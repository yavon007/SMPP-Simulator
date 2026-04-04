package config

import (
	"log"
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
	DBType string `yaml:"db_type"` // sqlite, postgres, mysql
	DBPath string `yaml:"db_path"` // for sqlite
	DBHost string `yaml:"db_host"`
	DBPort int    `yaml:"db_port"`
	DBName string `yaml:"db_name"`
	DBUser string `yaml:"db_user"`
	DBPassword string `yaml:"db_password"`

	// Redis
	RedisHost     string `yaml:"redis_host"`
	RedisPort     string `yaml:"redis_port"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
	RedisEnabled  bool   `yaml:"redis_enabled"`

	// Auth
	AdminPassword string `yaml:"admin_password"`
	JWTSecret     string `yaml:"jwt_secret"`
	JWTExpiry     int    `yaml:"jwt_expiry"` // hours

	// Security
	CORSOrigins    string `yaml:"cors_origins"`    // Comma-separated allowed origins
	LoginRateLimit int    `yaml:"login_rate_limit"` // Max login attempts per minute

	// UI
	Language string `yaml:"language"` // UI language: zh-CN, en-US
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		SMPPHost: "0.0.0.0",
		SMPPPort: "2775",
		HTTPHost: "0.0.0.0",
		HTTPPort: "8080",

		DBType: "sqlite",
		DBPath: "./smpp.db",
		DBPort: 5432, // default for postgres

		RedisEnabled: false,

		AdminPassword: "admin123",
		JWTSecret:     "smpp-simulator-secret-key",
		JWTExpiry:     24,

		CORSOrigins:    "*",
		LoginRateLimit: 5, // 5 attempts per minute

		Language: "zh-CN",
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
	// Database
	if v := os.Getenv("DB_TYPE"); v != "" {
		cfg.DBType = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.DBPort = i
		}
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.RedisHost = v
		cfg.RedisEnabled = true
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		cfg.RedisPort = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.RedisPassword = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.RedisDB = i
		}
	}
	if v := os.Getenv("REDIS_ENABLED"); v != "" {
		cfg.RedisEnabled = v == "true" || v == "1"
	}
	// Auth
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
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		cfg.CORSOrigins = v
	}
	if v := os.Getenv("LOGIN_RATE_LIMIT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.LoginRateLimit = i
		}
	}
	// UI
	if v := os.Getenv("LANGUAGE"); v != "" {
		cfg.Language = v
	}
}

// CheckSecurityWarnings checks for insecure default configurations
// and logs warnings if found
func (c *Config) CheckSecurityWarnings() {
	warnings := false

	if c.AdminPassword == "admin123" {
		log.Println("WARNING: Using default admin password 'admin123'. Please change it via ADMIN_PASSWORD environment variable or config file.")
		warnings = true
	}

	if c.JWTSecret == "smpp-simulator-secret-key" {
		log.Println("WARNING: Using default JWT secret. Please change it via JWT_SECRET environment variable or config file.")
		warnings = true
	}

	if warnings {
		log.Println("WARNING: Default credentials detected. This is insecure for production use.")
	}
}
