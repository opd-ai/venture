// Package engine provides tests for the NG+ Reward system.
// Phase 114: NG+ Exclusive Content
package engine

import (
	"testing"
)

func TestNewNGPlusRewardSystem(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	if sys == nil {
		t.Fatal("NewNGPlusRewardSystem returned nil")
	}

	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}

	if sys.rewardCheckInterval <= 0 {
		t.Errorf("rewardCheckInterval = %d, should be > 0", sys.rewardCheckInterval)
	}
}

func TestNGPlusRewardSystem_Update_NoNGPlus(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	// Entity without NG+ component should not get rewards
	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)

	// Should not have reward component
	if _, ok := entity.GetComponent("ngplus_reward"); ok {
		t.Error("Entity without NG+ should not get reward component")
	}
}

func TestNGPlusRewardSystem_ProcessEntity_WithNGPlus(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1) // Check every update for testing

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Add NG+ component at cycle 1
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 1
	entity.AddComponent(ngp)

	// Process
	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016) // Second update triggers check

	// Should now have reward component
	rewardComp, ok := entity.GetComponent("ngplus_reward")
	if !ok {
		t.Fatal("Entity with NG+ should get reward component")
	}

	reward := rewardComp.(*NGPlusRewardComponent)

	// Should have first cycle achievement
	if !reward.HasAchievement("ngp_first_cycle") {
		t.Error("Should have ngp_first_cycle achievement at cycle 1")
	}

	// Should have tier 1
	if reward.GetHighestTierReached() != 1 {
		t.Errorf("HighestTierReached = %d, want 1", reward.GetHighestTierReached())
	}
}

func TestNGPlusRewardSystem_CycleBasedAchievements(t *testing.T) {
	tests := []struct {
		cycle       int
		achievement string
		shouldHave  bool
	}{
		{1, "ngp_first_cycle", true},
		{1, "ngp_double", false},
		{2, "ngp_double", true},
		{3, "ngp_triple", true},
		{5, "ngp_veteran", true},
		{5, "ngp_master", false},
		{10, "ngp_master", true},
	}

	for _, tt := range tests {
		t.Run(tt.achievement, func(t *testing.T) {
			sys := NewNGPlusRewardSystem(nil, 12345)
			sys.SetRewardCheckInterval(1)

			entity := &Entity{ID: 1, Components: make(map[string]Component)}
			ngp := NewNewGamePlusComponent()
			ngp.Cycle = tt.cycle
			entity.AddComponent(ngp)

			// Process twice to trigger check
			sys.Update([]*Entity{entity}, 0.016)
			sys.Update([]*Entity{entity}, 0.016)

			reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

			if reward.HasAchievement(tt.achievement) != tt.shouldHave {
				t.Errorf("At cycle %d, HasAchievement(%s) = %v, want %v",
					tt.cycle, tt.achievement, reward.HasAchievement(tt.achievement), tt.shouldHave)
			}
		})
	}
}

func TestNGPlusRewardSystem_CycleBasedTitles(t *testing.T) {
	tests := []struct {
		cycle      int
		title      string
		shouldHave bool
	}{
		{1, "title_reborn", true},
		{1, "title_twice_fallen", false},
		{2, "title_twice_fallen", true},
		{5, "title_eternal_challenger", true},
		{10, "title_legendary", true},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			sys := NewNGPlusRewardSystem(nil, 12345)
			sys.SetRewardCheckInterval(1)

			entity := &Entity{ID: 1, Components: make(map[string]Component)}
			ngp := NewNewGamePlusComponent()
			ngp.Cycle = tt.cycle
			entity.AddComponent(ngp)

			sys.Update([]*Entity{entity}, 0.016)
			sys.Update([]*Entity{entity}, 0.016)

			reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

			if reward.HasTitle(tt.title) != tt.shouldHave {
				t.Errorf("At cycle %d, HasTitle(%s) = %v, want %v",
					tt.cycle, tt.title, reward.HasTitle(tt.title), tt.shouldHave)
			}
		})
	}
}

func TestNGPlusRewardSystem_ItemAwards(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1)

	// Create entity at cycle 1
	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 1
	entity.AddComponent(ngp)

	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016)

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

	// Should have cycle 1 item
	if !reward.HasItem("ngp_sword_cycle1") {
		t.Error("Should have cycle 1 legendary item")
	}

	// Should not have cycle 5 item
	if reward.HasItem("ngp_weapon_cycle5") {
		t.Error("Should not have cycle 5 item at cycle 1")
	}
}

func TestNGPlusRewardSystem_DialogVariations(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	entity.AddComponent(ngp)

	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016)

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

	// Should have basic variations
	if !reward.HasNPCDialogVariation("npc_blacksmith_ngplus") {
		t.Error("Should have blacksmith dialog variation")
	}

	// Should have cycle 5 variation
	if !reward.HasNPCDialogVariation("npc_king_ngplus") {
		t.Error("Should have king dialog variation at cycle 5")
	}

	// Should not have cycle 10 variation
	if reward.HasNPCDialogVariation("npc_elder_ngplus") {
		t.Error("Should not have elder dialog variation at cycle 5")
	}
}

func TestNGPlusRewardSystem_SpeedrunAchievement(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Award speedrun with time under 2 hours
	sys.AwardSpeedrunAchievement(entity, 6000000) // 100 minutes

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

	if !reward.HasAchievement("ngp_speedrun") {
		t.Error("Should have speedrun achievement for sub-2-hour time")
	}

	if reward.GetTimeAttackBest("challenge_time_2h") != 6000000 {
		t.Errorf("Time attack best = %d, want 6000000", reward.GetTimeAttackBest("challenge_time_2h"))
	}
}

func TestNGPlusRewardSystem_SpeedrunAchievement_TooSlow(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Award speedrun with time over 2 hours
	sys.AwardSpeedrunAchievement(entity, 8000000) // > 2 hours

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

	if reward.HasAchievement("ngp_speedrun") {
		t.Error("Should not have speedrun achievement for >2 hour time")
	}

	// Time should still be recorded
	if reward.GetTimeAttackBest("challenge_time_2h") != 8000000 {
		t.Errorf("Time should still be recorded")
	}
}

func TestNGPlusRewardSystem_NoDeathAchievement(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Start no-death challenge
	sys.StartNoDeathChallenge(entity, 1)

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

	// Check challenge is active
	active := reward.GetActiveChallenges()
	found := false
	for _, c := range active {
		if c == "challenge_nodeaths" {
			found = true
			break
		}
	}
	if !found {
		t.Error("No-death challenge should be active")
	}

	// Complete it
	sys.AwardNoDeathAchievement(entity, 1)

	if !reward.HasAchievement("ngp_nodeaths") {
		t.Error("Should have no-death achievement after completion")
	}

	if !reward.HasNoDeathCompletion() {
		t.Error("Should have no-death completion recorded")
	}
}

func TestNGPlusRewardSystem_NoDeathFail(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Start and fail
	sys.StartNoDeathChallenge(entity, 1)
	sys.FailNoDeathChallenge(entity, 1, "Boss Room")

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)

	data, ok := reward.GetNoDeathRunData(1)
	if !ok {
		t.Fatal("Should have run data")
	}

	if data.IsActive {
		t.Error("Run should not be active after failure")
	}
	if data.WasCompleted {
		t.Error("Run should not be completed after failure")
	}
	if data.FailedAt != "Boss Room" {
		t.Errorf("FailedAt = %q, want 'Boss Room'", data.FailedAt)
	}
}

func TestNGPlusRewardSystem_AllBossesAchievement(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Try at cycle 3 - should not work
	sys.AwardAllBossesAchievement(entity, 3)
	if rewardComp, ok := entity.GetComponent("ngplus_reward"); ok {
		reward := rewardComp.(*NGPlusRewardComponent)
		if reward.HasAchievement("ngp_allbosses") {
			t.Error("Should not have all-bosses achievement at cycle 3")
		}
	}

	// Try at cycle 5 - should work
	sys.AwardAllBossesAchievement(entity, 5)
	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)
	if !reward.HasAchievement("ngp_allbosses") {
		t.Error("Should have all-bosses achievement at cycle 5")
	}
}

func TestNGPlusRewardSystem_GetAvailableChallenges(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	// At cycle 1
	challenges := sys.GetAvailableChallenges(1)
	if len(challenges) < 3 {
		t.Errorf("Should have at least 3 challenges at cycle 1, got %d", len(challenges))
	}

	// At cycle 5
	challenges = sys.GetAvailableChallenges(5)
	if len(challenges) < 4 {
		t.Errorf("Should have at least 4 challenges at cycle 5, got %d", len(challenges))
	}
}

func TestNGPlusRewardSystem_GetCycleLegendaryItem(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	// Same seed + cycle = same item
	item1 := sys.GetCycleLegendaryItem(5)
	item2 := sys.GetCycleLegendaryItem(5)

	if item1 != item2 {
		t.Errorf("Same cycle should give same item, got %s and %s", item1, item2)
	}

	// Cycle 0 = no item
	item := sys.GetCycleLegendaryItem(0)
	if item != "" {
		t.Errorf("Cycle 0 should return empty, got %s", item)
	}
}

func TestNGPlusRewardSystem_GetPlayerRewardSummary(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	entity.AddComponent(ngp)

	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016)

	summary := sys.GetPlayerRewardSummary(entity)

	if summary.AchievementCount < 1 {
		t.Error("Should have at least 1 achievement")
	}
	if summary.HighestTier != 5 {
		t.Errorf("HighestTier = %d, want 5", summary.HighestTier)
	}
}

func TestNGPlusRewardSystem_CompleteChallenge(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}

	// Activate and complete
	reward := NewNGPlusRewardComponent()
	entity.AddComponent(reward)
	reward.ActivateChallenge("challenge_time_2h")

	completed := false
	sys.SetOnChallengeCompleted(func(entityID uint64, challengeID string) {
		completed = true
	})

	sys.CompleteChallenge(entity, "challenge_time_2h")

	if !completed {
		t.Error("Challenge completed callback should have been called")
	}

	if !reward.IsChallengeCompleted("challenge_time_2h") {
		t.Error("Challenge should be marked as completed")
	}
}

func TestNGPlusRewardSystem_Callbacks(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1)

	achievementCalled := false
	itemCalled := false
	titleCalled := false

	sys.SetOnAchievementUnlocked(func(entityID uint64, achievementID string) {
		achievementCalled = true
	})
	sys.SetOnItemAwarded(func(entityID uint64, itemID string) {
		itemCalled = true
	})
	sys.SetOnTitleUnlocked(func(entityID uint64, titleID string) {
		titleCalled = true
	})

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 1
	entity.AddComponent(ngp)

	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016)

	if !achievementCalled {
		t.Error("Achievement callback should have been called")
	}
	if !itemCalled {
		t.Error("Item callback should have been called")
	}
	if !titleCalled {
		t.Error("Title callback should have been called")
	}
}

func TestIsNGPlusExclusiveItem(t *testing.T) {
	if !IsNGPlusExclusiveItem("ngp_sword_cycle1") {
		t.Error("ngp_sword_cycle1 should be exclusive")
	}

	if IsNGPlusExclusiveItem("regular_sword") {
		t.Error("regular_sword should not be exclusive")
	}
}

func TestGetNGPlusItemDefinition(t *testing.T) {
	def, ok := GetNGPlusItemDefinition("ngp_sword_cycle1")
	if !ok {
		t.Fatal("Should find ngp_sword_cycle1")
	}
	if def.Name != "Blade of Rebirth" {
		t.Errorf("Name = %q, want 'Blade of Rebirth'", def.Name)
	}

	_, ok = GetNGPlusItemDefinition("nonexistent")
	if ok {
		t.Error("Should not find nonexistent item")
	}
}

func TestGetNGPlusAchievementDefinition(t *testing.T) {
	def, ok := GetNGPlusAchievementDefinition("ngp_first_cycle")
	if !ok {
		t.Fatal("Should find ngp_first_cycle")
	}
	if def.Name != "Reborn" {
		t.Errorf("Name = %q, want 'Reborn'", def.Name)
	}

	_, ok = GetNGPlusAchievementDefinition("nonexistent")
	if ok {
		t.Error("Should not find nonexistent achievement")
	}
}

func TestGetNGPlusTitleDefinition(t *testing.T) {
	def, ok := GetNGPlusTitleDefinition("title_reborn")
	if !ok {
		t.Fatal("Should find title_reborn")
	}
	if def.Display != "Reborn" {
		t.Errorf("Display = %q, want 'Reborn'", def.Display)
	}

	_, ok = GetNGPlusTitleDefinition("nonexistent")
	if ok {
		t.Error("Should not find nonexistent title")
	}
}

func TestGenerateNPCDialogVariation(t *testing.T) {
	// Cycle 0 returns empty
	dialog := GenerateNPCDialogVariation(12345, "blacksmith", 0)
	if dialog != "" {
		t.Errorf("Cycle 0 should return empty, got %q", dialog)
	}

	// Same seed/npc/cycle = same dialog
	dialog1 := GenerateNPCDialogVariation(12345, "blacksmith", 1)
	dialog2 := GenerateNPCDialogVariation(12345, "blacksmith", 1)
	if dialog1 != dialog2 {
		t.Error("Same inputs should give same dialog")
	}

	// Different NPC = different dialog (usually)
	dialog3 := GenerateNPCDialogVariation(12345, "merchant", 1)
	// May be same by chance, but we test the function works
	_ = dialog3
}

func TestGetUIIndicatorForNGPlusContent(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{"achievement", "[NG+] "},
		{"item", "⟳ "},
		{"title", "★ "},
		{"challenge", "⚔ "},
		{"unknown", "[NG+] "},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := GetUIIndicatorForNGPlusContent(tt.contentType)
			if got != tt.want {
				t.Errorf("GetUIIndicatorForNGPlusContent(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestNGPlusRewardSystem_ForceRewardCheck(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	// Don't set interval to 1 - use default

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 1
	entity.AddComponent(ngp)

	// Force check without waiting for interval
	sys.ForceRewardCheck(entity)

	reward := entity.Components["ngplus_reward"].(*NGPlusRewardComponent)
	if !reward.HasAchievement("ngp_first_cycle") {
		t.Error("Force check should award achievements")
	}
}

func TestNGPlusRewardSystem_DeathlessTitleFromNoDeathCompletion(t *testing.T) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 1
	entity.AddComponent(ngp)

	// Complete a no-death run first
	reward := NewNGPlusRewardComponent()
	entity.AddComponent(reward)
	reward.StartNoDeathRun(1)
	reward.CompleteNoDeathRun(1)

	// Now update system to check for titles
	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016)

	if !reward.HasTitle("title_deathless") {
		t.Error("Should have deathless title after completing no-death run")
	}
}

func BenchmarkNGPlusRewardSystem_Update(b *testing.B) {
	sys := NewNGPlusRewardSystem(nil, 12345)
	sys.SetRewardCheckInterval(1)

	entities := make([]*Entity, 100)
	for i := range entities {
		entities[i] = &Entity{ID: uint64(i), Components: make(map[string]Component)}
		ngp := NewNewGamePlusComponent()
		ngp.Cycle = 5
		entities[i].AddComponent(ngp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkNGPlusRewardSystem_ProcessEntity(b *testing.B) {
	sys := NewNGPlusRewardSystem(nil, 12345)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 10
	entity.AddComponent(ngp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ForceRewardCheck(entity)
	}
}
