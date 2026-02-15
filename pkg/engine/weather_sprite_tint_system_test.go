package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestWeatherSpriteTintComponent_Type(t *testing.T) {
	comp := NewWeatherSpriteTintComponent()
	if comp.Type() != "weather_sprite_tint" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "weather_sprite_tint")
	}
}

func TestWeatherSpriteTintComponent_Defaults(t *testing.T) {
	comp := NewWeatherSpriteTintComponent()
	if comp.TintR != 1.0 || comp.TintG != 1.0 || comp.TintB != 1.0 {
		t.Errorf("defaults = (%f,%f,%f), want (1,1,1)", comp.TintR, comp.TintG, comp.TintB)
	}
}

func TestWeatherSpriteTintSystem_Creation(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)
	if sys == nil {
		t.Fatal("NewWeatherSpriteTintSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestWeatherSpriteTintSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genre = %q, want %q", sys.genreID, "horror")
	}
	// Presets should be rebuilt with higher intensity
	rainPreset := sys.presets[particles.WeatherRain]
	if rainPreset.R >= 1.0 {
		t.Errorf("horror rain tint R = %f, want < 1.0", rainPreset.R)
	}
}

func TestWeatherSpriteTintSystem_GenreIntensityScale(t *testing.T) {
	tests := []struct {
		genre    string
		wantMin  float64
		wantMax  float64
	}{
		{"horror", 1.3, 1.5},
		{"cyberpunk", 1.1, 1.3},
		{"fantasy", 0.7, 0.9},
		{"sci-fi", 0.9, 1.1},
		{"unknown", 0.9, 1.1},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherSpriteTintSystem(world, 42)
			sys.genreID = tt.genre
			scale := sys.genreIntensityScale()
			if scale < tt.wantMin || scale > tt.wantMax {
				t.Errorf("scale = %f, want [%f, %f]", scale, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestWeatherSpriteTintSystem_FindActiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	// No weather entity -> inactive
	entities := []*Entity{}
	_, active := sys.findActiveWeather(entities)
	if active {
		t.Error("expected no active weather with empty entities")
	}

	// Add weather entity
	weatherEntity := NewEntity(1)
	weatherComp := &WeatherComponent{
		Config: particles.WeatherConfig{Type: particles.WeatherSnow},
		Active: true,
	}
	weatherEntity.AddComponent(weatherComp)
	entities = append(entities, weatherEntity)

	wType, active := sys.findActiveWeather(entities)
	if !active {
		t.Error("expected active weather")
	}
	if wType != particles.WeatherSnow {
		t.Errorf("weather type = %v, want Snow", wType)
	}
}

func TestWeatherSpriteTintSystem_ApplyTints_Rain(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	spriteEntity := NewEntity(2)
	spriteEntity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	entities := []*Entity{spriteEntity}

	sys.applyTints(entities, particles.WeatherRain, true)

	comp, ok := spriteEntity.GetComponent("weather_sprite_tint")
	if !ok {
		t.Fatal("weather_sprite_tint component not found after apply")
	}
	tint := comp.(*WeatherSpriteTintComponent)

	// Rain should darken R and G slightly, leave B unchanged
	if tint.TintR >= 1.0 {
		t.Errorf("rain TintR = %f, want < 1.0", tint.TintR)
	}
	if tint.TintB < 1.0 {
		t.Errorf("rain TintB = %f, want >= 1.0", tint.TintB)
	}
}

func TestWeatherSpriteTintSystem_ApplyTints_Clear(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	spriteEntity := NewEntity(3)
	spriteEntity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	// First apply rain tint
	entities := []*Entity{spriteEntity}
	sys.applyTints(entities, particles.WeatherRain, true)

	// Then clear it
	sys.applyTints(entities, particles.WeatherRain, false)

	comp, _ := spriteEntity.GetComponent("weather_sprite_tint")
	tint := comp.(*WeatherSpriteTintComponent)
	if tint.TintR != 1.0 || tint.TintG != 1.0 || tint.TintB != 1.0 {
		t.Errorf("cleared tint = (%f,%f,%f), want (1,1,1)", tint.TintR, tint.TintG, tint.TintB)
	}
}

func TestWeatherSpriteTintSystem_SkipsNonSpriteEntities(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	noSpriteEntity := NewEntity(4)
	entities := []*Entity{noSpriteEntity}

	sys.applyTints(entities, particles.WeatherRain, true)

	_, ok := noSpriteEntity.GetComponent("weather_sprite_tint")
	if ok {
		t.Error("non-sprite entity should not get weather tint component")
	}
}

func TestWeatherSpriteTintSystem_Update_Throttled(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	spriteEntity := NewEntity(5)
	spriteEntity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
	entities := []*Entity{spriteEntity}

	// First update at dt=0.1 should not trigger (under 0.5 interval)
	sys.Update(entities, 0.1)
	_, ok := spriteEntity.GetComponent("weather_sprite_tint")
	if ok {
		t.Error("update should be throttled at 0.1s")
	}

	// Accumulate past threshold
	sys.Update(entities, 0.5)
	// No weather entity present, so tints should still not be applied (no change from default)
}

func TestWeatherSpriteTintSystem_Update_WithWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	weatherEntity := NewEntity(10)
	weatherEntity.AddComponent(&WeatherComponent{
		Config: particles.WeatherConfig{Type: particles.WeatherBloodRain},
		Active: true,
	})

	spriteEntity := NewEntity(11)
	spriteEntity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

	entities := []*Entity{weatherEntity, spriteEntity}

	sys.Update(entities, 0.6) // Past throttle interval

	comp, ok := spriteEntity.GetComponent("weather_sprite_tint")
	if !ok {
		t.Fatal("expected weather_sprite_tint after update with active weather")
	}
	tint := comp.(*WeatherSpriteTintComponent)
	// Blood rain should darken green and blue channels
	if tint.TintG >= 1.0 {
		t.Errorf("blood rain TintG = %f, want < 1.0", tint.TintG)
	}
	if tint.TintB >= 1.0 {
		t.Errorf("blood rain TintB = %f, want < 1.0", tint.TintB)
	}
}

func TestWeatherSpriteTintSystem_AllWeatherTypes(t *testing.T) {
	tests := []struct {
		name        string
		weatherType particles.WeatherType
		wantRLess   bool // TintR < 1.0
		wantGLess   bool // TintG < 1.0
		wantBLess   bool // TintB < 1.0
	}{
		{"rain", particles.WeatherRain, true, true, false},
		{"snow", particles.WeatherSnow, false, false, false},
		{"fog", particles.WeatherFog, true, true, true},
		{"dust", particles.WeatherDust, false, true, true},
		{"ash", particles.WeatherAsh, true, true, true},
		{"neon_rain", particles.WeatherNeonRain, true, false, false},
		{"blood_rain", particles.WeatherBloodRain, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherSpriteTintSystem(world, 42)

			entity := NewEntity(100)
			entity.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})

			sys.applyTints([]*Entity{entity}, tt.weatherType, true)

			comp, ok := entity.GetComponent("weather_sprite_tint")
			if !ok {
				t.Fatal("tint component missing")
			}
			tint := comp.(*WeatherSpriteTintComponent)

			if tt.wantRLess && tint.TintR >= 1.0 {
				t.Errorf("TintR = %f, want < 1.0", tint.TintR)
			}
			if !tt.wantRLess && tint.TintR < 1.0 {
				t.Errorf("TintR = %f, want >= 1.0", tint.TintR)
			}
			if tt.wantGLess && tint.TintG >= 1.0 {
				t.Errorf("TintG = %f, want < 1.0", tint.TintG)
			}
			if !tt.wantGLess && tint.TintG < 1.0 {
				t.Errorf("TintG = %f, want >= 1.0", tint.TintG)
			}
			if tt.wantBLess && tint.TintB >= 1.0 {
				t.Errorf("TintB = %f, want < 1.0", tint.TintB)
			}
			if !tt.wantBLess && tint.TintB < 1.0 {
				t.Errorf("TintB = %f, want >= 1.0", tint.TintB)
			}
		})
	}
}

func TestClampTint(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0.3, 0.5},
		{0.5, 0.5},
		{0.8, 0.8},
		{1.0, 1.0},
		{1.1, 1.1},
		{1.5, 1.1},
	}
	for _, tt := range tests {
		got := clampTint(tt.input)
		if got != tt.want {
			t.Errorf("clampTint(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestWeatherSpriteTintSystem_PresetRebuildsOnGenreChange(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	fantasyRain := sys.presets[particles.WeatherRain]
	sys.SetGenre("horror")
	horrorRain := sys.presets[particles.WeatherRain]

	// Horror should have stronger (lower R) tint than fantasy
	if horrorRain.R >= fantasyRain.R {
		t.Errorf("horror rain R = %f, fantasy = %f, want horror < fantasy", horrorRain.R, fantasyRain.R)
	}
}

func BenchmarkWeatherSpriteTintSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherSpriteTintSystem(world, 42)

	weatherEntity := NewEntity(1)
	weatherEntity.AddComponent(&WeatherComponent{
		Config: particles.WeatherConfig{Type: particles.WeatherRain},
		Active: true,
	})

	entities := make([]*Entity, 101)
	entities[0] = weatherEntity
	for i := 1; i <= 100; i++ {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = sys.updateInterval // Force recheck
		sys.lastWeatherActive = false           // Force change detection
		sys.Update(entities, 0.6)
	}
}
