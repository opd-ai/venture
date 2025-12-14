package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// TestEventRewardIntegration tests the full event reward workflow.
func TestEventRewardIntegration_QuestCompletion(t *testing.T) {
	world := NewWorld()
	now := time.Date(2025, 3, 21, 12, 0, 0, 0, time.UTC) // During spring
	clock := &mockEventRewardClock{currentTime: now}

	// Create systems
	rewardSys := NewEventRewardSystem(world, clock)

	// Create world entity with seasonal events
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(12345, false)
	seasonalComp.CurrentTime = now
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:             "spring_festival",
			Name:           "Spring Festival",
			Theme:          EventThemeSpring,
			DurationDays:   7,
			StartDayOfYear: 80,
			Seed:           12345,
		},
		StartTime: now.AddDate(0, 0, -1),
		EndTime:   now.AddDate(0, 0, 6),
		Phase:     EventPhaseActive,
	})
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Create player entity with all components
	player := NewEntity(2)
	rewardComp := NewEventRewardComponent()
	questComp := NewEventQuestComponent(3)
	player.AddComponent(rewardComp)
	player.AddComponent(questComp)
	world.AddEntity(player)

	// Process pending entity additions
	world.Update(0)

	// Register participation
	rewardSys.RegisterEventParticipation(player)
	if rewardComp.TotalEventsParticipated != 1 {
		t.Errorf("TotalEventsParticipated = %d, want 1", rewardComp.TotalEventsParticipated)
	}

	// Simulate completing an event quest
	questComp.CompletedQuests = append(questComp.CompletedQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:        "spring_festival_collection",
			EventID:   "spring_festival",
			QuestType: EventQuestCollection,
			Name:      "Gather Spring Blossoms",
			Reward: quest.Reward{
				XP:   100,
				Gold: 50,
			},
		},
		Status: quest.StatusComplete,
	})

	// Run system update
	rewardSys.Update([]*Entity{player}, 0.016)

	// Check quest completion was recorded
	if rewardComp.TotalQuestsCompleted != 1 {
		t.Errorf("TotalQuestsCompleted = %d, want 1", rewardComp.TotalQuestsCompleted)
	}

	// Check currency was granted
	if rewardComp.GetCurrency("spring_festival") < 25 {
		t.Errorf("Currency = %d, should be >= 25", rewardComp.GetCurrency("spring_festival"))
	}
}

func TestEventRewardIntegration_VendorPurchase(t *testing.T) {
	world := NewWorld()
	now := time.Date(2025, 6, 21, 12, 0, 0, 0, time.UTC) // Summer
	clock := &mockEventRewardClock{currentTime: now}

	rewardSys := NewEventRewardSystem(world, clock)

	// Create world entity with summer event
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(54321, false)
	seasonalComp.CurrentTime = now
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "summer_solstice",
			Name:  "Summer Solstice",
			Theme: EventThemeSummer,
			Seed:  54321,
		},
		Phase: EventPhaseActive,
	})
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Create player with lots of currency
	player := NewEntity(2)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency("summer_solstice", 5000)
	player.AddComponent(rewardComp)
	world.AddEntity(player)

	// Process pending entity additions
	world.Update(0)

	// Get vendor inventory
	inventory := rewardSys.GetVendorInventory("summer_solstice")
	if len(inventory) == 0 {
		t.Fatal("Vendor inventory should be generated")
	}

	// Find a purchasable item
	var targetItem *EventVendorItem
	for _, item := range inventory {
		if item.Stock != 0 && item.CurrencyCost <= rewardComp.GetCurrency("summer_solstice") {
			targetItem = &item
			break
		}
	}

	if targetItem == nil {
		t.Fatal("Should have at least one purchasable item")
	}

	initialCurrency := rewardComp.GetCurrency("summer_solstice")

	// Purchase the item
	success := rewardSys.PurchaseFromVendor(player, "summer_solstice", targetItem.ID)
	if !success {
		t.Fatal("Purchase should succeed")
	}

	// Verify currency spent
	expectedCurrency := initialCurrency - targetItem.CurrencyCost
	if rewardComp.GetCurrency("summer_solstice") != expectedCurrency {
		t.Errorf("Currency = %d, want %d", rewardComp.GetCurrency("summer_solstice"), expectedCurrency)
	}

	// Verify reward received
	if len(rewardComp.EarnedRewards) != 1 {
		t.Errorf("EarnedRewards = %d, want 1", len(rewardComp.EarnedRewards))
	}
}

func TestEventRewardIntegration_AchievementUnlock(t *testing.T) {
	world := NewWorld()
	now := time.Date(2025, 9, 25, 12, 0, 0, 0, time.UTC) // Autumn
	clock := &mockEventRewardClock{currentTime: now}

	rewardSys := NewEventRewardSystem(world, clock)

	// Create world entity with autumn event
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(99999, false)
	seasonalComp.CurrentTime = now
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "autumn_harvest",
			Name:  "Autumn Harvest",
			Theme: EventThemeAutumn,
			Seed:  99999,
		},
		Phase: EventPhaseActive,
	}
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, event)
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Create player
	player := NewEntity(2)
	rewardComp := NewEventRewardComponent()
	questComp := NewEventQuestComponent(3)
	rewardComp.TotalEventsParticipated = 1 // Already participated
	player.AddComponent(rewardComp)
	player.AddComponent(questComp)
	world.AddEntity(player)

	// Process pending entity additions
	world.Update(0)

	// Run system to check for participation achievement
	rewardSys.Update([]*Entity{player}, 0.016)

	// Check if participation achievement was granted
	achievements := rewardSys.EventAchievements["autumn_harvest"]
	if len(achievements) == 0 {
		t.Fatal("Achievements should be generated")
	}

	// Find the participation achievement
	var participationAch *EventAchievementDef
	for i := range achievements {
		if achievements[i].Requirement == "participate" {
			participationAch = &achievements[i]
			break
		}
	}

	if participationAch == nil {
		t.Fatal("Should have a participation achievement")
	}

	if !rewardComp.HasAchievement(participationAch.ID) {
		t.Error("Participation achievement should be completed")
	}
}

func TestEventRewardIntegration_BossQuestAchievement(t *testing.T) {
	world := NewWorld()
	now := time.Date(2025, 12, 25, 12, 0, 0, 0, time.UTC) // Winter
	clock := &mockEventRewardClock{currentTime: now}

	rewardSys := NewEventRewardSystem(world, clock)

	// Create world entity with winter event
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(11111, false)
	seasonalComp.CurrentTime = now
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "winter_celebration",
			Name:  "Winter Celebration",
			Theme: EventThemeWinter,
			Seed:  11111,
		},
		Phase: EventPhaseActive,
	}
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, event)
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Create player who defeated the boss
	player := NewEntity(2)
	rewardComp := NewEventRewardComponent()
	questComp := NewEventQuestComponent(3)

	// Add completed boss quest
	questComp.CompletedQuests = append(questComp.CompletedQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:        "winter_celebration_boss",
			EventID:   "winter_celebration",
			QuestType: EventQuestBoss,
			Name:      "Defeat the Frost Giant",
			Reward:    quest.Reward{XP: 500, Gold: 250},
		},
		Status: quest.StatusComplete,
	})

	player.AddComponent(rewardComp)
	player.AddComponent(questComp)
	world.AddEntity(player)

	// Process pending entity additions
	world.Update(0)

	// Run system
	rewardSys.Update([]*Entity{player}, 0.016)

	// Find the champion achievement
	achievements := rewardSys.EventAchievements["winter_celebration"]
	var championAch *EventAchievementDef
	for i := range achievements {
		if achievements[i].Requirement == "defeat_boss" {
			championAch = &achievements[i]
			break
		}
	}

	if championAch == nil {
		t.Fatal("Should have a defeat_boss achievement")
	}

	if !rewardComp.HasAchievement(championAch.ID) {
		t.Error("Champion achievement should be completed after boss defeat")
	}
}

func TestEventRewardIntegration_MultipleRewardTypes(t *testing.T) {
	world := NewWorld()
	now := time.Now()
	clock := &mockEventRewardClock{currentTime: now}

	rewardSys := NewEventRewardSystem(world, clock)

	// Setup
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(12345, false)
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "test_event",
			Name:  "Test Event",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		Phase: EventPhaseActive,
	})
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Setup vendor with all reward types
	rewardSys.VendorInventory["test_event"] = []EventVendorItem{
		{ID: "item1", EventID: "test_event", Name: "Item", RewardType: EventRewardItem, CurrencyCost: 10, Stock: -1},
		{ID: "title1", EventID: "test_event", Name: "Title", RewardType: EventRewardTitle, CurrencyCost: 10, Stock: -1},
		{ID: "effect1", EventID: "test_event", Name: "Effect", RewardType: EventRewardEffect, CurrencyCost: 10, Stock: -1},
	}

	player := NewEntity(2)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency("test_event", 1000)
	player.AddComponent(rewardComp)
	world.AddEntity(player)

	// Process pending entity additions
	world.Update(0)

	// Purchase all types
	rewardSys.PurchaseFromVendor(player, "test_event", "item1")
	rewardSys.PurchaseFromVendor(player, "test_event", "title1")
	rewardSys.PurchaseFromVendor(player, "test_event", "effect1")

	// Verify
	if len(rewardComp.EarnedRewards) != 3 {
		t.Errorf("EarnedRewards = %d, want 3", len(rewardComp.EarnedRewards))
	}
	if len(rewardComp.EarnedTitles) != 1 {
		t.Errorf("EarnedTitles = %d, want 1", len(rewardComp.EarnedTitles))
	}
	if len(rewardComp.EarnedEffects) != 1 {
		t.Errorf("EarnedEffects = %d, want 1", len(rewardComp.EarnedEffects))
	}
}

func TestEventRewardIntegration_Persistence(t *testing.T) {
	// Create component with various data
	original := NewEventRewardComponent()

	// Add currency for multiple events
	original.AddCurrency("spring_festival", 500)
	original.AddCurrency("summer_solstice", 300)

	// Add rewards
	original.AddReward(EventReward{
		ID:      "reward1",
		EventID: "spring_festival",
		Type:    EventRewardItem,
		Name:    "Spring Crown",
		Rarity:  "rare",
	})

	// Add titles and effects
	original.AddTitle(EventCosmeticTitle{
		ID:          "title1",
		EventID:     "spring_festival",
		DisplayName: "Herald of Spring",
		Rarity:      "epic",
	})
	original.SetActiveTitle("title1")

	original.AddEffect(EventVisualEffect{
		ID:         "effect1",
		EventID:    "spring_festival",
		Name:       "Petal Aura",
		EffectType: "aura",
		ColorHex:   "#FF69B4",
		Intensity:  0.8,
	})
	original.SetActiveEffect("effect1")

	// Complete achievements
	original.CompleteAchievement("spring_quest_master")
	original.RecordEventParticipation()
	original.RecordQuestCompletion()
	original.RecordQuestCompletion()

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize into new component
	restored := &EventRewardComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify all data
	if restored.GetCurrency("spring_festival") != 500 {
		t.Errorf("Spring currency = %d, want 500", restored.GetCurrency("spring_festival"))
	}
	if restored.GetCurrency("summer_solstice") != 300 {
		t.Errorf("Summer currency = %d, want 300", restored.GetCurrency("summer_solstice"))
	}
	if len(restored.EarnedRewards) != 1 {
		t.Errorf("EarnedRewards = %d, want 1", len(restored.EarnedRewards))
	}
	if len(restored.EarnedTitles) != 1 {
		t.Errorf("EarnedTitles = %d, want 1", len(restored.EarnedTitles))
	}
	if restored.ActiveTitle != "title1" {
		t.Errorf("ActiveTitle = %q, want %q", restored.ActiveTitle, "title1")
	}
	if len(restored.EarnedEffects) != 1 {
		t.Errorf("EarnedEffects = %d, want 1", len(restored.EarnedEffects))
	}
	if restored.ActiveEffect != "effect1" {
		t.Errorf("ActiveEffect = %q, want %q", restored.ActiveEffect, "effect1")
	}
	if !restored.HasAchievement("spring_quest_master") {
		t.Error("Achievement not preserved")
	}
	if restored.TotalEventsParticipated != 1 {
		t.Errorf("TotalEventsParticipated = %d, want 1", restored.TotalEventsParticipated)
	}
	if restored.TotalQuestsCompleted != 2 {
		t.Errorf("TotalQuestsCompleted = %d, want 2", restored.TotalQuestsCompleted)
	}
}
