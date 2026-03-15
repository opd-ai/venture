// Package quality provides visual quality tier management for Venture.
// This file defines the QualitySettingsComponent for per-entity quality overrides.
package quality

import "encoding/json"

// QualitySettingsComponent is an ECS component that allows per-entity
// quality overrides. This enables specific entities to use different
// quality settings than the global configuration.
type QualitySettingsComponent struct {
	// Override enables this component's settings to override global quality
	Override bool

	// SpriteDetailOverride overrides sprite detail level (0.0-1.0)
	// Only used if Override is true
	SpriteDetailOverride float64

	// EnableAntiAliasingOverride forces anti-aliasing on/off for this entity
	// Only used if Override is true
	EnableAntiAliasingOverride bool

	// ParticleCountMultiplierOverride scales particle emissions
	// Only used if Override is true
	ParticleCountMultiplierOverride float64

	// DisableEffects completely disables visual effects for this entity
	// Useful for background entities or distant objects
	DisableEffects bool
}

// Type returns the component type identifier.
func (q QualitySettingsComponent) Type() string {
	return "quality_settings"
}

// NewQualitySettingsComponent creates a new quality settings component
// with default values (no overrides).
func NewQualitySettingsComponent() QualitySettingsComponent {
	return QualitySettingsComponent{
		Override:                        false,
		SpriteDetailOverride:            1.0,
		EnableAntiAliasingOverride:      true,
		ParticleCountMultiplierOverride: 1.0,
		DisableEffects:                  false,
	}
}

// WithSpriteDetail creates a component that overrides sprite detail.
func WithSpriteDetail(detail float64) QualitySettingsComponent {
	return QualitySettingsComponent{
		Override:                        true,
		SpriteDetailOverride:            detail,
		EnableAntiAliasingOverride:      true,
		ParticleCountMultiplierOverride: 1.0,
		DisableEffects:                  false,
	}
}

// WithParticleMultiplier creates a component that overrides particle count.
func WithParticleMultiplier(multiplier float64) QualitySettingsComponent {
	return QualitySettingsComponent{
		Override:                        true,
		SpriteDetailOverride:            1.0,
		EnableAntiAliasingOverride:      true,
		ParticleCountMultiplierOverride: multiplier,
		DisableEffects:                  false,
	}
}

// WithoutEffects creates a component that disables all effects.
func WithoutEffects() QualitySettingsComponent {
	return QualitySettingsComponent{
		DisableEffects: true,
	}
}

// Serialize encodes the component to JSON bytes for persistence across save/load cycles.
func (q *QualitySettingsComponent) Serialize() ([]byte, error) {
	return json.Marshal(q)
}

// Deserialize decodes JSON bytes into the component.
func (q *QualitySettingsComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, q)
}
