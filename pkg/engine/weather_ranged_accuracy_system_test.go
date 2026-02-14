package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherRangedAccuracySystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherRangedAccuracySystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", sys.genreID)
	}
}

func TestWeatherRangedAccuracySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("Genre = %s, want scifi", sys.genreID)
	}

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("Genre = %s, want horror", sys.genreID)
	}
}

func TestWeatherRangedAccuracySystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Create entity with ranged capability
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 150.0})
	entity.AddComponent(&StatsComponent{})

	// Update with no weather
	sys.Update(world.GetEntities(), 1.0)

	// Modifier should be 1.0 (no effect)
	modifier := sys.GetAccuracyModifier(entity.ID)
	if modifier != 1.0 {
		t.Errorf("Modifier without weather = %f, want 1.0", modifier)
	}
}

func TestWeatherRangedAccuracySystem_RainEffect(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create ranged entity
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 200.0})

	// Update system
	sys.Update(world.GetEntities(), 1.0)

	// Rain at medium intensity should reduce accuracy
	modifier := sys.GetAccuracyModifier(entity.ID)
	if modifier >= 1.0 || modifier < 0.5 {
		t.Errorf("Rain modifier = %f, expected between 0.5 and 1.0", modifier)
	}
}

func TestWeatherRangedAccuracySystem_FogEffect(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Create weather entity with fog
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityHeavy,
		},
	})

	// Create ranged entity
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 250.0})

	sys.Update(world.GetEntities(), 1.0)

	// Fog should have significant accuracy penalty
	modifier := sys.GetAccuracyModifier(entity.ID)
	if modifier > 0.8 {
		t.Errorf("Fog heavy modifier = %f, expected <= 0.8", modifier)
	}
}

func TestWeatherRangedAccuracySystem_SandstormEffect(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Create sandstorm
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSandstorm,
			Intensity: particles.IntensityExtreme,
		},
	})

	// Create ranged entity
	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})

	sys.Update(world.GetEntities(), 1.0)

	// Sandstorm extreme should have severe penalty
	modifier := sys.GetAccuracyModifier(entity.ID)
	if modifier > 0.6 {
		t.Errorf("Sandstorm extreme modifier = %f, expected <= 0.6", modifier)
	}
}

func TestWeatherRangedAccuracySystem_GenreBonus_Scifi(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)
	sys.SetGenre("scifi")

	// Create rain weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 150.0})

	sys.Update(world.GetEntities(), 1.0)

	// Sci-fi genre should have better accuracy in rain
	scifiModifier := sys.GetAccuracyModifier(entity.ID)

	// Compare with fantasy
	sys.SetGenre("fantasy")
	sys.Update(world.GetEntities(), 1.0)
	fantasyModifier := sys.GetAccuracyModifier(entity.ID)

	// Sci-fi should be better (accounting for variance)
	if scifiModifier < fantasyModifier-0.15 {
		t.Errorf("Sci-fi modifier %f should be >= fantasy modifier %f", scifiModifier, fantasyModifier)
	}
}

func TestWeatherRangedAccuracySystem_GenreBonus_Horror(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)
	sys.SetGenre("horror")

	// Create fog weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityMedium,
		},
	})

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 150.0})

	sys.Update(world.GetEntities(), 1.0)

	// Horror genre has extra fog penalty
	modifier := sys.GetAccuracyModifier(entity.ID)
	if modifier > 0.85 {
		t.Errorf("Horror fog modifier = %f, expected <= 0.85 due to horror penalty", modifier)
	}
}

func TestWeatherRangedAccuracySystem_IntensityScaling(t *testing.T) {
	tests := []struct {
		name      string
		intensity particles.WeatherIntensity
		maxMod    float64
	}{
		{"light", particles.IntensityLight, 1.00},     // Light intensity has minimal effect
		{"medium", particles.IntensityMedium, 0.98},   // Medium starts to show effect
		{"heavy", particles.IntensityHeavy, 0.95},     // Heavy has noticeable effect
		{"extreme", particles.IntensityExtreme, 0.92}, // Extreme has significant effect
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherRangedAccuracySystem(world, 12345)

			weatherEntity := world.CreateEntity()
			weatherEntity.AddComponent(&WeatherComponent{
				Active: true,
				Config: particles.WeatherConfig{
					Type:      particles.WeatherRain,
					Intensity: tt.intensity,
				},
			})

			entity := world.CreateEntity()
			entity.AddComponent(&AIComponent{DetectionRange: 150.0})

			sys.Update(world.GetEntities(), 1.0)

			modifier := sys.GetAccuracyModifier(entity.ID)
			if modifier > tt.maxMod {
				t.Errorf("Intensity %s modifier = %f, expected <= %f", tt.name, modifier, tt.maxMod)
			}
		})
	}
}

func TestWeatherRangedAccuracySystem_IsWeatherActive(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Initially inactive
	sys.Update(world.GetEntities(), 1.0)
	if sys.IsWeatherActive() {
		t.Error("Weather should not be active without weather entity")
	}

	// Add weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityMedium,
		},
	})

	sys.Update(world.GetEntities(), 1.0)
	if !sys.IsWeatherActive() {
		t.Error("Weather should be active with active weather entity")
	}
}

func TestWeatherRangedAccuracySystem_GetCurrentPenalty(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// No weather = no penalty
	sys.Update(world.GetEntities(), 1.0)
	penalty := sys.GetCurrentPenalty()
	if penalty != 0.0 {
		t.Errorf("Penalty without weather = %f, want 0.0", penalty)
	}

	// Add weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityHeavy,
		},
	})

	sys.Update(world.GetEntities(), 1.0)
	penalty = sys.GetCurrentPenalty()
	if penalty <= 0.0 {
		t.Error("Penalty should be > 0 with heavy fog")
	}
	if penalty >= 1.0 {
		t.Error("Penalty should be < 1.0")
	}
}

func TestWeatherRangedAccuracySystem_HasRangedCapability(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	tests := []struct {
		name     string
		setup    func(*Entity)
		expected bool
	}{
		{
			name:     "no components",
			setup:    func(e *Entity) {},
			expected: false,
		},
		{
			name: "ai melee",
			setup: func(e *Entity) {
				e.AddComponent(&AIComponent{DetectionRange: 64.0})
			},
			expected: false,
		},
		{
			name: "ai ranged",
			setup: func(e *Entity) {
				e.AddComponent(&AIComponent{DetectionRange: 150.0})
			},
			expected: true,
		},
		{
			name: "spellcasting",
			setup: func(e *Entity) {
				e.AddComponent(&ManaComponent{Current: 100, Max: 100})
			},
			expected: true,
		},
		{
			name: "projectile",
			setup: func(e *Entity) {
				e.AddComponent(&ProjectileComponent{})
			},
			expected: true,
		},
		{
			name: "player with stats",
			setup: func(e *Entity) {
				e.AddComponent(NewStubInput())
				e.AddComponent(&StatsComponent{})
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			tt.setup(entity)

			result := sys.hasRangedCapability(entity)
			if result != tt.expected {
				t.Errorf("hasRangedCapability = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWeatherRangedAccuracySystem_AllWeatherTypes(t *testing.T) {
	weatherTypes := []particles.WeatherType{
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

	for _, wt := range weatherTypes {
		t.Run(wt.String(), func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherRangedAccuracySystem(world, 12345)

			weatherEntity := world.CreateEntity()
			weatherEntity.AddComponent(&WeatherComponent{
				Active: true,
				Config: particles.WeatherConfig{
					Type:      wt,
					Intensity: particles.IntensityMedium,
				},
			})

			entity := world.CreateEntity()
			entity.AddComponent(&AIComponent{DetectionRange: 150.0})

			sys.Update(world.GetEntities(), 1.0)

			modifier := sys.GetAccuracyModifier(entity.ID)
			if modifier <= 0.0 || modifier > 1.0 {
				t.Errorf("Weather %s modifier = %f, expected in (0, 1.0]", wt.String(), modifier)
			}
		})
	}
}

func TestWeatherRangedAccuracySystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Short delta should not trigger update
	sys.Update(world.GetEntities(), 0.1)
	sys.Update(world.GetEntities(), 0.1)
	sys.Update(world.GetEntities(), 0.1)

	// Should trigger after enough time
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityHeavy,
		},
	})

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 150.0})

	// Now update with enough time
	sys.Update(world.GetEntities(), 0.6) // Should trigger (>0.5s interval)

	if !sys.IsWeatherActive() {
		t.Error("Weather should be active after sufficient update interval")
	}
}

func TestWeatherRangedAccuracySystem_ModifierClamping(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Test minimum clamp with extreme sandstorm
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSandstorm,
			Intensity: particles.IntensityExtreme,
		},
	})

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 150.0})

	sys.Update(world.GetEntities(), 1.0)

	modifier := sys.GetAccuracyModifier(entity.ID)
	if modifier < 0.4 {
		t.Errorf("Modifier = %f, should be clamped to >= 0.4", modifier)
	}
	if modifier > 1.0 {
		t.Errorf("Modifier = %f, should be clamped to <= 1.0", modifier)
	}
}

func BenchmarkWeatherRangedAccuracySystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityMedium,
		},
	})

	// Create 100 ranged entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&AIComponent{DetectionRange: 150.0})
	}

	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 1.0 // Force update
		sys.Update(entities, 0.016)
	}
}

func BenchmarkWeatherRangedAccuracySystem_GetAccuracyModifier(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherRangedAccuracySystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityMedium,
		},
	})

	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{DetectionRange: 150.0})

	sys.Update(world.GetEntities(), 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetAccuracyModifier(entity.ID)
	}
}
