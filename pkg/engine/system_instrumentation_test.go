package engine

import (
	"testing"
	"time"
)

// TestWorldPerformanceInstrumentation verifies that per-system timing is recorded.
func TestWorldPerformanceInstrumentation(t *testing.T) {
	world := NewWorld()

	// Add a simple test system
	testSys := &testTimedSystem{name: "TestSystem", delay: 5 * time.Millisecond}
	world.AddSystem(testSys)

	// Create a test entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.FlushPendingEntities()

	// Update world (should record system timing)
	world.Update(0.016)

	// Get performance metrics
	metrics := world.GetPerformanceMetrics()

	// Verify system timing was recorded
	if len(metrics.SystemTimes) == 0 {
		t.Fatal("expected system times to be recorded, got empty map")
	}

	// Check for our test system
	systemTime, found := metrics.SystemTimes["testTimedSystem"]
	if !found {
		t.Fatalf("expected 'testTimedSystem' in system times, got keys: %v", getKeys(metrics.SystemTimes))
	}

	// Verify timing is reasonable (should be >= delay)
	if systemTime < testSys.delay {
		t.Errorf("expected system time >= %v, got %v", testSys.delay, systemTime)
	}
}

// TestWorldPerformanceMetricsMultipleSystems verifies timing for multiple systems.
func TestWorldPerformanceMetricsMultipleSystems(t *testing.T) {
	world := NewWorld()

	// Add multiple systems with different delays
	// Note: Both have same type name, so map will have one entry with latest update
	sys1 := &testTimedSystem{name: "FastSystem", delay: 1 * time.Millisecond}
	sys2 := &testTimedSystem{name: "SlowSystem", delay: 10 * time.Millisecond}
	world.AddSystem(sys1)
	world.AddSystem(sys2)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.FlushPendingEntities()

	// Update world
	world.Update(0.016)

	metrics := world.GetPerformanceMetrics()

	// Verify system timing was recorded
	// Since both have same type, we get one entry (last system's time)
	if len(metrics.SystemTimes) < 1 {
		t.Fatalf("expected at least 1 system timed, got %d", len(metrics.SystemTimes))
	}

	// Verify timing is recorded
	for name, duration := range metrics.SystemTimes {
		if duration < 0 {
			t.Errorf("system %s has negative duration: %v", name, duration)
		}
	}
}

// TestGetSystemName verifies system name extraction logic.
func TestGetSystemName(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name           string
		system         System
		expectedSuffix string
	}{
		{
			name:           "test system 1",
			system:         &testTimedSystem{name: "System1"},
			expectedSuffix: "testTimedSystem",
		},
		{
			name:           "test system 2",
			system:         &testTimedSystem{name: "System2"},
			expectedSuffix: "testTimedSystem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			systemName := world.getSystemName(tt.system)
			if systemName == "" {
				t.Error("expected non-empty system name")
			}

			// Check if name contains expected suffix
			if len(systemName) < len(tt.expectedSuffix) ||
				systemName[len(systemName)-len(tt.expectedSuffix):] != tt.expectedSuffix {
				t.Errorf("expected system name to end with %q, got %q", tt.expectedSuffix, systemName)
			}
		})
	}
}

// TestPerformanceMetricsThreadSafety verifies concurrent access to metrics.
func TestPerformanceMetricsThreadSafety(t *testing.T) {
	world := NewWorld()
	sys := &testTimedSystem{delay: 1 * time.Millisecond}
	world.AddSystem(sys)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.FlushPendingEntities()

	// Run updates and metric reads concurrently
	done := make(chan bool)
	go func() {
		for i := 0; i < 50; i++ {
			world.Update(0.016)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			_ = world.GetPerformanceMetrics()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify we can still get metrics
	metrics := world.GetPerformanceMetrics()
	if metrics == nil {
		t.Error("expected non-nil metrics")
	}
}

// TestPerformanceMetricsSnapshot verifies snapshot independence.
func TestPerformanceMetricsSnapshot(t *testing.T) {
	world := NewWorld()
	sys := &testTimedSystem{delay: 5 * time.Millisecond}
	world.AddSystem(sys)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.FlushPendingEntities()

	// Update and get first snapshot
	world.Update(0.016)
	snapshot1 := world.GetPerformanceMetrics()
	systemTime1 := snapshot1.SystemTimes["testTimedSystem"]

	// Update again with different timing
	world.Update(0.016)
	snapshot2 := world.GetPerformanceMetrics()
	systemTime2 := snapshot2.SystemTimes["testTimedSystem"]

	// Verify snapshots are independent copies
	if systemTime1 == 0 || systemTime2 == 0 {
		t.Error("expected non-zero system times in snapshots")
	}

	// Modify original metrics should not affect snapshot1
	originalMetrics := world.performanceMetrics
	originalMetrics.SystemTimes["testTimedSystem"] = 999 * time.Second

	// snapshot1 should still have its original value
	if snapshot1.SystemTimes["testTimedSystem"] == 999*time.Second {
		t.Error("snapshot was modified after being created (not a deep copy)")
	}
}

// BenchmarkSystemInstrumentation measures overhead of per-system timing.
func BenchmarkSystemInstrumentation(b *testing.B) {
	world := NewWorld()

	// Add test systems (avoiding Ebiten dependencies)
	world.AddSystem(&testTimedSystem{delay: 0})
	world.AddSystem(&testTimedSystem{delay: 0})
	world.AddSystem(&testTimedSystem{delay: 0})

	// Create entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
	}
	world.FlushPendingEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world.Update(0.016)
	}
}

// BenchmarkSystemInstrumentationManySystems measures overhead with many systems.
func BenchmarkSystemInstrumentationManySystems(b *testing.B) {
	world := NewWorld()

	// Add many test systems to measure instrumentation overhead
	for i := 0; i < 20; i++ {
		world.AddSystem(&testTimedSystem{delay: 0})
	}

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.FlushPendingEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world.Update(0.016)
	}
}

// testTimedSystem is a test system that simulates work with a configurable delay.
type testTimedSystem struct {
	name  string
	delay time.Duration
}

func (s *testTimedSystem) Update(entities []*Entity, deltaTime float64) {
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
}

// Helper to extract map keys
func getKeys(m map[string]time.Duration) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
