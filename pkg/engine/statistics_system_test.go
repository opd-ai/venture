// Package engine provides tests for the StatisticsSystem.
//
// Phase 84: Player Statistics System (V15.0)
package engine

import (
	"testing"
)

func createTestWorldForStats() *World {
	world := NewWorld()
	// Set up a clock with a fixed time for deterministic tests.
	clock := NewSimulationClock(1000)
	world.Clock = clock
	return world
}

func TestNewStatisticsSystem(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	if sys == nil {
		t.Fatal("NewStatisticsSystem returned nil")
	}
	if sys.world != world {
		t.Error("System world reference incorrect")
	}
}

func TestNewStatisticsSystem_NilWorld(t *testing.T) {
	sys := NewStatisticsSystem(nil)

	if sys == nil {
		t.Fatal("NewStatisticsSystem returned nil for nil world")
	}
	// Should not panic when called with nil world.
	sys.StartSession(1, 1000)
	sys.IncrementStat(1, "combat_enemies_killed", 1)
}

func TestStatisticsSystem_StartSession(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.StartSession(entityID, 1000)

	// Should create component and start session.
	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		t.Fatal("Component should be created")
	}

	comp, ok := compRaw.(*PlayerStatisticsComponent)
	if !ok {
		t.Fatal("Component should be PlayerStatisticsComponent")
	}

	if comp.SessionStartTime != 1000 {
		t.Errorf("SessionStartTime = %d, want 1000", comp.SessionStartTime)
	}
	if comp.GetFirstPlayTime() != 1000 {
		t.Errorf("FirstPlayTime = %d, want 1000", comp.GetFirstPlayTime())
	}
}

func TestStatisticsSystem_EndSession(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.StartSession(entityID, 1000)
	sys.EndSession(entityID, 2000)

	sessionTime := sys.GetSessionPlayTime(entityID)
	if sessionTime != 1000 {
		t.Errorf("SessionPlayTime = %d, want 1000", sessionTime)
	}

	totalTime := sys.GetTotalPlayTime(entityID)
	if totalTime != 1000 {
		t.Errorf("TotalPlayTime = %d, want 1000", totalTime)
	}
}

func TestStatisticsSystem_IncrementStat(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.IncrementStat(entityID, "combat_enemies_killed", 5)

	if sys.GetLifetimeStat(entityID, "combat_enemies_killed") != 5 {
		t.Errorf("Lifetime stat = %d, want 5", sys.GetLifetimeStat(entityID, "combat_enemies_killed"))
	}
	if sys.GetSessionStat(entityID, "combat_enemies_killed") != 5 {
		t.Errorf("Session stat = %d, want 5", sys.GetSessionStat(entityID, "combat_enemies_killed"))
	}
}

func TestStatisticsSystem_SetMaxStat(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.SetMaxStat(entityID, "combat_highest_hit", 100)
	if sys.GetLifetimeStat(entityID, "combat_highest_hit") != 100 {
		t.Errorf("Max stat = %d, want 100", sys.GetLifetimeStat(entityID, "combat_highest_hit"))
	}

	// Lower value should not update.
	sys.SetMaxStat(entityID, "combat_highest_hit", 50)
	if sys.GetLifetimeStat(entityID, "combat_highest_hit") != 100 {
		t.Errorf("Max stat = %d, want 100 (should not decrease)", sys.GetLifetimeStat(entityID, "combat_highest_hit"))
	}

	// Higher value should update.
	sys.SetMaxStat(entityID, "combat_highest_hit", 200)
	if sys.GetLifetimeStat(entityID, "combat_highest_hit") != 200 {
		t.Errorf("Max stat = %d, want 200", sys.GetLifetimeStat(entityID, "combat_highest_hit"))
	}
}

func TestStatisticsSystem_RecordStat_UnknownStat(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	// Should not panic with unknown stat.
	sys.RecordStat(entityID, "unknown_stat_id", 10)

	// Should still be 0 since stat doesn't exist.
	if sys.GetLifetimeStat(entityID, "unknown_stat_id") != 0 {
		t.Error("Unknown stat should remain 0")
	}
}

func TestStatisticsSystem_OnEnemyKilled(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnEnemyKilled(entityID, false, 100)
	if sys.GetLifetimeStat(entityID, "combat_enemies_killed") != 1 {
		t.Errorf("combat_enemies_killed = %d, want 1", sys.GetLifetimeStat(entityID, "combat_enemies_killed"))
	}
	if sys.GetLifetimeStat(entityID, "combat_bosses_killed") != 0 {
		t.Errorf("combat_bosses_killed = %d, want 0", sys.GetLifetimeStat(entityID, "combat_bosses_killed"))
	}

	sys.OnEnemyKilled(entityID, true, 500)
	if sys.GetLifetimeStat(entityID, "combat_enemies_killed") != 2 {
		t.Errorf("combat_enemies_killed = %d, want 2", sys.GetLifetimeStat(entityID, "combat_enemies_killed"))
	}
	if sys.GetLifetimeStat(entityID, "combat_bosses_killed") != 1 {
		t.Errorf("combat_bosses_killed = %d, want 1", sys.GetLifetimeStat(entityID, "combat_bosses_killed"))
	}
}

func TestStatisticsSystem_OnDamageDealt(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnDamageDealt(entityID, 100, false)
	if sys.GetLifetimeStat(entityID, "combat_damage_dealt") != 100 {
		t.Errorf("combat_damage_dealt = %d, want 100", sys.GetLifetimeStat(entityID, "combat_damage_dealt"))
	}
	if sys.GetLifetimeStat(entityID, "combat_critical_hits") != 0 {
		t.Errorf("combat_critical_hits = %d, want 0", sys.GetLifetimeStat(entityID, "combat_critical_hits"))
	}
	if sys.GetLifetimeStat(entityID, "combat_highest_hit") != 100 {
		t.Errorf("combat_highest_hit = %d, want 100", sys.GetLifetimeStat(entityID, "combat_highest_hit"))
	}

	sys.OnDamageDealt(entityID, 200, true)
	if sys.GetLifetimeStat(entityID, "combat_damage_dealt") != 300 {
		t.Errorf("combat_damage_dealt = %d, want 300", sys.GetLifetimeStat(entityID, "combat_damage_dealt"))
	}
	if sys.GetLifetimeStat(entityID, "combat_critical_hits") != 1 {
		t.Errorf("combat_critical_hits = %d, want 1", sys.GetLifetimeStat(entityID, "combat_critical_hits"))
	}
	if sys.GetLifetimeStat(entityID, "combat_highest_hit") != 200 {
		t.Errorf("combat_highest_hit = %d, want 200", sys.GetLifetimeStat(entityID, "combat_highest_hit"))
	}
}

func TestStatisticsSystem_OnQuestCompleted(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	// Main story quest.
	sys.OnQuestCompleted(entityID, true, false, false, false)
	if sys.GetLifetimeStat(entityID, "quest_completed") != 1 {
		t.Errorf("quest_completed = %d, want 1", sys.GetLifetimeStat(entityID, "quest_completed"))
	}
	if sys.GetLifetimeStat(entityID, "quest_main_completed") != 1 {
		t.Errorf("quest_main_completed = %d, want 1", sys.GetLifetimeStat(entityID, "quest_main_completed"))
	}

	// Side quest.
	sys.OnQuestCompleted(entityID, false, true, false, false)
	if sys.GetLifetimeStat(entityID, "quest_completed") != 2 {
		t.Errorf("quest_completed = %d, want 2", sys.GetLifetimeStat(entityID, "quest_completed"))
	}
	if sys.GetLifetimeStat(entityID, "quest_side_completed") != 1 {
		t.Errorf("quest_side_completed = %d, want 1", sys.GetLifetimeStat(entityID, "quest_side_completed"))
	}

	// Daily quest.
	sys.OnQuestCompleted(entityID, false, false, true, false)
	if sys.GetLifetimeStat(entityID, "quest_daily_completed") != 1 {
		t.Errorf("quest_daily_completed = %d, want 1", sys.GetLifetimeStat(entityID, "quest_daily_completed"))
	}

	// Event quest.
	sys.OnQuestCompleted(entityID, false, false, false, true)
	if sys.GetLifetimeStat(entityID, "quest_event_completed") != 1 {
		t.Errorf("quest_event_completed = %d, want 1", sys.GetLifetimeStat(entityID, "quest_event_completed"))
	}
}

func TestStatisticsSystem_OnItemCrafted(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnItemCrafted(entityID, false, false, false, false)
	if sys.GetLifetimeStat(entityID, "craft_items_created") != 1 {
		t.Errorf("craft_items_created = %d, want 1", sys.GetLifetimeStat(entityID, "craft_items_created"))
	}

	sys.OnItemCrafted(entityID, true, true, true, true)
	if sys.GetLifetimeStat(entityID, "craft_items_created") != 2 {
		t.Errorf("craft_items_created = %d, want 2", sys.GetLifetimeStat(entityID, "craft_items_created"))
	}
	if sys.GetLifetimeStat(entityID, "craft_rare_items") != 1 {
		t.Errorf("craft_rare_items = %d, want 1", sys.GetLifetimeStat(entityID, "craft_rare_items"))
	}
	if sys.GetLifetimeStat(entityID, "craft_perfect_items") != 1 {
		t.Errorf("craft_perfect_items = %d, want 1", sys.GetLifetimeStat(entityID, "craft_perfect_items"))
	}
	if sys.GetLifetimeStat(entityID, "craft_potions_brewed") != 1 {
		t.Errorf("craft_potions_brewed = %d, want 1", sys.GetLifetimeStat(entityID, "craft_potions_brewed"))
	}
	if sys.GetLifetimeStat(entityID, "craft_enchantments_applied") != 1 {
		t.Errorf("craft_enchantments_applied = %d, want 1", sys.GetLifetimeStat(entityID, "craft_enchantments_applied"))
	}
}

func TestStatisticsSystem_PvPEvents(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnPvPMatchPlayed(entityID, true)
	if sys.GetLifetimeStat(entityID, "pvp_matches_played") != 1 {
		t.Errorf("pvp_matches_played = %d, want 1", sys.GetLifetimeStat(entityID, "pvp_matches_played"))
	}
	if sys.GetLifetimeStat(entityID, "pvp_matches_won") != 1 {
		t.Errorf("pvp_matches_won = %d, want 1", sys.GetLifetimeStat(entityID, "pvp_matches_won"))
	}

	sys.OnPvPMatchPlayed(entityID, false)
	if sys.GetLifetimeStat(entityID, "pvp_matches_played") != 2 {
		t.Errorf("pvp_matches_played = %d, want 2", sys.GetLifetimeStat(entityID, "pvp_matches_played"))
	}
	if sys.GetLifetimeStat(entityID, "pvp_matches_won") != 1 {
		t.Errorf("pvp_matches_won should not increment on loss")
	}

	sys.OnTournamentEntered(entityID)
	if sys.GetLifetimeStat(entityID, "pvp_tournaments_entered") != 1 {
		t.Errorf("pvp_tournaments_entered = %d, want 1", sys.GetLifetimeStat(entityID, "pvp_tournaments_entered"))
	}

	sys.OnTournamentWon(entityID)
	if sys.GetLifetimeStat(entityID, "pvp_tournaments_won") != 1 {
		t.Errorf("pvp_tournaments_won = %d, want 1", sys.GetLifetimeStat(entityID, "pvp_tournaments_won"))
	}

	sys.OnHonorEarned(entityID, 100)
	if sys.GetLifetimeStat(entityID, "pvp_honor_earned") != 100 {
		t.Errorf("pvp_honor_earned = %d, want 100", sys.GetLifetimeStat(entityID, "pvp_honor_earned"))
	}

	sys.OnRatingChanged(entityID, 1500)
	if sys.GetLifetimeStat(entityID, "pvp_highest_rating") != 1500 {
		t.Errorf("pvp_highest_rating = %d, want 1500", sys.GetLifetimeStat(entityID, "pvp_highest_rating"))
	}
}

func TestStatisticsSystem_EconomyEvents(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnGoldEarned(entityID, 1000)
	if sys.GetLifetimeStat(entityID, "economy_gold_earned") != 1000 {
		t.Errorf("economy_gold_earned = %d, want 1000", sys.GetLifetimeStat(entityID, "economy_gold_earned"))
	}

	sys.OnGoldSpent(entityID, 500)
	if sys.GetLifetimeStat(entityID, "economy_gold_spent") != 500 {
		t.Errorf("economy_gold_spent = %d, want 500", sys.GetLifetimeStat(entityID, "economy_gold_spent"))
	}

	sys.OnItemSold(entityID)
	if sys.GetLifetimeStat(entityID, "economy_items_sold") != 1 {
		t.Errorf("economy_items_sold = %d, want 1", sys.GetLifetimeStat(entityID, "economy_items_sold"))
	}

	sys.OnItemBought(entityID)
	if sys.GetLifetimeStat(entityID, "economy_items_bought") != 1 {
		t.Errorf("economy_items_bought = %d, want 1", sys.GetLifetimeStat(entityID, "economy_items_bought"))
	}

	sys.OnGoldHeld(entityID, 10000)
	if sys.GetLifetimeStat(entityID, "economy_highest_gold") != 10000 {
		t.Errorf("economy_highest_gold = %d, want 10000", sys.GetLifetimeStat(entityID, "economy_highest_gold"))
	}
}

func TestStatisticsSystem_ExplorationEvents(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnAreaVisited(entityID)
	if sys.GetLifetimeStat(entityID, "explore_areas_visited") != 1 {
		t.Errorf("explore_areas_visited = %d, want 1", sys.GetLifetimeStat(entityID, "explore_areas_visited"))
	}

	sys.OnDungeonCleared(entityID)
	if sys.GetLifetimeStat(entityID, "explore_dungeons_cleared") != 1 {
		t.Errorf("explore_dungeons_cleared = %d, want 1", sys.GetLifetimeStat(entityID, "explore_dungeons_cleared"))
	}

	sys.OnSecretFound(entityID)
	if sys.GetLifetimeStat(entityID, "explore_secrets_found") != 1 {
		t.Errorf("explore_secrets_found = %d, want 1", sys.GetLifetimeStat(entityID, "explore_secrets_found"))
	}

	sys.OnChestOpened(entityID)
	if sys.GetLifetimeStat(entityID, "explore_chests_opened") != 1 {
		t.Errorf("explore_chests_opened = %d, want 1", sys.GetLifetimeStat(entityID, "explore_chests_opened"))
	}

	sys.OnDistanceTraveled(entityID, 1000)
	if sys.GetLifetimeStat(entityID, "explore_distance_traveled") != 1000 {
		t.Errorf("explore_distance_traveled = %d, want 1000", sys.GetLifetimeStat(entityID, "explore_distance_traveled"))
	}

	sys.OnBiomeVisited(entityID)
	if sys.GetLifetimeStat(entityID, "explore_biomes_visited") != 1 {
		t.Errorf("explore_biomes_visited = %d, want 1", sys.GetLifetimeStat(entityID, "explore_biomes_visited"))
	}

	sys.OnMapRevealed(entityID, 50)
	if sys.GetLifetimeStat(entityID, "explore_map_revealed") != 50 {
		t.Errorf("explore_map_revealed = %d, want 50", sys.GetLifetimeStat(entityID, "explore_map_revealed"))
	}
}

func TestStatisticsSystem_SocialEvents(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnFriendAdded(entityID)
	if sys.GetLifetimeStat(entityID, "social_friends_added") != 1 {
		t.Errorf("social_friends_added = %d, want 1", sys.GetLifetimeStat(entityID, "social_friends_added"))
	}

	sys.OnGuildJoined(entityID)
	if sys.GetLifetimeStat(entityID, "social_guilds_joined") != 1 {
		t.Errorf("social_guilds_joined = %d, want 1", sys.GetLifetimeStat(entityID, "social_guilds_joined"))
	}

	sys.OnTradeCompleted(entityID)
	if sys.GetLifetimeStat(entityID, "social_trades_completed") != 1 {
		t.Errorf("social_trades_completed = %d, want 1", sys.GetLifetimeStat(entityID, "social_trades_completed"))
	}

	sys.OnChatMessage(entityID)
	if sys.GetLifetimeStat(entityID, "social_chat_messages") != 1 {
		t.Errorf("social_chat_messages = %d, want 1", sys.GetLifetimeStat(entityID, "social_chat_messages"))
	}

	sys.OnPartyActivity(entityID)
	if sys.GetLifetimeStat(entityID, "social_party_activities") != 1 {
		t.Errorf("social_party_activities = %d, want 1", sys.GetLifetimeStat(entityID, "social_party_activities"))
	}

	sys.OnEmoteUsed(entityID)
	if sys.GetLifetimeStat(entityID, "social_emotes_used") != 1 {
		t.Errorf("social_emotes_used = %d, want 1", sys.GetLifetimeStat(entityID, "social_emotes_used"))
	}
}

func TestStatisticsSystem_CombatEvents(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnDamageTaken(entityID, 50)
	if sys.GetLifetimeStat(entityID, "combat_damage_taken") != 50 {
		t.Errorf("combat_damage_taken = %d, want 50", sys.GetLifetimeStat(entityID, "combat_damage_taken"))
	}

	sys.OnDodge(entityID)
	if sys.GetLifetimeStat(entityID, "combat_dodges") != 1 {
		t.Errorf("combat_dodges = %d, want 1", sys.GetLifetimeStat(entityID, "combat_dodges"))
	}

	sys.OnDeath(entityID)
	if sys.GetLifetimeStat(entityID, "combat_deaths") != 1 {
		t.Errorf("combat_deaths = %d, want 1", sys.GetLifetimeStat(entityID, "combat_deaths"))
	}

	sys.OnHealing(entityID, 100)
	if sys.GetLifetimeStat(entityID, "combat_healing_done") != 100 {
		t.Errorf("combat_healing_done = %d, want 100", sys.GetLifetimeStat(entityID, "combat_healing_done"))
	}

	sys.OnSpellCast(entityID)
	if sys.GetLifetimeStat(entityID, "combat_spells_cast") != 1 {
		t.Errorf("combat_spells_cast = %d, want 1", sys.GetLifetimeStat(entityID, "combat_spells_cast"))
	}
}

func TestStatisticsSystem_GeneralEvents(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnCombatTime(entityID, 60)
	if sys.GetLifetimeStat(entityID, "general_combat_time") != 60 {
		t.Errorf("general_combat_time = %d, want 60", sys.GetLifetimeStat(entityID, "general_combat_time"))
	}

	sys.OnCraftingTime(entityID, 30)
	if sys.GetLifetimeStat(entityID, "general_crafting_time") != 30 {
		t.Errorf("general_crafting_time = %d, want 30", sys.GetLifetimeStat(entityID, "general_crafting_time"))
	}

	sys.OnLevelUp(entityID, 10)
	if sys.GetLifetimeStat(entityID, "general_level_reached") != 10 {
		t.Errorf("general_level_reached = %d, want 10", sys.GetLifetimeStat(entityID, "general_level_reached"))
	}

	sys.OnAchievementUnlocked(entityID)
	if sys.GetLifetimeStat(entityID, "general_achievements_unlocked") != 1 {
		t.Errorf("general_achievements_unlocked = %d, want 1", sys.GetLifetimeStat(entityID, "general_achievements_unlocked"))
	}
}

func TestStatisticsSystem_Update(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(NewPlayerStatisticsComponent())
	world.Update(0) // Process entity creation

	// Start session.
	compRaw, _ := entity.GetComponent("player_statistics")
	comp := compRaw.(*PlayerStatisticsComponent)
	comp.StartSession(1000)

	// Update with 1 second delta time.
	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Playtime should be updated.
	if comp.GetSessionPlayTime() != 1 {
		t.Errorf("SessionPlayTime = %d, want 1", comp.GetSessionPlayTime())
	}
}

func TestStatisticsSystem_GetStatsNonexistentEntity(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	// Should return 0 for nonexistent entity.
	if sys.GetLifetimeStat(999, "combat_enemies_killed") != 0 {
		t.Error("GetLifetimeStat on nonexistent entity should return 0")
	}
	if sys.GetSessionStat(999, "combat_enemies_killed") != 0 {
		t.Error("GetSessionStat on nonexistent entity should return 0")
	}
	if sys.GetTotalPlayTime(999) != 0 {
		t.Error("GetTotalPlayTime on nonexistent entity should return 0")
	}
	if sys.GetSessionPlayTime(999) != 0 {
		t.Error("GetSessionPlayTime on nonexistent entity should return 0")
	}
}

func TestStatisticsSystem_RecipeLearned(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnRecipeLearned(entityID)
	if sys.GetLifetimeStat(entityID, "craft_recipes_learned") != 1 {
		t.Errorf("craft_recipes_learned = %d, want 1", sys.GetLifetimeStat(entityID, "craft_recipes_learned"))
	}
}

func TestStatisticsSystem_QuestFailed(t *testing.T) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0) // Process entity creation
	entityID := entity.ID

	sys.OnQuestFailed(entityID)
	if sys.GetLifetimeStat(entityID, "quest_failed") != 1 {
		t.Errorf("quest_failed = %d, want 1", sys.GetLifetimeStat(entityID, "quest_failed"))
	}
}

func BenchmarkStatisticsSystem_IncrementStat(b *testing.B) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0)
	entityID := entity.ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.IncrementStat(entityID, "combat_enemies_killed", 1)
	}
}

func BenchmarkStatisticsSystem_GetStat(b *testing.B) {
	world := createTestWorldForStats()
	sys := NewStatisticsSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0)
	entityID := entity.ID
	sys.IncrementStat(entityID, "combat_enemies_killed", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.GetLifetimeStat(entityID, "combat_enemies_killed")
	}
}
