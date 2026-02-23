//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/class/advanced"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/qol"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/saveload"
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
