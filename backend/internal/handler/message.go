package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/model"
	"smpp-simulator/internal/repository"
	"smpp-simulator/internal/smpp"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	repo *repository.MessageRepository
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(repo *repository.MessageRepository) *MessageHandler {
	return &MessageHandler{repo: repo}
}

// List returns messages with pagination and filters
func (h *MessageHandler) List(c *gin.Context) {
	// Parse query parameters
	sessionID := c.Query("session_id")
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	messages, total, err := h.repo.GetList(sessionID, status, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      messages,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Get returns a single message by ID
func (h *MessageHandler) Get(c *gin.Context) {
	id := c.Param("id")

	msg, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	c.JSON(http.StatusOK, msg)
}

// Deliver triggers a delivery report for a message
func (h *MessageHandler) Deliver(c *gin.Context) {
	id := c.Param("id")

	// Get message
	msg, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	// Update status to delivered
	if err := h.repo.UpdateStatus(id, "delivered"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "delivery report sent",
		"message_id":   msg.MessageID,
		"source_addr":  msg.SourceAddr,
		"dest_addr":    msg.DestAddr,
		"status":       "delivered",
	})
}

// StatsHandler handles statistics requests
type StatsHandler struct {
	msgRepo     *repository.MessageRepository
	sessionRepo *repository.SessionRepository
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(msgRepo *repository.MessageRepository, sessionRepo *repository.SessionRepository) *StatsHandler {
	return &StatsHandler{
		msgRepo:     msgRepo,
		sessionRepo: sessionRepo,
	}
}

// Get returns statistics
func (h *StatsHandler) Get(c *gin.Context) {
	stats, err := h.msgRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get active sessions count
	sessions, err := h.sessionRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeCount := 0
	for _, s := range sessions {
		if s.Status == "active" {
			activeCount++
		}
	}
	stats.ActiveConnections = activeCount

	c.JSON(http.StatusOK, stats)
}

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
