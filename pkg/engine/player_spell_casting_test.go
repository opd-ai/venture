package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// TestPlayerSpellCastingSystem_NewPlayerSpellCastingSystem tests constructor
func TestPlayerSpellCastingSystem_NewPlayerSpellCastingSystem(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)

	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	if playerCastingSystem == nil {
		t.Fatal("NewPlayerSpellCastingSystem returned nil")
	}

	if playerCastingSystem.castingSystem != castingSystem {
		t.Error("Casting system not properly set")
	}

	if playerCastingSystem.world != world {
		t.Error("World not properly set")
	}

	// Check key bindings are initialized
	if playerCastingSystem.KeySpell1 == 0 {
		t.Error("KeySpell1 not initialized")
	}
	if playerCastingSystem.KeySpell2 == 0 {
		t.Error("KeySpell2 not initialized")
	}
}

// TestPlayerSpellCastingSystem_Update_NoPlayer tests system with no player entity
func TestPlayerSpellCastingSystem_Update_NoPlayer(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)
	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	// Create entities without input component (not players)
	entity1 := NewEntity(1)
	entity1.AddComponent(&PositionComponent{X: 0, Y: 0})
	entities := []*Entity{entity1}

	// Should not panic
	playerCastingSystem.Update(entities, 0.016)
}

// TestPlayerSpellCastingSystem_Update_NoSpellSlots tests player without spell slots
func TestPlayerSpellCastingSystem_Update_NoSpellSlots(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)
	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	// Create player without spell slots
	player := NewEntity(1)
	player.AddComponent(&EbitenInput{}) // Input component makes it a player
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{player}
	world.Update(0) // Process pending additions

	// Should not panic
	playerCastingSystem.Update(entities, 0.016)
}

// TestPlayerSpellCastingSystem_Update_SpellCasting tests actual spell casting
func TestPlayerSpellCastingSystem_Update_SpellCasting(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)
	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	// Create player with spell slots and mana
	player := NewEntity(1)
	input := &EbitenInput{}
	player.AddComponent(input)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&ManaComponent{Current: 100, Max: 100})

	// Create spell slots with a test spell
	slots := &SpellSlotComponent{Casting: -1} // -1 means not casting
	testSpell := &magic.Spell{
		Name:    "Test Fireball",
		Type:    magic.TypeOffensive,
		Element: magic.ElementFire,
		Stats: magic.Stats{
			ManaCost: 20,
			CastTime: 0.5,
			Cooldown: 2.0,
			Range:    200,
			Damage:   50,
		},
		Description: "Test spell",
	}
	slots.SetSlot(0, testSpell) // Spell in slot 1
	player.AddComponent(slots)

	entities := []*Entity{player}
	world.Update(0) // Process pending additions

	// Simulate pressing key 1
	input.Spell1Pressed = true

	// Update system
	playerCastingSystem.Update(entities, 0.016)

	// Check if spell is now casting
	if !slots.IsCasting() {
		t.Error("Spell should be casting after key press")
	}

	if slots.Casting != 0 {
		t.Errorf("Casting should be slot 0, got %d", slots.Casting)
	}
}

// TestPlayerSpellCastingSystem_Update_MultipleSlotsInput tests different spell slot keys
func TestPlayerSpellCastingSystem_Update_MultipleSlotsInput(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)
	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	tests := []struct {
		name         string
		slotIndex    int // 0-4
		expectedSlot int
		setInput     func(*EbitenInput)
	}{
		{"Spell Slot 1", 0, 0, func(i *EbitenInput) { i.Spell1Pressed = true }},
		{"Spell Slot 2", 1, 1, func(i *EbitenInput) { i.Spell2Pressed = true }},
		{"Spell Slot 3", 2, 2, func(i *EbitenInput) { i.Spell3Pressed = true }},
		{"Spell Slot 4", 3, 3, func(i *EbitenInput) { i.Spell4Pressed = true }},
		{"Spell Slot 5", 4, 4, func(i *EbitenInput) { i.Spell5Pressed = true }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh player for each test
			player := NewEntity(uint64(tt.slotIndex + 10))
			input := &EbitenInput{}
			player.AddComponent(input)
			player.AddComponent(&PositionComponent{X: 100, Y: 100})
			player.AddComponent(&ManaComponent{Current: 100, Max: 100})

			slots := &SpellSlotComponent{Casting: -1}
			testSpell := &magic.Spell{
				Name:  "Test Spell",
				Type:  magic.TypeOffensive,
				Stats: magic.Stats{ManaCost: 20, CastTime: 0.5},
			}
			slots.SetSlot(tt.slotIndex, testSpell)
			player.AddComponent(slots)

			entities := []*Entity{player}
			world.Update(0)

			// Press the specific spell key
			tt.setInput(input)

			// Update system
			playerCastingSystem.Update(entities, 0.016)

			// Verify correct slot activated
			if slots.Casting != tt.expectedSlot {
				t.Errorf("Expected casting slot %d, got %d", tt.expectedSlot, slots.Casting)
			}
		})
	}
}

// TestPlayerSpellCastingSystem_Update_AlreadyCasting tests that new casts don't interrupt
func TestPlayerSpellCastingSystem_Update_AlreadyCasting(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)
	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	// Create player
	player := NewEntity(1)
	input := &EbitenInput{}
	player.AddComponent(input)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&ManaComponent{Current: 100, Max: 100})

	slots := &SpellSlotComponent{Casting: -1}
	spell1 := &magic.Spell{
		Name:  "Spell 1",
		Type:  magic.TypeOffensive,
		Stats: magic.Stats{ManaCost: 20, CastTime: 1.0}, // Long cast time
	}
	spell2 := &magic.Spell{
		Name:  "Spell 2",
		Type:  magic.TypeOffensive,
		Stats: magic.Stats{ManaCost: 20, CastTime: 0.5},
	}
	slots.SetSlot(0, spell1)
	slots.SetSlot(1, spell2)
	player.AddComponent(slots)

	entities := []*Entity{player}
	world.Update(0)

	// Start casting spell 1
	input.Spell1Pressed = true
	playerCastingSystem.Update(entities, 0.016)

	if slots.Casting != 0 {
		t.Fatalf("First spell should be casting, got %d", slots.Casting)
	}

	initialCastingBar := slots.CastingBar

	// Try to cast spell 2 while still casting spell 1
	input.Spell1Pressed = false
	input.Spell2Pressed = true
	playerCastingSystem.Update(entities, 0.016)

	// Should still be casting spell 1
	if slots.Casting != 0 {
		t.Error("Should still be casting first spell")
	}

	if slots.CastingBar != initialCastingBar {
		t.Error("Casting bar should not have changed")
	}
}

// TestPlayerSpellCastingSystem_Update_EmptySlot tests casting from empty slot
func TestPlayerSpellCastingSystem_Update_EmptySlot(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusEffectSys := NewStatusEffectSystem(world, rng)
	castingSystem := NewSpellCastingSystem(world, statusEffectSys)
	playerCastingSystem := NewPlayerSpellCastingSystem(castingSystem, world)

	// Create player
	player := NewEntity(1)
	input := &EbitenInput{}
	player.AddComponent(input)
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&ManaComponent{Current: 100, Max: 100})

	slots := &SpellSlotComponent{Casting: -1}
	// Don't add any spells - all slots empty
	player.AddComponent(slots)

	entities := []*Entity{player}
	world.Update(0)

	// Try to cast from empty slot 1
	input.Spell1Pressed = true
	playerCastingSystem.Update(entities, 0.016)

	// Should not be casting anything
	if slots.IsCasting() {
		t.Error("Should not be casting from empty slot")
	}
}
