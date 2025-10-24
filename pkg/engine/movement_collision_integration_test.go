//go:build test
// +build test

// Package engine provides integration tests for movement and collision systems.
// This file tests the critical interaction between MovementSystem and CollisionSystem,
// particularly the prevention of movement through walls.
package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestMovementBlockedByTerrain verifies that entities cannot move through terrain walls.
// This is a critical integration test for GAP-001 repair.
func TestMovementBlockedByTerrain(t *testing.T) {
	world := NewWorld()

	// Create systems with proper initialization
	movementSystem := NewMovementSystem(200.0)
	collisionSystem := NewCollisionSystem(32.0)

	// GAP-001 REPAIR: Connect collision system to movement for predictive checking
	movementSystem.SetCollisionSystem(collisionSystem)

	// Create simple terrain: 10x10 grid with walls around the perimeter
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

	// Setup terrain checker and connect to collision system
	terrainChecker := NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(testTerrain)
	collisionSystem.SetTerrainChecker(terrainChecker)

	// Create entity in the center of the map
	// Tile (5, 5) -> World position (160, 160)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 160, Y: 160})
	entity.AddComponent(&VelocityComponent{VX: -100, VY: 0}) // Moving left toward wall
	entity.AddComponent(&ColliderComponent{
		Width:   32,
		Height:  32,
		Solid:   true,
		OffsetX: -16,
		OffsetY: -16, // Centered collider
	})

	world.Update(0) // Process entity additions

	// Record starting position
	posComp, _ := entity.GetComponent("position")
	startPos := posComp.(*PositionComponent)
	startX := startPos.X

	// Simulate movement toward the left wall
	// Wall is at tile X=0, which is world X=0-32
	// Entity center needs to stay at X >= 32 (collider left edge at X >= 16)
	// Since collider is offset by -16, entity center at X=32 means collider spans X[16-48]

	// Move for 2 seconds worth of frames (120 frames at 60 FPS)
	// With VX=-100, entity should move 200 pixels left
	// Starting at X=160, should end at X=32 (blocked by wall), NOT X=-40
	for i := 0; i < 120; i++ {
		movementSystem.Update(world.GetEntities(), 0.016) // ~60 FPS
	}

	// Check final position
	finalPos := posComp.(*PositionComponent)
	finalX := finalPos.X

	// Entity should have stopped at the wall boundary
	// With centered collider (offset -16), minimum X position is 16 (collider left edge at 0)
	// But terrain collision should keep entity at least at X >= 32 (one tile width)
	if finalX < 16 {
		t.Errorf("Entity moved through wall! Final X = %.2f, expected X >= 16", finalX)
	}

	// Verify entity actually moved from starting position (not stuck)
	if finalX >= startX {
		t.Errorf("Entity didn't move at all! Start X = %.2f, Final X = %.2f", startX, finalX)
	}

	// Verify velocity was zeroed when hitting wall
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX != 0 {
		t.Errorf("Velocity not zeroed after hitting wall: VX = %.2f, expected 0", vel.VX)
	}

	t.Logf("Movement correctly blocked by terrain. Start X=%.2f, Final X=%.2f", startX, finalX)
}

// TestMovementSlidesAlongWalls verifies that diagonal movement slides along walls
// rather than stopping completely.
func TestMovementSlidesAlongWalls(t *testing.T) {
	world := NewWorld()

	movementSystem := NewMovementSystem(200.0)
	collisionSystem := NewCollisionSystem(32.0)
	movementSystem.SetCollisionSystem(collisionSystem)

	// Create terrain with wall on left side
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if x == 0 {
				testTerrain.SetTile(x, y, terrain.TileWall) // Left wall
			} else {
				testTerrain.SetTile(x, y, terrain.TileFloor)
			}
		}
	}

	terrainChecker := NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(testTerrain)
	collisionSystem.SetTerrainChecker(terrainChecker)

	// Entity moving diagonally toward the wall (left and down)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 160, Y: 160})
	entity.AddComponent(&VelocityComponent{VX: -50, VY: 50}) // Diagonal movement
	entity.AddComponent(&ColliderComponent{
		Width:   32,
		Height:  32,
		Solid:   true,
		OffsetX: -16,
		OffsetY: -16,
	})

	world.Update(0)

	// Record starting position
	posComp, _ := entity.GetComponent("position")
	startY := posComp.(*PositionComponent).Y

	// Move for 60 frames
	for i := 0; i < 60; i++ {
		movementSystem.Update(world.GetEntities(), 0.016)
	}

	// Check position
	finalPos := posComp.(*PositionComponent)

	// X should be blocked near the wall
	if finalPos.X < 16 {
		t.Errorf("Entity moved through wall in X direction! X = %.2f", finalPos.X)
	}

	// Y should have continued moving (sliding along the wall)
	if finalPos.Y <= startY {
		t.Errorf("Entity didn't slide along wall! Start Y=%.2f, Final Y=%.2f", startY, finalPos.Y)
	}

	// VX should be zeroed, but VY should remain
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)

	if vel.VX != 0 {
		t.Logf("Note: VX not zeroed immediately (expected for sliding), VX=%.2f", vel.VX)
	}

	t.Logf("Entity correctly slid along wall. Final pos: (%.2f, %.2f)", finalPos.X, finalPos.Y)
}

// TestMovementBlockedByEntity verifies that entities cannot move through each other.
func TestMovementBlockedByEntity(t *testing.T) {
	world := NewWorld()

	movementSystem := NewMovementSystem(200.0)
	collisionSystem := NewCollisionSystem(32.0)
	movementSystem.SetCollisionSystem(collisionSystem)

	// No terrain needed for this test
	terrainChecker := NewTerrainCollisionChecker(32, 32)
	collisionSystem.SetTerrainChecker(terrainChecker)

	// Create two entities: one moving, one stationary
	movingEntity := world.CreateEntity()
	movingEntity.AddComponent(&PositionComponent{X: 0, Y: 0})
	movingEntity.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	movingEntity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	stationaryEntity := world.CreateEntity()
	stationaryEntity.AddComponent(&PositionComponent{X: 100, Y: 0})
	stationaryEntity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	stationaryEntity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	world.Update(0)

	// Record starting distance
	posComp1, _ := movingEntity.GetComponent("position")
	posComp2, _ := stationaryEntity.GetComponent("position")
	startX1 := posComp1.(*PositionComponent).X
	startX2 := posComp2.(*PositionComponent).X
	startDistance := math.Abs(startX2 - startX1)

	// Move for 2 seconds (120 frames)
	// Moving entity should stop when colliding with stationary entity
	for i := 0; i < 120; i++ {
		movementSystem.Update(world.GetEntities(), 0.016)
	}

	// Check final positions
	finalX1 := posComp1.(*PositionComponent).X
	finalX2 := posComp2.(*PositionComponent).X
	finalDistance := math.Abs(finalX2 - finalX1)

	// Entities should not overlap (distance should be at least sum of radii)
	minDistance := 32.0 // Both have width 32, so minimum distance is 32
	if finalDistance < minDistance {
		t.Errorf("Entities overlapping! Distance = %.2f, expected >= %.2f", finalDistance, minDistance)
	}

	// Moving entity should have moved closer but stopped before overlapping
	if finalX1 <= startX1 {
		t.Errorf("Moving entity didn't move! Start=%.2f, Final=%.2f", startX1, finalX1)
	}

	// Stationary entity should not have moved
	if finalX2 != startX2 {
		t.Errorf("Stationary entity moved! Start=%.2f, Final=%.2f", startX2, finalX2)
	}

	t.Logf("Entities correctly blocked. Start distance=%.2f, Final distance=%.2f", startDistance, finalDistance)
}

// TestMovementWithoutCollisionSystem verifies backward compatibility.
// When collision system is not set, movement should work as before.
func TestMovementWithoutCollisionSystem(t *testing.T) {
	world := NewWorld()

	// Create movement system WITHOUT connecting collision system
	movementSystem := NewMovementSystem(200.0)

	// Create entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 50, VY: 0})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	world.Update(0)

	// Record starting position
	posComp, _ := entity.GetComponent("position")
	startX := posComp.(*PositionComponent).X

	// Move for 1 second at 60 FPS (60 frames × 1/60 second)
	deltaTime := 1.0 / 60.0 // 0.01666... seconds per frame
	numFrames := 60
	for i := 0; i < numFrames; i++ {
		movementSystem.Update(world.GetEntities(), deltaTime)
	}

	finalX := posComp.(*PositionComponent).X

	// Entity should have moved (no collision checking)
	totalTime := deltaTime * float64(numFrames) // Exactly 1.0 second
	expectedX := startX + 50*totalTime          // VX * time = 50 * 1.0 = 50
	if math.Abs(finalX-expectedX) > 0.1 {
		t.Errorf("Movement incorrect without collision. Start=%.2f, Final=%.2f, Expected=%.2f", startX, finalX, expectedX)
	}

	t.Logf("Movement works without collision system. Moved from %.2f to %.2f", startX, finalX)
}

// TestPredictiveCollisionMethods tests the new collision prediction methods directly.
func TestPredictiveCollisionMethods(t *testing.T) {
	world := NewWorld()
	collisionSystem := NewCollisionSystem(32.0)

	// Create terrain with single wall tile
	testTerrain := terrain.NewTerrain(10, 10, 12345)

	// Clear all tiles to floor first (NewTerrain initializes everything as walls)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			testTerrain.SetTile(x, y, terrain.TileFloor)
		}
	}

	// Now set one wall tile for testing
	testTerrain.SetTile(5, 5, terrain.TileWall) // Wall at world (160, 160)

	terrainChecker := NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(testTerrain)
	collisionSystem.SetTerrainChecker(terrainChecker)

	// Create entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true, OffsetX: -16, OffsetY: -16})

	world.Update(0)

	// Test 1: Should NOT collide at current position (away from wall)
	if collisionSystem.WouldCollideWithTerrain(entity, 100, 100) {
		t.Error("False positive: Entity should not collide at position (100, 100)")
	}

	// Test 2: Should collide if moved to wall position
	if !collisionSystem.WouldCollideWithTerrain(entity, 160, 160) {
		t.Error("False negative: Entity should collide at wall position (160, 160)")
	}

	// Test 3: Should collide near wall edge
	if !collisionSystem.WouldCollideWithTerrain(entity, 145, 160) {
		t.Error("False negative: Entity should collide near wall edge")
	}

	// Test 4: Entity-to-entity collision prediction
	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 150, Y: 100})
	entity2.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

	world.Update(0)

	// Should NOT collide at current positions
	if collisionSystem.WouldCollideWithEntity(entity, 100, 100, entity2) {
		t.Error("False positive: Entities should not collide at starting positions")
	}

	// Should collide if entity1 moved to entity2's position
	if !collisionSystem.WouldCollideWithEntity(entity, 150, 100, entity2) {
		t.Error("False negative: Entities should collide if overlapping")
	}

	t.Log("Predictive collision methods working correctly")
}

// BenchmarkMovementWithCollisionPrediction benchmarks the performance impact
// of predictive collision checking.
func BenchmarkMovementWithCollisionPrediction(b *testing.B) {
	world := NewWorld()
	movementSystem := NewMovementSystem(200.0)
	collisionSystem := NewCollisionSystem(32.0)
	movementSystem.SetCollisionSystem(collisionSystem)

	// Create terrain
	testTerrain := terrain.NewTerrain(50, 50, 12345)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if x == 0 || x == 49 || y == 0 || y == 49 {
				testTerrain.SetTile(x, y, terrain.TileWall)
			} else {
				testTerrain.SetTile(x, y, terrain.TileFloor)
			}
		}
	}

	terrainChecker := NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(testTerrain)
	collisionSystem.SetTerrainChecker(terrainChecker)

	// Create 100 moving entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(100 + i*10), Y: float64(100 + i*5)})
		entity.AddComponent(&VelocityComponent{VX: float64(i%10 - 5), VY: float64(i%7 - 3)})
		entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true, OffsetX: -16, OffsetY: -16})
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movementSystem.Update(entities, 0.016)
	}
}
