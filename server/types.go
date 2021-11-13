package server

import (
    "dvr-protocol/types"
    "net"
    "sync"
    "time"
)

type Server struct {
    // The ID of the server
    Id uint16

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

    // The next availabe connection ID
    //
    // Only access this using atomics!
    nextID uint16

    // Listener that will accept incoming messages on
    listener net.PacketConn
} // }}}

type Topology struct {
    NumServers  int
    NumNeighbors int
    Neighbors   map[int]*Neighbor
    Routing     RoutingTable
} // }}}

type RoutingTable [][]int

type Neighbor struct {
    // The ID of the server, as a uint32 as that is the atomic
    // type we are using, and we're going to use this ID# to
    // load the neighbor from the sync.Map in the server.
    Id uint16

    // The address we want to bind to & listen for packets on
    Bindy string

    Cost int
} // }}}
