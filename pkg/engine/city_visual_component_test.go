package engine

import (
	"math/rand"
	"testing"
)

func TestNewCityVisualComponent(t *testing.T) {
	c := NewCityVisualComponent("city_test")

	if c.CityID != "city_test" {
		t.Errorf("CityID = %v, want city_test", c.CityID)
	}
	if c.VisualStyle != VisualStyleModest {
		t.Errorf("VisualStyle = %v, want modest", c.VisualStyle)
	}
	if c.BuildingCondition != 0.5 {
		t.Errorf("BuildingCondition = %v, want 0.5", c.BuildingCondition)
	}
	if c.PopulationActivity != 1.0 {
		t.Errorf("PopulationActivity = %v, want 1.0", c.PopulationActivity)
	}
}

func TestCityVisualComponent_Type(t *testing.T) {
	c := NewCityVisualComponent("test")
	if got := c.Type(); got != "city_visual" {
		t.Errorf("Type() = %v, want city_visual", got)
	}
}

func TestCityVisualComponent_UpdateFromCityState(t *testing.T) {
	tests := []struct {
		name           string
		prosperity     float64
		infrastructure float64
		defense        float64
		state          CityState
		wantStyle      CityVisualStyle
		wantBanners    int
	}{
		{
			name:           "struggling city",
			prosperity:     0.1,
			infrastructure: 0.2,
			defense:        0.1,
			state:          CityStateStruggling,
			wantStyle:      VisualStyleRundown,
			wantBanners:    0,
		},
		{
			name:           "stable city",
			prosperity:     0.5,
			infrastructure: 0.5,
			defense:        0.5,
			state:          CityStateStable,
			wantStyle:      VisualStyleModest,
			wantBanners:    2,
		},
		{
			name:           "thriving city",
			prosperity:     0.9,
			infrastructure: 0.9,
			defense:        0.8,
			state:          CityStateThriving,
			wantStyle:      VisualStyleProsperous,
			wantBanners:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visual := NewCityVisualComponent("test")
			cityState := NewCityStateComponent("test", "Test City", 12345)
			cityState.Prosperity = tt.prosperity
			cityState.Infrastructure = tt.infrastructure
			cityState.Defense = tt.defense
			cityState.State = tt.state
			cityState.Population = 100
			cityState.MaxPopulation = 200

			visual.UpdateFromCityState(cityState)

			if visual.VisualStyle != tt.wantStyle {
				t.Errorf("VisualStyle = %v, want %v", visual.VisualStyle, tt.wantStyle)
			}
			if visual.BannerCount != tt.wantBanners {
				t.Errorf("BannerCount = %d, want %d", visual.BannerCount, tt.wantBanners)
			}

			// Building condition should follow infrastructure
			if visual.BuildingCondition != tt.infrastructure {
				t.Errorf("BuildingCondition = %v, want %v", visual.BuildingCondition, tt.infrastructure)
			}

			// Guard presence should follow defense
			if visual.GuardPresence != tt.defense {
				t.Errorf("GuardPresence = %v, want %v", visual.GuardPresence, tt.defense)
			}

			// Debris should be inverse of prosperity
			expectedDebris := 1.0 - tt.prosperity
			if visual.DebrisDensity != expectedDebris {
				t.Errorf("DebrisDensity = %v, want %v", visual.DebrisDensity, expectedDebris)
			}
		})
	}
}

func TestCityVisualComponent_UpdateFromCityState_NilSafe(t *testing.T) {
	visual := NewCityVisualComponent("test")
	// Should not panic
	visual.UpdateFromCityState(nil)
}

func TestCityVisualComponent_GetBuildingSpriteVariant(t *testing.T) {
	tests := []struct {
		condition float64
		want      int
	}{
		{0.0, 0}, // damaged
		{0.2, 0}, // damaged
		{0.3, 1}, // normal
		{0.5, 1}, // normal
		{0.6, 1}, // normal
		{0.7, 2}, // fancy
		{1.0, 2}, // fancy
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			visual := NewCityVisualComponent("test")
			visual.BuildingCondition = tt.condition

			got := visual.GetBuildingSpriteVariant()
			if got != tt.want {
				t.Errorf("GetBuildingSpriteVariant() with condition %v = %d, want %d", tt.condition, got, tt.want)
			}
		})
	}
}

func TestCityVisualComponent_ShouldSpawnDecoration(t *testing.T) {
	visual := NewCityVisualComponent("test")
	rng := rand.New(rand.NewSource(12345))

	// High density should spawn more often
	visual.DecorationDensity = 1.0
	spawnCount := 0
	for i := 0; i < 100; i++ {
		if visual.ShouldSpawnDecoration(rng) {
			spawnCount++
		}
	}
	if spawnCount != 100 {
		t.Errorf("With density 1.0, should spawn 100%%, got %d%%", spawnCount)
	}

	// Zero density should never spawn
	visual.DecorationDensity = 0.0
	spawnCount = 0
	for i := 0; i < 100; i++ {
		if visual.ShouldSpawnDecoration(rng) {
			spawnCount++
		}
	}
	if spawnCount != 0 {
		t.Errorf("With density 0.0, should spawn 0%%, got %d%%", spawnCount)
	}
}

func TestCityVisualComponent_ShouldSpawnDebris(t *testing.T) {
	visual := NewCityVisualComponent("test")
	rng := rand.New(rand.NewSource(54321))

	visual.DebrisDensity = 1.0
	count := 0
	for i := 0; i < 100; i++ {
		if visual.ShouldSpawnDebris(rng) {
			count++
		}
	}
	if count != 100 {
		t.Errorf("With debris 1.0, should spawn 100%%, got %d%%", count)
	}
}

func TestCityVisualComponent_GetNPCSpawnMultiplier(t *testing.T) {
	visual := NewCityVisualComponent("test")

	visual.PopulationActivity = 1.5
	if got := visual.GetNPCSpawnMultiplier(); got != 1.5 {
		t.Errorf("GetNPCSpawnMultiplier() = %v, want 1.5", got)
	}

	visual.PopulationActivity = 0.5
	if got := visual.GetNPCSpawnMultiplier(); got != 0.5 {
		t.Errorf("GetNPCSpawnMultiplier() = %v, want 0.5", got)
	}
}

func TestCityVisualComponent_Serialization(t *testing.T) {
	original := NewCityVisualComponent("serialize_test")
	original.VisualStyle = VisualStyleProsperous
	original.BuildingCondition = 0.9
	original.RoadCondition = 0.85
	original.DecorationDensity = 0.7
	original.LightingLevel = 0.8
	original.PopulationActivity = 1.5
	original.MarketActivity = 0.9
	original.GuardPresence = 0.6
	original.DebrisDensity = 0.1
	original.BannerCount = 5
	original.PrimaryColor = 120
	original.SecondaryColor = 240

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize
	restored := &CityVisualComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify fields
	if restored.CityID != original.CityID {
		t.Errorf("CityID = %v, want %v", restored.CityID, original.CityID)
	}
	if restored.VisualStyle != original.VisualStyle {
		t.Errorf("VisualStyle = %v, want %v", restored.VisualStyle, original.VisualStyle)
	}
	if restored.BuildingCondition != original.BuildingCondition {
		t.Errorf("BuildingCondition = %v, want %v", restored.BuildingCondition, original.BuildingCondition)
	}
	if restored.BannerCount != original.BannerCount {
		t.Errorf("BannerCount = %d, want %d", restored.BannerCount, original.BannerCount)
	}
	if restored.PrimaryColor != original.PrimaryColor {
		t.Errorf("PrimaryColor = %d, want %d", restored.PrimaryColor, original.PrimaryColor)
	}
}

func TestCityVisualComponent_Deserialize_InvalidData(t *testing.T) {
	c := &CityVisualComponent{}
	err := c.Deserialize([]byte("not valid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func TestGenerateCityVisualFromSeed(t *testing.T) {
	visual := GenerateCityVisualFromSeed("seed_test", 12345)

	if visual.CityID != "seed_test" {
		t.Errorf("CityID = %v, want seed_test", visual.CityID)
	}
	if visual.PrimaryColor < 0 || visual.PrimaryColor >= 360 {
		t.Errorf("PrimaryColor = %d, should be 0-359", visual.PrimaryColor)
	}
	if visual.SecondaryColor < 0 || visual.SecondaryColor >= 360 {
		t.Errorf("SecondaryColor = %d, should be 0-359", visual.SecondaryColor)
	}
}

func TestGenerateCityVisualFromSeed_Determinism(t *testing.T) {
	seed := int64(99999)

	visual1 := GenerateCityVisualFromSeed("det_test", seed)
	visual2 := GenerateCityVisualFromSeed("det_test", seed)

	if visual1.PrimaryColor != visual2.PrimaryColor {
		t.Errorf("PrimaryColor not deterministic: %d vs %d", visual1.PrimaryColor, visual2.PrimaryColor)
	}
	if visual1.SecondaryColor != visual2.SecondaryColor {
		t.Errorf("SecondaryColor not deterministic: %d vs %d", visual1.SecondaryColor, visual2.SecondaryColor)
	}
}

func TestCityVisualComponent_UpdateFromCityState_PopulationActivity(t *testing.T) {
	visual := NewCityVisualComponent("test")
	cityState := NewCityStateComponent("test", "Test City", 12345)
	cityState.Prosperity = 1.0
	cityState.State = CityStateThriving
	cityState.Population = 200
	cityState.MaxPopulation = 200 // Full capacity

	visual.UpdateFromCityState(cityState)

	// Population activity should be clamped to max 2.0
	if visual.PopulationActivity > 2.0 {
		t.Errorf("PopulationActivity = %v, should be clamped to 2.0", visual.PopulationActivity)
	}
	if visual.PopulationActivity < 0.5 {
		t.Errorf("PopulationActivity = %v, should be at least 0.5", visual.PopulationActivity)
	}
}
