package smpp

import (
	"net"
	"sync"
	"time"
)

// SessionState represents the state of an SMPP session
type SessionState struct {
	ID          string
	SystemID    string
	Password    string
	BindType    string // transmitter, receiver, transceiver
	RemoteAddr  string
	Conn        net.Conn
	ConnectedAt time.Time
	Status      string // active, closed
	SequenceNum uint32
	mu          sync.Mutex
}

// NewSessionState creates a new session state
func NewSessionState(conn net.Conn) *SessionState {
	return &SessionState{
		ID:          generateID(),
		Conn:        conn,
		RemoteAddr:  conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
		Status:      "active",
		SequenceNum: 0,
	}
}

// NextSequenceNum returns the next sequence number
func (s *SessionState) NextSequenceNum() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SequenceNum++
	return s.SequenceNum
}

// Close closes the session
func (s *SessionState) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Conn != nil {
		s.Conn.Close()
	}
	s.Status = "closed"
}

// SetBindInfo sets bind information
func (s *SessionState) SetBindInfo(systemID, password, bindType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SystemID = systemID
	s.Password = password
	s.BindType = bindType
}

// SessionManager manages all active sessions
type SessionManager struct {
	sessions map[string]*SessionState
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionState),
	}
}

// Add adds a session
func (m *SessionManager) Add(session *SessionState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
}

// Remove removes a session
func (m *SessionManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

// Get gets a session by ID
func (m *SessionManager) Get(id string) *SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// GetAll returns all sessions
func (m *SessionManager) GetAll() []*SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*SessionState, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

// Count returns the number of active sessions
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// GetReceivers returns all sessions that can receive messages (receiver or transceiver)
func (m *SessionManager) GetReceivers() []*SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*SessionState, 0)
	for _, s := range m.sessions {
		if s.BindType == "receiver" || s.BindType == "transceiver" {
			result = append(result, s)
		}
	}
	return result
}

// generateID generates a unique session ID
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}
