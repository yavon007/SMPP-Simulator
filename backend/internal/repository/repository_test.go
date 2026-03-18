package repository

import (
	"os"
	"testing"
	"time"

	"smpp-simulator/internal/model"
)

var testDB *Database

func TestMain(m *testing.M) {
	// Setup
	var err error
	testDB, err = NewDatabase(":memory:")
	if err != nil {
		panic(err)
	}

	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

func TestSessionRepository_SaveAndGet(t *testing.T) {
	repo := NewSessionRepository(testDB)

	session := &model.Session{
		ID:          "test-session-1",
		SystemID:    "testuser",
		BindType:    "TX",
		RemoteAddr:  "127.0.0.1:12345",
		ConnectedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:      "active",
	}

	// Save
	err := repo.Save(session)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// GetByID
	got, err := repo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.SystemID != session.SystemID {
		t.Errorf("SystemID = %v, want %v", got.SystemID, session.SystemID)
	}
}

func TestSessionRepository_GetAll(t *testing.T) {
	repo := NewSessionRepository(testDB)

	// Clear first
	repo.DeleteAllSessions()

	// Add multiple sessions
	for i := 0; i < 3; i++ {
		session := &model.Session{
			ID:          "session-" + string(rune('a'+i)),
			SystemID:    "user" + string(rune('a'+i)),
			BindType:    "TX",
			RemoteAddr:  "127.0.0.1",
			ConnectedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:      "active",
		}
		repo.Save(session)
	}

	sessions, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("GetAll() returned %d sessions, want 3", len(sessions))
	}
}

func TestSessionRepository_UpdateStatus(t *testing.T) {
	repo := NewSessionRepository(testDB)

	session := &model.Session{
		ID:          "test-session-2",
		SystemID:    "testuser",
		BindType:    "TX",
		RemoteAddr:  "127.0.0.1",
		ConnectedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:      "active",
	}
	repo.Save(session)

	// Update status
	err := repo.UpdateStatus(session.ID, "closed")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// Verify
	got, _ := repo.GetByID(session.ID)
	if got.Status != "closed" {
		t.Errorf("Status = %v, want closed", got.Status)
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	repo := NewSessionRepository(testDB)

	session := &model.Session{
		ID:          "test-session-3",
		SystemID:    "testuser",
		BindType:    "TX",
		RemoteAddr:  "127.0.0.1",
		ConnectedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:      "active",
	}
	repo.Save(session)

	// Delete
	err := repo.Delete(session.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify
	_, err = repo.GetByID(session.ID)
	if err == nil {
		t.Error("GetByID() should return error for deleted session")
	}
}

func TestMessageRepository_SaveAndGet(t *testing.T) {
	repo := NewMessageRepository(testDB)

	msg := &model.Message{
		ID:          "msg-1",
		SessionID:   "session-1",
		MessageID:   "SM-001",
		SequenceNum: 1,
		SourceAddr:  "12345",
		DestAddr:    "67890",
		Content:     "Test message",
		Encoding:    "GSM7",
		Status:      "pending",
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Save
	err := repo.Save(msg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// GetByID
	got, err := repo.GetByID(msg.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Content != msg.Content {
		t.Errorf("Content = %v, want %v", got.Content, msg.Content)
	}
}

func TestMessageRepository_GetByMessageID(t *testing.T) {
	repo := NewMessageRepository(testDB)

	msg := &model.Message{
		ID:          "msg-2",
		SessionID:   "session-1",
		MessageID:   "SM-UNIQUE-001",
		SequenceNum: 2,
		SourceAddr:  "12345",
		DestAddr:    "67890",
		Content:     "Test message 2",
		Encoding:    "GSM7",
		Status:      "pending",
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	repo.Save(msg)

	got, err := repo.GetByMessageID("SM-UNIQUE-001")
	if err != nil {
		t.Fatalf("GetByMessageID() error = %v", err)
	}
	if got.ID != msg.ID {
		t.Errorf("ID = %v, want %v", got.ID, msg.ID)
	}
}

func TestMessageRepository_GetList(t *testing.T) {
	repo := NewMessageRepository(testDB)

	// Clear first
	repo.DeleteAllMessages()

	// Add messages with different statuses
	for i := 0; i < 5; i++ {
		status := "pending"
		if i >= 3 {
			status = "delivered"
		}
		msg := &model.Message{
			ID:          "msg-list-" + string(rune('a'+i)),
			SessionID:   "session-1",
			MessageID:   "SM-" + string(rune('0'+i)),
			SequenceNum: uint32(i),
			SourceAddr:  "12345",
			DestAddr:    "67890",
			Content:     "Test message",
			Encoding:    "GSM7",
			Status:      status,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		repo.Save(msg)
	}

	// Get all
	messages, total, err := repo.GetList(MessageFilter{}, 10, 0)
	if err != nil {
		t.Fatalf("GetList() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %v, want 5", total)
	}
	if len(messages) != 5 {
		t.Errorf("returned %d messages, want 5", len(messages))
	}

	// Filter by status
	pending, total, _ := repo.GetList(MessageFilter{Status: "pending"}, 10, 0)
	if total != 3 {
		t.Errorf("pending total = %v, want 3", total)
	}
	if len(pending) != 3 {
		t.Errorf("returned %d pending messages, want 3", len(pending))
	}
}

func TestMessageRepository_UpdateStatus(t *testing.T) {
	repo := NewMessageRepository(testDB)

	msg := &model.Message{
		ID:          "msg-status-1",
		SessionID:   "session-1",
		MessageID:   "SM-STATUS-001",
		SequenceNum: 1,
		SourceAddr:  "12345",
		DestAddr:    "67890",
		Content:     "Test",
		Encoding:    "GSM7",
		Status:      "pending",
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	repo.Save(msg)

	// Update to delivered
	err := repo.UpdateStatus(msg.ID, "delivered")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, _ := repo.GetByID(msg.ID)
	if got.Status != "delivered" {
		t.Errorf("Status = %v, want delivered", got.Status)
	}
	if got.DeliveredAt == nil {
		t.Error("DeliveredAt should be set for delivered status")
	}
}

func TestMessageRepository_GetStats(t *testing.T) {
	repo := NewMessageRepository(testDB)

	// Clear and add messages
	repo.DeleteAllMessages()

	statuses := []string{"pending", "pending", "delivered", "delivered", "failed"}
	for i, status := range statuses {
		msg := &model.Message{
			ID:          "msg-stats-" + string(rune('a'+i)),
			SessionID:   "session-1",
			MessageID:   "SM-STATS-" + string(rune('0'+i)),
			SequenceNum: uint32(i),
			SourceAddr:  "12345",
			DestAddr:    "67890",
			Content:     "Test",
			Encoding:    "GSM7",
			Status:      status,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		repo.Save(msg)
	}

	stats, err := repo.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.TotalMessages != 5 {
		t.Errorf("TotalMessages = %v, want 5", stats.TotalMessages)
	}
	if stats.PendingMessages != 2 {
		t.Errorf("PendingMessages = %v, want 2", stats.PendingMessages)
	}
	if stats.DeliveredMessages != 2 {
		t.Errorf("DeliveredMessages = %v, want 2", stats.DeliveredMessages)
	}
	if stats.FailedMessages != 1 {
		t.Errorf("FailedMessages = %v, want 1", stats.FailedMessages)
	}
}

func TestMockConfigRepository_GetAndSave(t *testing.T) {
	repo := NewMockConfigRepository(testDB)

	// Get default config
	config, err := repo.Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Modify and save
	config.AutoResponse = false
	config.SuccessRate = 80
	config.ResponseDelay = 100

	err = repo.Save(config)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify
	got, _ := repo.Get()
	if got.AutoResponse != false {
		t.Error("AutoResponse should be false")
	}
	if got.SuccessRate != 80 {
		t.Errorf("SuccessRate = %v, want 80", got.SuccessRate)
	}
}
