package tiles

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// Generator creates procedural tile images.
type Generator struct {
	paletteGen *palette.Generator
}

// NewGenerator creates a new tile generator.
func NewGenerator() *Generator {
	return &Generator{
		paletteGen: palette.NewGenerator(),
	}
}

// Generate creates a tile image from the given configuration.
func (g *Generator) Generate(config Config) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create RNG from seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Generate color palette for genre
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	// Create base image
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Generate tile based on type
	switch config.Type {
	case TileFloor:
		g.generateFloor(img, pal, rng, config)
	case TileWall:
		g.generateWall(img, pal, rng, config)
	case TileDoor:
		g.generateDoor(img, pal, rng, config)
	case TileCorridor:
		g.generateCorridor(img, pal, rng, config)
	case TileWater:
		g.generateWater(img, pal, rng, config)
	case TileLava:
		g.generateLava(img, pal, rng, config)
	case TileTrap:
		g.generateTrap(img, pal, rng, config)
	case TileStairs:
		g.generateStairs(img, pal, rng, config)
	default:
		return nil, fmt.Errorf("unknown tile type: %d", config.Type)
	}

	return img, nil
}

// generateFloor creates a floor tile with subtle texture.
func (g *Generator) generateFloor(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	baseColor := g.pickColor(pal, "floor", rng)
	
	// Use variant to select pattern
	pattern := g.selectPattern(config.Variant, rng, []Pattern{
		PatternSolid, PatternCheckerboard, PatternDots,
	})

	switch pattern {
	case PatternSolid:
		g.fillSolid(img, baseColor, config.Variant, rng)
	case PatternCheckerboard:
		g.fillCheckerboard(img, baseColor, config.Variant, rng)
	case PatternDots:
		g.fillDots(img, baseColor, config.Variant, rng)
	}
}

// generateWall creates a wall tile with brick or stone pattern.
func (g *Generator) generateWall(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	baseColor := g.pickColor(pal, "wall", rng)
	
	// Walls typically use brick or solid pattern
	pattern := g.selectPattern(config.Variant, rng, []Pattern{
		PatternSolid, PatternBrick, PatternLines,
	})

	switch pattern {
	case PatternSolid:
		g.fillSolid(img, baseColor, config.Variant, rng)
	case PatternBrick:
		g.fillBrick(img, baseColor, config.Variant, rng)
	case PatternLines:
		g.fillLines(img, baseColor, config.Variant, rng)
	}
}

// generateDoor creates a door tile.
func (g *Generator) generateDoor(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Doors use grain pattern to simulate wood
	baseColor := g.pickColor(pal, "door", rng)
	g.fillGrain(img, baseColor, config.Variant, rng)
	
	// Add door frame
	frameColor := g.darkenColor(baseColor, 0.3)
	g.drawFrame(img, frameColor, 2)
}

// generateCorridor creates a corridor tile (similar to floor but darker).
func (g *Generator) generateCorridor(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	baseColor := g.pickColor(pal, "floor", rng)
	baseColor = g.darkenColor(baseColor, 0.2)
	g.fillSolid(img, baseColor, config.Variant, rng)
}

// generateWater creates a water tile.
func (g *Generator) generateWater(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Water uses blue tones
	baseColor := color.RGBA{R: 30, G: 100, B: 180, A: 255}
	g.fillSolid(img, baseColor, config.Variant, rng)
}

// generateLava creates a lava tile.
func (g *Generator) generateLava(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Lava uses red/orange tones
	baseColor := color.RGBA{R: 200, G: 50, B: 0, A: 255}
	g.fillSolid(img, baseColor, config.Variant, rng)
}

// generateTrap creates a trap tile.
func (g *Generator) generateTrap(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Traps look like floors with a danger indicator
	baseColor := g.pickColor(pal, "floor", rng)
	g.fillSolid(img, baseColor, config.Variant, rng)
	
	// Add danger pattern in center
	dangerColor := g.pickColor(pal, "danger", rng)
	centerX := config.Width / 2
	centerY := config.Height / 2
	radius := config.Width / 4
	g.drawCircle(img, centerX, centerY, radius, dangerColor)
}

// generateStairs creates a stairs tile.
func (g *Generator) generateStairs(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config Config) {
	baseColor := g.pickColor(pal, "floor", rng)
	g.fillSolid(img, baseColor, 0.1, rng)
	
	// Draw stairs as horizontal lines
	stepColor := g.darkenColor(baseColor, 0.3)
	stepCount := 5
	stepHeight := config.Height / stepCount
	
	for i := 0; i < stepCount; i++ {
		y := i * stepHeight
		for x := 0; x < config.Width; x++ {
			img.Set(x, y, stepColor)
			if y+1 < config.Height {
				img.Set(x, y+1, stepColor)
			}
		}
	}
}

// Helper methods for pattern generation

func (g *Generator) fillSolid(img *image.RGBA, baseColor color.Color, variance float64, rng *rand.Rand) {
	bounds := img.Bounds()
	r, gr, b, a := baseColor.RGBA()
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Add slight random variation
			variation := 1.0 + (rng.Float64()*2.0-1.0)*variance*0.1
			varR := uint8(math.Min(255, float64(r>>8)*variation))
			varG := uint8(math.Min(255, float64(gr>>8)*variation))
			varB := uint8(math.Min(255, float64(b>>8)*variation))
			
			img.Set(x, y, color.RGBA{R: varR, G: varG, B: varB, A: uint8(a >> 8)})
		}
	}
}

func (g *Generator) fillCheckerboard(img *image.RGBA, baseColor color.Color, variance float64, rng *rand.Rand) {
	bounds := img.Bounds()
	altColor := g.lightenColor(baseColor, 0.1)
	checkSize := 4
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			useAlt := ((x / checkSize) + (y / checkSize)) % 2 == 0
			if useAlt {
				img.Set(x, y, altColor)
			} else {
				img.Set(x, y, baseColor)
			}
		}
	}
}

func (g *Generator) fillDots(img *image.RGBA, baseColor color.Color, variance float64, rng *rand.Rand) {
	g.fillSolid(img, baseColor, variance, rng)
	
	dotColor := g.darkenColor(baseColor, 0.2)
	spacing := 6
	radius := 1
	
	bounds := img.Bounds()
	for y := bounds.Min.Y + spacing/2; y < bounds.Max.Y; y += spacing {
		for x := bounds.Min.X + spacing/2; x < bounds.Max.X; x += spacing {
			g.drawCircle(img, x, y, radius, dotColor)
		}
	}
}

func (g *Generator) fillLines(img *image.RGBA, baseColor color.Color, variance float64, rng *rand.Rand) {
	g.fillSolid(img, baseColor, variance, rng)
	
	lineColor := g.darkenColor(baseColor, 0.15)
	spacing := 4
	
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += spacing {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, lineColor)
		}
	}
}

func (g *Generator) fillBrick(img *image.RGBA, baseColor color.Color, variance float64, rng *rand.Rand) {
	g.fillSolid(img, baseColor, variance, rng)
	
	mortarColor := g.darkenColor(baseColor, 0.3)
	brickWidth := 16
	brickHeight := 8
	
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += brickHeight {
		offset := 0
		if (y/brickHeight)%2 == 1 {
			offset = brickWidth / 2
		}
		
		// Horizontal mortar line
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, mortarColor)
		}
		
		// Vertical mortar lines
		for x := bounds.Min.X + offset; x < bounds.Max.X; x += brickWidth {
			for dy := 0; dy < brickHeight && y+dy < bounds.Max.Y; dy++ {
				img.Set(x, y+dy, mortarColor)
			}
		}
	}
}

func (g *Generator) fillGrain(img *image.RGBA, baseColor color.Color, variance float64, rng *rand.Rand) {
	bounds := img.Bounds()
	r, gr, b, a := baseColor.RGBA()
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Create horizontal grain lines with noise
		grainIntensity := math.Sin(float64(y)*0.3) * 0.1
		
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			variation := 1.0 + grainIntensity + (rng.Float64()*2.0-1.0)*variance*0.05
			varR := uint8(math.Min(255, float64(r>>8)*variation))
			varG := uint8(math.Min(255, float64(gr>>8)*variation))
			varB := uint8(math.Min(255, float64(b>>8)*variation))
			
			img.Set(x, y, color.RGBA{R: varR, G: varG, B: varB, A: uint8(a >> 8)})
		}
	}
}

func (g *Generator) drawFrame(img *image.RGBA, frameColor color.Color, thickness int) {
	bounds := img.Bounds()
	
	for t := 0; t < thickness; t++ {
		// Top and bottom
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, bounds.Min.Y+t, frameColor)
			img.Set(x, bounds.Max.Y-t-1, frameColor)
		}
		
		// Left and right
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(bounds.Min.X+t, y, frameColor)
			img.Set(bounds.Max.X-t-1, y, frameColor)
		}
	}
}

func (g *Generator) drawCircle(img *image.RGBA, cx, cy, radius int, col color.Color) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				bounds := img.Bounds()
				if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
					img.Set(x, y, col)
				}
			}
		}
	}
}

// Helper methods for color manipulation

func (g *Generator) pickColor(pal *palette.Palette, context string, rng *rand.Rand) color.Color {
	// Select appropriate color based on context
	switch context {
	case "wall":
		return pal.Background
	case "floor":
		return pal.Colors[rng.Intn(len(pal.Colors)/2)]
	case "door":
		return pal.Colors[rng.Intn(len(pal.Colors)/2)+len(pal.Colors)/2]
	case "danger":
		return pal.Danger
	default:
		return pal.Primary
	}
}

func (g *Generator) darkenColor(col color.Color, amount float64) color.Color {
	r, gr, b, a := col.RGBA()
	factor := 1.0 - amount
	return color.RGBA{
		R: uint8(float64(r>>8) * factor),
		G: uint8(float64(gr>>8) * factor),
		B: uint8(float64(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

func (g *Generator) lightenColor(col color.Color, amount float64) color.Color {
	r, gr, b, a := col.RGBA()
	factor := 1.0 + amount
	return color.RGBA{
		R: uint8(math.Min(255, float64(r>>8)*factor)),
		G: uint8(math.Min(255, float64(gr>>8)*factor)),
		B: uint8(math.Min(255, float64(b>>8)*factor)),
		A: uint8(a >> 8),
	}
}

func (g *Generator) selectPattern(variant float64, rng *rand.Rand, patterns []Pattern) Pattern {
	if len(patterns) == 0 {
		return PatternSolid
	}
	
	// Use variant to influence selection
	idx := int(variant * float64(len(patterns)))
	if idx >= len(patterns) {
		idx = len(patterns) - 1
	}
	
	return patterns[idx]
}

// Validate implements the procgen.Generator interface.
func (g *Generator) Validate(result interface{}) error {
	img, ok := result.(*image.RGBA)
	if !ok {
		return fmt.Errorf("result is not an *image.RGBA")
	}
	
	if img == nil {
		return fmt.Errorf("generated image is nil")
	}
	
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return fmt.Errorf("generated image has zero dimensions")
	}
	
	return nil
}
