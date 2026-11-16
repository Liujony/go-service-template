// Package config loads the process configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config is everything the process needs to start. Defaults are safe for local
// development; production deploys override the address, database and log level.
type Config struct {
	Addr            string
	DatabasePath    string
	LogLevel        string
	ShutdownTimeout time.Duration
	SeedDemo        bool
}

// Load reads the environment and applies defaults. It never panics: a bad value
// falls back to the default and is reported by the caller through Validate.
func Load() Config {
	cfg := Config{
		Addr:            env("HTTP_ADDR", ":8080"),
		DatabasePath:    env("DATABASE_PATH", "app.db"),
		LogLevel:        env("LOG_LEVEL", "INFO"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", 10*time.Second),
		SeedDemo:        env("SEED_DEMO", "false") == "true",
	}
	return cfg
}

// Validate checks values that cannot be expressed as defaults.
func (c Config) Validate() error {
	switch c.LogLevel {
	case "DEBUG", "INFO", "WARN", "ERROR":
	default:
		return fmt.Errorf("LOG_LEVEL must be DEBUG, INFO, WARN or ERROR, got %q", c.LogLevel)
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be positive")
	}
	return nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if parsed, err := time.ParseDuration(raw); err == nil {
		return parsed
	}
	return fallback
}
