package model

import "time"

// Session represents an SMPP connection session
type Session struct {
	ID           string     `json:"id"`
	SystemID     string     `json:"system_id"`
	Password     string     `json:"-"`
	BindType     string     `json:"bind_type"`
	RemoteAddr   string     `json:"remote_addr"`
	ConnectedAt  time.Time  `json:"connected_at"`
	Status       string     `json:"status"` // active, closed
}

// Message represents an SMS message
type Message struct {
	ID          string     `json:"id"`
	SessionID   string     `json:"session_id"`
	MessageID   string     `json:"message_id"`   // SMPP message_id returned to client
	SequenceNum uint32     `json:"sequence_num"` // PDU sequence number
	SourceAddr  string     `json:"source_addr"`
	DestAddr    string     `json:"dest_addr"`
	Content     string     `json:"content"`
	Encoding    string     `json:"encoding"` // ASCII, UCS2, GSM7
	Status      string     `json:"status"`   // pending, delivered, failed
	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

// MockConfig represents simulation behavior configuration
type MockConfig struct {
	AutoResponse  bool `json:"auto_response"`  // Auto respond to submit_sm
	SuccessRate   int  `json:"success_rate"`   // Success rate (0-100)
	ResponseDelay int  `json:"response_delay"` // Response delay in ms
	DeliverReport bool `json:"deliver_report"` // Auto send delivery report
	DeliverDelay  int  `json:"deliver_delay"`  // Delivery report delay in ms
}

// Stats represents system statistics
type Stats struct {
	ActiveConnections int `json:"active_connections"`
	TotalMessages     int `json:"total_messages"`
	PendingMessages   int `json:"pending_messages"`
	DeliveredMessages int `json:"delivered_messages"`
	FailedMessages    int `json:"failed_messages"`
}

// DefaultMockConfig returns default mock configuration
func DefaultMockConfig() *MockConfig {
	return &MockConfig{
		AutoResponse:  true,
		SuccessRate:   100,
		ResponseDelay: 0,
		DeliverReport: false,
		DeliverDelay:  1000,
	}
}
