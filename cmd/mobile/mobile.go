package mobile

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/opd-ai/venture/pkg/engine"
)

// Game is the mobile game instance
var gameInstance *engine.Game

// Init initializes the game for mobile platforms.
// This is called automatically by ebitenmobile.
func init() {
	// Create the game instance with mobile-friendly dimensions
	// Portrait mode: 720x1280 (9:16 aspect ratio)
	gameInstance = engine.NewGame(720, 1280)

	// Log initialization
	log.Println("Mobile game initialized")

	// Register the game with ebitenmobile
	mobile.SetGame(gameInstance)
}
