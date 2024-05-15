package btcore

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
)

func TestMessageTypeBytes(t *testing.T) {
	tests := []struct {
		name     string
		mt       messageType
		expected [12]byte
	}{
		{"PingMsg", PingMsg, pingMsgBytes},
		{"PongMsg", PongMsg, pongMsgBytes},
		{"VersionMsg", VersionMsg, versionMessageBytes},
		{"VerackMsg", VerackMsg, verackMsgBytes},
		{"UnknownMsg", Unknown, unknownMessageBytes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mt.Bytes()
			if !bytes.Equal(result[:], tt.expected[:]) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMessageTypeFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    [12]byte
		expected messageType
	}{
		{"VersionMsg", versionMessageBytes, VersionMsg},
		{"VerackMsg", verackMsgBytes, VerackMsg},
		{"UnknownMsg", unknownMessageBytes, Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MessageTypeFromBytes(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMessageBytes(t *testing.T) {
	payload := []byte("test payload")
	msg := NewMessage(VersionMsg, payload)
	msgBytes := msg.Bytes()

	if len(msgBytes) != 24+len(payload) {
		t.Errorf("expected length %d, got %d", 24+len(payload), len(msgBytes))
	}

	startString := binary.LittleEndian.Uint32(msgBytes[:4])
	if startString != magicNum {
		t.Errorf("expected startString %x, got %x", magicNum, startString)
	}

	command := MessageTypeFromBytes([12]byte(msgBytes[4:16]))
	if command != VersionMsg {
		t.Errorf("expected command %v, got %v", VersionMsg, command)
	}

	payloadLen := binary.LittleEndian.Uint32(msgBytes[16:20])
	if payloadLen != uint32(len(payload)) {
		t.Errorf("expected payloadLen %d, got %d", len(payload), payloadLen)
	}

	sum := binary.LittleEndian.Uint32(msgBytes[20:24])
	expectedChecksum := checksum(payload)
	if sum != expectedChecksum {
		t.Errorf("expected checksum %x, got %x", expectedChecksum, sum)
	}

	actualPayload := msgBytes[24:]
	if !bytes.Equal(actualPayload, payload) {
		t.Errorf("expected payload %v, got %v", payload, actualPayload)
	}
}

func TestMessageFromBytes(t *testing.T) {
	payload := []byte("test payload")
	msg := NewMessage(VersionMsg, payload)
	msgBytes := msg.Bytes()

	parsedMsg, err := messageFromBytes(msgBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedMsg.magicNumber != msg.magicNumber {
		t.Errorf("expected startString %x, got %x", msg.magicNumber, parsedMsg.magicNumber)
	}

	if parsedMsg.command != msg.command {
		t.Errorf("expected command %v, got %v", msg.command, parsedMsg.command)
	}

	if parsedMsg.payloadLen != msg.payloadLen {
		t.Errorf("expected payloadLen %d, got %d", msg.payloadLen, parsedMsg.payloadLen)
	}

	if parsedMsg.checksum != msg.checksum {
		t.Errorf("expected checksum %x, got %x", msg.checksum, parsedMsg.checksum)
	}

	if !bytes.Equal(parsedMsg.payload, msg.payload) {
		t.Errorf("expected payload %v, got %v", msg.payload, parsedMsg.payload)
	}
}

func TestNewVersionMessage(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")
	port := uint16(8333)
	msg := newVersionMessage(ip, port)

	if msg.command != VersionMsg {
		t.Errorf("expected command VersionMsg, got %v", msg.command)
	}

	if msg.magicNumber != magicNum {
		t.Errorf("expected startString %x, got %x", magicNum, msg.magicNumber)
	}

	expectedPayloadLen := 85 // Adjust based on actual payload size
	if int(msg.payloadLen) != expectedPayloadLen {
		t.Errorf("expected payloadLen %d, got %d", expectedPayloadLen, msg.payloadLen)
	}
}
