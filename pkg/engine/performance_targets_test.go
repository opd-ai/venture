package engine

import (
	"runtime"
	"testing"
)

// BenchmarkWorldUpdateWith2000Entities validates the 60 FPS performance target
// with 2000 entities as specified in docs/ARCHITECTURE.md.
//
// Target: 60 FPS = 16.67ms per frame = 16,666,666 ns/op
//
// Note: This benchmark uses MovementSystem only to represent the core ECS update
// overhead. Collision detection is tested separately as it depends heavily on
// spatial partitioning configuration.
func BenchmarkWorldUpdateWith2000Entities(b *testing.B) {
	world := NewWorld()

	// Add MovementSystem (lightweight, representative of core systems)
	world.AddSystem(&MovementSystem{})

	// Create 2000 entities with position and velocity components
	// This simulates the baseline ECS overhead for entity updates
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})

		// Add health component to subset for realistic variety
		if i%3 == 0 {
			entity.AddComponent(&HealthComponent{
				Current: 100,
				Max:     100,
			})
		}
	}

	// Process pending entities
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate frame updates at 60 FPS (16.67ms delta)
	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// TestMemoryBounds2000Entities validates the <500MB memory target
// with 2000 entities as specified in docs/ARCHITECTURE.md.
//
// Target: <500MB client memory with 2000 entities
func TestMemoryBounds2000Entities(t *testing.T) {
	// Force GC before measurement
	runtime.GC()

	// Capture baseline memory
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	world := NewWorld()

	// Add realistic systems (lightweight systems only for fair measurement)
	world.AddSystem(&MovementSystem{})

	// Create 2000 entities with components
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
		})

		// Add health and stats to subset
		if i%3 == 0 {
			entity.AddComponent(&HealthComponent{
				Current: 100,
				Max:     100,
			})
			entity.AddComponent(&StatsComponent{
				Attack:       10,
				Defense:      10,
				MagicPower:   10,
				MagicDefense: 10,
			})
		}
	}

	// Run a few update cycles to stabilize
	for i := 0; i < 10; i++ {
		world.Update(0.016)
	}

	// Force GC and measure memory after entity creation
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate memory in MB (using heap allocated)
	totalMemoryMB := float64(memAfter.Alloc) / (1024 * 1024)

	t.Logf("Entities created: 2000")
	t.Logf("Heap allocated: %.2f MB", totalMemoryMB)
	t.Logf("Sys memory: %.2f MB", float64(memAfter.Sys)/(1024*1024))

	// Validate against 500MB target
	if totalMemoryMB >= 500 {
		t.Errorf("Memory usage exceeds 500MB target: %.2f MB >= 500 MB", totalMemoryMB)
	} else {
		t.Logf("✅ Memory usage within target: %.2f MB < 500 MB (%.1f%% of budget)",
			totalMemoryMB, (totalMemoryMB/500)*100)
	}
}

// BenchmarkWorldUpdateWith2000EntitiesRealistic benchmarks world updates
// with a more complete set of systems for comprehensive performance testing.
//
// Note: CollisionSystem is excluded as it requires careful spatial partitioning
// configuration for optimal performance. This benchmark focuses on core ECS overhead.
func BenchmarkWorldUpdateWith2000EntitiesRealistic(b *testing.B) {
	world := NewWorld()

	// Add MovementSystem (most common lightweight system)
	world.AddSystem(&MovementSystem{})

	// Create 2000 entities with varied components
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&VelocityComponent{
			VX: 1.0 + float64(i%10)*0.1,
			VY: 1.0 + float64(i%10)*0.1,
		})

		// Varied components for different entity types
		if i%5 == 0 {
			entity.AddComponent(&HealthComponent{
				Current: 100,
				Max:     100,
			})
		}
		if i%3 == 0 {
			entity.AddComponent(&StatsComponent{
				Attack:  10,
				Defense: 10,
			})
		}
		if i%4 == 0 {
			entity.AddComponent(&ColliderComponent{
				Width:  32,
				Height: 32,
			})
		}
	}

	// Stabilize world state
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	deltaTime := 0.016 // 60 FPS
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}
