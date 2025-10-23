package engine

// VisualFeedbackComponent stores visual feedback state for an entity.
// GAP-012 REPAIR: Provides hit flashes, color tints, and other visual effects.
type VisualFeedbackComponent struct {
	// Flash effect (entity briefly turns white when damaged)
	FlashIntensity float64 // 0.0 = no flash, 1.0 = full white
	FlashDuration  float64 // Total flash duration in seconds
	FlashTimer     float64 // Remaining flash time in seconds

	// Color tint (for status effects, powerups, etc.)
	TintR, TintG, TintB float64 // RGB tint multipliers (1.0 = normal)
	TintA               float64 // Alpha multiplier (1.0 = normal, 0.0 = invisible)
}

// Type returns the component type identifier.
func (v *VisualFeedbackComponent) Type() string {
	return "visual_feedback"
}

// NewVisualFeedbackComponent creates a new visual feedback component with default values.
func NewVisualFeedbackComponent() *VisualFeedbackComponent {
	return &VisualFeedbackComponent{
		FlashIntensity: 0,
		FlashDuration:  0.1, // 100ms flash by default
		FlashTimer:     0,
		TintR:          1.0,
		TintG:          1.0,
		TintB:          1.0,
		TintA:          1.0,
	}
}

// TriggerFlash starts a white flash effect (typically when taking damage).
func (v *VisualFeedbackComponent) TriggerFlash(intensity float64) {
	v.FlashIntensity = intensity
	v.FlashTimer = v.FlashDuration
}

// SetTint sets a color tint for the entity.
func (v *VisualFeedbackComponent) SetTint(r, g, b, a float64) {
	v.TintR = r
	v.TintG = g
	v.TintB = b
	v.TintA = a
}

// ClearTint removes any color tint.
func (v *VisualFeedbackComponent) ClearTint() {
	v.TintR = 1.0
	v.TintG = 1.0
	v.TintB = 1.0
	v.TintA = 1.0
}

// IsFlashing returns true if currently flashing.
func (v *VisualFeedbackComponent) IsFlashing() bool {
	return v.FlashTimer > 0
}

// GetFlashAlpha returns the current flash alpha (0.0-1.0) for blending white color.
func (v *VisualFeedbackComponent) GetFlashAlpha() float64 {
	if v.FlashTimer <= 0 {
		return 0.0
	}
	// Linear fade based on remaining time
	return v.FlashIntensity * (v.FlashTimer / v.FlashDuration)
}

// VisualFeedbackSystem updates visual feedback effects over time.
type VisualFeedbackSystem struct{}

// NewVisualFeedbackSystem creates a new visual feedback system.
func NewVisualFeedbackSystem() *VisualFeedbackSystem {
	return &VisualFeedbackSystem{}
}

// Update decrements flash timers and updates visual effects.
func (s *VisualFeedbackSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		feedbackComp, ok := entity.GetComponent("visual_feedback")
		if !ok {
			continue
		}

		feedback := feedbackComp.(*VisualFeedbackComponent)

		// Update flash timer
		if feedback.FlashTimer > 0 {
			feedback.FlashTimer -= deltaTime
			if feedback.FlashTimer < 0 {
				feedback.FlashTimer = 0
			}
		}
	}
}
