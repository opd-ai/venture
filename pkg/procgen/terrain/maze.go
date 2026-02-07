// Package terrain provides maze generation using recursive backtracking.
// This file implements a maze generation algorithm that creates complex,
// winding corridors with optional rooms at dead ends.
package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// MazeGenerator generates mazes using recursive backtracking algorithm.
// It creates complex, winding corridors with optional rooms at dead ends.
type MazeGenerator struct {
	roomChance    float64 // Probability (0.0-1.0) of creating a room at a dead end
	corridorWidth int     // Width of corridors (1 = single tile, 2 = double-wide)
	logger        *logrus.Entry
}

// NewMazeGenerator creates a new maze generator with default parameters.
func NewMazeGenerator() *MazeGenerator {
	return NewMazeGeneratorWithLogger(nil)
}

// NewMazeGeneratorWithLogger creates a new maze generator with a logger.
func NewMazeGeneratorWithLogger(logger *logrus.Logger) *MazeGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "maze")
	}
	return &MazeGenerator{
		roomChance:    0.1, // 10% of dead ends become rooms
		corridorWidth: 1,   // Single-tile corridors
		logger:        logEntry,
	}
}

// Generate creates a maze using recursive backtracking algorithm.
func (g *MazeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"genreID":    params.GenreID,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
		}).Debug("starting maze terrain generation")
	}

	// Extract and validate dimensions
	width, height, err := g.extractDimensions(params)
	if err != nil {
		return nil, err
	}

	// Create RNG with seed
	rng := rand.New(rand.NewSource(seed))

	// Create terrain (starts filled with walls)
	terrain := NewTerrain(width, height, seed)

	// Generate maze structure
	g.generateMazeStructure(terrain, rng)

	// Add rooms and hazards to enhance maze
	g.enhanceMazeWithFeatures(terrain, rng)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":  terrain.Width,
			"height": terrain.Height,
		}).Info("maze terrain generation complete")
	}

	return terrain, nil
}

// extractDimensions extracts and validates maze dimensions from parameters.
func (g *MazeGenerator) extractDimensions(params procgen.GenerationParams) (int, int, error) {
	width, height := g.parseCustomDimensions(params)

	if err := g.validateDimensions(width, height); err != nil {
		return 0, 0, err
	}

	width, height = g.normalizeToOddDimensions(width, height)
	return width, height, nil
}

// parseCustomDimensions extracts dimensions and settings from custom parameters.
func (g *MazeGenerator) parseCustomDimensions(params procgen.GenerationParams) (int, int) {
	width := 80
	height := 50

	if params.Custom == nil {
		return width, height
	}

	if w, ok := params.Custom["width"].(int); ok {
		width = w
	}
	if h, ok := params.Custom["height"].(int); ok {
		height = h
	}
	if rc, ok := params.Custom["roomChance"].(float64); ok {
		g.roomChance = rc
	}
	if cw, ok := params.Custom["corridorWidth"].(int); ok {
		g.corridorWidth = cw
	}

	return width, height
}

// validateDimensions checks if dimensions are within valid bounds.
func (g *MazeGenerator) validateDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}

	if width > 1000 || height > 1000 {
		return fmt.Errorf("dimensions too large: width=%d, height=%d (max 1000x1000)", width, height)
	}

	return nil
}

// normalizeToOddDimensions ensures dimensions are odd for maze algorithm.
func (g *MazeGenerator) normalizeToOddDimensions(width, height int) (int, int) {
	if width%2 == 0 {
		width++
	}
	if height%2 == 0 {
		height++
	}
	return width, height
}

// generateMazeStructure creates the basic maze layout using recursive backtracking.
func (g *MazeGenerator) generateMazeStructure(terrain *Terrain, rng *rand.Rand) {
	// Start maze generation from a random odd position
	startX := 1 + (rng.Intn((terrain.Width-2)/2) * 2)
	startY := 1 + (rng.Intn((terrain.Height-2)/2) * 2)

	// Carve passages using recursive backtracking
	g.carvePassages(startX, startY, terrain, rng)
}

// enhanceMazeWithFeatures adds rooms, water hazards, and stairs to the maze.
func (g *MazeGenerator) enhanceMazeWithFeatures(terrain *Terrain, rng *rand.Rand) {
	// Find dead ends and potentially create rooms
	deadEnds := g.findDeadEnds(terrain)
	for _, point := range deadEnds {
		if rng.Float64() < g.roomChance {
			g.createRoomAtDeadEnd(point.X, point.Y, terrain, rng)
		}
	}

	// Add water hazards to some remaining dead ends (20% chance)
	g.addWaterHazards(terrain, deadEnds, rng)

	// Place stairs at furthest corners
	g.placeStairsInCorners(terrain, rng)
}

// carvePassages recursively carves passages through the maze using backtracking.
func (g *MazeGenerator) carvePassages(x, y int, terrain *Terrain, rng *rand.Rand) {
	terrain.SetTile(x, y, TileFloor)

	directions := g.shuffleMazeDirections(rng)

	for _, dir := range directions {
		nx := x + dir.dx
		ny := y + dir.dy

		if g.isValidUnvisitedCell(nx, ny, terrain) {
			g.carveCorridorWall(x, y, dir, terrain)
			g.carvePassages(nx, ny, terrain, rng)
		}
	}
}

// shuffleMazeDirections returns randomized cardinal directions for maze carving.
func (g *MazeGenerator) shuffleMazeDirections(rng *rand.Rand) []struct{ dx, dy int } {
	directions := []struct{ dx, dy int }{
		{0, -2}, {2, 0}, {0, 2}, {-2, 0},
	}
	for i := len(directions) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		directions[i], directions[j] = directions[j], directions[i]
	}
	return directions
}

// isValidUnvisitedCell checks if coordinates are valid and unvisited wall tiles.
func (g *MazeGenerator) isValidUnvisitedCell(x, y int, terrain *Terrain) bool {
	if x <= 0 || x >= terrain.Width-1 || y <= 0 || y >= terrain.Height-1 {
		return false
	}
	return terrain.GetTile(x, y) == TileWall
}

// carveCorridorWall carves the wall between current position and next cell.
func (g *MazeGenerator) carveCorridorWall(x, y int, dir struct{ dx, dy int }, terrain *Terrain) {
	wallX := x + dir.dx/2
	wallY := y + dir.dy/2
	terrain.SetTile(wallX, wallY, TileCorridor)

	if g.corridorWidth == 2 {
		g.carveAdjacentCorridorTile(wallX, wallY, dir, terrain)
	}
}

// carveAdjacentCorridorTile carves an additional tile for wide corridors.
func (g *MazeGenerator) carveAdjacentCorridorTile(wallX, wallY int, dir struct{ dx, dy int }, terrain *Terrain) {
	if dir.dx != 0 && wallY > 0 {
		terrain.SetTile(wallX, wallY-1, TileCorridor)
	} else if dir.dy != 0 && wallX > 0 {
		terrain.SetTile(wallX-1, wallY, TileCorridor)
	}
}

// findDeadEnds identifies all dead ends in the maze (cells with only one neighbor).
func (g *MazeGenerator) findDeadEnds(terrain *Terrain) []Point {
	deadEnds := make([]Point, 0)

	for y := 1; y < terrain.Height-1; y++ {
		for x := 1; x < terrain.Width-1; x++ {
			// Check if this is a floor tile
			if !terrain.IsWalkable(x, y) {
				continue
			}

			// Count walkable neighbors (orthogonal only)
			neighbors := 0
			if terrain.IsWalkable(x, y-1) {
				neighbors++
			}
			if terrain.IsWalkable(x+1, y) {
				neighbors++
			}
			if terrain.IsWalkable(x, y+1) {
				neighbors++
			}
			if terrain.IsWalkable(x-1, y) {
				neighbors++
			}

			// Dead end has exactly one neighbor
			if neighbors == 1 {
				deadEnds = append(deadEnds, Point{X: x, Y: y})
			}
		}
	}

	return deadEnds
}

// createRoomAtDeadEnd creates a small room at a dead end location.
func (g *MazeGenerator) createRoomAtDeadEnd(x, y int, terrain *Terrain, rng *rand.Rand) {
	// Random room size (3x3 to 7x7)
	roomWidth := 3 + rng.Intn(5)
	roomHeight := 3 + rng.Intn(5)

	// Calculate room position (centered on dead end)
	roomX := x - roomWidth/2
	roomY := y - roomHeight/2

	// Ensure room is within bounds
	if roomX < 1 {
		roomX = 1
	}
	if roomY < 1 {
		roomY = 1
	}
	if roomX+roomWidth >= terrain.Width-1 {
		roomX = terrain.Width - roomWidth - 1
	}
	if roomY+roomHeight >= terrain.Height-1 {
		roomY = terrain.Height - roomHeight - 1
	}

	// Create the room
	for ry := roomY; ry < roomY+roomHeight && ry < terrain.Height-1; ry++ {
		for rx := roomX; rx < roomX+roomWidth && rx < terrain.Width-1; rx++ {
			terrain.SetTile(rx, ry, TileFloor)
		}
	}

	// Add room to terrain's room list
	room := &Room{
		X:      roomX,
		Y:      roomY,
		Width:  roomWidth,
		Height: roomHeight,
		Type:   RoomNormal,
	}
	terrain.Rooms = append(terrain.Rooms, room)
}

// placeStairsInCorners places stairs up and down in opposite corners of the maze.
func (g *MazeGenerator) placeStairsInCorners(terrain *Terrain, rng *rand.Rand) {
	cornerTiles := g.collectCornerTiles(terrain)
	upCornerIdx := g.placeStairsUp(terrain, cornerTiles, rng)
	g.placeStairsDown(terrain, cornerTiles, rng, upCornerIdx)
}

// collectCornerTiles finds walkable tiles in each corner region.
func (g *MazeGenerator) collectCornerTiles(terrain *Terrain) [][]Point {
	const cornerSize = 10
	corners := []struct {
		name   string
		x, y   int
		width  int
		height int
	}{
		{"top-left", 1, 1, cornerSize, cornerSize},
		{"top-right", terrain.Width - cornerSize - 1, 1, cornerSize, cornerSize},
		{"bottom-left", 1, terrain.Height - cornerSize - 1, cornerSize, cornerSize},
		{"bottom-right", terrain.Width - cornerSize - 1, terrain.Height - cornerSize - 1, cornerSize, cornerSize},
	}

	cornerTiles := make([][]Point, len(corners))
	for i, corner := range corners {
		for y := corner.y; y < corner.y+corner.height && y < terrain.Height-1; y++ {
			for x := corner.x; x < corner.x+corner.width && x < terrain.Width-1; x++ {
				if terrain.IsWalkable(x, y) {
					cornerTiles[i] = append(cornerTiles[i], Point{X: x, Y: y})
				}
			}
		}
	}

	return cornerTiles
}

// placeStairsUp places stairs up in a random corner with walkable tiles.
func (g *MazeGenerator) placeStairsUp(terrain *Terrain, cornerTiles [][]Point, rng *rand.Rand) int {
	const maxAttempts = 10
	upCornerIdx := -1

	for attempts := 0; attempts < maxAttempts; attempts++ {
		cornerIdx := rng.Intn(len(cornerTiles))
		if len(cornerTiles[cornerIdx]) > 0 {
			tileIdx := rng.Intn(len(cornerTiles[cornerIdx]))
			point := cornerTiles[cornerIdx][tileIdx]
			terrain.AddStairs(point.X, point.Y, true)
			upCornerIdx = cornerIdx
			break
		}
	}

	return upCornerIdx
}

// placeStairsDown places stairs down in opposite corner from stairs up.
func (g *MazeGenerator) placeStairsDown(terrain *Terrain, cornerTiles [][]Point, rng *rand.Rand, upCornerIdx int) {
	const maxAttempts = 10

	for attempts := 0; attempts < maxAttempts; attempts++ {
		oppositeIdx := g.getOppositeCornerIndex(terrain, upCornerIdx)

		if len(cornerTiles[oppositeIdx]) > 0 {
			tileIdx := rng.Intn(len(cornerTiles[oppositeIdx]))
			point := cornerTiles[oppositeIdx][tileIdx]
			terrain.AddStairs(point.X, point.Y, false)
			break
		}
	}
}

// getOppositeCornerIndex determines the opposite corner index.
func (g *MazeGenerator) getOppositeCornerIndex(terrain *Terrain, upCornerIdx int) int {
	if len(terrain.StairsUp) == 0 {
		return 0
	}

	upPoint := terrain.StairsUp[0]
	isLeft := upPoint.X < terrain.Width/2
	isTop := upPoint.Y < terrain.Height/2

	if isLeft && isTop {
		return 3 // bottom-right
	} else if !isLeft && isTop {
		return 2 // bottom-left
	} else if isLeft && !isTop {
		return 1 // top-right
	}
	return 0 // top-left
}

// Validate checks if the generated maze meets quality requirements.
func (g *MazeGenerator) Validate(result interface{}) error {
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

	// Ensure at least 20% of tiles are walkable (mazes are typically sparser than dungeons)
	totalTiles := terrain.Width * terrain.Height
	if float64(walkable)/float64(totalTiles) < 0.2 {
		return fmt.Errorf("insufficient walkable tiles: %d/%d (%.1f%%, need >= 20%%)",
			walkable, totalTiles, float64(walkable)/float64(totalTiles)*100)
	}

	// Validate stair placement if stairs exist
	if len(terrain.StairsUp) > 0 || len(terrain.StairsDown) > 0 {
		if err := terrain.ValidateStairPlacement(); err != nil {
			return fmt.Errorf("stair validation failed: %w", err)
		}
	}

	return nil
}

// addWaterHazards fills some dead ends with water for additional maze challenges.
// Creates small water pools (2-3 tiles) at dead ends that weren't converted to rooms.
func (g *MazeGenerator) addWaterHazards(terrain *Terrain, deadEnds []Point, rng *rand.Rand) {
	for _, point := range deadEnds {
		if !g.isValidWaterLocation(terrain, point) {
			continue
		}

		if rng.Float64() < 0.2 {
			g.placeWaterPool(terrain, point, rng)
		}
	}
}

// isValidWaterLocation checks if a point is valid for water placement.
func (g *MazeGenerator) isValidWaterLocation(terrain *Terrain, point Point) bool {
	if terrain.GetTile(point.X, point.Y) != TileFloor {
		return false
	}
	return !g.isPointInRoom(terrain, point)
}

// isPointInRoom checks if a point is inside any room.
func (g *MazeGenerator) isPointInRoom(terrain *Terrain, point Point) bool {
	for _, room := range terrain.Rooms {
		if point.X >= room.X && point.X < room.X+room.Width &&
			point.Y >= room.Y && point.Y < room.Y+room.Height {
			return true
		}
	}
	return false
}

// placeWaterPool creates a water pool at the dead end.
func (g *MazeGenerator) placeWaterPool(terrain *Terrain, point Point, rng *rand.Rand) {
	poolSize := 2 + rng.Intn(2)
	directions := []struct{ dx, dy int }{
		{0, -1}, {1, 0}, {0, 1}, {-1, 0},
	}

	for _, dir := range directions {
		nx, ny := point.X+dir.dx, point.Y+dir.dy
		if terrain.IsInBounds(nx, ny) && terrain.GetTile(nx, ny) == TileFloor {
			terrain.SetTile(point.X, point.Y, TileWaterDeep)
			g.addShallowWaterTiles(terrain, point, dir, poolSize)
			break
		}
	}
}

// addShallowWaterTiles adds shallow water tiles toward corridor.
func (g *MazeGenerator) addShallowWaterTiles(terrain *Terrain, point Point, dir struct{ dx, dy int }, poolSize int) {
	for i := 1; i < poolSize; i++ {
		wx := point.X - dir.dx*i
		wy := point.Y - dir.dy*i
		if terrain.IsInBounds(wx, wy) && terrain.GetTile(wx, wy) == TileFloor {
			terrain.SetTile(wx, wy, TileWaterShallow)
		}
	}
}
