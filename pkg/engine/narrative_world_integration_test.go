// Package engine provides comprehensive integration tests for NarrativeSystem + narrative_world integration.
// Tests verify companion story event recording, quest generation triggers, and memory tracking
// during narrative events (combat victories, dialogue completion, discoveries).
package engine

import (
	"testing"
)

// mockCompanionStoryProvider implements CompanionStoryProvider for testing
type mockCompanionStoryProvider struct {
	combatEvents  []mockEventRecord
	bondingEvents []mockEventRecord
}

type mockEventRecord struct {
	companionID uint64
	description string
}

func (m *mockCompanionStoryProvider) RecordCombatEvent(companionID uint64, description string) {
	m.combatEvents = append(m.combatEvents, mockEventRecord{companionID, description})
}

func (m *mockCompanionStoryProvider) RecordBondingEvent(companionID uint64, description string) {
	m.bondingEvents = append(m.bondingEvents, mockEventRecord{companionID, description})
}

// TestNarrativeSystem_CompanionStoryIntegration_Injection verifies SetCompanionStoryProvider injection
func TestNarrativeSystem_CompanionStoryIntegration_Injection(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	if ns.companionStoryProvider != nil {
		t.Error("Expected nil provider before injection")
	}

	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	if ns.companionStoryProvider == nil {
		t.Error("Expected provider after injection")
	}
}

// TestNarrativeSystem_CombatVictory_RecordsCompanionEvents verifies combat events are recorded
func TestNarrativeSystem_CombatVictory_RecordsCompanionEvents(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative component
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActSetup,
		StoryProgress: 0.3,
		EventHistory:  []NarrativeEvent{},
	})

	// Create two companion entities
	companion1 := world.CreateEntity()
	companion1.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypePet,
		Loyalty:       0.75,
	})

	companion2 := world.CreateEntity()
	companion2.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeHireling,
		Loyalty:       0.85,
	})

	// Wire mock provider
	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	// Trigger combat victory
	narrative, _ := player.GetComponent("narrative")
	ns.OnCombatVictory(narrative.(*NarrativeComponent), nil, 100, 200)

	// Verify combat events recorded for both companions
	if len(mockProvider.combatEvents) != 2 {
		t.Errorf("Expected 2 combat events, got %d", len(mockProvider.combatEvents))
	}

	foundComp1 := false
	foundComp2 := false
	for _, event := range mockProvider.combatEvents {
		if event.companionID == companion1.ID {
			foundComp1 = true
		}
		if event.companionID == companion2.ID {
			foundComp2 = true
		}
		if event.description != "Defeated enemy in combat" {
			t.Errorf("Expected combat description, got %q", event.description)
		}
	}

	if !foundComp1 {
		t.Error("Expected combat event for companion1")
	}
	if !foundComp2 {
		t.Error("Expected combat event for companion2")
	}
}

// TestNarrativeSystem_BossFight_RecordsPowerfulEnemy verifies boss combat description
func TestNarrativeSystem_BossFight_RecordsPowerfulEnemy(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActConfrontation,
		StoryProgress: 0.6,
		EventHistory:  []NarrativeEvent{},
	})

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypePet,
		Loyalty:       0.9,
	})

	// Create boss enemy with high detection range
	boss := world.CreateEntity()
	boss.AddComponent(&AIComponent{
		DetectionRange: 350.0, // Above BossDetectionRange (300.0)
	})

	// Wire mock provider
	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	// Trigger boss combat victory
	narrative, _ := player.GetComponent("narrative")
	ns.OnCombatVictory(narrative.(*NarrativeComponent), boss, 500, 600)

	// Verify powerful enemy description
	if len(mockProvider.combatEvents) != 1 {
		t.Fatalf("Expected 1 combat event, got %d", len(mockProvider.combatEvents))
	}

	if mockProvider.combatEvents[0].description != "Defeated powerful enemy" {
		t.Errorf("Expected boss description, got %q", mockProvider.combatEvents[0].description)
	}
}

// TestNarrativeSystem_DialogueComplete_RecordsBonding verifies dialogue events trigger bonding
func TestNarrativeSystem_DialogueComplete_RecordsBonding(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActSetup,
		StoryProgress: 0.2,
		EventHistory:  []NarrativeEvent{},
	})

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeSummon,
		Loyalty:       0.7,
	})

	// Wire mock provider
	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	// Trigger dialogue completion
	npc := world.CreateEntity()
	narrative, _ := player.GetComponent("narrative")
	ns.OnDialogueComplete(narrative.(*NarrativeComponent), npc, 0)

	// Verify bonding event recorded
	if len(mockProvider.bondingEvents) != 1 {
		t.Fatalf("Expected 1 bonding event, got %d", len(mockProvider.bondingEvents))
	}

	event := mockProvider.bondingEvents[0]
	if event.companionID != companion.ID {
		t.Errorf("Expected bonding event for companion %d, got %d", companion.ID, event.companionID)
	}
	if event.description != "Witnessed player dialogue interaction" {
		t.Errorf("Expected dialogue description, got %q", event.description)
	}
}

// TestNarrativeSystem_NoProvider_GracefulFallback verifies system works without provider
func TestNarrativeSystem_NoProvider_GracefulFallback(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActSetup,
		StoryProgress: 0.1,
		EventHistory:  []NarrativeEvent{},
	})

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypePet,
		Loyalty:       0.8,
	})

	// No provider injected - should not panic
	narrative, _ := player.GetComponent("narrative")
	ns.OnCombatVictory(narrative.(*NarrativeComponent), nil, 10, 20)
	ns.OnDialogueComplete(narrative.(*NarrativeComponent), nil, 0)

	// Verify narrative events still triggered (base system functionality)
	narrativeComp := narrative.(*NarrativeComponent)
	if len(narrativeComp.EventHistory) != 2 {
		t.Errorf("Expected 2 narrative events, got %d", len(narrativeComp.EventHistory))
	}
}

// TestNarrativeSystem_MultipleCompanions_AllRecorded verifies all companions get events
func TestNarrativeSystem_MultipleCompanions_AllRecorded(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActConfrontation,
		StoryProgress: 0.5,
		EventHistory:  []NarrativeEvent{},
	})

	// Create 5 companions (max party size)
	companionIDs := make([]uint64, 5)
	for i := 0; i < 5; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			CompanionType: CompanionTypePet,
			Loyalty:       0.75 + float64(i)*0.05,
		})
		companionIDs[i] = companion.ID
	}

	// Wire mock provider
	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	// Trigger combat victory
	narrative, _ := player.GetComponent("narrative")
	ns.OnCombatVictory(narrative.(*NarrativeComponent), nil, 300, 400)

	// Verify all 5 companions recorded combat events
	if len(mockProvider.combatEvents) != 5 {
		t.Fatalf("Expected 5 combat events, got %d", len(mockProvider.combatEvents))
	}

	// Verify all companion IDs present
	for i, expectedID := range companionIDs {
		found := false
		for _, event := range mockProvider.combatEvents {
			if event.companionID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Companion %d (ID %d) missing combat event", i, expectedID)
		}
	}
}

// TestNarrativeSystem_NoCompanions_NoEvents verifies no events when no companions present
func TestNarrativeSystem_NoCompanions_NoEvents(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActSetup,
		StoryProgress: 0.1,
		EventHistory:  []NarrativeEvent{},
	})

	// Wire mock provider
	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	// Trigger combat without any companions
	narrative, _ := player.GetComponent("narrative")
	ns.OnCombatVictory(narrative.(*NarrativeComponent), nil, 50, 75)

	// Verify no companion events recorded
	if len(mockProvider.combatEvents) != 0 {
		t.Errorf("Expected 0 combat events without companions, got %d", len(mockProvider.combatEvents))
	}
}

// TestNarrativeSystem_MixedEntityTypes_OnlyCompanions verifies only companions get events
func TestNarrativeSystem_MixedEntityTypes_OnlyCompanions(t *testing.T) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActSetup,
		StoryProgress: 0.2,
		EventHistory:  []NarrativeEvent{},
	})

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypePet,
		Loyalty:       0.8,
	})

	// Create non-companion entities
	enemy := world.CreateEntity()
	enemy.AddComponent(&AIComponent{
		DetectionRange: 100,
	})

	npc := world.CreateEntity()
	npc.AddComponent(&PositionComponent{X: 50, Y: 50})

	// Wire mock provider
	mockProvider := &mockCompanionStoryProvider{}
	ns.SetCompanionStoryProvider(mockProvider)

	// Trigger combat victory
	narrative, _ := player.GetComponent("narrative")
	ns.OnCombatVictory(narrative.(*NarrativeComponent), enemy, 100, 150)

	// Verify only the 1 companion entity recorded event
	if len(mockProvider.combatEvents) != 1 {
		t.Fatalf("Expected 1 combat event, got %d", len(mockProvider.combatEvents))
	}

	if mockProvider.combatEvents[0].companionID != companion.ID {
		t.Errorf("Expected event for companion %d, got %d", companion.ID, mockProvider.combatEvents[0].companionID)
	}
}

// Benchmark companion story recording performance
func BenchmarkNarrativeSystem_CombatWithCompanions(b *testing.B) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActConfrontation,
		StoryProgress: 0.5,
		EventHistory:  make([]NarrativeEvent, 0, 100),
	})

	// Create 3 companions
	for i := 0; i < 3; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			CompanionType: CompanionTypePet,
			Loyalty:       0.8,
		})
	}

	mockProvider := &mockCompanionStoryProvider{
		combatEvents: make([]mockEventRecord, 0, b.N*3),
	}
	ns.SetCompanionStoryProvider(mockProvider)

	narrative, _ := player.GetComponent("narrative")
	narrativeComp := narrative.(*NarrativeComponent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.OnCombatVictory(narrativeComp, nil, 100, 200)
	}
}

// Benchmark baseline without provider for comparison
func BenchmarkNarrativeSystem_CombatWithoutProvider(b *testing.B) {
	world := NewWorld()
	ns := NewNarrativeSystem(world)

	// Create player with narrative
	player := world.CreateEntity()
	player.AddComponent(&NarrativeComponent{
		CurrentAct:    ActConfrontation,
		StoryProgress: 0.5,
		EventHistory:  make([]NarrativeEvent, 0, 100),
	})

	// Create 3 companions
	for i := 0; i < 3; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			CompanionType: CompanionTypePet,
			Loyalty:       0.8,
		})
	}

	// No provider - baseline performance
	narrative, _ := player.GetComponent("narrative")
	narrativeComp := narrative.(*NarrativeComponent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.OnCombatVictory(narrativeComp, nil, 100, 200)
	}
}
