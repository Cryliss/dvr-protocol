package network

import "fmt"

// routerThread is a thread for handling routing table updates
func (r *Router) routerThread() {
    for {
        select {
        case update := <-r.UpdateChan:
            r.UpdateTable(update)
        }
    }
}

// routerThread is a thread for handling new packet updates
func (r *Router) packetThread() {
    for {
        select {
        case packet := <- r.PacketChan:
            r.newPacket(packet)
            if r.log.Debug {
                r.DisplayTable()
            }
        }
    }
}

// UpdateTable updates the routing table
func (r *Router) UpdateTable(rt routingTable) {
    r.mu.Lock()
    defer r.mu.Unlock()

    var updated bool

    var senderID uint16
    for dest, n := range rt.Table {
        if dest == n.ID && n.Cost == 0 {
            senderID = n.ID
        }
    }

    for destination, n := range rt.Table {
        _, ok := r.table[destination]

        if destination == senderID || destination == r.ID {
            continue
        }

        cost := n.Cost + rt.Table[r.ID].Cost
        if ok {
            // If our cost is larger than the incoming cost, update our table.
            if r.table[destination].linkCost > cost && cost > 0 {
                r.table[destination].linkCost = cost
                r.table[destination].nextHop = senderID
                updated = true
            }
        } else {
            newn := neighbor{
                linkCost: n.Cost,
                ID: senderID,
            }
            r.table[destination] = &newn
            updated = true
        }
    }

    if updated {
        go r.sendToNeighbors()
    }
}

// sendToNeighbors sends routing table updates to the neighboring routers
// in the network
func (r *Router) sendToNeighbors() {
    r.mu.Lock()
    defer r.mu.Unlock()

    tableUp := make(map[uint16]tableUpdate, len(r.table))
    for id, server := range r.table {
        t := tableUpdate{
            ID: server.ID,
            Cost: server.linkCost,
        }
        tableUp[id] = t
    }

    rt := routingTable{
        ID: r.ID,
        Table: tableUp,
    }

    channels := r.network.Channels

    for id, c := range channels {
        if id == r.ID {
            continue
        }

        c <- rt
    }
}

// GetNeighborID returns the ID of the neighbor associated with the provided port
func (r *Router) GetNeighborID(port string) uint16 {
    var id uint16
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, server := range r.table {
        p := fmt.Sprintf("%d", server.port)
        if p == port {
            return server.ID
        }
    }
    return id
} // }}}
