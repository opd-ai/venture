// Package sprites provides top-down equipment overlay rendering.
// This file renders recognizable weapon, armor, shield, and accessory silhouettes
// as seen from directly above, with material-specific textures, damage wear, and
// enchantment glow effects. All rendering is seed-deterministic.
package sprites

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// EquipmentRenderer produces pixel-level top-down equipment overlays.
type EquipmentRenderer struct{}

// NewEquipmentRenderer creates a new equipment renderer.
func NewEquipmentRenderer() *EquipmentRenderer {
	return &EquipmentRenderer{}
}

// RenderEquipment draws a complete equipment overlay image for the given slot,
// applying material texture, damage wear, and enchantment glow.
func (r *EquipmentRenderer) RenderEquipment(
	width, height int,
	equip EquipmentVisual,
	rng *rand.Rand,
) *ebiten.Image {
	if width <= 0 || height <= 0 {
		return ebiten.NewImage(1, 1)
	}

	img := ebiten.NewImage(width, height)

	// Derive a stable weapon sub-type from seed so same item always looks the same
	subRng := rand.New(rand.NewSource(equip.Seed))

	materialProps := GetMaterialVisualProperties(equip.Material)
	damageEffects := GetDamageVisualEffects(equip.DamageState)

	// Base color from material with variation from seed
	baseColor := r.materialBaseColor(equip.Material, subRng)

	switch equip.Slot {
	case "weapon":
		r.renderWeapon(img, width, height, baseColor, materialProps, subRng)
	case "armor":
		r.renderArmor(img, width, height, baseColor, materialProps, subRng)
	case "helmet":
		r.renderHelmet(img, width, height, baseColor, materialProps, subRng)
	case "shield":
		r.renderShield(img, width, height, baseColor, materialProps, subRng)
	case "accessory":
		r.renderAccessory(img, width, height, baseColor, materialProps, subRng)
	default:
		r.renderWeapon(img, width, height, baseColor, materialProps, subRng)
	}

	// Apply material texture overlay
	r.applyMaterialTexture(img, width, height, equip.Material, materialProps, subRng)

	// Apply damage wear
	if damageEffects.CrackDensity > 0 || damageEffects.Dirtiness > 0 {
		r.applyDamageWear(img, width, height, damageEffects, subRng)
	}

	// Apply enchantment glow
	if equip.Enchantment.Active && equip.Enchantment.Intensity > 0 {
		r.applyEnchantGlow(img, width, height, equip.Enchantment, subRng)
	}

	return img
}

// --- Weapon silhouettes (top-down view) ---

func (r *EquipmentRenderer) renderWeapon(img *ebiten.Image, w, h int, baseColor color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	weaponType := rng.Intn(6) // sword, axe, spear, bow, staff, dagger
	switch weaponType {
	case 0:
		r.drawSword(img, w, h, baseColor, mat, rng)
	case 1:
		r.drawAxe(img, w, h, baseColor, mat, rng)
	case 2:
		r.drawSpear(img, w, h, baseColor, mat, rng)
	case 3:
		r.drawBow(img, w, h, baseColor, mat, rng)
	case 4:
		r.drawStaff(img, w, h, baseColor, mat, rng)
	case 5:
		r.drawDagger(img, w, h, baseColor, mat, rng)
	}
}

// drawSword: elongated blade extending diagonally from center, visible from above.
func (r *EquipmentRenderer) drawSword(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	cx := float64(w) / 2
	cy := float64(h) / 2
	bladeLen := float64(h) * 0.8
	bladeW := math.Max(2, float64(w)*0.12)

	angle := -math.Pi/4 + rng.Float64()*0.2
	cosA, sinA := math.Cos(angle), math.Sin(angle)

	// Blade highlight color
	highlight := r.lighten(base, 40+int(mat.Sheen*30))
	shadow := r.darken(base, 30)

	for i := 0; i < int(bladeLen); i++ {
		t := float64(i) / bladeLen
		// Taper blade toward tip
		curW := bladeW * (1.0 - t*0.6)
		px := cx + float64(i)*cosA
		py := cy + float64(i)*sinA

		for dw := -curW / 2; dw <= curW/2; dw++ {
			nx := int(px + dw*(-sinA))
			ny := int(py + dw*cosA)
			if nx < 0 || nx >= w || ny < 0 || ny >= h {
				continue
			}
			// Highlight on one edge, shadow on the other
			if dw < 0 {
				r.setPixel(img, nx, ny, r.blendColor(base, highlight, 0.3+t*0.3))
			} else {
				r.setPixel(img, nx, ny, r.blendColor(base, shadow, 0.2+t*0.2))
			}
		}
	}

	// Guard (cross-piece near the handle)
	guardLen := float64(w) * 0.3
	guardY := cy + 3*cosA
	guardX := cx + 3*sinA
	for g := -guardLen / 2; g <= guardLen/2; g++ {
		gx := int(guardX + g*(-sinA))
		gy := int(guardY + g*cosA)
		if gx >= 0 && gx < w && gy >= 0 && gy < h {
			r.setPixel(img, gx, gy, r.darken(base, 20))
		}
	}

	// Pommel
	r.drawFilledCircle(img, int(cx-2*cosA), int(cy-2*sinA), 2, r.darken(base, 40))
}

// drawAxe: T-shaped from above — short handle with wide head.
func (r *EquipmentRenderer) drawAxe(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	cx := w / 2
	handleLen := int(float64(h) * 0.6)
	handleW := maxInt(1, w/8)

	// Handle
	handleColor := color.RGBA{R: 120, G: 80, B: 40, A: 255} // Wood handle
	for y := cx - handleLen/2; y <= cx+handleLen/2; y++ {
		for x := h/2 - handleW; x <= h/2+handleW; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				r.setPixel(img, x, y, handleColor)
			}
		}
	}

	// Axe head: wide crescent on one side
	headW := int(float64(w) * 0.45)
	headH := int(float64(h) * 0.3)
	headY := cx - handleLen/2
	headX := w / 2
	highlight := r.lighten(base, 30+int(mat.Sheen*25))

	for dy := -headH / 2; dy <= headH/2; dy++ {
		// Crescent shape: wider in the middle
		t := 1.0 - math.Abs(float64(dy))/float64(headH/2)
		curW := int(float64(headW) * t)
		for dx := 0; dx < curW; dx++ {
			px := headX + dx
			py := headY + dy
			if px >= 0 && px < w && py >= 0 && py < h {
				edgeFade := float64(dx) / float64(curW)
				c := r.blendColor(base, highlight, edgeFade*0.5)
				r.setPixel(img, px, py, c)
			}
		}
	}
}

// drawSpear: thin long line with a triangular tip.
func (r *EquipmentRenderer) drawSpear(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, _ *rand.Rand) {
	cx := w / 2
	shaftColor := color.RGBA{R: 110, G: 75, B: 35, A: 255}
	tipLen := int(float64(h) * 0.2)

	// Shaft
	for y := tipLen; y < h; y++ {
		if cx >= 0 && cx < w && y >= 0 && y < h {
			r.setPixel(img, cx, y, shaftColor)
			if cx+1 < w {
				r.setPixel(img, cx+1, y, r.darken(shaftColor, 15))
			}
		}
	}

	// Spear tip (triangle)
	highlight := r.lighten(base, int(mat.Sheen*40))
	for y := 0; y < tipLen; y++ {
		t := float64(y) / float64(tipLen)
		halfW := int(t * float64(w) * 0.15)
		for dx := -halfW; dx <= halfW; dx++ {
			px := cx + dx
			if px >= 0 && px < w && y >= 0 && y < h {
				c := r.blendColor(base, highlight, 0.5-math.Abs(float64(dx))/float64(maxInt(1, halfW))*0.3)
				r.setPixel(img, px, y, c)
			}
		}
	}
}

// drawBow: curved arc visible from above.
func (r *EquipmentRenderer) drawBow(img *ebiten.Image, w, h int, base color.RGBA, _ MaterialVisualProperties, _ *rand.Rand) {
	bowColor := color.RGBA{R: 130, G: 90, B: 45, A: 255}
	stringColor := color.RGBA{R: 200, G: 190, B: 170, A: 200}
	cx := float64(w) / 2
	cy := float64(h) / 2
	radius := float64(h) * 0.4

	// Draw bow arc (right half of circle)
	for angle := -math.Pi * 0.4; angle <= math.Pi*0.4; angle += 0.05 {
		px := int(cx + radius*math.Cos(angle))
		py := int(cy + radius*math.Sin(angle))
		if px >= 0 && px < w && py >= 0 && py < h {
			r.setPixel(img, px, py, bowColor)
			// Thickness
			if px+1 < w {
				r.setPixel(img, px+1, py, r.darken(bowColor, 10))
			}
		}
	}

	// Bowstring (straight line)
	topY := int(cy - radius*math.Sin(math.Pi*0.4))
	botY := int(cy + radius*math.Sin(math.Pi*0.4))
	stringX := int(cx + radius*math.Cos(math.Pi*0.4))
	for y := topY; y <= botY; y++ {
		if stringX >= 0 && stringX < w && y >= 0 && y < h {
			r.setPixel(img, stringX, y, stringColor)
		}
	}
	_ = base // base used by callers for consistency
}

// drawStaff: thin vertical line with a glowing orb at the top.
func (r *EquipmentRenderer) drawStaff(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, _ *rand.Rand) {
	cx := w / 2
	staffColor := color.RGBA{R: 100, G: 70, B: 35, A: 255}
	orbRadius := maxInt(2, w/6)
	orbColor := base

	// Staff shaft
	for y := orbRadius * 2; y < h; y++ {
		if cx >= 0 && cx < w && y >= 0 && y < h {
			r.setPixel(img, cx, y, staffColor)
		}
	}

	// Orb at top with radial glow
	orbCy := orbRadius + 1
	highlight := r.lighten(orbColor, 60+int(mat.Sheen*30))
	for dy := -orbRadius; dy <= orbRadius; dy++ {
		for dx := -orbRadius; dx <= orbRadius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(orbRadius) {
				px := cx + dx
				py := orbCy + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					t := dist / float64(orbRadius)
					c := r.blendColor(highlight, orbColor, t*0.7)
					r.setPixel(img, px, py, c)
				}
			}
		}
	}
}

// drawDagger: short blade, compact.
func (r *EquipmentRenderer) drawDagger(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, _ *rand.Rand) {
	cx := float64(w) / 2
	cy := float64(h) / 2
	bladeLen := float64(h) * 0.45
	bladeW := math.Max(1.5, float64(w)*0.1)
	highlight := r.lighten(base, 35+int(mat.Sheen*25))

	// Blade
	for i := 0; i < int(bladeLen); i++ {
		t := float64(i) / bladeLen
		curW := bladeW * (1.0 - t*0.7)
		py := int(cy - bladeLen/2 + float64(i))
		for dw := -curW / 2; dw <= curW/2; dw++ {
			px := int(cx + dw)
			if px >= 0 && px < w && py >= 0 && py < h {
				edgeT := math.Abs(dw) / (curW / 2)
				c := r.blendColor(base, highlight, (1.0-edgeT)*0.4)
				r.setPixel(img, px, py, c)
			}
		}
	}

	// Handle
	handleColor := color.RGBA{R: 100, G: 65, B: 30, A: 255}
	handleStart := int(cy + bladeLen/2)
	handleEnd := minInt(h-1, handleStart+int(float64(h)*0.15))
	for y := handleStart; y <= handleEnd; y++ {
		for dx := -1; dx <= 1; dx++ {
			px := int(cx) + dx
			if px >= 0 && px < w && y >= 0 && y < h {
				r.setPixel(img, px, y, handleColor)
			}
		}
	}
}

// --- Armor overlay (top-down view: shoulder pads / body coverage) ---

func (r *EquipmentRenderer) renderArmor(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	cx := float64(w) / 2
	cy := float64(h) * 0.45 // Shoulders area from above
	highlight := r.lighten(base, 25+int(mat.Sheen*30))
	shadow := r.darken(base, 25)

	shoulderW := float64(w) * 0.42
	shoulderH := float64(h) * 0.18
	padThickness := float64(h) * 0.1

	// Left shoulder pad
	r.drawEllipseFilled(img, int(cx-shoulderW/2), int(cy), int(shoulderW*0.4), int(shoulderH), base, highlight, shadow, w, h)
	// Right shoulder pad
	r.drawEllipseFilled(img, int(cx+shoulderW/2), int(cy), int(shoulderW*0.4), int(shoulderH), base, highlight, shadow, w, h)

	// Chest plate connecting shoulders (visible between shoulder pads from above)
	chestTop := int(cy + shoulderH*0.3)
	chestBot := int(cy + shoulderH*0.3 + padThickness)
	for y := chestTop; y <= chestBot; y++ {
		halfW := int(shoulderW * 0.35)
		for dx := -halfW; dx <= halfW; dx++ {
			px := int(cx) + dx
			if px >= 0 && px < w && y >= 0 && y < h {
				edgeT := math.Abs(float64(dx)) / float64(maxInt(1, halfW))
				c := r.blendColor(base, shadow, edgeT*0.3)
				r.setPixel(img, px, y, c)
			}
		}
	}

	// Collar / neckguard (ring around where head meets torso)
	collarRadius := float64(w) * 0.12
	collarY := int(cy - shoulderH*0.2)
	for angle := 0.0; angle < 2*math.Pi; angle += 0.15 {
		px := int(cx + collarRadius*math.Cos(angle))
		py := collarY + int(collarRadius*0.6*math.Sin(angle))
		if px >= 0 && px < w && py >= 0 && py < h {
			r.setPixel(img, px, py, r.darken(base, 15))
		}
	}

	_ = rng
}

// --- Helmet overlay (top-down: visible as head decoration from above) ---

func (r *EquipmentRenderer) renderHelmet(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	cx := w / 2
	cy := int(float64(h) * 0.3)
	radius := int(float64(w) * 0.22)
	highlight := r.lighten(base, 30+int(mat.Sheen*35))

	// Helmet dome (circle from above)
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(radius) {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					t := dist / float64(radius)
					// Radial shading: bright center, dark edge
					c := r.blendColor(highlight, base, t*0.6)
					r.setPixel(img, px, py, c)
				}
			}
		}
	}

	// Crest / ridge down the center (visible line on top)
	crestColor := r.lighten(base, 50)
	for dy := -radius; dy <= radius/2; dy++ {
		py := cy + dy
		if cx >= 0 && cx < w && py >= 0 && py < h {
			r.setPixel(img, cx, py, crestColor)
		}
	}

	// Visor slit (darker horizontal line)
	visorY := cy + radius/3
	visorHalfW := radius * 2 / 3
	for dx := -visorHalfW; dx <= visorHalfW; dx++ {
		px := cx + dx
		if px >= 0 && px < w && visorY >= 0 && visorY < h {
			r.setPixel(img, px, visorY, r.darken(base, 60))
		}
	}

	_ = rng
}

// --- Shield overlay (top-down: crescent or circle on off-hand side) ---

func (r *EquipmentRenderer) renderShield(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	shieldType := rng.Intn(3)
	highlight := r.lighten(base, 25+int(mat.Sheen*30))
	shadow := r.darken(base, 30)

	// Shield positioned on the left side from above
	cx := int(float64(w) * 0.35)
	cy := h / 2

	switch shieldType {
	case 0: // Round shield
		radius := int(float64(w) * 0.25)
		r.drawRadialShield(img, cx, cy, radius, base, highlight, shadow, w, h)
	case 1: // Kite shield (teardrop from above)
		r.drawKiteShield(img, cx, cy, w, h, base, highlight, shadow)
	case 2: // Tower shield (rectangle)
		shieldW := int(float64(w) * 0.2)
		shieldH := int(float64(h) * 0.45)
		r.drawRectShield(img, cx, cy, shieldW, shieldH, base, highlight, shadow, w, h)
	}
}

func (r *EquipmentRenderer) drawRadialShield(img *ebiten.Image, cx, cy, radius int, base, highlight, shadow color.RGBA, w, h int) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(radius) {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					t := dist / float64(radius)
					c := r.blendColor(highlight, base, t*0.5)
					// Edge rim
					if t > 0.8 {
						c = r.blendColor(c, shadow, (t-0.8)*5*0.4)
					}
					r.setPixel(img, px, py, c)
				}
			}
		}
	}
	// Boss (center rivet)
	r.drawFilledCircle(img, cx, cy, maxInt(1, radius/4), r.lighten(base, 50))
}

func (r *EquipmentRenderer) drawKiteShield(img *ebiten.Image, cx, cy, w, h int, base, highlight, shadow color.RGBA) {
	shieldH := int(float64(h) * 0.45)
	shieldW := int(float64(w) * 0.22)
	topY := cy - shieldH/3
	botY := cy + shieldH*2/3

	for y := topY; y <= botY; y++ {
		t := float64(y-topY) / float64(maxInt(1, botY-topY))
		halfW := float64(shieldW) * (1.0 - t*0.7) // Tapers down
		for dx := -int(halfW); dx <= int(halfW); dx++ {
			px := cx + dx
			if px >= 0 && px < w && y >= 0 && y < h {
				edgeT := math.Abs(float64(dx)) / math.Max(1, halfW)
				c := r.blendColor(highlight, base, edgeT*0.4+t*0.2)
				if edgeT > 0.8 {
					c = r.blendColor(c, shadow, 0.3)
				}
				r.setPixel(img, px, y, c)
			}
		}
	}
}

func (r *EquipmentRenderer) drawRectShield(img *ebiten.Image, cx, cy, sw, sh int, base, highlight, shadow color.RGBA, w, h int) {
	for dy := -sh / 2; dy <= sh/2; dy++ {
		for dx := -sw / 2; dx <= sw/2; dx++ {
			px := cx + dx
			py := cy + dy
			if px >= 0 && px < w && py >= 0 && py < h {
				edgeX := math.Abs(float64(dx)) / float64(maxInt(1, sw/2))
				edgeY := math.Abs(float64(dy)) / float64(maxInt(1, sh/2))
				edge := math.Max(edgeX, edgeY)
				c := r.blendColor(highlight, base, edge*0.4)
				if edge > 0.85 {
					c = r.blendColor(c, shadow, 0.3)
				}
				r.setPixel(img, px, py, c)
			}
		}
	}
}

// --- Accessory overlay (top-down: ring, amulet, cape clasp) ---

func (r *EquipmentRenderer) renderAccessory(img *ebiten.Image, w, h int, base color.RGBA, mat MaterialVisualProperties, rng *rand.Rand) {
	accType := rng.Intn(3)
	highlight := r.lighten(base, 40+int(mat.Sheen*30))

	cx := w / 2
	cy := h / 2

	switch accType {
	case 0: // Amulet / pendant (small circle with chain)
		gemRadius := maxInt(2, w/8)
		// Chain lines
		chainColor := r.darken(base, 20)
		for y := 0; y < cy-gemRadius; y++ {
			if cx >= 0 && cx < w && y >= 0 && y < h {
				r.setPixel(img, cx, y, chainColor)
			}
		}
		// Gem
		for dy := -gemRadius; dy <= gemRadius; dy++ {
			for dx := -gemRadius; dx <= gemRadius; dx++ {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= float64(gemRadius) {
					px := cx + dx
					py := cy + dy
					if px >= 0 && px < w && py >= 0 && py < h {
						t := dist / float64(gemRadius)
						c := r.blendColor(highlight, base, t*0.6)
						r.setPixel(img, px, py, c)
					}
				}
			}
		}

	case 1: // Cape / cloak (draping behind, visible from above as arc behind body)
		capeW := int(float64(w) * 0.6)
		capeH := int(float64(h) * 0.35)
		capeY := int(float64(h) * 0.6)
		capeColor := base

		for dy := 0; dy < capeH; dy++ {
			t := float64(dy) / float64(maxInt(1, capeH))
			halfW := int(float64(capeW) * (0.5 + t*0.5))
			for dx := -halfW; dx <= halfW; dx++ {
				px := cx + dx
				py := capeY + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					edgeT := math.Abs(float64(dx)) / float64(maxInt(1, halfW))
					fade := t*0.3 + edgeT*0.2
					c := r.blendColor(capeColor, r.darken(capeColor, 30), fade)
					r.setPixel(img, px, py, c)
				}
			}
		}

	case 2: // Ring glow (subtle radial aura around character)
		ringRadius := int(float64(w) * 0.35)
		ringThickness := maxInt(1, w/10)
		for angle := 0.0; angle < 2*math.Pi; angle += 0.08 {
			for t := 0; t < ringThickness; t++ {
				rad := float64(ringRadius) + float64(t)
				px := int(float64(cx) + rad*math.Cos(angle))
				py := int(float64(cy) + rad*math.Sin(angle))
				if px >= 0 && px < w && py >= 0 && py < h {
					fade := float64(t) / float64(maxInt(1, ringThickness))
					c := r.withAlpha(highlight, uint8(180-int(fade*100)))
					r.setPixel(img, px, py, c)
				}
			}
		}
	}
}

// --- Material textures ---

func (r *EquipmentRenderer) applyMaterialTexture(img *ebiten.Image, w, h int, material MaterialType, props MaterialVisualProperties, rng *rand.Rand) {
	switch material {
	case MaterialMetal:
		r.applyMetalTexture(img, w, h, props, rng)
	case MaterialLeather:
		r.applyLeatherTexture(img, w, h, props, rng)
	case MaterialCloth:
		r.applyClothTexture(img, w, h)
	case MaterialWood:
		r.applyWoodTexture(img, w, h, rng)
	case MaterialCrystal:
		r.applyCrystalTexture(img, w, h, rng)
	case MaterialEnergy:
		r.applyEnergyTexture(img, w, h, rng)
	}
}

// applyMetalTexture adds specular highlight lines (horizontal streaks) for metal materials.
func (r *EquipmentRenderer) applyMetalTexture(img *ebiten.Image, w, h int, props MaterialVisualProperties, rng *rand.Rand) {
	numStreaks := 2 + rng.Intn(3)
	for i := 0; i < numStreaks; i++ {
		y := rng.Intn(h)
		length := w/4 + rng.Intn(w/3)
		startX := rng.Intn(w)
		for x := startX; x < startX+length && x < w; x++ {
			if y >= 0 && y < h && x >= 0 && x < w {
				existing := r.getPixel(img, x, y)
				if existing.A > 0 {
					bright := r.lighten(existing, 15+int(props.Sheen*15))
					r.setPixel(img, x, y, bright)
				}
			}
		}
	}
}

// applyLeatherTexture adds grain texture (scattered darker dots) for leather materials.
func (r *EquipmentRenderer) applyLeatherTexture(img *ebiten.Image, w, h int, props MaterialVisualProperties, rng *rand.Rand) {
	numDots := int(float64(w*h) * props.Roughness * 0.08)
	for i := 0; i < numDots; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		existing := r.getPixel(img, px, py)
		if existing.A > 0 {
			r.setPixel(img, px, py, r.darken(existing, 8+rng.Intn(12)))
		}
	}
}

// applyClothTexture adds weave pattern (alternating rows slightly lighter/darker) for cloth materials.
func (r *EquipmentRenderer) applyClothTexture(img *ebiten.Image, w, h int) {
	for y := 0; y < h; y += 2 {
		for x := 0; x < w; x += 2 {
			existing := r.getPixel(img, x, y)
			if existing.A > 0 {
				if (x/2+y/2)%2 == 0 {
					r.setPixel(img, x, y, r.lighten(existing, 5))
				} else {
					r.setPixel(img, x, y, r.darken(existing, 5))
				}
			}
		}
	}
}

// applyWoodTexture adds wood grain (horizontal lines with slight variation) for wood materials.
func (r *EquipmentRenderer) applyWoodTexture(img *ebiten.Image, w, h int, rng *rand.Rand) {
	for y := 0; y < h; y++ {
		if y%3 == 0 {
			offset := rng.Intn(3) - 1
			for x := 0; x < w; x++ {
				py := y + offset
				if py >= 0 && py < h {
					existing := r.getPixel(img, x, py)
					if existing.A > 0 {
						r.setPixel(img, x, py, r.darken(existing, 6))
					}
				}
			}
		}
	}
}

// applyCrystalTexture adds faceted highlights (bright spots at random positions) for crystal materials.
func (r *EquipmentRenderer) applyCrystalTexture(img *ebiten.Image, w, h int, rng *rand.Rand) {
	numFacets := 3 + rng.Intn(4)
	for i := 0; i < numFacets; i++ {
		fx := rng.Intn(w)
		fy := rng.Intn(h)
		existing := r.getPixel(img, fx, fy)
		if existing.A > 0 {
			bright := r.lighten(existing, 50)
			r.setPixel(img, fx, fy, bright)
			// Adjacent pixels slightly brighter
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				nx, ny := fx+d[0], fy+d[1]
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					adj := r.getPixel(img, nx, ny)
					if adj.A > 0 {
						r.setPixel(img, nx, ny, r.lighten(adj, 25))
					}
				}
			}
		}
	}
}

// applyEnergyTexture adds pulsing glow dots scattered across surface for energy materials.
func (r *EquipmentRenderer) applyEnergyTexture(img *ebiten.Image, w, h int, rng *rand.Rand) {
	numGlows := 4 + rng.Intn(5)
	for i := 0; i < numGlows; i++ {
		gx := rng.Intn(w)
		gy := rng.Intn(h)
		existing := r.getPixel(img, gx, gy)
		if existing.A > 0 {
			glowC := r.lighten(existing, 70)
			r.setPixel(img, gx, gy, glowC)
		}
	}
}

// --- Damage wear ---

func (r *EquipmentRenderer) applyDamageWear(img *ebiten.Image, w, h int, effects DamageVisualEffects, rng *rand.Rand) {
	// Cracks: dark lines across equipment pixels
	if effects.CrackDensity > 0 {
		numCracks := maxInt(1, int(effects.CrackDensity*8))
		crackColor := color.RGBA{R: 35, G: 30, B: 25, A: 220}
		for i := 0; i < numCracks; i++ {
			x := rng.Intn(w)
			y := rng.Intn(h)
			length := 2 + rng.Intn(maxInt(1, int(effects.CrackDensity*5)))
			dx := rng.Intn(3) - 1
			dy := 1
			if rng.Intn(2) == 0 {
				dy = 0
				dx = 1
			}
			for j := 0; j < length; j++ {
				px := x + j*dx
				py := y + j*dy
				if px >= 0 && px < w && py >= 0 && py < h {
					existing := r.getPixel(img, px, py)
					if existing.A > 0 {
						r.setPixel(img, px, py, crackColor)
					}
				}
			}
		}
	}

	// Dirt spots
	if effects.Dirtiness > 0 {
		numDirt := maxInt(1, int(effects.Dirtiness*6))
		dirtColor := color.RGBA{R: 55, G: 45, B: 35, A: 150}
		for i := 0; i < numDirt; i++ {
			px := rng.Intn(w)
			py := rng.Intn(h)
			existing := r.getPixel(img, px, py)
			if existing.A > 0 {
				blended := r.blendColor(existing, dirtColor, effects.Dirtiness*0.4)
				r.setPixel(img, px, py, blended)
			}
		}
	}

	// Edge roughness: randomly remove border pixels
	if effects.EdgeRoughness > 0 {
		numRemoves := maxInt(1, int(effects.EdgeRoughness*float64(w+h)*0.3))
		for i := 0; i < numRemoves; i++ {
			px := rng.Intn(w)
			py := rng.Intn(h)
			existing := r.getPixel(img, px, py)
			if existing.A > 0 && r.isEdgePixel(img, px, py, w, h) {
				r.setPixel(img, px, py, color.RGBA{A: 0})
			}
		}
	}
}

// --- Enchantment glow ---

func (r *EquipmentRenderer) applyEnchantGlow(img *ebiten.Image, w, h int, enchant EnchantmentGlow, rng *rand.Rand) {
	glowColor := r.enchantmentColor(enchant.Color)

	// Edge glow: find edge pixels and add colored halo
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			existing := r.getPixel(img, x, y)
			if existing.A > 0 && r.isEdgePixel(img, x, y, w, h) {
				intensity := enchant.Intensity * 0.6
				blended := r.blendColor(existing, glowColor, intensity)
				r.setPixel(img, x, y, blended)

				// Outer glow on adjacent transparent pixels
				for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
					nx, ny := x+d[0], y+d[1]
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						adj := r.getPixel(img, nx, ny)
						if adj.A == 0 {
							outerGlow := r.withAlpha(glowColor, uint8(enchant.Intensity*120))
							r.setPixel(img, nx, ny, outerGlow)
						}
					}
				}
			}
		}
	}

	// Sparkle particles for high-rarity items
	if enchant.ParticleCount > 4 {
		sparkleColor := r.lighten(glowColor, 60)
		numSparkles := enchant.ParticleCount / 3
		for i := 0; i < numSparkles; i++ {
			sx := rng.Intn(w)
			sy := rng.Intn(h)
			if sx >= 0 && sx < w && sy >= 0 && sy < h {
				r.setPixel(img, sx, sy, r.withAlpha(sparkleColor, uint8(150+rng.Intn(105))))
			}
		}
	}
}

// --- Color helpers ---

func (r *EquipmentRenderer) materialBaseColor(material MaterialType, rng *rand.Rand) color.RGBA {
	// Slight variation per seed for visual variety
	vary := func(base uint8, amount int) uint8 {
		return clampUint8(int(base) + rng.Intn(amount*2+1) - amount)
	}
	switch material {
	case MaterialMetal:
		return color.RGBA{R: vary(180, 20), G: vary(185, 15), B: vary(195, 15), A: 255}
	case MaterialLeather:
		return color.RGBA{R: vary(140, 20), G: vary(95, 15), B: vary(55, 10), A: 255}
	case MaterialCloth:
		return color.RGBA{R: vary(170, 30), G: vary(150, 30), B: vary(170, 30), A: 255}
	case MaterialWood:
		return color.RGBA{R: vary(130, 15), G: vary(90, 10), B: vary(50, 10), A: 255}
	case MaterialCrystal:
		return color.RGBA{R: vary(130, 25), G: vary(200, 20), B: vary(230, 15), A: 255}
	case MaterialEnergy:
		return color.RGBA{R: vary(80, 30), G: vary(150, 30), B: vary(240, 15), A: 255}
	default:
		return color.RGBA{R: vary(160, 20), G: vary(160, 20), B: vary(160, 20), A: 255}
	}
}

func (r *EquipmentRenderer) enchantmentColor(name string) color.RGBA {
	switch name {
	case "green":
		return color.RGBA{R: 50, G: 230, B: 100, A: 255}
	case "blue":
		return color.RGBA{R: 80, G: 140, B: 255, A: 255}
	case "purple":
		return color.RGBA{R: 180, G: 80, B: 255, A: 255}
	case "gold":
		return color.RGBA{R: 255, G: 210, B: 50, A: 255}
	case "red":
		return color.RGBA{R: 255, G: 60, B: 60, A: 255}
	default:
		return color.RGBA{R: 220, G: 220, B: 255, A: 255}
	}
}

func (r *EquipmentRenderer) lighten(c color.RGBA, amount int) color.RGBA {
	return color.RGBA{
		R: clampUint8(int(c.R) + amount),
		G: clampUint8(int(c.G) + amount),
		B: clampUint8(int(c.B) + amount),
		A: c.A,
	}
}

func (r *EquipmentRenderer) darken(c color.RGBA, amount int) color.RGBA {
	return color.RGBA{
		R: clampUint8(int(c.R) - amount),
		G: clampUint8(int(c.G) - amount),
		B: clampUint8(int(c.B) - amount),
		A: c.A,
	}
}

func (r *EquipmentRenderer) blendColor(a, b color.RGBA, t float64) color.RGBA {
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

func (r *EquipmentRenderer) withAlpha(c color.RGBA, a uint8) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: a}
}

func (r *EquipmentRenderer) setPixel(img *ebiten.Image, x, y int, c color.RGBA) {
	if c.A == 0 {
		return
	}
	img.Set(x, y, c)
}

func (r *EquipmentRenderer) getPixel(img *ebiten.Image, x, y int) (result color.RGBA) {
	defer func() {
		if recov := recover(); recov != nil {
			logrus.WithField("panic", recov).Debug("getPixel: recovered from panic (expected in tests without game loop)")
			result = color.RGBA{}
		}
	}()
	c := img.At(x, y)
	rr, gg, bb, aa := c.RGBA()
	return color.RGBA{R: uint8(rr >> 8), G: uint8(gg >> 8), B: uint8(bb >> 8), A: uint8(aa >> 8)}
}

func (r *EquipmentRenderer) isEdgePixel(img *ebiten.Image, x, y, w, h int) bool {
	// A pixel is an edge if any 4-neighbor is transparent
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		nx, ny := x+d[0], y+d[1]
		if nx < 0 || nx >= w || ny < 0 || ny >= h {
			return true
		}
		adj := r.getPixel(img, nx, ny)
		if adj.A == 0 {
			return true
		}
	}
	return false
}

func (r *EquipmentRenderer) drawFilledCircle(img *ebiten.Image, cx, cy, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := cx + dx
				py := cy + dy
				bounds := img.Bounds()
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					r.setPixel(img, px, py, c)
				}
			}
		}
	}
}

func (r *EquipmentRenderer) drawEllipseFilled(img *ebiten.Image, cx, cy, rx, ry int, base, highlight, shadow color.RGBA, w, h int) {
	if rx <= 0 || ry <= 0 {
		return
	}
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			// Ellipse equation
			ex := float64(dx) / float64(rx)
			ey := float64(dy) / float64(ry)
			if ex*ex+ey*ey <= 1.0 {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					dist := math.Sqrt(ex*ex + ey*ey)
					// Top-left highlight, bottom-right shadow
					lightAngle := (float64(dx)/float64(rx) - float64(dy)/float64(ry)) / 2.0
					if lightAngle > 0 {
						c := r.blendColor(base, highlight, lightAngle*0.4*(1.0-dist))
						r.setPixel(img, px, py, c)
					} else {
						c := r.blendColor(base, shadow, -lightAngle*0.3*(1.0-dist))
						r.setPixel(img, px, py, c)
					}
				}
			}
		}
	}
}

// maxInt is defined in hair_style_renderer.go

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
