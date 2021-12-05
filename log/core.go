// Package log provides terminal logging to the user
package log

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

// Logger for application logging to the terminal
type Logger struct {
	Log     zerolog.Logger
	Debug 	bool
}

// New initializes and returns a new Logger
func New() *Logger {
	// Set the time logging format
	zerolog.TimeFieldFormat = time.RFC3339
	l := Logger{
		Log: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
	return &l
}

// OutApp provides terminal logging for application related output
func (l Logger) OutApp(format string, b... interface{}) {
	// Create a green colored message using the given arguments
	msg := color.HiGreenString(format, b...)
	// Print the message to the user
	fmt.Fprintf(os.Stdout, msg)
}

// OutError provides terminal logging for errorr related output
func (l Logger) OutError(format string, b... interface{}) {
	// Create a red colored message using the given arguments
	msg := color.HiRedString(format, b...)
	// Print the message to the user
	fmt.Fprintf(os.Stdout, msg)
}

// OutServer provides terminal logging for server related output
func (l Logger) OutServer(format string, b... interface{}) {
	// Create a cyan colored message using the given arguments
	msg := color.HiCyanString(format, b...)
	// Print the message to the user
	fmt.Fprintf(os.Stdout, msg)
}

// OutDebug provides terminal logging for debugging related output
func (l Logger) OutDebug(format string, b... interface{}) {
	if l.Debug {
		// Create a magenta colored message using the given arguments
		msg := color.HiMagentaString(format, b...)
		// Print the message to the user
		fmt.Fprintf(os.Stdout, msg)
	}
}
