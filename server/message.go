package server

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "github.com/pkg/errors"
    "strings"
    "strconv"
)

// Format of the message:
//		    	               0  1  2  3  4  5  6  7                          0  1  2  3  4  5  6  7
//     0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F  0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |          NUMBER OF UPDATE FIELDS              |          HOST SERVER PORT NUMBER              |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |                                    HOST SERVER IP ADDRESS                                     |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |                                 HOST'S NEIGHBOR 1 IP ADDRESS                                  |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |         HOST'S NEIGHBOR 1 PORT NUMBER         |             0x0 [NULL BYTE]                   |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |       HOST'S NEIGHBOR 1 SERVER ID NUMBER      |         HOST'S NEIGHBOR 1 LINK COST           |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |                                 HOST'S NEIGHBOR 2 IP ADDRESS                                  |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |         HOST'S NEIGHBOR 2 PORT NUMBER         |             0x0 [NULL BYTE]                   |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |       HOST'S NEIGHBOR 2 SERVER ID NUMBER      |         HOST'S NEIGHBOR 2 LINK COST           |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |                                 HOST'S NEIGHBOR n IP ADDRESS                                  |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |         HOST'S NEIGHBOR n PORT NUMBER         |             0x0 [NULL BYTE]                   |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//   |       HOST'S NEIGHBOR n SERVER ID NUMBER      |         HOST'S NEIGHBOR n LINK COST           |
//   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//

// type Message struct {{{

// To store the unmarshalled message from the other host server
// The minimum possible size of an update message:
// 8 bytes for the host servers information
// A minimum 10 bytes per neighbor for their information
//          + 2 bytes for the null byte
// With a minimum of 2 expected neighbors, the minimum size of the message
// is expected to be 8 + 2(10+2) = 32 bytes
type Message struct {
    nUpdates uint16                 // Number of expected fields
    hPort   uint16                  // Port of the host server sending the msg
    hIP     string                  // IP of the host server sending the msg
    n       map[uint16]*mNeighbor   // Map of the neighbor information in the hosts routng table
} // }}}

// type mNeighbor struct {{{

// To store the information about the host servers neighbors
// Total size of each neighbors info is 10 bytes
type mNeighbor struct {
    nIP     string // Neighbor IP
    nPort   uint16 // Neighbor Port
    nID     uint16 // Neighbor ID
    nCost   uint16 // Neighbor Link cost
} // }}}

// func UnmarshalMessage {{{

// Unmarshalled message should like like:
//  {nUpdates:3 hPort:2001 hIP:192.168.200.80 n:map[1:0xc00000e318 2:0xc00000e348 3:0xc00000e378]}
// With the map containing
//  {nIP:192.168.200.80 nPort:2000 nID:1 nCost:7}
//  {nIP:192.168.200.80 nPort:2001 nID:2 nCost:0}
//  {nIP:192.168.200.80 nPort:2002 nID:3 nCost:2}
func UnmarshalMessage(msg []byte, m *Message) error {
    var neighbors map[uint16]*mNeighbor
    neighbors = make(map[uint16]*mNeighbor, 4)

    // Did we actually get a message struct?
    if m == nil {
		err := errors.Errorf("header must be non-nil")
		return err
	}

    // Check if the message is of the correct length
    if len(msg) < 32 {
        err := errors.Errorf("msg does not have the expected size - %d", len(msg))
        return err
    }

    // The number of expected updates is formed by the first two bytes such that
	// it results in an uint16.
    //
	// As this comes from the network in UDP packets we can assume that it comes
	// in BigEndian (network byte order).
    m.nUpdates = uint16(msg[1]) | uint16(msg[0]<<8)
    m.hPort = uint16(msg[3]) | uint16(msg[2])<<8
    m.hIP = fmt.Sprintf("%d.%d.%d.%d\n", msg[4], msg[5], msg[6], msg[7])

    // mNeighbor unmarshalled: {nIP:192.168.200.80 nPort:2000 nID:1 nCost:7}
    for b := 8; b <= len(msg)-12; b += 12 {
        var n mNeighbor

        n.nIP = fmt.Sprintf("%d.%d.%d.%d\n", msg[b], msg[b+1], msg[b+2], msg[b+3])
        n.nPort = uint16(msg[b+5]) | uint16(msg[b+4])<<8
        n.nID = uint16(msg[b+9]) | uint16(msg[b+8])<<8
        n.nCost = uint16(msg[b+11]) | uint16(msg[b+10])<<8

        neighbors[n.nID] = &n
    }
    m.n = neighbors
    return nil
}

func (m *Message) Marshal() ([]byte, error) {
    buf := new(bytes.Buffer)

    binary.Write(buf, binary.BigEndian, m.nUpdates)
    binary.Write(buf, binary.BigEndian, m.hPort)

    iparr := strings.Split(m.hIP, ".")

    ip064, _ := strconv.ParseUint(iparr[0], 10, 8)
    ip0 := uint8(ip064)
    ip164, _ := strconv.ParseUint(iparr[1], 10, 8)
    ip1 := uint8(ip164)
    ip264, _ := strconv.ParseUint(iparr[2], 10, 8)
    ip2 := uint8(ip264)
    ip364, _ := strconv.ParseUint(iparr[3], 10, 8)
    ip3 := uint8(ip364)

    buf.WriteByte(ip0)
    buf.WriteByte(ip1)
    buf.WriteByte(ip2)
    buf.WriteByte(ip3)

    for _, n := range m.n {
        iparr := strings.Split(n.nIP, ".")

        ip064, _ := strconv.ParseUint(iparr[0], 10, 8)
        ip0 := uint8(ip064)
        ip164, _ := strconv.ParseUint(iparr[1], 10, 8)
        ip1 := uint8(ip164)
        ip264, _ := strconv.ParseUint(iparr[2], 10, 8)
        ip2 := uint8(ip264)
        ip364, _ := strconv.ParseUint(iparr[3], 10, 8)
        ip3 := uint8(ip364)

        buf.WriteByte(ip0)
        buf.WriteByte(ip1)
        buf.WriteByte(ip2)
        buf.WriteByte(ip3)

        binary.Write(buf, binary.BigEndian, n.nPort)
        buf.WriteByte(0)
        buf.WriteByte(0)
        binary.Write(buf, binary.BigEndian, n.nID)
        binary.Write(buf, binary.BigEndian, n.nCost)
    }
    return buf.Bytes(), nil
}
