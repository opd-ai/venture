package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherMeleeDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherMeleeDamageSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set correctly")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre should be fantasy, got %s", sys.genreID)
	}
	if sys.modCache == nil {
		t.Error("Modifier cache not initialized")
	}
}

func TestWeatherMeleeDamageSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("Genre should be cyberpunk, got %s", sys.genreID)
	}
}

func TestWeatherMeleeDamageSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	// Without weather, all modifiers should be 1.0
	damageTypes := []combat.DamageType{
		combat.DamagePhysical,
		combat.DamageFire,
		combat.DamageIce,
		combat.DamageLightning,
		combat.DamagePoison,
	}

	for _, dt := range damageTypes {
		mod := sys.GetDamageModifier(dt)
		if mod != 1.0 {
			t.Errorf("No weather: %s modifier should be 1.0, got %f", dt.String(), mod)
		}
	}
}

func TestWeatherMeleeDamageSystem_RainModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	// Add rain weather entity
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherRain,
		},
	})

	tests := []struct {
		damageType combat.DamageType
		expected   float64
	}{
		{combat.DamagePhysical, 0.90},  // Slippery grip
		{combat.DamageLightning, 1.20}, // Rain conducts
		{combat.DamageFire, 0.85},      // Fire dampened
		{combat.DamageIce, 1.0},        // No modifier
		{combat.DamagePoison, 1.0},     // No modifier
	}

	for _, tt := range tests {
		mod := sys.GetDamageModifier(tt.damageType)
		if mod != tt.expected {
			t.Errorf("Rain + %s: expected %f, got %f", tt.damageType.String(), tt.expected, mod)
		}
	}
}

func TestWeatherMeleeDamageSystem_SnowModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherSnow,
		},
	})

	tests := []struct {
		damageType combat.DamageType
		expected   float64
	}{
		{combat.DamagePhysical, 0.85}, // Stiff joints
		{combat.DamageIce, 1.25},      // Cold empowers ice
		{combat.DamageFire, 0.90},     // Fire struggles
	}

	for _, tt := range tests {
		mod := sys.GetDamageModifier(tt.damageType)
		if mod != tt.expected {
			t.Errorf("Snow + %s: expected %f, got %f", tt.damageType.String(), tt.expected, mod)
		}
	}
}

func TestWeatherMeleeDamageSystem_NeonRainCyberpunk(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)
	sys.SetGenre("cyberpunk")

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherNeonRain,
		},
	})

	// Neon rain + lightning in cyberpunk = 1.30 base + 0.10 genre bonus = 1.40
	lightningMod := sys.GetDamageModifier(combat.DamageLightning)
	if lightningMod != 1.40 {
		t.Errorf("Cyberpunk neon rain + lightning: expected 1.40, got %f", lightningMod)
	}
}

func TestWeatherMeleeDamageSystem_BloodRainHorror(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)
	sys.SetGenre("horror")

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherBloodRain,
		},
	})

	// Blood rain + physical in horror = 1.20 base + 0.10 genre bonus = 1.30
	physicalMod := sys.GetDamageModifier(combat.DamagePhysical)
	if physicalMod != 1.30 {
		t.Errorf("Horror blood rain + physical: expected 1.30, got %f", physicalMod)
	}
}

func TestWeatherMeleeDamageSystem_RadiationPostApoc(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)
	sys.SetGenre("postapoc")

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherRadiation,
		},
	})

	// Radiation + physical in postapoc = 1.15 base + 0.10 genre bonus = 1.25
	physicalMod := sys.GetDamageModifier(combat.DamagePhysical)
	if physicalMod != 1.25 {
		t.Errorf("Postapoc radiation + physical: expected 1.25, got %f", physicalMod)
	}
}

func TestWeatherMeleeDamageSystem_CacheClearing(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherRain,
		},
	})

	// First call caches result
	mod1 := sys.GetDamageModifier(combat.DamagePhysical)
	if mod1 != 0.90 {
		t.Errorf("First call: expected 0.90, got %f", mod1)
	}

	// Clear cache
	sys.ClearCache()
	if len(sys.modCache) != 0 {
		t.Error("Cache should be empty after ClearCache")
	}

	// Second call should recompute
	mod2 := sys.GetDamageModifier(combat.DamagePhysical)
	if mod2 != 0.90 {
		t.Errorf("After clear: expected 0.90, got %f", mod2)
	}
}

func TestWeatherMeleeDamageSystem_InactiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: false, // Inactive
		Config: particles.WeatherConfig{
			Type: particles.WeatherRain,
		},
	})

	// Inactive weather should return 1.0
	mod := sys.GetDamageModifier(combat.DamagePhysical)
	if mod != 1.0 {
		t.Errorf("Inactive weather: expected 1.0, got %f", mod)
	}
}

func TestWeatherMeleeDamageSystem_GetCurrentWeatherType(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	// No weather
	_, ok := sys.GetCurrentWeatherType()
	if ok {
		t.Error("Should return false when no weather")
	}

	// Add weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherFog,
		},
	})

	wtype, ok := sys.GetCurrentWeatherType()
	if !ok {
		t.Error("Should return true when weather exists")
	}
	if wtype != particles.WeatherFog {
		t.Errorf("Weather type should be Fog, got %v", wtype)
	}
}

func TestWeatherMeleeDamageSystem_FogModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherFog,
		},
	})

	tests := []struct {
		damageType combat.DamageType
		expected   float64
	}{
		{combat.DamagePhysical, 1.10}, // Concealment for sneak attacks
		{combat.DamagePoison, 1.15},   // Poison spreads in moist air
	}

	for _, tt := range tests {
		mod := sys.GetDamageModifier(tt.damageType)
		if mod != tt.expected {
			t.Errorf("Fog + %s: expected %f, got %f", tt.damageType.String(), tt.expected, mod)
		}
	}
}

func TestWeatherMeleeDamageSystem_DustModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherDust,
		},
	})

	tests := []struct {
		damageType combat.DamageType
		expected   float64
	}{
		{combat.DamagePhysical, 1.10},  // Abrasive particles
		{combat.DamageFire, 1.15},      // Dust ignites
		{combat.DamageLightning, 0.85}, // Static interference
	}

	for _, tt := range tests {
		mod := sys.GetDamageModifier(tt.damageType)
		if mod != tt.expected {
			t.Errorf("Dust + %s: expected %f, got %f", tt.damageType.String(), tt.expected, mod)
		}
	}
}

func TestWeatherMeleeDamageSystem_SmogModifiers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherSmog,
		},
	})

	tests := []struct {
		damageType combat.DamageType
		expected   float64
	}{
		{combat.DamagePhysical, 1.05}, // Obscured vision
		{combat.DamageFire, 1.10},     // Smog is flammable
		{combat.DamagePoison, 1.25},   // Toxic atmosphere
	}

	for _, tt := range tests {
		mod := sys.GetDamageModifier(tt.damageType)
		if mod != tt.expected {
			t.Errorf("Smog + %s: expected %f, got %f", tt.damageType.String(), tt.expected, mod)
		}
	}
}

func TestWeatherMeleeDamageSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	// Update is a no-op for this query-based system
	entities := world.GetEntities()
	sys.Update(entities, 0.016) // Should not panic
}

func TestWeatherMeleeDamageSystem_NilWorld(t *testing.T) {
	sys := NewWeatherMeleeDamageSystem(nil, 12345)

	// Should not panic and return 1.0
	mod := sys.GetDamageModifier(combat.DamagePhysical)
	if mod != 1.0 {
		t.Errorf("Nil world: expected 1.0, got %f", mod)
	}

	_, ok := sys.GetCurrentWeatherType()
	if ok {
		t.Error("Nil world should return false for GetCurrentWeatherType")
	}
}

func BenchmarkWeatherMeleeDamageSystem_GetDamageModifier(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherRain,
		},
	})

	damageTypes := []combat.DamageType{
		combat.DamagePhysical,
		combat.DamageFire,
		combat.DamageLightning,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt := damageTypes[i%len(damageTypes)]
		sys.GetDamageModifier(dt)
	}
}

func BenchmarkWeatherMeleeDamageSystem_CacheHit(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherMeleeDamageSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type: particles.WeatherRain,
		},
	})

	// Prime the cache
	sys.GetDamageModifier(combat.DamagePhysical)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetDamageModifier(combat.DamagePhysical)
	}
}
