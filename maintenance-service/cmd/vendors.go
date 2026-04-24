package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── Vendors ───────────────────────────────────────────────────────────────────

func (h *Handler) ListVendors(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, contact_name, phone, email, address, services, rating, notes, created_at
		FROM maintenance_vendors
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name`, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var vendors []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id", "name", "contact_name", "phone", "email", "address", "services", "rating", "notes", "created_at"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		vendors = append(vendors, m)
	}
	c.JSON(http.StatusOK, gin.H{"data": vendors})
}

func (h *Handler) CreateVendor(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		Name        string   `json:"name" binding:"required"`
		ContactName string   `json:"contact_name"`
		Phone       string   `json:"phone"`
		Email       string   `json:"email"`
		Address     string   `json:"address"`
		Services    []string `json:"services"`
		Notes       string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id string
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO maintenance_vendors (tenant_id, name, contact_name, phone, email, address, services, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		tenantID, body.Name, body.ContactName, body.Phone, body.Email, body.Address, body.Services, body.Notes,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) UpdateVendor(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	id := c.Param("id")
	var body struct {
		Name        string   `json:"name"`
		ContactName string   `json:"contact_name"`
		Phone       string   `json:"phone"`
		Email       string   `json:"email"`
		Address     string   `json:"address"`
		Services    []string `json:"services"`
		Rating      *float64 `json:"rating"`
		Notes       string   `json:"notes"`
	}
	c.ShouldBindJSON(&body)
	_, err := h.pool.Exec(c.Request.Context(), `
		UPDATE maintenance_vendors SET
		  name         = COALESCE(NULLIF($3,''), name),
		  contact_name = COALESCE(NULLIF($4,''), contact_name),
		  phone        = COALESCE(NULLIF($5,''), phone),
		  email        = COALESCE(NULLIF($6,''), email),
		  address      = COALESCE(NULLIF($7,''), address),
		  rating       = COALESCE($8, rating),
		  notes        = COALESCE(NULLIF($9,''), notes),
		  updated_at   = now()
		WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID, body.Name, body.ContactName, body.Phone, body.Email, body.Address, body.Rating, body.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteVendor(c *gin.Context) {
	h.pool.Exec(c.Request.Context(),
		`UPDATE maintenance_vendors SET deleted_at=now() WHERE id=$1 AND tenant_id=$2`,
		c.Param("id"), c.GetHeader("X-Tenant-ID"))
	c.Status(http.StatusNoContent)
}

// ── Inspections ────────────────────────────────────────────────────────────────

func (h *Handler) ListInspections(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	vehicleID := c.Query("vehicle_id")

	q := `SELECT i.id, i.vehicle_id, v.registration, i.inspection_type, i.performed_by,
	             i.inspected_at, i.result, i.notes, i.checklist, i.next_due_at
	      FROM maintenance_inspections i
	      JOIN vehicles v ON v.id = i.vehicle_id
	      WHERE i.tenant_id = $1`
	args := []any{tenantID}
	if vehicleID != "" {
		q += ` AND i.vehicle_id = $2`
		args = append(args, vehicleID)
	}
	q += ` ORDER BY i.inspected_at DESC LIMIT 100`

	rows, err := h.pool.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var inspections []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id", "vehicle_id", "registration", "inspection_type", "performed_by",
			"inspected_at", "result", "notes", "checklist", "next_due_at"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		inspections = append(inspections, m)
	}
	c.JSON(http.StatusOK, gin.H{"data": inspections})
}

func (h *Handler) CreateInspection(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		VehicleID      string         `json:"vehicle_id" binding:"required"`
		InspectionType string         `json:"inspection_type" binding:"required"` // routine|safety|emission|annual
		PerformedBy    string         `json:"performed_by"`
		InspectedAt    string         `json:"inspected_at" binding:"required"`
		Result         string         `json:"result"` // pass|fail|conditional
		NextDueAt      string         `json:"next_due_at"`
		Checklist      map[string]any `json:"checklist"`
		Notes          string         `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checklistJSON := "{}"
	if len(body.Checklist) > 0 {
		b, _ := json.Marshal(body.Checklist)
		checklistJSON = string(b)
	}

	var id string
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO maintenance_inspections
		  (tenant_id, vehicle_id, inspection_type, performed_by, inspected_at, result, next_due_at, checklist, notes)
		VALUES ($1,$2,$3,$4,$5::timestamptz,$6,$7::date,$8::jsonb,$9)
		RETURNING id`,
		tenantID, body.VehicleID, body.InspectionType, body.PerformedBy,
		body.InspectedAt, body.Result, nilIfEmpty(body.NextDueAt), checklistJSON, body.Notes,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

// ── Tyre Management ───────────────────────────────────────────────────────────

func (h *Handler) ListTyres(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	vehicleID := c.Query("vehicle_id")

	q := `SELECT t.id, t.vehicle_id, v.registration, t.position, t.brand, t.size,
	             t.serial_number, t.fitted_at, t.fitted_km, t.replaced_at,
	             t.tread_depth_mm, t.condition, t.notes
	      FROM tyre_management t
	      JOIN vehicles v ON v.id = t.vehicle_id
	      WHERE t.tenant_id = $1`
	args := []any{tenantID}
	if vehicleID != "" {
		q += ` AND t.vehicle_id = $2`
		args = append(args, vehicleID)
	}
	q += ` ORDER BY v.registration, t.position`

	rows, err := h.pool.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var tyres []map[string]any
	for rows.Next() {
		vals, _ := rows.Values()
		fields := []string{"id", "vehicle_id", "registration", "position", "brand", "size",
			"serial_number", "fitted_at", "fitted_km", "replaced_at", "tread_depth_mm", "condition", "notes"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		tyres = append(tyres, m)
	}
	c.JSON(http.StatusOK, gin.H{"data": tyres})
}

func (h *Handler) CreateTyre(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var body struct {
		VehicleID    string   `json:"vehicle_id" binding:"required"`
		Position     string   `json:"position" binding:"required"` // FL|FR|RL|RR|spare
		Brand        string   `json:"brand"`
		Size         string   `json:"size"`
		SerialNumber string   `json:"serial_number"`
		FittedAt     string   `json:"fitted_at"`
		FittedKm     *int     `json:"fitted_km"`
		TreadDepthMM *float64 `json:"tread_depth_mm"`
		Condition    string   `json:"condition"` // good|worn|replace
		Notes        string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Condition == "" {
		body.Condition = "good"
	}
	var id string
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO tyre_management
		  (tenant_id, vehicle_id, position, brand, size, serial_number,
		   fitted_at, fitted_km, tread_depth_mm, condition, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7::date,$8,$9,$10,$11)
		RETURNING id`,
		tenantID, body.VehicleID, body.Position, body.Brand, body.Size, body.SerialNumber,
		nilIfEmpty(body.FittedAt), body.FittedKm, body.TreadDepthMM, body.Condition, body.Notes,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) UpdateTyre(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	id := c.Param("id")
	var body struct {
		TreadDepthMM *float64 `json:"tread_depth_mm"`
		Condition    string   `json:"condition"`
		ReplacedAt   string   `json:"replaced_at"`
		Notes        string   `json:"notes"`
	}
	c.ShouldBindJSON(&body)
	_, err := h.pool.Exec(c.Request.Context(), `
		UPDATE tyre_management SET
		  tread_depth_mm = COALESCE($3, tread_depth_mm),
		  condition      = COALESCE(NULLIF($4,''), condition),
		  replaced_at    = COALESCE($5::date, replaced_at),
		  notes          = COALESCE(NULLIF($6,''), notes),
		  updated_at     = now()
		WHERE id=$1 AND tenant_id=$2`,
		id, tenantID, body.TreadDepthMM, body.Condition, nilIfEmpty(body.ReplacedAt), body.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
