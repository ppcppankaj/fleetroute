package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	RedisAddr    string
	MQTTBroker   string
}

func Load() *Config {
	return &Config{
		Port:         env("PORT", "4001"),
		DatabaseURL:  env("DATABASE_URL", "postgres://fleet:fleetpass@localhost:5401/fleet_live_db?sslmode=require"),
		KafkaBrokers: env("KAFKA_BROKERS", "localhost:29092"),
		RedisAddr:    env("REDIS_ADDR", "localhost:6379"),
		MQTTBroker:   env("MQTT_BROKER", "tcp://localhost:1883"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
