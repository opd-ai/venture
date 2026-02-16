// Package sprites provides form-specific detail rendering for nonhumanoid
// creatures. After the base body-part shapes are drawn by the template
// renderer, this module adds pixel-level features — eyes, leg joints,
// scale patterns, wing membranes, etc. — that make each creature form
// immediately recognizable from the top-down aerial perspective.
//
// All rendering is seed-deterministic and operates on standard image.RGBA
// buffers so it can be composited onto the Ebiten sprite after generation.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// CreatureDetailParams configures creature detail rendering.
type CreatureDetailParams struct {
	Width, Height int
	Form          string // creature form key matching CreatureForm values
	Direction     string // facing direction (DirUp, DirDown, DirLeft, DirRight)
	Seed          int64
	SizeClass     string // "tiny", "small", "medium", "large", "huge"
	Genre         string
}

// RenderCreatureDetails draws form-specific visual details onto buf.
// Call this after base template parts have been rendered to add eyes,
// scale patterns, wing lines, and other recognizable creature features.
// Non-creature forms (humanoid, empty) are silently skipped.
func RenderCreatureDetails(buf *image.RGBA, params CreatureDetailParams) {
	if buf == nil || params.Width <= 0 || params.Height <= 0 {
		return
	}
	rng := rand.New(rand.NewSource(params.Seed * 7919))
	switch params.Form {
	case "quadruped":
		renderQuadrupedDetails(buf, params, rng)
	case "arachnid":
		renderArachnidDetails(buf, params, rng)
	case "serpentine":
		renderSerpentineDetails(buf, params, rng)
	case "flying":
		renderFlyingDetails(buf, params, rng)
	case "blob":
		renderBlobDetails(buf, params, rng)
	case "mechanical":
		renderMechanicalDetails(buf, params, rng)
	case "undead":
		renderUndeadDetails(buf, params, rng)
	case "insect":
		renderInsectDetails(buf, params, rng)
	case "multi_limbed":
		renderMultiLimbedDetails(buf, params, rng)
	}
}

// ----------------------------------------------------------------------------
// Quadruped details — ears, paw pads, tail, fur dither
// ----------------------------------------------------------------------------

func renderQuadrupedDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	// Head center depends on facing direction
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Ears: two small triangular tufts flanking the head ---
	earColor := color.RGBA{R: 180, G: 130, B: 90, A: 220}
	earInner := color.RGBA{R: 210, G: 160, B: 140, A: 200}
	earW := max(2, w/8)
	// Left ear
	setPixelSafe(buf, hx-earW, hy-2, earColor)
	setPixelSafe(buf, hx-earW, hy-1, earInner)
	setPixelSafe(buf, hx-earW-1, hy-1, earColor)
	// Right ear
	setPixelSafe(buf, hx+earW, hy-2, earColor)
	setPixelSafe(buf, hx+earW, hy-1, earInner)
	setPixelSafe(buf, hx+earW+1, hy-1, earColor)

	// --- Eyes: two bright dots on the head ---
	eyeColor := color.RGBA{R: 30, G: 30, B: 30, A: 255}
	eyeShine := color.RGBA{R: 220, G: 220, B: 240, A: 180}
	setPixelSafe(buf, hx-1, hy, eyeColor)
	setPixelSafe(buf, hx+1, hy, eyeColor)
	setPixelSafe(buf, hx-1, hy-1, eyeShine) // small highlight

	// --- Nose dot ---
	noseColor := color.RGBA{R: 60, G: 40, B: 40, A: 230}
	nx, ny := noseOffset(hx, hy, p.Direction)
	setPixelSafe(buf, nx, ny, noseColor)

	// --- Paw pads: four dots at each leg position ---
	pawColor := color.RGBA{R: 160, G: 120, B: 100, A: 200}
	legSpreadX := max(3, w*3/8)
	legSpreadY := max(2, h/5)
	pawPositions := [4][2]int{
		{cx - legSpreadX, cy - legSpreadY},     // front-left
		{cx + legSpreadX, cy - legSpreadY},     // front-right
		{cx - legSpreadX, cy + legSpreadY + 1}, // rear-left
		{cx + legSpreadX, cy + legSpreadY + 1}, // rear-right
	}
	for _, pos := range pawPositions {
		setPixelSafe(buf, pos[0], pos[1], pawColor)
		setPixelSafe(buf, pos[0]+1, pos[1], pawColor)
	}

	// --- Tail: short line extending from rear ---
	tailColor := color.RGBA{R: 150 + uint8(rng.Intn(40)), G: 110 + uint8(rng.Intn(30)), B: 80, A: 210}
	tx, ty := tailOffset(cx, cy, p.Direction, w, h)
	tailLen := max(2, h/6)
	for i := 0; i < tailLen; i++ {
		dx, dy := tailDirection(p.Direction, i)
		setPixelSafe(buf, tx+dx, ty+dy, tailColor)
	}

	// --- Fur dithering: subtle diagonal stripes across the body ---
	furHighlight := color.RGBA{R: 200, G: 170, B: 140, A: 50}
	for y := cy - h/4; y < cy+h/4; y++ {
		for x := cx - w/4; x < cx+w/4; x++ {
			if (x+y+int(p.Seed))%5 == 0 && isOpaqueAt(buf, x, y) {
				blendPixelSafe(buf, x, y, furHighlight)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Arachnid details — multi-eye cluster, leg joints, mandibles, web pattern
// ----------------------------------------------------------------------------

func renderArachnidDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Multiple eyes: cluster of 6-8 dots ---
	numEyes := 6 + rng.Intn(3)
	eyeColor := color.RGBA{R: 200, G: 20, B: 20, A: 240}
	eyeRadius := max(1, w/10)
	for i := 0; i < numEyes; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numEyes)
		ex := hx + int(float64(eyeRadius)*math.Cos(angle))
		ey := hy + int(float64(eyeRadius)*math.Sin(angle))
		setPixelSafe(buf, ex, ey, eyeColor)
	}
	// Central large eye
	centralEye := color.RGBA{R: 240, G: 30, B: 10, A: 255}
	setPixelSafe(buf, hx, hy, centralEye)

	// --- Mandibles: V-shape in front of head ---
	mandibleColor := color.RGBA{R: 80, G: 60, B: 40, A: 230}
	mx, my := noseOffset(hx, hy, p.Direction)
	dx1, dy1 := mandibleDir(p.Direction, -1)
	dx2, dy2 := mandibleDir(p.Direction, 1)
	for i := 0; i < 3; i++ {
		setPixelSafe(buf, mx+dx1*i, my+dy1*i, mandibleColor)
		setPixelSafe(buf, mx+dx2*i, my+dy2*i, mandibleColor)
	}

	// --- Leg joint highlights: bright dots along each leg ---
	jointColor := color.RGBA{R: 120, G: 100, B: 80, A: 180}
	jointHighlight := color.RGBA{R: 180, G: 160, B: 130, A: 120}
	legCount := 8
	for i := 0; i < legCount; i++ {
		angle := float64(i) * 2 * math.Pi / float64(legCount)
		for j := 1; j <= 3; j++ {
			radius := float64(max(3, w/4)) * float64(j) / 3.0
			jx := cx + int(radius*math.Cos(angle))
			jy := cy + int(radius*math.Sin(angle))
			setPixelSafe(buf, jx, jy, jointColor)
			if j == 2 {
				setPixelSafe(buf, jx, jy, jointHighlight)
			}
		}
	}

	// --- Abdomen pattern: subtle chevron/dots on rear body ---
	patternColor := color.RGBA{R: 100, G: 70, B: 50, A: 80}
	abdX, abdY := tailOffset(cx, cy, p.Direction, w, h)
	for i := -2; i <= 2; i++ {
		setPixelSafe(buf, abdX+i, abdY-abs(i), patternColor)
		setPixelSafe(buf, abdX+i, abdY+1-abs(i), patternColor)
	}
}

// ----------------------------------------------------------------------------
// Serpentine details — scale chevrons, forked tongue, eye slits
// ----------------------------------------------------------------------------

func renderSerpentineDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Eye slits: narrow vertical lines ---
	eyeColor := color.RGBA{R: 220, G: 200, B: 30, A: 250}
	pupilColor := color.RGBA{R: 30, G: 30, B: 10, A: 255}
	setPixelSafe(buf, hx-2, hy, eyeColor)
	setPixelSafe(buf, hx-2, hy-1, pupilColor)
	setPixelSafe(buf, hx+2, hy, eyeColor)
	setPixelSafe(buf, hx+2, hy-1, pupilColor)

	// --- Forked tongue: two-pixel fork extending from head ---
	tongueColor := color.RGBA{R: 200, G: 50, B: 50, A: 220}
	tx, ty := noseOffset(hx, hy, p.Direction)
	dtx, dty := tongueDir(p.Direction)
	setPixelSafe(buf, tx+dtx, ty+dty, tongueColor)
	setPixelSafe(buf, tx+dtx*2-1, ty+dty*2, tongueColor)
	setPixelSafe(buf, tx+dtx*2+1, ty+dty*2, tongueColor)

	// --- Scale chevron pattern along the body ---
	scaleLight := color.RGBA{R: 100, G: 140, B: 80, A: 70}
	scaleDark := color.RGBA{R: 40, G: 70, B: 30, A: 60}
	bodyLen := max(4, h/2)
	for i := 0; i < bodyLen; i++ {
		bx, by := bodyPointAlongSpine(cx, cy, p.Direction, w, h, i, bodyLen)
		// Chevron: V-shape at each segment
		spread := max(1, (w/6)-i/3)
		if i%2 == 0 {
			setPixelSafe(buf, bx-spread, by, scaleLight)
			setPixelSafe(buf, bx+spread, by, scaleLight)
			setPixelSafe(buf, bx, by-1, scaleLight)
		} else {
			setPixelSafe(buf, bx-spread, by, scaleDark)
			setPixelSafe(buf, bx+spread, by, scaleDark)
		}
	}

	// --- Belly ridge: lighter central stripe ---
	bellyColor := color.RGBA{R: 180, G: 200, B: 150, A: 50}
	for i := 0; i < bodyLen; i++ {
		bx, by := bodyPointAlongSpine(cx, cy, p.Direction, w, h, i, bodyLen)
		blendPixelSafe(buf, bx, by, bellyColor)
	}
}

// ----------------------------------------------------------------------------
// Flying details — wing membrane lines, feather tips, talons, beak
// ----------------------------------------------------------------------------

func renderFlyingDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Eyes: fierce glowing ---
	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	eyeGlow := color.RGBA{R: 255, G: 230, B: 100, A: 120}
	setPixelSafe(buf, hx-1, hy, eyeColor)
	setPixelSafe(buf, hx+1, hy, eyeColor)
	blendPixelSafe(buf, hx-2, hy, eyeGlow)
	blendPixelSafe(buf, hx+2, hy, eyeGlow)

	// --- Beak/horn: pointed shape at head front ---
	beakColor := color.RGBA{R: 200, G: 180, B: 60, A: 240}
	bx, by := noseOffset(hx, hy, p.Direction)
	setPixelSafe(buf, bx, by, beakColor)
	dx, dy := tongueDir(p.Direction)
	setPixelSafe(buf, bx+dx, by+dy, beakColor)

	// --- Wing membrane lines: radial lines from body center outward ---
	membraneColor := color.RGBA{R: 100, G: 80, B: 60, A: 90}
	membraneEdge := color.RGBA{R: 70, G: 55, B: 40, A: 130}
	wingSpan := max(4, w*2/5)
	numRibs := 4 + rng.Intn(3)
	for side := -1; side <= 1; side += 2 {
		for rib := 0; rib < numRibs; rib++ {
			angle := float64(side) * (0.3 + float64(rib)*0.3/float64(numRibs))
			for r := 2; r < wingSpan; r++ {
				rx := cx + int(float64(r)*math.Sin(angle))
				ry := cy + int(float64(r)*math.Cos(angle)*0.6) // flatter top-down
				if r == wingSpan-1 {
					setPixelSafe(buf, rx, ry, membraneEdge)
				} else {
					blendPixelSafe(buf, rx, ry, membraneColor)
				}
			}
		}
	}

	// --- Feather/scale tips along wing edge ---
	featherColor := color.RGBA{R: 140, G: 120, B: 100, A: 100}
	for side := -1; side <= 1; side += 2 {
		for i := 0; i < 5; i++ {
			fx := cx + side*(wingSpan-1-i)
			fy := cy + i - 2
			blendPixelSafe(buf, fx, fy, featherColor)
		}
	}

	// --- Talons at foot position ---
	talonColor := color.RGBA{R: 80, G: 60, B: 40, A: 220}
	ty := cy + h/4
	setPixelSafe(buf, cx-2, ty, talonColor)
	setPixelSafe(buf, cx-3, ty+1, talonColor)
	setPixelSafe(buf, cx+2, ty, talonColor)
	setPixelSafe(buf, cx+3, ty+1, talonColor)
}

// ----------------------------------------------------------------------------
// Blob details — bubbles, translucent highlights, nucleus, pseudopod tips
// ----------------------------------------------------------------------------

func renderBlobDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	// --- Internal bubbles: small circles inside the body ---
	bubbleHighlight := color.RGBA{R: 255, G: 255, B: 255, A: 60}
	bubbleShadow := color.RGBA{R: 80, G: 100, B: 80, A: 40}
	numBubbles := 4 + rng.Intn(4)
	bodyRadius := max(3, w/4)
	for i := 0; i < numBubbles; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * float64(bodyRadius) * 0.7
		bx := cx + int(dist*math.Cos(angle))
		by := cy + int(dist*math.Sin(angle))
		bubbleR := 1 + rng.Intn(2)
		// Draw bubble: highlight on top-left, shadow on bottom-right
		for dy := -bubbleR; dy <= bubbleR; dy++ {
			for dx := -bubbleR; dx <= bubbleR; dx++ {
				if dx*dx+dy*dy <= bubbleR*bubbleR {
					if dx+dy < 0 {
						blendPixelSafe(buf, bx+dx, by+dy, bubbleHighlight)
					} else {
						blendPixelSafe(buf, bx+dx, by+dy, bubbleShadow)
					}
				}
			}
		}
	}

	// --- Nucleus: dark central dot with glow ---
	nucleusColor := color.RGBA{R: 60, G: 80, B: 60, A: 180}
	nucleusGlow := color.RGBA{R: 100, G: 150, B: 100, A: 70}
	setPixelSafe(buf, cx, cy, nucleusColor)
	setPixelSafe(buf, cx+1, cy, nucleusColor)
	setPixelSafe(buf, cx, cy+1, nucleusColor)
	blendPixelSafe(buf, cx-1, cy, nucleusGlow)
	blendPixelSafe(buf, cx+1, cy+1, nucleusGlow)
	blendPixelSafe(buf, cx, cy-1, nucleusGlow)

	// --- Translucent sheen: bright arc across top ---
	sheenColor := color.RGBA{R: 255, G: 255, B: 255, A: 35}
	sheenRadius := max(3, w/3)
	for angle := -0.8; angle < 0.8; angle += 0.15 {
		sx := cx + int(float64(sheenRadius)*math.Sin(angle))
		sy := cy - int(float64(sheenRadius)*math.Cos(angle)*0.5) - 1
		blendPixelSafe(buf, sx, sy, sheenColor)
	}

	// --- Pseudopod tips: slightly brighter edges where pseudopods extend ---
	tipColor := color.RGBA{R: 160, G: 200, B: 160, A: 70}
	numTips := 3 + rng.Intn(3)
	for i := 0; i < numTips; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := float64(bodyRadius) * (0.8 + rng.Float64()*0.3)
		tx := cx + int(dist*math.Cos(angle))
		ty := cy + int(dist*math.Sin(angle))
		blendPixelSafe(buf, tx, ty, tipColor)
		blendPixelSafe(buf, tx+1, ty, tipColor)
	}
}

// ----------------------------------------------------------------------------
// Mechanical details — panel lines, rivets, sensor eyes, circuit traces
// ----------------------------------------------------------------------------

func renderMechanicalDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Sensor eyes: glowing colored dots ---
	sensorColor := color.RGBA{R: 50, G: 200, B: 255, A: 255}
	sensorGlow := color.RGBA{R: 50, G: 180, B: 230, A: 100}
	setPixelSafe(buf, hx-1, hy, sensorColor)
	setPixelSafe(buf, hx+1, hy, sensorColor)
	blendPixelSafe(buf, hx-2, hy, sensorGlow)
	blendPixelSafe(buf, hx+2, hy, sensorGlow)
	blendPixelSafe(buf, hx-1, hy-1, sensorGlow)
	blendPixelSafe(buf, hx+1, hy-1, sensorGlow)

	// --- Panel lines: grid pattern on torso ---
	panelColor := color.RGBA{R: 50, G: 50, B: 60, A: 100}
	gridSpacing := max(3, w/6)
	bodyTop := cy - h/4
	bodyBot := cy + h/4
	bodyLeft := cx - w/4
	bodyRight := cx + w/4
	for x := bodyLeft; x <= bodyRight; x += gridSpacing {
		for y := bodyTop; y <= bodyBot; y++ {
			if isOpaqueAt(buf, x, y) {
				blendPixelSafe(buf, x, y, panelColor)
			}
		}
	}
	for y := bodyTop; y <= bodyBot; y += gridSpacing {
		for x := bodyLeft; x <= bodyRight; x++ {
			if isOpaqueAt(buf, x, y) {
				blendPixelSafe(buf, x, y, panelColor)
			}
		}
	}

	// --- Rivets: bright dots at panel intersections ---
	rivetColor := color.RGBA{R: 180, G: 190, B: 200, A: 200}
	for x := bodyLeft; x <= bodyRight; x += gridSpacing {
		for y := bodyTop; y <= bodyBot; y += gridSpacing {
			if isOpaqueAt(buf, x, y) {
				setPixelSafe(buf, x, y, rivetColor)
			}
		}
	}

	// --- Circuit trace: thin glowing line from sensor to core ---
	traceColor := color.RGBA{R: 30, G: 160, B: 220, A: 90}
	traceBright := color.RGBA{R: 50, G: 200, B: 255, A: 140}
	// Vertical trace from head to center
	for y := hy + 1; y < cy; y++ {
		if y%2 == 0 {
			blendPixelSafe(buf, hx, y, traceColor)
		} else {
			blendPixelSafe(buf, hx, y, traceBright)
		}
	}

	// --- Core glow: central energy source ---
	coreColor := color.RGBA{R: 80, G: 220, B: 255, A: 180}
	coreGlow := color.RGBA{R: 50, G: 180, B: 230, A: 60}
	setPixelSafe(buf, cx, cy, coreColor)
	blendPixelSafe(buf, cx-1, cy, coreGlow)
	blendPixelSafe(buf, cx+1, cy, coreGlow)
	blendPixelSafe(buf, cx, cy-1, coreGlow)
	blendPixelSafe(buf, cx, cy+1, coreGlow)
}

// ----------------------------------------------------------------------------
// Undead details — bone highlights, rib arcs, eye socket glow, tattered edges
// ----------------------------------------------------------------------------

func renderUndeadDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Eye socket glow: eerie light from eye positions ---
	socketColor := color.RGBA{R: 100, G: 255, B: 100, A: 220}
	socketGlow := color.RGBA{R: 60, G: 200, B: 60, A: 80}
	if p.Genre == "horror" {
		socketColor = color.RGBA{R: 255, G: 80, B: 80, A: 220}
		socketGlow = color.RGBA{R: 200, G: 40, B: 40, A: 80}
	}
	setPixelSafe(buf, hx-1, hy, socketColor)
	setPixelSafe(buf, hx+1, hy, socketColor)
	blendPixelSafe(buf, hx-2, hy, socketGlow)
	blendPixelSafe(buf, hx+2, hy, socketGlow)
	blendPixelSafe(buf, hx-1, hy+1, socketGlow)
	blendPixelSafe(buf, hx+1, hy+1, socketGlow)

	// --- Skull nose: dark inverted triangle ---
	noseColor := color.RGBA{R: 30, G: 25, B: 25, A: 200}
	nx, ny := noseOffset(hx, hy, p.Direction)
	setPixelSafe(buf, nx, ny, noseColor)
	setPixelSafe(buf, nx-1, ny+1, noseColor)
	setPixelSafe(buf, nx+1, ny+1, noseColor)

	// --- Rib arcs: curved bone lines across torso ---
	boneColor := color.RGBA{R: 220, G: 215, B: 200, A: 120}
	boneShadow := color.RGBA{R: 160, G: 155, B: 140, A: 80}
	numRibs := 3 + rng.Intn(2)
	ribSpacing := max(2, h/(numRibs*3))
	for i := 0; i < numRibs; i++ {
		ribY := cy - h/8 + i*ribSpacing
		ribWidth := max(2, w/4-i)
		for x := -ribWidth; x <= ribWidth; x++ {
			curve := int(math.Abs(float64(x)) * 0.3)
			blendPixelSafe(buf, cx+x, ribY+curve, boneColor)
			if x%2 == 0 {
				blendPixelSafe(buf, cx+x, ribY+curve+1, boneShadow)
			}
		}
	}

	// --- Spine: vertical bone line down center ---
	spineColor := color.RGBA{R: 200, G: 195, B: 180, A: 100}
	for y := cy - h/6; y < cy+h/6; y++ {
		blendPixelSafe(buf, cx, y, spineColor)
	}

	// --- Tattered edges: random dark pixels along the outline ---
	tatterColor := color.RGBA{R: 40, G: 35, B: 30, A: 100}
	for i := 0; i < 12+rng.Intn(6); i++ {
		angle := rng.Float64() * 2 * math.Pi
		radius := float64(max(3, w/3)) + rng.Float64()*2
		tx := cx + int(radius*math.Cos(angle))
		ty := cy + int(radius*math.Sin(angle))
		blendPixelSafe(buf, tx, ty, tatterColor)
	}
}

// ============================================================================
// Helper functions
// ============================================================================

// headOffset returns the head center position based on facing direction.
func headOffset(cx, cy int, direction string, w, h int) (int, int) {
	offset := max(2, h/4)
	switch direction {
	case "up":
		return cx, cy - offset
	case "down":
		return cx, cy + offset
	case "left":
		return cx - offset, cy
	case "right":
		return cx + offset, cy
	default:
		return cx, cy - offset // default facing up in top-down
	}
}

// noseOffset returns the position just in front of the head.
func noseOffset(hx, hy int, direction string) (int, int) {
	switch direction {
	case "up":
		return hx, hy - 2
	case "down":
		return hx, hy + 2
	case "left":
		return hx - 2, hy
	case "right":
		return hx + 2, hy
	default:
		return hx, hy - 2
	}
}

// tailOffset returns the tail/rear position opposite the head.
func tailOffset(cx, cy int, direction string, w, h int) (int, int) {
	offset := max(2, h/4)
	switch direction {
	case "up":
		return cx, cy + offset
	case "down":
		return cx, cy - offset
	case "left":
		return cx + offset, cy
	case "right":
		return cx - offset, cy
	default:
		return cx, cy + offset
	}
}

// tailDirection returns the per-pixel offset for drawing a tail.
func tailDirection(direction string, i int) (int, int) {
	switch direction {
	case "up":
		return 0, i
	case "down":
		return 0, -i
	case "left":
		return i, 0
	case "right":
		return -i, 0
	default:
		return 0, i
	}
}

// tongueDir returns the unit direction for tongue/beak extension.
func tongueDir(direction string) (int, int) {
	switch direction {
	case "up":
		return 0, -1
	case "down":
		return 0, 1
	case "left":
		return -1, 0
	case "right":
		return 1, 0
	default:
		return 0, -1
	}
}

// mandibleDir returns offset for mandible drawing, with side = -1 or +1.
func mandibleDir(direction string, side int) (int, int) {
	switch direction {
	case "up":
		return side, -1
	case "down":
		return side, 1
	case "left":
		return -1, side
	case "right":
		return 1, side
	default:
		return side, -1
	}
}

// bodyPointAlongSpine returns a point along the creature's body axis.
func bodyPointAlongSpine(cx, cy int, direction string, w, h, index, total int) (int, int) {
	t := float64(index) / float64(max(1, total-1))
	halfLen := max(3, h/3)
	switch direction {
	case "up":
		return cx, cy - halfLen + int(float64(2*halfLen)*t)
	case "down":
		return cx, cy + halfLen - int(float64(2*halfLen)*t)
	case "left":
		return cx - halfLen + int(float64(2*halfLen)*t), cy
	case "right":
		return cx + halfLen - int(float64(2*halfLen)*t), cy
	default:
		return cx, cy - halfLen + int(float64(2*halfLen)*t)
	}
}

// isOpaqueAt checks if a pixel has non-zero alpha.
func isOpaqueAt(buf *image.RGBA, x, y int) bool {
	if x < buf.Rect.Min.X || x >= buf.Rect.Max.X || y < buf.Rect.Min.Y || y >= buf.Rect.Max.Y {
		return false
	}
	return buf.RGBAAt(x, y).A > 0
}

func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// ----------------------------------------------------------------------------
// Insect details — compound eyes, antennae, wing covers, leg joints
// ----------------------------------------------------------------------------

func renderInsectDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	// Head center depends on facing direction
	hx, hy := headOffset(cx, cy, p.Direction, w, h)

	// --- Compound eyes: two large colored dots flanking the head ---
	eyeBase := color.RGBA{
		R: uint8(80 + rng.Intn(120)),
		G: uint8(40 + rng.Intn(80)),
		B: uint8(20 + rng.Intn(60)),
		A: 255,
	}
	eyeHighlight := color.RGBA{R: 240, G: 240, B: 220, A: 160}
	eyeSpread := max(2, w/8)

	// Left compound eye (2px cluster)
	setPixelSafe(buf, hx-eyeSpread, hy, eyeBase)
	setPixelSafe(buf, hx-eyeSpread-1, hy, eyeBase)
	setPixelSafe(buf, hx-eyeSpread, hy-1, eyeBase)
	setPixelSafe(buf, hx-eyeSpread, hy+1, eyeHighlight)
	// Right compound eye
	setPixelSafe(buf, hx+eyeSpread, hy, eyeBase)
	setPixelSafe(buf, hx+eyeSpread+1, hy, eyeBase)
	setPixelSafe(buf, hx+eyeSpread, hy-1, eyeBase)
	setPixelSafe(buf, hx+eyeSpread, hy+1, eyeHighlight)

	// --- Antennae: two thin lines extending from the head ---
	antennaColor := color.RGBA{R: 90, G: 70, B: 50, A: 200}
	antennaTip := color.RGBA{R: 120, G: 100, B: 70, A: 230}
	antLen := max(3, h/5)
	adx, ady := antennaDir(p.Direction)
	for i := 0; i < antLen; i++ {
		// Left antenna curves outward
		lx := hx - eyeSpread + adx*i - i/2
		ly := hy + ady*i
		c := antennaColor
		if i == antLen-1 {
			c = antennaTip
		}
		setPixelSafe(buf, lx, ly, c)
		// Right antenna curves outward (mirrored)
		rx := hx + eyeSpread + adx*i + i/2
		setPixelSafe(buf, rx, ly, c)
	}

	// --- Mandibles: two small dark pincers in front of the head ---
	mandColor := color.RGBA{R: 50, G: 35, B: 25, A: 240}
	mx, my := noseOffset(hx, hy, p.Direction)
	setPixelSafe(buf, mx-1, my, mandColor)
	setPixelSafe(buf, mx+1, my, mandColor)
	setPixelSafe(buf, mx-2, my+intSign(ady), mandColor)
	setPixelSafe(buf, mx+2, my+intSign(ady), mandColor)

	// --- Wing covers (elytra): two faint lines on the abdomen ---
	elytraColor := color.RGBA{R: 140, G: 120, B: 80, A: 100}
	abdY := cy + h/6
	if p.Direction == "up" {
		abdY = cy - h/6
	}
	for i := -2; i <= 2; i++ {
		setPixelSafe(buf, cx, abdY+i, elytraColor)
	}

	// --- Leg joints: small dots at leg attachment points ---
	jointColor := color.RGBA{R: 70, G: 55, B: 40, A: 180}
	legPairY := [3]int{cy - h/6, cy, cy + h/6}
	legSpreadX := max(3, w*3/10)
	for _, ly := range legPairY {
		setPixelSafe(buf, cx-legSpreadX, ly, jointColor)
		setPixelSafe(buf, cx+legSpreadX, ly, jointColor)
	}

	// --- Segmentation line between thorax and abdomen ---
	segColor := color.RGBA{R: 40, G: 30, B: 20, A: 120}
	segY := cy + 1
	segW := max(2, w/6)
	for dx := -segW; dx <= segW; dx++ {
		setPixelSafe(buf, cx+dx, segY, segColor)
	}
}

// antennaDir returns the primary direction offsets for antennae extension.
func antennaDir(direction string) (dx, dy int) {
	switch direction {
	case "up":
		return 0, 1
	case "left":
		return -1, 0
	case "right":
		return 1, 0
	default: // down
		return 0, -1
	}
}

// intSign returns -1, 0, or 1.
func intSign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

// ----------------------------------------------------------------------------
// Multi-limbed horror details — multiple eyes, sucker patterns, tendril tips
// ----------------------------------------------------------------------------

func renderMultiLimbedDetails(buf *image.RGBA, p CreatureDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	// --- Multiple eyes: 3-7 eyes scattered across the central mass ---
	numEyes := 3 + rng.Intn(5)
	eyeRadius := max(2, w/8)
	for i := 0; i < numEyes; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numEyes)
		r := float64(eyeRadius) * (0.4 + rng.Float64()*0.6)
		ex := cx + int(r*math.Cos(angle))
		ey := cy + int(r*math.Sin(angle)) - h/8

		// Each eye has a unique color
		eyeColor := color.RGBA{
			R: uint8(100 + rng.Intn(155)),
			G: uint8(rng.Intn(100)),
			B: uint8(rng.Intn(80)),
			A: 255,
		}
		pupilColor := color.RGBA{R: 10, G: 10, B: 10, A: 255}
		glintColor := color.RGBA{R: 255, G: 255, B: 240, A: 200}

		setPixelSafe(buf, ex, ey, eyeColor)
		setPixelSafe(buf, ex+1, ey, eyeColor)
		setPixelSafe(buf, ex, ey+1, pupilColor)
		if rng.Float64() > 0.4 {
			setPixelSafe(buf, ex+1, ey-1, glintColor) // specular glint
		}
	}

	// --- Tentacle suckers: small circular dots along radial arms ---
	suckerColor := color.RGBA{R: 160, G: 120, B: 140, A: 140}
	suckerHighlight := color.RGBA{R: 200, G: 170, B: 180, A: 100}
	numTentacles := 5 + rng.Intn(3)
	for t := 0; t < numTentacles; t++ {
		angle := float64(t)*2.0*math.Pi/float64(numTentacles) + rng.Float64()*0.3
		tentLen := max(4, w/3)
		for s := 2; s < tentLen; s += 2 {
			sx := cx + int(float64(s)*math.Cos(angle))
			sy := cy + int(float64(s)*math.Sin(angle))
			setPixelSafe(buf, sx, sy, suckerColor)
			if s%4 == 0 {
				setPixelSafe(buf, sx+1, sy, suckerHighlight)
			}
		}

		// Tendril tip — slightly brighter/different colored point
		tipDist := float64(tentLen)
		tx := cx + int(tipDist*math.Cos(angle))
		ty := cy + int(tipDist*math.Sin(angle))
		tipColor := color.RGBA{
			R: uint8(180 + rng.Intn(75)),
			G: uint8(80 + rng.Intn(60)),
			B: uint8(100 + rng.Intn(60)),
			A: 220,
		}
		setPixelSafe(buf, tx, ty, tipColor)
		setPixelSafe(buf, tx-1, ty, tipColor)
		setPixelSafe(buf, tx+1, ty, tipColor)
	}

	// --- Central maw: a dark opening in the center ---
	mawColor := color.RGBA{R: 30, G: 15, B: 20, A: 230}
	mawRing := color.RGBA{R: 100, G: 50, B: 60, A: 180}
	setPixelSafe(buf, cx, cy, mawColor)
	setPixelSafe(buf, cx+1, cy, mawColor)
	setPixelSafe(buf, cx, cy+1, mawColor)
	setPixelSafe(buf, cx+1, cy+1, mawColor)
	// Ring around maw
	for _, d := range [][2]int{
		{-1, -1},
		{0, -1},
		{1, -1},
		{2, -1},
		{-1, 0},
		{2, 0},
		{-1, 1},
		{2, 1},
		{-1, 2},
		{0, 2},
		{1, 2},
		{2, 2},
	} {
		setPixelSafe(buf, cx+d[0], cy+d[1], mawRing)
	}

	// --- Veiny texture: faint irregular lines radiating from center ---
	veinColor := color.RGBA{R: 80, G: 40, B: 50, A: 70}
	numVeins := 4 + rng.Intn(4)
	for v := 0; v < numVeins; v++ {
		angle := rng.Float64() * 2 * math.Pi
		veinLen := 3 + rng.Intn(max(1, w/4))
		for i := 2; i < veinLen; i++ {
			vx := cx + int(float64(i)*math.Cos(angle))
			vy := cy + int(float64(i)*math.Sin(angle))
			// Wobble the vein path
			vx += rng.Intn(3) - 1
			setPixelSafe(buf, vx, vy, veinColor)
		}
	}
}
