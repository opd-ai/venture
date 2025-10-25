// Package engine provides animation system for updating entity animations.
package engine

import (
	"fmt"
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

		// Regenerate frames if dirty
		if animComp.Dirty {
			if err := s.regenerateFrames(entity, animComp, spriteComp); err != nil {
				return fmt.Errorf("failed to regenerate frames: %w", err)
			}
			animComp.Dirty = false
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

	// Generate each frame with state-specific variations
	for i := 0; i < frameCount; i++ {
		config.Variation = i
		frame, err := s.spriteGenerator.GenerateAnimationFrame(config, string(anim.CurrentState), i, frameCount)
		if err != nil {
			return nil, fmt.Errorf("frame %d generation failed: %w", i, err)
		}
		frames[i] = frame
	}

	return frames, nil
}

// buildSpriteConfig creates a sprite configuration from entity components.
func (s *AnimationSystem) buildSpriteConfig(entity *Entity, sprite *EbitenSprite, anim *AnimationComponent) sprites.Config {
	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      int(sprite.Width),
		Height:     int(sprite.Height),
		Seed:       anim.Seed,
		Complexity: 0.5, // Default complexity
	}

	// Get genre from entity if available
	if genreComp, ok := entity.GetComponent("genre"); ok && genreComp != nil {
		if gc, ok := genreComp.(interface{ GetGenreID() string }); ok {
			config.GenreID = gc.GetGenreID()
		}
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
