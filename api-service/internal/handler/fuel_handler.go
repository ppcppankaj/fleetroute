package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	pkgauth "gpsgo/pkg/auth"
)

// ── FuelHandler ────────────────────────────────────────────────────────────────

type FuelHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewFuelHandler(pool *pgxpool.Pool, logger *zap.Logger) *FuelHandler {
	return &FuelHandler{pool: pool, logger: logger}
}

// ListFuelLogs returns fuel logs for the tenant with optional vehicle filter.
func (h *FuelHandler) ListFuelLogs(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	vehicleID := c.Query("vehicle_id")
	limit := c.DefaultQuery("limit", "100")

	q := `SELECT fl.id, fl.vehicle_id, v.registration, fl.driver_id,
	             fl.liters, fl.cost_per_liter, fl.total_cost, fl.currency,
	             fl.odometer_km, fl.station_name, fl.fill_type, fl.filled_at
	      FROM fuel_logs fl
	      JOIN vehicles v ON v.id = fl.vehicle_id
	      WHERE fl.tenant_id = $1`
	args := []any{tenantID}

	if vehicleID != "" {
		q += ` AND fl.vehicle_id = $2`
		args = append(args, vehicleID)
	}
	q += ` ORDER BY fl.filled_at DESC LIMIT ` + limit

	rows, err := h.pool.Query(c.Request.Context(), q, args...)
	if err != nil {
		h.logger.Error("fuel logs list", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var logs []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			h.logger.Error("fuel logs: scan row", zap.Error(err))
			continue
		}
		fields := []string{"id", "vehicle_id", "registration", "driver_id",
			"liters", "cost_per_liter", "total_cost", "currency",
			"odometer_km", "station_name", "fill_type", "filled_at"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		logs = append(logs, m)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("fuel logs: row iteration", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondOK(c, logs)
}

// CreateFuelLog adds a manual fuel fill-up record.
func (h *FuelHandler) CreateFuelLog(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body struct {
		VehicleID    string   `json:"vehicle_id" binding:"required"`
		DriverID     *string  `json:"driver_id"`
		Liters       float64  `json:"liters" binding:"required,gt=0"`
		CostPerLiter float64  `json:"cost_per_liter" binding:"required,gt=0"`
		Currency     string   `json:"currency"`
		OdometerKm   *int64   `json:"odometer_km"`
		StationName  string   `json:"station_name"`
		FillType     string   `json:"fill_type"`
		FilledAt     string   `json:"filled_at" binding:"required"`
		ReceiptURL   string   `json:"receipt_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.Currency == "" {
		body.Currency = "INR"
	}
	if body.FillType == "" {
		body.FillType = "full"
	}

	var id string
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO fuel_logs
		  (tenant_id, vehicle_id, driver_id, liters, cost_per_liter, currency,
		   odometer_km, station_name, fill_type, filled_at, receipt_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::timestamptz,$11)
		RETURNING id`,
		tenantID, body.VehicleID, body.DriverID, body.Liters, body.CostPerLiter, body.Currency,
		body.OdometerKm, body.StationName, body.FillType, body.FilledAt, body.ReceiptURL,
	).Scan(&id)
	if err != nil {
		h.logger.Error("fuel log create", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondCreated(c, gin.H{"id": id})
}

// FuelSummary returns aggregated fuel stats per vehicle for the time range.
func (h *FuelHandler) FuelSummary(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	from := c.DefaultQuery("from", "now()-interval '30 days'")
	to := c.DefaultQuery("to", "now()")

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT fl.vehicle_id, v.registration,
		       COUNT(*)::int                         AS fill_count,
		       COALESCE(SUM(fl.liters), 0)           AS total_liters,
		       COALESCE(SUM(fl.total_cost), 0)       AS total_cost,
		       COALESCE(AVG(fl.cost_per_liter), 0)   AS avg_cost_per_liter,
		       MAX(fl.filled_at)                     AS last_fill
		FROM fuel_logs fl
		JOIN vehicles v ON v.id = fl.vehicle_id
		WHERE fl.tenant_id = $1
		  AND fl.filled_at BETWEEN $2::timestamptz AND $3::timestamptz
		GROUP BY fl.vehicle_id, v.registration
		ORDER BY total_liters DESC`, tenantID, from, to)
	if err != nil {
		h.logger.Error("fuel summary query", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var summary []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			h.logger.Error("fuel summary: scan row", zap.Error(err))
			continue
		}
		fields := []string{"vehicle_id", "registration", "fill_count",
			"total_liters", "total_cost", "avg_cost_per_liter", "last_fill"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		summary = append(summary, m)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("fuel summary: row iteration", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondOK(c, summary)
}

// ListAnomalies returns fuel theft/anomaly events.
func (h *FuelHandler) ListAnomalies(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	vehicleID := c.Query("vehicle_id")
	onlyUnconfirmed := c.Query("unconfirmed") == "true"

	q := `SELECT fa.id, fa.vehicle_id, v.registration, fa.anomaly_type,
	             fa.drop_liters, fa.drop_percent, fa.start_level, fa.end_level,
	             fa.detected_at, fa.confirmed, fa.notes
	      FROM fuel_anomalies fa
	      JOIN vehicles v ON v.id = fa.vehicle_id
	      WHERE fa.tenant_id = $1`
	args := []any{tenantID}

	if vehicleID != "" {
		args = append(args, vehicleID)
		q += ` AND fa.vehicle_id = $` + fmt.Sprint(len(args))
	}
	if onlyUnconfirmed {
		q += ` AND fa.confirmed = false`
	}
	q += ` ORDER BY fa.detected_at DESC LIMIT 200`

	rows, err := h.pool.Query(c.Request.Context(), q, args...)
	if err != nil {
		h.logger.Error("fuel anomalies list", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var anomalies []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			h.logger.Error("fuel anomalies: scan row", zap.Error(err))
			continue
		}
		fields := []string{"id", "vehicle_id", "registration", "anomaly_type",
			"drop_liters", "drop_percent", "start_level", "end_level",
			"detected_at", "confirmed", "notes"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		anomalies = append(anomalies, m)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("fuel anomalies: row iteration", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondOK(c, anomalies)
}

// ConfirmAnomaly marks a fuel anomaly as confirmed theft.
func (h *FuelHandler) ConfirmAnomaly(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	userID := pkgauth.UserID(c)
	id := c.Param("id")

	var body struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.pool.Exec(c.Request.Context(), `
		UPDATE fuel_anomalies SET
		  confirmed = true, confirmed_by = $3::uuid, confirmed_at = now(),
		  notes = COALESCE(NULLIF($4,''), notes)
		WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, userID, body.Notes)
	if err != nil {
		h.logger.Error("fuel anomaly confirm", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── Real AlertRules CRUD (replaces stub) ──────────────────────────────────────

// ListRulesFromDB returns all alert rules for the tenant from the database.
func (h *RuleHandler) ListFromDB(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, alert_type, severity, conditions, speed_limit,
		       COALESCE(cooldown_s, 300), enabled, trigger_count, last_triggered, created_at
		FROM alert_rules
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`, tenantID)
	if err != nil {
		h.logger.Error("list alert rules", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var rules []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			h.logger.Error("alert rules: scan row", zap.Error(err))
			continue
		}
		fields := []string{"id", "name", "alert_type", "severity", "conditions",
			"speed_limit", "cooldown_s", "enabled", "trigger_count",
			"last_triggered", "created_at"}
		m := make(map[string]any)
		for i, f := range fields {
			if i < len(vals) {
				m[f] = vals[i]
			}
		}
		rules = append(rules, m)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("alert rules: row iteration", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondOK(c, rules)
}

func (h *RuleHandler) CreateFromDB(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body struct {
		Name        string  `json:"name" binding:"required"`
		AlertType   string  `json:"alert_type" binding:"required"`
		Severity    string  `json:"severity"`
		Conditions  any     `json:"conditions"`
		SpeedLimit  *int    `json:"speed_limit"`
		CooldownS   int     `json:"cooldown_s"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.Severity == "" {
		body.Severity = "warning"
	}
	if body.CooldownS == 0 {
		body.CooldownS = 300
	}

	condJSON := "[]"
	if body.Conditions != nil {
		if b, err := json.Marshal(body.Conditions); err == nil {
			condJSON = string(b)
		}
	}

	var id string
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO alert_rules
		  (tenant_id, name, alert_type, severity, conditions, speed_limit, cooldown_s, description)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7,$8)
		RETURNING id`,
		tenantID, body.Name, body.AlertType, body.Severity, condJSON,
		body.SpeedLimit, body.CooldownS, body.Description,
	).Scan(&id)
	if err != nil {
		h.logger.Error("alert rule create", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	respondCreated(c, gin.H{"id": id})
}

func (h *RuleHandler) UpdateFromDB(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	id := c.Param("id")
	var body struct {
		Name       string `json:"name"`
		Severity   string `json:"severity"`
		SpeedLimit *int   `json:"speed_limit"`
		CooldownS  *int   `json:"cooldown_s"`
		Enabled    *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.pool.Exec(c.Request.Context(), `
		UPDATE alert_rules SET
		  name       = COALESCE(NULLIF($3,''), name),
		  severity   = COALESCE(NULLIF($4,''), severity),
		  speed_limit = COALESCE($5, speed_limit),
		  cooldown_s = COALESCE($6, cooldown_s),
		  enabled    = COALESCE($7, enabled),
		  updated_at = now()
		WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID, body.Name, body.Severity, body.SpeedLimit, body.CooldownS, body.Enabled)
	if err != nil {
		h.logger.Error("alert rule update", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *RuleHandler) DeleteFromDB(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	h.pool.Exec(c.Request.Context(),
		`UPDATE alert_rules SET deleted_at=now() WHERE id=$1 AND tenant_id=$2`,
		c.Param("id"), tenantID)
	c.JSON(http.StatusNoContent, nil)
}

// GetFromDB returns a single rule with its full conditions.
func (h *RuleHandler) GetFromDB(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, alert_type, severity, conditions, speed_limit,
		       cooldown_s, enabled, description, trigger_count, last_triggered
		FROM alert_rules WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		c.Param("id"), tenantID)
	if err != nil || !rows.Next() {
		respondError(c, http.StatusNotFound, "rule not found")
		return
	}
	defer rows.Close()
	vals, err := rows.Values()
	if err != nil {
		h.logger.Error("alert rule get: scan", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	fields := []string{"id", "name", "alert_type", "severity", "conditions",
		"speed_limit", "cooldown_s", "enabled", "description", "trigger_count", "last_triggered"}
	m := make(map[string]any)
	for i, f := range fields {
		if i < len(vals) {
			m[f] = vals[i]
		}
	}
	respondOK(c, m)
}
