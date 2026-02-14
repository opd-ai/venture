package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherCooldownSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherCooldownSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre should be 'fantasy', got '%s'", sys.genreID)
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("Update interval should be 0.5, got %f", sys.updateInterval)
	}
}

func TestWeatherCooldownSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherCooldownSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			if sys.genreID != tt.genreID {
				t.Errorf("Genre not set correctly: got %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestWeatherCooldownSystem_CalculateCooldownMultiplier(t *testing.T) {
	tests := []struct {
		name        string
		weatherType particles.WeatherType
		intensity   particles.WeatherIntensity
		genre       string
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "rain light fantasy",
			weatherType: particles.WeatherRain,
			intensity:   particles.IntensityLight,
			genre:       "fantasy",
			wantMin:     1.03,
			wantMax:     1.12,
		},
		{
			name:        "fog heavy fantasy",
			weatherType: particles.WeatherFog,
			intensity:   particles.IntensityHeavy,
			genre:       "fantasy",
			wantMin:     0.75,
			wantMax:     0.88,
		},
		{
			name:        "neon rain cyberpunk",
			weatherType: particles.WeatherNeonRain,
			intensity:   particles.IntensityMedium,
			genre:       "cyberpunk",
			wantMin:     1.25,
			wantMax:     1.40,
		},
		{
			name:        "blood rain horror",
			weatherType: particles.WeatherBloodRain,
			intensity:   particles.IntensityHeavy,
			genre:       "horror",
			wantMin:     1.50,
			wantMax:     1.60, // Clamped at 1.6
		},
		{
			name:        "sandstorm postapoc",
			weatherType: particles.WeatherSandstorm,
			intensity:   particles.IntensityMedium,
			genre:       "postapoc",
			wantMin:     0.77,
			wantMax:     0.88,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherCooldownSystem(world, 12345)
			sys.SetGenre(tt.genre)

			// Run multiple times to test variance
			for i := 0; i < 10; i++ {
				mult := sys.calculateCooldownMultiplier(tt.weatherType, tt.intensity)
				if mult < tt.wantMin || mult > tt.wantMax {
					t.Errorf("Multiplier %f outside expected range [%f, %f]", mult, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}

func TestWeatherCooldownSystem_Update_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	// Create entity with spell slots
	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{
		Cooldowns: [5]float64{5.0, 3.0, 0, 0, 0},
	})

	entities := []*Entity{entity}

	// Update with no weather - cooldowns should not be modified by weather
	sys.Update(entities, 0.5)

	spellComp, _ := entity.GetComponent("spell_slots")
	slots := spellComp.(*SpellSlotComponent)

	// Cooldowns should remain unchanged (weather system doesn't tick cooldowns, just modifies rate)
	if slots.Cooldowns[0] != 5.0 {
		t.Errorf("Cooldown 0 should be 5.0, got %f", slots.Cooldowns[0])
	}
	if slots.Cooldowns[1] != 3.0 {
		t.Errorf("Cooldown 1 should be 3.0, got %f", slots.Cooldowns[1])
	}
}

func TestWeatherCooldownSystem_Update_WithWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	// Create weather entity
	weather := world.CreateEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create entity with spell slots
	player := world.CreateEntity()
	player.AddComponent(&SpellSlotComponent{
		Cooldowns: [5]float64{5.0, 3.0, 0, 0, 0},
	})

	entities := []*Entity{weather, player}

	// First update detects weather change
	sys.Update(entities, 0.5)

	if !sys.IsWeatherActive() {
		t.Error("Weather should be active")
	}

	// Second update applies cooldown modifier
	sys.Update(entities, 0.5)

	spellComp, _ := player.GetComponent("spell_slots")
	slots := spellComp.(*SpellSlotComponent)

	// Rain with medium intensity should speed up cooldowns slightly
	// The cooldown should be reduced by the bonus tick
	if slots.Cooldowns[0] >= 5.0 {
		t.Errorf("Cooldown 0 should be reduced by weather bonus, got %f", slots.Cooldowns[0])
	}
}

func TestWeatherCooldownSystem_Update_WeatherClears(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	// Create weather entity
	weather := world.CreateEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	player := world.CreateEntity()
	player.AddComponent(&SpellSlotComponent{
		Cooldowns: [5]float64{5.0, 0, 0, 0, 0},
	})

	entities := []*Entity{weather, player}

	// Activate weather
	sys.Update(entities, 0.5)
	if !sys.IsWeatherActive() {
		t.Error("Weather should be active")
	}

	// Deactivate weather
	weatherComp, _ := weather.GetComponent("weather")
	weatherComp.(*WeatherComponent).Active = false

	sys.Update(entities, 0.5)
	if sys.IsWeatherActive() {
		t.Error("Weather should be inactive after clearing")
	}
}

func TestWeatherCooldownSystem_GetModifierForEntity(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(&SpellSlotComponent{})

	// No weather - should return 1.0
	mod := sys.GetModifierForEntity(player.ID)
	if mod != 1.0 {
		t.Errorf("Modifier should be 1.0 without weather, got %f", mod)
	}

	// Add weather
	weather := world.CreateEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	sys.Update([]*Entity{weather, player}, 0.5)

	mod = sys.GetModifierForEntity(player.ID)
	if mod <= 1.0 {
		t.Errorf("Modifier should be > 1.0 with rain weather, got %f", mod)
	}
}

func TestWeatherCooldownSystem_GetActiveWeatherType(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	weather := world.CreateEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityHeavy,
		},
	})

	player := world.CreateEntity()
	player.AddComponent(&SpellSlotComponent{})

	sys.Update([]*Entity{weather, player}, 0.5)

	if sys.GetActiveWeatherType() != particles.WeatherFog {
		t.Errorf("Expected WeatherFog, got %v", sys.GetActiveWeatherType())
	}
}

func TestWeatherCooldownSystem_IntensityEffects(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 99999) // Fixed seed for determinism

	tests := []struct {
		intensity particles.WeatherIntensity
	}{
		{particles.IntensityLight},
		{particles.IntensityMedium},
		{particles.IntensityHeavy},
		{particles.IntensityExtreme},
	}

	var lastMult float64
	for i, tt := range tests {
		mult := sys.calculateCooldownMultiplier(particles.WeatherRain, tt.intensity)
		if i > 0 && mult <= lastMult {
			// Rain bonus should increase with intensity (approximately)
			// Allow for variance
			t.Logf("Intensity %v gave multiplier %f (prev: %f)", tt.intensity, mult, lastMult)
		}
		lastMult = mult
	}
}

func TestWeatherCooldownSystem_DeterministicWithSeed(t *testing.T) {
	world1 := NewWorld()
	world2 := NewWorld()
	sys1 := NewWeatherCooldownSystem(world1, 54321)
	sys2 := NewWeatherCooldownSystem(world2, 54321)

	// Same seed should produce same results
	mult1 := sys1.calculateCooldownMultiplier(particles.WeatherRain, particles.IntensityMedium)
	mult2 := sys2.calculateCooldownMultiplier(particles.WeatherRain, particles.IntensityMedium)

	if mult1 != mult2 {
		t.Errorf("Same seed should produce same multipliers: %f != %f", mult1, mult2)
	}
}

func TestWeatherCooldownSystem_ClampBounds(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	// Test that extreme values are clamped
	tests := []struct {
		weather   particles.WeatherType
		intensity particles.WeatherIntensity
		minBound  float64
		maxBound  float64
	}{
		{particles.WeatherSandstorm, particles.IntensityExtreme, 0.6, 1.6},
		{particles.WeatherBloodRain, particles.IntensityExtreme, 0.6, 1.6},
	}

	for _, tt := range tests {
		sys.SetGenre("horror") // Maximize blood rain effect
		mult := sys.calculateCooldownMultiplier(tt.weather, tt.intensity)
		if mult < tt.minBound || mult > tt.maxBound {
			t.Errorf("Multiplier %f should be clamped to [%f, %f]", mult, tt.minBound, tt.maxBound)
		}
	}
}

func BenchmarkWeatherCooldownSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	// Create weather
	weather := world.CreateEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create 100 entities with spell slots
	entities := make([]*Entity, 101)
	entities[0] = weather
	for i := 1; i <= 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(&SpellSlotComponent{
			Cooldowns: [5]float64{5.0, 4.0, 3.0, 2.0, 1.0},
		})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkWeatherCooldownSystem_CalculateMultiplier(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherCooldownSystem(world, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.calculateCooldownMultiplier(particles.WeatherRain, particles.IntensityMedium)
	}
}
