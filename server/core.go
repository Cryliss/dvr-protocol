package server

import (
    //"context"
    "dvr-protocol/types"
    "errors"
    "fmt"
    "log"
    "sort"
    "time"
)

func New(file string, interval int) *Server {
    var s Server
    t, id, err := ParseTopologyFile(file)
    if err != nil {
        log.Fatalf("server.New: error parsing topology file: %v",err)
        return &s
    }
    s.Id = id
    s.t = t

    for _, n := range s.t.Neighbors {
        if n.Id == s.Id {
            s.Bindy = n.Bindy
            n.Cost = 0

            // Add the new neighbor to our sync map
            s.neighbors.Store(n.Id, n)
            s.ids = append(s.ids, int(n.Id))
            continue
        }
        // Add the new neighbor to our sync map
        s.neighbors.Store(n.Id, n)
        s.ids = append(s.ids, int(n.Id))
    }

    sort.Ints(s.ids)

    inv := fmt.Sprintf("%ds", interval)
    s.upint, err = time.ParseDuration(inv)
    if err != nil {
        log.Fatalf("server.New: error parse update interval! %v", err)
        return &s
    }

    return &s
} // }}}

// func s.SetApplication {{{
//
// Sets the application of the server, since we make the server
// prior to making the application
func (s *Server) SetApplication(app types.Application) {
    s.app = app
} // }}}

// func s.Topology {{{
//
// Returns the servers topology struct
func (s *Server) Topology() *Topology {
    return s.t
} //  }}}

// func s.Update {{{
//
// Sets the link cost between two neighbors to the given cost
func (s *Server) Update(id1, id2 uint16, newCost int) error {
    // Try loading our connection from the sync map
    _, ok := s.neighbors.Load(id2)

    // Check if it was loaded or not - if it didin't its not
    // its not one our neighbors
    if !ok {
        s.app.OutErr("s.Disable(%d): error updating link, link id not found.", id2)
        return nil
    }

    s.mu.Lock()
    n := s.t.Neighbors[int(id2)]
    s.mu.Unlock()

    n.Cost = -1

    //s.Write(context.Background(), n.Bindy, "test")

    s.app.Out("holder function for Update(%d, %d, %d)\n", id1, id2, newCost)
    return nil
} // }}}

// func s.Step {{{
//
// Sends the routing update immediately, instead of waiting for the update interval
func (s *Server) Step() error {
    s.app.Out("holder function for Step\n")
    return nil
} // }}}

// func s.Packets {{{
//
//
func (s *Server) Packets() error {
    s.mu.Lock()
    s.app.Out("Number of packets received since last call %d\n", s.p)
    s.p = 0
    s.mu.Unlock()

    return nil
} // }}}

// func s.RoutingTable {{{
//
// Displays the current routing table.
func (s *Server) RoutingTable() error {
    s.app.Out("\nsrc id | next hop id | link cost\n")
    s.app.Out("-------+-------------+-----------\n")

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

        s.mu.Lock()
        n := s.t.Neighbors[int(id)]
        s.mu.Unlock()

        if n.Cost == 0 || n.Cost == -1 {
            continue
        }

        s.app.Out("   %d   |      %d      |    %d \n", sid, n.Id, n.Cost)
    }
    s.app.Out("\n")

    return nil
} // }}}

// func s.Disable {{{
//
// Disables the link between two servers.
func (s *Server) Disable(id uint16) error {
    // Try loading our connection from the sync map
    _, ok := s.neighbors.Load(id)

    // Check if it was loaded or not - if it didin't its likely
    // been deleted from the map so just continue
    if !ok {
        e := fmt.Sprintf("s.Disable(%d): error disabling link, link id not found", id)
        return errors.New(e)
    }

    s.mu.Lock()
    n := s.t.Neighbors[int(id)]
    s.mu.Unlock()

    n.Cost = -1

    return nil
} // }}}

// func s.Crash {{{
//
// Simulates a server crashing
func (s *Server) Crash() error {
    s.mu.Lock()
    s.listener.Close()
    s.mu.Unlock()

    s.app.Out("holder function for crash\n")
    return nil
} // }}}
