package sprites

import (
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// SpriteType represents different categories of sprites.
type SpriteType int

const (
	// SpriteEntity represents character/monster sprites
	SpriteEntity SpriteType = iota
	// SpriteItem represents item/collectible sprites
	SpriteItem
	// SpriteTile represents terrain tile sprites
	SpriteTile
	// SpriteParticle represents particle effect sprites
	SpriteParticle
	// SpriteUI represents UI element sprites
	SpriteUI
)

// String returns the string representation of a sprite type.
func (s SpriteType) String() string {
	switch s {
	case SpriteEntity:
		return "entity"
	case SpriteItem:
		return "item"
	case SpriteTile:
		return "tile"
	case SpriteParticle:
		return "particle"
	case SpriteUI:
		return "ui"
	default:
		return "unknown"
	}
}

// Config contains parameters for sprite generation.
type Config struct {
	// Type of sprite to generate
	Type SpriteType

	// Width and height in pixels
	Width  int
	Height int

	// Seed for deterministic generation
	Seed int64

	// Palette to use for colors
	Palette *palette.Palette

	// Genre ID for style consistency
	GenreID string

	// Complexity level (0.0-1.0) - affects detail
	Complexity float64

	// Variation index for creating different sprites from same config
	Variation int

	// Custom parameters for specific sprite types
	Custom map[string]interface{}
}

// DefaultConfig returns a default sprite configuration.
func DefaultConfig() Config {
	return Config{
		Type:       SpriteEntity,
		Width:      32,
		Height:     32,
		Seed:       0,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  0,
		Custom:     make(map[string]interface{}),
	}
}

// Layer represents a single layer in a sprite composition.
type Layer struct {
	Shape     shapes.Shape
	OffsetX   int
	OffsetY   int
	ZIndex    int
	Opacity   float64
	BlendMode string
}

// Sprite represents a generated sprite with metadata.
type Sprite struct {
	Config Config
	Layers []Layer
	Width  int
	Height int
}
