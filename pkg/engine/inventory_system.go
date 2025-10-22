// Package engine provides the inventory management system.
// This file implements InventorySystem which handles item management,
// equipment, and inventory operations for entities.
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// InventorySystem manages inventory and equipment operations.
type InventorySystem struct {
	world *World
}

// NewInventorySystem creates a new inventory system.
func NewInventorySystem(world *World) *InventorySystem {
	return &InventorySystem{
		world: world,
	}
}

// AddItemToInventory adds an item to an entity's inventory.
// Returns true if successful, false if inventory is full.
func (s *InventorySystem) AddItemToInventory(entityID uint64, itm *item.Item) (bool, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return false, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return false, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	return invComp.AddItem(itm), nil
}

// RemoveItemFromInventory removes an item from inventory by index.
func (s *InventorySystem) RemoveItemFromInventory(entityID uint64, index int) (*item.Item, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return nil, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return nil, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	itm := invComp.RemoveItem(index)
	if itm == nil {
		return nil, fmt.Errorf("invalid item index %d", index)
	}

	return itm, nil
}

// EquipItem equips an item from inventory to the appropriate slot.
// The item is removed from inventory and placed in equipment.
func (s *InventorySystem) EquipItem(entityID uint64, inventoryIndex int) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	comp2, ok := entity.GetComponent("equipment")
	if !ok {
		return fmt.Errorf("entity %d does not have equipment component", entityID)
	}
	equipComp, ok := comp2.(*EquipmentComponent)
	if !ok {
		return fmt.Errorf("entity %d equipment component has wrong type", entityID)
	}

	// Get item from inventory
	if inventoryIndex < 0 || inventoryIndex >= len(invComp.Items) {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}
	itm := invComp.Items[inventoryIndex]

	// Check if item can be equipped
	slot, canEquip := equipComp.GetSlotForItem(itm)
	if !canEquip {
		return fmt.Errorf("item %s cannot be equipped", itm.Name)
	}

	// Equip the item (this may return a previously equipped item)
	previousItem := equipComp.Equip(itm, slot)

	// Remove from inventory
	invComp.RemoveItem(inventoryIndex)

	// If there was a previously equipped item, add it to inventory
	if previousItem != nil {
		if !invComp.AddItem(previousItem) {
			// Inventory is full, re-equip the previous item and return error
			equipComp.Equip(previousItem, slot)
			invComp.Items = append(invComp.Items[:inventoryIndex],
				append([]*item.Item{itm}, invComp.Items[inventoryIndex:]...)...)
			return fmt.Errorf("cannot equip: inventory full for swapped item")
		}
	}

	// Update entity stats based on new equipment
	s.applyEquipmentStats(entityID)

	return nil
}

// UnequipItem removes an item from an equipment slot and adds it to inventory.
func (s *InventorySystem) UnequipItem(entityID uint64, slot EquipmentSlot) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	comp2, ok := entity.GetComponent("equipment")
	if !ok {
		return fmt.Errorf("entity %d does not have equipment component", entityID)
	}
	equipComp, ok := comp2.(*EquipmentComponent)
	if !ok {
		return fmt.Errorf("entity %d equipment component has wrong type", entityID)
	}

	// Unequip the item
	itm := equipComp.Unequip(slot)
	if itm == nil {
		return fmt.Errorf("no item equipped in slot %s", slot.String())
	}

	// Add to inventory
	if !invComp.AddItem(itm) {
		// Inventory is full, re-equip the item
		equipComp.Equip(itm, slot)
		return fmt.Errorf("cannot unequip: inventory full")
	}

	// Update entity stats
	s.applyEquipmentStats(entityID)

	return nil
}

// UseConsumable uses a consumable item from inventory.
// The item is removed from inventory after use.
func (s *InventorySystem) UseConsumable(entityID uint64, inventoryIndex int) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Get item from inventory
	if inventoryIndex < 0 || inventoryIndex >= len(invComp.Items) {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}
	itm := invComp.Items[inventoryIndex]

	// Check if item is consumable
	if !itm.IsConsumable() {
		return fmt.Errorf("item %s is not consumable", itm.Name)
	}

	// Apply consumable effects
	if err := s.applyConsumableEffects(entityID, itm); err != nil {
		return fmt.Errorf("failed to use consumable: %w", err)
	}

	// Remove from inventory
	invComp.RemoveItem(inventoryIndex)

	return nil
}

// applyConsumableEffects applies the effects of a consumable item to an entity.
func (s *InventorySystem) applyConsumableEffects(entityID uint64, itm *item.Item) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	// Get health component if it exists
	comp, hasHealth := entity.GetComponent("health")
	var healthComp *HealthComponent
	if hasHealth {
		healthComp, _ = comp.(*HealthComponent)
	}

	// Apply effects based on consumable type
	switch itm.ConsumableType {
	case item.ConsumablePotion:
		// Health potions restore health
		if healthComp != nil {
			// Restore health based on item value (simple implementation)
			healAmount := float64(itm.Stats.Value) / 10.0
			healthComp.Heal(healAmount)
		}

	case item.ConsumableScroll:
		// Scrolls might cast a spell or provide a buff
		// For now, just a placeholder
		// In a full implementation, this would trigger a spell effect

	case item.ConsumableFood:
		// Food restores health over time
		if healthComp != nil {
			healAmount := float64(itm.Stats.Value) / 20.0
			healthComp.Heal(healAmount)
		}

	case item.ConsumableBomb:
		// Bombs would deal area damage
		// This would require position and collision detection
		// Placeholder for now
	}

	return nil
}

// applyEquipmentStats updates an entity's stats based on equipped items.
func (s *InventorySystem) applyEquipmentStats(entityID uint64) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return
	}

	comp, ok := entity.GetComponent("equipment")
	if !ok {
		return
	}
	equipComp, _ := comp.(*EquipmentComponent)
	if equipComp == nil {
		return
	}

	// Get equipment stats
	equipStats := equipComp.GetStats()

	// Update stats component if it exists
	comp2, ok := entity.GetComponent("stats")
	if ok {
		if statsComp, ok := comp2.(*StatsComponent); ok {
			// Apply equipment defense to defense stat
			// Note: This is additive. The base stats are assumed to be set elsewhere.
			// A full implementation might want to track base vs. equipment bonuses separately
			statsComp.Defense = float64(equipStats.Defense)
		}
	}

	// Update attack component if it exists
	comp3, ok := entity.GetComponent("attack")
	if ok {
		if attackComp, ok := comp3.(*AttackComponent); ok {
			// Apply weapon damage
			weaponDamage := equipComp.GetWeaponDamage()
			if weaponDamage > 0 {
				attackComp.Damage = float64(weaponDamage)
			}

			// Apply weapon speed
			weaponSpeed := equipComp.GetWeaponSpeed()
			if weaponSpeed > 0 {
				attackComp.Cooldown = 1.0 / weaponSpeed
			}

			// Set damage type based on weapon
			mainHand := equipComp.GetEquipped(SlotMainHand)
			if mainHand != nil {
				// Map weapon types to damage types
				// This is simplified; a full system would have more nuanced mapping
				switch mainHand.WeaponType {
				case item.WeaponStaff:
					attackComp.DamageType = combat.DamageMagical
				default:
					attackComp.DamageType = combat.DamagePhysical
				}
			}
		}
	}
}

// DropItem removes an item from inventory and places it in the world.
// For now, this just removes the item. A full implementation would create
// a world entity for the dropped item.
func (s *InventorySystem) DropItem(entityID uint64, inventoryIndex int) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Remove item from inventory
	itm := invComp.RemoveItem(inventoryIndex)
	if itm == nil {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}

	// TODO: Create a world entity for the dropped item
	// This would require position information and a new entity type
	// For now, the item is simply removed

	return nil
}

// TransferItem moves an item from one entity's inventory to another's.
func (s *InventorySystem) TransferItem(fromEntityID, toEntityID uint64, inventoryIndex int) error {
	// Get source entity
	fromEntity, ok := s.world.GetEntity(fromEntityID)
	if !ok {
		return fmt.Errorf("source entity %d not found", fromEntityID)
	}

	comp, ok := fromEntity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("source entity %d does not have inventory component", fromEntityID)
	}
	fromInv, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("source entity %d inventory component has wrong type", fromEntityID)
	}

	// Get destination entity
	toEntity, ok := s.world.GetEntity(toEntityID)
	if !ok {
		return fmt.Errorf("destination entity %d not found", toEntityID)
	}

	comp2, ok := toEntity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("destination entity %d does not have inventory component", toEntityID)
	}
	toInv, ok := comp2.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("destination entity %d inventory component has wrong type", toEntityID)
	}

	// Get item from source inventory
	if inventoryIndex < 0 || inventoryIndex >= len(fromInv.Items) {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}
	itm := fromInv.Items[inventoryIndex]

	// Check if destination can accept the item
	if !toInv.CanAddItem(itm) {
		return fmt.Errorf("destination inventory cannot accept item")
	}

	// Transfer the item
	fromInv.RemoveItem(inventoryIndex)
	toInv.AddItem(itm)

	return nil
}

// GetInventoryValue returns the total value of all items in an entity's inventory.
func (s *InventorySystem) GetInventoryValue(entityID uint64) (int, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return 0, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return 0, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return 0, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	totalValue := invComp.Gold
	for _, itm := range invComp.Items {
		totalValue += itm.GetValue()
	}

	return totalValue, nil
}

// SortInventoryByValue sorts inventory items by value (descending).
func (s *InventorySystem) SortInventoryByValue(entityID uint64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Simple bubble sort (good enough for small inventories)
	items := invComp.Items
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].GetValue() < items[j+1].GetValue() {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	return nil
}

// SortInventoryByWeight sorts inventory items by weight (ascending).
func (s *InventorySystem) SortInventoryByWeight(entityID uint64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Simple bubble sort
	items := invComp.Items
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].Stats.Weight > items[j+1].Stats.Weight {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	return nil
}

// SortInventoryByType sorts inventory items by type.
func (s *InventorySystem) SortInventoryByType(entityID uint64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Simple bubble sort by type
	items := invComp.Items
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].Type > items[j+1].Type {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	return nil
}

// Update implements the System interface.
// InventorySystem doesn't need per-frame updates, so this is a no-op.
func (s *InventorySystem) Update(entities []*Entity, deltaTime float64) {
	// InventorySystem is event-driven (AddItem, RemoveItem, etc.), not frame-driven
	// No per-frame updates needed
}
