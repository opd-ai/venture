// Package engine provides enhanced camera and visual feedback components - tests.
package engine

import (
	"math"
	"testing"
)

// TestScreenShakeComponent_New tests component creation.
func TestScreenShakeComponent_New(t *testing.T) {
	shake := NewScreenShakeComponent()

	if shake == nil {
		t.Fatal("NewScreenShakeComponent() returned nil")
	}

	if shake.Type() != "screenShake" {
		t.Errorf("expected type 'screenShake', got '%s'", shake.Type())
	}

	if shake.Active {
		t.Error("new shake should not be active")
	}

	if shake.Frequency != 15.0 {
		t.Errorf("expected default frequency 15.0, got %.2f", shake.Frequency)
	}
}

// TestScreenShakeComponent_TriggerShake tests shake triggering.
func TestScreenShakeComponent_TriggerShake(t *testing.T) {
	tests := []struct {
		name      string
		intensity float64
		duration  float64
		wantErr   bool
	}{
		{"valid shake", 5.0, 0.3, false},
		{"zero intensity", 0.0, 0.5, false},
		{"negative intensity", -1.0, 0.3, true},
		{"zero duration", 5.0, 0.0, true},
		{"negative duration", 5.0, -0.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shake := NewScreenShakeComponent()
			err := shake.TriggerShake(tt.intensity, tt.duration)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if !shake.Active {
					t.Error("shake should be active after trigger")
				}
				if shake.Intensity != tt.intensity {
					t.Errorf("expected intensity %.2f, got %.2f", tt.intensity, shake.Intensity)
				}
				if shake.Duration != tt.duration {
					t.Errorf("expected duration %.2f, got %.2f", tt.duration, shake.Duration)
				}
			}
		})
	}
}

// TestScreenShakeComponent_StackingShakes tests shake stacking behavior.
func TestScreenShakeComponent_StackingShakes(t *testing.T) {
	shake := NewScreenShakeComponent()

	// First shake
	err := shake.TriggerShake(5.0, 0.3)
	if err != nil {
		t.Fatalf("TriggerShake failed: %v", err)
	}

	// Simulate time passing
	shake.Elapsed = 0.1

	// Second shake (higher intensity)
	err = shake.TriggerShake(8.0, 0.2)
	if err != nil {
		t.Fatalf("TriggerShake failed: %v", err)
	}

	// Should take higher intensity
	if shake.Intensity != 8.0 {
		t.Errorf("expected intensity 8.0 from stacking, got %.2f", shake.Intensity)
	}

	// Duration should extend
	if shake.Duration < 0.3 {
		t.Errorf("expected duration extended to at least 0.3, got %.2f", shake.Duration)
	}
}

// TestScreenShakeComponent_IsShaking tests shake status.
func TestScreenShakeComponent_IsShaking(t *testing.T) {
	shake := NewScreenShakeComponent()

	if shake.IsShaking() {
		t.Error("new shake should not be shaking")
	}

	shake.TriggerShake(5.0, 0.3)
	if !shake.IsShaking() {
		t.Error("shake should be active after trigger")
	}

	// Simulate shake completion
	shake.Elapsed = 0.4
	if shake.IsShaking() {
		t.Error("shake should not be active after duration")
	}
}

// TestScreenShakeComponent_GetProgress tests progress calculation.
func TestScreenShakeComponent_GetProgress(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(5.0, 1.0)

	tests := []struct {
		elapsed  float64
		expected float64
	}{
		{0.0, 0.0},
		{0.25, 0.25},
		{0.5, 0.5},
		{0.75, 0.75},
		{1.0, 1.0},
		{1.5, 1.0}, // Clamped at 1.0
	}

	for _, tt := range tests {
		shake.Elapsed = tt.elapsed
		progress := shake.GetProgress()

		if math.Abs(progress-tt.expected) > 0.001 {
			t.Errorf("elapsed %.2f: expected progress %.2f, got %.2f", tt.elapsed, tt.expected, progress)
		}
	}
}

// TestScreenShakeComponent_GetCurrentIntensity tests intensity decay.
func TestScreenShakeComponent_GetCurrentIntensity(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(10.0, 1.0)

	tests := []struct {
		elapsed  float64
		expected float64
	}{
		{0.0, 10.0}, // Start: 100% intensity
		{0.5, 5.0},  // Middle: 50% intensity
		{0.75, 2.5}, // 75%: 25% intensity
		{1.0, 0.0},  // End: 0% intensity
	}

	for _, tt := range tests {
		shake.Elapsed = tt.elapsed
		intensity := shake.GetCurrentIntensity()

		if math.Abs(intensity-tt.expected) > 0.001 {
			t.Errorf("elapsed %.2f: expected intensity %.2f, got %.2f", tt.elapsed, tt.expected, intensity)
		}
	}
}

// TestScreenShakeComponent_CalculateOffset tests offset calculation.
func TestScreenShakeComponent_CalculateOffset(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(10.0, 1.0)

	// Calculate offset at different times
	shake.Elapsed = 0.0
	shake.CalculateOffset()

	// Should have some offset
	if shake.OffsetX == 0 && shake.OffsetY == 0 {
		t.Error("expected non-zero offset at shake start")
	}

	// Store initial offset
	initialX, initialY := shake.OffsetX, shake.OffsetY

	// Advance time
	shake.Elapsed = 0.1
	shake.CalculateOffset()

	// Offset should change (sine wave)
	if shake.OffsetX == initialX && shake.OffsetY == initialY {
		t.Error("offset should change over time")
	}

	// At end, offset should be zero
	shake.Elapsed = 1.0
	shake.CalculateOffset()
	if shake.OffsetX != 0 || shake.OffsetY != 0 {
		t.Errorf("expected zero offset at end, got (%.2f, %.2f)", shake.OffsetX, shake.OffsetY)
	}
}

// TestScreenShakeComponent_Reset tests reset functionality.
func TestScreenShakeComponent_Reset(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(5.0, 0.3)
	shake.Elapsed = 0.1
	shake.CalculateOffset()

	shake.Reset()

	if shake.Active {
		t.Error("shake should not be active after reset")
	}
	if shake.Elapsed != 0 {
		t.Error("elapsed should be 0 after reset")
	}
	if shake.OffsetX != 0 || shake.OffsetY != 0 {
		t.Error("offsets should be 0 after reset")
	}
}

// TestHitStopComponent_New tests component creation.
func TestHitStopComponent_New(t *testing.T) {
	hitStop := NewHitStopComponent()

	if hitStop == nil {
		t.Fatal("NewHitStopComponent() returned nil")
	}

	if hitStop.Type() != "hitStop" {
		t.Errorf("expected type 'hitStop', got '%s'", hitStop.Type())
	}

	if hitStop.Active {
		t.Error("new hit-stop should not be active")
	}

	if hitStop.TimeScale != 0.0 {
		t.Errorf("expected default time scale 0.0, got %.2f", hitStop.TimeScale)
	}
}

// TestHitStopComponent_TriggerHitStop tests hit-stop triggering.
func TestHitStopComponent_TriggerHitStop(t *testing.T) {
	tests := []struct {
		name      string
		duration  float64
		timeScale float64
		wantErr   bool
	}{
		{"valid hit-stop", 0.1, 0.0, false},
		{"slow motion", 0.2, 0.1, false},
		{"zero duration", 0.0, 0.0, true},
		{"negative duration", -0.1, 0.0, true},
		{"negative time scale", 0.1, -0.1, true},
		{"time scale > 1", 0.1, 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hitStop := NewHitStopComponent()
			err := hitStop.TriggerHitStop(tt.duration, tt.timeScale)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if !hitStop.Active {
					t.Error("hit-stop should be active after trigger")
				}
				if hitStop.Duration != tt.duration {
					t.Errorf("expected duration %.2f, got %.2f", tt.duration, hitStop.Duration)
				}
				if hitStop.TimeScale != tt.timeScale {
					t.Errorf("expected time scale %.2f, got %.2f", tt.timeScale, hitStop.TimeScale)
				}
			}
		})
	}
}

// TestHitStopComponent_StackingHitStops tests hit-stop stacking.
func TestHitStopComponent_StackingHitStops(t *testing.T) {
	hitStop := NewHitStopComponent()

	// First hit-stop
	err := hitStop.TriggerHitStop(0.1, 0.1)
	if err != nil {
		t.Fatalf("TriggerHitStop failed: %v", err)
	}

	// Simulate time passing
	hitStop.Elapsed = 0.05

	// Second hit-stop (more dramatic)
	err = hitStop.TriggerHitStop(0.15, 0.0)
	if err != nil {
		t.Fatalf("TriggerHitStop failed: %v", err)
	}

	// Should take lower time scale (more dramatic)
	if hitStop.TimeScale != 0.0 {
		t.Errorf("expected time scale 0.0 from stacking, got %.2f", hitStop.TimeScale)
	}

	// Duration should extend
	if hitStop.Duration < 0.15 {
		t.Errorf("expected duration extended to at least 0.15, got %.2f", hitStop.Duration)
	}
}

// TestHitStopComponent_IsActive tests active status.
func TestHitStopComponent_IsActive(t *testing.T) {
	hitStop := NewHitStopComponent()

	if hitStop.IsActive() {
		t.Error("new hit-stop should not be active")
	}

	hitStop.TriggerHitStop(0.1, 0.0)
	if !hitStop.IsActive() {
		t.Error("hit-stop should be active after trigger")
	}

	// Simulate completion
	hitStop.Elapsed = 0.15
	if hitStop.IsActive() {
		t.Error("hit-stop should not be active after duration")
	}
}

// TestHitStopComponent_GetTimeScale tests time scale retrieval.
func TestHitStopComponent_GetTimeScale(t *testing.T) {
	hitStop := NewHitStopComponent()

	// Not active: should return 1.0 (normal time)
	if hitStop.GetTimeScale() != 1.0 {
		t.Errorf("expected time scale 1.0 when inactive, got %.2f", hitStop.GetTimeScale())
	}

	// Active: should return set time scale
	hitStop.TriggerHitStop(0.1, 0.05)
	if hitStop.GetTimeScale() != 0.05 {
		t.Errorf("expected time scale 0.05 when active, got %.2f", hitStop.GetTimeScale())
	}

	// After duration: should return 1.0 again
	hitStop.Elapsed = 0.15
	if hitStop.GetTimeScale() != 1.0 {
		t.Errorf("expected time scale 1.0 after duration, got %.2f", hitStop.GetTimeScale())
	}
}

// TestHitStopComponent_Reset tests reset functionality.
func TestHitStopComponent_Reset(t *testing.T) {
	hitStop := NewHitStopComponent()
	hitStop.TriggerHitStop(0.1, 0.0)
	hitStop.Elapsed = 0.05

	hitStop.Reset()

	if hitStop.Active {
		t.Error("hit-stop should not be active after reset")
	}
	if hitStop.Elapsed != 0 {
		t.Error("elapsed should be 0 after reset")
	}
	if hitStop.GetTimeScale() != 1.0 {
		t.Errorf("time scale should be 1.0 after reset, got %.2f", hitStop.GetTimeScale())
	}
}

// TestCalculateShakeIntensity tests shake intensity calculation.
func TestCalculateShakeIntensity(t *testing.T) {
	tests := []struct {
		name         string
		damage       float64
		maxHP        float64
		scaleFactor  float64
		minIntensity float64
		maxIntensity float64
		expected     float64
	}{
		{"10 damage to 100 HP", 10, 100, 10, 1, 20, 1},    // 10/100*10 = 1 → clamped to min 1
		{"50 damage to 100 HP", 50, 100, 10, 1, 20, 5},    // 50/100*10 = 5
		{"100 damage to 100 HP", 100, 100, 10, 1, 20, 10}, // 100/100*10 = 10
		{"200 damage to 100 HP", 200, 100, 10, 1, 20, 20}, // 200/100*10 = 20 → clamped to max 20
		{"small damage", 1, 100, 10, 2, 20, 2},            // 1/100*10 = 0.1 → clamped to min 2
		{"zero max HP", 50, 0, 10, 1, 20, 5},              // Uses default 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateShakeIntensity(tt.damage, tt.maxHP, tt.scaleFactor, tt.minIntensity, tt.maxIntensity)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestCalculateShakeDuration tests shake duration calculation.
func TestCalculateShakeDuration(t *testing.T) {
	tests := []struct {
		name               string
		intensity          float64
		baseDuration       float64
		additionalDuration float64
		maxIntensity       float64
		expected           float64
	}{
		{"low intensity", 5, 0.1, 0.2, 20, 0.15},    // 0.1 + (5/20)*0.2 = 0.15
		{"medium intensity", 10, 0.1, 0.2, 20, 0.2}, // 0.1 + (10/20)*0.2 = 0.2
		{"high intensity", 15, 0.1, 0.2, 20, 0.25},  // 0.1 + (15/20)*0.2 = 0.25
		{"max intensity", 20, 0.1, 0.2, 20, 0.3},    // 0.1 + (20/20)*0.2 = 0.3
		{"over max", 30, 0.1, 0.2, 20, 0.3},         // Clamped to max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateShakeDuration(tt.intensity, tt.baseDuration, tt.additionalDuration, tt.maxIntensity)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %.3f, got %.3f", tt.expected, result)
			}
		})
	}
}

// Benchmark tests

func BenchmarkScreenShakeComponent_TriggerShake(b *testing.B) {
	shake := NewScreenShakeComponent()
	for i := 0; i < b.N; i++ {
		shake.TriggerShake(5.0, 0.3)
	}
}

func BenchmarkScreenShakeComponent_CalculateOffset(b *testing.B) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(5.0, 0.3)

	for i := 0; i < b.N; i++ {
		shake.Elapsed = float64(i) * 0.016 // 60 FPS simulation
		shake.CalculateOffset()
	}
}

func BenchmarkHitStopComponent_TriggerHitStop(b *testing.B) {
	hitStop := NewHitStopComponent()
	for i := 0; i < b.N; i++ {
		hitStop.TriggerHitStop(0.1, 0.0)
	}
}
