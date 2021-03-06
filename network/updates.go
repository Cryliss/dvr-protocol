package network

import (
    "dvr/client"
    "dvr/message"
    "fmt"
    "time"

    "github.com/pkg/errors"
)

// newPacket handles new packet updates
func (r *Router) newPacket(packet []byte) {
    // Create a new message and try to unmarshal the packet into it
    var msg = &message.Message{}
    if err := message.UnmarshalMessage(packet, msg); err != nil {
        r.log.OutError("\nr.newPacket(%+v): Error unmarshaling packet! err = %+v\n", packet, err)
        r.log.OutApp("\nPlease enter a command: ")
    }

    // Retrieve the sender ID & Port #
    senderPort := fmt.Sprintf("%d", msg.Port)
    senderID := r.GetNeighborID(senderPort)

    // I was occasionally sending to myself, somehow?
    if senderID == r.ID {
        return
    }

    r.mu.Lock()
    // Get the last time this sender was updated &
    // set this sender to be active, since we received a
    // message from them
    updated := r.table[senderID].updated
    if !r.table[senderID].active {
        r.table[senderID].active = true
    }
    r.mu.Unlock()

    // Ensure we didn't *JUST* get sent this packet
    now := time.Now()
    oneSecAgo := now.Add(-1*time.Second)
    if updated.After(oneSecAgo) {
        return
    }

    // Let the user know we just got a new packet
    r.log.OutServer("\nRECEIVED A MESSAGE FROM SERVER %d\n", senderID)
    r.log.OutApp("\nPlease enter a command: ")

    // Check if we need to forward the packet to someone else
    if r.checkForwarding(senderID, packet) {
        r.log.OutServer("\nSUCCESSFULLY FORWARDED MESSAGE\n")
        r.log.OutApp("\nPlease enter a command: ")
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    // Set the updated time for the server
    r.table[senderID].updated = time.Now()

    tableUp := make(map[uint16]tableUpdate, len(r.table))
    // Loop through each of our message neighbors and update the routing table
    // for the neighbor and the neighbor link costs accordingly.
    for _, n := range msg.N {
        t := tableUpdate{
            ID: n.ID,
            Cost: int(n.Cost),
        }
        tableUp[n.ID] = t

        if t.ID == r.ID {
            if r.table[senderID].linkCost != t.Cost {
                if t.Cost == r.table[senderID].directCost {
                    r.table[senderID].nextHop = senderID
                }
                r.table[senderID].linkCost = t.Cost
            }
        }
	}

    upd := routingTable{
        ID: senderID,
        Table: tableUp,
    }

    channels := r.network.Channels
    // We'll send the update to *all* channels in the network,
    // including our own router
    for _, c := range channels {
        c <- upd
    }
}

// CheckUpdates checks the routers neighbors and see if they've been updated
// within 3 update intervals & disables them if not
func (r *Router) CheckUpdates(interval time.Duration) error {
    r.mu.Lock()

    // Check to see if we've gotten an update within the last 3
    // update intervals for each neighbor
    now := time.Now()
    threeUpdates := now.Add(-3 * interval)

    for _, server := range r.table {
        if server.ID == r.ID {
            continue
        }
        if server.updated.Before(threeUpdates) && r.table[server.ID].active {
            r.table[server.ID].active = false
            r.log.OutError("\nr.CheckUpdates: Haven't received an update from server (%d) in 3 intervals, disabling the link.\n", server.ID)
            r.log.OutApp("\nPlease enter a command: ")
            go r.Disable(server.ID)
        }
    }
    r.mu.Unlock()
    return nil
}

// SendPacketUpdates sends packet updates to neighboring links
func (r *Router) SendPacketUpdates() error {
    packet, err := r.preparePacket()
    if err != nil {
        return err
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    for id, server := range r.table {
        if server.linkCost != Inf && server.linkCost != 0 {
            bindy := r.table[id].bindy

            // Make sure we don't send packets we're not directly linked
            // to, or to ourself
            if server.directCost == Inf || server.directCost == 0 {
                continue
            }

            // Create a new client connection and send the packet
            c, err := client.NewClient(bindy)
            if err != nil {
                return errors.Wrapf(err, "r.sendUpdates: failed to send updates to neighbor %d - bindy: %s", id, bindy)
            }

            c.SendPacket(packet, r.log)
        }
    }
    return nil
}

// SendPacket handles sending a packet update to a single server
func (r *Router) SendPacket(packet []byte, src, dst uint16) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if r.ID == dst {
        return errors.Wrapf(SendErr, "r.SendPacket: failed to send packet")
    }

    server := r.table[dst]
    bindy := server.bindy

    // Is our direct cost to the destination inf?
    if server.directCost == Inf {
        // Get the nexthop server's bind address then
        bindy = r.table[server.nextHop].bindy
    }

    // Create a new client connection and send the packet
    c, err := client.NewClient(bindy)
    if err != nil {
        return errors.Wrapf(err, "r.SendPacket: failed to send packet to neighbor %d", server.ID)
    }

    c.SendPacket(packet, r.log)
    //r.log.OutServer("\nSENT PACKET TO %d\n", id)
    return nil
}

// preparePacket prepares an update packet
func (r *Router) preparePacket() ([]byte, error) {
    r.mu.Lock()

    neighbors := r.table
    numUpdates := len(neighbors)

    // Create a new update message
    updateMsg := &message.Message{
        Updates: uint16(numUpdates),
        Port:    uint16(neighbors[r.ID].port),
        IP:      neighbors[r.ID].IP,
    }
    r.mu.Unlock()

    // Create a new map for our update message neighbors to go into
    var un map[uint16]*message.Neighbor
    un = make(map[uint16]*message.Neighbor, numUpdates)

    // Create update neighbors for each neighbor our server has
    // and one for our server itself.
    var i uint16 = 1
	for ; i <= uint16(NumServers); i++ {
        // Let's get the neighbors information
        n := neighbors[i]

        // Create a new mNeighbor
        updateNeighbor := message.Neighbor{
            IP:   n.IP,
            Port: uint16(n.port),
            ID:   n.ID,
            Cost: uint16(n.linkCost),
        }

        // Uncomment this line to see how the update neighbor is formatted
        //r.log.OutDebug("Update neighbor: %+v\n", updateNeighbor)

        // Formatting looks like this --
        // Update neighbor: {nIP:192.168.0.104 nPort:2000 nID:1 nCost:7}

        // Add the neighbor to the update neighbor map
        un[n.ID] = &updateNeighbor
    }

    //r.log.OutDebug("neighbors? %+v\n", un)

    // Set the update message neighbors map equal to our update neighbor map
    updateMsg.N = un

    // Marshal the message into a packet to be sent
    packet, err := updateMsg.Marshal()
    if err != nil {
        e := errors.Wrapf(err, "failed to marshal update message %+v", updateMsg)
        return packet, e
    }

    //r.log.OutDebug("Packet: %+v\n", packet)
    //r.log.OutDebug("Update Message: %+v\n", updateMsg)

    return packet, nil
}

// createDifferentSenderPacket(senderID, neighborID) marshals
// a new message using the first servers information in the sender
// bytes & the second servers information in the the next set of bytes
func (r *Router) createDifferentSenderPacket(senderID, neighborID uint16, newCost int) ([]byte, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    sender := r.table[senderID]

    // Create a new update message
    updateMsg := &message.Message{
        Updates: uint16(1),
        Port:    uint16(sender.port),
        IP:      sender.IP,
    }

    // Create a new map for our update message neighbors to go into
    var un map[uint16]*message.Neighbor
    un = make(map[uint16]*message.Neighbor, 1)

    // Create an update neighbors for the neighbor
    neighbor := r.table[neighborID]

    // Create a new mNeighbor
    updateNeighbor := &message.Neighbor{
        IP:   neighbor.IP,
        Port: uint16(neighbor.port),
        ID:   neighbor.ID,
        Cost: uint16(newCost),
    }

    // Uncomment this line to see how the update neighbor is formatted
    //r.log.OutServer("Update neighbor: %+v\n", *updateNeighbor)

    // Formatting looks like this --
    // Update neighbor: {nIP:192.168.0.104 nPort:2000 nID:1 nCost:7}

    // Add the neighbor to the update neighbor map
    un[neighbor.ID] = updateNeighbor

    // Set the update message neighbors map equal to our update neighbor map
    updateMsg.N = un

    // Marshal the message into a packet to be sent
    packet, err := updateMsg.Marshal()
    if err != nil {
        e := errors.Wrapf(err, "failed to marshal update message %+v", updateMsg)
        return packet, e
    }

    //s.log.OutServer("\nUpdate Message: %+v\n", updateMsg)

    return packet, nil
}

// checkForwarding checks to see if a new packet should be forwarded
func (r *Router) checkForwarding(senderID uint16, packet []byte) bool {
    r.mu.Lock()
    defer r.mu.Unlock()

    forwarded := false

    r.network.mu.Lock()
    // Get the router for this sender from the network
    router := r.network.Routers[senderID]
    r.network.mu.Unlock()

    router.mu.Lock()
    for dest, server := range router.table {
        server.mu.Lock()

        // Is the nextHop our router & we're not the destination?
        if server.nextHop == r.ID && server.ID != r.ID {
            forwarded := server.forwarded

            // Make sure we haven't *JUST* forwarded a packet to this server
            now := time.Now()
            tenSecAgo := now.Add(-10*time.Second)
            if forwarded.Before(tenSecAgo) {
                r.table[server.ID].forwarded = time.Now()
                r.forwardPacket(packet, r.table[dest].bindy, dest)
                //r.log.OutDebug("\nFORWARDED PACKET FROM %d TO %d\n", senderID, dest)
            }
        }
        server.mu.Unlock()
    }
    router.mu.Unlock()
    
    return forwarded
}

// forwardPacket handles forwarding the packet to the other server
func (r *Router) forwardPacket(packet []byte, bindy string, id uint16) error {
    // Create a new client connection and send the packet
    c, err := client.NewClient(bindy)
    if err != nil {
        return errors.Wrapf(err, "r.forwardPacket: failed to forward packet to neighbor %d", id)
    }

    c.SendPacket(packet, r.log)
    return nil
}
