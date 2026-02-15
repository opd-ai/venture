// Package sprites provides seed-based avatar variety for procedurally generated
// entity sprites. Each unique seed produces a distinct combination of skin tone,
// hair color, clothing colors, and body proportion modifiers, ensuring that
// different NPCs, players, and creatures are immediately distinguishable.
package sprites

import (
	"image/color"
	"math"
	"math/rand"
)

// AvatarTraits holds seed-derived visual characteristics for an entity sprite.
// These traits override default palette colors for body parts, producing unique
// appearances per entity while respecting genre palette constraints.
type AvatarTraits struct {
	// SkinTone is the base skin/body color.
	SkinTone color.RGBA
	// HairColor is the head/top color.
	HairColor color.RGBA
	// HairStyle is the seed-derived hairstyle type for aerial-view rendering.
	HairStyle HairStyle
	// ClothingPrimary is the main clothing/torso color.
	ClothingPrimary color.RGBA
	// ClothingSecondary is the accent clothing color (arms, trim).
	ClothingSecondary color.RGBA
	// LegColor is the lower body / pants / boots color.
	LegColor color.RGBA
	// ShoulderScale adjusts torso width (0.85–1.15, 1.0 = normal).
	ShoulderScale float64
	// HeadScale adjusts head size (0.90–1.10, 1.0 = normal).
	HeadScale float64
	// HeightScale adjusts overall vertical compression (0.92–1.08).
	HeightScale float64
}

// Predefined skin tone palettes spanning realistic human range plus fantasy tones.
var skinTones = []color.RGBA{
	{R: 255, G: 224, B: 196, A: 255}, // fair
	{R: 241, G: 194, B: 155, A: 255}, // light
	{R: 224, G: 172, B: 130, A: 255}, // light-medium
	{R: 198, G: 145, B: 104, A: 255}, // medium
	{R: 174, G: 119, B: 80, A: 255},  // medium-tan
	{R: 148, G: 95, B: 63, A: 255},   // tan
	{R: 124, G: 76, B: 48, A: 255},   // brown
	{R: 100, G: 60, B: 36, A: 255},   // dark brown
	{R: 78, G: 46, B: 28, A: 255},    // deep brown
	{R: 60, G: 35, B: 20, A: 255},    // very dark
}

// Hair color base palettes.
var hairColors = []color.RGBA{
	{R: 30, G: 20, B: 10, A: 255},    // black
	{R: 60, G: 40, B: 20, A: 255},    // dark brown
	{R: 100, G: 70, B: 40, A: 255},   // brown
	{R: 140, G: 100, B: 60, A: 255},  // light brown
	{R: 180, G: 140, B: 80, A: 255},  // dirty blonde
	{R: 220, G: 190, B: 120, A: 255}, // blonde
	{R: 240, G: 220, B: 170, A: 255}, // platinum
	{R: 160, G: 50, B: 30, A: 255},   // auburn
	{R: 200, G: 80, B: 40, A: 255},   // red
	{R: 110, G: 110, B: 120, A: 255}, // grey
	{R: 200, G: 200, B: 210, A: 255}, // white/silver
}

// Clothing hue presets — base hues that get varied per seed.
var clothingHues = []float64{
	0, 15, 30, 50, 80, 120, 160, 200, 220, 240, 270, 300, 330, 350,
}

// GenerateAvatarTraits produces unique visual traits from an entity seed.
// The same seed always produces the same traits (deterministic).
func GenerateAvatarTraits(seed int64) AvatarTraits {
	rng := rand.New(rand.NewSource(seed))

	// Select and vary skin tone
	skinIdx := rng.Intn(len(skinTones))
	skin := varyColor(skinTones[skinIdx], rng, 12)

	// Select and vary hair color
	hairIdx := rng.Intn(len(hairColors))
	hair := varyColor(hairColors[hairIdx], rng, 18)

	// Select hair style deterministically from seed
	hairStyle := HairStyle(rng.Intn(int(HairStyleCount)))

	// Generate clothing colors from hue wheel
	hueIdx := rng.Intn(len(clothingHues))
	baseHue := clothingHues[hueIdx] + rng.Float64()*20 - 10
	sat := 0.35 + rng.Float64()*0.45
	lum := 0.30 + rng.Float64()*0.35
	clothPrimary := hslToRGBA(baseHue, sat, lum)

	// Secondary clothing is offset in hue
	secHue := math.Mod(baseHue+90+rng.Float64()*60, 360)
	secSat := 0.25 + rng.Float64()*0.40
	secLum := 0.35 + rng.Float64()*0.30
	clothSecondary := hslToRGBA(secHue, secSat, secLum)

	// Leg color: darker variant of primary or secondary
	legHue := math.Mod(baseHue+rng.Float64()*40-20, 360)
	legLum := 0.18 + rng.Float64()*0.22
	legColor := hslToRGBA(legHue, sat*0.8, legLum)

	// Body proportion variety
	shoulderScale := 0.85 + rng.Float64()*0.30
	headScale := 0.90 + rng.Float64()*0.20
	heightScale := 0.92 + rng.Float64()*0.16

	return AvatarTraits{
		SkinTone:          skin,
		HairColor:         hair,
		HairStyle:         hairStyle,
		ClothingPrimary:   clothPrimary,
		ClothingSecondary: clothSecondary,
		LegColor:          legColor,
		ShoulderScale:     shoulderScale,
		HeadScale:         headScale,
		HeightScale:       heightScale,
	}
}

// GenerateCreatureTraits produces visual traits for nonhumanoid creatures.
// Uses the seed to vary base colors and proportions while keeping creature
// types recognizable.
func GenerateCreatureTraits(seed int64, entityType string) AvatarTraits {
	rng := rand.New(rand.NewSource(seed))

	// Base hue depends on creature type for recognizability
	var baseHue float64
	var satRange, lumRange [2]float64
	switch entityType {
	case "spider", "arachnid", "insect":
		baseHue = 25 + rng.Float64()*30 // brown-ish
		satRange = [2]float64{0.20, 0.50}
		lumRange = [2]float64{0.15, 0.35}
	case "serpent", "snake":
		baseHue = 80 + rng.Float64()*60 // green-ish
		satRange = [2]float64{0.40, 0.70}
		lumRange = [2]float64{0.25, 0.45}
	case "dragon", "flying", "winged":
		baseHue = rng.Float64() * 360 // any color
		satRange = [2]float64{0.50, 0.80}
		lumRange = [2]float64{0.30, 0.50}
	case "blob", "amorphous", "slime":
		baseHue = 90 + rng.Float64()*180 // green to purple
		satRange = [2]float64{0.50, 0.85}
		lumRange = [2]float64{0.35, 0.55}
	case "undead", "skeleton", "zombie":
		baseHue = 40 + rng.Float64()*20 // sickly yellow-green
		satRange = [2]float64{0.10, 0.30}
		lumRange = [2]float64{0.30, 0.50}
	default: // quadruped, beast, etc.
		baseHue = rng.Float64() * 60 // warm browns/reds
		satRange = [2]float64{0.25, 0.55}
		lumRange = [2]float64{0.25, 0.45}
	}

	sat := satRange[0] + rng.Float64()*(satRange[1]-satRange[0])
	lum := lumRange[0] + rng.Float64()*(lumRange[1]-lumRange[0])
	bodyColor := hslToRGBA(baseHue, sat, lum)

	// Belly/underside is lighter
	bellyColor := hslToRGBA(baseHue, sat*0.7, lum+0.15)

	// Accent markings
	markHue := math.Mod(baseHue+30+rng.Float64()*60, 360)
	markColor := hslToRGBA(markHue, sat*0.9, lum*0.8)

	return AvatarTraits{
		SkinTone:          bodyColor,
		HairColor:         markColor,
		ClothingPrimary:   bodyColor,
		ClothingSecondary: bellyColor,
		LegColor:          darkenRGBA(bodyColor, 0.7),
		ShoulderScale:     0.90 + rng.Float64()*0.25,
		HeadScale:         0.85 + rng.Float64()*0.30,
		HeightScale:       0.90 + rng.Float64()*0.20,
	}
}

// ColorForBodyPart returns the seed-varied color for a specific body part's color role.
// If the role doesn't map to an avatar trait, returns nil so the caller can fall back to the palette.
func (a *AvatarTraits) ColorForBodyPart(colorRole string, zIndex int) color.Color {
	switch colorRole {
	case "secondary":
		// Head and arms use skin/hair depending on z-index
		if zIndex >= 15 {
			return a.HairColor // head
		}
		return a.SkinTone // arms
	case "primary":
		if zIndex <= 5 {
			return a.LegColor // legs
		}
		return a.ClothingPrimary // torso
	case "accent1":
		return a.ClothingSecondary
	case "accent2":
		return a.HairColor
	default:
		return nil
	}
}

// ApplyProportions adjusts a PartSpec's relative dimensions using avatar trait scales.
func (a *AvatarTraits) ApplyProportions(spec PartSpec, bodyPart BodyPart) PartSpec {
	switch bodyPart {
	case PartHead:
		spec.RelativeWidth *= a.HeadScale
		spec.RelativeHeight *= a.HeadScale
	case PartTorso, PartArmor:
		spec.RelativeWidth *= a.ShoulderScale
	case PartArms:
		spec.RelativeWidth *= a.ShoulderScale
	case PartLegs:
		spec.RelativeHeight *= a.HeightScale
	}
	return spec
}

// varyColor adds small random variation to a base color.
func varyColor(base color.RGBA, rng *rand.Rand, maxDelta int) color.RGBA {
	vary := func(v uint8) uint8 {
		d := rng.Intn(maxDelta*2+1) - maxDelta
		result := int(v) + d
		if result < 0 {
			result = 0
		}
		if result > 255 {
			result = 255
		}
		return uint8(result)
	}
	return color.RGBA{R: vary(base.R), G: vary(base.G), B: vary(base.B), A: 255}
}

// hslToRGBA converts HSL to an RGBA color. H is in degrees [0,360), S and L in [0,1].
func hslToRGBA(h, s, l float64) color.RGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

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
		R: uint8(math.Round((r + m) * 255)),
		G: uint8(math.Round((g + m) * 255)),
		B: uint8(math.Round((b + m) * 255)),
		A: 255,
	}
}

// darkenRGBA returns a darker version of the color by the given factor (0-1).
func darkenRGBA(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

// IsHumanoidEntity returns true if the entity type should use humanoid avatar traits.
func IsHumanoidEntity(entityType string) bool {
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior", "merchant":
		return true
	}
	return false
}
