package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherBlockChanceSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherBlockChanceSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.genreID != "fantasy" {
		t.Errorf("default genre = %q, want fantasy", system.genreID)
	}
	if system.updateInterval != 0.5 {
		t.Errorf("updateInterval = %f, want 0.5", system.updateInterval)
	}
}

func TestWeatherBlockChanceSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			if system.genreID != tt.genre {
				t.Errorf("genreID = %q, want %q", system.genreID, tt.genre)
			}
		})
	}
}

func TestWeatherBlockChanceSystem_BaseModifiers(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	tests := []struct {
		weather  particles.WeatherType
		expected float64
	}{
		{particles.WeatherRain, -0.10},
		{particles.WeatherSnow, -0.12},
		{particles.WeatherBlizzard, -0.20},
		{particles.WeatherFog, -0.08},
		{particles.WeatherDust, -0.06},
		{particles.WeatherStorm, -0.18},
		{particles.WeatherClear, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.weather.String(), func(t *testing.T) {
			modifier := system.getBaseModifier(tt.weather)
			if modifier != tt.expected {
				t.Errorf("getBaseModifier(%s) = %f, want %f", tt.weather, modifier, tt.expected)
			}
		})
	}
}

func TestWeatherBlockChanceSystem_GenreMultipliers(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.5},
		{"horror", 2.0},
		{"cyberpunk", 0.7},
		{"postapoc", 1.2},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.genreID = tt.genre
			multiplier := system.getGenreMultiplier()
			if multiplier != tt.expected {
				t.Errorf("getGenreMultiplier(%s) = %f, want %f", tt.genre, multiplier, tt.expected)
			}
		})
	}
}

func TestWeatherBlockChanceSystem_UpdateWithWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	// Create entity with stats
	entity := world.CreateEntity()
	stats := &StatsComponent{BlockChance: 0.25}
	entity.AddComponent(stats)

	// Create weather entity with rain
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	// Initial update (accumulate time)
	system.Update([]*Entity{entity, weatherEntity}, 0.1)
	// Verify no change yet (interval not reached)
	if stats.BlockChance != 0.25 {
		t.Errorf("BlockChance changed before interval: %f", stats.BlockChance)
	}

	// Update past interval
	system.Update([]*Entity{entity, weatherEntity}, 0.5)

	// Rain should reduce block by 10%
	expectedBlock := 0.25 - 0.10
	if stats.BlockChance != expectedBlock {
		t.Errorf("BlockChance = %f, want %f", stats.BlockChance, expectedBlock)
	}
}

func TestWeatherBlockChanceSystem_UpdateNoWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	entity := world.CreateEntity()
	stats := &StatsComponent{BlockChance: 0.25}
	entity.AddComponent(stats)

	// Update without weather entity
	system.Update([]*Entity{entity}, 0.6)

	// Block chance should remain unchanged
	if stats.BlockChance != 0.25 {
		t.Errorf("BlockChance = %f, want 0.25 (unchanged)", stats.BlockChance)
	}
}

func TestWeatherBlockChanceSystem_UpdateInactiveWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	entity := world.CreateEntity()
	stats := &StatsComponent{BlockChance: 0.25}
	entity.AddComponent(stats)

	// Create inactive weather
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: false}
	weatherEntity.AddComponent(weatherComp)

	system.Update([]*Entity{entity, weatherEntity}, 0.6)

	// Block chance should remain unchanged
	if stats.BlockChance != 0.25 {
		t.Errorf("BlockChance = %f, want 0.25 (unchanged)", stats.BlockChance)
	}
}

func TestWeatherBlockChanceSystem_GenreAffectsModifier(t *testing.T) {
	tests := []struct {
		genre         string
		weather       particles.WeatherType
		baseBlock     float64
		expectedBlock float64
	}{
		// Fantasy: full rain penalty (-10%)
		{"fantasy", particles.WeatherRain, 0.30, 0.20},
		// Scifi: half rain penalty (-5%)
		{"scifi", particles.WeatherRain, 0.30, 0.25},
		// Horror: double rain penalty (-20%)
		{"horror", particles.WeatherRain, 0.30, 0.10},
		// Cyberpunk: 70% rain penalty (-7%)
		{"cyberpunk", particles.WeatherRain, 0.30, 0.23},
		// Postapoc: 120% rain penalty (-12%)
		{"postapoc", particles.WeatherRain, 0.30, 0.18},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			system := NewWeatherBlockChanceSystem(world, 12345)
			system.SetGenre(tt.genre)

			entity := world.CreateEntity()
			stats := &StatsComponent{BlockChance: tt.baseBlock}
			entity.AddComponent(stats)

			weatherEntity := world.CreateEntity()
			weatherConfig := &particles.WeatherConfig{Type: tt.weather}
			weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
			weatherEntity.AddComponent(weatherComp)

			system.Update([]*Entity{entity, weatherEntity}, 0.6)

			// Allow for floating point imprecision
			diff := stats.BlockChance - tt.expectedBlock
			if diff < -0.01 || diff > 0.01 {
				t.Errorf("%s genre: BlockChance = %f, want %f", tt.genre, stats.BlockChance, tt.expectedBlock)
			}
		})
	}
}

func TestWeatherBlockChanceSystem_ClampMin(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)
	system.SetGenre("horror") // Double penalties

	entity := world.CreateEntity()
	// Start with low block chance
	stats := &StatsComponent{BlockChance: 0.10}
	entity.AddComponent(stats)

	// Blizzard with horror = -20% * 2 = -40%, but clamped to 0
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherBlizzard}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	system.Update([]*Entity{entity, weatherEntity}, 0.6)

	if stats.BlockChance < 0.0 {
		t.Errorf("BlockChance = %f, should be clamped to 0.0", stats.BlockChance)
	}
}

func TestWeatherBlockChanceSystem_ClampMax(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	entity := world.CreateEntity()
	// Start with high block chance (shouldn't happen naturally, but test clamp)
	stats := &StatsComponent{BlockChance: 0.55}
	entity.AddComponent(stats)

	// Clear weather to trigger restore
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherClear}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	system.Update([]*Entity{entity, weatherEntity}, 0.6)

	// Clear weather has 0 modifier, but value should be clamped to 0.50 max
	if stats.BlockChance > 0.55 {
		t.Errorf("BlockChance = %f, should not exceed stored value", stats.BlockChance)
	}
}

func TestWeatherBlockChanceSystem_ClearBonuses(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	entity := world.CreateEntity()
	stats := &StatsComponent{BlockChance: 0.30}
	entity.AddComponent(stats)

	// Apply rain weather
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	system.Update([]*Entity{entity, weatherEntity}, 0.6)

	// Verify modifier applied
	if stats.BlockChance == 0.30 {
		t.Error("Weather modifier not applied")
	}

	// Deactivate weather
	weatherComp.Active = false
	system.Update([]*Entity{entity, weatherEntity}, 0.6)

	// Block chance should be restored
	if stats.BlockChance != 0.30 {
		t.Errorf("BlockChance = %f, want 0.30 (restored)", stats.BlockChance)
	}
}

func TestWeatherBlockChanceSystem_GetCurrentModifier(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	// No weather
	modifier := system.GetCurrentModifier()
	if modifier != 0.0 {
		t.Errorf("GetCurrentModifier() = %f, want 0.0 with no weather", modifier)
	}

	// Add rain weather
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	modifier = system.GetCurrentModifier()
	if modifier != -0.10 {
		t.Errorf("GetCurrentModifier() = %f, want -0.10 with rain", modifier)
	}
}

func TestWeatherBlockChanceSystem_NilWorld(t *testing.T) {
	system := NewWeatherBlockChanceSystem(nil, 12345)

	// Should not panic
	system.Update(nil, 0.6)

	modifier := system.GetCurrentModifier()
	if modifier != 0.0 {
		t.Errorf("GetCurrentModifier() = %f, want 0.0 with nil world", modifier)
	}
}

func TestWeatherBlockChanceSystem_EntityWithoutStats(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	// Entity without stats component
	entity := world.CreateEntity()

	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	// Should not panic
	system.Update([]*Entity{entity, weatherEntity}, 0.6)
}

func TestWeatherBlockChanceSystem_Determinism(t *testing.T) {
	// Verify same seed produces same results
	seeds := []int64{12345, 67890, 11111}

	for _, seed := range seeds {
		t.Run("seed", func(t *testing.T) {
			world1 := NewWorld()
			world2 := NewWorld()

			sys1 := NewWeatherBlockChanceSystem(world1, seed)
			sys2 := NewWeatherBlockChanceSystem(world2, seed)

			// Verify both systems have same RNG state
			val1 := sys1.rng.Float64()
			val2 := sys2.rng.Float64()

			if val1 != val2 {
				t.Errorf("RNG values differ: %f vs %f", val1, val2)
			}
		})
	}
}

func TestWeatherBlockChanceSystem_WeatherChangeResetsCache(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	entity := world.CreateEntity()
	stats := &StatsComponent{BlockChance: 0.30}
	entity.AddComponent(stats)

	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)

	// Apply rain
	system.Update([]*Entity{entity, weatherEntity}, 0.6)
	rainBlock := stats.BlockChance

	// Change to snow
	weatherConfig.Type = particles.WeatherSnow
	system.Update([]*Entity{entity, weatherEntity}, 0.6)
	snowBlock := stats.BlockChance

	// Snow has different penalty than rain
	if rainBlock == snowBlock {
		t.Error("Block chance should differ between rain and snow")
	}

	// Snow penalty (-12%) vs rain (-10%)
	expectedSnowBlock := 0.30 - 0.12
	if snowBlock != expectedSnowBlock {
		t.Errorf("Snow BlockChance = %f, want %f", snowBlock, expectedSnowBlock)
	}
}

func BenchmarkWeatherBlockChanceSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 12345)

	// Create 100 entities with stats
	entities := make([]*Entity, 101)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&StatsComponent{BlockChance: 0.25})
		entities[i] = entity
	}

	// Add weather entity
	weatherEntity := world.CreateEntity()
	weatherConfig := &particles.WeatherConfig{Type: particles.WeatherRain}
	weatherComp := &WeatherComponent{Config: weatherConfig, Active: true}
	weatherEntity.AddComponent(weatherComp)
	entities[100] = weatherEntity

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.timeSinceCheck = 0 // Reset to force update
		system.Update(entities, 0.6)
	}
}

// Compile-time verification of deterministic RNG usage
func TestWeatherBlockChanceSystem_UsesProvidedSeed(t *testing.T) {
	world := NewWorld()
	system := NewWeatherBlockChanceSystem(world, 42)

	// Verify RNG is seeded with provided value
	expected := rand.New(rand.NewSource(42))
	actual := system.rng

	// Generate same sequence
	for i := 0; i < 10; i++ {
		e := expected.Float64()
		a := actual.Float64()
		if e != a {
			t.Errorf("RNG mismatch at iteration %d: expected %f, got %f", i, e, a)
		}
	}
}
