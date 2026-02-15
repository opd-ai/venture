// Package sprites provides seed-based procedural surface texture patterns for
// entity body parts. Each creature form (fur, scales, chitin, metal, bone,
// ooze, feathers, bark) receives a distinct pixel-level micro-texture that
// makes it immediately recognizable at 32×32 resolution from the top-down
// aerial perspective. All rendering is seed-deterministic, genre-aware, and
// operates on standard image.RGBA buffers for compositing.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// SurfaceTextureType identifies the visual surface texture applied to a body part.
type SurfaceTextureType int

const (
	// TexNone applies no texture (smooth skin).
	TexNone SurfaceTextureType = iota
	// TexFur renders a directional noise dither simulating animal fur.
	TexFur
	// TexScales renders overlapping semi-circular tiles for reptiles/serpents.
	TexScales
	// TexChitin renders hard shell segmentation with specular highlights (insects).
	TexChitin
	// TexMetal renders directional specular streaks and rivet dots (mechanicals).
	TexMetal
	// TexBone renders cracked lines and rough pitting (undead/skeletons).
	TexBone
	// TexOoze renders translucent internal bubble/nucleus pattern (blobs/slimes).
	TexOoze
	// TexFeathers renders overlapping chevron patterns (flying creatures).
	TexFeathers
	// TexBark renders vertical grain lines for treants/plant creatures.
	TexBark
	// texTypeCount is a sentinel for random selection.
	texTypeCount
)

// String returns a human-readable name for the texture type.
func (t SurfaceTextureType) String() string {
	names := [...]string{
		"none", "fur", "scales", "chitin", "metal",
		"bone", "ooze", "feathers", "bark",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "unknown"
}

// SurfaceTextureParams configures how a texture is rendered onto a region.
type SurfaceTextureParams struct {
	// Type is the texture to render.
	Type SurfaceTextureType
	// Intensity controls how strongly the texture modifies base pixels (0.0-1.0).
	Intensity float64
	// Scale controls texture density (1.0=default, <1=denser, >1=sparser).
	Scale float64
	// PrimaryColor is the main texture overlay color.
	PrimaryColor color.RGBA
	// SecondaryColor is used for highlights/accents within the texture.
	SecondaryColor color.RGBA
}

// SurfaceTextureSet holds per-body-region texture assignments.
type SurfaceTextureSet struct {
	HeadTexture  SurfaceTextureParams
	TorsoTexture SurfaceTextureParams
	LimbTexture  SurfaceTextureParams
}

// TextureForCreatureForm returns the default texture type for a creature form.
func TextureForCreatureForm(form string) SurfaceTextureType {
	switch form {
	case "quadruped":
		return TexFur
	case "serpentine":
		return TexScales
	case "arachnid":
		return TexChitin
	case "mechanical":
		return TexMetal
	case "undead":
		return TexBone
	case "blob":
		return TexOoze
	case "flying":
		return TexFeathers
	case "humanoid":
		return TexNone
	default:
		return TexNone
	}
}

// GenerateSurfaceTextureSet deterministically produces a texture set from a seed
// and creature form. The genre influences color choices and intensity.
func GenerateSurfaceTextureSet(seed int64, form, genre string) SurfaceTextureSet {
	rng := rand.New(rand.NewSource(seed ^ 0x53555246)) // "SURF" XOR

	texType := TextureForCreatureForm(form)
	if texType == TexNone {
		return SurfaceTextureSet{}
	}

	baseIntensity := 0.25 + rng.Float64()*0.25 // 0.25-0.50
	baseScale := 0.8 + rng.Float64()*0.4       // 0.8-1.2

	primary, secondary := textureColors(texType, genre, rng)

	set := SurfaceTextureSet{
		HeadTexture: SurfaceTextureParams{
			Type:           texType,
			Intensity:      baseIntensity * 0.7, // Lighter on head
			Scale:          baseScale * 0.8,     // Finer on head
			PrimaryColor:   primary,
			SecondaryColor: secondary,
		},
		TorsoTexture: SurfaceTextureParams{
			Type:           texType,
			Intensity:      baseIntensity,
			Scale:          baseScale,
			PrimaryColor:   primary,
			SecondaryColor: secondary,
		},
		LimbTexture: SurfaceTextureParams{
			Type:           texType,
			Intensity:      baseIntensity * 0.85,
			Scale:          baseScale * 1.1, // Slightly coarser on limbs
			PrimaryColor:   primary,
			SecondaryColor: secondary,
		},
	}

	applyGenreTextureBias(&set, genre)
	return set
}

// textureColors picks genre-aware primary/secondary colors for a texture type.
func textureColors(tex SurfaceTextureType, genre string, rng *rand.Rand) (color.RGBA, color.RGBA) {
	vary := func(base uint8, amount int) uint8 {
		v := int(base) + rng.Intn(amount*2+1) - amount
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		return uint8(v)
	}

	switch tex {
	case TexFur:
		// Warm brown tones with lighter highlights
		r := vary(140, 30)
		g := vary(100, 25)
		b := vary(60, 20)
		return color.RGBA{R: r, G: g, B: b, A: 180},
			color.RGBA{R: vary(r+30, 10), G: vary(g+25, 10), B: vary(b+20, 10), A: 120}

	case TexScales:
		// Green/teal with darker edges
		r := vary(40, 20)
		g := vary(120, 30)
		b := vary(80, 25)
		if genre == "horror" {
			r, g, b = vary(80, 15), vary(60, 15), vary(50, 15)
		}
		return color.RGBA{R: r, G: g, B: b, A: 160},
			color.RGBA{R: vary(r-20, 8), G: vary(g-20, 8), B: vary(b-15, 8), A: 200}

	case TexChitin:
		// Dark glossy brown/black with specular white highlight
		r := vary(50, 15)
		g := vary(35, 10)
		b := vary(20, 10)
		return color.RGBA{R: r, G: g, B: b, A: 180},
			color.RGBA{R: 220, G: 220, B: 230, A: 100} // specular

	case TexMetal:
		// Silver/steel with bright specular
		r := vary(160, 20)
		g := vary(165, 20)
		b := vary(175, 20)
		if genre == "cyberpunk" {
			r, g = vary(100, 15), vary(180, 15)
		}
		return color.RGBA{R: r, G: g, B: b, A: 150},
			color.RGBA{R: 240, G: 245, B: 255, A: 160}

	case TexBone:
		// Off-white with dark cracks
		return color.RGBA{R: vary(200, 15), G: vary(190, 15), B: vary(170, 15), A: 140},
			color.RGBA{R: vary(50, 10), G: vary(40, 10), B: vary(30, 10), A: 180}

	case TexOoze:
		// Translucent green with bright nuclei
		r := vary(30, 15)
		g := vary(180, 25)
		b := vary(50, 20)
		if genre == "horror" {
			r, g, b = vary(120, 15), vary(40, 15), vary(30, 15) // Blood ooze
		}
		return color.RGBA{R: r, G: g, B: b, A: 100},
			color.RGBA{R: vary(200, 20), G: vary(255, 10), B: vary(200, 20), A: 200}

	case TexFeathers:
		// Neutral with lighter barb highlights
		r := vary(100, 30)
		g := vary(90, 25)
		b := vary(80, 25)
		return color.RGBA{R: r, G: g, B: b, A: 150},
			color.RGBA{R: vary(r+40, 15), G: vary(g+35, 15), B: vary(b+30, 15), A: 120}

	case TexBark:
		// Dark brown with lighter grain
		return color.RGBA{R: vary(80, 15), G: vary(55, 12), B: vary(30, 10), A: 170},
			color.RGBA{R: vary(120, 15), G: vary(90, 12), B: vary(50, 10), A: 130}
	}

	return color.RGBA{A: 0}, color.RGBA{A: 0}
}

// applyGenreTextureBias adjusts texture parameters based on genre mood.
func applyGenreTextureBias(set *SurfaceTextureSet, genre string) {
	switch genre {
	case "horror":
		// Heavier, darker textures
		set.HeadTexture.Intensity *= 1.2
		set.TorsoTexture.Intensity *= 1.3
		set.LimbTexture.Intensity *= 1.2
	case "fantasy":
		// Slightly lighter, more colorful
		set.TorsoTexture.Intensity *= 0.9
	case "cyberpunk":
		// Sharper, more defined for mechanical types
		set.TorsoTexture.Scale *= 0.85
	case "sci-fi", "scifi":
		set.TorsoTexture.Scale *= 0.9
	}

	// Clamp intensities
	clamp := func(v *float64) {
		if *v > 0.7 {
			*v = 0.7
		}
		if *v < 0 {
			*v = 0
		}
	}
	clamp(&set.HeadTexture.Intensity)
	clamp(&set.TorsoTexture.Intensity)
	clamp(&set.LimbTexture.Intensity)
}

// ApplySurfaceTexture renders a surface texture pattern onto the given region
// of an image.RGBA buffer. Only pixels with alpha > threshold are textured,
// preserving the existing silhouette. The texture modifies existing colors
// by blending, never introducing new opaque pixels.
func ApplySurfaceTexture(buf *image.RGBA, region image.Rectangle, params SurfaceTextureParams, seed int64) {
	if buf == nil || params.Type == TexNone || params.Intensity <= 0 {
		return
	}

	bounds := buf.Bounds()
	region = region.Intersect(bounds)
	if region.Empty() {
		return
	}

	rng := rand.New(rand.NewSource(seed ^ int64(params.Type)*7907))

	switch params.Type {
	case TexFur:
		applyFurTexture(buf, region, params, rng)
	case TexScales:
		applyScaleTexture(buf, region, params, rng)
	case TexChitin:
		applyChitinTexture(buf, region, params, rng)
	case TexMetal:
		applyMetalTexture(buf, region, params, rng)
	case TexBone:
		applyBoneTexture(buf, region, params, rng)
	case TexOoze:
		applyOozeTexture(buf, region, params, rng)
	case TexFeathers:
		applyFeatherTexture(buf, region, params, rng)
	case TexBark:
		applyBarkTexture(buf, region, params, rng)
	}
}

// --- Fur: directional noise dither simulating tufts ---

func applyFurTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	spacing := int(math.Max(2, 3*p.Scale))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			// Fur has directional noise: diagonal strokes with variation
			phase := float64(x+y*3) / float64(spacing)
			noise := math.Sin(phase*2.7+float64(rng.Intn(10))*0.3) * 0.5
			if rng.Float64() > 0.6 {
				noise += rng.Float64()*0.4 - 0.2
			}
			if math.Abs(noise) > 0.3 {
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*math.Abs(noise))
			}
			// Occasional highlight tuft
			if rng.Float64() < 0.08 {
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.5)
			}
		}
	}
}

// --- Scales: overlapping semi-circular tiles ---

func applyScaleTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	scaleSize := int(math.Max(2, 3*p.Scale))
	halfScale := scaleSize / 2
	if halfScale < 1 {
		halfScale = 1
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := (y - r.Min.Y) / scaleSize
		xOffset := 0
		if row%2 == 1 {
			xOffset = halfScale // Stagger alternate rows
		}
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			localX := ((x - r.Min.X + xOffset) % scaleSize) - halfScale
			localY := (y - r.Min.Y) % scaleSize
			// Semi-circle: darken edges of each scale tile
			distFromCenter := math.Sqrt(float64(localX*localX+localY*localY)) / float64(scaleSize)
			if distFromCenter > 0.6 {
				// Scale edge — darker
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.7)
			} else if distFromCenter < 0.25 {
				// Scale center — highlight
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.3)
			}
		}
	}
}

// --- Chitin: hard shell segments with specular highlights ---

func applyChitinTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	segH := int(math.Max(3, 4*p.Scale))
	w := r.Dx()
	for y := r.Min.Y; y < r.Max.Y; y++ {
		segIdx := (y - r.Min.Y) / segH
		localY := (y - r.Min.Y) % segH
		isSegBorder := localY == 0 || localY == segH-1
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			if isSegBorder {
				// Dark groove between segments
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.8)
			} else {
				// Specular highlight along center of each segment
				centerDist := math.Abs(float64(x-r.Min.X)-float64(w)/2) / float64(w)
				if centerDist < 0.15 && localY == segH/2 {
					blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.6)
				}
				// Subtle noise for rough chitin
				if rng.Float64() < 0.05 && segIdx%2 == 0 {
					blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.2)
				}
			}
		}
	}
}

// --- Metal: directional specular streaks and rivets ---

func applyMetalTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	streakSpacing := int(math.Max(3, 4*p.Scale))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			// Horizontal specular streaks (brushed metal)
			localY := (y - r.Min.Y) % streakSpacing
			if localY == streakSpacing/2 {
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.5)
			}
			// Subtle diagonal grain
			diag := (x + y*2) % (streakSpacing * 2)
			if diag == 0 {
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.3)
			}
		}
	}

	// Rivet dots at corners
	rivetPos := [][2]int{
		{r.Min.X + 1, r.Min.Y + 1},
		{r.Max.X - 2, r.Min.Y + 1},
		{r.Min.X + 1, r.Max.Y - 2},
		{r.Max.X - 2, r.Max.Y - 2},
	}
	for _, pos := range rivetPos {
		if pos[0] >= r.Min.X && pos[0] < r.Max.X && pos[1] >= r.Min.Y && pos[1] < r.Max.Y {
			blendPixel(buf, pos[0], pos[1], p.SecondaryColor, p.Intensity*0.7)
		}
	}
}

// --- Bone: cracked lines and rough pitting ---

func applyBoneTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	// Random crack lines
	numCracks := 2 + rng.Intn(3)
	for i := 0; i < numCracks; i++ {
		cx := r.Min.X + rng.Intn(r.Dx())
		cy := r.Min.Y + rng.Intn(r.Dy())
		length := 2 + rng.Intn(r.Dx()/2)
		dx := rng.Float64()*2 - 1
		dy := rng.Float64()*2 - 1
		for j := 0; j < length; j++ {
			px := cx + int(float64(j)*dx)
			py := cy + int(float64(j)*dy)
			if px >= r.Min.X && px < r.Max.X && py >= r.Min.Y && py < r.Max.Y {
				idx := (py*buf.Stride + px*4)
				if idx+3 < len(buf.Pix) && buf.Pix[idx+3] > 30 {
					blendPixel(buf, px, py, p.SecondaryColor, p.Intensity*0.7)
				}
			}
			// Slight jitter
			if rng.Float64() < 0.3 {
				dx += rng.Float64()*0.4 - 0.2
			}
		}
	}

	// Rough pitting noise
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			if rng.Float64() < 0.08 {
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.25)
			}
		}
	}
}

// --- Ooze: translucent internal bubbles and bright nuclei ---

func applyOozeTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2
	maxR := float64(r.Dx()+r.Dy()) / 4

	// Radial translucency gradient (edges more opaque than center)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 20 {
				continue
			}
			dist := math.Sqrt(float64((x-cx)*(x-cx)+(y-cy)*(y-cy))) / maxR
			// Center is lighter/brighter
			if dist < 0.4 {
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*(0.4-dist)*0.8)
			}
		}
	}

	// 2-4 internal bubble nuclei
	numBubbles := 2 + rng.Intn(3)
	for i := 0; i < numBubbles; i++ {
		bx := r.Min.X + 1 + rng.Intn(surfMaxInt(1, r.Dx()-2))
		by := r.Min.Y + 1 + rng.Intn(surfMaxInt(1, r.Dy()-2))
		// Bright nucleus dot
		if bx >= r.Min.X && bx < r.Max.X && by >= r.Min.Y && by < r.Max.Y {
			idx := (by*buf.Stride + bx*4)
			if idx+3 < len(buf.Pix) && buf.Pix[idx+3] > 20 {
				blendPixel(buf, bx, by, p.SecondaryColor, p.Intensity*0.9)
			}
		}
	}
}

// --- Feathers: overlapping chevron pattern ---

func applyFeatherTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	barbSpacing := int(math.Max(2, 3*p.Scale))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := (y - r.Min.Y) / barbSpacing
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			localX := x - (r.Min.X+r.Max.X)/2
			localY := (y - r.Min.Y) % barbSpacing
			// Chevron: V-shaped barbs
			chevronY := surfAbsInt(localX) / 2
			if localY == chevronY%barbSpacing {
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.5)
			}
			// Highlight on rachis (center line)
			if surfAbsInt(localX) <= 0 && row%2 == 0 {
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.4)
			}
		}
	}
}

// --- Bark: vertical grain lines ---

func applyBarkTexture(buf *image.RGBA, r image.Rectangle, p SurfaceTextureParams, rng *rand.Rand) {
	grainSpacing := int(math.Max(2, 3*p.Scale))
	// Seed-stable grain positions
	grainOffsets := make([]int, r.Dx()+1)
	for i := range grainOffsets {
		grainOffsets[i] = rng.Intn(2) // slight offset per column
	}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := (y*buf.Stride + x*4)
			if idx+3 >= len(buf.Pix) || buf.Pix[idx+3] < 30 {
				continue
			}
			col := x - r.Min.X
			grainOffset := 0
			if col < len(grainOffsets) {
				grainOffset = grainOffsets[col]
			}
			localX := (col + grainOffset) % grainSpacing
			if localX == 0 {
				// Dark grain line
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.6)
			} else if localX == grainSpacing/2 {
				// Lighter highlight between grains
				blendPixel(buf, x, y, p.SecondaryColor, p.Intensity*0.25)
			}
			// Occasional knot
			if rng.Float64() < 0.02 {
				blendPixel(buf, x, y, p.PrimaryColor, p.Intensity*0.4)
			}
		}
	}
}

// --- Utility helpers ---

// blendPixel blends an overlay color onto an existing pixel at the given alpha factor.
func blendPixel(buf *image.RGBA, x, y int, overlay color.RGBA, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	idx := (y*buf.Stride + x*4)
	if idx+3 >= len(buf.Pix) {
		return
	}
	existingA := buf.Pix[idx+3]
	if existingA < 10 {
		return
	}

	a := alpha * float64(overlay.A) / 255.0
	invA := 1.0 - a
	buf.Pix[idx+0] = surfClampU8(float64(buf.Pix[idx+0])*invA + float64(overlay.R)*a)
	buf.Pix[idx+1] = surfClampU8(float64(buf.Pix[idx+1])*invA + float64(overlay.G)*a)
	buf.Pix[idx+2] = surfClampU8(float64(buf.Pix[idx+2])*invA + float64(overlay.B)*a)
}

func surfClampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func surfAbsInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func surfMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
