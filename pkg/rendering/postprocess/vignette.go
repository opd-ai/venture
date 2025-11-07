// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
	"math"
)

// ApplyVignette applies a vignette effect that darkens the edges of the image.
func (p *Processor) ApplyVignette(img *image.RGBA) *image.RGBA {
	if !p.config.Vignette.Enabled {
		return img
	}

	config := p.config.Vignette
	if config.Intensity <= 0 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Calculate image center and radius
	centerX := float64(bounds.Min.X+bounds.Max.X) / 2.0
	centerY := float64(bounds.Min.Y+bounds.Max.Y) / 2.0
	maxRadius := math.Sqrt(
		math.Pow(float64(bounds.Dx())/2.0, 2) +
			math.Pow(float64(bounds.Dy())/2.0, 2),
	)

	// Get vignette color components
	vr, vg, vb, _ := config.Color.RGBA()
	vignetteR := float64(vr) / 65535.0
	vignetteG := float64(vg) / 65535.0
	vignetteB := float64(vb) / 65535.0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate distance from center (normalized 0-1)
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx+dy*dy) / maxRadius

			// Calculate vignette strength based on distance
			vignetteStrength := 0.0

			if dist > config.InnerRadius {
				// Map distance to vignette range
				t := (dist - config.InnerRadius) / (config.OuterRadius - config.InnerRadius)
				t = clamp(t, 0.0, 1.0)

				// Apply softness using smoothstep
				if config.Softness > 0 {
					vignetteStrength = smoothstep(0.0, 1.0, t)
				} else {
					vignetteStrength = t
				}

				vignetteStrength *= config.Intensity
			}

			// Get original pixel
			c := img.RGBAAt(x, y)
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0
			a := float64(c.A) / 255.0

			// Blend with vignette color
			r = lerp(r, vignetteR, vignetteStrength)
			g = lerp(g, vignetteG, vignetteStrength)
			b = lerp(b, vignetteB, vignetteStrength)

			result.Set(x, y, color.RGBA{
				R: clampUint8(r * 255.0),
				G: clampUint8(g * 255.0),
				B: clampUint8(b * 255.0),
				A: clampUint8(a * 255.0),
			})
		}
	}

	return result
}

// ApplyRadialGradient applies a radial gradient from center to edges.
// This is similar to vignette but with customizable colors.
func ApplyRadialGradient(img *image.RGBA, centerColor, edgeColor color.Color, intensity float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Calculate image center and radius
	centerX := float64(bounds.Min.X+bounds.Max.X) / 2.0
	centerY := float64(bounds.Min.Y+bounds.Max.Y) / 2.0
	maxRadius := math.Sqrt(
		math.Pow(float64(bounds.Dx())/2.0, 2) +
			math.Pow(float64(bounds.Dy())/2.0, 2),
	)

	// Get color components
	cr, cg, cb, _ := centerColor.RGBA()
	centerR := float64(cr) / 65535.0
	centerG := float64(cg) / 65535.0
	centerB := float64(cb) / 65535.0

	er, eg, eb, _ := edgeColor.RGBA()
	edgeR := float64(er) / 65535.0
	edgeG := float64(eg) / 65535.0
	edgeB := float64(eb) / 65535.0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate distance from center (normalized 0-1)
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx+dy*dy) / maxRadius
			dist = clamp(dist, 0.0, 1.0)

			// Interpolate gradient colors
			gradR := lerp(centerR, edgeR, dist)
			gradG := lerp(centerG, edgeG, dist)
			gradB := lerp(centerB, edgeB, dist)

			// Get original pixel
			c := img.RGBAAt(x, y)
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0
			a := float64(c.A) / 255.0

			// Blend with gradient
			r = lerp(r, gradR, intensity)
			g = lerp(g, gradG, intensity)
			b = lerp(b, gradB, intensity)

			result.Set(x, y, color.RGBA{
				R: clampUint8(r * 255.0),
				G: clampUint8(g * 255.0),
				B: clampUint8(b * 255.0),
				A: clampUint8(a * 255.0),
			})
		}
	}

	return result
}
