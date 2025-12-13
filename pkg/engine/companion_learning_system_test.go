package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/companion/learning"
)

func TestNewCompanionLearningSystem(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	if cls == nil {
		t.Fatal("Expected non-nil CompanionLearningSystem")
	}

	if cls.world != world {
		t.Error("Expected system to reference the world")
	}

	if cls.learningSystem == nil {
		t.Error("Expected learning system to be initialized")
	}
}

func TestAddCompanionLearning(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorAggressive,
	})
	world.Update(0) // Process entity addition

	err := cls.AddCompanionLearning(companion.ID, 1.2)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Expected companion to have learning component")
	}

	comp := learningCompRaw.(*learning.CompanionLearningComponent)
	if comp.LearningRate != 1.2 {
		t.Errorf("Expected learning rate 1.2, got %f", comp.LearningRate)
	}
}

func TestAddCompanionLearningInvalidEntity(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	err := cls.AddCompanionLearning(99999, 1.0)
	if err == nil {
		t.Error("Expected error for invalid entity ID")
	}
}

func TestProcessCombatAction(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorAggressive,
	})

	err := cls.AddCompanionLearning(companion.ID, 1.0)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	err = cls.ProcessCombatAction(companion.ID, true, true)
	if err != nil {
		t.Fatalf("Failed to process combat action: %v", err)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Failed to get learning component")
	}
	learningComp, ok := learningCompRaw.(*learning.CompanionLearningComponent)
	if !ok {
		t.Fatal("Failed to type assert learning component")
	}

	skill, ok := learningComp.SkillTree.Skills["Basic Attack"]
	if !ok {
		t.Fatal("Expected Basic Attack skill to exist")
	}

	if skill.Experience == 0 {
		t.Error("Expected skill to gain experience from combat")
	}

	if learningComp.Memory.TotalEvents == 0 {
		t.Error("Expected combat event to be recorded in memory")
	}
}

func TestProcessCombatActionInvalidEntity(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	err := cls.ProcessCombatAction(99999, true, true)
	if err == nil {
		t.Error("Expected error for invalid entity ID")
	}
}

func TestProcessSocialInteraction(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorPassive,
	})

	err := cls.AddCompanionLearning(companion.ID, 1.0)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	err = cls.ProcessSocialInteraction(companion.ID, "player_123", true)
	if err != nil {
		t.Fatalf("Failed to process social interaction: %v", err)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Failed to get learning component")
	}
	learningComp := learningCompRaw.(*learning.CompanionLearningComponent)

	skill, ok := learningComp.SkillTree.Skills["Persuasion"]
	if !ok {
		t.Fatal("Expected Persuasion skill to exist")
	}

	if skill.Experience == 0 {
		t.Error("Expected skill to gain experience from social interaction")
	}

	if learningComp.Memory.TotalEvents == 0 {
		t.Error("Expected social event to be recorded in memory")
	}
}

func TestGetCompanionSkillBonus(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorAggressive,
	})

	err := cls.AddCompanionLearning(companion.ID, 1.0)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	bonus := cls.GetCompanionSkillBonus(companion.ID, "Basic Attack")
	if bonus != 1.0 {
		t.Errorf("Expected base bonus 1.0, got %f", bonus)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Failed to get learning component")
	}
	learningComp := learningCompRaw.(*learning.CompanionLearningComponent)
	learningComp.SkillTree.Skills["Basic Attack"].Level = 5

	bonus = cls.GetCompanionSkillBonus(companion.ID, "Basic Attack")
	expectedBonus := 1.5
	if bonus != expectedBonus {
		t.Errorf("Expected bonus %f for level 5, got %f", expectedBonus, bonus)
	}
}

func TestGetCompanionSkillBonusInvalidEntity(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	bonus := cls.GetCompanionSkillBonus(99999, "Basic Attack")
	if bonus != 1.0 {
		t.Errorf("Expected default bonus 1.0 for invalid entity, got %f", bonus)
	}
}

func TestGetPersonalityInfluence(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorAggressive,
	})

	err := cls.AddCompanionLearning(companion.ID, 1.0)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	influence := cls.GetPersonalityInfluence(companion.ID, learning.TraitAggressive)
	if influence != 0.5 {
		t.Errorf("Expected default influence 0.5, got %f", influence)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Failed to get learning component")
	}
	learningComp := learningCompRaw.(*learning.CompanionLearningComponent)
	learningComp.Personality.AdjustTrait(learning.TraitAggressive, 0.3, "test")

	influence = cls.GetPersonalityInfluence(companion.ID, learning.TraitAggressive)
	if influence < 0.7 || influence > 0.9 {
		t.Errorf("Expected influence around 0.8, got %f", influence)
	}
}

func TestUpdateProcessesBehaviorChanges(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)
	world.AddSystem(cls)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorAggressive,
	})
	companion.AddComponent(&HealthComponent{
		Current: 100,
		Max:     100,
	})

	err := cls.AddCompanionLearning(companion.ID, 1.0)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Failed to get learning component")
	}
	learningComp := learningCompRaw.(*learning.CompanionLearningComponent)
	initialAggressive := learningComp.Personality.Traits[learning.TraitAggressive]

	world.Update(1.0)

	newAggressive := learningComp.Personality.Traits[learning.TraitAggressive]
	if newAggressive <= initialAggressive {
		t.Error("Expected aggressive trait to increase with aggressive behavior")
	}
}

func TestUpdateProcessesLowHealth(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)
	world.AddSystem(cls)

	companion := world.CreateEntity()
	world.Update(0) // Process entity addition
	companion.AddComponent(&CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
		Behavior:      BehaviorPassive,
	})
	companion.AddComponent(&HealthComponent{
		Current: 20,
		Max:     100,
	})

	err := cls.AddCompanionLearning(companion.ID, 1.0)
	if err != nil {
		t.Fatalf("Failed to add companion learning: %v", err)
	}

	learningCompRaw, ok := companion.GetComponent("companion_learning")
	if !ok {
		t.Fatal("Failed to get learning component")
	}
	learningComp := learningCompRaw.(*learning.CompanionLearningComponent)
	initialCautious := learningComp.Personality.Traits[learning.TraitCautious]

	world.Update(1.0)

	newCautious := learningComp.Personality.Traits[learning.TraitCautious]
	if newCautious <= initialCautious {
		t.Error("Expected cautious trait to increase when health is low")
	}
}

func TestGetManager(t *testing.T) {
	world := NewWorld()
	cls := NewCompanionLearningSystem(world)

	manager := cls.GetManager()
	if manager == nil {
		t.Error("Expected non-nil manager")
	}
}
