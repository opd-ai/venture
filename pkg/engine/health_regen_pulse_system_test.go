package engine

import (
	"testing"
)

func TestNewHealthRegenPulseSystem(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", system.seed)
	}
	if system.world != world {
		t.Error("expected world to be set")
	}
	if system.rng == nil {
		t.Error("expected non-nil rng")
	}
	if system.pulseInterval <= 0 {
		t.Error("expected positive pulse interval")
	}
}

func TestNewHealthRegenPulseSystem_NilWorld(t *testing.T) {
	system := NewHealthRegenPulseSystem(nil, 99)
	if system == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
	// Should not panic on Update with nil world
	system.Update(nil, 0.016)
}

func TestHealthRegenPulseSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name   string
		genre  string
		wantPT string
	}{
		{"fantasy", "fantasy", "sparkle"},
		{"scifi", "scifi", "magic"},
		{"horror", "horror", "ember"},
		{"cyberpunk", "cyberpunk", "spark"},
		{"postapoc", "postapoc", "dust"},
		{"unknown defaults to fantasy", "unknown_genre", "sparkle"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewHealthRegenPulseSystem(NewWorld(), 42)
			system.SetGenre(tt.genre)
			if system.genreID != tt.genre {
				t.Errorf("expected genreID %q, got %q", tt.genre, system.genreID)
			}
			got := system.preset.particleType.String()
			if got != tt.wantPT {
				t.Errorf("expected particle type %q for genre %q, got %q", tt.wantPT, tt.genre, got)
			}
		})
	}
}

func TestHealthRegenPulseComponent_Type(t *testing.T) {
	comp := &HealthRegenPulseComponent{}
	if comp.Type() != "health_regen_pulse" {
		t.Errorf("expected type 'health_regen_pulse', got %q", comp.Type())
	}
}

func TestHealthRegenPulseSystem_DetectsHealthIncrease(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})

	entities := []*Entity{entity}

	// First update: initializes PrevHealth, no particles
	system.Update(entities, 0.016)

	comp, ok := entity.GetComponent("health_regen_pulse")
	if !ok {
		t.Fatal("expected health_regen_pulse component to be added")
	}
	hrp := comp.(*HealthRegenPulseComponent)
	if !hrp.Initialized {
		t.Error("expected component to be initialized after first update")
	}
	if hrp.PrevHealth != 80 {
		t.Errorf("expected PrevHealth 80, got %f", hrp.PrevHealth)
	}

	// Simulate healing
	health := entity.GetHealth()
	health.Current = 90

	system.Update(entities, 0.016)
	if hrp.Accumulator != 10 {
		t.Errorf("expected accumulator 10, got %f", hrp.Accumulator)
	}
}

func TestHealthRegenPulseSystem_NoEffectWithoutHealthChange(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// Initialize
	system.Update(entities, 0.016)

	comp, _ := entity.GetComponent("health_regen_pulse")
	hrp := comp.(*HealthRegenPulseComponent)

	// No health change
	system.Update(entities, 0.016)
	if hrp.Accumulator != 0 {
		t.Errorf("expected accumulator 0 with no health change, got %f", hrp.Accumulator)
	}
}

func TestHealthRegenPulseSystem_IgnoresHealthDecrease(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	system.SetGenre("scifi")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Take damage
	entity.GetHealth().Current = 50
	system.Update(entities, 0.016)

	comp, _ := entity.GetComponent("health_regen_pulse")
	hrp := comp.(*HealthRegenPulseComponent)
	if hrp.Accumulator != 0 {
		t.Errorf("expected accumulator 0 for damage, got %f", hrp.Accumulator)
	}
}

func TestHealthRegenPulseSystem_SkipsEntitiesWithoutHealth(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	// No health component

	entities := []*Entity{entity}
	// Should not panic
	system.Update(entities, 0.016)

	_, ok := entity.GetComponent("health_regen_pulse")
	if ok {
		t.Error("should not add component to entity without health")
	}
}

func TestHealthRegenPulseSystem_SkipsEntitiesWithoutPosition(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100})
	// No position component

	entities := []*Entity{entity}
	// Should not panic
	system.Update(entities, 0.016)
}

func TestHealthRegenPulseSystem_PulseCooldown(t *testing.T) {
	world := NewWorld()
	system := NewHealthRegenPulseSystem(world, 12345)
	system.SetGenre("fantasy")
	// No particle system set - won't actually spawn, but accumulator logic runs

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})

	entities := []*Entity{entity}
	system.Update(entities, 0.016) // Initialize

	// Heal enough to trigger threshold
	entity.GetHealth().Current = 60
	system.Update(entities, 0.016)

	comp, _ := entity.GetComponent("health_regen_pulse")
	hrp := comp.(*HealthRegenPulseComponent)

	// healRatio = 10/100 = 0.1, which exceeds 0.02 threshold
	// But no particle system, so SpawnParticles won't run
	// The accumulator/timer logic should still work
	if hrp.PulseTimer < 0 {
		t.Error("expected pulse timer to be non-negative after heal detected")
	}
}

func TestHealthRegenPulseSystem_SetParticleSystem(t *testing.T) {
	system := NewHealthRegenPulseSystem(NewWorld(), 42)
	ps := &ParticleSystem{}
	system.SetParticleSystem(ps)
	if system.particleSystem != ps {
		t.Error("expected particle system to be set")
	}
}

func TestHealthRegenPulseSystem_PresetValues(t *testing.T) {
	system := NewHealthRegenPulseSystem(NewWorld(), 42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			system.SetGenre(genre)
			p := system.preset
			if p.baseCount <= 0 {
				t.Error("expected positive base count")
			}
			if p.duration <= 0 {
				t.Error("expected positive duration")
			}
			if p.minSize <= 0 || p.maxSize < p.minSize {
				t.Error("invalid size range")
			}
			if p.gravity >= 0 {
				t.Error("expected negative gravity for upward drift")
			}
		})
	}
}
