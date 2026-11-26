// Package config reads settings from environment variables,
// so the same binary can run on your laptop, in staging and in production.
package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return d
	}
	return fallback
}
