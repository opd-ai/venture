// Package engine provides the SurfaceTextureComponent which stores per-entity
// surface texture configuration. This is a transient visual component—not
// persisted in saves—that tells the sprite renderer what micro-texture
// (fur, scales, chitin, metal, bone, ooze, feathers, bark) to apply.
package engine

// SurfaceTextureComponent holds per-entity surface texture state.
// Pure data — all logic lives in SurfaceTextureSystem.
type SurfaceTextureComponent struct {
	// TextureType identifies the surface texture (maps to sprites.SurfaceTextureType).
	TextureType int

	// HeadIntensity controls how strongly the texture shows on the head (0.0-1.0).
	HeadIntensity float64
	// TorsoIntensity controls how strongly the texture shows on the torso.
	TorsoIntensity float64
	// LimbIntensity controls how strongly the texture shows on limbs.
	LimbIntensity float64

	// HeadScale controls head texture density.
	HeadScale float64
	// TorsoScale controls torso texture density.
	TorsoScale float64
	// LimbScale controls limb texture density.
	LimbScale float64

	// PrimaryR, G, B, A is the primary texture overlay color.
	PrimaryR, PrimaryG, PrimaryB, PrimaryA uint8
	// SecondaryR, G, B, A is the secondary/highlight texture color.
	SecondaryR, SecondaryG, SecondaryB, SecondaryA uint8

	// GenreID is the genre used when generating this texture set.
	GenreID string

	// Enabled controls whether surface textures are active for this entity.
	Enabled bool

	// Dirty flags sprite for regeneration when texture params change.
	Dirty bool
}

// Type returns the component type identifier.
func (c *SurfaceTextureComponent) Type() string {
	return "surface_texture"
}

// NewSurfaceTextureComponent creates a component with default (no-texture) values.
func NewSurfaceTextureComponent() *SurfaceTextureComponent {
	return &SurfaceTextureComponent{
		TextureType:    0, // TexNone
		HeadIntensity:  0,
		TorsoIntensity: 0,
		LimbIntensity:  0,
		HeadScale:      1.0,
		TorsoScale:     1.0,
		LimbScale:      1.0,
		Enabled:        true,
		Dirty:          false,
	}
}
