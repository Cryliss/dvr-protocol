package topology

// Topology setup for the network
type Topology struct {
    NumServers int
    NumNeighbors int
    Servers map[int]*Server
}

// Server details
type Server struct {
    ID uint16
    Cost int
    Bindy string
    IP string
    Port int
    Neighbor bool
}
