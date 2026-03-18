package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/repository"
	"smpp-simulator/internal/smpp"
)

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
