package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/repository"
)

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
