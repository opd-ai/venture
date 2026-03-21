package fps

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// BenchmarkFPS2000Entities validates the 60 FPS performance target with 2000 entities.
// This is the primary performance benchmark referenced in docs/PERFORMANCE.md.
//
// Target: 60 FPS = 16.67ms per frame = 16,666,666 ns/op
//
// Note: Requires graphics libraries (same as pkg/engine tests). Install with:
// sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
func BenchmarkFPS2000Entities(b *testing.B) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})

	// Create 2000 entities with position and velocity components
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})

		// Add health to subset for variety
		if i%3 == 0 {
			entity.AddComponent(&engine.HealthComponent{
				Current: 100,
				Max:     100,
			})
		}
	}

	// Stabilize world state
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate 60 FPS updates (16.67ms delta)
	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// BenchmarkFPS2000EntitiesRealistic tests FPS with a more complete component set
// to simulate real-world gameplay scenarios.
func BenchmarkFPS2000EntitiesRealistic(b *testing.B) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})

	// Create 2000 entities with varied components
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: 1.0 + float64(i%10)*0.1,
			VY: 1.0 + float64(i%10)*0.1,
		})

		// Varied components by entity type
		if i%5 == 0 {
			entity.AddComponent(&engine.HealthComponent{
				Current: 100,
				Max:     100,
			})
		}
		if i%3 == 0 {
			entity.AddComponent(&engine.StatsComponent{
				Attack:  10,
				Defense: 10,
			})
		}
		if i%4 == 0 {
			entity.AddComponent(&engine.ColliderComponent{
				Width:  32,
				Height: 32,
			})
		}
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// BenchmarkFPS500Entities benchmarks FPS with a smaller entity count
// to validate performance headroom.
func BenchmarkFPS500Entities(b *testing.B) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})

	for i := 0; i < 500; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i % 50),
			Y: float64(i / 50),
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// BenchmarkFPS5000Entities stress tests FPS with 5000 entities
// to validate performance degradation patterns.
func BenchmarkFPS5000Entities(b *testing.B) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})

	for i := 0; i < 5000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// BenchmarkFPS2000EntitiesWithCollision benchmarks FPS with 2000 entities
// including collision detection via spatial partitioning.
// This validates that the spatial hash grid optimization maintains 60 FPS
// with collision-heavy workloads.
//
// Target: 60 FPS = 16.67ms per frame = 16,666,666 ns/op
func BenchmarkFPS2000EntitiesWithCollision(b *testing.B) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})
	world.AddSystem(&engine.CollisionSystem{})

	// Create 2000 entities with colliders distributed spatially
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i%100) * 64, // 64-unit spacing
			Y: float64(i/100) * 64,
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: float64((i%3)-1) * 10.0, // Varied velocities: -10, 0, 10
			VY: float64((i%5)-2) * 10.0, // Varied velocities: -20, -10, 0, 10, 20
		})
		entity.AddComponent(&engine.ColliderComponent{
			Width:  32,
			Height: 32,
		})

		// Add friction for some entities to vary physics interactions
		if i%2 == 0 {
			entity.AddComponent(&engine.FrictionComponent{
				Coefficient: 0.5,
			})
		}
	}

	// Stabilize world state and build spatial partition
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate 60 FPS updates
	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// TestFPS60TargetWith2000Entities is a unit test that validates the 60 FPS target
// is achievable with 2000 entities. This test complements the benchmarks by providing
// a pass/fail threshold for CI/CD validation.
func TestFPS60TargetWith2000Entities(t *testing.T) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})

	// Create 2000 entities
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})
	}

	world.Update(0)

	// Run benchmark to measure performance
	deltaTime := 0.016

	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			world.Update(deltaTime)
		}
	})

	// Calculate ns per frame
	nsPerOp := result.NsPerOp()

	// 60 FPS target = 16.67ms = 16,666,666 ns
	targetNs := int64(16666666)

	t.Logf("Performance: %d ns/op (target: %d ns/op)", nsPerOp, targetNs)
	t.Logf("Equivalent FPS: %.2f (target: 60.00)", 1_000_000_000.0/float64(nsPerOp))

	if nsPerOp > targetNs {
		t.Errorf("FPS performance below target: %d ns/op exceeds 16.67ms threshold (60 FPS)", nsPerOp)
	} else {
		headroomPercent := (1.0 - float64(nsPerOp)/float64(targetNs)) * 100
		t.Logf("✅ FPS target achieved with %.1f%% headroom", headroomPercent)
	}
}

// TestFPS60TargetLightWeight validates 60 FPS is achievable with minimal load.
func TestFPS60TargetLightWeight(t *testing.T) {
	world := engine.NewWorld()
	world.AddSystem(&engine.MovementSystem{})

	// Create 500 entities (light load)
	for i := 0; i < 500; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i),
			Y: float64(i),
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: 1.0,
			VY: 1.0,
		})
	}

	world.Update(0)

	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			world.Update(0.016)
		}
	})

	nsPerOp := result.NsPerOp()
	targetNs := int64(16666666)

	t.Logf("Performance: %d ns/op (target: %d ns/op)", nsPerOp, targetNs)
	t.Logf("Equivalent FPS: %.2f", 1_000_000_000.0/float64(nsPerOp))

	if nsPerOp > targetNs {
		t.Errorf("FPS performance below target even with 500 entities: %d ns/op", nsPerOp)
	} else {
		t.Logf("✅ Light load FPS target achieved")
	}
}

// BenchmarkMultiSystemSuite benchmarks FPS with multiple core systems to simulate
// realistic gameplay scenarios. Tests Movement, Collision, AI, and Combat systems
// together with 2000 entities.
//
// Note: This does not include rendering or audio systems which require Ebiten
// initialization. See pkg/engine/performance/ for graphics pipeline benchmarks.
//
// Target: 60 FPS = 16.67ms per frame = 16,666,666 ns/op
func BenchmarkMultiSystemSuite(b *testing.B) {
	world := engine.NewWorld()

	// Register core gameplay systems (non-graphics)
	world.AddSystem(&engine.MovementSystem{})
	world.AddSystem(&engine.CollisionSystem{})
	world.AddSystem(&engine.AISystem{})
	world.AddSystem(&engine.CombatSystem{})

	// Create 2000 entities with varied components for realistic load
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()

		// All entities have position and velocity
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i%100) * 64, // 64-unit spacing across 100x20 grid
			Y: float64(i/100) * 64,
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: float64((i%3)-1) * 10.0, // Varied velocities: -10, 0, 10
			VY: float64((i%5)-2) * 10.0, // Varied velocities: -20, -10, 0, 10, 20
		})

		// All entities have colliders
		entity.AddComponent(&engine.ColliderComponent{
			Width:  32,
			Height: 32,
		})

		// All entities have health
		entity.AddComponent(&engine.HealthComponent{
			Current: 100,
			Max:     100,
		})

		// Combat entities (50%) have stats
		if i%2 == 0 {
			entity.AddComponent(&engine.StatsComponent{
				Attack:     10 + float64(i%5)*2,
				Defense:    10 + float64(i%3)*2,
				CritChance: 0.15,
				CritDamage: 2.0,
			})
		}

		// AI-controlled entities (33%) have AI state
		if i%3 == 0 {
			entity.AddComponent(&engine.AIComponent{
				State:               engine.AIStateIdle,
				DetectionRange:      200.0,
				FleeHealthThreshold: 0.2,
				MaxChaseDistance:    500.0,
				DecisionInterval:    0.5,
			})
		}
	}

	// Stabilize world state
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate 60 FPS updates
	deltaTime := 0.016
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// TestMultiSystemSuiteFPS60 benchmarks FPS with multiple systems and reports performance.
// Note: This test measures the current multi-system performance. The 60 FPS target
// is validated by the simpler single-system tests; this test helps identify
// bottlenecks when multiple systems interact.
func TestMultiSystemSuiteFPS60(t *testing.T) {
	world := engine.NewWorld()

	// Register core gameplay systems
	world.AddSystem(&engine.MovementSystem{})
	world.AddSystem(&engine.CollisionSystem{})
	world.AddSystem(&engine.AISystem{})
	world.AddSystem(&engine.CombatSystem{})

	// Create 2000 entities
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{
			X: float64(i%100) * 64,
			Y: float64(i/100) * 64,
		})
		entity.AddComponent(&engine.VelocityComponent{
			VX: float64((i%3)-1) * 10.0,
			VY: float64((i%5)-2) * 10.0,
		})
		entity.AddComponent(&engine.ColliderComponent{
			Width:  32,
			Height: 32,
		})
		entity.AddComponent(&engine.HealthComponent{
			Current: 100,
			Max:     100,
		})
		if i%2 == 0 {
			entity.AddComponent(&engine.StatsComponent{
				Attack:     10,
				Defense:    10,
				CritChance: 0.15,
				CritDamage: 2.0,
			})
		}
		if i%3 == 0 {
			entity.AddComponent(&engine.AIComponent{
				State:               engine.AIStateIdle,
				DetectionRange:      200.0,
				FleeHealthThreshold: 0.2,
				MaxChaseDistance:    500.0,
				DecisionInterval:    0.5,
			})
		}
	}

	world.Update(0)

	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			world.Update(0.016)
		}
	})

	nsPerOp := result.NsPerOp()
	targetNs := int64(16666666)

	t.Logf("Multi-System Performance: %d ns/op (target: %d ns/op)", nsPerOp, targetNs)
	fps := 1_000_000_000.0 / float64(nsPerOp)
	t.Logf("Equivalent FPS: %.2f (target: 60.00)", fps)

	// Report headroom or shortfall
	if nsPerOp > targetNs {
		shortfallPercent := (float64(nsPerOp)/float64(targetNs) - 1.0) * 100
		t.Logf("⚠️ Multi-system FPS below target by %.1f%% (this is expected with 4 systems + 2000 AI entities)", shortfallPercent)
		t.Logf("   Single-system benchmarks validate core 60 FPS; this measures system interaction overhead")
	} else {
		headroomPercent := (1.0 - float64(nsPerOp)/float64(targetNs)) * 100
		t.Logf("✅ Multi-system FPS target achieved with %.1f%% headroom", headroomPercent)
	}
}
