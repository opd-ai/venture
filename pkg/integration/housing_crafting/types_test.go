package housing_crafting

import (
	"testing"
)

// TestStationType tests the StationType enum
func TestStationType(t *testing.T) {
	tests := []struct {
		name     string
		st       StationType
		expected string
	}{
		{"Forge", StationTypeForge, "Forge"},
		{"Alchemy", StationTypeAlchemy, "Alchemy"},
		{"Enchanting", StationTypeEnchanting, "Enchanting"},
		{"Cooking", StationTypeCooking, "Cooking"},
		{"Tailoring", StationTypeTailoring, "Tailoring"},
		{"Woodworking", StationTypeWoodworking, "Woodworking"},
		{"Inscription", StationTypeInscription, "Inscription"},
		{"Engineering", StationTypeEngineering, "Engineering"},
		{"Unknown", StationType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.String(); got != tt.expected {
				t.Errorf("StationType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestQualityTier tests the QualityTier enum and multipliers
func TestQualityTier(t *testing.T) {
	tests := []struct {
		name         string
		qt           QualityTier
		expectedName string
		expectedMult float64
	}{
		{"Basic", QualityBasic, "Basic", 1.0},
		{"Standard", QualityStandard, "Standard", 1.2},
		{"Advanced", QualityAdvanced, "Advanced", 1.5},
		{"Master", QualityMaster, "Master", 2.0},
		{"Unknown", QualityTier(999), "Unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qt.String(); got != tt.expectedName {
				t.Errorf("QualityTier.String() = %v, want %v", got, tt.expectedName)
			}
			if got := tt.qt.Multiplier(); got != tt.expectedMult {
				t.Errorf("QualityTier.Multiplier() = %v, want %v", got, tt.expectedMult)
			}
		})
	}
}

// TestCraftingStation_GetSkillTrainingBonus tests skill bonus retrieval
func TestCraftingStation_GetSkillTrainingBonus(t *testing.T) {
	station := &CraftingStation{
		SkillBonus: map[string]int{
			"smithing": 50,
			"alchemy":  30,
		},
	}

	tests := []struct {
		name      string
		skillName string
		expected  int
	}{
		{"Smithing", "smithing", 50},
		{"Alchemy", "alchemy", 30},
		{"NonExistent", "cooking", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := station.GetSkillTrainingBonus(tt.skillName); got != tt.expected {
				t.Errorf("GetSkillTrainingBonus(%s) = %v, want %v", tt.skillName, got, tt.expected)
			}
		})
	}
}

// TestCraftingStation_HasRecipe tests recipe availability
func TestCraftingStation_HasRecipe(t *testing.T) {
	station := &CraftingStation{
		ActiveRecipes: []string{"recipe1", "recipe2", "recipe3"},
	}

	tests := []struct {
		name     string
		recipeID string
		expected bool
	}{
		{"Recipe1", "recipe1", true},
		{"Recipe2", "recipe2", true},
		{"Recipe3", "recipe3", true},
		{"NonExistent", "recipe4", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := station.HasRecipe(tt.recipeID); got != tt.expected {
				t.Errorf("HasRecipe(%s) = %v, want %v", tt.recipeID, got, tt.expected)
			}
		})
	}
}

// TestCraftingStation_NilSkillBonus tests nil skill bonus map
func TestCraftingStation_NilSkillBonus(t *testing.T) {
	station := &CraftingStation{
		SkillBonus: nil,
	}

	if got := station.GetSkillTrainingBonus("smithing"); got != 0 {
		t.Errorf("GetSkillTrainingBonus with nil map = %v, want 0", got)
	}
}

// TestSkillTrainingFacility_CanTrainSkill tests skill training availability
func TestSkillTrainingFacility_CanTrainSkill(t *testing.T) {
	facility := &SkillTrainingFacility{
		TrainableSkills: []string{"smithing", "alchemy", "enchanting"},
	}

	tests := []struct {
		name      string
		skillName string
		expected  bool
	}{
		{"Smithing", "smithing", true},
		{"Alchemy", "alchemy", true},
		{"Enchanting", "enchanting", true},
		{"Cooking", "cooking", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := facility.CanTrainSkill(tt.skillName); got != tt.expected {
				t.Errorf("CanTrainSkill(%s) = %v, want %v", tt.skillName, got, tt.expected)
			}
		})
	}
}

// TestSkillTrainingFacility_GetXPBonus tests XP bonus calculation
func TestSkillTrainingFacility_GetXPBonus(t *testing.T) {
	tests := []struct {
		name       string
		multiplier float64
		baseXP     float64
		expectedXP float64
	}{
		{"1.5x multiplier", 1.5, 100.0, 150.0},
		{"2.0x multiplier", 2.0, 100.0, 200.0},
		{"1.0x multiplier", 1.0, 100.0, 100.0},
		{"Zero multiplier", 0.0, 100.0, 100.0},
		{"Negative multiplier", -1.0, 100.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facility := &SkillTrainingFacility{
				XPMultiplier: tt.multiplier,
			}
			if got := facility.GetXPBonus(tt.baseXP); got != tt.expectedXP {
				t.Errorf("GetXPBonus(%v) = %v, want %v", tt.baseXP, got, tt.expectedXP)
			}
		})
	}
}

// Benchmark tests

func BenchmarkQualityTier_Multiplier(b *testing.B) {
	qt := QualityMaster
	for i := 0; i < b.N; i++ {
		_ = qt.Multiplier()
	}
}

func BenchmarkCraftingStation_HasRecipe(b *testing.B) {
	station := &CraftingStation{
		ActiveRecipes: []string{"recipe1", "recipe2", "recipe3", "recipe4", "recipe5"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = station.HasRecipe("recipe3")
	}
}

func BenchmarkSkillTrainingFacility_GetXPBonus(b *testing.B) {
	facility := &SkillTrainingFacility{
		XPMultiplier: 1.5,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = facility.GetXPBonus(100.0)
	}
}
