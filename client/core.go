package client

import (
	"dvr/log"
	"net"
	"time"

	"github.com/pkg/errors"
)

// Client struct for holding the client connection
type Client struct {
	Conn net.Conn
}

// NewClient creates and returns a new client using the provided bind address
func NewClient(address string) (*Client, error) {
	var c Client
	var err error

	// Create a new net Dialer and set the timeout to be 10 seconds
	// Timeout is max time allowed to wait for a dial to connect
	//
	// We're using a timeout so we don't completely break the program
	// if we never get a new connection
	timeout, _ := time.ParseDuration("5s")
	dialer := net.Dialer{Timeout: timeout}

	// Dial the connection adddress to establish the connection.
	c.Conn, err = dialer.Dial("udp", address)
	if err != nil {
		e := errors.Wrapf(err, "NewClient(%s): failed to create connection to address", address)
		return &c, e
	}

	return &c, nil
}

// SendPacket sends the provided packet to the client connection and then closes the connection
func (c *Client) SendPacket(packet []byte, log *log.Logger) {
	// Defer closing our client connection
	defer c.close()

	// Write the packet to the connection
	_, err := c.Conn.Write(packet)
	if err != nil {
		// Some error occurred ..
		// Create a new error message, print it to the user and return
		e := errors.Wrapf(err, "c.SendPacket: failed to write update packet %+v", packet)
		log.OutError("%+v\n", e)
	}
}

// close closes the UDP connection
func (c *Client) close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
