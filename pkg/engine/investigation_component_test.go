package engine

import (
	"testing"
	"time"
)

// TestInvestigationComponent_Type tests the Type method.
func TestInvestigationComponent_Type(t *testing.T) {
	comp := NewInvestigationComponent()
	if comp.Type() != "investigation" {
		t.Errorf("Expected type 'investigation', got '%s'", comp.Type())
	}
}

// TestNewInvestigationComponent tests component creation with defaults.
func TestNewInvestigationComponent(t *testing.T) {
	comp := NewInvestigationComponent()

	if comp.InvestigationRadius != 3.0 {
		t.Errorf("Expected radius 3.0, got %f", comp.InvestigationRadius)
	}

	if comp.IsInvestigating {
		t.Error("Expected IsInvestigating to be false initially")
	}

	if comp.InvestigationDuration != 2.0 {
		t.Errorf("Expected duration 2.0, got %f", comp.InvestigationDuration)
	}

	if comp.InvestigationCooldown != 1.0 {
		t.Errorf("Expected cooldown 1.0, got %f", comp.InvestigationCooldown)
	}

	if comp.DiscoveredAreas == nil {
		t.Error("Expected DiscoveredAreas map to be initialized")
	}

	if comp.RevealedFragments == nil {
		t.Error("Expected RevealedFragments map to be initialized")
	}
}

// TestInvestigationComponent_StartInvestigation tests starting an investigation.
func TestInvestigationComponent_StartInvestigation(t *testing.T) {
	tests := []struct {
		name            string
		cooldownElapsed float64
		wantSuccess     bool
	}{
		{"No cooldown", 1.0, true},
		{"Cooldown elapsed", 1.5, true},
		{"On cooldown", 0.5, false},
		{"Zero cooldown", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewInvestigationComponent()
			comp.CooldownElapsed = tt.cooldownElapsed

			success := comp.StartInvestigation()

			if success != tt.wantSuccess {
				t.Errorf("StartInvestigation() = %v, want %v", success, tt.wantSuccess)
			}

			if success && !comp.IsInvestigating {
				t.Error("Expected IsInvestigating to be true after successful start")
			}

			if success && comp.TotalInvestigations != 1 {
				t.Errorf("Expected TotalInvestigations to be 1, got %d", comp.TotalInvestigations)
			}

			if success && comp.CooldownElapsed != 0.0 {
				t.Errorf("Expected CooldownElapsed to reset to 0, got %f", comp.CooldownElapsed)
			}
		})
	}
}

// TestInvestigationComponent_StopInvestigation tests stopping an investigation.
func TestInvestigationComponent_StopInvestigation(t *testing.T) {
	comp := NewInvestigationComponent()
	comp.StartInvestigation()

	if !comp.IsInvestigating {
		t.Fatal("Investigation should be active")
	}

	comp.StopInvestigation()

	if comp.IsInvestigating {
		t.Error("Expected IsInvestigating to be false after stop")
	}

	if comp.LastInvestigationTime.IsZero() {
		t.Error("Expected LastInvestigationTime to be set")
	}
}

// TestInvestigationComponent_IsInvestigationComplete tests completion checking.
func TestInvestigationComponent_IsInvestigationComplete(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*InvestigationComponent)
		wantComplete bool
	}{
		{
			name: "Not investigating",
			setupFunc: func(c *InvestigationComponent) {
				c.IsInvestigating = false
			},
			wantComplete: false,
		},
		{
			name: "Just started",
			setupFunc: func(c *InvestigationComponent) {
				c.StartInvestigation()
			},
			wantComplete: false,
		},
		{
			name: "Duration elapsed",
			setupFunc: func(c *InvestigationComponent) {
				c.IsInvestigating = true
				c.InvestigationStartTime = time.Now().Add(-3 * time.Second)
			},
			wantComplete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewInvestigationComponent()
			tt.setupFunc(comp)

			complete := comp.IsInvestigationComplete()

			if complete != tt.wantComplete {
				t.Errorf("IsInvestigationComplete() = %v, want %v", complete, tt.wantComplete)
			}
		})
	}
}

// TestInvestigationComponent_Update tests cooldown timer update.
func TestInvestigationComponent_Update(t *testing.T) {
	comp := NewInvestigationComponent()
	comp.CooldownElapsed = 0.3

	comp.Update(0.5)

	if comp.CooldownElapsed != 0.8 {
		t.Errorf("Expected CooldownElapsed 0.8, got %f", comp.CooldownElapsed)
	}

	// Test clamping
	comp.Update(1.0)

	if comp.CooldownElapsed > comp.InvestigationCooldown {
		t.Errorf("CooldownElapsed should not exceed InvestigationCooldown")
	}
}

// TestInvestigationComponent_GetEffectiveRadius tests radius calculation with skill bonus.
func TestInvestigationComponent_GetEffectiveRadius(t *testing.T) {
	tests := []struct {
		name       string
		baseRadius float64
		skillBonus float64
		wantRadius float64
	}{
		{"No skill bonus", 3.0, 0.0, 3.0},
		{"25% skill bonus", 3.0, 0.25, 3.375}, // 3.0 + (3.0 * 0.5 * 0.25)
		{"50% skill bonus", 3.0, 0.5, 3.75},   // 3.0 + (3.0 * 0.5 * 0.5)
		{"100% skill bonus", 3.0, 1.0, 4.5},   // 3.0 + (3.0 * 0.5 * 1.0)
		{"Different base", 5.0, 0.5, 6.25},    // 5.0 + (5.0 * 0.5 * 0.5)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewInvestigationComponent()
			comp.InvestigationRadius = tt.baseRadius
			comp.InvestigationSkillBonus = tt.skillBonus

			radius := comp.GetEffectiveRadius()

			if radius != tt.wantRadius {
				t.Errorf("GetEffectiveRadius() = %f, want %f", radius, tt.wantRadius)
			}
		})
	}
}

// TestInvestigationComponent_AreaDiscovery tests area discovery tracking.
func TestInvestigationComponent_AreaDiscovery(t *testing.T) {
	comp := NewInvestigationComponent()

	// Initially not discovered
	if comp.HasDiscoveredArea(10, 20) {
		t.Error("Area should not be discovered initially")
	}

	// Mark as discovered
	comp.MarkAreaDiscovered(10, 20)

	if !comp.HasDiscoveredArea(10, 20) {
		t.Error("Area should be discovered after marking")
	}

	// Different area not discovered
	if comp.HasDiscoveredArea(15, 25) {
		t.Error("Different area should not be discovered")
	}

	// Mark multiple areas
	comp.MarkAreaDiscovered(15, 25)
	comp.MarkAreaDiscovered(20, 30)

	if len(comp.DiscoveredAreas) != 3 {
		t.Errorf("Expected 3 discovered areas, got %d", len(comp.DiscoveredAreas))
	}
}

// TestInvestigationComponent_FragmentRevealed tests fragment tracking.
func TestInvestigationComponent_FragmentRevealed(t *testing.T) {
	comp := NewInvestigationComponent()

	fragID := uint64(12345)

	// Initially not revealed
	if comp.HasRevealedFragment(fragID) {
		t.Error("Fragment should not be revealed initially")
	}

	// Mark as revealed
	comp.MarkFragmentRevealed(fragID)

	if !comp.HasRevealedFragment(fragID) {
		t.Error("Fragment should be revealed after marking")
	}

	// Mark multiple fragments
	comp.MarkFragmentRevealed(67890)
	comp.MarkFragmentRevealed(11111)

	if len(comp.RevealedFragments) != 3 {
		t.Errorf("Expected 3 revealed fragments, got %d", len(comp.RevealedFragments))
	}
}

// TestInvestigationComponent_GetDetectionChance tests detection probability calculation.
func TestInvestigationComponent_GetDetectionChance(t *testing.T) {
	tests := []struct {
		name       string
		skillBonus float64
		wantChance float64
	}{
		{"No skill", 0.0, 0.6},
		{"25% skill", 0.25, 0.7}, // 0.6 + (0.4 * 0.25)
		{"50% skill", 0.5, 0.8},  // 0.6 + (0.4 * 0.5)
		{"75% skill", 0.75, 0.9}, // 0.6 + (0.4 * 0.75)
		{"100% skill", 1.0, 1.0}, // 0.6 + (0.4 * 1.0)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewInvestigationComponent()
			comp.InvestigationSkillBonus = tt.skillBonus

			chance := comp.GetDetectionChance()

			if chance != tt.wantChance {
				t.Errorf("GetDetectionChance() = %f, want %f", chance, tt.wantChance)
			}
		})
	}
}

// TestInvestigationComponent_MultipleInvestigations tests multiple investigation cycles.
func TestInvestigationComponent_MultipleInvestigations(t *testing.T) {
	comp := NewInvestigationComponent()

	// First investigation
	if !comp.StartInvestigation() {
		t.Fatal("First investigation should start")
	}

	// Try starting again (should fail - already investigating)
	comp.CooldownElapsed = 2.0  // Reset cooldown manually for test
	comp.IsInvestigating = true // Still investigating
	time.Sleep(10 * time.Millisecond)

	comp.StopInvestigation()

	// Update cooldown
	comp.Update(1.1)

	// Second investigation (cooldown elapsed)
	if !comp.StartInvestigation() {
		t.Error("Second investigation should start after cooldown")
	}

	if comp.TotalInvestigations != 2 {
		t.Errorf("Expected 2 total investigations, got %d", comp.TotalInvestigations)
	}
}

// Benchmark investigation component operations
func BenchmarkInvestigationComponent_StartInvestigation(b *testing.B) {
	comp := NewInvestigationComponent()
	comp.CooldownElapsed = 1.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.StartInvestigation()
		comp.CooldownElapsed = 1.0 // Reset for next iteration
	}
}

func BenchmarkInvestigationComponent_GetEffectiveRadius(b *testing.B) {
	comp := NewInvestigationComponent()
	comp.InvestigationSkillBonus = 0.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.GetEffectiveRadius()
	}
}

func BenchmarkInvestigationComponent_HasDiscoveredArea(b *testing.B) {
	comp := NewInvestigationComponent()
	comp.MarkAreaDiscovered(10, 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.HasDiscoveredArea(10, 20)
	}
}
