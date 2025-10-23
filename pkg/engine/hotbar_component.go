// Package engine provides hotbar quickslot functionality.
// This file implements HotbarComponent which stores item quickslots for fast access during combat.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// HotbarComponent stores quickslot assignments for fast item access.
// Slots are typically mapped to number keys 1-6 for instant use during combat.
//
// Usage:
//
//	hotbar := engine.NewHotbarComponent()
//	hotbar.SetSlot(0, myPotion) // Assign potion to slot 1 (key 1)
//	if !hotbar.IsOnCooldown(0) {
//	    // Use item, trigger cooldown
//	    hotbar.TriggerCooldown(0)
//	}
type HotbarComponent struct {
	Slots         [6]*item.Item // Item references (nil = empty slot)
	Cooldowns     [6]float64    // Remaining cooldown per slot (seconds)
	MaxCooldowns  [6]float64    // Maximum cooldown per slot
	LastUsedIndex int           // Track last used for UI feedback (-1 = none)
}

// Type returns the component type identifier.
func (h *HotbarComponent) Type() string {
	return "hotbar"
}

// NewHotbarComponent creates a new hotbar with empty slots.
func NewHotbarComponent() *HotbarComponent {
	return &HotbarComponent{
		Slots:         [6]*item.Item{},
		Cooldowns:     [6]float64{},
		MaxCooldowns:  [6]float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}, // Default 1s cooldown
		LastUsedIndex: -1,
	}
}

// SetSlot assigns an item to a hotbar slot.
// Returns false if slotIndex out of range (valid: 0-5).
func (h *HotbarComponent) SetSlot(slotIndex int, itm *item.Item) bool {
	if slotIndex < 0 || slotIndex >= 6 {
		return false
	}
	h.Slots[slotIndex] = itm

	// Set cooldown based on item type
	if itm != nil && itm.Type == item.TypeConsumable {
		h.MaxCooldowns[slotIndex] = 2.0 // 2s cooldown for consumables
	}
	return true
}

// GetSlot retrieves the item in a hotbar slot.
// Returns nil if empty or out of range.
func (h *HotbarComponent) GetSlot(slotIndex int) *item.Item {
	if slotIndex < 0 || slotIndex >= 6 {
		return nil
	}
	return h.Slots[slotIndex]
}

// ClearSlot removes an item from a hotbar slot.
func (h *HotbarComponent) ClearSlot(slotIndex int) {
	if slotIndex >= 0 && slotIndex < 6 {
		h.Slots[slotIndex] = nil
		h.Cooldowns[slotIndex] = 0
	}
}

// IsOnCooldown checks if a slot is on cooldown.
// Returns true if the slot cannot be used yet.
func (h *HotbarComponent) IsOnCooldown(slotIndex int) bool {
	if slotIndex < 0 || slotIndex >= 6 {
		return true // Treat invalid slots as always on cooldown
	}
	return h.Cooldowns[slotIndex] > 0
}

// GetCooldownProgress returns cooldown progress fraction.
// Returns 0.0 if ready to use, 1.0 if just used, value in between during cooldown.
func (h *HotbarComponent) GetCooldownProgress(slotIndex int) float64 {
	if slotIndex < 0 || slotIndex >= 6 {
		return 0
	}
	if h.MaxCooldowns[slotIndex] == 0 {
		return 0
	}
	return h.Cooldowns[slotIndex] / h.MaxCooldowns[slotIndex]
}

// TriggerCooldown starts cooldown for a slot after item use.
func (h *HotbarComponent) TriggerCooldown(slotIndex int) {
	if slotIndex >= 0 && slotIndex < 6 {
		h.Cooldowns[slotIndex] = h.MaxCooldowns[slotIndex]
		h.LastUsedIndex = slotIndex
	}
}

// UpdateCooldowns decreases all active cooldowns by deltaTime.
// Called every frame by HotbarSystem.
func (h *HotbarComponent) UpdateCooldowns(deltaTime float64) {
	for i := range h.Cooldowns {
		if h.Cooldowns[i] > 0 {
			h.Cooldowns[i] -= deltaTime
			if h.Cooldowns[i] < 0 {
				h.Cooldowns[i] = 0
			}
		}
	}
}
