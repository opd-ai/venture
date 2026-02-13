package narrative_world

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

func TestStoryEventManager_Serialize_Deserialize(t *testing.T) {
	// Create manager with test data
	manager := NewStoryEventManager(12345)

	// Add some memory events
	manager.RecordMemory(1, EventTypeCombat, "Fought together against bandits")
	manager.RecordMemory(1, EventTypeTreasure, "Found a hidden chest")
	manager.RecordMemory(2, EventTypeBonding, "Shared a meal by the campfire")

	// Add a conflict
	manager.conflicts = append(manager.conflicts, CompanionConflict{
		Companion1:     1,
		Companion2:     2,
		ConflictType:   ConflictPersonality,
		Description:    "They have conflicting personalities",
		Severity:       0.5,
		Active:         true,
		TimeSinceStart: time.Hour * 24,
	})

	// Add a cross-companion story
	manager.crossStories = append(manager.crossStories, &CrossCompanionStory{
		StoryID:      "story-1",
		Title:        "Test Story",
		Description:  "A test cross-companion story",
		Participants: []uint64{1, 2},
		Events: []MemoryEvent{
			{
				Timestamp:    now(),
				Type:         EventTypeBonding,
				Description:  "Bonding event",
				Participants: []uint64{1, 2},
				Location:     "tavern",
				Importance:   0.7,
			},
		},
		Outcome: OutcomeFriendship,
		Active:  true,
	})

	// Serialize
	data, err := manager.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Serialized data is empty")
	}

	// Create new manager and deserialize
	newManager := NewStoryEventManager(0) // Seed will be restored
	err = newManager.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify seed restored
	if newManager.seed != manager.seed {
		t.Errorf("Seed mismatch: got %d, want %d", newManager.seed, manager.seed)
	}

	// Verify settings restored
	if newManager.conflictChance != manager.conflictChance {
		t.Errorf("ConflictChance mismatch: got %f, want %f", newManager.conflictChance, manager.conflictChance)
	}
	if newManager.maxMemoryEvents != manager.maxMemoryEvents {
		t.Errorf("MaxMemoryEvents mismatch: got %d, want %d", newManager.maxMemoryEvents, manager.maxMemoryEvents)
	}

	// Verify memories restored
	if len(newManager.memories) != len(manager.memories) {
		t.Errorf("Memories count mismatch: got %d, want %d", len(newManager.memories), len(manager.memories))
	}
	for id, mem := range manager.memories {
		newMem, exists := newManager.memories[id]
		if !exists {
			t.Errorf("Memory for companion %d not restored", id)
			continue
		}
		if len(newMem.Events) != len(mem.Events) {
			t.Errorf("Memory events count mismatch for companion %d: got %d, want %d", id, len(newMem.Events), len(mem.Events))
		}
		if newMem.TotalEvents != mem.TotalEvents {
			t.Errorf("TotalEvents mismatch for companion %d: got %d, want %d", id, newMem.TotalEvents, mem.TotalEvents)
		}
	}

	// Verify conflicts restored
	if len(newManager.conflicts) != len(manager.conflicts) {
		t.Errorf("Conflicts count mismatch: got %d, want %d", len(newManager.conflicts), len(manager.conflicts))
	}
	if len(newManager.conflicts) > 0 {
		if newManager.conflicts[0].Companion1 != manager.conflicts[0].Companion1 {
			t.Errorf("Conflict Companion1 mismatch: got %d, want %d", newManager.conflicts[0].Companion1, manager.conflicts[0].Companion1)
		}
		if newManager.conflicts[0].Severity != manager.conflicts[0].Severity {
			t.Errorf("Conflict Severity mismatch: got %f, want %f", newManager.conflicts[0].Severity, manager.conflicts[0].Severity)
		}
		if newManager.conflicts[0].TimeSinceStart != manager.conflicts[0].TimeSinceStart {
			t.Errorf("Conflict TimeSinceStart mismatch: got %v, want %v", newManager.conflicts[0].TimeSinceStart, manager.conflicts[0].TimeSinceStart)
		}
	}

	// Verify cross stories restored
	if len(newManager.crossStories) != len(manager.crossStories) {
		t.Errorf("CrossStories count mismatch: got %d, want %d", len(newManager.crossStories), len(manager.crossStories))
	}
	if len(newManager.crossStories) > 0 {
		if newManager.crossStories[0].StoryID != manager.crossStories[0].StoryID {
			t.Errorf("CrossStory StoryID mismatch: got %s, want %s", newManager.crossStories[0].StoryID, manager.crossStories[0].StoryID)
		}
		if newManager.crossStories[0].Outcome != manager.crossStories[0].Outcome {
			t.Errorf("CrossStory Outcome mismatch: got %d, want %d", newManager.crossStories[0].Outcome, manager.crossStories[0].Outcome)
		}
	}
}

func TestSerializeMemoryEvent(t *testing.T) {
	event := MemoryEvent{
		Timestamp:    1234567890,
		Type:         EventTypeCombat,
		Description:  "Test event",
		Participants: []uint64{1, 2, 3},
		Location:     "dungeon",
		Importance:   0.8,
	}

	serialized := serializeMemoryEvent(event)

	if serialized.Timestamp != event.Timestamp {
		t.Errorf("Timestamp mismatch: got %d, want %d", serialized.Timestamp, event.Timestamp)
	}
	if serialized.Type != int(event.Type) {
		t.Errorf("Type mismatch: got %d, want %d", serialized.Type, int(event.Type))
	}
	if serialized.Description != event.Description {
		t.Errorf("Description mismatch: got %s, want %s", serialized.Description, event.Description)
	}
	if serialized.Importance != event.Importance {
		t.Errorf("Importance mismatch: got %f, want %f", serialized.Importance, event.Importance)
	}

	// Test round-trip
	restored := deserializeMemoryEvent(serialized)
	if restored.Timestamp != event.Timestamp {
		t.Errorf("Round-trip Timestamp mismatch: got %d, want %d", restored.Timestamp, event.Timestamp)
	}
	if restored.Type != event.Type {
		t.Errorf("Round-trip Type mismatch: got %d, want %d", restored.Type, event.Type)
	}
}

func TestSerializePersonalQuest(t *testing.T) {
	quest := &PersonalQuest{
		QuestID:       "quest-123",
		CompanionID:   42,
		CompanionType: engine.CompanionTypePet,
		Title:         "Lost Home",
		Description:   "Find the companion's origin",
		Objectives: []QuestObjective{
			{Description: "Visit the forest", Type: ObjectiveVisit, Target: "dark_forest", Progress: 0, Required: 1, Completed: false},
			{Description: "Defeat the guardian", Type: ObjectiveDefeat, Target: "forest_guardian", Progress: 2, Required: 3, Completed: false},
		},
		UnlockLoyalty: 0.7,
		Completed:     false,
		Started:       true,
		Consequences: []Consequence{
			{Type: ConsequenceLoyaltyChange, Description: "Loyalty will change", Permanent: true, Severity: 0.6},
		},
	}

	serialized := serializePersonalQuest(quest)

	if serialized.QuestID != quest.QuestID {
		t.Errorf("QuestID mismatch: got %s, want %s", serialized.QuestID, quest.QuestID)
	}
	if serialized.CompanionID != quest.CompanionID {
		t.Errorf("CompanionID mismatch: got %d, want %d", serialized.CompanionID, quest.CompanionID)
	}
	if len(serialized.Objectives) != len(quest.Objectives) {
		t.Errorf("Objectives count mismatch: got %d, want %d", len(serialized.Objectives), len(quest.Objectives))
	}

	// Test round-trip
	restored := deserializePersonalQuest(serialized)
	if restored.QuestID != quest.QuestID {
		t.Errorf("Round-trip QuestID mismatch: got %s, want %s", restored.QuestID, quest.QuestID)
	}
	if restored.CompanionType != quest.CompanionType {
		t.Errorf("Round-trip CompanionType mismatch: got %d, want %d", restored.CompanionType, quest.CompanionType)
	}
	if restored.Started != quest.Started {
		t.Errorf("Round-trip Started mismatch: got %v, want %v", restored.Started, quest.Started)
	}
	if len(restored.Objectives) != len(quest.Objectives) {
		t.Errorf("Round-trip Objectives count mismatch: got %d, want %d", len(restored.Objectives), len(quest.Objectives))
	}
	if len(restored.Objectives) > 1 {
		if restored.Objectives[1].Progress != quest.Objectives[1].Progress {
			t.Errorf("Round-trip Objective Progress mismatch: got %d, want %d", restored.Objectives[1].Progress, quest.Objectives[1].Progress)
		}
	}
}

func TestSerializeCompanionConflict(t *testing.T) {
	conflict := CompanionConflict{
		Companion1:     1,
		Companion2:     2,
		ConflictType:   ConflictRivalry,
		Description:    "They are rivals",
		Severity:       0.75,
		Active:         true,
		TimeSinceStart: time.Hour * 48,
	}

	serialized := serializeCompanionConflict(conflict)

	if serialized.Companion1 != conflict.Companion1 {
		t.Errorf("Companion1 mismatch: got %d, want %d", serialized.Companion1, conflict.Companion1)
	}
	if serialized.ConflictType != int(conflict.ConflictType) {
		t.Errorf("ConflictType mismatch: got %d, want %d", serialized.ConflictType, int(conflict.ConflictType))
	}
	if serialized.TimeSinceStartNs != conflict.TimeSinceStart.Nanoseconds() {
		t.Errorf("TimeSinceStartNs mismatch: got %d, want %d", serialized.TimeSinceStartNs, conflict.TimeSinceStart.Nanoseconds())
	}

	// Test round-trip
	restored := deserializeCompanionConflict(serialized)
	if restored.Companion1 != conflict.Companion1 {
		t.Errorf("Round-trip Companion1 mismatch: got %d, want %d", restored.Companion1, conflict.Companion1)
	}
	if restored.ConflictType != conflict.ConflictType {
		t.Errorf("Round-trip ConflictType mismatch: got %d, want %d", restored.ConflictType, conflict.ConflictType)
	}
	if restored.TimeSinceStart != conflict.TimeSinceStart {
		t.Errorf("Round-trip TimeSinceStart mismatch: got %v, want %v", restored.TimeSinceStart, conflict.TimeSinceStart)
	}
}

func TestSerializeCrossCompanionStory(t *testing.T) {
	story := &CrossCompanionStory{
		StoryID:      "story-456",
		Title:        "Alliance Formed",
		Description:  "Two companions form an alliance",
		Participants: []uint64{10, 20},
		Events: []MemoryEvent{
			{Timestamp: 100, Type: EventTypeBonding, Description: "Initial meeting", Participants: []uint64{10, 20}, Location: "tavern", Importance: 0.5},
		},
		Outcome: OutcomeRomance,
		Active:  false,
	}

	serialized := serializeCrossCompanionStory(story)

	if serialized.StoryID != story.StoryID {
		t.Errorf("StoryID mismatch: got %s, want %s", serialized.StoryID, story.StoryID)
	}
	if serialized.Outcome != int(story.Outcome) {
		t.Errorf("Outcome mismatch: got %d, want %d", serialized.Outcome, int(story.Outcome))
	}
	if len(serialized.Participants) != len(story.Participants) {
		t.Errorf("Participants count mismatch: got %d, want %d", len(serialized.Participants), len(story.Participants))
	}

	// Test round-trip
	restored := deserializeCrossCompanionStory(serialized)
	if restored.StoryID != story.StoryID {
		t.Errorf("Round-trip StoryID mismatch: got %s, want %s", restored.StoryID, story.StoryID)
	}
	if restored.Outcome != story.Outcome {
		t.Errorf("Round-trip Outcome mismatch: got %d, want %d", restored.Outcome, story.Outcome)
	}
	if restored.Active != story.Active {
		t.Errorf("Round-trip Active mismatch: got %v, want %v", restored.Active, story.Active)
	}
}

func TestDeserialize_EmptyData(t *testing.T) {
	manager := NewStoryEventManager(12345)
	err := manager.Deserialize([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
}

func TestDeserialize_InvalidJSON(t *testing.T) {
	manager := NewStoryEventManager(12345)
	err := manager.Deserialize([]byte("not valid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestSerialize_EmptyManager(t *testing.T) {
	manager := NewStoryEventManager(99999)

	data, err := manager.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	newManager := NewStoryEventManager(0)
	err = newManager.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if newManager.seed != manager.seed {
		t.Errorf("Seed mismatch: got %d, want %d", newManager.seed, manager.seed)
	}
	if len(newManager.memories) != 0 {
		t.Errorf("Expected empty memories, got %d", len(newManager.memories))
	}
	if len(newManager.conflicts) != 0 {
		t.Errorf("Expected empty conflicts, got %d", len(newManager.conflicts))
	}
}

// Benchmark serialization performance
func BenchmarkSerialize(b *testing.B) {
	manager := NewStoryEventManager(12345)
	// Add substantial data
	for i := uint64(0); i < 10; i++ {
		for j := 0; j < 50; j++ {
			manager.RecordMemory(i, EventType(j%8), "Test memory event description")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Serialize()
	}
}

func BenchmarkDeserialize(b *testing.B) {
	manager := NewStoryEventManager(12345)
	// Add substantial data
	for i := uint64(0); i < 10; i++ {
		for j := 0; j < 50; j++ {
			manager.RecordMemory(i, EventType(j%8), "Test memory event description")
		}
	}
	data, _ := manager.Serialize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newManager := NewStoryEventManager(0)
		_ = newManager.Deserialize(data)
	}
}
