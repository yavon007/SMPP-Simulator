package smpp

import (
	"bytes"
	"testing"
)

func TestDecodePDU(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantErr     bool
		wantCmdID   uint32
		wantSeqNum  uint32
		wantStatus  uint32
	}{
		{
			name:    "data too short",
			data:    []byte{0x00, 0x01, 0x02},
			wantErr: true,
		},
		{
			name: "valid enquire_link",
			data: func() []byte {
				buf := new(bytes.Buffer)
				writeUint32(buf, 16)                  // CommandLength
				writeUint32(buf, CmdEnquireLink)      // CommandID
				writeUint32(buf, StatusOK)            // CommandStatus
				writeUint32(buf, 1)                   // SequenceNum
				return buf.Bytes()
			}(),
			wantErr:    false,
			wantCmdID:  CmdEnquireLink,
			wantSeqNum: 1,
			wantStatus: StatusOK,
		},
		{
			name: "valid bind_transmitter with body",
			data: func() []byte {
				buf := new(bytes.Buffer)
				writeUint32(buf, 30)                  // CommandLength
				writeUint32(buf, CmdBindTransmitter)  // CommandID
				writeUint32(buf, StatusOK)            // CommandStatus
				writeUint32(buf, 2)                   // SequenceNum
				buf.Write([]byte("test\x00pass\x00")) // Body
				return buf.Bytes()
			}(),
			wantErr:    false,
			wantCmdID:  CmdBindTransmitter,
			wantSeqNum: 2,
			wantStatus: StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdu, err := DecodePDU(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("DecodePDU() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("DecodePDU() unexpected error: %v", err)
				return
			}
			if pdu.CommandID != tt.wantCmdID {
				t.Errorf("CommandID = %v, want %v", pdu.CommandID, tt.wantCmdID)
			}
			if pdu.SequenceNum != tt.wantSeqNum {
				t.Errorf("SequenceNum = %v, want %v", pdu.SequenceNum, tt.wantSeqNum)
			}
			if pdu.CommandStatus != tt.wantStatus {
				t.Errorf("CommandStatus = %v, want %v", pdu.CommandStatus, tt.wantStatus)
			}
		})
	}
}

func TestEncodePDU(t *testing.T) {
	tests := []struct {
		name       string
		pdu        *PDU
		wantLen    uint32
	}{
		{
			name: "empty body",
			pdu: &PDU{
				CommandID:     CmdEnquireLink,
				CommandStatus: StatusOK,
				SequenceNum:   1,
				Body:          nil,
			},
			wantLen: 16,
		},
		{
			name: "with body",
			pdu: &PDU{
				CommandID:     CmdBindTransmitter,
				CommandStatus: StatusOK,
				SequenceNum:   2,
				Body:          []byte("test\x00pass\x00"),
			},
			wantLen: 26,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := EncodePDU(tt.pdu)
			if uint32(len(data)) != tt.wantLen {
				t.Errorf("encoded length = %v, want %v", len(data), tt.wantLen)
			}

			// Verify we can decode it back
			decoded, err := DecodePDU(data)
			if err != nil {
				t.Errorf("failed to decode encoded PDU: %v", err)
				return
			}
			if decoded.CommandID != tt.pdu.CommandID {
				t.Errorf("decoded CommandID = %v, want %v", decoded.CommandID, tt.pdu.CommandID)
			}
			if decoded.SequenceNum != tt.pdu.SequenceNum {
				t.Errorf("decoded SequenceNum = %v, want %v", decoded.SequenceNum, tt.pdu.SequenceNum)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := &PDU{
		CommandID:     CmdSubmitSM,
		CommandStatus: StatusOK,
		SequenceNum:   12345,
		Body:          []byte("test body data with null\x00bytes"),
	}

	encoded := EncodePDU(original)
	decoded, err := DecodePDU(encoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded.CommandID != original.CommandID {
		t.Errorf("CommandID mismatch")
	}
	if decoded.CommandStatus != original.CommandStatus {
		t.Errorf("CommandStatus mismatch")
	}
	if decoded.SequenceNum != original.SequenceNum {
		t.Errorf("SequenceNum mismatch")
	}
	if !bytes.Equal(decoded.Body, original.Body) {
		t.Errorf("Body mismatch")
	}
}

// Helper function to write uint32 in big-endian
func writeUint32(buf *bytes.Buffer, val uint32) {
	buf.WriteByte(byte(val >> 24))
	buf.WriteByte(byte(val >> 16))
	buf.WriteByte(byte(val >> 8))
	buf.WriteByte(byte(val))
}
