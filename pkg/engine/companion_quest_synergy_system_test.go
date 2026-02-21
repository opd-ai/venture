package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

func TestCompanionSkillType_String(t *testing.T) {
	tests := []struct {
		skill    CompanionSkillType
		expected string
	}{
		{SkillNone, "None"},
		{SkillTracker, "Tracker"},
		{SkillHunter, "Hunter"},
		{SkillGatherer, "Gatherer"},
		{SkillGuardian, "Guardian"},
		{SkillDiplomat, "Diplomat"},
		{SkillCombatant, "Combatant"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.skill.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCompanionSkillType_MatchesQuestType(t *testing.T) {
	tests := []struct {
		name      string
		skill     CompanionSkillType
		questType quest.QuestType
		want      bool
	}{
		{"Tracker matches Explore", SkillTracker, quest.TypeExplore, true},
		{"Tracker no match Kill", SkillTracker, quest.TypeKill, false},
		{"Hunter matches Kill", SkillHunter, quest.TypeKill, true},
		{"Hunter no match Collect", SkillHunter, quest.TypeCollect, false},
		{"Gatherer matches Collect", SkillGatherer, quest.TypeCollect, true},
		{"Guardian matches Escort", SkillGuardian, quest.TypeEscort, true},
		{"Diplomat matches Talk", SkillDiplomat, quest.TypeTalk, true},
		{"Diplomat matches FactionConflict", SkillDiplomat, quest.TypeFactionConflict, true},
		{"Combatant matches Boss", SkillCombatant, quest.TypeBoss, true},
		{"None matches nothing", SkillNone, quest.TypeKill, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.skill.MatchesQuestType(tt.questType); got != tt.want {
				t.Errorf("MatchesQuestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCompanionSkillForType(t *testing.T) {
	tests := []struct {
		companionType CompanionType
		expectedSkill CompanionSkillType
	}{
		{CompanionTypePet, SkillTracker},
		{CompanionTypeSummon, SkillCombatant},
		{CompanionTypeHireling, SkillGatherer},
		{CompanionTypeElemental, SkillCombatant},
		{CompanionTypeUndead, SkillHunter},
		{CompanionTypeRobot, SkillTracker},
		{CompanionTypeSpirit, SkillDiplomat},
		{CompanionTypeInsect, SkillGatherer},
	}

	for _, tt := range tests {
		t.Run(tt.expectedSkill.String(), func(t *testing.T) {
			if got := GetCompanionSkillForType(tt.companionType); got != tt.expectedSkill {
				t.Errorf("GetCompanionSkillForType(%v) = %v, want %v", tt.companionType, got, tt.expectedSkill)
			}
		})
	}
}

func TestCompanionQuestSynergyComponent_Type(t *testing.T) {
	comp := NewCompanionQuestSynergyComponent()
	if got := comp.Type(); got != "companion_quest_synergy" {
		t.Errorf("Type() = %v, want companion_quest_synergy", got)
	}
}

func TestCompanionQuestSynergyComponent_AddSynergy(t *testing.T) {
	comp := NewCompanionQuestSynergyComponent()

	comp.AddSynergy("quest_1", 100, SkillTracker, 50.0)

	if !comp.HasActiveSynergy("quest_1") {
		t.Error("Expected active synergy for quest_1")
	}

	// Verify base bonuses are applied
	objBonus := comp.GetObjectiveBonus("quest_1")
	if objBonus <= 1.0 {
		t.Errorf("Expected objective bonus > 1.0, got %v", objBonus)
	}

	rewardBonus := comp.GetRewardBonus("quest_1")
	if rewardBonus <= 1.0 {
		t.Errorf("Expected reward bonus > 1.0, got %v", rewardBonus)
	}
}

func TestCompanionQuestSynergyComponent_ApplySkillMatch(t *testing.T) {
	comp := NewCompanionQuestSynergyComponent()
	comp.AddSynergy("quest_1", 100, SkillTracker, 50.0)

	beforeObj := comp.GetObjectiveBonus("quest_1")
	beforeReward := comp.GetRewardBonus("quest_1")

	comp.ApplySkillMatch("quest_1", true)

	afterObj := comp.GetObjectiveBonus("quest_1")
	afterReward := comp.GetRewardBonus("quest_1")

	if afterObj <= beforeObj {
		t.Errorf("Expected objective bonus to increase with skill match, before=%v after=%v", beforeObj, afterObj)
	}
	if afterReward <= beforeReward {
		t.Errorf("Expected reward bonus to increase with skill match, before=%v after=%v", beforeReward, afterReward)
	}
}

func TestCompanionQuestSynergyComponent_NoActiveSynergy(t *testing.T) {
	comp := NewCompanionQuestSynergyComponent()

	objBonus := comp.GetObjectiveBonus("nonexistent_quest")
	if objBonus != 1.0 {
		t.Errorf("Expected 1.0 for nonexistent quest, got %v", objBonus)
	}

	rewardBonus := comp.GetRewardBonus("nonexistent_quest")
	if rewardBonus != 1.0 {
		t.Errorf("Expected 1.0 for nonexistent quest, got %v", rewardBonus)
	}
}

func TestCompanionQuestSynergyComponent_CompleteSynergy(t *testing.T) {
	comp := NewCompanionQuestSynergyComponent()
	comp.AddSynergy("quest_1", 100, SkillTracker, 75.0)
	comp.ApplySkillMatch("quest_1", true)

	comp.CompleteSynergy("quest_1", 50, 25)

	if comp.HasActiveSynergy("quest_1") {
		t.Error("Expected synergy to be inactive after completion")
	}
	if comp.TotalBonusXP != 50 {
		t.Errorf("Expected TotalBonusXP = 50, got %v", comp.TotalBonusXP)
	}
	if comp.TotalBonusGold != 25 {
		t.Errorf("Expected TotalBonusGold = 25, got %v", comp.TotalBonusGold)
	}
	if comp.QuestsCompletedWithSynergy != 1 {
		t.Errorf("Expected QuestsCompletedWithSynergy = 1, got %v", comp.QuestsCompletedWithSynergy)
	}
	if len(comp.CompletedSynergies) != 1 {
		t.Errorf("Expected 1 completed synergy, got %v", len(comp.CompletedSynergies))
	}
}

func TestCompanionQuestSynergyComponent_RemoveSynergy(t *testing.T) {
	comp := NewCompanionQuestSynergyComponent()
	comp.AddSynergy("quest_1", 100, SkillTracker, 50.0)

	comp.RemoveSynergy("quest_1")

	if comp.HasActiveSynergy("quest_1") {
		t.Error("Expected synergy to be removed")
	}
}

func TestCompanionQuestSynergyComponent_LoyaltyBonus(t *testing.T) {
	// Test that higher loyalty gives higher bonuses
	lowLoyaltyComp := NewCompanionQuestSynergyComponent()
	lowLoyaltyComp.AddSynergy("quest_1", 100, SkillTracker, 10.0) // Low loyalty

	highLoyaltyComp := NewCompanionQuestSynergyComponent()
	highLoyaltyComp.AddSynergy("quest_1", 100, SkillTracker, 90.0) // High loyalty

	lowObjBonus := lowLoyaltyComp.GetObjectiveBonus("quest_1")
	highObjBonus := highLoyaltyComp.GetObjectiveBonus("quest_1")

	if highObjBonus <= lowObjBonus {
		t.Errorf("Expected high loyalty bonus (%v) > low loyalty bonus (%v)", highObjBonus, lowObjBonus)
	}
}

func TestNewCompanionQuestSynergySystem(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	if system == nil {
		t.Fatal("Expected non-nil system")
	}
	if system.world != world {
		t.Error("Expected system to reference world")
	}
	if system.synergyDistance <= 0 {
		t.Error("Expected positive synergy distance")
	}
}

func TestCompanionQuestSynergySystem_Update_NoEntities(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Should not panic with empty entities
	system.Update([]*Entity{}, 1.0)
}

func TestCompanionQuestSynergySystem_FindActiveCompanions(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion near owner
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 150, Y: 150}) // Within 300px
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       75.0,
	})

	companions := system.findActiveCompanions(owner.ID)
	if len(companions) != 1 {
		t.Errorf("Expected 1 companion, got %v", len(companions))
	}
}

func TestCompanionQuestSynergySystem_FindActiveCompanions_TooFar(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion far from owner
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 1000, Y: 1000}) // Beyond 300px
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       75.0,
	})

	companions := system.findActiveCompanions(owner.ID)
	if len(companions) != 0 {
		t.Errorf("Expected 0 companions (too far), got %v", len(companions))
	}
}

func TestCompanionQuestSynergySystem_ApplyObjectiveProgress(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Create player with synergy component
	player := world.CreateEntity()
	synergy := NewCompanionQuestSynergyComponent()
	synergy.AddSynergy("quest_1", 100, SkillTracker, 80.0)
	synergy.ApplySkillMatch("quest_1", true) // Adds bonus
	player.AddComponent(synergy)

	// Base progress should be boosted
	baseProgress := 10
	boostedProgress := system.ApplyObjectiveProgress(player.ID, "quest_1", baseProgress)

	if boostedProgress <= baseProgress {
		t.Errorf("Expected boosted progress > %v, got %v", baseProgress, boostedProgress)
	}
}

func TestCompanionQuestSynergySystem_ApplyObjectiveProgress_NoSynergy(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Player without synergy component
	player := world.CreateEntity()

	baseProgress := 10
	result := system.ApplyObjectiveProgress(player.ID, "quest_1", baseProgress)

	if result != baseProgress {
		t.Errorf("Expected unchanged progress %v, got %v", baseProgress, result)
	}
}

func TestCompanionQuestSynergySystem_ApplyRewardBonus(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Create player with synergy
	player := world.CreateEntity()
	synergy := NewCompanionQuestSynergyComponent()
	synergy.AddSynergy("quest_1", 100, SkillCombatant, 90.0)
	synergy.ApplySkillMatch("quest_1", true)
	player.AddComponent(synergy)

	baseXP := 100
	baseGold := 50
	boostedXP, boostedGold := system.ApplyRewardBonus(player.ID, "quest_1", baseXP, baseGold)

	if boostedXP <= baseXP {
		t.Errorf("Expected boosted XP > %v, got %v", baseXP, boostedXP)
	}
	if boostedGold <= baseGold {
		t.Errorf("Expected boosted gold > %v, got %v", baseGold, boostedGold)
	}

	// Verify synergy was completed
	if synergy.HasActiveSynergy("quest_1") {
		t.Error("Expected synergy to be completed after reward bonus")
	}
}

func TestCompanionQuestSynergySystem_OnQuestAccepted(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Create player with position
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 150, Y: 150})
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypePet, // Tracker skill
		Loyalty:       60.0,
	})

	// Accept an explore quest (matches Tracker skill)
	q := &quest.Quest{
		ID:   "explore_quest_1",
		Name: "Explore the Ruins",
		Type: quest.TypeExplore,
	}
	system.OnQuestAccepted(player.ID, q)

	// Verify synergy was created
	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	if !hasSynergy {
		t.Fatal("Expected synergy component to be created")
	}
	synergy := synergyComp.(*CompanionQuestSynergyComponent)
	if !synergy.HasActiveSynergy("explore_quest_1") {
		t.Error("Expected active synergy for quest")
	}

	// Verify skill match bonus was applied (Tracker + Explore = match)
	objBonus := synergy.GetObjectiveBonus("explore_quest_1")
	// Base (1.0) + companion (0.05) + loyalty (0.09) + skill match (0.25) = ~1.39
	if objBonus < 1.3 {
		t.Errorf("Expected significant bonus from skill match, got %v", objBonus)
	}
}

func TestCompanionQuestSynergySystem_OnQuestAbandoned(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Create player with synergy
	player := world.CreateEntity()
	synergy := NewCompanionQuestSynergyComponent()
	synergy.AddSynergy("quest_1", 100, SkillTracker, 50.0)
	player.AddComponent(synergy)

	system.OnQuestAbandoned(player.ID, "quest_1")

	if synergy.HasActiveSynergy("quest_1") {
		t.Error("Expected synergy to be removed after abandon")
	}
}

func TestCompanionQuestSynergySystem_GetSynergyStats(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	player := world.CreateEntity()
	synergy := NewCompanionQuestSynergyComponent()
	synergy.AddSynergy("quest_1", 100, SkillTracker, 50.0)
	synergy.CompleteSynergy("quest_1", 100, 50)
	synergy.AddSynergy("quest_2", 100, SkillHunter, 75.0)
	synergy.CompleteSynergy("quest_2", 150, 75)
	player.AddComponent(synergy)

	bonusXP, bonusGold, questCount := system.GetSynergyStats(player.ID)

	if bonusXP != 250 {
		t.Errorf("Expected total bonus XP = 250, got %v", bonusXP)
	}
	if bonusGold != 125 {
		t.Errorf("Expected total bonus gold = 125, got %v", bonusGold)
	}
	if questCount != 2 {
		t.Errorf("Expected quest count = 2, got %v", questCount)
	}
}

func TestCompanionQuestSynergySystem_GetActiveSynergyBonus(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	player := world.CreateEntity()
	synergy := NewCompanionQuestSynergyComponent()
	synergy.AddSynergy("quest_1", 100, SkillTracker, 100.0) // Max loyalty
	synergy.ApplySkillMatch("quest_1", true)
	player.AddComponent(synergy)

	objBonus, rewardBonus := system.GetActiveSynergyBonus(player.ID, "quest_1")

	if objBonus <= 1.0 {
		t.Errorf("Expected objective bonus > 1.0, got %v", objBonus)
	}
	if rewardBonus <= 1.0 {
		t.Errorf("Expected reward bonus > 1.0, got %v", rewardBonus)
	}
}

func TestCompanionQuestSynergySystem_GetActiveSynergyBonus_NoPlayer(t *testing.T) {
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	objBonus, rewardBonus := system.GetActiveSynergyBonus(999, "quest_1")

	if objBonus != 1.0 || rewardBonus != 1.0 {
		t.Errorf("Expected default bonuses (1.0, 1.0), got (%v, %v)", objBonus, rewardBonus)
	}
}

func TestCompanionQuestSynergySystem_IntegrationScenario(t *testing.T) {
	// Full integration test: player with companion accepts quest, progresses, completes
	world := NewWorld()
	system := NewCompanionQuestSynergySystem(world)

	// Setup player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	questTracker := NewQuestTrackerComponent(5)
	player.AddComponent(questTracker)

	// Setup companion (Hunter type for kill quests)
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 120, Y: 120})
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypeUndead, // Hunter skill
		Loyalty:       85.0,
	})

	// Accept a kill quest
	killQuest := &quest.Quest{
		ID:   "kill_goblins",
		Name: "Hunt the Goblins",
		Type: quest.TypeKill,
		Objectives: []quest.Objective{
			{Description: "Kill goblins", Target: "Goblin", Required: 10, Current: 0},
		},
		Reward: quest.Reward{XP: 200, Gold: 100},
	}

	// Step 1: Accept quest and trigger synergy
	system.OnQuestAccepted(player.ID, killQuest)

	// Verify synergy established
	synergyComp, _ := player.GetComponent("companion_quest_synergy")
	synergy := synergyComp.(*CompanionQuestSynergyComponent)
	if !synergy.HasActiveSynergy("kill_goblins") {
		t.Fatal("Expected synergy to be established")
	}

	// Step 2: Apply objective progress boost
	baseProgress := 5
	boostedProgress := system.ApplyObjectiveProgress(player.ID, "kill_goblins", baseProgress)
	if boostedProgress <= baseProgress {
		t.Errorf("Expected boosted progress, base=%v boosted=%v", baseProgress, boostedProgress)
	}

	// Step 3: Apply reward bonus on completion
	boostedXP, boostedGold := system.ApplyRewardBonus(player.ID, "kill_goblins", 200, 100)
	if boostedXP <= 200 {
		t.Errorf("Expected boosted XP > 200, got %v", boostedXP)
	}
	if boostedGold <= 100 {
		t.Errorf("Expected boosted gold > 100, got %v", boostedGold)
	}

	// Verify stats were recorded
	totalXP, totalGold, questCount := system.GetSynergyStats(player.ID)
	if questCount != 1 {
		t.Errorf("Expected 1 completed quest, got %v", questCount)
	}
	if totalXP <= 0 {
		t.Errorf("Expected positive bonus XP, got %v", totalXP)
	}
	if totalGold <= 0 {
		t.Errorf("Expected positive bonus gold, got %v", totalGold)
	}
}
