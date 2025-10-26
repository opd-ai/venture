// Package engine provides animation system for updating entity animations.
package engine

import (
	"fmt"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// AnimationSystem updates animation components and manages frame transitions.
// Integrates with sprite generator to create procedural animation frames.
type AnimationSystem struct {
	spriteGenerator *sprites.Generator
	frameCache      map[string][]*ebiten.Image // Cache by key: seed_state
	cacheMutex      sync.RWMutex
	maxCacheSize    int
	cacheKeys       []string // For LRU eviction
}

// NewAnimationSystem creates a new animation system.
func NewAnimationSystem(spriteGenerator *sprites.Generator) *AnimationSystem {
	return &AnimationSystem{
		spriteGenerator: spriteGenerator,
		frameCache:      make(map[string][]*ebiten.Image),
		maxCacheSize:    100, // Cache up to 100 animation sequences
		cacheKeys:       make([]string, 0, 100),
	}
}

// Update processes all entities with animation components.
// Updates frame timers, transitions states, and regenerates frames if needed.
func (s *AnimationSystem) Update(entities []*Entity, deltaTime float64) error {
	for _, entity := range entities {
		// Get animation component
		animComp := s.getAnimationComponent(entity)
		if animComp == nil {
			continue
		}

		// Get sprite component for size information
		spriteComp := s.getSpriteComponent(entity)
		if spriteComp == nil {
			continue
		}

		// Regenerate frames if dirty (state changed)
		if animComp.Dirty {
			// DEBUG: Log animation frame generation
			if entity.HasComponent("input") { // Only log for player
				fmt.Printf("[ANIMATION] Entity %d: Generating %d frames for state=%s (sprite=%dx%d)\n",
					entity.ID, s.getFrameCount(animComp.CurrentState), animComp.CurrentState,
					int(spriteComp.Width), int(spriteComp.Height))
			}

			if err := s.regenerateFrames(entity, animComp, spriteComp); err != nil {
				return fmt.Errorf("failed to regenerate frames: %w", err)
			}
			animComp.Dirty = false

			// DEBUG: Verify frames were generated
			if entity.HasComponent("input") && len(animComp.Frames) > 0 {
				fmt.Printf("[ANIMATION] Entity %d: Successfully generated %d frames, now showing frame %d\n",
					entity.ID, len(animComp.Frames), animComp.FrameIndex)
			}
		}

		// Update animation if playing
		if animComp.Playing && len(animComp.Frames) > 0 {
			s.updateFrame(animComp, deltaTime)
		}

		// Update sprite component with current frame
		if frame := animComp.CurrentFrame(); frame != nil {
			spriteComp.Image = frame
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
	case "walk", "run":
		// Bobbing motion - increase amplitude for visibility
		cycle := math.Sin(t * 2 * math.Pi)
		offset.Y = cycle * 4.0 // Increased from 2.0 to 4.0 pixels

	case "jump":
		// Parabolic arc
		offset.Y = -4.0 * (t - t*t) * 15.0 // Jump up and down

	case "attack":
		// Forward lunge - increase amplitude for visibility
		if t < 0.5 {
			offset.X = t * 8.0 // Increased from 4.0 to 8.0 pixels
		} else {
			offset.X = (1.0 - t) * 8.0
		}

	case "hit":
		// Knockback
		offset.X = -(1.0 - t) * 6.0 // Increased from 3.0 to 6.0 pixels

	case "death":
		// Fall down
		offset.Y = t * 12.0 // Increased from 8.0 to 12.0 pixels
	}

	return offset
}

// calculateAnimationRotation computes rotation for animation frame.
func calculateAnimationRotation(state string, frameIndex, frameCount int) float64 {
	t := float64(frameIndex) / float64(frameCount)

	switch state {
	case "attack":
		// Swing arc
		if t < 0.3 {
			return -t * 0.5 // Wind up
		} else if t < 0.6 {
			return (t - 0.3) * 1.5 // Swing through
		} else {
			return (1.0 - t) * 0.3 // Follow through
		}

	case "death":
		// Rotate while falling
		return t * math.Pi / 2 // 90 degree rotation

	case "cast":
		// Gentle sway
		return math.Sin(t*2*math.Pi) * 0.1
	}

	return 0
}

// calculateAnimationScale computes scale factor for animation frame.
func calculateAnimationScale(state string, frameIndex, frameCount int) float64 {
	t := float64(frameIndex) / float64(frameCount)

	switch state {
	case "jump":
		// Squash and stretch
		if t < 0.2 {
			return 1.0 - t*0.5 // Squash before jump
		} else if t < 0.8 {
			return 0.9 + (t-0.2)*0.3 // Stretch during jump
		} else {
			return 1.0 - (t-0.8)*0.5 // Squash on landing
		}

	case "hit":
		// Squash on impact
		return 1.0 - t*0.2

	case "attack":
		// Slight scale up during strike
		if t > 0.3 && t < 0.6 {
			return 1.0 + (t-0.3)*0.3
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

	// CRITICAL: Set entity type to trigger template-based generation
	// Check if entity has input component (player) or team component (enemy/NPC)
	if entity.HasComponent("input") {
		// Player character - use humanoid template
		config.Custom["entityType"] = "humanoid"
		config.Custom["facing"] = "down" // Default facing direction

		// Check for equipment to show on sprite
		if entity.HasComponent("equipment") {
			config.Custom["hasWeapon"] = true
			config.Custom["hasShield"] = false // Could be enhanced to check actual equipment
		}
	} else if teamComp, ok := entity.GetComponent("team"); ok {
		team := teamComp.(*TeamComponent)
		if team.TeamID == 2 { // Enemy team
			// Determine monster type based on entity characteristics
			entityType := "humanoid" // Default

			// Check if it's a boss (high damage indicates boss)
			if attackComp, ok := entity.GetComponent("attack"); ok {
				attack := attackComp.(*AttackComponent)
				if attack.Damage > 20 {
					entityType = "boss"
					config.Custom["isBoss"] = true
					config.Custom["bossScale"] = 1.5
				}
			}

			// Check size based on collider
			if colliderComp, ok := entity.GetComponent("collider"); ok {
				collider := colliderComp.(*ColliderComponent)
				if collider.Width > 48 {
					entityType = "monster" // Large monster
				} else if collider.Width < 24 {
					entityType = "minion" // Small creature
				}
			}

			config.Custom["entityType"] = entityType
			config.Custom["facing"] = "down"
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

	return config
}

// getFrameCount returns the number of frames for an animation state.
func (s *AnimationSystem) getFrameCount(state AnimationState) int {
	switch state {
	case AnimationStateIdle:
		return 4
	case AnimationStateWalk:
		return 8 // 8-frame walk cycle
	case AnimationStateRun:
		return 8
	case AnimationStateAttack:
		return 6 // Wind-up, strike, follow-through
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
