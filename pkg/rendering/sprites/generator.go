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
		var pal *palette.Palette
		var err error
		if config.PaletteOptions != nil {
			// Use advanced palette options (Phase 5.4)
			pal, err = g.paletteGen.GenerateWithOptions(config.GenreID, config.Seed, *config.PaletteOptions)
		} else {
			// Use default palette generation
			pal, err = g.paletteGen.Generate(config.GenreID, config.Seed)
		}
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

	direction, genre := extractDirectionAndGenre(config)
	hasWeapon, hasShield := extractEquipmentFlags(config)
	isBoss, bossScale := extractBossConfig(config)
	useAerial := extractAerialFlag(config)

	template := selectEntityTemplate(entityType, genre, direction, hasWeapon, hasShield, useAerial)

	if isBoss {
		template = applyBossModifications(template, bossScale, config.Complexity)
	}

	g.renderTemplateParts(img, template, config, rng)

	return img, nil
}

// extractDirectionAndGenre extracts direction and genre from the config.
func extractDirectionAndGenre(config Config) (Direction, string) {
	direction := DirDown // Default facing down
	if config.Custom != nil {
		if dir, ok := config.Custom["facing"].(string); ok {
			direction = Direction(dir)
		}
	}

	genre := ""
	if config.Custom != nil {
		if g, ok := config.Custom["genre"].(string); ok {
			genre = g
		}
	}

	return direction, genre
}

// extractEquipmentFlags extracts equipment flags from the config.
func extractEquipmentFlags(config Config) (bool, bool) {
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
	return hasWeapon, hasShield
}

// extractBossConfig extracts boss configuration from the config.
func extractBossConfig(config Config) (bool, float64) {
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
	return isBoss, bossScale
}

// extractAerialFlag extracts the aerial flag from the config.
func extractAerialFlag(config Config) bool {
	if config.Custom != nil {
		if aerial, ok := config.Custom["useAerial"].(bool); ok {
			return aerial
		}
	}
	return false
}

// selectEntityTemplate selects the appropriate anatomical template based on entity type and configuration.
func selectEntityTemplate(entityType, genre string, direction Direction, hasWeapon, hasShield, useAerial bool) AnatomicalTemplate {
	isHumanoid := isHumanoidType(entityType)

	if useAerial && isHumanoid {
		return SelectAerialTemplate(entityType, genre, direction)
	} else if isHumanoid && (hasWeapon || hasShield) {
		return HumanoidWithEquipment(direction, hasWeapon, hasShield)
	} else if isHumanoid && genre != "" {
		return SelectHumanoidTemplate(genre, entityType, direction)
	} else if isHumanoid {
		return HumanoidDirectionalTemplate(direction)
	}
	return SelectTemplate(entityType)
}

// isHumanoidType checks if the entity type is humanoid.
func isHumanoidType(entityType string) bool {
	switch entityType {
	case "humanoid", "player", "npc", "knight", "mage", "warrior":
		return true
	}
	return false
}

// applyBossModifications applies boss scaling and enhancements to the template.
func applyBossModifications(template AnatomicalTemplate, bossScale, complexity float64) AnatomicalTemplate {
	template = BossTemplate(template, bossScale)
	if complexity > 0.6 {
		template = ApplyBossEnhancements(template)
	}
	return template
}

// renderTemplateParts renders all parts of the anatomical template to the image.
func (g *Generator) renderTemplateParts(img *ebiten.Image, template AnatomicalTemplate, config Config, rng *rand.Rand) {
	parts := template.GetSortedParts()

	for _, partData := range parts {
		g.renderTemplatePart(img, partData.Spec, config, rng)
	}
}

// renderTemplatePart renders a single template part to the image.
func (g *Generator) renderTemplatePart(img *ebiten.Image, spec PartSpec, config Config, rng *rand.Rand) {
	partWidth := int(float64(config.Width) * spec.RelativeWidth)
	partHeight := int(float64(config.Height) * spec.RelativeHeight)

	if partWidth <= 0 || partHeight <= 0 {
		return
	}

	shapeType := selectShapeType(spec.ShapeTypes, rng)
	partColor := g.getColorForRole(spec.ColorRole, config.Palette)

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
		return
	}

	opts := &ebiten.DrawImageOptions{}
	x := float64(config.Width)*spec.RelativeX - float64(partWidth)/2
	y := float64(config.Height)*spec.RelativeY - float64(partHeight)/2
	opts.GeoM.Translate(x, y)

	if spec.Opacity < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(spec.Opacity))
	}

	img.DrawImage(shape, opts)
}

// selectShapeType selects a shape type from the allowed types.
func selectShapeType(shapeTypes []shapes.ShapeType, rng *rand.Rand) shapes.ShapeType {
	if len(shapeTypes) > 0 {
		return shapeTypes[rng.Intn(len(shapeTypes))]
	}
	return shapes.ShapeCircle
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

	// Fallback to procedural random generation when no item type specified
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

// GenerateDirectionalSprites generates a 4-directional sprite sheet (Phase 4).
// Returns map[int]*ebiten.Image where keys are Direction constants (0-3).
func (g *Generator) GenerateDirectionalSprites(config Config) (map[int]*ebiten.Image, error) {
	g.logDirectionalSpriteStart(config)

	useAerial := g.checkUseAerialFlag(config)
	sprites := make(map[int]*ebiten.Image)

	if err := g.generateAllDirections(config, useAerial, sprites); err != nil {
		return nil, err
	}

	g.logDirectionalSpriteComplete(config, useAerial, len(sprites))
	return sprites, nil
}

// logDirectionalSpriteStart logs the start of directional sprite generation.
func (g *Generator) logDirectionalSpriteStart(config Config) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"type":       config.Type,
			"genreID":    config.GenreID,
			"seed":       config.Seed,
			"entityType": config.Custom["entityType"],
		}).Debug("generating directional sprite sheet")
	}
}

// checkUseAerialFlag checks if the aerial view flag is set in config.
func (g *Generator) checkUseAerialFlag(config Config) bool {
	if config.Custom != nil {
		if aerial, ok := config.Custom["useAerial"].(bool); ok {
			return aerial
		}
	}
	return false
}

// generateAllDirections generates sprites for all four directions.
func (g *Generator) generateAllDirections(config Config, useAerial bool, sprites map[int]*ebiten.Image) error {
	directions := []struct {
		index int
		name  string
	}{
		{0, "up"},
		{1, "down"},
		{2, "left"},
		{3, "right"},
	}

	for _, dir := range directions {
		dirConfig := g.createDirectionalConfig(config, dir.name, useAerial)

		sprite, err := g.Generate(dirConfig)
		if err != nil {
			if g.logger != nil {
				g.logger.WithError(err).WithField("direction", dir.name).Error("directional sprite generation failed")
			}
			return err
		}

		sprites[dir.index] = sprite
	}

	return nil
}

// createDirectionalConfig creates a config for a specific direction.
func (g *Generator) createDirectionalConfig(config Config, direction string, useAerial bool) Config {
	dirConfig := config
	if dirConfig.Custom == nil {
		dirConfig.Custom = make(map[string]interface{})
	}

	for k, v := range config.Custom {
		dirConfig.Custom[k] = v
	}

	dirConfig.Custom["facing"] = direction
	if useAerial {
		dirConfig.Custom["useAerial"] = true
	}

	return dirConfig
}

// logDirectionalSpriteComplete logs completion of directional sprite generation.
func (g *Generator) logDirectionalSpriteComplete(config Config, useAerial bool, count int) {
	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"seed":      config.Seed,
			"useAerial": useAerial,
			"count":     count,
		}).Info("directional sprite sheet generated")
	}
}
