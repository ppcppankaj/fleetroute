package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
}

func Load() *Config {
	return &Config{
		Port:         env("PORT", "4016"),
		DatabaseURL:  env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5416/fleet_activity_db?sslmode=require"),
		KafkaBrokers: env("KAFKA_BROKERS", "localhost:29092"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
