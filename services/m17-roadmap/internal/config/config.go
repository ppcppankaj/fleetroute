package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() *Config {
	return &Config{
		Port:        env("PORT", "4017"),
		DatabaseURL: env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5417/fleet_roadmap_db?sslmode=require"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
