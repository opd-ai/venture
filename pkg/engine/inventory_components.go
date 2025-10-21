package engine

import "github.com/opd-ai/venture/pkg/procgen/item"

// InventoryComponent manages an entity's item storage.
type InventoryComponent struct {
	// Items stored in inventory
	Items []*item.Item
	
	// MaxCapacity is the maximum number of items
	MaxCapacity int
	
	// MaxWeight is the maximum total weight that can be carried
	MaxWeight float64
}

// Type returns the component type identifier.
func (i *InventoryComponent) Type() string {
	return "inventory"
}

// NewInventoryComponent creates a new inventory with the given capacity.
func NewInventoryComponent(maxCapacity int, maxWeight float64) *InventoryComponent {
	return &InventoryComponent{
		Items:       make([]*item.Item, 0, maxCapacity),
		MaxCapacity: maxCapacity,
		MaxWeight:   maxWeight,
	}
}

// IsFull returns true if the inventory is at maximum capacity.
func (i *InventoryComponent) IsFull() bool {
	return len(i.Items) >= i.MaxCapacity
}

// GetCurrentWeight returns the total weight of items in inventory.
func (i *InventoryComponent) GetCurrentWeight() float64 {
	total := 0.0
	for _, itm := range i.Items {
		total += itm.Stats.Weight
	}
	return total
}

// IsOverweight returns true if current weight exceeds maximum.
func (i *InventoryComponent) IsOverweight() bool {
	return i.GetCurrentWeight() > i.MaxWeight
}

// CanAddItem returns true if the item can be added to inventory.
func (i *InventoryComponent) CanAddItem(itm *item.Item) bool {
	if i.IsFull() {
		return false
	}
	futureWeight := i.GetCurrentWeight() + itm.Stats.Weight
	return futureWeight <= i.MaxWeight
}

// AddItem adds an item to the inventory.
// Returns true if successful, false if inventory is full or would exceed weight limit.
func (i *InventoryComponent) AddItem(itm *item.Item) bool {
	if !i.CanAddItem(itm) {
		return false
	}
	i.Items = append(i.Items, itm)
	return true
}

// RemoveItem removes an item at the given index.
// Returns the removed item, or nil if index is invalid.
func (i *InventoryComponent) RemoveItem(index int) *item.Item {
	if index < 0 || index >= len(i.Items) {
		return nil
	}
	itm := i.Items[index]
	i.Items = append(i.Items[:index], i.Items[index+1:]...)
	return itm
}

// FindItem returns the index of the first item matching the given predicate.
// Returns -1 if no item is found.
func (i *InventoryComponent) FindItem(predicate func(*item.Item) bool) int {
	for idx, itm := range i.Items {
		if predicate(itm) {
			return idx
		}
	}
	return -1
}

// GetItemCount returns the number of items in inventory.
func (i *InventoryComponent) GetItemCount() int {
	return len(i.Items)
}

// Clear removes all items from inventory.
func (i *InventoryComponent) Clear() {
	i.Items = i.Items[:0]
}

// EquipmentSlot represents a type of equipment slot.
type EquipmentSlot int

const (
	// SlotWeapon is the main hand weapon slot
	SlotWeapon EquipmentSlot = iota
	// SlotOffhand is the off-hand slot (shields, secondary weapons)
	SlotOffhand
	// SlotHead is the helmet/head armor slot
	SlotHead
	// SlotChest is the chest/torso armor slot
	SlotChest
	// SlotLegs is the leg armor slot
	SlotLegs
	// SlotBoots is the boot/feet armor slot
	SlotBoots
	// SlotGloves is the glove/hand armor slot
	SlotGloves
	// SlotAccessory1 is the first accessory slot
	SlotAccessory1
	// SlotAccessory2 is the second accessory slot
	SlotAccessory2
)

// String returns the string representation of an equipment slot.
func (s EquipmentSlot) String() string {
	switch s {
	case SlotWeapon:
		return "weapon"
	case SlotOffhand:
		return "offhand"
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
		return "accessory1"
	case SlotAccessory2:
		return "accessory2"
	default:
		return "unknown"
	}
}

// EquipmentComponent manages equipped items and their stat bonuses.
type EquipmentComponent struct {
	// Slots maps equipment slot to equipped item
	Slots map[EquipmentSlot]*item.Item
}

// Type returns the component type identifier.
func (e *EquipmentComponent) Type() string {
	return "equipment"
}

// NewEquipmentComponent creates a new equipment component.
func NewEquipmentComponent() *EquipmentComponent {
	return &EquipmentComponent{
		Slots: make(map[EquipmentSlot]*item.Item),
	}
}

// GetSlotForItem returns the appropriate equipment slot for an item type.
// Returns -1 if the item cannot be equipped.
func GetSlotForItem(itm *item.Item) EquipmentSlot {
	switch itm.Type {
	case item.TypeWeapon:
		return SlotWeapon
	case item.TypeArmor:
		switch itm.ArmorType {
		case item.ArmorHelmet:
			return SlotHead
		case item.ArmorChest:
			return SlotChest
		case item.ArmorLegs:
			return SlotLegs
		case item.ArmorBoots:
			return SlotBoots
		case item.ArmorGloves:
			return SlotGloves
		case item.ArmorShield:
			return SlotOffhand
		}
	case item.TypeAccessory:
		// Check if first accessory slot is empty
		// This is determined by the equipment system during equip
		return SlotAccessory1
	}
	return EquipmentSlot(-1)
}

// Equip equips an item to the appropriate slot.
// Returns the previously equipped item in that slot, or nil.
func (e *EquipmentComponent) Equip(itm *item.Item) *item.Item {
	slot := GetSlotForItem(itm)
	if slot == EquipmentSlot(-1) {
		return nil
	}
	
	// For accessories, use the first available slot
	if itm.Type == item.TypeAccessory {
		if _, exists := e.Slots[SlotAccessory1]; !exists {
			slot = SlotAccessory1
		} else {
			slot = SlotAccessory2
		}
	}
	
	previous := e.Slots[slot]
	e.Slots[slot] = itm
	return previous
}

// Unequip removes an item from the given slot.
// Returns the unequipped item, or nil if slot was empty.
func (e *EquipmentComponent) Unequip(slot EquipmentSlot) *item.Item {
	itm := e.Slots[slot]
	delete(e.Slots, slot)
	return itm
}

// GetEquipped returns the item in the given slot, or nil if empty.
func (e *EquipmentComponent) GetEquipped(slot EquipmentSlot) *item.Item {
	return e.Slots[slot]
}

// IsSlotEmpty returns true if the given slot has no item equipped.
func (e *EquipmentComponent) IsSlotEmpty(slot EquipmentSlot) bool {
	_, exists := e.Slots[slot]
	return !exists
}

// GetTotalDamage returns the total damage from equipped weapons.
func (e *EquipmentComponent) GetTotalDamage() int {
	total := 0
	if weapon := e.Slots[SlotWeapon]; weapon != nil {
		total += weapon.Stats.Damage
	}
	return total
}

// GetTotalDefense returns the total defense from equipped armor.
func (e *EquipmentComponent) GetTotalDefense() int {
	total := 0
	for slot, itm := range e.Slots {
		// Only count armor pieces for defense
		if slot != SlotWeapon && itm != nil {
			total += itm.Stats.Defense
		}
	}
	return total
}

// GetAllEquipped returns all currently equipped items.
func (e *EquipmentComponent) GetAllEquipped() []*item.Item {
	items := make([]*item.Item, 0, len(e.Slots))
	for _, itm := range e.Slots {
		if itm != nil {
			items = append(items, itm)
		}
	}
	return items
}

// Clear removes all equipped items.
func (e *EquipmentComponent) Clear() {
	e.Slots = make(map[EquipmentSlot]*item.Item)
}
