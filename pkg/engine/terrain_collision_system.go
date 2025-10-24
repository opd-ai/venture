// Package engine provides efficient terrain collision checking.
// This file implements terrain collision detection by extending the existing
// collision system to check terrain data directly instead of creating entities.
package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TerrainCollisionChecker provides efficient terrain collision detection.
// This integrates with the existing collision system to check terrain walls
// without creating thousands of collision entities.
type TerrainCollisionChecker struct {
	terrain    *terrain.Terrain
	tileWidth  int
	tileHeight int
}

// NewTerrainCollisionChecker creates a new terrain collision checker.
func NewTerrainCollisionChecker(tileWidth, tileHeight int) *TerrainCollisionChecker {
	return &TerrainCollisionChecker{
		tileWidth:  tileWidth,
		tileHeight: tileHeight,
	}
}

// SetTerrain sets the terrain data for collision checking.
func (t *TerrainCollisionChecker) SetTerrain(terrain *terrain.Terrain) {
	t.terrain = terrain
}

// CheckCollision checks if a world position collides with a terrain wall.
func (t *TerrainCollisionChecker) CheckCollision(worldX, worldY, width, height float64) bool {
	if t.terrain == nil {
		return false
	}

	// Calculate bounding box in world coordinates
	minX := worldX - width/2
	minY := worldY - height/2
	maxX := worldX + width/2
	maxY := worldY + height/2

	return t.CheckCollisionBounds(minX, minY, maxX, maxY)
}

// CheckCollisionBounds checks if a bounding box collides with terrain walls.
// This method accepts explicit bounds coordinates for more precise collision detection.
// minX, minY are the top-left corner of the bounding box
// maxX, maxY are the bottom-right corner of the bounding box
func (t *TerrainCollisionChecker) CheckCollisionBounds(minX, minY, maxX, maxY float64) bool {
	if t.terrain == nil {
		return false
	}

	// Convert to tile coordinates
	minTileX := int(math.Floor(minX / float64(t.tileWidth)))
	minTileY := int(math.Floor(minY / float64(t.tileHeight)))
	maxTileX := int(math.Floor(maxX / float64(t.tileWidth)))
	maxTileY := int(math.Floor(maxY / float64(t.tileHeight)))

	// Check all tiles that the bounding box overlaps
	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			if t.terrain.GetTile(x, y) == terrain.TileWall {
				return true
			}
		}
	}

	return false
}

// CheckEntityCollision checks if an entity collides with terrain walls.
func (t *TerrainCollisionChecker) CheckEntityCollision(entity *Entity) bool {
	if !entity.HasComponent("position") || !entity.HasComponent("collider") {
		return false
	}

	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	pos := posComp.(*PositionComponent)
	collider := colliderComp.(*ColliderComponent)

	return t.CheckCollision(pos.X, pos.Y, collider.Width, collider.Height)
}
