package config_test

import (
	"os"
	"testing"

	"familyshare/internal/config"
)

func TestLoad_WithEnvVars(t *testing.T) {
	os.Setenv("SERVER_ADDR", ":9999")
	os.Setenv("DATABASE_PATH", "./tmp/db.sqlite")
	os.Setenv("DATA_DIR", "./tmp/data")
	defer func() {
		os.Unsetenv("SERVER_ADDR")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("DATA_DIR")
	}()

	cfg := config.Load()
	if cfg.ServerAddr != ":9999" {
		t.Fatalf("expected SERVER_ADDR :9999, got %s", cfg.ServerAddr)
	}
	if cfg.DatabasePath != "./tmp/db.sqlite" {
		t.Fatalf("expected DATABASE_PATH ./tmp/db.sqlite, got %s", cfg.DatabasePath)
	}
	if cfg.DataDir != "./tmp/data" {
		t.Fatalf("expected DATA_DIR ./tmp/data, got %s", cfg.DataDir)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// ensure env is clear
	os.Unsetenv("SERVER_ADDR")
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("DATA_DIR")

	cfg := config.Load()
	if cfg.ServerAddr == "" {
		t.Fatalf("expected default SERVER_ADDR, got empty")
	}
	if cfg.DatabasePath == "" {
		t.Fatalf("expected default DATABASE_PATH, got empty")
	}
	if cfg.DataDir == "" {
		t.Fatalf("expected default DATA_DIR, got empty")
	}
}
