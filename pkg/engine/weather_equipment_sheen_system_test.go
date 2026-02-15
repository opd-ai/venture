package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherEquipmentSheenSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 42)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.seed != 42 {
		t.Errorf("seed = %d, want 42", sys.seed)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want fantasy", sys.genreID)
	}
}

func TestWeatherEquipmentSheenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	genres := []string{"fantasy", "horror", "cyberpunk", "scifi", "postapoc"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("SetGenre(%q): got %q", g, sys.genreID)
		}
	}
}

func TestWeatherEquipmentSheenSystem_NeutralModifier(t *testing.T) {
	mod := neutralModifier()
	if mod.SheenScale != 1.0 {
		t.Errorf("SheenScale = %f, want 1.0", mod.SheenScale)
	}
	if mod.ReflectivityAdd != 0.0 {
		t.Errorf("ReflectivityAdd = %f, want 0.0", mod.ReflectivityAdd)
	}
	if mod.RoughnessAdd != 0.0 {
		t.Errorf("RoughnessAdd = %f, want 0.0", mod.RoughnessAdd)
	}
	if mod.PulseSpeedScale != 1.0 {
		t.Errorf("PulseSpeedScale = %f, want 1.0", mod.PulseSpeedScale)
	}
}

func TestWeatherEquipmentSheenSystem_GetWeatherModifier(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	tests := []struct {
		name      string
		weather   particles.WeatherType
		intensity particles.WeatherIntensity
		checkFn   func(t *testing.T, mod weatherSheenModifier)
	}{
		{
			"rain_increases_sheen",
			particles.WeatherRain,
			particles.IntensityHeavy,
			func(t *testing.T, mod weatherSheenModifier) {
				if mod.SheenScale <= 1.0 {
					t.Errorf("rain SheenScale = %f, want > 1.0", mod.SheenScale)
				}
				if mod.ReflectivityAdd <= 0.0 {
					t.Errorf("rain ReflectivityAdd = %f, want > 0", mod.ReflectivityAdd)
				}
			},
		},
		{
			"snow_dampens_sheen",
			particles.WeatherSnow,
			particles.IntensityMedium,
			func(t *testing.T, mod weatherSheenModifier) {
				if mod.SheenScale >= 1.0 {
					t.Errorf("snow SheenScale = %f, want < 1.0", mod.SheenScale)
				}
				if mod.RoughnessAdd <= 0.0 {
					t.Errorf("snow RoughnessAdd = %f, want > 0", mod.RoughnessAdd)
				}
			},
		},
		{
			"dust_roughens_surface",
			particles.WeatherDust,
			particles.IntensityHeavy,
			func(t *testing.T, mod weatherSheenModifier) {
				if mod.RoughnessAdd <= 0.0 {
					t.Errorf("dust RoughnessAdd = %f, want > 0", mod.RoughnessAdd)
				}
				if mod.ReflectivityAdd >= 0.0 {
					t.Errorf("dust ReflectivityAdd = %f, want < 0", mod.ReflectivityAdd)
				}
			},
		},
		{
			"sandstorm_roughens_surface",
			particles.WeatherSandstorm,
			particles.IntensityExtreme,
			func(t *testing.T, mod weatherSheenModifier) {
				if mod.RoughnessAdd <= 0.0 {
					t.Errorf("sandstorm RoughnessAdd = %f, want > 0", mod.RoughnessAdd)
				}
			},
		},
		{
			"fog_mutes_softly",
			particles.WeatherFog,
			particles.IntensityLight,
			func(t *testing.T, mod weatherSheenModifier) {
				if mod.SheenScale >= 1.0 {
					t.Errorf("fog SheenScale = %f, want < 1.0", mod.SheenScale)
				}
				if mod.PulseSpeedScale >= 1.0 {
					t.Errorf("fog PulseSpeedScale = %f, want < 1.0", mod.PulseSpeedScale)
				}
			},
		},
		{
			"light_intensity_lower_effect",
			particles.WeatherRain,
			particles.IntensityLight,
			func(t *testing.T, mod weatherSheenModifier) {
				heavyMod := sys.getWeatherModifier(particles.WeatherRain, particles.IntensityHeavy)
				if mod.SheenScale >= heavyMod.SheenScale {
					t.Error("light rain should have less sheen boost than heavy")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := sys.getWeatherModifier(tt.weather, tt.intensity)
			tt.checkFn(t, mod)
		})
	}
}

func TestWeatherEquipmentSheenSystem_GenreModifiers(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name    string
		genre   string
		weather particles.WeatherType
		checkFn func(t *testing.T, mod weatherSheenModifier, baseMod weatherSheenModifier)
	}{
		{
			"cyberpunk_rain_amplified",
			"cyberpunk",
			particles.WeatherRain,
			func(t *testing.T, mod, base weatherSheenModifier) {
				if mod.SheenScale <= base.SheenScale {
					t.Error("cyberpunk rain should amplify sheen beyond base")
				}
			},
		},
		{
			"horror_dampened",
			"horror",
			particles.WeatherRain,
			func(t *testing.T, mod, base weatherSheenModifier) {
				if mod.SheenScale >= base.SheenScale {
					t.Error("horror should dampen sheen")
				}
			},
		},
		{
			"postapoc_extra_roughness",
			"postapoc",
			particles.WeatherDust,
			func(t *testing.T, mod, base weatherSheenModifier) {
				if mod.RoughnessAdd <= base.RoughnessAdd {
					t.Error("postapoc should add extra roughness")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewWeatherEquipmentSheenSystem(world, 1)
			baseSys := NewWeatherEquipmentSheenSystem(world, 1)

			sys.SetGenre(tt.genre)
			baseSys.SetGenre("fantasy") // baseline

			mod := sys.getWeatherModifier(tt.weather, particles.IntensityHeavy)
			baseMod := baseSys.getWeatherModifier(tt.weather, particles.IntensityHeavy)
			tt.checkFn(t, mod, baseMod)
		})
	}
}

func TestWeatherEquipmentSheenSystem_FindActiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	// No weather entity
	entities := []*Entity{NewEntity(1)}
	_, _, active := sys.findActiveWeather(entities)
	if active {
		t.Error("should not find weather when no weather component")
	}

	// With weather entity
	weatherEntity := NewEntity(2)
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityHeavy,
		},
	})
	entities = append(entities, weatherEntity)
	wType, wIntensity, active := sys.findActiveWeather(entities)
	if !active {
		t.Error("should find active weather")
	}
	if wType != particles.WeatherSnow {
		t.Errorf("weather type = %v, want snow", wType)
	}
	if wIntensity != particles.IntensityHeavy {
		t.Errorf("intensity = %v, want heavy", wIntensity)
	}
}

func TestWeatherEquipmentSheenSystem_FindActiveWeather_Inactive(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	weatherEntity := NewEntity(1)
	weatherEntity.AddComponent(&WeatherComponent{
		Active: false,
		Config: particles.WeatherConfig{Type: particles.WeatherRain},
	})

	_, _, active := sys.findActiveWeather([]*Entity{weatherEntity})
	if active {
		t.Error("inactive weather should not be found")
	}
}

func TestWeatherEquipmentSheenSystem_Update_NilWorld(t *testing.T) {
	sys := NewWeatherEquipmentSheenSystem(nil, 1)
	sys.world = nil
	// Should not panic
	sys.Update([]*Entity{}, 0.5)
}

func TestWeatherEquipmentSheenSystem_Update_AppliesModifier(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	// Create weather entity
	weatherEntity := NewEntity(1)
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityHeavy,
		},
	})

	// Create entity with material sheen
	equipEntity := NewEntity(2)
	sheen := NewMaterialSheenComponent()
	sheen.Enabled = true
	sheen.SheenIntensity = 0.5
	sheen.Reflectivity = 0.3
	sheen.Roughness = 0.4
	equipEntity.AddComponent(sheen)

	entities := []*Entity{weatherEntity, equipEntity}

	// First update triggers weather check (timeSinceCheck starts at 0, interval is 0.5)
	sys.timeSinceCheck = 0.5 // Force check
	sys.Update(entities, 0.01)

	// Rain should increase reflectivity
	if sheen.Reflectivity <= 0.3 {
		t.Errorf("rain should increase reflectivity, got %f", sheen.Reflectivity)
	}
}

func TestWeatherEquipmentSheenSystem_Update_SkipsDisabledSheen(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	weatherEntity := NewEntity(1)
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityExtreme,
		},
	})

	equipEntity := NewEntity(2)
	sheen := NewMaterialSheenComponent()
	sheen.Enabled = false // Disabled
	sheen.Reflectivity = 0.3
	equipEntity.AddComponent(sheen)

	entities := []*Entity{weatherEntity, equipEntity}
	sys.timeSinceCheck = 0.5
	sys.Update(entities, 0.01)

	if sheen.Reflectivity != 0.3 {
		t.Errorf("disabled sheen should not be modified, got reflectivity %f", sheen.Reflectivity)
	}
}

func TestWeatherEquipmentSheenSystem_Update_ThrottlesCheck(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentSheenSystem(world, 1)

	entity := NewEntity(1)
	entities := []*Entity{entity}

	// Small deltaTime, should not trigger check
	sys.Update(entities, 0.1)
	if sys.timeSinceCheck < 0.1 {
		t.Error("timeSinceCheck should accumulate")
	}
}

func TestWeatherEquipmentSheenSystem_ClampF(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		min, max float64
		want     float64
	}{
		{"within_range", 0.5, 0.0, 1.0, 0.5},
		{"below_min", -0.5, 0.0, 1.0, 0.0},
		{"above_max", 1.5, 0.0, 1.0, 1.0},
		{"at_min", 0.0, 0.0, 1.0, 0.0},
		{"at_max", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampF(tt.v, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clampF(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestWeatherEquipmentSheenSystem_Deterministic(t *testing.T) {
	world := NewWorld()
	sys1 := NewWeatherEquipmentSheenSystem(world, 12345)
	sys2 := NewWeatherEquipmentSheenSystem(world, 12345)

	for i := 0; i < 10; i++ {
		v1 := sys1.rng.Float64()
		v2 := sys2.rng.Float64()
		if v1 != v2 {
			t.Errorf("iteration %d: different rng values %f != %f", i, v1, v2)
		}
	}
}
