package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestWeatherCombatSystem_Update(t *testing.T) {
	tests := []struct {
		name          string
		weatherType   particles.WeatherType
		intensity     particles.WeatherIntensity
		wantEffect    string
		wantNoEffect  bool
		wantMagnitude float64
	}{
		{
			name:          "rain applies wet effect",
			weatherType:   particles.WeatherRain,
			intensity:     particles.IntensityExtreme,
			wantEffect:    "wet",
			wantMagnitude: 0.1,
		},
		{
			name:          "snow applies chilled effect",
			weatherType:   particles.WeatherSnow,
			intensity:     particles.IntensityExtreme,
			wantEffect:    "chilled",
			wantMagnitude: 0.15,
		},
		{
			name:          "sandstorm applies sandblasted effect",
			weatherType:   particles.WeatherSandstorm,
			intensity:     particles.IntensityExtreme,
			wantEffect:    "sandblasted",
			wantMagnitude: 2.0,
		},
		{
			name:          "ash applies burning effect",
			weatherType:   particles.WeatherAsh,
			intensity:     particles.IntensityExtreme,
			wantEffect:    "burning",
			wantMagnitude: 1.0,
		},
		{
			name:         "fog applies no effect",
			weatherType:  particles.WeatherFog,
			intensity:    particles.IntensityExtreme,
			wantNoEffect: true,
		},
		{
			name:         "dust applies no effect",
			weatherType:  particles.WeatherDust,
			intensity:    particles.IntensityExtreme,
			wantNoEffect: true,
		},
		{
			name:          "medium intensity scales magnitude",
			weatherType:   particles.WeatherRain,
			intensity:     particles.IntensityMedium,
			wantEffect:    "wet",
			wantMagnitude: 0.05,
		},
		{
			name:          "light intensity scales magnitude",
			weatherType:   particles.WeatherRain,
			intensity:     particles.IntensityLight,
			wantEffect:    "wet",
			wantMagnitude: 0.025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherCombatSystem(world)

			// Create a weather entity.
			weatherEntity := world.CreateEntity()
			wc := &WeatherComponent{
				Config: particles.WeatherConfig{
					Type:      tt.weatherType,
					Intensity: tt.intensity,
				},
				Active: true,
			}
			weatherEntity.AddComponent(wc)

			// Create a target entity with health.
			target := world.CreateEntity()
			target.AddComponent(&HealthComponent{Current: 100, Max: 100})

			entities := []*Entity{weatherEntity, target}
			sys.Update(entities, 0.016)

			// Check whether a status effect was applied.
			hasEffect := false
			var appliedEffect *StatusEffectComponent
			for _, comp := range target.Components {
				if se, ok := comp.(*StatusEffectComponent); ok {
					hasEffect = true
					appliedEffect = se
					break
				}
			}

			if tt.wantNoEffect {
				if hasEffect {
					t.Errorf("expected no effect, but got %q", appliedEffect.EffectType)
				}
				return
			}

			if !hasEffect {
				t.Fatalf("expected effect %q, but none was applied", tt.wantEffect)
			}
			if appliedEffect.EffectType != tt.wantEffect {
				t.Errorf("effect type = %q, want %q", appliedEffect.EffectType, tt.wantEffect)
			}
			if appliedEffect.Magnitude != tt.wantMagnitude {
				t.Errorf("magnitude = %f, want %f", appliedEffect.Magnitude, tt.wantMagnitude)
			}
		})
	}
}

func TestWeatherCombatSystem_Cooldown(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCombatSystem(world)

	weatherEntity := world.CreateEntity()
	wc := &WeatherComponent{
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityExtreme,
		},
		Active: true,
	}
	weatherEntity.AddComponent(wc)

	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{weatherEntity, target}

	// First update applies effect.
	sys.Update(entities, 0.016)

	// Remove the effect to test cooldown prevents re-application.
	target.RemoveComponent("status_effect")

	// Second update within cooldown should NOT re-apply.
	sys.Update(entities, 0.016)
	for _, comp := range target.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			t.Error("effect should not re-apply within cooldown period")
			return
		}
	}

	// After cooldown expires, effect should be re-applied.
	sys.Update(entities, 6.0) // Advance past 5s cooldown.
	sys.Update(entities, 0.016)
	for _, comp := range target.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			return // OK: effect re-applied after cooldown.
		}
	}
	t.Error("effect should re-apply after cooldown expires")
}

func TestWeatherCombatSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCombatSystem(world)

	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{target}
	sys.Update(entities, 0.016)

	for _, comp := range target.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			t.Error("no weather present, but effect was applied")
		}
	}
}

func TestWeatherCombatSystem_InactiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCombatSystem(world)

	weatherEntity := world.CreateEntity()
	wc := &WeatherComponent{
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityExtreme,
		},
		Active: false, // Inactive weather.
	}
	weatherEntity.AddComponent(wc)

	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{weatherEntity, target}
	sys.Update(entities, 0.016)

	for _, comp := range target.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			t.Error("inactive weather should not apply effects")
		}
	}
}

func TestWeatherCombatSystem_DuplicateEffectPrevented(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCombatSystem(world)
	// Reset cooldowns to isolate duplicate-check logic.
	sys.applyCooldown = 0

	weatherEntity := world.CreateEntity()
	wc := &WeatherComponent{
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityExtreme,
		},
		Active: true,
	}
	weatherEntity.AddComponent(wc)

	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{weatherEntity, target}

	// First update applies effect.
	sys.Update(entities, 0.016)

	// Second update should NOT add a duplicate because hasWeatherEffect is true.
	sys.Update(entities, 0.016)

	effectCount := 0
	for _, comp := range target.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			effectCount++
		}
	}
	if effectCount > 1 {
		t.Errorf("expected at most 1 weather effect, got %d", effectCount)
	}
}

func TestWeatherCombatSystem_NoHealthEntity(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherCombatSystem(world)

	weatherEntity := world.CreateEntity()
	wc := &WeatherComponent{
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityExtreme,
		},
		Active: true,
	}
	weatherEntity.AddComponent(wc)

	// Entity without health should not get effects.
	noHealthEntity := world.CreateEntity()
	noHealthEntity.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{weatherEntity, noHealthEntity}
	sys.Update(entities, 0.016)

	for _, comp := range noHealthEntity.Components {
		if _, ok := comp.(*StatusEffectComponent); ok {
			t.Error("entity without health should not receive weather effects")
		}
	}
}
