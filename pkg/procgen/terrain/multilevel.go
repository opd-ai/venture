package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// LevelGenerator generates connected multi-level dungeons.
// It manages the creation of multiple terrain levels and ensures
// proper connectivity between them via stairs.
type LevelGenerator struct {
	generators map[int]procgen.Generator // Depth -> Generator mapping
}

// NewLevelGenerator creates a new multi-level generator.
// By default, it uses BSP for all depths, but generators can be
// customized per depth using SetGenerator.
func NewLevelGenerator() *LevelGenerator {
	return &LevelGenerator{
		generators: make(map[int]procgen.Generator),
	}
}

// SetGenerator assigns a specific generator to a depth level.
// This allows mixing different generation algorithms across levels.
// For example: BSP for upper levels, Cellular for caves, Maze for deep levels.
func (g *LevelGenerator) SetGenerator(depth int, gen procgen.Generator) {
	g.generators[depth] = gen
}

// GetGenerator retrieves the generator for a specific depth.
// Returns BSP generator as default if no specific generator is set.
func (g *LevelGenerator) GetGenerator(depth int) procgen.Generator {
	if gen, ok := g.generators[depth]; ok {
		return gen
	}
	// Default to BSP for all levels
	return NewBSPGenerator()
}

// GenerateMultiLevel creates a connected multi-level dungeon.
// Each level uses the appropriate generator for its depth, with difficulty
// scaling as levels go deeper. Stairs are automatically placed to connect levels.
//
// Parameters:
//   - numLevels: Total number of levels to generate (1-20)
//   - seed: Base seed for deterministic generation
//   - params: Base generation parameters (difficulty scales with depth)
//
// Returns a slice of Terrain, one per level, with stairs properly connected.
func (g *LevelGenerator) GenerateMultiLevel(numLevels int, seed int64, params procgen.GenerationParams) ([]*Terrain, error) {
	if numLevels < 1 || numLevels > 20 {
		return nil, fmt.Errorf("numLevels must be between 1 and 20, got %d", numLevels)
	}

	levels := make([]*Terrain, numLevels)
	rng := rand.New(rand.NewSource(seed))

	// Generate each level
	for i := 0; i < numLevels; i++ {
		// Scale difficulty with depth
		levelParams := params
		levelParams.Depth = i + 1
		levelParams.Difficulty = params.Difficulty + float64(i)*0.1
		if levelParams.Difficulty > 1.0 {
			levelParams.Difficulty = 1.0
		}

		// Use depth-specific seed
		levelSeed := seed + int64(i)*1000

		// Get generator for this depth
		gen := g.GetGenerator(i)

		// Generate the level
		result, err := gen.Generate(levelSeed, levelParams)
		if err != nil {
			return nil, fmt.Errorf("failed to generate level %d: %w", i, err)
		}

		terrain := result.(*Terrain)
		terrain.Level = i
		levels[i] = terrain
	}

	// Connect levels with stairs
	for i := 0; i < numLevels-1; i++ {
		if err := g.ConnectLevels(levels[i], levels[i+1], rng); err != nil {
			return nil, fmt.Errorf("failed to connect levels %d and %d: %w", i, i+1, err)
		}
	}

	// Validate all connections
	if err := g.ValidateMultiLevelConnectivity(levels); err != nil {
		return nil, fmt.Errorf("multi-level connectivity validation failed: %w", err)
	}

	return levels, nil
}

// ConnectLevels connects two adjacent levels by placing stairs.
// Stairs down in the upper level should roughly align with stairs up in the lower level.
// This creates a sense of vertical continuity in the dungeon.
func (g *LevelGenerator) ConnectLevels(above, below *Terrain, rng *rand.Rand) error {
	// Place stairs down in upper level
	stairsDownPos, err := g.findStairLocation(above, rng)
	if err != nil {
		return fmt.Errorf("failed to find stairs down location in level %d: %w", above.Level, err)
	}
	above.AddStairs(stairsDownPos.X, stairsDownPos.Y, false)

	// Place stairs up in lower level, roughly aligned with stairs down
	stairsUpPos, err := g.findAlignedStairLocation(below, stairsDownPos, rng)
	if err != nil {
		// Fallback: place stairs anywhere if alignment fails
		stairsUpPos, err = g.findStairLocation(below, rng)
		if err != nil {
			return fmt.Errorf("failed to find stairs up location in level %d: %w", below.Level, err)
		}
	}
	below.AddStairs(stairsUpPos.X, stairsUpPos.Y, true)

	return nil
}

// findStairLocation finds a suitable location for stairs in a terrain.
// Prefers room centers, but will use any walkable floor tile if needed.
func (g *LevelGenerator) findStairLocation(terrain *Terrain, rng *rand.Rand) (Point, error) {
	// First try: place in a room center
	if len(terrain.Rooms) > 0 {
		room := terrain.Rooms[rng.Intn(len(terrain.Rooms))]
		cx, cy := room.Center()
		if terrain.IsWalkable(cx, cy) {
			return Point{X: cx, Y: cy}, nil
		}
	}

	// Second try: find any walkable floor tile
	walkable := make([]Point, 0, 100)
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileFloor && terrain.IsWalkable(x, y) {
				walkable = append(walkable, Point{X: x, Y: y})
			}
		}
	}

	if len(walkable) == 0 {
		return Point{}, fmt.Errorf("no walkable floor tiles found for stairs")
	}

	return walkable[rng.Intn(len(walkable))], nil
}

// findAlignedStairLocation finds a stair location roughly aligned with the stairs from above.
// Searches in a radius around the aligned position, falling back to any walkable tile.
func (g *LevelGenerator) findAlignedStairLocation(terrain *Terrain, abovePos Point, rng *rand.Rand) (Point, error) {
	// Search radius for alignment (within ~10 tiles)
	searchRadius := 10

	candidates := make([]Point, 0, 50)

	// Collect candidates within radius
	for dy := -searchRadius; dy <= searchRadius; dy++ {
		for dx := -searchRadius; dx <= searchRadius; dx++ {
			x := abovePos.X + dx
			y := abovePos.Y + dy
			if terrain.IsInBounds(x, y) && terrain.GetTile(x, y) == TileFloor && terrain.IsWalkable(x, y) {
				candidates = append(candidates, Point{X: x, Y: y})
			}
		}
	}

	if len(candidates) == 0 {
		return Point{}, fmt.Errorf("no walkable tiles near aligned position (%d, %d)", abovePos.X, abovePos.Y)
	}

	return candidates[rng.Intn(len(candidates))], nil
}

// ValidateMultiLevelConnectivity ensures all levels are properly connected.
// Checks that:
// - Every level except first has stairs up
// - Every level except last has stairs down
// - All stairs are placed in walkable areas
func (g *LevelGenerator) ValidateMultiLevelConnectivity(levels []*Terrain) error {
	if len(levels) == 0 {
		return fmt.Errorf("no levels to validate")
	}

	for i, level := range levels {
		// First level should have stairs down (except if only one level)
		if i == 0 && len(levels) > 1 {
			if len(level.StairsDown) == 0 {
				return fmt.Errorf("level %d (first level) missing stairs down", i)
			}
		}

		// Middle levels should have both stairs
		if i > 0 && i < len(levels)-1 {
			if len(level.StairsUp) == 0 {
				return fmt.Errorf("level %d (middle level) missing stairs up", i)
			}
			if len(level.StairsDown) == 0 {
				return fmt.Errorf("level %d (middle level) missing stairs down", i)
			}
		}

		// Last level should have stairs up
		if i == len(levels)-1 && len(levels) > 1 {
			if len(level.StairsUp) == 0 {
				return fmt.Errorf("level %d (last level) missing stairs up", i)
			}
		}

		// Validate all stairs are in walkable positions
		for _, stair := range level.StairsUp {
			if !level.IsWalkable(stair.X, stair.Y) {
				return fmt.Errorf("level %d stairs up at (%d, %d) is not walkable", i, stair.X, stair.Y)
			}
		}
		for _, stair := range level.StairsDown {
			if !level.IsWalkable(stair.X, stair.Y) {
				return fmt.Errorf("level %d stairs down at (%d, %d) is not walkable", i, stair.X, stair.Y)
			}
		}
	}

	return nil
}

// PlaceStairsRandom places stairs at random walkable locations in the terrain.
// This is a simple strategy suitable for any terrain type.
func PlaceStairsRandom(terrain *Terrain, up, down bool, rng *rand.Rand) error {
	walkable := make([]Point, 0, 100)
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileFloor && terrain.IsWalkable(x, y) {
				walkable = append(walkable, Point{X: x, Y: y})
			}
		}
	}

	if len(walkable) == 0 {
		return fmt.Errorf("no walkable tiles for stairs")
	}

	if up {
		pos := walkable[rng.Intn(len(walkable))]
		terrain.AddStairs(pos.X, pos.Y, true)
	}

	if down {
		pos := walkable[rng.Intn(len(walkable))]
		terrain.AddStairs(pos.X, pos.Y, false)
	}

	return nil
}

// PlaceStairsInRoom places stairs in a specific room type.
// Useful for placing stairs in boss rooms, treasure rooms, etc.
func PlaceStairsInRoom(terrain *Terrain, roomType RoomType, up, down bool, rng *rand.Rand) error {
	// Find rooms of the specified type
	candidates := make([]*Room, 0, len(terrain.Rooms))
	for _, room := range terrain.Rooms {
		if room.Type == roomType {
			candidates = append(candidates, room)
		}
	}

	if len(candidates) == 0 {
		return fmt.Errorf("no rooms of type %v found for stairs", roomType)
	}

	room := candidates[rng.Intn(len(candidates))]
	cx, cy := room.Center()

	if up {
		terrain.AddStairs(cx, cy, true)
	}

	if down {
		// Place stairs down in a different location within the room
		offsetX := rng.Intn(3) - 1 // -1, 0, or 1
		offsetY := rng.Intn(3) - 1
		dx := cx + offsetX
		dy := cy + offsetY
		if terrain.IsInBounds(dx, dy) && terrain.IsWalkable(dx, dy) {
			terrain.AddStairs(dx, dy, false)
		} else {
			terrain.AddStairs(cx, cy, false) // Fallback to center
		}
	}

	return nil
}

// PlaceStairsSymmetric places stairs in opposite corners or edges of the terrain.
// Creates visual balance and encourages full exploration of the level.
func PlaceStairsSymmetric(terrain *Terrain, up, down bool, rng *rand.Rand) error {
	// Define corner regions (each 1/4 of the map)
	quarterWidth := terrain.Width / 4
	quarterHeight := terrain.Height / 4

	corners := []struct {
		x, y, w, h int
	}{
		{0, 0, quarterWidth, quarterHeight},                                                         // Top-left
		{terrain.Width - quarterWidth, 0, quarterWidth, quarterHeight},                              // Top-right
		{0, terrain.Height - quarterHeight, quarterWidth, quarterHeight},                            // Bottom-left
		{terrain.Width - quarterWidth, terrain.Height - quarterHeight, quarterWidth, quarterHeight}, // Bottom-right
	}

	// Find walkable tiles in each corner
	cornerTiles := make([][]Point, 4)
	for i, corner := range corners {
		tiles := make([]Point, 0, 20)
		for y := corner.y; y < corner.y+corner.h && y < terrain.Height; y++ {
			for x := corner.x; x < corner.x+corner.w && x < terrain.Width; x++ {
				if terrain.GetTile(x, y) == TileFloor && terrain.IsWalkable(x, y) {
					tiles = append(tiles, Point{X: x, Y: y})
				}
			}
		}
		cornerTiles[i] = tiles
	}

	// Find corners with available tiles
	available := make([]int, 0, 4)
	for i, tiles := range cornerTiles {
		if len(tiles) > 0 {
			available = append(available, i)
		}
	}

	if len(available) < 2 {
		return fmt.Errorf("insufficient walkable corners for symmetric stairs")
	}

	// Place stairs up in one corner
	if up {
		corner1 := available[rng.Intn(len(available))]
		tiles := cornerTiles[corner1]
		pos := tiles[rng.Intn(len(tiles))]
		terrain.AddStairs(pos.X, pos.Y, true)

		// Remove this corner from available
		available = removeElement(available, corner1)
	}

	// Place stairs down in opposite corner
	if down && len(available) > 0 {
		corner2 := available[rng.Intn(len(available))]
		tiles := cornerTiles[corner2]
		pos := tiles[rng.Intn(len(tiles))]
		terrain.AddStairs(pos.X, pos.Y, false)
	}

	return nil
}

// removeElement removes the first occurrence of val from slice.
func removeElement(slice []int, val int) []int {
	for i, v := range slice {
		if v == val {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
