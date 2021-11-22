package server

import (
	"dvr-protocol/types"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/pkg/errors"
)

var inf int = 99999

// New initializes and returns a new server
func New(file string, interval int) *Server {
	// Parse the given topology file
	t, id, err := ParseTopologyFile(file)
	if err != nil {
		log.Fatalf("server.New(%s, %d): error parsing topology file: %v", file, interval, err)
		return &Server{}
	}

	// Create a new server
	s := &Server{
		Id:  id,
		t:   t,
		bye: make(chan struct{}, 0),
	}

	// Initialize the neighbors sync map
	s.initializeNeighbors()

	// Set the update interval for the routing updates
	inv := fmt.Sprintf("%ds", interval)
	s.upint, err = time.ParseDuration(inv)
	if err != nil {
		log.Fatalf("server.New(%s, %d): error parsing update interval! %v", file, interval, err)
		return s
	}

	return s
}

// initializeNeighbors initializes the neighbors sync map
func (s *Server) initializeNeighbors() {
	for _, n := range s.t.Neighbors {
		if n.Id == s.Id {
			s.Bindy = n.Bindy
			n.Cost = 0
		}

		// Add the new neighbor to our sync map
		s.neighbors.Store(n.Id, n)
		s.ids = append(s.ids, int(n.Id))
	}

	// Sort the array of ids.
	sort.Ints(s.ids)
}

// SetApplication sets the application of the server, since we make the server
// prior to making the application
func (s *Server) SetApplication(app types.Application) {
	s.app = app
}

// Topology returns the servers topology struct
func (s *Server) Topology() *Topology {
	return s.t
}

// Update sets the link cost between two neighbors to the given cost
func (s *Server) Update(id1, id2 uint16, newCost int) error {
	if s.Id == id1 {
		// Try loading our connection from the sync map
		_, ok := s.neighbors.Load(id2)

		// Check if it was loaded or not - if it didin't its not
		// its not one our neighbors
		if !ok {
			return errors.Errorf("s.Update(%d, %d): error updating link, link id not found.", id1, id2)
		}

		s.t.mu.Lock()
		n := s.t.Neighbors[int(id2)]
		s.t.mu.Unlock()

		// Update the neighbors link cost
		n.mu.Lock()
		if n.Cost == inf {
			n.mu.Unlock()
			return errors.Errorf("s.Update(%d, %d): error updating link, the server to update is not active", id1, id2)
		}
		n.Cost = newCost
		n.mu.Unlock()
	} else if s.Id == id2 {
		// Try loading our connection from the sync map
		_, ok := s.neighbors.Load(id1)

		// Check if it was loaded or not - if it didin't its not
		// its not one our neighbors
		if !ok {
			return errors.Errorf("s.Update(%d, %d): error updating link, link id not found.", id1, id2)
		}

		s.t.mu.Lock()
		n := s.t.Neighbors[int(id1)]
		s.t.mu.Unlock()

		// Update the neighbors link cost
		n.mu.Lock()
		if n.Cost == inf {
			n.mu.Unlock()
			return errors.Errorf("s.Update(%d, %d): error updating link, the server to update is not active", id1, id2)
		}
		n.Cost = newCost
		n.mu.Unlock()
	}

	// Set the link cost in the routing table to the new cost
	// NOTE: Costs are bi-directional, so update both entries in the routing table
	s.t.mu.Lock()
	s.t.Routing[int(id1)-1][int(id2)-1] = newCost
	s.t.Routing[int(id2)-1][int(id1)-1] = newCost
	rt := s.t.Routing
	s.t.mu.Unlock()

	s.updateRoutingTable(rt)
	s.Step()

	return nil
}

// Step sends the routing update immediately, instead of waiting for the update interval
func (s *Server) Step() error {
	// Prepare the message packet
	packet, err := s.preparePacket()
	if err != nil {
		return errors.Errorf("s.Step: failed to send prepare packet for update: %+v", err)
	}

	// Send the update messages
	if err := s.sendUpdates(packet); err != nil {
		return errors.Errorf("s.Step: failed to send packet update: %+v", err)
	}

	return nil
}

// Packets prints the number of packets the server has received since the last time
// this function was called.
func (s *Server) Packets() error {
	s.mu.Lock()
	packets := s.p
	s.p = 0
	s.mu.Unlock()

	s.app.OutCyan("Number of packets received since last call %d\n", packets)
	return nil
}

// Display displays the current routing table.
func (s *Server) Display() error {
	s.app.OutCyan("\nsrc id | next hop id | link cost\n")
	s.app.OutCyan("-------+-------------+-----------\n")

	// Let's grab the ids the server currently has
	s.mu.Lock()
	ids := s.ids
	sid := s.Id
	s.mu.Unlock()

	// Range over our array of ids and print them.
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

		// Make sure we don't add ourself or disabled links to the table
		if n.Cost == 0 || n.Cost == inf {
			continue
		}

		s.app.OutCyan("   %d   |      %d      |    %d \n", sid, n.Id, n.Cost)
	}
	s.app.Out("\n")

	return nil
}

// Disable disables the link between two servers.
func (s *Server) Disable(id uint16) error {
	// Try loading our connection from the sync map
	_, ok := s.neighbors.Load(id)

	// Check if it was loaded or not
	if !ok {
		return errors.Errorf("s.Disable(%d): error disabling link, link id not found", id)
	}

	s.t.mu.Lock()
	n := s.t.Neighbors[int(id)]
	s.t.Routing[int(s.Id)-1][int(n.Id)-1] = inf
	s.t.Routing[int(id)-1][int(s.Id)-1] = inf
	s.t.mu.Unlock()

	// Set the link cost to infinity (-1) to indicate its been disabled
	n.mu.Lock()
	n.disabled = true
	n.Cost = inf
	n.mu.Unlock()

	// Prepare the message packet
	packet, err := s.preparePacket()
	if err != nil {
		return errors.Errorf("s.Step: failed to send prepare packet for update: %+v", err)
	}

	// Send the update messages
	if err := s.sendDisableUpdate(packet); err != nil {
		return errors.Errorf("s.Step: failed to send packet update: %+v", err)
	}

	return nil
}

// Crash simulates a server crashing
func (s *Server) Crash() {
	s.mu.Lock()
	// Closing s.bye will cause the s.Listen and the s.Loopy goroutines to stop
	close(s.bye)
	s.mu.Unlock()

	s.app.Out("Crashing server now .. bye!\n")
	os.Exit(0)
}
