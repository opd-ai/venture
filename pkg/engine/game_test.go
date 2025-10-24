package engine

// StubGame is a test implementation of GameRunner interface.
// It provides a simple in-memory game state for unit testing without Ebiten dependencies.
type StubGame struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
	Paused       bool

	// Rendering systems (minimal stubs for testing)
	CameraSystem *CameraSystem

	// Player entity reference
	PlayerEntity *Entity

	// Test state tracking
	UpdateCount int
	InventorySystem *InventorySystem
	InputCallbacksSetup bool
}

// NewStubGame creates a new stub game instance for testing.
func NewStubGame(screenWidth, screenHeight int) *StubGame {
	world := NewWorld()
	cameraSystem := NewCameraSystem(screenWidth, screenHeight)

	return &StubGame{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		CameraSystem: cameraSystem,
		Paused:       false,
		UpdateCount:  0,
	}
}

// GetWorld returns the ECS world instance (implements GameRunner).
func (g *StubGame) GetWorld() *World {
	return g.World
}

// GetScreenSize returns the screen dimensions (implements GameRunner).
func (g *StubGame) GetScreenSize() (width, height int) {
	return g.ScreenWidth, g.ScreenHeight
}

// IsPaused returns whether the game is paused (implements GameRunner).
func (g *StubGame) IsPaused() bool {
	return g.Paused
}

// SetPaused sets the game pause state (implements GameRunner).
func (g *StubGame) SetPaused(paused bool) {
	g.Paused = paused
}

// SetPlayerEntity sets the player entity (implements GameRunner).
func (g *StubGame) SetPlayerEntity(entity *Entity) {
	g.PlayerEntity = entity
}

// GetPlayerEntity returns the player entity (implements GameRunner).
func (g *StubGame) GetPlayerEntity() *Entity {
	return g.PlayerEntity
}

// Update updates game state (implements GameRunner).
// For testing, this just increments a counter and updates the world if not paused.
func (g *StubGame) Update() error {
	g.UpdateCount++
	if !g.Paused {
		g.World.Update(0.016) // Simulate 60 FPS (1/60 ≈ 0.016 seconds)
	}
	return nil
}

// SetInventorySystem connects the inventory system (implements GameRunner).
func (g *StubGame) SetInventorySystem(system *InventorySystem) {
	g.InventorySystem = system
}

// SetupInputCallbacks connects input callbacks (implements GameRunner).
// For testing, this just sets a flag to track that it was called.
func (g *StubGame) SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) {
	g.InputCallbacksSetup = true
}

// Compile-time interface check
var _ GameRunner = (*StubGame)(nil)
