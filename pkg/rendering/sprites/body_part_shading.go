// Package sprites provides per-body-part depth shading for procedurally
// generated entity sprites. It applies radial highlight, edge darkening,
// and genre-aware tinting to each body part during sprite generation,
// transforming flat single-color shapes into shaded, visually rich sprites.
package sprites

import (
	"image"
	"image/color"
	"math"
)

// ShadingConfig controls how depth shading is applied to a body part.
type ShadingConfig struct {
	// LightAngle is the direction of the overhead light in radians (0 = top).
	LightAngle float64
	// LightIntensity controls highlight strength (0.0-1.0).
	LightIntensity float64
	// EdgeDarkening controls how much darker edges become (0.0-1.0).
	EdgeDarkening float64
	// AmbientOcclusion controls darkening at overlapping boundaries (0.0-1.0).
	AmbientOcclusion float64
	// DitherStrength adds subtle noise texture (0.0-1.0).
	DitherStrength float64
	// TintR, TintG, TintB is the light color tint (1.0 = neutral white).
	TintR, TintG, TintB float64
}

// DefaultShadingConfig returns shading config suitable for top-down aerial sprites.
func DefaultShadingConfig() ShadingConfig {
	return ShadingConfig{
		LightAngle:       0,
		LightIntensity:   0.35,
		EdgeDarkening:    0.25,
		AmbientOcclusion: 0.15,
		DitherStrength:   0.06,
		TintR:            1.0,
		TintG:            0.98,
		TintB:            0.93,
	}
}

// GenreShadingConfig returns genre-specific shading parameters.
func GenreShadingConfig(genre string) ShadingConfig {
	cfg := DefaultShadingConfig()
	switch genre {
	case "horror":
		cfg.LightIntensity = 0.20
		cfg.EdgeDarkening = 0.40
		cfg.AmbientOcclusion = 0.30
		cfg.TintR, cfg.TintG, cfg.TintB = 0.85, 0.75, 0.70
	case "cyberpunk":
		cfg.LightIntensity = 0.40
		cfg.EdgeDarkening = 0.30
		cfg.DitherStrength = 0.08
		cfg.TintR, cfg.TintG, cfg.TintB = 0.80, 0.95, 1.0
	case "sci-fi", "scifi":
		cfg.LightIntensity = 0.38
		cfg.EdgeDarkening = 0.22
		cfg.TintR, cfg.TintG, cfg.TintB = 0.90, 0.95, 1.0
	case "post-apocalyptic", "postapoc":
		cfg.LightIntensity = 0.28
		cfg.EdgeDarkening = 0.35
		cfg.DitherStrength = 0.10
		cfg.TintR, cfg.TintG, cfg.TintB = 1.0, 0.90, 0.75
	case "fantasy":
		cfg.LightIntensity = 0.35
		cfg.EdgeDarkening = 0.25
		cfg.TintR, cfg.TintG, cfg.TintB = 1.0, 0.98, 0.90
	}
	return cfg
}

// ApplyBodyPartShading modifies pixels in the given RGBA image to apply
// radial highlight, edge darkening, and dithering. The image should contain
// a single rendered body part shape (non-transparent pixels are the shape).
// seed provides deterministic dithering.
func ApplyBodyPartShading(img *image.RGBA, cfg ShadingConfig, seed int64) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w == 0 || h == 0 {
		return
	}

	// Precompute center and light direction offset
	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	lightDx := math.Sin(cfg.LightAngle) * 0.3
	lightDy := -math.Cos(cfg.LightAngle) * 0.3

	// Precompute max radius for normalization
	maxR := math.Sqrt(cx*cx + cy*cy)
	if maxR < 1 {
		maxR = 1
	}

	// Build a simple alpha mask for edge detection (distance to nearest transparent pixel)
	alphaMap := buildAlphaMap(img)

	// Deterministic dither hash
	dSeed := uint64(seed)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			a := img.Pix[idx+3]
			if a == 0 {
				continue // skip fully transparent pixels
			}

			r := float64(img.Pix[idx])
			g := float64(img.Pix[idx+1])
			b := float64(img.Pix[idx+2])

			// 1. Radial highlight: brighter near light-shifted center
			nx := (float64(x) - cx) / cx
			ny := (float64(y) - cy) / cy
			distFromLight := math.Sqrt((nx-lightDx)*(nx-lightDx) + (ny-lightDy)*(ny-lightDy))
			highlight := (1.0 - clampF(distFromLight/1.4, 0, 1)) * cfg.LightIntensity
			r += highlight * 60.0 * cfg.TintR
			g += highlight * 60.0 * cfg.TintG
			b += highlight * 60.0 * cfg.TintB

			// 2. Edge darkening: darker pixels near shape boundaries
			edgeDist := float64(alphaMap[y*w+x])
			maxEdge := 3.0 // pixels from edge that are darkened
			if edgeDist < maxEdge {
				edgeFactor := (1.0 - edgeDist/maxEdge) * cfg.EdgeDarkening
				r *= (1.0 - edgeFactor)
				g *= (1.0 - edgeFactor)
				b *= (1.0 - edgeFactor)
			}

			// 3. Ambient occlusion: darker toward bottom of body part
			aoFactor := clampF(float64(y)/float64(h), 0, 1) * cfg.AmbientOcclusion
			r *= (1.0 - aoFactor)
			g *= (1.0 - aoFactor)
			b *= (1.0 - aoFactor)

			// 4. Sub-pixel dithering (deterministic noise)
			if cfg.DitherStrength > 0 {
				dither := ditherNoise(x, y, dSeed) * cfg.DitherStrength * 20.0
				r += dither
				g += dither
				b += dither
			}

			img.Pix[idx] = clampByte(r)
			img.Pix[idx+1] = clampByte(g)
			img.Pix[idx+2] = clampByte(b)
		}
	}
}

// buildAlphaMap returns per-pixel distance to nearest transparent neighbor.
// Uses a simple BFS-like expansion from edges. Values: 0 = edge pixel,
// higher = further from edge. Capped at 4 for efficiency.
func buildAlphaMap(img *image.RGBA) []int {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dist := make([]int, w*h)
	const maxDist = 4

	// Initialize: opaque pixels adjacent to transparent get dist 0
	for i := range dist {
		dist[i] = maxDist
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			if img.Pix[idx+3] == 0 {
				dist[y*w+x] = 0
				continue
			}
			// Check if this opaque pixel neighbors a transparent pixel
			if hasTransparentNeighbor(img, x, y, w, h, bounds) {
				dist[y*w+x] = 1
			}
		}
	}

	// Propagate distances (simple iterative approach, capped iterations)
	for d := 2; d <= maxDist; d++ {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if dist[y*w+x] != maxDist {
					continue
				}
				idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
				if img.Pix[idx+3] == 0 {
					continue
				}
				if hasDistNeighbor(dist, x, y, w, h, d-1) {
					dist[y*w+x] = d
				}
			}
		}
	}

	return dist
}

// hasTransparentNeighbor checks if pixel at (x,y) has a transparent 4-neighbor.
func hasTransparentNeighbor(img *image.RGBA, x, y, w, h int, bounds image.Rectangle) bool {
	offsets := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, off := range offsets {
		nx, ny := x+off[0], y+off[1]
		if nx < 0 || nx >= w || ny < 0 || ny >= h {
			// Image boundary counts as transparent neighbor
			return true
		}
		idx := (ny-bounds.Min.Y)*img.Stride + (nx-bounds.Min.X)*4
		if img.Pix[idx+3] == 0 {
			return true
		}
	}
	return false
}

// hasDistNeighbor checks if pixel at (x,y) has a 4-neighbor with dist value d.
func hasDistNeighbor(dist []int, x, y, w, h, d int) bool {
	offsets := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, off := range offsets {
		nx, ny := x+off[0], y+off[1]
		if nx >= 0 && nx < w && ny >= 0 && ny < h {
			if dist[ny*w+nx] == d {
				return true
			}
		}
	}
	return false
}

// ditherNoise returns deterministic per-pixel noise in range [-1.0, 1.0].
func ditherNoise(x, y int, seed uint64) float64 {
	// Simple hash-based noise
	h := seed ^ (uint64(x)*374761393 + uint64(y)*668265263)
	h = (h ^ (h >> 13)) * 1274126177
	h = h ^ (h >> 16)
	return (float64(h%1000)/500.0 - 1.0)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// ExtractRGBA converts an image.Image to *image.RGBA for pixel manipulation.
func ExtractRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
	return dst
}

// ShadingConfigForPart returns slightly adjusted shading for different body part types.
// Head gets stronger highlight, legs get stronger AO, torso gets balanced shading.
func ShadingConfigForPart(base ShadingConfig, colorRole string, zIndex int) ShadingConfig {
	cfg := base
	switch {
	case zIndex >= 15: // Head and above
		cfg.LightIntensity *= 1.3
		cfg.EdgeDarkening *= 0.8
	case zIndex >= 10: // Torso
		cfg.LightIntensity *= 1.0
		cfg.AmbientOcclusion *= 1.2
	case zIndex >= 5: // Legs
		cfg.LightIntensity *= 0.7
		cfg.AmbientOcclusion *= 1.5
		cfg.EdgeDarkening *= 1.2
	case colorRole == "shadow": // Shadow
		cfg.LightIntensity = 0
		cfg.EdgeDarkening = 0
		cfg.DitherStrength = 0
		cfg.AmbientOcclusion = 0
	}
	return cfg
}

// ApplyOverlapDarkening darkens the bottom N rows of an image to simulate
// ambient occlusion from the body part rendered above it.
func ApplyOverlapDarkening(img *image.RGBA, rows int, strength float64) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if rows > h {
		rows = h
	}
	startY := 0 // darken top rows (parts below are rendered first, occluded at top by parts above)
	for y := startY; y < startY+rows; y++ {
		factor := 1.0 - strength*(1.0-float64(y-startY)/float64(rows))
		for x := 0; x < w; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			if img.Pix[idx+3] == 0 {
				continue
			}
			img.Pix[idx] = clampByte(float64(img.Pix[idx]) * factor)
			img.Pix[idx+1] = clampByte(float64(img.Pix[idx+1]) * factor)
			img.Pix[idx+2] = clampByte(float64(img.Pix[idx+2]) * factor)
		}
	}
}

// toRGBAColor converts any color.Color to color.RGBA.
func toRGBAColor(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}
