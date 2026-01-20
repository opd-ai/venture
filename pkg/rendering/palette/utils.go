// Package palette provides utility functions for color manipulation.
// This file contains shared helper functions used across palette generation.
// Code relocated from: generator.go, timeofday.go
package palette

import "math"

// clamp restricts a value to a given range.
// Originally defined in: generator.go
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// max returns the maximum of two integers.
// Originally defined in: generator.go
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers.
// Originally defined in: generator.go
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// hueToRGB is a helper function for HSL to RGB conversion.
// Originally defined in: generator.go, timeofday.go
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// rgbToHSL converts RGB (0-255) to HSL (H: 0-360, S: 0-1, L: 0-1).
// Originally defined in: timeofday.go
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	l = (max + min) / 2.0

	if max == min {
		h = 0.0
		s = 0.0
	} else {
		delta := max - min

		if l > 0.5 {
			s = delta / (2.0 - max - min)
		} else {
			s = delta / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / delta
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = (bf-rf)/delta + 2.0
		case bf:
			h = (rf-gf)/delta + 4.0
		}

		h *= 60.0
	}

	return h, s, l
}

// hslToRGB converts HSL (H: 0-360, S: 0-1, L: 0-1) to RGB (0-255).
// Originally defined in: timeofday.go
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0.0 {
		// Achromatic (gray)
		val := uint8(l * 255.0)
		return val, val, val
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}

	p := 2.0*l - q

	hk := h / 360.0

	tr := hk + 1.0/3.0
	tg := hk
	tb := hk - 1.0/3.0

	r = uint8(hueToRGB(p, q, tr) * 255.0)
	g = uint8(hueToRGB(p, q, tg) * 255.0)
	b = uint8(hueToRGB(p, q, tb) * 255.0)

	return r, g, b
}
