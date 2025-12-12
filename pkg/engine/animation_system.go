// Package engine provides animation system for updating entity animations.
package engine

import (
	"fmt"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// AnimationSystem updates animation components and manages frame transitions.
// Integrates with sprite generator to create procedural animation frames.
type AnimationSystem struct {
	spriteGenerator *sprites.Generator
	spriteCache     *cache.SpriteCache // Phase 1.2: External sprite cache for base sprites
	frameCache      map[string][]*ebiten.Image // Cache by key: seed_state
	cacheMutex      sync.RWMutex
	maxCacheSize    int
	cacheKeys       []string // For LRU eviction
	logger          *logrus.Entry

	// Phase 14.2: Viewport culling and distance-based optimization
	cameraSystem        *CameraSystem
	enableViewportCull  bool
	enableDistanceLOD   bool
	playerEntity        *Entity        // Cached player reference for distance calculations
	distanceCloseThresh float64        // Distance threshold for full animation (default: 200px)
	distanceMidThresh   float64        // Distance threshold for half rate (default: 400px)
	stats               AnimationStats // Performance statistics
}

// AnimationStats holds performance statistics for the animation system.
type AnimationStats struct {
	TotalEntities    int // Total entities processed
	AnimatedEntities int // Entities with active animations
	CulledByViewport int // Entities culled by viewport check
	FullRateEntities int // Entities animated at full rate (close)
	HalfRateEntities int // Entities animated at half rate (mid distance)
	StaticEntities   int // Entities rendered as static (far distance)
}

// NewAnimationSystem creates a new animation system.
func NewAnimationSystem(spriteGenerator *sprites.Generator) *AnimationSystem {
	return NewAnimationSystemWithLogger(spriteGenerator, nil)
}

// NewAnimationSystemWithLogger creates a new animation system with a logger.
func NewAnimationSystemWithLogger(spriteGenerator *sprites.Generator, logger *logrus.Logger) *AnimationSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "animation",
		})
		logEntry.WithFields(logrus.Fields{
			"max_cache_size":        100,
			"viewport_cull_enabled": true,
			"distance_lod_enabled":  true,
			"close_threshold":       200.0,
			"mid_threshold":         400.0,
		}).Debug("initializing animation system")
	}

	return &AnimationSystem{
		spriteGenerator: spriteGenerator,
		frameCache:      make(map[string][]*ebiten.Image),
		maxCacheSize:    100, // Cache up to 100 animation sequences
		cacheKeys:       make([]string, 0, 100),
		logger:          logEntry,
		// Phase 14.2: Default optimization settings
		enableViewportCull:  true,  // Enabled by default for performance
		enableDistanceLOD:   true,  // Enabled by default for performance
		distanceCloseThresh: 200.0, // Full animation within 200px
		distanceMidThresh:   400.0, // Half rate 200-400px, static beyond
	}
}

// Phase 14.2: Configuration methods for viewport culling and distance-based LOD

// SetCameraSystem sets the camera system for viewport culling.
// Call this during initialization to enable viewport-based optimization.
func (s *AnimationSystem) SetCameraSystem(cameraSystem *CameraSystem) {
	s.cameraSystem = cameraSystem
	if s.logger != nil {
		s.logger.Debug("camera system set for viewport culling")
	}
}

// SetPlayerEntity sets the player entity reference for distance calculations.
// Call this during initialization to enable distance-based frame rate adjustment.
func (s *AnimationSystem) SetPlayerEntity(player *Entity) {
	s.playerEntity = player
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": player.ID,
		}).Debug("player entity set for distance calculations")
	}
}

// SetSpriteCache sets the external sprite cache for base sprite caching.
// Phase 1.2: Integrates with pkg/rendering/cache for efficient sprite reuse.
// When set, base sprites are cached before animation frame transformation,
// significantly reducing regeneration overhead for repeated sprite types.
func (s *AnimationSystem) SetSpriteCache(spriteCache *cache.SpriteCache) {
	s.spriteCache = spriteCache
	if s.logger != nil {
		if spriteCache != nil {
			s.logger.WithFields(logrus.Fields{
				"max_size": spriteCache.MaxSize(),
			}).Info("sprite cache connected to animation system")
		} else {
			s.logger.Debug("sprite cache disconnected from animation system")
		}
	}
}

// GetSpriteCache returns the current sprite cache, or nil if not set.
func (s *AnimationSystem) GetSpriteCache() *cache.SpriteCache {
	return s.spriteCache
}

// EnableViewportCulling enables or disables viewport culling optimization.
// When enabled, only animates entities visible in the current viewport.
func (s *AnimationSystem) EnableViewportCulling(enable bool) {
	s.enableViewportCull = enable
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"enabled": enable,
		}).Debug("viewport culling configuration changed")
	}
}

// EnableDistanceLOD enables or disables distance-based level-of-detail.
// When enabled, adjusts animation frame rate based on distance from player.
func (s *AnimationSystem) EnableDistanceLOD(enable bool) {
	s.enableDistanceLOD = enable
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"enabled": enable,
		}).Debug("distance LOD configuration changed")
	}
}

// SetDistanceThresholds sets the distance thresholds for LOD tiers.
// closeThreshold: Full animation rate (default 200px)
// midThreshold: Half animation rate (default 400px)
// Beyond midThreshold: Static pose (no animation updates)
func (s *AnimationSystem) SetDistanceThresholds(closeThreshold, midThreshold float64) {
	s.distanceCloseThresh = closeThreshold
	s.distanceMidThresh = midThreshold
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"close_threshold": closeThreshold,
			"mid_threshold":   midThreshold,
		}).Debug("distance thresholds updated")
	}
}

// SetMaxCacheSize sets the maximum number of animation sequences to cache.
// Larger cache reduces regeneration overhead but uses more memory.
// Default is 100. For WASM/browser environments, consider 200-500 for better performance.
// Each cached sequence typically uses 100-400KB depending on sprite size and frame count.
func (s *AnimationSystem) SetMaxCacheSize(maxSize int) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	oldSize := s.maxCacheSize
	s.maxCacheSize = maxSize

	// If new size is smaller than current cache, trigger eviction
	if len(s.frameCache) > maxSize {
		// Evict oldest entries until within limit
		toEvict := len(s.frameCache) - maxSize
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"old_max_size":    oldSize,
				"new_max_size":    maxSize,
				"current_size":    len(s.frameCache),
				"entries_evicted": toEvict,
			}).Debug("cache size reduced, evicting entries")
		}
		for i := 0; i < toEvict && len(s.cacheKeys) > 0; i++ {
			oldestKey := s.cacheKeys[0]
			delete(s.frameCache, oldestKey)
			s.cacheKeys = s.cacheKeys[1:]
		}
	} else if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"old_max_size": oldSize,
			"new_max_size": maxSize,
			"current_size": len(s.frameCache),
		}).Debug("max cache size updated")
	}
}

// GetStats returns current animation performance statistics.
// Useful for monitoring and debugging performance.
func (s *AnimationSystem) GetStats() AnimationStats {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"total_entities":    s.stats.TotalEntities,
			"animated_entities": s.stats.AnimatedEntities,
			"culled_viewport":   s.stats.CulledByViewport,
			"full_rate":         s.stats.FullRateEntities,
			"half_rate":         s.stats.HalfRateEntities,
			"static":            s.stats.StaticEntities,
		}).Debug("retrieving animation stats")
	}
	return s.stats
}

// Update processes all entities with animation components.
// Updates frame timers, transitions states, and regenerates frames if needed.
func (s *AnimationSystem) Update(entities []*Entity, deltaTime float64) error {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("animation system update started")
	}

	s.resetStatistics(len(entities))
	playerX, playerY := s.getPlayerPosition()
	viewportBounds, hasViewport := s.calculateViewportBounds()

	for _, entity := range entities {
		if err := s.updateEntityAnimation(entity, deltaTime, playerX, playerY, viewportBounds, hasViewport); err != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"error":     err.Error(),
				}).Error("failed to update entity animation")
			}
			return err
		}
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"total_entities":    s.stats.TotalEntities,
			"animated_entities": s.stats.AnimatedEntities,
			"culled_viewport":   s.stats.CulledByViewport,
			"full_rate":         s.stats.FullRateEntities,
			"half_rate":         s.stats.HalfRateEntities,
			"static":            s.stats.StaticEntities,
		}).Debug("animation system update completed")
	}

	return nil
}

// resetStatistics resets animation statistics for the current frame.
func (s *AnimationSystem) resetStatistics(entityCount int) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": entityCount,
		}).Debug("resetting animation statistics")
	}
	s.stats = AnimationStats{
		TotalEntities: entityCount,
	}
}

// getPlayerPosition retrieves the current player position for distance calculations.
func (s *AnimationSystem) getPlayerPosition() (float64, float64) {
	var playerX, playerY float64
	if s.playerEntity != nil {
		if posComp, ok := s.playerEntity.GetComponent("position"); ok {
			if pos, ok := posComp.(*PositionComponent); ok {
				playerX = pos.X
				playerY = pos.Y
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"player_x": playerX,
						"player_y": playerY,
					}).Debug("retrieved player position")
				}
			}
		}
	}
	return playerX, playerY
}

// viewportBounds holds calculated viewport boundaries for culling.
type viewportBounds struct {
	minX, minY, maxX, maxY float64
}

// calculateViewportBounds computes viewport boundaries for culling optimization.
func (s *AnimationSystem) calculateViewportBounds() (viewportBounds, bool) {
	var bounds viewportBounds
	if !s.enableViewportCull || s.cameraSystem == nil || s.cameraSystem.activeCamera == nil {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.Debug("viewport culling disabled or camera not available")
		}
		return bounds, false
	}

	camComp, ok := s.cameraSystem.activeCamera.GetComponent("camera")
	if !ok {
		if s.logger != nil {
			s.logger.Warn("active camera missing camera component")
		}
		return bounds, false
	}

	camera, ok := camComp.(*CameraComponent)
	if !ok {
		if s.logger != nil {
			s.logger.Warn("camera component has incorrect type")
		}
		return bounds, false
	}

	margin := 100.0
	halfWidth := float64(s.cameraSystem.ScreenWidth) / (2.0 * camera.Zoom)
	halfHeight := float64(s.cameraSystem.ScreenHeight) / (2.0 * camera.Zoom)
	bounds.minX = camera.X - halfWidth - margin
	bounds.minY = camera.Y - halfHeight - margin
	bounds.maxX = camera.X + halfWidth + margin
	bounds.maxY = camera.Y + halfHeight + margin

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"camera_x":        camera.X,
			"camera_y":        camera.Y,
			"zoom":            camera.Zoom,
			"viewport_width":  halfWidth * 2,
			"viewport_height": halfHeight * 2,
		}).Debug("viewport bounds calculated")
	}

	return bounds, true
}

// updateEntityAnimation updates animation for a single entity.
func (s *AnimationSystem) updateEntityAnimation(entity *Entity, deltaTime, playerX, playerY float64, viewport viewportBounds, hasViewport bool) error {
	animComp := s.getAnimationComponent(entity)
	if animComp == nil {
		return nil
	}

	s.stats.AnimatedEntities++

	spriteComp := s.getSpriteComponent(entity)
	if spriteComp == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Warn("entity has animation but no sprite component")
		}
		return nil
	}

	pos, ok := s.getEntityPosition(entity)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Warn("entity has animation but no position component")
		}
		return nil
	}

	if s.shouldCullEntity(entity, pos, viewport, hasViewport) {
		s.stats.CulledByViewport++
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"pos_x":     pos.X,
				"pos_y":     pos.Y,
			}).Debug("entity culled by viewport")
		}
		return nil
	}

	effectiveDeltaTime := s.applyDistanceLOD(animComp, pos, playerX, playerY, deltaTime)

	if err := s.regenerateFramesIfDirty(entity, animComp, spriteComp); err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"error":     err.Error(),
			}).Error("failed to regenerate animation frames")
		}
		return err
	}

	s.updateAnimationFrame(animComp, effectiveDeltaTime)
	s.syncSpriteFrame(entity, animComp, spriteComp)

	return nil
}

// getEntityPosition retrieves entity position component.
func (s *AnimationSystem) getEntityPosition(entity *Entity) (*PositionComponent, bool) {
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("entity missing position component")
		}
		return nil, false
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok && s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "position",
		}).Warn("position component has incorrect type")
	}
	return pos, ok
}

// shouldCullEntity determines if entity should be culled from animation updates.
func (s *AnimationSystem) shouldCullEntity(entity *Entity, pos *PositionComponent, viewport viewportBounds, hasViewport bool) bool {
	if !hasViewport {
		return false
	}

	isLocalPlayer := entity.HasComponent("input")
	if isLocalPlayer {
		return false
	}

	return pos.X < viewport.minX || pos.X > viewport.maxX ||
		pos.Y < viewport.minY || pos.Y > viewport.maxY
}

// applyDistanceLOD applies distance-based level-of-detail adjustments.
func (s *AnimationSystem) applyDistanceLOD(animComp *AnimationComponent, pos *PositionComponent, playerX, playerY, deltaTime float64) float64 {
	if !s.enableDistanceLOD || s.playerEntity == nil {
		return deltaTime
	}

	dx := pos.X - playerX
	dy := pos.Y - playerY
	dist := math.Sqrt(dx*dx + dy*dy)

	var targetFrameTime float64
	var lodTier string
	if dist <= s.distanceCloseThresh {
		targetFrameTime = 1.0 / 12.0
		s.stats.FullRateEntities++
		lodTier = "full"
	} else if dist <= s.distanceMidThresh {
		targetFrameTime = 1.0 / 6.0
		s.stats.HalfRateEntities++
		lodTier = "half"
	} else {
		targetFrameTime = 1.0 / 3.0
		s.stats.StaticEntities++
		lodTier = "static"
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"distance":         dist,
			"lod_tier":         lodTier,
			"target_frametime": targetFrameTime,
		}).Debug("applied distance LOD")
	}

	animComp.FrameTime = targetFrameTime
	return deltaTime
}

// regenerateFramesIfDirty regenerates animation frames if component state changed.
func (s *AnimationSystem) regenerateFramesIfDirty(entity *Entity, animComp *AnimationComponent, spriteComp *EbitenSprite) error {
	if !animComp.Dirty {
		return nil
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"state":     animComp.CurrentState,
			"dirty":     animComp.Dirty,
		}).Debug("regenerating frames for dirty animation")
	}

	s.logFrameGeneration(entity, animComp, spriteComp)

	if err := s.regenerateFrames(entity, animComp, spriteComp); err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"state":     animComp.CurrentState,
				"error":     err.Error(),
			}).Error("failed to regenerate frames")
		}
		return fmt.Errorf("failed to regenerate frames: %w", err)
	}

	animComp.Dirty = false
	s.logGenerationResult(entity, animComp)

	return nil
}

// logFrameGeneration logs animation frame generation at debug level.
func (s *AnimationSystem) logFrameGeneration(entity *Entity, animComp *AnimationComponent, spriteComp *EbitenSprite) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		if entity.HasComponent("input") {
			s.logger.WithFields(logrus.Fields{
				"entityID":   entity.ID,
				"state":      animComp.CurrentState,
				"frameCount": s.getFrameCount(animComp.CurrentState),
				"width":      int(spriteComp.Width),
				"height":     int(spriteComp.Height),
			}).Debug("generating animation frames")
		}
	}

	// Player frame regeneration logged via logger if debug level enabled
}

// logGenerationResult logs the result of frame generation.
func (s *AnimationSystem) logGenerationResult(entity *Entity, animComp *AnimationComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		if entity.HasComponent("input") && len(animComp.Frames) > 0 {
			s.logger.WithFields(logrus.Fields{
				"entityID":        entity.ID,
				"framesGenerated": len(animComp.Frames),
				"currentFrame":    animComp.FrameIndex,
			}).Debug("animation frames generated successfully")
		}
	}
}

// updateAnimationFrame advances animation frame if playing.
func (s *AnimationSystem) updateAnimationFrame(animComp *AnimationComponent, effectiveDeltaTime float64) {
	if animComp.Playing && len(animComp.Frames) > 0 && effectiveDeltaTime > 0 {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"state":       animComp.CurrentState,
				"frame_index": animComp.FrameIndex,
				"frame_count": len(animComp.Frames),
				"delta_time":  effectiveDeltaTime,
				"playing":     animComp.Playing,
			}).Debug("updating animation frame")
		}
		s.updateFrame(animComp, effectiveDeltaTime)
	}
}

// syncSpriteFrame synchronizes sprite component with current animation frame.
func (s *AnimationSystem) syncSpriteFrame(entity *Entity, animComp *AnimationComponent, spriteComp *EbitenSprite) {
	if frame := animComp.CurrentFrame(); frame != nil {
		spriteComp.Image = frame
	} else if entity.HasComponent("input") && s.logger != nil {
		// Log warning if player frame is missing (unexpected condition)
		s.logger.WithFields(logrus.Fields{
			"frameIndex": animComp.FrameIndex,
			"frameCount": len(animComp.Frames),
		}).Warn("player animation frame is nil - regeneration may be needed")
	}
}

// updateFrame advances the animation frame based on delta time.
func (s *AnimationSystem) updateFrame(anim *AnimationComponent, deltaTime float64) {
	anim.TimeAccumulator += deltaTime

	// Check if it's time to advance frame
	if anim.TimeAccumulator >= anim.FrameTime {
		oldFrame := anim.FrameIndex
		anim.TimeAccumulator -= anim.FrameTime
		anim.FrameIndex++

		// Handle loop or completion
		if anim.FrameIndex >= len(anim.Frames) {
			if anim.Loop {
				anim.FrameIndex = 0
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"state":     anim.CurrentState,
						"old_frame": oldFrame,
						"new_frame": anim.FrameIndex,
						"loop":      true,
					}).Debug("animation looped")
				}
			} else {
				anim.FrameIndex = len(anim.Frames) - 1
				anim.Playing = false
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"state":       anim.CurrentState,
						"final_frame": anim.FrameIndex,
					}).Debug("animation completed")
				}
				if anim.OnComplete != nil {
					anim.OnComplete()
				}
			}
		}
	}
}

// regenerateFrames generates animation frames for the current state.
func (s *AnimationSystem) regenerateFrames(entity *Entity, anim *AnimationComponent, sprite *EbitenSprite) error {
	// Check cache first
	cacheKey := s.getCacheKey(anim.Seed, anim.CurrentState)

	s.cacheMutex.RLock()
	if frames, exists := s.frameCache[cacheKey]; exists {
		s.cacheMutex.RUnlock()
		anim.Frames = frames
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"cache_key":   cacheKey,
				"frame_count": len(frames),
			}).Debug("animation frames loaded from cache")
		}
		return nil
	}
	s.cacheMutex.RUnlock()

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"cache_key": cacheKey,
			"state":     anim.CurrentState,
			"seed":      anim.Seed,
		}).Debug("cache miss, generating new animation frames")
	}

	// Generate frames using sprite generator
	frames, err := s.generateFrames(entity, anim, sprite)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"state":     anim.CurrentState,
				"error":     err.Error(),
			}).Error("frame generation failed")
		}
		return err
	}

	// Cache frames
	s.cacheFrames(cacheKey, frames)
	anim.Frames = frames

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"cache_key":   cacheKey,
			"frame_count": len(frames),
			"cache_size":  len(s.frameCache),
		}).Debug("animation frames generated and cached")
	}

	return nil
}

// generateFrames creates animation frames using the sprite generator.
func (s *AnimationSystem) generateFrames(entity *Entity, anim *AnimationComponent, sprite *EbitenSprite) ([]*ebiten.Image, error) {
	// Determine frame count based on animation state
	frameCount := s.getFrameCount(anim.CurrentState)
	if anim.FrameCount > 0 {
		frameCount = anim.FrameCount
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"state":       anim.CurrentState,
			"frame_count": frameCount,
			"seed":        anim.Seed,
		}).Debug("generating animation frames")
	}

	frames := make([]*ebiten.Image, frameCount)

	// Get sprite configuration from entity
	config := s.buildSpriteConfig(entity, sprite, anim)

	// Phase 1.2: Generate the base sprite using cache if available
	// CRITICAL FIX: Generate the base sprite ONCE, then transform it for each frame
	// This prevents the "mutating shapes" issue where each frame is a different sprite
	var baseSprite *ebiten.Image
	var err error

	if s.spriteCache != nil {
		// Create cache key from sprite config
		cacheKey := cache.GenerateKey(config.Seed, config.GenreID, config.Variation)
		baseSprite, _ = s.spriteCache.Get(cacheKey)
		if baseSprite == nil {
			// Cache miss: generate and cache
			baseSprite, err = s.spriteGenerator.Generate(config)
			if err == nil {
				s.spriteCache.Put(cacheKey, baseSprite)
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					stats := s.spriteCache.Stats()
					s.logger.WithFields(logrus.Fields{
						"entity_id":  entity.ID,
						"cache_key":  string(cacheKey),
						"cache_size": stats.EntryCount,
						"hit_rate":   stats.HitRate(),
					}).Debug("base sprite cached")
				}
			}
		} else if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"cache_key": string(cacheKey),
			}).Debug("base sprite cache hit")
		}
	} else {
		// No cache: generate directly
		baseSprite, err = s.spriteGenerator.Generate(config)
	}

	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"config":    fmt.Sprintf("%+v", config),
				"error":     err.Error(),
			}).Error("failed to generate base sprite")
		}
		return nil, fmt.Errorf("failed to generate base sprite: %w", err)
	}

	// Now generate each animation frame by applying transformations to the base sprite
	for i := 0; i < frameCount; i++ {
		frame, err := s.generateTransformedFrame(baseSprite, config, string(anim.CurrentState), i, frameCount)
		if err != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":   entity.ID,
					"frame_index": i,
					"error":       err.Error(),
				}).Error("frame transformation failed")
			}
			return nil, fmt.Errorf("frame %d generation failed: %w", i, err)
		}
		frames[i] = frame
	}

	return frames, nil
}

// generateTransformedFrame creates a single animation frame by applying transformations to a base sprite.
// This ensures consistent sprite appearance across all frames, with only position/rotation/scale changing.
func (s *AnimationSystem) generateTransformedFrame(baseSprite *ebiten.Image, config sprites.Config, state string, frameIndex, frameCount int) (*ebiten.Image, error) {
	// Calculate transformations for this frame
	offset := calculateAnimationOffset(state, frameIndex, frameCount)
	rotation := calculateAnimationRotation(state, frameIndex, frameCount)
	scale := calculateAnimationScale(state, frameIndex, frameCount)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"state":       state,
			"frame_index": frameIndex,
			"frame_count": frameCount,
			"offset_x":    offset.X,
			"offset_y":    offset.Y,
			"rotation":    rotation,
			"scale":       scale,
		}).Debug("generating transformed frame")
	}

	// Create output image with room for transformations
	outputWidth := config.Width + int(math.Abs(offset.X)*2) + 10
	outputHeight := config.Height + int(math.Abs(offset.Y)*2) + 10
	img := ebiten.NewImage(outputWidth, outputHeight)

	// Apply transformations to the base sprite
	opts := &ebiten.DrawImageOptions{}

	// Center sprite in output image
	centerX := float64(outputWidth) / 2
	centerY := float64(outputHeight) / 2

	// Apply scale around center
	if scale != 1.0 {
		opts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply rotation around center
	if rotation != 0 {
		opts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		opts.GeoM.Rotate(rotation)
		opts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply position offset and center in output
	opts.GeoM.Translate(centerX-float64(config.Width)/2+offset.X, centerY-float64(config.Height)/2+offset.Y)

	img.DrawImage(baseSprite, opts)

	return img, nil
}

// Animation transformation helper functions

// calculateAnimationOffset computes position offset for animation frame.
func calculateAnimationOffset(state string, frameIndex, frameCount int) struct{ X, Y float64 } {
	t := float64(frameIndex) / float64(frameCount)
	offset := struct{ X, Y float64 }{X: 0, Y: 0}

	switch state {
	case "idle":
		// Phase 15.2: Subtle breathing animation
		// Gentle vertical oscillation with slight horizontal sway
		breathCycle := math.Sin(t * 2 * math.Pi)
		offset.Y = breathCycle * 0.8           // Very subtle 0.8px vertical breathing
		offset.X = math.Sin(t*4*math.Pi) * 0.3 // Even more subtle horizontal sway

	case "walk", "run":
		// Bobbing motion - SIGNIFICANTLY increased amplitude for visibility
		cycle := math.Sin(t * 2 * math.Pi)
		offset.Y = cycle * 8.0 // Increased from 4.0 to 8.0 pixels (200% increase)

	case "jump":
		// Parabolic arc
		offset.Y = -4.0 * (t - t*t) * 15.0 // Jump up and down

	case "attack":
		// Phase 15.2: Enhanced forward lunge with better follow-through
		// Wind-up (0-0.2), strike (0.2-0.5), follow-through (0.5-1.0)
		if t < 0.2 {
			// Wind-up: slight backward movement
			offset.X = -(t / 0.2) * 2.0
		} else if t < 0.5 {
			// Strike: rapid forward lunge
			strikeT := (t - 0.2) / 0.3
			offset.X = -2.0 + strikeT*18.0 // From -2 to +16 pixels
		} else {
			// Follow-through: gradual return with slight overextension
			followT := (t - 0.5) / 0.5
			offset.X = 16.0 - followT*followT*16.0 // Quadratic easing for smooth return
		}

	case "hit":
		// Knockback - SIGNIFICANTLY increased amplitude
		offset.X = -(1.0 - t) * 12.0 // Increased from 6.0 to 12.0 pixels (200% increase)

	case "death":
		// Fall down - SIGNIFICANTLY increased amplitude
		offset.Y = t * 20.0 // Increased from 12.0 to 20.0 pixels (167% increase)
	}

	return offset
}

// calculateAnimationRotation computes rotation for animation frame.
func calculateAnimationRotation(state string, frameIndex, frameCount int) float64 {
	t := float64(frameIndex) / float64(frameCount)

	switch state {
	case "idle":
		// Phase 15.2: Very subtle head tilt for breathing animation
		return math.Sin(t*2*math.Pi) * 0.03 // Tiny oscillation (0.03 radians ≈ 1.7 degrees)

	case "attack":
		// Phase 15.2: Enhanced swing arc with better follow-through
		// Wind-up (0-0.2), strike (0.2-0.5), follow-through (0.5-1.0)
		if t < 0.2 {
			// Wind up: slight backward rotation
			windupT := t / 0.2
			return -windupT * 0.4 // -0.4 radians (~23 degrees)
		} else if t < 0.5 {
			// Swing through: rapid forward rotation
			strikeT := (t - 0.2) / 0.3
			return -0.4 + strikeT*1.8 // From -0.4 to +1.4 radians
		} else {
			// Follow through: continued rotation with deceleration
			followT := (t - 0.5) / 0.5
			// Use sine easing for smooth deceleration
			easedT := math.Sin(followT * math.Pi / 2)
			return 1.4 - easedT*0.8 // From +1.4 to +0.6 radians, smooth follow-through
		}

	case "death":
		// Rotate while falling
		return t * math.Pi / 2 // 90 degree rotation

	case "cast":
		// Gentle sway - increased
		return math.Sin(t*2*math.Pi) * 0.2 // Increased from 0.1 to 0.2
	}

	return 0
}

// calculateAnimationScale computes scale factor for animation frame.
func calculateAnimationScale(state string, frameIndex, frameCount int) float64 {
	t := float64(frameIndex) / float64(frameCount)

	switch state {
	case "idle":
		// Phase 15.2: Subtle breathing scale (chest expansion/contraction)
		breathCycle := math.Sin(t * 2 * math.Pi)
		return 1.0 + breathCycle*0.015 // Very subtle 1.5% scale change

	case "jump":
		// Squash and stretch - more dramatic
		if t < 0.2 {
			return 1.0 - t // Squash before jump (more pronounced)
		} else if t < 0.8 {
			return 0.8 + (t-0.2)*0.5 // Stretch during jump
		} else {
			return 1.0 - (t - 0.8) // Squash on landing
		}

	case "hit":
		// Squash on impact - more dramatic
		return 1.0 - t*0.4 // Increased from 0.2 to 0.4

	case "attack":
		// Phase 15.2: Enhanced anticipation and follow-through scale
		// Slight anticipation squat, then expansion during strike
		if t < 0.2 {
			// Anticipation: slight compression
			return 1.0 - (t/0.2)*0.05
		} else if t < 0.5 {
			// Strike: expand for power
			strikeT := (t - 0.2) / 0.3
			return 0.95 + strikeT*0.15 // From 0.95 to 1.10
		} else {
			// Follow-through: gradual return to normal
			followT := (t - 0.5) / 0.5
			return 1.10 - followT*0.10 // From 1.10 to 1.00
		}
	}

	return 1.0
}

// buildSpriteConfig creates a sprite configuration from entity components.
func (s *AnimationSystem) buildSpriteConfig(entity *Entity, sprite *EbitenSprite, anim *AnimationComponent) sprites.Config {
	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      int(sprite.Width),
		Height:     int(sprite.Height),
		Seed:       anim.Seed,
		Complexity: 0.7,
		Palette:    nil,
		Custom:     make(map[string]interface{}),
	}

	config.Custom["useAerial"] = true

	if entity.HasComponent("input") {
		s.configurePlayerSprite(&config, entity, anim)
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"entity_type": "player",
				"seed":        anim.Seed,
			}).Debug("configured player sprite")
		}
	} else if teamComp, ok := entity.GetComponent("team"); ok {
		if team, ok := teamComp.(*TeamComponent); ok {
			if team.TeamID == 2 {
				s.configureEnemySprite(&config, entity, anim)
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"entity_id": entity.ID,
						"team_id":   team.TeamID,
						"seed":      anim.Seed,
					}).Debug("configured enemy sprite")
				}
			}
		}
	}

	s.applyGenreConfig(&config, entity)
	config.Custom["useAerial"] = true

	return config
}

// configurePlayerSprite configures sprite settings for player entities.
func (s *AnimationSystem) configurePlayerSprite(config *sprites.Config, entity *Entity, anim *AnimationComponent) {
	config.Custom["entityType"] = "humanoid"
	facing := s.determineFacingDirection(entity, anim)
	config.Custom["facing"] = facing

	if entity.HasComponent("equipment") {
		config.Custom["hasWeapon"] = true
		config.Custom["hasShield"] = false
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"has_weapon": true,
				"facing":     facing,
			}).Debug("player equipment configured")
		}
	}
}

// configureEnemySprite configures sprite settings for enemy entities.
func (s *AnimationSystem) configureEnemySprite(config *sprites.Config, entity *Entity, anim *AnimationComponent) {
	entityType := s.determineEnemyType(entity, config)
	config.Custom["entityType"] = entityType
	facing := s.determineFacingDirection(entity, anim)
	config.Custom["facing"] = facing

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"entity_type": entityType,
			"facing":      facing,
		}).Debug("enemy sprite configured")
	}
}

// determineEnemyType analyzes entity components to determine monster type.
func (s *AnimationSystem) determineEnemyType(entity *Entity, config *sprites.Config) string {
	if bossType := s.checkBossType(entity, config); bossType != "" {
		return bossType
	}
	if sizeType := s.checkSizeBasedType(entity); sizeType != "" {
		return sizeType
	}
	return "humanoid"
}

// checkBossType determines if the entity is a boss based on attack damage.
func (s *AnimationSystem) checkBossType(entity *Entity, config *sprites.Config) string {
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		return ""
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok || attack.Damage <= 20 {
		return ""
	}

	config.Custom["isBoss"] = true
	config.Custom["bossScale"] = 1.5
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"entity_type": "boss",
			"damage":      attack.Damage,
			"boss_scale":  1.5,
		}).Debug("boss enemy detected")
	}
	return "boss"
}

// checkSizeBasedType determines entity type based on collider size.
func (s *AnimationSystem) checkSizeBasedType(entity *Entity) string {
	colliderComp, ok := entity.GetComponent("collider")
	if !ok {
		return ""
	}
	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		return ""
	}

	if collider.Width > 48 {
		s.logEntityType(entity.ID, "monster", collider.Width)
		return "monster"
	}
	if collider.Width < 24 {
		s.logEntityType(entity.ID, "minion", collider.Width)
		return "minion"
	}
	return ""
}

// logEntityType logs the detected entity type at debug level.
func (s *AnimationSystem) logEntityType(entityID uint64, entityType string, width float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"entity_type":    entityType,
			"collider_width": width,
		}).Debug("entity type detected")
	}
}

// determineFacingDirection calculates entity facing direction from velocity.
func (s *AnimationSystem) determineFacingDirection(entity *Entity, anim *AnimationComponent) string {
	facing := "down"

	velComp, hasVel := entity.GetComponent("velocity")
	if !hasVel {
		if anim.LastFacing != "" {
			return anim.LastFacing
		}
		return facing
	}

	vel, ok := velComp.(*VelocityComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "velocity",
			}).Warn("velocity component has incorrect type")
		}
		return facing
	}

	if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
		facing = s.calculateFacingFromVelocity(vel)
		anim.LastFacing = facing
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"facing":    facing,
				"vx":        vel.VX,
				"vy":        vel.VY,
			}).Debug("updated facing direction from velocity")
		}
	} else if anim.LastFacing != "" {
		facing = anim.LastFacing
	}

	return facing
}

// calculateFacingFromVelocity determines facing direction from velocity vector.
func (s *AnimationSystem) calculateFacingFromVelocity(vel *VelocityComponent) string {
	if math.Abs(vel.VX) > math.Abs(vel.VY) {
		if vel.VX > 0 {
			return "right"
		}
		return "left"
	}

	if vel.VY > 0 {
		return "down"
	}
	return "up"
}

// applyGenreConfig applies genre configuration to sprite config.
func (s *AnimationSystem) applyGenreConfig(config *sprites.Config, entity *Entity) {
	if genreComp, ok := entity.GetComponent("genre"); ok && genreComp != nil {
		if gc, ok := genreComp.(interface{ GetGenreID() string }); ok {
			genreID := gc.GetGenreID()
			config.GenreID = genreID
			config.Custom["genre"] = genreID
			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"genre_id":  genreID,
				}).Debug("applied genre config from entity")
			}
			return
		}
	}

	if config.GenreID == "" {
		config.GenreID = "fantasy"
		config.Custom["genre"] = "fantasy"
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"genre_id":  "fantasy",
			}).Debug("applied default fantasy genre")
		}
	}
}

// getFrameCount returns the number of frames for an animation state.
func (s *AnimationSystem) getFrameCount(state AnimationState) int {
	switch state {
	case AnimationStateIdle:
		return 8 // Phase 15.2: Increased from 4 to 8 for smoother breathing animation
	case AnimationStateWalk:
		return 8 // 8-frame walk cycle
	case AnimationStateRun:
		return 8
	case AnimationStateAttack:
		return 8 // Phase 15.2: Increased from 6 to 8 for better follow-through
	case AnimationStateCast:
		return 8 // Cast preparation, cast, recovery
	case AnimationStateHit:
		return 3 // Quick hit reaction
	case AnimationStateDeath:
		return 6 // Death animation
	case AnimationStateJump:
		return 4
	case AnimationStateCrouch:
		return 2
	case AnimationStateUse:
		return 4
	default:
		return 4
	}
}

// cacheFrames stores frames in cache with LRU eviction.
func (s *AnimationSystem) cacheFrames(key string, frames []*ebiten.Image) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Check if already cached
	if _, exists := s.frameCache[key]; exists {
		return
	}

	// Evict oldest entry if cache is full
	if len(s.frameCache) >= s.maxCacheSize {
		oldestKey := s.cacheKeys[0]
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"evicted_key": oldestKey,
				"cache_size":  len(s.frameCache),
				"max_size":    s.maxCacheSize,
			}).Debug("evicting oldest cache entry")
		}
		delete(s.frameCache, oldestKey)
		s.cacheKeys = s.cacheKeys[1:]
	}

	// Add to cache
	s.frameCache[key] = frames
	s.cacheKeys = append(s.cacheKeys, key)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"cache_key":   key,
			"frame_count": len(frames),
			"cache_size":  len(s.frameCache),
		}).Debug("frames added to cache")
	}
}

// getCacheKey generates a cache key for animation frames.
func (s *AnimationSystem) getCacheKey(seed int64, state AnimationState) string {
	key := fmt.Sprintf("%d_%s", seed, state)
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"seed":      seed,
			"state":     state,
			"cache_key": key,
		}).Debug("generated cache key")
	}
	return key
}

// ClearCache clears the animation frame cache.
func (s *AnimationSystem) ClearCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	entriesCleared := len(s.frameCache)
	s.frameCache = make(map[string][]*ebiten.Image)
	s.cacheKeys = make([]string, 0, s.maxCacheSize)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entries_cleared": entriesCleared,
		}).Info("animation cache cleared")
	}
}

// GetCacheSize returns the current number of cached animation sequences.
func (s *AnimationSystem) GetCacheSize() int {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	size := len(s.frameCache)
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"cache_size": size,
			"max_size":   s.maxCacheSize,
		}).Debug("retrieved cache size")
	}
	return size
}

// Helper methods to get components

func (s *AnimationSystem) getAnimationComponent(entity *Entity) *AnimationComponent {
	comp, ok := entity.GetComponent("animation")
	if !ok || comp == nil {
		return nil
	}
	animComp, ok := comp.(*AnimationComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "animation",
			}).Warn("animation component has incorrect type")
		}
		return nil
	}
	return animComp
}

func (s *AnimationSystem) getSpriteComponent(entity *Entity) *EbitenSprite {
	comp, ok := entity.GetComponent("sprite")
	if !ok || comp == nil {
		return nil
	}
	spriteComp, ok := comp.(*EbitenSprite)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "sprite",
			}).Warn("sprite component has incorrect type")
		}
		return nil
	}
	return spriteComp
}

// TransitionState safely transitions an entity to a new animation state.
// Returns false if entity has no animation component.
func (s *AnimationSystem) TransitionState(entity *Entity, newState AnimationState) bool {
	animComp := s.getAnimationComponent(entity)
	if animComp == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Warn("cannot transition animation state: no animation component")
		}
		return false
	}

	oldState := animComp.CurrentState
	animComp.SetState(newState)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"old_state": oldState,
			"new_state": newState,
		}).Debug("animation state transitioned")
	}

	return true
}
