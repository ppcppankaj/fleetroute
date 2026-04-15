package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jung-kurt/gofpdf"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	dbURL := getenv("DATABASE_URL", "postgres://gpsgo:gpsgo@localhost:5432/gpsgo")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("db connect", zap.Error(err))
	}
	defer pool.Close()

	nc, err := nats.Connect(getenv("NATS_URL", "nats://localhost:4222"))
	if err != nil {
		log.Fatal("nats connect", zap.Error(err))
	}
	defer nc.Close()
	js, _ := nc.JetStream()

	// Ensure stream
	js.AddStream(&nats.StreamConfig{
		Name:     "REPORTS",
		Subjects: []string{"REPORTS.>"},
		Storage:  nats.FileStorage,
	})

	// S3 client
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Warn("AWS config not available, using local storage fallback", zap.Error(err))
	}
	var s3Client *s3.Client
	if err == nil {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	w := &Worker{pool: pool, nc: nc, s3: s3Client, log: log,
		bucket: getenv("S3_BUCKET", "gpsgo-reports"),
		baseURL: getenv("REPORT_BASE_URL", "http://localhost:8085/reports"),
	}

	// Consume report jobs
	sub, err := js.QueueSubscribeSync("REPORTS.generate", "report-workers")
	if err != nil {
		log.Fatal("subscribe reports", zap.Error(err))
	}

	go func() {
		for {
			msg, err := sub.NextMsg(30 * time.Second)
			if err != nil {
				continue
			}
			var job ReportJob
			if err := json.Unmarshal(msg.Data, &job); err != nil {
				msg.Nak()
				continue
			}
			w.Process(job)
			msg.Ack()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("report-service stopped")
}

// ── Types ─────────────────────────────────────────────────────────────────────

type ReportJob struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	ReportType string         `json:"report_type"`
	Format     string         `json:"format"`
	Parameters map[string]any `json:"parameters"`
}

type Worker struct {
	pool    *pgxpool.Pool
	nc      *nats.Conn
	s3      *s3.Client
	log     *zap.Logger
	bucket  string
	baseURL string
}

// ── Process ───────────────────────────────────────────────────────────────────

func (w *Worker) Process(job ReportJob) {
	w.log.Info("processing report", zap.String("id", job.ID), zap.String("type", job.ReportType))
	w.updateStatus(job.ID, "processing", 0)

	var (
		data    [][]string
		headers []string
		err     error
	)

	switch job.ReportType {
	case "trip":
		headers, data, err = w.queryTrips(job)
	case "fuel":
		headers, data, err = w.queryFuel(job)
	case "driver_behavior":
		headers, data, err = w.queryDriverBehavior(job)
	case "geofence_violations":
		headers, data, err = w.queryGeofenceViolations(job)
	case "idle":
		headers, data, err = w.queryIdleEvents(job)
	case "overspeed":
		headers, data, err = w.queryOverspeedEvents(job)
	case "maintenance":
		headers, data, err = w.queryMaintenance(job)
	case "ais140_audit":
		headers, data, err = w.queryAIS140Audit(job)
	default:
		err = fmt.Errorf("unknown report type: %s", job.ReportType)
	}

	if err != nil {
		w.log.Error("query report data", zap.Error(err))
		w.updateStatusFailed(job.ID, err.Error())
		return
	}

	w.updateStatus(job.ID, "processing", 50)

	var fileBytes []byte
	var ext string

	if job.Format == "pdf" {
		fileBytes, err = renderPDF(job.ReportType, headers, data)
		ext = "pdf"
	} else {
		fileBytes, err = renderCSV(headers, data)
		ext = "csv"
	}

	if err != nil {
		w.updateStatusFailed(job.ID, err.Error())
		return
	}

	// Upload to S3 or local
	outputURL := w.upload(job, fileBytes, ext)

	w.log.Info("report completed", zap.String("id", job.ID), zap.String("url", outputURL))
	w.updateStatusComplete(job.ID, outputURL, int64(len(fileBytes)))
}

// ── Query Functions ───────────────────────────────────────────────────────────

func (w *Worker) queryTrips(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT t.id::text, v.registration, COALESCE(d.name,'Unknown'),
		       t.started_at::text, t.ended_at::text,
		       ROUND(t.distance_m/1000.0, 2)::text || ' km',
		       (t.duration_s/60)::text || ' min',
		       t.max_speed::text || ' km/h',
		       COALESCE(t.harsh_accel::text,'0'), COALESCE(t.harsh_brake::text,'0')
		FROM trips t
		JOIN vehicles v ON v.id = t.vehicle_id
		LEFT JOIN drivers d ON d.id = t.driver_id
		WHERE t.tenant_id = $1 AND t.started_at BETWEEN $2::timestamptz AND $3::timestamptz
		ORDER BY t.started_at DESC`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	headers := []string{"Trip ID", "Vehicle", "Driver", "Start", "End", "Distance", "Duration", "Max Speed", "Harsh Accel", "Harsh Brake"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = fmt.Sprint(v)
		}
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryFuel(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT v.registration,
		       date_trunc('day', a.timestamp)::date::text AS date,
		       ROUND(AVG(a.fuel_level_pct)::numeric, 1)::text AS avg_fuel_pct,
		       MIN(a.fuel_level_pct)::text AS min_fuel_pct
		FROM avl_records a
		JOIN vehicles v ON v.device_id = a.device_id
		WHERE a.tenant_id = $1 AND a.timestamp BETWEEN $2::timestamptz AND $3::timestamptz
		  AND a.fuel_level_pct > 0
		GROUP BY v.registration, date ORDER BY date DESC`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Vehicle", "Date", "Avg Fuel %", "Min Fuel %"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryDriverBehavior(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT d.name, COUNT(t.id)::text, SUM(t.harsh_accel)::text,
		       SUM(t.harsh_brake)::text, SUM(t.harsh_corner)::text,
		       SUM(t.overspeed_count)::text,
		       GREATEST(0, 100 - (SUM(t.harsh_accel)*2 + SUM(t.harsh_brake)*2 + SUM(t.harsh_corner) + SUM(t.overspeed_count)))::text AS score
		FROM trips t JOIN drivers d ON d.id = t.driver_id
		WHERE t.tenant_id=$1 AND t.started_at BETWEEN $2::timestamptz AND $3::timestamptz
		GROUP BY d.name ORDER BY score DESC`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Driver", "Trips", "Harsh Accel", "Harsh Brake", "Harsh Corner", "Overspeed", "Score"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryGeofenceViolations(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT v.registration, g.name AS geofence, ge.event_type,
		       ge.entered_at::text, ge.exited_at::text,
		       EXTRACT(EPOCH FROM (ge.exited_at - ge.entered_at))::int/60 AS dwell_minutes
		FROM geofence_events ge
		JOIN vehicles v ON v.id = ge.vehicle_id
		JOIN geofences g ON g.id = ge.geofence_id
		WHERE ge.tenant_id=$1 AND ge.entered_at BETWEEN $2::timestamptz AND $3::timestamptz
		ORDER BY ge.entered_at DESC`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Vehicle", "Geofence", "Event Type", "Entered At", "Exited At", "Dwell (min)"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryIdleEvents(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT v.registration, a.timestamp::text, a.ignition::text, a.speed::text,
		       a.lat::text, a.lng::text
		FROM avl_records a JOIN vehicles v ON v.device_id = a.device_id
		WHERE a.tenant_id=$1 AND a.timestamp BETWEEN $2::timestamptz AND $3::timestamptz
		  AND a.ignition = true AND a.speed = 0
		ORDER BY a.timestamp DESC LIMIT 5000`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Vehicle", "Timestamp", "Ignition", "Speed", "Lat", "Lng"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryOverspeedEvents(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	threshold := 80
	if v, ok := job.Parameters["speed_threshold"]; ok {
		fmt.Sscanf(fmt.Sprint(v), "%d", &threshold)
	}
	rows, err := w.pool.Query(context.Background(), `
		SELECT v.registration, a.timestamp::text, a.speed::text, a.lat::text, a.lng::text
		FROM avl_records a JOIN vehicles v ON v.device_id = a.device_id
		WHERE a.tenant_id=$1 AND a.timestamp BETWEEN $2::timestamptz AND $3::timestamptz
		  AND a.speed > $4
		ORDER BY a.timestamp DESC LIMIT 5000`, job.TenantID, from, to, threshold)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Vehicle", "Timestamp", "Speed (km/h)", "Lat", "Lng"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryMaintenance(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT v.registration, sl.service_type, sl.serviced_at::text,
		       sl.odometer_m::text, sl.technician, sl.cost::text, sl.notes
		FROM service_log sl JOIN vehicles v ON v.id = sl.vehicle_id
		WHERE sl.tenant_id=$1 AND sl.serviced_at BETWEEN $2::timestamptz AND $3::timestamptz
		ORDER BY sl.serviced_at DESC`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Vehicle", "Service Type", "Date", "Odometer", "Technician", "Cost", "Notes"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

func (w *Worker) queryAIS140Audit(job ReportJob) ([]string, [][]string, error) {
	from, to := getDateRange(job.Parameters)
	rows, err := w.pool.Query(context.Background(), `
		SELECT a.device_id::text, d.imei, a.timestamp::text,
		       a.lat::text, a.lng::text, a.speed::text, a.heading::text,
		       a.ignition::text, a.valid::text
		FROM avl_records a JOIN devices d ON d.id = a.device_id
		WHERE a.tenant_id=$1 AND a.timestamp BETWEEN $2::timestamptz AND $3::timestamptz
		  AND d.protocol = 'ais140'
		ORDER BY a.timestamp`, job.TenantID, from, to)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	headers := []string{"Device ID", "IMEI", "Timestamp", "Latitude", "Longitude", "Speed", "Heading", "Ignition", "GPS Valid"}
	var data [][]string
	for rows.Next() {
		vals, _ := rows.Values()
		row := make([]string, len(vals))
		for i, v := range vals { row[i] = fmt.Sprint(v) }
		data = append(data, row)
	}
	return headers, data, nil
}

// ── Render ────────────────────────────────────────────────────────────────────

func renderCSV(headers []string, data [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write(headers)
	w.WriteAll(data)
	w.Flush()
	return buf.Bytes(), w.Error()
}

func renderPDF(title string, headers []string, data [][]string) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("FleetOS Report: %s", title))
	pdf.Ln(12)
	pdf.SetFont("Arial", "B", 9)
	colW := 270.0 / float64(len(headers))
	for _, h := range headers {
		pdf.CellFormat(colW, 7, h, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 8)
	fill := false
	pdf.SetFillColor(240, 240, 240)
	for _, row := range data {
		for i, cell := range row {
			if i >= len(headers) { break }
			if len(cell) > 30 { cell = cell[:30] + "…" }
			pdf.CellFormat(colW, 6, cell, "1", 0, "", fill, 0, "")
		}
		pdf.Ln(-1)
		fill = !fill
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ── Upload ────────────────────────────────────────────────────────────────────

func (w *Worker) upload(job ReportJob, data []byte, ext string) string {
	key := fmt.Sprintf("reports/%s/%s_%s.%s", job.TenantID, job.ReportType, job.ID[:8], ext)
	if w.s3 != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		contentType := "text/csv"
		if ext == "pdf" {
			contentType = "application/pdf"
		}
		_, err := w.s3.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(w.bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(data),
			ContentType: aws.String(contentType),
		})
		if err == nil {
			return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", w.bucket, key)
		}
		w.log.Warn("S3 upload failed, saving locally", zap.Error(err))
	}
	// Fallback: local file
	localPath := fmt.Sprintf("/tmp/%s", key)
	os.MkdirAll(fmt.Sprintf("/tmp/reports/%s", job.TenantID), 0755)
	os.WriteFile(localPath, data, 0644)
	return fmt.Sprintf("%s/%s", w.baseURL, key)
}

// ── DB Status Updates ─────────────────────────────────────────────────────────

func (w *Worker) updateStatus(id, status string, progress int) {
	w.pool.Exec(context.Background(), `
		UPDATE report_jobs SET status=$2, progress_pct=$3, started_at=COALESCE(started_at,now()), updated_at=now()
		WHERE id=$1`, id, status, progress)
}

func (w *Worker) updateStatusFailed(id, errMsg string) {
	w.pool.Exec(context.Background(), `
		UPDATE report_jobs SET status='failed', error_msg=$2, completed_at=now(), updated_at=now()
		WHERE id=$1`, id, errMsg)
}

func (w *Worker) updateStatusComplete(id, url string, size int64) {
	w.pool.Exec(context.Background(), `
		UPDATE report_jobs SET status='completed', output_url=$2, output_size_b=$3,
		  progress_pct=100, completed_at=now(), expires_at=now()+interval '1 hour', updated_at=now()
		WHERE id=$1`, id, url, size)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func getDateRange(params map[string]any) (string, string) {
	from := "1970-01-01"
	to := time.Now().Format(time.RFC3339)
	if v, ok := params["from"]; ok { from = fmt.Sprint(v) }
	if v, ok := params["to"]; ok   { to = fmt.Sprint(v) }
	return from, to
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
