package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/animation"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewAnimationAdapter(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	if adapter == nil {
		t.Fatal("NewAnimationAdapter returned nil")
	}
	if !adapter.IsEnabled() {
		t.Error("expected animation adapter to be enabled by default")
	}
}

func TestAnimationAdapter_SetEnabled(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	adapter.SetEnabled(false)
	if adapter.IsEnabled() {
		t.Error("expected animation adapter to be disabled")
	}

	adapter.SetEnabled(true)
	if !adapter.IsEnabled() {
		t.Error("expected animation adapter to be enabled")
	}
}

func TestAnimationAdapter_Update(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	// Update should not panic (it's a no-op for animation adapter)
	entities := []*Entity{}
	adapter.Update(entities, 0.016)
}

func TestAnimationAdapter_GenerateFrame(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	config := sprites.Config{
		Type:    sprites.SpriteEntity,
		Seed:    12345,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Custom:  make(map[string]interface{}),
	}

	frame, err := adapter.GenerateFrame(12345, "walk", 0, 8, animation.Dir8South, config)
	if err != nil {
		t.Fatalf("GenerateFrame failed: %v", err)
	}
	if frame == nil {
		t.Error("expected frame to be generated")
	}
}

func TestAnimationAdapter_GenerateFrameDisabled(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)
	adapter.SetEnabled(false)

	config := sprites.Config{
		Type:    sprites.SpriteEntity,
		Seed:    12345,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Custom:  make(map[string]interface{}),
	}

	frame, err := adapter.GenerateFrame(12345, "walk", 0, 8, animation.Dir8South, config)
	if err != nil {
		t.Fatalf("GenerateFrame failed when disabled: %v", err)
	}
	if frame != nil {
		t.Error("expected nil frame when disabled")
	}
}

func TestAnimationAdapter_GenerateSequence(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	config := sprites.Config{
		Type:    sprites.SpriteEntity,
		Seed:    12345,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Custom:  make(map[string]interface{}),
	}

	frames, err := adapter.GenerateSequence(12345, "walk", animation.Dir8South, config)
	if err != nil {
		t.Fatalf("GenerateSequence failed: %v", err)
	}
	if len(frames) == 0 {
		t.Error("expected animation frames to be generated")
	}
}

func TestAnimationAdapter_GetCacheStats(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	hitRate, cacheSize := adapter.GetCacheStats()

	// Initially cache should be empty
	if hitRate != 0.0 {
		t.Errorf("expected initial hit rate of 0, got %f", hitRate)
	}
	if cacheSize != 0 {
		t.Errorf("expected initial cache size of 0, got %d", cacheSize)
	}
}

func TestAnimationAdapter_ClearCache(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	// Generate some frames to populate cache
	config := sprites.Config{
		Type:    sprites.SpriteEntity,
		Seed:    12345,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Custom:  make(map[string]interface{}),
	}
	adapter.GenerateFrame(12345, "walk", 0, 8, animation.Dir8South, config)

	// Clear cache
	adapter.ClearCache()

	_, cacheSize := adapter.GetCacheStats()
	if cacheSize != 0 {
		t.Errorf("expected cache size of 0 after clear, got %d", cacheSize)
	}
}

func TestAnimationAdapter_SetArticulationConfig(t *testing.T) {
	generator := sprites.NewGenerator()
	adapter := NewAnimationAdapter(generator, nil)

	config := animation.DefaultArticulationConfig()

	// Should not panic
	adapter.SetArticulationConfig(config)
}
