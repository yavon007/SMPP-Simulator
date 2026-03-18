package smpp

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"unicode/utf16"

	"smpp-simulator/internal/model"
)

// Server represents the SMPP server
type Server struct {
	host          string
	port          string
	listener      net.Listener
	sessions      *SessionManager
	messageStore  MessageStore
	mockConfig    *model.MockConfig
	configMu      sync.RWMutex
	eventHandler  EventHandler
	running       bool
}

// MessageStore interface for message storage
type MessageStore interface {
	Save(msg *model.Message) error
	GetByID(id string) (*model.Message, error)
	UpdateStatus(id string, status string) error
}

// EventHandler interface for events
type EventHandler interface {
	OnSessionConnect(session *model.Session)
	OnSessionDisconnect(sessionID string)
	OnMessageReceived(msg *model.Message)
	OnMessageDelivered(msgID string)
}

// NewServer creates a new SMPP server
func NewServer(host, port string, store MessageStore) *Server {
	return &Server{
		host:         host,
		port:         port,
		sessions:     NewSessionManager(),
		messageStore: store,
		mockConfig:   model.DefaultMockConfig(),
	}
}

// SetEventHandler sets the event handler
func (s *Server) SetEventHandler(handler EventHandler) {
	s.eventHandler = handler
}

// SetMockConfig sets the mock configuration
func (s *Server) SetMockConfig(config *model.MockConfig) {
	s.configMu.Lock()
	defer s.configMu.Unlock()
	s.mockConfig = config
}

// GetMockConfig gets the mock configuration
func (s *Server) GetMockConfig() *model.MockConfig {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	return s.mockConfig
}

// Start starts the SMPP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener
	s.running = true

	log.Printf("SMPP server started on %s", addr)

	go s.acceptLoop()
	return nil
}

// Stop stops the SMPP server
func (s *Server) Stop() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	// Close all sessions
	for _, session := range s.sessions.GetAll() {
		session.Close()
	}
	log.Println("SMPP server stopped")
}

// acceptLoop accepts incoming connections
func (s *Server) acceptLoop() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("Accept error: %v", err)
			}
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	session := NewSessionState(conn)
	s.sessions.Add(session)

	log.Printf("New connection from %s (session: %s)", session.RemoteAddr, session.ID)

	reader := bufio.NewReader(conn)
	defer func() {
		session.Close()
		s.sessions.Remove(session.ID)
		if s.eventHandler != nil {
			s.eventHandler.OnSessionDisconnect(session.ID)
		}
		log.Printf("Connection closed: %s (session: %s)", session.RemoteAddr, session.ID)
	}()

	for s.running {
		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// Read PDU length (first 4 bytes)
		lengthBuf := make([]byte, 4)
		_, err := reader.Read(lengthBuf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && s.running {
				log.Printf("Read error from %s: %v", session.RemoteAddr, err)
			}
			return
		}

		// Decode length
		length := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])
		if length < 16 || length > 10240 {
			log.Printf("Invalid PDU length %d from %s", length, session.RemoteAddr)
			return
		}

		// Read remaining PDU
		remaining := length - 4
		pduBuf := make([]byte, remaining)
		_, err = reader.Read(pduBuf)
		if err != nil {
			log.Printf("Read error from %s: %v", session.RemoteAddr, err)
			return
		}

		// Combine and decode PDU
		fullPDU := append(lengthBuf, pduBuf...)
		pdu, err := DecodePDU(fullPDU)
		if err != nil {
			log.Printf("PDU decode error from %s: %v", session.RemoteAddr, err)
			continue
		}

		log.Printf("Received PDU from %s: %s (seq: %d)", session.RemoteAddr, CommandName(pdu.CommandID), pdu.SequenceNum)

		// Handle PDU
		response := s.handlePDU(session, pdu)
		if response != nil {
			if _, err := conn.Write(response); err != nil {
				log.Printf("Write error to %s: %v", session.RemoteAddr, err)
				return
			}
			log.Printf("Sent response to %s: %x", session.RemoteAddr, hex.EncodeToString(response[:20]))
		}
	}
}

// handlePDU handles a PDU and returns response
func (s *Server) handlePDU(session *SessionState, pdu *PDU) []byte {
	switch pdu.CommandID {
	case CmdBindTransmitter:
		return s.handleBind(session, pdu, "transmitter")
	case CmdBindReceiver:
		return s.handleBind(session, pdu, "receiver")
	case CmdBindTransceiver:
		return s.handleBind(session, pdu, "transceiver")
	case CmdUnbind:
		return s.handleUnbind(session, pdu)
	case CmdSubmitSM:
		return s.handleSubmitSM(session, pdu)
	case CmdEnquireLink:
		return EncodeEnquireLinkResp(pdu.SequenceNum)
	default:
		log.Printf("Unhandled PDU command: %s", CommandName(pdu.CommandID))
		return nil
	}
}

// handleBind handles bind request
func (s *Server) handleBind(session *SessionState, pdu *PDU, bindType string) []byte {
	params := DecodeBind(pdu.Body)
	session.SetBindInfo(params.SystemID, params.Password, bindType)

	log.Printf("Bind %s from %s: system_id=%s", bindType, session.RemoteAddr, params.SystemID)

	// Notify event handler
	if s.eventHandler != nil {
		s.eventHandler.OnSessionConnect(&model.Session{
			ID:          session.ID,
			SystemID:    params.SystemID,
			BindType:    bindType,
			RemoteAddr:  session.RemoteAddr,
			ConnectedAt: session.ConnectedAt,
			Status:      "active",
		})
	}

	// Always accept bind (no authentication in this simulator)
	switch bindType {
	case "transmitter":
		return EncodeBindTransmitterResp("SMSC", pdu.SequenceNum, StatusOK)
	case "receiver":
		return EncodeBindReceiverResp("SMSC", pdu.SequenceNum, StatusOK)
	case "transceiver":
		return EncodeBindTransceiverResp("SMSC", pdu.SequenceNum, StatusOK)
	}
	return nil
}

// handleUnbind handles unbind request
func (s *Server) handleUnbind(session *SessionState, pdu *PDU) []byte {
	log.Printf("Unbind from %s (session: %s)", session.RemoteAddr, session.ID)
	return EncodeUnbindResp(pdu.SequenceNum)
}

// handleSubmitSM handles submit_sm request
func (s *Server) handleSubmitSM(session *SessionState, pdu *PDU) []byte {
	params := DecodeSubmitSM(pdu.Body)

	// Generate message ID
	messageID := generateMessageID()

	// Determine encoding and decode message content properly
	var content string
	var encoding string

	switch params.DataCoding {
	case 8: // UCS2 (UTF-16BE)
		encoding = "UCS2"
		content = decodeUCS2(params.ShortMessageBytes)
	case 0: // GSM7 or ASCII
		encoding = "GSM7"
		content = params.ShortMessage
	default: // Default to ASCII/latin1
		encoding = "ASCII"
		content = params.ShortMessage
	}

	// Create message record
	msg := &model.Message{
		ID:          generateMessageID(),
		SessionID:   session.ID,
		MessageID:   messageID,
		SequenceNum: pdu.SequenceNum,
		SourceAddr:  params.SourceAddr,
		DestAddr:    params.DestAddr,
		Content:     content,
		Encoding:    encoding,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	// Save message
	if s.messageStore != nil {
		if err := s.messageStore.Save(msg); err != nil {
			log.Printf("Failed to save message: %v", err)
		}
	}

	// Notify event handler
	if s.eventHandler != nil {
		s.eventHandler.OnMessageReceived(msg)
	}

	log.Printf("SubmitSM from %s: src=%s, dst=%s, msg_id=%s",
		session.RemoteAddr, params.SourceAddr, params.DestAddr, messageID)

	// Apply mock config
	config := s.GetMockConfig()

	// Simulate response delay
	if config.ResponseDelay > 0 {
		time.Sleep(time.Duration(config.ResponseDelay) * time.Millisecond)
	}

	// Determine response status
	var status uint32 = StatusOK
	if !config.AutoResponse {
		// Don't respond immediately
		return nil
	}

	// Check success rate
	if config.SuccessRate < 100 {
		// Random check would go here, for now just use 100% success
	}

	// Send response
	response := EncodeSubmitSMResp(messageID, pdu.SequenceNum, status)

	// Schedule delivery report if configured
	if config.DeliverReport {
		go s.scheduleDeliveryReport(session, msg, config.DeliverDelay)
	}

	return response
}

// scheduleDeliveryReport schedules a delivery report
func (s *Server) scheduleDeliveryReport(session *SessionState, msg *model.Message, delayMs int) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Update message status
	if s.messageStore != nil {
		s.messageStore.UpdateStatus(msg.ID, "delivered")
	}

	// Notify event handler
	if s.eventHandler != nil {
		s.eventHandler.OnMessageDelivered(msg.ID)
	}

	// Send deliver_sm if session is a receiver or transceiver
	if session.BindType == "receiver" || session.BindType == "transceiver" {
		// Build delivery report
		now := time.Now().Format("060102150405") // YYMMDDhhmmss
		shortMsg := fmt.Sprintf("id:%s sub:001 dlvrd:001 submit date:%s done date:%s stat:DELIVRD err:000",
			msg.MessageID, now, now)

		params := &DeliverSMParams{
			SourceAddrTon:     0,
			SourceAddrNpi:     1,
			SourceAddr:        msg.DestAddr,
			DestAddrTon:       0,
			DestAddrNpi:       1,
			DestAddr:          msg.SourceAddr,
			ESMClass:          0x04, // Delivery report
			DataCoding:        0,
			SMLength:          byte(len(shortMsg)),
			ShortMessage:      shortMsg,
		}

		seqNum := session.NextSequenceNum()
		deliverPDU := EncodeDeliverSM(params, seqNum)

		if _, err := session.Conn.Write(deliverPDU); err != nil {
			log.Printf("Failed to send deliver_sm: %v", err)
		} else {
			log.Printf("Sent deliver_sm to %s for message %s", session.RemoteAddr, msg.MessageID)
		}
	}
}

// DisconnectSession disconnects a session by ID
func (s *Server) DisconnectSession(sessionID string) error {
	session := s.sessions.Get(sessionID)
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	session.Close()
	return nil
}

// GetSessions returns all active sessions
func (s *Server) GetSessions() []*SessionState {
	return s.sessions.GetAll()
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("%d%06d", time.Now().Unix(), time.Now().Nanosecond()/1000)
}

// decodeUCS2 decodes UCS2 (UTF-16BE) encoded bytes to string
func decodeUCS2(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Ensure even length for UTF-16
	if len(data)%2 != 0 {
		data = append(data, 0)
	}

	// Convert bytes to UTF-16 code points (big-endian)
	codePoints := make([]uint16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		codePoints[i/2] = uint16(data[i])<<8 | uint16(data[i+1])
	}

	// Convert UTF-16 to string
	runes := utf16.Decode(codePoints)
	return string(runes)
}
