// Package terrain provides cellular automata cave generation.
// This file implements cellular automata algorithm for organic
// cave and cavern generation.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// neighborOffsets contains pre-computed neighbor coordinate offsets for
// the 8 surrounding cells. Using pre-computed offsets avoids nested loop overhead.
var neighborOffsets = [8][2]int{
	{-1, -1},
	{0, -1},
	{1, -1},
	{-1, 0},
	{1, 0},
	{-1, 1},
	{0, 1},
	{1, 1},
}

// CellularGenerator generates cave-like terrain using cellular automata.
// It starts with random noise and applies rules to create organic structures.
// Pre-allocated buffers reduce allocations during simulation iterations.
type CellularGenerator struct {
	fillProbability float64
	iterations      int
	birthLimit      int
	deathLimit      int
	logger          *logrus.Entry
	// Pre-allocated buffers for simulation (double-buffering pattern)
	bufferA [][]TileType
	bufferB [][]TileType
	// Track which buffer is currently active (true = bufferA, false = bufferB)
	useBufferA bool
	// Pre-allocated visited tracking for flood fill
	visitedBuffer [][]bool
}

// NewCellularGenerator creates a new cellular automata generator.
func NewCellularGenerator() *CellularGenerator {
	return NewCellularGeneratorWithLogger(nil)
}

// NewCellularGeneratorWithLogger creates a new cellular automata generator with a logger.
func NewCellularGeneratorWithLogger(logger *logrus.Logger) *CellularGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "cellular")
	}
	return &CellularGenerator{
		fillProbability: 0.40,
		iterations:      5,
		birthLimit:      4,
		deathLimit:      3,
		logger:          logEntry,
	}
}

// Generate creates cave-like terrain using cellular automata.
func (g *CellularGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting cellular automata terrain generation")
	}

	width, height := g.parseTerrainDimensions(params)
	if err := g.validateTerrainDimensions(width, height); err != nil {
		return nil, err
	}

	terrain, addedLakes := g.generateCellularTerrain(seed, width, height)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":      terrain.Width,
			"height":     terrain.Height,
			"iterations": g.iterations,
			"lakes":      addedLakes,
		}).Info("cellular automata terrain generation complete")
	}

	return terrain, nil
}

// parseTerrainDimensions extracts width and height from custom parameters.
func (g *CellularGenerator) parseTerrainDimensions(params procgen.GenerationParams) (width, height int) {
	width = 80
	height = 50

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

	return width, height
}

// ensureBuffers allocates or resizes buffers if needed for given dimensions.
// This minimizes allocations by reusing buffers across Generate calls.
func (g *CellularGenerator) ensureBuffers(width, height int) {
	// Check if buffers need allocation or resizing
	needsResize := len(g.bufferA) != height
	if !needsResize && len(g.bufferA) > 0 {
		needsResize = len(g.bufferA[0]) != width
	}

	if needsResize {
		// Allocate double buffers for simulation iterations
		g.bufferA = make([][]TileType, height)
		g.bufferB = make([][]TileType, height)
		for y := 0; y < height; y++ {
			g.bufferA[y] = make([]TileType, width)
			g.bufferB[y] = make([]TileType, width)
		}

		// Allocate visited buffer for flood fill
		g.visitedBuffer = make([][]bool, height)
		for y := 0; y < height; y++ {
			g.visitedBuffer[y] = make([]bool, width)
		}
	}
}

// validateTerrainDimensions validates width and height are within acceptable bounds.
func (g *CellularGenerator) validateTerrainDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width and height must be positive (got width=%d, height=%d)", width, height)
	}

	const maxDimension = 10000
	if width > maxDimension || height > maxDimension {
		return fmt.Errorf("dimensions too large: maximum is %d (got width=%d, height=%d)", maxDimension, width, height)
	}

	return nil
}

// generateCellularTerrain creates terrain using cellular automata.
func (g *CellularGenerator) generateCellularTerrain(seed int64, width, height int) (*Terrain, bool) {
	rng := rand.New(rand.NewSource(seed))

	// Ensure buffers are allocated for this terrain size
	g.ensureBuffers(width, height)

	// Create terrain with one of our pre-allocated buffers
	// This avoids the initial terrain allocation and final copy
	terrain := &Terrain{
		Width:  width,
		Height: height,
		Seed:   seed,
		Tiles:  g.bufferA, // Start with bufferA
	}

	g.initializeNoise(terrain, rng)

	// Start with bufferA (already assigned above), iterations will swap
	g.useBufferA = true
	for i := 0; i < g.iterations; i++ {
		g.simulateStep(terrain)
	}

	// Copy final result to terrain's own buffer to avoid aliasing
	// This ensures the terrain owns its data and isn't affected by future generations
	finalTiles := make([][]TileType, height)
	for y := 0; y < height; y++ {
		finalTiles[y] = make([]TileType, width)
		copy(finalTiles[y], terrain.Tiles[y])
	}
	terrain.Tiles = finalTiles

	g.ensureConnectivity(terrain)

	addedLakes := false
	if rng.Float64() < 0.3 {
		g.addUndergroundLakes(terrain, rng)
		addedLakes = true
	}

	return terrain, addedLakes
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
// Uses pre-allocated buffers to eliminate per-iteration allocations.
func (g *CellularGenerator) simulateStep(terrain *Terrain) {
	width := terrain.Width
	height := terrain.Height
	tiles := terrain.Tiles

	// Determine which buffer to use as output (swap between A and B)
	// This eliminates the need to allocate newTiles on every iteration
	var outputBuffer [][]TileType
	if g.useBufferA {
		outputBuffer = g.bufferA
		g.useBufferA = false
	} else {
		outputBuffer = g.bufferB
		g.useBufferA = true
	}

	// Copy current state to output buffer (edges stay as walls)
	for y := 0; y < height; y++ {
		copy(outputBuffer[y], tiles[y])
	}

	// Apply rules to each interior cell (edges stay as walls)
	for y := 1; y < height-1; y++ {
		row := tiles[y]
		rowAbove := tiles[y-1]
		rowBelow := tiles[y+1]
		outputRow := outputBuffer[y]

		for x := 1; x < width-1; x++ {
			// Count wall neighbors using direct array access (no bounds checks needed)
			neighbors := g.countWallNeighborsFast(row, rowAbove, rowBelow, x)

			// Apply birth/death rules
			if row[x] == TileWall {
				// Death rule: become floor if too few neighbors
				if neighbors < g.deathLimit {
					outputRow[x] = TileFloor
				}
			} else {
				// Birth rule: become wall if enough neighbors
				if neighbors > g.birthLimit {
					outputRow[x] = TileWall
				}
			}
		}
	}

	// Swap buffer reference in terrain
	terrain.Tiles = outputBuffer
}

// countWallNeighbors counts the number of wall tiles in the 8 surrounding cells.
// Uses pre-computed neighbor offsets to avoid nested loop overhead.
func (g *CellularGenerator) countWallNeighbors(terrain *Terrain, x, y int) int {
	count := 0
	tiles := terrain.Tiles
	for _, offset := range neighborOffsets {
		nx, ny := x+offset[0], y+offset[1]
		if nx >= 0 && nx < terrain.Width && ny >= 0 && ny < terrain.Height {
			if tiles[ny][nx] == TileWall {
				count++
			}
		} else {
			// Out of bounds counts as wall
			count++
		}
	}
	return count
}

// countWallNeighborsFast counts wall neighbors using direct row references.
// This avoids bounds checking for interior cells where neighbors are guaranteed to exist.
// row is the current row, rowAbove and rowBelow are the adjacent rows.
func (g *CellularGenerator) countWallNeighborsFast(row, rowAbove, rowBelow []TileType, x int) int {
	count := 0
	// Check row above: x-1, x, x+1
	if rowAbove[x-1] == TileWall {
		count++
	}
	if rowAbove[x] == TileWall {
		count++
	}
	if rowAbove[x+1] == TileWall {
		count++
	}
	// Check current row: x-1, x+1 (skip center)
	if row[x-1] == TileWall {
		count++
	}
	if row[x+1] == TileWall {
		count++
	}
	// Check row below: x-1, x, x+1
	if rowBelow[x-1] == TileWall {
		count++
	}
	if rowBelow[x] == TileWall {
		count++
	}
	if rowBelow[x+1] == TileWall {
		count++
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
// Uses pre-allocated visitedBuffer to reduce allocations.
func (g *CellularGenerator) findRegions(terrain *Terrain) [][]*Tile {
	// Reset visited buffer (zero all values)
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			g.visitedBuffer[y][x] = false
		}
	}

	regions := make([][]*Tile, 0, 8) // pre-allocate for typical region count

	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if !g.visitedBuffer[y][x] && terrain.IsWalkable(x, y) {
				region := g.floodFill(terrain, x, y, g.visitedBuffer)
				if len(region) > 10 { // Ignore very small regions
					regions = append(regions, region)
				}
			}
		}
	}

	return regions
}

// floodFill performs flood fill to find all connected floor tiles.
// Pre-allocates region and stack buffers to reduce slice growth allocations.
func (g *CellularGenerator) floodFill(terrain *Terrain, startX, startY int, visited [][]bool) []*Tile {
	// Estimate capacity: one-quarter of total tiles is a reasonable typical region size.
	estimatedRegionSize := terrain.Width * terrain.Height / 4
	if estimatedRegionSize < 32 {
		estimatedRegionSize = 32
	}
	region := make([]*Tile, 0, estimatedRegionSize)
	stack := make([]struct{ x, y int }, 0, estimatedRegionSize)
	stack = append(stack, struct{ x, y int }{startX, startY})

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
// Corridors are 3 tiles wide to accommodate 64×64 player sprites.
func (g *CellularGenerator) connectRegions(terrain *Terrain, tile1, tile2 *Tile) {
	x1, y1 := tile1.X, tile1.Y
	x2, y2 := tile2.X, tile2.Y

	// Create L-shaped corridor (3 tiles wide)
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		for dy := -corridorHalfWidth; dy <= corridorHalfWidth; dy++ {
			if terrain.IsInBounds(x, y1+dy) {
				terrain.SetTile(x, y1+dy, TileCorridor)
			}
		}
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		for dx := -corridorHalfWidth; dx <= corridorHalfWidth; dx++ {
			if terrain.IsInBounds(x2+dx, y) {
				terrain.SetTile(x2+dx, y, TileCorridor)
			}
		}
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
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"floorCount":  floorCount,
				"totalTiles":  totalTiles,
				"percentage":  float64(floorCount) / float64(totalTiles) * 100,
				"minRequired": 30.0,
			}).Warn("terrain validation failed: insufficient walkable tiles")
		}
		return fmt.Errorf("too few walkable tiles: %d/%d", floorCount, totalTiles)
	}

	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"floorCount": floorCount,
			"totalTiles": totalTiles,
			"percentage": float64(floorCount) / float64(totalTiles) * 100,
		}).Debug("terrain validation passed")
	}

	return nil
}

// addUndergroundLakes creates small lakes in cave chambers.
// Lakes are placed in open floor areas, giving caves a natural water feature.
func (g *CellularGenerator) addUndergroundLakes(terrain *Terrain, rng *rand.Rand) {
	candidates := g.findLakeCandidates(terrain)
	if len(candidates) == 0 {
		return
	}

	lakeLocations := g.selectLakeLocations(candidates, rng)
	g.placeLakes(terrain, lakeLocations, rng)
}

// findLakeCandidates identifies suitable locations for underground lakes.
func (g *CellularGenerator) findLakeCandidates(terrain *Terrain) []Point {
	candidates := make([]Point, 0, 50)

	for y := 5; y < terrain.Height-5; y++ {
		for x := 5; x < terrain.Width-5; x++ {
			if terrain.GetTile(x, y) == TileFloor && g.hasOpenArea(terrain, x, y) {
				candidates = append(candidates, Point{x, y})
			}
		}
	}

	return candidates
}

// hasOpenArea checks if location is surrounded by mostly floor tiles.
func (g *CellularGenerator) hasOpenArea(terrain *Terrain, x, y int) bool {
	floorNeighbors := 0
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			nx, ny := x+dx, y+dy
			if terrain.IsInBounds(nx, ny) && terrain.GetTile(nx, ny) == TileFloor {
				floorNeighbors++
			}
		}
	}
	return floorNeighbors > 15
}

// selectLakeLocations picks 1-3 non-overlapping lake locations.
func (g *CellularGenerator) selectLakeLocations(candidates []Point, rng *rand.Rand) []Point {
	locations := []Point{}
	numLakes := 1 + rng.Intn(3)

	for i := 0; i < numLakes && len(candidates) > 0; i++ {
		idx := rng.Intn(len(candidates))
		center := candidates[idx]
		locations = append(locations, center)

		candidates = g.removeNearbyPoints(candidates, center)
	}

	return locations
}

// removeNearbyPoints filters out candidates near a given point.
func (g *CellularGenerator) removeNearbyPoints(candidates []Point, center Point) []Point {
	newCandidates := make([]Point, 0, len(candidates))
	for _, c := range candidates {
		dist := abs(c.X-center.X) + abs(c.Y-center.Y)
		if dist > 15 {
			newCandidates = append(newCandidates, c)
		}
	}
	return newCandidates
}

// placeLakes generates lakes at selected locations.
func (g *CellularGenerator) placeLakes(terrain *Terrain, locations []Point, rng *rand.Rand) {
	for _, center := range locations {
		radius := 3 + rng.Intn(4)
		GenerateLake(center, radius, terrain, rng)
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
