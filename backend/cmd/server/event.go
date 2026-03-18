package main

import (
	"encoding/json"
	"log"

	"smpp-simulator/internal/handler"
	"smpp-simulator/internal/model"
	"smpp-simulator/internal/repository"
)

// eventHandler implements smpp.EventHandler
type eventHandler struct {
	sessionRepo *repository.SessionRepository
	wsHub       *handler.WebSocketHub
}

// NewEventHandler creates a new event handler
func NewEventHandler(sessionRepo *repository.SessionRepository, wsHub *handler.WebSocketHub) *eventHandler {
	return &eventHandler{
		sessionRepo: sessionRepo,
		wsHub:       wsHub,
	}
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
