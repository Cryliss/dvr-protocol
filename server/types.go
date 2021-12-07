// Package server provides server functionality for the Chat application
package server

import (
	"dvr/log"
	"dvr/types"
	"errors"
	"net"
	"sync"
)

var StepErr error = errors.New("the server crashed")
var ByeErr error = errors.New("stopped checking for updates")

// type Server struct {{{

// Server holds private information related to the server
type Server struct {
	// The ID of the server
	ID uint16

	// The address we want to bind to & accept connections on
	bindy string

	// Locks reading on this struct, avoids data races!
	mu sync.Mutex

	// Listener that will accept incoming packets
	listener net.PacketConn

	// The network router for the server
	router types.Router

	// Number of packets the server has received
	packets int

	// Logger for terminal logging
	log *log.Logger

	// Channel that will inform us if we need to stop sending update
	// messages or not
	bye chan struct{}

	active bool 

	// Channel that we'll send incoming packets on
	packetChan chan []byte
} // }}}
