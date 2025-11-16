package animation

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewController(t *testing.T) {
	gen := sprites.NewGenerator()
	controller := NewController(gen)

	if controller == nil {
		t.Fatal("NewController returned nil")
	}
	if controller.generator == nil {
		t.Error("Controller generator not set")
	}
	if controller.cache == nil {
		t.Error("Controller cache not set")
	}
}

func TestGetFrameCount(t *testing.T) {
	tests := []struct {
		state string
		want  int
	}{
		{"idle", 8},
		{"walk", 8},
		{"run", 8},
		{"attack", 8},
		{"cast", 8},
		{"hit", 4},
		{"death", 8},
		{"jump", 6},
		{"crouch", 4},
		{"use", 6},
		{"unknown", 8}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := GetFrameCount(tt.state)
			if got != tt.want {
				t.Errorf("GetFrameCount(%v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestGetFrameTime(t *testing.T) {
	tests := []struct {
		state   string
		wantMin float64
		wantMax float64
	}{
		{"idle", 0.12, 0.13},   // ~1/8 = 0.125
		{"walk", 0.08, 0.09},   // ~1/12 = 0.083
		{"run", 0.06, 0.07},    // ~1/16 = 0.0625
		{"attack", 0.06, 0.07}, // ~1/16 = 0.0625
		{"hit", 0.04, 0.06},    // ~1/20 = 0.05
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := GetFrameTime(tt.state)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetFrameTime(%v) = %v, want between %v and %v", tt.state, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateDirection8(t *testing.T) {
	// Test convenience wrapper
	dir := CalculateDirection8(1.0, -1.0)
	if dir != Dir8NorthEast {
		t.Errorf("CalculateDirection8(1.0, -1.0) = %v, want NorthEast", dir)
	}
}

func TestSmoothDirectionTransition(t *testing.T) {
	tests := []struct {
		name      string
		oldDir    Direction8
		newDir    Direction8
		deltaTime float64
		wantMin   float64
		wantMax   float64
	}{
		{"fast transition", Dir8North, Dir8East, 1.0, 0.9, 1.0},
		{"slow transition", Dir8North, Dir8East, 0.01, 0.0, 0.2},
		{"zero time", Dir8North, Dir8South, 0.0, 0.0, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blend := SmoothDirectionTransition(tt.oldDir, tt.newDir, tt.deltaTime)
			if blend < tt.wantMin || blend > tt.wantMax {
				t.Errorf("SmoothDirectionTransition() = %v, want between %v and %v", blend, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestControllerSetArticulationConfig(t *testing.T) {
	gen := sprites.NewGenerator()
	controller := NewController(gen)

	customConfig := ArticulationConfig{
		ArmOffsetMax:    5.0,
		LegOffsetMax:    6.0,
		HeadOffsetMax:   3.0,
		TailOffsetMax:   7.0,
		ArmRotationMax:  0.5,
		LegRotationMax:  0.6,
		HeadRotationMax: 0.3,
		TailRotationMax: 0.7,
	}

	controller.SetArticulationConfig(customConfig)

	if controller.config.ArmOffsetMax != 5.0 {
		t.Error("SetArticulationConfig did not update config")
	}
}

func TestControllerClearCache(t *testing.T) {
	gen := sprites.NewGenerator()
	controller := NewController(gen)

	// Cache should start empty
	if controller.cache.Count() != 0 {
		t.Error("Cache should start empty")
	}

	// Clear should not panic on empty cache
	controller.ClearCache()

	if controller.cache.Count() != 0 {
		t.Error("Cache should still be empty after clear")
	}
}

func TestControllerPerformanceMetrics(t *testing.T) {
	gen := sprites.NewGenerator()
	controller := NewController(gen)

	metrics := controller.GetPerformanceMetrics()

	// Should have default values
	if metrics.CacheSize < 0 {
		t.Error("CacheSize should be non-negative")
	}
	if metrics.CacheCount < 0 {
		t.Error("CacheCount should be non-negative")
	}
	if metrics.CacheHitRate < 0 || metrics.CacheHitRate > 100 {
		t.Error("CacheHitRate should be between 0 and 100")
	}
}

func TestGetFrameCountConsistency(t *testing.T) {
	// Verify frame counts are consistent across calls
	states := []string{"idle", "walk", "run", "attack"}

	for _, state := range states {
		count1 := GetFrameCount(state)
		count2 := GetFrameCount(state)
		if count1 != count2 {
			t.Errorf("GetFrameCount(%v) not consistent: %v vs %v", state, count1, count2)
		}
	}
}

func TestGetFrameTimeConsistency(t *testing.T) {
	// Verify frame times are consistent across calls
	states := []string{"idle", "walk", "run"}

	for _, state := range states {
		time1 := GetFrameTime(state)
		time2 := GetFrameTime(state)
		if time1 != time2 {
			t.Errorf("GetFrameTime(%v) not consistent: %v vs %v", state, time1, time2)
		}
	}
}

func TestGetFrameCountAllStates(t *testing.T) {
	// All states should return 8 frames (Phase 46 requirement)
	// except for specific shorter animations
	states := []string{"idle", "walk", "run", "attack", "cast", "death"}

	for _, state := range states {
		count := GetFrameCount(state)
		if count != 8 {
			t.Errorf("GetFrameCount(%v) = %v, Phase 46 requires 8 frames for standard animations", state, count)
		}
	}
}

func BenchmarkGetFrameCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFrameCount("walk")
	}
}

func BenchmarkGetFrameTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFrameTime("walk")
	}
}

func BenchmarkCalculateDirection8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateDirection8(1.0, -1.0)
	}
}

func BenchmarkSmoothDirectionTransition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SmoothDirectionTransition(Dir8North, Dir8East, 0.016)
	}
}
