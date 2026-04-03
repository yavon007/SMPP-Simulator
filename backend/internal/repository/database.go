package repository

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// Database represents a database connection
type Database struct {
	db     *sql.DB
	mu     sync.RWMutex
	dbType string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string
	Path     string // for SQLite
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// NewDatabase creates a new database connection based on config
func NewDatabase(cfg *DatabaseConfig) (*Database, error) {
	var db *sql.DB
	var err error
	var driverName string
	var dsn string

	switch cfg.Type {
	case "sqlite", "":
		driverName = "sqlite"
		dsn = cfg.Path
		if dsn == "" {
			dsn = "./smpp.db"
		}
	case "postgres", "postgresql":
		driverName = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	case "mysql":
		driverName = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err = sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// SQLite-specific optimizations
	if cfg.Type == "sqlite" || cfg.Type == "" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	database := &Database{db: db, dbType: cfg.Type}
	if err := database.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

// NewDatabaseFromPath creates a new SQLite database (backward compatible)
func NewDatabaseFromPath(dbPath string) (*Database, error) {
	return NewDatabase(&DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	})
}

// createTables creates the database tables
func (d *Database) createTables() error {
	// Use appropriate syntax for each database
	var autoIncrement string
	var timestampType string

	switch d.dbType {
	case "postgres", "postgresql":
		autoIncrement = "SERIAL"
		timestampType = "TIMESTAMP"
	case "mysql":
		autoIncrement = "INT AUTO_INCREMENT"
		timestampType = "DATETIME"
	default: // sqlite
		autoIncrement = "INTEGER"
		timestampType = "DATETIME"
	}

	queries := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			system_id TEXT NOT NULL,
			bind_type TEXT NOT NULL,
			remote_addr TEXT NOT NULL,
			connected_at %s NOT NULL,
			status TEXT NOT NULL DEFAULT 'active'
		)`, timestampType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			sequence_num INTEGER NOT NULL,
			source_addr TEXT NOT NULL,
			dest_addr TEXT NOT NULL,
			content TEXT,
			encoding TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at %s NOT NULL,
			delivered_at %s
		)`, timestampType, timestampType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS mock_config (
			id %s PRIMARY KEY CHECK (id = 1),
			auto_response INTEGER NOT NULL DEFAULT 1,
			success_rate INTEGER NOT NULL DEFAULT 100,
			response_delay INTEGER NOT NULL DEFAULT 0,
			deliver_report INTEGER NOT NULL DEFAULT 0,
			deliver_delay INTEGER NOT NULL DEFAULT 1000
		)`, autoIncrement),
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
	switch d.dbType {
	case "postgres", "postgresql":
		_, _ = d.db.Exec(`INSERT INTO mock_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING`)
	case "mysql":
		_, _ = d.db.Exec(`INSERT IGNORE INTO mock_config (id) VALUES (1)`)
	default: // sqlite
		_, _ = d.db.Exec(`INSERT OR IGNORE INTO mock_config (id) VALUES (1)`)
	}

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

// GetDB returns the underlying database connection
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Type returns the database type
func (d *Database) Type() string {
	return d.dbType
}
