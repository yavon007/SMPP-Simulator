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

// MessageTemplate represents a predefined message template
type MessageTemplate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Encoding  string    `json:"encoding"`
	CreatedAt time.Time `json:"created_at"`
}

// OperationLog represents an operation log entry
type OperationLog struct {
	ID          string    `json:"id"`
	Operation   string    `json:"operation"`   // login, send_message, config_change, data_clear
	Content     string    `json:"content"`     // detailed content
	Operator    string    `json:"operator"`
	IP          string    `json:"ip"`          // IP address
	CreatedAt   time.Time `json:"created_at"`
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

// DefaultMessageTemplates returns default message templates
func DefaultMessageTemplates() []MessageTemplate {
	return []MessageTemplate{
		{
			Name:     "问候语",
			Content:  "您好，感谢您的关注！",
			Encoding: "UCS2",
		},
		{
			Name:     "验证码",
			Content:  "您的验证码是：{code}，5分钟内有效。",
			Encoding: "UCS2",
		},
		{
			Name:     "通知模板",
			Content:  "尊敬的用户，您的订单已发货，单号：{order_id}，请注意查收。",
			Encoding: "UCS2",
		},
		{
			Name:     "错误消息",
			Content:  "系统繁忙，请稍后重试。",
			Encoding: "UCS2",
		},
	}
}
