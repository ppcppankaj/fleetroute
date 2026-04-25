package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	JWTSecret    string // HS256 separate secret for super-admin JWT
}

func Load() *Config {
	return &Config{
		Port:         env("PORT", "4015"),
		DatabaseURL:  env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5415/fleet_admin_db?sslmode=require"),
		KafkaBrokers: env("KAFKA_BROKERS", "localhost:29092"),
		JWTSecret:    env("ADMIN_JWT_SECRET", "change-me-in-production"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
