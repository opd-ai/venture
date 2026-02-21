// Package sprites provides creature eye pattern rendering for nonhumanoid top-down sprites.
// This renderer draws creature-type-specific eye arrangements based on CreatureEyePatternComponent
// data: 8 eyes for arachnids, slit pupils for serpents, compound faceted eyes for insects, etc.
package sprites

import (
	"image"
	"image/color"
	"math"
)

// CreatureEyeRenderParams controls how creature eyes are drawn onto the head/body area.
type CreatureEyeRenderParams struct {
	// SpriteWidth and SpriteHeight are the full sprite dimensions
	SpriteWidth  int
	SpriteHeight int

	// HeadSpec is the head/body part where eyes are drawn
	HeadSpec PartSpec

	// EyePattern identifies which pattern type (e.g., "arachnid_8", "serpent_slit")
	EyePattern string

	// EyeCount is the number of eyes to render
	EyeCount int

	// EyePositions stores [x,y] pairs for each eye (relative to head bounds, 0.0-1.0)
	EyePositions []float64

	// EyeSizes stores relative size multiplier for each eye
	EyeSizes []float64

	// EyeColor (RGB 0.0-1.0) for the eye/iris
	EyeR, EyeG, EyeB float64

	// PupilStyle controls pupil rendering ("round", "slit_vertical", "slit_horizontal", "faceted", "none")
	PupilStyle string

	// GlowIntensity for mechanical/magical eyes (0.0-1.0)
	GlowIntensity float64

	// Direction the entity is facing
	Direction Direction

	// Seed for per-entity variation
	Seed int64
}

// RenderCreatureEyes draws creature-type-specific eye patterns onto the sprite image.
// Eyes are positioned within the head bounding box based on the pattern configuration.
func RenderCreatureEyes(img *image.RGBA, params CreatureEyeRenderParams) {
	if params.EyeCount <= 0 || len(params.EyePositions) < 2 {
		return
	}

	// Compute head bounding box in pixel coordinates
	headW := int(float64(params.SpriteWidth) * params.HeadSpec.RelativeWidth)
	headH := int(float64(params.SpriteHeight) * params.HeadSpec.RelativeHeight)
	headCX := int(float64(params.SpriteWidth) * params.HeadSpec.RelativeX)
	headCY := int(float64(params.SpriteHeight) * params.HeadSpec.RelativeY)
	headLeft := headCX - headW/2
	headTop := headCY - headH/2

	if headW < 4 || headH < 4 {
		return // Too small for eyes
	}

	bounds := img.Bounds()

	// Convert eye color to RGBA
	eyeColor := color.RGBA{
		R: uint8(clampCEP(params.EyeR*255, 0, 255)),
		G: uint8(clampCEP(params.EyeG*255, 0, 255)),
		B: uint8(clampCEP(params.EyeB*255, 0, 255)),
		A: 255,
	}

	// Draw each eye
	for i := 0; i < params.EyeCount; i++ {
		if i*2+1 >= len(params.EyePositions) {
			break
		}

		relX := params.EyePositions[i*2]
		relY := params.EyePositions[i*2+1]

		// Adjust positions based on facing direction
		adjX, adjY := adjustEyePositionForDirection(relX, relY, params.Direction)

		// Convert to pixel coordinates within head bounds
		eyeX := headLeft + int(adjX*float64(headW))
		eyeY := headTop + int(adjY*float64(headH))

		// Get size for this eye
		size := 1.0
		if i < len(params.EyeSizes) {
			size = params.EyeSizes[i]
		}

		// Base eye radius in pixels (scaled by head size and eye size multiplier)
		baseRadius := clampCEP(float64(min(headW, headH))/8.0*size, 1, 4)

		// Draw the eye based on pattern
		switch params.PupilStyle {
		case "slit_vertical":
			drawSlitEye(img, eyeX, eyeY, int(baseRadius), eyeColor, true, bounds, params.GlowIntensity)
		case "slit_horizontal":
			drawSlitEye(img, eyeX, eyeY, int(baseRadius), eyeColor, false, bounds, params.GlowIntensity)
		case "faceted":
			drawFacetedEye(img, eyeX, eyeY, int(baseRadius), eyeColor, bounds, params.GlowIntensity)
		case "none":
			// Simple dark bead eyes (spiders, blobs)
			drawBeadEye(img, eyeX, eyeY, int(baseRadius), eyeColor, bounds, params.GlowIntensity)
		default:
			// Round pupil (quadrupeds, flying, undead)
			drawRoundEye(img, eyeX, eyeY, int(baseRadius), eyeColor, bounds, params.GlowIntensity)
		}
	}
}

// adjustEyePositionForDirection shifts eye positions based on facing direction.
// From above, when facing down the face is visible; facing up shows back of head.
func adjustEyePositionForDirection(relX, relY float64, dir Direction) (float64, float64) {
	switch dir {
	case DirUp:
		// Back of head visible - shift eyes toward bottom (hidden)
		return relX, clampCEP(relY+0.3, 0.0, 1.0)
	case DirDown:
		// Face visible at bottom of head
		return relX, clampCEP(relY+0.15, 0.0, 1.0)
	case DirLeft:
		// Eyes on left side
		return clampCEP(relX-0.15, 0.0, 1.0), relY
	case DirRight:
		// Eyes on right side
		return clampCEP(relX+0.15, 0.0, 1.0), relY
	default:
		return relX, relY
	}
}

// drawRoundEye draws a standard round eye with pupil and highlight.
func drawRoundEye(img *image.RGBA, cx, cy, radius int, eyeColor color.RGBA, bounds image.Rectangle, glow float64) {
	if radius < 1 {
		radius = 1
	}

	// Draw sclera (white of eye) first
	scleraColor := color.RGBA{R: 245, G: 245, B: 240, A: 255}
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px, py := cx+dx, cy+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					img.SetRGBA(px, py, scleraColor)
				}
			}
		}
	}

	// Draw iris/pupil (smaller, centered)
	pupilRadius := max(1, radius/2)
	for dy := -pupilRadius; dy <= pupilRadius; dy++ {
		for dx := -pupilRadius; dx <= pupilRadius; dx++ {
			if dx*dx+dy*dy <= pupilRadius*pupilRadius {
				px, py := cx+dx, cy+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					img.SetRGBA(px, py, eyeColor)
				}
			}
		}
	}

	// Add specular highlight
	addEyeHighlight(img, cx, cy, radius, bounds, glow)
}

// drawSlitEye draws a reptilian eye with vertical or horizontal slit pupil.
func drawSlitEye(img *image.RGBA, cx, cy, radius int, eyeColor color.RGBA, vertical bool, bounds image.Rectangle, glow float64) {
	if radius < 1 {
		radius = 1
	}

	// Draw elliptical iris
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			// Ellipse check
			dist := float64(dx*dx+dy*dy) / float64(radius*radius)
			if dist <= 1.0 {
				px, py := cx+dx, cy+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					img.SetRGBA(px, py, eyeColor)
				}
			}
		}
	}

	// Draw slit pupil (dark line)
	slitColor := darkenColorCEP(eyeColor, 0.3)
	if vertical {
		// Vertical slit
		for dy := -radius; dy <= radius; dy++ {
			px, py := cx, cy+dy
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				img.SetRGBA(px, py, slitColor)
			}
		}
	} else {
		// Horizontal slit
		for dx := -radius; dx <= radius; dx++ {
			px, py := cx+dx, cy
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				img.SetRGBA(px, py, slitColor)
			}
		}
	}

	addEyeHighlight(img, cx, cy, radius, bounds, glow)
}

// drawFacetedEye draws a compound insect eye with faceted appearance.
func drawFacetedEye(img *image.RGBA, cx, cy, radius int, eyeColor color.RGBA, bounds image.Rectangle, glow float64) {
	if radius < 1 {
		radius = 1
	}

	// Draw main compound eye shape
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := float64(dx*dx+dy*dy) / float64(radius*radius)
			if dist <= 1.0 {
				px, py := cx+dx, cy+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					// Create faceted effect with checkerboard-like pattern
					isFacetLight := (px+py)%2 == 0
					if isFacetLight {
						img.SetRGBA(px, py, brightenColorCEP(eyeColor, 1.2))
					} else {
						img.SetRGBA(px, py, darkenColorCEP(eyeColor, 0.85))
					}
				}
			}
		}
	}

	// Add multiple small highlights for compound effect
	if radius >= 2 {
		highlightColor := color.RGBA{R: 255, G: 255, B: 255, A: 150}
		offsets := []struct{ dx, dy int }{{-radius / 2, -radius / 2}, {radius / 3, -radius / 3}}
		for _, off := range offsets {
			hx, hy := cx+off.dx, cy+off.dy
			if hx >= bounds.Min.X && hx < bounds.Max.X && hy >= bounds.Min.Y && hy < bounds.Max.Y {
				img.SetRGBA(hx, hy, highlightColor)
			}
		}
	}
}

// drawBeadEye draws a simple dark bead eye (for spiders, blobs).
func drawBeadEye(img *image.RGBA, cx, cy, radius int, eyeColor color.RGBA, bounds image.Rectangle, glow float64) {
	if radius < 1 {
		radius = 1
	}

	// Dark beady eyes - use eye color darkened
	beadColor := darkenColorCEP(eyeColor, 0.5)

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px, py := cx+dx, cy+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					img.SetRGBA(px, py, beadColor)
				}
			}
		}
	}

	// Small highlight for shine
	addEyeHighlight(img, cx, cy, radius, bounds, glow)
}

// addEyeHighlight adds a specular highlight to the eye.
func addEyeHighlight(img *image.RGBA, cx, cy, radius int, bounds image.Rectangle, glow float64) {
	// Position highlight at upper-left
	hx := cx - radius/2
	hy := cy - radius/2

	// Highlight intensity based on glow
	alpha := uint8(180 + int(glow*75))
	highlightColor := color.RGBA{R: 255, G: 255, B: 255, A: alpha}

	if hx >= bounds.Min.X && hx < bounds.Max.X && hy >= bounds.Min.Y && hy < bounds.Max.Y {
		img.SetRGBA(hx, hy, highlightColor)
	}

	// For glowing eyes, add additional glow pixels
	if glow > 0.3 && radius >= 2 {
		glowAlpha := uint8(int(glow * 100))
		glowColor := color.RGBA{R: 255, G: 255, B: 255, A: glowAlpha}
		for _, off := range []struct{ dx, dy int }{{0, -1}, {-1, 0}} {
			gx, gy := hx+off.dx, hy+off.dy
			if gx >= bounds.Min.X && gx < bounds.Max.X && gy >= bounds.Min.Y && gy < bounds.Max.Y {
				img.SetRGBA(gx, gy, glowColor)
			}
		}
	}
}

// darkenColorCEP returns a darker version of the color.
func darkenColorCEP(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(clampCEP(float64(c.R)*factor, 0, 255)),
		G: uint8(clampCEP(float64(c.G)*factor, 0, 255)),
		B: uint8(clampCEP(float64(c.B)*factor, 0, 255)),
		A: c.A,
	}
}

// brightenColorCEP returns a brighter version of the color.
func brightenColorCEP(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(clampCEP(float64(c.R)*factor, 0, 255)),
		G: uint8(clampCEP(float64(c.G)*factor, 0, 255)),
		B: uint8(clampCEP(float64(c.B)*factor, 0, 255)),
		A: c.A,
	}
}

// clampCEP clamps a value to [min, max].
func clampCEP(v, minV, maxV float64) float64 {
	return math.Max(minV, math.Min(maxV, v))
}
