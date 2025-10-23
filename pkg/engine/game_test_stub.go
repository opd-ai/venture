//go:build test
// +build test

// Package engine provides test stubs for Game type and related systems.
// This file provides stub implementations when building with the test tag,
// allowing unit tests to compile without Ebiten/X11 dependencies.
package engine

// Game represents the main game instance (test stub).
type Game struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
	Paused       bool

	// Rendering systems (stubs)
	CameraSystem        *CameraSystem
	RenderSystem        *RenderSystem
	TerrainRenderSystem *TerrainRenderSystem
	HUDSystem           *HUDSystem
	TutorialSystem      *TutorialSystem
	HelpSystem          *HelpSystem
	MenuSystem          *MenuSystem

	// UI systems (stubs)
	InventoryUI *InventoryUI
	QuestUI     *QuestUI
	CharacterUI *CharacterUI
	SkillsUI    *SkillsUI
	MapUI       *MapUI

	// Player entity reference
	PlayerEntity *Entity
}

// NewGame creates a new game instance (test stub).
func NewGame(screenWidth, screenHeight int) *Game {
	world := NewWorld()
	cameraSystem := NewCameraSystem(screenWidth, screenHeight)

	return &Game{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		CameraSystem: cameraSystem,
	}
}

// SetPlayerEntity sets the player entity for the game and UI systems (test stub).
func (g *Game) SetPlayerEntity(entity *Entity) {
	g.PlayerEntity = entity
}

// SetInventorySystem connects the inventory system to the inventory UI (test stub).
func (g *Game) SetInventorySystem(system *InventorySystem) {
	// Stub - no op in tests
}

// SetupInputCallbacks connects the input system callbacks to the UI systems (test stub).
func (g *Game) SetupInputCallbacks(inputSystem *InputSystem) {
	// Stub - no op in tests
}

// Run starts the game loop (test stub - does nothing).
func (g *Game) Run(title string) error {
	// Stub - tests don't actually run game loop
	return nil
}

// Stub system types for test builds
type TerrainRenderSystem struct{}

func (t *TerrainRenderSystem) Update(world *World) {}

type TutorialSystem struct{}

func (t *TutorialSystem) Update(world *World) {}

type HelpSystem struct{}

func (h *HelpSystem) Update(world *World) {}

type InputSystem struct{}

func (i *InputSystem) Update(world *World) {}
