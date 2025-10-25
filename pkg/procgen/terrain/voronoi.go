// Package terrain provides Voronoi diagram utilities for composite terrain generation.
package terrain

import (
	"math"
	"math/rand"
)

// VoronoiRegion represents a region in a Voronoi diagram.
type VoronoiRegion struct {
	Seed  Point   // Center point of the region
	Tiles []Point // All tiles belonging to this region
	ID    int     // Unique identifier for the region
}

// VoronoiDiagram represents a complete Voronoi partitioning of the terrain.
type VoronoiDiagram struct {
	Regions    []*VoronoiRegion
	Assignment [][]int // 2D grid mapping (x,y) -> region ID
	Width      int
	Height     int
}

// GenerateVoronoiDiagram creates a Voronoi diagram with the specified number of regions.
// Uses Manhattan distance for performance. Regions are evenly distributed spatially.
func GenerateVoronoiDiagram(width, height, numRegions int, rng *rand.Rand) *VoronoiDiagram {
	if numRegions < 1 {
		numRegions = 1
	}
	if numRegions > width*height {
		numRegions = width * height
	}

	// Generate seed points with spatial distribution
	seeds := generateSeedPoints(width, height, numRegions, rng)

	// Initialize regions
	regions := make([]*VoronoiRegion, numRegions)
	for i := 0; i < numRegions; i++ {
		regions[i] = &VoronoiRegion{
			Seed:  seeds[i],
			Tiles: make([]Point, 0),
			ID:    i,
		}
	}

	// Create assignment grid
	assignment := make([][]int, height)
	for y := range assignment {
		assignment[y] = make([]int, width)
	}

	// Assign each tile to nearest seed (Manhattan distance)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			minDist := math.MaxInt32
			closestRegion := 0

			for i, seed := range seeds {
				dist := manhattanDistance(x, y, seed.X, seed.Y)
				if dist < minDist {
					minDist = dist
					closestRegion = i
				}
			}

			assignment[y][x] = closestRegion
			regions[closestRegion].Tiles = append(regions[closestRegion].Tiles, Point{X: x, Y: y})
		}
	}

	return &VoronoiDiagram{
		Regions:    regions,
		Assignment: assignment,
		Width:      width,
		Height:     height,
	}
}

// generateSeedPoints creates evenly distributed seed points using Poisson disc sampling.
// This provides better spatial distribution than pure random placement.
func generateSeedPoints(width, height, count int, rng *rand.Rand) []Point {
	if count == 1 {
		// Single region, place in center
		return []Point{{X: width / 2, Y: height / 2}}
	}

	// Use grid-based placement for even distribution
	seeds := make([]Point, 0, count)

	// Calculate grid dimensions
	cols := int(math.Sqrt(float64(count))) + 1
	rows := (count + cols - 1) / cols

	cellWidth := width / cols
	cellHeight := height / rows

	for i := 0; i < count; i++ {
		row := i / cols
		col := i % cols

		// Place seed in cell with random offset
		baseX := col * cellWidth
		baseY := row * cellHeight

		offsetX := rng.Intn(cellWidth)
		offsetY := rng.Intn(cellHeight)

		x := baseX + offsetX
		y := baseY + offsetY

		// Clamp to bounds
		if x >= width {
			x = width - 1
		}
		if y >= height {
			y = height - 1
		}

		seeds = append(seeds, Point{X: x, Y: y})
	}

	return seeds
}

// manhattanDistance calculates Manhattan distance between two points.
func manhattanDistance(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// GetRegion returns the region ID for a given coordinate.
func (vd *VoronoiDiagram) GetRegion(x, y int) int {
	if x < 0 || x >= vd.Width || y < 0 || y >= vd.Height {
		return -1
	}
	return vd.Assignment[y][x]
}

// FindBoundaryTiles returns all tiles that are on the boundary between regions.
// A tile is a boundary if any of its 4 neighbors belongs to a different region.
func (vd *VoronoiDiagram) FindBoundaryTiles() []Point {
	boundaries := make([]Point, 0)
	visited := make(map[Point]bool)

	for y := 0; y < vd.Height; y++ {
		for x := 0; x < vd.Width; x++ {
			point := Point{X: x, Y: y}
			if visited[point] {
				continue
			}

			currentRegion := vd.Assignment[y][x]

			// Check 4-connected neighbors
			neighbors := []Point{
				{X: x - 1, Y: y},
				{X: x + 1, Y: y},
				{X: x, Y: y - 1},
				{X: x, Y: y + 1},
			}

			isBoundary := false
			for _, n := range neighbors {
				if n.X >= 0 && n.X < vd.Width && n.Y >= 0 && n.Y < vd.Height {
					if vd.Assignment[n.Y][n.X] != currentRegion {
						isBoundary = true
						break
					}
				}
			}

			if isBoundary {
				boundaries = append(boundaries, point)
				visited[point] = true
			}
		}
	}

	return boundaries
}

// GetRegionBounds returns the bounding box for a specific region.
func (vd *VoronoiDiagram) GetRegionBounds(regionID int) (minX, minY, maxX, maxY int) {
	if regionID < 0 || regionID >= len(vd.Regions) {
		return 0, 0, 0, 0
	}

	region := vd.Regions[regionID]
	if len(region.Tiles) == 0 {
		return 0, 0, 0, 0
	}

	minX, minY = vd.Width, vd.Height
	maxX, maxY = 0, 0

	for _, tile := range region.Tiles {
		if tile.X < minX {
			minX = tile.X
		}
		if tile.X > maxX {
			maxX = tile.X
		}
		if tile.Y < minY {
			minY = tile.Y
		}
		if tile.Y > maxY {
			maxY = tile.Y
		}
	}

	return minX, minY, maxX, maxY
}

// ExpandBoundaryZone expands the boundary tiles by the specified radius.
// This creates a transition zone around region boundaries.
func ExpandBoundaryZone(boundaries []Point, radius, width, height int) []Point {
	zone := make(map[Point]bool)

	// Add original boundaries
	for _, p := range boundaries {
		zone[p] = true
	}

	// Expand by radius
	for r := 0; r < radius; r++ {
		toAdd := make([]Point, 0)

		for point := range zone {
			// Add 4-connected neighbors
			neighbors := []Point{
				{X: point.X - 1, Y: point.Y},
				{X: point.X + 1, Y: point.Y},
				{X: point.X, Y: point.Y - 1},
				{X: point.X, Y: point.Y + 1},
			}

			for _, n := range neighbors {
				if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height {
					if !zone[n] {
						toAdd = append(toAdd, n)
					}
				}
			}
		}

		for _, p := range toAdd {
			zone[p] = true
		}
	}

	// Convert map to slice
	result := make([]Point, 0, len(zone))
	for p := range zone {
		result = append(result, p)
	}

	return result
}
