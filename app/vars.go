package app

import "errors"

// Create some common errors for input mistakes we may see

// ErrUpd is an error message for our update command
var ErrUpd error = errors.New("update ERROR: You must provide 2 server ID#s and the new link cost, which can be either a number or 'inf'")

// ErrDis is an error message for our disable command
var ErrDis error = errors.New("disable ERROR: You must give the server id you wish to disable\nType `display` to view the current routing table")

// ErrInp is an error message for invalid user input
var ErrInp error = errors.New("invalid ERROR: You must give one of the accepted app commands\nType 'help' to get a list of available commands")

// ExitErr is the error to signal we want to exit the wait for input for loop
var ExitErr error = errors.New("exiting application")

// commands the user can give the application
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
	"8": "exit",
}

// The helpText to display for each command
var helpText = map[string]string{
	"help": "1. help - Displays available application commands\n",
	"update": "2. update <server-ID1> <server-ID2> <new-link-cost> - Updates the link cost between the two servers\n",
	"step": "3. step - Triggers the server to send the routing update right away\n",
	"packets": "4. packets - Displays the number of DVR packets this server has received since the last time this command was used\n",
	"display": "5. display - Displays the current routing table, with the servers sorted in ascending order\n",
	"disable": "6. disable <server-ID> - Disables the link between to a given server\n",
	"crash": "7. crash - 'Closes' all connections, to simulate a server crash\n",
	"exit": "8. exit - Exits the aplication.",
}
