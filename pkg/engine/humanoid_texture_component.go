// Package engine provides the HumanoidTextureComponent which stores per-entity
// humanoid-specific surface texture configuration including skin textures
// (freckled, scarred, weathered, tattooed), clothing fabric textures (linen,
// leather, silk, wool, chainmail, plate), and hair textures (straight, wavy,
// curly, braided). This is a transient visual component—not persisted in saves—
// that tells the sprite renderer what micro-textures to apply to humanoid entities.
package engine

import "image/color"

// HumanoidTextureComponent holds per-entity humanoid texture state.
// Pure data — all logic lives in HumanoidTextureSystem.
type HumanoidTextureComponent struct {
	// SkinTextureType identifies the skin texture (maps to sprites.HumanoidTextureType).
	SkinTextureType int
	// SkinIntensity controls how strongly the skin texture shows (0.0-1.0).
	SkinIntensity float64
	// SkinScale controls skin texture density.
	SkinScale float64
	// SkinPrimaryColor is the primary skin texture overlay color.
	SkinPrimaryColor color.RGBA
	// SkinSecondaryColor is the secondary skin texture color.
	SkinSecondaryColor color.RGBA

	// ClothingTopTextureType identifies the upper body clothing fabric.
	ClothingTopTextureType int
	// ClothingTopIntensity controls upper body fabric texture strength.
	ClothingTopIntensity float64
	// ClothingTopScale controls upper body fabric texture density.
	ClothingTopScale float64
	// ClothingTopPrimaryColor is the primary upper body fabric color.
	ClothingTopPrimaryColor color.RGBA
	// ClothingTopSecondaryColor is the secondary upper body fabric color.
	ClothingTopSecondaryColor color.RGBA

	// ClothingBottomTextureType identifies the lower body clothing fabric.
	ClothingBottomTextureType int
	// ClothingBottomIntensity controls lower body fabric texture strength.
	ClothingBottomIntensity float64
	// ClothingBottomScale controls lower body fabric texture density.
	ClothingBottomScale float64
	// ClothingBottomPrimaryColor is the primary lower body fabric color.
	ClothingBottomPrimaryColor color.RGBA
	// ClothingBottomSecondaryColor is the secondary lower body fabric color.
	ClothingBottomSecondaryColor color.RGBA

	// HairTextureType identifies the hair texture.
	HairTextureType int
	// HairIntensity controls hair texture strength.
	HairIntensity float64
	// HairScale controls hair texture density.
	HairScale float64
	// HairPrimaryColor is the primary hair color.
	HairPrimaryColor color.RGBA
	// HairSecondaryColor is the secondary hair highlight color.
	HairSecondaryColor color.RGBA
	// HairDirection controls directional hair patterns (radians).
	HairDirection float64

	// GenreID is the genre used when generating this texture set.
	GenreID string

	// Enabled controls whether humanoid textures are active for this entity.
	Enabled bool

	// Dirty flags sprite for regeneration when texture params change.
	Dirty bool
}

// Type returns the component type identifier.
func (c *HumanoidTextureComponent) Type() string {
	return "humanoid_texture"
}

// NewHumanoidTextureComponent creates a component with default (no-texture) values.
func NewHumanoidTextureComponent() *HumanoidTextureComponent {
	return &HumanoidTextureComponent{
		SkinTextureType:           0, // SkinSmooth
		SkinIntensity:             0.2,
		SkinScale:                 1.0,
		ClothingTopTextureType:    5, // FabricLinen
		ClothingTopIntensity:      0.25,
		ClothingTopScale:          1.0,
		ClothingBottomTextureType: 5, // FabricLinen
		ClothingBottomIntensity:   0.25,
		ClothingBottomScale:       1.0,
		HairTextureType:           11, // HairStraight
		HairIntensity:             0.35,
		HairScale:                 1.0,
		Enabled:                   true,
		Dirty:                     false,
	}
}
