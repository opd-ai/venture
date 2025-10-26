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
	"github.com/sirupsen/logrus"
)

// Generator creates procedural sprites.
type Generator struct {
	paletteGen *palette.Generator
	shapeGen   *shapes.Generator
	logger     *logrus.Entry
}

// NewGenerator creates a new sprite generator.
func NewGenerator() *Generator {
	return NewGeneratorWithLogger(nil)
}

// NewGeneratorWithLogger creates a new sprite generator with a logger.
func NewGeneratorWithLogger(logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "sprite",
		})
	}
	return &Generator{
		paletteGen: palette.NewGenerator(),
		shapeGen:   shapes.NewGenerator(),
		logger:     logEntry,
	}
}

// GetPaletteGenerator returns the palette generator.
func (g *Generator) GetPaletteGenerator() *palette.Generator {
	return g.paletteGen
}

// Generate creates a sprite from the configuration.
func (g *Generator) Generate(config Config) (*ebiten.Image, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"type":       config.Type,
			"genreID":    config.GenreID,
			"seed":       config.Seed,
			"width":      config.Width,
			"height":     config.Height,
			"complexity": config.Complexity,
		}).Debug("generating sprite")
	}

	// Generate palette if not provided
	if config.Palette == nil {
		pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
		if err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Error("palette generation failed")
			}
			return nil, err
		}
		config.Palette = pal
	}

	// Create seed generator for consistent random values
	seedGen := procgen.NewSeedGenerator(config.Seed)
	rng := rand.New(rand.NewSource(seedGen.GetSeed("sprite", config.Variation)))

	// Generate sprite based on type
	var img *ebiten.Image
	var err error
	switch config.Type {
	case SpriteEntity:
		img, err = g.generateEntity(config, rng)
	case SpriteItem:
		img, err = g.generateItem(config, rng)
	case SpriteTile:
		img, err = g.generateTile(config, rng)
	case SpriteParticle:
		img, err = g.generateParticle(config, rng)
	case SpriteUI:
		img, err = g.generateUI(config, rng)
	default:
		img, err = g.generateEntity(config, rng)
	}

	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).WithField("type", config.Type).Error("sprite generation failed")
		}
		return nil, err
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"type": config.Type,
			"seed": config.Seed,
		}).Info("sprite generated")
	}

	return img, nil
}

// generateEntity creates an entity/character sprite.
func (g *Generator) generateEntity(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	// Phase 5.1: Check if we should use template-based generation
	// Use templates for complexity >= 0.3 (Tier 2+), fallback to random for low complexity
	useTemplate := config.Complexity >= 0.3

	// Check if entity type is specified in Custom config
	var entityType string
	if config.Custom != nil {
		if et, ok := config.Custom["entityType"].(string); ok {
			entityType = et
			useTemplate = true // Always use templates when entity type is specified
		}
	}

	// Use template-based generation if enabled
	if useTemplate && entityType != "" {
		return g.generateEntityWithTemplate(config, entityType, rng)
	}

	// Fallback to original random generation for simple entities or when no type specified
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

// generateEntityWithTemplate creates an entity sprite using anatomical templates (Phase 5.1 & 5.2).
func (g *Generator) generateEntityWithTemplate(config Config, entityType string, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Extract direction from config (Phase 5.2)
	direction := DirDown // Default facing down
	if config.Custom != nil {
		if dir, ok := config.Custom["facing"].(string); ok {
			direction = Direction(dir)
		}
	}

	// Extract genre from config (Phase 5.2)
	genre := ""
	if config.Custom != nil {
		if g, ok := config.Custom["genre"].(string); ok {
			genre = g
		}
	}

	// Extract equipment flags (Phase 5.2)
	hasWeapon := false
	hasShield := false
	if config.Custom != nil {
		if w, ok := config.Custom["hasWeapon"].(bool); ok {
			hasWeapon = w
		}
		if s, ok := config.Custom["hasShield"].(bool); ok {
			hasShield = s
		}
	}

	// Extract boss flag and scale (Phase 5.3)
	isBoss := false
	bossScale := 2.5 // Default boss scale
	if config.Custom != nil {
		if b, ok := config.Custom["isBoss"].(bool); ok {
			isBoss = b
		}
		if scale, ok := config.Custom["bossScale"].(float64); ok {
			bossScale = scale
		}
	}

	// Select appropriate template based on entity type, genre, direction, and equipment
	var template AnatomicalTemplate

	// Check if humanoid with equipment
	isHumanoid := false
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior":
		isHumanoid = true
	}

	if isHumanoid && (hasWeapon || hasShield) {
		// Use equipment template
		template = HumanoidWithEquipment(direction, hasWeapon, hasShield)
	} else if isHumanoid && genre != "" {
		// Use genre-specific humanoid template
		template = SelectHumanoidTemplate(genre, entityType, direction)
	} else if isHumanoid {
		// Use directional template
		template = HumanoidDirectionalTemplate(direction)
	} else {
		// Use basic template for non-humanoids
		template = SelectTemplate(entityType)
	}

	// Apply boss scaling if needed (Phase 5.3)
	if isBoss {
		template = BossTemplate(template, bossScale)
		// Optionally add boss enhancements
		if config.Complexity > 0.6 {
			template = ApplyBossEnhancements(template)
		}
	}

	// Get sorted parts for correct rendering order (Z-index)
	parts := template.GetSortedParts()

	for _, partData := range parts {
		spec := partData.Spec

		// Calculate actual dimensions and position from relative values
		partWidth := int(float64(config.Width) * spec.RelativeWidth)
		partHeight := int(float64(config.Height) * spec.RelativeHeight)

		// Skip parts with invalid dimensions
		if partWidth <= 0 || partHeight <= 0 {
			continue
		}

		// Select shape type for this part (randomly from allowed shapes)
		var shapeType shapes.ShapeType
		if len(spec.ShapeTypes) > 0 {
			shapeType = spec.ShapeTypes[rng.Intn(len(spec.ShapeTypes))]
		} else {
			shapeType = shapes.ShapeCircle // Default fallback
		}

		// Get color based on color role
		partColor := g.getColorForRole(spec.ColorRole, config.Palette)

		// Generate shape for this body part
		shapeConfig := shapes.Config{
			Type:      shapeType,
			Width:     partWidth,
			Height:    partHeight,
			Color:     partColor,
			Seed:      config.Seed + int64(spec.ZIndex),
			Smoothing: 0.2,
			Rotation:  spec.Rotation,
		}

		shape, err := g.shapeGen.Generate(shapeConfig)
		if err != nil {
			continue // Skip on error
		}

		// Position shape according to template
		opts := &ebiten.DrawImageOptions{}

		// Calculate position (relative to sprite center)
		x := float64(config.Width)*spec.RelativeX - float64(partWidth)/2
		y := float64(config.Height)*spec.RelativeY - float64(partHeight)/2
		opts.GeoM.Translate(x, y)

		// Apply opacity
		if spec.Opacity < 1.0 {
			opts.ColorScale.ScaleAlpha(float32(spec.Opacity))
		}

		img.DrawImage(shape, opts)
	}

	return img, nil
}

// getColorForRole returns the appropriate color based on the role string.
func (g *Generator) getColorForRole(role string, pal *palette.Palette) color.Color {
	switch role {
	case "primary":
		return pal.Primary
	case "secondary":
		return pal.Secondary
	case "accent1":
		return pal.Accent1
	case "accent2":
		return pal.Accent2
	case "accent3":
		return pal.Accent3
	case "highlight1":
		return pal.Highlight1
	case "highlight2":
		return pal.Highlight2
	case "shadow":
		// Return dark semi-transparent color for shadows
		return color.RGBA{R: 0, G: 0, B: 0, A: 80}
	default:
		// Default to random color from palette Colors slice
		if len(pal.Colors) > 0 {
			return pal.Colors[0]
		}
		return pal.Primary
	}
}

// generateItem creates an item sprite using item templates (Phase 5.4).
func (g *Generator) generateItem(config Config, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Extract item type and rarity from config
	var itemType ItemType
	var rarity ItemRarity = RarityCommon

	if config.Custom != nil {
		if it, ok := config.Custom["itemType"].(string); ok {
			itemType = ItemType(it)
		}
		if r, ok := config.Custom["rarity"].(int); ok {
			rarity = ItemRarity(r)
		} else if r, ok := config.Custom["rarity"].(ItemRarity); ok {
			rarity = r
		}
	}

	// Use template-based generation if item type specified
	if itemType != "" {
		return g.generateItemWithTemplate(config, itemType, rarity, rng)
	}

	// Fallback to original random generation for backward compatibility
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

// generateItemWithTemplate creates an item sprite using item templates (Phase 5.4).
func (g *Generator) generateItemWithTemplate(config Config, itemType ItemType, rarity ItemRarity, rng *rand.Rand) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Select appropriate template
	template := SelectItemTemplate(itemType, rarity)

	// Render each part
	for _, part := range template.Parts {
		// Calculate actual dimensions
		partWidth := int(float64(config.Width) * part.RelativeWidth)
		partHeight := int(float64(config.Height) * part.RelativeHeight)

		// Skip invalid parts
		if partWidth <= 0 || partHeight <= 0 {
			continue
		}

		// Select shape type
		var shapeType shapes.ShapeType
		if len(part.ShapeTypes) > 0 {
			shapeType = part.ShapeTypes[rng.Intn(len(part.ShapeTypes))]
		} else {
			shapeType = shapes.ShapeCircle
		}

		// Get color based on role
		partColor := g.getColorForRole(part.ColorRole, config.Palette)

		// Generate shape
		shapeConfig := shapes.Config{
			Type:      shapeType,
			Width:     partWidth,
			Height:    partHeight,
			Color:     partColor,
			Seed:      config.Seed + int64(part.ZIndex),
			Smoothing: 0.2,
			Rotation:  part.Rotation,
		}

		shape, err := g.shapeGen.Generate(shapeConfig)
		if err != nil {
			continue
		}

		// Position and draw
		opts := &ebiten.DrawImageOptions{}
		x := float64(config.Width)*part.RelativeX - float64(partWidth)/2
		y := float64(config.Height)*part.RelativeY - float64(partHeight)/2
		opts.GeoM.Translate(x, y)

		// Apply opacity
		if part.Opacity < 1.0 {
			opts.ColorScale.ScaleAlpha(float32(part.Opacity))
		}

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
