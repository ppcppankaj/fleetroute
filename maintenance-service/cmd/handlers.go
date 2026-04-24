package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// ── Service Schedules ─────────────────────────────────────────────────────────

func (h *Handler) ListSchedules(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	vehicleID := c.Query("vehicle_id")

	q := `SELECT ss.id, ss.vehicle_id, v.registration, ss.service_type, ss.description,
	             ss.interval_days, ss.interval_km, ss.last_service_at, ss.last_odometer_m,
	             ss.next_due_at, ss.next_due_odometer, ss.enabled,
	             COALESCE(v.current_odometer_m, 0) AS current_odometer_m,
	             EXTRACT(DAY FROM ss.next_due_at - now())::int AS days_until_due,
	             CASE
	               WHEN ss.next_due_at < now() THEN 'overdue'
	               WHEN ss.next_due_at < now() + (ss.warn_days_before || ' days')::interval THEN 'due_soon'
	               ELSE 'ok'
	             END AS status
	      FROM service_schedules ss
	      JOIN vehicles v ON v.id = ss.vehicle_id
	      WHERE ss.tenant_id = $1 AND ss.deleted_at IS NULL`
	args := []any{tenantID}
	if vehicleID != "" {
		q += ` AND ss.vehicle_id = $2`
		args = append(args, vehicleID)
	}
	q += ` ORDER BY ss.next_due_at ASC NULLS LAST`

	rows, err := h.pool.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var schedules []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id", "vehicle_id", "registration", "service_type", "description",
			"interval_days", "interval_km", "last_service_at", "last_odometer_m",
			"next_due_at", "next_due_odometer", "enabled",
			"current_odometer_m", "days_until_due", "status"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		schedules = append(schedules, m)
	}
	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

func (h *Handler) CreateSchedule(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		VehicleID       string  `json:"vehicle_id" binding:"required"`
		ServiceType     string  `json:"service_type" binding:"required"`
		Description     string  `json:"description"`
		IntervalDays    *int    `json:"interval_days"`
		IntervalKm      *int    `json:"interval_km"`
		WarnDaysBefore  int     `json:"warn_days_before"`
		WarnKmBefore    int     `json:"warn_km_before"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.WarnDaysBefore == 0 {
		body.WarnDaysBefore = 7
	}
	if body.WarnKmBefore == 0 {
		body.WarnKmBefore = 500
	}

	var id string
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO service_schedules
		  (tenant_id, vehicle_id, service_type, description,
		   interval_days, interval_km, warn_days_before, warn_km_before,
		   next_due_at, next_due_odometer)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,
		        CASE WHEN $5 IS NOT NULL THEN now() + ($5 || ' days')::interval ELSE NULL END,
		        CASE WHEN $6 IS NOT NULL THEN (SELECT COALESCE(current_odometer_m,0)+$6*1000 FROM vehicles WHERE id=$2) ELSE NULL END)
		RETURNING id`,
		tenantID, body.VehicleID, body.ServiceType, body.Description,
		body.IntervalDays, body.IntervalKm, body.WarnDaysBefore, body.WarnKmBefore,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) UpdateSchedule(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	id := c.Param("id")
	var body struct {
		Description  string `json:"description"`
		IntervalDays *int   `json:"interval_days"`
		IntervalKm   *int   `json:"interval_km"`
		Enabled      *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.pool.Exec(c.Request.Context(), `
		UPDATE service_schedules SET
		  description = COALESCE(NULLIF($3,''), description),
		  interval_days = COALESCE($4, interval_days),
		  interval_km = COALESCE($5, interval_km),
		  enabled = COALESCE($6, enabled),
		  updated_at = now()
		WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID, body.Description, body.IntervalDays, body.IntervalKm, body.Enabled,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteSchedule(c *gin.Context) {
	h.pool.Exec(c.Request.Context(),
		`UPDATE service_schedules SET deleted_at=now() WHERE id=$1 AND tenant_id=$2`,
		c.Param("id"), c.GetHeader("X-Tenant-ID"))
	c.Status(http.StatusNoContent)
}

// CompleteService marks a service as done, creates a log entry, and resets next_due
func (h *Handler) CompleteService(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	scheduleID := c.Param("id")
	var body struct {
		ServicedAt    string   `json:"serviced_at" binding:"required"`
		OdometerM     *int64   `json:"odometer_m"`
		Technician    string   `json:"technician"`
		ServiceCenter string   `json:"service_center"`
		Cost          *float64 `json:"cost"`
		Notes         string   `json:"notes"`
		PartsUsed     []map[string]any `json:"parts_used"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get schedule info
	var vehicleID, serviceType string
	var intervalDays, intervalKm *int
	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT vehicle_id, service_type, interval_days, interval_km
		FROM service_schedules WHERE id=$1 AND tenant_id=$2`,
		scheduleID, tenantID,
	).Scan(&vehicleID, &serviceType, &intervalDays, &intervalKm)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	partsJSON := "[]"
	if len(body.PartsUsed) > 0 {
		partsJSON = fmt.Sprintf("%v", body.PartsUsed)
	}

	// Insert service log
	var logID string
	err = h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO service_log
		  (tenant_id, vehicle_id, schedule_id, service_type, serviced_at,
		   odometer_m, technician, service_center, cost, notes, parts_used)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb)
		RETURNING id`,
		tenantID, vehicleID, scheduleID, serviceType, body.ServicedAt,
		body.OdometerM, body.Technician, body.ServiceCenter,
		body.Cost, body.Notes, partsJSON,
	).Scan(&logID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reset schedule: update last_service and compute next_due
	h.pool.Exec(c.Request.Context(), `
		UPDATE service_schedules SET
		  last_service_at = $3,
		  last_odometer_m = $4,
		  next_due_at = CASE WHEN interval_days IS NOT NULL THEN $3::timestamptz + (interval_days || ' days')::interval ELSE NULL END,
		  next_due_odometer = CASE WHEN interval_km IS NOT NULL THEN $4 + interval_km*1000 ELSE NULL END,
		  updated_at = now()
		WHERE id=$1 AND tenant_id=$2`,
		scheduleID, tenantID, body.ServicedAt, body.OdometerM,
	)

	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"log_id": logID}})
}

// ── Service Log ───────────────────────────────────────────────────────────────

func (h *Handler) ListServiceLog(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT sl.id, sl.vehicle_id, v.registration, sl.service_type,
		       sl.serviced_at, sl.odometer_m, sl.technician,
		       sl.service_center, sl.cost, sl.notes, sl.parts_used
		FROM service_log sl
		JOIN vehicles v ON v.id = sl.vehicle_id
		WHERE sl.tenant_id = $1
		ORDER BY sl.serviced_at DESC LIMIT 200`, tenantID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var logs []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id", "vehicle_id", "registration", "service_type",
			"serviced_at", "odometer_m", "technician", "service_center", "cost", "notes", "parts_used"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		logs = append(logs, m)
	}
	c.JSON(200, gin.H{"data": logs})
}

func (h *Handler) GetServiceLogEntry(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var vals []any
	h.pool.QueryRow(c.Request.Context(),
		`SELECT id, vehicle_id, service_type, serviced_at, odometer_m, technician, service_center, cost, notes, parts_used
         FROM service_log WHERE id=$1 AND tenant_id=$2`, c.Param("id"), tenantID).Scan(&vals)
	c.JSON(200, gin.H{"data": vals})
}

// ── Documents ─────────────────────────────────────────────────────────────────

func (h *Handler) ListDocuments(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT vd.id, vd.vehicle_id, v.registration, vd.doc_type, vd.doc_number,
		       vd.issued_at, vd.expires_at, vd.file_url, vd.issuer,
		       (vd.expires_at - CURRENT_DATE)::int AS days_until_expiry,
		       CASE
		         WHEN vd.expires_at < CURRENT_DATE THEN 'expired'
		         WHEN vd.expires_at < CURRENT_DATE + INTERVAL '30 days' THEN 'expiring_soon'
		         ELSE 'valid'
		       END AS expiry_status
		FROM vehicle_documents vd
		JOIN vehicles v ON v.id = vd.vehicle_id
		WHERE vd.tenant_id = $1 AND vd.deleted_at IS NULL
		ORDER BY vd.expires_at ASC NULLS LAST`, tenantID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var docs []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id","vehicle_id","registration","doc_type","doc_number",
			"issued_at","expires_at","file_url","issuer","days_until_expiry","expiry_status"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) { m[f] = vals[i] }
		}
		docs = append(docs, m)
	}
	c.JSON(200, gin.H{"data": docs})
}

func (h *Handler) CreateDocument(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		VehicleID string `json:"vehicle_id" binding:"required"`
		DocType   string `json:"doc_type" binding:"required"`
		DocNumber string `json:"doc_number"`
		IssuedAt  string `json:"issued_at"`
		ExpiresAt string `json:"expires_at"`
		FileURL   string `json:"file_url"`
		Issuer    string `json:"issuer"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var id string
	h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO vehicle_documents (tenant_id, vehicle_id, doc_type, doc_number, issued_at, expires_at, file_url, issuer)
		VALUES ($1,$2,$3,$4,$5::date,$6::date,$7,$8) RETURNING id`,
		tenantID, body.VehicleID, body.DocType, body.DocNumber,
		body.IssuedAt, body.ExpiresAt, body.FileURL, body.Issuer,
	).Scan(&id)
	c.JSON(201, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	h.pool.Exec(c.Request.Context(),
		`UPDATE vehicle_documents SET deleted_at=now() WHERE id=$1 AND tenant_id=$2`,
		c.Param("id"), c.GetHeader("X-Tenant-ID"))
	c.Status(204)
}

// ── Spare Parts ───────────────────────────────────────────────────────────────

func (h *Handler) ListParts(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, part_number, description, qty_in_stock, reorder_threshold, unit_cost, currency, supplier
		FROM spare_parts WHERE tenant_id=$1 AND deleted_at IS NULL ORDER BY name`, tenantID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var parts []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id","name","part_number","description","qty_in_stock","reorder_threshold","unit_cost","currency","supplier"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) { m[f] = vals[i] }
		}
		parts = append(parts, m)
	}
	c.JSON(200, gin.H{"data": parts})
}

func (h *Handler) CreatePart(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		Name             string   `json:"name" binding:"required"`
		PartNumber       string   `json:"part_number"`
		Description      string   `json:"description"`
		QtyInStock       int      `json:"qty_in_stock"`
		ReorderThreshold int      `json:"reorder_threshold"`
		UnitCost         *float64 `json:"unit_cost"`
		Currency         string   `json:"currency"`
		Supplier         string   `json:"supplier"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if body.Currency == "" { body.Currency = "INR" }
	var id string
	h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO spare_parts (tenant_id,name,part_number,description,qty_in_stock,reorder_threshold,unit_cost,currency,supplier)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		tenantID, body.Name, body.PartNumber, body.Description,
		body.QtyInStock, body.ReorderThreshold, body.UnitCost, body.Currency, body.Supplier,
	).Scan(&id)
	c.JSON(201, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) UpdatePart(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		QtyInStock       *int     `json:"qty_in_stock"`
		ReorderThreshold *int     `json:"reorder_threshold"`
		UnitCost         *float64 `json:"unit_cost"`
	}
	c.ShouldBindJSON(&body)
	h.pool.Exec(c.Request.Context(), `
		UPDATE spare_parts SET
		  qty_in_stock = COALESCE($3, qty_in_stock),
		  reorder_threshold = COALESCE($4, reorder_threshold),
		  unit_cost = COALESCE($5, unit_cost),
		  updated_at = now()
		WHERE id=$1 AND tenant_id=$2`,
		c.Param("id"), tenantID, body.QtyInStock, body.ReorderThreshold, body.UnitCost,
	)
	c.Status(204)
}
