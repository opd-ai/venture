// Package engine provides the core ECS (Entity-Component-System) interfaces.
// This file contains the fundamental interfaces for the ECS architecture and
// dependency injection interfaces that enable testability by abstracting
// Ebiten-specific implementations.
//
// Interface Design Principles:
//   - Production implementations use Ebiten types (*ebiten.Image, input handling, etc.)
//   - Test implementations use simple stubs without external dependencies
//   - All interfaces support both production and test scenarios
//   - No build tags required - *_test.go suffix provides automatic test-only code
package engine

import (
	"image/color"
)

// Component represents a data container attached to an Entity.
// Components should be pure data structures without behavior.
// Originally from: ecs.go
type Component interface {
	// Type returns a unique identifier for this component type
	Type() string
}

// System represents a behavior that operates on entities with specific components.
// Systems should be stateless where possible and operate on entity data.
// Originally from: ecs.go
type System interface {
	// Update is called every frame to update entities managed by this system
	Update(entities []*Entity, deltaTime float64)
}

// GameRunner manages the main game loop and state.
// This interface abstracts the game from Ebiten-specific implementation details.
//
// Implementations:
//   - EbitenGame: Production implementation that integrates with Ebiten game loop
//   - StubGame: Test implementation for unit testing without Ebiten dependencies
type GameRunner interface {
	// GetWorld returns the ECS world instance containing all entities and systems
	GetWorld() *World

	// GetScreenSize returns the current screen dimensions in pixels
	GetScreenSize() (width, height int)

	// IsPaused returns whether the game is currently paused
	IsPaused() bool

	// SetPaused sets the game pause state
	SetPaused(paused bool)

	// SetPlayerEntity sets the player entity reference for UI systems and game logic
	SetPlayerEntity(entity *Entity)

	// GetPlayerEntity returns the current player entity (can be nil)
	GetPlayerEntity() *Entity

	// Update is called every frame to update game state
	// Returns an error to comply with ebiten.Game interface requirements
	Update() error

	// SetInventorySystem connects the inventory system to the UI for player interaction
	SetInventorySystem(system *InventorySystem)

	// SetupInputCallbacks connects input system callbacks to UI systems
	SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem)
}

// ImageProvider provides access to image data without Ebiten-specific types.
// This allows image handling to be abstracted for testing.
//
// Implementations:
//   - EbitenImage: Wraps *ebiten.Image for production use
//   - StubImage: Simple test implementation with size and color data
type ImageProvider interface {
	// GetSize returns the image dimensions in pixels
	GetSize() (width, height int)

	// GetPixel returns the color at the given position
	// Returns color.Transparent if position is out of bounds
	GetPixel(x, y int) color.Color
}

// DrawOptions contains optional parameters for rendering operations.
// All fields are optional and have sensible defaults.
type DrawOptions struct {
	Rotation float64 // Rotation in radians
	ScaleX   float64 // Horizontal scale factor (1.0 = no scaling)
	ScaleY   float64 // Vertical scale factor (1.0 = no scaling)
	Alpha    float32 // Alpha transparency (0.0 = transparent, 1.0 = opaque)
	OffsetX  float64 // X offset from position
	OffsetY  float64 // Y offset from position
}

// Renderer handles drawing of visual elements to the screen.
// This interface abstracts rendering operations from Ebiten-specific implementation.
//
// Implementations:
//   - EbitenRenderer: Production renderer using *ebiten.Image operations
//   - StubRenderer: Test renderer that records draw calls for verification
type Renderer interface {
	// DrawImage draws an image at the specified position with optional transformations
	DrawImage(img ImageProvider, x, y float64, opts *DrawOptions)

	// DrawRect draws a filled rectangle at the specified position and size
	DrawRect(x, y, width, height float64, col color.Color)

	// DrawText draws text at the specified position
	// Position is in screen coordinates (top-left origin)
	DrawText(text string, x, y int, col color.Color)

	// Clear clears the entire screen with the specified color
	Clear(col color.Color)

	// GetBounds returns the rendering bounds (screen size) in pixels
	GetBounds() (width, height int)
}

// SpriteProvider provides sprite visual data without Ebiten dependencies.
// Sprites represent the visual appearance of entities in the game.
//
// Implementations:
//   - EbitenSprite: Production sprite with *ebiten.Image
//   - StubSprite: Test sprite with just properties (no actual image data)
type SpriteProvider interface {
	Component // Inherits Type() string

	// GetImage returns the sprite's image data
	// Returns nil if no image is set (will use solid color rendering)
	GetImage() ImageProvider

	// GetSize returns the sprite dimensions in game units (not pixels)
	GetSize() (width, height float64)

	// GetColor returns the sprite tint color
	// This color is multiplied with the image (or used as solid color if no image)
	GetColor() color.Color

	// GetRotation returns the sprite rotation in radians
	GetRotation() float64

	// GetLayer returns the rendering layer (higher values render on top)
	// Typical range: 0 (background) to 100 (foreground/UI)
	GetLayer() int

	// IsVisible returns whether the sprite should be rendered
	IsVisible() bool

	// SetVisible sets the sprite visibility state
	SetVisible(visible bool)

	// SetColor sets the sprite tint color
	SetColor(col color.Color)

	// SetRotation sets the sprite rotation in radians
	SetRotation(rotation float64)
}

// InputProvider provides access to player input state.
// This interface abstracts input handling from Ebiten's keyboard/mouse implementation.
//
// Implementations:
//   - EbitenInput: Production input reading from ebiten.IsKeyPressed, ebiten.CursorPosition
//   - StubInput: Test input with manually controllable state for deterministic testing
type InputProvider interface {
	Component // Inherits Type() string

	// GetMovement returns normalized movement input on both axes
	// Returns values in range [-1.0, 1.0] for each axis
	// (0, 0) means no movement input
	GetMovement() (x, y float64)

	// IsActionPressed returns whether the primary action button is currently held down
	// Typically mapped to SPACE or mouse click
	IsActionPressed() bool

	// IsActionJustPressed returns whether the action button was pressed THIS frame
	// Used for actions that should trigger once per press (attacks, confirmations)
	IsActionJustPressed() bool

	// IsAnyKeyPressed returns whether any key was pressed this frame
	// Used for "press any key to continue" interactions
	IsAnyKeyPressed() bool

	// IsUseItemPressed returns whether the use item button is currently held down
	IsUseItemPressed() bool

	// IsUseItemJustPressed returns whether use item was pressed THIS frame
	IsUseItemJustPressed() bool

	// IsSpellPressed returns whether a spell hotkey (1-5) is currently pressed
	// slot must be in range [1, 5]
	IsSpellPressed(slot int) bool

	// GetMousePosition returns the current mouse cursor position in screen coordinates
	GetMousePosition() (x, y int)

	// IsMousePressed returns whether the primary mouse button is pressed
	IsMousePressed() bool

	// SetMovement sets the movement input (primarily for testing)
	// Values should be in range [-1.0, 1.0]
	SetMovement(x, y float64)

	// SetActionPressed sets the action button state (primarily for testing)
	SetActionPressed(pressed bool)
}

// RenderingSystem handles visual rendering of entities.
// This interface abstracts the render system from Ebiten-specific drawing operations.
//
// Implementations:
//   - EbitenRenderSystem: Production renderer using Ebiten drawing primitives
//   - StubRenderSystem: Test renderer for verifying render logic without actual drawing
type RenderingSystem interface {
	System // Inherits Update(entities []*Entity, deltaTime float64)

	// Draw renders all visible entities to the screen
	// The screen parameter is interface{} to support both *ebiten.Image (production)
	// and Renderer interface (testing). Concrete implementations should type-assert.
	Draw(screen interface{}, entities []*Entity)

	// SetShowColliders enables or disables debug visualization of collision boxes
	SetShowColliders(show bool)

	// SetShowGrid enables or disables debug visualization of spatial grid
	SetShowGrid(show bool)
}

// UISystem handles user interface rendering and interaction.
// This interface provides a common contract for all UI systems (HUD, menus, dialogs, etc.)
//
// Implementations:
//   - Production: EbitenHUDSystem, EbitenMenuSystem, EbitenCharacterUI, etc.
//   - Test: StubHUDSystem, StubMenuSystem, StubCharacterUI, etc.
type UISystem interface {
	System // Inherits Update(entities []*Entity, deltaTime float64)

	// Draw renders the UI to the screen
	// The screen parameter is interface{} to support both *ebiten.Image (production)
	// and Renderer interface (testing)
	Draw(screen interface{})

	// IsActive returns whether the UI is currently visible and processing input
	IsActive() bool

	// SetActive sets the UI visibility and input processing state
	SetActive(active bool)

	// Note: HandleInput(input InputProvider) bool is optional and not part of the base
	// interface. Individual UI systems can implement it if they need direct input handling.
	// Return value indicates whether the UI consumed the input (blocking lower layers).
}
