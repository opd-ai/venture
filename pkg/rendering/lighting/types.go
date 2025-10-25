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

	// EnableShadows enables shadow casting (not implemented yet)
	EnableShadows bool
}

// DefaultConfig returns a default lighting configuration.
func DefaultConfig() LightingConfig {
	return LightingConfig{
		AmbientColor:     color.RGBA{30, 30, 40, 255},
		AmbientIntensity: 0.2,
		MaxLights:        32,
		GammaCorrection:  2.2,
		EnableShadows:    false,
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
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "lighting: " + e.Field + " " + e.Message
}
