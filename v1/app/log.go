package app

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Out prints message to the standard output device
func (a *Application) Out(format string, b ...interface{}) {
	// Create a green colored message using the given arguments
	msg := color.HiGreenString(format, b...)
	// Print the message to the user
	fmt.Fprintf(os.Stdout, msg)
}

// OutCyan prints message to the standard output device in cyan
func (a *Application) OutCyan(format string, b ...interface{}) {
	// Create a green colored message using the given arguments
	msg := color.HiCyanString(format, b...)
	// Print the message to the user
	fmt.Fprintf(os.Stdout, msg)
}

// OutErr prints message to the standard output error device
func (a *Application) OutErr(format string, b ...interface{}) {
	// Create a green colored message using the given arguments
	msg := color.HiRedString(format, b...)
	// Print the message to the user
	fmt.Fprintf(os.Stdout, msg)
}

// startupText prints the text that should be displayed on application startup
func (a *Application) startupText() {
	sText := `
DVR: Distance Vector Routing Protocol
--------------------------------------
A simplified version of the distance vector routing protocol.
By Sabra Bilodeau

Available commands:
    1. help
    2. update <server-ID1> <server-ID2> <new-link-cost>
    3. step
    4. packets
    5. display
    6. disable <server-id>
    7. crash

You may either type the command name, i.e. 'disable <server-id>', or the command number, i.e. '2 <server-id>'
Type 'help' for an explanation of each command or type 'help <command>' to get the explanation for a specific command.

Now beginning topology setup..
`
	a.Out("%s", sText)
}
