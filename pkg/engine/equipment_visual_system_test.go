package engine

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestNewEquipmentVisualSystem tests system initialization.
func TestNewEquipmentVisualSystem(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	if sys == nil {
		t.Fatal("Expected non-nil system")
	}

	if sys.spriteGenerator == nil {
		t.Error("Expected sprite generator to be set")
	}
}

// TestEquipmentVisualSystem_Update_NoEquipment tests update with no equipment.
func TestEquipmentVisualSystem_Update_NoEquipment(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	// Add equipment visual component but no sprite component
	equipComp := NewEquipmentVisualComponent()
	entity.AddComponent(equipComp)

	entities := []*Entity{entity}
	err := sys.Update(entities, 0.016)

	if err != nil {
		t.Errorf("Update should not error: %v", err)
	}

	// Should remain dirty if no sprite component exists
	if !equipComp.Dirty {
		t.Error("Component should remain dirty if no sprite component exists")
	}
}

// TestEquipmentVisualSystem_Update_DirtyEquipment tests sprite regeneration.
func TestEquipmentVisualSystem_Update_DirtyEquipment(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	// Add equipment visual component
	equipComp := NewEquipmentVisualComponent()
	equipComp.SetWeapon("sword_001", 12345)
	equipComp.SetArmor("plate_armor_001", 54321)
	entity.AddComponent(equipComp)

	// Add sprite component
	spriteComp := NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	entity.AddComponent(spriteComp)

	// Add animation component for base sprite
	animComp := NewAnimationComponent(12345)
	entity.AddComponent(animComp)

	if !equipComp.Dirty {
		t.Fatal("Equipment component should start dirty")
	}

	entities := []*Entity{entity}
	err := sys.Update(entities, 0.016)

	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if equipComp.Dirty {
		t.Error("Equipment component should be clean after update")
	}

	if spriteComp.Image == nil {
		t.Error("Expected sprite image to be generated")
	}
}

// TestEquipmentVisualSystem_Update_CleanEquipment tests skipping clean entities.
func TestEquipmentVisualSystem_Update_CleanEquipment(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	equipComp := NewEquipmentVisualComponent()
	equipComp.SetWeapon("sword_001", 12345)
	equipComp.MarkClean()
	entity.AddComponent(equipComp)

	spriteComp := NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	entity.AddComponent(spriteComp)

	animComp := NewAnimationComponent(12345)
	entity.AddComponent(animComp)

	// Store original image
	originalImg := spriteComp.Image

	entities := []*Entity{entity}
	err := sys.Update(entities, 0.016)

	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Should not regenerate since not dirty
	if spriteComp.Image != originalImg {
		t.Error("Should not regenerate sprite if component is clean")
	}
}

// TestEquipmentVisualSystem_EquipItem tests equipping items.
func TestEquipmentVisualSystem_EquipItem(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	equipComp := NewEquipmentVisualComponent()
	equipComp.Dirty = false
	entity.AddComponent(equipComp)

	spriteComp := NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	entity.AddComponent(spriteComp)

	animComp := NewAnimationComponent(12345)
	entity.AddComponent(animComp)

	// Equip weapon
	sys.EquipItem(entity, "weapon", "sword_001", 12345)

	if !equipComp.HasWeapon() {
		t.Error("Expected weapon to be equipped")
	}

	if !equipComp.Dirty {
		t.Error("Expected equipment component to be dirty after equipping")
	}

	// Update to regenerate sprite
	entities := []*Entity{entity}
	err := sys.Update(entities, 0.016)

	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if spriteComp.Image == nil {
		t.Error("Expected sprite image to be generated")
	}
}

// TestEquipmentVisualSystem_UnequipItem tests unequipping items.
func TestEquipmentVisualSystem_UnequipItem(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	equipComp := NewEquipmentVisualComponent()
	equipComp.SetWeapon("sword_001", 12345)
	equipComp.SetArmor("plate_armor_001", 54321)
	equipComp.MarkClean()
	entity.AddComponent(equipComp)

	spriteComp := NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	entity.AddComponent(spriteComp)

	animComp := NewAnimationComponent(12345)
	entity.AddComponent(animComp)

	// Unequip weapon
	sys.UnequipItem(entity, "weapon")

	if equipComp.HasWeapon() {
		t.Error("Expected weapon to be unequipped")
	}

	if !equipComp.Dirty {
		t.Error("Expected equipment component to be dirty after unequipping")
	}

	// Update to regenerate sprite
	entities := []*Entity{entity}
	err := sys.Update(entities, 0.016)

	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if spriteComp.Image == nil {
		t.Error("Expected sprite image to be regenerated")
	}

	if !equipComp.HasArmor() {
		t.Error("Armor should still be equipped")
	}
}

// BenchmarkEquipmentVisualSystem_Update benchmarks system update.
func BenchmarkEquipmentVisualSystem_Update(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	// Create 100 entities with various equipment
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))

		equipComp := NewEquipmentVisualComponent()
		if i%3 == 0 {
			equipComp.SetWeapon("sword_001", int64(i*1000))
		}
		if i%5 == 0 {
			equipComp.SetArmor("plate_armor_001", int64(i*2000))
		}
		entity.AddComponent(equipComp)

		entity.AddComponent(NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255}))
		entity.AddComponent(NewAnimationComponent(int64(i * 1000)))

		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.Update(entities, 0.016)

		// Mark some dirty for next iteration
		if i%10 == 0 {
			for j := 0; j < 10; j++ {
				equipComp, _ := entities[j].GetComponent("equipment_visual")
				equipComp.(*EquipmentVisualComponent).MarkDirty()
			}
		}
	}
}

// BenchmarkEquipmentVisualSystem_EquipItem benchmarks item equipping.
func BenchmarkEquipmentVisualSystem_EquipItem(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)
	entity.AddComponent(NewEquipmentVisualComponent())
	entity.AddComponent(NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255}))
	entity.AddComponent(NewAnimationComponent(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.EquipItem(entity, "weapon", "sword_001", 12345)
		sys.UnequipItem(entity, "weapon")
	}
}
