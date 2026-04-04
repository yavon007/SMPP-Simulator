package repository

import (
	"database/sql"
	"fmt"
	"time"

	"smpp-simulator/internal/model"
)

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
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

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
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

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
	SessionID  string
	Status     string
	SourceAddr string
	DestAddr   string
	Content    string
	StartTime  string
	EndTime    string
}

// GetList gets a paginated list of messages with filters
func (r *MessageRepository) GetList(filter MessageFilter, limit, offset int) ([]model.Message, int, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

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
	if filter.Content != "" {
		where += " AND content LIKE ?"
		args = append(args, "%"+filter.Content+"%")
	}
	if filter.StartTime != "" {
		// 解析前端传来的时间（本地时间格式）并转换为 RFC3339 格式
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", filter.StartTime, time.Local)
		if err == nil {
			where += " AND created_at >= ?"
			args = append(args, startTime.Format(time.RFC3339))
		}
	}
	if filter.EndTime != "" {
		// 解析前端传来的时间（本地时间格式）并转换为 RFC3339 格式
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", filter.EndTime, time.Local)
		if err == nil {
			where += " AND created_at <= ?"
			args = append(args, endTime.Format(time.RFC3339))
		}
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

	messages := make([]model.Message, 0)
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

	if status == "delivered" {
		now := time.Now()
		_, err := r.db.db.Exec(
			`UPDATE messages SET status = ?, delivered_at = ? WHERE id = ?`,
			status, now, id,
		)
		return err
	}

	_, err := r.db.db.Exec(`UPDATE messages SET status = ? WHERE id = ?`, status, id)
	return err
}

// GetStats gets message statistics
func (r *MessageRepository) GetStats() (*model.Stats, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

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

// DeleteMessage deletes a message by ID
func (r *MessageRepository) DeleteMessage(id string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(`DELETE FROM messages WHERE id = ?`, id)
	return err
}

// DeleteAllMessages deletes all messages
func (r *MessageRepository) DeleteAllMessages() error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(`DELETE FROM messages`)
	return err
}
