// Package engine provides tests for the CarryOver component.
// Tests cover selection limits, locking, serialization, and carry-over categories.
//
// Phase 112: Carry-Over System Tests
package engine

import (
	"encoding/json"
	"testing"
)

func TestCarryOverComponent_Type(t *testing.T) {
	c := NewCarryOverComponent()
	if c.Type() != "carryover" {
		t.Errorf("Type() = %q, want %q", c.Type(), "carryover")
	}
}

func TestNewCarryOverComponent(t *testing.T) {
	c := NewCarryOverComponent()

	if c == nil {
		t.Fatal("NewCarryOverComponent() returned nil")
	}
	if c.EquipmentSlotLimit != 3 {
		t.Errorf("EquipmentSlotLimit = %d, want 3", c.EquipmentSlotLimit)
	}
	if c.CurrencyPercentLimit != 50.0 {
		t.Errorf("CurrencyPercentLimit = %f, want 50.0", c.CurrencyPercentLimit)
	}
	if c.SelectionLocked {
		t.Error("SelectionLocked should be false initially")
	}
	if c.TransferComplete {
		t.Error("TransferComplete should be false initially")
	}
}

func TestCarryOverComponent_SetLimitsFromNGPlus(t *testing.T) {
	tests := []struct {
		name           string
		ngpCycle       int
		totalSkills    int
		wantSkillLimit int
	}{
		{
			name:           "first playthrough",
			ngpCycle:       0,
			totalSkills:    10,
			wantSkillLimit: 5, // 50% of 10
		},
		{
			name:           "NG+1",
			ngpCycle:       1,
			totalSkills:    10,
			wantSkillLimit: 5, // 55% -> 5.5 -> 5
		},
		{
			name:           "NG+5",
			ngpCycle:       5,
			totalSkills:    10,
			wantSkillLimit: 7, // 75% of 10
		},
		{
			name:           "NG+10 caps at 100%",
			ngpCycle:       10,
			totalSkills:    10,
			wantSkillLimit: 10, // 100% capped
		},
		{
			name:           "no skills",
			ngpCycle:       0,
			totalSkills:    0,
			wantSkillLimit: 0,
		},
		{
			name:           "minimum 1 skill when available",
			ngpCycle:       0,
			totalSkills:    1,
			wantSkillLimit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCarryOverComponent()
			ngp := NewNewGamePlusComponent()
			ngp.Cycle = tt.ngpCycle

			c.SetLimitsFromNGPlus(ngp, tt.totalSkills)

			if c.SkillSlotLimit != tt.wantSkillLimit {
				t.Errorf("SkillSlotLimit = %d, want %d", c.SkillSlotLimit, tt.wantSkillLimit)
			}
		})
	}
}

func TestCarryOverComponent_EquipmentSelection(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 3

	// Test adding equipment
	if !c.SelectEquipment("sword_123") {
		t.Error("SelectEquipment(sword_123) should succeed")
	}
	if !c.SelectEquipment("armor_456") {
		t.Error("SelectEquipment(armor_456) should succeed")
	}
	if !c.SelectEquipment("ring_789") {
		t.Error("SelectEquipment(ring_789) should succeed")
	}

	// Should fail at limit
	if c.SelectEquipment("boots_000") {
		t.Error("SelectEquipment should fail at limit")
	}

	// Test duplicate prevention
	if c.SelectEquipment("sword_123") {
		t.Error("SelectEquipment should prevent duplicates")
	}

	// Test count
	if c.GetEquipmentSelectionCount() != 3 {
		t.Errorf("GetEquipmentSelectionCount() = %d, want 3", c.GetEquipmentSelectionCount())
	}

	// Test IsEquipmentSelected
	if !c.IsEquipmentSelected("sword_123") {
		t.Error("IsEquipmentSelected(sword_123) should return true")
	}
	if c.IsEquipmentSelected("boots_000") {
		t.Error("IsEquipmentSelected(boots_000) should return false")
	}

	// Test deselection
	if !c.DeselectEquipment("armor_456") {
		t.Error("DeselectEquipment should succeed")
	}
	if c.GetEquipmentSelectionCount() != 2 {
		t.Errorf("GetEquipmentSelectionCount() = %d, want 2", c.GetEquipmentSelectionCount())
	}

	// Now we can add again
	if !c.SelectEquipment("boots_000") {
		t.Error("SelectEquipment should succeed after deselection")
	}
}

func TestCarryOverComponent_SkillSelection(t *testing.T) {
	c := NewCarryOverComponent()
	c.SkillSlotLimit = 2

	// Test adding skills
	if !c.SelectSkill("fireball") {
		t.Error("SelectSkill(fireball) should succeed")
	}
	if !c.SelectSkill("heal") {
		t.Error("SelectSkill(heal) should succeed")
	}

	// Should fail at limit
	if c.SelectSkill("shield") {
		t.Error("SelectSkill should fail at limit")
	}

	// Test duplicate prevention
	if c.SelectSkill("fireball") {
		t.Error("SelectSkill should prevent duplicates")
	}

	// Test count
	if c.GetSkillSelectionCount() != 2 {
		t.Errorf("GetSkillSelectionCount() = %d, want 2", c.GetSkillSelectionCount())
	}

	// Test IsSkillSelected
	if !c.IsSkillSelected("fireball") {
		t.Error("IsSkillSelected(fireball) should return true")
	}
	if c.IsSkillSelected("shield") {
		t.Error("IsSkillSelected(shield) should return false")
	}

	// Test deselection
	if !c.DeselectSkill("heal") {
		t.Error("DeselectSkill should succeed")
	}
	if c.GetSkillSelectionCount() != 1 {
		t.Errorf("GetSkillSelectionCount() = %d, want 1", c.GetSkillSelectionCount())
	}
}

func TestCarryOverComponent_CurrencyCarryOver(t *testing.T) {
	c := NewCarryOverComponent()
	c.CurrencyPercentLimit = 50.0

	// Set currency amounts
	c.SetCurrencyCarryOver("gold", 1000)
	c.SetCurrencyCarryOver("gems", 50)

	if c.GetCurrencyCarryOver("gold") != 1000 {
		t.Errorf("GetCurrencyCarryOver(gold) = %d, want 1000", c.GetCurrencyCarryOver("gold"))
	}

	// Test CalculateFinalCurrencyAmount
	tests := []struct {
		name        string
		currType    string
		totalAmount int64
		want        int64
	}{
		{
			name:        "within limit",
			currType:    "gold",
			totalAmount: 2500, // 50% = 1250, requested 1000, get 1000
			want:        1000,
		},
		{
			name:        "exceeds limit",
			currType:    "gold",
			totalAmount: 1500, // 50% = 750, requested 1000, capped to 750
			want:        750,
		},
		{
			name:        "exact limit",
			currType:    "gold",
			totalAmount: 2000, // 50% = 1000, requested 1000, get 1000
			want:        1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.CalculateFinalCurrencyAmount(tt.currType, tt.totalAmount)
			if result != tt.want {
				t.Errorf("CalculateFinalCurrencyAmount() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestCarryOverComponent_Cosmetics(t *testing.T) {
	c := NewCarryOverComponent()

	// Add cosmetics
	if !c.AddCosmetic("hat_001") {
		t.Error("AddCosmetic should succeed")
	}
	if !c.AddCosmetic("cape_002") {
		t.Error("AddCosmetic should succeed")
	}

	// Duplicate prevention
	if c.AddCosmetic("hat_001") {
		t.Error("AddCosmetic should prevent duplicates")
	}

	// Test HasCosmetic
	if !c.HasCosmetic("hat_001") {
		t.Error("HasCosmetic should return true")
	}
	if c.HasCosmetic("boots_003") {
		t.Error("HasCosmetic should return false for unknown")
	}

	// Test GetCosmetics
	cosmetics := c.GetCosmetics()
	if len(cosmetics) != 2 {
		t.Errorf("GetCosmetics() length = %d, want 2", len(cosmetics))
	}
}

func TestCarryOverComponent_Achievements(t *testing.T) {
	c := NewCarryOverComponent()

	// Add achievements
	if !c.AddAchievement("first_boss") {
		t.Error("AddAchievement should succeed")
	}
	if !c.AddAchievement("100_kills") {
		t.Error("AddAchievement should succeed")
	}

	// Duplicate prevention
	if c.AddAchievement("first_boss") {
		t.Error("AddAchievement should prevent duplicates")
	}

	// Test HasAchievement
	if !c.HasAchievement("first_boss") {
		t.Error("HasAchievement should return true")
	}
	if c.HasAchievement("speedrun") {
		t.Error("HasAchievement should return false for unknown")
	}

	// Test GetAchievements
	achievements := c.GetAchievements()
	if len(achievements) != 2 {
		t.Errorf("GetAchievements() length = %d, want 2", len(achievements))
	}
}

func TestCarryOverComponent_Locking(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 5

	// Initially unlocked
	if c.IsLocked() {
		t.Error("IsLocked should return false initially")
	}

	// Can modify before locking
	if !c.SelectEquipment("sword_1") {
		t.Error("Should be able to select before locking")
	}

	// Lock
	c.Lock()
	if !c.IsLocked() {
		t.Error("IsLocked should return true after Lock()")
	}

	// Cannot modify after locking
	if c.SelectEquipment("sword_2") {
		t.Error("Should not be able to select after locking")
	}
	if c.DeselectEquipment("sword_1") {
		t.Error("Should not be able to deselect after locking")
	}
	if c.CanSelectEquipment() {
		t.Error("CanSelectEquipment should return false after locking")
	}

	// Unlock
	c.Unlock()
	if c.IsLocked() {
		t.Error("IsLocked should return false after Unlock()")
	}

	// Can modify again
	if !c.SelectEquipment("sword_2") {
		t.Error("Should be able to select after unlocking")
	}
}

func TestCarryOverComponent_Confirmation(t *testing.T) {
	c := NewCarryOverComponent()

	// Initially not confirmed
	if c.IsConfirmed() {
		t.Error("IsConfirmed should return false initially")
	}

	// Confirm
	if !c.ConfirmSelection() {
		t.Error("ConfirmSelection should succeed")
	}
	if !c.IsConfirmed() {
		t.Error("IsConfirmed should return true after confirmation")
	}

	// Lock prevents confirmation
	c.Lock()
	if c.ConfirmSelection() {
		t.Error("ConfirmSelection should fail when locked")
	}

	// Unlock resets confirmation
	c.Unlock()
	if c.IsConfirmed() {
		t.Error("Unlock should reset confirmation")
	}
}

func TestCarryOverComponent_TransferComplete(t *testing.T) {
	c := NewCarryOverComponent()

	if c.IsTransferComplete() {
		t.Error("IsTransferComplete should return false initially")
	}

	c.MarkTransferComplete()

	if !c.IsTransferComplete() {
		t.Error("IsTransferComplete should return true after MarkTransferComplete")
	}
}

func TestCarryOverComponent_ClearSelections(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 5
	c.SkillSlotLimit = 5

	// Add various selections
	c.SelectEquipment("sword_1")
	c.SelectSkill("heal")
	c.SetCurrencyCarryOver("gold", 100)
	c.AddCosmetic("hat_1")
	c.AddAchievement("boss_1")

	// Clear
	c.ClearSelections()

	// Check equipment, skills, currency are cleared
	if c.GetEquipmentSelectionCount() != 0 {
		t.Error("Equipment should be cleared")
	}
	if c.GetSkillSelectionCount() != 0 {
		t.Error("Skills should be cleared")
	}
	if c.GetCurrencyCarryOver("gold") != 0 {
		t.Error("Currency should be cleared")
	}

	// Cosmetics and achievements should be preserved
	if !c.HasCosmetic("hat_1") {
		t.Error("Cosmetics should be preserved")
	}
	if !c.HasAchievement("boss_1") {
		t.Error("Achievements should be preserved")
	}
}

func TestCarryOverComponent_Reset(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 5
	c.SkillSlotLimit = 5

	// Add selections and lock
	c.SelectEquipment("sword_1")
	c.AddCosmetic("hat_1")
	c.Lock()
	c.MarkTransferComplete()

	// Reset
	c.Reset()

	// Check state is reset
	if c.IsLocked() {
		t.Error("IsLocked should be false after Reset")
	}
	if c.IsTransferComplete() {
		t.Error("IsTransferComplete should be false after Reset")
	}
	if c.GetEquipmentSelectionCount() != 0 {
		t.Error("Equipment should be cleared after Reset")
	}

	// Cosmetics preserved
	if !c.HasCosmetic("hat_1") {
		t.Error("Cosmetics should be preserved after Reset")
	}
}

func TestCarryOverComponent_GetSummary(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 5
	c.SkillSlotLimit = 3
	c.CurrencyPercentLimit = 60.0

	c.SelectEquipment("sword_1")
	c.SelectEquipment("armor_1")
	c.SelectSkill("heal")
	c.SetCurrencyCarryOver("gold", 100)
	c.AddCosmetic("hat_1")
	c.AddAchievement("boss_1")

	summary := c.GetSummary()

	if summary.EquipmentCount != 2 {
		t.Errorf("EquipmentCount = %d, want 2", summary.EquipmentCount)
	}
	if summary.EquipmentLimit != 5 {
		t.Errorf("EquipmentLimit = %d, want 5", summary.EquipmentLimit)
	}
	if summary.SkillCount != 1 {
		t.Errorf("SkillCount = %d, want 1", summary.SkillCount)
	}
	if summary.SkillLimit != 3 {
		t.Errorf("SkillLimit = %d, want 3", summary.SkillLimit)
	}
	if summary.CurrencyTypes != 1 {
		t.Errorf("CurrencyTypes = %d, want 1", summary.CurrencyTypes)
	}
	if summary.CurrencyLimit != 60.0 {
		t.Errorf("CurrencyLimit = %f, want 60.0", summary.CurrencyLimit)
	}
	if summary.CosmeticCount != 1 {
		t.Errorf("CosmeticCount = %d, want 1", summary.CosmeticCount)
	}
	if summary.AchievementCount != 1 {
		t.Errorf("AchievementCount = %d, want 1", summary.AchievementCount)
	}
}

func TestCarryOverComponent_Serialize(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 5
	c.SkillSlotLimit = 3
	c.CurrencyPercentLimit = 75.0

	c.SelectEquipment("sword_123")
	c.SelectSkill("fireball")
	c.SetCurrencyCarryOver("gold", 500)
	c.AddCosmetic("hat_001")
	c.AddAchievement("first_boss")
	c.ConfirmSelection()

	// Serialize
	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	c2 := NewCarryOverComponent()
	if err := c2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify state matches
	if c2.GetEquipmentSelectionCount() != 1 {
		t.Errorf("Equipment count mismatch after deserialize")
	}
	if !c2.IsEquipmentSelected("sword_123") {
		t.Error("sword_123 should be selected after deserialize")
	}
	if c2.GetSkillSelectionCount() != 1 {
		t.Error("Skill count mismatch after deserialize")
	}
	if !c2.IsSkillSelected("fireball") {
		t.Error("fireball should be selected after deserialize")
	}
	if c2.GetCurrencyCarryOver("gold") != 500 {
		t.Error("gold amount mismatch after deserialize")
	}
	if !c2.HasCosmetic("hat_001") {
		t.Error("hat_001 should be present after deserialize")
	}
	if !c2.HasAchievement("first_boss") {
		t.Error("first_boss should be present after deserialize")
	}
	if !c2.IsConfirmed() {
		t.Error("Should be confirmed after deserialize")
	}
	if c2.EquipmentSlotLimit != 5 {
		t.Errorf("EquipmentSlotLimit = %d, want 5", c2.EquipmentSlotLimit)
	}
	if c2.CurrencyPercentLimit != 75.0 {
		t.Errorf("CurrencyPercentLimit = %f, want 75.0", c2.CurrencyPercentLimit)
	}
}

func TestCarryOverComponent_DeserializeNilFields(t *testing.T) {
	// Test deserializing with nil fields
	data := []byte(`{"equipment_slot_limit":3}`)

	c := NewCarryOverComponent()
	if err := c.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Nil fields should be initialized
	if c.SelectedEquipment == nil {
		t.Error("SelectedEquipment should not be nil")
	}
	if c.CurrencyCarryOver == nil {
		t.Error("CurrencyCarryOver should not be nil")
	}
	if c.SkillsToKeep == nil {
		t.Error("SkillsToKeep should not be nil")
	}
	if c.CosmeticsUnlocked == nil {
		t.Error("CosmeticsUnlocked should not be nil")
	}
	if c.AchievementsUnlocked == nil {
		t.Error("AchievementsUnlocked should not be nil")
	}
}

func TestCarryOverComponent_ThreadSafety(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 100
	c.SkillSlotLimit = 100

	done := make(chan bool, 10)

	// Concurrent equipment selection
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				itemID := "item_" + string('A'+byte(id)) + "_" + string('0'+byte(j))
				c.SelectEquipment(itemID)
				c.IsEquipmentSelected(itemID)
			}
			done <- true
		}(i)
	}

	// Concurrent skill selection
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				skillID := "skill_" + string('A'+byte(id)) + "_" + string('0'+byte(j))
				c.SelectSkill(skillID)
				c.IsSkillSelected(skillID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no panics occurred
	summary := c.GetSummary()
	if summary.EquipmentCount == 0 && summary.SkillCount == 0 {
		t.Error("Expected some selections to succeed")
	}
}

func TestCarryOverComponent_GetSelectedEquipmentCopy(t *testing.T) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 5

	c.SelectEquipment("sword_1")
	c.SelectEquipment("armor_1")

	// Get copy
	selected := c.GetSelectedEquipment()

	// Modify copy
	selected[0] = "modified"

	// Original should be unchanged
	if c.GetSelectedEquipment()[0] == "modified" {
		t.Error("GetSelectedEquipment should return a copy, not reference")
	}
}

func TestCarryOverComponent_GetAllCurrencyCarryOverCopy(t *testing.T) {
	c := NewCarryOverComponent()

	c.SetCurrencyCarryOver("gold", 100)
	c.SetCurrencyCarryOver("gems", 50)

	// Get copy
	currencies := c.GetAllCurrencyCarryOver()

	// Modify copy
	currencies["gold"] = 9999

	// Original should be unchanged
	if c.GetCurrencyCarryOver("gold") == 9999 {
		t.Error("GetAllCurrencyCarryOver should return a copy, not reference")
	}
}

func TestCarryOverSummary_JSON(t *testing.T) {
	summary := CarryOverSummary{
		EquipmentCount:   2,
		EquipmentLimit:   5,
		SkillCount:       1,
		SkillLimit:       3,
		CurrencyTypes:    1,
		CurrencyLimit:    50.0,
		CosmeticCount:    3,
		AchievementCount: 10,
		IsLocked:         true,
		IsConfirmed:      true,
		IsComplete:       false,
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded CarryOverSummary
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.EquipmentCount != 2 {
		t.Errorf("EquipmentCount = %d, want 2", decoded.EquipmentCount)
	}
	if decoded.IsLocked != true {
		t.Error("IsLocked should be true")
	}
}

func BenchmarkCarryOverComponent_SelectEquipment(b *testing.B) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = b.N + 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		itemID := "item_" + string('0'+byte(i%10))
		c.SelectEquipment(itemID)
	}
}

func BenchmarkCarryOverComponent_IsEquipmentSelected(b *testing.B) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 100

	// Pre-populate
	for i := 0; i < 50; i++ {
		c.SelectEquipment("item_" + string('0'+byte(i%10)) + string('0'+byte((i/10)%10)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.IsEquipmentSelected("item_25")
	}
}

func BenchmarkCarryOverComponent_Serialize(b *testing.B) {
	c := NewCarryOverComponent()
	c.EquipmentSlotLimit = 10

	for i := 0; i < 5; i++ {
		c.SelectEquipment("item_" + string('0'+byte(i)))
		c.SelectSkill("skill_" + string('0'+byte(i)))
		c.AddCosmetic("cosmetic_" + string('0'+byte(i)))
	}
	c.SetCurrencyCarryOver("gold", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Serialize()
	}
}
