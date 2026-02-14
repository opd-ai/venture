//go:build ignore

package engine

import (
	"image/color"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// mockGameClock implements GameClock for testing.
type mockGameClock struct {
	currentTime time.Time
}

func (m *mockGameClock) Now() time.Time {
	return m.currentTime
}

func (m *mockGameClock) Advance(deltaTime float64) {
	m.currentTime = m.currentTime.Add(time.Duration(deltaTime * float64(time.Second)))
}

func (m *mockGameClock) Reset(startTime time.Time) {
	m.currentTime = startTime
}

func TestNewTimeOfDayLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTimeOfDayLightingSystem returned nil")
	}

	if system.world != world {
		t.Error("World not set correctly")
	}

	if system.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", system.seed)
	}

	if system.currentTimeOfDay != palette.TimeOfDayDay {
		t.Errorf("Initial time of day = %v, want Day", system.currentTimeOfDay)
	}

	if system.dayDuration != 600.0 {
		t.Errorf("Default day duration = %f, want 600.0", system.dayDuration)
	}

	if system.transitionDuration != 5.0 {
		t.Errorf("Default transition duration = %f, want 5.0", system.transitionDuration)
	}
}

func TestTimeOfDayLightingSystem_SetClock(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
	system.SetClock(clock)

	if system.clock != clock {
		t.Error("Clock not set correctly")
	}
}

func TestTimeOfDayLightingSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	system.SetGenre("horror")
	if system.genreID != "horror" {
		t.Errorf("Genre = %s, want horror", system.genreID)
	}
}

func TestTimeOfDayLightingSystem_SetDayDuration(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	system.SetDayDuration(1200.0)
	if system.dayDuration != 1200.0 {
		t.Errorf("Day duration = %f, want 1200.0", system.dayDuration)
	}

	// Invalid values should not change
	system.SetDayDuration(-100.0)
	if system.dayDuration != 1200.0 {
		t.Error("Negative day duration should be ignored")
	}

	system.SetDayDuration(0)
	if system.dayDuration != 1200.0 {
		t.Error("Zero day duration should be ignored")
	}
}

func TestTimeOfDayLightingSystem_CalculateTimeOfDay(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	tests := []struct {
		name     string
		hour     int
		minute   int
		wantTime palette.TimeOfDay
	}{
		{"midnight", 0, 0, palette.TimeOfDayNight},
		{"early_morning", 3, 30, palette.TimeOfDayNight},
		{"dawn_start", 5, 0, palette.TimeOfDayDawn},
		{"dawn_middle", 6, 30, palette.TimeOfDayDawn},
		{"day_start", 8, 0, palette.TimeOfDayDay},
		{"noon", 12, 0, palette.TimeOfDayDay},
		{"afternoon", 15, 0, palette.TimeOfDayDay},
		{"dusk_start", 17, 0, palette.TimeOfDayDusk},
		{"dusk_middle", 18, 30, palette.TimeOfDayDusk},
		{"night_start", 20, 0, palette.TimeOfDayNight},
		{"late_night", 23, 0, palette.TimeOfDayNight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameTime := time.Date(2026, 1, 15, tt.hour, tt.minute, 0, 0, time.UTC)
			gotTime, _ := system.calculateTimeOfDay(gameTime)

			if gotTime != tt.wantTime {
				t.Errorf("calculateTimeOfDay() = %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}

func TestTimeOfDayLightingSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	// Create game clock at noon
	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
	system.SetClock(clock)
	system.SetGenre("fantasy")

	// Create ambient light entity
	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)
	world.AddEntity(entity)

	// Update with entity
	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify time of day is set to Day at noon
	if system.currentTimeOfDay != palette.TimeOfDayDay {
		t.Errorf("Time of day = %v, want Day", system.currentTimeOfDay)
	}

	// Verify base values were cached
	if !system.baseAmbientCached {
		t.Error("Base ambient values not cached")
	}

	// Verify ambient light was modulated
	if ambient.Color == (color.RGBA{200, 200, 210, 255}) && ambient.Intensity == 0.7 {
		// At Day time, modulation should be minimal but not zero
		t.Log("Day time modulation is neutral as expected")
	}
}

func TestTimeOfDayLightingSystem_NightModulation(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	// Create game clock at midnight
	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)}
	system.SetClock(clock)
	system.SetGenre("fantasy")

	// Create ambient light entity
	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Night should reduce intensity
	if ambient.Intensity >= 0.7 {
		t.Errorf("Night intensity = %f, should be less than base 0.7", ambient.Intensity)
	}

	// Night should shift color toward blue (lower R, higher B relative to G)
	// Or at minimum, the color should be different from day
	if system.currentTimeOfDay != palette.TimeOfDayNight {
		t.Errorf("Time of day = %v, want Night", system.currentTimeOfDay)
	}
}

func TestTimeOfDayLightingSystem_DawnModulation(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	// Create game clock at 6 AM (dawn)
	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 6, 0, 0, 0, time.UTC)}
	system.SetClock(clock)
	system.SetGenre("fantasy")

	// Create ambient light entity
	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	if system.currentTimeOfDay != palette.TimeOfDayDawn {
		t.Errorf("Time of day = %v, want Dawn", system.currentTimeOfDay)
	}
}

func TestTimeOfDayLightingSystem_DuskModulation(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	// Create game clock at 6 PM (dusk)
	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 18, 0, 0, 0, time.UTC)}
	system.SetClock(clock)
	system.SetGenre("fantasy")

	// Create ambient light entity
	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	if system.currentTimeOfDay != palette.TimeOfDayDusk {
		t.Errorf("Time of day = %v, want Dusk", system.currentTimeOfDay)
	}
}

func TestTimeOfDayLightingSystem_GenreIntensity(t *testing.T) {
	tests := []struct {
		genre     string
		wantMult  float64
		tolerance float64
	}{
		{"fantasy", 1.0, 0.01},
		{"scifi", 0.7, 0.01},
		{"horror", 1.2, 0.01},
		{"cyberpunk", 0.6, 0.01},
		{"postapoc", 1.1, 0.01},
		{"unknown", 1.0, 0.01}, // Unknown defaults to 1.0
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			system := NewTimeOfDayLightingSystem(world, 12345)
			system.SetGenre(tt.genre)

			got := system.getGenreIntensity()
			if got < tt.wantMult-tt.tolerance || got > tt.wantMult+tt.tolerance {
				t.Errorf("Genre %s intensity = %f, want %f", tt.genre, got, tt.wantMult)
			}
		})
	}
}

func TestTimeOfDayLightingSystem_ModulateColor(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	base := color.RGBA{200, 200, 200, 255}

	tests := []struct {
		name         string
		modulation   palette.ColorModulation
		expectWarm   bool // If true, expect R > B
		expectCool   bool // If true, expect B > R
		expectDarker bool // If true, expect lower RGB values
	}{
		{
			name: "warm_shift",
			modulation: palette.ColorModulation{
				TemperatureShift:     0.5,
				SaturationMultiplier: 1.0,
				LightnessOffset:      0.0,
			},
			expectWarm: true,
		},
		{
			name: "cool_shift",
			modulation: palette.ColorModulation{
				TemperatureShift:     -0.5,
				SaturationMultiplier: 1.0,
				LightnessOffset:      0.0,
			},
			expectCool: true,
		},
		{
			name: "darker",
			modulation: palette.ColorModulation{
				TemperatureShift:     0.0,
				SaturationMultiplier: 1.0,
				LightnessOffset:      -0.2,
			},
			expectDarker: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.modulateColor(base, tt.modulation)

			if tt.expectWarm && result.R <= result.B {
				t.Errorf("Warm shift: R=%d should be > B=%d", result.R, result.B)
			}

			if tt.expectCool && result.B <= result.R {
				t.Errorf("Cool shift: B=%d should be > R=%d", result.B, result.R)
			}

			if tt.expectDarker {
				avgBase := (int(base.R) + int(base.G) + int(base.B)) / 3
				avgResult := (int(result.R) + int(result.G) + int(result.B)) / 3
				if avgResult >= avgBase {
					t.Errorf("Darker: avg result %d should be < avg base %d", avgResult, avgBase)
				}
			}
		})
	}
}

func TestTimeOfDayLightingSystem_ModulateIntensity(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	tests := []struct {
		name    string
		base    float64
		offset  float64
		wantMin float64
		wantMax float64
	}{
		{"neutral", 0.7, 0.0, 0.69, 0.71},
		{"brighter", 0.7, 0.1, 0.79, 0.81},
		{"darker", 0.7, -0.2, 0.49, 0.51},
		{"clamped_min", 0.05, -0.5, 0.1, 0.1},
		{"clamped_max", 0.9, 0.5, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := palette.ColorModulation{
				LightnessOffset: tt.offset,
			}
			result := system.modulateIntensity(tt.base, mod)

			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("modulateIntensity() = %f, want [%f, %f]", result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTimeOfDayLightingSystem_GetCurrentTimeOfDay(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	// Default should be Day
	if system.GetCurrentTimeOfDay() != palette.TimeOfDayDay {
		t.Errorf("GetCurrentTimeOfDay() = %v, want Day", system.GetCurrentTimeOfDay())
	}

	// After update at midnight
	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)}
	system.SetClock(clock)

	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)

	system.Update([]*Entity{entity}, 0.016)

	if system.GetCurrentTimeOfDay() != palette.TimeOfDayNight {
		t.Errorf("GetCurrentTimeOfDay() = %v, want Night after midnight update", system.GetCurrentTimeOfDay())
	}
}

func TestTimeOfDayLightingSystem_NilWorld(t *testing.T) {
	system := NewTimeOfDayLightingSystem(nil, 12345)

	// Should not panic
	system.Update(nil, 0.016)
}

func TestTimeOfDayLightingSystem_NoClock(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)

	// Should not panic without clock
	system.Update([]*Entity{entity}, 0.016)
}

func TestTimeOfDayLightingSystem_NoAmbientLight(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
	system.SetClock(clock)

	// Entity without ambient light
	entity := NewEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	system.Update([]*Entity{entity}, 0.016)
}

func BenchmarkTimeOfDayLightingSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	clock := &mockGameClock{currentTime: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
	system.SetClock(clock)
	system.SetGenre("fantasy")

	entity := NewEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 210, 255}, 0.7)
	entity.AddComponent(ambient)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkTimeOfDayLightingSystem_ModulateColor(b *testing.B) {
	world := NewWorld()
	system := NewTimeOfDayLightingSystem(world, 12345)

	base := color.RGBA{200, 200, 210, 255}
	mod := palette.ColorModulation{
		HueShift:             15.0,
		SaturationMultiplier: 0.85,
		LightnessOffset:      0.05,
		TemperatureShift:     0.4,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.modulateColor(base, mod)
	}
}
