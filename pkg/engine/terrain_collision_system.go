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
	return t.CheckCollisionBoundsWithLayer(minX, minY, maxX, maxY, 0)
}

// CheckCollisionBoundsWithLayer checks if a bounding box collides with terrain walls
// on a specific layer. Entities only collide with terrain on their current layer.
// Phase 11.1: Multi-layer terrain collision detection.
//
// Layer semantics:
//   - Layer 0 (ground): Normal floor/wall tiles
//   - Layer 1 (water/pit): Deep water, pits block movement unless entity can swim/fly
//   - Layer 2 (platform): Elevated platforms, entities on ground don't collide
func (t *TerrainCollisionChecker) CheckCollisionBoundsWithLayer(minX, minY, maxX, maxY float64, layer int) bool {
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

			// Skip if tile is not on the entity's layer
			if !t.tileMatchesLayer(tile, layer) {
				continue
			}

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

			// Check pits - block movement unless on platform layer or flying
			if tile == terrain.TilePit && layer != 2 {
				return true
			}
		}
	}

	return false
}

// tileMatchesLayer checks if a tile type is relevant for collision on the given layer.
// Phase 11.1: Multi-layer terrain support.
func (t *TerrainCollisionChecker) tileMatchesLayer(tile terrain.TileType, layer int) bool {
	switch layer {
	case 0: // Ground layer
		// Ground entities collide with walls, diagonal walls, and pits
		return tile == terrain.TileWall ||
			tile.IsDiagonalWall() ||
			tile == terrain.TilePit
	case 1: // Water/Pit layer
		// Water layer entities collide with walls and diagonal walls
		// They don't collide with pits (they're IN the pit)
		return tile == terrain.TileWall || tile.IsDiagonalWall()
	case 2: // Platform layer
		// Platform entities collide with walls and diagonal walls
		// They don't collide with pits (they're above them)
		return tile == terrain.TileWall || tile.IsDiagonalWall()
	default:
		return false
	}
}

// CheckEntityCollision checks if an entity collides with terrain walls.
// Phase 11.1: Enhanced to support multi-layer terrain collision.
func (t *TerrainCollisionChecker) CheckEntityCollision(entity *Entity) bool {
	if !entity.HasComponent("position") || !entity.HasComponent("collider") {
		return false
	}

	posComp, _ := entity.GetComponent("position")
	colliderComp, _ := entity.GetComponent("collider")

	pos := posComp.(*PositionComponent)
	collider := colliderComp.(*ColliderComponent)

	// Get entity's layer (default to ground layer if no layer component)
	layer := 0
	if entity.HasComponent("layer") {
		layerComp, _ := entity.GetComponent("layer")
		layerComponent := layerComp.(*LayerComponent)
		layer = layerComponent.GetEffectiveLayer()
	}

	return t.CheckCollisionWithLayer(pos.X, pos.Y, collider.Width, collider.Height, layer)
}

// CheckCollisionWithLayer checks if a world position collides with terrain
// considering the entity's layer. Phase 11.1: Multi-layer collision support.
func (t *TerrainCollisionChecker) CheckCollisionWithLayer(worldX, worldY, width, height float64, layer int) bool {
	if t.terrain == nil {
		return false
	}

	// Calculate bounding box in world coordinates
	minX := worldX - width/2
	minY := worldY - height/2
	maxX := worldX + width/2
	maxY := worldY + height/2

	return t.CheckCollisionBoundsWithLayer(minX, minY, maxX, maxY, layer)
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
	// Check if AABB is degenerate (zero or near-zero size)
	// For points or very small AABBs, use inclusive checks
	epsilon := 1e-10
	isPoint := (maxX-minX < epsilon) && (maxY-minY < epsilon)

	// Test 1: Check if any triangle vertex is strictly inside the AABB (not just touching boundary)
	// Phase 11.1 Week 3: Use strict inequality to exclude adjacent (touching) shapes
	if pointInAABBStrict(t1X, t1Y, minX, minY, maxX, maxY) {
		return true
	}
	if pointInAABBStrict(t2X, t2Y, minX, minY, maxX, maxY) {
		return true
	}
	if pointInAABBStrict(t3X, t3Y, minX, minY, maxX, maxY) {
		return true
	}

	// Test 2: Check if any AABB corner is inside the triangle
	// For points (zero-size AABBs), use non-strict to catch points on edges
	// For regular AABBs, use strict to exclude adjacent touching
	if isPoint {
		// Point AABB: check if point is in triangle (including boundary)
		if pointInTriangle(minX, minY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
			return true
		}
	} else {
		// Regular AABB: check if corners are strictly inside triangle
		if pointInTriangleStrict(minX, minY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
			return true
		}
		if pointInTriangleStrict(maxX, minY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
			return true
		}
		if pointInTriangleStrict(minX, maxY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
			return true
		}
		if pointInTriangleStrict(maxX, maxY, t1X, t1Y, t2X, t2Y, t3X, t3Y) {
			return true
		}
	}

	// Test 3: Check if any triangle edge intersects any AABB edge
	// Phase 11.1 Week 3: Use strict intersection to exclude adjacent (touching) shapes
	// AABB edges: top, right, bottom, left
	// Triangle edges: (t1-t2), (t2-t3), (t3-t1)

	// Top edge of AABB vs all triangle edges
	if lineSegmentsIntersectStrict(minX, minY, maxX, minY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersectStrict(minX, minY, maxX, minY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersectStrict(minX, minY, maxX, minY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// Right edge of AABB vs all triangle edges
	if lineSegmentsIntersectStrict(maxX, minY, maxX, maxY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersectStrict(maxX, minY, maxX, maxY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersectStrict(maxX, minY, maxX, maxY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// Bottom edge of AABB vs all triangle edges
	if lineSegmentsIntersectStrict(maxX, maxY, minX, maxY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersectStrict(maxX, maxY, minX, maxY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersectStrict(maxX, maxY, minX, maxY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// Left edge of AABB vs all triangle edges
	if lineSegmentsIntersectStrict(minX, maxY, minX, minY, t1X, t1Y, t2X, t2Y) {
		return true
	}
	if lineSegmentsIntersectStrict(minX, maxY, minX, minY, t2X, t2Y, t3X, t3Y) {
		return true
	}
	if lineSegmentsIntersectStrict(minX, maxY, minX, minY, t3X, t3Y, t1X, t1Y) {
		return true
	}

	// No intersection found
	return false
}

// pointInAABB checks if a point is inside an axis-aligned bounding box.
func pointInAABB(x, y, minX, minY, maxX, maxY float64) bool {
	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

// pointInAABBStrict checks if a point is strictly inside an AABB (not on boundary).
// Phase 11.1 Week 3: Strict version excludes boundary points to avoid false positives
// for adjacent (touching) shapes in collision detection.
func pointInAABBStrict(x, y, minX, minY, maxX, maxY float64) bool {
	return x > minX && x < maxX && y > minY && y < maxY
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

// pointInTriangleStrict checks if a point is strictly inside a triangle (not on edge).
// Phase 11.1 Week 3: Strict version excludes edge points to avoid false positives
// for adjacent (touching) shapes in collision detection.
func pointInTriangleStrict(px, py, t1X, t1Y, t2X, t2Y, t3X, t3Y float64) bool {
	// Calculate barycentric coordinates using cross products
	// A point is strictly inside if all cross products are strictly positive or strictly negative

	// Edge 1-2
	d1 := (px-t2X)*(t1Y-t2Y) - (t1X-t2X)*(py-t2Y)
	// Edge 2-3
	d2 := (px-t3X)*(t2Y-t3Y) - (t2X-t3X)*(py-t3Y)
	// Edge 3-1
	d3 := (px-t1X)*(t3Y-t1Y) - (t3X-t1X)*(py-t1Y)

	// Point is strictly inside if all have the same sign AND none are zero
	if d1 > 0 && d2 > 0 && d3 > 0 {
		return true
	}
	if d1 < 0 && d2 < 0 && d3 < 0 {
		return true
	}
	return false
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

	// Parallel lines (cross product ≈ 0)
	if math.Abs(cross) < 1e-10 {
		// Check for collinear segments that overlap
		// For collinear segments, check if they share any points

		// Check if segments are on the same line using cross product with a connecting vector
		dx := q1X - p1X
		dy := q1Y - p1Y
		crossP := dx*d1Y - dy*d1X

		if math.Abs(crossP) > 1e-10 {
			// Not on the same line
			return false
		}

		// Segments are collinear, check for overlap
		// Project all points onto the line direction and check for overlap
		// Use the primary direction (larger component) for projection
		if math.Abs(d1X) > math.Abs(d1Y) {
			// Project onto X axis
			p1 := p1X
			p2 := p2X
			q1 := q1X
			q2 := q2X

			// Ensure p1 <= p2 and q1 <= q2
			if p1 > p2 {
				p1, p2 = p2, p1
			}
			if q1 > q2 {
				q1, q2 = q2, q1
			}

			// Check for overlap: segments overlap if max(p1, q1) <= min(p2, q2)
			return math.Max(p1, q1) <= math.Min(p2, q2)
		} else {
			// Project onto Y axis
			p1 := p1Y
			p2 := p2Y
			q1 := q1Y
			q2 := q2Y

			// Ensure p1 <= p2 and q1 <= q2
			if p1 > p2 {
				p1, p2 = p2, p1
			}
			if q1 > q2 {
				q1, q2 = q2, q1
			}

			// Check for overlap: segments overlap if max(p1, q1) <= min(p2, q2)
			return math.Max(p1, q1) <= math.Min(p2, q2)
		}
	}

	// Calculate parameters for intersection point
	t := ((q1X-p1X)*d2Y - (q1Y-p1Y)*d2X) / cross
	u := ((q1X-p1X)*d1Y - (q1Y-p1Y)*d1X) / cross

	// Intersection occurs if both parameters are in [0, 1]
	return t >= 0 && t <= 1 && u >= 0 && u <= 1
}

// lineSegmentsIntersectStrict checks if two line segments have a proper intersection.
// Phase 11.1 Week 3: Strict version excludes endpoint touching to avoid false positives
// for adjacent (touching) shapes in collision detection. Only returns true for proper
// intersections where segments actually cross or overlap with non-zero length.
func lineSegmentsIntersectStrict(p1X, p1Y, p2X, p2Y, q1X, q1Y, q2X, q2Y float64) bool {
	// Calculate direction vectors
	d1X := p2X - p1X
	d1Y := p2Y - p1Y
	d2X := q2X - q1X
	d2Y := q2Y - q1Y

	// Calculate cross product of directions
	cross := d1X*d2Y - d1Y*d2X

	// Parallel lines (cross product ≈ 0)
	if math.Abs(cross) < 1e-10 {
		// For collinear segments, check for proper overlap (not just touching at endpoints)

		// Check if segments are on the same line
		dx := q1X - p1X
		dy := q1Y - p1Y
		crossP := dx*d1Y - dy*d1X

		if math.Abs(crossP) > 1e-10 {
			// Not on the same line
			return false
		}

		// Segments are collinear, check for proper overlap (strict inequality)
		// Use the primary direction (larger component) for projection
		if math.Abs(d1X) > math.Abs(d1Y) {
			// Project onto X axis
			p1 := p1X
			p2 := p2X
			q1 := q1X
			q2 := q2X

			// Ensure p1 <= p2 and q1 <= q2
			if p1 > p2 {
				p1, p2 = p2, p1
			}
			if q1 > q2 {
				q1, q2 = q2, q1
			}

			// Proper overlap requires: max(p1, q1) < min(p2, q2)
			return math.Max(p1, q1) < math.Min(p2, q2)
		} else {
			// Project onto Y axis
			p1 := p1Y
			p2 := p2Y
			q1 := q1Y
			q2 := q2Y

			// Ensure p1 <= p2 and q1 <= q2
			if p1 > p2 {
				p1, p2 = p2, p1
			}
			if q1 > q2 {
				q1, q2 = q2, q1
			}

			// Proper overlap requires: max(p1, q1) < min(p2, q2)
			return math.Max(p1, q1) < math.Min(p2, q2)
		}
	}

	// Calculate parameters for intersection point
	t := ((q1X-p1X)*d2Y - (q1Y-p1Y)*d2X) / cross
	u := ((q1X-p1X)*d1Y - (q1Y-p1Y)*d1X) / cross

	// Proper intersection requires strict inequality (not at endpoints)
	return t > 0 && t < 1 && u > 0 && u < 1
}
