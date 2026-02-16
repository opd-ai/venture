package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestAnimationSystem_DefaultRegenPerFrame verifies V1 increased default from 8 to 16.
func TestAnimationSystem_DefaultRegenPerFrame(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	if sys.maxRegenPerFrame != 16 {
		t.Errorf("Expected default maxRegenPerFrame 16 (V1 fix), got %d", sys.maxRegenPerFrame)
	}
}

// TestAnimationSystem_PreGenerateSprites tests batch pre-generation during loading.
func TestAnimationSystem_PreGenerateSprites(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	// Create entities with dirty animation components
	entities := make([]*Entity, 5)
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 0})
		entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		animComp := NewAnimationComponent(int64(i + 100))
		animComp.Dirty = true
		animComp.CurrentState = AnimationStateIdle
		entity.AddComponent(animComp)
		entities[i] = entity
	}

	// Pre-generate all at once (bypasses per-frame limit)
	generated := sys.PreGenerateSprites(entities)
	if generated != 5 {
		t.Errorf("Expected 5 sprites pre-generated, got %d", generated)
	}

	// Verify entities are no longer dirty
	for i, entity := range entities {
		animComp := entity.GetAnimation()
		if animComp == nil {
			t.Fatalf("Entity %d has no animation component", i)
		}
		if animComp.Dirty {
			t.Errorf("Entity %d should not be dirty after pre-generation", i)
		}
	}
}

// TestAnimationSystem_PreGenerateSprites_SkipsClean tests that clean entities are skipped.
func TestAnimationSystem_PreGenerateSprites_SkipsClean(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	// Create 3 entities: 2 dirty, 1 clean
	entities := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 0})
		entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		animComp := NewAnimationComponent(int64(i + 200))
		animComp.Dirty = i != 1 // Middle entity is clean
		animComp.CurrentState = AnimationStateWalk
		entity.AddComponent(animComp)
		entities[i] = entity
	}

	generated := sys.PreGenerateSprites(entities)
	if generated != 2 {
		t.Errorf("Expected 2 sprites pre-generated (1 skipped clean), got %d", generated)
	}
}

// TestAnimationSystem_PreGenerateSprites_SkipsMissingComponents tests robustness.
func TestAnimationSystem_PreGenerateSprites_SkipsMissingComponents(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	// Entity without animation component
	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	// Entity without sprite component
	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 100, Y: 0})
	animComp := NewAnimationComponent(300)
	animComp.Dirty = true
	e2.AddComponent(animComp)

	// Valid entity
	e3 := world.CreateEntity()
	e3.AddComponent(&PositionComponent{X: 200, Y: 0})
	e3.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	animComp3 := NewAnimationComponent(301)
	animComp3.Dirty = true
	animComp3.CurrentState = AnimationStateIdle
	e3.AddComponent(animComp3)

	generated := sys.PreGenerateSprites([]*Entity{e1, e2, e3})
	if generated != 1 {
		t.Errorf("Expected 1 sprite pre-generated (2 skipped), got %d", generated)
	}
}

// TestAnimationSystem_PreGenerateSprites_Empty tests empty input.
func TestAnimationSystem_PreGenerateSprites_Empty(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	generated := sys.PreGenerateSprites(nil)
	if generated != 0 {
		t.Errorf("Expected 0 sprites pre-generated for nil input, got %d", generated)
	}

	generated = sys.PreGenerateSprites([]*Entity{})
	if generated != 0 {
		t.Errorf("Expected 0 sprites pre-generated for empty input, got %d", generated)
	}
}

// TestAnimationSystem_SetPredictiveWarmer tests warmer integration.
func TestAnimationSystem_SetPredictiveWarmer(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Initially nil
	if sys.GetPredictiveWarmer() != nil {
		t.Error("Expected predictive warmer to be nil initially")
	}

	// Set a warmer
	spriteCache := cache.NewSpriteCache(cache.DefaultCacheSize)
	pregen := cache.NewPreGenerator(spriteCache)
	config := cache.DefaultWarmerConfig()
	warmer := cache.NewPredictiveCacheWarmer(spriteCache, pregen, config)

	sys.SetPredictiveWarmer(warmer)
	if sys.GetPredictiveWarmer() != warmer {
		t.Error("Expected predictive warmer to be set")
	}

	// Disconnect
	sys.SetPredictiveWarmer(nil)
	if sys.GetPredictiveWarmer() != nil {
		t.Error("Expected predictive warmer to be nil after disconnection")
	}
}

// TestAnimationSystem_RecordWarmerAccess tests that access patterns are recorded.
func TestAnimationSystem_RecordWarmerAccess(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	spriteCache := cache.NewSpriteCache(cache.DefaultCacheSize)
	pregen := cache.NewPreGenerator(spriteCache)
	config := cache.DefaultWarmerConfig()
	warmer := cache.NewPredictiveCacheWarmer(spriteCache, pregen, config)
	sys.SetPredictiveWarmer(warmer)

	// Simulate access pattern recording
	key1 := cache.GenerateKey(12345, "idle", 0)
	key2 := cache.GenerateKey(12345, "walk", 0)
	sys.recordWarmerAccess(key1, true)
	sys.recordWarmerAccess(key2, false)

	stats := warmer.Stats()
	if stats.AccessLogSize != 2 {
		t.Errorf("Expected 2 access records, got %d", stats.AccessLogSize)
	}
	if stats.PatternCount != 2 {
		t.Errorf("Expected 2 patterns, got %d", stats.PatternCount)
	}
}

// TestAnimationSystem_RecordWarmerAccess_NilWarmer tests no-op when warmer is nil.
func TestAnimationSystem_RecordWarmerAccess_NilWarmer(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	// Should not panic when warmer is nil
	key := cache.GenerateKey(12345, "idle", 0)
	sys.recordWarmerAccess(key, true)
	sys.recordWarmerAccess(key, false)
}

// TestAnimationSystem_WarmPredictedSprites_NilWarmer tests no-op when warmer is nil.
func TestAnimationSystem_WarmPredictedSprites_NilWarmer(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	count := sys.WarmPredictedSprites()
	if count != 0 {
		t.Errorf("Expected 0 when warmer is nil, got %d", count)
	}
}

// TestAnimationSystem_WarmPredictedSprites tests prediction-based warming.
func TestAnimationSystem_WarmPredictedSprites(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	spriteCache := cache.NewSpriteCache(cache.DefaultCacheSize)
	pregen := cache.NewPreGenerator(spriteCache)
	config := cache.DefaultWarmerConfig()
	config.HotThreshold = 1 // Low threshold for testing
	warmer := cache.NewPredictiveCacheWarmer(spriteCache, pregen, config)
	sys.SetPredictiveWarmer(warmer)
	sys.SetSpriteCache(spriteCache)

	// Record some access patterns to build predictions
	key1 := cache.GenerateKey(555, "idle", 0)
	key2 := cache.GenerateKey(555, "walk", 0)
	warmer.RecordAccess(key1, true, 1)
	warmer.RecordAccess(key2, false, 2)

	// Should return count of predicted sprites (may be 0 if all cached)
	count := sys.WarmPredictedSprites()
	// Result depends on whether predictions are generated; at minimum, no panic
	_ = count
}

// TestAnimationSystem_PreGenerateSprites_CachesResults tests that pre-generated
// sprites are cached and subsequent Update calls don't re-generate.
func TestAnimationSystem_PreGenerateSprites_CachesResults(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	animComp := NewAnimationComponent(999)
	animComp.Dirty = true
	animComp.CurrentState = AnimationStateIdle
	entity.AddComponent(animComp)

	// Pre-generate
	generated := sys.PreGenerateSprites([]*Entity{entity})
	if generated != 1 {
		t.Fatalf("Expected 1 sprite pre-generated, got %d", generated)
	}

	// Cache should have the entry now
	cacheSize := sys.GetCacheSize()
	if cacheSize == 0 {
		t.Error("Expected cache to have entries after pre-generation")
	}

	// Running Update should not trigger any regeneration (entity is clean)
	sys.SetMaxRegenPerFrame(1) // Very low limit
	err := sys.Update([]*Entity{entity}, 0.016)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	stats := sys.GetStats()
	if stats.CompletedRegen != 0 {
		t.Errorf("Expected 0 regenerations after pre-generation, got %d", stats.CompletedRegen)
	}
}

// TestAnimationSystem_WarmerTickIncrement tests that tick counter advances.
func TestAnimationSystem_WarmerTickIncrement(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	spriteCache := cache.NewSpriteCache(cache.DefaultCacheSize)
	pregen := cache.NewPreGenerator(spriteCache)
	warmer := cache.NewPredictiveCacheWarmer(spriteCache, pregen, cache.DefaultWarmerConfig())
	sys.SetPredictiveWarmer(warmer)

	if sys.warmerTickCount != 0 {
		t.Errorf("Expected initial tick count 0, got %d", sys.warmerTickCount)
	}

	key := cache.GenerateKey(1, "idle", 0)
	sys.recordWarmerAccess(key, true)
	if sys.warmerTickCount != 1 {
		t.Errorf("Expected tick count 1 after first access, got %d", sys.warmerTickCount)
	}

	sys.recordWarmerAccess(key, false)
	if sys.warmerTickCount != 2 {
		t.Errorf("Expected tick count 2 after second access, got %d", sys.warmerTickCount)
	}
}

// TestAnimationSystem_PreGenerate_PopulatesSpriteImage verifies entities get sprite images.
func TestAnimationSystem_PreGenerate_PopulatesSpriteImage(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)

	world := NewWorld()

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true, Image: ebiten.NewImage(32, 32)})
	animComp := NewAnimationComponent(42)
	animComp.Dirty = true
	animComp.CurrentState = AnimationStateWalk
	animComp.Playing = true
	animComp.Loop = true
	entity.AddComponent(animComp)

	sys.PreGenerateSprites([]*Entity{entity})

	// Animation should have frames populated
	anim := entity.GetAnimation()
	if anim == nil {
		t.Fatal("Animation component is nil")
	}
	if len(anim.Frames) == 0 {
		t.Error("Expected animation frames to be populated after pre-generation")
	}
}
