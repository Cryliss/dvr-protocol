package types

import "time"

// Server functionality ..
type Server interface {
    // Update performs the application update comand
    Update(id1, id2 uint16, newCost int) error

    // Step sends the routing update immediately, instead of waiting
    // for the update interval
    Step() error

    // Packets prints the number of packets the server has received
    // since the last time this function was called.
    Packets() error

    // Display displays the current routing table.
    Display() error

    // Disable disables the link between this server and another
    Disable(id uint16) error

    // Crash simulates a server crashing
    Crash() error
}

// Router interface ..
type Router interface {
    // Update sets the link cost between two neighbors to the given cost
    Update(id1, id2 uint16, newCost int) error
    // SendUpdates sends update packets to neighbors
    SendPacketUpdates() error
    // CheckUpdates checks for invalid links
    CheckUpdates(interval time.Duration) error 
    // DisplayTable displays the routing table
    DisplayTable()
    // Disable disables the link between this server and another
    Disable(id uint16) error
}
