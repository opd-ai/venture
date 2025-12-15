package engine

import (
	"testing"
)

func TestAchievementCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category AchievementCategory
		want     string
	}{
		{"Combat", AchievementCategoryCombat, "Combat"},
		{"Quest", AchievementCategoryQuest, "Quest"},
		{"Crafting", AchievementCategoryCrafting, "Crafting"},
		{"Exploration", AchievementCategoryExploration, "Exploration"},
		{"Social", AchievementCategorySocial, "Social"},
		{"PvP", AchievementCategoryPvP, "PvP"},
		{"Unknown", AchievementCategory(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("AchievementCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAchievementTier_String(t *testing.T) {
	tests := []struct {
		name string
		tier AchievementTier
		want string
	}{
		{"None", AchievementTierNone, "None"},
		{"Bronze", AchievementTierBronze, "Bronze"},
		{"Silver", AchievementTierSilver, "Silver"},
		{"Gold", AchievementTierGold, "Gold"},
		{"Platinum", AchievementTierPlatinum, "Platinum"},
		{"Unknown", AchievementTier(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("AchievementTier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAchievementTier_Points(t *testing.T) {
	tests := []struct {
		name string
		tier AchievementTier
		want int
	}{
		{"None", AchievementTierNone, 0},
		{"Bronze", AchievementTierBronze, 10},
		{"Silver", AchievementTierSilver, 25},
		{"Gold", AchievementTierGold, 50},
		{"Platinum", AchievementTierPlatinum, 100},
		{"Unknown", AchievementTier(99), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.Points(); got != tt.want {
				t.Errorf("AchievementTier.Points() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewExtendedAchievementComponent(t *testing.T) {
	comp := NewExtendedAchievementComponent()

	if comp == nil {
		t.Fatal("NewExtendedAchievementComponent() returned nil")
	}

	if comp.Type() != "extended_achievement" {
		t.Errorf("Type() = %v, want extended_achievement", comp.Type())
	}

	if comp.Achievements == nil {
		t.Error("Achievements map should be initialized")
	}

	if comp.CategoryPoints == nil {
		t.Error("CategoryPoints map should be initialized")
	}

	if comp.TotalPoints != 0 {
		t.Errorf("TotalPoints = %v, want 0", comp.TotalPoints)
	}
}

func TestExtendedAchievementComponent_GetAchievement(t *testing.T) {
	comp := NewExtendedAchievementComponent()

	// Getting a non-existent achievement should create it
	entry := comp.GetAchievement("test_achievement", AchievementCategoryCombat)

	if entry == nil {
		t.Fatal("GetAchievement() returned nil")
	}

	if entry.ID != "test_achievement" {
		t.Errorf("ID = %v, want test_achievement", entry.ID)
	}

	if entry.Category != AchievementCategoryCombat {
		t.Errorf("Category = %v, want Combat", entry.Category)
	}

	if entry.CurrentTier != AchievementTierNone {
		t.Errorf("CurrentTier = %v, want None", entry.CurrentTier)
	}

	// Getting the same achievement should return the same entry
	entry2 := comp.GetAchievement("test_achievement", AchievementCategoryCombat)

	if entry != entry2 {
		t.Error("GetAchievement should return the same entry for the same ID")
	}
}

func TestExtendedAchievementComponent_SetProgress(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	// Set progress below first threshold
	unlocked := comp.SetProgress("test", AchievementCategoryCombat, 0, thresholds, timestamp)

	if unlocked {
		t.Error("Should not unlock at 0 progress")
	}

	if comp.GetTier("test") != AchievementTierNone {
		t.Errorf("Tier = %v, want None", comp.GetTier("test"))
	}

	// Set progress to bronze threshold
	unlocked = comp.SetProgress("test", AchievementCategoryCombat, 1, thresholds, timestamp)

	if !unlocked {
		t.Error("Should unlock at bronze threshold")
	}

	if comp.GetTier("test") != AchievementTierBronze {
		t.Errorf("Tier = %v, want Bronze", comp.GetTier("test"))
	}

	if comp.TotalPoints != 10 {
		t.Errorf("TotalPoints = %v, want 10", comp.TotalPoints)
	}

	// Set progress to silver threshold
	unlocked = comp.SetProgress("test", AchievementCategoryCombat, 10, thresholds, timestamp)

	if !unlocked {
		t.Error("Should unlock at silver threshold")
	}

	if comp.GetTier("test") != AchievementTierSilver {
		t.Errorf("Tier = %v, want Silver", comp.GetTier("test"))
	}

	if comp.TotalPoints != 35 { // 10 (bronze) + 25 (silver)
		t.Errorf("TotalPoints = %v, want 35", comp.TotalPoints)
	}

	// Skip to platinum
	unlocked = comp.SetProgress("test", AchievementCategoryCombat, 1000, thresholds, timestamp)

	if !unlocked {
		t.Error("Should unlock at platinum threshold")
	}

	if comp.GetTier("test") != AchievementTierPlatinum {
		t.Errorf("Tier = %v, want Platinum", comp.GetTier("test"))
	}

	// 10 (bronze) + 25 (silver) + 50 (gold) + 100 (platinum) = 185
	if comp.TotalPoints != 185 {
		t.Errorf("TotalPoints = %v, want 185", comp.TotalPoints)
	}

	// Setting same progress should not unlock again
	unlocked = comp.SetProgress("test", AchievementCategoryCombat, 1000, thresholds, timestamp)

	if unlocked {
		t.Error("Should not unlock again at same progress")
	}
}

func TestExtendedAchievementComponent_IncrementProgress(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{5, 10, 20, 50}
	timestamp := int64(1000000)

	// Increment from 0
	unlocked := comp.IncrementProgress("test", AchievementCategoryCombat, 3, thresholds, timestamp)

	if unlocked {
		t.Error("Should not unlock at 3")
	}

	if comp.GetProgress("test") != 3 {
		t.Errorf("Progress = %v, want 3", comp.GetProgress("test"))
	}

	// Increment to reach bronze
	unlocked = comp.IncrementProgress("test", AchievementCategoryCombat, 2, thresholds, timestamp)

	if !unlocked {
		t.Error("Should unlock bronze at 5")
	}

	if comp.GetProgress("test") != 5 {
		t.Errorf("Progress = %v, want 5", comp.GetProgress("test"))
	}
}

func TestExtendedAchievementComponent_CategoryPoints(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	// Add combat achievement
	comp.SetProgress("combat1", AchievementCategoryCombat, 1, thresholds, timestamp)

	if comp.GetCategoryPoints(AchievementCategoryCombat) != 10 {
		t.Errorf("Combat points = %v, want 10", comp.GetCategoryPoints(AchievementCategoryCombat))
	}

	// Add quest achievement
	comp.SetProgress("quest1", AchievementCategoryQuest, 1, thresholds, timestamp)

	if comp.GetCategoryPoints(AchievementCategoryQuest) != 10 {
		t.Errorf("Quest points = %v, want 10", comp.GetCategoryPoints(AchievementCategoryQuest))
	}

	// Combat points should be unchanged
	if comp.GetCategoryPoints(AchievementCategoryCombat) != 10 {
		t.Errorf("Combat points = %v, want 10 (unchanged)", comp.GetCategoryPoints(AchievementCategoryCombat))
	}

	// Total should be 20
	if comp.GetTotalPoints() != 20 {
		t.Errorf("TotalPoints = %v, want 20", comp.GetTotalPoints())
	}
}

func TestExtendedAchievementComponent_GetAchievementsByCategory(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	// Add achievements in different categories
	comp.SetProgress("combat1", AchievementCategoryCombat, 1, thresholds, timestamp)
	comp.SetProgress("combat2", AchievementCategoryCombat, 1, thresholds, timestamp)
	comp.SetProgress("quest1", AchievementCategoryQuest, 1, thresholds, timestamp)

	combatAchievements := comp.GetAchievementsByCategory(AchievementCategoryCombat)

	if len(combatAchievements) != 2 {
		t.Errorf("Combat achievements count = %v, want 2", len(combatAchievements))
	}

	questAchievements := comp.GetAchievementsByCategory(AchievementCategoryQuest)

	if len(questAchievements) != 1 {
		t.Errorf("Quest achievements count = %v, want 1", len(questAchievements))
	}
}

func TestExtendedAchievementComponent_UnlockedCount(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	if comp.GetUnlockedCount() != 0 {
		t.Errorf("UnlockedCount = %v, want 0", comp.GetUnlockedCount())
	}

	// Add achievements - only count those with Bronze+
	comp.SetProgress("unlocked", AchievementCategoryCombat, 1, thresholds, timestamp)
	comp.SetProgress("not_unlocked", AchievementCategoryQuest, 0, thresholds, timestamp)

	if comp.GetUnlockedCount() != 1 {
		t.Errorf("UnlockedCount = %v, want 1", comp.GetUnlockedCount())
	}
}

func TestExtendedAchievementComponent_MaxTierCount(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	if comp.GetMaxTierCount() != 0 {
		t.Errorf("MaxTierCount = %v, want 0", comp.GetMaxTierCount())
	}

	// Add a platinum achievement
	comp.SetProgress("platinum", AchievementCategoryCombat, 1000, thresholds, timestamp)

	// Add a gold achievement
	comp.SetProgress("gold", AchievementCategoryQuest, 100, thresholds, timestamp)

	if comp.GetMaxTierCount() != 1 {
		t.Errorf("MaxTierCount = %v, want 1", comp.GetMaxTierCount())
	}
}

func TestExtendedAchievementComponent_Serialization(t *testing.T) {
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	comp.SetProgress("combat1", AchievementCategoryCombat, 50, thresholds, timestamp)
	comp.SetProgress("quest1", AchievementCategoryQuest, 1, thresholds, timestamp)

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	comp2 := NewExtendedAchievementComponent()
	err = comp2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify
	if comp2.GetTotalPoints() != comp.GetTotalPoints() {
		t.Errorf("TotalPoints = %v, want %v", comp2.GetTotalPoints(), comp.GetTotalPoints())
	}

	if comp2.GetProgress("combat1") != 50 {
		t.Errorf("combat1 progress = %v, want 50", comp2.GetProgress("combat1"))
	}

	if comp2.GetTier("combat1") != AchievementTierSilver {
		t.Errorf("combat1 tier = %v, want Silver", comp2.GetTier("combat1"))
	}

	if comp2.GetProgress("quest1") != 1 {
		t.Errorf("quest1 progress = %v, want 1", comp2.GetProgress("quest1"))
	}
}

func TestNewAchievementEntry(t *testing.T) {
	entry := NewAchievementEntry("test_id", AchievementCategoryPvP)

	if entry.ID != "test_id" {
		t.Errorf("ID = %v, want test_id", entry.ID)
	}

	if entry.Category != AchievementCategoryPvP {
		t.Errorf("Category = %v, want PvP", entry.Category)
	}

	if entry.CurrentTier != AchievementTierNone {
		t.Errorf("CurrentTier = %v, want None", entry.CurrentTier)
	}

	if entry.Progress != 0 {
		t.Errorf("Progress = %v, want 0", entry.Progress)
	}

	if entry.UnlockedAt == nil {
		t.Error("UnlockedAt should be initialized")
	}
}
