// Package btcore provides a bitcoin core client which is able to connect and
// handshake with a node. It does nothing else.
package btcore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	magicNum        uint32 = 0xDAB5BFFA // testnet/regtest
	protocolVerison int32  = 70012      // Bitcoin Core 0.12.0 (Feb 2016)
	serviceMask     uint64 = 0x01       // NODE_NETWORK
)

type messageType int

const (
	Unknown messageType = iota
	PongMsg
	PingMsg
	VersionMsg
	VerackMsg
)

var (
	versionMessageBytes = [12]byte{'v', 'e', 'r', 's', 'i', 'o', 'n', 0x0, 0x0, 0x0, 0x0, 0x0}
	verackMsgBytes      = [12]byte{'v', 'e', 'r', 'a', 'c', 'k', 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	pongMsgBytes        = [12]byte{'p', 'o', 'n', 'g', 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	pingMsgBytes        = [12]byte{'p', 'i', 'n', 'g', 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	unknownMessageBytes = [12]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
)

func (mt *messageType) Bytes() [12]byte {
	switch *mt {
	case PingMsg:
		return pingMsgBytes
	case PongMsg:
		return pongMsgBytes
	case VersionMsg:
		return versionMessageBytes
	case VerackMsg:
		return verackMsgBytes
	default:
		return unknownMessageBytes
	}
}

// MessageTypeFromBytes takes a message type (command) as a [12]byte array and
// converts it to a MessageType.
func MessageTypeFromBytes(input [12]byte) messageType {
	if bytes.Equal(input[:], versionMessageBytes[:]) {
		return VersionMsg
	} else if bytes.Equal(input[:], verackMsgBytes[:]) {
		return VerackMsg
	} else {
		return Unknown
	}
}

// Message encapsulates a bitcoin coin message.
type Message struct {
	// Indicates the originating network.
	magicNumber uint32
	//  ASCII string indicating the payload message type (right padded with
	//  nulls [0x00]).
	command messageType
	// The length of the payload.
	payloadLen uint32
	// The first 4 bytes of the output to the function SHA256(SHA256(payload).
	checksum uint32
	payload  []byte
}

// Bytes serialises a Message into a byte slice for transmission over the wire.
func (m *Message) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, m.magicNumber)
	binary.Write(buf, binary.LittleEndian, m.command.Bytes())
	binary.Write(buf, binary.LittleEndian, m.payloadLen)

	binary.Write(buf, binary.LittleEndian, checksum(m.payload))

	binary.Write(buf, binary.LittleEndian, m.payload)

	return buf.Bytes()
}

// pretty is a simple helper which might be useful when debugging.
func (m *Message) pretty() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%.15s: %d\n", "magic", m.magicNumber))
	sb.WriteString(fmt.Sprintf("%.15s: %s\n", "command", m.command.Bytes()))
	sb.WriteString(fmt.Sprintf("%.15s: %d\n", "payload length", m.payloadLen))
	sb.WriteString(fmt.Sprintf("%.15s: %d\n", "checksum", m.checksum))
	sb.WriteString(fmt.Sprintf("%.15s: %s\n", "payload", m.payload))

	return sb.String()
}

// NewMessage creates a bitcoin message that can later be serialised.
// See spec: https://en.bitcoin.it/wiki/Protocol_documentation#Message_structure
func NewMessage(msgType messageType, payload []byte) *Message {
	return &Message{
		magicNumber: magicNum,
		command:     msgType,
		payloadLen:  uint32(len(payload)),
		payload:     payload,
		checksum:    checksum(payload),
	}
}

// messageFromBytes marshals a byte slice into a Message type. This method can
// return an ErrMessageSize error.
func messageFromBytes(data []byte) (*Message, error) {
	msg := &Message{
		payload: make([]byte, 0),
	}
	if len(data) < 24 {
		return nil, ErrMessageSize
	}
	msg.magicNumber = binary.LittleEndian.Uint32(data[:4])
	msg.command = MessageTypeFromBytes([12]byte(data[4:16]))
	msg.payloadLen = binary.LittleEndian.Uint32(data[16:20])
	msg.checksum = binary.LittleEndian.Uint32(data[20:24])
	if msg.payloadLen > 0 {
		msg.payload = data[24:]
	}

	return msg, nil
}

// newVersionMessage creates a new version message which is required to perform
// a handshake with the server.
// See spec: https://en.bitcoin.it/wiki/Protocol_documentation#version
func newVersionMessage(destinationIP net.IP, destinationPort uint16) *Message {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, protocolVerison)
	binary.Write(buf, binary.LittleEndian, serviceMask)
	binary.Write(buf, binary.LittleEndian, time.Now().UTC().Unix())

	// Network Address
	binary.Write(buf, binary.LittleEndian, serviceMask)
	binary.Write(buf, binary.BigEndian, destinationIP.To16())
	binary.Write(buf, binary.BigEndian, destinationPort)

	// addr_from (redundant).
	binary.Write(buf, binary.LittleEndian, []byte{
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00})

	binary.Write(buf, binary.LittleEndian, rand.Uint64()) // nonce.
	// user agent can be empty
	binary.Write(buf, binary.LittleEndian, []byte{0x00})
	binary.Write(buf, binary.LittleEndian, int32(0))
	binary.Write(buf, binary.LittleEndian, false)

	return NewMessage(VersionMsg, buf.Bytes())
}
