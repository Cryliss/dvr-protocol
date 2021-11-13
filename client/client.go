package client

import (
    "fmt"
    "net"

    "github.com/pkg/errors"
)

type Client struct {
    conn net.Conn
}

func NewClient(address string) (*Client, error) {
    var c Client

    c.conn, err := net.Dial("udp", address)
    if err != nil {
        err = errors.Wrapf(err, "NewClient(%s): failed to create connection to address", address)
        return &c, err
    }

    return &c, nil
}

func (c *Client) Update() {
    var payload []byte

    updateMsg := &Message{

    }
}
