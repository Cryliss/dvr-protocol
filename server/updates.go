package server

import (
	"dvr-protocol/client"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Loopy sends the routing updates at the specified time interval
func (s *Server) Loopy() error {
	// Basic tracking ticker, set to tick at the same time interval
	// as update interval
	tick := time.NewTicker(s.upint)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// Log the auto packet update
			s.app.OutCyan("\ns.Loopy: Sending packet update now..\n")
			s.app.Out("\nPlease enter a command: ")

			// Time to send a new packet update; prepare the message packet
			packet, err := s.preparePacket()
			if err != nil {
				s.app.OutErr("\ns.Loopy: failed to prepare packets for routing update! err = %+v\n", err)
				s.app.Out("\nPlease enter a command: ")
				continue
			}

			// Send the update messages
			if err := s.sendUpdates(packet); err != nil {
				s.app.OutErr("\ns.Loopy: failed to send routing updates! err = 5+v\n", err)
				s.app.Out("\nPlease enter a command: ")
				continue
			}

			// Log the suuccess of the update
			s.app.OutCyan("\ns.Loopy: Successfully sent packets!\n")
			s.app.Out("\nPlease enter a command: ")

			// Check to see if we've gotten an update within the last 3
			// update intervals for each neighbor
			now := time.Now()
			threeUpdates := now.Add(-3 * s.upint)

			s.t.mu.Lock()
			for _, id := range s.ids {
				if id == int(s.ID) {
					continue
				}
				n := s.t.Neighbors[id]
				n.mu.Lock()
				ts := n.ts
				id := n.ID
				n.mu.Unlock()
				if ts.Before(threeUpdates) {
					s.app.OutErr("\nHaven't received an update from server (%d) in 3 intervals, disabling the link.\n", n.ID)
					s.app.Out("\nPlease enter a command: ")

					// Safely disable our neighbor
					n.mu.Lock()
					n.disabled = true
					n.Cost = inf
					n.mu.Unlock()

					// Update our routing table entries
					s.t.Routing[int(s.ID)-1][int(id)-1] = inf
					for i := 0; i < s.t.NumServers; i++ {
						s.t.Routing[int(id)-1][i] = inf
					}
				}
			}
			s.t.mu.Unlock()
		case _, ok := <-s.bye:
			if !ok {
				e := errors.New("\ns.Loopy: our bye channel was closed! The server must have crashed")
				return e
			}
		}
	}
}

// preparePacket prepares the packet for the update messages
func (s *Server) preparePacket() ([]byte, error) {
	// Let's grab the ids of all the servers in the network
	s.mu.Lock()
	ids := s.ids
	s.mu.Unlock()

	// Let's grab the IP & port of the server from the bind address
	bindyarr := strings.Split(s.Bindy, ":")
	ip := bindyarr[0]
	port, _ := strconv.Atoi(bindyarr[1])

	// Create a new update message
	updateMsg := &Message{
		nUpdates: uint16(s.t.NumNeighbors + 1),
		hPort:    uint16(port),
		hIP:      ip,
	}

	// Create a new map for our update message neighbors to go into
	var un map[uint16]*mNeighbor
	un = make(map[uint16]*mNeighbor, s.t.NumNeighbors+1)

	// Create update neighbors for each neighbor our server has
	// and one for our server itself.
	for _, id := range ids {
		// Try loading our connection from the sync map
		_, ok := s.neighbors.Load(uint16(id))

		// Check if it was loaded or not - if it didin't its likely
		// been deleted from the map so just continue
		if !ok {
			continue
		}

		s.t.mu.Lock()
		n := s.t.Neighbors[id]
		s.t.mu.Unlock()

		// Let's grab the IP & port from the bind address
		bindyarr = strings.Split(n.Bindy, ":")
		ip = bindyarr[0]
		port, _ = strconv.Atoi(bindyarr[1])

		// Create a new mNeighbor
		updateNeighbor := &mNeighbor{
			nIP:   ip,
			nPort: uint16(port),
			nID:   n.ID,
			nCost: uint16(n.Cost),
		}

		// Uncomment this line to see how the update neighbor is formatted
		//s.app.Out("Update neighbor: %+v\n", *updateNeighbor)

		// Formatting looks like this --
		// Update neighbor: {nIP:192.168.0.104 nPort:2000 nID:1 nCost:7}

		// Add the neighbor to the update neighbor map
		un[n.ID] = updateNeighbor
	}

	// Set the update message neighbors map equal to our update neighbor map
	updateMsg.n = un

	// Marshal the message into a packet to be sent
	packet, err := updateMsg.Marshal()
	if err != nil {
		e := errors.Wrapf(err, "failed to marshal update message %+v", updateMsg)
		return packet, e
	}

	//s.app.Out("Packet: %+v\n", packet)
	//s.app.Out("Update Message: %+v\n", updateMsg)

	return packet, nil
}

// sendUpdates Sends the packet update to each neighboring server
func (s *Server) sendUpdates(packet []byte) error {
	for i := 1; i <= s.t.NumServers; i++ {
		// Is the id we're on our servers ID?
		if i == int(s.ID) {
			// Yep, let's keep going.
			continue
		}

		// Load our connection from the provided connection I
		_, ok := s.neighbors.Load(uint16(i))

		// Do we actually have a neighbor with that ID?
		if !ok {
			// Nope, let's keep going.
			continue
		}

		s.t.mu.Lock()
		n := s.t.Neighbors[i]
		s.t.mu.Unlock()

		// Is our link cost infinity?
		n.mu.Lock()
		if n.Cost == inf {
			// Yep, don't try to send the packet
			n.mu.Unlock()
			continue
		}

		// Create a new client connection and send the packet
		c, err := client.NewClient(n.Bindy, n.Cost)
		n.mu.Unlock()
		if err != nil {
			return errors.Wrapf(err, "s.sendUpdates: failed to send updates to neighbor %d", n.ID)
		}

		go c.SendPacket(packet, s.app)

		// Code to use the server to send messages .. not sure which I'm supposed
		// to do, but I could see this causing some locking issues

		/* Create a new net Dialer and set the timeout to be 10 seconds
		   // Timeout is max time allowed to wait for a dial to connect
		   //
		   // We're using a timeout so we don't completely break the program
		   // if we never get a new connection
		   duraton := fmt.Sprintf("%ds", n.Cost)
		   timeout, _ := time.ParseDuration(duraton)
		   deadline := time.Now().Add(timeout)

		   s.mu.Lock()
		   // Set a write deadline and send the packets contents
		   err := s.listener.SetWriteDeadline(deadline)
		   if err != nil {
		       return errors.Errorf("s.sendUpdates: failed to set write deadline. err := %+v", err)
		   }

		   // Get the net.UDPAddr for the neighbor
		   raddr, err := net.ResolveUDPAddr("udp", n.Bindy)
		   if err != nil {
		       return errors.Wrapf(err, "s.sendUpdates: error resolving udp add for server %d", i)
		   }

		   // Write the packet's contents to the neighboring server
		   _, err = s.listener.WriteTo(packet, raddr)
		   s.mu.Unlock()
		   if err != nil {
		       return errors.Wrapf(err, "s.sendUpdates: error sending routing update to server %d", i)
		   }*/
	}
	return nil
}

// sendUpdates Sends the packet update to each neighboring server
func (s *Server) sendDisableUpdate(packet []byte) error {
	for i := 1; i <= s.t.NumServers; i++ {
		// Is the id we're on our servers ID?
		if i == int(s.ID) {
			// Yep, let's keep going.
			continue
		}

		// Load our connection from the provided connection I
		_, ok := s.neighbors.Load(uint16(i))

		// Do we actually have a neighbor with that ID?
		if !ok {
			// Nope, let's keep going.
			continue
		}

		s.t.mu.Lock()
		n := s.t.Neighbors[i]
		s.t.mu.Unlock()

		// Create a new client connection and send the packet
		c, err := client.NewClient(n.Bindy, n.Cost)
		if err != nil {
			return errors.Wrapf(err, "s.sendUpdates: failed to send updates to neighbor %d", n.ID)
		}

		go c.SendPacket(packet, s.app)
	}
	return nil
}
