package client

import (
    "dvr-protocol/types"
    "fmt"
    "github.com/pkg/errors"
    "net"
    "time"
)

// type Client struct {{{

type Client struct {
    Conn net.Conn
} // }}}

// func NewClient {{{
//
// Creates and returns a new client using the provided bind address
func NewClient(address string, cost int) (*Client, error) {
    var c Client
    var err error

    // Create a new net Dialer and set the timeout to be 10 seconds
    // Timeout is max time allowed to wait for a dial to connect
    //
    // We're using a timeout so we don't completely break the program
    // if we never get a new connection
    duraton := fmt.Sprintf("%ds", cost)
    timeout, _ := time.ParseDuration(duraton)
    dialer := net.Dialer{ Timeout: timeout }

    // Dial the connection adddress to establish the connection.
    c.Conn, err = dialer.Dial("udp", address)
    if err != nil {
        e := errors.Wrapf(err, "NewClient(%s): failed to create connection to address", address)
        return &c, e
    }

    return &c, nil
} // }}}

// func c.SendPacket {{{
//
// Sends the provided packet to the client connection and then closes the connection
func (c *Client) SendPacket(packet []byte, app types.Application) {
    // Defer closing our client connection
    defer c.close()

    // Write the packet to the connection
    _, err := c.Conn.Write(packet)
    if err != nil {
        // Some error occurred ..
        // Create a new error message, print it to the user and return
        e := errors.Wrapf(err, "failed to write update packet %+v", packet)
        app.OutErr("%+v\n", e)
    }
} // }}}

// func c.Close {{{
//
// Closes the UDP connection
func (c *Client) close() {
    if c.Conn != nil {
        c.Conn.Close()
    }
} // }}}
