package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Upstream targets
	apiURL    := mustParseURL(getenv("API_SERVICE_URL",    "http://localhost:8080"))
	wsURL     := mustParseURL(getenv("WS_SERVICE_URL",     "http://localhost:8081"))
	mntURL    := mustParseURL(getenv("MAINT_SERVICE_URL",  "http://localhost:8084"))
	rptURL    := mustParseURL(getenv("REPORT_SERVICE_URL", "http://localhost:8085"))

	apiProxy  := newReverseProxy(apiURL)
	wsProxy   := newReverseProxy(wsURL)
	mntProxy  := newReverseProxy(mntURL)
	rptProxy  := newReverseProxy(rptURL)

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "gateway", "ts": time.Now().Unix()})
	})

	// REST API  →  api-service
	r.Any("/api/*path", func(c *gin.Context) {
		apiProxy.ServeHTTP(c.Writer, c.Request)
	})

	// Maintenance API  →  maintenance-service
	r.Any("/api/v1/maintenance/*path", func(c *gin.Context) {
		mntProxy.ServeHTTP(c.Writer, c.Request)
	})

	// Report API  →  report-service
	r.Any("/api/v1/reports/*path", func(c *gin.Context) {
		rptProxy.ServeHTTP(c.Writer, c.Request)
	})

	// WebSocket  →  websocket-service
	r.GET("/ws", func(c *gin.Context) {
		wsProxy.ServeHTTP(c.Writer, c.Request)
	})

	port := getenv("PORT", "8000")
	log.Info("gateway listening", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatal("server", zap.Error(err))
	}
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Server")
		return nil
	}
	return proxy
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := strings.Split(getenv("CORS_ORIGINS", "*"), ",")
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Tenant-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic("invalid URL: " + raw)
	}
	return u
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
