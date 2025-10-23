//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestTerrainCollisionSystem_NewSystem tests system creation.
func TestTerrainCollisionSystem_NewSystem(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCollisionSystem(world, 32, 32)

	if system == nil {
		t.Fatal("NewTerrainCollisionSystem returned nil")
	}

	if system.world != world {
		t.Error("World reference not set correctly")
	}

	if system.tileWidth != 32 || system.tileHeight != 32 {
		t.Errorf("Tile size not set correctly: expected 32x32, got %dx%d", system.tileWidth, system.tileHeight)
	}

	if system.initialized {
		t.Error("System should not be initialized without terrain")
	}

	if len(system.wallEntities) != 0 {
		t.Error("Wall entities should be empty initially")
	}
}

// TestTerrainCollisionSystem_SetTerrain tests terrain setting and wall creation.
func TestTerrainCollisionSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCollisionSystem(world, 32, 32)

	// Create simple test terrain
	testTerrain := terrain.NewTerrain(5, 5, 12345)
	testTerrain.SetTile(1, 1, terrain.TileFloor)
	testTerrain.SetTile(2, 1, terrain.TileFloor)
	testTerrain.SetTile(1, 2, terrain.TileFloor)
	testTerrain.SetTile(2, 2, terrain.TileFloor)

	// Set terrain
	err := system.SetTerrain(testTerrain)
	if err != nil {
		t.Fatalf("SetTerrain failed: %v", err)
	}

	// Check system state
	if !system.IsInitialized() {
		t.Error("System should be initialized after setting terrain")
	}

	// Should create collision entities for walls (21 walls in 5x5 grid with 4 floor tiles)
	expectedWalls := 25 - 4 // Total tiles - floor tiles
	if system.GetWallEntityCount() != expectedWalls {
		t.Errorf("Expected %d wall entities, got %d", expectedWalls, system.GetWallEntityCount())
	}

	// Check that wall entities were added to world
	entities := world.GetEntities()
	wallEntityCount := 0
	for _, entity := range entities {
		if entity.HasComponent("wall") {
			wallEntityCount++
		}
	}

	if wallEntityCount != expectedWalls {
		t.Errorf("Expected %d wall entities in world, got %d", expectedWalls, wallEntityCount)
	}
}

// TestTerrainCollisionSystem_SetTerrainNil tests error handling for nil terrain.
func TestTerrainCollisionSystem_SetTerrainNil(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCollisionSystem(world, 32, 32)

	err := system.SetTerrain(nil)
	if err == nil {
		t.Error("SetTerrain should return error for nil terrain")
	}
}

// TestTerrainCollisionSystem_WallColliderProperties tests wall collider properties.
func TestTerrainCollisionSystem_WallColliderProperties(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCollisionSystem(world, 32, 32)

	// Create simple terrain with one wall
	testTerrain := terrain.NewTerrain(3, 3, 12345)
	testTerrain.SetTile(1, 1, terrain.TileFloor) // Make center floor, rest walls

	err := system.SetTerrain(testTerrain)
	if err != nil {
		t.Fatalf("SetTerrain failed: %v", err)
	}

	// Find a wall entity
	entities := world.GetEntities()
	var wallEntity *Entity
	for _, entity := range entities {
		if entity.HasComponent("wall") {
			wallEntity = entity
			break
		}
	}

	if wallEntity == nil {
		t.Fatal("No wall entity found")
	}

	// Check position component
	if !wallEntity.HasComponent("position") {
		t.Error("Wall entity should have position component")
	}

	// Check collider component
	if !wallEntity.HasComponent("collider") {
		t.Error("Wall entity should have collider component")
	}

	colliderComp, _ := wallEntity.GetComponent("collider")
	collider := colliderComp.(*ColliderComponent)

	if !collider.Solid {
		t.Error("Wall collider should be solid")
	}

	if collider.IsTrigger {
		t.Error("Wall collider should not be a trigger")
	}

	if collider.Layer != 0 {
		t.Errorf("Wall collider should be on layer 0, got layer %d", collider.Layer)
	}

	if collider.Width != 32 || collider.Height != 32 {
		t.Errorf("Wall collider should be 32x32, got %.1fx%.1f", collider.Width, collider.Height)
	}
}

// TestTerrainCollisionSystem_Cleanup tests entity cleanup on re-initialization.
func TestTerrainCollisionSystem_Cleanup(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCollisionSystem(world, 32, 32)

	// Create terrain and initialize
	testTerrain1 := terrain.NewTerrain(3, 3, 12345)
	testTerrain1.SetTile(1, 1, terrain.TileFloor)

	err := system.SetTerrain(testTerrain1)
	if err != nil {
		t.Fatalf("SetTerrain failed: %v", err)
	}

	initialWallCount := system.GetWallEntityCount()
	if initialWallCount == 0 {
		t.Fatal("Should have created some wall entities")
	}

	// Re-initialize with different terrain
	testTerrain2 := terrain.NewTerrain(2, 2, 54321)
	// All walls in 2x2 terrain

	err = system.SetTerrain(testTerrain2)
	if err != nil {
		t.Fatalf("Second SetTerrain failed: %v", err)
	}

	// Should have cleaned up old entities and created new ones
	newWallCount := system.GetWallEntityCount()
	if newWallCount != 4 { // 2x2 = 4 walls
		t.Errorf("Expected 4 wall entities after re-init, got %d", newWallCount)
	}

	// Check that old entities are no longer in world
	entities := world.GetEntities()
	wallEntityCount := 0
	for _, entity := range entities {
		if entity.HasComponent("wall") {
			wallEntityCount++
		}
	}

	if wallEntityCount != newWallCount {
		t.Errorf("Cleanup failed: expected %d wall entities in world, got %d", newWallCount, wallEntityCount)
	}
}

// TestTerrainCollisionSystem_Update tests the update method (should be no-op).
func TestTerrainCollisionSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewTerrainCollisionSystem(world, 32, 32)

	// Create terrain
	testTerrain := terrain.NewTerrain(3, 3, 12345)
	err := system.SetTerrain(testTerrain)
	if err != nil {
		t.Fatalf("SetTerrain failed: %v", err)
	}

	initialCount := system.GetWallEntityCount()

	// Update should not change anything
	system.Update(world.GetEntities(), 0.016)

	if system.GetWallEntityCount() != initialCount {
		t.Error("Update should not modify wall entity count")
	}

	if !system.IsInitialized() {
		t.Error("Update should not change initialization state")
	}
}

// TestTerrainCollisionSystem_WallComponent tests the WallComponent.
func TestTerrainCollisionSystem_WallComponent(t *testing.T) {
	wallComp := &WallComponent{TileX: 5, TileY: 10}

	if wallComp.Type() != "wall" {
		t.Errorf("WallComponent.Type() should return 'wall', got '%s'", wallComp.Type())
	}

	if wallComp.TileX != 5 {
		t.Errorf("TileX should be 5, got %d", wallComp.TileX)
	}

	if wallComp.TileY != 10 {
		t.Errorf("TileY should be 10, got %d", wallComp.TileY)
	}
}
