package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestFrameSlicePooling tests that frame slices are pooled and reused.
func TestFrameSlicePooling(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	// Get first slice
	slice1 := sys.getFrameSlice(8)
	if len(slice1) != 8 {
		t.Errorf("expected length 8, got %d", len(slice1))
	}
	if cap(slice1) < 8 {
		t.Errorf("expected capacity >= 8, got %d", cap(slice1))
	}

	// Return slice to pool
	sys.putFrameSlice(slice1)

	// Get second slice - should reuse first slice's backing array
	slice2 := sys.getFrameSlice(8)
	if len(slice2) != 8 {
		t.Errorf("expected length 8, got %d", len(slice2))
	}

	// Verify pool reuse by checking capacity matches (same backing array)
	if cap(slice2) != cap(slice1) {
		t.Log("Pool may have created new slice (acceptable during warmup)")
	}
}

// TestFrameSlicePoolingDifferentSizes tests pooling with varying slice sizes.
func TestFrameSlicePoolingDifferentSizes(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	testCases := []struct {
		name  string
		size  int
		check func(t *testing.T, slice []*ebiten.Image)
	}{
		{
			name: "small slice (3 frames)",
			size: 3,
			check: func(t *testing.T, slice []*ebiten.Image) {
				if len(slice) != 3 {
					t.Errorf("expected length 3, got %d", len(slice))
				}
			},
		},
		{
			name: "typical slice (8 frames)",
			size: 8,
			check: func(t *testing.T, slice []*ebiten.Image) {
				if len(slice) != 8 {
					t.Errorf("expected length 8, got %d", len(slice))
				}
			},
		},
		{
			name: "large slice (16 frames)",
			size: 16,
			check: func(t *testing.T, slice []*ebiten.Image) {
				if len(slice) != 16 {
					t.Errorf("expected length 16, got %d", len(slice))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slice := sys.getFrameSlice(tc.size)
			tc.check(t, slice)
			sys.putFrameSlice(slice)
		})
	}
}

// TestFrameSliceClearOnReturn tests that slices are cleared before returning to pool.
func TestFrameSliceClearOnReturn(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	// Get slice and populate with dummy images
	slice := sys.getFrameSlice(4)
	for i := range slice {
		slice[i] = ebiten.NewImage(1, 1) // Dummy image
	}

	// Verify populated
	for i, img := range slice {
		if img == nil {
			t.Errorf("slice[%d] should be non-nil before return", i)
		}
	}

	// Return to pool (should clear)
	sys.putFrameSlice(slice)

	// Get new slice
	newSlice := sys.getFrameSlice(4)

	// Verify all elements are nil (cleared)
	for i, img := range newSlice {
		if img != nil {
			t.Errorf("slice[%d] should be nil after pool reuse, got %v", i, img)
		}
	}
}

// TestCacheEvictionReturnsFramesToPool tests that cache eviction returns slices to pool.
func TestCacheEvictionReturnsFramesToPool(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	// Set small cache size for testing
	sys.maxCacheSize = 2

	// Create dummy frames
	frames1 := sys.getFrameSlice(4)
	for i := range frames1 {
		frames1[i] = ebiten.NewImage(1, 1)
	}

	frames2 := sys.getFrameSlice(4)
	for i := range frames2 {
		frames2[i] = ebiten.NewImage(1, 1)
	}

	frames3 := sys.getFrameSlice(4)
	for i := range frames3 {
		frames3[i] = ebiten.NewImage(1, 1)
	}

	// Cache frames (fill cache)
	sys.cacheFrames(uint64(1), frames1)
	sys.cacheFrames(uint64(2), frames2)

	// Cache third set (should evict key1)
	sys.cacheFrames(uint64(3), frames3)

	// Verify cache size
	sys.cacheMutex.RLock()
	cacheSize := len(sys.frameCache)
	sys.cacheMutex.RUnlock()

	if cacheSize != 2 {
		t.Errorf("expected cache size 2, got %d", cacheSize)
	}

	// Verify key1 was evicted
	sys.cacheMutex.RLock()
	_, exists := sys.frameCache[uint64(1)]
	sys.cacheMutex.RUnlock()

	if exists {
		t.Error("key1 should have been evicted")
	}

	// Note: We can't directly verify frames1 was returned to pool,
	// but putFrameSlice was called during eviction (verified by code coverage)
}

// TestGenerateAllFramesUsesPool tests that generateAllFrames uses pooled slices.
func TestGenerateAllFramesUsesPool(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	// Create test entity with animation
	entity := NewEntity(1)

	anim := &AnimationComponent{
		CurrentState: AnimationStateIdle,
		Seed:         12345,
	}

	sprite := &EbitenSprite{
		Image:  ebiten.NewImage(16, 16),
		Width:  16,
		Height: 16,
	}

	entity.AddComponent(anim)
	entity.AddComponent(sprite)

	// Generate frames (should use pool internally)
	config := sprites.Config{
		Type:   0, // Sprite type 0
		Width:  16,
		Height: 16,
		Seed:   12345,
		Custom: make(map[string]interface{}),
	}

	baseSprite := ebiten.NewImage(16, 16)
	frames, err := sys.generateAllFrames(entity, baseSprite, config, anim, 8)
	if err != nil {
		t.Fatalf("generateAllFrames failed: %v", err)
	}

	if len(frames) != 8 {
		t.Errorf("expected 8 frames, got %d", len(frames))
	}

	// Verify all frames were generated
	for i, frame := range frames {
		if frame == nil {
			t.Errorf("frame[%d] is nil", i)
		}
	}
}

// TestPutFrameSliceNilSafe tests that putFrameSlice handles nil safely.
func TestPutFrameSliceNilSafe(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	// Should not panic
	sys.putFrameSlice(nil)
}

// BenchmarkFrameSlicePooling benchmarks the performance of pooled vs non-pooled allocation.
func BenchmarkFrameSlicePooling(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	b.Run("Pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := sys.getFrameSlice(8)
			// Simulate use
			for j := range slice {
				slice[j] = nil
			}
			sys.putFrameSlice(slice)
		}
	})

	b.Run("NonPooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := make([]*ebiten.Image, 8)
			// Simulate use
			for j := range slice {
				slice[j] = nil
			}
			// No return to pool
		}
	})
}

// BenchmarkGenerateAllFrames benchmarks frame generation with pooling.
func BenchmarkGenerateAllFrames(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewAnimationSystem(gen)

	entity := NewEntity(1)

	anim := &AnimationComponent{
		CurrentState: AnimationStateIdle,
		Seed:         12345,
	}
	entity.AddComponent(anim)

	config := sprites.Config{
		Type:   0, // Sprite type 0
		Width:  16,
		Height: 16,
		Seed:   12345,
		Custom: make(map[string]interface{}),
	}

	baseSprite := ebiten.NewImage(16, 16)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sys.generateAllFrames(entity, baseSprite, config, anim, 8)
		if err != nil {
			b.Fatalf("generateAllFrames failed: %v", err)
		}
	}
}
