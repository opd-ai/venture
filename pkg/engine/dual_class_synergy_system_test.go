package engine

import (
	"testing"
)

// TestDualClassSynergySystem_Creation tests system initialization.
func TestDualClassSynergySystem_Creation(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)

	if system == nil {
		t.Fatal("NewDualClassSynergySystem returned nil")
	}

	if len(system.synergies) == 0 {
		t.Error("No synergies were initialized")
	}

	// We define unique synergies for each ordered pair (30 total)
	// 6 base classes × 5 other classes = 30 unique synergies
	synergies := system.GetAllSynergies()
	if len(synergies) < 15 {
		t.Errorf("Expected at least 15 unique synergies, got %d", len(synergies))
	}
}

// TestDualClassSynergySystem_GenreMultipliers tests genre-based bonus scaling.
func TestDualClassSynergySystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre          string
		wantMultiplier float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.9},
		{"horror", 1.1},
		{"cyberpunk", 0.95},
		{"postapoc", 1.05},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			system := NewDualClassSynergySystem(world, 12345)
			system.SetGenre(tt.genre)

			if system.bonusMultiplier != tt.wantMultiplier {
				t.Errorf("bonusMultiplier = %f, want %f", system.bonusMultiplier, tt.wantMultiplier)
			}
		})
	}
}

// TestDualClassSynergySystem_GetSynergy tests synergy lookup.
func TestDualClassSynergySystem_GetSynergy(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)

	tests := []struct {
		name      string
		primary   CharacterClass
		secondary CharacterClass
		wantName  string
		wantNil   bool
	}{
		{
			name:      "Warrior+Mage = Battle Arcanist",
			primary:   ClassWarrior,
			secondary: ClassMage,
			wantName:  "Battle Arcanist",
			wantNil:   false,
		},
		{
			name:      "Mage+Warrior = Arcane Duelist",
			primary:   ClassMage,
			secondary: ClassWarrior,
			wantName:  "Arcane Duelist",
			wantNil:   false,
		},
		{
			name:      "Rogue+Cleric = Divine Agent",
			primary:   ClassRogue,
			secondary: ClassCleric,
			wantName:  "Divine Agent",
			wantNil:   false,
		},
		{
			name:      "Necromancer+Ranger = Plague Doctor",
			primary:   ClassNecromancer,
			secondary: ClassRanger,
			wantName:  "Plague Doctor",
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synergy := system.GetSynergy(tt.primary, tt.secondary)

			if tt.wantNil {
				if synergy != nil {
					t.Errorf("GetSynergy() = %v, want nil", synergy.Name)
				}
				return
			}

			if synergy == nil {
				t.Fatal("GetSynergy() returned nil, want synergy")
			}

			if synergy.Name != tt.wantName {
				t.Errorf("GetSynergy().Name = %q, want %q", synergy.Name, tt.wantName)
			}
		})
	}
}

// TestDualClassSynergySystem_ApplyBonus tests synergy bonus application.
func TestDualClassSynergySystem_ApplyBonus(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)
	system.updateInterval = 0 // Immediate updates for testing

	// Create entity with dual-class
	entity := NewEntity(1)
	secondaryClass := ClassMage
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		SecondaryClass: &secondaryClass,
		SecondaryLevel: 1,
	})
	entity.AddComponent(&StatsComponent{
		Attack:     10,
		Defense:    8,
		MagicPower: 5,
		CritChance: 0.05,
	})
	entity.AddComponent(&BaseStatsComponent{BaseSpeed: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100})

	// Get initial values
	statsComp, _ := entity.GetComponent("stats")
	stats := statsComp.(*StatsComponent)
	initialAttack := stats.Attack
	initialMagicPower := stats.MagicPower

	// Run update
	system.Update([]*Entity{entity}, 1.0)

	// Verify bonuses applied
	if stats.Attack <= initialAttack {
		t.Errorf("Attack not increased: got %f, want > %f", stats.Attack, initialAttack)
	}
	if stats.MagicPower <= initialMagicPower {
		t.Errorf("MagicPower not increased: got %f, want > %f", stats.MagicPower, initialMagicPower)
	}

	// Verify tracking
	if !system.HasActiveSynergy(entity.ID) {
		t.Error("Entity should have active synergy")
	}

	synergyName := system.GetActiveSynergyName(entity.ID)
	if synergyName != "Battle Arcanist" {
		t.Errorf("GetActiveSynergyName() = %q, want %q", synergyName, "Battle Arcanist")
	}
}

// TestDualClassSynergySystem_NoSecondaryClass tests entities without dual-class.
func TestDualClassSynergySystem_NoSecondaryClass(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)
	system.updateInterval = 0

	// Create entity without dual-class
	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15, // Below level 20, no secondary
		SecondaryClass: nil,
	})
	entity.AddComponent(&StatsComponent{Attack: 10, Defense: 8})

	statsComp, _ := entity.GetComponent("stats")
	stats := statsComp.(*StatsComponent)
	initialAttack := stats.Attack

	// Run update
	system.Update([]*Entity{entity}, 1.0)

	// Verify no bonuses applied
	if stats.Attack != initialAttack {
		t.Errorf("Attack changed for non-dual-class: got %f, want %f", stats.Attack, initialAttack)
	}

	if system.HasActiveSynergy(entity.ID) {
		t.Error("Non-dual-class entity should not have active synergy")
	}
}

// TestDualClassSynergySystem_RegenEffects tests health/mana regeneration.
func TestDualClassSynergySystem_RegenEffects(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)
	system.updateInterval = 0

	// Use Warrior+Cleric which has HealthRegenRate
	entity := NewEntity(1)
	secondaryClass := ClassCleric
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		SecondaryClass: &secondaryClass,
		SecondaryLevel: 1,
	})
	entity.AddComponent(&StatsComponent{Attack: 10, Defense: 8})
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100}) // Start at half health
	entity.AddComponent(&ManaComponent{Current: 25, Max: 100})

	// First update applies bonus
	system.Update([]*Entity{entity}, 1.0)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	afterFirstUpdate := health.Current

	// Second update should trigger regen
	system.Update([]*Entity{entity}, 1.0)

	if health.Current <= afterFirstUpdate {
		t.Errorf("Health not regenerating: got %f, want > %f", health.Current, afterFirstUpdate)
	}
}

// TestDualClassSynergySystem_BonusRemoval tests removing bonuses when losing dual-class.
func TestDualClassSynergySystem_BonusRemoval(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)
	system.updateInterval = 0

	// Create dual-classed entity
	entity := NewEntity(1)
	secondaryClass := ClassMage
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		SecondaryClass: &secondaryClass,
	})
	entity.AddComponent(&StatsComponent{Attack: 10, Defense: 8, MagicPower: 5})

	// Apply bonus
	system.Update([]*Entity{entity}, 1.0)

	if !system.HasActiveSynergy(entity.ID) {
		t.Fatal("Entity should have synergy after first update")
	}

	// Remove secondary class
	comp, _ := entity.GetComponent("class_progression")
	progression := comp.(*ClassProgressionComponent)
	progression.SecondaryClass = nil

	// Update should remove bonus
	system.Update([]*Entity{entity}, 1.0)

	if system.HasActiveSynergy(entity.ID) {
		t.Error("Entity should not have synergy after losing dual-class")
	}
}

// TestDualClassSynergySystem_AllClassCombinations tests that all base class pairs have synergies.
func TestDualClassSynergySystem_AllClassCombinations(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)

	baseClasses := []CharacterClass{
		ClassWarrior, ClassMage, ClassRogue, ClassRanger, ClassCleric, ClassNecromancer,
	}

	for _, primary := range baseClasses {
		for _, secondary := range baseClasses {
			if primary == secondary {
				continue // Same class, no synergy expected
			}

			t.Run(primary.String()+"+"+secondary.String(), func(t *testing.T) {
				synergy := system.GetSynergy(primary, secondary)
				if synergy == nil {
					t.Errorf("No synergy defined for %s + %s", primary.String(), secondary.String())
					return
				}

				// Verify synergy has at least one bonus
				hasBonus := synergy.AttackBonus > 0 ||
					synergy.DefenseBonus > 0 ||
					synergy.ManaRegenRate > 0 ||
					synergy.CritBonus > 0 ||
					synergy.HealthRegenRate > 0 ||
					synergy.SpeedBonus > 0 ||
					synergy.SpellDamageBonus > 0

				if !hasBonus {
					t.Errorf("Synergy %q has no bonuses", synergy.Name)
				}

				// Verify synergy has a name
				if synergy.Name == "" {
					t.Error("Synergy has empty name")
				}
			})
		}
	}
}

// TestDualClassSynergySystem_SynergyBalance tests that synergies are reasonably balanced.
func TestDualClassSynergySystem_SynergyBalance(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)

	synergies := system.GetAllSynergies()

	for _, syn := range synergies {
		t.Run(syn.Name, func(t *testing.T) {
			// Check attack bonus is reasonable (0-10)
			if syn.AttackBonus < 0 || syn.AttackBonus > 10 {
				t.Errorf("AttackBonus %f out of range [0, 10]", syn.AttackBonus)
			}

			// Check defense bonus is reasonable (0-10)
			if syn.DefenseBonus < 0 || syn.DefenseBonus > 10 {
				t.Errorf("DefenseBonus %f out of range [0, 10]", syn.DefenseBonus)
			}

			// Check crit bonus is reasonable (0-0.2 = 20%)
			if syn.CritBonus < 0 || syn.CritBonus > 0.2 {
				t.Errorf("CritBonus %f out of range [0, 0.2]", syn.CritBonus)
			}

			// Check speed bonus is reasonable (0-0.2 = 20%)
			if syn.SpeedBonus < 0 || syn.SpeedBonus > 0.2 {
				t.Errorf("SpeedBonus %f out of range [0, 0.2]", syn.SpeedBonus)
			}

			// Check health regen is reasonable (0-2 per second)
			if syn.HealthRegenRate < 0 || syn.HealthRegenRate > 2 {
				t.Errorf("HealthRegenRate %f out of range [0, 2]", syn.HealthRegenRate)
			}

			// Check mana regen is reasonable (0-2 per second)
			if syn.ManaRegenRate < 0 || syn.ManaRegenRate > 2 {
				t.Errorf("ManaRegenRate %f out of range [0, 2]", syn.ManaRegenRate)
			}
		})
	}
}

// TestDualClassSynergySystem_UpdateInterval tests that updates respect the interval.
func TestDualClassSynergySystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)
	system.updateInterval = 1.0 // 1 second

	entity := NewEntity(1)
	secondaryClass := ClassMage
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		SecondaryClass: &secondaryClass,
	})
	entity.AddComponent(&StatsComponent{Attack: 10})

	// First call with small delta - should not trigger
	system.Update([]*Entity{entity}, 0.1)
	if system.HasActiveSynergy(entity.ID) {
		t.Error("Synergy applied before interval elapsed")
	}

	// Accumulate more time
	system.Update([]*Entity{entity}, 0.5)
	if system.HasActiveSynergy(entity.ID) {
		t.Error("Synergy applied before interval elapsed (0.6s)")
	}

	// Now exceed interval
	system.Update([]*Entity{entity}, 0.5)
	if !system.HasActiveSynergy(entity.ID) {
		t.Error("Synergy not applied after interval elapsed (1.1s)")
	}
}

// TestDualClassSynergySystem_Count tests the active synergy count tracking.
func TestDualClassSynergySystem_Count(t *testing.T) {
	world := NewWorld()
	system := NewDualClassSynergySystem(world, 12345)
	system.updateInterval = 0

	if system.GetActiveSynergyCount() != 0 {
		t.Errorf("Initial count should be 0, got %d", system.GetActiveSynergyCount())
	}

	// Add first dual-class entity
	e1 := NewEntity(1)
	secondaryClass1 := ClassMage
	e1.AddComponent(&ClassProgressionComponent{
		Class: ClassWarrior, Level: 20, SecondaryClass: &secondaryClass1,
	})
	e1.AddComponent(&StatsComponent{Attack: 10})

	// Add second dual-class entity
	e2 := NewEntity(2)
	secondaryClass2 := ClassCleric
	e2.AddComponent(&ClassProgressionComponent{
		Class: ClassRogue, Level: 20, SecondaryClass: &secondaryClass2,
	})
	e2.AddComponent(&StatsComponent{Attack: 10})

	// Add non-dual-class entity
	e3 := NewEntity(3)
	e3.AddComponent(&ClassProgressionComponent{
		Class: ClassMage, Level: 15, SecondaryClass: nil,
	})
	e3.AddComponent(&StatsComponent{Attack: 10})

	system.Update([]*Entity{e1, e2, e3}, 1.0)

	if system.GetActiveSynergyCount() != 2 {
		t.Errorf("Count should be 2, got %d", system.GetActiveSynergyCount())
	}
}
