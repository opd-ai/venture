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

// worldToTileCoords converts world coordinates to tile coordinates.
// Phase 11.1 Week 3: Extracted helper to eliminate code duplication.
func (t *TerrainCollisionChecker) worldToTileCoords(worldX, worldY float64) (tileX, tileY int) {
	tileX = int(math.Floor(worldX / float64(t.tileWidth)))
	tileY = int(math.Floor(worldY / float64(t.tileHeight)))
	return tileX, tileY
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
// Phase 11.1 Week 3: Enhanced to support diagonal walls and multi-layer terrain.
func (t *TerrainCollisionChecker) CheckCollisionBounds(minX, minY, maxX, maxY float64) bool {
	if t.terrain == nil {
		return false
	}

	// Convert to tile coordinates using helper method
	minTileX, minTileY := t.worldToTileCoords(minX, minY)
	maxTileX, maxTileY := t.worldToTileCoords(maxX, maxY)

	// Check all tiles that the bounding box overlaps
	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			tile := t.terrain.GetTile(x, y)

			// Check standard axis-aligned walls
			if tile == terrain.TileWall {
				return true
			}

			// Check diagonal walls using triangle collision
			if tile.IsDiagonalWall() {
				if t.checkDiagonalWallCollision(x, y, tile, minX, minY, maxX, maxY) {
					return true
				}
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

// checkDiagonalWallCollision checks if a bounding box collides with a diagonal wall tile.
// Uses triangle-AABB intersection test for accurate diagonal wall collision.
// Phase 11.1 Week 3: Diagonal Wall Collision Detection
//
// Diagonal walls are represented as triangles filling half of a tile:
// - TileWallNE: Triangle from bottom-left to top-right (/)
// - TileWallNW: Triangle from bottom-right to top-left (\)
// - TileWallSE: Triangle from top-left to bottom-right (\)
// - TileWallSW: Triangle from top-right to bottom-left (/)
func (t *TerrainCollisionChecker) checkDiagonalWallCollision(
	tileX, tileY int,
	tileType terrain.TileType,
	minX, minY, maxX, maxY float64,
) bool {
	// Calculate tile bounds in world coordinates
	tileminX := float64(tileX * t.tileWidth)
	tileminY := float64(tileY * t.tileHeight)
	tilemaxX := tileminX + float64(t.tileWidth)
	tilemaxY := tileminY + float64(t.tileHeight)

	// Define triangle vertices based on diagonal wall orientation
	// Each diagonal wall fills half the tile with a triangle
	var v1X, v1Y, v2X, v2Y, v3X, v3Y float64

	switch tileType {
	case terrain.TileWallNE: // / diagonal (bottom-left to top-right)
		v1X, v1Y = tileminX, tilemaxY // Bottom-left
		v2X, v2Y = tilemaxX, tilemaxY // Bottom-right
		v3X, v3Y = tilemaxX, tileminY // Top-right
	case terrain.TileWallNW: // \ diagonal (bottom-right to top-left)
		v1X, v1Y = tileminX, tilemaxY // Bottom-left
		v2X, v2Y = tileminX, tileminY // Top-left
		v3X, v3Y = tilemaxX, tilemaxY // Bottom-right
	case terrain.TileWallSE: // \ diagonal (top-left to bottom-right)
		v1X, v1Y = tileminX, tileminY // Top-left
		v2X, v2Y = tilemaxX, tileminY // Top-right
		v3X, v3Y = tilemaxX, tilemaxY // Bottom-right
	case terrain.TileWallSW: // / diagonal (top-right to bottom-left)
		v1X, v1Y = tileminX, tileminY // Top-left
		v2X, v2Y = tileminX, tilemaxY // Bottom-left
		v3X, v3Y = tilemaxX, tileminY // Top-right
	default:
		return false
	}

	// Triangle-AABB intersection test
	// Uses separating axis theorem (SAT) for accurate collision detection
	return triangleAABBIntersection(v1X, v1Y, v2X, v2Y, v3X, v3Y, minX, minY, maxX, maxY)
}

// triangleAABBIntersection tests if a triangle intersects with an axis-aligned bounding box.
// Uses the separating axis theorem (SAT) for accurate collision detection.
// Phase 11.1 Week 3: Triangle-AABB intersection algorithm
//
// Returns true if the triangle and AABB overlap.
// Algorithm steps:
// 1. Check if any triangle vertex is inside the AABB
// 2. Check if any AABB vertex is inside the triangle
// 3. Check if any triangle edge intersects the AABB edges
func triangleAABBIntersection(
	t1X, t1Y, t2X, t2Y, t3X, t3Y float64, // Triangle vertices
	minX, minY, maxX, maxY float64, // AABB bounds
) bool {
	// Test 1: Check if any triangle vertex is inside the AABB
	if pointInAABB(t1X, t1Y, minX, minY, maxX, maxY) {
		return true
	}
	if pointInAABB(t2X, t2Y, minX, minY, maxX, maxY) {
		return true
	}
	if pointInAABB(t3X, t3Y, minX, minY, maxX, maxY) {
		return true
	}

	// Test 2: Check if any AABB corner is inside the triangle
	// AABB corners
	if pointInTriangle(minX, minY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if pointInTriangle(maxX, minY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if pointInTriangle(minX, maxY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if pointInTriangle(maxX, maxY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
		return true
	}

	// Test 3: Check if any triangle edge intersects any AABB edge
	// Phase 11.1 Week 3: Unrolled loops to avoid allocations
	// AABB edges: top, right, bottom, left
	// Triangle edges: (t1-t2), (t2-t3), (t3-t1)

	// Top edge of AABB vs all triangle edges
	if lineSegmentsIntersect(minX, minY, maxX, minY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersect(minX, minY, maxX, minY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersect(minX, minY, maxX, minY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// Right edge of AABB vs all triangle edges
	if lineSegmentsIntersect(maxX, minY, maxX, maxY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersect(maxX, minY, maxX, maxY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersect(maxX, minY, maxX, maxY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// Bottom edge of AABB vs all triangle edges
	if lineSegmentsIntersect(maxX, maxY, minX, maxY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersect(maxX, maxY, minX, maxY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersect(maxX, maxY, minX, maxY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// Left edge of AABB vs all triangle edges
	if lineSegmentsIntersect(minX, maxY, minX, minY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersect(minX, maxY, minX, minY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersect(minX, maxY, minX, minY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// No intersection found
	return false
}

// pointInAABB checks if a point is inside an axis-aligned bounding box.
func pointInAABB(x, y, minX, minY, maxX, maxY float64) bool {
	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

// pointInTriangle checks if a point is inside a triangle using barycentric coordinates.
// Phase 11.1 Week 3: Point-in-triangle test using cross products
func pointInTriangle(px, py, t1X, t1Y, t2X, t2Y, t3X, t3Y float64) bool {
	// Calculate barycentric coordinates using cross products
	// A point is inside the triangle if all three cross products have the same sign

	// Edge 1-2
	d1 := (px-t2X)*(t1Y-t2Y) - (t1X-t2X)*(py-t2Y)
	// Edge 2-3
	d2 := (px-t3X)*(t2Y-t3Y) - (t2X-t3X)*(py-t3Y)
	// Edge 3-1
	d3 := (px-t1X)*(t3Y-t1Y) - (t3X-t1X)*(py-t1Y)

	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)

	// Point is inside if all have the same sign (not both negative and positive)
	return !(hasNeg && hasPos)
}

// lineSegmentsIntersect checks if two line segments intersect.
// Phase 11.1 Week 3: Line segment intersection test
func lineSegmentsIntersect(p1X, p1Y, p2X, p2Y, q1X, q1Y, q2X, q2Y float64) bool {
	// Calculate direction vectors
	d1X := p2X - p1X
	d1Y := p2Y - p1Y
	d2X := q2X - q1X
	d2Y := q2Y - q1Y

	// Calculate cross product of directions
	cross := d1X*d2Y - d1Y*d2X

	// Parallel lines (cross product = 0) don't intersect unless collinear
	if math.Abs(cross) < 1e-10 {
		return false
	}

	// Calculate parameters for intersection point
	t := ((q1X-p1X)*d2Y - (q1Y-p1Y)*d2X) / cross
	u := ((q1X-p1X)*d1Y - (q1Y-p1Y)*d1X) / cross

	// Intersection occurs if both parameters are in [0, 1]
	return t >= 0 && t <= 1 && u >= 0 && u <= 1
}
