// Package engine provides the GleamSweepComponent for per-entity animated
// specular gleam sweep state. The render pipeline uses this to draw a moving
// highlight band across equipped items, conveying material quality and rarity.
package engine

// GleamSweepComponent stores the state of an animated specular highlight sweep
// across an entity's equipped items. This is a transient visual component—not
// persisted in saves.
type GleamSweepComponent struct {
	// SweepPosition is the normalized position of the gleam band (0.0–1.0).
	// 0.0 = left/bottom edge, 1.0 = right/top edge of the entity sprite.
	SweepPosition float64

	// SweepSpeed controls how fast the gleam moves across the sprite (units/sec).
	SweepSpeed float64

	// SweepWidth is the normalized width of the highlight band (0.0–1.0).
	SweepWidth float64

	// Intensity is the peak brightness of the gleam (0.0–1.0).
	Intensity float64

	// ColorR, ColorG, ColorB are the gleam tint (genre and material derived).
	ColorR float64
	ColorG float64
	ColorB float64

	// CooldownRemaining is the pause between sweep cycles (seconds).
	CooldownRemaining float64

	// CooldownDuration is the full pause duration between sweeps (seconds).
	CooldownDuration float64

	// Active indicates whether the sweep is currently animating.
	Active bool

	// Enabled indicates the entity qualifies for gleam sweeps.
	Enabled bool

	// MaterialHint stores the dominant material name for render-side logic.
	MaterialHint string
}

// Type returns the component type identifier.
func (g *GleamSweepComponent) Type() string {
	return "gleam_sweep"
}

// NewGleamSweepComponent creates a gleam sweep component with sensible defaults.
func NewGleamSweepComponent() *GleamSweepComponent {
	return &GleamSweepComponent{
		SweepPosition:     0.0,
		SweepSpeed:        0.4,
		SweepWidth:        0.15,
		Intensity:         0.0,
		ColorR:            1.0,
		ColorG:            1.0,
		ColorB:            1.0,
		CooldownRemaining: 0.0,
		CooldownDuration:  3.0,
		Active:            false,
		Enabled:           false,
	}
}
