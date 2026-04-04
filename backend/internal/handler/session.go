package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/repository"
	"smpp-simulator/internal/smpp"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	smppServer *smpp.Server
	msgRepo    *repository.MessageRepository
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(smppServer *smpp.Server, msgRepo *repository.MessageRepository) *SessionHandler {
	return &SessionHandler{
		smppServer: smppServer,
		msgRepo:    msgRepo,
	}
}

// List returns all sessions
func (h *SessionHandler) List(c *gin.Context) {
	sessions := h.smppServer.GetSessions()

	result := make([]gin.H, len(sessions))
	for i, s := range sessions {
		result[i] = gin.H{
			"id":           s.ID,
			"system_id":    s.SystemID,
			"bind_type":    s.BindType,
			"remote_addr":  s.RemoteAddr,
			"connected_at": s.ConnectedAt,
			"status":       s.Status,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": len(result),
	})
}

// Delete disconnects a session
func (h *SessionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.smppServer.DisconnectSession(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session disconnected"})
}

// GetStats returns statistics for a specific session
func (h *SessionHandler) GetStats(c *gin.Context) {
	id := c.Param("id")

	// Get session info
	sessions := h.smppServer.GetSessions()
	var sessionInfo gin.H
	for _, s := range sessions {
		if s.ID == id {
			sessionInfo = gin.H{
				"id":           s.ID,
				"system_id":    s.SystemID,
				"bind_type":    s.BindType,
				"remote_addr":  s.RemoteAddr,
				"connected_at": s.ConnectedAt,
				"status":       s.Status,
			}
			break
		}
	}

	// Get message stats
	stats, err := h.msgRepo.GetStatsBySessionID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get recent messages (last 5)
	recentMessages, err := h.msgRepo.GetRecentBySessionID(id, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate success rate
	successRate := 0.0
	if stats.Total > 0 {
		successRate = float64(stats.Delivered) / float64(stats.Total) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"session": sessionInfo,
		"stats": gin.H{
			"total":        stats.Total,
			"delivered":    stats.Delivered,
			"failed":       stats.Failed,
			"pending":      stats.Pending,
			"success_rate": successRate,
		},
		"recent_messages": recentMessages,
	})
}
