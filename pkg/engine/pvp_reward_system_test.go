// Package engine provides tests for the PvP reward system.
package engine

import (
	"testing"
)

func TestNewPvPRewardSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	sys := NewPvPRewardSystem(world, seed)

	if sys == nil {
		t.Fatal("Expected non-nil system")
	}

	if sys.world != world {
		t.Error("Expected world to be set")
	}

	if len(sys.achievements) == 0 {
		t.Error("Expected non-empty achievements")
	}

	if len(sys.vendorItems) == 0 {
		t.Error("Expected non-empty vendor items")
	}

	// Verify determinism
	sys2 := NewPvPRewardSystem(world, seed)
	if len(sys.achievements) != len(sys2.achievements) {
		t.Error("Expected same achievements for same seed")
	}
	if len(sys.vendorItems) != len(sys2.vendorItems) {
		t.Error("Expected same vendor items for same seed")
	}
}

func TestPvPRewardSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	// Create entity with PvP components
	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))
	entity.AddComponent(NewPvPRatingComponent("season_1"))

	// Get components
	rewardComp, _ := entity.GetComponent("pvp_reward")
	pvpReward := rewardComp.(*PvPRewardComponent)

	ratingComp, _ := entity.GetComponent("pvp_rating")
	pvpRating := ratingComp.(*PvPRatingComponent)

	// Add wins to trigger achievement
	pvpRating.Wins = 1

	// Run update
	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Check if "first_blood" achievement was granted
	if !pvpReward.HasAchievement("pvp_first_blood") {
		t.Error("Expected 'pvp_first_blood' achievement after 1 win")
	}
}

func TestPvPRewardSystem_GrantMatchReward(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))
	entity.AddComponent(NewPvPRatingComponent("season_1"))

	rewardComp, _ := entity.GetComponent("pvp_reward")
	pvpReward := rewardComp.(*PvPRewardComponent)

	// Grant win reward
	sys.GrantMatchReward(entity, true, 15, 1)

	// Base win reward is 25
	if pvpReward.HonorPoints < 25 {
		t.Errorf("Expected at least 25 honor for win, got %d", pvpReward.HonorPoints)
	}

	// Reset and test loss
	pvpReward.HonorPoints = 0
	sys.GrantMatchReward(entity, false, -15, -1)

	// Base loss reward is 5
	if pvpReward.HonorPoints != 5 {
		t.Errorf("Expected 5 honor for loss, got %d", pvpReward.HonorPoints)
	}
}

func TestPvPRewardSystem_GrantMatchReward_Streak(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))

	rewardComp, _ := entity.GetComponent("pvp_reward")
	pvpReward := rewardComp.(*PvPRewardComponent)

	// Grant win with 3-win streak (should get streak bonus)
	sys.GrantMatchReward(entity, true, 15, 3)

	// Base 25 + (streak-1) * 5 = 25 + 10 = 35
	expected := 25 + (3-1)*5
	if pvpReward.HonorPoints != expected {
		t.Errorf("Expected %d honor with 3 streak, got %d", expected, pvpReward.HonorPoints)
	}
}

func TestPvPRewardSystem_GrantMatchReward_HighRating(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))

	ratingComp := NewPvPRatingComponent("season_1")
	ratingComp.Rating = 1700 // Above threshold
	entity.AddComponent(ratingComp)

	rewardComp, _ := entity.GetComponent("pvp_reward")
	pvpReward := rewardComp.(*PvPRewardComponent)

	// Grant win reward
	sys.GrantMatchReward(entity, true, 15, 1)

	// Base 25 * 1.5 = 37.5 -> 37
	baseHonor := float64(25)
	expected := int(baseHonor * 1.5)
	if pvpReward.HonorPoints != expected {
		t.Errorf("Expected %d honor with high rating, got %d", expected, pvpReward.HonorPoints)
	}
}

func TestPvPRewardSystem_GrantTournamentReward(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))

	rewardComp, _ := entity.GetComponent("pvp_reward")
	pvpReward := rewardComp.(*PvPRewardComponent)

	// Grant tournament win
	sys.GrantTournamentReward(entity, 1, 16, true)

	// Should include: participation (50) + placement (4*100) + win (500) = 950
	expected := 50 + 4*100 + 500
	if pvpReward.HonorPoints != expected {
		t.Errorf("Expected %d honor for tournament win, got %d", expected, pvpReward.HonorPoints)
	}

	if pvpReward.TournamentWins != 1 {
		t.Errorf("Expected 1 tournament win, got %d", pvpReward.TournamentWins)
	}

	if pvpReward.TournamentParticipations != 1 {
		t.Errorf("Expected 1 tournament participation, got %d", pvpReward.TournamentParticipations)
	}

	// Reset and test 4th place
	pvpReward.HonorPoints = 0
	sys.GrantTournamentReward(entity, 4, 16, false)

	// participation (50) + placement (1*100) = 150
	expected = 50 + 100
	if pvpReward.HonorPoints != expected {
		t.Errorf("Expected %d honor for 4th place, got %d", expected, pvpReward.HonorPoints)
	}

	// Reset and test 5th place (no placement bonus)
	pvpReward.HonorPoints = 0
	sys.GrantTournamentReward(entity, 5, 16, false)

	if pvpReward.HonorPoints != 50 {
		t.Errorf("Expected 50 honor for 5th place, got %d", pvpReward.HonorPoints)
	}
}

func TestPvPRewardSystem_DistributeSeasonRewards(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.UpdateHighestRank(RankGold)
	entity.AddComponent(rewardComp)

	entities := []*Entity{entity}
	sys.DistributeSeasonRewards(entities, "season_1", 12345)

	// Should have rewards for Bronze, Silver, and Gold
	unclaimed := rewardComp.GetUnclaimedSeasonalRewards()
	if len(unclaimed) != 3 {
		t.Errorf("Expected 3 seasonal reward tiers, got %d", len(unclaimed))
	}
}

func TestPvPRewardSystem_PurchaseFromVendor(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.AddHonor(1000)
	entity.AddComponent(rewardComp)
	entity.AddComponent(NewPvPRatingComponent("season_1"))

	// Get a basic vendor item
	vendorItems := sys.GetVendorInventory()
	if len(vendorItems) == 0 {
		t.Fatal("Expected non-empty vendor inventory")
	}

	// Find a basic item with no rank requirement
	var basicItem *PvPVendorItem
	for i := range vendorItems {
		if vendorItems[i].RankRequirement == "" {
			basicItem = &vendorItems[i]
			break
		}
	}

	if basicItem == nil {
		t.Fatal("Expected at least one item with no rank requirement")
	}

	initialHonor := rewardComp.HonorPoints

	// Purchase item
	if !sys.PurchaseFromVendor(entity, basicItem.ID) {
		t.Error("Expected successful purchase")
	}

	// Verify honor spent
	expectedHonor := initialHonor - basicItem.HonorCost
	if rewardComp.HonorPoints != expectedHonor {
		t.Errorf("Expected %d honor after purchase, got %d", expectedHonor, rewardComp.HonorPoints)
	}

	// Verify reward added
	if len(rewardComp.EarnedRewards) == 0 {
		t.Error("Expected reward to be added after purchase")
	}
}

func TestPvPRewardSystem_PurchaseFromVendor_InsufficientHonor(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.AddHonor(10) // Not enough for most items
	entity.AddComponent(rewardComp)

	// Find expensive item
	vendorItems := sys.GetVendorInventory()
	var expensiveItem *PvPVendorItem
	for i := range vendorItems {
		if vendorItems[i].HonorCost > 10 {
			expensiveItem = &vendorItems[i]
			break
		}
	}

	if expensiveItem == nil {
		t.Skip("No expensive items found")
	}

	// Try to purchase
	if sys.PurchaseFromVendor(entity, expensiveItem.ID) {
		t.Error("Expected failed purchase due to insufficient honor")
	}
}

func TestPvPRewardSystem_PurchaseFromVendor_RankRequirement(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.AddHonor(50000) // Plenty of honor
	entity.AddComponent(rewardComp)

	ratingComp := NewPvPRatingComponent("season_1")
	ratingComp.RankTier = RankBronze // Low rank
	entity.AddComponent(ratingComp)

	// Find item with high rank requirement
	vendorItems := sys.GetVendorInventory()
	var highRankItem *PvPVendorItem
	for i := range vendorItems {
		if vendorItems[i].RankRequirement == RankLegend {
			highRankItem = &vendorItems[i]
			break
		}
	}

	if highRankItem == nil {
		t.Skip("No Legend-tier items found")
	}

	// Try to purchase
	if sys.PurchaseFromVendor(entity, highRankItem.ID) {
		t.Error("Expected failed purchase due to rank requirement")
	}
}

func TestPvPRewardSystem_GetVendorItemsForRank(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	// Get items for Bronze rank
	bronzeItems := sys.GetVendorItemsForRank(RankBronze)

	// Get items for Legend rank
	legendItems := sys.GetVendorItemsForRank(RankLegend)

	// Legend should have access to more items
	if len(legendItems) < len(bronzeItems) {
		t.Errorf("Expected Legend to have at least as many items as Bronze")
	}

	// All bronze items should be accessible to Legend
	for _, bi := range bronzeItems {
		found := false
		for _, li := range legendItems {
			if li.ID == bi.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected Legend to have access to item %s", bi.ID)
		}
	}
}

func TestPvPRewardSystem_GetPlayerPvPStats(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.AddHonor(500)
	rewardComp.RecordTournamentWin()
	rewardComp.AddTitle("champion")
	entity.AddComponent(rewardComp)

	ratingComp := NewPvPRatingComponent("season_1")
	ratingComp.Rating = 1500
	ratingComp.Wins = 25
	ratingComp.Losses = 10
	entity.AddComponent(ratingComp)

	stats := sys.GetPlayerPvPStats(entity)

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if stats["honor_points"].(int) != 500 {
		t.Errorf("Expected 500 honor, got %v", stats["honor_points"])
	}

	if stats["tournament_wins"].(int) != 1 {
		t.Errorf("Expected 1 tournament win, got %v", stats["tournament_wins"])
	}

	if stats["titles_earned"].(int) != 1 {
		t.Errorf("Expected 1 title, got %v", stats["titles_earned"])
	}

	if stats["current_rating"].(int) != 1500 {
		t.Errorf("Expected 1500 rating, got %v", stats["current_rating"])
	}

	if stats["wins"].(int) != 25 {
		t.Errorf("Expected 25 wins, got %v", stats["wins"])
	}
}

func TestPvPRewardSystem_GetPlayerAchievementProgress(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	entity.AddComponent(rewardComp)

	ratingComp := NewPvPRatingComponent("season_1")
	ratingComp.Wins = 25
	entity.AddComponent(ratingComp)

	progress := sys.GetPlayerAchievementProgress(entity)

	if len(progress) == 0 {
		t.Error("Expected non-empty achievement progress")
	}

	// Check first blood achievement (requires 1 win)
	firstBlood, exists := progress["pvp_first_blood"]
	if !exists {
		t.Error("Expected 'pvp_first_blood' in progress")
	} else {
		if firstBlood.CurrentProgress != 25 {
			t.Errorf("Expected 25 current progress for wins-based achievement, got %d", firstBlood.CurrentProgress)
		}
		if !firstBlood.Completed {
			t.Error("Expected 'pvp_first_blood' to be completed with 25 wins")
		}
	}
}

func TestPvPRewardSystem_HonorConfig(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	// Get default config
	config := sys.GetHonorConfig()

	if config.MatchWin != 25 {
		t.Errorf("Expected default MatchWin of 25, got %d", config.MatchWin)
	}

	// Set custom config
	customConfig := HonorRewardConfig{
		MatchWin:                50,
		MatchLoss:               10,
		TournamentWin:           1000,
		TournamentParticipation: 100,
		TopPlacement:            200,
		WinStreakBonus:          10,
		RatingBonusThreshold:    1800,
		HighRatingMultiplier:    2.0,
	}

	sys.SetHonorConfig(customConfig)

	// Verify new config
	config = sys.GetHonorConfig()
	if config.MatchWin != 50 {
		t.Errorf("Expected custom MatchWin of 50, got %d", config.MatchWin)
	}

	// Test with new config
	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))

	rewardComp, _ := entity.GetComponent("pvp_reward")
	pvpReward := rewardComp.(*PvPRewardComponent)

	sys.GrantMatchReward(entity, true, 15, 1)

	if pvpReward.HonorPoints != 50 {
		t.Errorf("Expected 50 honor with custom config, got %d", pvpReward.HonorPoints)
	}
}

func TestPvPRewardSystem_VendorItemStock(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	// Find an item with limited stock (any rank requirement)
	vendorItems := sys.GetVendorInventory()
	var limitedItem *PvPVendorItem
	var itemIdx int
	for i := range vendorItems {
		if vendorItems[i].Stock > 0 {
			limitedItem = &vendorItems[i]
			itemIdx = i
			break
		}
	}

	if limitedItem == nil {
		t.Skip("No limited stock items found")
	}

	initialStock := limitedItem.Stock

	// Create entity with enough honor and high rank to purchase any item
	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.AddHonor(100000)
	entity.AddComponent(rewardComp)

	// Add rating component with Legend rank to meet any rank requirement
	ratingComp := NewPvPRatingComponent("season_1")
	ratingComp.RankTier = RankLegend
	entity.AddComponent(ratingComp)

	// Purchase item
	if !sys.PurchaseFromVendor(entity, limitedItem.ID) {
		t.Error("Expected successful purchase of limited item")
	}

	// Check stock decreased
	currentStock := sys.vendorItems[itemIdx].Stock
	if currentStock != initialStock-1 {
		t.Errorf("Expected stock to decrease from %d to %d, got %d", initialStock, initialStock-1, currentStock)
	}
}

func TestPvPRewardSystem_AchievementGrant_Rating(t *testing.T) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	entity.AddComponent(rewardComp)

	ratingComp := NewPvPRatingComponent("season_1")
	ratingComp.Rating = RankThreshold[RankGold]
	ratingComp.PeakRating = RankThreshold[RankGold]
	entity.AddComponent(ratingComp)

	// Run update to process achievements
	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Check rising star achievement
	if !rewardComp.HasAchievement("pvp_rising_star") {
		t.Error("Expected 'pvp_rising_star' achievement at Gold rank")
	}
}

func TestDefaultHonorConfig(t *testing.T) {
	config := DefaultHonorConfig()

	if config.MatchWin <= 0 {
		t.Error("Expected positive MatchWin")
	}
	if config.MatchLoss < 0 {
		t.Error("Expected non-negative MatchLoss")
	}
	if config.TournamentWin <= 0 {
		t.Error("Expected positive TournamentWin")
	}
	if config.HighRatingMultiplier <= 0 {
		t.Error("Expected positive HighRatingMultiplier")
	}
}

func BenchmarkPvPRewardSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	// Create entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewPvPRewardComponent("season_1"))
		entity.AddComponent(NewPvPRatingComponent("season_1"))
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkPvPRewardSystem_GrantMatchReward(b *testing.B) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewPvPRewardComponent("season_1"))
	entity.AddComponent(NewPvPRatingComponent("season_1"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GrantMatchReward(entity, true, 15, 1)
	}
}

func BenchmarkPvPRewardSystem_PurchaseFromVendor(b *testing.B) {
	world := NewWorld()
	sys := NewPvPRewardSystem(world, 12345)

	entity := world.CreateEntity()
	rewardComp := NewPvPRewardComponent("season_1")
	rewardComp.AddHonor(1000000)
	entity.AddComponent(rewardComp)

	vendorItems := sys.GetVendorInventory()
	itemID := vendorItems[0].ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewardComp.AddHonor(100) // Ensure we have enough
		sys.PurchaseFromVendor(entity, itemID)
	}
}
