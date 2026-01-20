package furniture

import "image/color"

// This file contains all type definitions for the furniture package including:
// - Enums: FurnitureType, MaterialType, RarityTier, Direction
// - Structs: Furniture (generated items), Template (generation blueprints)
// - All enum String() and helper methods

// FurnitureType represents categories of furniture
type FurnitureType int

const (
	TypeSeating FurnitureType = iota
	TypeStorage
	TypeCrafting
	TypeDecoration
	TypeLighting
	TypeBedding
	TypeTable
	TypeUtility
)

// String returns the string representation of FurnitureType
func (f FurnitureType) String() string {
	switch f {
	case TypeSeating:
		return "Seating"
	case TypeStorage:
		return "Storage"
	case TypeCrafting:
		return "Crafting"
	case TypeDecoration:
		return "Decoration"
	case TypeLighting:
		return "Lighting"
	case TypeBedding:
		return "Bedding"
	case TypeTable:
		return "Table"
	case TypeUtility:
		return "Utility"
	default:
		return "Unknown"
	}
}

// MaterialType represents furniture materials
type MaterialType int

const (
	MaterialWood MaterialType = iota
	MaterialMetal
	MaterialStone
	MaterialCrystal
	MaterialFabric
)

// String returns the string representation of MaterialType
func (m MaterialType) String() string {
	switch m {
	case MaterialWood:
		return "Wood"
	case MaterialMetal:
		return "Metal"
	case MaterialStone:
		return "Stone"
	case MaterialCrystal:
		return "Crystal"
	case MaterialFabric:
		return "Fabric"
	default:
		return "Unknown"
	}
}

// RarityTier affects visual detail and functionality
type RarityTier int

const (
	RarityCommon RarityTier = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

// String returns the string representation of RarityTier
func (r RarityTier) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityEpic:
		return "Epic"
	case RarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// DetailMultiplier returns the visual detail multiplier for this rarity
func (r RarityTier) DetailMultiplier() float64 {
	switch r {
	case RarityCommon:
		return 1.0
	case RarityUncommon:
		return 1.2
	case RarityRare:
		return 1.5
	case RarityEpic:
		return 2.0
	case RarityLegendary:
		return 3.0
	default:
		return 1.0
	}
}

// Direction represents furniture rotation
type Direction int

const (
	DirNorth Direction = iota
	DirEast
	DirSouth
	DirWest
	DirNorthEast
	DirSouthEast
	DirSouthWest
	DirNorthWest
)

// String returns the string representation of Direction
func (d Direction) String() string {
	switch d {
	case DirNorth:
		return "North"
	case DirEast:
		return "East"
	case DirSouth:
		return "South"
	case DirWest:
		return "West"
	case DirNorthEast:
		return "NorthEast"
	case DirSouthEast:
		return "SouthEast"
	case DirSouthWest:
		return "SouthWest"
	case DirNorthWest:
		return "NorthWest"
	default:
		return "Unknown"
	}
}

// RotationSteps returns number of 90-degree steps (4 directions) or 45-degree steps (8 directions)
func (d Direction) RotationSteps() int {
	return int(d)
}

// Furniture represents a generated furniture item
type Furniture struct {
	ID          string
	Type        FurnitureType
	SubType     string // "Chair", "Table", "Chest", etc.
	Material    MaterialType
	Rarity      RarityTier
	GenreID     string
	Name        string
	Description string

	// Dimensions in tiles
	Width  float64
	Height float64
	Depth  float64

	// Visual properties
	PrimaryColor   color.RGBA
	SecondaryColor color.RGBA
	DetailLevel    float64 // 0.0-3.0 based on rarity

	// Placement properties
	Direction      Direction
	Walkable       bool
	CollisionWidth float64
	CollisionDepth float64

	// Functionality
	Functional     bool
	Capacity       int     // Storage slots if applicable
	CraftingType   string  // Type of crafting if applicable
	LightIntensity float64 // Light output if applicable
}

// Template defines base properties for furniture generation
type Template struct {
	Type     FurnitureType
	SubType  string
	BaseName string

	// Size ranges (in tiles)
	MinWidth  float64
	MaxWidth  float64
	MinHeight float64
	MaxHeight float64
	MinDepth  float64
	MaxDepth  float64

	// Material compatibility
	AllowedMaterials []MaterialType

	// Properties
	Walkable       bool
	Functional     bool
	BaseCapacity   int     // Storage
	BaseLightLevel float64 // Lighting

	// Visual style
	DetailComplexity float64 // Base complexity, modified by rarity
}
