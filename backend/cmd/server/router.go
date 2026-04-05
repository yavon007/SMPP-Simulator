package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/handler"
	"smpp-simulator/internal/middleware"
)

// RouterConfig holds dependencies for router setup
type RouterConfig struct {
	JWTSecret       string
	CORSOrigins     string
	LoginRateLimit  int
	LoginLimiter    *middleware.RateLimiter
	AuthHandler     *handler.AuthHandler
	SessionHandler  *handler.SessionHandler
	MessageHandler  *handler.MessageHandler
	StatsHandler    *handler.StatsHandler
	MockHandler     *handler.MockHandler
	DataHandler     *handler.DataHandler
	SendHandler     *handler.SendMessageHandler
	WsHandler       *handler.WebSocketHandler
	SystemHandler   *handler.SystemHandler
	TemplateHandler *handler.TemplateHandler
	OutboundHandler *handler.OutboundHandler
}

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Parse CORS origins
	allowOrigins := []string{"*"}
	if cfg.CORSOrigins != "" && cfg.CORSOrigins != "*" {
		allowOrigins = strings.Split(cfg.CORSOrigins, ",")
		for i, origin := range allowOrigins {
			allowOrigins[i] = strings.TrimSpace(origin)
		}
	}

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Create login rate limiter (default 5 attempts per minute)
	loginLimiter := cfg.LoginLimiter
	if loginLimiter == nil {
		loginLimiter = middleware.NewRateLimiter(cfg.LoginRateLimit, time.Minute)
	}

	// API routes
	api := router.Group("/api")
	{
		// Public routes (no auth required)
		api.POST("/auth/login", middleware.RateLimitMiddleware(loginLimiter), cfg.AuthHandler.Login)
		api.GET("/auth/status", cfg.AuthHandler.Status)
		api.GET("/stats", cfg.StatsHandler.Get)
		api.GET("/messages", cfg.MessageHandler.List)
		api.GET("/messages/export", cfg.MessageHandler.Export)
		api.GET("/messages/:id", cfg.MessageHandler.Get)
	}

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/sessions", cfg.SessionHandler.List)
		protected.GET("/sessions/:id/stats", cfg.SessionHandler.GetStats)
		protected.DELETE("/sessions/:id", cfg.SessionHandler.Delete)
		protected.DELETE("/messages/batch", cfg.MessageHandler.BatchDelete)
		protected.POST("/messages/:id/deliver", cfg.MessageHandler.Deliver)
		protected.POST("/messages/:id/fail", cfg.MessageHandler.Fail)
		protected.GET("/mock/config", cfg.MockHandler.Get)
		protected.PUT("/mock/config", cfg.MockHandler.Update)
		protected.DELETE("/data/messages", cfg.DataHandler.DeleteAllMessages)
		protected.DELETE("/data/sessions", cfg.DataHandler.DeleteAllSessions)
		protected.DELETE("/data/all", cfg.DataHandler.ClearAllData)
		// Send message (deliver_sm)
		protected.GET("/send/receivers", cfg.SendHandler.ListReceivers)
		protected.POST("/send", cfg.SendHandler.SendMessage)
		// System configuration
		protected.GET("/system/config", cfg.SystemHandler.GetConfig)
		protected.PUT("/system/config", cfg.SystemHandler.UpdateConfig)
		protected.GET("/system/redis", cfg.SystemHandler.CheckRedis)
		// Message templates
		protected.GET("/templates", cfg.TemplateHandler.List)
		protected.GET("/templates/:id", cfg.TemplateHandler.Get)
		protected.POST("/templates", cfg.TemplateHandler.Create)
		protected.PUT("/templates/:id", cfg.TemplateHandler.Update)
		protected.DELETE("/templates/:id", cfg.TemplateHandler.Delete)
		// Rate limit status
		protected.GET("/stats/rate-limit", cfg.StatsHandler.GetRateLimit)
		// Outbound SMPP connections
		protected.GET("/outbound", cfg.OutboundHandler.List)
		protected.POST("/outbound/connect", cfg.OutboundHandler.Connect)
		protected.DELETE("/outbound/:id", cfg.OutboundHandler.Delete)
		protected.POST("/outbound/:id/send", cfg.OutboundHandler.SendMessage)
	}

	// WebSocket
	router.GET("/ws", cfg.WsHandler.Handle)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
