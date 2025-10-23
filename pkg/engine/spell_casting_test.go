//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// TestManaComponent tests the mana component.
func TestManaComponent(t *testing.T) {
	mana := &ManaComponent{
		Current: 50,
		Max:     100,
		Regen:   5.0,
	}

	if mana.Type() != "mana" {
		t.Errorf("Type() = %s, want 'mana'", mana.Type())
	}
}

// TestSpellSlotComponent tests spell slot management.
func TestSpellSlotComponent(t *testing.T) {
	slots := &SpellSlotComponent{
		Casting: -1,
	}

	if slots.Type() != "spell_slots" {
		t.Errorf("Type() = %s, want 'spell_slots'", slots.Type())
	}

	// Test slot operations
	testSpell := &magic.Spell{
		Name: "Fireball",
		Type: magic.TypeOffensive,
	}

	slots.SetSlot(0, testSpell)
	if slots.GetSlot(0) != testSpell {
		t.Error("Slot 0 does not contain set spell")
	}

	if slots.GetSlot(5) != nil {
		t.Error("Invalid slot should return nil")
	}

	// Test cooldown
	if slots.IsOnCooldown(0) {
		t.Error("Slot should not be on cooldown initially")
	}

	slots.Cooldowns[0] = 5.0
	if !slots.IsOnCooldown(0) {
		t.Error("Slot should be on cooldown")
	}

	// Test casting state
	if slots.IsCasting() {
		t.Error("Should not be casting initially")
	}

	slots.Casting = 0
	if !slots.IsCasting() {
		t.Error("Should be casting")
	}
}

// TestSpellCastingSystem tests spell casting mechanics.
func TestSpellCastingSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	// Create caster entity
	caster := world.CreateEntity()
	caster.AddComponent(&PositionComponent{X: 100, Y: 100})
	caster.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

	slots := &SpellSlotComponent{Casting: -1}
	testSpell := &magic.Spell{
		Name:   "Fireball",
		Type:   magic.TypeOffensive,
		Target: magic.TargetSingle, // Target single enemy
		Stats: magic.Stats{
			Damage:   30,
			ManaCost: 20,
			Cooldown: 5.0,
			CastTime: 1.0,
			Range:    50.0,
		},
	}
	slots.SetSlot(0, testSpell)
	caster.AddComponent(slots)

	// Create target enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 120, Y: 100})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)
	entities := world.GetEntities()

	// Start casting
	if !system.StartCast(caster, 0) {
		t.Fatal("Failed to start cast")
	}

	if !slots.IsCasting() {
		t.Error("Should be casting after StartCast")
	}

	// Simulate casting time
	system.Update(entities, 0.5) // 50% progress
	if slots.CastingBar < 0.4 || slots.CastingBar > 0.6 {
		t.Errorf("CastingBar = %f, want ~0.5", slots.CastingBar)
	}

	// Complete cast
	system.Update(entities, 0.5) // 100% progress

	// Verify spell completed
	if slots.IsCasting() {
		t.Error("Should not be casting after completion")
	}

	// Verify mana cost
	manaComp, _ := caster.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Current != 80 {
		t.Errorf("Mana = %d, want 80 (100 - 20 cost)", mana.Current)
	}

	// Verify cooldown started
	if !slots.IsOnCooldown(0) {
		t.Error("Slot should be on cooldown after cast")
	}

	// Verify enemy took damage
	healthComp, _ := enemy.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 70 {
		t.Errorf("Enemy health = %f, want 70 (100 - 30 damage)", health.Current)
	}
}

// TestSpellCastingSystem_InsufficientMana tests casting without enough mana.
func TestSpellCastingSystem_InsufficientMana(t *testing.T) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	caster := world.CreateEntity()
	caster.AddComponent(&PositionComponent{X: 100, Y: 100})
	caster.AddComponent(&ManaComponent{Current: 10, Max: 100, Regen: 5.0}) // Low mana

	slots := &SpellSlotComponent{Casting: -1}
	expensiveSpell := &magic.Spell{
		Name: "Meteor",
		Type: magic.TypeOffensive,
		Stats: magic.Stats{
			Damage:   100,
			ManaCost: 50, // More than available
			Cooldown: 10.0,
			CastTime: 2.0,
		},
	}
	slots.SetSlot(0, expensiveSpell)
	caster.AddComponent(slots)

	// Try to start cast
	if system.StartCast(caster, 0) {
		t.Error("Should not be able to start cast with insufficient mana")
	}
}

// TestSpellCastingSystem_Cooldown tests cooldown mechanics.
func TestSpellCastingSystem_Cooldown(t *testing.T) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	caster := world.CreateEntity()
	caster.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

	slots := &SpellSlotComponent{Casting: -1}
	slots.Cooldowns[0] = 5.0 // On cooldown
	slots.SetSlot(0, &magic.Spell{
		Name:  "Test Spell",
		Stats: magic.Stats{ManaCost: 10},
	})
	caster.AddComponent(slots)

	// Try to cast while on cooldown
	if system.StartCast(caster, 0) {
		t.Error("Should not be able to cast while on cooldown")
	}

	world.Update(0)
	entities := world.GetEntities()

	// Update to reduce cooldown
	system.Update(entities, 3.0)
	if slots.Cooldowns[0] != 2.0 {
		t.Errorf("Cooldown = %f, want 2.0 (5.0 - 3.0)", slots.Cooldowns[0])
	}

	// Continue until cooldown expires
	system.Update(entities, 2.0)
	if slots.Cooldowns[0] != 0 {
		t.Errorf("Cooldown = %f, want 0", slots.Cooldowns[0])
	}

	// Should be able to cast now
	if system.StartCast(caster, 0) == false {
		t.Error("Should be able to cast after cooldown expires")
	}
}

// TestCancelCast tests spell cast cancellation.
func TestCancelCast(t *testing.T) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	caster := world.CreateEntity()
	caster.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

	slots := &SpellSlotComponent{Casting: -1}
	slots.SetSlot(0, &magic.Spell{
		Name: "Test Spell",
		Stats: magic.Stats{
			ManaCost: 20,
			CastTime: 2.0,
		},
	})
	caster.AddComponent(slots)

	// Start casting
	system.StartCast(caster, 0)
	if !slots.IsCasting() {
		t.Fatal("Should be casting")
	}

	// Cancel
	system.CancelCast(caster)
	if slots.IsCasting() {
		t.Error("Should not be casting after cancel")
	}

	if slots.CastingBar != 0 {
		t.Error("CastingBar should be reset to 0")
	}
}

// TestManaRegenSystem tests mana regeneration.
func TestManaRegenSystem(t *testing.T) {
	world := NewWorld()
	system := &ManaRegenSystem{}

	entity := world.CreateEntity()
	mana := &ManaComponent{
		Current: 50,
		Max:     100,
		Regen:   10.0, // 10 mana per second
	}
	entity.AddComponent(mana)

	world.Update(0)
	entities := world.GetEntities()

	// Regenerate for 2 seconds
	system.Update(entities, 2.0)

	if mana.Current != 70 {
		t.Errorf("Mana = %d, want 70 (50 + 10*2)", mana.Current)
	}

	// Regenerate past max
	system.Update(entities, 5.0)
	if mana.Current != 100 {
		t.Errorf("Mana = %d, want 100 (capped at max)", mana.Current)
	}
}

// TestLoadPlayerSpells tests spell loading for player.
func TestLoadPlayerSpells(t *testing.T) {
	player := NewWorld().CreateEntity()
	player.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

	seed := int64(12345)
	genreID := "fantasy"
	depth := 5

	err := LoadPlayerSpells(player, seed, genreID, depth)
	if err != nil {
		t.Fatalf("LoadPlayerSpells failed: %v", err)
	}

	// Verify spell slots were added
	if !player.HasComponent("spell_slots") {
		t.Fatal("Player missing spell_slots component")
	}

	slotsComp, _ := player.GetComponent("spell_slots")
	slots := slotsComp.(*SpellSlotComponent)

	// Verify all 5 slots have spells
	for i := 0; i < 5; i++ {
		spell := slots.GetSlot(i)
		if spell == nil {
			t.Errorf("Slot %d is empty, want spell", i)
		} else {
			t.Logf("Slot %d: %s (%s, %s element)", i, spell.Name, spell.Type, spell.Element)
		}
	}
}

// TestLoadPlayerSpells_Determinism tests deterministic spell generation.
func TestLoadPlayerSpells_Determinism(t *testing.T) {
	seed := int64(99999)
	genreID := "scifi"

	// Generate spells twice with same seed
	player1 := NewWorld().CreateEntity()
	LoadPlayerSpells(player1, seed, genreID, 10)

	player2 := NewWorld().CreateEntity()
	LoadPlayerSpells(player2, seed, genreID, 10)

	slotsComp1, _ := player1.GetComponent("spell_slots")
	slotsComp2, _ := player2.GetComponent("spell_slots")
	slots1 := slotsComp1.(*SpellSlotComponent)
	slots2 := slotsComp2.(*SpellSlotComponent)

	// Verify spells are identical
	for i := 0; i < 5; i++ {
		spell1 := slots1.GetSlot(i)
		spell2 := slots2.GetSlot(i)

		if spell1.Name != spell2.Name {
			t.Errorf("Slot %d: spell1.Name = %s, spell2.Name = %s (expected identical)", i, spell1.Name, spell2.Name)
		}
		if spell1.Stats.Damage != spell2.Stats.Damage {
			t.Errorf("Slot %d: different damage (expected identical)", i)
		}
	}
}

// TestHealingSpell tests healing spell execution.
func TestHealingSpell(t *testing.T) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	caster := world.CreateEntity()
	caster.AddComponent(&PositionComponent{X: 100, Y: 100})
	caster.AddComponent(&HealthComponent{Current: 50, Max: 100})
	caster.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

	slots := &SpellSlotComponent{Casting: -1}
	healSpell := &magic.Spell{
		Name: "Heal",
		Type: magic.TypeHealing,
		Stats: magic.Stats{
			Healing:  40,
			ManaCost: 25,
			CastTime: 0.5,
		},
	}
	slots.SetSlot(0, healSpell)
	caster.AddComponent(slots)

	world.Update(0)
	entities := world.GetEntities()

	// Cast heal
	system.StartCast(caster, 0)
	system.Update(entities, 0.5) // Complete cast

	// Verify healing
	healthComp, _ := caster.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 90 {
		t.Errorf("Health = %f, want 90 (50 + 40 healing)", health.Current)
	}
}

// TestInstantCast tests instant cast spells (0 cast time).
func TestInstantCast(t *testing.T) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	caster := world.CreateEntity()
	caster.AddComponent(&PositionComponent{X: 100, Y: 100})
	caster.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

	slots := &SpellSlotComponent{Casting: -1}
	instantSpell := &magic.Spell{
		Name:   "Quick Shot",
		Type:   magic.TypeOffensive,
		Target: magic.TargetSingle, // Target single enemy
		Stats: magic.Stats{
			Damage:   20,
			ManaCost: 10,
			CastTime: 0, // Instant
			Range:    50.0,
		},
	}
	slots.SetSlot(0, instantSpell)
	caster.AddComponent(slots)

	// Create target
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 110, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)
	entities := world.GetEntities()

	// Cast instant spell
	system.StartCast(caster, 0)
	system.Update(entities, 0.016) // Single frame

	// Should complete immediately
	if slots.IsCasting() {
		t.Error("Instant cast should complete immediately")
	}

	// Verify damage was applied
	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 80 {
		t.Errorf("Target health = %f, want 80", health.Current)
	}
}

// BenchmarkSpellCastingSystem benchmarks the spell casting system.
func BenchmarkSpellCastingSystem(b *testing.B) {
	world := NewWorld()
	system := NewSpellCastingSystem(world)

	// Create 10 casters with spells
	for i := 0; i < 10; i++ {
		caster := world.CreateEntity()
		caster.AddComponent(&PositionComponent{X: float64(i * 50), Y: 100})
		caster.AddComponent(&ManaComponent{Current: 100, Max: 100, Regen: 5.0})

		slots := &SpellSlotComponent{Casting: -1}
		spell := &magic.Spell{
			Name: "Test Spell",
			Type: magic.TypeOffensive,
			Stats: magic.Stats{
				Damage:   30,
				ManaCost: 20,
				Cooldown: 5.0,
				CastTime: 1.0,
			},
		}
		slots.SetSlot(0, spell)
		caster.AddComponent(slots)
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
