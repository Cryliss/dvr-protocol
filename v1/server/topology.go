package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ParseTopologyFile parses the provided topology file and returns the topology setup
func ParseTopologyFile(file string) (*Topology, uint16, error) {
	var t Topology
	t.Neighbors = make(map[int]*Neighbor, 4)

	var sid uint16

	// Open the file
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening topology file! %v", err)
	}
	defer f.Close()

	// Create a new bufio scanner so we can read line by line
	scanner := bufio.NewScanner(f)
	line := 1
	for scanner.Scan() {
		switch line {
		case 1:
			numServers, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return &t, sid, err
			}
			t.NumServers = numServers
			line++
			break
		case 2:
			numNeighbors, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return &t, sid, err
			}
			t.NumNeighbors = numNeighbors
			line++
			break
		case 3:
			fallthrough
		case 4:
			fallthrough
		case 5:
			fallthrough
		case 6:
			text := scanner.Text()
			textArr := strings.Split(text, " ")
			if len(textArr) != 3 {
				e := errors.Errorf("ParseTopologyFile: error parsing topology file, incorrect number of arguments in line %d", line)
				return &t, sid, e
			}

			tid, err := strconv.Atoi(textArr[0])
			if err != nil {
				e := errors.Errorf("s.ParseTopologyFile: error parsing topology file, non integer in first column of line %d", line)
				return &t, sid, e
			}
			id := uint16(tid)

			port := textArr[2]
			/* Project specification part 3.1 Topology Establishment
			   "The host server here is the one which will read this topology file).
			   Note: the IPs of servers may change when you are running the
			   program in a wireless network environment.
			   So, we need to use ifconfig or ipconfig to obtain the IP first
			   and then set up the topology file before the demo."

			   Calling GetOutboundIP right here makes it so we don't have to
			   do that at all. If this this is the first server (i.e. host),
			   then let's set the ip to the outbound ip of the machine
			*/
			ip := GetOutboundIP(port)

			n := Neighbor{
				ID:    id,
				Bindy: ip + ":" + port,
				Cost:  inf,
				ts:    time.Now(),
			}

			t.Neighbors[tid] = &n
			line++
			break
		case 7:
			text := scanner.Text()
			textArr := strings.Split(text, " ")
			if len(textArr) != 3 {
				e := errors.Errorf("ParseTopologyFile: error parsing topology file, incorrect number of arguments in line %d", line)
				return &t, sid, e
			}

			id1, err := strconv.Atoi(textArr[0])
			if err != nil {
				e := errors.Errorf("s.ParseTopologyFile: error parsing topology file, non integer in first column of line %d", line)
				return &t, sid, e
			}

			sid = uint16(id1)

			id2, err := strconv.Atoi(textArr[1])
			if err != nil {
				e := errors.Errorf("s.ParseTopologyFile: error parsing topology file, non integer in first column of line %d", line)
				return &t, sid, e
			}

			cost, err := strconv.Atoi(textArr[2])
			if err != nil {
				e := errors.Errorf("s.ParseTopologyFile: error parsing topology file, non integer in first column of line %d", line)
				return &t, sid, e
			}

			if _, ok := t.Neighbors[id2]; ok {
				t.Neighbors[id2].Cost = cost
			}

			line++
			break
		default:
			text := scanner.Text()
			textArr := strings.Split(text, " ")
			if len(textArr) != 3 {
				e := errors.Errorf("ParseTopologyFile: error parsing topology file, incorrect number of arguments in line %d", line)
				return &t, sid, e
			}
			// I could care less about the server id in textArr[0]
			id, err := strconv.Atoi(textArr[1])
			if err != nil {
				e := errors.Errorf("s.ParseTopologyFile: error parsing topology file, non integer in first column of line %d", line)
				return &t, sid, e
			}

			cost, err := strconv.Atoi(textArr[2])
			if err != nil {
				e := errors.Errorf("s.ParseTopologyFile: error parsing topology file, non integer in first column of line %d", line)
				return &t, sid, e
			}

			if _, ok := t.Neighbors[id]; ok {
				t.Neighbors[id].Cost = cost
			}

			line++
			break
		}
	}

	return &t, sid, nil
}

// GetOutboundIP gets the preferred outbound ip of this machine
// src: https://stackoverflow.com/a/37382208
func GetOutboundIP(port string) string {
	s := "8.8.8.8:" + port
	conn, err := net.Dial("udp", s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ip := fmt.Sprintf("%v", localAddr.IP)
	ipArr := strings.Split(ip, ":")
	return ipArr[0]
}

// GetNeighborID returns the ID of the neighbor associated with the provided port
func (t *Topology) GetNeighborID(port string) uint16 {
	var id uint16
	t.mu.Lock()
	neighbors := t.Neighbors
	t.mu.Unlock()

	for _, n := range neighbors {
		bindyarr := strings.Split(n.Bindy, ":")
		p := bindyarr[1]
		if p == port {
			id = n.ID
			break
		}
	}
	return id
} // }}}
