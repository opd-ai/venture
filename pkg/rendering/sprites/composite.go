// Package sprites provides composite sprite generation with multi-layer composition.
package sprites

import (
	"fmt"
	"image/color"
	"math/rand"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// GenerateComposite creates a multi-layer composite sprite.
// Layers are rendered in ZIndex order (lowest to highest).
func (g *Generator) GenerateComposite(config CompositeConfig) (*ebiten.Image, error) {
	// Validate configuration
	if err := g.validateCompositeConfig(config); err != nil {
		return nil, fmt.Errorf("invalid composite config: %w", err)
	}

	// Create base image
	img := ebiten.NewImage(config.BaseConfig.Width, config.BaseConfig.Height)

	// Sort layers by ZIndex for proper rendering order
	sortedLayers := make([]LayerConfig, len(config.Layers))
	copy(sortedLayers, config.Layers)
	sort.Slice(sortedLayers, func(i, j int) bool {
		return sortedLayers[i].ZIndex < sortedLayers[j].ZIndex
	})

	// Render each layer
	for _, layerCfg := range sortedLayers {
		if !layerCfg.Visible {
			continue
		}

		layerImg, err := g.generateLayer(layerCfg, config.BaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate layer %s: %w", layerCfg.Type, err)
		}

		// Apply layer transformations and composite
		g.compositeLayer(img, layerImg, layerCfg)
	}

	// Apply equipment visuals
	for _, equip := range config.Equipment {
		equipImg, err := g.generateEquipmentLayer(equip, config.BaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate equipment %s: %w", equip.Slot, err)
		}

		// Find layer config for equipment
		layerCfg := g.getLayerConfigForEquipment(equip, config.Layers)
		g.compositeLayer(img, equipImg, layerCfg)
	}

	// Apply status effects
	for _, effect := range config.StatusEffects {
		if err := g.applyStatusEffect(img, effect); err != nil {
			return nil, fmt.Errorf("failed to apply status effect %s: %w", effect.Type, err)
		}
	}

	return img, nil
}

// generateLayer creates a single layer of the composite sprite.
func (g *Generator) generateLayer(layerCfg LayerConfig, baseConfig Config) (*ebiten.Image, error) {
	rng := rand.New(rand.NewSource(layerCfg.Seed))

	// Create layer-specific config
	layerConfig := shapes.Config{
		Type:      layerCfg.ShapeType,
		Width:     int(float64(baseConfig.Width) * layerCfg.Scale * g.getLayerSizeMultiplier(layerCfg.Type)),
		Height:    int(float64(baseConfig.Height) * layerCfg.Scale * g.getLayerSizeMultiplier(layerCfg.Type)),
		Color:     g.getLayerColor(layerCfg, baseConfig, rng),
		Seed:      layerCfg.Seed,
		Smoothing: 0.2,
		AntiAlias: baseConfig.AntiAlias,
	}

	// Generate shape for layer
	return g.shapeGen.Generate(layerConfig)
}

// getLayerSizeMultiplier returns the size multiplier for a layer type.
func (g *Generator) getLayerSizeMultiplier(layerType LayerType) float64 {
	switch layerType {
	case LayerBody:
		return 0.7 // Main body is 70% of sprite size
	case LayerHead:
		return 0.4 // Head is 40% of sprite size
	case LayerLegs:
		return 0.5 // Legs are 50% of sprite size
	case LayerWeapon:
		return 0.6 // Weapon is 60% of sprite size
	case LayerArmor:
		return 0.75 // Armor slightly larger than body
	case LayerAccessory:
		return 0.3 // Accessories are small
	case LayerEffect:
		return 1.0 // Effects cover full sprite
	default:
		return 0.5
	}
}

// getLayerColor determines the color for a layer.
func (g *Generator) getLayerColor(layerCfg LayerConfig, baseConfig Config, rng *rand.Rand) color.Color {
	if layerCfg.ColorTint != nil {
		// Use layer-specific color tint
		colors := []color.Color{
			layerCfg.ColorTint.Primary,
			layerCfg.ColorTint.Secondary,
			layerCfg.ColorTint.Accent1,
		}
		return colors[rng.Intn(len(colors))]
	}

	if baseConfig.Palette != nil {
		// Use base palette
		switch layerCfg.Type {
		case LayerBody:
			return baseConfig.Palette.Primary
		case LayerHead:
			return baseConfig.Palette.Secondary
		case LayerLegs:
			return baseConfig.Palette.Primary
		case LayerWeapon:
			return baseConfig.Palette.Accent1
		case LayerArmor:
			return baseConfig.Palette.Accent2
		case LayerAccessory:
			return baseConfig.Palette.Secondary
		default:
			return baseConfig.Palette.Primary
		}
	}

	return color.White
}

// compositeLayer draws a layer onto the base image with transformations.
func (g *Generator) compositeLayer(base, layer *ebiten.Image, layerCfg LayerConfig) {
	if layer == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{}

	// Apply scale
	if layerCfg.Scale != 1.0 {
		opts.GeoM.Scale(layerCfg.Scale, layerCfg.Scale)
	}

	// Calculate centered position with offset
	baseW, baseH := base.Bounds().Dx(), base.Bounds().Dy()
	layerW, layerH := layer.Bounds().Dx(), layer.Bounds().Dy()

	centerX := float64(baseW-layerW) / 2.0
	centerY := float64(baseH-layerH) / 2.0

	opts.GeoM.Translate(centerX+layerCfg.OffsetX, centerY+layerCfg.OffsetY)

	base.DrawImage(layer, opts)
}

// generateEquipmentLayer creates a visual for equipped items with Phase 15.3 enhancements.
func (g *Generator) generateEquipmentLayer(equip EquipmentVisual, baseConfig Config) (*ebiten.Image, error) {
	rng := rand.New(rand.NewSource(equip.Seed))

	// Determine equipment shape based on slot
	shapeType := g.getEquipmentShapeType(equip.Slot, rng)

	// Get material visual properties for Phase 15.3
	materialProps := GetMaterialVisualProperties(equip.Material)

	// Get damage visual effects for Phase 15.3
	damageEffects := GetDamageVisualEffects(equip.DamageState)

	// Base equipment color (from palette or material)
	equipColor := baseConfig.Palette.Accent1
	if baseConfig.Palette != nil {
		equipColor = baseConfig.Palette.Accent1
	}

	// Apply material properties to color (adjust based on material)
	equipColor = g.applyMaterialColor(equipColor, materialProps, rng)

	// Apply damage darkening
	equipColor = g.applyDamageDarkening(equipColor, damageEffects.ColorDarken)

	// Size based on detail level (Phase 15.3)
	baseSize := 0.5
	if equip.DetailLevel > 0 {
		baseSize = 0.4 + (equip.DetailLevel * 0.3) // Range: 0.4 to 0.7
	}

	config := shapes.Config{
		Type:      shapeType,
		Width:     int(float64(baseConfig.Width) * baseSize),
		Height:    int(float64(baseConfig.Height) * baseSize),
		Color:     equipColor,
		Seed:      equip.Seed,
		Smoothing: 0.15 * (1.0 - damageEffects.EdgeRoughness), // Rougher edges for damaged items
		AntiAlias: baseConfig.AntiAlias,
	}

	// Generate base equipment shape
	equipImg, err := g.shapeGen.Generate(config)
	if err != nil {
		return nil, err
	}

	// Apply enchantment glow if active (Phase 15.3)
	if equip.Enchantment.Active {
		equipImg = g.applyEnchantmentGlow(equipImg, equip.Enchantment, rng)
	}

	// Apply damage effects (cracks, dirt) if damaged (Phase 15.3)
	if damageEffects.CrackDensity > 0 {
		equipImg = g.applyDamageEffects(equipImg, damageEffects, rng)
	}

	return equipImg, nil
}

// getEquipmentShapeType returns appropriate shape for equipment slot.
func (g *Generator) getEquipmentShapeType(slot string, rng *rand.Rand) shapes.ShapeType {
	switch slot {
	case "weapon":
		// Weapons are typically rectangular or star-shaped
		shapes := []shapes.ShapeType{
			shapes.ShapeRectangle,
			shapes.ShapeStar,
			shapes.ShapeTriangle,
		}
		return shapes[rng.Intn(len(shapes))]
	case "armor":
		// Armor is typically rectangular or polygonal
		return shapes.ShapeRectangle
	case "accessory":
		// Accessories are typically circular or star-shaped
		shapes := []shapes.ShapeType{
			shapes.ShapeCircle,
			shapes.ShapeStar,
			shapes.ShapeRing,
		}
		return shapes[rng.Intn(len(shapes))]
	default:
		return shapes.ShapeCircle
	}
}

// getLayerConfigForEquipment finds or creates layer config for equipment.
func (g *Generator) getLayerConfigForEquipment(equip EquipmentVisual, layers []LayerConfig) LayerConfig {
	// Create a layer config for the equipment
	return LayerConfig{
		Type:      equip.Layer,
		ZIndex:    g.getEquipmentZIndex(equip.Layer),
		OffsetX:   g.getEquipmentOffsetX(equip.Slot),
		OffsetY:   g.getEquipmentOffsetY(equip.Slot),
		Scale:     1.0,
		Visible:   true,
		Seed:      equip.Seed,
		ShapeType: shapes.ShapeCircle, // Default, overridden during generation
	}
}

// getEquipmentZIndex returns the rendering order for equipment layers.
// Uses standardized Z-index constants to ensure correct rendering order.
// Order: Legs < Body < Armor < Head < Weapon < Accessory < Effect
func (g *Generator) getEquipmentZIndex(layerType LayerType) int {
	switch layerType {
	case LayerLegs:
		return ZIndexLegs
	case LayerBody:
		return ZIndexBody
	case LayerArmor:
		return ZIndexArmor
	case LayerHead:
		return ZIndexHead
	case LayerWeapon:
		return ZIndexWeapon
	case LayerAccessory:
		return ZIndexAccessory
	case LayerEffect:
		return ZIndexEffect
	default:
		return 0
	}
}

// getEquipmentOffsetX returns horizontal offset for equipment slot.
func (g *Generator) getEquipmentOffsetX(slot string) float64 {
	switch slot {
	case "weapon":
		return 5.0 // Slightly to the right (weapon hand)
	case "shield":
		return -5.0 // Slightly to the left
	default:
		return 0.0
	}
}

// getEquipmentOffsetY returns vertical offset for equipment slot.
func (g *Generator) getEquipmentOffsetY(slot string) float64 {
	switch slot {
	case "helmet":
		return -8.0 // Top of sprite
	case "boots":
		return 8.0 // Bottom of sprite
	default:
		return 0.0
	}
}

// applyStatusEffect overlays a visual status effect on the sprite.
func (g *Generator) applyStatusEffect(img *ebiten.Image, effect StatusEffect) error {
	bounds := img.Bounds()
	rng := rand.New(rand.NewSource(hashString(effect.Type)))

	// Get effect color
	effectColor := g.getStatusEffectColor(effect.Type, effect.Color)

	// Generate effect particles
	for i := 0; i < effect.ParticleCount; i++ {
		// Random position within sprite bounds
		x := rng.Intn(bounds.Dx())
		y := rng.Intn(bounds.Dy())

		// Create small particle
		particleSize := 2 + rng.Intn(2)
		particle := ebiten.NewImage(particleSize, particleSize)
		particle.Fill(effectColor)

		// Draw particle with transparency
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(x), float64(y))
		opts.ColorScale.ScaleAlpha(float32(effect.Intensity * 0.7))

		img.DrawImage(particle, opts)
	}

	return nil
}

// getStatusEffectColor returns the color for a status effect type.
func (g *Generator) getStatusEffectColor(effectType, customColor string) color.Color {
	if customColor != "" {
		// Parse custom color (simplified - just return predefined colors)
		switch customColor {
		case "red":
			return color.RGBA{255, 0, 0, 255}
		case "blue":
			return color.RGBA{0, 100, 255, 255}
		case "green":
			return color.RGBA{0, 255, 0, 255}
		case "purple":
			return color.RGBA{200, 0, 255, 255}
		}
	}

	// Default colors by effect type
	switch effectType {
	case "burning":
		return color.RGBA{255, 100, 0, 255} // Orange/red
	case "frozen":
		return color.RGBA{100, 200, 255, 255} // Light blue
	case "poisoned":
		return color.RGBA{100, 255, 50, 255} // Green
	case "stunned":
		return color.RGBA{255, 255, 0, 255} // Yellow
	case "blessed":
		return color.RGBA{255, 255, 200, 255} // Golden
	case "cursed":
		return color.RGBA{150, 0, 150, 255} // Purple
	default:
		return color.RGBA{255, 255, 255, 128} // White/transparent
	}
}

// validateCompositeConfig checks if the composite configuration is valid.
func (g *Generator) validateCompositeConfig(config CompositeConfig) error {
	if config.BaseConfig.Width <= 0 || config.BaseConfig.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", config.BaseConfig.Width, config.BaseConfig.Height)
	}

	if len(config.Layers) == 0 {
		return fmt.Errorf("at least one layer required")
	}

	// Validate Z-order integrity: ensure layers follow expected ordering conventions
	if err := validateZOrderIntegrity(config.Layers, config.Equipment); err != nil {
		return fmt.Errorf("z-order validation failed: %w", err)
	}

	return nil
}

// validateZOrderIntegrity ensures equipment layers have proper Z-order relationships.
// Validates that equipment layers use appropriate Z-index values relative to base layers.
func validateZOrderIntegrity(layers []LayerConfig, equipment []EquipmentVisual) error {
	// Build a map of layer types to their Z-indices
	layerZIndices := make(map[LayerType]int)
	for _, layer := range layers {
		layerZIndices[layer.Type] = layer.ZIndex
	}

	// Check that base body layer exists and has expected Z-index
	if bodyZ, hasBody := layerZIndices[LayerBody]; hasBody {
		if bodyZ != ZIndexBody {
			return fmt.Errorf("body layer has unexpected Z-index %d (expected %d)", bodyZ, ZIndexBody)
		}
	}

	// Check that head layer exists and has higher Z-index than body
	if headZ, hasHead := layerZIndices[LayerHead]; hasHead {
		if headZ != ZIndexHead {
			return fmt.Errorf("head layer has unexpected Z-index %d (expected %d)", headZ, ZIndexHead)
		}
		if bodyZ, hasBody := layerZIndices[LayerBody]; hasBody && headZ <= bodyZ {
			return fmt.Errorf("head layer Z-index %d must be higher than body Z-index %d", headZ, bodyZ)
		}
	}

	// Verify equipment layers would have correct Z-indices
	// Note: Equipment Z-indices are assigned during rendering, not in config
	// This validation ensures the system would assign proper values
	for _, equip := range equipment {
		expectedZ := getStandardZIndex(equip.Layer)

		// Ensure equipment Z-index is higher than body
		if bodyZ, hasBody := layerZIndices[LayerBody]; hasBody && expectedZ <= bodyZ {
			return fmt.Errorf("equipment layer %s would have Z-index %d which is not higher than body Z-index %d",
				equip.Layer.String(), expectedZ, bodyZ)
		}
	}

	return nil
}

// getStandardZIndex returns the standard Z-index for a layer type.
// This is a helper for validation and mirrors getEquipmentZIndex.
func getStandardZIndex(layerType LayerType) int {
	switch layerType {
	case LayerLegs:
		return ZIndexLegs
	case LayerBody:
		return ZIndexBody
	case LayerArmor:
		return ZIndexArmor
	case LayerHead:
		return ZIndexHead
	case LayerWeapon:
		return ZIndexWeapon
	case LayerAccessory:
		return ZIndexAccessory
	case LayerEffect:
		return ZIndexEffect
	default:
		return 0
	}
}

// applyMaterialColor applies material-specific color adjustments.
// Phase 15.3: Material-specific visual effects (metallic sheen, etc.)
func (g *Generator) applyMaterialColor(baseColor color.Color, materialProps MaterialVisualProperties, rng *rand.Rand) color.Color {
	rVal, gVal, bVal, aVal := baseColor.RGBA()

	// Convert to 8-bit values
	r8, g8, b8, a8 := uint8(rVal>>8), uint8(gVal>>8), uint8(bVal>>8), uint8(aVal>>8)

	// Apply sheen (brightening for metallic/crystalline materials)
	if materialProps.Sheen > 0.5 {
		brighten := uint8(materialProps.Sheen * 40)
		r8 = clampUint8(int(r8) + int(brighten))
		g8 = clampUint8(int(g8) + int(brighten))
		b8 = clampUint8(int(b8) + int(brighten))
	}

	// Apply roughness (slight desaturation for rough materials)
	if materialProps.Roughness > 0.6 {
		desat := uint8(materialProps.Roughness * 20)
		avg := (int(r8) + int(g8) + int(b8)) / 3
		r8 = uint8(int(r8) + (avg-int(r8))*int(desat)/100)
		g8 = uint8(int(g8) + (avg-int(g8))*int(desat)/100)
		b8 = uint8(int(b8) + (avg-int(b8))*int(desat)/100)
	}

	return color.RGBA{R: r8, G: g8, B: b8, A: a8}
}

// applyDamageDarkening darkens colors for damaged equipment.
// Phase 15.3: Damage state visual effects
func (g *Generator) applyDamageDarkening(c color.Color, darkenAmount float64) color.Color {
	if darkenAmount <= 0 {
		return c
	}

	rVal, gVal, bVal, aVal := c.RGBA()

	// Convert to 8-bit and darken
	multiplier := 1.0 - darkenAmount
	r8 := uint8(float64(rVal>>8) * multiplier)
	g8 := uint8(float64(gVal>>8) * multiplier)
	b8 := uint8(float64(bVal>>8) * multiplier)
	a8 := uint8(aVal >> 8)

	return color.RGBA{R: r8, G: g8, B: b8, A: a8}
}

// applyEnchantmentGlow adds glowing effect for enchanted equipment.
// Phase 15.3: Enchantment glow effects
func (g *Generator) applyEnchantmentGlow(img *ebiten.Image, enchant EnchantmentGlow, rng *rand.Rand) *ebiten.Image {
	if !enchant.Active || enchant.Intensity <= 0 {
		return img
	}

	// Get glow color
	glowColor := g.getGlowColor(enchant.Color)

	// Create glow overlay
	bounds := img.Bounds()
	glowImg := ebiten.NewImage(bounds.Dx(), bounds.Dy())

	// Fill with glow color
	glowImg.Fill(glowColor)

	// Draw glow with transparency
	opts := &ebiten.DrawImageOptions{}
	opts.ColorScale.ScaleAlpha(float32(enchant.Intensity * 0.3)) // Subtle glow
	img.DrawImage(glowImg, opts)

	// Add particles for active enchantments
	if enchant.ParticleCount > 0 {
		for i := 0; i < enchant.ParticleCount; i++ {
			x := rng.Intn(bounds.Dx())
			y := rng.Intn(bounds.Dy())

			particle := ebiten.NewImage(2, 2)
			particle.Fill(glowColor)

			particleOpts := &ebiten.DrawImageOptions{}
			particleOpts.GeoM.Translate(float64(x), float64(y))
			particleOpts.ColorScale.ScaleAlpha(float32(enchant.Intensity * 0.6))
			img.DrawImage(particle, particleOpts)
		}
	}

	return img
}

// applyDamageEffects adds cracks and dirt to damaged equipment.
// Phase 15.3: Damage state visual effects (cracks, dirt)
func (g *Generator) applyDamageEffects(img *ebiten.Image, effects DamageVisualEffects, rng *rand.Rand) *ebiten.Image {
	bounds := img.Bounds()

	// Apply cracks
	if effects.CrackDensity > 0 {
		crackColor := color.RGBA{R: 40, G: 40, B: 40, A: 200}
		numCracks := int(effects.CrackDensity * 10)

		for i := 0; i < numCracks; i++ {
			// Simple line crack
			x1 := rng.Intn(bounds.Dx())
			y1 := rng.Intn(bounds.Dy())
			length := 2 + rng.Intn(4)

			for j := 0; j < length; j++ {
				crack := ebiten.NewImage(1, 1)
				crack.Fill(crackColor)

				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(x1+j), float64(y1))
				opts.ColorScale.ScaleAlpha(float32(effects.CrackDensity))
				img.DrawImage(crack, opts)
			}
		}
	}

	// Apply dirtiness
	if effects.Dirtiness > 0 {
		dirtColor := color.RGBA{R: 60, G: 50, B: 40, A: 150}
		numDirtSpots := int(effects.Dirtiness * 8)

		for i := 0; i < numDirtSpots; i++ {
			x := rng.Intn(bounds.Dx())
			y := rng.Intn(bounds.Dy())
			size := 1 + rng.Intn(2)

			dirt := ebiten.NewImage(size, size)
			dirt.Fill(dirtColor)

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(x), float64(y))
			opts.ColorScale.ScaleAlpha(float32(effects.Dirtiness * 0.5))
			img.DrawImage(dirt, opts)
		}
	}

	// Apply opacity reduction for heavily damaged items
	if effects.OpacityMultiplier < 1.0 {
		// Create a copy with reduced opacity
		// Note: Ebiten doesn't support direct opacity, so we simulate with ColorScale
		// This is a simplification - in production might use more sophisticated approach
	}

	return img
}

// getGlowColor returns the color for enchantment glow.
func (g *Generator) getGlowColor(colorName string) color.Color {
	switch colorName {
	case "green":
		return color.RGBA{R: 0, G: 255, B: 100, A: 255}
	case "blue":
		return color.RGBA{R: 100, G: 150, B: 255, A: 255}
	case "purple":
		return color.RGBA{R: 200, G: 100, B: 255, A: 255}
	case "gold":
		return color.RGBA{R: 255, G: 215, B: 0, A: 255}
	case "red":
		return color.RGBA{R: 255, G: 50, B: 50, A: 255}
	case "cyan":
		return color.RGBA{R: 0, G: 255, B: 255, A: 255}
	default:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
}

// clampUint8 clamps an integer to valid uint8 range.
func clampUint8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
