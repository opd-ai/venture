// Package terrain provides forest generation with natural features.
// This file implements a forest generation algorithm that creates
// natural-looking environments with trees, clearings, and water features.
package terrain

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// ForestGenerator generates natural forest environments with trees, clearings, and water.
type ForestGenerator struct {
	treeDensity   float64 // Percentage of tiles that should be trees (0.0-1.0)
	clearingCount int     // Number of open clearings to create
	waterChance   float64 // Probability of water features (0.0-1.0)
}

// NewForestGenerator creates a new forest generator with default parameters.
func NewForestGenerator() *ForestGenerator {
	return &ForestGenerator{
		treeDensity:   0.3, // 30% of tiles are trees
		clearingCount: 3,   // 3-5 clearings
		waterChance:   0.3, // 30% chance of water features
	}
}

// Generate creates a forest environment using Poisson disc sampling for trees.
func (g *ForestGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Use custom parameters if provided
	width := 80
	height := 50
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

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}

	if width > 1000 || height > 1000 {
		return nil, fmt.Errorf("dimensions too large: width=%d, height=%d (max 1000x1000)", width, height)
	}

	// Create RNG with seed
	rng := rand.New(rand.NewSource(seed))

	// Create terrain (starts with all floors - grassland)
	terrain := NewTerrain(width, height, seed)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Create clearings first (before trees)
	clearings := g.createClearings(terrain, rng)

	// Add water features if chance succeeds
	if rng.Float64() < g.waterChance {
		g.addWaterFeatures(terrain, clearings, rng)
	}

	// Generate trees using Poisson disc sampling
	g.generateTrees(terrain, clearings, rng)

	// Create organic paths between clearings
	g.connectClearings(terrain, clearings, rng)

	// Auto-place bridges where paths cross water
	g.placeAutoBridges(terrain)

	// Place stairs in largest clearings
	g.placeStairsInClearings(terrain, clearings, rng)

	return terrain, nil
}

// createClearings creates circular or elliptical open areas in the forest.
func (g *ForestGenerator) createClearings(terrain *Terrain, rng *rand.Rand) []*Room {
	clearings := make([]*Room, 0)
	attempts := g.clearingCount * 5 // Allow multiple attempts per clearing

	for i := 0; i < attempts && len(clearings) < g.clearingCount; i++ {
		// Random position and size
		width := 8 + rng.Intn(10)  // 8-17 tiles wide
		height := 8 + rng.Intn(10) // 8-17 tiles tall
		x := 2 + rng.Intn(terrain.Width-width-4)
		y := 2 + rng.Intn(terrain.Height-height-4)

		// Create clearing
		clearing := &Room{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
			Type:   RoomNormal,
		}

		// Check for overlap with existing clearings
		overlaps := false
		for _, existing := range clearings {
			if clearing.Overlaps(existing) {
				overlaps = true
				break
			}
		}

		if !overlaps {
			// Create elliptical clearing
			cx, cy := clearing.Center()
			radiusX := float64(width) / 2.0
			radiusY := float64(height) / 2.0

			for dy := 0; dy < height; dy++ {
				for dx := 0; dx < width; dx++ {
					px := x + dx
					py := y + dy

					// Check if point is inside ellipse
					normX := (float64(px) - float64(cx)) / radiusX
					normY := (float64(py) - float64(cy)) / radiusY
					if normX*normX+normY*normY <= 1.0 {
						terrain.SetTile(px, py, TileFloor)
					}
				}
			}

			clearings = append(clearings, clearing)
		}
	}

	terrain.Rooms = clearings
	return clearings
}

// generateTrees places trees using Poisson disc sampling for natural distribution.
func (g *ForestGenerator) generateTrees(terrain *Terrain, clearings []*Room, rng *rand.Rand) {
	// Calculate minimum distance between trees based on density
	minDist := 3.0 / math.Sqrt(g.treeDensity)
	if minDist < 2.0 {
		minDist = 2.0
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

	// Grid to store point indices (-1 = empty)
	grid := make([][]int, gridH)
	for i := range grid {
		grid[i] = make([]int, gridW)
		for j := range grid[i] {
			grid[i][j] = -1
		}
	}

	points := make([]Point, 0)
	activeList := make([]int, 0)

	// Start with random point
	startX := rng.Intn(width)
	startY := rng.Intn(height)
	startPoint := Point{X: startX, Y: startY}
	points = append(points, startPoint)
	activeList = append(activeList, 0)
	startGridX := int(float64(startX) / cellSize)
	startGridY := int(float64(startY) / cellSize)
	if startGridX >= 0 && startGridX < gridW && startGridY >= 0 && startGridY < gridH {
		grid[startGridY][startGridX] = 0
	}

	// Process active list
	for len(activeList) > 0 {
		// Pick random active point
		activeIdx := rng.Intn(len(activeList))
		pointIdx := activeList[activeIdx]
		point := points[pointIdx]

		// Try to generate new points around it
		found := false
		for i := 0; i < 30; i++ { // 30 attempts per point
			// Random point in annulus (minDist to 2*minDist from point)
			angle := rng.Float64() * 2.0 * math.Pi
			radius := minDist * (1.0 + rng.Float64())
			newX := point.X + int(radius*math.Cos(angle))
			newY := point.Y + int(radius*math.Sin(angle))

			// Check if point is valid
			if newX >= 0 && newX < width && newY >= 0 && newY < height {
				gridX := int(float64(newX) / cellSize)
				gridY := int(float64(newY) / cellSize)
				// Ensure grid indices are in bounds
				if gridX >= 0 && gridX < gridW && gridY >= 0 && gridY < gridH {
					newPoint := Point{X: newX, Y: newY}
					if g.isValidPoissonPoint(newPoint, points, grid, cellSize, minDist, width, height) {
						points = append(points, newPoint)
						activeList = append(activeList, len(points)-1)
						grid[gridY][gridX] = len(points) - 1
						found = true
					}
				}
			}
		}

		// Remove from active list if no new points found
		if !found {
			activeList = append(activeList[:activeIdx], activeList[activeIdx+1:]...)
		}
	}

	return points
}

// isValidPoissonPoint checks if a point is valid for Poisson disc sampling.
func (g *ForestGenerator) isValidPoissonPoint(point Point, points []Point, grid [][]int,
	cellSize, minDist float64, width, height int) bool {

	// Get grid cell
	gridX := int(float64(point.X) / cellSize)
	gridY := int(float64(point.Y) / cellSize)

	// Check neighboring cells
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			checkY := gridY + dy
			checkX := gridX + dx

			if checkY >= 0 && checkY < len(grid) && checkX >= 0 && checkX < len(grid[0]) {
				if grid[checkY][checkX] != -1 {
					neighbor := points[grid[checkY][checkX]]
					dist := point.Distance(neighbor)
					if dist < minDist {
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
	// Random center position (avoid clearings)
	maxAttempts := 50
	var centerX, centerY int
	foundValidPos := false

	for attempt := 0; attempt < maxAttempts && !foundValidPos; attempt++ {
		centerX = 10 + rng.Intn(terrain.Width-20)
		centerY = 10 + rng.Intn(terrain.Height-20)

		// Check if far enough from clearings
		farFromClearings := true
		for _, clearing := range clearings {
			cx, cy := clearing.Center()
			dist := math.Sqrt(float64((centerX-cx)*(centerX-cx) + (centerY-cy)*(centerY-cy)))
			if dist < 15.0 {
				farFromClearings = false
				break
			}
		}
		foundValidPos = farFromClearings
	}

	if !foundValidPos {
		return // Skip this lake
	}

	// Random lake size
	radiusX := 4.0 + rng.Float64()*4.0 // 4-8 tiles
	radiusY := 4.0 + rng.Float64()*4.0

	// Create irregular lake using ellipse with noise
	for dy := -int(radiusY) - 2; dy <= int(radiusY)+2; dy++ {
		for dx := -int(radiusX) - 2; dx <= int(radiusX)+2; dx++ {
			x := centerX + dx
			y := centerY + dy

			if !terrain.IsInBounds(x, y) {
				continue
			}

			// Distance from center
			normX := float64(dx) / radiusX
			normY := float64(dy) / radiusY
			dist := math.Sqrt(normX*normX + normY*normY)

			// Add some noise to make it irregular
			noise := rng.Float64()*0.3 - 0.15

			if dist+noise < 0.7 {
				// Deep water in center
				terrain.SetTile(x, y, TileWaterDeep)
			} else if dist+noise < 1.0 {
				// Shallow water at edges
				terrain.SetTile(x, y, TileWaterShallow)
			}
		}
	}
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
