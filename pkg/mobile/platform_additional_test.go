package mobile

import (
	"testing"
)

// TestTriggerHaptic tests haptic feedback triggering.
func TestTriggerHaptic(t *testing.T) {
	// TriggerHaptic should not panic on any platform
	// It's a no-op on most platforms except mobile
	TriggerHaptic(HapticLight)
	TriggerHaptic(HapticMedium)
	TriggerHaptic(HapticHeavy)
}

// TestHapticFeedback tests HapticFeedback type.
func TestHapticFeedback(t *testing.T) {
	tests := []HapticFeedback{
		HapticLight,
		HapticMedium,
		HapticHeavy,
	}

	for _, feedback := range tests {
		// Just verify the constants exist and can be used
		TriggerHaptic(feedback)
	}
}
