package main

import (
    "dvr-protocol/app"
    "dvr-protocol/server"
    "errors"
    "flag"
    "fmt"
    "os"
)

// func usage {{{
//
// Prints information on how to use the program and then exits
func usage() {
    fmt.Printf("usage: %s\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(-1)
} // }}}


// func main {{{
//
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
    app := app.New(server)
    server.SetApplication(app)

    t := server.Topology()
    server.InitalizeRt()
    server.Updates()
    fmt.Println("\nTOPOLOGY\n")
    fmt.Println("========\n")
    fmt.Printf("Num Servers: %d\n", t.NumServers)
    fmt.Printf("Num Neighbors: %d\n\n", t.NumNeighbors)
    fmt.Println("--------------\n")

    fmt.Printf("Server #%d: %v\n\n", server.Id, server.Bindy)

    for _, val := range t.Neighbors {
        if val.Cost == -1 || val.Cost == 0 {
            continue
        }
        fmt.Printf("Neighbor #%d: \nAddr: %v\nCost: %d\n\n", val.Id, val.Bindy, val.Cost)
    }
    fmt.Println("======================\n")
    fmt.Printf("Starting the DVR protocol .. Now accepting user input.\n")

    for {
        err := app.WaitForInput()
        if err != nil {
            app.OutErr("ERROR: %v\nExiting application now\n", err)
            os.Exit(-1)
        }
    }
}  // }}}
