// Package tiles provides utility functions for tile rendering.
// This file contains shared helper functions used across tile generation systems.
// Code relocated from: transitions.go, phase11_rendering.go
package tiles

import (
	"image/color"
	"math"
)

// smoothstep provides smooth Hermite interpolation between 0 and 1.
// Originally from: transitions.go
func smoothstep(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// lerp performs linear interpolation between a and b.
// Originally from: transitions.go
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// blendColors blends two colors with the given blend factor (0.0-1.0).
// blend=0 returns c1, blend=1 returns c2.
// Originally from: transitions.go
func blendColors(c1, c2 color.Color, blend float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return color.RGBA{
		R: uint8(lerp(float64(r1>>8), float64(r2>>8), blend)),
		G: uint8(lerp(float64(g1>>8), float64(g2>>8), blend)),
		B: uint8(lerp(float64(b1>>8), float64(b2>>8), blend)),
		A: uint8(lerp(float64(a1>>8), float64(a2>>8), blend)),
	}
}

// NOTE: Go 1.21+ provides built-in min() and max() functions for comparable types.
// Previous local definitions were removed in 2026-02-23 cleanup to use builtins.
