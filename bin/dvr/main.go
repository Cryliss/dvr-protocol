package main

import (
    "dvr/app"
    "dvr/network"
    "dvr/topology"
    "errors"
    "flag"
    "fmt"
    "os"
    "time"
)

var file string
var interval int
var debug bool

// usage prints information on how to use the program and then exits
func usage() {
    fmt.Printf("usage: %s\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(-1)
}

// checkFlags checks the command line flags given at startup
func checkFlags() {
    // Lets load our flags.
    flag.StringVar(&file, "t", "", "Topology file name.")
    flag.IntVar(&interval, "i", -1, "Routing update interval, in seconds.")
    flag.BoolVar(&debug, "d", false, "Whether or not to show routing tables for debugging.")
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
}

func main() {
    checkFlags()

    a := app.New()
    a.Log.Debug = debug
    a.StartupText()

    //a.Log.OutDebug("Parsing topology file .. \n")

    top, serverID, err := topology.ParseTopology(file)
    if err != nil {
        fmt.Printf("Failed to parse topology file - %s\n", err.Error())
        os.Exit(-1)
    }
    //a.Log.OutDebug("Successfully parsed topology file.\nStarting network setup now ..\n")
    a.Server = network.New(top, serverID, a.Log)

    go a.Server.Listen()
    go a.Server.Loopy(interval)

    // Print the current topology setup
    a.Log.OutServer("\nTOPOLOGY\n")
    a.Log.OutServer("========\n")
    a.Server.Display()

    // Begin waiting for user input
    for {
        err := a.WaitForInput()
        if err != nil {
            time.Sleep(5*time.Millisecond)
            a.Log.OutApp("\nExiting application now\n")
            os.Exit(0)
        }
    }
}
