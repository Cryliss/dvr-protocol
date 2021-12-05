package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Code assistance: https://github.com/cirocosta/rawdns

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

// Message to store the unmarshaled message from the other host server
// The minimum possible size of an update message:
// 8 bytes for the host servers information
// A minimum 10 bytes per neighbor for their information
//          + 2 bytes for the null byte
// With a minimum of 1 expected neighbors, the minimum size of the message
// is expected to be 8 + 12 = 20 bytes
type Message struct {
	Updates uint16                // Number of expected fields
	Port    uint16                // Port of the host server sending the msg
	IP      string                // IP of the host server sending the msg
	N       map[uint16]*Neighbor  // Map of the neighbor information in the hosts routng table
}

// Neighbor To store the information about the host servers neighbors
// Total size of each neighbors info is 12 bytes with null byte
type Neighbor struct {
	IP   string // Neighbor IP
	Port uint16 // Neighbor Port
	ID   uint16 // Neighbor ID
	Cost uint16 // Neighbor Link cost
}

// UnmarshalMessage unmarshals an update message into a message struct
//
// Unmarshaled message should like like:
//  {nUpdates:3 hPort:2001 hIP:192.168.200.80 n:map[1:0xc00000e318 2:0xc00000e348 3:0xc00000e378]}
//
// With the map containing
//  {nIP:192.168.200.80 nPort:2000 nID:1 nCost:7}
//  {nIP:192.168.200.80 nPort:2001 nID:2 nCost:0}
//  {nIP:192.168.200.80 nPort:2002 nID:3 nCost:2}
func UnmarshalMessage(msg []byte, m *Message) error {
	// Create a new map for the unmarshalled neighbors
	var neighbors map[uint16]*Neighbor
	neighbors = make(map[uint16]*Neighbor, 4)

	// Did we actually get a message struct?
	if m == nil {
		return errors.Errorf("header must be non-nil")
	}

	// Check if the message is of the correct length
	if len(msg) < 20 {
		return errors.Errorf("msg does not have the expected size - %d", len(msg))
	}

	// The number of expected updates is formed by the first two bytes such that
	// it results in an uint16.
	//
	// As this comes from the network in UDP packets we can assume that it comes
	// in BigEndian (network byte order).
	m.Updates = uint16(msg[1]) | uint16(msg[0])<<8
	m.Port = uint16(msg[3]) | uint16(msg[2])<<8
	m.IP = fmt.Sprintf("%d.%d.%d.%d\n", msg[4], msg[5], msg[6], msg[7])

	// Loop through the rest of the bytes in the messae and unmarshal each
	// set of 12 bytes into a new message struct into a new neighbor struct
	for b := 8; b <= len(msg)-12; b += 12 {
		var n Neighbor

		n.IP = fmt.Sprintf("%d.%d.%d.%d\n", msg[b], msg[b+1], msg[b+2], msg[b+3])
		n.Port = uint16(msg[b+5]) | uint16(msg[b+4])<<8
		n.ID = uint16(msg[b+9]) | uint16(msg[b+8])<<8
		n.Cost = uint16(msg[b+11]) | uint16(msg[b+10])<<8

		neighbors[n.ID] = &n
	}

	// Set the message neighbors map to be the neighbors map we just initialized
	m.N = neighbors
	return nil
}

// Marshal marshal's the update message into an array of bytes
func (m *Message) Marshal() ([]byte, error) {
	// Create a new buffer to write the message to
	buf := new(bytes.Buffer)

	// Write the number of updates and the host port number
	// into the buffer, encoded as binary using Big Endian
	binary.Write(buf, binary.BigEndian, m.Updates)
	binary.Write(buf, binary.BigEndian, m.Port)

	// Split up the IP address
	iparr := strings.Split(m.IP, ".")

	// Create new 8 bit unsigned integer for each part of the IP address
	ip064, _ := strconv.ParseUint(iparr[0], 10, 8)
	ip0 := uint8(ip064)
	ip164, _ := strconv.ParseUint(iparr[1], 10, 8)
	ip1 := uint8(ip164)
	ip264, _ := strconv.ParseUint(iparr[2], 10, 8)
	ip2 := uint8(ip264)
	ip364, _ := strconv.ParseUint(iparr[3], 10, 8)
	ip3 := uint8(ip364)

	// Write each part of the IP address to the buffer
	buf.WriteByte(ip0)
	buf.WriteByte(ip1)
	buf.WriteByte(ip2)
	buf.WriteByte(ip3)

	// For each neighbor in our message -
	for _, n := range m.N {
		// Split up the neighbors IP address
		iparr := strings.Split(n.IP, ".")

		// Create new 8 bit unsigned integer for each part of the IP address
		ip064, _ := strconv.ParseUint(iparr[0], 10, 8)
		ip0 := uint8(ip064)
		ip164, _ := strconv.ParseUint(iparr[1], 10, 8)
		ip1 := uint8(ip164)
		ip264, _ := strconv.ParseUint(iparr[2], 10, 8)
		ip2 := uint8(ip264)
		ip364, _ := strconv.ParseUint(iparr[3], 10, 8)
		ip3 := uint8(ip364)

		// Write each part of the IP address to the buffer
		buf.WriteByte(ip0)
		buf.WriteByte(ip1)
		buf.WriteByte(ip2)
		buf.WriteByte(ip3)

		// Write the neighbors port number into the buffer,
		// encoded as binary using Big Endian
		binary.Write(buf, binary.BigEndian, n.Port)

		// Write two null bytes
		buf.WriteByte(0)
		buf.WriteByte(0)

		// Write the neighbors id number and link cost into the buffer,
		// encoded as binary using Big Endian
		binary.Write(buf, binary.BigEndian, n.ID)
		binary.Write(buf, binary.BigEndian, n.Cost)
	}

	// Return the bytes writen to the buffer
	return buf.Bytes(), nil
}
