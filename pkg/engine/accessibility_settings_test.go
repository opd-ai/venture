package engine

import "testing"

func TestNewAccessibilitySettings(t *testing.T) {
	settings := NewAccessibilitySettings()

	if settings.ScreenShakeIntensity != 1.0 {
		t.Errorf("Expected default ScreenShakeIntensity 1.0, got %v", settings.ScreenShakeIntensity)
	}
	if !settings.HitStopEnabled {
		t.Error("Expected HitStopEnabled to be true by default")
	}
	if !settings.VisualFlashEnabled {
		t.Error("Expected VisualFlashEnabled to be true by default")
	}
	if settings.ReducedMotion {
		t.Error("Expected ReducedMotion to be false by default")
	}
}

func TestApplyShakeIntensity(t *testing.T) {
	tests := []struct {
		name              string
		shakeIntensity    float64
		reducedMotion     bool
		baseIntensity     float64
		expectedIntensity float64
	}{
		{
			name:              "full intensity",
			shakeIntensity:    1.0,
			reducedMotion:     false,
			baseIntensity:     10.0,
			expectedIntensity: 10.0,
		},
		{
			name:              "half intensity",
			shakeIntensity:    0.5,
			reducedMotion:     false,
			baseIntensity:     10.0,
			expectedIntensity: 5.0,
		},
		{
			name:              "zero intensity",
			shakeIntensity:    0.0,
			reducedMotion:     false,
			baseIntensity:     10.0,
			expectedIntensity: 0.0,
		},
		{
			name:              "reduced motion overrides",
			shakeIntensity:    1.0,
			reducedMotion:     true,
			baseIntensity:     10.0,
			expectedIntensity: 0.0,
		},
		{
			name:              "double intensity (power users)",
			shakeIntensity:    2.0,
			reducedMotion:     false,
			baseIntensity:     10.0,
			expectedIntensity: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := NewAccessibilitySettings()
			settings.ScreenShakeIntensity = tt.shakeIntensity
			settings.ReducedMotion = tt.reducedMotion

			result := settings.ApplyShakeIntensity(tt.baseIntensity)
			if result != tt.expectedIntensity {
				t.Errorf("ApplyShakeIntensity() = %v, want %v", result, tt.expectedIntensity)
			}
		})
	}
}

func TestShouldApplyHitStop(t *testing.T) {
	tests := []struct {
		name           string
		hitStopEnabled bool
		reducedMotion  bool
		expected       bool
	}{
		{
			name:           "enabled and no reduced motion",
			hitStopEnabled: true,
			reducedMotion:  false,
			expected:       true,
		},
		{
			name:           "disabled",
			hitStopEnabled: false,
			reducedMotion:  false,
			expected:       false,
		},
		{
			name:           "reduced motion overrides enabled",
			hitStopEnabled: true,
			reducedMotion:  true,
			expected:       false,
		},
		{
			name:           "both disabled",
			hitStopEnabled: false,
			reducedMotion:  true,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := NewAccessibilitySettings()
			settings.HitStopEnabled = tt.hitStopEnabled
			settings.ReducedMotion = tt.reducedMotion

			result := settings.ShouldApplyHitStop()
			if result != tt.expected {
				t.Errorf("ShouldApplyHitStop() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldApplyVisualFlash(t *testing.T) {
	tests := []struct {
		name               string
		visualFlashEnabled bool
		reducedMotion      bool
		expected           bool
	}{
		{
			name:               "enabled and no reduced motion",
			visualFlashEnabled: true,
			reducedMotion:      false,
			expected:           true,
		},
		{
			name:               "disabled",
			visualFlashEnabled: false,
			reducedMotion:      false,
			expected:           false,
		},
		{
			name:               "reduced motion overrides enabled",
			visualFlashEnabled: true,
			reducedMotion:      true,
			expected:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := NewAccessibilitySettings()
			settings.VisualFlashEnabled = tt.visualFlashEnabled
			settings.ReducedMotion = tt.reducedMotion

			result := settings.ShouldApplyVisualFlash()
			if result != tt.expected {
				t.Errorf("ShouldApplyVisualFlash() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSetReducedMotion(t *testing.T) {
	settings := NewAccessibilitySettings()

	// Test enabling
	settings.SetReducedMotion(true)
	if !settings.ReducedMotion {
		t.Error("SetReducedMotion(true) did not enable reduced motion")
	}

	// Test disabling
	settings.SetReducedMotion(false)
	if settings.ReducedMotion {
		t.Error("SetReducedMotion(false) did not disable reduced motion")
	}
}

func TestSetScreenShakeIntensity(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "normal value",
			input:    0.5,
			expected: 0.5,
		},
		{
			name:     "zero",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "full intensity",
			input:    1.0,
			expected: 1.0,
		},
		{
			name:     "negative clamped to zero",
			input:    -0.5,
			expected: 0.0,
		},
		{
			name:     "over 1.0 allowed",
			input:    2.0,
			expected: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := NewAccessibilitySettings()
			settings.SetScreenShakeIntensity(tt.input)

			if settings.ScreenShakeIntensity != tt.expected {
				t.Errorf("SetScreenShakeIntensity(%v) resulted in %v, want %v",
					tt.input, settings.ScreenShakeIntensity, tt.expected)
			}
		})
	}
}

func TestSetHitStopEnabled(t *testing.T) {
	settings := NewAccessibilitySettings()

	// Test disabling
	settings.SetHitStopEnabled(false)
	if settings.HitStopEnabled {
		t.Error("SetHitStopEnabled(false) did not disable hit-stop")
	}

	// Test enabling
	settings.SetHitStopEnabled(true)
	if !settings.HitStopEnabled {
		t.Error("SetHitStopEnabled(true) did not enable hit-stop")
	}
}

func TestSetVisualFlashEnabled(t *testing.T) {
	settings := NewAccessibilitySettings()

	// Test disabling
	settings.SetVisualFlashEnabled(false)
	if settings.VisualFlashEnabled {
		t.Error("SetVisualFlashEnabled(false) did not disable visual flash")
	}

	// Test enabling
	settings.SetVisualFlashEnabled(true)
	if !settings.VisualFlashEnabled {
		t.Error("SetVisualFlashEnabled(true) did not enable visual flash")
	}
}

func TestAccessibilitySettings_Integration(t *testing.T) {
	// Test realistic usage scenario
	settings := NewAccessibilitySettings()

	// User with motion sensitivity
	settings.SetScreenShakeIntensity(0.3) // Reduce shake to 30%
	settings.SetHitStopEnabled(false)     // Disable time dilation

	// Apply to shake
	baseShake := 10.0
	appliedShake := settings.ApplyShakeIntensity(baseShake)
	expectedShake := 3.0
	if appliedShake != expectedShake {
		t.Errorf("Expected shake %v, got %v", expectedShake, appliedShake)
	}

	// Check hit-stop
	if settings.ShouldApplyHitStop() {
		t.Error("Hit-stop should be disabled")
	}

	// Enable reduced motion (should override all)
	settings.SetReducedMotion(true)
	appliedShake = settings.ApplyShakeIntensity(baseShake)
	if appliedShake != 0.0 {
		t.Errorf("Reduced motion should disable shake, got %v", appliedShake)
	}
}
