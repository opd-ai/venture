package choice_consequences

import (
	"bytes"
	"os"
	"testing"
	"time"
)

// testTimestamp is a fixed timestamp for deterministic testing
const testTimestamp = int64(1640000000)

// setupTestTime configures a fixed time provider for deterministic tests
func setupTestTime(t *testing.T) {
	t.Helper()
	SetTimeProvider(FixedTimeProvider{Timestamp: testTimestamp})
	t.Cleanup(ResetTimeProvider)
}

func TestPlayerChoiceBasics(t *testing.T) {
	setupTestTime(t)
	choice := &PlayerChoice{
		ChoiceID:    "test_choice_1",
		StoryNodeID: "node_1",
		Timestamp:   testTimestamp,
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
	setupTestTime(t)
	tracker := NewChoiceTracker()
	playerID := "player_1"

	choice := &PlayerChoice{
		ChoiceID:    "choice_save_village",
		StoryNodeID: "village_attack",
		Timestamp:   testTimestamp,
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
	setupTestTime(t)
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
				Timestamp:   testTimestamp,
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
	setupTestTime(t)
	tracker := NewChoiceTracker()
	tracker.choiceLimit = 10 // Set low limit for testing
	playerID := "player_test"

	// Record 15 choices (5 over limit)
	for i := 0; i < 15; i++ {
		choice := &PlayerChoice{
			ChoiceID:     "choice_" + string(rune('a'+i)),
			StoryNodeID:  "node",
			Timestamp:    testTimestamp,
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
	setupTestTime(t)
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
		Timestamp:    testTimestamp,
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
				Timestamp:    testTimestamp,
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
	setupTestTime(t)
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
		Timestamp:   testTimestamp,
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
	setupTestTime(t)
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
		Timestamp:   testTimestamp,
	})

	// Still missing one prerequisite
	if tracker.IsQuestBranchAvailable(playerID, "quest_alliance", "join_mages") {
		t.Error("Expected branch unavailable with partial prerequisites")
	}

	// Record second prerequisite
	tracker.RecordChoice(playerID, &PlayerChoice{
		ChoiceID:    "choice_reject_warriors",
		StoryNodeID: "warrior_guild",
		Timestamp:   testTimestamp,
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
					Timestamp:      testTimestamp,
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
	setupTestTime(t)
	tracker := NewChoiceTracker()
	playerID := "player_save_test"

	// Record some data
	for i := 0; i < 5; i++ {
		choice := &PlayerChoice{
			ChoiceID:    "choice_" + string(rune('a'+i)),
			StoryNodeID: "node",
			Timestamp:   testTimestamp,
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
	setupTestTime(t)
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
			Timestamp:   testTimestamp,
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
			Timestamp:    testTimestamp,
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
		Timestamp:    testTimestamp,
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
		Timestamp:   testTimestamp,
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

func TestTimeProvider(t *testing.T) {
	// Test with fixed time provider
	fixedTime := int64(1234567890)
	SetTimeProvider(FixedTimeProvider{Timestamp: fixedTime})
	defer ResetTimeProvider()

	tracker := NewChoiceTracker()
	playerID := "time_test_player"

	choice := &PlayerChoice{
		ChoiceID:    "test_choice",
		StoryNodeID: "test_node",
		Timestamp:   fixedTime,
		MoralAlignment: &AlignmentShift{
			GoodEvil: 0.5,
		},
	}

	err := tracker.RecordChoice(playerID, choice)
	if err != nil {
		t.Fatalf("RecordChoice failed: %v", err)
	}

	alignment := tracker.GetAlignment(playerID)
	if alignment.UpdatedAt != fixedTime {
		t.Errorf("Expected UpdatedAt = %d, got %d", fixedTime, alignment.UpdatedAt)
	}

	// Verify player state LastUpdate
	tracker.mu.RLock()
	state := tracker.players[playerID]
	tracker.mu.RUnlock()

	if state.LastUpdate != fixedTime {
		t.Errorf("Expected LastUpdate = %d, got %d", fixedTime, state.LastUpdate)
	}
}

func TestRealTimeProvider(t *testing.T) {
	rtp := RealTimeProvider{}
	before := time.Now().Unix()
	ts := rtp.Now()
	after := time.Now().Unix()

	if ts < before || ts > after {
		t.Errorf("RealTimeProvider.Now() = %d, expected between %d and %d", ts, before, after)
	}
}

func TestIsContentAvailableThreadSafety(t *testing.T) {
	setupTestTime(t)
	// Test that IsContentAvailable properly handles the unlock path
	// without data races (uses write lock when deleting)
	tracker := NewChoiceTracker()
	playerID := "thread_safety_test"

	// Create a non-permanent lock with unlock conditions
	choice := &PlayerChoice{
		ChoiceID:     "initial_choice",
		StoryNodeID:  "test_node",
		Timestamp:    testTimestamp,
		Irreversible: true,
		Consequences: []string{"lock_quest_test_content"},
	}
	tracker.RecordChoice(playerID, choice)

	// Modify the lock to be non-permanent with conditions
	tracker.mu.Lock()
	state := tracker.players[playerID]
	lock := state.ContentLocks["test_content"]
	lock.Permanent = false
	lock.UnlockConditions = []string{"unlock_choice"}
	tracker.mu.Unlock()

	// Content should be locked since we don't have the unlock choice
	if tracker.IsContentAvailable(playerID, "test_content") {
		t.Error("Expected content to be locked without unlock choice")
	}

	// Record the unlock choice
	unlockChoice := &PlayerChoice{
		ChoiceID:    "unlock_choice",
		StoryNodeID: "test_node",
		Timestamp:   testTimestamp,
	}
	tracker.RecordChoice(playerID, unlockChoice)

	// Now content should be available (and lock removed)
	if !tracker.IsContentAvailable(playerID, "test_content") {
		t.Error("Expected content to be available after unlock choice")
	}

	// Verify lock was removed
	tracker.mu.RLock()
	_, stillLocked := tracker.players[playerID].ContentLocks["test_content"]
	tracker.mu.RUnlock()

	if stillLocked {
		t.Error("Expected lock to be removed after unlock conditions met")
	}
}

func TestChoiceTrackerComponentSerialize(t *testing.T) {
	tests := []struct {
		name string
		comp *ChoiceTrackerComponent
	}{
		{
			name: "empty component",
			comp: &ChoiceTrackerComponent{
				PlayerID: "player_1",
			},
		},
		{
			name: "component with data",
			comp: &ChoiceTrackerComponent{
				PlayerID: "player_2",
				ChoiceHistory: []*PlayerChoice{
					{
						ChoiceID:     "choice_1",
						StoryNodeID:  "node_1",
						Timestamp:    1234567890,
						Irreversible: true,
					},
				},
				NPCRelationships: map[string]*NPCRelationship{
					"npc_1": {
						NPCID:      "npc_1",
						Attitude:   0.5,
						TrustLevel: 0.3,
					},
				},
				ContentLocks: map[string]*ContentLock{
					"quest_1": {
						ContentID: "quest_1",
						LockedBy:  "choice_1",
						LockType:  LockTypeQuest,
						Permanent: true,
					},
				},
				Alignment: &PlayerAlignment{
					GoodEvil:      0.5,
					LawChaos:      -0.2,
					HonorDishonor: 0.3,
				},
				CompanionReactions: []*CompanionReaction{
					{
						CompanionID:  "companion_1",
						ChoiceID:     "choice_1",
						LoyaltyDelta: 0.1,
						Approval:     true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := tt.comp.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}
			if len(data) == 0 {
				t.Error("Serialize() returned empty data")
			}

			// Deserialize into new component
			newComp := &ChoiceTrackerComponent{}
			if err := newComp.Deserialize(data); err != nil {
				t.Fatalf("Deserialize() error = %v", err)
			}

			// Verify round-trip
			if newComp.PlayerID != tt.comp.PlayerID {
				t.Errorf("PlayerID = %v, want %v", newComp.PlayerID, tt.comp.PlayerID)
			}
			if len(newComp.ChoiceHistory) != len(tt.comp.ChoiceHistory) {
				t.Errorf("ChoiceHistory len = %v, want %v", len(newComp.ChoiceHistory), len(tt.comp.ChoiceHistory))
			}
			if len(newComp.NPCRelationships) != len(tt.comp.NPCRelationships) {
				t.Errorf("NPCRelationships len = %v, want %v", len(newComp.NPCRelationships), len(tt.comp.NPCRelationships))
			}
			if len(newComp.ContentLocks) != len(tt.comp.ContentLocks) {
				t.Errorf("ContentLocks len = %v, want %v", len(newComp.ContentLocks), len(tt.comp.ContentLocks))
			}
		})
	}
}

func TestChoiceTrackerComponentDeserializeErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    []byte("not valid json"),
			wantErr: true,
		},
		{
			name:    "valid JSON",
			data:    []byte(`{"PlayerID":"test"}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &ChoiceTrackerComponent{}
			err := comp.Deserialize(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkSerialize(b *testing.B) {
	comp := &ChoiceTrackerComponent{
		PlayerID:      "benchmark_player",
		ChoiceHistory: make([]*PlayerChoice, 100),
		NPCRelationships: map[string]*NPCRelationship{
			"npc_1": {NPCID: "npc_1", Attitude: 0.5},
		},
		ContentLocks: make(map[string]*ContentLock),
		Alignment: &PlayerAlignment{
			GoodEvil:      0.5,
			LawChaos:      -0.2,
			HonorDishonor: 0.3,
		},
	}

	for i := 0; i < 100; i++ {
		comp.ChoiceHistory[i] = &PlayerChoice{
			ChoiceID:    "choice_" + string(rune('a'+i%26)),
			StoryNodeID: "node_1",
			Timestamp:   int64(1234567890 + i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Serialize()
	}
}

func BenchmarkDeserialize(b *testing.B) {
	comp := &ChoiceTrackerComponent{
		PlayerID:      "benchmark_player",
		ChoiceHistory: make([]*PlayerChoice, 100),
		NPCRelationships: map[string]*NPCRelationship{
			"npc_1": {NPCID: "npc_1", Attitude: 0.5},
		},
		ContentLocks: make(map[string]*ContentLock),
		Alignment: &PlayerAlignment{
			GoodEvil:      0.5,
			LawChaos:      -0.2,
			HonorDishonor: 0.3,
		},
	}

	for i := 0; i < 100; i++ {
		comp.ChoiceHistory[i] = &PlayerChoice{
			ChoiceID:    "choice_" + string(rune('a'+i%26)),
			StoryNodeID: "node_1",
			Timestamp:   int64(1234567890 + i),
		}
	}

	data, _ := comp.Serialize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newComp := &ChoiceTrackerComponent{}
		newComp.Deserialize(data)
	}
}

// TestSaveToLoadFrom tests the WASM-compatible io.Reader/io.Writer based save/load methods.
func TestSaveToLoadFrom(t *testing.T) {
	tracker := NewChoiceTracker()
	playerID := "player_io_test"

	// Record some data
	for i := 0; i < 5; i++ {
		choice := &PlayerChoice{
			ChoiceID:    "choice_io_" + string(rune('a'+i)),
			StoryNodeID: "node_io",
			Timestamp:   int64(1700000000 + i), // Fixed timestamp for determinism
			MoralAlignment: &AlignmentShift{
				GoodEvil: 0.1 * float64(i),
				LawChaos: 0.05 * float64(i),
			},
			NPCsAffected: []string{"npc_io_" + string(rune('a'+i))},
		}
		tracker.RecordChoice(playerID, choice)
	}

	// Save to bytes.Buffer (simulating WASM localStorage)
	var buf bytes.Buffer
	err := tracker.SaveTo(&buf)
	if err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	// Verify buffer has content
	if buf.Len() == 0 {
		t.Fatal("SaveTo produced empty buffer")
	}

	// Create new tracker and load from buffer
	newTracker := NewChoiceTracker()
	err = newTracker.LoadFrom(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	// Verify data
	count := newTracker.GetChoiceCount(playerID)
	if count != 5 {
		t.Errorf("Expected 5 choices after LoadFrom, got %d", count)
	}

	npcCount := newTracker.GetNPCRelationshipCount(playerID)
	if npcCount != 5 {
		t.Errorf("Expected 5 NPC relationships after LoadFrom, got %d", npcCount)
	}

	alignment := newTracker.GetAlignment(playerID)
	if alignment.GoodEvil != 1.0 { // Sum of 0.0 + 0.1 + 0.2 + 0.3 + 0.4 = 1.0 (clamped)
		t.Errorf("Expected GoodEvil 1.0 after LoadFrom, got %f", alignment.GoodEvil)
	}
}

// TestSaveToLoadFromEmptyTracker verifies SaveTo/LoadFrom work with empty tracker.
func TestSaveToLoadFromEmptyTracker(t *testing.T) {
	tracker := NewChoiceTracker()

	// Save empty tracker
	var buf bytes.Buffer
	err := tracker.SaveTo(&buf)
	if err != nil {
		t.Fatalf("SaveTo empty tracker failed: %v", err)
	}

	// Load into new tracker
	newTracker := NewChoiceTracker()
	err = newTracker.LoadFrom(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("LoadFrom empty data failed: %v", err)
	}

	// Verify empty state
	if newTracker.GetChoiceCount("any") != 0 {
		t.Error("Expected 0 choices for empty tracker")
	}
}

// TestLoadFromInvalidData verifies LoadFrom handles corrupt data gracefully.
func TestLoadFromInvalidData(t *testing.T) {
	tracker := NewChoiceTracker()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"garbage", []byte("not gzip data at all")},
		{"truncated_gzip", []byte{0x1f, 0x8b, 0x08}}, // Valid gzip header, truncated
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tracker.LoadFrom(bytes.NewReader(tc.data))
			if err == nil {
				t.Error("Expected error for invalid data, got nil")
			}
		})
	}
}

// BenchmarkSaveToLoadFrom benchmarks the io-based save/load cycle.
func BenchmarkSaveToLoadFrom(b *testing.B) {
	tracker := NewChoiceTracker()

	// Populate with data
	for p := 0; p < 10; p++ {
		playerID := "player_" + string(rune('a'+p))
		for i := 0; i < 50; i++ {
			choice := &PlayerChoice{
				ChoiceID:    "choice_" + string(rune('a'+i%26)),
				StoryNodeID: "node_bench",
				Timestamp:   int64(1700000000 + i),
				MoralAlignment: &AlignmentShift{
					GoodEvil: 0.1,
				},
				NPCsAffected: []string{"npc_" + string(rune('a'+i%26))},
			}
			tracker.RecordChoice(playerID, choice)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		tracker.SaveTo(&buf)

		newTracker := NewChoiceTracker()
		newTracker.LoadFrom(bytes.NewReader(buf.Bytes()))
	}
}
