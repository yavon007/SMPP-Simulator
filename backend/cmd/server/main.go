package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Check for insecure defaults
	cfg.CheckSecurityWarnings()

	// Initialize database
	dbCfg := &repository.DatabaseConfig{
		Type:     cfg.DBType,
		Path:     cfg.DBPath,
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
	}
	db, err := repository.NewDatabase(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database connected: %s", cfg.DBType)

	// Initialize repositories
	sessionRepo := repository.NewSessionRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	mockConfigRepo := repository.NewMockConfigRepository(db)
	templateRepo := repository.NewTemplateRepository(db)

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
	smppServer.SetEventHandler(NewEventHandler(sessionRepo, wsHub))

	// Start SMPP server
	if err := smppServer.Start(); err != nil {
		log.Fatalf("Failed to start SMPP server: %v", err)
	}
	defer smppServer.Stop()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg.AdminPassword, cfg.JWTSecret, cfg.JWTExpiry)
	sessionHandler := handler.NewSessionHandler(smppServer, messageRepo)
	messageHandler := handler.NewMessageHandler(messageRepo)
	loginLimiter := middleware.NewRateLimiter(cfg.LoginRateLimit, time.Minute)
	statsHandler := handler.NewStatsHandler(messageRepo, smppServer, loginLimiter)
	mockHandler := handler.NewMockHandler(mockConfigRepo, smppServer)
	dataHandler := handler.NewDataHandler(messageRepo, sessionRepo)
	sendMessageHandler := handler.NewSendMessageHandler(smppServer)
	wsHandler := handler.NewWebSocketHandler(wsHub, cfg.JWTSecret, []string{cfg.CORSOrigins})
	systemHandler := handler.NewSystemHandler(cfg, "")

	// Setup router
	router := SetupRouter(&RouterConfig{
		JWTSecret:      cfg.JWTSecret,
		CORSOrigins:    cfg.CORSOrigins,
		LoginRateLimit: cfg.LoginRateLimit,
		LoginLimiter:   loginLimiter,
		AuthHandler:    authHandler,
		SessionHandler: sessionHandler,
		MessageHandler: messageHandler,
		StatsHandler:   statsHandler,
		MockHandler:    mockHandler,
		DataHandler:    dataHandler,
		SendHandler:    sendMessageHandler,
		WsHandler:      wsHandler,
		SystemHandler:  systemHandler,
	})

	// Start HTTP server
	httpAddr := fmt.Sprintf("%s:%s", cfg.HTTPHost, cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	// Shutdown WebSocket hub
	wsHub.Shutdown()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
