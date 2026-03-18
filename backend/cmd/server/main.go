package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/config"
	"smpp-simulator/internal/handler"
	"smpp-simulator/internal/middleware"
	"smpp-simulator/internal/model"
	"smpp-simulator/internal/repository"
	"smpp-simulator/internal/smpp"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	sessionRepo := repository.NewSessionRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	mockConfigRepo := repository.NewMockConfigRepository(db)

	// Create SMPP server
	smppServer := smpp.NewServer(cfg.SMPPHost, cfg.SMPPPort, messageRepo)

	// Load mock config
	mockConfig, err := mockConfigRepo.Get()
	if err != nil {
		mockConfig = model.DefaultMockConfig()
	}
	smppServer.SetMockConfig(mockConfig)

	// Create WebSocket hub
	wsHub := handler.NewWebSocketHub()
	go wsHub.Run()

	// Set event handler for real-time updates
	smppServer.SetEventHandler(&eventHandler{
		sessionRepo: sessionRepo,
		wsHub:       wsHub,
	})

	// Start SMPP server
	if err := smppServer.Start(); err != nil {
		log.Fatalf("Failed to start SMPP server: %v", err)
	}
	defer smppServer.Stop()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg.AdminPassword, cfg.JWTSecret, cfg.JWTExpiry)
	sessionHandler := handler.NewSessionHandler(smppServer)
	messageHandler := handler.NewMessageHandler(messageRepo)
	statsHandler := handler.NewStatsHandler(messageRepo, smppServer)
	mockHandler := handler.NewMockHandler(mockConfigRepo, smppServer)
	dataHandler := handler.NewDataHandler(messageRepo, sessionRepo)
	wsHandler := handler.NewWebSocketHandler(wsHub)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	api := router.Group("/api")
	{
		// Public routes (no auth required)
		api.POST("/auth/login", authHandler.Login)
		api.GET("/auth/status", authHandler.Status)
		api.GET("/stats", statsHandler.Get)
		api.GET("/messages", messageHandler.List)
		api.GET("/messages/:id", messageHandler.Get)
	}

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/sessions", sessionHandler.List)
		protected.DELETE("/sessions/:id", sessionHandler.Delete)
		protected.POST("/messages/:id/deliver", messageHandler.Deliver)
		protected.POST("/messages/:id/fail", messageHandler.Fail)
		protected.GET("/mock/config", mockHandler.Get)
		protected.PUT("/mock/config", mockHandler.Update)
		protected.DELETE("/data/messages", dataHandler.DeleteAllMessages)
		protected.DELETE("/data/sessions", dataHandler.DeleteAllSessions)
		protected.DELETE("/data/all", dataHandler.ClearAllData)
	}

	// WebSocket
	router.GET("/ws", wsHandler.Handle)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start HTTP server
	httpAddr := fmt.Sprintf("%s:%s", cfg.HTTPHost, cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	go func() {
		log.Printf("HTTP server started on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Printf("SMPP Simulator started successfully")
	log.Printf("SMPP: %s:%s", cfg.SMPPHost, cfg.SMPPPort)
	log.Printf("HTTP: %s", httpAddr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// eventHandler implements smpp.EventHandler
type eventHandler struct {
	sessionRepo *repository.SessionRepository
	wsHub       *handler.WebSocketHub
}

func (h *eventHandler) OnSessionConnect(session *model.Session) {
	// Save session to database
	if err := h.sessionRepo.Save(session); err != nil {
		log.Printf("Failed to save session: %v", err)
	}

	// Broadcast to WebSocket clients
	data, _ := json.Marshal(map[string]interface{}{
		"type":    "session_connect",
		"session": session,
	})
	h.wsHub.Broadcast(data)
}

func (h *eventHandler) OnSessionDisconnect(sessionID string) {
	// Update session status
	if err := h.sessionRepo.UpdateStatus(sessionID, "closed"); err != nil {
		log.Printf("Failed to update session status: %v", err)
	}

	// Broadcast to WebSocket clients
	data, _ := json.Marshal(map[string]interface{}{
		"type":       "session_disconnect",
		"session_id": sessionID,
	})
	h.wsHub.Broadcast(data)
}

func (h *eventHandler) OnMessageReceived(msg *model.Message) {
	// Broadcast to WebSocket clients
	data, _ := json.Marshal(map[string]interface{}{
		"type":    "message_received",
		"message": msg,
	})
	h.wsHub.Broadcast(data)
}

func (h *eventHandler) OnMessageDelivered(msgID string) {
	// Broadcast to WebSocket clients
	data, _ := json.Marshal(map[string]interface{}{
		"type":       "message_delivered",
		"message_id": msgID,
	})
	h.wsHub.Broadcast(data)
}
