package handler

import (
	"fmt"
	"net/http"
	"unicode/utf8"

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

// GSM7 character set - characters that can be encoded in GSM 7-bit
// Note: This is a simplified check. Full GSM7 includes extension characters.
var gsm7Chars = map[rune]bool{
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
	// GSM7MaxChars is the maximum characters for GSM7 encoding in a single SMS
	GSM7MaxChars = 160
	// UCS2MaxChars is the maximum characters for UCS2 encoding in a single SMS
	UCS2MaxChars = 70
)

// isGSM7Compatible checks if a string can be encoded in GSM7
func isGSM7Compatible(s string) bool {
	for _, r := range s {
		if !gsm7Chars[r] {
			return false
		}
	}
	return true
}

// validateMessageLength validates message content length based on encoding
func validateMessageLength(content string, encoding string) (int, string) {
	charCount := utf8.RuneCountInString(content)

	switch encoding {
	case "UCS2":
		if charCount > UCS2MaxChars {
			return charCount, fmt.Sprintf("UCS2 encoding allows maximum %d characters, got %d", UCS2MaxChars, charCount)
		}
	case "GSM7", "":
		// If encoding is not specified, check if content is GSM7 compatible
		if encoding == "" && !isGSM7Compatible(content) {
			return charCount, "content contains non-GSM7 characters, please use UCS2 encoding"
		}
		if charCount > GSM7MaxChars {
			return charCount, fmt.Sprintf("GSM7 encoding allows maximum %d characters, got %d", GSM7MaxChars, charCount)
		}
	}

	return charCount, ""
}

// SendMessage sends a message to a connected session
func (h *SendMessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate message length
	charCount, errMsg := validateMessageLength(req.Content, req.Encoding)
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
		"char_count": charCount,
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
