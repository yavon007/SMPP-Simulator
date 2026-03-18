package smpp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// SMPP PDU Command IDs
const (
	CmdBindTransmitter   uint32 = 0x00000002
	CmdBindTransmitterResp uint32 = 0x80000002
	CmdBindReceiver      uint32 = 0x00000001
	CmdBindReceiverResp  uint32 = 0x80000001
	CmdBindTransceiver   uint32 = 0x00000009
	CmdBindTransceiverResp uint32 = 0x80000009
	CmdUnbind            uint32 = 0x00000006
	CmdUnbindResp        uint32 = 0x80000006
	CmdSubmitSM          uint32 = 0x00000004
	CmdSubmitSMResp      uint32 = 0x80000004
	CmdDeliverSM         uint32 = 0x00000005
	CmdDeliverSMResp     uint32 = 0x80000005
	CmdQuerySM           uint32 = 0x00000003
	CmdQuerySMResp       uint32 = 0x80000003
	CmdEnquireLink       uint32 = 0x00000015
	CmdEnquireLinkResp   uint32 = 0x80000015
)

// SMPP Command Status
const (
	StatusOK              uint32 = 0x00000000
	StatusInvSysID        uint32 = 0x0000000E
	StatusInvPassWD       uint32 = 0x0000000E
	StatusBindFailed      uint32 = 0x0000000D
	StatusInvMsgID        uint32 = 0x0000000C
	StatusInvDestAddr     uint32 = 0x0000000B
	StatusInvSrcAddr      uint32 = 0x0000000A
	StatusInvMsgLen       uint32 = 0x00000009
	StatusInvESMClass     uint32 = 0x00000043
	StatusGenericError    uint32 = 0x000000FF
)

// PDU represents a Protocol Data Unit
type PDU struct {
	CommandLength uint32
	CommandID     uint32
	CommandStatus uint32
	SequenceNum   uint32
	Body          []byte
}

// BindParams represents bind request parameters
type BindParams struct {
	SystemID       string
	Password       string
	SystemType     string
	InterfaceVer   byte
	AddrTon        byte
	AddrNpi        byte
	AddressRange   string
}

// SubmitSMParams represents submit_sm parameters
type SubmitSMParams struct {
	ServiceType       string
	SourceAddrTon     byte
	SourceAddrNpi     byte
	SourceAddr        string
	DestAddrTon       byte
	DestAddrNpi       byte
	DestAddr          string
	ESMClass          byte
	ProtocolID        byte
	PriorityFlag      byte
	ScheduleDelTime   string
	ValidityPeriod    string
	RegisteredDelivery byte
	ReplaceIfPresent  byte
	DataCoding        byte
	SMDefaultMsgID    byte
	SMLength          byte
	ShortMessage      string
}

// DeliverSMParams represents deliver_sm parameters
type DeliverSMParams struct {
	ServiceType       string
	SourceAddrTon     byte
	SourceAddrNpi     byte
	SourceAddr        string
	DestAddrTon       byte
	DestAddrNpi       byte
	DestAddr          string
	ESMClass          byte
	ProtocolID        byte
	PriorityFlag      byte
	ScheduleDelTime   string
	ValidityPeriod    string
	RegisteredDelivery byte
	ReplaceIfPresent  byte
	DataCoding        byte
	SMDefaultMsgID    byte
	SMLength          byte
	ShortMessage      string
}

// DecodePDU decodes a PDU from bytes
func DecodePDU(data []byte) (*PDU, error) {
	if len(data) < 16 {
		return nil, errors.New("data too short for PDU header")
	}

	pdu := &PDU{}
	buf := bytes.NewReader(data)

	if err := binary.Read(buf, binary.BigEndian, &pdu.CommandLength); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &pdu.CommandID); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &pdu.CommandStatus); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &pdu.SequenceNum); err != nil {
		return nil, err
	}

	if pdu.CommandLength > 16 {
		pdu.Body = data[16:pdu.CommandLength]
	}

	return pdu, nil
}

// EncodePDU encodes a PDU to bytes
func EncodePDU(pdu *PDU) []byte {
	buf := new(bytes.Buffer)
	pdu.CommandLength = 16 + uint32(len(pdu.Body))

	binary.Write(buf, binary.BigEndian, pdu.CommandLength)
	binary.Write(buf, binary.BigEndian, pdu.CommandID)
	binary.Write(buf, binary.BigEndian, pdu.CommandStatus)
	binary.Write(buf, binary.BigEndian, pdu.SequenceNum)
	buf.Write(pdu.Body)

	return buf.Bytes()
}

// decodeCString decodes a null-terminated string
func decodeCString(data []byte, offset int) (string, int) {
	end := offset
	for end < len(data) && data[end] != 0 {
		end++
	}
	return string(data[offset:end]), end + 1
}

// encodeCString encodes a null-terminated string
func encodeCString(s string) []byte {
	return append([]byte(s), 0)
}

// DecodeBind decodes bind request body
func DecodeBind(body []byte) *BindParams {
	params := &BindParams{}
	offset := 0

	params.SystemID, offset = decodeCString(body, offset)
	params.Password, offset = decodeCString(body, offset)
	params.SystemType, offset = decodeCString(body, offset)

	if offset < len(body) {
		params.InterfaceVer = body[offset]
		offset++
	}
	if offset < len(body) {
		params.AddrTon = body[offset]
		offset++
	}
	if offset < len(body) {
		params.AddrNpi = body[offset]
		offset++
	}
	if offset < len(body) {
		params.AddressRange, _ = decodeCString(body, offset)
	}

	return params
}

// EncodeBindTransmitterResp encodes bind_transmitter response
func EncodeBindTransmitterResp(systemID string, sequenceNum uint32, status uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdBindTransmitterResp,
		CommandStatus: status,
		SequenceNum:   sequenceNum,
		Body:          encodeCString(systemID),
	}
	return EncodePDU(pdu)
}

// EncodeBindReceiverResp encodes bind_receiver response
func EncodeBindReceiverResp(systemID string, sequenceNum uint32, status uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdBindReceiverResp,
		CommandStatus: status,
		SequenceNum:   sequenceNum,
		Body:          encodeCString(systemID),
	}
	return EncodePDU(pdu)
}

// EncodeBindTransceiverResp encodes bind_transceiver response
func EncodeBindTransceiverResp(systemID string, sequenceNum uint32, status uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdBindTransceiverResp,
		CommandStatus: status,
		SequenceNum:   sequenceNum,
		Body:          encodeCString(systemID),
	}
	return EncodePDU(pdu)
}

// DecodeSubmitSM decodes submit_sm body
func DecodeSubmitSM(body []byte) *SubmitSMParams {
	params := &SubmitSMParams{}
	offset := 0

	params.ServiceType, offset = decodeCString(body, offset)
	if offset < len(body) {
		params.SourceAddrTon = body[offset]
		offset++
	}
	if offset < len(body) {
		params.SourceAddrNpi = body[offset]
		offset++
	}
	params.SourceAddr, offset = decodeCString(body, offset)

	if offset < len(body) {
		params.DestAddrTon = body[offset]
		offset++
	}
	if offset < len(body) {
		params.DestAddrNpi = body[offset]
		offset++
	}
	params.DestAddr, offset = decodeCString(body, offset)

	if offset < len(body) {
		params.ESMClass = body[offset]
		offset++
	}
	if offset < len(body) {
		params.ProtocolID = body[offset]
		offset++
	}
	if offset < len(body) {
		params.PriorityFlag = body[offset]
		offset++
	}
	params.ScheduleDelTime, offset = decodeCString(body, offset)
	params.ValidityPeriod, offset = decodeCString(body, offset)

	if offset < len(body) {
		params.RegisteredDelivery = body[offset]
		offset++
	}
	if offset < len(body) {
		params.ReplaceIfPresent = body[offset]
		offset++
	}
	if offset < len(body) {
		params.DataCoding = body[offset]
		offset++
	}
	if offset < len(body) {
		params.SMDefaultMsgID = body[offset]
		offset++
	}
	if offset < len(body) {
		params.SMLength = body[offset]
		offset++
	}
	if offset < len(body) {
		params.ShortMessage = string(body[offset : offset+int(params.SMLength)])
	}

	return params
}

// EncodeSubmitSMResp encodes submit_sm response
func EncodeSubmitSMResp(messageID string, sequenceNum uint32, status uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdSubmitSMResp,
		CommandStatus: status,
		SequenceNum:   sequenceNum,
		Body:          encodeCString(messageID),
	}
	return EncodePDU(pdu)
}

// EncodeDeliverSM encodes deliver_sm PDU
func EncodeDeliverSM(params *DeliverSMParams, sequenceNum uint32) []byte {
	body := new(bytes.Buffer)

	body.Write(encodeCString(params.ServiceType))
	body.WriteByte(params.SourceAddrTon)
	body.WriteByte(params.SourceAddrNpi)
	body.Write(encodeCString(params.SourceAddr))
	body.WriteByte(params.DestAddrTon)
	body.WriteByte(params.DestAddrNpi)
	body.Write(encodeCString(params.DestAddr))
	body.WriteByte(params.ESMClass)
	body.WriteByte(params.ProtocolID)
	body.WriteByte(params.PriorityFlag)
	body.Write(encodeCString(params.ScheduleDelTime))
	body.Write(encodeCString(params.ValidityPeriod))
	body.WriteByte(params.RegisteredDelivery)
	body.WriteByte(params.ReplaceIfPresent)
	body.WriteByte(params.DataCoding)
	body.WriteByte(params.SMDefaultMsgID)
	body.WriteByte(params.SMLength)
	body.Write([]byte(params.ShortMessage))

	pdu := &PDU{
		CommandID:     CmdDeliverSM,
		CommandStatus: 0,
		SequenceNum:   sequenceNum,
		Body:          body.Bytes(),
	}
	return EncodePDU(pdu)
}

// EncodeUnbindResp encodes unbind response
func EncodeUnbindResp(sequenceNum uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdUnbindResp,
		CommandStatus: StatusOK,
		SequenceNum:   sequenceNum,
		Body:          nil,
	}
	return EncodePDU(pdu)
}

// EncodeEnquireLinkResp encodes enquire_link response
func EncodeEnquireLinkResp(sequenceNum uint32) []byte {
	pdu := &PDU{
		CommandID:     CmdEnquireLinkResp,
		CommandStatus: StatusOK,
		SequenceNum:   sequenceNum,
		Body:          nil,
	}
	return EncodePDU(pdu)
}

// CommandName returns human readable command name
func CommandName(cmdID uint32) string {
	switch cmdID {
	case CmdBindTransmitter:
		return "bind_transmitter"
	case CmdBindTransmitterResp:
		return "bind_transmitter_resp"
	case CmdBindReceiver:
		return "bind_receiver"
	case CmdBindReceiverResp:
		return "bind_receiver_resp"
	case CmdBindTransceiver:
		return "bind_transceiver"
	case CmdBindTransceiverResp:
		return "bind_transceiver_resp"
	case CmdUnbind:
		return "unbind"
	case CmdUnbindResp:
		return "unbind_resp"
	case CmdSubmitSM:
		return "submit_sm"
	case CmdSubmitSMResp:
		return "submit_sm_resp"
	case CmdDeliverSM:
		return "deliver_sm"
	case CmdDeliverSMResp:
		return "deliver_sm_resp"
	case CmdQuerySM:
		return "query_sm"
	case CmdQuerySMResp:
		return "query_sm_resp"
	case CmdEnquireLink:
		return "enquire_link"
	case CmdEnquireLinkResp:
		return "enquire_link_resp"
	default:
		return fmt.Sprintf("unknown(0x%08X)", cmdID)
	}
}
