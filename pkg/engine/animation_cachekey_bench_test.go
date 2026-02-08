package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkGetCacheKey_Uint64 benchmarks the optimized uint64 cache key generation.
func BenchmarkGetCacheKey_Uint64(b *testing.B) {
	sys := NewAnimationSystem(nil)
	seed := int64(12345)
	state := AnimationStateWalk

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.getCacheKey(seed, state)
	}
}

// BenchmarkGetCacheKey_VariousStates benchmarks key generation across different states.
func BenchmarkGetCacheKey_VariousStates(b *testing.B) {
	sys := NewAnimationSystem(nil)
	states := []AnimationState{
		AnimationStateIdle, AnimationStateWalk, AnimationStateRun,
		AnimationStateAttack, AnimationStateCast, AnimationStateHit,
		AnimationStateDeath, AnimationStateJump, AnimationStateCrouch,
		AnimationStateUse,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := states[i%len(states)]
		_ = sys.getCacheKey(int64(i), state)
	}
}

// BenchmarkStateToInt benchmarks the state-to-int conversion.
func BenchmarkStateToInt(b *testing.B) {
	state := AnimationStateWalk

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stateToInt(state)
	}
}

// BenchmarkCacheFrames_Uint64Key benchmarks frame caching with uint64 keys.
func BenchmarkCacheFrames_Uint64Key(b *testing.B) {
	sys := NewAnimationSystem(nil)
	frames := make([]*ebiten.Image, 8) // Typical 8-frame animation

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := sys.getCacheKey(int64(i), AnimationStateWalk)
		sys.cacheFrames(key, frames)
	}
}

// BenchmarkMapLookup_Uint64 benchmarks map lookup performance with uint64 keys.
func BenchmarkMapLookup_Uint64(b *testing.B) {
	sys := NewAnimationSystem(nil)
	frames := make([]*ebiten.Image, 8)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := sys.getCacheKey(int64(i), AnimationStateWalk)
		sys.frameCache[key] = frames
	}

	lookupKey := sys.getCacheKey(50, AnimationStateWalk)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sys.frameCache[lookupKey]
	}
}

// BenchmarkMapLookup_String provides baseline comparison with string keys.
func BenchmarkMapLookup_String(b *testing.B) {
	cache := make(map[string][]*ebiten.Image)
	frames := make([]*ebiten.Image, 8)

	// Pre-populate cache with string keys for comparison
	for i := 0; i < 100; i++ {
		key := "12345_walk" // Simulating old string key format
		cache[key] = frames
	}

	lookupKey := "12345_walk"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache[lookupKey]
	}
}
