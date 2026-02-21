// Package engine provides tests for the equipment set bonus system.
package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestEquipmentSetBonusComponent_Type(t *testing.T) {
	comp := NewEquipmentSetBonusComponent()
	if comp.Type() != "equipment_set_bonus" {
		t.Errorf("Type() = %v, want equipment_set_bonus", comp.Type())
	}
}

func TestEquipmentSetBonusComponent_NewComponent(t *testing.T) {
	comp := NewEquipmentSetBonusComponent()

	if comp.ActiveSets == nil {
		t.Error("ActiveSets should not be nil")
	}
	if len(comp.ActiveSets) != 0 {
		t.Errorf("ActiveSets should be empty, got %d", len(comp.ActiveSets))
	}
	if !comp.Dirty {
		t.Error("Dirty should be true initially")
	}
}

func TestEquipmentSetBonusComponent_TotalBonuses(t *testing.T) {
	comp := NewEquipmentSetBonusComponent()

	// Add test active sets
	comp.ActiveSets["set1"] = &ActiveSetBonus{
		SetID: "set1",
		CombinedBonus: SetBonusTier{
			DamageBonus:         10,
			DefenseBonus:        5,
			AttackSpeedBonus:    0.1,
			CriticalChanceBonus: 0.05,
			MovementSpeedBonus:  0.08,
			HealthBonus:         50,
			ManaRegenBonus:      0.15,
		},
	}
	comp.ActiveSets["set2"] = &ActiveSetBonus{
		SetID: "set2",
		CombinedBonus: SetBonusTier{
			DamageBonus:         5,
			DefenseBonus:        10,
			AttackSpeedBonus:    0.05,
			CriticalChanceBonus: 0.03,
			MovementSpeedBonus:  0.05,
			HealthBonus:         25,
			ManaRegenBonus:      0.1,
		},
	}

	tests := []struct {
		name     string
		getter   func() interface{}
		expected interface{}
	}{
		{"TotalDamageBonus", func() interface{} { return comp.GetTotalDamageBonus() }, 15},
		{"TotalDefenseBonus", func() interface{} { return comp.GetTotalDefenseBonus() }, 15},
		{"TotalAttackSpeedBonus", func() interface{} { return comp.GetTotalAttackSpeedBonus() }, 0.15},
		{"TotalCriticalChanceBonus", func() interface{} { return comp.GetTotalCriticalChanceBonus() }, 0.08},
		{"TotalMovementSpeedBonus", func() interface{} { return comp.GetTotalMovementSpeedBonus() }, 0.13},
		{"TotalHealthBonus", func() interface{} { return comp.GetTotalHealthBonus() }, 75},
		{"TotalManaRegenBonus", func() interface{} { return comp.GetTotalManaRegenBonus() }, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getter()
			switch expected := tt.expected.(type) {
			case int:
				if got.(int) != expected {
					t.Errorf("%s = %v, want %v", tt.name, got, expected)
				}
			case float64:
				if got.(float64) < expected-0.001 || got.(float64) > expected+0.001 {
					t.Errorf("%s = %v, want %v", tt.name, got, expected)
				}
			}
		})
	}
}

func TestEquipmentSetBonusComponent_HasActiveSet(t *testing.T) {
	comp := NewEquipmentSetBonusComponent()

	if comp.HasActiveSet("nonexistent") {
		t.Error("HasActiveSet should return false for nonexistent set")
	}

	comp.ActiveSets["test_set"] = &ActiveSetBonus{
		SetID:          "test_set",
		PiecesEquipped: 2,
	}

	if !comp.HasActiveSet("test_set") {
		t.Error("HasActiveSet should return true for active set")
	}

	comp.ActiveSets["empty_set"] = &ActiveSetBonus{
		SetID:          "empty_set",
		PiecesEquipped: 0,
	}

	if comp.HasActiveSet("empty_set") {
		t.Error("HasActiveSet should return false for set with 0 pieces")
	}
}

func TestEquipmentSetBonusComponent_GetActiveTierCount(t *testing.T) {
	comp := NewEquipmentSetBonusComponent()

	if comp.GetActiveTierCount("nonexistent") != 0 {
		t.Error("GetActiveTierCount should return 0 for nonexistent set")
	}

	comp.ActiveSets["test_set"] = &ActiveSetBonus{
		SetID:       "test_set",
		ActiveTiers: []int{0, 1},
	}

	if comp.GetActiveTierCount("test_set") != 2 {
		t.Errorf("GetActiveTierCount = %d, want 2", comp.GetActiveTierCount("test_set"))
	}
}

func TestEquipmentSetBonusSystem_NewSystem(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		genreID string
	}{
		{"fantasy seed", 12345, "fantasy"},
		{"scifi seed", 67890, "scifi"},
		{"horror seed", 11111, "horror"},
		{"cyberpunk seed", 22222, "cyberpunk"},
		{"postapoc seed", 33333, "postapoc"},
		{"default genre", 44444, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewEquipmentSetBonusSystem(tt.seed, tt.genreID, nil)

			if system == nil {
				t.Fatal("NewEquipmentSetBonusSystem returned nil")
			}

			if system.setRegistry == nil {
				t.Error("setRegistry should not be nil")
			}

			if len(system.setRegistry) == 0 {
				t.Error("setRegistry should have generated sets")
			}
		})
	}
}

func TestEquipmentSetBonusSystem_Determinism(t *testing.T) {
	seed := int64(99999)
	genreID := "fantasy"

	system1 := NewEquipmentSetBonusSystem(seed, genreID, nil)
	system2 := NewEquipmentSetBonusSystem(seed, genreID, nil)

	defs1 := system1.GetAllSetDefinitions()
	defs2 := system2.GetAllSetDefinitions()

	if len(defs1) != len(defs2) {
		t.Fatalf("Different number of sets: %d vs %d", len(defs1), len(defs2))
	}

	for _, def1 := range defs1 {
		def2 := system2.GetSetDefinition(def1.SetID)
		if def2 == nil {
			t.Errorf("Set %s missing in second system", def1.SetID)
			continue
		}

		if def1.SetName != def2.SetName {
			t.Errorf("Set names differ for %s: %s vs %s", def1.SetID, def1.SetName, def2.SetName)
		}

		if len(def1.Tiers) != len(def2.Tiers) {
			t.Errorf("Tier counts differ for %s: %d vs %d", def1.SetID, len(def1.Tiers), len(def2.Tiers))
		}
	}
}

func TestEquipmentSetBonusSystem_Update(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	// Get a set ID from the registry
	defs := system.GetAllSetDefinitions()
	if len(defs) == 0 {
		t.Fatal("No set definitions generated")
	}
	testSetID := defs[0].SetID

	// Create test entity with equipment
	entity := NewEntity(1)
	equipment := NewEquipmentComponent()

	// Create items that are part of the same set
	helmet := &item.Item{
		ID:        "helmet1",
		Name:      "Test Helmet",
		Type:      item.TypeArmor,
		ArmorType: item.ArmorHelmet,
		SetID:     testSetID,
		Rarity:    item.RarityRare,
	}
	chest := &item.Item{
		ID:        "chest1",
		Name:      "Test Chest",
		Type:      item.TypeArmor,
		ArmorType: item.ArmorChest,
		SetID:     testSetID,
		Rarity:    item.RarityRare,
	}

	equipment.Slots[SlotHead] = helmet
	equipment.Slots[SlotChest] = chest
	entity.AddComponent(equipment)

	// Run update
	system.Update([]*Entity{entity}, 0.016)

	// Check that set bonus component was created
	setBonusComp, hasComp := entity.GetComponent("equipment_set_bonus")
	if !hasComp || setBonusComp == nil {
		t.Fatal("EquipmentSetBonusComponent not created")
	}

	setBonus := setBonusComp.(*EquipmentSetBonusComponent)

	// Check that the set is active
	if !setBonus.HasActiveSet(testSetID) {
		t.Errorf("Set %s should be active with 2 pieces", testSetID)
	}

	// Check piece count
	activeSet := setBonus.ActiveSets[testSetID]
	if activeSet == nil {
		t.Fatal("ActiveSet should not be nil")
	}

	if activeSet.PiecesEquipped != 2 {
		t.Errorf("PiecesEquipped = %d, want 2", activeSet.PiecesEquipped)
	}

	// Should have 2-piece bonus active
	if len(activeSet.ActiveTiers) != 1 {
		t.Errorf("ActiveTiers count = %d, want 1", len(activeSet.ActiveTiers))
	}
}

func TestEquipmentSetBonusSystem_UpdateWithMorePieces(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	defs := system.GetAllSetDefinitions()
	if len(defs) == 0 {
		t.Fatal("No set definitions generated")
	}
	testSetID := defs[0].SetID

	entity := NewEntity(1)
	equipment := NewEquipmentComponent()

	// Equip 4 pieces of the same set
	equipment.Slots[SlotHead] = &item.Item{
		ID: "h1", Type: item.TypeArmor, ArmorType: item.ArmorHelmet, SetID: testSetID,
	}
	equipment.Slots[SlotChest] = &item.Item{
		ID: "c1", Type: item.TypeArmor, ArmorType: item.ArmorChest, SetID: testSetID,
	}
	equipment.Slots[SlotLegs] = &item.Item{
		ID: "l1", Type: item.TypeArmor, ArmorType: item.ArmorLegs, SetID: testSetID,
	}
	equipment.Slots[SlotBoots] = &item.Item{
		ID: "b1", Type: item.TypeArmor, ArmorType: item.ArmorBoots, SetID: testSetID,
	}

	entity.AddComponent(equipment)
	system.Update([]*Entity{entity}, 0.016)

	setBonusComp, _ := entity.GetComponent("equipment_set_bonus")
	setBonus := setBonusComp.(*EquipmentSetBonusComponent)
	activeSet := setBonus.ActiveSets[testSetID]

	if activeSet.PiecesEquipped != 4 {
		t.Errorf("PiecesEquipped = %d, want 4", activeSet.PiecesEquipped)
	}

	// Should have 2-piece AND 4-piece bonuses active
	if len(activeSet.ActiveTiers) != 2 {
		t.Errorf("ActiveTiers count = %d, want 2", len(activeSet.ActiveTiers))
	}

	// Bonuses should be cumulative
	if activeSet.CombinedBonus.DamageBonus < 0 {
		t.Error("CombinedBonus.DamageBonus should be positive")
	}
}

func TestEquipmentSetBonusSystem_NoUpdateWhenUnchanged(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	defs := system.GetAllSetDefinitions()
	testSetID := defs[0].SetID

	entity := NewEntity(1)
	equipment := NewEquipmentComponent()
	equipment.Slots[SlotHead] = &item.Item{
		ID: "h1", Type: item.TypeArmor, ArmorType: item.ArmorHelmet, SetID: testSetID,
	}
	entity.AddComponent(equipment)

	// First update
	system.Update([]*Entity{entity}, 0.016)
	setBonusComp, _ := entity.GetComponent("equipment_set_bonus")
	setBonus := setBonusComp.(*EquipmentSetBonusComponent)
	originalHash := setBonus.LastEquipmentHash

	// Second update with no changes
	system.Update([]*Entity{entity}, 0.016)

	// Hash should remain the same
	if setBonus.LastEquipmentHash != originalHash {
		t.Error("Hash should not change when equipment is unchanged")
	}
}

func TestEquipmentSetBonusSystem_AssignSetToItem(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)
	rng := rand.New(rand.NewSource(54321))

	tests := []struct {
		name       string
		item       *item.Item
		canAssign  bool
		checkSetID bool
	}{
		{
			name:       "legendary weapon can be assigned",
			item:       &item.Item{ID: "w1", Type: item.TypeWeapon, Rarity: item.RarityLegendary},
			canAssign:  true,
			checkSetID: true,
		},
		{
			name:       "common item cannot be assigned",
			item:       &item.Item{ID: "c1", Type: item.TypeArmor, Rarity: item.RarityCommon},
			canAssign:  false,
			checkSetID: false,
		},
		{
			name:       "consumable cannot be assigned",
			item:       &item.Item{ID: "con1", Type: item.TypeConsumable, Rarity: item.RarityLegendary},
			canAssign:  false,
			checkSetID: false,
		},
		{
			name:      "nil item",
			item:      nil,
			canAssign: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset RNG for determinism
			rng = rand.New(rand.NewSource(54321))

			// Try multiple times for probabilistic assignment
			assigned := false
			for i := 0; i < 100; i++ {
				if tt.item != nil {
					tt.item.SetID = "" // Reset
				}
				if system.AssignSetToItem(tt.item, rng) {
					assigned = true
					break
				}
			}

			if tt.canAssign && !assigned && tt.item != nil && tt.item.Rarity == item.RarityLegendary {
				t.Error("Expected assignment for legendary item")
			}

			if !tt.canAssign && assigned {
				t.Error("Did not expect assignment")
			}

			if tt.checkSetID && assigned && tt.item.SetID == "" {
				t.Error("SetID should be assigned")
			}
		})
	}
}

func TestEquipmentSetBonusSystem_RegisterSet(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	customSet := &EquipmentSetDefinition{
		SetID:       "custom_test_set",
		SetName:     "Test Set",
		Description: "A custom test set",
		TotalPieces: 3,
		Tiers: []SetBonusTier{
			{PiecesRequired: 2, DamageBonus: 100},
			{PiecesRequired: 3, DamageBonus: 200},
		},
	}

	system.RegisterSet(customSet)

	retrieved := system.GetSetDefinition("custom_test_set")
	if retrieved == nil {
		t.Fatal("Custom set should be registered")
	}

	if retrieved.SetName != "Test Set" {
		t.Errorf("SetName = %s, want Test Set", retrieved.SetName)
	}

	if len(retrieved.Tiers) != 2 {
		t.Errorf("Tiers count = %d, want 2", len(retrieved.Tiers))
	}
}

func TestEquipmentSetBonusSystem_GetRandomSetID(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)
	rng := rand.New(rand.NewSource(11111))

	// Get multiple random IDs
	seenIDs := make(map[string]bool)
	for i := 0; i < 50; i++ {
		setID := system.GetRandomSetID(rng)
		if setID == "" {
			t.Error("GetRandomSetID returned empty string")
		}
		seenIDs[setID] = true
	}

	// Should see multiple different sets
	if len(seenIDs) < 2 {
		t.Error("Expected variation in random set IDs")
	}
}

func TestEquipmentSetBonusSystem_EmptyEntitySlice(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	// Should not panic
	system.Update([]*Entity{}, 0.016)
	system.Update(nil, 0.016)
}

func TestEquipmentSetBonusSystem_EntityWithoutEquipment(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	entity := NewEntity(2)
	// No equipment component added

	// Should not panic
	system.Update([]*Entity{entity}, 0.016)

	// Should not have set bonus component
	if _, has := entity.GetComponent("equipment_set_bonus"); has {
		t.Error("Entity without equipment should not have set bonus component")
	}
}

func TestEquipmentSetBonusSystem_MixedSets(t *testing.T) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	defs := system.GetAllSetDefinitions()
	if len(defs) < 2 {
		t.Skip("Need at least 2 sets for this test")
	}

	setID1 := defs[0].SetID
	setID2 := defs[1].SetID

	entity := NewEntity(3)
	equipment := NewEquipmentComponent()

	// 2 pieces from set 1
	equipment.Slots[SlotHead] = &item.Item{
		ID: "h1", Type: item.TypeArmor, ArmorType: item.ArmorHelmet, SetID: setID1,
	}
	equipment.Slots[SlotChest] = &item.Item{
		ID: "c1", Type: item.TypeArmor, ArmorType: item.ArmorChest, SetID: setID1,
	}

	// 2 pieces from set 2
	equipment.Slots[SlotLegs] = &item.Item{
		ID: "l1", Type: item.TypeArmor, ArmorType: item.ArmorLegs, SetID: setID2,
	}
	equipment.Slots[SlotBoots] = &item.Item{
		ID: "b1", Type: item.TypeArmor, ArmorType: item.ArmorBoots, SetID: setID2,
	}

	entity.AddComponent(equipment)
	system.Update([]*Entity{entity}, 0.016)

	setBonusComp, _ := entity.GetComponent("equipment_set_bonus")
	setBonus := setBonusComp.(*EquipmentSetBonusComponent)

	// Both sets should be active
	if !setBonus.HasActiveSet(setID1) {
		t.Errorf("Set %s should be active", setID1)
	}
	if !setBonus.HasActiveSet(setID2) {
		t.Errorf("Set %s should be active", setID2)
	}

	// Each should have 2 pieces
	if setBonus.ActiveSets[setID1].PiecesEquipped != 2 {
		t.Errorf("Set1 pieces = %d, want 2", setBonus.ActiveSets[setID1].PiecesEquipped)
	}
	if setBonus.ActiveSets[setID2].PiecesEquipped != 2 {
		t.Errorf("Set2 pieces = %d, want 2", setBonus.ActiveSets[setID2].PiecesEquipped)
	}
}

func TestSetBonusTier_Fields(t *testing.T) {
	tier := SetBonusTier{
		PiecesRequired:      4,
		DamageBonus:         20,
		DefenseBonus:        15,
		AttackSpeedBonus:    0.15,
		CriticalChanceBonus: 0.08,
		MovementSpeedBonus:  0.1,
		HealthBonus:         100,
		ManaRegenBonus:      0.2,
		SpecialEffect:       "Test effect",
	}

	if tier.PiecesRequired != 4 {
		t.Errorf("PiecesRequired = %d, want 4", tier.PiecesRequired)
	}
	if tier.DamageBonus != 20 {
		t.Errorf("DamageBonus = %d, want 20", tier.DamageBonus)
	}
	if tier.SpecialEffect != "Test effect" {
		t.Errorf("SpecialEffect = %s, want Test effect", tier.SpecialEffect)
	}
}

func TestEquipmentSetDefinition_Fields(t *testing.T) {
	def := EquipmentSetDefinition{
		SetID:       "test_set",
		SetName:     "Test Set Name",
		Description: "A test description",
		TotalPieces: 6,
		Tiers: []SetBonusTier{
			{PiecesRequired: 2},
			{PiecesRequired: 4},
			{PiecesRequired: 6},
		},
		GenreID: "fantasy",
	}

	if def.SetID != "test_set" {
		t.Errorf("SetID = %s, want test_set", def.SetID)
	}
	if def.TotalPieces != 6 {
		t.Errorf("TotalPieces = %d, want 6", def.TotalPieces)
	}
	if len(def.Tiers) != 3 {
		t.Errorf("Tiers count = %d, want 3", len(def.Tiers))
	}
}

func BenchmarkEquipmentSetBonusSystem_Update(b *testing.B) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)
	defs := system.GetAllSetDefinitions()
	testSetID := defs[0].SetID

	// Create entities with equipment
	entities := make([]*Entity, 100)
	for i := range entities {
		entity := NewEntity(uint64(i))
		equipment := NewEquipmentComponent()
		equipment.Slots[SlotHead] = &item.Item{
			ID: "h", Type: item.TypeArmor, ArmorType: item.ArmorHelmet, SetID: testSetID,
		}
		equipment.Slots[SlotChest] = &item.Item{
			ID: "c", Type: item.TypeArmor, ArmorType: item.ArmorChest, SetID: testSetID,
		}
		entity.AddComponent(equipment)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkEquipmentSetBonusSystem_EquipmentHashCalculation(b *testing.B) {
	system := NewEquipmentSetBonusSystem(12345, "fantasy", nil)

	equipment := NewEquipmentComponent()
	equipment.Slots[SlotHead] = &item.Item{ID: "h1", SetID: "set1"}
	equipment.Slots[SlotChest] = &item.Item{ID: "c1", SetID: "set1"}
	equipment.Slots[SlotLegs] = &item.Item{ID: "l1", SetID: "set1"}
	equipment.Slots[SlotBoots] = &item.Item{ID: "b1", SetID: "set1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.calculateEquipmentHash(equipment)
	}
}
