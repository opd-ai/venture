package engine

import (
	"math"
	"testing"
)

func TestDeathDissolveComponentType(t *testing.T) {
	c := &DeathDissolveComponent{}
	if got := c.Type(); got != "death_dissolve" {
		t.Errorf("Type() = %q, want %q", got, "death_dissolve")
	}
}

func TestNewEntityDeathDissolveSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 42)
	if sys == nil {
		t.Fatal("NewEntityDeathDissolveSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("default genre = %q, want %q", sys.genreID, "fantasy")
	}
	if len(sys.presets) != 5 {
		t.Errorf("preset count = %d, want 5", len(sys.presets))
	}
}

func TestEntityDeathDissolveSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 42)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"horror"},
		{"scifi"},
		{"cyberpunk"},
		{"postapoc"},
	}
	for _, tt := range tests {
		sys.SetGenre(tt.genre)
		if sys.genreID != tt.genre {
			t.Errorf("SetGenre(%q): genreID = %q", tt.genre, sys.genreID)
		}
	}
}

func TestEntityDeathDissolveSystem_DetectDeadEntities(t *testing.T) {
	tests := []struct {
		name      string
		health    float64
		hasSprite bool
		wantComp  bool
	}{
		{"alive entity", 50.0, true, false},
		{"dead entity with sprite", 0.0, true, true},
		{"dead entity without sprite", 0.0, false, false},
		{"negative health with sprite", -10.0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEntityDeathDissolveSystem(world, 123)

			entity := NewEntity(1)
			entity.AddComponent(&HealthComponent{Current: tt.health, Max: 100})
			if tt.hasSprite {
				entity.AddComponent(&StubSprite{})
			}

			entities := []*Entity{entity}
			// Run update with enough time to trigger scan
			sys.Update(entities, 0.2)

			_, hasDissolve := entity.GetComponent("death_dissolve")
			if hasDissolve != tt.wantComp {
				t.Errorf("has death_dissolve = %v, want %v", hasDissolve, tt.wantComp)
			}
		})
	}
}

func TestEntityDeathDissolveSystem_AnimateDissolve(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 99)

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})
	entity.AddComponent(&StubSprite{})
	entities := []*Entity{entity}

	// First update attaches the component
	sys.Update(entities, 0.2)

	comp, ok := entity.GetComponent("death_dissolve")
	if !ok {
		t.Fatal("death_dissolve component not attached")
	}
	dc := comp.(*DeathDissolveComponent)

	if dc.Opacity < 0.9 {
		t.Errorf("initial opacity = %f, want ~1.0", dc.Opacity)
	}

	// Animate for half the duration
	halfDuration := dc.Duration / 2
	sys.Update(entities, halfDuration)
	if dc.Opacity >= 1.0 || dc.Opacity <= 0.0 {
		t.Errorf("mid-dissolve opacity = %f, want between 0 and 1", dc.Opacity)
	}
	if dc.Complete {
		t.Error("should not be complete at half duration")
	}

	// Animate past the full duration
	sys.Update(entities, dc.Duration+0.5)
	if dc.Opacity != 0.0 {
		t.Errorf("final opacity = %f, want 0.0", dc.Opacity)
	}
	if !dc.Complete {
		t.Error("should be complete after full duration")
	}
}

func TestEntityDeathDissolveSystem_NoDoubleAttach(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 55)

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})
	entity.AddComponent(&StubSprite{})
	entities := []*Entity{entity}

	// First scan attaches the component
	sys.Update(entities, 0.2)
	comp, _ := entity.GetComponent("death_dissolve")
	dc := comp.(*DeathDissolveComponent)
	origDuration := dc.Duration

	// Second scan should not replace the component
	sys.Update(entities, 0.2)
	comp2, _ := entity.GetComponent("death_dissolve")
	dc2 := comp2.(*DeathDissolveComponent)
	if dc2.Duration != origDuration {
		t.Error("component was replaced on second scan")
	}
}

func TestEntityDeathDissolveSystem_GenrePresets(t *testing.T) {
	tests := []struct {
		genre    string
		wantR    float64
		wantG    float64
		wantB    float64
		style    int
		duration float64
	}{
		{"fantasy", 0.95, 0.85, 0.40, 0, 0.8},
		{"horror", 0.60, 0.05, 0.05, 3, 1.2},
		{"scifi", 0.15, 0.80, 0.90, 2, 0.5},
		{"cyberpunk", 0.85, 0.15, 0.75, 2, 0.45},
		{"postapoc", 0.55, 0.45, 0.30, 4, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			presets := buildDissolvePresets()
			p, ok := presets[tt.genre]
			if !ok {
				t.Fatalf("missing preset for %q", tt.genre)
			}
			if p.Style != tt.style {
				t.Errorf("style = %d, want %d", p.Style, tt.style)
			}
			if math.Abs(p.Duration-tt.duration) > 0.01 {
				t.Errorf("duration = %f, want %f", p.Duration, tt.duration)
			}
			if math.Abs(p.R-tt.wantR) > 0.01 || math.Abs(p.G-tt.wantG) > 0.01 || math.Abs(p.B-tt.wantB) > 0.01 {
				t.Errorf("color = (%f,%f,%f), want (%f,%f,%f)", p.R, p.G, p.B, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestEntityDeathDissolveSystem_ZeroDuration(t *testing.T) {
	dc := &DeathDissolveComponent{
		Duration: 0,
		Opacity:  1.0,
	}
	entity := NewEntity(1)
	entity.AddComponent(dc)
	entity.AddComponent(&StubSprite{})
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})

	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 77)
	sys.Update([]*Entity{entity}, 0.016)

	if !dc.Complete {
		t.Error("zero-duration dissolve should immediately complete")
	}
	if dc.Opacity != 0.0 {
		t.Errorf("zero-duration opacity = %f, want 0.0", dc.Opacity)
	}
}

func TestEntityDeathDissolveSystem_FallbackGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 88)
	sys.SetGenre("unknown_genre")

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})
	entity.AddComponent(&StubSprite{})

	sys.Update([]*Entity{entity}, 0.2)

	_, ok := entity.GetComponent("death_dissolve")
	if !ok {
		t.Error("should fall back to fantasy preset for unknown genre")
	}
}

func TestClampDissolve(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		got := clampDissolve(tt.input)
		if got != tt.want {
			t.Errorf("clampDissolve(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func BenchmarkEntityDeathDissolveSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewEntityDeathDissolveSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&StubSprite{})
		if i%5 == 0 {
			// 20% dead entities
			e.AddComponent(&HealthComponent{Current: 0, Max: 100})
		} else {
			e.AddComponent(&HealthComponent{Current: 50, Max: 100})
		}
		entities[i] = e
	}

	// First pass to attach components
	sys.Update(entities, 0.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
