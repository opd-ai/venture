//go:build !test
// +build !test

// Package engine provides the main game loop and Ebiten integration.
// This file implements Game which ties together the ECS world, rendering
// systems, and the Ebiten game engine.
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

	// Rendering systems
	CameraSystem        *CameraSystem
	RenderSystem        *RenderSystem
	TerrainRenderSystem *TerrainRenderSystem
	HUDSystem           *HUDSystem
	TutorialSystem      *TutorialSystem
	HelpSystem          *HelpSystem
}

// NewGame creates a new game instance.
func NewGame(screenWidth, screenHeight int) *Game {
	cameraSystem := NewCameraSystem(screenWidth, screenHeight)
	renderSystem := NewRenderSystem(cameraSystem)
	hudSystem := NewHUDSystem(screenWidth, screenHeight)
	// TerrainRenderSystem will be initialized later with specific genre/seed

	return &Game{
		World:          NewWorld(),
		lastUpdateTime: time.Now(),
		ScreenWidth:    screenWidth,
		ScreenHeight:   screenHeight,
		CameraSystem:   cameraSystem,
		RenderSystem:   renderSystem,
		HUDSystem:      hudSystem,
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

	// Update camera system
	g.CameraSystem.Update(g.World.GetEntities(), deltaTime)

	return nil
}

// Draw implements ebiten.Game interface. Called every frame.
func (g *Game) Draw(screen *ebiten.Image) {
	// Render terrain (if available)
	if g.TerrainRenderSystem != nil {
		g.TerrainRenderSystem.Draw(screen, g.CameraSystem)
	}

	// Render all entities
	g.RenderSystem.Draw(screen, g.World.GetEntities())

	// Render HUD overlay
	g.HUDSystem.Draw(screen)

	// Render tutorial overlay (if active)
	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Draw(screen)
	}

	// Render help overlay (if visible)
	if g.HelpSystem != nil && g.HelpSystem.Visible {
		g.HelpSystem.Draw(screen)
	}
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
