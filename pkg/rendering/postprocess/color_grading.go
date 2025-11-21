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
			r, g, b, a := p.processPixel(c, config)
			result.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return result
}

// processPixel processes a single pixel with color grading adjustments.
func (p *Processor) processPixel(c color.RGBA, config ColorGradingConfig) (r, g, b, a uint8) {
	rf, gf, bf, af := normalizeColor(c)

	rf, gf, bf = applyBrightnessAdjustment(rf, gf, bf, config.Brightness)
	rf, gf, bf = applyContrastAdjustment(rf, gf, bf, config.Contrast)
	rf, gf, bf = applySaturationAdjustment(rf, gf, bf, config.Saturation)
	rf, gf, bf = applyTemperatureAndTint(rf, gf, bf, config.Temperature, config.Tint)

	return clampUint8(rf * 255.0), clampUint8(gf * 255.0), clampUint8(bf * 255.0), clampUint8(af * 255.0)
}

// normalizeColor converts RGBA uint8 values to float64 in range 0.0-1.0.
func normalizeColor(c color.RGBA) (r, g, b, a float64) {
	return float64(c.R) / 255.0, float64(c.G) / 255.0, float64(c.B) / 255.0, float64(c.A) / 255.0
}

// applyBrightnessAdjustment applies brightness adjustment to RGB values.
func applyBrightnessAdjustment(r, g, b, brightness float64) (float64, float64, float64) {
	if brightness == 0.0 {
		return r, g, b
	}
	return clamp(r+brightness, 0.0, 1.0), clamp(g+brightness, 0.0, 1.0), clamp(b+brightness, 0.0, 1.0)
}

// applyContrastAdjustment applies contrast adjustment to RGB values.
func applyContrastAdjustment(r, g, b, contrast float64) (float64, float64, float64) {
	if contrast == 1.0 {
		return r, g, b
	}
	return clamp((r-0.5)*contrast+0.5, 0.0, 1.0), clamp((g-0.5)*contrast+0.5, 0.0, 1.0), clamp((b-0.5)*contrast+0.5, 0.0, 1.0)
}

// applySaturationAdjustment applies saturation adjustment to RGB values.
func applySaturationAdjustment(r, g, b, saturation float64) (float64, float64, float64) {
	if saturation == 1.0 {
		return r, g, b
	}
	lum := luminance(r, g, b)
	return clamp(lerp(lum, r, saturation), 0.0, 1.0), clamp(lerp(lum, g, saturation), 0.0, 1.0), clamp(lerp(lum, b, saturation), 0.0, 1.0)
}

// applyTemperatureAndTint applies temperature and tint adjustments to RGB values.
func applyTemperatureAndTint(r, g, b, temperature, tint float64) (float64, float64, float64) {
	if temperature == 0.0 && tint == 0.0 {
		return r, g, b
	}

	r, g, b = applyTemperature(r, g, b, temperature)
	r, g, b = applyTint(r, g, b, tint)
	return r, g, b
}

// applyTemperature applies temperature adjustment (cool/warm).
func applyTemperature(r, g, b, temperature float64) (float64, float64, float64) {
	if temperature < 0 {
		strength := -temperature
		return clamp(r-strength*0.2, 0.0, 1.0), g, clamp(b+strength*0.3, 0.0, 1.0)
	} else if temperature > 0 {
		strength := temperature
		return clamp(r+strength*0.3, 0.0, 1.0), clamp(g+strength*0.15, 0.0, 1.0), clamp(b-strength*0.2, 0.0, 1.0)
	}
	return r, g, b
}

// applyTint applies tint adjustment (green/magenta).
func applyTint(r, g, b, tint float64) (float64, float64, float64) {
	if tint < 0 {
		strength := -tint
		return clamp(r-strength*0.1, 0.0, 1.0), clamp(g+strength*0.2, 0.0, 1.0), clamp(b-strength*0.1, 0.0, 1.0)
	} else if tint > 0 {
		strength := tint
		return clamp(r+strength*0.2, 0.0, 1.0), clamp(g-strength*0.15, 0.0, 1.0), clamp(b+strength*0.2, 0.0, 1.0)
	}
	return r, g, b
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
