package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherCompanionBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherCompanionBonusSystem returned nil")
	}

	if system.world != world {
		t.Error("world not set correctly")
	}

	if system.rng == nil {
		t.Error("rng not initialized")
	}

	if len(system.weatherBonuses) == 0 {
		t.Error("weather bonuses not initialized")
	}

	if len(system.genreMultipliers) == 0 {
		t.Error("genre multipliers not initialized")
	}
}

func TestWeatherCompanionBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)

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
			system.SetGenre(tt.genreID)
			if system.genre != tt.genreID {
				t.Errorf("SetGenre(%s) did not set genre correctly, got %s", tt.genreID, system.genre)
			}
		})
	}
}

func TestWeatherCompanionBonusComponent_Type(t *testing.T) {
	comp := &WeatherCompanionBonusComponent{
		AttackBonus:       1.2,
		DefenseBonus:      1.15,
		SpeedBonus:        1.1,
		WeatherType:       "rain",
		CompanionTypeName: "Elemental",
	}

	if comp.Type() != "weather_companion_bonus" {
		t.Errorf("Type() = %s, want weather_companion_bonus", comp.Type())
	}
}

func TestWeatherCompanionBonusSystem_Update_NoWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create companion entity without weather
	companion := NewEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeElemental,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{companion}

	// Run update multiple times to exceed throttle
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// No weather means no bonus should be applied
	if system.GetBonusCount() != 0 {
		t.Errorf("expected no bonuses without weather, got %d", system.GetBonusCount())
	}

	if system.IsWeatherActive() {
		t.Error("expected weather to be inactive")
	}
}

func TestWeatherCompanionBonusSystem_Update_WithRain(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather entity with rain
	weather := NewEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create elemental companion (should get bonuses in rain)
	companion := NewEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeElemental,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})

	entities := []*Entity{weather, companion}

	// Run update to exceed throttle
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// Check bonus was applied
	if !system.IsWeatherActive() {
		t.Error("expected weather to be active")
	}

	if system.GetActiveWeatherType() != particles.WeatherRain {
		t.Errorf("expected rain weather, got %v", system.GetActiveWeatherType())
	}

	// Check companion has weather bonus component
	bonusComp, ok := companion.GetComponent("weather_companion_bonus")
	if !ok {
		t.Fatal("expected weather_companion_bonus component")
	}

	bonus := bonusComp.(*WeatherCompanionBonusComponent)
	if bonus.AttackBonus <= 1.0 {
		t.Errorf("expected attack bonus > 1.0, got %f", bonus.AttackBonus)
	}

	if bonus.WeatherType != "Rain" {
		t.Errorf("expected weather type Rain, got %s", bonus.WeatherType)
	}
}

func TestWeatherCompanionBonusSystem_RobotPenaltyInRain(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather entity with rain
	weather := NewEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create robot companion (should get penalties in rain)
	companion := NewEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeRobot,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})

	entities := []*Entity{weather, companion}

	// Run update to exceed throttle
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// Check bonus was applied (should be penalty for robot)
	bonusComp, ok := companion.GetComponent("weather_companion_bonus")
	if !ok {
		t.Fatal("expected weather_companion_bonus component")
	}

	bonus := bonusComp.(*WeatherCompanionBonusComponent)
	if bonus.AttackBonus >= 1.0 {
		t.Errorf("expected robot attack penalty < 1.0, got %f", bonus.AttackBonus)
	}
}

func TestWeatherCompanionBonusSystem_GenreMultiplier(t *testing.T) {
	tests := []struct {
		name         string
		genre        string
		expectedMult float64
		checkHigher  bool // true = expect stronger effect than fantasy
	}{
		{"horror_stronger", "horror", 1.3, true},
		{"cyberpunk_weaker", "cyberpunk", 0.5, false},
		{"scifi_weaker", "scifi", 0.7, false},
		{"postapoc_stronger", "postapoc", 1.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewWeatherCompanionBonusSystem(world, 12345)
			system.SetGenre(tt.genre)

			mult, ok := system.genreMultipliers[tt.genre]
			if !ok {
				t.Errorf("genre %s not found in multipliers", tt.genre)
				return
			}

			if mult != tt.expectedMult {
				t.Errorf("genre %s multiplier = %f, want %f", tt.genre, mult, tt.expectedMult)
			}
		})
	}
}

func TestWeatherCompanionBonusSystem_IntensityModifier(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)

	tests := []struct {
		name      string
		intensity particles.WeatherIntensity
		expected  float64
	}{
		{"light", particles.IntensityLight, 0.5},
		{"medium", particles.IntensityMedium, 1.0},
		{"heavy", particles.IntensityHeavy, 1.3},
		{"extreme", particles.IntensityExtreme, 1.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := system.getIntensityModifier(tt.intensity)
			if mod != tt.expected {
				t.Errorf("getIntensityModifier(%v) = %f, want %f", tt.intensity, mod, tt.expected)
			}
		})
	}
}

func TestWeatherCompanionBonusSystem_WeatherClears(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather entity with rain
	weather := NewEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create companion
	companion := NewEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeElemental,
	})
	originalAttack := 10.0
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  originalAttack,
		Defense: 10.0,
		Speed:   10.0,
	})

	entities := []*Entity{weather, companion}

	// Apply weather bonus
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// Verify bonus was applied
	if !companion.HasComponent("weather_companion_bonus") {
		t.Fatal("expected bonus component to be added")
	}

	// Disable weather
	weatherComp, _ := weather.GetComponent("weather")
	weatherComp.(*WeatherComponent).Active = false

	// Run update again
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// Verify bonus was removed
	if companion.HasComponent("weather_companion_bonus") {
		t.Error("expected bonus component to be removed when weather clears")
	}

	// Check stats were restored
	statsComp, _ := companion.GetComponent("companionstats")
	stats := statsComp.(*CompanionStatsComponent)

	// Allow small floating point error
	if stats.Attack < originalAttack*0.99 || stats.Attack > originalAttack*1.01 {
		t.Errorf("expected attack to be restored to ~%f, got %f", originalAttack, stats.Attack)
	}
}

func TestWeatherCompanionBonusSystem_MultipleCompanionTypes(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather entity with fog (spirits benefit)
	weather := NewEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create spirit companion (should get big bonus)
	spirit := NewEntity()
	spirit.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeSpirit,
	})
	spirit.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})

	// Create robot companion (should get penalty)
	robot := NewEntity()
	robot.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeRobot,
	})
	robot.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})

	entities := []*Entity{weather, spirit, robot}

	// Run update
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// Check spirit got bonus
	spiritBonus, ok := spirit.GetComponent("weather_companion_bonus")
	if !ok {
		t.Fatal("expected spirit to have bonus")
	}
	if spiritBonus.(*WeatherCompanionBonusComponent).AttackBonus <= 1.0 {
		t.Error("expected spirit attack bonus > 1.0 in fog")
	}

	// Check robot got penalty
	robotBonus, ok := robot.GetComponent("weather_companion_bonus")
	if !ok {
		t.Fatal("expected robot to have bonus component")
	}
	if robotBonus.(*WeatherCompanionBonusComponent).AttackBonus >= 1.0 {
		t.Error("expected robot attack penalty < 1.0 in fog")
	}
}

func TestWeatherCompanionBonusSystem_HasActiveBonus(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather
	weather := NewEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create companion
	companion := NewEntity()
	companion.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeElemental,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})

	entities := []*Entity{weather, companion}

	// Before update, no bonus
	if system.HasActiveBonus(companion.ID) {
		t.Error("expected no active bonus before update")
	}

	// Run update
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.1)
	}

	// After update with weather, should have bonus
	if !system.HasActiveBonus(companion.ID) {
		t.Error("expected active bonus after update")
	}
}

func TestWeatherCompanionBonusSystem_CompanionTypeName(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)

	tests := []struct {
		compType CompanionType
		expected string
	}{
		{CompanionTypePet, "Pet"},
		{CompanionTypeSummon, "Summon"},
		{CompanionTypeHireling, "Hireling"},
		{CompanionTypeElemental, "Elemental"},
		{CompanionTypeUndead, "Undead"},
		{CompanionTypeRobot, "Robot"},
		{CompanionTypeSpirit, "Spirit"},
		{CompanionTypeInsect, "Insect"},
		{CompanionType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := system.companionTypeName(tt.compType)
			if name != tt.expected {
				t.Errorf("companionTypeName(%v) = %s, want %s", tt.compType, name, tt.expected)
			}
		})
	}
}

func TestWeatherCompanionBonusSystem_AllWeatherTypes(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)

	// Test that all weather types have at least one companion bonus defined
	weatherTypes := []particles.WeatherType{
		particles.WeatherRain,
		particles.WeatherSnow,
		particles.WeatherFog,
		particles.WeatherDust,
		particles.WeatherAsh,
		particles.WeatherNeonRain,
		particles.WeatherSmog,
		particles.WeatherRadiation,
		particles.WeatherBloodRain,
		particles.WeatherSandstorm,
	}

	for _, wt := range weatherTypes {
		t.Run(wt.String(), func(t *testing.T) {
			bonuses, ok := system.weatherBonuses[wt]
			if !ok {
				t.Errorf("weather type %v has no bonus definitions", wt)
				return
			}
			if len(bonuses) == 0 {
				t.Errorf("weather type %v has empty bonus map", wt)
			}
		})
	}
}

func BenchmarkWeatherCompanionBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create weather
	weather := NewEntity()
	weather.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create multiple companions
	entities := []*Entity{weather}
	for i := 0; i < 50; i++ {
		companion := NewEntity()
		companion.AddComponent(&CompanionComponent{
			CompanionType: CompanionType(i % 8),
		})
		companion.AddComponent(&CompanionStatsComponent{
			Attack:  10.0,
			Defense: 10.0,
			Speed:   10.0,
		})
		entities = append(entities, companion)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
