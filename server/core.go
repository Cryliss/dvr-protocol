// Package server provides server functionality
package server

import (
	"dvr/log"
	"dvr/types"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// func New {{{

// New initializes and returns a new Server.
func New(packetChan chan []byte, id uint16, bindy string, router types.Router, l *log.Logger) *Server {
	s := Server{
		ID:  id,
		bindy: bindy,
		log: l,
        bye: make(chan struct{}, 0),
		packetChan: packetChan,
		router: router,
	}

	// Return the new server
	return &s
} // }}}


// Loopy sends the routing updates at the specified time interval
func (s *Server) Loopy(updateInterval int) error {
	// Set the update interval for the routing updates
	inv := fmt.Sprintf("%ds", updateInterval)
	interval, err := time.ParseDuration(inv)
	if err != nil {
		return errors.Wrapf(err, "server.New: error parsing update interval '%d'", updateInterval)
	}

	// Basic tracking ticker, set to tick at the same time interval
	// as update interval
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// Log the auto packet update
			s.log.OutServer("\ns.Loopy: Sending packet update now..\n")
			s.log.OutApp("\nPlease enter a command: ")

			// Send the update messages
			if err := s.router.SendPacketUpdates(); err != nil {
				s.log.OutError("\ns.Loopy: failed to send routing updates! err = 5+v\n", err)
				s.log.OutApp("\nPlease enter a command: ")
			}

			// Log the suuccess of the update
			s.log.OutServer("\ns.Loopy: Successfully sent packets!\n")
			s.log.OutApp("\nPlease enter a command: ")

			if err := s.router.CheckUpdates(interval); err != nil {
				s.log.OutError("s.Loopy: error while checking updates - %s", err.Error())
			}
		case _, ok := <-s.bye:
			if !ok {
				e := errors.New("\ns.Loopy: our bye channel was closed! The server must have crashed")
				return e
			}
		}
	}
}

// Update sets the link cost between two neighbors to the given cost
func (s *Server) Update(id1, id2 uint16, newCost int) error {
    return s.router.Update(id1,id2,newCost)
}

// Step sends the routing update immediately, instead of waiting for the update interval
func (s *Server) Step() error {
	// Send the update messages
	if err := s.router.SendPacketUpdates(); err != nil {
		return errors.Errorf("s.Step: failed to send packet update: %+v", err)
	}

	return nil
}

// Packets prints the number of packets the server has received since the last time
// this function was called.
func (s *Server) Packets() error {
	s.mu.Lock()
	packets := s.packets
	s.packets = 0
	s.mu.Unlock()

	s.log.OutServer("Number of packets received since last call: %d\n", packets)
	return nil
}

// Display displays the current routing table.
func (s *Server) Display() error {
	s.router.DisplayTable()
	return nil
}

// Disable disables the link between this server and another
func (s *Server) Disable(id uint16) error {
    return s.router.Disable(id)
}

// Crash simulates a server crashing
func (s *Server) Crash() error {
	s.log.OutServer("Crashing server now .. bye!\n")
	s.mu.Lock()
	// Closing s.bye will cause the s.Listen and the s.Loopy goroutines to stop
	close(s.bye)
	s.mu.Unlock()
	return nil
}
