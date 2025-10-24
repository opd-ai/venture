# Interface Refactoring - Quick Reference

**For Developers: How to Use the New Interface Pattern**

## Quick Start

### Before Build Tag Pattern (❌ Broken)
```go
// game.go - //go:build !test
type Game struct { World *World }

// game_test_stub.go - //go:build test
type Game struct { World *World }  // Conflict!
```

### After Interface Pattern (✅ Standard Go)
```go
// interfaces.go
type GameRunner interface { Update() error }

// game.go
type EbitenGame struct { World *World }
func (g *EbitenGame) Update() error { /* real */ }

// game_test.go (automatically test-only)
type StubGame struct { World *World }
func (g *StubGame) Update() error { return nil }
```

## Core Interfaces Reference

### GameRunner
```go
type GameRunner interface {
    GetWorld() *World
    GetScreenSize() (width, height int)
    IsPaused() bool
    SetPaused(paused bool)
    SetPlayerEntity(entity *Entity)
    GetPlayerEntity() *Entity
    Update() error
}

// Production: EbitenGame
// Test: StubGame
```

### SpriteProvider
```go
type SpriteProvider interface {
    Component
    GetImage() ImageProvider
    GetSize() (width, height float64)
    GetColor() color.Color
    GetRotation() float64
    GetLayer() int
    IsVisible() bool
    SetVisible(visible bool)
}

// Production: EbitenSprite
// Test: StubSprite
```

### InputProvider
```go
type InputProvider interface {
    Component
    GetMovement() (x, y float64)
    IsActionPressed() bool
    IsActionJustPressed() bool
    IsUseItemPressed() bool
    IsUseItemJustPressed() bool
    IsSpellPressed(slot int) bool
    GetMousePosition() (x, y int)
    IsMousePressed() bool
}

// Production: EbitenInput
// Test: StubInput
```

### RenderingSystem
```go
type RenderingSystem interface {
    System  // Inherits Update(entities, deltaTime)
    Draw(renderer Renderer, entities []*Entity)
    SetShowColliders(show bool)
    SetShowGrid(show bool)
}

// Production: EbitenRenderSystem
// Test: StubRenderSystem
```

### UISystem
```go
type UISystem interface {
    System  // Inherits Update(entities, deltaTime)
    Draw(renderer Renderer)
    IsActive() bool
    SetActive(active bool)
    HandleInput(input InputProvider) bool
}

// Production: Ebiten*System (EbitenHUDSystem, etc.)
// Test: Stub*System (StubHUDSystem, etc.)
```

## Common Patterns

### Creating a Game Instance

**Production:**
```go
game := engine.NewEbitenGame(800, 600)
ebiten.RunGame(game)  // Implements ebiten.Game
```

**Tests:**
```go
game := &engine.StubGame{
    World: engine.NewWorld(),
}
game.Update()  // No Ebiten dependency
```

### Using Components

**Same pattern for both:**
```go
comp, ok := entity.GetComponent("sprite")
if ok {
    sprite := comp.(engine.SpriteProvider)
    w, h := sprite.GetSize()
    color := sprite.GetColor()
}
```

### Creating Systems

**Production:**
```go
renderSystem := engine.NewRenderSystem(cameraSystem)
world.AddSystem(renderSystem)
```

**Tests:**
```go
renderSystem := engine.NewStubRenderSystem()
world.AddSystem(renderSystem)
```

## Migration Checklist for Each Type

- [ ] Define interface in `interfaces.go`
- [ ] Rename production struct (e.g., `Foo` → `EbitenFoo`)
- [ ] Remove `//go:build !test` from production file
- [ ] Create test impl in `foo_test.go` (e.g., `StubFoo`)
- [ ] Update constructor to return interface
- [ ] Update all references to use interface type
- [ ] Delete `foo_test_stub.go` if it exists
- [ ] Run tests: `go test ./...`
- [ ] Run build: `go build ./...`
- [ ] Commit

## File Naming Convention

```
pkg/engine/
├── interfaces.go           # All interfaces (no build tags)
├── game.go                 # EbitenGame (no build tags)
├── game_test.go            # StubGame + tests (auto test-only)
├── sprite_component.go     # EbitenSprite (no build tags)
├── sprite_component_test.go # StubSprite + tests (auto test-only)
└── ...
```

**Key Rule:** `*_test.go` files are automatically excluded from production builds. No build tags needed.

## Testing Example

```go
// component_test.go

func TestSpriteComponent(t *testing.T) {
    // Create test entity
    entity := engine.NewEntity(1)
    
    // Create stub sprite
    sprite := &engine.StubSprite{
        Width: 32,
        Height: 32,
        Color: color.RGBA{255, 0, 0, 255},
        Visible: true,
    }
    
    // Add to entity (works same as production)
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

## Build Commands

**All commands work without build tags:**

```bash
# Build everything
go build ./...

# Test everything
go test ./...

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Vet
go vet ./...

# Run specific package
go test ./pkg/engine

# Run specific test
go test ./pkg/engine -run TestSpriteComponent
```

## Common Errors After Migration

### Error: "undefined: Game"
**Fix:** Use `EbitenGame` (production) or `StubGame` (test)

### Error: "cannot use X as type Y"
**Fix:** Make sure type implements the interface, check method signatures

### Error: "cannot convert X to Y"
**Fix:** Use type assertion to interface: `comp.(SpriteProvider)`

## Documentation References

- **BUILD_TAG_ISSUES.md** - Why we're doing this
- **REFACTORING_ANALYSIS.md** - What needs to change
- **INTERFACE_DESIGN.md** - How to change it (full details)
- **REFACTORING_SUMMARY.md** - Executive summary
- **REFACTORING_PROGRESS.md** - Current status

## Questions?

**Q: Why not keep build tags?**  
A: They create mutual exclusivity and break standard Go tooling. Interface pattern is idiomatic Go.

**Q: Will this hurt performance?**  
A: No. Interface calls in Go are very fast. Benchmarks show negligible overhead.

**Q: Do I need to change my code?**  
A: Minimal changes. Constructor names change, but usage stays mostly the same.

**Q: What about Ebiten dependencies?**  
A: Production code uses Ebiten normally. Interfaces abstract it for testing.

**Q: Can I mix production and test implementations?**  
A: Yes! That's the point. Tests can use real or stub implementations as needed.

---

**Status:** Design complete, ready for implementation  
**Estimated effort:** 13 hours total  
**Current phase:** Phase 0 complete, Phase 2 ready to start
