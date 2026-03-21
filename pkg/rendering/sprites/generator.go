// Package sprites provides procedural sprite generation.
// This file implements sprite generators that create entity visuals
// at runtime without external assets.
package sprites

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/sirupsen/logrus"
)

// safeReadPixels attempts to read pixels from an ebiten image.
// Returns false if ReadPixels is not available (e.g., game not started in tests).
func safeReadPixels(img *ebiten.Image, pix []byte) (ok bool) {
	defer func() {
		if recov := recover(); recov != nil {
			logrus.WithField("panic", recov).Debug("safeReadPixels: recovered from panic (expected in tests without game loop)")
			ok = false
		}
	}()
	img.ReadPixels(pix)
	return true
}

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
// logGenerationStart logs the start of sprite generation at debug level
func (g *Generator) logGenerationStart(config Config) {
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
}

// ensurePalette generates a palette for the config if not already provided
func (g *Generator) ensurePalette(config *Config) error {
	if config.Palette != nil {
		return nil
	}

	var pal *palette.Palette
	var err error
	if config.PaletteOptions != nil {
		pal, err = g.paletteGen.GenerateWithOptions(config.GenreID, config.Seed, *config.PaletteOptions)
	} else {
		pal, err = g.paletteGen.Generate(config.GenreID, config.Seed)
	}
	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("palette generation failed")
		}
		return err
	}
	config.Palette = pal
	return nil
}

// generateByType creates a sprite image based on the configured sprite type
func (g *Generator) generateByType(config Config, rng *rand.Rand) (*ebiten.Image, error) {
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

func (g *Generator) Generate(config Config) (*ebiten.Image, error) {
	g.logGenerationStart(config)

	if err := g.ensurePalette(&config); err != nil {
		return nil, err
	}

	seedGen := procgen.NewSeedGenerator(config.Seed)
	rng := rand.New(rand.NewSource(seedGen.GetSeed("sprite", config.Variation)))

	img, err := g.generateByType(config, rng)
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
		AntiAlias: config.AntiAlias,
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
			AntiAlias: config.AntiAlias,
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

	var traits *AvatarTraits
	if useAerial {
		traits = g.configureAerialTraits(config, entityType, &template)
	}

	g.renderTemplatePartsWithTraits(img, template, config, rng, traits)

	ctx := &entityRenderContext{
		img:        img,
		config:     config,
		entityType: entityType,
		direction:  direction,
		genre:      genre,
		useAerial:  useAerial,
		template:   &template,
		traits:     traits,
	}

	ctx.renderCreatureDetails()
	ctx.renderGarmentDetails()
	ctx.renderRoleDetails()

	if useAerial {
		overlayEquipmentVisuals(img, config)
	}

	ctx.renderBackAccessory()
	ctx.renderHeadgear()
	ctx.applySurfaceTextures()
	ctx.applyDepthEnhancement()
	ctx.applyColorTemperature()
	ctx.finalizeSprite()

	return img, nil
}

// configureAerialTraits generates and applies traits for aerial view rendering.
func (g *Generator) configureAerialTraits(config Config, entityType string, template *AnatomicalTemplate) *AvatarTraits {
	t := generateTraitsForEntity(config.Seed, entityType)
	*template = applyTraitProportions(*template, &t)

	bodyType := t.BodyBuild
	if config.Custom != nil {
		if btVal, ok := config.Custom["bodyType"].(int); ok {
			bodyType = BodyType(btVal)
		}
	}
	*template = ApplyBodyTypeToTemplate(*template, bodyType)

	sizeClass := extractSizeClass(config)
	*template = ApplySizeScaling(*template, sizeClass)

	return &t
}

// extractSizeClass extracts the size class from config custom data.
// hasEquipmentHelmet returns true if the config contains an explicit helmet
// equipment visual, in which case the default headgear overlay is skipped.
func hasEquipmentHelmet(config Config) bool {
	if config.Custom == nil {
		return false
	}
	slots, ok := config.Custom["equipmentVisuals"].([]EquipmentVisual)
	if !ok {
		return false
	}
	for _, s := range slots {
		if s.Slot == "helmet" {
			return true
		}
	}
	return false
}

// resolveHeadgearType extracts the headgear type from Config.Custom or
// derives one from the entity seed and role. Returns HeadgearNone if no
// headgear should be drawn.
func resolveHeadgearType(config Config) HeadgearType {
	if config.Custom != nil {
		if ht, ok := config.Custom["headgearType"].(int); ok {
			return HeadgearType(ht)
		}
	}
	// Derive from seed + entity type
	entityType := ""
	genre := ""
	if config.Custom != nil {
		if et, ok := config.Custom["entityType"].(string); ok {
			entityType = et
		}
		if g, ok := config.Custom["genre"].(string); ok {
			genre = g
		}
	}
	role := string(MapEntityTypeToRole(entityType))
	if role == "" {
		role = "npc"
	}
	return SelectHeadgearForRole(role, genre, config.Seed)
}

// resolveBackAccessoryType extracts the back accessory type from Config.Custom
// or derives one from the entity seed and role. Returns BackAccessoryNone if
// no back accessory should be drawn.
func resolveBackAccessoryType(config Config) BackAccessoryType {
	if config.Custom != nil {
		if bat, ok := config.Custom["backAccessoryType"].(int); ok {
			return BackAccessoryType(bat)
		}
	}
	entityType := ""
	genre := ""
	if config.Custom != nil {
		if et, ok := config.Custom["entityType"].(string); ok {
			entityType = et
		}
		if g, ok := config.Custom["genre"].(string); ok {
			genre = g
		}
	}
	role := string(MapEntityTypeToRole(entityType))
	if role == "" {
		role = "npc"
	}
	return SelectBackAccessoryForRole(role, genre, config.Seed)
}

func extractSizeClass(config Config) string {
	if config.Custom != nil {
		if sc, ok := config.Custom["sizeClass"].(string); ok {
			return sc
		}
	}
	return "medium"
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
// Aerial (top-down) view is the default since the game uses a top-down camera.
// Set useAerial=false in Custom to opt out (legacy profile-view).
func extractAerialFlag(config Config) bool {
	if config.Custom != nil {
		if aerial, ok := config.Custom["useAerial"].(bool); ok {
			return aerial
		}
	}
	return true
}

// selectEntityTemplate selects the appropriate anatomical template based on entity type and configuration.
// Aerial (top-down) templates are preferred since the game uses a top-down camera.
func selectEntityTemplate(entityType, genre string, direction Direction, hasWeapon, hasShield, useAerial bool) AnatomicalTemplate {
	if useAerial {
		return SelectAerialTemplate(entityType, genre, direction)
	}
	// Legacy fallback when useAerial is explicitly false
	isHumanoid := isHumanoidType(entityType)
	if isHumanoid && (hasWeapon || hasShield) {
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
	return IsHumanoidEntity(entityType)
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
	g.renderTemplatePartsWithTraits(img, template, config, rng, nil)
}

// renderTemplatePartsWithTraits renders all template parts, applying optional seed-based
// avatar traits for color and proportion variety. For humanoid entities with traits,
// a hair overlay is rendered after the head part for visual distinctness.
func (g *Generator) renderTemplatePartsWithTraits(img *ebiten.Image, template AnatomicalTemplate, config Config, rng *rand.Rand, traits *AvatarTraits) {
	parts := template.GetSortedParts()

	// Extract per-frame body part offsets from config for animated frame generation.
	var frameOffsets FrameOffsetMap
	if config.Custom != nil {
		if fo, ok := config.Custom["frameOffsets"].(FrameOffsetMap); ok {
			frameOffsets = fo
		}
	}

	for _, partData := range parts {
		spec := partData.Spec
		if frameOffsets != nil {
			spec = applyFrameOffsetsToSpec(spec, partData.Part, frameOffsets, config.Width, config.Height)
		}
		g.renderTemplatePartWithTraits(img, spec, config, rng, traits)
	}

	// Render hair overlay for humanoid entities with aerial traits
	if traits != nil && traits.HairStyle >= 0 {
		headSpec, hasHead := template.BodyPartLayout[PartHead]
		if hasHead {
			direction := DirDown
			if config.Custom != nil {
				if dir, ok := config.Custom["facing"].(string); ok {
					direction = Direction(dir)
				}
			}
			hairParams := ComputeHairRenderParams(config.Width, config.Height, headSpec, traits, direction, config.Seed)
			hairBuf := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
			RenderHairOverlay(hairBuf, hairParams)
			hairImg := ebiten.NewImageFromImage(hairBuf)
			img.DrawImage(hairImg, nil)

			// Render face detail (eyes, mouth) on top of hair
			faceBuf := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
			faceParams := g.buildFaceParams(config, headSpec, direction)
			RenderFaceDetail(faceBuf, faceParams)
			faceImg := ebiten.NewImageFromImage(faceBuf)
			img.DrawImage(faceImg, nil)
		}
	}
}

// buildFaceParams constructs FaceRenderParams from config, using NPC facial
// detail data from Custom map if available, otherwise seed-based defaults.
func (g *Generator) buildFaceParams(config Config, headSpec PartSpec, direction Direction) FaceRenderParams {
	if config.Custom != nil {
		if eyeR, ok := config.Custom["faceEyeR"].(float64); ok {
			eyeG, _ := config.Custom["faceEyeG"].(float64)
			eyeB, _ := config.Custom["faceEyeB"].(float64)
			mouthR, _ := config.Custom["faceMouthR"].(float64)
			mouthG, _ := config.Custom["faceMouthG"].(float64)
			mouthB, _ := config.Custom["faceMouthB"].(float64)
			eyeSize := 2.0
			if es, ok := config.Custom["faceEyeSize"].(float64); ok {
				eyeSize = es
			}
			mouthSize := 1.0
			if ms, ok := config.Custom["faceMouthSize"].(float64); ok {
				mouthSize = ms
			}
			expression := "neutral"
			if expr, ok := config.Custom["faceExpression"].(string); ok {
				expression = expr
			}
			return ComputeFaceParamsFromComponent(
				config.Width, config.Height, headSpec, direction, config.Seed,
				eyeR, eyeG, eyeB, mouthR, mouthG, mouthB,
				eyeSize, mouthSize, expression,
			)
		}
	}
	return ComputeFaceParams(config.Width, config.Height, headSpec, direction, config.Seed)
}

// renderTemplatePart renders a single template part to the image with depth shading.
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
		AntiAlias: config.AntiAlias,
	}

	// Generate as raw RGBA for pixel-level shading
	rgbaImg, err := g.shapeGen.GenerateRGBA(shapeConfig)
	if err != nil {
		return
	}

	// Apply per-body-part depth shading
	genre := config.GenreID
	baseCfg := GenreShadingConfig(genre)
	partCfg := ShadingConfigForPart(baseCfg, spec.ColorRole, spec.ZIndex)
	ApplyBodyPartShading(rgbaImg, partCfg, config.Seed+int64(spec.ZIndex)*31)

	shape := ebiten.NewImageFromImage(rgbaImg)

	opts := &ebiten.DrawImageOptions{}
	x := float64(config.Width)*spec.RelativeX - float64(partWidth)/2
	y := float64(config.Height)*spec.RelativeY - float64(partHeight)/2
	opts.GeoM.Translate(x, y)

	if spec.Opacity < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(spec.Opacity))
	}

	img.DrawImage(shape, opts)
}

// renderTemplatePartWithTraits renders a body part with optional seed-based color variety.
func (g *Generator) renderTemplatePartWithTraits(img *ebiten.Image, spec PartSpec, config Config, rng *rand.Rand, traits *AvatarTraits) {
	if traits == nil {
		g.renderTemplatePart(img, spec, config, rng)
		return
	}

	partWidth := int(float64(config.Width) * spec.RelativeWidth)
	partHeight := int(float64(config.Height) * spec.RelativeHeight)

	if partWidth <= 0 || partHeight <= 0 {
		return
	}

	shapeType := selectShapeType(spec.ShapeTypes, rng)

	// Use trait-derived color if available, otherwise fall back to palette
	partColor := traits.ColorForBodyPart(spec.ColorRole, spec.ZIndex)
	if partColor == nil {
		partColor = g.getColorForRole(spec.ColorRole, config.Palette)
	}

	shapeConfig := shapes.Config{
		Type:      shapeType,
		Width:     partWidth,
		Height:    partHeight,
		Color:     partColor,
		Seed:      config.Seed + int64(spec.ZIndex),
		Smoothing: 0.2,
		Rotation:  spec.Rotation,
		AntiAlias: config.AntiAlias,
	}

	rgbaImg, err := g.shapeGen.GenerateRGBA(shapeConfig)
	if err != nil {
		return
	}

	genre := config.GenreID
	baseCfg := GenreShadingConfig(genre)
	partCfg := ShadingConfigForPart(baseCfg, spec.ColorRole, spec.ZIndex)
	ApplyBodyPartShading(rgbaImg, partCfg, config.Seed+int64(spec.ZIndex)*31)

	// Apply clothing patterns to garment body parts (torso, arms, legs)
	clothingPattern := traits.ClothingPatterns.PatternForBodyRegion(spec.ColorRole, spec.ZIndex)
	ApplyClothingPattern(rgbaImg, clothingPattern, config.Seed+int64(spec.ZIndex)*17)

	shape := ebiten.NewImageFromImage(rgbaImg)

	opts := &ebiten.DrawImageOptions{}
	x := float64(config.Width)*spec.RelativeX - float64(partWidth)/2
	y := float64(config.Height)*spec.RelativeY - float64(partHeight)/2
	opts.GeoM.Translate(x, y)

	if spec.Opacity < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(spec.Opacity))
	}

	img.DrawImage(shape, opts)
}

// generateTraitsForEntity creates seed-based visual traits for the given entity.
func generateTraitsForEntity(seed int64, entityType string) AvatarTraits {
	if IsHumanoidEntity(entityType) {
		return GenerateAvatarTraits(seed)
	}
	return GenerateCreatureTraits(seed, entityType)
}

// applyTraitProportions adjusts template part dimensions using trait scales.
func applyTraitProportions(template AnatomicalTemplate, traits *AvatarTraits) AnatomicalTemplate {
	if traits == nil {
		return template
	}
	adjusted := AnatomicalTemplate{
		Name:           template.Name,
		BodyPartLayout: make(map[BodyPart]PartSpec, len(template.BodyPartLayout)),
	}
	for part, spec := range template.BodyPartLayout {
		adjusted.BodyPartLayout[part] = traits.ApplyProportions(spec, part)
	}
	return adjusted
}

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

	itemType, rarity := g.extractItemMetadata(config)

	// Use template-based generation if item type specified
	if itemType != "" {
		return g.generateItemWithTemplate(config, itemType, rarity, rng)
	}

	// Fallback to procedural random generation when no item type specified
	g.generateProceduralItemShapes(img, config, rng)

	return img, nil
}

// extractItemMetadata extracts item type and rarity from config custom data.
func (g *Generator) extractItemMetadata(config Config) (ItemType, ItemRarity) {
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

	return itemType, rarity
}

// generateProceduralItemShapes generates multiple random shapes for item sprite.
func (g *Generator) generateProceduralItemShapes(img *ebiten.Image, config Config, rng *rand.Rand) {
	numShapes := 1 + int(config.Complexity*2)

	for i := 0; i < numShapes; i++ {
		colorChoice := g.selectShapeColor(i, config, rng)
		shapeConfig := g.createItemShapeConfig(i, config, colorChoice, rng)

		shape, err := g.shapeGen.Generate(shapeConfig)
		if err != nil {
			continue
		}

		g.drawCenteredShape(img, shape, config.Width, config.Height, shapeConfig.Width, shapeConfig.Height)
	}
}

// selectShapeColor chooses color for item shape based on index.
func (g *Generator) selectShapeColor(index int, config Config, rng *rand.Rand) color.Color {
	if index == 0 {
		return config.Palette.Secondary
	}
	return config.Palette.Colors[rng.Intn(len(config.Palette.Colors))]
}

// createItemShapeConfig creates shape configuration for procedural item generation.
func (g *Generator) createItemShapeConfig(index int, config Config, colorChoice color.Color, rng *rand.Rand) shapes.Config {
	return shapes.Config{
		Type:       shapes.ShapeType(rng.Intn(6)),
		Width:      int(float64(config.Width) * (0.5 + rng.Float64()*0.4)),
		Height:     int(float64(config.Height) * (0.5 + rng.Float64()*0.4)),
		Color:      colorChoice,
		Seed:       config.Seed + int64(index),
		Sides:      4 + rng.Intn(4),
		InnerRatio: 0.3 + rng.Float64()*0.4,
		Rotation:   rng.Float64() * 360,
		Smoothing:  0.1,
		AntiAlias:  config.AntiAlias,
	}
}

// drawCenteredShape draws a shape centered within the target image.
func (g *Generator) drawCenteredShape(img, shape *ebiten.Image, targetWidth, targetHeight, shapeWidth, shapeHeight int) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		float64(targetWidth-shapeWidth)/2,
		float64(targetHeight-shapeHeight)/2,
	)
	img.DrawImage(shape, opts)
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
			AntiAlias: config.AntiAlias,
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
		AntiAlias: config.AntiAlias,
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
				AntiAlias: config.AntiAlias,
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
		AntiAlias: config.AntiAlias,
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
		AntiAlias: config.AntiAlias,
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
		AntiAlias:  config.AntiAlias,
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

// GenerateFromParams implements the procgen.Generator interface.
// It creates a sprite based on the seed and generation parameters.
func (g *Generator) GenerateFromParams(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Build Config from GenerationParams
	config := Config{
		Type:       SpriteEntity, // Default type
		Width:      32,           // Default width
		Height:     32,           // Default height
		Seed:       seed,
		GenreID:    params.GenreID,
		Complexity: params.Difficulty,
	}

	// Extract custom parameters if provided
	if params.Custom != nil {
		if w, ok := params.Custom["width"].(int); ok {
			config.Width = w
		}
		if h, ok := params.Custom["height"].(int); ok {
			config.Height = h
		}
		if t, ok := params.Custom["type"].(SpriteType); ok {
			config.Type = t
		} else if ts, ok := params.Custom["type"].(string); ok {
			switch ts {
			case "entity":
				config.Type = SpriteEntity
			case "item":
				config.Type = SpriteItem
			case "tile":
				config.Type = SpriteTile
			case "particle":
				config.Type = SpriteParticle
			case "ui":
				config.Type = SpriteUI
			}
		}
		if v, ok := params.Custom["variation"].(int); ok {
			config.Variation = v
		}
		// Pass through remaining custom parameters
		config.SetCustom(params.Custom)
	}

	return g.Generate(config)
}

// Validate implements the procgen.Generator interface.
// It checks if the generated sprite meets minimum quality thresholds.
func (g *Generator) Validate(result interface{}) error {
	img, ok := result.(*ebiten.Image)
	if !ok {
		return fmt.Errorf("invalid result type: expected *ebiten.Image, got %T", result)
	}

	if img == nil {
		return fmt.Errorf("sprite is nil")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid sprite dimensions: %dx%d", width, height)
	}

	// Check for minimum opacity coverage (at least 1% of pixels should be non-transparent)
	// This ensures the sprite isn't completely empty
	minOpaquePixels := (width * height) / 100
	if minOpaquePixels < 1 {
		minOpaquePixels = 1
	}

	opaqueCount := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				opaqueCount++
				if opaqueCount >= minOpaquePixels {
					return nil // Early exit once threshold met
				}
			}
		}
	}

	if opaqueCount < minOpaquePixels {
		return fmt.Errorf("sprite has insufficient opacity coverage: %d opaque pixels (minimum: %d)", opaqueCount, minOpaquePixels)
	}

	return nil
}
