//go:build !android && !ios
// +build !android,!ios

package main

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/class/advanced"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/qol"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

// TestMapCharacterClassToAdvancedClass tests class mapping for all character classes.
func TestMapCharacterClassToAdvancedClass(t *testing.T) {
	tests := []struct {
		name     string
		class    engine.CharacterClass
		expected advanced.ClassID
	}{
		{"warrior maps to advanced warrior", engine.ClassWarrior, advanced.ClassWarrior},
		{"mage maps to advanced mage", engine.ClassMage, advanced.ClassMage},
		{"rogue maps to advanced rogue", engine.ClassRogue, advanced.ClassRogue},
		{"ranger maps to advanced ranger", engine.ClassRanger, advanced.ClassRanger},
		{"cleric maps to advanced cleric", engine.ClassCleric, advanced.ClassCleric},
		{"necromancer maps to advanced necromancer", engine.ClassNecromancer, advanced.ClassNecromancer},
		{"battlemage maps to warrior (default hybrid)", engine.ClassBattlemage, advanced.ClassWarrior},
		{"unknown class returns empty", engine.CharacterClass(999), ""},
		{"negative class returns empty", engine.CharacterClass(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapCharacterClassToAdvancedClass(tt.class)
			if result != tt.expected {
				t.Errorf("mapCharacterClassToAdvancedClass(%q) = %q, want %q", tt.class, result, tt.expected)
			}
		})
	}
}

// TestMapCharacterClassToAdvancedClassDeterminism verifies consistent results.
func TestMapCharacterClassToAdvancedClassDeterminism(t *testing.T) {
	classes := []engine.CharacterClass{
		engine.ClassWarrior,
		engine.ClassMage,
		engine.ClassRogue,
		engine.ClassRanger,
		engine.ClassCleric,
		engine.ClassNecromancer,
		engine.ClassBattlemage,
	}

	for _, class := range classes {
		result1 := mapCharacterClassToAdvancedClass(class)
		result2 := mapCharacterClassToAdvancedClass(class)
		if result1 != result2 {
			t.Errorf("class %q: inconsistent results %q vs %q", class, result1, result2)
		}
	}
}

// TestCreateGenerationParams tests generation params creation.
func TestCreateGenerationParams(t *testing.T) {
	// Set test values for the global flags
	oldGenreID := *genreID
	oldSeed := *seed
	defer func() {
		*genreID = oldGenreID
		*seed = oldSeed
	}()

	*genreID = "fantasy"
	*seed = 12345

	params := createGenerationParams()

	// Verify expected defaults
	if params.Difficulty != defaultDifficulty {
		t.Errorf("Difficulty = %f, want %f", params.Difficulty, defaultDifficulty)
	}
	if params.Depth != defaultDepth {
		t.Errorf("Depth = %d, want %d", params.Depth, defaultDepth)
	}
	if params.GenreID != "fantasy" {
		t.Errorf("GenreID = %q, want %q", params.GenreID, "fantasy")
	}

	// Verify custom map has expected keys
	if params.Custom == nil {
		t.Fatal("Custom map is nil")
	}
	if _, ok := params.Custom["width"]; !ok {
		t.Error("Custom map missing 'width' key")
	}
	if _, ok := params.Custom["height"]; !ok {
		t.Error("Custom map missing 'height' key")
	}
	if _, ok := params.Custom["theme"]; !ok {
		t.Error("Custom map missing 'theme' key")
	}

	// Verify width and height values
	if w, ok := params.Custom["width"].(int); !ok || w != defaultTerrainWidth {
		t.Errorf("Custom[width] = %v, want %d", params.Custom["width"], defaultTerrainWidth)
	}
	if h, ok := params.Custom["height"].(int); !ok || h != defaultTerrainHeight {
		t.Errorf("Custom[height] = %v, want %d", params.Custom["height"], defaultTerrainHeight)
	}
}

// TestCreateGenerationParamsGenres tests params creation for all genres.
func TestCreateGenerationParamsGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			// Set genre flag
			oldGenreID := *genreID
			*genreID = g
			defer func() { *genreID = oldGenreID }()

			params := createGenerationParams()

			if params.GenreID != g {
				t.Errorf("GenreID = %q, want %q", params.GenreID, g)
			}

			// Theme should be non-nil
			if params.Custom["theme"] == nil {
				t.Error("theme is nil")
			}
		})
	}
}

// TestSerializePosition tests position serialization.
func TestSerializePosition(t *testing.T) {
	tests := []struct {
		name  string
		x, y  float64
		wantX float64
		wantY float64
	}{
		{"positive coords", 100.5, 200.7, 100.5, 200.7},
		{"zero coords", 0, 0, 0, 0},
		{"negative coords", -50.5, -75.3, -50.5, -75.3},
		{"large coords", 10000.0, 20000.0, 10000.0, 20000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := engine.NewEntity(1)
			player.AddComponent(&engine.PositionComponent{X: tt.x, Y: tt.y})

			state := &saveload.PlayerState{}
			serializePosition(player, state)

			if state.X != tt.wantX {
				t.Errorf("X = %f, want %f", state.X, tt.wantX)
			}
			if state.Y != tt.wantY {
				t.Errorf("Y = %f, want %f", state.Y, tt.wantY)
			}
		})
	}
}

// TestSerializePositionNoComponent tests serialization without position component.
func TestSerializePositionNoComponent(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{X: 999, Y: 888} // Pre-set values

	serializePosition(player, state)

	// Should remain unchanged when no component exists
	if state.X != 999 || state.Y != 888 {
		t.Errorf("State modified without position component: X=%f, Y=%f", state.X, state.Y)
	}
}

// TestSerializeHealth tests health serialization.
func TestSerializeHealth(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		max      float64
		wantCurr float64
		wantMax  float64
	}{
		{"full health", 100, 100, 100, 100},
		{"partial health", 50, 100, 50, 100},
		{"zero health", 0, 100, 0, 100},
		{"boosted max", 150, 200, 150, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := engine.NewEntity(1)
			player.AddComponent(&engine.HealthComponent{Current: tt.current, Max: tt.max})

			state := &saveload.PlayerState{}
			serializeHealth(player, state)

			if state.CurrentHealth != tt.wantCurr {
				t.Errorf("CurrentHealth = %f, want %f", state.CurrentHealth, tt.wantCurr)
			}
			if state.MaxHealth != tt.wantMax {
				t.Errorf("MaxHealth = %f, want %f", state.MaxHealth, tt.wantMax)
			}
		})
	}
}

// TestSerializeStats tests stats serialization.
func TestSerializeStats(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.StatsComponent{
		Attack:     25,
		Defense:    15,
		MagicPower: 30,
	})

	state := &saveload.PlayerState{}
	serializeStats(player, state)

	if state.Attack != 25 {
		t.Errorf("Attack = %f, want 25", state.Attack)
	}
	if state.Defense != 15 {
		t.Errorf("Defense = %f, want 15", state.Defense)
	}
	if state.MagicPower != 30 {
		t.Errorf("MagicPower = %f, want 30", state.MagicPower)
	}
}

// TestSerializeExperience tests experience serialization.
func TestSerializeExperience(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		xp      int
		wantLvl int
		wantXP  int
	}{
		{"level 1 fresh", 1, 0, 1, 0},
		{"level 5 with xp", 5, 2500, 5, 2500},
		{"max level", 100, 999999, 100, 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := engine.NewEntity(1)
			player.AddComponent(&engine.ExperienceComponent{
				Level:     tt.level,
				CurrentXP: tt.xp,
			})

			state := &saveload.PlayerState{}
			serializeExperience(player, state)

			if state.Level != tt.wantLvl {
				t.Errorf("Level = %d, want %d", state.Level, tt.wantLvl)
			}
			if state.Experience != tt.wantXP {
				t.Errorf("Experience = %d, want %d", state.Experience, tt.wantXP)
			}
		})
	}
}

// TestDeserializePosition tests position deserialization.
func TestDeserializePosition(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.PositionComponent{X: 0, Y: 0})

	state := &saveload.PlayerState{X: 500.5, Y: 300.3}
	deserializePosition(player, state)

	posComp, _ := player.GetComponent("position")
	pos := posComp.(*engine.PositionComponent)

	if pos.X != 500.5 {
		t.Errorf("X = %f, want 500.5", pos.X)
	}
	if pos.Y != 300.3 {
		t.Errorf("Y = %f, want 300.3", pos.Y)
	}
}

// TestDeserializeHealth tests health deserialization.
func TestDeserializeHealth(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})

	state := &saveload.PlayerState{CurrentHealth: 75, MaxHealth: 150}
	deserializeHealth(player, state)

	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*engine.HealthComponent)

	if health.Current != 75 {
		t.Errorf("Current = %f, want 75", health.Current)
	}
	if health.Max != 150 {
		t.Errorf("Max = %f, want 150", health.Max)
	}
}

// TestDeserializeStats tests stats deserialization.
func TestDeserializeStats(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.StatsComponent{Attack: 10, Defense: 5, MagicPower: 10})

	state := &saveload.PlayerState{Attack: 50, Defense: 30, MagicPower: 75}
	deserializeStats(player, state)

	statsComp, _ := player.GetComponent("stats")
	stats := statsComp.(*engine.StatsComponent)

	if stats.Attack != 50 {
		t.Errorf("Attack = %f, want 50", stats.Attack)
	}
	if stats.Defense != 30 {
		t.Errorf("Defense = %f, want 30", stats.Defense)
	}
	if stats.MagicPower != 75 {
		t.Errorf("MagicPower = %f, want 75", stats.MagicPower)
	}
}

// TestDeserializeExperience tests experience deserialization.
func TestDeserializeExperience(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.ExperienceComponent{Level: 1, CurrentXP: 0})

	state := &saveload.PlayerState{Level: 15, Experience: 50000}
	deserializeExperience(player, state)

	expComp, _ := player.GetComponent("experience")
	exp := expComp.(*engine.ExperienceComponent)

	if exp.Level != 15 {
		t.Errorf("Level = %d, want 15", exp.Level)
	}
	if exp.CurrentXP != 50000 {
		t.Errorf("CurrentXP = %d, want 50000", exp.CurrentXP)
	}
}

// TestSerializeDeserializeRoundTrip tests that serialization is reversible.
func TestSerializeDeserializeRoundTrip(t *testing.T) {
	// Create player with all components
	player := engine.NewEntity(1)
	player.AddComponent(&engine.PositionComponent{X: 123.456, Y: 789.012})
	player.AddComponent(&engine.HealthComponent{Current: 85, Max: 120})
	player.AddComponent(&engine.StatsComponent{Attack: 42, Defense: 28, MagicPower: 55})
	player.AddComponent(&engine.ExperienceComponent{Level: 12, CurrentXP: 15000})

	// Serialize
	state := &saveload.PlayerState{}
	serializePosition(player, state)
	serializeHealth(player, state)
	serializeStats(player, state)
	serializeExperience(player, state)

	// Create new player and deserialize
	player2 := engine.NewEntity(2)
	player2.AddComponent(&engine.PositionComponent{})
	player2.AddComponent(&engine.HealthComponent{})
	player2.AddComponent(&engine.StatsComponent{})
	player2.AddComponent(&engine.ExperienceComponent{})

	deserializePosition(player2, state)
	deserializeHealth(player2, state)
	deserializeStats(player2, state)
	deserializeExperience(player2, state)

	// Verify all values match
	pos1, _ := player.GetComponent("position")
	pos2, _ := player2.GetComponent("position")
	if pos1.(*engine.PositionComponent).X != pos2.(*engine.PositionComponent).X {
		t.Error("Position X mismatch after round-trip")
	}
	if pos1.(*engine.PositionComponent).Y != pos2.(*engine.PositionComponent).Y {
		t.Error("Position Y mismatch after round-trip")
	}

	health1, _ := player.GetComponent("health")
	health2, _ := player2.GetComponent("health")
	if health1.(*engine.HealthComponent).Current != health2.(*engine.HealthComponent).Current {
		t.Error("Health Current mismatch after round-trip")
	}

	stats1, _ := player.GetComponent("stats")
	stats2, _ := player2.GetComponent("stats")
	if stats1.(*engine.StatsComponent).Attack != stats2.(*engine.StatsComponent).Attack {
		t.Error("Attack mismatch after round-trip")
	}

	exp1, _ := player.GetComponent("experience")
	exp2, _ := player2.GetComponent("experience")
	if exp1.(*engine.ExperienceComponent).Level != exp2.(*engine.ExperienceComponent).Level {
		t.Error("Level mismatch after round-trip")
	}
}

// TestSerializeInventory tests inventory serialization.
func TestSerializeInventory(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.InventoryComponent{
		Gold:     500,
		MaxItems: 20,
		Items:    nil, // No items for basic test
	})

	state := &saveload.PlayerState{}
	serializeInventory(player, state)

	if state.Gold != 500 {
		t.Errorf("Gold = %d, want 500", state.Gold)
	}
	if state.Items == nil {
		t.Error("Items should not be nil")
	}
	if len(state.Items) != 0 {
		t.Errorf("Items length = %d, want 0", len(state.Items))
	}
}

// TestSerializeInventoryNoComponent tests serialization without inventory component.
func TestSerializeInventoryNoComponent(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{Gold: 999} // Pre-set value

	serializeInventory(player, state)

	// Should remain unchanged when no component exists
	if state.Gold != 999 {
		t.Errorf("Gold modified without inventory component: %d", state.Gold)
	}
}

// TestSerializeEquipment tests equipment serialization with no items.
func TestSerializeEquipment(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.EquipmentComponent{
		Slots: make(map[engine.EquipmentSlot]*item.Item),
	})

	state := &saveload.PlayerState{}
	serializeEquipment(player, state)

	// With empty slots, all equipped items should be nil
	if state.EquippedItems.Weapon != nil {
		t.Error("Weapon should be nil with empty slots")
	}
	if state.EquippedItems.Armor != nil {
		t.Error("Armor should be nil with empty slots")
	}
	if state.EquippedItems.Accessory != nil {
		t.Error("Accessory should be nil with empty slots")
	}
}

// TestSerializeEquipmentWithItems tests serialization with actual equipped items.
func TestSerializeEquipmentWithItems(t *testing.T) {
	player := engine.NewEntity(1)
	equipment := &engine.EquipmentComponent{
		Slots: make(map[engine.EquipmentSlot]*item.Item),
	}

	// Create test items
	weapon := &item.Item{
		ID:   "test-weapon-1",
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Damage: 25,
			Value:  100,
		},
	}
	armor := &item.Item{
		ID:   "test-armor-1",
		Name: "Test Chestplate",
		Type: item.TypeArmor,
		Stats: item.Stats{
			Defense: 15,
			Value:   200,
		},
	}
	accessory := &item.Item{
		ID:   "test-accessory-1",
		Name: "Test Ring",
		Type: item.TypeAccessory,
		Stats: item.Stats{
			Value: 50,
		},
	}

	equipment.Slots[engine.SlotMainHand] = weapon
	equipment.Slots[engine.SlotChest] = armor
	equipment.Slots[engine.SlotAccessory1] = accessory
	player.AddComponent(equipment)

	state := &saveload.PlayerState{}
	serializeEquipment(player, state)

	// Verify weapon was serialized
	if state.EquippedItems.Weapon == nil {
		t.Fatal("Weapon should not be nil")
	}
	if state.EquippedItems.Weapon.ID != weapon.ID {
		t.Errorf("Weapon ID = %s, want %s", state.EquippedItems.Weapon.ID, weapon.ID)
	}
	if state.EquippedItems.Weapon.Name != weapon.Name {
		t.Errorf("Weapon Name = %s, want %s", state.EquippedItems.Weapon.Name, weapon.Name)
	}

	// Verify armor was serialized
	if state.EquippedItems.Armor == nil {
		t.Fatal("Armor should not be nil")
	}
	if state.EquippedItems.Armor.ID != armor.ID {
		t.Errorf("Armor ID = %s, want %s", state.EquippedItems.Armor.ID, armor.ID)
	}

	// Verify accessory was serialized
	if state.EquippedItems.Accessory == nil {
		t.Fatal("Accessory should not be nil")
	}
	if state.EquippedItems.Accessory.ID != accessory.ID {
		t.Errorf("Accessory ID = %s, want %s", state.EquippedItems.Accessory.ID, accessory.ID)
	}
}

// TestSerializeEquipmentNoComponent tests serialization without equipment component.
func TestSerializeEquipmentNoComponent(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{}

	serializeEquipment(player, state)

	// Should not panic and equipped items should remain nil
	if state.EquippedItems.Weapon != nil {
		t.Error("Weapon should remain nil")
	}
}

// TestDeserializeInventory tests inventory deserialization.
func TestDeserializeInventory(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.InventoryComponent{
		Gold:     0,
		MaxItems: 20,
		Items:    nil,
	})

	state := &saveload.PlayerState{
		Gold:  750,
		Items: []saveload.ItemData{},
	}
	deserializeInventory(player, state)

	invComp, _ := player.GetComponent("inventory")
	inv := invComp.(*engine.InventoryComponent)

	if inv.Gold != 750 {
		t.Errorf("Gold = %d, want 750", inv.Gold)
	}
}

// BenchmarkMapCharacterClassToAdvancedClass benchmarks class mapping.
func BenchmarkMapCharacterClassToAdvancedClass(b *testing.B) {
	classes := []engine.CharacterClass{
		engine.ClassWarrior,
		engine.ClassMage,
		engine.ClassRogue,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, class := range classes {
			mapCharacterClassToAdvancedClass(class)
		}
	}
}

// BenchmarkSerializePosition benchmarks position serialization.
func BenchmarkSerializePosition(b *testing.B) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	state := &saveload.PlayerState{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serializePosition(player, state)
	}
}

// BenchmarkDeserializePosition benchmarks position deserialization.
func BenchmarkDeserializePosition(b *testing.B) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.PositionComponent{})
	state := &saveload.PlayerState{X: 100, Y: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deserializePosition(player, state)
	}
}

// TestSerializeManaAndSpells tests mana serialization.
func TestSerializeManaAndSpells(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.ManaComponent{
		Current: 75,
		Max:     100,
	})

	state := &saveload.PlayerState{}
	serializeManaAndSpells(player, state)

	if state.CurrentMana != 75 {
		t.Errorf("CurrentMana = %d, want 75", state.CurrentMana)
	}
	if state.MaxMana != 100 {
		t.Errorf("MaxMana = %d, want 100", state.MaxMana)
	}
}

// TestSerializeManaAndSpellsWithSpells tests mana and spell serialization with actual spells.
func TestSerializeManaAndSpellsWithSpells(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.ManaComponent{
		Current: 80,
		Max:     120,
	})

	spellSlots := &engine.SpellSlotComponent{}
	spellSlots.Slots[0] = &magic.Spell{
		Name:    "Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
		Stats: magic.Stats{
			Damage:   50,
			ManaCost: 25,
			CastTime: 1.5,
			Cooldown: 3.0,
			AreaSize: 100,
			Range:    200,
		},
	}
	spellSlots.Slots[1] = &magic.Spell{
		Name:    "Heal",
		Type:    magic.TypeHealing,
		Element: magic.ElementLight,
		Stats: magic.Stats{
			Healing:  30,
			ManaCost: 20,
			CastTime: 2.0,
			Cooldown: 5.0,
		},
	}
	spellSlots.Slots[2] = &magic.Spell{
		Name:    "Frost Nova",
		Type:    magic.TypeOffensive,
		Element: magic.ElementIce,
		Stats: magic.Stats{
			Damage:   35,
			ManaCost: 30,
		},
	}
	player.AddComponent(spellSlots)

	state := &saveload.PlayerState{}
	serializeManaAndSpells(player, state)

	// Verify mana
	if state.CurrentMana != 80 {
		t.Errorf("CurrentMana = %d, want 80", state.CurrentMana)
	}
	if state.MaxMana != 120 {
		t.Errorf("MaxMana = %d, want 120", state.MaxMana)
	}

	// Verify spells
	if len(state.Spells) != 3 {
		t.Fatalf("len(Spells) = %d, want 3", len(state.Spells))
	}

	if state.Spells[0].Name != "Fireball" {
		t.Errorf("Spell 0 Name = %q, want %q", state.Spells[0].Name, "Fireball")
	}
	if state.Spells[1].Name != "Heal" {
		t.Errorf("Spell 1 Name = %q, want %q", state.Spells[1].Name, "Heal")
	}
	if state.Spells[2].Name != "Frost Nova" {
		t.Errorf("Spell 2 Name = %q, want %q", state.Spells[2].Name, "Frost Nova")
	}
}

// TestSerializeManaAndSpellsNoComponent tests mana serialization without component.
func TestSerializeManaAndSpellsNoComponent(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{CurrentMana: 999, MaxMana: 888}

	serializeManaAndSpells(player, state)

	// Should remain unchanged when no component exists
	if state.CurrentMana != 999 || state.MaxMana != 888 {
		t.Errorf("State modified without mana component: CurrentMana=%d, MaxMana=%d", state.CurrentMana, state.MaxMana)
	}
}

// TestSerializeQoLState tests QoL state serialization.
func TestSerializeQoLState(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&qol.QoLComponent{
		PlayerID:        12345,
		AutoLootEnabled: true,
		AutoLootRadius:  150.0,
		SortPreset:      "type",
		MountWhistle:    true,
		RecipeTracking:  true,
	})

	state := &saveload.PlayerState{}
	serializeQoLState(player, state)

	if state.QoLData == nil {
		t.Fatal("QoLData should not be nil")
	}
	if state.QoLData.PlayerID != 12345 {
		t.Errorf("PlayerID = %d, want %d", state.QoLData.PlayerID, 12345)
	}
	if !state.QoLData.AutoLootEnabled {
		t.Error("AutoLootEnabled should be true")
	}
	if state.QoLData.AutoLootRadius != 150.0 {
		t.Errorf("AutoLootRadius = %f, want 150.0", state.QoLData.AutoLootRadius)
	}
	if state.QoLData.SortPreset != "type" {
		t.Errorf("SortPreset = %q, want %q", state.QoLData.SortPreset, "type")
	}
	if !state.QoLData.RecipeTracking {
		t.Error("RecipeTracking should be true")
	}
}

// TestSerializeQoLStateNoComponent tests QoL serialization without component.
func TestSerializeQoLStateNoComponent(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{}

	serializeQoLState(player, state)

	// QoLData should remain nil when no component exists
	if state.QoLData != nil {
		t.Error("QoLData should be nil without QoL component")
	}
}

// TestDeserializeQoLState tests QoL state deserialization.
func TestDeserializeQoLState(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{
		QoLData: &saveload.QoLStateData{
			PlayerID:        99999,
			AutoLootEnabled: true,
			AutoLootRadius:  200.0,
			SortPreset:      "rarity",
			MountWhistle:    true,
			RecipeTracking:  true,
		},
	}

	deserializeQoLState(player, state)

	qolComp, ok := player.GetComponent("qol")
	if !ok {
		t.Fatal("QoL component should be created")
	}
	q := qolComp.(*qol.QoLComponent)

	if q.PlayerID != 99999 {
		t.Errorf("PlayerID = %d, want %d", q.PlayerID, 99999)
	}
	if !q.AutoLootEnabled {
		t.Error("AutoLootEnabled should be true")
	}
	if q.AutoLootRadius != 200.0 {
		t.Errorf("AutoLootRadius = %f, want 200.0", q.AutoLootRadius)
	}
	if q.SortPreset != "rarity" {
		t.Errorf("SortPreset = %q, want %q", q.SortPreset, "rarity")
	}
}

// TestDeserializeQoLStateNilData tests QoL deserialization with nil data.
func TestDeserializeQoLStateNilData(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{QoLData: nil}

	// Should not panic
	deserializeQoLState(player, state)

	// No QoL component should be added
	_, ok := player.GetComponent("qol")
	if ok {
		t.Error("QoL component should not be created with nil data")
	}
}

// TestDeserializeEquipment tests equipment deserialization with all slots.
func TestDeserializeEquipment(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(engine.NewEquipmentComponent())

	weaponData := saveload.ItemData{
		ID:   "test-weapon",
		Name: "Test Sword",
		Type: "weapon",
	}
	armorData := saveload.ItemData{
		ID:   "test-armor",
		Name: "Test Plate",
		Type: "armor",
	}
	accessoryData := saveload.ItemData{
		ID:   "test-accessory",
		Name: "Test Ring",
		Type: "accessory",
	}

	state := &saveload.PlayerState{
		EquippedItems: saveload.EquipmentData{
			Weapon:    &weaponData,
			Armor:     &armorData,
			Accessory: &accessoryData,
		},
	}

	deserializeEquipment(player, state)

	equipComp, ok := player.GetComponent("equipment")
	if !ok {
		t.Fatal("Equipment component not found")
	}
	equipment := equipComp.(*engine.EquipmentComponent)

	if weapon := equipment.Slots[engine.SlotMainHand]; weapon == nil || weapon.Name != "Test Sword" {
		t.Error("Weapon not deserialized correctly")
	}
	if armor := equipment.Slots[engine.SlotChest]; armor == nil || armor.Name != "Test Plate" {
		t.Error("Armor not deserialized correctly")
	}
	if accessory := equipment.Slots[engine.SlotAccessory1]; accessory == nil || accessory.Name != "Test Ring" {
		t.Error("Accessory not deserialized correctly")
	}
	if !equipment.StatsDirty {
		t.Error("StatsDirty should be true after deserialization")
	}
}

// TestDeserializeEquipmentNoComponent tests equipment deserialization without component.
func TestDeserializeEquipmentNoComponent(t *testing.T) {
	player := engine.NewEntity(1)

	weaponData := saveload.ItemData{ID: "test-weapon", Name: "Test Sword"}
	state := &saveload.PlayerState{
		EquippedItems: saveload.EquipmentData{
			Weapon: &weaponData,
		},
	}

	// Should not panic
	deserializeEquipment(player, state)
}

// TestDeserializeEquipmentPartialSlots tests equipment deserialization with partial slots.
func TestDeserializeEquipmentPartialSlots(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(engine.NewEquipmentComponent())

	weaponData := saveload.ItemData{ID: "test-weapon", Name: "Test Sword"}
	state := &saveload.PlayerState{
		EquippedItems: saveload.EquipmentData{
			Weapon:    &weaponData,
			Armor:     nil,
			Accessory: nil,
		},
	}

	deserializeEquipment(player, state)

	equipComp, _ := player.GetComponent("equipment")
	equipment := equipComp.(*engine.EquipmentComponent)

	if weapon := equipment.Slots[engine.SlotMainHand]; weapon == nil || weapon.Name != "Test Sword" {
		t.Error("Weapon not deserialized correctly")
	}
	if equipment.Slots[engine.SlotChest] != nil {
		t.Error("Armor should be nil when not provided")
	}
	if equipment.Slots[engine.SlotAccessory1] != nil {
		t.Error("Accessory should be nil when not provided")
	}
}

// TestDeserializeManaAndSpells tests mana and spell deserialization.
func TestDeserializeManaAndSpells(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.ManaComponent{Current: 0, Max: 0})
	player.AddComponent(&engine.SpellSlotComponent{})

	spell1Data := saveload.SpellData{Name: "Fireball", ManaCost: 25}
	spell2Data := saveload.SpellData{Name: "Ice Bolt", ManaCost: 20}

	state := &saveload.PlayerState{
		CurrentMana: 85,
		MaxMana:     150,
		Spells:      []saveload.SpellData{spell1Data, spell2Data},
	}

	deserializeManaAndSpells(player, state)

	manaComp, ok := player.GetComponent("mana")
	if !ok {
		t.Fatal("Mana component not found")
	}
	mana := manaComp.(*engine.ManaComponent)

	if mana.Current != 85 {
		t.Errorf("Mana.Current = %d, want 85", mana.Current)
	}
	if mana.Max != 150 {
		t.Errorf("Mana.Max = %d, want 150", mana.Max)
	}

	slotsComp, ok := player.GetComponent("spell_slots")
	if !ok {
		t.Fatal("Spell slots component not found")
	}
	slots := slotsComp.(*engine.SpellSlotComponent)

	if slots.Slots[0] == nil || slots.Slots[0].Name != "Fireball" {
		t.Error("Spell slot 0 not deserialized correctly")
	}
	if slots.Slots[1] == nil || slots.Slots[1].Name != "Ice Bolt" {
		t.Error("Spell slot 1 not deserialized correctly")
	}
}

// TestDeserializeManaAndSpellsNoComponent tests deserialization without components.
func TestDeserializeManaAndSpellsNoComponent(t *testing.T) {
	player := engine.NewEntity(1)
	state := &saveload.PlayerState{CurrentMana: 100, MaxMana: 200}

	// Should not panic
	deserializeManaAndSpells(player, state)
}

// TestDeserializeManaAndSpellsMoreThanFiveSpells tests spell slot limit.
func TestDeserializeManaAndSpellsMoreThanFiveSpells(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&engine.ManaComponent{})
	player.AddComponent(&engine.SpellSlotComponent{})

	// Create 7 spells
	var spells []saveload.SpellData
	for i := 0; i < 7; i++ {
		spells = append(spells, saveload.SpellData{
			Name: "Spell " + string(rune('A'+i)),
		})
	}

	state := &saveload.PlayerState{Spells: spells}
	deserializeManaAndSpells(player, state)

	slotsComp, _ := player.GetComponent("spell_slots")
	slots := slotsComp.(*engine.SpellSlotComponent)

	// Only first 5 should be set
	for i := 0; i < 5; i++ {
		if slots.Slots[i] == nil {
			t.Errorf("Spell slot %d should not be nil", i)
		}
	}
}

// TestApplyEquipmentLoadout tests that loadout bonuses are correctly applied to player stats.
func TestApplyEquipmentLoadout(t *testing.T) {
	logger, clientLogger := initializeLogger()

	tests := []struct {
		name            string
		loadout         *engine.EquipmentLoadout
		initialHP       float64
		initialAttack   float64
		initialDefense  float64
		expectedHP      float64
		expectedAttack  float64
		expectedDefense float64
	}{
		{
			name: "heavy armor loadout",
			loadout: &engine.EquipmentLoadout{
				Name:         "Heavy Armor",
				BonusHP:      50,
				BonusAttack:  10,
				BonusDefense: 20,
			},
			initialHP: 100, initialAttack: 10, initialDefense: 5,
			expectedHP: 150, expectedAttack: 20, expectedDefense: 25,
		},
		{
			name: "berserker loadout",
			loadout: &engine.EquipmentLoadout{
				Name:         "Berserker",
				BonusHP:      0,
				BonusAttack:  30,
				BonusDefense: -5,
			},
			initialHP: 100, initialAttack: 15, initialDefense: 10,
			expectedHP: 100, expectedAttack: 45, expectedDefense: 5,
		},
		{
			name: "balanced loadout",
			loadout: &engine.EquipmentLoadout{
				Name:         "Balanced",
				BonusHP:      25,
				BonusAttack:  15,
				BonusDefense: 10,
			},
			initialHP: 80, initialAttack: 12, initialDefense: 8,
			expectedHP: 105, expectedAttack: 27, expectedDefense: 18,
		},
		{
			name: "zero bonus loadout",
			loadout: &engine.EquipmentLoadout{
				Name:         "Basic",
				BonusHP:      0,
				BonusAttack:  0,
				BonusDefense: 0,
			},
			initialHP: 100, initialAttack: 10, initialDefense: 10,
			expectedHP: 100, expectedAttack: 10, expectedDefense: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := engine.NewEntity(1)
			player.AddComponent(&engine.HealthComponent{
				Current: tt.initialHP,
				Max:     tt.initialHP,
			})
			player.AddComponent(&engine.StatsComponent{
				Attack:  tt.initialAttack,
				Defense: tt.initialDefense,
			})

			charData := &engine.CharacterData{
				Name:            "TestPlayer",
				Class:           engine.ClassWarrior,
				StartingLoadout: tt.loadout,
			}

			applyEquipmentLoadout(player, charData, clientLogger)

			// Verify health
			healthComp, _ := player.GetComponent("health")
			health := healthComp.(*engine.HealthComponent)
			if health.Max != tt.expectedHP {
				t.Errorf("Max HP = %v, want %v", health.Max, tt.expectedHP)
			}
			if health.Current != tt.expectedHP {
				t.Errorf("Current HP = %v, want %v", health.Current, tt.expectedHP)
			}

			// Verify stats
			statsComp, _ := player.GetComponent("stats")
			stats := statsComp.(*engine.StatsComponent)
			if stats.Attack != tt.expectedAttack {
				t.Errorf("Attack = %v, want %v", stats.Attack, tt.expectedAttack)
			}
			if stats.Defense != tt.expectedDefense {
				t.Errorf("Defense = %v, want %v", stats.Defense, tt.expectedDefense)
			}
		})
	}
	_ = logger // Suppress unused warning
}

// TestApplyEquipmentLoadoutNilLoadout tests that nil loadout is handled gracefully.
func TestApplyEquipmentLoadoutNilLoadout(t *testing.T) {
	logger, clientLogger := initializeLogger()

	player := engine.NewEntity(1)
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.StatsComponent{Attack: 10, Defense: 10})

	charData := &engine.CharacterData{
		Name:            "TestPlayer",
		Class:           engine.ClassWarrior,
		StartingLoadout: nil, // No loadout selected
	}

	// Should not panic and not modify stats
	applyEquipmentLoadout(player, charData, clientLogger)

	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*engine.HealthComponent)
	if health.Max != 100 {
		t.Errorf("Max HP modified when loadout was nil: got %v, want 100", health.Max)
	}

	statsComp, _ := player.GetComponent("stats")
	stats := statsComp.(*engine.StatsComponent)
	if stats.Attack != 10 || stats.Defense != 10 {
		t.Errorf("Stats modified when loadout was nil: attack=%v, defense=%v", stats.Attack, stats.Defense)
	}
	_ = logger
}

// TestApplyEquipmentLoadoutMissingComponents tests handling of missing components.
func TestApplyEquipmentLoadoutMissingComponents(t *testing.T) {
	logger, clientLogger := initializeLogger()

	loadout := &engine.EquipmentLoadout{
		Name:         "Test",
		BonusHP:      50,
		BonusAttack:  20,
		BonusDefense: 15,
	}
	charData := &engine.CharacterData{
		Name:            "TestPlayer",
		Class:           engine.ClassWarrior,
		StartingLoadout: loadout,
	}

	t.Run("missing health component", func(t *testing.T) {
		player := engine.NewEntity(1)
		player.AddComponent(&engine.StatsComponent{Attack: 10, Defense: 10})

		// Should not panic
		applyEquipmentLoadout(player, charData, clientLogger)

		statsComp, _ := player.GetComponent("stats")
		stats := statsComp.(*engine.StatsComponent)
		if stats.Attack != 30 {
			t.Errorf("Attack not applied: got %v, want 30", stats.Attack)
		}
	})

	t.Run("missing stats component", func(t *testing.T) {
		player := engine.NewEntity(2)
		player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})

		// Should not panic
		applyEquipmentLoadout(player, charData, clientLogger)

		healthComp, _ := player.GetComponent("health")
		health := healthComp.(*engine.HealthComponent)
		if health.Max != 150 {
			t.Errorf("HP not applied: got %v, want 150", health.Max)
		}
	})

	t.Run("no components at all", func(t *testing.T) {
		player := engine.NewEntity(3)

		// Should not panic
		applyEquipmentLoadout(player, charData, clientLogger)
	})
	_ = logger
}

// TestApplyEquipmentLoadoutNegativeBonuses tests that negative bonuses work correctly.
func TestApplyEquipmentLoadoutNegativeBonuses(t *testing.T) {
	logger, clientLogger := initializeLogger()

	player := engine.NewEntity(1)
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.StatsComponent{Attack: 20, Defense: 15})

	charData := &engine.CharacterData{
		Name:  "TestPlayer",
		Class: engine.ClassWarrior,
		StartingLoadout: &engine.EquipmentLoadout{
			Name:         "Glass Cannon",
			BonusHP:      -20,
			BonusAttack:  40,
			BonusDefense: -10,
		},
	}

	applyEquipmentLoadout(player, charData, clientLogger)

	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*engine.HealthComponent)
	if health.Max != 80 {
		t.Errorf("Max HP with negative bonus = %v, want 80", health.Max)
	}

	statsComp, _ := player.GetComponent("stats")
	stats := statsComp.(*engine.StatsComponent)
	if stats.Attack != 60 {
		t.Errorf("Attack = %v, want 60", stats.Attack)
	}
	if stats.Defense != 5 {
		t.Errorf("Defense with negative bonus = %v, want 5", stats.Defense)
	}
	_ = logger
}

// BenchmarkApplyEquipmentLoadout benchmarks loadout application performance.
func BenchmarkApplyEquipmentLoadout(b *testing.B) {
	logger, clientLogger := initializeLogger()

	loadout := &engine.EquipmentLoadout{
		Name:         "Heavy Armor",
		BonusHP:      50,
		BonusAttack:  10,
		BonusDefense: 20,
	}
	charData := &engine.CharacterData{
		Name:            "BenchmarkPlayer",
		Class:           engine.ClassWarrior,
		StartingLoadout: loadout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player := engine.NewEntity(uint64(i))
		player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
		player.AddComponent(&engine.StatsComponent{Attack: 10, Defense: 10})
		applyEquipmentLoadout(player, charData, clientLogger)
	}
	_ = logger
}

// TestPrestigeEntityAdapterGetID tests the GetID method of prestigeEntityAdapter.
func TestPrestigeEntityAdapterGetID(t *testing.T) {
	tests := []struct {
		name     string
		entityID uint64
		expected string
	}{
		{"single digit", 1, "1"},
		{"double digit", 42, "42"},
		{"large id", 123456789, "123456789"},
		{"zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := engine.NewEntity(tt.entityID)
			adapter := &prestigeEntityAdapter{entity: entity}

			if got := adapter.GetID(); got != tt.expected {
				t.Errorf("GetID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestPrestigeEntityAdapterHasComponent tests the HasComponent method.
func TestPrestigeEntityAdapterHasComponent(t *testing.T) {
	entity := engine.NewEntity(1)
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	adapter := &prestigeEntityAdapter{entity: entity}

	tests := []struct {
		name          string
		componentType string
		expected      bool
	}{
		{"has position", "position", true},
		{"has health", "health", true},
		{"no velocity", "velocity", false},
		{"no stats", "stats", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adapter.HasComponent(tt.componentType); got != tt.expected {
				t.Errorf("HasComponent(%q) = %v, want %v", tt.componentType, got, tt.expected)
			}
		})
	}
}

// TestPrestigeEntityAdapterGetComponent tests the GetComponent method.
func TestPrestigeEntityAdapterGetComponent(t *testing.T) {
	entity := engine.NewEntity(1)
	posComp := &engine.PositionComponent{X: 100, Y: 200}
	entity.AddComponent(posComp)
	adapter := &prestigeEntityAdapter{entity: entity}

	t.Run("existing component", func(t *testing.T) {
		comp := adapter.GetComponent("position")
		if comp == nil {
			t.Fatal("GetComponent returned nil for existing component")
		}
		pos, ok := comp.(*engine.PositionComponent)
		if !ok {
			t.Fatalf("Component type mismatch: got %T", comp)
		}
		if pos.X != 100 || pos.Y != 200 {
			t.Errorf("Position = (%v, %v), want (100, 200)", pos.X, pos.Y)
		}
	})

	t.Run("non-existing component", func(t *testing.T) {
		comp := adapter.GetComponent("velocity")
		if comp != nil {
			t.Errorf("GetComponent returned non-nil for non-existing component: %v", comp)
		}
	})
}

// TestPrestigeEntityAdapterAddComponent tests the AddComponent method.
func TestPrestigeEntityAdapterAddComponent(t *testing.T) {
	entity := engine.NewEntity(1)
	adapter := &prestigeEntityAdapter{entity: entity}

	// Verify no velocity initially
	if adapter.HasComponent("velocity") {
		t.Fatal("Entity should not have velocity initially")
	}

	// Add component via adapter
	velComp := &engine.VelocityComponent{VX: 10, VY: 20}
	adapter.AddComponent(velComp)

	// Verify component was added
	if !adapter.HasComponent("velocity") {
		t.Error("Component was not added")
	}

	comp := adapter.GetComponent("velocity")
	if comp == nil {
		t.Fatal("GetComponent returned nil after AddComponent")
	}
	vel, ok := comp.(*engine.VelocityComponent)
	if !ok {
		t.Fatalf("Component type mismatch: got %T", comp)
	}
	if vel.VX != 10 || vel.VY != 20 {
		t.Errorf("Velocity = (%v, %v), want (10, 20)", vel.VX, vel.VY)
	}
}

// TestPrestigeEntityAdapterRemoveComponent tests the RemoveComponent method.
func TestPrestigeEntityAdapterRemoveComponent(t *testing.T) {
	entity := engine.NewEntity(1)
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	adapter := &prestigeEntityAdapter{entity: entity}

	// Verify position exists
	if !adapter.HasComponent("position") {
		t.Fatal("Entity should have position initially")
	}

	// Remove position
	adapter.RemoveComponent("position")

	// Verify position was removed
	if adapter.HasComponent("position") {
		t.Error("Position component was not removed")
	}

	// Verify health is still there
	if !adapter.HasComponent("health") {
		t.Error("Health component should still exist")
	}
}

// TestPrestigeEntityAdapterRemoveNonExistingComponent tests removing a non-existing component.
func TestPrestigeEntityAdapterRemoveNonExistingComponent(t *testing.T) {
	entity := engine.NewEntity(1)
	adapter := &prestigeEntityAdapter{entity: entity}

	// Should not panic when removing non-existing component
	adapter.RemoveComponent("nonexistent")
}

// BenchmarkPrestigeEntityAdapterGetID benchmarks the GetID method.
func BenchmarkPrestigeEntityAdapterGetID(b *testing.B) {
	entity := engine.NewEntity(123456789)
	adapter := &prestigeEntityAdapter{entity: entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.GetID()
	}
}

// BenchmarkPrestigeEntityAdapterHasComponent benchmarks the HasComponent method.
func BenchmarkPrestigeEntityAdapterHasComponent(b *testing.B) {
	entity := engine.NewEntity(1)
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	adapter := &prestigeEntityAdapter{entity: entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.HasComponent("position")
	}
}

// mockDisconnector is a test mock for testing cleanupNetworkClient.
type mockDisconnector struct {
	disconnectCalled bool
	disconnectErr    error
}

func (m *mockDisconnector) Disconnect() error {
	m.disconnectCalled = true
	return m.disconnectErr
}

// createHandlerTestLogger creates a logger for handler tests with suppressed output.
func createHandlerTestLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.WarnLevel)
	return logger.WithField("test", "handlers")
}

// TestCleanupNetworkClientNil tests cleanup with nil client.
func TestCleanupNetworkClientNil(t *testing.T) {
	logger := createHandlerTestLogger()
	// Should not panic
	cleanupNetworkClient(nil, logger)
}

// TestCleanupNetworkClientSuccess tests successful cleanup.
func TestCleanupNetworkClientSuccess(t *testing.T) {
	logger := createHandlerTestLogger()
	mock := &mockDisconnector{}

	cleanupNetworkClient(mock, logger)

	if !mock.disconnectCalled {
		t.Error("Disconnect() was not called")
	}
}

// TestCleanupNetworkClientError tests cleanup with disconnect error.
func TestCleanupNetworkClientError(t *testing.T) {
	logger := createHandlerTestLogger()
	mock := &mockDisconnector{disconnectErr: errors.New("disconnect failed")}

	// Should not panic, just log warning
	cleanupNetworkClient(mock, logger)

	if !mock.disconnectCalled {
		t.Error("Disconnect() was not called despite error")
	}
}

// TestCleanupNetworkClientNoDisconnector tests cleanup with non-disconnector interface.
func TestCleanupNetworkClientNoDisconnector(t *testing.T) {
	logger := createHandlerTestLogger()
	// Pass an int - it doesn't implement disconnector
	cleanupNetworkClient(42, logger)
	// Should complete without panic
}

// TestCleanupNetworkClientStruct tests cleanup with struct not implementing disconnector.
func TestCleanupNetworkClientStruct(t *testing.T) {
	logger := createHandlerTestLogger()
	type noMethodsStruct struct {
		value int
	}
	cleanupNetworkClient(&noMethodsStruct{value: 1}, logger)
	// Should complete without panic
}

// TestConnectUIComponents tests the connectUIComponents helper function.
func TestConnectUIComponents(t *testing.T) {
	// Track which setters were called
	var calledSetters []int

	connections := []struct {
		ui     interface{}
		setter func(interface{})
	}{
		{ui: "ui1", setter: func(ui interface{}) { calledSetters = append(calledSetters, 0) }},
		{ui: nil, setter: func(ui interface{}) { calledSetters = append(calledSetters, 1) }}, // nil should be skipped
		{ui: "ui2", setter: func(ui interface{}) { calledSetters = append(calledSetters, 2) }},
		{ui: nil, setter: func(ui interface{}) { calledSetters = append(calledSetters, 3) }}, // nil should be skipped
	}

	connectUIComponents(connections)

	// Should only call setters for non-nil UIs (indices 0 and 2)
	if len(calledSetters) != 2 {
		t.Errorf("expected 2 setter calls, got %d", len(calledSetters))
	}
	if len(calledSetters) >= 1 && calledSetters[0] != 0 {
		t.Errorf("expected first call to setter 0, got %d", calledSetters[0])
	}
	if len(calledSetters) >= 2 && calledSetters[1] != 2 {
		t.Errorf("expected second call to setter 2, got %d", calledSetters[1])
	}
}

// TestConnectUIComponentsEmpty tests with empty connections slice.
func TestConnectUIComponentsEmpty(t *testing.T) {
	connections := []struct {
		ui     interface{}
		setter func(interface{})
	}{}

	// Should not panic with empty slice
	connectUIComponents(connections)
}

// TestConnectUIComponentsAllNil tests with all nil UIs.
func TestConnectUIComponentsAllNil(t *testing.T) {
	var calledCount int

	connections := []struct {
		ui     interface{}
		setter func(interface{})
	}{
		{ui: nil, setter: func(ui interface{}) { calledCount++ }},
		{ui: nil, setter: func(ui interface{}) { calledCount++ }},
		{ui: nil, setter: func(ui interface{}) { calledCount++ }},
	}

	connectUIComponents(connections)

	if calledCount != 0 {
		t.Errorf("expected 0 setter calls for all nil UIs, got %d", calledCount)
	}
}

// TestConnectUIComponentsCorrectValuePassed tests that correct UI value is passed to setter.
func TestConnectUIComponentsCorrectValuePassed(t *testing.T) {
	var receivedValues []interface{}

	connections := []struct {
		ui     interface{}
		setter func(interface{})
	}{
		{ui: "first", setter: func(ui interface{}) { receivedValues = append(receivedValues, ui) }},
		{ui: 42, setter: func(ui interface{}) { receivedValues = append(receivedValues, ui) }},
		{ui: nil, setter: func(ui interface{}) { receivedValues = append(receivedValues, ui) }},
		{ui: true, setter: func(ui interface{}) { receivedValues = append(receivedValues, ui) }},
	}

	connectUIComponents(connections)

	if len(receivedValues) != 3 {
		t.Fatalf("expected 3 values, got %d", len(receivedValues))
	}

	if receivedValues[0] != "first" {
		t.Errorf("expected 'first', got %v", receivedValues[0])
	}
	if receivedValues[1] != 42 {
		t.Errorf("expected 42, got %v", receivedValues[1])
	}
	if receivedValues[2] != true {
		t.Errorf("expected true, got %v", receivedValues[2])
	}
}

// BenchmarkConnectUIComponents benchmarks the connectUIComponents function.
func BenchmarkConnectUIComponents(b *testing.B) {
	connections := []struct {
		ui     interface{}
		setter func(interface{})
	}{
		{ui: "ui1", setter: func(ui interface{}) {}},
		{ui: nil, setter: func(ui interface{}) {}},
		{ui: "ui2", setter: func(ui interface{}) {}},
		{ui: "ui3", setter: func(ui interface{}) {}},
		{ui: nil, setter: func(ui interface{}) {}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectUIComponents(connections)
	}
}

// TestSerializeQoLStateWithCraftQueue tests QoL state serialization with craft queue items.
func TestSerializeQoLStateWithCraftQueue(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&qol.QoLComponent{
		PlayerID:        12345,
		AutoLootEnabled: true,
		AutoLootRadius:  150.0,
		SortPreset:      "type",
		MountWhistle:    true,
		RecipeTracking:  true,
		CraftQueue: []*qol.CraftQueueEntry{
			{RecipeID: "sword_iron", Quantity: 2, MaterialsReady: true, Position: 0},
			{RecipeID: "potion_health", Quantity: 5, MaterialsReady: false, Position: 1},
		},
	})

	state := &saveload.PlayerState{}
	serializeQoLState(player, state)

	if state.QoLData == nil {
		t.Fatal("QoLData should not be nil")
	}
	if len(state.QoLData.CraftQueueJSON) == 0 {
		t.Error("CraftQueueJSON should not be empty when craft queue has items")
	}
	// Verify JSON is valid by checking it contains expected recipe IDs
	jsonStr := string(state.QoLData.CraftQueueJSON)
	if !contains(jsonStr, "sword_iron") {
		t.Error("CraftQueueJSON should contain 'sword_iron'")
	}
	if !contains(jsonStr, "potion_health") {
		t.Error("CraftQueueJSON should contain 'potion_health'")
	}
}

// contains is a helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestDeserializeQoLStateWithCraftQueue tests QoL deserialization with craft queue JSON.
func TestDeserializeQoLStateWithCraftQueue(t *testing.T) {
	player := engine.NewEntity(1)
	craftQueueJSON := []byte(`[{"RecipeID":"sword_iron","Quantity":2,"MaterialsReady":true,"Position":0},{"RecipeID":"potion_health","Quantity":5,"MaterialsReady":false,"Position":1}]`)

	state := &saveload.PlayerState{
		QoLData: &saveload.QoLStateData{
			PlayerID:        99999,
			AutoLootEnabled: true,
			AutoLootRadius:  200.0,
			SortPreset:      "rarity",
			MountWhistle:    true,
			RecipeTracking:  true,
			CraftQueueJSON:  craftQueueJSON,
		},
	}

	deserializeQoLState(player, state)

	qolComp, ok := player.GetComponent("qol")
	if !ok {
		t.Fatal("QoL component should be created")
	}
	q, ok := qolComp.(*qol.QoLComponent)
	if !ok {
		t.Fatal("Component should be QoLComponent type")
	}

	if len(q.CraftQueue) != 2 {
		t.Errorf("CraftQueue should have 2 items, got %d", len(q.CraftQueue))
	}
	if len(q.CraftQueue) > 0 && q.CraftQueue[0].RecipeID != "sword_iron" {
		t.Errorf("First recipe should be 'sword_iron', got %q", q.CraftQueue[0].RecipeID)
	}
	if len(q.CraftQueue) > 1 && q.CraftQueue[1].Quantity != 5 {
		t.Errorf("Second recipe quantity should be 5, got %d", q.CraftQueue[1].Quantity)
	}
}

// TestDeserializeQoLStateExistingComponent tests deserialization when entity already has QoL component.
func TestDeserializeQoLStateExistingComponent(t *testing.T) {
	player := engine.NewEntity(1)
	// Pre-add QoL component with different values
	existingQoL := &qol.QoLComponent{
		PlayerID:        11111,
		AutoLootEnabled: false,
		AutoLootRadius:  50.0,
		SortPreset:      "name",
		MountWhistle:    false,
		RecipeTracking:  false,
	}
	player.AddComponent(existingQoL)

	state := &saveload.PlayerState{
		QoLData: &saveload.QoLStateData{
			PlayerID:        99999,
			AutoLootEnabled: true,
			AutoLootRadius:  200.0,
			SortPreset:      "rarity",
			MountWhistle:    true,
			RecipeTracking:  true,
		},
	}

	deserializeQoLState(player, state)

	qolComp, ok := player.GetComponent("qol")
	if !ok {
		t.Fatal("QoL component should exist")
	}
	q, ok := qolComp.(*qol.QoLComponent)
	if !ok {
		t.Fatal("Component should be QoLComponent type")
	}

	// Values should be overwritten from state
	if q.PlayerID != 99999 {
		t.Errorf("PlayerID = %d, want 99999", q.PlayerID)
	}
	if !q.AutoLootEnabled {
		t.Error("AutoLootEnabled should be true after deserialization")
	}
	if q.AutoLootRadius != 200.0 {
		t.Errorf("AutoLootRadius = %f, want 200.0", q.AutoLootRadius)
	}
	if q.SortPreset != "rarity" {
		t.Errorf("SortPreset = %q, want 'rarity'", q.SortPreset)
	}
}

// TestDeserializeQoLStateInvalidCraftQueue tests deserialization with invalid JSON gracefully ignores bad data.
func TestDeserializeQoLStateInvalidCraftQueue(t *testing.T) {
	player := engine.NewEntity(1)
	// Invalid JSON should not cause panic, just skip craft queue restoration
	invalidJSON := []byte(`[{"invalid json missing closing bracket`)

	state := &saveload.PlayerState{
		QoLData: &saveload.QoLStateData{
			PlayerID:        99999,
			AutoLootEnabled: true,
			CraftQueueJSON:  invalidJSON,
		},
	}

	// Should not panic
	deserializeQoLState(player, state)

	qolComp, ok := player.GetComponent("qol")
	if !ok {
		t.Fatal("QoL component should be created")
	}
	q, ok := qolComp.(*qol.QoLComponent)
	if !ok {
		t.Fatal("Component should be QoLComponent type")
	}

	// Other fields should still be set
	if q.PlayerID != 99999 {
		t.Errorf("PlayerID = %d, want 99999", q.PlayerID)
	}
	// CraftQueue should be nil/empty due to invalid JSON
	if len(q.CraftQueue) != 0 {
		t.Errorf("CraftQueue should be empty with invalid JSON, got %d items", len(q.CraftQueue))
	}
}

// TestSerializeQoLStateEmptyCraftQueue tests serialization with empty craft queue.
func TestSerializeQoLStateEmptyCraftQueue(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&qol.QoLComponent{
		PlayerID:        12345,
		AutoLootEnabled: true,
		CraftQueue:      []*qol.CraftQueueEntry{}, // Empty slice
	})

	state := &saveload.PlayerState{}
	serializeQoLState(player, state)

	if state.QoLData == nil {
		t.Fatal("QoLData should not be nil")
	}
	// Empty craft queue should not produce JSON
	if len(state.QoLData.CraftQueueJSON) != 0 {
		t.Errorf("CraftQueueJSON should be empty for empty craft queue, got %d bytes", len(state.QoLData.CraftQueueJSON))
	}
}

// wrongTypeComponent is a component with wrong type for testing type assertion failures.
type wrongTypeComponent struct{}

func (w *wrongTypeComponent) Type() string { return "qol" }

// wrongPositionComponent returns "position" type but isn't PositionComponent.
type wrongPositionComponent struct{}

func (w *wrongPositionComponent) Type() string { return "position" }

// wrongHealthComponent returns "health" type but isn't HealthComponent.
type wrongHealthComponent struct{}

func (w *wrongHealthComponent) Type() string { return "health" }

// wrongStatsComponent returns "stats" type but isn't StatsComponent.
type wrongStatsComponent struct{}

func (w *wrongStatsComponent) Type() string { return "stats" }

// wrongExperienceComponent returns "experience" type but isn't ExperienceComponent.
type wrongExperienceComponent struct{}

func (w *wrongExperienceComponent) Type() string { return "experience" }

// wrongInventoryComponent returns "inventory" type but isn't InventoryComponent.
type wrongInventoryComponent struct{}

func (w *wrongInventoryComponent) Type() string { return "inventory" }

// wrongEquipmentComponent returns "equipment" type but isn't EquipmentComponent.
type wrongEquipmentComponent struct{}

func (w *wrongEquipmentComponent) Type() string { return "equipment" }

// wrongManaComponent returns "mana" type but isn't ManaComponent.
type wrongManaComponent struct{}

func (w *wrongManaComponent) Type() string { return "mana" }

// wrongSpellSlotComponent returns "spell_slots" type but isn't SpellSlotComponent.
type wrongSpellSlotComponent struct{}

func (w *wrongSpellSlotComponent) Type() string { return "spell_slots" }

// TestSerializeQoLStateWrongComponentType tests serialization with wrong component type.
func TestSerializeQoLStateWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	// Add a component that returns "qol" type but isn't actually a QoLComponent
	player.AddComponent(&wrongTypeComponent{})

	state := &saveload.PlayerState{}
	serializeQoLState(player, state)

	// Should return early and not set QoLData
	if state.QoLData != nil {
		t.Error("QoLData should be nil when component type assertion fails")
	}
}

// TestDeserializeQoLStateWrongComponentType tests deserialization with wrong component type.
func TestDeserializeQoLStateWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	// Add a component that returns "qol" type but isn't actually a QoLComponent
	player.AddComponent(&wrongTypeComponent{})

	state := &saveload.PlayerState{
		QoLData: &saveload.QoLStateData{
			PlayerID:        99999,
			AutoLootEnabled: true,
		},
	}

	// Should not panic, just return early
	deserializeQoLState(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("qol")
	if !ok {
		t.Fatal("Component should still exist")
	}
	// Should still be the wrong type
	_, isWrong := comp.(*wrongTypeComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializePositionWrongComponentType tests position deserialization with wrong component type.
func TestDeserializePositionWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongPositionComponent{})

	state := &saveload.PlayerState{X: 100, Y: 200}

	// Should not panic, just return early
	deserializePosition(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("position")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongPositionComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeHealthWrongComponentType tests health deserialization with wrong component type.
func TestDeserializeHealthWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongHealthComponent{})

	state := &saveload.PlayerState{CurrentHealth: 50, MaxHealth: 100}

	// Should not panic, just return early
	deserializeHealth(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("health")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongHealthComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeStatsWrongComponentType tests stats deserialization with wrong component type.
func TestDeserializeStatsWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongStatsComponent{})

	state := &saveload.PlayerState{Attack: 10, Defense: 5, MagicPower: 8}

	// Should not panic, just return early
	deserializeStats(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("stats")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongStatsComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeExperienceWrongComponentType tests experience deserialization with wrong component type.
func TestDeserializeExperienceWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongExperienceComponent{})

	state := &saveload.PlayerState{Level: 5, Experience: 1500}

	// Should not panic, just return early
	deserializeExperience(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("experience")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongExperienceComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeInventoryWrongComponentType tests inventory deserialization with wrong component type.
func TestDeserializeInventoryWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongInventoryComponent{})

	state := &saveload.PlayerState{Gold: 500, Items: []saveload.ItemData{}}

	// Should not panic, just return early
	deserializeInventory(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("inventory")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongInventoryComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeEquipmentWrongComponentType tests equipment deserialization with wrong component type.
func TestDeserializeEquipmentWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongEquipmentComponent{})

	weaponData := saveload.ItemData{ID: "test", Name: "Test", Type: "weapon"}
	state := &saveload.PlayerState{
		EquippedItems: saveload.EquipmentData{Weapon: &weaponData},
	}

	// Should not panic, just return early
	deserializeEquipment(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("equipment")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongEquipmentComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeManaAndSpellsWrongManaType tests mana deserialization with wrong component type.
func TestDeserializeManaAndSpellsWrongManaType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongManaComponent{})

	state := &saveload.PlayerState{CurrentMana: 50, MaxMana: 100}

	// Should not panic, just return early
	deserializeManaAndSpells(player, state)

	// The wrong component should still be there (not replaced)
	comp, ok := player.GetComponent("mana")
	if !ok {
		t.Fatal("Component should still exist")
	}
	_, isWrong := comp.(*wrongManaComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestDeserializeManaAndSpellsWrongSpellSlotType tests spell slot deserialization with wrong component type.
func TestDeserializeManaAndSpellsWrongSpellSlotType(t *testing.T) {
	player := engine.NewEntity(1)
	// Add correct mana component but wrong spell slots component
	player.AddComponent(&engine.ManaComponent{Current: 0, Max: 100})
	player.AddComponent(&wrongSpellSlotComponent{})

	state := &saveload.PlayerState{
		CurrentMana: 50,
		MaxMana:     100,
		Spells:      []saveload.SpellData{{Name: "Fireball", ManaCost: 25}},
	}

	// Should not panic, just return early when spell slots assertion fails
	deserializeManaAndSpells(player, state)

	// Mana should be updated (correct type)
	manaComp, _ := player.GetComponent("mana")
	mana := manaComp.(*engine.ManaComponent)
	if mana.Current != 50 || mana.Max != 100 {
		t.Errorf("Mana = %d/%d, want 50/100", mana.Current, mana.Max)
	}

	// Spell slots should still be wrong type (not replaced)
	comp, ok := player.GetComponent("spell_slots")
	if !ok {
		t.Fatal("Spell slot component should still exist")
	}
	_, isWrong := comp.(*wrongSpellSlotComponent)
	if !isWrong {
		t.Error("Component should still be wrong type (not replaced)")
	}
}

// TestSerializeExperienceWrongComponentType tests experience serialization with wrong component type.
func TestSerializeExperienceWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongExperienceComponent{})

	state := &saveload.PlayerState{}

	// Should not panic, just return early
	serializeExperience(player, state)

	// Values should not be set (remain default)
	if state.Level != 0 || state.Experience != 0 {
		t.Errorf("Level/Experience = %d/%d, want 0/0", state.Level, state.Experience)
	}
}

// TestSerializeInventoryWrongComponentType tests inventory serialization with wrong component type.
func TestSerializeInventoryWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongInventoryComponent{})

	state := &saveload.PlayerState{}

	// Should not panic, just return early
	serializeInventory(player, state)

	// Values should not be set (remain default)
	if state.Gold != 0 || state.Items != nil {
		t.Errorf("Gold/Items = %d/%v, want 0/nil", state.Gold, state.Items)
	}
}

// TestSerializeEquipmentWrongComponentType tests equipment serialization with wrong component type.
func TestSerializeEquipmentWrongComponentType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongEquipmentComponent{})

	state := &saveload.PlayerState{}

	// Should not panic, just return early
	serializeEquipment(player, state)

	// Values should not be set (remain default)
	if state.EquippedItems.Weapon != nil || state.EquippedItems.Armor != nil {
		t.Error("EquippedItems should remain empty")
	}
}

// TestSerializeManaAndSpellsWrongManaType tests mana serialization with wrong component type.
func TestSerializeManaAndSpellsWrongManaType(t *testing.T) {
	player := engine.NewEntity(1)
	player.AddComponent(&wrongManaComponent{})

	state := &saveload.PlayerState{}

	// Should not panic, just return early
	serializeManaAndSpells(player, state)

	// Values should not be set (remain default)
	if state.CurrentMana != 0 || state.MaxMana != 0 {
		t.Errorf("Mana = %d/%d, want 0/0", state.CurrentMana, state.MaxMana)
	}
}

// TestSerializeManaAndSpellsWrongSpellSlotType tests spell slot serialization with wrong component type.
func TestSerializeManaAndSpellsWrongSpellSlotType(t *testing.T) {
	player := engine.NewEntity(1)
	// Add correct mana component but wrong spell slots component
	player.AddComponent(&engine.ManaComponent{Current: 75, Max: 150})
	player.AddComponent(&wrongSpellSlotComponent{})

	state := &saveload.PlayerState{}

	// Should not panic, mana should be serialized, spells should not
	serializeManaAndSpells(player, state)

	// Mana values should be set (correct type)
	if state.CurrentMana != 75 || state.MaxMana != 150 {
		t.Errorf("Mana = %d/%d, want 75/150", state.CurrentMana, state.MaxMana)
	}

	// Spells should not be set (wrong type)
	if state.Spells != nil {
		t.Errorf("Spells should be nil, got %v", state.Spells)
	}
}

// TestFindNearestNPC tests finding the nearest NPC with dialog component.
func TestFindNearestNPC(t *testing.T) {
	tests := []struct {
		name        string
		playerX     float64
		playerY     float64
		npcs        []struct{ x, y float64 }
		expectedIdx int // -1 means no NPC expected
		noPlayer    bool
		noPlayerPos bool
	}{
		{
			name:        "single NPC in range",
			playerX:     100,
			playerY:     100,
			npcs:        []struct{ x, y float64 }{{120, 100}},
			expectedIdx: 0,
		},
		{
			name:        "multiple NPCs - returns nearest",
			playerX:     100,
			playerY:     100,
			npcs:        []struct{ x, y float64 }{{200, 100}, {130, 100}, {150, 100}},
			expectedIdx: 1, // 130,100 is closest
		},
		{
			name:        "NPC out of range",
			playerX:     0,
			playerY:     0,
			npcs:        []struct{ x, y float64 }{{500, 500}}, // > 5*32 = 160
			expectedIdx: -1,
		},
		{
			name:        "no NPCs",
			playerX:     100,
			playerY:     100,
			npcs:        nil,
			expectedIdx: -1,
		},
		{
			name:        "nil player",
			noPlayer:    true,
			expectedIdx: -1,
		},
		{
			name:        "player without position component",
			playerX:     100,
			playerY:     100,
			noPlayerPos: true,
			expectedIdx: -1,
		},
		{
			name:        "NPC at exact max range boundary",
			playerX:     0,
			playerY:     0,
			npcs:        []struct{ x, y float64 }{{159, 0}}, // just under 160
			expectedIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &engine.EbitenGame{
				World: engine.NewWorld(),
			}

			var player *engine.Entity
			if !tt.noPlayer {
				player = game.World.CreateEntity()
				if !tt.noPlayerPos {
					player.AddComponent(&engine.PositionComponent{X: tt.playerX, Y: tt.playerY})
				}
			}

			// Create NPCs with dialog components
			var npcEntities []*engine.Entity
			for _, npc := range tt.npcs {
				entity := game.World.CreateEntity()
				entity.AddComponent(&engine.PositionComponent{X: npc.x, Y: npc.y})
				entity.AddComponent(&engine.DialogComponent{})
				npcEntities = append(npcEntities, entity)
			}

			result := findNearestNPC(game, player)

			if tt.expectedIdx == -1 {
				if result != nil {
					t.Errorf("expected nil, got entity %d", result.ID)
				}
			} else {
				if result == nil {
					t.Error("expected NPC, got nil")
				} else if result.ID != npcEntities[tt.expectedIdx].ID {
					t.Errorf("expected NPC %d, got NPC %d", npcEntities[tt.expectedIdx].ID, result.ID)
				}
			}
		})
	}
}

// TestFindNearestNPCSkipsSelf tests that player entity is not returned as nearest NPC.
func TestFindNearestNPCSkipsSelf(t *testing.T) {
	game := &engine.EbitenGame{
		World: engine.NewWorld(),
	}

	player := game.World.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	// Give player a dialog component (unusual but should still be skipped)
	player.AddComponent(&engine.DialogComponent{})

	result := findNearestNPC(game, player)

	// Player should not be returned as nearest NPC (self)
	if result != nil && result.ID == player.ID {
		t.Error("findNearestNPC should not return the player entity itself")
	}
}

// TestFindNearestNPCEntitiesWithoutDialogIgnored tests that entities without dialog are ignored.
func TestFindNearestNPCEntitiesWithoutDialogIgnored(t *testing.T) {
	game := &engine.EbitenGame{
		World: engine.NewWorld(),
	}

	player := game.World.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	// Create nearby entity without dialog (should be ignored)
	nonNPC := game.World.CreateEntity()
	nonNPC.AddComponent(&engine.PositionComponent{X: 101, Y: 100})

	// Create farther entity with dialog (should be selected)
	npc := game.World.CreateEntity()
	npc.AddComponent(&engine.PositionComponent{X: 120, Y: 100})
	npc.AddComponent(&engine.DialogComponent{})

	result := findNearestNPC(game, player)

	if result == nil {
		t.Fatal("expected NPC, got nil")
	}
	if result.ID != npc.ID {
		t.Errorf("expected NPC %d, got entity %d", npc.ID, result.ID)
	}
}

// BenchmarkFindNearestNPC benchmarks NPC finding performance.
func BenchmarkFindNearestNPC(b *testing.B) {
	game := &engine.EbitenGame{
		World: engine.NewWorld(),
	}

	player := game.World.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 500, Y: 500})

	// Create 100 NPCs spread around
	for i := 0; i < 100; i++ {
		entity := game.World.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&engine.DialogComponent{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findNearestNPC(game, player)
	}
}

// TestSerializePlayerStateComplete tests full player state serialization.
func TestSerializePlayerStateComplete(t *testing.T) {
	player := engine.NewEntity(42)
	player.AddComponent(&engine.PositionComponent{X: 100.5, Y: 200.5})
	player.AddComponent(&engine.HealthComponent{Current: 75, Max: 100})
	player.AddComponent(&engine.StatsComponent{Attack: 20, Defense: 15, MagicPower: 25})
	player.AddComponent(&engine.ExperienceComponent{Level: 5, CurrentXP: 1500})
	player.AddComponent(&engine.InventoryComponent{Gold: 500})
	player.AddComponent(&engine.ManaComponent{Current: 50, Max: 100})
	player.AddComponent(&engine.EquipmentComponent{
		Slots: make(map[engine.EquipmentSlot]*item.Item),
	})
	player.AddComponent(&qol.QoLComponent{
		AutoLootEnabled: true,
		AutoLootRadius:  150.0,
	})

	game := &engine.EbitenGame{}

	state := serializePlayerState(player, game)

	// Verify all fields are serialized
	if state.EntityID != 42 {
		t.Errorf("EntityID = %d, want 42", state.EntityID)
	}
	if state.X != 100.5 || state.Y != 200.5 {
		t.Errorf("Position = (%f, %f), want (100.5, 200.5)", state.X, state.Y)
	}
	if state.CurrentHealth != 75 || state.MaxHealth != 100 {
		t.Errorf("Health = %f/%f, want 75/100", state.CurrentHealth, state.MaxHealth)
	}
	if state.Attack != 20 || state.Defense != 15 || state.MagicPower != 25 {
		t.Errorf("Stats = (%f, %f, %f), want (20, 15, 25)", state.Attack, state.Defense, state.MagicPower)
	}
	if state.Level != 5 || state.Experience != 1500 {
		t.Errorf("Experience = %d/%d, want 5/1500", state.Level, state.Experience)
	}
	if state.Gold != 500 {
		t.Errorf("Gold = %d, want 500", state.Gold)
	}
	if state.CurrentMana != 50 || state.MaxMana != 100 {
		t.Errorf("Mana = %d/%d, want 50/100", state.CurrentMana, state.MaxMana)
	}
	if state.Speed != 1.0 {
		t.Errorf("Speed = %f, want 1.0", state.Speed)
	}
	if state.QoLData == nil {
		t.Error("QoLData should not be nil")
	} else {
		if !state.QoLData.AutoLootEnabled {
			t.Error("QoLData.AutoLootEnabled should be true")
		}
		if state.QoLData.AutoLootRadius != 150.0 {
			t.Errorf("QoLData.AutoLootRadius = %f, want 150.0", state.QoLData.AutoLootRadius)
		}
	}
}

// TestSerializePlayerStateMinimal tests serialization with minimal components.
func TestSerializePlayerStateMinimal(t *testing.T) {
	player := engine.NewEntity(1)
	game := &engine.EbitenGame{}

	state := serializePlayerState(player, game)

	// Should still return valid state with default values
	if state == nil {
		t.Fatal("state should not be nil")
	}
	if state.EntityID != 1 {
		t.Errorf("EntityID = %d, want 1", state.EntityID)
	}
	// Speed should always be set to default
	if state.Speed != 1.0 {
		t.Errorf("Speed = %f, want 1.0", state.Speed)
	}
}

// TestDeserializePlayerStateComplete tests full player state deserialization.
func TestDeserializePlayerStateComplete(t *testing.T) {
	player := engine.NewEntity(42)
	player.AddComponent(&engine.PositionComponent{})
	player.AddComponent(&engine.HealthComponent{})
	player.AddComponent(&engine.StatsComponent{})
	player.AddComponent(&engine.ExperienceComponent{})
	player.AddComponent(&engine.InventoryComponent{})
	player.AddComponent(&engine.ManaComponent{})
	player.AddComponent(&engine.SpellSlotComponent{})
	player.AddComponent(&engine.EquipmentComponent{
		Slots: make(map[engine.EquipmentSlot]*item.Item),
	})
	player.AddComponent(&qol.QoLComponent{})

	state := &saveload.PlayerState{
		EntityID:      42,
		X:             100.5,
		Y:             200.5,
		CurrentHealth: 75,
		MaxHealth:     100,
		Attack:        20,
		Defense:       15,
		MagicPower:    25,
		Level:         5,
		Experience:    1500,
		Gold:          500,
		CurrentMana:   50,
		MaxMana:       100,
		QoLData: &saveload.QoLStateData{
			AutoLootEnabled: true,
			AutoLootRadius:  150.0,
		},
	}

	game := &engine.EbitenGame{}

	deserializePlayerState(player, state, game)

	// Verify position
	if pos, ok := player.GetComponent("position"); ok {
		p := pos.(*engine.PositionComponent)
		if p.X != 100.5 || p.Y != 200.5 {
			t.Errorf("Position = (%f, %f), want (100.5, 200.5)", p.X, p.Y)
		}
	} else {
		t.Error("Position component not found")
	}

	// Verify health
	if h, ok := player.GetComponent("health"); ok {
		health := h.(*engine.HealthComponent)
		if health.Current != 75 || health.Max != 100 {
			t.Errorf("Health = %f/%f, want 75/100", health.Current, health.Max)
		}
	} else {
		t.Error("Health component not found")
	}

	// Verify stats
	if s, ok := player.GetComponent("stats"); ok {
		stats := s.(*engine.StatsComponent)
		if stats.Attack != 20 || stats.Defense != 15 || stats.MagicPower != 25 {
			t.Errorf("Stats = (%f, %f, %f), want (20, 15, 25)", stats.Attack, stats.Defense, stats.MagicPower)
		}
	} else {
		t.Error("Stats component not found")
	}

	// Verify QoL state
	if q, ok := player.GetComponent("qol"); ok {
		qolComp := q.(*qol.QoLComponent)
		if !qolComp.AutoLootEnabled {
			t.Error("QoL AutoLootEnabled should be true")
		}
		if qolComp.AutoLootRadius != 150.0 {
			t.Errorf("QoL AutoLootRadius = %f, want 150.0", qolComp.AutoLootRadius)
		}
	} else {
		t.Error("QoL component not found")
	}
}

// TestSerializeDeserializePlayerStateRoundTrip tests that serialize/deserialize is a valid round-trip.
func TestSerializeDeserializePlayerStateRoundTrip(t *testing.T) {
	// Create player with full state
	original := engine.NewEntity(99)
	original.AddComponent(&engine.PositionComponent{X: 123.456, Y: 789.012})
	original.AddComponent(&engine.HealthComponent{Current: 80, Max: 120})
	original.AddComponent(&engine.StatsComponent{Attack: 30, Defense: 25, MagicPower: 40})
	original.AddComponent(&engine.ExperienceComponent{Level: 10, CurrentXP: 5000})
	original.AddComponent(&engine.InventoryComponent{Gold: 2500})
	original.AddComponent(&engine.ManaComponent{Current: 80, Max: 150})
	original.AddComponent(&engine.EquipmentComponent{
		Slots: make(map[engine.EquipmentSlot]*item.Item),
	})
	original.AddComponent(&qol.QoLComponent{
		AutoLootEnabled: true,
		AutoLootRadius:  200.0,
		MountWhistle:    true,
		RecipeTracking:  true,
	})

	game := &engine.EbitenGame{}

	// Serialize
	state := serializePlayerState(original, game)

	// Create new player and deserialize
	restored := engine.NewEntity(99)
	restored.AddComponent(&engine.PositionComponent{})
	restored.AddComponent(&engine.HealthComponent{})
	restored.AddComponent(&engine.StatsComponent{})
	restored.AddComponent(&engine.ExperienceComponent{})
	restored.AddComponent(&engine.InventoryComponent{})
	restored.AddComponent(&engine.ManaComponent{})
	restored.AddComponent(&engine.EquipmentComponent{
		Slots: make(map[engine.EquipmentSlot]*item.Item),
	})
	restored.AddComponent(&qol.QoLComponent{})

	deserializePlayerState(restored, state, game)

	// Verify round-trip for position
	origPosComp, _ := original.GetComponent("position")
	origPos := origPosComp.(*engine.PositionComponent)
	restoredPosComp, _ := restored.GetComponent("position")
	restoredPos := restoredPosComp.(*engine.PositionComponent)
	if origPos.X != restoredPos.X || origPos.Y != restoredPos.Y {
		t.Errorf("Position mismatch: original (%f, %f), restored (%f, %f)",
			origPos.X, origPos.Y, restoredPos.X, restoredPos.Y)
	}

	// Verify round-trip for health
	origHealthComp, _ := original.GetComponent("health")
	origHealth := origHealthComp.(*engine.HealthComponent)
	restoredHealthComp, _ := restored.GetComponent("health")
	restoredHealth := restoredHealthComp.(*engine.HealthComponent)
	if origHealth.Current != restoredHealth.Current || origHealth.Max != restoredHealth.Max {
		t.Errorf("Health mismatch: original %f/%f, restored %f/%f",
			origHealth.Current, origHealth.Max, restoredHealth.Current, restoredHealth.Max)
	}

	// Verify round-trip for QoL
	origQoLComp, _ := original.GetComponent("qol")
	origQoL := origQoLComp.(*qol.QoLComponent)
	restoredQoLComp, _ := restored.GetComponent("qol")
	restoredQoL := restoredQoLComp.(*qol.QoLComponent)
	if origQoL.AutoLootEnabled != restoredQoL.AutoLootEnabled {
		t.Errorf("QoL AutoLootEnabled mismatch: original %v, restored %v",
			origQoL.AutoLootEnabled, restoredQoL.AutoLootEnabled)
	}
	if origQoL.AutoLootRadius != restoredQoL.AutoLootRadius {
		t.Errorf("QoL AutoLootRadius mismatch: original %f, restored %f",
			origQoL.AutoLootRadius, restoredQoL.AutoLootRadius)
	}
}

// mockClientConnection implements network.ClientConnection for testing voice transport wiring.
type mockClientConnection struct {
	playerID  uint64
	connected bool
	lastInput string
	lastData  []byte
}

func (m *mockClientConnection) Connect() error               { return nil }
func (m *mockClientConnection) Disconnect() error             { return nil }
func (m *mockClientConnection) IsConnected() bool             { return m.connected }
func (m *mockClientConnection) GetPlayerID() uint64           { return m.playerID }
func (m *mockClientConnection) SetPlayerID(id uint64)         { m.playerID = id }
func (m *mockClientConnection) GetLatency() time.Duration     { return 0 }
func (m *mockClientConnection) ReceiveError() <-chan error     { return make(chan error) }

func (m *mockClientConnection) SendInput(inputType string, data []byte) error {
	m.lastInput = inputType
	m.lastData = data
	return nil
}

func (m *mockClientConnection) ReceiveStateUpdate() <-chan *network.StateUpdate {
	return make(chan *network.StateUpdate)
}

// TestInitializeVoiceTransportNilAudioManager verifies graceful handling.
func TestInitializeVoiceTransportNilAudioManager(t *testing.T) {
	sys := &systemsContainer{}
	logger := logrus.NewEntry(logrus.StandardLogger())
	// Should not panic with nil audio manager
	initializeVoiceTransport(sys, &mockClientConnection{connected: true}, logger)
}

// TestInitializeVoiceTransportNilNetworkClient verifies offline mode handling.
func TestInitializeVoiceTransportNilNetworkClient(t *testing.T) {
	sys := &systemsContainer{
		baseAudioManager: audio.NewManager(44100, 42),
	}
	logger := logrus.NewEntry(logrus.StandardLogger())
	// Should not panic with nil network client
	initializeVoiceTransport(sys, nil, logger)
}

// TestInitializeVoiceTransportSuccess verifies full wiring path.
func TestInitializeVoiceTransportSuccess(t *testing.T) {
	sys := &systemsContainer{
		baseAudioManager: audio.NewManager(44100, 42),
	}
	logger := logrus.NewEntry(logrus.StandardLogger())
	mock := &mockClientConnection{playerID: 12345, connected: true}

	initializeVoiceTransport(sys, mock, logger)

	// After initialization, the audio manager should report voice as enabled
	if !sys.baseAudioManager.IsVoiceEnabled() {
		t.Error("voice should be enabled after transport initialization")
	}
}

// TestInitializeVoiceTransportZeroPlayerID verifies that voice is skipped when
// the server hasn't assigned a player ID yet.
func TestInitializeVoiceTransportZeroPlayerID(t *testing.T) {
	sys := &systemsContainer{
		baseAudioManager: audio.NewManager(44100, 42),
	}
	logger := logrus.NewEntry(logrus.StandardLogger())
	mock := &mockClientConnection{playerID: 0, connected: true}

	initializeVoiceTransport(sys, mock, logger)

	// Voice should NOT be enabled when player ID is 0
	if sys.baseAudioManager.IsVoiceEnabled() {
		t.Error("voice should not be enabled when player ID is 0")
	}
}
