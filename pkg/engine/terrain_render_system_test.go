//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

// TerrainRenderSystem stub for testing (actual implementation has !test build tag)
type TerrainRenderSystem struct {
	tileCache  *TileCache
	terrain    *terrain.Terrain
	genreID    string
	seed       int64
	tileWidth  int
	tileHeight int
	tileImages map[string]interface{}
}

// NewTerrainRenderSystem creates a new terrain rendering system stub for testing.
func NewTerrainRenderSystem(tileWidth, tileHeight int, genreID string, seed int64) *TerrainRenderSystem {
	return &TerrainRenderSystem{
		tileCache:  NewTileCache(1000),
		tileWidth:  tileWidth,
		tileHeight: tileHeight,
		genreID:    genreID,
		seed:       seed,
		tileImages: make(map[string]interface{}),
	}
}

// SetTerrain updates the terrain to be rendered.
func (t *TerrainRenderSystem) SetTerrain(terrain *terrain.Terrain) {
	t.terrain = terrain
}

// SetGenre updates the genre for tile generation.
func (t *TerrainRenderSystem) SetGenre(genreID string) {
	t.genreID = genreID
	t.tileImages = make(map[string]interface{})
}

// ClearCache clears the tile cache.
func (t *TerrainRenderSystem) ClearCache() {
	t.tileCache.Clear()
	t.tileImages = make(map[string]interface{})
}

// GetCacheStats returns statistics about tile cache performance.
func (t *TerrainRenderSystem) GetCacheStats() (hits, misses uint64, hitRate float64) {
	h, m := t.tileCache.Stats()
	return h, m, t.tileCache.HitRate()
}

// terrainTileToRenderTile converts a terrain.TileType to a tiles.TileType.
func (t *TerrainRenderSystem) terrainTileToRenderTile(tileType terrain.TileType) tiles.TileType {
	switch tileType {
	case terrain.TileWall:
		return tiles.TileWall
	case terrain.TileFloor:
		return tiles.TileFloor
	case terrain.TileDoor:
		return tiles.TileDoor
	case terrain.TileCorridor:
		return tiles.TileCorridor
	default:
		return tiles.TileFloor
	}
}

// Test functions

func TestTerrainTileToRenderTile(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	tests := []struct {
		name     string
		input    terrain.TileType
		expected tiles.TileType
	}{
		{"Wall", terrain.TileWall, tiles.TileWall},
		{"Floor", terrain.TileFloor, tiles.TileFloor},
		{"Door", terrain.TileDoor, tiles.TileDoor},
		{"Corridor", terrain.TileCorridor, tiles.TileCorridor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.terrainTileToRenderTile(tt.input)
			if result != tt.expected {
				t.Errorf("terrainTileToRenderTile(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTerrainRenderSystem_SetTerrain(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	// Generate a simple terrain
	gen := terrain.NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Failed to generate terrain: %v", err)
	}

	terr := result.(*terrain.Terrain)

	// Set terrain
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("SetTerrain did not update terrain correctly")
	}
}

func TestTerrainRenderSystem_SetGenre(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	// Pre-populate tile images
	sys.tileImages["test-key"] = nil

	if len(sys.tileImages) != 1 {
		t.Error("Expected 1 tile image before genre change")
	}

	// Change genre
	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("Expected genreID to be 'scifi', got %s", sys.genreID)
	}

	if len(sys.tileImages) != 0 {
		t.Errorf("Expected tile images to be cleared, got %d entries", len(sys.tileImages))
	}
}

func TestTerrainRenderSystem_ClearCache(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	// Pre-populate caches
	sys.tileImages["test-key"] = nil

	// Add some entries to tile cache
	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}
	sys.tileCache.Get(key)

	if sys.tileCache.Size() != 1 {
		t.Error("Expected tile cache to have 1 entry")
	}
	if len(sys.tileImages) != 1 {
		t.Error("Expected 1 tile image")
	}

	// Clear cache
	sys.ClearCache()

	if sys.tileCache.Size() != 0 {
		t.Errorf("Expected tile cache to be empty, got %d entries", sys.tileCache.Size())
	}
	if len(sys.tileImages) != 0 {
		t.Errorf("Expected tile images to be cleared, got %d entries", len(sys.tileImages))
	}
}

func TestTerrainRenderSystem_GetCacheStats(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	// Initial stats should be zero
	hits, misses, hitRate := sys.GetCacheStats()
	if hits != 0 || misses != 0 || hitRate != 0.0 {
		t.Errorf("Expected initial stats to be zero, got hits=%d, misses=%d, hitRate=%.1f", hits, misses, hitRate)
	}

	// Access cache to generate stats
	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}

	// First access - miss
	sys.tileCache.Get(key)
	hits, misses, hitRate = sys.GetCacheStats()
	if hits != 0 || misses != 1 {
		t.Errorf("Expected 0 hits, 1 miss, got hits=%d, misses=%d", hits, misses)
	}

	// Second access - hit
	sys.tileCache.Get(key)
	hits, misses, hitRate = sys.GetCacheStats()
	if hits != 1 || misses != 1 {
		t.Errorf("Expected 1 hit, 1 miss, got hits=%d, misses=%d", hits, misses)
	}
	if hitRate != 50.0 {
		t.Errorf("Expected 50%% hit rate, got %.1f%%", hitRate)
	}
}

func TestNewTerrainRenderSystem(t *testing.T) {
	sys := NewTerrainRenderSystem(32, 32, "fantasy", 12345)

	if sys == nil {
		t.Fatal("Expected non-nil TerrainRenderSystem")
	}
	if sys.tileWidth != 32 {
		t.Errorf("Expected tileWidth=32, got %d", sys.tileWidth)
	}
	if sys.tileHeight != 32 {
		t.Errorf("Expected tileHeight=32, got %d", sys.tileHeight)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Expected genreID='fantasy', got %s", sys.genreID)
	}
	if sys.seed != 12345 {
		t.Errorf("Expected seed=12345, got %d", sys.seed)
	}
	if sys.tileCache == nil {
		t.Error("Expected non-nil tile cache")
	}
	if sys.tileImages == nil {
		t.Error("Expected non-nil tile images map")
	}
}
