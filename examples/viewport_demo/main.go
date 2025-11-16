// Phase 44 Viewport Optimization Example
//
// This example demonstrates the viewport optimization features:
// - Enhanced viewport culling for 1920x1080 resolution
// - Memory monitoring for sprite cache
// - Batch pre-generation for cache warming
//
// Run: go run examples/viewport_demo/main.go
package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/display"
)

func main() {
	fmt.Println("=== Phase 44: Viewport Optimization Demo ===")
	fmt.Println()

	// 1. Display Configuration (Phase 43)
	fmt.Println("1. Display Configuration:")
	cfg := display.NewConfigDefault() // 1920x1080
	fmt.Printf("   Resolution: %dx%d (%s)\n", cfg.Width, cfg.Height, cfg.GetResolution().Name)
	fmt.Printf("   Aspect Ratio: %.2f\n", cfg.AspectRatio())
	fmt.Println()

	// 2. Viewport Optimizer (Phase 44)
	fmt.Println("2. Viewport Optimizer:")
	optimizer := engine.NewViewportOptimizer()
	optimizer.SetTileSize(32.0)
	optimizer.SetMarginTiles(1) // 1-tile margin

	camera := engine.NewCameraComponent()
	camera.X = 960 // Center of 1920px width
	camera.Y = 540 // Center of 1080px height
	camera.Zoom = 1.0

	// Calculate viewport bounds
	bounds := optimizer.CalculateViewportBounds(
		camera.X, camera.Y,
		float64(cfg.Width), float64(cfg.Height),
		camera.Zoom,
	)
	fmt.Printf("   Viewport Bounds: X=%.1f, Y=%.1f, W=%.1f, H=%.1f\n",
		bounds.X, bounds.Y, bounds.Width, bounds.Height)
	fmt.Printf("   Margin: %d tiles (%.0fpx)\n", 1, 32.0)
	fmt.Println()

	// 3. Frustum Culling
	fmt.Println("3. Frustum Culling:")
	testEntities := []struct {
		name string
		x, y float64
		w, h float64
	}{
		{"Center Entity", 960, 540, 32, 32},
		{"Edge Entity", 1900, 540, 32, 32},
		{"Off-Screen Entity", 3000, 3000, 32, 32},
	}

	for _, entity := range testEntities {
		visible := optimizer.FrustumCull(entity.x, entity.y, entity.w, entity.h, bounds)
		status := "CULLED"
		if visible {
			status = "VISIBLE"
		}
		fmt.Printf("   %s: %s\n", entity.name, status)
	}
	fmt.Println()

	// 4. Sprite Cache (Phase 44)
	fmt.Println("4. Sprite Cache:")
	spriteCache := cache.NewSpriteCache(300 * 1024 * 1024) // 300MB limit
	fmt.Printf("   Max Size: %d MB\n", spriteCache.MaxSize()/(1024*1024))

	// Estimate capacity for 64x64 sprites
	const sprite64Size = 64 * 64 * 4 // 16KB
	capacity := int(spriteCache.MaxSize() / sprite64Size)
	fmt.Printf("   Estimated Capacity (64x64 sprites): %d sprites\n", capacity)
	fmt.Println()

	// 5. Memory Monitor (Phase 44)
	fmt.Println("5. Memory Monitor:")
	monitor := cache.NewMemoryMonitor(spriteCache)
	fmt.Printf("   Soft Limit: %d MB\n", monitor.Stats().CurrentUsage/(1024*1024))
	fmt.Printf("   Hard Limit: %d MB\n", 300)
	fmt.Printf("   Current Usage: %.1f%%\n", monitor.UsagePercentage())
	fmt.Printf("   Health: %v\n", monitor.IsHealthy())
	fmt.Println()

	// 6. Pre-Generation (Phase 44)
	fmt.Println("6. Batch Pre-Generation:")
	pregen := cache.NewPreGenerator(spriteCache)

	// Queue some sprite generation requests
	for i := 0; i < 10; i++ {
		key := cache.GenerateKey(int64(i), "idle", 0)
		// In real usage, would have actual generator function
		pregen.Queue(key, nil) // nil generator for demo
	}

	fmt.Printf("   Queued Requests: %d\n", pregen.QueueSize())
	stats := pregen.Stats()
	fmt.Printf("   Total Queued: %d\n", stats.RequestsQueued)
	fmt.Println()

	// 7. Performance Metrics
	fmt.Println("7. Target Metrics (Phase 44):")
	fmt.Println("   ✓ Off-Screen Rendered: <5%")
	fmt.Println("   ✓ Quadtree Query Time: <0.1ms")
	fmt.Println("   ✓ Cache Hit Rate: ≥90%")
	fmt.Println("   ✓ Memory Usage: <300MB")
	fmt.Println()

	// 8. Comparison with Old System
	fmt.Println("8. Improvement over 800x600:")
	oldScreenArea := 800.0 * 600.0
	newScreenArea := 1920.0 * 1080.0
	increase := newScreenArea / oldScreenArea
	fmt.Printf("   Screen Area Increase: %.2fx\n", increase)
	fmt.Printf("   Viewport Culling: %d tile margin for smooth rendering\n", 1)
	fmt.Printf("   Cache Size: 300MB (supports ~19,200 64x64 sprites)\n")
	fmt.Println()

	log.Println("Phase 44 Viewport Optimization Demo Complete!")
}
