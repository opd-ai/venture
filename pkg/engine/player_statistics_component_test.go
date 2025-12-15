// Package engine provides tests for the PlayerStatisticsComponent.
//
// Phase 84: Player Statistics System (V15.0)
package engine

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestNewPlayerStatisticsComponent(t *testing.T) {
	comp := NewPlayerStatisticsComponent()

	if comp == nil {
		t.Fatal("NewPlayerStatisticsComponent returned nil")
	}
	if comp.Type() != "player_statistics" {
		t.Errorf("Type() = %s, want player_statistics", comp.Type())
	}
	if comp.Lifetime == nil {
		t.Error("Lifetime map should be initialized")
	}
	if comp.Session == nil {
		t.Error("Session map should be initialized")
	}
	if comp.FirstPlayTime != 0 {
		t.Errorf("FirstPlayTime = %d, want 0", comp.FirstPlayTime)
	}
	if comp.TotalPlayTime != 0 {
		t.Errorf("TotalPlayTime = %d, want 0", comp.TotalPlayTime)
	}
}

func TestPlayerStatisticsComponent_StartSession(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	timestamp := int64(1000000)

	comp.StartSession(timestamp)

	if comp.SessionStartTime != timestamp {
		t.Errorf("SessionStartTime = %d, want %d", comp.SessionStartTime, timestamp)
	}
	if comp.FirstPlayTime != timestamp {
		t.Errorf("FirstPlayTime = %d, want %d", comp.FirstPlayTime, timestamp)
	}
	if comp.GetLifetimeStat("general_sessions_played") != 1 {
		t.Errorf("general_sessions_played = %d, want 1", comp.GetLifetimeStat("general_sessions_played"))
	}

	// Start another session, FirstPlayTime should not change.
	timestamp2 := int64(2000000)
	comp.StartSession(timestamp2)

	if comp.FirstPlayTime != timestamp {
		t.Errorf("FirstPlayTime should not change, got %d, want %d", comp.FirstPlayTime, timestamp)
	}
	if comp.GetLifetimeStat("general_sessions_played") != 2 {
		t.Errorf("general_sessions_played = %d, want 2", comp.GetLifetimeStat("general_sessions_played"))
	}
}

func TestPlayerStatisticsComponent_EndSession(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	startTime := int64(1000)
	endTime := int64(2000)

	comp.StartSession(startTime)
	comp.EndSession(endTime)

	expectedDuration := endTime - startTime
	if comp.SessionPlayTime != expectedDuration {
		t.Errorf("SessionPlayTime = %d, want %d", comp.SessionPlayTime, expectedDuration)
	}
	if comp.TotalPlayTime != expectedDuration {
		t.Errorf("TotalPlayTime = %d, want %d", comp.TotalPlayTime, expectedDuration)
	}
	if comp.GetLifetimeStat("general_playtime") != expectedDuration {
		t.Errorf("general_playtime = %d, want %d", comp.GetLifetimeStat("general_playtime"), expectedDuration)
	}
}

func TestPlayerStatisticsComponent_UpdatePlaytime(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)

	comp.UpdatePlaytime(100)

	if comp.SessionPlayTime != 100 {
		t.Errorf("SessionPlayTime = %d, want 100", comp.SessionPlayTime)
	}
	if comp.TotalPlayTime != 100 {
		t.Errorf("TotalPlayTime = %d, want 100", comp.TotalPlayTime)
	}
	if comp.GetLifetimeStat("general_playtime") != 100 {
		t.Errorf("Lifetime general_playtime = %d, want 100", comp.GetLifetimeStat("general_playtime"))
	}
	if comp.GetSessionStat("general_playtime") != 100 {
		t.Errorf("Session general_playtime = %d, want 100", comp.GetSessionStat("general_playtime"))
	}

	// Update again.
	comp.UpdatePlaytime(50)

	if comp.SessionPlayTime != 150 {
		t.Errorf("SessionPlayTime = %d, want 150", comp.SessionPlayTime)
	}
	if comp.TotalPlayTime != 150 {
		t.Errorf("TotalPlayTime = %d, want 150", comp.TotalPlayTime)
	}
}

func TestPlayerStatisticsComponent_IncrementStat(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)

	comp.IncrementStat("combat_enemies_killed", 5)

	if comp.GetLifetimeStat("combat_enemies_killed") != 5 {
		t.Errorf("Lifetime combat_enemies_killed = %d, want 5", comp.GetLifetimeStat("combat_enemies_killed"))
	}
	if comp.GetSessionStat("combat_enemies_killed") != 5 {
		t.Errorf("Session combat_enemies_killed = %d, want 5", comp.GetSessionStat("combat_enemies_killed"))
	}

	// Increment again.
	comp.IncrementStat("combat_enemies_killed", 3)

	if comp.GetLifetimeStat("combat_enemies_killed") != 8 {
		t.Errorf("Lifetime combat_enemies_killed = %d, want 8", comp.GetLifetimeStat("combat_enemies_killed"))
	}
}

func TestPlayerStatisticsComponent_SetMaxStat(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)

	comp.SetMaxStat("combat_highest_hit", 100)

	if comp.GetLifetimeStat("combat_highest_hit") != 100 {
		t.Errorf("combat_highest_hit = %d, want 100", comp.GetLifetimeStat("combat_highest_hit"))
	}

	// Set lower value, should not change.
	comp.SetMaxStat("combat_highest_hit", 50)

	if comp.GetLifetimeStat("combat_highest_hit") != 100 {
		t.Errorf("combat_highest_hit = %d, want 100 (should not decrease)", comp.GetLifetimeStat("combat_highest_hit"))
	}

	// Set higher value, should change.
	comp.SetMaxStat("combat_highest_hit", 200)

	if comp.GetLifetimeStat("combat_highest_hit") != 200 {
		t.Errorf("combat_highest_hit = %d, want 200", comp.GetLifetimeStat("combat_highest_hit"))
	}
}

func TestPlayerStatisticsComponent_GetAllStats(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)
	comp.IncrementStat("combat_enemies_killed", 10)
	comp.IncrementStat("quest_completed", 5)

	lifetime := comp.GetAllLifetimeStats()
	session := comp.GetAllSessionStats()

	if lifetime["combat_enemies_killed"] != 10 {
		t.Errorf("Lifetime combat_enemies_killed = %d, want 10", lifetime["combat_enemies_killed"])
	}
	if lifetime["quest_completed"] != 5 {
		t.Errorf("Lifetime quest_completed = %d, want 5", lifetime["quest_completed"])
	}
	if session["combat_enemies_killed"] != 10 {
		t.Errorf("Session combat_enemies_killed = %d, want 10", session["combat_enemies_killed"])
	}

	// Verify returned maps are copies.
	lifetime["combat_enemies_killed"] = 999
	if comp.GetLifetimeStat("combat_enemies_killed") != 10 {
		t.Error("GetAllLifetimeStats should return a copy")
	}
}

func TestPlayerStatisticsComponent_GetStatsByCategory(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.IncrementStat("combat_enemies_killed", 10)
	comp.IncrementStat("combat_damage_dealt", 1000)
	comp.IncrementStat("quest_completed", 5)

	combatStats := comp.GetStatsByCategory(StatCategoryCombat)

	if combatStats["combat_enemies_killed"] != 10 {
		t.Errorf("combat_enemies_killed = %d, want 10", combatStats["combat_enemies_killed"])
	}
	if combatStats["combat_damage_dealt"] != 1000 {
		t.Errorf("combat_damage_dealt = %d, want 1000", combatStats["combat_damage_dealt"])
	}
	// Quest stat should not be in combat category.
	if _, exists := combatStats["quest_completed"]; exists {
		t.Error("quest_completed should not be in combat category")
	}
}

func TestPlayerStatisticsComponent_SessionReset(t *testing.T) {
	comp := NewPlayerStatisticsComponent()

	// First session.
	comp.StartSession(1000)
	comp.IncrementStat("combat_enemies_killed", 10)

	if comp.GetSessionStat("combat_enemies_killed") != 10 {
		t.Errorf("Session stat = %d, want 10", comp.GetSessionStat("combat_enemies_killed"))
	}

	// Start new session, session stats should reset.
	comp.StartSession(2000)

	if comp.GetSessionStat("combat_enemies_killed") != 0 {
		t.Errorf("Session stat after reset = %d, want 0", comp.GetSessionStat("combat_enemies_killed"))
	}
	// Lifetime should persist.
	if comp.GetLifetimeStat("combat_enemies_killed") != 10 {
		t.Errorf("Lifetime stat = %d, want 10", comp.GetLifetimeStat("combat_enemies_killed"))
	}
}

func TestPlayerStatisticsComponent_Serialize(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)
	comp.IncrementStat("combat_enemies_killed", 50)
	comp.IncrementStat("quest_completed", 10)
	comp.SetMaxStat("combat_highest_hit", 500)
	comp.UpdatePlaytime(3600)

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized data should not be empty")
	}

	// Verify it's valid JSON.
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		t.Errorf("Serialized data is not valid JSON: %v", err)
	}
}

func TestPlayerStatisticsComponent_Deserialize(t *testing.T) {
	original := NewPlayerStatisticsComponent()
	original.StartSession(1000)
	original.IncrementStat("combat_enemies_killed", 50)
	original.IncrementStat("quest_completed", 10)
	original.SetMaxStat("combat_highest_hit", 500)
	original.UpdatePlaytime(3600)

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	restored := NewPlayerStatisticsComponent()
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify stats match.
	if restored.GetLifetimeStat("combat_enemies_killed") != 50 {
		t.Errorf("combat_enemies_killed = %d, want 50", restored.GetLifetimeStat("combat_enemies_killed"))
	}
	if restored.GetLifetimeStat("quest_completed") != 10 {
		t.Errorf("quest_completed = %d, want 10", restored.GetLifetimeStat("quest_completed"))
	}
	if restored.GetLifetimeStat("combat_highest_hit") != 500 {
		t.Errorf("combat_highest_hit = %d, want 500", restored.GetLifetimeStat("combat_highest_hit"))
	}
	if restored.GetTotalPlayTime() != 3600 {
		t.Errorf("TotalPlayTime = %d, want 3600", restored.GetTotalPlayTime())
	}
	if restored.GetFirstPlayTime() != 1000 {
		t.Errorf("FirstPlayTime = %d, want 1000", restored.GetFirstPlayTime())
	}
}

func TestPlayerStatisticsComponent_DeserializeEmptyMaps(t *testing.T) {
	// Test deserialization with null maps.
	data := []byte(`{"lifetime":null,"session":null,"first_play_time":0,"total_play_time":0}`)

	comp := NewPlayerStatisticsComponent()
	if err := comp.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if comp.Lifetime == nil {
		t.Error("Lifetime map should be initialized after deserialize")
	}
	if comp.Session == nil {
		t.Error("Session map should be initialized after deserialize")
	}
}

func TestPlayerStatisticsComponent_ConcurrentAccess(t *testing.T) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent increments.
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comp.IncrementStat("combat_enemies_killed", 1)
		}()
	}

	// Concurrent reads.
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = comp.GetLifetimeStat("combat_enemies_killed")
			_ = comp.GetSessionStat("combat_enemies_killed")
			_ = comp.GetAllLifetimeStats()
		}()
	}

	wg.Wait()

	// Should have all increments.
	if comp.GetLifetimeStat("combat_enemies_killed") != int64(iterations) {
		t.Errorf("combat_enemies_killed = %d, want %d", comp.GetLifetimeStat("combat_enemies_killed"), iterations)
	}
}

func TestStatCategory_String(t *testing.T) {
	tests := []struct {
		category StatCategory
		want     string
	}{
		{StatCategoryCombat, "Combat"},
		{StatCategoryQuest, "Quest"},
		{StatCategoryCrafting, "Crafting"},
		{StatCategoryExploration, "Exploration"},
		{StatCategorySocial, "Social"},
		{StatCategoryPvP, "PvP"},
		{StatCategoryEconomy, "Economy"},
		{StatCategoryGeneral, "General"},
		{StatCategory(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("StatCategory.String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestStatDefinitions(t *testing.T) {
	// Verify we have 40+ statistics as required by Phase 84.
	count := GetStatCount()
	if count < 40 {
		t.Errorf("GetStatCount() = %d, want >= 40", count)
	}

	// Verify all definitions are valid.
	defs := GetAllStatDefinitions()
	for id, def := range defs {
		if def.ID != id {
			t.Errorf("Definition ID mismatch: key=%s, def.ID=%s", id, def.ID)
		}
		if def.Name == "" {
			t.Errorf("Definition %s has empty Name", id)
		}
		if def.Description == "" {
			t.Errorf("Definition %s has empty Description", id)
		}
	}
}

func TestGetStatDefinition(t *testing.T) {
	tests := []struct {
		statID    string
		wantExist bool
	}{
		{"combat_enemies_killed", true},
		{"quest_completed", true},
		{"craft_items_created", true},
		{"explore_areas_visited", true},
		{"social_friends_added", true},
		{"pvp_matches_won", true},
		{"economy_gold_earned", true},
		{"general_playtime", true},
		{"nonexistent_stat", false},
	}

	for _, tt := range tests {
		t.Run(tt.statID, func(t *testing.T) {
			def, exists := GetStatDefinition(tt.statID)
			if exists != tt.wantExist {
				t.Errorf("GetStatDefinition(%s) exists = %v, want %v", tt.statID, exists, tt.wantExist)
			}
			if tt.wantExist && def.ID != tt.statID {
				t.Errorf("GetStatDefinition(%s).ID = %s, want %s", tt.statID, def.ID, tt.statID)
			}
		})
	}
}

func TestGetStatDefinitionsByCategory(t *testing.T) {
	tests := []struct {
		category  StatCategory
		wantCount int // Minimum expected count
	}{
		{StatCategoryCombat, 10},
		{StatCategoryQuest, 6},
		{StatCategoryCrafting, 6},
		{StatCategoryExploration, 7},
		{StatCategorySocial, 6},
		{StatCategoryPvP, 6},
		{StatCategoryEconomy, 5},
		{StatCategoryGeneral, 6},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			defs := GetStatDefinitionsByCategory(tt.category)
			if len(defs) < tt.wantCount {
				t.Errorf("GetStatDefinitionsByCategory(%s) returned %d, want >= %d", tt.category, len(defs), tt.wantCount)
			}

			// Verify all returned definitions are in the correct category.
			for _, def := range defs {
				if def.Category != tt.category {
					t.Errorf("Definition %s has category %s, want %s", def.ID, def.Category, tt.category)
				}
			}
		})
	}
}

func TestPlayerStatisticsComponent_GettersWithoutData(t *testing.T) {
	comp := NewPlayerStatisticsComponent()

	// Test getters on empty component.
	if comp.GetLifetimeStat("nonexistent") != 0 {
		t.Error("GetLifetimeStat on nonexistent stat should return 0")
	}
	if comp.GetSessionStat("nonexistent") != 0 {
		t.Error("GetSessionStat on nonexistent stat should return 0")
	}
	if comp.GetTotalPlayTime() != 0 {
		t.Error("GetTotalPlayTime on new component should return 0")
	}
	if comp.GetSessionPlayTime() != 0 {
		t.Error("GetSessionPlayTime on new component should return 0")
	}
	if comp.GetFirstPlayTime() != 0 {
		t.Error("GetFirstPlayTime on new component should return 0")
	}
}

func BenchmarkPlayerStatisticsComponent_IncrementStat(b *testing.B) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.IncrementStat("combat_enemies_killed", 1)
	}
}

func BenchmarkPlayerStatisticsComponent_GetStat(b *testing.B) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)
	comp.IncrementStat("combat_enemies_killed", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.GetLifetimeStat("combat_enemies_killed")
	}
}

func BenchmarkPlayerStatisticsComponent_Serialize(b *testing.B) {
	comp := NewPlayerStatisticsComponent()
	comp.StartSession(1000)
	// Add some data.
	for id := range statDefinitions {
		comp.IncrementStat(id, 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = comp.Serialize()
	}
}
