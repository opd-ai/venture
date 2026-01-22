// Package terrain provides forest generation with natural features.
// This file implements a forest generation algorithm that creates
// natural-looking environments with trees, clearings, and water features.
package terrain

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// ForestGenerator generates natural forest environments with trees, clearings, and water.
type ForestGenerator struct {
	treeDensity   float64 // Percentage of tiles that should be trees (0.0-1.0)
	clearingCount int     // Number of open clearings to create
	waterChance   float64 // Probability of water features (0.0-1.0)
	logger        *logrus.Entry
}

// NewForestGenerator creates a new forest generator with default parameters.
func NewForestGenerator() *ForestGenerator {
	return NewForestGeneratorWithLogger(nil)
}

// NewForestGeneratorWithLogger creates a new forest generator with a logger.
func NewForestGeneratorWithLogger(logger *logrus.Logger) *ForestGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "forest")
	}
	return &ForestGenerator{
		treeDensity:   0.3, // 30% of tiles are trees
		clearingCount: 3,   // 3-5 clearings
		waterChance:   0.3, // 30% chance of water features
		logger:        logEntry,
	}
}

// Generate creates a forest environment using Poisson disc sampling for trees.
func (g *ForestGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	g.logGenerationStart(seed, params)

	width, height := g.extractDimensions(params)
	if err := g.validateDimensions(width, height); err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(seed))
	terrain := g.initializeTerrain(width, height, seed)
	clearings := g.createClearings(terrain, rng)

	g.populateForestFeatures(terrain, clearings, rng)
	g.logGenerationComplete(terrain, clearings)

	return terrain, nil
}

// extractDimensions extracts and applies custom dimension parameters.
func (g *ForestGenerator) extractDimensions(params procgen.GenerationParams) (width, height int) {
	width, height = 80, 50
	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok {
			width = w
		}
		if h, ok := params.Custom["height"].(int); ok {
			height = h
		}
		if td, ok := params.Custom["treeDensity"].(float64); ok {
			g.treeDensity = td
		}
		if cc, ok := params.Custom["clearingCount"].(int); ok {
			g.clearingCount = cc
		}
		if wc, ok := params.Custom["waterChance"].(float64); ok {
			g.waterChance = wc
		}
	}
	return width, height
}

// validateDimensions validates terrain dimensions.
func (g *ForestGenerator) validateDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}
	if width > 1000 || height > 1000 {
		return fmt.Errorf("dimensions too large: width=%d, height=%d (max 1000x1000)", width, height)
	}
	return nil
}

// initializeTerrain creates a new terrain filled with floor tiles.
func (g *ForestGenerator) initializeTerrain(width, height int, seed int64) *Terrain {
	terrain := NewTerrain(width, height, seed)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}
	return terrain
}

// populateForestFeatures adds water, trees, paths and stairs to the forest.
func (g *ForestGenerator) populateForestFeatures(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	if rng.Float64() < g.waterChance {
		g.addWaterFeatures(terrain, clearings, rng)
	}
	g.generateTrees(terrain, clearings, rng)
	g.connectClearings(terrain, clearings, rng)
	g.placeAutoBridges(terrain)
	g.placeStairsInClearings(terrain, clearings, rng)
}

// logGenerationStart logs the start of forest generation.
func (g *ForestGenerator) logGenerationStart(seed int64, params procgen.GenerationParams) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting forest terrain generation")
	}
}

// logGenerationComplete logs the completion of forest generation.
func (g *ForestGenerator) logGenerationComplete(terrain *Terrain, clearings []*Room) {
	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":     terrain.Width,
			"height":    terrain.Height,
			"clearings": len(clearings),
		}).Info("forest terrain generation complete")
	}
}

// createClearings creates circular or elliptical open areas in the forest.
// calculateClearingDimensions generates random dimensions for a clearing
func calculateClearingDimensions(rng *rand.Rand) (width, height int) {
	width = 8 + rng.Intn(10)
	height = 8 + rng.Intn(10)
	return width, height
}

// validateTerrainSize checks if terrain is large enough for a clearing of given dimensions
func validateTerrainSize(terrain *Terrain, width, height int) (maxX, maxY int, valid bool) {
	maxX = terrain.Width - width - 4
	maxY = terrain.Height - height - 4
	if maxX <= 0 || maxY <= 0 {
		return 0, 0, false
	}
	return maxX, maxY, true
}

// checkClearingOverlap determines if a clearing overlaps with any existing clearings
func checkClearingOverlap(clearing *Room, existingClearings []*Room) bool {
	for _, existing := range existingClearings {
		if clearing.Overlaps(existing) {
			return true
		}
	}
	return false
}

// createEllipticalClearing fills a clearing area with floor tiles in an elliptical shape
func createEllipticalClearing(terrain *Terrain, clearing *Room) {
	cx, cy := clearing.Center()
	radiusX := float64(clearing.Width) / 2.0
	radiusY := float64(clearing.Height) / 2.0

	for dy := 0; dy < clearing.Height; dy++ {
		for dx := 0; dx < clearing.Width; dx++ {
			px := clearing.X + dx
			py := clearing.Y + dy

			normX := (float64(px) - float64(cx)) / radiusX
			normY := (float64(py) - float64(cy)) / radiusY
			if normX*normX+normY*normY <= 1.0 {
				terrain.SetTile(px, py, TileFloor)
			}
		}
	}
}

func (g *ForestGenerator) createClearings(terrain *Terrain, rng *rand.Rand) []*Room {
	clearings := make([]*Room, 0)
	attempts := g.clearingCount * 5

	for i := 0; i < attempts && len(clearings) < g.clearingCount; i++ {
		width, height := calculateClearingDimensions(rng)

		maxX, maxY, valid := validateTerrainSize(terrain, width, height)
		if !valid {
			continue
		}

		x := 2 + rng.Intn(maxX)
		y := 2 + rng.Intn(maxY)

		clearing := &Room{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
			Type:   RoomNormal,
		}

		if checkClearingOverlap(clearing, clearings) {
			continue
		}

		createEllipticalClearing(terrain, clearing)
		clearings = append(clearings, clearing)
	}

	terrain.Rooms = clearings
	return clearings
}

// generateTrees places trees using Poisson disc sampling for natural distribution.
func (g *ForestGenerator) generateTrees(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	// Calculate minimum distance between trees based on density
	// Higher density = smaller minimum distance between trees
	// Formula: minDist ≈ 1/sqrt(density) gives points with spacing inversely proportional to density
	minDist := 1.0 / math.Sqrt(g.treeDensity)
	if minDist < 1.5 {
		minDist = 1.5
	}
	if minDist > 5.0 {
		minDist = 5.0
	}

	// Get tree positions using Poisson disc sampling
	treePositions := g.poissonDiscSampling(terrain.Width, terrain.Height, minDist, rng)

	// Place trees, avoiding clearings
	for _, pos := range treePositions {
		// Check if position is in a clearing
		inClearing := false
		for _, clearing := range clearings {
			if pos.X >= clearing.X && pos.X < clearing.X+clearing.Width &&
				pos.Y >= clearing.Y && pos.Y < clearing.Y+clearing.Height {
				inClearing = true
				break
			}
		}

		// Place tree if not in clearing
		if !inClearing && terrain.GetTile(pos.X, pos.Y) == TileFloor {
			terrain.SetTile(pos.X, pos.Y, TileTree)
		}
	}
}

// poissonDiscSampling generates evenly distributed points using Poisson disc sampling.
func (g *ForestGenerator) poissonDiscSampling(width, height int, minDist float64, rng *rand.Rand) []Point {
	cellSize := minDist / math.Sqrt(2.0)
	gridW := int(math.Ceil(float64(width) / cellSize))
	gridH := int(math.Ceil(float64(height) / cellSize))

	grid := g.initializePoissonGrid(gridW, gridH)
	points, activeList := g.initializeWithStartPoint(width, height, grid, gridW, gridH, cellSize, rng)
	g.processActivePoints(&points, &activeList, grid, cellSize, minDist, width, height, rng)

	return points
}

// initializePoissonGrid creates an empty grid for Poisson disc sampling.
func (g *ForestGenerator) initializePoissonGrid(gridW, gridH int) [][]int {
	grid := make([][]int, gridH)
	for i := range grid {
		grid[i] = make([]int, gridW)
		for j := range grid[i] {
			grid[i][j] = -1
		}
	}
	return grid
}

// initializeWithStartPoint creates the initial point and active list.
func (g *ForestGenerator) initializeWithStartPoint(width, height int, grid [][]int, gridW, gridH int, cellSize float64, rng *rand.Rand) ([]Point, []int) {
	startX := rng.Intn(width)
	startY := rng.Intn(height)
	startPoint := Point{X: startX, Y: startY}

	points := []Point{startPoint}
	activeList := []int{0}

	startGridX := int(float64(startX) / cellSize)
	startGridY := int(float64(startY) / cellSize)
	if startGridX >= 0 && startGridX < gridW && startGridY >= 0 && startGridY < gridH {
		grid[startGridY][startGridX] = 0
	}

	return points, activeList
}

// processActivePoints generates new points around active points.
func (g *ForestGenerator) processActivePoints(points *[]Point, activeList *[]int, grid [][]int, cellSize, minDist float64, width, height int, rng *rand.Rand) {
	for len(*activeList) > 0 {
		activeIdx := rng.Intn(len(*activeList))
		pointIdx := (*activeList)[activeIdx]
		point := (*points)[pointIdx]

		found := g.tryGenerateNewPoints(point, points, activeList, grid, cellSize, minDist, width, height, rng)
		if !found {
			*activeList = append((*activeList)[:activeIdx], (*activeList)[activeIdx+1:]...)
		}
	}
}

// tryGenerateNewPoints attempts to generate new points around a point.
func (g *ForestGenerator) tryGenerateNewPoints(point Point, points *[]Point, activeList *[]int, grid [][]int, cellSize, minDist float64, width, height int, rng *rand.Rand) bool {
	for i := 0; i < 30; i++ {
		newPoint := g.generateCandidatePoint(point, minDist, rng)
		if g.isValidAndAddPoint(newPoint, points, activeList, grid, cellSize, minDist, width, height) {
			return true
		}
	}
	return false
}

// generateCandidatePoint generates a random point in the annulus around a point.
// Uses rejection sampling with cartesian offsets to avoid expensive trigonometric calls.
func (g *ForestGenerator) generateCandidatePoint(point Point, minDist float64, rng *rand.Rand) Point {
	// Use rejection sampling instead of polar coordinates to avoid cos/sin
	// Generate random offset in [-2*minDist, 2*minDist] range
	maxRange := minDist * 2.0
	for {
		dx := (rng.Float64()*2.0 - 1.0) * maxRange
		dy := (rng.Float64()*2.0 - 1.0) * maxRange
		distSq := dx*dx + dy*dy
		minDistSq := minDist * minDist
		maxDistSq := maxRange * maxRange

		// Check if point is in the annulus [minDist, 2*minDist]
		if distSq >= minDistSq && distSq <= maxDistSq {
			return Point{
				X: point.X + int(dx),
				Y: point.Y + int(dy),
			}
		}
	}
}

// isValidAndAddPoint validates and adds a point if valid.
func (g *ForestGenerator) isValidAndAddPoint(newPoint Point, points *[]Point, activeList *[]int, grid [][]int, cellSize, minDist float64, width, height int) bool {
	if newPoint.X < 0 || newPoint.X >= width || newPoint.Y < 0 || newPoint.Y >= height {
		return false
	}

	gridW := len(grid[0])
	gridH := len(grid)
	gridX := int(float64(newPoint.X) / cellSize)
	gridY := int(float64(newPoint.Y) / cellSize)

	if gridX < 0 || gridX >= gridW || gridY < 0 || gridY >= gridH {
		return false
	}

	if !g.isValidPoissonPoint(newPoint, *points, grid, cellSize, minDist, width, height) {
		return false
	}

	*points = append(*points, newPoint)
	*activeList = append(*activeList, len(*points)-1)
	grid[gridY][gridX] = len(*points) - 1
	return true
}

// isValidPoissonPoint checks if a point is valid for Poisson disc sampling.
// Uses squared distance comparison to avoid expensive sqrt operations.
func (g *ForestGenerator) isValidPoissonPoint(point Point, points []Point, grid [][]int,
	cellSize, minDist float64, width, height int,
) bool {
	// Get grid cell
	gridX := int(float64(point.X) / cellSize)
	gridY := int(float64(point.Y) / cellSize)

	// Pre-compute squared minimum distance for comparison
	minDistSq := minDist * minDist

	// Check neighboring cells
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			checkY := gridY + dy
			checkX := gridX + dx

			if checkY >= 0 && checkY < len(grid) && checkX >= 0 && checkX < len(grid[0]) {
				if grid[checkY][checkX] != -1 {
					neighbor := points[grid[checkY][checkX]]
					// Use squared distance to avoid sqrt
					distSq := point.DistanceSquared(neighbor)
					if distSq < minDistSq {
						return false
					}
				}
			}
		}
	}

	return true
}

// addWaterFeatures adds lakes or rivers to the forest.
func (g *ForestGenerator) addWaterFeatures(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	// Choose feature type
	featureType := rng.Intn(2) // 0 = lake, 1 = river

	if featureType == 0 {
		// Create 1-2 lakes
		lakeCount := 1 + rng.Intn(2)
		for i := 0; i < lakeCount; i++ {
			g.createLake(terrain, clearings, rng)
		}
	} else {
		// Create a river
		g.createRiver(terrain, rng)
	}
}

// createLake creates a natural-looking lake.
func (g *ForestGenerator) createLake(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	centerX, centerY, valid := g.findValidLakePosition(terrain, clearings, rng)
	if !valid {
		return
	}

	radiusX := 4.0 + rng.Float64()*4.0
	radiusY := 4.0 + rng.Float64()*4.0

	g.fillLakeArea(terrain, centerX, centerY, radiusX, radiusY, rng)
}

// findValidLakePosition finds a valid position for a lake away from clearings.
func (g *ForestGenerator) findValidLakePosition(terrain *Terrain, clearings []*Room, rng *rand.Rand) (int, int, bool) {
	if terrain.Width <= 20 || terrain.Height <= 20 {
		return 0, 0, false
	}

	maxAttempts := 50
	for attempt := 0; attempt < maxAttempts; attempt++ {
		centerX := 10 + rng.Intn(terrain.Width-20)
		centerY := 10 + rng.Intn(terrain.Height-20)

		if g.isFarFromClearings(centerX, centerY, clearings) {
			return centerX, centerY, true
		}
	}
	return 0, 0, false
}

// isFarFromClearings checks if a position is far enough from all clearings.
func (g *ForestGenerator) isFarFromClearings(x, y int, clearings []*Room) bool {
	minDist := 15.0
	for _, clearing := range clearings {
		cx, cy := clearing.Center()
		dist := math.Sqrt(float64((x-cx)*(x-cx) + (y-cy)*(y-cy)))
		if dist < minDist {
			return false
		}
	}
	return true
}

// fillLakeArea fills the lake area with water tiles using irregular ellipse.
func (g *ForestGenerator) fillLakeArea(terrain *Terrain, centerX, centerY int, radiusX, radiusY float64, rng *rand.Rand) {
	for dy := -int(radiusY) - 2; dy <= int(radiusY)+2; dy++ {
		for dx := -int(radiusX) - 2; dx <= int(radiusX)+2; dx++ {
			x := centerX + dx
			y := centerY + dy

			if !terrain.IsInBounds(x, y) {
				continue
			}

			waterType := g.determineWaterType(dx, dy, radiusX, radiusY, rng)
			if waterType != TileFloor {
				terrain.SetTile(x, y, waterType)
			}
		}
	}
}

// determineWaterType determines water tile type based on distance from center.
func (g *ForestGenerator) determineWaterType(dx, dy int, radiusX, radiusY float64, rng *rand.Rand) TileType {
	normX := float64(dx) / radiusX
	normY := float64(dy) / radiusY
	dist := math.Sqrt(normX*normX + normY*normY)
	noise := rng.Float64()*0.3 - 0.15

	if dist+noise < 0.7 {
		return TileWaterDeep
	} else if dist+noise < 1.0 {
		return TileWaterShallow
	}
	return TileFloor
}

// createRiver creates a winding river across the map.
func (g *ForestGenerator) createRiver(terrain *Terrain, rng *rand.Rand) {
	// River flows from one edge to opposite edge
	startEdge := rng.Intn(4) // 0=top, 1=right, 2=bottom, 3=left

	var x, y int
	var dx, dy float64

	switch startEdge {
	case 0: // Top to bottom
		x = rng.Intn(terrain.Width)
		y = 0
		dx = (rng.Float64() - 0.5) * 0.3
		dy = 1.0
	case 1: // Right to left
		x = terrain.Width - 1
		y = rng.Intn(terrain.Height)
		dx = -1.0
		dy = (rng.Float64() - 0.5) * 0.3
	case 2: // Bottom to top
		x = rng.Intn(terrain.Width)
		y = terrain.Height - 1
		dx = (rng.Float64() - 0.5) * 0.3
		dy = -1.0
	case 3: // Left to right
		x = 0
		y = rng.Intn(terrain.Height)
		dx = 1.0
		dy = (rng.Float64() - 0.5) * 0.3
	}

	// Trace river path
	riverWidth := 2 + rng.Intn(2) // 2-3 tiles wide
	step := 0

	for terrain.IsInBounds(x, y) && step < terrain.Width+terrain.Height {
		// Place water at current position
		for wy := -riverWidth / 2; wy <= riverWidth/2; wy++ {
			for wx := -riverWidth / 2; wx <= riverWidth/2; wx++ {
				rx := x + wx
				ry := y + wy
				if terrain.IsInBounds(rx, ry) {
					if wx == 0 && wy == 0 {
						terrain.SetTile(rx, ry, TileWaterDeep)
					} else {
						terrain.SetTile(rx, ry, TileWaterShallow)
					}
				}
			}
		}

		// Add some randomness to direction
		dx += (rng.Float64() - 0.5) * 0.2
		dy += (rng.Float64() - 0.5) * 0.2

		// Normalize direction
		length := math.Sqrt(dx*dx + dy*dy)
		if length > 0 {
			dx /= length
			dy /= length
		}

		// Move to next position
		x += int(dx * 2.0)
		y += int(dy * 2.0)
		step++
	}
}

// connectClearings creates organic paths between clearings.
func (g *ForestGenerator) connectClearings(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	if len(clearings) < 2 {
		return
	}

	// Connect each clearing to at least one other
	for i := 0; i < len(clearings); i++ {
		// Find nearest unconnected clearing
		nearest := (i + 1) % len(clearings)
		from := clearings[i]
		to := clearings[nearest]

		fromX, fromY := from.Center()
		toX, toY := to.Center()

		// Create winding path
		g.createOrganicPath(Point{X: fromX, Y: fromY}, Point{X: toX, Y: toY}, terrain, rng)
	}
}

// createOrganicPath creates a natural-looking path between two points.
func (g *ForestGenerator) createOrganicPath(start, end Point, terrain *Terrain, rng *rand.Rand) {
	current := start

	for current.ManhattanDistance(end) > 2 {
		// Move toward target with some randomness
		dx := 0
		dy := 0

		if current.X < end.X {
			dx = 1
		} else if current.X > end.X {
			dx = -1
		}

		if current.Y < end.Y {
			dy = 1
		} else if current.Y > end.Y {
			dy = -1
		}

		// Add randomness
		if rng.Float64() < 0.3 {
			dx += rng.Intn(3) - 1
			dy += rng.Intn(3) - 1
		}

		current.X += dx
		current.Y += dy

		// Ensure in bounds
		if !current.IsInBounds(terrain.Width, terrain.Height) {
			break
		}

		// Clear path (remove trees, keep water)
		tile := terrain.GetTile(current.X, current.Y)
		if tile == TileTree {
			terrain.SetTile(current.X, current.Y, TileFloor)
		}
		// Note: Don't overwrite water tiles - bridges will be placed later
	}
}

// placeAutoBridges automatically places bridges where paths cross water.
func (g *ForestGenerator) placeAutoBridges(terrain *Terrain) {
	for y := 1; y < terrain.Height-1; y++ {
		for x := 1; x < terrain.Width-1; x++ {
			tile := terrain.GetTile(x, y)

			// Check if this is water
			if tile == TileWaterShallow || tile == TileWaterDeep {
				// Check if floor tiles are on opposite sides (horizontal or vertical)
				hasPathH := (terrain.GetTile(x-1, y) == TileFloor || terrain.GetTile(x-1, y) == TileBridge) &&
					(terrain.GetTile(x+1, y) == TileFloor || terrain.GetTile(x+1, y) == TileBridge)
				hasPathV := (terrain.GetTile(x, y-1) == TileFloor || terrain.GetTile(x, y-1) == TileBridge) &&
					(terrain.GetTile(x, y+1) == TileFloor || terrain.GetTile(x, y+1) == TileBridge)

				if hasPathH || hasPathV {
					terrain.SetTile(x, y, TileBridge)
				}
			}
		}
	}
}

// placeStairsInClearings places stairs in the largest clearings.
func (g *ForestGenerator) placeStairsInClearings(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	if len(clearings) == 0 {
		return
	}

	// Find two largest clearings
	var largest, secondLargest *Room
	for _, clearing := range clearings {
		size := clearing.Width * clearing.Height
		if largest == nil || size > largest.Width*largest.Height {
			secondLargest = largest
			largest = clearing
		} else if secondLargest == nil || size > secondLargest.Width*secondLargest.Height {
			secondLargest = clearing
		}
	}

	// Place stairs up in largest clearing
	if largest != nil {
		cx, cy := largest.Center()
		terrain.AddStairs(cx, cy, true)
	}

	// Place stairs down in second largest (or any other clearing)
	if secondLargest != nil {
		cx, cy := secondLargest.Center()
		terrain.AddStairs(cx, cy, false)
	} else if len(clearings) > 1 {
		// Use any other clearing
		cx, cy := clearings[1].Center()
		terrain.AddStairs(cx, cy, false)
	}
}

// Validate checks if the generated forest meets quality requirements.
func (g *ForestGenerator) Validate(result interface{}) error {
	terrain, ok := result.(*Terrain)
	if !ok {
		return fmt.Errorf("result is not a Terrain")
	}

	// Count walkable tiles
	walkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkable++
			}
		}
	}

	// Ensure at least 40% of tiles are walkable (forests should have decent open space)
	totalTiles := terrain.Width * terrain.Height
	if float64(walkable)/float64(totalTiles) < 0.4 {
		return fmt.Errorf("insufficient walkable tiles: %d/%d (%.1f%%, need >= 40%%)",
			walkable, totalTiles, float64(walkable)/float64(totalTiles)*100)
	}

	// Ensure at least one clearing was created
	if len(terrain.Rooms) == 0 {
		return fmt.Errorf("no clearings created")
	}

	// Validate stair placement if stairs exist
	if len(terrain.StairsUp) > 0 || len(terrain.StairsDown) > 0 {
		if err := terrain.ValidateStairPlacement(); err != nil {
			return fmt.Errorf("stair validation failed: %w", err)
		}
	}

	return nil
}
