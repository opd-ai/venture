// Package postprocess provides post-processing effects for rendered scenes.
// Code relocated from: chromatic_aberration.go (constants moved to constants.go)
package postprocess

import (
	"image"
	"image/color"
	"math"
)

// ApplyChromaticAberration applies chromatic aberration effect that separates color channels.
// This simulates lens distortion found in analog cameras.
func (p *Processor) ApplyChromaticAberration(img *image.RGBA) *image.RGBA {
	if !p.config.ChromaticAberration.Enabled {
		return img
	}

	config := p.config.ChromaticAberration
	if config.Intensity <= 0 || config.Samples < 2 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Calculate image center
	centerX := float64(bounds.Min.X+bounds.Max.X) / 2.0
	centerY := float64(bounds.Min.Y+bounds.Max.Y) / 2.0
	maxRadius := math.Sqrt(
		math.Pow(float64(bounds.Dx())/2.0, 2) +
			math.Pow(float64(bounds.Dy())/2.0, 2),
	)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate distance from center (normalized 0-1)
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx+dy*dy) / maxRadius

			// Aberration strength increases with distance from center
			aberrationStrength := dist * config.Intensity * chromaticAberrationScale

			// Sample red channel with offset
			var rTotal, gTotal, bTotal, aTotal float64
			samples := config.Samples

			for i := 0; i < samples; i++ {
				// Calculate offset for this sample
				t := float64(i) / float64(samples-1) // 0 to 1
				offset := (t - 0.5) * 2.0            // -1 to 1

				// Red channel - shift in direction
				rx := float64(x) + config.DirectionX*aberrationStrength*offset
				ry := float64(y) + config.DirectionY*aberrationStrength*offset
				rc := sampleBilinear(img, rx, ry)
				rTotal += float64(rc.R)

				// Green channel - no offset (center)
				gc := img.RGBAAt(x, y)
				gTotal += float64(gc.G)

				// Blue channel - shift opposite direction
				bx := float64(x) - config.DirectionX*aberrationStrength*offset
				by := float64(y) - config.DirectionY*aberrationStrength*offset
				bc := sampleBilinear(img, bx, by)
				bTotal += float64(bc.B)

				// Alpha from center pixel
				aTotal += float64(gc.A)
			}

			// Average samples
			count := float64(samples)
			result.Set(x, y, color.RGBA{
				R: clampUint8(rTotal / count),
				G: clampUint8(gTotal / count),
				B: clampUint8(bTotal / count),
				A: clampUint8(aTotal / count),
			})
		}
	}

	return result
}

// ApplyPrismaticAberration applies angle-based chromatic separation creating
// a rainbow-like prismatic color split. angle is in radians [0, 2π] and
// controls the direction of color channel displacement. intensity controls
// the magnitude of displacement [0.0, 1.0] where 0 is a no-op.
// This is a standalone utility function rather than a Processor method since it
// uses custom parameters not part of the standard ChromaticAberrationConfig.
func ApplyPrismaticAberration(img *image.RGBA, intensity, angle float64) *image.RGBA {
	if intensity <= 0 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Calculate direction vector from angle
	dirX := math.Cos(angle * math.Pi / 180.0)
	dirY := math.Sin(angle * math.Pi / 180.0)

	// Calculate image center
	centerX := float64(bounds.Min.X+bounds.Max.X) / 2.0
	centerY := float64(bounds.Min.Y+bounds.Max.Y) / 2.0
	maxRadius := math.Sqrt(
		math.Pow(float64(bounds.Dx())/2.0, 2) +
			math.Pow(float64(bounds.Dy())/2.0, 2),
	)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate distance from center (normalized 0-1)
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx+dy*dy) / maxRadius

			// Aberration strength increases with distance from center
			aberrationStrength := dist * intensity * prismaticAberrationScale

			// Sample each color channel with different offsets
			// Red shifts most, blue shifts least (simulating refraction)
			rx := float64(x) + dirX*aberrationStrength*1.0
			ry := float64(y) + dirY*aberrationStrength*1.0
			rc := sampleBilinear(img, rx, ry)

			gx := float64(x) + dirX*aberrationStrength*0.5
			gy := float64(y) + dirY*aberrationStrength*0.5
			gc := sampleBilinear(img, gx, gy)

			bx := float64(x) + dirX*aberrationStrength*0.0
			by := float64(y) + dirY*aberrationStrength*0.0
			bc := sampleBilinear(img, bx, by)

			// Use center pixel alpha
			centerColor := img.RGBAAt(x, y)

			result.Set(x, y, color.RGBA{
				R: rc.R,
				G: gc.G,
				B: bc.B,
				A: centerColor.A,
			})
		}
	}

	return result
}
