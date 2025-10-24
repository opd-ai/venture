//go:build test
// +build test

// Package engine provides tests for proper system initialization.
// This file tests that systems are correctly initialized with required
// parameters and validates against common initialization errors.
package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestCollisionSystemRequiresConstructor verifies that CollisionSystem
// must be initialized with NewCollisionSystem constructor.
func TestCollisionSystemRequiresConstructor(t *testing.T) {
	cellSize := 64.0
	sys := NewCollisionSystem(cellSize)

	if sys == nil {
		t.Fatal("NewCollisionSystem returned nil")
	}

	if sys.CellSize != cellSize {
		t.Errorf("CellSize = %f, want %f", sys.CellSize, cellSize)
	}

	if sys.grid == nil {
		t.Error("grid map not initialized")
	}
}

// TestCollisionSystemInvalidInitialization tests that direct struct
// instantiation produces invalid configuration.
func TestCollisionSystemInvalidInitialization(t *testing.T) {
	// Simulate the bug: direct struct instantiation
	sys := &CollisionSystem{}

	// Verify that this produces invalid state
	if sys.CellSize != 0.0 {
		t.Errorf("Expected CellSize=0.0 from direct instantiation, got %f", sys.CellSize)
	}

	// This is the state that causes division by zero errors
	t.Logf("WARNING: Direct instantiation produces CellSize=0, causing collision detection failure")
}

// TestCollisionSystemTerrainIntegration tests that properly initialized
// collision system can detect terrain collisions.
func TestCollisionSystemTerrainIntegration(t *testing.T) {
	world := NewWorld()
	sys := NewCollisionSystem(32.0) // Proper initialization

	// Create simple terrain with walls
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if x == 0 || x == 9 || y == 0 || y == 9 {
				testTerrain.SetTile(x, y, terrain.TileWall)
			} else {
				testTerrain.SetTile(x, y, terrain.TileFloor)
			}
		}
	}

	// Setup terrain checker
	terrainChecker := NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(testTerrain)
	sys.SetTerrainChecker(terrainChecker)

	// Create entity at center (should be on floor)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 160, Y: 160}) // Tile (5, 5)
	entity.AddComponent(&VelocityComponent{VX: -100, VY: 0})
	entity.AddComponent(&ColliderComponent{
		Width:  32,
		Height: 32,
		Solid:  true,
	})

	world.Update(0)

	// Initial position should not collide
	if terrainChecker.CheckEntityCollision(entity) {
		t.Error("Entity at center should not collide with terrain")
	}

	// Move entity towards wall (left)
	// Entity needs to move from X=160 to X<48 to collide with wall at tile 0
	// Distance needed: >112 pixels. With VX=-100, need deltaTime > 1.12s
	movementSys := NewMovementSystem(200.0)
	movementSys.Update(world.GetEntities(), 1.2) // Move 120 pixels left (to X=40)

	// New position should collide with left wall
	if !terrainChecker.CheckEntityCollision(entity) {
		t.Error("Entity should collide with wall after moving left")
	}

	// Collision system should resolve the collision
	sys.Update(world.GetEntities(), 0.016)

	// After resolution, entity should not penetrate wall
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)

	// Entity should be pushed back from wall (x > 32, which is wall boundary)
	if pos.X < 32 {
		t.Errorf("Entity penetrated wall: X = %f, expected X >= 32", pos.X)
	}
}

// TestMovementSystemRequiresConstructor verifies that MovementSystem
// must be initialized with NewMovementSystem constructor.
func TestMovementSystemRequiresConstructor(t *testing.T) {
	maxSpeed := 200.0
	sys := NewMovementSystem(maxSpeed)

	if sys == nil {
		t.Fatal("NewMovementSystem returned nil")
	}

	if sys.MaxSpeed != maxSpeed {
		t.Errorf("MaxSpeed = %f, want %f", sys.MaxSpeed, maxSpeed)
	}
}

// TestMovementSystemInvalidInitialization tests that direct struct
// instantiation produces invalid configuration.
func TestMovementSystemInvalidInitialization(t *testing.T) {
	// Simulate the bug: direct struct instantiation
	sys := &MovementSystem{}

	// Verify that this produces invalid state
	if sys.MaxSpeed != 0.0 {
		t.Errorf("Expected MaxSpeed=0.0 from direct instantiation, got %f", sys.MaxSpeed)
	}

	t.Logf("WARNING: Direct instantiation produces MaxSpeed=0, disabling speed limiting")
}

// TestMovementSystemSpeedLimiting verifies that properly initialized
// movement system enforces speed limits.
func TestMovementSystemSpeedLimiting(t *testing.T) {
	world := NewWorld()
	maxSpeed := 100.0
	sys := NewMovementSystem(maxSpeed) // Proper initialization

	// Create entity with excessive velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 500, VY: 500}) // Way over limit

	world.Update(0)

	// Update movement system
	sys.Update(world.GetEntities(), 1.0) // 1 second

	// Velocity should be clamped to maxSpeed
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)

	// Calculate actual speed
	actualSpeed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)

	// Should be clamped to maxSpeed (within tolerance for floating point)
	tolerance := 0.01
	if actualSpeed > maxSpeed+tolerance {
		t.Errorf("Velocity not clamped: speed = %f, maxSpeed = %f", actualSpeed, maxSpeed)
	}

	// Position should reflect clamped velocity
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)

	// After 1 second at maxSpeed of 100, entity should move roughly 100 units
	// (diagonal, so ~70.7 units in each axis)
	expectedDistance := maxSpeed
	actualDistance := math.Sqrt(pos.X*pos.X + pos.Y*pos.Y)

	if math.Abs(actualDistance-expectedDistance) > tolerance {
		t.Errorf("Incorrect distance traveled: got %f, want ~%f", actualDistance, expectedDistance)
	}
}

// TestMovementSystemNoSpeedLimit verifies that MaxSpeed=0 means no limit.
func TestMovementSystemNoSpeedLimit(t *testing.T) {
	world := NewWorld()
	sys := NewMovementSystem(0) // 0 = no limit

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 10000, VY: 10000})

	world.Update(0)

	sys.Update(world.GetEntities(), 1.0)

	// Velocity should NOT be clamped
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)

	if vel.VX != 10000 || vel.VY != 10000 {
		t.Errorf("Velocity was clamped when MaxSpeed=0: VX=%f, VY=%f", vel.VX, vel.VY)
	}
}

// TestSystemInitializationIntegration tests full initialization pattern
// used in cmd/client/main.go after the fix.
func TestSystemInitializationIntegration(t *testing.T) {
	world := NewWorld()

	// Correct initialization pattern (after fix)
	movementSystem := NewMovementSystem(200.0)
	collisionSystem := NewCollisionSystem(64.0)

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	// Create terrain
	testTerrain := terrain.NewTerrain(20, 20, 12345)
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if x == 0 || x == 19 || y == 0 || y == 19 {
				testTerrain.SetTile(x, y, terrain.TileWall)
			} else {
				testTerrain.SetTile(x, y, terrain.TileFloor)
			}
		}
	}

	// Setup terrain collision
	terrainChecker := NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(testTerrain)
	collisionSystem.SetTerrainChecker(terrainChecker)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 320, Y: 320}) // Center of map
	player.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&ColliderComponent{
		Width:  32,
		Height: 32,
		Solid:  true,
	})

	world.Update(0)

	// Test 1: Player stays in place with zero velocity
	world.Update(1.0)

	posComp, _ := player.GetComponent("position")
	pos := posComp.(*PositionComponent)

	if pos.X != 320 || pos.Y != 320 {
		t.Errorf("Player moved with zero velocity: (%f, %f)", pos.X, pos.Y)
	}

	// Test 2: Player moves but stops at wall
	velComp, _ := player.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	vel.VX = -100 // Move left towards wall

	// Simulate movement towards wall
	for i := 0; i < 30; i++ { // 3 seconds of movement
		world.Update(0.1)
	}

	// Player should have hit the wall and stopped
	posComp, _ = player.GetComponent("position")
	pos = posComp.(*PositionComponent)

	// Left wall is at x=32 (tile 1, which is floor boundary)
	// Player collider is 32 pixels wide, centered at position
	// So minimum X position is around 48 (32 + 16 for half-width)
	if pos.X < 32 {
		t.Errorf("Player penetrated wall: X = %f, expected X >= 32", pos.X)
	}

	if pos.X > 100 {
		t.Errorf("Player didn't reach wall: X = %f, expected X ~ 48-64", pos.X)
	}

	t.Logf("Player stopped at wall: X = %.1f (correct behavior)", pos.X)
}

// BenchmarkCollisionSystemProperInit benchmarks properly initialized collision system.
func BenchmarkCollisionSystemProperInit(b *testing.B) {
	world := NewWorld()
	sys := NewCollisionSystem(64.0) // Proper initialization

	// Create 100 entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10 * 50),
			Y: float64(i / 10 * 50),
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
			Solid:  true,
		})
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

// BenchmarkCollisionSystemInvalidInit benchmarks improperly initialized collision system.
// This demonstrates the performance impact of the bug.
func BenchmarkCollisionSystemInvalidInit(b *testing.B) {
	world := NewWorld()
	sys := &CollisionSystem{} // Bug: direct instantiation, CellSize=0

	// Create 100 entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10 * 50),
			Y: float64(i / 10 * 50),
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
			Solid:  true,
		})
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
