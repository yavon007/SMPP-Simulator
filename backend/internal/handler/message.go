package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/repository"
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
		Content:    c.Query("content"),
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
		"message":     "status updated",
		"message_id":  msg.MessageID,
		"source_addr": msg.SourceAddr,
		"dest_addr":   msg.DestAddr,
		"status":      status,
	})
}
