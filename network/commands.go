package network

import "github.com/pkg/errors"

// DisplayTable displays the routers table
func (r *Router) DisplayTable() {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.log.OutServer("\n dst id |  next hop id  | link cost\n")
    r.log.OutServer("--------+---------------+-----------\n")

    var i uint16 = 1
    for ; i <= uint16(NumServers); i++ {
        //r.log.OutDebug("Table: %+v\n", r.table[i])
        if r.table[i].linkCost == Inf || r.table[i].linkCost == 0 {
            continue
        }
        r.log.OutServer("%d\t|\t%d\t| %d \n", r.table[i].ID, r.table[i].nextHop, r.table[i].linkCost)
    }

    r.log.OutApp("\nPlease enter a command: ")
}

// Update updates the link cost between to servers
func (r *Router) Update(id1, id2 uint16, newCost int) error {
    r.mu.Lock()

    id := r.ID
    if id1 == id || id2 == id {
        tableUp := make(map[uint16]tableUpdate, len(r.table))
        for id, server := range r.table {
            if id == id1 && id1 != id {
                t := tableUpdate{
                    ID: server.ID,
                    Cost: newCost,
                }
                tableUp[id] = t
                continue
            }
            if id == id2 && id2 != id {
                t := tableUpdate{
                    ID: server.ID,
                    Cost: newCost,
                }
                tableUp[id] = t
                continue
            }
            t := tableUpdate{
                ID: server.ID,
                Cost: server.linkCost,
            }
            tableUp[id] = t
        }

        rt := routingTable{
            ID: id,
            Table: tableUp,
        }

        go r.sendToNeighbors()
        go r.UpdateTable(rt)
    }
    r.mu.Unlock()

    if id1 != id {
        packet, err := r.createDifferentSenderPacket(id1, id2, newCost)
        if err != nil {
            return errors.Wrapf(err, "r.Update: failed to perform update, couldn't create sender packet")
        }
        err = r.SendPacket(packet, id1, id2)
        if err != nil {
            return errors.Wrapf(err, "r.Update: failed to perform update, couldn't send the packet to other server ")
        }
    }

    if id2 != id {
        packet, err := r.createDifferentSenderPacket(id2, id1, newCost)
        if err != nil {
            return errors.Wrapf(err, "r.Update: failed to perform update, couldn't create sender packet")
        }
        err = r.SendPacket(packet, id2, id1)
        if err != nil {
            return errors.Wrapf(err, "r.Update: failed to perform update, couldn't send the packet to other server ")
        }
    }

    return nil
}

// Disable disables a link between two routers
func (r *Router) Disable(id uint16) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if id == r.ID {
        return errors.Wrapf(nil, "r.Disable: failed to disable link, cannot disable link to self")
    }

    if r.table[id].directCost == Inf {
        return errors.Wrapf(nil, "r.Disable: failed to disable link, cannot disable a non neighbor link")
    }

    r.table[id].directCost = Inf
    r.table[id].linkCost = Inf

    tableUp := make(map[uint16]tableUpdate, len(r.table))
    for _, server := range r.table {
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

    go r.sendToNeighbors()
    go r.UpdateTable(rt)
    return nil
}
