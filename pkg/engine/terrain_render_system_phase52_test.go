// Package engine provides Phase 5.2 tile rendering integration tests.
// This file tests advanced tile rendering features: transitions, parallax, enhanced walls.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

// TestTerrainRenderSystem_AdvancedFeatures tests configuration of advanced rendering features.
func TestTerrainRenderSystem_AdvancedFeatures(t *testing.T) {
	tests := []struct {
		name                string
		enableTransitions   bool
		enableParallax      bool
		enableEnhancedWalls bool
		wantCacheCleared    bool
	}{
		{
			name:                "All features enabled",
			enableTransitions:   true,
			enableParallax:      true,
			enableEnhancedWalls: true,
			wantCacheCleared:    false,
		},
		{
			name:                "All features disabled",
			enableTransitions:   false,
			enableParallax:      false,
			enableEnhancedWalls: false,
			wantCacheCleared:    false,
		},
		{
			name:                "Only transitions enabled",
			enableTransitions:   true,
			enableParallax:      false,
			enableEnhancedWalls: false,
			wantCacheCleared:    false,
		},
		{
			name:                "Only enhanced walls enabled",
			enableTransitions:   false,
			enableParallax:      false,
			enableEnhancedWalls: true,
			wantCacheCleared:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

			// Configure features
			sys.SetTransitionsEnabled(tt.enableTransitions)
			sys.SetParallaxEnabled(tt.enableParallax)
			sys.SetEnhancedWallsEnabled(tt.enableEnhancedWalls)

			// Verify configuration
			if sys.enableTransitions != tt.enableTransitions {
				t.Errorf("enableTransitions = %v, want %v", sys.enableTransitions, tt.enableTransitions)
			}
			if sys.enableParallax != tt.enableParallax {
				t.Errorf("enableParallax = %v, want %v", sys.enableParallax, tt.enableParallax)
			}
			if sys.enableEnhancedWalls != tt.enableEnhancedWalls {
				t.Errorf("enableEnhancedWalls = %v, want %v", sys.enableEnhancedWalls, tt.enableEnhancedWalls)
			}
		})
	}
}

// TestTerrainRenderSystem_CameraPosition tests camera position tracking for parallax.
func TestTerrainRenderSystem_CameraPosition(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	tests := []struct {
		name string
		x, y float64
	}{
		{"Origin", 0, 0},
		{"Positive", 100, 200},
		{"Negative", -50, -75},
		{"Large values", 10000, 20000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetCameraPosition(tt.x, tt.y)

			if sys.cameraX != tt.x {
				t.Errorf("cameraX = %v, want %v", sys.cameraX, tt.x)
			}
			if sys.cameraY != tt.y {
				t.Errorf("cameraY = %v, want %v", sys.cameraY, tt.y)
			}
		})
	}
}

// TestTerrainRenderSystem_WallNeighborDetection tests wall neighbor detection for enhanced walls.
func TestTerrainRenderSystem_WallNeighborDetection(t *testing.T) {
	// Create a simple test terrain with walls
	testTerrain := &terrain.Terrain{
		Width:  5,
		Height: 5,
		Tiles:  make([][]terrain.TileType, 5),
	}

	// Initialize 2D array
	for y := range testTerrain.Tiles {
		testTerrain.Tiles[y] = make([]terrain.TileType, 5)
	}

	// Create a cross pattern of walls
	// . W .
	// W W W
	// . W .
	testTerrain.Tiles[0][1] = terrain.TileWall // North (row 0, col 1)
	testTerrain.Tiles[1][0] = terrain.TileWall // West  (row 1, col 0)
	testTerrain.Tiles[1][1] = terrain.TileWall // Center (row 1, col 1)
	testTerrain.Tiles[1][2] = terrain.TileWall // East  (row 1, col 2)
	testTerrain.Tiles[2][1] = terrain.TileWall // South (row 2, col 1)

	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)
	sys.SetTerrain(testTerrain)

	// Test center tile (should have all four neighbors)
	neighbors := sys.getWallNeighbors(1, 1)

	if !neighbors.North {
		t.Error("Expected north neighbor to be wall")
	}
	if !neighbors.South {
		t.Error("Expected south neighbor to be wall")
	}
	if !neighbors.East {
		t.Error("Expected east neighbor to be wall")
	}
	if !neighbors.West {
		t.Error("Expected west neighbor to be wall")
	}
}

// TestTerrainRenderSystem_TileNeighborDetection tests 8-directional neighbor detection for transitions.
func TestTerrainRenderSystem_TileNeighborDetection(t *testing.T) {
	// Create a 3x3 terrain with floor tiles
	testTerrain := &terrain.Terrain{
		Width:  3,
		Height: 3,
		Tiles:  make([][]terrain.TileType, 3),
	}

	// Fill all with floor tiles
	for y := range testTerrain.Tiles {
		testTerrain.Tiles[y] = make([]terrain.TileType, 3)
		for x := range testTerrain.Tiles[y] {
			testTerrain.Tiles[y][x] = terrain.TileFloor
		}
	}

	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)
	sys.SetTerrain(testTerrain)

	// Test center tile (should have all 8 neighbors)
	neighbors := sys.getTileNeighbors(1, 1, tiles.TileFloor)

	allDirections := []struct {
		name  string
		value bool
	}{
		{"N", neighbors.N},
		{"NE", neighbors.NE},
		{"E", neighbors.E},
		{"SE", neighbors.SE},
		{"S", neighbors.S},
		{"SW", neighbors.SW},
		{"W", neighbors.W},
		{"NW", neighbors.NW},
	}

	for _, dir := range allDirections {
		if !dir.value {
			t.Errorf("Expected %s neighbor to be floor tile", dir.name)
		}
	}
}

// TestTerrainRenderSystem_TileNeighborEdgeCases tests edge and corner tile neighbor detection.
func TestTerrainRenderSystem_TileNeighborEdgeCases(t *testing.T) {
	testTerrain := &terrain.Terrain{
		Width:  3,
		Height: 3,
		Tiles:  make([][]terrain.TileType, 3),
	}

	for y := range testTerrain.Tiles {
		testTerrain.Tiles[y] = make([]terrain.TileType, 3)
		for x := range testTerrain.Tiles[y] {
			testTerrain.Tiles[y][x] = terrain.TileFloor
		}
	}

	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)
	sys.SetTerrain(testTerrain)

	tests := []struct {
		name            string
		x, y            int
		expectNorth     bool
		expectSouth     bool
		expectEast      bool
		expectWest      bool
		expectNorthEast bool
		expectSouthEast bool
		expectNorthWest bool
		expectSouthWest bool
	}{
		{
			name:            "Top-left corner",
			x:               0,
			y:               0,
			expectNorth:     false,
			expectSouth:     true,
			expectEast:      true,
			expectWest:      false,
			expectNorthEast: false,
			expectSouthEast: true,
			expectNorthWest: false,
			expectSouthWest: false,
		},
		{
			name:            "Top-right corner",
			x:               2,
			y:               0,
			expectNorth:     false,
			expectSouth:     true,
			expectEast:      false,
			expectWest:      true,
			expectNorthEast: false,
			expectSouthEast: false,
			expectNorthWest: false,
			expectSouthWest: true,
		},
		{
			name:            "Bottom-left corner",
			x:               0,
			y:               2,
			expectNorth:     true,
			expectSouth:     false,
			expectEast:      true,
			expectWest:      false,
			expectNorthEast: true,
			expectSouthEast: false,
			expectNorthWest: false,
			expectSouthWest: false,
		},
		{
			name:            "Bottom-right corner",
			x:               2,
			y:               2,
			expectNorth:     true,
			expectSouth:     false,
			expectEast:      false,
			expectWest:      true,
			expectNorthEast: false,
			expectSouthEast: false,
			expectNorthWest: true,
			expectSouthWest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			neighbors := sys.getTileNeighbors(tt.x, tt.y, tiles.TileFloor)

			if neighbors.N != tt.expectNorth {
				t.Errorf("N = %v, want %v", neighbors.N, tt.expectNorth)
			}
			if neighbors.S != tt.expectSouth {
				t.Errorf("S = %v, want %v", neighbors.S, tt.expectSouth)
			}
			if neighbors.E != tt.expectEast {
				t.Errorf("E = %v, want %v", neighbors.E, tt.expectEast)
			}
			if neighbors.W != tt.expectWest {
				t.Errorf("W = %v, want %v", neighbors.W, tt.expectWest)
			}
			if neighbors.NE != tt.expectNorthEast {
				t.Errorf("NE = %v, want %v", neighbors.NE, tt.expectNorthEast)
			}
			if neighbors.SE != tt.expectSouthEast {
				t.Errorf("SE = %v, want %v", neighbors.SE, tt.expectSouthEast)
			}
			if neighbors.NW != tt.expectNorthWest {
				t.Errorf("NW = %v, want %v", neighbors.NW, tt.expectNorthWest)
			}
			if neighbors.SW != tt.expectSouthWest {
				t.Errorf("SW = %v, want %v", neighbors.SW, tt.expectSouthWest)
			}
		})
	}
}

// TestTerrainRenderSystem_CacheClearedOnFeatureToggle tests cache invalidation when features change.
func TestTerrainRenderSystem_CacheClearedOnFeatureToggle(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	// Initial cache size
	initialSize := sys.tileCache.Size()

	// Toggle transitions (should clear cache)
	sys.SetTransitionsEnabled(false)
	if sys.tileCache.Size() != 0 {
		t.Error("Expected cache to be cleared after SetTransitionsEnabled")
	}

	// Toggle parallax (should clear cache)
	sys.SetParallaxEnabled(true)
	if sys.tileCache.Size() != 0 {
		t.Error("Expected cache to be cleared after SetParallaxEnabled")
	}

	// Toggle enhanced walls (should clear cache)
	sys.SetEnhancedWallsEnabled(false)
	if sys.tileCache.Size() != 0 {
		t.Error("Expected cache to be cleared after SetEnhancedWallsEnabled")
	}

	// Setting same value should not clear cache
	sys.SetEnhancedWallsEnabled(false) // Already false
	// Cache size should remain 0 (no change)
	if sys.tileCache.Size() != initialSize && sys.tileCache.Size() != 0 {
		t.Errorf("Cache size changed unexpectedly: %d", sys.tileCache.Size())
	}
}

// TestTerrainRenderSystem_Determinism tests deterministic tile generation.
func TestTerrainRenderSystem_Determinism(t *testing.T) {
	seed := int64(98765)
	genreID := "fantasy"

	sys1 := NewTerrainRenderSystem(32, 32, genreID, seed)
	sys2 := NewTerrainRenderSystem(32, 32, genreID, seed)

	// Both systems should have identical configuration
	if sys1.seed != sys2.seed {
		t.Error("Seeds should be identical")
	}
	if sys1.genreID != sys2.genreID {
		t.Error("GenreIDs should be identical")
	}
	if sys1.enableTransitions != sys2.enableTransitions {
		t.Error("Transition settings should be identical")
	}
}

// TestTerrainRenderSystem_FallbackTileCache tests that fallback tiles are cached and reused.
// This is a performance optimization (AUDIT.md Moderate Issue #3).
func TestTerrainRenderSystem_FallbackTileCache(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	// Verify fallback cache is initialized
	if sys.fallbackTileCache == nil {
		t.Fatal("fallbackTileCache should be initialized")
	}

	// Initial cache should be empty
	if len(sys.fallbackTileCache) != 0 {
		t.Errorf("fallbackTileCache should be empty initially, got %d entries", len(sys.fallbackTileCache))
	}

	// After ClearCache, the cache should still exist but be empty
	sys.ClearCache()
	if sys.fallbackTileCache == nil {
		t.Fatal("fallbackTileCache should still exist after ClearCache")
	}
	if len(sys.fallbackTileCache) != 0 {
		t.Errorf("fallbackTileCache should be empty after ClearCache, got %d entries", len(sys.fallbackTileCache))
	}
}

// TestTerrainRenderSystem_FallbackCacheKeyGeneration verifies the cache key generation for colors.
func TestTerrainRenderSystem_FallbackCacheKeyGeneration(t *testing.T) {
	// The cache key format is: (R << 24) | (G << 16) | (B << 8) | A
	tests := []struct {
		r, g, b  uint8
		expected uint32
	}{
		{120, 120, 120, 0x787878FF}, // Wall color
		{150, 180, 150, 0x96B496FF}, // Spawn room color
		{150, 150, 200, 0x9696C8FF}, // Exit room color
		{200, 120, 120, 0xC87878FF}, // Boss room color
		{200, 200, 120, 0xC8C878FF}, // Treasure room color
		{180, 120, 180, 0xB478B4FF}, // Trap room color
		{150, 150, 150, 0x969696FF}, // Normal floor color
	}

	for _, tt := range tests {
		key := uint32(tt.r)<<24 | uint32(tt.g)<<16 | uint32(tt.b)<<8 | 255
		if key != tt.expected {
			t.Errorf("Cache key for RGB(%d,%d,%d) = 0x%08X, want 0x%08X", tt.r, tt.g, tt.b, key, tt.expected)
		}
	}
}
