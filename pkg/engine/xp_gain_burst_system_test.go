package engine

import (
	"testing"
)

func TestXPGainBurstComponentType(t *testing.T) {
	comp := &XPGainBurstComponent{}
	if got := comp.Type(); got != "xp_gain_burst" {
		t.Errorf("Type() = %q, want %q", got, "xp_gain_burst")
	}
}

func TestNewXPGainBurstSystem(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewXPGainBurstSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.seed != 12345 {
		t.Error("seed not set")
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
}

func TestXPGainBurstSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		wantPT  string
	}{
		{"fantasy", "fantasy", "sparkle"},
		{"scifi", "scifi", "magic"},
		{"horror", "horror", "ember"},
		{"cyberpunk", "cyberpunk", "spark"},
		{"postapoc", "postapoc", "dust"},
		{"default", "unknown", "sparkle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewXPGainBurstSystem(world, 42)
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.genre)
			}
			got := sys.preset.particleType.String()
			if got != tt.wantPT {
				t.Errorf("particle type = %q, want %q", got, tt.wantPT)
			}
		})
	}
}

func TestXPGainBurstSystem_UpdateNoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)
	// Should not panic with nil particle system
	entity := NewEntity(1)
	sys.Update([]*Entity{entity}, 0.016)
}

func TestXPGainBurstSystem_UpdateSkipsWithoutExperience(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No experience component - should skip
	sys.Update([]*Entity{entity}, 0.016)

	// Verify no burst component was added
	if _, ok := entity.GetComponent("xp_gain_burst"); ok {
		t.Error("burst component should not be added without experience")
	}
}

func TestXPGainBurstSystem_InitializesPrevXP(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	exp := NewExperienceComponent()
	exp.TotalXP = 50
	entity.AddComponent(exp)

	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("xp_gain_burst")
	if !ok {
		t.Fatal("burst component not created")
	}
	burst := comp.(*XPGainBurstComponent)
	if !burst.Initialized {
		t.Error("component should be initialized after first update")
	}
	if burst.PrevTotalXP != 50 {
		t.Errorf("PrevTotalXP = %d, want 50", burst.PrevTotalXP)
	}
}

func TestXPGainBurstSystem_DetectsXPGain(t *testing.T) {
	tests := []struct {
		name       string
		startXP    int
		gainXP     int
		requiredXP int
		genre      string
	}{
		{"small_fantasy", 100, 10, 200, "fantasy"},
		{"medium_scifi", 500, 100, 400, "scifi"},
		{"large_horror", 200, 300, 500, "horror"},
		{"cyberpunk_gain", 0, 50, 100, "cyberpunk"},
		{"postapoc_gain", 1000, 25, 300, "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewXPGainBurstSystem(world, 42)
			sys.SetGenre(tt.genre)
			ps := NewParticleSystem()
			sys.SetParticleSystem(ps)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			exp := NewExperienceComponent()
			exp.TotalXP = tt.startXP
			exp.RequiredXP = tt.requiredXP
			entity.AddComponent(exp)

			// First update: initialize
			sys.Update([]*Entity{entity}, 0.016)

			// Simulate XP gain
			exp.TotalXP += tt.gainXP

			// Second update: should detect gain
			sys.Update([]*Entity{entity}, 0.016)

			comp, _ := entity.GetComponent("xp_gain_burst")
			burst := comp.(*XPGainBurstComponent)
			if burst.PrevTotalXP != tt.startXP+tt.gainXP {
				t.Errorf("PrevTotalXP = %d, want %d", burst.PrevTotalXP, tt.startXP+tt.gainXP)
			}
			// Burst timer should have been set (cooldown active)
			if burst.BurstTimer <= 0 {
				t.Error("burst timer should be positive after XP gain")
			}
		})
	}
}

func TestXPGainBurstSystem_CooldownPreventsSpam(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	exp := NewExperienceComponent()
	exp.TotalXP = 100
	entity.AddComponent(exp)

	// Initialize
	sys.Update([]*Entity{entity}, 0.016)

	// First XP gain
	exp.TotalXP = 150
	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("xp_gain_burst")
	burst := comp.(*XPGainBurstComponent)
	firstTimer := burst.BurstTimer

	// Second XP gain during cooldown (small deltaTime)
	exp.TotalXP = 200
	sys.Update([]*Entity{entity}, 0.01)

	// Timer should have decreased but still be positive
	if burst.BurstTimer >= firstTimer {
		t.Error("timer should decrease over time")
	}
}

func TestXPGainBurstSystem_NoXPChange(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	exp := NewExperienceComponent()
	exp.TotalXP = 100
	entity.AddComponent(exp)

	// Initialize
	sys.Update([]*Entity{entity}, 0.016)

	// No XP change
	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("xp_gain_burst")
	burst := comp.(*XPGainBurstComponent)
	if burst.BurstTimer > 0 {
		t.Error("timer should not be set when no XP gained")
	}
}

func TestXPGainBurstSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)

	if sys.particleSystem != nil {
		t.Error("particle system should be nil initially")
	}

	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func BenchmarkXPGainBurstSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewXPGainBurstSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		exp := NewExperienceComponent()
		exp.TotalXP = i * 100
		e.AddComponent(exp)
		entities[i] = e
	}

	// Initialize all
	sys.Update(entities, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
