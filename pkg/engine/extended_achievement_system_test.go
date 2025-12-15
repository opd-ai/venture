package engine

import (
	"testing"
	"time"
)

func TestGetAchievementDefinition(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantName string
		wantOK   bool
	}{
		{"combat_first_blood exists", "combat_first_blood", "First Blood", true},
		{"combat_boss_slayer exists", "combat_boss_slayer", "Boss Slayer", true},
		{"quest_adventurer exists", "quest_adventurer", "Adventurer", true},
		{"craft_apprentice exists", "craft_apprentice", "Apprentice Crafter", true},
		{"explore_wanderer exists", "explore_wanderer", "Wanderer", true},
		{"social_friend_maker exists", "social_friend_maker", "Friend Maker", true},
		{"pvp_gladiator exists", "pvp_gladiator", "Gladiator", true},
		{"non_existent does not exist", "non_existent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, ok := GetAchievementDefinition(tt.id)

			if ok != tt.wantOK {
				t.Errorf("GetAchievementDefinition() ok = %v, want %v", ok, tt.wantOK)
			}

			if ok && def.Name != tt.wantName {
				t.Errorf("GetAchievementDefinition() name = %v, want %v", def.Name, tt.wantName)
			}
		})
	}
}

func TestGetAllAchievementDefinitions(t *testing.T) {
	defs := GetAllAchievementDefinitions()

	if len(defs) < 60 {
		t.Errorf("Expected at least 60 achievement definitions, got %d", len(defs))
	}

	// Check that each category has at least 10 achievements
	categories := make(map[AchievementCategory]int)
	for _, def := range defs {
		categories[def.Category]++
	}

	for cat := AchievementCategoryCombat; cat <= AchievementCategoryPvP; cat++ {
		if categories[cat] < 10 {
			t.Errorf("Category %v has %d achievements, want at least 10", cat.String(), categories[cat])
		}
	}
}

func TestGetAchievementDefinitionsByCategory(t *testing.T) {
	combatDefs := GetAchievementDefinitionsByCategory(AchievementCategoryCombat)

	if len(combatDefs) < 10 {
		t.Errorf("Combat category has %d achievements, want at least 10", len(combatDefs))
	}

	for _, def := range combatDefs {
		if def.Category != AchievementCategoryCombat {
			t.Errorf("Achievement %s has category %v, want Combat", def.ID, def.Category)
		}
	}
}

func TestGetAchievementCount(t *testing.T) {
	count := GetAchievementCount()

	if count < 60 {
		t.Errorf("GetAchievementCount() = %d, want at least 60", count)
	}
}

func TestNewExtendedAchievementSystem(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	if system == nil {
		t.Fatal("NewExtendedAchievementSystem() returned nil")
	}

	if system.world != world {
		t.Error("System world reference incorrect")
	}
}

func TestExtendedAchievementSystem_RecordProgress(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	// Create an entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Record progress for first blood
	unlocked := system.RecordProgress(entity.ID, "combat_first_blood", 1)

	if !unlocked {
		t.Error("RecordProgress should unlock bronze tier at 1 kill")
	}

	tier := system.GetTier(entity.ID, "combat_first_blood")
	if tier != AchievementTierBronze {
		t.Errorf("Tier = %v, want Bronze", tier)
	}

	progress := system.GetProgress(entity.ID, "combat_first_blood")
	if progress != 1 {
		t.Errorf("Progress = %v, want 1", progress)
	}
}

func TestExtendedAchievementSystem_IncrementProgress(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Increment multiple times
	system.IncrementProgress(entity.ID, "combat_first_blood", 1)
	system.IncrementProgress(entity.ID, "combat_first_blood", 1)
	system.IncrementProgress(entity.ID, "combat_first_blood", 1)

	progress := system.GetProgress(entity.ID, "combat_first_blood")
	if progress != 3 {
		t.Errorf("Progress = %v, want 3", progress)
	}
}

func TestExtendedAchievementSystem_GetTotalPoints(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// No achievements yet
	points := system.GetTotalPoints(entity.ID)
	if points != 0 {
		t.Errorf("TotalPoints = %v, want 0", points)
	}

	// Unlock bronze (10 points)
	system.RecordProgress(entity.ID, "combat_first_blood", 1)

	points = system.GetTotalPoints(entity.ID)
	if points != 10 {
		t.Errorf("TotalPoints = %v, want 10", points)
	}

	// Unlock another bronze (10 more points)
	system.RecordProgress(entity.ID, "quest_adventurer", 1)

	points = system.GetTotalPoints(entity.ID)
	if points != 20 {
		t.Errorf("TotalPoints = %v, want 20", points)
	}
}

func TestExtendedAchievementSystem_GetCategoryPoints(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Unlock combat achievement
	system.RecordProgress(entity.ID, "combat_first_blood", 1)

	combatPoints := system.GetCategoryPoints(entity.ID, AchievementCategoryCombat)
	if combatPoints != 10 {
		t.Errorf("Combat points = %v, want 10", combatPoints)
	}

	questPoints := system.GetCategoryPoints(entity.ID, AchievementCategoryQuest)
	if questPoints != 0 {
		t.Errorf("Quest points = %v, want 0", questPoints)
	}
}

func TestExtendedAchievementSystem_OnUnlockCallback(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	callbackCalled := false
	var callbackEntityID uint64
	var callbackAchievementID string
	var callbackTier AchievementTier

	system.SetOnUnlockCallback(func(entityID uint64, achievementID string, tier AchievementTier) {
		callbackCalled = true
		callbackEntityID = entityID
		callbackAchievementID = achievementID
		callbackTier = tier
	})

	system.RecordProgress(entity.ID, "combat_first_blood", 1)

	if !callbackCalled {
		t.Error("Callback should have been called")
	}

	if callbackEntityID != entity.ID {
		t.Errorf("Callback entityID = %v, want %v", callbackEntityID, entity.ID)
	}

	if callbackAchievementID != "combat_first_blood" {
		t.Errorf("Callback achievementID = %v, want combat_first_blood", callbackAchievementID)
	}

	if callbackTier != AchievementTierBronze {
		t.Errorf("Callback tier = %v, want Bronze", callbackTier)
	}
}

func TestExtendedAchievementSystem_OnEnemyKilled(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Kill a regular enemy
	system.OnEnemyKilled(entity.ID, false)

	progress := system.GetProgress(entity.ID, "combat_first_blood")
	if progress != 1 {
		t.Errorf("first_blood progress = %v, want 1", progress)
	}

	bossProgress := system.GetProgress(entity.ID, "combat_boss_slayer")
	if bossProgress != 0 {
		t.Errorf("boss_slayer progress = %v, want 0", bossProgress)
	}

	// Kill a boss
	system.OnEnemyKilled(entity.ID, true)

	progress = system.GetProgress(entity.ID, "combat_first_blood")
	if progress != 2 {
		t.Errorf("first_blood progress = %v, want 2", progress)
	}

	bossProgress = system.GetProgress(entity.ID, "combat_boss_slayer")
	if bossProgress != 1 {
		t.Errorf("boss_slayer progress = %v, want 1", bossProgress)
	}
}

func TestExtendedAchievementSystem_OnDamageDealt(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Deal regular damage
	system.OnDamageDealt(entity.ID, 100, false)

	damageProgress := system.GetProgress(entity.ID, "combat_damage_dealer")
	if damageProgress != 100 {
		t.Errorf("damage_dealer progress = %v, want 100", damageProgress)
	}

	critProgress := system.GetProgress(entity.ID, "combat_critical_master")
	if critProgress != 0 {
		t.Errorf("critical_master progress = %v, want 0", critProgress)
	}

	// Deal critical damage
	system.OnDamageDealt(entity.ID, 200, true)

	damageProgress = system.GetProgress(entity.ID, "combat_damage_dealer")
	if damageProgress != 300 {
		t.Errorf("damage_dealer progress = %v, want 300", damageProgress)
	}

	critProgress = system.GetProgress(entity.ID, "combat_critical_master")
	if critProgress != 1 {
		t.Errorf("critical_master progress = %v, want 1", critProgress)
	}
}

func TestExtendedAchievementSystem_OnQuestCompleted(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Complete a regular quest
	system.OnQuestCompleted(entity.ID, false, false, false)

	adventurerProgress := system.GetProgress(entity.ID, "quest_adventurer")
	if adventurerProgress != 1 {
		t.Errorf("adventurer progress = %v, want 1", adventurerProgress)
	}

	// Complete a main story quest
	system.OnQuestCompleted(entity.ID, true, false, false)

	storyProgress := system.GetProgress(entity.ID, "quest_main_story")
	if storyProgress != 1 {
		t.Errorf("main_story progress = %v, want 1", storyProgress)
	}

	// Complete a side quest
	system.OnQuestCompleted(entity.ID, false, true, false)

	sideProgress := system.GetProgress(entity.ID, "quest_side_tracker")
	if sideProgress != 1 {
		t.Errorf("side_tracker progress = %v, want 1", sideProgress)
	}

	// Complete a legendary quest
	system.OnQuestCompleted(entity.ID, false, false, true)

	legendaryProgress := system.GetProgress(entity.ID, "quest_legendary")
	if legendaryProgress != 1 {
		t.Errorf("legendary progress = %v, want 1", legendaryProgress)
	}
}

func TestExtendedAchievementSystem_OnItemCrafted(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Craft a regular item
	system.OnItemCrafted(entity.ID, false, false)

	apprenticeProgress := system.GetProgress(entity.ID, "craft_apprentice")
	if apprenticeProgress != 1 {
		t.Errorf("apprentice progress = %v, want 1", apprenticeProgress)
	}

	// Craft a rare item
	system.OnItemCrafted(entity.ID, true, false)

	qualityProgress := system.GetProgress(entity.ID, "craft_quality")
	if qualityProgress != 1 {
		t.Errorf("quality progress = %v, want 1", qualityProgress)
	}

	// Craft a perfect item
	system.OnItemCrafted(entity.ID, false, true)

	perfectProgress := system.GetProgress(entity.ID, "craft_perfect")
	if perfectProgress != 1 {
		t.Errorf("perfect progress = %v, want 1", perfectProgress)
	}
}

func TestExtendedAchievementSystem_OnExplorationEvents(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Visit an area
	system.OnAreaVisited(entity.ID)

	wandererProgress := system.GetProgress(entity.ID, "explore_wanderer")
	if wandererProgress != 1 {
		t.Errorf("wanderer progress = %v, want 1", wandererProgress)
	}

	// Clear a dungeon
	system.OnDungeonCleared(entity.ID)

	delverProgress := system.GetProgress(entity.ID, "explore_dungeon_delver")
	if delverProgress != 1 {
		t.Errorf("dungeon_delver progress = %v, want 1", delverProgress)
	}

	// Find a secret
	system.OnSecretFound(entity.ID)

	secretProgress := system.GetProgress(entity.ID, "explore_secret_finder")
	if secretProgress != 1 {
		t.Errorf("secret_finder progress = %v, want 1", secretProgress)
	}

	// Travel distance
	system.OnDistanceTraveled(entity.ID, 1000)

	distanceProgress := system.GetProgress(entity.ID, "explore_distance")
	if distanceProgress != 1000 {
		t.Errorf("distance progress = %v, want 1000", distanceProgress)
	}
}

func TestExtendedAchievementSystem_OnPvPEvents(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Win a 1v1 match
	system.OnPvPMatchWon(entity.ID, true, false)

	gladiatorProgress := system.GetProgress(entity.ID, "pvp_gladiator")
	if gladiatorProgress != 1 {
		t.Errorf("gladiator progress = %v, want 1", gladiatorProgress)
	}

	duelistProgress := system.GetProgress(entity.ID, "pvp_duelist")
	if duelistProgress != 1 {
		t.Errorf("duelist progress = %v, want 1", duelistProgress)
	}

	// Win a team match
	system.OnPvPMatchWon(entity.ID, false, true)

	teamProgress := system.GetProgress(entity.ID, "pvp_team_player")
	if teamProgress != 1 {
		t.Errorf("team_player progress = %v, want 1", teamProgress)
	}

	// Win a tournament
	system.OnTournamentWon(entity.ID)

	tournamentProgress := system.GetProgress(entity.ID, "pvp_tournament_victor")
	if tournamentProgress != 1 {
		t.Errorf("tournament_victor progress = %v, want 1", tournamentProgress)
	}

	// Earn honor
	system.OnHonorEarned(entity.ID, 500)

	honorProgress := system.GetProgress(entity.ID, "pvp_honor_bound")
	if honorProgress != 500 {
		t.Errorf("honor_bound progress = %v, want 500", honorProgress)
	}
}

func TestExtendedAchievementSystem_OnSocialEvents(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Add a friend
	system.OnFriendAdded(entity.ID)

	friendProgress := system.GetProgress(entity.ID, "social_friend_maker")
	if friendProgress != 1 {
		t.Errorf("friend_maker progress = %v, want 1", friendProgress)
	}

	// Join a guild
	system.OnGuildJoined(entity.ID)

	guildProgress := system.GetProgress(entity.ID, "social_guild_member")
	if guildProgress != 1 {
		t.Errorf("guild_member progress = %v, want 1", guildProgress)
	}

	// Complete a trade
	system.OnTradeCompleted(entity.ID)

	tradeProgress := system.GetProgress(entity.ID, "social_trade_master")
	if tradeProgress != 1 {
		t.Errorf("trade_master progress = %v, want 1", tradeProgress)
	}

	// Send a chat message
	system.OnChatMessage(entity.ID)

	chatProgress := system.GetProgress(entity.ID, "social_chat_active")
	if chatProgress != 1 {
		t.Errorf("chat_active progress = %v, want 1", chatProgress)
	}
}

func TestExtendedAchievementSystem_NonExistentEntity(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	// Try to record progress for non-existent entity
	unlocked := system.RecordProgress(99999, "combat_first_blood", 1)

	if unlocked {
		t.Error("RecordProgress should return false for non-existent entity")
	}

	// Get methods should return zero values
	progress := system.GetProgress(99999, "combat_first_blood")
	if progress != 0 {
		t.Errorf("GetProgress for non-existent entity = %v, want 0", progress)
	}

	tier := system.GetTier(99999, "combat_first_blood")
	if tier != AchievementTierNone {
		t.Errorf("GetTier for non-existent entity = %v, want None", tier)
	}

	points := system.GetTotalPoints(99999)
	if points != 0 {
		t.Errorf("GetTotalPoints for non-existent entity = %v, want 0", points)
	}
}

func TestExtendedAchievementSystem_NonExistentAchievement(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	// Try to record progress for non-existent achievement
	unlocked := system.RecordProgress(entity.ID, "non_existent_achievement", 1)

	if unlocked {
		t.Error("RecordProgress should return false for non-existent achievement")
	}
}

func TestExtendedAchievementSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	// Update should not panic with nil or empty entities
	system.Update(nil, 0.016)
	system.Update([]*Entity{}, 0.016)
}

func TestExtendedAchievementSystem_Determinism(t *testing.T) {
	// Same actions should produce same results
	results := make([]int, 2)

	for i := 0; i < 2; i++ {
		world := NewWorld()
		system := NewExtendedAchievementSystem(world)

		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: 0, Y: 0})

		// Perform the same sequence of actions
		system.OnEnemyKilled(entity.ID, false)
		system.OnEnemyKilled(entity.ID, false)
		system.OnEnemyKilled(entity.ID, true)
		system.OnDamageDealt(entity.ID, 1000, false)
		system.OnQuestCompleted(entity.ID, true, false, false)

		results[i] = system.GetTotalPoints(entity.ID)
	}

	if results[0] != results[1] {
		t.Errorf("Determinism check failed: run 1 = %d, run 2 = %d", results[0], results[1])
	}
}

func BenchmarkExtendedAchievementSystem_IncrementProgress(b *testing.B) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.IncrementProgress(entity.ID, "combat_first_blood", 1)
	}
}

func BenchmarkExtendedAchievementSystem_GetProgress(b *testing.B) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	system.RecordProgress(entity.ID, "combat_first_blood", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.GetProgress(entity.ID, "combat_first_blood")
	}
}

func BenchmarkGetAchievementDefinition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetAchievementDefinition("combat_first_blood")
	}
}

func TestExtendedAchievementSystem_MultiTierUnlock(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	unlockCount := 0
	system.SetOnUnlockCallback(func(entityID uint64, achievementID string, tier AchievementTier) {
		unlockCount++
	})

	// Jump straight to platinum (1000 kills for combat_first_blood)
	system.RecordProgress(entity.ID, "combat_first_blood", 1000)

	// Callback should be called once (for platinum, which includes all previous tiers in one call)
	if unlockCount != 1 {
		t.Errorf("unlockCount = %v, want 1", unlockCount)
	}

	tier := system.GetTier(entity.ID, "combat_first_blood")
	if tier != AchievementTierPlatinum {
		t.Errorf("Tier = %v, want Platinum", tier)
	}

	// Points should be for all tiers: 10 + 25 + 50 + 100 = 185
	points := system.GetTotalPoints(entity.ID)
	if points != 185 {
		t.Errorf("TotalPoints = %v, want 185", points)
	}
}

func TestExtendedAchievementSystem_OnSurvived(t *testing.T) {
	world := NewWorld()
	system := NewExtendedAchievementSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation

	system.OnSurvived(entity.ID)

	progress := system.GetProgress(entity.ID, "combat_survivor")
	if progress != 1 {
		t.Errorf("survivor progress = %v, want 1", progress)
	}
}

func TestExtendedAchievementSystem_NilWorld(t *testing.T) {
	system := NewExtendedAchievementSystem(nil)

	// All methods should handle nil world gracefully
	unlocked := system.RecordProgress(1, "combat_first_blood", 1)
	if unlocked {
		t.Error("RecordProgress should return false with nil world")
	}

	progress := system.GetProgress(1, "combat_first_blood")
	if progress != 0 {
		t.Errorf("GetProgress with nil world = %v, want 0", progress)
	}

	tier := system.GetTier(1, "combat_first_blood")
	if tier != AchievementTierNone {
		t.Errorf("GetTier with nil world = %v, want None", tier)
	}

	points := system.GetTotalPoints(1)
	if points != 0 {
		t.Errorf("GetTotalPoints with nil world = %v, want 0", points)
	}
}

func TestAchievementDefinitions_Thresholds(t *testing.T) {
	defs := GetAllAchievementDefinitions()

	for id, def := range defs {
		// Verify thresholds are in ascending order
		for i := 0; i < 3; i++ {
			if def.Thresholds[i] > def.Thresholds[i+1] {
				t.Errorf("Achievement %s has non-ascending thresholds: %v", id, def.Thresholds)
			}
		}

		// Verify first threshold is at least 1
		if def.Thresholds[0] < 1 {
			t.Errorf("Achievement %s has invalid first threshold: %v", id, def.Thresholds[0])
		}

		// Verify ID is not empty
		if def.ID == "" {
			t.Errorf("Achievement has empty ID")
		}

		// Verify Name is not empty
		if def.Name == "" {
			t.Errorf("Achievement %s has empty name", id)
		}
	}
}

// TestConcurrentAchievementUpdates tests thread safety of the component itself
func TestConcurrentAchievementUpdates(t *testing.T) {
	// Test component-level thread safety directly
	comp := NewExtendedAchievementComponent()
	thresholds := [4]int64{1, 10, 100, 1000}
	timestamp := int64(1000000)

	done := make(chan bool)
	iterations := 100

	// Run multiple goroutines updating the component concurrently
	for i := 0; i < 4; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				comp.IncrementProgress("combat_test", AchievementCategoryCombat, 1, thresholds, timestamp)
				comp.GetProgress("combat_test")
				comp.GetTotalPoints()
				time.Sleep(time.Microsecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify final state is consistent
	progress := comp.GetProgress("combat_test")
	if progress != int64(iterations*4) {
		t.Errorf("Final progress = %v, want %v", progress, iterations*4)
	}
}
