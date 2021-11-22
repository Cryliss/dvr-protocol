package types

// This package is required to avoid import cycles in Go
//
// It defines the three types of interfaces we have in our progam
// and the functions that are required to be in each of them

// Application interface for outputting to the user
type Application interface {
	Out(format string, a ...interface{})
	OutCyan(format string, a ...interface{})
	OutErr(format string, a ...interface{})
	WaitForInput() error
}

// Client interface for the client connections used in the updates
type Client interface {
	NewClient(address string) (Client, error)
	Update()
	Close() error
}

// Server interface for the server comands
type Server interface {
	Update(id1, id2 uint16, newCost int) error
	Step() error
	Packets() error
	Display() error
	Disable(id uint16) error
	Crash()
}
