package types

// This package is required to avoid import cycles in Go
//
// It defines the three types of interfaces we have in our progam
// and the functions that are required to be in each of them

type Application interface {
    Out(format string, a ...interface{})
    OutErr(format string, a ...interface{})
    WaitForInput() error
}

type Client interface {
    NewClient(address string) (Client, error)
    Update()
    Close() error
}

type Server interface {
    Update(id1, id2 uint16, newCost int) error
    Step() error
    Packets() error
    Display() error
    Disable(id uint16) error
    Crash()
}
