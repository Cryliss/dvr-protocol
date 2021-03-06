package main

import (
	"dvr-protocol/app"
	"dvr-protocol/server"
	"errors"
	"flag"
	"fmt"
	"os"
)

// usage prints information on how to use the program and then exits
func usage() {
	fmt.Printf("usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(-1)
}

func main() {
	var file string
	var interval int

	// Lets load our flags.
	flag.StringVar(&file, "t", "", "Topology file name.")
	flag.IntVar(&interval, "i", -1, "Routing update interval, in seconds.")
	flag.Parse()

	// Did we get a file name or interval to update?
	if interval == -1 || file == "" {
		usage()
	}

	// Check that the file path actually exists
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("The provided file path does not exist!")
		os.Exit(-1)
	}

	// Create a new server
	server := server.New(file, interval)

	// Create a new application
	app := app.New(server)

	// Pass the app to the server
	server.SetApplication(app)

	// Initialize the servers routing table
	server.InitializeRoutingTable()

	// Create a goroutine to start listening for new packets
	go server.Listen()

	// Create a goroutine for sending updates at the specified interval
	go server.Loopy()

	// Print the current topology setup
	app.OutCyan("\nTOPOLOGY\n")
	app.OutCyan("========\n")
	server.Display()
	app.OutCyan("======================\n")
	app.OutCyan("Starting the DVR protocol ..\n\nNow accepting user input.\n")

	// Begin waiting for user input
	for {
		err := app.WaitForInput()
		if err != nil {
			app.OutErr("ERROR: %v\nExiting application now\n", err)
			os.Exit(-1)
		}
	}
}
