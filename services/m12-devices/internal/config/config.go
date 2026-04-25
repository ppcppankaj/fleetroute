package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	RedisAddr    string
}

func Load() *Config {
	return &Config{
		Port:         env("PORT", "4012"),
		DatabaseURL:  env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5412/fleet_devices_db?sslmode=require"),
		KafkaBrokers: env("KAFKA_BROKERS", "localhost:29092"),
		RedisAddr:    env("REDIS_ADDR", "localhost:6379"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
