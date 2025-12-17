// Package engine provides tests for the NG+ Reward component.
// Phase 114: NG+ Exclusive Content
package engine

import (
	"encoding/json"
	"testing"
)

func TestNGPlusRewardComponent_Type(t *testing.T) {
	comp := NewNGPlusRewardComponent()
	if comp.Type() != "ngplus_reward" {
		t.Errorf("Type() = %v, want ngplus_reward", comp.Type())
	}
}

func TestNewNGPlusRewardComponent(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	if len(comp.ExclusiveAchievements) != 0 {
		t.Errorf("ExclusiveAchievements should be empty, got %d", len(comp.ExclusiveAchievements))
	}
	if len(comp.ExclusiveItems) != 0 {
		t.Errorf("ExclusiveItems should be empty, got %d", len(comp.ExclusiveItems))
	}
	if len(comp.TitlesUnlocked) != 0 {
		t.Errorf("TitlesUnlocked should be empty, got %d", len(comp.TitlesUnlocked))
	}
	if len(comp.ChallengesCompleted) != 0 {
		t.Errorf("ChallengesCompleted should be empty, got %d", len(comp.ChallengesCompleted))
	}
	if comp.HighestTierReached != 0 {
		t.Errorf("HighestTierReached = %d, want 0", comp.HighestTierReached)
	}
	if comp.CurrentTitle != "" {
		t.Errorf("CurrentTitle = %q, want empty", comp.CurrentTitle)
	}
}

func TestNGPlusRewardComponent_Achievements(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	// Initially no achievements
	if comp.HasAchievement("ngp_first_cycle") {
		t.Error("Should not have achievement before unlocking")
	}

	// Unlock achievement
	if !comp.UnlockAchievement("ngp_first_cycle") {
		t.Error("UnlockAchievement should return true for new achievement")
	}

	// Should now have it
	if !comp.HasAchievement("ngp_first_cycle") {
		t.Error("Should have achievement after unlocking")
	}

	// Duplicate unlock should return false
	if comp.UnlockAchievement("ngp_first_cycle") {
		t.Error("UnlockAchievement should return false for duplicate")
	}

	// Get all achievements
	achievements := comp.GetAchievements()
	if len(achievements) != 1 || achievements[0] != "ngp_first_cycle" {
		t.Errorf("GetAchievements() = %v, want [ngp_first_cycle]", achievements)
	}
}

func TestNGPlusRewardComponent_Items(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	if comp.HasItem("ngp_sword_cycle1") {
		t.Error("Should not have item before adding")
	}

	if !comp.AddItem("ngp_sword_cycle1") {
		t.Error("AddItem should return true for new item")
	}

	if !comp.HasItem("ngp_sword_cycle1") {
		t.Error("Should have item after adding")
	}

	if comp.AddItem("ngp_sword_cycle1") {
		t.Error("AddItem should return false for duplicate")
	}

	items := comp.GetItems()
	if len(items) != 1 || items[0] != "ngp_sword_cycle1" {
		t.Errorf("GetItems() = %v, want [ngp_sword_cycle1]", items)
	}
}

func TestNGPlusRewardComponent_Titles(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	if comp.HasTitle("title_reborn") {
		t.Error("Should not have title before unlocking")
	}

	if !comp.UnlockTitle("title_reborn") {
		t.Error("UnlockTitle should return true for new title")
	}

	if !comp.HasTitle("title_reborn") {
		t.Error("Should have title after unlocking")
	}

	if comp.UnlockTitle("title_reborn") {
		t.Error("UnlockTitle should return false for duplicate")
	}

	// Test current title
	if comp.GetCurrentTitle() != "" {
		t.Error("Current title should be empty initially")
	}

	comp.SetCurrentTitle("title_reborn")
	if comp.GetCurrentTitle() != "title_reborn" {
		t.Errorf("GetCurrentTitle() = %q, want title_reborn", comp.GetCurrentTitle())
	}

	titles := comp.GetTitles()
	if len(titles) != 1 || titles[0] != "title_reborn" {
		t.Errorf("GetTitles() = %v, want [title_reborn]", titles)
	}
}

func TestNGPlusRewardComponent_Challenges(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	if comp.IsChallengeCompleted("challenge_nodeaths") {
		t.Error("Challenge should not be completed initially")
	}

	// Activate challenge
	comp.ActivateChallenge("challenge_nodeaths")
	active := comp.GetActiveChallenges()
	if len(active) != 1 || active[0] != "challenge_nodeaths" {
		t.Errorf("GetActiveChallenges() = %v, want [challenge_nodeaths]", active)
	}

	// Activating again should not duplicate
	comp.ActivateChallenge("challenge_nodeaths")
	active = comp.GetActiveChallenges()
	if len(active) != 1 {
		t.Errorf("Duplicate activate should not add, got %d items", len(active))
	}

	// Complete challenge
	if !comp.CompleteChallenge("challenge_nodeaths") {
		t.Error("CompleteChallenge should return true for new completion")
	}

	if !comp.IsChallengeCompleted("challenge_nodeaths") {
		t.Error("Challenge should be completed after completion")
	}

	if comp.CompleteChallenge("challenge_nodeaths") {
		t.Error("CompleteChallenge should return false for duplicate")
	}

	// Should be removed from active
	active = comp.GetActiveChallenges()
	if len(active) != 0 {
		t.Errorf("Completed challenge should be removed from active, got %v", active)
	}

	// Cannot activate completed challenge
	comp.ActivateChallenge("challenge_nodeaths")
	active = comp.GetActiveChallenges()
	if len(active) != 0 {
		t.Error("Should not be able to activate completed challenge")
	}

	completed := comp.GetCompletedChallenges()
	if len(completed) != 1 {
		t.Errorf("GetCompletedChallenges() = %v, want 1 item", completed)
	}
}

func TestNGPlusRewardComponent_HighestTier(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	if comp.GetHighestTierReached() != 0 {
		t.Errorf("Initial tier = %d, want 0", comp.GetHighestTierReached())
	}

	comp.UpdateHighestTier(5)
	if comp.GetHighestTierReached() != 5 {
		t.Errorf("After update to 5, tier = %d, want 5", comp.GetHighestTierReached())
	}

	// Lower value should not update
	comp.UpdateHighestTier(3)
	if comp.GetHighestTierReached() != 5 {
		t.Errorf("After update to 3, tier = %d, want 5 (should not decrease)", comp.GetHighestTierReached())
	}

	comp.UpdateHighestTier(10)
	if comp.GetHighestTierReached() != 10 {
		t.Errorf("After update to 10, tier = %d, want 10", comp.GetHighestTierReached())
	}
}

func TestNGPlusRewardComponent_TimeAttack(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	// No initial time
	if comp.GetTimeAttackBest("challenge_time_2h") != 0 {
		t.Error("Initial time should be 0")
	}

	// Record first time
	if !comp.RecordTimeAttack("challenge_time_2h", 7000000) {
		t.Error("First record should return true")
	}

	if comp.GetTimeAttackBest("challenge_time_2h") != 7000000 {
		t.Errorf("GetTimeAttackBest() = %d, want 7000000", comp.GetTimeAttackBest("challenge_time_2h"))
	}

	// Worse time should not update
	if comp.RecordTimeAttack("challenge_time_2h", 8000000) {
		t.Error("Worse time should return false")
	}

	if comp.GetTimeAttackBest("challenge_time_2h") != 7000000 {
		t.Error("Worse time should not update best")
	}

	// Better time should update
	if !comp.RecordTimeAttack("challenge_time_2h", 6000000) {
		t.Error("Better time should return true")
	}

	if comp.GetTimeAttackBest("challenge_time_2h") != 6000000 {
		t.Errorf("After better time, GetTimeAttackBest() = %d, want 6000000", comp.GetTimeAttackBest("challenge_time_2h"))
	}
}

func TestNGPlusRewardComponent_NoDeathRun(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	// No run initially
	if _, ok := comp.GetNoDeathRunData(1); ok {
		t.Error("Should not have no-death data initially")
	}

	if comp.HasNoDeathCompletion() {
		t.Error("Should not have any completion initially")
	}

	// Start run
	comp.StartNoDeathRun(1)
	data, ok := comp.GetNoDeathRunData(1)
	if !ok {
		t.Error("Should have no-death data after starting")
	}
	if !data.IsActive {
		t.Error("Run should be active after starting")
	}
	if data.WasCompleted {
		t.Error("Run should not be completed after starting")
	}

	// Update progress
	comp.UpdateNoDeathRun(1, 5, 3)
	data, _ = comp.GetNoDeathRunData(1)
	if data.BossesDefeated != 5 || data.AreasCleared != 3 {
		t.Errorf("Progress = bosses:%d areas:%d, want bosses:5 areas:3", data.BossesDefeated, data.AreasCleared)
	}

	// Fail run
	comp.FailNoDeathRun(1, "Boss Room 3")
	data, _ = comp.GetNoDeathRunData(1)
	if data.IsActive {
		t.Error("Run should not be active after failing")
	}
	if data.WasCompleted {
		t.Error("Run should not be completed after failing")
	}
	if data.FailedAt != "Boss Room 3" {
		t.Errorf("FailedAt = %q, want 'Boss Room 3'", data.FailedAt)
	}

	// Start new run for cycle 2
	comp.StartNoDeathRun(2)
	comp.CompleteNoDeathRun(2)
	data, _ = comp.GetNoDeathRunData(2)
	if data.IsActive {
		t.Error("Run should not be active after completing")
	}
	if !data.WasCompleted {
		t.Error("Run should be completed after completing")
	}

	if !comp.HasNoDeathCompletion() {
		t.Error("Should have completion after completing run")
	}
}

func TestNGPlusRewardComponent_NPCDialogVariations(t *testing.T) {
	comp := NewNGPlusRewardComponent()

	if comp.HasNPCDialogVariation("npc_blacksmith_ngplus") {
		t.Error("Should not have variation initially")
	}

	if !comp.UnlockNPCDialogVariation("npc_blacksmith_ngplus") {
		t.Error("First unlock should return true")
	}

	if !comp.HasNPCDialogVariation("npc_blacksmith_ngplus") {
		t.Error("Should have variation after unlocking")
	}

	if comp.UnlockNPCDialogVariation("npc_blacksmith_ngplus") {
		t.Error("Duplicate unlock should return false")
	}

	variations := comp.GetNPCDialogVariations()
	if len(variations) != 1 || variations[0] != "npc_blacksmith_ngplus" {
		t.Errorf("GetNPCDialogVariations() = %v, want [npc_blacksmith_ngplus]", variations)
	}
}

func TestNGPlusRewardComponent_Serialize(t *testing.T) {
	comp := NewNGPlusRewardComponent()
	comp.UnlockAchievement("ngp_first_cycle")
	comp.AddItem("ngp_sword_cycle1")
	comp.UnlockTitle("title_reborn")
	comp.SetCurrentTitle("title_reborn")
	comp.CompleteChallenge("challenge_nodeaths")
	comp.UpdateHighestTier(5)
	comp.RecordTimeAttack("challenge_time_2h", 6500000)
	comp.StartNoDeathRun(1)
	comp.CompleteNoDeathRun(1)
	comp.UnlockNPCDialogVariation("npc_test")

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Verify it's valid JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		t.Fatalf("Serialized data is not valid JSON: %v", err)
	}
}

func TestNGPlusRewardComponent_Deserialize(t *testing.T) {
	// Create and populate original
	original := NewNGPlusRewardComponent()
	original.UnlockAchievement("ngp_first_cycle")
	original.UnlockAchievement("ngp_double")
	original.AddItem("ngp_sword_cycle1")
	original.UnlockTitle("title_reborn")
	original.SetCurrentTitle("title_reborn")
	original.CompleteChallenge("challenge_nodeaths")
	original.ActivateChallenge("challenge_time_2h")
	original.UpdateHighestTier(5)
	original.RecordTimeAttack("challenge_time_2h", 6500000)
	original.StartNoDeathRun(1)
	original.CompleteNoDeathRun(1)
	original.UnlockNPCDialogVariation("npc_test")

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	restored := NewNGPlusRewardComponent()
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify all fields restored
	if !restored.HasAchievement("ngp_first_cycle") {
		t.Error("Achievement ngp_first_cycle not restored")
	}
	if !restored.HasAchievement("ngp_double") {
		t.Error("Achievement ngp_double not restored")
	}
	if !restored.HasItem("ngp_sword_cycle1") {
		t.Error("Item ngp_sword_cycle1 not restored")
	}
	if !restored.HasTitle("title_reborn") {
		t.Error("Title title_reborn not restored")
	}
	if restored.GetCurrentTitle() != "title_reborn" {
		t.Errorf("CurrentTitle = %q, want title_reborn", restored.GetCurrentTitle())
	}
	if !restored.IsChallengeCompleted("challenge_nodeaths") {
		t.Error("Challenge challenge_nodeaths not restored as completed")
	}
	active := restored.GetActiveChallenges()
	if len(active) != 1 || active[0] != "challenge_time_2h" {
		t.Errorf("Active challenges = %v, want [challenge_time_2h]", active)
	}
	if restored.GetHighestTierReached() != 5 {
		t.Errorf("HighestTierReached = %d, want 5", restored.GetHighestTierReached())
	}
	if restored.GetTimeAttackBest("challenge_time_2h") != 6500000 {
		t.Errorf("TimeAttackBest = %d, want 6500000", restored.GetTimeAttackBest("challenge_time_2h"))
	}
	if !restored.HasNoDeathCompletion() {
		t.Error("NoDeathCompletion not restored")
	}
	if !restored.HasNPCDialogVariation("npc_test") {
		t.Error("NPC dialog variation not restored")
	}
}

func TestGetNGPlusAchievements(t *testing.T) {
	achievements := GetNGPlusAchievements()
	if len(achievements) != 10 {
		t.Errorf("GetNGPlusAchievements() returned %d items, want 10", len(achievements))
	}

	// Verify unique IDs
	ids := make(map[string]bool)
	for _, a := range achievements {
		if ids[a.ID] {
			t.Errorf("Duplicate achievement ID: %s", a.ID)
		}
		ids[a.ID] = true

		if a.Name == "" {
			t.Errorf("Achievement %s has empty name", a.ID)
		}
		if a.Description == "" {
			t.Errorf("Achievement %s has empty description", a.ID)
		}
		if a.MinCycle < 1 {
			t.Errorf("Achievement %s has invalid MinCycle: %d", a.ID, a.MinCycle)
		}
	}
}

func TestGetNGPlusLegendaryItems(t *testing.T) {
	items := GetNGPlusLegendaryItems()
	if len(items) != 10 {
		t.Errorf("GetNGPlusLegendaryItems() returned %d items, want 10", len(items))
	}

	ids := make(map[string]bool)
	for _, item := range items {
		if ids[item.ID] {
			t.Errorf("Duplicate item ID: %s", item.ID)
		}
		ids[item.ID] = true

		if item.Name == "" {
			t.Errorf("Item %s has empty name", item.ID)
		}
		if item.Description == "" {
			t.Errorf("Item %s has empty description", item.ID)
		}
		if item.MinCycle < 1 {
			t.Errorf("Item %s has invalid MinCycle: %d", item.ID, item.MinCycle)
		}
		if item.SpecialEffect == "" {
			t.Errorf("Item %s has empty special effect", item.ID)
		}
	}
}

func TestGetNGPlusTitles(t *testing.T) {
	titles := GetNGPlusTitles()
	if len(titles) < 10 {
		t.Errorf("GetNGPlusTitles() returned %d items, want at least 10", len(titles))
	}

	ids := make(map[string]bool)
	for _, title := range titles {
		if ids[title.ID] {
			t.Errorf("Duplicate title ID: %s", title.ID)
		}
		ids[title.ID] = true

		if title.Display == "" {
			t.Errorf("Title %s has empty display name", title.ID)
		}
		if title.MinCycle < 1 {
			t.Errorf("Title %s has invalid MinCycle: %d", title.ID, title.MinCycle)
		}
	}
}

func TestGetNGPlusChallenges(t *testing.T) {
	challenges := GetNGPlusChallenges()
	if len(challenges) < 5 {
		t.Errorf("GetNGPlusChallenges() returned %d items, want at least 5", len(challenges))
	}

	ids := make(map[string]bool)
	for _, c := range challenges {
		if ids[c.ID] {
			t.Errorf("Duplicate challenge ID: %s", c.ID)
		}
		ids[c.ID] = true

		if c.Name == "" {
			t.Errorf("Challenge %s has empty name", c.ID)
		}
		if c.Description == "" {
			t.Errorf("Challenge %s has empty description", c.ID)
		}
		if c.ChallengeTyp == "" {
			t.Errorf("Challenge %s has empty type", c.ID)
		}
	}
}

func TestGenerateDeterministicLegendaryItem(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		cycle int
	}{
		{"cycle_0", 12345, 0},
		{"cycle_1", 12345, 1},
		{"cycle_5", 12345, 5},
		{"cycle_10", 12345, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate twice with same seed/cycle - should be identical
			item1 := GenerateDeterministicLegendaryItem(tt.seed, tt.cycle)
			item2 := GenerateDeterministicLegendaryItem(tt.seed, tt.cycle)

			if item1 != item2 {
				t.Errorf("Same seed/cycle gave different items: %s vs %s", item1, item2)
			}

			// Cycle 0 should return empty
			if tt.cycle == 0 && item1 != "" {
				t.Errorf("Cycle 0 should return empty, got %s", item1)
			}

			// Non-zero cycles should return valid item
			if tt.cycle > 0 && item1 == "" {
				t.Error("Non-zero cycle should return item")
			}
		})
	}

	// Different seeds should (usually) give different items
	item1 := GenerateDeterministicLegendaryItem(12345, 5)
	item2 := GenerateDeterministicLegendaryItem(54321, 5)
	// Note: Could theoretically be same, but unlikely with different seeds
	_ = item1
	_ = item2
}

func TestGenerateDeterministicLegendaryItem_Determinism(t *testing.T) {
	seed := int64(999)
	cycle := 7

	// Generate multiple times
	results := make([]string, 100)
	for i := 0; i < 100; i++ {
		results[i] = GenerateDeterministicLegendaryItem(seed, cycle)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Result %d differs: %s vs %s", i, results[i], results[0])
		}
	}
}

func TestGetTierForCycle(t *testing.T) {
	tests := []struct {
		cycle    int
		wantTier int
	}{
		{-1, 0},
		{0, 0},
		{1, 1},
		{5, 5},
		{9, 9},
		{10, 10},
		{15, 10},
		{99, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := GetTierForCycle(tt.cycle); got != tt.wantTier {
				t.Errorf("GetTierForCycle(%d) = %d, want %d", tt.cycle, got, tt.wantTier)
			}
		})
	}
}

func BenchmarkNGPlusRewardComponent_UnlockAchievement(b *testing.B) {
	comp := NewNGPlusRewardComponent()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comp.UnlockAchievement("test_achievement")
	}
}

func BenchmarkNGPlusRewardComponent_HasAchievement(b *testing.B) {
	comp := NewNGPlusRewardComponent()
	comp.UnlockAchievement("test_achievement")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comp.HasAchievement("test_achievement")
	}
}

func BenchmarkNGPlusRewardComponent_Serialize(b *testing.B) {
	comp := NewNGPlusRewardComponent()
	for i := 0; i < 10; i++ {
		comp.UnlockAchievement("achievement_" + string(rune('a'+i)))
		comp.AddItem("item_" + string(rune('a'+i)))
		comp.UnlockTitle("title_" + string(rune('a'+i)))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = comp.Serialize()
	}
}
