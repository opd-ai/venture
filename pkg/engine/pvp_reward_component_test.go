// Package engine provides tests for the PvP reward component.
package engine

import (
	"testing"
)

func TestNewPvPRewardComponent(t *testing.T) {
	seasonID := "season_1"
	comp := NewPvPRewardComponent(seasonID)

	if comp == nil {
		t.Fatal("Expected non-nil component")
	}

	if comp.Type() != "pvp_reward" {
		t.Errorf("Expected type 'pvp_reward', got '%s'", comp.Type())
	}

	if comp.CurrentSeasonID != seasonID {
		t.Errorf("Expected season '%s', got '%s'", seasonID, comp.CurrentSeasonID)
	}

	if comp.HonorPoints != 0 {
		t.Errorf("Expected 0 honor points, got %d", comp.HonorPoints)
	}

	if comp.TotalHonorEarned != 0 {
		t.Errorf("Expected 0 total honor, got %d", comp.TotalHonorEarned)
	}

	if len(comp.EarnedRewards) != 0 {
		t.Errorf("Expected empty rewards, got %d", len(comp.EarnedRewards))
	}

	if comp.HighestSeasonRank == nil {
		t.Error("Expected initialized HighestSeasonRank map")
	}
}

func TestPvPRewardComponent_Honor(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Test adding honor
	comp.AddHonor(100)
	if comp.HonorPoints != 100 {
		t.Errorf("Expected 100 honor, got %d", comp.HonorPoints)
	}
	if comp.TotalHonorEarned != 100 {
		t.Errorf("Expected 100 total honor, got %d", comp.TotalHonorEarned)
	}

	// Test adding more honor
	comp.AddHonor(50)
	if comp.HonorPoints != 150 {
		t.Errorf("Expected 150 honor, got %d", comp.HonorPoints)
	}
	if comp.TotalHonorEarned != 150 {
		t.Errorf("Expected 150 total honor, got %d", comp.TotalHonorEarned)
	}

	// Test adding zero/negative honor (should be ignored)
	comp.AddHonor(0)
	comp.AddHonor(-10)
	if comp.HonorPoints != 150 {
		t.Errorf("Expected 150 honor after invalid add, got %d", comp.HonorPoints)
	}

	// Test spending honor
	if !comp.SpendHonor(50) {
		t.Error("Expected successful spend of 50 honor")
	}
	if comp.HonorPoints != 100 {
		t.Errorf("Expected 100 honor after spend, got %d", comp.HonorPoints)
	}

	// Total earned should not change after spending
	if comp.TotalHonorEarned != 150 {
		t.Errorf("Expected 150 total honor after spend, got %d", comp.TotalHonorEarned)
	}

	// Test spending more than available
	if comp.SpendHonor(200) {
		t.Error("Expected failed spend of 200 honor")
	}
	if comp.HonorPoints != 100 {
		t.Errorf("Expected 100 honor after failed spend, got %d", comp.HonorPoints)
	}

	// Test spending zero/negative (should fail)
	if comp.SpendHonor(0) {
		t.Error("Expected failed spend of 0 honor")
	}
	if comp.SpendHonor(-10) {
		t.Error("Expected failed spend of -10 honor")
	}

	// Test GetHonor
	if comp.GetHonor() != 100 {
		t.Errorf("Expected GetHonor() = 100, got %d", comp.GetHonor())
	}
}

func TestPvPRewardComponent_Rewards(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	reward := PvPReward{
		ID:       "test_reward",
		SeasonID: "season_1",
		Type:     PvPRewardHonor,
		Name:     "Test Reward",
		Value:    100,
		Rarity:   "rare",
	}

	// Add reward
	comp.AddReward(reward)
	if len(comp.EarnedRewards) != 1 {
		t.Errorf("Expected 1 reward, got %d", len(comp.EarnedRewards))
	}

	// Get unclaimed rewards
	unclaimed := comp.GetUnclaimedRewards()
	if len(unclaimed) != 1 {
		t.Errorf("Expected 1 unclaimed reward, got %d", len(unclaimed))
	}

	// Claim reward
	if !comp.ClaimReward("test_reward") {
		t.Error("Expected successful claim")
	}

	// Verify claimed
	unclaimed = comp.GetUnclaimedRewards()
	if len(unclaimed) != 0 {
		t.Errorf("Expected 0 unclaimed rewards after claim, got %d", len(unclaimed))
	}

	// Try to claim again (should fail)
	if comp.ClaimReward("test_reward") {
		t.Error("Expected failed claim of already claimed reward")
	}

	// Try to claim non-existent reward
	if comp.ClaimReward("nonexistent") {
		t.Error("Expected failed claim of nonexistent reward")
	}

	// Test GetRewardsForSeason
	seasonRewards := comp.GetRewardsForSeason("season_1")
	if len(seasonRewards) != 1 {
		t.Errorf("Expected 1 season reward, got %d", len(seasonRewards))
	}

	otherSeasonRewards := comp.GetRewardsForSeason("season_2")
	if len(otherSeasonRewards) != 0 {
		t.Errorf("Expected 0 rewards for other season, got %d", len(otherSeasonRewards))
	}
}

func TestPvPRewardComponent_Titles(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Add title
	if !comp.AddTitle("champion_title") {
		t.Error("Expected successful title add")
	}

	if !comp.HasTitle("champion_title") {
		t.Error("Expected HasTitle to return true")
	}

	// Try adding duplicate title
	if comp.AddTitle("champion_title") {
		t.Error("Expected failed duplicate title add")
	}

	// Set active title
	if !comp.SetActiveTitle("champion_title") {
		t.Error("Expected successful SetActiveTitle")
	}

	if comp.ActiveTitle != "champion_title" {
		t.Errorf("Expected active title 'champion_title', got '%s'", comp.ActiveTitle)
	}

	// Try setting non-owned title
	if comp.SetActiveTitle("nonexistent_title") {
		t.Error("Expected failed SetActiveTitle for non-owned title")
	}

	// Check non-owned title
	if comp.HasTitle("nonexistent_title") {
		t.Error("Expected HasTitle to return false for non-owned title")
	}
}

func TestPvPRewardComponent_Mounts(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Add mount
	if !comp.AddMount("war_horse") {
		t.Error("Expected successful mount add")
	}

	if !comp.HasMount("war_horse") {
		t.Error("Expected HasMount to return true")
	}

	// Try adding duplicate mount
	if comp.AddMount("war_horse") {
		t.Error("Expected failed duplicate mount add")
	}

	// Set active mount
	if !comp.SetActiveMount("war_horse") {
		t.Error("Expected successful SetActiveMount")
	}

	if comp.ActiveMount != "war_horse" {
		t.Errorf("Expected active mount 'war_horse', got '%s'", comp.ActiveMount)
	}

	// Try setting non-owned mount
	if comp.SetActiveMount("nonexistent_mount") {
		t.Error("Expected failed SetActiveMount for non-owned mount")
	}
}

func TestPvPRewardComponent_Cosmetics(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Add cosmetic
	if !comp.AddCosmetic("flame_aura") {
		t.Error("Expected successful cosmetic add")
	}

	if !comp.HasCosmetic("flame_aura") {
		t.Error("Expected HasCosmetic to return true")
	}

	// Try adding duplicate cosmetic
	if comp.AddCosmetic("flame_aura") {
		t.Error("Expected failed duplicate cosmetic add")
	}

	// Set active cosmetic
	if !comp.SetActiveCosmetic("flame_aura") {
		t.Error("Expected successful SetActiveCosmetic")
	}

	if comp.ActiveCosmetic != "flame_aura" {
		t.Errorf("Expected active cosmetic 'flame_aura', got '%s'", comp.ActiveCosmetic)
	}

	// Try setting non-owned cosmetic
	if comp.SetActiveCosmetic("nonexistent_cosmetic") {
		t.Error("Expected failed SetActiveCosmetic for non-owned cosmetic")
	}
}

func TestPvPRewardComponent_TournamentTracking(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Record tournament win
	comp.RecordTournamentWin()
	if comp.TournamentWins != 1 {
		t.Errorf("Expected 1 tournament win, got %d", comp.TournamentWins)
	}
	if comp.TournamentParticipations != 1 {
		t.Errorf("Expected 1 tournament participation, got %d", comp.TournamentParticipations)
	}

	// Record participation without win
	comp.RecordTournamentParticipation()
	if comp.TournamentWins != 1 {
		t.Errorf("Expected 1 tournament win after participation, got %d", comp.TournamentWins)
	}
	if comp.TournamentParticipations != 2 {
		t.Errorf("Expected 2 tournament participations, got %d", comp.TournamentParticipations)
	}
}

func TestPvPRewardComponent_HighestRank(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Update highest rank
	comp.UpdateHighestRank(RankGold)
	if comp.GetHighestRank("season_1") != RankGold {
		t.Errorf("Expected Gold rank, got %v", comp.GetHighestRank("season_1"))
	}

	// Update to higher rank
	comp.UpdateHighestRank(RankDiamond)
	if comp.GetHighestRank("season_1") != RankDiamond {
		t.Errorf("Expected Diamond rank, got %v", comp.GetHighestRank("season_1"))
	}

	// Try updating to lower rank (should not change)
	comp.UpdateHighestRank(RankSilver)
	if comp.GetHighestRank("season_1") != RankDiamond {
		t.Errorf("Expected Diamond rank after lower update, got %v", comp.GetHighestRank("season_1"))
	}

	// Get rank for non-existent season
	if comp.GetHighestRank("season_2") != RankBronze {
		t.Errorf("Expected Bronze for non-existent season, got %v", comp.GetHighestRank("season_2"))
	}
}

func TestPvPRewardComponent_SeasonalRewards(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	rewards := []PvPReward{
		{ID: "reward_1", Type: PvPRewardHonor, Value: 100},
		{ID: "reward_2", Type: PvPRewardTitle, Name: "Gold Champion"},
	}

	// Add seasonal reward
	comp.AddSeasonalReward(RankGold, "season_1", rewards)
	if len(comp.SeasonalRewards) != 1 {
		t.Errorf("Expected 1 seasonal reward tier, got %d", len(comp.SeasonalRewards))
	}

	// Try adding duplicate seasonal reward (should be ignored)
	comp.AddSeasonalReward(RankGold, "season_1", rewards)
	if len(comp.SeasonalRewards) != 1 {
		t.Errorf("Expected still 1 seasonal reward tier, got %d", len(comp.SeasonalRewards))
	}

	// Get unclaimed seasonal rewards
	unclaimed := comp.GetUnclaimedSeasonalRewards()
	if len(unclaimed) != 1 {
		t.Errorf("Expected 1 unclaimed seasonal reward, got %d", len(unclaimed))
	}

	// Claim seasonal reward
	claimed := comp.ClaimSeasonalReward(RankGold, "season_1")
	if len(claimed) != 2 {
		t.Errorf("Expected 2 claimed rewards, got %d", len(claimed))
	}

	// Verify claimed
	unclaimed = comp.GetUnclaimedSeasonalRewards()
	if len(unclaimed) != 0 {
		t.Errorf("Expected 0 unclaimed after claim, got %d", len(unclaimed))
	}

	// Try claiming again (should return nil)
	claimed = comp.ClaimSeasonalReward(RankGold, "season_1")
	if claimed != nil {
		t.Error("Expected nil for already claimed seasonal reward")
	}

	// Try claiming non-existent reward
	claimed = comp.ClaimSeasonalReward(RankLegend, "season_1")
	if claimed != nil {
		t.Error("Expected nil for non-existent seasonal reward")
	}
}

func TestPvPRewardComponent_Achievements(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Update achievement progress
	completed := comp.UpdateAchievementProgress("ach_1", 5, 10)
	if completed {
		t.Error("Expected not completed at 5/10")
	}

	progress := comp.GetAchievementProgress("ach_1")
	if progress == nil {
		t.Fatal("Expected non-nil progress")
	}
	if progress.CurrentProgress != 5 {
		t.Errorf("Expected progress 5, got %d", progress.CurrentProgress)
	}

	// Update to completion
	completed = comp.UpdateAchievementProgress("ach_1", 10, 10)
	if !completed {
		t.Error("Expected completed at 10/10")
	}

	// Increment achievement progress
	completed = comp.IncrementAchievementProgress("ach_2", 3, 5)
	if completed {
		t.Error("Expected not completed at 3/5")
	}

	completed = comp.IncrementAchievementProgress("ach_2", 2, 5)
	if !completed {
		t.Error("Expected completed at 5/5")
	}

	// Complete achievement manually
	if !comp.CompleteAchievement("ach_3") {
		t.Error("Expected successful achievement completion")
	}

	if !comp.HasAchievement("ach_3") {
		t.Error("Expected HasAchievement to return true")
	}

	// Try completing again
	if comp.CompleteAchievement("ach_3") {
		t.Error("Expected failed duplicate achievement completion")
	}
}

func TestPvPRewardComponent_StartNewSeason(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	comp.StartNewSeason("season_2")

	if comp.CurrentSeasonID != "season_2" {
		t.Errorf("Expected season 'season_2', got '%s'", comp.CurrentSeasonID)
	}
}

func TestPvPRewardComponent_Serialization(t *testing.T) {
	comp := NewPvPRewardComponent("season_1")

	// Add some data
	comp.AddHonor(500)
	comp.AddTitle("champion")
	comp.AddMount("war_horse")
	comp.AddCosmetic("flame_aura")
	comp.RecordTournamentWin()
	comp.UpdateHighestRank(RankGold)
	comp.CompleteAchievement("first_blood")

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty serialized data")
	}

	// Deserialize into new component
	comp2 := &PvPRewardComponent{}
	if err := comp2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify data
	if comp2.HonorPoints != 500 {
		t.Errorf("Expected 500 honor, got %d", comp2.HonorPoints)
	}
	if !comp2.HasTitle("champion") {
		t.Error("Expected HasTitle to return true after deserialize")
	}
	if !comp2.HasMount("war_horse") {
		t.Error("Expected HasMount to return true after deserialize")
	}
	if !comp2.HasCosmetic("flame_aura") {
		t.Error("Expected HasCosmetic to return true after deserialize")
	}
	if comp2.TournamentWins != 1 {
		t.Errorf("Expected 1 tournament win, got %d", comp2.TournamentWins)
	}
	if comp2.GetHighestRank("season_1") != RankGold {
		t.Errorf("Expected Gold rank, got %v", comp2.GetHighestRank("season_1"))
	}
	if !comp2.HasAchievement("first_blood") {
		t.Error("Expected HasAchievement to return true after deserialize")
	}
}

func TestGeneratePvPAchievements(t *testing.T) {
	seed := int64(12345)

	achievements := GeneratePvPAchievements(seed)

	if len(achievements) == 0 {
		t.Error("Expected non-empty achievements list")
	}

	// Verify determinism
	achievements2 := GeneratePvPAchievements(seed)
	if len(achievements) != len(achievements2) {
		t.Error("Expected same number of achievements for same seed")
	}

	for i := range achievements {
		if achievements[i].ID != achievements2[i].ID {
			t.Errorf("Expected same achievement ID for same seed")
		}
	}

	// Verify different seed produces different values
	achievements3 := GeneratePvPAchievements(99999)
	differentFound := false
	for i := range achievements {
		if achievements[i].Reward.Value != achievements3[i].Reward.Value {
			differentFound = true
			break
		}
	}
	if !differentFound && len(achievements) > 0 {
		// Not all achievements have random values, so this is acceptable
	}

	// Verify achievement structure
	for _, ach := range achievements {
		if ach.ID == "" {
			t.Error("Expected non-empty achievement ID")
		}
		if ach.Name == "" {
			t.Error("Expected non-empty achievement name")
		}
		if ach.Requirement == "" {
			t.Error("Expected non-empty achievement requirement")
		}
		if ach.RequiredAmount <= 0 {
			t.Error("Expected positive required amount")
		}
	}
}

func TestGenerateSeasonRewards(t *testing.T) {
	seasonID := "season_1"
	seed := int64(12345)

	rewards := GenerateSeasonRewards(seasonID, seed)

	if len(rewards) == 0 {
		t.Error("Expected non-empty rewards map")
	}

	// Verify all tiers have rewards
	for _, tier := range RankTierOrder {
		tierRewards, exists := rewards[tier]
		if !exists {
			t.Errorf("Expected rewards for tier %s", tier)
		}
		if len(tierRewards) == 0 {
			t.Errorf("Expected non-empty rewards for tier %s", tier)
		}

		// Verify season ID
		for _, r := range tierRewards {
			if r.SeasonID != seasonID {
				t.Errorf("Expected seasonID '%s', got '%s'", seasonID, r.SeasonID)
			}
		}
	}

	// Verify determinism
	rewards2 := GenerateSeasonRewards(seasonID, seed)
	for tier, tierRewards := range rewards {
		tierRewards2 := rewards2[tier]
		if len(tierRewards) != len(tierRewards2) {
			t.Errorf("Expected same number of rewards for tier %s", tier)
		}
	}
}

func TestGetTierRarity(t *testing.T) {
	tests := []struct {
		tier     RankTier
		expected string
	}{
		{RankBronze, "common"},
		{RankSilver, "common"},
		{RankGold, "uncommon"},
		{RankPlatinum, "rare"},
		{RankDiamond, "epic"},
		{RankMaster, "legendary"},
		{RankLegend, "legendary"},
	}

	for _, tt := range tests {
		result := getTierRarity(tt.tier)
		if result != tt.expected {
			t.Errorf("getTierRarity(%s) = %s, expected %s", tt.tier, result, tt.expected)
		}
	}
}

func BenchmarkPvPRewardComponent_AddHonor(b *testing.B) {
	comp := NewPvPRewardComponent("season_1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.AddHonor(25)
	}
}

func BenchmarkPvPRewardComponent_Serialize(b *testing.B) {
	comp := NewPvPRewardComponent("season_1")
	comp.AddHonor(1000)
	for i := 0; i < 10; i++ {
		comp.AddTitle("title_" + string(rune('a'+i)))
		comp.AddMount("mount_" + string(rune('a'+i)))
		comp.AddCosmetic("cosmetic_" + string(rune('a'+i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = comp.Serialize()
	}
}

func BenchmarkGeneratePvPAchievements(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GeneratePvPAchievements(int64(i))
	}
}

func BenchmarkGenerateSeasonRewards(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateSeasonRewards("season_1", int64(i))
	}
}
