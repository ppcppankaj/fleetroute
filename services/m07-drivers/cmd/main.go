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
	
	"gpsgo/services/m07-drivers/internal/config"
	"gpsgo/services/m07-drivers/internal/handler"
	"gpsgo/services/m07-drivers/internal/kafka"
	"gpsgo/services/m07-drivers/internal/repository"
	"gpsgo/services/m07-drivers/internal/service"
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

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	repo := repository.New(pool)
	svc := service.New(repo, producer)
	h := handler.New(svc)

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, svc)
	go consumer.Start(ctx)
	defer consumer.Close()

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"m07-drivers"}`))
	})
	r.Handle("/metrics", promhttp.Handler())
	
	r.Route("/api/drivers", func(r chi.Router) {
		h.RegisterRoutes(r)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("m07-drivers listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
