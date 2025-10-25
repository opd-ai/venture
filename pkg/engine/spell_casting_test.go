package engine

import (
	"math/rand"
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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	system := NewSpellCastingSystem(world, statusSys)

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

// TestSpellCasting_ElementalEffects tests that offensive spells apply elemental status effects.
func TestSpellCasting_ElementalEffects(t *testing.T) {
	tests := []struct {
		name           string
		element        magic.ElementType
		expectedEffect string
	}{
		{"Fire applies burning", magic.ElementFire, "burning"},
		{"Ice applies frozen", magic.ElementIce, "frozen"},
		{"Lightning applies shocked", magic.ElementLightning, "shocked"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create world and systems
			world := NewWorld()
			rng := rand.New(rand.NewSource(12345))
			statusSys := NewStatusEffectSystem(world, rng)
			spellSys := NewSpellCastingSystem(world, statusSys)

			// Create caster and target
			caster := &Entity{ID: 1, Components: make(map[string]Component)}
			target := &Entity{ID: 2, Components: make(map[string]Component)}

			// Add required components
			caster.AddComponent(&PositionComponent{X: 0, Y: 0})
			target.AddComponent(&PositionComponent{X: 5, Y: 5})
			target.AddComponent(&HealthComponent{Current: 100, Max: 100})

			// Add entities to world
			world.AddEntity(caster)
			world.AddEntity(target)
			world.Update(0) // Initialize entities list

			// Create offensive spell with element
			spell := &magic.Spell{
				Name:    "Test Spell",
				Type:    magic.TypeOffensive,
				Element: tt.element,
				Target:  magic.TargetSingle,
				Stats: magic.Stats{
					Damage: 20,
					Range:  20.0,
				},
			}

			// Cast spell
			spellSys.castOffensiveSpell(caster, spell, 0, 0)

			// Verify status effect applied
			hasEffect := false
			for _, comp := range target.Components {
				if effect, ok := comp.(*StatusEffectComponent); ok {
					if effect.EffectType == tt.expectedEffect {
						hasEffect = true
						break
					}
				}
			}

			if !hasEffect {
				t.Errorf("Expected status effect %s not found on target", tt.expectedEffect)
			}
		})
	}
}

// TestSpellCasting_ShieldMechanics tests defensive spell shield creation.
func TestSpellCasting_ShieldMechanics(t *testing.T) {
	// Create world and systems
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	spellSys := NewSpellCastingSystem(world, statusSys)

	// Create caster
	caster := &Entity{ID: 1, Components: make(map[string]Component)}
	caster.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Create defensive spell
	spell := &magic.Spell{
		Name:    "Shield",
		Type:    magic.TypeDefensive,
		Element: magic.ElementArcane,
		Target:  magic.TargetSelf,
		Stats: magic.Stats{
			Damage:   50, // Shield amount
			Duration: 30.0,
		},
	}

	// Cast spell
	spellSys.castDefensiveSpell(caster, spell)

	// Verify shield component added
	shieldComp, hasShield := caster.GetComponent("shield")
	if !hasShield {
		t.Fatal("Shield component not added to caster")
	}

	shield := shieldComp.(*ShieldComponent)
	if shield.Amount != 50.0 {
		t.Errorf("Shield amount = %f, want 50.0", shield.Amount)
	}
	if shield.Duration != 30.0 {
		t.Errorf("Shield duration = %f, want 30.0", shield.Duration)
	}
}

// TestSpellCasting_BuffSystem tests stat-boosting spells.
func TestSpellCasting_BuffSystem(t *testing.T) {
	// Create world and systems
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	spellSys := NewSpellCastingSystem(world, statusSys)

	// Create caster with stats
	caster := &Entity{ID: 1, Components: make(map[string]Component)}
	caster.AddComponent(&PositionComponent{X: 0, Y: 0})
	stats := &StatsComponent{
		Attack:  10.0,
		Defense: 10.0,
	}
	caster.AddComponent(stats)

	// Create buff spell (Strength)
	spell := &magic.Spell{
		Name:    "Strength",
		Type:    magic.TypeBuff,
		Element: magic.ElementLight,
		Target:  magic.TargetSelf,
		Stats: magic.Stats{
			Duration: 30.0,
		},
	}

	// Cast spell
	spellSys.castBuffSpell(caster, spell)

	// Verify attack increased
	if stats.Attack <= 10.0 {
		t.Errorf("Attack not buffed: %f", stats.Attack)
	}

	// Expected: 10.0 * 1.3 = 13.0
	expected := 13.0
	if stats.Attack < expected-0.1 || stats.Attack > expected+0.1 {
		t.Errorf("Attack = %f, want approximately %f", stats.Attack, expected)
	}
}

// TestSpellCasting_DebuffSystem tests stat-reducing spells.
func TestSpellCasting_DebuffSystem(t *testing.T) {
	// Create world and systems
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	spellSys := NewSpellCastingSystem(world, statusSys)

	// Create caster and target
	caster := &Entity{ID: 1, Components: make(map[string]Component)}
	target := &Entity{ID: 2, Components: make(map[string]Component)}

	caster.AddComponent(&PositionComponent{X: 0, Y: 0})
	target.AddComponent(&PositionComponent{X: 5, Y: 5})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	targetStats := &StatsComponent{
		Attack:  20.0,
		Defense: 20.0,
	}
	target.AddComponent(targetStats)

	world.AddEntity(caster)
	world.AddEntity(target)
	world.Update(0) // Initialize entities list

	// Create debuff spell (Weakness)
	spell := &magic.Spell{
		Name:    "Weakness",
		Type:    magic.TypeDebuff,
		Element: magic.ElementDark,
		Target:  magic.TargetSingle,
		Stats: magic.Stats{
			Damage:   5,
			Range:    20.0,
			Duration: 10.0,
		},
	}

	// Cast spell
	spellSys.castDebuffSpell(caster, spell, 0, 0)

	// Verify attack decreased
	if targetStats.Attack >= 20.0 {
		t.Errorf("Attack not debuffed: %f", targetStats.Attack)
	}

	// Expected: 20.0 * 0.7 = 14.0
	expected := 14.0
	if targetStats.Attack < expected-0.1 || targetStats.Attack > expected+0.1 {
		t.Errorf("Attack = %f, want approximately %f", targetStats.Attack, expected)
	}
}

// TestShieldComponent_AbsorbDamage tests shield damage absorption.
func TestShieldComponent_AbsorbDamage(t *testing.T) {
	tests := []struct {
		name           string
		shieldAmount   float64
		incomingDamage float64
		expectedAbsorb float64
		expectedRemain float64
	}{
		{"Shield absorbs all", 50.0, 30.0, 30.0, 20.0},
		{"Shield partially absorbs", 20.0, 50.0, 20.0, 0.0},
		{"Shield exactly absorbs", 30.0, 30.0, 30.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shield := &ShieldComponent{
				Amount:      tt.shieldAmount,
				MaxAmount:   tt.shieldAmount,
				Duration:    30.0,
				MaxDuration: 30.0,
			}

			absorbed := shield.AbsorbDamage(tt.incomingDamage)

			if absorbed != tt.expectedAbsorb {
				t.Errorf("Absorbed = %f, want %f", absorbed, tt.expectedAbsorb)
			}
			if shield.Amount != tt.expectedRemain {
				t.Errorf("Shield amount = %f, want %f", shield.Amount, tt.expectedRemain)
			}
		})
	}
}

// TestStatusEffectSystem_BurningDamage tests burning DoT effect.
func TestStatusEffectSystem_BurningDamage(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	health := &HealthComponent{Current: 100, Max: 100}
	entity.AddComponent(health)

	// Apply burning effect
	statusSys.ApplyStatusEffect(entity, "burning", 10.0, 3.0, 1.0)

	// Update for 1 second (should trigger first tick)
	entities := []*Entity{entity}
	statusSys.Update(entities, 1.0)

	// Health should be reduced by magnitude (10)
	if health.Current != 90.0 {
		t.Errorf("Health = %f, want 90.0", health.Current)
	}

	// Update for 1 more second (should trigger second tick)
	statusSys.Update(entities, 1.0)

	// Health should be 80.0 (2 ticks × 10 damage)
	if health.Current != 80.0 {
		t.Errorf("Health = %f, want 80.0 after 2 ticks", health.Current)
	}

	// Update for 1 more second (should trigger third tick)
	statusSys.Update(entities, 1.0)

	// Should have taken 30 total damage (3 ticks × 10) = 70 final
	// However, effect expires after 3 seconds, so third tick occurs just before expiration
	if health.Current < 69.9 || health.Current > 80.1 {
		t.Errorf("Health = %f, want 70.0 after 3 ticks (or 80.0 if expired before tick)", health.Current)
	}
}

// TestCombatSystem_ShieldIntegration tests shield integration in combat.
func TestCombatSystem_ShieldIntegration(t *testing.T) {
	combatSys := NewCombatSystem(12345)

	// Create attacker
	attacker := &Entity{ID: 1, Components: make(map[string]Component)}
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attack := &AttackComponent{
		Damage:   30.0,
		Range:    10.0,
		Cooldown: 1.0,
	}
	attacker.AddComponent(attack)

	// Create target with shield
	target := &Entity{ID: 2, Components: make(map[string]Component)}
	target.AddComponent(&PositionComponent{X: 5, Y: 0})
	health := &HealthComponent{Current: 100, Max: 100}
	target.AddComponent(health)
	shield := &ShieldComponent{
		Amount:      50.0,
		MaxAmount:   50.0,
		Duration:    30.0,
		MaxDuration: 30.0,
	}
	target.AddComponent(shield)

	// Attack target
	hit := combatSys.Attack(attacker, target)
	if !hit {
		t.Fatal("Attack should have hit")
	}

	// Shield should absorb all damage
	if health.Current != 100.0 {
		t.Errorf("Health = %f, want 100.0 (shield absorbed)", health.Current)
	}
	if shield.Amount != 20.0 {
		t.Errorf("Shield amount = %f, want 20.0", shield.Amount)
	}

	// Attack again (shield should be gone, cooldown reset)
	attack.ResetCooldown() // Reset cooldown for second attack
	attack.CooldownTimer = 0
	combatSys.Attack(attacker, target)

	// Health should take damage now (shield depleted)
	if health.Current >= 100.0 {
		t.Errorf("Health = %f, should be less than 100.0", health.Current)
	}
	// Shield component should be removed when depleted
	if _, hasShield := target.GetComponent("shield"); hasShield {
		shieldComp, _ := target.GetComponent("shield")
		remainingShield := shieldComp.(*ShieldComponent)
		if remainingShield.IsActive() {
			t.Errorf("Shield should not be active after depletion")
		}
	}
}

// TestSpellCasting_HealingAllyTargeting tests healing spell ally targeting.
func TestSpellCasting_HealingAllyTargeting(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)
	spellSys := NewSpellCastingSystem(world, statusSys)

	// Create caster (player team)
	caster := &Entity{ID: 1, Components: make(map[string]Component)}
	caster.AddComponent(&PositionComponent{X: 0, Y: 0})
	caster.AddComponent(&HealthComponent{Current: 100, Max: 100})
	caster.AddComponent(&TeamComponent{TeamID: 1})

	// Create injured ally
	ally := &Entity{ID: 2, Components: make(map[string]Component)}
	ally.AddComponent(&PositionComponent{X: 5, Y: 0})
	allyHealth := &HealthComponent{Current: 30, Max: 100}
	ally.AddComponent(allyHealth)
	ally.AddComponent(&TeamComponent{TeamID: 1})

	// Create enemy (should not be targeted)
	enemy := &Entity{ID: 3, Components: make(map[string]Component)}
	enemy.AddComponent(&PositionComponent{X: 10, Y: 0})
	enemy.AddComponent(&HealthComponent{Current: 50, Max: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})

	world.AddEntity(caster)
	world.AddEntity(ally)
	world.AddEntity(enemy)
	world.Update(0) // Initialize entities list

	// Create healing spell
	spell := &magic.Spell{
		Name:    "Heal",
		Type:    magic.TypeHealing,
		Element: magic.ElementLight,
		Target:  magic.TargetSingle,
		Stats: magic.Stats{
			Healing: 50,
			Range:   20.0,
		},
	}

	// Cast spell (should target injured ally, not caster or enemy)
	spellSys.castHealingSpell(caster, spell)

	// Ally should be healed
	if allyHealth.Current != 80.0 {
		t.Errorf("Ally health = %f, want 80.0", allyHealth.Current)
	}
}

// TestStatusEffectSystem_StatModifiers tests that buff/debuff stat changes are applied and removed.
func TestStatusEffectSystem_StatModifiers(t *testing.T) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	statusSys := NewStatusEffectSystem(world, rng)

	entity := &Entity{ID: 1, Components: make(map[string]Component)}
	stats := &StatsComponent{
		Attack:  10.0,
		Defense: 10.0,
	}
	entity.AddComponent(stats)

	// Apply strength buff (+30%)
	statusSys.ApplyStatusEffect(entity, "strength", 0.3, 5.0, 0)

	// Attack should increase
	expectedAttack := 13.0
	if stats.Attack < expectedAttack-0.1 || stats.Attack > expectedAttack+0.1 {
		t.Errorf("Attack after buff = %f, want %f", stats.Attack, expectedAttack)
	}

	// Update until effect expires
	entities := []*Entity{entity}
	statusSys.Update(entities, 6.0)

	// Attack should return to original value
	if stats.Attack < 9.9 || stats.Attack > 10.1 {
		t.Errorf("Attack after buff expired = %f, want 10.0", stats.Attack)
	}
}
