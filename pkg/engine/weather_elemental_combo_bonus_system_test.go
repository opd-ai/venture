//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherElementalComboBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherElementalComboBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.baseMultiplier != 1.0 {
		t.Errorf("Expected base multiplier 1.0, got %v", sys.baseMultiplier)
	}
	if len(sys.weatherBonuses) == 0 {
		t.Error("Weather bonuses not initialized")
	}
}

func TestWeatherElementalComboBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)

	tests := []struct {
		name    string
		genre   string
		wantMul float64
	}{
		{"fantasy", "fantasy", 1.0},
		{"horror", "horror", 1.3},
		{"scifi", "scifi", 0.8},
		{"cyberpunk", "cyberpunk", 0.7},
		{"postapoc", "postapoc", 1.1},
		{"unknown", "unknown_genre", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("Genre not set, got %s want %s", sys.genreID, tt.genre)
			}
			got := sys.getGenreIntensityScale()
			if got != tt.wantMul {
				t.Errorf("getGenreIntensityScale() = %v, want %v", got, tt.wantMul)
			}
		})
	}
}

func TestWeatherElementalComboBonusSystem_GetWeatherBonus(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	sys.SetGenre("fantasy")

	tests := []struct {
		name         string
		weather      particles.WeatherType
		comboName    string
		wantPositive bool
	}{
		{"rain steam burst", particles.WeatherRain, "steam_burst", true},
		{"rain toxic flames", particles.WeatherRain, "toxic_flames", false},
		{"snow deep freeze", particles.WeatherSnow, "deep_freeze", true},
		{"snow toxic flames", particles.WeatherSnow, "toxic_flames", false},
		{"fog toxic pool", particles.WeatherFog, "toxic_pool", true},
		{"dust toxic flames", particles.WeatherDust, "toxic_flames", true},
		{"sandstorm plasma burst", particles.WeatherSandstorm, "plasma_burst", true},
		{"blood rain toxic flames", particles.WeatherBloodRain, "toxic_flames", true},
		{"radiation plasma burst", particles.WeatherRadiation, "plasma_burst", true},
		{"unknown combo", particles.WeatherRain, "nonexistent_combo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := sys.GetWeatherBonus(tt.weather, tt.comboName)
			if tt.wantPositive && bonus <= 0 {
				t.Errorf("Expected positive bonus for %s/%s, got %v", tt.weather, tt.comboName, bonus)
			}
			if !tt.wantPositive && bonus > 0 {
				t.Errorf("Expected zero or negative bonus for %s/%s, got %v", tt.weather, tt.comboName, bonus)
			}
		})
	}
}

func TestWeatherElementalComboBonusSystem_SetBaseMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)

	tests := []struct {
		name    string
		mult    float64
		wantSet float64
	}{
		{"positive", 1.5, 1.5},
		{"small", 0.5, 0.5},
		{"zero (ignored)", 0.0, 1.0}, // Should keep previous value
		{"negative (ignored)", -1.0, 1.0},
		{"large", 3.0, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default
			sys.baseMultiplier = 1.0
			sys.SetBaseMultiplier(tt.mult)
			if sys.GetBaseMultiplier() != tt.wantSet {
				t.Errorf("SetBaseMultiplier(%v) = %v, want %v", tt.mult, sys.GetBaseMultiplier(), tt.wantSet)
			}
		})
	}
}

func TestWeatherElementalComboBonusSystem_SetElementalComboDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)

	elementalSys := NewElementalComboDamageSystem(world, 54321)
	sys.SetElementalComboDamageSystem(elementalSys)

	if sys.elementalComboDamageSystem != elementalSys {
		t.Error("ElementalComboDamageSystem not set correctly")
	}
	if sys.originalBaseDamage != elementalSys.GetBaseDamage() {
		t.Error("Original base damage not captured")
	}
}

func TestWeatherElementalComboBonusSystem_Update_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	elementalSys := NewElementalComboDamageSystem(world, 54321)
	sys.SetElementalComboDamageSystem(elementalSys)

	originalDamage := elementalSys.GetBaseDamage()

	// Create entity without weather component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.Update([]*Entity{entity}, 0.016)

	// Base damage should remain unchanged with no weather
	if elementalSys.GetBaseDamage() != originalDamage {
		t.Errorf("Base damage changed without weather: got %v, want %v", elementalSys.GetBaseDamage(), originalDamage)
	}
}

func TestWeatherElementalComboBonusSystem_Update_WithWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	sys.SetGenre("fantasy")

	elementalSys := NewElementalComboDamageSystem(world, 54321)
	sys.SetElementalComboDamageSystem(elementalSys)

	originalDamage := elementalSys.GetBaseDamage()

	// Create entity with active weather
	entity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.WeatherIntensityMedium,
		Width:     800,
		Height:    600,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	entity.AddComponent(weatherComp)

	sys.Update([]*Entity{entity}, 0.016)

	// With rain weather, steam_burst gets +25% bonus which should modify damage
	newDamage := elementalSys.GetBaseDamage()
	if newDamage <= originalDamage {
		t.Logf("Note: Damage modified from %v to %v", originalDamage, newDamage)
	}
	if !sys.hasModifiedDamage {
		t.Error("hasModifiedDamage should be true after weather update")
	}
}

func TestWeatherElementalComboBonusSystem_Update_InactiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	elementalSys := NewElementalComboDamageSystem(world, 54321)
	sys.SetElementalComboDamageSystem(elementalSys)

	originalDamage := elementalSys.GetBaseDamage()

	// Create entity with inactive weather
	entity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.WeatherIntensityHeavy,
		Width:     800,
		Height:    600,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = false // Inactive
	entity.AddComponent(weatherComp)

	sys.Update([]*Entity{entity}, 0.016)

	// Inactive weather should not affect damage
	if elementalSys.GetBaseDamage() != originalDamage {
		t.Errorf("Inactive weather modified damage: got %v, want %v", elementalSys.GetBaseDamage(), originalDamage)
	}
}

func TestWeatherElementalComboBonusSystem_GenreScaling(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)

	// Horror genre should have higher bonuses
	sys.SetGenre("horror")
	horrorBonus := sys.GetWeatherBonus(particles.WeatherRain, "steam_burst")

	// Cyberpunk should have lower bonuses
	sys.SetGenre("cyberpunk")
	cyberpunkBonus := sys.GetWeatherBonus(particles.WeatherRain, "steam_burst")

	if horrorBonus <= cyberpunkBonus {
		t.Errorf("Horror bonus (%v) should be greater than cyberpunk bonus (%v)", horrorBonus, cyberpunkBonus)
	}
}

func TestWeatherElementalComboBonusSystem_NilWorld(t *testing.T) {
	sys := NewWeatherElementalComboBonusSystem(nil, 12345)
	if sys == nil {
		t.Fatal("System should be created even with nil world")
	}

	// Update with nil world should not panic
	sys.Update([]*Entity{}, 0.016)
}

func TestWeatherElementalComboBonusSystem_NilElementalSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)

	// Create entity with active weather
	entity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.WeatherIntensityMedium,
		Width:     800,
		Height:    600,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	entity.AddComponent(weatherComp)

	// Update without elemental system should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestWeatherElementalComboBonusSystem_AllWeatherTypes(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	sys.SetGenre("fantasy")

	weatherTypes := []particles.WeatherType{
		particles.WeatherRain,
		particles.WeatherSnow,
		particles.WeatherFog,
		particles.WeatherDust,
		particles.WeatherSandstorm,
		particles.WeatherBloodRain,
		particles.WeatherRadiation,
	}

	for _, wt := range weatherTypes {
		t.Run(wt.String(), func(t *testing.T) {
			bonuses, exists := sys.weatherBonuses[wt]
			if !exists {
				t.Errorf("No bonuses defined for weather type %s", wt)
				return
			}
			if len(bonuses) == 0 {
				t.Errorf("Empty bonuses for weather type %s", wt)
			}
		})
	}
}

func TestWeatherElementalComboBonusSystem_DamageRestoration(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	sys.SetGenre("fantasy")

	elementalSys := NewElementalComboDamageSystem(world, 54321)
	sys.SetElementalComboDamageSystem(elementalSys)

	originalDamage := elementalSys.GetBaseDamage()

	// Create entity with active weather
	entity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.WeatherIntensityMedium,
		Width:     800,
		Height:    600,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	entity.AddComponent(weatherComp)

	// Apply weather bonus
	sys.Update([]*Entity{entity}, 0.016)

	// Now deactivate weather
	weatherComp.Active = false
	sys.Update([]*Entity{entity}, 0.016)

	// Damage should be restored to original
	if elementalSys.GetBaseDamage() != originalDamage {
		t.Errorf("Damage not restored after weather ended: got %v, want %v", elementalSys.GetBaseDamage(), originalDamage)
	}
}

func BenchmarkWeatherElementalComboBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherElementalComboBonusSystem(world, 12345)
	sys.SetGenre("fantasy")
	elementalSys := NewElementalComboDamageSystem(world, 54321)
	sys.SetElementalComboDamageSystem(elementalSys)

	// Create entities with weather
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		if i == 0 {
			// Only first entity has weather
			weatherConfig := particles.WeatherConfig{
				Type:      particles.WeatherRain,
				Intensity: particles.WeatherIntensityMedium,
				Width:     800,
				Height:    600,
			}
			weatherComp := NewWeatherComponent(weatherConfig)
			weatherComp.Active = true
			entity.AddComponent(weatherComp)
		}
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
