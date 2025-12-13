package narrative_world

import (
	"testing"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
)

func TestSystem_Creation(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(12345)

	system := NewSystem(world, seed)

	if system == nil {
		t.Fatal("NewSystem returned nil")
	}
	if system.world != world {
		t.Error("System world not set correctly")
	}
	if system.manager == nil {
		t.Error("System manager is nil")
	}
	if system.logger == nil {
		t.Error("System logger is nil")
	}
	if system.seed != seed {
		t.Errorf("System seed mismatch: expected %d, got %d", seed, system.seed)
	}
}

func TestSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(12345)
	system := NewSystem(world, seed)

	// Create companion entity with high loyalty
	companion := world.CreateEntity()
	companionComp := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Level:         5,
		Loyalty:       0.8,
	}
	companion.AddComponent(companionComp)

	entities := []*engine.Entity{companion}
	system.Update(entities, 0.016)

	// Verify quest was generated (loyalty >= 0.7)
	quests := system.GetActiveQuests(companion.ID)
	if len(quests) == 0 {
		t.Error("Expected personal quest to be generated for loyal companion")
	}
}

func TestSystem_ConflictDetection(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(54321)
	system := NewSystem(world, seed)

	// Create two companions with opposing personalities
	companion1 := world.CreateEntity()
	comp1 := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Level:         5,
		Loyalty:       0.8,
	}
	companion1.AddComponent(comp1)

	personality1 := learning.NewPersonalityEvolution()
	personality1.Traits[learning.TraitAggressive] = 0.9
	personality1.Traits[learning.TraitPacifist] = 0.1
	learning1 := &learning.CompanionLearningComponent{
		SkillTree:   &learning.SkillProgression{Skills: make(map[string]*learning.Skill)},
		Personality: personality1,
		Memory:      learning.NewEventMemory(100),
	}
	companion1.AddComponent(learning1)

	companion2 := world.CreateEntity()
	comp2 := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypeHireling,
		Level:         5,
		Loyalty:       0.8,
	}
	companion2.AddComponent(comp2)

	personality2 := learning.NewPersonalityEvolution()
	personality2.Traits[learning.TraitAggressive] = 0.1
	personality2.Traits[learning.TraitPacifist] = 0.9
	learning2 := &learning.CompanionLearningComponent{
		SkillTree:   &learning.SkillProgression{Skills: make(map[string]*learning.Skill)},
		Personality: personality2,
		Memory:      learning.NewEventMemory(100),
	}
	companion2.AddComponent(learning2)

	entities := []*engine.Entity{companion1, companion2}

	// Run multiple updates to trigger conflict detection (probabilistic)
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.016)
	}

	// Check if conflict was detected (probabilistic, so just log result)
	conflicts := system.GetActiveConflicts()
	t.Logf("Detected %d conflicts after 10 updates", len(conflicts))
}

func TestSystem_RecordEvents(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(99999)
	system := NewSystem(world, seed)

	companionID := uint64(12345)

	// Record various events
	system.RecordCombatEvent(companionID, "Defeated dragon")
	system.RecordBondingEvent(companionID, "Shared campfire stories")

	// Verify events were recorded
	memoryCount := system.manager.GetMemoryCount(companionID)
	if memoryCount != 2 {
		t.Errorf("Expected 2 memory events, got %d", memoryCount)
	}
}

func TestSystem_DialogueContext(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(88888)
	system := NewSystem(world, seed)

	companionID := uint64(67890)

	// Record some events
	system.RecordCombatEvent(companionID, "Battle 1")
	system.RecordBondingEvent(companionID, "Bonding moment")
	system.RecordCombatEvent(companionID, "Battle 2")

	// Get dialogue context
	context := system.GetDialogueContext(companionID)

	if context == nil {
		t.Fatal("GetDialogueContext returned nil")
	}
	if len(context.RecentEvents) == 0 {
		t.Error("Expected recent events in dialogue context")
	}
}

func TestSystem_QuestManagement(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(77777)
	system := NewSystem(world, seed)

	// Create companion with high loyalty
	companion := world.CreateEntity()
	companionComp := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Level:         10,
		Loyalty:       0.9,
	}
	companion.AddComponent(companionComp)

	entities := []*engine.Entity{companion}
	system.Update(entities, 0.016)

	// Get active quests
	quests := system.GetActiveQuests(companion.ID)
	if len(quests) == 0 {
		t.Fatal("Expected at least one quest for loyal companion")
	}

	quest := quests[0]

	// Update quest objective
	err := system.UpdateQuestObjective(companion.ID, quest.QuestID, 0, 5)
	if err != nil {
		t.Errorf("Failed to update quest objective: %v", err)
	}

	// Mark all objectives completed
	for i := range quest.Objectives {
		quest.Objectives[i].Progress = quest.Objectives[i].Required
		quest.Objectives[i].Completed = true
	}

	// Complete quest
	err = system.CompleteQuest(companion.ID, quest.QuestID)
	if err != nil {
		t.Errorf("Failed to complete quest: %v", err)
	}
}

func TestSystem_LoyaltyThreshold(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(66666)
	system := NewSystem(world, seed)

	tests := []struct {
		name           string
		loyalty        float64
		expectQuestGen bool
	}{
		{"Low loyalty", 0.5, false},
		{"Threshold loyalty", 0.7, true},
		{"High loyalty", 0.9, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companion := world.CreateEntity()
			companionComp := &engine.CompanionComponent{
				CompanionType: engine.CompanionTypePet,
				Level:         5,
				Loyalty:       tt.loyalty,
			}
			companion.AddComponent(companionComp)

			entities := []*engine.Entity{companion}
			system.Update(entities, 0.016)

			quests := system.GetActiveQuests(companion.ID)
			hasQuest := len(quests) > 0

			if hasQuest != tt.expectQuestGen {
				t.Errorf("Loyalty %f: expected quest generation = %v, got %v", tt.loyalty, tt.expectQuestGen, hasQuest)
			}
		})
	}
}

func TestSystem_QuestLimit(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(55555)
	system := NewSystem(world, seed)

	// Create companion with high loyalty
	companion := world.CreateEntity()
	companionComp := &engine.CompanionComponent{
		CompanionType: engine.CompanionTypePet,
		Level:         10,
		Loyalty:       0.95,
	}
	companion.AddComponent(companionComp)

	entities := []*engine.Entity{companion}

	// Run multiple updates to try generating quests
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.016)
	}

	// Check quest limit (should be capped at 2)
	quests := system.GetActiveQuests(companion.ID)
	if len(quests) > 2 {
		t.Errorf("Expected max 2 active quests, got %d", len(quests))
	}
}

func TestSystem_GetManager(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(44444)
	system := NewSystem(world, seed)

	manager := system.GetManager()
	if manager == nil {
		t.Error("GetManager returned nil")
	}
	if manager != system.manager {
		t.Error("GetManager returned different manager instance")
	}
}
