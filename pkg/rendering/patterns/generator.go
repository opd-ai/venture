// Package patterns provides procedural texture pattern generation.
// This file implements the Generator struct and its texture generation methods.
// Code relocated from: original generator.go (types moved to types.go)
package patterns

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// Generator creates procedural texture patterns.
type Generator struct {
	logger *logrus.Entry
}

// NewGenerator creates a new pattern generator.
func NewGenerator() *Generator {
	return NewGeneratorWithLogger(nil)
}

// NewGeneratorWithLogger creates a new pattern generator with a logger.
func NewGeneratorWithLogger(logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "patterns",
		})
	}
	return &Generator{
		logger: logEntry,
	}
}

// Generate creates a texture pattern image from the configuration.
func (g *Generator) Generate(config TextureConfig) (*image.RGBA, error) {
	if config.Width <= 0 || config.Height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", config.Width, config.Height)
	}

	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"texture": config.Texture.String(),
			"genre":   config.GenreID,
			"size":    fmt.Sprintf("%dx%d", config.Width, config.Height),
			"seed":    config.Seed,
		}).Debug("generating texture pattern")
	}

	// Create base image
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Generate RNG from seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Apply genre-specific variations
	config = g.applyGenreVariations(config, rng)

	// Generate base texture
	switch config.Texture {
	case TextureStone:
		g.generateStoneTexture(img, config, rng)
	case TextureWood:
		g.generateWoodTexture(img, config, rng)
	case TextureMetal:
		g.generateMetalTexture(img, config, rng)
	case TextureOrganic:
		g.generateOrganicTexture(img, config, rng)
	default:
		return nil, fmt.Errorf("unknown texture type: %d", config.Texture)
	}

	// Add detail layers based on detail level
	if config.DetailLevel > 0 {
		g.addDetailLayer(img, config, rng)
	}

	// Add variation to prevent repetition
	g.addVariation(img, config, rng)

	// Add normal map approximation for depth
	g.addDepthEffect(img, config, rng)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"texture": config.Texture.String(),
			"genre":   config.GenreID,
		}).Info("texture pattern generated")
	}

	return img, nil
}

// applyGenreVariations modifies config.Scale and config.DetailLevel in-place
// based on the specified genre, then returns the modified config.
func (g *Generator) applyGenreVariations(config TextureConfig, rng *rand.Rand) TextureConfig {
	switch config.GenreID {
	case "fantasy":
		// Fantasy: organic, earthy, natural variations
		config.Scale = config.Scale * (0.8 + rng.Float64()*0.4) // 0.8-1.2x
		config.DetailLevel = config.DetailLevel * 1.1           // More detail
	case "scifi":
		// Sci-fi: geometric, clean, metallic
		config.Scale = config.Scale * (1.0 + rng.Float64()*0.3) // 1.0-1.3x
		config.DetailLevel = config.DetailLevel * 0.9           // Less organic detail
	case "horror":
		// Horror: distorted, irregular, unsettling
		config.Scale = config.Scale * (0.6 + rng.Float64()*0.6) // 0.6-1.2x
		config.DetailLevel = config.DetailLevel * 1.2           // More chaotic detail
	case "cyberpunk":
		// Cyberpunk: angular, tech-augmented, neon accents
		config.Scale = config.Scale * (0.9 + rng.Float64()*0.4) // 0.9-1.3x
		config.DetailLevel = config.DetailLevel * 1.0           // Standard detail
	case "postapocalyptic":
		// Post-apocalyptic: weathered, damaged, decayed
		config.Scale = config.Scale * (0.7 + rng.Float64()*0.5) // 0.7-1.2x
		config.DetailLevel = config.DetailLevel * 1.15          // More weathering detail
	}
	return config
}

// generateStoneTexture creates a stone/rock texture pattern.
func (g *Generator) generateStoneTexture(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	// Extract colors
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Multi-octave Perlin-like noise for stone texture
			noise := g.perlinNoise(float64(x)*config.Scale, float64(y)*config.Scale, rng)
			noise += 0.5 * g.perlinNoise(float64(x)*config.Scale*2, float64(y)*config.Scale*2, rng)
			noise += 0.25 * g.perlinNoise(float64(x)*config.Scale*4, float64(y)*config.Scale*4, rng)
			noise /= 1.75 // Normalize

			// Clamp noise to [0, 1]
			if noise < 0 {
				noise = 0
			}
			if noise > 1 {
				noise = 1
			}

			// Blend between colors based on noise
			r := uint8((float64(r1>>8)*(1-noise) + float64(r2>>8)*noise))
			g := uint8((float64(g1>>8)*(1-noise) + float64(g2>>8)*noise))
			b := uint8((float64(b1>>8)*(1-noise) + float64(b2>>8)*noise))
			a := uint8((float64(a1>>8)*(1-noise) + float64(a2>>8)*noise))

			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

// generateWoodTexture creates a wood grain texture pattern.
func (g *Generator) generateWoodTexture(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	// Extract colors
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	// Wood grain parameters
	grainFrequency := 0.1 * config.Scale
	grainTurbulence := 5.0

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Distance from center for radial pattern
			centerX := float64(config.Width) / 2.0
			centerY := float64(config.Height) / 2.0
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Add turbulence to create wood grain
			turbulence := g.perlinNoise(float64(x)*0.1, float64(y)*0.1, rng) * grainTurbulence
			grainValue := math.Sin((dist+turbulence)*grainFrequency)*0.5 + 0.5

			// Add fine grain detail
			fineGrain := g.perlinNoise(float64(x)*0.5, float64(y)*0.5, rng) * 0.2
			grainValue = grainValue*0.8 + fineGrain

			// Clamp to [0, 1]
			if grainValue < 0 {
				grainValue = 0
			}
			if grainValue > 1 {
				grainValue = 1
			}

			// Blend between colors
			r := uint8((float64(r1>>8)*(1-grainValue) + float64(r2>>8)*grainValue))
			g := uint8((float64(g1>>8)*(1-grainValue) + float64(g2>>8)*grainValue))
			b := uint8((float64(b1>>8)*(1-grainValue) + float64(b2>>8)*grainValue))
			a := uint8((float64(a1>>8)*(1-grainValue) + float64(a2>>8)*grainValue))

			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

// generateMetalTexture creates a metallic texture pattern.
func (g *Generator) generateMetalTexture(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	// Extract colors
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Fine scratches and brushed metal effect
			horizontalNoise := g.perlinNoise(float64(x)*config.Scale*0.1, float64(y)*config.Scale*2.0, rng)
			verticalNoise := g.perlinNoise(float64(x)*config.Scale*2.0, float64(y)*config.Scale*0.1, rng)

			// Combine for anisotropic brushed metal look
			noise := (horizontalNoise + verticalNoise) * 0.5

			// Add small specular highlights
			highlight := g.perlinNoise(float64(x)*config.Scale*4, float64(y)*config.Scale*4, rng)
			if highlight > 0.7 {
				noise += (highlight - 0.7) * 0.5
			}

			// Clamp to [0, 1]
			if noise < 0 {
				noise = 0
			}
			if noise > 1 {
				noise = 1
			}

			// Blend between colors
			r := uint8((float64(r1>>8)*(1-noise) + float64(r2>>8)*noise))
			g := uint8((float64(g1>>8)*(1-noise) + float64(g2>>8)*noise))
			b := uint8((float64(b1>>8)*(1-noise) + float64(b2>>8)*noise))
			a := uint8((float64(a1>>8)*(1-noise) + float64(a2>>8)*noise))

			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

// generateOrganicTexture creates an organic/biological texture pattern.
func (g *Generator) generateOrganicTexture(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	// Extract colors
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Multiple octaves for organic complexity
			noise := g.perlinNoise(float64(x)*config.Scale*0.5, float64(y)*config.Scale*0.5, rng) * 0.5
			noise += g.perlinNoise(float64(x)*config.Scale*1.0, float64(y)*config.Scale*1.0, rng) * 0.3
			noise += g.perlinNoise(float64(x)*config.Scale*2.0, float64(y)*config.Scale*2.0, rng) * 0.15
			noise += g.perlinNoise(float64(x)*config.Scale*4.0, float64(y)*config.Scale*4.0, rng) * 0.05

			// Add cellular pattern for biological appearance
			cellValue := g.cellularNoise(float64(x)*config.Scale, float64(y)*config.Scale, rng)
			noise = noise*0.7 + cellValue*0.3

			// Clamp to [0, 1]
			if noise < 0 {
				noise = 0
			}
			if noise > 1 {
				noise = 1
			}

			// Blend between colors
			r := uint8((float64(r1>>8)*(1-noise) + float64(r2>>8)*noise))
			g := uint8((float64(g1>>8)*(1-noise) + float64(g2>>8)*noise))
			b := uint8((float64(b1>>8)*(1-noise) + float64(b2>>8)*noise))
			a := uint8((float64(a1>>8)*(1-noise) + float64(a2>>8)*noise))

			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

// addDetailLayer adds fine detail to the texture based on detail level.
func (g *Generator) addDetailLayer(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	if config.DetailLevel <= 0 {
		return
	}

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			g.applyDetailToPixel(img, x, y, config, rng)
		}
	}
}

// applyDetailToPixel applies high-frequency noise detail to a single pixel.
func (g *Generator) applyDetailToPixel(img *image.RGBA, x, y int, config TextureConfig, rng *rand.Rand) {
	detail := g.calculatePixelDetail(x, y, config, rng)
	c := img.RGBAAt(x, y)
	r, gr, b := g.applyDetailToChannels(c, detail)
	img.SetRGBA(x, y, color.RGBA{R: uint8(r), G: uint8(gr), B: uint8(b), A: c.A})
}

// calculatePixelDetail computes the detail value for a pixel using high-frequency noise.
func (g *Generator) calculatePixelDetail(x, y int, config TextureConfig, rng *rand.Rand) float64 {
	detail := g.perlinNoise(float64(x)*config.Scale*8, float64(y)*config.Scale*8, rng)
	return detail * config.DetailLevel * 0.15
}

// applyDetailToChannels applies detail to RGB channels and clamps values to 0-255.
func (g *Generator) applyDetailToChannels(c color.RGBA, detail float64) (float64, float64, float64) {
	r := g.clampColorValue(float64(c.R) + detail*255)
	gr := g.clampColorValue(float64(c.G) + detail*255)
	b := g.clampColorValue(float64(c.B) + detail*255)
	return r, gr, b
}

// addVariation adds per-pixel variation to prevent repetitive patterns.
func (g *Generator) addVariation(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			g.applyPixelVariation(img, x, y, rng)
		}
	}
}

// applyPixelVariation applies random color variation to a single pixel.
func (g *Generator) applyPixelVariation(img *image.RGBA, x, y int, rng *rand.Rand) {
	variation := (rng.Float64() - 0.5) * 0.08 // ±4% variation
	c := img.RGBAAt(x, y)

	r := g.clampColorValue(float64(c.R) * (1 + variation))
	green := g.clampColorValue(float64(c.G) * (1 + variation))
	b := g.clampColorValue(float64(c.B) * (1 + variation))

	img.SetRGBA(x, y, color.RGBA{R: uint8(r), G: uint8(green), B: uint8(b), A: c.A})
}

// addDepthEffect adds normal map approximation for depth perception.
func (g *Generator) addDepthEffect(img *image.RGBA, config TextureConfig, rng *rand.Rand) {
	temp := g.copyImageForGradient(img, config)
	g.applyLightingEffect(img, temp, config)
}

// copyImageForGradient creates a temporary copy for gradient calculation.
func (g *Generator) copyImageForGradient(img *image.RGBA, config TextureConfig) *image.RGBA {
	temp := image.NewRGBA(img.Bounds())
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			temp.Set(x, y, img.At(x, y))
		}
	}
	return temp
}

// applyLightingEffect applies subtle lighting based on luminance gradients.
func (g *Generator) applyLightingEffect(img, temp *image.RGBA, config TextureConfig) {
	for y := 1; y < config.Height-1; y++ {
		for x := 1; x < config.Width-1; x++ {
			lightEffect := g.calculateLightingAtPixel(temp, x, y)
			g.applyLightToPixel(img, x, y, lightEffect)
		}
	}
}

// calculateLightingAtPixel calculates lighting effect for a single pixel.
func (g *Generator) calculateLightingAtPixel(temp *image.RGBA, x, y int) float64 {
	leftLum := g.luminance(temp.RGBAAt(x-1, y))
	rightLum := g.luminance(temp.RGBAAt(x+1, y))
	topLum := g.luminance(temp.RGBAAt(x, y-1))
	bottomLum := g.luminance(temp.RGBAAt(x, y+1))

	gradX := rightLum - leftLum
	gradY := bottomLum - topLum

	lightDir := -gradX + -gradY
	return lightDir * 0.15
}

// applyLightToPixel applies lighting effect to a pixel with clamping.
func (g *Generator) applyLightToPixel(img *image.RGBA, x, y int, lightEffect float64) {
	c := img.RGBAAt(x, y)
	r := g.clampColorValue(float64(c.R) + lightEffect*255)
	green := g.clampColorValue(float64(c.G) + lightEffect*255)
	b := g.clampColorValue(float64(c.B) + lightEffect*255)

	img.SetRGBA(x, y, color.RGBA{R: uint8(r), G: uint8(green), B: uint8(b), A: c.A})
}

// clampColorValue clamps a color value to the range [0, 255].
func (g *Generator) clampColorValue(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}

// perlinNoise generates Perlin-like noise at given coordinates.
// This is a simplified Perlin noise implementation for deterministic texture generation.
func (g *Generator) perlinNoise(x, y float64, rng *rand.Rand) float64 {
	// Grid coordinates
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	x1 := x0 + 1
	y1 := y0 + 1

	// Interpolation weights
	sx := x - x0
	sy := y - y0

	// Gradient vectors at corners (deterministic based on grid position)
	n00 := g.dotGridGradient(int(x0), int(y0), x, y, rng)
	n10 := g.dotGridGradient(int(x1), int(y0), x, y, rng)
	n01 := g.dotGridGradient(int(x0), int(y1), x, y, rng)
	n11 := g.dotGridGradient(int(x1), int(y1), x, y, rng)

	// Interpolate using smoothstep
	sx = g.smoothstep(sx)
	sy = g.smoothstep(sy)

	n0 := n00*(1-sx) + n10*sx
	n1 := n01*(1-sx) + n11*sx

	return n0*(1-sy) + n1*sy
}

// dotGridGradient computes the dot product of distance and gradient vectors.
func (g *Generator) dotGridGradient(ix, iy int, x, y float64, rng *rand.Rand) float64 {
	// Create deterministic gradient based on grid position
	// Use simple hash function for deterministic pseudo-random gradient
	hash := (ix*73856093 ^ iy*19349663) & 0x7FFFFFFF
	angle := float64(hash%360) * math.Pi / 180.0

	// Gradient vector
	gx := math.Cos(angle)
	gy := math.Sin(angle)

	// Distance vector
	dx := x - float64(ix)
	dy := y - float64(iy)

	// Dot product
	return dx*gx + dy*gy
}

// cellularNoise generates cellular/Worley noise for organic patterns.
// Optimized version using hash-based pseudo-random instead of creating RNG instances.
func (g *Generator) cellularNoise(x, y float64, rng *rand.Rand) float64 {
	// Grid cell
	cellX := math.Floor(x)
	cellY := math.Floor(y)

	minDist := math.MaxFloat64

	// Check neighboring cells
	for dy := -1.0; dy <= 1.0; dy++ {
		for dx := -1.0; dx <= 1.0; dx++ {
			// Cell position
			cx := cellX + dx
			cy := cellY + dy

			// Feature point in cell (deterministic based on cell position)
			// Use simple hash for pseudo-random position within cell
			hash := int((cx*73856093 + cy*19349663)) & 0x7FFFFFFF
			// Convert hash to [0,1] range using bit manipulation
			hashX := float64((hash>>8)&0xFFFF) / 65536.0
			hashY := float64((hash>>16)&0xFFFF) / 65536.0

			px := cx + hashX
			py := cy + hashY

			// Distance to feature point
			dx := x - px
			dy := y - py
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < minDist {
				minDist = dist
			}
		}
	}

	// Normalize distance (typical range is 0-1.5)
	return math.Min(minDist/1.5, 1.0)
}

// smoothstep interpolation function for smoother noise.
func (g *Generator) smoothstep(t float64) float64 {
	return t * t * (3 - 2*t)
}

// luminance calculates the perceived brightness of a color.
func (g *Generator) luminance(c color.RGBA) float64 {
	// Standard luminance formula
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}

// GeneratePattern creates a basic pattern image from the configuration.
// This method generates primitive patterns (stripes, dots, gradients, noise,
// checkerboard, circles) as opposed to material textures from Generate().
func (g *Generator) GeneratePattern(config Config) (*image.RGBA, error) {
	if config.Width <= 0 || config.Height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", config.Width, config.Height)
	}

	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"pattern": config.Type.String(),
			"size":    fmt.Sprintf("%dx%d", config.Width, config.Height),
			"seed":    config.Seed,
		}).Debug("generating pattern")
	}

	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
	rng := rand.New(rand.NewSource(config.Seed))

	switch config.Type {
	case PatternStripes:
		g.generateStripesPattern(img, config, rng)
	case PatternDots:
		g.generateDotsPattern(img, config, rng)
	case PatternGradient:
		g.generateGradientPattern(img, config)
	case PatternNoise:
		g.generateNoisePattern(img, config, rng)
	case PatternCheckerboard:
		g.generateCheckerboardPattern(img, config)
	case PatternCircles:
		g.generateCirclesPattern(img, config)
	default:
		return nil, fmt.Errorf("unknown pattern type: %d", config.Type)
	}

	// Apply opacity if less than 1.0
	if config.Opacity < 1.0 && config.Opacity > 0 {
		g.applyOpacity(img, config.Opacity)
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"pattern": config.Type.String(),
		}).Info("pattern generated")
	}

	return img, nil
}

// generateStripesPattern creates parallel line stripes.
func (g *Generator) generateStripesPattern(img *image.RGBA, config Config, rng *rand.Rand) {
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	angleRad := config.Angle * math.Pi / 180.0
	cosA, sinA := math.Cos(angleRad), math.Sin(angleRad)

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Rotate coordinates by angle
			rotX := float64(x)*cosA - float64(y)*sinA

			// Calculate stripe position
			stripePos := rotX * config.Frequency / float64(config.Width)
			stripeValue := math.Sin(stripePos*2*math.Pi)*0.5 + 0.5

			// Blend colors
			r := uint8(float64(r1>>8)*(1-stripeValue) + float64(r2>>8)*stripeValue)
			gr := uint8(float64(g1>>8)*(1-stripeValue) + float64(g2>>8)*stripeValue)
			b := uint8(float64(b1>>8)*(1-stripeValue) + float64(b2>>8)*stripeValue)
			a := uint8(float64(a1>>8)*(1-stripeValue) + float64(a2>>8)*stripeValue)

			img.SetRGBA(x, y, color.RGBA{R: r, G: gr, B: b, A: a})
		}
	}
}

// generateDotsPattern creates a dot grid pattern.
func (g *Generator) generateDotsPattern(img *image.RGBA, config Config, rng *rand.Rand) {
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	spacing := float64(config.Width) / config.Frequency
	if spacing < 1 {
		spacing = 1
	}
	dotRadius := spacing * config.Amplitude * 0.4

	// Fill with background color
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(r2 >> 8), G: uint8(g2 >> 8), B: uint8(b2 >> 8), A: uint8(a2 >> 8),
			})
		}
	}

	// Draw dots
	for cy := spacing / 2; cy < float64(config.Height); cy += spacing {
		for cx := spacing / 2; cx < float64(config.Width); cx += spacing {
			for y := int(cy - dotRadius); y <= int(cy+dotRadius); y++ {
				for x := int(cx - dotRadius); x <= int(cx+dotRadius); x++ {
					if x < 0 || x >= config.Width || y < 0 || y >= config.Height {
						continue
					}
					dx, dy := float64(x)-cx, float64(y)-cy
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist <= dotRadius {
						img.SetRGBA(x, y, color.RGBA{
							R: uint8(r1 >> 8), G: uint8(g1 >> 8), B: uint8(b1 >> 8), A: uint8(a1 >> 8),
						})
					}
				}
			}
		}
	}
}

// generateGradientPattern creates a smooth color gradient.
func (g *Generator) generateGradientPattern(img *image.RGBA, config Config) {
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	angleRad := config.Angle * math.Pi / 180.0
	cosA, sinA := math.Cos(angleRad), math.Sin(angleRad)

	// Calculate gradient direction extents
	maxDist := float64(config.Width)*math.Abs(cosA) + float64(config.Height)*math.Abs(sinA)
	if maxDist == 0 {
		maxDist = 1
	}

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Project point onto gradient direction
			proj := float64(x)*cosA + float64(y)*sinA
			t := proj / maxDist
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}

			r := uint8(float64(r1>>8)*(1-t) + float64(r2>>8)*t)
			gr := uint8(float64(g1>>8)*(1-t) + float64(g2>>8)*t)
			b := uint8(float64(b1>>8)*(1-t) + float64(b2>>8)*t)
			a := uint8(float64(a1>>8)*(1-t) + float64(a2>>8)*t)

			img.SetRGBA(x, y, color.RGBA{R: r, G: gr, B: b, A: a})
		}
	}
}

// generateNoisePattern creates a Perlin noise pattern.
func (g *Generator) generateNoisePattern(img *image.RGBA, config Config, rng *rand.Rand) {
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	scale := config.Frequency / 10.0
	if scale <= 0 {
		scale = 0.1
	}

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			noise := g.perlinNoise(float64(x)*scale, float64(y)*scale, rng)
			noise = noise*config.Amplitude + 0.5
			if noise < 0 {
				noise = 0
			}
			if noise > 1 {
				noise = 1
			}

			r := uint8(float64(r1>>8)*(1-noise) + float64(r2>>8)*noise)
			gr := uint8(float64(g1>>8)*(1-noise) + float64(g2>>8)*noise)
			b := uint8(float64(b1>>8)*(1-noise) + float64(b2>>8)*noise)
			a := uint8(float64(a1>>8)*(1-noise) + float64(a2>>8)*noise)

			img.SetRGBA(x, y, color.RGBA{R: r, G: gr, B: b, A: a})
		}
	}
}

// generateCheckerboardPattern creates a checkerboard pattern.
func (g *Generator) generateCheckerboardPattern(img *image.RGBA, config Config) {
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	squareSize := float64(config.Width) / config.Frequency
	if squareSize < 1 {
		squareSize = 1
	}

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			// Calculate which square this pixel is in
			sx := int(float64(x) / squareSize)
			sy := int(float64(y) / squareSize)

			// Alternate colors based on checkerboard pattern
			if (sx+sy)%2 == 0 {
				img.SetRGBA(x, y, color.RGBA{
					R: uint8(r1 >> 8), G: uint8(g1 >> 8), B: uint8(b1 >> 8), A: uint8(a1 >> 8),
				})
			} else {
				img.SetRGBA(x, y, color.RGBA{
					R: uint8(r2 >> 8), G: uint8(g2 >> 8), B: uint8(b2 >> 8), A: uint8(a2 >> 8),
				})
			}
		}
	}
}

// generateCirclesPattern creates concentric circles pattern.
func (g *Generator) generateCirclesPattern(img *image.RGBA, config Config) {
	r1, g1, b1, a1 := config.Color1.RGBA()
	r2, g2, b2, a2 := config.Color2.RGBA()

	centerX := float64(config.Width) / 2.0
	centerY := float64(config.Height) / 2.0
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Create concentric rings
			ringValue := math.Sin(dist*config.Frequency*2*math.Pi/maxDist)*0.5 + 0.5
			ringValue *= config.Amplitude
			ringValue = ringValue*0.5 + 0.25 // Normalize to 0.25-0.75 range

			r := uint8(float64(r1>>8)*(1-ringValue) + float64(r2>>8)*ringValue)
			gr := uint8(float64(g1>>8)*(1-ringValue) + float64(g2>>8)*ringValue)
			b := uint8(float64(b1>>8)*(1-ringValue) + float64(b2>>8)*ringValue)
			a := uint8(float64(a1>>8)*(1-ringValue) + float64(a2>>8)*ringValue)

			img.SetRGBA(x, y, color.RGBA{R: r, G: gr, B: b, A: a})
		}
	}
}

// applyOpacity reduces alpha channel by given opacity factor.
func (g *Generator) applyOpacity(img *image.RGBA, opacity float64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			c.A = uint8(float64(c.A) * opacity)
			img.SetRGBA(x, y, c)
		}
	}
}

// Validate checks if the generated texture meets quality requirements.
func (g *Generator) Validate(img *image.RGBA) error {
	if err := g.validateImageBasics(img); err != nil {
		return err
	}

	bounds := img.Bounds()
	avgR, avgG, avgB, err := g.calculateAverageColor(img, bounds)
	if err != nil {
		return err
	}

	if err := g.checkColorVariation(img, bounds, avgR, avgG, avgB); err != nil {
		return err
	}

	return nil
}

// validateImageBasics checks basic image validity (nil check and dimensions).
func (g *Generator) validateImageBasics(img *image.RGBA) error {
	if img == nil {
		return fmt.Errorf("texture image is nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("texture has invalid dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	return nil
}

// calculateAverageColor computes the average color of sampled pixels.
func (g *Generator) calculateAverageColor(img *image.RGBA, bounds image.Rectangle) (uint8, uint8, uint8, error) {
	var sumR, sumG, sumB, sumA uint32
	sampleCount := 0
	step := bounds.Dy() / 4
	if step == 0 {
		step = 1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		stepX := bounds.Dx() / 4
		if stepX == 0 {
			stepX = 1
		}
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			c := img.RGBAAt(x, y)
			sumR += uint32(c.R)
			sumG += uint32(c.G)
			sumB += uint32(c.B)
			sumA += uint32(c.A)
			sampleCount++
		}
	}

	if sampleCount == 0 {
		return 0, 0, 0, fmt.Errorf("no pixels sampled for validation")
	}

	avgR := uint8(sumR / uint32(sampleCount))
	avgG := uint8(sumG / uint32(sampleCount))
	avgB := uint8(sumB / uint32(sampleCount))

	return avgR, avgG, avgB, nil
}

// checkColorVariation validates that the image has sufficient color variation.
func (g *Generator) checkColorVariation(img *image.RGBA, bounds image.Rectangle, avgR, avgG, avgB uint8) error {
	step := bounds.Dy() / 4
	if step == 0 {
		step = 1
	}

	allSame := true
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		stepX := bounds.Dx() / 4
		if stepX == 0 {
			stepX = 1
		}
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			c := img.RGBAAt(x, y)
			if math.Abs(float64(c.R)-float64(avgR)) > float64(avgR)*0.05 ||
				math.Abs(float64(c.G)-float64(avgG)) > float64(avgG)*0.05 ||
				math.Abs(float64(c.B)-float64(avgB)) > float64(avgB)*0.05 {
				allSame = false
				break
			}
		}
		if !allSame {
			break
		}
	}

	if allSame {
		return fmt.Errorf("texture lacks variation (appears to be solid color)")
	}

	return nil
}
