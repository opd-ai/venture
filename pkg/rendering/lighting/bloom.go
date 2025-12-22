// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
	"math"
)

// BloomConfig contains configuration for bloom/glow effects.
type BloomConfig struct {
	// Enabled toggles bloom effect
	Enabled bool

	// Threshold is the brightness threshold for bloom (0.0-1.0)
	// Pixels brighter than this will bloom
	Threshold float64

	// Intensity controls bloom strength (0.0-2.0 typical)
	Intensity float64

	// Radius is the bloom spread in pixels (1-20 typical)
	Radius int

	// Samples is the number of blur samples per direction (3-9 typical)
	// Higher values create smoother bloom but cost more performance
	Samples int
}

// DefaultBloomConfig returns a sensible default bloom configuration.
// Phase 45: Radius and Samples scaled for 64×64 tiles (was 8/5 for 32×32).
func DefaultBloomConfig() BloomConfig {
	return BloomConfig{
		Enabled:   true,
		Threshold: 0.8,
		Intensity: 1.0,
		Radius:    16, // Phase 45: scaled from 8 to 16 for 64×64 tiles
		Samples:   7,  // Phase 45: scaled from 5 to 7 for smoother bloom
	}
}

// ApplyBloom applies a bloom/glow effect to bright areas of an image.
// Uses a two-pass separable Gaussian blur for performance.
func ApplyBloom(img *image.RGBA, config BloomConfig) *image.RGBA {
	if !config.Enabled || config.Intensity <= 0 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Step 1: Extract bright pixels above threshold
	brightMap := extractBrightPixels(img, config.Threshold)

	// Step 2: Apply two-pass separable Gaussian blur to bright map
	blurredH := gaussianBlurHorizontal(brightMap, config.Radius, config.Samples)
	blurred := gaussianBlurVertical(blurredH, config.Radius, config.Samples)

	// Step 3: Composite bloom with original image (additive blending)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Original pixel
			or, og, ob, oa := img.At(x, y).RGBA()
			orf := float64(or) / 65535.0
			ogf := float64(og) / 65535.0
			obf := float64(ob) / 65535.0
			oaf := float64(oa) / 65535.0

			// Bloom pixel
			br, bg, bb, _ := blurred.At(x, y).RGBA()
			brf := float64(br) / 65535.0
			bgf := float64(bg) / 65535.0
			bbf := float64(bb) / 65535.0

			// Additive blend with intensity control
			finalR := clamp(orf+brf*config.Intensity, 0, 1)
			finalG := clamp(ogf+bgf*config.Intensity, 0, 1)
			finalB := clamp(obf+bbf*config.Intensity, 0, 1)

			result.Set(x, y, color.RGBA{
				R: uint8(finalR * 255),
				G: uint8(finalG * 255),
				B: uint8(finalB * 255),
				A: uint8(oaf * 255),
			})
		}
	}

	return result
}

// extractBrightPixels creates a new image with only pixels above brightness threshold.
func extractBrightPixels(img *image.RGBA, threshold float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Calculate perceived brightness (luminance)
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0
			brightness := 0.2126*rf + 0.7152*gf + 0.0722*bf

			// Keep pixel if above threshold
			if brightness > threshold {
				// Scale down by threshold to maintain energy conservation
				scale := (brightness - threshold) / (1.0 - threshold)
				result.Set(x, y, color.RGBA{
					R: uint8(rf * scale * 255),
					G: uint8(gf * scale * 255),
					B: uint8(bf * scale * 255),
					A: uint8(float64(a) / 257),
				})
			}
		}
	}

	return result
}

// gaussianBlurHorizontal applies a horizontal Gaussian blur.
func gaussianBlurHorizontal(img *image.RGBA, radius, samples int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Pre-calculate Gaussian weights
	weights := calculateGaussianWeights(radius, samples)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sumR, sumG, sumB, sumW float64

			// Sample horizontally with Gaussian weights
			for i := 0; i < len(weights); i++ {
				offset := (i - samples) * radius / samples
				sampleX := x + offset

				if sampleX >= bounds.Min.X && sampleX < bounds.Max.X {
					r, g, b, _ := img.At(sampleX, y).RGBA()
					weight := weights[i]

					sumR += float64(r) / 65535.0 * weight
					sumG += float64(g) / 65535.0 * weight
					sumB += float64(b) / 65535.0 * weight
					sumW += weight
				}
			}

			// Normalize by total weight
			if sumW > 0 {
				result.Set(x, y, color.RGBA{
					R: uint8(clamp(sumR/sumW, 0, 1) * 255),
					G: uint8(clamp(sumG/sumW, 0, 1) * 255),
					B: uint8(clamp(sumB/sumW, 0, 1) * 255),
					A: 255,
				})
			}
		}
	}

	return result
}

// gaussianBlurVertical applies a vertical Gaussian blur.
func gaussianBlurVertical(img *image.RGBA, radius, samples int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Pre-calculate Gaussian weights
	weights := calculateGaussianWeights(radius, samples)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sumR, sumG, sumB, sumW float64

			// Sample vertically with Gaussian weights
			for i := 0; i < len(weights); i++ {
				offset := (i - samples) * radius / samples
				sampleY := y + offset

				if sampleY >= bounds.Min.Y && sampleY < bounds.Max.Y {
					r, g, b, _ := img.At(x, sampleY).RGBA()
					weight := weights[i]

					sumR += float64(r) / 65535.0 * weight
					sumG += float64(g) / 65535.0 * weight
					sumB += float64(b) / 65535.0 * weight
					sumW += weight
				}
			}

			// Normalize by total weight
			if sumW > 0 {
				result.Set(x, y, color.RGBA{
					R: uint8(clamp(sumR/sumW, 0, 1) * 255),
					G: uint8(clamp(sumG/sumW, 0, 1) * 255),
					B: uint8(clamp(sumB/sumW, 0, 1) * 255),
					A: 255,
				})
			}
		}
	}

	return result
}

// calculateGaussianWeights pre-calculates Gaussian weights for blur kernel.
// Returns an array of weights centered around 0 with specified samples.
func calculateGaussianWeights(radius, samples int) []float64 {
	weights := make([]float64, samples*2+1)
	sigma := float64(radius) / 3.0 // Standard deviation

	var sum float64
	for i := 0; i < len(weights); i++ {
		x := float64(i - samples)
		weight := math.Exp(-(x * x) / (2 * sigma * sigma))
		weights[i] = weight
		sum += weight
	}

	// Normalize weights to sum to 1
	for i := range weights {
		weights[i] /= sum
	}

	return weights
}
