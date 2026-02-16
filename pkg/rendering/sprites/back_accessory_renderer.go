// Package sprites provides seed-based top-down back accessory rendering for
// humanoid entity sprites. Back accessories (capes, cloaks, quivers, backpacks,
// banners, scarves, and wing-capes) are rendered as overlays behind the torso,
// visible from above draped over shoulders and extending below the body. Each
// accessory type has a distinct silhouette making entities immediately
// distinguishable. All rendering is seed-deterministic.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// BackAccessoryType identifies the kind of back-worn accessory.
type BackAccessoryType int

const (
	// BackAccessoryNone means no back accessory.
	BackAccessoryNone BackAccessoryType = iota
	// BackAccessoryCape is a short shoulder cape ending at mid-back.
	BackAccessoryCape
	// BackAccessoryCloak is a full-length cloak draping past the legs.
	BackAccessoryCloak
	// BackAccessoryQuiver is a diagonal arrow quiver across one shoulder.
	BackAccessoryQuiver
	// BackAccessoryBackpack is a rectangular pack centered on the back.
	BackAccessoryBackpack
	// BackAccessoryBanner is a tall vertical banner or flag on a pole.
	BackAccessoryBanner
	// BackAccessoryScarf is a flowing scarf trailing behind one shoulder.
	BackAccessoryScarf
	// BackAccessoryWingCape is a wide decorative wing-shaped mantle.
	BackAccessoryWingCape

	// backAccessoryCount is the number of accessory types (used for random selection).
	backAccessoryCount
)

// String returns a human-readable name for the accessory type.
func (b BackAccessoryType) String() string {
	switch b {
	case BackAccessoryNone:
		return "none"
	case BackAccessoryCape:
		return "cape"
	case BackAccessoryCloak:
		return "cloak"
	case BackAccessoryQuiver:
		return "quiver"
	case BackAccessoryBackpack:
		return "backpack"
	case BackAccessoryBanner:
		return "banner"
	case BackAccessoryScarf:
		return "scarf"
	case BackAccessoryWingCape:
		return "wing_cape"
	default:
		return "unknown"
	}
}

// BackAccessoryParams controls rendering of a single back accessory overlay.
type BackAccessoryParams struct {
	// SpriteWidth and SpriteHeight are the full sprite dimensions.
	SpriteWidth  int
	SpriteHeight int

	// TorsoSpec is the torso body part layout for positioning.
	TorsoSpec PartSpec

	// AccessoryType selects the shape to render.
	AccessoryType BackAccessoryType

	// PrimaryColor is the main fabric/material color.
	PrimaryColor color.RGBA
	// AccentColor is a trim/detail color.
	AccentColor color.RGBA

	// Direction the entity is facing.
	Direction Direction

	// Seed for deterministic variation.
	Seed int64

	// Genre for style adaptation.
	Genre string
}

// SelectBackAccessoryForRole selects an appropriate back accessory type based
// on entity role, genre, and seed. Returns BackAccessoryNone ~30% of the time.
func SelectBackAccessoryForRole(role, genre string, seed int64) BackAccessoryType {
	rng := rand.New(rand.NewSource(seed ^ 0x4241434B)) // "BACK" XOR salt

	// 30% chance of no accessory
	if rng.Float64() < 0.30 {
		return BackAccessoryNone
	}

	// Role-weighted selection
	weights := roleBackAccessoryWeights(role, genre)
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total == 0 {
		return BackAccessoryNone
	}
	roll := rng.Float64() * total
	cumulative := 0.0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return BackAccessoryType(i + 1) // +1 because index 0 = None
		}
	}
	return BackAccessoryCape
}

// roleBackAccessoryWeights returns probability weights for each accessory type
// indexed 0=Cape, 1=Cloak, 2=Quiver, 3=Backpack, 4=Banner, 5=Scarf, 6=WingCape.
func roleBackAccessoryWeights(role, genre string) [7]float64 {
	base := [7]float64{1.0, 1.0, 0.5, 0.5, 0.3, 0.8, 0.3}

	switch role {
	case "warrior", "knight":
		base = [7]float64{2.0, 0.5, 0.2, 0.3, 1.5, 0.2, 0.3}
	case "mage", "wizard":
		base = [7]float64{0.5, 2.5, 0.1, 0.3, 0.2, 1.0, 1.5}
	case "rogue", "thief":
		base = [7]float64{1.5, 1.5, 0.5, 0.5, 0.1, 1.5, 0.2}
	case "ranger", "archer":
		base = [7]float64{0.8, 1.2, 3.0, 0.8, 0.2, 0.5, 0.3}
	case "merchant", "npc":
		base = [7]float64{0.5, 0.5, 0.1, 2.5, 0.1, 1.0, 0.1}
	case "priest", "healer":
		base = [7]float64{0.5, 2.0, 0.1, 0.3, 0.5, 1.5, 1.0}
	case "player":
		base = [7]float64{1.5, 1.2, 1.0, 0.8, 0.5, 1.0, 0.8}
	}

	// Genre modifiers
	switch genre {
	case "horror":
		base[1] *= 1.5 // More cloaks
		base[6] *= 1.5 // More wing capes
		base[3] *= 0.5 // Fewer backpacks
	case "cyberpunk":
		base[5] *= 1.8 // More scarves
		base[0] *= 0.5 // Fewer traditional capes
		base[3] *= 1.5 // More backpacks/tech packs
	case "sci-fi", "scifi":
		base[3] *= 1.5 // More backpacks/jetpacks
		base[0] *= 1.2 // Capes viable
		base[2] *= 0.3 // Fewer quivers
	case "post-apocalyptic", "postapoc":
		base[3] *= 2.0 // Backpacks essential
		base[5] *= 1.5 // Scarves/wraps
		base[4] *= 0.2 // No banners
	}

	return base
}

// ComputeBackAccessoryParams builds rendering parameters from sprite data and
// component information, using seed-based color generation.
func ComputeBackAccessoryParams(
	spriteW, spriteH int,
	torsoSpec PartSpec,
	accessoryType BackAccessoryType,
	direction Direction,
	seed int64,
	genre string,
) BackAccessoryParams {
	rng := rand.New(rand.NewSource(seed ^ 0x43415045)) // "CAPE" XOR salt

	primary := generateAccessoryColor(rng, genre, accessoryType)
	accent := generateAccentColor(primary, rng)

	return BackAccessoryParams{
		SpriteWidth:   spriteW,
		SpriteHeight:  spriteH,
		TorsoSpec:     torsoSpec,
		AccessoryType: accessoryType,
		PrimaryColor:  primary,
		AccentColor:   accent,
		Direction:     direction,
		Seed:          seed,
		Genre:         genre,
	}
}

// generateAccessoryColor produces a genre-aware base color for the accessory.
func generateAccessoryColor(rng *rand.Rand, genre string, aType BackAccessoryType) color.RGBA {
	var hue, sat, val float64
	switch genre {
	case "horror":
		hue = 260 + rng.Float64()*60 // purples, dark reds
		sat = 0.3 + rng.Float64()*0.3
		val = 0.25 + rng.Float64()*0.25
	case "cyberpunk":
		hue = float64([]int{180, 200, 280, 320}[rng.Intn(4)]) + rng.Float64()*30
		sat = 0.5 + rng.Float64()*0.4
		val = 0.5 + rng.Float64()*0.3
	case "fantasy":
		hue = rng.Float64() * 360
		sat = 0.4 + rng.Float64()*0.4
		val = 0.4 + rng.Float64()*0.4
	default:
		hue = rng.Float64() * 360
		sat = 0.3 + rng.Float64()*0.4
		val = 0.35 + rng.Float64()*0.35
	}

	// Quiver/backpack skew toward browns/greens
	if aType == BackAccessoryQuiver || aType == BackAccessoryBackpack {
		hue = 20 + rng.Float64()*60 // browns/tans/greens
		sat = 0.3 + rng.Float64()*0.3
		val = 0.3 + rng.Float64()*0.3
	}

	return accessoryHSV(hue, sat, val)
}

// generateAccentColor produces a complementary trim color.
func generateAccentColor(primary color.RGBA, rng *rand.Rand) color.RGBA {
	// Lighter or darker variation of primary
	if rng.Float64() < 0.5 {
		return color.RGBA{
			R: clampU8(float64(int(primary.R) + 40 + rng.Intn(30))),
			G: clampU8(float64(int(primary.G) + 40 + rng.Intn(30))),
			B: clampU8(float64(int(primary.B) + 40 + rng.Intn(30))),
			A: 255,
		}
	}
	return color.RGBA{
		R: clampU8(float64(int(primary.R) - 30 - rng.Intn(20))),
		G: clampU8(float64(int(primary.G) - 30 - rng.Intn(20))),
		B: clampU8(float64(int(primary.B) - 30 - rng.Intn(20))),
		A: 255,
	}
}

// RenderBackAccessoryOverlay draws the back accessory onto the given RGBA buffer.
func RenderBackAccessoryOverlay(buf *image.RGBA, params BackAccessoryParams) {
	switch params.AccessoryType {
	case BackAccessoryCape:
		renderCape(buf, params)
	case BackAccessoryCloak:
		renderCloak(buf, params)
	case BackAccessoryQuiver:
		renderQuiver(buf, params)
	case BackAccessoryBackpack:
		renderBackpack(buf, params)
	case BackAccessoryBanner:
		renderBanner(buf, params)
	case BackAccessoryScarf:
		renderScarf(buf, params)
	case BackAccessoryWingCape:
		renderWingCape(buf, params)
	}
}

// --- Cape: short shoulder cape ending at mid-back ---

func renderCape(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0xCA7E))
	w, h := p.SpriteWidth, p.SpriteHeight

	// Cape region: below the head, covering shoulders and upper back
	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	torsoW := int(float64(w) * p.TorsoSpec.RelativeWidth)
	cx := w / 2

	// Cape starts at upper torso and extends to ~60% down
	startY := torsoY - torsoW/4
	endY := startY + int(float64(h)*0.30)
	capeHalfW := int(float64(torsoW) * 0.55)

	for y := startY; y < endY && y < h; y++ {
		if y < 0 {
			continue
		}
		t := float64(y-startY) / float64(endY-startY)
		// Cape widens slightly as it falls, then narrows at bottom
		widthFactor := 0.8 + 0.3*math.Sin(t*math.Pi)
		halfW := int(float64(capeHalfW) * widthFactor)

		// Direction offset — cape trails behind the facing direction
		offsetX := directionOffsetX(p.Direction, t, w)

		for x := cx - halfW + offsetX; x <= cx+halfW+offsetX; x++ {
			if x < 0 || x >= w {
				continue
			}

			// Radial distance from center for shading
			dx := float64(x-(cx+offsetX)) / float64(halfW)
			edgeDark := math.Abs(dx) * 0.25

			// Fabric fold highlights: vertical sinusoidal pattern
			foldX := math.Sin(float64(x)*1.2+float64(rng.Intn(3))) * 0.08
			foldY := math.Cos(float64(y)*0.8) * 0.05

			// Vertical gradient — darker toward bottom
			gradDark := t * 0.20

			shade := 1.0 - edgeDark - gradDark + foldX + foldY
			shade = clampF(shade, 0.4, 1.0)

			c := shadeColor(p.PrimaryColor, shade)

			// Trim border on edges
			if math.Abs(dx) > 0.8 || t > 0.9 {
				c = accessoryBlend(c, p.AccentColor, 0.5)
			}

			accessorySetPixel(buf, x, y, c)
		}
	}
}

// --- Cloak: full-length draping past the torso ---

func renderCloak(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0xC10A))
	w, h := p.SpriteWidth, p.SpriteHeight

	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	torsoW := int(float64(w) * p.TorsoSpec.RelativeWidth)
	cx := w / 2

	// Cloak extends from upper torso almost to the bottom of the sprite
	startY := torsoY - torsoW/3
	endY := int(float64(h) * 0.88)
	cloakHalfW := int(float64(torsoW) * 0.65)

	for y := startY; y < endY && y < h; y++ {
		if y < 0 {
			continue
		}
		t := float64(y-startY) / float64(endY-startY)
		// Cloak gradually widens and then gathers at the bottom
		widthFactor := 0.6 + 0.5*t - 0.15*t*t
		halfW := int(float64(cloakHalfW) * widthFactor)

		offsetX := directionOffsetX(p.Direction, t, w)

		for x := cx - halfW + offsetX; x <= cx+halfW+offsetX; x++ {
			if x < 0 || x >= w {
				continue
			}

			dx := float64(x-(cx+offsetX)) / float64(accessoryMaxI(halfW, 1))
			edgeDark := math.Abs(dx) * 0.30

			// Deep fabric folds for cloak
			foldPhase := float64(x)*0.9 + float64(rng.Intn(5))
			fold := math.Sin(foldPhase) * 0.12

			gradDark := t * 0.30 // Darker at bottom
			shade := 1.0 - edgeDark - gradDark + fold
			shade = clampF(shade, 0.30, 1.0)

			c := shadeColor(p.PrimaryColor, shade)

			// Hood shadow at top (first 15%)
			if t < 0.15 {
				hoodDark := (0.15 - t) / 0.15 * 0.2
				c = shadeColor(c, 1.0-hoodDark)
			}

			// Bottom trim
			if t > 0.92 {
				c = accessoryBlend(c, p.AccentColor, 0.6)
			}

			accessorySetPixel(buf, x, y, c)
		}
	}
}

// --- Quiver: diagonal strap with arrow tips visible ---

func renderQuiver(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0xAF20))
	w, h := p.SpriteWidth, p.SpriteHeight

	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	cx := w / 2

	// Quiver sits on the right side of the back, diagonal
	side := 1 // right side
	if rng.Float64() < 0.5 {
		side = -1 // left side
	}

	quiverW := accessoryMaxI(2, w/8)
	quiverH := int(float64(h) * 0.40)
	qStartY := torsoY - quiverH/3
	qEndY := qStartY + quiverH
	qCX := cx + side*int(float64(w)*0.18)

	for y := qStartY; y < qEndY && y < h; y++ {
		if y < 0 {
			continue
		}
		t := float64(y-qStartY) / float64(qEndY-qStartY)

		// Slight diagonal tilt
		tiltX := int(float64(side) * t * 2)

		for dx := -quiverW; dx <= quiverW; dx++ {
			x := qCX + dx + tiltX
			if x < 0 || x >= w {
				continue
			}

			edgeFade := math.Abs(float64(dx)) / float64(quiverW)
			shade := 1.0 - edgeFade*0.3 - t*0.15
			c := shadeColor(p.PrimaryColor, clampF(shade, 0.5, 1.0))

			// Leather strap color
			if edgeFade > 0.7 {
				c = accessoryBlend(c, p.AccentColor, 0.4)
			}

			accessorySetPixel(buf, x, y, c)
		}
	}

	// Arrow tips poking out above the quiver
	arrowCount := 2 + rng.Intn(3)
	arrowColor := color.RGBA{R: 180, G: 180, B: 170, A: 255} // Metallic tip
	for i := 0; i < arrowCount; i++ {
		ax := qCX - quiverW + rng.Intn(quiverW*2)
		ay := qStartY - 1 - rng.Intn(3)
		if ax >= 0 && ax < w && ay >= 0 && ay < h {
			accessorySetPixel(buf, ax, ay, arrowColor)
		}
		// Small fletching below
		fletchY := ay + 1
		if fletchY >= 0 && fletchY < h && ax >= 0 && ax < w {
			fletchColor := color.RGBA{R: 200, G: 180, B: 140, A: 220}
			accessorySetPixel(buf, ax, fletchY, fletchColor)
		}
	}
}

// --- Backpack: rectangular pack centered on back ---

func renderBackpack(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0xBAC4))
	w, h := p.SpriteWidth, p.SpriteHeight

	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	torsoW := int(float64(w) * p.TorsoSpec.RelativeWidth)
	cx := w / 2

	// Pack is centered on back, slightly lower than center
	packW := int(float64(torsoW) * 0.50)
	packH := int(float64(h) * 0.25)
	packStartY := torsoY
	packEndY := packStartY + packH

	offsetX := directionOffsetX(p.Direction, 0.5, w)

	for y := packStartY; y < packEndY && y < h; y++ {
		if y < 0 {
			continue
		}
		t := float64(y-packStartY) / float64(packEndY-packStartY)

		for dx := -packW / 2; dx <= packW/2; dx++ {
			x := cx + dx + offsetX
			if x < 0 || x >= w {
				continue
			}

			edgeDX := math.Abs(float64(dx)) / float64(accessoryMaxI(packW/2, 1))
			edgeDY := 0.0
			if t < 0.1 || t > 0.9 {
				if t < 0.1 {
					edgeDY = (0.1 - t) / 0.1 * 0.2
				} else {
					edgeDY = (t - 0.9) / 0.1 * 0.2
				}
			}

			// Bulging center: lighter in the middle
			bulge := (1.0 - edgeDX) * (1.0 - math.Abs(t-0.5)*2) * 0.15
			shade := 1.0 - edgeDX*0.25 - edgeDY + bulge
			shade = clampF(shade, 0.45, 1.0)

			c := shadeColor(p.PrimaryColor, shade)

			// Border/frame
			if edgeDX > 0.85 || t < 0.05 || t > 0.95 {
				c = accessoryBlend(c, p.AccentColor, 0.5)
			}

			accessorySetPixel(buf, x, y, c)
		}
	}

	// Straps: two thin vertical lines from top of pack to shoulders
	strapColor := shadeColor(p.AccentColor, 0.8)
	for _, sx := range []int{cx - packW/4 + offsetX, cx + packW/4 + offsetX} {
		for sy := packStartY - int(float64(h)*0.08); sy < packStartY; sy++ {
			if sx >= 0 && sx < w && sy >= 0 && sy < h {
				accessorySetPixel(buf, sx, sy, strapColor)
			}
		}
	}

	// Buckle detail
	buckleY := packStartY + packH/2
	buckleX := cx + offsetX
	buckleC := color.RGBA{R: 180, G: 160, B: 100, A: 255}
	if buckleX >= 0 && buckleX < w && buckleY >= 0 && buckleY < h {
		accessorySetPixel(buf, buckleX, buckleY, buckleC)
	}
	_ = rng // used for seed stability
}

// --- Banner: vertical flag on a pole extending above ---

func renderBanner(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0xF1A9))
	w, h := p.SpriteWidth, p.SpriteHeight

	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	cx := w / 2

	// Banner pole on one side
	side := 1
	if rng.Float64() < 0.5 {
		side = -1
	}

	poleX := cx + side*int(float64(w)*0.22)
	poleStartY := accessoryMaxI(0, torsoY-int(float64(h)*0.30))
	poleEndY := torsoY + int(float64(h)*0.15)
	poleColor := color.RGBA{R: 110, G: 80, B: 50, A: 255}

	// Pole
	for y := poleStartY; y < poleEndY && y < h; y++ {
		if poleX >= 0 && poleX < w {
			accessorySetPixel(buf, poleX, y, poleColor)
		}
	}

	// Flag/banner: rectangle hanging from pole
	flagW := int(float64(w) * 0.18)
	flagH := int(float64(h) * 0.22)
	flagStartY := poleStartY
	flagEndY := flagStartY + flagH

	for y := flagStartY; y < flagEndY && y < h; y++ {
		if y < 0 {
			continue
		}
		t := float64(y-flagStartY) / float64(flagEndY-flagStartY)

		// Slight wave
		wave := int(math.Sin(t*math.Pi*2+float64(rng.Intn(3)))*1.5 + 0.5)

		for dx := 0; dx < flagW; dx++ {
			x := poleX + side*dx + wave
			if x < 0 || x >= w {
				continue
			}

			edgeFade := float64(dx) / float64(flagW)
			shade := 1.0 - edgeFade*0.2 - t*0.15
			c := shadeColor(p.PrimaryColor, clampF(shade, 0.5, 1.0))

			// Emblem stripe across the middle
			if t > 0.35 && t < 0.65 {
				c = accessoryBlend(c, p.AccentColor, 0.5)
			}

			accessorySetPixel(buf, x, y, c)
		}
	}
}

// --- Scarf: flowing fabric trailing from one shoulder ---

func renderScarf(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0x5CAF))
	w, h := p.SpriteWidth, p.SpriteHeight

	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	torsoW := int(float64(w) * p.TorsoSpec.RelativeWidth)
	cx := w / 2

	// Scarf wraps from one shoulder and trails behind
	side := 1
	if rng.Float64() < 0.5 {
		side = -1
	}

	startX := cx + side*int(float64(torsoW)*0.35)
	startY := torsoY - int(float64(h)*0.12)

	// Scarf path: flowing curve from shoulder down and across
	scarfLen := int(float64(h) * 0.35)
	scarfW := accessoryMaxI(2, w/10)

	for i := 0; i < scarfLen; i++ {
		t := float64(i) / float64(scarfLen)

		// Curved path
		curvature := math.Sin(t*math.Pi*1.5) * float64(w) * 0.12
		sx := startX + int(curvature)*(-side)
		sy := startY + i

		if sy < 0 || sy >= h {
			continue
		}

		// Width narrows toward end
		curW := int(float64(scarfW) * (1.0 - t*0.4))

		for dx := -curW; dx <= curW; dx++ {
			x := sx + dx
			if x < 0 || x >= w {
				continue
			}

			edgeFade := math.Abs(float64(dx)) / float64(accessoryMaxI(curW, 1))
			shade := 1.0 - edgeFade*0.3 - t*0.2
			c := shadeColor(p.PrimaryColor, clampF(shade, 0.45, 1.0))

			// Fringe detail at the end
			if t > 0.85 {
				c = accessoryBlend(c, p.AccentColor, (t-0.85)/0.15*0.6)
			}

			accessorySetPixel(buf, x, sy, c)
		}
	}

	// Wrap around neck/shoulder: small arc at the starting point
	for dx := -scarfW - 1; dx <= scarfW+1; dx++ {
		x := startX + dx
		y := startY
		if x >= 0 && x < w && y >= 0 && y < h {
			shade := 1.0 - math.Abs(float64(dx))/float64(scarfW+1)*0.2
			accessorySetPixel(buf, x, y, shadeColor(p.PrimaryColor, shade))
		}
	}
	_ = rng
}

// --- Wing Cape: wide decorative mantle spread like wings ---

func renderWingCape(buf *image.RGBA, p BackAccessoryParams) {
	rng := rand.New(rand.NewSource(p.Seed ^ 0xA1C0))
	w, h := p.SpriteWidth, p.SpriteHeight

	torsoY := int(float64(h) * p.TorsoSpec.RelativeY)
	torsoW := int(float64(w) * p.TorsoSpec.RelativeWidth)
	cx := w / 2

	// Wing cape spreads wide from shoulders
	startY := torsoY - int(float64(h)*0.05)
	endY := startY + int(float64(h)*0.35)
	maxHalfW := int(float64(torsoW)*0.70) + 2

	for y := startY; y < endY && y < h; y++ {
		if y < 0 {
			continue
		}
		t := float64(y-startY) / float64(endY-startY)

		// Wing shape: widens then narrows, with scalloped edges
		widthCurve := math.Sin(t * math.Pi)
		halfW := int(float64(maxHalfW) * widthCurve)

		offsetX := directionOffsetX(p.Direction, t, w)

		for dx := -halfW; dx <= halfW; dx++ {
			x := cx + dx + offsetX
			if x < 0 || x >= w {
				continue
			}

			normDX := math.Abs(float64(dx)) / float64(accessoryMaxI(halfW, 1))

			// Scalloped outer edge
			scallop := math.Sin(normDX*math.Pi*3+float64(rng.Intn(3))) * 0.06
			edgeDark := normDX * 0.30

			// Feather/rib lines radiating from center
			ribAngle := math.Atan2(float64(y-startY), float64(dx))
			rib := math.Abs(math.Sin(ribAngle*4)) * 0.08

			shade := 1.0 - edgeDark - t*0.15 + scallop + rib
			shade = clampF(shade, 0.35, 1.0)

			c := shadeColor(p.PrimaryColor, shade)

			// Outer trim
			if normDX > 0.85 {
				c = accessoryBlend(c, p.AccentColor, 0.5)
			}

			accessorySetPixel(buf, x, y, c)
		}
	}
}

// --- Utility functions ---

// directionOffsetX shifts accessory rendering opposite to facing direction
// so the cape/cloak appears to trail behind the entity.
func directionOffsetX(dir Direction, t float64, spriteW int) int {
	trail := int(math.Ceil(t * float64(spriteW) * 0.05))
	switch dir {
	case DirLeft:
		return trail
	case DirRight:
		return -trail
	default:
		return 0
	}
}

// accessoryBlend linearly interpolates between two colors by factor t (0=a, 1=b).
func accessoryBlend(a, b color.RGBA, t float64) color.RGBA {
	t = clampF(t, 0, 1)
	return color.RGBA{
		R: clampU8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: clampU8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: clampU8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: clampU8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

// accessorySetPixel sets a pixel in an RGBA image, blending with existing content.
func accessorySetPixel(buf *image.RGBA, x, y int, c color.RGBA) {
	if x < buf.Rect.Min.X || x >= buf.Rect.Max.X || y < buf.Rect.Min.Y || y >= buf.Rect.Max.Y {
		return
	}
	if c.A == 0 {
		return
	}
	idx := buf.PixOffset(x, y)
	if c.A == 255 {
		buf.Pix[idx+0] = c.R
		buf.Pix[idx+1] = c.G
		buf.Pix[idx+2] = c.B
		buf.Pix[idx+3] = c.A
		return
	}
	alpha := float64(c.A) / 255.0
	invA := 1.0 - alpha
	buf.Pix[idx+0] = clampU8(float64(c.R)*alpha + float64(buf.Pix[idx+0])*invA)
	buf.Pix[idx+1] = clampU8(float64(c.G)*alpha + float64(buf.Pix[idx+1])*invA)
	buf.Pix[idx+2] = clampU8(float64(c.B)*alpha + float64(buf.Pix[idx+2])*invA)
	buf.Pix[idx+3] = clampU8(float64(buf.Pix[idx+3])*invA + float64(c.A)*alpha)
}

// accessoryHSV converts HSV (hue 0–360, sat 0–1, val 0–1) to RGBA.
func accessoryHSV(h, s, v float64) color.RGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: clampU8((r + m) * 255),
		G: clampU8((g + m) * 255),
		B: clampU8((b + m) * 255),
		A: 255,
	}
}

// accessoryMaxI returns the larger of two ints.
func accessoryMaxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
