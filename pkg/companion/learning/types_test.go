package learning

import (
	"testing"
)

// TestCompanionLearningComponent_Type tests the ECS component Type method.
func TestCompanionLearningComponent_Type(t *testing.T) {
	comp := &CompanionLearningComponent{}
	got := comp.Type()
	want := "companion_learning"
	if got != want {
		t.Errorf("CompanionLearningComponent.Type() = %q, want %q", got, want)
	}
}

// TestEventType_String tests all EventType String() method cases.
func TestEventType_String(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		expected string
	}{
		{"EventCombat", EventCombat, "Combat"},
		{"EventDialog", EventDialog, "Dialog"},
		{"EventTrade", EventTrade, "Trade"},
		{"EventQuest", EventQuest, "Quest"},
		{"EventExploration", EventExploration, "Exploration"},
		{"EventCrafting", EventCrafting, "Crafting"},
		{"EventDeath", EventDeath, "Death"},
		{"EventRevival", EventRevival, "Revival"},
		{"EventGift", EventGift, "Gift"},
		{"EventBetrayal", EventBetrayal, "Betrayal"},
		{"Unknown", EventType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.String()
			if got != tt.expected {
				t.Errorf("EventType(%d).String() = %q, want %q", tt.event, got, tt.expected)
			}
		})
	}
}

// TestSkillType_String tests all SkillType String() method cases.
func TestSkillType_String(t *testing.T) {
	tests := []struct {
		name     string
		skill    SkillType
		expected string
	}{
		{"SkillCombat", SkillCombat, "Combat"},
		{"SkillUtility", SkillUtility, "Utility"},
		{"SkillSocial", SkillSocial, "Social"},
		{"SkillCrafting", SkillCrafting, "Crafting"},
		{"SkillMagic", SkillMagic, "Magic"},
		{"SkillDefense", SkillDefense, "Defense"},
		{"SkillHealing", SkillHealing, "Healing"},
		{"SkillStealth", SkillStealth, "Stealth"},
		{"Unknown", SkillType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skill.String()
			if got != tt.expected {
				t.Errorf("SkillType(%d).String() = %q, want %q", tt.skill, got, tt.expected)
			}
		})
	}
}

// TestPersonalityTrait_String tests all PersonalityTrait String() method cases.
func TestPersonalityTrait_String(t *testing.T) {
	tests := []struct {
		name     string
		trait    PersonalityTrait
		expected string
	}{
		{"TraitCautious", TraitCautious, "Cautious"},
		{"TraitBrave", TraitBrave, "Brave"},
		{"TraitShy", TraitShy, "Shy"},
		{"TraitOutgoing", TraitOutgoing, "Outgoing"},
		{"TraitAggressive", TraitAggressive, "Aggressive"},
		{"TraitPacifist", TraitPacifist, "Pacifist"},
		{"TraitLoyal", TraitLoyal, "Loyal"},
		{"TraitIndependent", TraitIndependent, "Independent"},
		{"TraitCurious", TraitCurious, "Curious"},
		{"TraitPractical", TraitPractical, "Practical"},
		{"Unknown", PersonalityTrait(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.trait.String()
			if got != tt.expected {
				t.Errorf("PersonalityTrait(%d).String() = %q, want %q", tt.trait, got, tt.expected)
			}
		})
	}
}

// TestCompanionLearningComponent_SerializeDeserialize tests round-trip serialization.
func TestCompanionLearningComponent_SerializeDeserialize(t *testing.T) {
	// Create a fully populated component
	manager := NewManager()
	original := manager.AddCompanion("test-companion", 1.5)

	// Add some XP and learn a skill
	original.SkillTree.AvailablePoints = 5
	err := original.SkillTree.LearnSkill("Basic Attack")
	if err != nil {
		t.Fatalf("failed to learn skill: %v", err)
	}
	err = original.SkillTree.AddExperience("Basic Attack", 50.0, original.LearningRate)
	if err != nil {
		t.Fatalf("failed to add experience: %v", err)
	}

	// Adjust personality
	original.Personality.AdjustTrait(TraitBrave, 0.3, "test adjustment")

	// Add a memory event
	ProcessCombatAction(original, true, true)

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	restored := &CompanionLearningComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify fields
	if restored.CompanionID != original.CompanionID {
		t.Errorf("CompanionID = %q, want %q", restored.CompanionID, original.CompanionID)
	}
	if restored.LearningRate != original.LearningRate {
		t.Errorf("LearningRate = %f, want %f", restored.LearningRate, original.LearningRate)
	}

	// Verify skill tree
	if restored.SkillTree == nil {
		t.Fatal("SkillTree is nil after deserialize")
	}
	if restored.SkillTree.TotalXP != original.SkillTree.TotalXP {
		t.Errorf("TotalXP = %f, want %f", restored.SkillTree.TotalXP, original.SkillTree.TotalXP)
	}
	restoredSkill := restored.SkillTree.Skills["Basic Attack"]
	if restoredSkill == nil {
		t.Fatal("Basic Attack skill not found after deserialize")
	}
	origSkill := original.SkillTree.Skills["Basic Attack"]
	if restoredSkill.Level != origSkill.Level {
		t.Errorf("Skill level = %d, want %d", restoredSkill.Level, origSkill.Level)
	}

	// Verify personality
	if restored.Personality == nil {
		t.Fatal("Personality is nil after deserialize")
	}
	restoredBrave := restored.Personality.Traits[TraitBrave]
	origBrave := original.Personality.Traits[TraitBrave]
	if restoredBrave != origBrave {
		t.Errorf("Brave trait = %f, want %f", restoredBrave, origBrave)
	}
	if len(restored.Personality.Changes) != len(original.Personality.Changes) {
		t.Errorf("Personality changes count = %d, want %d",
			len(restored.Personality.Changes), len(original.Personality.Changes))
	}

	// Verify memory
	if restored.Memory == nil {
		t.Fatal("Memory is nil after deserialize")
	}
	if restored.Memory.TotalEvents != original.Memory.TotalEvents {
		t.Errorf("TotalEvents = %d, want %d", restored.Memory.TotalEvents, original.Memory.TotalEvents)
	}
	if len(restored.Memory.Events) != len(original.Memory.Events) {
		t.Errorf("Events count = %d, want %d",
			len(restored.Memory.Events), len(original.Memory.Events))
	}
}

// TestCompanionLearningComponent_Deserialize_InvalidJSON tests error handling.
func TestCompanionLearningComponent_Deserialize_InvalidJSON(t *testing.T) {
	comp := &CompanionLearningComponent{}
	err := comp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestCompanionLearningComponent_Serialize_EmptyComponent tests serializing empty component.
func TestCompanionLearningComponent_Serialize_EmptyComponent(t *testing.T) {
	comp := &CompanionLearningComponent{
		CompanionID:  "empty-test",
		LearningRate: 1.0,
	}

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	restored := &CompanionLearningComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if restored.CompanionID != comp.CompanionID {
		t.Errorf("CompanionID = %q, want %q", restored.CompanionID, comp.CompanionID)
	}
}

// TestCompanionLearningComponent_PrerequisitesAndCostPreservation verifies that
// skill prerequisites and costs are correctly preserved through serialization.
func TestCompanionLearningComponent_PrerequisitesAndCostPreservation(t *testing.T) {
	manager := NewManager()
	original := manager.AddCompanion("prereq-test", 1.0)

	// Verify initial prerequisites exist for "Power Strike" (requires "Basic Attack")
	powerStrikeNode := original.SkillTree.SkillTree["Power Strike"]
	if powerStrikeNode == nil {
		t.Fatal("Power Strike skill node not found")
	}
	if len(powerStrikeNode.Prerequisites) == 0 {
		t.Fatal("Power Strike should have prerequisites")
	}
	if powerStrikeNode.Prerequisites[0] != "Basic Attack" {
		t.Errorf("Power Strike prerequisite = %q, want %q", powerStrikeNode.Prerequisites[0], "Basic Attack")
	}

	// Verify cost for Combat Mastery (should be 2)
	combatMasteryNode := original.SkillTree.SkillTree["Combat Mastery"]
	if combatMasteryNode == nil {
		t.Fatal("Combat Mastery skill node not found")
	}
	if combatMasteryNode.Cost != 2 {
		t.Errorf("Combat Mastery cost = %d, want %d", combatMasteryNode.Cost, 2)
	}

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	restored := &CompanionLearningComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify prerequisites are preserved
	restoredPowerStrike := restored.SkillTree.SkillTree["Power Strike"]
	if restoredPowerStrike == nil {
		t.Fatal("Power Strike skill node not found after restore")
	}
	if len(restoredPowerStrike.Prerequisites) != len(powerStrikeNode.Prerequisites) {
		t.Errorf("Restored prerequisites count = %d, want %d",
			len(restoredPowerStrike.Prerequisites), len(powerStrikeNode.Prerequisites))
	}
	if len(restoredPowerStrike.Prerequisites) > 0 && restoredPowerStrike.Prerequisites[0] != "Basic Attack" {
		t.Errorf("Restored Power Strike prerequisite = %q, want %q",
			restoredPowerStrike.Prerequisites[0], "Basic Attack")
	}

	// Verify cost is preserved
	restoredCombatMastery := restored.SkillTree.SkillTree["Combat Mastery"]
	if restoredCombatMastery == nil {
		t.Fatal("Combat Mastery skill node not found after restore")
	}
	if restoredCombatMastery.Cost != combatMasteryNode.Cost {
		t.Errorf("Restored Combat Mastery cost = %d, want %d",
			restoredCombatMastery.Cost, combatMasteryNode.Cost)
	}
}
