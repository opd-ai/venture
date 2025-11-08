// Package ui provides procedural UI element generation.
// This file implements UI transition animations.
package ui

import (
	"image"
	"image/color"
	"math"
)

// ApplyEasing applies an easing function to a progress value (0.0 to 1.0).
func ApplyEasing(progress float64, easing EasingFunction) float64 {
	// Clamp progress to [0, 1]
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	switch easing {
	case EaseLinear:
		return progress

	case EaseInQuad:
		return progress * progress

	case EaseOutQuad:
		return progress * (2 - progress)

	case EaseInOutQuad:
		if progress < 0.5 {
			return 2 * progress * progress
		}
		return 1 - math.Pow(-2*progress+2, 2)/2

	case EaseInCubic:
		return progress * progress * progress

	case EaseOutCubic:
		return 1 - math.Pow(1-progress, 3)

	case EaseInOutCubic:
		if progress < 0.5 {
			return 4 * progress * progress * progress
		}
		return 1 - math.Pow(-2*progress+2, 3)/2

	default:
		return progress
	}
}

// ApplyTransition applies a transition effect to an image based on configuration.
// Returns a new image with the transition applied.
func (g *Generator) ApplyTransition(img *image.RGBA, config TransitionConfig) *image.RGBA {
	if config.Type == TransitionNone {
		return img
	}

	// Apply easing to progress
	easedProgress := ApplyEasing(config.Progress, config.Easing)

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	switch config.Type {
	case TransitionFade:
		g.applyFadeTransition(result, img, easedProgress)

	case TransitionSlideLeft:
		g.applySlideTransition(result, img, easedProgress, -1, 0)

	case TransitionSlideRight:
		g.applySlideTransition(result, img, easedProgress, 1, 0)

	case TransitionSlideUp:
		g.applySlideTransition(result, img, easedProgress, 0, -1)

	case TransitionSlideDown:
		g.applySlideTransition(result, img, easedProgress, 0, 1)

	case TransitionZoom:
		g.applyZoomTransition(result, img, easedProgress)

	default:
		// Unknown transition type, return original
		return img
	}

	return result
}

// applyFadeTransition applies a fade effect by modulating alpha channel.
func (g *Generator) applyFadeTransition(dst, src *image.RGBA, progress float64) {
	bounds := src.Bounds()
	alpha := uint8(progress * 255)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			srcColor := src.RGBAAt(x, y)

			// Modulate alpha channel by progress
			dstColor := color.RGBA{
				R: srcColor.R,
				G: srcColor.G,
				B: srcColor.B,
				A: uint8((uint32(srcColor.A) * uint32(alpha)) / 255),
			}

			dst.Set(x, y, dstColor)
		}
	}
}

// applySlideTransition applies a slide effect by offsetting the image.
// dirX and dirY are -1, 0, or 1 to indicate slide direction.
func (g *Generator) applySlideTransition(dst, src *image.RGBA, progress float64, dirX, dirY int) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate offset based on direction and progress
	offsetX := int(float64(width) * float64(dirX) * (1 - progress))
	offsetY := int(float64(height) * float64(dirY) * (1 - progress))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate destination position with offset
			dstX := x + offsetX
			dstY := y + offsetY

			// Only copy if destination is within bounds
			if dstX >= bounds.Min.X && dstX < bounds.Max.X &&
				dstY >= bounds.Min.Y && dstY < bounds.Max.Y {
				dst.Set(dstX, dstY, src.At(x, y))
			}
		}
	}
}

// applyZoomTransition applies a zoom effect by scaling the image.
func (g *Generator) applyZoomTransition(dst, src *image.RGBA, progress float64) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scale (0.0 at start, 1.0 at end)
	scale := progress

	// Calculate scaled dimensions
	scaledWidth := int(float64(width) * scale)
	scaledHeight := int(float64(height) * scale)

	// Calculate offset to center the scaled image
	offsetX := (width - scaledWidth) / 2
	offsetY := (height - scaledHeight) / 2

	// Simple nearest-neighbor scaling
	for dstY := 0; dstY < scaledHeight; dstY++ {
		for dstX := 0; dstX < scaledWidth; dstX++ {
			// Map destination pixel to source pixel
			srcX := bounds.Min.X + int(float64(dstX)/scale)
			srcY := bounds.Min.Y + int(float64(dstY)/scale)

			// Clamp to source bounds
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			if srcY >= bounds.Max.Y {
				srcY = bounds.Max.Y - 1
			}

			// Set destination pixel with offset
			finalX := offsetX + dstX
			finalY := offsetY + dstY

			if finalX >= 0 && finalX < width && finalY >= 0 && finalY < height {
				dst.Set(finalX, finalY, src.At(srcX, srcY))
			}
		}
	}
}

// InterpolateTransition creates a smooth transition between two UI elements.
// This is useful for crossfading or morphing between states.
func (g *Generator) InterpolateTransition(img1, img2 *image.RGBA, progress float64) *image.RGBA {
	if img1 == nil || img2 == nil {
		if img1 != nil {
			return img1
		}
		return img2
	}

	// Use the bounds of the first image
	bounds := img1.Bounds()
	result := image.NewRGBA(bounds)

	// Ensure progress is in [0, 1]
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	// Interpolate between the two images
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := color.RGBA{0, 0, 0, 0}

			// Get color from second image if within bounds
			if x < img2.Bounds().Max.X && y < img2.Bounds().Max.Y {
				c2 = img2.RGBAAt(x, y)
			}

			// Linear interpolation of colors
			r := uint8(float64(c1.R)*(1-progress) + float64(c2.R)*progress)
			gr := uint8(float64(c1.G)*(1-progress) + float64(c2.G)*progress)
			b := uint8(float64(c1.B)*(1-progress) + float64(c2.B)*progress)
			a := uint8(float64(c1.A)*(1-progress) + float64(c2.A)*progress)

			result.Set(x, y, color.RGBA{R: r, G: gr, B: b, A: a})
		}
	}

	return result
}

// GetTransitionDuration returns the duration of a transition in milliseconds.
func (t TransitionConfig) GetTransitionDuration() float64 {
	return t.Duration
}

// IsComplete returns true if the transition has completed.
func (t TransitionConfig) IsComplete() bool {
	return t.Progress >= 1.0
}

// UpdateProgress updates the transition progress based on elapsed time in milliseconds.
// Returns the new progress value (clamped to [0, 1]).
func (t *TransitionConfig) UpdateProgress(deltaTimeMs float64) float64 {
	if t.Duration <= 0 {
		t.Progress = 1.0
		return t.Progress
	}

	t.Progress += deltaTimeMs / t.Duration
	if t.Progress > 1.0 {
		t.Progress = 1.0
	}

	return t.Progress
}
