package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/animation"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// AnimationAdapter wraps pkg/rendering/animation.Controller for advanced animation features.
// Provides Phase 2.2 articulation, direction-based animation, and enhanced caching.
// This adapter enhances the existing AnimationSystem with the advanced features
// from pkg/rendering/animation without replacing it.
type AnimationAdapter struct {
	controller *animation.Controller
	enabled    bool
	logger     *logrus.Entry
}

// NewAnimationAdapter creates an animation adapter with the default cache.
func NewAnimationAdapter(generator *sprites.Generator, logger *logrus.Entry) *AnimationAdapter {
	return NewAnimationAdapterWithCache(generator, nil, logger)
}

// NewAnimationAdapterWithCache creates an animation adapter with a custom cache.
// Pass nil to use the default 50 MB / 1000-entry cache.
func NewAnimationAdapterWithCache(generator *sprites.Generator, cache *animation.AnimationCache, logger *logrus.Entry) *AnimationAdapter {
	var ctrl *animation.Controller
	if cache == nil {
		ctrl = animation.NewController(generator)
	} else {
		ctrl = animation.NewControllerWithCache(generator, cache)
	}
	return &AnimationAdapter{
		controller: ctrl,
		enabled:    true,
		logger:     logger,
	}
}

// SetEnabled enables or disables the advanced animation features.
func (a *AnimationAdapter) SetEnabled(enabled bool) {
	a.enabled = enabled
	if a.logger != nil {
		a.logger.WithField("enabled", enabled).Debug("animation adapter enabled state changed")
	}
}

// IsEnabled returns whether advanced animation features are enabled.
func (a *AnimationAdapter) IsEnabled() bool {
	return a.enabled
}

// SetArticulationConfig updates the articulation configuration.
func (a *AnimationAdapter) SetArticulationConfig(config animation.ArticulationConfig) {
	a.controller.SetArticulationConfig(config)
}

// GenerateFrame generates a single animation frame with articulation and direction.
func (a *AnimationAdapter) GenerateFrame(seed int64, state string, frameIndex, frameCount int, direction animation.Direction8, spriteConfig sprites.Config) (*ebiten.Image, error) {
	if !a.enabled {
		return nil, nil
	}

	return a.controller.GenerateFrame(seed, state, frameIndex, frameCount, direction, spriteConfig)
}

// GenerateSequence generates a complete animation sequence (all frames).
func (a *AnimationAdapter) GenerateSequence(seed int64, state string, direction animation.Direction8, spriteConfig sprites.Config) ([]*ebiten.Image, error) {
	if !a.enabled {
		return nil, nil
	}

	return a.controller.GenerateSequence(seed, state, direction, spriteConfig)
}

// ClearCache clears the animation cache.
func (a *AnimationAdapter) ClearCache() {
	a.controller.ClearCache()
}

// GetCacheStats returns cache hit rate and other statistics.
func (a *AnimationAdapter) GetCacheStats() (hitRate float64, cacheSize int) {
	stats := a.controller.GetCacheStats()
	return stats.HitRate(), stats.EntryCount
}

// Update implements System interface for ECS integration.
// This system doesn't need per-frame updates as it's used on-demand during rendering.
func (a *AnimationAdapter) Update(entities []*Entity, deltaTime float64) {
	// No per-frame updates needed - animation generation is on-demand
}
