package network

import "github.com/pkg/errors"

// DisplayTable displays the routers table
func (r *Router) DisplayTable() {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.log.OutServer("\ndst | next hop | cost\n")
    r.log.OutServer("----+----------+-------\n")

    var i uint16 = 1
    for ; i <= uint16(NumServers); i++ {
        if r.table[i].linkCost == Inf || r.table[i].linkCost == 0 || r.table[i].nextHop == 0 {
            continue
        }
        r.log.OutServer(" %d  |\t  %d    |  %d\n", r.table[i].ID, r.table[i].nextHop, r.table[i].linkCost)
    }

    //r.log.OutApp("\nPlease enter a command: ")
}

// Update updates the link cost between to servers
func (r *Router) Update(id1, id2 uint16, newCost int) error {
    r.mu.Lock()

    id := r.ID
    var rt routingTable
    if id1 == id || id2 == id {
        tableUp := make(map[uint16]tableUpdate, len(r.table))
        for id, server := range r.table {
            if id == id1 && id1 != id {
                r.table[id].directCost = newCost
                r.table[id].linkCost = newCost
                t := tableUpdate{
                    ID: server.ID,
                    Cost: newCost,
                }
                tableUp[id] = t
                continue
            }
            if id == id2 && id2 != id {
                r.table[id].directCost = newCost
                r.table[id].linkCost = newCost
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

        rt.ID = id
        rt.Table = tableUp
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

    if rt.Table != nil {
        go r.UpdateTable(rt)
    }

    return nil
}

// Disable disables a link between two routers
func (r *Router) Disable(id uint16) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if id == r.ID {
        return errors.Wrapf(DisSErr, "r.Disable: failed to disable link")
    }

    if r.table[id].directCost == Inf {
        return errors.Wrapf(DisErr, "r.Disable: failed to disable link")
    }

    if r.table[id].directCost != r.table[id].linkCost {
        n := r.table[id].nextHop
        if n != id {
            r.table[n].linkCost = r.table[n].directCost
            r.table[n].nextHop = n
        }
    }

    r.table[id].directCost = Inf
    r.table[id].linkCost = Inf

    tableUp := make(map[uint16]tableUpdate, len(r.table))
    for i, server := range r.table {
        if server.nextHop == id {
            r.table[i].linkCost = r.table[i].directCost
            r.table[i].nextHop = i
            t := tableUpdate{
                ID: server.ID,
                Cost: server.directCost,
            }
            tableUp[id] = t
            continue
        }
        if server.ID == id {
            r.table[i].directCost = Inf
            r.table[i].linkCost = Inf
            r.table[i].nextHop = uint16(Inf)
            t := tableUpdate{
                ID: server.ID,
                Cost: Inf,
            }
            tableUp[id] = t
            continue
        }
        r.table[i].linkCost = r.table[i].directCost
        r.table[i].nextHop = i
        t := tableUpdate{
            ID: server.ID,
            Cost: server.directCost,
        }
        tableUp[id] = t
    }

    rt := routingTable{
        ID: r.ID,
        Table: tableUp,
    }
    r.UpdateChan <- rt

    return nil
}
