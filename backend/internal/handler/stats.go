package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/middleware"
	"smpp-simulator/internal/repository"
	"smpp-simulator/internal/smpp"
)

// StatsHandler handles statistics requests
type StatsHandler struct {
	msgRepo      *repository.MessageRepository
	smppServer   *smpp.Server
	loginLimiter *middleware.RateLimiter
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(msgRepo *repository.MessageRepository, smppServer *smpp.Server, loginLimiter *middleware.RateLimiter) *StatsHandler {
	return &StatsHandler{
		msgRepo:      msgRepo,
		smppServer:   smppServer,
		loginLimiter: loginLimiter,
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

// GetRateLimit returns the current rate limit status for the client
func (h *StatsHandler) GetRateLimit(c *gin.Context) {
	if h.loginLimiter == nil {
		c.JSON(http.StatusOK, gin.H{
			"remaining":      -1,
			"reset_at":       nil,
			"total":          -1,
			"is_limited":     false,
			"window_seconds": 0,
		})
		return
	}

	ip := c.ClientIP()
	status := h.loginLimiter.GetStatus(ip)

	c.JSON(http.StatusOK, status)
}
