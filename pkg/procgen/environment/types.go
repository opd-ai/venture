// Package environment provides environmental object type definitions.
package environment

import (
	"fmt"
	"image"
)

// ObjectType defines the category of environmental object.
type ObjectType int

const (
	// ObjectFurniture represents interactive furniture objects
	ObjectFurniture ObjectType = iota
	// ObjectDecoration represents non-interactive decorative objects
	ObjectDecoration
	// ObjectObstacle represents blocking but non-harmful objects
	ObjectObstacle
	// ObjectHazard represents harmful environmental hazards
	ObjectHazard
)

// String returns the string representation of an object type.
func (o ObjectType) String() string {
	switch o {
	case ObjectFurniture:
		return "Furniture"
	case ObjectDecoration:
		return "Decoration"
	case ObjectObstacle:
		return "Obstacle"
	case ObjectHazard:
		return "Hazard"
	default:
		return "Unknown"
	}
}

// SubType defines specific object subtypes within each category.
type SubType int

const (
	// SubTypeTable represents a table furniture item (furniture subtypes: 0-19).
	SubTypeTable SubType = iota
	// SubTypeChair represents a chair furniture item.
	SubTypeChair
	// SubTypeBed represents a bed furniture item.
	SubTypeBed
	// SubTypeShelf represents a shelf furniture item.
	SubTypeShelf
	// SubTypeChest represents a chest furniture item.
	SubTypeChest
	// SubTypeDesk represents a desk furniture item.
	SubTypeDesk
	// SubTypeBench represents a bench furniture item.
	SubTypeBench
	// SubTypeCabinet represents a cabinet furniture item.
	SubTypeCabinet
)

const (
	// SubTypePlant represents a plant decoration (decoration subtypes: 20-39).
	SubTypePlant SubType = 20 + iota
	// SubTypeStatue represents a statue decoration.
	SubTypeStatue
	// SubTypePainting represents a painting decoration.
	SubTypePainting
	// SubTypeBanner represents a banner decoration.
	SubTypeBanner
	// SubTypeTorch represents a torch decoration.
	SubTypeTorch
	// SubTypeCandlestick represents a candlestick decoration.
	SubTypeCandlestick
	// SubTypeVase represents a vase decoration.
	SubTypeVase
	// SubTypeTapestry represents a tapestry decoration.
	SubTypeTapestry
	// SubTypeCrystal represents a crystal decoration.
	SubTypeCrystal
	// SubTypeBook represents a book decoration.
	SubTypeBook
)

const (
	// SubTypeBarrel represents a barrel obstacle (obstacle subtypes: 40-59).
	SubTypeBarrel SubType = 40 + iota
	// SubTypeCrate represents a crate obstacle.
	SubTypeCrate
	// SubTypeRubble represents rubble obstacle.
	SubTypeRubble
	// SubTypePillar represents a pillar obstacle.
	SubTypePillar
	// SubTypeBoulder represents a boulder obstacle.
	SubTypeBoulder
	// SubTypeDebris represents debris obstacle.
	SubTypeDebris
	// SubTypeWreckage represents wreckage obstacle.
	SubTypeWreckage
	// SubTypeColumn represents a column obstacle.
	SubTypeColumn
)

const (
	// SubTypeSpikes represents spikes hazard (hazard subtypes: 60-79).
	SubTypeSpikes SubType = 60 + iota
	// SubTypeFirePit represents a fire pit hazard.
	SubTypeFirePit
	// SubTypeAcidPool represents an acid pool hazard.
	SubTypeAcidPool
	// SubTypeBearTrap represents a bear trap hazard.
	SubTypeBearTrap
	// SubTypePoisonGas represents poison gas hazard.
	SubTypePoisonGas
	// SubTypeLavaPit represents a lava pit hazard.
	SubTypeLavaPit
	// SubTypeElectricField represents an electric field hazard.
	SubTypeElectricField
	// SubTypeIceField represents an ice field hazard.
	SubTypeIceField
)

// String returns the string representation of a subtype.
func (s SubType) String() string {
	switch s {
	// Furniture
	case SubTypeTable:
		return "Table"
	case SubTypeChair:
		return "Chair"
	case SubTypeBed:
		return "Bed"
	case SubTypeShelf:
		return "Shelf"
	case SubTypeChest:
		return "Chest"
	case SubTypeDesk:
		return "Desk"
	case SubTypeBench:
		return "Bench"
	case SubTypeCabinet:
		return "Cabinet"

	// Decorations
	case SubTypePlant:
		return "Plant"
	case SubTypeStatue:
		return "Statue"
	case SubTypePainting:
		return "Painting"
	case SubTypeBanner:
		return "Banner"
	case SubTypeTorch:
		return "Torch"
	case SubTypeCandlestick:
		return "Candlestick"
	case SubTypeVase:
		return "Vase"
	case SubTypeTapestry:
		return "Tapestry"
	case SubTypeCrystal:
		return "Crystal"
	case SubTypeBook:
		return "Book"

	// Obstacles
	case SubTypeBarrel:
		return "Barrel"
	case SubTypeCrate:
		return "Crate"
	case SubTypeRubble:
		return "Rubble"
	case SubTypePillar:
		return "Pillar"
	case SubTypeBoulder:
		return "Boulder"
	case SubTypeDebris:
		return "Debris"
	case SubTypeWreckage:
		return "Wreckage"
	case SubTypeColumn:
		return "Column"

	// Hazards
	case SubTypeSpikes:
		return "Spikes"
	case SubTypeFirePit:
		return "FirePit"
	case SubTypeAcidPool:
		return "AcidPool"
	case SubTypeBearTrap:
		return "BearTrap"
	case SubTypePoisonGas:
		return "PoisonGas"
	case SubTypeLavaPit:
		return "LavaPit"
	case SubTypeElectricField:
		return "ElectricField"
	case SubTypeIceField:
		return "IceField"

	default:
		return "Unknown"
	}
}

// GetObjectType returns the category for a given subtype.
func (s SubType) GetObjectType() ObjectType {
	switch {
	case s >= 0 && s < 20:
		return ObjectFurniture
	case s >= 20 && s < 40:
		return ObjectDecoration
	case s >= 40 && s < 60:
		return ObjectObstacle
	case s >= 60 && s < 80:
		return ObjectHazard
	default:
		return ObjectFurniture
	}
}

// EnvironmentalObject represents a generated environmental object.
type EnvironmentalObject struct {
	// Type and subtype
	Type    ObjectType
	SubType SubType

	// Visual representation
	Sprite *image.RGBA
	Width  int
	Height int

	// Gameplay properties
	Collidable   bool
	Interactable bool
	Harmful      bool
	Damage       int // Damage per tick if harmful

	// Genre and seed for reproduction
	GenreID string
	Seed    int64

	// Descriptive name
	Name string
}

// Config contains parameters for object generation.
type Config struct {
	// SubType to generate
	SubType SubType

	// Size in pixels
	Width  int
	Height int

	// Genre for styling
	GenreID string

	// Seed for deterministic generation
	Seed int64

	// Custom parameters
	Custom map[string]interface{}
}

// DefaultConfig returns default object configuration.
func DefaultConfig() Config {
	return Config{
		SubType: SubTypeTable,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    0,
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
	if c.GenreID == "" {
		return fmt.Errorf("genreID cannot be empty")
	}
	return nil
}

// GetProperties returns default properties for a subtype.
func GetProperties(subType SubType) (collidable, interactable, harmful bool, damage int) {
	objectType := subType.GetObjectType()

	switch objectType {
	case ObjectFurniture:
		// Furniture is collidable and interactable but not harmful
		return true, true, false, 0

	case ObjectDecoration:
		// Decorations vary
		switch subType {
		case SubTypeTorch, SubTypeCandlestick:
			// Light sources are not collidable
			return false, true, false, 0
		default:
			// Most decorations are not collidable or interactable
			return false, false, false, 0
		}

	case ObjectObstacle:
		// Obstacles are collidable but not interactable or harmful
		return true, false, false, 0

	case ObjectHazard:
		// Hazards are harmful, some are collidable
		switch subType {
		case SubTypePoisonGas, SubTypeElectricField, SubTypeIceField:
			// Area hazards are not collidable
			return false, false, true, 5
		default:
			// Physical hazards are collidable
			return true, false, true, 10
		}

	default:
		return true, false, false, 0
	}
}
