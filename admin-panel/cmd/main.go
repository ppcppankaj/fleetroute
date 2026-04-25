package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/GoAdminGroup/go-admin/adapter/gin"
	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
	"github.com/GoAdminGroup/go-admin/plugins/admin"
	"github.com/GoAdminGroup/go-admin/template"
	_ "github.com/GoAdminGroup/themes/sword"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"gpsgo/admin-panel/pages"
	"gpsgo/admin-panel/tables"
)

func main() {
	// ── Database ─────────────────────────────────────────────────────────────
	dsn := getenv("DATABASE_URL", "postgres://gpsgo:gpsgo@localhost:5432/gpsgo?sslmode=require")
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("database open: %v", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("database ping: %v", err)
	}
	log.Println("✅ database connected")

	// ── GoAdmin Engine ────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	configureTrustedProxies(r)

	eng := engine.Default()
	cfg := config.Config{
		Databases: config.DatabaseList{
			"default": {
				Host:       getenvOr("DB_HOST", "localhost"),
				Port:       getenvOr("DB_PORT", "5432"),
				User:       getenvOr("DB_USER", "gpsgo"),
				Pwd:        getenvOr("DB_PASSWORD", "gpsgo"),
				Name:       getenvOr("DB_NAME", "gpsgo"),
				MaxIdleCon: 10,
				MaxOpenCon: 100,
				Driver:     db.DriverPostgresql,
			},
		},
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Theme:       "sword",
		Language:    "en",
		UrlPrefix:   "admin",
		IndexUrl:    "/custom/dashboard",
		Debug:       os.Getenv("GIN_MODE") != "release",
		ColorScheme: "skin-black",
		Title:       "FleetOS Admin",
		Logo:        template.HTML(`<b>Fleet</b>OS`),
		MiniLogo:    template.HTML(`<b>F</b>`),
		LoginTitle:  "FleetOS Operations",
		LoginLogo:   template.HTML(`<b>Fleet</b>OS`),
		FooterInfo:  template.HTML(`<p>FleetOS Admin Panel © 2026</p>`),
	}

	// Register all data tables
	adminPlugin := admin.NewAdmin(tables.GetGenerators(sqlDB))

	if err = eng.AddConfig(&cfg).
		AddPlugins(adminPlugin).
		Use(r); err != nil {
		log.Fatalf("GoAdmin init: %v", err)
	}

	// Register custom pages
	eng.HTML("GET", "/admin/custom/dashboard", pages.DashboardHandler(sqlDB))
	eng.HTML("GET", "/admin/admin", pages.DashboardHandler(sqlDB))
	eng.HTML("GET", "/admin/custom/packet-inspector", pages.PacketInspectorHandler(sqlDB))
	eng.HTML("GET", "/admin/custom/protocol-stats", pages.ProtocolStatsHandler(sqlDB))
	eng.HTML("GET", "/admin/custom/nats-monitor", pages.NatsMonitorHandler())
	eng.HTML("GET", "/admin/custom/live-map", pages.LiveMapHandler(sqlDB))

	// Health check (not behind admin auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "admin-panel"})
	})

	// Root path convenience redirect for local browser access.
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
	})
	
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/admin/login")
	})

	// Avoid noisy 404s when browsers auto-request favicon.
	r.GET("/favicon.ico", func(c *gin.Context) {


		c.Status(http.StatusNoContent)
	})


	port := getenv("ADMIN_PORT", "8090")
	log.Printf("🚀 GoAdmin panel listening on :%s  →  http://localhost:%s/admin", port, port)
	if err = r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func getenv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func configureTrustedProxies(r *gin.Engine) {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if raw == "" {
		if err := r.SetTrustedProxies(nil); err != nil {
			log.Fatalf("trusted proxies: %v", err)
		}
		return
	}

	parts := strings.Split(raw, ",")
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		proxy := strings.TrimSpace(part)
		if proxy != "" {
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) == 0 {
		if err := r.SetTrustedProxies(nil); err != nil {
			log.Fatalf("trusted proxies: %v", err)
		}
		return
	}

	if err := r.SetTrustedProxies(proxies); err != nil {
		log.Fatalf("trusted proxies: %v", err)
	}
}

func getenvOr(key, defaultVal string) string { return getenv(key, defaultVal) }
