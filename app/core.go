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
		commands: commands,
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
	case a.commands["1"]:
		// If we have more than 2 input args, that means they want to see
		// the command information for multiple commands, so let's loop
		// through each of them
		if numArgs > 2 {
			for i := 1; i < numArgs; i++ {
				a.help(inputArgs[i])
			}
			return nil
		}

		// If we only have 2, then the user just wants to see this one command
		if numArgs == 2 {
			a.help(inputArgs[1])
			return nil
		}

		// We only had help in our input, so they want the full list
		a.help("")
		return nil
	case "2":
		fallthrough
	case a.commands["2"]:
		// Do we have the proper number of arguments?
		if numArgs < 4 {
			return UPDERR
		}

		// Let's get the first ID# given
		id1, err := strconv.ParseInt(inputArgs[1], 10, 32)
		if err != nil {
			e := fmt.Sprintf("%s ERROR: error parsing input id1: %v\n", a.commands["2"], err)
			return errors.New(e)
		}

		// Let's get the second ID# given
		id2, err := strconv.ParseInt(inputArgs[2], 10, 32)
		if err != nil {
			e := fmt.Sprintf("%s ERROR: error parsing input id2: %v\n", a.commands["2"], err)
			return errors.New(e)
		}

		// Let's check if the link cost given was "inf"
		if inputArgs[3] == "inf" {
			inputArgs[3] = "-1"
		}

		// Parse the link cost into an int
		cost, err := strconv.Atoi(inputArgs[3])
		if err != nil {
			e := fmt.Sprintf("%s ERROR: error parsing input cost: %v\n", a.commands["2"], err)
			return errors.New(e)
		}

		// Perform the update
		if err := a.server.Update(uint16(id1), uint16(id2), cost); err != nil {
			e := fmt.Sprintf("%s ERROR: %v\n", a.commands["2"], err)
			return errors.New(e)
		}
		a.Out("%s SUCCESS\n", a.commands["2"])
		return nil
	case "3":
		fallthrough
	case a.commands["3"]:
		// Call the servers step function and check for any errors
		if err := a.server.Step(); err != nil {
			e := fmt.Sprintf("%s ERROR: %v\n", a.commands["3"], err)
			return errors.New(e)
		}
		a.Out("%s SUCCESS\n", a.commands["3"])
		return nil
	case "4":
		fallthrough
	case a.commands["4"]:
		// Call the servers packets function and check for any errors
		if err := a.server.Packets(); err != nil {
			e := fmt.Sprintf("%s ERROR: %v\n", a.commands["4"], err)
			return errors.New(e)
		}
		a.Out("%s SUCCESS\n", a.commands["4"])
		return nil
	case "5":
		fallthrough
	case a.commands["5"]:
		// Call the servers routing table function and check for any errors
		if err := a.server.Display(); err != nil {
			e := fmt.Sprintf("%s ERROR: %v\n", a.commands["5"], err)
			return errors.New(e)
		}
		a.Out("%s SUCCESS\n", a.commands["5"])
		return nil
	case "6":
		fallthrough
	case a.commands["6"]:
		// Do we have the proper number of arguments?
		if numArgs != 2 {
			return DISERR
		}

		// Yes, so let's get the connection that we need to terminate
		// and attempt to disable it
		id, _ := strconv.ParseInt(inputArgs[1], 10, 64)

		// Call the servers disable function and check for any errors
		if err := a.server.Disable(uint16(id)); err != nil {
			e := fmt.Sprintf("%s ERROR: %v\n", a.commands["6"], err)
			return errors.New(e)
		}
		a.Out("%s SUCCESS\n", a.commands["6"])
		return nil
	case "7":
		fallthrough
	case a.commands["7"]:
		// Call the servers crash function and print the success message
		a.server.Crash()
		a.Out("%s SUCCESS\n", a.commands["7"])
		return nil
	default:
		// We didn't find a matching command for their input, let's throw an error
		return INPERR
	}
	return nil
}

// help prints the application commands
func (a *Application) help(command string) {
	a.Out("\nApplication Commands\n")
	a.Out("--------------------\n")

	switch command {
	case "":
		cmds := `1. help - Displays available application commands
2. update <server-ID1> <server-ID2> <new-link-cost> - Updates the link cost between the two servers
3. step - Triggers the server to send the routing update right away
4. packets - Displays the number of DVR packets this server has received since the last time this command was used
5. display - Displays the current routing table, with the servers sorted in ascending order
6. disable <server-ID> - Disables the link between to a given server
7. crash - "Closes" all connections, to simulate a server crash
`
		a.Out(cmds)
		break
	case "1":
		fallthrough
	case "help":
		a.Out("1. help - Displays available application commands\n")
		break
	case "2":
		fallthrough
	case "update":
		a.Out("2. update <server-ID1> <server-ID2> <new-link-cost> - Updates the link cost between the two servers\n")
		break
	case "3":
		fallthrough
	case "step":
		a.Out("3. step - Triggers the server to send the routing update right away\n")
		break
	case "4":
		fallthrough
	case "packets":
		a.Out("4. packets - Displays the number of DVR packets this server has received since the last time this command was used\n")
		break
	case "5":
		fallthrough
	case "display":
		a.Out("5. display - Displays the current routing table, with the servers sorted in ascending order\n")
		break
	case "6":
		fallthrough
	case "disable":
		a.Out("6. disable <server-ID> - Disables the link between to a given server\n")
		break
	case "7":
		fallthrough
	case "crash":
		a.Out("7. crash - 'Closes' all connections, to simulate a server crash\n")
		break
	default:
		return
	}
}
