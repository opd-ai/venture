// Package sprites provides seed-based garment structure line rendering for
// procedurally generated entity sprites viewed from above. After body part
// shapes, shading, and texture patterns are drawn, this module adds structural
// detail lines — collars, belts, hems, seams, fasteners, layering edges —
// that transform flat colored shapes into recognizable clothing items.
//
// Each entity's seed deterministically selects a garment type (tunic, robe,
// vest, armor plate, shirt, cloak) and genre influences the distribution.
// All rendering operates on standard image.RGBA buffers for compositing.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// GarmentType identifies the structural clothing type drawn on the torso.
type GarmentType int

const (
	// GarmentTunic has a V-neckline, belt at waist, hem, and side seams.
	GarmentTunic GarmentType = iota
	// GarmentRobe has a round neckline, center fold, flowing hem.
	GarmentRobe
	// GarmentVest has an open front with V-cut and button line.
	GarmentVest
	// GarmentPlateArmor has plate boundaries, rivet dots, and pauldrons.
	GarmentPlateArmor
	// GarmentShirt has a round collar, button line, and sleeve cuffs.
	GarmentShirt
	// GarmentCloak has a hood edge, broad drape, and clasp at throat.
	GarmentCloak
	// garmentTypeCount is the number of garment types.
	garmentTypeCount
)

// String returns a human-readable name for the garment type.
func (g GarmentType) String() string {
	switch g {
	case GarmentTunic:
		return "tunic"
	case GarmentRobe:
		return "robe"
	case GarmentVest:
		return "vest"
	case GarmentPlateArmor:
		return "plate_armor"
	case GarmentShirt:
		return "shirt"
	case GarmentCloak:
		return "cloak"
	default:
		return "unknown"
	}
}

// GarmentDetailParams configures garment structure line rendering.
type GarmentDetailParams struct {
	Width, Height int
	Seed          int64
	Genre         string
	Direction     string // facing direction ("up","down","left","right")
}

// genreGarmentWeights returns per-genre weights for each garment type.
// Higher weight = more likely to be selected for that genre.
func genreGarmentWeights(genre string) [garmentTypeCount]int {
	switch genre {
	case "fantasy":
		return [garmentTypeCount]int{30, 25, 10, 15, 10, 10} // tunic & robe heavy
	case "sci-fi", "scifi":
		return [garmentTypeCount]int{5, 5, 15, 30, 35, 10} // shirt & armor heavy
	case "horror":
		return [garmentTypeCount]int{15, 30, 10, 5, 10, 30} // robe & cloak heavy
	case "cyberpunk":
		return [garmentTypeCount]int{5, 5, 30, 25, 20, 15} // vest & armor heavy
	case "post-apocalyptic", "postapoc":
		return [garmentTypeCount]int{20, 10, 25, 15, 15, 15} // tunic & vest heavy
	default:
		return [garmentTypeCount]int{20, 15, 15, 15, 20, 15} // balanced
	}
}

// SelectGarmentType deterministically picks a garment type from seed and genre.
func SelectGarmentType(seed int64, genre string) GarmentType {
	rng := rand.New(rand.NewSource(seed * 8419))
	weights := genreGarmentWeights(genre)
	total := 0
	for _, w := range weights {
		total += w
	}
	roll := rng.Intn(total)
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return GarmentType(i)
		}
	}
	return GarmentTunic
}

// RenderGarmentDetail draws structural garment lines onto buf.
// Call after body part shapes, shading, and clothing patterns, but before
// role-specific details and sprite finalization.
func RenderGarmentDetail(buf *image.RGBA, params GarmentDetailParams) {
	if buf == nil || params.Width <= 0 || params.Height <= 0 {
		return
	}
	rng := rand.New(rand.NewSource(params.Seed * 8419))
	garment := SelectGarmentType(params.Seed, params.Genre)

	// Sample the dominant body color from the torso center for detail line colors
	baseColor := sampleBodyColor(buf, params.Width, params.Height)
	lineLight := garmentLighten(baseColor, 1.35)
	lineDark := garmentDarken(baseColor, 0.65)
	lineSubtle := garmentDarken(baseColor, 0.80)

	switch garment {
	case GarmentTunic:
		renderTunicDetails(buf, params, lineLight, lineDark, lineSubtle, rng)
	case GarmentRobe:
		renderRobeDetails(buf, params, lineLight, lineDark, lineSubtle, rng)
	case GarmentVest:
		renderVestDetails(buf, params, lineLight, lineDark, lineSubtle, rng)
	case GarmentPlateArmor:
		renderArmorDetails(buf, params, lineLight, lineDark, lineSubtle, rng)
	case GarmentShirt:
		renderShirtDetails(buf, params, lineLight, lineDark, lineSubtle, rng)
	case GarmentCloak:
		renderCloakDetails(buf, params, lineLight, lineDark, lineSubtle, rng)
	}
}

// sampleBodyColor reads the pixel at the torso center to derive line colors.
func sampleBodyColor(buf *image.RGBA, w, h int) color.RGBA {
	cx, cy := w/2, h/2+h/8 // slightly below center = torso area
	if cx < 0 || cx >= w || cy < 0 || cy >= h {
		return color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}
	idx := (cy-buf.Bounds().Min.Y)*buf.Stride + (cx-buf.Bounds().Min.X)*4
	if idx+3 >= len(buf.Pix) || idx < 0 {
		return color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}
	return color.RGBA{R: buf.Pix[idx], G: buf.Pix[idx+1], B: buf.Pix[idx+2], A: buf.Pix[idx+3]}
}

// garmentLighten returns a lighter version of the color by the given factor.
func garmentLighten(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: clampByte(float64(c.R) * factor),
		G: clampByte(float64(c.G) * factor),
		B: clampByte(float64(c.B) * factor),
		A: c.A,
	}
}

// garmentDarken returns a darker version of the color by the given factor.
func garmentDarken(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: clampByte(float64(c.R) * factor),
		G: clampByte(float64(c.G) * factor),
		B: clampByte(float64(c.B) * factor),
		A: c.A,
	}
}

// --- Tunic: V-neckline, belt, hem, side seams ---

func renderTunicDetails(buf *image.RGBA, p GarmentDetailParams, light, dark, subtle color.RGBA, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx := w / 2

	// Torso region in aerial view: roughly 25%-65% height
	torsoTop := h * 25 / 100
	torsoBot := h * 65 / 100
	torsoMid := (torsoTop + torsoBot) / 2

	// --- V-neckline: two short diagonal lines from center outward ---
	neckY := torsoTop + 1
	neckSpread := garmentMax(1, w/10)
	for i := 0; i <= neckSpread; i++ {
		leftX := cx - i
		rightX := cx + i
		ny := neckY + i
		blendPixelSafe(buf, leftX, ny, dark)
		blendPixelSafe(buf, rightX, ny, dark)
		// Highlight just above the dark line
		if i > 0 {
			blendPixelSafe(buf, leftX, ny-1, light)
			blendPixelSafe(buf, rightX, ny-1, light)
		}
	}

	// --- Belt line: horizontal dark line across waist ---
	beltY := torsoMid + garmentMax(1, h/16)
	beltHalfW := garmentMax(2, w*3/10)
	beltColor := garmentDarken(dark, 0.85)
	for dx := -beltHalfW; dx <= beltHalfW; dx++ {
		px := cx + dx
		if garmentIsOpaque(buf, px, beltY) {
			blendPixelSafe(buf, px, beltY, beltColor)
			// Belt highlight on top edge
			blendPixelSafe(buf, px, beltY-1, color.RGBA{R: light.R, G: light.G, B: light.B, A: 80})
		}
	}
	// Belt buckle: bright pixel at center
	blendPixelSafe(buf, cx, beltY, light)

	// --- Hem line: subtle dark line at torso bottom ---
	hemY := torsoBot - 1
	hemHalfW := garmentMax(2, w*3/10)
	for dx := -hemHalfW; dx <= hemHalfW; dx++ {
		px := cx + dx
		if garmentIsOpaque(buf, px, hemY) {
			blendPixelSafe(buf, px, hemY, subtle)
		}
	}

	// --- Side seams: vertical subtle lines along torso edges ---
	seamOffX := garmentMax(2, w*5/20)
	for y := torsoTop + 2; y < torsoBot-1; y++ {
		if garmentIsOpaque(buf, cx-seamOffX, y) {
			blendPixelSafe(buf, cx-seamOffX, y, color.RGBA{R: subtle.R, G: subtle.G, B: subtle.B, A: 60})
		}
		if garmentIsOpaque(buf, cx+seamOffX, y) {
			blendPixelSafe(buf, cx+seamOffX, y, color.RGBA{R: subtle.R, G: subtle.G, B: subtle.B, A: 60})
		}
	}
}

// --- Robe: round neckline, center fold, flowing hem ---

func renderRobeDetails(buf *image.RGBA, p GarmentDetailParams, light, dark, subtle color.RGBA, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx := w / 2

	torsoTop := h * 25 / 100
	torsoBot := h * 70 / 100 // robes extend lower

	// --- Round neckline: small arc of dark pixels around head base ---
	neckRadius := garmentMax(2, w/8)
	neckCY := torsoTop + 1
	for angle := 0.0; angle < math.Pi; angle += 0.25 {
		px := cx + int(float64(neckRadius)*math.Cos(angle))
		py := neckCY + int(float64(neckRadius)*0.5*math.Sin(angle))
		blendPixelSafe(buf, px, py, dark)
	}

	// --- Center fold line: vertical line down the front ---
	foldTop := torsoTop + neckRadius/2 + 1
	foldBot := torsoBot - 1
	for y := foldTop; y <= foldBot; y++ {
		if garmentIsOpaque(buf, cx, y) {
			blendPixelSafe(buf, cx, y, subtle)
			// Slight highlight to left of fold
			blendPixelSafe(buf, cx-1, y, color.RGBA{R: light.R, G: light.G, B: light.B, A: 45})
		}
	}

	// --- Flowing hem: wavy dark line at bottom, wider than torso ---
	hemHalfW := garmentMax(3, w*7/20)
	hemY := torsoBot
	waveAmp := rng.Intn(2) // 0 or 1 pixel wave
	for dx := -hemHalfW; dx <= hemHalfW; dx++ {
		px := cx + dx
		wave := 0
		if waveAmp > 0 && dx%3 == 0 {
			wave = 1
		}
		py := hemY + wave
		if garmentIsOpaque(buf, px, py) {
			blendPixelSafe(buf, px, py, dark)
		}
		// Light edge above
		if garmentIsOpaque(buf, px, py-1) {
			blendPixelSafe(buf, px, py-1, color.RGBA{R: light.R, G: light.G, B: light.B, A: 50})
		}
	}

	// --- Sash / cord at waist ---
	sashY := (torsoTop + torsoBot) / 2
	sashW := garmentMax(1, w/6)
	sashColor := garmentDarken(dark, 0.90)
	for dx := -sashW; dx <= sashW; dx++ {
		if garmentIsOpaque(buf, cx+dx, sashY) {
			blendPixelSafe(buf, cx+dx, sashY, sashColor)
		}
	}
}

// --- Vest: V-cut front, button line, no sleeves ---

func renderVestDetails(buf *image.RGBA, p GarmentDetailParams, light, dark, subtle color.RGBA, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx := w / 2

	torsoTop := h * 25 / 100
	torsoBot := h * 60 / 100

	// --- V-cut lapels: two lines converging toward center ---
	lapelLen := garmentMax(2, h/8)
	lapelSpread := garmentMax(2, w/7)
	for i := 0; i < lapelLen; i++ {
		t := float64(i) / float64(garmentMax(1, lapelLen))
		offX := int(float64(lapelSpread) * (1.0 - t))
		py := torsoTop + i
		// Left lapel edge
		blendPixelSafe(buf, cx-offX, py, dark)
		blendPixelSafe(buf, cx-offX+1, py, color.RGBA{R: light.R, G: light.G, B: light.B, A: 60})
		// Right lapel edge
		blendPixelSafe(buf, cx+offX, py, dark)
		blendPixelSafe(buf, cx+offX-1, py, color.RGBA{R: light.R, G: light.G, B: light.B, A: 60})
	}

	// --- Button line: dots down center ---
	buttonStart := torsoTop + lapelLen
	buttonEnd := torsoBot - 1
	buttonSpacing := garmentMax(2, h/10)
	for y := buttonStart; y <= buttonEnd; y += buttonSpacing {
		if garmentIsOpaque(buf, cx, y) {
			blendPixelSafe(buf, cx, y, light)
		}
	}

	// --- Bottom hem ---
	hemHalfW := garmentMax(2, w/4)
	for dx := -hemHalfW; dx <= hemHalfW; dx++ {
		if garmentIsOpaque(buf, cx+dx, torsoBot) {
			blendPixelSafe(buf, cx+dx, torsoBot, subtle)
		}
	}
}

// --- Plate Armor: plate boundaries, rivets, pauldrons ---

func renderArmorDetails(buf *image.RGBA, p GarmentDetailParams, light, dark, subtle color.RGBA, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx := w / 2

	torsoTop := h * 25 / 100
	torsoBot := h * 62 / 100
	torsoMid := (torsoTop + torsoBot) / 2

	// --- Chest plate boundary: horizontal line separating upper and lower plate ---
	plateHalfW := garmentMax(2, w*5/20)
	for dx := -plateHalfW; dx <= plateHalfW; dx++ {
		if garmentIsOpaque(buf, cx+dx, torsoMid) {
			blendPixelSafe(buf, cx+dx, torsoMid, dark)
			blendPixelSafe(buf, cx+dx, torsoMid+1, color.RGBA{R: light.R, G: light.G, B: light.B, A: 70})
		}
	}

	// --- Pauldron outlines on shoulders ---
	shoulderY := torsoTop + 1
	shoulderW := garmentMax(2, w/5)
	// Left pauldron arc
	for i := 0; i < shoulderW; i++ {
		arc := int(math.Sqrt(float64(shoulderW*shoulderW-i*i)) * 0.4)
		blendPixelSafe(buf, cx-w/4-i, shoulderY+arc, dark)
		blendPixelSafe(buf, cx-w/4-i, shoulderY+arc-1, light)
	}
	// Right pauldron arc
	for i := 0; i < shoulderW; i++ {
		arc := int(math.Sqrt(float64(shoulderW*shoulderW-i*i)) * 0.4)
		blendPixelSafe(buf, cx+w/4+i, shoulderY+arc, dark)
		blendPixelSafe(buf, cx+w/4+i, shoulderY+arc-1, light)
	}

	// --- Rivet dots along plate edges ---
	rivetSpacing := garmentMax(2, w/6)
	rivetColor := garmentLighten(light, 1.2)
	for dx := -plateHalfW; dx <= plateHalfW; dx += rivetSpacing {
		if garmentIsOpaque(buf, cx+dx, torsoMid-1) {
			blendPixelSafe(buf, cx+dx, torsoMid-1, rivetColor)
		}
	}

	// --- Gorget: collar plate around neck ---
	gorgetRadius := garmentMax(2, w/7)
	for angle := 0.0; angle < math.Pi; angle += 0.3 {
		px := cx + int(float64(gorgetRadius)*math.Cos(angle))
		py := torsoTop + int(float64(gorgetRadius)*0.35*math.Sin(angle))
		blendPixelSafe(buf, px, py, dark)
	}

	// --- Waist guard line ---
	waistY := torsoBot - 2
	for dx := -plateHalfW; dx <= plateHalfW; dx++ {
		if garmentIsOpaque(buf, cx+dx, waistY) {
			blendPixelSafe(buf, cx+dx, waistY, subtle)
		}
	}
}

// --- Shirt: round collar, button line, sleeve cuffs ---

func renderShirtDetails(buf *image.RGBA, p GarmentDetailParams, light, dark, subtle color.RGBA, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx := w / 2

	torsoTop := h * 25 / 100
	torsoBot := h * 60 / 100

	// --- Round collar ---
	collarRadius := garmentMax(2, w/8)
	collarY := torsoTop
	for angle := 0.0; angle <= math.Pi; angle += 0.2 {
		px := cx + int(float64(collarRadius)*math.Cos(angle))
		py := collarY + int(float64(collarRadius)*0.4*math.Sin(angle))
		blendPixelSafe(buf, px, py, dark)
	}
	// Collar highlight inner edge
	innerRadius := garmentMax(1, collarRadius-1)
	for angle := 0.3; angle <= math.Pi-0.3; angle += 0.25 {
		px := cx + int(float64(innerRadius)*math.Cos(angle))
		py := collarY + int(float64(innerRadius)*0.35*math.Sin(angle))
		blendPixelSafe(buf, px, py, color.RGBA{R: light.R, G: light.G, B: light.B, A: 70})
	}

	// --- Button line (3-4 small dots) ---
	numButtons := 3 + rng.Intn(2)
	buttonSpacing := (torsoBot - torsoTop - 3) / garmentMax(1, numButtons)
	for i := 0; i < numButtons; i++ {
		by := torsoTop + collarRadius/2 + 2 + i*buttonSpacing
		if garmentIsOpaque(buf, cx, by) {
			blendPixelSafe(buf, cx, by, light)
		}
	}

	// --- Sleeve cuffs: short horizontal lines at arm edges ---
	cuffY := torsoMidY(torsoTop, torsoBot) + (torsoBot-torsoTop)/4
	cuffW := garmentMax(1, w/10)
	armOffX := garmentMax(3, w*3/10)
	// Left cuff
	for dx := 0; dx < cuffW; dx++ {
		if garmentIsOpaque(buf, cx-armOffX-dx, cuffY) {
			blendPixelSafe(buf, cx-armOffX-dx, cuffY, dark)
		}
	}
	// Right cuff
	for dx := 0; dx < cuffW; dx++ {
		if garmentIsOpaque(buf, cx+armOffX+dx, cuffY) {
			blendPixelSafe(buf, cx+armOffX+dx, cuffY, dark)
		}
	}

	// --- Bottom hem ---
	hemHalfW := garmentMax(2, w/4)
	for dx := -hemHalfW; dx <= hemHalfW; dx++ {
		if garmentIsOpaque(buf, cx+dx, torsoBot) {
			blendPixelSafe(buf, cx+dx, torsoBot, subtle)
		}
	}
}

// --- Cloak: hood edge, broad drape, clasp ---

func renderCloakDetails(buf *image.RGBA, p GarmentDetailParams, light, dark, subtle color.RGBA, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx := w / 2

	headTop := h * 10 / 100
	torsoBot := h * 70 / 100

	// --- Hood rim: arc around the head area ---
	hoodRadius := garmentMax(3, w/5)
	hoodCY := headTop + hoodRadius
	for angle := -math.Pi; angle <= 0; angle += 0.15 {
		px := cx + int(float64(hoodRadius)*math.Cos(angle))
		py := hoodCY + int(float64(hoodRadius)*0.7*math.Sin(angle))
		blendPixelSafe(buf, px, py, dark)
		// Outer highlight
		outerPx := cx + int(float64(hoodRadius+1)*math.Cos(angle))
		outerPy := hoodCY + int(float64(hoodRadius+1)*0.7*math.Sin(angle))
		blendPixelSafe(buf, outerPx, outerPy, color.RGBA{R: light.R, G: light.G, B: light.B, A: 60})
	}

	// --- Clasp at throat: bright pixel pair ---
	claspY := hoodCY + hoodRadius/2 + 1
	blendPixelSafe(buf, cx-1, claspY, light)
	blendPixelSafe(buf, cx, claspY, garmentLighten(light, 1.3))
	blendPixelSafe(buf, cx+1, claspY, light)

	// --- Drape edges: diagonal lines from shoulders downward ---
	drapeStartY := claspY + 2
	drapeLen := torsoBot - drapeStartY
	drapeSpread := garmentMax(2, w/5)
	for i := 0; i < drapeLen; i++ {
		t := float64(i) / float64(garmentMax(1, drapeLen))
		offX := int(float64(drapeSpread) * (0.5 + t*0.5))
		py := drapeStartY + i
		// Left drape edge
		if garmentIsOpaque(buf, cx-offX, py) {
			blendPixelSafe(buf, cx-offX, py, subtle)
		}
		// Right drape edge
		if garmentIsOpaque(buf, cx+offX, py) {
			blendPixelSafe(buf, cx+offX, py, subtle)
		}
	}

	// --- Bottom drape hem: wider dark line ---
	hemHalfW := garmentMax(3, w*7/20)
	hemY := torsoBot
	for dx := -hemHalfW; dx <= hemHalfW; dx++ {
		if garmentIsOpaque(buf, cx+dx, hemY) {
			blendPixelSafe(buf, cx+dx, hemY, dark)
		}
	}
}

// --- Helpers ---

// garmentIsOpaque checks whether the pixel at (x,y) is opaque (alpha > 40).
// Garment lines should only draw where the body part already has content.
func garmentIsOpaque(buf *image.RGBA, x, y int) bool {
	bounds := buf.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return false
	}
	idx := (y-bounds.Min.Y)*buf.Stride + (x-bounds.Min.X)*4
	if idx+3 >= len(buf.Pix) || idx < 0 {
		return false
	}
	return buf.Pix[idx+3] > 40
}

func garmentMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func torsoMidY(top, bot int) int {
	return (top + bot) / 2
}
