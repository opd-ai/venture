package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// CompanionInventoryComponent manages a companion's item storage and carrying.
// Companions can fetch items, carry loot, and store items for the player.
type CompanionInventoryComponent struct {
	// Items stored in companion's inventory
	Items []*item.Item

	// MaxItems is the maximum number of items companion can carry
	MaxItems int

	// MaxWeight is the maximum weight companion can carry
	MaxWeight float64

	// AutoFetch enables automatic item pickup when near items
	AutoFetch bool

	// FetchRadius is the distance within which companion can fetch items
	FetchRadius float64
}

// Type returns the component type identifier.
func (c *CompanionInventoryComponent) Type() string {
	return "companioninventory"
}

// NewCompanionInventoryComponent creates a new companion inventory.
func NewCompanionInventoryComponent(maxItems int, maxWeight float64) *CompanionInventoryComponent {
	return &CompanionInventoryComponent{
		Items:       make([]*item.Item, 0, maxItems),
		MaxItems:    maxItems,
		MaxWeight:   maxWeight,
		AutoFetch:   false,
		FetchRadius: 100.0,
	}
}

// GetCurrentWeight returns the total weight of all items in companion's inventory.
func (c *CompanionInventoryComponent) GetCurrentWeight() float64 {
	totalWeight := 0.0
	for _, itm := range c.Items {
		totalWeight += itm.Stats.Weight
	}
	return totalWeight
}

// CanAddItem checks if an item can be added to companion's inventory.
func (c *CompanionInventoryComponent) CanAddItem(itm *item.Item) bool {
	// Check item count limit
	if len(c.Items) >= c.MaxItems {
		return false
	}

	// Check weight limit
	if c.GetCurrentWeight()+itm.Stats.Weight > c.MaxWeight {
		return false
	}

	return true
}

// AddItem adds an item to the companion's inventory if possible.
// Returns true if successful, false if inventory is full or weight exceeded.
func (c *CompanionInventoryComponent) AddItem(itm *item.Item) bool {
	if !c.CanAddItem(itm) {
		return false
	}

	c.Items = append(c.Items, itm)
	return true
}

// RemoveItem removes an item from inventory by index.
// Returns the removed item or nil if index is invalid.
func (c *CompanionInventoryComponent) RemoveItem(index int) *item.Item {
	if index < 0 || index >= len(c.Items) {
		return nil
	}

	itm := c.Items[index]
	c.Items = append(c.Items[:index], c.Items[index+1:]...)
	return itm
}

// RemoveItemByReference removes a specific item instance from inventory.
// Returns true if the item was found and removed.
func (c *CompanionInventoryComponent) RemoveItemByReference(itm *item.Item) bool {
	for idx, invItem := range c.Items {
		if invItem == itm {
			c.RemoveItem(idx)
			return true
		}
	}
	return false
}

// TransferToOwner transfers all items to the owner's inventory.
// Returns the items that couldn't be transferred due to owner inventory limits.
func (c *CompanionInventoryComponent) TransferToOwner(ownerInventory *InventoryComponent) []*item.Item {
	untransferred := make([]*item.Item, 0)

	for len(c.Items) > 0 {
		itm := c.RemoveItem(0)
		if itm != nil {
			if !ownerInventory.AddItem(itm) {
				untransferred = append(untransferred, itm)
			}
		}
	}

	// Add back items that couldn't be transferred
	for _, itm := range untransferred {
		c.AddItem(itm)
	}

	return untransferred
}

// GetItemCount returns the number of items in companion's inventory.
func (c *CompanionInventoryComponent) GetItemCount() int {
	return len(c.Items)
}

// IsFull returns true if companion inventory cannot accept more items.
func (c *CompanionInventoryComponent) IsFull() bool {
	return len(c.Items) >= c.MaxItems || c.GetCurrentWeight() >= c.MaxWeight
}

// Clear removes all items from companion's inventory.
func (c *CompanionInventoryComponent) Clear() {
	c.Items = c.Items[:0]
}
