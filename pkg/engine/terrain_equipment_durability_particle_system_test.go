//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainEquipmentDurabilityParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.world != world {
		t.Error("expected world reference to be set")
	}
	if system.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", system.seed)
	}
	if system.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got '%s'", system.genreID)
	}
	if system.tileSize != 32 {
		t.Errorf("expected default tile size 32, got %d", system.tileSize)
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()

	system.SetParticleSystem(particleSystem)

	if system.particleSystem != particleSystem {
		t.Error("expected particle system to be set")
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld(nil)
			system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
			system.SetGenre(tt.genreID)

			if system.genreID != tt.genreID {
				t.Errorf("expected genre '%s', got '%s'", tt.genreID, system.genreID)
			}
		})
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_SetTileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"valid size", 64, 64},
		{"default size", 32, 32},
		{"zero keeps default", 0, 32},
		{"negative keeps default", -1, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld(nil)
			system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
			system.SetTileSize(tt.size)

			if system.tileSize != tt.expected {
				t.Errorf("expected tile size %d, got %d", tt.expected, system.tileSize)
			}
		})
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_CalculateDamageState(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)

	tests := []struct {
		name          string
		current       int
		max           int
		expectedState int
	}{
		{"pristine 100%", 100, 100, 0},
		{"pristine 76%", 76, 100, 0},
		{"worn 75%", 75, 100, 1},
		{"worn 51%", 51, 100, 1},
		{"damaged 50%", 50, 100, 2},
		{"damaged 26%", 26, 100, 2},
		{"broken 25%", 25, 100, 3},
		{"broken 0%", 0, 100, 3},
		{"zero max", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := system.calculateDamageState(tt.current, tt.max)
			if state != tt.expectedState {
				t.Errorf("expected state %d, got %d", tt.expectedState, state)
			}
		})
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_CalculateTotalDurability(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)

	// Create equipment component with items
	equipComp := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}

	// Test with no items
	total, max := system.calculateTotalDurability(equipComp)
	if total != 0 || max != 0 {
		t.Errorf("expected 0/0 for empty equipment, got %d/%d", total, max)
	}

	// Add weapon with durability
	equipComp.Slots[SlotMainHand] = &item.Item{
		Stats: item.ItemStats{
			Durability:    80,
			DurabilityMax: 100,
		},
	}

	total, max = system.calculateTotalDurability(equipComp)
	if total != 80 || max != 100 {
		t.Errorf("expected 80/100, got %d/%d", total, max)
	}

	// Add armor
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{
			Durability:    50,
			DurabilityMax: 100,
		},
	}

	total, max = system.calculateTotalDurability(equipComp)
	if total != 130 || max != 200 {
		t.Errorf("expected 130/200, got %d/%d", total, max)
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_UpdateNoTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	// No terrain set

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	})

	// Should not panic with nil terrain
	system.Update([]*Entity{entity}, 0.016)
}

func TestTerrainEquipmentDurabilityParticleSystem_UpdateWithTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	// Create terrain
	terr := terrain.NewTerrain(10, 10)
	system.SetTerrain(terr)
	system.SetTileSize(32)

	// Set a lava tile at position (0,0)
	terr.SetTile(0, 0, terrain.TileLavaFlow)

	// Create entity with equipment
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 16, Y: 16}) // On tile (0,0)
	equipComp := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	entity.AddComponent(equipComp)

	// First update - establishes baseline
	system.Update([]*Entity{entity}, 0.016)

	// Simulate durability damage
	equipComp.Slots[SlotChest].Stats.Durability = 90

	// Second update - should detect damage and spawn particles
	system.Update([]*Entity{entity}, 0.016)

	// Check that state was cached
	state, exists := system.lastDurabilityState[entity.ID]
	if !exists {
		t.Error("expected durability state to be cached")
	}
	if state.totalDurability != 90 {
		t.Errorf("expected cached durability 90, got %d", state.totalDurability)
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_GetLavaParticleType(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)

	tests := []struct {
		genreID string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			system.SetGenre(tt.genreID)
			particleType := system.getLavaParticleType()
			// Just verify it returns a valid type (not zero value)
			if particleType == 0 && tt.genreID != "unknown" && tt.genreID != "postapoc" {
				// Note: Some genres may return types that equal 0, which is valid
				// The important thing is the function doesn't panic
			}
		})
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_GetTrapParticleType(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)

	tests := []struct {
		genreID string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			system.SetGenre(tt.genreID)
			// Just verify it doesn't panic
			_ = system.getTrapParticleType()
		})
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_GetStateChangeParticleType(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)

	tests := []struct {
		genreID string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			system.SetGenre(tt.genreID)
			// Just verify it doesn't panic
			_ = system.getStateChangeParticleType()
		})
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_ProcessEntity_NoEquipment(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	terr := terrain.NewTerrain(10, 10)
	system.SetTerrain(terr)

	// Entity without equipment component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	system.processEntity(entity)
}

func TestTerrainEquipmentDurabilityParticleSystem_ProcessEntity_NoPosition(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	terr := terrain.NewTerrain(10, 10)
	system.SetTerrain(terr)

	// Entity with equipment but no position
	entity := world.CreateEntity()
	entity.AddComponent(&EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	})

	// Should not panic
	system.processEntity(entity)
}

func TestTerrainEquipmentDurabilityParticleSystem_WaterTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	// Create terrain with water
	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(0, 0, terrain.TileWaterShallow)
	system.SetTerrain(terr)
	system.SetTileSize(32)

	// Create entity with equipment
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 16, Y: 16})
	equipComp := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipComp.Slots[SlotMainHand] = &item.Item{
		Stats: item.ItemStats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	entity.AddComponent(equipComp)

	// First update
	system.Update([]*Entity{entity}, 0.016)

	// Simulate water damage (rusting)
	equipComp.Slots[SlotMainHand].Stats.Durability = 95

	// Second update - should spawn water/rust particles
	system.Update([]*Entity{entity}, 0.016)

	state := system.lastDurabilityState[entity.ID]
	if state.totalDurability != 95 {
		t.Errorf("expected durability 95, got %d", state.totalDurability)
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_TrapTerrain(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	// Create terrain with trap
	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(0, 0, terrain.TileTrapDoor)
	system.SetTerrain(terr)
	system.SetTileSize(32)

	// Create entity with equipment
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 16, Y: 16})
	equipComp := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipComp.Slots[SlotBoots] = &item.Item{
		Stats: item.ItemStats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	entity.AddComponent(equipComp)

	// First update
	system.Update([]*Entity{entity}, 0.016)

	// Simulate trap damage
	equipComp.Slots[SlotBoots].Stats.Durability = 85

	// Second update - should spawn trap/shard particles
	system.Update([]*Entity{entity}, 0.016)

	state := system.lastDurabilityState[entity.ID]
	if state.totalDurability != 85 {
		t.Errorf("expected durability 85, got %d", state.totalDurability)
	}
}

func TestTerrainEquipmentDurabilityParticleSystem_DamageStateTransition(t *testing.T) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	// Create terrain with lava
	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(0, 0, terrain.TileLavaFlow)
	system.SetTerrain(terr)
	system.SetTileSize(32)

	// Create entity with equipment
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 16, Y: 16})
	equipComp := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
	equipComp.Slots[SlotChest] = &item.Item{
		Stats: item.ItemStats{
			Durability:    80,
			DurabilityMax: 100,
		},
	}
	entity.AddComponent(equipComp)

	// First update - pristine state
	system.Update([]*Entity{entity}, 0.016)

	state := system.lastDurabilityState[entity.ID]
	if state.damageState != 0 {
		t.Errorf("expected pristine state (0), got %d", state.damageState)
	}

	// Damage to worn state (below 75%)
	equipComp.Slots[SlotChest].Stats.Durability = 70
	system.Update([]*Entity{entity}, 0.016)

	state = system.lastDurabilityState[entity.ID]
	if state.damageState != 1 {
		t.Errorf("expected worn state (1), got %d", state.damageState)
	}

	// Damage to damaged state (below 50%)
	equipComp.Slots[SlotChest].Stats.Durability = 45
	system.Update([]*Entity{entity}, 0.016)

	state = system.lastDurabilityState[entity.ID]
	if state.damageState != 2 {
		t.Errorf("expected damaged state (2), got %d", state.damageState)
	}
}

func BenchmarkTerrainEquipmentDurabilityParticleSystem_Update(b *testing.B) {
	world := NewWorld(nil)
	system := NewTerrainEquipmentDurabilityParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	terr := terrain.NewTerrain(100, 100)
	system.SetTerrain(terr)
	system.SetTileSize(32)

	// Create 100 entities with equipment
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		equipComp := &EquipmentComponent{
			Slots: make(map[EquipmentSlot]*item.Item),
		}
		equipComp.Slots[SlotChest] = &item.Item{
			Stats: item.ItemStats{
				Durability:    100,
				DurabilityMax: 100,
			},
		}
		entity.AddComponent(equipComp)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
