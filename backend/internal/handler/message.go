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
	filter := repository.MessageFilter{
		SessionID:  c.Query("session_id"),
		Status:     c.Query("status"),
		SourceAddr: c.Query("source_addr"),
		DestAddr:   c.Query("dest_addr"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	messages, total, err := h.repo.GetList(filter, pageSize, offset)
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
	h.updateStatus(c, "delivered")
}

// Fail marks a message as failed
func (h *MessageHandler) Fail(c *gin.Context) {
	h.updateStatus(c, "failed")
}

// updateStatus updates message status
func (h *MessageHandler) updateStatus(c *gin.Context, status string) {
	id := c.Param("id")

	// Get message
	msg, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	// Update status
	if err := h.repo.UpdateStatus(id, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "status updated",
		"message_id":   msg.MessageID,
		"source_addr":  msg.SourceAddr,
		"dest_addr":    msg.DestAddr,
		"status":       status,
	})
}

// StatsHandler handles statistics requests
type StatsHandler struct {
	msgRepo    *repository.MessageRepository
	smppServer *smpp.Server
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(msgRepo *repository.MessageRepository, smppServer *smpp.Server) *StatsHandler {
	return &StatsHandler{
		msgRepo:    msgRepo,
		smppServer: smppServer,
	}
}

// Get returns statistics
func (h *StatsHandler) Get(c *gin.Context) {
	stats, err := h.msgRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get active sessions count from SMPP server (real-time)
	stats.ActiveConnections = len(h.smppServer.GetSessions())

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

// DataHandler handles data management requests
type DataHandler struct {
	msgRepo     *repository.MessageRepository
	sessionRepo *repository.SessionRepository
}

// NewDataHandler creates a new data handler
func NewDataHandler(msgRepo *repository.MessageRepository, sessionRepo *repository.SessionRepository) *DataHandler {
	return &DataHandler{
		msgRepo:     msgRepo,
		sessionRepo: sessionRepo,
	}
}

// DeleteAllMessages deletes all messages
func (h *DataHandler) DeleteAllMessages(c *gin.Context) {
	if err := h.msgRepo.DeleteAllMessages(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all messages deleted"})
}

// DeleteAllSessions deletes all sessions
func (h *DataHandler) DeleteAllSessions(c *gin.Context) {
	if err := h.sessionRepo.DeleteAllSessions(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all sessions deleted"})
}

// ClearAllData deletes all messages and sessions
func (h *DataHandler) ClearAllData(c *gin.Context) {
	// Delete all messages
	if err := h.msgRepo.DeleteAllMessages(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Delete all sessions
	if err := h.sessionRepo.DeleteAllSessions(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all data cleared"})
}
