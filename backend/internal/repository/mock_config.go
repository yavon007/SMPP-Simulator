package repository

import (
	"smpp-simulator/internal/model"
)

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
