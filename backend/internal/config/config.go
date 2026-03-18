package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents application configuration
type Config struct {
	// SMPP Server
	SMPPHost string `yaml:"smpp_host"`
	SMPPPort string `yaml:"smpp_port"`

	// HTTP Server
	HTTPHost string `yaml:"http_host"`
	HTTPPort string `yaml:"http_port"`

	// Database
	DBPath string `yaml:"db_path"`

	// Auth
	AdminPassword string `yaml:"admin_password"`
	JWTSecret     string `yaml:"jwt_secret"`
	JWTExpiry     int `yaml:"jwt_expiry"` // hours
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		SMPPHost: "0.0.0.0",
		SMPPPort: "2775",
		HTTPHost: "0.0.0.0",
		HTTPPort: "8080",
		DBPath:   "./smpp.db",

		AdminPassword: "admin123",
		JWTSecret:     "smpp-simulator-secret-key",
		JWTExpiry:     24,
	}
}

// Load loads configuration from file and environment variables
// Priority: environment variables > config file > defaults
func Load() *Config {
	cfg := DefaultConfig()

	// Try to load config file
	configPath := getConfigPath()
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			yaml.Unmarshal(data, cfg)
		}
	}

	// Override with environment variables
	overrideWithEnv(cfg)

	return cfg
}

// getConfigPath returns config file path
// Checks: ./config.yaml, ./config.yml, ./configs/config.yaml
func getConfigPath() string {
	candidates := []string{
		"config.yaml",
		"config.yml",
		"./configs/config.yaml",
		"./configs/config.yml",
	}

	// Also check CONFIG_PATH env
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		candidates = append([]string{envPath}, candidates...)
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return ""
}

// overrideWithEnv overrides config with environment variables
func overrideWithEnv(cfg *Config) {
	if v := os.Getenv("SMPP_HOST"); v != "" {
		cfg.SMPPHost = v
	}
	if v := os.Getenv("SMPP_PORT"); v != "" {
		cfg.SMPPPort = v
	}
	if v := os.Getenv("HTTP_HOST"); v != "" {
		cfg.HTTPHost = v
	}
	if v := os.Getenv("HTTP_PORT"); v != "" {
		cfg.HTTPPort = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		cfg.AdminPassword = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("JWT_EXPIRY"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.JWTExpiry = i
		}
	}
}
