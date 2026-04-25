package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// TestCompleteCast_NoCooldownOnManaDrainedMidCast is the regression test for G39.
// A caster with sufficient mana starts a cast; mana is drained to below spell
// cost before the cast bar reaches 1.0. On completion, executeCast silently
// no-ops and completeCast must NOT apply the full cooldown.
func TestCompleteCast_NoCooldownOnManaDrainedMidCast(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(42))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

	caster := world.CreateEntity()
	caster.AddComponent(&PositionComponent{X: 100, Y: 100, Initialized: true})
	mana := &ManaComponent{Current: 60, Max: 100, Regen: 0}
	caster.AddComponent(mana)

	spell := &magic.Spell{
		Name:   "Fireball",
		Type:   magic.TypeOffensive,
		Target: magic.TargetSingle,
		Stats: magic.Stats{
			Damage:   30,
			ManaCost: 50,
			Cooldown: 8.0,
			CastTime: 1.0,
			Range:    200.0,
		},
	}
	slots := &SpellSlotComponent{Casting: -1}
	slots.SetSlot(0, spell)
	caster.AddComponent(slots)

	world.Update(0)
	entities := world.GetEntities()

	// Start the cast (60 >= 50 mana, so it is allowed).
	if !system.StartCast(caster, 0) {
		t.Fatal("StartCast should succeed when mana is sufficient")
	}

	// Advance cast bar to 50% without completing.
	system.Update(entities, 0.5)
	if !slots.IsCasting() {
		t.Fatal("entity should still be casting at 50% progress")
	}

	// Drain mana below the spell cost mid-cast.
	mana.Current = 30

	// Advance cast bar past 100% — completeCast fires.
	system.Update(entities, 0.6)

	// Casting state must be cleared.
	if slots.IsCasting() {
		t.Error("IsCasting should be false after cast bar reached 1.0")
	}
	if slots.CastingBar != 0 {
		t.Errorf("CastingBar = %f, want 0 after completion", slots.CastingBar)
	}

	// No cooldown must be applied because executeCast returned false.
	if slots.IsOnCooldown(0) {
		t.Errorf("Cooldown = %f, want 0: cooldown must not be set when spell silently no-ops due to drained mana", slots.Cooldowns[0])
	}

	// Mana must be unchanged (spell was not consumed).
	if mana.Current != 30 {
		t.Errorf("Mana = %d, want 30: mana must not be consumed when executeCast no-ops", mana.Current)
	}
}

// TestCompleteCast_CooldownAppliedOnSuccess verifies the happy path is intact
// after the G39 fix: a successful cast still sets the full cooldown.
func TestCompleteCast_CooldownAppliedOnSuccess(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(42))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

	caster := world.CreateEntity()
	caster.AddComponent(&PositionComponent{X: 100, Y: 100, Initialized: true})
	mana := &ManaComponent{Current: 100, Max: 100, Regen: 0}
	caster.AddComponent(mana)

	spell := &magic.Spell{
		Name:   "Fireball",
		Type:   magic.TypeOffensive,
		Target: magic.TargetSelf,
		Stats: magic.Stats{
			ManaCost: 20,
			Cooldown: 5.0,
			CastTime: 1.0,
		},
	}
	slots := &SpellSlotComponent{Casting: -1}
	slots.SetSlot(0, spell)
	caster.AddComponent(slots)

	world.Update(0)
	entities := world.GetEntities()

	if !system.StartCast(caster, 0) {
		t.Fatal("StartCast should succeed")
	}

	// Complete the cast in one step.
	system.Update(entities, 1.1)

	if slots.IsCasting() {
		t.Error("IsCasting should be false after completion")
	}
	if !slots.IsOnCooldown(0) {
		t.Errorf("Cooldown = %f, want %f: cooldown must be set after successful cast", slots.Cooldowns[0], spell.Stats.Cooldown)
	}
	if mana.Current != 80 {
		t.Errorf("Mana = %d, want 80 after 20-cost spell", mana.Current)
	}
}
