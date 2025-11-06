// Package engine provides enhanced camera and visual feedback components.
// Phase 10.3: Screen Shake & Impact Feedback
//
// This file adds advanced screen shake and hit-stop components to complement
// the existing CameraSystem (camera_system.go). The existing CameraComponent
// has basic shake support; these components provide more advanced control.
package engine

import (
	"fmt"
	"math"
)

// ScreenShakeComponent adds advanced procedural screen shake effects.
type ScreenShakeComponent struct {
	// Shake intensity (pixels)
	Intensity float64

	// Shake duration (seconds)
	Duration float64

	// Elapsed time (seconds)
	Elapsed float64

	// Shake frequency (Hz)
	Frequency float64

	// Current offset (calculated by system)
	OffsetX, OffsetY float64

	// Active flag
	Active bool
}

// Type returns the component type identifier.
func (s *ScreenShakeComponent) Type() string {
	return "screenShake"
}

// NewScreenShakeComponent creates a new screen shake component.
func NewScreenShakeComponent() *ScreenShakeComponent {
	return &ScreenShakeComponent{
		Intensity: 0,
		Duration:  0,
		Elapsed:   0,
		Frequency: 15.0, // 15 Hz default (fast shake)
		OffsetX:   0,
		OffsetY:   0,
		Active:    false,
	}
}

// TriggerShake starts a new screen shake effect.
func (s *ScreenShakeComponent) TriggerShake(intensity, duration float64) error {
	if intensity < 0 {
		return fmt.Errorf("intensity must be non-negative, got %.2f", intensity)
	}
	if duration <= 0 {
		return fmt.Errorf("duration must be positive, got %.2f", duration)
	}

	// If already shaking, add to intensity (stack shakes)
	if s.Active && s.Elapsed < s.Duration {
		// Take the maximum intensity and extend duration
		if intensity > s.Intensity {
			s.Intensity = intensity
		}
		if duration > (s.Duration - s.Elapsed) {
			s.Duration = s.Elapsed + duration
		}
	} else {
		// Start new shake
		s.Intensity = intensity
		s.Duration = duration
		s.Elapsed = 0
		s.Active = true
	}

	return nil
}

// IsShaking returns true if shake is active.
func (s *ScreenShakeComponent) IsShaking() bool {
	return s.Active && s.Elapsed < s.Duration
}

// GetProgress returns shake progress (0 to 1).
func (s *ScreenShakeComponent) GetProgress() float64 {
	if s.Duration <= 0 {
		return 1.0
	}
	return math.Min(s.Elapsed/s.Duration, 1.0)
}

// GetCurrentIntensity returns intensity with decay applied.
func (s *ScreenShakeComponent) GetCurrentIntensity() float64 {
	if !s.IsShaking() {
		return 0
	}
	// Linear decay
	progress := s.GetProgress()
	return s.Intensity * (1.0 - progress)
}

// CalculateOffset updates the shake offset based on elapsed time.
func (s *ScreenShakeComponent) CalculateOffset() {
	if !s.IsShaking() {
		s.OffsetX = 0
		s.OffsetY = 0
		return
	}

	currentIntensity := s.GetCurrentIntensity()

	// Use sine wave for smooth oscillation
	angle := s.Elapsed * s.Frequency * 2 * math.Pi

	// Two perpendicular sine waves for circular-ish shake
	s.OffsetX = currentIntensity * math.Sin(angle)
	s.OffsetY = currentIntensity * math.Sin(angle*1.3+math.Pi/4) // Slightly different frequency and phase
}

// Reset stops the shake and resets to zero.
func (s *ScreenShakeComponent) Reset() {
	s.Active = false
	s.Elapsed = 0
	s.OffsetX = 0
	s.OffsetY = 0
}

// HitStopComponent adds time dilation / hit-stop effects.
type HitStopComponent struct {
	// Duration of hit-stop (seconds)
	Duration float64

	// Elapsed time in hit-stop (seconds)
	Elapsed float64

	// Active flag
	Active bool

	// Time scale during hit-stop (0 = full stop, 0.1 = slow motion)
	TimeScale float64
}

// Type returns the component type identifier.
func (h *HitStopComponent) Type() string {
	return "hitStop"
}

// NewHitStopComponent creates a new hit-stop component.
func NewHitStopComponent() *HitStopComponent {
	return &HitStopComponent{
		Duration:  0,
		Elapsed:   0,
		Active:    false,
		TimeScale: 0.0, // Full stop by default
	}
}

// TriggerHitStop starts a hit-stop effect.
func (h *HitStopComponent) TriggerHitStop(duration, timeScale float64) error {
	if duration <= 0 {
		return fmt.Errorf("duration must be positive, got %.2f", duration)
	}
	if timeScale < 0 || timeScale > 1 {
		return fmt.Errorf("timeScale must be between 0 and 1, got %.2f", timeScale)
	}

	// If already in hit-stop, extend duration
	if h.Active && h.Elapsed < h.Duration {
		if duration > (h.Duration - h.Elapsed) {
			h.Duration = h.Elapsed + duration
		}
		// Use minimum time scale (most dramatic slowdown)
		if timeScale < h.TimeScale {
			h.TimeScale = timeScale
		}
	} else {
		// Start new hit-stop
		h.Duration = duration
		h.TimeScale = timeScale
		h.Elapsed = 0
		h.Active = true
	}

	return nil
}

// IsActive returns true if hit-stop is active.
func (h *HitStopComponent) IsActive() bool {
	return h.Active && h.Elapsed < h.Duration
}

// GetTimeScale returns the current time scale.
func (h *HitStopComponent) GetTimeScale() float64 {
	if h.IsActive() {
		return h.TimeScale
	}
	return 1.0 // Normal time
}

// Reset stops the hit-stop effect.
func (h *HitStopComponent) Reset() {
	h.Active = false
	h.Elapsed = 0
}

// CalculateShakeIntensity calculates shake intensity based on damage.
// Phase 10.3: Helper for damage-based shake scaling.
// Formula: intensity = clamp(damage / maxHP * scaleFactor, minIntensity, maxIntensity)
func CalculateShakeIntensity(damage, maxHP, scaleFactor, minIntensity, maxIntensity float64) float64 {
	if maxHP <= 0 {
		maxHP = 100 // Default to avoid division by zero
	}

	intensity := (damage / maxHP) * scaleFactor
	intensity = math.Max(minIntensity, intensity)
	intensity = math.Min(maxIntensity, intensity)

	return intensity
}

// CalculateShakeDuration calculates shake duration based on intensity.
// Phase 10.3: Helper for intensity-based duration scaling.
// Formula: duration = baseDuration + (intensity / maxIntensity) * additionalDuration
func CalculateShakeDuration(intensity, baseDuration, additionalDuration, maxIntensity float64) float64 {
	if maxIntensity <= 0 {
		maxIntensity = 20 // Default
	}

	ratio := math.Min(intensity/maxIntensity, 1.0)
	return baseDuration + ratio*additionalDuration
}
