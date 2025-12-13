package engine

import (
	"runtime"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/parallel"
	"github.com/opd-ai/venture/pkg/rendering/pool"
)

func TestImagePoolAdapter(t *testing.T) {
	// Create pool and adapter
	p := pool.NewImagePool()
	adapter := NewImagePoolAdapter(p)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Player size", pool.SizePlayer, pool.SizePlayer},
		{"Small size", pool.SizeSmall, pool.SizeSmall},
		{"Medium size", pool.SizeMedium, pool.SizeMedium},
		{"Large size", pool.SizeLarge, pool.SizeLarge},
		{"Custom size", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get image from pool
			img := adapter.GetImage(tt.width, tt.height)
			if img == nil {
				t.Errorf("GetImage(%d, %d) returned nil", tt.width, tt.height)
				return
			}

			// Check size
			bounds := img.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("GetImage(%d, %d) returned image with size (%d, %d)",
					tt.width, tt.height, bounds.Dx(), bounds.Dy())
			}

			// Return to pool
			adapter.PutImage(img)
		})
	}
}

func TestImagePoolAdapter_NilImage(t *testing.T) {
	p := pool.NewImagePool()
	adapter := NewImagePoolAdapter(p)

	// Putting nil should not panic
	adapter.PutImage(nil)
}

func TestParallelRendererAdapter(t *testing.T) {
	// Create worker pool and adapter
	workerCount := runtime.NumCPU()
	if workerCount > 4 {
		workerCount = 4 // Limit for test
	}
	wp := parallel.NewWorkerPool(workerCount)
	adapter := NewParallelRendererAdapter(wp)

	// Initially not running
	if adapter.IsRunning() {
		t.Error("ParallelRendererAdapter should not be running initially")
	}

	// Start workers
	adapter.Start()
	if !adapter.IsRunning() {
		t.Error("ParallelRendererAdapter should be running after Start()")
	}

	// Starting again should be idempotent
	adapter.Start()
	if !adapter.IsRunning() {
		t.Error("ParallelRendererAdapter should still be running after second Start()")
	}

	// Stop workers
	adapter.Stop()
	if adapter.IsRunning() {
		t.Error("ParallelRendererAdapter should not be running after Stop()")
	}

	// Stopping again should be idempotent
	adapter.Stop()
	if adapter.IsRunning() {
		t.Error("ParallelRendererAdapter should still not be running after second Stop()")
	}
}

func TestRenderSystem_SetPool(t *testing.T) {
	// Create render system
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Create pool adapter
	p := pool.NewImagePool()
	adapter := NewImagePoolAdapter(p)

	// Set pool should not panic
	renderSystem.SetPool(adapter)

	// Verify pool is set (we can't directly check private field, but we can verify no panic)
}

func TestRenderSystem_SetParallelRenderer(t *testing.T) {
	// Create render system
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Create parallel renderer adapter
	wp := parallel.NewWorkerPool(2)
	adapter := NewParallelRendererAdapter(wp)

	// Set parallel renderer should not panic
	renderSystem.SetParallelRenderer(adapter)

	// Verify renderer is set (we can't directly check private field, but we can verify no panic)
}

func BenchmarkImagePoolAdapter_GetPut(b *testing.B) {
	p := pool.NewImagePool()
	adapter := NewImagePoolAdapter(p)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := adapter.GetImage(pool.SizeSmall, pool.SizeSmall)
		adapter.PutImage(img)
	}
}

func BenchmarkImagePoolAdapter_StandardSizes(b *testing.B) {
	p := pool.NewImagePool()
	adapter := NewImagePoolAdapter(p)

	sizes := []int{pool.SizePlayer, pool.SizeSmall, pool.SizeMedium, pool.SizeLarge}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		img := adapter.GetImage(size, size)
		adapter.PutImage(img)
	}
}
