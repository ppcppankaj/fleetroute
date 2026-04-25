package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	pkgauth "gpsgo/pkg/auth"
)

// ── UserHandler ───────────────────────────────────────────────────────────────

type UserHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewUserHandler(pool *pgxpool.Pool, logger *zap.Logger) *UserHandler {
	return &UserHandler{pool: pool, logger: logger}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, email, name, role, phone, is_active, created_at
		 FROM users WHERE tenant_id=$1 AND deleted_at IS NULL ORDER BY name`,
		tenantID,
	)
	if err != nil {
		h.logger.Error("users list", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	var users []map[string]any
	for rows.Next() {
		var id, email, name, role string
		var phone *string
		var isActive bool
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &name, &role, &phone, &isActive, &createdAt); err != nil {
			h.logger.Error("users list scan", zap.Error(err))
			continue
		}
		users = append(users, map[string]any{
			"id": id, "email": email, "name": name, "role": role,
			"phone": phone, "is_active": isActive, "created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("users list rows", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if users == nil {
		users = []map[string]any{}
	}
	respondOK(c, users)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	id := c.Param("id")
	var userID, email, name, role string
	var phone *string
	var isActive bool
	var createdAt time.Time
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, email, name, role, phone, is_active, created_at
		 FROM users WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID,
	).Scan(&userID, &email, &name, &role, &phone, &isActive, &createdAt)
	if err != nil {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}
	respondOK(c, map[string]any{
		"id": userID, "email": email, "name": name, "role": role,
		"phone": phone, "is_active": isActive, "created_at": createdAt,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("bcrypt", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "internal error")
		return
	}
	var id string
	err = h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO users (tenant_id, email, name, role, password_hash, phone)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		tenantID, body.Email, body.Name, body.Role, string(hash), body.Phone,
	).Scan(&id)
	if err != nil {
		h.logger.Error("user create", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	respondCreated(c, gin.H{"id": id})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	id := c.Param("id")
	var body struct {
		Name     string `json:"name"`
		Role     string `json:"role"`
		Phone    string `json:"phone"`
		IsActive *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE users SET
		   name     = COALESCE(NULLIF($3,''), name),
		   role     = COALESCE(NULLIF($4,''), role),
		   phone    = COALESCE(NULLIF($5,''), phone),
		   is_active = COALESCE($6, is_active),
		   updated_at = now()
		 WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID, body.Name, body.Role, body.Phone, body.IsActive,
	)
	if err != nil {
		h.logger.Error("user update", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if tag.RowsAffected() == 0 {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE users SET deleted_at=now() WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		c.Param("id"), tenantID,
	)
	if err != nil {
		h.logger.Error("user delete", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if tag.RowsAffected() == 0 {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── RoleHandler ────────────────────────────────────────────────────────────────

func (h *RuleHandler) ListRoles(c *gin.Context) {
	// Return role names that match the actual JWT/DB role values
	respondOK(c, []gin.H{
		{
			"name": "tenant_admin",
			"permissions": []string{
				"devices:read", "devices:write", "trips:read", "alerts:read",
				"alerts:ack", "rules:write", "reports:read", "users:manage", "commands:send",
			},
		},
		{
			"name": "fleet_manager",
			"permissions": []string{
				"devices:read", "trips:read", "alerts:read",
				"alerts:ack", "rules:write", "reports:read", "commands:send",
			},
		},
		{
			"name": "dispatcher",
			"permissions": []string{
				"devices:read", "trips:read", "alerts:read", "alerts:ack", "reports:read",
			},
		},
		{
			"name":        "driver",
			"permissions": []string{"devices:read", "trips:read"},
		},
	})
}

// ── TenantHandler ─────────────────────────────────────────────────────────────

type TenantHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewTenantHandler(pool *pgxpool.Pool, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{pool: pool, logger: logger}
}

func (h *TenantHandler) GetSettings(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var name, slug, plan, timezone string
	var settingsRaw []byte
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT name, slug, plan, COALESCE(settings->>'timezone','Asia/Kolkata'), settings
		 FROM tenants WHERE id=$1 AND deleted_at IS NULL`,
		tenantID,
	).Scan(&name, &slug, &plan, &timezone, &settingsRaw)
	if err != nil {
		h.logger.Error("tenant settings get", zap.Error(err))
		respondError(c, http.StatusNotFound, "tenant not found")
		return
	}
	var settings map[string]any
	json.Unmarshal(settingsRaw, &settings)
	if settings == nil {
		settings = map[string]any{}
	}
	settings["tenant_id"] = tenantID
	settings["company_name"] = name
	settings["slug"] = slug
	settings["plan"] = plan
	settings["timezone"] = timezone
	respondOK(c, settings)
}

func (h *TenantHandler) UpdateSettings(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	b, _ := json.Marshal(body)
	_, err := h.pool.Exec(c.Request.Context(),
		`UPDATE tenants SET settings = settings || $2::jsonb, updated_at=now()
		 WHERE id=$1 AND deleted_at IS NULL`,
		tenantID, string(b),
	)
	if err != nil {
		h.logger.Error("tenant settings update", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *TenantHandler) GetBranding(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var brandingRaw []byte
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COALESCE(settings->'branding', '{}') FROM tenants WHERE id=$1 AND deleted_at IS NULL`,
		tenantID,
	).Scan(&brandingRaw)
	if err != nil {
		respondError(c, http.StatusNotFound, "tenant not found")
		return
	}
	var branding map[string]any
	json.Unmarshal(brandingRaw, &branding)
	if branding == nil {
		branding = map[string]any{}
	}
	respondOK(c, branding)
}

func (h *TenantHandler) UpdateBranding(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	b, _ := json.Marshal(body)
	_, err := h.pool.Exec(c.Request.Context(),
		`UPDATE tenants
		 SET settings = jsonb_set(COALESCE(settings,'{}'), '{branding}', $2::jsonb, true),
		     updated_at = now()
		 WHERE id=$1 AND deleted_at IS NULL`,
		tenantID, string(b),
	)
	if err != nil {
		h.logger.Error("tenant branding update", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *TenantHandler) GetFeatureFlags(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var flagsRaw []byte
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT COALESCE(settings->'feature_flags', '{}') FROM tenants WHERE id=$1 AND deleted_at IS NULL`,
		tenantID,
	).Scan(&flagsRaw)
	if err != nil {
		respondError(c, http.StatusNotFound, "tenant not found")
		return
	}
	var flags map[string]any
	json.Unmarshal(flagsRaw, &flags)
	if flags == nil {
		flags = map[string]any{}
	}
	// Return as array for API consistency
	var result []gin.H
	for feature, enabled := range flags {
		result = append(result, gin.H{"feature": feature, "enabled": enabled})
	}
	if result == nil {
		result = []gin.H{}
	}
	respondOK(c, result)
}

func (h *TenantHandler) SetFeatureFlag(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	feature := c.Param("feature")
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.pool.Exec(c.Request.Context(),
		`UPDATE tenants
		 SET settings = jsonb_set(
		     COALESCE(settings,'{}'),
		     ARRAY['feature_flags', $2],
		     $3::jsonb, true),
		     updated_at = now()
		 WHERE id=$1 AND deleted_at IS NULL`,
		tenantID, feature, json.RawMessage(func() string {
			if body.Enabled { return "true" }
			return "false"
		}()),
	)
	if err != nil {
		h.logger.Error("set feature flag", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── AuditHandler ──────────────────────────────────────────────────────────────

type AuditHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewAuditHandler(pool *pgxpool.Pool, logger *zap.Logger) *AuditHandler {
	return &AuditHandler{pool: pool, logger: logger}
}

func (h *AuditHandler) List(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, user_id, action, resource, resource_id, ip_address, created_at
		 FROM audit_logs WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 200`,
		tenantID,
	)
	if err != nil {
		h.logger.Error("audit list", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	var logs []map[string]any
	for rows.Next() {
		var id, action, resource string
		var userID, resourceID, ipAddress *string
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &action, &resource, &resourceID, &ipAddress, &createdAt); err != nil {
			h.logger.Error("audit scan", zap.Error(err))
			continue
		}
		logs = append(logs, map[string]any{
			"id": id, "user_id": userID, "action": action, "resource": resource,
			"resource_id": resourceID, "ip_address": ipAddress, "created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if logs == nil {
		logs = []map[string]any{}
	}
	respondOK(c, logs)
}

// ── BillingHandler ────────────────────────────────────────────────────────────

type BillingHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewBillingHandler(pool *pgxpool.Pool, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{pool: pool, logger: logger}
}

func (h *BillingHandler) GetPlans(c *gin.Context) {
	rows, err := h.pool.Query(context.Background(),
		`SELECT id, name, price, currency, max_vehicles, max_users, features FROM subscription_plans ORDER BY price`,
	)
	if err != nil {
		h.logger.Error("billing plans", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	var plans []map[string]any
	for rows.Next() {
		var id, name, currency string
		var price float64
		var maxV, maxU int
		var featuresRaw []byte
		if err := rows.Scan(&id, &name, &price, &currency, &maxV, &maxU, &featuresRaw); err != nil {
			continue
		}
		var features map[string]any
		json.Unmarshal(featuresRaw, &features)
		plans = append(plans, map[string]any{
			"id": id, "name": name, "price": price, "currency": currency,
			"max_vehicles": maxV, "max_users": maxU, "features": features,
		})
	}
	if plans == nil {
		plans = []map[string]any{}
	}
	respondOK(c, plans)
}

func (h *BillingHandler) GetSubscription(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var id, planID, status string
	var stripeSub, stripeCus *string
	var periodStart, periodEnd *time.Time
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, plan_id, status, stripe_sub_id, stripe_cus_id,
		        current_period_start, current_period_end
		 FROM subscriptions WHERE tenant_id=$1`,
		tenantID,
	).Scan(&id, &planID, &status, &stripeSub, &stripeCus, &periodStart, &periodEnd)
	if err != nil {
		respondError(c, http.StatusNotFound, "no active subscription found")
		return
	}
	respondOK(c, map[string]any{
		"id": id, "plan": planID, "status": status,
		"stripe_subscription_id": stripeSub,
		"current_period_start":   periodStart,
		"current_period_end":     periodEnd,
	})
}

func (h *BillingHandler) GetInvoices(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, amount, currency, status, due_date, paid_at, pdf_url, created_at
		 FROM invoices WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 50`,
		tenantID,
	)
	if err != nil {
		h.logger.Error("billing invoices", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	var invoices []map[string]any
	for rows.Next() {
		var id, currency, status string
		var amount float64
		var dueDate, paidAt *time.Time
		var pdfURL *string
		var createdAt time.Time
		if err := rows.Scan(&id, &amount, &currency, &status, &dueDate, &paidAt, &pdfURL, &createdAt); err != nil {
			continue
		}
		invoices = append(invoices, map[string]any{
			"id": id, "amount": amount, "currency": currency, "status": status,
			"due_date": dueDate, "paid_at": paidAt, "pdf_url": pdfURL, "created_at": createdAt,
		})
	}
	if invoices == nil {
		invoices = []map[string]any{}
	}
	respondOK(c, invoices)
}

func (h *BillingHandler) GetUsage(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var deviceCount, userCount int
	h.pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM devices WHERE tenant_id=$1 AND deleted_at IS NULL`, tenantID,
	).Scan(&deviceCount)
	h.pool.QueryRow(c.Request.Context(),
		`SELECT COUNT(*) FROM users WHERE tenant_id=$1 AND deleted_at IS NULL`, tenantID,
	).Scan(&userCount)
	respondOK(c, gin.H{
		"devices_count": deviceCount,
		"users_count":   userCount,
	})
}

// ── DeviceManagementHandler ───────────────────────────────────────────────────

type DeviceManagementHandler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDeviceManagementHandler(pool *pgxpool.Pool, logger *zap.Logger) *DeviceManagementHandler {
	return &DeviceManagementHandler{pool: pool, logger: logger}
}

func (h *DeviceManagementHandler) ListProtocols(c *gin.Context) {
	// Protocol config is static — not stored in DB
	respondOK(c, []gin.H{
		{"name": "Teltonika", "port": 6030, "models": []string{"FMB920", "FMB140", "FMT100"}},
		{"name": "Concox",   "port": 6000, "models": []string{"GT06E", "GV75", "WeTrack2"}},
		{"name": "Queclink", "port": 6020, "models": []string{"GL300", "GL500", "GV65"}},
		{"name": "Ruptela",  "port": 6050, "models": []string{"FM Eco4+", "Pro4"}},
		{"name": "AIS140",   "port": 6040, "models": []string{"VLTD Certified"}},
		{"name": "Custom",   "port": 6100, "models": []string{}},
	})
}

func (h *DeviceManagementHandler) GetHealth(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	deviceID := c.Param("id")
	var lastSeenAt *time.Time
	var firmware, protocol *string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT last_seen_at, firmware_ver, protocol
		 FROM devices WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		deviceID, tenantID,
	).Scan(&lastSeenAt, &firmware, &protocol)
	if err != nil {
		respondError(c, http.StatusNotFound, "device not found")
		return
	}
	online := lastSeenAt != nil && time.Since(*lastSeenAt) < 5*time.Minute
	respondOK(c, gin.H{
		"device_id":        deviceID,
		"connected":        online,
		"last_seen":        lastSeenAt,
		"firmware_version": firmware,
		"protocol":         protocol,
	})
}

func (h *DeviceManagementHandler) TriggerOTA(c *gin.Context) {
	// OTA job table does not exist yet — return 501 rather than fake 201
	respondError(c, http.StatusNotImplemented, "OTA job dispatch not yet implemented")
}
