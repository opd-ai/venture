package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// mockEventQuestClock implements GameClock for testing.
type mockEventQuestClock struct {
	currentTime time.Time
}

func (m *mockEventQuestClock) Now() time.Time {
	return m.currentTime
}

func (m *mockEventQuestClock) Advance(deltaTime float64) {
	m.currentTime = m.currentTime.Add(time.Duration(deltaTime * float64(time.Second)))
}

func (m *mockEventQuestClock) Reset(startTime time.Time) {
	m.currentTime = startTime
}

func (m *mockEventQuestClock) AdvanceDuration(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

func TestNewEventQuestSystem(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}

	system := NewEventQuestSystem(world, clock)

	if system == nil {
		t.Fatal("NewEventQuestSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.clock != clock {
		t.Error("clock not set correctly")
	}
}

func TestEventQuestSystem_Update_NoClock(t *testing.T) {
	world := NewWorld()
	system := NewEventQuestSystem(world, nil)

	// Should not panic with nil clock
	entity := NewEntity(1)
	system.Update([]*Entity{entity}, 0.016)
}

func TestEventQuestSystem_Update_GeneratesQuests(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)}
	system := NewEventQuestSystem(world, clock)

	// Create world entity with active seasonal event
	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()

	// Add an active event
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Bloom Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		StartTime: clock.Now().Add(-time.Hour),
		EndTime:   clock.Now().Add(7 * 24 * time.Hour),
		Phase:     EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	// Create player entity with event quest component
	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}
	system.Update(entities, 0.016)

	// Should have generated 3 quests
	if len(questComp.AvailableQuests) != 3 {
		t.Errorf("Expected 3 available quests, got %d", len(questComp.AvailableQuests))
	}

	// Verify quests are for the correct event
	for _, q := range questComp.AvailableQuests {
		if q.EventID != "spring_festival" {
			t.Errorf("Quest EventID = %v, want spring_festival", q.EventID)
		}
	}

	// Verify last generation event ID is set
	if questComp.LastGenerationEventID != "spring_festival" {
		t.Errorf("LastGenerationEventID = %v, want spring_festival", questComp.LastGenerationEventID)
	}
}

func TestEventQuestSystem_Update_DoesNotRegenerateQuests(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)}
	system := NewEventQuestSystem(world, clock)

	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Bloom Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		StartTime: clock.Now().Add(-time.Hour),
		EndTime:   clock.Now().Add(7 * 24 * time.Hour),
		Phase:     EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}

	// First update generates quests
	system.Update(entities, 0.016)
	initialCount := len(questComp.AvailableQuests)

	// Second update should not add more quests
	system.Update(entities, 0.016)

	if len(questComp.AvailableQuests) != initialCount {
		t.Errorf("Quests regenerated: %d != %d", len(questComp.AvailableQuests), initialCount)
	}
}

func TestEventQuestSystem_Update_ExpiresQuests(t *testing.T) {
	world := NewWorld()
	now := time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)
	clock := &mockEventQuestClock{currentTime: now}
	system := NewEventQuestSystem(world, clock)

	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Bloom Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		StartTime: clock.Now().Add(-time.Hour),
		EndTime:   clock.Now().Add(7 * 24 * time.Hour),
		Phase:     EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)

	// Add an active quest that expires in the past
	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:      "expired_quest",
			EventID: "spring_festival",
		},
		Status:    quest.StatusActive,
		ExpiresAt: now.Add(-time.Hour), // Expired 1 hour ago
		Progress:  make(map[int]int),
	})
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}
	system.Update(entities, 0.016)

	if len(questComp.ActiveQuests) != 0 {
		t.Errorf("Expected 0 active quests after expiration, got %d", len(questComp.ActiveQuests))
	}

	if len(questComp.ExpiredQuests) != 1 {
		t.Errorf("Expected 1 expired quest, got %d", len(questComp.ExpiredQuests))
	}
}

func TestEventQuestSystem_Update_CleansUpEndedEvents(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)}
	system := NewEventQuestSystem(world, clock)

	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	// No active events
	worldEntity.AddComponent(eventComp)

	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)

	// Add quests for an event that's no longer active
	questComp.AvailableQuests = append(questComp.AvailableQuests, EventQuestDefinition{
		ID:      "old_quest",
		EventID: "ended_event",
	})
	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:      "old_active_quest",
			EventID: "ended_event",
		},
		ExpiresAt: clock.Now().Add(time.Hour),
		Progress:  make(map[int]int),
	})
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}
	system.Update(entities, 0.016)

	if len(questComp.AvailableQuests) != 0 {
		t.Errorf("Expected 0 available quests for ended event, got %d", len(questComp.AvailableQuests))
	}

	if len(questComp.ActiveQuests) != 0 {
		t.Errorf("Expected 0 active quests for ended event, got %d", len(questComp.ActiveQuests))
	}

	if len(questComp.ExpiredQuests) != 1 {
		t.Errorf("Expected 1 expired quest, got %d", len(questComp.ExpiredQuests))
	}
}

func TestEventQuestSystem_Update_CompletesQuests(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)}
	system := NewEventQuestSystem(world, clock)

	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Bloom Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		Phase: EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)

	// Add an active quest with completed objectives
	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:      "completed_quest",
			EventID: "spring_festival",
			Objectives: []quest.Objective{
				{Required: 5},
			},
		},
		Status:    quest.StatusActive,
		ExpiresAt: clock.Now().Add(24 * time.Hour),
		Progress:  map[int]int{0: 5}, // Objective complete
	})
	playerEntity.AddComponent(questComp)

	entities := []*Entity{worldEntity, playerEntity}
	system.Update(entities, 0.016)

	if len(questComp.ActiveQuests) != 0 {
		t.Errorf("Expected 0 active quests after completion, got %d", len(questComp.ActiveQuests))
	}

	if len(questComp.CompletedQuests) != 1 {
		t.Errorf("Expected 1 completed quest, got %d", len(questComp.CompletedQuests))
	}
}

func TestEventQuestSystem_GetEventNPCDialogOptions(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1)
	questComp := NewEventQuestComponent(3)

	// Add available quests
	questComp.AvailableQuests = append(questComp.AvailableQuests,
		EventQuestDefinition{ID: "quest1", Name: "Gather Blossoms"},
		EventQuestDefinition{ID: "quest2", Name: "Find the Grove"},
	)

	// Add a completed active quest
	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:   "quest3",
			Name: "Defeat the Guardian",
			Objectives: []quest.Objective{
				{Required: 1},
			},
		},
		Progress: map[int]int{0: 1}, // Complete
	})

	playerEntity.AddComponent(questComp)

	npcEntity := NewEntity(2)

	options := system.GetEventNPCDialogOptions(playerEntity, npcEntity)

	// Should have 2 available + 1 completed = 3 options
	if len(options) != 3 {
		t.Errorf("Expected 3 dialog options, got %d", len(options))
	}

	// Check option types
	offerCount := 0
	completeCount := 0
	for _, opt := range options {
		if opt.Action == ActionOfferEventQuest {
			offerCount++
		} else if opt.Action == ActionCompleteEventQuest {
			completeCount++
		}
	}

	if offerCount != 2 {
		t.Errorf("Expected 2 offer options, got %d", offerCount)
	}
	if completeCount != 1 {
		t.Errorf("Expected 1 complete option, got %d", completeCount)
	}
}

func TestEventQuestSystem_GetEventNPCDialogOptions_NoComponent(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1) // No event quest component
	npcEntity := NewEntity(2)

	options := system.GetEventNPCDialogOptions(playerEntity, npcEntity)

	if len(options) != 0 {
		t.Errorf("Expected 0 dialog options without component, got %d", len(options))
	}
}

func TestEventQuestSystem_AcceptEventQuest(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Date(2025, 3, 25, 12, 0, 0, 0, time.UTC)}
	system := NewEventQuestSystem(world, clock)

	worldEntity := NewEntity(1)
	eventComp := NewSeasonalEventComponent(12345, false)
	eventComp.CurrentTime = clock.Now()
	eventComp.ActiveEvents = append(eventComp.ActiveEvents, EventInstance{
		Definition: EventDefinition{
			ID:    "spring_festival",
			Name:  "Bloom Festival",
			Theme: EventThemeSpring,
			Seed:  12345,
		},
		EndTime: clock.Now().Add(7 * 24 * time.Hour),
		Phase:   EventPhaseActive,
	})
	worldEntity.AddComponent(eventComp)

	playerEntity := NewEntity(2)
	questComp := NewEventQuestComponent(3)
	questComp.AvailableQuests = append(questComp.AvailableQuests, EventQuestDefinition{
		ID:      "test_quest",
		EventID: "spring_festival",
		Name:    "Test Quest",
	})
	playerEntity.AddComponent(questComp)

	result := system.AcceptEventQuest(playerEntity, worldEntity, "test_quest")

	if !result {
		t.Error("AcceptEventQuest should succeed")
	}

	if len(questComp.ActiveQuests) != 1 {
		t.Errorf("Expected 1 active quest, got %d", len(questComp.ActiveQuests))
	}

	// Check expiration time is set to event end time
	active := questComp.ActiveQuests[0]
	expectedEnd := eventComp.ActiveEvents[0].EndTime
	if !active.ExpiresAt.Equal(expectedEnd) {
		t.Errorf("ExpiresAt = %v, want %v", active.ExpiresAt, expectedEnd)
	}
}

func TestEventQuestSystem_TurnInEventQuest(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1)
	questComp := NewEventQuestComponent(3)

	// Add a completed quest
	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID:   "test_quest",
			Name: "Test Quest",
			Objectives: []quest.Objective{
				{Required: 5},
			},
			Reward: quest.Reward{XP: 100, Gold: 50},
		},
		Progress: map[int]int{0: 5},
	})
	playerEntity.AddComponent(questComp)

	success, reward := system.TurnInEventQuest(playerEntity, "test_quest")

	if !success {
		t.Error("TurnInEventQuest should succeed")
	}

	if reward == nil {
		t.Fatal("Expected reward definition")
	}

	if reward.Reward.XP != 100 {
		t.Errorf("Reward XP = %d, want 100", reward.Reward.XP)
	}

	if len(questComp.ActiveQuests) != 0 {
		t.Errorf("Expected 0 active quests, got %d", len(questComp.ActiveQuests))
	}

	if len(questComp.CompletedQuests) != 1 {
		t.Errorf("Expected 1 completed quest, got %d", len(questComp.CompletedQuests))
	}
}

func TestEventQuestSystem_TurnInEventQuest_NotComplete(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1)
	questComp := NewEventQuestComponent(3)

	// Add an incomplete quest
	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID: "test_quest",
			Objectives: []quest.Objective{
				{Required: 5},
			},
		},
		Progress: map[int]int{0: 3}, // Not complete
	})
	playerEntity.AddComponent(questComp)

	success, reward := system.TurnInEventQuest(playerEntity, "test_quest")

	if success {
		t.Error("TurnInEventQuest should fail for incomplete quest")
	}

	if reward != nil {
		t.Error("Should not return reward for incomplete quest")
	}
}

func TestEventQuestSystem_GetEventQuestProgress(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1)
	questComp := NewEventQuestComponent(3)

	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID: "test_quest",
			Objectives: []quest.Objective{
				{Required: 5},
				{Required: 3},
			},
		},
		Progress: map[int]int{0: 3, 1: 2},
	})
	playerEntity.AddComponent(questComp)

	current, required, complete := system.GetEventQuestProgress(playerEntity, "test_quest")

	if current != 5 { // 3 + 2
		t.Errorf("Current = %d, want 5", current)
	}

	if required != 8 { // 5 + 3
		t.Errorf("Required = %d, want 8", required)
	}

	if complete {
		t.Error("Quest should not be complete")
	}
}

func TestEventQuestSystem_GetEventQuestProgress_Complete(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1)
	questComp := NewEventQuestComponent(3)

	questComp.ActiveQuests = append(questComp.ActiveQuests, EventQuestInstance{
		Definition: EventQuestDefinition{
			ID: "test_quest",
			Objectives: []quest.Objective{
				{Required: 5},
			},
		},
		Progress: map[int]int{0: 5},
	})
	playerEntity.AddComponent(questComp)

	current, required, complete := system.GetEventQuestProgress(playerEntity, "test_quest")

	if current != 5 {
		t.Errorf("Current = %d, want 5", current)
	}

	if required != 5 {
		t.Errorf("Required = %d, want 5", required)
	}

	if !complete {
		t.Error("Quest should be complete")
	}
}

func TestEventQuestSystem_GetEventQuestProgress_NotFound(t *testing.T) {
	world := NewWorld()
	clock := &mockEventQuestClock{currentTime: time.Now()}
	system := NewEventQuestSystem(world, clock)

	playerEntity := NewEntity(1)
	questComp := NewEventQuestComponent(3)
	playerEntity.AddComponent(questComp)

	current, required, complete := system.GetEventQuestProgress(playerEntity, "nonexistent")

	if current != 0 || required != 0 || complete {
		t.Error("Should return zeros for nonexistent quest")
	}
}
