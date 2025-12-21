// Package tiles provides Phase 47 advanced wall rendering.
// This file implements anti-aliased wall rendering with seamless corner blending,
// edge smoothing, and shadow integration for enhanced 1920x1080 resolution.
package tiles

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// CornerType represents the type of wall corner junction.
type CornerType int

const (
	// CornerNone indicates no corner (standalone wall segment)
	CornerNone CornerType = iota
	// CornerL indicates an L-shaped corner (90° junction)
	CornerL
	// CornerT indicates a T-shaped junction (3-way)
	CornerT
	// CornerCross indicates a cross junction (4-way)
	CornerCross
)

// String returns the string representation of a corner type.
func (c CornerType) String() string {
	switch c {
	case CornerNone:
		return "none"
	case CornerL:
		return "L"
	case CornerT:
		return "T"
	case CornerCross:
		return "cross"
	default:
		return "unknown"
	}
}

// WallNeighbors indicates which neighboring tiles are also walls.
// Used for corner detection and blending.
type WallNeighbors struct {
	North bool
	South bool
	East  bool
	West  bool
}

// DetectCornerType determines the corner type from neighboring walls.
func (wn WallNeighbors) DetectCornerType() CornerType {
	count := 0
	if wn.North {
		count++
	}
	if wn.South {
		count++
	}
	if wn.East {
		count++
	}
	if wn.West {
		count++
	}

	switch count {
	case 4:
		return CornerCross
	case 3:
		return CornerT
	case 2:
		// L corner if adjacent sides, corridor if opposite sides
		if (wn.North && wn.East) || (wn.East && wn.South) || (wn.South && wn.West) || (wn.West && wn.North) {
			return CornerL
		}
		return CornerNone
	default:
		return CornerNone
	}
}

// EnhancedWallConfig extends Config with wall-specific rendering options.
type EnhancedWallConfig struct {
	Config
	// Neighbors indicates which adjacent tiles are walls
	Neighbors WallNeighbors
	// EnableAntialiasing enables 2x2 super-sampling
	EnableAntialiasing bool
	// EnableShadows enables shadow gradient rendering
	EnableShadows bool
	// EnableHeightEdges enables wall height edge indicators for top-down view
	EnableHeightEdges bool
	// BlendRadius is the pixel radius for corner blending (default 4)
	BlendRadius int
}

// DefaultEnhancedWallConfig returns default enhanced wall configuration.
func DefaultEnhancedWallConfig() EnhancedWallConfig {
	return EnhancedWallConfig{
		Config:             DefaultConfig(),
		Neighbors:          WallNeighbors{},
		EnableAntialiasing: true,
		EnableShadows:      true,
		EnableHeightEdges:  true,
		BlendRadius:        4,
	}
}

// GenerateEnhancedWall creates an anti-aliased wall tile with corner blending.
func (g *Generator) GenerateEnhancedWall(config EnhancedWallConfig) (*image.RGBA, error) {
	if err := config.Config.Validate(); err != nil {
		return nil, err
	}

	// Create RNG from seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Generate color palette
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, err
	}

	var img *image.RGBA

	// Apply anti-aliasing if enabled
	if config.EnableAntialiasing {
		// Render at 2x resolution
		hiresWidth := config.Width * 2
		hiresHeight := config.Height * 2
		hiresImg := image.NewRGBA(image.Rect(0, 0, hiresWidth, hiresHeight))

		// Render high-res wall
		g.renderWallContent(hiresImg, pal, rng, config, hiresWidth, hiresHeight)

		// Downsample to target resolution
		img = g.downsample2x2(hiresImg, config.Width, config.Height)
	} else {
		// Render at target resolution
		img = image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
		g.renderWallContent(img, pal, rng, config, config.Width, config.Height)
	}

	// Apply corner blending if neighbors detected
	cornerType := config.Neighbors.DetectCornerType()
	if cornerType != CornerNone {
		g.blendCorner(img, pal, rng, config, cornerType)
	}

	// Apply wall/floor boundary blending
	g.blendWallFloorBoundary(img, pal, rng, config)

	// Apply shadow gradient if enabled
	if config.EnableShadows {
		g.applyWallShadow(img, pal, config)
	}

	// Apply wall height edge indicators for top-down 3D effect
	if config.EnableHeightEdges {
		g.applyWallHeightEdges(img, pal, config)
	}

	return img, nil
}

// renderWallContent renders the wall pattern at the specified resolution.
func (g *Generator) renderWallContent(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config EnhancedWallConfig, width, height int) {
	baseColor := g.pickColor(pal, "wall", rng)

	// Select wall pattern
	pattern := g.selectPattern(config.Variant, rng, []Pattern{
		PatternSolid, PatternBrick, PatternLines,
	})

	// Create temporary config for rendering
	tempConfig := Config{
		Type:    TileWall,
		Width:   width,
		Height:  height,
		GenreID: config.GenreID,
		Seed:    config.Seed,
		Variant: config.Variant,
		Custom:  config.Custom,
	}

	switch pattern {
	case PatternSolid:
		g.fillSolid(img, baseColor, config.Variant, rng)
	case PatternBrick:
		g.fillBrick(img, baseColor, config.Variant, rng)
	case PatternLines:
		g.fillLines(img, baseColor, config.Variant, rng)
	}

	// Add depth texture
	g.addWallDepthTexture(img, baseColor, rng, tempConfig)
}

// addWallDepthTexture adds subtle depth variations to wall surfaces.
func (g *Generator) addWallDepthTexture(img *image.RGBA, baseColor color.Color, rng *rand.Rand, config Config) {
	bounds := img.Bounds()

	// Add perlin-like noise for depth
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Pseudo-perlin noise using sine waves
			noise := math.Sin(float64(x)*0.1) * math.Sin(float64(y)*0.1) * 0.05
			noise += (rng.Float64()*2.0 - 1.0) * 0.02

			existing := img.At(x, y)
			er, eg, eb, ea := existing.RGBA()

			// Blend noise with existing color
			factor := 1.0 + noise
			newR := uint8(math.Min(255, float64(er>>8)*factor))
			newG := uint8(math.Min(255, float64(eg>>8)*factor))
			newB := uint8(math.Min(255, float64(eb>>8)*factor))

			img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(ea >> 8)})
		}
	}
}

// downsample2x2 performs 2x2 super-sampling downsampling for anti-aliasing.
func (g *Generator) downsample2x2(src *image.RGBA, targetWidth, targetHeight int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			// Sample 2x2 block from source
			srcX := x * 2
			srcY := y * 2

			// Average the 4 pixels
			var rSum, gSum, bSum, aSum uint32
			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					r, g, b, a := src.At(srcX+dx, srcY+dy).RGBA()
					rSum += r
					gSum += g
					bSum += b
					aSum += a
				}
			}

			// Average and convert back to 8-bit
			dst.Set(x, y, color.RGBA{
				R: uint8(rSum / 4 >> 8),
				G: uint8(gSum / 4 >> 8),
				B: uint8(bSum / 4 >> 8),
				A: uint8(aSum / 4 >> 8),
			})
		}
	}

	return dst
}

// blendCorner applies 4px blending radius at corner junctions.
func (g *Generator) blendCorner(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config EnhancedWallConfig, cornerType CornerType) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	radius := config.BlendRadius

	// Get blend color (slightly lighter than wall)
	baseColor := g.pickColor(pal, "wall", rng)
	blendColor := g.lightenColor(baseColor, 0.1)

	switch cornerType {
	case CornerL:
		// Blend at corners where two walls meet
		g.blendLCorner(img, blendColor, radius, config.Neighbors, width, height)
	case CornerT:
		// Blend at T-junctions
		g.blendTJunction(img, blendColor, radius, config.Neighbors, width, height)
	case CornerCross:
		// Blend at cross junctions
		g.blendCrossJunction(img, blendColor, radius, width, height)
	}
}

// blendLCorner blends L-shaped corner junctions.
func (g *Generator) blendLCorner(img *image.RGBA, blendColor color.Color, radius int, neighbors WallNeighbors, width, height int) {
	bounds := img.Bounds()

	// Determine which corner to blend
	var cornerX, cornerY int
	if neighbors.North && neighbors.East {
		cornerX = bounds.Min.X + width - 1
		cornerY = bounds.Min.Y
	} else if neighbors.East && neighbors.South {
		cornerX = bounds.Min.X + width - 1
		cornerY = bounds.Min.Y + height - 1
	} else if neighbors.South && neighbors.West {
		cornerX = bounds.Min.X
		cornerY = bounds.Min.Y + height - 1
	} else if neighbors.West && neighbors.North {
		cornerX = bounds.Min.X
		cornerY = bounds.Min.Y
	} else {
		return
	}

	// Blend in circular area around corner
	g.blendCircularArea(img, cornerX, cornerY, radius, blendColor)
}

// blendTJunction blends T-shaped junctions.
func (g *Generator) blendTJunction(img *image.RGBA, blendColor color.Color, radius int, neighbors WallNeighbors, width, height int) {
	bounds := img.Bounds()
	centerX := bounds.Min.X + width/2
	centerY := bounds.Min.Y + height/2

	// Blend at center and along the T stem
	g.blendCircularArea(img, centerX, centerY, radius, blendColor)

	// Additional blending along connecting edges
	if !neighbors.North {
		// T pointing up
		for y := bounds.Min.Y; y < bounds.Min.Y+radius && y < bounds.Max.Y; y++ {
			g.blendCircularArea(img, centerX, y, radius/2, blendColor)
		}
	} else if !neighbors.South {
		// T pointing down
		for y := bounds.Max.Y - radius; y < bounds.Max.Y; y++ {
			g.blendCircularArea(img, centerX, y, radius/2, blendColor)
		}
	} else if !neighbors.East {
		// T pointing right
		for x := bounds.Max.X - radius; x < bounds.Max.X; x++ {
			g.blendCircularArea(img, x, centerY, radius/2, blendColor)
		}
	} else if !neighbors.West {
		// T pointing left
		for x := bounds.Min.X; x < bounds.Min.X+radius && x < bounds.Max.X; x++ {
			g.blendCircularArea(img, x, centerY, radius/2, blendColor)
		}
	}
}

// blendCrossJunction blends cross (4-way) junctions.
func (g *Generator) blendCrossJunction(img *image.RGBA, blendColor color.Color, radius, width, height int) {
	bounds := img.Bounds()
	centerX := bounds.Min.X + width/2
	centerY := bounds.Min.Y + height/2

	// Blend at center with larger radius
	g.blendCircularArea(img, centerX, centerY, radius, blendColor)
}

// blendCircularArea blends colors in a circular area.
func (g *Generator) blendCircularArea(img *image.RGBA, cx, cy, radius int, blendColor color.Color) {
	bounds := img.Bounds()
	br, bg, bb, _ := blendColor.RGBA()

	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			// Check bounds
			if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
				continue
			}

			// Check if within circle
			dx := x - cx
			dy := y - cy
			distSq := dx*dx + dy*dy
			if distSq > radius*radius {
				continue
			}

			// Calculate blend factor based on distance (smoother at edges)
			dist := math.Sqrt(float64(distSq))
			blendFactor := 1.0 - (dist / float64(radius))
			blendFactor = math.Max(0, math.Min(1, blendFactor))

			// Get existing pixel
			existing := img.At(x, y)
			er, eg, eb, ea := existing.RGBA()

			// Blend colors
			newR := uint8((float64(er>>8)*(1.0-blendFactor) + float64(br>>8)*blendFactor))
			newG := uint8((float64(eg>>8)*(1.0-blendFactor) + float64(bg>>8)*blendFactor))
			newB := uint8((float64(eb>>8)*(1.0-blendFactor) + float64(bb>>8)*blendFactor))

			img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(ea >> 8)})
		}
	}
}

// blendWallFloorBoundary applies 50/50 color blending at wall/floor boundaries.
func (g *Generator) blendWallFloorBoundary(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config EnhancedWallConfig) {
	bounds := img.Bounds()
	floorColor := g.pickColor(pal, "floor", rng)
	fr, fg, fb, _ := floorColor.RGBA()
	thickness := 1

	for t := 0; t < thickness; t++ {
		if !config.Neighbors.South {
			g.blendBottomEdge(img, bounds, t, fr, fg, fb)
		}
		if !config.Neighbors.North {
			g.blendTopEdge(img, bounds, t, fr, fg, fb)
		}
		if !config.Neighbors.West {
			g.blendLeftEdge(img, bounds, t, fr, fg, fb)
		}
		if !config.Neighbors.East {
			g.blendRightEdge(img, bounds, t, fr, fg, fb)
		}
	}
}

// blendBottomEdge blends the bottom edge with floor color.
func (g *Generator) blendBottomEdge(img *image.RGBA, bounds image.Rectangle, t int, fr, fg, fb uint32) {
	y := bounds.Max.Y - t - 1
	if y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		g.blendPixel(img, x, y, fr, fg, fb)
	}
}

// blendTopEdge blends the top edge with floor color.
func (g *Generator) blendTopEdge(img *image.RGBA, bounds image.Rectangle, t int, fr, fg, fb uint32) {
	y := bounds.Min.Y + t
	if y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		g.blendPixel(img, x, y, fr, fg, fb)
	}
}

// blendLeftEdge blends the left edge with floor color.
func (g *Generator) blendLeftEdge(img *image.RGBA, bounds image.Rectangle, t int, fr, fg, fb uint32) {
	x := bounds.Min.X + t
	if x < bounds.Min.X || x >= bounds.Max.X {
		return
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		g.blendPixel(img, x, y, fr, fg, fb)
	}
}

// blendRightEdge blends the right edge with floor color.
func (g *Generator) blendRightEdge(img *image.RGBA, bounds image.Rectangle, t int, fr, fg, fb uint32) {
	x := bounds.Max.X - t - 1
	if x < bounds.Min.X || x >= bounds.Max.X {
		return
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		g.blendPixel(img, x, y, fr, fg, fb)
	}
}

// blendPixel performs 50/50 color blending on a single pixel.
func (g *Generator) blendPixel(img *image.RGBA, x, y int, fr, fg, fb uint32) {
	existing := img.At(x, y)
	er, eg, eb, ea := existing.RGBA()
	newR := uint8((er>>8)/2 + (fr>>8)/2)
	newG := uint8((eg>>8)/2 + (fg>>8)/2)
	newB := uint8((eb>>8)/2 + (fb>>8)/2)
	img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(ea >> 8)})
}

// applyWallShadow applies directional shadow gradient to walls.
func (g *Generator) applyWallShadow(img *image.RGBA, pal *palette.Palette, config EnhancedWallConfig) {
	bounds := img.Bounds()
	height := bounds.Dy()

	// Apply vertical gradient (lighter at top, darker at bottom)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Shadow intensity increases towards bottom
		shadowFactor := float64(y-bounds.Min.Y) / float64(height) * 0.25

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			existing := img.At(x, y)
			r, g, b, a := existing.RGBA()

			// Darken based on shadow factor
			darkFactor := 1.0 - shadowFactor
			newR := uint8(float64(r>>8) * darkFactor)
			newG := uint8(float64(g>>8) * darkFactor)
			newB := uint8(float64(b>>8) * darkFactor)

			img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
		}
	}
}

// applyWallHeightEdges adds visible edge indicators for wall height in top-down view.
// Creates a darker top edge (player-facing, casting shadow) and lighter bottom edge
// (visible top surface reflecting light) to simulate 3D wall projection.
func (g *Generator) applyWallHeightEdges(img *image.RGBA, pal *palette.Palette, config EnhancedWallConfig) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Edge thickness scales with tile size (4px for 64x64 tiles)
	edgeThickness := height / 16
	if edgeThickness < 2 {
		edgeThickness = 2
	}

	// Apply top edge (darker - represents the player-facing wall face casting shadow)
	// This is visible when looking down at the wall from above
	for t := 0; t < edgeThickness; t++ {
		y := bounds.Min.Y + t
		if y >= bounds.Max.Y {
			continue
		}

		// Darken factor decreases as we move down from the very top
		darkenFactor := 0.6 + 0.4*float64(t)/float64(edgeThickness)

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Skip corners if they're adjacent to walls (they're handled by corner blending)
			if t < edgeThickness/2 {
				if (config.Neighbors.West && x < bounds.Min.X+edgeThickness) ||
					(config.Neighbors.East && x >= bounds.Max.X-edgeThickness) {
					continue
				}
			}

			existing := img.At(x, y)
			r, gr, b, a := existing.RGBA()

			newR := uint8(float64(r>>8) * darkenFactor)
			newG := uint8(float64(gr>>8) * darkenFactor)
			newB := uint8(float64(b>>8) * darkenFactor)

			img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
		}
	}

	// Apply bottom edge (lighter - represents the top surface of the wall visible from above)
	// This simulates light hitting the top of the wall
	for t := 0; t < edgeThickness; t++ {
		y := bounds.Max.Y - 1 - t
		if y < bounds.Min.Y {
			continue
		}

		// Lighten factor decreases as we move up from the very bottom
		lightenFactor := 1.15 + 0.1*float64(edgeThickness-1-t)/float64(edgeThickness)

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Skip corners if they're adjacent to walls
			if t < edgeThickness/2 {
				if (config.Neighbors.West && x < bounds.Min.X+edgeThickness) ||
					(config.Neighbors.East && x >= bounds.Max.X-edgeThickness) {
					continue
				}
			}

			existing := img.At(x, y)
			r, gr, b, a := existing.RGBA()

			newR := uint8(math.Min(255, float64(r>>8)*lightenFactor))
			newG := uint8(math.Min(255, float64(gr>>8)*lightenFactor))
			newB := uint8(math.Min(255, float64(b>>8)*lightenFactor))

			img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
		}
	}

	// Apply left and right edge shading for complete 3D effect
	g.applyWallSideEdges(img, config, edgeThickness, width, height)
}

// applyWallSideEdges adds left/right edge shading for walls.
func (g *Generator) applyWallSideEdges(img *image.RGBA, config EnhancedWallConfig, edgeThickness, width, height int) {
	bounds := img.Bounds()

	// Apply left edge (slightly darker - light coming from right)
	if !config.Neighbors.West {
		for t := 0; t < edgeThickness; t++ {
			x := bounds.Min.X + t
			if x >= bounds.Max.X {
				continue
			}

			darkenFactor := 0.8 + 0.2*float64(t)/float64(edgeThickness)

			for y := bounds.Min.Y + edgeThickness; y < bounds.Max.Y-edgeThickness; y++ {
				existing := img.At(x, y)
				r, gr, b, a := existing.RGBA()

				newR := uint8(float64(r>>8) * darkenFactor)
				newG := uint8(float64(gr>>8) * darkenFactor)
				newB := uint8(float64(b>>8) * darkenFactor)

				img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
			}
		}
	}

	// Apply right edge (slightly lighter - facing the light)
	if !config.Neighbors.East {
		for t := 0; t < edgeThickness; t++ {
			x := bounds.Max.X - 1 - t
			if x < bounds.Min.X {
				continue
			}

			lightenFactor := 1.1 + 0.1*float64(edgeThickness-1-t)/float64(edgeThickness)

			for y := bounds.Min.Y + edgeThickness; y < bounds.Max.Y-edgeThickness; y++ {
				existing := img.At(x, y)
				r, gr, b, a := existing.RGBA()

				newR := uint8(math.Min(255, float64(r>>8)*lightenFactor))
				newG := uint8(math.Min(255, float64(gr>>8)*lightenFactor))
				newB := uint8(math.Min(255, float64(b>>8)*lightenFactor))

				img.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
			}
		}
	}
}
