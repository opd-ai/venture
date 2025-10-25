// Package engine provides equipment visual component for rendering equipped items on sprites.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// EquipmentVisualComponent manages visual representation of equipped items.
// This component tracks which items are equipped and triggers sprite regeneration
// when equipment changes.
type EquipmentVisualComponent struct {
	// Weapon visual layer
	WeaponLayer *ebiten.Image

	// Armor visual layer
	ArmorLayer *ebiten.Image

	// Accessory visual layers (hat, cape, rings, etc.)
	AccessoryLayers []*ebiten.Image

	// Equipment IDs for tracking changes
	WeaponID     string
	ArmorID      string
	AccessoryIDs []string

	// Seeds for deterministic generation
	WeaponSeed     int64
	ArmorSeed      int64
	AccessorySeeds []int64

	// Dirty flag - regenerate layers if true
	Dirty bool

	// Visibility flags per layer
	ShowWeapon      bool
	ShowArmor       bool
	ShowAccessories bool
}

// Type returns the component type identifier.
func (e *EquipmentVisualComponent) Type() string {
	return "equipment_visual"
}

// NewEquipmentVisualComponent creates a new equipment visual component.
func NewEquipmentVisualComponent() *EquipmentVisualComponent {
	return &EquipmentVisualComponent{
		AccessoryLayers: make([]*ebiten.Image, 0),
		AccessoryIDs:    make([]string, 0),
		AccessorySeeds:  make([]int64, 0),
		Dirty:           true,
		ShowWeapon:      true,
		ShowArmor:       true,
		ShowAccessories: true,
	}
}

// SetWeapon updates the weapon visual.
func (e *EquipmentVisualComponent) SetWeapon(itemID string, seed int64) {
	if e.WeaponID != itemID {
		e.WeaponID = itemID
		e.WeaponSeed = seed
		e.Dirty = true
	}
}

// SetArmor updates the armor visual.
func (e *EquipmentVisualComponent) SetArmor(itemID string, seed int64) {
	if e.ArmorID != itemID {
		e.ArmorID = itemID
		e.ArmorSeed = seed
		e.Dirty = true
	}
}

// AddAccessory adds an accessory visual.
func (e *EquipmentVisualComponent) AddAccessory(itemID string, seed int64) {
	e.AccessoryIDs = append(e.AccessoryIDs, itemID)
	e.AccessorySeeds = append(e.AccessorySeeds, seed)
	e.AccessoryLayers = append(e.AccessoryLayers, nil) // Will be generated
	e.Dirty = true
}

// RemoveAccessory removes an accessory by index.
func (e *EquipmentVisualComponent) RemoveAccessory(index int) {
	if index < 0 || index >= len(e.AccessoryIDs) {
		return
	}

	// Remove from all slices
	e.AccessoryIDs = append(e.AccessoryIDs[:index], e.AccessoryIDs[index+1:]...)
	e.AccessorySeeds = append(e.AccessorySeeds[:index], e.AccessorySeeds[index+1:]...)
	e.AccessoryLayers = append(e.AccessoryLayers[:index], e.AccessoryLayers[index+1:]...)
	e.Dirty = true
}

// ClearWeapon removes the weapon visual.
func (e *EquipmentVisualComponent) ClearWeapon() {
	if e.WeaponID != "" {
		e.WeaponID = ""
		e.WeaponLayer = nil
		e.Dirty = true
	}
}

// ClearArmor removes the armor visual.
func (e *EquipmentVisualComponent) ClearArmor() {
	if e.ArmorID != "" {
		e.ArmorID = ""
		e.ArmorLayer = nil
		e.Dirty = true
	}
}

// ClearAccessories removes all accessory visuals.
func (e *EquipmentVisualComponent) ClearAccessories() {
	if len(e.AccessoryIDs) > 0 {
		e.AccessoryIDs = make([]string, 0)
		e.AccessorySeeds = make([]int64, 0)
		e.AccessoryLayers = make([]*ebiten.Image, 0)
		e.Dirty = true
	}
}

// HasWeapon returns true if a weapon is equipped.
func (e *EquipmentVisualComponent) HasWeapon() bool {
	return e.WeaponID != ""
}

// HasArmor returns true if armor is equipped.
func (e *EquipmentVisualComponent) HasArmor() bool {
	return e.ArmorID != ""
}

// HasAccessories returns true if any accessories are equipped.
func (e *EquipmentVisualComponent) HasAccessories() bool {
	return len(e.AccessoryIDs) > 0
}

// MarkClean resets the dirty flag after regeneration.
func (e *EquipmentVisualComponent) MarkClean() {
	e.Dirty = false
}

// MarkDirty forces regeneration on next update.
func (e *EquipmentVisualComponent) MarkDirty() {
	e.Dirty = true
}
