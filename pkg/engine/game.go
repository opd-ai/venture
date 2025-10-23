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
	MenuSystem          *MenuSystem

	// UI systems
	InventoryUI *InventoryUI
	QuestUI     *QuestUI

	// Player entity reference (for UI systems)
	PlayerEntity *Entity
}

// NewGame creates a new game instance.
func NewGame(screenWidth, screenHeight int) *Game {
	world := NewWorld()
	cameraSystem := NewCameraSystem(screenWidth, screenHeight)
	renderSystem := NewRenderSystem(cameraSystem)
	hudSystem := NewHUDSystem(screenWidth, screenHeight)
	// TerrainRenderSystem will be initialized later with specific genre/seed

	// Create UI systems
	inventoryUI := NewInventoryUI(world, screenWidth, screenHeight)
	questUI := NewQuestUI(world, screenWidth, screenHeight)

	// Create menu system with save directory
	menuSystem, err := NewMenuSystem(world, screenWidth, screenHeight, "./saves")
	if err != nil {
		// Log error but continue (save/load won't work but game can run)
		fmt.Printf("Warning: Failed to initialize menu system: %v\n", err)
	}

	return &Game{
		World:          world,
		lastUpdateTime: time.Now(),
		ScreenWidth:    screenWidth,
		ScreenHeight:   screenHeight,
		CameraSystem:   cameraSystem,
		RenderSystem:   renderSystem,
		HUDSystem:      hudSystem,
		MenuSystem:     menuSystem,
		InventoryUI:    inventoryUI,
		QuestUI:        questUI,
	}
}

// Update implements ebiten.Game interface. Called every frame.
func (g *Game) Update() error {
	// Calculate delta time
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Cap delta time to prevent spiral of death
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	// If menu is visible, pause game world (but allow menu input)
	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.Paused = true
		// Update menu even when paused
		g.MenuSystem.Update(g.World.GetEntities(), deltaTime)
		return nil
	}

	if g.Paused {
		return nil
	}

	// Update UI systems first (they capture input if visible)
	g.InventoryUI.Update()
	g.QuestUI.Update()

	// Gap #6: Always update tutorial system for progress tracking (even when UI visible)
	if g.TutorialSystem != nil && g.TutorialSystem.Enabled {
		g.TutorialSystem.Update(g.World.GetEntities(), deltaTime)
	}

	// Update the world (unless UI is blocking input)
	if !g.InventoryUI.IsVisible() && !g.QuestUI.IsVisible() {
		g.World.Update(deltaTime)
	}

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

	// Render menu overlay (if active)
	if g.MenuSystem != nil && g.MenuSystem.IsActive() {
		g.MenuSystem.Draw(screen)
	}

	// Render UI overlays (drawn last so they're on top)
	g.InventoryUI.Draw(screen)
	g.QuestUI.Draw(screen)
}

// Layout implements ebiten.Game interface. Returns the game's screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ScreenWidth, g.ScreenHeight
}

// SetPlayerEntity sets the player entity for the game and UI systems.
// This should be called after creating the player entity.
func (g *Game) SetPlayerEntity(entity *Entity) {
	g.PlayerEntity = entity
	g.InventoryUI.SetPlayerEntity(entity)
	g.QuestUI.SetPlayerEntity(entity)
}

// SetInventorySystem connects the inventory system to the inventory UI for item actions.
func (g *Game) SetInventorySystem(system *InventorySystem) {
	g.InventoryUI.SetInventorySystem(system)
}

// SetupInputCallbacks connects the input system callbacks to the UI systems.
// This should be called after the InputSystem is added to the world.
func (g *Game) SetupInputCallbacks(inputSystem *InputSystem) {
	// Connect inventory toggle
	inputSystem.SetInventoryCallback(func() {
		g.InventoryUI.Toggle()
	})

	// Connect quest log toggle
	inputSystem.SetQuestsCallback(func() {
		g.QuestUI.Toggle()
	})

	// Connect pause menu toggle (ESC key)
	if g.MenuSystem != nil {
		inputSystem.SetMenuToggleCallback(func() {
			g.MenuSystem.Toggle()
		})
	}

	// TODO: Connect other callbacks when character/skills/map UIs are implemented
	// inputSystem.SetCharacterCallback(func() { ... })
	// inputSystem.SetSkillsCallback(func() { ... })
	// inputSystem.SetMapCallback(func() { ... })
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
