package mobile

import (
	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

// Game is the mobile game instance
var (
	gameInstance *engine.EbitenGame
	logger       *logrus.Logger
)

func init() {
	// Initialize logger for mobile
	logger = logging.TestUtilityLogger("mobile")
}

// Init initializes the game for mobile platforms.
// This must be called before any other functions.
func Init() {
	if gameInstance != nil {
		return // Already initialized
	}

	// Create the game instance with mobile-friendly dimensions
	// Portrait mode: 720x1280 (9:16 aspect ratio)
	gameInstance = engine.NewEbitenGameWithLogger(720, 1280, logger)

	// Log initialization
	logger.Info("mobile game initialized")

	// Register the game with ebitenmobile
	mobile.SetGame(gameInstance)
}

// Start starts the game loop.
// This is called automatically by the mobile platform.
func Start() {
	if gameInstance == nil {
		Init()
	}
}

// Update updates the game state.
// Returns true to continue running, false to quit.
func Update() bool {
	return gameInstance != nil
}

// GetScreenWidth returns the screen width.
func GetScreenWidth() int {
	if gameInstance == nil {
		return 0
	}
	return 720
}

// GetScreenHeight returns the screen height.
func GetScreenHeight() int {
	if gameInstance == nil {
		return 0
	}
	return 1280
}
