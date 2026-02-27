// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
)

// LightType defines the type of light source.
type LightType int

const (
	// TypeAmbient provides uniform lighting across the entire scene
	TypeAmbient LightType = iota
	// TypePoint emits light from a single point in all directions
	TypePoint
	// TypeDirectional emits parallel light rays in a specific direction
	TypeDirectional
)

// String returns the string representation of a light type.
func (t LightType) String() string {
	switch t {
	case TypeAmbient:
		return "Ambient"
	case TypePoint:
		return "Point"
	case TypeDirectional:
		return "Directional"
	default:
		return "Unknown"
	}
}

// FalloffType defines how light intensity decreases with distance.
type FalloffType int

const (
	// FalloffNone means no falloff (constant intensity)
	FalloffNone FalloffType = iota
	// FalloffLinear decreases linearly with distance
	FalloffLinear
	// FalloffQuadratic decreases with the square of distance
	FalloffQuadratic
	// FalloffInverseSquare follows inverse square law (realistic)
	FalloffInverseSquare
)

const (
	// CornerDetectionThreshold is the minimum depth difference to consider a neighbor significant
	// Originally from: ambient_occlusion.go
	CornerDetectionThreshold = 0.1

	// MinNeighborsForCorner is the minimum number of higher neighbors to detect a corner
	// Originally from: ambient_occlusion.go
	MinNeighborsForCorner = 5
)

// String returns the string representation of a falloff type.
func (f FalloffType) String() string {
	switch f {
	case FalloffNone:
		return "None"
	case FalloffLinear:
		return "Linear"
	case FalloffQuadratic:
		return "Quadratic"
	case FalloffInverseSquare:
		return "InverseSquare"
	default:
		return "Unknown"
	}
}

// Light represents a single light source in the scene.
type Light struct {
	// Type of light source
	Type LightType

	// Position in world coordinates (for point lights)
	Position image.Point

	// Direction for directional lights (unit vector)
	Direction image.Point

	// Color of the light
	Color color.Color

	// Intensity multiplier (0.0 to 1.0, can go higher for bright lights)
	Intensity float64

	// Radius of effect in pixels (for point lights)
	Radius float64

	// Falloff type for distance attenuation
	Falloff FalloffType

	// Whether this light is enabled
	Enabled bool
}

// LightingConfig contains configuration for the lighting system.
type LightingConfig struct {
	// AmbientColor is the base ambient light color
	AmbientColor color.Color

	// AmbientIntensity is the ambient light intensity (0.0 to 1.0)
	AmbientIntensity float64

	// MaxLights is the maximum number of lights to process
	MaxLights int

	// GammaCorrection applies gamma correction (typically 2.2)
	GammaCorrection float64

	// EnableShadows is reserved for future shadow casting support and is
	// currently a no-op. Setting this to true has no effect on rendering.
	//
	// Deprecated: This field exists for API forward-compatibility but has no
	// implementation. Shadow casting may be added in a future release via
	// pkg/engine/shadow_system.go. Do not rely on this field for any logic.
	EnableShadows bool

	// BloomConfig configures bloom/glow effects (Phase 17.1)
	BloomConfig BloomConfig

	// AOConfig configures ambient occlusion (Phase 17.1)
	AOConfig EnhancedAOConfig
}

// DefaultConfig returns a default lighting configuration.
func DefaultConfig() LightingConfig {
	return LightingConfig{
		AmbientColor:     color.RGBA{30, 30, 40, 255},
		AmbientIntensity: 0.2,
		MaxLights:        32,
		GammaCorrection:  2.2,
		EnableShadows:    false,
		BloomConfig:      DefaultBloomConfig(),
		AOConfig:         DefaultEnhancedAOConfig(),
	}
}

// LightingResult contains the calculated lighting for a pixel.
type LightingResult struct {
	// FinalColor is the modulated color after lighting
	FinalColor color.Color

	// LightIntensity is the total light intensity (0.0 to 1.0+)
	LightIntensity float64

	// LightColor is the combined light color
	LightColor color.Color
}

// Validate checks if the light configuration is valid.
func (l *Light) Validate() error {
	if l.Intensity < 0 {
		return &ValidationError{Field: "Intensity", Message: "must be non-negative"}
	}
	if l.Radius < 0 {
		return &ValidationError{Field: "Radius", Message: "must be non-negative"}
	}
	return nil
}

// ValidationError represents a lighting configuration validation error.
// It supports error wrapping via the Unwrap method, allowing nested
// error chains for better error context preservation.
type ValidationError struct {
	Field   string
	Message string
	Err     error // Wrapped error (optional)
}

func (e *ValidationError) Error() string {
	msg := "lighting: " + e.Field + " " + e.Message
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

// Unwrap returns the wrapped error, enabling error unwrapping with errors.Is and errors.As.
func (e *ValidationError) Unwrap() error {
	return e.Err
}
