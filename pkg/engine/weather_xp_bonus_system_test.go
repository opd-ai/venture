package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// TestNewWeatherXPBonusSystem verifies system creation.
func TestNewWeatherXPBonusSystem(t *testing.T) {
	world := NewWorld()

	sys := NewWeatherXPBonusSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewWeatherXPBonusSystem returned nil")
	}

	if sys.world != world {
		t.Error("World not set correctly")
	}

	if sys.rng == nil {
		t.Error("RNG not initialized")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %q, want 'fantasy'", sys.genreID)
	}
}

// TestWeatherXPBonusSystem_SetGenre verifies genre setting.
func TestWeatherXPBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

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
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("Genre = %q, want %q", sys.genreID, tt.genre)
			}
		})
	}
}

// TestWeatherXPBonusSystem_Update verifies update processes weather.
func TestWeatherXPBonusSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create weather entity with active weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	entities := []*Entity{player, weatherEntity}

	// First update should detect weather
	sys.Update(entities, 0.6) // Exceed update interval

	if !sys.IsWeatherActive() {
		t.Error("Weather should be detected as active")
	}

	if sys.GetActiveWeatherType() != particles.WeatherRain {
		t.Errorf("Weather type = %v, want Rain", sys.GetActiveWeatherType())
	}

	// Player should have multiplier now
	mult := sys.GetXPMultiplier(player.ID)
	if mult <= 1.0 {
		t.Errorf("Player multiplier = %f, want > 1.0", mult)
	}
}

// TestWeatherXPBonusSystem_NoWeather verifies no bonus without weather.
func TestWeatherXPBonusSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())

	entities := []*Entity{player}
	sys.Update(entities, 0.6)

	if sys.IsWeatherActive() {
		t.Error("Weather should not be active without weather entity")
	}

	mult := sys.GetXPMultiplier(player.ID)
	if mult != 1.0 {
		t.Errorf("Multiplier without weather = %f, want 1.0", mult)
	}
}

// TestWeatherXPBonusSystem_ApplyBonusToXP verifies XP bonus calculation.
func TestWeatherXPBonusSystem_ApplyBonusToXP(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())

	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityHeavy,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	entities := []*Entity{player, weatherEntity}
	sys.Update(entities, 0.6)

	baseXP := 100
	totalXP := sys.ApplyBonusToXP(player.ID, baseXP)

	if totalXP <= baseXP {
		t.Errorf("Total XP = %d, want > %d", totalXP, baseXP)
	}

	// Bonus should be reasonable (between 1% and 50%)
	bonus := totalXP - baseXP
	if bonus < 1 || bonus > 50 {
		t.Errorf("Bonus XP = %d, want between 1 and 50", bonus)
	}
}

// TestWeatherXPBonusSystem_GenreSpecificBonuses verifies genre affects bonuses.
func TestWeatherXPBonusSystem_GenreSpecificBonuses(t *testing.T) {
	tests := []struct {
		genre       string
		weatherType particles.WeatherType
		wantBonus   bool
	}{
		{"fantasy", particles.WeatherFog, true},      // Stealth bonus in fantasy
		{"horror", particles.WeatherBloodRain, true}, // Dark ritual bonus
		{"cyberpunk", particles.WeatherNeonRain, true},
		{"postapoc", particles.WeatherRadiation, true},
		{"scifi", particles.WeatherRadiation, true},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.weatherType.String(), func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherXPBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)

			player := world.CreateEntity()
			player.AddComponent(NewStubInput())

			weatherEntity := world.CreateEntity()
			weatherConfig := particles.WeatherConfig{
				Type:      tt.weatherType,
				Intensity: particles.IntensityMedium,
			}
			weatherComp := NewWeatherComponent(weatherConfig)
			weatherComp.Active = true
			weatherEntity.AddComponent(weatherComp)

			entities := []*Entity{player, weatherEntity}
			sys.Update(entities, 0.6)

			mult := sys.GetXPMultiplier(player.ID)
			hasBonus := mult > 1.0

			if hasBonus != tt.wantBonus {
				t.Errorf("Genre %s weather %s: hasBonus = %v, want %v (mult=%f)",
					tt.genre, tt.weatherType.String(), hasBonus, tt.wantBonus, mult)
			}
		})
	}
}

// TestWeatherXPBonusSystem_IntensityScaling verifies intensity affects bonus.
func TestWeatherXPBonusSystem_IntensityScaling(t *testing.T) {
	world := NewWorld()

	intensities := []particles.WeatherIntensity{
		particles.IntensityLight,
		particles.IntensityMedium,
		particles.IntensityHeavy,
		particles.IntensityExtreme,
	}

	var prevMult float64 = 0.0
	for _, intensity := range intensities {
		t.Run(intensity.String(), func(t *testing.T) {
			sys := NewWeatherXPBonusSystem(world, 99999) // Fixed seed for determinism

			player := world.CreateEntity()
			player.AddComponent(NewStubInput())

			weatherEntity := world.CreateEntity()
			weatherConfig := particles.WeatherConfig{
				Type:      particles.WeatherFog, // Good bonus in fantasy
				Intensity: intensity,
			}
			weatherComp := NewWeatherComponent(weatherConfig)
			weatherComp.Active = true
			weatherEntity.AddComponent(weatherComp)

			entities := []*Entity{player, weatherEntity}
			sys.Update(entities, 0.6)

			mult := sys.GetXPMultiplier(player.ID)

			// Higher intensity should give higher or equal multiplier
			if mult < prevMult*0.9 { // Allow 10% tolerance for variance
				t.Errorf("Intensity %s mult = %f, want >= %f", intensity.String(), mult, prevMult)
			}
			prevMult = mult
		})
	}
}

// TestWeatherXPBonusSystem_NonPlayerEntities verifies non-players don't get bonuses.
func TestWeatherXPBonusSystem_NonPlayerEntities(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

	// Create NPC entity (no input component)
	npc := world.CreateEntity()
	npc.AddComponent(&PositionComponent{X: 100, Y: 100})
	npc.AddComponent(&HealthComponent{Current: 100, Max: 100})

	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityHeavy,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	entities := []*Entity{npc, weatherEntity}
	sys.Update(entities, 0.6)

	mult := sys.GetXPMultiplier(npc.ID)
	if mult != 1.0 {
		t.Errorf("NPC multiplier = %f, want 1.0 (NPCs should not get bonuses)", mult)
	}
}

// TestWeatherXPBonusSystem_InactiveWeather verifies inactive weather gives no bonus.
func TestWeatherXPBonusSystem_InactiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())

	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityHeavy,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = false // Weather not active
	weatherEntity.AddComponent(weatherComp)

	entities := []*Entity{player, weatherEntity}
	sys.Update(entities, 0.6)

	if sys.IsWeatherActive() {
		t.Error("Inactive weather should not be detected")
	}

	mult := sys.GetXPMultiplier(player.ID)
	if mult != 1.0 {
		t.Errorf("Multiplier with inactive weather = %f, want 1.0", mult)
	}
}

// TestWeatherXPBonusSystem_Deterministic verifies same seed gives same results.
func TestWeatherXPBonusSystem_Deterministic(t *testing.T) {
	var results []float64
	for i := 0; i < 3; i++ {
		world := NewWorld()
		sys := NewWeatherXPBonusSystem(world, 54321)

		player := world.CreateEntity()
		player.AddComponent(NewStubInput())

		weatherEntity := world.CreateEntity()
		weatherConfig := particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		}
		weatherComp := NewWeatherComponent(weatherConfig)
		weatherComp.Active = true
		weatherEntity.AddComponent(weatherComp)

		entities := []*Entity{player, weatherEntity}
		sys.Update(entities, 0.6)

		results = append(results, sys.GetXPMultiplier(player.ID))
	}

	// All runs with same seed should give same result
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Run %d multiplier = %f, want %f (determinism)", i, results[i], results[0])
		}
	}
}

// TestWeatherXPBonusSystem_MultiplierBounds verifies multipliers stay in bounds.
func TestWeatherXPBonusSystem_MultiplierBounds(t *testing.T) {
	allWeatherTypes := []particles.WeatherType{
		particles.WeatherRain,
		particles.WeatherSnow,
		particles.WeatherFog,
		particles.WeatherDust,
		particles.WeatherAsh,
		particles.WeatherNeonRain,
		particles.WeatherSmog,
		particles.WeatherRadiation,
		particles.WeatherSandstorm,
		particles.WeatherBloodRain,
	}

	for _, wt := range allWeatherTypes {
		t.Run(wt.String(), func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherXPBonusSystem(world, 12345)

			player := world.CreateEntity()
			player.AddComponent(NewStubInput())

			weatherEntity := world.CreateEntity()
			weatherConfig := particles.WeatherConfig{
				Type:      wt,
				Intensity: particles.IntensityExtreme,
			}
			weatherComp := NewWeatherComponent(weatherConfig)
			weatherComp.Active = true
			weatherEntity.AddComponent(weatherComp)

			entities := []*Entity{player, weatherEntity}
			sys.Update(entities, 0.6)

			mult := sys.GetXPMultiplier(player.ID)
			if mult < 1.0 || mult > 1.5 {
				t.Errorf("Multiplier for %s = %f, want [1.0, 1.5]", wt.String(), mult)
			}
		})
	}
}

// BenchmarkWeatherXPBonusSystem_Update benchmarks the update cycle.
func BenchmarkWeatherXPBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherXPBonusSystem(world, 12345)

	// Create 100 player entities
	entities := make([]*Entity, 0, 101)
	for i := 0; i < 100; i++ {
		player := world.CreateEntity()
		player.AddComponent(NewStubInput())
		player.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entities = append(entities, player)
	}

	// Add weather entity
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)
	entities = append(entities, weatherEntity)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.6)
	}
}
