// Package patterns provides pattern type definitions and configuration.
package patterns

import (
	"image/color"
)

// PatternType represents different pattern primitives.
type PatternType int

const (
	// PatternStripes represents parallel lines pattern
	PatternStripes PatternType = iota
	// PatternDots represents a dot/circle grid pattern
	PatternDots
	// PatternGradient represents a smooth color gradient
	PatternGradient
	// PatternNoise represents pseudo-random noise
	PatternNoise
	// PatternCheckerboard represents a checkerboard/chess pattern
	PatternCheckerboard
	// PatternCircles represents concentric or scattered circles
	PatternCircles
)

// String returns the string representation of a pattern type.
func (p PatternType) String() string {
	switch p {
	case PatternStripes:
		return "stripes"
	case PatternDots:
		return "dots"
	case PatternGradient:
		return "gradient"
	case PatternNoise:
		return "noise"
	case PatternCheckerboard:
		return "checkerboard"
	case PatternCircles:
		return "circles"
	default:
		return "unknown"
	}
}

// Config holds configuration for pattern generation.
type Config struct {
	Type   PatternType
	Width  int
	Height int
	Seed   int64

	// Pattern-specific parameters
	Frequency float64 // For stripes, dots, checkerboard (spacing between elements)
	Amplitude float64 // For noise intensity, wave amplitude
	Angle     float64 // Rotation angle in degrees (0-360)

	// Color parameters
	Color1 color.Color // Primary/foreground color
	Color2 color.Color // Secondary/background color

	// Blending
	Opacity   float64 // Pattern opacity (0.0-1.0)
	BlendMode string  // "overlay", "multiply", "screen", "add"
}

// DefaultConfig returns a default pattern configuration.
func DefaultConfig() Config {
	return Config{
		Type:      PatternStripes,
		Width:     32,
		Height:    32,
		Seed:      0,
		Frequency: 4.0,
		Amplitude: 0.5,
		Angle:     0,
		Color1:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Color2:    color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Opacity:   0.5,
		BlendMode: "overlay",
	}
}
