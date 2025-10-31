// Package tiles provides Phase 11.1 diagonal wall and multi-layer terrain rendering.
// This file implements rendering for diagonal walls (45° angles) and multi-layer
// terrain features (platforms, pits, ramps).
package tiles

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// DiagonalDirection represents the direction of a diagonal wall.
type DiagonalDirection int

const (
	// DiagonalNE represents a diagonal from bottom-left to top-right (/)
	DiagonalNE DiagonalDirection = iota
	// DiagonalNW represents a diagonal from bottom-right to top-left (\)
	DiagonalNW
	// DiagonalSE represents a diagonal from top-left to bottom-right (\)
	DiagonalSE
	// DiagonalSW represents a diagonal from top-right to bottom-left (/)
	DiagonalSW
)

// generateDiagonalWall creates a diagonal wall tile.
// Phase 11.1: Renders diagonal walls at 45° angles using triangle fill.
func (g *Generator) generateDiagonalWall(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config, direction DiagonalDirection) {
	baseColor := g.pickColor(pal, "wall", rng)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create background (floor-like for contrast)
	floorColor := g.pickColor(pal, "floor", rng)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, floorColor)
		}
	}

	// Define triangle vertices based on direction
	var x1, y1, x2, y2, x3, y3 int
	switch direction {
	case DiagonalNE: // /
		x1, y1 = 0, height
		x2, y2 = 0, 0
		x3, y3 = width, 0
	case DiagonalNW: // \
		x1, y1 = 0, 0
		x2, y2 = width, 0
		x3, y3 = width, height
	case DiagonalSE: // \
		x1, y1 = 0, 0
		x2, y2 = 0, height
		x3, y3 = width, height
	case DiagonalSW: // /
		x1, y1 = 0, height
		x2, y2 = width, height
		x3, y3 = width, 0
	}

	// Fill triangle with wall pattern
	g.fillTriangle(img, x1, y1, x2, y2, x3, y3, baseColor, rng, config.Variant)

	// Add depth with shadow gradient
	shadowColor := g.darkenColor(baseColor, 0.3)
	g.addDiagonalShadow(img, direction, shadowColor, 2)
}

// fillTriangle fills a triangle defined by three vertices.
func (g *Generator) fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, baseColor color.Color, rng *rand.Rand, variance float64) {
	bounds := img.Bounds()

	// Compute bounding box
	minX := min(x1, min(x2, x3))
	maxX := max(x1, max(x2, x3))
	minY := min(y1, min(y2, y3))
	maxY := max(y1, max(y2, y3))

	// Clip to image bounds
	if minX < bounds.Min.X {
		minX = bounds.Min.X
	}
	if maxX > bounds.Max.X {
		maxX = bounds.Max.X
	}
	if minY < bounds.Min.Y {
		minY = bounds.Min.Y
	}
	if maxY > bounds.Max.Y {
		maxY = bounds.Max.Y
	}

	r, gr, b, a := baseColor.RGBA()

	// Fill pixels inside triangle
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			if g.isInsideTriangle(x, y, x1, y1, x2, y2, x3, y3) {
				// Add subtle texture variation
				variation := 1.0 + (rng.Float64()*2.0-1.0)*variance*0.1
				varR := uint8(math.Min(255, float64(r>>8)*variation))
				varG := uint8(math.Min(255, float64(gr>>8)*variation))
				varB := uint8(math.Min(255, float64(b>>8)*variation))

				img.Set(x, y, color.RGBA{R: varR, G: varG, B: varB, A: uint8(a >> 8)})
			}
		}
	}
}

// isInsideTriangle checks if a point (px, py) is inside triangle (x1,y1), (x2,y2), (x3,y3).
func (g *Generator) isInsideTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	// Use barycentric coordinates
	d1 := g.sign(px, py, x1, y1, x2, y2)
	d2 := g.sign(px, py, x2, y2, x3, y3)
	d3 := g.sign(px, py, x3, y3, x1, y1)

	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)

	return !(hasNeg && hasPos)
}

// sign computes the sign of the cross product for triangle point test.
func (g *Generator) sign(px, py, x1, y1, x2, y2 int) float64 {
	return float64((px-x2)*(y1-y2) - (x1-x2)*(py-y2))
}

// addDiagonalShadow adds a shadow gradient along the diagonal edge.
func (g *Generator) addDiagonalShadow(img *image.RGBA, direction DiagonalDirection, shadowColor color.Color, thickness int) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Add shadow on the "inner" edge of the diagonal
	for t := 0; t < thickness; t++ {
		switch direction {
		case DiagonalNE: // /
			for i := 0; i < width && i < height; i++ {
				if i+t < width && i+t < height {
					img.Set(bounds.Min.X+i+t, bounds.Min.Y+height-1-i-t, shadowColor)
				}
			}
		case DiagonalNW: // \
			for i := 0; i < width && i < height; i++ {
				if i+t < width && i+t < height {
					img.Set(bounds.Min.X+i+t, bounds.Min.Y+i+t, shadowColor)
				}
			}
		case DiagonalSE: // \
			for i := 0; i < width && i < height; i++ {
				if i+t < width && i+t < height {
					img.Set(bounds.Min.X+i+t, bounds.Min.Y+i+t, shadowColor)
				}
			}
		case DiagonalSW: // /
			for i := 0; i < width && i < height; i++ {
				if i+t < width && i+t < height {
					img.Set(bounds.Min.X+i+t, bounds.Min.Y+height-1-i-t, shadowColor)
				}
			}
		}
	}
}

// generatePlatform creates an elevated platform tile.
// Phase 11.1: Platforms appear raised with 3D shading.
func (g *Generator) generatePlatform(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	baseColor := g.pickColor(pal, "floor", rng)
	bounds := img.Bounds()

	// Fill platform surface
	g.fillSolid(img, baseColor, config.Variant, rng)

	// Add raised edge effect (lighter on top/left, darker on bottom/right)
	highlightColor := g.lightenColor(baseColor, 0.3)
	shadowColor := g.darkenColor(baseColor, 0.3)

	thickness := 3

	// Top edge highlight
	for y := bounds.Min.Y; y < bounds.Min.Y+thickness; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, highlightColor)
		}
	}

	// Left edge highlight
	for x := bounds.Min.X; x < bounds.Min.X+thickness; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(x, y, highlightColor)
		}
	}

	// Bottom edge shadow
	for y := bounds.Max.Y - thickness; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, shadowColor)
		}
	}

	// Right edge shadow
	for x := bounds.Max.X - thickness; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(x, y, shadowColor)
		}
	}
}

// generateRamp creates a ramp tile for layer transitions.
// Phase 11.1: Ramps show gradient from low to high.
func (g *Generator) generateRamp(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	baseColor := g.pickColor(pal, "floor", rng)
	bounds := img.Bounds()
	height := bounds.Dy()

	r, gr, b, a := baseColor.RGBA()

	// Vertical gradient from dark (top/Y=0) to light (bottom/Y=max)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Calculate gradient factor (0.7 at top, 1.0 at bottom)
		factor := 0.7 + float64(y-bounds.Min.Y)/float64(height)*0.3

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			variation := 1.0 + (rng.Float64()*2.0-1.0)*config.Variant*0.05
			finalFactor := factor * variation

			varR := uint8(math.Min(255, float64(r>>8)*finalFactor))
			varG := uint8(math.Min(255, float64(gr>>8)*finalFactor))
			varB := uint8(math.Min(255, float64(b>>8)*finalFactor))

			img.Set(x, y, color.RGBA{R: varR, G: varG, B: varB, A: uint8(a >> 8)})
		}
	}

	// Add horizontal lines to suggest steps
	stepCount := 4
	stepSpacing := height / stepCount
	lineColor := g.darkenColor(baseColor, 0.2)

	for i := 1; i < stepCount; i++ {
		y := bounds.Min.Y + i*stepSpacing
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if y < bounds.Max.Y {
				img.Set(x, y, lineColor)
			}
		}
	}
}

// generatePit creates a pit/chasm tile.
// Phase 11.1: Pits appear as dark voids with depth.
func (g *Generator) generatePit(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Use dark background color
	pitColor := g.darkenColor(pal.Background, 0.6)
	bounds := img.Bounds()

	// Fill with dark base
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, pitColor)
		}
	}

	// Add vignette effect to center (darker in middle)
	centerX := bounds.Min.X + bounds.Dx()/2
	centerY := bounds.Min.Y + bounds.Dy()/2
	maxDist := math.Sqrt(float64(bounds.Dx()*bounds.Dx()+bounds.Dy()*bounds.Dy())) / 2.0

	r, gr, b, a := pitColor.RGBA()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			// Darken based on distance from center
			darkFactor := 1.0 - (dist / maxDist * 0.4)
			if darkFactor < 0.6 {
				darkFactor = 0.6
			}

			varR := uint8(float64(r>>8) * darkFactor)
			varG := uint8(float64(gr>>8) * darkFactor)
			varB := uint8(float64(b>>8) * darkFactor)

			img.Set(x, y, color.RGBA{R: varR, G: varG, B: varB, A: uint8(a >> 8)})
		}
	}

	// Add subtle edge highlights to show depth
	edgeColor := g.lightenColor(pitColor, 0.2)
	thickness := 2

	for t := 0; t < thickness; t++ {
		// Top edge
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, bounds.Min.Y+t, edgeColor)
		}
		// Left edge
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(bounds.Min.X+t, y, edgeColor)
		}
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
