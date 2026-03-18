package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/model"
	"smpp-simulator/internal/repository"
	"smpp-simulator/internal/smpp"
)

// MockHandler handles mock configuration
type MockHandler struct {
	repo       *repository.MockConfigRepository
	smppServer *smpp.Server
}

// NewMockHandler creates a new mock handler
func NewMockHandler(repo *repository.MockConfigRepository, smppServer *smpp.Server) *MockHandler {
	return &MockHandler{
		repo:       repo,
		smppServer: smppServer,
	}
}

// Get returns mock configuration
func (h *MockHandler) Get(c *gin.Context) {
	config, err := h.repo.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Update updates mock configuration
func (h *MockHandler) Update(c *gin.Context) {
	var config model.MockConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate
	if config.SuccessRate < 0 || config.SuccessRate > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "success_rate must be between 0 and 100"})
		return
	}

	// Save to database
	if err := h.repo.Save(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update SMPP server
	h.smppServer.SetMockConfig(&config)

	c.JSON(http.StatusOK, config)
}
