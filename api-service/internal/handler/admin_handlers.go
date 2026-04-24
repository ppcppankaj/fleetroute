package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgauth "gpsgo/pkg/auth"
)

// ── UserHandler ────────────────────────────────────────────────────────────────

type UserHandler struct{}

func NewUserHandler() *UserHandler { return &UserHandler{} }

func (h *UserHandler) ListUsers(c *gin.Context) {
	respondOK(c, []gin.H{})
}
func (h *UserHandler) GetUser(c *gin.Context) {
	respondOK(c, gin.H{"id": c.Param("id")})
}
func (h *UserHandler) CreateUser(c *gin.Context) {
	respondCreated(c, gin.H{"id": "new-user"})
}
func (h *UserHandler) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}
func (h *UserHandler) DeleteUser(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// ── RoleHandler (extended) ─────────────────────────────────────────────────────

func (h *RuleHandler) ListRoles(c *gin.Context) {
	respondOK(c, []gin.H{
		{"name": "admin", "permissions": []string{"*"}},
		{"name": "manager", "permissions": []string{"tracking.*", "vehicles.*", "drivers.*", "reports.*", "alerts.*"}},
		{"name": "dispatcher", "permissions": []string{"tracking.view", "alerts.view", "alerts.acknowledge"}},
		{"name": "driver", "permissions": []string{"tracking.view"}},
		{"name": "viewer", "permissions": []string{"tracking.view", "reports.view"}},
	})
}

// ── TenantHandler ──────────────────────────────────────────────────────────────

type TenantHandler struct{}

func NewTenantHandler() *TenantHandler { return &TenantHandler{} }

func (h *TenantHandler) GetSettings(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	respondOK(c, gin.H{
		"tenant_id":       tenantID,
		"company_name":    "Fleet Corp",
		"timezone":        "Asia/Kolkata",
		"locale":          "en",
		"primary_color":   "#1a73e8",
		"secondary_color": "#34a853",
	})
}

func (h *TenantHandler) UpdateSettings(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (h *TenantHandler) GetBranding(c *gin.Context) {
	respondOK(c, gin.H{
		"logo_url":        "",
		"favicon_url":     "",
		"primary_color":   "#1a73e8",
		"secondary_color": "#34a853",
		"custom_domain":   "",
	})
}

func (h *TenantHandler) UpdateBranding(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (h *TenantHandler) GetFeatureFlags(c *gin.Context) {
	respondOK(c, []gin.H{
		{"feature": "video_telematics", "enabled": false},
		{"feature": "ai_routing", "enabled": false},
		{"feature": "cold_chain", "enabled": false},
		{"feature": "adas", "enabled": false},
	})
}

func (h *TenantHandler) SetFeatureFlag(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// ── AuditHandler ──────────────────────────────────────────────────────────────

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler { return &AuditHandler{} }

func (h *AuditHandler) List(c *gin.Context) {
	respondOK(c, []gin.H{})
}

// ── BillingHandler ────────────────────────────────────────────────────────────

type BillingHandler struct{}

func NewBillingHandler() *BillingHandler { return &BillingHandler{} }

func (h *BillingHandler) GetPlans(c *gin.Context) {
	respondOK(c, []gin.H{
		{
			"id": "starter", "name": "Starter", "price": 999, "currency": "INR",
			"billing_cycle": "monthly", "features": gin.H{"max_devices": 25, "max_users": 5},
		},
		{
			"id": "pro", "name": "Professional", "price": 2999, "currency": "INR",
			"billing_cycle": "monthly", "features": gin.H{"max_devices": 100, "max_users": 20},
		},
		{
			"id": "enterprise", "name": "Enterprise", "price": 0, "currency": "INR",
			"billing_cycle": "annual", "features": gin.H{"max_devices": -1, "max_users": -1},
		},
	})
}

func (h *BillingHandler) GetSubscription(c *gin.Context) {
	respondOK(c, gin.H{
		"plan":   "pro",
		"status": "active",
		"current_period_end": "2026-05-23",
		"auto_renew": true,
	})
}

func (h *BillingHandler) GetInvoices(c *gin.Context) {
	respondOK(c, []gin.H{})
}

func (h *BillingHandler) GetUsage(c *gin.Context) {
	respondOK(c, gin.H{
		"devices_count":  42,
		"api_calls":      125430,
		"storage_gb":     12.4,
		"video_hours":    0,
	})
}

// ── DeviceManagementHandler ────────────────────────────────────────────────────
// (device provisioning, OTA, health — separate from ingestion-service)

type DeviceManagementHandler struct{}

func NewDeviceManagementHandler() *DeviceManagementHandler { return &DeviceManagementHandler{} }

func (h *DeviceManagementHandler) ListProtocols(c *gin.Context) {
	respondOK(c, []gin.H{
		{"name": "Teltonika",  "port": 6030, "models": []string{"FMB920", "FMB140", "FMT100"}},
		{"name": "Concox",    "port": 6000, "models": []string{"GT06E", "GV75", "WeTrack2"}},
		{"name": "Queclink",  "port": 6020, "models": []string{"GL300", "GL500", "GV65"}},
		{"name": "Ruptela",   "port": 6050, "models": []string{"FM Eco4+", "Pro4"}},
		{"name": "Custom",    "port": 6100, "models": []string{}},
	})
}

func (h *DeviceManagementHandler) GetHealth(c *gin.Context) {
	respondOK(c, gin.H{
		"device_id":        c.Param("id"),
		"connected":        true,
		"signal_strength":  87,
		"battery_percent":  92,
		"firmware_version": "2.1.5",
		"last_seen":        "2026-04-23T10:00:00Z",
	})
}

func (h *DeviceManagementHandler) TriggerOTA(c *gin.Context) {
	respondCreated(c, gin.H{"job_id": "ota-xyz", "status": "queued"})
}
