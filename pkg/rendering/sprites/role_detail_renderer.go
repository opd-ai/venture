// Package sprites provides role-specific pixel-level detail rendering for
// humanoid NPC sprites viewed from above. After the base body-part shapes,
// hair, face, and clothing patterns are drawn, this module adds distinctive
// visual markers — arcane runes, weapon belts, shoulder plate highlights,
// hood shadows, belt pouches, quiver lines, holy symbols — that make each
// NPC role immediately identifiable at 32×32 pixel resolution.
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

// RoleDetailParams configures role-specific detail rendering.
type RoleDetailParams struct {
	Width, Height int
	Role          VisualRole
	Direction     string // facing direction ("up", "down", "left", "right")
	Seed          int64
	Genre         string
}

// RenderRoleDetails draws role-specific visual markers onto buf.
// Call this after base template parts, hair, face, and clothing patterns
// have been rendered. Non-role or empty roles are silently skipped.
func RenderRoleDetails(buf *image.RGBA, params RoleDetailParams) {
	if buf == nil || params.Width <= 0 || params.Height <= 0 {
		return
	}
	rng := rand.New(rand.NewSource(params.Seed * 6271))
	switch params.Role {
	case RoleMage:
		renderMageDetails(buf, params, rng)
	case RoleWarrior:
		renderWarriorDetails(buf, params, rng)
	case RoleKnight:
		renderKnightDetails(buf, params, rng)
	case RoleRogue:
		renderRogueDetails(buf, params, rng)
	case RoleMerchant:
		renderMerchantDetails(buf, params, rng)
	case RoleRanger:
		renderRangerDetails(buf, params, rng)
	case RolePriest:
		renderPriestDetails(buf, params, rng)
	}
}

// ----------------------------------------------------------------------------
// Mage — arcane rune on robe, glowing orb at staff tip, hat shadow crease
// ----------------------------------------------------------------------------

func renderMageDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	// Genre-aware magic color
	arcaneColor := mageArcaneColor(p.Genre, rng)
	arcaneGlow := color.RGBA{R: arcaneColor.R, G: arcaneColor.G, B: arcaneColor.B, A: 100}

	// --- Arcane rune on torso (small diamond/star pattern) ---
	runeY := cy + 1 // torso center in top-down
	runeSize := maxDetail(1, w/10)
	// Diamond pattern: top, right, bottom, left, center
	blendPixelSafe(buf, cx, runeY-runeSize, arcaneColor)
	blendPixelSafe(buf, cx+runeSize, runeY, arcaneColor)
	blendPixelSafe(buf, cx, runeY+runeSize, arcaneColor)
	blendPixelSafe(buf, cx-runeSize, runeY, arcaneColor)
	blendPixelSafe(buf, cx, runeY, arcaneColor) // center bright
	// Glow halo around rune
	for dx := -runeSize - 1; dx <= runeSize+1; dx++ {
		for dy := -runeSize - 1; dy <= runeSize+1; dy++ {
			dist := math.Abs(float64(dx)) + math.Abs(float64(dy))
			if dist > float64(runeSize) && dist <= float64(runeSize+2) {
				blendPixelSafe(buf, cx+dx, runeY+dy, arcaneGlow)
			}
		}
	}

	// --- Glowing orb near the staff hand ---
	orbX, orbY := staffTipPosition(cx, cy, p.Direction, w, h)
	orbColor := color.RGBA{R: arcaneColor.R, G: arcaneColor.G, B: arcaneColor.B, A: 220}
	orbGlow := color.RGBA{R: arcaneColor.R, G: arcaneColor.G, B: arcaneColor.B, A: 70}
	setPixelSafe(buf, orbX, orbY, orbColor)
	// Soft glow around orb
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= 2.2 {
				alpha := uint8(float64(orbGlow.A) * (1.0 - dist/2.5))
				glow := color.RGBA{R: orbGlow.R, G: orbGlow.G, B: orbGlow.B, A: alpha}
				blendPixelSafe(buf, orbX+dx, orbY+dy, glow)
			}
		}
	}

	// --- Hat brim shadow crease (dark arc across top of head) ---
	hx, hy := headOffset(cx, cy, p.Direction, w, h)
	brimShadow := color.RGBA{R: 30, G: 20, B: 40, A: 80}
	brimW := maxDetail(2, w/6)
	for dx := -brimW; dx <= brimW; dx++ {
		blendPixelSafe(buf, hx+dx, hy+1, brimShadow)
	}

	// --- Subtle robe hem sparkles at bottom edge ---
	sparkle := color.RGBA{R: arcaneColor.R, G: arcaneColor.G, B: arcaneColor.B, A: 90}
	for i := 0; i < 3; i++ {
		sx := cx - w/4 + rng.Intn(maxDetail(1, w/2))
		sy := h - 2 - rng.Intn(2)
		if isOpaqueAt(buf, sx, sy) {
			blendPixelSafe(buf, sx, sy, sparkle)
		}
	}
}

// ----------------------------------------------------------------------------
// Warrior — weapon belt, arm wraps, battle scar
// ----------------------------------------------------------------------------

func renderWarriorDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	beltColor := color.RGBA{R: 90, G: 60, B: 30, A: 200}
	buckleColor := color.RGBA{R: 200, G: 180, B: 100, A: 230}
	wrapColor := color.RGBA{R: 120, G: 90, B: 60, A: 160}
	scarColor := color.RGBA{R: 180, G: 80, B: 70, A: 130}

	// --- Weapon belt across torso (horizontal dark line with buckle) ---
	beltY := cy + 2
	beltW := maxDetail(2, w/4)
	for dx := -beltW; dx <= beltW; dx++ {
		blendPixelSafe(buf, cx+dx, beltY, beltColor)
	}
	// Belt buckle: bright dot at center
	setPixelSafe(buf, cx, beltY, buckleColor)

	// --- Arm wraps (short diagonal marks on arm positions) ---
	armSpreadX := maxDetail(3, w*3/10)
	armY := cy
	// Left arm wraps
	blendPixelSafe(buf, cx-armSpreadX, armY, wrapColor)
	blendPixelSafe(buf, cx-armSpreadX, armY+1, wrapColor)
	blendPixelSafe(buf, cx-armSpreadX+1, armY, wrapColor)
	// Right arm wraps
	blendPixelSafe(buf, cx+armSpreadX, armY, wrapColor)
	blendPixelSafe(buf, cx+armSpreadX, armY+1, wrapColor)
	blendPixelSafe(buf, cx+armSpreadX-1, armY, wrapColor)

	// --- Battle scar (diagonal line on torso or arm) ---
	scarStart := rng.Intn(2) // 0 = torso scar, 1 = arm scar
	if scarStart == 0 {
		// Torso scar: short diagonal
		sy := cy - 1
		for i := 0; i < 3; i++ {
			blendPixelSafe(buf, cx-1+i, sy+i, scarColor)
		}
	} else {
		// Arm scar
		sy := armY - 1
		for i := 0; i < 2; i++ {
			blendPixelSafe(buf, cx-armSpreadX+i, sy+i, scarColor)
		}
	}

	// --- Subtle muscle highlight on shoulders ---
	muscleHL := color.RGBA{R: 220, G: 200, B: 180, A: 50}
	for dx := -1; dx <= 1; dx++ {
		blendPixelSafe(buf, cx+armSpreadX+dx, armY-1, muscleHL)
		blendPixelSafe(buf, cx-armSpreadX+dx, armY-1, muscleHL)
	}
}

// ----------------------------------------------------------------------------
// Knight — shoulder plate highlights, visor slit, chest emblem
// ----------------------------------------------------------------------------

func renderKnightDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	metalHL := color.RGBA{R: 230, G: 230, B: 240, A: 180}
	metalShadow := color.RGBA{R: 60, G: 60, B: 80, A: 120}
	emblemColor := knightEmblemColor(p.Genre, rng)

	// --- Shoulder plate highlights (bright dots on outer shoulders) ---
	shoulderX := maxDetail(3, w*3/8)
	shoulderY := cy - 1
	// Left plate highlight
	setPixelSafe(buf, cx-shoulderX, shoulderY, metalHL)
	setPixelSafe(buf, cx-shoulderX+1, shoulderY, metalHL)
	blendPixelSafe(buf, cx-shoulderX, shoulderY+1, metalShadow)
	// Right plate highlight
	setPixelSafe(buf, cx+shoulderX, shoulderY, metalHL)
	setPixelSafe(buf, cx+shoulderX-1, shoulderY, metalHL)
	blendPixelSafe(buf, cx+shoulderX, shoulderY+1, metalShadow)

	// --- Visor slit on helmet (horizontal dark line on head) ---
	hx, hy := headOffset(cx, cy, p.Direction, w, h)
	visorW := maxDetail(1, w/8)
	visorColor := color.RGBA{R: 20, G: 20, B: 30, A: 200}
	for dx := -visorW; dx <= visorW; dx++ {
		setPixelSafe(buf, hx+dx, hy, visorColor)
	}
	// Visor rim highlight above
	visorRim := color.RGBA{R: 180, G: 180, B: 200, A: 140}
	for dx := -visorW; dx <= visorW; dx++ {
		blendPixelSafe(buf, hx+dx, hy-1, visorRim)
	}

	// --- Chest emblem (cross or diamond on torso center) ---
	emblemY := cy + 1
	emblemType := rng.Intn(2)
	if emblemType == 0 {
		// Cross emblem
		for i := -1; i <= 1; i++ {
			blendPixelSafe(buf, cx+i, emblemY, emblemColor)
			blendPixelSafe(buf, cx, emblemY+i, emblemColor)
		}
	} else {
		// Diamond emblem
		blendPixelSafe(buf, cx, emblemY-1, emblemColor)
		blendPixelSafe(buf, cx-1, emblemY, emblemColor)
		blendPixelSafe(buf, cx+1, emblemY, emblemColor)
		blendPixelSafe(buf, cx, emblemY+1, emblemColor)
	}

	// --- Armor edge highlight along torso border ---
	armorEdge := color.RGBA{R: 200, G: 200, B: 220, A: 60}
	torsoW := maxDetail(2, w/4)
	for dy := cy - 2; dy <= cy+3; dy++ {
		blendPixelSafe(buf, cx-torsoW, dy, armorEdge)
		blendPixelSafe(buf, cx+torsoW, dy, armorEdge)
	}
	_ = h // suppress unused
}

// ----------------------------------------------------------------------------
// Rogue — hood shadow, dagger glint, shadow wisps
// ----------------------------------------------------------------------------

func renderRogueDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	shadowColor := color.RGBA{R: 20, G: 15, B: 30, A: 100}
	daggerGlint := color.RGBA{R: 220, G: 230, B: 240, A: 210}
	wispColor := color.RGBA{R: 40, G: 30, B: 60, A: 70}

	// --- Hood shadow (dark ring around head, deeper at front) ---
	hx, hy := headOffset(cx, cy, p.Direction, w, h)
	hoodR := maxDetail(2, w/6)
	for angle := 0.0; angle < 2*math.Pi; angle += 0.3 {
		rx := hx + int(float64(hoodR)*math.Cos(angle))
		ry := hy + int(float64(hoodR)*math.Sin(angle))
		// Deeper shadow at front of hood
		front := isHoodFront(angle, p.Direction)
		if front {
			setPixelSafe(buf, rx, ry, color.RGBA{R: 15, G: 10, B: 25, A: 140})
		} else {
			blendPixelSafe(buf, rx, ry, shadowColor)
		}
	}

	// --- Dagger glint at belt position ---
	daggerX, daggerY := daggerPosition(cx, cy, p.Direction, w, h)
	setPixelSafe(buf, daggerX, daggerY, daggerGlint)
	// Faint blade line
	bladeDir := bladeDirection(p.Direction)
	setPixelSafe(buf, daggerX+bladeDir[0], daggerY+bladeDir[1],
		color.RGBA{R: 180, G: 190, B: 200, A: 160})
	setPixelSafe(buf, daggerX+bladeDir[0]*2, daggerY+bladeDir[1]*2,
		color.RGBA{R: 150, G: 160, B: 170, A: 100})

	// --- Shadow wisps trailing behind ---
	tx, ty := roleRearPosition(cx, cy, p.Direction, w, h)
	wispCount := 3 + rng.Intn(3)
	for i := 0; i < wispCount; i++ {
		wx := tx + rng.Intn(5) - 2
		wy := ty + rng.Intn(3) - 1
		if wx >= 0 && wx < w && wy >= 0 && wy < h {
			blendPixelSafe(buf, wx, wy, wispColor)
		}
	}
	_ = h // suppress unused
}

// ----------------------------------------------------------------------------
// Merchant — belt pouches, gold sparkle, apron hem
// ----------------------------------------------------------------------------

func renderMerchantDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	pouchColor := color.RGBA{R: 100, G: 70, B: 40, A: 200}
	pouchShadow := color.RGBA{R: 60, G: 40, B: 25, A: 150}
	goldColor := color.RGBA{R: 255, G: 220, B: 80, A: 220}
	apronColor := color.RGBA{R: 200, G: 190, B: 170, A: 120}

	// --- Belt pouches (two small bumps at waist sides) ---
	beltY := cy + 2
	pouchOff := maxDetail(2, w/5)
	// Left pouch
	setPixelSafe(buf, cx-pouchOff, beltY, pouchColor)
	setPixelSafe(buf, cx-pouchOff, beltY+1, pouchColor)
	blendPixelSafe(buf, cx-pouchOff+1, beltY+1, pouchShadow)
	// Right pouch
	setPixelSafe(buf, cx+pouchOff, beltY, pouchColor)
	setPixelSafe(buf, cx+pouchOff, beltY+1, pouchColor)
	blendPixelSafe(buf, cx+pouchOff-1, beltY+1, pouchShadow)

	// --- Gold coin sparkle (1-2 bright yellow pixels near pouch) ---
	goldCount := 1 + rng.Intn(2)
	for i := 0; i < goldCount; i++ {
		side := -1
		if rng.Intn(2) == 0 {
			side = 1
		}
		gx := cx + side*pouchOff + rng.Intn(3) - 1
		gy := beltY + rng.Intn(2)
		blendPixelSafe(buf, gx, gy, goldColor)
	}

	// --- Apron front (lighter band down center of torso) ---
	apronTop := cy - 1
	apronBot := cy + 4
	apronW := maxDetail(1, w/8)
	for dy := apronTop; dy <= apronBot && dy < h; dy++ {
		for dx := -apronW; dx <= apronW; dx++ {
			if isOpaqueAt(buf, cx+dx, dy) {
				blendPixelSafe(buf, cx+dx, dy, apronColor)
			}
		}
	}

	// --- Apron hem line (slightly darker line at bottom of apron) ---
	hemColor := color.RGBA{R: 140, G: 130, B: 110, A: 100}
	for dx := -apronW - 1; dx <= apronW+1; dx++ {
		blendPixelSafe(buf, cx+dx, apronBot, hemColor)
	}
}

// ----------------------------------------------------------------------------
// Ranger — quiver detail on back, leaf accents, nature emblem
// ----------------------------------------------------------------------------

func renderRangerDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	quiverColor := color.RGBA{R: 100, G: 70, B: 40, A: 190}
	arrowTip := color.RGBA{R: 180, G: 180, B: 190, A: 200}
	leafColor := color.RGBA{R: 60, G: 140, B: 50, A: 130}
	leafDark := color.RGBA{R: 40, G: 100, B: 35, A: 100}

	// --- Quiver on back (diagonal lines on one shoulder, arrow tips peeking) ---
	qx, qy := quiverPosition(cx, cy, p.Direction, w, h)
	// Quiver body: 2-3 pixel vertical strip
	for dy := 0; dy < maxDetail(2, h/8); dy++ {
		setPixelSafe(buf, qx, qy+dy, quiverColor)
		if dy > 0 {
			setPixelSafe(buf, qx+1, qy+dy, quiverColor)
		}
	}
	// Arrow tips peeking above
	setPixelSafe(buf, qx, qy-1, arrowTip)
	setPixelSafe(buf, qx+1, qy-1, arrowTip)
	blendPixelSafe(buf, qx-1, qy-1, arrowTip)

	// --- Leaf accents (green pixels scattered at edges of clothing) ---
	leafCount := 2 + rng.Intn(3)
	for i := 0; i < leafCount; i++ {
		lx := cx - w/3 + rng.Intn(maxDetail(1, 2*w/3))
		ly := cy - h/4 + rng.Intn(maxDetail(1, h/2))
		if isOpaqueAt(buf, lx, ly) && isRoleEdgePixel(buf, lx, ly) {
			blendPixelSafe(buf, lx, ly, leafColor)
			// Leaf shadow
			blendPixelSafe(buf, lx+1, ly, leafDark)
		}
	}

	// --- Nature emblem on chest (small circle/dot) ---
	emblemY := cy
	natureEmblem := color.RGBA{R: 70, G: 160, B: 60, A: 170}
	setPixelSafe(buf, cx, emblemY, natureEmblem)
	blendPixelSafe(buf, cx-1, emblemY, leafDark)
	blendPixelSafe(buf, cx+1, emblemY, leafDark)
	_ = h // suppress unused
}

// ----------------------------------------------------------------------------
// Priest — halo rim above head, holy symbol, robe hem glow
// ----------------------------------------------------------------------------

func renderPriestDetails(buf *image.RGBA, p RoleDetailParams, rng *rand.Rand) {
	w, h := p.Width, p.Height
	cx, cy := w/2, h/2

	holyColor := priestHolyColor(p.Genre, rng)
	holyGlow := color.RGBA{R: holyColor.R, G: holyColor.G, B: holyColor.B, A: 60}

	// --- Halo rim (bright arc above head) ---
	hx, hy := headOffset(cx, cy, p.Direction, w, h)
	haloR := maxDetail(2, w/5)
	// Draw upper arc only (semicircle above head)
	for angle := math.Pi; angle <= 2*math.Pi; angle += 0.25 {
		rx := hx + int(float64(haloR)*math.Cos(angle))
		ry := hy - 1 + int(float64(haloR)*math.Sin(angle))
		setPixelSafe(buf, rx, ry, holyColor)
	}
	// Glow around halo
	for angle := math.Pi; angle <= 2*math.Pi; angle += 0.3 {
		outerR := float64(haloR) + 1.0
		rx := hx + int(outerR*math.Cos(angle))
		ry := hy - 1 + int(outerR*math.Sin(angle))
		blendPixelSafe(buf, rx, ry, holyGlow)
	}

	// --- Holy symbol on chest (cross or star based on seed) ---
	symbolY := cy + 1
	symbolType := rng.Intn(2)
	if symbolType == 0 {
		// Cross
		setPixelSafe(buf, cx, symbolY-1, holyColor)
		setPixelSafe(buf, cx, symbolY, holyColor)
		setPixelSafe(buf, cx, symbolY+1, holyColor)
		setPixelSafe(buf, cx-1, symbolY, holyColor)
		setPixelSafe(buf, cx+1, symbolY, holyColor)
	} else {
		// Star (4-point)
		setPixelSafe(buf, cx, symbolY-1, holyColor)
		setPixelSafe(buf, cx, symbolY+1, holyColor)
		setPixelSafe(buf, cx-1, symbolY, holyColor)
		setPixelSafe(buf, cx+1, symbolY, holyColor)
		// Diagonal glow
		blendPixelSafe(buf, cx-1, symbolY-1, holyGlow)
		blendPixelSafe(buf, cx+1, symbolY-1, holyGlow)
		blendPixelSafe(buf, cx-1, symbolY+1, holyGlow)
		blendPixelSafe(buf, cx+1, symbolY+1, holyGlow)
	}

	// --- Robe hem glow (faint light at bottom of robes) ---
	hemY := h - 2
	hemW := maxDetail(2, w/4)
	for dx := -hemW; dx <= hemW; dx++ {
		if isOpaqueAt(buf, cx+dx, hemY) {
			blendPixelSafe(buf, cx+dx, hemY, holyGlow)
		}
		if isOpaqueAt(buf, cx+dx, hemY-1) {
			dimGlow := color.RGBA{R: holyColor.R, G: holyColor.G, B: holyColor.B, A: 30}
			blendPixelSafe(buf, cx+dx, hemY-1, dimGlow)
		}
	}
}

// ============================================================================
// Helper functions for role details
// ============================================================================

// maxDetail returns the larger of two ints (local helper to avoid name conflicts).
func maxDetail(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// mageArcaneColor returns genre-aware arcane magic color.
func mageArcaneColor(genre string, rng *rand.Rand) color.RGBA {
	switch genre {
	case "horror":
		return color.RGBA{R: 120, G: 40, B: 180, A: 200} // dark purple
	case "cyberpunk":
		return color.RGBA{R: 0, G: 220, B: 255, A: 200} // neon cyan
	case "sci-fi", "scifi":
		return color.RGBA{R: 100, G: 180, B: 255, A: 200} // soft blue
	case "post-apocalyptic", "postapoc":
		return color.RGBA{R: 200, G: 150, B: 50, A: 200} // amber
	default: // fantasy
		hueOff := 20
		if rng != nil {
			hueOff = rng.Intn(40)
		}
		return color.RGBA{R: 100 + uint8(hueOff), G: 80, B: 220, A: 200} // blue-violet
	}
}

// knightEmblemColor returns genre-aware emblem color.
func knightEmblemColor(genre string, rng *rand.Rand) color.RGBA {
	switch genre {
	case "horror":
		return color.RGBA{R: 180, G: 30, B: 30, A: 180} // blood red
	case "cyberpunk":
		return color.RGBA{R: 255, G: 200, B: 0, A: 180} // gold-yellow
	default:
		heraldic := []color.RGBA{
			{R: 200, G: 50, B: 50, A: 180},  // red
			{R: 50, G: 80, B: 200, A: 180},   // blue
			{R: 220, G: 200, B: 50, A: 180},  // gold
			{R: 200, G: 200, B: 210, A: 180}, // silver
		}
		idx := 0
		if rng != nil {
			idx = rng.Intn(len(heraldic))
		}
		return heraldic[idx]
	}
}

// priestHolyColor returns genre-aware holy light color.
func priestHolyColor(genre string, rng *rand.Rand) color.RGBA {
	switch genre {
	case "horror":
		return color.RGBA{R: 200, G: 200, B: 220, A: 190} // pale silver
	case "cyberpunk":
		return color.RGBA{R: 255, G: 100, B: 255, A: 190} // magenta
	case "sci-fi", "scifi":
		return color.RGBA{R: 180, G: 220, B: 255, A: 190} // cool white
	default:
		_ = rng
		return color.RGBA{R: 255, G: 235, B: 150, A: 190} // warm gold
	}
}

// staffTipPosition returns the position of a staff/wand tip based on direction.
func staffTipPosition(cx, cy int, direction string, w, h int) (int, int) {
	armOff := maxDetail(3, w*3/10)
	switch direction {
	case "up":
		return cx + armOff, cy - maxDetail(2, h/5)
	case "down":
		return cx - armOff, cy + maxDetail(2, h/5)
	case "left":
		return cx - maxDetail(2, w/5), cy - armOff
	case "right":
		return cx + maxDetail(2, w/5), cy - armOff
	default:
		return cx + armOff, cy - maxDetail(2, h/4)
	}
}

// daggerPosition returns the dagger location at the belt on the off-hand side.
func daggerPosition(cx, cy int, direction string, w, h int) (int, int) {
	beltOff := maxDetail(2, w/5)
	switch direction {
	case "left":
		return cx + beltOff, cy + 1
	case "right":
		return cx - beltOff, cy + 1
	default:
		return cx + beltOff, cy + 2
	}
}

// bladeDirection returns the (dx, dy) direction for a dagger blade line.
func bladeDirection(direction string) [2]int {
	switch direction {
	case "up":
		return [2]int{0, -1}
	case "down":
		return [2]int{0, 1}
	case "left":
		return [2]int{-1, 0}
	case "right":
		return [2]int{1, 0}
	default:
		return [2]int{0, -1}
	}
}

// roleRearPosition returns the position behind the entity (for trailing effects).
func roleRearPosition(cx, cy int, direction string, w, h int) (int, int) {
	off := maxDetail(2, h/4)
	switch direction {
	case "up":
		return cx, cy + off
	case "down":
		return cx, cy - off
	case "left":
		return cx + off, cy
	case "right":
		return cx - off, cy
	default:
		return cx, cy + off
	}
}

// quiverPosition returns the quiver placement on the back.
func quiverPosition(cx, cy int, direction string, w, h int) (int, int) {
	off := maxDetail(2, w/4)
	switch direction {
	case "up":
		return cx + off, cy + 1
	case "down":
		return cx - off, cy - 1
	case "left":
		return cx + 1, cy + off
	case "right":
		return cx - 1, cy - off
	default:
		return cx + off, cy
	}
}

// isHoodFront returns true if the angle is at the front of the hood for the given direction.
func isHoodFront(angle float64, direction string) bool {
	switch direction {
	case "up":
		return angle >= math.Pi && angle <= 2*math.Pi
	case "down":
		return angle >= 0 && angle <= math.Pi
	case "left":
		return angle >= math.Pi/2 && angle <= 3*math.Pi/2
	case "right":
		return angle <= math.Pi/2 || angle >= 3*math.Pi/2
	default:
		return angle >= math.Pi && angle <= 2*math.Pi
	}
}

// isRoleEdgePixel returns true if the pixel at (x,y) has at least one transparent neighbor.
func isRoleEdgePixel(buf *image.RGBA, x, y int) bool {
	if !isOpaqueAt(buf, x, y) {
		return false
	}
	return !isOpaqueAt(buf, x-1, y) || !isOpaqueAt(buf, x+1, y) ||
		!isOpaqueAt(buf, x, y-1) || !isOpaqueAt(buf, x, y+1)
}
