package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestEquipmentWearTintComponentType(t *testing.T) {
	comp := NewEquipmentWearTintComponent()
	if comp.Type() != "equipment_wear_tint" {
		t.Errorf("expected equipment_wear_tint, got %s", comp.Type())
	}
}

func TestEquipmentWearTintComponentDefaults(t *testing.T) {
	comp := NewEquipmentWearTintComponent()
	if comp.Enabled {
		t.Error("expected disabled by default")
	}
	if comp.OpacityMultiplier != 1.0 {
		t.Errorf("expected 1.0 opacity, got %f", comp.OpacityMultiplier)
	}
	if comp.ColorDarken != 0.0 {
		t.Errorf("expected 0.0 darken, got %f", comp.ColorDarken)
	}
	if comp.CrackDensity != 0.0 {
		t.Errorf("expected 0.0 crack density, got %f", comp.CrackDensity)
	}
}

func TestEquipmentDamageStateTintSystemCreation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected fantasy default genre, got %s", sys.genreID)
	}
}

func TestEquipmentDamageStateTintSystemSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"sci-fi", "sci-fi"},
		{"post-apocalyptic", "post-apocalyptic"},
		{"unknown", "steampunk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentDamageStateTintSystem(world, 42)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected %s, got %s", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestEquipmentDamageStateTintSystemNoEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	entities := []*Entity{entity}

	// Trigger full scan
	sys.Update(entities, 1.1)

	_, hasTint := entity.GetComponent("equipment_wear_tint")
	if hasTint {
		t.Error("entity without equipment should not have equipment_wear_tint")
	}
}

func TestEquipmentDamageStateTintSystemPristine(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	sword := &item.Item{
		ID:   "sword-1",
		Seed: 100,
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	equipComp.Equip(sword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, hasTint := entity.GetComponent("equipment_wear_tint")
	if !hasTint {
		t.Fatal("expected equipment_wear_tint component")
	}
	tint := comp.(*EquipmentWearTintComponent)

	// Pristine equipment should not enable tint
	if tint.Enabled {
		t.Error("pristine equipment should not enable tint")
	}
	if tint.OpacityMultiplier != 1.0 {
		t.Errorf("expected 1.0 opacity for pristine, got %f", tint.OpacityMultiplier)
	}
	if tint.ColorDarken != 0.0 {
		t.Errorf("expected 0.0 darken for pristine, got %f", tint.ColorDarken)
	}
}

func TestEquipmentDamageStateTintSystemWorn(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	armor := &item.Item{
		ID:        "armor-1",
		Seed:      200,
		Type:      item.TypeArmor,
		ArmorType: item.ArmorChest,
		Stats: item.Stats{
			Durability:    60,
			DurabilityMax: 100,
		},
	}
	equipComp.Equip(armor, SlotChest)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, hasTint := entity.GetComponent("equipment_wear_tint")
	if !hasTint {
		t.Fatal("expected equipment_wear_tint component")
	}
	tint := comp.(*EquipmentWearTintComponent)

	if !tint.Enabled {
		t.Error("worn equipment should enable tint")
	}
	if tint.OpacityMultiplier >= 1.0 {
		t.Errorf("worn equipment should reduce opacity, got %f", tint.OpacityMultiplier)
	}
	if tint.ColorDarken <= 0.0 {
		t.Errorf("worn equipment should have some darkening, got %f", tint.ColorDarken)
	}
	if tint.EquippedCount != 1 {
		t.Errorf("expected 1 equipped item, got %d", tint.EquippedCount)
	}
}

func TestEquipmentDamageStateTintSystemBroken(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	brokenSword := &item.Item{
		ID:   "sword-broken",
		Seed: 300,
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Durability:    5,
			DurabilityMax: 100,
		},
	}
	equipComp.Equip(brokenSword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("equipment_wear_tint")
	tint := comp.(*EquipmentWearTintComponent)

	if !tint.Enabled {
		t.Error("broken equipment should enable tint")
	}
	if tint.OpacityMultiplier > 0.75 {
		t.Errorf("broken equipment should heavily reduce opacity, got %f", tint.OpacityMultiplier)
	}
	if tint.CrackDensity < 0.5 {
		t.Errorf("broken equipment should have high crack density, got %f", tint.CrackDensity)
	}
	if tint.WorstState != "broken" {
		t.Errorf("expected worst state broken, got %s", tint.WorstState)
	}
}

func TestEquipmentDamageStateTintSystemMultipleItems(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()

	// Mix pristine and damaged items
	pristineSword := &item.Item{
		ID: "sword-p", Seed: 100, Type: item.TypeWeapon,
		Stats: item.Stats{Durability: 100, DurabilityMax: 100},
	}
	damagedArmor := &item.Item{
		ID: "armor-d", Seed: 200, Type: item.TypeArmor, ArmorType: item.ArmorChest,
		Stats: item.Stats{Durability: 30, DurabilityMax: 100},
	}
	equipComp.Equip(pristineSword, SlotMainHand)
	equipComp.Equip(damagedArmor, SlotChest)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("equipment_wear_tint")
	tint := comp.(*EquipmentWearTintComponent)

	if !tint.Enabled {
		t.Error("mixed damage should enable tint")
	}
	if tint.EquippedCount != 2 {
		t.Errorf("expected 2 equipped items, got %d", tint.EquippedCount)
	}
	// Average of pristine(1.0) and damaged(0.85) = 0.925
	if tint.OpacityMultiplier > 0.97 || tint.OpacityMultiplier < 0.88 {
		t.Errorf("expected blended opacity ~0.925, got %f", tint.OpacityMultiplier)
	}
}

func TestEquipmentDamageStateTintSystemGenrePresets(t *testing.T) {
	tests := []struct {
		name           string
		genreID        string
		expectHighDirt bool
		expectHighDark bool
	}{
		{"horror darkens more", "horror", true, true},
		{"scifi minimal dirt", "sci-fi", false, false},
		{"postapoc max dirt", "postapoc", true, true},
		{"cyberpunk moderate", "cyberpunk", false, false},
		{"fantasy default", "fantasy", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentDamageStateTintSystem(world, 42)
			sys.SetGenre(tt.genreID)

			entity := NewEntity(1)
			equipComp := NewEquipmentComponent()
			worn := &item.Item{
				ID: "helm-w", Seed: 400, Type: item.TypeArmor,
				Stats: item.Stats{Durability: 60, DurabilityMax: 100},
			}
			equipComp.Equip(worn, SlotHead)
			entity.AddComponent(equipComp)

			sys.Update([]*Entity{entity}, 1.1)

			comp, _ := entity.GetComponent("equipment_wear_tint")
			tint := comp.(*EquipmentWearTintComponent)

			if tt.expectHighDirt && tint.Dirtiness <= 0.15 {
				t.Errorf("expected higher dirtiness for %s, got %f", tt.genreID, tint.Dirtiness)
			}
		})
	}
}

func TestEquipmentDamageStateTintSystemThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	sword := &item.Item{
		ID: "sword-1", Seed: 100, Type: item.TypeWeapon,
		Stats: item.Stats{Durability: 50, DurabilityMax: 100},
	}
	equipComp.Equip(sword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// First update triggers full scan
	sys.Update(entities, 1.1)

	_, hasTint := entity.GetComponent("equipment_wear_tint")
	if !hasTint {
		t.Fatal("expected component after full scan")
	}

	// Small delta should NOT trigger full scan for new entities
	entity2 := NewEntity(2)
	entity2.AddComponent(NewEquipmentComponent())
	entities = append(entities, entity2)
	sys.Update(entities, 0.1)

	_, hasTint2 := entity2.GetComponent("equipment_wear_tint")
	if hasTint2 {
		t.Error("short delta should not trigger full scan for new entities")
	}
}

func TestEquipmentDamageStateTintSystemNoItems(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, hasTint := entity.GetComponent("equipment_wear_tint")
	if !hasTint {
		t.Fatal("expected component even with empty equipment")
	}
	tint := comp.(*EquipmentWearTintComponent)

	if tint.Enabled {
		t.Error("empty equipment should not enable tint")
	}
	if tint.EquippedCount != 0 {
		t.Errorf("expected 0 equipped, got %d", tint.EquippedCount)
	}
}

func BenchmarkEquipmentDamageStateTintSystem(b *testing.B) {
	world := NewWorld()
	sys := NewEquipmentDamageStateTintSystem(world, 42)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		equipComp := NewEquipmentComponent()
		sword := &item.Item{
			ID: "sword", Seed: int64(i), Type: item.TypeWeapon,
			Stats: item.Stats{Durability: 50 + i%50, DurabilityMax: 100},
		}
		equipComp.Equip(sword, SlotMainHand)
		e.AddComponent(equipComp)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 2.0 // Force full scan
		sys.Update(entities, 0.016)
	}
}
