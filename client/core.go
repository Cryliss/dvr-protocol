package client
/*
import (
    "dvr-protocol/types"
    "errors"
    "fmt"
    "net"
    "strings"
    "time"
)

// func New {{{
//
// Initializes and returns a new Client struct
func New(conn *net.TCPConn, addr []string, id uint32, app types.Application) *Client {
    client := Client{
        app: app,
        Id: id,
        IP: addr[0],
        Port: addr[1],
        Conn: conn,
    }
    return &client
} // }}}

// func c.HandleClient {{{
//
// Handler for client connections - reads from the connection and displays
// the message to the user.
// help src: https://ipfs.io/ipfs/QmfYeDhGH9bZzihBUDEQbCbTc5k5FZKURMUoUvfmc27BwL/socket/tcp_sockets.html
func (c *Client) HandleClient() {
    // Defer closing the client
    defer c.Conn.Close()

    // Create a buffer for incoming messages
    bufferSize := 1024
    buffer := make([]byte, bufferSize)
    for {
        // Read the message from the connection -
        n, err := c.Conn.Read(buffer)
        if err != nil {
            // Closed?
            if errors.Is(err, net.ErrClosed) {
                // We were told to shutdown, so just return.
                // Some other goroutine logged the reason for the closure.
                return
            }

            e := fmt.Sprintf("%v", err)

            // EOF?
            if e == "EOF" {
                c.app.OutErr("\n\nPeer has terminated the connection - closing client# %d now.\n\nPlease enter a command: ", c.Id)
            }
            break
        }

        // We don't output empty messages, so check the length.
        if n > 0 {
            // Let's read the message from the buffer
            msg := string(buffer[:n])

            // I want to include the time received to the message output,
            // so this creates the variables we need to do that.
            now := time.Now()
            now.Format(time.Stamp)
            nowArr := strings.Split(now.String(), ".")
            ts := nowArr[0]

            // Print the message to the user
            c.app.Out("\n\n====================================\nNEW MESSAGE FROM %v:%v\n\n", c.IP, c.Port)
            c.app.Out("%s:\t%s\n\nEND MESSAGE\n===================================\n\nPlease enter a command: ", ts, msg)
        }
    }
} // }}}

// func c.closeConn {{{
//
// Handles closing connections, sending a message to the connection
// prior to closing it. Returns any errors that may occur
func (c *Client) CloseConn() error {
    // Try closing the connection and handle any errors that may happen
    err := c.Conn.Close()
    if err != nil {
        // Closed?
        if errors.Is(err, net.ErrClosed) {
            // We already closed it, just return
            return err
        }
        // Hmm, something else went wrong ..
        c.app.OutErr("c.CloseConn: error closing connection!\n")
        return err
    }

    // Connction was successfully closed, let the user know and return
    c.app.Out("Successfully closed connection %d\n", c.Id)
    return nil
} // }}}
*/
