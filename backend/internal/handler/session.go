package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/smpp"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	smppServer *smpp.Server
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(smppServer *smpp.Server) *SessionHandler {
	return &SessionHandler{smppServer: smppServer}
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
