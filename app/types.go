package app

import (
	"dvr-protocol/types"
	"errors"
)

// Create some common errors for input mistakes we may see
var UPDERR error = errors.New("update input error: You must provide 2 server ID#s and the new link cost, which can be either a number or 'inf'")
var DISERR error = errors.New("disable input error: You must give the server id you wish to disable\nType `display` to view the current routing table")
var INPERR error = errors.New("invalid input error: You must give one of the accepted app commands\nType 'help' to get a list of available commands")

// Commands the user can give the application
// We are using a map[string]string here, just in case
// the user wants to be lazy and not type the whole thing out
var commands = map[string]string{
	"1": "help",
	"2": "update",
	"3": "step",
	"4": "packets",
	"5": "display",
	"6": "disable",
	"7": "crash",
}

// Application structure for our application that holds the availabe commands
// and our host server
type Application struct {
	// The applicaton's server
	server types.Server

	// Map of the application commands, see above
	commands map[string]string
}
