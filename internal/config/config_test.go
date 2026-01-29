package config_test

import (
	"net/netip"
	"os"
	"testing"
	"time"

	"familyshare/internal/config"
)

func TestLoad_WithEnvVars(t *testing.T) {
	os.Setenv("SERVER_ADDR", ":9999")
	os.Setenv("DATABASE_PATH", "./tmp/db.sqlite")
	os.Setenv("DATA_DIR", "./tmp/data")
	os.Setenv("TEMP_UPLOAD_DIR", "./tmp/uploads")
	os.Setenv("RATE_LIMIT_SHARE", "120")
	os.Setenv("RATE_LIMIT_ADMIN", "20")
	os.Setenv("ADMIN_PASSWORD_HASH", "$2a$10$test_hash")
	os.Setenv("JANITOR_INTERVAL", "2h30m")
	os.Setenv("APP_ENV", "production")
	os.Setenv("VIEWER_HASH_SECRET", "test-secret")
	os.Setenv("VIEWER_HASH_SECRET_REQUIRED", "true")
	os.Setenv("FORCE_HTTPS", "true")
	os.Setenv("COOKIE_SAMESITE", "Strict")
	os.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8, 192.168.0.0/16")
	defer func() {
		os.Unsetenv("SERVER_ADDR")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("DATA_DIR")
		os.Unsetenv("TEMP_UPLOAD_DIR")
		os.Unsetenv("RATE_LIMIT_SHARE")
		os.Unsetenv("RATE_LIMIT_ADMIN")
		os.Unsetenv("ADMIN_PASSWORD_HASH")
		os.Unsetenv("JANITOR_INTERVAL")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("VIEWER_HASH_SECRET")
		os.Unsetenv("VIEWER_HASH_SECRET_REQUIRED")
		os.Unsetenv("FORCE_HTTPS")
		os.Unsetenv("COOKIE_SAMESITE")
		os.Unsetenv("TRUSTED_PROXY_CIDRS")
	}()

	cfg := config.Load()

	if cfg.ServerAddr != ":9999" {
		t.Errorf("expected SERVER_ADDR :9999, got %s", cfg.ServerAddr)
	}
	if cfg.DatabasePath != "./tmp/db.sqlite" {
		t.Errorf("expected DATABASE_PATH ./tmp/db.sqlite, got %s", cfg.DatabasePath)
	}
	if cfg.DataDir != "./tmp/data" {
		t.Errorf("expected DATA_DIR ./tmp/data, got %s", cfg.DataDir)
	}
	if cfg.TempUploadDir != "./tmp/uploads" {
		t.Errorf("expected TEMP_UPLOAD_DIR ./tmp/uploads, got %s", cfg.TempUploadDir)
	}
	if cfg.RateLimitShare != 120 {
		t.Errorf("expected RATE_LIMIT_SHARE 120, got %d", cfg.RateLimitShare)
	}
	if cfg.RateLimitAdmin != 20 {
		t.Errorf("expected RATE_LIMIT_ADMIN 20, got %d", cfg.RateLimitAdmin)
	}
	if cfg.AdminPasswordHash != "$2a$10$test_hash" {
		t.Errorf("expected ADMIN_PASSWORD_HASH $2a$10$test_hash, got %s", cfg.AdminPasswordHash)
	}
	if cfg.JanitorInterval != 2*time.Hour+30*time.Minute {
		t.Errorf("expected JANITOR_INTERVAL 2h30m, got %v", cfg.JanitorInterval)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected APP_ENV production, got %s", cfg.Environment)
	}
	if cfg.ViewerHashSecret != "test-secret" {
		t.Errorf("expected VIEWER_HASH_SECRET test-secret, got %s", cfg.ViewerHashSecret)
	}
	if !cfg.RequireViewerHashSecret {
		t.Errorf("expected VIEWER_HASH_SECRET_REQUIRED true, got false")
	}
	if !cfg.ForceHTTPS {
		t.Errorf("expected FORCE_HTTPS true, got false")
	}
	if cfg.CookieSameSite != "Strict" {
		t.Errorf("expected COOKIE_SAMESITE Strict, got %s", cfg.CookieSameSite)
	}
	if len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatalf("expected 2 trusted proxy CIDRs, got %d", len(cfg.TrustedProxyCIDRs))
	}
	if cfg.TrustedProxyCIDRs[0] != netip.MustParsePrefix("10.0.0.0/8") {
		t.Errorf("expected first trusted CIDR 10.0.0.0/8, got %s", cfg.TrustedProxyCIDRs[0])
	}
}

func TestLoad_Defaults(t *testing.T) {
	// ensure env is clear
	os.Unsetenv("SERVER_ADDR")
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("DATA_DIR")
	os.Unsetenv("TEMP_UPLOAD_DIR")
	os.Unsetenv("RATE_LIMIT_SHARE")
	os.Unsetenv("RATE_LIMIT_ADMIN")
	os.Unsetenv("ADMIN_PASSWORD_HASH")
	os.Unsetenv("JANITOR_INTERVAL")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("VIEWER_HASH_SECRET")
	os.Unsetenv("VIEWER_HASH_SECRET_REQUIRED")
	os.Unsetenv("FORCE_HTTPS")
	os.Unsetenv("COOKIE_SAMESITE")
	os.Unsetenv("TRUSTED_PROXY_CIDRS")

	cfg := config.Load()

	if cfg.ServerAddr != ":8080" {
		t.Errorf("expected default SERVER_ADDR :8080, got %s", cfg.ServerAddr)
	}
	if cfg.DatabasePath != "./data/familyshare.db" {
		t.Errorf("expected default DATABASE_PATH ./data/familyshare.db, got %s", cfg.DatabasePath)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("expected default DATA_DIR ./data, got %s", cfg.DataDir)
	}
	if cfg.TempUploadDir != "" {
		t.Errorf("expected default TEMP_UPLOAD_DIR empty, got %s", cfg.TempUploadDir)
	}
	if cfg.RateLimitShare != 60 {
		t.Errorf("expected default RATE_LIMIT_SHARE 60, got %d", cfg.RateLimitShare)
	}
	if cfg.RateLimitAdmin != 10 {
		t.Errorf("expected default RATE_LIMIT_ADMIN 10, got %d", cfg.RateLimitAdmin)
	}
	if cfg.AdminPasswordHash != "" {
		t.Errorf("expected default ADMIN_PASSWORD_HASH empty, got %s", cfg.AdminPasswordHash)
	}
	if cfg.JanitorInterval != 6*time.Hour {
		t.Errorf("expected default JANITOR_INTERVAL 6h, got %v", cfg.JanitorInterval)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected default APP_ENV development, got %s", cfg.Environment)
	}
	if cfg.ViewerHashSecret != "" {
		t.Errorf("expected default VIEWER_HASH_SECRET empty, got %s", cfg.ViewerHashSecret)
	}
	if cfg.RequireViewerHashSecret {
		t.Errorf("expected default VIEWER_HASH_SECRET_REQUIRED false, got true")
	}
	if cfg.ForceHTTPS {
		t.Errorf("expected default FORCE_HTTPS false, got true")
	}
	if cfg.CookieSameSite != "Lax" {
		t.Errorf("expected default COOKIE_SAMESITE Lax, got %s", cfg.CookieSameSite)
	}
	if len(cfg.TrustedProxyCIDRs) != 0 {
		t.Errorf("expected default TRUSTED_PROXY_CIDRS empty, got %d", len(cfg.TrustedProxyCIDRs))
	}
}

func TestLoad_ViewerHashSecretRequiredInProduction(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("VIEWER_HASH_SECRET_REQUIRED")
	defer os.Unsetenv("APP_ENV")

	cfg := config.Load()

	if !cfg.RequireViewerHashSecret {
		t.Errorf("expected viewer hash secret required in production")
	}
}

func TestLoad_InvalidRateLimitFallbackToDefault(t *testing.T) {
	os.Setenv("RATE_LIMIT_SHARE", "invalid")
	os.Setenv("RATE_LIMIT_ADMIN", "not-a-number")
	defer func() {
		os.Unsetenv("RATE_LIMIT_SHARE")
		os.Unsetenv("RATE_LIMIT_ADMIN")
	}()

	cfg := config.Load()

	if cfg.RateLimitShare != 60 {
		t.Errorf("expected fallback RATE_LIMIT_SHARE 60, got %d", cfg.RateLimitShare)
	}
	if cfg.RateLimitAdmin != 10 {
		t.Errorf("expected fallback RATE_LIMIT_ADMIN 10, got %d", cfg.RateLimitAdmin)
	}
}

func TestLoad_InvalidJanitorIntervalFallbackToDefault(t *testing.T) {
	os.Setenv("JANITOR_INTERVAL", "not-a-duration")
	defer os.Unsetenv("JANITOR_INTERVAL")

	cfg := config.Load()

	if cfg.JanitorInterval != 6*time.Hour {
		t.Errorf("expected fallback JANITOR_INTERVAL 6h, got %v", cfg.JanitorInterval)
	}
}

func TestLoad_VariousJanitorIntervalFormats(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{"minutes", "30m", 30 * time.Minute},
		{"hours", "12h", 12 * time.Hour},
		{"combined", "1h30m", time.Hour + 30*time.Minute},
		{"seconds", "300s", 5 * time.Minute},
		{"complex", "2h45m30s", 2*time.Hour + 45*time.Minute + 30*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JANITOR_INTERVAL", tt.envValue)
			defer os.Unsetenv("JANITOR_INTERVAL")

			cfg := config.Load()

			if cfg.JanitorInterval != tt.expected {
				t.Errorf("expected JANITOR_INTERVAL %v, got %v", tt.expected, cfg.JanitorInterval)
			}
		})
	}
}
