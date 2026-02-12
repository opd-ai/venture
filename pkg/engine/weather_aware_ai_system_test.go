package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherAwareAISystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherAwareAISystem returned nil")
	}

	if system.world != world {
		t.Error("world reference not set")
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.originalRanges == nil {
		t.Error("originalRanges map not initialized")
	}
}

func TestWeatherAwareAISystem_VisibilityMultiplier(t *testing.T) {
	system := NewWeatherAwareAISystem(NewWorld(), 12345)

	tests := []struct {
		name        string
		weatherType particles.WeatherType
		intensity   particles.WeatherIntensity
		minMult     float64
		maxMult     float64
	}{
		{
			name:        "fog light",
			weatherType: particles.WeatherFog,
			intensity:   particles.IntensityLight,
			minMult:     0.75,
			maxMult:     0.85,
		},
		{
			name:        "fog heavy",
			weatherType: particles.WeatherFog,
			intensity:   particles.IntensityHeavy,
			minMult:     0.30,
			maxMult:     0.50,
		},
		{
			name:        "rain medium",
			weatherType: particles.WeatherRain,
			intensity:   particles.IntensityMedium,
			minMult:     0.75,
			maxMult:     0.85,
		},
		{
			name:        "sandstorm heavy",
			weatherType: particles.WeatherSandstorm,
			intensity:   particles.IntensityHeavy,
			minMult:     0.30,
			maxMult:     0.35,
		},
		{
			name:        "snow light",
			weatherType: particles.WeatherSnow,
			intensity:   particles.IntensityLight,
			minMult:     0.85,
			maxMult:     0.95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mult := system.GetVisibilityMultiplier(tt.weatherType, tt.intensity)
			if mult < tt.minMult || mult > tt.maxMult {
				t.Errorf("multiplier = %v, want between %v and %v", mult, tt.minMult, tt.maxMult)
			}
		})
	}
}

func TestWeatherAwareAISystem_DetectionRangeReduction(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Create AI entity with known detection range
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	aiEntity.AddComponent(&AIComponent{
		DetectionRange: 200.0,
		State:          AIStateIdle,
	})

	// Create weather entity with active fog
	weatherEntity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherFog
	config.Intensity = particles.IntensityMedium
	weatherComp := NewWeatherComponent(config)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Process pending entity additions
	world.Update(0)

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	// Force immediate update by setting timeSinceCheck high
	system.timeSinceCheck = system.updateInterval

	// Run update
	system.Update(entities, 0.01)

	// Check detection range was reduced
	aiComp, _ := aiEntity.GetComponent("ai")
	ai := aiComp.(*AIComponent)

	expectedMult := system.GetVisibilityMultiplier(particles.WeatherFog, particles.IntensityMedium)
	expectedRange := 200.0 * expectedMult

	if ai.DetectionRange != expectedRange {
		t.Errorf("DetectionRange = %v, want %v", ai.DetectionRange, expectedRange)
	}

	// Verify original range was stored
	if stored, exists := system.originalRanges[aiEntity.ID]; !exists || stored != 200.0 {
		t.Errorf("originalRanges[%d] = %v, want 200.0", aiEntity.ID, stored)
	}
}

func TestWeatherAwareAISystem_DetectionRangeRestored(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Create AI entity
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	aiEntity.AddComponent(&AIComponent{
		DetectionRange: 200.0,
		State:          AIStateIdle,
	})

	// Create active weather
	weatherEntity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherFog
	config.Intensity = particles.IntensityHeavy
	weatherComp := NewWeatherComponent(config)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	// Force immediate update
	system.timeSinceCheck = system.updateInterval

	// Apply weather effect
	system.Update(entities, 0.01)

	// Deactivate weather
	weatherComp.Active = false

	// Force another update
	system.timeSinceCheck = system.updateInterval

	// Update again to restore
	system.Update(entities, 0.01)

	// Check detection range was restored
	aiComp, _ := aiEntity.GetComponent("ai")
	ai := aiComp.(*AIComponent)

	if ai.DetectionRange != 200.0 {
		t.Errorf("DetectionRange after restore = %v, want 200.0", ai.DetectionRange)
	}

	// Verify original range was removed from cache
	if _, exists := system.originalRanges[aiEntity.ID]; exists {
		t.Error("originalRanges should be cleared after weather stops")
	}
}

func TestWeatherAwareAISystem_NoWeatherNoChange(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Create AI entity with no weather
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	aiEntity.AddComponent(&AIComponent{
		DetectionRange: 200.0,
		State:          AIStateIdle,
	})

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	// Force immediate update
	system.timeSinceCheck = system.updateInterval

	// Update without weather
	system.Update(entities, 0.01)

	// Check detection range unchanged
	aiComp, _ := aiEntity.GetComponent("ai")
	ai := aiComp.(*AIComponent)

	if ai.DetectionRange != 200.0 {
		t.Errorf("DetectionRange = %v, want 200.0 (unchanged)", ai.DetectionRange)
	}
}

func TestWeatherAwareAISystem_MultipleAIEntities(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Create multiple AI entities with different ranges
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&AIComponent{DetectionRange: 200.0, State: AIStateIdle})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 200, Y: 200})
	entity2.AddComponent(&AIComponent{DetectionRange: 300.0, State: AIStatePatrol})

	entity3 := world.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 300, Y: 300})
	entity3.AddComponent(&AIComponent{DetectionRange: 150.0, State: AIStateIdle})

	// Create fog weather
	weatherEntity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherFog
	config.Intensity = particles.IntensityMedium
	weatherComp := NewWeatherComponent(config)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	// Force immediate update
	system.timeSinceCheck = system.updateInterval

	// Update
	system.Update(entities, 0.01)

	// All AI entities should have reduced ranges
	mult := system.GetVisibilityMultiplier(particles.WeatherFog, particles.IntensityMedium)

	checks := []struct {
		entity   *Entity
		expected float64
	}{
		{entity1, 200.0 * mult},
		{entity2, 300.0 * mult},
		{entity3, 150.0 * mult},
	}

	for i, check := range checks {
		aiComp, _ := check.entity.GetComponent("ai")
		ai := aiComp.(*AIComponent)
		if ai.DetectionRange != check.expected {
			t.Errorf("entity %d DetectionRange = %v, want %v", i+1, ai.DetectionRange, check.expected)
		}
	}
}

func TestWeatherAwareAISystem_DifferentWeatherTypes(t *testing.T) {
	tests := []struct {
		name        string
		weatherType particles.WeatherType
		intensity   particles.WeatherIntensity
	}{
		{"rain_light", particles.WeatherRain, particles.IntensityLight},
		{"rain_heavy", particles.WeatherRain, particles.IntensityHeavy},
		{"fog_medium", particles.WeatherFog, particles.IntensityMedium},
		{"snow_heavy", particles.WeatherSnow, particles.IntensityHeavy},
		{"sandstorm_heavy", particles.WeatherSandstorm, particles.IntensityHeavy},
		{"smog_medium", particles.WeatherSmog, particles.IntensityMedium},
		{"dust_light", particles.WeatherDust, particles.IntensityLight},
		{"neon_rain_medium", particles.WeatherNeonRain, particles.IntensityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewWeatherAwareAISystem(world, 12345)

			// Create AI entity
			aiEntity := world.CreateEntity()
			aiEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
			aiEntity.AddComponent(&AIComponent{
				DetectionRange: 200.0,
				State:          AIStateIdle,
			})

			// Create weather
			weatherEntity := world.CreateEntity()
			config := particles.DefaultWeatherConfig()
			config.Type = tt.weatherType
			config.Intensity = tt.intensity
			weatherComp := NewWeatherComponent(config)
			weatherComp.Active = true
			weatherEntity.AddComponent(weatherComp)

			world.Update(0) // Process pending entity additions
			entities := world.GetEntities()

			// Force immediate update
			system.timeSinceCheck = system.updateInterval

			// Update
			system.Update(entities, 0.01)

			// Verify detection was reduced
			aiComp, _ := aiEntity.GetComponent("ai")
			ai := aiComp.(*AIComponent)

			if ai.DetectionRange >= 200.0 {
				t.Errorf("DetectionRange = %v, should be less than 200.0 with weather", ai.DetectionRange)
			}

			mult := system.GetVisibilityMultiplier(tt.weatherType, tt.intensity)
			expected := 200.0 * mult
			if ai.DetectionRange != expected {
				t.Errorf("DetectionRange = %v, want %v", ai.DetectionRange, expected)
			}
		})
	}
}

func TestWeatherAwareAISystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Default interval
	if system.updateInterval != 0.5 {
		t.Errorf("default updateInterval = %v, want 0.5", system.updateInterval)
	}

	// Set custom interval
	system.SetUpdateInterval(1.0)
	if system.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0", system.updateInterval)
	}

	// Invalid interval should not change
	system.SetUpdateInterval(-1.0)
	if system.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0 (unchanged)", system.updateInterval)
	}
}

func TestWeatherAwareAISystem_SkipsNonAIEntities(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Create non-AI entity
	nonAI := world.CreateEntity()
	nonAI.AddComponent(&PositionComponent{X: 100, Y: 100})
	nonAI.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Create weather
	weatherEntity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherFog
	config.Intensity = particles.IntensityHeavy
	weatherComp := NewWeatherComponent(config)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	// Force immediate update
	system.timeSinceCheck = system.updateInterval

	// Should not panic or error
	system.Update(entities, 0.01)

	// Non-AI entity should not be in originalRanges
	if _, exists := system.originalRanges[nonAI.ID]; exists {
		t.Error("non-AI entity should not be in originalRanges")
	}
}

func TestWeatherAwareAISystem_InactiveWeatherIgnored(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)

	// Create AI entity
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	aiEntity.AddComponent(&AIComponent{
		DetectionRange: 200.0,
		State:          AIStateIdle,
	})

	// Create inactive weather
	weatherEntity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherFog
	config.Intensity = particles.IntensityHeavy
	weatherComp := NewWeatherComponent(config)
	weatherComp.Active = false // Not active
	weatherEntity.AddComponent(weatherComp)

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	// Force immediate update
	system.timeSinceCheck = system.updateInterval

	// Update
	system.Update(entities, 0.01)

	// Detection range should be unchanged
	aiComp, _ := aiEntity.GetComponent("ai")
	ai := aiComp.(*AIComponent)

	if ai.DetectionRange != 200.0 {
		t.Errorf("DetectionRange = %v, want 200.0 (unchanged with inactive weather)", ai.DetectionRange)
	}
}

func TestWeatherAwareAISystem_MaxReductionCap(t *testing.T) {
	system := NewWeatherAwareAISystem(NewWorld(), 12345)

	// Even with worst conditions, should not go below 30% visibility
	mult := system.GetVisibilityMultiplier(particles.WeatherSandstorm, particles.IntensityHeavy)

	if mult < 0.30 {
		t.Errorf("visibility multiplier = %v, should not be below 0.30", mult)
	}
}

func BenchmarkWeatherAwareAISystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherAwareAISystem(world, 12345)
	rng := rand.New(rand.NewSource(12345))

	// Create many AI entities
	for i := 0; i < 500; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(rng.Intn(1000)),
			Y: float64(rng.Intn(1000)),
		})
		entity.AddComponent(&AIComponent{
			DetectionRange: 150.0 + float64(rng.Intn(100)),
			State:          AIStateIdle,
		})
	}

	// Create weather
	weatherEntity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherFog
	config.Intensity = particles.IntensityMedium
	weatherComp := NewWeatherComponent(config)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	world.Update(0) // Process pending entity additions
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset timer to force update
		system.timeSinceCheck = system.updateInterval
		system.Update(entities, 0.016)
	}
}
