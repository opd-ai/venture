package engine

import (
	"testing"
)

// BenchmarkHazardZoneTracker_Update benchmarks the Update method with buffer reuse.
func BenchmarkHazardZoneTracker_Update(b *testing.B) {
	tracker := NewHazardZoneTracker(500)

	// Add some zones with varying durations
	for i := 0; i < 100; i++ {
		zone := &HazardZone{
			ID:                uint64(i + 1),
			X:                 float64(i * 10),
			Y:                 float64(i * 10),
			Radius:            50.0,
			RemainingDuration: float64(i%10) + 1.0, // 1-10 seconds
			Intensity:         1.0,
			DamagePerSecond:   10.0,
			HazardType:        HazardPoison,
		}
		tracker.AddZone(zone)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tracker.Update(0.016) // 60 FPS
	}
}

// BenchmarkHazardZoneTracker_UpdateWithExpiring benchmarks Update with zones expiring.
func BenchmarkHazardZoneTracker_UpdateWithExpiring(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tracker := NewHazardZoneTracker(500)

		// Add zones that will expire quickly
		for j := 0; j < 50; j++ {
			zone := &HazardZone{
				ID:                uint64(j + 1),
				X:                 float64(j * 10),
				Y:                 float64(j * 10),
				Radius:            50.0,
				RemainingDuration: 0.1, // Will expire soon
				Intensity:         1.0,
				DamagePerSecond:   10.0,
				HazardType:        HazardPoison,
			}
			tracker.AddZone(zone)
		}

		// Run multiple updates to expire zones
		for k := 0; k < 10; k++ {
			tracker.Update(0.016)
		}
	}
}

// BenchmarkHazardZoneTracker_GetZonesAtInto benchmarks zero-allocation zone queries.
func BenchmarkHazardZoneTracker_GetZonesAtInto(b *testing.B) {
	tracker := NewHazardZoneTracker(500)

	// Add overlapping zones around center
	for i := 0; i < 20; i++ {
		zone := &HazardZone{
			ID:                uint64(i + 1),
			X:                 100 + float64(i*5),
			Y:                 100 + float64(i*5),
			Radius:            100.0, // Large overlap
			RemainingDuration: -1,    // Permanent
			Intensity:         1.0,
			DamagePerSecond:   10.0,
			HazardType:        HazardPoison,
		}
		tracker.AddZone(zone)
	}

	// Pre-allocate buffer for reuse
	buffer := make([]*HazardZone, 0, 8)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buffer = buffer[:0]
		buffer = tracker.GetZonesAtInto(100, 100, buffer)
	}
}

// BenchmarkHazardZoneTracker_GetZonesAt_Baseline benchmarks the original allocation pattern
// by simulating how the old code worked (allocating per call).
func BenchmarkHazardZoneTracker_GetZonesAt_Baseline(b *testing.B) {
	tracker := NewHazardZoneTracker(500)

	// Add overlapping zones around center
	for i := 0; i < 20; i++ {
		zone := &HazardZone{
			ID:                uint64(i + 1),
			X:                 100 + float64(i*5),
			Y:                 100 + float64(i*5),
			Radius:            100.0, // Large overlap
			RemainingDuration: -1,    // Permanent
			Intensity:         1.0,
			DamagePerSecond:   10.0,
			HazardType:        HazardPoison,
		}
		tracker.AddZone(zone)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate old allocation pattern: allocate new slice each call
		_ = tracker.GetZonesAt(100, 100)
	}
}

// BenchmarkHazardSystem_ApplyZoneEffects benchmarks zone effect application with buffer reuse.
func BenchmarkHazardSystem_ApplyZoneEffects(b *testing.B) {
	world := NewWorld()
	system := NewHazardSystem()
	system.SetWorld(world)

	// Add hazard zones
	for i := 0; i < 10; i++ {
		zone := &HazardZone{
			ID:                 uint64(i + 1),
			X:                  float64(i * 50),
			Y:                  float64(i * 50),
			Radius:             100.0,
			RemainingDuration:  -1,
			Intensity:          1.0,
			DamagePerSecond:    5.0,
			MovementMultiplier: 0.5,
			HazardType:         HazardPoison,
		}
		system.zoneTracker.AddZone(zone)
	}

	// Add entities with position and health
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 5), Y: float64(i * 5)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	// Force entity processing by calling Update with an empty system list
	world.Update(0.016)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.applyZoneEffects(0.016)
	}
}
