// Package engine provides inventory and equipment components.
// This file defines components for item storage, equipment slots, and
// inventory management used by the inventory system.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// InventoryComponent manages an entity's item collection.
type InventoryComponent struct {
	// Items stored in the inventory
	Items []*item.Item

	// MaxItems is the maximum number of items that can be carried
	MaxItems int

	// MaxWeight is the maximum weight that can be carried
	MaxWeight float64

	// Gold/currency amount
	Gold int
}

// Type returns the component type identifier.
func (i *InventoryComponent) Type() string {
	return "inventory"
}

// NewInventoryComponent creates a new inventory with default limits.
func NewInventoryComponent(maxItems int, maxWeight float64) *InventoryComponent {
	return &InventoryComponent{
		Items:     make([]*item.Item, 0, maxItems),
		MaxItems:  maxItems,
		MaxWeight: maxWeight,
		Gold:      0,
	}
}

// GetCurrentWeight returns the total weight of all items in inventory.
func (i *InventoryComponent) GetCurrentWeight() float64 {
	totalWeight := 0.0
	for _, itm := range i.Items {
		totalWeight += itm.Stats.Weight
	}
	return totalWeight
}

// CanAddItem checks if an item can be added to inventory.
func (i *InventoryComponent) CanAddItem(itm *item.Item) bool {
	// Check item count limit
	if len(i.Items) >= i.MaxItems {
		return false
	}

	// Check weight limit
	if i.GetCurrentWeight()+itm.Stats.Weight > i.MaxWeight {
		return false
	}

	return true
}

// AddItem adds an item to the inventory if possible.
// Returns true if successful, false if inventory is full or weight exceeded.
func (i *InventoryComponent) AddItem(itm *item.Item) bool {
	if !i.CanAddItem(itm) {
		return false
	}

	i.Items = append(i.Items, itm)
	return true
}

// RemoveItem removes an item from inventory by index.
// Returns the removed item or nil if index is invalid.
func (i *InventoryComponent) RemoveItem(index int) *item.Item {
	if index < 0 || index >= len(i.Items) {
		return nil
	}

	itm := i.Items[index]
	i.Items = append(i.Items[:index], i.Items[index+1:]...)
	return itm
}

// RemoveItemByReference removes a specific item instance from inventory.
// Returns true if the item was found and removed.
func (i *InventoryComponent) RemoveItemByReference(itm *item.Item) bool {
	for idx, invItem := range i.Items {
		if invItem == itm {
			i.RemoveItem(idx)
			return true
		}
	}
	return false
}

// FindItem searches for an item by name.
// Returns the first matching item or nil if not found.
func (i *InventoryComponent) FindItem(name string) *item.Item {
	for _, itm := range i.Items {
		if itm.Name == name {
			return itm
		}
	}
	return nil
}

// GetItemCount returns the number of items in inventory.
func (i *InventoryComponent) GetItemCount() int {
	return len(i.Items)
}

// IsFull returns true if inventory cannot accept more items.
func (i *InventoryComponent) IsFull() bool {
	return len(i.Items) >= i.MaxItems || i.GetCurrentWeight() >= i.MaxWeight
}

// Clear removes all items from inventory.
func (i *InventoryComponent) Clear() {
	i.Items = i.Items[:0]
}

// EquipmentSlot represents a slot where equipment can be placed.
type EquipmentSlot int

const (
	// SlotMainHand is the primary weapon slot
	SlotMainHand EquipmentSlot = iota
	// SlotOffHand is the secondary weapon or shield slot
	SlotOffHand
	// SlotHead is the helmet slot
	SlotHead
	// SlotChest is the body armor slot
	SlotChest
	// SlotLegs is the leg armor slot
	SlotLegs
	// SlotBoots is the footwear slot
	SlotBoots
	// SlotGloves is the hand armor slot
	SlotGloves
	// SlotAccessory1 is the first accessory slot
	SlotAccessory1
	// SlotAccessory2 is the second accessory slot
	SlotAccessory2
	// SlotAccessory3 is the third accessory slot
	SlotAccessory3
)

// String returns the string representation of an equipment slot.
func (s EquipmentSlot) String() string {
	switch s {
	case SlotMainHand:
		return "main_hand"
	case SlotOffHand:
		return "off_hand"
	case SlotHead:
		return "head"
	case SlotChest:
		return "chest"
	case SlotLegs:
		return "legs"
	case SlotBoots:
		return "boots"
	case SlotGloves:
		return "gloves"
	case SlotAccessory1:
		return "accessory_1"
	case SlotAccessory2:
		return "accessory_2"
	case SlotAccessory3:
		return "accessory_3"
	default:
		return "unknown"
	}
}

// EquipmentComponent manages equipped items and their effects.
type EquipmentComponent struct {
	// Slots maps equipment slots to items
	Slots map[EquipmentSlot]*item.Item

	// CachedStats stores the total bonuses from all equipped items
	CachedStats item.Stats

	// StatsDirty indicates if cached stats need recalculation
	StatsDirty bool
}

// Type returns the component type identifier.
func (e *EquipmentComponent) Type() string {
	return "equipment"
}

// NewEquipmentComponent creates a new equipment component.
func NewEquipmentComponent() *EquipmentComponent {
	return &EquipmentComponent{
		Slots:      make(map[EquipmentSlot]*item.Item),
		StatsDirty: true,
	}
}

// CanEquip checks if an item can be equipped in a slot.
func (e *EquipmentComponent) CanEquip(itm *item.Item, slot EquipmentSlot) bool {
	if !itm.IsEquippable() {
		return false
	}

	// Check if item type matches slot
	switch slot {
	case SlotMainHand, SlotOffHand:
		return itm.Type == item.TypeWeapon
	case SlotHead:
		return itm.Type == item.TypeArmor && itm.ArmorType == item.ArmorHelmet
	case SlotChest:
		return itm.Type == item.TypeArmor && itm.ArmorType == item.ArmorChest
	case SlotLegs:
		return itm.Type == item.TypeArmor && itm.ArmorType == item.ArmorLegs
	case SlotBoots:
		return itm.Type == item.TypeArmor && itm.ArmorType == item.ArmorBoots
	case SlotGloves:
		return itm.Type == item.TypeArmor && itm.ArmorType == item.ArmorGloves
	case SlotAccessory1, SlotAccessory2, SlotAccessory3:
		return itm.Type == item.TypeAccessory
	default:
		return false
	}
}

// Equip places an item in the specified slot.
// Returns the previously equipped item (if any).
func (e *EquipmentComponent) Equip(itm *item.Item, slot EquipmentSlot) *item.Item {
	if !e.CanEquip(itm, slot) {
		return nil
	}

	// Get previously equipped item
	previousItem := e.Slots[slot]

	// Equip new item
	e.Slots[slot] = itm
	e.StatsDirty = true

	return previousItem
}

// Unequip removes an item from the specified slot.
// Returns the unequipped item or nil if slot is empty.
func (e *EquipmentComponent) Unequip(slot EquipmentSlot) *item.Item {
	itm, exists := e.Slots[slot]
	if !exists {
		return nil
	}

	delete(e.Slots, slot)
	e.StatsDirty = true

	return itm
}

// GetEquipped returns the item in the specified slot.
func (e *EquipmentComponent) GetEquipped(slot EquipmentSlot) *item.Item {
	return e.Slots[slot]
}

// IsEquipped checks if a specific item is currently equipped.
func (e *EquipmentComponent) IsEquipped(itm *item.Item) bool {
	for _, equipped := range e.Slots {
		if equipped == itm {
			return true
		}
	}
	return false
}

// GetSlotForItem determines the appropriate slot for an item.
// Returns the slot and true if valid, or 0 and false if item can't be equipped.
func (e *EquipmentComponent) GetSlotForItem(itm *item.Item) (EquipmentSlot, bool) {
	if !itm.IsEquippable() {
		return 0, false
	}

	switch itm.Type {
	case item.TypeWeapon:
		return SlotMainHand, true
	case item.TypeArmor:
		switch itm.ArmorType {
		case item.ArmorHelmet:
			return SlotHead, true
		case item.ArmorChest:
			return SlotChest, true
		case item.ArmorLegs:
			return SlotLegs, true
		case item.ArmorBoots:
			return SlotBoots, true
		case item.ArmorGloves:
			return SlotGloves, true
		case item.ArmorShield:
			return SlotOffHand, true
		}
	case item.TypeAccessory:
		// Find first empty accessory slot
		if e.Slots[SlotAccessory1] == nil {
			return SlotAccessory1, true
		}
		if e.Slots[SlotAccessory2] == nil {
			return SlotAccessory2, true
		}
		if e.Slots[SlotAccessory3] == nil {
			return SlotAccessory3, true
		}
		// All accessory slots full, return first one for swapping
		return SlotAccessory1, true
	}

	return 0, false
}

// RecalculateStats updates the cached stat bonuses from all equipped items.
func (e *EquipmentComponent) RecalculateStats() {
	// Reset cached stats
	e.CachedStats = item.Stats{}

	// Sum stats from all equipped items
	for _, itm := range e.Slots {
		if itm == nil {
			continue
		}

		e.CachedStats.Damage += itm.Stats.Damage
		e.CachedStats.Defense += itm.Stats.Defense

		// Attack speed: use the weapon's speed, don't sum
		if itm.Type == item.TypeWeapon && itm.Stats.AttackSpeed > 0 {
			e.CachedStats.AttackSpeed = itm.Stats.AttackSpeed
		}

		e.CachedStats.Value += itm.Stats.Value
		e.CachedStats.Weight += itm.Stats.Weight
	}

	e.StatsDirty = false
}

// GetStats returns the total stat bonuses from equipped items.
func (e *EquipmentComponent) GetStats() item.Stats {
	if e.StatsDirty {
		e.RecalculateStats()
	}
	return e.CachedStats
}

// GetTotalDefense returns the sum of defense from all equipped armor.
func (e *EquipmentComponent) GetTotalDefense() int {
	if e.StatsDirty {
		e.RecalculateStats()
	}
	return e.CachedStats.Defense
}

// GetWeaponDamage returns the damage from equipped main hand weapon.
func (e *EquipmentComponent) GetWeaponDamage() int {
	mainHand := e.Slots[SlotMainHand]
	if mainHand != nil && mainHand.Type == item.TypeWeapon {
		return mainHand.Stats.Damage
	}
	return 0
}

// GetWeaponSpeed returns the attack speed from equipped main hand weapon.
func (e *EquipmentComponent) GetWeaponSpeed() float64 {
	mainHand := e.Slots[SlotMainHand]
	if mainHand != nil && mainHand.Type == item.TypeWeapon {
		return mainHand.Stats.AttackSpeed
	}
	return 1.0 // Default speed
}

// UnequipAll removes all equipped items and returns them.
func (e *EquipmentComponent) UnequipAll() []*item.Item {
	items := make([]*item.Item, 0, len(e.Slots))
	for _, itm := range e.Slots {
		if itm != nil {
			items = append(items, itm)
		}
	}

	e.Slots = make(map[EquipmentSlot]*item.Item)
	e.StatsDirty = true

	return items
}
