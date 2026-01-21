package choice_consequences

import (
	"os"
	"testing"
	"time"
)

func TestPlayerChoiceBasics(t *testing.T) {
	choice := &PlayerChoice{
		ChoiceID:    "test_choice_1",
		StoryNodeID: "node_1",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &AlignmentShift{
			GoodEvil: 0.5,
			LawChaos: -0.2,
		},
		Irreversible: true,
	}

	if choice.ChoiceID != "test_choice_1" {
		t.Errorf("Expected ChoiceID test_choice_1, got %s", choice.ChoiceID)
	}
	if !choice.Irreversible {
		t.Error("Expected Irreversible to be true")
	}
}

func TestAlignmentShiftApplication(t *testing.T) {
	alignment := &PlayerAlignment{
		GoodEvil:      0.0,
		LawChaos:      0.0,
		HonorDishonor: 0.0,
	}

	shift := &AlignmentShift{
		GoodEvil:      0.3,
		LawChaos:      -0.2,
		HonorDishonor: 0.1,
	}

	alignment.ApplyShift(shift)

	if alignment.GoodEvil != 0.3 {
		t.Errorf("Expected GoodEvil 0.3, got %f", alignment.GoodEvil)
	}
	if alignment.LawChaos != -0.2 {
		t.Errorf("Expected LawChaos -0.2, got %f", alignment.LawChaos)
	}
	if alignment.HonorDishonor != 0.1 {
		t.Errorf("Expected HonorDishonor 0.1, got %f", alignment.HonorDishonor)
	}
}

func TestAlignmentClamping(t *testing.T) {
	alignment := &PlayerAlignment{
		GoodEvil:      0.9,
		LawChaos:      -0.8,
		HonorDishonor: 0.7,
	}

	shift := &AlignmentShift{
		GoodEvil:      0.5,  // Would exceed 1.0
		LawChaos:      -0.5, // Would exceed -1.0
		HonorDishonor: 0.5,  // Within bounds
	}

	alignment.ApplyShift(shift)

	if alignment.GoodEvil != 1.0 {
		t.Errorf("Expected GoodEvil clamped to 1.0, got %f", alignment.GoodEvil)
	}
	if alignment.LawChaos != -1.0 {
		t.Errorf("Expected LawChaos clamped to -1.0, got %f", alignment.LawChaos)
	}
	if alignment.HonorDishonor != 1.0 {
		t.Errorf("Expected HonorDishonor 1.0, got %f", alignment.HonorDishonor)
	}
}

func TestAlignmentRequirementCheck(t *testing.T) {
	alignment := &PlayerAlignment{
		GoodEvil:      0.5,
		LawChaos:      0.2,
		HonorDishonor: 0.3,
	}

	tests := []struct {
		name string
		req  *AlignmentRequirement
		want bool
	}{
		{
			name: "nil requirement",
			req:  nil,
			want: true,
		},
		{
			name: "within range",
			req: &AlignmentRequirement{
				MinGoodEvil:      0.0,
				MaxGoodEvil:      1.0,
				MinLawChaos:      0.0,
				MaxLawChaos:      1.0,
				MinHonorDishonor: 0.0,
				MaxHonorDishonor: 1.0,
			},
			want: true,
		},
		{
			name: "too evil",
			req: &AlignmentRequirement{
				MinGoodEvil:      0.8, // Requires very good
				MaxGoodEvil:      1.0,
				MinLawChaos:      -1.0,
				MaxLawChaos:      1.0,
				MinHonorDishonor: -1.0,
				MaxHonorDishonor: 1.0,
			},
			want: false,
		},
		{
			name: "too chaotic",
			req: &AlignmentRequirement{
				MinGoodEvil:      -1.0,
				MaxGoodEvil:      1.0,
				MinLawChaos:      0.5, // Requires very lawful
				MaxLawChaos:      1.0,
				MinHonorDishonor: -1.0,
				MaxHonorDishonor: 1.0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := alignment.ChecksAlignment(tt.req)
			if got != tt.want {
				t.Errorf("ChecksAlignment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLockTypeString(t *testing.T) {
	tests := []struct {
		lockType LockType
		want     string
	}{
		{LockTypeQuest, "Quest"},
		{LockTypeNPC, "NPC"},
		{LockTypeArea, "Area"},
		{LockTypeDialogue, "Dialogue"},
		{LockTypeReward, "Reward"},
		{LockTypeCompanion, "Companion"},
		{LockType(999), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.lockType.String()
		if got != tt.want {
			t.Errorf("LockType.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestChoiceTrackerComponentType(t *testing.T) {
	comp := ChoiceTrackerComponent{}
	if comp.Type() != "choice_tracker" {
		t.Errorf("Expected Type() = choice_tracker, got %s", comp.Type())
	}
}

func TestNewChoiceTracker(t *testing.T) {
	tracker := NewChoiceTracker()

	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}
	if tracker.players == nil {
		t.Error("Expected initialized players map")
	}
	if tracker.questBranches == nil {
		t.Error("Expected initialized questBranches map")
	}
	if tracker.npcMemoryLimit != 50 {
		t.Errorf("Expected npcMemoryLimit 50, got %d", tracker.npcMemoryLimit)
	}
	if tracker.choiceLimit != 200 {
		t.Errorf("Expected choiceLimit 200, got %d", tracker.choiceLimit)
	}
}

func TestRecordChoice(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"

	choice := &PlayerChoice{
		ChoiceID:    "choice_save_village",
		StoryNodeID: "village_attack",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &AlignmentShift{
			GoodEvil:      0.3,
			LawChaos:      0.1,
			HonorDishonor: 0.2,
		},
		Irreversible: true,
		NPCsAffected: []string{"villager_elder", "village_guard"},
	}

	err := tracker.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("RecordChoice failed: %v", err)
	}

	// Verify choice was recorded
	count := tracker.GetChoiceCount(playerID)
	if count != 1 {
		t.Errorf("Expected 1 choice, got %d", count)
	}

	// Verify alignment was updated
	alignment := tracker.GetAlignment(playerID)
	if alignment.GoodEvil != 0.3 {
		t.Errorf("Expected GoodEvil 0.3, got %f", alignment.GoodEvil)
	}

	// Verify NPC relationships were created
	npcCount := tracker.GetNPCRelationshipCount(playerID)
	if npcCount != 2 {
		t.Errorf("Expected 2 NPC relationships, got %d", npcCount)
	}
}

func TestRecordChoiceErrors(t *testing.T) {
	tracker := NewChoiceTracker()

	tests := []struct {
		name    string
		choice  *PlayerChoice
		wantErr bool
	}{
		{
			name:    "nil choice",
			choice:  nil,
			wantErr: true,
		},
		{
			name: "empty choice ID",
			choice: &PlayerChoice{
				ChoiceID: "",
			},
			wantErr: true,
		},
		{
			name: "valid choice",
			choice: &PlayerChoice{
				ChoiceID:    "test",
				StoryNodeID: "node",
				Timestamp:   time.Now().Unix(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tracker.RecordChoice("player", tt.choice)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordChoice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChoiceLimitEnforcement(t *testing.T) {
	tracker := NewChoiceTracker()
	tracker.choiceLimit = 10 // Set low limit for testing
	playerID := "player_test"

	// Record 15 choices (5 over limit)
	for i := 0; i < 15; i++ {
		choice := &PlayerChoice{
			ChoiceID:     "choice_" + string(rune('a'+i)),
			StoryNodeID:  "node",
			Timestamp:    time.Now().Unix(),
			Irreversible: i%2 == 0, // Half are irreversible
		}
		tracker.RecordChoice(playerID, choice)
	}

	count := tracker.GetChoiceCount(playerID)
	if count > tracker.choiceLimit {
		t.Errorf("Expected choice count <= %d, got %d", tracker.choiceLimit, count)
	}
}

func TestIsContentAvailable(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"

	// Initially, content should be available
	if !tracker.IsContentAvailable(playerID, "quest_forbidden") {
		t.Error("Expected content to be available initially")
	}

	// Record choice that locks content
	choice := &PlayerChoice{
		ChoiceID:     "choice_betray_guild",
		StoryNodeID:  "guild_confrontation",
		Timestamp:    time.Now().Unix(),
		Irreversible: true,
		Consequences: []string{"lock_quest_forbidden"},
	}

	tracker.RecordChoice(playerID, choice)

	// Now content should be locked
	if tracker.IsContentAvailable(playerID, "forbidden") {
		t.Error("Expected content to be locked after choice")
	}
}

// TestAllLockTypes verifies that all 6 LockType values are correctly parsed.
func TestAllLockTypes(t *testing.T) {
	tests := []struct {
		name         string
		consequence  string
		wantLockType LockType
		wantContent  string
	}{
		{
			name:         "lock_quest prefix",
			consequence:  "lock_quest_main_story",
			wantLockType: LockTypeQuest,
			wantContent:  "main_story",
		},
		{
			name:         "lock_npc prefix",
			consequence:  "lock_npc_merchant_john",
			wantLockType: LockTypeNPC,
			wantContent:  "merchant_john",
		},
		{
			name:         "lock_area prefix",
			consequence:  "lock_area_dark_forest",
			wantLockType: LockTypeArea,
			wantContent:  "dark_forest",
		},
		{
			name:         "lock_dialogue prefix",
			consequence:  "lock_dialogue_peaceful_option",
			wantLockType: LockTypeDialogue,
			wantContent:  "peaceful_option",
		},
		{
			name:         "lock_reward prefix",
			consequence:  "lock_reward_legendary_sword",
			wantLockType: LockTypeReward,
			wantContent:  "legendary_sword",
		},
		{
			name:         "lock_companion prefix",
			consequence:  "lock_companion_aria_knight",
			wantLockType: LockTypeCompanion,
			wantContent:  "aria_knight",
		},
		{
			name:         "unknown prefix defaults to quest",
			consequence:  "some_unknown_consequence",
			wantLockType: LockTypeQuest,
			wantContent:  "some_unknown_consequence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewChoiceTracker()
			playerID := "test_player"

			choice := &PlayerChoice{
				ChoiceID:     "test_choice",
				StoryNodeID:  "test_node",
				Timestamp:    time.Now().Unix(),
				Irreversible: true,
				Consequences: []string{tt.consequence},
			}

			tracker.RecordChoice(playerID, choice)

			// Verify content is locked with correct type
			if tracker.IsContentAvailable(playerID, tt.wantContent) {
				t.Errorf("Expected content %q to be locked", tt.wantContent)
			}

			// Access internal state to verify lock type
			tracker.mu.RLock()
			defer tracker.mu.RUnlock()

			state := tracker.players[playerID]
			if state == nil {
				t.Fatal("Player state not found")
			}

			lock, ok := state.ContentLocks[tt.wantContent]
			if !ok {
				t.Fatalf("Content lock for %q not found", tt.wantContent)
			}

			if lock.LockType != tt.wantLockType {
				t.Errorf("LockType = %v, want %v", lock.LockType, tt.wantLockType)
			}
		})
	}
}

func TestGetNPCAttitude(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"
	npcID := "merchant_john"

	// Initially neutral
	attitude := tracker.GetNPCAttitude(playerID, npcID)
	if attitude != 0.0 {
		t.Errorf("Expected initial attitude 0.0, got %f", attitude)
	}

	// Make positive choice affecting NPC
	choice := &PlayerChoice{
		ChoiceID:    "help_merchant",
		StoryNodeID: "merchant_trouble",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &AlignmentShift{
			GoodEvil:      0.5,
			HonorDishonor: 0.3,
		},
		NPCsAffected: []string{npcID},
	}

	tracker.RecordChoice(playerID, choice)

	attitude = tracker.GetNPCAttitude(playerID, npcID)
	if attitude <= 0.0 {
		t.Errorf("Expected positive attitude, got %f", attitude)
	}
}

func TestRegisterQuestBranch(t *testing.T) {
	tracker := NewChoiceTracker()

	branch := &QuestBranch{
		QuestID:       "quest_civil_war",
		BranchID:      "support_rebels",
		Prerequisites: []string{"choice_meet_rebels", "choice_distrust_king"},
	}

	err := tracker.RegisterQuestBranch(branch)
	if err != nil {
		t.Fatalf("RegisterQuestBranch failed: %v", err)
	}

	// Test errors
	if err := tracker.RegisterQuestBranch(nil); err == nil {
		t.Error("Expected error for nil branch")
	}

	emptyBranch := &QuestBranch{QuestID: ""}
	if err := tracker.RegisterQuestBranch(emptyBranch); err == nil {
		t.Error("Expected error for empty quest ID")
	}
}

func TestIsQuestBranchAvailable(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"

	// Register branch with prerequisites
	branch := &QuestBranch{
		QuestID:       "quest_alliance",
		BranchID:      "join_mages",
		Prerequisites: []string{"choice_study_magic", "choice_reject_warriors"},
	}
	tracker.RegisterQuestBranch(branch)

	// Without prerequisites, branch unavailable
	if tracker.IsQuestBranchAvailable(playerID, "quest_alliance", "join_mages") {
		t.Error("Expected branch unavailable without prerequisites")
	}

	// Record first prerequisite
	tracker.RecordChoice(playerID, &PlayerChoice{
		ChoiceID:    "choice_study_magic",
		StoryNodeID: "magic_academy",
		Timestamp:   time.Now().Unix(),
	})

	// Still missing one prerequisite
	if tracker.IsQuestBranchAvailable(playerID, "quest_alliance", "join_mages") {
		t.Error("Expected branch unavailable with partial prerequisites")
	}

	// Record second prerequisite
	tracker.RecordChoice(playerID, &PlayerChoice{
		ChoiceID:    "choice_reject_warriors",
		StoryNodeID: "warrior_guild",
		Timestamp:   time.Now().Unix(),
	})

	// Now branch should be available
	if !tracker.IsQuestBranchAvailable(playerID, "quest_alliance", "join_mages") {
		t.Error("Expected branch available with all prerequisites")
	}
}

func TestRegisterClassQuest(t *testing.T) {
	tracker := NewChoiceTracker()

	quest := &ClassSpecificQuest{
		QuestID:       "quest_paladin_trial",
		RequiredClass: "Paladin",
		MinLevel:      20,
		AlignmentReq: &AlignmentRequirement{
			MinGoodEvil:      0.5,
			MaxGoodEvil:      1.0,
			MinLawChaos:      0.3,
			MaxLawChaos:      1.0,
			MinHonorDishonor: 0.4,
			MaxHonorDishonor: 1.0,
		},
	}

	err := tracker.RegisterClassQuest(quest)
	if err != nil {
		t.Fatalf("RegisterClassQuest failed: %v", err)
	}

	// Test errors
	if err := tracker.RegisterClassQuest(nil); err == nil {
		t.Error("Expected error for nil quest")
	}

	emptyQuest := &ClassSpecificQuest{QuestID: ""}
	if err := tracker.RegisterClassQuest(emptyQuest); err == nil {
		t.Error("Expected error for empty quest ID")
	}
}

func TestIsClassQuestAvailable(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"

	// Register paladin quest
	quest := &ClassSpecificQuest{
		QuestID:       "quest_holy_light",
		RequiredClass: "Paladin",
		MinLevel:      15,
		AlignmentReq: &AlignmentRequirement{
			MinGoodEvil:      0.5,
			MaxGoodEvil:      1.0,
			MinLawChaos:      0.0,
			MaxLawChaos:      1.0,
			MinHonorDishonor: 0.0,
			MaxHonorDishonor: 1.0,
		},
	}
	tracker.RegisterClassQuest(quest)

	tests := []struct {
		name        string
		playerClass string
		playerLevel int
		alignment   *AlignmentShift
		want        bool
	}{
		{
			name:        "wrong class",
			playerClass: "Warrior",
			playerLevel: 20,
			alignment:   &AlignmentShift{GoodEvil: 0.8},
			want:        false,
		},
		{
			name:        "too low level",
			playerClass: "Paladin",
			playerLevel: 10,
			alignment:   &AlignmentShift{GoodEvil: 0.8},
			want:        false,
		},
		{
			name:        "wrong alignment",
			playerClass: "Paladin",
			playerLevel: 20,
			alignment:   &AlignmentShift{GoodEvil: -0.5}, // Evil
			want:        false,
		},
		{
			name:        "all requirements met",
			playerClass: "Paladin",
			playerLevel: 20,
			alignment:   &AlignmentShift{GoodEvil: 0.8},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPlayerID := playerID + "_" + tt.name

			if tt.alignment != nil {
				choice := &PlayerChoice{
					ChoiceID:       "setup",
					StoryNodeID:    "test",
					Timestamp:      time.Now().Unix(),
					MoralAlignment: tt.alignment,
				}
				tracker.RecordChoice(testPlayerID, choice)
			}

			got := tracker.IsClassQuestAvailable(testPlayerID, "quest_holy_light", tt.playerClass, tt.playerLevel)
			if got != tt.want {
				t.Errorf("IsClassQuestAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecordCompanionReaction(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"

	reaction := &CompanionReaction{
		CompanionID:  "companion_wolf",
		ChoiceID:     "choice_hunt_deer",
		LoyaltyDelta: 0.1,
		Approval:     true,
		Comment:      "Wolf approves of the hunt",
	}

	err := tracker.RecordCompanionReaction(playerID, reaction)
	if err != nil {
		t.Fatalf("RecordCompanionReaction failed: %v", err)
	}

	reactions := tracker.GetCompanionReactions(playerID, "companion_wolf")
	if len(reactions) != 1 {
		t.Errorf("Expected 1 reaction, got %d", len(reactions))
	}

	// Test nil reaction
	if err := tracker.RecordCompanionReaction(playerID, nil); err == nil {
		t.Error("Expected error for nil reaction")
	}
}

func TestCompanionReactionLimit(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_1"
	companionID := "companion_test"

	// Record 25 reactions (should keep only last 20)
	for i := 0; i < 25; i++ {
		reaction := &CompanionReaction{
			CompanionID:  companionID,
			ChoiceID:     "choice_" + string(rune('a'+i)),
			LoyaltyDelta: 0.01,
			Approval:     true,
		}
		tracker.RecordCompanionReaction(playerID, reaction)
	}

	// Get state to check total reactions
	tracker.mu.RLock()
	state := tracker.players[playerID]
	tracker.mu.RUnlock()

	if len(state.CompanionReactions) > 20 {
		t.Errorf("Expected max 20 reactions, got %d", len(state.CompanionReactions))
	}
}

func TestSaveLoad(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_save_test"

	// Record some data
	for i := 0; i < 5; i++ {
		choice := &PlayerChoice{
			ChoiceID:    "choice_" + string(rune('a'+i)),
			StoryNodeID: "node",
			Timestamp:   time.Now().Unix(),
			MoralAlignment: &AlignmentShift{
				GoodEvil: 0.1 * float64(i),
			},
			NPCsAffected: []string{"npc_" + string(rune('a'+i))},
		}
		tracker.RecordChoice(playerID, choice)
	}

	// Save to file
	filename := "test_choices.json.gz"
	defer os.Remove(filename)

	err := tracker.Save(filename)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new tracker and load
	newTracker := NewChoiceTracker()
	err = newTracker.Load(filename)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data
	count := newTracker.GetChoiceCount(playerID)
	if count != 5 {
		t.Errorf("Expected 5 choices after load, got %d", count)
	}

	npcCount := newTracker.GetNPCRelationshipCount(playerID)
	if npcCount != 5 {
		t.Errorf("Expected 5 NPC relationships after load, got %d", npcCount)
	}

	alignment := newTracker.GetAlignment(playerID)
	if alignment.GoodEvil != 1.0 { // Sum of 0.0 + 0.1 + 0.2 + 0.3 + 0.4 = 1.0 (clamped)
		t.Errorf("Expected GoodEvil 1.0 after load, got %f", alignment.GoodEvil)
	}
}

func TestGetAlignment(t *testing.T) {
	tracker := NewChoiceTracker()

	// Non-existent player should return neutral alignment
	alignment := tracker.GetAlignment("nonexistent")
	if alignment.GoodEvil != 0.0 || alignment.LawChaos != 0.0 || alignment.HonorDishonor != 0.0 {
		t.Error("Expected neutral alignment for non-existent player")
	}
}

func TestGetCompanionReactionsEmpty(t *testing.T) {
	tracker := NewChoiceTracker()

	reactions := tracker.GetCompanionReactions("nonexistent", "companion")
	if reactions != nil {
		t.Error("Expected nil reactions for non-existent player")
	}
}

// Benchmark tests
func BenchmarkRecordChoice(b *testing.B) {
	tracker := NewChoiceTracker()
	playerID := "bench_player"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		choice := &PlayerChoice{
			ChoiceID:    "choice_bench",
			StoryNodeID: "node",
			Timestamp:   time.Now().Unix(),
			MoralAlignment: &AlignmentShift{
				GoodEvil: 0.1,
			},
		}
		tracker.RecordChoice(playerID, choice)
	}
}

func BenchmarkIsContentAvailable(b *testing.B) {
	tracker := NewChoiceTracker()
	playerID := "bench_player"

	// Setup some locks
	for i := 0; i < 10; i++ {
		choice := &PlayerChoice{
			ChoiceID:     "choice_" + string(rune('a'+i)),
			StoryNodeID:  "node",
			Timestamp:    time.Now().Unix(),
			Irreversible: true,
			Consequences: []string{"lock_quest_" + string(rune('a'+i))},
		}
		tracker.RecordChoice(playerID, choice)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.IsContentAvailable(playerID, "quest_test")
	}
}

func BenchmarkGetNPCAttitude(b *testing.B) {
	tracker := NewChoiceTracker()
	playerID := "bench_player"
	npcID := "npc_test"

	// Setup relationship
	choice := &PlayerChoice{
		ChoiceID:     "choice_test",
		StoryNodeID:  "node",
		Timestamp:    time.Now().Unix(),
		NPCsAffected: []string{npcID},
	}
	tracker.RecordChoice(playerID, choice)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetNPCAttitude(playerID, npcID)
	}
}

func BenchmarkGetAlignment(b *testing.B) {
	tracker := NewChoiceTracker()
	playerID := "bench_player"

	// Setup alignment
	choice := &PlayerChoice{
		ChoiceID:    "choice_test",
		StoryNodeID: "node",
		Timestamp:   time.Now().Unix(),
		MoralAlignment: &AlignmentShift{
			GoodEvil: 0.5,
		},
	}
	tracker.RecordChoice(playerID, choice)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetAlignment(playerID)
	}
}
