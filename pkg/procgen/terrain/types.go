// Package terrain provides terrain type definitions.
// This file defines tile types, terrain data structures, and
// generation parameters used by terrain generators.
package terrain

import "fmt"

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
	// TileWaterShallow represents shallow water that is walkable but slows movement
	TileWaterShallow
	// TileWaterDeep represents deep water that blocks movement
	TileWaterDeep
	// TileTree represents a tree or natural obstacle that blocks movement
	TileTree
	// TileStairsUp represents stairs leading to an upper level
	TileStairsUp
	// TileStairsDown represents stairs leading to a lower level
	TileStairsDown
	// TileTrapDoor represents a hidden or revealed trap door
	TileTrapDoor
	// TileSecretDoor represents a hidden door
	TileSecretDoor
	// TileBridge represents a walkable bridge over water
	TileBridge
	// TileStructure represents a building or ruins
	TileStructure

	// Diagonal wall tiles (45° angles)
	// TileWallNE represents a diagonal wall from bottom-left to top-right (/)
	TileWallNE
	// TileWallNW represents a diagonal wall from bottom-right to top-left (\)
	TileWallNW
	// TileWallSE represents a diagonal wall from top-left to bottom-right (\)
	TileWallSE
	// TileWallSW represents a diagonal wall from top-right to bottom-left (/)
	TileWallSW

	// Multi-layer terrain tiles
	// TilePlatform represents an elevated platform entities can walk on
	TilePlatform
	// TileRamp represents a ramp for transitioning between layers
	TileRamp
	// TileLavaFlow represents flowing lava that damages entities
	TileLavaFlow
	// TilePit represents a pit or chasm that blocks movement
	TilePit
	// TileRampUp represents a ramp going up to a higher layer
	TileRampUp
	// TileRampDown represents a ramp going down to a lower layer
	TileRampDown
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
	case TileWaterShallow:
		return "shallow_water"
	case TileWaterDeep:
		return "deep_water"
	case TileTree:
		return "tree"
	case TileStairsUp:
		return "stairs_up"
	case TileStairsDown:
		return "stairs_down"
	case TileTrapDoor:
		return "trap_door"
	case TileSecretDoor:
		return "secret_door"
	case TileBridge:
		return "bridge"
	case TileStructure:
		return "structure"
	case TileWallNE:
		return "wall_ne"
	case TileWallNW:
		return "wall_nw"
	case TileWallSE:
		return "wall_se"
	case TileWallSW:
		return "wall_sw"
	case TilePlatform:
		return "platform"
	case TileRamp:
		return "ramp"
	case TileLavaFlow:
		return "lava_flow"
	case TilePit:
		return "pit"
	case TileRampUp:
		return "ramp_up"
	case TileRampDown:
		return "ramp_down"
	default:
		return "unknown"
	}
}

// IsWalkableTile returns true if this tile type is walkable.
func (t TileType) IsWalkableTile() bool {
	return t == TileFloor || t == TileDoor || t == TileCorridor ||
		t == TileWaterShallow || t == TileStairsUp || t == TileStairsDown ||
		t == TileTrapDoor || t == TileSecretDoor || t == TileBridge ||
		t == TilePlatform || t == TileRamp || t == TileRampUp || t == TileRampDown
}

// IsTransparent returns true if this tile type doesn't block vision.
func (t TileType) IsTransparent() bool {
	// Most tiles are transparent except walls (including diagonal walls) and structures
	return t != TileWall && t != TileStructure && t != TileTree &&
		t != TileSecretDoor && t != TileWallNE && t != TileWallNW &&
		t != TileWallSE && t != TileWallSW
}

// MovementCost returns the movement cost multiplier for this tile type.
// 1.0 is normal speed, higher values slow movement, -1 means impassable.
func (t TileType) MovementCost() float64 {
	switch t {
	case TileFloor, TileDoor, TileCorridor, TileBridge, TileStairsUp, TileStairsDown, TilePlatform:
		return 1.0
	case TileRamp, TileRampUp, TileRampDown:
		return 1.2 // Slightly slower on ramps
	case TileWaterShallow:
		return 2.0 // Half speed in shallow water
	case TileLavaFlow:
		return 3.0 // Very slow movement through lava (with damage)
	case TileTrapDoor:
		return 1.5 // Slightly slower on trap doors
	case TileSecretDoor:
		return 1.0 // Once discovered, normal speed
	case TileWall, TileWallNE, TileWallNW, TileWallSE, TileWallSW:
		return -1 // All walls are impassable
	case TileWaterDeep, TileTree, TileStructure, TilePit:
		return -1 // Impassable terrain
	default:
		return -1 // Unknown tiles are impassable
	}
}

// IsDiagonalWall returns true if this tile is a diagonal wall.
func (t TileType) IsDiagonalWall() bool {
	return t == TileWallNE || t == TileWallNW || t == TileWallSE || t == TileWallSW
}

// IsWall returns true if this tile is any type of wall (axis-aligned or diagonal).
func (t TileType) IsWall() bool {
	return t == TileWall || t.IsDiagonalWall()
}

// Layer represents the vertical layer of a terrain tile.
type Layer int

const (
	// LayerGround is the base layer (default for most tiles)
	LayerGround Layer = iota
	// LayerWater is the water/lava layer (below ground)
	LayerWater
	// LayerPlatform is the elevated platform layer (above ground)
	LayerPlatform
)

// GetLayer returns the layer this tile type belongs to.
func (t TileType) GetLayer() Layer {
	switch t {
	case TileWaterShallow, TileWaterDeep, TileLavaFlow, TilePit:
		return LayerWater
	case TilePlatform:
		return LayerPlatform
	case TileBridge:
		return LayerPlatform // Bridges are elevated over water
	default:
		return LayerGround // Most tiles are at ground level
	}
}

// CanTransitionToLayer returns true if an entity can move from this tile type
// to a tile of the target layer (e.g., via ramps or stairs).
func (t TileType) CanTransitionToLayer(target Layer) bool {
	// Ramps allow transitioning between layers
	if t == TileRamp || t == TileRampUp || t == TileRampDown {
		return true
	}
	// Stairs allow vertical movement
	if t == TileStairsUp || t == TileStairsDown {
		return true
	}
	// Same layer transitions are always allowed
	return t.GetLayer() == target
}

// Tile represents a single position in the terrain grid with layer information.
type Tile struct {
	Type  TileType
	X, Y  int
	Layer Layer // Vertical layer for multi-layer terrain
}

// RoomType represents the special purpose or theme of a room.
type RoomType int

const (
	// RoomNormal represents a standard empty room
	RoomNormal RoomType = iota
	// RoomTreasure represents a room with valuable loot
	RoomTreasure
	// RoomBoss represents a room with a boss enemy
	RoomBoss
	// RoomTrap represents a room with hazards or traps
	RoomTrap
	// RoomSpawn represents the player starting room
	RoomSpawn
	// RoomExit represents the dungeon exit/stairs
	RoomExit
)

// String returns the string representation of a room type.
func (rt RoomType) String() string {
	switch rt {
	case RoomNormal:
		return "normal"
	case RoomTreasure:
		return "treasure"
	case RoomBoss:
		return "boss"
	case RoomTrap:
		return "trap"
	case RoomSpawn:
		return "spawn"
	case RoomExit:
		return "exit"
	default:
		return "unknown"
	}
}

// Room represents a rectangular area in the dungeon.
type Room struct {
	X, Y          int // Top-left corner
	Width, Height int
	Type          RoomType // Special room type
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
	Width      int
	Height     int
	Tiles      [][]TileType
	Rooms      []*Room
	Seed       int64
	Level      int     // Dungeon level (0 = first level)
	StairsUp   []Point // Positions of upward stairs
	StairsDown []Point // Positions of downward stairs
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
		Width:      width,
		Height:     height,
		Tiles:      tiles,
		Rooms:      make([]*Room, 0),
		Seed:       seed,
		Level:      0,
		StairsUp:   make([]Point, 0),
		StairsDown: make([]Point, 0),
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
	return tile.IsWalkableTile()
}

// AddStairs adds stairs at the specified position and updates the stairs list.
func (t *Terrain) AddStairs(x, y int, up bool) {
	if !t.IsInBounds(x, y) {
		return
	}

	point := Point{X: x, Y: y}
	if up {
		t.SetTile(x, y, TileStairsUp)
		// Check if point already exists
		for _, p := range t.StairsUp {
			if p.Equals(point) {
				return
			}
		}
		t.StairsUp = append(t.StairsUp, point)
	} else {
		t.SetTile(x, y, TileStairsDown)
		// Check if point already exists
		for _, p := range t.StairsDown {
			if p.Equals(point) {
				return
			}
		}
		t.StairsDown = append(t.StairsDown, point)
	}
}

// IsInBounds checks if the given coordinates are within the terrain bounds.
func (t *Terrain) IsInBounds(x, y int) bool {
	return x >= 0 && x < t.Width && y >= 0 && y < t.Height
}

// ValidateStairPlacement checks that all stairs are placed in valid, walkable locations.
func (t *Terrain) ValidateStairPlacement() error {
	// Check stairs up
	for _, p := range t.StairsUp {
		tile := t.GetTile(p.X, p.Y)
		if tile != TileStairsUp {
			return fmt.Errorf("stairs up at (%d, %d) not placed correctly (tile is %s)", p.X, p.Y, tile.String())
		}
		// Check that at least one adjacent tile is walkable (for accessibility)
		accessible := false
		for _, neighbor := range p.Neighbors() {
			if t.IsInBounds(neighbor.X, neighbor.Y) {
				neighborTile := t.GetTile(neighbor.X, neighbor.Y)
				if neighborTile.IsWalkableTile() && neighborTile != TileStairsUp && neighborTile != TileStairsDown {
					accessible = true
					break
				}
			}
		}
		if !accessible {
			return fmt.Errorf("stairs up at (%d, %d) is not accessible from walkable tiles", p.X, p.Y)
		}
	}

	// Check stairs down
	for _, p := range t.StairsDown {
		tile := t.GetTile(p.X, p.Y)
		if tile != TileStairsDown {
			return fmt.Errorf("stairs down at (%d, %d) not placed correctly (tile is %s)", p.X, p.Y, tile.String())
		}
		// Check that at least one adjacent tile is walkable (for accessibility)
		accessible := false
		for _, neighbor := range p.Neighbors() {
			if t.IsInBounds(neighbor.X, neighbor.Y) {
				neighborTile := t.GetTile(neighbor.X, neighbor.Y)
				if neighborTile.IsWalkableTile() && neighborTile != TileStairsUp && neighborTile != TileStairsDown {
					accessible = true
					break
				}
			}
		}
		if !accessible {
			return fmt.Errorf("stairs down at (%d, %d) is not accessible from walkable tiles", p.X, p.Y)
		}
	}

	return nil
}
