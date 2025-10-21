package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// CellularGenerator generates cave-like terrain using cellular automata.
// It starts with random noise and applies rules to create organic structures.
type CellularGenerator struct {
	fillProbability float64
	iterations      int
	birthLimit      int
	deathLimit      int
}

// NewCellularGenerator creates a new cellular automata generator.
func NewCellularGenerator() *CellularGenerator {
	return &CellularGenerator{
		fillProbability: 0.40,
		iterations:      5,
		birthLimit:      4,
		deathLimit:      3,
	}
}

// Generate creates cave-like terrain using cellular automata.
func (g *CellularGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Use custom parameters if provided, otherwise use defaults
	width := 80
	height := 50
	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok {
			width = w
		}
		if h, ok := params.Custom["height"].(int); ok {
			height = h
		}
		if f, ok := params.Custom["fillProbability"].(float64); ok {
			g.fillProbability = f
		}
		if i, ok := params.Custom["iterations"].(int); ok {
			g.iterations = i
		}
	}

	// Create random source from seed for deterministic generation
	rng := rand.New(rand.NewSource(seed))

	// Create terrain
	terrain := NewTerrain(width, height, seed)

	// Initialize with random noise
	g.initializeNoise(terrain, rng)

	// Apply cellular automata rules
	for i := 0; i < g.iterations; i++ {
		g.simulateStep(terrain)
	}

	// Post-process to ensure connectivity
	g.ensureConnectivity(terrain)

	return terrain, nil
}

// initializeNoise fills the map with random walls and floors.
func (g *CellularGenerator) initializeNoise(terrain *Terrain, rng *rand.Rand) {
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			// Keep edges as walls
			if x == 0 || x == terrain.Width-1 || y == 0 || y == terrain.Height-1 {
				terrain.SetTile(x, y, TileWall)
			} else if rng.Float64() < g.fillProbability {
				terrain.SetTile(x, y, TileWall)
			} else {
				terrain.SetTile(x, y, TileFloor)
			}
		}
	}
}

// simulateStep performs one iteration of the cellular automata rules.
func (g *CellularGenerator) simulateStep(terrain *Terrain) {
	// Create a copy of the current state
	newTiles := make([][]TileType, terrain.Height)
	for y := range newTiles {
		newTiles[y] = make([]TileType, terrain.Width)
		copy(newTiles[y], terrain.Tiles[y])
	}

	// Apply rules to each cell
	for y := 1; y < terrain.Height-1; y++ {
		for x := 1; x < terrain.Width-1; x++ {
			neighbors := g.countWallNeighbors(terrain, x, y)

			// Apply birth/death rules
			if terrain.GetTile(x, y) == TileWall {
				// Death rule: become floor if too few neighbors
				if neighbors < g.deathLimit {
					newTiles[y][x] = TileFloor
				}
			} else {
				// Birth rule: become wall if enough neighbors
				if neighbors > g.birthLimit {
					newTiles[y][x] = TileWall
				}
			}
		}
	}

	// Update terrain with new state
	terrain.Tiles = newTiles
}

// countWallNeighbors counts the number of wall tiles in the 8 surrounding cells.
func (g *CellularGenerator) countWallNeighbors(terrain *Terrain, x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if terrain.GetTile(x+dx, y+dy) == TileWall {
				count++
			}
		}
	}
	return count
}

// ensureConnectivity uses flood fill to find and connect isolated regions.
func (g *CellularGenerator) ensureConnectivity(terrain *Terrain) {
	// Find all floor regions using flood fill
	regions := g.findRegions(terrain)

	// If there's only one region, we're done
	if len(regions) <= 1 {
		return
	}

	// Connect the largest region to all others
	largestIdx := 0
	largestSize := len(regions[0])
	for i, region := range regions {
		if len(region) > largestSize {
			largestSize = len(region)
			largestIdx = i
		}
	}

	// Connect each smaller region to the largest one
	for i, region := range regions {
		if i != largestIdx && len(region) > 0 {
			g.connectRegions(terrain, regions[largestIdx][0], region[0])
		}
	}
}

// findRegions finds all connected floor regions using flood fill.
func (g *CellularGenerator) findRegions(terrain *Terrain) [][]*Tile {
	visited := make([][]bool, terrain.Height)
	for y := range visited {
		visited[y] = make([]bool, terrain.Width)
	}

	regions := make([][]*Tile, 0)

	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if !visited[y][x] && terrain.IsWalkable(x, y) {
				region := g.floodFill(terrain, x, y, visited)
				if len(region) > 10 { // Ignore very small regions
					regions = append(regions, region)
				}
			}
		}
	}

	return regions
}

// floodFill performs flood fill to find all connected floor tiles.
func (g *CellularGenerator) floodFill(terrain *Terrain, startX, startY int, visited [][]bool) []*Tile {
	region := make([]*Tile, 0)
	stack := []struct{ x, y int }{{startX, startY}}

	for len(stack) > 0 {
		// Pop from stack
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		x, y := pos.x, pos.y

		// Check bounds
		if x < 0 || x >= terrain.Width || y < 0 || y >= terrain.Height {
			continue
		}

		// Skip if already visited or not walkable
		if visited[y][x] || !terrain.IsWalkable(x, y) {
			continue
		}

		// Mark as visited and add to region
		visited[y][x] = true
		region = append(region, &Tile{Type: terrain.GetTile(x, y), X: x, Y: y})

		// Add neighbors to stack
		stack = append(stack,
			struct{ x, y int }{x + 1, y},
			struct{ x, y int }{x - 1, y},
			struct{ x, y int }{x, y + 1},
			struct{ x, y int }{x, y - 1},
		)
	}

	return region
}

// connectRegions creates a corridor between two regions.
func (g *CellularGenerator) connectRegions(terrain *Terrain, tile1, tile2 *Tile) {
	x1, y1 := tile1.X, tile1.Y
	x2, y2 := tile2.X, tile2.Y

	// Create L-shaped corridor
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		terrain.SetTile(x, y1, TileCorridor)
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		terrain.SetTile(x2, y, TileCorridor)
	}
}

// Validate checks if the generated terrain is valid.
func (g *CellularGenerator) Validate(result interface{}) error {
	terrain, ok := result.(*Terrain)
	if !ok {
		return fmt.Errorf("result is not a Terrain")
	}

	// Count floor tiles
	floorCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				floorCount++
			}
		}
	}

	// Check that we have a reasonable amount of open space (at least 30%)
	totalTiles := terrain.Width * terrain.Height
	if floorCount < totalTiles*3/10 {
		return fmt.Errorf("too few walkable tiles: %d/%d", floorCount, totalTiles)
	}

	return nil
}
