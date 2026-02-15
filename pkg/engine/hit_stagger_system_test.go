package engine

import (
	"math"
	"testing"
)

func getHitStagger(entity *Entity) *HitStaggerComponent {
	comp, ok := entity.GetComponent("hit_stagger")
	if !ok {
		return nil
	}
	return comp.(*HitStaggerComponent)
}

func TestCombatHitStaggerSystem_Creation(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.rng == nil {
		t.Error("expected non-nil RNG")
	}
}

func TestCombatHitStaggerSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name       string
		genreID    string
		wantMaxOff float64
	}{
		{"fantasy", "fantasy", 4.0},
		{"horror", "horror", 5.0},
		{"cyberpunk", "cyberpunk", 4.0},
		{"scifi", "scifi", 3.5},
		{"postapoc", "postapoc", 5.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCombatHitStaggerSystem(nil, 42)
			sys.SetGenre(tt.genreID)
			if sys.preset.MaxOffset != tt.wantMaxOff {
				t.Errorf("genre %q: expected MaxOffset=%.1f, got %.1f", tt.genreID, tt.wantMaxOff, sys.preset.MaxOffset)
			}
		})
	}
}

func TestCombatHitStaggerSystem_TriggerOnDamage(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 99)

	entity := NewEntity(1)
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}

	// First update: record baseline health
	sys.Update(entities, 0.016)
	stagger := getHitStagger(entity)
	if stagger.Active {
		t.Error("should not stagger on first frame")
	}

	// Take damage
	health.Current = 70
	sys.Update(entities, 0.016)

	if !stagger.Active {
		t.Error("expected stagger after taking damage")
	}
	magnitude := math.Sqrt(stagger.OffsetX*stagger.OffsetX + stagger.OffsetY*stagger.OffsetY)
	if magnitude < 0.5 {
		t.Errorf("expected meaningful offset, got magnitude %.3f", magnitude)
	}
}

func TestCombatHitStaggerSystem_NoStaggerOnHeal(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 99)

	entity := NewEntity(2)
	health := &HealthComponent{Current: 50, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}

	sys.Update(entities, 0.016)

	// Heal
	health.Current = 80
	sys.Update(entities, 0.016)

	stagger := getHitStagger(entity)
	if stagger.Active {
		t.Error("should not stagger on healing")
	}
}

func TestCombatHitStaggerSystem_DecayToZero(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 42)

	entity := NewEntity(3)
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Trigger damage
	health.Current = 50
	sys.Update(entities, 0.016)

	stagger := getHitStagger(entity)
	if !stagger.Active {
		t.Fatal("expected active stagger")
	}

	// Run enough frames to fully decay (fantasy duration ~0.27s max)
	for i := 0; i < 30; i++ {
		sys.Update(entities, 0.016)
	}

	if stagger.Active {
		t.Errorf("expected stagger to finish after %.2fs, timer=%.4f", 0.48, stagger.Timer)
	}
	if stagger.OffsetX != 0 || stagger.OffsetY != 0 {
		t.Errorf("expected zero offset after decay, got (%.3f, %.3f)", stagger.OffsetX, stagger.OffsetY)
	}
}

func TestCombatHitStaggerSystem_IntensityScaling(t *testing.T) {
	tests := []struct {
		name     string
		damage   float64
		wantMin  float64
		wantMax  float64
	}{
		{"small_hit", 5, 0.5, 2.5},
		{"medium_hit", 30, 1.5, 4.0},
		{"big_hit", 80, 3.0, 5.0},
		{"full_kill", 100, 3.5, 5.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCombatHitStaggerSystem(nil, 12345)

			entity := NewEntity(10)
			health := &HealthComponent{Current: 100, Max: 100}
			entity.AddComponent(health)

			entities := []*Entity{entity}
			sys.Update(entities, 0.016)

			health.Current = 100 - tt.damage
			sys.Update(entities, 0.016)

			stagger := getHitStagger(entity)
			magnitude := math.Sqrt(stagger.OffsetX*stagger.OffsetX + stagger.OffsetY*stagger.OffsetY)
			if magnitude < tt.wantMin || magnitude > tt.wantMax {
				t.Errorf("expected magnitude in [%.1f, %.1f], got %.3f", tt.wantMin, tt.wantMax, magnitude)
			}
		})
	}
}

func TestCombatHitStaggerSystem_BelowMinDamageNoStagger(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 42)

	entity := NewEntity(4)
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Tiny damage below 3% threshold (fantasy preset)
	health.Current = 98
	sys.Update(entities, 0.016)

	stagger := getHitStagger(entity)
	if stagger.Active {
		t.Error("should not stagger for damage below minimum threshold")
	}
}

func TestCombatHitStaggerSystem_NoHealthNoStagger(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 42)

	entity := NewEntity(5)
	// No HealthComponent
	entities := []*Entity{entity}

	sys.Update(entities, 0.016)

	_, hasComp := entity.GetComponent("hit_stagger")
	if hasComp {
		t.Error("should not create stagger component without health")
	}
}

func TestCombatHitStaggerSystem_CleanupStaleEntries(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 42)
	sys.cleanupInterval = 0.0 // Force immediate cleanup

	entity := NewEntity(6)
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	if _, ok := sys.prevHealth[6]; !ok {
		t.Error("expected entity tracked in prevHealth")
	}

	// Remove entity from list and trigger cleanup
	sys.Update([]*Entity{}, 0.016)

	if _, ok := sys.prevHealth[6]; ok {
		t.Error("expected stale entry to be cleaned up")
	}
}

func TestCombatHitStaggerSystem_RepeatedHitsResetStagger(t *testing.T) {
	sys := NewCombatHitStaggerSystem(nil, 42)

	entity := NewEntity(7)
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// First hit
	health.Current = 80
	sys.Update(entities, 0.016)
	stagger := getHitStagger(entity)
	if !stagger.Active {
		t.Fatal("expected active stagger after first hit")
	}

	// Partially decay
	sys.Update(entities, 0.05)

	// Second hit while still staggering
	health.Current = 60
	sys.Update(entities, 0.016)
	if !stagger.Active {
		t.Error("expected stagger still active after second hit")
	}
	// Timer should be reset to full duration
	if stagger.Timer < stagger.Duration*0.9 {
		t.Errorf("expected timer near full duration after re-trigger, got %.3f/%.3f", stagger.Timer, stagger.Duration)
	}
}

func TestHitStaggerComponent_Type(t *testing.T) {
	comp := &HitStaggerComponent{}
	if comp.Type() != "hit_stagger" {
		t.Errorf("expected type 'hit_stagger', got %q", comp.Type())
	}
}

func BenchmarkCombatHitStaggerSystem_Update(b *testing.B) {
	sys := NewCombatHitStaggerSystem(nil, 42)

	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&HealthComponent{Current: 100, Max: 100})
		e.AddComponent(&HitStaggerComponent{Active: true, Timer: 0.1, Duration: 0.2, OffsetX: 2, OffsetY: 1, InitialOffsetX: 3, InitialOffsetY: 2})
		entities[i] = e
	}

	// Warm up health tracking
	sys.Update(entities, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
