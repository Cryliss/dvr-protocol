package app

import (
    "dvr/log"
    "dvr/network"
    "dvr/server"
)

// Application for all things related to our application.
type Application struct {
    Commands map[string]string
    Log *log.Logger
    Network *network.Network
    Server *server.Server
}
