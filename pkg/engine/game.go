//go:build !test
// +build !test

package engine

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game represents the main game instance with the ECS world and game loop.
type Game struct {
	World          *World
	lastUpdateTime time.Time
	ScreenWidth    int
	ScreenHeight   int
	Paused         bool
}

// NewGame creates a new game instance.
func NewGame(screenWidth, screenHeight int) *Game {
	return &Game{
		World:          NewWorld(),
		lastUpdateTime: time.Now(),
		ScreenWidth:    screenWidth,
		ScreenHeight:   screenHeight,
	}
}

// Update implements ebiten.Game interface. Called every frame.
func (g *Game) Update() error {
	if g.Paused {
		return nil
	}

	// Calculate delta time
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Cap delta time to prevent spiral of death
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	// Update the world
	g.World.Update(deltaTime)

	return nil
}

// Draw implements ebiten.Game interface. Called every frame.
func (g *Game) Draw(screen *ebiten.Image) {
	// Drawing is handled by rendering systems that are part of the World
	// Systems can access the screen through a component or global state
}

// Layout implements ebiten.Game interface. Returns the game's screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ScreenWidth, g.ScreenHeight
}

// Run starts the game loop.
func (g *Game) Run(title string) error {
	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("failed to run game: %w", err)
	}

	return nil
}
