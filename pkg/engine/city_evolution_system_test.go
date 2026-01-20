package engine

import (
	"testing"
)

func TestNewCityEvolutionSystem(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)

	sys := NewCityEvolutionSystem(world, clock)

	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.clock != clock {
		t.Error("clock not set correctly")
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0", sys.updateInterval)
	}
}

func TestCityEvolutionSystem_Update_NoCities(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	// Should not panic with no entities
	sys.Update([]*Entity{}, 1.0)
}

func TestCityEvolutionSystem_Update_ProcessesTriggers(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	// Create city entity
	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	cityState.Prosperity = 0.5
	entity.AddComponent(cityState)

	triggers := NewCityEvolutionTriggersComponent("city_1")
	triggers.QueueTrigger(EvolutionTrigger{
		TriggerType: EvolutionTradeComplete,
		Magnitude:   1.0,
	})
	entity.AddComponent(triggers)

	initialProsperity := cityState.Prosperity

	// Update system with enough time to trigger
	sys.Update([]*Entity{entity}, 1.5)

	if cityState.Prosperity <= initialProsperity {
		t.Error("Trade trigger should increase prosperity")
	}
	if triggers.HasPendingTriggers() {
		t.Error("Trigger should be processed")
	}
	if len(triggers.RecentTriggers) != 1 {
		t.Errorf("RecentTriggers len = %d, want 1", len(triggers.RecentTriggers))
	}
}

func TestCityEvolutionSystem_Update_StateTransition(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	cityState.Prosperity = 0.69 // Just below thriving (0.70 threshold)
	sys.UpdateCityState(cityState)
	entity.AddComponent(cityState)

	triggers := NewCityEvolutionTriggersComponent("city_1")
	triggers.QueueTrigger(EvolutionTrigger{
		TriggerType: EvolutionQuestComplete,
		Magnitude:   1.0, // +0.02 should push to 0.71, above thriving threshold
	})
	entity.AddComponent(triggers)

	if cityState.State != CityStateStable {
		t.Fatalf("Initial state should be stable, got %v", cityState.State)
	}

	sys.Update([]*Entity{entity}, 1.5)

	if cityState.State != CityStateThriving {
		t.Errorf("State should transition to thriving, got %v", cityState.State)
	}
}

func TestCityEvolutionSystem_Update_RaidDamage(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	cityState.Prosperity = 0.5
	cityState.Population = 100
	cityState.ResourceStockpile = 200.0
	entity.AddComponent(cityState)

	triggers := NewCityEvolutionTriggersComponent("city_1")
	triggers.QueueTrigger(EvolutionTrigger{
		TriggerType: EvolutionRaidAttack,
		Magnitude:   1.0,
	})
	entity.AddComponent(triggers)

	initialProsperity := cityState.Prosperity
	initialPopulation := cityState.Population
	initialResources := cityState.ResourceStockpile

	sys.Update([]*Entity{entity}, 1.5)

	if cityState.Prosperity >= initialProsperity {
		t.Error("Raid should decrease prosperity")
	}
	if cityState.Population >= initialPopulation {
		t.Error("Raid should decrease population")
	}
	if cityState.ResourceStockpile >= initialResources {
		t.Error("Raid should decrease resources")
	}
}

func TestCityEvolutionSystem_Update_DisabledProcessing(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	cityState.Prosperity = 0.5
	entity.AddComponent(cityState)

	triggers := NewCityEvolutionTriggersComponent("city_1")
	triggers.ProcessingEnabled = false
	triggers.QueueTrigger(EvolutionTrigger{
		TriggerType: EvolutionTradeComplete,
		Magnitude:   1.0,
	})
	entity.AddComponent(triggers)

	initialProsperity := cityState.Prosperity

	sys.Update([]*Entity{entity}, 1.5)

	if cityState.Prosperity != initialProsperity {
		t.Error("Triggers should not be processed when disabled")
	}
	if !triggers.HasPendingTriggers() {
		t.Error("Trigger should remain pending when processing disabled")
	}
}

func TestCityEvolutionSystem_NaturalEvolution_Thriving(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	cityState.Prosperity = 0.8
	cityState.State = CityStateThriving
	cityState.Population = 100
	cityState.MaxPopulation = 200
	cityState.ResourceStockpile = 100.0
	entity.AddComponent(cityState)

	initialPopulation := cityState.Population

	// Multiple updates
	for i := 0; i < 10; i++ {
		sys.Update([]*Entity{entity}, 1.5)
	}

	if cityState.Population <= initialPopulation {
		t.Error("Thriving city should grow population")
	}
}

func TestCityEvolutionSystem_NaturalEvolution_Struggling(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	cityState.Prosperity = 0.2
	cityState.State = CityStateStruggling
	cityState.Population = 100
	entity.AddComponent(cityState)

	initialProsperity := cityState.Prosperity
	initialPopulation := cityState.Population

	// Multiple updates
	for i := 0; i < 10; i++ {
		sys.Update([]*Entity{entity}, 1.5)
	}

	if cityState.Prosperity >= initialProsperity {
		t.Error("Struggling city should decline in prosperity")
	}
	if cityState.Population >= initialPopulation {
		t.Error("Struggling city should lose population")
	}
}

func TestCityEvolutionSystem_TriggerHelpers(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	entity.AddComponent(cityState)

	// Test TriggerTradeEvent
	sys.TriggerTradeEvent(entity, 500.0, "player_1")
	triggers, _ := entity.GetComponent("city_evolution_triggers")
	if triggers.(*CityEvolutionTriggersComponent).GetPendingCount() != 1 {
		t.Error("TriggerTradeEvent should queue trigger")
	}

	// Test EvolutionQuestComplete
	sys.EvolutionQuestComplete(entity, 0.5, "player_1")
	if triggers.(*CityEvolutionTriggersComponent).GetPendingCount() != 2 {
		t.Error("EvolutionQuestComplete should queue trigger")
	}

	// Test TriggerRaid
	sys.TriggerRaid(entity, 0.8, false, "enemy_1")
	if triggers.(*CityEvolutionTriggersComponent).GetPendingCount() != 3 {
		t.Error("TriggerRaid should queue trigger")
	}

	// Test TriggerBuildingChange
	sys.TriggerBuildingChange(entity, true, 0.5, "builder_1")
	if triggers.(*CityEvolutionTriggersComponent).GetPendingCount() != 4 {
		t.Error("TriggerBuildingChange should queue trigger")
	}

	// Test EvolutionResourceDonation
	sys.EvolutionResourceDonation(entity, 250.0, "donor_1")
	if triggers.(*CityEvolutionTriggersComponent).GetPendingCount() != 5 {
		t.Error("EvolutionResourceDonation should queue trigger")
	}
}

func TestCityEvolutionSystem_TriggerHelpers_CreatesComponent(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	// Entity without triggers component
	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test City", 12345)
	entity.AddComponent(cityState)

	// Should create triggers component
	sys.TriggerTradeEvent(entity, 100.0, "player_1")

	if !entity.HasComponent("city_evolution_triggers") {
		t.Error("Should create triggers component if missing")
	}
}

func TestGenerateCity(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name     string
		cityID   string
		cityName string
		seed     int64
		x, y     float64
	}{
		{"basic city", "city_1", "Test City", 12345, 100.0, 200.0},
		{"negative coords", "city_2", "Negative", 54321, -100.0, -200.0},
		{"zero seed", "city_3", "Zero Seed", 0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := GenerateCity(world, tt.cityID, tt.cityName, tt.seed, tt.x, tt.y)

			if entity == nil {
				t.Fatal("GenerateCity returned nil")
			}

			// Check position
			posComp, ok := entity.GetComponent("position")
			if !ok {
				t.Fatal("City should have position component")
			}
			pos := posComp.(*PositionComponent)
			if pos.X != tt.x || pos.Y != tt.y {
				t.Errorf("Position = (%v, %v), want (%v, %v)", pos.X, pos.Y, tt.x, tt.y)
			}

			// Check city state
			cityStateComp, ok := entity.GetComponent("city_state")
			if !ok {
				t.Fatal("City should have city_state component")
			}
			cityState := cityStateComp.(*CityStateComponent)
			if cityState.CityID != tt.cityID {
				t.Errorf("CityID = %v, want %v", cityState.CityID, tt.cityID)
			}
			if cityState.CityName != tt.cityName {
				t.Errorf("CityName = %v, want %v", cityState.CityName, tt.cityName)
			}
			if cityState.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", cityState.Seed, tt.seed)
			}

			// Check triggers
			if !entity.HasComponent("city_evolution_triggers") {
				t.Error("City should have city_evolution_triggers component")
			}
		})
	}
}

func TestGenerateCity_Determinism(t *testing.T) {
	seed := int64(99999)

	world1 := NewWorld()
	city1 := GenerateCity(world1, "city_det", "Determinism Test", seed, 0, 0)
	state1, _ := city1.GetComponent("city_state")
	cs1 := state1.(*CityStateComponent)

	world2 := NewWorld()
	city2 := GenerateCity(world2, "city_det", "Determinism Test", seed, 0, 0)
	state2, _ := city2.GetComponent("city_state")
	cs2 := state2.(*CityStateComponent)

	// Same seed should produce identical results
	if cs1.Prosperity != cs2.Prosperity {
		t.Errorf("Prosperity not deterministic: %v vs %v", cs1.Prosperity, cs2.Prosperity)
	}
	if cs1.Infrastructure != cs2.Infrastructure {
		t.Errorf("Infrastructure not deterministic: %v vs %v", cs1.Infrastructure, cs2.Infrastructure)
	}
	if cs1.Defense != cs2.Defense {
		t.Errorf("Defense not deterministic: %v vs %v", cs1.Defense, cs2.Defense)
	}
	if cs1.Population != cs2.Population {
		t.Errorf("Population not deterministic: %v vs %v", cs1.Population, cs2.Population)
	}
	if cs1.State != cs2.State {
		t.Errorf("State not deterministic: %v vs %v", cs1.State, cs2.State)
	}
}

func TestCityEvolutionSystem_GetCityByID(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	GenerateCity(world, "city_1", "City One", 1, 0, 0)
	GenerateCity(world, "city_2", "City Two", 2, 100, 0)
	GenerateCity(world, "city_3", "City Three", 3, 200, 0)

	// Find existing city
	found := sys.GetCityByID("city_2")
	if found == nil {
		t.Fatal("GetCityByID should find existing city")
	}
	cityState, _ := found.GetComponent("city_state")
	if cityState.(*CityStateComponent).CityName != "City Two" {
		t.Error("Found wrong city")
	}

	// Find non-existent city
	notFound := sys.GetCityByID("city_999")
	if notFound != nil {
		t.Error("GetCityByID should return nil for non-existent city")
	}
}

func TestCityEvolutionSystem_GetCitiesInState(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	// Create cities with different states
	city1 := GenerateCity(world, "city_1", "Struggling", 1, 0, 0)
	city1State, _ := city1.GetComponent("city_state")
	city1State.(*CityStateComponent).Prosperity = 0.1
	sys.UpdateCityState(city1State.(*CityStateComponent))

	city2 := GenerateCity(world, "city_2", "Thriving", 2, 100, 0)
	city2State, _ := city2.GetComponent("city_state")
	city2State.(*CityStateComponent).Prosperity = 0.9
	sys.UpdateCityState(city2State.(*CityStateComponent))

	city3 := GenerateCity(world, "city_3", "Also Thriving", 3, 200, 0)
	city3State, _ := city3.GetComponent("city_state")
	city3State.(*CityStateComponent).Prosperity = 0.8
	sys.UpdateCityState(city3State.(*CityStateComponent))

	// Get struggling cities
	struggling := sys.GetCitiesInState(CityStateStruggling)
	if len(struggling) != 1 {
		t.Errorf("GetCitiesInState(struggling) len = %d, want 1", len(struggling))
	}

	// Get thriving cities
	thriving := sys.GetCitiesInState(CityStateThriving)
	if len(thriving) != 2 {
		t.Errorf("GetCitiesInState(thriving) len = %d, want 2", len(thriving))
	}

	// Get stable cities (none)
	stable := sys.GetCitiesInState(CityStateStable)
	if len(stable) != 0 {
		t.Errorf("GetCitiesInState(stable) len = %d, want 0", len(stable))
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		value, min, max, want float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		got := clampFloat(tt.value, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clampFloat(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestCityEvolutionSystem_Update_BelowInterval(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	entity := world.CreateEntity()
	cityState := NewCityStateComponent("city_1", "Test", 12345)
	cityState.Prosperity = 0.5
	entity.AddComponent(cityState)

	triggers := NewCityEvolutionTriggersComponent("city_1")
	triggers.QueueTrigger(EvolutionTrigger{
		TriggerType: EvolutionTradeComplete,
		Magnitude:   1.0,
	})
	entity.AddComponent(triggers)

	// Update with time below interval
	sys.Update([]*Entity{entity}, 0.5)

	// Trigger should still be pending
	if !triggers.HasPendingTriggers() {
		t.Error("Trigger should remain pending when update interval not reached")
	}
}

func BenchmarkCityEvolutionSystem_Update(b *testing.B) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewCityEvolutionSystem(world, clock)

	// Create 100 cities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := GenerateCity(world, "city_"+string(rune(i)), "City", int64(i), float64(i*100), 0)
		entities[i] = entity

		// Add some triggers
		triggers, _ := entity.GetComponent("city_evolution_triggers")
		for j := 0; j < 5; j++ {
			triggers.(*CityEvolutionTriggersComponent).QueueTrigger(EvolutionTrigger{
				TriggerType: EvolutionTradeComplete,
				Magnitude:   0.5,
			})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 1.5)
	}
}
