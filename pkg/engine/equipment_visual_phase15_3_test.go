// Package engine provides Phase 15.3 equipment visual refinement tests.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestEquipmentVisual_Phase15_3_MaterialTypes tests material type detection.
func TestEquipmentVisual_Phase15_3_MaterialTypes(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	tests := []struct {
		name         string
		weaponType   item.WeaponType
		armorType    item.ArmorType
		tags         []string
		wantMaterial sprites.MaterialType
	}{
		{"metal sword", item.WeaponSword, item.ArmorChest, []string{"metal", "heavy"}, sprites.MaterialMetal},
		{"wooden bow", item.WeaponBow, item.ArmorChest, []string{"wood", "ranged"}, sprites.MaterialWood},
		{"leather armor", item.WeaponSword, item.ArmorLegs, []string{"leather"}, sprites.MaterialLeather},
		{"cloth robe", item.WeaponStaff, item.ArmorChest, []string{"cloth", "magical"}, sprites.MaterialCloth},
		{"crystal staff", item.WeaponStaff, item.ArmorChest, []string{"crystal", "magical"}, sprites.MaterialCrystal},
		{"energy gun", item.WeaponGun, item.ArmorChest, []string{"energy", "advanced"}, sprites.MaterialEnergy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test item
			testItem := &item.Item{
				ID:         "test-item",
				Name:       "Test Item",
				Type:       item.TypeWeapon,
				WeaponType: tt.weaponType,
				ArmorType:  tt.armorType,
				Tags:       tt.tags,
				Seed:       12345,
			}

			// Get material type
			material := sys.getMaterialType(testItem, "weapon", "fantasy")

			if material != tt.wantMaterial {
				t.Errorf("getMaterialType() = %v, want %v", material, tt.wantMaterial)
			}
		})
	}
}

// TestEquipmentVisual_Phase15_3_DamageStates tests damage state from durability.
func TestEquipmentVisual_Phase15_3_DamageStates(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	tests := []struct {
		name       string
		durability int
		maxDur     int
		wantState  sprites.DamageState
	}{
		{"pristine 100%", 100, 100, sprites.DamageStatePristine},
		{"worn 75%", 75, 100, sprites.DamageStateWorn},
		{"damaged 40%", 40, 100, sprites.DamageStateDamaged},
		{"broken 10%", 10, 100, sprites.DamageStateBroken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test item with specific durability
			testItem := &item.Item{
				ID:   "test-item",
				Name: "Test Item",
				Type: item.TypeWeapon,
				Stats: item.Stats{
					Durability:    tt.durability,
					DurabilityMax: tt.maxDur,
				},
				Seed: 12345,
			}

			// Build equipment visual
			visual := sys.buildEquipmentVisual("weapon", testItem.ID, testItem.Seed, sprites.LayerWeapon, testItem, "fantasy")

			if visual.DamageState != tt.wantState {
				t.Errorf("DamageState = %v, want %v", visual.DamageState, tt.wantState)
			}
		})
	}
}

// TestEquipmentVisual_Phase15_3_EnchantmentGlow tests enchantment effects from rarity.
func TestEquipmentVisual_Phase15_3_EnchantmentGlow(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	tests := []struct {
		name       string
		rarity     item.Rarity
		wantActive bool
		wantColor  string
	}{
		{"common no glow", item.RarityCommon, false, "white"},
		{"uncommon green", item.RarityUncommon, true, "green"},
		{"rare blue", item.RarityRare, true, "blue"},
		{"epic purple", item.RarityEpic, true, "purple"},
		{"legendary gold", item.RarityLegendary, true, "gold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test item with specific rarity
			testItem := &item.Item{
				ID:     "test-item",
				Name:   "Test Item",
				Type:   item.TypeWeapon,
				Rarity: tt.rarity,
				Seed:   12345,
			}

			// Build equipment visual
			visual := sys.buildEquipmentVisual("weapon", testItem.ID, testItem.Seed, sprites.LayerWeapon, testItem, "fantasy")

			if visual.Enchantment.Active != tt.wantActive {
				t.Errorf("Enchantment.Active = %v, want %v", visual.Enchantment.Active, tt.wantActive)
			}
			if visual.Enchantment.Color != tt.wantColor {
				t.Errorf("Enchantment.Color = %v, want %v", visual.Enchantment.Color, tt.wantColor)
			}
			if tt.wantActive && visual.Enchantment.Intensity <= 0 {
				t.Errorf("Expected positive intensity for active enchantment, got %v", visual.Enchantment.Intensity)
			}
		})
	}
}

// TestEquipmentVisual_Phase15_3_DetailLevel tests detail level from rarity.
func TestEquipmentVisual_Phase15_3_DetailLevel(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	tests := []struct {
		name      string
		rarity    item.Rarity
		wantLevel float64
	}{
		{"common low", item.RarityCommon, 0.3},
		{"uncommon medium-low", item.RarityUncommon, 0.4},
		{"rare medium", item.RarityRare, 0.6},
		{"epic high", item.RarityEpic, 0.8},
		{"legendary max", item.RarityLegendary, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test item with specific rarity
			testItem := &item.Item{
				ID:     "test-item",
				Name:   "Test Item",
				Type:   item.TypeWeapon,
				Rarity: tt.rarity,
				Seed:   12345,
			}

			// Build equipment visual
			visual := sys.buildEquipmentVisual("weapon", testItem.ID, testItem.Seed, sprites.LayerWeapon, testItem, "fantasy")

			if visual.DetailLevel != tt.wantLevel {
				t.Errorf("DetailLevel = %v, want %v", visual.DetailLevel, tt.wantLevel)
			}
		})
	}
}

// TestEquipmentVisual_Phase15_3_Integration tests full integration with entity.
func TestEquipmentVisual_Phase15_3_Integration(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	// Create entity with equipment
	entity := NewEntity(1)
	
	// Add genre component
	genreComp := &GenreComponent{}
	genreComp.GenreID =("fantasy")
	entity.AddComponent(genreComp)

	// Add sprite component
	spriteComp := &EbitenSprite{
		Width:  32,
		Height: 32,
	}
	entity.AddComponent(spriteComp)

	// Add equipment component with items
	equipComp := NewEquipmentComponent()
	
	// Create legendary weapon with low durability (damaged but enchanted)
	weapon := &item.Item{
		ID:         "legendary-sword",
		Name:       "Legendary Excalibur",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Rarity:     item.RarityLegendary,
		Tags:       []string{"metal", "legendary"},
		Seed:       99999,
		Stats: item.Stats{
			Durability:    30,
			DurabilityMax: 100,
		},
	}
	equipComp.Equip(weapon, SlotMainHand)
	entity.AddComponent(equipComp)

	// Add equipment visual component
	equipVisualComp := NewEquipmentVisualComponent()
	equipVisualComp.SetWeapon(weapon.ID, weapon.Seed)
	entity.AddComponent(equipVisualComp)

	// Update system
	sys.Update([]*Entity{entity}, 0.0)

	// Verify equipment visual was built with Phase 15.3 properties
	config := sys.buildCompositeConfig(entity, equipVisualComp, spriteComp)
	
	if len(config.Equipment) != 1 {
		t.Fatalf("Expected 1 equipment item, got %d", len(config.Equipment))
	}

	equip := config.Equipment[0]

	// Verify material type
	if equip.Material != sprites.MaterialMetal {
		t.Errorf("Expected MaterialMetal, got %v", equip.Material)
	}

	// Verify damage state (30/100 = damaged)
	if equip.DamageState != sprites.DamageStateDamaged {
		t.Errorf("Expected DamageStateDamaged, got %v", equip.DamageState)
	}

	// Verify enchantment (legendary = gold glow)
	if !equip.Enchantment.Active {
		t.Error("Expected active enchantment for legendary item")
	}
	if equip.Enchantment.Color != "gold" {
		t.Errorf("Expected gold enchantment, got %v", equip.Enchantment.Color)
	}

	// Verify detail level (legendary = 1.0)
	if equip.DetailLevel != 1.0 {
		t.Errorf("Expected DetailLevel 1.0, got %v", equip.DetailLevel)
	}
}

// TestEquipmentVisual_Phase15_3_NoItemData tests graceful handling when item data is unavailable.
func TestEquipmentVisual_Phase15_3_NoItemData(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	// Build equipment visual without item data (nil item)
	visual := sys.buildEquipmentVisual("weapon", "test-id", 12345, sprites.LayerWeapon, nil, "fantasy")

	// Should have default values
	if visual.Material != sprites.MaterialMetal {
		t.Errorf("Expected default MaterialMetal, got %v", visual.Material)
	}
	if visual.DamageState != sprites.DamageStatePristine {
		t.Errorf("Expected default DamageStatePristine, got %v", visual.DamageState)
	}
	if visual.Enchantment.Active {
		t.Error("Expected no enchantment when item data unavailable")
	}
	if visual.DetailLevel != 0.5 {
		t.Errorf("Expected default DetailLevel 0.5, got %v", visual.DetailLevel)
	}
}

// TestEquipmentVisual_Phase15_3_MultipleItems tests multiple equipped items with different properties.
func TestEquipmentVisual_Phase15_3_MultipleItems(t *testing.T) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	// Create entity
	entity := NewEntity(1)
	
	genreComp := &GenreComponent{}
	genreComp.GenreID =("fantasy")
	entity.AddComponent(genreComp)

	spriteComp := &EbitenSprite{Width: 32, Height: 32}
	entity.AddComponent(spriteComp)

	// Add equipment component
	equipComp := NewEquipmentComponent()

	// Weapon: Common, pristine condition
	weapon := &item.Item{
		ID:         "common-sword",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Rarity:     item.RarityCommon,
		Tags:       []string{"metal"},
		Seed:       1000,
		Stats:      item.Stats{Durability: 100, DurabilityMax: 100},
	}
	equipComp.Equip(weapon, SlotMainHand)

	// Armor: Epic, worn condition
	armor := &item.Item{
		ID:        "epic-armor",
		Type:      item.TypeArmor,
		ArmorType: item.ArmorChest,
		Rarity:    item.RarityEpic,
		Tags:      []string{"metal"},
		Seed:      2000,
		Stats:     item.Stats{Durability: 60, DurabilityMax: 100},
	}
	equipComp.Equip(armor, SlotChest)

	entity.AddComponent(equipComp)

	// Equipment visual component
	equipVisualComp := NewEquipmentVisualComponent()
	equipVisualComp.SetWeapon(weapon.ID, weapon.Seed)
	equipVisualComp.SetArmor(armor.ID, armor.Seed)
	entity.AddComponent(equipVisualComp)

	// Build config
	config := sys.buildCompositeConfig(entity, equipVisualComp, spriteComp)

	if len(config.Equipment) != 2 {
		t.Fatalf("Expected 2 equipment items, got %d", len(config.Equipment))
	}

	// Find weapon and armor in equipment list
	var weaponVisual, armorVisual *sprites.EquipmentVisual
	for i := range config.Equipment {
		if config.Equipment[i].Slot == "weapon" {
			weaponVisual = &config.Equipment[i]
		} else if config.Equipment[i].Slot == "armor" {
			armorVisual = &config.Equipment[i]
		}
	}

	// Verify weapon (common, pristine)
	if weaponVisual == nil {
		t.Fatal("Weapon visual not found")
	}
	if weaponVisual.DamageState != sprites.DamageStatePristine {
		t.Errorf("Weapon should be pristine, got %v", weaponVisual.DamageState)
	}
	if weaponVisual.Enchantment.Active {
		t.Error("Common weapon should not have enchantment")
	}
	if weaponVisual.DetailLevel != 0.3 {
		t.Errorf("Common weapon DetailLevel should be 0.3, got %v", weaponVisual.DetailLevel)
	}

	// Verify armor (epic, worn)
	if armorVisual == nil {
		t.Fatal("Armor visual not found")
	}
	if armorVisual.DamageState != sprites.DamageStateWorn {
		t.Errorf("Armor should be worn, got %v", armorVisual.DamageState)
	}
	if !armorVisual.Enchantment.Active {
		t.Error("Epic armor should have enchantment")
	}
	if armorVisual.Enchantment.Color != "purple" {
		t.Errorf("Epic enchantment should be purple, got %v", armorVisual.Enchantment.Color)
	}
	if armorVisual.DetailLevel != 0.8 {
		t.Errorf("Epic armor DetailLevel should be 0.8, got %v", armorVisual.DetailLevel)
	}
}

// Benchmark tests for performance verification (<0.2ms requirement)
func BenchmarkEquipmentVisual_Phase15_3_BuildVisual(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	testItem := &item.Item{
		ID:         "test-item",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Rarity:     item.RarityLegendary,
		Tags:       []string{"metal", "legendary"},
		Seed:       12345,
		Stats:      item.Stats{Durability: 50, DurabilityMax: 100},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.buildEquipmentVisual("weapon", testItem.ID, testItem.Seed, sprites.LayerWeapon, testItem, "fantasy")
	}
}

func BenchmarkEquipmentVisual_Phase15_3_GetMaterialType(b *testing.B) {
	gen := sprites.NewGenerator()
	sys := NewEquipmentVisualSystem(gen)

	testItem := &item.Item{
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Tags:       []string{"metal", "heavy"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.getMaterialType(testItem, "weapon", "fantasy")
	}
}
