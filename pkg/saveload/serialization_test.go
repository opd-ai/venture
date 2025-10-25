package saveload

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// TestItemToData tests item serialization.
func TestItemToData(t *testing.T) {
	testItem := &item.Item{
		Name:        "Iron Sword",
		Type:        item.TypeWeapon,
		WeaponType:  item.WeaponSword,
		Rarity:      item.RarityCommon,
		Seed:        12345,
		Tags:        []string{"melee", "iron"},
		Description: "A sturdy iron sword",
		Stats: item.Stats{
			Damage:        25,
			AttackSpeed:   1.2,
			Value:         100,
			Weight:        5.0,
			RequiredLevel: 1,
			DurabilityMax: 100,
			Durability:    100,
		},
	}

	data := ItemToData(testItem)

	if data.Name != "Iron Sword" {
		t.Errorf("Name = %s, want 'Iron Sword'", data.Name)
	}
	if data.Type != "weapon" {
		t.Errorf("Type = %s, want 'weapon'", data.Type)
	}
	if data.WeaponType != "sword" {
		t.Errorf("WeaponType = %s, want 'sword'", data.WeaponType)
	}
	if data.Rarity != "common" {
		t.Errorf("Rarity = %s, want 'common'", data.Rarity)
	}
	if data.Damage != 25 {
		t.Errorf("Damage = %d, want 25", data.Damage)
	}
	if data.Value != 100 {
		t.Errorf("Value = %d, want 100", data.Value)
	}
}

// TestDataToItem tests item deserialization.
func TestDataToItem(t *testing.T) {
	data := ItemData{
		Name:        "Leather Armor",
		Type:        "armor",
		ArmorType:   "chest",
		Rarity:      "uncommon",
		Seed:        54321,
		Tags:        []string{"light", "leather"},
		Description: "Light leather chest armor",
		Defense:     15,
		Value:       150,
		Weight:      8.0,
	}

	itm := DataToItem(data)

	if itm.Name != "Leather Armor" {
		t.Errorf("Name = %s, want 'Leather Armor'", itm.Name)
	}
	if itm.Type != item.TypeArmor {
		t.Errorf("Type = %v, want TypeArmor", itm.Type)
	}
	if itm.ArmorType != item.ArmorChest {
		t.Errorf("ArmorType = %v, want ArmorChest", itm.ArmorType)
	}
	if itm.Rarity != item.RarityUncommon {
		t.Errorf("Rarity = %v, want RarityUncommon", itm.Rarity)
	}
	if itm.Stats.Defense != 15 {
		t.Errorf("Defense = %d, want 15", itm.Stats.Defense)
	}
}

// TestItemRoundTrip tests item serialization and deserialization.
func TestItemRoundTrip(t *testing.T) {
	original := &item.Item{
		Name:        "Epic Fire Staff",
		Type:        item.TypeWeapon,
		WeaponType:  item.WeaponStaff,
		Rarity:      item.RarityEpic,
		Seed:        99999,
		Tags:        []string{"magic", "fire", "staff"},
		Description: "A powerful staff imbued with fire magic",
		Stats: item.Stats{
			Damage:        50,
			AttackSpeed:   0.8,
			Value:         500,
			Weight:        3.0,
			RequiredLevel: 10,
			DurabilityMax: 150,
			Durability:    150,
		},
	}

	// Serialize
	data := ItemToData(original)

	// Deserialize
	restored := DataToItem(data)

	// Verify
	if restored.Name != original.Name {
		t.Errorf("Name = %s, want %s", restored.Name, original.Name)
	}
	if restored.Type != original.Type {
		t.Errorf("Type = %v, want %v", restored.Type, original.Type)
	}
	if restored.Rarity != original.Rarity {
		t.Errorf("Rarity = %v, want %v", restored.Rarity, original.Rarity)
	}
	if restored.Stats.Damage != original.Stats.Damage {
		t.Errorf("Damage = %d, want %d", restored.Stats.Damage, original.Stats.Damage)
	}
	if restored.Stats.Value != original.Stats.Value {
		t.Errorf("Value = %d, want %d", restored.Stats.Value, original.Stats.Value)
	}
}

// TestSpellToData tests spell serialization.
func TestSpellToData(t *testing.T) {
	testSpell := &magic.Spell{
		Name:        "Fireball",
		Type:        magic.TypeOffensive,
		Element:     magic.ElementFire,
		Target:      magic.TargetSingle,
		Rarity:      magic.RarityRare,
		Seed:        11111,
		Tags:        []string{"fire", "damage", "projectile"},
		Description: "A blazing ball of fire",
		Stats: magic.Stats{
			Damage:   40,
			ManaCost: 25,
			Cooldown: 5.0,
			CastTime: 1.5,
			Range:    30.0,
		},
	}

	data := SpellToData(testSpell)

	if data.Name != "Fireball" {
		t.Errorf("Name = %s, want 'Fireball'", data.Name)
	}
	if data.Type != "offensive" {
		t.Errorf("Type = %s, want 'offensive'", data.Type)
	}
	if data.Element != "fire" {
		t.Errorf("Element = %s, want 'fire'", data.Element)
	}
	if data.Damage != 40 {
		t.Errorf("Damage = %d, want 40", data.Damage)
	}
	if data.ManaCost != 25 {
		t.Errorf("ManaCost = %d, want 25", data.ManaCost)
	}
}

// TestDataToSpell tests spell deserialization.
func TestDataToSpell(t *testing.T) {
	data := SpellData{
		Name:        "Heal",
		Type:        "healing",
		Element:     "light",
		Target:      "self",
		Rarity:      "common",
		Seed:        22222,
		Tags:        []string{"healing", "holy"},
		Description: "Restores health",
		Healing:     50,
		ManaCost:    20,
		Cooldown:    8.0,
		CastTime:    2.0,
	}

	spell := DataToSpell(data)

	if spell.Name != "Heal" {
		t.Errorf("Name = %s, want 'Heal'", spell.Name)
	}
	if spell.Type != magic.TypeHealing {
		t.Errorf("Type = %v, want TypeHealing", spell.Type)
	}
	if spell.Element != magic.ElementLight {
		t.Errorf("Element = %v, want ElementLight", spell.Element)
	}
	if spell.Stats.Healing != 50 {
		t.Errorf("Healing = %d, want 50", spell.Stats.Healing)
	}
}

// TestSpellRoundTrip tests spell serialization and deserialization.
func TestSpellRoundTrip(t *testing.T) {
	original := &magic.Spell{
		Name:        "Lightning Strike",
		Type:        magic.TypeOffensive,
		Element:     magic.ElementLightning,
		Target:      magic.TargetArea,
		Rarity:      magic.RarityEpic,
		Seed:        33333,
		Tags:        []string{"lightning", "area", "damage"},
		Description: "Call down lightning from the sky",
		Stats: magic.Stats{
			Damage:   80,
			ManaCost: 50,
			Cooldown: 12.0,
			CastTime: 3.0,
			Range:    40.0,
			AreaSize: 10.0,
		},
	}

	// Serialize
	data := SpellToData(original)

	// Deserialize
	restored := DataToSpell(data)

	// Verify
	if restored.Name != original.Name {
		t.Errorf("Name = %s, want %s", restored.Name, original.Name)
	}
	if restored.Type != original.Type {
		t.Errorf("Type = %v, want %v", restored.Type, original.Type)
	}
	if restored.Element != original.Element {
		t.Errorf("Element = %v, want %v", restored.Element, original.Element)
	}
	if restored.Stats.Damage != original.Stats.Damage {
		t.Errorf("Damage = %d, want %d", restored.Stats.Damage, original.Stats.Damage)
	}
	if restored.Stats.AreaSize != original.Stats.AreaSize {
		t.Errorf("AreaSize = %f, want %f", restored.Stats.AreaSize, original.Stats.AreaSize)
	}
}

// TestItemToData_Nil tests nil item handling.
func TestItemToData_Nil(t *testing.T) {
	data := ItemToData(nil)

	if data.Name != "" {
		t.Errorf("Expected empty ItemData for nil item")
	}
}

// TestSpellToData_Nil tests nil spell handling.
func TestSpellToData_Nil(t *testing.T) {
	data := SpellToData(nil)

	if data.Name != "" {
		t.Errorf("Expected empty SpellData for nil spell")
	}
}

// TestParseItemType tests item type parsing.
func TestParseItemType(t *testing.T) {
	tests := []struct {
		input string
		want  item.ItemType
	}{
		{"weapon", item.TypeWeapon},
		{"armor", item.TypeArmor},
		{"consumable", item.TypeConsumable},
		{"accessory", item.TypeAccessory},
		{"invalid", item.TypeWeapon}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseItemType(tt.input)
			if got != tt.want {
				t.Errorf("parseItemType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseRarity tests rarity parsing.
func TestParseRarity(t *testing.T) {
	tests := []struct {
		input string
		want  item.Rarity
	}{
		{"common", item.RarityCommon},
		{"uncommon", item.RarityUncommon},
		{"rare", item.RarityRare},
		{"epic", item.RarityEpic},
		{"legendary", item.RarityLegendary},
		{"invalid", item.RarityCommon}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseItemRarity(tt.input)
			if got != tt.want {
				t.Errorf("parseItemRarity(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestConsumableItem tests consumable item serialization.
func TestConsumableItem(t *testing.T) {
	original := &item.Item{
		Name:           "Health Potion",
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumablePotion,
		Rarity:         item.RarityCommon,
		Seed:           44444,
		Description:    "Restores 50 health",
		Stats: item.Stats{
			Value:  25,
			Weight: 0.5,
		},
	}

	data := ItemToData(original)
	restored := DataToItem(data)

	if restored.Type != item.TypeConsumable {
		t.Errorf("Type = %v, want TypeConsumable", restored.Type)
	}
	if restored.ConsumableType != item.ConsumablePotion {
		t.Errorf("ConsumableType = %v, want ConsumablePotion", restored.ConsumableType)
	}
}

// BenchmarkItemToData benchmarks item serialization.
func BenchmarkItemToData(b *testing.B) {
	testItem := &item.Item{
		Name:   "Test Item",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Damage: 10, Value: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ItemToData(testItem)
	}
}

// BenchmarkDataToItem benchmarks item deserialization.
func BenchmarkDataToItem(b *testing.B) {
	data := ItemData{
		Name:   "Test Item",
		Type:   "weapon",
		Rarity: "common",
		Damage: 10,
		Value:  50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DataToItem(data)
	}
}

// TestAnimationStateToData tests animation state serialization.
func TestAnimationStateToData(t *testing.T) {
	tests := []struct {
		name           string
		state          string
		frameIndex     uint8
		loop           bool
		lastUpdateTime float64
	}{
		{"idle state", "idle", 0, true, 0.0},
		{"walk state", "walk", 3, true, 1.5},
		{"attack state", "attack", 5, false, 2.75},
		{"death state", "death", 7, false, 10.0},
		{"run state", "run", 2, true, 0.333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := AnimationStateToData(tt.state, tt.frameIndex, tt.loop, tt.lastUpdateTime)

			if data == nil {
				t.Fatal("AnimationStateToData returned nil")
			}

			if data.State != tt.state {
				t.Errorf("State = %s, want %s", data.State, tt.state)
			}
			if data.FrameIndex != tt.frameIndex {
				t.Errorf("FrameIndex = %d, want %d", data.FrameIndex, tt.frameIndex)
			}
			if data.Loop != tt.loop {
				t.Errorf("Loop = %v, want %v", data.Loop, tt.loop)
			}
			if data.LastUpdateTime != tt.lastUpdateTime {
				t.Errorf("LastUpdateTime = %f, want %f", data.LastUpdateTime, tt.lastUpdateTime)
			}
		})
	}
}

// TestDataToAnimationState tests animation state deserialization.
func TestDataToAnimationState(t *testing.T) {
	tests := []struct {
		name           string
		data           *AnimationStateData
		wantState      string
		wantFrame      uint8
		wantLoop       bool
		wantUpdateTime float64
	}{
		{
			name: "valid data",
			data: &AnimationStateData{
				State:          "walk",
				FrameIndex:     3,
				Loop:           true,
				LastUpdateTime: 1.5,
			},
			wantState:      "walk",
			wantFrame:      3,
			wantLoop:       true,
			wantUpdateTime: 1.5,
		},
		{
			name: "attack data",
			data: &AnimationStateData{
				State:          "attack",
				FrameIndex:     5,
				Loop:           false,
				LastUpdateTime: 2.75,
			},
			wantState:      "attack",
			wantFrame:      5,
			wantLoop:       false,
			wantUpdateTime: 2.75,
		},
		{
			name:           "nil data defaults",
			data:           nil,
			wantState:      "idle",
			wantFrame:      0,
			wantLoop:       true,
			wantUpdateTime: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, frame, loop, updateTime := DataToAnimationState(tt.data)

			if state != tt.wantState {
				t.Errorf("state = %s, want %s", state, tt.wantState)
			}
			if frame != tt.wantFrame {
				t.Errorf("frame = %d, want %d", frame, tt.wantFrame)
			}
			if loop != tt.wantLoop {
				t.Errorf("loop = %v, want %v", loop, tt.wantLoop)
			}
			if updateTime != tt.wantUpdateTime {
				t.Errorf("updateTime = %f, want %f", updateTime, tt.wantUpdateTime)
			}
		})
	}
}

// TestAnimationStateRoundTrip tests serialization and deserialization round-trip.
func TestAnimationStateRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		frameIndex uint8
		loop       bool
		updateTime float64
	}{
		{"idle", "idle", 0, true, 0.0},
		{"walk", "walk", 3, true, 1.5},
		{"run", "run", 2, true, 0.75},
		{"attack", "attack", 5, false, 2.0},
		{"cast", "cast", 4, false, 1.8},
		{"hit", "hit", 1, false, 0.2},
		{"death", "death", 7, false, 3.5},
		{"jump", "jump", 2, false, 0.5},
		{"crouch", "crouch", 0, true, 0.0},
		{"use", "use", 1, false, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data := AnimationStateToData(tt.state, tt.frameIndex, tt.loop, tt.updateTime)

			// Deserialize
			gotState, gotFrame, gotLoop, gotTime := DataToAnimationState(data)

			// Verify round-trip
			if gotState != tt.state {
				t.Errorf("Round-trip state = %s, want %s", gotState, tt.state)
			}
			if gotFrame != tt.frameIndex {
				t.Errorf("Round-trip frame = %d, want %d", gotFrame, tt.frameIndex)
			}
			if gotLoop != tt.loop {
				t.Errorf("Round-trip loop = %v, want %v", gotLoop, tt.loop)
			}
			if gotTime != tt.updateTime {
				t.Errorf("Round-trip time = %f, want %f", gotTime, tt.updateTime)
			}
		})
	}
}

// BenchmarkAnimationStateToData benchmarks serialization.
func BenchmarkAnimationStateToData(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AnimationStateToData("walk", 3, true, 1.5)
	}
}

// BenchmarkDataToAnimationState benchmarks deserialization.
func BenchmarkDataToAnimationState(b *testing.B) {
	data := &AnimationStateData{
		State:          "walk",
		FrameIndex:     3,
		Loop:           true,
		LastUpdateTime: 1.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = DataToAnimationState(data)
	}
}
