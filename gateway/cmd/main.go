package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	sharedkafka "gpsgo/shared/kafka"
	"gpsgo/shared/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

type serviceRoute struct {
	ID  string
	URL string
}

type gateway struct {
	breakers map[string]*gobreaker.CircuitBreaker
	rdb      *redis.Client
	writer   *kafkago.Writer
	pubKey   *rsa.PublicKey
	timeout  time.Duration
}

func main() {
	ctx := context.Background()
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(mustReadFile(env("JWT_PUBLIC_KEY_PATH", "secrets\\jwt_public.pem")))
	if err != nil {
		log.Fatalf("parse jwt public key: %v", err)
	}

	gw := &gateway{
		breakers: buildBreakers(),
		rdb: redis.NewClient(&redis.Options{
			Addr: env("REDIS_ADDR", "localhost:6379"),
		}),
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")...),
			Async:        true,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
		},
		pubKey:  pubKey,
		timeout: 30 * time.Second,
	}
	defer func() {
		_ = gw.rdb.Close()
		_ = gw.writer.Close()
	}()
	_ = gw.rdb.Ping(ctx).Err()

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(gw.ipRateLimit)
	r.Use(gw.requestAudit)

	r.Post("/auth/login", gw.proxy("m11", "http://m11-users-access:4011", false, false))
	r.Post("/auth/refresh", gw.proxy("m11", "http://m11-users-access:4011", false, false))
	r.Mount("/api/tracking", chiRouter(gw.proxy("m01", "http://m01-live-tracking:4001", true, false)))
	r.Mount("/api/routes", chiRouter(gw.proxy("m02", "http://m02-routes-trips:4002", true, false)))
	r.Mount("/api/geofencing", chiRouter(gw.proxy("m03", "http://m03-geofencing:4003", true, false)))
	r.Mount("/api/alerts", chiRouter(gw.proxy("m04", "http://m04-alerts:4004", true, false)))
	r.Mount("/api/reports", chiRouter(gw.proxy("m05", "http://m05-reports:4005", true, false)))
	r.Mount("/api/vehicles", chiRouter(gw.proxy("m06", "http://m06-vehicles:4006", true, false)))
	r.Mount("/api/drivers", chiRouter(gw.proxy("m07", "http://m07-drivers:4007", true, false)))
	r.Mount("/api/maintenance", chiRouter(gw.proxy("m08", "http://m08-maintenance:4008", true, false)))
	r.Mount("/api/fuel", chiRouter(gw.proxy("m09", "http://m09-fuel:4009", true, false)))
	r.Mount("/api/tenants", chiRouter(gw.proxy("m10", "http://m10-multi-tenant:4010", true, false)))
	r.Mount("/api/users", chiRouter(gw.proxy("m11", "http://m11-users-access:4011", true, false)))
	r.Mount("/api/devices", chiRouter(gw.proxy("m12", "http://m12-devices:4012", true, false)))
	r.Mount("/api/security", chiRouter(gw.proxy("m13", "http://m13-security:4013", true, false)))
	r.Mount("/api/billing", chiRouter(gw.proxy("m14", "http://m14-billing:4014", true, false)))
	r.Mount("/api/admin", chiRouter(gw.proxy("m15", "http://m15-admin-panel:4015", true, true)))
	r.Mount("/api/activity", chiRouter(gw.proxy("m16", "http://m16-activity-log:4016", true, false)))
	r.Mount("/api/roadmap", chiRouter(gw.proxy("m17", "http://m17-roadmap:4017", true, false)))
	r.Get("/health/{service}", gw.healthAggregator)

	addr := ":" + env("PORT", "3000")
	log.Printf("gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func buildBreakers() map[string]*gobreaker.CircuitBreaker {
	names := []string{"m01", "m02", "m03", "m04", "m05", "m06", "m07", "m08", "m09", "m10", "m11", "m12", "m13", "m14", "m15", "m16", "m17"}
	out := make(map[string]*gobreaker.CircuitBreaker, len(names))
	for _, name := range names {
		service := name
		out[service] = gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        service,
			MaxRequests: 1,
			Interval:    10 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures >= 5 },
		})
	}
	return out
}

func (g *gateway) proxy(serviceID, target string, auth, adminOnly bool) http.HandlerFunc {
	targetURL := mustURL(target)
	upstream := httputil.NewSingleHostReverseProxy(targetURL)
	upstream.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		response.Error(w, http.StatusBadGateway, "UPSTREAM_ERROR", map[string]any{"service": serviceID})
	}
	upstream.Transport = &http.Transport{
		ResponseHeaderTimeout: g.timeout,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if auth {
			claims, err := g.verifyJWT(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", nil)
				return
			}
			tenantID := fmt.Sprintf("%v", claims["tenantId"])
			userID := fmt.Sprintf("%v", claims["userId"])
			role := fmt.Sprintf("%v", claims["role"])
			if tenantID == "" || userID == "" {
				response.Error(w, http.StatusUnauthorized, "INVALID_CLAIMS", nil)
				return
			}
			if adminOnly && role != "SUPER_ADMIN" {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", nil)
				return
			}
			if err := g.tenantRateLimit(r.Context(), tenantID); err != nil {
				response.Error(w, http.StatusTooManyRequests, "TENANT_RATE_LIMITED", nil)
				return
			}
			r.Header.Set("X-Tenant-Id", tenantID)
			r.Header.Set("X-User-Id", userID)
			r.Header.Set("X-User-Role", role)
		}
		requestID := requestID(r)
		r.Header.Set("X-Request-Id", requestID)

		cb := g.breakers[serviceID]
		_, err := cb.Execute(func() (any, error) {
			ctx, cancel := context.WithTimeout(r.Context(), g.timeout)
			defer cancel()
			upstream.ServeHTTP(w, r.WithContext(ctx))
			if statusFromWriter(w) >= 500 {
				return nil, errors.New("upstream 5xx")
			}
			return nil, nil
		})
		if err != nil {
			if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
				response.Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", map[string]any{"service": serviceID})
				return
			}
			response.Error(w, http.StatusBadGateway, "UPSTREAM_ERROR", map[string]any{"service": serviceID})
		}
	}
}

func (g *gateway) verifyJWT(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("invalid method")
		}
		return g.pubKey, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (g *gateway) ipRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		key := "rl:ip:" + ip + ":" + time.Now().Format("200601021504")
		count, err := g.rdb.Incr(r.Context(), key).Result()
		if err == nil && count == 1 {
			_ = g.rdb.Expire(r.Context(), key, 65*time.Second).Err()
		}
		if count > 200 {
			response.Error(w, http.StatusTooManyRequests, "IP_RATE_LIMITED", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (g *gateway) tenantRateLimit(ctx context.Context, tenantID string) error {
	key := "rl:tenant:" + tenantID + ":" + time.Now().Format("200601021504")
	count, err := g.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		_ = g.rdb.Expire(ctx, key, 65*time.Second).Err()
	}
	if count > 2000 {
		return errors.New("tenant rate limited")
	}
	return nil
}

func (g *gateway) requestAudit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		go func(req *http.Request) {
			body, _ := json.Marshal(map[string]any{
				"request_id": requestID(req),
				"path":       req.URL.Path,
				"method":     req.Method,
				"tenant_id":  req.Header.Get("X-Tenant-Id"),
				"user_id":    req.Header.Get("X-User-Id"),
				"ip":         req.RemoteAddr,
				"created_at": time.Now().UTC(),
			})
			_ = g.writer.WriteMessages(context.Background(), kafkago.Message{
				Topic: sharedkafka.TopicGatewayRequest,
				Key:   []byte(requestID(req)),
				Value: body,
			})
		}(r.Clone(context.Background()))
	})
}

func (g *gateway) healthAggregator(w http.ResponseWriter, r *http.Request) {
	services := []serviceRoute{
		{"m01", "http://m01-live-tracking:4001/health"},
		{"m02", "http://m02-routes-trips:4002/health"},
		{"m03", "http://m03-geofencing:4003/health"},
		{"m04", "http://m04-alerts:4004/health"},
		{"m05", "http://m05-reports:4005/health"},
		{"m06", "http://m06-vehicles:4006/health"},
		{"m07", "http://m07-drivers:4007/health"},
		{"m08", "http://m08-maintenance:4008/health"},
		{"m09", "http://m09-fuel:4009/health"},
		{"m10", "http://m10-multi-tenant:4010/health"},
		{"m11", "http://m11-users-access:4011/health"},
		{"m12", "http://m12-devices:4012/health"},
		{"m13", "http://m13-security:4013/health"},
		{"m14", "http://m14-billing:4014/health"},
		{"m15", "http://m15-admin-panel:4015/health"},
		{"m16", "http://m16-activity-log:4016/health"},
		{"m17", "http://m17-roadmap:4017/health"},
	}
	type status struct {
		Service string `json:"service"`
		Up      bool   `json:"up"`
	}
	out := make([]status, 0, len(services))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, s := range services {
		svc := s
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, svc.URL, nil)
			resp, err := http.DefaultClient.Do(req)
			up := err == nil && resp.StatusCode < 500
			mu.Lock()
			out = append(out, status{Service: svc.ID, Up: up})
			mu.Unlock()
		}()
	}
	wg.Wait()
	response.JSON(w, http.StatusOK, out)
}

func chiRouter(handler http.HandlerFunc) http.Handler {
	r := chi.NewRouter()
	r.Handle("/*", handler)
	return r
}

func splitCSV(input string) []string {
	values := strings.Split(input, ",")
	out := make([]string, 0, len(values))
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func requestID(r *http.Request) string {
	if v := r.Header.Get("X-Request-Id"); v != "" {
		return v
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func statusFromWriter(http.ResponseWriter) int { return 200 }

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustReadFile(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read file %s: %v", path, err)
	}
	return b
}

func mustURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid url %s: %v", raw, err)
	}
	return u
}
