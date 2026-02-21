// Package sprites provides seed-based creature markings for procedurally
// generated nonhumanoid entity sprites. Markings (spots, stripes, patches,
// gradients, rings, scales) are applied after the base body parts are rendered,
// making each creature visually unique and immediately distinguishable from
// others of the same type.
//
// All rendering is seed-deterministic and operates on standard image.RGBA
// buffers for compositing onto Ebiten sprites.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// CreatureMarkingType identifies the visual marking pattern applied to a creature.
type CreatureMarkingType int

const (
	// MarkingNone applies no markings (solid color).
	MarkingNone CreatureMarkingType = iota
	// MarkingSpots draws scattered circular spots.
	MarkingSpots
	// MarkingStripes draws parallel stripe bands.
	MarkingStripes
	// MarkingPatches draws irregular blotchy patches.
	MarkingPatches
	// MarkingRings draws concentric ring patterns (for serpents/blobs).
	MarkingRings
	// MarkingScales draws overlapping scale patterns (for serpents/dragons).
	MarkingScales
	// MarkingGradient draws a body-length color gradient.
	MarkingGradient
	// MarkingDappled draws small dappled light/dark spots like a fawn.
	MarkingDappled
	// MarkingBanded draws thick alternating color bands.
	MarkingBanded
	// MarkingMottled draws irregular splotchy coloring.
	MarkingMottled
	// MarkingTiger draws diagonal stripe patterns like a tiger.
	MarkingTiger
	// MarkingLeopard draws rosette patterns like a leopard.
	MarkingLeopard
	// MarkingZebra draws wavy zebra-like stripes.
	MarkingZebra
	// markingCount is the total number of marking types.
	markingCount
)

// String returns a human-readable name for the marking type.
func (m CreatureMarkingType) String() string {
	names := []string{
		"none", "spots", "stripes", "patches", "rings", "scales",
		"gradient", "dappled", "banded", "mottled", "tiger", "leopard", "zebra",
	}
	if int(m) < len(names) {
		return names[m]
	}
	return "unknown"
}

// CreatureMarkings holds parameters for a creature's visual markings.
type CreatureMarkings struct {
	// Type is the primary marking pattern.
	Type CreatureMarkingType
	// SecondaryType is an optional secondary pattern (layered on top).
	SecondaryType CreatureMarkingType
	// PrimaryColor is the main marking color.
	PrimaryColor color.RGBA
	// SecondaryColor is used for secondary patterns or highlights.
	SecondaryColor color.RGBA
	// Density controls how many marks are drawn (0.0-1.0).
	Density float64
	// Scale controls the size of individual marks (0.5-2.0).
	Scale float64
	// Intensity controls mark visibility/opacity (0.0-1.0).
	Intensity float64
	// Rotation offsets the pattern angle in degrees.
	Rotation float64
}

// CreatureMarkingParams contains all info needed to render markings.
type CreatureMarkingParams struct {
	Width, Height int
	Form          string // creature form key (quadruped, serpentine, etc.)
	Direction     string
	Seed          int64
	Markings      CreatureMarkings
}

// GenerateCreatureMarkings deterministically produces markings from seed and creature type.
// Different creature forms have different probabilities for each marking type.
func GenerateCreatureMarkings(seed int64, creatureForm string) CreatureMarkings {
	rng := rand.New(rand.NewSource(seed ^ 0x4D41524B)) // "MARK" XOR

	// Base chance of having any markings (most creatures should)
	if rng.Float64() > 0.85 {
		return CreatureMarkings{Type: MarkingNone}
	}

	markings := CreatureMarkings{
		Density:   0.3 + rng.Float64()*0.5,
		Scale:     0.7 + rng.Float64()*0.8,
		Intensity: 0.4 + rng.Float64()*0.5,
		Rotation:  rng.Float64() * 30,
	}

	// Select marking type based on creature form
	markings.Type = selectMarkingForForm(creatureForm, rng)

	// 25% chance of secondary markings
	if rng.Float64() < 0.25 {
		markings.SecondaryType = selectMarkingForForm(creatureForm, rng)
		if markings.SecondaryType == markings.Type {
			markings.SecondaryType = MarkingNone
		}
	}

	// Generate colors based on creature form
	markings.PrimaryColor, markings.SecondaryColor = generateMarkingColors(creatureForm, rng)

	return markings
}

// selectMarkingForForm chooses appropriate marking types for each creature form.
func selectMarkingForForm(form string, rng *rand.Rand) CreatureMarkingType {
	var weights []float64
	switch form {
	case "quadruped":
		// Natural animal patterns: spots, stripes, dappled, tiger, leopard
		weights = []float64{0, 0.15, 0.15, 0.1, 0.05, 0.02, 0.08, 0.12, 0.05, 0.08, 0.10, 0.08, 0.02}
	case "serpentine":
		// Snake patterns: rings, scales, banded, gradient
		weights = []float64{0, 0.05, 0.12, 0.05, 0.20, 0.18, 0.15, 0.05, 0.15, 0.03, 0.02, 0, 0}
	case "arachnid":
		// Spider patterns: spots, patches, mottled
		weights = []float64{0, 0.25, 0.05, 0.20, 0.05, 0, 0.10, 0.10, 0.05, 0.18, 0.02, 0, 0}
	case "insect":
		// Insect patterns: banded, stripes, spots
		weights = []float64{0, 0.18, 0.15, 0.08, 0.05, 0.05, 0.08, 0.08, 0.20, 0.08, 0.05, 0, 0}
	case "flying":
		// Dragon/bird patterns: scales, gradient, patches
		weights = []float64{0, 0.10, 0.08, 0.15, 0.05, 0.20, 0.15, 0.05, 0.10, 0.07, 0.05, 0, 0}
	case "blob":
		// Slime patterns: gradient, mottled, rings
		weights = []float64{0, 0.05, 0.02, 0.08, 0.20, 0, 0.30, 0.05, 0.05, 0.20, 0.05, 0, 0}
	case "mechanical":
		// Robot patterns: banded (warning stripes), gradient
		weights = []float64{0, 0.05, 0.25, 0.08, 0, 0, 0.20, 0, 0.35, 0.02, 0.05, 0, 0}
	case "undead":
		// Undead patterns: mottled, patches, dappled (decay)
		weights = []float64{0, 0.08, 0.05, 0.20, 0.05, 0, 0.10, 0.12, 0.05, 0.30, 0.05, 0, 0}
	case "multi_limbed":
		// Eldritch patterns: mottled, patches, rings, spots
		weights = []float64{0, 0.15, 0.05, 0.18, 0.12, 0.05, 0.10, 0.08, 0.07, 0.15, 0.05, 0, 0}
	default:
		// Generic: equal distribution
		weights = make([]float64, int(markingCount))
		for i := 1; i < int(markingCount); i++ {
			weights[i] = 1.0 / float64(markingCount-1)
		}
	}

	// Weighted random selection
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := rng.Float64() * total
	cumulative := 0.0
	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return CreatureMarkingType(i)
		}
	}
	return MarkingSpots
}

// generateMarkingColors produces appropriate marking colors for each form.
func generateMarkingColors(form string, rng *rand.Rand) (primary, secondary color.RGBA) {
	var hueRange [2]float64
	var satRange [2]float64
	var lumRange [2]float64

	switch form {
	case "quadruped":
		hueRange = [2]float64{20, 60} // browns/oranges
		satRange = [2]float64{0.3, 0.6}
		lumRange = [2]float64{0.2, 0.4}
	case "serpentine":
		hueRange = [2]float64{60, 150} // greens/yellows
		satRange = [2]float64{0.4, 0.7}
		lumRange = [2]float64{0.3, 0.5}
	case "arachnid":
		hueRange = [2]float64{10, 50} // dark browns/reds
		satRange = [2]float64{0.2, 0.5}
		lumRange = [2]float64{0.15, 0.35}
	case "insect":
		hueRange = [2]float64{40, 80} // yellows/oranges
		satRange = [2]float64{0.5, 0.8}
		lumRange = [2]float64{0.4, 0.6}
	case "flying":
		hueRange = [2]float64{0, 360} // any color
		satRange = [2]float64{0.5, 0.8}
		lumRange = [2]float64{0.3, 0.6}
	case "blob":
		hueRange = [2]float64{90, 280} // greens to purples
		satRange = [2]float64{0.4, 0.7}
		lumRange = [2]float64{0.4, 0.6}
	case "mechanical":
		hueRange = [2]float64{40, 60} // warning yellow/orange
		satRange = [2]float64{0.7, 0.9}
		lumRange = [2]float64{0.5, 0.7}
	case "undead":
		hueRange = [2]float64{80, 140} // sickly greens
		satRange = [2]float64{0.15, 0.35}
		lumRange = [2]float64{0.2, 0.4}
	case "multi_limbed":
		hueRange = [2]float64{260, 320} // purples/magentas
		satRange = [2]float64{0.3, 0.6}
		lumRange = [2]float64{0.25, 0.45}
	default:
		hueRange = [2]float64{0, 360}
		satRange = [2]float64{0.3, 0.6}
		lumRange = [2]float64{0.3, 0.5}
	}

	h1 := hueRange[0] + rng.Float64()*(hueRange[1]-hueRange[0])
	s1 := satRange[0] + rng.Float64()*(satRange[1]-satRange[0])
	l1 := lumRange[0] + rng.Float64()*(lumRange[1]-lumRange[0])
	primary = hslToRGBA(h1, s1, l1)

	// Secondary is offset in hue or brightness
	if rng.Float64() < 0.5 {
		h2 := math.Mod(h1+30+rng.Float64()*60, 360)
		secondary = hslToRGBA(h2, s1*0.9, l1+0.1)
	} else {
		secondary = hslToRGBA(h1, s1*0.8, l1*0.6)
	}

	return primary, secondary
}

// RenderCreatureMarkings applies markings to an existing creature sprite buffer.
// Call this after base template parts and creature details have been rendered.
func RenderCreatureMarkings(buf *image.RGBA, params CreatureMarkingParams) {
	if buf == nil || params.Width <= 0 || params.Height <= 0 {
		return
	}
	if params.Markings.Type == MarkingNone {
		return
	}

	rng := rand.New(rand.NewSource(params.Seed * 8191))

	// Render primary markings
	renderMarkingPattern(buf, params, params.Markings.Type, params.Markings.PrimaryColor, rng)

	// Render secondary markings if present
	if params.Markings.SecondaryType != MarkingNone {
		// Secondary uses lower intensity
		secondaryParams := params
		secondaryParams.Markings.Intensity *= 0.6
		renderMarkingPattern(buf, secondaryParams, params.Markings.SecondaryType, params.Markings.SecondaryColor, rng)
	}
}

// renderMarkingPattern draws a specific marking type onto the buffer.
func renderMarkingPattern(buf *image.RGBA, params CreatureMarkingParams, markType CreatureMarkingType, markColor color.RGBA, rng *rand.Rand) {
	switch markType {
	case MarkingSpots:
		renderSpots(buf, params, markColor, rng)
	case MarkingStripes:
		renderStripes(buf, params, markColor, rng)
	case MarkingPatches:
		renderPatches(buf, params, markColor, rng)
	case MarkingRings:
		renderRings(buf, params, markColor, rng)
	case MarkingScales:
		renderScales(buf, params, markColor, rng)
	case MarkingGradient:
		renderMarkingGradient(buf, params, markColor, rng)
	case MarkingDappled:
		renderDappled(buf, params, markColor, rng)
	case MarkingBanded:
		renderBanded(buf, params, markColor, rng)
	case MarkingMottled:
		renderMottled(buf, params, markColor, rng)
	case MarkingTiger:
		renderTiger(buf, params, markColor, rng)
	case MarkingLeopard:
		renderLeopard(buf, params, markColor, rng)
	case MarkingZebra:
		renderZebra(buf, params, markColor, rng)
	}
}

// renderSpots draws scattered circular spots.
func renderSpots(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	numSpots := int(float64(w*h) * params.Markings.Density * 0.02)
	spotSize := int(float64(w) * params.Markings.Scale * 0.15)
	if spotSize < 1 {
		spotSize = 1
	}

	for i := 0; i < numSpots; i++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)

		// Only draw on existing opaque pixels
		if !isOpaqueAt(buf, cx, cy) {
			continue
		}

		// Draw filled circle
		for dy := -spotSize; dy <= spotSize; dy++ {
			for dx := -spotSize; dx <= spotSize; dx++ {
				if dx*dx+dy*dy <= spotSize*spotSize {
					px, py := cx+dx, cy+dy
					if isOpaqueAt(buf, px, py) {
						// Apply with intensity-based alpha
						spotC := c
						spotC.A = uint8(float64(c.A) * params.Markings.Intensity)
						blendPixelSafe(buf, px, py, spotC)
					}
				}
			}
		}
	}
}

// renderStripes draws parallel stripe bands.
func renderStripes(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	stripeWidth := int(float64(w) * params.Markings.Scale * 0.12)
	if stripeWidth < 1 {
		stripeWidth = 1
	}
	stripeGap := int(float64(stripeWidth) * (1.5 + rng.Float64()))

	// Stripe direction based on creature form
	vertical := params.Form == "serpentine" || rng.Float64() < 0.5
	angle := params.Markings.Rotation * math.Pi / 180

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			var pos int
			if vertical {
				pos = int(float64(x)*math.Cos(angle) + float64(y)*math.Sin(angle))
			} else {
				pos = int(float64(y)*math.Cos(angle) + float64(x)*math.Sin(angle))
			}

			if (pos/(stripeWidth+stripeGap))%2 == 0 && pos%(stripeWidth+stripeGap) < stripeWidth {
				stripeC := c
				stripeC.A = uint8(float64(c.A) * params.Markings.Intensity)
				blendPixelSafe(buf, x, y, stripeC)
			}
		}
	}
}

// renderPatches draws irregular blotchy patches using noise-based shapes.
func renderPatches(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	numPatches := 2 + int(float64(w)*params.Markings.Density*0.2)

	for p := 0; p < numPatches; p++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)
		patchSize := int(float64(w) * params.Markings.Scale * (0.15 + rng.Float64()*0.15))

		// Draw irregular patch using random walk
		for i := 0; i < patchSize*3; i++ {
			px := cx + rng.Intn(patchSize*2) - patchSize
			py := cy + rng.Intn(patchSize*2) - patchSize

			// Bias toward center
			dist := math.Sqrt(float64((px-cx)*(px-cx) + (py-cy)*(py-cy)))
			if dist > float64(patchSize)*1.2 {
				continue
			}

			if isOpaqueAt(buf, px, py) {
				// Fade at edges
				edgeFade := 1.0 - (dist / float64(patchSize) * 0.5)
				patchC := c
				patchC.A = uint8(float64(c.A) * params.Markings.Intensity * edgeFade)
				blendPixelSafe(buf, px, py, patchC)
			}
		}
	}
}

// renderRings draws concentric ring patterns.
func renderRings(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	cx, cy := w/2, h/2
	ringSpacing := int(float64(w) * params.Markings.Scale * 0.15)
	if ringSpacing < 2 {
		ringSpacing = 2
	}
	ringWidth := max(1, ringSpacing/3)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			dist := int(math.Sqrt(float64((x-cx)*(x-cx) + (y-cy)*(y-cy))))
			if dist%ringSpacing < ringWidth {
				ringC := c
				ringC.A = uint8(float64(c.A) * params.Markings.Intensity)
				blendPixelSafe(buf, x, y, ringC)
			}
		}
	}
}

// renderScales draws overlapping scale patterns.
func renderScales(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	scaleSize := int(float64(w) * params.Markings.Scale * 0.10)
	if scaleSize < 2 {
		scaleSize = 2
	}

	// Draw overlapping semi-circles in a grid offset pattern
	for row := 0; row < h/scaleSize+1; row++ {
		offset := 0
		if row%2 == 1 {
			offset = scaleSize / 2
		}
		for col := -1; col < w/scaleSize+2; col++ {
			cx := col*scaleSize + offset
			cy := row * scaleSize

			// Draw scale edge (top arc)
			for angle := 0.0; angle < math.Pi; angle += 0.2 {
				px := cx + int(float64(scaleSize/2)*math.Cos(angle))
				py := cy - int(float64(scaleSize/2)*math.Sin(angle))
				if isOpaqueAt(buf, px, py) {
					scaleC := c
					scaleC.A = uint8(float64(c.A) * params.Markings.Intensity)
					blendPixelSafe(buf, px, py, scaleC)
				}
			}
		}
	}
}

// renderMarkingGradient draws a body-length color gradient.
func renderMarkingGradient(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	vertical := params.Form != "serpentine" || rng.Float64() < 0.5

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			var t float64
			if vertical {
				t = float64(y) / float64(h)
			} else {
				t = float64(x) / float64(w)
			}

			// Apply gradient as darkening/lightening
			gradC := c
			gradC.A = uint8(float64(c.A) * params.Markings.Intensity * t)
			blendPixelSafe(buf, x, y, gradC)
		}
	}
}

// renderDappled draws small dappled light/dark spots.
func renderDappled(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	numDapples := int(float64(w*h) * params.Markings.Density * 0.05)

	for i := 0; i < numDapples; i++ {
		x := rng.Intn(w)
		y := rng.Intn(h)

		if !isOpaqueAt(buf, x, y) {
			continue
		}

		// Small 1-2 pixel dapple
		dappleC := c
		dappleC.A = uint8(float64(c.A) * params.Markings.Intensity * (0.5 + rng.Float64()*0.5))
		setPixelSafe(buf, x, y, dappleC)
		if rng.Float64() < 0.5 {
			setPixelSafe(buf, x+1, y, dappleC)
		}
	}
}

// renderBanded draws thick alternating color bands.
func renderBanded(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	bandWidth := int(float64(w) * params.Markings.Scale * 0.20)
	if bandWidth < 2 {
		bandWidth = 2
	}

	vertical := params.Form == "serpentine"

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			var pos int
			if vertical {
				pos = x
			} else {
				pos = y
			}

			if (pos/bandWidth)%2 == 0 {
				bandC := c
				bandC.A = uint8(float64(c.A) * params.Markings.Intensity)
				blendPixelSafe(buf, x, y, bandC)
			}
		}
	}
}

// renderMottled draws irregular splotchy coloring using Perlin-like noise.
func renderMottled(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	noiseScale := params.Markings.Scale * 0.2

	// Generate noise field
	noiseOffsetX := rng.Float64() * 1000
	noiseOffsetY := rng.Float64() * 1000

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			// Simple pseudo-noise based on coordinates
			nx := (float64(x)*noiseScale + noiseOffsetX)
			ny := (float64(y)*noiseScale + noiseOffsetY)
			noise := math.Sin(nx)*math.Cos(ny)*0.5 + 0.5

			if noise > 0.5-params.Markings.Density*0.3 {
				mottleC := c
				mottleC.A = uint8(float64(c.A) * params.Markings.Intensity * noise)
				blendPixelSafe(buf, x, y, mottleC)
			}
		}
	}
}

// renderTiger draws diagonal stripe patterns like a tiger.
func renderTiger(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	stripeWidth := int(float64(w) * params.Markings.Scale * 0.08)
	if stripeWidth < 1 {
		stripeWidth = 1
	}
	stripeGap := stripeWidth * 2

	angle := params.Markings.Rotation + 45

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			// Diagonal stripe calculation with waviness
			pos := int(float64(x)*math.Cos(angle*math.Pi/180) + float64(y)*math.Sin(angle*math.Pi/180))
			waveOffset := int(math.Sin(float64(y)*0.3) * float64(stripeWidth) * 0.5)
			pos += waveOffset

			if pos%(stripeWidth+stripeGap) < stripeWidth {
				tigerC := c
				tigerC.A = uint8(float64(c.A) * params.Markings.Intensity)
				blendPixelSafe(buf, x, y, tigerC)
			}
		}
	}
}

// renderLeopard draws rosette patterns like a leopard.
func renderLeopard(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	numRosettes := int(float64(w*h) * params.Markings.Density * 0.015)
	rosetteSize := int(float64(w) * params.Markings.Scale * 0.15)
	if rosetteSize < 2 {
		rosetteSize = 2
	}

	for i := 0; i < numRosettes; i++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)

		if !isOpaqueAt(buf, cx, cy) {
			continue
		}

		// Draw rosette: darker ring around lighter center
		innerR := rosetteSize / 2
		outerR := rosetteSize

		for dy := -outerR; dy <= outerR; dy++ {
			for dx := -outerR; dx <= outerR; dx++ {
				dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
				px, py := cx+dx, cy+dy

				if !isOpaqueAt(buf, px, py) {
					continue
				}

				// Outer ring
				if dist >= innerR && dist <= outerR {
					rosetteC := c
					rosetteC.A = uint8(float64(c.A) * params.Markings.Intensity)
					blendPixelSafe(buf, px, py, rosetteC)
				}
			}
		}
	}
}

// renderZebra draws wavy zebra-like stripes.
func renderZebra(buf *image.RGBA, params CreatureMarkingParams, c color.RGBA, rng *rand.Rand) {
	w, h := params.Width, params.Height
	stripeWidth := int(float64(w) * params.Markings.Scale * 0.10)
	if stripeWidth < 1 {
		stripeWidth = 1
	}
	stripeGap := stripeWidth

	waveAmp := float64(stripeWidth) * 1.5
	waveFreq := 0.2 + rng.Float64()*0.2

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isOpaqueAt(buf, x, y) {
				continue
			}

			// Vertical stripes with horizontal wave
			wave := int(waveAmp * math.Sin(float64(y)*waveFreq))
			pos := x + wave

			if pos%(stripeWidth+stripeGap) < stripeWidth {
				zebraC := c
				zebraC.A = uint8(float64(c.A) * params.Markings.Intensity)
				blendPixelSafe(buf, x, y, zebraC)
			}
		}
	}
}
