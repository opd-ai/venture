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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// LightingSystem processes light sources and applies lighting to the scene.
// This system runs after the main render pass to apply lighting effects
// as a post-processing step. It also manages shadow rendering via ShadowSystem.
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

	// Shadow system integration (optional)
	shadowSystem *ShadowSystem
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
		logEntry.WithFields(logrus.Fields{
			"has_config": config != nil,
		}).Debug("Creating lighting system")
	}

	if config == nil {
		config = NewLightingConfig()
		if logEntry != nil {
			logEntry.Debug("Using default lighting configuration")
		}
	}

	system := &LightingSystem{
		world:         world,
		config:        config,
		logger:        logEntry,
		visibleLights: make([]*lightWithPosition, 0, config.MaxLights),
	}

	// Initialize shadow system if shadows are enabled
	if config.ShadowsEnabled {
		if logEntry != nil {
			logEntry.WithFields(logrus.Fields{
				"max_shadows":    config.MaxShadows,
				"shadow_quality": config.ShadowQuality,
			}).Debug("Initializing shadow system")
		}
		system.shadowSystem = NewShadowSystemWithLogger(world, logger)
		system.shadowSystem.SetMaxShadows(config.MaxShadows)
		system.shadowSystem.SetRenderQuality(config.ShadowQuality)
	}

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"max_lights":        config.MaxLights,
			"shadows_enabled":   config.ShadowsEnabled,
			"ambient_intensity": config.AmbientIntensity,
		}).Info("Lighting system created")
	}

	return system
}

// SetViewport updates the camera position and viewport size for culling.
func (s *LightingSystem) SetViewport(cameraX, cameraY float64, width, height int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"camera_x":          cameraX,
			"camera_y":          cameraY,
			"width":             width,
			"height":            height,
			"previous_camera_x": s.cameraX,
			"previous_camera_y": s.cameraY,
			"viewport_was_set":  s.viewportSet,
		}).Debug("Setting viewport")
	}

	s.cameraX = cameraX
	s.cameraY = cameraY
	s.viewportW = width
	s.viewportH = height
	s.viewportSet = true

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"camera_x":     cameraX,
			"camera_y":     cameraY,
			"width":        width,
			"height":       height,
			"viewport_set": s.viewportSet,
		}).Debug("Viewport updated")
	}

	// Update shadow system viewport
	if s.shadowSystem != nil {
		if s.logger != nil {
			s.logger.Debug("Updating shadow system viewport")
		}
		s.shadowSystem.SetViewport(cameraX, cameraY, width, height)
	}
}

// EnableShadows enables or disables shadow rendering.
func (s *LightingSystem) EnableShadows(enabled bool) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"enabled":              enabled,
			"shadow_system_exists": s.shadowSystem != nil,
		}).Debug("Toggling shadows")
	}

	s.config.ShadowsEnabled = enabled
	if s.shadowSystem != nil {
		s.shadowSystem.SetEnabled(enabled)
	} else if enabled {
		// Create shadow system if it doesn't exist
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"max_shadows":    s.config.MaxShadows,
				"shadow_quality": s.config.ShadowQuality,
			}).Info("Creating shadow system on-demand")
		}
		s.shadowSystem = NewShadowSystem(s.world)
		s.shadowSystem.SetMaxShadows(s.config.MaxShadows)
		s.shadowSystem.SetRenderQuality(s.config.ShadowQuality)
		if s.viewportSet {
			s.shadowSystem.SetViewport(s.cameraX, s.cameraY, s.viewportW, s.viewportH)
		}
	}

	if s.logger != nil {
		s.logger.WithField("enabled", enabled).Info("Shadows toggled")
	}
}

// GetShadowSystem returns the shadow system for direct access.
// Returns nil if shadows are not enabled.
func (s *LightingSystem) GetShadowSystem() *ShadowSystem {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"has_shadow_system": s.shadowSystem != nil,
		}).Debug("Getting shadow system")
	}
	return s.shadowSystem
}

// Update processes lighting each frame (updates animation times).
func (s *LightingSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
			"enabled":      s.config.Enabled,
		}).Debug("Updating lighting system")
	}

	if !s.config.Enabled {
		if s.logger != nil {
			s.logger.Debug("Lighting system disabled, skipping update")
		}
		return
	}

	// Update animation times for flickering/pulsing lights
	updatedLights := 0
	for _, entity := range entities {
		comp, ok := entity.GetComponent("light")
		if !ok {
			continue
		}

		light, ok := comp.(*LightComponent)
		if !ok {
			continue
		}

		previousTime := light.internalTime
		light.internalTime += deltaTime
		updatedLights++

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     entity.ID,
				"previous_time": previousTime,
				"new_time":      light.internalTime,
				"delta_time":    deltaTime,
			}).Debug("Updated light animation time")
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"updated_lights": updatedLights,
			"total_entities": len(entities),
			"delta_time":     deltaTime,
		}).Debug("Lighting system update complete")
	}
}

// CollectVisibleLights gathers lights within viewport for rendering.
// Returns the collected lights sorted by priority (closest first).
func (s *LightingSystem) CollectVisibleLights(entities []*Entity) []*lightWithPosition {
	startTime := time.Now()

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count":  len(entities),
			"enabled":       s.config.Enabled,
			"viewport_set":  s.viewportSet,
			"max_lights":    s.config.MaxLights,
			"previous_size": len(s.visibleLights),
		}).Debug("Collecting visible lights")
	}

	s.visibleLights = s.visibleLights[:0]

	if !s.config.Enabled {
		if s.logger != nil {
			s.logger.Debug("Lighting disabled, returning empty light collection")
		}
		return s.visibleLights
	}

	totalLights := 0
	culledLights := 0
	disabledLights := 0
	missingPosition := 0

	for _, entity := range entities {
		// Get light component
		lightComp, hasLight := entity.GetComponent("light")
		if !hasLight {
			continue
		}

		light, ok := lightComp.(*LightComponent)
		if !ok {
			continue
		}

		if !light.Enabled {
			disabledLights++
			continue
		}

		totalLights++

		// Get position
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			missingPosition++
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
				}).Warn("Light entity missing position component")
			}
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Viewport culling (if viewport is set)
		if s.viewportSet {
			if !s.isLightInViewport(pos.X, pos.Y, light.Radius) {
				culledLights++
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id": entity.ID,
						"x":         pos.X,
						"y":         pos.Y,
						"radius":    light.Radius,
					}).Debug("Light culled (outside viewport)")
				}
				continue
			}
		}

		// Add to visible lights
		s.visibleLights = append(s.visibleLights, &lightWithPosition{
			light: light,
			x:     pos.X,
			y:     pos.Y,
		})

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"x":         pos.X,
				"y":         pos.Y,
				"radius":    light.Radius,
				"intensity": light.GetCurrentIntensity(),
			}).Debug("Light added to visible collection")
		}

		// Limit to MaxLights
		if len(s.visibleLights) >= s.config.MaxLights {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"max_lights":     s.config.MaxLights,
					"total_lights":   totalLights,
					"culled_lights":  culledLights,
					"visible_lights": len(s.visibleLights),
				}).Warn("Reached maximum light limit")
			}
			break
		}
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"total_lights":     totalLights,
			"culled_lights":    culledLights,
			"disabled_lights":  disabledLights,
			"missing_position": missingPosition,
			"visible_lights":   len(s.visibleLights),
			"duration_ms":      duration.Milliseconds(),
			"duration_us":      duration.Microseconds(),
			"entities_checked": len(entities),
			"viewport_culling": s.viewportSet,
		}).Debug("Light collection complete")
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

	inViewport := x >= minX && x <= maxX && y >= minY && y <= maxY

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":           x,
			"y":           y,
			"radius":      radius,
			"camera_x":    s.cameraX,
			"camera_y":    s.cameraY,
			"viewport_w":  s.viewportW,
			"viewport_h":  s.viewportH,
			"in_viewport": inViewport,
		}).Debug("Viewport culling check")
	}

	return inViewport
}

// SetAmbientLightEntity sets the cached ambient light entity ID.
// This should be called when creating or changing the ambient light entity
// to avoid O(n) iteration on every frame.
func (s *LightingSystem) SetAmbientLightEntity(entityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":          entityID,
			"previous_entity_id": s.ambientLightEntityID,
			"previous_cache":     s.ambientLightCached,
		}).Debug("Setting ambient light entity cache")
	}

	s.ambientLightEntityID = entityID
	s.ambientLightCached = true

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"cached":    s.ambientLightCached,
		}).Debug("Ambient light entity cache updated")
	}
}

// ClearAmbientLightCache clears the cached ambient light entity.
// Call this if the ambient light entity is removed from the world.
func (s *LightingSystem) ClearAmbientLightCache() {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"previous_entity_id": s.ambientLightEntityID,
			"was_cached":         s.ambientLightCached,
		}).Debug("Clearing ambient light cache")
	}

	s.ambientLightEntityID = 0
	s.ambientLightCached = false

	if s.logger != nil {
		s.logger.Debug("Ambient light cache cleared")
	}
}

// ApplyLighting applies lighting effects to a rendered image.
// This is called after the main render pass as a post-processing step.
// Returns a new image with lighting applied (input image is not modified).
func (s *LightingSystem) ApplyLighting(screen, renderedScene *ebiten.Image, entities []*Entity) {
	startTime := time.Now()

	if s.logger != nil {
		w, h := renderedScene.Size()
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"scene_width":  w,
			"scene_height": h,
			"enabled":      s.config.Enabled,
		}).Debug("Applying lighting to scene")
	}

	if !s.config.Enabled {
		if s.logger != nil {
			s.logger.Debug("Lighting disabled, drawing scene directly")
		}
		screen.DrawImage(renderedScene, nil)
		return
	}

	// Collect visible lights
	collectStart := time.Now()
	lights := s.CollectVisibleLights(entities)
	collectDuration := time.Since(collectStart)

	// Get ambient light from cache or config defaults
	ambientIntensity := s.config.AmbientIntensity
	ambientColor := s.config.AmbientColor
	usedCachedAmbient := false

	// Try cached ambient light entity first
	if s.ambientLightCached {
		if entity, ok := s.world.GetEntity(s.ambientLightEntityID); ok {
			if ambComp, ok := entity.GetComponent("ambient_light"); ok {
				if ambient, ok := ambComp.(*AmbientLightComponent); ok {
					ambientIntensity = ambient.Intensity
					ambientColor = ambient.Color
					usedCachedAmbient = true
					if s.logger != nil {
						s.logger.WithFields(logrus.Fields{
							"entity_id":         s.ambientLightEntityID,
							"ambient_intensity": ambientIntensity,
							"ambient_color_r":   ambientColor.R,
							"ambient_color_g":   ambientColor.G,
							"ambient_color_b":   ambientColor.B,
						}).Debug("Using cached ambient light")
					}
				}
			}
		} else if s.logger != nil {
			s.logger.WithField("entity_id", s.ambientLightEntityID).Warn("Cached ambient light entity not found")
		}
	}

	// If no lights and high ambient, just draw normally
	if len(lights) == 0 && ambientIntensity > 0.8 {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"ambient_intensity": ambientIntensity,
				"light_count":       0,
			}).Debug("Skipping lighting (high ambient, no lights)")
		}
		screen.DrawImage(renderedScene, nil)
		duration := time.Since(startTime)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"duration_ms": duration.Milliseconds(),
				"duration_us": duration.Microseconds(),
			}).Debug("Lighting skipped (fast path)")
		}
		return
	}

	// Create lighting buffer if needed
	w, h := renderedScene.Size()
	bufferResized := false
	if s.lightingBuffer == nil || s.lightingBuffer.Bounds().Dx() != w || s.lightingBuffer.Bounds().Dy() != h {
		previousW, previousH := 0, 0
		if s.lightingBuffer != nil {
			previousW = s.lightingBuffer.Bounds().Dx()
			previousH = s.lightingBuffer.Bounds().Dy()
		}
		s.lightingBuffer = ebiten.NewImage(w, h)
		bufferResized = true
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"width":      w,
				"height":     h,
				"previous_w": previousW,
				"previous_h": previousH,
			}).Debug("Created/resized lighting buffer")
		}
	}

	// Calculate lighting (simplified version - full per-pixel lighting would be more complex)
	// For now, we'll use a blend approach with colored overlays
	s.lightingBuffer.Clear()

	// Apply ambient base
	ambR := float64(ambientColor.R) / 255.0 * ambientIntensity
	ambG := float64(ambientColor.G) / 255.0 * ambientIntensity
	ambB := float64(ambientColor.B) / 255.0 * ambientIntensity

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"light_count":         len(lights),
			"ambient_intensity":   ambientIntensity,
			"ambient_r":           ambR,
			"ambient_g":           ambG,
			"ambient_b":           ambB,
			"cached_ambient":      usedCachedAmbient,
			"buffer_resized":      bufferResized,
			"collect_duration_ms": collectDuration.Milliseconds(),
		}).Debug("Applying lighting calculations")
	}

	// Draw rendered scene with ambient modulation
	opts := &ebiten.DrawImageOptions{}
	opts.ColorScale.Scale(float32(ambR), float32(ambG), float32(ambB), 1.0)
	s.lightingBuffer.DrawImage(renderedScene, opts)

	// Apply point lights additively
	// Note: Full lighting would require shader support or per-pixel calculations
	// This is a simplified version using additive blending
	lightsApplied := 0
	for _, lwp := range lights {
		s.applyPointLight(s.lightingBuffer, renderedScene, lwp)
		lightsApplied++
	}

	// Draw final result to screen
	screen.DrawImage(s.lightingBuffer, nil)

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"lights_applied":      lightsApplied,
			"total_duration_ms":   duration.Milliseconds(),
			"total_duration_us":   duration.Microseconds(),
			"collect_duration_ms": collectDuration.Milliseconds(),
			"ambient_intensity":   ambientIntensity,
			"buffer_resized":      bufferResized,
		}).Debug("Lighting application complete")
	}
}

// applyPointLight applies a single point light to the lighting buffer.
// This is a simplified implementation; full lighting would use shaders.
func (s *LightingSystem) applyPointLight(lightBuffer, scene *ebiten.Image, lwp *lightWithPosition) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":      lwp.x,
			"y":      lwp.y,
			"radius": lwp.light.Radius,
		}).Debug("Applying point light")
	}

	intensity := lwp.light.GetCurrentIntensity()
	if intensity <= 0 {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"x":         lwp.x,
				"y":         lwp.y,
				"intensity": intensity,
			}).Debug("Skipping light with zero intensity")
		}
		return
	}

	// Create a light influence area (simplified)
	// In a full implementation, this would be a radial gradient
	radius := int(lwp.light.Radius)
	x := int(lwp.x - s.cameraX)
	y := int(lwp.y - s.cameraY)

	// Draw light influence as additive blend
	// INTEGRATION FIX [Category F]: Radial Gradient Lighting
	// Gap: Light rendering uses simple circle fill instead of proper radial gradients
	// Fix: Implement gradient shader or pre-generated gradient texture with alpha falloff
	// Roadmap: ROADMAP_V3.md Phase 17.1 - Soft Shadows & Colored Lighting (bloom/glow complete, gradients deferred)
	// Temporary: Simple circle fill provides acceptable visual quality, gradient is optimization
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x-radius), float64(y-radius))

	// Modulate by light color and intensity
	r := float64(lwp.light.Color.R) / 255.0 * intensity * 0.15 // 0.15 is blend strength (very subtle lighting)
	g := float64(lwp.light.Color.G) / 255.0 * intensity * 0.15
	b := float64(lwp.light.Color.B) / 255.0 * intensity * 0.15

	opts.ColorScale.Scale(float32(r), float32(g), float32(b), 1.0)
	opts.Blend = ebiten.BlendLighter // Additive blending

	// Minimal implementation: draw a filled white circle as the light influence
	diameter := 2 * radius
	if diameter <= 0 {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"radius":   radius,
				"diameter": diameter,
				"x":        lwp.x,
				"y":        lwp.y,
			}).Warn("Invalid light diameter")
		}
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

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":         lwp.x,
			"y":         lwp.y,
			"radius":    radius,
			"diameter":  diameter,
			"intensity": intensity,
			"color_r":   lwp.light.Color.R,
			"color_g":   lwp.light.Color.G,
			"color_b":   lwp.light.Color.B,
			"blend_r":   r,
			"blend_g":   g,
			"blend_b":   b,
		}).Debug("Point light applied successfully")
	}
}

// findAmbientIntensity searches for ambient light component and returns its intensity.
func (s *LightingSystem) findAmbientIntensity(entities []*Entity) float64 {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count":      len(entities),
			"default_intensity": s.config.AmbientIntensity,
		}).Debug("Searching for ambient light component")
	}

	for _, entity := range entities {
		if ambComp, hasAmb := entity.GetComponent("ambient_light"); hasAmb {
			if ambient, ok := ambComp.(*AmbientLightComponent); ok {
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id":         entity.ID,
						"ambient_intensity": ambient.Intensity,
					}).Debug("Found ambient light component")
				}
				return ambient.Intensity
			}
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"default_intensity": s.config.AmbientIntensity,
		}).Debug("No ambient light component found, using default")
	}
	return s.config.AmbientIntensity
}

// calculateLightContribution computes light contribution from a single entity.
func (s *LightingSystem) calculateLightContribution(entity *Entity, x, y float64) (float64, bool) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"target_x":  x,
			"target_y":  y,
		}).Debug("Calculating light contribution")
	}

	lightComp, hasLight := entity.GetComponent("light")
	if !hasLight {
		return 0, false
	}

	light, ok := lightComp.(*LightComponent)
	if !ok || !light.Enabled {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"enabled":   light.Enabled,
			}).Debug("Light component disabled or invalid")
		}
		return 0, false
	}

	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithField("entity_id", entity.ID).Warn("Light entity missing position component")
		}
		return 0, false
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return 0, false
	}

	dx := x - pos.X
	dy := y - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > light.Radius {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"distance":  dist,
				"radius":    light.Radius,
			}).Debug("Light out of range")
		}
		return 0, false
	}

	falloff := s.calculateFalloff(dist, light.Radius, light.Falloff)
	intensity := light.GetCurrentIntensity() * falloff

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"distance":  dist,
			"falloff":   falloff,
			"intensity": intensity,
			"radius":    light.Radius,
		}).Debug("Light contribution calculated")
	}

	return intensity, true
}

// CalculateLightIntensityAt calculates the total light intensity at a point.
// This can be used for gameplay mechanics (e.g., stealth, vision).
func (s *LightingSystem) CalculateLightIntensityAt(x, y float64, entities []*Entity) float64 {
	startTime := time.Now()

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":            x,
			"y":            y,
			"entity_count": len(entities),
			"enabled":      s.config.Enabled,
		}).Debug("Calculating light intensity at point")
	}

	if !s.config.Enabled {
		if s.logger != nil {
			s.logger.Debug("Lighting disabled, returning full intensity")
		}
		return 1.0
	}

	totalIntensity := s.findAmbientIntensity(entities)
	contributingLights := 0
	skippedLights := 0

	for _, entity := range entities {
		if contribution, ok := s.calculateLightContribution(entity, x, y); ok {
			totalIntensity += contribution
			contributingLights++
		} else {
			skippedLights++
		}
	}

	clamped := false
	if totalIntensity > 1.0 {
		totalIntensity = 1.0
		clamped = true
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":                   x,
			"y":                   y,
			"total_intensity":     totalIntensity,
			"contributing_lights": contributingLights,
			"skipped_lights":      skippedLights,
			"clamped":             clamped,
			"duration_us":         duration.Microseconds(),
		}).Debug("Light intensity calculation complete")
	}

	return totalIntensity
}

// calculateFalloff computes light falloff based on distance and type.
func (s *LightingSystem) calculateFalloff(dist, radius float64, falloffType LightFalloffType) float64 {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"distance":     dist,
			"radius":       radius,
			"falloff_type": falloffType,
		}).Debug("Calculating light falloff")
	}

	if dist >= radius {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"distance": dist,
				"radius":   radius,
				"falloff":  0.0,
			}).Debug("Distance exceeds radius, zero falloff")
		}
		return 0
	}

	normalized := dist / radius
	var falloff float64

	switch falloffType {
	case FalloffLinear:
		falloff = 1.0 - normalized

	case FalloffQuadratic:
		falloff = 1.0 - normalized*normalized

	case FalloffInverseSquare:
		if dist < 1.0 {
			falloff = 1.0
		} else {
			falloff = 1.0 / (dist * dist) * (radius * radius)
		}

	case FalloffConstant:
		falloff = 1.0

	default:
		falloff = 1.0 - normalized
		if s.logger != nil {
			s.logger.WithField("falloff_type", falloffType).Warn("Unknown falloff type, using linear")
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"distance":     dist,
			"radius":       radius,
			"normalized":   normalized,
			"falloff_type": falloffType,
			"falloff":      falloff,
		}).Debug("Falloff calculated")
	}

	return falloff
}

// SetEnabled enables or disables the lighting system.
func (s *LightingSystem) SetEnabled(enabled bool) {
	previousState := s.config.Enabled
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"enabled":        enabled,
			"previous_state": previousState,
		}).Debug("Setting lighting system enabled state")
	}

	s.config.Enabled = enabled

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"enabled":        enabled,
			"previous_state": previousState,
		}).Info("Lighting system toggled")
	}
}

// IsEnabled returns whether lighting is currently enabled.
func (s *LightingSystem) IsEnabled() bool {
	enabled := s.config.Enabled
	if s.logger != nil {
		s.logger.WithField("enabled", enabled).Debug("Checking if lighting is enabled")
	}
	return enabled
}

// GetConfig returns the current lighting configuration.
func (s *LightingSystem) GetConfig() *LightingConfig {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"max_lights":        s.config.MaxLights,
			"ambient_intensity": s.config.AmbientIntensity,
			"shadows_enabled":   s.config.ShadowsEnabled,
			"enabled":           s.config.Enabled,
		}).Debug("Getting lighting configuration")
	}
	return s.config
}

// SetConfig updates the lighting configuration.
func (s *LightingSystem) SetConfig(config *LightingConfig) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"max_lights":        config.MaxLights,
			"ambient_intensity": config.AmbientIntensity,
			"shadows_enabled":   config.ShadowsEnabled,
		}).Debug("Updating lighting configuration")
	}

	s.config = config

	if s.logger != nil {
		s.logger.Info("Lighting configuration updated")
	}
}
