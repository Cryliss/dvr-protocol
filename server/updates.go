package server

import (
    "net"
    "strconv"
    "strings"

    "github.com/pkg/errors"
)

type Client struct {
    conn net.Conn
}

func NewClient(address string) (*Client, error) {
    var c Client
    var err error
    raddr, err := net.ResolveUDPAddr("udp", address)
    if err != nil {
        e := errors.Wrapf(err, "NewClient(%s): failed to resolve UDP Address", address)
        return &c, e
    }

    c.conn, err = net.DialUDP("udp", nil, raddr)
    if err != nil {
        e := errors.Wrapf(err, "NewClient(%s): failed to create connection to address", address)
        return &c, e
    }

    return &c, nil
}

func (c *Client) Close() {
    if c.conn != nil {
        c.conn.Close()
    }
}

func (s *Server) Updates() error {
    bindyarr := strings.Split(s.Bindy, ":")
    port, _ := strconv.Atoi(bindyarr[1])

    updateMsg := &Message{
        nUpdates: uint16(s.t.NumNeighbors+1),
        hPort: uint16(port),
        hIP: bindyarr[0],
    }

    s.mu.Lock()
    neighbors := s.t.Neighbors
    s.mu.Unlock()

    var un map[uint16]*mNeighbor
    un = make(map[uint16]*mNeighbor)

    for _, n := range neighbors {
        if n.Cost == -1 {
            continue
        }
        bindyarr = strings.Split(n.Bindy, ":")
        port, _ = strconv.Atoi(bindyarr[1])

        updateNeighbor := &mNeighbor{
            nIP: bindyarr[0],
            nPort: uint16(port),
            nID: n.Id,
            nCost: uint16(n.Cost),
        }
        s.app.Out("Update neighbor: %+v\n", *updateNeighbor)
        un[n.Id] = updateNeighbor
    }
    updateMsg.n = un

    payload, err := updateMsg.Marshal()
    if err != nil {
        e := errors.Wrapf(err, "failed to marshal update message %+v", updateMsg)
        return e
    }

    s.app.Out("Payload: %+v\n", payload)
    s.app.Out("Update Message: %+v\n", updateMsg)

    s.app.Out("Testing unmarshalling .. \n")
    resp := &Message{}
    UnmarshalMessage(payload, resp)
    s.app.Out("Unmarshalled message: %+v\n", *resp)
    return nil
    //return s.SendUpdates(payload)
}

func (s *Server) SendUpdates(payload []byte) error {
    s.mu.Lock()
    neighbors := s.t.Neighbors
    s.mu.Unlock()

    for i := 0; i < s.t.NumServers; i++ {
        n, ok := neighbors[i]
        if !ok {
            continue
        }

        if n.Id == s.Id {
            continue
        }

        c, err := NewClient(n.Bindy)
        if err != nil {
            return err
        }

        go func() {
            defer c.Close()

            _, err = c.conn.Write(payload)
            if err != nil {
                e := errors.Wrapf(err, "failed to write update message payload %+v", payload)
                s.app.OutErr("%v", e)
            }

            s.app.Out("Update message sent to neighbor!")

            buf := make([]byte, 1024)
            _, err := c.conn.Read(buf)
            if err != nil {
                e := errors.Wrapf(err, "failed to read from client connection!")
                s.app.OutErr("%v", e)
            }

            resp := &Message{}
            err = UnmarshalMessage(buf, resp)
            if err != nil {
                e := errors.Wrapf(err, "failed to unmarshal client message from connection")
                s.app.OutErr("%v", e)
            }
            s.app.Out("MESSAGE: %+v\n", resp)
        }()
    }
    return nil
}
