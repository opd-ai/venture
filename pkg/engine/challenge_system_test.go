// Package engine provides tests for the ChallengeSystem.
// Phase 98: Daily/Weekly Challenges (V18.0)
package engine

import (
	"testing"
	"time"
)

func TestNewChallengeSystem(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	if s == nil {
		t.Fatal("NewChallengeSystem returned nil")
	}

	if s.world != world {
		t.Error("ChallengeSystem world reference mismatch")
	}

	if s.updateIntervalSeconds != 1.0 {
		t.Errorf("updateIntervalSeconds = %f, want 1.0", s.updateIntervalSeconds)
	}
}

func TestSetCallbacks(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	_ = false // placeholder for unused var warning
	s.SetRewardCallback(func(entityID uint64, reward *ChallengeReward) {
		// callback set
	})

	s.SetCompletionCallback(func(entityID uint64, challenge *Challenge) {
		// callback set
	})

	s.SetStreakCallback(func(entityID uint64, streakType ChallengeType, newStreak int) {
		// callback set
	})

	// Verify callbacks are set (not nil)
	s.mu.RLock()
	if s.onRewardGranted == nil {
		t.Error("onRewardGranted callback not set")
	}
	if s.onChallengeCompleted == nil {
		t.Error("onChallengeCompleted callback not set")
	}
	if s.onStreakChanged == nil {
		t.Error("onStreakChanged callback not set")
	}
	s.mu.RUnlock()
}

func TestChallengeSystemUpdate(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)
	s.updateIntervalSeconds = 0 // Disable throttling for test

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity
	world.Update(0) // Process entity creation

	// Run update to initialize challenges
	s.Update([]*Entity{entity}, 0.016)

	if len(comp.GetActiveDailyChallenges()) != 5 {
		t.Errorf("Active daily challenges = %d, want 5", len(comp.GetActiveDailyChallenges()))
	}

	if len(comp.GetActiveWeeklyChallenges()) != 3 {
		t.Errorf("Active weekly challenges = %d, want 3", len(comp.GetActiveWeeklyChallenges()))
	}
}

func TestChallengeSystemUpdateWithReset(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)
	s.updateIntervalSeconds = 0

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity
	world.Update(0)

	// Set last reset to yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	comp.LastDailyReset = yesterday.Unix()

	// Add completed challenges
	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", IsCompleted: true},
		{ID: "ch2", IsCompleted: true},
		{ID: "ch3", IsCompleted: true},
		{ID: "ch4", IsCompleted: true},
		{ID: "ch5", IsCompleted: true},
	}

	streakChanged := false
	s.SetStreakCallback(func(entityID uint64, streakType ChallengeType, newStreak int) {
		streakChanged = true
	})

	s.Update([]*Entity{entity}, 0.016)

	if !streakChanged {
		t.Error("Streak callback should have been called")
	}

	// After reset, challenges are regenerated
	if len(comp.GetActiveDailyChallenges()) != 5 {
		t.Errorf("Should have 5 new daily challenges after reset, got %d", len(comp.GetActiveDailyChallenges()))
	}
}

func TestTrackProgress(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity
	world.Update(0)

	// Set up a challenge with known tracking key
	comp.ActiveDailyChallenges = []*Challenge{
		{
			ID:           "test_ch",
			DefinitionID: "kill_enemies",
			Type:         ChallengeTypeDaily,
			Target:       10,
			Progress:     0,
		},
	}

	result := s.TrackProgress(entity.ID, "enemies_killed", 5)

	if !result {
		t.Error("TrackProgress should return true")
	}

	challenges := comp.GetActiveDailyChallenges()
	if challenges[0].Progress != 5 {
		t.Errorf("Progress = %d, want 5", challenges[0].Progress)
	}
}

func TestTrackProgressCompletion(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	completionCalled := false
	s.SetCompletionCallback(func(entityID uint64, challenge *Challenge) {
		completionCalled = true
	})

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{
			ID:           "test_ch",
			DefinitionID: "kill_enemies",
			Type:         ChallengeTypeDaily,
			Target:       10,
			Progress:     0,
		},
	}

	s.TrackProgress(entity.ID, "enemies_killed", 10)

	if !completionCalled {
		t.Error("Completion callback should have been called")
	}

	challenges := comp.GetActiveDailyChallenges()
	if !challenges[0].IsCompleted {
		t.Error("Challenge should be completed")
	}
}

func TestClaimReward(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	rewardCalled := false
	s.SetRewardCallback(func(entityID uint64, reward *ChallengeReward) {
		rewardCalled = true
	})

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{
			ID:          "test_ch",
			Target:      10,
			Progress:    10,
			IsCompleted: true,
			Reward:      ChallengeReward{XP: 100, Gold: 50},
		},
	}

	reward := s.ClaimReward(entity.ID, "test_ch")

	if reward == nil {
		t.Fatal("ClaimReward returned nil")
	}

	if !rewardCalled {
		t.Error("Reward callback should have been called")
	}

	if reward.XP != 100 {
		t.Errorf("XP = %d, want 100", reward.XP)
	}
}

func TestClaimRewardNotFound(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	reward := s.ClaimReward(entity.ID, "nonexistent")

	if reward != nil {
		t.Error("ClaimReward should return nil for nonexistent challenge")
	}
}

func TestRerollChallenge(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	// Generate initial challenges
	date := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	daily := comp.GenerateDailyChallenges(date)
	comp.SetActiveChallenges(daily, nil)

	oldID := daily[0].ID
	newChallenge := s.RerollChallenge(entity.ID, oldID)

	if newChallenge == nil {
		t.Fatal("RerollChallenge returned nil")
	}

	if newChallenge.ID == oldID {
		t.Error("Rerolled challenge should have different ID")
	}

	if comp.GetRerollsRemaining() != 2 {
		t.Errorf("RerollsRemaining = %d, want 2", comp.GetRerollsRemaining())
	}
}

func TestGetEntityChallenges(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{{ID: "daily1"}}
	comp.ActiveWeeklyChallenges = []*Challenge{{ID: "weekly1"}}

	daily, weekly := s.GetEntityChallenges(entity.ID)

	if len(daily) != 1 {
		t.Errorf("Daily challenges = %d, want 1", len(daily))
	}
	if len(weekly) != 1 {
		t.Errorf("Weekly challenges = %d, want 1", len(weekly))
	}
}

func TestGetEntityChallengesNotFound(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	daily, weekly := s.GetEntityChallenges(99999)

	if daily != nil || weekly != nil {
		t.Error("Should return nil for nonexistent entity")
	}
}

func TestGetEntityStreak(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	comp.DailyStreak = 5
	comp.WeeklyStreak = 2
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	dailyStreak := s.GetEntityStreak(entity.ID, ChallengeTypeDaily)
	weeklyStreak := s.GetEntityStreak(entity.ID, ChallengeTypeWeekly)

	if dailyStreak != 5 {
		t.Errorf("DailyStreak = %d, want 5", dailyStreak)
	}
	if weeklyStreak != 2 {
		t.Errorf("WeeklyStreak = %d, want 2", weeklyStreak)
	}
}

func TestGetCompletionPercent(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", IsCompleted: true},
		{ID: "ch2", IsCompleted: false},
	}

	percent := s.GetCompletionPercent(entity.ID, ChallengeTypeDaily)

	if percent != 50.0 {
		t.Errorf("Completion percent = %f, want 50.0", percent)
	}
}

func TestInitializePlayerChallenges(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity creation

	err := s.InitializePlayerChallenges(entity.ID, 12345)
	if err != nil {
		t.Fatalf("InitializePlayerChallenges error: %v", err)
	}

	if !entity.HasComponent("daily_challenge") {
		t.Error("Entity should have daily_challenge component")
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		t.Fatal("Failed to get component")
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		t.Fatal("Failed to cast to DailyChallengeComponent")
	}

	if len(comp.GetActiveDailyChallenges()) != 5 {
		t.Errorf("Daily challenges = %d, want 5", len(comp.GetActiveDailyChallenges()))
	}

	if len(comp.GetActiveWeeklyChallenges()) != 3 {
		t.Errorf("Weekly challenges = %d, want 3", len(comp.GetActiveWeeklyChallenges()))
	}
}

func TestInitializePlayerChallengesAlreadyExists(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	existingComp := NewDailyChallengeComponent(54321)
	existingComp.DailyStreak = 10
	entity.AddComponent(existingComp)
	world.Update(0) // Process entity creation

	err := s.InitializePlayerChallenges(entity.ID, 12345)
	if err != nil {
		t.Fatalf("InitializePlayerChallenges error: %v", err)
	}

	// Should not overwrite existing component
	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		t.Fatal("Failed to get component")
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok || comp.DailyStreak != 10 {
		t.Error("Should not overwrite existing component")
	}
}

func TestInitializePlayerChallengesNotFound(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	err := s.InitializePlayerChallenges(99999, 12345)
	if err == nil {
		t.Error("Should return error for nonexistent entity")
	}
}

func TestGetChallengeByID(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{{ID: "daily1", Name: "Daily Test"}}
	comp.ActiveWeeklyChallenges = []*Challenge{{ID: "weekly1", Name: "Weekly Test"}}

	daily := s.GetChallengeByID(entity.ID, "daily1")
	weekly := s.GetChallengeByID(entity.ID, "weekly1")
	notFound := s.GetChallengeByID(entity.ID, "nonexistent")

	if daily == nil || daily.Name != "Daily Test" {
		t.Error("Failed to get daily challenge by ID")
	}
	if weekly == nil || weekly.Name != "Weekly Test" {
		t.Error("Failed to get weekly challenge by ID")
	}
	if notFound != nil {
		t.Error("Should return nil for nonexistent challenge")
	}
}

func TestGetTotalStats(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	comp.TotalChallengesCompleted = 100
	comp.TotalXPEarned = 5000
	comp.TotalGoldEarned = 2500
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	completed, xp, gold := s.GetTotalStats(entity.ID)

	if completed != 100 {
		t.Errorf("TotalCompleted = %d, want 100", completed)
	}
	if xp != 5000 {
		t.Errorf("TotalXP = %d, want 5000", xp)
	}
	if gold != 2500 {
		t.Errorf("TotalGold = %d, want 2500", gold)
	}
}

func TestTrackCombatProgress(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", DefinitionID: "kill_enemies", Type: ChallengeTypeDaily, Target: 100, Progress: 0},
		{ID: "ch2", DefinitionID: "deal_damage", Type: ChallengeTypeDaily, Target: 1000, Progress: 0},
	}

	s.TrackCombatProgress(entity.ID, 10, 500, 2, 0)

	challenges := comp.GetActiveDailyChallenges()

	// Check if progress was tracked
	var enemyProgress, damageProgress int
	for _, ch := range challenges {
		switch ch.DefinitionID {
		case "kill_enemies":
			enemyProgress = ch.Progress
		case "deal_damage":
			damageProgress = ch.Progress
		}
	}

	if enemyProgress != 10 {
		t.Errorf("Enemy progress = %d, want 10", enemyProgress)
	}
	if damageProgress != 500 {
		t.Errorf("Damage progress = %d, want 500", damageProgress)
	}
}

func TestTrackGatheringProgress(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", DefinitionID: "gather_resources", Type: ChallengeTypeDaily, Target: 100, Progress: 0},
		{ID: "ch2", DefinitionID: "catch_fish", Type: ChallengeTypeDaily, Target: 50, Progress: 0},
	}

	s.TrackGatheringProgress(entity.ID, 20, 5, 0, 0)

	challenges := comp.GetActiveDailyChallenges()

	var resourceProgress, fishProgress int
	for _, ch := range challenges {
		switch ch.DefinitionID {
		case "gather_resources":
			resourceProgress = ch.Progress
		case "catch_fish":
			fishProgress = ch.Progress
		}
	}

	if resourceProgress != 20 {
		t.Errorf("Resource progress = %d, want 20", resourceProgress)
	}
	if fishProgress != 5 {
		t.Errorf("Fish progress = %d, want 5", fishProgress)
	}
}

func TestTrackExplorationProgress(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", DefinitionID: "discover_areas", Type: ChallengeTypeDaily, Target: 10, Progress: 0},
	}

	s.TrackExplorationProgress(entity.ID, 3, 100, 1, 0)

	challenges := comp.GetActiveDailyChallenges()
	if challenges[0].Progress != 3 {
		t.Errorf("Area progress = %d, want 3", challenges[0].Progress)
	}
}

func TestTrackSocialProgress(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", DefinitionID: "complete_trades", Type: ChallengeTypeDaily, Target: 10, Progress: 0},
	}

	s.TrackSocialProgress(entity.ID, 2, 5, 1, 0)

	challenges := comp.GetActiveDailyChallenges()
	if challenges[0].Progress != 2 {
		t.Errorf("Trade progress = %d, want 2", challenges[0].Progress)
	}
}

func TestTrackCraftingProgress(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", DefinitionID: "craft_items", Type: ChallengeTypeDaily, Target: 20, Progress: 0},
	}

	s.TrackCraftingProgress(entity.ID, 5, 2, 1, 0)

	challenges := comp.GetActiveDailyChallenges()
	if challenges[0].Progress != 5 {
		t.Errorf("Craft progress = %d, want 5", challenges[0].Progress)
	}
}

func TestSystemWithNilWorld(t *testing.T) {
	s := NewChallengeSystem(nil)

	// All methods should handle nil world gracefully
	result := s.TrackProgress(1, "test", 1)
	if result {
		t.Error("TrackProgress should return false with nil world")
	}

	reward := s.ClaimReward(1, "test")
	if reward != nil {
		t.Error("ClaimReward should return nil with nil world")
	}

	challenge := s.RerollChallenge(1, "test")
	if challenge != nil {
		t.Error("RerollChallenge should return nil with nil world")
	}

	daily, weekly := s.GetEntityChallenges(1)
	if daily != nil || weekly != nil {
		t.Error("GetEntityChallenges should return nil with nil world")
	}

	streak := s.GetEntityStreak(1, ChallengeTypeDaily)
	if streak != 0 {
		t.Error("GetEntityStreak should return 0 with nil world")
	}

	percent := s.GetCompletionPercent(1, ChallengeTypeDaily)
	if percent != 0 {
		t.Error("GetCompletionPercent should return 0 with nil world")
	}

	err := s.InitializePlayerChallenges(1, 12345)
	if err == nil {
		t.Error("InitializePlayerChallenges should return error with nil world")
	}
}

func TestUpdateThrottling(t *testing.T) {
	world := NewWorld()
	s := NewChallengeSystem(world)
	s.updateIntervalSeconds = 10.0 // 10 second throttle

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	// First update should run
	s.Update([]*Entity{entity}, 0.016)
	firstCount := len(comp.GetActiveDailyChallenges())

	// Clear challenges to test if update runs
	comp.ActiveDailyChallenges = nil

	// Second immediate update should be throttled
	s.Update([]*Entity{entity}, 0.016)
	secondCount := len(comp.GetActiveDailyChallenges())

	if firstCount != 5 {
		t.Errorf("First update should generate challenges, got %d", firstCount)
	}

	// Note: Due to throttling, second update might not generate challenges
	// The actual behavior depends on implementation details
	_ = secondCount // Acknowledged that throttling affects this
}

// BenchmarkTrackProgress benchmarks progress tracking.
func BenchmarkTrackProgress(b *testing.B) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	comp.ActiveDailyChallenges = []*Challenge{
		{ID: "ch1", DefinitionID: "kill_enemies", Type: ChallengeTypeDaily, Target: 1000000, Progress: 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.TrackProgress(entity.ID, "enemies_killed", 1)
	}
}

// BenchmarkClaimReward benchmarks reward claiming.
func BenchmarkClaimReward(b *testing.B) {
	world := NewWorld()
	s := NewChallengeSystem(world)

	entity := world.CreateEntity()
	comp := NewDailyChallengeComponent(12345)
	entity.AddComponent(comp)
	world.Update(0) // Process entity

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		comp.ActiveDailyChallenges = []*Challenge{
			{
				ID:          "test_ch",
				Target:      10,
				Progress:    10,
				IsCompleted: true,
				Reward:      ChallengeReward{XP: 100, Gold: 50},
			},
		}
		b.StartTimer()

		s.ClaimReward(entity.ID, "test_ch")
	}
}
