// Package engine provides tests for EquipmentDurabilityParticleSystem.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestNewEquipmentDurabilityParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewEquipmentDurabilityParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.lastStateCache == nil {
		t.Error("lastStateCache not initialized")
	}
	if len(sys.particleCounts) != 3 {
		t.Errorf("particleCounts has %d entries, want 3", len(sys.particleCounts))
	}
}

func TestEquipmentDurabilityParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestEquipmentDurabilityParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "horror")
	}
}

func TestEquipmentDurabilityParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	// Don't set particle system

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entities := []*Entity{entity}

	// Should not panic
	sys.Update(entities, 0.016)
}

func TestEquipmentDurabilityParticleSystem_Update_NoDegradation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	equipComp := NewEquipmentComponent()
	equipComp.Slots[SlotMainHand] = &item.Item{
		ID:    "sword1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// First update should initialize cache but not spawn particles
	sys.Update(entities, 0.016)

	// Second update with same durability should not spawn particles
	sys.Update(entities, 0.016)
}

func TestEquipmentDurabilityParticleSystem_Update_Degradation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	equipComp := NewEquipmentComponent()
	equipComp.Slots[SlotMainHand] = &item.Item{
		ID:    "sword1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// First update initializes cache at Pristine
	sys.Update(entities, 0.016)

	// Degrade to Worn state
	equipComp.Slots[SlotMainHand].Stats.Durability = 60

	// Second update should detect transition and spawn particles
	sys.Update(entities, 0.016)

	// Verify cache was updated
	cachedState := sys.lastStateCache[entity.ID][SlotMainHand]
	if cachedState != sprites.DamageStateWorn {
		t.Errorf("cached state = %v, want %v", cachedState, sprites.DamageStateWorn)
	}
}

func TestEquipmentDurabilityParticleSystem_Update_MultipleDegradations(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("cyberpunk")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	equipComp := NewEquipmentComponent()
	equipComp.Slots[SlotChest] = &item.Item{
		ID:    "armor1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// Initialize at Pristine
	sys.Update(entities, 0.016)

	// Degrade to Worn
	equipComp.Slots[SlotChest].Stats.Durability = 60
	sys.Update(entities, 0.016)

	// Degrade to Damaged
	equipComp.Slots[SlotChest].Stats.Durability = 30
	sys.Update(entities, 0.016)

	cachedState := sys.lastStateCache[entity.ID][SlotChest]
	if cachedState != sprites.DamageStateDamaged {
		t.Errorf("cached state = %v, want %v", cachedState, sprites.DamageStateDamaged)
	}

	// Degrade to Broken
	equipComp.Slots[SlotChest].Stats.Durability = 10
	sys.Update(entities, 0.016)

	cachedState = sys.lastStateCache[entity.ID][SlotChest]
	if cachedState != sprites.DamageStateBroken {
		t.Errorf("cached state = %v, want %v", cachedState, sprites.DamageStateBroken)
	}
}

func TestEquipmentDurabilityParticleSystem_OnEquipmentDegraded(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(entity)

	// Test callback for Pristine -> Worn transition
	sys.OnEquipmentDegraded(entity, SlotMainHand, 100, 60, 100)

	// Test callback for Worn -> Damaged transition
	sys.OnEquipmentDegraded(entity, SlotMainHand, 60, 30, 100)

	// Test callback for Damaged -> Broken transition
	sys.OnEquipmentDegraded(entity, SlotMainHand, 30, 10, 100)
}

func TestEquipmentDurabilityParticleSystem_OnEquipmentDegraded_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	// Don't set particle system

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.OnEquipmentDegraded(entity, SlotMainHand, 100, 60, 100)
}

func TestEquipmentDurabilityParticleSystem_OnEquipmentDegraded_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	// No position component

	// Should not panic
	sys.OnEquipmentDegraded(entity, SlotMainHand, 100, 60, 100)
}

func TestEquipmentDurabilityParticleSystem_OnEquipmentDegraded_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Should not panic
	sys.OnEquipmentDegraded(nil, SlotMainHand, 100, 60, 100)
}

func TestEquipmentDurabilityParticleSystem_ClearEntityCache(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	equipComp := NewEquipmentComponent()
	equipComp.Slots[SlotMainHand] = &item.Item{
		ID:    "sword1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// Initialize cache
	sys.Update(entities, 0.016)

	if _, ok := sys.lastStateCache[entity.ID]; !ok {
		t.Fatal("cache not initialized")
	}

	// Clear cache
	sys.ClearEntityCache(entity.ID)

	if _, ok := sys.lastStateCache[entity.ID]; ok {
		t.Error("cache not cleared")
	}
}

func TestEquipmentDurabilityParticleSystem_GetParticleType(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)

	tests := []struct {
		state sprites.DamageState
	}{
		{sprites.DamageStateWorn},
		{sprites.DamageStateDamaged},
		{sprites.DamageStateBroken},
		{sprites.DamageStatePristine},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			// Should not panic and return a valid type
			_ = sys.getParticleType(tt.state)
		})
	}
}

func TestEquipmentDurabilityParticleSystem_GetDuration(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)

	tests := []struct {
		state  sprites.DamageState
		minDur float64
		maxDur float64
	}{
		{sprites.DamageStateWorn, 0.3, 0.5},
		{sprites.DamageStateDamaged, 0.5, 0.7},
		{sprites.DamageStateBroken, 0.8, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			dur := sys.getDuration(tt.state)
			if dur < tt.minDur || dur > tt.maxDur {
				t.Errorf("duration = %f, want between %f and %f", dur, tt.minDur, tt.maxDur)
			}
		})
	}
}

func TestEquipmentDurabilityParticleSystem_GetGravity(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)

	// Worn should float upward (negative gravity)
	wornGrav := sys.getGravity(sprites.DamageStateWorn)
	if wornGrav >= 0 {
		t.Errorf("worn gravity = %f, want negative", wornGrav)
	}

	// Damaged and Broken should fall (positive gravity)
	damagedGrav := sys.getGravity(sprites.DamageStateDamaged)
	if damagedGrav <= 0 {
		t.Errorf("damaged gravity = %f, want positive", damagedGrav)
	}

	brokenGrav := sys.getGravity(sprites.DamageStateBroken)
	if brokenGrav <= damagedGrav {
		t.Errorf("broken gravity = %f should be greater than damaged %f", brokenGrav, damagedGrav)
	}
}

func TestEquipmentDurabilityParticleSystem_GetSizes(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)

	states := []sprites.DamageState{
		sprites.DamageStateWorn,
		sprites.DamageStateDamaged,
		sprites.DamageStateBroken,
	}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			minSize := sys.getMinSize(state)
			maxSize := sys.getMaxSize(state)

			if minSize <= 0 {
				t.Errorf("minSize = %f, want positive", minSize)
			}
			if maxSize <= minSize {
				t.Errorf("maxSize = %f should be greater than minSize %f", maxSize, minSize)
			}
		})
	}
}

func TestEquipmentDurabilityParticleSystem_GenreVariation(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentDurabilityParticleSystem(world, 12345)
			ps := NewParticleSystem()
			sys.SetParticleSystem(ps)
			sys.SetGenre(genre)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			world.AddEntity(entity)

			// Should spawn particles without panic for any genre
			sys.OnEquipmentDegraded(entity, SlotMainHand, 100, 60, 100)
		})
	}
}

func TestEquipmentDurabilityParticleSystem_MultipleSlots(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	equipComp := NewEquipmentComponent()
	equipComp.Slots[SlotMainHand] = &item.Item{
		ID:    "sword1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	equipComp.Slots[SlotChest] = &item.Item{
		ID:    "armor1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	equipComp.Slots[SlotBoots] = &item.Item{
		ID:    "boots1",
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// Initialize cache
	sys.Update(entities, 0.016)

	// Degrade different slots to different states
	equipComp.Slots[SlotMainHand].Stats.Durability = 60 // Worn
	equipComp.Slots[SlotChest].Stats.Durability = 30    // Damaged
	equipComp.Slots[SlotBoots].Stats.Durability = 10    // Broken

	sys.Update(entities, 0.016)

	// Verify each slot is cached correctly
	if sys.lastStateCache[entity.ID][SlotMainHand] != sprites.DamageStateWorn {
		t.Error("main hand state incorrect")
	}
	if sys.lastStateCache[entity.ID][SlotChest] != sprites.DamageStateDamaged {
		t.Error("chest state incorrect")
	}
	if sys.lastStateCache[entity.ID][SlotBoots] != sprites.DamageStateBroken {
		t.Error("boots state incorrect")
	}
}

func BenchmarkEquipmentDurabilityParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewEquipmentDurabilityParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Create 100 entities with equipment
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		equipComp := NewEquipmentComponent()
		equipComp.Slots[SlotMainHand] = &item.Item{
			ID:    "sword",
			Stats: item.Stats{Durability: 100, DurabilityMax: 100},
		}
		entity.AddComponent(equipComp)
		entities[i] = entity
	}

	// Initialize cache
	sys.Update(entities, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
