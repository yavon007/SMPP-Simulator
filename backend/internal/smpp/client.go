package smpp

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// responseWaiter represents a pending response wait
type responseWaiter struct {
	ch     chan *PDU
	expire time.Time
}

// OutboundSession represents an active outbound SMPP connection
type OutboundSession struct {
	ID           string
	SystemID     string
	Password     string
	BindType     string // transmitter, receiver, transceiver
	RemoteAddr   string
	Conn         net.Conn
	ConnectedAt  time.Time
	Status       string // connecting, active, closed, error
	ErrorMessage string
	SequenceNum  uint32
	reader       *bufio.Reader
	waiters      map[uint32]*responseWaiter
	waitersMu    sync.Mutex
	mu           sync.Mutex
}

// NewOutboundSession creates a new outbound session
func NewOutboundSession(systemID, password, bindType string) *OutboundSession {
	return &OutboundSession{
		ID:          generateID(),
		SystemID:    systemID,
		Password:    password,
		BindType:    bindType,
		ConnectedAt: time.Now(),
		Status:      "connecting",
		SequenceNum: 0,
		waiters:     make(map[uint32]*responseWaiter),
	}
}

// NextSequenceNum returns the next sequence number
func (s *OutboundSession) NextSequenceNum() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SequenceNum++
	return s.SequenceNum
}

// Close closes the session
func (s *OutboundSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Conn != nil {
		s.Conn.Close()
	}
	s.Status = "closed"
	// Close all waiters
	s.waitersMu.Lock()
	for _, w := range s.waiters {
		close(w.ch)
	}
	s.waiters = make(map[uint32]*responseWaiter)
	s.waitersMu.Unlock()
}

// SetStatus sets the session status
func (s *OutboundSession) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// SetError sets the error message and status
func (s *OutboundSession) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "error"
	if err != nil {
		s.ErrorMessage = err.Error()
	}
}

// registerWaiter registers a channel to wait for a response
func (s *OutboundSession) registerWaiter(seqNum uint32, timeout time.Duration) chan *PDU {
	ch := make(chan *PDU, 1)
	s.waitersMu.Lock()
	s.waiters[seqNum] = &responseWaiter{
		ch:     ch,
		expire: time.Now().Add(timeout),
	}
	s.waitersMu.Unlock()
	return ch
}

// unregisterWaiter removes a waiter
func (s *OutboundSession) unregisterWaiter(seqNum uint32) {
	s.waitersMu.Lock()
	if w, ok := s.waiters[seqNum]; ok {
		close(w.ch)
		delete(s.waiters, seqNum)
	}
	s.waitersMu.Unlock()
}

// deliverResponse delivers a response to a waiting goroutine
func (s *OutboundSession) deliverResponse(pdu *PDU) {
	s.waitersMu.Lock()
	if w, ok := s.waiters[pdu.SequenceNum]; ok {
		select {
		case w.ch <- pdu:
		default:
		}
		delete(s.waiters, pdu.SequenceNum)
	}
	s.waitersMu.Unlock()
}

// Client represents an SMPP client that can initiate connections
type Client struct {
	sessions map[string]*OutboundSession
	mu       sync.RWMutex
}

// NewClient creates a new SMPP client
func NewClient() *Client {
	return &Client{
		sessions: make(map[string]*OutboundSession),
	}
}

// ConnectParams represents parameters for connecting to an SMSC
type ConnectParams struct {
	Host     string
	Port     string
	SystemID string
	Password string
	BindType string // transmitter, receiver, transceiver
}

// Connect connects to a remote SMSC
func (c *Client) Connect(params *ConnectParams) (*OutboundSession, error) {
	if params.Host == "" || params.Port == "" {
		return nil, errors.New("host and port are required")
	}

	session := NewOutboundSession(params.SystemID, params.Password, params.BindType)

	addr := fmt.Sprintf("%s:%s", params.Host, params.Port)
	session.RemoteAddr = addr

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		session.SetError(err)
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	session.Conn = conn
	session.reader = bufio.NewReader(conn)

	// Add session to manager
	c.mu.Lock()
	c.sessions[session.ID] = session
	c.mu.Unlock()

	// Prepare bind request
	seqNum := session.NextSequenceNum()
	var bindPDU []byte
	switch params.BindType {
	case "transmitter":
		bindPDU = EncodeBindTransmitter(params.SystemID, params.Password, seqNum)
	case "receiver":
		bindPDU = EncodeBindReceiver(params.SystemID, params.Password, seqNum)
	case "transceiver":
		bindPDU = EncodeBindTransceiver(params.SystemID, params.Password, seqNum)
	default:
		session.Close()
		c.Remove(session.ID)
		return nil, fmt.Errorf("invalid bind type: %s", params.BindType)
	}

	// Register waiter BEFORE starting response handler and sending request
	respCh := session.registerWaiter(seqNum, 10*time.Second)

	// Start response handler
	go c.handleResponses(session)

	// Send bind request
	if _, err := conn.Write(bindPDU); err != nil {
		session.unregisterWaiter(seqNum)
		session.SetError(err)
		c.Remove(session.ID)
		return nil, fmt.Errorf("failed to send bind request: %w", err)
	}

	select {
	case respPDU := <-respCh:
		if respPDU == nil {
			session.Close()
			c.Remove(session.ID)
			return nil, errors.New("connection closed")
		}
		if !isBindResp(respPDU.CommandID) {
			session.Close()
			c.Remove(session.ID)
			return nil, fmt.Errorf("unexpected response: %s", CommandName(respPDU.CommandID))
		}
		if respPDU.CommandStatus != StatusOK {
			session.Close()
			c.Remove(session.ID)
			return nil, fmt.Errorf("bind failed: %s", statusText(respPDU.CommandStatus))
		}
	case <-time.After(10 * time.Second):
		session.unregisterWaiter(seqNum)
		session.Close()
		c.Remove(session.ID)
		return nil, errors.New("bind timeout")
	}

	session.SetStatus("active")
	log.Printf("Outbound session %s connected to %s (bind_type: %s)", session.ID, addr, params.BindType)

	return session, nil
}

// handleResponses handles incoming PDUs for a session
func (c *Client) handleResponses(session *OutboundSession) {
	defer func() {
		session.Close()
		c.Remove(session.ID)
		log.Printf("Outbound session %s disconnected", session.ID)
	}()

	for {
		session.mu.Lock()
		conn := session.Conn
		reader := session.reader
		status := session.Status
		session.mu.Unlock()

		if conn == nil || status == "closed" || status == "error" {
			return
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		pdu, err := readPDU(reader)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && session.Status == "active" {
				log.Printf("Outbound session %s read error: %v", session.ID, err)
				session.SetError(err)
			}
			return
		}

		log.Printf("Outbound session %s received PDU: %s (seq: %d, status: %d)",
			session.ID, CommandName(pdu.CommandID), pdu.SequenceNum, pdu.CommandStatus)

		// Handle PDU based on command
		switch pdu.CommandID {
		case CmdDeliverSM:
			// Send deliver_sm_resp
			resp := EncodeDeliverSMResp(pdu.SequenceNum, StatusOK)
			session.mu.Lock()
			if session.Conn != nil {
				session.Conn.Write(resp)
			}
			session.mu.Unlock()
			log.Printf("Sent deliver_sm_resp for seq %d", pdu.SequenceNum)

		case CmdEnquireLink:
			// Send enquire_link_resp
			resp := EncodeEnquireLinkResp(pdu.SequenceNum)
			session.mu.Lock()
			if session.Conn != nil {
				session.Conn.Write(resp)
			}
			session.mu.Unlock()

		case CmdUnbind:
			// Send unbind_resp and close
			resp := EncodeUnbindResp(pdu.SequenceNum)
			session.mu.Lock()
			if session.Conn != nil {
				session.Conn.Write(resp)
			}
			session.mu.Unlock()
			return

		default:
			// Deliver response to waiter (submit_sm_resp, bind_resp, etc.)
			session.deliverResponse(pdu)
		}
	}
}

// Remove removes a session
func (c *Client) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, id)
}

// Get gets a session by ID
func (c *Client) Get(id string) *OutboundSession {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessions[id]
}

// GetAll returns all sessions
func (c *Client) GetAll() []*OutboundSession {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*OutboundSession, 0, len(c.sessions))
	for _, s := range c.sessions {
		result = append(result, s)
	}
	return result
}

// Disconnect disconnects a session by ID
func (c *Client) Disconnect(id string) error {
	session := c.Get(id)
	if session == nil {
		return fmt.Errorf("session not found: %s", id)
	}
	session.Close()
	c.Remove(id)
	return nil
}

// SendMessageParams represents parameters for sending a message
type OutboundSendMessageParams struct {
	SessionID  string
	SourceAddr string
	DestAddr   string
	Content    string
	DataCoding byte // 0=GSM7, 8=UCS2
}

// SendMessage sends a submit_sm to a specific session
func (c *Client) SendMessage(params *OutboundSendMessageParams) (string, error) {
	session := c.Get(params.SessionID)
	if session == nil {
		return "", fmt.Errorf("session not found: %s", params.SessionID)
	}

	session.mu.Lock()
	conn := session.Conn
	status := session.Status
	session.mu.Unlock()

	if conn == nil || status != "active" {
		return "", fmt.Errorf("session is not active")
	}

	// Encode message content
	var shortMessage []byte
	if params.DataCoding == 8 {
		// UCS2 encoding
		shortMessage = encodeUCS2(params.Content)
	} else {
		// GSM7/ASCII
		shortMessage = []byte(params.Content)
	}

	// Build submit_sm
	submitParams := &SubmitSMParams{
		ServiceType:        "",
		SourceAddrTon:      0,
		SourceAddrNpi:      1,
		SourceAddr:         params.SourceAddr,
		DestAddrTon:        0,
		DestAddrNpi:        1,
		DestAddr:           params.DestAddr,
		ESMClass:           0,
		ProtocolID:         0,
		PriorityFlag:       0,
		ScheduleDelTime:    "",
		ValidityPeriod:     "",
		RegisteredDelivery: 1, // Request delivery report
		ReplaceIfPresent:   0,
		DataCoding:         params.DataCoding,
		SMDefaultMsgID:     0,
		SMLength:           byte(len(shortMessage)),
		ShortMessage:       string(shortMessage),
	}

	seqNum := session.NextSequenceNum()
	submitPDU := EncodeSubmitSM(submitParams, seqNum)

	// Register waiter before sending
	respCh := session.registerWaiter(seqNum, 30*time.Second)

	session.mu.Lock()
	if session.Conn != nil {
		_, err := session.Conn.Write(submitPDU)
		session.mu.Unlock()
		if err != nil {
			session.unregisterWaiter(seqNum)
			session.SetError(err)
			return "", fmt.Errorf("failed to send submit_sm: %w", err)
		}
	} else {
		session.mu.Unlock()
		session.unregisterWaiter(seqNum)
		return "", fmt.Errorf("connection closed")
	}

	log.Printf("Sent submit_sm to session %s: from=%s, to=%s", params.SessionID, params.SourceAddr, params.DestAddr)

	// Wait for response
	select {
	case respPDU := <-respCh:
		if respPDU == nil {
			return "", errors.New("connection closed")
		}
		if respPDU.CommandID != CmdSubmitSMResp {
			return "", fmt.Errorf("unexpected response: %s", CommandName(respPDU.CommandID))
		}
		if respPDU.CommandStatus != StatusOK {
			return "", fmt.Errorf("submit_sm failed: %s", statusText(respPDU.CommandStatus))
		}
		// Extract message ID from response
		messageID, _ := decodeCString(respPDU.Body, 0)
		log.Printf("submit_sm_resp received: message_id=%s", messageID)
		return messageID, nil

	case <-time.After(30 * time.Second):
		session.unregisterWaiter(seqNum)
		return "", errors.New("submit_sm timeout")
	}
}

// EncodeBindTransmitter encodes bind_transmitter PDU
func EncodeBindTransmitter(systemID, password string, sequenceNum uint32) []byte {
	body := new(bufBuffer)
	body.writeCString(systemID)
	body.writeCString(password)
	body.writeCString("") // system_type
	body.WriteByte(0x34)  // interface_version
	body.WriteByte(0)     // addr_ton
	body.WriteByte(0)     // addr_npi
	body.writeCString("") // address_range

	pdu := &PDU{
		CommandID:     CmdBindTransmitter,
		CommandStatus: 0,
		SequenceNum:   sequenceNum,
		Body:          body.Bytes(),
	}
	return EncodePDU(pdu)
}

// EncodeBindReceiver encodes bind_receiver PDU
func EncodeBindReceiver(systemID, password string, sequenceNum uint32) []byte {
	body := new(bufBuffer)
	body.writeCString(systemID)
	body.writeCString(password)
	body.writeCString("") // system_type
	body.WriteByte(0x34)  // interface_version
	body.WriteByte(0)     // addr_ton
	body.WriteByte(0)     // addr_npi
	body.writeCString("") // address_range

	pdu := &PDU{
		CommandID:     CmdBindReceiver,
		CommandStatus: 0,
		SequenceNum:   sequenceNum,
		Body:          body.Bytes(),
	}
	return EncodePDU(pdu)
}

// EncodeBindTransceiver encodes bind_transceiver PDU
func EncodeBindTransceiver(systemID, password string, sequenceNum uint32) []byte {
	body := new(bufBuffer)
	body.writeCString(systemID)
	body.writeCString(password)
	body.writeCString("") // system_type
	body.WriteByte(0x34)  // interface_version
	body.WriteByte(0)     // addr_ton
	body.WriteByte(0)     // addr_npi
	body.writeCString("") // address_range

	pdu := &PDU{
		CommandID:     CmdBindTransceiver,
		CommandStatus: 0,
		SequenceNum:   sequenceNum,
		Body:          body.Bytes(),
	}
	return EncodePDU(pdu)
}

// EncodeSubmitSM encodes submit_sm PDU
func EncodeSubmitSM(params *SubmitSMParams, sequenceNum uint32) []byte {
	body := new(bufBuffer)
	body.writeCString(params.ServiceType)
	body.WriteByte(params.SourceAddrTon)
	body.WriteByte(params.SourceAddrNpi)
	body.writeCString(params.SourceAddr)
	body.WriteByte(params.DestAddrTon)
	body.WriteByte(params.DestAddrNpi)
	body.writeCString(params.DestAddr)
	body.WriteByte(params.ESMClass)
	body.WriteByte(params.ProtocolID)
	body.WriteByte(params.PriorityFlag)
	body.writeCString(params.ScheduleDelTime)
	body.writeCString(params.ValidityPeriod)
	body.WriteByte(params.RegisteredDelivery)
	body.WriteByte(params.ReplaceIfPresent)
	body.WriteByte(params.DataCoding)
	body.WriteByte(params.SMDefaultMsgID)
	body.WriteByte(params.SMLength)
	body.Write([]byte(params.ShortMessage))

	pdu := &PDU{
		CommandID:     CmdSubmitSM,
		CommandStatus: 0,
		SequenceNum:   sequenceNum,
		Body:          body.Bytes(),
	}
	return EncodePDU(pdu)
}

// EncodeDeliverSMResp encodes deliver_sm_resp PDU
func EncodeDeliverSMResp(sequenceNum uint32, status uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdDeliverSMResp,
		CommandStatus: status,
		SequenceNum:   sequenceNum,
		Body:          encodeCString(""), // message_id
	}
	return EncodePDU(pdu)
}

// readPDU reads a PDU from a reader
func readPDU(reader *bufio.Reader) (*PDU, error) {
	// Read PDU length (first 4 bytes)
	lengthBuf := make([]byte, 4)
	_, err := reader.Read(lengthBuf)
	if err != nil {
		return nil, err
	}

	// Decode length
	length := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])
	if length < 16 || length > 10240 {
		return nil, fmt.Errorf("invalid PDU length: %d", length)
	}

	// Read remaining PDU
	remaining := length - 4
	pduBuf := make([]byte, remaining)
	_, err = reader.Read(pduBuf)
	if err != nil {
		return nil, err
	}

	// Combine and decode PDU
	fullPDU := append(lengthBuf, pduBuf...)
	return DecodePDU(fullPDU)
}

// isBindResp checks if command ID is a bind response
func isBindResp(cmdID uint32) bool {
	return cmdID == CmdBindTransmitterResp ||
		cmdID == CmdBindReceiverResp ||
		cmdID == CmdBindTransceiverResp
}

// statusText returns human readable status text
func statusText(status uint32) string {
	switch status {
	case StatusOK:
		return "OK"
	case StatusInvSysID:
		return "Invalid System ID or Password"
	case StatusBindFailed:
		return "Bind Failed"
	case StatusInvMsgID:
		return "Invalid Message ID"
	case StatusInvDestAddr:
		return "Invalid Destination Address"
	case StatusInvSrcAddr:
		return "Invalid Source Address"
	case StatusInvMsgLen:
		return "Invalid Message Length"
	case StatusESMERROUT:
		return "ESME Generic Error"
	default:
		return fmt.Sprintf("Error(0x%08X)", status)
	}
}

// bufBuffer is a helper for building PDU bodies
type bufBuffer struct {
	data []byte
}

func (b *bufBuffer) Write(p []byte) {
	b.data = append(b.data, p...)
}

func (b *bufBuffer) WriteByte(c byte) {
	b.data = append(b.data, c)
}

func (b *bufBuffer) writeCString(s string) {
	b.data = append(b.data, []byte(s)...)
	b.data = append(b.data, 0)
}

func (b *bufBuffer) Bytes() []byte {
	return b.data
}

// generateOutboundID generates a unique ID
func generateOutboundID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d%06d", time.Now().Unix(), time.Now().Nanosecond()/1000)
	}
	return fmt.Sprintf("out_%d%s", time.Now().Unix(), hex.EncodeToString(b))
}
