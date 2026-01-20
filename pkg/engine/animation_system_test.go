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
		{AnimationStateIdle, 8}, // Phase 15.2: Increased from 4 to 8
		{AnimationStateWalk, 8},
		{AnimationStateRun, 8},
		{AnimationStateAttack, 8}, // Phase 15.2: Increased from 6 to 8
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

// BenchmarkAnimationSystem_ViewportCulling benchmarks viewport culling performance.
func BenchmarkAnimationSystem_ViewportCulling(b *testing.B) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	// Create camera
	cameraSystem := NewCameraSystem(800, 600)
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 400, Y: 300})
	cameraComp := NewCameraComponent()
	cameraComp.X = 400
	cameraComp.Y = 300
	player.AddComponent(cameraComp)
	cameraSystem.SetActiveCamera(player)

	sys.SetCameraSystem(cameraSystem)
	sys.EnableViewportCulling(true)

	// Create 100 entities (mix of visible and culled)
	entities := []*Entity{player}
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		// Half inside viewport, half outside
		x := float64(400 + (i-50)*50)
		y := 300.0
		entity.AddComponent(&PositionComponent{X: x, Y: y})
		entity.AddComponent(&EbitenSprite{Width: 28, Height: 28, Visible: true, Image: ebiten.NewImage(28, 28)})

		animComp := NewAnimationComponent(int64(i))
		animComp.Frames = []*ebiten.Image{ebiten.NewImage(28, 28)}
		entity.AddComponent(animComp)

		entities = append(entities, entity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

// BenchmarkAnimationSystem_DistanceLOD benchmarks distance-based LOD performance.
func BenchmarkAnimationSystem_DistanceLOD(b *testing.B) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 400, Y: 300})

	sys.SetPlayerEntity(player)
	sys.EnableDistanceLOD(true)

	// Create 100 entities at various distances
	entities := []*Entity{player}
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		distance := float64(i * 10)
		entity.AddComponent(&PositionComponent{X: 400 + distance, Y: 300})
		entity.AddComponent(&EbitenSprite{Width: 28, Height: 28, Visible: true, Image: ebiten.NewImage(28, 28)})

		animComp := NewAnimationComponent(int64(i))
		animComp.Frames = []*ebiten.Image{ebiten.NewImage(28, 28)}
		animComp.Playing = true
		entity.AddComponent(animComp)

		entities = append(entities, entity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

// TestAnimationSystem_SetMaxCacheSize tests cache size configuration and eviction.
func TestAnimationSystem_SetMaxCacheSize(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Verify default size
	if sys.maxCacheSize != 100 {
		t.Errorf("Expected default cache size 100, got %d", sys.maxCacheSize)
	}

	// Test increasing cache size
	sys.SetMaxCacheSize(200)
	if sys.maxCacheSize != 200 {
		t.Errorf("Expected cache size 200, got %d", sys.maxCacheSize)
	}

	// Add test entries to cache
	for i := 0; i < 150; i++ {
		key := sys.getCacheKey(int64(i), AnimationStateWalk)
		frames := []*ebiten.Image{ebiten.NewImage(28, 28)}
		sys.cacheFrames(key, frames)
	}

	// Verify cache has 150 entries
	sys.cacheMutex.RLock()
	cacheSize := len(sys.frameCache)
	sys.cacheMutex.RUnlock()

	if cacheSize != 150 {
		t.Errorf("Expected cache to have 150 entries, got %d", cacheSize)
	}

	// Test decreasing cache size triggers eviction
	sys.SetMaxCacheSize(50)

	sys.cacheMutex.RLock()
	cacheSize = len(sys.frameCache)
	keyCount := len(sys.cacheKeys)
	sys.cacheMutex.RUnlock()

	if sys.maxCacheSize != 50 {
		t.Errorf("Expected cache size 50, got %d", sys.maxCacheSize)
	}

	if cacheSize != 50 {
		t.Errorf("Expected cache to be evicted to 50 entries, got %d", cacheSize)
	}

	if keyCount != 50 {
		t.Errorf("Expected 50 cache keys after eviction, got %d", keyCount)
	}

	// Verify cache still works after eviction
	sys.SetMaxCacheSize(100)
	key := sys.getCacheKey(99999, AnimationStateRun)
	frames := []*ebiten.Image{ebiten.NewImage(28, 28)}
	sys.cacheFrames(key, frames)

	sys.cacheMutex.RLock()
	_, exists := sys.frameCache[key]
	sys.cacheMutex.RUnlock()

	if !exists {
		t.Error("Cache should work after eviction")
	}
}

// TestAnimationSystem_SetMaxRegenPerFrame tests per-frame regeneration limit configuration.
func TestAnimationSystem_SetMaxRegenPerFrame(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Verify default limit (8)
	if sys.maxRegenPerFrame != 8 {
		t.Errorf("Expected default maxRegenPerFrame 8, got %d", sys.maxRegenPerFrame)
	}

	// Test GetMaxRegenPerFrame
	if sys.GetMaxRegenPerFrame() != 8 {
		t.Errorf("Expected GetMaxRegenPerFrame() to return 8, got %d", sys.GetMaxRegenPerFrame())
	}

	// Test setting custom limit
	sys.SetMaxRegenPerFrame(16)
	if sys.maxRegenPerFrame != 16 {
		t.Errorf("Expected maxRegenPerFrame 16, got %d", sys.maxRegenPerFrame)
	}

	// Test disabling limit (0 = unlimited)
	sys.SetMaxRegenPerFrame(0)
	if sys.maxRegenPerFrame != 0 {
		t.Errorf("Expected maxRegenPerFrame 0, got %d", sys.maxRegenPerFrame)
	}
}

// TestAnimationSystem_RegenLimitEnforced tests that regeneration limit is enforced.
func TestAnimationSystem_RegenLimitEnforced(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Set a low limit for testing
	sys.SetMaxRegenPerFrame(2)

	// Create a world with test entities
	world := NewWorld()

	// Create 5 entities, all with dirty animation components
	entities := make([]*Entity, 5)
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 0})
		entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		animComp := NewAnimationComponent(int64(i))
		animComp.Dirty = true
		animComp.CurrentState = AnimationStateIdle
		entity.AddComponent(animComp)
		entities[i] = entity
	}

	// Run one Update cycle
	err := sys.Update(entities, 0.016)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Check stats - should have completed 2 and deferred 3
	stats := sys.GetStats()
	if stats.CompletedRegen != 2 {
		t.Errorf("Expected 2 completed regenerations, got %d", stats.CompletedRegen)
	}
	if stats.DeferredRegen != 3 {
		t.Errorf("Expected 3 deferred regenerations, got %d", stats.DeferredRegen)
	}

	// Run another Update cycle
	err = sys.Update(entities, 0.016)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Now 2 more should be completed and 1 deferred
	stats = sys.GetStats()
	if stats.CompletedRegen != 2 {
		t.Errorf("Expected 2 completed regenerations in second frame, got %d", stats.CompletedRegen)
	}
	if stats.DeferredRegen != 1 {
		t.Errorf("Expected 1 deferred regeneration in second frame, got %d", stats.DeferredRegen)
	}

	// Third Update should complete the last one
	err = sys.Update(entities, 0.016)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	stats = sys.GetStats()
	if stats.CompletedRegen != 1 {
		t.Errorf("Expected 1 completed regeneration in third frame, got %d", stats.CompletedRegen)
	}
	if stats.DeferredRegen != 0 {
		t.Errorf("Expected 0 deferred regenerations in third frame, got %d", stats.DeferredRegen)
	}
}

// TestAnimationSystem_PlayerBypassesRegenLimit tests that player entity always regenerates.
func TestAnimationSystem_PlayerBypassesRegenLimit(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Set limit to 1 (only one NPC regeneration allowed; player bypasses limit)
	sys.SetMaxRegenPerFrame(1)

	world := NewWorld()

	// Create entities: 1 NPC, 1 player (with input component)
	npcEntity := world.CreateEntity()
	npcEntity.AddComponent(&PositionComponent{X: 0, Y: 0})
	npcEntity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	npcAnim := NewAnimationComponent(1)
	npcAnim.Dirty = true
	npcAnim.CurrentState = AnimationStateIdle
	npcEntity.AddComponent(npcAnim)

	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&PositionComponent{X: 100, Y: 0})
	playerEntity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	playerEntity.AddComponent(NewStubInput()) // This makes it a player
	playerAnim := NewAnimationComponent(2)
	playerAnim.Dirty = true
	playerAnim.CurrentState = AnimationStateIdle
	playerEntity.AddComponent(playerAnim)

	entities := []*Entity{npcEntity, playerEntity}

	// Run Update - both should regenerate (NPC uses budget, player bypasses)
	err := sys.Update(entities, 0.016)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Player should always regenerate regardless of limit
	if playerAnim.Dirty {
		t.Error("Player animation should have regenerated (Dirty should be false)")
	}

	// NPC should have regenerated (used the 1 slot budget)
	if npcAnim.Dirty {
		t.Error("NPC animation should have regenerated (Dirty should be false)")
	}
}
