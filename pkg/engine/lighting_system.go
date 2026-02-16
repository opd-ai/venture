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
	"github.com/opd-ai/venture/pkg/rendering/lighting"
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
	visibleLights []lightWithPosition

	// Lighting buffer (reused each frame)
	lightingBuffer *ebiten.Image

	// Cached ambient light entity (avoid O(n) search each frame)
	ambientLightEntityID uint64
	ambientLightCached   bool

	// Shadow system integration (optional)
	shadowSystem *ShadowSystem

	// Light circle image cache - keyed by (diameter, falloff) to avoid per-frame allocations
	// This dramatically reduces allocations in the hot applyPointLight path
	// Uses radial gradients for realistic light falloff
	lightCircleCache map[lightCacheKey]*ebiten.Image

	// Reusable DrawImageOptions to eliminate per-light allocations in hot path
	// Reset and reused for each point light in applyPointLight
	lightDrawOpts ebiten.DrawImageOptions

	// Reusable DrawImageOptions for ambient lighting to eliminate per-frame allocation
	ambientDrawOpts ebiten.DrawImageOptions

	// Reusable lightMetrics to eliminate per-frame allocation in CollectVisibleLights
	metrics lightMetrics

	// GPU bloom processor for hardware-accelerated bloom effects (V2 fix)
	// Uses Kage shaders instead of CPU pixel iteration, reducing bloom from 15-50ms to <2ms
	gpuBloom *lighting.GPUBloom

	// Tracked light entities for dirty-marked collection (V4 optimization).
	// Populated during Update() to avoid O(N_all_entities) scan in CollectVisibleLights.
	// When available, CollectVisibleLights iterates only these entities instead of all.
	trackedLightEntities []*Entity
	lightTrackingValid   bool
}

// lightWithPosition combines a light component with its world position.
type lightWithPosition struct {
	light *LightComponent
	x     float64
	y     float64
}

// lightCacheKey uniquely identifies a cached light circle image.
// This allows efficient reuse of gradient textures across frames.
type lightCacheKey struct {
	diameter int
	falloff  LightFalloffType
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
		world:                world,
		config:               config,
		logger:               logEntry,
		visibleLights:        make([]lightWithPosition, 0, config.MaxLights),
		lightCircleCache:     make(map[lightCacheKey]*ebiten.Image), // Initialize light circle cache to avoid per-frame allocations
		trackedLightEntities: make([]*Entity, 0, 16),
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
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.Debug("Updating shadow system viewport")
		}
		s.shadowSystem.SetViewport(cameraX, cameraY, width, height)
	}
}

// EnableShadows enables or disables shadow rendering.
func (s *LightingSystem) EnableShadows(enabled bool) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"has_shadow_system": s.shadowSystem != nil,
		}).Debug("Getting shadow system")
	}
	return s.shadowSystem
}

// Update processes lighting each frame (updates animation times).
func (s *LightingSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
			"enabled":      s.config.Enabled,
		}).Debug("Updating lighting system")
	}

	if !s.config.Enabled {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.Debug("Lighting system disabled, skipping update")
		}
		return
	}

	// Update animation times for flickering/pulsing lights
	// Also build tracked light entity list (V4 optimization) to avoid
	// O(N_all_entities) scan in CollectVisibleLights
	updatedLights := 0
	s.trackedLightEntities = s.trackedLightEntities[:0]
	for _, entity := range entities {
		comp, ok := entity.GetComponent("light")
		if !ok {
			continue
		}

		light, ok := comp.(*LightComponent)
		if !ok {
			continue
		}

		s.trackedLightEntities = append(s.trackedLightEntities, entity)

		previousTime := light.internalTime
		light.internalTime += deltaTime
		updatedLights++

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     entity.ID,
				"previous_time": previousTime,
				"new_time":      light.internalTime,
				"delta_time":    deltaTime,
			}).Debug("Updated light animation time")
		}
	}
	s.lightTrackingValid = true

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"updated_lights": updatedLights,
			"total_entities": len(entities),
			"delta_time":     deltaTime,
		}).Debug("Lighting system update complete")
	}
}

// CollectVisibleLights gathers lights within viewport for rendering.
// Returns the collected lights sorted by priority (closest first).
// Uses tracked light entities from Update() when available (V4 optimization),
// reducing iteration from O(N_all_entities) to O(N_light_entities).
func (s *LightingSystem) CollectVisibleLights(entities []*Entity) []lightWithPosition {
	startTime := time.Now()

	if !s.validateLightCollection(len(entities)) {
		return s.visibleLights
	}

	s.visibleLights = s.visibleLights[:0]
	metrics := s.initLightMetrics()

	// V4 optimization: use tracked light entities if available,
	// avoiding O(N) scan of all entities in the render hot path
	scanEntities := entities
	if s.lightTrackingValid && len(s.trackedLightEntities) > 0 {
		scanEntities = s.trackedLightEntities
	}

	for _, entity := range scanEntities {
		if len(s.visibleLights) >= s.config.MaxLights {
			s.logMaxLightsReached(metrics)
			break
		}

		light, pos, skip := s.extractLightAndPosition(entity, metrics)
		if skip {
			continue
		}

		if !s.isLightVisibleInViewport(entity, pos, light, metrics) {
			continue
		}

		s.addVisibleLight(entity, light, pos)
	}

	s.logLightCollectionComplete(metrics, len(entities), startTime)
	return s.visibleLights
}

// validateLightCollection performs initial validation for light collection.
func (s *LightingSystem) validateLightCollection(entityCount int) bool {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count":  entityCount,
			"enabled":       s.config.Enabled,
			"viewport_set":  s.viewportSet,
			"max_lights":    s.config.MaxLights,
			"previous_size": len(s.visibleLights),
		}).Debug("Collecting visible lights")
	}

	if !s.config.Enabled {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.Debug("Lighting disabled, returning empty light collection")
		}
		return false
	}
	return true
}

// lightMetrics tracks statistics during light collection.
// This struct is reused across frames to avoid allocations.
type lightMetrics struct {
	totalLights     int
	culledLights    int
	disabledLights  int
	missingPosition int
}

// reset clears all metrics for reuse in a new frame.
func (m *lightMetrics) reset() {
	m.totalLights = 0
	m.culledLights = 0
	m.disabledLights = 0
	m.missingPosition = 0
}

// initLightMetrics returns a pointer to the reusable metrics struct after resetting it.
func (s *LightingSystem) initLightMetrics() *lightMetrics {
	s.metrics.reset()
	return &s.metrics
}

// extractLightAndPosition extracts light and position components from entity.
func (s *LightingSystem) extractLightAndPosition(entity *Entity, metrics *lightMetrics) (*LightComponent, *PositionComponent, bool) {
	lightComp, hasLight := entity.GetComponent("light")
	if !hasLight {
		return nil, nil, true
	}

	light, ok := lightComp.(*LightComponent)
	if !ok || !light.Enabled {
		if !ok || !light.Enabled {
			metrics.disabledLights++
		}
		return nil, nil, true
	}

	metrics.totalLights++

	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		metrics.missingPosition++
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Warn("Light entity missing position component")
		}
		return nil, nil, true
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil, true
	}

	return light, pos, false
}

// isLightVisibleInViewport checks if light is visible in viewport.
func (s *LightingSystem) isLightVisibleInViewport(entity *Entity, pos *PositionComponent, light *LightComponent, metrics *lightMetrics) bool {
	if s.viewportSet && !s.isLightInViewport(pos.X, pos.Y, light.Radius) {
		metrics.culledLights++
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"x":         pos.X,
				"y":         pos.Y,
				"radius":    light.Radius,
			}).Debug("Light culled (outside viewport)")
		}
		return false
	}
	return true
}

// addVisibleLight adds a light to the visible lights collection.
func (s *LightingSystem) addVisibleLight(entity *Entity, light *LightComponent, pos *PositionComponent) {
	s.visibleLights = append(s.visibleLights, lightWithPosition{
		light: light,
		x:     pos.X,
		y:     pos.Y,
	})

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"x":         pos.X,
			"y":         pos.Y,
			"radius":    light.Radius,
			"intensity": light.GetCurrentIntensity(),
		}).Debug("Light added to visible collection")
	}
}

// logMaxLightsReached logs when maximum light limit is reached.
func (s *LightingSystem) logMaxLightsReached(metrics *lightMetrics) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"max_lights":     s.config.MaxLights,
			"total_lights":   metrics.totalLights,
			"culled_lights":  metrics.culledLights,
			"visible_lights": len(s.visibleLights),
		}).Warn("Reached maximum light limit")
	}
}

// logLightCollectionComplete logs the completion of light collection.
func (s *LightingSystem) logLightCollectionComplete(metrics *lightMetrics, entityCount int, startTime time.Time) {
	duration := time.Since(startTime)
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"total_lights":     metrics.totalLights,
			"culled_lights":    metrics.culledLights,
			"disabled_lights":  metrics.disabledLights,
			"missing_position": metrics.missingPosition,
			"visible_lights":   len(s.visibleLights),
			"duration_ms":      duration.Milliseconds(),
			"duration_us":      duration.Microseconds(),
			"entities_checked": entityCount,
			"viewport_culling": s.viewportSet,
		}).Debug("Light collection complete")
	}
}

// isLightInViewport checks if a light affects the viewport.
func (s *LightingSystem) isLightInViewport(x, y, radius float64) bool {
	// Expand viewport by light radius for overlap detection
	minX := s.cameraX - radius
	maxX := s.cameraX + float64(s.viewportW) + radius
	minY := s.cameraY - radius
	maxY := s.cameraY + float64(s.viewportH) + radius

	inViewport := x >= minX && x <= maxX && y >= minY && y <= maxY

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":          entityID,
			"previous_entity_id": s.ambientLightEntityID,
			"previous_cache":     s.ambientLightCached,
		}).Debug("Setting ambient light entity cache")
	}

	s.ambientLightEntityID = entityID
	s.ambientLightCached = true

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"cached":    s.ambientLightCached,
		}).Debug("Ambient light entity cache updated")
	}
}

// ClearAmbientLightCache clears the cached ambient light entity.
// Call this if the ambient light entity is removed from the world.
func (s *LightingSystem) ClearAmbientLightCache() {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"previous_entity_id": s.ambientLightEntityID,
			"was_cached":         s.ambientLightCached,
		}).Debug("Clearing ambient light cache")
	}

	s.ambientLightEntityID = 0
	s.ambientLightCached = false

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("Ambient light cache cleared")
	}
}

// MarkLightsDirty invalidates the tracked light entity list.
// Call this when light entities are added or removed outside of the normal Update cycle.
// The next CollectVisibleLights call will fall back to scanning all entities
// until the next Update() repopulates the tracked list.
func (s *LightingSystem) MarkLightsDirty() {
	s.lightTrackingValid = false
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("Light tracking invalidated")
	}
}

// ApplyLighting applies lighting effects to a rendered image.
// This is called after the main render pass as a post-processing step.
// Returns a new image with lighting applied (input image is not modified).
func (s *LightingSystem) ApplyLighting(screen, renderedScene *ebiten.Image, entities []*Entity) {
	startTime := time.Now()

	if !s.validateLightingSetup(renderedScene, len(entities)) {
		screen.DrawImage(renderedScene, nil)
		return
	}

	collectStart := time.Now()
	lights := s.CollectVisibleLights(entities)
	collectDuration := time.Since(collectStart)

	ambientIntensity, ambientColor, usedCachedAmbient := s.getCachedAmbientLight()

	if s.shouldSkipLighting(lights, ambientIntensity, startTime) {
		screen.DrawImage(renderedScene, nil)
		return
	}

	w, h := renderedScene.Size()
	bufferResized := s.createOrResizeLightingBuffer(w, h)

	lightsApplied := s.applyAmbientAndPointLights(renderedScene, lights, ambientIntensity, ambientColor, usedCachedAmbient, bufferResized, collectDuration)

	// Apply bloom effect if enabled
	if s.config.EnableBloom && s.config.BloomIntensity > 0 {
		bloomStart := time.Now()
		s.applyBloomEffect()
		bloomDuration := time.Since(bloomStart)
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"bloom_duration_ms": bloomDuration.Milliseconds(),
				"bloom_threshold":   s.config.BloomThreshold,
				"bloom_intensity":   s.config.BloomIntensity,
				"bloom_radius":      s.config.BloomRadius,
			}).Debug("Bloom effect applied")
		}
	}

	screen.DrawImage(s.lightingBuffer, nil)

	s.logLightingComplete(lightsApplied, startTime, collectDuration, ambientIntensity, bufferResized)
}

// validateLightingSetup performs initial validation and logging for lighting application.
func (s *LightingSystem) validateLightingSetup(renderedScene *ebiten.Image, entityCount int) bool {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		w, h := renderedScene.Size()
		s.logger.WithFields(logrus.Fields{
			"entity_count": entityCount,
			"scene_width":  w,
			"scene_height": h,
			"enabled":      s.config.Enabled,
		}).Debug("Applying lighting to scene")
	}

	if !s.config.Enabled {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.Debug("Lighting disabled, drawing scene directly")
		}
		return false
	}
	return true
}

// getCachedAmbientLight retrieves ambient light from cache or config defaults.
func (s *LightingSystem) getCachedAmbientLight() (float64, color.RGBA, bool) {
	ambientIntensity := s.config.AmbientIntensity
	ambientColor := s.config.AmbientColor
	usedCachedAmbient := false

	if s.ambientLightCached {
		if entity, ok := s.world.GetEntity(s.ambientLightEntityID); ok {
			if ambComp, ok := entity.GetComponent("ambient_light"); ok {
				if ambient, ok := ambComp.(*AmbientLightComponent); ok {
					ambientIntensity = ambient.Intensity
					ambientColor = ambient.Color
					usedCachedAmbient = true
					if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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

	return ambientIntensity, ambientColor, usedCachedAmbient
}

// shouldSkipLighting determines if lighting can be skipped for performance.
func (s *LightingSystem) shouldSkipLighting(lights []lightWithPosition, ambientIntensity float64, startTime time.Time) bool {
	if len(lights) == 0 && ambientIntensity > 0.8 {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"ambient_intensity": ambientIntensity,
				"light_count":       0,
			}).Debug("Skipping lighting (high ambient, no lights)")
		}
		duration := time.Since(startTime)
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"duration_ms": duration.Milliseconds(),
				"duration_us": duration.Microseconds(),
			}).Debug("Lighting skipped (fast path)")
		}
		return true
	}
	return false
}

// createOrResizeLightingBuffer creates or resizes the lighting buffer as needed.
func (s *LightingSystem) createOrResizeLightingBuffer(w, h int) bool {
	bufferResized := false
	if s.lightingBuffer == nil || s.lightingBuffer.Bounds().Dx() != w || s.lightingBuffer.Bounds().Dy() != h {
		previousW, previousH := 0, 0
		if s.lightingBuffer != nil {
			previousW = s.lightingBuffer.Bounds().Dx()
			previousH = s.lightingBuffer.Bounds().Dy()
		}
		s.lightingBuffer = ebiten.NewImage(w, h)
		bufferResized = true
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"width":      w,
				"height":     h,
				"previous_w": previousW,
				"previous_h": previousH,
			}).Debug("Created/resized lighting buffer")
		}
	}
	return bufferResized
}

// applyAmbientAndPointLights applies ambient lighting and point lights to the scene.
func (s *LightingSystem) applyAmbientAndPointLights(renderedScene *ebiten.Image, lights []lightWithPosition, ambientIntensity float64, ambientColor color.RGBA, usedCachedAmbient, bufferResized bool, collectDuration time.Duration) int {
	s.lightingBuffer.Clear()

	ambR := float64(ambientColor.R) / 255.0 * ambientIntensity
	ambG := float64(ambientColor.G) / 255.0 * ambientIntensity
	ambB := float64(ambientColor.B) / 255.0 * ambientIntensity

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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

	// OPTIMIZATION: Reuse pre-allocated DrawImageOptions to eliminate per-frame heap allocation
	// Previous: opts := &ebiten.DrawImageOptions{} - allocated 1 struct per frame
	// Now: Reuse s.ambientDrawOpts with Reset() - zero allocations in hot path
	s.ambientDrawOpts.GeoM.Reset()
	s.ambientDrawOpts.ColorScale.Reset()
	s.ambientDrawOpts.ColorScale.Scale(float32(ambR), float32(ambG), float32(ambB), 1.0)
	s.lightingBuffer.DrawImage(renderedScene, &s.ambientDrawOpts)

	lightsApplied := 0
	for i := range lights {
		s.applyPointLight(s.lightingBuffer, renderedScene, &lights[i])
		lightsApplied++
	}

	return lightsApplied
}

// logLightingComplete logs the completion of lighting application.
func (s *LightingSystem) logLightingComplete(lightsApplied int, startTime time.Time, collectDuration time.Duration, ambientIntensity float64, bufferResized bool) {
	duration := time.Since(startTime)
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"x":      lwp.x,
			"y":      lwp.y,
			"radius": lwp.light.Radius,
		}).Debug("Applying point light")
	}

	intensity := lwp.light.GetCurrentIntensity()
	if intensity <= 0 {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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

	// Draw light influence as additive blend with radial gradient falloff
	// OPTIMIZATION: Reuse cached DrawImageOptions to eliminate per-light heap allocations
	// Previous: opts := &ebiten.DrawImageOptions{} - allocated 1 struct per light per frame
	// Now: Reuse s.lightDrawOpts with Reset() - zero allocations in hot path
	s.lightDrawOpts.GeoM.Reset()
	s.lightDrawOpts.ColorScale.Reset()
	s.lightDrawOpts.GeoM.Translate(float64(x-radius), float64(y-radius))

	// Modulate by light color and intensity
	r := float64(lwp.light.Color.R) / 255.0 * intensity * 0.15 // 0.15 is blend strength (very subtle lighting)
	g := float64(lwp.light.Color.G) / 255.0 * intensity * 0.15
	b := float64(lwp.light.Color.B) / 255.0 * intensity * 0.15

	s.lightDrawOpts.ColorScale.Scale(float32(r), float32(g), float32(b), 1.0)
	s.lightDrawOpts.Blend = ebiten.BlendLighter // Additive blending

	// Generate radial gradient based on light's falloff curve
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

	// Use cached light circle with radial gradient to avoid per-frame allocations
	lightImg := s.getCachedLightCircle(diameter, lwp.light.Falloff)
	lightBuffer.DrawImage(lightImg, &s.lightDrawOpts)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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

// getCachedLightCircle returns a cached light circle image with radial gradient.
// This avoids per-frame allocations in the hot applyPointLight path.
// The gradient is calculated based on the falloff type for realistic light propagation.
func (s *LightingSystem) getCachedLightCircle(diameter int, falloff LightFalloffType) *ebiten.Image {
	cacheKey := lightCacheKey{diameter: diameter, falloff: falloff}
	if img, ok := s.lightCircleCache[cacheKey]; ok {
		return img
	}

	// Create new light circle image with radial gradient and cache it
	img := ebiten.NewImage(diameter, diameter)
	radius := float64(diameter) / 2.0
	cx, cy := radius, radius

	for py := 0; py < diameter; py++ {
		for px := 0; px < diameter; px++ {
			dx := float64(px) - cx
			dy := float64(py) - cy
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance <= radius {
				// Calculate intensity based on falloff type
				// normalized distance: 0.0 at center, 1.0 at edge
				normalizedDist := distance / radius
				intensity := s.calculateFalloffIntensity(normalizedDist, falloff)

				// Use white color with alpha based on intensity
				// The actual color will be applied via ColorScale at draw time
				alpha := uint8(intensity * 255.0)
				img.Set(px, py, color.RGBA{255, 255, 255, alpha})
			}
		}
	}

	s.lightCircleCache[cacheKey] = img
	return img
}

// calculateFalloffIntensity computes light intensity at a given normalized distance
// based on the falloff curve. Distance is normalized to [0, 1] where 0 is center, 1 is edge.
func (s *LightingSystem) calculateFalloffIntensity(normalizedDist float64, falloff LightFalloffType) float64 {
	// Clamp to valid range
	if normalizedDist <= 0 {
		return 1.0
	}
	if normalizedDist >= 1.0 {
		return 0.0
	}

	switch falloff {
	case FalloffLinear:
		// Linear falloff: intensity = 1 - distance
		return 1.0 - normalizedDist

	case FalloffQuadratic:
		// Quadratic falloff: intensity = (1 - distance)^2
		remaining := 1.0 - normalizedDist
		return remaining * remaining

	case FalloffInverseSquare:
		// Inverse square falloff: intensity = 1 / (1 + distance^2)
		// Modified to work with normalized distance and avoid division by zero
		return 1.0 / (1.0 + normalizedDist*normalizedDist*4.0)

	case FalloffConstant:
		// Constant falloff: full intensity until edge
		return 1.0

	default:
		// Default to linear if unknown falloff type
		return 1.0 - normalizedDist
	}
}

// findAmbientIntensity searches for ambient light component and returns its intensity.
func (s *LightingSystem) findAmbientIntensity(entities []*Entity) float64 {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count":      len(entities),
			"default_intensity": s.config.AmbientIntensity,
		}).Debug("Searching for ambient light component")
	}

	for _, entity := range entities {
		if ambComp, hasAmb := entity.GetComponent("ambient_light"); hasAmb {
			if ambient, ok := ambComp.(*AmbientLightComponent); ok {
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"entity_id":         entity.ID,
						"ambient_intensity": ambient.Intensity,
					}).Debug("Found ambient light component")
				}
				return ambient.Intensity
			}
		}
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"default_intensity": s.config.AmbientIntensity,
		}).Debug("No ambient light component found, using default")
	}
	return s.config.AmbientIntensity
}

// calculateLightContribution computes light contribution from a single entity.
func (s *LightingSystem) calculateLightContribution(entity *Entity, x, y float64) (float64, bool) {
	s.logCalculationStart(entity.ID, x, y)

	light := s.getLightComponent(entity)
	if light == nil {
		return 0, false
	}

	pos := s.getPositionComponent(entity)
	if pos == nil {
		return 0, false
	}

	dist := calculateLightDistance(x, y, pos.X, pos.Y)

	if !s.isWithinRange(entity.ID, dist, light.Radius) {
		return 0, false
	}

	intensity := s.computeIntensity(light, dist)
	s.logCalculationResult(entity.ID, dist, light.Radius, intensity)

	return intensity, true
}

// logCalculationStart logs the beginning of light contribution calculation.
func (s *LightingSystem) logCalculationStart(entityID uint64, x, y float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"target_x":  x,
			"target_y":  y,
		}).Debug("Calculating light contribution")
	}
}

// getLightComponent retrieves and validates the light component.
func (s *LightingSystem) getLightComponent(entity *Entity) *LightComponent {
	lightComp, hasLight := entity.GetComponent("light")
	if !hasLight {
		return nil
	}

	light, ok := lightComp.(*LightComponent)
	if !ok || !light.Enabled {
		s.logDisabledLight(entity.ID, light)
		return nil
	}

	return light
}

// logDisabledLight logs when a light component is disabled or invalid.
func (s *LightingSystem) logDisabledLight(entityID uint64, light *LightComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel && light != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"enabled":   light.Enabled,
		}).Debug("Light component disabled or invalid")
	}
}

// getPositionComponent retrieves and validates the position component.
func (s *LightingSystem) getPositionComponent(entity *Entity) *PositionComponent {
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		s.logMissingPosition(entity.ID)
		return nil
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil
	}

	return pos
}

// logMissingPosition logs when a light entity lacks a position component.
func (s *LightingSystem) logMissingPosition(entityID uint64) {
	if s.logger != nil {
		s.logger.WithField("entity_id", entityID).Warn("Light entity missing position component")
	}
}

// calculateLightDistance computes the Euclidean distance between two points.
func calculateLightDistance(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

// isWithinRange checks if the distance is within the light radius.
func (s *LightingSystem) isWithinRange(entityID uint64, dist, radius float64) bool {
	if dist > radius {
		s.logOutOfRange(entityID, dist, radius)
		return false
	}
	return true
}

// logOutOfRange logs when a light is out of range.
func (s *LightingSystem) logOutOfRange(entityID uint64, dist, radius float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"distance":  dist,
			"radius":    radius,
		}).Debug("Light out of range")
	}
}

// computeIntensity calculates the final light intensity with falloff.
func (s *LightingSystem) computeIntensity(light *LightComponent, dist float64) float64 {
	falloff := s.calculateFalloff(dist, light.Radius, light.Falloff)
	return light.GetCurrentIntensity() * falloff
}

// logCalculationResult logs the final light contribution calculation.
func (s *LightingSystem) logCalculationResult(entityID uint64, dist, radius, intensity float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		falloff := intensity / (intensity + 1) // Approximate falloff for logging
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"distance":  dist,
			"falloff":   falloff,
			"intensity": intensity,
			"radius":    radius,
		}).Debug("Light contribution calculated")
	}
}

// CalculateLightIntensityAt calculates the total light intensity at a point.
// This can be used for gameplay mechanics (e.g., stealth, vision).
func (s *LightingSystem) CalculateLightIntensityAt(x, y float64, entities []*Entity) float64 {
	startTime := time.Now()

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"x":            x,
			"y":            y,
			"entity_count": len(entities),
			"enabled":      s.config.Enabled,
		}).Debug("Calculating light intensity at point")
	}

	if !s.config.Enabled {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
// calculateFalloff computes light falloff at a given distance for gameplay calculations.
// This is used by CalculateLightIntensityAt for gameplay mechanics (stealth, vision, etc.).
// It delegates to calculateFalloffIntensity for consistent falloff behavior across
// both rendering (gradient generation) and gameplay (intensity calculation).
func (s *LightingSystem) calculateFalloff(dist, radius float64, falloffType LightFalloffType) float64 {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"distance":     dist,
			"radius":       radius,
			"falloff_type": falloffType,
		}).Debug("Calculating light falloff")
	}

	if dist >= radius {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"distance": dist,
				"radius":   radius,
				"falloff":  0.0,
			}).Debug("Distance exceeds radius, zero falloff")
		}
		return 0
	}

	// Normalize distance to [0, 1] and use the same falloff calculation as gradient generation
	normalizedDist := dist / radius
	falloff := s.calculateFalloffIntensity(normalizedDist, falloffType)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"distance":        dist,
			"radius":          radius,
			"normalized_dist": normalizedDist,
			"falloff":         falloff,
		}).Debug("Falloff calculated")
	}

	return falloff
}

// SetEnabled enables or disables the lighting system.
func (s *LightingSystem) SetEnabled(enabled bool) {
	previousState := s.config.Enabled
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("enabled", enabled).Debug("Checking if lighting is enabled")
	}
	return enabled
}

// GetConfig returns the current lighting configuration.
func (s *LightingSystem) GetConfig() *LightingConfig {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"max_lights":        s.config.MaxLights,
			"ambient_intensity": s.config.AmbientIntensity,
			"shadows_enabled":   s.config.ShadowsEnabled,
			"enabled":           s.config.Enabled,
		}).Debug("Getting lighting configuration")
	}
	return s.config
}

// applyBloomEffect applies bloom/glow effect to bright areas of the lighting buffer.
// This creates a soft glow around bright lights for enhanced visual quality.
// Uses GPU-accelerated Kage shaders for performance (<2ms vs 15-50ms CPU-based).
func (s *LightingSystem) applyBloomEffect() {
	if s.lightingBuffer == nil {
		return
	}

	// Lazy-initialize GPU bloom processor
	if s.gpuBloom == nil {
		s.gpuBloom = lighting.NewGPUBloom()
	}

	// Configure bloom from lighting config
	bloomConfig := lighting.BloomConfig{
		Enabled:   true,
		Threshold: s.config.BloomThreshold,
		Intensity: s.config.BloomIntensity,
		Radius:    s.config.BloomRadius,
		Samples:   7, // Not used by GPU bloom, kept for compatibility
	}
	s.gpuBloom.SetConfig(bloomConfig)

	// Apply GPU-accelerated bloom directly to the lighting buffer
	s.gpuBloom.ApplyToBuffer(s.lightingBuffer)
}

// GetConfig returns the current lighting configuration.

// SetConfig updates the lighting configuration.
func (s *LightingSystem) SetConfig(config *LightingConfig) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
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
