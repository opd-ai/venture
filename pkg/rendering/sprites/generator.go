//go:build !test
// +build !test

// Package sprites provides procedural sprite generation.
// This file implements sprite generators that create entity visuals
// at runtime without external assets.
package sprites

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// Generator creates procedural sprites.
type Generator struct {
	paletteGen *palette.Generator
	shapeGen   *shapes.Generator
}

// NewGenerator creates a new sprite generator.
func NewGenerator() *Generator {
	return &Generator{
		paletteGen: palette.NewGenerator(),
		shapeGen:   shapes.NewGenerator(),
	}
}

// GetPaletteGenerator returns the palette generator.
func (g *Generator) GetPaletteGenerator() *palette.Generator {
	return g.paletteGen
}

// Generate creates a sprite from the configuration.
func (g *Generator) Generate(config Config) (*ebiten.Image, error) {
	// Generate palette if not provided
	if config.Palette == nil {
		pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
		if err != nil {
			return nil, err
		}
		config.Palette = pal
	}

	// Create seed generator for consistent random values
	seedGen := procgen.NewSeedGenerator(config.Seed)
	rng := rand.New(rand.NewSource(seedGen.GetSeed("sprite", config.Variation)))

	// Generate sprite based on type
	switch config.Type {
	case SpriteEntity:
		return g.generateEntity(config, rng)
	case SpriteItem:
		return g.generateItem(config, rng)
	case SpriteTile:
		return g.generateTile(config, rng)
	case SpriteParticle:
		return g.generateParticle(config, rng)
	case SpriteUI:
		return g.generateUI(config, rng)
	default:
		return g.generateEntity(config, rng)
	}
}

// generateEntity creates an entity/character sprite.
func (g *Generator) generateEntity(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Determine number of shapes based on complexity
	numShapes := 1 + int(config.Complexity*4)

	// Generate body (main shape)
	bodyConfig := shapes.Config{
		Type:      shapes.ShapeType(rng.Intn(3)), // Circle, Rectangle, or Triangle
		Width:     int(float64(config.Width) * 0.7),
		Height:    int(float64(config.Height) * 0.7),
		Color:     config.Palette.Primary,
		Seed:      config.Seed,
		Smoothing: 0.2,
	}

	bodyShape, err := g.shapeGen.Generate(bodyConfig)
	if err != nil {
		return nil, err
	}

	// Draw body centered
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		float64(config.Width-bodyConfig.Width)/2,
		float64(config.Height-bodyConfig.Height)/2,
	)
	img.DrawImage(bodyShape, opts)

	// Add detail shapes based on complexity
	for i := 1; i < numShapes; i++ {
		detailConfig := shapes.Config{
			Type:      shapes.ShapeType(rng.Intn(6)),
			Width:     int(float64(config.Width) * (0.2 + rng.Float64()*0.3)),
			Height:    int(float64(config.Height) * (0.2 + rng.Float64()*0.3)),
			Color:     config.Palette.Colors[rng.Intn(len(config.Palette.Colors))],
			Seed:      config.Seed + int64(i),
			Sides:     3 + rng.Intn(5),
			Smoothing: rng.Float64() * 0.3,
		}

		detailShape, err := g.shapeGen.Generate(detailConfig)
		if err != nil {
			continue // Skip on error
		}

		// Position detail randomly
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(
			float64(rng.Intn(config.Width-detailConfig.Width)),
			float64(rng.Intn(config.Height-detailConfig.Height)),
		)
		img.DrawImage(detailShape, opts)
	}

	return img, nil
}

// generateItem creates an item sprite.
func (g *Generator) generateItem(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Items are typically simpler with 1-3 shapes
	numShapes := 1 + int(config.Complexity*2)

	for i := 0; i < numShapes; i++ {
		var colorChoice color.Color
		if i == 0 {
			colorChoice = config.Palette.Secondary
		} else {
			colorChoice = config.Palette.Colors[rng.Intn(len(config.Palette.Colors))]
		}

		itemConfig := shapes.Config{
			Type:       shapes.ShapeType(rng.Intn(6)),
			Width:      int(float64(config.Width) * (0.5 + rng.Float64()*0.4)),
			Height:     int(float64(config.Height) * (0.5 + rng.Float64()*0.4)),
			Color:      colorChoice,
			Seed:       config.Seed + int64(i),
			Sides:      4 + rng.Intn(4),
			InnerRatio: 0.3 + rng.Float64()*0.4,
			Rotation:   rng.Float64() * 360,
			Smoothing:  0.1,
		}

		shape, err := g.shapeGen.Generate(itemConfig)
		if err != nil {
			continue
		}

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(
			float64(config.Width-itemConfig.Width)/2,
			float64(config.Height-itemConfig.Height)/2,
		)
		img.DrawImage(shape, opts)
	}

	return img, nil
}

// generateTile creates a terrain tile sprite.
func (g *Generator) generateTile(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Tiles are simple filled rectangles with optional patterns
	tileConfig := shapes.Config{
		Type:      shapes.ShapeRectangle,
		Width:     config.Width,
		Height:    config.Height,
		Color:     config.Palette.Background,
		Seed:      config.Seed,
		Smoothing: 0,
	}

	tile, err := g.shapeGen.Generate(tileConfig)
	if err != nil {
		return nil, err
	}

	img.DrawImage(tile, nil)

	// Add pattern detail based on complexity
	if config.Complexity > 0.3 {
		numPatterns := int(config.Complexity * 5)
		for i := 0; i < numPatterns; i++ {
			patternConfig := shapes.Config{
				Type:      shapes.ShapeCircle,
				Width:     2 + rng.Intn(4),
				Height:    2 + rng.Intn(4),
				Color:     config.Palette.Colors[rng.Intn(len(config.Palette.Colors))],
				Seed:      config.Seed + int64(i),
				Smoothing: 0.5,
			}

			pattern, err := g.shapeGen.Generate(patternConfig)
			if err != nil {
				continue
			}

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(
				float64(rng.Intn(config.Width)),
				float64(rng.Intn(config.Height)),
			)
			opts.ColorScale.ScaleAlpha(0.3)
			img.DrawImage(pattern, opts)
		}
	}

	return img, nil
}

// generateParticle creates a particle effect sprite.
func (g *Generator) generateParticle(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Particles are small, simple shapes
	particleConfig := shapes.Config{
		Type:      shapes.ShapeType(rng.Intn(3)), // Circle, Rectangle, or Triangle
		Width:     config.Width,
		Height:    config.Height,
		Color:     config.Palette.Accent1,
		Seed:      config.Seed,
		Smoothing: 0.5,
	}

	particle, err := g.shapeGen.Generate(particleConfig)
	if err != nil {
		return nil, err
	}

	img.DrawImage(particle, nil)

	return img, nil
}

// generateUI creates a UI element sprite.
func (g *Generator) generateUI(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	// UI elements are typically rectangles with borders
	uiConfig := shapes.Config{
		Type:      shapes.ShapeRectangle,
		Width:     config.Width,
		Height:    config.Height,
		Color:     config.Palette.Background,
		Seed:      config.Seed,
		Smoothing: 0.1,
	}

	uiShape, err := g.shapeGen.Generate(uiConfig)
	if err != nil {
		return nil, err
	}

	result := ebiten.NewImage(config.Width, config.Height)
	result.DrawImage(uiShape, nil)

	// Add border
	borderConfig := shapes.Config{
		Type:       shapes.ShapeRing,
		Width:      config.Width,
		Height:     config.Height,
		Color:      config.Palette.Primary,
		Seed:       config.Seed,
		InnerRatio: 0.9,
		Smoothing:  0,
	}

	border, err := g.shapeGen.Generate(borderConfig)
	if err == nil {
		result.DrawImage(border, nil)
	}

	return result, nil
}
