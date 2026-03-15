// Package tiles provides Phase 16.3 parallax depth effects.
// Phase 16.3 added multi-layer parallax rendering for 3D depth perception.
// This file implements multi-layer tile rendering with parallax scrolling,
// ambient occlusion, and height-based shadows for depth perception.
package tiles

import (
	"image"
	"image/color"
	"math"
)

// TileLayer represents a rendering layer with depth information.
type TileLayer int

const (
	// LayerBackground - furthest layer (e.g., sky, distant walls)
	LayerBackground TileLayer = iota
	// LayerBase - middle layer (main tile content)
	LayerBase
	// LayerForeground - closest layer (details, decorations)
	LayerForeground
)

// String returns the string representation of a tile layer.
func (l TileLayer) String() string {
	switch l {
	case LayerBackground:
		return "background"
	case LayerBase:
		return "base"
	case LayerForeground:
		return "foreground"
	default:
		return "unknown"
	}
}

// ParallaxConfig contains parameters for parallax depth rendering.
type ParallaxConfig struct {
	// BaseConfig for the tile
	BaseConfig Config

	// Layer to render
	Layer TileLayer

	// CameraX and CameraY represent camera position for parallax offset
	CameraX float64
	CameraY float64

	// ParallaxDepth controls offset amount (0.0 = no parallax, 1.0 = full parallax)
	// Background: typically 0.2-0.4 (moves slower)
	// Base: typically 1.0 (moves with camera)
	// Foreground: typically 1.2-1.5 (moves faster for enhanced depth)
	ParallaxDepth float64

	// AOIntensity controls ambient occlusion darkness (0.0 = none, 1.0 = full)
	AOIntensity float64

	// ShadowHeight controls height-based shadow length (0.0 = none, 1.0 = full)
	ShadowHeight float64

	// ShadowAngle controls shadow direction in radians (default: math.Pi/4 for 45° down-right)
	ShadowAngle float64
}

// DefaultParallaxConfig returns a default parallax configuration.
func DefaultParallaxConfig() ParallaxConfig {
	return ParallaxConfig{
		BaseConfig:    DefaultConfig(),
		Layer:         LayerBase,
		CameraX:       0.0,
		CameraY:       0.0,
		ParallaxDepth: 1.0,
		AOIntensity:   0.5,
		ShadowHeight:  0.3,
		ShadowAngle:   math.Pi / 4, // 45° down-right
	}
}

// Validate checks if the parallax configuration is valid.
func (c ParallaxConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	if c.ParallaxDepth < 0.0 || c.ParallaxDepth > 2.0 {
		return ErrInvalidParallaxDepth
	}
	if c.AOIntensity < 0.0 || c.AOIntensity > 1.0 {
		return ErrInvalidAOIntensity
	}
	if c.ShadowHeight < 0.0 || c.ShadowHeight > 1.0 {
		return ErrInvalidShadowHeight
	}
	return nil
}

// ParallaxOffset calculates the pixel offset for parallax effect.
// Returns (offsetX, offsetY) based on camera position and layer depth.
func (c ParallaxConfig) ParallaxOffset() (float64, float64) {
	// Layer-specific parallax multipliers
	var depthMultiplier float64
	switch c.Layer {
	case LayerBackground:
		depthMultiplier = 0.3 // Background moves slower
	case LayerBase:
		depthMultiplier = 1.0 // Base moves with camera
	case LayerForeground:
		depthMultiplier = 1.4 // Foreground moves faster
	default:
		depthMultiplier = 1.0
	}

	// Apply parallax depth and multiplier
	offsetX := c.CameraX * c.ParallaxDepth * depthMultiplier
	offsetY := c.CameraY * c.ParallaxDepth * depthMultiplier

	return offsetX, offsetY
}

// GenerateWithParallax creates a tile image with parallax depth effects.
// This generates a multi-layer tile with ambient occlusion and height-based shadows.
func (g *Generator) GenerateWithParallax(config ParallaxConfig) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("invalid parallax config")
		}
		return nil, err
	}

	// Generate base tile
	baseTile, err := g.Generate(config.BaseConfig)
	if err != nil {
		return nil, err
	}

	// Create layer-specific content
	img := image.NewRGBA(baseTile.Bounds())

	// Copy base tile
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, baseTile.At(x, y))
		}
	}

	// Apply layer-specific effects
	switch config.Layer {
	case LayerBackground:
		g.applyBackgroundEffects(img, config)
	case LayerBase:
		// Base layer gets standard rendering with AO and shadows
		if config.AOIntensity > 0 {
			g.applyAmbientOcclusion(img, config)
		}
		if config.ShadowHeight > 0 {
			g.applyHeightShadows(img, config)
		}
	case LayerForeground:
		g.applyForegroundEffects(img, config)
	}

	return img, nil
}

// applyBackgroundEffects applies background layer effects (darkening, blur simulation).
func (g *Generator) applyBackgroundEffects(img *image.RGBA, config ParallaxConfig) {
	bounds := img.Bounds()

	// Darken and desaturate background layer for depth
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)

			// Darken by 30%
			r := uint8(float64(c.R) * 0.7)
			g := uint8(float64(c.G) * 0.7)
			b := uint8(float64(c.B) * 0.7)

			// Desaturate slightly
			gray := uint8((float64(r) + float64(g) + float64(b)) / 3.0)
			r = uint8(float64(r)*0.8 + float64(gray)*0.2)
			g = uint8(float64(g)*0.8 + float64(gray)*0.2)
			b = uint8(float64(b)*0.8 + float64(gray)*0.2)

			img.SetRGBA(x, y, color.RGBA{r, g, b, c.A})
		}
	}
}

// applyForegroundEffects applies foreground layer effects (slight brightening).
func (g *Generator) applyForegroundEffects(img *image.RGBA, config ParallaxConfig) {
	bounds := img.Bounds()

	// Brighten foreground layer slightly for emphasis
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)

			// Brighten by 10%
			r := uint8(math.Min(float64(c.R)*1.1, 255))
			g := uint8(math.Min(float64(c.G)*1.1, 255))
			b := uint8(math.Min(float64(c.B)*1.1, 255))

			img.SetRGBA(x, y, color.RGBA{r, g, b, c.A})
		}
	}
}

// applyAmbientOcclusion generates and applies ambient occlusion to tile corners/edges.
// AO darkens concave corners and edges for depth perception.
func (g *Generator) applyAmbientOcclusion(img *image.RGBA, config ParallaxConfig) {
	aoMap := g.computeAOMap(img, config)
	g.applyAOToImage(img, aoMap, config)
}

// computeAOMap creates an ambient occlusion map using edge detection.
func (g *Generator) computeAOMap(img *image.RGBA, config ParallaxConfig) [][]float64 {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	aoMap := make([][]float64, height)
	for y := 0; y < height; y++ {
		aoMap[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			aoMap[y][x] = g.calculatePixelOcclusion(img, x, y, width, height)
		}
	}
	return aoMap
}

// calculatePixelOcclusion computes occlusion for a single pixel by sampling neighbors.
func (g *Generator) calculatePixelOcclusion(img *image.RGBA, x, y, width, height int) float64 {
	occlusionSum := 0.0
	sampleCount := 0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			occ, samples := g.sampleNeighborOcclusion(img, x, y, dx, dy, width, height)
			occlusionSum += occ
			sampleCount += samples
		}
	}

	return occlusionSum / float64(sampleCount)
}

// sampleNeighborOcclusion samples a single neighbor for occlusion contribution.
func (g *Generator) sampleNeighborOcclusion(img *image.RGBA, x, y, dx, dy, width, height int) (float64, int) {
	nx, ny := x+dx, y+dy

	if nx < 0 || nx >= width || ny < 0 || ny >= height {
		return 1.0, 1 // Edge counts as occluded
	}

	if g.isNeighborDarker(img, x, y, nx, ny) {
		return 1.0, 1
	}
	return 0.0, 1
}

// isNeighborDarker checks if neighbor pixel is significantly darker than current.
func (g *Generator) isNeighborDarker(img *image.RGBA, x, y, nx, ny int) bool {
	c1 := img.RGBAAt(x, y)
	c2 := img.RGBAAt(nx, ny)

	brightness1 := float64(c1.R+c1.G+c1.B) / 3.0
	brightness2 := float64(c2.R+c2.G+c2.B) / 3.0

	return brightness2 < brightness1*0.7
}

// applyAOToImage applies the computed AO map to darken the image.
func (g *Generator) applyAOToImage(img *image.RGBA, aoMap [][]float64, config ParallaxConfig) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			g.darkenPixel(img, x, y, aoMap[y][x], config.AOIntensity)
		}
	}
}

// darkenPixel darkens a single pixel based on AO factor.
func (g *Generator) darkenPixel(img *image.RGBA, x, y int, aoValue, aoIntensity float64) {
	c := img.RGBAAt(x, y)
	aoFactor := 1.0 - (aoValue * aoIntensity * 0.3)

	r := uint8(float64(c.R) * aoFactor)
	gr := uint8(float64(c.G) * aoFactor)
	b := uint8(float64(c.B) * aoFactor)

	img.SetRGBA(x, y, color.RGBA{r, gr, b, c.A})
}

// applyHeightShadows generates height-based shadows for 3D depth effect.
// Shadows are cast based on tile type and height parameter.
func (g *Generator) applyHeightShadows(img *image.RGBA, config ParallaxConfig) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	shadowDX, shadowDY := calculateShadowOffset(config, width, height)
	shadowIntensity := determineShadowIntensity(config)
	shadowMap := createShadowMap(img, width, height, shadowDX, shadowDY, shadowIntensity)
	applyShadowToImage(img, shadowMap, width, height)
}

// calculateShadowOffset computes shadow offset based on angle and height.
func calculateShadowOffset(config ParallaxConfig, width, height int) (int, int) {
	shadowDX := int(math.Cos(config.ShadowAngle) * config.ShadowHeight * float64(width) * 0.3)
	shadowDY := int(math.Sin(config.ShadowAngle) * config.ShadowHeight * float64(height) * 0.3)
	return shadowDX, shadowDY
}

// determineShadowIntensity returns shadow intensity based on tile type.
func determineShadowIntensity(config ParallaxConfig) float64 {
	if config.BaseConfig.Type == TileWall ||
		config.BaseConfig.Type == TileWallNE ||
		config.BaseConfig.Type == TileWallNW ||
		config.BaseConfig.Type == TileWallSE ||
		config.BaseConfig.Type == TileWallSW {
		return 0.4
	}
	return 0.2
}

// createShadowMap generates a shadow overlay from edge detection.
func createShadowMap(img *image.RGBA, width, height, shadowDX, shadowDY int, shadowIntensity float64) [][]float64 {
	shadowMap := make([][]float64, height)
	for y := 0; y < height; y++ {
		shadowMap[y] = make([]float64, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if isEdgePixel(img, x, y, width, height) {
				castShadowFromEdge(shadowMap, x, y, shadowDX, shadowDY, width, height, shadowIntensity)
			}
		}
	}

	return shadowMap
}

// isEdgePixel detects if a pixel is an edge (bright to dark transition).
func isEdgePixel(img *image.RGBA, x, y, width, height int) bool {
	c := img.RGBAAt(x, y)
	brightness := float64(c.R+c.G+c.B) / 3.0

	if x < width-1 {
		cRight := img.RGBAAt(x+1, y)
		brightRight := float64(cRight.R+cRight.G+cRight.B) / 3.0
		if brightness > brightRight*1.3 {
			return true
		}
	}
	if y < height-1 {
		cDown := img.RGBAAt(x, y+1)
		brightDown := float64(cDown.R+cDown.G+cDown.B) / 3.0
		if brightness > brightDown*1.3 {
			return true
		}
	}
	return false
}

// castShadowFromEdge projects shadow from an edge pixel to the shadow map.
func castShadowFromEdge(shadowMap [][]float64, x, y, shadowDX, shadowDY, width, height int, shadowIntensity float64) {
	sx := x + shadowDX
	sy := y + shadowDY
	if sx >= 0 && sx < width && sy >= 0 && sy < height {
		shadowMap[sy][sx] = math.Max(shadowMap[sy][sx], shadowIntensity)
	}
}

// applyShadowToImage applies the shadow map to the image.
func applyShadowToImage(img *image.RGBA, shadowMap [][]float64, width, height int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if shadowMap[y][x] > 0 {
				c := img.RGBAAt(x, y)
				shadowFactor := 1.0 - shadowMap[y][x]
				r := uint8(float64(c.R) * shadowFactor)
				g := uint8(float64(c.G) * shadowFactor)
				b := uint8(float64(c.B) * shadowFactor)
				img.SetRGBA(x, y, color.RGBA{r, g, b, c.A})
			}
		}
	}
}

// GenerateLayeredTile generates all three layers for a tile at once.
// Returns (background, base, foreground) images.
func (g *Generator) GenerateLayeredTile(baseConfig Config, cameraX, cameraY float64) (*image.RGBA, *image.RGBA, *image.RGBA, error) {
	// Validate base config
	if err := baseConfig.Validate(); err != nil {
		return nil, nil, nil, err
	}

	// Generate background layer
	bgConfig := ParallaxConfig{
		BaseConfig:    baseConfig,
		Layer:         LayerBackground,
		CameraX:       cameraX,
		CameraY:       cameraY,
		ParallaxDepth: 0.3,
		AOIntensity:   0.3,
		ShadowHeight:  0.1,
		ShadowAngle:   math.Pi / 4,
	}
	background, err := g.GenerateWithParallax(bgConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	// Generate base layer
	baseLayerConfig := ParallaxConfig{
		BaseConfig:    baseConfig,
		Layer:         LayerBase,
		CameraX:       cameraX,
		CameraY:       cameraY,
		ParallaxDepth: 1.0,
		AOIntensity:   0.5,
		ShadowHeight:  0.3,
		ShadowAngle:   math.Pi / 4,
	}
	base, err := g.GenerateWithParallax(baseLayerConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	// Generate foreground layer (lighter, for details)
	fgConfig := ParallaxConfig{
		BaseConfig:    baseConfig,
		Layer:         LayerForeground,
		CameraX:       cameraX,
		CameraY:       cameraY,
		ParallaxDepth: 1.4,
		AOIntensity:   0.2,
		ShadowHeight:  0.5,
		ShadowAngle:   math.Pi / 4,
	}
	foreground, err := g.GenerateWithParallax(fgConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	return background, base, foreground, nil
}

// CompositeLayers composites the three layers into a single image.
// This is useful for rendering when parallax offset isn't needed.
func CompositeLayers(background, base, foreground *image.RGBA) *image.RGBA {
	bounds := base.Bounds()
	composite := image.NewRGBA(bounds)

	// Composite in order: background -> base -> foreground
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Start with background
			result := background.RGBAAt(x, y)

			// Alpha blend base layer
			baseColor := base.RGBAAt(x, y)
			result = alphaBlend(result, baseColor)

			// Alpha blend foreground layer
			fgColor := foreground.RGBAAt(x, y)
			result = alphaBlend(result, fgColor)

			composite.SetRGBA(x, y, result)
		}
	}

	return composite
}

// alphaBlend blends two colors using alpha compositing.
func alphaBlend(dst, src color.RGBA) color.RGBA {
	if src.A == 0 {
		return dst
	}
	if src.A == 255 {
		return src
	}

	// Alpha blending formula: C = (src.A * src + (1 - src.A) * dst) / 255
	alpha := float64(src.A) / 255.0
	invAlpha := 1.0 - alpha

	r := uint8(float64(src.R)*alpha + float64(dst.R)*invAlpha)
	g := uint8(float64(src.G)*alpha + float64(dst.G)*invAlpha)
	b := uint8(float64(src.B)*alpha + float64(dst.B)*invAlpha)
	a := uint8(math.Max(float64(src.A), float64(dst.A)))

	return color.RGBA{r, g, b, a}
}

// GenerateAOMap generates a standalone ambient occlusion map for a tile.
// Returns a grayscale image where darker values indicate more occlusion.
func (g *Generator) GenerateAOMap(config Config, intensity float64) (*image.RGBA, error) {
	// Generate base tile
	baseTile, err := g.Generate(config)
	if err != nil {
		return nil, err
	}

	bounds := baseTile.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	aoImage := image.NewRGBA(bounds)

	// Generate AO using same algorithm as applyAmbientOcclusion
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			occlusionSum := 0.0
			sampleCount := 0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx < 0 || nx >= width || ny < 0 || ny >= height {
						occlusionSum += 1.0
						sampleCount++
						continue
					}

					c1 := baseTile.RGBAAt(x, y)
					c2 := baseTile.RGBAAt(nx, ny)

					brightness1 := float64(c1.R+c1.G+c1.B) / 3.0
					brightness2 := float64(c2.R+c2.G+c2.B) / 3.0

					if brightness2 < brightness1*0.7 {
						occlusionSum += 1.0
					}
					sampleCount++
				}
			}

			// Convert to grayscale (darker = more occluded)
			ao := occlusionSum / float64(sampleCount)
			aoValue := uint8((1.0 - ao*intensity) * 255)

			aoImage.SetRGBA(x, y, color.RGBA{aoValue, aoValue, aoValue, 255})
		}
	}

	return aoImage, nil
}
