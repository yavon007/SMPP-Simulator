package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/handler"
	"smpp-simulator/internal/middleware"
)

// RouterConfig holds dependencies for router setup
type RouterConfig struct {
	JWTSecret      string
	AuthHandler    *handler.AuthHandler
	SessionHandler *handler.SessionHandler
	MessageHandler *handler.MessageHandler
	StatsHandler   *handler.StatsHandler
	MockHandler    *handler.MockHandler
	DataHandler    *handler.DataHandler
	SendHandler    *handler.SendMessageHandler
	WsHandler      *handler.WebSocketHandler
}

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	// Note: AllowCredentials cannot be true when AllowOrigins is "*"
	// For production, configure specific origins via environment variable
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	api := router.Group("/api")
	{
		// Public routes (no auth required)
		api.POST("/auth/login", cfg.AuthHandler.Login)
		api.GET("/auth/status", cfg.AuthHandler.Status)
		api.GET("/stats", cfg.StatsHandler.Get)
		api.GET("/messages", cfg.MessageHandler.List)
		api.GET("/messages/:id", cfg.MessageHandler.Get)
	}

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/sessions", cfg.SessionHandler.List)
		protected.DELETE("/sessions/:id", cfg.SessionHandler.Delete)
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
	}

	// WebSocket
	router.GET("/ws", cfg.WsHandler.Handle)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
