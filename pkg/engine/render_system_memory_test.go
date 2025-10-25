package engine

import (
	"runtime"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkRenderSystem_Memory_2000Entities measures memory allocations with 2000 entities
func BenchmarkRenderSystem_Memory_2000Entities(b *testing.B) {
	// Create system
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)

	// Create 2000 entities
	entities := make([]*Entity, 2000)
	sharedSprites := make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		sharedSprites[i] = ebiten.NewImage(32, 32)
	}

	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 100 * 50),
			Y: float64(i / 100 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   sharedSprites[i%10],
			Width:   32,
			Height:  32,
			Visible: true,
		}
		entity.AddComponent(sprite)

		entities[i] = entity
	}

	// Build spatial partition
	spatialPartition.Update(entities, 0)

	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		renderSystem.Draw(screen, entities)
	}
}

// BenchmarkRenderSystem_Memory_5000Entities stress tests with 5000 entities
func BenchmarkRenderSystem_Memory_5000Entities(b *testing.B) {
	// Create system
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)

	// Create 5000 entities
	entities := make([]*Entity, 5000)
	sharedSprites := make([]*ebiten.Image, 20)
	for i := 0; i < 20; i++ {
		sharedSprites[i] = ebiten.NewImage(32, 32)
	}

	for i := 0; i < 5000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 100 * 50),
			Y: float64(i / 100 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   sharedSprites[i%20],
			Width:   32,
			Height:  32,
			Visible: true,
		}
		entity.AddComponent(sprite)

		entities[i] = entity
	}

	// Build spatial partition
	spatialPartition.Update(entities, 0)

	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		renderSystem.Draw(screen, entities)
	}
}

// BenchmarkRenderSystem_Memory_NoCulling benchmarks memory without culling
func BenchmarkRenderSystem_Memory_NoCulling(b *testing.B) {
	// Create system without culling
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	renderSystem.EnableCulling(false)

	// Create 2000 entities
	entities := make([]*Entity, 2000)
	sharedSprites := make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		sharedSprites[i] = ebiten.NewImage(32, 32)
	}

	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 100 * 50),
			Y: float64(i / 100 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   sharedSprites[i%10],
			Width:   32,
			Height:  32,
			Visible: true,
		}
		entity.AddComponent(sprite)

		entities[i] = entity
	}

	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		renderSystem.Draw(screen, entities)
	}
}

// TestRenderSystem_MemoryProfile tests memory usage over extended rendering
func TestRenderSystem_MemoryProfile(t *testing.T) {
	// Create system
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)

	// Create 1000 entities
	entities := make([]*Entity, 1000)
	sharedSprites := make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		sharedSprites[i] = ebiten.NewImage(32, 32)
	}

	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 50 * 50),
			Y: float64(i / 50 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   sharedSprites[i%10],
			Width:   32,
			Height:  32,
			Visible: true,
		}
		entity.AddComponent(sprite)

		entities[i] = entity
	}

	// Build spatial partition
	spatialPartition.Update(entities, 0)

	screen := ebiten.NewImage(800, 600)

	// Measure memory before
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	heapBefore := m1.HeapAlloc

	// Render 1000 frames
	for i := 0; i < 1000; i++ {
		renderSystem.Draw(screen, entities)
	}

	// Measure memory after
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	heapAfter := m2.HeapAlloc

	growth := int64(heapAfter) - int64(heapBefore)
	growthKB := float64(growth) / 1024

	t.Logf("Heap memory before: %.2f MB", float64(heapBefore)/(1024*1024))
	t.Logf("Heap memory after:  %.2f MB", float64(heapAfter)/(1024*1024))
	t.Logf("Memory growth over 1000 frames: %.2f KB", growthKB)

	if growth > 100*1024 { // 100 KB threshold
		t.Logf("Warning: Heap grew by %.2f KB (may indicate leak)", growthKB)
	} else {
		t.Logf("✅ No memory leak detected (%.2f KB growth is acceptable)", growthKB)
	}
}

// TestRenderSystem_MemoryScaling analyzes memory usage scaling
func TestRenderSystem_MemoryScaling(t *testing.T) {
	entityCounts := []int{100, 500, 1000, 2000, 5000}

	t.Log("\n=== MEMORY SCALING ANALYSIS ===")

	for _, count := range entityCounts {
		// Measure heap before
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		heapBefore := m1.HeapAlloc

		// Create system
		camera := NewCameraSystem(800, 600)
		renderSystem := NewRenderSystem(camera)
		spatialPartition := NewSpatialPartitionSystem(5000, 5000)
		renderSystem.SetSpatialPartition(spatialPartition)

		// Create entities
		entities := make([]*Entity, count)
		sharedSprites := make([]*ebiten.Image, 10)
		for i := 0; i < 10; i++ {
			sharedSprites[i] = ebiten.NewImage(32, 32)
		}

		for i := 0; i < count; i++ {
			entity := NewEntity(uint64(i))

			pos := &PositionComponent{
				X: float64(i % 100 * 50),
				Y: float64(i / 100 * 50),
			}
			entity.AddComponent(pos)

			sprite := &EbitenSprite{
				Image:   sharedSprites[i%10],
				Width:   32,
				Height:  32,
				Visible: true,
			}
			entity.AddComponent(sprite)

			entities[i] = entity
		}

		// Build spatial partition
		spatialPartition.Update(entities, 0)

		// Measure heap after
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		heapAfter := m2.HeapAlloc

		allocMB := float64(heapAfter-heapBefore) / (1024 * 1024)
		bytesPerEntity := (heapAfter - heapBefore) / uint64(count)

		t.Logf("%5d entities: %.2f MB (%.0f bytes/entity)", count, allocMB, float64(bytesPerEntity))

		_ = renderSystem
	}
}

// TestRenderSystem_TotalMemoryFootprint measures total memory consumption
func TestRenderSystem_TotalMemoryFootprint(t *testing.T) {
	t.Log("\n=== TOTAL MEMORY FOOTPRINT ===")

	// Start clean
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	baseHeap := m1.HeapAlloc

	// Create complete system
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)

	// Create 2000 entities (target test case)
	entities := make([]*Entity, 2000)
	sharedSprites := make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		sharedSprites[i] = ebiten.NewImage(32, 32)
	}

	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 100 * 50),
			Y: float64(i / 100 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   sharedSprites[i%10],
			Width:   32,
			Height:  32,
			Visible: true,
		}
		entity.AddComponent(sprite)

		entities[i] = entity
	}

	// Build spatial partition
	spatialPartition.Update(entities, 0)

	// Render a few frames to initialize everything
	screen := ebiten.NewImage(800, 600)
	for i := 0; i < 10; i++ {
		renderSystem.Draw(screen, entities)
	}

	// Measure final heap
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	finalHeap := m2.HeapAlloc

	totalMB := float64(finalHeap-baseHeap) / (1024 * 1024)

	t.Logf("Base heap:  %.2f MB", float64(baseHeap)/(1024*1024))
	t.Logf("Final heap: %.2f MB", float64(finalHeap)/(1024*1024))
	t.Logf("Total used: %.2f MB", totalMB)
	t.Logf("Per entity: %.0f bytes", float64(finalHeap-baseHeap)/2000)

	// Check against 400MB target
	if totalMB < 400 {
		t.Logf("✅ PASS: Total memory (%.2f MB) is well under 400 MB target", totalMB)
	} else {
		t.Errorf("❌ FAIL: Total memory (%.2f MB) exceeds 400 MB target", totalMB)
	}

	// Additional context
	t.Logf("\nSystem Stats:")
	t.Logf("  Sys:          %.2f MB (total from OS)", float64(m2.Sys)/(1024*1024))
	t.Logf("  HeapInuse:    %.2f MB (in-use heap)", float64(m2.HeapInuse)/(1024*1024))
	t.Logf("  HeapIdle:     %.2f MB (idle heap)", float64(m2.HeapIdle)/(1024*1024))
	t.Logf("  NumGC:        %d (garbage collections)", m2.NumGC)
}
