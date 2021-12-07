// Package app handles performing the application commands
package app

import (
    "bufio"
    "dvr/log"
	"os"
	"strconv"
	"strings"

    "github.com/pkg/errors"
)

// New initializes and returns a new Application.
func New() *Application {
    a := Application{
        Commands: commands,
        Log: log.New(),
    }
    return &a
}

// StartupText prints the text that should be displayed on application startup
func (a *Application) StartupText() {
    sText := `
DVR: Distance Vector Routing Server
--------------------------------------
Available commands:
    1. help
    2. update <server-ID1> <server-ID2> <new-link-cost>
    3. step
    4. packets
    5. display
    6. disable <server-id>
    7. crash

Type 'help' or 'help <command>' to get the explanation
for the commands.
`
    a.Log.OutApp("%s", sText)
}


// WaitForInput waits for user input and handles parsing it once received.
func (a *Application) WaitForInput() error {
    // Create a new bufio reader to read user input from the command line
    reader := bufio.NewReader(os.Stdin)

    // Variable to store our scanned input into
    var userInput string

    for {
        // Prompt the user for a command
        a.Log.OutApp("\nPlease enter a command: ")

        // Read user input and save into userInput variable
        userInput, _ = reader.ReadString('\n')
        userInput = strings.Replace(userInput, "\n", "", -1)

        // Parse and handle the users input
        // If the request resulted in an error, let's let the user know
        err := a.parseInput(userInput)
        if err != nil {
            if err == ExitErr {
                return err
            }
            a.Log.OutError("%v\n", err)
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
    command := strings.ToLower(inputArgs[0])
    switch command {
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
    case "8":
        fallthrough
    case a.Commands["7"]:
        a.Log.OutApp("Shutting down server .. \n")
        a.Server.Crash()
        return ExitErr
    default:
        // We didn't find a matching command for their input, let's throw an error
        return ErrInp
    }
}

// help prints the requested help commands
func (a *Application) help(inputArgs []string) error {
    command := strings.ToUpper(a.Commands["1"])
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
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}

// printHelp prints the application commands
func (a *Application) printHelp(command string) {
    a.Log.OutApp("\nApplication Commands\n")
    a.Log.OutApp("--------------------\n")

    if command == "" {
        for _, cmd := range helpText {
            a.Log.OutApp(cmd)
        }
        return
    }

    if len(command) == 1 {
        if cmd, ok := helpText[commands[command]]; ok {
            a.Log.OutApp(cmd)
        }
        return
    }

    if cmd, ok := helpText[command]; ok {
        a.Log.OutApp(cmd)
    }
}

// update updates the link cost between two servers
func (a *Application) update(inputArgs []string) error {
    command := strings.ToUpper(a.Commands["2"])

    // Let's get the first ID# given
    id1, err := strconv.ParseInt(inputArgs[1], 10, 32)
    if err != nil {
        return errors.Wrapf(err, "%s ERROR: error parsing input id1: %v\n", command, err)
    }

    // Let's get the second ID# given
    id2, err := strconv.ParseInt(inputArgs[2], 10, 32)
    if err != nil {
        return errors.Wrapf(err, "%s ERROR: error parsing input id2: %v\n", command, err)
    }

    // Let's check if the link cost given was "inf"
    if inputArgs[3] == "inf" {
        inputArgs[3] = "99999"
    }

    // Parse the link cost into an int
    cost, err := strconv.Atoi(inputArgs[3])
    if err != nil {
        return errors.Wrapf(err, "%s ERROR: error parsing input cost: %v\n", command, err)
    }

    // Perform the update
    if err := a.Server.Update(uint16(id1), uint16(id2), cost); err != nil {
        return errors.Wrapf(err, "%s ERROR: %v\n", command, err)
    }
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}

// step triggers the server to send a routing update automatically
func (a *Application) step() error {
    command := strings.ToUpper(a.Commands["3"])
    // Call the servers step function and check for any errors
    if err := a.Server.Step(); err != nil {
        return errors.Wrapf(err, "%s ERROR: %v\n", command, err)
    }
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}

// packets calls the server to display the # of packets received
func (a *Application) packets() error {
    command := strings.ToUpper(a.Commands["4"])
    // Call the servers packets function and check for any errors
    if err := a.Server.Packets(); err != nil {
        return errors.Wrapf(err, "%s ERROR: %v\n", command, err)
    }
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}

// display calls the server to display the current routing table
func (a *Application) display() error {
    command := strings.ToUpper(a.Commands["5"])
    // Call the servers routing table function and check for any errors
    if err := a.Server.Display(); err != nil {
        return errors.Wrapf(err, "%s ERROR: %v\n", command, err)
    }
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}

// disable calls the server to disable a server
func (a *Application) disable(idString string) error {
    command := strings.ToUpper(a.Commands["6"])
    // Yes, so let's get the connection that we need to terminate
    // and attempt to disable it
    id, _ := strconv.Atoi(idString)

    // Call the servers disable function and check for any errors
    if err := a.Server.Disable(uint16(id)); err != nil {
        return errors.Wrapf(err, "%s ERROR: %v\n", command, err)
    }
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}

// crash calls the server to crash it
func (a *Application) crash() error {
    command := strings.ToUpper(a.Commands["7"])
    // Call the servers crash function and print the success message
    a.Server.Crash()
    a.Log.OutApp("\n%s SUCCESS\n", command)
    return nil
}
