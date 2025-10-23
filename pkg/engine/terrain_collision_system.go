//go:build !test
// +build !test

// Package engine provides terrain collision system.
// This file implements TerrainCollisionSystem which creates collision entities
// for terrain walls to enable proper physics and movement blocking.
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TerrainCollisionSystem creates collision entities for terrain walls.
// This system converts terrain data into physical collision entities that
// can interact with the physics system.
type TerrainCollisionSystem struct {
	world       *World
	terrain     *terrain.Terrain
	tileWidth   int
	tileHeight  int
	initialized bool
	wallEntities []*Entity // Track created wall entities for cleanup
}

// NewTerrainCollisionSystem creates a new terrain collision system.
func NewTerrainCollisionSystem(world *World, tileWidth, tileHeight int) *TerrainCollisionSystem {
	return &TerrainCollisionSystem{
		world:        world,
		tileWidth:    tileWidth,
		tileHeight:   tileHeight,
		initialized:  false,
		wallEntities: make([]*Entity, 0),
	}
}

// SetTerrain sets the terrain and creates collision entities for all walls.
// This should be called after terrain generation and before the game starts.
func (t *TerrainCollisionSystem) SetTerrain(terrain *terrain.Terrain) error {
	if terrain == nil {
		return fmt.Errorf("terrain cannot be nil")
	}

	// Clean up existing wall entities if re-initializing
	t.cleanup()

	t.terrain = terrain
	t.initialized = false

	// Create collision entities for all wall tiles
	return t.initializeWallColliders()
}

// initializeWallColliders creates collision entities for all wall tiles in the terrain.
func (t *TerrainCollisionSystem) initializeWallColliders() error {
	if t.terrain == nil {
		return fmt.Errorf("terrain not set")
	}

	wallCount := 0

	// Iterate through all tiles and create collision entities for walls
	for y := 0; y < t.terrain.Height; y++ {
		for x := 0; x < t.terrain.Width; x++ {
			tileType := t.terrain.GetTile(x, y)
			
			// Create collision entity for walls only
			if tileType == terrain.TileWall {
				err := t.createWallCollider(x, y)
				if err != nil {
					return fmt.Errorf("failed to create wall collider at (%d, %d): %w", x, y, err)
				}
				wallCount++
			}
		}
	}

	t.initialized = true
	return nil
}

// createWallCollider creates a collision entity for a single wall tile.
func (t *TerrainCollisionSystem) createWallCollider(tileX, tileY int) error {
	// Create wall entity
	wallEntity := t.world.CreateEntity()

	// Calculate world position (center of tile)
	worldX := float64(tileX*t.tileWidth) + float64(t.tileWidth)/2
	worldY := float64(tileY*t.tileHeight) + float64(t.tileHeight)/2

	// Add position component
	wallEntity.AddComponent(&PositionComponent{
		X: worldX,
		Y: worldY,
	})

	// Add collision component
	wallEntity.AddComponent(&ColliderComponent{
		Width:     float64(t.tileWidth),
		Height:    float64(t.tileHeight),
		Solid:     true,
		IsTrigger: false,
		Layer:     0, // Environment layer (collides with all entities)
		OffsetX:   -float64(t.tileWidth) / 2,  // Center the collider
		OffsetY:   -float64(t.tileHeight) / 2,
	})

	// Add a component to identify this as a wall entity for debugging
	wallEntity.AddComponent(&WallComponent{
		TileX: tileX,
		TileY: tileY,
	})

	// Track this entity for cleanup
	t.wallEntities = append(t.wallEntities, wallEntity)

	return nil
}

// cleanup removes all previously created wall entities.
func (t *TerrainCollisionSystem) cleanup() {
	for _, entity := range t.wallEntities {
		t.world.RemoveEntity(entity.ID)
	}
	t.wallEntities = make([]*Entity, 0)
}

// Update is called every frame but terrain collision is static.
func (t *TerrainCollisionSystem) Update(entities []*Entity, deltaTime float64) {
	// Static terrain collision doesn't need per-frame updates
}

// IsInitialized returns whether the terrain collision system has been set up.
func (t *TerrainCollisionSystem) IsInitialized() bool {
	return t.initialized
}

// GetWallEntityCount returns the number of wall collision entities created.
func (t *TerrainCollisionSystem) GetWallEntityCount() int {
	return len(t.wallEntities)
}

// WallComponent identifies an entity as a terrain wall for debugging.
type WallComponent struct {
	TileX int
	TileY int
}

// Type implements Component interface.
func (w *WallComponent) Type() string {
	return "wall"
}