package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr   string
	DatabasePath string
	DataDir      string
	
	// Rate limiting configuration
	RateLimitShare int // requests per minute for share links
	RateLimitAdmin int // requests per minute for admin endpoints
	
	// Admin authentication
	AdminPasswordHash string // bcrypt hash of admin password
	
	// Janitor configuration
	JanitorInterval time.Duration // interval for cleanup tasks
}

func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()
		
	return &Config{
		ServerAddr:        getEnv("SERVER_ADDR", ":8080"),
		DatabasePath:      getEnv("DATABASE_PATH", "./data/familyshare.db"),
		DataDir:           getEnv("DATA_DIR", "./data"),
		RateLimitShare:    getEnvInt("RATE_LIMIT_SHARE", 60),
		RateLimitAdmin:    getEnvInt("RATE_LIMIT_ADMIN", 10),
		AdminPasswordHash: getEnv("ADMIN_PASSWORD_HASH", ""),
		JanitorInterval:   getEnvDuration("JANITOR_INTERVAL", 6*time.Hour),
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
