package repository

import (
	"database/sql"
	"fmt"

	"smpp-simulator/internal/model"
)

// TemplateRepository handles message template data
type TemplateRepository struct {
	db *Database
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *Database) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Save saves a message template
func (r *TemplateRepository) Save(template *model.MessageTemplate) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := r.db.RebindQuery(`INSERT INTO message_templates (id, name, content, encoding, created_at)
		 VALUES (?, ?, ?, ?, ?)`)
	_, err := r.db.db.Exec(query,
		template.ID, template.Name, template.Content, template.Encoding, template.CreatedAt,
	)
	return err
}

// GetByID gets a template by ID
func (r *TemplateRepository) GetByID(id string) (*model.MessageTemplate, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	template := &model.MessageTemplate{}
	query := r.db.RebindQuery(`SELECT id, name, content, encoding, created_at
		 FROM message_templates WHERE id = ?`)
	err := r.db.db.QueryRow(query, id).Scan(
		&template.ID, &template.Name, &template.Content, &template.Encoding, &template.CreatedAt)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// GetList gets all templates
func (r *TemplateRepository) GetList() ([]model.MessageTemplate, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `SELECT id, name, content, encoding, created_at
		FROM message_templates ORDER BY created_at DESC`

	rows, err := r.db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get templates: %w", err)
	}
	defer rows.Close()

	templates := make([]model.MessageTemplate, 0)
	for rows.Next() {
		var t model.MessageTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Content, &t.Encoding, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan template row: %w", err)
		}
		templates = append(templates, t)
	}

	return templates, nil
}

// Update updates a template
func (r *TemplateRepository) Update(template *model.MessageTemplate) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := r.db.RebindQuery(`UPDATE message_templates SET name = ?, content = ?, encoding = ? WHERE id = ?`)
	result, err := r.db.db.Exec(query,
		template.Name, template.Content, template.Encoding, template.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deletes a template by ID
func (r *TemplateRepository) Delete(id string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := r.db.RebindQuery(`DELETE FROM message_templates WHERE id = ?`)
	_, err := r.db.db.Exec(query, id)
	return err
}
