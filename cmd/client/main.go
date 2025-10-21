package main

import (
	"flag"
	"log"

	"github.com/opd-ai/venture/pkg/engine"
)

var (
	width  = flag.Int("width", 800, "Screen width")
	height = flag.Int("height", 600, "Screen height")
	seed   = flag.Int64("seed", 12345, "World generation seed")
)

func main() {
	flag.Parse()

	log.Printf("Starting Venture - Procedural Action RPG")
	log.Printf("Screen: %dx%d, Seed: %d", *width, *height, *seed)

	// Create the game instance
	game := engine.NewGame(*width, *height)

	// TODO: Initialize game systems here
	// - Add rendering systems
	// - Add gameplay systems
	// - Generate initial world
	// - Create player entity

	log.Println("Game initialized successfully")

	// Run the game loop
	if err := game.Run("Venture - Procedural Action RPG"); err != nil {
		log.Fatalf("Error running game: %v", err)
	}
}
