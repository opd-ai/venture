# Build Tag Refactoring - Implementation Plan

**Project:** Venture - Procedural Action RPG  
**Date:** October 24, 2025  
**Branch:** `interfaces`  
**Estimated Time:** 13 hours  
**Status:** Ready for Execution

---

## Executive Summary

**Problem:** Build tags (`//go:build test`) create mutual exclusivity preventing production and test code from coexisting, breaking standard Go workflows.

**Solution:** Interface-based dependency injection following idiomatic Go patterns.

**Goal:** Zero build tags, full build/test compatibility, improved testability.

---

## Prerequisites

### Understanding the Problem

Read these documents in order:
1. `BUILD_TAG_ISSUES.md` - Understand why build tags fail
2. `REFACTORING_SUMMARY.md` - High-level overview
3. `INTERFACE_DESIGN.md` - Detailed solution architecture
4. `QUICK_REFERENCE.md` - Developer patterns

### Verify Current State

```bash
# Should succeed
go build ./...

# Should fail with build tag errors
go test ./cmd/client
go test ./cmd/server

# Should succeed but limited to pkg/*
go test -tags test ./pkg/...

# Count build tag usage
grep -r "//go:build test" pkg/ | wc -l
# Expected: ~28 matches in pkg/engine
```

### Create Backup Branch

```bash
git checkout main
git pull
git checkout -b interfaces-backup
git checkout interfaces
```

---

## Phase 2: Implementation (13 hours)

### Phase 2a: Create Core Interfaces (1 hour)

**File:** `pkg/engine/interfaces.go`

#### Step 2a.1: Create interfaces.go skeleton (10 min)

```go
// Package engine provides the core game engine interfaces.
// These interfaces enable dependency injection and testability by abstracting
// Ebiten-specific implementations.
package engine

import (
	"image/color"
)

// Component is the base interface for all ECS components.
// This already exists in the codebase - reference it here for clarity.
type Component interface {
	Type() string
}

// System is the base interface for all ECS systems.
// This already exists in the codebase - reference it here for clarity.
type System interface {
	Update(entities []*Entity, deltaTime float64)
}
```

#### Step 2a.2: Add GameRunner interface (10 min)

```go
// GameRunner manages the main game loop and state.
// Implementations: EbitenGame (production), StubGame (test)
type GameRunner interface {
	// GetWorld returns the ECS world instance
	GetWorld() *World
	
	// GetScreenSize returns the current screen dimensions
	GetScreenSize() (width, height int)
	
	// IsPaused returns whether the game is currently paused
	IsPaused() bool
	
	// SetPaused sets the pause state
	SetPaused(paused bool)
	
	// SetPlayerEntity sets the player entity for UI systems
	SetPlayerEntity(entity *Entity)
	
	// GetPlayerEntity returns the current player entity
	GetPlayerEntity() *Entity
	
	// Update is called every frame (returns error for Ebiten compatibility)
	Update() error
	
	// SetInventorySystem connects the inventory system to the UI
	SetInventorySystem(system *InventorySystem)
	
	// SetupInputCallbacks connects input callbacks to UI systems
	SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem)
}
```

#### Step 2a.3: Add rendering interfaces (15 min)

```go
// ImageProvider provides access to image data without Ebiten dependency.
// Implementations: EbitenImage (wraps *ebiten.Image), StubImage (test)
type ImageProvider interface {
	// GetSize returns the image dimensions
	GetSize() (width, height int)
	
	// GetPixel returns the color at the given position (optional, for testing)
	GetPixel(x, y int) color.Color
}

// DrawOptions contains optional rendering parameters
type DrawOptions struct {
	Rotation   float64
	ScaleX     float64
	ScaleY     float64
	Alpha      float32
	OffsetX    float64
	OffsetY    float64
}

// Renderer handles drawing of visual elements.
// Implementations: EbitenRenderer (production), StubRenderer (test)
type Renderer interface {
	// DrawImage draws an image at the given position with options
	DrawImage(img ImageProvider, x, y float64, opts *DrawOptions)
	
	// DrawRect draws a filled rectangle
	DrawRect(x, y, width, height float64, col color.Color)
	
	// DrawText draws text at the given position
	DrawText(text string, x, y int, col color.Color)
	
	// Clear clears the screen with the given color
	Clear(col color.Color)
	
	// GetBounds returns the rendering bounds
	GetBounds() (width, height int)
}
```

#### Step 2a.4: Add component interfaces (15 min)

```go
// SpriteProvider provides sprite visual data without Ebiten dependency.
// Implementations: EbitenSprite (production), StubSprite (test)
type SpriteProvider interface {
	Component // Inherits Type() string
	
	// GetImage returns the sprite image (can be nil)
	GetImage() ImageProvider
	
	// GetSize returns sprite dimensions
	GetSize() (width, height float64)
	
	// GetColor returns the sprite tint color
	GetColor() color.Color
	
	// GetRotation returns rotation in radians
	GetRotation() float64
	
	// GetLayer returns the render layer (higher = on top)
	GetLayer() int
	
	// IsVisible returns whether the sprite should be drawn
	IsVisible() bool
	
	// SetVisible sets the visibility state
	SetVisible(visible bool)
	
	// SetColor sets the sprite tint color
	SetColor(col color.Color)
	
	// SetRotation sets the rotation in radians
	SetRotation(rotation float64)
}

// InputProvider provides input state without Ebiten dependency.
// Implementations: EbitenInput (production), StubInput (test)
type InputProvider interface {
	Component // Inherits Type() string
	
	// GetMovement returns movement input (-1.0 to 1.0 for each axis)
	GetMovement() (x, y float64)
	
	// IsActionPressed returns whether the action button is currently pressed
	IsActionPressed() bool
	
	// IsActionJustPressed returns whether action was just pressed this frame
	IsActionJustPressed() bool
	
	// IsUseItemPressed returns whether the use item button is currently pressed
	IsUseItemPressed() bool
	
	// IsUseItemJustPressed returns whether use item was just pressed this frame
	IsUseItemJustPressed() bool
	
	// IsSpellPressed returns whether a spell hotkey (1-5) is pressed
	IsSpellPressed(slot int) bool
	
	// GetMousePosition returns mouse coordinates
	GetMousePosition() (x, y int)
	
	// IsMousePressed returns whether mouse button is pressed
	IsMousePressed() bool
	
	// SetMovement sets movement input (for testing)
	SetMovement(x, y float64)
	
	// SetActionPressed sets action button state (for testing)
	SetActionPressed(pressed bool)
}
```

#### Step 2a.5: Add system interfaces (10 min)

```go
// RenderingSystem handles visual rendering of entities.
// Implementations: EbitenRenderSystem (production), StubRenderSystem (test)
type RenderingSystem interface {
	System // Inherits Update()
	
	// Draw renders all entities to the screen/renderer
	// Parameter can be *ebiten.Image (production) or Renderer interface (test)
	Draw(screen interface{}, entities []*Entity)
	
	// SetShowColliders enables/disables collider visualization
	SetShowColliders(show bool)
	
	// SetShowGrid enables/disables grid visualization
	SetShowGrid(show bool)
}

// UISystem handles user interface rendering and interaction.
// Implementations: Various (HUD, Menu, Character, Skills, Map, Inventory, Quest)
type UISystem interface {
	System // Inherits Update()
	
	// Draw renders the UI to the screen
	Draw(screen interface{})
	
	// IsActive returns whether the UI is currently visible
	IsActive() bool
	
	// SetActive sets the UI visibility
	SetActive(active bool)
	
	// HandleInput processes input for the UI (returns true if input was consumed)
	// Optional - not all UI systems need this
	// HandleInput(input InputProvider) bool
}
```

**Checkpoint:**
```bash
go build ./pkg/engine
# Should succeed - interfaces don't break anything
git add pkg/engine/interfaces.go
git commit -m "feat(engine): add core interfaces for dependency injection

- GameRunner: abstract game loop
- Renderer/ImageProvider: abstract rendering
- SpriteProvider/InputProvider: abstract components
- RenderingSystem/UISystem: abstract systems

Part of build tag refactoring (Phase 2a)"
```

---

### Phase 2b: Migrate Game Type (2 hours)

#### Step 2b.1: Rename Game to EbitenGame (20 min)

**File:** `pkg/engine/game.go`

1. Change build tag from `//go:build !test` to remove it entirely
2. Rename `type Game struct` to `type EbitenGame struct`
3. Update all method receivers: `func (g *Game)` → `func (g *EbitenGame)`
4. Update constructor: `func NewGame(...)` → `func NewEbitenGame(...)`
5. Add comment that it implements both `ebiten.Game` and `GameRunner`

```go
// Remove these lines at top:
//go:build !test
// +build !test

// Update struct:
// EbitenGame represents the main game instance with Ebiten integration.
// Implements both ebiten.Game and GameRunner interfaces.
type EbitenGame struct {
    World          *World
    lastUpdateTime time.Time
    ScreenWidth    int
    ScreenHeight   int
    // ... rest of fields
}

// Update constructor:
func NewEbitenGame(screenWidth, screenHeight int) *EbitenGame {
    // ... implementation
    return &EbitenGame{
        // ... fields
    }
}

// Update all methods:
func (g *EbitenGame) Update() error { ... }
func (g *EbitenGame) Draw(screen *ebiten.Image) { ... }
func (g *EbitenGame) Layout(outsideWidth, outsideHeight int) (int, int) { ... }
func (g *EbitenGame) SetPlayerEntity(entity *Entity) { ... }
// etc.
```

**Verify GameRunner interface implementation:**

Add at bottom of file:
```go
// Compile-time interface check
var _ GameRunner = (*EbitenGame)(nil)
var _ ebiten.Game = (*EbitenGame)(nil)
```

#### Step 2b.2: Create StubGame (20 min)

**File:** `pkg/engine/game_test.go`

Create new file (no build tags needed - `*_test.go` auto-excludes):

```go
package engine

// StubGame is a test implementation of GameRunner without Ebiten dependencies.
type StubGame struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
	Paused       bool
	PlayerEntity *Entity
	
	// Test controls
	UpdateFunc func() error
	UpdateCalls int
}

// NewStubGame creates a new stub game for testing.
func NewStubGame(screenWidth, screenHeight int) *StubGame {
	return &StubGame{
		World:       NewWorld(),
		ScreenWidth: screenWidth,
		ScreenHeight: screenHeight,
	}
}

// GetWorld implements GameRunner.
func (g *StubGame) GetWorld() *World {
	return g.World
}

// GetScreenSize implements GameRunner.
func (g *StubGame) GetScreenSize() (width, height int) {
	return g.ScreenWidth, g.ScreenHeight
}

// IsPaused implements GameRunner.
func (g *StubGame) IsPaused() bool {
	return g.Paused
}

// SetPaused implements GameRunner.
func (g *StubGame) SetPaused(paused bool) {
	g.Paused = paused
}

// SetPlayerEntity implements GameRunner.
func (g *StubGame) SetPlayerEntity(entity *Entity) {
	g.PlayerEntity = entity
}

// GetPlayerEntity implements GameRunner.
func (g *StubGame) GetPlayerEntity() *Entity {
	return g.PlayerEntity
}

// Update implements GameRunner.
func (g *StubGame) Update() error {
	g.UpdateCalls++
	if g.UpdateFunc != nil {
		return g.UpdateFunc()
	}
	// Default: update world
	if g.World != nil && !g.Paused {
		g.World.Update(0.016) // 60 FPS
	}
	return nil
}

// SetInventorySystem implements GameRunner.
func (g *StubGame) SetInventorySystem(system *InventorySystem) {
	// Stub - no-op in tests
}

// SetupInputCallbacks implements GameRunner.
func (g *StubGame) SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem) {
	// Stub - no-op in tests
}

// Compile-time interface check
var _ GameRunner = (*StubGame)(nil)
```

#### Step 2b.3: Update Game references (60 min)

Search and replace across codebase:

```bash
# Find all references to *Game type
grep -r "\*Game" --include="*.go" pkg/ cmd/ examples/

# Update references systematically:
# In production code: *Game → *EbitenGame
# In test code: *Game → *StubGame or GameRunner interface
# In interfaces: accept GameRunner interface
```

**Key files to update:**
- `cmd/client/main.go` - use `NewEbitenGame`
- `cmd/server/main.go` - if it uses Game
- Any system constructors that take `*Game` - change to `GameRunner`
- Test files - use `NewStubGame` or `&StubGame{}`

#### Step 2b.4: Delete old stub file (5 min)

```bash
git rm pkg/engine/game_test_stub.go
```

**Checkpoint:**
```bash
go build ./pkg/engine
go test ./pkg/engine
go build ./cmd/client
# All should succeed

git add -A
git commit -m "refactor(engine): migrate Game to interface-based pattern

- Renamed Game → EbitenGame (production implementation)
- Created StubGame in game_test.go (test implementation)
- Both implement GameRunner interface
- Removed //go:build tags from game.go
- Deleted game_test_stub.go
- Updated all references to use interface or concrete types

Part of build tag refactoring (Phase 2b)"
```

---

### Phase 2c: Migrate Components (2 hours)

#### Step 2c.1: Migrate SpriteComponent (60 min)

**Step 2c.1a: Create EbitenSprite**

**File:** `pkg/engine/sprite_component.go` (rename from components.go or create new)

```go
// Remove build tags if present

package engine

import (
	"image/color"
	
	"github.com/hajimehoshi/ebiten/v2"
)

// EbitenSprite is the production sprite component using Ebiten.
type EbitenSprite struct {
	Image    *ebiten.Image
	Width    float64
	Height   float64
	Color    color.Color
	Rotation float64
	Visible  bool
	Layer    int
}

// Type implements Component.
func (s *EbitenSprite) Type() string {
	return "sprite"
}

// GetImage implements SpriteProvider.
func (s *EbitenSprite) GetImage() ImageProvider {
	if s.Image == nil {
		return nil
	}
	return &EbitenImage{image: s.Image}
}

// GetSize implements SpriteProvider.
func (s *EbitenSprite) GetSize() (width, height float64) {
	return s.Width, s.Height
}

// GetColor implements SpriteProvider.
func (s *EbitenSprite) GetColor() color.Color {
	if s.Color == nil {
		return color.White
	}
	return s.Color
}

// GetRotation implements SpriteProvider.
func (s *EbitenSprite) GetRotation() float64 {
	return s.Rotation
}

// GetLayer implements SpriteProvider.
func (s *EbitenSprite) GetLayer() int {
	return s.Layer
}

// IsVisible implements SpriteProvider.
func (s *EbitenSprite) IsVisible() bool {
	return s.Visible
}

// SetVisible implements SpriteProvider.
func (s *EbitenSprite) SetVisible(visible bool) {
	s.Visible = visible
}

// SetColor implements SpriteProvider.
func (s *EbitenSprite) SetColor(col color.Color) {
	s.Color = col
}

// SetRotation implements SpriteProvider.
func (s *EbitenSprite) SetRotation(rotation float64) {
	s.Rotation = rotation
}

// NewSpriteComponent creates a new Ebiten sprite component.
func NewSpriteComponent(width, height float64, col color.Color) *EbitenSprite {
	return &EbitenSprite{
		Width:   width,
		Height:  height,
		Color:   col,
		Visible: true,
		Layer:   0,
	}
}

// EbitenImage wraps an Ebiten image for the ImageProvider interface.
type EbitenImage struct {
	image *ebiten.Image
}

// GetSize implements ImageProvider.
func (e *EbitenImage) GetSize() (width, height int) {
	if e.image == nil {
		return 0, 0
	}
	return e.image.Bounds().Dx(), e.image.Bounds().Dy()
}

// GetPixel implements ImageProvider.
func (e *EbitenImage) GetPixel(x, y int) color.Color {
	if e.image == nil {
		return color.Transparent
	}
	return e.image.At(x, y)
}

// Compile-time interface check
var _ SpriteProvider = (*EbitenSprite)(nil)
var _ ImageProvider = (*EbitenImage)(nil)
```

**Step 2c.1b: Create StubSprite**

**File:** `pkg/engine/sprite_component_test.go`

```go
package engine

import "image/color"

// StubSprite is a test sprite component without Ebiten dependencies.
type StubSprite struct {
	Width    float64
	Height   float64
	Color    color.Color
	Rotation float64
	Visible  bool
	Layer    int
}

// Type implements Component.
func (s *StubSprite) Type() string {
	return "sprite"
}

// GetImage implements SpriteProvider.
func (s *StubSprite) GetImage() ImageProvider {
	return nil // Stub has no actual image
}

// GetSize implements SpriteProvider.
func (s *StubSprite) GetSize() (width, height float64) {
	return s.Width, s.Height
}

// GetColor implements SpriteProvider.
func (s *StubSprite) GetColor() color.Color {
	if s.Color == nil {
		return color.White
	}
	return s.Color
}

// GetRotation implements SpriteProvider.
func (s *StubSprite) GetRotation() float64 {
	return s.Rotation
}

// GetLayer implements SpriteProvider.
func (s *StubSprite) GetLayer() int {
	return s.Layer
}

// IsVisible implements SpriteProvider.
func (s *StubSprite) IsVisible() bool {
	return s.Visible
}

// SetVisible implements SpriteProvider.
func (s *StubSprite) SetVisible(visible bool) {
	s.Visible = visible
}

// SetColor implements SpriteProvider.
func (s *StubSprite) SetColor(col color.Color) {
	s.Color = col
}

// SetRotation implements SpriteProvider.
func (s *StubSprite) SetRotation(rotation float64) {
	s.Rotation = rotation
}

// NewStubSprite creates a new stub sprite for testing.
func NewStubSprite(width, height float64, col color.Color) *StubSprite {
	return &StubSprite{
		Width:   width,
		Height:  height,
		Color:   col,
		Visible: true,
		Layer:   0,
	}
}

// Compile-time interface check
var _ SpriteProvider = (*StubSprite)(nil)
```

**Step 2c.1c: Update SpriteComponent references**

Update code that uses `*SpriteComponent` to use `SpriteProvider` interface:

```go
// Old:
sprite := comp.(*SpriteComponent)

// New:
sprite := comp.(SpriteProvider)
```

Delete `pkg/engine/components_test_stub.go` (the SpriteComponent part).

#### Step 2c.2: Migrate InputComponent (60 min)

**Similar process:**
1. Create `pkg/engine/input_component.go` with `EbitenInput`
2. Create `pkg/engine/input_component_test.go` with `StubInput`
3. Update references to use `InputProvider` interface
4. Delete stub file

**File:** `pkg/engine/input_component.go`

```go
package engine

// EbitenInput is the production input component.
type EbitenInput struct {
	MoveX, MoveY           float64
	ActionPressed          bool
	ActionJustPressed      bool
	UseItemPressed         bool
	UseItemJustPressed     bool
	Spell1Pressed          bool
	Spell2Pressed          bool
	Spell3Pressed          bool
	Spell4Pressed          bool
	Spell5Pressed          bool
	AnyKeyPressed          bool
	SecondaryAction        bool
	Action                 bool // Alias
	MouseX, MouseY         int
	MousePressed           bool
}

// Type implements Component.
func (i *EbitenInput) Type() string {
	return "input"
}

// GetMovement implements InputProvider.
func (i *EbitenInput) GetMovement() (x, y float64) {
	return i.MoveX, i.MoveY
}

// IsActionPressed implements InputProvider.
func (i *EbitenInput) IsActionPressed() bool {
	return i.ActionPressed
}

// IsActionJustPressed implements InputProvider.
func (i *EbitenInput) IsActionJustPressed() bool {
	return i.ActionJustPressed
}

// IsUseItemPressed implements InputProvider.
func (i *EbitenInput) IsUseItemPressed() bool {
	return i.UseItemPressed
}

// IsUseItemJustPressed implements InputProvider.
func (i *EbitenInput) IsUseItemJustPressed() bool {
	return i.UseItemJustPressed
}

// IsSpellPressed implements InputProvider.
func (i *EbitenInput) IsSpellPressed(slot int) bool {
	switch slot {
	case 1:
		return i.Spell1Pressed
	case 2:
		return i.Spell2Pressed
	case 3:
		return i.Spell3Pressed
	case 4:
		return i.Spell4Pressed
	case 5:
		return i.Spell5Pressed
	default:
		return false
	}
}

// GetMousePosition implements InputProvider.
func (i *EbitenInput) GetMousePosition() (x, y int) {
	return i.MouseX, i.MouseY
}

// IsMousePressed implements InputProvider.
func (i *EbitenInput) IsMousePressed() bool {
	return i.MousePressed
}

// SetMovement implements InputProvider.
func (i *EbitenInput) SetMovement(x, y float64) {
	i.MoveX = x
	i.MoveY = y
}

// SetActionPressed implements InputProvider.
func (i *EbitenInput) SetActionPressed(pressed bool) {
	i.ActionPressed = pressed
	i.Action = pressed // Keep alias in sync
}

// Compile-time interface check
var _ InputProvider = (*EbitenInput)(nil)
```

**File:** `pkg/engine/input_component_test.go`

```go
package engine

// StubInput is a test input component for controllable input state.
type StubInput struct {
	MoveX, MoveY       float64
	ActionPressed      bool
	ActionJustPressed  bool
	UseItemPressed     bool
	UseItemJustPressed bool
	Spell1Pressed      bool
	Spell2Pressed      bool
	Spell3Pressed      bool
	Spell4Pressed      bool
	Spell5Pressed      bool
	AnyKeyPressed      bool
	MouseX, MouseY     int
	MousePressed       bool
}

// Type implements Component.
func (i *StubInput) Type() string {
	return "input"
}

// Implement all InputProvider methods (same as EbitenInput)
// ... (copy implementation from above)

// Compile-time interface check
var _ InputProvider = (*StubInput)(nil)
```

**Checkpoint:**
```bash
go build ./pkg/engine
go test ./pkg/engine
git add -A
git commit -m "refactor(engine): migrate components to interface pattern

- Created EbitenSprite and StubSprite implementing SpriteProvider
- Created EbitenInput and StubInput implementing InputProvider
- Removed build tags from component files
- Updated references to use interfaces
- Deleted components_test_stub.go

Part of build tag refactoring (Phase 2c)"
```

---

### Phase 2d: Migrate Core Systems (3 hours)

#### Step 2d.1: Migrate RenderSystem (60 min)

**Process:**
1. Rename `RenderSystem` → `EbitenRenderSystem` in `render_system.go`
2. Remove build tags
3. Create `StubRenderSystem` in `render_system_test.go`
4. Update constructor to return `RenderingSystem` interface
5. Delete `render_system_test_stub.go`

**Pattern for each system:**
```go
// render_system.go (remove build tags)
type EbitenRenderSystem struct { ... }
func NewRenderSystem(...) RenderingSystem {
    return &EbitenRenderSystem{...}
}

// render_system_test.go
type StubRenderSystem struct { ... }
func NewStubRenderSystem() RenderingSystem {
    return &StubRenderSystem{}
}
```

#### Step 2d.2: Migrate TutorialSystem (60 min)

Same pattern - rename, create stub, remove build tags.

Note: `TutorialSystem` already has both production (`tutorial_system.go`) and test (`tutorial_system_test.go`) versions. Ensure they follow the new pattern.

#### Step 2d.3: Migrate HelpSystem (60 min)

Same pattern for HelpSystem.

**Checkpoint:**
```bash
go build ./pkg/engine
go test ./pkg/engine
git add -A
git commit -m "refactor(engine): migrate core systems to interface pattern

- Migrated RenderSystem → EbitenRenderSystem/StubRenderSystem
- Migrated TutorialSystem → EbitenTutorialSystem/StubTutorialSystem
- Migrated HelpSystem → EbitenHelpSystem/StubHelpSystem
- All implement appropriate interfaces
- Removed build tags
- Deleted stub files

Part of build tag refactoring (Phase 2d)"
```

---

### Phase 2e: Migrate UI Systems (3 hours)

Migrate each UI system following the same pattern:

#### Systems to migrate:
1. **HUDSystem** (30 min)
2. **MenuSystem** (30 min)
3. **CharacterUI** (20 min)
4. **SkillsUI** (20 min)
5. **MapUI** (20 min)
6. **InventoryUI** (20 min)
7. **QuestUI** (20 min)

**For each system:**
```bash
# Example for HUDSystem:
1. Edit pkg/engine/hud_system.go
   - Remove build tags
   - Rename HUDSystem → EbitenHUDSystem
   - Update constructor: func NewHUDSystem(...) UISystem

2. Create pkg/engine/hud_system_test.go
   - type StubHUDSystem struct
   - Implement UISystem interface
   - Simple no-op implementations

3. Delete pkg/engine/hud_system_test_stub.go

4. Test:
   go build ./pkg/engine
   go test ./pkg/engine

5. Commit:
   git add -A
   git commit -m "refactor(engine): migrate HUDSystem to interface pattern"
```

**Bulk checkpoint:**
```bash
go build ./pkg/engine
go test ./pkg/engine
go build ./cmd/client
# Should all succeed now!

git add -A
git commit -m "refactor(engine): migrate all UI systems to interface pattern

- Migrated 7 UI systems to interface-based pattern
- Each has Ebiten* (production) and Stub* (test) implementations
- All implement UISystem interface
- Removed all build tags
- Deleted all *_test_stub.go files

Part of build tag refactoring (Phase 2e)"
```

---

## Phase 3: Cleanup & Validation (2 hours)

### Step 3.1: Verify all build tag stubs are removed (15 min)

```bash
# Check for remaining stub files
find pkg/engine -name "*_test_stub.go"
# Should return empty

# Check for remaining build tags in pkg/engine
grep -r "//go:build test" pkg/engine/*.go
grep -r "// +build test" pkg/engine/*.go
# Should return empty (except in *_test.go files which is OK)

# Count actual build tag usage
grep -r "//go:build test" pkg/engine --include="*.go" | grep -v "_test.go" | wc -l
# Should be 0
```

### Step 3.2: Verify builds (20 min)

```bash
# Clean build cache
go clean -cache

# Build all packages
go build ./...
# Must succeed

# Build specific binaries
go build -o venture-client ./cmd/client
go build -o venture-server ./cmd/server
# Must succeed

# Test all packages
go test ./...
# Must succeed

# Test with race detector
go test -race ./pkg/...
# Must succeed

# Vet all code
go vet ./...
# Must pass
```

### Step 3.3: Measure coverage (20 min)

```bash
# Get baseline (if not already done)
go test -cover ./... -coverprofile=coverage_after.out 2>&1 | tee coverage_after.txt

# Generate HTML report
go tool cover -html=coverage_after.out -o coverage_after.html

# Compare key metrics
echo "Coverage Summary:"
go tool cover -func=coverage_after.out | grep total
```

**Expected results:**
- All packages should compile
- Test coverage should be >= baseline
- No build tag conflicts

### Step 3.4: Test cmd/* packages (15 min)

These should now work:

```bash
# Test client
go test ./cmd/client
# Should succeed (even if no tests, should compile)

# Test server
go test ./cmd/server
# Should succeed

# Test examples
go test ./examples/...
# Check results
```

### Step 3.5: Update existing tests (30 min)

Review test files that may need updates:

```bash
# Find tests that might reference old types
grep -r "Game{" pkg/engine/*_test.go
grep -r "SpriteComponent{" pkg/engine/*_test.go
grep -r "InputComponent{" pkg/engine/*_test.go

# Update to use new types:
# Game{} → StubGame{} or EbitenGame{}
# &SpriteComponent{} → &StubSprite{} or &EbitenSprite{}
# &InputComponent{} → &StubInput{} or &EbitenInput{}
```

### Step 3.6: Create TESTING.md (20 min)

**File:** `docs/TESTING.md`

```markdown
# Testing Guide

## Overview

Venture uses interface-based dependency injection for testability. Production code uses Ebiten implementations, while tests use stub implementations.

## Test Structure

### Component Testing

```go
func TestSpriteComponent(t *testing.T) {
    entity := engine.NewEntity(1)
    sprite := &engine.StubSprite{
        Width: 32,
        Height: 32,
        Color: color.RGBA{255, 0, 0, 255},
        Visible: true,
    }
    entity.AddComponent(sprite)
    
    // Test via interface
    comp, ok := entity.GetComponent("sprite")
    if !ok {
        t.Fatal("Expected sprite component")
    }
    
    sp := comp.(engine.SpriteProvider)
    w, h := sp.GetSize()
    if w != 32 || h != 32 {
        t.Errorf("Expected 32x32, got %fx%f", w, h)
    }
}
```

### System Testing

```go
func TestRenderSystem(t *testing.T) {
    world := engine.NewWorld()
    system := engine.NewStubRenderSystem()
    world.AddSystem(system)
    
    // Add test entities
    entity := world.CreateEntity()
    entity.AddComponent(&engine.StubSprite{Width: 32, Height: 32})
    entity.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
    
    // Test system
    world.Update(0.016)
}
```

### Game Loop Testing

```go
func TestGameLoop(t *testing.T) {
    game := engine.NewStubGame(800, 600)
    game.World.AddSystem(engine.NewMovementSystem(200.0))
    
    err := game.Update()
    if err != nil {
        t.Fatal(err)
    }
}
```

## Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/engine

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Verbose
go test -v ./pkg/engine
```

## Interface Implementations

### Components

- **SpriteProvider**: `EbitenSprite` (prod), `StubSprite` (test)
- **InputProvider**: `EbitenInput` (prod), `StubInput` (test)

### Systems

- **RenderingSystem**: `EbitenRenderSystem` (prod), `StubRenderSystem` (test)
- **UISystem**: Various Ebiten* and Stub* implementations

### Game

- **GameRunner**: `EbitenGame` (prod), `StubGame` (test)

## Writing New Tests

1. Use stub implementations from `*_test.go` files
2. Test against interfaces, not concrete types
3. Create entities with stub components
4. Use `NewStubGame` for game loop tests
5. No build tags needed - `*_test.go` suffix handles exclusion

## Integration Testing

For tests that need real Ebiten:
- Use production implementations
- Mark as integration tests with build tag if needed
- Most tests should use stubs for speed and simplicity
```

### Step 3.7: Update documentation (15 min)

**Update README.md:**
```markdown
## Testing

Run tests without any build flags:

```bash
go test ./...
```

See [Testing Guide](docs/TESTING.md) for details.
```

**Update DEVELOPMENT.md:**
Add section about interface-based testing pattern.

**Update API_REFERENCE.md:**
Update examples to show new type names and interfaces.

### Step 3.8: Final validation (10 min)

```bash
# Complete validation
echo "=== Build Validation ==="
go build ./... && echo "✅ Build successful" || echo "❌ Build failed"

echo "=== Test Validation ==="
go test ./... && echo "✅ Tests successful" || echo "❌ Tests failed"

echo "=== Vet Validation ==="
go vet ./... && echo "✅ Vet successful" || echo "❌ Vet failed"

echo "=== Build Tag Audit ==="
count=$(grep -r "//go:build test" pkg/engine --include="*.go" | grep -v "_test.go" | wc -l)
if [ $count -eq 0 ]; then
    echo "✅ No build tags in production code"
else
    echo "❌ Found $count build tags in production code"
fi

echo "=== Coverage Check ==="
go test -cover ./pkg/... | grep -E "coverage:|ok"
```

**Checkpoint:**
```bash
git add -A
git commit -m "docs: add testing guide and update documentation

- Created docs/TESTING.md with interface testing patterns
- Updated README.md with testing instructions
- Updated API_REFERENCE.md with new type names
- Documented all interface implementations

Part of build tag refactoring (Phase 3)"
```

---

## Final Steps

### Create Summary Report

**File:** `REFACTORING_COMPLETE.md`

```markdown
# Build Tag Refactoring - Completion Report

**Date:** [Date]
**Branch:** interfaces
**Status:** ✅ COMPLETE

## Changes Summary

### Interfaces Created
- GameRunner (game loop abstraction)
- Renderer/ImageProvider (rendering abstraction)
- SpriteProvider/InputProvider (component abstraction)
- RenderingSystem/UISystem (system abstraction)

### Types Migrated
- Game → EbitenGame/StubGame
- SpriteComponent → EbitenSprite/StubSprite
- InputComponent → EbitenInput/StubInput
- RenderSystem → EbitenRenderSystem/StubRenderSystem
- TutorialSystem → EbitenTutorialSystem/StubTutorialSystem
- HelpSystem → EbitenHelpSystem/StubHelpSystem
- 7 UI Systems → Ebiten*/Stub* implementations

### Files Deleted
- game_test_stub.go
- components_test_stub.go
- render_system_test_stub.go
- hud_system_test_stub.go
- menu_system_test_stub.go
- character_ui_test_stub.go
- skills_ui_test_stub.go
- map_ui_test_stub.go
- ui_systems_test_stub.go

### Build Tags Removed
- Removed from all production files in pkg/engine
- Zero `//go:build test` tags in non-test files

## Validation Results

```
✅ go build ./...
✅ go test ./...
✅ go vet ./...
✅ go test -race ./pkg/...
✅ Build tag audit: 0 tags in production code
```

## Coverage Comparison

Before: [baseline coverage]
After: [final coverage]
Change: [+/-]%

## Benefits Achieved

1. ✅ Standard Go patterns - no build tags
2. ✅ Build/test independence - both work without flags
3. ✅ Improved testability - easy to mock components
4. ✅ IDE support - no conflicting definitions
5. ✅ Tool compatibility - works with all Go tools

## Documentation

- docs/TESTING.md - Testing guide with examples
- QUICK_REFERENCE.md - Developer patterns
- Updated API_REFERENCE.md - New type names
- Updated README.md - Testing instructions

## Next Steps

1. Merge to main via PR
2. Update CI/CD to remove -tags test flags
3. Notify team of new patterns
4. Archive old documentation

---

**Refactoring successful.** All goals achieved.
```

### Commit and Push

```bash
# Final commit
git add REFACTORING_COMPLETE.md
git commit -m "docs: mark build tag refactoring as complete

All phases complete:
- 7 interfaces created
- 15+ types migrated to interface pattern
- 9 stub files deleted
- Zero build tags in production code
- All tests passing
- Documentation updated

Closes refactoring epic."

# Push branch
git push origin interfaces

# Create PR
gh pr create --title "Refactor: Replace build tags with interface-based dependency injection" \
  --body "See REFACTORING_SUMMARY.md for complete details.

This PR eliminates build tag conflicts by implementing standard Go interface patterns.

**Changes:**
- 7 new interfaces for dependency injection
- 15+ types migrated (Game, Components, Systems)
- Removed all build tag stubs
- Updated documentation

**Benefits:**
- ✅ Standard Go patterns
- ✅ No build tags needed
- ✅ Better testability
- ✅ Full tool support

**Validation:**
- All builds succeed
- All tests pass
- Coverage maintained
- No build tag conflicts"
```

---

## Troubleshooting

### Build Errors

**Error:** `undefined: TutorialSystem`
**Fix:** Ensure TutorialSystem has both production and test implementations, or use UISystem interface.

**Error:** `cannot use X as type Y`
**Fix:** Check interface implementation. Add missing methods.

**Error:** `multiple definitions of X`
**Fix:** Ensure old stub files are deleted. Check for duplicate types.

### Test Failures

**Error:** `cannot convert X to Y`
**Fix:** Use type assertion to interface: `comp.(SpriteProvider)` not `comp.(*SpriteComponent)`

**Error:** Tests work with -tags test but fail without
**Fix:** Tests should use stub implementations from *_test.go files, not production types.

### Coverage Issues

**Coverage drops:** Review which tests were lost. May need to port tests from stub files to new test files.

**Coverage won't measure:** Ensure package builds without -tags test flag.

---

## Success Criteria Checklist

- [ ] All interfaces defined in interfaces.go
- [ ] Game type migrated (EbitenGame/StubGame)
- [ ] SpriteComponent migrated (EbitenSprite/StubSprite)
- [ ] InputComponent migrated (EbitenInput/StubInput)
- [ ] RenderSystem migrated
- [ ] TutorialSystem migrated
- [ ] HelpSystem migrated
- [ ] All UI systems migrated (7 systems)
- [ ] All *_test_stub.go files deleted
- [ ] Zero build tags in pkg/engine production files
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] `go test -race ./pkg/...` passes
- [ ] Coverage >= baseline
- [ ] docs/TESTING.md created
- [ ] Documentation updated
- [ ] REFACTORING_COMPLETE.md created
- [ ] Changes committed and pushed
- [ ] PR created

---

## Estimated Timeline

- **Phase 2a:** 1 hour (interfaces)
- **Phase 2b:** 2 hours (Game type)
- **Phase 2c:** 2 hours (Components)
- **Phase 2d:** 3 hours (Core systems)
- **Phase 2e:** 3 hours (UI systems)
- **Phase 3:** 2 hours (Cleanup & validation)

**Total:** 13 hours

**Recommended approach:** Work in focused 2-3 hour blocks, commit frequently, test after each phase.

---

## Support

**Questions?** See:
- `QUICK_REFERENCE.md` - Developer patterns
- `INTERFACE_DESIGN.md` - Detailed architecture
- `BUILD_TAG_ISSUES.md` - Problem explanation
- `docs/TESTING.md` - Testing guide

**Issues?** 
- Revert to last working commit
- Check interface implementations with `var _ Interface = (*Type)(nil)`
- Use `go build -x` to see compilation details

---

**Ready to begin?** Start with Phase 2a: Create Core Interfaces.
