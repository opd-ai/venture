// Package engine provides tests for multi-layer terrain collision detection.
// Phase 11.1: Multi-layer collision testing.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestTerrainCollisionChecker_CheckCollisionBoundsWithLayer tests layer-aware collision detection.
func TestTerrainCollisionChecker_CheckCollisionBoundsWithLayer(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create test terrain with multi-layer tiles
	testTerrain := &terrain.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([]terrain.TileType, 100),
	}

	// Setup test tiles:
	// (1, 1): Wall
	// (2, 2): Pit
	// (3, 3): Diagonal wall NE
	testTerrain.SetTile(1, 1, terrain.TileWall)
	testTerrain.SetTile(2, 2, terrain.TilePit)
	testTerrain.SetTile(3, 3, terrain.TileWallNE)

	checker.SetTerrain(testTerrain)

	tests := []struct {
		name        string
		minX, minY  float64
		maxX, maxY  float64
		layer       int
		wantCollide bool
		description string
	}{
		// Wall collision tests (all layers collide with walls)
		{
			name: "ground layer collides with wall",
			minX: 32.0, minY: 32.0, maxX: 48.0, maxY: 48.0,
			layer:       0,
			wantCollide: true,
			description: "Ground entities hit walls",
		},
		{
			name: "water layer collides with wall",
			minX: 32.0, minY: 32.0, maxX: 48.0, maxY: 48.0,
			layer:       1,
			wantCollide: true,
			description: "Water entities hit walls",
		},
		{
			name: "platform layer collides with wall",
			minX: 32.0, minY: 32.0, maxX: 48.0, maxY: 48.0,
			layer:       2,
			wantCollide: true,
			description: "Platform entities hit walls",
		},

		// Pit collision tests (only ground layer collides)
		{
			name: "ground layer collides with pit",
			minX: 64.0, minY: 64.0, maxX: 80.0, maxY: 80.0,
			layer:       0,
			wantCollide: true,
			description: "Ground entities can't enter pits",
		},
		{
			name: "water layer doesn't collide with pit",
			minX: 64.0, minY: 64.0, maxX: 80.0, maxY: 80.0,
			layer:       1,
			wantCollide: false,
			description: "Water entities are IN pits",
		},
		{
			name: "platform layer doesn't collide with pit",
			minX: 64.0, minY: 64.0, maxX: 80.0, maxY: 80.0,
			layer:       2,
			wantCollide: false,
			description: "Platform entities are ABOVE pits",
		},

		// Diagonal wall collision tests (all layers)
		{
			name: "ground layer collides with diagonal wall",
			minX: 96.0, minY: 96.0, maxX: 112.0, maxY: 112.0,
			layer:       0,
			wantCollide: true,
			description: "Ground entities hit diagonal walls",
		},
		{
			name: "platform layer collides with diagonal wall",
			minX: 96.0, minY: 96.0, maxX: 112.0, maxY: 112.0,
			layer:       2,
			wantCollide: true,
			description: "Platform entities hit diagonal walls",
		},

		// No collision tests
		{
			name: "ground layer - empty space",
			minX: 160.0, minY: 160.0, maxX: 176.0, maxY: 176.0,
			layer:       0,
			wantCollide: false,
			description: "Empty space has no collision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.CheckCollisionBoundsWithLayer(tt.minX, tt.minY, tt.maxX, tt.maxY, tt.layer)
			if got != tt.wantCollide {
				t.Errorf("CheckCollisionBoundsWithLayer() = %v, want %v\n%s", got, tt.wantCollide, tt.description)
			}
		})
	}
}

// TestTerrainCollisionChecker_tileMatchesLayer tests the layer matching logic.
func TestTerrainCollisionChecker_tileMatchesLayer(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	tests := []struct {
		name      string
		tile      terrain.TileType
		layer     int
		wantMatch bool
	}{
		// Layer 0 (ground)
		{"ground - wall", terrain.TileWall, 0, true},
		{"ground - pit", terrain.TilePit, 0, true},
		{"ground - diagonal NE", terrain.TileWallNE, 0, true},
		{"ground - floor", terrain.TileFloor, 0, false},

		// Layer 1 (water/pit)
		{"water - wall", terrain.TileWall, 1, true},
		{"water - pit (no collision)", terrain.TilePit, 1, false},
		{"water - diagonal NW", terrain.TileWallNW, 1, true},
		{"water - floor", terrain.TileFloor, 1, false},

		// Layer 2 (platform)
		{"platform - wall", terrain.TileWall, 2, true},
		{"platform - pit (no collision)", terrain.TilePit, 2, false},
		{"platform - diagonal SE", terrain.TileWallSE, 2, true},
		{"platform - floor", terrain.TileFloor, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.tileMatchesLayer(tt.tile, tt.layer)
			if got != tt.wantMatch {
				t.Errorf("tileMatchesLayer(%v, %d) = %v, want %v",
					tt.tile, tt.layer, got, tt.wantMatch)
			}
		})
	}
}

// TestTerrainCollisionChecker_CheckEntityCollision_WithLayer tests entity collision with layer support.
func TestTerrainCollisionChecker_CheckEntityCollision_WithLayer(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create test terrain
	testTerrain := &terrain.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([]terrain.TileType, 100),
	}
	testTerrain.SetTile(2, 2, terrain.TilePit)
	checker.SetTerrain(testTerrain)

	tests := []struct {
		name        string
		entity      *Entity
		wantCollide bool
		description string
	}{
		{
			name: "ground entity collides with pit",
			entity: func() *Entity {
				e := NewEntity()
				e.AddComponent(&PositionComponent{X: 70, Y: 70})
				e.AddComponent(&ColliderComponent{Width: 16, Height: 16})
				e.AddComponent(NewLayerComponent()) // Ground layer
				return e
			}(),
			wantCollide: true,
			description: "Ground entities can't enter pits",
		},
		{
			name: "flying entity doesn't collide with pit",
			entity: func() *Entity {
				e := NewEntity()
				e.AddComponent(&PositionComponent{X: 70, Y: 70})
				e.AddComponent(&ColliderComponent{Width: 16, Height: 16})
				layerComp := NewFlyingLayerComponent()
				layerComp.CurrentLayer = 2 // On platform layer
				e.AddComponent(&layerComp)
				return e
			}(),
			wantCollide: false,
			description: "Flying entities above pits don't collide",
		},
		{
			name: "swimming entity in pit doesn't collide",
			entity: func() *Entity {
				e := NewEntity()
				e.AddComponent(&PositionComponent{X: 70, Y: 70})
				e.AddComponent(&ColliderComponent{Width: 16, Height: 16})
				e.AddComponent(NewSwimmingLayerComponent()) // Water layer (1)
				return e
			}(),
			wantCollide: false,
			description: "Swimming entities in pits don't collide",
		},
		{
			name: "entity without layer component (defaults to ground)",
			entity: func() *Entity {
				e := NewEntity()
				e.AddComponent(&PositionComponent{X: 70, Y: 70})
				e.AddComponent(&ColliderComponent{Width: 16, Height: 16})
				// No layer component - should default to ground (0)
				return e
			}(),
			wantCollide: true,
			description: "Entities without layer component use ground layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.CheckEntityCollision(tt.entity)
			if got != tt.wantCollide {
				t.Errorf("CheckEntityCollision() = %v, want %v\n%s",
					got, tt.wantCollide, tt.description)
			}
		})
	}
}

// TestTerrainCollisionChecker_CheckCollisionWithLayer tests the new layer-aware method.
func TestTerrainCollisionChecker_CheckCollisionWithLayer(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create test terrain with a pit
	testTerrain := &terrain.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([]terrain.TileType, 100),
	}
	testTerrain.SetTile(1, 1, terrain.TilePit)
	checker.SetTerrain(testTerrain)

	tests := []struct {
		name        string
		x, y        float64
		width       float64
		height      float64
		layer       int
		wantCollide bool
	}{
		{
			name: "ground layer at pit - collides",
			x:    48, y: 48, width: 16, height: 16,
			layer:       0,
			wantCollide: true,
		},
		{
			name: "water layer at pit - no collision",
			x:    48, y: 48, width: 16, height: 16,
			layer:       1,
			wantCollide: false,
		},
		{
			name: "platform layer at pit - no collision",
			x:    48, y: 48, width: 16, height: 16,
			layer:       2,
			wantCollide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.CheckCollisionWithLayer(tt.x, tt.y, tt.width, tt.height, tt.layer)
			if got != tt.wantCollide {
				t.Errorf("CheckCollisionWithLayer() = %v, want %v", got, tt.wantCollide)
			}
		})
	}
}

// TestTerrainCollisionChecker_MultiLayer_Integration tests the complete multi-layer system.
func TestTerrainCollisionChecker_MultiLayer_Integration(t *testing.T) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create complex test terrain
	testTerrain := &terrain.Terrain{
		Width:  5,
		Height: 5,
		Tiles:  make([]terrain.TileType, 25),
	}

	// Setup terrain:
	// [W][P][D][F][F]
	// [P][F][W][F][F]
	// [D][W][F][P][F]
	// [F][F][P][F][W]
	// [F][F][F][W][F]
	// W=Wall, P=Pit, D=DiagonalNE, F=Floor

	testTerrain.SetTile(0, 0, terrain.TileWall)
	testTerrain.SetTile(1, 0, terrain.TilePit)
	testTerrain.SetTile(2, 0, terrain.TileWallNE)

	testTerrain.SetTile(0, 1, terrain.TilePit)
	testTerrain.SetTile(2, 1, terrain.TileWall)

	testTerrain.SetTile(0, 2, terrain.TileWallNE)
	testTerrain.SetTile(1, 2, terrain.TileWall)
	testTerrain.SetTile(3, 2, terrain.TilePit)

	testTerrain.SetTile(2, 3, terrain.TilePit)
	testTerrain.SetTile(4, 3, terrain.TileWall)

	testTerrain.SetTile(3, 4, terrain.TileWall)

	checker.SetTerrain(testTerrain)

	// Test entity on ground layer
	groundEntity := NewEntity()
	groundEntity.AddComponent(&PositionComponent{X: 48, Y: 16}) // Position at pit (1,0)
	groundEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16})
	groundEntity.AddComponent(NewLayerComponent())

	if !checker.CheckEntityCollision(groundEntity) {
		t.Error("Ground entity should collide with pit at (1,0)")
	}

	// Test entity on platform layer
	platformEntity := NewEntity()
	platformEntity.AddComponent(&PositionComponent{X: 48, Y: 16}) // Same position
	platformEntity.AddComponent(&ColliderComponent{Width: 16, Height: 16})
	layerComp := NewLayerComponent()
	layerComp.CurrentLayer = 2
	platformEntity.AddComponent(&layerComp)

	if checker.CheckEntityCollision(platformEntity) {
		t.Error("Platform entity should NOT collide with pit at (1,0)")
	}

	// Test wall collision (all layers)
	for layer := 0; layer <= 2; layer++ {
		entity := NewEntity()
		entity.AddComponent(&PositionComponent{X: 16, Y: 16}) // Position at wall (0,0)
		entity.AddComponent(&ColliderComponent{Width: 16, Height: 16})
		layerC := NewLayerComponent()
		layerC.CurrentLayer = layer
		entity.AddComponent(&layerC)

		if !checker.CheckEntityCollision(entity) {
			t.Errorf("Entity on layer %d should collide with wall at (0,0)", layer)
		}
	}
}

// BenchmarkCheckCollisionBoundsWithLayer benchmarks multi-layer collision detection.
func BenchmarkCheckCollisionBoundsWithLayer(b *testing.B) {
	checker := NewTerrainCollisionChecker(32, 32)

	// Create realistic terrain
	testTerrain := &terrain.Terrain{
		Width:  100,
		Height: 100,
		Tiles:  make([]terrain.TileType, 10000),
	}

	// Fill with some obstacles
	for i := 0; i < 10000; i += 10 {
		testTerrain.Tiles[i] = terrain.TileWall
	}
	for i := 5; i < 10000; i += 15 {
		testTerrain.Tiles[i] = terrain.TilePit
	}

	checker.SetTerrain(testTerrain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test collision at various positions and layers
		layer := i % 3
		x := float64((i * 7) % 3000)
		y := float64((i * 11) % 3000)
		_ = checker.CheckCollisionBoundsWithLayer(x, y, x+16, y+16, layer)
	}
}
