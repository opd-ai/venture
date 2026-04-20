// Package engine provides procedural terrain rendering.
// This file implements TerrainRenderSystem which handles rendering of
// procedurally generated terrain tiles with caching for performance.
package engine

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

// TerrainRenderSystem handles rendering of procedural terrain tiles.
type TerrainRenderSystem struct {
	tileCache           *TileCache
	terrain             *terrain.Terrain
	genreID             string
	seed                int64
	tileWidth           int
	tileHeight          int
	tileImages          map[string]*ebiten.Image // Pre-converted ebiten images
	enableTransitions   bool                     // Enable auto-tiling transitions
	enableParallax      bool                     // Enable parallax depth effects
	enableEnhancedWalls bool                     // Enable anti-aliased walls
	cameraX             float64                  // Camera X for parallax
	cameraY             float64                  // Camera Y for parallax
	fallbackTileCache   map[uint32]*ebiten.Image // PERF: Cache fallback tiles by RGBA color (Issue #3)
	drawTileOpts        ebiten.DrawImageOptions  // PERF V4: Pre-allocated options to avoid per-tile heap allocation
}

// NewTerrainRenderSystem creates a new terrain rendering system.
func NewTerrainRenderSystem(tileWidth, tileHeight int, genreID string, seed int64) *TerrainRenderSystem {
	return &TerrainRenderSystem{
		tileCache:           NewTileCache(1000), // Cache up to 1000 tiles (~4MB at 32x32)
		tileWidth:           tileWidth,
		tileHeight:          tileHeight,
		genreID:             genreID,
		seed:                seed,
		tileImages:          make(map[string]*ebiten.Image),
		enableTransitions:   true,  // Enable by default for smooth terrain
		enableParallax:      false, // Disable by default (performance)
		enableEnhancedWalls: true,  // Enable by default for better visuals
		fallbackTileCache:   make(map[uint32]*ebiten.Image),
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

// SetTransitionsEnabled enables or disables auto-tiling transitions.
func (t *TerrainRenderSystem) SetTransitionsEnabled(enabled bool) {
	if t.enableTransitions != enabled {
		t.enableTransitions = enabled
		t.ClearCache() // Clear cache when transition mode changes
	}
}

// SetParallaxEnabled enables or disables parallax depth effects.
func (t *TerrainRenderSystem) SetParallaxEnabled(enabled bool) {
	if t.enableParallax != enabled {
		t.enableParallax = enabled
		t.ClearCache()
	}
}

// SetEnhancedWallsEnabled enables or disables enhanced wall rendering.
func (t *TerrainRenderSystem) SetEnhancedWallsEnabled(enabled bool) {
	if t.enableEnhancedWalls != enabled {
		t.enableEnhancedWalls = enabled
		t.ClearCache()
	}
}

// SetCameraPosition updates camera position for parallax calculations.
func (t *TerrainRenderSystem) SetCameraPosition(x, y float64) {
	t.cameraX = x
	t.cameraY = y
}

// Draw renders the terrain using the camera system for viewport culling.
func (t *TerrainRenderSystem) Draw(screen *ebiten.Image, camera *CameraSystem) {
	if t.terrain == nil {
		return
	}
	if t.tileWidth <= 0 || t.tileHeight <= 0 {
		return
	}

	// Update camera position for parallax calculations
	camX, camY := camera.GetPosition()
	t.cameraX = camX
	t.cameraY = camY

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

	// Draw tile - PERF V4: Reuse pre-allocated options
	t.drawTileOpts.GeoM.Reset()
	t.drawTileOpts.GeoM.Translate(screenX, screenY)
	screen.DrawImage(img, &t.drawTileOpts)
}

// getTileImage retrieves or generates an ebiten.Image for the given tile type.
func (t *TerrainRenderSystem) getTileImage(tileType tiles.TileType, tileX, tileY int) (*ebiten.Image, error) {
	variant := float64((tileX*7+tileY*13)%100) / 100.0
	keyStr := t.buildCacheKey(tileType, variant)

	if img, ok := t.tileImages[keyStr]; ok {
		return img, nil
	}

	rgbaImg, err := t.generateTileImage(tileType, tileX, tileY, variant)
	if err != nil {
		return nil, err
	}

	ebitenImg := ebiten.NewImageFromImage(rgbaImg)
	t.tileImages[keyStr] = ebitenImg

	return ebitenImg, nil
}

// buildCacheKey creates a cache key for tile images.
func (t *TerrainRenderSystem) buildCacheKey(tileType tiles.TileType, variant float64) string {
	return fmt.Sprintf("%s-%s-%d-%.2f-%dx%d-t%v-p%v-w%v",
		tileType, t.genreID, t.seed, variant, t.tileWidth, t.tileHeight,
		t.enableTransitions, t.enableParallax, t.enableEnhancedWalls)
}

// generateTileImage generates the RGBA image for a tile using advanced features.
func (t *TerrainRenderSystem) generateTileImage(tileType tiles.TileType, tileX, tileY int, variant float64) (*image.RGBA, error) {
	if tileType == tiles.TileWall && t.enableEnhancedWalls {
		return t.generateEnhancedWall(tileX, tileY, variant)
	}

	if t.enableTransitions && (tileType == tiles.TileFloor || tileType == tiles.TileCorridor) {
		return t.generateWithTransition(tileType, tileX, tileY, variant)
	}

	if t.enableParallax {
		return t.generateWithParallax(tileType, variant)
	}

	return t.generateBasicTile(tileType, variant)
}

// generateEnhancedWall generates an enhanced wall tile with corner detection.
func (t *TerrainRenderSystem) generateEnhancedWall(tileX, tileY int, variant float64) (*image.RGBA, error) {
	neighbors := t.getWallNeighbors(tileX, tileY)
	wallConfig := tiles.EnhancedWallConfig{
		Config: tiles.Config{
			Type:    tiles.TileWall,
			Width:   t.tileWidth,
			Height:  t.tileHeight,
			GenreID: t.genreID,
			Seed:    t.seed,
			Variant: variant,
		},
		Neighbors:          neighbors,
		EnableAntialiasing: true,
		EnableShadows:      true,
		BlendRadius:        4,
	}
	return t.tileCache.gen.GenerateEnhancedWall(wallConfig)
}

// generateWithTransition generates a tile with smooth transitions.
func (t *TerrainRenderSystem) generateWithTransition(tileType tiles.TileType, tileX, tileY int, variant float64) (*image.RGBA, error) {
	neighbors := t.getTileNeighbors(tileX, tileY, tileType)
	transitionType := tiles.DetermineTransition(neighbors)
	transConfig := tiles.TransitionConfig{
		BaseConfig: tiles.Config{
			Type:    tileType,
			Width:   t.tileWidth,
			Height:  t.tileHeight,
			GenreID: t.genreID,
			Seed:    t.seed,
			Variant: variant,
		},
		Transition:   transitionType,
		Neighbors:    neighbors,
		BlendRadius:  0.3,
		CornerRadius: 0.25,
		Smoothness:   0.5,
	}
	return t.tileCache.gen.GenerateWithTransition(transConfig)
}

// generateWithParallax generates a tile with parallax depth effects.
func (t *TerrainRenderSystem) generateWithParallax(tileType tiles.TileType, variant float64) (*image.RGBA, error) {
	parallaxConfig := tiles.ParallaxConfig{
		BaseConfig: tiles.Config{
			Type:    tileType,
			Width:   t.tileWidth,
			Height:  t.tileHeight,
			GenreID: t.genreID,
			Seed:    t.seed,
			Variant: variant,
		},
		Layer:         tiles.LayerBase,
		CameraX:       t.cameraX,
		CameraY:       t.cameraY,
		ParallaxDepth: 1.0,
		AOIntensity:   0.5,
		ShadowHeight:  0.3,
		ShadowAngle:   math.Pi / 4,
	}
	return t.tileCache.gen.GenerateWithParallax(parallaxConfig)
}

// generateBasicTile generates a basic tile without advanced features.
func (t *TerrainRenderSystem) generateBasicTile(tileType tiles.TileType, variant float64) (*image.RGBA, error) {
	key := TileCacheKey{
		TileType: tileType,
		GenreID:  t.genreID,
		Seed:     t.seed,
		Variant:  variant,
		Width:    t.tileWidth,
		Height:   t.tileHeight,
	}
	return t.tileCache.Get(key)
}

// drawFallbackTile draws a colored rectangle when tile generation fails.
// PERF: Uses cached fallback images to avoid per-draw allocations (Moderate Issue #3).
func (t *TerrainRenderSystem) drawFallbackTile(screen *ebiten.Image, camera *CameraSystem, tileX, tileY int, tileType terrain.TileType) {
	// Calculate world position
	worldX := float64(tileX * t.tileWidth)
	worldY := float64(tileY * t.tileHeight)

	// Convert to screen coordinates
	screenX, screenY := camera.WorldToScreen(worldX, worldY)

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

	// PERF: Use composite key (RGBA as uint32) to cache fallback tiles
	cacheKey := uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | 255
	fallbackImg, ok := t.fallbackTileCache[cacheKey]
	if !ok {
		// Create and cache the fallback image
		fallbackImg = ebiten.NewImage(t.tileWidth, t.tileHeight)
		fallbackImg.Fill(color.RGBA{R: r, G: g, B: b, A: 255})
		t.fallbackTileCache[cacheKey] = fallbackImg
	}

	// PERF V4: Reuse pre-allocated options
	t.drawTileOpts.GeoM.Reset()
	t.drawTileOpts.GeoM.Translate(screenX, screenY)
	// GAP REPAIR: Remove redundant color scaling - image is already colored
	screen.DrawImage(fallbackImg, &t.drawTileOpts)
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
	case terrain.TileWallNE:
		return tiles.TileWallNE
	case terrain.TileWallNW:
		return tiles.TileWallNW
	case terrain.TileWallSE:
		return tiles.TileWallSE
	case terrain.TileWallSW:
		return tiles.TileWallSW
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
	// PERF: Also clear fallback tile cache when main cache is cleared
	t.fallbackTileCache = make(map[uint32]*ebiten.Image)
}

// getWallNeighbors returns which neighboring tiles are walls for enhanced wall rendering.
func (t *TerrainRenderSystem) getWallNeighbors(tileX, tileY int) tiles.WallNeighbors {
	neighbors := tiles.WallNeighbors{}

	// Check north
	if tileY > 0 && t.terrain.GetTile(tileX, tileY-1) == terrain.TileWall {
		neighbors.North = true
	}

	// Check south
	if tileY < t.terrain.Height-1 && t.terrain.GetTile(tileX, tileY+1) == terrain.TileWall {
		neighbors.South = true
	}

	// Check east
	if tileX < t.terrain.Width-1 && t.terrain.GetTile(tileX+1, tileY) == terrain.TileWall {
		neighbors.East = true
	}

	// Check west
	if tileX > 0 && t.terrain.GetTile(tileX-1, tileY) == terrain.TileWall {
		neighbors.West = true
	}

	return neighbors
}

// getTileNeighbors returns 8-directional neighbors for transition detection.
func (t *TerrainRenderSystem) getTileNeighbors(tileX, tileY int, targetType tiles.TileType) tiles.TileNeighbors {
	neighbors := tiles.TileNeighbors{}

	// Helper to check if tile matches target type
	isTargetType := func(x, y int) bool {
		if x < 0 || x >= t.terrain.Width || y < 0 || y >= t.terrain.Height {
			return false
		}
		terrainTile := t.terrain.GetTile(x, y)
		renderTile := t.terrainTileToRenderTile(terrainTile)
		return renderTile == targetType
	}

	// Check all 8 directions
	neighbors.N = isTargetType(tileX, tileY-1)
	neighbors.NE = isTargetType(tileX+1, tileY-1)
	neighbors.E = isTargetType(tileX+1, tileY)
	neighbors.SE = isTargetType(tileX+1, tileY+1)
	neighbors.S = isTargetType(tileX, tileY+1)
	neighbors.SW = isTargetType(tileX-1, tileY+1)
	neighbors.W = isTargetType(tileX-1, tileY)
	neighbors.NW = isTargetType(tileX-1, tileY-1)

	return neighbors
}
