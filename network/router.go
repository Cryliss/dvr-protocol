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
        //r.log.OutDebug("\ndest=%d|n.ID=%d|rt.ID=%d|r.ID=%d", destination, n.ID, rt.ID, r.ID)
        _, ok := r.table[destination]

        if senderID == destination {
            continue
        }

        cost := n.Cost + rt.Table[r.ID].Cost
        if ok {
            // If our cost is larger than the incoming cost, update our table.
            if r.table[destination].linkCost > cost && cost > 0 {
                r.table[destination].linkCost = cost
                if r.table[destination].linkCost > 12 {
                    r.table[destination].linkCost = Inf
                    r.table[destination].directCost = Inf
                    updated = true
                    continue
                }

                if _, ok := r.table[senderID]; ok {
                    //r.log.OutDebug("senderID: %d | sender.nexthop: %d | n.ID: %d | destination: %d | destination.nexthop: %d | r.ID: %d | rt.ID: %d\n", senderID, r.table[senderID].nextHop, n.ID, destination, r.table[destination].nextHop, r.ID, rt.ID)
                    if r.ID == rt.ID {
                        r.table[destination].nextHop = destination
                        //r.log.OutDebug("using destination value\n")
                        updated = true
                        continue
                    }
                    if r.table[senderID].nextHop == r.table[destination].nextHop  {
                        //r.log.OutDebug("using destination valu2e\n")
                        r.table[destination].nextHop = destination
                        updated = true
                        continue
                    }
                }
                r.table[destination].nextHop = senderID
                updated = true
            }
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
