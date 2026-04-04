package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/model"
	"smpp-simulator/internal/repository"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	repo *repository.MessageRepository
}

// BatchDeleteRequest represents the request body for batch delete
type BatchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(repo *repository.MessageRepository) *MessageHandler {
	return &MessageHandler{repo: repo}
}

// List returns messages with pagination and filters
func (h *MessageHandler) List(c *gin.Context) {
	// Parse query parameters
	filter := repository.MessageFilter{
		SessionID:  c.Query("session_id"),
		Status:     c.Query("status"),
		SourceAddr: c.Query("source_addr"),
		DestAddr:   c.Query("dest_addr"),
		Content:    c.Query("content"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	messages, total, err := h.repo.GetList(filter, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      messages,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Get returns a single message by ID
func (h *MessageHandler) Get(c *gin.Context) {
	id := c.Param("id")

	msg, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	c.JSON(http.StatusOK, msg)
}

// Deliver triggers a delivery report for a message
func (h *MessageHandler) Deliver(c *gin.Context) {
	h.updateStatus(c, "delivered")
}

// Fail marks a message as failed
func (h *MessageHandler) Fail(c *gin.Context) {
	h.updateStatus(c, "failed")
}

// updateStatus updates message status
func (h *MessageHandler) updateStatus(c *gin.Context, status string) {
	id := c.Param("id")

	// Get message
	msg, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	// Update status
	if err := h.repo.UpdateStatus(id, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "status updated",
		"message_id":  msg.MessageID,
		"source_addr": msg.SourceAddr,
		"dest_addr":   msg.DestAddr,
		"status":      status,
	})
}

// Export exports messages in CSV or JSON format
func (h *MessageHandler) Export(c *gin.Context) {
	// Parse query parameters
	filter := repository.MessageFilter{
		SessionID:  c.Query("session_id"),
		Status:     c.Query("status"),
		SourceAddr: c.Query("source_addr"),
		DestAddr:   c.Query("dest_addr"),
		Content:    c.Query("content"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
	}

	// Get all messages (no pagination for export)
	messages, _, err := h.repo.GetList(filter, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	format := c.DefaultQuery("format", "csv")
	switch strings.ToLower(format) {
	case "json":
		h.exportJSON(c, messages)
	default:
		h.exportCSV(c, messages)
	}
}

// exportCSV exports messages as CSV format
func (h *MessageHandler) exportCSV(c *gin.Context, messages []repository.MessageWithDetails) {
	// Set response headers for file download
	filename := fmt.Sprintf("messages_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Add UTF-8 BOM for Excel compatibility
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	header := []string{"消息ID", "发送方", "接收方", "内容", "状态", "时间"}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write CSV header"})
		return
	}

	// Write data rows
	for _, msg := range messages {
		statusText := getStatusText(msg.Status)
		row := []string{
			msg.MessageID,
			msg.SourceAddr,
			msg.DestAddr,
			msg.Content,
			statusText,
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write CSV row"})
			return
		}
	}
}

// exportJSON exports messages as JSON format
func (h *MessageHandler) exportJSON(c *gin.Context, messages []repository.MessageWithDetails) {
	// Set response headers for file download
	filename := fmt.Sprintf("messages_%s.json", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Convert to export format
	exportData := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		exportData = append(exportData, map[string]interface{}{
			"message_id":  msg.MessageID,
			"source_addr": msg.SourceAddr,
			"dest_addr":   msg.DestAddr,
			"content":     msg.Content,
			"encoding":    msg.Encoding,
			"status":      msg.Status,
			"created_at":  msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	encoder := json.NewEncoder(c.Writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode JSON"})
		return
	}
}

// getStatusText converts status code to Chinese text
func getStatusText(status string) string {
	switch status {
	case "pending":
		return "待处理"
	case "delivered":
		return "已送达"
	case "failed":
		return "失败"
	default:
		return status
	}
}

// BatchDelete deletes multiple messages by IDs
func (h *MessageHandler) BatchDelete(c *gin.Context) {
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: ids array is required"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids array cannot be empty"})
		return
	}

	// Limit batch size to prevent abuse
	if len(req.IDs) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete more than 1000 messages at once"})
		return
	}

	deleted, err := h.repo.DeleteByIDs(req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "batch delete completed",
		"deleted_count": deleted,
	})
}
