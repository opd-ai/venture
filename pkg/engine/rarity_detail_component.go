// Package engine provides the RarityDetailComponent for per-entity equipment
// visual detail scaling based on item rarity tiers. The render pipeline uses
// these values to control shape complexity, color vibrancy, and border quality.
package engine

// RarityDetailComponent stores computed visual detail parameters derived from
// the rarity tiers of all equipped items. This is a transient visual
// component—not persisted in saves.
type RarityDetailComponent struct {
	// DetailLevel is the aggregate detail tier (0.0–1.0) across equipped items.
	// Common=0.3, Uncommon=0.4, Rare=0.6, Epic=0.8, Legendary=1.0.
	DetailLevel float64

	// ShapeComplexity controls vertex count / shape intricacy (0.0–1.0).
	// Higher rarity yields more complex silhouettes and ornamental edges.
	ShapeComplexity float64

	// ColorVibrancy scales saturation boost applied to equipment sprites (0.0–1.0).
	ColorVibrancy float64

	// BorderSharpness controls outline crispness (0.0–1.0).
	// Legendary items get sharp, glowing borders; common items get soft edges.
	BorderSharpness float64

	// MaterialFidelity controls how distinctly material properties render (0.0–1.0).
	MaterialFidelity float64

	// HighestRarity is the string name of the highest rarity item equipped.
	HighestRarity string

	// EquippedCount is the number of equipped items used in the calculation.
	EquippedCount int

	// Enabled indicates whether the detail scaling is active.
	Enabled bool
}

// Type returns the component type identifier.
func (r *RarityDetailComponent) Type() string {
	return "rarity_detail"
}

// NewRarityDetailComponent creates a rarity detail component with common-tier defaults.
func NewRarityDetailComponent() *RarityDetailComponent {
	return &RarityDetailComponent{
		DetailLevel:      0.3,
		ShapeComplexity:  0.3,
		ColorVibrancy:    0.3,
		BorderSharpness:  0.3,
		MaterialFidelity: 0.3,
		HighestRarity:    "common",
		EquippedCount:    0,
		Enabled:          false,
	}
}
