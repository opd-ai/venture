// Package engine provides the ClothingPatternComponent which stores per-entity
// clothing pattern configuration. This is a transient visual component—not
// persisted in saves.
package engine

// ClothingPatternComponent holds per-entity clothing pattern state that
// controls what textile patterns are rendered on the entity's garment areas.
type ClothingPatternComponent struct {
	// TorsoPatternType identifies the torso garment pattern (0 = none).
	TorsoPatternType int
	// TorsoPatternR, G, B is the torso pattern overlay color.
	TorsoPatternR, TorsoPatternG, TorsoPatternB uint8
	// TorsoScale controls torso pattern density.
	TorsoScale float64
	// TorsoIntensity controls torso pattern strength (0.0-1.0).
	TorsoIntensity float64

	// ArmPatternType identifies the arm garment pattern.
	ArmPatternType int
	// ArmPatternR, G, B is the arm pattern overlay color.
	ArmPatternR, ArmPatternG, ArmPatternB uint8
	// ArmScale controls arm pattern density.
	ArmScale float64
	// ArmIntensity controls arm pattern strength (0.0-1.0).
	ArmIntensity float64

	// LegPatternType identifies the leg garment pattern.
	LegPatternType int
	// LegPatternR, G, B is the leg pattern overlay color.
	LegPatternR, LegPatternG, LegPatternB uint8
	// LegScale controls leg pattern density.
	LegScale float64
	// LegIntensity controls leg pattern strength (0.0-1.0).
	LegIntensity float64

	// GenreID is the genre used to derive pattern preferences.
	GenreID string

	// Enabled controls whether clothing patterns are active for this entity.
	Enabled bool

	// Dirty flags sprite for regeneration when pattern params change.
	Dirty bool
}

// Type returns the component type identifier.
func (c *ClothingPatternComponent) Type() string {
	return "clothing_pattern"
}

// NewClothingPatternComponent creates a component with default (no-pattern) values.
func NewClothingPatternComponent() *ClothingPatternComponent {
	return &ClothingPatternComponent{
		TorsoScale:     1.0,
		TorsoIntensity: 0.0,
		ArmScale:       1.0,
		ArmIntensity:   0.0,
		LegScale:       1.0,
		LegIntensity:   0.0,
		Enabled:        true,
		Dirty:          false,
	}
}
