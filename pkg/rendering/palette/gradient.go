// Package palette provides gradient generation for color transitions.
// This file implements linear, radial, and angular gradient generation
// for procedural color effects.
// Phase 19.2: Dynamic Palette System - Gradient Generation
package palette

import (
	"image"
	"image/color"
	"math"
)

// GenerateGradient creates a gradient image based on the configuration.
// Width and height must be positive; values <= 0 are treated as 1 to prevent
// division by zero while still generating a valid (minimal) image.
func GenerateGradient(width, height int, config GradientConfig) *image.RGBA {
	// Prevent division by zero - default to 1x1 for invalid dimensions
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	if len(config.Colors) < 2 {
		// Default to black-white gradient if not enough colors
		config.Colors = []color.Color{color.Black, color.White}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Normalize coordinates (0.0-1.0)
			nx := float64(x) / float64(width)
			ny := float64(y) / float64(height)

			// Calculate gradient position (0.0-1.0)
			var t float64
			switch config.Type {
			case GradientLinear:
				t = calculateLinearGradient(nx, ny, config.Angle)
			case GradientRadial:
				t = calculateRadialGradient(nx, ny, config.CenterX, config.CenterY, config.Radius)
			case GradientAngular:
				t = calculateAngularGradient(nx, ny, config.CenterX, config.CenterY)
			case GradientDiamond:
				t = calculateDiamondGradient(nx, ny, config.CenterX, config.CenterY)
			case GradientSpiral:
				t = calculateSpiralGradient(nx, ny, config.CenterX, config.CenterY, config.SpiralRotations)
			case GradientConic:
				t = calculateConicGradient(nx, ny, config.CenterX, config.CenterY, config.Angle)
			default:
				t = nx // Fallback to simple horizontal gradient
			}

			// Apply smoothness (ease in/out)
			t = applySmoothness(t, config.Smoothness)

			// Reverse if needed
			if config.Reverse {
				t = 1.0 - t
			}

			// Interpolate color from palette — returns color.RGBA to avoid interface boxing.
			c := interpolateColors(config.Colors, t)
			img.SetRGBA(x, y, c)
		}
	}

	return img
}

// calculateLinearGradient computes gradient position for linear gradients.
func calculateLinearGradient(x, y, angle float64) float64 {
	// Convert angle to radians
	rad := angle * math.Pi / 180.0
	// Calculate projection along gradient direction
	return x*math.Cos(rad) + y*math.Sin(rad)
}

// calculateRadialGradient computes gradient position for radial gradients.
// Radius values <= 0 are treated as 1.0 to prevent division by zero.
func calculateRadialGradient(x, y, cx, cy, radius float64) float64 {
	dx := x - cx
	dy := y - cy
	dist := math.Sqrt(dx*dx + dy*dy)
	// Prevent division by zero - default to 1.0 for invalid radius
	if radius <= 0 {
		radius = 1.0
	}
	// Normalize by radius and clamp to [0, 1]
	t := dist / radius
	if t > 1.0 {
		t = 1.0
	}
	return t
}

// calculateAngularGradient computes gradient position for angular gradients.
func calculateAngularGradient(x, y, cx, cy float64) float64 {
	dx := x - cx
	dy := y - cy
	// Calculate angle from center (0 to 2π)
	angle := math.Atan2(dy, dx)
	// Normalize to [0, 1]
	t := (angle + math.Pi) / (2 * math.Pi)
	return t
}

// calculateDiamondGradient computes gradient position for diamond gradients.
func calculateDiamondGradient(x, y, cx, cy float64) float64 {
	dx := math.Abs(x - cx)
	dy := math.Abs(y - cy)
	// Manhattan distance
	dist := dx + dy
	// Normalize (max distance is 1.0 for normalized coords)
	t := dist
	if t > 1.0 {
		t = 1.0
	}
	return t
}

// calculateSpiralGradient computes gradient position for spiral gradients.
func calculateSpiralGradient(x, y, cx, cy, rotations float64) float64 {
	dx := x - cx
	dy := y - cy
	dist := math.Sqrt(dx*dx + dy*dy)
	angle := math.Atan2(dy, dx)

	// Combine distance and angle for spiral effect
	t := math.Mod(dist*rotations+(angle/(2*math.Pi)), 1.0)
	if t < 0 {
		t += 1.0
	}
	return t
}

// calculateConicGradient computes gradient position for conic gradients.
func calculateConicGradient(x, y, cx, cy, startAngle float64) float64 {
	dx := x - cx
	dy := y - cy
	angle := math.Atan2(dy, dx)
	// Apply start angle offset
	offset := startAngle * math.Pi / 180.0
	angle -= offset
	// Normalize to [0, 1]
	t := (angle + math.Pi) / (2 * math.Pi)
	if t < 0 {
		t += 1.0
	}
	if t > 1.0 {
		t -= 1.0
	}
	return t
}

// applySmoothness applies easing function to gradient position.
func applySmoothness(t, smoothness float64) float64 {
	if smoothness <= 0.0 {
		return t // No smoothing
	}
	if smoothness >= 1.0 {
		// Full smoothstep
		return t * t * (3.0 - 2.0*t)
	}
	// Linear interpolation between no smoothing and full smoothstep
	smoothed := t * t * (3.0 - 2.0*t)
	return t*(1.0-smoothness) + smoothed*smoothness
}

// interpolateColors interpolates between multiple colors based on position t (0.0-1.0).
// Returns color.RGBA directly (concrete type) to avoid interface boxing and heap escape.
func interpolateColors(colors []color.Color, t float64) color.RGBA {
	if len(colors) == 0 {
		r, g, b, a := color.Black.RGBA()
		return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	}
	if len(colors) == 1 {
		r, g, b, a := colors[0].RGBA()
		return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	}

	// Clamp t to [0, 1]
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// Find which color pair to interpolate between
	n := len(colors) - 1
	pos := t * float64(n)
	idx := int(pos)
	if idx >= n {
		r, g, b, a := colors[n].RGBA()
		return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	}

	// Calculate interpolation factor within this segment
	segmentT := pos - float64(idx)

	// Get the two colors to interpolate
	c1 := colors[idx]
	c2 := colors[idx+1]

	// Interpolate RGB values
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8((float64(r1>>8)*(1-segmentT) + float64(r2>>8)*segmentT))
	g := uint8((float64(g1>>8)*(1-segmentT) + float64(g2>>8)*segmentT))
	b := uint8((float64(b1>>8)*(1-segmentT) + float64(b2>>8)*segmentT))
	a := uint8((float64(a1>>8)*(1-segmentT) + float64(a2>>8)*segmentT))

	return color.RGBA{R: r, G: g, B: b, A: a}
}

// CreateGradientPalette generates a palette from gradient colors.
// This creates a palette suitable for genre theming with gradient-based color selection.
func CreateGradientPalette(colors []color.Color, steps int) *Palette {
	if len(colors) < 2 {
		colors = []color.Color{color.Black, color.White}
	}
	if steps < 12 {
		steps = 12
	}

	palette := &Palette{
		Colors: make([]color.Color, steps),
	}

	// Generate colors along the gradient
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		palette.Colors[i] = interpolateColors(colors, t)
	}

	// Set key palette colors from gradient samples
	palette.Primary = palette.Colors[0]
	palette.Secondary = palette.Colors[steps/3]
	palette.Background = palette.Colors[0]

	// Set text based on background brightness
	r, g, b, _ := palette.Background.RGBA()
	brightness := (float64(r>>8) + float64(g>>8) + float64(b>>8)) / 3.0
	if brightness < 128 {
		palette.Text = color.RGBA{R: 240, G: 240, B: 240, A: 255}
	} else {
		palette.Text = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	}

	palette.Accent1 = palette.Colors[steps/4]
	palette.Accent2 = palette.Colors[steps/2]
	palette.Accent3 = palette.Colors[3*steps/4]

	palette.Highlight1 = palette.Colors[steps-1]
	palette.Highlight2 = palette.Colors[2*steps/3]

	palette.Shadow1 = palette.Colors[steps/6]
	palette.Shadow2 = palette.Colors[steps/5]

	palette.Neutral = palette.Colors[steps/2]

	// Standard UI colors
	palette.Danger = color.RGBA{R: 220, G: 53, B: 69, A: 255}
	palette.Success = color.RGBA{R: 40, G: 167, B: 69, A: 255}
	palette.Warning = color.RGBA{R: 255, G: 193, B: 7, A: 255}
	palette.Info = color.RGBA{R: 23, G: 162, B: 184, A: 255}

	return palette
}
