//go:build !test
// +build !test

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

// generateEquipmentLayer creates a visual for equipped items.
func (g *Generator) generateEquipmentLayer(equip EquipmentVisual, baseConfig Config) (*ebiten.Image, error) {
	rng := rand.New(rand.NewSource(equip.Seed))

	// Determine equipment shape based on slot
	shapeType := g.getEquipmentShapeType(equip.Slot, rng)

	config := shapes.Config{
		Type:      shapeType,
		Width:     int(float64(baseConfig.Width) * 0.5),
		Height:    int(float64(baseConfig.Height) * 0.5),
		Color:     baseConfig.Palette.Accent1,
		Seed:      equip.Seed,
		Smoothing: 0.15,
	}

	return g.shapeGen.Generate(config)
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
func (g *Generator) getEquipmentZIndex(layerType LayerType) int {
	switch layerType {
	case LayerBody:
		return 10
	case LayerHead:
		return 20
	case LayerLegs:
		return 5
	case LayerArmor:
		return 15
	case LayerWeapon:
		return 25
	case LayerAccessory:
		return 30
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

	return nil
}
