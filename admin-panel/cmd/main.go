package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/chartjs"
	"github.com/GoAdminGroup/themes/sword"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"gpsgo/admin-panel/pages"
	"gpsgo/admin-panel/tables"
)

func main() {
	// ── Database ─────────────────────────────────────────────────────────────
	dsn := getenv("DATABASE_URL", "postgres://gpsgo:gpsgo@localhost:5432/gpsgo?sslmode=disable")
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
	r := gin.Default()

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
		Language:    config.EN,
		UrlPrefix:   "admin",
		IndexUrl:    "/admin",
		Debug:       os.Getenv("GIN_MODE") != "release",
		ColorScheme: sword.ColorschemeSwordDark,
		Title:       "FleetOS Admin",
		Logo:        template.HTML(`<b>Fleet</b>OS`),
		MiniLogo:    template.HTML(`<b>F</b>`),
		LoginTitle:  "FleetOS Operations",
		LoginLogo:   template.HTML(`<b>Fleet</b>OS`),
		FooterInfo:  template.HTML(`<p>FleetOS Admin Panel © 2026</p>`),
	}

	// Register all data tables
	adminPlugin := admin.NewAdmin(tables.GetGenerators(sqlDB))

	// Add chart template
	template.AddComp(chartjs.NewChart())

	if err = eng.AddConfigFromJSON(marshalConfig(cfg)).
		AddPlugins(adminPlugin).
		Use(r); err != nil {
		log.Fatalf("GoAdmin init: %v", err)
	}

	// Register custom pages
	eng.HTML("GET", "/admin/custom/dashboard", pages.DashboardHandler(sqlDB))
	eng.HTML("GET", "/admin/custom/packet-inspector", pages.PacketInspectorHandler(sqlDB))
	eng.HTML("GET", "/admin/custom/protocol-stats", pages.ProtocolStatsHandler(sqlDB))
	eng.HTML("GET", "/admin/custom/nats-monitor", pages.NatsMonitorHandler())
	eng.HTML("GET", "/admin/custom/live-map", pages.LiveMapHandler(sqlDB))

	// Health check (not behind admin auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "admin-panel"})
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

func getenvOr(key, defaultVal string) string { return getenv(key, defaultVal) }

func marshalConfig(cfg config.Config) []byte {
	b, err := json.Marshal(cfg)
	if err != nil {
		log.Fatalf("marshal config: %v", err)
	}
	return b
}

// Suppress unused import warning for fmt
var _ = fmt.Sprintf
