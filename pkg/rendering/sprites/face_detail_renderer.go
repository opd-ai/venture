// Package sprites provides pixel-level face detail rendering for top-down
// aerial entity sprites. This renderer draws eyes, mouth, and expression
// variations onto the head region of humanoid sprites, bridging the gap
// between the NpcFacialDetailComponent data and actual sprite pixels.
// All rendering is seed-deterministic and direction-aware.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// FaceRenderParams controls how facial features are drawn onto the head area.
type FaceRenderParams struct {
	// SpriteWidth and SpriteHeight are the full sprite dimensions.
	SpriteWidth  int
	SpriteHeight int

	// HeadSpec is the head body part layout from the anatomical template.
	HeadSpec PartSpec

	// EyeColor is the iris/pupil color.
	EyeColor color.RGBA

	// EyeSize in pixels (1-3); 2 is standard for humanoids.
	EyeSize int

	// MouthColor for the mouth feature.
	MouthColor color.RGBA

	// MouthSize in pixels (1-2).
	MouthSize int

	// Expression controls eye/mouth shape: "neutral", "hostile", "friendly", "scared".
	Expression string

	// Direction the entity is facing.
	Direction Direction

	// Seed for deterministic variation.
	Seed int64
}

// Predefined eye color palette for seed-based selection.
var eyeColors = []color.RGBA{
	{R: 80, G: 50, B: 25, A: 255},    // brown
	{R: 50, G: 35, B: 18, A: 255},    // dark brown
	{R: 45, G: 90, B: 160, A: 255},   // blue
	{R: 40, G: 120, B: 70, A: 255},   // green
	{R: 120, G: 95, B: 45, A: 255},   // hazel
	{R: 110, G: 120, B: 130, A: 255}, // grey
	{R: 155, G: 105, B: 35, A: 255},  // amber
}

// ComputeFaceParams builds FaceRenderParams from seed-based traits and template data.
func ComputeFaceParams(spriteW, spriteH int, headSpec PartSpec, direction Direction, seed int64) FaceRenderParams {
	rng := rand.New(rand.NewSource(seed ^ 0x7ACE))

	eyeIdx := rng.Intn(len(eyeColors))
	eye := varyColor(eyeColors[eyeIdx], rng, 15)

	mouthR := uint8(150 + rng.Intn(60))
	mouthG := uint8(80 + rng.Intn(50))
	mouthB := uint8(80 + rng.Intn(40))

	return FaceRenderParams{
		SpriteWidth:  spriteW,
		SpriteHeight: spriteH,
		HeadSpec:     headSpec,
		EyeColor:     eye,
		EyeSize:      2,
		MouthColor:   color.RGBA{R: mouthR, G: mouthG, B: mouthB, A: 255},
		MouthSize:    1,
		Expression:   "neutral",
		Direction:    direction,
		Seed:         seed,
	}
}

// ComputeFaceParamsFromComponent builds FaceRenderParams from NPC facial detail
// component data (EyeR/G/B, MouthR/G/B, EyeSize, MouthSize, ExpressionType).
func ComputeFaceParamsFromComponent(spriteW, spriteH int, headSpec PartSpec, direction Direction, seed int64,
	eyeR, eyeG, eyeB, mouthR, mouthG, mouthB, eyeSize, mouthSize float64, expression string,
) FaceRenderParams {
	clamp8 := func(v float64) uint8 {
		c := int(v * 255)
		if c < 0 {
			c = 0
		}
		if c > 255 {
			c = 255
		}
		return uint8(c)
	}

	es := int(math.Round(eyeSize))
	if es < 1 {
		es = 1
	}
	if es > 3 {
		es = 3
	}
	ms := int(math.Round(mouthSize))
	if ms < 1 {
		ms = 1
	}
	if ms > 2 {
		ms = 2
	}

	return FaceRenderParams{
		SpriteWidth:  spriteW,
		SpriteHeight: spriteH,
		HeadSpec:     headSpec,
		EyeColor:     color.RGBA{R: clamp8(eyeR), G: clamp8(eyeG), B: clamp8(eyeB), A: 255},
		EyeSize:      es,
		MouthColor:   color.RGBA{R: clamp8(mouthR), G: clamp8(mouthG), B: clamp8(mouthB), A: 255},
		MouthSize:    ms,
		Expression:   expression,
		Direction:    direction,
		Seed:         seed,
	}
}

// RenderFaceDetail draws eyes and mouth onto the sprite image at the head region.
// For top-down view: when facing down the face is visible on the lower portion
// of the head; when facing up the back of the head is shown (no face); left/right
// show the face on the corresponding side.
func RenderFaceDetail(img *image.RGBA, params FaceRenderParams) {
	if params.Direction == DirUp {
		// Back of head visible — no facial features
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
		return // Too small for facial features
	}

	// Eye positions depend on facing direction
	switch params.Direction {
	case DirDown:
		renderFaceDown(img, params, headLeft, headTop, headW, headH)
	case DirLeft:
		renderFaceLeft(img, params, headLeft, headTop, headW, headH)
	case DirRight:
		renderFaceRight(img, params, headLeft, headTop, headW, headH)
	}
}

// renderFaceDown draws eyes and mouth for a downward-facing (toward viewer) entity.
// From above: eyes appear on the lower half of the head circle.
func renderFaceDown(img *image.RGBA, params FaceRenderParams, hx, hy, hw, hh int) {
	bounds := img.Bounds()

	// Eyes positioned at ~65% down the head, spaced ~40% of head width apart
	eyeY := hy + int(float64(hh)*0.65)
	eyeSpacing := int(float64(hw) * 0.20)
	leftEyeX := hx + hw/2 - eyeSpacing
	rightEyeX := hx + hw/2 + eyeSpacing

	// Expression modifies eye shape
	eyeColor := params.EyeColor
	switch params.Expression {
	case "hostile":
		// Narrow angry eyes — darken and shift down slightly
		eyeY += 1
		eyeColor = darkenRGBA(eyeColor, 0.7)
	case "scared":
		// Wide eyes — brighten
		eyeColor = brightenRGBA(eyeColor, 1.3)
	}

	drawEye(img, leftEyeX, eyeY, params.EyeSize, eyeColor, params.Expression, bounds)
	drawEye(img, rightEyeX, eyeY, params.EyeSize, eyeColor, params.Expression, bounds)

	// Mouth — small feature below eyes
	if params.MouthSize > 0 {
		mouthY := eyeY + max(2, params.EyeSize)
		mouthX := hx + hw/2
		drawMouth(img, mouthX, mouthY, params.MouthSize, params.MouthColor, params.Expression, bounds)
	}
}

// renderFaceLeft draws facial features for a left-facing entity (face on left side).
func renderFaceLeft(img *image.RGBA, params FaceRenderParams, hx, hy, hw, hh int) {
	bounds := img.Bounds()

	// From above, looking left: one eye visible on left edge of head, at ~55% height
	eyeX := hx + int(float64(hw)*0.25)
	eyeY := hy + int(float64(hh)*0.55)

	eyeColor := expressionEyeColor(params.EyeColor, params.Expression)
	drawEye(img, eyeX, eyeY, params.EyeSize, eyeColor, params.Expression, bounds)

	// Mouth on left side
	if params.MouthSize > 0 {
		mouthX := eyeX + 1
		mouthY := eyeY + max(2, params.EyeSize)
		drawMouth(img, mouthX, mouthY, params.MouthSize, params.MouthColor, params.Expression, bounds)
	}
}

// renderFaceRight draws facial features for a right-facing entity.
func renderFaceRight(img *image.RGBA, params FaceRenderParams, hx, hy, hw, hh int) {
	bounds := img.Bounds()

	eyeX := hx + int(float64(hw)*0.75)
	eyeY := hy + int(float64(hh)*0.55)

	eyeColor := expressionEyeColor(params.EyeColor, params.Expression)
	drawEye(img, eyeX, eyeY, params.EyeSize, eyeColor, params.Expression, bounds)

	if params.MouthSize > 0 {
		mouthX := eyeX - 1
		mouthY := eyeY + max(2, params.EyeSize)
		drawMouth(img, mouthX, mouthY, params.MouthSize, params.MouthColor, params.Expression, bounds)
	}
}

// drawEye renders a single eye at the given position with specular highlight.
func drawEye(img *image.RGBA, cx, cy, size int, eyeColor color.RGBA, expression string, bounds image.Rectangle) {
	// Draw the dark pupil/iris area
	for dy := -size / 2; dy <= size/2; dy++ {
		for dx := -size / 2; dx <= size/2; dx++ {
			px := cx + dx
			py := cy + dy
			if px < bounds.Min.X || px >= bounds.Max.X || py < bounds.Min.Y || py >= bounds.Max.Y {
				continue
			}

			// For hostile expression, skip top row to make eyes narrow
			if expression == "hostile" && dy == -size/2 && size > 1 {
				continue
			}

			img.SetRGBA(px, py, eyeColor)
		}
	}

	// Specular highlight — 1px white dot at upper-left of eye
	if size >= 2 {
		hx := cx - size/2
		hy := cy - size/2
		if hx >= bounds.Min.X && hx < bounds.Max.X && hy >= bounds.Min.Y && hy < bounds.Max.Y {
			highlight := color.RGBA{R: 255, G: 255, B: 255, A: 180}
			img.SetRGBA(hx, hy, highlight)
		}
	}
}

// drawMouth renders a mouth line at the given position.
func drawMouth(img *image.RGBA, cx, cy, size int, mouthColor color.RGBA, expression string, bounds image.Rectangle) {
	// Mouth width depends on expression
	width := size
	switch expression {
	case "hostile":
		width = size + 1 // Wider grimace
		mouthColor = darkenRGBA(mouthColor, 0.8)
	case "friendly":
		width = size + 1 // Wider smile
	case "scared":
		// Round mouth — just a dot
		width = 1
	}

	halfW := width / 2
	for dx := -halfW; dx <= halfW; dx++ {
		px := cx + dx
		if px >= bounds.Min.X && px < bounds.Max.X && cy >= bounds.Min.Y && cy < bounds.Max.Y {
			img.SetRGBA(px, cy, mouthColor)
		}
	}
}

// expressionEyeColor adjusts eye color based on expression.
func expressionEyeColor(base color.RGBA, expression string) color.RGBA {
	switch expression {
	case "hostile":
		return darkenRGBA(base, 0.7)
	case "scared":
		return brightenRGBA(base, 1.3)
	default:
		return base
	}
}

// brightenRGBA returns a brighter version of a color, clamping at 255.
func brightenRGBA(c color.RGBA, factor float64) color.RGBA {
	clamp := func(v float64) uint8 {
		if v > 255 {
			return 255
		}
		if v < 0 {
			return 0
		}
		return uint8(v)
	}
	return color.RGBA{
		R: clamp(float64(c.R) * factor),
		G: clamp(float64(c.G) * factor),
		B: clamp(float64(c.B) * factor),
		A: c.A,
	}
}
