package server

import (
	"math"
)

// InitializeRoutingTable initializes the routing table
func (s *Server) InitializeRoutingTable() error {
	s.mu.Lock()
	neighbors := s.t.Neighbors
	s.mu.Unlock()

	var rt RoutingTable

	for y := 1; y <= s.t.NumServers; y++ {
		var yRt []int
		for w := 1; w <= s.t.NumServers; w++ {
			// Is id # y = to our servers id?
			if y == int(s.ID) {
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
		rt = append(rt, yRt)
	}

	s.app.OutCyan("\nInitialized routing table:\n")
	for _, r := range rt {
		s.app.OutCyan("%+v\n",r)
	}
	s.t.Routing = rt
	return nil
}

// updateRoutingTable updates the servers routing table.
func (s *Server) updateRoutingTable(rt RoutingTable) error {
	x := int(s.ID) - 1
	s.t.mu.Lock()
	neighbors := s.t.Neighbors
	s.t.mu.Unlock()

	for y := 0; y < s.t.NumServers; y++ {
		n := neighbors[y+1]
		for v := 1; v < s.t.NumServers; v++ {
			// 34463 is some number I kept getting during testing that would
			// break other conditions checks cos that's not equal to inf lol
			// so this is to avoid that .. :D
			if rt[x][y] == int(34463) {
				rt[x][y] = inf
				rt[y][x] = inf
			}
			if rt[y][y] == int(34463) {
				rt[y][y] = inf
			}
			if rt[x][v] == int(34463) {
				rt[x][v] = inf
				rt[v][x] = inf
			}
			if rt[v][y] == int(34463) {
				rt[v][y] = inf
				rt[y][v] = inf
			}

			// Bellman-Ford Equation:
			// Dx(y) = min {cost(x,y)+Dy(y), cost(x,v)+Dv(y)}
			cxy := rt[x][y] // cost(x,y)
			dyy := rt[y][y] // Dy(y)

			cxv := rt[x][v] // cost(x,v)
			dvy := rt[v][y] // Dv(y)

			minf := math.Min(float64(cxy+dyy), float64(cxv+dvy))
			min := int(minf)

			if min > inf || min == int(34463) {
				min = inf
			}
			if rt[x][y] < min {
				continue
			}

			n.mu.Lock()
			// Let's make sure our link isn't disabled before we update the cost.
			if !n.disabled {
				rt[x][y] = min
			}
			n.mu.Unlock()
		}
		// Since we've checked all the neighbors, v, we can go ahead and
		// update our neighbors link cost now.
		n.mu.Lock()
		// Let's make sure our link isn't disabled before we update the cost.
		if !n.disabled {
			n.Cost = rt[x][y]
		}
		n.mu.Unlock()
	}

	// Update our routing table to reflect the changes
	s.t.mu.Lock()
	s.t.Routing = rt
	s.t.mu.Unlock()

	// Print the new routing table for debugging.
	s.app.OutCyan("\nNew Routing Table:\n")
	for _, r := range rt {
		s.app.OutCyan("%+v\n",r)
	}
	s.app.Out("\nPlease enter a command: ")
	return nil
}
