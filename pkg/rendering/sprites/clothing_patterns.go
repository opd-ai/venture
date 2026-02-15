// Package sprites provides seed-based clothing pattern rendering for procedurally
// generated entity sprites. Patterns (stripes, checks, dots, borders, herringbone,
// diamond lattice) are applied to torso, arm, and leg body parts during sprite
// generation, transforming flat-colored garments into textured, visually distinct
// clothing that makes every NPC and player immediately distinguishable.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// ClothingPatternType identifies the visual pattern applied to a garment.
type ClothingPatternType int

const (
	// PatternNone applies no pattern (solid color).
	PatternNone ClothingPatternType = iota
	// PatternHStripes draws horizontal stripes.
	PatternHStripes
	// PatternVStripes draws vertical stripes.
	PatternVStripes
	// PatternCheckerboard draws a checkerboard grid.
	PatternCheckerboard
	// PatternDots draws a polka-dot pattern.
	PatternDots
	// PatternBorder draws a contrasting trim border around the part edge.
	PatternBorder
	// PatternHerringbone draws a V-shaped herringbone weave.
	PatternHerringbone
	// PatternDiamondLattice draws a diagonal diamond grid.
	PatternDiamondLattice
	// PatternGradientV draws a vertical color gradient (light top to dark bottom).
	PatternGradientV
	// patternCount is the number of pattern types (used for random selection).
	patternCount
)

// String returns a human-readable name for the pattern type.
func (p ClothingPatternType) String() string {
	switch p {
	case PatternNone:
		return "none"
	case PatternHStripes:
		return "horizontal_stripes"
	case PatternVStripes:
		return "vertical_stripes"
	case PatternCheckerboard:
		return "checkerboard"
	case PatternDots:
		return "dots"
	case PatternBorder:
		return "border"
	case PatternHerringbone:
		return "herringbone"
	case PatternDiamondLattice:
		return "diamond_lattice"
	case PatternGradientV:
		return "gradient_v"
	default:
		return "unknown"
	}
}

// ClothingPattern holds parameters for a single garment pattern layer.
type ClothingPattern struct {
	// Type is the pattern to render.
	Type ClothingPatternType
	// PatternColor is the contrasting color used for pattern strokes.
	PatternColor color.RGBA
	// Scale controls pattern density (1=default, <1=denser, >1=sparser).
	Scale float64
	// Intensity controls how strongly the pattern shows (0.0-1.0).
	Intensity float64
}

// ClothingPatternSet holds per-region patterns derived from an entity seed.
type ClothingPatternSet struct {
	// TorsoPattern is applied to the main body/chest area.
	TorsoPattern ClothingPattern
	// ArmPattern is applied to arm regions.
	ArmPattern ClothingPattern
	// LegPattern is applied to leg/pants regions.
	LegPattern ClothingPattern
}

// GenerateClothingPatternSet deterministically produces a pattern set from a seed.
// Roughly 30% of entities get no pattern (solid clothing), ensuring variety.
func GenerateClothingPatternSet(seed int64) ClothingPatternSet {
	rng := rand.New(rand.NewSource(seed ^ 0x434C4F5448)) // "CLOTH" XOR

	set := ClothingPatternSet{}

	// Torso has the highest chance of having a pattern
	if rng.Float64() < 0.70 {
		set.TorsoPattern = generateOnePattern(rng)
	}
	// Arms match torso 60% of the time, else get their own or none
	if rng.Float64() < 0.60 {
		set.ArmPattern = set.TorsoPattern
	} else if rng.Float64() < 0.40 {
		set.ArmPattern = generateOnePattern(rng)
	}
	// Legs get an independent pattern 45% of the time
	if rng.Float64() < 0.45 {
		set.LegPattern = generateOnePattern(rng)
	}

	return set
}

// generateOnePattern picks a random pattern type with randomised parameters.
func generateOnePattern(rng *rand.Rand) ClothingPattern {
	// Skip PatternNone (index 0) since caller already decided to add a pattern
	ptype := ClothingPatternType(1 + rng.Intn(int(patternCount)-1))

	// Pattern color: pick a random hue with moderate saturation
	hue := rng.Float64() * 360
	sat := 0.15 + rng.Float64()*0.35
	lum := 0.30 + rng.Float64()*0.40
	pcolor := hslToRGBA(hue, sat, lum)

	scale := 0.7 + rng.Float64()*0.6     // 0.7–1.3
	intensity := 0.3 + rng.Float64()*0.5 // 0.3–0.8

	return ClothingPattern{
		Type:         ptype,
		PatternColor: pcolor,
		Scale:        scale,
		Intensity:    intensity,
	}
}

// ApplyClothingPattern overlays a clothing pattern onto an already-rendered body
// part RGBA image. Only non-transparent pixels in the source image are affected,
// preserving the body part's silhouette. The pattern blends with existing colors
// at the specified intensity.
func ApplyClothingPattern(img *image.RGBA, pattern ClothingPattern, seed int64) {
	if pattern.Type == PatternNone || pattern.Intensity <= 0 {
		return
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return
	}

	intensity := pattern.Intensity
	pc := pattern.PatternColor

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			if idx+3 >= len(img.Pix) {
				continue
			}
			a := img.Pix[idx+3]
			if a == 0 {
				continue // skip transparent
			}

			patternValue := evaluatePattern(pattern.Type, x, y, w, h, pattern.Scale, seed)
			if patternValue <= 0 {
				continue
			}

			blend := intensity * patternValue
			if blend > 1.0 {
				blend = 1.0
			}

			// Blend pattern color into existing pixel
			invBlend := 1.0 - blend
			img.Pix[idx+0] = uint8(float64(img.Pix[idx+0])*invBlend + float64(pc.R)*blend)
			img.Pix[idx+1] = uint8(float64(img.Pix[idx+1])*invBlend + float64(pc.G)*blend)
			img.Pix[idx+2] = uint8(float64(img.Pix[idx+2])*invBlend + float64(pc.B)*blend)
		}
	}
}

// evaluatePattern returns 0.0-1.0 for how strongly the pattern is present at (x,y).
func evaluatePattern(ptype ClothingPatternType, x, y, w, h int, scale float64, seed int64) float64 {
	if scale <= 0 {
		scale = 1.0
	}
	// Base stripe/grid period in pixels, scaled
	period := math.Max(2.0, 3.0*scale)

	switch ptype {
	case PatternHStripes:
		return horizontalStripe(y, period)
	case PatternVStripes:
		return verticalStripe(x, period)
	case PatternCheckerboard:
		return checkerboard(x, y, period)
	case PatternDots:
		return dots(x, y, period)
	case PatternBorder:
		return border(x, y, w, h)
	case PatternHerringbone:
		return herringbone(x, y, period)
	case PatternDiamondLattice:
		return diamondLattice(x, y, period)
	case PatternGradientV:
		return gradientV(y, h)
	default:
		return 0
	}
}

func horizontalStripe(y int, period float64) float64 {
	pos := math.Mod(float64(y), period*2)
	if pos < period {
		return 1.0
	}
	return 0
}

func verticalStripe(x int, period float64) float64 {
	pos := math.Mod(float64(x), period*2)
	if pos < period {
		return 1.0
	}
	return 0
}

func checkerboard(x, y int, period float64) float64 {
	cx := int(math.Floor(float64(x) / period))
	cy := int(math.Floor(float64(y) / period))
	if (cx+cy)%2 == 0 {
		return 1.0
	}
	return 0
}

func dots(x, y int, period float64) float64 {
	p := period * 1.5
	// Center of nearest dot
	cx := math.Floor(float64(x)/p)*p + p/2
	cy := math.Floor(float64(y)/p)*p + p/2
	dx := float64(x) - cx
	dy := float64(y) - cy
	dist := math.Sqrt(dx*dx + dy*dy)
	radius := p * 0.3
	if dist <= radius {
		return 1.0 - dist/radius // soft falloff
	}
	return 0
}

func border(x, y, w, h int) float64 {
	borderWidth := 1
	if w > 8 || h > 8 {
		borderWidth = 2
	}
	if x < borderWidth || x >= w-borderWidth || y < borderWidth || y >= h-borderWidth {
		return 1.0
	}
	return 0
}

func herringbone(x, y int, period float64) float64 {
	p := period * 2
	fy := math.Mod(float64(y), p)
	var offset float64
	if fy < p/2 {
		offset = fy // ascending diagonal
	} else {
		offset = p - fy // descending diagonal
	}
	pos := math.Mod(float64(x)+offset, p)
	if pos < period*0.5 {
		return 1.0
	}
	return 0
}

func diamondLattice(x, y int, period float64) float64 {
	p := period * 2
	// Diamond pattern: |x mod p - p/2| + |y mod p - p/2| near p/2
	mx := math.Abs(math.Mod(float64(x), p) - p/2)
	my := math.Abs(math.Mod(float64(y), p) - p/2)
	dist := mx + my
	threshold := p * 0.45
	if dist > threshold && dist < threshold+1.5 {
		return 1.0
	}
	return 0
}

func gradientV(y, h int) float64 {
	if h <= 1 {
		return 0
	}
	return float64(y) / float64(h-1)
}

// PatternForBodyRegion returns the appropriate pattern from the set based on
// the color role and z-index of the body part being rendered.
func (s *ClothingPatternSet) PatternForBodyRegion(colorRole string, zIndex int) ClothingPattern {
	switch colorRole {
	case "primary":
		if zIndex <= 5 {
			return s.LegPattern
		}
		return s.TorsoPattern
	case "accent1":
		return s.ArmPattern
	default:
		return ClothingPattern{Type: PatternNone}
	}
}
