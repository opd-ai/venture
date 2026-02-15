// Package engine provides the MaterialSheenComponent for per-entity equipment
// specular highlight state. The render system uses these values to draw animated
// material-appropriate sheen on equipped items.
package engine

// MaterialSheenComponent stores computed specular highlight parameters derived
// from the material visual properties of equipped items. This is a transient
// visual component—not persisted in saves.
type MaterialSheenComponent struct {
	// Aggregate sheen intensity across all equipped items (0.0–1.0)
	SheenIntensity float64

	// Aggregate reflectivity for highlight brightness (0.0–1.0)
	Reflectivity float64

	// Aggregate roughness dampening (higher = softer highlights) (0.0–1.0)
	Roughness float64

	// Highlight color tint RGB (derived from genre and dominant material)
	ColorR float64
	ColorG float64
	ColorB float64

	// Animated highlight phase (0.0–2π, advances each frame)
	Phase float64

	// Pulse speed in radians per second
	PulseSpeed float64

	// Highlight size in pixels (proportional to entity sprite)
	HighlightSize float64

	// Whether the sheen is currently visible
	Enabled bool

	// Dominant material name for debugging/rendering hints
	DominantMaterial string
}

// Type returns the component type identifier.
func (m *MaterialSheenComponent) Type() string {
	return "material_sheen"
}

// NewMaterialSheenComponent creates a material sheen with sensible defaults.
func NewMaterialSheenComponent() *MaterialSheenComponent {
	return &MaterialSheenComponent{
		SheenIntensity: 0.0,
		Reflectivity:   0.0,
		Roughness:      0.5,
		ColorR:         1.0,
		ColorG:         1.0,
		ColorB:         1.0,
		Phase:          0.0,
		PulseSpeed:     1.5,
		HighlightSize:  4.0,
		Enabled:        false,
	}
}
