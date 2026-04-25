package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
}

func Load() *Config {
	return &Config{
		Port:         env("PORT", "4002"),
		DatabaseURL:  env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5402/fleet_trips_db?sslmode=require"),
		KafkaBrokers: env("KAFKA_BROKERS", "localhost:29092"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
