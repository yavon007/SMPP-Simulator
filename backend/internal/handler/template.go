package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"smpp-simulator/internal/model"
	"smpp-simulator/internal/repository"
)

// TemplateHandler handles template-related requests
type TemplateHandler struct {
	repo *repository.TemplateRepository
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(repo *repository.TemplateRepository) *TemplateHandler {
	return &TemplateHandler{repo: repo}
}

// TemplateCreateRequest represents the request body for creating a template
type TemplateCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Encoding string `json:"encoding"`
}

// TemplateUpdateRequest represents the request body for updating a template
type TemplateUpdateRequest struct {
	Name     string `json:"name" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Encoding string `json:"encoding"`
}

// List returns all templates
func (h *TemplateHandler) List(c *gin.Context) {
	templates, err := h.repo.GetList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": templates,
	})
}

// Get returns a single template by ID
func (h *TemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")

	template, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// Create creates a new template
func (h *TemplateHandler) Create(c *gin.Context) {
	var req TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: name and content are required"})
		return
	}

	// Set default encoding
	encoding := req.Encoding
	if encoding == "" {
		encoding = "GSM7"
	}

	template := &model.MessageTemplate{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Content:   req.Content,
		Encoding:  encoding,
		CreatedAt: time.Now(),
	}

	if err := h.repo.Save(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// Update updates an existing template
func (h *TemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: name and content are required"})
		return
	}

	// Check if template exists
	existing, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	// Update fields
	existing.Name = req.Name
	existing.Content = req.Content
	if req.Encoding != "" {
		existing.Encoding = req.Encoding
	}

	if err := h.repo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// Delete deletes a template
func (h *TemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// Check if template exists
	_, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}
