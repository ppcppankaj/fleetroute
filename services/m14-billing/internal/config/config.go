package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	StripeKey    string
	WebhookSecret string
}

func Load() *Config {
	return &Config{
		Port:          env("PORT", "4014"),
		DatabaseURL:   env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5414/fleet_billing_db?sslmode=disable"),
		KafkaBrokers:  env("KAFKA_BROKERS", "localhost:29092"),
		StripeKey:     env("STRIPE_SECRET_KEY", ""),
		WebhookSecret: env("STRIPE_WEBHOOK_SECRET", ""),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
