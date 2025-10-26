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
		if rc, ok := params.Custom["roomChance"].(float64); ok {
			g.roomChance = rc
		}
		if cw, ok := params.Custom["corridorWidth"].(int); ok {
			g.corridorWidth = cw
		}
	}

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width=%d, height=%d (must be positive)", width, height)
	}

	if width > 1000 || height > 1000 {
		return nil, fmt.Errorf("dimensions too large: width=%d, height=%d (max 1000x1000)", width, height)
	}

	// Ensure dimensions are odd for maze algorithm (walls on edges, floors in between)
	if width%2 == 0 {
		width++
	}
	if height%2 == 0 {
		height++
	}

	// Create RNG with seed
	rng := rand.New(rand.NewSource(seed))

	// Create terrain (starts filled with walls)
	terrain := NewTerrain(width, height, seed)

	// Start maze generation from a random odd position
	startX := 1 + (rng.Intn((width-2)/2) * 2)
	startY := 1 + (rng.Intn((height-2)/2) * 2)

	// Carve passages using recursive backtracking
	g.carvePassages(startX, startY, terrain, rng)

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

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"width":    terrain.Width,
			"height":   terrain.Height,
			"deadEnds": len(deadEnds),
		}).Info("maze terrain generation complete")
	}

	return terrain, nil
}

// carvePassages recursively carves passages through the maze using backtracking.
func (g *MazeGenerator) carvePassages(x, y int, terrain *Terrain, rng *rand.Rand) {
	// Mark current cell as floor
	terrain.SetTile(x, y, TileFloor)

	// Define directions: North, East, South, West
	directions := []struct{ dx, dy int }{
		{0, -2}, // North
		{2, 0},  // East
		{0, 2},  // South
		{-2, 0}, // West
	}

	// Shuffle directions for randomness
	for i := len(directions) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		directions[i], directions[j] = directions[j], directions[i]
	}

	// Try each direction
	for _, dir := range directions {
		nx := x + dir.dx
		ny := y + dir.dy

		// Check if new position is valid and unvisited
		if nx > 0 && nx < terrain.Width-1 && ny > 0 && ny < terrain.Height-1 {
			if terrain.GetTile(nx, ny) == TileWall {
				// Carve the wall between current and next cell
				wallX := x + dir.dx/2
				wallY := y + dir.dy/2
				terrain.SetTile(wallX, wallY, TileCorridor)

				// If corridor width is 2, carve adjacent tile
				if g.corridorWidth == 2 {
					if dir.dx != 0 {
						// Horizontal movement - carve tile above or below
						if wallY > 0 {
							terrain.SetTile(wallX, wallY-1, TileCorridor)
						}
					} else {
						// Vertical movement - carve tile left or right
						if wallX > 0 {
							terrain.SetTile(wallX-1, wallY, TileCorridor)
						}
					}
				}

				// Recursively carve from new position
				g.carvePassages(nx, ny, terrain, rng)
			}
		}
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
	// Find walkable tiles in each corner region
	cornerSize := 10
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

	// Collect walkable positions in each corner
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

	// Place stairs up in a random corner with walkable tiles
	stairsUpPlaced := false
	attempts := 0
	for !stairsUpPlaced && attempts < 10 {
		cornerIdx := rng.Intn(len(corners))
		if len(cornerTiles[cornerIdx]) > 0 {
			tileIdx := rng.Intn(len(cornerTiles[cornerIdx]))
			point := cornerTiles[cornerIdx][tileIdx]
			terrain.AddStairs(point.X, point.Y, true)
			stairsUpPlaced = true
		}
		attempts++
	}

	// Place stairs down in opposite corner
	stairsDownPlaced := false
	attempts = 0
	for !stairsDownPlaced && attempts < 10 {
		// Choose opposite corner (0↔3, 1↔2)
		var oppositeIdx int
		if len(terrain.StairsUp) > 0 {
			upPoint := terrain.StairsUp[0]
			// Determine which corner based on position
			isLeft := upPoint.X < terrain.Width/2
			isTop := upPoint.Y < terrain.Height/2

			if isLeft && isTop {
				oppositeIdx = 3 // bottom-right
			} else if !isLeft && isTop {
				oppositeIdx = 2 // bottom-left
			} else if isLeft && !isTop {
				oppositeIdx = 1 // top-right
			} else {
				oppositeIdx = 0 // top-left
			}
		} else {
			oppositeIdx = rng.Intn(len(corners))
		}

		if len(cornerTiles[oppositeIdx]) > 0 {
			tileIdx := rng.Intn(len(cornerTiles[oppositeIdx]))
			point := cornerTiles[oppositeIdx][tileIdx]
			terrain.AddStairs(point.X, point.Y, false)
			stairsDownPlaced = true
		}
		attempts++
	}
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
		// Skip if this dead end was turned into a room or stairs
		tile := terrain.GetTile(point.X, point.Y)
		if tile != TileFloor {
			continue
		}

		// Skip if this point is part of a room
		isInRoom := false
		for _, room := range terrain.Rooms {
			if point.X >= room.X && point.X < room.X+room.Width &&
				point.Y >= room.Y && point.Y < room.Y+room.Height {
				isInRoom = true
				break
			}
		}
		if isInRoom {
			continue
		}

		// 20% chance to create water hazard
		if rng.Float64() < 0.2 {
			// Create small water pool (2-3 tiles from the dead end)
			poolSize := 2 + rng.Intn(2)

			// Find the direction of the corridor leading to this dead end
			directions := []struct{ dx, dy int }{
				{0, -1}, // North
				{1, 0},  // East
				{0, 1},  // South
				{-1, 0}, // West
			}

			for _, dir := range directions {
				nx, ny := point.X+dir.dx, point.Y+dir.dy
				if terrain.IsInBounds(nx, ny) && terrain.GetTile(nx, ny) == TileFloor {
					// Found corridor direction, place water in dead end
					terrain.SetTile(point.X, point.Y, TileWaterDeep)

					// Add 1-2 more shallow water tiles back toward corridor
					for i := 1; i < poolSize; i++ {
						wx := point.X - dir.dx*i
						wy := point.Y - dir.dy*i
						if terrain.IsInBounds(wx, wy) && terrain.GetTile(wx, wy) == TileFloor {
							terrain.SetTile(wx, wy, TileWaterShallow)
						}
					}
					break
				}
			}
		}
	}
}
