package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkRenderSystem_Performance_Baseline measures baseline performance (no optimizations)
func BenchmarkRenderSystem_Performance_Baseline(b *testing.B) {
	// Create system without optimizations
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	renderSystem.EnableCulling(false)
	renderSystem.EnableBatching(false)

	// Create 2000 entities
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 100 * 50),
			Y: float64(i / 100 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   ebiten.NewImage(32, 32),
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

// BenchmarkRenderSystem_Performance_CullingOnly measures culling-only optimization
func BenchmarkRenderSystem_Performance_CullingOnly(b *testing.B) {
	// Create system with culling only
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)
	renderSystem.EnableBatching(false)

	// Create 2000 entities
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 100 * 50),
			Y: float64(i / 100 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   ebiten.NewImage(32, 32),
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

// BenchmarkRenderSystem_Performance_BatchingOnly measures batching-only optimization
func BenchmarkRenderSystem_Performance_BatchingOnly(b *testing.B) {
	// Create system with batching only
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	renderSystem.EnableCulling(false)
	renderSystem.EnableBatching(true)

	// Create 2000 entities with shared sprites (ideal for batching)
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

// BenchmarkRenderSystem_Performance_AllOptimizations measures all optimizations combined
func BenchmarkRenderSystem_Performance_AllOptimizations(b *testing.B) {
	// Create system with all optimizations
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)
	renderSystem.EnableBatching(true)

	// Create 2000 entities with shared sprites
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

// BenchmarkRenderSystem_Performance_5000Entities stress tests with 5000 entities
func BenchmarkRenderSystem_Performance_5000Entities(b *testing.B) {
	// Create system with all optimizations
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)
	renderSystem.EnableBatching(true)

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

// BenchmarkRenderSystem_Performance_10000Entities extreme stress test
func BenchmarkRenderSystem_Performance_10000Entities(b *testing.B) {
	// Create system with all optimizations
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(10000, 10000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)
	renderSystem.EnableBatching(true)

	// Create 10000 entities
	entities := make([]*Entity, 10000)
	sharedSprites := make([]*ebiten.Image, 50)
	for i := 0; i < 50; i++ {
		sharedSprites[i] = ebiten.NewImage(32, 32)
	}

	for i := 0; i < 10000; i++ {
		entity := NewEntity(uint64(i))

		pos := &PositionComponent{
			X: float64(i % 200 * 50),
			Y: float64(i / 200 * 50),
		}
		entity.AddComponent(pos)

		sprite := &EbitenSprite{
			Image:   sharedSprites[i%50],
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

// BenchmarkRenderSystem_Performance_VariableViewport tests different viewport sizes
func BenchmarkRenderSystem_Performance_VariableViewport(b *testing.B) {
	viewportSizes := []struct {
		name   string
		width  int
		height int
	}{
		{"640x480_VGA", 640, 480},
		{"800x600_SVGA", 800, 600},
		{"1920x1080_FullHD", 1920, 1080},
	}

	for _, size := range viewportSizes {
		b.Run(size.name, func(b *testing.B) {
			camera := NewCameraSystem(size.width, size.height)
			renderSystem := NewRenderSystem(camera)
			spatialPartition := NewSpatialPartitionSystem(5000, 5000)
			renderSystem.SetSpatialPartition(spatialPartition)
			renderSystem.EnableCulling(true)
			renderSystem.EnableBatching(true)

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

			screen := ebiten.NewImage(size.width, size.height)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				renderSystem.Draw(screen, entities)
			}
		})
	}
}

// BenchmarkRenderSystem_Performance_SpriteDiversity tests batching efficiency
func BenchmarkRenderSystem_Performance_SpriteDiversity(b *testing.B) {
	spriteCounts := []struct {
		name   string
		unique int
	}{
		{"LowDiversity_5Sprites", 5},
		{"MediumDiversity_20Sprites", 20},
		{"HighDiversity_100Sprites", 100},
	}

	for _, sc := range spriteCounts {
		b.Run(sc.name, func(b *testing.B) {
			camera := NewCameraSystem(800, 600)
			renderSystem := NewRenderSystem(camera)
			spatialPartition := NewSpatialPartitionSystem(5000, 5000)
			renderSystem.SetSpatialPartition(spatialPartition)
			renderSystem.EnableCulling(true)
			renderSystem.EnableBatching(true)

			// Create 2000 entities with varying sprite diversity
			entities := make([]*Entity, 2000)
			sharedSprites := make([]*ebiten.Image, sc.unique)
			for i := 0; i < sc.unique; i++ {
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
					Image:   sharedSprites[i%sc.unique],
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
		})
	}
}

// BenchmarkRenderSystem_Performance_EntityDensity tests culling efficiency
func BenchmarkRenderSystem_Performance_EntityDensity(b *testing.B) {
	densities := []struct {
		name     string
		entities int
		spread   int // Area spread multiplier
	}{
		{"HighDensity_2000_Small", 2000, 50},
		{"MediumDensity_2000_Medium", 2000, 100},
		{"LowDensity_2000_Large", 2000, 200},
	}

	for _, density := range densities {
		b.Run(density.name, func(b *testing.B) {
			camera := NewCameraSystem(800, 600)
			renderSystem := NewRenderSystem(camera)
			spatialPartition := NewSpatialPartitionSystem(10000, 10000)
			renderSystem.SetSpatialPartition(spatialPartition)
			renderSystem.EnableCulling(true)
			renderSystem.EnableBatching(true)

			// Create entities with varying density
			entities := make([]*Entity, density.entities)
			sharedSprites := make([]*ebiten.Image, 10)
			for i := 0; i < 10; i++ {
				sharedSprites[i] = ebiten.NewImage(32, 32)
			}

			for i := 0; i < density.entities; i++ {
				entity := NewEntity(uint64(i))

				pos := &PositionComponent{
					X: float64(i % 100 * density.spread),
					Y: float64(i / 100 * density.spread),
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
		})
	}
}

// TestRenderSystem_Performance_FrameTimeTarget validates 60 FPS target
func TestRenderSystem_Performance_FrameTimeTarget(t *testing.T) {
	// Target: <16.67ms per frame for 60 FPS
	const targetFrameTimeNS = 16670000 // 16.67ms in nanoseconds

	// Create system with all optimizations
	camera := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(camera)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)
	renderSystem.EnableBatching(true)

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

	// Benchmark frame time
	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			renderSystem.Draw(screen, entities)
		}
	})

	frameTimeNS := result.NsPerOp()
	frameTimeMS := float64(frameTimeNS) / 1000000.0
	fps := 1000.0 / frameTimeMS

	t.Logf("Frame time: %.3f ms (%d ns)", frameTimeMS, frameTimeNS)
	t.Logf("Theoretical FPS: %.0f", fps)

	if frameTimeNS <= targetFrameTimeNS {
		t.Logf("✅ PASS: Frame time (%.3f ms) meets 60 FPS target (<16.67 ms)", frameTimeMS)
	} else {
		t.Errorf("❌ FAIL: Frame time (%.3f ms) exceeds 60 FPS target (16.67 ms)", frameTimeMS)
	}

	// Additional context
	stats := renderSystem.GetStats()
	t.Logf("Rendered entities: %d / %d", stats.RenderedEntities, stats.TotalEntities)
	t.Logf("Culled entities: %d (%.1f%%)", stats.CulledEntities, float64(stats.CulledEntities)/float64(stats.TotalEntities)*100)
	t.Logf("Batch count: %d", stats.BatchCount)
}

// TestRenderSystem_Performance_StressTest validates system under extreme load
func TestRenderSystem_Performance_StressTest(t *testing.T) {
	testCases := []struct {
		name         string
		entityCount  int
		targetTimeMS float64
	}{
		{"Comfortable_2000", 2000, 16.67},
		{"Heavy_5000", 5000, 16.67},
		{"Extreme_10000", 10000, 50.0}, // Allow more time for extreme case
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			camera := NewCameraSystem(800, 600)
			renderSystem := NewRenderSystem(camera)
			spatialPartition := NewSpatialPartitionSystem(10000, 10000)
			renderSystem.SetSpatialPartition(spatialPartition)
			renderSystem.EnableCulling(true)
			renderSystem.EnableBatching(true)

			// Create entities
			entities := make([]*Entity, tc.entityCount)
			sharedSprites := make([]*ebiten.Image, 20)
			for i := 0; i < 20; i++ {
				sharedSprites[i] = ebiten.NewImage(32, 32)
			}

			for i := 0; i < tc.entityCount; i++ {
				entity := NewEntity(uint64(i))

				pos := &PositionComponent{
					X: float64(i % 200 * 50),
					Y: float64(i / 200 * 50),
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

			// Benchmark
			result := testing.Benchmark(func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					renderSystem.Draw(screen, entities)
				}
			})

			frameTimeMS := float64(result.NsPerOp()) / 1000000.0
			fps := 1000.0 / frameTimeMS

			t.Logf("%d entities: %.3f ms/frame (%.0f FPS)", tc.entityCount, frameTimeMS, fps)

			stats := renderSystem.GetStats()
			t.Logf("  Rendered: %d (%.1f%%)", stats.RenderedEntities,
				float64(stats.RenderedEntities)/float64(tc.entityCount)*100)
			t.Logf("  Culled: %d (%.1f%%)", stats.CulledEntities,
				float64(stats.CulledEntities)/float64(tc.entityCount)*100)
			t.Logf("  Batches: %d", stats.BatchCount)

			if frameTimeMS <= tc.targetTimeMS {
				t.Logf("✅ PASS: Meets performance target (%.1f ms)", tc.targetTimeMS)
			} else {
				t.Logf("⚠️  Warning: Exceeds target (%.1f ms target, %.1f ms actual)", tc.targetTimeMS, frameTimeMS)
			}
		})
	}
}
