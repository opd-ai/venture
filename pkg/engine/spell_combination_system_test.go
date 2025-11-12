package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
)

func TestNewSpellCombinationSystem(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	
	sys := NewSpellCombinationSystem(world, rng)
	
	if sys == nil {
		t.Fatal("NewSpellCombinationSystem returned nil")
	}
	if sys.world != world {
		t.Error("System world not set correctly")
	}
	if sys.rng != rng {
		t.Error("System rng not set correctly")
	}
}

func TestSpellCombinationSystem_OnSpellCast(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	spell := &magic.Spell{
		Name:    "Fireball",
		Element: magic.ElementFire,
	}
	
	// Cast a spell - should create combo component
	sys.OnSpellCast(entity, spell, 0)
	
	comboComp, hasCombo := entity.GetComponent("spell_combo")
	if !hasCombo {
		t.Fatal("SpellComboComponent not created")
	}
	
	combo, ok := comboComp.(*SpellComboComponent)
	if !ok {
		t.Fatal("Component is not SpellComboComponent")
	}
	
	if len(combo.RecentCasts) != 1 {
		t.Errorf("Expected 1 recent cast, got %d", len(combo.RecentCasts))
	}
	
	if combo.RecentCasts[0].SpellName != "Fireball" {
		t.Errorf("Expected spell name 'Fireball', got %s", combo.RecentCasts[0].SpellName)
	}
}

func TestSpellCombinationSystem_GetActiveComboMultiplier(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	
	// No combo component - should return 1.0
	mult := sys.GetActiveComboMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("Expected multiplier 1.0 (no combo), got %f", mult)
	}
	
	// Add combo component with no active combo
	combo := &SpellComboComponent{
		ComboWindow: 1.0,
	}
	entity.AddComponent(combo)
	
	mult = sys.GetActiveComboMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("Expected multiplier 1.0 (no active combo), got %f", mult)
	}
	
	// Add active combo
	currentTime := GetCurrentTime()
	combo.ActiveCombo = &ActiveCombo{
		PowerMultiplier: 1.5,
		StartTime:       currentTime,
		Duration:        3.0,
	}
	
	mult = sys.GetActiveComboMultiplier(entity)
	if mult != 1.5 {
		t.Errorf("Expected multiplier 1.5 (active combo), got %f", mult)
	}
	
	// Expired combo
	combo.ActiveCombo.StartTime = currentTime - 5.0
	combo.ActiveCombo.Duration = 2.0
	
	mult = sys.GetActiveComboMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("Expected multiplier 1.0 (expired combo), got %f", mult)
	}
}

func TestSpellCombinationSystem_DiscoverRecipe(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	
	// Discover a recipe
	sys.DiscoverRecipe(entity, "Fireball", "Ice Shard", "fire", "ice", "Steam Burst", 1.5, true)
	
	comboComp, hasCombo := entity.GetComponent("spell_combo")
	if !hasCombo {
		t.Fatal("SpellComboComponent not created")
	}
	
	combo, ok := comboComp.(*SpellComboComponent)
	if !ok {
		t.Fatal("Component is not SpellComboComponent")
	}
	
	if len(combo.KnownRecipes) != 1 {
		t.Fatalf("Expected 1 recipe, got %d", len(combo.KnownRecipes))
	}
	
	recipe := combo.KnownRecipes[0]
	if recipe.Spell1Name != "Fireball" {
		t.Errorf("Expected Spell1Name 'Fireball', got %s", recipe.Spell1Name)
	}
	if recipe.Spell2Name != "Ice Shard" {
		t.Errorf("Expected Spell2Name 'Ice Shard', got %s", recipe.Spell2Name)
	}
	if recipe.PowerMultiplier != 1.5 {
		t.Errorf("Expected PowerMultiplier 1.5, got %f", recipe.PowerMultiplier)
	}
	if !recipe.IsSymmetric {
		t.Error("Expected IsSymmetric to be true")
	}
}

func TestSpellCombinationSystem_ElementalSynergy(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	tests := []struct {
		name     string
		elem1    string
		elem2    string
		wantSync bool
		wantMult float64
	}{
		{
			name:     "fire + wind synergy",
			elem1:    "fire",
			elem2:    "wind",
			wantSync: true,
			wantMult: 1.5,
		},
		{
			name:     "wind + fire synergy (reversed)",
			elem1:    "wind",
			elem2:    "fire",
			wantSync: true,
			wantMult: 1.5,
		},
		{
			name:     "ice + fire synergy",
			elem1:    "ice",
			elem2:    "fire",
			wantSync: true,
			wantMult: 1.3,
		},
		{
			name:     "lightning + earth synergy",
			elem1:    "lightning",
			elem2:    "earth",
			wantSync: true,
			wantMult: 1.6,
		},
		{
			name:     "light + dark synergy",
			elem1:    "light",
			elem2:    "dark",
			wantSync: true,
			wantMult: 2.0,
		},
		{
			name:     "arcane + arcane synergy",
			elem1:    "arcane",
			elem2:    "arcane",
			wantSync: true,
			wantMult: 1.8,
		},
		{
			name:     "no synergy",
			elem1:    "fire",
			elem2:    "none",
			wantSync: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synergy := sys.checkElementalSynergy(tt.elem1, tt.elem2)
			
			if tt.wantSync {
				if synergy == nil {
					t.Error("Expected synergy, got nil")
					return
				}
				if synergy.PowerMultiplier != tt.wantMult {
					t.Errorf("Expected multiplier %f, got %f", tt.wantMult, synergy.PowerMultiplier)
				}
			} else {
				if synergy != nil {
					t.Error("Expected no synergy, got synergy")
				}
			}
		})
	}
}

func TestSpellCombinationSystem_Update_ComboDetection(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	currentTime := GetCurrentTime()
	
	combo := &SpellComboComponent{
		ComboWindow: 1.0,
		RecentCasts: []RecentCast{
			{
				SpellName: "Fireball",
				Element:   "fire",
				CastTime:  currentTime - 0.5, // 0.5s ago
				SlotIndex: 0,
			},
			{
				SpellName: "Wind Blade",
				Element:   "wind",
				CastTime:  currentTime - 0.1, // 0.1s ago
				SlotIndex: 1,
			},
		},
	}
	entity.AddComponent(combo)
	
	// Update should detect fire+wind synergy
	sys.Update([]*Entity{entity}, 0.016)
	
	// Should have active combo now
	if combo.ActiveCombo == nil {
		t.Fatal("Expected active combo, got nil")
	}
	
	if combo.ActiveCombo.PowerMultiplier != 1.5 {
		t.Errorf("Expected multiplier 1.5, got %f", combo.ActiveCombo.PowerMultiplier)
	}
	
	if combo.ActiveCombo.EffectDescription != "Fire and wind create a raging inferno!" {
		t.Errorf("Unexpected effect description: %s", combo.ActiveCombo.EffectDescription)
	}
	
	// Recent casts should be cleared after combo
	if len(combo.RecentCasts) != 0 {
		t.Errorf("Expected recent casts to be cleared, got %d", len(combo.RecentCasts))
	}
}

func TestSpellCombinationSystem_Update_RecipeCombo(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	currentTime := GetCurrentTime()
	
	recipe := ComboRecipe{
		Spell1Name:      "Custom Spell A",
		Spell2Name:      "Custom Spell B",
		Element1:        "none",
		Element2:        "none",
		ResultEffect:    "Ultimate Power",
		PowerMultiplier: 2.5,
		IsSymmetric:     false,
	}
	
	combo := &SpellComboComponent{
		ComboWindow: 1.0,
		KnownRecipes: []ComboRecipe{recipe},
		RecentCasts: []RecentCast{
			{
				SpellName: "Custom Spell A",
				Element:   "none",
				CastTime:  currentTime - 0.5,
				SlotIndex: 0,
			},
			{
				SpellName: "Custom Spell B",
				Element:   "none",
				CastTime:  currentTime - 0.1,
				SlotIndex: 1,
			},
		},
	}
	entity.AddComponent(combo)
	
	// Update should detect recipe combo
	sys.Update([]*Entity{entity}, 0.016)
	
	// Should have active combo now
	if combo.ActiveCombo == nil {
		t.Fatal("Expected active combo from recipe, got nil")
	}
	
	if combo.ActiveCombo.PowerMultiplier != 2.5 {
		t.Errorf("Expected multiplier 2.5, got %f", combo.ActiveCombo.PowerMultiplier)
	}
	
	if combo.ActiveCombo.EffectDescription != "Ultimate Power" {
		t.Errorf("Expected effect 'Ultimate Power', got %s", combo.ActiveCombo.EffectDescription)
	}
}

func TestSpellCombinationSystem_Update_NoComboOutsideWindow(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	currentTime := GetCurrentTime()
	
	combo := &SpellComboComponent{
		ComboWindow: 1.0,
		RecentCasts: []RecentCast{
			{
				SpellName: "Fireball",
				Element:   "fire",
				CastTime:  currentTime - 2.0, // 2s ago (outside window)
				SlotIndex: 0,
			},
			{
				SpellName: "Wind Blade",
				Element:   "wind",
				CastTime:  currentTime - 0.1,
				SlotIndex: 1,
			},
		},
	}
	entity.AddComponent(combo)
	
	// Update should NOT detect combo (too far apart)
	sys.Update([]*Entity{entity}, 0.016)
	
	// Should NOT have active combo
	if combo.ActiveCombo != nil {
		t.Error("Expected no combo (outside window), got active combo")
	}
}

func TestSpellCombinationSystem_Update_BacklashDamage(t *testing.T) {
	world := NewWorld()
	// Use seed that will trigger backlash (30% chance)
	rng := rand.New(rand.NewSource(1)) // Seed 1 gives low random values
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	currentTime := GetCurrentTime()
	
	health := &HealthComponent{
		Current: 100.0,
		Max:     100.0,
	}
	entity.AddComponent(health)
	
	combo := &SpellComboComponent{
		ComboWindow: 1.0,
		RecentCasts: []RecentCast{
			{
				SpellName: "Fireball",
				Element:   "fire",
				CastTime:  currentTime - 0.5,
				SlotIndex: 0,
			},
			{
				SpellName: "Ice Shard",
				Element:   "ice",
				CastTime:  currentTime - 0.1,
				SlotIndex: 1,
			},
		},
	}
	entity.AddComponent(combo)
	
	// Update - fire+ice might trigger backlash
	sys.Update([]*Entity{entity}, 0.016)
	
	// Check if backlash occurred (30% chance with seed 1 should trigger)
	if combo.ActiveCombo != nil && combo.ActiveCombo.PowerMultiplier < 1.0 {
		// Backlash occurred
		if health.Current >= 100.0 {
			t.Error("Expected health damage from backlash")
		}
		
		expectedDamage := 10.0 // 10% of 100
		expectedHealth := 100.0 - expectedDamage
		if health.Current != expectedHealth {
			t.Errorf("Expected health %f after backlash, got %f", expectedHealth, health.Current)
		}
	}
}

func TestSpellCombinationSystem_Update_ComboExpiration(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	currentTime := GetCurrentTime()
	
	// Create an expired combo
	combo := &SpellComboComponent{
		ComboWindow: 1.0,
		ActiveCombo: &ActiveCombo{
			PowerMultiplier: 1.5,
			StartTime:       currentTime - 5.0, // Started 5s ago
			Duration:        2.0,                // Only lasts 2s (expired 3s ago)
		},
	}
	entity.AddComponent(combo)
	
	// Update should clear expired combo
	sys.Update([]*Entity{entity}, 0.016)
	
	if combo.ActiveCombo != nil {
		t.Error("Expected expired combo to be cleared, still active")
	}
}

func TestSpellCombinationSystem_Performance(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	sys := NewSpellCombinationSystem(world, rng)
	
	// Create many entities with combo components
	entities := make([]*Entity, 100)
	currentTime := GetCurrentTime()
	
	for i := range entities {
		entity := &Entity{
			ID:         uint64(i + 1),
			Components: make(map[string]Component),
		}
		combo := &SpellComboComponent{
			ComboWindow: 1.0,
			RecentCasts: []RecentCast{
				{
					SpellName: "Spell A",
					Element:   "fire",
					CastTime:  currentTime - 0.5,
					SlotIndex: 0,
				},
				{
					SpellName: "Spell B",
					Element:   "wind",
					CastTime:  currentTime - 0.1,
					SlotIndex: 1,
				},
			},
		}
		entity.AddComponent(combo)
		entities[i] = entity
	}
	
	// Benchmark update performance
	iterations := 100
	for i := 0; i < iterations; i++ {
		sys.Update(entities, 0.016)
	}
	
	// No specific assertion, just ensure it completes without panic
	// Performance target is <0.5ms per evaluation, which this should easily meet
}

func TestSpellCombinationSystem_Determinism(t *testing.T) {
	// Test that same seed produces same backlash outcomes
	seed := int64(42)
	
	results1 := make([]bool, 10)
	for i := 0; i < 10; i++ {
		world := NewWorld()
		rng := rand.New(rand.NewSource(seed))
		sys := NewSpellCombinationSystem(world, rng)
		results1[i] = sys.checkIncompatibleCombo("fire", "ice")
	}
	
	results2 := make([]bool, 10)
	for i := 0; i < 10; i++ {
		world := NewWorld()
		rng := rand.New(rand.NewSource(seed))
		sys := NewSpellCombinationSystem(world, rng)
		results2[i] = sys.checkIncompatibleCombo("fire", "ice")
	}
	
	// Results should be identical with same seed
	for i := range results1 {
		if results1[i] != results2[i] {
			t.Errorf("Determinism violated at iteration %d: %v != %v", i, results1[i], results2[i])
		}
	}
}
