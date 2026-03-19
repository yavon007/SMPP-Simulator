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

	_, err := r.db.db.Exec(
		`INSERT OR REPLACE INTO sessions (id, system_id, bind_type, remote_addr, connected_at, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID, session.SystemID, session.BindType, session.RemoteAddr, session.ConnectedAt, session.Status,
	)
	return err
}

// GetByID gets a session by ID
func (r *SessionRepository) GetByID(id string) (*model.Session, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

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
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	rows, err := r.db.db.Query(
		`SELECT id, system_id, bind_type, remote_addr, connected_at, status FROM sessions ORDER BY connected_at DESC`,
	)
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

// DeleteAllSessions deletes all sessions
func (r *SessionRepository) DeleteAllSessions() error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(`DELETE FROM sessions`)
	return err
}
