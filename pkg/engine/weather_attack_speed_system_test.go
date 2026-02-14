//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherAttackSpeedSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewWeatherAttackSpeedSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", sys.genreID)
	}
}

func TestWeatherAttackSpeedSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("Genre = %s, want horror", sys.genreID)
	}
}

func TestWeatherAttackSpeedSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	entity := world.CreateEntity()
	entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
	sys.Update(world.GetEntities(), 1.0)
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.Cooldown != 1.0 {
		t.Errorf("Cooldown = %f, want 1.0 (no weather)", attack.Cooldown)
	}
}

func TestWeatherAttackSpeedSystem_SnowSlowsAttacks(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{Type: particles.WeatherSnow, Intensity: particles.IntensityMedium},
	})
	entity := world.CreateEntity()
	entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
	sys.Update(world.GetEntities(), 1.0)
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.Cooldown <= 1.0 {
		t.Errorf("Cooldown = %f, want > 1.0 (snow should slow attacks)", attack.Cooldown)
	}
}

func TestWeatherAttackSpeedSystem_RainSpeedsAttacks(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{Type: particles.WeatherRain, Intensity: particles.IntensityMedium},
	})
	entity := world.CreateEntity()
	entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
	sys.Update(world.GetEntities(), 1.0)
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.Cooldown >= 1.0 {
		t.Errorf("Cooldown = %f, want < 1.0 (rain should speed attacks)", attack.Cooldown)
	}
}

func TestWeatherAttackSpeedSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name, genre    string
		weather        particles.WeatherType
		expectSlowdown bool
	}{
		{"fantasy_snow", "fantasy", particles.WeatherSnow, true},
		{"scifi_snow", "scifi", particles.WeatherSnow, true},
		{"horror_snow", "horror", particles.WeatherSnow, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherAttackSpeedSystem(world, 12345)
			sys.SetGenre(tt.genre)
			weatherEntity := world.CreateEntity()
			weatherEntity.AddComponent(&WeatherComponent{
				Active: true,
				Config: particles.Config{Type: tt.weather, Intensity: particles.IntensityMedium},
			})
			entity := world.CreateEntity()
			entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
			sys.Update(world.GetEntities(), 1.0)
			attackComp, _ := entity.GetComponent("attack")
			attack := attackComp.(*AttackComponent)
			if tt.expectSlowdown && attack.Cooldown <= 1.0 {
				t.Errorf("Genre %s: Cooldown = %f, want > 1.0 for slowdown", tt.genre, attack.Cooldown)
			}
		})
	}
}

func TestWeatherAttackSpeedSystem_WeatherClears(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{Type: particles.WeatherSnow, Intensity: particles.IntensityMedium},
	})
	entity := world.CreateEntity()
	entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
	sys.Update(world.GetEntities(), 1.0)
	// Clear weather
	weatherComp, _ := weatherEntity.GetComponent("weather")
	weather := weatherComp.(*WeatherComponent)
	weather.Active = false
	sys.Update(world.GetEntities(), 1.0)
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.Cooldown != 1.0 {
		t.Errorf("Cooldown after weather clear = %f, want 1.0", attack.Cooldown)
	}
}

func TestWeatherAttackSpeedSystem_AllWeatherTypes(t *testing.T) {
	tests := []struct {
		name          string
		weather       particles.WeatherType
		expectPenalty bool
	}{
		{"rain", particles.WeatherRain, false},
		{"snow", particles.WeatherSnow, true},
		{"fog", particles.WeatherFog, false},
		{"dust", particles.WeatherDust, true},
		{"ash", particles.WeatherAsh, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherAttackSpeedSystem(world, 12345)
			weatherEntity := world.CreateEntity()
			weatherEntity.AddComponent(&WeatherComponent{
				Active: true,
				Config: particles.Config{Type: tt.weather, Intensity: particles.IntensityMedium},
			})
			entity := world.CreateEntity()
			entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
			sys.Update(world.GetEntities(), 1.0)
			attackComp, _ := entity.GetComponent("attack")
			attack := attackComp.(*AttackComponent)
			if tt.expectPenalty && attack.Cooldown <= 1.0 {
				t.Errorf("%s: Cooldown = %f, want > 1.0", tt.name, attack.Cooldown)
			}
			if !tt.expectPenalty && attack.Cooldown > 1.0 {
				t.Errorf("%s: Cooldown = %f, want <= 1.0", tt.name, attack.Cooldown)
			}
		})
	}
}

func TestWeatherAttackSpeedSystem_GetAttackSpeedModifier(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	entity := world.CreateEntity()
	entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
	mod := sys.GetAttackSpeedModifier(entity.ID)
	if mod != 1.0 {
		t.Errorf("Initial modifier = %f, want 1.0", mod)
	}
}

func BenchmarkWeatherAttackSpeedSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherAttackSpeedSystem(world, 12345)
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.Config{Type: particles.WeatherSnow, Intensity: particles.IntensityMedium},
	})
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0})
	}
	entities := world.GetEntities()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
