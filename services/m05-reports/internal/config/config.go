package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	MinIOEndpoint string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket   string
}

func Load() *Config {
	return &Config{
		Port:           env("PORT", "4005"),
		DatabaseURL:    env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5405/fleet_reports_db?sslmode=require"),
		KafkaBrokers:   env("KAFKA_BROKERS", "localhost:29092"),
		MinIOEndpoint:  env("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: env("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: env("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    env("MINIO_BUCKET", "reports"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
