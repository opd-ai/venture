# Testing Guide

## Overview

Venture uses interface-based dependency injection for comprehensive testability without external dependencies. Production code uses Ebiten implementations for graphics and input, while tests use lightweight stub implementations that require no GUI or X11 libraries.

This architecture enables:
- ✅ Fast test execution (no window initialization)
- ✅ Headless CI/CD environments
- ✅ Standard Go testing without build tags
- ✅ Easy mocking and behavior verification
- ✅ Deterministic test outcomes

## Test Structure

### Component Testing

Components use the interface pattern to enable testing without Ebiten dependencies:

```go
func TestSpriteComponent(t *testing.T) {
    // Use stub implementation in tests
    sprite := &StubSprite{
        Width:    32,
        Height:   32,
        Color:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
        Visible:  true,
    }
    
    // Test interface methods
    if w, h := sprite.GetSize(); w != 32 || h != 32 {
        t.Errorf("expected 32x32, got %vx%v", w, h)
    }
    
    // Test state changes
    sprite.SetVisible(false)
    if sprite.IsVisible() {
        t.Error("sprite should be invisible")
    }
}
```

### System Testing

Systems receive interfaces, enabling test-time dependency injection:

```go
func TestRenderSystem(t *testing.T) {
    world := NewWorld()
    
    // Create stub render system (no Ebiten required)
    renderSystem := &StubRenderSystem{}
    world.AddSystem(renderSystem)
    
    // Create entity with stub sprite
    entity := world.CreateEntity()
    entity.AddComponent(&PositionComponent{X: 100, Y: 100})
    entity.AddComponent(&StubSprite{Width: 32, Height: 32})
    
    // Update system
    world.Update(0.016)
    
    // Verify behavior (stub can track calls)
    if renderSystem.UpdateCallCount != 1 {
        t.Errorf("expected 1 update, got %d", renderSystem.UpdateCallCount)
    }
}
```

### Game Loop Testing

The `GameRunner` interface enables testing game state without running Ebiten:

```go
func TestGameLoop(t *testing.T) {
    // Create stub game (no window)
    game := NewStubGame(800, 600)
    
    // Set up test scenario
    player := game.GetWorld().CreateEntity()
    player.AddComponent(&PositionComponent{X: 100, Y: 100})
    game.SetPlayerEntity(player)
    
    // Run multiple frames
    for i := 0; i < 60; i++ {
        if err := game.Update(); err != nil {
            t.Fatalf("update failed: %v", err)
        }
    }
    
    // Verify state after 60 frames
    if game.UpdateCalls != 60 {
        t.Errorf("expected 60 updates, got %d", game.UpdateCalls)
    }
}
```

### Input Testing

Input components use stubs to simulate player actions deterministically:

```go
func TestPlayerMovement(t *testing.T) {
    world := NewWorld()
    world.AddSystem(NewMovementSystem())
    
    player := world.CreateEntity()
    player.AddComponent(&PositionComponent{X: 0, Y: 0})
    
    // Create stub input with simulated movement
    input := &StubInput{
        MoveX: 1.0,  // Move right
        MoveY: 0.0,
    }
    player.AddComponent(input)
    player.AddComponent(&VelocityComponent{MaxSpeed: 100})
    
    // Update for one frame (0.016s = 60 FPS)
    world.Update(0.016)
    
    // Verify player moved right
    pos := player.GetComponent("position").(*PositionComponent)
    if pos.X <= 0 {
        t.Error("player should have moved right")
    }
}
```

## Running Tests

### Basic Testing

```bash
# Run all tests (no build tags needed!)
go test ./...

# Test specific package
go test ./pkg/engine

# Test with verbose output
go test -v ./pkg/engine

# Test specific function
go test -v ./pkg/engine -run TestPlayerCombat
```

### Coverage Analysis

```bash
# Generate coverage report
go test -cover ./...

# Detailed coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Race Detection

```bash
# Detect data races (important for concurrent systems)
go test -race ./...

# Race detection for specific package
go test -race ./pkg/engine
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. ./...

# Benchmarks with memory stats
go test -bench=. -benchmem ./...

# Specific benchmark
go test -bench=BenchmarkWorldUpdate ./pkg/engine
```

## Interface Implementations

### Components

| Interface | Production | Test | Description |
|-----------|-----------|------|-------------|
| `SpriteProvider` | `EbitenSprite` | `StubSprite` | Visual sprite data |
| `InputProvider` | `EbitenInput` | `StubInput` | Player input state |
| `ClientConnection` | `TCPClient` | `MockClient` | Network client operations |
| `ServerConnection` | `TCPServer` | `MockServer` | Network server operations |

### Systems

| Interface | Production | Test | Description |
|-----------|-----------|------|-------------|
| `RenderingSystem` | `EbitenRenderSystem` | `StubRenderSystem` | Entity rendering |
| `UISystem` | `Ebiten*System` | `Stub*System` | UI rendering (HUD, Menu, etc.) |

### Game Runner

| Interface | Production | Test | Description |
|-----------|-----------|------|-------------|
| `GameRunner` | `EbitenGame` | `StubGame` | Main game loop |

### Complete Mapping

**Production (Ebiten-based):**
- `EbitenGame` - Game loop with Ebiten integration
- `EbitenSprite` - Sprite with `*ebiten.Image`
- `EbitenInput` - Input from keyboard/mouse
- `EbitenRenderSystem` - Renders to `*ebiten.Image`
- `EbitenHUDSystem`, `EbitenMenuSystem`, etc. - UI systems with Ebiten

**Test (Stub implementations):**
- `StubGame` - Game loop without window
- `StubSprite` - Sprite without image data
- `StubInput` - Controllable input state
- `StubRenderSystem` - Render system without drawing
- `StubHUDSystem`, `StubMenuSystem`, etc. - UI systems without rendering

## Writing New Tests

### 1. Test Against Interfaces

**Do:**
```go
func TestSystem(t *testing.T) {
    var sprite SpriteProvider = &StubSprite{}
    // Use interface methods
    sprite.SetVisible(false)
}
```

**Don't:**
```go
func TestSystem(t *testing.T) {
    // Avoid concrete types in test signatures
    sprite := &EbitenSprite{}  // Requires Ebiten!
}
```

### 2. Use Stub Implementations

**Do:**
```go
func TestCombat(t *testing.T) {
    player := world.CreateEntity()
    player.AddComponent(&StubInput{ActionPressed: true})
    player.AddComponent(&StubSprite{Width: 32, Height: 32})
}
```

**Don't:**
```go
func TestCombat(t *testing.T) {
    // Avoid production types in tests
    player.AddComponent(&EbitenInput{})  // Needs Ebiten initialization
}
```

### 3. Create Entities with Stub Components

```go
func setupTestEntity(world *World) *Entity {
    entity := world.CreateEntity()
    entity.AddComponent(&PositionComponent{X: 100, Y: 100})
    entity.AddComponent(&VelocityComponent{X: 10, Y: 0, MaxSpeed: 100})
    entity.AddComponent(&StubSprite{Width: 32, Height: 32})
    entity.AddComponent(&StubInput{})
    return entity
}

func TestEntityBehavior(t *testing.T) {
    world := NewWorld()
    entity := setupTestEntity(world)
    // Test entity behavior
}
```

### 4. Use Table-Driven Tests

```go
func TestDamageCalculation(t *testing.T) {
    tests := []struct {
        name           string
        attackDamage   int
        targetArmor    int
        expectedDamage int
    }{
        {"no armor", 100, 0, 100},
        {"half armor", 100, 50, 50},
        {"full armor", 100, 100, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := calculateDamage(tt.attackDamage, tt.targetArmor)
            if result != tt.expectedDamage {
                t.Errorf("expected %d, got %d", tt.expectedDamage, result)
            }
        })
    }
}
```

### 5. Test Deterministic Generation

All procedural generation should be deterministic:

```go
func TestDeterministicGeneration(t *testing.T) {
    seed := int64(12345)
    params := procgen.GenerationParams{
        Difficulty: 0.5,
        Depth:      5,
        GenreID:    "fantasy",
    }
    
    // Generate twice with same seed
    gen := terrain.NewGenerator()
    result1, _ := gen.Generate(seed, params)
    result2, _ := gen.Generate(seed, params)
    
    // Results must be identical
    terrain1 := result1.(*terrain.Terrain)
    terrain2 := result2.(*terrain.Terrain)
    
    if !terrainEquals(terrain1, terrain2) {
        t.Error("same seed produced different terrain")
    }
}
```

## Integration Testing

For tests requiring real Ebiten integration:

```go
//go:build integration
// +build integration

package engine_test

import (
    "testing"
    "github.com/hajimehoshi/ebiten/v2"
)

func TestRealRendering(t *testing.T) {
    // Use actual Ebiten types
    game := NewEbitenGame(800, 600)
    screen := ebiten.NewImage(800, 600)
    
    // Test real rendering
    game.Draw(screen)
}
```

Run integration tests separately:
```bash
go test -tags integration ./...
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      # No X11 or Ebiten dependencies needed!
      - name: Run tests
        run: go test ./...
      
      - name: Race detection
        run: go test -race ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

## Common Testing Patterns

### Testing System Interactions

```go
func TestCombatSystem(t *testing.T) {
    world := NewWorld()
    world.AddSystem(NewCombatSystem())
    
    attacker := world.CreateEntity()
    attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
    attacker.AddComponent(&AttackComponent{Damage: 50, Range: 100})
    
    target := world.CreateEntity()
    target.AddComponent(&PositionComponent{X: 50, Y: 0})
    target.AddComponent(&HealthComponent{Current: 100, Max: 100})
    
    // Trigger attack
    world.Update(0.016)
    
    // Verify damage applied
    health := target.GetComponent("health").(*HealthComponent)
    if health.Current != 50 {
        t.Errorf("expected 50 health, got %d", health.Current)
    }
}
```

### Testing Event Systems

```go
func TestDamageEvent(t *testing.T) {
    var eventReceived bool
    var damageAmount int
    
    world := NewWorld()
    world.RegisterEventHandler("damage", func(e Event) {
        eventReceived = true
        damageAmount = e.Data.(int)
    })
    
    world.EmitEvent(Event{Type: "damage", Data: 25})
    
    if !eventReceived {
        t.Error("event not received")
    }
    if damageAmount != 25 {
        t.Errorf("expected damage 25, got %d", damageAmount)
    }
}
```

### Testing Inventory System

```go
func TestInventoryAddItem(t *testing.T) {
    inventory := NewInventoryComponent(10, 100.0)
    
    item := &ItemData{
        ID:     "potion",
        Name:   "Health Potion",
        Weight: 1.0,
        Type:   item.TypeConsumable,
    }
    
    if !inventory.AddItem(item) {
        t.Error("failed to add item to empty inventory")
    }
    
    if inventory.GetItemCount() != 1 {
        t.Errorf("expected 1 item, got %d", inventory.GetItemCount())
    }
}
```

### Testing Network Communication

The network package uses mock implementations to test client-server communication without real network I/O:

```go
func TestClientServerCommunication(t *testing.T) {
    // Create mock client and server
    client := network.NewMockClient()
    server := network.NewMockServer()
    
    // Simulate server starting
    if err := server.Start(); err != nil {
        t.Fatalf("server start failed: %v", err)
    }
    
    // Simulate client connecting
    if err := client.Connect(); err != nil {
        t.Fatalf("client connect failed: %v", err)
    }
    
    playerID := uint64(123)
    client.SetPlayerID(playerID)
    
    // Simulate player joining server
    server.SimulatePlayerJoin(playerID)
    
    // Test sending input from client
    err := client.SendInput("move", []byte{1, 0})
    if err != nil {
        t.Errorf("send input failed: %v", err)
    }
    
    // Verify input was recorded
    if client.GetSentInputCount() != 1 {
        t.Errorf("expected 1 sent input, got %d", client.GetSentInputCount())
    }
    
    // Test server broadcasting state
    update := &network.StateUpdate{
        Sequence:  1,
        Timestamp: 1000,
    }
    server.BroadcastStateUpdate(update)
    
    // Verify broadcast was recorded
    if server.BroadcastCalls != 1 {
        t.Errorf("expected 1 broadcast, got %d", server.BroadcastCalls)
    }
}
```

**Testing Network Errors:**

```go
func TestNetworkErrorHandling(t *testing.T) {
    client := network.NewMockClient()
    
    // Configure mock to return error
    client.ConnectError = fmt.Errorf("connection refused")
    
    // Attempt connection
    err := client.Connect()
    if err == nil {
        t.Error("expected connection error")
    }
    
    // Verify error message
    if !strings.Contains(err.Error(), "connection refused") {
        t.Errorf("unexpected error: %v", err)
    }
}
```

**Testing Latency Simulation:**

```go
func TestNetworkLatency(t *testing.T) {
    client := network.NewMockClient()
    
    // Set high latency
    client.SetLatency(500 * time.Millisecond)
    
    if client.GetLatency() != 500*time.Millisecond {
        t.Errorf("expected 500ms latency, got %v", client.GetLatency())
    }
    
    // Test behavior under high latency conditions
    // ... your latency-dependent logic here
}
```

**Testing Server Player Management:**

```go
func TestServerPlayerManagement(t *testing.T) {
    server := network.NewMockServer()
    server.Start()
    
    // Simulate multiple players joining
    server.SimulatePlayerJoin(1)
    server.SimulatePlayerJoin(2)
    server.SimulatePlayerJoin(3)
    
    // Verify player count
    if count := server.GetPlayerCount(); count != 3 {
        t.Errorf("expected 3 players, got %d", count)
    }
    
    // Get player list
    players := server.GetPlayers()
    if len(players) != 3 {
        t.Errorf("expected 3 players in list, got %d", len(players))
    }
    
    // Simulate player leaving
    server.SimulatePlayerLeave(2)
    
    if count := server.GetPlayerCount(); count != 2 {
        t.Errorf("expected 2 players after leave, got %d", count)
    }
}
```

**Resetting Mocks Between Tests:**

```go
func TestMultipleScenarios(t *testing.T) {
    client := network.NewMockClient()
    
    t.Run("scenario 1", func(t *testing.T) {
        client.Connect()
        client.SendInput("action", []byte{1})
        // ... test logic
        client.Reset() // Clean state for next test
    })
    
    t.Run("scenario 2", func(t *testing.T) {
        // Client is clean, no previous state
        if client.IsConnected() {
            t.Error("client should not be connected after reset")
        }
        if client.GetSentInputCount() != 0 {
            t.Error("sent input count should be 0 after reset")
        }
    })
}
```

## Test Organization

```
pkg/engine/
├── game.go                    # EbitenGame (production)
├── game_test.go              # StubGame + tests
├── sprite_component.go       # EbitenSprite (production)
├── sprite_component_test.go  # StubSprite + tests
├── input_component.go        # EbitenInput (production)
├── input_component_test.go   # StubInput + tests
└── ...
```

**Conventions:**
- Production code in `*.go` (no build tags)
- Test stubs and tests in `*_test.go` (automatically excluded from builds)
- Integration tests in `*_integration_test.go` with `//go:build integration`

## Performance Testing

### Benchmark Example

```go
func BenchmarkWorldUpdate(b *testing.B) {
    world := NewWorld()
    world.AddSystem(NewMovementSystem())
    world.AddSystem(NewCombatSystem())
    
    // Create 1000 entities
    for i := 0; i < 1000; i++ {
        entity := world.CreateEntity()
        entity.AddComponent(&PositionComponent{X: float64(i), Y: 0})
        entity.AddComponent(&VelocityComponent{X: 1, Y: 0, MaxSpeed: 100})
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        world.Update(0.016)
    }
}
```

**Performance Targets:**
- World update with 1000 entities: <1ms
- Terrain generation: <500ms
- Entity generation: <10ms per entity
- 60 FPS sustained with 2000+ entities

## Debugging Tests

### Enable Verbose Output

```bash
go test -v ./pkg/engine
```

### Run Single Test

```bash
go test -v ./pkg/engine -run TestCombatSystem
```

### Print Debug Information

```go
func TestDebug(t *testing.T) {
    world := NewWorld()
    
    // Use t.Log for debug output (shown with -v)
    t.Logf("World has %d entities", len(world.GetEntities()))
    
    // Use t.Helper() for helper functions
    assertEqual := func(expected, actual int) {
        t.Helper()
        if expected != actual {
            t.Errorf("expected %d, got %d", expected, actual)
        }
    }
    
    assertEqual(0, len(world.GetEntities()))
}
```

### Test Failures with Context

```go
func TestWithContext(t *testing.T) {
    for i, testCase := range testCases {
        result := processInput(testCase.input)
        if result != testCase.expected {
            t.Errorf("test case %d failed:\n  input: %v\n  expected: %v\n  got: %v", 
                i, testCase.input, testCase.expected, result)
        }
    }
}
```

## Best Practices

### ✅ Do

1. **Test interfaces, not implementations**
2. **Use stub implementations in all tests**
3. **Write table-driven tests for multiple scenarios**
4. **Test error paths and edge cases**
5. **Verify determinism in generation**
6. **Use meaningful test names**
7. **Keep tests focused and small**
8. **Use setup/teardown helpers**
9. **Run race detector regularly**
10. **Maintain test coverage above 80%**

### ❌ Don't

1. **Don't use production Ebiten types in tests**
2. **Don't test implementation details**
3. **Don't write flaky tests**
4. **Don't skip error checking in tests**
5. **Don't use sleep() for timing**
6. **Don't test external dependencies directly**
7. **Don't write tests that depend on execution order**
8. **Don't ignore race warnings**
9. **Don't commit commented-out tests**
10. **Don't use magic numbers without explanation**

## Coverage Goals

| Package | Target | Current | Notes |
|---------|--------|---------|-------|
| pkg/engine | 80%+ | 70.7% | UI systems need integration tests |
| pkg/procgen/* | 90%+ | 100% | Excellent coverage |
| pkg/rendering/* | 90%+ | 92-100% | Excellent coverage |
| pkg/audio/* | 90%+ | 85-100% | Good coverage |
| pkg/network | 70%+ | 54.1% | Mock implementations added, coverage will improve as used |
| pkg/combat | 90%+ | 100% | Excellent coverage |
| pkg/world | 90%+ | 100% | Excellent coverage |

**Focus areas for improvement:**
- `pkg/engine`: UI system coverage (needs integration tests)
- `pkg/network`: Write tests using new MockClient/MockServer to improve from 54.1% to 70%+

## Troubleshooting

### Build Errors

**Problem:** `undefined: ebiten.Image`  
**Solution:** Use stub types (`StubSprite`) instead of production types (`EbitenSprite`)

**Problem:** `multiple definitions of X`  
**Solution:** Build tags are deprecated - ensure using new interface pattern

### Test Failures

**Problem:** `panic: X window required`  
**Solution:** Tests should not require GUI - use stub implementations

**Problem:** Tests flaky/non-deterministic  
**Solution:** Ensure all RNG uses seeds, avoid time-based logic

### Coverage Issues

**Problem:** Coverage not measuring  
**Solution:** Ensure package compiles with `go test` (no build tags)

**Problem:** Low coverage reported  
**Solution:** Check if tests actually exercise code paths

## Resources

- **Architecture**: See `docs/ARCHITECTURE.md`
- **Interface Reference**: See `QUICK_REFERENCE.md`
- **Refactoring Details**: See `REFACTORING_COMPLETE.md`
- **Development Guide**: See `docs/DEVELOPMENT.md`
- **Contributing**: See `docs/CONTRIBUTING.md`

## Summary

Venture's testing infrastructure is built on **interface-based dependency injection**, enabling:
- Fast, reliable tests without GUI dependencies
- Comprehensive coverage across all systems
- Deterministic test execution
- Easy CI/CD integration
- Clear separation between production and test code

**Key Principle:** Test against interfaces, implement with stubs, verify behavior—not implementation.

Happy testing! 🧪✅
