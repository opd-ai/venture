package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestInventorySystem_ClassRestrictions tests equipment restrictions by class.
// Phase 25.2: Class-specific equipment restrictions.
func TestInventorySystem_ClassRestrictions(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	tests := []struct {
		name             string
		class            CharacterClass
		itemRestrictions []string
		wantCanEquip     bool
	}{
		{
			name:             "warrior can equip warrior item",
			class:            ClassWarrior,
			itemRestrictions: []string{"warrior"},
			wantCanEquip:     true,
		},
		{
			name:             "warrior cannot equip mage item",
			class:            ClassWarrior,
			itemRestrictions: []string{"mage"},
			wantCanEquip:     false,
		},
		{
			name:             "mage can equip mage item",
			class:            ClassMage,
			itemRestrictions: []string{"mage"},
			wantCanEquip:     true,
		},
		{
			name:             "mage cannot equip warrior item",
			class:            ClassMage,
			itemRestrictions: []string{"warrior"},
			wantCanEquip:     false,
		},
		{
			name:             "rogue can equip no-restriction item",
			class:            ClassRogue,
			itemRestrictions: []string{},
			wantCanEquip:     true,
		},
		{
			name:             "ranger can equip multi-class item (includes ranger)",
			class:            ClassRanger,
			itemRestrictions: []string{"warrior", "ranger", "rogue"},
			wantCanEquip:     true,
		},
		{
			name:             "cleric cannot equip fighter-only item",
			class:            ClassCleric,
			itemRestrictions: []string{"warrior", "ranger", "rogue"},
			wantCanEquip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create entity with class
			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:     tt.class,
				Level:     5,
				Abilities: GetClassAbilities(tt.class),
			})
			entity.AddComponent(NewInventoryComponent(10, 100.0))
			entity.AddComponent(NewEquipmentComponent())

			world.Update(0.0) // Process entity additions

			// Create test item (weapon)
			testItem := &item.Item{
				ID:                "test_sword",
				Name:              "Test Sword",
				Type:              item.TypeWeapon,
				WeaponType:        item.WeaponSword,
				Rarity:            item.RarityCommon,
				ClassRestrictions: tt.itemRestrictions,
				Stats: item.Stats{
					Damage: 10,
					Value:  50,
					Weight: 3.0,
				},
			}

			// Add item to inventory
			comp, ok := entity.GetComponent("inventory")
			if !ok {
				t.Fatal("Failed to get inventory component")
			}
			inv := comp.(*InventoryComponent)
			inv.AddItem(testItem)

			// Try to equip item
			err := system.EquipItem(entity.ID, 0)

			if tt.wantCanEquip {
				if err != nil {
					t.Errorf("EquipItem() error = %v, want success", err)
				}

				// Verify item was equipped
				comp2, ok := entity.GetComponent("equipment")
				if !ok {
					t.Fatal("Failed to get equipment component")
				}
				equip := comp2.(*EquipmentComponent)
				equipped := equip.Slots[SlotMainHand]
				if equipped == nil {
					t.Error("Item was not equipped in main hand")
				} else if equipped.ID != testItem.ID {
					t.Errorf("Wrong item equipped, got ID %s, want %s", equipped.ID, testItem.ID)
				}

				// Verify item removed from inventory
				if len(inv.Items) != 0 {
					t.Errorf("Item still in inventory, want removed")
				}
			} else {
				if err == nil {
					t.Error("EquipItem() succeeded, want error for class restriction")
				}

				// Verify item still in inventory
				if len(inv.Items) != 1 {
					t.Errorf("Item removed from inventory, want to remain after failed equip")
				}
			}
		})
	}
}

// TestInventorySystem_DualClassRestrictions tests equipment with dual-classing.
func TestInventorySystem_DualClassRestrictions(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	// Create entity: Warrior primary, Mage secondary
	entity := world.CreateEntity()
	secondaryClass := ClassMage
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		Abilities:      GetClassAbilities(ClassWarrior),
		SecondaryClass: &secondaryClass,
		SecondaryLevel: 5,
	})
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())

	world.Update(0.0) // Process entity additions

	tests := []struct {
		name             string
		itemRestrictions []string
		wantCanEquip     bool
		reason           string
	}{
		{
			name:             "can equip warrior item (primary class)",
			itemRestrictions: []string{"warrior"},
			wantCanEquip:     true,
			reason:           "primary class",
		},
		{
			name:             "can equip mage item (secondary class)",
			itemRestrictions: []string{"mage"},
			wantCanEquip:     true,
			reason:           "secondary class",
		},
		{
			name:             "cannot equip rogue-only item",
			itemRestrictions: []string{"rogue"},
			wantCanEquip:     false,
			reason:           "not primary or secondary",
		},
		{
			name:             "can equip no-restriction item",
			itemRestrictions: []string{},
			wantCanEquip:     true,
			reason:           "no restrictions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear inventory and equipment
			comp1, ok := entity.GetComponent("inventory")
			if !ok {
				t.Fatal("Failed to get inventory component")
			}
			inv := comp1.(*InventoryComponent)

			comp2, ok := entity.GetComponent("equipment")
			if !ok {
				t.Fatal("Failed to get equipment component")
			}
			equip := comp2.(*EquipmentComponent)

			inv.Items = []*item.Item{}
			equip.Slots = make(map[EquipmentSlot]*item.Item)

			// Create test item
			testItem := &item.Item{
				ID:                "test_item",
				Name:              "Test Item",
				Type:              item.TypeWeapon,
				WeaponType:        item.WeaponStaff,
				Rarity:            item.RarityCommon,
				ClassRestrictions: tt.itemRestrictions,
				Stats: item.Stats{
					Damage: 8,
					Value:  60,
					Weight: 2.0,
				},
			}

			// Add to inventory
			inv.AddItem(testItem)

			// Try to equip
			err := system.EquipItem(entity.ID, 0)

			if tt.wantCanEquip {
				if err != nil {
					t.Errorf("EquipItem() error = %v, want success (%s)", err, tt.reason)
				}
			} else {
				if err == nil {
					t.Errorf("EquipItem() succeeded, want error (%s)", tt.reason)
				}
			}
		})
	}
}

// TestInventorySystem_ClassRestrictions_NoClassComponent tests behavior without class component.
func TestInventorySystem_ClassRestrictions_NoClassComponent(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	// Create entity without class progression component
	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())

	world.Update(0.0) // Process entity additions

	// Create restricted item
	testItem := &item.Item{
		ID:                "test_sword",
		Name:              "Test Sword",
		Type:              item.TypeWeapon,
		WeaponType:        item.WeaponSword,
		Rarity:            item.RarityCommon,
		ClassRestrictions: []string{"warrior"},
		Stats: item.Stats{
			Damage: 10,
			Value:  50,
			Weight: 3.0,
		},
	}

	// Add to inventory
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		t.Fatal("Failed to get inventory component")
	}
	inv := comp.(*InventoryComponent)
	inv.AddItem(testItem)

	// Try to equip - should succeed since entity has no class (no restrictions apply)
	err := system.EquipItem(entity.ID, 0)
	if err != nil {
		t.Errorf("EquipItem() error = %v, want success (no class component means no restrictions)", err)
	}

	// Verify equipped
	comp2, ok := entity.GetComponent("equipment")
	if !ok {
		t.Fatal("Failed to get equipment component")
	}
	equip := comp2.(*EquipmentComponent)
	if equip.Slots[SlotMainHand] == nil {
		t.Error("Item was not equipped")
	}
}
