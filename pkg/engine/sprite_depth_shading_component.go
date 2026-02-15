// Package engine provides the SpriteDepthShadingComponent which stores
// per-entity shading configuration for the depth shading system. This is
// a transient visual component—not persisted in saves.
package engine

// SpriteDepthShadingComponent holds per-entity shading parameters that control
// how the sprite depth shading system renders highlight, edge darkening, and
// ambient occlusion on entity sprites.
type SpriteDepthShadingComponent struct {
	// LightIntensity controls highlight strength (0.0-1.0).
	LightIntensity float64

	// EdgeDarkening controls how much darker edges become (0.0-1.0).
	EdgeDarkening float64

	// AmbientOcclusion controls darkening at overlapping boundaries (0.0-1.0).
	AmbientOcclusion float64

	// DitherStrength adds subtle noise texture (0.0-1.0).
	DitherStrength float64

	// TintR, TintG, TintB is the light color tint derived from genre.
	TintR float64
	TintG float64
	TintB float64

	// Enabled controls whether shading is active for this entity.
	Enabled bool

	// Dirty flags sprite for regeneration when shading params change.
	Dirty bool
}

// Type returns the component type identifier.
func (c *SpriteDepthShadingComponent) Type() string {
	return "sprite_depth_shading"
}

// NewSpriteDepthShadingComponent creates a component with default shading values.
func NewSpriteDepthShadingComponent() *SpriteDepthShadingComponent {
	return &SpriteDepthShadingComponent{
		LightIntensity:   0.35,
		EdgeDarkening:    0.25,
		AmbientOcclusion: 0.15,
		DitherStrength:   0.06,
		TintR:            1.0,
		TintG:            0.98,
		TintB:            0.93,
		Enabled:          true,
		Dirty:            false,
	}
}
