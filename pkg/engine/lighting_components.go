// Package engine provides lighting components for the ECS.
// This file defines components for dynamic lighting including point lights,
// ambient light, and light falloff calculations. The lighting system enhances
// visual atmosphere and immersion with genre-appropriate lighting configurations.
//
// Design Philosophy:
// - Components contain only data, no behavior
// - Lighting calculations happen in LightingSystem
// - Integration with existing rendering pipeline via post-processing
// - Performance-conscious with configurable light limits and culling
package engine

import (
	"image/color"
	"math"
)

// LightFalloffType defines how light intensity decreases with distance.
type LightFalloffType int

const (
	// FalloffLinear decreases light intensity linearly with distance
	FalloffLinear LightFalloffType = iota
	// FalloffQuadratic decreases with distance squared (physically accurate)
	FalloffQuadratic
	// FalloffInverseSquare decreases with 1/distance^2 (realistic physics)
	FalloffInverseSquare
	// FalloffConstant maintains constant intensity until radius cutoff
	FalloffConstant
)

// String returns the string representation of falloff type.
func (f LightFalloffType) String() string {
	switch f {
	case FalloffLinear:
		return "linear"
	case FalloffQuadratic:
		return "quadratic"
	case FalloffInverseSquare:
		return "inverse_square"
	case FalloffConstant:
		return "constant"
	default:
		return "unknown"
	}
}

// LightComponent marks an entity as a light source.
// Lights emit colored illumination that affects the visual appearance
// of nearby entities and terrain. Multiple lights can combine additively.
type LightComponent struct {
	// Color of the light (RGB)
	Color color.RGBA

	// Radius is the maximum distance the light reaches (in pixels)
	Radius float64

	// Intensity is the brightness multiplier (1.0 = full brightness)
	Intensity float64

	// Falloff determines how light dims with distance
	Falloff LightFalloffType

	// Enabled allows lights to be toggled on/off without removal
	Enabled bool

	// Flickering enables random intensity variation
	Flickering bool

	// FlickerSpeed controls how fast flicker changes occur (Hz)
	FlickerSpeed float64

	// FlickerAmount controls intensity variation range (0.0-1.0)
	FlickerAmount float64

	// Pulsing enables smooth periodic intensity changes
	Pulsing bool

	// PulseSpeed controls pulse frequency (Hz)
	PulseSpeed float64

	// PulseAmount controls intensity variation range (0.0-1.0)
	PulseAmount float64

	// internalTime tracks animation time for effects (managed by system)
	internalTime float64
}

// Type returns the component type identifier.
func (l *LightComponent) Type() string {
	return "light"
}

// NewLightComponent creates a light with default values.
// radius: maximum light reach (default 200)
// color: light color (default white)
// intensity: brightness multiplier (default 1.0)
func NewLightComponent(radius float64, lightColor color.RGBA, intensity float64) *LightComponent {
	if radius <= 0 {
		radius = 200
	}
	if intensity <= 0 {
		intensity = 1.0
	}

	return &LightComponent{
		Color:         lightColor,
		Radius:        radius,
		Intensity:     intensity,
		Falloff:       FalloffQuadratic,
		Enabled:       true,
		Flickering:    false,
		FlickerSpeed:  2.0,
		FlickerAmount: 0.2,
		Pulsing:       false,
		PulseSpeed:    1.0,
		PulseAmount:   0.3,
		internalTime:  0,
	}
}

// NewTorchLight creates a flickering torch light (warm orange).
func NewTorchLight(radius float64) *LightComponent {
	light := NewLightComponent(radius, color.RGBA{255, 180, 100, 255}, 1.0)
	light.Flickering = true
	light.FlickerSpeed = 3.0
	light.FlickerAmount = 0.15
	return light
}

// NewSpellLight creates a bright spell light with specified color.
func NewSpellLight(radius float64, spellColor color.RGBA) *LightComponent {
	light := NewLightComponent(radius, spellColor, 1.2)
	light.Falloff = FalloffLinear
	light.Pulsing = true
	light.PulseSpeed = 2.0
	light.PulseAmount = 0.2
	return light
}

// NewCrystalLight creates a pulsing magical crystal light.
func NewCrystalLight(radius float64, crystalColor color.RGBA) *LightComponent {
	light := NewLightComponent(radius, crystalColor, 0.8)
	light.Pulsing = true
	light.PulseSpeed = 0.5
	light.PulseAmount = 0.3
	return light
}

// NewSpellBurstLight creates a bright, short-lived spell burst effect.
// This is intended for spell cast effects that rapidly fade.
func NewSpellBurstLight(radius float64, spellColor color.RGBA) *LightComponent {
	light := NewLightComponent(radius, spellColor, 2.0) // Very bright
	light.Falloff = FalloffQuadratic
	light.Pulsing = true
	light.PulseSpeed = 4.0  // Rapid pulsing
	light.PulseAmount = 0.8 // Large variation
	return light
}

// NewDynamicLight creates a light with custom animation settings.
// This provides full control over light animation for special effects.
func NewDynamicLight(radius float64, lightColor color.RGBA, intensity float64,
	flickering bool, flickerSpeed, flickerAmount float64,
	pulsing bool, pulseSpeed, pulseAmount float64,
) *LightComponent {
	light := NewLightComponent(radius, lightColor, intensity)
	light.Flickering = flickering
	light.FlickerSpeed = flickerSpeed
	light.FlickerAmount = flickerAmount
	light.Pulsing = pulsing
	light.PulseSpeed = pulseSpeed
	light.PulseAmount = pulseAmount
	return light
}

// twoPi is 2π used for animation calculations
const twoPi = 6.283185307179586

// GetCurrentIntensity calculates the effective intensity with animations applied.
// This is called by LightingSystem during rendering.
func (l *LightComponent) GetCurrentIntensity() float64 {
	if !l.Enabled {
		return 0
	}

	intensity := l.Intensity

	// Apply flickering effect
	if l.Flickering {
		// Use internal time for pseudo-random flicker
		// Range: [1.0 - FlickerAmount, 1.0]
		flicker := 1.0 - l.FlickerAmount/2.0 + l.FlickerAmount/2.0*l.fastSin(l.internalTime*l.FlickerSpeed*twoPi)
		intensity *= flicker
	}

	// Apply pulsing effect
	if l.Pulsing {
		// Range: [1.0 - PulseAmount, 1.0]
		pulse := 1.0 - l.PulseAmount/2.0 + l.PulseAmount/2.0*l.fastSin(l.internalTime*l.PulseSpeed*twoPi)
		intensity *= pulse
	}

	return intensity
}

// fastSin is a fast sine approximation for animation effects.
// Uses math.Sin for accuracy as approximations have insufficient precision.
func (l *LightComponent) fastSin(x float64) float64 {
	return math.Sin(x)
}

// AmbientLightComponent defines global ambient lighting for a scene.
// Only one ambient light should exist per world/area.
type AmbientLightComponent struct {
	// Color of ambient light (affects all rendered pixels)
	Color color.RGBA

	// Intensity is the ambient brightness (0.0 = pitch black, 1.0 = fully lit)
	Intensity float64
}

// Type returns the component type identifier.
func (a *AmbientLightComponent) Type() string {
	return "ambient_light"
}

// NewAmbientLightComponent creates ambient light with default values.
func NewAmbientLightComponent(ambientColor color.RGBA, intensity float64) *AmbientLightComponent {
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1 {
		intensity = 1
	}

	return &AmbientLightComponent{
		Color:     ambientColor,
		Intensity: intensity,
	}
}

// LightingConfig stores global lighting settings.
// This is not a component but a configuration object.
type LightingConfig struct {
	// Enabled toggles lighting system on/off
	Enabled bool

	// MaxLights is the maximum number of lights to process per frame
	MaxLights int

	// GammaCorrection applies gamma correction for realistic lighting
	GammaCorrection bool

	// Gamma value (typically 2.2 for monitors)
	Gamma float64

	// AmbientIntensity is the default ambient light level (0.0-1.0)
	AmbientIntensity float64

	// AmbientColor is the default ambient light color
	AmbientColor color.RGBA

	// ShadowsEnabled toggles shadow rendering
	ShadowsEnabled bool

	// ShadowOpacity controls default shadow darkness (0.0-1.0)
	ShadowOpacity float64

	// ShadowQuality controls shadow rendering quality (0.5-2.0)
	ShadowQuality float64

	// MaxShadows is the maximum number of shadow casters per frame
	MaxShadows int
}

// NewLightingConfig creates default lighting configuration.
func NewLightingConfig() *LightingConfig {
	return &LightingConfig{
		Enabled:          true,
		MaxLights:        16,
		GammaCorrection:  true,
		Gamma:            2.2,
		AmbientIntensity: 0.7,                            // Increased for better visibility
		AmbientColor:     color.RGBA{200, 200, 210, 255}, // Brighter with slight blue tint
		ShadowsEnabled:   true,
		ShadowOpacity:    0.5,
		ShadowQuality:    1.0,
		MaxShadows:       100,
	}
}

// SetGenrePreset configures lighting for a specific genre.
// This includes ambient light, shadow settings, and atmospheric parameters.
func (c *LightingConfig) SetGenrePreset(genreID string) {
	switch genreID {
	case "fantasy":
		c.AmbientIntensity = 0.75
		c.AmbientColor = color.RGBA{220, 210, 190, 255} // Warm tone, brighter
		c.ShadowOpacity = 0.5                           // Medium shadows
		c.MaxLights = 20                                // More torches/magic lights
	case "scifi", "sci-fi": // Support both canonical and hyphenated form
		c.AmbientIntensity = 0.7
		c.AmbientColor = color.RGBA{190, 210, 230, 255} // Cool blue, brighter
		c.ShadowOpacity = 0.6                           // Sharper shadows
		c.MaxLights = 24                                // Many tech lights
	case "horror":
		c.AmbientIntensity = 0.5                        // Even horror needs to be visible
		c.AmbientColor = color.RGBA{160, 155, 170, 255} // Dark but visible
		c.ShadowOpacity = 0.8                           // Deep, oppressive shadows
		c.MaxLights = 12                                // Limited, scary lighting
	case "cyberpunk":
		c.AmbientIntensity = 0.65
		c.AmbientColor = color.RGBA{190, 170, 210, 255} // Purple tint, brighter
		c.ShadowOpacity = 0.7                           // Strong neon shadows
		c.MaxLights = 28                                // Neon everywhere
	case "postapoc", "post-apocalyptic": // Support both canonical and hyphenated form
		c.AmbientIntensity = 0.7
		c.AmbientColor = color.RGBA{210, 200, 180, 255} // Dusty, harsh, but visible
		c.ShadowOpacity = 0.4                           // Diffuse, dusty shadows
		c.MaxLights = 16                                // Standard
	default:
		c.AmbientIntensity = 0.7
		c.AmbientColor = color.RGBA{200, 200, 210, 255}
		c.ShadowOpacity = 0.5
		c.MaxLights = 16
	}
}
