package engine

import (
	"testing"
)

func TestNewWorldMemoryComponent(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	if wm.WorldSeed != 12345 {
		t.Errorf("WorldSeed = %d; want 12345", wm.WorldSeed)
	}
	if wm.CityStates == nil {
		t.Error("CityStates is nil; want initialized map")
	}
	if wm.NPCStates == nil {
		t.Error("NPCStates is nil; want initialized map")
	}
	if wm.EventHistory == nil {
		t.Error("EventHistory is nil; want initialized slice")
	}
	if wm.PlayerReputations == nil {
		t.Error("PlayerReputations is nil; want initialized map")
	}
	if wm.MaxEventHistory != 100 {
		t.Errorf("MaxEventHistory = %d; want 100", wm.MaxEventHistory)
	}
	if wm.TimeProgressionEnabled {
		t.Error("TimeProgressionEnabled = true; want false")
	}
	if wm.TimeProgressionRate != 0.1 {
		t.Errorf("TimeProgressionRate = %f; want 0.1", wm.TimeProgressionRate)
	}
}

func TestWorldMemoryComponent_Type(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)
	if wm.Type() != "world_memory" {
		t.Errorf("Type() = %s; want world_memory", wm.Type())
	}
}

func TestWorldMemoryComponent_SaveLoadCityState(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Create a city state
	city := NewCityStateComponent("city1", "Test City", 99999)
	city.Prosperity = 0.75
	city.Population = 150
	city.State = CityStateThriving

	// Save city state
	wm.SaveCityState(city)

	// Verify it's saved
	if wm.GetCityCount() != 1 {
		t.Errorf("GetCityCount() = %d; want 1", wm.GetCityCount())
	}

	// Load city state
	loaded := wm.LoadCityState("city1")
	if loaded == nil {
		t.Fatal("LoadCityState returned nil; want valid city")
	}

	if loaded.CityID != "city1" {
		t.Errorf("CityID = %s; want city1", loaded.CityID)
	}
	if loaded.CityName != "Test City" {
		t.Errorf("CityName = %s; want Test City", loaded.CityName)
	}
	if loaded.Prosperity != 0.75 {
		t.Errorf("Prosperity = %f; want 0.75", loaded.Prosperity)
	}
	if loaded.Population != 150 {
		t.Errorf("Population = %d; want 150", loaded.Population)
	}
	if loaded.State != CityStateThriving {
		t.Errorf("State = %s; want thriving", loaded.State)
	}

	// Test loading non-existent city
	missing := wm.LoadCityState("nonexistent")
	if missing != nil {
		t.Error("LoadCityState for nonexistent city should return nil")
	}
}

func TestWorldMemoryComponent_SaveLoadNPCState(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Create a schedule
	schedule := NewScheduleComponent(100.0, 200.0)
	schedule.AddActivity(ActivityWork, 8, 17, 150.0, 100.0, "Shop")
	schedule.CurrentActivityIdx = 0
	schedule.IsMoving = true

	// Save NPC state
	wm.SaveNPCState("npc1", "Merchant Bob", 120.0, 150.0, schedule)

	// Verify it's saved
	if wm.GetNPCCount() != 1 {
		t.Errorf("GetNPCCount() = %d; want 1", wm.GetNPCCount())
	}

	// Load NPC state
	loaded := wm.LoadNPCState("npc1")
	if loaded == nil {
		t.Fatal("LoadNPCState returned nil; want valid NPC")
	}

	if loaded.EntityID != "npc1" {
		t.Errorf("EntityID = %s; want npc1", loaded.EntityID)
	}
	if loaded.Name != "Merchant Bob" {
		t.Errorf("Name = %s; want Merchant Bob", loaded.Name)
	}
	if loaded.X != 120.0 {
		t.Errorf("X = %f; want 120.0", loaded.X)
	}
	if loaded.Y != 150.0 {
		t.Errorf("Y = %f; want 150.0", loaded.Y)
	}
	if loaded.HomeX != 100.0 {
		t.Errorf("HomeX = %f; want 100.0", loaded.HomeX)
	}
	if loaded.HomeY != 200.0 {
		t.Errorf("HomeY = %f; want 200.0", loaded.HomeY)
	}
	if loaded.IsMoving != true {
		t.Error("IsMoving = false; want true")
	}
	if len(loaded.Schedule) != 1 {
		t.Errorf("Schedule length = %d; want 1", len(loaded.Schedule))
	}

	// Test loading non-existent NPC
	missing := wm.LoadNPCState("nonexistent")
	if missing != nil {
		t.Error("LoadNPCState for nonexistent NPC should return nil")
	}
}

func TestWorldMemoryComponent_RecordEvent(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)
	wm.MaxEventHistory = 3

	// Record events
	events := []WorldEventRecord{
		{EventID: "e1", EventType: "trade", Description: "Trade completed", GameTime: 100.0, Magnitude: 0.5},
		{EventID: "e2", EventType: "raid", Description: "Raid occurred", GameTime: 200.0, AffectedCityID: "city1", Magnitude: 0.8},
		{EventID: "e3", EventType: "trade", Description: "Another trade", GameTime: 300.0, Magnitude: 0.3},
		{EventID: "e4", EventType: "build", Description: "Building built", GameTime: 400.0, AffectedCityID: "city1", Magnitude: 0.6},
	}

	for _, e := range events {
		wm.RecordEvent(e)
	}

	// Should only keep last 3 events
	if wm.GetEventCount() != 3 {
		t.Errorf("GetEventCount() = %d; want 3", wm.GetEventCount())
	}

	// Check oldest event was removed
	recent := wm.GetRecentEvents(10)
	if len(recent) != 3 {
		t.Errorf("GetRecentEvents(10) = %d; want 3", len(recent))
	}
	if recent[0].EventID != "e2" {
		t.Errorf("First event ID = %s; want e2", recent[0].EventID)
	}
}

func TestWorldMemoryComponent_GetEventsByCity(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	events := []WorldEventRecord{
		{EventID: "e1", EventType: "trade", AffectedCityID: "city1"},
		{EventID: "e2", EventType: "raid", AffectedCityID: "city2"},
		{EventID: "e3", EventType: "build", AffectedCityID: "city1"},
	}
	for _, e := range events {
		wm.RecordEvent(e)
	}

	city1Events := wm.GetEventsByCity("city1")
	if len(city1Events) != 2 {
		t.Errorf("GetEventsByCity(city1) = %d; want 2", len(city1Events))
	}

	city2Events := wm.GetEventsByCity("city2")
	if len(city2Events) != 1 {
		t.Errorf("GetEventsByCity(city2) = %d; want 1", len(city2Events))
	}
}

func TestWorldMemoryComponent_GetEventsByType(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	events := []WorldEventRecord{
		{EventID: "e1", EventType: "trade"},
		{EventID: "e2", EventType: "raid"},
		{EventID: "e3", EventType: "trade"},
	}
	for _, e := range events {
		wm.RecordEvent(e)
	}

	tradeEvents := wm.GetEventsByType("trade")
	if len(tradeEvents) != 2 {
		t.Errorf("GetEventsByType(trade) = %d; want 2", len(tradeEvents))
	}
}

func TestWorldMemoryComponent_PlayerCityReputation(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Set reputation
	wm.SetPlayerCityReputation("player1", "city1", 50.0)
	wm.SetPlayerCityReputation("player1", "city2", -30.0)

	// Get reputation
	rep1 := wm.GetPlayerCityReputation("player1", "city1")
	if rep1 != 50.0 {
		t.Errorf("Reputation city1 = %f; want 50.0", rep1)
	}

	rep2 := wm.GetPlayerCityReputation("player1", "city2")
	if rep2 != -30.0 {
		t.Errorf("Reputation city2 = %f; want -30.0", rep2)
	}

	// Get unknown reputation (should be 0)
	repUnknown := wm.GetPlayerCityReputation("player1", "unknown")
	if repUnknown != 0.0 {
		t.Errorf("Reputation unknown = %f; want 0.0", repUnknown)
	}

	// Adjust reputation
	wm.AdjustPlayerCityReputation("player1", "city1", 25.0)
	repAdjusted := wm.GetPlayerCityReputation("player1", "city1")
	if repAdjusted != 75.0 {
		t.Errorf("Adjusted reputation = %f; want 75.0", repAdjusted)
	}

	// Test clamping
	wm.SetPlayerCityReputation("player1", "city3", 150.0)
	repClamped := wm.GetPlayerCityReputation("player1", "city3")
	if repClamped != 100.0 {
		t.Errorf("Clamped max reputation = %f; want 100.0", repClamped)
	}

	wm.SetPlayerCityReputation("player1", "city4", -150.0)
	repClampedMin := wm.GetPlayerCityReputation("player1", "city4")
	if repClampedMin != -100.0 {
		t.Errorf("Clamped min reputation = %f; want -100.0", repClampedMin)
	}
}

func TestWorldMemoryComponent_GetPlayerAllCityReputations(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	wm.SetPlayerCityReputation("player1", "city1", 50.0)
	wm.SetPlayerCityReputation("player1", "city2", -30.0)

	reps := wm.GetPlayerAllCityReputations("player1")
	if len(reps) != 2 {
		t.Errorf("GetPlayerAllCityReputations = %d; want 2", len(reps))
	}
	if reps["city1"] != 50.0 {
		t.Errorf("city1 reputation = %f; want 50.0", reps["city1"])
	}
	if reps["city2"] != -30.0 {
		t.Errorf("city2 reputation = %f; want -30.0", reps["city2"])
	}

	// Unknown player should return empty map
	unknown := wm.GetPlayerAllCityReputations("unknown")
	if len(unknown) != 0 {
		t.Errorf("GetPlayerAllCityReputations(unknown) = %d; want 0", len(unknown))
	}
}

func TestWorldMemoryComponent_GetReputationTier(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	tests := []struct {
		reputation float64
		wantTier   string
	}{
		{75.0, "Revered"},
		{100.0, "Revered"},
		{50.0, "Honored"},
		{74.9, "Honored"},
		{25.0, "Friendly"},
		{49.9, "Friendly"},
		{0.0, "Neutral"},
		{24.9, "Neutral"},
		{-24.9, "Neutral"},
		{-25.0, "Unfriendly"},
		{-49.9, "Unfriendly"},
		{-50.0, "Hostile"},
		{-74.9, "Hostile"},
		{-75.0, "Hated"},
		{-100.0, "Hated"},
	}

	for _, tt := range tests {
		t.Run(tt.wantTier, func(t *testing.T) {
			got := wm.GetReputationTier(tt.reputation)
			if got != tt.wantTier {
				t.Errorf("GetReputationTier(%f) = %s; want %s", tt.reputation, got, tt.wantTier)
			}
		})
	}
}

func TestWorldMemoryComponent_TimeProgression(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Disabled by default
	progression := wm.CalculateTimeProgression(1000.0, 100.0)
	if progression != 0.0 {
		t.Errorf("Disabled progression = %f; want 0.0", progression)
	}

	// Enable and test
	wm.TimeProgressionEnabled = true
	wm.TimeProgressionRate = 0.5

	progression = wm.CalculateTimeProgression(1000.0, 100.0)
	if progression != 50.0 {
		t.Errorf("Enabled progression = %f; want 50.0", progression)
	}
}

func TestWorldMemoryComponent_SerializeDeserialize(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Add some data
	city := NewCityStateComponent("city1", "Test City", 99999)
	city.Prosperity = 0.75
	wm.SaveCityState(city)

	schedule := NewScheduleComponent(100.0, 200.0)
	wm.SaveNPCState("npc1", "Bob", 50.0, 60.0, schedule)

	wm.RecordEvent(WorldEventRecord{
		EventID:   "e1",
		EventType: "trade",
		GameTime:  100.0,
	})

	wm.SetPlayerCityReputation("player1", "city1", 50.0)
	wm.TimeProgressionEnabled = true
	wm.LastSaveTime = 500.0

	// Serialize
	data, err := wm.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Deserialize into new component
	wm2 := &WorldMemoryComponent{}
	err = wm2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify all data
	if wm2.WorldSeed != 12345 {
		t.Errorf("WorldSeed = %d; want 12345", wm2.WorldSeed)
	}
	if wm2.LastSaveTime != 500.0 {
		t.Errorf("LastSaveTime = %f; want 500.0", wm2.LastSaveTime)
	}
	if wm2.GetCityCount() != 1 {
		t.Errorf("GetCityCount() = %d; want 1", wm2.GetCityCount())
	}
	if wm2.GetNPCCount() != 1 {
		t.Errorf("GetNPCCount() = %d; want 1", wm2.GetNPCCount())
	}
	if wm2.GetEventCount() != 1 {
		t.Errorf("GetEventCount() = %d; want 1", wm2.GetEventCount())
	}
	if !wm2.TimeProgressionEnabled {
		t.Error("TimeProgressionEnabled = false; want true")
	}

	rep := wm2.GetPlayerCityReputation("player1", "city1")
	if rep != 50.0 {
		t.Errorf("Player reputation = %f; want 50.0", rep)
	}

	loaded := wm2.LoadCityState("city1")
	if loaded == nil || loaded.Prosperity != 0.75 {
		t.Error("Loaded city state has wrong prosperity")
	}
}

func TestWorldMemoryComponent_NilHandling(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Save nil city should not panic
	wm.SaveCityState(nil)
	if wm.GetCityCount() != 0 {
		t.Error("Saving nil city should not add entry")
	}

	// Save NPC with nil schedule should not panic
	wm.SaveNPCState("npc1", "Test", 0, 0, nil)
	if wm.GetNPCCount() != 0 {
		t.Error("Saving NPC with nil schedule should not add entry")
	}
}

func TestWorldMemoryComponent_GetRecentEvents(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	// Empty case
	empty := wm.GetRecentEvents(5)
	if len(empty) != 0 {
		t.Errorf("GetRecentEvents on empty = %d; want 0", len(empty))
	}

	// Invalid count
	invalid := wm.GetRecentEvents(-1)
	if len(invalid) != 0 {
		t.Errorf("GetRecentEvents(-1) = %d; want 0", len(invalid))
	}

	// Add some events
	for i := 0; i < 5; i++ {
		wm.RecordEvent(WorldEventRecord{EventID: string(rune('a' + i))})
	}

	// Get more than available
	all := wm.GetRecentEvents(10)
	if len(all) != 5 {
		t.Errorf("GetRecentEvents(10) = %d; want 5", len(all))
	}

	// Get exact count
	three := wm.GetRecentEvents(3)
	if len(three) != 3 {
		t.Errorf("GetRecentEvents(3) = %d; want 3", len(three))
	}
}

func TestWorldMemoryComponent_ClearHistory(t *testing.T) {
	wm := NewWorldMemoryComponent(12345)

	wm.RecordEvent(WorldEventRecord{EventID: "e1"})
	wm.RecordEvent(WorldEventRecord{EventID: "e2"})

	if wm.GetEventCount() != 2 {
		t.Errorf("GetEventCount() = %d; want 2", wm.GetEventCount())
	}

	wm.ClearHistory()

	if wm.GetEventCount() != 0 {
		t.Errorf("After clear, GetEventCount() = %d; want 0", wm.GetEventCount())
	}
}
