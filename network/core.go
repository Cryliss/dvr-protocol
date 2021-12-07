package network

import (
    "dvr/log"
    "dvr/topology"
    "dvr/server"
    "time"
)

var (
    // NumServers is the number of servers in the network
    NumServers int
    // Inf is the inifinity link cost value
    Inf int = 99999
)

// New initializes and returns a new network.
func New(top *topology.Topology, sid uint16, l *log.Logger) *server.Server {
    var n Network
    NumServers = top.NumServers
    n.Channels = make(map[uint16]chan routingTable, NumServers)
    s := n.parseTopology(top, sid, l)
    return s
}

// parseTopology will parse the topology configuration and create
// the necessary routers and server
func (n *Network) parseTopology(top *topology.Topology, sid uint16, l *log.Logger) *server.Server {
    var routers map[uint16]*Router
    routers = make(map[uint16]*Router, NumServers)

    var table map[uint16]*neighbor
    table = make(map[uint16]*neighbor, NumServers)

    var bindy string
    for _, server := range top.Servers {
        if server.ID == sid {
            bindy = server.Bindy

            s := neighbor{
                ID: server.ID,

                IP: server.IP,
                port: server.Port,
                bindy: server.Bindy,

                active: true,
                nextHop: uint16(0),
                directCost: 0,
                linkCost: 0,

                updated: time.Now(),
                forwarded: time.Now(),
            }
            table[server.ID] = &s
            continue
        }

        s := neighbor{
            ID: server.ID,

            IP: server.IP,
            port: server.Port,
            nextHop: uint16(0),
            bindy: server.Bindy,

            active: true,
            directCost: server.Cost,
            linkCost: server.Cost,

            updated: time.Now(),
            forwarded: time.Now(),
        }
        if s.directCost != Inf {
            s.nextHop = s.ID
        }
        table[server.ID] = &s
        r := n.createRouter(server.ID, server.Cost, l)
        routers[server.ID] = r
    }

    r := Router{
        ID: sid,
        network: n,
        table: table,
        PacketChan: make(chan []byte, 50000),
        UpdateChan: make(chan routingTable, 100),
        log: l,
    }
    go r.routerThread()
    go r.packetThread()

    routers[sid] = &r
    n.Routers = routers
    n.Channels[r.ID] = r.UpdateChan

    server := server.New(r.PacketChan, sid, bindy, &r, l)
    return server
}

// createRouter creates a new router that will be used to
// keep track of the other servers routing tables
func (n *Network) createRouter(id uint16, cost int, l *log.Logger) *Router {
    var table map[uint16]*neighbor
    table = make(map[uint16]*neighbor, NumServers)

    var i uint16 = 1
    for ; i <= uint16(NumServers); i++ {

        if i == id {
            s := neighbor{
                ID: i,
                active: true,
                nextHop: uint16(0),
                directCost: cost,
                linkCost: Inf,
                updated: time.Now(),
                forwarded: time.Now(),
            }
            table[i] = &s
            continue
        }

        s := neighbor{
            ID: i,
            nextHop: uint16(0),
            active: false,
            directCost: Inf,
            linkCost: Inf,
            updated: time.Now(),
            forwarded: time.Now(),
        }
        table[i] = &s
    }

    r := Router{
        ID: id,
        network: n,
        table: table,
        PacketChan: make(chan []byte, 50000),
        UpdateChan: make(chan routingTable, 100),
        log: l,
    }
    go r.routerThread()

    n.Channels[r.ID] = r.UpdateChan
    return &r
}
