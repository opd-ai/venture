package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestLootDropFromInventory tests that items are spawned when an entity dies
// Priority 1.4: Loot Drop System
func TestLootDropFromInventory(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	droppedItemCount := 0
	var droppedEntity *Entity
	callbackExecuted := false

	// Set death callback that tracks drops
	combatSystem.SetDeathCallback(func(entity *Entity) {
		// Only process once (callback called every frame while dead)
		if entity.HasComponent("dead") {
			return
		}

		callbackExecuted = true

		// Add dead component
		entity.AddComponent(NewDeadComponent(0.0))

		// Check if entity has inventory
		if invComp, hasInv := entity.GetComponent("inventory"); hasInv {
			inventory := invComp.(*InventoryComponent)
			droppedItemCount = len(inventory.Items)
			droppedEntity = entity

			// Spawn items (simplified version without scatter)
			posComp, _ := entity.GetComponent("position")
			pos := posComp.(*PositionComponent)

			deadComp, _ := entity.GetComponent("dead")
			dead := deadComp.(*DeadComponent)

			for _, itm := range inventory.Items {
				itemEntity := SpawnItemInWorld(world, itm, pos.X, pos.Y)
				if itemEntity != nil {
					dead.AddDroppedItem(itemEntity.ID)
				}
			}

			inventory.Clear()
		}
	})

	world.AddSystem(combatSystem)

	// Create entity with inventory
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	inventory := NewInventoryComponent(10, 100.0)

	// Add test items to inventory
	testItem1 := &item.Item{
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Damage: 10,
			Weight: 5.0,
		},
	}
	testItem2 := &item.Item{
		Name: "Test Potion",
		Type: item.TypeConsumable,
		Stats: item.Stats{
			Weight: 1.0,
		},
	}

	inventory.AddItem(testItem1)
	inventory.AddItem(testItem2)
	entity.AddComponent(inventory)

	world.Update(0)

	// Kill entity
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(100)

	// Trigger death callback
	world.Update(0.1)

	// Process entity additions to world
	world.Update(0)

	// Verify callback was executed
	if !callbackExecuted {
		t.Fatal("death callback should have been executed")
	}

	// Verify death component added
	if !entity.HasComponent("dead") {
		t.Fatal("entity should have dead component after death")
	}

	// Verify items were tracked
	if droppedItemCount != 2 {
		t.Errorf("expected 2 items to be dropped, got %d", droppedItemCount)
	}

	// Verify inventory was cleared
	invComp, _ := droppedEntity.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if len(inv.Items) != 0 {
		t.Errorf("inventory should be empty after death, got %d items", len(inv.Items))
	}

	// Verify dropped items tracked in dead component
	deadComp, _ := entity.GetComponent("dead")
	dead := deadComp.(*DeadComponent)
	if len(dead.DroppedItems) != 2 {
		t.Errorf("expected 2 dropped item IDs tracked, got %d", len(dead.DroppedItems))
	}

	// Verify item entities were created in world
	itemEntities := 0
	for _, e := range world.GetEntities() {
		if e.HasComponent("item_entity") {
			itemEntities++
		}
	}
	if itemEntities != 2 {
		t.Errorf("expected 2 item entities in world, got %d", itemEntities)
	}
}

// TestLootDropFromEquipment tests that equipped items are dropped on death
func TestLootDropFromEquipment(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	var droppedEquipCount int

	combatSystem.SetDeathCallback(func(entity *Entity) {
		// Only process once
		if entity.HasComponent("dead") {
			return
		}

		entity.AddComponent(NewDeadComponent(0.0))

		if equipComp, hasEquip := entity.GetComponent("equipment"); hasEquip {
			equipment := equipComp.(*EquipmentComponent)
			equippedItems := equipment.UnequipAll()
			droppedEquipCount = len(equippedItems)

			posComp, _ := entity.GetComponent("position")
			pos := posComp.(*PositionComponent)

			deadComp, _ := entity.GetComponent("dead")
			dead := deadComp.(*DeadComponent)

			for _, itm := range equippedItems {
				itemEntity := SpawnItemInWorld(world, itm, pos.X, pos.Y)
				if itemEntity != nil {
					dead.AddDroppedItem(itemEntity.ID)
				}
			}
		}
	})

	world.AddSystem(combatSystem)

	// Create entity with equipped items
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	equipment := NewEquipmentComponent()

	weapon := &item.Item{
		Name:       "Equipped Sword",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Stats:      item.Stats{Damage: 20, Weight: 10.0},
	}
	armor := &item.Item{
		Name:      "Equipped Helmet",
		Type:      item.TypeArmor,
		ArmorType: item.ArmorHelmet,
		Stats:     item.Stats{Defense: 5, Weight: 5.0},
	}

	equipment.Equip(weapon, SlotMainHand)
	equipment.Equip(armor, SlotHead)
	entity.AddComponent(equipment)

	world.Update(0)

	// Kill entity
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(100)

	world.Update(0.1)

	// Process entity additions
	world.Update(0)

	// Verify equipped items were dropped
	if droppedEquipCount != 2 {
		t.Errorf("expected 2 equipped items to be dropped, got %d", droppedEquipCount)
	}

	// Verify equipment slots are empty
	equipComp, _ := entity.GetComponent("equipment")
	equip := equipComp.(*EquipmentComponent)
	if len(equip.Slots) != 0 {
		t.Errorf("equipment slots should be empty after death, got %d", len(equip.Slots))
	}

	// Verify item entities created
	itemEntities := 0
	for _, e := range world.GetEntities() {
		if e.HasComponent("item_entity") {
			itemEntities++
		}
	}
	if itemEntities != 2 {
		t.Errorf("expected 2 item entities in world, got %d", itemEntities)
	}
}

// TestLootDropWithPhysics tests that dropped items have velocity and friction
func TestLootDropWithPhysics(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	var droppedItemEntity *Entity

	combatSystem.SetDeathCallback(func(entity *Entity) {
		// Only process once
		if entity.HasComponent("dead") {
			return
		}

		entity.AddComponent(NewDeadComponent(0.0))

		if invComp, hasInv := entity.GetComponent("inventory"); hasInv {
			inventory := invComp.(*InventoryComponent)
			posComp, _ := entity.GetComponent("position")
			pos := posComp.(*PositionComponent)

			if len(inventory.Items) > 0 {
				// Spawn first item with physics
				itemEntity := SpawnItemInWorld(world, inventory.Items[0], pos.X, pos.Y)
				if itemEntity != nil {
					// Add velocity
					itemEntity.AddComponent(&VelocityComponent{VX: 50, VY: 50})
					// Add friction
					itemEntity.AddComponent(NewFrictionComponent(0.12))
					droppedItemEntity = itemEntity
				}
			}

			inventory.Clear()
		}
	})

	world.AddSystem(combatSystem)

	// Create entity with one item
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	inventory := NewInventoryComponent(10, 100.0)
	testItem := &item.Item{
		Name:  "Physics Test Item",
		Type:  item.TypeWeapon,
		Stats: item.Stats{Weight: 5.0},
	}
	inventory.AddItem(testItem)
	entity.AddComponent(inventory)

	world.Update(0)

	// Kill entity
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(100)

	world.Update(0.1)

	// Verify item has velocity component
	if !droppedItemEntity.HasComponent("velocity") {
		t.Fatal("dropped item should have velocity component")
	}

	velComp, _ := droppedItemEntity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 || vel.VY == 0 {
		t.Error("dropped item should have non-zero initial velocity")
	}

	// Verify item has friction component
	if !droppedItemEntity.HasComponent("friction") {
		t.Fatal("dropped item should have friction component")
	}

	frictionComp, _ := droppedItemEntity.GetComponent("friction")
	friction := frictionComp.(*FrictionComponent)
	if friction.Coefficient <= 0 {
		t.Error("friction coefficient should be positive")
	}
}

// TestLootDropEmptyInventory tests that entities without items don't crash
func TestLootDropEmptyInventory(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	callbackCalled := false

	combatSystem.SetDeathCallback(func(entity *Entity) {
		// Only process once
		if entity.HasComponent("dead") {
			return
		}

		callbackCalled = true
		entity.AddComponent(NewDeadComponent(0.0))

		// Try to process empty inventory - should not crash
		if invComp, hasInv := entity.GetComponent("inventory"); hasInv {
			inventory := invComp.(*InventoryComponent)

			posComp, _ := entity.GetComponent("position")
			pos := posComp.(*PositionComponent)

			for _, itm := range inventory.Items {
				SpawnItemInWorld(world, itm, pos.X, pos.Y)
			}

			inventory.Clear()
		}
	})

	world.AddSystem(combatSystem)

	// Create entity with empty inventory
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(NewInventoryComponent(10, 100.0)) // Empty

	world.Update(0)

	// Kill entity - should not crash
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(100)

	world.Update(0.1)

	if !callbackCalled {
		t.Error("death callback should have been called")
	}

	// Verify no item entities created
	itemEntities := 0
	for _, e := range world.GetEntities() {
		if e.HasComponent("item_entity") {
			itemEntities++
		}
	}
	if itemEntities != 0 {
		t.Errorf("expected 0 item entities from empty inventory, got %d", itemEntities)
	}
}

// TestLootDropNoInventory tests entities without inventory component
func TestLootDropNoInventory(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	callbackCalled := false

	combatSystem.SetDeathCallback(func(entity *Entity) {
		// Only process once
		if entity.HasComponent("dead") {
			return
		}

		callbackCalled = true
		entity.AddComponent(NewDeadComponent(0.0))

		// Should handle missing inventory gracefully
		if invComp, hasInv := entity.GetComponent("inventory"); hasInv {
			inventory := invComp.(*InventoryComponent)
			inventory.Clear() // Won't be called since no inventory
		}
	})

	world.AddSystem(combatSystem)

	// Create entity WITHOUT inventory
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 10, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No inventory component

	world.Update(0)

	// Kill entity - should not crash
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(100)

	world.Update(0.1)

	if !callbackCalled {
		t.Error("death callback should have been called")
	}

	// Should complete without errors
}
