// Package postprocess provides post-processing effects for rendered scenes.
// This file contains internal constants used by various effects.
package postprocess

// Chromatic aberration constants
// Originally from: chromatic_aberration.go
const (
	// chromaticAberrationScale is the base scaling factor for chromatic aberration effect
	chromaticAberrationScale = 5.0

	// prismaticAberrationScale is the scaling factor for prismatic aberration effect
	prismaticAberrationScale = 8.0
)

// Depth blur constants
// Originally from: depth_blur.go
const (
	// maxBlurRadiusPixels is the maximum blur radius in pixels for depth blur
	maxBlurRadiusPixels = 10.0
)
