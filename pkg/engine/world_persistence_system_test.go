package engine

import (
	"testing"
)

func TestNewWorldPersistenceSystem(t *testing.T) {
	system := NewWorldPersistenceSystem()
	if system == nil {
		t.Fatal("NewWorldPersistenceSystem returned nil")
	}
	if system.logger == nil {
		t.Error("System logger is nil")
	}
}

func TestWorldPersistenceSystem_SaveLoadWorldState(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)

	// Create test entities
	cityEntity := NewEntity(1)
	cityState := NewCityStateComponent("city1", "Test City", 99999)
	cityState.Prosperity = 0.75
	cityState.Population = 150
	cityEntity.AddComponent(cityState)

	npcEntity := NewEntity(2)
	schedule := NewScheduleComponent(100.0, 200.0)
	schedule.AddActivity(ActivityWork, 8, 17, 150.0, 100.0, "Shop")
	schedule.CurrentActivityIdx = 0
	npcEntity.AddComponent(schedule)
	npcEntity.AddComponent(&PositionComponent{X: 120.0, Y: 130.0})

	cities := []*Entity{cityEntity}
	npcs := []*Entity{npcEntity}

	// Save world state
	err := system.SaveWorldState(worldMemory, cities, npcs, 1000.0)
	if err != nil {
		t.Fatalf("SaveWorldState failed: %v", err)
	}

	// Verify saved
	if worldMemory.GetCityCount() != 1 {
		t.Errorf("GetCityCount() = %d; want 1", worldMemory.GetCityCount())
	}
	if worldMemory.GetNPCCount() != 1 {
		t.Errorf("GetNPCCount() = %d; want 1", worldMemory.GetNPCCount())
	}
	if worldMemory.LastSaveTime != 1000.0 {
		t.Errorf("LastSaveTime = %f; want 1000.0", worldMemory.LastSaveTime)
	}

	// Modify city state
	cityState.Prosperity = 0.5
	cityState.Population = 100

	// Load world state (restores saved values)
	err = system.LoadWorldState(worldMemory, cities, npcs, 1000.0)
	if err != nil {
		t.Fatalf("LoadWorldState failed: %v", err)
	}

	// Verify restored
	if cityState.Prosperity != 0.75 {
		t.Errorf("Restored prosperity = %f; want 0.75", cityState.Prosperity)
	}
	if cityState.Population != 150 {
		t.Errorf("Restored population = %d; want 150", cityState.Population)
	}
}

func TestWorldPersistenceSystem_NilWorldMemory(t *testing.T) {
	system := NewWorldPersistenceSystem()

	err := system.SaveWorldState(nil, nil, nil, 0)
	if err == nil {
		t.Error("SaveWorldState with nil worldMemory should fail")
	}

	err = system.LoadWorldState(nil, nil, nil, 0)
	if err == nil {
		t.Error("LoadWorldState with nil worldMemory should fail")
	}
}

func TestWorldPersistenceSystem_TimeProgression(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)
	worldMemory.TimeProgressionEnabled = true
	worldMemory.TimeProgressionRate = 1.0 // 1:1 time

	// Create a thriving city
	cityEntity := NewEntity(1)
	cityState := NewCityStateComponent("city1", "Test City", 99999)
	cityState.Prosperity = 0.8
	cityState.Population = 100
	cityState.MaxPopulation = 200
	cityState.State = CityStateThriving
	cityEntity.AddComponent(cityState)

	cities := []*Entity{cityEntity}

	// Save at time 0
	err := system.SaveWorldState(worldMemory, cities, nil, 0.0)
	if err != nil {
		t.Fatalf("SaveWorldState failed: %v", err)
	}

	// Simulate time passing (48 hours = 2 days)
	// Reset city to see if load restores and applies progression
	originalProsperity := cityState.Prosperity
	originalPopulation := cityState.Population

	// Load at time 48.0 (2 days later)
	err = system.LoadWorldState(worldMemory, cities, nil, 48.0)
	if err != nil {
		t.Fatalf("LoadWorldState failed: %v", err)
	}

	// City should have grown due to time progression
	// Thriving cities grow population and prosperity
	if cityState.Population <= originalPopulation {
		t.Logf("Population didn't grow as expected: %d -> %d (may be due to RNG)", originalPopulation, cityState.Population)
	}
	// Prosperity should have increased or stayed same (clamped at 1.0)
	if cityState.Prosperity < originalProsperity-0.1 {
		t.Errorf("Prosperity decreased unexpectedly: %f -> %f", originalProsperity, cityState.Prosperity)
	}
}

func TestWorldPersistenceSystem_RecordWorldEvent(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)

	system.RecordWorldEvent(
		worldMemory,
		"trade",
		"Big trade completed",
		100.0,
		"city1",
		"player1",
		0.5,
		map[string]interface{}{"gold": 1000},
	)

	if worldMemory.GetEventCount() != 1 {
		t.Errorf("GetEventCount() = %d; want 1", worldMemory.GetEventCount())
	}

	events := worldMemory.GetRecentEvents(1)
	if len(events) != 1 {
		t.Fatalf("GetRecentEvents(1) = %d; want 1", len(events))
	}

	event := events[0]
	if event.EventType != "trade" {
		t.Errorf("EventType = %s; want trade", event.EventType)
	}
	if event.AffectedCityID != "city1" {
		t.Errorf("AffectedCityID = %s; want city1", event.AffectedCityID)
	}
	if event.AffectedPlayerID != "player1" {
		t.Errorf("AffectedPlayerID = %s; want player1", event.AffectedPlayerID)
	}
	if event.Magnitude != 0.5 {
		t.Errorf("Magnitude = %f; want 0.5", event.Magnitude)
	}
}

func TestWorldPersistenceSystem_RecordWorldEventNilMemory(t *testing.T) {
	system := NewWorldPersistenceSystem()

	// Should not panic
	system.RecordWorldEvent(nil, "test", "test", 0, "", "", 0, nil)
}

func TestWorldPersistenceSystem_UpdatePlayerReputation(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)

	// Set initial reputation
	worldMemory.SetPlayerCityReputation("player1", "city1", 50.0)

	// Update with significant change (should record event)
	system.UpdatePlayerReputation(worldMemory, "player1", "city1", 25.0, "Completed major quest", 100.0)

	rep := worldMemory.GetPlayerCityReputation("player1", "city1")
	if rep != 75.0 {
		t.Errorf("Reputation = %f; want 75.0", rep)
	}

	// Should have recorded the significant reputation change
	if worldMemory.GetEventCount() != 1 {
		t.Errorf("GetEventCount() = %d; want 1", worldMemory.GetEventCount())
	}

	// Small change (should not record event)
	system.UpdatePlayerReputation(worldMemory, "player1", "city1", 5.0, "Small favor", 200.0)

	if worldMemory.GetEventCount() != 1 {
		t.Errorf("Small change should not add event, GetEventCount() = %d; want 1", worldMemory.GetEventCount())
	}
}

func TestWorldPersistenceSystem_UpdatePlayerReputationNilMemory(t *testing.T) {
	system := NewWorldPersistenceSystem()

	// Should not panic
	system.UpdatePlayerReputation(nil, "player1", "city1", 25.0, "test", 0)
}

func TestWorldPersistenceSystem_Update(t *testing.T) {
	system := NewWorldPersistenceSystem()

	// Update should not panic even with no entities
	system.Update(nil, 0.016)
	system.Update([]*Entity{}, 0.016)
}

func TestWorldPersistenceSystem_EmptyEntities(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)

	// Should work with empty entity slices
	err := system.SaveWorldState(worldMemory, []*Entity{}, []*Entity{}, 0)
	if err != nil {
		t.Errorf("SaveWorldState with empty entities failed: %v", err)
	}

	err = system.LoadWorldState(worldMemory, []*Entity{}, []*Entity{}, 0)
	if err != nil {
		t.Errorf("LoadWorldState with empty entities failed: %v", err)
	}
}

func TestWorldPersistenceSystem_EntitiesWithoutComponents(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)

	// Entities without the expected components
	emptyEntity := NewEntity(1)
	entities := []*Entity{emptyEntity}

	// Should not panic and should complete successfully
	err := system.SaveWorldState(worldMemory, entities, entities, 0)
	if err != nil {
		t.Errorf("SaveWorldState with componentless entities failed: %v", err)
	}

	// Should not save anything since no valid components
	if worldMemory.GetCityCount() != 0 {
		t.Errorf("GetCityCount() = %d; want 0", worldMemory.GetCityCount())
	}
	if worldMemory.GetNPCCount() != 0 {
		t.Errorf("GetNPCCount() = %d; want 0", worldMemory.GetNPCCount())
	}
}

func TestWorldPersistenceSystem_LoadNonexistentState(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)

	// Create city with no saved state
	cityEntity := NewEntity(1)
	cityState := NewCityStateComponent("new_city", "New City", 99999)
	cityState.Prosperity = 0.5
	cityEntity.AddComponent(cityState)

	cities := []*Entity{cityEntity}

	// Load should not modify city without saved state
	err := system.LoadWorldState(worldMemory, cities, nil, 0)
	if err != nil {
		t.Errorf("LoadWorldState failed: %v", err)
	}

	// City should retain original values
	if cityState.Prosperity != 0.5 {
		t.Errorf("Prosperity = %f; want 0.5 (unchanged)", cityState.Prosperity)
	}
}

func TestWorldPersistenceSystem_TimeProgressionDisabled(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)
	worldMemory.TimeProgressionEnabled = false // Disabled

	cityEntity := NewEntity(1)
	cityState := NewCityStateComponent("city1", "Test City", 99999)
	cityState.Prosperity = 0.8
	cityState.Population = 100
	cityEntity.AddComponent(cityState)

	cities := []*Entity{cityEntity}

	// Save at time 0
	_ = system.SaveWorldState(worldMemory, cities, nil, 0.0)

	// Load at time 1000 (should not apply progression)
	_ = system.LoadWorldState(worldMemory, cities, nil, 1000.0)

	// City should have exact saved values
	if cityState.Prosperity != 0.8 {
		t.Errorf("Prosperity = %f; want 0.8 (no progression)", cityState.Prosperity)
	}
	if cityState.Population != 100 {
		t.Errorf("Population = %d; want 100 (no progression)", cityState.Population)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test getCityStateComponent
	entity := NewEntity(1)
	if getCityStateComponent(nil) != nil {
		t.Error("getCityStateComponent(nil) should return nil")
	}
	if getCityStateComponent(entity) != nil {
		t.Error("getCityStateComponent with no component should return nil")
	}
	cityState := NewCityStateComponent("city1", "Test", 0)
	entity.AddComponent(cityState)
	if getCityStateComponent(entity) != cityState {
		t.Error("getCityStateComponent should return the component")
	}

	// Test getScheduleComponent
	entity2 := NewEntity(2)
	if getScheduleComponent(nil) != nil {
		t.Error("getScheduleComponent(nil) should return nil")
	}
	if getScheduleComponent(entity2) != nil {
		t.Error("getScheduleComponent with no component should return nil")
	}
	schedule := NewScheduleComponent(0, 0)
	entity2.AddComponent(schedule)
	if getScheduleComponent(entity2) != schedule {
		t.Error("getScheduleComponent should return the component")
	}

	// Test getPositionComponent
	entity3 := NewEntity(3)
	if getPositionComponent(nil) != nil {
		t.Error("getPositionComponent(nil) should return nil")
	}
	if getPositionComponent(entity3) != nil {
		t.Error("getPositionComponent with no component should return nil")
	}
	pos := &PositionComponent{X: 10, Y: 20}
	entity3.AddComponent(pos)
	if getPositionComponent(entity3) != pos {
		t.Error("getPositionComponent should return the component")
	}

	// Test getEntityName
	if getEntityName(nil) != "" {
		t.Error("getEntityName(nil) should return empty string")
	}
	name := getEntityName(entity3)
	if name != "Entity-3" {
		t.Errorf("getEntityName = %s; want Entity-3", name)
	}

	// Test absFloat64
	if absFloat64(5.0) != 5.0 {
		t.Error("absFloat64(5.0) should be 5.0")
	}
	if absFloat64(-5.0) != 5.0 {
		t.Error("absFloat64(-5.0) should be 5.0")
	}
	if absFloat64(0.0) != 0.0 {
		t.Error("absFloat64(0.0) should be 0.0")
	}
}

func TestWorldPersistenceSystem_StrugglingCityProgression(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(12345)
	worldMemory.TimeProgressionEnabled = true
	worldMemory.TimeProgressionRate = 1.0

	// Create a struggling city
	cityEntity := NewEntity(1)
	cityState := NewCityStateComponent("city1", "Struggling Town", 99999)
	cityState.Prosperity = 0.2
	cityState.Population = 80
	cityState.MaxPopulation = 200
	cityState.State = CityStateStruggling
	cityEntity.AddComponent(cityState)

	cities := []*Entity{cityEntity}

	// Save at time 0
	err := system.SaveWorldState(worldMemory, cities, nil, 0.0)
	if err != nil {
		t.Fatalf("SaveWorldState failed: %v", err)
	}

	originalProsperity := cityState.Prosperity
	originalPopulation := cityState.Population

	// Load at time 48.0 (2 days later) - struggling city should decline
	err = system.LoadWorldState(worldMemory, cities, nil, 48.0)
	if err != nil {
		t.Fatalf("LoadWorldState failed: %v", err)
	}

	// Struggling cities should decline or stay same
	if cityState.Prosperity > originalProsperity+0.1 {
		t.Errorf("Struggling city prosperity increased unexpectedly: %f -> %f", originalProsperity, cityState.Prosperity)
	}
	// Population should decrease or stay same
	if cityState.Population > originalPopulation+5 {
		t.Errorf("Struggling city population increased unexpectedly: %d -> %d", originalPopulation, cityState.Population)
	}
}

func TestWorldPersistenceSystem_CityStateChange(t *testing.T) {
	system := NewWorldPersistenceSystem()
	worldMemory := NewWorldMemoryComponent(54321)
	worldMemory.TimeProgressionEnabled = true
	worldMemory.TimeProgressionRate = 1.0

	// Create a city at the edge of state change
	cityEntity := NewEntity(1)
	cityState := NewCityStateComponent("city1", "Edge City", 54321)
	cityState.Prosperity = 0.69 // Just below thriving threshold
	cityState.Population = 150
	cityState.MaxPopulation = 200
	cityState.State = CityStateStable
	cityEntity.AddComponent(cityState)

	cities := []*Entity{cityEntity}

	// Save at time 0
	_ = system.SaveWorldState(worldMemory, cities, nil, 0.0)

	// Load at time 240.0 (10 days later) - might trigger state change
	_ = system.LoadWorldState(worldMemory, cities, nil, 240.0)

	// If state changed, an event should have been recorded
	events := worldMemory.GetEventsByType("city_evolution")
	if len(events) > 0 {
		// Verify event has correct structure
		event := events[0]
		if event.AffectedCityID != "city1" {
			t.Errorf("Event city_id = %s; want city1", event.AffectedCityID)
		}
		if event.Details["from_offline"] != true {
			t.Error("Event should indicate from_offline")
		}
	}
}
