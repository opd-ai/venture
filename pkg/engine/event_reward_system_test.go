package engine

import (
	"testing"
	"time"
)

// mockEventRewardClock implements GameClock for testing.
type mockEventRewardClock struct {
	currentTime time.Time
}

func (m *mockEventRewardClock) Now() time.Time {
	return m.currentTime
}

func (m *mockEventRewardClock) Advance(deltaTime float64) {
	m.currentTime = m.currentTime.Add(time.Duration(deltaTime * float64(time.Second)))
}

func (m *mockEventRewardClock) Reset(startTime time.Time) {
	m.currentTime = startTime
}

func TestNewEventRewardSystem(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}

	sys := NewEventRewardSystem(world, clock)

	if sys == nil {
		t.Fatal("NewEventRewardSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.VendorInventory == nil {
		t.Error("VendorInventory should be initialized")
	}
	if sys.EventAchievements == nil {
		t.Error("EventAchievements should be initialized")
	}
}

func TestEventRewardSystem_Update_NoComponents(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	// Create entity without event_reward component
	entity := NewEntity(1)

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestEventRewardSystem_Update_WithComponents(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	// Create entity with event_reward component
	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	entity.AddComponent(rewardComp)

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestEventRewardSystem_RegisterEventParticipation(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	// Create world entity with seasonal events
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(12345, false)
	// Add an active event
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "test_event",
			Name:  "Test Event",
			Theme: EventThemeSpring,
		},
		Phase: EventPhaseActive,
	})
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Process pending entity additions
	world.Update(0)

	// Create player entity
	entity := NewEntity(2)
	rewardComp := NewEventRewardComponent()
	entity.AddComponent(rewardComp)

	// Register participation
	sys.RegisterEventParticipation(entity)

	if rewardComp.TotalEventsParticipated != 1 {
		t.Errorf("TotalEventsParticipated = %d, want 1", rewardComp.TotalEventsParticipated)
	}
}

func TestEventRewardSystem_GrantEventCurrency(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	entity.AddComponent(rewardComp)

	sys.GrantEventCurrency(entity, "spring_festival", 100, "bonus")

	if rewardComp.GetCurrency("spring_festival") != 100 {
		t.Errorf("Currency = %d, want 100", rewardComp.GetCurrency("spring_festival"))
	}
}

func TestEventRewardSystem_GrantEventCurrency_NoComponent(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	entity := NewEntity(1)

	// Should not panic
	sys.GrantEventCurrency(entity, "test_event", 50, "test")
}

func TestEventRewardSystem_GetPlayerEventStats(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.TotalEventsParticipated = 3
	rewardComp.TotalQuestsCompleted = 5
	rewardComp.TotalCurrencyEarned = 500
	rewardComp.CompleteAchievement("test_ach")
	rewardComp.AddTitle(EventCosmeticTitle{ID: "test_title"})
	rewardComp.AddEffect(EventVisualEffect{ID: "test_effect"})
	rewardComp.AddReward(EventReward{ID: "test_reward"})
	entity.AddComponent(rewardComp)

	stats := sys.GetPlayerEventStats(entity)

	if stats == nil {
		t.Fatal("GetPlayerEventStats returned nil")
	}
	if stats["total_events_participated"].(int) != 3 {
		t.Errorf("total_events_participated = %v, want 3", stats["total_events_participated"])
	}
	if stats["total_quests_completed"].(int) != 5 {
		t.Errorf("total_quests_completed = %v, want 5", stats["total_quests_completed"])
	}
	if stats["achievements_completed"].(int) != 1 {
		t.Errorf("achievements_completed = %v, want 1", stats["achievements_completed"])
	}
}

func TestEventRewardSystem_GetPlayerEventStats_NoComponent(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	entity := NewEntity(1)

	stats := sys.GetPlayerEventStats(entity)
	if stats != nil {
		t.Error("GetPlayerEventStats should return nil for entity without component")
	}
}

func TestEventRewardSystem_PurchaseFromVendor(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	// Setup vendor inventory
	sys.VendorInventory[eventID] = []EventVendorItem{
		{
			ID:           "test_item",
			EventID:      eventID,
			Name:         "Test Item",
			RewardType:   EventRewardItem,
			CurrencyCost: 50,
			Stock:        2,
			Rarity:       "common",
		},
	}

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency(eventID, 100)
	entity.AddComponent(rewardComp)

	// Successful purchase
	success := sys.PurchaseFromVendor(entity, eventID, "test_item")
	if !success {
		t.Error("PurchaseFromVendor should succeed with sufficient currency")
	}

	if rewardComp.GetCurrency(eventID) != 50 {
		t.Errorf("Currency after purchase = %d, want 50", rewardComp.GetCurrency(eventID))
	}

	// Check stock decreased
	if sys.VendorInventory[eventID][0].Stock != 1 {
		t.Errorf("Stock = %d, want 1", sys.VendorInventory[eventID][0].Stock)
	}

	// Check reward added
	if len(rewardComp.EarnedRewards) != 1 {
		t.Errorf("EarnedRewards = %d, want 1", len(rewardComp.EarnedRewards))
	}
}

func TestEventRewardSystem_PurchaseFromVendor_InsufficientCurrency(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	sys.VendorInventory[eventID] = []EventVendorItem{
		{
			ID:           "expensive_item",
			EventID:      eventID,
			Name:         "Expensive Item",
			RewardType:   EventRewardItem,
			CurrencyCost: 1000,
			Stock:        1,
		},
	}

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency(eventID, 100)
	entity.AddComponent(rewardComp)

	success := sys.PurchaseFromVendor(entity, eventID, "expensive_item")
	if success {
		t.Error("PurchaseFromVendor should fail with insufficient currency")
	}

	// Currency should not change
	if rewardComp.GetCurrency(eventID) != 100 {
		t.Errorf("Currency should not change after failed purchase")
	}
}

func TestEventRewardSystem_PurchaseFromVendor_OutOfStock(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	sys.VendorInventory[eventID] = []EventVendorItem{
		{
			ID:           "out_of_stock_item",
			EventID:      eventID,
			Name:         "Sold Out Item",
			RewardType:   EventRewardItem,
			CurrencyCost: 10,
			Stock:        0, // Out of stock
		},
	}

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency(eventID, 100)
	entity.AddComponent(rewardComp)

	success := sys.PurchaseFromVendor(entity, eventID, "out_of_stock_item")
	if success {
		t.Error("PurchaseFromVendor should fail when out of stock")
	}
}

func TestEventRewardSystem_PurchaseFromVendor_UnlimitedStock(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	sys.VendorInventory[eventID] = []EventVendorItem{
		{
			ID:           "unlimited_item",
			EventID:      eventID,
			Name:         "Unlimited Item",
			RewardType:   EventRewardTitle,
			CurrencyCost: 10,
			Stock:        -1, // Unlimited
		},
	}

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency(eventID, 100)
	entity.AddComponent(rewardComp)

	// Buy multiple times
	for i := 0; i < 3; i++ {
		success := sys.PurchaseFromVendor(entity, eventID, "unlimited_item")
		if !success {
			t.Errorf("Purchase %d should succeed with unlimited stock", i+1)
		}
	}

	// Stock should still be -1
	if sys.VendorInventory[eventID][0].Stock != -1 {
		t.Errorf("Unlimited stock should remain -1")
	}
}

func TestEventRewardSystem_PurchaseFromVendor_NonexistentItem(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency("spring_festival", 100)
	entity.AddComponent(rewardComp)

	success := sys.PurchaseFromVendor(entity, "spring_festival", "nonexistent_item")
	if success {
		t.Error("PurchaseFromVendor should fail for nonexistent item")
	}
}

func TestEventRewardSystem_PurchaseFromVendor_TitleReward(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	sys.VendorInventory[eventID] = []EventVendorItem{
		{
			ID:           "title_item",
			EventID:      eventID,
			Name:         "Champion Title",
			RewardType:   EventRewardTitle,
			CurrencyCost: 100,
			Stock:        -1,
			Rarity:       "epic",
		},
	}

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency(eventID, 200)
	entity.AddComponent(rewardComp)

	success := sys.PurchaseFromVendor(entity, eventID, "title_item")
	if !success {
		t.Error("Purchase should succeed")
	}

	// Check title was added
	if len(rewardComp.EarnedTitles) != 1 {
		t.Errorf("EarnedTitles = %d, want 1", len(rewardComp.EarnedTitles))
	}
	if rewardComp.EarnedTitles[0].DisplayName != "Champion Title" {
		t.Errorf("Title name = %q, want %q", rewardComp.EarnedTitles[0].DisplayName, "Champion Title")
	}
}

func TestEventRewardSystem_PurchaseFromVendor_EffectReward(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	sys.VendorInventory[eventID] = []EventVendorItem{
		{
			ID:           "effect_item",
			EventID:      eventID,
			Name:         "Glowing Aura",
			RewardType:   EventRewardEffect,
			CurrencyCost: 500,
			Stock:        1,
			Rarity:       "legendary",
		},
	}

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	rewardComp.AddCurrency(eventID, 1000)
	entity.AddComponent(rewardComp)

	success := sys.PurchaseFromVendor(entity, eventID, "effect_item")
	if !success {
		t.Error("Purchase should succeed")
	}

	// Check effect was added
	if len(rewardComp.EarnedEffects) != 1 {
		t.Errorf("EarnedEffects = %d, want 1", len(rewardComp.EarnedEffects))
	}
}

func TestEventRewardSystem_GetVendorInventory_Generate(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	// Create world entity with seasonal events
	worldEntity := NewEntity(1)
	seasonalComp := NewSeasonalEventComponent(12345, false)
	seasonalComp.ActiveEvents = append(seasonalComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Spring Festival",
			Theme: EventThemeSpring,
			Seed:  54321,
		},
		Phase: EventPhaseActive,
	})
	worldEntity.AddComponent(seasonalComp)
	world.AddEntity(worldEntity)

	// Process pending entity additions
	world.Update(0)

	inventory := sys.GetVendorInventory("spring_festival")

	if len(inventory) == 0 {
		t.Error("GetVendorInventory should generate inventory")
	}

	// Should be cached now
	inventory2 := sys.GetVendorInventory("spring_festival")
	if len(inventory) != len(inventory2) {
		t.Error("Second call should return same cached inventory")
	}
}

func TestEventRewardSystem_GetVendorInventory_NoSeasonalComponent(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	inventory := sys.GetVendorInventory("nonexistent_event")

	if inventory != nil {
		t.Error("GetVendorInventory should return nil when no seasonal component exists")
	}
}

func TestGenerateEventVendorInventory(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Spring Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
	}

	inventory := GenerateEventVendorInventory(event, 12345)

	if len(inventory) != 5 {
		t.Errorf("GenerateEventVendorInventory returned %d items, want 5", len(inventory))
	}

	// Check each item has required fields
	for _, item := range inventory {
		if item.ID == "" {
			t.Error("Item ID should not be empty")
		}
		if item.Name == "" {
			t.Error("Item Name should not be empty")
		}
		if item.EventID != event.Definition.ID {
			t.Errorf("Item EventID = %q, want %q", item.EventID, event.Definition.ID)
		}
		if item.CurrencyCost <= 0 {
			t.Errorf("Item %s has invalid cost: %d", item.ID, item.CurrencyCost)
		}
	}

	// Check we have different rarities
	rarities := make(map[string]bool)
	for _, item := range inventory {
		rarities[item.Rarity] = true
	}
	if len(rarities) < 3 {
		t.Errorf("Expected at least 3 different rarities, got %d", len(rarities))
	}
}

func TestGenerateEventVendorInventory_Determinism(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "summer_solstice",
			Name:  "Summer Solstice",
			Theme: EventThemeSummer,
			Seed:  99999,
		},
	}

	inv1 := GenerateEventVendorInventory(event, 99999)
	inv2 := GenerateEventVendorInventory(event, 99999)

	if len(inv1) != len(inv2) {
		t.Fatalf("Different inventory sizes: %d vs %d", len(inv1), len(inv2))
	}

	for i := range inv1 {
		if inv1[i].ID != inv2[i].ID {
			t.Errorf("Item %d ID differs: %q vs %q", i, inv1[i].ID, inv2[i].ID)
		}
		if inv1[i].CurrencyCost != inv2[i].CurrencyCost {
			t.Errorf("Item %d cost differs: %d vs %d", i, inv1[i].CurrencyCost, inv2[i].CurrencyCost)
		}
	}
}

func TestGenerateEventVendorInventory_AllThemes(t *testing.T) {
	themes := []EventTheme{
		EventThemeSpring,
		EventThemeSummer,
		EventThemeAutumn,
		EventThemeWinter,
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			event := EventInstance{
				Definition: EventDefinition{
					ID:    string(theme) + "_event",
					Name:  string(theme) + " Event",
					Theme: theme,
					Seed:  int64(len(theme)),
				},
			}

			inventory := GenerateEventVendorInventory(event, int64(len(theme)))

			if len(inventory) != 5 {
				t.Errorf("Expected 5 items, got %d", len(inventory))
			}
		})
	}
}

func TestEventRewardSystem_CheckAchievementCompletion(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	eventID := "spring_festival"

	tests := []struct {
		name        string
		requirement string
		required    int
		setupReward func(*EventRewardComponent)
		setupQuest  func(*EventQuestComponent)
		want        bool
	}{
		{
			name:        "participate - met",
			requirement: "participate",
			required:    1,
			setupReward: func(c *EventRewardComponent) {
				c.TotalEventsParticipated = 1
			},
			want: true,
		},
		{
			name:        "participate - not met",
			requirement: "participate",
			required:    5,
			setupReward: func(c *EventRewardComponent) {
				c.TotalEventsParticipated = 2
			},
			want: false,
		},
		{
			name:        "earn_currency - met",
			requirement: "earn_currency",
			required:    100,
			setupReward: func(c *EventRewardComponent) {
				c.AddCurrency(eventID, 150)
			},
			want: true,
		},
		{
			name:        "complete_quests - met",
			requirement: "complete_quests",
			required:    2,
			setupQuest: func(c *EventQuestComponent) {
				c.CompletedQuests = append(c.CompletedQuests,
					EventQuestInstance{Definition: EventQuestDefinition{EventID: eventID}},
					EventQuestInstance{Definition: EventQuestDefinition{EventID: eventID}},
				)
			},
			want: true,
		},
		{
			name:        "defeat_boss - met",
			requirement: "defeat_boss",
			required:    1,
			setupQuest: func(c *EventQuestComponent) {
				c.CompletedQuests = append(c.CompletedQuests,
					EventQuestInstance{Definition: EventQuestDefinition{
						EventID:   eventID,
						QuestType: EventQuestBoss,
					}},
				)
			},
			want: true,
		},
		{
			name:        "explore_location - met",
			requirement: "explore_location",
			required:    1,
			setupQuest: func(c *EventQuestComponent) {
				c.CompletedQuests = append(c.CompletedQuests,
					EventQuestInstance{Definition: EventQuestDefinition{
						EventID:   eventID,
						QuestType: EventQuestExploration,
					}},
				)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewardComp := NewEventRewardComponent()
			questComp := NewEventQuestComponent(3)

			if tt.setupReward != nil {
				tt.setupReward(rewardComp)
			}
			if tt.setupQuest != nil {
				tt.setupQuest(questComp)
			}

			ach := EventAchievementDef{
				ID:             "test_ach",
				EventID:        eventID,
				Requirement:    tt.requirement,
				RequiredAmount: tt.required,
			}

			got := sys.checkAchievementCompletion(rewardComp, questComp, ach, eventID)
			if got != tt.want {
				t.Errorf("checkAchievementCompletion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventRewardSystem_GrantAchievementReward(t *testing.T) {
	world := NewWorld()
	clock := &mockEventRewardClock{currentTime: time.Now()}
	sys := NewEventRewardSystem(world, clock)

	entity := NewEntity(1)
	rewardComp := NewEventRewardComponent()
	entity.AddComponent(rewardComp)

	ach := EventAchievementDef{
		ID:          "test_ach",
		EventID:     "spring_festival",
		Name:        "Test Achievement",
		Description: "A test achievement",
		Requirement: "test",
		Reward: EventReward{
			ID:      "test_ach_reward",
			EventID: "spring_festival",
			Type:    EventRewardCurrency,
			Name:    "Achievement Bonus",
			Value:   100,
		},
	}

	sys.grantAchievementReward(entity, rewardComp, ach)

	if !rewardComp.HasAchievement("test_ach") {
		t.Error("Achievement should be marked as completed")
	}

	if rewardComp.GetCurrency("spring_festival") != 100 {
		t.Errorf("Currency = %d, want 100", rewardComp.GetCurrency("spring_festival"))
	}
}

func TestGetThemePotionName(t *testing.T) {
	themes := []EventTheme{EventThemeSpring, EventThemeSummer, EventThemeAutumn, EventThemeWinter, "unknown"}

	for _, theme := range themes {
		name := getThemePotionName(theme)
		if name == "" {
			t.Errorf("getThemePotionName(%q) returned empty string", theme)
		}
	}
}

func TestGetThemeAccessoryName(t *testing.T) {
	themes := []EventTheme{EventThemeSpring, EventThemeSummer, EventThemeAutumn, EventThemeWinter, "unknown"}

	for _, theme := range themes {
		name := getThemeAccessoryName(theme)
		if name == "" {
			t.Errorf("getThemeAccessoryName(%q) returned empty string", theme)
		}
	}
}
