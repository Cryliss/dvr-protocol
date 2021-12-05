package app

import (
	"bufio"
	"dvr-protocol/types"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// New returns a new app
func New(s types.Server) *Application {
	a := Application{
		Commands: commands,
		server:   s,
	}

	// Print out the application startup message
	a.startupText()
	return &a
}

// WaitForInput waits for user input and handles parsing it once received.
func (a *Application) WaitForInput() error {
	// Create a new bufio reader to read user input from the command line
	reader := bufio.NewReader(os.Stdin)

	// Variable to store our scanned input into
	var userInput string

	for {
		// Prompt the user for a command
		a.Out("\nPlease enter a command: ")

		// Read user input and save into userInput variable
		userInput, _ = reader.ReadString('\n')
		userInput = strings.Replace(userInput, "\n", "", -1)

		// Parse and handle the users input
		// If the request resulted in an error, let's let the user know
		err := a.parseInput(userInput)
		if err != nil {
			a.OutErr("%v\n", err)
		}
	}
	return nil
}

// parseInput parses the users input and calls the function associated with
// the given command
func (a *Application) parseInput(userInput string) error {
	// Split the users input into array of strings
	inputArgs := strings.SplitN(userInput, " ", 4)
	numArgs := len(inputArgs)

	// Check what the command was, the first item in the input, and
	// perform the actions necessary for that command
	switch inputArgs[0] {
	case "1":
		fallthrough
	case a.Commands["1"]:
		return a.help(inputArgs)
	case "2":
		fallthrough
	case a.Commands["2"]:
		// Do we have the proper number of arguments?
		if numArgs < 4 {
			return ErrUpd
		}
		return a.update(inputArgs)
	case "3":
		fallthrough
	case a.Commands["3"]:
		return a.step()
	case "4":
		fallthrough
	case a.Commands["4"]:
		return a.packets()
	case "5":
		fallthrough
	case a.Commands["5"]:
		return a.display()
	case "6":
		fallthrough
	case a.Commands["6"]:
		// Do we have the proper number of arguments?
		if numArgs != 2 {
			return ErrDis
		}
		return a.disable(inputArgs[1])
	case "7":
		fallthrough
	case a.Commands["7"]:
		return a.crash()
	default:
		// We didn't find a matching command for their input, let's throw an error
		return ErrInp
	}
}

// help prints the requested help commands
func (a *Application) help(inputArgs []string) error {
	numArgs := len(inputArgs)

	// If we have more than 2 input args, that means they want to see
	// the command information for multiple commands, so let's loop
	// through each of them
	if numArgs > 2 {
		for i := 1; i < numArgs; i++ {
			a.printHelp(inputArgs[i])
		}
		return nil
	}

	// If we only have 2, then the user just wants to see this one command
	if numArgs == 2 {
		a.printHelp(inputArgs[1])
		return nil
	}

	// We only had printHelp in our input, so they want the full list
	a.printHelp("")
	return nil
}

// printHelp prints the application commands
func (a *Application) printHelp(command string) {
	a.Out("\nApplication Commands\n")
	a.Out("--------------------\n")

	if command == "" {
		for _, cmd := range helpText {
			a.Out(cmd)
		}
		return
	}

	if len(command) == 1 {
		if cmd, ok := helpText[commands[command]]; ok {
			a.Out(cmd)
		}
		return
	}

	if cmd, ok := helpText[command]; ok {
		a.Out(cmd)
	}
}

// update updates the link cost between two servers
func (a *Application) update(inputArgs []string) error {
	// Let's get the first ID# given
	id1, err := strconv.ParseInt(inputArgs[1], 10, 32)
	if err != nil {
		e := fmt.Sprintf("%s ERROR: error parsing input id1: %v\n", a.Commands["2"], err)
		return errors.New(e)
	}

	// Let's get the second ID# given
	id2, err := strconv.ParseInt(inputArgs[2], 10, 32)
	if err != nil {
		e := fmt.Sprintf("%s ERROR: error parsing input id2: %v\n", a.Commands["2"], err)
		return errors.New(e)
	}

	// Let's check if the link cost given was "inf"
	if inputArgs[3] == "inf" {
		inputArgs[3] = "-1"
	}

	// Parse the link cost into an int
	cost, err := strconv.Atoi(inputArgs[3])
	if err != nil {
		e := fmt.Sprintf("%s ERROR: error parsing input cost: %v\n", a.Commands["2"], err)
		return errors.New(e)
	}

	// Perform the update
	if err := a.server.Update(uint16(id1), uint16(id2), cost); err != nil {
		e := fmt.Sprintf("%s ERROR: %v\n", a.Commands["2"], err)
		return errors.New(e)
	}
	a.Out("%s SUCCESS\n", a.Commands["2"])
	return nil
}

// step triggers the server to send a routing update automatically
func (a *Application) step() error {
	// Call the servers step function and check for any errors
	if err := a.server.Step(); err != nil {
		e := fmt.Sprintf("%s ERROR: %v\n", a.Commands["3"], err)
		return errors.New(e)
	}
	a.Out("%s SUCCESS\n", a.Commands["3"])
	return nil
}

// packets calls the server to display the # of packets received
func (a *Application) packets() error {
	// Call the servers packets function and check for any errors
	if err := a.server.Packets(); err != nil {
		e := fmt.Sprintf("%s ERROR: %v\n", a.Commands["4"], err)
		return errors.New(e)
	}
	a.Out("%s SUCCESS\n", a.Commands["4"])
	return nil
}

// display calls the server to display the current routing table
func (a *Application) display() error {
	// Call the servers routing table function and check for any errors
	if err := a.server.Display(); err != nil {
		e := fmt.Sprintf("%s ERROR: %v\n", a.Commands["5"], err)
		return errors.New(e)
	}
	a.Out("%s SUCCESS\n", a.Commands["5"])
	return nil
}

// disable calls the server to disable a server
func (a *Application) disable(idString string) error {
	// Yes, so let's get the connection that we need to terminate
	// and attempt to disable it
	id, _ := strconv.ParseInt(idString, 10, 64)

	// Call the servers disable function and check for any errors
	if err := a.server.Disable(uint16(id)); err != nil {
		e := fmt.Sprintf("%s ERROR: %v\n", a.Commands["6"], err)
		return errors.New(e)
	}
	a.Out("%s SUCCESS\n", a.Commands["6"])
	return nil
}

// crash calls the server to crash it
func (a *Application) crash() error {
	// Call the servers crash function and print the success message
	a.server.Crash()
	a.Out("%s SUCCESS\n", a.Commands["7"])
	return nil
}
