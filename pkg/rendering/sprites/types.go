// Package sprites provides sprite type definitions.
// This file defines sprite data structures, animation parameters,
// and rendering state used by the sprite generator.
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

// LayerType represents different sprite layers for composition.
type LayerType int

const (
	// LayerBody is the main body layer
	LayerBody LayerType = iota
	// LayerHead is the head/face layer
	LayerHead
	// LayerLegs is the legs/lower body layer
	LayerLegs
	// LayerWeapon is the weapon layer
	LayerWeapon
	// LayerArmor is the armor/clothing layer
	LayerArmor
	// LayerAccessory is accessories (hat, cape, etc.)
	LayerAccessory
	// LayerEffect is status effects overlay
	LayerEffect
)

// Z-index constants for layer rendering order.
// These define the standard ordering: Legs < Body < Armor < Head < Weapon < Accessory < Effect.
// Higher values render on top of lower values.
const (
	ZIndexLegs      = 5  // Legs render below body
	ZIndexBody      = 10 // Body is the base layer
	ZIndexArmor     = 15 // Armor renders over body
	ZIndexHead      = 20 // Head renders over armor
	ZIndexWeapon    = 25 // Weapon renders over head
	ZIndexAccessory = 30 // Accessories render on top
	ZIndexEffect    = 40 // Effects render above everything
)

// String returns the string representation of a layer type.
func (l LayerType) String() string {
	switch l {
	case LayerBody:
		return "body"
	case LayerHead:
		return "head"
	case LayerLegs:
		return "legs"
	case LayerWeapon:
		return "weapon"
	case LayerArmor:
		return "armor"
	case LayerAccessory:
		return "accessory"
	case LayerEffect:
		return "effect"
	default:
		return "unknown"
	}
}

// LayerConfig defines a single layer in a composite sprite.
type LayerConfig struct {
	// Type of layer
	Type LayerType

	// Z-index for rendering order (higher = drawn on top)
	ZIndex int

	// Offset from base position
	OffsetX, OffsetY float64

	// Scale factor (1.0 = normal size)
	Scale float64

	// Color tint (nil = no tint)
	ColorTint *palette.Palette

	// Visibility flag
	Visible bool

	// Seed for this layer's generation
	Seed int64

	// Shape type for this layer (if applicable)
	ShapeType shapes.ShapeType
}

// CompositeConfig contains parameters for multi-layer sprite composition.
type CompositeConfig struct {
	// Base configuration
	BaseConfig Config

	// Layers to composite (rendered in order of ZIndex)
	Layers []LayerConfig

	// Equipment visuals
	Equipment []EquipmentVisual

	// Status effects to overlay
	StatusEffects []StatusEffect
}

// MaterialType represents the material an equipment item is made from.
type MaterialType int

const (
	// MaterialMetal represents metallic materials (iron, steel, silver, gold)
	MaterialMetal MaterialType = iota
	// MaterialLeather represents leather and hide materials
	MaterialLeather
	// MaterialCloth represents fabric and cloth materials
	MaterialCloth
	// MaterialWood represents wooden materials
	MaterialWood
	// MaterialCrystal represents crystalline materials (gems, glass)
	MaterialCrystal
	// MaterialEnergy represents energy-based or magical materials
	MaterialEnergy
)

// String returns the string representation of a material type.
func (m MaterialType) String() string {
	switch m {
	case MaterialMetal:
		return "metal"
	case MaterialLeather:
		return "leather"
	case MaterialCloth:
		return "cloth"
	case MaterialWood:
		return "wood"
	case MaterialCrystal:
		return "crystal"
	case MaterialEnergy:
		return "energy"
	default:
		return "unknown"
	}
}

// DamageState represents the condition/wear of an equipment item.
type DamageState int

const (
	// DamageStatePristine represents perfect condition (100% durability)
	DamageStatePristine DamageState = iota
	// DamageStateWorn represents light wear (50-99% durability)
	DamageStateWorn
	// DamageStateDamaged represents heavy damage (25-49% durability)
	DamageStateDamaged
	// DamageStateBroken represents severely damaged (0-24% durability)
	DamageStateBroken
)

// String returns the string representation of a damage state.
func (d DamageState) String() string {
	switch d {
	case DamageStatePristine:
		return "pristine"
	case DamageStateWorn:
		return "worn"
	case DamageStateDamaged:
		return "damaged"
	case DamageStateBroken:
		return "broken"
	default:
		return "unknown"
	}
}

// EnchantmentGlow represents visual properties of magical enchantments.
type EnchantmentGlow struct {
	// Active indicates if enchantment glow is enabled
	Active bool

	// Color of the glow effect (e.g., "blue", "red", "purple", "gold")
	Color string

	// Intensity of the glow (0.0-1.0)
	Intensity float64

	// PulseSpeed controls animation speed (0.0 = static, 1.0 = normal)
	PulseSpeed float64

	// ParticleCount for magical particle effects
	ParticleCount int
}

// EquipmentVisual represents visual properties of equipped items.
type EquipmentVisual struct {
	// Slot type (weapon, armor, accessory)
	Slot string

	// Item ID for deterministic generation
	ItemID string

	// Seed for visual generation
	Seed int64

	// Layer to render on
	Layer LayerType

	// Material type for visual styling
	Material MaterialType

	// Damage state for wear/condition visuals
	DamageState DamageState

	// Enchantment glow effects
	Enchantment EnchantmentGlow

	// DetailLevel controls visual complexity (0.0-1.0, typically based on rarity)
	DetailLevel float64

	// Custom visual parameters
	Params map[string]interface{}
}

// StatusEffect represents a visual status effect overlay.
type StatusEffect struct {
	// Effect type (burning, frozen, poisoned, etc.)
	Type string

	// Intensity (0.0-1.0)
	Intensity float64

	// Color for the effect
	Color string

	// Animation speed modifier
	AnimSpeed float64

	// Particle count for effect
	ParticleCount int
}
