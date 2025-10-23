//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestTerrainCollisionChecker_NewChecker tests checker creation.
func TestTerrainCollisionChecker_NewChecker(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	if checker == nil {
		t.Fatal("NewTerrainCollisionChecker returned nil")
	}

	if checker.tileWidth != 32 || checker.tileHeight != 32 {
		t.Errorf("Tile size not set correctly: expected 32x32, got %dx%d", checker.tileWidth, checker.tileHeight)
	}

	if checker.terrain != nil {
		t.Error("Terrain should be nil initially")
	}
}

// TestTerrainCollisionChecker_SetTerrain tests terrain setting.
func TestTerrainCollisionChecker_SetTerrain(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create simple test terrain
	testTerrain := terrain.NewTerrain(5, 5, 12345)
	testTerrain.SetTile(1, 1, terrain.TileFloor)
	testTerrain.SetTile(2, 1, terrain.TileFloor)
	testTerrain.SetTile(1, 2, terrain.TileFloor)
	testTerrain.SetTile(2, 2, terrain.TileFloor)

	// Set terrain
	checker.SetTerrain(testTerrain)

	// Check terrain was set
	if checker.terrain != testTerrain {
		t.Error("Terrain not set correctly")
	}
}

// TestTerrainCollisionChecker_SetTerrainNil tests nil terrain handling.
func TestTerrainCollisionChecker_SetTerrainNil(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// SetTerrain accepts nil (no error returned)
	checker.SetTerrain(nil)

	if checker.terrain != nil {
		t.Error("Terrain should be nil after setting to nil")
	}
}

// TestTerrainCollisionChecker_CheckCollision tests collision detection.
func TestTerrainCollisionChecker_CheckCollision(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create simple terrain with one floor tile at (1,1), rest walls
	testTerrain := terrain.NewTerrain(3, 3, 12345)
	testTerrain.SetTile(1, 1, terrain.TileFloor)
	checker.SetTerrain(testTerrain)

	tests := []struct {
		name     string
		x, y     float64
		width    float64
		height   float64
		wantColl bool
	}{
		{"center of floor tile", 48.0, 48.0, 16.0, 16.0, false},   // 32+16 = center of tile (1,1)
		{"center of wall tile", 16.0, 16.0, 16.0, 16.0, true},     // center of tile (0,0) which is wall
		{"edge of floor into wall", 32.0, 48.0, 16.0, 16.0, true}, // overlapping wall at x=0
		{"entirely in floor", 48.0, 48.0, 8.0, 8.0, false},        // small entity in floor
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotColl := checker.CheckCollision(tt.x, tt.y, tt.width, tt.height)
			if gotColl != tt.wantColl {
				t.Errorf("CheckCollision(%v, %v, %v, %v) = %v, want %v",
					tt.x, tt.y, tt.width, tt.height, gotColl, tt.wantColl)
			}
		})
	}
}

// TestTerrainCollisionChecker_CheckEntityCollision tests entity collision detection.
func TestTerrainCollisionChecker_CheckEntityCollision(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create terrain with floor at (1,1), rest walls
	testTerrain := terrain.NewTerrain(3, 3, 12345)
	testTerrain.SetTile(1, 1, terrain.TileFloor)
	checker.SetTerrain(testTerrain)

	world := NewWorld()

	// Create entity in floor tile (should not collide)
	floorEntity := world.CreateEntity()
	floorEntity.AddComponent(&PositionComponent{X: 48.0, Y: 48.0}) // center of tile (1,1)
	floorEntity.AddComponent(&ColliderComponent{Width: 16.0, Height: 16.0})

	// Create entity in wall tile (should collide)
	wallEntity := world.CreateEntity()
	wallEntity.AddComponent(&PositionComponent{X: 16.0, Y: 16.0}) // center of tile (0,0)
	wallEntity.AddComponent(&ColliderComponent{Width: 16.0, Height: 16.0})

	// Test floor entity
	if checker.CheckEntityCollision(floorEntity) {
		t.Error("Entity in floor tile should not collide with terrain")
	}

	// Test wall entity
	if !checker.CheckEntityCollision(wallEntity) {
		t.Error("Entity in wall tile should collide with terrain")
	}
}

// TestTerrainCollisionChecker_NoTerrain tests behavior when no terrain is set.
func TestTerrainCollisionChecker_NoTerrain(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Without terrain, should not detect collisions
	if checker.CheckCollision(0, 0, 16, 16) {
		t.Error("Checker without terrain should not detect collisions")
	}

	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&ColliderComponent{Width: 16, Height: 16})

	if checker.CheckEntityCollision(entity) {
		t.Error("Checker without terrain should not detect entity collisions")
	}
}

// TestTerrainCollisionChecker_MissingComponents tests entity without required components.
func TestTerrainCollisionChecker_MissingComponents(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)
	testTerrain := terrain.NewTerrain(3, 3, 12345)
	checker.SetTerrain(testTerrain)

	world := NewWorld()

	// Entity without components
	entity1 := world.CreateEntity()
	if checker.CheckEntityCollision(entity1) {
		t.Error("Entity without components should not collide")
	}

	// Entity with only position
	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 0, Y: 0})
	if checker.CheckEntityCollision(entity2) {
		t.Error("Entity without collider should not collide")
	}

	// Entity with only collider
	entity3 := world.CreateEntity()
	entity3.AddComponent(&ColliderComponent{Width: 16, Height: 16})
	if checker.CheckEntityCollision(entity3) {
		t.Error("Entity without position should not collide")
	}
}
