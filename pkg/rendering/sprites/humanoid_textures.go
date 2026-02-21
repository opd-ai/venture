// Package sprites provides seed-based humanoid surface textures for procedurally
// generated entity sprites. This extends the creature texture system to humanoid
// characters, adding skin texture variations (freckles, scars, weathering) and
// clothing fabric textures (linen, leather, silk, wool) that make every player
// character and NPC visually distinctive at 32×32 resolution from top-down view.
// All rendering is seed-deterministic, genre-aware, and cached for performance.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// HumanoidTextureType identifies surface texture applied to humanoid body parts.
type HumanoidTextureType int

const (
	// SkinSmooth is default smooth skin with subtle radial shading.
	SkinSmooth HumanoidTextureType = iota
	// SkinFreckled adds seed-based freckle dots concentrated on face/shoulders.
	SkinFreckled
	// SkinScarred adds linear scar marks with healed-over coloring.
	SkinScarred
	// SkinWeathered adds age/weather lines and darker creases.
	SkinWeathered
	// SkinTattooed adds geometric pattern marks in accent colors.
	SkinTattooed
	// FabricLinen adds fine crosshatch weave pattern.
	FabricLinen
	// FabricLeather adds subtle grain with wear highlights.
	FabricLeather
	// FabricSilk adds smooth sheen with light reflection.
	FabricSilk
	// FabricWool adds fuzzy dithered texture.
	FabricWool
	// FabricChainmail adds repeating ring pattern for armor.
	FabricChainmail
	// FabricPlate adds brushed metal sheen for plate armor.
	FabricPlate
	// HumTexHairStraight adds directional parallel strokes.
	HumTexHairStraight
	// HumTexHairWavy adds sinusoidal wave pattern.
	HumTexHairWavy
	// HumTexHairCurly adds tight spiral pattern.
	HumTexHairCurly
	// HumTexHairBraided adds interlocking diagonal patterns.
	HumTexHairBraided
	// humanoidTexTypeCount is sentinel for random selection.
	humanoidTexTypeCount
)

// String returns a human-readable name for the humanoid texture type.
func (t HumanoidTextureType) String() string {
	names := [...]string{
		"skin_smooth", "skin_freckled", "skin_scarred", "skin_weathered", "skin_tattooed",
		"fabric_linen", "fabric_leather", "fabric_silk", "fabric_wool",
		"fabric_chainmail", "fabric_plate",
		"hair_straight", "hair_wavy", "hair_curly", "hair_braided",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "unknown"
}

// HumanoidTextureParams configures how a humanoid texture is rendered.
type HumanoidTextureParams struct {
	// Type is the texture to render.
	Type HumanoidTextureType
	// Intensity controls how strongly the texture modifies base pixels (0.0-1.0).
	Intensity float64
	// Scale controls texture density (1.0=default, <1=denser, >1=sparser).
	Scale float64
	// PrimaryColor is the main texture overlay color (e.g., freckle color).
	PrimaryColor color.RGBA
	// SecondaryColor is used for highlights/shadows within the texture.
	SecondaryColor color.RGBA
	// Direction controls directional textures (0=up, PI/2=right, PI=down, etc.).
	Direction float64
}

// HumanoidTextureSet holds per-body-region texture assignments for humanoids.
type HumanoidTextureSet struct {
	SkinTexture    HumanoidTextureParams
	ClothingTop    HumanoidTextureParams
	ClothingBottom HumanoidTextureParams
	HairTexture    HumanoidTextureParams
}

// GenerateHumanoidTextureSet deterministically produces a texture set for humanoids.
// The seed ensures consistency, genre influences style choices.
func GenerateHumanoidTextureSet(seed int64, genre string) HumanoidTextureSet {
	rng := rand.New(rand.NewSource(seed ^ 0x484D4E44)) // "HMND" XOR

	set := HumanoidTextureSet{
		SkinTexture:    generateSkinTexture(rng, genre),
		ClothingTop:    generateClothingTexture(rng, genre, true),
		ClothingBottom: generateClothingTexture(rng, genre, false),
		HairTexture:    generateHairTexture(rng, genre),
	}

	return set
}

// generateSkinTexture creates a skin texture type and parameters.
func generateSkinTexture(rng *rand.Rand, genre string) HumanoidTextureParams {
	// 40% smooth, 25% freckled, 15% scarred, 15% weathered, 5% tattooed
	roll := rng.Float64()
	var skinType HumanoidTextureType
	switch {
	case roll < 0.40:
		skinType = SkinSmooth
	case roll < 0.65:
		skinType = SkinFreckled
	case roll < 0.80:
		skinType = SkinScarred
	case roll < 0.95:
		skinType = SkinWeathered
	default:
		skinType = SkinTattooed
	}

	// Genre biases
	switch genre {
	case "horror":
		if rng.Float64() < 0.4 {
			skinType = SkinScarred
		}
	case "post-apocalyptic", "postapoc":
		if rng.Float64() < 0.5 {
			skinType = SkinWeathered
		}
	case "cyberpunk":
		if rng.Float64() < 0.3 {
			skinType = SkinTattooed
		}
	}

	// Base skin undertone for texture overlays
	baseR := uint8(180 + rng.Intn(60))
	baseG := uint8(140 + rng.Intn(50))
	baseB := uint8(100 + rng.Intn(40))

	return HumanoidTextureParams{
		Type:           skinType,
		Intensity:      0.15 + rng.Float64()*0.20, // Subtle 0.15-0.35
		Scale:          0.8 + rng.Float64()*0.4,
		PrimaryColor:   color.RGBA{R: baseR - 40, G: baseG - 30, B: baseB - 20, A: 150},
		SecondaryColor: color.RGBA{R: baseR + 20, G: baseG + 15, B: baseB + 10, A: 120},
	}
}

// generateClothingTexture creates a fabric texture type and parameters.
func generateClothingTexture(rng *rand.Rand, genre string, isTop bool) HumanoidTextureParams {
	// Default distribution
	roll := rng.Float64()
	var fabricType HumanoidTextureType
	switch {
	case roll < 0.30:
		fabricType = FabricLinen
	case roll < 0.50:
		fabricType = FabricLeather
	case roll < 0.65:
		fabricType = FabricSilk
	case roll < 0.80:
		fabricType = FabricWool
	case roll < 0.90:
		fabricType = FabricChainmail
	default:
		fabricType = FabricPlate
	}

	// Genre biases
	switch genre {
	case "fantasy":
		if rng.Float64() < 0.4 && isTop {
			fabricType = FabricLeather
		}
	case "sci-fi", "scifi":
		if rng.Float64() < 0.5 {
			fabricType = FabricSilk // Synthetic smooth fabrics
		}
	case "cyberpunk":
		if rng.Float64() < 0.4 {
			fabricType = FabricLeather
		}
	case "horror":
		if rng.Float64() < 0.3 {
			fabricType = FabricWool // Worn, tattered look
		}
	case "post-apocalyptic", "postapoc":
		if rng.Float64() < 0.5 {
			fabricType = FabricLeather
		}
	}

	// Generate fabric colors
	hue := rng.Float64() * 360
	sat := 0.3 + rng.Float64()*0.5
	val := 0.3 + rng.Float64()*0.4
	r, g, b := hsvToRGB(hue, sat, val)

	return HumanoidTextureParams{
		Type:           fabricType,
		Intensity:      0.20 + rng.Float64()*0.25, // 0.20-0.45
		Scale:          0.8 + rng.Float64()*0.4,
		PrimaryColor:   color.RGBA{R: r, G: g, B: b, A: 160},
		SecondaryColor: color.RGBA{R: uint8(min255(int(r) + 30)), G: uint8(min255(int(g) + 25)), B: uint8(min255(int(b) + 20)), A: 140},
	}
}

// generateHairTexture creates a hair texture type and parameters.
func generateHairTexture(rng *rand.Rand, genre string) HumanoidTextureParams {
	roll := rng.Float64()
	var hairType HumanoidTextureType
	switch {
	case roll < 0.35:
		hairType = HumTexHairStraight
	case roll < 0.60:
		hairType = HumTexHairWavy
	case roll < 0.85:
		hairType = HumTexHairCurly
	default:
		hairType = HumTexHairBraided
	}

	// Hair colors: 60% dark, 25% medium, 15% light/red
	colorRoll := rng.Float64()
	var r, g, b uint8
	switch {
	case colorRoll < 0.60:
		// Dark hair (black to dark brown)
		r = uint8(30 + rng.Intn(40))
		g = uint8(20 + rng.Intn(30))
		b = uint8(15 + rng.Intn(25))
	case colorRoll < 0.85:
		// Medium hair (brown to auburn)
		r = uint8(80 + rng.Intn(50))
		g = uint8(50 + rng.Intn(40))
		b = uint8(30 + rng.Intn(30))
	default:
		// Light/red hair
		r = uint8(160 + rng.Intn(60))
		g = uint8(100 + rng.Intn(50))
		b = uint8(40 + rng.Intn(40))
	}

	// Genre variation
	if genre == "cyberpunk" && rng.Float64() < 0.3 {
		// Neon highlights possible
		r, g, b = uint8(rng.Intn(100)+100), uint8(rng.Intn(255)), uint8(rng.Intn(255))
	}

	direction := rng.Float64() * math.Pi * 2 // Random hair direction

	return HumanoidTextureParams{
		Type:           hairType,
		Intensity:      0.30 + rng.Float64()*0.25, // 0.30-0.55
		Scale:          0.7 + rng.Float64()*0.3,
		PrimaryColor:   color.RGBA{R: r, G: g, B: b, A: 180},
		SecondaryColor: color.RGBA{R: uint8(min255(int(r) + 40)), G: uint8(min255(int(g) + 35)), B: uint8(min255(int(b) + 30)), A: 150},
		Direction:      direction,
	}
}

// ApplyHumanoidTexture renders a humanoid texture pattern onto the given region.
// Only pixels with alpha > threshold are textured, preserving silhouette.
func ApplyHumanoidTexture(buf *image.RGBA, region image.Rectangle, params HumanoidTextureParams, seed int64) {
	if buf == nil || params.Intensity <= 0 {
		return
	}

	bounds := buf.Bounds()
	region = region.Intersect(bounds)
	if region.Empty() {
		return
	}

	rng := rand.New(rand.NewSource(seed ^ int64(params.Type)*8191))

	switch params.Type {
	case SkinSmooth:
		applySkinSmoothTexture(buf, region, params, rng)
	case SkinFreckled:
		applySkinFreckledTexture(buf, region, params, rng)
	case SkinScarred:
		applySkinScarredTexture(buf, region, params, rng)
	case SkinWeathered:
		applySkinWeatheredTexture(buf, region, params, rng)
	case SkinTattooed:
		applySkinTattooedTexture(buf, region, params, rng)
	case FabricLinen:
		applyFabricLinenTexture(buf, region, params, rng)
	case FabricLeather:
		applyFabricLeatherTexture(buf, region, params, rng)
	case FabricSilk:
		applyFabricSilkTexture(buf, region, params, rng)
	case FabricWool:
		applyFabricWoolTexture(buf, region, params, rng)
	case FabricChainmail:
		applyFabricChainmailTexture(buf, region, params, rng)
	case FabricPlate:
		applyFabricPlateTexture(buf, region, params, rng)
	case HumTexHairStraight:
		applyHairStraightTexture(buf, region, params, rng)
	case HumTexHairWavy:
		applyHairWavyTexture(buf, region, params, rng)
	case HumTexHairCurly:
		applyHairCurlyTexture(buf, region, params, rng)
	case HumTexHairBraided:
		applyHairBraidedTexture(buf, region, params, rng)
	}
}

// --- Skin Textures ---

// applySkinSmoothTexture adds subtle radial highlight variation.
func applySkinSmoothTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	cx := float64(r.Min.X+r.Max.X) / 2.0
	cy := float64(r.Min.Y+r.Max.Y) / 2.0
	maxR := float64(r.Dx()+r.Dy()) / 4.0
	if maxR < 1 {
		maxR = 1
	}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			dist := math.Sqrt(float64((x-int(cx))*(x-int(cx))+(y-int(cy))*(y-int(cy)))) / maxR
			// Subtle center highlight
			if dist < 0.5 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*(0.5-dist)*0.4)
			}
			// Very subtle noise for skin variation
			if rng.Float64() < 0.02 {
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.15)
			}
		}
	}
}

// applySkinFreckledTexture adds clustered freckle dots.
func applySkinFreckledTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	// First apply smooth base
	applySkinSmoothTexture(buf, r, p, rng)

	// Freckle concentration varies by position (more on nose/cheeks area)
	cx := float64(r.Min.X+r.Max.X) / 2.0
	cy := float64(r.Min.Y) + float64(r.Dy())*0.35 // Concentrate toward upper area (face)

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Distance from concentration center
			dist := math.Sqrt(float64((x-int(cx))*(x-int(cx))+(y-int(cy))*(y-int(cy)))) / float64(r.Dx())

			// Probability decreases with distance from center
			prob := 0.12 * math.Max(0, 1.0-dist*1.5)
			if rng.Float64() < prob {
				// Freckle dot: darker than base skin
				freckleColor := color.RGBA{
					R: uint8(max0(int(p.PrimaryColor.R) - 30)),
					G: uint8(max0(int(p.PrimaryColor.G) - 25)),
					B: uint8(max0(int(p.PrimaryColor.B) - 20)),
					A: 200,
				}
				blendHumanoidPixel(buf, x, y, freckleColor, p.Intensity*0.6)
			}
		}
	}
}

// applySkinScarredTexture adds linear scar marks.
func applySkinScarredTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	// First apply smooth base
	applySkinSmoothTexture(buf, r, p, rng)

	// 1-3 scar lines
	numScars := 1 + rng.Intn(3)
	for i := 0; i < numScars; i++ {
		// Random start point
		sx := r.Min.X + rng.Intn(r.Dx())
		sy := r.Min.Y + rng.Intn(r.Dy())
		// Random direction and length
		angle := rng.Float64() * math.Pi
		length := 3 + rng.Intn(humTexMaxInt(1, r.Dx()/2))

		// Scar is slightly lighter (healed tissue)
		scarColor := color.RGBA{
			R: uint8(min255(int(p.SecondaryColor.R) + 20)),
			G: uint8(min255(int(p.SecondaryColor.G) + 15)),
			B: uint8(min255(int(p.SecondaryColor.B) + 10)),
			A: 180,
		}

		dx := math.Cos(angle)
		dy := math.Sin(angle)
		for j := 0; j < length; j++ {
			px := sx + int(float64(j)*dx)
			py := sy + int(float64(j)*dy)
			if px >= r.Min.X && px < r.Max.X && py >= r.Min.Y && py < r.Max.Y {
				idx := py*buf.Stride + px*4
				if idx+3 < len(buf.Pix) && buf.Pix[idx+3] > 30 {
					blendHumanoidPixel(buf, px, py, scarColor, p.Intensity*0.5)
				}
			}
		}
	}
}

// applySkinWeatheredTexture adds age/weather lines and creases.
func applySkinWeatheredTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	// First apply smooth base
	applySkinSmoothTexture(buf, r, p, rng)

	// Add horizontal wrinkle lines (forehead, etc.)
	wSpacing := humTexMaxInt(2, int(3*p.Scale))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		isWrinkleLine := (y-r.Min.Y)%wSpacing == 0
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			if isWrinkleLine {
				// Darker crease line
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.4)
			}
			// General skin roughness
			if rng.Float64() < 0.05 {
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.2)
			}
		}
	}
}

// applySkinTattooedTexture adds geometric pattern marks.
func applySkinTattooedTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	// First apply smooth base
	applySkinSmoothTexture(buf, r, p, rng)

	// Tattoo pattern: geometric lines/spirals
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2

	// Tattoo ink color (dark blue-black or genre-specific)
	inkColor := color.RGBA{R: 30, G: 40, B: 80, A: 200}

	// Simple geometric pattern
	patternType := rng.Intn(3)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			draw := false
			switch patternType {
			case 0: // Diagonal stripes
				if (x+y)%(humTexMaxInt(1, int(4*p.Scale))) == 0 {
					draw = true
				}
			case 1: // Circles
				dist := math.Sqrt(float64((x-cx)*(x-cx) + (y-cy)*(y-cy)))
				if int(dist)%(humTexMaxInt(1, int(3*p.Scale))) == 0 {
					draw = true
				}
			case 2: // Crosshatch
				if (x-r.Min.X)%(humTexMaxInt(1, int(3*p.Scale))) == 0 || (y-r.Min.Y)%(humTexMaxInt(1, int(3*p.Scale))) == 0 {
					draw = true
				}
			}

			if draw {
				blendHumanoidPixel(buf, x, y, inkColor, p.Intensity*0.5)
			}
		}
	}
}

// --- Fabric Textures ---

// applyFabricLinenTexture adds fine crosshatch weave pattern.
func applyFabricLinenTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	spacing := humTexMaxInt(2, int(2*p.Scale))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Crosshatch pattern: both horizontal and vertical threads
			hThread := (y-r.Min.Y)%spacing == 0
			vThread := (x-r.Min.X)%spacing == 0

			if hThread && vThread {
				// Thread intersection: slightly brighter
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.4)
			} else if hThread || vThread {
				// Single thread: standard color
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.25)
			}
		}
	}
}

// applyFabricLeatherTexture adds subtle grain with wear highlights.
func applyFabricLeatherTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Random grain noise
			if rng.Float64() < 0.08 {
				if rng.Float64() < 0.5 {
					// Dark grain
					blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.3)
				} else {
					// Wear highlight
					blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.25)
				}
			}
		}
	}

	// Add edge wear (lighter at fold points)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < humTexMinInt(r.Min.X+2, r.Max.X); x++ {
			idx := y*buf.Stride + x*4
			if idx+3 < len(buf.Pix) && buf.Pix[idx+3] > 30 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.35)
			}
		}
		for x := humTexMaxInt(r.Min.X, r.Max.X-2); x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 < len(buf.Pix) && buf.Pix[idx+3] > 30 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.35)
			}
		}
	}
}

// applyFabricSilkTexture adds smooth sheen with light reflection.
func applyFabricSilkTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	cx := float64(r.Min.X+r.Max.X) / 2.0
	cy := float64(r.Min.Y+r.Max.Y) / 2.0

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Diagonal sheen highlight
			diagDist := math.Abs(float64(x-int(cx)) - float64(y-int(cy)))
			if diagDist < float64(r.Dx())*0.15 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.4)
			}
		}
	}
}

// applyFabricWoolTexture adds fuzzy dithered texture.
func applyFabricWoolTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Heavy dithering for fuzzy appearance
			if rng.Float64() < 0.25 {
				if rng.Float64() < 0.5 {
					blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.35)
				} else {
					blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.30)
				}
			}
		}
	}
}

// applyFabricChainmailTexture adds repeating ring pattern.
func applyFabricChainmailTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	ringSize := humTexMaxInt(2, int(3*p.Scale))
	halfRing := ringSize / 2

	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := (y - r.Min.Y) / ringSize
		xOffset := 0
		if row%2 == 1 {
			xOffset = halfRing
		}
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			localX := ((x - r.Min.X + xOffset) % ringSize) - halfRing
			localY := (y - r.Min.Y) % ringSize

			// Ring edge: darker (metal shadow)
			distFromCenter := math.Sqrt(float64(localX*localX + localY*localY))
			if distFromCenter > float64(halfRing)*0.6 && distFromCenter < float64(ringSize)*0.8 {
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.5)
			} else if distFromCenter < float64(halfRing)*0.4 {
				// Ring hole: darker
				blendHumanoidPixel(buf, x, y, color.RGBA{R: 40, G: 40, B: 50, A: 180}, p.Intensity*0.3)
			}
		}
	}
}

// applyFabricPlateTexture adds brushed metal sheen.
func applyFabricPlateTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	streakSpacing := humTexMaxInt(3, int(4*p.Scale))

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Horizontal brushed metal streaks
			localY := (y - r.Min.Y) % streakSpacing
			if localY == streakSpacing/2 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.5)
			}

			// Subtle diagonal grain
			diag := (x + y*2) % (streakSpacing * 2)
			if diag == 0 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.25)
			}
		}
	}
}

// --- Hair Textures ---

// applyHairStraightTexture adds directional parallel strokes.
func applyHairStraightTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	spacing := humTexMaxInt(2, int(2*p.Scale))
	dx := math.Cos(p.Direction)
	dy := math.Sin(p.Direction)

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Project point onto direction vector to create parallel lines
			proj := float64(x-r.Min.X)*dy - float64(y-r.Min.Y)*dx
			if int(math.Abs(proj))%spacing == 0 {
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.5)
			}

			// Occasional highlight strand
			if rng.Float64() < 0.03 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.4)
			}
		}
	}
}

// applyHairWavyTexture adds sinusoidal wave pattern.
func applyHairWavyTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	wavelength := float64(humTexMaxInt(4, int(5*p.Scale)))
	amplitude := float64(humTexMaxInt(1, int(2*p.Scale)))

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Sinusoidal offset based on y
			waveOffset := amplitude * math.Sin(float64(y-r.Min.Y)*2*math.Pi/wavelength)
			localX := float64(x - r.Min.X)

			// Check if we're on a wave crest
			if math.Abs(localX-waveOffset-float64(r.Dx())/2) < 1.5 {
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.5)
			}

			// Highlight on wave peaks
			if rng.Float64() < 0.04 {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.35)
			}
		}
	}
}

// applyHairCurlyTexture adds tight spiral pattern.
func applyHairCurlyTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	curlSize := humTexMaxInt(2, int(3*p.Scale))

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Checkerboard-like curl pattern
			cellX := (x - r.Min.X) / curlSize
			cellY := (y - r.Min.Y) / curlSize
			localX := (x - r.Min.X) % curlSize
			localY := (y - r.Min.Y) % curlSize

			// Alternate curls
			if (cellX+cellY)%2 == 0 {
				// Curl shadow on one side
				if localX < curlSize/2 {
					blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.4)
				}
			} else {
				// Curl highlight on other side
				if localY < curlSize/2 {
					blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.35)
				}
			}
		}
	}
}

// applyHairBraidedTexture adds interlocking diagonal patterns.
func applyHairBraidedTexture(buf *image.RGBA, r image.Rectangle, p HumanoidTextureParams, rng *rand.Rand) {
	braidWidth := humTexMaxInt(3, int(4*p.Scale))

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := y*buf.Stride + x*4
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}

			// Diagonal crossover pattern
			diag1 := (x - r.Min.X + y - r.Min.Y) % (braidWidth * 2)
			diag2 := (x - r.Min.X - y + r.Min.Y + r.Dy()) % (braidWidth * 2)

			// Create interlocking diamond pattern
			if diag1 < braidWidth && diag2 < braidWidth {
				blendHumanoidPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.45)
			} else if diag1 >= braidWidth && diag2 >= braidWidth {
				blendHumanoidPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.35)
			}
		}
	}
}

// --- Utility helpers ---

func blendHumanoidPixel(buf *image.RGBA, x, y int, overlay color.RGBA, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	idx := y*buf.Stride + x*4
	if idx+3 >= len(buf.Pix) {
		return
	}
	existingA := buf.Pix[idx+3]
	if existingA < 10 {
		return
	}

	a := alpha * float64(overlay.A) / 255.0
	invA := 1.0 - a
	buf.Pix[idx+0] = humanoidClampU8(float64(buf.Pix[idx+0])*invA + float64(overlay.R)*a)
	buf.Pix[idx+1] = humanoidClampU8(float64(buf.Pix[idx+1])*invA + float64(overlay.G)*a)
	buf.Pix[idx+2] = humanoidClampU8(float64(buf.Pix[idx+2])*invA + float64(overlay.B)*a)
}

func humanoidClampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
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

	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}

func min255(v int) int {
	if v > 255 {
		return 255
	}
	return v
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

func humTexMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func humTexMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
