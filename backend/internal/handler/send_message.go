package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/smpp"
)

// SendMessageHandler handles send message requests
type SendMessageHandler struct {
	smppServer *smpp.Server
}

// NewSendMessageHandler creates a new send message handler
func NewSendMessageHandler(smppServer *smpp.Server) *SendMessageHandler {
	return &SendMessageHandler{smppServer: smppServer}
}

// SendMessageRequest represents send message request
type SendMessageRequest struct {
	SessionID  string `json:"session_id" binding:"required"`
	SourceAddr string `json:"source_addr" binding:"required"`
	DestAddr   string `json:"dest_addr" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Encoding   string `json:"encoding"` // "GSM7" or "UCS2"
}

// SendMessage sends a message to a connected session
func (h *SendMessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine data coding
	var dataCoding byte = 0 // Default GSM7
	if req.Encoding == "UCS2" {
		dataCoding = 8
	}

	params := &smpp.SendMessageParams{
		SessionID:  req.SessionID,
		SourceAddr: req.SourceAddr,
		DestAddr:   req.DestAddr,
		Content:    req.Content,
		DataCoding: dataCoding,
	}

	if err := h.smppServer.SendMessage(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "message sent",
		"session_id": req.SessionID,
	})
}

// ListReceivers returns all sessions that can receive messages
func (h *SendMessageHandler) ListReceivers(c *gin.Context) {
	receivers := h.smppServer.GetReceivers()
	result := make([]gin.H, 0, len(receivers))
	for _, s := range receivers {
		result = append(result, gin.H{
			"id":          s.ID,
			"system_id":   s.SystemID,
			"bind_type":   s.BindType,
			"remote_addr": s.RemoteAddr,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
