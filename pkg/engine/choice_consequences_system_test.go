package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/integration/choice_consequences"
)

func TestNewChoiceConsequencesSystem(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)

	if system == nil {
		t.Fatal("Expected non-nil system")
	}
	if system.world != world {
		t.Error("Expected system to reference world")
	}
	if system.tracker == nil {
		t.Error("Expected non-nil tracker")
	}
}

func TestChoiceConsequencesSystem_RecordChoice(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_123"

	choice := &choice_consequences.PlayerChoice{
		ChoiceID:    "quest_village_burned_spare_bandit",
		StoryNodeID: "village_burned_confrontation",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &choice_consequences.AlignmentShift{
			GoodEvil: 0.2,
			LawChaos: -0.1,
		},
		Irreversible: true,
		Consequences: []string{"lock_quest_bandit_execution"},
	}

	err := system.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("Failed to record choice: %v", err)
	}

	// Verify alignment was updated
	alignment := system.GetAlignment(playerID)
	if alignment.GoodEvil != 0.2 {
		t.Errorf("Expected GoodEvil=0.2, got %f", alignment.GoodEvil)
	}
	if alignment.LawChaos != -0.1 {
		t.Errorf("Expected LawChaos=-0.1, got %f", alignment.LawChaos)
	}
}

func TestChoiceConsequencesSystem_IsContentAvailable(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_456"

	// All content available initially
	if !system.IsContentAvailable(playerID, "quest_bandit_redemption") {
		t.Error("Expected content to be available initially")
	}

	// Make choice that locks content
	choice := &choice_consequences.PlayerChoice{
		ChoiceID:     "quest_village_burned_execute_bandit",
		StoryNodeID:  "village_burned_confrontation",
		Timestamp:    time.Now().Unix(),
		Irreversible: true,
		Consequences: []string{"lock_quest_bandit_redemption"},
	}

	err := system.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("Failed to record choice: %v", err)
	}

	// Content should now be locked
	if system.IsContentAvailable(playerID, "bandit_redemption") {
		t.Error("Expected content to be locked after choice")
	}
}

func TestChoiceConsequencesSystem_NPCAttitude(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_789"
	npcID := "villager_elder"

	// Neutral attitude initially
	attitude := system.GetNPCAttitude(playerID, npcID)
	if attitude != 0.0 {
		t.Errorf("Expected neutral attitude (0.0), got %f", attitude)
	}

	// Make choice affecting NPC
	choice := &choice_consequences.PlayerChoice{
		ChoiceID:    "quest_village_help",
		StoryNodeID: "village_quest",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &choice_consequences.AlignmentShift{
			GoodEvil: 0.5,
		},
		NPCsAffected: []string{npcID},
	}

	err := system.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("Failed to record choice: %v", err)
	}

	// Attitude should improve
	attitude = system.GetNPCAttitude(playerID, npcID)
	if attitude <= 0.0 {
		t.Errorf("Expected positive attitude after good deed, got %f", attitude)
	}
}

func TestChoiceConsequencesSystem_QuestBranching(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_branch"

	// Register quest branch with prerequisites
	branch := &choice_consequences.QuestBranch{
		QuestID:       "main_quest_chapter2",
		BranchID:      "good_path",
		Prerequisites: []string{"choice_spare_enemy"},
	}

	err := system.RegisterQuestBranch(branch)
	if err != nil {
		t.Fatalf("Failed to register quest branch: %v", err)
	}

	// Branch not available without prerequisite
	if system.IsQuestBranchAvailable(playerID, "main_quest_chapter2", "good_path") {
		t.Error("Expected branch to be unavailable without prerequisite")
	}

	// Make prerequisite choice
	choice := &choice_consequences.PlayerChoice{
		ChoiceID:    "choice_spare_enemy",
		StoryNodeID: "combat_aftermath",
		Timestamp:   time.Now().Unix(),
	}

	err = system.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("Failed to record choice: %v", err)
	}

	// Branch should now be available
	if !system.IsQuestBranchAvailable(playerID, "main_quest_chapter2", "good_path") {
		t.Error("Expected branch to be available after prerequisite")
	}
}

func TestChoiceConsequencesSystem_ClassQuests(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_class"

	// Register class-specific quest
	quest := &choice_consequences.ClassSpecificQuest{
		QuestID:       "warrior_epic_quest",
		RequiredClass: "warrior",
		MinLevel:      10,
		AlignmentReq: &choice_consequences.AlignmentRequirement{
			MinGoodEvil: -0.5,
			MaxGoodEvil: 0.5,
		},
	}

	err := system.RegisterClassQuest(quest)
	if err != nil {
		t.Fatalf("Failed to register class quest: %v", err)
	}

	tests := []struct {
		name        string
		playerClass string
		playerLevel int
		wantAvail   bool
	}{
		{"wrong class", "mage", 10, false},
		{"too low level", "warrior", 5, false},
		{"perfect match", "warrior", 10, true},
		{"higher level", "warrior", 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avail := system.IsClassQuestAvailable(playerID, "warrior_epic_quest", tt.playerClass, tt.playerLevel)
			if avail != tt.wantAvail {
				t.Errorf("IsClassQuestAvailable() = %v, want %v", avail, tt.wantAvail)
			}
		})
	}
}

func TestChoiceConsequencesSystem_CompanionReactions(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_companion"
	companionID := "companion_loyal_knight"

	reaction := &choice_consequences.CompanionReaction{
		CompanionID:  companionID,
		ChoiceID:     "choice_help_innocent",
		LoyaltyDelta: 0.1,
		Approval:     true,
		Comment:      "I admire your compassion, friend.",
	}

	err := system.RecordCompanionReaction(playerID, reaction)
	if err != nil {
		t.Fatalf("Failed to record companion reaction: %v", err)
	}

	reactions := system.GetCompanionReactions(playerID, companionID)
	if len(reactions) != 1 {
		t.Errorf("Expected 1 reaction, got %d", len(reactions))
	}
	if reactions[0].Approval != true {
		t.Error("Expected approval to be true")
	}
}

func TestChoiceConsequencesSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)

	// Create player entity with choice tracker component
	player := world.CreateEntity()
	playerID := "player_test_update"

	choiceComp := &choice_consequences.ChoiceTrackerComponent{
		PlayerID:         playerID,
		ChoiceHistory:    make([]*choice_consequences.PlayerChoice, 0),
		NPCRelationships: make(map[string]*choice_consequences.NPCRelationship),
		ContentLocks:     make(map[string]*choice_consequences.ContentLock),
		Alignment: &choice_consequences.PlayerAlignment{
			GoodEvil:      0.0,
			LawChaos:      0.0,
			HonorDishonor: 0.0,
			UpdatedAt:     time.Now().Unix(),
		},
		CompanionReactions: make([]*choice_consequences.CompanionReaction, 0),
	}
	player.AddComponent(choiceComp)

	// Make a choice to change alignment
	choice := &choice_consequences.PlayerChoice{
		ChoiceID:    "test_choice",
		StoryNodeID: "test_node",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &choice_consequences.AlignmentShift{
			GoodEvil: 0.3,
		},
	}
	system.RecordChoice(playerID, choice)

	// Update system
	entities := []*Entity{player}
	system.Update(entities, 0.016)

	// Component alignment should be synced
	comp, _ := player.GetComponent("choice_tracker")
	choiceComp = comp.(*choice_consequences.ChoiceTrackerComponent)
	if choiceComp.Alignment.GoodEvil != 0.3 {
		t.Errorf("Expected alignment sync, got GoodEvil=%f", choiceComp.Alignment.GoodEvil)
	}
}

func TestChoiceConsequencesSystem_InvalidChoice(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_invalid"

	tests := []struct {
		name    string
		choice  *choice_consequences.PlayerChoice
		wantErr bool
	}{
		{
			name:    "nil choice",
			choice:  nil,
			wantErr: true,
		},
		{
			name: "empty choice ID",
			choice: &choice_consequences.PlayerChoice{
				ChoiceID:    "",
				StoryNodeID: "node",
				Timestamp:   time.Now().Unix(),
			},
			wantErr: true,
		},
		{
			name: "valid choice",
			choice: &choice_consequences.PlayerChoice{
				ChoiceID:    "valid_choice",
				StoryNodeID: "valid_node",
				Timestamp:   time.Now().Unix(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := system.RecordChoice(playerID, tt.choice)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordChoice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChoiceConsequencesSystem_SaveLoad(t *testing.T) {
	world := NewWorld()
	system := NewChoiceConsequencesSystem(world)
	playerID := "player_test_saveload"

	// Record some choices
	choice := &choice_consequences.PlayerChoice{
		ChoiceID:    "save_test_choice",
		StoryNodeID: "save_test_node",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &choice_consequences.AlignmentShift{
			GoodEvil: 0.5,
		},
	}
	err := system.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("Failed to record choice: %v", err)
	}

	// Save to file
	filename := "/tmp/choice_test_save.gz"
	defer func() {
		// Clean up
		_ = system.tracker.Load("/dev/null") // Reset tracker
	}()

	err = system.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Create new system and load
	system2 := NewChoiceConsequencesSystem(world)
	err = system2.Load(filename)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify data was loaded
	alignment := system2.GetAlignment(playerID)
	if alignment.GoodEvil != 0.5 {
		t.Errorf("Expected loaded alignment GoodEvil=0.5, got %f", alignment.GoodEvil)
	}
}
