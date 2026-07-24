// Package engine provides animation system for updating entity animations.
package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// AnimationSyncer is implemented by network-layer managers (e.g.
// *network.AnimationSyncManager) that provide delta-compressed, jitter-buffered
// animation state synchronisation for multiplayer. When set on AnimationSystem
// via SetSyncManager, game code can call ShouldSync before transmitting a state
// change and RecordSync after confirming the packet was queued.
//
// The interface uses only engine-package types to avoid a circular import with
// pkg/network. Set nil to disable multiplayer animation sync (offline/singleplayer).
type AnimationSyncer interface {
	// ShouldSync returns true if the state change for entityID should be
	// transmitted (i.e. the state actually changed since the last sync).
	ShouldSync(entityID uint64, newState AnimationState) bool
	// RecordSync records that a state packet of bytesSent bytes was queued for
	// entityID so that ShouldSync can correctly detect the next delta.
	RecordSync(entityID uint64, state AnimationState, bytesSent int)
	// DrainRemoteState pops the next buffered animation state for entityID from
	// the jitter buffer. Returns false when the buffer is empty.
	// state is the new AnimationState; frameIdx is the frame number to resume from.
	DrainRemoteState(entityID uint64) (state AnimationState, frameIdx int, ok bool)
}

// animStatePacketBytes is the fixed wire size of one animation state packet.
// Format: [EntityID:8][State:1][FrameIndex:2][Timestamp:8][Loop:1] = 20 bytes.
// Defined here because AnimationSystem cannot import pkg/network without
// creating a circular dependency. Keep in sync with AnimationStatePacket.Encode
// in pkg/network/animation_sync.go whenever the packet layout changes.
const animStatePacketBytes = 20

// Animation Timing:
//   - Default FPS: 12 FPS for close-range entities (0.083s per frame)
//   - Frame Count: 8 frames per animation state (idle, walk, attack, etc.)
//   - Total Duration: ~0.67 seconds per animation cycle at base rate
//
// Distance-Based LOD (Level of Detail):
//   - Close range (≤200px from player): 12 FPS (full animation rate)
//   - Medium range (200-400px): 6 FPS (half animation rate for performance)
//   - Far range (>400px): 3 FPS (minimal animation for distant entities)
//   - Player entity: Always rendered at full rate regardless of distance
//
// Animation States and Typical Usage:
//   - AnimationStateIdle: Looping idle animation (breathing, standing)
//   - AnimationStateWalk: Looping walk animation (movement at normal speed)
//   - AnimationStateRun: Looping run animation (faster movement)
//   - AnimationStateAttack: One-shot attack animation (triggers on combat action)
//   - AnimationStateCast: One-shot spell casting animation
//   - AnimationStateHit: One-shot damage reaction animation
//   - AnimationStateDeath: One-shot death animation (stops at final frame)
//
// Performance Optimizations:
//   - Viewport Culling: Entities outside camera view are not animated
//   - Frame Caching: Animation frames are cached per seed+state combination
//   - Per-Frame Limits: Maximum 16 sprite regenerations per frame to prevent lag
//   - LRU Eviction: Frame cache limited to 100 sequences to manage memory
//   - Pre-Generation: Batch sprite pre-generation during loading screens
//   - Predictive Warming: Access pattern analysis for proactive cache population
//
// Tuning Animation Speed:
//   - Modify AnimationComponent.FrameTime directly to change speed
//   - Lower values = faster animation (e.g., 0.05s = 20 FPS)
//   - Higher values = slower animation (e.g., 0.2s = 5 FPS)
//   - Default: 1.0/12.0 (~0.083s) provides smooth motion without overhead
//
// Example:
//
//	// Create custom animation with faster playback
//	anim := NewAnimationComponent(seed)
//	anim.FrameTime = 0.05 // 20 FPS for fast combat animation
//	anim.CurrentState = AnimationStateAttack
//	anim.Loop = false // One-shot animation
//	anim.OnComplete = func() {
//	    // Trigger when attack animation finishes
//	}
type AnimationSystem struct {
	spriteGenerator *sprites.Generator
	spriteCache     *cache.SpriteCache         // Phase 1.2: External sprite cache for base sprites
	frameCache      map[uint64][]*ebiten.Image // Cache by key: uint64 combining seed + state
	cacheMutex      sync.RWMutex
	maxCacheSize    int
	cacheKeys       []uint64 // For LRU eviction
	logger          *logrus.Entry
	paletteOptions  *palette.GenerationOptions // Phase 5.4: Custom palette generation options

	// Phase 14.2: Viewport culling and distance-based optimization
	cameraSystem        *CameraSystem
	enableViewportCull  bool
	enableDistanceLOD   bool
	playerEntity        *Entity        // Cached player reference for distance calculations
	distanceCloseThresh float64        // Distance threshold for full animation (default: 200px)
	distanceMidThresh   float64        // Distance threshold for half rate (default: 400px)
	stats               AnimationStats // Performance statistics

	// Performance fix: Per-frame regeneration limit to prevent startup lag
	// When many entities need sprite regeneration, this spreads the work across frames
	maxRegenPerFrame int // Maximum sprite regenerations per frame (0 = unlimited)
	regenCount       int // Current regeneration count this frame

	// V1/V3 Performance fix: Predictive cache warmer for proactive sprite pre-generation
	predictiveWarmer *cache.PredictiveCacheWarmer
	warmerTickCount  int64 // Frame counter for warmer access tracking

	// Performance optimization: Pool for frame slices to reduce allocations
	frameSlicePool sync.Pool

	// V7 Performance fix: Pool for animation frame images and reusable DrawImageOptions
	// Animation frames typically have dimensions near 74-84 pixels (64px sprite + offset + margin)
	// The pool uses bucketed sizes (64, 80, 96, 128) to maximize reuse while minimizing wasted space
	frameImagePool    *animationImagePool
	transformDrawOpts ebiten.DrawImageOptions // Reusable DrawImageOptions for frame generation

	// syncManager provides delta-compressed animation state synchronisation for
	// multiplayer. Nil in offline/singleplayer mode. Set via SetSyncManager.
	syncManager AnimationSyncer

	// stateSender is an optional callback invoked when ShouldSync decides a local
	// animation state change must be transmitted to the server. Set via
	// SetStateSender; nil in offline/singleplayer mode.
	stateSender func(entityID uint64, state AnimationState, frameIdx int)
}

// AnimationStats holds performance statistics for the animation system.
type AnimationStats struct {
	TotalEntities    int // Total entities processed
	AnimatedEntities int // Entities with active animations
	CulledByViewport int // Entities culled by viewport check
	FullRateEntities int // Entities animated at full rate (close)
	HalfRateEntities int // Entities animated at half rate (mid distance)
	StaticEntities   int // Entities rendered as static (far distance)
	DeferredRegen    int // Entities with deferred regeneration (hit per-frame limit)
	CompletedRegen   int // Entities that completed regeneration this frame
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
			"max_regen_per_frame":   16,
		}).Debug("initializing animation system")
	}

	sys := &AnimationSystem{
		spriteGenerator: spriteGenerator,
		frameCache:      make(map[uint64][]*ebiten.Image),
		maxCacheSize:    100, // Cache up to 100 animation sequences
		cacheKeys:       make([]uint64, 0, 100),
		logger:          logEntry,
		// Phase 14.2: Default optimization settings
		enableViewportCull:  true,  // Enabled by default for performance
		enableDistanceLOD:   true,  // Enabled by default for performance
		distanceCloseThresh: 200.0, // Full animation within 200px
		distanceMidThresh:   400.0, // Half rate 200-400px, static beyond
		// Performance fix: Limit sprite regenerations per frame to prevent startup lag
		// V1: Increased from 8 to 16 to reduce spawn stutter from ~200ms to ~100ms
		// With 16 regenerations per frame at 60 FPS, 100 entities regenerate in ~6 frames (~100ms)
		maxRegenPerFrame: 16,
	}

	// Initialize frame slice pool for reuse (reduces allocations during regeneration)
	sys.frameSlicePool = sync.Pool{
		New: func() interface{} {
			// Pre-allocate with capacity 16 to handle all animation types (max 8 frames typical)
			return make([]*ebiten.Image, 0, 16)
		},
	}

	// V7 Performance fix: Initialize animation frame image pool
	// Animation frames use dimensions around 74-84 pixels for standard 64px sprites
	sys.frameImagePool = newAnimationImagePool()

	return sys
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

// SetSyncManager wires a network-layer AnimationSyncer for multiplayer animation
// delta compression. Pass nil to disable (offline / singleplayer mode).
// The manager is used by game code to determine when to transmit a state change
// (ShouldSync) and to record confirmed transmissions (RecordSync).
func (s *AnimationSystem) SetSyncManager(mgr AnimationSyncer) {
	s.syncManager = mgr
}

// GetSyncManager returns the current AnimationSyncer, or nil if not set.
func (s *AnimationSystem) GetSyncManager() AnimationSyncer {
	return s.syncManager
}

// SetStateSender wires the outgoing animation-state send path for multiplayer.
// fn is called whenever ShouldSync decides a local animation state change must be
// transmitted; it should encode the packet and call TCPClient.SendInput. Pass nil
// to disable (offline / singleplayer mode).
func (s *AnimationSystem) SetStateSender(fn func(entityID uint64, state AnimationState, frameIdx int)) {
	s.stateSender = fn
}

// SetPaletteOptions sets custom palette generation options for all sprites (Phase 5.4).
// These options control color harmony, mood, and rarity/intensity of generated palettes.
// If nil, uses default palette generation.
func (s *AnimationSystem) SetPaletteOptions(opts *palette.GenerationOptions) {
	s.paletteOptions = opts
	// Clear frame cache to regenerate sprites with new palette options
	s.cacheMutex.Lock()
	// V7: Return all cached images to pool before clearing
	for _, frames := range s.frameCache {
		for _, img := range frames {
			s.frameImagePool.Put(img)
		}
	}
	s.frameCache = make(map[uint64][]*ebiten.Image)
	s.cacheKeys = make([]uint64, 0, s.maxCacheSize)
	s.cacheMutex.Unlock()

	if s.logger != nil {
		if opts != nil {
			s.logger.WithFields(logrus.Fields{
				"harmony":    opts.Harmony,
				"mood":       opts.Mood,
				"rarity":     opts.Rarity,
				"min_colors": opts.MinColors,
			}).Info("palette options set, frame cache cleared")
		} else {
			s.logger.Debug("palette options cleared, using defaults")
		}
	}
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
			// V7: Return evicted images to pool
			if evictedFrames, ok := s.frameCache[oldestKey]; ok {
				for _, img := range evictedFrames {
					s.frameImagePool.Put(img)
				}
			}
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

// SetMaxRegenPerFrame sets the maximum sprite regenerations allowed per frame.
// This prevents startup lag by spreading sprite generation over multiple frames.
// Default is 8. Set to 0 for unlimited (may cause lag with many entities).
// Recommended values: 4-16 depending on hardware and entity count.
// Negative values are clamped to 0 (unlimited).
func (s *AnimationSystem) SetMaxRegenPerFrame(maxRegen int) {
	if maxRegen < 0 {
		maxRegen = 0
	}
	s.maxRegenPerFrame = maxRegen
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"max_regen_per_frame": maxRegen,
		}).Debug("max regeneration per frame updated")
	}
}

// GetMaxRegenPerFrame returns the current per-frame regeneration limit.
func (s *AnimationSystem) GetMaxRegenPerFrame() int {
	return s.maxRegenPerFrame
}

// SetPredictiveWarmer sets the predictive cache warmer for proactive sprite pre-generation.
// When set, the animation system records access patterns and can warm predicted sprites.
func (s *AnimationSystem) SetPredictiveWarmer(warmer *cache.PredictiveCacheWarmer) {
	s.predictiveWarmer = warmer
	if s.logger != nil {
		if warmer != nil {
			s.logger.Info("predictive cache warmer connected to animation system")
		} else {
			s.logger.Debug("predictive cache warmer disconnected from animation system")
		}
	}
}

// GetPredictiveWarmer returns the current predictive cache warmer, or nil if not set.
func (s *AnimationSystem) GetPredictiveWarmer() *cache.PredictiveCacheWarmer {
	return s.predictiveWarmer
}

// PreGenerateSprites pre-generates animation frames for a batch of entities.
// Call this during loading screens to avoid first-use sprite generation stutter.
// Bypasses the per-frame regeneration limit since this runs outside the game loop.
// Returns the number of entities whose sprites were successfully pre-generated.
func (s *AnimationSystem) PreGenerateSprites(entities []*Entity) int {
	generated := 0
	for _, entity := range entities {
		animComp := s.getAnimationComponent(entity)
		spriteComp := s.getSpriteComponent(entity)
		if animComp == nil || spriteComp == nil {
			continue
		}
		if !animComp.Dirty {
			continue
		}

		if err := s.regenerateFrames(entity, animComp, spriteComp); err != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"error":     err.Error(),
				}).Warn("pre-generation failed for entity")
			}
			continue
		}
		animComp.Dirty = false
		generated++
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"total_entities": len(entities),
			"generated":      generated,
		}).Info("sprite pre-generation completed")
	}
	return generated
}

// WarmPredictedSprites uses the predictive cache warmer to queue predicted sprites
// for pre-generation. Call periodically (e.g., every 60 frames) to maintain high
// cache hit rates. Returns the number of sprites queued for warming.
func (s *AnimationSystem) WarmPredictedSprites() int {
	if s.predictiveWarmer == nil || s.spriteCache == nil {
		return 0
	}

	predictions := s.predictiveWarmer.PredictNext()
	if len(predictions) == 0 {
		return 0
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"predicted_count": len(predictions),
		}).Debug("warming predicted sprites")
	}

	return len(predictions)
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
			"deferred_regen":    s.stats.DeferredRegen,
			"completed_regen":   s.stats.CompletedRegen,
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
			"deferred_regen":    s.stats.DeferredRegen,
			"completed_regen":   s.stats.CompletedRegen,
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
	// Reset per-frame regeneration counter
	s.regenCount = 0
}

// getPlayerPosition retrieves the current player position for distance calculations.
// Uses typed getter for ~91x faster access vs generic GetComponent.
func (s *AnimationSystem) getPlayerPosition() (float64, float64) {
	var playerX, playerY float64
	if s.playerEntity != nil {
		if pos := s.playerEntity.GetPosition(); pos != nil {
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

	// Apply any buffered remote animation state before local processing so
	// that jitter-buffered packets from remote players are consumed each tick.
	s.drainRemoteBuffer(entity, animComp)

	spriteComp, pos, shouldContinue := s.validateAnimationComponents(entity)
	if !shouldContinue {
		return nil
	}

	if s.shouldCullEntity(entity, pos, viewport, hasViewport) {
		s.logAndCountCulled(entity.ID, pos)
		return nil
	}

	return s.processAnimation(entity, animComp, spriteComp, pos, playerX, playerY, deltaTime)
}

// validateAnimationComponents checks sprite and position components.
func (s *AnimationSystem) validateAnimationComponents(entity *Entity) (*EbitenSprite, *PositionComponent, bool) {
	spriteComp := s.getSpriteComponent(entity)
	if spriteComp == nil {
		s.logMissingSprite(entity.ID)
		return nil, nil, false
	}

	pos, ok := s.getEntityPosition(entity)
	if !ok {
		s.logMissingPosition(entity.ID)
		return nil, nil, false
	}

	return spriteComp, pos, true
}

// logMissingSprite logs when an entity lacks a sprite component.
func (s *AnimationSystem) logMissingSprite(entityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("entity has animation but no sprite component")
	}
}

// logMissingPosition logs when an entity lacks a position component.
func (s *AnimationSystem) logMissingPosition(entityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("entity has animation but no position component")
	}
}

// logAndCountCulled logs and counts culled entities.
func (s *AnimationSystem) logAndCountCulled(entityID uint64, pos *PositionComponent) {
	s.stats.CulledByViewport++
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"pos_x":     pos.X,
			"pos_y":     pos.Y,
		}).Debug("entity culled by viewport")
	}
}

// processAnimation applies LOD, regenerates frames, and updates animation.
func (s *AnimationSystem) processAnimation(entity *Entity, animComp *AnimationComponent, spriteComp *EbitenSprite, pos *PositionComponent, playerX, playerY, deltaTime float64) error {
	effectiveDeltaTime := s.applyDistanceLOD(animComp, pos, playerX, playerY, deltaTime)

	if err := s.regenerateFramesIfDirty(entity, animComp, spriteComp); err != nil {
		s.logFrameRegenerationError(entity.ID, err)
		return err
	}

	s.updateAnimationFrame(animComp, effectiveDeltaTime)
	s.syncSpriteFrame(entity, animComp, spriteComp)
	return nil
}

// logFrameRegenerationError logs frame regeneration failures.
func (s *AnimationSystem) logFrameRegenerationError(entityID uint64, err error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"error":     err.Error(),
		}).Error("failed to regenerate animation frames")
	}
}

// getEntityPosition retrieves entity position component.
// Uses typed getter for ~91x faster access vs generic GetComponent.
func (s *AnimationSystem) getEntityPosition(entity *Entity) (*PositionComponent, bool) {
	pos := entity.GetPosition()
	if pos == nil {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("entity missing position component")
		}
		return nil, false
	}
	return pos, true
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
// Respects per-frame regeneration limit to prevent startup lag.
func (s *AnimationSystem) regenerateFramesIfDirty(entity *Entity, animComp *AnimationComponent, spriteComp *EbitenSprite) error {
	if !animComp.Dirty {
		return nil
	}

	// Performance fix: Check if we've hit the per-frame regeneration limit
	// Always allow player entity regeneration (entities with "input" component)
	isPlayer := entity.HasComponent("input")
	if s.maxRegenPerFrame > 0 && s.regenCount >= s.maxRegenPerFrame && !isPlayer {
		// Defer regeneration to next frame to prevent lag
		s.stats.DeferredRegen++
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     entity.ID,
				"regen_count":   s.regenCount,
				"max_per_frame": s.maxRegenPerFrame,
			}).Debug("deferring sprite regeneration to next frame")
		}
		return nil // Leave Dirty=true so it regenerates next frame
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
	s.regenCount++
	s.stats.CompletedRegen++
	s.logGenerationResult(entity, animComp)
	// animComp.FrameIndex is the correct value here: regenerateSprite always
	// regenerates from FrameIndex 0 when a state transition occurs, so it
	// reflects the first frame of the newly-generated state.
	s.notifyStateChange(entity.ID, animComp.CurrentState, animComp.FrameIndex)

	return nil
}

// drainRemoteBuffer pops the next buffered animation state for a remote entity
// (an entity without a local "input" component) and applies it to animComp.
// This gives jitter-buffered, delta-compressed playback of remote-player
// animations received via the network layer. No-op when syncManager is nil or
// the buffer is empty.
func (s *AnimationSystem) drainRemoteBuffer(entity *Entity, animComp *AnimationComponent) {
	if s.syncManager == nil || entity.HasComponent("input") {
		return
	}
	state, frameIdx, ok := s.syncManager.DrainRemoteState(entity.ID)
	if !ok {
		return
	}
	// Update state and frame independently so a same-state packet can still
	// advance the frame index (reviewer suggestion: separate conditionals).
	if state != animComp.CurrentState {
		animComp.CurrentState = state
		animComp.Dirty = true
	}
	if frameIdx != animComp.FrameIndex {
		animComp.FrameIndex = frameIdx
	}
}

// notifyStateChange calls the sync manager and (when a stateSender is wired)
// dispatches an outgoing animation-state packet to the server so other clients
// receive remote-player animation updates. No-op when syncManager is nil.
func (s *AnimationSystem) notifyStateChange(entityID uint64, state AnimationState, frameIdx int) {
	if s.syncManager == nil {
		return
	}
	if s.syncManager.ShouldSync(entityID, state) {
		s.syncManager.RecordSync(entityID, state, animStatePacketBytes)
		if s.stateSender != nil {
			s.stateSender(entityID, state, frameIdx)
		}
	}
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
	warmerKey := cache.GenerateKey(anim.Seed, string(anim.CurrentState), 0)

	s.cacheMutex.RLock()
	if frames, exists := s.frameCache[cacheKey]; exists {
		s.cacheMutex.RUnlock()
		anim.Frames = frames
		// V3: Record cache hit for predictive warming
		s.recordWarmerAccess(warmerKey, true)
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

	// V3: Record cache miss for predictive warming
	s.recordWarmerAccess(warmerKey, false)

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
	frameCount := s.determineFrameCount(anim)
	s.logFrameGeneration(entity, anim, sprite)

	config := s.buildSpriteConfig(entity, sprite, anim)

	// Check if entity uses template-based aerial rendering (supports body part animation).
	useBodyPartAnim := false
	if config.Custom != nil {
		if aerial, ok := config.Custom["useAerial"].(bool); ok && aerial {
			if _, ok := config.Custom["entityType"].(string); ok {
				useBodyPartAnim = true
			}
		}
	}

	if useBodyPartAnim {
		return s.generateBodyPartAnimatedFrames(entity, config, anim, frameCount)
	}

	// Fallback: static base sprite with geometric transforms
	baseSprite, err := s.getOrGenerateBaseSprite(entity, config)
	if err != nil {
		return nil, err
	}
	return s.generateAllFrames(entity, baseSprite, config, anim, frameCount)
}

// generateBodyPartAnimatedFrames generates per-frame sprites where each frame
// has body parts in state-specific positions (e.g., legs alternating during walk,
// arms extending during attack). Each frame is a fully rendered template sprite
// with per-body-part offsets baked in, producing visibly animated entity sprites
// instead of static cutouts with geometric transforms.
func (s *AnimationSystem) generateBodyPartAnimatedFrames(entity *Entity, baseConfig sprites.Config, anim *AnimationComponent, frameCount int) ([]*ebiten.Image, error) {
	frames := s.getFrameSlice(frameCount)
	state := string(anim.CurrentState)

	// Extract entity type for creature-aware animation offsets.
	entityType := ""
	if baseConfig.Custom != nil {
		if et, ok := baseConfig.Custom["entityType"].(string); ok {
			entityType = et
		}
	}

	for i := 0; i < frameCount; i++ {
		offsets := sprites.ComputeCreatureFrameOffsets(state, i, frameCount, entityType)

		// Create per-frame config with body part offsets injected
		frameConfig := baseConfig
		frameCustom := make(map[string]interface{}, len(baseConfig.Custom)+1)
		for k, v := range baseConfig.Custom {
			frameCustom[k] = v
		}
		frameCustom["frameOffsets"] = offsets
		frameConfig.Custom = frameCustom

		frame, err := s.spriteGenerator.Generate(frameConfig)
		if err != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":   entity.ID,
					"frame_index": i,
					"state":       state,
					"error":       err.Error(),
				}).Error("body-part animated frame generation failed")
			}
			return nil, fmt.Errorf("frame %d body-part generation failed: %w", i, err)
		}
		frames[i] = frame
	}

	return frames, nil
}

// determineFrameCount calculates the number of frames for an animation.
func (s *AnimationSystem) determineFrameCount(anim *AnimationComponent) int {
	frameCount := s.getFrameCount(anim.CurrentState)
	if anim.FrameCount > 0 {
		frameCount = anim.FrameCount
	}
	return frameCount
}

// getOrGenerateBaseSprite retrieves a cached sprite or generates a new one.
func (s *AnimationSystem) getOrGenerateBaseSprite(entity *Entity, config sprites.Config) (*ebiten.Image, error) {
	if s.spriteCache != nil {
		return s.getCachedBaseSprite(entity, config)
	}
	return s.spriteGenerator.Generate(config)
}

// getCachedBaseSprite retrieves or generates a sprite using the cache.
func (s *AnimationSystem) getCachedBaseSprite(entity *Entity, config sprites.Config) (*ebiten.Image, error) {
	cacheKey := s.generateSpriteCacheKey(config)
	baseSprite, hit := s.spriteCache.Get(cacheKey)
	if hit {
		s.logSpriteHit(entity, cacheKey)
		return baseSprite, nil
	}

	baseSprite, err := s.spriteGenerator.Generate(config)
	if err == nil {
		s.spriteCache.Put(cacheKey, baseSprite)
		s.logSpriteMiss(entity, cacheKey)
	}
	return baseSprite, err
}

// logSpriteHit logs a cache hit event.
func (s *AnimationSystem) logSpriteHit(entity *Entity, cacheKey cache.CacheKey) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"cache_key": string(cacheKey),
		}).Debug("base sprite cache hit")
	}
}

// logSpriteMiss logs a cache miss event with statistics.
func (s *AnimationSystem) logSpriteMiss(entity *Entity, cacheKey cache.CacheKey) {
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

// generateAllFrames creates all animation frames from a base sprite.
// Uses pooled frame slices to reduce allocations during regeneration.
func (s *AnimationSystem) generateAllFrames(entity *Entity, baseSprite *ebiten.Image, config sprites.Config, anim *AnimationComponent, frameCount int) ([]*ebiten.Image, error) {
	// Get slice from pool and resize to needed capacity
	frames := s.getFrameSlice(frameCount)

	for i := 0; i < frameCount; i++ {
		frame, err := s.generateTransformedFrame(baseSprite, config, string(anim.CurrentState), i, frameCount)
		if err != nil {
			s.logFrameTransformError(entity, i, err)
			return nil, fmt.Errorf("frame %d generation failed: %w", i, err)
		}
		frames[i] = frame
	}
	return frames, nil
}

// logFrameTransformError logs an error during frame transformation.
func (s *AnimationSystem) logFrameTransformError(entity *Entity, frameIndex int, err error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"frame_index": frameIndex,
			"error":       err.Error(),
		}).Error("frame transformation failed")
	}
}

// generateTransformedFrame creates a single animation frame by applying transformations to a base sprite.
// This ensures consistent sprite appearance across all frames, with only position/rotation/scale changing.
// V7 Performance fix: Uses pooled images and reusable DrawImageOptions to reduce allocations.
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
	// V7: Use pooled images instead of ebiten.NewImage
	outputWidth := config.Width + int(math.Abs(offset.X)*2) + 10
	outputHeight := config.Height + int(math.Abs(offset.Y)*2) + 10
	img := s.frameImagePool.Get(outputWidth, outputHeight)

	// V7: Reuse pre-allocated DrawImageOptions, reset GeoM for this frame
	s.transformDrawOpts.GeoM.Reset()
	s.transformDrawOpts.ColorScale.Reset()

	// Center sprite in output image
	centerX := float64(outputWidth) / 2
	centerY := float64(outputHeight) / 2

	// Apply scale around center
	if scale != 1.0 {
		s.transformDrawOpts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		s.transformDrawOpts.GeoM.Scale(scale, scale)
		s.transformDrawOpts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply rotation around center
	if rotation != 0 {
		s.transformDrawOpts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		s.transformDrawOpts.GeoM.Rotate(rotation)
		s.transformDrawOpts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply position offset and center in output
	s.transformDrawOpts.GeoM.Translate(centerX-float64(config.Width)/2+offset.X, centerY-float64(config.Height)/2+offset.Y)

	img.DrawImage(baseSprite, &s.transformDrawOpts)

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
		Type:           sprites.SpriteEntity,
		Width:          int(sprite.Width),
		Height:         int(sprite.Height),
		Seed:           anim.Seed,
		Complexity:     0.7,
		Palette:        nil,
		Custom:         make(map[string]interface{}),
		PaletteOptions: s.paletteOptions, // Phase 5.4: Use custom palette options if set
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

		// Pass equipment visual data for template pipeline overlay rendering
		s.attachEquipmentVisuals(config, entity)

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

	// Override entity type from CreatureVisualComponent if available.
	if cvComp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := cvComp.(*CreatureVisualComponent); ok && cv.Form != FormHumanoid {
			entityType = string(cv.Form)
		}
	}

	// Override with NPC role visual for role-specific humanoid templates
	if roleComp, ok := entity.GetComponent("npc_role_visual"); ok {
		if npcRole, ok := roleComp.(*NpcRoleVisualComponent); ok && npcRole.Role != "" {
			entityType = npcRole.Role
		}
	}

	config.Custom["entityType"] = entityType
	facing := s.determineFacingDirection(entity, anim)
	config.Custom["facing"] = facing

	// Propagate size class from CreatureVisualComponent for size-based anatomy scaling
	if cvComp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := cvComp.(*CreatureVisualComponent); ok && cv.SizeClass != "" {
			config.Custom["sizeClass"] = cv.SizeClass
		}
	}

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

// attachEquipmentVisuals reads the entity's EquipmentVisualComponent and
// EquipmentComponent to build a slice of sprites.EquipmentVisual, storing it
// in config.Custom["equipmentVisuals"] for the template pipeline's equipment
// overlay renderer.
func (s *AnimationSystem) attachEquipmentVisuals(config *sprites.Config, entity *Entity) {
	equipVisComp, ok := entity.GetComponent("equipment_visual")
	if !ok {
		return
	}
	evc, ok := equipVisComp.(*EquipmentVisualComponent)
	if !ok || evc == nil {
		return
	}

	equipComp, _ := entity.GetComponent("equipment")
	ec, _ := equipComp.(*EquipmentComponent)

	var visuals []sprites.EquipmentVisual

	if evc.HasWeapon() && evc.ShowWeapon {
		v := sprites.EquipmentVisual{
			Slot:   "weapon",
			ItemID: evc.WeaponID,
			Seed:   evc.WeaponSeed,
			Layer:  sprites.LayerWeapon,
		}
		s.enrichEquipmentVisual(&v, ec, SlotMainHand)
		visuals = append(visuals, v)
	}

	if evc.HasArmor() && evc.ShowArmor {
		v := sprites.EquipmentVisual{
			Slot:   "armor",
			ItemID: evc.ArmorID,
			Seed:   evc.ArmorSeed,
			Layer:  sprites.LayerArmor,
		}
		s.enrichEquipmentVisual(&v, ec, SlotChest)
		visuals = append(visuals, v)
	}

	if evc.HasAccessories() && evc.ShowAccessories {
		for i, accID := range evc.AccessoryIDs {
			accSlot := SlotAccessory1
			if i == 1 {
				accSlot = SlotAccessory2
			} else if i >= 2 {
				accSlot = SlotAccessory3
			}
			v := sprites.EquipmentVisual{
				Slot:   "accessory",
				ItemID: accID,
				Seed:   evc.AccessorySeeds[i],
				Layer:  sprites.LayerAccessory,
			}
			s.enrichEquipmentVisual(&v, ec, accSlot)
			visuals = append(visuals, v)
		}
	}

	if len(visuals) > 0 {
		config.Custom["equipmentVisuals"] = visuals
	}
}

// enrichEquipmentVisual populates material, damage state, and enchantment
// from the actual equipped item data when available.
func (s *AnimationSystem) enrichEquipmentVisual(v *sprites.EquipmentVisual, ec *EquipmentComponent, slot EquipmentSlot) {
	v.Material = sprites.MaterialMetal
	v.DamageState = sprites.DamageStatePristine
	v.Enchantment = sprites.EnchantmentGlow{Active: false}
	v.DetailLevel = 0.5

	if ec == nil {
		return
	}
	itm := ec.GetEquipped(slot)
	if itm == nil {
		return
	}

	v.Material = sprites.GetMaterialTypeFromTags(itm.Tags, "")
	v.DamageState = sprites.GetDamageStateFromDurability(itm.Stats.Durability, itm.Stats.DurabilityMax)
	v.Enchantment = sprites.GetEnchantmentFromRarity(itm.Rarity.String())
	v.DetailLevel = sprites.GetDetailLevelFromRarity(itm.Rarity.String())
}

// determineFacingDirection calculates entity facing direction from velocity.
// Uses typed getter for ~91x faster access vs generic GetComponent.
func (s *AnimationSystem) determineFacingDirection(entity *Entity, anim *AnimationComponent) string {
	facing := "down"

	vel := entity.GetVelocity()
	if vel == nil {
		if anim.LastFacing != "" {
			return anim.LastFacing
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
func (s *AnimationSystem) cacheFrames(key uint64, frames []*ebiten.Image) {
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
		// V7: Return evicted images to pool, then return slice to pool
		if evictedFrames, ok := s.frameCache[oldestKey]; ok {
			for _, img := range evictedFrames {
				s.frameImagePool.Put(img)
			}
			s.putFrameSlice(evictedFrames)
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

// stateToInt converts an AnimationState to its integer representation.
// This enables zero-allocation cache keys by using uint64 instead of strings.
func stateToInt(state AnimationState) uint8 {
	switch state {
	case AnimationStateIdle:
		return 1
	case AnimationStateWalk:
		return 2
	case AnimationStateRun:
		return 3
	case AnimationStateAttack:
		return 4
	case AnimationStateCast:
		return 5
	case AnimationStateHit:
		return 6
	case AnimationStateDeath:
		return 7
	case AnimationStateJump:
		return 8
	case AnimationStateCrouch:
		return 9
	case AnimationStateUse:
		return 10
	default:
		return 255 // Unknown state
	}
}

// recordWarmerAccess records a cache access event in the predictive warmer.
func (s *AnimationSystem) recordWarmerAccess(key cache.CacheKey, hit bool) {
	if s.predictiveWarmer == nil {
		return
	}
	s.warmerTickCount++
	s.predictiveWarmer.RecordAccess(key, hit, s.warmerTickCount)
}

// getCacheKey generates a cache key for animation frames.
// Optimized: Uses uint64 combining seed + state for zero-allocation lookups.
// Layout: upper 56 bits = seed (int64), lower 8 bits = state ID (uint8)
func (s *AnimationSystem) getCacheKey(seed int64, state AnimationState) uint64 {
	stateID := stateToInt(state)
	// Combine: shift seed left 8 bits, OR with state ID in lower 8 bits
	// stateToInt returns 1-based IDs (1–10 for known states, 255 for unknown),
	// guaranteeing the key is never zero even when seed=0.
	key := (uint64(seed) << 8) | uint64(stateID)
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"seed":      seed,
			"state":     state,
			"state_id":  stateID,
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

	// V7: Return all cached images to pool before clearing
	for _, frames := range s.frameCache {
		for _, img := range frames {
			s.frameImagePool.Put(img)
		}
	}

	s.frameCache = make(map[uint64][]*ebiten.Image)
	s.cacheKeys = make([]uint64, 0, s.maxCacheSize)

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

// getAnimationComponent retrieves entity animation component.
// Uses typed getter for ~91x faster access vs generic GetComponent.
func (s *AnimationSystem) getAnimationComponent(entity *Entity) *AnimationComponent {
	return entity.GetAnimation()
}

// getSpriteComponent retrieves entity sprite component.
// Uses cached GetSprite() getter for ~36x faster access vs generic GetComponent + type assertion.
func (s *AnimationSystem) getSpriteComponent(entity *Entity) *EbitenSprite {
	return entity.GetSprite()
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

// generateSpriteCacheKey creates a cache key from a sprite config.
// The key uniquely identifies a sprite based on its generation parameters.
// Format: "sprite:type:seed:genre:width:height:complexity:variation:customHash"
// The customHash includes sprite-affecting Custom field values: entityType,
// facing, hasWeapon, hasShield, isBoss, bossScale, useAerial, and genre.
// Optimized: Uses strings.Builder instead of fmt.Sprintf to avoid allocations.
func (s *AnimationSystem) generateSpriteCacheKey(config sprites.Config) cache.CacheKey {
	// Build a deterministic string from relevant Custom fields that affect sprite appearance
	customKey := s.buildCustomFieldsKey(config.Custom, config.GenreID)

	// Include all relevant config fields that affect sprite generation
	// Type is included to prevent cache collisions between different sprite categories
	// Pre-allocate approximate capacity: "sprite:" + type(~10) + seed(~20) + genre(~10)
	// + width/height(~8) + complexity(~6) + variation(~4) + customKey + separators
	var b strings.Builder
	b.Grow(80 + len(customKey))

	b.WriteString("sprite:")
	b.WriteString(config.Type.String())
	b.WriteByte(':')
	b.Write(strconv.AppendInt(nil, config.Seed, 10))
	b.WriteByte(':')
	b.WriteString(config.GenreID)
	b.WriteByte(':')
	b.Write(strconv.AppendInt(nil, int64(config.Width), 10))
	b.WriteByte(':')
	b.Write(strconv.AppendInt(nil, int64(config.Height), 10))
	b.WriteByte(':')
	// Format complexity as fixed-point with 2 decimal places (e.g., 0.75 -> "75")
	// This avoids float formatting overhead while maintaining precision
	b.Write(strconv.AppendInt(nil, int64(config.Complexity*100), 10))
	b.WriteByte(':')
	b.Write(strconv.AppendInt(nil, int64(config.Variation), 10))
	b.WriteByte(':')
	b.WriteString(customKey)

	return cache.CacheKey(b.String())
}

// buildCustomFieldsKey creates a deterministic key string from Custom field values
// that affect sprite appearance. This ensures sprites with different entity types,
// facing directions, equipment, or boss configurations get unique cache entries.
// The configGenreID parameter is used to avoid duplicating genre in the key when
// the Custom["genre"] matches the config.GenreID.
func (s *AnimationSystem) buildCustomFieldsKey(custom map[string]interface{}, configGenreID string) string {
	if custom == nil {
		return ""
	}

	parts := s.extractCustomFieldParts(custom, configGenreID)

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "|")
}

// extractCustomFieldParts extracts sprite-affecting fields in deterministic order.
func (s *AnimationSystem) extractCustomFieldParts(custom map[string]interface{}, configGenreID string) []string {
	var parts []string

	parts = s.appendEntityType(parts, custom)
	parts = s.appendFacing(parts, custom)
	parts = s.appendEquipmentFlags(parts, custom)
	parts = s.appendBossConfiguration(parts, custom)
	parts = s.appendAerialFlag(parts, custom)
	parts = s.appendGenreIfDifferent(parts, custom, configGenreID)

	return parts
}

// appendEntityType appends entity type to parts if present.
func (s *AnimationSystem) appendEntityType(parts []string, custom map[string]interface{}) []string {
	if entityType, ok := custom["entityType"].(string); ok {
		parts = append(parts, "et:"+entityType)
	}
	return parts
}

// appendFacing appends facing direction to parts if present.
func (s *AnimationSystem) appendFacing(parts []string, custom map[string]interface{}) []string {
	if facing, ok := custom["facing"].(string); ok {
		parts = append(parts, "f:"+facing)
	}
	return parts
}

// appendEquipmentFlags appends weapon and shield flags to parts if present.
func (s *AnimationSystem) appendEquipmentFlags(parts []string, custom map[string]interface{}) []string {
	if hasWeapon, ok := custom["hasWeapon"].(bool); ok && hasWeapon {
		parts = append(parts, "w:1")
	}
	if hasShield, ok := custom["hasShield"].(bool); ok && hasShield {
		parts = append(parts, "s:1")
	}
	return parts
}

// appendBossConfiguration appends boss flag and scale to parts if present.
func (s *AnimationSystem) appendBossConfiguration(parts []string, custom map[string]interface{}) []string {
	if isBoss, ok := custom["isBoss"].(bool); ok && isBoss {
		parts = append(parts, "boss:1")
		if bossScale, ok := custom["bossScale"].(float64); ok {
			parts = append(parts, "bs:"+strconv.FormatInt(int64(bossScale*10), 10))
		}
	}
	return parts
}

// appendAerialFlag appends aerial sprite flag to parts if present.
func (s *AnimationSystem) appendAerialFlag(parts []string, custom map[string]interface{}) []string {
	if useAerial, ok := custom["useAerial"].(bool); ok && useAerial {
		parts = append(parts, "a:1")
	}
	return parts
}

// appendGenreIfDifferent appends genre to parts if different from config genre.
func (s *AnimationSystem) appendGenreIfDifferent(parts []string, custom map[string]interface{}, configGenreID string) []string {
	if genre, ok := custom["genre"].(string); ok && genre != configGenreID {
		parts = append(parts, "g:"+genre)
	}
	return parts
}

// getFrameSlice retrieves a frame slice from the pool and resizes it to the needed count.
// The slice will have the exact length needed and adequate capacity.
func (s *AnimationSystem) getFrameSlice(count int) []*ebiten.Image {
	slice := s.frameSlicePool.Get().([]*ebiten.Image)
	// Reset length but preserve capacity, then resize to exact count
	slice = slice[:0]
	if cap(slice) < count {
		// If capacity insufficient, make new slice with headroom
		slice = make([]*ebiten.Image, count, count+8)
	} else {
		// Resize to exact count within existing capacity
		slice = slice[:count]
	}
	return slice
}

// putFrameSlice returns a frame slice to the pool for reuse.
// The slice is cleared before returning to prevent memory leaks.
func (s *AnimationSystem) putFrameSlice(slice []*ebiten.Image) {
	if slice == nil {
		return
	}
	// Clear references to allow GC
	for i := range slice {
		slice[i] = nil
	}
	slice = slice[:0]
	s.frameSlicePool.Put(slice)
}

// animationImagePool manages pools of Ebiten images for animation frames.
// V7 Performance fix: Reduces GPU texture allocations during animation frame generation.
// Uses bucketed sizes (64, 80, 96, 128, 160) to balance memory efficiency with reuse rate.
// Animation frames typically need ~74-84 pixels for 64px sprites with transformations.
type animationImagePool struct {
	pool64  sync.Pool // For sizes up to 64
	pool80  sync.Pool // For sizes up to 80 (typical animation frame size)
	pool96  sync.Pool // For sizes up to 96
	pool128 sync.Pool // For sizes up to 128
	pool160 sync.Pool // For sizes up to 160 (large sprites with transforms)
}

// newAnimationImagePool creates a new animation frame image pool.
func newAnimationImagePool() *animationImagePool {
	p := &animationImagePool{}

	p.pool64.New = func() interface{} {
		return ebiten.NewImage(64, 64)
	}
	p.pool80.New = func() interface{} {
		return ebiten.NewImage(80, 80)
	}
	p.pool96.New = func() interface{} {
		return ebiten.NewImage(96, 96)
	}
	p.pool128.New = func() interface{} {
		return ebiten.NewImage(128, 128)
	}
	p.pool160.New = func() interface{} {
		return ebiten.NewImage(160, 160)
	}

	return p
}

// getBucket returns the appropriate bucket size for the given dimensions.
// Rounds up to the nearest bucket to maximize reuse.
func (p *animationImagePool) getBucket(width, height int) int {
	maxDim := width
	if height > maxDim {
		maxDim = height
	}
	switch {
	case maxDim <= 64:
		return 64
	case maxDim <= 80:
		return 80
	case maxDim <= 96:
		return 96
	case maxDim <= 128:
		return 128
	case maxDim <= 160:
		return 160
	default:
		return 0 // No bucket available, will create non-pooled image
	}
}

// Get retrieves an image from the pool, sized to at least the requested dimensions.
// Returns a pooled image for common sizes, or creates a new one for non-standard sizes.
// The caller must ensure the image is cleared before use (Clear() is called by Put).
func (p *animationImagePool) Get(width, height int) *ebiten.Image {
	bucket := p.getBucket(width, height)
	switch bucket {
	case 64:
		return p.pool64.Get().(*ebiten.Image)
	case 80:
		return p.pool80.Get().(*ebiten.Image)
	case 96:
		return p.pool96.Get().(*ebiten.Image)
	case 128:
		return p.pool128.Get().(*ebiten.Image)
	case 160:
		return p.pool160.Get().(*ebiten.Image)
	default:
		// Non-standard size: create new image (not pooled)
		return ebiten.NewImage(width, height)
	}
}

// Put returns an image to the appropriate pool.
// The image is cleared before being returned to the pool.
// Only images matching bucket sizes are pooled; others are left for GC.
func (p *animationImagePool) Put(img *ebiten.Image) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Clear the image before returning to pool
	img.Clear()

	// Only pool images that match our bucket sizes exactly
	if width == height {
		switch width {
		case 64:
			p.pool64.Put(img)
			return
		case 80:
			p.pool80.Put(img)
			return
		case 96:
			p.pool96.Put(img)
			return
		case 128:
			p.pool128.Put(img)
			return
		case 160:
			p.pool160.Put(img)
			return
		}
	}
	// Non-standard size: let it be garbage collected
}
