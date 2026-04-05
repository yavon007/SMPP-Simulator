package handler

import (
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"smpp-simulator/internal/smpp"
)

// OutboundHandler handles outbound SMPP connection requests
type OutboundHandler struct {
	client *smpp.Client
}

// NewOutboundHandler creates a new outbound handler
func NewOutboundHandler(client *smpp.Client) *OutboundHandler {
	return &OutboundHandler{client: client}
}

// ConnectRequest represents connect request
type ConnectRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     string `json:"port" binding:"required"`
	SystemID string `json:"system_id" binding:"required"`
	Password string `json:"password"`
	BindType string `json:"bind_type"` // transmitter, receiver, transceiver (default: transceiver)
}

// Connect connects to a remote SMSC
func (h *OutboundHandler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default bind type to transceiver
	bindType := req.BindType
	if bindType == "" {
		bindType = "transceiver"
	}

	// Validate bind type
	if bindType != "transmitter" && bindType != "receiver" && bindType != "transceiver" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bind_type, must be transmitter, receiver, or transceiver"})
		return
	}

	params := &smpp.ConnectParams{
		Host:     req.Host,
		Port:     req.Port,
		SystemID: req.SystemID,
		Password: req.Password,
		BindType: bindType,
	}

	session, err := h.client.Connect(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "connected successfully",
		"data": gin.H{
			"id":           session.ID,
			"system_id":    session.SystemID,
			"bind_type":    session.BindType,
			"remote_addr":  session.RemoteAddr,
			"connected_at": session.ConnectedAt,
			"status":       session.Status,
		},
	})
}

// List returns all outbound sessions
func (h *OutboundHandler) List(c *gin.Context) {
	sessions := h.client.GetAll()
	result := make([]gin.H, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, gin.H{
			"id":           s.ID,
			"system_id":    s.SystemID,
			"bind_type":    s.BindType,
			"remote_addr":  s.RemoteAddr,
			"connected_at": s.ConnectedAt,
			"status":       s.Status,
			"error":        s.ErrorMessage,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete disconnects an outbound session
func (h *OutboundHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.client.Disconnect(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "disconnected"})
}

// SendMessageRequest represents send message request
type OutboundSendMessageRequest struct {
	SourceAddr string `json:"source_addr" binding:"required"`
	DestAddr   string `json:"dest_addr" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Encoding   string `json:"encoding"` // "GSM7" or "UCS2"
}

// GSM7 character set - characters that can be encoded in GSM 7-bit
var gsm7CharsOutbound = map[rune]bool{
	'@': true, '£': true, '$': true, '¥': true, 'è': true, 'é': true, 'ù': true, 'ì': true,
	'ò': true, 'Ç': true, '\n': true, 'Ø': true, 'ø': true, '\r': true, 'Å': true, 'å': true,
	'Δ': true, '_': true, 'Φ': true, 'Γ': true, 'Λ': true, 'Ω': true, 'Π': true, 'Ψ': true,
	'Σ': true, 'Θ': true, 'Ξ': true, 'Æ': true, 'æ': true, 'ß': true, 'É': true,
	' ': true, '!': true, '"': true, '#': true, '¤': true, '%': true, '&': true, '\'': true,
	'(': true, ')': true, '*': true, '+': true, ',': true, '-': true, '.': true, '/': true,
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true,
	'8': true, '9': true, ':': true, ';': true, '<': true, '=': true, '>': true, '?': true,
	'¡': true, 'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true,
	'H': true, 'I': true, 'J': true, 'K': true, 'L': true, 'M': true, 'N': true, 'O': true,
	'P': true, 'Q': true, 'R': true, 'S': true, 'T': true, 'U': true, 'V': true, 'W': true,
	'X': true, 'Y': true, 'Z': true, 'Ä': true, 'Ö': true, 'Ñ': true, 'Ü': true, '§': true,
	'¿': true, 'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true,
	'h': true, 'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true,
	'p': true, 'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true,
	'x': true, 'y': true, 'z': true, 'ä': true, 'ö': true, 'ñ': true, 'ü': true, 'à': true,
}

const (
	GSM7MaxCharsOutbound = 160
	UCS2MaxCharsOutbound  = 70
)

// isGSM7CompatibleOutbound checks if a string can be encoded in GSM7
func isGSM7CompatibleOutbound(s string) bool {
	for _, r := range s {
		if !gsm7CharsOutbound[r] {
			return false
		}
	}
	return true
}

// validateMessageLengthOutbound validates message content length based on encoding
func validateMessageLengthOutbound(content string, encoding string) (int, string) {
	charCount := utf8.RuneCountInString(content)

	switch encoding {
	case "UCS2":
		if charCount > UCS2MaxCharsOutbound {
			return charCount, "UCS2 encoding allows maximum 70 characters"
		}
	case "GSM7", "":
		if encoding == "" && !isGSM7CompatibleOutbound(content) {
			return charCount, "content contains non-GSM7 characters, please use UCS2 encoding"
		}
		if charCount > GSM7MaxCharsOutbound {
			return charCount, "GSM7 encoding allows maximum 160 characters"
		}
	}

	return charCount, ""
}

// SendMessage sends a message through an outbound session
func (h *OutboundHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("id")

	var req OutboundSendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate message length
	charCount, errMsg := validateMessageLengthOutbound(req.Content, req.Encoding)
	if errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      errMsg,
			"char_count": charCount,
		})
		return
	}

	// Determine data coding
	var dataCoding byte = 0 // Default GSM7
	if req.Encoding == "UCS2" {
		dataCoding = 8
	}

	params := &smpp.OutboundSendMessageParams{
		SessionID:  sessionID,
		SourceAddr: req.SourceAddr,
		DestAddr:   req.DestAddr,
		Content:    req.Content,
		DataCoding: dataCoding,
	}

	messageID, err := h.client.SendMessage(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "message sent",
		"message_id": messageID,
		"session_id": sessionID,
		"char_count": charCount,
	})
}
