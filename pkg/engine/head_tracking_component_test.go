package engine

import (
	"encoding/json"
	"math"
	"testing"
)

func TestNewHeadTrackingComponent(t *testing.T) {
	c := NewHeadTrackingComponent()

	if c == nil {
		t.Fatal("NewHeadTrackingComponent returned nil")
	}

	if c.Enabled {
		t.Error("Expected Enabled to be false by default")
	}

	if c.Pitch != 0 || c.Yaw != 0 || c.Roll != 0 {
		t.Error("Expected all rotations to be 0 by default")
	}

	if c.PositionX != 0 || c.PositionY != 0 || c.PositionZ != 0 {
		t.Error("Expected all positions to be 0 by default")
	}

	if c.SmoothingFactor != DefaultSmoothingFactor {
		t.Errorf("Expected SmoothingFactor %v, got %v", DefaultSmoothingFactor, c.SmoothingFactor)
	}

	if c.PredictionMs != DefaultPredictionMs {
		t.Errorf("Expected PredictionMs %v, got %v", DefaultPredictionMs, c.PredictionMs)
	}
}

func TestHeadTrackingComponent_Type(t *testing.T) {
	c := NewHeadTrackingComponent()
	if c.Type() != "head_tracking" {
		t.Errorf("Expected type 'head_tracking', got '%s'", c.Type())
	}
}

func TestHeadTrackingComponent_SetEnabled(t *testing.T) {
	c := NewHeadTrackingComponent()

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("Expected enabled after SetEnabled(true)")
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("Expected disabled after SetEnabled(false)")
	}
}

func TestHeadTrackingComponent_SetOrientation(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0 // Disable smoothing for direct testing
	c.PredictionMs = 0    // Disable prediction for direct testing

	// Test normal values
	c.SetOrientation(0.5, 1.0, 0.2)
	p, y, r := c.GetOrientation()

	if math.Abs(p-0.5) > 0.01 {
		t.Errorf("Expected pitch 0.5, got %v", p)
	}
	if math.Abs(y-1.0) > 0.01 {
		t.Errorf("Expected yaw 1.0, got %v", y)
	}
	if math.Abs(r-0.2) > 0.01 {
		t.Errorf("Expected roll 0.2, got %v", r)
	}
}

func TestHeadTrackingComponent_SetOrientation_PitchClamp(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0

	// Test pitch clamping at max
	c.SetOrientation(2.0, 0, 0) // Above MaxPitch
	p, _, _ := c.GetOrientation()
	if p > MaxPitch+0.01 {
		t.Errorf("Pitch should be clamped to MaxPitch (%v), got %v", MaxPitch, p)
	}

	// Test pitch clamping at min
	c.SetOrientation(-2.0, 0, 0) // Below MinPitch
	p, _, _ = c.GetOrientation()
	if p < MinPitch-0.01 {
		t.Errorf("Pitch should be clamped to MinPitch (%v), got %v", MinPitch, p)
	}
}

func TestHeadTrackingComponent_SetOrientation_YawNormalize(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0
	c.PredictionMs = 0

	tests := []struct {
		name     string
		inputYaw float64
		wantMin  float64
		wantMax  float64
	}{
		{"positive within range", 1.0, 0.9, 1.1},
		{"negative wrap", -1.0, 2*math.Pi - 1.1, 2*math.Pi - 0.9},
		{"large positive wrap", 3 * math.Pi, math.Pi - 0.1, math.Pi + 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetOrientation(0, tt.inputYaw, 0)
			_, y, _ := c.GetOrientation()
			if y < tt.wantMin || y > tt.wantMax {
				t.Errorf("Yaw %v: got %v, want [%v, %v]", tt.inputYaw, y, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestHeadTrackingComponent_SetPosition(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0
	c.PredictionMs = 0

	c.SetPosition(1.0, 2.0, 3.0)
	x, y, z := c.GetPosition()

	if x != 1.0 {
		t.Errorf("Expected x 1.0, got %v", x)
	}
	if y != 2.0 {
		t.Errorf("Expected y 2.0, got %v", y)
	}
	if z != 3.0 {
		t.Errorf("Expected z 3.0, got %v", z)
	}
}

func TestHeadTrackingComponent_Smoothing(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0.5 // 50% smoothing
	c.PredictionMs = 0

	// Set initial position
	c.SetPosition(0, 0, 0)

	// Set new position - with 50% smoothing, should be halfway
	c.SetPosition(10, 0, 0)
	x, _, _ := c.GetPosition()

	// With 50% smoothing: new = old + (current - old) * 0.5 = 0 + 10 * 0.5 = 5
	if math.Abs(x-5.0) > 0.01 {
		t.Errorf("Expected smoothed x ~5.0, got %v", x)
	}
}

func TestHeadTrackingComponent_SetSmoothingFactor(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 0.5, 0.5},
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
		{"negative clamp", -0.5, 0.0},
		{"above one clamp", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewHeadTrackingComponent()
			c.SetSmoothingFactor(tt.input)
			if c.GetSmoothingFactor() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, c.GetSmoothingFactor())
			}
		})
	}
}

func TestHeadTrackingComponent_SetPredictionMs(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 15.0, 15.0},
		{"zero", 0.0, 0.0},
		{"negative clamp", -10.0, 0.0},
		{"above max clamp", 200.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewHeadTrackingComponent()
			c.SetPredictionMs(tt.input)
			if c.GetPredictionMs() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, c.GetPredictionMs())
			}
		})
	}
}

func TestHeadTrackingComponent_RecenterView(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0
	c.PredictionMs = 0

	// Set initial yaw
	c.SetOrientation(0, 1.5, 0)

	// Recenter
	c.RecenterView()

	// Get orientation - should be adjusted by recenter offset
	_, yaw, _ := c.GetOrientation()

	// After recentering at 1.5, new forward should be at yaw 0
	if math.Abs(yaw) > 0.01 && math.Abs(yaw-2*math.Pi) > 0.01 {
		t.Errorf("Expected yaw ~0 after recenter, got %v", yaw)
	}
}

func TestHeadTrackingComponent_ResetRecenter(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0
	c.PredictionMs = 0

	c.SetOrientation(0, 1.5, 0)
	c.RecenterView()

	// Reset recenter
	c.ResetRecenter()

	_, yaw, _ := c.GetOrientation()

	// After reset, yaw should be back to original
	if math.Abs(yaw-1.5) > 0.01 {
		t.Errorf("Expected yaw ~1.5 after reset, got %v", yaw)
	}
}

func TestHeadTrackingComponent_Prediction(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SmoothingFactor = 0
	c.PredictionMs = 16.67 // One frame prediction

	// Set initial position
	c.SetPosition(0, 0, 0)

	// Set moving position (velocity = 10 per frame)
	c.SetPosition(10, 0, 0)

	// Get predicted position
	x, _, _ := c.GetPosition()

	// With 1 frame prediction, should predict 10 + 10 = 20
	if math.Abs(x-20.0) > 0.5 {
		t.Errorf("Expected predicted x ~20.0, got %v", x)
	}
}

func TestHeadTrackingComponent_Serialize(t *testing.T) {
	c := NewHeadTrackingComponent()
	c.SetEnabled(true)
	c.SmoothingFactor = 0
	c.SetOrientation(0.5, 1.0, 0.2)
	c.SetPosition(1, 2, 3)
	c.SetSmoothingFactor(0.4)
	c.SetPredictionMs(20)

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if m["enabled"] != true {
		t.Error("Expected enabled true")
	}
}

func TestHeadTrackingComponent_Deserialize(t *testing.T) {
	original := NewHeadTrackingComponent()
	original.SetEnabled(true)
	original.SmoothingFactor = 0
	original.SetOrientation(0.5, 1.0, 0.2)
	original.SetPosition(1, 2, 3)
	original.SetSmoothingFactor(0.4)
	original.SetPredictionMs(20)

	data, _ := original.Serialize()

	restored := NewHeadTrackingComponent()
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.IsEnabled() != original.IsEnabled() {
		t.Error("Enabled mismatch")
	}

	// Compare values (allow small tolerance due to float serialization)
	if math.Abs(restored.Pitch-original.Pitch) > 0.001 {
		t.Errorf("Pitch mismatch: %v vs %v", restored.Pitch, original.Pitch)
	}
	if math.Abs(restored.GetSmoothingFactor()-original.GetSmoothingFactor()) > 0.001 {
		t.Errorf("Smoothing mismatch: %v vs %v", restored.GetSmoothingFactor(), original.GetSmoothingFactor())
	}
}

func TestHeadTrackingComponent_ThreadSafety(t *testing.T) {
	c := NewHeadTrackingComponent()

	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetOrientation(float64(i)*0.01, float64(i)*0.02, 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.GetOrientation()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetPosition(float64(i), 0, 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.GetPosition()
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestNormalizeAngleSignedVR(t *testing.T) {
	tests := []struct {
		input   float64
		wantMin float64
		wantMax float64
	}{
		{0, -0.01, 0.01},
		{math.Pi, math.Pi - 0.01, math.Pi + 0.01},
		{-math.Pi, -math.Pi - 0.01, -math.Pi + 0.01},
		{2 * math.Pi, -0.01, 0.01},  // Wraps to 0
		{-2 * math.Pi, -0.01, 0.01}, // Wraps to 0
	}

	for _, tt := range tests {
		result := normalizeAngleSignedVR(tt.input)
		if result < tt.wantMin || result > tt.wantMax {
			t.Errorf("normalizeAngleSignedVR(%v) = %v, want [%v, %v]", tt.input, result, tt.wantMin, tt.wantMax)
		}
		// Should always be in [-π, π]
		if result < -math.Pi-0.01 || result > math.Pi+0.01 {
			t.Errorf("normalizeAngleSignedVR(%v) = %v, outside [-π, π]", tt.input, result)
		}
	}
}

func TestShortestAngularDistanceVR(t *testing.T) {
	tests := []struct {
		from, to float64
		wantMin  float64
		wantMax  float64
	}{
		{0, 0.1, 0.09, 0.11},
		{0, -0.1, -0.11, -0.09},
		{0.1, 2*math.Pi - 0.1, -0.21, -0.19}, // Should wrap around
		{2*math.Pi - 0.1, 0.1, 0.19, 0.21},   // Should wrap around
	}

	for _, tt := range tests {
		result := shortestAngularDistance(tt.from, tt.to)
		if result < tt.wantMin || result > tt.wantMax {
			t.Errorf("shortestAngularDistance(%v, %v) = %v, want [%v, %v]", tt.from, tt.to, result, tt.wantMin, tt.wantMax)
		}
	}
}

func BenchmarkHeadTrackingComponent_SetOrientation(b *testing.B) {
	c := NewHeadTrackingComponent()
	c.SetSmoothingFactor(0.3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetOrientation(float64(i)*0.01, float64(i)*0.02, 0)
	}
}

func BenchmarkHeadTrackingComponent_GetOrientation(b *testing.B) {
	c := NewHeadTrackingComponent()
	c.SetOrientation(0.5, 1.0, 0.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetOrientation()
	}
}
