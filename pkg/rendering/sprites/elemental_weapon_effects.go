// Package sprites provides elemental weapon effect rendering for top-down sprites.
// This file adds visual special effects to weapons based on elemental enchantments
// (fire, ice, lightning, poison, holy, shadow). Effects are rendered as additional
// overlays with animated particles and color tinting. All rendering is seed-deterministic.
package sprites

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// ElementType represents the elemental damage type of a weapon enchantment.
type ElementType int

const (
	// ElementNone represents no elemental enchantment
	ElementNone ElementType = iota
	// ElementFire adds flames and orange/red glow
	ElementFire
	// ElementIce adds frost crystals and blue/white tint
	ElementIce
	// ElementLightning adds electric arcs and yellow/white sparks
	ElementLightning
	// ElementPoison adds dripping green particles
	ElementPoison
	// ElementHoly adds golden radiance
	ElementHoly
	// ElementShadow adds dark wisps and purple tint
	ElementShadow
)

// String returns the string representation of an element type.
func (e ElementType) String() string {
	switch e {
	case ElementFire:
		return "fire"
	case ElementIce:
		return "ice"
	case ElementLightning:
		return "lightning"
	case ElementPoison:
		return "poison"
	case ElementHoly:
		return "holy"
	case ElementShadow:
		return "shadow"
	default:
		return "none"
	}
}

// ParseElementType converts a string to ElementType.
func ParseElementType(s string) ElementType {
	switch s {
	case "fire", "flame", "burning":
		return ElementFire
	case "ice", "frost", "frozen", "cold":
		return ElementIce
	case "lightning", "electric", "shock", "thunder":
		return ElementLightning
	case "poison", "toxic", "venom", "acid":
		return ElementPoison
	case "holy", "sacred", "divine", "radiant", "light":
		return ElementHoly
	case "shadow", "dark", "void", "unholy", "necrotic":
		return ElementShadow
	default:
		return ElementNone
	}
}

// ElementalEffectParams controls how elemental effects are rendered.
type ElementalEffectParams struct {
	// Element type for the effect
	Element ElementType

	// Intensity controls effect strength (0.0-1.0)
	Intensity float64

	// AnimationPhase controls the current animation frame (0.0-1.0, cycles)
	AnimationPhase float64

	// ParticleCount is the number of particles to render
	ParticleCount int

	// Seed for deterministic randomness
	Seed int64
}

// DefaultElementalParams returns default parameters for an element type.
func DefaultElementalParams(element ElementType, seed int64) ElementalEffectParams {
	return ElementalEffectParams{
		Element:        element,
		Intensity:      0.7,
		AnimationPhase: 0.0,
		ParticleCount:  6,
		Seed:           seed,
	}
}

// ElementalWeaponRenderer renders elemental effects on weapon sprites.
type ElementalWeaponRenderer struct{}

// NewElementalWeaponRenderer creates a new elemental weapon renderer.
func NewElementalWeaponRenderer() *ElementalWeaponRenderer {
	return &ElementalWeaponRenderer{}
}

// ApplyElementalEffect applies elemental visual effects to an existing weapon image.
// The weapon image is modified in-place with element-specific coloring and particles.
func (r *ElementalWeaponRenderer) ApplyElementalEffect(
	img *ebiten.Image,
	params ElementalEffectParams,
) {
	if img == nil || params.Element == ElementNone || params.Intensity <= 0 {
		return
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return
	}

	rng := rand.New(rand.NewSource(params.Seed ^ int64(params.Element)))

	switch params.Element {
	case ElementFire:
		r.applyFireEffect(img, w, h, params, rng)
	case ElementIce:
		r.applyIceEffect(img, w, h, params, rng)
	case ElementLightning:
		r.applyLightningEffect(img, w, h, params, rng)
	case ElementPoison:
		r.applyPoisonEffect(img, w, h, params, rng)
	case ElementHoly:
		r.applyHolyEffect(img, w, h, params, rng)
	case ElementShadow:
		r.applyShadowEffect(img, w, h, params, rng)
	}
}

// applyFireEffect adds flames, orange tint, and ember particles.
func (r *ElementalWeaponRenderer) applyFireEffect(img *ebiten.Image, w, h int, params ElementalEffectParams, rng *rand.Rand) {
	// Color palette for fire
	flameColors := []color.RGBA{
		{R: 255, G: 200, B: 50, A: 255},  // yellow core
		{R: 255, G: 140, B: 30, A: 255},  // orange
		{R: 255, G: 80, B: 20, A: 255},   // red-orange
		{R: 200, G: 40, B: 10, A: 255},   // dark red
		{R: 255, G: 240, B: 200, A: 255}, // white-hot
	}

	// Tint existing weapon pixels with warm orange
	r.tintWeaponPixels(img, w, h, color.RGBA{R: 255, G: 160, B: 80, A: 255}, params.Intensity*0.4)

	// Add flame particles rising from the blade
	for i := 0; i < params.ParticleCount; i++ {
		px := rng.Intn(w)
		baseY := rng.Intn(h/2) + h/4

		// Flames rise upward with flicker
		phase := params.AnimationPhase + float64(i)*0.15
		flickerY := int(math.Sin(phase*math.Pi*2)*3) - 2 // Rise and flicker
		py := baseY + flickerY

		if py >= 0 && py < h && px >= 0 && px < w {
			// Check if near an existing weapon pixel
			if r.isNearOpaquePixel(img, px, py, w, h, 3) {
				flameColor := flameColors[rng.Intn(len(flameColors))]
				// Vary alpha based on intensity and position
				alpha := uint8(180 + int(params.Intensity*75) - rng.Intn(50))
				flameColor.A = alpha
				img.Set(px, py, flameColor)

				// Add glow around flame particle
				r.addGlow(img, px, py, w, h, flameColor, 2, 0.3*params.Intensity)
			}
		}
	}

	// Add edge glow effect
	r.addElementalEdgeGlow(img, w, h, color.RGBA{R: 255, G: 120, B: 40, A: 255}, params.Intensity*0.5)
}

// applyIceEffect adds frost crystals, blue tint, and icicle formations.
func (r *ElementalWeaponRenderer) applyIceEffect(img *ebiten.Image, w, h int, params ElementalEffectParams, rng *rand.Rand) {
	// Color palette for ice
	iceColors := []color.RGBA{
		{R: 200, G: 240, B: 255, A: 255}, // light blue
		{R: 150, G: 220, B: 255, A: 255}, // cyan
		{R: 100, G: 180, B: 240, A: 255}, // blue
		{R: 255, G: 255, B: 255, A: 255}, // white frost
		{R: 180, G: 200, B: 255, A: 255}, // pale blue
	}

	// Tint existing weapon pixels with cold blue
	r.tintWeaponPixels(img, w, h, color.RGBA{R: 150, G: 200, B: 255, A: 255}, params.Intensity*0.35)

	// Add frost crystal particles
	for i := 0; i < params.ParticleCount; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)

		if r.isNearOpaquePixel(img, px, py, w, h, 2) {
			// Crystal pattern: small cross shape
			iceColor := iceColors[rng.Intn(len(iceColors))]
			alpha := uint8(160 + int(params.Intensity*80))
			iceColor.A = alpha

			// Center pixel
			if px >= 0 && px < w && py >= 0 && py < h {
				img.Set(px, py, iceColor)
			}

			// Cross arms for crystal effect (small)
			offsets := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
			for _, off := range offsets {
				nx, ny := px+off[0], py+off[1]
				if nx >= 0 && nx < w && ny >= 0 && ny < h && rng.Float64() < 0.5 {
					dimColor := r.dimColor(iceColor, 0.7)
					img.Set(nx, ny, dimColor)
				}
			}
		}
	}

	// Add specular highlights for icy sheen
	numHighlights := 2 + rng.Intn(3)
	for i := 0; i < numHighlights; i++ {
		hx := rng.Intn(w)
		hy := rng.Intn(h)
		if r.getPixelAlpha(img, hx, hy) > 0 {
			highlight := color.RGBA{R: 255, G: 255, B: 255, A: uint8(200 * params.Intensity)}
			img.Set(hx, hy, highlight)
		}
	}

	// Add cold edge glow
	r.addElementalEdgeGlow(img, w, h, color.RGBA{R: 150, G: 220, B: 255, A: 255}, params.Intensity*0.4)
}

// applyLightningEffect adds electric arcs, sparks, and crackling energy.
func (r *ElementalWeaponRenderer) applyLightningEffect(img *ebiten.Image, w, h int, params ElementalEffectParams, rng *rand.Rand) {
	// Color palette for lightning
	lightningColors := []color.RGBA{
		{R: 255, G: 255, B: 255, A: 255}, // white core
		{R: 255, G: 255, B: 180, A: 255}, // pale yellow
		{R: 200, G: 200, B: 255, A: 255}, // electric blue
		{R: 180, G: 180, B: 255, A: 255}, // blue-white
	}

	// Subtle blue-white tint to weapon
	r.tintWeaponPixels(img, w, h, color.RGBA{R: 220, G: 220, B: 255, A: 255}, params.Intensity*0.25)

	// Draw lightning arcs along the blade
	numArcs := 1 + int(params.Intensity*2)
	for arc := 0; arc < numArcs; arc++ {
		startX := w/4 + rng.Intn(w/2)
		startY := h/4 + rng.Intn(h/2)
		arcLen := 4 + rng.Intn(6)

		// Animate arc position with phase
		phaseOffset := int(math.Sin(params.AnimationPhase*math.Pi*2+float64(arc)) * 2)
		startX += phaseOffset

		x, y := startX, startY
		arcColor := lightningColors[rng.Intn(len(lightningColors))]
		arcColor.A = uint8(200 + int(params.Intensity*55))

		for step := 0; step < arcLen; step++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, arcColor)
			}

			// Zigzag movement
			dx := rng.Intn(3) - 1
			dy := rng.Intn(3) - 1
			x += dx
			y += dy

			// Fade alpha along arc
			arcColor.A = uint8(float64(arcColor.A) * 0.85)
		}
	}

	// Add spark particles
	for i := 0; i < params.ParticleCount; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)

		if r.isNearOpaquePixel(img, px, py, w, h, 4) {
			// Flicker based on animation phase
			if math.Sin(params.AnimationPhase*math.Pi*4+float64(i)*0.5) > 0 {
				sparkColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(220 * params.Intensity)}
				img.Set(px, py, sparkColor)
			}
		}
	}

	// Electric edge glow
	r.addElementalEdgeGlow(img, w, h, color.RGBA{R: 200, G: 200, B: 255, A: 255}, params.Intensity*0.45)
}

// applyPoisonEffect adds dripping green particles and toxic aura.
func (r *ElementalWeaponRenderer) applyPoisonEffect(img *ebiten.Image, w, h int, params ElementalEffectParams, rng *rand.Rand) {
	// Color palette for poison
	poisonColors := []color.RGBA{
		{R: 100, G: 200, B: 50, A: 255},  // bright green
		{R: 80, G: 180, B: 40, A: 255},   // green
		{R: 60, G: 150, B: 30, A: 255},   // dark green
		{R: 150, G: 220, B: 80, A: 255},  // yellow-green
		{R: 120, G: 100, B: 180, A: 255}, // sickly purple
	}

	// Tint weapon with sickly green
	r.tintWeaponPixels(img, w, h, color.RGBA{R: 120, G: 180, B: 80, A: 255}, params.Intensity*0.35)

	// Add dripping particles that fall downward
	for i := 0; i < params.ParticleCount; i++ {
		px := rng.Intn(w)
		baseY := rng.Intn(h / 2)

		// Drips fall downward with animation phase
		phase := params.AnimationPhase + float64(i)*0.2
		dripY := baseY + int(math.Mod(phase*float64(h), float64(h/2)))

		if r.isNearOpaquePixel(img, px, dripY, w, h, 3) {
			dropColor := poisonColors[rng.Intn(len(poisonColors))]
			alpha := uint8(180 + int(params.Intensity*60))
			dropColor.A = alpha

			// Draw drip trail (2-3 pixels)
			for dy := 0; dy < 2+rng.Intn(2); dy++ {
				py := dripY + dy
				if py >= 0 && py < h && px >= 0 && px < w {
					// Fade trail
					trailColor := r.dimColor(dropColor, 1.0-float64(dy)*0.3)
					img.Set(px, py, trailColor)
				}
			}
		}
	}

	// Add toxic bubbles
	numBubbles := 2 + rng.Intn(3)
	for i := 0; i < numBubbles; i++ {
		bx := rng.Intn(w)
		by := rng.Intn(h)
		if r.isNearOpaquePixel(img, bx, by, w, h, 2) {
			bubbleColor := poisonColors[0]
			bubbleColor.A = uint8(100 + rng.Intn(80))
			img.Set(bx, by, bubbleColor)
		}
	}

	// Toxic edge glow
	r.addElementalEdgeGlow(img, w, h, color.RGBA{R: 100, G: 200, B: 60, A: 255}, params.Intensity*0.5)
}

// applyHolyEffect adds golden radiance, light rays, and divine glow.
func (r *ElementalWeaponRenderer) applyHolyEffect(img *ebiten.Image, w, h int, params ElementalEffectParams, rng *rand.Rand) {
	// Color palette for holy
	holyColors := []color.RGBA{
		{R: 255, G: 240, B: 150, A: 255}, // gold
		{R: 255, G: 220, B: 100, A: 255}, // bright gold
		{R: 255, G: 255, B: 200, A: 255}, // pale gold
		{R: 255, G: 255, B: 255, A: 255}, // white
		{R: 255, G: 200, B: 80, A: 255},  // deep gold
	}

	// Golden tint to weapon
	r.tintWeaponPixels(img, w, h, color.RGBA{R: 255, G: 230, B: 160, A: 255}, params.Intensity*0.4)

	// Add radiant light particles
	for i := 0; i < params.ParticleCount; i++ {
		// Particles radiate outward from center
		cx, cy := float64(w)/2, float64(h)/2
		angle := float64(i) * (2 * math.Pi / float64(params.ParticleCount))
		angle += params.AnimationPhase * math.Pi * 2 // Rotate with animation

		radius := float64(w)/4 + float64(rng.Intn(w/4))
		px := int(cx + radius*math.Cos(angle))
		py := int(cy + radius*math.Sin(angle))

		if px >= 0 && px < w && py >= 0 && py < h {
			if r.isNearOpaquePixel(img, px, py, w, h, 4) {
				rayColor := holyColors[rng.Intn(len(holyColors))]
				rayColor.A = uint8(150 + int(params.Intensity*80))
				img.Set(px, py, rayColor)
			}
		}
	}

	// Add central glow/halo
	cx, cy := w/2, h/3
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= 3 {
				px, py := cx+dx, cy+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					alpha := uint8((1 - dist/3) * params.Intensity * 100)
					haloColor := color.RGBA{R: 255, G: 240, B: 180, A: alpha}
					r.blendPixel(img, px, py, haloColor)
				}
			}
		}
	}

	// Strong golden edge glow
	r.addElementalEdgeGlow(img, w, h, color.RGBA{R: 255, G: 220, B: 100, A: 255}, params.Intensity*0.6)
}

// applyShadowEffect adds dark wisps, void particles, and shadowy tendrils.
func (r *ElementalWeaponRenderer) applyShadowEffect(img *ebiten.Image, w, h int, params ElementalEffectParams, rng *rand.Rand) {
	// Color palette for shadow
	shadowColors := []color.RGBA{
		{R: 60, G: 20, B: 80, A: 255},   // dark purple
		{R: 40, G: 10, B: 60, A: 255},   // deep purple
		{R: 80, G: 30, B: 100, A: 255},  // medium purple
		{R: 20, G: 20, B: 30, A: 255},   // near-black
		{R: 100, G: 50, B: 120, A: 255}, // light purple accent
		{R: 150, G: 80, B: 200, A: 255}, // bright purple highlight
	}

	// Dark purple tint to weapon
	r.tintWeaponPixels(img, w, h, color.RGBA{R: 80, G: 40, B: 100, A: 255}, params.Intensity*0.35)

	// Add wispy shadow particles that drift upward
	for i := 0; i < params.ParticleCount; i++ {
		px := rng.Intn(w)
		baseY := h/2 + rng.Intn(h/2)

		// Wisps drift upward slowly
		phase := params.AnimationPhase + float64(i)*0.15
		driftY := baseY - int(math.Mod(phase*float64(h/3), float64(h/2)))
		driftX := px + int(math.Sin(phase*math.Pi*2)*2) // Slight horizontal sway

		if driftY >= 0 && driftY < h && driftX >= 0 && driftX < w {
			if r.isNearOpaquePixel(img, driftX, driftY, w, h, 3) {
				wispColor := shadowColors[rng.Intn(len(shadowColors))]
				alpha := uint8(160 + int(params.Intensity*80) - rng.Intn(40))
				wispColor.A = alpha
				img.Set(driftX, driftY, wispColor)
			}
		}
	}

	// Add void tendrils (small dark lines)
	numTendrils := 1 + int(params.Intensity*2)
	for t := 0; t < numTendrils; t++ {
		tx := rng.Intn(w)
		ty := rng.Intn(h)
		tendrilLen := 3 + rng.Intn(4)

		tendrilColor := shadowColors[3] // Near-black
		tendrilColor.A = uint8(180 * params.Intensity)

		x, y := tx, ty
		for step := 0; step < tendrilLen; step++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, tendrilColor)
			}
			// Tendrils curve upward and outward
			y--
			x += rng.Intn(3) - 1
			tendrilColor.A = uint8(float64(tendrilColor.A) * 0.8)
		}
	}

	// Dark purple edge glow
	r.addElementalEdgeGlow(img, w, h, color.RGBA{R: 100, G: 50, B: 140, A: 255}, params.Intensity*0.5)
}

// --- Helper methods ---

// tintWeaponPixels applies a color tint to all opaque pixels in the image.
func (r *ElementalWeaponRenderer) tintWeaponPixels(img *ebiten.Image, w, h int, tint color.RGBA, intensity float64) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			alpha := r.getPixelAlpha(img, x, y)
			if alpha > 0 {
				existing := r.getPixelRGBA(img, x, y)
				blended := r.blendColors(existing, tint, intensity)
				blended.A = existing.A // Preserve original alpha
				img.Set(x, y, blended)
			}
		}
	}
}

// addElementalEdgeGlow adds a glowing outline effect to weapon edges.
func (r *ElementalWeaponRenderer) addElementalEdgeGlow(img *ebiten.Image, w, h int, glowColor color.RGBA, intensity float64) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if r.getPixelAlpha(img, x, y) > 0 && r.isEdgePixelFast(img, x, y, w, h) {
				// Blend glow color onto edge pixels
				existing := r.getPixelRGBA(img, x, y)
				blended := r.blendColors(existing, glowColor, intensity)
				blended.A = existing.A
				img.Set(x, y, blended)

				// Add outer glow on adjacent transparent pixels
				for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					nx, ny := x+d[0], y+d[1]
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						if r.getPixelAlpha(img, nx, ny) == 0 {
							outerGlow := glowColor
							outerGlow.A = uint8(float64(glowColor.A) * intensity * 0.6)
							img.Set(nx, ny, outerGlow)
						}
					}
				}
			}
		}
	}
}

// addGlow adds a radial glow around a center point.
func (r *ElementalWeaponRenderer) addGlow(img *ebiten.Image, cx, cy, w, h int, glowColor color.RGBA, radius int, intensity float64) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(radius) {
				px, py := cx+dx, cy+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					fade := 1.0 - dist/float64(radius)
					alpha := uint8(float64(glowColor.A) * intensity * fade)
					c := glowColor
					c.A = alpha
					r.blendPixel(img, px, py, c)
				}
			}
		}
	}
}

// isNearOpaquePixel checks if there's an opaque pixel within radius distance.
func (r *ElementalWeaponRenderer) isNearOpaquePixel(img *ebiten.Image, x, y, w, h, radius int) bool {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				if r.getPixelAlpha(img, nx, ny) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// isEdgePixelFast checks if a pixel is on the edge (has transparent neighbor).
func (r *ElementalWeaponRenderer) isEdgePixelFast(img *ebiten.Image, x, y, w, h int) bool {
	offsets := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, off := range offsets {
		nx, ny := x+off[0], y+off[1]
		if nx < 0 || nx >= w || ny < 0 || ny >= h {
			return true
		}
		if r.getPixelAlpha(img, nx, ny) == 0 {
			return true
		}
	}
	return false
}

// getPixelAlpha returns the alpha value of a pixel.
func (r *ElementalWeaponRenderer) getPixelAlpha(img *ebiten.Image, x, y int) uint8 {
	c := img.At(x, y)
	_, _, _, a := c.RGBA()
	return uint8(a >> 8)
}

// getPixelRGBA returns the RGBA value of a pixel.
func (r *ElementalWeaponRenderer) getPixelRGBA(img *ebiten.Image, x, y int) color.RGBA {
	c := img.At(x, y)
	rr, gg, bb, aa := c.RGBA()
	return color.RGBA{R: uint8(rr >> 8), G: uint8(gg >> 8), B: uint8(bb >> 8), A: uint8(aa >> 8)}
}

// blendColors blends two colors by a factor (0.0 = a, 1.0 = b).
func (r *ElementalWeaponRenderer) blendColors(a, b color.RGBA, t float64) color.RGBA {
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
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

// dimColor darkens a color by a factor (0.0 = black, 1.0 = unchanged).
func (r *ElementalWeaponRenderer) dimColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

// blendPixel blends a color onto an existing pixel using alpha compositing.
func (r *ElementalWeaponRenderer) blendPixel(img *ebiten.Image, x, y int, src color.RGBA) {
	if src.A == 0 {
		return
	}
	existing := r.getPixelRGBA(img, x, y)
	if existing.A == 0 {
		img.Set(x, y, src)
		return
	}
	// Simple alpha blend
	srcA := float64(src.A) / 255.0
	result := color.RGBA{
		R: uint8(float64(existing.R)*(1-srcA) + float64(src.R)*srcA),
		G: uint8(float64(existing.G)*(1-srcA) + float64(src.G)*srcA),
		B: uint8(float64(existing.B)*(1-srcA) + float64(src.B)*srcA),
		A: uint8(math.Max(float64(existing.A), float64(src.A))),
	}
	img.Set(x, y, result)
}

// GetElementFromEnchantment extracts element type from an enchantment description or tags.
func GetElementFromEnchantment(enchantType string, tags []string) ElementType {
	// Check enchantment type first
	element := ParseElementType(enchantType)
	if element != ElementNone {
		return element
	}

	// Check tags for elemental keywords
	for _, tag := range tags {
		element = ParseElementType(tag)
		if element != ElementNone {
			return element
		}
	}

	return ElementNone
}
