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
	smppServer.SetEventHandler(NewEventHandler(sessionRepo, wsHub))

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
	sendMessageHandler := handler.NewSendMessageHandler(smppServer)
	wsHandler := handler.NewWebSocketHandler(wsHub, cfg.JWTSecret, []string{cfg.CORSOrigins})

	// Setup router
	router := SetupRouter(&RouterConfig{
		JWTSecret:      cfg.JWTSecret,
		CORSOrigins:    cfg.CORSOrigins,
		LoginRateLimit: cfg.LoginRateLimit,
		AuthHandler:    authHandler,
		SessionHandler: sessionHandler,
		MessageHandler: messageHandler,
		StatsHandler:   statsHandler,
		MockHandler:    mockHandler,
		DataHandler:    dataHandler,
		SendHandler:    sendMessageHandler,
		WsHandler:      wsHandler,
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
