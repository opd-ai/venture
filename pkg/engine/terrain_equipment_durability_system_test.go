package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainEquipmentDurabilitySystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainEquipmentDurabilitySystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
	if len(sys.genreMultipliers) == 0 {
		t.Error("genre multipliers not initialized")
	}
	if len(sys.terrainDamageRates) == 0 {
		t.Error("terrain damage rates not initialized")
	}
}

func TestTerrainEquipmentDurabilitySystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	terr := terrain.New(100, 100, rand.New(rand.NewSource(12345)))
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("terrain not set correctly")
	}
}

func TestTerrainEquipmentDurabilitySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "horror")
	}
}

func TestTerrainEquipmentDurabilitySystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Invalid tile size should not change
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d on invalid input", sys.tileSize)
	}
}

func TestTerrainEquipmentDurabilitySystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		name     string
		genre    string
		wantMult float64
	}{
		{"fantasy default", "fantasy", 1.0},
		{"scifi protection", "scifi", 0.6},
		{"horror harsh", "horror", 1.4},
		{"cyberpunk moderate", "cyberpunk", 0.8},
		{"postapoc harsh", "postapoc", 1.5},
		{"unknown default", "unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.getGenreMultiplier()
			if mult != tt.wantMult {
				t.Errorf("getGenreMultiplier() = %v, want %v", mult, tt.wantMult)
			}
		})
	}
}

func TestTerrainEquipmentDurabilitySystem_SlotMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		name     string
		tile     terrain.TileType
		slot     EquipmentSlot
		wantMult float64
	}{
		{"lava boots high", terrain.TileLavaFlow, SlotBoots, 2.0},
		{"lava chest moderate", terrain.TileLavaFlow, SlotChest, 1.5},
		{"water weapon rust", terrain.TileWaterShallow, SlotMainHand, 2.0},
		{"water boots wet", terrain.TileWaterShallow, SlotBoots, 1.5},
		{"trap boots high", terrain.TileTrapDoor, SlotBoots, 2.5},
		{"trap legs moderate", terrain.TileTrapDoor, SlotLegs, 1.5},
		{"normal floor no mult", terrain.TileFloor, SlotChest, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mult := sys.getSlotMultiplier(tt.tile, tt.slot)
			if mult != tt.wantMult {
				t.Errorf("getSlotMultiplier(%v, %v) = %v, want %v",
					tt.tile, tt.slot, mult, tt.wantMult)
			}
		})
	}
}

func TestTerrainEquipmentDurabilitySystem_DamageStateChanged(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		name        string
		oldDur      int
		newDur      int
		maxDur      int
		wantChanged bool
	}{
		{"no change same", 100, 100, 100, false},
		{"slight decrease", 100, 95, 100, false},
		{"cross 75 threshold", 76, 74, 100, true},
		{"cross 50 threshold", 51, 49, 100, true},
		{"cross 25 threshold", 26, 24, 100, true},
		{"stay above 75", 90, 80, 100, false},
		{"stay between 50-75", 60, 55, 100, false},
		{"zero max durability", 10, 5, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := sys.damageStateChanged(tt.oldDur, tt.newDur, tt.maxDur)
			if changed != tt.wantChanged {
				t.Errorf("damageStateChanged(%d, %d, %d) = %v, want %v",
					tt.oldDur, tt.newDur, tt.maxDur, changed, tt.wantChanged)
			}
		})
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_NilTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic with nil terrain
	sys.Update([]*Entity{entity}, 1.0)
}

func TestTerrainEquipmentDurabilitySystem_Update_LavaDamage(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create terrain with lava
	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	// Create entity with equipment
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()

	boots := &item.Item{
		ID:   "test_boots",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
		ArmorType: item.ArmorBoots,
	}
	equipComp.Equip(boots, SlotBoots)
	entity.AddComponent(equipComp)

	// Position on lava tile (tile 3,3 = pixel 96-128)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Run update for 1 second
	sys.Update([]*Entity{entity}, 1.0)

	// Expect durability loss: 5.0 (lava) * 1.0 (fantasy) * 2.0 (boots slot) * 1.0 (time) = 10
	expectedDurability := 90
	if boots.Stats.Durability != expectedDurability {
		t.Errorf("boots durability = %d, want %d", boots.Stats.Durability, expectedDurability)
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_WaterRust(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create terrain with water
	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(2, 2, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	// Create entity with weapon
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()

	sword := &item.Item{
		ID:   "test_sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	equipComp.Equip(sword, SlotMainHand)
	entity.AddComponent(equipComp)

	// Position on water tile
	entity.AddComponent(&PositionComponent{X: 70, Y: 70})

	// Run update for 2 seconds
	sys.Update([]*Entity{entity}, 2.0)

	// Expect: 0.5 (water) * 1.0 (fantasy) * 2.0 (weapon slot) * 2.0 (time) = 2
	expectedDurability := 98
	if sword.Stats.Durability != expectedDurability {
		t.Errorf("sword durability = %d, want %d", sword.Stats.Durability, expectedDurability)
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_GenreScifi(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("scifi") // 0.6 multiplier

	// Create terrain with lava
	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	// Create entity with equipment
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()

	boots := &item.Item{
		ID:   "test_boots",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
		ArmorType: item.ArmorBoots,
	}
	equipComp.Equip(boots, SlotBoots)
	entity.AddComponent(equipComp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.Update([]*Entity{entity}, 1.0)

	// Expect: 5.0 (lava) * 0.6 (scifi) * 2.0 (boots) * 1.0 (time) = 6
	expectedDurability := 94
	if boots.Stats.Durability != expectedDurability {
		t.Errorf("boots durability = %d, want %d", boots.Stats.Durability, expectedDurability)
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_NoEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)

	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	// Entity without equipment
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.Update([]*Entity{entity}, 1.0)
}

func TestTerrainEquipmentDurabilitySystem_Update_SafeTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create terrain with safe floor
	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileFloor)
	sys.SetTerrain(terr)

	// Create entity with equipment
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()

	boots := &item.Item{
		ID:   "test_boots",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
		ArmorType: item.ArmorBoots,
	}
	equipComp.Equip(boots, SlotBoots)
	entity.AddComponent(equipComp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.Update([]*Entity{entity}, 1.0)

	// No damage on safe terrain
	if boots.Stats.Durability != 100 {
		t.Errorf("boots durability = %d, want 100 (no damage)", boots.Stats.Durability)
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_MarkVisualDirty(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()
	visualComp := NewEquipmentVisualComponent()

	// Start at 76% durability, will cross 75% threshold
	boots := &item.Item{
		ID:   "test_boots",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Durability:    76,
			DurabilityMax: 100,
		},
		ArmorType: item.ArmorBoots,
	}
	equipComp.Equip(boots, SlotBoots)
	entity.AddComponent(equipComp)
	entity.AddComponent(visualComp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Clear dirty flag
	visualComp.MarkClean()

	sys.Update([]*Entity{entity}, 1.0)

	// Should have crossed 75% threshold and marked visual dirty
	if !visualComp.Dirty {
		t.Error("visual component should be dirty after crossing damage threshold")
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_ZeroDurabilityMax(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()

	// Item with no durability tracking
	boots := &item.Item{
		ID:   "test_boots",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Durability:    0,
			DurabilityMax: 0, // No durability system
		},
		ArmorType: item.ArmorBoots,
	}
	equipComp.Equip(boots, SlotBoots)
	entity.AddComponent(equipComp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.Update([]*Entity{entity}, 1.0)

	// Durability unchanged
	if boots.Stats.Durability != 0 {
		t.Errorf("boots durability changed unexpectedly to %d", boots.Stats.Durability)
	}
}

func TestTerrainEquipmentDurabilitySystem_Update_MinDurabilityZero(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("postapoc") // 1.5 multiplier for faster damage

	terr := terrain.New(10, 10, rand.New(rand.NewSource(12345)))
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()

	// Low durability item
	boots := &item.Item{
		ID:   "test_boots",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Durability:    5,
			DurabilityMax: 100,
		},
		ArmorType: item.ArmorBoots,
	}
	equipComp.Equip(boots, SlotBoots)
	entity.AddComponent(equipComp)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Long time should not go negative
	sys.Update([]*Entity{entity}, 10.0)

	if boots.Stats.Durability < 0 {
		t.Errorf("boots durability went negative: %d", boots.Stats.Durability)
	}
}

func BenchmarkTerrainEquipmentDurabilitySystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	terr := terrain.New(100, 100, rand.New(rand.NewSource(12345)))
	// Scatter some hazardous tiles
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if (i+j)%5 == 0 {
				terr.SetTile(i, j, terrain.TileLavaFlow)
			}
		}
	}
	sys.SetTerrain(terr)

	// Create 100 entities with equipment
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		equipComp := NewEquipmentComponent()
		boots := &item.Item{
			ID:   "boots",
			Type: item.TypeArmor,
			Stats: item.Stats{
				Durability:    100,
				DurabilityMax: 100,
			},
			ArmorType: item.ArmorBoots,
		}
		equipComp.Equip(boots, SlotBoots)
		entity.AddComponent(equipComp)
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // ~60 FPS
	}
}
