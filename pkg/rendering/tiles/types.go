package tiles

import (
	"fmt"
)

// TileType represents different types of tiles that can be rendered.
type TileType int

const (
	// TileFloor represents a walkable floor tile
	TileFloor TileType = iota
	// TileWall represents a solid wall tile
	TileWall
	// TileDoor represents a door tile
	TileDoor
	// TileCorridor represents a corridor tile
	TileCorridor
	// TileWater represents a water tile
	TileWater
	// TileLava represents a lava tile
	TileLava
	// TileTrap represents a trap tile
	TileTrap
	// TileStairs represents stairs (up or down)
	TileStairs
)

// String returns the string representation of a tile type.
func (t TileType) String() string {
	switch t {
	case TileFloor:
		return "floor"
	case TileWall:
		return "wall"
	case TileDoor:
		return "door"
	case TileCorridor:
		return "corridor"
	case TileWater:
		return "water"
	case TileLava:
		return "lava"
	case TileTrap:
		return "trap"
	case TileStairs:
		return "stairs"
	default:
		return "unknown"
	}
}

// Config contains parameters for tile generation.
type Config struct {
	// Type of tile to generate
	Type TileType

	// Width and height in pixels
	Width  int
	Height int

	// GenreID for style selection
	GenreID string

	// Seed for deterministic generation
	Seed int64

	// Variant controls visual variation (0.0 - 1.0)
	Variant float64

	// Custom parameters for specific tile types
	Custom map[string]interface{}
}

// DefaultConfig returns a default tile configuration.
func DefaultConfig() Config {
	return Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    0,
		Variant: 0.5,
		Custom:  make(map[string]interface{}),
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Width <= 0 {
		return fmt.Errorf("width must be positive, got %d", c.Width)
	}
	if c.Height <= 0 {
		return fmt.Errorf("height must be positive, got %d", c.Height)
	}
	if c.Variant < 0.0 || c.Variant > 1.0 {
		return fmt.Errorf("variant must be between 0.0 and 1.0, got %f", c.Variant)
	}
	if c.GenreID == "" {
		return fmt.Errorf("genreID cannot be empty")
	}
	return nil
}

// Pattern represents a visual pattern that can be applied to tiles.
type Pattern int

const (
	// PatternSolid fills with a solid color
	PatternSolid Pattern = iota
	// PatternCheckerboard creates a checkerboard pattern
	PatternCheckerboard
	// PatternDots creates a dot pattern
	PatternDots
	// PatternLines creates parallel lines
	PatternLines
	// PatternBrick creates a brick pattern (for walls)
	PatternBrick
	// PatternGrain creates a wood grain pattern
	PatternGrain
)

// String returns the string representation of a pattern type.
func (p Pattern) String() string {
	switch p {
	case PatternSolid:
		return "solid"
	case PatternCheckerboard:
		return "checkerboard"
	case PatternDots:
		return "dots"
	case PatternLines:
		return "lines"
	case PatternBrick:
		return "brick"
	case PatternGrain:
		return "grain"
	default:
		return "unknown"
	}
}
