// Package terrain provides water generation utilities for terrain generators.
// This file contains functions for creating lakes, rivers, moats, and bridges.
package terrain

import (
	"math"
	"math/rand"
)

// WaterType identifies the type of water feature.
type WaterType int

const (
	WaterLake  WaterType = iota // Circular/elliptical body of water
	WaterRiver                  // Linear flowing water feature
	WaterMoat                   // Water surrounding a room/structure
)

// String returns a human-readable name for the water type.
func (w WaterType) String() string {
	switch w {
	case WaterLake:
		return "Lake"
	case WaterRiver:
		return "River"
	case WaterMoat:
		return "Moat"
	default:
		return "Unknown"
	}
}

// WaterFeature represents a generated water feature in the terrain.
type WaterFeature struct {
	Type    WaterType // Type of water feature
	Tiles   []Point   // Coordinates of all water tiles
	Bridges []Point   // Auto-placed bridge locations
}

// GenerateLake creates a circular or elliptical lake centered at the given point.
//
// Parameters:
//   - center: Center point of the lake
//   - radius: Approximate radius in tiles (actual size varies by 10-30%)
//   - terrain: Terrain to modify
//   - rng: Random number generator for shape variation
//
// The lake uses TileWaterDeep for the center and TileWaterShallow for edges.
// Returns a WaterFeature describing the created lake.
func GenerateLake(center Point, radius int, terrain *Terrain, rng *rand.Rand) *WaterFeature {
	feature := &WaterFeature{
		Type:  WaterLake,
		Tiles: make([]Point, 0, radius*radius),
	}

	if radius < 1 {
		radius = 1
	}

	// Add shape variation (10-30% size variance)
	radiusX := float64(radius) * (0.9 + rng.Float64()*0.4)
	radiusY := float64(radius) * (0.9 + rng.Float64()*0.4)

	// Generate elliptical lake
	for dy := -radius - 2; dy <= radius+2; dy++ {
		for dx := -radius - 2; dx <= radius+2; dx++ {
			x := center.X + dx
			y := center.Y + dy

			if !terrain.IsInBounds(x, y) {
				continue
			}

			// Don't overwrite important tiles
			tile := terrain.GetTile(x, y)
			if tile == TileWall || tile == TileDoor || tile == TileStairsUp || tile == TileStairsDown {
				continue
			}

			// Calculate normalized distance from center
			normDist := math.Sqrt(float64(dx*dx)/(radiusX*radiusX) + float64(dy*dy)/(radiusY*radiusY))

			// Deep water in center, shallow at edges
			if normDist <= 0.6 {
				terrain.SetTile(x, y, TileWaterDeep)
				feature.Tiles = append(feature.Tiles, Point{x, y})
			} else if normDist <= 1.0 {
				terrain.SetTile(x, y, TileWaterShallow)
				feature.Tiles = append(feature.Tiles, Point{x, y})
			}
		}
	}

	return feature
}

// GenerateRiver creates a winding river between two points.
//
// Parameters:
//   - start: Starting point of the river
//   - end: Ending point of the river
//   - width: Width of the river in tiles (1-5)
//   - terrain: Terrain to modify
//   - rng: Random number generator for path variation
//
// The river follows a path from start to end with random meandering.
// Uses TileWaterShallow for narrow rivers, TileWaterDeep for wide centers.
// Returns a WaterFeature describing the created river.
func GenerateRiver(start, end Point, width int, terrain *Terrain, rng *rand.Rand) *WaterFeature {
	feature := &WaterFeature{
		Type:  WaterRiver,
		Tiles: make([]Point, 0, 100),
	}

	if width < 1 {
		width = 1
	}
	if width > 5 {
		width = 5
	}

	// Generate path points from start to end with meandering
	pathPoints := generateRiverPath(start, end, terrain, rng)

	// Track placed tiles to avoid duplicates
	placed := make(map[Point]bool)

	// Expand path points to river width
	for _, point := range pathPoints {
		// Place water tiles around each path point
		for dy := -width; dy <= width; dy++ {
			for dx := -width; dx <= width; dx++ {
				x := point.X + dx
				y := point.Y + dy

				if !terrain.IsInBounds(x, y) {
					continue
				}

				// Skip if already placed
				p := Point{x, y}
				if placed[p] {
					continue
				}

				// Don't overwrite important tiles
				tile := terrain.GetTile(x, y)
				if tile == TileWall || tile == TileDoor || tile == TileStairsUp || tile == TileStairsDown {
					continue
				}

				// Calculate distance from path center
				dist := math.Sqrt(float64(dx*dx + dy*dy))

				// Deep water in center, shallow at edges
				if width > 2 && dist <= float64(width)*0.5 {
					terrain.SetTile(x, y, TileWaterDeep)
					feature.Tiles = append(feature.Tiles, p)
					placed[p] = true
				} else if dist <= float64(width) {
					terrain.SetTile(x, y, TileWaterShallow)
					feature.Tiles = append(feature.Tiles, p)
					placed[p] = true
				}
			}
		}
	}

	return feature
}

// generateRiverPath creates a winding path between two points.
func generateRiverPath(start, end Point, terrain *Terrain, rng *rand.Rand) []Point {
	path := make([]Point, 0, 50)

	// Calculate direction vector
	dx := end.X - start.X
	dy := end.Y - start.Y
	distance := math.Sqrt(float64(dx*dx + dy*dy))

	if distance < 1 {
		return []Point{start}
	}

	// Step size based on distance
	steps := int(distance)
	if steps < 5 {
		steps = 5
	}

	stepX := float64(dx) / float64(steps)
	stepY := float64(dy) / float64(steps)

	// Generate path with meandering
	for i := 0; i <= steps; i++ {
		// Base position
		baseX := float64(start.X) + stepX*float64(i)
		baseY := float64(start.Y) + stepY*float64(i)

		// Add perpendicular offset for meandering
		perpendicularDist := int(distance * 0.2) // 20% of total distance
		if perpendicularDist < 2 {
			perpendicularDist = 2
		}

		offset := rng.Intn(perpendicularDist*2+1) - perpendicularDist

		// Calculate perpendicular direction
		perpX := -stepY
		perpY := stepX
		perpLen := math.Sqrt(perpX*perpX + perpY*perpY)
		if perpLen > 0 {
			perpX /= perpLen
			perpY /= perpLen
		}

		// Apply offset
		x := int(baseX + perpX*float64(offset))
		y := int(baseY + perpY*float64(offset))

		// Clamp to terrain bounds
		if x < 0 {
			x = 0
		}
		if x >= terrain.Width {
			x = terrain.Width - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= terrain.Height {
			y = terrain.Height - 1
		}

		path = append(path, Point{x, y})
	}

	return path
}

// GenerateMoat creates a water-filled moat around a room.
//
// Parameters:
//   - room: Room to surround with water
//   - width: Width of the moat in tiles (1-3)
//   - terrain: Terrain to modify
//
// The moat is placed width tiles outside the room boundaries.
// Uses TileWaterShallow for narrow moats, mixed shallow/deep for wide moats.
// Returns a WaterFeature describing the created moat.
func GenerateMoat(room *Room, width int, terrain *Terrain) *WaterFeature {
	feature := &WaterFeature{
		Type:  WaterMoat,
		Tiles: make([]Point, 0, (room.Width+room.Height)*2*width),
	}

	if width < 1 {
		width = 1
	}
	if width > 3 {
		width = 3
	}

	// Place water tiles around room perimeter
	for dy := -width; dy <= room.Height+width-1; dy++ {
		for dx := -width; dx <= room.Width+width-1; dx++ {
			// Skip interior of room
			if dx >= 0 && dx < room.Width && dy >= 0 && dy < room.Height {
				continue
			}

			x := room.X + dx
			y := room.Y + dy

			if !terrain.IsInBounds(x, y) {
				continue
			}

			// Calculate distance from room edge
			minDistX := 0
			if dx < 0 {
				minDistX = -dx
			} else if dx >= room.Width {
				minDistX = dx - room.Width + 1
			}

			minDistY := 0
			if dy < 0 {
				minDistY = -dy
			} else if dy >= room.Height {
				minDistY = dy - room.Height + 1
			}

			dist := minDistX
			if minDistY > dist {
				dist = minDistY
			}

			// Place water based on distance (only if within moat width)
			if dist > 0 && dist <= width {
				if width > 1 && dist <= width/2 {
					terrain.SetTile(x, y, TileWaterDeep)
				} else {
					terrain.SetTile(x, y, TileWaterShallow)
				}
				feature.Tiles = append(feature.Tiles, Point{x, y})
			}
		}
	}

	return feature
}

// PlaceBridges automatically places bridges where paths cross water.
//
// Parameters:
//   - feature: Water feature to place bridges over
//   - terrain: Terrain to modify
//   - rng: Random number generator (unused, for future variations)
//
// Bridges are placed where TileCorridor or TileFloor tiles are adjacent to
// water tiles, creating crossings. Uses TileBridge tile type.
// Updates feature.Bridges with placed bridge locations.
func PlaceBridges(feature *WaterFeature, terrain *Terrain, rng *rand.Rand) {
	bridgeLocations := make(map[Point]bool)

	// Find water tiles adjacent to paths
	for _, waterTile := range feature.Tiles {
		// Check all 8 neighbors
		for _, neighbor := range waterTile.Neighbors() {
			if !terrain.IsInBounds(neighbor.X, neighbor.Y) {
				continue
			}

			tile := terrain.GetTile(neighbor.X, neighbor.Y)

			// If neighbor is a walkable path tile
			if tile == TileCorridor || tile == TileFloor {
				// Check if this water tile is between two path tiles
				hasPathOnOppositeSide := false

				// Check opposite neighbors
				dx := neighbor.X - waterTile.X
				dy := neighbor.Y - waterTile.Y

				opposite := Point{waterTile.X - dx, waterTile.Y - dy}
				if terrain.IsInBounds(opposite.X, opposite.Y) {
					oppositeTile := terrain.GetTile(opposite.X, opposite.Y)
					if oppositeTile == TileCorridor || oppositeTile == TileFloor {
						hasPathOnOppositeSide = true
					}
				}

				// Place bridge if water is between paths
				if hasPathOnOppositeSide {
					bridgeLocations[waterTile] = true
				}
			}
		}
	}

	// Apply bridges to terrain
	for bridgePoint := range bridgeLocations {
		terrain.SetTile(bridgePoint.X, bridgePoint.Y, TileBridge)
		feature.Bridges = append(feature.Bridges, bridgePoint)
	}
}

// FloodFill performs a flood fill starting from a point, filling up to maxTiles.
//
// Parameters:
//   - start: Starting point for flood fill
//   - maxTiles: Maximum number of tiles to fill (limits lake size)
//   - terrain: Terrain to read walkability from
//
// The flood fill only traverses walkable tiles (TileFloor, TileCorridor, TileDoor).
// Returns a slice of all reachable points within the maxTiles limit.
// Useful for creating natural lake shapes or validating connectivity.
func FloodFill(start Point, maxTiles int, terrain *Terrain) []Point {
	if !terrain.IsInBounds(start.X, start.Y) {
		return nil
	}

	// Check if start tile is walkable
	startTile := terrain.GetTile(start.X, start.Y)
	if startTile != TileFloor && startTile != TileCorridor && startTile != TileDoor {
		return nil
	}

	filled := make([]Point, 0, maxTiles)
	visited := make(map[Point]bool)
	queue := []Point{start}
	visited[start] = true

	for len(queue) > 0 && len(filled) < maxTiles {
		current := queue[0]
		queue = queue[1:]

		filled = append(filled, current)

		// Check all 4 cardinal neighbors
		neighbors := []Point{
			{current.X - 1, current.Y},
			{current.X + 1, current.Y},
			{current.X, current.Y - 1},
			{current.X, current.Y + 1},
		}

		for _, neighbor := range neighbors {
			if !terrain.IsInBounds(neighbor.X, neighbor.Y) {
				continue
			}

			if visited[neighbor] {
				continue
			}

			tile := terrain.GetTile(neighbor.X, neighbor.Y)

			// Only traverse walkable tiles
			if tile == TileFloor || tile == TileCorridor || tile == TileDoor {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return filled
}

// FloodFillWater is a specialized flood fill that creates natural-looking water bodies.
//
// Parameters:
//   - start: Starting point for flood fill
//   - maxTiles: Maximum number of tiles to fill
//   - deepWaterRatio: Ratio of deep to shallow water (0.0-1.0)
//   - terrain: Terrain to modify
//   - rng: Random number generator for variation
//
// Creates water by flood-filling from the start point. Uses deepWaterRatio to
// determine the proportion of TileWaterDeep vs TileWaterShallow tiles.
// Returns a WaterFeature describing the created water body.
func FloodFillWater(start Point, maxTiles int, deepWaterRatio float64, terrain *Terrain, rng *rand.Rand) *WaterFeature {
	feature := &WaterFeature{
		Type:  WaterLake,
		Tiles: make([]Point, 0, maxTiles),
	}

	if !terrain.IsInBounds(start.X, start.Y) {
		return feature
	}

	// Get reachable tiles
	reachable := FloodFill(start, maxTiles, terrain)

	// Convert to water
	for _, point := range reachable {
		// Randomly choose deep or shallow based on ratio
		if rng.Float64() < deepWaterRatio {
			terrain.SetTile(point.X, point.Y, TileWaterDeep)
		} else {
			terrain.SetTile(point.X, point.Y, TileWaterShallow)
		}
		feature.Tiles = append(feature.Tiles, point)
	}

	return feature
}
