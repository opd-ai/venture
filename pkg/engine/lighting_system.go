// Package engine provides the dynamic lighting system.
// This file implements LightingSystem which processes light sources and applies
// lighting calculations to the rendered scene. The system supports point lights,
// ambient lighting, and various falloff curves for realistic illumination.
//
// Design Philosophy:
// - Performance-conscious: viewport culling, light limits, deferred rendering
// - Genre-aware: ambient light configured per genre for appropriate atmosphere
// - Extensible: supports multiple light types and falloff curves
// - Integration: works with existing render pipeline via post-processing
package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// LightingSystem processes light sources and applies lighting to the scene.
// This system runs after the main render pass to apply lighting effects
// as a post-processing step.
type LightingSystem struct {
	world  *World
	config *LightingConfig
	logger *logrus.Entry

	// Viewport tracking for culling
	cameraX     float64
	cameraY     float64
	viewportW   int
	viewportH   int
	viewportSet bool

	// Light collection (reused each frame)
	visibleLights []*lightWithPosition

	// Lighting buffer (reused each frame)
	lightingBuffer *ebiten.Image

	// Cached ambient light entity (avoid O(n) search each frame)
	ambientLightEntityID uint64
	ambientLightCached   bool
}

// lightWithPosition combines a light component with its world position.
type lightWithPosition struct {
	light *LightComponent
	x     float64
	y     float64
}

// NewLightingSystem creates a new lighting system.
func NewLightingSystem(world *World, config *LightingConfig) *LightingSystem {
	return NewLightingSystemWithLogger(world, config, nil)
}

// NewLightingSystemWithLogger creates a new lighting system with a logger.
func NewLightingSystemWithLogger(world *World, config *LightingConfig, logger *logrus.Logger) *LightingSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "lighting")
	}

	if config == nil {
		config = NewLightingConfig()
	}

	return &LightingSystem{
		world:         world,
		config:        config,
		logger:        logEntry,
		visibleLights: make([]*lightWithPosition, 0, config.MaxLights),
	}
}

// SetViewport updates the camera position and viewport size for culling.
func (s *LightingSystem) SetViewport(cameraX, cameraY float64, width, height int) {
	s.cameraX = cameraX
	s.cameraY = cameraY
	s.viewportW = width
	s.viewportH = height
	s.viewportSet = true
}

// Update processes lighting each frame (updates animation times).
func (s *LightingSystem) Update(entities []*Entity, deltaTime float64) {
	if !s.config.Enabled {
		return
	}

	// Update animation times for flickering/pulsing lights
	for _, entity := range entities {
		comp, ok := entity.GetComponent("light")
		if !ok {
			continue
		}

		light, ok := comp.(*LightComponent)
		if !ok {
			continue
		}

		light.internalTime += deltaTime
	}
}

// CollectVisibleLights gathers lights within viewport for rendering.
// Returns the collected lights sorted by priority (closest first).
func (s *LightingSystem) CollectVisibleLights(entities []*Entity) []*lightWithPosition {
	s.visibleLights = s.visibleLights[:0]

	if !s.config.Enabled {
		return s.visibleLights
	}

	for _, entity := range entities {
		// Get light component
		lightComp, hasLight := entity.GetComponent("light")
		if !hasLight {
			continue
		}

		light, ok := lightComp.(*LightComponent)
		if !ok || !light.Enabled {
			continue
		}

		// Get position
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Viewport culling (if viewport is set)
		if s.viewportSet {
			if !s.isLightInViewport(pos.X, pos.Y, light.Radius) {
				continue
			}
		}

		// Add to visible lights
		s.visibleLights = append(s.visibleLights, &lightWithPosition{
			light: light,
			x:     pos.X,
			y:     pos.Y,
		})

		// Limit to MaxLights
		if len(s.visibleLights) >= s.config.MaxLights {
			break
		}
	}

	return s.visibleLights
}

// isLightInViewport checks if a light affects the viewport.
func (s *LightingSystem) isLightInViewport(x, y, radius float64) bool {
	// Expand viewport by light radius for overlap detection
	minX := s.cameraX - radius
	maxX := s.cameraX + float64(s.viewportW) + radius
	minY := s.cameraY - radius
	maxY := s.cameraY + float64(s.viewportH) + radius

	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

// SetAmbientLightEntity sets the cached ambient light entity ID.
// This should be called when creating or changing the ambient light entity
// to avoid O(n) iteration on every frame.
func (s *LightingSystem) SetAmbientLightEntity(entityID uint64) {
	s.ambientLightEntityID = entityID
	s.ambientLightCached = true
}

// ClearAmbientLightCache clears the cached ambient light entity.
// Call this if the ambient light entity is removed from the world.
func (s *LightingSystem) ClearAmbientLightCache() {
	s.ambientLightEntityID = 0
	s.ambientLightCached = false
}

// ApplyLighting applies lighting effects to a rendered image.
// This is called after the main render pass as a post-processing step.
// Returns a new image with lighting applied (input image is not modified).
func (s *LightingSystem) ApplyLighting(screen, renderedScene *ebiten.Image, entities []*Entity) {
	if !s.config.Enabled {
		// No lighting, just draw the scene
		screen.DrawImage(renderedScene, nil)
		return
	}

	// Collect visible lights
	lights := s.CollectVisibleLights(entities)

	// Get ambient light from cache or config defaults
	ambientIntensity := s.config.AmbientIntensity
	ambientColor := s.config.AmbientColor

	// Try cached ambient light entity first
	if s.ambientLightCached {
		if entity, ok := s.world.GetEntity(s.ambientLightEntityID); ok {
			if ambComp, ok := entity.GetComponent("ambient_light"); ok {
				if ambient, ok := ambComp.(*AmbientLightComponent); ok {
					ambientIntensity = ambient.Intensity
					ambientColor = ambient.Color
				}
			}
		}
	}

	// If no lights and high ambient, just draw normally
	if len(lights) == 0 && ambientIntensity > 0.8 {
		screen.DrawImage(renderedScene, nil)
		return
	}

	// Create lighting buffer if needed
	w, h := renderedScene.Size()
	if s.lightingBuffer == nil || s.lightingBuffer.Bounds().Dx() != w || s.lightingBuffer.Bounds().Dy() != h {
		s.lightingBuffer = ebiten.NewImage(w, h)
	}

	// Calculate lighting (simplified version - full per-pixel lighting would be more complex)
	// For now, we'll use a blend approach with colored overlays
	s.lightingBuffer.Clear()

	// Apply ambient base
	ambR := float64(ambientColor.R) / 255.0 * ambientIntensity
	ambG := float64(ambientColor.G) / 255.0 * ambientIntensity
	ambB := float64(ambientColor.B) / 255.0 * ambientIntensity

	// Draw rendered scene with ambient modulation
	opts := &ebiten.DrawImageOptions{}
	opts.ColorScale.Scale(float32(ambR), float32(ambG), float32(ambB), 1.0)
	s.lightingBuffer.DrawImage(renderedScene, opts)

	// Apply point lights additively
	// Note: Full lighting would require shader support or per-pixel calculations
	// This is a simplified version using additive blending
	for _, lwp := range lights {
		s.applyPointLight(s.lightingBuffer, renderedScene, lwp)
	}

	// Draw final result to screen
	screen.DrawImage(s.lightingBuffer, nil)
}

// applyPointLight applies a single point light to the lighting buffer.
// This is a simplified implementation; full lighting would use shaders.
func (s *LightingSystem) applyPointLight(lightBuffer, scene *ebiten.Image, lwp *lightWithPosition) {
	intensity := lwp.light.GetCurrentIntensity()
	if intensity <= 0 {
		return
	}

	// Create a light influence area (simplified)
	// In a full implementation, this would be a radial gradient
	radius := int(lwp.light.Radius)
	x := int(lwp.x - s.cameraX)
	y := int(lwp.y - s.cameraY)

	// Draw light influence as additive blend
	// This is a placeholder - real implementation would need proper radial gradients
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x-radius), float64(y-radius))

	// Modulate by light color and intensity
	r := float64(lwp.light.Color.R) / 255.0 * intensity * 0.3 // 0.3 is blend strength
	g := float64(lwp.light.Color.G) / 255.0 * intensity * 0.3
	b := float64(lwp.light.Color.B) / 255.0 * intensity * 0.3

	opts.ColorScale.Scale(float32(r), float32(g), float32(b), 1.0)
	opts.Blend = ebiten.BlendLighter // Additive blending

	// Minimal implementation: draw a filled white circle as the light influence
	diameter := 2 * radius
	if diameter <= 0 {
		return
	}
	lightImg := ebiten.NewImage(diameter, diameter)
	// Fill with white, but only inside the circle
	cx, cy := float64(radius), float64(radius)
	for py := 0; py < diameter; py++ {
		for px := 0; px < diameter; px++ {
			dx := float64(px) - cx
			dy := float64(py) - cy
			if dx*dx+dy*dy <= float64(radius*radius) {
				lightImg.Set(px, py, color.White)
			}
		}
	}
	lightBuffer.DrawImage(lightImg, opts)
}

// CalculateLightIntensityAt calculates the total light intensity at a point.
// This can be used for gameplay mechanics (e.g., stealth, vision).
func (s *LightingSystem) CalculateLightIntensityAt(x, y float64, entities []*Entity) float64 {
	if !s.config.Enabled {
		return 1.0 // Full brightness if lighting disabled
	}

	totalIntensity := s.config.AmbientIntensity

	// Check for ambient light component in entities first
	for _, entity := range entities {
		if ambComp, hasAmb := entity.GetComponent("ambient_light"); hasAmb {
			if ambient, ok := ambComp.(*AmbientLightComponent); ok {
				totalIntensity = ambient.Intensity
				break // Only use first ambient light found
			}
		}
	}

	// Check each light
	for _, entity := range entities {
		lightComp, hasLight := entity.GetComponent("light")
		if !hasLight {
			continue
		}

		light, ok := lightComp.(*LightComponent)
		if !ok || !light.Enabled {
			continue
		}

		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance
		dx := x - pos.X
		dy := y - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > light.Radius {
			continue
		}

		// Calculate falloff
		falloff := s.calculateFalloff(dist, light.Radius, light.Falloff)

		// Add light contribution
		intensity := light.GetCurrentIntensity() * falloff
		totalIntensity += intensity
	}

	// Clamp to [0, 1]
	if totalIntensity > 1.0 {
		totalIntensity = 1.0
	}

	return totalIntensity
}

// calculateFalloff computes light falloff based on distance and type.
func (s *LightingSystem) calculateFalloff(dist, radius float64, falloffType LightFalloffType) float64 {
	if dist >= radius {
		return 0
	}

	normalized := dist / radius

	switch falloffType {
	case FalloffLinear:
		return 1.0 - normalized

	case FalloffQuadratic:
		return 1.0 - normalized*normalized

	case FalloffInverseSquare:
		if dist < 1.0 {
			return 1.0
		}
		return 1.0 / (dist * dist) * (radius * radius)

	case FalloffConstant:
		return 1.0

	default:
		return 1.0 - normalized
	}
}

// SetEnabled enables or disables the lighting system.
func (s *LightingSystem) SetEnabled(enabled bool) {
	s.config.Enabled = enabled
	if s.logger != nil {
		s.logger.WithField("enabled", enabled).Info("lighting system toggled")
	}
}

// IsEnabled returns whether lighting is currently enabled.
func (s *LightingSystem) IsEnabled() bool {
	return s.config.Enabled
}

// GetConfig returns the current lighting configuration.
func (s *LightingSystem) GetConfig() *LightingConfig {
	return s.config
}

// SetConfig updates the lighting configuration.
func (s *LightingSystem) SetConfig(config *LightingConfig) {
	s.config = config
}
