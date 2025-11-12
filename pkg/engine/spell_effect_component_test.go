package engine

import (
	"testing"
)

func TestEffectType_String(t *testing.T) {
	tests := []struct {
		name     string
		effect   EffectType
		expected string
	}{
		{"terrain manipulation", EffectTerrainManipulation, "terrain_manipulation"},
		{"transmutation", EffectTransmutation, "transmutation"},
		{"summoning", EffectSummoning, "summoning"},
		{"illusion", EffectIllusion, "illusion"},
		{"time manipulation", EffectTimeManipulation, "time_manipulation"},
		{"gravity control", EffectGravityControl, "gravity_control"},
		{"elemental fusion", EffectElementalFusion, "elemental_fusion"},
		{"life drain", EffectLifeDrain, "life_drain"},
		{"teleportation", EffectTeleportation, "teleportation"},
		{"metamagic", EffectMetamagic, "metamagic"},
		{"unknown", EffectType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.effect.String()
			if got != tt.expected {
				t.Errorf("EffectType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTargetType_String(t *testing.T) {
	tests := []struct {
		name     string
		target   TargetType
		expected string
	}{
		{"self", TargetSelf, "self"},
		{"entity", TargetEntity, "entity"},
		{"area", TargetArea, "area"},
		{"terrain", TargetTerrain, "terrain"},
		{"unknown", TargetType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.target.String()
			if got != tt.expected {
				t.Errorf("TargetType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSpellEffectComponent_Type(t *testing.T) {
	effect := &SpellEffectComponent{}
	if effect.Type() != "spell_effect" {
		t.Errorf("SpellEffectComponent.Type() = %v, want spell_effect", effect.Type())
	}
}

func TestSpellEffectComponent_IsExpired(t *testing.T) {
	tests := []struct {
		name        string
		duration    float64
		elapsedTime float64
		expected    bool
	}{
		{"instant effect", 0, 0, true},
		{"not expired", 5.0, 2.0, false},
		{"expired", 5.0, 5.0, true},
		{"expired past duration", 5.0, 6.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := &SpellEffectComponent{
				Duration:    tt.duration,
				ElapsedTime: tt.elapsedTime,
			}
			got := effect.IsExpired()
			if got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSpellEffectComponent_Update(t *testing.T) {
	effect := &SpellEffectComponent{
		Duration:    5.0,
		ElapsedTime: 0.0,
	}

	// Update with 1 second
	effect.Update(1.0)
	if effect.ElapsedTime != 1.0 {
		t.Errorf("After 1s update, ElapsedTime = %v, want 1.0", effect.ElapsedTime)
	}

	// Update with another 2 seconds
	effect.Update(2.0)
	if effect.ElapsedTime != 3.0 {
		t.Errorf("After 2s update, ElapsedTime = %v, want 3.0", effect.ElapsedTime)
	}
}

func TestSpellEffectComponent_GetProgress(t *testing.T) {
	tests := []struct {
		name        string
		duration    float64
		elapsedTime float64
		expected    float64
	}{
		{"instant effect", 0, 0, 1.0},
		{"25% complete", 4.0, 1.0, 0.25},
		{"50% complete", 4.0, 2.0, 0.5},
		{"75% complete", 4.0, 3.0, 0.75},
		{"100% complete", 4.0, 4.0, 1.0},
		{"past completion", 4.0, 5.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := &SpellEffectComponent{
				Duration:    tt.duration,
				ElapsedTime: tt.elapsedTime,
			}
			got := effect.GetProgress()
			if got != tt.expected {
				t.Errorf("GetProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSpellEffectComponent_TerrainManipulation(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType:      EffectTerrainManipulation,
		Duration:        0, // Instant
		Magnitude:       10.0,
		TargetType:      TargetTerrain,
		TerrainModifier: 5,
		CasterID:        100,
		TargetX:         50.0,
		TargetY:         75.0,
		Radius:          3.0,
		Active:          true,
	}

	if effect.Type() != "spell_effect" {
		t.Errorf("Type() = %v, want spell_effect", effect.Type())
	}
	if effect.EffectType != EffectTerrainManipulation {
		t.Errorf("EffectType = %v, want %v", effect.EffectType, EffectTerrainManipulation)
	}
	if effect.IsExpired() != true {
		t.Errorf("IsExpired() = %v, want true for instant effect", effect.IsExpired())
	}
}

func TestSpellEffectComponent_Summoning(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType:     EffectSummoning,
		Duration:       30.0, // 30 second summon
		Magnitude:      1.0,
		TargetType:     TargetArea,
		SummonTemplate: "fire_elemental",
		CasterID:       100,
		TargetX:        50.0,
		TargetY:        75.0,
		Active:         true,
	}

	if effect.IsExpired() != false {
		t.Errorf("IsExpired() = %v, want false for new summon", effect.IsExpired())
	}

	// Simulate 15 seconds elapsed
	effect.Update(15.0)
	progress := effect.GetProgress()
	if progress != 0.5 {
		t.Errorf("GetProgress() = %v, want 0.5 at 50%% duration", progress)
	}

	// Complete the duration
	effect.Update(15.0)
	if effect.IsExpired() != true {
		t.Errorf("IsExpired() = %v, want true after full duration", effect.IsExpired())
	}
}

func TestSpellEffectComponent_Metamagic(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType:          EffectMetamagic,
		Duration:            0, // Instant modifier
		Magnitude:           2.0,
		TargetType:          TargetEntity,
		MetamagicMultiplier: 2.5, // 2.5x damage multiplier
		CasterID:            100,
		TargetID:            200,
		Active:              true,
	}

	if effect.MetamagicMultiplier != 2.5 {
		t.Errorf("MetamagicMultiplier = %v, want 2.5", effect.MetamagicMultiplier)
	}
	if effect.EffectType.String() != "metamagic" {
		t.Errorf("EffectType.String() = %v, want metamagic", effect.EffectType.String())
	}
}

func TestSpellEffectComponent_ElementalFusion(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType:     EffectElementalFusion,
		Duration:       0, // Instant fusion
		Magnitude:      50.0,
		TargetType:     TargetArea,
		FusionElements: "fire,ice",
		CasterID:       100,
		TargetX:        100.0,
		TargetY:        150.0,
		Radius:         5.0,
		Active:         true,
	}

	if effect.FusionElements != "fire,ice" {
		t.Errorf("FusionElements = %v, want fire,ice", effect.FusionElements)
	}
	if effect.EffectType != EffectElementalFusion {
		t.Errorf("EffectType = %v, want %v", effect.EffectType, EffectElementalFusion)
	}
}

func TestSpellEffectComponent_LifeDrain(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType: EffectLifeDrain,
		Duration:   3.0, // 3 second drain
		Magnitude:  10.0, // 10 HP per second
		TargetType: TargetEntity,
		CasterID:   100,
		TargetID:   200,
		Active:     true,
	}

	// Test progression over time
	for i := 0; i < 3; i++ {
		if effect.IsExpired() {
			t.Errorf("IsExpired() = true at iteration %d, want false", i)
		}
		effect.Update(1.0)
	}

	if !effect.IsExpired() {
		t.Errorf("IsExpired() = false after 3 seconds, want true")
	}
}

func TestSpellEffectComponent_Teleportation(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType: EffectTeleportation,
		Duration:   0, // Instant teleport
		Magnitude:  1.0,
		TargetType: TargetSelf,
		CasterID:   100,
		TargetX:    200.0,
		TargetY:    300.0,
		Active:     true,
	}

	if effect.TargetType != TargetSelf {
		t.Errorf("TargetType = %v, want %v", effect.TargetType, TargetSelf)
	}
	if effect.EffectType.String() != "teleportation" {
		t.Errorf("EffectType.String() = %v, want teleportation", effect.EffectType.String())
	}
}

func TestSpellEffectComponent_TimeManipulation(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType: EffectTimeManipulation,
		Duration:   5.0, // 5 second slow effect
		Magnitude:  0.5, // 50% speed reduction
		TargetType: TargetArea,
		CasterID:   100,
		TargetX:    150.0,
		TargetY:    150.0,
		Radius:     10.0,
		Active:     true,
	}

	// Verify setup
	if effect.Duration != 5.0 {
		t.Errorf("Duration = %v, want 5.0", effect.Duration)
	}
	if effect.Magnitude != 0.5 {
		t.Errorf("Magnitude = %v, want 0.5", effect.Magnitude)
	}

	// Test progress
	effect.Update(2.5)
	progress := effect.GetProgress()
	if progress != 0.5 {
		t.Errorf("GetProgress() = %v, want 0.5 at 50%% duration", progress)
	}
}

func TestSpellEffectComponent_GravityControl(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType: EffectGravityControl,
		Duration:   10.0, // 10 second levitation
		Magnitude:  2.0,  // 2x normal gravity (can be negative for levitation)
		TargetType: TargetEntity,
		CasterID:   100,
		TargetID:   100, // Self-cast levitation
		Active:     true,
	}

	// Test multiple update cycles
	for i := 0; i < 5; i++ {
		effect.Update(1.0)
	}

	if effect.ElapsedTime != 5.0 {
		t.Errorf("ElapsedTime = %v, want 5.0", effect.ElapsedTime)
	}
	if effect.GetProgress() != 0.5 {
		t.Errorf("GetProgress() = %v, want 0.5", effect.GetProgress())
	}
}

func TestSpellEffectComponent_Illusion(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType: EffectIllusion,
		Duration:   15.0, // 15 second invisibility
		Magnitude:  1.0,  // Full invisibility
		TargetType: TargetSelf,
		CasterID:   100,
		TargetID:   100,
		Active:     true,
	}

	// Verify illusion properties
	if effect.EffectType != EffectIllusion {
		t.Errorf("EffectType = %v, want %v", effect.EffectType, EffectIllusion)
	}
	if effect.TargetType != TargetSelf {
		t.Errorf("TargetType = %v, want %v", effect.TargetType, TargetSelf)
	}

	// Test duration
	effect.Update(15.0)
	if !effect.IsExpired() {
		t.Errorf("IsExpired() = false after full duration, want true")
	}
}

func TestSpellEffectComponent_Transmutation(t *testing.T) {
	effect := &SpellEffectComponent{
		EffectType:      EffectTransmutation,
		Duration:        0,   // Instant
		Magnitude:       1.0, // Full conversion
		TargetType:      TargetTerrain,
		TerrainModifier: 7, // Convert to new terrain type
		CasterID:        100,
		TargetX:         50.0,
		TargetY:         50.0,
		Radius:          2.0, // 2 tile radius
		Active:          true,
	}

	// Verify transmutation setup
	if effect.TerrainModifier != 7 {
		t.Errorf("TerrainModifier = %v, want 7", effect.TerrainModifier)
	}
	if effect.IsExpired() != true {
		t.Errorf("IsExpired() = false, want true for instant effect")
	}
}
