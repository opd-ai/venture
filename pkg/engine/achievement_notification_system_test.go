// Package engine provides tests for the achievement notification system.
//
// Phase 85: Achievement Notifications & Rewards (V15.0)
package engine

import (
	"sync"
	"testing"
	"time"
)

// mockAchievementClock implements GameClock for testing.
type mockAchievementClock struct {
	currentTime time.Time
}

func (c *mockAchievementClock) Now() time.Time {
	return c.currentTime
}

func (c *mockAchievementClock) Advance(deltaTime float64) {
	c.currentTime = c.currentTime.Add(time.Duration(deltaTime * float64(time.Second)))
}

func (c *mockAchievementClock) Reset(startTime time.Time) {
	c.currentTime = startTime
}

func (c *mockAchievementClock) SetTime(t time.Time) {
	c.currentTime = t
}

func createAchievementNotificationTestWorld() (*World, *mockAchievementClock) {
	world := NewWorld()
	clock := &mockAchievementClock{currentTime: time.Unix(1000, 0)}
	world.Clock = clock
	return world, clock
}

// commitEntities processes pending entity additions
func commitEntities(world *World) {
	world.Update(0.0)
}

func TestNewAchievementNotificationSystem(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	if system == nil {
		t.Fatal("System should not be nil")
	}
	if system.world != world {
		t.Error("World reference mismatch")
	}
	if system.customRewards == nil {
		t.Error("customRewards should be initialized")
	}
}

func TestNewAchievementNotificationSystemNilWorld(t *testing.T) {
	system := NewAchievementNotificationSystem(nil)

	if system == nil {
		t.Fatal("System should not be nil even with nil world")
	}
}

func TestHandleAchievementUnlock(t *testing.T) {
	world, clock := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world) // Process entity additions

	// Unlock achievement
	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze)

	// Check notification was queued
	pending := system.GetPendingNotifications(player.ID)
	if pending != 1 {
		t.Errorf("Expected 1 pending notification, got %d", pending)
	}

	// Pop and verify notification
	notification := system.PopNotification(player.ID)
	if notification == nil {
		t.Fatal("Expected notification, got nil")
	}
	if notification.AchievementID != "combat_first_blood" {
		t.Errorf("Expected 'combat_first_blood', got '%s'", notification.AchievementID)
	}
	if notification.AchievementName != "First Blood" {
		t.Errorf("Expected 'First Blood', got '%s'", notification.AchievementName)
	}
	if notification.Tier != AchievementTierBronze {
		t.Errorf("Expected Bronze tier, got %s", notification.Tier.String())
	}
	if notification.Points != 10 {
		t.Errorf("Expected 10 points, got %d", notification.Points)
	}
	if notification.Timestamp != clock.currentTime.Unix() {
		t.Errorf("Timestamp mismatch")
	}
	if len(notification.Rewards) == 0 {
		t.Error("Expected rewards")
	}
}

func TestHandleAchievementUnlockWithRewards(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	player.AddComponent(NewExperienceComponent())
	commitEntities(world)

	// Track rewards via callback
	var grantedRewards []AchievementReward
	system.SetOnRewardGrantedCallback(func(entityID uint64, reward AchievementReward) {
		grantedRewards = append(grantedRewards, reward)
	})

	// Unlock Gold tier (should give XP and Currency)
	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierGold)

	// Verify rewards were granted
	if len(grantedRewards) != 2 {
		t.Errorf("Expected 2 rewards, got %d", len(grantedRewards))
	}

	hasXP := false
	hasCurrency := false
	for _, r := range grantedRewards {
		if r.Type == AchievementRewardXP {
			hasXP = true
			if r.Value != 300 {
				t.Errorf("Expected 300 XP, got %d", r.Value)
			}
		}
		if r.Type == AchievementRewardCurrency {
			hasCurrency = true
			if r.Value != 75 {
				t.Errorf("Expected 75 gold, got %d", r.Value)
			}
		}
	}
	if !hasXP {
		t.Error("Expected XP reward")
	}
	if !hasCurrency {
		t.Error("Expected currency reward")
	}
}

func TestHandleAchievementUnlockPlatinum(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	// Unlock Platinum tier (should give XP, Currency, and Title)
	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierPlatinum)

	notification := system.PopNotification(player.ID)
	if notification == nil {
		t.Fatal("Expected notification")
	}
	if len(notification.Rewards) != 3 {
		t.Errorf("Expected 3 rewards for Platinum, got %d", len(notification.Rewards))
	}

	hasTitle := false
	for _, r := range notification.Rewards {
		if r.Type == AchievementRewardTitle {
			hasTitle = true
		}
	}
	if !hasTitle {
		t.Error("Platinum tier should include title reward")
	}
}

func TestHandleAchievementUnlockUnknownAchievement(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	commitEntities(world)

	// Try to unlock unknown achievement
	system.HandleAchievementUnlock(player.ID, "unknown_achievement", AchievementTierBronze)

	// Should not queue notification
	pending := system.GetPendingNotifications(player.ID)
	if pending != 0 {
		t.Errorf("Expected 0 pending for unknown achievement, got %d", pending)
	}
}

func TestHandleAchievementUnlockNonexistentEntity(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	// Should not panic
	system.HandleAchievementUnlock(999999, "combat_first_blood", AchievementTierBronze)
}

func TestRegisterCustomRewards(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	// Register custom rewards
	customDef := AchievementRewardDefinition{
		AchievementID: "combat_first_blood",
		TierRewards: map[AchievementTier][]AchievementReward{
			AchievementTierBronze: {
				{Type: AchievementRewardXP, Name: "Custom XP", Value: 999},
			},
		},
	}
	system.RegisterCustomRewards(customDef)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze)

	notification := system.PopNotification(player.ID)
	if notification == nil {
		t.Fatal("Expected notification")
	}
	if len(notification.Rewards) != 1 {
		t.Fatalf("Expected 1 custom reward, got %d", len(notification.Rewards))
	}
	if notification.Rewards[0].Value != 999 {
		t.Errorf("Expected custom reward value 999, got %d", notification.Rewards[0].Value)
	}
}

func TestSetRewardSeed(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	// Set custom seed
	system.SetRewardSeed(54321)

	// Register custom item reward
	customDef := AchievementRewardDefinition{
		AchievementID: "combat_first_blood",
		TierRewards: map[AchievementTier][]AchievementReward{
			AchievementTierBronze: {
				{Type: AchievementRewardItem, Name: "Test Item", Value: 1},
			},
		},
	}
	system.RegisterCustomRewards(customDef)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze)

	notification := system.PopNotification(player.ID)
	if notification == nil {
		t.Fatal("Expected notification")
	}
	if len(notification.Rewards) == 0 {
		t.Fatal("Expected rewards")
	}
	if notification.Rewards[0].ItemSeed == 0 {
		t.Error("Item reward should have generated seed")
	}
}

func TestOnPlayUnlockSoundCallback(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	var soundPlayed bool
	var soundEntityID uint64
	var soundTier AchievementTier

	system.SetOnPlayUnlockSoundCallback(func(entityID uint64, tier AchievementTier) {
		soundPlayed = true
		soundEntityID = entityID
		soundTier = tier
	})

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierSilver)

	if !soundPlayed {
		t.Error("Sound callback should have been called")
	}
	if soundEntityID != player.ID {
		t.Errorf("Entity ID mismatch: %d vs %d", soundEntityID, player.ID)
	}
	if soundTier != AchievementTierSilver {
		t.Errorf("Tier mismatch: %s", soundTier.String())
	}
}

func TestOnPlayUnlockSoundDisabled(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	soundPlayed := false
	system.SetOnPlayUnlockSoundCallback(func(entityID uint64, tier AchievementTier) {
		soundPlayed = true
	})

	player := world.CreateEntity()
	notifComp := NewAchievementNotificationComponent()
	notifComp.SetPlaySound(false)
	player.AddComponent(notifComp)
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze)

	if soundPlayed {
		t.Error("Sound should not play when disabled")
	}
}

func TestGetTotalAchievementPoints(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	// Initial points
	points := system.GetTotalAchievementPoints(player.ID)
	if points != 0 {
		t.Errorf("Expected 0 initial points, got %d", points)
	}

	// Unlock achievements
	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze) // 10 points
	system.HandleAchievementUnlock(player.ID, "quest_adventurer", AchievementTierSilver)   // 25 points

	points = system.GetTotalAchievementPoints(player.ID)
	if points != 35 {
		t.Errorf("Expected 35 points, got %d", points)
	}
}

func TestGetPendingNotificationsNilWorld(t *testing.T) {
	system := NewAchievementNotificationSystem(nil)

	pending := system.GetPendingNotifications(1)
	if pending != 0 {
		t.Errorf("Expected 0 for nil world, got %d", pending)
	}
}

func TestPopNotificationNilWorld(t *testing.T) {
	system := NewAchievementNotificationSystem(nil)

	notification := system.PopNotification(1)
	if notification != nil {
		t.Error("Expected nil for nil world")
	}
}

func TestGetDefaultTierRewards(t *testing.T) {
	tests := []struct {
		tier          AchievementTier
		expectedCount int
	}{
		{AchievementTierNone, 0},
		{AchievementTierBronze, 1},
		{AchievementTierSilver, 2},
		{AchievementTierGold, 2},
		{AchievementTierPlatinum, 3},
	}

	for _, tc := range tests {
		rewards := GetDefaultTierRewards(tc.tier)
		if len(rewards) != tc.expectedCount {
			t.Errorf("Tier %s: expected %d rewards, got %d", tc.tier.String(), tc.expectedCount, len(rewards))
		}
	}
}

func TestGrantXPIntegration(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	xpComp := NewExperienceComponent()
	player.AddComponent(xpComp)
	commitEntities(world)

	initialXP := xpComp.CurrentXP

	// Unlock Bronze (50 XP)
	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze)

	if xpComp.CurrentXP != initialXP+50 {
		t.Errorf("Expected %d XP, got %d", initialXP+50, xpComp.CurrentXP)
	}
}

func TestGrantCurrencyIntegration(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	invComp := NewInventoryComponent(20, 100.0)
	invComp.Gold = 100
	player.AddComponent(invComp)
	commitEntities(world)

	// Unlock Silver (25 gold)
	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierSilver)

	if invComp.Gold != 125 {
		t.Errorf("Expected 125 gold, got %d", invComp.Gold)
	}
}

func TestUpdateMethod(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	entities := []*Entity{player}

	// Should not panic
	system.Update(entities, 0.016)
}

func TestAchievementNotificationSystemConcurrency(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	var wg sync.WaitGroup
	done := make(chan bool)

	// Create player with notification component already attached
	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	player.AddComponent(NewAchievementNotificationComponent())
	commitEntities(world)

	// Concurrent unlocks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(tier AchievementTier) {
			defer wg.Done()
			system.HandleAchievementUnlock(player.ID, "combat_first_blood", tier)
		}(AchievementTier((i % 4) + 1))
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = system.GetPendingNotifications(player.ID)
				_ = system.GetTotalAchievementPoints(player.ID)
			}
		}()
	}

	// Concurrent pops
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				system.PopNotification(player.ID)
			}
		}()
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}
}

func TestHandleAchievementUnlockStatisticsUpdate(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	statsComp := NewPlayerStatisticsComponent()
	player.AddComponent(statsComp)
	commitEntities(world)

	system.HandleAchievementUnlock(player.ID, "combat_first_blood", AchievementTierBronze)

	unlocked := statsComp.GetLifetimeStat("general_achievements_unlocked")
	if unlocked != 1 {
		t.Errorf("Expected 1 achievement unlocked stat, got %d", unlocked)
	}
}

func TestHandleAchievementUnlockNilWorld(t *testing.T) {
	system := NewAchievementNotificationSystem(nil)

	// Should not panic
	system.HandleAchievementUnlock(1, "combat_first_blood", AchievementTierBronze)
}

func TestMultipleAchievementUnlocks(t *testing.T) {
	world, _ := createAchievementNotificationTestWorld()
	system := NewAchievementNotificationSystem(world)

	player := world.CreateEntity()
	player.AddComponent(NewExtendedAchievementComponent())
	commitEntities(world)

	achievements := []string{
		"combat_first_blood",
		"quest_adventurer",
		"craft_apprentice",
		"explore_wanderer",
		"social_friend_maker",
		"pvp_gladiator",
	}

	for _, achID := range achievements {
		system.HandleAchievementUnlock(player.ID, achID, AchievementTierBronze)
	}

	pending := system.GetPendingNotifications(player.ID)
	if pending != 6 {
		t.Errorf("Expected 6 pending notifications, got %d", pending)
	}

	// Each Bronze = 10 points
	totalPoints := system.GetTotalAchievementPoints(player.ID)
	if totalPoints != 60 {
		t.Errorf("Expected 60 total points, got %d", totalPoints)
	}
}
