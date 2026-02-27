// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
)

// ApplyMotionBlur applies velocity-based motion blur to an image.
// If velocityMap is nil, uses the global velocity from config.
func (p *Processor) ApplyMotionBlur(img *image.RGBA, velocityMap *VelocityMap) *image.RGBA {
	if !p.config.MotionBlur.Enabled {
		return img
	}

	config := p.config.MotionBlur
	if config.Intensity <= 0 || config.Samples < 2 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// For each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get velocity for this pixel
			var vx, vy float64
			if velocityMap != nil {
				vx, vy = velocityMap.GetVelocity(x, y)
			} else {
				vx = config.VelocityX
				vy = config.VelocityY
			}

			// If no velocity, just copy pixel
			if vx == 0 && vy == 0 {
				result.Set(x, y, img.At(x, y))
				continue
			}

			// Sample along motion direction
			var r, g, b, a float64
			samples := config.Samples

			for i := 0; i < samples; i++ {
				// Calculate sample position
				t := float64(i) / float64(samples-1)   // 0 to 1
				offset := (t - 0.5) * config.Intensity // -0.5 to 0.5 scaled by intensity
				sx := float64(x) + vx*offset
				sy := float64(y) + vy*offset

				// Sample with bilinear interpolation
				c := sampleBilinear(img, sx, sy)
				r += float64(c.R)
				g += float64(c.G)
				b += float64(c.B)
				a += float64(c.A)
			}

			// Average samples
			count := float64(samples)
			result.Set(x, y, color.RGBA{
				R: clampUint8(r / count),
				G: clampUint8(g / count),
				B: clampUint8(b / count),
				A: clampUint8(a / count),
			})
		}
	}

	return result
}

// CreateUniformVelocityMap creates a velocity map with uniform velocity across all pixels.
// Useful for camera panning effects or constant-direction motion.
//
// Example:
//
//	// Camera panning right with 5 pixel/frame velocity
//	velMap := CreateUniformVelocityMap(image.Rect(0, 0, 800, 600), 5.0, 0.0)
//	blurred := ApplyMotionBlur(srcImage, velMap, 8)
func CreateUniformVelocityMap(bounds image.Rectangle, vx, vy float64) *VelocityMap {
	velMap := NewVelocityMap(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			velMap.SetVelocity(x, y, vx, vy)
		}
	}

	return velMap
}

// CreateRadialVelocityMap creates a velocity map with radial velocity from a center point.
// Useful for explosion effects or radial camera motion.
//
// Example:
//
//	// Explosion at center of screen with strength 10.0
//	centerX, centerY := 400, 300
//	velMap := CreateRadialVelocityMap(image.Rect(0, 0, 800, 600), centerX, centerY, 10.0)
//	blurred := ApplyMotionBlur(explosionImage, velMap, 16)
func CreateRadialVelocityMap(bounds image.Rectangle, centerX, centerY int, strength float64) *VelocityMap {
	velMap := NewVelocityMap(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate direction from center
			dx := float64(x - centerX)
			dy := float64(y - centerY)

			// Normalize and scale by strength
			dist := (dx*dx + dy*dy)
			if dist > 0 {
				dist = 1.0 / (dist + 1.0) // Inverse distance falloff
				vx := dx * strength * dist
				vy := dy * strength * dist
				velMap.SetVelocity(x, y, vx, vy)
			}
		}
	}

	return velMap
}
