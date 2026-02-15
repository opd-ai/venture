//go:build ignore

package engine

import (
	"testing"
)

func TestStatusEffectGroundTrailComponentType(t *testing.T) {
	c := &StatusEffectGroundTrailComponent{DropTimers: make(map[string]float64)}
	if c.Type() != "status_effect_ground_trail" {
		t.Errorf("expected 'status_effect_ground_trail', got %q", c.Type())
	}
}

func TestNewStatusEffectGroundTrailSystem(t *testing.T) {
	sys := NewStatusEffectGroundTrailSystem(nil, 12345)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", sys.seed)
	}
	if sys.dropInterval <= 0 {
		t.Errorf("expected positive dropInterval, got %f", sys.dropInterval)
	}
	if len(sys.trailEffectTypes) != 4 {
		t.Errorf("expected 4 trail effect types, got %d", len(sys.trailEffectTypes))
	}
}

func TestStatusEffectGroundTrailSystem_SetGenre(t *testing.T) {
	sys := NewStatusEffectGroundTrailSystem(nil, 1)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestStatusEffectGroundTrailSystem_IsTrailEffect(t *testing.T) {
	sys := NewStatusEffectGroundTrailSystem(nil, 1)
	tests := []struct {
		effectType string
		want       bool
	}{
		{"burning", true},
		{"poisoned", true},
		{"bleeding", true},
		{"frozen", true},
		{"stunned", false},
		{"haste", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			got := sys.isTrailEffect(tt.effectType)
			if got != tt.want {
				t.Errorf("isTrailEffect(%q) = %v, want %v", tt.effectType, got, tt.want)
			}
		})
	}
}

func TestStatusEffectGroundTrailSystem_GetPreset(t *testing.T) {
	sys := NewStatusEffectGroundTrailSystem(nil, 1)
	tests := []struct {
		name       string
		effectType string
		genre      string
		wantValid  bool
	}{
		{"burning_fantasy", "burning", "fantasy", true},
		{"poisoned_scifi", "poisoned", "scifi", true},
		{"bleeding_horror", "bleeding", "horror", true},
		{"frozen_cyberpunk", "frozen", "cyberpunk", true},
		{"unknown_effect", "stunned", "fantasy", true},  // fallback
		{"unknown_genre", "burning", "steampunk", true}, // falls back to fantasy
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := sys.getPreset(tt.effectType, tt.genre)
			if tt.wantValid && preset.duration <= 0 {
				t.Errorf("expected positive duration for %s/%s", tt.effectType, tt.genre)
			}
		})
	}
}

func TestStatusEffectGroundTrailSystem_UpdateNoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectGroundTrailSystem(world, 1)
	// No particle system set — should return without panic
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StatusEffectComponent{EffectType: "burning", Duration: 5.0, Magnitude: 1.0})
	sys.Update([]*Entity{entity}, 0.016)
}

func TestStatusEffectGroundTrailSystem_UpdateInitializesComponent(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectGroundTrailSystem(world, 42)
	sys.SetGenre("fantasy")
	// Set a non-nil particle system stub via the internal field for testing
	sys.particleSystem = NewParticleSystem()

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&StatusEffectComponent{EffectType: "poisoned", Duration: 3.0, Magnitude: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("status_effect_ground_trail")
	if !ok {
		t.Fatal("expected ground trail component to be attached")
	}
	trail := comp.(*StatusEffectGroundTrailComponent)
	if !trail.Initialized {
		t.Error("expected Initialized to be true after first update")
	}
	if trail.PrevX != 50 || trail.PrevY != 50 {
		t.Errorf("expected prev position (50,50), got (%f,%f)", trail.PrevX, trail.PrevY)
	}
}

func TestStatusEffectGroundTrailSystem_SkipsExpiredEffects(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectGroundTrailSystem(world, 42)
	sys.SetGenre("fantasy")
	sys.particleSystem = NewParticleSystem()

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&StatusEffectComponent{EffectType: "burning", Duration: -1.0, Magnitude: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("status_effect_ground_trail")
	if ok {
		t.Error("expected no ground trail component for expired effect")
	}
}

func TestStatusEffectGroundTrailSystem_SkipsNonTrailEffects(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectGroundTrailSystem(world, 42)
	sys.SetGenre("fantasy")
	sys.particleSystem = NewParticleSystem()

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&StatusEffectComponent{EffectType: "stunned", Duration: 3.0, Magnitude: 1.0})

	sys.Update([]*Entity{entity}, 0.016)

	_, ok := entity.GetComponent("status_effect_ground_trail")
	if ok {
		t.Error("expected no ground trail component for non-trail effect")
	}
}

func TestStatusEffectGroundTrailSystem_MovementTriggersTrail(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectGroundTrailSystem(world, 42)
	sys.SetGenre("fantasy")
	sys.particleSystem = NewParticleSystem()

	entity := world.CreateEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	entity.AddComponent(pos)
	entity.AddComponent(&StatusEffectComponent{EffectType: "bleeding", Duration: 5.0, Magnitude: 2.0})

	// First update: initialize
	sys.Update([]*Entity{entity}, 0.016)

	// Move entity far enough
	pos.X = 80
	pos.Y = 80

	// Second update: should detect movement and update PrevX/PrevY
	sys.Update([]*Entity{entity}, 0.5) // enough time to pass cooldown

	comp, _ := entity.GetComponent("status_effect_ground_trail")
	trail := comp.(*StatusEffectGroundTrailComponent)

	if trail.PrevX != 80 || trail.PrevY != 80 {
		t.Errorf("expected prev position updated to (80,80), got (%f,%f)", trail.PrevX, trail.PrevY)
	}
}

func TestStatusEffectGroundTrailSystem_DropTimerCooldown(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectGroundTrailSystem(world, 42)
	sys.SetGenre("scifi")
	sys.particleSystem = NewParticleSystem()

	entity := world.CreateEntity()
	pos := &PositionComponent{X: 10, Y: 10}
	entity.AddComponent(pos)
	entity.AddComponent(&StatusEffectComponent{EffectType: "frozen", Duration: 10.0, Magnitude: 1.0})

	// Initialize
	sys.Update([]*Entity{entity}, 0.016)

	// Move and trigger first drop
	pos.X = 50
	pos.Y = 50
	sys.Update([]*Entity{entity}, 0.5)

	comp, _ := entity.GetComponent("status_effect_ground_trail")
	trail := comp.(*StatusEffectGroundTrailComponent)
	timer := trail.DropTimers["frozen"]
	if timer <= 0 {
		t.Error("expected positive drop timer after trail spawn")
	}

	// Move again but within cooldown
	pos.X = 90
	pos.Y = 90
	sys.Update([]*Entity{entity}, 0.01) // tiny delta, timer still positive

	timer2 := trail.DropTimers["frozen"]
	if timer2 >= timer {
		t.Error("expected timer to decrease after update")
	}
}

func TestDefaultGroundTrailPresets(t *testing.T) {
	presets := defaultGroundTrailPresets()
	effects := []string{"burning", "poisoned", "bleeding", "frozen"}
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, effect := range effects {
		for _, genre := range genres {
			t.Run(effect+"_"+genre, func(t *testing.T) {
				p, ok := presets[effect][genre]
				if !ok {
					t.Fatalf("missing preset for %s/%s", effect, genre)
				}
				if p.baseCount < 1 {
					t.Errorf("expected baseCount >= 1, got %d", p.baseCount)
				}
				if p.duration <= 0 {
					t.Errorf("expected positive duration, got %f", p.duration)
				}
			})
		}
	}
}
