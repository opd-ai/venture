// Package engine provides visual feedback system tests.
package engine

import (
	"testing"
)

// TestVisualFeedbackComponent_TriggerFlash tests flash triggering.
func TestVisualFeedbackComponent_TriggerFlash(t *testing.T) {
	comp := NewVisualFeedbackComponent()

	// Initially not flashing
	if comp.IsFlashing() {
		t.Error("Component should not be flashing initially")
	}

	// Trigger flash
	comp.TriggerFlash(0.8)

	// Should be flashing now
	if !comp.IsFlashing() {
		t.Error("Component should be flashing after TriggerFlash")
	}

	// Flash intensity should be set
	if comp.FlashIntensity != 0.8 {
		t.Errorf("Flash intensity = %f, want 0.8", comp.FlashIntensity)
	}

	// Flash timer should be positive
	if comp.FlashTimer <= 0 {
		t.Errorf("Flash timer = %f, want > 0", comp.FlashTimer)
	}
}

// TestVisualFeedbackComponent_GetFlashAlpha tests alpha calculation.
func TestVisualFeedbackComponent_GetFlashAlpha(t *testing.T) {
	comp := NewVisualFeedbackComponent()

	// Initially no flash
	alpha := comp.GetFlashAlpha()
	if alpha != 0.0 {
		t.Errorf("Initial flash alpha = %f, want 0.0", alpha)
	}

	// Trigger flash
	comp.TriggerFlash(1.0)

	// At start of flash, alpha should be maximum (1.0 * intensity)
	alpha = comp.GetFlashAlpha()
	if alpha <= 0.0 || alpha > 1.0 {
		t.Errorf("Flash alpha = %f, want 0.0 < alpha <= 1.0", alpha)
	}

	// After half duration, alpha should be ~0.5
	comp.FlashTimer = comp.FlashDuration / 2
	alpha = comp.GetFlashAlpha()
	if alpha < 0.4 || alpha > 0.6 {
		t.Errorf("Mid-flash alpha = %f, want ~0.5", alpha)
	}

	// After full duration, alpha should be 0
	comp.FlashTimer = 0
	alpha = comp.GetFlashAlpha()
	if alpha != 0.0 {
		t.Errorf("Expired flash alpha = %f, want 0.0", alpha)
	}
}

// TestVisualFeedbackComponent_Tint tests color tinting.
func TestVisualFeedbackComponent_Tint(t *testing.T) {
	comp := NewVisualFeedbackComponent()

	// Initially no tint (all 1.0)
	if comp.TintR != 1.0 || comp.TintG != 1.0 || comp.TintB != 1.0 || comp.TintA != 1.0 {
		t.Errorf("Initial tint = (%f,%f,%f,%f), want (1,1,1,1)",
			comp.TintR, comp.TintG, comp.TintB, comp.TintA)
	}

	// Set red tint
	comp.SetTint(1.0, 0.5, 0.5, 1.0)
	if comp.TintR != 1.0 || comp.TintG != 0.5 || comp.TintB != 0.5 || comp.TintA != 1.0 {
		t.Errorf("Red tint = (%f,%f,%f,%f), want (1,0.5,0.5,1)",
			comp.TintR, comp.TintG, comp.TintB, comp.TintA)
	}

	// Clear tint
	comp.ClearTint()
	if comp.TintR != 1.0 || comp.TintG != 1.0 || comp.TintB != 1.0 || comp.TintA != 1.0 {
		t.Errorf("After clear tint = (%f,%f,%f,%f), want (1,1,1,1)",
			comp.TintR, comp.TintG, comp.TintB, comp.TintA)
	}
}

// TestVisualFeedbackSystem_Update tests flash timer decay.
func TestVisualFeedbackSystem_Update(t *testing.T) {
	system := NewVisualFeedbackSystem()
	world := NewWorld()

	// Create entity with visual feedback
	entity := world.CreateEntity()
	comp := NewVisualFeedbackComponent()
	comp.TriggerFlash(1.0)
	initialTimer := comp.FlashTimer
	entity.AddComponent(comp)

	// Process pending additions
	world.Update(0.0)

	// Update system with 0.05 seconds
	system.Update(world.GetEntities(), 0.05)

	// Flash timer should have decreased
	if comp.FlashTimer >= initialTimer {
		t.Errorf("Flash timer = %f, want < %f", comp.FlashTimer, initialTimer)
	}

	// Flash should still be active (default duration is 0.1s)
	if !comp.IsFlashing() {
		t.Error("Flash should still be active after 0.05s")
	}

	// Update past flash duration
	system.Update(world.GetEntities(), 0.1)

	// Flash should be expired
	if comp.IsFlashing() {
		t.Error("Flash should be expired after 0.15s total")
	}

	// Timer should be zero
	if comp.FlashTimer != 0.0 {
		t.Errorf("Flash timer = %f, want 0.0", comp.FlashTimer)
	}
}

// TestVisualFeedbackComponent_MultipleFlashes tests flash stacking behavior.
func TestVisualFeedbackComponent_MultipleFlashes(t *testing.T) {
	comp := NewVisualFeedbackComponent()

	// Trigger first flash
	comp.TriggerFlash(0.5)

	// Wait a bit (simulate time passing)
	comp.FlashTimer -= 0.05
	timer1 := comp.FlashTimer

	// Trigger second flash (should reset timer and update intensity)
	comp.TriggerFlash(1.0)
	timer2 := comp.FlashTimer

	// Timer should be reset to full duration (greater than reduced timer)
	if timer2 <= timer1 {
		t.Errorf("Second flash timer = %f, want > %f (should reset to default)", timer2, timer1)
	}

	// Intensity should be updated
	if comp.FlashIntensity != 1.0 {
		t.Errorf("Flash intensity = %f, want 1.0", comp.FlashIntensity)
	}
}

// TestVisualFeedbackComponent_Type tests component type identifier.
func TestVisualFeedbackComponent_Type(t *testing.T) {
	comp := NewVisualFeedbackComponent()
	if comp.Type() != "visual_feedback" {
		t.Errorf("Component type = %s, want 'visual_feedback'", comp.Type())
	}
}

// TestVisualFeedbackComponent_IntensityClamping tests intensity bounds.
func TestVisualFeedbackComponent_IntensityClamping(t *testing.T) {
	comp := NewVisualFeedbackComponent()

	tests := []struct {
		name      string
		intensity float64
		wantAlpha float64 // Maximum possible alpha
	}{
		{"zero intensity", 0.0, 0.0},
		{"low intensity", 0.3, 0.3},
		{"normal intensity", 0.8, 0.8},
		{"max intensity", 1.0, 1.0},
		{"over max intensity", 1.5, 1.5}, // System doesn't clamp, combat system does
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.TriggerFlash(tt.intensity)
			alpha := comp.GetFlashAlpha()

			// Alpha should be positive if flashing
			if tt.intensity > 0 && alpha <= 0 {
				t.Errorf("Flash alpha = %f, want > 0", alpha)
			}

			// Alpha should not exceed intensity * timer ratio
			maxAlpha := tt.intensity
			if alpha > maxAlpha+0.01 { // Small epsilon for float comparison
				t.Errorf("Flash alpha = %f, exceeds max %f", alpha, maxAlpha)
			}
		})
	}
}
