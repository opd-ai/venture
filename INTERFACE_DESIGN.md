# Interface-Based Architecture Design

**Date:** October 24, 2025  
**Branch:** interfaces  
**Goal:** Replace build tag system with interface-based dependency injection

## Design Philosophy

1. **Production code depends on interfaces, not concrete types**
2. **Interfaces live in regular .go files (no build tags)**
3. **Production implementations in regular .go files (no build tags)**
4. **Test implementations in *_test.go files (automatically test-only)**
5. **No build tags anywhere** (except documented exceptions)

## Interface Hierarchy

### Core Interfaces

#### 1. GameRunner Interface

**Purpose:** Abstract the game loop from Ebiten specifics

```go
// File: pkg/engine/interfaces.go

// GameRunner manages the main game loop and state.
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
}
```

**Implementations:**
- **Production:** `EbitenGame` (implements `ebiten.Game` + `GameRunner`)
- **Test:** `StubGame` (implements `GameRunner` only)

#### 2. Renderer Interface

**Purpose:** Abstract rendering from Ebiten image types

```go
// Renderer handles drawing of visual elements
type Renderer interface {
    // DrawSprite draws a sprite at the given position
    DrawSprite(sprite SpriteProvider, x, y float64, opts *DrawOptions)
    
    // DrawRect draws a filled rectangle
    DrawRect(x, y, width, height float64, color color.Color)
    
    // DrawText draws text at the given position
    DrawText(text string, x, y int, color color.Color)
    
    // Clear clears the screen with the given color
    Clear(color color.Color)
    
    // GetBounds returns the rendering bounds
    GetBounds() (width, height int)
}

// DrawOptions contains optional rendering parameters
type DrawOptions struct {
    Rotation float64
    ScaleX, ScaleY float64
    Alpha float32
}
```

**Implementations:**
- **Production:** `EbitenRenderer` (wraps `*ebiten.Image`)
- **Test:** `StubRenderer` (no-op or records draw calls for verification)

#### 3. ImageProvider Interface

**Purpose:** Abstract image handling from Ebiten

```go
// ImageProvider provides access to image data
type ImageProvider interface {
    // GetSize returns the image dimensions
    GetSize() (width, height int)
    
    // GetPixel returns the color at the given position
    GetPixel(x, y int) color.Color
}
```

**Implementations:**
- **Production:** `EbitenImage` (wraps `*ebiten.Image`)
- **Test:** `StubImage` (simple in-memory representation)

#### 4. SpriteProvider Interface

**Purpose:** Abstract sprite components from Ebiten dependencies

```go
// SpriteProvider provides sprite visual data
type SpriteProvider interface {
    Component  // Inherits Type() string
    
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
}
```

**Implementations:**
- **Production:** `EbitenSprite` (has `*ebiten.Image`)
- **Test:** `StubSprite` (just properties, no actual image)

#### 5. InputProvider Interface

**Purpose:** Abstract input handling

```go
// InputProvider provides input state
type InputProvider interface {
    Component  // Inherits Type() string
    
    // GetMovement returns movement input (-1.0 to 1.0)
    GetMovement() (x, y float64)
    
    // IsActionPressed returns whether the action button is pressed
    IsActionPressed() bool
    
    // IsActionJustPressed returns whether action was just pressed this frame
    IsActionJustPressed() bool
    
    // IsUseItemPressed returns whether the use item button is pressed
    IsUseItemPressed() bool
    
    // IsUseItemJustPressed returns whether use item was just pressed this frame
    IsUseItemJustPressed() bool
    
    // IsSpellPressed returns whether a spell hotkey (1-5) is pressed
    IsSpellPressed(slot int) bool
    
    // GetMousePosition returns mouse coordinates
    GetMousePosition() (x, y int)
    
    // IsMousePressed returns whether mouse button is pressed
    IsMousePressed() bool
}
```

**Implementations:**
- **Production:** `EbitenInput` (reads from `ebiten.IsKeyPressed`)
- **Test:** `StubInput` (controllable state for testing)

### System Interfaces

#### 6. System Interface (Already Exists)

The existing `System` interface is good:

```go
type System interface {
    Update(entities []*Entity, deltaTime float64)
}
```

**No changes needed** - systems already use interfaces.

#### 7. RenderSystem Interface

**Purpose:** Define rendering system contract

```go
// RenderingSystem handles visual rendering of entities
type RenderingSystem interface {
    System  // Inherits Update()
    
    // Draw renders all entities to the renderer
    Draw(renderer Renderer, entities []*Entity)
    
    // SetShowColliders enables/disables collider visualization
    SetShowColliders(show bool)
    
    // SetShowGrid enables/disables grid visualization
    SetShowGrid(show bool)
}
```

**Implementations:**
- **Production:** `EbitenRenderSystem`
- **Test:** `StubRenderSystem`

#### 8. UISystem Interface

**Purpose:** Define UI system contract for HUD, menus, etc.

```go
// UISystem handles user interface rendering and interaction
type UISystem interface {
    System  // Inherits Update()
    
    // Draw renders the UI to the renderer
    Draw(renderer Renderer)
    
    // IsActive returns whether the UI is currently visible
    IsActive() bool
    
    // SetActive sets the UI visibility
    SetActive(active bool)
    
    // HandleInput processes input for the UI (returns true if input was consumed)
    HandleInput(input InputProvider) bool
}
```

**Implementations:**
- **Production versions:** `EbitenHUDSystem`, `EbitenMenuSystem`, etc.
- **Test versions:** `StubHUDSystem`, `StubMenuSystem`, etc.

## Migration Strategy

### Step 1: Create Interfaces

Create `pkg/engine/interfaces.go` with all interface definitions.

### Step 2: Migrate Game Type

**Before:**
```go
// game.go - //go:build !test
type Game struct { ... }

// game_test_stub.go - //go:build test
type Game struct { ... }
```

**After:**
```go
// interfaces.go - No build tags
type GameRunner interface { ... }

// game.go - No build tags
type EbitenGame struct {
    World *World
    screen *ebiten.Image
    // ... other fields
}

func (g *EbitenGame) Update() error { ... }
// Implements GameRunner interface

// game_test.go - No build tags, *_test.go suffix excludes from production
type StubGame struct {
    World *World
    // ... minimal fields
}

func (g *StubGame) Update() error { return nil }
// Implements GameRunner interface
```

**Update Dependencies:**
```go
// Before
func NewSomeSystem(game *Game) *SomeSystem { ... }

// After
func NewSomeSystem(game GameRunner) *SomeSystem { ... }
```

### Step 3: Migrate Components

**Before:**
```go
// components.go - //go:build !test
type SpriteComponent struct {
    Image *ebiten.Image  // Ebiten dependency
    // ...
}

// components_test_stub.go - //go:build test
type SpriteComponent struct {
    // No Image field
    // ...
}
```

**After:**
```go
// interfaces.go - No build tags
type SpriteProvider interface { ... }

// sprite_component.go - No build tags
type EbitenSprite struct {
    image *ebiten.Image
    width, height float64
    color color.Color
    // ...
}

func (s *EbitenSprite) GetImage() ImageProvider { 
    return &EbitenImage{image: s.image} 
}
// Implements SpriteProvider interface

// sprite_component_test.go - No build tags
type StubSprite struct {
    width, height float64
    color color.Color
    visible bool
    // ...
}

func (s *StubSprite) GetImage() ImageProvider { return nil }
// Implements SpriteProvider interface
```

**Update Entity Component Access:**
```go
// Before
comp, ok := entity.GetComponent("sprite")
sprite := comp.(*SpriteComponent)

// After - Same pattern, just different concrete types
comp, ok := entity.GetComponent("sprite")
sprite := comp.(SpriteProvider)  // Use interface
```

### Step 4: Migrate Systems

For each system with build tags:

1. Define interface in `interfaces.go` (if needed beyond System interface)
2. Rename production struct (e.g., `RenderSystem` → `EbitenRenderSystem`)
3. Remove `//go:build !test` from production file
4. Create test implementation in `*_test.go` file (no build tags)
5. Update constructors to return interfaces

**Example - RenderSystem:**

```go
// interfaces.go
type RenderingSystem interface {
    System
    Draw(renderer Renderer, entities []*Entity)
    SetShowColliders(show bool)
    SetShowGrid(show bool)
}

// render_system.go - No build tags
type EbitenRenderSystem struct {
    cameraSystem *CameraSystem
    showColliders bool
    showGrid bool
}

func NewRenderSystem(cameraSystem *CameraSystem) RenderingSystem {
    return &EbitenRenderSystem{
        cameraSystem: cameraSystem,
    }
}

// render_system_test.go - No build tags
type StubRenderSystem struct {
    ShowColliders bool
    ShowGrid bool
}

func NewStubRenderSystem() RenderingSystem {
    return &StubRenderSystem{}
}

func (r *StubRenderSystem) Update(entities []*Entity, deltaTime float64) {}
func (r *StubRenderSystem) Draw(renderer Renderer, entities []*Entity) {}
func (r *StubRenderSystem) SetShowColliders(show bool) { r.ShowColliders = show }
func (r *StubRenderSystem) SetShowGrid(show bool) { r.ShowGrid = show }
```

### Step 5: Remove Build Tags

After all types are migrated:

1. Delete all `*_test_stub.go` files
2. Remove all `//go:build` directives
3. Verify builds: `go build ./...`
4. Verify tests: `go test ./...`

## File Organization

```
pkg/engine/
├── interfaces.go              # All interfaces (no build tags)
│
├── game.go                    # EbitenGame implementation (no build tags)
├── game_test.go               # StubGame + tests (auto-excluded by *_test.go)
│
├── components.go              # Component base types (no build tags)
├── sprite_component.go        # EbitenSprite (no build tags)
├── sprite_component_test.go   # StubSprite + tests (auto-excluded)
├── input_component.go         # EbitenInput (no build tags)
├── input_component_test.go    # StubInput + tests (auto-excluded)
│
├── render_system.go           # EbitenRenderSystem (no build tags)
├── render_system_test.go      # StubRenderSystem + tests (auto-excluded)
│
├── hud_system.go              # EbitenHUDSystem (no build tags)
├── hud_system_test.go         # StubHUDSystem + tests (auto-excluded)
│
└── ... (similar pattern for all systems)
```

**Key Benefits:**
- No build tags anywhere
- Production and test code can coexist
- `*_test.go` suffix provides automatic test-only behavior
- Standard Go patterns, works with all tools

## Testing Strategy

### Unit Tests

```go
// render_system_test.go

func TestRenderSystem(t *testing.T) {
    // Create stub renderer to verify draw calls
    renderer := &StubRenderer{}
    
    // Create stub camera system
    camera := &StubCameraSystem{}
    
    // Create system under test
    system := NewRenderSystem(camera)
    
    // Create test entity with stub sprite
    entity := NewEntity(1)
    sprite := &StubSprite{
        width: 32,
        height: 32,
        color: color.RGBA{255, 0, 0, 255},
        visible: true,
    }
    entity.AddComponent(sprite)
    
    // Test draw
    system.Draw(renderer, []*Entity{entity})
    
    // Verify renderer was called
    if len(renderer.DrawCalls) != 1 {
        t.Errorf("Expected 1 draw call, got %d", len(renderer.DrawCalls))
    }
}
```

### Integration Tests

Integration tests can use real implementations or stubs as needed:

```go
// integration_test.go

func TestGameLoop(t *testing.T) {
    // Use stub game for test
    game := &StubGame{
        World: NewWorld(),
    }
    
    // Add real or stub systems as needed
    game.World.AddSystem(NewStubRenderSystem())
    
    // Test game loop
    err := game.Update()
    if err != nil {
        t.Fatal(err)
    }
}
```

## Compatibility Notes

### Ebiten Integration

Production code continues to use Ebiten normally:

```go
// main.go
func main() {
    game := engine.NewEbitenGame(800, 600)
    
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowTitle("Venture")
    
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
```

`EbitenGame` implements `ebiten.Game` interface:
```go
type EbitenGame struct {
    // ...
}

// Implements ebiten.Game
func (g *EbitenGame) Update() error { ... }
func (g *EbitenGame) Draw(screen *ebiten.Image) { ... }
func (g *EbitenGame) Layout(outsideWidth, outsideHeight int) (int, int) { ... }
```

### Backward Compatibility

External code that depends on specific types needs minimal changes:

```go
// Before
game := engine.NewGame(800, 600)  // Returns *Game

// After
game := engine.NewEbitenGame(800, 600)  // Returns *EbitenGame (implements GameRunner)
```

Most code using interfaces already doesn't need changes.

## Benefits Summary

1. ✅ **No build tags** - Standard Go compilation
2. ✅ **Testability** - Easy to mock/stub any component
3. ✅ **Clarity** - Clear separation of interfaces and implementations
4. ✅ **Flexibility** - Can swap implementations without changing dependencies
5. ✅ **IDE Support** - No conflicting type definitions
6. ✅ **Tool Compatibility** - Works with all Go tools (go test, go build, gopls, etc.)
7. ✅ **Performance** - Minimal overhead (interface calls are fast in Go)
8. ✅ **Maintainability** - Easier to understand and modify

## Migration Checklist

- [ ] Create `interfaces.go` with all interface definitions
- [ ] Migrate `Game` type (highest priority)
- [ ] Migrate `SpriteComponent` and `InputComponent`
- [ ] Migrate `RenderSystem`
- [ ] Migrate UI systems (HUD, Menu, Character, Skills, Map, Inventory, Quest)
- [ ] Migrate remaining systems with build tags
- [ ] Remove all `*_test_stub.go` files
- [ ] Remove all `//go:build` directives
- [ ] Update documentation
- [ ] Verify all builds pass
- [ ] Verify all tests pass
- [ ] Measure test coverage

## Timeline Estimate

**Phase 1: Design & Interfaces** - 1-2 hours
- Create interfaces.go
- Document interface contracts

**Phase 2: Core Migration** - 4-6 hours
- Game type (1-2 hours)
- Components (1-2 hours)
- Core systems (2 hours)

**Phase 3: System Migration** - 6-8 hours
- UI systems (3-4 hours)
- Remaining systems (3-4 hours)

**Phase 4: Cleanup & Validation** - 2-3 hours
- Remove build tags
- Delete stub files
- Fix any remaining issues
- Documentation

**Total:** 13-19 hours of focused work

**Risk Mitigation:**
- Migrate one type at a time
- Commit after each successful migration
- Run tests frequently
- Keep old code commented out temporarily
