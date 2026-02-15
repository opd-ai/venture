// Package sprites provides seed-based hairstyle rendering for top-down aerial
// entity sprites. Each seed produces a distinct hairstyle that is drawn as a
// pixel overlay on top of the head body part, dramatically increasing visual
// variety between humanoid entities when viewed from above.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// HairStyle represents a distinct hairstyle type visible from top-down view.
type HairStyle int

const (
	// HairBald shows mostly scalp with thin fringe.
	HairBald HairStyle = iota
	// HairBuzz shows a uniform short stubble covering the head.
	HairBuzz
	// HairShort shows neat short hair slightly larger than the head.
	HairShort
	// HairMedium shows hair extending past head edges.
	HairMedium
	// HairLong shows flowing hair extending well past head in back.
	HairLong
	// HairPonytail shows a gathered tail extending in one direction.
	HairPonytail
	// HairMohawk shows a narrow ridge along the center.
	HairMohawk
	// HairSpiky shows short protrusions radiating from center.
	HairSpiky
	// HairTopKnot shows a small bun on top of the head.
	HairTopKnot
	// HairHooded shows a hood covering the head (rogues, mages).
	HairHooded
	// HairBraided shows two braids extending from the sides.
	HairBraided
	// HairStyleCount is the number of hair styles (must be last).
	HairStyleCount
)

// String returns the display name of a hair style.
func (h HairStyle) String() string {
	switch h {
	case HairBald:
		return "bald"
	case HairBuzz:
		return "buzz"
	case HairShort:
		return "short"
	case HairMedium:
		return "medium"
	case HairLong:
		return "long"
	case HairPonytail:
		return "ponytail"
	case HairMohawk:
		return "mohawk"
	case HairSpiky:
		return "spiky"
	case HairTopKnot:
		return "topknot"
	case HairHooded:
		return "hooded"
	case HairBraided:
		return "braided"
	default:
		return "unknown"
	}
}

// HairRenderParams contains parameters for rendering hair on a sprite.
type HairRenderParams struct {
	// HeadCenterX is the X center of the head in sprite pixels.
	HeadCenterX int
	// HeadCenterY is the Y center of the head in sprite pixels.
	HeadCenterY int
	// HeadRadius is the approximate radius of the head in pixels.
	HeadRadius int
	// Style is the hairstyle to render.
	Style HairStyle
	// Color is the hair base color.
	Color color.RGBA
	// Direction is the entity's facing direction.
	Direction Direction
	// Seed provides deterministic sub-variation within the style.
	Seed int64
}

// RenderHairOverlay draws a hairstyle overlay onto the given RGBA image.
// The hair is rendered relative to the head center, respecting direction.
// Deterministic: same params always produce the same pixel output.
func RenderHairOverlay(img *image.RGBA, params HairRenderParams) {
	if params.HeadRadius <= 0 {
		return
	}

	rng := rand.New(rand.NewSource(params.Seed))
	darker := darkenRGBA(params.Color, 0.7)
	lighter := lightenRGBA(params.Color, 1.3)

	switch params.Style {
	case HairBald:
		renderBaldHair(img, params, rng, darker)
	case HairBuzz:
		renderBuzzHair(img, params, rng, darker)
	case HairShort:
		renderShortHair(img, params, rng, darker, lighter)
	case HairMedium:
		renderMediumHair(img, params, rng, darker, lighter)
	case HairLong:
		renderLongHair(img, params, rng, darker, lighter)
	case HairPonytail:
		renderPonytailHair(img, params, rng, darker, lighter)
	case HairMohawk:
		renderMohawkHair(img, params, rng, darker, lighter)
	case HairSpiky:
		renderSpikyHair(img, params, rng, darker, lighter)
	case HairTopKnot:
		renderTopKnotHair(img, params, rng, darker, lighter)
	case HairHooded:
		renderHoodedHair(img, params, rng, darker)
	case HairBraided:
		renderBraidedHair(img, params, rng, darker, lighter)
	default:
		renderShortHair(img, params, rng, darker, lighter)
	}
}

// renderBaldHair renders a bald head with thin hair fringe around edges.
func renderBaldHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist >= float64(r)*0.85 && dist <= float64(r)*1.05 {
				// Thin fringe ring around edges with gaps
				angle := math.Atan2(float64(dy), float64(dx))
				noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed))
				if noise > 0.3 {
					setPixelSafe(img, cx+dx, cy+dy, darker)
				}
				_ = angle
			}
		}
	}
}

// renderBuzzHair renders uniform short stubble covering the head.
func renderBuzzHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(r)*0.95 {
				noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed))
				if noise > -0.2 {
					// Stippled pattern for buzz cut texture
					alpha := uint8(180 + int(noise*40))
					c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
					blendPixelSafe(img, cx+dx, cy+dy, c)
				}
			}
		}
	}
}

// renderShortHair renders neat short hair slightly larger than head.
func renderShortHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	outerR := float64(r) * 1.1

	for dy := -int(outerR) - 1; dy <= int(outerR)+1; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > outerR {
				continue
			}
			// Hair covers top half more than bottom
			yBias := float64(dy) / float64(r)
			coverage := outerR - yBias*float64(r)*0.3
			if dist > coverage {
				continue
			}
			// Shading: lighter at top-center, darker at edges
			shade := 1.0 - dist/outerR*0.3
			noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed)) * 0.08
			shade += noise
			c := shadeColor(p.Color, shade)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}
	// Part line (subtle darker line)
	partOffset := rng.Intn(3) - 1 // -1, 0, or 1 pixel offset
	for dy := -r; dy <= 0; dy++ {
		px := cx + partOffset
		py := cy + dy
		if math.Sqrt(float64(partOffset*partOffset+dy*dy)) < float64(r)*0.8 {
			blendPixelSafe(img, px, py, color.RGBA{R: darker.R, G: darker.G, B: darker.B, A: 120})
		}
	}
}

// renderMediumHair renders hair extending past head edges.
func renderMediumHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	outerR := float64(r) * 1.3

	// Direction offset for hair flow
	backDx, backDy := directionBackOffset(p.Direction)

	for dy := -int(outerR) - 1; dy <= int(outerR)+2; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			// Hair extends further on the back side
			backBias := float64(dx)*backDx + float64(dy)*backDy
			effR := outerR + backBias*float64(r)*0.2
			if dist > effR {
				continue
			}
			shade := 1.0 - dist/outerR*0.25
			noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed)) * 0.06
			shade += noise
			// Lighter highlight near top-front
			frontBias := -backBias
			if frontBias > 0 && dist < float64(r)*0.5 {
				shade += 0.08
			}
			c := shadeColor(p.Color, shade)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}
}

// renderLongHair renders flowing hair extending well past head.
func renderLongHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	backDx, backDy := directionBackOffset(p.Direction)

	// Main hair dome on top of head
	outerR := float64(r) * 1.15
	for dy := -int(outerR) - 1; dy <= int(outerR)+1; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > outerR {
				continue
			}
			shade := 1.0 - dist/outerR*0.2
			noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed)) * 0.05
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}

	// Long flowing extension in back direction
	tailLen := int(float64(r) * 1.8)
	tailWidth := int(float64(r) * 0.8)
	for i := 0; i < tailLen; i++ {
		progress := float64(i) / float64(tailLen)
		width := float64(tailWidth) * (1.0 - progress*0.5) // Taper slightly
		tx := cx + int(backDx*float64(i))
		ty := cy + int(backDy*float64(i))
		halfW := int(width / 2)
		for w := -halfW; w <= halfW; w++ {
			// Perpendicular to back direction
			px := tx + int(-backDy*float64(w))
			py := ty + int(backDx*float64(w))
			edgeDist := math.Abs(float64(w)) / width
			shade := 0.9 - edgeDist*0.2 - progress*0.15
			noise := ditherNoise(px, py, uint64(p.Seed)) * 0.06
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, px, py, c)
		}
	}
}

// renderPonytailHair renders a gathered tail extending away from facing.
func renderPonytailHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	backDx, backDy := directionBackOffset(p.Direction)

	// Base hair dome
	outerR := float64(r) * 1.05
	for dy := -int(outerR) - 1; dy <= int(outerR)+1; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > outerR {
				continue
			}
			shade := 1.0 - dist/outerR*0.2
			noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed)) * 0.05
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}

	// Gathered band (slightly darker)
	bandDist := int(float64(r) * 0.8)
	bx := cx + int(backDx*float64(bandDist))
	by := cy + int(backDy*float64(bandDist))
	for w := -1; w <= 1; w++ {
		px := bx + int(-backDy*float64(w))
		py := by + int(backDx*float64(w))
		setPixelSafe(img, px, py, darker)
	}

	// Tail extension (narrow, extends 1.5x head radius)
	tailLen := int(float64(r) * 1.5)
	for i := 0; i < tailLen; i++ {
		progress := float64(i) / float64(tailLen)
		width := math.Max(1, float64(r)*0.4*(1.0-progress*0.6))
		tx := bx + int(backDx*float64(i))
		ty := by + int(backDy*float64(i))
		halfW := int(width / 2)
		for w := -halfW; w <= halfW; w++ {
			px := tx + int(-backDy*float64(w))
			py := ty + int(backDx*float64(w))
			shade := 0.85 - progress*0.2
			noise := ditherNoise(px, py, uint64(p.Seed)) * 0.06
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, px, py, c)
		}
	}
}

// renderMohawkHair renders a narrow ridge along the head center.
func renderMohawkHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	frontDx, frontDy := directionFrontOffset(p.Direction)

	// Ridge runs perpendicular to facing direction
	ridgeLen := int(float64(r) * 2.2)
	ridgeWidth := maxInt(1, int(float64(r)*0.35))

	for i := -ridgeLen / 2; i <= ridgeLen/2; i++ {
		progress := math.Abs(float64(i)) / float64(ridgeLen/2)
		// Taper at ends
		curWidth := float64(ridgeWidth) * (1.0 - progress*0.5)
		// Ridge center runs perpendicular to front
		rx := cx + int(-frontDy*float64(i))
		ry := cy + int(frontDx*float64(i))

		halfW := int(curWidth / 2)
		for w := -halfW; w <= halfW; w++ {
			px := rx + int(frontDx*float64(w))
			py := ry + int(frontDy*float64(w))
			edgeFrac := math.Abs(float64(w)) / math.Max(1, curWidth/2)
			shade := 1.1 - edgeFrac*0.3 - progress*0.15
			noise := ditherNoise(px, py, uint64(p.Seed)) * 0.05
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, px, py, c)
		}
	}
}

// renderSpikyHair renders short protrusions radiating from head center.
func renderSpikyHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY

	// Base hair dome
	outerR := float64(r) * 1.0
	for dy := -int(outerR) - 1; dy <= int(outerR)+1; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > outerR {
				continue
			}
			shade := 1.0 - dist/outerR*0.2
			c := shadeColor(p.Color, shade)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}

	// Spikes radiating outward (5-8 spikes based on seed)
	numSpikes := 5 + rng.Intn(4)
	for i := 0; i < numSpikes; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numSpikes)
		angle += rng.Float64() * 0.3 // Slight random offset
		spikeLen := int(float64(r) * (0.6 + rng.Float64()*0.4))
		for s := 0; s < spikeLen; s++ {
			progress := float64(s) / float64(spikeLen)
			px := cx + int(math.Cos(angle)*float64(r)*0.8) + int(math.Cos(angle)*float64(s))
			py := cy + int(math.Sin(angle)*float64(r)*0.8) + int(math.Sin(angle)*float64(s))
			shade := 1.0 - progress*0.3
			c := shadeColor(p.Color, shade)
			setPixelSafe(img, px, py, c)
			// Width of spike decreases
			if progress < 0.5 {
				perpX := int(-math.Sin(angle))
				perpY := int(math.Cos(angle))
				setPixelSafe(img, px+perpX, py+perpY, shadeColor(p.Color, shade*0.9))
			}
		}
	}
}

// renderTopKnotHair renders a small bun gathered at the top.
func renderTopKnotHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY

	// Thin hair covering
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(r)*0.9 {
				noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed))
				shade := 0.85 + noise*0.05
				c := shadeColor(p.Color, shade)
				blendPixelSafe(img, cx+dx, cy+dy, color.RGBA{R: c.R, G: c.G, B: c.B, A: 200})
			}
		}
	}

	// Topknot bun — small circle slightly above center
	knotR := maxInt(2, int(float64(r)*0.4))
	knotY := cy - int(float64(r)*0.2)
	for dy := -knotR; dy <= knotR; dy++ {
		for dx := -knotR; dx <= knotR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(knotR) {
				shade := 1.1 - dist/float64(knotR)*0.3
				noise := ditherNoise(dx+cx, dy+knotY, uint64(p.Seed)) * 0.04
				c := shadeColor(p.Color, shade+noise)
				setPixelSafe(img, cx+dx, knotY+dy, c)
			}
		}
	}
}

// renderHoodedHair renders a hood covering the head.
func renderHoodedHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	_, frontDy := directionFrontOffset(p.Direction)

	// Hood is slightly larger than head, darker than hair color
	outerR := float64(r) * 1.2
	hoodColor := darkenRGBA(p.Color, 0.5)

	for dy := -int(outerR) - 1; dy <= int(outerR)+1; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			// Hood is more pointed in front
			effR := outerR
			if float64(dy)*frontDy > 0 {
				// Front side: slightly pointed
				effR -= math.Abs(float64(dx)) * 0.15
			}
			if dist > effR {
				continue
			}
			// Fold shading
			edgeFrac := dist / effR
			shade := 0.9 - edgeFrac*0.2
			noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed)) * 0.08
			shade += noise
			c := shadeColor(hoodColor, shade)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}

	// Hood opening (darker inner area suggesting face shadow)
	openR := float64(r) * 0.55
	for dy := -int(openR); dy <= int(openR); dy++ {
		for dx := -int(openR); dx <= int(openR); dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > openR {
				continue
			}
			// Only show opening on front-facing side
			if float64(dy)*frontDy < -float64(r)*0.1 {
				continue
			}
			shadow := color.RGBA{R: 20, G: 15, B: 15, A: 140}
			blendPixelSafe(img, cx+dx, cy+dy, shadow)
		}
	}
}

// renderBraidedHair renders two braids extending from the sides.
func renderBraidedHair(img *image.RGBA, p HairRenderParams, rng *rand.Rand, darker, lighter color.RGBA) {
	r := p.HeadRadius
	cx, cy := p.HeadCenterX, p.HeadCenterY
	backDx, backDy := directionBackOffset(p.Direction)

	// Base hair dome
	outerR := float64(r) * 1.08
	for dy := -int(outerR) - 1; dy <= int(outerR)+1; dy++ {
		for dx := -int(outerR) - 1; dx <= int(outerR)+1; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > outerR {
				continue
			}
			shade := 1.0 - dist/outerR*0.2
			noise := ditherNoise(dx+cx, dy+cy, uint64(p.Seed)) * 0.05
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, cx+dx, cy+dy, c)
		}
	}

	// Two braids extending in back direction, offset to sides
	perpDx, perpDy := -backDy, backDx
	braidLen := int(float64(r) * 1.4)
	for side := -1; side <= 1; side += 2 {
		offset := float64(side) * float64(r) * 0.5
		startX := cx + int(perpDx*offset)
		startY := cy + int(perpDy*offset)

		for i := 0; i < braidLen; i++ {
			progress := float64(i) / float64(braidLen)
			// Braid weave pattern: alternate slight sideways offsets
			weave := math.Sin(float64(i)*1.5) * 1.0
			bx := startX + int(backDx*float64(i)) + int(perpDx*weave)
			by := startY + int(backDy*float64(i)) + int(perpDy*weave)

			shade := 0.9 - progress*0.2
			noise := ditherNoise(bx, by, uint64(p.Seed)) * 0.06
			c := shadeColor(p.Color, shade+noise)
			setPixelSafe(img, bx, by, c)
			// Width of 2px for braids
			setPixelSafe(img, bx+int(perpDx), by+int(perpDy), shadeColor(p.Color, shade*0.85+noise))
		}
	}
}

// directionBackOffset returns a unit vector pointing away from the facing direction.
func directionBackOffset(dir Direction) (float64, float64) {
	switch dir {
	case DirUp:
		return 0, 1 // Back is down
	case DirDown:
		return 0, -1 // Back is up
	case DirLeft:
		return 1, 0 // Back is right
	case DirRight:
		return -1, 0 // Back is left
	default:
		return 0, 1
	}
}

// directionFrontOffset returns a unit vector pointing in the facing direction.
func directionFrontOffset(dir Direction) (float64, float64) {
	dx, dy := directionBackOffset(dir)
	return -dx, -dy
}

// setPixelSafe writes a pixel if the coordinates are within image bounds.
func setPixelSafe(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
	img.Pix[idx] = c.R
	img.Pix[idx+1] = c.G
	img.Pix[idx+2] = c.B
	img.Pix[idx+3] = c.A
}

// blendPixelSafe alpha-blends a color onto the existing pixel.
func blendPixelSafe(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
	srcA := float64(c.A) / 255.0
	dstA := 1.0 - srcA
	img.Pix[idx] = clampByte(float64(c.R)*srcA + float64(img.Pix[idx])*dstA)
	img.Pix[idx+1] = clampByte(float64(c.G)*srcA + float64(img.Pix[idx+1])*dstA)
	img.Pix[idx+2] = clampByte(float64(c.B)*srcA + float64(img.Pix[idx+2])*dstA)
	if img.Pix[idx+3] < c.A {
		img.Pix[idx+3] = c.A
	}
}

// shadeColor adjusts a color's brightness by a factor (>1 = lighter, <1 = darker).
func shadeColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: clampByte(float64(c.R) * factor),
		G: clampByte(float64(c.G) * factor),
		B: clampByte(float64(c.B) * factor),
		A: c.A,
	}
}

// lightenRGBA returns a lighter version of the color.
func lightenRGBA(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: clampByte(float64(c.R) * factor),
		G: clampByte(float64(c.G) * factor),
		B: clampByte(float64(c.B) * factor),
		A: c.A,
	}
}

// maxInt returns the larger of two ints.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ComputeHairRenderParams calculates hair rendering parameters from sprite
// dimensions and template head specification.
func ComputeHairRenderParams(spriteWidth, spriteHeight int, headSpec PartSpec, traits *AvatarTraits, direction Direction, seed int64) HairRenderParams {
	partWidth := int(float64(spriteWidth) * headSpec.RelativeWidth)
	partHeight := int(float64(spriteHeight) * headSpec.RelativeHeight)
	if headSpec.PreferredPixelSize != nil {
		partWidth = headSpec.PreferredPixelSize.Width
		partHeight = headSpec.PreferredPixelSize.Height
	}

	headCX := int(float64(spriteWidth) * headSpec.RelativeX)
	headCY := int(float64(spriteHeight) * headSpec.RelativeY)

	radius := partWidth / 2
	if partHeight/2 < radius {
		radius = partHeight / 2
	}
	if radius < 1 {
		radius = 1
	}

	return HairRenderParams{
		HeadCenterX: headCX,
		HeadCenterY: headCY,
		HeadRadius:  radius,
		Style:       traits.HairStyle,
		Color:       traits.HairColor,
		Direction:   direction,
		Seed:        seed,
	}
}
