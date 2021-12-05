package server

import (
	"log"
	"net"
)

// Code assistance: https://ops.tips/blog/udp-client-and-server-in-go/

// Define the max buffer size for the buffer to hold incoming packets
const maxBufferSize = 1024

// Listen creates a new packet listener and starts listening for new packets
func (s *Server) Listen() error {
	var err error

	s.mu.Lock()
	bindy := s.bindy
	s.mu.Unlock()

	// Lets set our protocols listener by calling net.ListenPacket to
	// specifically create a UDP packet listener.
	s.listener, err = net.ListenPacket("udp", bindy)
	if err != nil {
		log.Fatalf("s.Listen: error creating a new packet listener! %s", err.Error())
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

			s.mu.Lock()
			s.packets++
			s.mu.Unlock()

			packet := buffer[:n]
			s.packetChan <- packet
		}
	}()

	select {
	case err = <-errChan:
		if err != nil {
			s.log.OutError("\ns.Listen: error reading packet := %+v\n\nPlease enter a command: ", err)
		}
	case _, ok := <-s.bye:
		if !ok {
			s.log.OutError("\ns.Listen: our bye channel was closed! The server must have crashed!\n")
			s.log.OutApp("\nPlease enter a command: ")
			return nil
		}
	}

	return nil
}
