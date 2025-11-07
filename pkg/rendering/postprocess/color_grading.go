// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
	"math"
)

// ApplyColorGrading applies color grading adjustments to an image.
// This includes saturation, contrast, brightness, temperature, and tint adjustments.
func (p *Processor) ApplyColorGrading(img *image.RGBA) *image.RGBA {
	if !p.config.ColorGrading.Enabled {
		return img
	}

	config := p.config.ColorGrading
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)

			// Convert to float 0.0-1.0
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0
			a := float64(c.A) / 255.0

			// Apply brightness
			if config.Brightness != 0.0 {
				r = clamp(r+config.Brightness, 0.0, 1.0)
				g = clamp(g+config.Brightness, 0.0, 1.0)
				b = clamp(b+config.Brightness, 0.0, 1.0)
			}

			// Apply contrast
			if config.Contrast != 1.0 {
				r = clamp((r-0.5)*config.Contrast+0.5, 0.0, 1.0)
				g = clamp((g-0.5)*config.Contrast+0.5, 0.0, 1.0)
				b = clamp((b-0.5)*config.Contrast+0.5, 0.0, 1.0)
			}

			// Apply saturation
			if config.Saturation != 1.0 {
				lum := luminance(r, g, b)
				r = clamp(lerp(lum, r, config.Saturation), 0.0, 1.0)
				g = clamp(lerp(lum, g, config.Saturation), 0.0, 1.0)
				b = clamp(lerp(lum, b, config.Saturation), 0.0, 1.0)
			}

			// Apply temperature and tint
			if config.Temperature != 0.0 || config.Tint != 0.0 {
				// Temperature: negative = cooler (more blue), positive = warmer (more orange)
				if config.Temperature < 0 {
					// Cool (add blue, reduce red)
					strength := -config.Temperature
					r = clamp(r-strength*0.2, 0.0, 1.0)
					b = clamp(b+strength*0.3, 0.0, 1.0)
				} else if config.Temperature > 0 {
					// Warm (add red/orange, reduce blue)
					strength := config.Temperature
					r = clamp(r+strength*0.3, 0.0, 1.0)
					g = clamp(g+strength*0.15, 0.0, 1.0)
					b = clamp(b-strength*0.2, 0.0, 1.0)
				}

				// Tint: negative = green, positive = magenta
				if config.Tint < 0 {
					// Green tint
					strength := -config.Tint
					g = clamp(g+strength*0.2, 0.0, 1.0)
					r = clamp(r-strength*0.1, 0.0, 1.0)
					b = clamp(b-strength*0.1, 0.0, 1.0)
				} else if config.Tint > 0 {
					// Magenta tint
					strength := config.Tint
					r = clamp(r+strength*0.2, 0.0, 1.0)
					b = clamp(b+strength*0.2, 0.0, 1.0)
					g = clamp(g-strength*0.15, 0.0, 1.0)
				}
			}

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

// ApplyGrayscale converts an image to grayscale using luminance.
func ApplyGrayscale(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			lum := luminance(
				float64(c.R)/255.0,
				float64(c.G)/255.0,
				float64(c.B)/255.0,
			)
			gray := clampUint8(lum * 255.0)

			result.Set(x, y, color.RGBA{
				R: gray,
				G: gray,
				B: gray,
				A: c.A,
			})
		}
	}

	return result
}

// ApplySepia applies a sepia tone effect to an image.
func ApplySepia(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0

			// Sepia transformation matrix
			tr := r*0.393 + g*0.769 + b*0.189
			tg := r*0.349 + g*0.686 + b*0.168
			tb := r*0.272 + g*0.534 + b*0.131

			result.Set(x, y, color.RGBA{
				R: clampUint8(tr * 255.0),
				G: clampUint8(tg * 255.0),
				B: clampUint8(tb * 255.0),
				A: c.A,
			})
		}
	}

	return result
}

// ApplyHueShift shifts the hue of all colors in an image.
// shift is in the range -1.0 to 1.0 (maps to -180° to +180°).
func ApplyHueShift(img *image.RGBA, shift float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			r := float64(c.R) / 255.0
			g := float64(c.G) / 255.0
			b := float64(c.B) / 255.0

			// Convert to HSL
			h, s, l := rgbToHSL(r, g, b)

			// Shift hue
			h = math.Mod(h+shift+1.0, 1.0)

			// Convert back to RGB
			nr, ng, nb := hslToRGB(h, s, l)

			result.Set(x, y, color.RGBA{
				R: clampUint8(nr * 255.0),
				G: clampUint8(ng * 255.0),
				B: clampUint8(nb * 255.0),
				A: c.A,
			})
		}
	}

	return result
}
