package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type packetType int32

const (
	packetTypeCommand packetType = 2
	packetTypeAuth    packetType = 3
	badLoginID        int32      = -1
	maxPacketSize     int        = 1460
)

// Client is our representation of an RCON client
type Client struct {
	host     string
	port     int
	password string
	authed   bool
	conn     net.Conn
}

type header struct {
	Size       int32
	RequestID  int32
	PacketType packetType
}

// Dial connects to the given host
func Dial(host string, port int, password string) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	// Establish a connection
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	// Create a Client object
	c := Client{
		host:     host,
		port:     port,
		password: password,
		authed:   false,
		conn:     conn,
	}
	return &c, nil
}

// Authenticate attempts to authenticate to RCON.
func (c *Client) Authenticate() error {
	// Make sure we're not already authenticated
	if c.authed {
		return errors.New("already authenticated")
	}
	// Send our auth packet
	head, resp, err := c.sendPacket(packetTypeAuth, []byte(c.password))
	if err != nil {
		return err
	}
	// Read the response to see if we authenticated
	if head.RequestID == badLoginID {
		return fmt.Errorf("unable to authenticate: %s", string(resp))
	}
	c.authed = true
	return nil
}

// Close closes the connection to the RCON server.
func (c *Client) Close() error {
	return c.conn.Close()
}

// SendCommand sends a command to the RCON server and returns any result.
func (c *Client) SendCommand(cmd string) (string, error) {
	// Make sure we're authenticated to RCON
	if !c.authed {
		return "", errors.New("cannot send command when not authenticated")
	}
	// Send the packet to the RCON server
	head, payload, err := c.sendPacket(packetTypeCommand, []byte(cmd))
	if err != nil {
		return "", err
	}
	// Our authentication is bad
	if head.RequestID == badLoginID {
		return "", errors.New("unable to send command: bad auth")
	}
	// Return the response
	return string(payload), nil
}

// sendPacket sends a packet over the connection and returns the response
func (c *Client) sendPacket(t packetType, payload []byte) (header, []byte, error) {
	// Generate a binary packet
	packet, err := createPacket(t, payload)
	if err != nil {
		return header{}, nil, err
	}
	// Send the packet over the connection
	if _, err := c.conn.Write(packet); err != nil {
		return header{}, nil, err
	}
	// Read the response
	head, resp, err := readPacket(c.conn)
	if err != nil {
		return header{}, nil, err
	}
	// Return the response
	return head, resp, nil
}

// createPacket encodes a packet to send over the connection.
func createPacket(t packetType, payload []byte) ([]byte, error) {
	pad := [2]byte{}
	length := int32(len(payload) + 10)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, length)
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, t)
	binary.Write(&buf, binary.LittleEndian, payload)
	binary.Write(&buf, binary.LittleEndian, pad)
	// Make sure we're under the size limit
	if buf.Len() >= maxPacketSize {
		return nil, errors.New("packet size too lage")
	}
	// Return the bytes
	return buf.Bytes(), nil
}

// readPacket decodes a binary packet into Go objects.
func readPacket(reader io.Reader) (header, []byte, error) {
	head := header{}
	// Read the packet header
	if err := binary.Read(reader, binary.LittleEndian, &head); err != nil {
		return header{}, nil, err
	}
	// Read the response body
	resp := make([]byte, head.Size-8)
	if _, err := io.ReadFull(reader, resp); err != nil {
		return header{}, nil, err
	}
	// Return the data
	return head, resp[:len(resp)-2], nil
}
