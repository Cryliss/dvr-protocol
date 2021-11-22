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
		rt = append(rt, yRt)
	}
	s.app.OutCyan("\nInitialized routing table:\n%+v\n", rt)
	s.t.Routing = rt
	return nil
}

// updateRoutingTable updates the servers routing table.
func (s *Server) updateRoutingTable(rt RoutingTable) error {
	x := int(s.Id) - 1

	for y := 0; y < s.t.NumServers; y++ {
		for v := 1; v < s.t.NumServers; v++ {
			//fmt.Printf("rt[x][y]: %d, rt[y][y]: %d, rt[x][v]: %d, rt[v][y]: %d\n", rt[x][y], rt[y][y], rt[x][v], rt[v][y])
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
			cxy := rt[x][y]
			dyy := rt[y][y]

			cxv := rt[x][v]
			dvy := rt[v][y]

			minf := math.Min(float64(cxy+dyy), float64(cxv+dvy))
			min := int(minf)

			if min > inf || min == int(34463) {
				min = inf
			}
			if rt[x][y] < min {
				continue
			}
			rt[x][y] = min
		}

		s.t.mu.Lock()
		n := s.t.Neighbors[y+1]
		s.t.mu.Unlock()

		n.mu.Lock()
		if !n.disabled {
			n.Cost = rt[x][y]
		}
		n.mu.Unlock()
	}

	s.t.mu.Lock()
	s.t.Routing = rt
	s.t.mu.Unlock()

	s.app.OutCyan("\nNew Routing Table:\n%+v\n\nPlease enter a command: ", rt)
	return nil
}
