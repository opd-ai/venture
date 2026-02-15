package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestMaterialSheenComponentType(t *testing.T) {
	comp := NewMaterialSheenComponent()
	if comp.Type() != "material_sheen" {
		t.Errorf("expected material_sheen, got %s", comp.Type())
	}
}

func TestMaterialSheenComponentDefaults(t *testing.T) {
	comp := NewMaterialSheenComponent()
	if comp.Enabled {
		t.Error("expected disabled by default")
	}
	if comp.SheenIntensity != 0.0 {
		t.Errorf("expected 0.0 sheen, got %f", comp.SheenIntensity)
	}
	if comp.PulseSpeed != 1.5 {
		t.Errorf("expected 1.5 pulse speed, got %f", comp.PulseSpeed)
	}
}

func TestEquipmentMaterialSheenSystemCreation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected fantasy default genre, got %s", sys.genreID)
	}
}

func TestEquipmentMaterialSheenSystemSetGenre(t *testing.T) {
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
			sys := NewEquipmentMaterialSheenSystem(world, 42)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected %s, got %s", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestEquipmentMaterialSheenSystemNoEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)
	entities := []*Entity{entity}

	// First update with full scan
	sys.Update(entities, 1.1)

	// Entity without equipment should not get a sheen component
	_, hasSheen := entity.GetComponent("material_sheen")
	if hasSheen {
		t.Error("entity without equipment should not have material_sheen")
	}
}

func TestEquipmentMaterialSheenSystemWithEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)

	// Add equipment component with a metal weapon
	equipComp := NewEquipmentComponent()
	metalSword := &item.Item{
		ID:   "sword-1",
		Seed: 100,
		Tags: []string{"metal"},
		Type: item.TypeWeapon,
	}
	equipComp.Equip(metalSword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// Trigger full scan
	sys.Update(entities, 1.1)

	comp, hasSheen := entity.GetComponent("material_sheen")
	if !hasSheen {
		t.Fatal("entity with equipment should have material_sheen")
	}

	sheen, ok := comp.(*MaterialSheenComponent)
	if !ok {
		t.Fatal("expected *MaterialSheenComponent")
	}

	// Metal has Sheen=0.9 -> should be high intensity
	if sheen.SheenIntensity < 0.5 {
		t.Errorf("metal sheen intensity should be high, got %f", sheen.SheenIntensity)
	}
	if !sheen.Enabled {
		t.Error("sheen should be enabled for metal equipment")
	}
	if sheen.DominantMaterial != "metal" {
		t.Errorf("expected dominant material metal, got %s", sheen.DominantMaterial)
	}
}

func TestEquipmentMaterialSheenSystemClothLowSheen(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	clothRobe := &item.Item{
		ID:        "robe-1",
		Seed:      200,
		Tags:      []string{"cloth"},
		Type:      item.TypeArmor,
		ArmorType: item.ArmorChest,
	}
	equipComp.Equip(clothRobe, SlotChest)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("material_sheen")
	sheen := comp.(*MaterialSheenComponent)

	// Cloth has Sheen=0.1 -> should be low intensity
	if sheen.SheenIntensity > 0.3 {
		t.Errorf("cloth sheen should be low, got %f", sheen.SheenIntensity)
	}
}

func TestEquipmentMaterialSheenSystemPhaseAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	metalSword := &item.Item{
		ID:   "sword-1",
		Seed: 100,
		Tags: []string{"metal"},
		Type: item.TypeWeapon,
	}
	equipComp.Equip(metalSword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// Trigger full scan to attach component
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("material_sheen")
	sheen := comp.(*MaterialSheenComponent)
	initialPhase := sheen.Phase

	// Subsequent update should advance phase without full scan
	sys.Update(entities, 0.1)

	if sheen.Phase <= initialPhase {
		t.Error("phase should advance each update")
	}
}

func TestEquipmentMaterialSheenSystemPhaseWraps(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	metalSword := &item.Item{
		ID:   "sword-1",
		Seed: 100,
		Tags: []string{"metal"},
		Type: item.TypeWeapon,
	}
	equipComp.Equip(metalSword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// Trigger full scan
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("material_sheen")
	sheen := comp.(*MaterialSheenComponent)

	// Advance past 2π to test wrap
	sheen.Phase = 6.0
	sys.Update(entities, 1.0)

	if sheen.Phase > 2*math.Pi {
		t.Errorf("phase should wrap at 2π, got %f", sheen.Phase)
	}
}

func TestEquipmentMaterialSheenSystemGenrePresets(t *testing.T) {
	tests := []struct {
		name            string
		genre           string
		wantHigherPulse bool
		wantLowerScale  bool
	}{
		{"cyberpunk has high pulse", "cyberpunk", true, false},
		{"horror has low intensity", "horror", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEquipmentMaterialSheenSystem(world, 42)
			sys.SetGenre(tt.genre)

			fantasyPreset := sys.getPreset("fantasy")

			if tt.wantHigherPulse && sys.preset.PulseSpeed <= fantasyPreset.PulseSpeed {
				t.Errorf("expected higher pulse speed than fantasy for %s", tt.genre)
			}
			if tt.wantLowerScale && sys.preset.IntensityScale >= fantasyPreset.IntensityScale {
				t.Errorf("expected lower intensity scale than fantasy for %s", tt.genre)
			}
		})
	}
}

func TestEquipmentMaterialSheenSystemMultipleItems(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()

	// Equip metal weapon + cloth armor => averaged sheen
	metalSword := &item.Item{ID: "sword-1", Seed: 100, Tags: []string{"metal"}, Type: item.TypeWeapon}
	clothArmor := &item.Item{ID: "robe-1", Seed: 200, Tags: []string{"cloth"}, Type: item.TypeArmor, ArmorType: item.ArmorChest}
	equipComp.Equip(metalSword, SlotMainHand)
	equipComp.Equip(clothArmor, SlotChest)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	sys.Update(entities, 1.1)

	comp, _ := entity.GetComponent("material_sheen")
	sheen := comp.(*MaterialSheenComponent)

	// Average of metal (0.9) and cloth (0.1) sheen = 0.5
	if sheen.SheenIntensity < 0.3 || sheen.SheenIntensity > 0.7 {
		t.Errorf("expected averaged sheen ~0.5, got %f", sheen.SheenIntensity)
	}
}

func TestEquipmentMaterialSheenSystemThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentMaterialSheenSystem(world, 42)

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	metalSword := &item.Item{ID: "sword-1", Seed: 100, Tags: []string{"metal"}, Type: item.TypeWeapon}
	equipComp.Equip(metalSword, SlotMainHand)
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}

	// First update at dt=0.5 (below interval) should not create component
	sys.Update(entities, 0.5)
	_, hasSheen := entity.GetComponent("material_sheen")
	if hasSheen {
		t.Error("should not attach component before interval elapses")
	}

	// Next update pushes past interval, triggering full scan
	sys.Update(entities, 0.6)
	_, hasSheen = entity.GetComponent("material_sheen")
	if !hasSheen {
		t.Error("should attach component after interval elapses")
	}
}
