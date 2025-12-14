package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// TestEventQuestIntegration_FullLifecycle tests the complete event quest lifecycle.
func TestEventQuestIntegration_FullLifecycle(t *testing.T) {
	// Setup
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)}
	eventCalendarSystem := NewEventCalendarSystem(world, clock)
	eventQuestSystem := NewEventQuestSystem(world, clock)

	// Create world entity with seasonal events
	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	worldEntity.AddComponent(eventComp)

	// Create player entity
	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	// Create NPC entity
	npcEntity := NewEntity(3)
	npcEntity.AddComponent(&ScheduleComponent{
		Activities: []ScheduledActivity{
			{ActivityType: ActivityWork, StartHour: 8, EndHour: 18, LocationName: "market"},
		},
	})

	entities := []*Entity{worldEntity, playerEntity, npcEntity}

	// Manually trigger a spring event
	eventCalendarSystem.TriggerEvent(worldEntity, "spring_festival", 7)

	// Run event calendar system to process the event
	eventCalendarSystem.Update(entities, 0.016)

	// Run event quest system to generate quests
	eventQuestSystem.Update(entities, 0.016)

	// Verify quests were generated
	if len(questComp.AvailableQuests) != 3 {
		t.Errorf("Expected 3 available quests, got %d", len(questComp.AvailableQuests))
	}

	// Get dialog options (simulating player talking to NPC)
	dialogOptions := eventQuestSystem.GetEventNPCDialogOptions(playerEntity, npcEntity)
	if len(dialogOptions) != 3 {
		t.Errorf("Expected 3 dialog options for available quests, got %d", len(dialogOptions))
	}

	// Accept a quest
	collectionQuestID := ""
	for _, q := range questComp.AvailableQuests {
		if q.QuestType == EventQuestCollection {
			collectionQuestID = q.ID
			break
		}
	}

	if collectionQuestID == "" {
		t.Fatal("No collection quest found")
	}

	if !eventQuestSystem.AcceptEventQuest(playerEntity, worldEntity, collectionQuestID) {
		t.Fatal("Failed to accept quest")
	}

	if len(questComp.ActiveQuests) != 1 {
		t.Errorf("Expected 1 active quest, got %d", len(questComp.ActiveQuests))
	}

	if len(questComp.AvailableQuests) != 2 {
		t.Errorf("Expected 2 available quests, got %d", len(questComp.AvailableQuests))
	}

	// Simulate quest progress
	activeQuest := questComp.GetActiveQuest(collectionQuestID)
	if activeQuest == nil {
		t.Fatal("Active quest not found")
	}

	// Increment progress
	required := activeQuest.Definition.Objectives[0].Required
	for i := 0; i < required; i++ {
		questComp.IncrementProgress(collectionQuestID, 0, 1)
	}

	// Verify quest is complete
	if !questComp.IsQuestComplete(collectionQuestID) {
		t.Error("Quest should be complete")
	}

	// Get updated dialog options (should have turn-in option)
	dialogOptions = eventQuestSystem.GetEventNPCDialogOptions(playerEntity, npcEntity)
	hasCompleteOption := false
	for _, opt := range dialogOptions {
		if opt.Action == ActionCompleteEventQuest && opt.Payload == collectionQuestID {
			hasCompleteOption = true
			break
		}
	}
	if !hasCompleteOption {
		t.Error("Should have complete option for finished quest")
	}

	// Turn in quest
	success, reward := eventQuestSystem.TurnInEventQuest(playerEntity, collectionQuestID)
	if !success {
		t.Error("Turn in should succeed")
	}
	if reward == nil {
		t.Error("Should receive reward")
	}

	if len(questComp.CompletedQuests) != 1 {
		t.Errorf("Expected 1 completed quest, got %d", len(questComp.CompletedQuests))
	}
}

// TestEventQuestIntegration_EventExpiration tests quest expiration when event ends.
func TestEventQuestIntegration_EventExpiration(t *testing.T) {
	world := NewWorld()
	startTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC) // Mid-June, no default events
	clock := &mockEventQuestClock{currentTime: startTime}
	eventQuestSystem := NewEventQuestSystem(world, clock)

	// Create world entity with custom event (no default calendar)
	worldEntity := NewEntity(1)
	eventComp := &SeasonalEventComponent{
		ActiveEvents:  make([]EventInstance, 0),
		EventCalendar: make([]EventDefinition, 0), // Empty calendar to avoid defaults
		CurrentTime:   clock.Now(),
		UseRealTime:   false,
		CalendarSeed:  12345,
	}

	// Add a manually created event that ends in 1 day
	eventEndTime := clock.Now().Add(24 * time.Hour)
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "test_event",
			Name:  "Test Event",
			Theme: EventThemeSummer,
			Seed:  12345,
		},
		StartTime: clock.Now().Add(-time.Hour),
		EndTime:   eventEndTime,
		Phase:     EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	// Create player entity
	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}

	// Generate quests for the event
	eventQuestSystem.Update(entities, 0.016)

	// Accept a quest
	if len(questComp.AvailableQuests) == 0 {
		t.Fatal("No quests available")
	}
	questID := questComp.AvailableQuests[0].ID
	eventQuestSystem.AcceptEventQuest(playerEntity, worldEntity, questID)

	if len(questComp.ActiveQuests) != 1 {
		t.Fatal("Quest not accepted")
	}

	// Verify the quest has the correct expiration
	activeQuest := questComp.ActiveQuests[0]
	if !activeQuest.ExpiresAt.Equal(eventEndTime) {
		t.Errorf("Quest expires at %v, want %v", activeQuest.ExpiresAt, eventEndTime)
	}

	// Advance time past event end
	clock.AdvanceDuration(26 * time.Hour)
	eventComp.CurrentTime = clock.Now()

	// Clear the event from active events (simulating event ended)
	eventComp.ActiveEvents = []EventInstance{}

	// Run system - should expire the quest
	eventQuestSystem.Update(entities, 0.016)

	// Quest should be expired
	if len(questComp.ActiveQuests) != 0 {
		t.Errorf("Expected 0 active quests after expiration, got %d", len(questComp.ActiveQuests))
	}

	if len(questComp.ExpiredQuests) != 1 {
		t.Errorf("Expected 1 expired quest, got %d", len(questComp.ExpiredQuests))
	}
}

// TestEventQuestIntegration_MultipleEvents tests handling multiple concurrent events.
func TestEventQuestIntegration_MultipleEvents(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 6, 21, 12, 0, 0, 0, time.UTC)}
	eventQuestSystem := NewEventQuestSystem(world, clock)

	// Create world entity with two active events
	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()

	// Add two active events
	eventComp.ActiveEvents = append(eventComp.ActiveEvents,
		EventInstance{
			Definition: EventDefinition{
				ID:    "summer_solstice",
				Name:  "Summer Solstice",
				Theme: EventThemeSummer,
				Seed:  11111,
			},
			EndTime: clock.Now().Add(5 * 24 * time.Hour),
			Phase:   EventPhaseActive,
		},
		EventInstance{
			Definition: EventDefinition{
				ID:    "midsummer_fair",
				Name:  "Midsummer Fair",
				Theme: EventThemeSummer,
				Seed:  22222,
			},
			EndTime: clock.Now().Add(3 * 24 * time.Hour),
			Phase:   EventPhaseActive,
		},
	)
	worldEntity.AddComponent(eventComp)

	// Create player entity
	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(6) // Allow more quests
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}
	eventQuestSystem.Update(entities, 0.016)

	// Should have 6 quests (3 per event)
	if len(questComp.AvailableQuests) != 6 {
		t.Errorf("Expected 6 available quests (3 per event), got %d", len(questComp.AvailableQuests))
	}

	// Verify quests are distributed between events
	solsticeQuests := questComp.GetAvailableQuestsForEvent("summer_solstice")
	fairQuests := questComp.GetAvailableQuestsForEvent("midsummer_fair")

	if len(solsticeQuests) != 3 {
		t.Errorf("Expected 3 solstice quests, got %d", len(solsticeQuests))
	}

	if len(fairQuests) != 3 {
		t.Errorf("Expected 3 fair quests, got %d", len(fairQuests))
	}
}

// TestEventQuestIntegration_QuestTypeVariety tests that all quest types generate correctly.
func TestEventQuestIntegration_QuestTypeVariety(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 12, 25, 12, 0, 0, 0, time.UTC)}
	eventQuestSystem := NewEventQuestSystem(world, clock)

	// Create world entity with winter event
	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "winter_celebration",
			Name:  "Hearthfire Festival",
			Theme: EventThemeWinter,
			Seed:  99999,
		},
		EndTime: clock.Now().Add(14 * 24 * time.Hour),
		Phase:   EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}
	eventQuestSystem.Update(entities, 0.016)

	// Verify quest types
	types := make(map[EventQuestType]bool)
	for _, q := range questComp.AvailableQuests {
		types[q.QuestType] = true

		// Verify quest has appropriate content
		if q.Name == "" {
			t.Errorf("Quest %s has empty name", q.ID)
		}
		if q.Description == "" {
			t.Errorf("Quest %s has empty description", q.ID)
		}
		if len(q.Objectives) == 0 {
			t.Errorf("Quest %s has no objectives", q.ID)
		}
		if q.Reward.XP <= 0 {
			t.Errorf("Quest %s has no XP reward", q.ID)
		}
	}

	if !types[EventQuestCollection] {
		t.Error("Missing collection quest type")
	}
	if !types[EventQuestExploration] {
		t.Error("Missing exploration quest type")
	}
	if !types[EventQuestBoss] {
		t.Error("Missing boss quest type")
	}
}

// TestEventQuestIntegration_WithLivingWorldNPC tests integration with Living World NPCs.
func TestEventQuestIntegration_WithLivingWorldNPC(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 10, 1, 14, 0, 0, 0, time.UTC)}
	eventQuestSystem := NewEventQuestSystem(world, clock)

	// Create world entity with harvest event
	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "harvest_festival",
			Name:  "Crimson Gathering",
			Theme: EventThemeAutumn,
			Seed:  77777,
		},
		EndTime: clock.Now().Add(10 * 24 * time.Hour),
		Phase:   EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	// Create player entity
	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	// Create event NPC with schedule (Living World integration)
	eventNPC := NewEntity(3)
	scheduleComp := &ScheduleComponent{
		Activities: []ScheduledActivity{
			{
				ActivityType: ActivityWork,
				StartHour:    8,
				EndHour:      20,
				LocationName: "festival_grounds",
			},
		},
	}
	eventNPC.AddComponent(scheduleComp)

	// Add operating hours for event NPC
	operatingHours := NewOperatingHoursComponent()
	operatingHours.SetHours(8, 20)
	eventNPC.AddComponent(operatingHours)

	entities := []*Entity{worldEntity, playerEntity, eventNPC}

	// Generate quests
	eventQuestSystem.Update(entities, 0.016)

	// Verify NPC can offer quests during operating hours
	dialogOptions := eventQuestSystem.GetEventNPCDialogOptions(playerEntity, eventNPC)
	if len(dialogOptions) == 0 {
		t.Error("Event NPC should offer quest dialog options")
	}

	// Accept a quest through the NPC
	if len(questComp.AvailableQuests) == 0 {
		t.Fatal("No available quests")
	}
	questID := questComp.AvailableQuests[0].ID
	success := eventQuestSystem.AcceptEventQuest(playerEntity, worldEntity, questID)
	if !success {
		t.Error("Should be able to accept quest from event NPC")
	}

	// Verify quest tracking
	current, required, complete := eventQuestSystem.GetEventQuestProgress(playerEntity, questID)
	if current != 0 {
		t.Errorf("Initial current progress should be 0, got %d", current)
	}
	if required <= 0 {
		t.Error("Required should be positive")
	}
	if complete {
		t.Error("Quest should not be complete initially")
	}
}

// TestEventQuestIntegration_DeterministicGeneration tests determinism of quest generation.
func TestEventQuestIntegration_DeterministicGeneration(t *testing.T) {
	// Run generation twice with same seed
	generateQuests := func(seed int64) []EventQuestDefinition {
		world := NewWorld()
		clock := &mockEventQuestClock{currentTime: time.Date(2025, 4, 15, 12, 0, 0, 0, time.UTC)}
		eventQuestSystem := NewEventQuestSystem(world, clock)

		worldEntity := NewEntity(1)
		eventComp := NewSeasonalEventComponent(seed, false)
		eventComp.CurrentTime = clock.Now()
		eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
			Definition: EventDefinition{
				ID:    "test_event",
				Name:  "Test Event",
				Theme: EventThemeSpring,
				Seed:  seed,
			},
			Phase: EventPhaseActive,
		})
		worldEntity.AddComponent(eventComp)

		playerEntity := NewEntity(2)
		questComp := NewEventQuestComponent(3)
		playerEntity.AddComponent(questComp)

		entities := []*Entity{worldEntity, playerEntity}
		eventQuestSystem.Update(entities, 0.016)

		return questComp.AvailableQuests
	}

	quests1 := generateQuests(42)
	quests2 := generateQuests(42)

	if len(quests1) != len(quests2) {
		t.Fatal("Quest count differs between generations")
	}

	for i := range quests1 {
		if quests1[i].ID != quests2[i].ID {
			t.Errorf("Quest ID differs at index %d: %s != %s", i, quests1[i].ID, quests2[i].ID)
		}
		if quests1[i].Name != quests2[i].Name {
			t.Errorf("Quest name differs at index %d: %s != %s", i, quests1[i].Name, quests2[i].Name)
		}
		if quests1[i].Description != quests2[i].Description {
			t.Errorf("Quest description differs at index %d", i)
		}
		if len(quests1[i].Objectives) > 0 && len(quests2[i].Objectives) > 0 {
			if quests1[i].Objectives[0].Required != quests2[i].Objectives[0].Required {
				t.Errorf("Objective required differs at index %d", i)
			}
		}
	}
}

// TestEventQuestIntegration_RewardDistribution tests that rewards are appropriate.
func TestEventQuestIntegration_RewardDistribution(t *testing.T) {
	event := EventInstance{
		Definition: EventDefinition{
			ID:    "test_event",
			Theme: EventThemeWinter,
			Seed:  12345,
		},
	}

	quests := GenerateEventQuests(event, 12345)

	for _, q := range quests {
		// All quests should have positive XP
		if q.Reward.XP <= 0 {
			t.Errorf("Quest %s has no XP reward", q.ID)
		}

		// All quests should have positive gold
		if q.Reward.Gold <= 0 {
			t.Errorf("Quest %s has no gold reward", q.ID)
		}

		// Boss quests should have skill points
		if q.QuestType == EventQuestBoss && q.Reward.SkillPoints <= 0 {
			t.Errorf("Boss quest %s should have skill points", q.ID)
		}

		// Boss quests should require higher level
		if q.QuestType == EventQuestBoss && q.RequiredLevel < 5 {
			t.Errorf("Boss quest %s should require level 5+, got %d", q.ID, q.RequiredLevel)
		}
	}
}

// TestEventQuestComponent_MultipleObjectives tests quests with multiple objectives.
func TestEventQuestComponent_MultipleObjectives(t *testing.T) {
	comp := NewEventQuestComponent(3)

	// Create quest with multiple objectives
	multiObjQuest := EventQuestDefinition{
		ID:        "multi_obj_quest",
		EventID:   "test_event",
		Name:      "Complete Tasks",
		QuestType: EventQuestCollection,
		Objectives: []quest.Objective{
			{Description: "Collect apples", Required: 5},
			{Description: "Collect oranges", Required: 3},
			{Description: "Collect berries", Required: 10},
		},
	}

	comp.AvailableQuests = append(comp.AvailableQuests, multiObjQuest)
	comp.AcceptQuest("multi_obj_quest", time.Now().Add(24*time.Hour))

	// Complete first two objectives
	comp.UpdateProgress("multi_obj_quest", 0, 5)
	comp.UpdateProgress("multi_obj_quest", 1, 3)

	// Quest should not be complete yet
	if comp.IsQuestComplete("multi_obj_quest") {
		t.Error("Quest should not be complete with one incomplete objective")
	}

	// Complete last objective
	comp.UpdateProgress("multi_obj_quest", 2, 10)

	// Now quest should be complete
	if !comp.IsQuestComplete("multi_obj_quest") {
		t.Error("Quest should be complete with all objectives done")
	}
}
