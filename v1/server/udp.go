package server

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Code assistance: https://ops.tips/blog/udp-client-and-server-in-go/

// Define the max buffer size for the buffer to hold incoming packets
const maxBufferSize = 1024

// Listen creates a new packet listener and starts listening for new packets
func (s *Server) Listen() error {
	var err error

	// Lets set our server listener by calling net.ListenPacket to specifically create a UDP packet listener.
	s.listener, err = net.ListenPacket("udp", s.Bindy)
	if err != nil {
		log.Fatalf("server.Write: error creating a new packet listener! %v", err)
		return nil
	}

	// Defer closing the listener
	defer func() {
		s.mu.Lock()
		s.listener.Close()
		s.mu.Unlock()
	}()

	errChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	// Given that waiting for packets to arrive is blocking by nature and we want
	// to be able of canceling such action if desired, we do that in a separate
	// go routine.
	go func() {
		for {
			// By reading from the connection into the buffer, we block until there's
			// new content in the socket that we're listening for new packets.
			//
			// Whenever new packets arrive, `buffer` gets filled and we can continue
			// the execution.
			n, _, err := s.listener.ReadFrom(buffer)
			if err != nil {
				// An error occurred so let's send it to our
				// error channel and return
				errChan <- err
				return
			}

			packet := buffer[:n]
			go s.newPacket(packet)
		}
	}()

	select {
	case err = <-errChan:
		if err != nil {
			s.app.OutErr("\ns.Listen: error reading packet := %+v\n\nPlease enter a command: ", err)
		}
	case _, ok := <-s.bye:
		if !ok {
			s.app.OutErr("\ns.Listen: our bye channel was closed! The server must have crashed!\n")
			return nil
		}
	}

	return nil
}

// newPacket handles unmarshaling and dealing with the new packet received.
func (s *Server) newPacket(packet []byte) {
	// Create a new message and try to unmarshal the packet into it
	var msg = &Message{}
	if err := UnmarshalMessage(packet, msg); err != nil {
		s.app.OutErr("\ns.newPacket(%+v): Error unmarshaling packet! err = %+v\n\nPlease enter a command: ", packet, err)
		return
	}

	// Incrememnt the number of packets received.
	s.mu.Lock()
	s.p++
	s.mu.Unlock()

	// Retrieve the sender ID & Port #
	senderPort := fmt.Sprintf("%d", msg.hPort)
	senderID := s.t.GetNeighborID(senderPort)

	// Let the user know we just got a new packet
	s.app.Out("\nRECEIVED A MESSAGE FROM SERVER %d\n\nPlease enter a command: ", senderID)

	// Create an index into our routing table (IDs start at 1, but we're zero indexed..)
	x := int(senderID) - 1

	s.t.mu.Lock()
	rt := s.t.Routing
	s.t.mu.Unlock()

	// Loop through each of our message neighbors and update the routing table
	// for the neighbor and the neighbor link costs accordingly.
	for _, n := range msg.n {
		rt[x][int(n.nID)-1] = int(n.nCost)
		if n.nID == s.ID {
			if nb, ok := s.t.Neighbors[int(senderID)]; ok {
				nb.mu.Lock()
				if !nb.disabled {
					rt[int(n.nID)-1][x] = int(n.nCost)
					nb.Cost = int(n.nCost)
					nb.ts = time.Now()
				}
				nb.disabled = false
				nb.mu.Unlock()
			}
		}
	}

	// Update the rest of the routing table
	s.updateRoutingTable(rt)
}
