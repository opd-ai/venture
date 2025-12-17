// Package engine provides tests for the NG+ Difficulty component.
// Phase 113: Difficulty Scaling System
package engine

import (
	"encoding/json"
	"math"
	"testing"
)

func TestNGPlusDifficultyComponent_Type(t *testing.T) {
	comp := NewNGPlusDifficultyComponent()
	if comp.Type() != "ngplus_difficulty" {
		t.Errorf("Type() = %v, want ngplus_difficulty", comp.Type())
	}
}

func TestNewNGPlusDifficultyComponent(t *testing.T) {
	comp := NewNGPlusDifficultyComponent()

	// Verify default values
	if comp.HealthMultiplier != 1.0 {
		t.Errorf("HealthMultiplier = %v, want 1.0", comp.HealthMultiplier)
	}
	if comp.DamageMultiplier != 1.0 {
		t.Errorf("DamageMultiplier = %v, want 1.0", comp.DamageMultiplier)
	}
	if comp.DefenseMultiplier != 1.0 {
		t.Errorf("DefenseMultiplier = %v, want 1.0", comp.DefenseMultiplier)
	}
	if comp.LootQualityBonus != 0.0 {
		t.Errorf("LootQualityBonus = %v, want 0.0", comp.LootQualityBonus)
	}
	if comp.XPMultiplier != 1.0 {
		t.Errorf("XPMultiplier = %v, want 1.0", comp.XPMultiplier)
	}
	if comp.NewMechanicsLevel != 0 {
		t.Errorf("NewMechanicsLevel = %v, want 0", comp.NewMechanicsLevel)
	}
	if comp.NGPlusCycle != 0 {
		t.Errorf("NGPlusCycle = %v, want 0", comp.NGPlusCycle)
	}
	if comp.IsScaled {
		t.Errorf("IsScaled = true, want false")
	}
}

func TestNewNGPlusDifficultyComponentForCycle(t *testing.T) {
	tests := []struct {
		name              string
		cycle             int
		expectHealthAbove float64
		expectDamageAbove float64
		expectMechanics   int
	}{
		{"first_playthrough", 0, 0.99, 0.99, 0},
		{"ngplus_1", 1, 1.1, 1.05, 0},
		{"ngplus_3", 3, 1.2, 1.1, 1},
		{"ngplus_5", 5, 1.3, 1.2, 2},
		{"ngplus_7", 7, 1.35, 1.25, 3},
		{"ngplus_10", 10, 1.4, 1.3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewNGPlusDifficultyComponentForCycle(tt.cycle)

			if comp.GetHealthMultiplier() < tt.expectHealthAbove {
				t.Errorf("HealthMultiplier = %v, want > %v", comp.GetHealthMultiplier(), tt.expectHealthAbove)
			}
			if comp.GetDamageMultiplier() < tt.expectDamageAbove {
				t.Errorf("DamageMultiplier = %v, want > %v", comp.GetDamageMultiplier(), tt.expectDamageAbove)
			}
			if comp.GetNewMechanicsLevel() != tt.expectMechanics {
				t.Errorf("NewMechanicsLevel = %v, want %v", comp.GetNewMechanicsLevel(), tt.expectMechanics)
			}
			if tt.cycle > 0 && !comp.IsScaled {
				t.Errorf("IsScaled = false, want true for cycle %d", tt.cycle)
			}
		})
	}
}

func TestApplyScalingForCycle(t *testing.T) {
	comp := NewNGPlusDifficultyComponent()

	// Apply scaling for NG+5
	comp.ApplyScalingForCycle(5)

	if comp.GetNGPlusCycle() != 5 {
		t.Errorf("NGPlusCycle = %v, want 5", comp.GetNGPlusCycle())
	}
	if !comp.IsScaled {
		t.Error("IsScaled = false, want true")
	}

	// Health should be scaled up
	if comp.GetHealthMultiplier() <= 1.0 {
		t.Errorf("HealthMultiplier = %v, should be > 1.0 for NG+5", comp.GetHealthMultiplier())
	}

	// XP should be slightly reduced
	if comp.GetXPMultiplier() >= 1.0 {
		t.Errorf("XPMultiplier = %v, should be < 1.0 for NG+5", comp.GetXPMultiplier())
	}

	// Loot quality should be improved
	if comp.GetLootQualityBonus() <= 0.0 {
		t.Errorf("LootQualityBonus = %v, should be > 0.0 for NG+5", comp.GetLootQualityBonus())
	}
}

func TestApplyScalingForCycle_XPMinimum(t *testing.T) {
	// Test that XP multiplier never goes below 0.5
	comp := NewNGPlusDifficultyComponent()
	comp.ApplyScalingForCycle(99) // Very high NG+ level

	if comp.GetXPMultiplier() < 0.5 {
		t.Errorf("XPMultiplier = %v, should not go below 0.5", comp.GetXPMultiplier())
	}
}

func TestApplyBossEnhancements(t *testing.T) {
	tests := []struct {
		name          string
		cycle         int
		expectLevel   int
		expectEnraged bool
	}{
		{"cycle_0", 0, 0, false},
		{"cycle_1", 1, 0, false},
		{"cycle_2", 2, 1, false},
		{"cycle_5", 5, 2, false},
		{"cycle_10", 10, 3, true},
		{"cycle_15", 15, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewNGPlusDifficultyComponent()
			comp.ApplyBossEnhancements(tt.cycle)

			if comp.GetBossEnhancementLevel() != tt.expectLevel {
				t.Errorf("BossEnhancementLevel = %v, want %v", comp.GetBossEnhancementLevel(), tt.expectLevel)
			}
			if comp.HasEnragedPhase != tt.expectEnraged {
				t.Errorf("HasEnragedPhase = %v, want %v", comp.HasEnragedPhase, tt.expectEnraged)
			}
		})
	}
}

func TestGetAbilitiesForMechanicsLevel(t *testing.T) {
	tests := []struct {
		level          int
		expectCount    int
		expectContains string
	}{
		{0, 0, ""},
		{1, 2, "counter_attack"},
		{2, 4, "elemental_shield"},
		{3, 6, "summon_minions"},
		{4, 8, "ultimate_attack"},
	}

	for _, tt := range tests {
		t.Run("level_"+string(rune('0'+tt.level)), func(t *testing.T) {
			abilities := getAbilitiesForMechanicsLevel(tt.level)

			if len(abilities) != tt.expectCount {
				t.Errorf("len(abilities) = %v, want %v", len(abilities), tt.expectCount)
			}

			if tt.expectContains != "" {
				found := false
				for _, a := range abilities {
					if a == tt.expectContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("abilities should contain %s", tt.expectContains)
				}
			}
		})
	}
}

func TestHasAbility(t *testing.T) {
	comp := NewNGPlusDifficultyComponentForCycle(5) // Level 2 mechanics

	// Should have level 1 and 2 abilities
	if !comp.HasAbility("counter_attack") {
		t.Error("Should have counter_attack ability at NG+5")
	}
	if !comp.HasAbility("elemental_shield") {
		t.Error("Should have elemental_shield ability at NG+5")
	}
	// Should not have level 3+ abilities
	if comp.HasAbility("summon_minions") {
		t.Error("Should not have summon_minions ability at NG+5")
	}
}

func TestGetAdditionalAbilities(t *testing.T) {
	comp := NewNGPlusDifficultyComponentForCycle(3)
	abilities := comp.GetAdditionalAbilities()

	// Should return a copy, not the original slice
	if len(abilities) == 0 {
		t.Error("Should have abilities at NG+3")
	}

	// Modifying the returned slice should not affect the component
	originalLen := len(comp.GetAdditionalAbilities())
	abilities[0] = "modified"
	if comp.GetAdditionalAbilities()[0] == "modified" {
		t.Error("GetAdditionalAbilities should return a copy")
	}
	if len(comp.GetAdditionalAbilities()) != originalLen {
		t.Error("Component abilities should be unchanged")
	}
}

func TestScaledHealth(t *testing.T) {
	comp := NewNGPlusDifficultyComponentForCycle(5)
	baseHealth := 100.0
	scaled := comp.ScaledHealth(baseHealth)

	// Should be higher than base
	if scaled <= baseHealth {
		t.Errorf("ScaledHealth = %v, want > %v", scaled, baseHealth)
	}

	// Verify it uses the multiplier correctly
	expected := baseHealth * comp.GetHealthMultiplier()
	if math.Abs(scaled-expected) > 0.001 {
		t.Errorf("ScaledHealth = %v, want %v", scaled, expected)
	}
}

func TestScaledDamage(t *testing.T) {
	comp := NewNGPlusDifficultyComponentForCycle(5)
	baseDamage := 50.0
	scaled := comp.ScaledDamage(baseDamage)

	if scaled <= baseDamage {
		t.Errorf("ScaledDamage = %v, want > %v", scaled, baseDamage)
	}
}

func TestScaledDefense(t *testing.T) {
	comp := NewNGPlusDifficultyComponentForCycle(5)
	baseDefense := 25.0
	scaled := comp.ScaledDefense(baseDefense)

	if scaled <= baseDefense {
		t.Errorf("ScaledDefense = %v, want > %v", scaled, baseDefense)
	}
}

func TestScaledXP(t *testing.T) {
	comp := NewNGPlusDifficultyComponentForCycle(5)
	baseXP := 100.0
	scaled := comp.ScaledXP(baseXP)

	// XP should be reduced at higher NG+ levels
	if scaled >= baseXP {
		t.Errorf("ScaledXP = %v, want < %v for NG+5", scaled, baseXP)
	}
}

func TestNGPlusDifficultyComponent_SerializeDeserialize(t *testing.T) {
	original := NewNGPlusDifficultyComponentForCycle(7)
	original.ApplyBossEnhancements(7)

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize
	restored := NewNGPlusDifficultyComponent()
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify values match
	if restored.GetHealthMultiplier() != original.GetHealthMultiplier() {
		t.Errorf("HealthMultiplier mismatch: %v != %v", restored.GetHealthMultiplier(), original.GetHealthMultiplier())
	}
	if restored.GetDamageMultiplier() != original.GetDamageMultiplier() {
		t.Errorf("DamageMultiplier mismatch: %v != %v", restored.GetDamageMultiplier(), original.GetDamageMultiplier())
	}
	if restored.GetNGPlusCycle() != original.GetNGPlusCycle() {
		t.Errorf("NGPlusCycle mismatch: %v != %v", restored.GetNGPlusCycle(), original.GetNGPlusCycle())
	}
	if restored.GetBossEnhancementLevel() != original.GetBossEnhancementLevel() {
		t.Errorf("BossEnhancementLevel mismatch: %v != %v", restored.GetBossEnhancementLevel(), original.GetBossEnhancementLevel())
	}
	if len(restored.GetAdditionalAbilities()) != len(original.GetAdditionalAbilities()) {
		t.Errorf("AdditionalAbilities length mismatch: %v != %v", len(restored.GetAdditionalAbilities()), len(original.GetAdditionalAbilities()))
	}
}

func TestDeserialize_InvalidData(t *testing.T) {
	comp := NewNGPlusDifficultyComponent()
	err := comp.Deserialize([]byte("invalid json"))

	if err == nil {
		t.Error("Deserialize should fail on invalid JSON")
	}
}

func TestGetDifficultyLabel(t *testing.T) {
	tests := []struct {
		cycle       int
		expectLabel string
	}{
		{0, "Normal"},
		{1, "Challenging"},
		{2, "Challenging"},
		{3, "Hard"},
		{5, "Hell"},
		{7, "Nightmare"},
		{10, "Legendary"},
		{15, "Legendary"},
	}

	for _, tt := range tests {
		t.Run("cycle_"+string(rune('0'+tt.cycle%10)), func(t *testing.T) {
			comp := NewNGPlusDifficultyComponentForCycle(tt.cycle)
			label := comp.GetDifficultyLabel()

			if label != tt.expectLabel {
				t.Errorf("GetDifficultyLabel() = %v, want %v for cycle %d", label, tt.expectLabel, tt.cycle)
			}
		})
	}
}

func TestThreadSafety(t *testing.T) {
	comp := NewNGPlusDifficultyComponent()
	done := make(chan bool, 10)

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = comp.GetHealthMultiplier()
				_ = comp.GetDamageMultiplier()
				_ = comp.GetXPMultiplier()
				_ = comp.HasAbility("test")
			}
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(cycle int) {
			for j := 0; j < 100; j++ {
				comp.ApplyScalingForCycle(cycle)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDeterminism(t *testing.T) {
	// Same cycle should always produce same values
	comp1 := NewNGPlusDifficultyComponentForCycle(5)
	comp2 := NewNGPlusDifficultyComponentForCycle(5)

	if comp1.GetHealthMultiplier() != comp2.GetHealthMultiplier() {
		t.Error("HealthMultiplier should be deterministic")
	}
	if comp1.GetDamageMultiplier() != comp2.GetDamageMultiplier() {
		t.Error("DamageMultiplier should be deterministic")
	}
	if comp1.GetNewMechanicsLevel() != comp2.GetNewMechanicsLevel() {
		t.Error("NewMechanicsLevel should be deterministic")
	}
}

func BenchmarkApplyScalingForCycle(b *testing.B) {
	comp := NewNGPlusDifficultyComponent()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comp.ApplyScalingForCycle(i % 20)
	}
}

func BenchmarkScaledHealth(b *testing.B) {
	comp := NewNGPlusDifficultyComponentForCycle(5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = comp.ScaledHealth(100.0)
	}
}

func BenchmarkNGPlusDifficultySerialize(b *testing.B) {
	comp := NewNGPlusDifficultyComponentForCycle(10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = comp.Serialize()
	}
}

func BenchmarkNGPlusDifficultyDeserialize(b *testing.B) {
	comp := NewNGPlusDifficultyComponentForCycle(10)
	data, _ := comp.Serialize()
	target := NewNGPlusDifficultyComponent()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = target.Deserialize(data)
	}
}

func TestSerializeJSON_Roundtrip(t *testing.T) {
	// Test that JSON structure is as expected
	comp := NewNGPlusDifficultyComponentForCycle(5)
	comp.ApplyBossEnhancements(5)

	data, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify key fields are present
	if _, ok := m["health_multiplier"]; !ok {
		t.Error("health_multiplier field missing")
	}
	if _, ok := m["ngplus_cycle"]; !ok {
		t.Error("ngplus_cycle field missing")
	}
	if _, ok := m["additional_abilities"]; !ok {
		t.Error("additional_abilities field missing")
	}
}
