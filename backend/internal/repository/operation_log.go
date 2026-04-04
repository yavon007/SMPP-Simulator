package repository

import (
	"database/sql"
	"fmt"
	"time"

	"smpp-simulator/internal/model"
)

// OperationLogRepository handles operation log data
type OperationLogRepository struct {
	db *Database
}

// NewOperationLogRepository creates a new operation log repository
func NewOperationLogRepository(db *Database) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Save saves an operation log
func (r *OperationLogRepository) Save(log *model.OperationLog) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	_, err := r.db.db.Exec(
		`INSERT INTO operation_logs (operation, content, operator, ip, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		log.Operation, log.Content, log.Operator, log.IP, log.CreatedAt,
	)
	return err
}

// OperationLogFilter represents log filter parameters
type OperationLogFilter struct {
	Operation string
	Operator  string
	StartTime string
	EndTime   string
}

// GetList gets a paginated list of operation logs with filters
func (r *OperationLogRepository) GetList(filter OperationLogFilter, limit, offset int) ([]model.OperationLog, int, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	// Build query with filters
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Operation != "" {
		where += " AND operation = ?"
		args = append(args, filter.Operation)
	}
	if filter.Operator != "" {
		where += " AND operator LIKE ?"
		args = append(args, "%"+filter.Operator+"%")
	}
	if filter.StartTime != "" {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", filter.StartTime, time.Local)
		if err == nil {
			where += " AND created_at >= ?"
			args = append(args, startTime.Format(time.RFC3339))
		}
	}
	if filter.EndTime != "" {
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", filter.EndTime, time.Local)
		if err == nil {
			where += " AND created_at <= ?"
			args = append(args, endTime.Format(time.RFC3339))
		}
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM operation_logs %s", where)
	err := r.db.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	query := fmt.Sprintf(
		`SELECT id, operation, content, operator, ip, created_at
		 FROM operation_logs %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		where,
	)
	args = append(args, limit, offset)

	rows, err := r.db.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]model.OperationLog, 0)
	for rows.Next() {
		var log model.OperationLog
		if err := rows.Scan(&log.ID, &log.Operation, &log.Content, &log.Operator, &log.IP, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// DeleteOldLogs deletes logs older than specified days
func (r *OperationLogRepository) DeleteOldLogs(days int) (int64, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	result, err := r.db.db.Exec(`DELETE FROM operation_logs WHERE created_at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete old logs: %w", err)
	}
	return result.RowsAffected()
}

// GetOperationTypes returns distinct operation types
func (r *OperationLogRepository) GetOperationTypes() ([]string, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	rows, err := r.db.db.Query(`SELECT DISTINCT operation FROM operation_logs ORDER BY operation`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]string, 0)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		types = append(types, t)
	}

	return types, nil
}

// LogOperation is a helper function to create and save a log entry
func (r *OperationLogRepository) LogOperation(operation, content, operator, ip string) error {
	log := &model.OperationLog{
		Operation: operation,
		Content:   content,
		Operator:  operator,
		IP:        ip,
		CreatedAt: time.Now(),
	}
	return r.Save(log)
}

// GetDB returns the underlying database for table creation
func (r *OperationLogRepository) GetDB() *sql.DB {
	return r.db.db
}
