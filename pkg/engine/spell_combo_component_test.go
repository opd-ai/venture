package engine

import (
	"testing"
)

func TestSpellComboComponent_Type(t *testing.T) {
	comp := &SpellComboComponent{}
	if got := comp.Type(); got != "spell_combo" {
		t.Errorf("Type() = %v, want %v", got, "spell_combo")
	}
}

func TestSpellComboComponent_AddRecentCast(t *testing.T) {
	comp := &SpellComboComponent{ComboWindow: 1.0}
	
	comp.AddRecentCast("Fireball", "fire", 100.0, 0)
	
	if len(comp.RecentCasts) != 1 {
		t.Errorf("Expected 1 recent cast, got %d", len(comp.RecentCasts))
	}
	
	cast := comp.RecentCasts[0]
	if cast.SpellName != "Fireball" {
		t.Errorf("Expected spell name 'Fireball', got %s", cast.SpellName)
	}
	if cast.Element != "fire" {
		t.Errorf("Expected element 'fire', got %s", cast.Element)
	}
	if cast.CastTime != 100.0 {
		t.Errorf("Expected cast time 100.0, got %f", cast.CastTime)
	}
	if cast.SlotIndex != 0 {
		t.Errorf("Expected slot index 0, got %d", cast.SlotIndex)
	}
}

func TestSpellComboComponent_CleanOldCasts(t *testing.T) {
	tests := []struct {
		name        string
		comboWindow float64
		casts       []RecentCast
		currentTime float64
		wantCount   int
	}{
		{
			name:        "all casts within window",
			comboWindow: 1.0,
			casts: []RecentCast{
				{SpellName: "Fireball", CastTime: 99.0},
				{SpellName: "Ice Shard", CastTime: 99.5},
			},
			currentTime: 100.0,
			wantCount:   2,
		},
		{
			name:        "one cast expired",
			comboWindow: 1.0,
			casts: []RecentCast{
				{SpellName: "Fireball", CastTime: 98.0},
				{SpellName: "Ice Shard", CastTime: 99.5},
			},
			currentTime: 100.0,
			wantCount:   1,
		},
		{
			name:        "all casts expired",
			comboWindow: 1.0,
			casts: []RecentCast{
				{SpellName: "Fireball", CastTime: 97.0},
				{SpellName: "Ice Shard", CastTime: 98.0},
			},
			currentTime: 100.0,
			wantCount:   0,
		},
		{
			name:        "empty cast list",
			comboWindow: 1.0,
			casts:       []RecentCast{},
			currentTime: 100.0,
			wantCount:   0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &SpellComboComponent{
				ComboWindow: tt.comboWindow,
				RecentCasts: tt.casts,
			}
			
			comp.CleanOldCasts(tt.currentTime)
			
			if len(comp.RecentCasts) != tt.wantCount {
				t.Errorf("CleanOldCasts() resulted in %d casts, want %d", len(comp.RecentCasts), tt.wantCount)
			}
		})
	}
}

func TestSpellComboComponent_HasRecentCasts(t *testing.T) {
	tests := []struct {
		name        string
		comboWindow float64
		casts       []RecentCast
		currentTime float64
		want        bool
	}{
		{
			name:        "has recent casts",
			comboWindow: 1.0,
			casts: []RecentCast{
				{SpellName: "Fireball", CastTime: 99.5},
			},
			currentTime: 100.0,
			want:        true,
		},
		{
			name:        "no recent casts",
			comboWindow: 1.0,
			casts: []RecentCast{
				{SpellName: "Fireball", CastTime: 98.0},
			},
			currentTime: 100.0,
			want:        false,
		},
		{
			name:        "empty cast list",
			comboWindow: 1.0,
			casts:       []RecentCast{},
			currentTime: 100.0,
			want:        false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &SpellComboComponent{
				ComboWindow: tt.comboWindow,
				RecentCasts: tt.casts,
			}
			
			if got := comp.HasRecentCasts(tt.currentTime); got != tt.want {
				t.Errorf("HasRecentCasts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpellComboComponent_GetRecentCastsCount(t *testing.T) {
	comp := &SpellComboComponent{
		ComboWindow: 1.0,
		RecentCasts: []RecentCast{
			{SpellName: "Fireball", CastTime: 98.0},
			{SpellName: "Ice Shard", CastTime: 99.5},
			{SpellName: "Lightning", CastTime: 99.8},
		},
	}
	
	// At time 100.0, only casts at 99.0+ are valid (within 1s window)
	count := comp.GetRecentCastsCount(100.0)
	if count != 2 {
		t.Errorf("GetRecentCastsCount() = %d, want 2", count)
	}
}

func TestSpellComboComponent_HasRecipe(t *testing.T) {
	tests := []struct {
		name    string
		recipes []ComboRecipe
		spell1  string
		spell2  string
		want    bool
	}{
		{
			name: "exact match",
			recipes: []ComboRecipe{
				{Spell1Name: "Fireball", Spell2Name: "Ice Shard", IsSymmetric: false},
			},
			spell1: "Fireball",
			spell2: "Ice Shard",
			want:   true,
		},
		{
			name: "symmetric match forward",
			recipes: []ComboRecipe{
				{Spell1Name: "Fireball", Spell2Name: "Ice Shard", IsSymmetric: true},
			},
			spell1: "Fireball",
			spell2: "Ice Shard",
			want:   true,
		},
		{
			name: "symmetric match reverse",
			recipes: []ComboRecipe{
				{Spell1Name: "Fireball", Spell2Name: "Ice Shard", IsSymmetric: true},
			},
			spell1: "Ice Shard",
			spell2: "Fireball",
			want:   true,
		},
		{
			name: "asymmetric no reverse match",
			recipes: []ComboRecipe{
				{Spell1Name: "Fireball", Spell2Name: "Ice Shard", IsSymmetric: false},
			},
			spell1: "Ice Shard",
			spell2: "Fireball",
			want:   false,
		},
		{
			name:    "no recipes",
			recipes: []ComboRecipe{},
			spell1:  "Fireball",
			spell2:  "Ice Shard",
			want:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &SpellComboComponent{
				KnownRecipes: tt.recipes,
			}
			
			if got := comp.HasRecipe(tt.spell1, tt.spell2); got != tt.want {
				t.Errorf("HasRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpellComboComponent_AddRecipe(t *testing.T) {
	comp := &SpellComboComponent{
		KnownRecipes: []ComboRecipe{},
	}
	
	// Add first recipe
	recipe1 := ComboRecipe{
		Spell1Name:      "Fireball",
		Spell2Name:      "Ice Shard",
		PowerMultiplier: 1.5,
		IsSymmetric:     true,
	}
	comp.AddRecipe(recipe1)
	
	if len(comp.KnownRecipes) != 1 {
		t.Errorf("Expected 1 recipe after first add, got %d", len(comp.KnownRecipes))
	}
	
	// Try to add duplicate (should be ignored)
	comp.AddRecipe(recipe1)
	
	if len(comp.KnownRecipes) != 1 {
		t.Errorf("Expected 1 recipe after duplicate add, got %d", len(comp.KnownRecipes))
	}
	
	// Add second different recipe
	recipe2 := ComboRecipe{
		Spell1Name:      "Lightning",
		Spell2Name:      "Earth",
		PowerMultiplier: 2.0,
		IsSymmetric:     false,
	}
	comp.AddRecipe(recipe2)
	
	if len(comp.KnownRecipes) != 2 {
		t.Errorf("Expected 2 recipes after second add, got %d", len(comp.KnownRecipes))
	}
}

func TestSpellComboComponent_GetRecipe(t *testing.T) {
	recipe1 := ComboRecipe{
		Spell1Name:      "Fireball",
		Spell2Name:      "Ice Shard",
		PowerMultiplier: 1.5,
		IsSymmetric:     true,
	}
	recipe2 := ComboRecipe{
		Spell1Name:      "Lightning",
		Spell2Name:      "Earth",
		PowerMultiplier: 2.0,
		IsSymmetric:     false,
	}
	
	comp := &SpellComboComponent{
		KnownRecipes: []ComboRecipe{recipe1, recipe2},
	}
	
	// Test symmetric match
	got := comp.GetRecipe("Fireball", "Ice Shard")
	if got == nil {
		t.Error("Expected recipe, got nil")
	} else if got.PowerMultiplier != 1.5 {
		t.Errorf("Expected multiplier 1.5, got %f", got.PowerMultiplier)
	}
	
	// Test symmetric reverse match
	got = comp.GetRecipe("Ice Shard", "Fireball")
	if got == nil {
		t.Error("Expected recipe (symmetric), got nil")
	}
	
	// Test asymmetric match
	got = comp.GetRecipe("Lightning", "Earth")
	if got == nil {
		t.Error("Expected recipe, got nil")
	} else if got.PowerMultiplier != 2.0 {
		t.Errorf("Expected multiplier 2.0, got %f", got.PowerMultiplier)
	}
	
	// Test asymmetric no reverse
	got = comp.GetRecipe("Earth", "Lightning")
	if got != nil {
		t.Error("Expected nil (asymmetric reverse), got recipe")
	}
	
	// Test non-existent
	got = comp.GetRecipe("Wind", "Water")
	if got != nil {
		t.Error("Expected nil (non-existent), got recipe")
	}
}

func TestSpellComboComponent_IsComboActive(t *testing.T) {
	tests := []struct {
		name        string
		activeCombo *ActiveCombo
		currentTime float64
		want        bool
	}{
		{
			name:        "no active combo",
			activeCombo: nil,
			currentTime: 100.0,
			want:        false,
		},
		{
			name: "combo is active",
			activeCombo: &ActiveCombo{
				StartTime: 99.0,
				Duration:  2.0,
			},
			currentTime: 100.0,
			want:        true,
		},
		{
			name: "combo expired",
			activeCombo: &ActiveCombo{
				StartTime: 97.0,
				Duration:  2.0,
			},
			currentTime: 100.0,
			want:        false,
		},
		{
			name: "combo at exact end time",
			activeCombo: &ActiveCombo{
				StartTime: 98.0,
				Duration:  2.0,
			},
			currentTime: 100.0,
			want:        false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &SpellComboComponent{
				ActiveCombo: tt.activeCombo,
			}
			
			if got := comp.IsComboActive(tt.currentTime); got != tt.want {
				t.Errorf("IsComboActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCurrentTime(t *testing.T) {
	// Just verify it returns a reasonable timestamp
	time1 := GetCurrentTime()
	time2 := GetCurrentTime()
	
	if time1 <= 0 {
		t.Error("GetCurrentTime() returned non-positive value")
	}
	
	if time2 < time1 {
		t.Error("GetCurrentTime() should be monotonically increasing")
	}
	
	// Should be within a reasonable range (after year 2000, before year 2100)
	if time1 < 946684800 || time1 > 4102444800 {
		t.Errorf("GetCurrentTime() returned unreasonable value: %f", time1)
	}
}
