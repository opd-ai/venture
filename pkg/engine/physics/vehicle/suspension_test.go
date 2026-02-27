package vehicle

import (
	"math"
	"testing"
)

func TestNewSuspensionComponent(t *testing.T) {
	tests := []struct {
		name       string
		wheelCount int
		wantWheels int
	}{
		{"one wheel", 1, 1},
		{"two wheels", 2, 2},
		{"three wheels", 3, 3},
		{"four wheels", 4, 4},
		{"six wheels", 6, 6},
		{"zero wheels defaults to four", 0, 4},
		{"negative wheels defaults to four", -1, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSuspensionComponent(tt.wheelCount)
			if comp == nil {
				t.Fatal("NewSuspensionComponent returned nil")
			}
			if len(comp.Wheels) != tt.wantWheels {
				t.Errorf("got %d wheels, want %d", len(comp.Wheels), tt.wantWheels)
			}
			if comp.Type() != "suspension" {
				t.Errorf("got type %q, want %q", comp.Type(), "suspension")
			}
		})
	}
}

func TestSuspensionComponent_Update(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	tests := []struct {
		name           string
		wheelCount     int
		terrainHeights []float64
		deltaTime      float64
		wantOffset     bool // true if offset should be non-zero
	}{
		{
			name:           "four wheels flat terrain",
			wheelCount:     4,
			terrainHeights: []float64{0, 0, 0, 0},
			deltaTime:      0.016,
			wantOffset:     false, // Should stabilize near zero
		},
		{
			name:           "four wheels uneven terrain",
			wheelCount:     4,
			terrainHeights: []float64{0, 5, 0, 5},
			deltaTime:      0.016,
			wantOffset:     true, // Should have offset
		},
		{
			name:           "mismatched terrain count",
			wheelCount:     4,
			terrainHeights: []float64{0, 0}, // Too few
			deltaTime:      0.016,
			wantOffset:     false, // Should return 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSuspensionComponent(tt.wheelCount)

			// Run multiple frames to stabilize
			var offset float64
			for i := 0; i < 10; i++ {
				offset = sys.UpdateSuspensionPhysics(comp, tt.deltaTime, tt.terrainHeights)
			}

			if tt.wantOffset && math.Abs(offset) < 0.01 {
				t.Errorf("expected non-zero offset, got %f", offset)
			}
		})
	}
}

func TestSuspensionComponent_GetWheelLoad(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewSuspensionComponent(4)
	terrainHeights := []float64{0, 0, 0, 0}

	// Update to calculate loads
	sys.UpdateSuspensionPhysics(comp, 0.016, terrainHeights)

	tests := []struct {
		name       string
		wheelIndex int
		wantZero   bool
	}{
		{"valid wheel 0", 0, false},
		{"valid wheel 3", 3, false},
		{"invalid negative", -1, true},
		{"invalid too large", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load := GetWheelLoad(comp, tt.wheelIndex)
			isZero := load == 0.0
			if isZero != tt.wantZero {
				t.Errorf("wheel %d: got load=%f, wantZero=%v", tt.wheelIndex, load, tt.wantZero)
			}
		})
	}
}

func TestSuspensionComponent_GetWheelCompression(t *testing.T) {
	comp := NewSuspensionComponent(4)

	tests := []struct {
		name       string
		wheelIndex int
		want       float64
	}{
		{"valid wheel 0", 0, 0.5}, // Initial compression
		{"valid wheel 3", 3, 0.5},
		{"invalid negative", -1, 0.0},
		{"invalid too large", 10, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compression := comp.GetWheelCompression(tt.wheelIndex)
			if math.Abs(compression-tt.want) > 0.01 {
				t.Errorf("wheel %d: got compression=%f, want=%f", tt.wheelIndex, compression, tt.want)
			}
		})
	}
}

func TestSuspensionComponent_IsWheelGrounded(t *testing.T) {
	comp := NewSuspensionComponent(4)

	// All wheels start grounded
	for i := 0; i < 4; i++ {
		if !comp.IsWheelGrounded(i) {
			t.Errorf("wheel %d should be grounded initially", i)
		}
	}

	// Invalid indices
	if comp.IsWheelGrounded(-1) {
		t.Error("invalid index -1 should return false")
	}
	if comp.IsWheelGrounded(10) {
		t.Error("invalid index 10 should return false")
	}
}

func TestSuspensionComponent_GetGroundedWheelCount(t *testing.T) {
	comp := NewSuspensionComponent(4)

	count := comp.GetGroundedWheelCount()
	if count != 4 {
		t.Errorf("got %d grounded wheels, want 4", count)
	}

	// Manually set one wheel to not grounded
	comp.Wheels[0].IsGrounded = false
	count = comp.GetGroundedWheelCount()
	if count != 3 {
		t.Errorf("got %d grounded wheels, want 3", count)
	}
}

func TestSuspensionComponent_SetWheelLoad(t *testing.T) {
	comp := NewSuspensionComponent(4)

	tests := []struct {
		name       string
		wheelIndex int
		load       float64
		wantSet    bool
	}{
		{"valid wheel 0", 0, 1000.0, true},
		{"valid wheel 3", 3, 2000.0, true},
		{"invalid negative", -1, 500.0, false},
		{"invalid too large", 10, 500.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.SetWheelLoad(tt.wheelIndex, tt.load)

			if tt.wantSet {
				actualLoad := GetWheelLoad(comp, tt.wheelIndex)
				if math.Abs(actualLoad-tt.load) > 0.01 {
					t.Errorf("wheel %d: got load=%f, want=%f", tt.wheelIndex, actualLoad, tt.load)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkSuspensionComponent_Update(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewSuspensionComponent(4)
	terrainHeights := []float64{0, 0, 0, 0}
	deltaTime := 0.016

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.UpdateSuspensionPhysics(comp, deltaTime, terrainHeights)
	}
}

func BenchmarkSuspensionComponent_GetWheelLoad(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewSuspensionComponent(4)
	terrainHeights := []float64{0, 0, 0, 0}
	sys.UpdateSuspensionPhysics(comp, 0.016, terrainHeights)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetWheelLoad(comp, i%4)
	}
}
