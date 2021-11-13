package server

import (
    "context"
    "log"
    "net"
    "time"
)

//https://ops.tips/blog/udp-client-and-server-in-go/

// maxBufferSize specifies the size of the buffers that
// are used to temporarily hold data from the UDP packets
// that we receive.
const maxBufferSize = 1024

// serve wraps all the UDP echo server functionality.
// ps.: the server is capable of answering to a single
// client at a time.
func (s *Server) Listen(ctx context.Context) error {
    var err error

    // Lets set our server listener by calling net.ListenPacket to specifically create a UDP packet listener.
	s.listener, err = net.ListenPacket("udp", s.Bindy)
	if err != nil {
        log.Fatalf("server.Write: error creating a new packet listener! %v", err)
        return nil
	}

    errChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	// Given that waiting for packets to arrive is blocking by nature and we want
	// to be able of canceling such action if desired, we do that in a separate
	// go routine.
	go func() {
        // Create a timeout duration
        timeout, _ := time.ParseDuration("10s")
        for {
			// By reading from the connection into the buffer, we block until there's
			// new content in the socket that we're listening for new packets.
			//
			// Whenever new packets arrive, `buffer` gets filled and we can continue
			// the execution.
			//
			// note.: `buffer` is not being reset between runs.
			//	  It's expected that only `n` reads are read from it whenever
			//	  inspecting its contents.
			n, addr, err := s.listener.ReadFrom(buffer)
			if err != nil {
                // An error occurred so let's send it to our
                // error channel and return
				errChan <- err
				return
			}

			s.app.Out("packet-received: bytes=%d from=%s\n", n, addr.String())

            // Setting a deadline for the `write` operation allows us to not block
			// for longer than a specific timeout.
			//
			// In the case of a write operation, that'd mean waiting for the send
			// queue to be freed enough so that we are able to proceed.
			deadline := time.Now().Add(timeout)
			err = s.listener.SetWriteDeadline(deadline)
			if err != nil {
                // An error occurred so let's send it to our
                // error channel and return
				errChan <- err
				return
			}

			// Write the packet's contents back to the client.
			n, err = s.listener.WriteTo(buffer[:n], addr)
			if err != nil {
                // An error occurred so let's send it to our
                // error channel and return
				errChan <- err
				return
			}

			s.app.Out("packet-written: bytes=%d to=%s\n", len(buffer), addr.String())
            return
		}
	}()

	select {
	case <-ctx.Done():
        err = ctx.Err()
        s.app.OutErr("error reading packet, connection closed.")
	case err = <-errChan:
        if err != nil {
            err = ctx.Err()
            s.app.OutErr("error reading packet, connection closed.")
        }
	}

	return nil
}
