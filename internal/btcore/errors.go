package btcore

import "errors"

var (
	ErrChecksum              = errors.New("wrong checksum")
	ErrMessageSend           = errors.New("problem sending message")
	ErrMessageReceive        = errors.New("problem recieving message")
	ErrUnexpectedMessageType = errors.New("unexpected message type")
	ErrHandshake             = errors.New("handshake error")
	ErrMessageSize           = errors.New("incorrect message size")
	ErrContext               = errors.New("Context closed")
)
