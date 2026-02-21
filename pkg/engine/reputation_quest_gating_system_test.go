package engine

import (
	"testing"
)

func TestReputationTier_String(t *testing.T) {
	tests := []struct {
		tier     ReputationTier
		expected string
	}{
		{TierHated, "Hated"},
		{TierHostile, "Hostile"},
		{TierUnfriendly, "Unfriendly"},
		{TierNeutral, "Neutral"},
		{TierFriendly, "Friendly"},
		{TierHonored, "Honored"},
		{TierRevered, "Revered"},
		{ReputationTier(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.expected {
				t.Errorf("ReputationTier.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReputationTier_MinReputation(t *testing.T) {
	tests := []struct {
		tier     ReputationTier
		expected float64
	}{
		{TierHated, -100.0},
		{TierHostile, -75.0},
		{TierUnfriendly, -50.0},
		{TierNeutral, -25.0},
		{TierFriendly, 25.0},
		{TierHonored, 50.0},
		{TierRevered, 75.0},
	}

	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			if got := tt.tier.MinReputation(); got != tt.expected {
				t.Errorf("ReputationTier.MinReputation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReputationFromValue(t *testing.T) {
	tests := []struct {
		name     string
		rep      float64
		expected ReputationTier
	}{
		{"highly revered", 100.0, TierRevered},
		{"revered boundary", 75.0, TierRevered},
		{"honored high", 74.9, TierHonored},
		{"honored boundary", 50.0, TierHonored},
		{"friendly high", 49.9, TierFriendly},
		{"friendly boundary", 25.0, TierFriendly},
		{"neutral positive", 24.9, TierNeutral},
		{"neutral zero", 0.0, TierNeutral},
		{"neutral negative", -24.9, TierNeutral},
		{"unfriendly boundary", -25.0, TierUnfriendly},
		{"unfriendly high", -49.9, TierUnfriendly},
		{"hostile boundary", -50.0, TierHostile},
		{"hostile high", -74.9, TierHostile},
		{"hated boundary", -75.0, TierHated},
		{"deeply hated", -100.0, TierHated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReputationFromValue(tt.rep); got != tt.expected {
				t.Errorf("ReputationFromValue(%v) = %v, want %v", tt.rep, got, tt.expected)
			}
		})
	}
}

func TestQuestReputationRequirement_MeetsRequirement(t *testing.T) {
	tests := []struct {
		name    string
		req     QuestReputationRequirement
		rep     float64
		expects bool
	}{
		{
			name:    "friendly meets friendly",
			req:     QuestReputationRequirement{FactionID: "guild", MinTier: TierFriendly, MaxTier: TierRevered},
			rep:     30.0,
			expects: true,
		},
		{
			name:    "neutral fails friendly",
			req:     QuestReputationRequirement{FactionID: "guild", MinTier: TierFriendly, MaxTier: TierRevered},
			rep:     0.0,
			expects: false,
		},
		{
			name:    "revered meets honored",
			req:     QuestReputationRequirement{FactionID: "guild", MinTier: TierHonored, MaxTier: TierRevered},
			rep:     80.0,
			expects: true,
		},
		{
			name:    "revered fails max honored",
			req:     QuestReputationRequirement{FactionID: "guild", MinTier: TierNeutral, MaxTier: TierHonored},
			rep:     80.0,
			expects: false,
		},
		{
			name:    "hostile meets hostile-only",
			req:     QuestReputationRequirement{FactionID: "enemy", MinTier: TierHostile, MaxTier: TierHostile},
			rep:     -60.0,
			expects: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.MeetsRequirement(tt.rep); got != tt.expects {
				t.Errorf("MeetsRequirement(%v) = %v, want %v", tt.rep, got, tt.expects)
			}
		})
	}
}

func TestReputationQuestGatingComponent_Type(t *testing.T) {
	comp := NewReputationQuestGatingComponent()
	if got := comp.Type(); got != "reputation_quest_gating" {
		t.Errorf("Type() = %v, want reputation_quest_gating", got)
	}
}

func TestReputationQuestGatingComponent_RegisterAndCheckQuest(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	quest := &GatedQuest{
		QuestID:   "test_quest",
		FactionID: "guild",
		Requirements: []QuestReputationRequirement{
			{FactionID: "guild", MinTier: TierFriendly, MaxTier: TierRevered},
		},
	}
	comp.RegisterGatedQuest(quest)

	// Test with neutral reputation (should fail)
	factionReps := map[string]float64{"guild": 0.0}
	if comp.IsQuestAvailable("test_quest", factionReps) {
		t.Error("Quest should not be available at neutral reputation")
	}

	// Test with friendly reputation (should pass)
	factionReps["guild"] = 30.0
	if !comp.IsQuestAvailable("test_quest", factionReps) {
		t.Error("Quest should be available at friendly reputation")
	}
}

func TestReputationQuestGatingComponent_LockedOutQuests(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	quest := &GatedQuest{
		QuestID:   "locked_quest",
		FactionID: "guild",
		Requirements: []QuestReputationRequirement{
			{FactionID: "guild", MinTier: TierNeutral, MaxTier: TierRevered},
		},
	}
	comp.RegisterGatedQuest(quest)

	// Quest should be available at neutral
	factionReps := map[string]float64{"guild": 0.0}
	if !comp.IsQuestAvailable("locked_quest", factionReps) {
		t.Error("Quest should be available before lockout")
	}

	// Lock out the quest
	comp.MarkQuestLockedOut("locked_quest", "chose opposing faction")

	// Quest should no longer be available
	if comp.IsQuestAvailable("locked_quest", factionReps) {
		t.Error("Quest should not be available after lockout")
	}

	// Check lockout reason
	reason := comp.GetQuestBlockReason("locked_quest", factionReps)
	if reason != "chose opposing faction" {
		t.Errorf("GetQuestBlockReason() = %v, want 'chose opposing faction'", reason)
	}
}

func TestReputationQuestGatingComponent_ExclusiveQuests(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	// Register two competing faction quests
	questA := &GatedQuest{
		QuestID:          "faction_a_final",
		FactionID:        "faction_a",
		IsExclusive:      true,
		ExcludesFactions: []string{"faction_b"},
		Requirements: []QuestReputationRequirement{
			{FactionID: "faction_a", MinTier: TierHonored, MaxTier: TierRevered},
		},
	}
	questB := &GatedQuest{
		QuestID:   "faction_b_quest",
		FactionID: "faction_b",
		Requirements: []QuestReputationRequirement{
			{FactionID: "faction_b", MinTier: TierFriendly, MaxTier: TierRevered},
		},
	}
	comp.RegisterGatedQuest(questA)
	comp.RegisterGatedQuest(questB)

	// Complete faction A's exclusive quest
	lockedOut := comp.RecordFactionQuestCompletion("faction_a_final")

	// Faction B quests should now be locked
	if len(lockedOut) != 1 || lockedOut[0] != "faction_b_quest" {
		t.Errorf("Expected faction_b_quest to be locked out, got %v", lockedOut)
	}

	// Verify lockout
	factionReps := map[string]float64{"faction_b": 30.0}
	if comp.IsQuestAvailable("faction_b_quest", factionReps) {
		t.Error("Faction B quest should be locked out after completing faction A exclusive")
	}
}

func TestReputationQuestGatingComponent_RecentNotifications(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	// Mark some unlocks and lockouts
	comp.MarkQuestUnlocked("quest1")
	comp.MarkQuestUnlocked("quest2")
	comp.MarkQuestLockedOut("quest3", "reason")

	// Get recent unlocks
	unlocks := comp.GetRecentUnlocks()
	if len(unlocks) != 2 {
		t.Errorf("Expected 2 recent unlocks, got %d", len(unlocks))
	}

	// Calling again should return empty (consumed)
	unlocks = comp.GetRecentUnlocks()
	if len(unlocks) != 0 {
		t.Errorf("Expected 0 recent unlocks after consumption, got %d", len(unlocks))
	}

	// Get recent lockouts
	lockouts := comp.GetRecentLockouts()
	if len(lockouts) != 1 {
		t.Errorf("Expected 1 recent lockout, got %d", len(lockouts))
	}
}

func TestReputationQuestGatingComponent_FactionProgress(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	quest := &GatedQuest{
		QuestID:   "guild_quest1",
		FactionID: "guild",
	}
	comp.RegisterGatedQuest(quest)

	// Initial count should be 0
	if count := comp.GetFactionQuestCount("guild"); count != 0 {
		t.Errorf("Expected 0 faction quests, got %d", count)
	}

	// Complete the quest
	comp.RecordFactionQuestCompletion("guild_quest1")

	// Count should increment
	if count := comp.GetFactionQuestCount("guild"); count != 1 {
		t.Errorf("Expected 1 faction quest, got %d", count)
	}
}

func TestReputationQuestGatingComponent_GetBlockReason(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	quest := &GatedQuest{
		QuestID:   "honor_quest",
		FactionID: "guild",
		Requirements: []QuestReputationRequirement{
			{
				FactionID:      "guild",
				MinTier:        TierHonored,
				MaxTier:        TierRevered,
				FailureMessage: "You must be Honored with the guild",
			},
		},
	}
	comp.RegisterGatedQuest(quest)

	// Check block reason at neutral
	factionReps := map[string]float64{"guild": 0.0}
	reason := comp.GetQuestBlockReason("honor_quest", factionReps)
	if reason != "You must be Honored with the guild" {
		t.Errorf("GetQuestBlockReason() = %v, want custom message", reason)
	}
}

func TestReputationQuestGatingComponent_UnregisteredQuest(t *testing.T) {
	comp := NewReputationQuestGatingComponent()

	// Unregistered quests should be available
	factionReps := map[string]float64{"guild": 0.0}
	if !comp.IsQuestAvailable("unregistered_quest", factionReps) {
		t.Error("Unregistered quests should be available")
	}

	// Block reason should be empty
	reason := comp.GetQuestBlockReason("unregistered_quest", factionReps)
	if reason != "" {
		t.Errorf("Block reason for unregistered quest should be empty, got %v", reason)
	}
}

func TestReputationQuestGatingSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	// Create entity with both components
	entity := world.CreateEntity()
	repComp := NewReputationComponent()
	gatingComp := NewReputationQuestGatingComponent()

	entity.AddComponent(repComp)
	entity.AddComponent(gatingComp)

	// Register a quest requiring friendly standing
	gatedQuest := &GatedQuest{
		QuestID:   "friendly_quest",
		FactionID: "guild",
		Requirements: []QuestReputationRequirement{
			{FactionID: "guild", MinTier: TierFriendly, MaxTier: TierRevered},
		},
		UnlockMessage: "New guild quest available!",
	}
	gatingComp.RegisterGatedQuest(gatedQuest)

	// Initially at neutral - quest should not unlock
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	if gatingComp.UnlockedQuests["friendly_quest"] {
		t.Error("Quest should not be unlocked at neutral reputation")
	}

	// Increase reputation to friendly
	repComp.SetReputation("guild", 30.0)
	system.Update(entities, 0.016)

	if !gatingComp.UnlockedQuests["friendly_quest"] {
		t.Error("Quest should be unlocked at friendly reputation")
	}
}

func TestReputationQuestGatingSystem_RegisterFactionQuests(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	entity := world.CreateEntity()
	repComp := NewReputationComponent()
	entity.AddComponent(repComp)

	configs := []FactionQuestConfig{
		{QuestID: "guild_q1", MinTier: TierNeutral},
		{QuestID: "guild_q2", MinTier: TierFriendly},
		{QuestID: "guild_q3", MinTier: TierHonored},
	}

	system.RegisterFactionQuests(entity, "guild", configs)

	// Get gating component that was created
	comp, _ := entity.GetComponent("reputation_quest_gating")
	gatingComp := comp.(*ReputationQuestGatingComponent)

	// All three quests should be registered
	if len(gatingComp.GatedQuests) != 3 {
		t.Errorf("Expected 3 gated quests, got %d", len(gatingComp.GatedQuests))
	}
}

func TestReputationQuestGatingSystem_GetAvailableQuests(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	entity := world.CreateEntity()
	repComp := NewReputationComponent()
	gatingComp := NewReputationQuestGatingComponent()
	entity.AddComponent(repComp)
	entity.AddComponent(gatingComp)

	// Register quests at different tiers
	gatingComp.RegisterGatedQuest(&GatedQuest{
		QuestID: "q_neutral", FactionID: "guild",
		Requirements: []QuestReputationRequirement{{FactionID: "guild", MinTier: TierNeutral, MaxTier: TierRevered}},
	})
	gatingComp.RegisterGatedQuest(&GatedQuest{
		QuestID: "q_friendly", FactionID: "guild",
		Requirements: []QuestReputationRequirement{{FactionID: "guild", MinTier: TierFriendly, MaxTier: TierRevered}},
	})
	gatingComp.RegisterGatedQuest(&GatedQuest{
		QuestID: "q_honored", FactionID: "guild",
		Requirements: []QuestReputationRequirement{{FactionID: "guild", MinTier: TierHonored, MaxTier: TierRevered}},
	})

	// At neutral - only neutral quest available
	available := system.GetAvailableQuests(entity)
	if len(available) != 1 {
		t.Errorf("Expected 1 available quest at neutral, got %d", len(available))
	}

	// At friendly - neutral and friendly available
	repComp.SetReputation("guild", 30.0)
	available = system.GetAvailableQuests(entity)
	if len(available) != 2 {
		t.Errorf("Expected 2 available quests at friendly, got %d", len(available))
	}

	// At honored - all available
	repComp.SetReputation("guild", 55.0)
	available = system.GetAvailableQuests(entity)
	if len(available) != 3 {
		t.Errorf("Expected 3 available quests at honored, got %d", len(available))
	}
}

func TestReputationQuestGatingSystem_GetBlockedQuests(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	entity := world.CreateEntity()
	repComp := NewReputationComponent()
	gatingComp := NewReputationQuestGatingComponent()
	entity.AddComponent(repComp)
	entity.AddComponent(gatingComp)

	gatingComp.RegisterGatedQuest(&GatedQuest{
		QuestID: "blocked_quest", FactionID: "guild",
		Requirements: []QuestReputationRequirement{
			{FactionID: "guild", MinTier: TierRevered, MaxTier: TierRevered, FailureMessage: "Must be Revered"},
		},
	})

	blocked := system.GetBlockedQuests(entity)
	if len(blocked) != 1 {
		t.Errorf("Expected 1 blocked quest, got %d", len(blocked))
	}
	if blocked["blocked_quest"] != "Must be Revered" {
		t.Errorf("Wrong block reason: %v", blocked["blocked_quest"])
	}
}

func TestReputationQuestGatingSystem_OnQuestCompleted(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	entity := world.CreateEntity()
	repComp := NewReputationComponent()
	gatingComp := NewReputationQuestGatingComponent()
	entity.AddComponent(repComp)
	entity.AddComponent(gatingComp)

	gatingComp.RegisterGatedQuest(&GatedQuest{
		QuestID:               "bonus_quest",
		FactionID:             "guild",
		RewardReputationBonus: 15.0,
	})

	// Initial reputation
	repComp.SetReputation("guild", 10.0)

	// Complete quest
	system.OnQuestCompleted(entity, "bonus_quest")

	// Check reputation was increased
	if rep := repComp.GetReputation("guild"); rep != 25.0 {
		t.Errorf("Expected reputation 25.0, got %v", rep)
	}
}

func TestReputationQuestGatingSystem_GenerateFactionQuestLine(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	configs := system.GenerateFactionQuestLine("guild", 54321, 5)

	if len(configs) != 5 {
		t.Errorf("Expected 5 quest configs, got %d", len(configs))
	}

	// Check progression of tiers
	expectedTiers := []ReputationTier{TierNeutral, TierFriendly, TierFriendly, TierHonored, TierHonored}
	for i, config := range configs {
		if config.MinTier != expectedTiers[i] {
			t.Errorf("Quest %d: expected tier %v, got %v", i, expectedTiers[i], config.MinTier)
		}
	}

	// Last quest should be exclusive
	if !configs[4].IsExclusive {
		t.Error("Final quest should be exclusive")
	}
}

func TestReputationQuestGatingSystem_NilEntity(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	// Should handle nil entities gracefully
	entities := []*Entity{nil}
	system.Update(entities, 0.016)
	// No panic = pass
}

func TestReputationQuestGatingSystem_MissingComponents(t *testing.T) {
	world := NewWorld()
	system := NewReputationQuestGatingSystem(world, nil, 12345)

	// Entity without components
	entity := world.CreateEntity()
	entities := []*Entity{entity}
	system.Update(entities, 0.016)
	// No panic = pass

	// Entity with only reputation
	entity2 := world.CreateEntity()
	entity2.AddComponent(NewReputationComponent())
	entities = []*Entity{entity2}
	system.Update(entities, 0.016)
	// No panic = pass

	// Entity with only gating
	entity3 := world.CreateEntity()
	entity3.AddComponent(NewReputationQuestGatingComponent())
	entities = []*Entity{entity3}
	system.Update(entities, 0.016)
	// No panic = pass
}
