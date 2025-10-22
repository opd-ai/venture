package world

// TileType represents different types of terrain tiles.
type TileType int

// Tile type constants.
const (
	TileEmpty TileType = iota
	TileFloor
	TileWall
	TileDoor
	TileWater
	TileLava
	TileGrass
	TileStone
)

// Tile represents a single tile in the world.
type Tile struct {
	Type     TileType
	Walkable bool
	X        int
	Y        int
}

// Map represents a game map/level.
type Map struct {
	// Width and height in tiles
	Width  int
	Height int

	// Tiles stored in row-major order
	Tiles []Tile

	// Seed used to generate this map
	Seed int64

	// Genre of this map
	Genre string
}

// NewMap creates a new map with the given dimensions.
func NewMap(width, height int, seed int64) *Map {
	tiles := make([]Tile, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			tiles[idx] = Tile{
				Type:     TileEmpty,
				Walkable: true,
				X:        x,
				Y:        y,
			}
		}
	}

	return &Map{
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Seed:   seed,
	}
}

// GetTile returns the tile at the given coordinates.
func (m *Map) GetTile(x, y int) *Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil
	}
	idx := y*m.Width + x
	return &m.Tiles[idx]
}

// SetTile updates the tile at the given coordinates.
func (m *Map) SetTile(x, y int, tile Tile) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}
	idx := y*m.Width + x
	tile.X = x
	tile.Y = y
	m.Tiles[idx] = tile
}

// IsWalkable checks if a position is walkable.
func (m *Map) IsWalkable(x, y int) bool {
	tile := m.GetTile(x, y)
	if tile == nil {
		return false
	}
	return tile.Walkable
}

// WorldState represents the complete state of the game world.
type WorldState struct {
	// Current map
	CurrentMap *Map

	// Global game time
	Time float64

	// Player entity IDs
	PlayerIDs []uint64

	// Additional state data
	State map[string]interface{}
}

// NewWorldState creates a new world state.
func NewWorldState() *WorldState {
	return &WorldState{
		PlayerIDs: make([]uint64, 0),
		State:     make(map[string]interface{}),
	}
}
