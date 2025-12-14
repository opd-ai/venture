package engine

import (
	"testing"
)

func TestNewCityStateComponent(t *testing.T) {
	tests := []struct {
		name     string
		cityID   string
		cityName string
		seed     int64
	}{
		{"basic city", "city_1", "Test City", 12345},
		{"empty name", "city_2", "", 0},
		{"large seed", "city_3", "Large Seed City", 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCityStateComponent(tt.cityID, tt.cityName, tt.seed)

			if c.CityID != tt.cityID {
				t.Errorf("CityID = %v, want %v", c.CityID, tt.cityID)
			}
			if c.CityName != tt.cityName {
				t.Errorf("CityName = %v, want %v", c.CityName, tt.cityName)
			}
			if c.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", c.Seed, tt.seed)
			}
			// Verify defaults
			if c.Prosperity != 0.5 {
				t.Errorf("Prosperity = %v, want 0.5", c.Prosperity)
			}
			if c.State != CityStateStable {
				t.Errorf("State = %v, want stable", c.State)
			}
			if c.Population != 100 {
				t.Errorf("Population = %v, want 100", c.Population)
			}
		})
	}
}

func TestCityStateComponent_Type(t *testing.T) {
	c := NewCityStateComponent("test", "Test", 0)
	if got := c.Type(); got != "city_state" {
		t.Errorf("Type() = %v, want city_state", got)
	}
}

func TestCityStateComponent_UpdateState(t *testing.T) {
	tests := []struct {
		name       string
		prosperity float64
		wantState  CityState
		wantChange bool
	}{
		{"struggling low", 0.0, CityStateStruggling, true},
		{"struggling boundary", 0.29, CityStateStruggling, true},
		{"stable low", 0.3, CityStateStable, false},
		{"stable mid", 0.5, CityStateStable, false},
		{"stable high", 0.69, CityStateStable, false},
		{"thriving boundary", 0.7, CityStateThriving, true},
		{"thriving high", 1.0, CityStateThriving, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCityStateComponent("test", "Test", 0)
			c.Prosperity = tt.prosperity

			changed := c.UpdateState()
			if c.State != tt.wantState {
				t.Errorf("UpdateState() state = %v, want %v", c.State, tt.wantState)
			}
			if changed != tt.wantChange {
				t.Errorf("UpdateState() changed = %v, want %v", changed, tt.wantChange)
			}
		})
	}
}

func TestCityStateComponent_GetProsperityTier(t *testing.T) {
	tests := []struct {
		state CityState
		want  string
	}{
		{CityStateStruggling, "Struggling"},
		{CityStateStable, "Stable"},
		{CityStateThriving, "Thriving"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			c := NewCityStateComponent("test", "Test", 0)
			c.State = tt.state

			if got := c.GetProsperityTier(); got != tt.want {
				t.Errorf("GetProsperityTier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCityStateComponent_GetPopulationRatio(t *testing.T) {
	tests := []struct {
		name          string
		population    int
		maxPopulation int
		want          float64
	}{
		{"half capacity", 100, 200, 0.5},
		{"full capacity", 200, 200, 1.0},
		{"empty", 0, 200, 0.0},
		{"zero max", 100, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCityStateComponent("test", "Test", 0)
			c.Population = tt.population
			c.MaxPopulation = tt.maxPopulation

			got := c.GetPopulationRatio()
			if got != tt.want {
				t.Errorf("GetPopulationRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCityStateComponent_CanGrowPopulation(t *testing.T) {
	tests := []struct {
		name          string
		population    int
		maxPopulation int
		prosperity    float64
		want          bool
	}{
		{"can grow", 100, 200, 0.5, true},
		{"at capacity", 200, 200, 0.5, false},
		{"low prosperity", 100, 200, 0.2, false},
		{"at threshold", 100, 200, 0.3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCityStateComponent("test", "Test", 0)
			c.Population = tt.population
			c.MaxPopulation = tt.maxPopulation
			c.Prosperity = tt.prosperity

			if got := c.CanGrowPopulation(); got != tt.want {
				t.Errorf("CanGrowPopulation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCityStateComponent_IsOvercrowded(t *testing.T) {
	tests := []struct {
		name          string
		population    int
		maxPopulation int
		want          bool
	}{
		{"not crowded", 100, 200, false},
		{"at threshold", 180, 200, false},
		{"overcrowded", 181, 200, true},
		{"full", 200, 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCityStateComponent("test", "Test", 0)
			c.Population = tt.population
			c.MaxPopulation = tt.maxPopulation

			if got := c.IsOvercrowded(); got != tt.want {
				t.Errorf("IsOvercrowded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCityStateComponent_Serialization(t *testing.T) {
	original := NewCityStateComponent("city_serialize", "Serialize Test", 99999)
	original.Prosperity = 0.75
	original.Population = 150
	original.MaxPopulation = 250
	original.Infrastructure = 0.8
	original.Defense = 0.6
	original.State = CityStateThriving
	original.TradeVolume = 5000.0
	original.ResourceStockpile = 200.0

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize() returned empty data")
	}

	// Deserialize into new component
	restored := &CityStateComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify all fields
	if restored.CityID != original.CityID {
		t.Errorf("CityID = %v, want %v", restored.CityID, original.CityID)
	}
	if restored.CityName != original.CityName {
		t.Errorf("CityName = %v, want %v", restored.CityName, original.CityName)
	}
	if restored.Prosperity != original.Prosperity {
		t.Errorf("Prosperity = %v, want %v", restored.Prosperity, original.Prosperity)
	}
	if restored.Population != original.Population {
		t.Errorf("Population = %v, want %v", restored.Population, original.Population)
	}
	if restored.MaxPopulation != original.MaxPopulation {
		t.Errorf("MaxPopulation = %v, want %v", restored.MaxPopulation, original.MaxPopulation)
	}
	if restored.State != original.State {
		t.Errorf("State = %v, want %v", restored.State, original.State)
	}
	if restored.TradeVolume != original.TradeVolume {
		t.Errorf("TradeVolume = %v, want %v", restored.TradeVolume, original.TradeVolume)
	}
	if restored.Seed != original.Seed {
		t.Errorf("Seed = %v, want %v", restored.Seed, original.Seed)
	}
}

func TestCityStateComponent_Deserialize_InvalidData(t *testing.T) {
	c := &CityStateComponent{}
	err := c.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}
