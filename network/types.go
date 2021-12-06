package network

import (
    "dvr/log"
    "errors"
    "sync"
    "time"
)

// DisErr is the error message to display on disable error - non-neighbor
var DisErr error = errors.New("cannot disable a non neighbor link")
// DisSErr is the error message to display on disable error - self
var DisSErr error = errors.New("cannot disable link to youself")
// SendErr is the error message to display on send error - self 
var SendErr error = errors.New("cannot send packet to yourself")

type tableUpdate struct {
    ID uint16
    Cost int
}

// routingTable holds the routing table for the router
type routingTable struct {
    ID uint16
    Table map[uint16]tableUpdate
}

// Router holds the routing information for a server
type Router struct {
    ID uint16
    network *Network
    table map[uint16]*neighbor
    PacketChan chan []byte
    UpdateChan chan routingTable
    log *log.Logger
    mu sync.RWMutex
}

// Network holds routing information for all servers on the network
type Network struct {
    Channels map[uint16]chan routingTable
    Routers map[uint16]*Router

    mu sync.RWMutex
}

// neighbor is essentially a server, but I want my server to
// have different information than this lol
type neighbor struct {
    ID uint16

    // The UDP address that we bind connections to this neighbor to
    bindy string

    IP string
    port int

    active bool

    // The direct cost between source and destination servers
    directCost int

    // The last cost of the link that was in the routing table
    lastCost int

    // The current link cost between the server & destination that's been
    // updated using the bellman-ford algorithm
    linkCost int

    // The server that'll be hopped to in order to get to destination server
    nextHop uint16

    // The last time this neighbors was updated (last time we received a
    // packet from this neighbor)
    updated time.Time

    // The last time we forwarded a message to a neighbor
    forwarded time.Time

    // Whether or not this link has been disabled or updated, will be used
    // to keep data in sync
    changed bool

    mu sync.Mutex
}
