package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"smpp-simulator/pkg/jwt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	adminPassword string
	jwtSecret     string
	jwtExpiry     int
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(adminPassword, jwtSecret string, jwtExpiry int) *AuthHandler {
	return &AuthHandler{
		adminPassword: adminPassword,
		jwtSecret:     jwtSecret,
		jwtExpiry:     jwtExpiry,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles login requests
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Only admin user is allowed
	if req.Username != "admin" || req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := jwt.GenerateToken("admin", h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// StatusResponse represents auth status response
type StatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
}

// Status checks if the current token is valid
func (h *AuthHandler) Status(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusOK, StatusResponse{Authenticated: false})
		return
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusOK, StatusResponse{Authenticated: false})
		return
	}

	tokenString := parts[1]
	claims, err := jwt.ValidateToken(tokenString, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusOK, StatusResponse{Authenticated: false})
		return
	}

	c.JSON(http.StatusOK, StatusResponse{
		Authenticated: true,
		Username:      claims.Username,
	})
}
