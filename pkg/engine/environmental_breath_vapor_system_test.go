package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestBreathVaporComponentType(t *testing.T) {
	c := &BreathVaporComponent{}
	if c.Type() != "breath_vapor" {
		t.Errorf("expected 'breath_vapor', got %q", c.Type())
	}
}

func TestNewEnvironmentalBreathVaporSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEnvironmentalBreathVaporSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.coldWeatherActive {
		t.Error("expected cold weather inactive by default")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		wantR   float64
		wantG   float64
		wantB   float64
		opacity float64
	}{
		{"fantasy", "fantasy", 0.9, 0.93, 1.0, 0.4},
		{"horror", "horror", 0.7, 0.75, 0.8, 0.55},
		{"cyberpunk", "cyberpunk", 0.6, 0.85, 1.0, 0.45},
		{"scifi", "scifi", 0.8, 0.9, 1.0, 0.35},
		{"postapoc", "postapoc", 0.75, 0.72, 0.68, 0.5},
		{"unknown", "steampunk", 0.9, 0.93, 1.0, 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEnvironmentalBreathVaporSystem(world, 42)
			sys.SetGenre(tt.genre)

			if sys.preset.R != tt.wantR {
				t.Errorf("R: got %f, want %f", sys.preset.R, tt.wantR)
			}
			if sys.preset.G != tt.wantG {
				t.Errorf("G: got %f, want %f", sys.preset.G, tt.wantG)
			}
			if sys.preset.B != tt.wantB {
				t.Errorf("B: got %f, want %f", sys.preset.B, tt.wantB)
			}
			if sys.preset.Opacity != tt.opacity {
				t.Errorf("Opacity: got %f, want %f", sys.preset.Opacity, tt.opacity)
			}
		})
	}
}

func TestDetectColdWeather(t *testing.T) {
	tests := []struct {
		name       string
		weatherTyp particles.WeatherType
		active     bool
		wantCold   bool
	}{
		{"snow active", particles.WeatherSnow, true, true},
		{"fog active", particles.WeatherFog, true, true},
		{"ash active", particles.WeatherAsh, true, true},
		{"rain active", particles.WeatherRain, false, false},
		{"rain not cold", particles.WeatherRain, true, false},
		{"snow inactive", particles.WeatherSnow, false, false},
		{"dust not cold", particles.WeatherDust, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEnvironmentalBreathVaporSystem(world, 42)

			weatherEntity := NewEntity(1)
			wc := NewWeatherComponent(particles.WeatherConfig{Type: tt.weatherTyp})
			wc.Active = tt.active
			weatherEntity.AddComponent(wc)

			entities := []*Entity{weatherEntity}
			got := sys.detectColdWeather(entities)
			if got != tt.wantCold {
				t.Errorf("detectColdWeather() = %v, want %v", got, tt.wantCold)
			}
		})
	}
}

func TestUpdateSpawnsPuffsInColdWeather(t *testing.T) {
	world := NewWorld()
	sys := NewEnvironmentalBreathVaporSystem(world, 42)
	sys.SetGenre("fantasy")

	// Create weather entity with snow
	weatherEntity := NewEntity(1)
	wc := NewWeatherComponent(particles.WeatherConfig{Type: particles.WeatherSnow})
	wc.Active = true
	weatherEntity.AddComponent(wc)

	// Create character entity with position + sprite
	charEntity := NewEntity(2)
	charEntity.AddComponent(&PositionComponent{X: 50, Y: 50})
	charEntity.AddComponent(&StubSprite{Visible: true})

	entities := []*Entity{weatherEntity, charEntity}

	// Force weather check on first update
	sys.timeSinceCheck = sys.weatherCheckInterval

	// Run multiple updates to let cooldown expire and puffs spawn
	for i := 0; i < 20; i++ {
		sys.Update(entities, 0.1)
	}

	// Check puffs were spawned
	comp, ok := charEntity.GetComponent("breath_vapor")
	if !ok {
		t.Fatal("expected breath_vapor component to be created")
	}
	vapor, ok := comp.(*BreathVaporComponent)
	if !ok {
		t.Fatal("expected *BreathVaporComponent")
	}
	if !vapor.Active {
		t.Error("expected vapor to be active in cold weather")
	}
	if len(vapor.Puffs) == 0 {
		t.Error("expected at least one puff to be spawned")
	}
}

func TestUpdateNoPuffsInWarmWeather(t *testing.T) {
	world := NewWorld()
	sys := NewEnvironmentalBreathVaporSystem(world, 42)

	// Create weather entity with rain (warm)
	weatherEntity := NewEntity(1)
	wc := NewWeatherComponent(particles.WeatherConfig{Type: particles.WeatherRain})
	wc.Active = true
	weatherEntity.AddComponent(wc)

	charEntity := NewEntity(2)
	charEntity.AddComponent(&PositionComponent{X: 50, Y: 50})
	charEntity.AddComponent(&StubSprite{Visible: true})

	entities := []*Entity{weatherEntity, charEntity}

	sys.timeSinceCheck = sys.weatherCheckInterval
	for i := 0; i < 20; i++ {
		sys.Update(entities, 0.1)
	}

	comp, ok := charEntity.GetComponent("breath_vapor")
	if !ok {
		return // No component created is valid
	}
	vapor := comp.(*BreathVaporComponent)
	if vapor.Active {
		t.Error("expected vapor to be inactive in warm weather")
	}
}

func TestPuffsExpireOverTime(t *testing.T) {
	world := NewWorld()
	sys := NewEnvironmentalBreathVaporSystem(world, 42)
	sys.SetGenre("fantasy")

	vapor := &BreathVaporComponent{
		Active:   true,
		Cooldown: 100.0, // prevent new spawns
		Puffs: []BreathPuff{
			{X: 10, Y: 10, VY: -5, Age: 0, MaxAge: 0.5, Opacity: 0.4, Size: 2},
		},
	}

	// Age the puff past its max
	sys.agePuffs(vapor, 0.6)
	if len(vapor.Puffs) != 0 {
		t.Errorf("expected 0 puffs after expiry, got %d", len(vapor.Puffs))
	}
}

func TestPuffCapLimit(t *testing.T) {
	world := NewWorld()
	sys := NewEnvironmentalBreathVaporSystem(world, 42)

	vapor := &BreathVaporComponent{
		Active: true,
		Puffs:  make([]BreathPuff, 4, 4), // already at cap
	}
	pos := &PositionComponent{X: 50, Y: 50}

	sys.spawnPuff(vapor, pos)
	if len(vapor.Puffs) != 4 {
		t.Errorf("expected puff cap at 4, got %d", len(vapor.Puffs))
	}
}

func TestGetActiveVaporPuffs(t *testing.T) {
	entity := NewEntity(1)

	// No component
	puffs := GetActiveVaporPuffs(entity)
	if puffs != nil {
		t.Error("expected nil puffs when no component")
	}

	// Inactive component
	v := &BreathVaporComponent{Active: false, Puffs: []BreathPuff{{X: 1}}}
	entity.AddComponent(v)
	puffs = GetActiveVaporPuffs(entity)
	if puffs != nil {
		t.Error("expected nil puffs when inactive")
	}

	// Active component with puffs
	v.Active = true
	puffs = GetActiveVaporPuffs(entity)
	if len(puffs) != 1 {
		t.Errorf("expected 1 puff, got %d", len(puffs))
	}
}

func TestUpdateWithNilWorld(t *testing.T) {
	sys := &EnvironmentalBreathVaporSystem{}
	// Should not panic
	sys.Update([]*Entity{}, 0.016)
}

func TestNoSpriteEntitySkipped(t *testing.T) {
	world := NewWorld()
	sys := NewEnvironmentalBreathVaporSystem(world, 42)
	sys.coldWeatherActive = true

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	// No sprite component

	sys.processEntity(entity, 0.1)

	_, ok := entity.GetComponent("breath_vapor")
	if ok {
		t.Error("expected no breath_vapor on entity without sprite")
	}
}
