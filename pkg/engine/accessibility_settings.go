// Package engine provides accessibility settings for game feel customization.
// Phase 10.3: Screen Shake & Impact Feedback accessibility
package engine

// AccessibilitySettings controls accessibility features for visual effects.
// Phase 10.3: Allows players to customize or disable screen shake and other
// potentially uncomfortable visual effects.
type AccessibilitySettings struct {
	// Screen shake intensity multiplier (0.0 = disabled, 1.0 = full intensity)
	ScreenShakeIntensity float64

	// Hit-stop enabled flag
	HitStopEnabled bool

	// Visual flash enabled flag (for damage feedback)
	VisualFlashEnabled bool

	// Reduced motion mode (disables all camera effects)
	ReducedMotion bool
}

// NewAccessibilitySettings creates default accessibility settings.
func NewAccessibilitySettings() *AccessibilitySettings {
	return &AccessibilitySettings{
		ScreenShakeIntensity: 1.0,   // Full intensity by default
		HitStopEnabled:       true,  // Enabled by default
		VisualFlashEnabled:   true,  // Enabled by default
		ReducedMotion:        false, // Disabled by default
	}
}

// ApplyShakeIntensity applies accessibility multiplier to shake intensity.
// Returns 0 if reduced motion is enabled or shake is disabled.
func (a *AccessibilitySettings) ApplyShakeIntensity(baseIntensity float64) float64 {
	if a.ReducedMotion {
		return 0.0
	}
	return baseIntensity * a.ScreenShakeIntensity
}

// ShouldApplyHitStop returns true if hit-stop should be applied.
func (a *AccessibilitySettings) ShouldApplyHitStop() bool {
	if a.ReducedMotion {
		return false
	}
	return a.HitStopEnabled
}

// ShouldApplyVisualFlash returns true if visual flash should be applied.
func (a *AccessibilitySettings) ShouldApplyVisualFlash() bool {
	if a.ReducedMotion {
		return false
	}
	return a.VisualFlashEnabled
}

// SetReducedMotion enables or disables reduced motion mode.
// When enabled, disables all camera effects that could cause discomfort.
func (a *AccessibilitySettings) SetReducedMotion(enabled bool) {
	a.ReducedMotion = enabled
}

// SetScreenShakeIntensity sets the screen shake intensity multiplier.
// Value should be between 0.0 (disabled) and 1.0 (full intensity).
// Values > 1.0 are allowed for players who want more intense feedback.
func (a *AccessibilitySettings) SetScreenShakeIntensity(intensity float64) {
	if intensity < 0.0 {
		intensity = 0.0
	}
	a.ScreenShakeIntensity = intensity
}

// SetHitStopEnabled enables or disables hit-stop effects.
func (a *AccessibilitySettings) SetHitStopEnabled(enabled bool) {
	a.HitStopEnabled = enabled
}

// SetVisualFlashEnabled enables or disables visual flash effects.
func (a *AccessibilitySettings) SetVisualFlashEnabled(enabled bool) {
	a.VisualFlashEnabled = enabled
}
