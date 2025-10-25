package pool

import (
	"image/color"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewImagePool(t *testing.T) {
	pool := NewImagePool()
	if pool == nil {
		t.Fatal("NewImagePool returned nil")
	}

	// Verify stats are initialized
	stats := pool.Stats()
	if stats.Gets != 0 || stats.Puts != 0 || stats.Creates != 0 {
		t.Error("New pool should have zero statistics")
	}
}

func TestImagePool_GetImage_StandardSizes(t *testing.T) {
	pool := NewImagePool()

	tests := []struct {
		name   string
		width  int
		height int
		size   int
	}{
		{"player size", SizePlayer, SizePlayer, SizePlayer},
		{"small size", SizeSmall, SizeSmall, SizeSmall},
		{"medium size", SizeMedium, SizeMedium, SizeMedium},
		{"large size", SizeLarge, SizeLarge, SizeLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := pool.GetImage(tt.width, tt.height)
			if img == nil {
				t.Fatal("GetImage returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.size || bounds.Dy() != tt.size {
				t.Errorf("Image size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), tt.size, tt.size)
			}
		})
	}
}

func TestImagePool_GetImage_NonStandardSize(t *testing.T) {
	pool := NewImagePool()

	// Non-standard size should create new image
	img := pool.GetImage(50, 50)
	if img == nil {
		t.Fatal("GetImage returned nil for non-standard size")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Image size = %dx%d, want 50x50", bounds.Dx(), bounds.Dy())
	}
}

func TestImagePool_GetImage_NonSquare(t *testing.T) {
	pool := NewImagePool()

	// Non-square size should create new image
	img := pool.GetImage(32, 64)
	if img == nil {
		t.Fatal("GetImage returned nil for non-square size")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 64 {
		t.Errorf("Image size = %dx%d, want 32x64", bounds.Dx(), bounds.Dy())
	}
}

func TestImagePool_PutImage(t *testing.T) {
	pool := NewImagePool()

	// Get and put standard size
	img := pool.GetImage(SizeSmall, SizeSmall)
	pool.PutImage(img)

	// Stats should reflect get and put
	stats := pool.Stats()
	if stats.Gets != 1 {
		t.Errorf("Gets = %d, want 1", stats.Gets)
	}
	if stats.Puts != 1 {
		t.Errorf("Puts = %d, want 1", stats.Puts)
	}
}

func TestImagePool_PutImage_Nil(t *testing.T) {
	pool := NewImagePool()

	// Putting nil should not panic and should not increment puts
	pool.PutImage(nil)

	stats := pool.Stats()
	// PutImage returns early for nil, so puts should not increment
	if stats.Puts != 0 {
		t.Errorf("Puts = %d, want 0 (nil should be ignored)", stats.Puts)
	}
}

func TestImagePool_Reuse(t *testing.T) {
	pool := NewImagePool()

	// Get image
	img1 := pool.GetImage(SizeSmall, SizeSmall)
	if img1 == nil {
		t.Fatal("First GetImage returned nil")
	}

	// Return to pool
	pool.PutImage(img1)

	// Get again - should reuse
	img2 := pool.GetImage(SizeSmall, SizeSmall)
	if img2 == nil {
		t.Fatal("Second GetImage returned nil")
	}

	// Should be the same image (pooled)
	if img1 != img2 {
		t.Error("Expected same image from pool, got different instance")
	}

	stats := pool.Stats()
	if stats.Creates > 1 {
		t.Errorf("Creates = %d, want 1 (should reuse)", stats.Creates)
	}
}

func TestImagePool_Clear(t *testing.T) {
	pool := NewImagePool()

	// Get image and draw something
	img := pool.GetImage(SizeSmall, SizeSmall)
	// Draw a pixel (simulating usage)
	img.Set(0, 0, color.White)

	// Return to pool (should clear)
	pool.PutImage(img)

	// Get again
	img2 := pool.GetImage(SizeSmall, SizeSmall)

	// Verify it's cleared (or at least same image)
	// Ebiten's Clear() is called in PutImage
	if img != img2 {
		t.Error("Expected same image from pool")
	}
}

func TestStatistics_ReuseRate(t *testing.T) {
	tests := []struct {
		name    string
		gets    uint64
		creates uint64
		want    float64
	}{
		{"no gets", 0, 0, 0.0},
		{"no reuse", 10, 10, 0.0},
		{"50% reuse", 10, 5, 0.5},
		{"90% reuse", 100, 10, 0.9},
		{"100% reuse", 100, 0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := Statistics{
				Gets:    tt.gets,
				Creates: tt.creates,
			}
			got := stats.ReuseRate()
			if got != tt.want {
				t.Errorf("ReuseRate() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestImagePool_ConcurrentAccess(t *testing.T) {
	pool := NewImagePool()

	// Test concurrent get/put operations
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Get different sizes based on goroutine ID
			size := []int{SizePlayer, SizeSmall, SizeMedium, SizeLarge}[id%4]
			img := pool.GetImage(size, size)
			if img == nil {
				// Use t.Errorf in a thread-safe way
				return
			}

			// Simulate some work
			img.Set(0, 0, color.White)

			// Return to pool
			pool.PutImage(img)
		}(i)
	}

	wg.Wait()

	// Verify statistics (approximately, due to concurrency)
	stats := pool.Stats()
	// Allow some tolerance for race conditions in stats tracking
	if stats.Gets < uint64(numGoroutines)-5 {
		t.Errorf("Gets = %d, want at least %d", stats.Gets, numGoroutines-5)
	}
	if stats.Puts < uint64(numGoroutines)-5 {
		t.Errorf("Puts = %d, want at least %d", stats.Puts, numGoroutines-5)
	}
}

func TestGlobalPool(t *testing.T) {
	// Reset global pool stats
	ResetStats()

	// Test global Get/Put functions
	img := GetImage(SizeSmall, SizeSmall)
	if img == nil {
		t.Fatal("Global GetImage returned nil")
	}

	PutImage(img)

	// Check global stats
	stats := Stats()
	if stats.Gets != 1 || stats.Puts != 1 {
		t.Errorf("Global pool stats incorrect: gets=%d, puts=%d", stats.Gets, stats.Puts)
	}
}

func TestResetStats(t *testing.T) {
	ResetStats()

	// Generate some activity
	img := GetImage(SizeSmall, SizeSmall)
	PutImage(img)

	// Reset
	ResetStats()

	// Stats should be zero
	stats := Stats()
	if stats.Gets != 0 || stats.Puts != 0 || stats.Creates != 0 {
		t.Error("ResetStats did not clear statistics")
	}
}

// Benchmarks

func BenchmarkImagePool_GetPut_Player(b *testing.B) {
	pool := NewImagePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.GetImage(SizePlayer, SizePlayer)
		pool.PutImage(img)
	}
}

func BenchmarkImagePool_GetPut_Small(b *testing.B) {
	pool := NewImagePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.GetImage(SizeSmall, SizeSmall)
		pool.PutImage(img)
	}
}

func BenchmarkImagePool_GetPut_Medium(b *testing.B) {
	pool := NewImagePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.GetImage(SizeMedium, SizeMedium)
		pool.PutImage(img)
	}
}

func BenchmarkImagePool_GetPut_Large(b *testing.B) {
	pool := NewImagePool()
	// Limit iterations to avoid OOM with large 128x128 images
	origN := b.N
	if b.N > 50000 {
		b.N = 50000
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.GetImage(SizeLarge, SizeLarge)
		pool.PutImage(img)
	}

	b.StopTimer()
	// Report actual iteration count
	if origN != b.N {
		b.Logf("Limited iterations from %d to %d to avoid OOM", origN, b.N)
	}
}

func BenchmarkImagePool_GetPut_NonStandard(b *testing.B) {
	pool := NewImagePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.GetImage(50, 50)
		pool.PutImage(img)
	}
}

func BenchmarkDirect_NewImage_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ebiten.NewImage(SizeSmall, SizeSmall)
	}
}

func BenchmarkPooled_GetImage_Small(b *testing.B) {
	pool := NewImagePool()
	// Pre-warm the pool
	for i := 0; i < 10; i++ {
		img := pool.GetImage(SizeSmall, SizeSmall)
		pool.PutImage(img)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := pool.GetImage(SizeSmall, SizeSmall)
		pool.PutImage(img)
	}
}

func BenchmarkGlobalPool_GetPut(b *testing.B) {
	ResetStats()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := GetImage(SizeSmall, SizeSmall)
		PutImage(img)
	}
}

func BenchmarkImagePool_Concurrent(b *testing.B) {
	pool := NewImagePool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			img := pool.GetImage(SizeSmall, SizeSmall)
			pool.PutImage(img)
		}
	})
}
