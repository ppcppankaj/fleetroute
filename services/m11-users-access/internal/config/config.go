package config

import "os"

type Config struct {
	Port             string
	DatabaseURL      string
	JWTPrivateKeyPEM []byte
	JWTPublicKeyPEM  []byte
}

func Load() Config {
	privatePath := env("JWT_PRIVATE_KEY_PATH", "secrets\\jwt_private.pem")
	publicPath := env("JWT_PUBLIC_KEY_PATH", "secrets\\jwt_public.pem")
	return Config{
		Port:             env("PORT", "4011"),
		DatabaseURL:      env("DATABASE_URL", "postgres://fleet:pass@localhost:5432/fleet_users_db?sslmode=require"),
		JWTPrivateKeyPEM: mustRead(privatePath),
		JWTPublicKeyPEM:  mustRead(publicPath),
	}
}

func env(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustRead(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}
