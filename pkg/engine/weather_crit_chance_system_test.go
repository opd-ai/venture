package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherCritChanceSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherCritChanceSystem returned nil")
	}
	if system.world != world {
		t.Error("World reference not set correctly")
	}
	if system.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", system.genreID)
	}
}

func TestWeatherCritChanceSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("SetGenre(%s) failed, got %s", genre, system.genreID)
		}
	}
}

func TestWeatherCritChanceSystem_GetBaseModifier(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	tests := []struct {
		name     string
		weather  particles.WeatherType
		expected float64
	}{
		{"rain reduces crit", particles.WeatherRain, -0.03},
		{"snow reduces crit", particles.WeatherSnow, -0.02},
		{"fog increases crit", particles.WeatherFog, 0.08},
		{"dust increases crit", particles.WeatherDust, 0.05},
		{"sandstorm increases crit", particles.WeatherSandstorm, 0.05},
		{"ash increases crit", particles.WeatherAsh, 0.03},
		{"neon rain reduces crit", particles.WeatherNeonRain, -0.02},
		{"smog increases crit", particles.WeatherSmog, 0.05},
		{"radiation increases crit", particles.WeatherRadiation, 0.06},
		{"blood rain increases crit", particles.WeatherBloodRain, 0.10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.getBaseModifier(tt.weather)
			if result != tt.expected {
				t.Errorf("getBaseModifier(%v) = %v, want %v", tt.weather, result, tt.expected)
			}
		})
	}
}

func TestWeatherCritChanceSystem_GetGenreBonus(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	tests := []struct {
		name     string
		genre    string
		weather  particles.WeatherType
		expected float64
	}{
		{"horror blood rain", "horror", particles.WeatherBloodRain, 0.05},
		{"horror fog", "horror", particles.WeatherFog, 0.03},
		{"cyberpunk smog", "cyberpunk", particles.WeatherSmog, 0.03},
		{"postapoc radiation", "postapoc", particles.WeatherRadiation, 0.04},
		{"postapoc sandstorm", "postapoc", particles.WeatherSandstorm, 0.02},
		{"scifi dust", "scifi", particles.WeatherDust, 0.02},
		{"fantasy fog", "fantasy", particles.WeatherFog, 0.02},
		{"no bonus scenario", "fantasy", particles.WeatherRain, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.SetGenre(tt.genre)
			result := system.getGenreBonus(tt.weather)
			if result != tt.expected {
				t.Errorf("getGenreBonus(%v) with genre %s = %v, want %v",
					tt.weather, tt.genre, result, tt.expected)
			}
		})
	}
}

func TestWeatherCritChanceSystem_UpdateWithNoWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	// Create entity with stats
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Update without weather
	system.Update(entities, 1.0)

	// CritChance should be unchanged
	if stats.CritChance != 0.10 {
		t.Errorf("CritChance = %v, want 0.10 (no weather)", stats.CritChance)
	}
}

func TestWeatherCritChanceSystem_UpdateWithFogWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherFog,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity with stats
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Update with fog weather
	system.Update(entities, 1.0)

	// Expected: 0.10 base + 0.08 fog + 0.02 fantasy bonus = 0.20
	expectedCrit := 0.20
	if stats.CritChance != expectedCrit {
		t.Errorf("CritChance = %v, want %v (fog weather)", stats.CritChance, expectedCrit)
	}
}

func TestWeatherCritChanceSystem_UpdateWithRainWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherRain,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity with stats
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.15
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Update with rain weather
	system.Update(entities, 1.0)

	// Expected: 0.15 base - 0.03 rain penalty = 0.12
	expectedCrit := 0.12
	if stats.CritChance != expectedCrit {
		t.Errorf("CritChance = %v, want %v (rain weather)", stats.CritChance, expectedCrit)
	}
}

func TestWeatherCritChanceSystem_CritChanceClamping(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)
	system.SetGenre("horror")

	// Create blood rain weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherBloodRain,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity with high base crit
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.70 // High base crit
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Update - should cap at 0.75
	system.Update(entities, 1.0)

	// Expected: 0.70 + 0.10 + 0.05 = 0.85, clamped to 0.75
	if stats.CritChance > 0.75 {
		t.Errorf("CritChance = %v, should be clamped to 0.75 max", stats.CritChance)
	}
}

func TestWeatherCritChanceSystem_CritChanceNegativeClamping(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	// Create rain weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherRain,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity with very low base crit
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.01 // Very low base
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Update - rain penalty should not make crit negative
	system.Update(entities, 1.0)

	// Expected: 0.01 - 0.03 = -0.02, clamped to 0.0
	if stats.CritChance < 0.0 {
		t.Errorf("CritChance = %v, should be clamped to 0.0 min", stats.CritChance)
	}
}

func TestWeatherCritChanceSystem_ClearAllBonuses(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	// Create fog weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherFog,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Apply bonus
	system.Update(entities, 1.0)
	if stats.CritChance == 0.10 {
		t.Error("Bonus should have been applied")
	}

	// Clear bonuses
	system.clearAllBonuses()

	// Check original restored
	if stats.CritChance != 0.10 {
		t.Errorf("CritChance = %v, want 0.10 after clearAllBonuses", stats.CritChance)
	}
}

func TestWeatherCritChanceSystem_GetCurrentModifier(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)
	system.SetGenre("fantasy")

	// No weather - should return 0
	modifier := system.GetCurrentModifier()
	if modifier != 0.0 {
		t.Errorf("GetCurrentModifier() = %v, want 0.0 (no weather)", modifier)
	}

	// Add fog weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherFog,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Should return fog modifier + fantasy bonus
	modifier = system.GetCurrentModifier()
	expected := 0.08 + 0.02 // fog base + fantasy bonus
	if modifier != expected {
		t.Errorf("GetCurrentModifier() = %v, want %v", modifier, expected)
	}
}

func TestWeatherCritChanceSystem_WeatherChange(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherFog,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Apply fog bonus
	system.Update(entities, 1.0)
	fogCrit := stats.CritChance

	// Change to rain
	weatherComp.Config.Type = particles.WeatherRain
	system.Update(entities, 1.0)

	// Should have different crit now
	if stats.CritChance == fogCrit {
		t.Error("CritChance should change when weather changes")
	}
}

func TestWeatherCritChanceSystem_EntityWithoutStats(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: WeatherConfig{
			Type: particles.WeatherFog,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity WITHOUT stats
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{entity}

	// Should not panic
	system.Update(entities, 1.0)
}

func TestWeatherCritChanceSystem_InactiveWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCritChanceSystem(world, 12345)

	// Create INACTIVE weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: false, // Not active
		Config: WeatherConfig{
			Type: particles.WeatherFog,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// Update - should not apply bonus (weather inactive)
	system.Update(entities, 1.0)

	if stats.CritChance != 0.10 {
		t.Errorf("CritChance = %v, want 0.10 (inactive weather)", stats.CritChance)
	}
}
