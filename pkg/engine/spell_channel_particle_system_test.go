package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewSpellChannelParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewSpellChannelParticleSystem returned nil")
	}

	if sys.world != world {
		t.Error("world not set correctly")
	}

	if sys.rng == nil {
		t.Error("rng not initialized")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}

	if sys.spawnInterval <= 0 {
		t.Error("spawn interval should be positive")
	}

	if sys.particleCount <= 0 {
		t.Error("particle count should be positive")
	}
}

func TestSpellChannelParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestSpellChannelParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSys != ps {
		t.Error("particle system not set correctly")
	}
}

func TestSpellChannelParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	// No particle system set

	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{Casting: 0})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	// Should not panic
	sys.Update(entities, 0.016)
}

func TestSpellChannelParticleSystem_Update_NoWorld(t *testing.T) {
	sys := NewSpellChannelParticleSystem(nil, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	// Should not panic with nil world
	sys.Update([]*Entity{}, 0.016)
}

func TestSpellChannelParticleSystem_Update_NoSpellSlots(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No spell_slots component

	entities := []*Entity{entity}

	// Should not panic
	sys.Update(entities, 0.016)
}

func TestSpellChannelParticleSystem_Update_NotCasting(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	sys.SetParticleSystem(NewParticleSystem())

	entity := world.CreateEntity()
	entity.AddComponent(&SpellSlotComponent{Casting: -1}) // Not casting
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	sys.Update(entities, 0.016)

	// Should not track entity since not casting
	if len(sys.baseParticleMap) > 0 {
		t.Error("should not track non-casting entity")
	}
}

func TestSpellChannelParticleSystem_Update_Casting(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	slots := &SpellSlotComponent{
		Casting:    0,
		CastingBar: 0.5,
	}
	slots.Slots[0] = &magic.Spell{
		Name:    "Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
	}
	entity.AddComponent(slots)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	sys.Update(entities, 0.016)

	// Should track entity
	if !sys.baseParticleMap[entity.ID] {
		t.Error("should track casting entity")
	}
}

func TestSpellChannelParticleSystem_Update_CastingEnds(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	slots := &SpellSlotComponent{
		Casting:    0,
		CastingBar: 0.5,
	}
	slots.Slots[0] = &magic.Spell{
		Name:    "Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
	}
	entity.AddComponent(slots)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	// Start casting
	sys.Update(entities, 0.016)

	if !sys.baseParticleMap[entity.ID] {
		t.Fatal("should track casting entity")
	}

	// Stop casting
	slots.Casting = -1
	sys.Update(entities, 0.016)

	// Should no longer track
	if sys.baseParticleMap[entity.ID] {
		t.Error("should stop tracking when casting ends")
	}
}

func TestSpellChannelParticleSystem_Update_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	slots := &SpellSlotComponent{
		Casting:    0,
		CastingBar: 0.5,
	}
	slots.Slots[0] = &magic.Spell{
		Name:    "Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
	}
	entity.AddComponent(slots)
	// No position component

	entities := []*Entity{entity}

	// Should not panic
	sys.Update(entities, 0.016)
}

func TestSpellChannelParticleSystem_getChannelParticleType(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)

	tests := []struct {
		element  magic.ElementType
		expected particles.ParticleType
	}{
		{magic.ElementFire, particles.ParticleEmber},
		{magic.ElementIce, particles.ParticleSparkle},
		{magic.ElementLightning, particles.ParticleSpark},
		{magic.ElementEarth, particles.ParticleDust},
		{magic.ElementWind, particles.ParticleDust},
		{magic.ElementLight, particles.ParticleSparkle},
		{magic.ElementDark, particles.ParticleSmoke},
		{magic.ElementArcane, particles.ParticleMagic},
		{magic.ElementNone, particles.ParticleMagic},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			ptype := sys.getChannelParticleType(tt.element)
			if ptype != tt.expected {
				t.Errorf("getChannelParticleType(%v) = %v, want %v", tt.element, ptype, tt.expected)
			}
		})
	}
}

func TestSpellChannelParticleSystem_SpawnCooldown(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	slots := &SpellSlotComponent{
		Casting:    0,
		CastingBar: 0.5,
	}
	slots.Slots[0] = &magic.Spell{
		Name:    "Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
	}
	entity.AddComponent(slots)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	// First update should spawn and set cooldown
	sys.Update(entities, 0.016)

	cooldown := sys.spawnCooldowns[entity.ID]
	if cooldown <= 0 {
		t.Error("cooldown should be set after spawning")
	}

	// Second immediate update should not spawn (cooldown active)
	initialCooldown := cooldown
	sys.Update(entities, 0.016)

	// Cooldown should have decreased
	if sys.spawnCooldowns[entity.ID] >= initialCooldown {
		t.Error("cooldown should decrease over time")
	}
}

func TestSpellChannelParticleSystem_CleanupStaleEntries(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	slots := &SpellSlotComponent{
		Casting:    0,
		CastingBar: 0.5,
	}
	slots.Slots[0] = &magic.Spell{
		Name:    "Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
	}
	entity.AddComponent(slots)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	// Start tracking
	sys.Update(entities, 0.016)

	if !sys.baseParticleMap[entity.ID] {
		t.Fatal("entity should be tracked")
	}

	// Update with empty entity list (entity removed from world)
	sys.Update([]*Entity{}, 0.016)

	// Should clean up stale entry
	if sys.baseParticleMap[entity.ID] {
		t.Error("should clean up stale entry when entity removed")
	}
}

func TestSpellChannelParticleSystem_NilSpellInSlot(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	slots := &SpellSlotComponent{
		Casting:    0, // Casting but...
		CastingBar: 0.5,
	}
	// No spell in slot 0 (nil)
	entity.AddComponent(slots)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	// Should not panic with nil spell
	sys.Update(entities, 0.016)

	// Should not track entity with nil spell
	if sys.baseParticleMap[entity.ID] {
		t.Error("should not track entity with nil spell in slot")
	}
}

func TestSpellChannelParticleSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create two casting entities
	entity1 := world.CreateEntity()
	slots1 := &SpellSlotComponent{Casting: 0, CastingBar: 0.3}
	slots1.Slots[0] = &magic.Spell{Name: "Fireball", Element: magic.ElementFire}
	entity1.AddComponent(slots1)
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})

	entity2 := world.CreateEntity()
	slots2 := &SpellSlotComponent{Casting: 1, CastingBar: 0.7}
	slots2.Slots[1] = &magic.Spell{Name: "Ice Blast", Element: magic.ElementIce}
	entity2.AddComponent(slots2)
	entity2.AddComponent(&PositionComponent{X: 200, Y: 200})

	// One non-casting entity
	entity3 := world.CreateEntity()
	slots3 := &SpellSlotComponent{Casting: -1}
	entity3.AddComponent(slots3)
	entity3.AddComponent(&PositionComponent{X: 300, Y: 300})

	entities := []*Entity{entity1, entity2, entity3}

	sys.Update(entities, 0.016)

	// Should track both casting entities
	if !sys.baseParticleMap[entity1.ID] {
		t.Error("should track entity1")
	}
	if !sys.baseParticleMap[entity2.ID] {
		t.Error("should track entity2")
	}
	if sys.baseParticleMap[entity3.ID] {
		t.Error("should not track non-casting entity3")
	}
}

func BenchmarkSpellChannelParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSpellChannelParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create 100 entities, 50 casting
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		casting := -1
		if i < 50 {
			casting = 0
		}
		slots := &SpellSlotComponent{Casting: casting, CastingBar: 0.5}
		if casting >= 0 {
			slots.Slots[0] = &magic.Spell{Name: "Test", Element: magic.ElementFire}
		}
		entity.AddComponent(slots)
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
