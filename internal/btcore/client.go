package btcore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
)

// Client encapsulates methods required for connecting and interacting with the
// bitcoin core daemon. The zero value is not ready to use, call the NewClient()
// function instead.
type Client struct {
	logger *slog.Logger
	addr   string
	port   uint16
	ctx    context.Context
}

// NewClient returns a new initialised btcore client.
func NewClient(address string, port uint16) *Client {
	logger := slog.Default().With("component", "btcore.Client")

	return &Client{
		addr:   address,
		port:   port,
		logger: logger,
	}
}

// Connect is blocking an establishes a connection to a node, performs the
// handshake and then listens for messsages. The method always returns a
// non-nil error.
func (c *Client) Connect(ctx context.Context) error {
	c.ctx = ctx

	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.addr, c.port))
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		c.logger.Error("err", err)
		return err
	}
	defer conn.Close()

	c.logger.Info("performing handshake")
	if err := c.handshake(addr, conn); err != nil {
		return err
	}
	c.logger.Info("handshake successful")
	msgChan := make(chan *Message)
	errChan := make(chan error)

	go c.readMessageAsync(conn, msgChan, errChan)

	for {
		select {
		case <-c.ctx.Done():
			return ErrContext
		case m := <-msgChan:
			switch m.command {
			case PingMsg:
				c.logger.Debug("ping message recieved")
				c.logger.Debug("sending pong message")
				// Nonce from ping included in pong.
				c.sendMessage(conn, NewMessage(PongMsg, m.payload))
			default:
				c.logger.Info("message recieved", "len", m.payloadLen, "type", m.command.Bytes())
			}
		case e := <-errChan:
			return errors.Join(ErrMessageReceive, e)
		}
	}
}

func (c *Client) Disconnect() {
	c.ctx.Done()
}

func (c *Client) sendMessage(conn net.Conn, msg *Message) error {
	_, err := conn.Write(msg.Bytes())
	if err != nil {
		return errors.Join(ErrMessageSend, err)
	}
	return nil
}

func (c *Client) readMessage(conn net.Conn) (*Message, error) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	msg, err := messageFromBytes(buffer[:n])
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Client) readMessageAsync(conn net.Conn, msgChan chan *Message, errChan chan error) {
	for {
		msg, err := c.readMessage(conn)
		if err != nil {
			errChan <- err
		} else {
			msgChan <- msg
		}
	}
}

func (c *Client) handshake(addr *net.TCPAddr, conn net.Conn) error {
	var (
		verackSeen  bool
		versionSeen bool
	)

	msg := newVersionMessage(addr.IP, uint16(addr.Port))
	_, err := conn.Write(msg.Bytes())
	if err != nil {
		return errors.Join(ErrHandshake, err)
	}

	for {
		msg, err = c.readMessage(conn)
		if err != nil {
			return errors.Join(ErrHandshake, err)
		}
		if msg.command == VersionMsg {
			versionSeen = true
			if err := c.sendMessage(conn, NewMessage(VerackMsg, nil)); err != nil {
				return errors.Join(ErrHandshake, err)
			}
		} else if msg.command == VerackMsg {
			verackSeen = true
		} else {
			return errors.Join(ErrHandshake, ErrUnexpectedMessageType)
		}

		if verackSeen && versionSeen {
			break
		}
	}

	return nil
}
