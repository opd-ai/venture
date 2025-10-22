// Package terrain provides terrain type definitions.
// This file defines tile types, terrain data structures, and
// generation parameters used by terrain generators.
package terrain

// TileType represents different types of terrain tiles.
type TileType int

const (
	// TileWall represents a solid wall that blocks movement
	TileWall TileType = iota
	// TileFloor represents walkable floor space
	TileFloor
	// TileDoor represents a doorway between rooms
	TileDoor
	// TileCorridor represents a connecting passage
	TileCorridor
)

// String returns the string representation of a tile type.
func (t TileType) String() string {
	switch t {
	case TileWall:
		return "wall"
	case TileFloor:
		return "floor"
	case TileDoor:
		return "door"
	case TileCorridor:
		return "corridor"
	default:
		return "unknown"
	}
}

// Tile represents a single position in the terrain grid.
type Tile struct {
	Type TileType
	X, Y int
}

// Room represents a rectangular area in the dungeon.
type Room struct {
	X, Y          int // Top-left corner
	Width, Height int
}

// Center returns the center coordinates of the room.
func (r *Room) Center() (int, int) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

// Overlaps checks if this room overlaps with another room.
func (r *Room) Overlaps(other *Room) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// Terrain represents a generated terrain map.
type Terrain struct {
	Width  int
	Height int
	Tiles  [][]TileType
	Rooms  []*Room
	Seed   int64
}

// NewTerrain creates a new terrain map filled with walls.
func NewTerrain(width, height int, seed int64) *Terrain {
	tiles := make([][]TileType, height)
	for y := range tiles {
		tiles[y] = make([]TileType, width)
		for x := range tiles[y] {
			tiles[y][x] = TileWall
		}
	}

	return &Terrain{
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Rooms:  make([]*Room, 0),
		Seed:   seed,
	}
}

// GetTile safely retrieves a tile at the given coordinates.
func (t *Terrain) GetTile(x, y int) TileType {
	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return TileWall
	}
	return t.Tiles[y][x]
}

// SetTile safely sets a tile at the given coordinates.
func (t *Terrain) SetTile(x, y int, tileType TileType) {
	if x >= 0 && x < t.Width && y >= 0 && y < t.Height {
		t.Tiles[y][x] = tileType
	}
}

// IsWalkable returns true if the tile at the given coordinates is walkable.
func (t *Terrain) IsWalkable(x, y int) bool {
	tile := t.GetTile(x, y)
	return tile == TileFloor || tile == TileDoor || tile == TileCorridor
}
