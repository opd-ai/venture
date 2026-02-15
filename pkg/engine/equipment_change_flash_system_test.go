package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewEquipmentChangeFlashSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)

	if sys == nil {
		t.Fatal("NewEquipmentChangeFlashSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.lastEquipState == nil {
		t.Error("lastEquipState not initialized")
	}
}

func TestEquipmentChangeFlashSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		want    string
	}{
		{"fantasy", "fantasy", "gold"},
		{"scifi", "scifi", "cyan"},
		{"horror", "horror", "dark_red"},
		{"cyberpunk", "cyberpunk", "neon_green"},
		{"postapoc", "postapoc", "amber"},
		{"unknown", "steampunk", "white"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewEquipmentChangeFlashSystem(nil, 1)
			sys.SetGenre(tt.genreID)
			got := sys.genreFlashColor()
			if got != tt.want {
				t.Errorf("genreFlashColor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEquipmentChangeFlashSystem_UpdateNoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	// No particle system set — should not panic
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(NewEquipmentComponent())
	sys.Update([]*Entity{entity}, 0.016)
}

func TestEquipmentChangeFlashSystem_FirstSightSnapshot(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	ec := NewEquipmentComponent()
	entity.AddComponent(ec)

	// First update should snapshot, not flash
	sys.Update([]*Entity{entity}, 0.016)

	state := sys.GetLastEquipState()
	if _, ok := state[entity.ID]; !ok {
		t.Fatal("entity state not tracked after first update")
	}
}

func TestEquipmentChangeFlashSystem_NoChangeNoFlash(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	ec := NewEquipmentComponent()
	entity.AddComponent(ec)

	// First update: snapshot
	sys.Update([]*Entity{entity}, 0.016)
	// Second update: no change
	sys.Update([]*Entity{entity}, 0.016)

	// State should remain the same (no panic, no crash)
	state := sys.GetLastEquipState()
	if _, ok := state[entity.ID]; !ok {
		t.Fatal("entity state lost after second update")
	}
}

func TestEquipmentChangeFlashSystem_DetectsEquipChange(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	ec := NewEquipmentComponent()
	entity.AddComponent(ec)

	// First update: snapshot empty state
	sys.Update([]*Entity{entity}, 0.016)

	// Equip a weapon
	weapon := &item.Item{
		ID:   "sword_01",
		Type: item.TypeWeapon,
	}
	ec.Slots[SlotMainHand] = weapon

	// Second update: should detect change and update state
	sys.Update([]*Entity{entity}, 0.016)

	state := sys.GetLastEquipState()
	if state[entity.ID][SlotMainHand] != "sword_01" {
		t.Errorf("state not updated, got %q want %q", state[entity.ID][SlotMainHand], "sword_01")
	}
}

func TestEquipmentChangeFlashSystem_DetectsUnequip(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("scifi")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5})
	ec := NewEquipmentComponent()
	ec.Slots[SlotChest] = &item.Item{ID: "armor_01", Type: item.TypeArmor}
	entity.AddComponent(ec)

	// Snapshot with armor
	sys.Update([]*Entity{entity}, 0.016)
	if sys.GetLastEquipState()[entity.ID][SlotChest] != "armor_01" {
		t.Fatal("initial snapshot wrong")
	}

	// Remove armor
	delete(ec.Slots, SlotChest)
	sys.Update([]*Entity{entity}, 0.016)

	if sys.GetLastEquipState()[entity.ID][SlotChest] != "" {
		t.Error("state not updated after unequip")
	}
}

func TestEquipmentChangeFlashSystem_SkipsNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	// No position component
	ec := NewEquipmentComponent()
	entity.AddComponent(ec)

	sys.Update([]*Entity{entity}, 0.016)

	if _, ok := sys.GetLastEquipState()[entity.ID]; ok {
		t.Error("should not track entity without position")
	}
}

func TestEquipmentChangeFlashSystem_SkipsNoEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 1, Y: 1})

	sys.Update([]*Entity{entity}, 0.016)

	if _, ok := sys.GetLastEquipState()[entity.ID]; ok {
		t.Error("should not track entity without equipment")
	}
}

func TestEquipmentChangeFlashSystem_MultipleSlotChange(t *testing.T) {
	world := NewWorld()
	sys := NewEquipmentChangeFlashSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("horror")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	ec := NewEquipmentComponent()
	entity.AddComponent(ec)

	// Snapshot empty
	sys.Update([]*Entity{entity}, 0.016)

	// Equip multiple items at once
	ec.Slots[SlotMainHand] = &item.Item{ID: "axe", Type: item.TypeWeapon}
	ec.Slots[SlotHead] = &item.Item{ID: "helm", Type: item.TypeArmor}

	sys.Update([]*Entity{entity}, 0.016)

	state := sys.GetLastEquipState()[entity.ID]
	if state[SlotMainHand] != "axe" || state[SlotHead] != "helm" {
		t.Errorf("multi-slot change not tracked: weapon=%q head=%q", state[SlotMainHand], state[SlotHead])
	}
}

func TestEquipmentChangeFlashSystem_ItemIDForSlotNil(t *testing.T) {
	sys := NewEquipmentChangeFlashSystem(nil, 1)
	ec := NewEquipmentComponent()
	got := sys.itemIDForSlot(ec, SlotMainHand)
	if got != "" {
		t.Errorf("itemIDForSlot empty slot = %q, want empty", got)
	}
}

func TestEquipmentChangeFlashSystem_SnapshotSlots(t *testing.T) {
	sys := NewEquipmentChangeFlashSystem(nil, 1)
	ec := NewEquipmentComponent()
	ec.Slots[SlotMainHand] = &item.Item{ID: "sword", Type: item.TypeWeapon}
	ec.Slots[SlotBoots] = &item.Item{ID: "boots", Type: item.TypeArmor}

	snap := sys.snapshotSlots(ec)
	if snap[SlotMainHand] != "sword" {
		t.Errorf("snapshot weapon = %q, want sword", snap[SlotMainHand])
	}
	if snap[SlotBoots] != "boots" {
		t.Errorf("snapshot boots = %q, want boots", snap[SlotBoots])
	}
	if snap[SlotHead] != "" {
		t.Errorf("snapshot empty head = %q, want empty", snap[SlotHead])
	}
}
