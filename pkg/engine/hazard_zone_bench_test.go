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
