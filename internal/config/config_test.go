package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("SEED_DEMO", "")

	cfg := Load()
	if cfg.Addr != ":8080" || cfg.DatabasePath != "app.db" || cfg.LogLevel != "INFO" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected shutdown timeout: %v", cfg.ShutdownTimeout)
	}
	if cfg.SeedDemo {
		t.Fatal("seed should default to false")
	}
}

func TestLoadOverridesAndDurationForms(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("SHUTDOWN_TIMEOUT", "3")
	t.Setenv("SEED_DEMO", "true")

	cfg := Load()
	if cfg.Addr != "127.0.0.1:9090" || cfg.LogLevel != "DEBUG" || !cfg.SeedDemo {
		t.Fatalf("overrides not applied: %+v", cfg)
	}
	if cfg.ShutdownTimeout != 3*time.Second {
		t.Fatalf("integer seconds not parsed: %v", cfg.ShutdownTimeout)
	}

	t.Setenv("SHUTDOWN_TIMEOUT", "1500ms")
	if got := Load().ShutdownTimeout; got != 1500*time.Millisecond {
		t.Fatalf("duration string not parsed: %v", got)
	}
}

func TestValidate(t *testing.T) {
	valid := Config{LogLevel: "WARN", ShutdownTimeout: time.Second}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	badLevel := Config{LogLevel: "LOUD", ShutdownTimeout: time.Second}
	if err := badLevel.Validate(); err == nil {
		t.Fatal("invalid log level accepted")
	}
	badTimeout := Config{LogLevel: "INFO", ShutdownTimeout: 0}
	if err := badTimeout.Validate(); err == nil {
		t.Fatal("non-positive timeout accepted")
	}
}
