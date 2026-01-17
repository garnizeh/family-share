package config

import "os"

type Config struct {
	ServerAddr   string
	DatabasePath string
	DataDir      string
}

func Load() *Config {
	return &Config{
		ServerAddr:   getEnv("SERVER_ADDR", ":8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/familyshare.db"),
		DataDir:      getEnv("DATA_DIR", "./data"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
