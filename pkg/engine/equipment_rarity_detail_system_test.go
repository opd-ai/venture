package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestRarityDetailComponent_Type(t *testing.T) {
	comp := NewRarityDetailComponent()
	if comp.Type() != "rarity_detail" {
		t.Errorf("expected type 'rarity_detail', got %q", comp.Type())
	}
}

func TestRarityDetailComponent_Defaults(t *testing.T) {
	comp := NewRarityDetailComponent()
	if comp.DetailLevel != 0.3 {
		t.Errorf("expected default DetailLevel 0.3, got %f", comp.DetailLevel)
	}
	if comp.Enabled {
		t.Error("expected Enabled to be false by default")
	}
}

func TestEquipmentRarityDetailSystem_NilWorld(t *testing.T) {
	sys := NewEquipmentRarityDetailSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
}

func TestEquipmentRarityDetailSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"sci-fi", "sci-fi"},
		{"postapoc", "postapoc"},
		{"unknown", "unknown_genre"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewEquipmentRarityDetailSystem(nil, 42)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected genreID %q, got %q", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestEquipmentRarityDetailSystem_NoEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentRarityDetailSystem(world, 42)

	entity := NewEntity(0)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0) // Triggers full scan

	_, ok := entity.GetComponent("rarity_detail")
	if ok {
		t.Error("should not attach rarity_detail to entity without equipment")
	}
}

func TestEquipmentRarityDetailSystem_WithEquipment(t *testing.T) {
	tests := []struct {
		name            string
		rarity          item.Rarity
		expectedMinDL   float64
		expectedMaxDL   float64
		expectedHighest string
	}{
		{"common_item", item.RarityCommon, 0.25, 0.35, "common"},
		{"uncommon_item", item.RarityUncommon, 0.35, 0.45, "uncommon"},
		{"rare_item", item.RarityRare, 0.55, 0.65, "rare"},
		{"epic_item", item.RarityEpic, 0.75, 0.85, "epic"},
		{"legendary_item", item.RarityLegendary, 0.95, 1.05, "legendary"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentRarityDetailSystem(world, 42)

			entity := NewEntity(0)
			equip := NewEquipmentComponent()
			equip.Slots[SlotChest] = &item.Item{
				Name:   "Test Armor",
				Type:   item.TypeArmor,
				Rarity: tt.rarity,
				Stats:  item.Stats{Defense: 10, Durability: 100, DurabilityMax: 100},
			}
			entity.AddComponent(equip)
			world.AddEntity(entity)

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			comp, ok := entity.GetComponent("rarity_detail")
			if !ok {
				t.Fatal("expected rarity_detail component to be attached")
			}
			detail := comp.(*RarityDetailComponent)
			if !detail.Enabled {
				t.Error("expected Enabled to be true")
			}
			if detail.DetailLevel < tt.expectedMinDL || detail.DetailLevel > tt.expectedMaxDL {
				t.Errorf("DetailLevel %f not in range [%f, %f]", detail.DetailLevel, tt.expectedMinDL, tt.expectedMaxDL)
			}
			if detail.HighestRarity != tt.expectedHighest {
				t.Errorf("expected HighestRarity %q, got %q", tt.expectedHighest, detail.HighestRarity)
			}
			if detail.EquippedCount != 1 {
				t.Errorf("expected EquippedCount 1, got %d", detail.EquippedCount)
			}
		})
	}
}

func TestEquipmentRarityDetailSystem_MixedRarities(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentRarityDetailSystem(world, 42)

	entity := NewEntity(0)
	equip := NewEquipmentComponent()
	// Common weapon + Legendary chest = blended detail
	equip.Slots[SlotMainHand] = &item.Item{
		Name:   "Rusty Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Damage: 5, Durability: 100, DurabilityMax: 100},
	}
	equip.Slots[SlotChest] = &item.Item{
		Name:   "Legendary Plate",
		Type:   item.TypeArmor,
		Rarity: item.RarityLegendary,
		Stats:  item.Stats{Defense: 50, Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equip)
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	comp, _ := entity.GetComponent("rarity_detail")
	detail := comp.(*RarityDetailComponent)

	if detail.HighestRarity != "legendary" {
		t.Errorf("expected highest rarity 'Legendary', got %q", detail.HighestRarity)
	}
	if detail.EquippedCount != 2 {
		t.Errorf("expected 2 equipped items, got %d", detail.EquippedCount)
	}
	// avg = (0.3 + 1.0)/2 = 0.65, peak = 1.0, blended = 0.65*0.6 + 1.0*0.4 = 0.79
	if detail.DetailLevel < 0.7 || detail.DetailLevel > 0.85 {
		t.Errorf("blended detail %f not in expected range [0.7, 0.85]", detail.DetailLevel)
	}
}

func TestEquipmentRarityDetailSystem_Throttled(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentRarityDetailSystem(world, 42)

	entity := NewEntity(0)
	equip := NewEquipmentComponent()
	equip.Slots[SlotChest] = &item.Item{
		Name:   "Iron Armor",
		Type:   item.TypeArmor,
		Rarity: item.RarityRare,
		Stats:  item.Stats{Defense: 20, Durability: 100, DurabilityMax: 100},
	}
	entity.AddComponent(equip)
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// First update at dt=1.0 triggers full scan
	sys.Update(entities, 1.0)

	comp, ok := entity.GetComponent("rarity_detail")
	if !ok {
		t.Fatal("expected component after first scan")
	}
	detail := comp.(*RarityDetailComponent)
	if !detail.Enabled {
		t.Error("expected enabled after first scan")
	}

	// Short dt (0.1s) should NOT trigger a full rescan
	detail.HighestRarity = "test_marker"
	sys.Update(entities, 0.1)
	if detail.HighestRarity != "test_marker" {
		t.Error("expected throttled update to skip recomputation")
	}
}

func TestEquipmentRarityDetailSystem_GenreScaling(t *testing.T) {
	tests := []struct {
		name               string
		genreID            string
		expectHighVibrancy bool
	}{
		{"cyberpunk_high_vibrancy", "cyberpunk", true},
		{"postapoc_low_vibrancy", "postapoc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentRarityDetailSystem(world, 42)
			sys.SetGenre(tt.genreID)

			entity := NewEntity(0)
			equip := NewEquipmentComponent()
			equip.Slots[SlotChest] = &item.Item{
				Name:   "Epic Armor",
				Type:   item.TypeArmor,
				Rarity: item.RarityEpic,
				Stats:  item.Stats{Defense: 30, Durability: 100, DurabilityMax: 100},
			}
			entity.AddComponent(equip)
			world.AddEntity(entity)

			sys.Update([]*Entity{entity}, 1.0)

			comp, _ := entity.GetComponent("rarity_detail")
			detail := comp.(*RarityDetailComponent)

			if tt.expectHighVibrancy && detail.ColorVibrancy < 0.7 {
				t.Errorf("expected high vibrancy for %s, got %f", tt.genreID, detail.ColorVibrancy)
			}
			if !tt.expectHighVibrancy && detail.ColorVibrancy > 0.5 {
				t.Errorf("expected low vibrancy for %s, got %f", tt.genreID, detail.ColorVibrancy)
			}
		})
	}
}

func BenchmarkEquipmentRarityDetailSystem(b *testing.B) {
	world := NewWorld()
	sys := NewEquipmentRarityDetailSystem(world, 42)

	entities := make([]*Entity, 200)
	rarities := []item.Rarity{item.RarityCommon, item.RarityUncommon, item.RarityRare, item.RarityEpic, item.RarityLegendary}
	for i := range entities {
		e := NewEntity(0)
		equip := NewEquipmentComponent()
		equip.Slots[SlotChest] = &item.Item{
			Name:   "Armor",
			Type:   item.TypeArmor,
			Rarity: rarities[i%len(rarities)],
			Stats:  item.Stats{Defense: 10, Durability: 100, DurabilityMax: 100},
		}
		equip.Slots[SlotMainHand] = &item.Item{
			Name:   "Sword",
			Type:   item.TypeWeapon,
			Rarity: rarities[(i+2)%len(rarities)],
			Stats:  item.Stats{Damage: 10, Durability: 100, DurabilityMax: 100},
		}
		e.AddComponent(equip)
		world.AddEntity(e)
		entities[i] = e
	}

	// Prime: attach components
	sys.Update(entities, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 1.0)
	}
}
