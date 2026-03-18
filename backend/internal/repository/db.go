package repository

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"smpp-simulator/internal/model"
)

// Database represents the SQLite database
type Database struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

// createTables creates the database tables
func (d *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			system_id TEXT NOT NULL,
			bind_type TEXT NOT NULL,
			remote_addr TEXT NOT NULL,
			connected_at DATETIME NOT NULL,
			status TEXT NOT NULL DEFAULT 'active'
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			sequence_num INTEGER NOT NULL,
			source_addr TEXT NOT NULL,
			dest_addr TEXT NOT NULL,
			content TEXT,
			encoding TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			delivered_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS mock_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			auto_response INTEGER NOT NULL DEFAULT 1,
			success_rate INTEGER NOT NULL DEFAULT 100,
			response_delay INTEGER NOT NULL DEFAULT 0,
			deliver_report INTEGER NOT NULL DEFAULT 0,
			deliver_delay INTEGER NOT NULL DEFAULT 1000
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return err
		}
	}

	// Insert default mock config if not exists
	_, _ = d.db.Exec(`INSERT OR IGNORE INTO mock_config (id) VALUES (1)`)

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// SessionRepository handles session data
type SessionRepository struct {
	db *Database
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// Save saves a session
func (r *SessionRepository) Save(session *model.Session) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(
		`INSERT OR REPLACE INTO sessions (id, system_id, bind_type, remote_addr, connected_at, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID, session.SystemID, session.BindType, session.RemoteAddr, session.ConnectedAt, session.Status,
	)
	return err
}

// GetByID gets a session by ID
func (r *SessionRepository) GetByID(id string) (*model.Session, error) {
	session := &model.Session{}
	err := r.db.db.QueryRow(
		`SELECT id, system_id, bind_type, remote_addr, connected_at, status FROM sessions WHERE id = ?`,
		id,
	).Scan(&session.ID, &session.SystemID, &session.BindType, &session.RemoteAddr, &session.ConnectedAt, &session.Status)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetAll gets all sessions
func (r *SessionRepository) GetAll() ([]model.Session, error) {
	rows, err := r.db.db.Query(
		`SELECT id, system_id, bind_type, remote_addr, connected_at, status FROM sessions ORDER BY connected_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		if err := rows.Scan(&s.ID, &s.SystemID, &s.BindType, &s.RemoteAddr, &s.ConnectedAt, &s.Status); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// UpdateStatus updates session status
func (r *SessionRepository) UpdateStatus(id string, status string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(`UPDATE sessions SET status = ? WHERE id = ?`, status, id)
	return err
}

// Delete deletes a session
func (r *SessionRepository) Delete(id string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// MessageRepository handles message data
type MessageRepository struct {
	db *Database
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *Database) *MessageRepository {
	return &MessageRepository{db: db}
}

// Save saves a message
func (r *MessageRepository) Save(msg *model.Message) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(
		`INSERT INTO messages (id, session_id, message_id, sequence_num, source_addr, dest_addr, content, encoding, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.SessionID, msg.MessageID, msg.SequenceNum, msg.SourceAddr, msg.DestAddr, msg.Content, msg.Encoding, msg.Status, msg.CreatedAt,
	)
	return err
}

// GetByID gets a message by ID
func (r *MessageRepository) GetByID(id string) (*model.Message, error) {
	msg := &model.Message{}
	var deliveredAt sql.NullTime
	err := r.db.db.QueryRow(
		`SELECT id, session_id, message_id, sequence_num, source_addr, dest_addr, content, encoding, status, created_at, delivered_at
		 FROM messages WHERE id = ?`,
		id,
	).Scan(&msg.ID, &msg.SessionID, &msg.MessageID, &msg.SequenceNum, &msg.SourceAddr, &msg.DestAddr,
		&msg.Content, &msg.Encoding, &msg.Status, &msg.CreatedAt, &deliveredAt)
	if err != nil {
		return nil, err
	}
	if deliveredAt.Valid {
		msg.DeliveredAt = &deliveredAt.Time
	}
	return msg, nil
}

// GetByMessageID gets a message by SMPP message_id
func (r *MessageRepository) GetByMessageID(messageID string) (*model.Message, error) {
	msg := &model.Message{}
	var deliveredAt sql.NullTime
	err := r.db.db.QueryRow(
		`SELECT id, session_id, message_id, sequence_num, source_addr, dest_addr, content, encoding, status, created_at, delivered_at
		 FROM messages WHERE message_id = ?`,
		messageID,
	).Scan(&msg.ID, &msg.SessionID, &msg.MessageID, &msg.SequenceNum, &msg.SourceAddr, &msg.DestAddr,
		&msg.Content, &msg.Encoding, &msg.Status, &msg.CreatedAt, &deliveredAt)
	if err != nil {
		return nil, err
	}
	if deliveredAt.Valid {
		msg.DeliveredAt = &deliveredAt.Time
	}
	return msg, nil
}

// MessageFilter represents message filter parameters
type MessageFilter struct {
	SessionID string
	Status    string
	SourceAddr string
	DestAddr   string
	StartTime  string
	EndTime    string
}

// GetList gets a paginated list of messages with filters
func (r *MessageRepository) GetList(filter MessageFilter, limit, offset int) ([]model.Message, int, error) {
	// Build query with filters
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.SessionID != "" {
		where += " AND session_id = ?"
		args = append(args, filter.SessionID)
	}
	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.SourceAddr != "" {
		where += " AND source_addr LIKE ?"
		args = append(args, "%"+filter.SourceAddr+"%")
	}
	if filter.DestAddr != "" {
		where += " AND dest_addr LIKE ?"
		args = append(args, "%"+filter.DestAddr+"%")
	}
	if filter.StartTime != "" {
		where += " AND created_at >= ?"
		args = append(args, filter.StartTime)
	}
	if filter.EndTime != "" {
		where += " AND created_at <= ?"
		args = append(args, filter.EndTime)
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM messages %s", where)
	err := r.db.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get messages
	query := fmt.Sprintf(
		`SELECT id, session_id, message_id, sequence_num, source_addr, dest_addr, content, encoding, status, created_at, delivered_at
		 FROM messages %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		where,
	)
	args = append(args, limit, offset)

	rows, err := r.db.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var deliveredAt sql.NullTime
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.MessageID, &msg.SequenceNum, &msg.SourceAddr,
			&msg.DestAddr, &msg.Content, &msg.Encoding, &msg.Status, &msg.CreatedAt, &deliveredAt); err != nil {
			return nil, 0, err
		}
		if deliveredAt.Valid {
			msg.DeliveredAt = &deliveredAt.Time
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// UpdateStatus updates message status
func (r *MessageRepository) UpdateStatus(id string, status string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	var deliveredAt interface{}
	if status == "delivered" {
		now := time.Now()
		deliveredAt = now
		_, err := r.db.db.Exec(
			`UPDATE messages SET status = ?, delivered_at = ? WHERE id = ?`,
			status, deliveredAt, id,
		)
		return err
	}

	_, err := r.db.db.Exec(`UPDATE messages SET status = ? WHERE id = ?`, status, id)
	return err
}

// GetStats gets message statistics
func (r *MessageRepository) GetStats() (*model.Stats, error) {
	stats := &model.Stats{}

	err := r.db.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&stats.TotalMessages)
	if err != nil {
		return nil, err
	}

	err = r.db.db.QueryRow(`SELECT COUNT(*) FROM messages WHERE status = 'pending'`).Scan(&stats.PendingMessages)
	if err != nil {
		return nil, err
	}

	err = r.db.db.QueryRow(`SELECT COUNT(*) FROM messages WHERE status = 'delivered'`).Scan(&stats.DeliveredMessages)
	if err != nil {
		return nil, err
	}

	err = r.db.db.QueryRow(`SELECT COUNT(*) FROM messages WHERE status = 'failed'`).Scan(&stats.FailedMessages)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// MockConfigRepository handles mock configuration
type MockConfigRepository struct {
	db *Database
}

// NewMockConfigRepository creates a new mock config repository
func NewMockConfigRepository(db *Database) *MockConfigRepository {
	return &MockConfigRepository{db: db}
}

// Get gets the mock configuration
func (r *MockConfigRepository) Get() (*model.MockConfig, error) {
	config := &model.MockConfig{}
	err := r.db.db.QueryRow(
		`SELECT auto_response, success_rate, response_delay, deliver_report, deliver_delay FROM mock_config WHERE id = 1`,
	).Scan(&config.AutoResponse, &config.SuccessRate, &config.ResponseDelay, &config.DeliverReport, &config.DeliverDelay)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// Save saves the mock configuration
func (r *MockConfigRepository) Save(config *model.MockConfig) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(
		`UPDATE mock_config SET auto_response = ?, success_rate = ?, response_delay = ?, deliver_report = ?, deliver_delay = ? WHERE id = 1`,
		config.AutoResponse, config.SuccessRate, config.ResponseDelay, config.DeliverReport, config.DeliverDelay,
	)
	return err
}
