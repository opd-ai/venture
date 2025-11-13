// Package engine provides animation system for updating entity animations.
package engine

import (
	"fmt"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// AnimationSystem updates animation components and manages frame transitions.
// Integrates with sprite generator to create procedural animation frames.
type AnimationSystem struct {
	spriteGenerator *sprites.Generator
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
}

// SetPlayerEntity sets the player entity reference for distance calculations.
// Call this during initialization to enable distance-based frame rate adjustment.
func (s *AnimationSystem) SetPlayerEntity(player *Entity) {
	s.playerEntity = player
}

// EnableViewportCulling enables or disables viewport culling optimization.
// When enabled, only animates entities visible in the current viewport.
func (s *AnimationSystem) EnableViewportCulling(enable bool) {
	s.enableViewportCull = enable
}

// EnableDistanceLOD enables or disables distance-based level-of-detail.
// When enabled, adjusts animation frame rate based on distance from player.
func (s *AnimationSystem) EnableDistanceLOD(enable bool) {
	s.enableDistanceLOD = enable
}

// SetDistanceThresholds sets the distance thresholds for LOD tiers.
// closeThreshold: Full animation rate (default 200px)
// midThreshold: Half animation rate (default 400px)
// Beyond midThreshold: Static pose (no animation updates)
func (s *AnimationSystem) SetDistanceThresholds(closeThreshold, midThreshold float64) {
	s.distanceCloseThresh = closeThreshold
	s.distanceMidThresh = midThreshold
}

// SetMaxCacheSize sets the maximum number of animation sequences to cache.
// Larger cache reduces regeneration overhead but uses more memory.
// Default is 100. For WASM/browser environments, consider 200-500 for better performance.
// Each cached sequence typically uses 100-400KB depending on sprite size and frame count.
func (s *AnimationSystem) SetMaxCacheSize(maxSize int) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.maxCacheSize = maxSize

	// If new size is smaller than current cache, trigger eviction
	if len(s.frameCache) > maxSize {
		// Evict oldest entries until within limit
		toEvict := len(s.frameCache) - maxSize
		for i := 0; i < toEvict && len(s.cacheKeys) > 0; i++ {
			oldestKey := s.cacheKeys[0]
			delete(s.frameCache, oldestKey)
			s.cacheKeys = s.cacheKeys[1:]
		}
	}
}

// GetStats returns current animation performance statistics.
// Useful for monitoring and debugging performance.
func (s *AnimationSystem) GetStats() AnimationStats {
	return s.stats
}

// Update processes all entities with animation components.
// Updates frame timers, transitions states, and regenerates frames if needed.
func (s *AnimationSystem) Update(entities []*Entity, deltaTime float64) error {
	// DEBUG: Log when system runs
	fmt.Printf("[DEBUG] AnimationSystem.Update called with %d entities\n", len(entities))

	// Phase 14.2: Reset statistics for this frame
	s.stats = AnimationStats{
		TotalEntities: len(entities),
	}

	// Phase 14.2: Get player position for distance calculations
	var playerX, playerY float64
	if s.playerEntity != nil {
		if posComp, ok := s.playerEntity.GetComponent("position"); ok {
			if pos, ok := posComp.(*PositionComponent); ok {
				playerX = pos.X
				playerY = pos.Y
			}
		}
	}

	// Phase 14.2: Get viewport bounds for culling
	var viewportMinX, viewportMinY, viewportMaxX, viewportMaxY float64
	hasViewport := false
	if s.enableViewportCull && s.cameraSystem != nil && s.cameraSystem.activeCamera != nil {
		if camComp, ok := s.cameraSystem.activeCamera.GetComponent("camera"); ok {
			if camera, ok := camComp.(*CameraComponent); ok {
				// Calculate viewport bounds with margin for sprites
				margin := 100.0 // Extra margin to start animating before entity enters view
				halfWidth := float64(s.cameraSystem.ScreenWidth) / (2.0 * camera.Zoom)
				halfHeight := float64(s.cameraSystem.ScreenHeight) / (2.0 * camera.Zoom)
				viewportMinX = camera.X - halfWidth - margin
				viewportMinY = camera.Y - halfHeight - margin
				viewportMaxX = camera.X + halfWidth + margin
				viewportMaxY = camera.Y + halfHeight + margin
				hasViewport = true
			}
		}
	}

	for _, entity := range entities {
		// Get animation component
		animComp := s.getAnimationComponent(entity)
		if animComp == nil {
			continue
		}

		s.stats.AnimatedEntities++

		// Get sprite component for size information
		spriteComp := s.getSpriteComponent(entity)
		if spriteComp == nil {
			continue
		}

		// Get entity position for viewport and distance checks
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Phase 14.2: Viewport culling check
		if hasViewport {
			if pos.X < viewportMinX || pos.X > viewportMaxX ||
				pos.Y < viewportMinY || pos.Y > viewportMaxY {
				// Entity is outside viewport - skip animation update but keep current frame
				s.stats.CulledByViewport++
				continue
			}
		}

		// Phase 14.2: Distance-based frame rate adjustment
		// Phase 15.2: Enhanced LOD with explicit frame rates (12 FPS close, 6 FPS medium, 3 FPS far)
		var effectiveDeltaTime float64 = deltaTime
		var targetFrameTime float64 = animComp.FrameTime // Default frame time
		if s.enableDistanceLOD && s.playerEntity != nil {
			// Calculate distance from player
			dx := pos.X - playerX
			dy := pos.Y - playerY
			distSq := dx*dx + dy*dy
			dist := math.Sqrt(distSq)

			if dist <= s.distanceCloseThresh {
				// Close range: 12 FPS (1/12 ≈ 0.083s per frame)
				targetFrameTime = 1.0 / 12.0
				effectiveDeltaTime = deltaTime
				s.stats.FullRateEntities++
			} else if dist <= s.distanceMidThresh {
				// Mid range: 6 FPS (1/6 ≈ 0.167s per frame)
				targetFrameTime = 1.0 / 6.0
				effectiveDeltaTime = deltaTime
				s.stats.HalfRateEntities++
			} else {
				// Far range: 3 FPS (1/3 ≈ 0.333s per frame) - very slow for distant entities
				targetFrameTime = 1.0 / 3.0
				effectiveDeltaTime = deltaTime
				s.stats.StaticEntities++
			}

			// Update component's frame time for this frame
			animComp.FrameTime = targetFrameTime
		}

		// Regenerate frames if dirty (state changed)
		if animComp.Dirty {
			// Only log for player entities at debug level
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

			// DEBUG: Always log for player
			if entity.HasComponent("input") {
				fmt.Printf("[DEBUG] AnimationSystem: Regenerating frames for player (state=%v)\n", animComp.CurrentState)
			}

			if err := s.regenerateFrames(entity, animComp, spriteComp); err != nil {
				return fmt.Errorf("failed to regenerate frames: %w", err)
			}
			animComp.Dirty = false

			// DEBUG: Log result for player
			if entity.HasComponent("input") {
				fmt.Printf("[DEBUG] AnimationSystem: Player frames generated. Count=%d, FirstFrame nil=%v\n",
					len(animComp.Frames), animComp.Frames == nil || len(animComp.Frames) == 0 || animComp.Frames[0] == nil)
			}

			// Verify frames were generated (debug level)
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

		// Update animation if playing and not static (far away)
		if animComp.Playing && len(animComp.Frames) > 0 && effectiveDeltaTime > 0 {
			s.updateFrame(animComp, effectiveDeltaTime)
		}

		// Update sprite component with current frame
		if frame := animComp.CurrentFrame(); frame != nil {
			spriteComp.Image = frame
			// DEBUG: Log for player
			if entity.HasComponent("input") {
				fmt.Printf("[DEBUG] AnimationSystem: Updated player sprite.Image (frame %d, nil=%v)\n",
					animComp.FrameIndex, frame == nil)
			}
		} else if entity.HasComponent("input") {
			fmt.Printf("[DEBUG] AnimationSystem: WARNING - Player CurrentFrame() returned NIL (frameIndex=%d, frameCount=%d)\n",
				animComp.FrameIndex, len(animComp.Frames))
		}
	}

	return nil
}

// updateFrame advances the animation frame based on delta time.
func (s *AnimationSystem) updateFrame(anim *AnimationComponent, deltaTime float64) {
	anim.TimeAccumulator += deltaTime

	// Check if it's time to advance frame
	if anim.TimeAccumulator >= anim.FrameTime {
		anim.TimeAccumulator -= anim.FrameTime
		anim.FrameIndex++

		// Handle loop or completion
		if anim.FrameIndex >= len(anim.Frames) {
			if anim.Loop {
				anim.FrameIndex = 0
			} else {
				anim.FrameIndex = len(anim.Frames) - 1
				anim.Playing = false
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
		return nil
	}
	s.cacheMutex.RUnlock()

	// Generate frames using sprite generator
	frames, err := s.generateFrames(entity, anim, sprite)
	if err != nil {
		return err
	}

	// Cache frames
	s.cacheFrames(cacheKey, frames)
	anim.Frames = frames

	return nil
}

// generateFrames creates animation frames using the sprite generator.
func (s *AnimationSystem) generateFrames(entity *Entity, anim *AnimationComponent, sprite *EbitenSprite) ([]*ebiten.Image, error) {
	// Determine frame count based on animation state
	frameCount := s.getFrameCount(anim.CurrentState)
	if anim.FrameCount > 0 {
		frameCount = anim.FrameCount
	}

	frames := make([]*ebiten.Image, frameCount)

	// Get sprite configuration from entity
	config := s.buildSpriteConfig(entity, sprite, anim)

	// CRITICAL FIX: Generate the base sprite ONCE, then transform it for each frame
	// This prevents the "mutating shapes" issue where each frame is a different sprite
	baseSprite, err := s.spriteGenerator.Generate(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base sprite: %w", err)
	}

	// Now generate each animation frame by applying transformations to the base sprite
	for i := 0; i < frameCount; i++ {
		frame, err := s.generateTransformedFrame(baseSprite, config, string(anim.CurrentState), i, frameCount)
		if err != nil {
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
		Complexity: 0.7, // Higher complexity for better detail (was 0.5)
		Palette:    nil, // Will be generated by sprite generator if nil
		Custom:     make(map[string]interface{}),
	}

	// Phase 10.1: Enable aerial-view sprites for top-down gameplay
	// Aerial templates use 35/50/15 proportions (head/torso/legs) optimized for overhead view
	config.Custom["useAerial"] = true

	// CRITICAL: Set entity type to trigger template-based generation
	// Check if entity has input component (player) or team component (enemy/NPC)
	if entity.HasComponent("input") {
		// Player character - use humanoid template
		config.Custom["entityType"] = "humanoid"

		// GAP FIX: Determine facing direction based on velocity
		facing := "down" // Default
		if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
			if vel, ok := velComp.(*VelocityComponent); ok {
				// Use velocity direction if moving, otherwise keep last facing
				if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
					if math.Abs(vel.VX) > math.Abs(vel.VY) {
						if vel.VX > 0 {
							facing = "right"
						} else {
							facing = "left"
						}
					} else {
						if vel.VY > 0 {
							facing = "down"
						} else {
							facing = "up"
						}
					}
					// Store facing for idle state
					anim.LastFacing = facing
				} else if anim.LastFacing != "" {
					// Use last facing direction when idle
					facing = anim.LastFacing
				}
			}
		}
		config.Custom["facing"] = facing

		// Check for equipment to show on sprite
		if entity.HasComponent("equipment") {
			config.Custom["hasWeapon"] = true
			config.Custom["hasShield"] = false // Could be enhanced to check actual equipment
		}
	} else if teamComp, ok := entity.GetComponent("team"); ok {
		if team, ok := teamComp.(*TeamComponent); ok {
			if team.TeamID == 2 { // Enemy team
				// Determine monster type based on entity characteristics
				entityType := "humanoid" // Default

				// Check if it's a boss (high damage indicates boss)
				if attackComp, ok := entity.GetComponent("attack"); ok {
					if attack, ok := attackComp.(*AttackComponent); ok {
						if attack.Damage > 20 {
							entityType = "boss"
							config.Custom["isBoss"] = true
							config.Custom["bossScale"] = 1.5
						}
					}
				}

				// Check size based on collider
				if colliderComp, ok := entity.GetComponent("collider"); ok {
					if collider, ok := colliderComp.(*ColliderComponent); ok {
						if collider.Width > 48 {
							entityType = "monster" // Large monster
						} else if collider.Width < 24 {
							entityType = "minion" // Small creature
						}
					}
				}

				config.Custom["entityType"] = entityType

				// GAP FIX: Determine facing direction based on velocity
				facing := "down" // Default
				if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
					if vel, ok := velComp.(*VelocityComponent); ok {
						// Use velocity direction if moving, otherwise keep last facing
						if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
							if math.Abs(vel.VX) > math.Abs(vel.VY) {
								if vel.VX > 0 {
									facing = "right"
								} else {
									facing = "left"
								}
							} else {
								if vel.VY > 0 {
									facing = "down"
								} else {
									facing = "up"
								}
							}
							// Store facing for idle state
							anim.LastFacing = facing
						} else if anim.LastFacing != "" {
							// Use last facing direction when idle
							facing = anim.LastFacing
						}
					}
				}
				config.Custom["facing"] = facing
			}
		}
	}

	// Get genre from entity if available
	if genreComp, ok := entity.GetComponent("genre"); ok && genreComp != nil {
		if gc, ok := genreComp.(interface{ GetGenreID() string }); ok {
			config.GenreID = gc.GetGenreID()
			config.Custom["genre"] = gc.GetGenreID()
		}
	}

	// Try to get genreID from world or use default
	if config.GenreID == "" {
		config.GenreID = "fantasy" // Default genre
		config.Custom["genre"] = "fantasy"
	}

	// Enable aerial/top-down view for all animated sprites by default
	// Provides better perspective for action-RPG gameplay
	config.Custom["useAerial"] = true

	return config
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
		delete(s.frameCache, oldestKey)
		s.cacheKeys = s.cacheKeys[1:]
	}

	// Add to cache
	s.frameCache[key] = frames
	s.cacheKeys = append(s.cacheKeys, key)
}

// getCacheKey generates a cache key for animation frames.
func (s *AnimationSystem) getCacheKey(seed int64, state AnimationState) string {
	return fmt.Sprintf("%d_%s", seed, state)
}

// ClearCache clears the animation frame cache.
func (s *AnimationSystem) ClearCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.frameCache = make(map[string][]*ebiten.Image)
	s.cacheKeys = make([]string, 0, s.maxCacheSize)
}

// GetCacheSize returns the current number of cached animation sequences.
func (s *AnimationSystem) GetCacheSize() int {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	return len(s.frameCache)
}

// Helper methods to get components

func (s *AnimationSystem) getAnimationComponent(entity *Entity) *AnimationComponent {
	comp, ok := entity.GetComponent("animation")
	if !ok || comp == nil {
		return nil
	}
	animComp, ok := comp.(*AnimationComponent)
	if !ok {
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
		return nil
	}
	return spriteComp
}

// TransitionState safely transitions an entity to a new animation state.
// Returns false if entity has no animation component.
func (s *AnimationSystem) TransitionState(entity *Entity, newState AnimationState) bool {
	animComp := s.getAnimationComponent(entity)
	if animComp == nil {
		return false
	}

	animComp.SetState(newState)
	return true
}
