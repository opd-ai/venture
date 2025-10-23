//go:build !test
// +build !test

// Package engine provides procedural terrain rendering.
// This file implements TerrainRenderSystem which handles rendering of
// procedurally generated terrain tiles with caching for performance.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

// TerrainRenderSystem handles rendering of procedural terrain tiles.
type TerrainRenderSystem struct {
	tileCache  *TileCache
	terrain    *terrain.Terrain
	genreID    string
	seed       int64
	tileWidth  int
	tileHeight int
	tileImages map[string]*ebiten.Image // Pre-converted ebiten images
}

// NewTerrainRenderSystem creates a new terrain rendering system.
func NewTerrainRenderSystem(tileWidth, tileHeight int, genreID string, seed int64) *TerrainRenderSystem {
	return &TerrainRenderSystem{
		tileCache:  NewTileCache(1000), // Cache up to 1000 tiles (~4MB at 32x32)
		tileWidth:  tileWidth,
		tileHeight: tileHeight,
		genreID:    genreID,
		seed:       seed,
		tileImages: make(map[string]*ebiten.Image),
	}
}

// SetTerrain updates the terrain to be rendered.
func (t *TerrainRenderSystem) SetTerrain(terrain *terrain.Terrain) {
	t.terrain = terrain
}

// SetGenre updates the genre for tile generation.
func (t *TerrainRenderSystem) SetGenre(genreID string) {
	t.genreID = genreID
	// Clear tile image cache when genre changes
	t.tileImages = make(map[string]*ebiten.Image)
}

// Draw renders the terrain using the camera system for viewport culling.
func (t *TerrainRenderSystem) Draw(screen *ebiten.Image, camera *CameraSystem) {
	if t.terrain == nil {
		return
	}

	// Calculate viewport bounds in tile coordinates
	viewportMinX, viewportMinY := camera.ScreenToWorld(0, 0)
	viewportMaxX, viewportMaxY := camera.ScreenToWorld(float64(camera.ScreenWidth), float64(camera.ScreenHeight))

	// Convert to tile coordinates
	minTileX := int(viewportMinX / float64(t.tileWidth))
	minTileY := int(viewportMinY / float64(t.tileHeight))
	maxTileX := int(viewportMaxX/float64(t.tileWidth)) + 1
	maxTileY := int(viewportMaxY/float64(t.tileHeight)) + 1

	// Clamp to terrain bounds
	if minTileX < 0 {
		minTileX = 0
	}
	if minTileY < 0 {
		minTileY = 0
	}
	if maxTileX > t.terrain.Width {
		maxTileX = t.terrain.Width
	}
	if maxTileY > t.terrain.Height {
		maxTileY = t.terrain.Height
	}

	// Render visible tiles
	for y := minTileY; y < maxTileY; y++ {
		for x := minTileX; x < maxTileX; x++ {
			t.drawTile(screen, camera, x, y)
		}
	}
}

// drawTile renders a single tile at the given terrain coordinates.
func (t *TerrainRenderSystem) drawTile(screen *ebiten.Image, camera *CameraSystem, tileX, tileY int) {
	if tileX < 0 || tileX >= t.terrain.Width || tileY < 0 || tileY >= t.terrain.Height {
		return
	}

	// Get tile type from terrain
	terrainTileType := t.terrain.GetTile(tileX, tileY)

	tileType := t.terrainTileToRenderTile(terrainTileType)

	// Get or generate tile image
	img, err := t.getTileImage(tileType, tileX, tileY)
	if err != nil {
		// Fallback: render as colored rectangle
		t.drawFallbackTile(screen, camera, tileX, tileY, terrainTileType)
		return
	}

	// Calculate world position
	worldX := float64(tileX * t.tileWidth)
	worldY := float64(tileY * t.tileHeight)

	// Convert to screen coordinates
	screenX, screenY := camera.WorldToScreen(worldX, worldY)

	// Draw tile
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenX, screenY)
	screen.DrawImage(img, opts)
}

// getTileImage retrieves or generates an ebiten.Image for the given tile type.
func (t *TerrainRenderSystem) getTileImage(tileType tiles.TileType, tileX, tileY int) (*ebiten.Image, error) {
	// Create cache key
	// Use position-based variant for visual variety
	variant := float64((tileX*7+tileY*13)%100) / 100.0
	key := TileCacheKey{
		TileType: tileType,
		GenreID:  t.genreID,
		Seed:     t.seed,
		Variant:  variant,
		Width:    t.tileWidth,
		Height:   t.tileHeight,
	}

	keyStr := key.String()

	// Check if we already have an ebiten image
	if img, ok := t.tileImages[keyStr]; ok {
		return img, nil
	}

	// Get RGBA image from cache
	rgbaImg, err := t.tileCache.Get(key)
	if err != nil {
		return nil, err
	}

	// Convert to ebiten.Image
	ebitenImg := ebiten.NewImageFromImage(rgbaImg)
	t.tileImages[keyStr] = ebitenImg

	return ebitenImg, nil
}

// drawFallbackTile draws a colored rectangle when tile generation fails.
func (t *TerrainRenderSystem) drawFallbackTile(screen *ebiten.Image, camera *CameraSystem, tileX, tileY int, tileType terrain.TileType) {
	// Calculate world position
	worldX := float64(tileX * t.tileWidth)
	worldY := float64(tileY * t.tileHeight)

	// Convert to screen coordinates
	screenX, screenY := camera.WorldToScreen(worldX, worldY)

	// Create a small fallback image
	fallbackImg := ebiten.NewImage(t.tileWidth, t.tileHeight)

	// Color based on tile type and room type - make colors brighter for visibility
	var r, g, b uint8
	if tileType == terrain.TileWall {
		r, g, b = 120, 120, 120 // Brighter gray for walls (was 60,60,60)
	} else {
		// GAP-006 REPAIR: Check room type for floor color theming
		roomType := t.getRoomTypeAt(tileX, tileY)
		switch roomType {
		case terrain.RoomSpawn:
			r, g, b = 150, 180, 150 // Brighter green for spawn
		case terrain.RoomExit:
			r, g, b = 150, 150, 200 // Brighter blue for exit
		case terrain.RoomBoss:
			r, g, b = 200, 120, 120 // Brighter red for boss
		case terrain.RoomTreasure:
			r, g, b = 200, 200, 120 // Brighter gold for treasure
		case terrain.RoomTrap:
			r, g, b = 180, 120, 180 // Brighter purple for traps
		default:
			r, g, b = 150, 150, 150 // Brighter gray for normal floors
		}
	}
	fallbackImg.Fill(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenX, screenY)
	// GAP REPAIR: Remove redundant color scaling - image is already colored
	screen.DrawImage(fallbackImg, opts)
}

// getRoomTypeAt returns the room type for the tile at the given coordinates.
// Returns RoomNormal if the tile is not in any room.
func (t *TerrainRenderSystem) getRoomTypeAt(tileX, tileY int) terrain.RoomType {
	if t.terrain == nil {
		return terrain.RoomNormal
	}

	// Check which room contains this tile
	for _, room := range t.terrain.Rooms {
		if tileX >= room.X && tileX < room.X+room.Width &&
			tileY >= room.Y && tileY < room.Y+room.Height {
			return room.Type
		}
	}

	return terrain.RoomNormal
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

// Update is called every frame but terrain rendering is stateless.
func (t *TerrainRenderSystem) Update(entities []*Entity, deltaTime float64) {
	// Terrain rendering doesn't need per-frame updates
}

// GetCacheStats returns statistics about tile cache performance.
func (t *TerrainRenderSystem) GetCacheStats() (hits, misses uint64, hitRate float64) {
	h, m := t.tileCache.Stats()
	return h, m, t.tileCache.HitRate()
}

// ClearCache clears the tile cache (useful when changing genres or seeds).
func (t *TerrainRenderSystem) ClearCache() {
	t.tileCache.Clear()
	t.tileImages = make(map[string]*ebiten.Image)
}
