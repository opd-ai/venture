package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestHotbarComponent_Type tests the Type method
func TestHotbarComponent_Type(t *testing.T) {
	hotbar := NewHotbarComponent()
	
	if hotbar.Type() != "hotbar" {
		t.Errorf("Expected type 'hotbar', got '%s'", hotbar.Type())
	}
}

// TestNewHotbarComponent tests constructor
func TestNewHotbarComponent(t *testing.T) {
	hotbar := NewHotbarComponent()

	if hotbar == nil {
		t.Fatal("NewHotbarComponent returned nil")
	}

	// Verify all slots are empty
	for i := 0; i < 6; i++ {
		if hotbar.Slots[i] != nil {
			t.Errorf("Slot %d should be nil, got %v", i, hotbar.Slots[i])
		}
		if hotbar.Cooldowns[i] != 0 {
			t.Errorf("Cooldown %d should be 0, got %f", i, hotbar.Cooldowns[i])
		}
		if hotbar.MaxCooldowns[i] != 1.0 {
			t.Errorf("MaxCooldown %d should be 1.0, got %f", i, hotbar.MaxCooldowns[i])
		}
	}

	if hotbar.LastUsedIndex != -1 {
		t.Errorf("LastUsedIndex should be -1, got %d", hotbar.LastUsedIndex)
	}
}

// TestHotbarComponent_SetSlot tests setting items in slots
func TestHotbarComponent_SetSlot(t *testing.T) {
	hotbar := NewHotbarComponent()
	testItem := &item.Item{
		Name: "Test Item",
		Type: item.TypeWeapon,
	}

	tests := []struct {
		name      string
		slotIndex int
		item      *item.Item
		wantOK    bool
	}{
		{"Valid slot 0", 0, testItem, true},
		{"Valid slot 5", 5, testItem, true},
		{"Invalid slot -1", -1, testItem, false},
		{"Invalid slot 6", 6, testItem, false},
		{"Valid nil item", 0, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := hotbar.SetSlot(tt.slotIndex, tt.item)
			if ok != tt.wantOK {
				t.Errorf("SetSlot() = %v, want %v", ok, tt.wantOK)
			}
			
			// Verify item was set if operation was successful and slot is valid
			if ok && tt.slotIndex >= 0 && tt.slotIndex < 6 {
				if hotbar.Slots[tt.slotIndex] != tt.item {
					t.Errorf("Slot %d item = %v, want %v", tt.slotIndex, hotbar.Slots[tt.slotIndex], tt.item)
				}
			}
		})
	}
}

// TestHotbarComponent_SetSlot_Consumable tests consumable cooldown setting
func TestHotbarComponent_SetSlot_Consumable(t *testing.T) {
	hotbar := NewHotbarComponent()
	consumable := &item.Item{
		Name: "Health Potion",
		Type: item.TypeConsumable,
	}

	ok := hotbar.SetSlot(0, consumable)
	if !ok {
		t.Fatal("SetSlot failed")
	}

	// Verify consumable sets 2s cooldown
	if hotbar.MaxCooldowns[0] != 2.0 {
		t.Errorf("Consumable max cooldown should be 2.0, got %f", hotbar.MaxCooldowns[0])
	}

	// Verify non-consumable keeps 1s cooldown
	weapon := &item.Item{
		Name: "Sword",
		Type: item.TypeWeapon,
	}
	hotbar.SetSlot(1, weapon)
	if hotbar.MaxCooldowns[1] != 1.0 {
		t.Errorf("Non-consumable max cooldown should remain 1.0, got %f", hotbar.MaxCooldowns[1])
	}
}

// TestHotbarComponent_GetSlot tests retrieving items from slots
func TestHotbarComponent_GetSlot(t *testing.T) {
	hotbar := NewHotbarComponent()
	testItem := &item.Item{
		Name: "Test Item",
	}

	hotbar.SetSlot(2, testItem)

	tests := []struct {
		name      string
		slotIndex int
		want      *item.Item
	}{
		{"Get valid slot with item", 2, testItem},
		{"Get valid empty slot", 0, nil},
		{"Get invalid slot -1", -1, nil},
		{"Get invalid slot 6", 6, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hotbar.GetSlot(tt.slotIndex)
			if got != tt.want {
				t.Errorf("GetSlot(%d) = %v, want %v", tt.slotIndex, got, tt.want)
			}
		})
	}
}

// TestHotbarComponent_ClearSlot tests clearing slots
func TestHotbarComponent_ClearSlot(t *testing.T) {
	hotbar := NewHotbarComponent()
	testItem := &item.Item{
		Name: "Test Item",
	}

	// Set up slots with items and cooldowns
	hotbar.SetSlot(0, testItem)
	hotbar.SetSlot(3, testItem)
	hotbar.Cooldowns[0] = 1.5
	hotbar.Cooldowns[3] = 2.0

	tests := []struct {
		name      string
		slotIndex int
	}{
		{"Clear valid slot", 0},
		{"Clear another valid slot", 3},
		{"Clear invalid slot -1", -1}, // Should not panic
		{"Clear invalid slot 6", 6},   // Should not panic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			hotbar.ClearSlot(tt.slotIndex)

			// Verify clearance if valid slot
			if tt.slotIndex >= 0 && tt.slotIndex < 6 {
				if hotbar.Slots[tt.slotIndex] != nil {
					t.Errorf("Slot %d should be nil after clear", tt.slotIndex)
				}
				if hotbar.Cooldowns[tt.slotIndex] != 0 {
					t.Errorf("Cooldown %d should be 0 after clear", tt.slotIndex)
				}
			}
		})
	}
}

// TestHotbarComponent_IsOnCooldown tests cooldown checking
func TestHotbarComponent_IsOnCooldown(t *testing.T) {
	hotbar := NewHotbarComponent()

	// Set up different cooldown states
	hotbar.Cooldowns[0] = 0     // Not on cooldown
	hotbar.Cooldowns[1] = 1.5   // On cooldown
	hotbar.Cooldowns[2] = 0.001 // Barely on cooldown

	tests := []struct {
		name      string
		slotIndex int
		want      bool
	}{
		{"No cooldown", 0, false},
		{"On cooldown", 1, true},
		{"Barely on cooldown", 2, true},
		{"Empty slot no cooldown", 3, false},
		{"Invalid slot -1", -1, true}, // Treat invalid as on cooldown
		{"Invalid slot 6", 6, true},   // Treat invalid as on cooldown
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hotbar.IsOnCooldown(tt.slotIndex)
			if got != tt.want {
				t.Errorf("IsOnCooldown(%d) = %v, want %v", tt.slotIndex, got, tt.want)
			}
		})
	}
}

// TestHotbarComponent_GetCooldownProgress tests cooldown progress calculation
func TestHotbarComponent_GetCooldownProgress(t *testing.T) {
	hotbar := NewHotbarComponent()

	// Set up different cooldown states
	hotbar.Cooldowns[0] = 0   // No cooldown
	hotbar.Cooldowns[1] = 2.0 // Full cooldown (max = 1.0, so 200%)
	hotbar.MaxCooldowns[1] = 2.0
	hotbar.Cooldowns[2] = 0.5 // Half cooldown
	hotbar.MaxCooldowns[2] = 1.0
	hotbar.Cooldowns[3] = 0.25 // Quarter cooldown
	hotbar.MaxCooldowns[3] = 1.0

	tests := []struct {
		name      string
		slotIndex int
		want      float64
	}{
		{"No cooldown", 0, 0.0},
		{"Full cooldown", 1, 1.0},
		{"Half cooldown", 2, 0.5},
		{"Quarter cooldown", 3, 0.25},
		{"Empty slot", 4, 0.0},
		{"Invalid slot -1", -1, 0.0},
		{"Invalid slot 6", 6, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hotbar.GetCooldownProgress(tt.slotIndex)
			tolerance := 0.0001
			if diff := got - tt.want; diff < -tolerance || diff > tolerance {
				t.Errorf("GetCooldownProgress(%d) = %f, want %f", tt.slotIndex, got, tt.want)
			}
		})
	}
}

// TestHotbarComponent_GetCooldownProgress_ZeroMax tests zero max cooldown edge case
func TestHotbarComponent_GetCooldownProgress_ZeroMax(t *testing.T) {
	hotbar := NewHotbarComponent()
	hotbar.MaxCooldowns[0] = 0 // Zero max cooldown
	hotbar.Cooldowns[0] = 5.0  // Non-zero current cooldown

	// Should return 0 to avoid division by zero
	got := hotbar.GetCooldownProgress(0)
	if got != 0.0 {
		t.Errorf("GetCooldownProgress with zero max should return 0.0, got %f", got)
	}
}

// TestHotbarComponent_TriggerCooldown tests triggering cooldowns
func TestHotbarComponent_TriggerCooldown(t *testing.T) {
	hotbar := NewHotbarComponent()

	// Set max cooldowns
	hotbar.MaxCooldowns[0] = 1.5
	hotbar.MaxCooldowns[2] = 2.0

	tests := []struct {
		name      string
		slotIndex int
		wantCD    float64
	}{
		{"Trigger slot 0", 0, 1.5},
		{"Trigger slot 2", 2, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hotbar.TriggerCooldown(tt.slotIndex)

			if hotbar.Cooldowns[tt.slotIndex] != tt.wantCD {
				t.Errorf("Cooldown %d = %f, want %f", tt.slotIndex, hotbar.Cooldowns[tt.slotIndex], tt.wantCD)
			}
			if hotbar.LastUsedIndex != tt.slotIndex {
				t.Errorf("LastUsedIndex = %d, want %d", hotbar.LastUsedIndex, tt.slotIndex)
			}
		})
	}
}

// TestHotbarComponent_TriggerCooldown_Invalid tests triggering invalid slots
func TestHotbarComponent_TriggerCooldown_Invalid(t *testing.T) {
	hotbar := NewHotbarComponent()
	initialLastUsed := hotbar.LastUsedIndex

	// Should not panic
	hotbar.TriggerCooldown(-1)
	hotbar.TriggerCooldown(6)

	// LastUsedIndex should not change for invalid slots
	if hotbar.LastUsedIndex != initialLastUsed {
		t.Errorf("LastUsedIndex changed after invalid trigger: %d", hotbar.LastUsedIndex)
	}
}

// TestHotbarComponent_UpdateCooldowns tests cooldown updates over time
func TestHotbarComponent_UpdateCooldowns(t *testing.T) {
	hotbar := NewHotbarComponent()

	// Set up cooldowns
	hotbar.Cooldowns[0] = 2.0
	hotbar.Cooldowns[1] = 0.5
	hotbar.Cooldowns[2] = 0.1
	hotbar.Cooldowns[3] = 0.0

	// Update by 0.5 seconds
	hotbar.UpdateCooldowns(0.5)

	tests := []struct {
		slotIndex int
		want      float64
	}{
		{0, 1.5},  // 2.0 - 0.5
		{1, 0.0},  // 0.5 - 0.5 (clamped to 0)
		{2, 0.0},  // 0.1 - 0.5 (clamped to 0)
		{3, 0.0},  // Already 0
	}

	for _, tt := range tests {
		t.Run("Slot "+string(rune(tt.slotIndex+'0')), func(t *testing.T) {
			tolerance := 0.0001
			if diff := hotbar.Cooldowns[tt.slotIndex] - tt.want; diff < -tolerance || diff > tolerance {
				t.Errorf("After update, cooldown %d = %f, want %f", tt.slotIndex, hotbar.Cooldowns[tt.slotIndex], tt.want)
			}
		})
	}
}

// TestHotbarComponent_UpdateCooldowns_Multiple tests multiple updates
func TestHotbarComponent_UpdateCooldowns_Multiple(t *testing.T) {
	hotbar := NewHotbarComponent()
	hotbar.Cooldowns[0] = 3.0

	// Update multiple times
	for i := 0; i < 5; i++ {
		hotbar.UpdateCooldowns(0.5)
	}

	// After 5 updates of 0.5s each (2.5s total), cooldown should be 0.5s
	expected := 0.5
	tolerance := 0.0001
	if diff := hotbar.Cooldowns[0] - expected; diff < -tolerance || diff > tolerance {
		t.Errorf("After 5 updates, cooldown = %f, want %f", hotbar.Cooldowns[0], expected)
	}
}

// TestHotbarComponent_UpdateCooldowns_Negative tests that cooldowns don't go negative
func TestHotbarComponent_UpdateCooldowns_Negative(t *testing.T) {
	hotbar := NewHotbarComponent()
	hotbar.Cooldowns[0] = 0.1

	// Update by more than remaining cooldown
	hotbar.UpdateCooldowns(1.0)

	// Should clamp to 0, not go negative
	if hotbar.Cooldowns[0] < 0 {
		t.Errorf("Cooldown went negative: %f", hotbar.Cooldowns[0])
	}
	if hotbar.Cooldowns[0] != 0 {
		t.Errorf("Cooldown should be exactly 0, got %f", hotbar.Cooldowns[0])
	}
}

// TestHotbarComponent_Integration tests full workflow
func TestHotbarComponent_Integration(t *testing.T) {
	hotbar := NewHotbarComponent()

	// Create test items
	potion := &item.Item{
		Name: "Health Potion",
		Type: item.TypeConsumable,
	}
	weapon := &item.Item{
		Name: "Iron Sword",
		Type: item.TypeWeapon,
	}

	// Set up hotbar
	if !hotbar.SetSlot(0, potion) {
		t.Fatal("Failed to set potion in slot 0")
	}
	if !hotbar.SetSlot(1, weapon) {
		t.Fatal("Failed to set weapon in slot 1")
	}

	// Verify consumable has 2s cooldown
	if hotbar.MaxCooldowns[0] != 2.0 {
		t.Errorf("Potion max cooldown should be 2.0, got %f", hotbar.MaxCooldowns[0])
	}

	// Use potion (slot 0)
	if hotbar.IsOnCooldown(0) {
		t.Error("Slot 0 should not be on cooldown initially")
	}

	hotbar.TriggerCooldown(0)

	if !hotbar.IsOnCooldown(0) {
		t.Error("Slot 0 should be on cooldown after trigger")
	}
	if hotbar.GetCooldownProgress(0) != 1.0 {
		t.Errorf("Progress should be 1.0 immediately after trigger, got %f", hotbar.GetCooldownProgress(0))
	}
	if hotbar.LastUsedIndex != 0 {
		t.Errorf("LastUsedIndex should be 0, got %d", hotbar.LastUsedIndex)
	}

	// Update cooldown by 1 second
	hotbar.UpdateCooldowns(1.0)

	if hotbar.Cooldowns[0] != 1.0 {
		t.Errorf("After 1s, cooldown should be 1.0, got %f", hotbar.Cooldowns[0])
	}
	if hotbar.GetCooldownProgress(0) != 0.5 {
		t.Errorf("Progress should be 0.5, got %f", hotbar.GetCooldownProgress(0))
	}

	// Update by another second
	hotbar.UpdateCooldowns(1.0)

	if hotbar.IsOnCooldown(0) {
		t.Error("Slot 0 should no longer be on cooldown")
	}
	if hotbar.Cooldowns[0] != 0 {
		t.Errorf("Cooldown should be 0, got %f", hotbar.Cooldowns[0])
	}

	// Clear slot
	hotbar.ClearSlot(0)

	if hotbar.GetSlot(0) != nil {
		t.Error("Slot 0 should be nil after clear")
	}
}

// TestHotbarComponent_AllSlotsUsable tests that all 6 slots work
func TestHotbarComponent_AllSlotsUsable(t *testing.T) {
	hotbar := NewHotbarComponent()
	testItem := &item.Item{
		Name: "Test",
	}

	// Test all 6 slots
	for i := 0; i < 6; i++ {
		if !hotbar.SetSlot(i, testItem) {
			t.Errorf("Failed to set slot %d", i)
		}
		if hotbar.GetSlot(i) != testItem {
			t.Errorf("Slot %d doesn't contain expected item", i)
		}

		hotbar.TriggerCooldown(i)
		if !hotbar.IsOnCooldown(i) {
			t.Errorf("Slot %d should be on cooldown", i)
		}

		hotbar.UpdateCooldowns(2.0) // Clear cooldown
		hotbar.ClearSlot(i)

		if hotbar.GetSlot(i) != nil {
			t.Errorf("Slot %d should be clear", i)
		}
	}
}
