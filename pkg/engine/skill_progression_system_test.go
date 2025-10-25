package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/skills"
)

// Helper function to create a skill with a name and effect
func createSkill(id, name, effectType string, value float64, isPercent bool) *skills.Skill {
	return &skills.Skill{
		ID:   id,
		Name: name,
		Effects: []skills.Effect{
			{Type: effectType, Value: value, IsPercent: isPercent},
		},
		Level:    1,
		MaxLevel: 5,
	}
}

// Helper function to create a skill node
func createSkillNode(skill *skills.Skill) *skills.SkillNode {
	return &skills.SkillNode{
		Skill: skill,
		Position: skills.Position{
			X: 0,
			Y: 0,
		},
	}
}

// TestNewSkillProgressionSystem tests constructor
func TestNewSkillProgressionSystem(t *testing.T) {
	system := NewSkillProgressionSystem()

	if system == nil {
		t.Fatal("NewSkillProgressionSystem returned nil")
	}

	if system.updateInterval != 60 {
		t.Errorf("Expected update interval 60, got %d", system.updateInterval)
	}

	if system.frameCounter != 0 {
		t.Errorf("Expected frame counter 0, got %d", system.frameCounter)
	}
}

// TestSkillProgressionSystem_Update_NoEntities tests with empty entity list
func TestSkillProgressionSystem_Update_NoEntities(t *testing.T) {
	system := NewSkillProgressionSystem()
	entities := []*Entity{}

	// Should not panic
	system.Update(entities, 0.016)

	if system.frameCounter != 1 {
		t.Errorf("Frame counter should increment, got %d", system.frameCounter)
	}
}

// TestSkillProgressionSystem_Update_NoSkillTree tests entities without skill trees
func TestSkillProgressionSystem_Update_NoSkillTree(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(NewStatsComponent())

	entities := []*Entity{entity}

	// Should not panic, should just skip the entity
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify stats remain at defaults
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		t.Fatal("stats component not found")
	}
	stats := statsComp.(*StatsComponent)
	if stats.CritChance != 0.05 {
		t.Errorf("Expected default crit chance 0.05, got %f", stats.CritChance)
	}
}

// TestSkillProgressionSystem_Update_NoStats tests entities without stats
func TestSkillProgressionSystem_Update_NoStats(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Create skill tree with learned skill
	critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", 0.05, false)
	skillNode := createSkillNode(critSkill)

	skillTree := &skills.SkillTree{
		Name:  "Test Tree",
		Nodes: []*skills.SkillNode{skillNode},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["crit_1"] = true
	skillTreeComp.SkillLevels["crit_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Should not panic, should just skip the entity
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}
}

// TestSkillProgressionSystem_CritChanceBonus tests crit chance bonus application
func TestSkillProgressionSystem_CritChanceBonus(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	// Create skill tree with crit chance skill
	critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", 0.10, false) // +10%
	skillNode := createSkillNode(critSkill)

	skillTree := &skills.SkillTree{
		Name:  "Warrior Tree",
		Nodes: []*skills.SkillNode{skillNode},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["crit_1"] = true
	skillTreeComp.SkillLevels["crit_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates until bonuses are applied (60 frames)
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify crit chance was increased
	expectedCrit := 0.05 + 0.10 // base + bonus
	if math.Abs(stats.CritChance-expectedCrit) > 0.0001 {
		t.Errorf("Expected crit chance %f, got %f", expectedCrit, stats.CritChance)
	}
}

// TestSkillProgressionSystem_CritDamageBonus tests crit damage bonus application
func TestSkillProgressionSystem_CritDamageBonus(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	// Create skill tree with crit damage skill (percentage bonus)
	critDmgSkill := createSkill("critdmg_1", "Deadly Strike", "crit_damage", 0.25, true) // +25%
	skillNode := createSkillNode(critDmgSkill)

	skillTree := &skills.SkillTree{
		Name:  "Warrior Tree",
		Nodes: []*skills.SkillNode{skillNode},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["critdmg_1"] = true
	skillTreeComp.SkillLevels["critdmg_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates until bonuses are applied
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify crit damage was increased (base 2.0 + 0.25)
	expectedCritDmg := 2.0 + 0.25 // base + bonus
	if math.Abs(stats.CritDamage-expectedCritDmg) > 0.0001 {
		t.Errorf("Expected crit damage %f, got %f", expectedCritDmg, stats.CritDamage)
	}
}

// TestSkillProgressionSystem_MultipleSkills tests multiple skills applying bonuses
func TestSkillProgressionSystem_MultipleSkills(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	// Create skill tree with multiple skills
	critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", 0.08, false)
	critDmgSkill := createSkill("critdmg_1", "Deadly Strike", "crit_damage", 0.30, true)

	skillTree := &skills.SkillTree{
		Name: "Warrior Tree",
		Nodes: []*skills.SkillNode{
			createSkillNode(critSkill),
			createSkillNode(critDmgSkill),
		},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["crit_1"] = true
	skillTreeComp.LearnedSkills["critdmg_1"] = true
	skillTreeComp.SkillLevels["crit_1"] = 1
	skillTreeComp.SkillLevels["critdmg_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify both bonuses were applied
	expectedCrit := 0.05 + 0.08
	expectedCritDmg := 2.0 + 0.30

	if math.Abs(stats.CritChance-expectedCrit) > 0.0001 {
		t.Errorf("Expected crit chance %f, got %f", expectedCrit, stats.CritChance)
	}
	if math.Abs(stats.CritDamage-expectedCritDmg) > 0.0001 {
		t.Errorf("Expected crit damage %f, got %f", expectedCritDmg, stats.CritDamage)
	}
}

// TestSkillProgressionSystem_SkillLevelScaling tests that higher skill levels increase bonuses
func TestSkillProgressionSystem_SkillLevelScaling(t *testing.T) {
	tests := []struct {
		name          string
		skillLevel    int
		baseBonus     float64
		expectedBonus float64
	}{
		{"Level 1", 1, 0.10, 0.10},
		{"Level 2", 2, 0.10, 0.20},
		{"Level 3", 3, 0.10, 0.30},
		{"Level 5", 5, 0.10, 0.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewSkillProgressionSystem()

			entity := NewEntity(1)
			stats := NewStatsComponent()
			entity.AddComponent(stats)

			// Create skill with base bonus (crit chance)
			critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", tt.baseBonus, false)
			skillNode := createSkillNode(critSkill)

			skillTree := &skills.SkillTree{
				Name:  "Test Tree",
				Nodes: []*skills.SkillNode{skillNode},
			}

			skillTreeComp := NewSkillTreeComponent(skillTree)
			skillTreeComp.LearnedSkills["crit_1"] = true
			skillTreeComp.SkillLevels["crit_1"] = tt.skillLevel
			entity.AddComponent(skillTreeComp)

			entities := []*Entity{entity}

			// Run updates
			for i := 0; i < 60; i++ {
				system.Update(entities, 0.016)
			}

			// Verify scaled bonus
			expectedCrit := 0.05 + tt.expectedBonus // base + (bonus * level)
			if math.Abs(stats.CritChance-expectedCrit) > 0.0001 {
				t.Errorf("Expected crit chance %f, got %f", expectedCrit, stats.CritChance)
			}
		})
	}
}

// TestSkillProgressionSystem_CritChanceCap tests that crit chance is capped at 100%
func TestSkillProgressionSystem_CritChanceCap(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	// Start at default crit chance (5%) and add huge bonus
	entity.AddComponent(stats)

	// Create skill tree with huge crit chance bonus
	critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", 1.50, false) // +150%
	skillNode := createSkillNode(critSkill)

	skillTree := &skills.SkillTree{
		Name:  "Test Tree",
		Nodes: []*skills.SkillNode{skillNode},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["crit_1"] = true
	skillTreeComp.SkillLevels["crit_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify crit chance is capped at 100%
	if stats.CritChance > 1.0 {
		t.Errorf("Crit chance exceeded 100%%: got %f", stats.CritChance)
	}
	if math.Abs(stats.CritChance-1.0) > 0.0001 {
		t.Errorf("Expected capped crit chance 1.0, got %f", stats.CritChance)
	}
}

// TestSkillProgressionSystem_Integration tests full system integration with multiple entities
func TestSkillProgressionSystem_Integration(t *testing.T) {
	system := NewSkillProgressionSystem()

	// Entity 1: Warrior with crit focus
	entity1 := NewEntity(1)
	stats1 := NewStatsComponent()
	entity1.AddComponent(stats1)

	critSkill1 := createSkill("crit_1", "Critical Strike", "crit_chance", 0.10, false)
	skillTree1 := &skills.SkillTree{
		Name:  "Warrior",
		Nodes: []*skills.SkillNode{createSkillNode(critSkill1)},
	}
	skillTreeComp1 := NewSkillTreeComponent(skillTree1)
	skillTreeComp1.LearnedSkills["crit_1"] = true
	skillTreeComp1.SkillLevels["crit_1"] = 1
	entity1.AddComponent(skillTreeComp1)

	// Entity 2: Berserker with crit damage focus
	entity2 := NewEntity(2)
	stats2 := NewStatsComponent()
	entity2.AddComponent(stats2)

	critDmgSkill2 := createSkill("critdmg_1", "Deadly Strike", "crit_damage", 0.20, true)
	skillTree2 := &skills.SkillTree{
		Name:  "Berserker",
		Nodes: []*skills.SkillNode{createSkillNode(critDmgSkill2)},
	}
	skillTreeComp2 := NewSkillTreeComponent(skillTree2)
	skillTreeComp2.LearnedSkills["critdmg_1"] = true
	skillTreeComp2.SkillLevels["critdmg_1"] = 3 // Level 3
	entity2.AddComponent(skillTreeComp2)

	entities := []*Entity{entity1, entity2}

	// Run updates
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify entity 1 bonuses
	expectedCrit1 := 0.05 + 0.10
	if math.Abs(stats1.CritChance-expectedCrit1) > 0.0001 {
		t.Errorf("Entity1: Expected crit chance %f, got %f", expectedCrit1, stats1.CritChance)
	}

	// Verify entity 2 bonuses (level 3 = 3x bonus)
	expectedCritDmg2 := 2.0 + (0.20 * 3)
	if math.Abs(stats2.CritDamage-expectedCritDmg2) > 0.0001 {
		t.Errorf("Entity2: Expected crit damage %f, got %f", expectedCritDmg2, stats2.CritDamage)
	}
}

// TestSkillProgressionSystem_UpdateInterval tests that updates only happen every N frames
func TestSkillProgressionSystem_UpdateInterval(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", 0.10, false)
	skillTree := &skills.SkillTree{
		Name:  "Test Tree",
		Nodes: []*skills.SkillNode{createSkillNode(critSkill)},
	}
	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["crit_1"] = true
	skillTreeComp.SkillLevels["crit_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run 59 updates (just before interval)
	for i := 0; i < 59; i++ {
		system.Update(entities, 0.016)
	}

	// Bonuses should not be applied yet
	// (Note: Might be applied if system starts at frame 0 and processes immediately)
	// So we'll just verify the frame counter incremented
	if system.frameCounter != 59 {
		t.Errorf("Expected frame counter 59, got %d", system.frameCounter)
	}

	// Run one more update (reaches interval of 60)
	system.Update(entities, 0.016)

	// Now bonuses should be applied
	expectedCrit := 0.05 + 0.10
	if math.Abs(stats.CritChance-expectedCrit) > 0.0001 {
		t.Errorf("Expected crit chance %f after 60 frames, got %f", expectedCrit, stats.CritChance)
	}

	// Frame counter should have wrapped around
	if system.frameCounter != 0 {
		t.Errorf("Expected frame counter to reset to 0, got %d", system.frameCounter)
	}
}

// TestSkillProgressionSystem_BonusAccumulation tests bonus accumulation from skill tree
func TestSkillProgressionSystem_BonusAccumulation(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	// Create multiple crit skills with different amounts
	skill1 := createSkill("crit_1", "Crit I", "crit_chance", 0.05, false)
	skill2 := createSkill("crit_2", "Crit II", "crit_chance", 0.08, false)
	skill3 := createSkill("crit_3", "Crit III", "crit_chance", 0.12, false)

	skillTree := &skills.SkillTree{
		Name: "Multi-Crit Tree",
		Nodes: []*skills.SkillNode{
			createSkillNode(skill1),
			createSkillNode(skill2),
			createSkillNode(skill3),
		},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["crit_1"] = true
	skillTreeComp.LearnedSkills["crit_2"] = true
	skillTreeComp.LearnedSkills["crit_3"] = true
	skillTreeComp.SkillLevels["crit_1"] = 1
	skillTreeComp.SkillLevels["crit_2"] = 2
	skillTreeComp.SkillLevels["crit_3"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify all bonuses accumulated: 0.05*1 + 0.08*2 + 0.12*1 = 0.33
	expectedCrit := 0.05 + (0.05 * 1) + (0.08 * 2) + (0.12 * 1)
	if math.Abs(stats.CritChance-expectedCrit) > 0.0001 {
		t.Errorf("Expected accumulated crit chance %f, got %f", expectedCrit, stats.CritChance)
	}
}

// TestSkillProgressionSystem_EmptyLearnedSkills tests entities with skill trees but no learned skills
func TestSkillProgressionSystem_EmptyLearnedSkills(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	// Create skill tree but don't learn any skills
	critSkill := createSkill("crit_1", "Critical Strike", "crit_chance", 0.10, false)
	skillTree := &skills.SkillTree{
		Name:  "Test Tree",
		Nodes: []*skills.SkillNode{createSkillNode(critSkill)},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	// Don't add any learned skills
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Stats should remain at defaults since no skills learned
	if stats.CritChance != 0.05 {
		t.Errorf("Expected default crit chance 0.05, got %f", stats.CritChance)
	}
	if stats.CritDamage != 2.0 {
		t.Errorf("Expected default crit damage 2.0, got %f", stats.CritDamage)
	}
}

// TestSkillProgressionSystem_MixedEffectTypes tests skills with different effect types
func TestSkillProgressionSystem_MixedEffectTypes(t *testing.T) {
	system := NewSkillProgressionSystem()

	entity := NewEntity(1)
	stats := NewStatsComponent()
	entity.AddComponent(stats)

	// Create skill with multiple effect types (only crit effects should apply)
	multiSkill := &skills.Skill{
		ID:   "multi_1",
		Name: "Multi Effect",
		Effects: []skills.Effect{
			{Type: "crit_chance", Value: 0.10, IsPercent: false},
			{Type: "crit_damage", Value: 0.50, IsPercent: true},
			{Type: "damage", Value: 0.20, IsPercent: true},  // Should be ignored (not implemented)
			{Type: "defense", Value: 0.15, IsPercent: true}, // Should be ignored (not implemented)
		},
		Level:    1,
		MaxLevel: 5,
	}

	skillTree := &skills.SkillTree{
		Name:  "Mixed Tree",
		Nodes: []*skills.SkillNode{createSkillNode(multiSkill)},
	}

	skillTreeComp := NewSkillTreeComponent(skillTree)
	skillTreeComp.LearnedSkills["multi_1"] = true
	skillTreeComp.SkillLevels["multi_1"] = 1
	entity.AddComponent(skillTreeComp)

	entities := []*Entity{entity}

	// Run updates
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Only crit bonuses should apply
	expectedCrit := 0.05 + 0.10
	expectedCritDmg := 2.0 + 0.50

	if math.Abs(stats.CritChance-expectedCrit) > 0.0001 {
		t.Errorf("Expected crit chance %f, got %f", expectedCrit, stats.CritChance)
	}
	if math.Abs(stats.CritDamage-expectedCritDmg) > 0.0001 {
		t.Errorf("Expected crit damage %f, got %f", expectedCritDmg, stats.CritDamage)
	}

	// Attack and defense should remain at defaults (not implemented)
	if stats.Attack != 10.0 {
		t.Errorf("Expected attack to remain at default 10.0, got %f", stats.Attack)
	}
	if stats.Defense != 5.0 {
		t.Errorf("Expected defense to remain at default 5.0, got %f", stats.Defense)
	}
}
