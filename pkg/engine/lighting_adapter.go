package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/lighting"
	"github.com/sirupsen/logrus"
)

// LightingAdapter wraps pkg/rendering/lighting.System for ECS integration.
// Provides Phase 2.2 dynamic lighting effects with multiple light sources.
type LightingAdapter struct {
	system  *lighting.System
	enabled bool
	logger  *logrus.Entry
}

// NewLightingAdapter creates a lighting adapter with default configuration.
func NewLightingAdapter(logger *logrus.Entry) *LightingAdapter {
	return NewLightingAdapterWithConfig(lighting.DefaultConfig(), logger)
}

// NewLightingAdapterWithConfig creates a lighting adapter with custom configuration.
func NewLightingAdapterWithConfig(config lighting.LightingConfig, logger *logrus.Entry) *LightingAdapter {
	return &LightingAdapter{
		system:  lighting.NewSystemWithConfig(config),
		enabled: true,
		logger:  logger,
	}
}

// SetEnabled enables or disables lighting effects.
func (l *LightingAdapter) SetEnabled(enabled bool) {
	l.enabled = enabled
	if l.logger != nil {
		l.logger.WithField("enabled", enabled).Debug("lighting enabled state changed")
	}
}

// IsEnabled returns whether lighting is enabled.
func (l *LightingAdapter) IsEnabled() bool {
	return l.enabled
}

// AddLight adds a light source to the system.
func (l *LightingAdapter) AddLight(light lighting.Light) error {
	return l.system.AddLight(light)
}

// RemoveLight removes a light at the specified index.
func (l *LightingAdapter) RemoveLight(index int) error {
	return l.system.RemoveLight(index)
}

// ClearLights removes all light sources.
func (l *LightingAdapter) ClearLights() {
	l.system.ClearLights()
}

// LightCount returns the number of active lights.
func (l *LightingAdapter) LightCount() int {
	return l.system.LightCount()
}

// Update processes entity lighting components (implements System interface).
func (l *LightingAdapter) Update(entities []*Entity, deltaTime float64) {
	if !l.enabled {
		return
	}

	l.system.ClearLights()

	for _, entity := range entities {
		l.processEntityLight(entity)
	}
}

// processEntityLight extracts light data from entity and adds to lighting system.
func (l *LightingAdapter) processEntityLight(entity *Entity) {
	lightComp, posComp := l.extractLightComponents(entity)
	if lightComp == nil || posComp == nil {
		return
	}

	light := l.convertToLight(lightComp, posComp)
	if err := l.system.AddLight(light); err != nil {
		if l.logger != nil {
			// Maximum lights reached is an expected operational condition during culling
			// Log at debug level to avoid flooding logs
			l.logger.WithError(err).Debug("failed to add light from entity")
		}
	}
}

// extractLightComponents extracts and validates light and position components.
func (l *LightingAdapter) extractLightComponents(entity *Entity) (*LightComponent, *PositionComponent) {
	lightComp, ok := entity.GetComponent("light")
	if !ok || lightComp == nil {
		return nil, nil
	}

	posComp, ok := entity.GetComponent("position")
	if !ok || posComp == nil {
		return nil, nil
	}

	lc, ok := lightComp.(*LightComponent)
	if !ok {
		return nil, nil
	}

	pc, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil
	}

	return lc, pc
}

// convertToLight converts ECS light components to lighting.Light.
func (l *LightingAdapter) convertToLight(lc *LightComponent, pc *PositionComponent) lighting.Light {
	falloff := l.convertFalloffType(lc.Falloff)
	return lighting.Light{
		Type:      lighting.TypePoint,
		Position:  image.Point{X: int(pc.X), Y: int(pc.Y)},
		Color:     lc.Color,
		Intensity: lc.Intensity,
		Radius:    lc.Radius,
		Falloff:   falloff,
		Enabled:   lc.Enabled,
	}
}

// convertFalloffType converts LightFalloffType to lighting.FalloffType.
func (l *LightingAdapter) convertFalloffType(falloff LightFalloffType) lighting.FalloffType {
	switch falloff {
	case FalloffConstant:
		return lighting.FalloffNone
	case FalloffLinear:
		return lighting.FalloffLinear
	case FalloffQuadratic:
		return lighting.FalloffQuadratic
	case FalloffInverseSquare:
		return lighting.FalloffInverseSquare
	default:
		return lighting.FalloffLinear
	}
}

// ApplyLighting applies lighting effects to the rendered scene.
// Should be called during rendering pass, not in Update.
func (l *LightingAdapter) ApplyLighting(scene *ebiten.Image) *ebiten.Image {
	if !l.enabled {
		return scene
	}

	// Convert Ebiten image to RGBA
	bounds := scene.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, scene.At(x, y))
		}
	}

	// Apply lighting
	lit := l.system.ApplyLighting(rgba)

	// Convert back to Ebiten image
	result := ebiten.NewImageFromImage(lit)
	return result
}
