package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewHazardProximityWarningParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 42)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 42 {
		t.Errorf("expected seed 42, got %d", sys.seed)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre fantasy, got %s", sys.genreID)
	}
	if sys.warningMargin != 48.0 {
		t.Errorf("expected warning margin 48, got %f", sys.warningMargin)
	}
}

func TestHazardProximityWarningParticleSystem_SetGenre(t *testing.T) {
	sys := NewHazardProximityWarningParticleSystem(nil, 1)
	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("expected genre %s, got %s", g, sys.genreID)
		}
	}
}

func TestHazardProximityWarningParticleSystem_SetDependencies(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 1)

	hs := NewHazardSystem()
	sys.SetHazardSystem(hs)
	if sys.hazardSystem != hs {
		t.Error("hazard system not set")
	}

	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestHazardProximityWarningParticleSystem_UpdateNilDeps(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 1)

	// Should not panic with nil deps
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.016)
}

func TestHazardProximityWarningParticleSystem_UpdateThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 1)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)

	// First call with small delta should accumulate but not emit
	sys.Update([]*Entity{}, 0.01)
	if sys.timeSinceEmit != 0.01 {
		t.Errorf("expected accumulator 0.01, got %f", sys.timeSinceEmit)
	}

	// Not enough time passed — should not reset
	sys.Update([]*Entity{}, 0.01)
	if sys.timeSinceEmit < 0.01 {
		t.Error("accumulator should still be accumulating")
	}
}

func TestHazardProximityWarningParticleSystem_UpdateWithNearbyHazard(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 99)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Create a hazard zone at (100, 100) with radius 40
	hs.CreateHazard(HazardPoison, 100, 100, 10.0, 40.0)
	// Update hazard system to register the zone
	hs.Update(world.GetEntities(), 0.016)

	// Create an entity near the hazard edge (at distance ~50 from center, just outside radius 40)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 150, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// First update: accumulate time past emit interval
	sys.Update(entities, 0.4)

	// Entity at distance 50 from hazard center, hazard radius 40.
	// Warning margin is 48, so entity is within warningMargin+radius (88).
	// Entity should have been warned
	if _, ok := sys.lastWarned[entity.ID]; !ok {
		t.Error("expected entity to be warned about nearby hazard")
	}
}

func TestHazardProximityWarningParticleSystem_NoWarningWhenInside(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 99)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)

	// Create a hazard zone at (100, 100) with radius 60
	hs.CreateHazard(HazardPoison, 100, 100, 10.0, 60.0)
	hs.Update(world.GetEntities(), 0.016)

	// Entity deep inside hazard (at center)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.4)

	// Entity is at dist 0 which is <= radius*0.8 (48), so no warning
	if _, ok := sys.lastWarned[entity.ID]; ok {
		t.Error("entity inside hazard should not get proximity warning")
	}
}

func TestHazardProximityWarningParticleSystem_NoWarningWhenFar(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 99)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)

	// Hazard at (100,100) radius 40
	hs.CreateHazard(HazardOil, 100, 100, 10.0, 40.0)
	hs.Update(world.GetEntities(), 0.016)

	// Entity far away at (300, 300) — distance ~283, well beyond warning margin
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 300, Y: 300})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 0.4)

	if _, ok := sys.lastWarned[entity.ID]; ok {
		t.Error("entity far from hazard should not be warned")
	}
}

func TestHazardProximityWarningParticleSystem_GenrePresets(t *testing.T) {
	tests := []struct {
		genre        string
		expectedType particles.ParticleType
	}{
		{"fantasy", particles.ParticleMagic},
		{"horror", particles.ParticleSmoke},
		{"scifi", particles.ParticleSpark},
		{"cyberpunk", particles.ParticleSpark},
		{"postapoc", particles.ParticleDebris},
		{"unknown", particles.ParticleMagic},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewHazardProximityWarningParticleSystem(nil, 1)
			sys.SetGenre(tt.genre)
			preset := sys.getGenrePreset()
			if preset.ParticleType != tt.expectedType {
				t.Errorf("genre %s: expected particle type %v, got %v",
					tt.genre, tt.expectedType, preset.ParticleType)
			}
			if preset.Duration <= 0 {
				t.Errorf("genre %s: duration should be positive", tt.genre)
			}
			if preset.MinSize >= preset.MaxSize {
				t.Errorf("genre %s: min size should be less than max size", tt.genre)
			}
		})
	}
}

func TestHazardProximityWarningParticleSystem_HazardParticleType(t *testing.T) {
	sys := NewHazardProximityWarningParticleSystem(nil, 1)
	tests := []struct {
		hazardType   HazardType
		defaultType  particles.ParticleType
		expectedType particles.ParticleType
	}{
		{HazardPoison, particles.ParticleMagic, particles.ParticleSmoke},
		{HazardOil, particles.ParticleMagic, particles.ParticleDebris},
		{HazardSmoke, particles.ParticleMagic, particles.ParticleSmoke},
		{HazardWater, particles.ParticleMagic, particles.ParticleMagic},
	}

	for _, tt := range tests {
		t.Run(tt.hazardType.String(), func(t *testing.T) {
			got := sys.hazardParticleType(tt.hazardType, tt.defaultType)
			if got != tt.expectedType {
				t.Errorf("hazard %s: expected %v, got %v", tt.hazardType.String(), tt.expectedType, got)
			}
		})
	}
}

func TestHazardProximityWarningParticleSystem_HazardCountBonus(t *testing.T) {
	sys := NewHazardProximityWarningParticleSystem(nil, 1)
	tests := []struct {
		hazardType HazardType
		expected   int
	}{
		{HazardPoison, 2},
		{HazardOil, 1},
		{HazardWater, 0},
		{HazardSmoke, 0},
	}

	for _, tt := range tests {
		t.Run(tt.hazardType.String(), func(t *testing.T) {
			got := sys.hazardCountBonus(tt.hazardType)
			if got != tt.expected {
				t.Errorf("hazard %s: expected bonus %d, got %d", tt.hazardType.String(), tt.expected, got)
			}
		})
	}
}

func TestHazardProximityWarningParticleSystem_CooldownPreventsSpam(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 99)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)

	hs.CreateHazard(HazardPoison, 100, 100, 10.0, 40.0)
	hs.Update(world.GetEntities(), 0.016)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 145, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}

	// First warning
	sys.Update(entities, 0.4)
	if _, ok := sys.lastWarned[entity.ID]; !ok {
		t.Fatal("expected entity to be warned")
	}

	// Second update immediately — should be on cooldown, cooldown value should decrease
	prevCooldown := sys.lastWarned[entity.ID]
	sys.Update(entities, 0.4)
	if cd, ok := sys.lastWarned[entity.ID]; ok && cd >= prevCooldown {
		t.Error("cooldown should have decreased")
	}
}

func TestHazardProximityWarningParticleSystem_SkipsEntitiesWithoutHealth(t *testing.T) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 99)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)

	hs.CreateHazard(HazardPoison, 100, 100, 10.0, 40.0)
	hs.Update(world.GetEntities(), 0.016)

	// Entity without health component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 145, Y: 100})

	sys.Update([]*Entity{entity}, 0.4)
	if _, ok := sys.lastWarned[entity.ID]; ok {
		t.Error("entity without health should not be warned")
	}
}

func BenchmarkHazardProximityWarningParticleSystem(b *testing.B) {
	world := NewWorld()
	sys := NewHazardProximityWarningParticleSystem(world, 42)
	hs := NewHazardSystemWithLogger(nil)
	hs.SetWorld(world)
	ps := NewParticleSystem()
	sys.SetHazardSystem(hs)
	sys.SetParticleSystem(ps)

	// Create several hazards
	for i := 0; i < 10; i++ {
		hs.CreateHazard(HazardPoison, float64(i*80+50), float64(i*60+50), 30.0, 40.0)
	}
	hs.Update(world.GetEntities(), 0.016)

	// Create entities
	entities := make([]*Entity, 100)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 8)})
		e.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceEmit = 0
		sys.lastWarned = make(map[uint64]float64, 32)
		sys.Update(entities, 0.4)
	}
}
