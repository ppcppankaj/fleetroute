package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gpsgo/services/m14-billing/internal/config"
	"gpsgo/services/m14-billing/internal/handler"
	"gpsgo/services/m14-billing/internal/repository"
	"gpsgo/services/m14-billing/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	repo := repository.New(pool)
	svc := service.New(repo, cfg.KafkaBrokers)
	defer svc.Close()

	h := handler.New(svc)
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"m14-billing"}`))
	})
	r.Handle("/metrics", promhttp.Handler())
	r.Route("/api/billing", func(r chi.Router) {
		h.RegisterRoutes(r)
	})

	server := &http.Server{
		Addr: ":" + cfg.Port, Handler: r,
		ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutCtx)
	}()

	log.Printf("m14-billing listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
