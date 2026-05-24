package config

import (
	"os"
	"time"
)

type Config struct {
	Port      string
	Host      string
	DBPath    string
	JWTSecret string
	JWTExpiry time.Duration
	BaseURL   string
}

func Load() *Config {
	return &Config{
		Port:      envOrDefault("PORT", "8080"),
		Host:      envOrDefault("HOST", "0.0.0.0"),
		DBPath:    envOrDefault("DB_PATH", "./portal.db"),
		JWTSecret: envOrDefault("JWT_SECRET", "change-me-to-a-random-string"),
		JWTExpiry: parseDuration(envOrDefault("JWT_EXPIRY", "24h")),
		BaseURL:   os.Getenv("BASE_URL"),
	}
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
