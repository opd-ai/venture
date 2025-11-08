// Package environment provides procedural generation of environmental objects.
package environment

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// VisualVariation represents visual modifiers for an object.
// Phase 20.1: Adds rotation, scale, and color variations to decorations.
type VisualVariation struct {
	// Rotation in degrees (0-360)
	Rotation float64
	// Scale multiplier (0.5-2.0 typical range)
	Scale float64
	// ColorShift adjusts hue (-180 to 180 degrees)
	ColorShift float64
	// Brightness adjustment (-0.3 to 0.3)
	Brightness float64
	// FlipHorizontal flips the sprite horizontally
	FlipHorizontal bool
	// FlipVertical flips the sprite vertically
	FlipVertical bool
}

// DefaultVariation returns neutral variation (no changes).
func DefaultVariation() VisualVariation {
	return VisualVariation{
		Rotation:       0,
		Scale:          1.0,
		ColorShift:     0,
		Brightness:     0,
		FlipHorizontal: false,
		FlipVertical:   false,
	}
}

// GenerateVariation creates a random variation based on subtype.
func GenerateVariation(subType SubType, rng *rand.Rand) VisualVariation {
	v := DefaultVariation()

	switch subType.GetObjectType() {
	case ObjectDecoration:
		// Decorations get more variation
		v.Rotation = rng.Float64() * 360.0
		v.Scale = 0.8 + rng.Float64()*0.4           // 0.8-1.2
		v.ColorShift = (rng.Float64() - 0.5) * 60.0 // -30 to +30 degrees
		v.Brightness = (rng.Float64() - 0.5) * 0.4  // -0.2 to +0.2
		v.FlipHorizontal = rng.Float64() < 0.3

		// Floor decorations don't flip vertically
		switch subType {
		case SubTypeBloodstain, SubTypeGrass, SubTypeMushroom, SubTypeMoss:
			v.FlipVertical = false
		default:
			v.FlipVertical = rng.Float64() < 0.2
		}

	case ObjectFurniture:
		// Furniture gets rotation but minimal other variation
		v.Rotation = float64(rng.Intn(4)) * 90.0 // 0, 90, 180, 270
		v.Scale = 0.9 + rng.Float64()*0.2        // 0.9-1.1
		v.ColorShift = (rng.Float64() - 0.5) * 20.0
		v.FlipHorizontal = rng.Float64() < 0.5

	case ObjectObstacle:
		// Obstacles get random rotation and slight scale variation
		v.Rotation = rng.Float64() * 360.0
		v.Scale = 0.85 + rng.Float64()*0.3 // 0.85-1.15
		v.ColorShift = (rng.Float64() - 0.5) * 40.0
		v.Brightness = (rng.Float64() - 0.5) * 0.3

	case ObjectHazard:
		// Hazards minimal variation to maintain recognition
		v.Rotation = float64(rng.Intn(8)) * 45.0 // 0, 45, 90, ...
		v.Scale = 0.95 + rng.Float64()*0.1       // 0.95-1.05
	}

	return v
}

// ApplyVariation modifies an image according to the variation settings.
func ApplyVariation(img *image.RGBA, v VisualVariation) *image.RGBA {
	if v.Rotation == 0 && v.Scale == 1.0 && v.ColorShift == 0 &&
		v.Brightness == 0 && !v.FlipHorizontal && !v.FlipVertical {
		return img // No changes needed
	}

	// Apply color adjustments first
	if v.ColorShift != 0 || v.Brightness != 0 {
		img = applyColorAdjustments(img, v.ColorShift, v.Brightness)
	}

	// Apply flips
	if v.FlipHorizontal || v.FlipVertical {
		img = applyFlips(img, v.FlipHorizontal, v.FlipVertical)
	}

	// Apply rotation and scale (combined for efficiency)
	if v.Rotation != 0 || v.Scale != 1.0 {
		img = applyTransform(img, v.Rotation, v.Scale)
	}

	return img
}

// applyColorAdjustments adjusts hue and brightness of the image.
func applyColorAdjustments(img *image.RGBA, hueShift, brightnessAdjust float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.A == 0 {
				continue // Skip transparent pixels
			}

			// Convert to HSL
			h, s, l := rgbToHSL(c.R, c.G, c.B)

			// Apply hue shift
			h = math.Mod(h+hueShift, 360.0)
			if h < 0 {
				h += 360.0
			}

			// Apply brightness
			l = clamp(l+brightnessAdjust, 0, 1)

			// Convert back to RGB
			r, g, b := hslToRGB(h, s, l)
			result.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: c.A})
		}
	}

	return result
}

// applyFlips flips the image horizontally and/or vertically.
func applyFlips(img *image.RGBA, horizontal, vertical bool) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	result := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x
			srcY := y

			if horizontal {
				srcX = width - 1 - x
			}
			if vertical {
				srcY = height - 1 - y
			}

			result.Set(x, y, img.At(srcX, srcY))
		}
	}

	return result
}

// applyTransform applies rotation and scale to the image.
func applyTransform(img *image.RGBA, rotation, scale float64) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions after scaling
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	if newWidth <= 0 || newHeight <= 0 {
		return img
	}

	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Convert rotation to radians
	rad := rotation * math.Pi / 180.0
	cosTheta := math.Cos(rad)
	sinTheta := math.Sin(rad)

	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0
	newCenterX := float64(newWidth) / 2.0
	newCenterY := float64(newHeight) / 2.0

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Transform destination to source coordinates
			dx := float64(x) - newCenterX
			dy := float64(y) - newCenterY

			// Apply inverse scale
			dx /= scale
			dy /= scale

			// Apply inverse rotation
			srcX := dx*cosTheta + dy*sinTheta + centerX
			srcY := -dx*sinTheta + dy*cosTheta + centerY

			// Bilinear interpolation for smooth results
			c := bilinearSample(img, srcX, srcY)
			result.Set(x, y, c)
		}
	}

	return result
}

// bilinearSample samples color from an image using bilinear interpolation.
func bilinearSample(img *image.RGBA, x, y float64) color.Color {
	bounds := img.Bounds()

	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1

	// Check bounds
	if x0 < bounds.Min.X || x1 >= bounds.Max.X || y0 < bounds.Min.Y || y1 >= bounds.Max.Y {
		return color.RGBA{A: 0} // Transparent for out-of-bounds
	}

	// Get four corner colors
	c00 := img.RGBAAt(x0, y0)
	c10 := img.RGBAAt(x1, y0)
	c01 := img.RGBAAt(x0, y1)
	c11 := img.RGBAAt(x1, y1)

	// Interpolation weights
	fx := x - float64(x0)
	fy := y - float64(y0)

	// Interpolate
	r := interpolate(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R), fx, fy)
	g := interpolate(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G), fx, fy)
	b := interpolate(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B), fx, fy)
	a := interpolate(float64(c00.A), float64(c10.A), float64(c01.A), float64(c11.A), fx, fy)

	return color.RGBA{
		R: uint8(clamp(r, 0, 255)),
		G: uint8(clamp(g, 0, 255)),
		B: uint8(clamp(b, 0, 255)),
		A: uint8(clamp(a, 0, 255)),
	}
}

// interpolate performs bilinear interpolation.
func interpolate(c00, c10, c01, c11, fx, fy float64) float64 {
	c0 := c00*(1-fx) + c10*fx
	c1 := c01*(1-fx) + c11*fx
	return c0*(1-fy) + c1*fy
}

// rgbToHSL converts RGB color to HSL color space.
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(math.Max(rf, gf), bf)
	min := math.Min(math.Min(rf, gf), bf)
	l = (max + min) / 2.0

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = (bf-rf)/d + 2.0
		case bf:
			h = (rf-gf)/d + 4.0
		}
		h *= 60.0
	}

	return h, s, l
}

// hslToRGB converts HSL color to RGB color space.
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		val := uint8(clamp(l*255.0, 0, 255))
		return val, val, val
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}
	p := 2.0*l - q

	h = h / 360.0
	tr := h + 1.0/3.0
	tg := h
	tb := h - 1.0/3.0

	r = uint8(clamp(hueToRGB(p, q, tr)*255.0, 0, 255))
	g = uint8(clamp(hueToRGB(p, q, tg)*255.0, 0, 255))
	b = uint8(clamp(hueToRGB(p, q, tb)*255.0, 0, 255))

	return r, g, b
}

// hueToRGB is a helper function for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6.0*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

// clamp restricts a value to a range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
