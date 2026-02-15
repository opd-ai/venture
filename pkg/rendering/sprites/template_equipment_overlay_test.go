package sprites

import (
	"testing"
)

func TestOverlayEquipmentVisuals_NilCustom(t *testing.T) {
	// No equipment data — should not panic
	config := Config{
		Width:  32,
		Height: 32,
		Custom: nil,
	}
	// Use a 1x1 image as stand-in (can't create ebiten.Image in test)
	overlayEquipmentVisuals(nil, config)
}

func TestOverlayEquipmentVisuals_EmptySlots(t *testing.T) {
	config := Config{
		Width:  32,
		Height: 32,
		Custom: map[string]interface{}{
			"equipmentVisuals": []EquipmentVisual{},
		},
	}
	overlayEquipmentVisuals(nil, config)
}

func TestOverlayEquipmentVisuals_WrongType(t *testing.T) {
	config := Config{
		Width:  32,
		Height: 32,
		Custom: map[string]interface{}{
			"equipmentVisuals": "not a slice",
		},
	}
	// Should not panic when config.Custom has wrong type
	overlayEquipmentVisuals(nil, config)
}

func TestSortEquipmentByZIndex(t *testing.T) {
	slots := []EquipmentVisual{
		{Slot: "accessory", Seed: 1},
		{Slot: "armor", Seed: 2},
		{Slot: "weapon", Seed: 3},
		{Slot: "helmet", Seed: 4},
	}

	sorted := sortEquipmentByZIndex(slots)

	expected := []string{"armor", "helmet", "weapon", "accessory"}
	for i, slot := range sorted {
		if slot.Slot != expected[i] {
			t.Errorf("position %d: got %q, want %q", i, slot.Slot, expected[i])
		}
	}

	// Verify original slice is not modified
	if slots[0].Slot != "accessory" {
		t.Error("sortEquipmentByZIndex modified original slice")
	}
}

func TestSlotZIndex(t *testing.T) {
	tests := []struct {
		slot     string
		expected int
	}{
		{"armor", ZIndexArmor},
		{"helmet", ZIndexHead},
		{"weapon", ZIndexWeapon},
		{"shield", ZIndexWeapon - 1},
		{"accessory", ZIndexAccessory},
		{"unknown", ZIndexAccessory},
	}

	for _, tt := range tests {
		t.Run(tt.slot, func(t *testing.T) {
			got := slotZIndex(tt.slot)
			if got != tt.expected {
				t.Errorf("slotZIndex(%q) = %d, want %d", tt.slot, got, tt.expected)
			}
		})
	}
}

func TestEquipmentSlotBounds(t *testing.T) {
	tests := []struct {
		slot    string
		spriteW int
		spriteH int
	}{
		{"weapon", 32, 32},
		{"armor", 32, 32},
		{"helmet", 32, 32},
		{"shield", 32, 32},
		{"accessory", 32, 32},
		{"unknown", 32, 32},
		{"weapon", 64, 64},
		{"armor", 16, 16},
	}

	for _, tt := range tests {
		t.Run(tt.slot+"_"+string(rune('0'+tt.spriteW/16)), func(t *testing.T) {
			w, h, offsetX, offsetY := equipmentSlotBounds(tt.slot, tt.spriteW, tt.spriteH)

			if w <= 0 || h <= 0 {
				t.Errorf("slot %q: got non-positive size (%d×%d)", tt.slot, w, h)
			}
			if w > tt.spriteW || h > tt.spriteH {
				t.Errorf("slot %q: overlay larger than sprite (%d×%d > %d×%d)", tt.slot, w, h, tt.spriteW, tt.spriteH)
			}
			if offsetX < 0 || offsetY < 0 {
				t.Errorf("slot %q: negative offset (%d, %d)", tt.slot, offsetX, offsetY)
			}
			if offsetX+w > tt.spriteW+4 || offsetY+h > tt.spriteH+4 {
				t.Logf("slot %q: overlay extends slightly beyond sprite bounds (offset=%d,%d size=%d×%d sprite=%d×%d)",
					tt.slot, offsetX, offsetY, w, h, tt.spriteW, tt.spriteH)
			}
		})
	}
}

func TestEquipmentSlotBounds_SmallSprite(t *testing.T) {
	// Minimum sprite size — ensure no zero-size overlays
	slots := []string{"weapon", "armor", "helmet", "shield", "accessory"}
	for _, slot := range slots {
		w, h, _, _ := equipmentSlotBounds(slot, 8, 8)
		if w <= 0 || h <= 0 {
			t.Errorf("slot %q at 8×8: got non-positive size (%d×%d)", slot, w, h)
		}
	}
}

func TestSortEquipmentByZIndex_SingleItem(t *testing.T) {
	slots := []EquipmentVisual{{Slot: "weapon", Seed: 42}}
	sorted := sortEquipmentByZIndex(slots)
	if len(sorted) != 1 || sorted[0].Slot != "weapon" {
		t.Error("single item sort failed")
	}
}

func TestSortEquipmentByZIndex_AlreadySorted(t *testing.T) {
	slots := []EquipmentVisual{
		{Slot: "armor"},
		{Slot: "helmet"},
		{Slot: "weapon"},
	}
	sorted := sortEquipmentByZIndex(slots)
	expected := []string{"armor", "helmet", "weapon"}
	for i, s := range sorted {
		if s.Slot != expected[i] {
			t.Errorf("position %d: got %q, want %q", i, s.Slot, expected[i])
		}
	}
}

func TestEquipmentSlotBounds_Deterministic(t *testing.T) {
	// Same input produces same output
	w1, h1, ox1, oy1 := equipmentSlotBounds("weapon", 32, 32)
	w2, h2, ox2, oy2 := equipmentSlotBounds("weapon", 32, 32)
	if w1 != w2 || h1 != h2 || ox1 != ox2 || oy1 != oy2 {
		t.Error("equipmentSlotBounds not deterministic")
	}
}
