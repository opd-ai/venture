package item

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewItemGenerator(t *testing.T) {
	gen := NewItemGenerator()
	if gen == nil {
		t.Fatal("NewItemGenerator returned nil")
	}
	if len(gen.weaponTemplates) == 0 {
		t.Error("weaponTemplates not initialized")
	}
	if len(gen.armorTemplates) == 0 {
		t.Error("armorTemplates not initialized")
	}
	if len(gen.consumableTemplates) == 0 {
		t.Error("consumableTemplates not initialized")
	}
}

func TestItemGeneration(t *testing.T) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 20,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	items, ok := result.([]*Item)
	if !ok {
		t.Fatal("Generate did not return []*Item")
	}

	if len(items) != 20 {
		t.Errorf("Expected 20 items, got %d", len(items))
	}

	// Verify all items are valid
	for i, item := range items {
		if item == nil {
			t.Errorf("Item %d is nil", i)
			continue
		}
		if item.ID == "" {
			t.Errorf("Item %d has empty ID", i)
		}
		if item.Name == "" {
			t.Errorf("Item %d has empty name", i)
		}
		if item.Stats.Value < 0 {
			t.Errorf("Item %d has negative value", i)
		}
	}
}

func TestItemGenerationDeterministic(t *testing.T) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Depth:      3,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
		},
	}

	seed := int64(54321)

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generate failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generate failed: %v", err2)
	}

	items1 := result1.([]*Item)
	items2 := result2.([]*Item)

	if len(items1) != len(items2) {
		t.Fatalf("Different number of items: %d vs %d", len(items1), len(items2))
	}

	// Verify items are identical
	for i := range items1 {
		if items1[i].Name != items2[i].Name {
			t.Errorf("Item %d name differs: %s vs %s", i, items1[i].Name, items2[i].Name)
		}
		if items1[i].Type != items2[i].Type {
			t.Errorf("Item %d type differs", i)
		}
		if items1[i].Rarity != items2[i].Rarity {
			t.Errorf("Item %d rarity differs", i)
		}
		if items1[i].Stats.Damage != items2[i].Stats.Damage {
			t.Errorf("Item %d damage differs", i)
		}
		if items1[i].Stats.Defense != items2[i].Stats.Defense {
			t.Errorf("Item %d defense differs", i)
		}
	}
}

func TestItemGenerationSciFi(t *testing.T) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"count": 15,
		},
	}

	result, err := gen.Generate(99999, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	items := result.([]*Item)
	if len(items) != 15 {
		t.Errorf("Expected 15 items, got %d", len(items))
	}

	// Verify items exist and have valid properties
	for _, item := range items {
		if item.Name == "" {
			t.Error("Item has empty name")
		}
	}
}

func TestItemValidation(t *testing.T) {
	gen := NewItemGenerator()

	// Test valid items
	validItems := []*Item{
		{
			Name: "Test Sword",
			Type: TypeWeapon,
			Stats: Stats{
				Damage: 10,
				Value:  100,
			},
		},
		{
			Name: "Test Armor",
			Type: TypeArmor,
			Stats: Stats{
				Defense: 15,
				Value:   200,
			},
		},
	}

	err := gen.Validate(validItems)
	if err != nil {
		t.Errorf("Validation failed for valid items: %v", err)
	}

	// Test invalid: nil item
	invalidItems1 := []*Item{nil}
	err = gen.Validate(invalidItems1)
	if err == nil {
		t.Error("Expected error for nil item")
	}

	// Test invalid: empty name
	invalidItems2 := []*Item{
		{
			Name: "",
			Type: TypeWeapon,
			Stats: Stats{
				Damage: 10,
			},
		},
	}
	err = gen.Validate(invalidItems2)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Test invalid: weapon with no damage
	invalidItems3 := []*Item{
		{
			Name: "Broken Sword",
			Type: TypeWeapon,
			Stats: Stats{
				Damage: 0,
			},
		},
	}
	err = gen.Validate(invalidItems3)
	if err == nil {
		t.Error("Expected error for weapon with no damage")
	}

	// Test invalid: armor with no defense
	invalidItems4 := []*Item{
		{
			Name: "Broken Armor",
			Type: TypeArmor,
			Stats: Stats{
				Defense: 0,
			},
		},
	}
	err = gen.Validate(invalidItems4)
	if err == nil {
		t.Error("Expected error for armor with no defense")
	}

	// Test invalid: wrong type
	err = gen.Validate("not a slice")
	if err == nil {
		t.Error("Expected error for wrong type")
	}
}

func TestItemTypes(t *testing.T) {
	tests := []struct {
		itemType ItemType
		expected string
	}{
		{TypeWeapon, "weapon"},
		{TypeArmor, "armor"},
		{TypeConsumable, "consumable"},
		{TypeAccessory, "accessory"},
	}

	for _, tt := range tests {
		if tt.itemType.String() != tt.expected {
			t.Errorf("ItemType.String() = %s, expected %s", tt.itemType.String(), tt.expected)
		}
	}
}

func TestWeaponTypes(t *testing.T) {
	tests := []struct {
		weaponType WeaponType
		expected   string
	}{
		{WeaponSword, "sword"},
		{WeaponAxe, "axe"},
		{WeaponBow, "bow"},
		{WeaponStaff, "staff"},
		{WeaponDagger, "dagger"},
		{WeaponSpear, "spear"},
	}

	for _, tt := range tests {
		if tt.weaponType.String() != tt.expected {
			t.Errorf("WeaponType.String() = %s, expected %s", tt.weaponType.String(), tt.expected)
		}
	}
}

func TestArmorTypes(t *testing.T) {
	tests := []struct {
		armorType ArmorType
		expected  string
	}{
		{ArmorHelmet, "helmet"},
		{ArmorChest, "chest"},
		{ArmorLegs, "legs"},
		{ArmorBoots, "boots"},
		{ArmorGloves, "gloves"},
		{ArmorShield, "shield"},
	}

	for _, tt := range tests {
		if tt.armorType.String() != tt.expected {
			t.Errorf("ArmorType.String() = %s, expected %s", tt.armorType.String(), tt.expected)
		}
	}
}

func TestConsumableTypes(t *testing.T) {
	tests := []struct {
		consumableType ConsumableType
		expected       string
	}{
		{ConsumablePotion, "potion"},
		{ConsumableScroll, "scroll"},
		{ConsumableFood, "food"},
		{ConsumableBomb, "bomb"},
	}

	for _, tt := range tests {
		if tt.consumableType.String() != tt.expected {
			t.Errorf("ConsumableType.String() = %s, expected %s", tt.consumableType.String(), tt.expected)
		}
	}
}

func TestRarity(t *testing.T) {
	tests := []struct {
		rarity   Rarity
		expected string
	}{
		{RarityCommon, "common"},
		{RarityUncommon, "uncommon"},
		{RarityRare, "rare"},
		{RarityEpic, "epic"},
		{RarityLegendary, "legendary"},
	}

	for _, tt := range tests {
		if tt.rarity.String() != tt.expected {
			t.Errorf("Rarity.String() = %s, expected %s", tt.rarity.String(), tt.expected)
		}
	}
}

func TestItemIsEquippable(t *testing.T) {
	tests := []struct {
		itemType ItemType
		expected bool
	}{
		{TypeWeapon, true},
		{TypeArmor, true},
		{TypeAccessory, true},
		{TypeConsumable, false},
	}

	for _, tt := range tests {
		item := &Item{Type: tt.itemType}
		if item.IsEquippable() != tt.expected {
			t.Errorf("Item type %s: IsEquippable() = %v, expected %v",
				tt.itemType, item.IsEquippable(), tt.expected)
		}
	}
}

func TestItemIsConsumable(t *testing.T) {
	tests := []struct {
		itemType ItemType
		expected bool
	}{
		{TypeWeapon, false},
		{TypeArmor, false},
		{TypeAccessory, false},
		{TypeConsumable, true},
	}

	for _, tt := range tests {
		item := &Item{Type: tt.itemType}
		if item.IsConsumable() != tt.expected {
			t.Errorf("Item type %s: IsConsumable() = %v, expected %v",
				tt.itemType, item.IsConsumable(), tt.expected)
		}
	}
}

func TestItemGetValue(t *testing.T) {
	tests := []struct {
		name               string
		baseValue          int
		durability         int
		durabilityMax      int
		expectedValue      int
		expectValueReduced bool
	}{
		{"full durability", 100, 100, 100, 100, false},
		{"half durability", 100, 50, 100, 50, true},
		{"no durability system", 100, 0, 0, 100, false},
		{"quarter durability", 100, 25, 100, 25, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{
				Stats: Stats{
					Value:         tt.baseValue,
					Durability:    tt.durability,
					DurabilityMax: tt.durabilityMax,
				},
			}
			value := item.GetValue()
			if value != tt.expectedValue {
				t.Errorf("GetValue() = %d, expected %d", value, tt.expectedValue)
			}
		})
	}
}

func TestGetFantasyWeaponTemplates(t *testing.T) {
	templates := GetFantasyWeaponTemplates()
	if len(templates) == 0 {
		t.Error("No fantasy weapon templates returned")
	}

	for i, template := range templates {
		if template.BaseType != TypeWeapon {
			t.Errorf("Template %d is not a weapon", i)
		}
		if len(template.NamePrefixes) == 0 {
			t.Errorf("Template %d has no name prefixes", i)
		}
		if len(template.NameSuffixes) == 0 {
			t.Errorf("Template %d has no name suffixes", i)
		}
		if template.DamageRange[1] <= 0 {
			t.Errorf("Template %d has invalid damage range", i)
		}
	}
}

func TestGetFantasyArmorTemplates(t *testing.T) {
	templates := GetFantasyArmorTemplates()
	if len(templates) == 0 {
		t.Error("No fantasy armor templates returned")
	}

	for i, template := range templates {
		if template.BaseType != TypeArmor {
			t.Errorf("Template %d is not armor", i)
		}
		if len(template.NamePrefixes) == 0 {
			t.Errorf("Template %d has no name prefixes", i)
		}
		if len(template.NameSuffixes) == 0 {
			t.Errorf("Template %d has no name suffixes", i)
		}
		if template.DefenseRange[1] <= 0 {
			t.Errorf("Template %d has invalid defense range", i)
		}
	}
}

func TestGetFantasyConsumableTemplates(t *testing.T) {
	templates := GetFantasyConsumableTemplates()
	if len(templates) == 0 {
		t.Error("No fantasy consumable templates returned")
	}

	for i, template := range templates {
		if template.BaseType != TypeConsumable {
			t.Errorf("Template %d is not a consumable", i)
		}
		if len(template.NamePrefixes) == 0 {
			t.Errorf("Template %d has no name prefixes", i)
		}
		if len(template.NameSuffixes) == 0 {
			t.Errorf("Template %d has no name suffixes", i)
		}
	}
}

func TestGetSciFiWeaponTemplates(t *testing.T) {
	templates := GetSciFiWeaponTemplates()
	if len(templates) == 0 {
		t.Error("No sci-fi weapon templates returned")
	}

	for i, template := range templates {
		if template.BaseType != TypeWeapon {
			t.Errorf("Template %d is not a weapon", i)
		}
	}
}

func TestGetSciFiArmorTemplates(t *testing.T) {
	templates := GetSciFiArmorTemplates()
	if len(templates) == 0 {
		t.Error("No sci-fi armor templates returned")
	}

	for i, template := range templates {
		if template.BaseType != TypeArmor {
			t.Errorf("Template %d is not armor", i)
		}
	}
}

func TestItemLevelScaling(t *testing.T) {
	gen := NewItemGenerator()

	// Generate items at different depths
	depths := []int{1, 5, 10, 20}
	for _, depth := range depths {
		params := procgen.GenerationParams{
			Depth:      depth,
			Difficulty: 0.5,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"count": 5,
			},
		}

		result, err := gen.Generate(12345, params)
		if err != nil {
			t.Fatalf("Generate at depth %d failed: %v", depth, err)
		}

		items := result.([]*Item)
		for _, item := range items {
			// Higher depth should generally mean higher required levels
			if item.Stats.RequiredLevel < 1 {
				t.Errorf("Depth %d: item has invalid required level: %d", depth, item.Stats.RequiredLevel)
			}
		}
	}
}

func TestItemTypeFiltering(t *testing.T) {
	gen := NewItemGenerator()

	// Test weapon filtering
	params := procgen.GenerationParams{
		Depth:      5,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
			"type":  "weapon",
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	items := result.([]*Item)
	for i, item := range items {
		if item.Type != TypeWeapon {
			t.Errorf("Item %d is not a weapon (type: %s)", i, item.Type)
		}
		if item.Stats.Damage <= 0 {
			t.Errorf("Weapon %d has no damage", i)
		}
	}

	// Test armor filtering
	params.Custom["type"] = "armor"
	result, err = gen.Generate(54321, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	items = result.([]*Item)
	for i, item := range items {
		if item.Type != TypeArmor {
			t.Errorf("Item %d is not armor (type: %s)", i, item.Type)
		}
		if item.Stats.Defense <= 0 {
			t.Errorf("Armor %d has no defense", i)
		}
	}
}

func TestRarityDistribution(t *testing.T) {
	gen := NewItemGenerator()

	// Generate many items at low depth - should be mostly common
	params := procgen.GenerationParams{
		Depth:      1,
		Difficulty: 0.5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 100,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	items := result.([]*Item)
	rarityCounts := make(map[Rarity]int)
	for _, item := range items {
		rarityCounts[item.Rarity]++
	}

	// At depth 1, common should be most prevalent
	if rarityCounts[RarityCommon] < 30 {
		t.Errorf("Expected more common items at depth 1, got %d", rarityCounts[RarityCommon])
	}

	// Generate items at high depth - should have more rare items
	params.Depth = 20
	result, err = gen.Generate(54321, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	items = result.([]*Item)
	rarityCounts = make(map[Rarity]int)
	for _, item := range items {
		rarityCounts[item.Rarity]++
	}

	// At depth 20, should have at least some uncommon+ items
	uncommonPlus := rarityCounts[RarityUncommon] + rarityCounts[RarityRare] +
		rarityCounts[RarityEpic] + rarityCounts[RarityLegendary]
	if uncommonPlus < 20 {
		t.Errorf("Expected more uncommon+ items at depth 20, got %d", uncommonPlus)
	}
}
