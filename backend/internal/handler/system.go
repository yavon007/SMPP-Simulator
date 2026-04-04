package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/config"
)

// SystemHandler handles system configuration requests
type SystemHandler struct {
	cfg           *config.Config
	configPath    string
	passwordStore PasswordStore
}

// PasswordStore defines interface for password operations
type PasswordStore interface {
	UpdatePassword(newPassword string) error
	VerifyPassword(password string) bool
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(cfg *config.Config, configPath string) *SystemHandler {
	return &SystemHandler{
		cfg:        cfg,
		configPath: configPath,
	}
}

// SystemConfigResponse represents system config response (sanitized)
type SystemConfigResponse struct {
	// SMPP Server (read-only)
	SMPPPort string `json:"smpp_port"`

	// HTTP Server (read-only)
	HTTPPort string `json:"http_port"`

	// Database (read-only)
	DBType string `json:"db_type"`

	// Redis (read-only)
	RedisEnabled bool   `json:"redis_enabled"`
	RedisStatus  string `json:"redis_status"`

	// Auth (modifiable)
	AdminPassword string `json:"admin_password"` // masked
	JWTExpiry     int    `json:"jwt_expiry"`     // hours

	// Security (modifiable)
	CORSOrigins    string `json:"cors_origins"`
	LoginRateLimit int    `json:"login_rate_limit"`
}

// UpdateConfigRequest represents config update request
type UpdateConfigRequest struct {
	// Password change (requires old password)
	OldPassword    string `json:"old_password"`
	NewPassword    string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`

	// Other settings
	JWTExpiry      *int   `json:"jwt_expiry"`
	CORSOrigins    string `json:"cors_origins"`
	LoginRateLimit *int   `json:"login_rate_limit"`
}

// UpdateConfigResponse represents config update response
type UpdateConfigResponse struct {
	Message string `json:"message"`
}

// GetConfig returns current system configuration (sanitized)
func (h *SystemHandler) GetConfig(c *gin.Context) {
	// Check Redis status
	redisStatus := "disabled"
	if h.cfg.RedisEnabled {
		redisStatus = "enabled"
	}

	response := SystemConfigResponse{
		SMPPPort:       h.cfg.SMPPPort,
		HTTPPort:       h.cfg.HTTPPort,
		DBType:         h.cfg.DBType,
		RedisEnabled:   h.cfg.RedisEnabled,
		RedisStatus:    redisStatus,
		AdminPassword:  "********", // Always masked
		JWTExpiry:      h.cfg.JWTExpiry,
		CORSOrigins:    h.cfg.CORSOrigins,
		LoginRateLimit: h.cfg.LoginRateLimit,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateConfig updates system configuration
func (h *SystemHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	changes := []string{}

	// Handle password change
	if req.NewPassword != "" {
		// Validate old password
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old password is required to change password"})
			return
		}

		if req.OldPassword != h.cfg.AdminPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect old password"})
			return
		}

		// Validate new password
		if len(req.NewPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 6 characters"})
			return
		}

		if req.NewPassword != req.ConfirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
			return
		}

		// Update password
		h.cfg.AdminPassword = req.NewPassword
		changes = append(changes, "admin password")
	}

	// Update JWT expiry
	if req.JWTExpiry != nil {
		if *req.JWTExpiry < 1 || *req.JWTExpiry > 720 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JWT expiry must be between 1 and 720 hours"})
			return
		}
		h.cfg.JWTExpiry = *req.JWTExpiry
		changes = append(changes, "JWT expiry")
	}

	// Update CORS origins
	if req.CORSOrigins != "" && req.CORSOrigins != h.cfg.CORSOrigins {
		h.cfg.CORSOrigins = req.CORSOrigins
		changes = append(changes, "CORS origins")
	}

	// Update rate limit
	if req.LoginRateLimit != nil {
		if *req.LoginRateLimit < 1 || *req.LoginRateLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "login rate limit must be between 1 and 100"})
			return
		}
		h.cfg.LoginRateLimit = *req.LoginRateLimit
		changes = append(changes, "login rate limit")
	}

	if len(changes) == 0 {
		c.JSON(http.StatusOK, UpdateConfigResponse{Message: "No changes made"})
		return
	}

	c.JSON(http.StatusOK, UpdateConfigResponse{
		Message: "Configuration updated successfully",
	})
}

// RedisStatusResponse represents Redis status check response
type RedisStatusResponse struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// CheckRedis checks Redis connection status
func (h *SystemHandler) CheckRedis(c *gin.Context) {
	// This would need actual Redis client to check connection
	// For now, just return the configured status
	response := RedisStatusResponse{
		Connected: h.cfg.RedisEnabled,
	}

	if !h.cfg.RedisEnabled {
		response.Error = "Redis is not enabled"
	}

	c.JSON(http.StatusOK, response)
}
