package engine

import "github.com/opd-ai/venture/pkg/procgen/item"

// InventorySystem manages inventory and equipment operations.
type InventorySystem struct {
	// Callbacks
	OnItemAdded    func(entityID uint64, itm *item.Item)
	OnItemRemoved  func(entityID uint64, itm *item.Item)
	OnItemEquipped func(entityID uint64, itm *item.Item, slot EquipmentSlot)
	OnItemUnequipped func(entityID uint64, itm *item.Item, slot EquipmentSlot)
}

// NewInventorySystem creates a new inventory system.
func NewInventorySystem() *InventorySystem {
	return &InventorySystem{}
}

// Update implements the System interface (inventory doesn't need periodic updates).
func (s *InventorySystem) Update(world *World, deltaTime float64) {
	// Inventory system is event-driven, no periodic updates needed
}

// AddItemToInventory adds an item to an entity's inventory.
// Returns true if successful, false if inventory is full or would exceed weight limit.
func (s *InventorySystem) AddItemToInventory(world *World, entityID uint64, itm *item.Item) bool {
	entity, ok := world.GetEntity(entityID)
	if !ok {
		return false
	}
	
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return false
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return false
	}
	
	if inv.AddItem(itm) {
		if s.OnItemAdded != nil {
			s.OnItemAdded(entityID, itm)
		}
		return true
	}
	return false
}

// RemoveItemFromInventory removes an item at the given index from inventory.
// Returns the removed item, or nil if index is invalid.
func (s *InventorySystem) RemoveItemFromInventory(world *World, entityID uint64, index int) *item.Item {
	entity, ok := world.GetEntity(entityID)
	if !ok {
		return nil
	}
	
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return nil
	}
	
	itm := inv.RemoveItem(index)
	if itm != nil && s.OnItemRemoved != nil {
		s.OnItemRemoved(entityID, itm)
	}
	return itm
}

// EquipItem equips an item from inventory to the appropriate equipment slot.
// Returns true if successful, false if item cannot be equipped or entity doesn't have equipment component.
func (s *InventorySystem) EquipItem(world *World, entityID uint64, inventoryIndex int) bool {
	entity, ok := world.GetEntity(entityID)
	if !ok {
		return false
	}
	
	invComp, hasInv := entity.GetComponent("inventory")
	equipComp, hasEquip := entity.GetComponent("equipment")
	
	if !hasInv || !hasEquip {
		return false
	}
	
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return false
	}
	equip, ok := equipComp.(*EquipmentComponent)
	if !ok {
		return false
	}
	
	// Validate inventory index
	if inventoryIndex < 0 || inventoryIndex >= len(inv.Items) {
		return false
	}
	
	itm := inv.Items[inventoryIndex]
	
	// Check if item is equippable
	if !itm.IsEquippable() {
		return false
	}
	
	// Remove from inventory
	inv.RemoveItem(inventoryIndex)
	
	// Equip item (this may return a previously equipped item)
	previousItem := equip.Equip(itm)
	
	// If there was a previously equipped item, add it back to inventory
	if previousItem != nil {
		// Try to add back to inventory
		if !inv.AddItem(previousItem) {
			// If inventory is full, drop the item (implementation-specific)
			// For now, we force it back into the slot
			equip.Equip(itm) // Re-equip current item
			inv.Items = append(inv.Items[:inventoryIndex], append([]*item.Item{itm}, inv.Items[inventoryIndex:]...)...)
			return false
		}
	}
	
	// Update stats based on equipment
	s.UpdateStatsFromEquipment(entity)
	
	if s.OnItemEquipped != nil {
		slot := GetSlotForItem(itm)
		s.OnItemEquipped(entityID, itm, slot)
	}
	
	return true
}

// UnequipItem unequips an item from a slot and adds it to inventory.
// Returns true if successful, false if slot is empty or inventory is full.
func (s *InventorySystem) UnequipItem(world *World, entityID uint64, slot EquipmentSlot) bool {
	entity, ok := world.GetEntity(entityID)
	if !ok {
		return false
	}
	
	invComp, hasInv := entity.GetComponent("inventory")
	equipComp, hasEquip := entity.GetComponent("equipment")
	
	if !hasInv || !hasEquip {
		return false
	}
	
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return false
	}
	equip, ok := equipComp.(*EquipmentComponent)
	if !ok {
		return false
	}
	
	// Get the equipped item
	itm := equip.GetEquipped(slot)
	if itm == nil {
		return false // Slot is empty
	}
	
	// Check if there's space in inventory
	if !inv.CanAddItem(itm) {
		return false // Inventory full or would exceed weight
	}
	
	// Unequip and add to inventory
	equip.Unequip(slot)
	inv.AddItem(itm)
	
	// Update stats based on equipment
	s.UpdateStatsFromEquipment(entity)
	
	if s.OnItemUnequipped != nil {
		s.OnItemUnequipped(entityID, itm, slot)
	}
	
	return true
}

// UseConsumable uses a consumable item from inventory.
// Returns true if successful, false if item is not consumable or doesn't exist.
func (s *InventorySystem) UseConsumable(world *World, entityID uint64, inventoryIndex int) bool {
	entity, ok := world.GetEntity(entityID)
	if !ok {
		return false
	}
	
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return false
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return false
	}
	
	// Validate inventory index
	if inventoryIndex < 0 || inventoryIndex >= len(inv.Items) {
		return false
	}
	
	itm := inv.Items[inventoryIndex]
	
	// Check if item is consumable
	if !itm.IsConsumable() {
		return false
	}
	
	// Apply consumable effects based on type
	healthComp, hasHealth := entity.GetComponent("health")
	var health *HealthComponent
	if hasHealth {
		health, _ = healthComp.(*HealthComponent)
	}
	
	switch itm.ConsumableType {
	case item.ConsumablePotion:
		// Healing potions restore health
		if hasHealth {
			// Use item value as healing amount (scaled)
			healAmount := float64(itm.Stats.Value) / 10.0
			health.Heal(healAmount)
		}
	case item.ConsumableFood:
		// Food provides small healing over time (instant for now)
		if hasHealth {
			healAmount := float64(itm.Stats.Value) / 20.0
			health.Heal(healAmount)
		}
	// Other consumable types can be implemented as needed
	default:
		// Generic consumable effect
		if hasHealth {
			healAmount := float64(itm.Stats.Value) / 15.0
			health.Heal(healAmount)
		}
	}
	
	// Remove the consumed item
	inv.RemoveItem(inventoryIndex)
	
	if s.OnItemRemoved != nil {
		s.OnItemRemoved(entityID, itm)
	}
	
	return true
}

// UpdateStatsFromEquipment updates an entity's stats based on equipped items.
func (s *InventorySystem) UpdateStatsFromEquipment(entity *Entity) {
	equipComp, hasEquip := entity.GetComponent("equipment")
	statsComp, hasStats := entity.GetComponent("stats")
	
	if !hasEquip || !hasStats {
		return
	}
	
	equip, ok := equipComp.(*EquipmentComponent)
	if !ok {
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return
	}
	
	// Calculate total bonuses from equipment
	totalDamage := equip.GetTotalDamage()
	totalDefense := equip.GetTotalDefense()
	
	// Apply bonuses to stats
	// Note: This is a simple implementation. In a real game, you might want
	// to track base stats separately from equipment bonuses
	stats.Attack = 10.0 + float64(totalDamage)
	stats.Defense = 5.0 + float64(totalDefense)
	
	// Update attack component if present
	attackComp, hasAttack := entity.GetComponent("attack")
	if hasAttack {
		attack, ok := attackComp.(*AttackComponent)
		if ok {
			attack.Damage = stats.Attack
		}
	}
}

// DropItem removes an item from inventory (simulates dropping on ground).
// In a full implementation, this would create an item entity in the world.
func (s *InventorySystem) DropItem(world *World, entityID uint64, inventoryIndex int) *item.Item {
	entity, ok := world.GetEntity(entityID)
	if !ok {
		return nil
	}
	
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return nil
	}
	
	itm := inv.RemoveItem(inventoryIndex)
	if itm != nil && s.OnItemRemoved != nil {
		s.OnItemRemoved(entityID, itm)
	}
	return itm
}

// GetInventoryInfo returns a summary of an entity's inventory.
func (s *InventorySystem) GetInventoryInfo(world *World, entityID uint64) (itemCount, maxCapacity int, currentWeight, maxWeight float64, ok bool) {
	entity, exists := world.GetEntity(entityID)
	if !exists {
		return 0, 0, 0, 0, false
	}
	
	comp, hasInv := entity.GetComponent("inventory")
	if !hasInv {
		return 0, 0, 0, 0, false
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return 0, 0, 0, 0, false
	}
	
	return len(inv.Items), inv.MaxCapacity, inv.GetCurrentWeight(), inv.MaxWeight, true
}

// GetEquipmentInfo returns information about equipped items.
func (s *InventorySystem) GetEquipmentInfo(world *World, entityID uint64) (totalDamage, totalDefense int, ok bool) {
	entity, exists := world.GetEntity(entityID)
	if !exists {
		return 0, 0, false
	}
	
	comp, hasEquip := entity.GetComponent("equipment")
	if !hasEquip {
		return 0, 0, false
	}
	equip, ok := comp.(*EquipmentComponent)
	if !ok {
		return 0, 0, false
	}
	
	return equip.GetTotalDamage(), equip.GetTotalDefense(), true
}
