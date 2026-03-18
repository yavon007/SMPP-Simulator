package config

import (
	"os"
	"strconv"
)

type Config struct {
	// SMPP Server
	SMPPHost string
	SMPPPort string

	// HTTP Server
	HTTPHost string
	HTTPPort string

	// Database
	DBPath string

	// Auth
	AdminPassword string
	JWTSecret     string
	JWTExpiry     int // hours
}

func Load() *Config {
	return &Config{
		SMPPHost: getEnv("SMPP_HOST", "0.0.0.0"),
		SMPPPort: getEnv("SMPP_PORT", "2775"),
		HTTPHost: getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		DBPath:   getEnv("DB_PATH", "./smpp.db"),

		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		JWTSecret:     getEnv("JWT_SECRET", "smpp-simulator-secret-key"),
		JWTExpiry:     getEnvInt("JWT_EXPIRY", 24),
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
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
