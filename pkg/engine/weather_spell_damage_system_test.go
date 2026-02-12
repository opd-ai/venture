package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherSpellDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherSpellDamageSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got '%s'", sys.genreID)
	}
}

func TestWeatherSpellDamageSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("Genre not set, expected 'cyberpunk', got '%s'", sys.genreID)
	}
}

func TestWeatherSpellDamageSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	// No weather entity, should return 1.0 for all elements
	tests := []magic.ElementType{
		magic.ElementFire,
		magic.ElementIce,
		magic.ElementLightning,
		magic.ElementEarth,
		magic.ElementWind,
		magic.ElementLight,
		magic.ElementDark,
		magic.ElementArcane,
	}

	for _, element := range tests {
		t.Run(element.String(), func(t *testing.T) {
			mod := sys.GetDamageModifier(element)
			if mod != 1.0 {
				t.Errorf("Expected 1.0 without weather, got %f for %s", mod, element.String())
			}
		})
	}
}

func TestWeatherSpellDamageSystem_RainModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	// Create weather entity with rain
	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: 0.5,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	tests := []struct {
		element  magic.ElementType
		expected float64
		desc     string
	}{
		{magic.ElementLightning, 1.25, "Lightning boosted in rain"},
		{magic.ElementFire, 0.75, "Fire weakened in rain"},
		{magic.ElementIce, 1.10, "Ice slightly boosted in rain"},
		{magic.ElementEarth, 1.0, "Earth unaffected by rain"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			mod := sys.GetDamageModifier(tt.element)
			if mod != tt.expected {
				t.Errorf("Expected %f for %s in rain, got %f", tt.expected, tt.element.String(), mod)
			}
		})
	}
}

func TestWeatherSpellDamageSystem_SnowModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherSnow,
		Intensity: 0.7,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	tests := []struct {
		element  magic.ElementType
		expected float64
	}{
		{magic.ElementIce, 1.30},
		{magic.ElementFire, 0.80},
		{magic.ElementWind, 1.15},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			mod := sys.GetDamageModifier(tt.element)
			if mod != tt.expected {
				t.Errorf("Expected %f for %s in snow, got %f", tt.expected, tt.element.String(), mod)
			}
		})
	}
}

func TestWeatherSpellDamageSystem_NeonRainCyberpunk(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)
	sys.SetGenre("cyberpunk")

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherNeonRain,
		Intensity: 0.8,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Lightning in neon rain with cyberpunk genre: 1.35 + 0.10 = 1.45
	mod := sys.GetDamageModifier(magic.ElementLightning)
	expected := 1.45
	if mod != expected {
		t.Errorf("Expected %f for lightning in cyberpunk neon rain, got %f", expected, mod)
	}
}

func TestWeatherSpellDamageSystem_BloodRainHorror(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)
	sys.SetGenre("horror")

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherBloodRain,
		Intensity: 0.9,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Dark in blood rain with horror genre: 1.35 + 0.15 = 1.50
	mod := sys.GetDamageModifier(magic.ElementDark)
	expected := 1.50
	if mod != expected {
		t.Errorf("Expected %f for dark in horror blood rain, got %f", expected, mod)
	}

	// Light in blood rain: 0.70
	lightMod := sys.GetDamageModifier(magic.ElementLight)
	if lightMod != 0.70 {
		t.Errorf("Expected 0.70 for light in blood rain, got %f", lightMod)
	}
}

func TestWeatherSpellDamageSystem_RadiationPostApoc(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)
	sys.SetGenre("postapoc")

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRadiation,
		Intensity: 0.6,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Arcane in radiation with postapoc: 1.30 + 0.10 = 1.40
	mod := sys.GetDamageModifier(magic.ElementArcane)
	expected := 1.40
	if mod != expected {
		t.Errorf("Expected %f for arcane in postapoc radiation, got %f", expected, mod)
	}
}

func TestWeatherSpellDamageSystem_CacheClearing(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: 0.5,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Get modifier (caches it)
	mod1 := sys.GetDamageModifier(magic.ElementLightning)

	// Clear cache
	sys.ClearCache()

	// Get again (should compute fresh)
	mod2 := sys.GetDamageModifier(magic.ElementLightning)

	if mod1 != mod2 {
		t.Errorf("Cache clearing changed result: %f vs %f", mod1, mod2)
	}
}

func TestWeatherSpellDamageSystem_InactiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: 0.5,
	})
	weatherComp.Active = false // Inactive
	weatherEntity.AddComponent(weatherComp)

	// Should return 1.0 for inactive weather
	mod := sys.GetDamageModifier(magic.ElementLightning)
	if mod != 1.0 {
		t.Errorf("Expected 1.0 for inactive weather, got %f", mod)
	}
}

func TestWeatherSpellDamageSystem_GetCurrentWeatherType(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	// No weather
	_, found := sys.GetCurrentWeatherType()
	if found {
		t.Error("Expected no weather type when none exists")
	}

	// Add active weather
	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherFog,
		Intensity: 0.4,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	weatherType, found := sys.GetCurrentWeatherType()
	if !found {
		t.Error("Expected to find weather type")
	}
	if weatherType != particles.WeatherFog {
		t.Errorf("Expected WeatherFog, got %v", weatherType)
	}
}

func TestWeatherSpellDamageSystem_DustModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherDust,
		Intensity: 0.6,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	tests := []struct {
		element  magic.ElementType
		expected float64
	}{
		{magic.ElementEarth, 1.25},
		{magic.ElementFire, 1.15},
		{magic.ElementLightning, 0.80},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			mod := sys.GetDamageModifier(tt.element)
			if mod != tt.expected {
				t.Errorf("Expected %f for %s in dust, got %f", tt.expected, tt.element.String(), mod)
			}
		})
	}
}

func TestWeatherSpellDamageSystem_SmogModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherSmog,
		Intensity: 0.5,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	tests := []struct {
		element  magic.ElementType
		expected float64
	}{
		{magic.ElementDark, 1.25},
		{magic.ElementFire, 1.10},
		{magic.ElementWind, 1.15},
		{magic.ElementLight, 0.80},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			mod := sys.GetDamageModifier(tt.element)
			if mod != tt.expected {
				t.Errorf("Expected %f for %s in smog, got %f", tt.expected, tt.element.String(), mod)
			}
		})
	}
}

func TestWeatherSpellDamageSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	// Update should not panic with empty entities
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.016)

	// Update with entities should also work
	entity := world.CreateEntity()
	sys.Update([]*Entity{entity}, 0.016)
}

func BenchmarkWeatherSpellDamageSystem_GetDamageModifier(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: 0.5,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetDamageModifier(magic.ElementLightning)
	}
}

func BenchmarkWeatherSpellDamageSystem_GetDamageModifierUncached(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherSpellDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: 0.5,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ClearCache()
		sys.GetDamageModifier(magic.ElementLightning)
	}
}
