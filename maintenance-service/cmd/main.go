package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

	// Subscribe to enriched GPS events to track live odometer
	go subscribeOdometer(js, pool, log)

	// Schedule document expiry check daily at 08:00
	go runExpiryChecker(pool, nc, log)

	// HTTP API for CRUD
	r := gin.New()
	r.Use(gin.Recovery())
	h := &Handler{pool: pool, log: log}

	v1 := r.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Service schedules
	v1.GET("/maintenance/schedules", h.ListSchedules)
	v1.POST("/maintenance/schedules", h.CreateSchedule)
	v1.PUT("/maintenance/schedules/:id", h.UpdateSchedule)
	v1.DELETE("/maintenance/schedules/:id", h.DeleteSchedule)
	v1.POST("/maintenance/schedules/:id/complete", h.CompleteService)

	// Service log
	v1.GET("/maintenance/log", h.ListServiceLog)
	v1.GET("/maintenance/log/:id", h.GetServiceLogEntry)

	// Documents
	v1.GET("/maintenance/documents", h.ListDocuments)
	v1.POST("/maintenance/documents", h.CreateDocument)
	v1.DELETE("/maintenance/documents/:id", h.DeleteDocument)

	// Spare parts
	v1.GET("/maintenance/parts", h.ListParts)
	v1.POST("/maintenance/parts", h.CreatePart)
	v1.PUT("/maintenance/parts/:id", h.UpdatePart)

	port := getenv("PORT", "8084")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Info("maintenance-service listening", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("maintenance-service stopped")
}

// subscribeOdometer reads enriched events from NATS and updates vehicle odometer
func subscribeOdometer(js nats.JetStreamContext, pool *pgxpool.Pool, log *zap.Logger) {
	sub, err := js.QueueSubscribeSync("GPS_ENRICHED.>", "maintenance-odometer")
	if err != nil {
		log.Error("subscribe enriched", zap.Error(err))
		return
	}
	for {
		msg, err := sub.NextMsg(30 * time.Second)
		if err != nil {
			continue
		}
		var ev struct {
			DeviceID   string `json:"device_id"`
			TenantID   string `json:"tenant_id"`
			TotalOdoM  int64  `json:"total_odometer_m"`
			Timestamp  string `json:"timestamp"`
		}
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			msg.Nak()
			continue
		}
		if ev.TotalOdoM > 0 {
			updateVehicleOdometer(pool, ev.DeviceID, ev.TotalOdoM, log)
		}
		msg.Ack()
	}
}

func updateVehicleOdometer(pool *pgxpool.Pool, deviceID string, odoM int64, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, `
		UPDATE vehicles SET current_odometer_m = $2, updated_at = now()
		WHERE device_id = (SELECT id FROM devices WHERE imei = $1 OR id::text = $1 LIMIT 1)
		  AND current_odometer_m < $2
	`, deviceID, odoM)
	if err != nil {
		log.Warn("update odometer", zap.Error(err))
	}
}

// runExpiryChecker fires daily to detect near-expiry documents and overdue services
func runExpiryChecker(pool *pgxpool.Pool, nc *nats.Conn, log *zap.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	// Also run immediately on startup
	checkExpiries(pool, nc, log)
	for range ticker.C {
		checkExpiries(pool, nc, log)
	}
}

func checkExpiries(pool *pgxpool.Pool, nc *nats.Conn, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Documents expiring within 30 days
	rows, err := pool.Query(ctx, `
		SELECT vd.id, vd.tenant_id, vd.vehicle_id, vd.doc_type, vd.expires_at,
		       EXTRACT(DAY FROM vd.expires_at - CURRENT_DATE)::int AS days_left
		FROM vehicle_documents vd
		WHERE vd.deleted_at IS NULL
		  AND vd.expires_at BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
	`)
	if err != nil {
		log.Error("check doc expiry", zap.Error(err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, tenantID, vehicleID, docType string
		var expiresAt time.Time
		var daysLeft int
		if err := rows.Scan(&id, &tenantID, &vehicleID, &docType, &expiresAt, &daysLeft); err != nil {
			continue
		}
		alert := map[string]any{
			"type":       "document_expiry",
			"tenant_id":  tenantID,
			"vehicle_id": vehicleID,
			"doc_type":   docType,
			"expires_at": expiresAt,
			"days_left":  daysLeft,
		}
		b, _ := json.Marshal(alert)
		_ = nc.Publish("ALERTS.document_expiry", b)
		log.Info("document expiry alert", zap.String("doc_type", docType), zap.Int("days_left", daysLeft))
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
