package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestWetnessComponentType(t *testing.T) {
	c := &WetnessComponent{}
	if got := c.Type(); got != "wetness" {
		t.Errorf("Type() = %q, want %q", got, "wetness")
	}
}

func TestNewWeatherEntityWetnessSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	if sys == nil {
		t.Fatal("NewWeatherEntityWetnessSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestWeatherEntityWetnessSetGenre(t *testing.T) {
	genres := []struct {
		name      string
		maxDarken float64
	}{
		{"fantasy", 0.2},
		{"horror", 0.28},
		{"scifi", 0.15},
		{"cyberpunk", 0.22},
		{"postapoc", 0.3},
		{"unknown", 0.2},
	}
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	for _, tt := range genres {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.name)
			if sys.preset.MaxDarken != tt.maxDarken {
				t.Errorf("genre %q MaxDarken = %v, want %v", tt.name, sys.preset.MaxDarken, tt.maxDarken)
			}
		})
	}
}

func TestWeatherEntityWetnessRainAccumulation(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)

	// Create weather entity with active rain
	weatherEntity := NewEntity(0)
	wc := NewWeatherComponent(particles.WeatherConfig{Type: particles.WeatherRain})
	wc.Active = true
	weatherEntity.AddComponent(wc)

	// Create a visible entity
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{weatherEntity, entity}

	// Force weather check immediately
	sys.timeSinceCheck = sys.checkInterval

	// Run multiple frames
	for i := 0; i < 60; i++ {
		sys.Update(entities, 0.016)
	}

	wet := getWetnessComponent(entity)
	if wet == nil {
		t.Fatal("entity missing WetnessComponent after Update")
	}
	if !wet.Active {
		t.Error("expected Active=true during rain")
	}
	if wet.Level <= 0 {
		t.Errorf("expected positive wetness level, got %v", wet.Level)
	}
	if wet.DarkenAmount <= 0 {
		t.Error("expected positive DarkenAmount during rain")
	}
}

func TestWeatherEntityWetnessDrying(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)

	// Create weather entity with active rain
	weatherEntity := NewEntity(0)
	wc := NewWeatherComponent(particles.WeatherConfig{Type: particles.WeatherRain})
	wc.Active = true
	weatherEntity.AddComponent(wc)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{weatherEntity, entity}

	// Force weather check and build up wetness
	sys.timeSinceCheck = sys.checkInterval
	for i := 0; i < 100; i++ {
		sys.Update(entities, 0.016)
	}

	wet := getWetnessComponent(entity)
	if wet == nil {
		t.Fatal("missing WetnessComponent")
	}
	peakLevel := wet.Level
	if peakLevel <= 0 {
		t.Fatal("wetness should have accumulated")
	}

	// Stop rain
	wc.Active = false
	sys.timeSinceCheck = sys.checkInterval

	// Dry off
	for i := 0; i < 200; i++ {
		sys.Update(entities, 0.016)
	}

	if wet.Level >= peakLevel {
		t.Errorf("wetness should decrease, got %v >= %v", wet.Level, peakLevel)
	}
}

func TestWeatherEntityWetnessFullDry(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	// Pre-attach wetness with some level
	wet := &WetnessComponent{Level: 0.5, Active: true}
	entity.AddComponent(wet)

	entities := []*Entity{entity}

	// No rain — should dry to zero
	sys.rainActive = false
	sys.timeSinceCheck = 0 // Skip weather scan (no weather entity)

	for i := 0; i < 500; i++ {
		sys.Update(entities, 0.016)
	}

	if wet.Active {
		t.Error("expected Active=false after full drying")
	}
	if wet.Level != 0 {
		t.Errorf("expected Level=0, got %v", wet.Level)
	}
	if wet.DarkenAmount != 0 {
		t.Errorf("expected DarkenAmount=0, got %v", wet.DarkenAmount)
	}
}

func TestWeatherEntityWetnessZeroDelta(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)

	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})
	entities := []*Entity{entity}

	sys.Update(entities, 0)
	if _, ok := entity.GetComponent("wetness"); ok {
		t.Error("should not create component with zero delta")
	}
}

func TestWeatherEntityWetnessNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	sys.rainActive = true

	entity := NewEntity(0)
	entity.AddComponent(&StubSprite{})
	entities := []*Entity{entity}

	sys.Update(entities, 0.016)
	if _, ok := entity.GetComponent("wetness"); ok {
		t.Error("should not add wetness to entity without position")
	}
}

func TestWeatherEntityWetnessNoSprite(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	sys.rainActive = true

	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entities := []*Entity{entity}

	sys.Update(entities, 0.016)
	if _, ok := entity.GetComponent("wetness"); ok {
		t.Error("should not add wetness to entity without sprite")
	}
}

func TestWeatherEntityWetnessLevelCapped(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	sys.rainActive = true

	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{entity}

	// Run for a long time
	for i := 0; i < 1000; i++ {
		sys.Update(entities, 0.016)
	}

	wet := getWetnessComponent(entity)
	if wet == nil {
		t.Fatal("missing WetnessComponent")
	}
	if wet.Level > 1.0 {
		t.Errorf("wetness level should cap at 1.0, got %v", wet.Level)
	}
}

func TestWeatherEntityWetnessDarkenProportional(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	sys.rainActive = true

	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{entity}

	// Run to full saturation
	for i := 0; i < 1000; i++ {
		sys.Update(entities, 0.016)
	}

	wet := getWetnessComponent(entity)
	if wet == nil {
		t.Fatal("missing WetnessComponent")
	}
	expectedDarken := wet.Level * sys.preset.MaxDarken
	if math.Abs(wet.DarkenAmount-expectedDarken) > 0.001 {
		t.Errorf("DarkenAmount = %v, want %v", wet.DarkenAmount, expectedDarken)
	}
}

func TestWeatherEntityWetnessBloodRain(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)

	weatherEntity := NewEntity(0)
	wc := NewWeatherComponent(particles.WeatherConfig{Type: particles.WeatherBloodRain})
	wc.Active = true
	weatherEntity.AddComponent(wc)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{weatherEntity, entity}
	sys.timeSinceCheck = sys.checkInterval

	for i := 0; i < 60; i++ {
		sys.Update(entities, 0.016)
	}

	wet := getWetnessComponent(entity)
	if wet == nil {
		t.Fatal("missing WetnessComponent")
	}
	if !wet.Active {
		t.Error("blood rain should activate wetness")
	}
}

func TestWeatherEntityWetnessNonRainWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)

	weatherEntity := NewEntity(0)
	wc := NewWeatherComponent(particles.WeatherConfig{Type: particles.WeatherFog})
	wc.Active = true
	weatherEntity.AddComponent(wc)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{weatherEntity, entity}
	sys.timeSinceCheck = sys.checkInterval

	for i := 0; i < 60; i++ {
		sys.Update(entities, 0.016)
	}

	wet := getWetnessComponent(entity)
	if wet == nil {
		return // Component may not be created if no rain detected
	}
	if wet.Active {
		t.Error("fog should not cause wetness")
	}
}

func TestWeatherEntityWetnessLargeDeltaClamped(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	sys.rainActive = true

	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StubSprite{})

	entities := []*Entity{entity}

	sys.Update(entities, 5.0) // Large delta should be clamped to 0.1

	wet := getWetnessComponent(entity)
	if wet == nil {
		t.Fatal("missing WetnessComponent")
	}
	// With clamped dt=0.1 and WetRate=0.15, max accumulation is 0.015
	if wet.Level > 0.02 {
		t.Errorf("clamped delta should limit accumulation, got %v", wet.Level)
	}
}

func TestWeatherEntityWetnessSheenTint(t *testing.T) {
	genres := []struct {
		name  string
		tintR float64
	}{
		{"fantasy", 0.85},
		{"horror", 0.9},
		{"cyberpunk", 0.7},
	}

	for _, tt := range genres {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherEntityWetnessSystem(world, 42)
			sys.SetGenre(tt.name)
			sys.rainActive = true

			entity := NewEntity(0)
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&StubSprite{})

			entities := []*Entity{entity}

			for i := 0; i < 100; i++ {
				sys.Update(entities, 0.016)
			}

			wet := getWetnessComponent(entity)
			if wet == nil {
				t.Fatal("missing WetnessComponent")
			}
			// Sheen tint R = TintR * Level * SheenIntensity
			expectedR := tt.tintR * wet.Level * sys.preset.SheenIntensity
			if math.Abs(wet.SheenTintR-expectedR) > 0.01 {
				t.Errorf("SheenTintR = %v, want ~%v", wet.SheenTintR, expectedR)
			}
		})
	}
}

func BenchmarkWeatherEntityWetnessSystem(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherEntityWetnessSystem(world, 42)
	sys.rainActive = true

	entities := make([]*Entity, 500)
	for i := range entities {
		e := NewEntity(0)
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		e.AddComponent(&StubSprite{})
		e.AddComponent(&WetnessComponent{})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func getWetnessComponent(entity *Entity) *WetnessComponent {
	comp, ok := entity.GetComponent("wetness")
	if !ok {
		return nil
	}
	wet, _ := comp.(*WetnessComponent)
	return wet
}
