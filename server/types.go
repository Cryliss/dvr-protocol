package server

import (
	"dvr-protocol/types"
	"net"
	"sync"
	"time"
)

// Server To store attributes related to our server
type Server struct {
	// The ID of the server
	ID uint16

	// The address we want to bind to & listen for packets on
	Bindy string

	// Our application so we can display messages to the user
	//
	// We are using types.Application rather than app.Application
	// because that causes an import loop which is not allowed ...
	app types.Application

	// The parsed topology information from the topology file
	t *Topology

	// The number of packets this server has received since the
	// last time packets was called
	p int

	// Interval of time between updates
	upint time.Duration

	// A sync map for our neighbor servers
	//
	// This allows us to use atomics to safely get new connection
	// ID values, without fear of data races
	neighbors sync.Map

	// List of connection IDs .. only using this so that my list
	// of connections will be sorted .. doesn't work that way if I just
	// range over the sync map
	ids []int

	// Locks reading on this struct, avoids data races!
	mu sync.Mutex

	// Listener that will accept incoming messages on
	listener net.PacketConn

	// Channel that will inform us if we need to stop sending update
	// messages or not
	bye chan struct{}
}

// Topology for details related to our network topology
type Topology struct {
	// Total number of servers in the topology network
	NumServers int

	// Number of neighbors the host server has
	NumNeighbors int

	// Map of all the neighboring servers
	Neighbors map[int]*Neighbor

	// Routing table to hold all the servers link costs
	Routing RoutingTable

	// Locks reading on this struct, avoids data races!
	mu sync.Mutex
}

// RoutingTable to store the topology routing table
type RoutingTable [][]int

// Neighbor for details related to the servers neighbors
type Neighbor struct {
	// The ID of the server, as a uint32 as that is the atomic
	// type we are using, and we're going to use this ID# to
	// load the neighbor from the sync.Map in the server.
	ID uint16

	// The address we want to bind to & listen for packets on
	Bindy string

	// The link cost between the neighbor and the host server
	Cost int

	// The number of times we failed to send the routing update
	failed int

	// Whether or not this link is disabled
	disabled bool

	// Last time this neighbor was updated
	ts time.Time

	// Locks reading on this struct, avoids data races!
	mu sync.Mutex
}
