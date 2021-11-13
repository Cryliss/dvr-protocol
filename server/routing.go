package server

// func s.initalizeRt {{{

// Initializes the routing table
func (s *Server) InitalizeRt() error {
    inf := -1

    s.mu.Lock()
    neighbors := s.t.Neighbors
    s.mu.Unlock()

    var rt RoutingTable

    for y := 1; y <= s.t.NumServers; y++ {
        var yRt []int
        for w := 1; w <= s.t.NumServers; w++ {
            // Is id # y = to our servers id?
            if y == int(s.Id) {
                // Nope, is the neighbor id we're looking at one of our neighbors?
                if n, ok := neighbors[w]; ok {
                    // Yep, set the routing table cost equal to our neighbors cost
                    yRt = append(yRt, n.Cost)
                    continue
                }
                // The id we're looking at is not one of our neighbors,
                // set the link cost to inf
                yRt = append(yRt, inf)
                continue
            }
            // We're not looking at our server, which means we don't care what
            // the actual routing costs are, because we don't have them yet..
            yRt = append(yRt, inf)
        }
        s.app.Out("%v\n", yRt)
        rt = append(rt, yRt)
    }
    s.t.Routing = rt
    // Send routing table
    return nil
} // }}}

/*
func checkRt() error {
    inf := -1
    s.mu.Lock()
    neighbors = s.t.Neighbors
    s.mu.Unlock()


}*/
