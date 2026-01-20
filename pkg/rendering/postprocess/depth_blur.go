// Package postprocess provides post-processing effects for rendered scenes.
// Code relocated from: depth_blur.go (constants moved to constants.go)
package postprocess

import (
	"image"
	"image/color"
	"math"
)

// ApplyDepthBlur applies depth-of-field blur to an image based on a depth map.
// The depth map should be an RGBA image where luminance represents depth (0=near, 1=far).
// Objects at the focal distance remain sharp, while objects outside the focal range are blurred.
func (p *Processor) ApplyDepthBlur(img, depthMap *image.RGBA) *image.RGBA {
	if !p.config.DepthBlur.Enabled || depthMap == nil {
		return img
	}

	config := p.config.DepthBlur
	if config.BlurStrength <= 0 || config.Samples < 1 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// For each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get depth value (0.0 = near, 1.0 = far)
			depthColor := depthMap.RGBAAt(x, y)
			depth := luminance(
				float64(depthColor.R)/255.0,
				float64(depthColor.G)/255.0,
				float64(depthColor.B)/255.0,
			)

			// Calculate blur amount based on distance from focal plane
			distFromFocus := math.Abs(depth - config.FocalDistance)
			blurAmount := 0.0

			if distFromFocus > config.FocalRange {
				// Outside focal range - apply blur
				blurAmount = (distFromFocus - config.FocalRange) * config.BlurStrength
				blurAmount = clamp(blurAmount, 0.0, 1.0)
			}

			// If no blur needed, just copy pixel
			if blurAmount < 0.01 {
				result.Set(x, y, img.At(x, y))
				continue
			}

			// Calculate blur radius (0 to maxBlurRadiusPixels based on blur amount)
			blurRadius := blurAmount * maxBlurRadiusPixels

			// Sample in a circle around the pixel with even spatial distribution
			var r, g, b, a float64
			samples := config.Samples
			count := 0

			for i := 0; i < samples; i++ {
				// Distribute samples evenly in a circle
				angle := (float64(i) / float64(samples)) * 2.0 * math.Pi
				// Square root for even spatial distribution (uniform area coverage)
				radius := blurRadius * math.Sqrt(float64(i)/float64(samples))

				sx := float64(x) + math.Cos(angle)*radius
				sy := float64(y) + math.Sin(angle)*radius

				// Sample with bilinear interpolation
				c := sampleBilinear(img, sx, sy)
				r += float64(c.R)
				g += float64(c.G)
				b += float64(c.B)
				a += float64(c.A)
				count++
			}

			// Also include center pixel
			centerColor := img.RGBAAt(x, y)
			r += float64(centerColor.R)
			g += float64(centerColor.G)
			b += float64(centerColor.B)
			a += float64(centerColor.A)
			count++

			// Average samples
			finalCount := float64(count)
			result.Set(x, y, color.RGBA{
				R: clampUint8(r / finalCount),
				G: clampUint8(g / finalCount),
				B: clampUint8(b / finalCount),
				A: clampUint8(a / finalCount),
			})
		}
	}

	return result
}

// GenerateDepthMapFromLuminance creates a simple depth map from image luminance.
// This is a fallback when no actual depth information is available.
// Brighter areas are assumed to be closer (typical for lit scenes).
func GenerateDepthMapFromLuminance(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	depthMap := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			lum := luminance(
				float64(c.R)/255.0,
				float64(c.G)/255.0,
				float64(c.B)/255.0,
			)

			// Invert luminance (bright = near = low depth value)
			depth := 1.0 - lum
			depthValue := clampUint8(depth * 255.0)

			depthMap.Set(x, y, color.RGBA{
				R: depthValue,
				G: depthValue,
				B: depthValue,
				A: 255,
			})
		}
	}

	return depthMap
}

// GenerateDepthMapFromY creates a simple depth map based on Y coordinate.
// Top of screen is near (0.0), bottom of screen is far (1.0).
// Useful for side-view or top-down games.
func GenerateDepthMapFromY(bounds image.Rectangle) *image.RGBA {
	depthMap := image.NewRGBA(bounds)
	height := float64(bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Calculate depth based on Y position (0 at top, 1 at bottom)
		depth := float64(y-bounds.Min.Y) / height
		depthValue := clampUint8(depth * 255.0)

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			depthMap.Set(x, y, color.RGBA{
				R: depthValue,
				G: depthValue,
				B: depthValue,
				A: 255,
			})
		}
	}

	return depthMap
}
