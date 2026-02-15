// Package sprites integrates equipment overlay rendering into the template-based
// sprite pipeline. When entities are rendered via generateEntityWithTemplate,
// this module overlays weapon, armor, shield, helmet, and accessory visuals
// on top of the anatomy-rendered body, bridging the gap between the template
// pipeline (which produces high-quality aerial sprites) and the equipment
// renderer (which was previously only available through the composite pipeline).
package sprites

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// equipmentRenderer is a package-level renderer instance to avoid repeated allocation.
var equipmentRenderer = NewEquipmentRenderer()

// overlayEquipmentVisuals reads equipment visual data from config.Custom and
// renders equipment overlays on top of the given template-rendered sprite image.
// Equipment is rendered in ZIndex order: armor(15), helmet(20), weapon(25), accessory(30).
// Each slot occupies a specific region of the 32×32 sprite for aerial/top-down view.
func overlayEquipmentVisuals(img *ebiten.Image, config Config) {
	if config.Custom == nil {
		return
	}

	equipSlots, ok := config.Custom["equipmentVisuals"].([]EquipmentVisual)
	if !ok || len(equipSlots) == 0 {
		return
	}

	w := config.Width
	h := config.Height

	// Sort equipment by ZIndex order for correct layering
	sorted := sortEquipmentByZIndex(equipSlots)

	for _, equip := range sorted {
		rng := rand.New(rand.NewSource(equip.Seed))
		overlayW, overlayH, offsetX, offsetY := equipmentSlotBounds(equip.Slot, w, h)

		if overlayW <= 0 || overlayH <= 0 {
			continue
		}

		equipImg := equipmentRenderer.RenderEquipment(overlayW, overlayH, equip, rng)
		if equipImg == nil {
			continue
		}

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(offsetX), float64(offsetY))
		// Equipment layers are semi-transparent so the body shows through slightly
		opts.ColorScale.ScaleAlpha(0.92)

		img.DrawImage(equipImg, opts)
	}
}

// sortEquipmentByZIndex returns a copy sorted by rendering order.
func sortEquipmentByZIndex(slots []EquipmentVisual) []EquipmentVisual {
	sorted := make([]EquipmentVisual, len(slots))
	copy(sorted, slots)

	// Simple insertion sort — typically ≤5 items
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && slotZIndex(sorted[j].Slot) > slotZIndex(key.Slot) {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	return sorted
}

// slotZIndex returns the rendering z-order for an equipment slot type.
func slotZIndex(slot string) int {
	switch slot {
	case "armor":
		return ZIndexArmor
	case "helmet":
		return ZIndexHead
	case "weapon":
		return ZIndexWeapon
	case "shield":
		return ZIndexWeapon - 1
	case "accessory":
		return ZIndexAccessory
	default:
		return ZIndexAccessory
	}
}

// equipmentSlotBounds returns the overlay size and position for each equipment
// slot in a top-down aerial sprite. Sizes are relative to the sprite dimensions
// to support different sprite sizes.
func equipmentSlotBounds(slot string, spriteW, spriteH int) (w, h, offsetX, offsetY int) {
	switch slot {
	case "weapon":
		// Weapon extends from right shoulder outward — visible from above
		w = max(spriteW*40/100, 4)
		h = max(spriteH*55/100, 4)
		offsetX = spriteW*60/100 - w/4
		offsetY = spriteH*20/100
	case "armor":
		// Armor covers the torso area — shoulders and chest from above
		w = max(spriteW*65/100, 4)
		h = max(spriteH*45/100, 4)
		offsetX = (spriteW - w) / 2
		offsetY = spriteH*30/100
	case "helmet":
		// Helmet covers the top of the head
		w = max(spriteW*40/100, 4)
		h = max(spriteH*35/100, 4)
		offsetX = (spriteW - w) / 2
		offsetY = spriteH*5/100
	case "shield":
		// Shield on the left side, opposite the weapon
		w = max(spriteW*30/100, 4)
		h = max(spriteH*35/100, 4)
		offsetX = spriteW*5/100
		offsetY = spriteH*30/100
	case "accessory":
		// Accessories float near the head/shoulders
		w = max(spriteW*20/100, 3)
		h = max(spriteH*20/100, 3)
		offsetX = spriteW*10/100
		offsetY = spriteH*10/100
	default:
		w = max(spriteW*25/100, 3)
		h = max(spriteH*25/100, 3)
		offsetX = (spriteW - w) / 2
		offsetY = (spriteH - h) / 2
	}
	return
}


