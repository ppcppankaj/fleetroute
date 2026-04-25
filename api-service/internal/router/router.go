// Package router configures all API routes and middleware.
package router

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	pkgauth "gpsgo/pkg/auth"
	"gpsgo/api-service/internal/handler"
	"gpsgo/api-service/internal/middleware"
)

// New builds and returns the configured Gin engine.
func New(
	pool *pgxpool.Pool,
	rdb *redis.Client,
	authMgr *pkgauth.Manager,
	logger *zap.Logger,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// ── CORS — read allowed origins from environment (no wildcard default) ────
	allowedOrigins := parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	// ── Global middleware ─────────────────────────────────────────────────────
	r.Use(middleware.RequestLogger(logger))
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// ── Health endpoints (no auth) ────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "api"})
	})
	r.GET("/ready", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler     := handler.NewAuthHandler(pool, rdb, authMgr, logger)
	deviceHandler   := handler.NewDeviceHandler(pool, rdb, logger)
	vehicleHandler  := handler.NewVehicleHandler(pool, logger)
	geofenceHandler := handler.NewGeofenceHandler(pool, logger)
	alertHandler    := handler.NewAlertHandler(pool, logger)
	ruleHandler     := handler.NewRuleHandler(pool, logger)
	reportHandler   := handler.NewReportHandler(pool, logger)
	driverHandler   := handler.NewDriverHandler(pool, logger)
	userHandler     := handler.NewUserHandler()
	tenantHandler   := handler.NewTenantHandler()
	billingHandler  := handler.NewBillingHandler()
	auditHandler    := handler.NewAuditHandler()
	devMgmtHandler  := handler.NewDeviceManagementHandler()
	fuelHandler     := handler.NewFuelHandler(pool, logger)

	// ── v1 routes ─────────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected routes — require valid JWT + RLS transaction + rate limit
	protected := v1.Group("")
	protected.Use(pkgauth.Middleware(authMgr))
	protected.Use(middleware.RLS(pool, logger))
	protected.Use(middleware.RateLimit(rdb, logger))
	{
		// ── Devices (C1: RBAC enforced) ───────────────────────────────────────
		// Read operations: all authenticated roles
		// Write operations: require devices:write
		// Delete: require devices:write
		// OTA push: require commands:send (privileged operation)
		devices := protected.Group("/devices")
		devices.GET("", deviceHandler.List)
		devices.POST("",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			deviceHandler.Create,
		)
		devices.GET("/:id", deviceHandler.Get)
		devices.PUT("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			deviceHandler.Update,
		)
		devices.DELETE("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			deviceHandler.Delete,
		)
		devices.GET("/:id/live", deviceHandler.Live)
		devices.GET("/:id/history", deviceHandler.History)
		devices.GET("/:id/trips", deviceHandler.Trips)
		devices.GET("/:id/telemetry", deviceHandler.Telemetry)

		// ── Vehicles ──────────────────────────────────────────────────────────
		vehicles := protected.Group("/vehicles")
		vehicles.GET("", vehicleHandler.List)
		vehicles.POST("",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			vehicleHandler.Create,
		)
		vehicles.GET("/:id", vehicleHandler.Get)
		vehicles.PUT("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			vehicleHandler.Update,
		)
		vehicles.DELETE("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			vehicleHandler.Delete,
		)
		vehicles.POST("/:id/command",
			pkgauth.RequirePermission(pkgauth.PermSendCommands),
			vehicleHandler.SendCommand,
		)

		// ── Drivers ───────────────────────────────────────────────────────────
		drivers := protected.Group("/drivers")
		drivers.GET("", driverHandler.List)
		drivers.POST("",
			pkgauth.RequirePermission(pkgauth.PermManageUsers),
			driverHandler.Create,
		)
		drivers.GET("/:id", driverHandler.Get)
		drivers.GET("/:id/score", driverHandler.Score)

		// ── Geofences ─────────────────────────────────────────────────────────
		geofences := protected.Group("/geofences")
		geofences.GET("", geofenceHandler.List)
		geofences.POST("",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			geofenceHandler.Create,
		)
		geofences.GET("/:id", geofenceHandler.Get)
		geofences.PUT("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			geofenceHandler.Update,
		)
		geofences.DELETE("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			geofenceHandler.Delete,
		)
		geofences.GET("/:id/events", geofenceHandler.Events)

		// ── Alerts ────────────────────────────────────────────────────────────
		alerts := protected.Group("/alerts")
		alerts.GET("", alertHandler.List)
		alerts.POST("/:id/acknowledge", alertHandler.Acknowledge)

		// ── Rules (DB-backed CRUD) ────────────────────────────────────────────
		rules := protected.Group("/rules")
		rules.GET("", ruleHandler.ListFromDB)
		rules.POST("",
			pkgauth.RequirePermission(pkgauth.PermWriteRules),
			ruleHandler.CreateFromDB,
		)
		rules.GET("/:id", ruleHandler.GetFromDB)
		rules.PUT("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteRules),
			ruleHandler.UpdateFromDB,
		)
		rules.DELETE("/:id",
			pkgauth.RequirePermission(pkgauth.PermWriteRules),
			ruleHandler.DeleteFromDB,
		)
		rules.GET("/templates", ruleHandler.Templates)

		// ── Fuel Management (M09) ─────────────────────────────────────────────
		fuel := protected.Group("/fuel")
		fuel.GET("/logs", fuelHandler.ListFuelLogs)
		fuel.POST("/logs",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			fuelHandler.CreateFuelLog,
		)
		fuel.GET("/summary", fuelHandler.FuelSummary)
		fuel.GET("/anomalies", fuelHandler.ListAnomalies)
		fuel.POST("/anomalies/:id/confirm",
			pkgauth.RequirePermission(pkgauth.PermWriteDevices),
			fuelHandler.ConfirmAnomaly,
		)

		// ── Video Telematics (M15) ────────────────────────────────────────────
		video := protected.Group("/video")
		videoHandler := handler.NewVideoHandler(pool, logger)
		video.GET("/events", videoHandler.ListEvents)
		video.POST("/devices/:deviceId/snapshot",
			pkgauth.RequirePermission(pkgauth.PermSendCommands),
			videoHandler.TriggerSnapshot,
		)
		video.GET("/devices/:deviceId/livestream", videoHandler.GetLiveStreamCredentials)

		// ── Reports ───────────────────────────────────────────────────────────
		reports := protected.Group("/reports")
		reports.GET("/trips", reportHandler.Trips)
		reports.GET("/fuel", reportHandler.Fuel)
		reports.GET("/driver-behavior", reportHandler.DriverBehavior)
		reports.GET("/geofence-violations", reportHandler.GeofenceViolations)

		// ── Users & Roles (M11) — require user management permission ─────────
		users := protected.Group("/users")
		users.Use(pkgauth.RequirePermission(pkgauth.PermManageUsers))
		users.GET("", userHandler.ListUsers)
		users.POST("", userHandler.CreateUser)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
		protected.GET("/roles", ruleHandler.ListRoles)

		// ── Tenant settings & branding (M10) — require tenant admin ──────────
		tenant := protected.Group("/tenant")
		tenant.Use(pkgauth.RequirePermission(pkgauth.PermManageTenants))
		tenant.GET("/settings", tenantHandler.GetSettings)
		tenant.PUT("/settings", tenantHandler.UpdateSettings)
		tenant.GET("/branding", tenantHandler.GetBranding)
		tenant.PUT("/branding", tenantHandler.UpdateBranding)
		tenant.GET("/feature-flags", tenantHandler.GetFeatureFlags)
		tenant.PUT("/feature-flags/:feature", tenantHandler.SetFeatureFlag)

		// ── Billing (M14) — read: all; write: tenant admin ───────────────────
		billing := protected.Group("/billing")
		billing.GET("/plans", billingHandler.GetPlans)
		billing.GET("/subscription", billingHandler.GetSubscription)
		billing.GET("/invoices", billingHandler.GetInvoices)
		billing.GET("/usage", billingHandler.GetUsage)

		// ── Audit log (M16) — fleet manager and above ────────────────────────
		protected.GET("/audit",
			pkgauth.RequirePermission(pkgauth.PermReadReports),
			auditHandler.List,
		)

		// ── Device management extras (M12) ────────────────────────────────────
		protected.GET("/device-protocols", devMgmtHandler.ListProtocols)
		protected.GET("/devices/:id/health", devMgmtHandler.GetHealth)
		protected.POST("/devices/:id/ota",
			pkgauth.RequirePermission(pkgauth.PermSendCommands),
			devMgmtHandler.TriggerOTA,
		)
	}

	return r
}

// parseCORSOrigins splits a comma-separated origin allowlist from the
// CORS_ALLOWED_ORIGINS environment variable. Returns a slice with a single
// empty string (which gin-cors treats as "no origins allowed") if unset,
// ensuring the service is restrictive by default.
func parseCORSOrigins(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
