package repository

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
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

// ClearAllData deletes all messages and sessions
func (d *Database) ClearAllData() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`DELETE FROM messages`)
	if err != nil {
		return err
	}
	_, err = d.db.Exec(`DELETE FROM sessions`)
	return err
}
