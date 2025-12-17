// Package engine provides tests for the CarryOver system.
// Tests cover preparation, selection, transfer, and callbacks.
//
// Phase 112: Carry-Over System Tests
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewCarryOverSystem(t *testing.T) {
	sys := NewCarryOverSystem(nil)
	if sys == nil {
		t.Fatal("NewCarryOverSystem() returned nil")
	}
}

func TestCarryOverSystem_PrepareForNGPlus(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	// Add NG+ component with cycle 2 and properly set carry-over benefits
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 2
	// Manually set the carry-over benefits that would have been earned
	// through completing cycles (base 3 slots + 1 per cycle)
	ngp.CarryOverSlots = 5              // 3 base + 2 cycles
	ngp.CurrencyCarryOverPercent = 60.0 // 50% base + 5% per cycle * 2
	entity.AddComponent(ngp)

	// Add skill book with some skills
	skillBook := NewSkillBookComponent()
	skillBook.LearnedSkills["fireball"] = &LearnedSkill{Name: "Fireball", Level: 3}
	skillBook.LearnedSkills["heal"] = &LearnedSkill{Name: "Heal", Level: 2}
	skillBook.LearnedSkills["shield"] = &LearnedSkill{Name: "Shield", Level: 1}
	entity.AddComponent(skillBook)

	// Prepare for NG+
	err := sys.PrepareForNGPlus(entity)
	if err != nil {
		t.Fatalf("PrepareForNGPlus() error = %v", err)
	}

	// Check carry-over component was created
	carryComp, ok := entity.GetComponent("carryover")
	if !ok {
		t.Fatal("Carry-over component not created")
	}

	carry, ok := carryComp.(*CarryOverComponent)
	if !ok {
		t.Fatal("Invalid carry-over component type")
	}

	// Check limits were set based on NG+ component values
	if carry.GetEquipmentSlotLimit() != 5 {
		t.Errorf("EquipmentSlotLimit = %d, want 5", carry.GetEquipmentSlotLimit())
	}
	if carry.GetCurrencyPercentLimit() != 60.0 {
		t.Errorf("CurrencyPercentLimit = %f, want 60.0", carry.GetCurrencyPercentLimit())
	}
}

func TestCarryOverSystem_PrepareForNGPlus_NoNGPComponent(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	// No NG+ component, should still work with defaults
	err := sys.PrepareForNGPlus(entity)
	if err != nil {
		t.Fatalf("PrepareForNGPlus() error = %v", err)
	}

	// Check carry-over component was created with defaults
	carryComp, ok := entity.GetComponent("carryover")
	if !ok {
		t.Fatal("Carry-over component not created")
	}

	carry, ok := carryComp.(*CarryOverComponent)
	if !ok {
		t.Fatal("Invalid carry-over component type")
	}

	if carry.GetEquipmentSlotLimit() != 3 {
		t.Errorf("EquipmentSlotLimit = %d, want 3 (default)", carry.GetEquipmentSlotLimit())
	}
}

func TestCarryOverSystem_SelectEquipmentFromInventory(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	// Add carry-over component
	carry := NewCarryOverComponent()
	carry.EquipmentSlotLimit = 3
	entity.AddComponent(carry)

	// Add inventory with items
	inv := NewInventoryComponent(10, 100.0)
	sword := &item.Item{
		Name: "Flame Sword",
		Type: item.TypeWeapon,
		Seed: 12345,
	}
	armor := &item.Item{
		Name: "Iron Armor",
		Type: item.TypeArmor,
		Seed: 67890,
	}
	potion := &item.Item{
		Name: "Health Potion",
		Type: item.TypeConsumable,
		Seed: 11111,
	}
	inv.Items = append(inv.Items, sword, armor, potion)
	entity.AddComponent(inv)

	// Select first item
	if !sys.SelectEquipmentFromInventory(entity, 0) {
		t.Error("SelectEquipmentFromInventory(0) should succeed")
	}

	// Check selection
	if carry.GetEquipmentSelectionCount() != 1 {
		t.Errorf("Equipment count = %d, want 1", carry.GetEquipmentSelectionCount())
	}

	// Select second item
	if !sys.SelectEquipmentFromInventory(entity, 1) {
		t.Error("SelectEquipmentFromInventory(1) should succeed")
	}

	// Invalid index should fail
	if sys.SelectEquipmentFromInventory(entity, 10) {
		t.Error("SelectEquipmentFromInventory(10) should fail for invalid index")
	}

	// Negative index should fail
	if sys.SelectEquipmentFromInventory(entity, -1) {
		t.Error("SelectEquipmentFromInventory(-1) should fail")
	}
}

func TestCarryOverSystem_SetCurrencyAmount(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	entity.AddComponent(carry)

	sys.SetCurrencyAmount(entity, "gold", 500)
	sys.SetCurrencyAmount(entity, "gems", 25)

	if carry.GetCurrencyCarryOver("gold") != 500 {
		t.Errorf("gold = %d, want 500", carry.GetCurrencyCarryOver("gold"))
	}
	if carry.GetCurrencyCarryOver("gems") != 25 {
		t.Errorf("gems = %d, want 25", carry.GetCurrencyCarryOver("gems"))
	}
}

func TestCarryOverSystem_SetGoldCarryOver(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	entity.AddComponent(carry)

	sys.SetGoldCarryOver(entity, 1000)

	if carry.GetCurrencyCarryOver("gold") != 1000 {
		t.Errorf("gold = %d, want 1000", carry.GetCurrencyCarryOver("gold"))
	}
}

func TestCarryOverSystem_SelectSkill(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	carry.SkillSlotLimit = 2
	entity.AddComponent(carry)

	if !sys.SelectSkill(entity, "fireball") {
		t.Error("SelectSkill(fireball) should succeed")
	}
	if !sys.SelectSkill(entity, "heal") {
		t.Error("SelectSkill(heal) should succeed")
	}

	// At limit
	if sys.SelectSkill(entity, "shield") {
		t.Error("SelectSkill(shield) should fail at limit")
	}

	// Deselect
	if !sys.DeselectSkill(entity, "heal") {
		t.Error("DeselectSkill(heal) should succeed")
	}

	// Now can add again
	if !sys.SelectSkill(entity, "shield") {
		t.Error("SelectSkill(shield) should succeed after deselect")
	}
}

func TestCarryOverSystem_ConfirmSelection(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	entity.AddComponent(carry)

	// Track callback
	callbackCalled := false
	sys.SetOnSelectionComplete(func(entityID uint64) {
		callbackCalled = true
	})

	if !sys.ConfirmSelection(entity) {
		t.Error("ConfirmSelection should succeed")
	}

	if !carry.IsConfirmed() {
		t.Error("Carry should be confirmed")
	}

	if !callbackCalled {
		t.Error("OnSelectionComplete callback should have been called")
	}
}

func TestCarryOverSystem_LockAndTransfer(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	entity.AddComponent(carry)

	// Cannot lock without confirmation
	if sys.LockAndTransfer(entity) {
		t.Error("LockAndTransfer should fail without confirmation")
	}

	// Confirm first
	carry.ConfirmSelection()

	// Now can lock
	if !sys.LockAndTransfer(entity) {
		t.Error("LockAndTransfer should succeed after confirmation")
	}

	if !carry.IsLocked() {
		t.Error("Carry should be locked")
	}
}

func TestCarryOverSystem_Update_ProcessesTransfer(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	carry.EquipmentSlotLimit = 5
	entity.AddComponent(carry)

	// Add inventory
	inv := NewInventoryComponent(10, 100.0)
	sword := &item.Item{
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Seed: 12345,
	}
	inv.Items = append(inv.Items, sword)
	inv.Gold = 1000
	entity.AddComponent(inv)

	// Select items
	sys.SelectEquipmentFromInventory(entity, 0)
	sys.SetGoldCarryOver(entity, 400)

	// Confirm and lock
	carry.ConfirmSelection()
	carry.Lock()

	// Track transfer callback
	transferCalled := false
	sys.SetOnTransferComplete(func(entityID uint64, summary CarryOverSummary) {
		transferCalled = true
	})

	// Update should process transfer
	sys.Update([]*Entity{entity}, 0.016)

	if !carry.IsTransferComplete() {
		t.Error("Transfer should be complete after Update")
	}

	if !transferCalled {
		t.Error("OnTransferComplete callback should have been called")
	}
}

func TestCarryOverSystem_CancelCarryOver(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	carry.EquipmentSlotLimit = 5
	entity.AddComponent(carry)

	// Make selections
	carry.SelectEquipment("sword_1")
	carry.SelectSkill("heal")

	// Cancel
	sys.CancelCarryOver(entity)

	if carry.GetEquipmentSelectionCount() != 0 {
		t.Error("Equipment should be cleared after cancel")
	}
	if carry.GetSkillSelectionCount() != 0 {
		t.Error("Skills should be cleared after cancel")
	}
}

func TestCarryOverSystem_GetAvailableEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	// Add carry-over component
	carry := NewCarryOverComponent()
	carry.EquipmentSlotLimit = 5
	entity.AddComponent(carry)

	// Add inventory with equippable and non-equippable items
	inv := NewInventoryComponent(10, 100.0)
	sword := &item.Item{
		Name:   "Flame Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityRare,
		Stats:  item.Stats{RequiredLevel: 5},
		Seed:   12345,
	}
	armor := &item.Item{
		Name:   "Iron Armor",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{RequiredLevel: 1},
		Seed:   67890,
	}
	potion := &item.Item{
		Name: "Health Potion",
		Type: item.TypeConsumable,
		Seed: 11111,
	}
	inv.Items = append(inv.Items, sword, armor, potion)
	entity.AddComponent(inv)

	// Select first item
	sys.SelectEquipmentFromInventory(entity, 0)

	// Get available equipment
	options := sys.GetAvailableEquipment(entity)

	// Should have 2 options (weapon and armor, not consumable)
	if len(options) != 2 {
		t.Errorf("len(options) = %d, want 2", len(options))
	}

	// First option (sword) should be selected
	found := false
	for _, opt := range options {
		if opt.Name == "Flame Sword" {
			found = true
			if !opt.Selected {
				t.Error("Flame Sword should be marked as selected")
			}
			if opt.Type != "weapon" {
				t.Errorf("Type = %s, want weapon", opt.Type)
			}
			if opt.Rarity != "rare" {
				t.Errorf("Rarity = %s, want rare", opt.Rarity)
			}
		}
	}
	if !found {
		t.Error("Flame Sword not found in options")
	}
}

func TestCarryOverSystem_GetAvailableSkills(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	// Add carry-over component
	carry := NewCarryOverComponent()
	carry.SkillSlotLimit = 5
	entity.AddComponent(carry)

	// Add skill book
	skillBook := NewSkillBookComponent()
	skillBook.LearnedSkills["fireball"] = &LearnedSkill{Name: "Fireball", Type: "offensive", Level: 3}
	skillBook.LearnedSkills["heal"] = &LearnedSkill{Name: "Heal", Type: "support", Level: 2}
	entity.AddComponent(skillBook)

	// Select one skill
	sys.SelectSkill(entity, "fireball")

	// Get available skills
	options := sys.GetAvailableSkills(entity)

	if len(options) != 2 {
		t.Errorf("len(options) = %d, want 2", len(options))
	}

	// Check fireball is marked as selected
	for _, opt := range options {
		if opt.SkillID == "fireball" && !opt.Selected {
			t.Error("fireball should be marked as selected")
		}
		if opt.SkillID == "heal" && opt.Selected {
			t.Error("heal should not be marked as selected")
		}
	}
}

func TestCarryOverSystem_GetCarryOverStatus(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	// No component
	status := sys.GetCarryOverStatus(entity)
	if status != nil {
		t.Error("GetCarryOverStatus should return nil when no component")
	}

	// Add component
	carry := NewCarryOverComponent()
	carry.EquipmentSlotLimit = 5
	carry.SelectEquipment("sword_1")
	entity.AddComponent(carry)

	status = sys.GetCarryOverStatus(entity)
	if status == nil {
		t.Fatal("GetCarryOverStatus should not return nil")
	}

	if status.EquipmentCount != 1 {
		t.Errorf("EquipmentCount = %d, want 1", status.EquipmentCount)
	}
	if status.EquipmentLimit != 5 {
		t.Errorf("EquipmentLimit = %d, want 5", status.EquipmentLimit)
	}
}

func TestCarryOverSystem_SetScaleItemLevel(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	// Default is true
	if !sys.scaleItemLevel {
		t.Error("scaleItemLevel should be true by default")
	}

	sys.SetScaleItemLevel(false)

	if sys.scaleItemLevel {
		t.Error("scaleItemLevel should be false after SetScaleItemLevel(false)")
	}
}

func TestCarryOverSystem_Callbacks(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()

	carry := NewCarryOverComponent()
	carry.EquipmentSlotLimit = 5
	entity.AddComponent(carry)

	// Add inventory
	inv := NewInventoryComponent(10, 100.0)
	sword := &item.Item{
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Seed: 12345,
	}
	inv.Items = append(inv.Items, sword)
	inv.Gold = 1000
	entity.AddComponent(inv)

	// Select and set currency
	sys.SelectEquipmentFromInventory(entity, 0)
	sys.SetGoldCarryOver(entity, 400)
	sys.SelectSkill(entity, "test_skill")

	// Set up callbacks
	equipmentTransferred := false
	currencyTransferred := false
	skillsTransferred := false

	sys.SetOnEquipmentTransfer(func(entityID uint64, items []*item.Item) {
		equipmentTransferred = true
	})
	sys.SetOnCurrencyTransfer(func(entityID uint64, currency map[string]int64) {
		currencyTransferred = true
	})
	sys.SetOnSkillsTransfer(func(entityID uint64, skills []string) {
		skillsTransferred = true
	})

	// Confirm, lock, and update
	carry.ConfirmSelection()
	carry.Lock()
	sys.Update([]*Entity{entity}, 0.016)

	if !equipmentTransferred {
		t.Error("Equipment transfer callback should have been called")
	}
	if !currencyTransferred {
		t.Error("Currency transfer callback should have been called")
	}
	if !skillsTransferred {
		t.Error("Skills transfer callback should have been called")
	}
}

func TestCarryOverSystem_ApplyCarryOverToNewCharacter(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	// Create carry-over data
	carryData := NewCarryOverComponent()
	carryData.AddCosmetic("hat_001")
	carryData.AddCosmetic("cape_002")
	carryData.AddAchievement("first_boss")
	carryData.MarkTransferComplete()

	// Create new character entity
	entity := world.CreateEntity()

	// Add cosmetic component
	cosmetic := NewCosmeticComponent()
	entity.AddComponent(cosmetic)

	// Add achievement component (using basic struct init)
	achieve := &AchievementComponent{
		Achievements:     []Achievement{},
		UniqueExpression: make(map[ExpressionType]int),
	}
	entity.AddComponent(achieve)

	// Apply carry-over
	err := sys.ApplyCarryOverToNewCharacter(entity, carryData)
	if err != nil {
		t.Fatalf("ApplyCarryOverToNewCharacter() error = %v", err)
	}

	// Check cosmetics were applied
	if !cosmetic.HasCosmetic("hat_001") {
		t.Error("hat_001 should be unlocked")
	}
	if !cosmetic.HasCosmetic("cape_002") {
		t.Error("cape_002 should be unlocked")
	}

	// Check achievements were applied
	if !achieve.IsUnlocked("first_boss") {
		t.Error("first_boss should be unlocked")
	}
}

func TestCarryOverSystem_ApplyCarryOverToNewCharacter_NotComplete(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	// Create carry-over data that's not complete
	carryData := NewCarryOverComponent()
	carryData.AddCosmetic("hat_001")
	// Not marked as complete

	entity := world.CreateEntity()
	cosmetic := NewCosmeticComponent()
	entity.AddComponent(cosmetic)

	// Should not apply anything
	err := sys.ApplyCarryOverToNewCharacter(entity, carryData)
	if err != nil {
		t.Fatalf("ApplyCarryOverToNewCharacter() error = %v", err)
	}

	if cosmetic.HasCosmetic("hat_001") {
		t.Error("Should not apply carry-over when not complete")
	}
}

func TestCarryOverSystem_NoEntityComponent(t *testing.T) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()
	// No carry-over component

	// These should not panic
	sys.SetCurrencyAmount(entity, "gold", 100)
	sys.CancelCarryOver(entity)

	if sys.ConfirmSelection(entity) {
		t.Error("ConfirmSelection should fail without component")
	}
	if sys.LockAndTransfer(entity) {
		t.Error("LockAndTransfer should fail without component")
	}
	if sys.SelectSkill(entity, "test") {
		t.Error("SelectSkill should fail without component")
	}
	if sys.DeselectSkill(entity, "test") {
		t.Error("DeselectSkill should fail without component")
	}
	if sys.DeselectEquipment(entity, "test") {
		t.Error("DeselectEquipment should fail without component")
	}
}

func TestFormatSeed(t *testing.T) {
	tests := []struct {
		seed int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{12345, "12345"},
		{-12345, "-12345"},
		{9876543210, "9876543210"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSeed(tt.seed)
			if got != tt.want {
				t.Errorf("formatSeed(%d) = %q, want %q", tt.seed, got, tt.want)
			}
		})
	}
}

func BenchmarkCarryOverSystem_PrepareForNGPlus(b *testing.B) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()
	ngp := NewNewGamePlusComponent()
	entity.AddComponent(ngp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.PrepareForNGPlus(entity)
	}
}

func BenchmarkCarryOverSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewCarryOverSystem(world)

	entity := world.CreateEntity()
	carry := NewCarryOverComponent()
	entity.AddComponent(carry)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
