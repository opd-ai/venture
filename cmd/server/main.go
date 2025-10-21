package main

import (
	"flag"
	"log"
)

var (
	port = flag.String("port", "8080", "Server port")
	maxPlayers = flag.Int("max-players", 4, "Maximum number of players")
)

func main() {
	flag.Parse()

	log.Printf("Starting Venture Game Server")
	log.Printf("Port: %s, Max Players: %d", *port, *maxPlayers)

	// TODO: Initialize server
	// - Create game world
	// - Start network listener
	// - Handle client connections
	// - Run authoritative game loop

	log.Println("Server initialized successfully")
	log.Printf("Server running on port %s", *port)

	// Keep server running
	select {}
}
