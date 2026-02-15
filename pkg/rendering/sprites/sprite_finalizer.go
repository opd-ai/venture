// Package sprites provides sprite finalization with adaptive outlines and rim lighting.
// FinalizeEntitySprite applies post-processing to generated entity sprites:
// a 1px adaptive-color outline for terrain visibility, directional rim lighting
// for depth perception, and edge shadow for grounding. All effects are baked into
// the sprite image at generation time and cached, incurring zero per-frame cost.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// FinalizerConfig controls which post-processing effects are applied.
type FinalizerConfig struct {
	// EnableOutline enables 1px adaptive outline around opaque pixels.
	EnableOutline bool
	// OutlineDarkenFactor controls how dark the outline is relative to dominant color (0-1).
	OutlineDarkenFactor float64
	// EnableRimLight enables directional rim lighting along the top edge.
	EnableRimLight bool
	// RimLightIntensity controls rim light brightness boost (0-1).
	RimLightIntensity float64
	// RimLightAngle is the light direction in degrees (0=top, 90=right, 180=bottom, 270=left).
	RimLightAngle float64
	// EnableEdgeShadow enables subtle darkening along the bottom edge.
	EnableEdgeShadow bool
	// EdgeShadowIntensity controls the shadow darkening factor (0-1).
	EdgeShadowIntensity float64
	// Seed for deterministic variation.
	Seed int64
}

// DefaultFinalizerConfig returns a config suitable for most entity sprites.
func DefaultFinalizerConfig(seed int64) FinalizerConfig {
	return FinalizerConfig{
		EnableOutline:       true,
		OutlineDarkenFactor: 0.40,
		EnableRimLight:      true,
		RimLightIntensity:   0.30,
		RimLightAngle:       315, // top-left light
		EnableEdgeShadow:    true,
		EdgeShadowIntensity: 0.25,
		Seed:                seed,
	}
}

// FinalizeEntitySprite applies outline, rim lighting, and edge shadow to a sprite.
// The input image is not modified; a new image is returned.
// Uses seed for deterministic minor per-pixel variation so outlines look organic.
func FinalizeEntitySprite(img image.Image, cfg FinalizerConfig) *image.RGBA {
	if img == nil {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	// Build opacity map and copy source pixels
	src := image.NewRGBA(image.Rect(0, 0, w, h))
	opaque := make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			src.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8), G: uint8(g >> 8),
				B: uint8(b >> 8), A: uint8(a >> 8),
			})
			opaque[y*w+x] = a > 0x8000
		}
	}

	// Compute dominant edge color for adaptive outline
	dominantColor := computeDominantEdgeColor(src, opaque, w, h)

	result := image.NewRGBA(image.Rect(0, 0, w, h))

	// Pass 1: draw outline behind the sprite
	if cfg.EnableOutline {
		outlineColor := darkenColor(dominantColor, cfg.OutlineDarkenFactor)
		drawAdaptiveOutline(result, opaque, outlineColor, w, h, cfg.Seed)
	}

	// Pass 2: copy sprite pixels on top
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.RGBAAt(x, y)
			if c.A > 0 {
				result.SetRGBA(x, y, c)
			}
		}
	}

	// Pass 3: apply rim lighting to opaque edge pixels along light direction
	if cfg.EnableRimLight {
		applyRimLighting(result, opaque, w, h, cfg.RimLightAngle, cfg.RimLightIntensity, cfg.Seed)
	}

	// Pass 4: apply edge shadow on the opposite side
	if cfg.EnableEdgeShadow {
		shadowAngle := math.Mod(cfg.RimLightAngle+180, 360)
		applyEdgeShadow(result, opaque, w, h, shadowAngle, cfg.EdgeShadowIntensity)
	}

	return result
}

// computeDominantEdgeColor samples opaque pixels near the edge and averages their color.
func computeDominantEdgeColor(src *image.RGBA, opaque []bool, w, h int) color.RGBA {
	var rSum, gSum, bSum, count uint64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !opaque[y*w+x] {
				continue
			}
			if !isEdgePixel(opaque, x, y, w, h) {
				continue
			}
			c := src.RGBAAt(x, y)
			// Skip very dark shadow pixels
			if c.R < 30 && c.G < 30 && c.B < 30 {
				continue
			}
			rSum += uint64(c.R)
			gSum += uint64(c.G)
			bSum += uint64(c.B)
			count++
		}
	}
	if count == 0 {
		return color.RGBA{R: 20, G: 20, B: 25, A: 255}
	}
	return color.RGBA{
		R: uint8(rSum / count),
		G: uint8(gSum / count),
		B: uint8(bSum / count),
		A: 255,
	}
}

// isEdgePixel returns true if an opaque pixel borders at least one transparent pixel.
func isEdgePixel(opaque []bool, x, y, w, h int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				return true // border of image counts as edge
			}
			if !opaque[ny*w+nx] {
				return true
			}
		}
	}
	return false
}

// drawAdaptiveOutline draws a 1px outline in transparent pixels adjacent to opaque pixels.
// Adds subtle per-pixel variation using the seed so outlines look hand-drawn.
func drawAdaptiveOutline(dst *image.RGBA, opaque []bool, outlineColor color.RGBA, w, h int, seed int64) {
	rng := rand.New(rand.NewSource(seed ^ 0x4F55544C))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if opaque[y*w+x] {
				continue // skip opaque pixels
			}
			if !hasOpaqueNeighbor(opaque, x, y, w, h) {
				continue
			}
			// Slight color variation for organic feel
			vary := rng.Intn(15) - 7
			c := color.RGBA{
				R: clampU8(float64(int(outlineColor.R) + vary)),
				G: clampU8(float64(int(outlineColor.G) + vary)),
				B: clampU8(float64(int(outlineColor.B) + vary)),
				A: 220 + uint8(rng.Intn(36)), // 220-255 alpha
			}
			dst.SetRGBA(x, y, c)
		}
	}
}

// hasOpaqueNeighbor returns true if any 4-connected neighbor is opaque.
func hasOpaqueNeighbor(opaque []bool, x, y, w, h int) bool {
	offsets := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, o := range offsets {
		nx, ny := x+o[0], y+o[1]
		if nx >= 0 && ny >= 0 && nx < w && ny < h && opaque[ny*w+nx] {
			return true
		}
	}
	return false
}

// applyRimLighting brightens opaque edge pixels facing the light direction.
func applyRimLighting(dst *image.RGBA, opaque []bool, w, h int, lightAngle, intensity float64, seed int64) {
	// Light direction as unit vector
	rad := lightAngle * math.Pi / 180
	lx := math.Cos(rad)
	ly := -math.Sin(rad) // screen Y is inverted

	rng := rand.New(rand.NewSource(seed ^ 0x52494D4C))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !opaque[y*w+x] {
				continue
			}
			if !isEdgePixel(opaque, x, y, w, h) {
				continue
			}

			// Compute edge normal (direction toward nearest transparent pixel)
			nx, ny := edgeNormal(opaque, x, y, w, h)
			// Dot product: how much does the edge face the light?
			dot := nx*lx + ny*ly
			if dot <= 0 {
				continue
			}

			// Scale brightness by dot product and intensity
			boost := dot * intensity
			// Tiny random variation
			boost += (rng.Float64() - 0.5) * 0.05

			c := dst.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			c.R = clampU8(float64(c.R) + boost*80)
			c.G = clampU8(float64(c.G) + boost*80)
			c.B = clampU8(float64(c.B) + boost*80)
			dst.SetRGBA(x, y, c)
		}
	}
}

// applyEdgeShadow darkens opaque edge pixels on the shadow side.
func applyEdgeShadow(dst *image.RGBA, opaque []bool, w, h int, shadowAngle, intensity float64) {
	rad := shadowAngle * math.Pi / 180
	sx := math.Cos(rad)
	sy := -math.Sin(rad)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !opaque[y*w+x] {
				continue
			}
			if !isEdgePixel(opaque, x, y, w, h) {
				continue
			}

			nx, ny := edgeNormal(opaque, x, y, w, h)
			dot := nx*sx + ny*sy
			if dot <= 0 {
				continue
			}

			darken := dot * intensity
			c := dst.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			c.R = clampU8(float64(c.R) * (1 - darken))
			c.G = clampU8(float64(c.G) * (1 - darken))
			c.B = clampU8(float64(c.B) * (1 - darken))
			dst.SetRGBA(x, y, c)
		}
	}
}

// edgeNormal computes the approximate outward-facing normal of an edge pixel.
func edgeNormal(opaque []bool, x, y, w, h int) (float64, float64) {
	var nx, ny float64
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			px, py := x+dx, y+dy
			if px < 0 || py < 0 || px >= w || py >= h || !opaque[py*w+px] {
				// Transparent neighbor: normal points toward it
				nx += float64(dx)
				ny += float64(dy)
			}
		}
	}
	length := math.Sqrt(nx*nx + ny*ny)
	if length < 0.001 {
		return 0, -1 // default: upward
	}
	return nx / length, ny / length
}

// darkenColor returns a color darkened by the given factor (0=no change, 1=black).
func darkenColor(c color.RGBA, factor float64) color.RGBA {
	f := 1 - factor
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: c.A,
	}
}


