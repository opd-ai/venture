package engine

// AttackTelegraphComponent stores per-entity visual telegraph state for
// incoming attacks. The render system uses these values to draw a genre-colored
// warning glow that ramps up as the entity's attack cooldown nears ready.
type AttackTelegraphComponent struct {
	// Intensity of the telegraph glow (0.0 = off, 1.0 = maximum warning)
	Intensity float64

	// Genre-themed glow color RGB (0.0-1.0 each)
	ColorR float64
	ColorG float64
	ColorB float64

	// Glow radius in pixels (scales with entity size)
	Radius float64

	// Whether the telegraph is currently active
	Active bool
}

// Type returns the component type identifier.
func (a *AttackTelegraphComponent) Type() string {
	return "attack_telegraph"
}

// NewAttackTelegraphComponent creates a telegraph component with default values.
func NewAttackTelegraphComponent() *AttackTelegraphComponent {
	return &AttackTelegraphComponent{
		Intensity: 0.0,
		ColorR:    0.9,
		ColorG:    0.2,
		ColorB:    0.1,
		Radius:    16.0,
		Active:    false,
	}
}
