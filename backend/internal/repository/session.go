package repository

import (
	"smpp-simulator/internal/model"
)

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

	var query string
	switch r.db.dbType {
	case "postgres", "postgresql":
		query = `INSERT INTO sessions (id, system_id, bind_type, remote_addr, connected_at, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET 
		   system_id = EXCLUDED.system_id,
		   bind_type = EXCLUDED.bind_type,
		   remote_addr = EXCLUDED.remote_addr,
		   connected_at = EXCLUDED.connected_at,
		   status = EXCLUDED.status`
	case "mysql":
		query = `INSERT INTO sessions (id, system_id, bind_type, remote_addr, connected_at, status)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE 
		   system_id = VALUES(system_id),
		   bind_type = VALUES(bind_type),
		   remote_addr = VALUES(remote_addr),
		   connected_at = VALUES(connected_at),
		   status = VALUES(status)`
	default: // sqlite
		query = `INSERT OR REPLACE INTO sessions (id, system_id, bind_type, remote_addr, connected_at, status)
		 VALUES (?, ?, ?, ?, ?, ?)`
	}

	_, err := r.db.db.Exec(query,
		session.ID, session.SystemID, session.BindType, session.RemoteAddr, session.ConnectedAt, session.Status,
	)
	return err
}

// GetByID gets a session by ID
func (r *SessionRepository) GetByID(id string) (*model.Session, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := r.db.RebindQuery(`SELECT id, system_id, bind_type, remote_addr, connected_at, status FROM sessions WHERE id = ?`)
	session := &model.Session{}
	err := r.db.db.QueryRow(query, id).Scan(
		&session.ID, &session.SystemID, &session.BindType, &session.RemoteAddr, &session.ConnectedAt, &session.Status,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetAll gets all sessions
func (r *SessionRepository) GetAll() ([]model.Session, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `SELECT id, system_id, bind_type, remote_addr, connected_at, status FROM sessions ORDER BY connected_at DESC`
	rows, err := r.db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]model.Session, 0)
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

	query := r.db.RebindQuery(`UPDATE sessions SET status = ? WHERE id = ?`)
	_, err := r.db.db.Exec(query, status, id)
	return err
}

// Delete deletes a session
func (r *SessionRepository) Delete(id string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := r.db.RebindQuery(`DELETE FROM sessions WHERE id = ?`)
	_, err := r.db.db.Exec(query, id)
	return err
}

// DeleteAllSessions deletes all sessions
func (r *SessionRepository) DeleteAllSessions() error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(`DELETE FROM sessions`)
	return err
}
