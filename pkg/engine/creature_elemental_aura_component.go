// Package engine provides the CreatureElementalAuraComponent for storing
// inherent elemental visual effects on creatures. Fire creatures glow with
// flame-colored aura, ice creatures have frost shimmer, poison creatures
// emit sickly green glow, etc. These are persistent visual effects (not
// status-effect-based) that make elemental creature types immediately
// distinguishable at a glance.
package engine

import "github.com/opd-ai/venture/pkg/procgen/magic"

// CreatureElementalAuraComponent stores the procedurally-assigned elemental
// aura parameters for a creature. The render pipeline reads these values
// to apply a colored glow/tint overlay beneath or around the sprite.
type CreatureElementalAuraComponent struct {
	// Element is the inherent elemental affinity of this creature.
	Element magic.ElementType

	// Aura color (RGB 0.0-1.0) derived from element and genre.
	AuraR float64
	AuraG float64
	AuraB float64

	// BaseIntensity controls overall aura brightness (0.0-1.0).
	BaseIntensity float64

	// CurrentIntensity after pulse modulation (0.0-1.0).
	CurrentIntensity float64

	// PulseSpeed in cycles per second (0 = no pulse).
	PulseSpeed float64

	// PulseAmplitude as fraction of BaseIntensity (0.0-0.5).
	PulseAmplitude float64

	// PulsePhase accumulated radians for animation.
	PulsePhase float64

	// AuraRadius in pixels around the sprite (4.0-12.0 typical).
	AuraRadius float64

	// SecondaryColor for gradient/shimmer effects (RGB 0.0-1.0).
	SecondaryR float64
	SecondaryG float64
	SecondaryB float64

	// ParticleEmission controls whether ambient particles spawn (fire sparks, ice crystals).
	ParticleEmission bool

	// ParticleRate particles per second when ParticleEmission is true.
	ParticleRate float64

	// Enabled indicates whether the aura should render.
	Enabled bool
}

// Type returns the component type identifier.
func (c *CreatureElementalAuraComponent) Type() string {
	return "creature_elemental_aura"
}

// NewCreatureElementalAuraComponent creates a component with disabled defaults.
func NewCreatureElementalAuraComponent() *CreatureElementalAuraComponent {
	return &CreatureElementalAuraComponent{
		Element:          magic.ElementNone,
		AuraR:            0.0,
		AuraG:            0.0,
		AuraB:            0.0,
		BaseIntensity:    0.0,
		CurrentIntensity: 0.0,
		PulseSpeed:       0.0,
		PulseAmplitude:   0.0,
		PulsePhase:       0.0,
		AuraRadius:       6.0,
		SecondaryR:       0.0,
		SecondaryG:       0.0,
		SecondaryB:       0.0,
		ParticleEmission: false,
		ParticleRate:     0.0,
		Enabled:          false,
	}
}

// IsElemental returns true if the creature has an elemental affinity.
func (c *CreatureElementalAuraComponent) IsElemental() bool {
	return c.Enabled && c.Element != magic.ElementNone
}

// GetPrimaryColor returns the aura color as (R, G, B) in 0-255 range.
func (c *CreatureElementalAuraComponent) GetPrimaryColor() (uint8, uint8, uint8) {
	return uint8(c.AuraR * 255), uint8(c.AuraG * 255), uint8(c.AuraB * 255)
}

// GetSecondaryColor returns the secondary color as (R, G, B) in 0-255 range.
func (c *CreatureElementalAuraComponent) GetSecondaryColor() (uint8, uint8, uint8) {
	return uint8(c.SecondaryR * 255), uint8(c.SecondaryG * 255), uint8(c.SecondaryB * 255)
}
