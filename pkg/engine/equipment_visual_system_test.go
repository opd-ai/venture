package engine

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
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
	sys.Update(entities, 0.016)

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
	sys.Update(entities, 0.016)

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
	sys.Update(entities, 0.016)

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
	sys.Update(entities, 0.016)

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
	sys.Update(entities, 0.016)

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
		sys.Update(entities, 0.016)

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

// TestEquipmentVisualSystem_AccessorySync tests accessory slot syncing.
func TestEquipmentVisualSystem_AccessorySync(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	// Add equipment component with 3 accessory slots
	equipComp := NewEquipmentComponent()
	entity.AddComponent(equipComp)

	// Add equipment visual component
	equipVisualComp := NewEquipmentVisualComponent()
	entity.AddComponent(equipVisualComp)

	// Add sprite and animation components
	spriteComp := NewSpriteComponent(28, 28, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	entity.AddComponent(spriteComp)
	animComp := NewAnimationComponent(12345)
	entity.AddComponent(animComp)

	// Create test accessories
	ring1 := createTestAccessory("ring_001", 1001)
	ring2 := createTestAccessory("ring_002", 1002)
	amulet := createTestAccessory("amulet_001", 1003)

	// Test 1: Equip single accessory
	equipComp.Equip(ring1, SlotAccessory1)
	sys.syncEquipmentChanges(entity)

	if len(equipVisualComp.AccessoryIDs) != 1 {
		t.Errorf("Expected 1 accessory, got %d", len(equipVisualComp.AccessoryIDs))
	}
	if equipVisualComp.AccessoryIDs[0] != "ring_001" {
		t.Errorf("Expected ring_001, got %s", equipVisualComp.AccessoryIDs[0])
	}

	// Test 2: Equip second accessory
	equipComp.Equip(ring2, SlotAccessory2)
	sys.syncEquipmentChanges(entity)

	if len(equipVisualComp.AccessoryIDs) != 2 {
		t.Errorf("Expected 2 accessories, got %d", len(equipVisualComp.AccessoryIDs))
	}
	if equipVisualComp.AccessoryIDs[1] != "ring_002" {
		t.Errorf("Expected ring_002, got %s", equipVisualComp.AccessoryIDs[1])
	}

	// Test 3: Equip third accessory (all slots filled)
	equipComp.Equip(amulet, SlotAccessory3)
	sys.syncEquipmentChanges(entity)

	if len(equipVisualComp.AccessoryIDs) != 3 {
		t.Errorf("Expected 3 accessories, got %d", len(equipVisualComp.AccessoryIDs))
	}
	if equipVisualComp.AccessoryIDs[2] != "amulet_001" {
		t.Errorf("Expected amulet_001, got %s", equipVisualComp.AccessoryIDs[2])
	}

	// Test 4: Unequip middle accessory
	equipComp.Unequip(SlotAccessory2)
	sys.syncEquipmentChanges(entity)

	if len(equipVisualComp.AccessoryIDs) != 2 {
		t.Errorf("Expected 2 accessories after unequip, got %d", len(equipVisualComp.AccessoryIDs))
	}
	// Should now have ring_001 and amulet_001 (ring_002 removed)
	if equipVisualComp.AccessoryIDs[0] != "ring_001" || equipVisualComp.AccessoryIDs[1] != "amulet_001" {
		t.Errorf("Unexpected accessory order after unequip: %v", equipVisualComp.AccessoryIDs)
	}

	// Test 5: Unequip all accessories
	equipComp.Unequip(SlotAccessory1)
	equipComp.Unequip(SlotAccessory3)
	sys.syncEquipmentChanges(entity)

	if len(equipVisualComp.AccessoryIDs) != 0 {
		t.Errorf("Expected 0 accessories after unequip all, got %d", len(equipVisualComp.AccessoryIDs))
	}
}

// TestEquipmentVisualSystem_AccessorySyncNoChanges tests that sync doesn't trigger on unchanged accessories.
func TestEquipmentVisualSystem_AccessorySyncNoChanges(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	equipComp := NewEquipmentComponent()
	entity.AddComponent(equipComp)

	equipVisualComp := NewEquipmentVisualComponent()
	entity.AddComponent(equipVisualComp)

	// Equip an accessory
	ring := createTestAccessory("ring_001", 1001)
	equipComp.Equip(ring, SlotAccessory1)
	sys.syncEquipmentChanges(entity)

	// Mark as clean
	equipVisualComp.MarkClean()

	// Sync again without changes - should remain clean
	sys.syncEquipmentChanges(entity)

	if equipVisualComp.Dirty {
		t.Error("Component should remain clean when accessories haven't changed")
	}
}

// TestEquipmentVisualSystem_AccessorySeedTracking tests that accessory seeds are tracked correctly.
func TestEquipmentVisualSystem_AccessorySeedTracking(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(1)

	equipComp := NewEquipmentComponent()
	entity.AddComponent(equipComp)

	equipVisualComp := NewEquipmentVisualComponent()
	entity.AddComponent(equipVisualComp)

	// Equip accessories with different seeds
	ring1 := createTestAccessory("ring_001", 5001)
	ring2 := createTestAccessory("ring_002", 5002)

	equipComp.Equip(ring1, SlotAccessory1)
	equipComp.Equip(ring2, SlotAccessory2)
	sys.syncEquipmentChanges(entity)

	// Verify seeds are tracked
	if len(equipVisualComp.AccessorySeeds) != 2 {
		t.Errorf("Expected 2 accessory seeds, got %d", len(equipVisualComp.AccessorySeeds))
	}
	if equipVisualComp.AccessorySeeds[0] != 5001 {
		t.Errorf("Expected seed 5001, got %d", equipVisualComp.AccessorySeeds[0])
	}
	if equipVisualComp.AccessorySeeds[1] != 5002 {
		t.Errorf("Expected seed 5002, got %d", equipVisualComp.AccessorySeeds[1])
	}
}

// createTestAccessory creates a test accessory item for testing.
func createTestAccessory(id string, seed int64) *item.Item {
	return &item.Item{
		ID:   id,
		Name: id,
		Type: item.TypeAccessory,
		Seed: seed,
		Stats: item.Stats{
			Weight: 0.1,
		},
	}
}

// TestResolveEntityType_Default tests that a bare entity resolves to humanoid.
func TestResolveEntityType_Default(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(100)
	got := sys.resolveEntityType(entity)
	if got != "humanoid" {
		t.Errorf("resolveEntityType = %q, want %q", got, "humanoid")
	}
}

// TestResolveEntityType_WithCreatureVisual tests nonhumanoid entity type resolution.
func TestResolveEntityType_WithCreatureVisual(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(101)
	cv := &CreatureVisualComponent{Form: FormQuadruped}
	entity.AddComponent(cv)

	got := sys.resolveEntityType(entity)
	if got != string(FormQuadruped) {
		t.Errorf("resolveEntityType = %q, want %q", got, FormQuadruped)
	}
}

// TestResolveEntityType_WithNpcRole tests NPC role-based entity type resolution.
func TestResolveEntityType_WithNpcRole(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(102)
	npc := &NpcRoleVisualComponent{Role: "merchant"}
	entity.AddComponent(npc)

	got := sys.resolveEntityType(entity)
	if got != "merchant" {
		t.Errorf("resolveEntityType = %q, want %q", got, "merchant")
	}
}

// TestResolveEntityFacing_Default tests default facing direction.
func TestResolveEntityFacing_Default(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(103)
	got := sys.resolveEntityFacing(entity)
	if got != "down" {
		t.Errorf("resolveEntityFacing = %q, want %q", got, "down")
	}
}

// TestResolveEntityFacing_WithAnimation tests facing from animation component.
func TestResolveEntityFacing_WithAnimation(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	entity := NewEntity(104)
	anim := &AnimationComponent{LastFacing: "left"}
	entity.AddComponent(anim)

	got := sys.resolveEntityFacing(entity)
	if got != "left" {
		t.Errorf("resolveEntityFacing = %q, want %q", got, "left")
	}
}
