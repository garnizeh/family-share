package config

import (
	"log"
	"net/netip"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"familyshare/internal/requestip"
)

type Config struct {
	ServerAddr     string
	DatabasePath   string
	DataDir        string
	TempUploadDir  string
	Environment    string
	ForceHTTPS     bool
	CookieSameSite string

	TrustedProxyCIDRs []netip.Prefix

	// Rate limiting configuration
	RateLimitShare int // requests per minute for share links
	RateLimitAdmin int // requests per minute for admin endpoints

	// Admin authentication
	AdminPasswordHash       string // bcrypt hash of admin password
	ViewerHashSecret        string // HMAC secret for viewer hash
	RequireViewerHashSecret bool   // require viewer hash secret (fail if missing)

	// Janitor configuration
	JanitorInterval time.Duration // interval for cleanup tasks
}

func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	env := getEnv("APP_ENV", "development")
	requireViewerHashSecret := getEnvBool("VIEWER_HASH_SECRET_REQUIRED", env == "production")
	trustedProxyCIDRs, err := requestip.ParseTrustedProxyCIDRs(getEnv("TRUSTED_PROXY_CIDRS", ""))
	if err != nil {
		log.Printf("invalid TRUSTED_PROXY_CIDRS: %v", err)
	}

	return &Config{
		ServerAddr:              getEnv("SERVER_ADDR", ":8080"),
		DatabasePath:            getEnv("DATABASE_PATH", "./data/familyshare.db"),
		DataDir:                 getEnv("DATA_DIR", "./data"),
		TempUploadDir:           getEnv("TEMP_UPLOAD_DIR", ""),
		Environment:             env,
		ForceHTTPS:              getEnvBool("FORCE_HTTPS", env == "production"),
		CookieSameSite:          getEnv("COOKIE_SAMESITE", "Lax"),
		TrustedProxyCIDRs:       trustedProxyCIDRs,
		RateLimitShare:          getEnvInt("RATE_LIMIT_SHARE", 60),
		RateLimitAdmin:          getEnvInt("RATE_LIMIT_ADMIN", 10),
		AdminPasswordHash:       getEnv("ADMIN_PASSWORD_HASH", ""),
		ViewerHashSecret:        getEnv("VIEWER_HASH_SECRET", ""),
		RequireViewerHashSecret: requireViewerHashSecret,
		JanitorInterval:         getEnvDuration("JANITOR_INTERVAL", 6*time.Hour),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "1", "true", "TRUE", "yes", "YES", "on", "ON":
			return true
		case "0", "false", "FALSE", "no", "NO", "off", "OFF":
			return false
		}
	}
	return defaultValue
}
