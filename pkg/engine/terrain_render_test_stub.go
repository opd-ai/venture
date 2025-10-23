//go:build test
// +build test

// Package engine provides test stubs for TerrainRenderSystem.
package engine

import "github.com/opd-ai/venture/pkg/procgen/terrain"

// TerrainRenderSystem renders procedural terrain tiles (test stub).
type TerrainRenderSystem struct {
	TileWidth  int
	TileHeight int
	GenreID    string
	Seed       int64
}

// NewTerrainRenderSystem creates a new terrain render system (test stub).
func NewTerrainRenderSystem(tileWidth, tileHeight int, genreID string, seed int64) *TerrainRenderSystem {
	return &TerrainRenderSystem{
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
		GenreID:    genreID,
		Seed:       seed,
	}
}

// SetTerrain sets the terrain to render (test stub).
func (t *TerrainRenderSystem) SetTerrain(terrain *terrain.Terrain) {
	// Stub - no op in tests
}

// Draw renders terrain tiles (test stub).
func (t *TerrainRenderSystem) Draw(screen interface{}, cameraSystem *CameraSystem) {
	// Stub - no op in tests
}
