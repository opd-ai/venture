// Package sprites provides seed-based headgear rendering for top-down aerial
// entity sprites. Different headgear types are immediately distinguishable from
// above, making headgear one of the highest-impact visual differentiators at
// 32×32. Genre and entity role influence which headgear types are selected.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// HeadgearType represents a distinct headgear style visible from a top-down camera.
type HeadgearType int

const (
	// HeadgearNone means no headgear is worn.
	HeadgearNone HeadgearType = iota
	// HeadgearCirclet is a thin metallic band around the head with optional gem.
	HeadgearCirclet
	// HeadgearCrown is a wide band with raised points/peaks around it.
	HeadgearCrown
	// HeadgearWizardHat is a pointed cone visible as a circle with radial gradient.
	HeadgearWizardHat
	// HeadgearHood is a fabric drape covering head and trailing behind.
	HeadgearHood
	// HeadgearHornedHelm has a dome with two prominent horn protrusions.
	HeadgearHornedHelm
	// HeadgearWideBrim is a large circular hat (ranger/witch/pilgrim).
	HeadgearWideBrim
	// HeadgearTurban is wrapped fabric wider than the head.
	HeadgearTurban
	// HeadgearSkullCap is a small close-fitting cap on top of the head.
	HeadgearSkullCap
	// HeadgearFullHelm is a heavy closed helmet covering the entire head.
	HeadgearFullHelm
	// HeadgearPlumed has a crest/plume trailing from a helmet.
	HeadgearPlumed
	// HeadgearTiara is a half-ring with upward points on the front.
	HeadgearTiara
	// HeadgearBandana is a tied cloth with trailing knot.
	HeadgearBandana
	// HeadgearCount is the total number of headgear types (must be last).
	HeadgearCount
)

// String returns the display name of a headgear type.
func (h HeadgearType) String() string {
	names := [...]string{
		"none", "circlet", "crown", "wizard_hat", "hood",
		"horned_helm", "wide_brim", "turban", "skull_cap",
		"full_helm", "plumed", "tiara", "bandana",
	}
	if int(h) < len(names) {
		return names[h]
	}
	return "unknown"
}

// HeadgearRenderParams controls how a headgear overlay is drawn onto the head area.
type HeadgearRenderParams struct {
	// SpriteWidth and SpriteHeight are the full sprite dimensions.
	SpriteWidth  int
	SpriteHeight int

	// HeadSpec defines the head body part position from the anatomy template.
	HeadSpec PartSpec

	// Type is the headgear style to render.
	Type HeadgearType

	// PrimaryColor is the main headgear color.
	PrimaryColor color.RGBA

	// AccentColor is the secondary/trim color.
	AccentColor color.RGBA

	// GemColor is used for jewels on circlets, crowns, and tiaras.
	GemColor color.RGBA

	// Direction the entity is facing.
	Direction Direction

	// MaterialSheen controls specular highlight strength (0.0-1.0).
	MaterialSheen float64

	// Seed for deterministic variation.
	Seed int64
}

// ComputeHeadgearParams derives render params from entity seed, role, and genre.
func ComputeHeadgearParams(spriteW, spriteH int, headSpec PartSpec, hType HeadgearType, direction Direction, seed int64, genre string) HeadgearRenderParams {
	rng := rand.New(rand.NewSource(seed ^ 0x48EA6))

	// Genre-influenced color selection
	primary, accent, gem := headgearColorsForGenre(rng, genre)

	sheen := 0.3 + rng.Float64()*0.5

	return HeadgearRenderParams{
		SpriteWidth:   spriteW,
		SpriteHeight:  spriteH,
		HeadSpec:      headSpec,
		Type:          hType,
		PrimaryColor:  primary,
		AccentColor:   accent,
		GemColor:      gem,
		Direction:     direction,
		MaterialSheen: sheen,
		Seed:          seed,
	}
}

// SelectHeadgearForRole returns a headgear type appropriate for an entity role and genre.
func SelectHeadgearForRole(role, genre string, seed int64) HeadgearType {
	rng := rand.New(rand.NewSource(seed ^ 0x9AE1F))

	candidates := headgearCandidatesForRole(role)
	if len(candidates) == 0 {
		return HeadgearNone
	}

	// Genre can filter or reweight candidates
	filtered := filterByGenre(candidates, genre, rng)
	if len(filtered) == 0 {
		filtered = candidates
	}

	return filtered[rng.Intn(len(filtered))]
}

// headgearCandidatesForRole returns headgear types appropriate for each role.
func headgearCandidatesForRole(role string) []HeadgearType {
	switch role {
	case "mage", "wizard", "elementalist", "necromancer", "enchanter":
		return []HeadgearType{HeadgearWizardHat, HeadgearHood, HeadgearCirclet, HeadgearTiara}
	case "warrior", "knight", "paladin", "berserker":
		return []HeadgearType{HeadgearFullHelm, HeadgearHornedHelm, HeadgearPlumed, HeadgearSkullCap}
	case "rogue", "assassin", "ninja":
		return []HeadgearType{HeadgearHood, HeadgearBandana, HeadgearSkullCap}
	case "ranger", "druid":
		return []HeadgearType{HeadgearWideBrim, HeadgearHood, HeadgearBandana, HeadgearCirclet}
	case "priest", "cleric":
		return []HeadgearType{HeadgearTiara, HeadgearCirclet, HeadgearHood, HeadgearTurban}
	case "merchant":
		return []HeadgearType{HeadgearWideBrim, HeadgearTurban, HeadgearSkullCap, HeadgearBandana}
	case "bard":
		return []HeadgearType{HeadgearWideBrim, HeadgearBandana, HeadgearPlumed}
	case "boss":
		return []HeadgearType{HeadgearCrown, HeadgearHornedHelm, HeadgearPlumed, HeadgearFullHelm}
	default:
		// Generic NPCs get lighter headgear
		return []HeadgearType{HeadgearNone, HeadgearBandana, HeadgearSkullCap, HeadgearCirclet, HeadgearWideBrim}
	}
}

// filterByGenre adjusts headgear candidates based on genre.
func filterByGenre(candidates []HeadgearType, genre string, rng *rand.Rand) []HeadgearType {
	var filtered []HeadgearType
	for _, c := range candidates {
		if headgearFitsGenre(c, genre) {
			filtered = append(filtered, c)
		}
	}
	// If filtering removed everything, allow a random pick from any candidate
	if len(filtered) == 0 {
		_ = rng
		return candidates
	}
	return filtered
}

// headgearFitsGenre returns true if a headgear type is thematically appropriate.
func headgearFitsGenre(h HeadgearType, genre string) bool {
	switch genre {
	case "fantasy":
		return true // All types fit fantasy
	case "horror":
		// Prefer dark/skull/hood, exclude flashy crowns
		return h != HeadgearCrown && h != HeadgearTiara
	case "sci-fi", "scifi":
		// Prefer sleek: circlets, skull caps, full helms
		return h == HeadgearCirclet || h == HeadgearSkullCap ||
			h == HeadgearFullHelm || h == HeadgearBandana || h == HeadgearNone
	case "cyberpunk":
		// Tech-adjacent: circlets, skull caps, hoods, bandanas
		return h == HeadgearCirclet || h == HeadgearSkullCap ||
			h == HeadgearHood || h == HeadgearBandana || h == HeadgearNone
	case "post-apocalyptic", "postapoc":
		// Scrappy: bandanas, hoods, skull caps, horned helms
		return h == HeadgearBandana || h == HeadgearHood ||
			h == HeadgearSkullCap || h == HeadgearHornedHelm || h == HeadgearNone
	default:
		return true
	}
}

// RenderHeadgearOverlay draws headgear onto the given RGBA buffer.
func RenderHeadgearOverlay(dst *image.RGBA, p HeadgearRenderParams) {
	if p.Type == HeadgearNone {
		return
	}

	// Compute head center and radius in sprite pixel coordinates
	headCX := int(float64(p.SpriteWidth) * p.HeadSpec.RelativeX)
	headCY := int(float64(p.SpriteHeight) * p.HeadSpec.RelativeY)
	headW := p.HeadSpec.GetEffectiveWidth(p.SpriteWidth)
	headH := p.HeadSpec.GetEffectiveHeight(p.SpriteHeight)
	headR := maxHeadgear(headW, headH) / 2

	rng := rand.New(rand.NewSource(p.Seed ^ 0x3CF8A))

	switch p.Type {
	case HeadgearCirclet:
		renderCirclet(dst, headCX, headCY, headR, p, rng)
	case HeadgearCrown:
		renderCrown(dst, headCX, headCY, headR, p, rng)
	case HeadgearWizardHat:
		renderWizardHat(dst, headCX, headCY, headR, p, rng)
	case HeadgearHood:
		renderHoodGear(dst, headCX, headCY, headR, p, rng)
	case HeadgearHornedHelm:
		renderHornedHelm(dst, headCX, headCY, headR, p, rng)
	case HeadgearWideBrim:
		renderWideBrim(dst, headCX, headCY, headR, p, rng)
	case HeadgearTurban:
		renderTurban(dst, headCX, headCY, headR, p, rng)
	case HeadgearSkullCap:
		renderSkullCap(dst, headCX, headCY, headR, p, rng)
	case HeadgearFullHelm:
		renderFullHelm(dst, headCX, headCY, headR, p, rng)
	case HeadgearPlumed:
		renderPlumedHelm(dst, headCX, headCY, headR, p, rng)
	case HeadgearTiara:
		renderTiara(dst, headCX, headCY, headR, p, rng)
	case HeadgearBandana:
		renderBandana(dst, headCX, headCY, headR, p, rng)
	}
}

// --- Individual headgear renderers ---

// renderCirclet draws a thin metallic band around the head with 1-2 small gems.
func renderCirclet(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	bandR := headR + 1
	highlight := lightenHeadgear(p.PrimaryColor, 50+int(p.MaterialSheen*40))

	for angle := 0.0; angle < 2*math.Pi; angle += 0.08 {
		px := int(float64(cx) + float64(bandR)*math.Cos(angle))
		py := int(float64(cy) + float64(bandR)*0.85*math.Sin(angle))
		if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
			// Slight shading variation around the band
			t := (math.Cos(angle*2) + 1) / 2
			c := blendHeadgear(p.PrimaryColor, highlight, float64(t)*0.4)
			setPixelAlpha(dst, px, py, c, 220)
		}
	}

	// Gem at the front of the circlet
	gemAngle := directionAngleOffset(p.Direction)
	gemX := int(float64(cx) + float64(bandR)*math.Cos(gemAngle))
	gemY := int(float64(cy) + float64(bandR)*0.85*math.Sin(gemAngle))
	drawGem(dst, gemX, gemY, 1, p.GemColor, p.SpriteWidth, p.SpriteHeight)
}

// renderCrown draws a wide band with 3-5 raised triangular peaks.
func renderCrown(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	bandR := headR + 1
	bandThickness := maxHeadgear(1, headR/3)
	highlight := lightenHeadgear(p.PrimaryColor, 40+int(p.MaterialSheen*35))

	// Thick band around head
	for angle := 0.0; angle < 2*math.Pi; angle += 0.06 {
		for t := 0; t < bandThickness; t++ {
			r := float64(bandR) - float64(t)
			px := int(float64(cx) + r*math.Cos(angle))
			py := int(float64(cy) + r*0.85*math.Sin(angle))
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				shade := float64(t) / float64(maxHeadgear(1, bandThickness))
				c := blendHeadgear(highlight, p.PrimaryColor, shade*0.5)
				setPixelAlpha(dst, px, py, c, 240)
			}
		}
	}

	// Crown peaks (3-5 points rising outward from band)
	numPeaks := 3 + rng.Intn(3)
	peakHeight := maxHeadgear(2, headR/2)
	for i := 0; i < numPeaks; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numPeaks)
		baseX := float64(cx) + float64(bandR)*math.Cos(angle)
		baseY := float64(cy) + float64(bandR)*0.85*math.Sin(angle)
		tipX := float64(cx) + float64(bandR+peakHeight)*math.Cos(angle)
		tipY := float64(cy) + float64(bandR+peakHeight)*0.85*math.Sin(angle)

		// Draw peak as 1px wide line from base to tip
		steps := peakHeight + 1
		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(maxHeadgear(1, steps))
			px := int(baseX + (tipX-baseX)*t)
			py := int(baseY + (tipY-baseY)*t)
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				c := blendHeadgear(highlight, p.AccentColor, t*0.6)
				setPixelAlpha(dst, px, py, c, 230)
			}
		}

		// Gem at every other peak tip
		if i%2 == 0 {
			drawGem(dst, int(tipX), int(tipY), 1, p.GemColor, p.SpriteWidth, p.SpriteHeight)
		}
	}
}

// renderWizardHat draws a pointed cone visible from above as a large circle
// with a radial gradient from brim to a bright center point.
func renderWizardHat(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	brimR := headR + maxHeadgear(2, headR/2)
	brimColor := darkenHeadgear(p.PrimaryColor, 0.7)
	tipColor := lightenHeadgear(p.PrimaryColor, 40+int(p.MaterialSheen*30))

	// Brim: large filled circle bigger than head
	for dy := -brimR; dy <= brimR; dy++ {
		for dx := -brimR; dx <= brimR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(brimR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(brimR)
					// Cone from above: bright center (tip), dark edge (brim)
					c := blendHeadgear(tipColor, brimColor, t)
					setPixelAlpha(dst, px, py, c, 210)
				}
			}
		}
	}

	// Decorative band near the brim edge
	bandR := brimR - 1
	for angle := 0.0; angle < 2*math.Pi; angle += 0.08 {
		px := int(float64(cx) + float64(bandR)*math.Cos(angle))
		py := int(float64(cy) + float64(bandR)*math.Sin(angle))
		if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
			setPixelAlpha(dst, px, py, p.AccentColor, 200)
		}
	}

	// Bright tip highlight at center
	if inBoundsHeadgear(cx, cy, p.SpriteWidth, p.SpriteHeight) {
		setPixelAlpha(dst, cx, cy, lightenHeadgear(tipColor, 60), 255)
	}

	_ = rng
}

// renderHoodGear draws a fabric hood draping over the head with trailing back.
func renderHoodGear(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	hoodR := headR + 2
	shadowColor := darkenHeadgear(p.PrimaryColor, 0.55)

	// Hood dome — darker at edges to show depth
	for dy := -hoodR; dy <= hoodR; dy++ {
		for dx := -hoodR; dx <= hoodR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(hoodR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(hoodR)
					c := blendHeadgear(p.PrimaryColor, shadowColor, t*0.7)
					setPixelAlpha(dst, px, py, c, 200)
				}
			}
		}
	}

	// Trailing back of hood (extends behind depending on direction)
	trailDX, trailDY := directionTrailVector(p.Direction)
	trailLen := maxHeadgear(3, headR)
	for i := 1; i <= trailLen; i++ {
		t := float64(i) / float64(trailLen)
		halfW := maxHeadgear(1, int(float64(hoodR)*(1.0-t*0.6)))
		for w := -halfW; w <= halfW; w++ {
			px := cx + trailDX*i + w*(1-absInt(trailDX))
			py := cy + trailDY*i + w*(1-absInt(trailDY))
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				edgeFade := float64(absInt(w)) / float64(maxHeadgear(1, halfW))
				c := blendHeadgear(shadowColor, darkenHeadgear(shadowColor, 0.8), edgeFade*0.4+t*0.3)
				setPixelAlpha(dst, px, py, c, uint8(200-int(t*80)))
			}
		}
	}

	_ = rng
}

// renderHornedHelm draws a helmet dome with two horn protrusions extending outward.
func renderHornedHelm(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	// Helmet dome base
	domeR := headR + 1
	highlight := lightenHeadgear(p.PrimaryColor, 30+int(p.MaterialSheen*35))
	for dy := -domeR; dy <= domeR; dy++ {
		for dx := -domeR; dx <= domeR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(domeR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(domeR)
					c := blendHeadgear(highlight, p.PrimaryColor, t*0.5)
					setPixelAlpha(dst, px, py, c, 230)
				}
			}
		}
	}

	// Two horns extending outward from the sides
	hornLen := maxHeadgear(3, headR+1)
	hornColor := p.AccentColor
	hornHighlight := lightenHeadgear(hornColor, 30)

	for side := -1; side <= 1; side += 2 {
		// Horns curve slightly upward from sides
		baseX := float64(cx) + float64(side*domeR)*0.7
		baseY := float64(cy) - float64(domeR)*0.3

		for i := 0; i < hornLen; i++ {
			t := float64(i) / float64(hornLen)
			// Horns spread outward and slightly forward
			hx := int(baseX + float64(side)*float64(i)*1.2)
			hy := int(baseY - float64(i)*0.8)

			if inBoundsHeadgear(hx, hy, p.SpriteWidth, p.SpriteHeight) {
				c := blendHeadgear(hornHighlight, hornColor, t*0.6)
				setPixelAlpha(dst, hx, hy, c, uint8(240-int(t*60)))
			}
			// Horn thickness (2px near base, 1px at tip)
			if t < 0.6 && inBoundsHeadgear(hx, hy+1, p.SpriteWidth, p.SpriteHeight) {
				setPixelAlpha(dst, hx, hy+1, hornColor, uint8(200-int(t*80)))
			}
		}
	}

	_ = rng
}

// renderWideBrim draws a large flat circular hat (ranger/witch/pilgrim style).
func renderWideBrim(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	brimR := headR + maxHeadgear(3, headR)
	crownR := headR - 1
	shadow := darkenHeadgear(p.PrimaryColor, 0.65)
	highlight := lightenHeadgear(p.PrimaryColor, 25)

	// Brim — large flat circle with slight shading
	for dy := -brimR; dy <= brimR; dy++ {
		for dx := -brimR; dx <= brimR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(brimR) && dist > float64(crownR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := (dist - float64(crownR)) / float64(maxHeadgear(1, brimR-crownR))
					c := blendHeadgear(p.PrimaryColor, shadow, t*0.5)
					setPixelAlpha(dst, px, py, c, 200)
				}
			}
		}
	}

	// Inner crown — slightly raised center
	for dy := -crownR; dy <= crownR; dy++ {
		for dx := -crownR; dx <= crownR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(crownR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(maxHeadgear(1, crownR))
					c := blendHeadgear(highlight, p.PrimaryColor, t*0.5)
					setPixelAlpha(dst, px, py, c, 220)
				}
			}
		}
	}

	// Hatband
	bandR := crownR + 1
	for angle := 0.0; angle < 2*math.Pi; angle += 0.08 {
		px := int(float64(cx) + float64(bandR)*math.Cos(angle))
		py := int(float64(cy) + float64(bandR)*math.Sin(angle))
		if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
			setPixelAlpha(dst, px, py, p.AccentColor, 220)
		}
	}

	_ = rng
}

// renderTurban draws a wrapped fabric circle wider than the head with fold lines.
func renderTurban(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	turbanR := headR + maxHeadgear(2, headR/2+1)
	foldColor := darkenHeadgear(p.PrimaryColor, 0.8)
	highlight := lightenHeadgear(p.PrimaryColor, 20)

	// Main turban body
	for dy := -turbanR; dy <= turbanR; dy++ {
		for dx := -turbanR; dx <= turbanR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(turbanR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(turbanR)
					c := blendHeadgear(highlight, p.PrimaryColor, t*0.4)
					setPixelAlpha(dst, px, py, c, 215)
				}
			}
		}
	}

	// Fabric fold lines (diagonal wrapping lines)
	numFolds := 3 + rng.Intn(3)
	foldSpacing := float64(turbanR*2) / float64(numFolds+1)
	for i := 1; i <= numFolds; i++ {
		offsetY := -turbanR + int(float64(i)*foldSpacing)
		for dx := -turbanR; dx <= turbanR; dx++ {
			px := cx + dx
			py := cy + offsetY + dx/3
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				distFromCenter := math.Sqrt(float64(dx*dx + offsetY*offsetY))
				if distFromCenter < float64(turbanR) {
					setPixelAlpha(dst, px, py, foldColor, 160)
				}
			}
		}
	}

	// Center ornament/jewel
	drawGem(dst, cx, cy, 1, p.GemColor, p.SpriteWidth, p.SpriteHeight)
}

// renderSkullCap draws a small close-fitting cap on top of the head.
func renderSkullCap(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	capR := headR - maxHeadgear(0, headR/4)
	if capR < 2 {
		capR = 2
	}
	highlight := lightenHeadgear(p.PrimaryColor, 30+int(p.MaterialSheen*25))

	// Cap covers the top half of the head
	for dy := -capR; dy <= capR/2; dy++ {
		for dx := -capR; dx <= capR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(capR) {
				px := cx + dx
				py := cy + dy - capR/4
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(capR)
					c := blendHeadgear(highlight, p.PrimaryColor, t*0.5)
					setPixelAlpha(dst, px, py, c, 200)
				}
			}
		}
	}

	// Rim line at the bottom edge
	rimY := cy + capR/2 - capR/4
	for dx := -capR; dx <= capR; dx++ {
		dist := math.Abs(float64(dx))
		if dist <= float64(capR) {
			px := cx + dx
			if inBoundsHeadgear(px, rimY, p.SpriteWidth, p.SpriteHeight) {
				setPixelAlpha(dst, px, rimY, p.AccentColor, 180)
			}
		}
	}

	_ = rng
}

// renderFullHelm draws a heavy closed helmet covering the entire head with eye slit.
func renderFullHelm(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	helmR := headR + 2
	highlight := lightenHeadgear(p.PrimaryColor, 35+int(p.MaterialSheen*40))
	shadow := darkenHeadgear(p.PrimaryColor, 0.6)

	// Full dome covering the head
	for dy := -helmR; dy <= helmR; dy++ {
		for dx := -helmR; dx <= helmR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(helmR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(helmR)
					c := blendHeadgear(highlight, p.PrimaryColor, t*0.6)
					// Edge darkening for heavy armor feel
					if t > 0.75 {
						c = blendHeadgear(c, shadow, (t-0.75)*4*0.4)
					}
					setPixelAlpha(dst, px, py, c, 240)
				}
			}
		}
	}

	// Central ridge/crest
	for dy := -helmR; dy <= 0; dy++ {
		py := cy + dy
		if inBoundsHeadgear(cx, py, p.SpriteWidth, p.SpriteHeight) {
			setPixelAlpha(dst, cx, py, highlight, 220)
		}
	}

	// T-shaped visor slit
	slitY := cy + helmR/3
	slitHalfW := helmR / 2
	slitColor := darkenHeadgear(p.PrimaryColor, 0.3)
	for dx := -slitHalfW; dx <= slitHalfW; dx++ {
		px := cx + dx
		if inBoundsHeadgear(px, slitY, p.SpriteWidth, p.SpriteHeight) {
			setPixelAlpha(dst, px, slitY, slitColor, 230)
		}
	}
	// Vertical nose guard
	for dy := slitY - 1; dy <= slitY+1; dy++ {
		if inBoundsHeadgear(cx, dy, p.SpriteWidth, p.SpriteHeight) {
			setPixelAlpha(dst, cx, dy, highlight, 200)
		}
	}

	_ = rng
}

// renderPlumedHelm draws a helmet with a feathered plume trailing behind.
func renderPlumedHelm(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	// Base helmet dome (smaller than full helm)
	domeR := headR + 1
	highlight := lightenHeadgear(p.PrimaryColor, 30+int(p.MaterialSheen*30))

	for dy := -domeR; dy <= domeR; dy++ {
		for dx := -domeR; dx <= domeR; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(domeR) {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
					t := dist / float64(domeR)
					c := blendHeadgear(highlight, p.PrimaryColor, t*0.5)
					setPixelAlpha(dst, px, py, c, 230)
				}
			}
		}
	}

	// Plume: colorful feather crest trailing behind the helmet
	plumeColor := p.AccentColor
	plumeHighlight := lightenHeadgear(plumeColor, 40)
	trailDX, trailDY := directionTrailVector(p.Direction)
	plumeLen := maxHeadgear(4, headR*2)

	for i := 0; i < plumeLen; i++ {
		t := float64(i) / float64(plumeLen)
		halfW := maxHeadgear(1, int(float64(domeR)*0.5*(1.0-t*0.7)))

		for w := -halfW; w <= halfW; w++ {
			px := cx + trailDX*i + w*(1-absInt(trailDX))
			py := cy + trailDY*i + w*(1-absInt(trailDY))
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				edgeFade := float64(absInt(w)) / float64(maxHeadgear(1, halfW))
				c := blendHeadgear(plumeHighlight, plumeColor, edgeFade*0.5+t*0.4)
				alpha := uint8(230 - int(t*100))
				setPixelAlpha(dst, px, py, c, alpha)
			}
		}
	}

	_ = rng
}

// renderTiara draws a half-ring on the front of the head with upward points.
func renderTiara(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	bandR := headR + 1
	highlight := lightenHeadgear(p.PrimaryColor, 50+int(p.MaterialSheen*40))

	// Half-ring arc on the front half of the head
	frontAngle := directionAngleOffset(p.Direction)
	arcSpan := math.Pi * 0.7

	for angle := frontAngle - arcSpan/2; angle <= frontAngle+arcSpan/2; angle += 0.06 {
		px := int(float64(cx) + float64(bandR)*math.Cos(angle))
		py := int(float64(cy) + float64(bandR)*0.85*math.Sin(angle))
		if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
			t := math.Abs(angle-frontAngle) / (arcSpan / 2)
			c := blendHeadgear(highlight, p.PrimaryColor, t*0.3)
			setPixelAlpha(dst, px, py, c, 230)
		}
	}

	// 3 upward points along the arc
	numPoints := 3
	for i := 0; i < numPoints; i++ {
		pointAngle := frontAngle - arcSpan/3 + float64(i)*arcSpan/3
		baseX := float64(cx) + float64(bandR)*math.Cos(pointAngle)
		baseY := float64(cy) + float64(bandR)*0.85*math.Sin(pointAngle)
		peakLen := maxHeadgear(2, headR/2)
		tipX := baseX + float64(peakLen)*math.Cos(pointAngle)
		tipY := baseY + float64(peakLen)*0.85*math.Sin(pointAngle)

		for s := 0; s <= peakLen; s++ {
			st := float64(s) / float64(maxHeadgear(1, peakLen))
			px := int(baseX + (tipX-baseX)*st)
			py := int(baseY + (tipY-baseY)*st)
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				c := blendHeadgear(highlight, p.AccentColor, st*0.5)
				setPixelAlpha(dst, px, py, c, 220)
			}
		}

		// Gem at the center point tip
		if i == 1 {
			drawGem(dst, int(tipX), int(tipY), 1, p.GemColor, p.SpriteWidth, p.SpriteHeight)
		}
	}

	_ = rng
}

// renderBandana draws a tied cloth around the head with a trailing knot.
func renderBandana(dst *image.RGBA, cx, cy, headR int, p HeadgearRenderParams, rng *rand.Rand) {
	bandR := headR
	bandWidth := maxHeadgear(1, headR/3)
	shadow := darkenHeadgear(p.PrimaryColor, 0.75)

	// Band wrapping around the head (full circle, cloth texture)
	for angle := 0.0; angle < 2*math.Pi; angle += 0.06 {
		for t := 0; t < bandWidth; t++ {
			r := float64(bandR) + float64(t)*0.5 - float64(bandWidth)*0.25
			px := int(float64(cx) + r*math.Cos(angle))
			py := int(float64(cy) + r*0.85*math.Sin(angle))
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				shade := float64(t) / float64(maxHeadgear(1, bandWidth))
				c := blendHeadgear(p.PrimaryColor, shadow, shade*0.3)
				setPixelAlpha(dst, px, py, c, 190)
			}
		}
	}

	// Trailing knot: two short tails extending from the back
	trailDX, trailDY := directionTrailVector(p.Direction)
	knotLen := maxHeadgear(2, headR)

	for side := -1; side <= 1; side += 2 {
		for i := 0; i < knotLen; i++ {
			t := float64(i) / float64(knotLen)
			sway := int(float64(side) * t * 2)
			px := cx + trailDX*(i+1) + sway*(1-absInt(trailDX))
			py := cy + trailDY*(i+1) + sway*(1-absInt(trailDY))
			if inBoundsHeadgear(px, py, p.SpriteWidth, p.SpriteHeight) {
				c := blendHeadgear(p.PrimaryColor, shadow, t*0.4)
				setPixelAlpha(dst, px, py, c, uint8(180-int(t*60)))
			}
		}
	}

	_ = rng
}

// --- Utility functions for headgear rendering ---

func headgearColorsForGenre(rng *rand.Rand, genre string) (primary, accent, gem color.RGBA) {
	switch genre {
	case "fantasy":
		primary = hslToRGBA(30+rng.Float64()*40, 0.3+rng.Float64()*0.4, 0.35+rng.Float64()*0.25)
		accent = hslToRGBA(40+rng.Float64()*20, 0.5+rng.Float64()*0.3, 0.55+rng.Float64()*0.2)
		gem = hslToRGBA(rng.Float64()*360, 0.7+rng.Float64()*0.3, 0.5+rng.Float64()*0.2)
	case "horror":
		primary = hslToRGBA(rng.Float64()*30, 0.1+rng.Float64()*0.2, 0.15+rng.Float64()*0.15)
		accent = hslToRGBA(0+rng.Float64()*20, 0.4+rng.Float64()*0.3, 0.2+rng.Float64()*0.15)
		gem = color.RGBA{R: uint8(180 + rng.Intn(75)), G: uint8(rng.Intn(40)), B: uint8(rng.Intn(40)), A: 255}
	case "sci-fi", "scifi":
		primary = hslToRGBA(200+rng.Float64()*60, 0.2+rng.Float64()*0.3, 0.45+rng.Float64()*0.2)
		accent = hslToRGBA(180+rng.Float64()*40, 0.5+rng.Float64()*0.4, 0.6+rng.Float64()*0.2)
		gem = color.RGBA{R: uint8(100 + rng.Intn(100)), G: uint8(200 + rng.Intn(55)), B: 255, A: 255}
	case "cyberpunk":
		primary = hslToRGBA(270+rng.Float64()*60, 0.3+rng.Float64()*0.4, 0.2+rng.Float64()*0.2)
		accent = hslToRGBA(300+rng.Float64()*60, 0.6+rng.Float64()*0.4, 0.5+rng.Float64()*0.3)
		gem = color.RGBA{R: uint8(200 + rng.Intn(55)), G: uint8(rng.Intn(80)), B: uint8(200 + rng.Intn(55)), A: 255}
	case "post-apocalyptic", "postapoc":
		primary = hslToRGBA(30+rng.Float64()*30, 0.15+rng.Float64()*0.25, 0.25+rng.Float64()*0.2)
		accent = hslToRGBA(20+rng.Float64()*20, 0.2+rng.Float64()*0.2, 0.3+rng.Float64()*0.15)
		gem = hslToRGBA(rng.Float64()*60, 0.4+rng.Float64()*0.3, 0.4+rng.Float64()*0.2)
	default:
		primary = hslToRGBA(rng.Float64()*360, 0.3+rng.Float64()*0.4, 0.35+rng.Float64()*0.25)
		accent = hslToRGBA(rng.Float64()*360, 0.4+rng.Float64()*0.4, 0.45+rng.Float64()*0.25)
		gem = hslToRGBA(rng.Float64()*360, 0.7+rng.Float64()*0.3, 0.5+rng.Float64()*0.2)
	}
	return primary, accent, gem
}

func directionAngleOffset(d Direction) float64 {
	switch d {
	case DirUp:
		return -math.Pi / 2
	case DirDown:
		return math.Pi / 2
	case DirLeft:
		return math.Pi
	case DirRight:
		return 0
	default:
		return math.Pi / 2
	}
}

func directionTrailVector(d Direction) (dx, dy int) {
	// Trail extends BEHIND the entity (opposite of facing direction)
	switch d {
	case DirUp:
		return 0, 1
	case DirDown:
		return 0, -1
	case DirLeft:
		return 1, 0
	case DirRight:
		return -1, 0
	default:
		return 0, -1
	}
}

func drawGem(dst *image.RGBA, cx, cy, radius int, gemColor color.RGBA, maxW, maxH int) {
	highlight := lightenHeadgear(gemColor, 60)
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := cx + dx
				py := cy + dy
				if inBoundsHeadgear(px, py, maxW, maxH) {
					t := math.Sqrt(float64(dx*dx+dy*dy)) / float64(maxHeadgear(1, radius))
					c := blendHeadgear(highlight, gemColor, t*0.7)
					setPixelAlpha(dst, px, py, c, 245)
				}
			}
		}
	}
}

func inBoundsHeadgear(x, y, w, h int) bool {
	return x >= 0 && x < w && y >= 0 && y < h
}

func setPixelAlpha(dst *image.RGBA, x, y int, c color.RGBA, alpha uint8) {
	idx := (y-dst.Rect.Min.Y)*dst.Stride + (x-dst.Rect.Min.X)*4
	if idx < 0 || idx+3 >= len(dst.Pix) {
		return
	}
	// Alpha-composite over existing pixel
	srcA := float64(alpha) / 255.0
	dstA := float64(dst.Pix[idx+3]) / 255.0
	outA := srcA + dstA*(1-srcA)
	if outA == 0 {
		return
	}
	dst.Pix[idx+0] = uint8((float64(c.R)*srcA + float64(dst.Pix[idx+0])*dstA*(1-srcA)) / outA)
	dst.Pix[idx+1] = uint8((float64(c.G)*srcA + float64(dst.Pix[idx+1])*dstA*(1-srcA)) / outA)
	dst.Pix[idx+2] = uint8((float64(c.B)*srcA + float64(dst.Pix[idx+2])*dstA*(1-srcA)) / outA)
	dst.Pix[idx+3] = uint8(outA * 255)
}

func maxHeadgear(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func lightenHeadgear(c color.RGBA, amount int) color.RGBA {
	add := func(v uint8, a int) uint8 {
		r := int(v) + a
		if r > 255 {
			return 255
		}
		return uint8(r)
	}
	return color.RGBA{R: add(c.R, amount), G: add(c.G, amount), B: add(c.B, amount), A: c.A}
}

func darkenHeadgear(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

func blendHeadgear(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 255,
	}
}
