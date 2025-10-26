package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestAnimationSystem_NewAnimationSystem tests system initialization.
func TestAnimationSystem_NewAnimationSystem(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	if sys == nil {
		t.Fatal("NewAnimationSystem returned nil")
	}

	if sys.spriteGenerator == nil {
		t.Error("Expected sprite generator to be set")
	}

	if sys.frameCache == nil {
		t.Error("Expected frame cache to be initialized")
	}

	if sys.maxCacheSize != 100 {
		t.Errorf("Expected max cache size 100, got %d", sys.maxCacheSize)
	}
}

// TestAnimationSystem_GetFrameCount tests frame count determination.
func TestAnimationSystem_GetFrameCount(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	tests := []struct {
		state    AnimationState
		expected int
	}{
		{AnimationStateIdle, 4},
		{AnimationStateWalk, 8},
		{AnimationStateRun, 8},
		{AnimationStateAttack, 6},
		{AnimationStateCast, 8},
		{AnimationStateHit, 3},
		{AnimationStateDeath, 6},
		{AnimationStateJump, 4},
		{AnimationStateCrouch, 2},
		{AnimationStateUse, 4},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			count := sys.getFrameCount(tt.state)
			if count != tt.expected {
				t.Errorf("Expected %d frames for %s, got %d", tt.expected, tt.state, count)
			}
		})
	}
}

// TestAnimationSystem_GetCacheKey tests cache key generation.
func TestAnimationSystem_GetCacheKey(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	key1 := sys.getCacheKey(12345, AnimationStateWalk)
	key2 := sys.getCacheKey(12345, AnimationStateWalk)
	key3 := sys.getCacheKey(12345, AnimationStateRun)
	key4 := sys.getCacheKey(54321, AnimationStateWalk)

	// Same seed and state should produce same key
	if key1 != key2 {
		t.Error("Expected identical keys for same seed and state")
	}

	// Different state should produce different key
	if key1 == key3 {
		t.Error("Expected different keys for different states")
	}

	// Different seed should produce different key
	if key1 == key4 {
		t.Error("Expected different keys for different seeds")
	}
}

// TestAnimationSystem_CacheFrames tests frame caching.
func TestAnimationSystem_CacheFrames(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	frames := []*ebiten.Image{
		ebiten.NewImage(28, 28),
		ebiten.NewImage(28, 28),
	}

	key := "test_key"
	sys.cacheFrames(key, frames)

	// Check cache size
	if sys.GetCacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", sys.GetCacheSize())
	}

	// Retrieve from cache
	sys.cacheMutex.RLock()
	cached, exists := sys.frameCache[key]
	sys.cacheMutex.RUnlock()

	if !exists {
		t.Error("Expected frames to be cached")
	}

	if len(cached) != len(frames) {
		t.Errorf("Expected %d cached frames, got %d", len(frames), len(cached))
	}
}

// TestAnimationSystem_CacheEviction tests LRU cache eviction.
func TestAnimationSystem_CacheEviction(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)
	sys.maxCacheSize = 3 // Small cache for testing

	frames := []*ebiten.Image{ebiten.NewImage(28, 28)}

	// Fill cache
	sys.cacheFrames("key1", frames)
	sys.cacheFrames("key2", frames)
	sys.cacheFrames("key3", frames)

	if sys.GetCacheSize() != 3 {
		t.Errorf("Expected cache size 3, got %d", sys.GetCacheSize())
	}

	// Add one more - should evict oldest
	sys.cacheFrames("key4", frames)

	if sys.GetCacheSize() != 3 {
		t.Errorf("Expected cache size 3 after eviction, got %d", sys.GetCacheSize())
	}

	// key1 should be evicted
	sys.cacheMutex.RLock()
	_, exists := sys.frameCache["key1"]
	sys.cacheMutex.RUnlock()

	if exists {
		t.Error("Expected oldest entry to be evicted")
	}

	// key4 should exist
	sys.cacheMutex.RLock()
	_, exists = sys.frameCache["key4"]
	sys.cacheMutex.RUnlock()

	if !exists {
		t.Error("Expected newest entry to be cached")
	}
}

// TestAnimationSystem_ClearCache tests cache clearing.
func TestAnimationSystem_ClearCache(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	frames := []*ebiten.Image{ebiten.NewImage(28, 28)}
	sys.cacheFrames("key1", frames)
	sys.cacheFrames("key2", frames)

	if sys.GetCacheSize() != 2 {
		t.Errorf("Expected cache size 2, got %d", sys.GetCacheSize())
	}

	sys.ClearCache()

	if sys.GetCacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", sys.GetCacheSize())
	}
}

// TestAnimationSystem_UpdateFrame tests frame advancement logic.
func TestAnimationSystem_UpdateFrame(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	anim := NewAnimationComponent(12345)
	anim.Frames = []*ebiten.Image{
		ebiten.NewImage(28, 28),
		ebiten.NewImage(28, 28),
		ebiten.NewImage(28, 28),
	}
	anim.FrameTime = 0.1 // 100ms per frame
	anim.Playing = true
	anim.Loop = true

	// Update with small delta (should not advance)
	sys.updateFrame(anim, 0.05)
	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index 0, got %d", anim.FrameIndex)
	}

	// Update with enough delta to advance
	sys.updateFrame(anim, 0.06)
	if anim.FrameIndex != 1 {
		t.Errorf("Expected frame index 1, got %d", anim.FrameIndex)
	}

	// Advance to end and loop
	anim.FrameIndex = 2
	anim.TimeAccumulator = 0.0
	sys.updateFrame(anim, 0.1)

	if anim.FrameIndex != 0 {
		t.Errorf("Expected frame index to loop to 0, got %d", anim.FrameIndex)
	}
}

// TestAnimationSystem_UpdateFrame_NonLooping tests non-looping animation.
func TestAnimationSystem_UpdateFrame_NonLooping(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	callbackCalled := false
	anim := NewAnimationComponent(12345)
	anim.Frames = []*ebiten.Image{
		ebiten.NewImage(28, 28),
		ebiten.NewImage(28, 28),
	}
	anim.FrameTime = 0.1
	anim.Playing = true
	anim.Loop = false
	anim.OnComplete = func() {
		callbackCalled = true
	}

	// Advance to last frame
	anim.FrameIndex = 1
	anim.TimeAccumulator = 0.0
	sys.updateFrame(anim, 0.1)

	// Should stop at last frame
	if anim.FrameIndex != 1 {
		t.Errorf("Expected frame index to stay at 1, got %d", anim.FrameIndex)
	}

	if anim.Playing {
		t.Error("Expected animation to stop")
	}

	if !callbackCalled {
		t.Error("Expected OnComplete callback to be called")
	}
}

// TestAnimationSystem_GetComponents tests component retrieval.
func TestAnimationSystem_GetComponents(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	entity := NewEntity(1)

	// No components
	if anim := sys.getAnimationComponent(entity); anim != nil {
		t.Error("Expected nil animation component for entity without component")
	}

	if sprite := sys.getSpriteComponent(entity); sprite != nil {
		t.Error("Expected nil sprite component for entity without component")
	}

	// Add components
	animComp := NewAnimationComponent(12345)
	spriteComp := &EbitenSprite{
		Image:   ebiten.NewImage(28, 28),
		Width:   28,
		Height:  28,
		Visible: true,
	}

	entity.AddComponent(animComp)
	entity.AddComponent(spriteComp)

	// Retrieve components
	if anim := sys.getAnimationComponent(entity); anim == nil {
		t.Error("Expected non-nil animation component")
	}

	if sprite := sys.getSpriteComponent(entity); sprite == nil {
		t.Error("Expected non-nil sprite component")
	}
}

// TestAnimationSystem_TransitionState tests state transition method.
func TestAnimationSystem_TransitionState(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	entity := NewEntity(1)

	// No animation component
	if sys.TransitionState(entity, AnimationStateWalk) {
		t.Error("Expected TransitionState to return false for entity without animation")
	}

	// Add animation component
	animComp := NewAnimationComponent(12345)
	entity.AddComponent(animComp)

	// Transition state
	if !sys.TransitionState(entity, AnimationStateWalk) {
		t.Error("Expected TransitionState to return true")
	}

	if animComp.CurrentState != AnimationStateWalk {
		t.Errorf("Expected state %v, got %v", AnimationStateWalk, animComp.CurrentState)
	}
}

// BenchmarkAnimationSystem_UpdateFrame benchmarks frame updates.
func BenchmarkAnimationSystem_UpdateFrame(b *testing.B) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	anim := NewAnimationComponent(12345)
	anim.Frames = make([]*ebiten.Image, 8)
	for i := range anim.Frames {
		anim.Frames[i] = ebiten.NewImage(28, 28)
	}
	anim.Playing = true
	anim.Loop = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.updateFrame(anim, 0.016) // 60 FPS delta
	}
}

// BenchmarkAnimationSystem_CacheFrames benchmarks frame caching.
func BenchmarkAnimationSystem_CacheFrames(b *testing.B) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	frames := []*ebiten.Image{
		ebiten.NewImage(28, 28),
		ebiten.NewImage(28, 28),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := sys.getCacheKey(int64(i), AnimationStateWalk)
		sys.cacheFrames(key, frames)
	}
}
