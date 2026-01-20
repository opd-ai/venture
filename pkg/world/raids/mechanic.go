// MechanicGenerator generates procedural boss mechanics (summons, debuffs, AoE).
// This file creates diverse boss abilities that scale with raid tier and
// provide varied encounter mechanics.
package raids

import (
	"fmt"
	"math/rand"
	"time"
)

// MechanicGenerator generates procedural boss mechanics.
type MechanicGenerator struct{}

// NewMechanicGenerator creates a new mechanic generator.
func NewMechanicGenerator() *MechanicGenerator {
	return &MechanicGenerator{}
}

// GenerateMechanic creates a boss mechanic.
func (m *MechanicGenerator) GenerateMechanic(rng *rand.Rand, tier RaidTier, index int) BossMechanic {
	// Select mechanic type based on index
	mechanicTypes := []MechanicType{
		MechanicInstant,
		MechanicGroundEffect,
		MechanicSummon,
		MechanicDebuff,
		MechanicChanneled,
		MechanicPeriodic,
		MechanicBuff,
	}

	mechType := mechanicTypes[index%len(mechanicTypes)]

	// Generate mechanic properties based on type
	mechanic := BossMechanic{
		ID:   fmt.Sprintf("mechanic-%d", index),
		Type: mechType,
	}

	baseDamage := 100 + (int(tier) * 50)

	switch mechType {
	case MechanicSummon:
		mechanic.Name = "Summon Adds"
		mechanic.Description = "Boss summons additional enemies to assist"
		mechanic.Cooldown = time.Duration(20+rng.Intn(20)) * time.Second
		mechanic.Damage = 0 // Adds deal damage, not the summon itself
		mechanic.AoE = false
		mechanic.Radius = 0

	case MechanicGroundEffect:
		mechanic.Name = "Ground Effect"
		mechanic.Description = "Creates a damaging area on the ground"
		mechanic.Cooldown = time.Duration(15+rng.Intn(15)) * time.Second
		mechanic.Damage = baseDamage
		mechanic.AoE = true
		mechanic.Radius = 3.0 + float64(rng.Intn(4))

	case MechanicDebuff:
		mechanic.Name = "Debuff"
		mechanic.Description = "Applies a harmful status effect to players"
		mechanic.Cooldown = time.Duration(25+rng.Intn(15)) * time.Second
		mechanic.Damage = baseDamage / 2
		mechanic.AoE = true
		mechanic.Radius = 5.0 + float64(rng.Intn(5))

	case MechanicBuff:
		mechanic.Name = "Enrage"
		mechanic.Description = "Boss gains increased damage and speed"
		mechanic.Cooldown = time.Duration(45+rng.Intn(15)) * time.Second
		mechanic.Damage = 0
		mechanic.AoE = false
		mechanic.Radius = 0

	case MechanicChanneled:
		mechanic.Name = "Channeled Beam"
		mechanic.Description = "Channels a beam of energy at a target"
		mechanic.Cooldown = time.Duration(30+rng.Intn(20)) * time.Second
		mechanic.Damage = baseDamage * 2
		mechanic.AoE = false
		mechanic.Radius = 0

	case MechanicInstant:
		mechanic.Name = "Instant Strike"
		mechanic.Description = "Instant high-damage attack on primary target"
		mechanic.Cooldown = time.Duration(10+rng.Intn(10)) * time.Second
		mechanic.Damage = baseDamage * 3
		mechanic.AoE = false
		mechanic.Radius = 0

	case MechanicPeriodic:
		mechanic.Name = "Periodic Pulse"
		mechanic.Description = "Deals damage in waves around the boss"
		mechanic.Cooldown = time.Duration(20+rng.Intn(20)) * time.Second
		mechanic.Damage = baseDamage / 3
		mechanic.AoE = true
		mechanic.Radius = 8.0 + float64(rng.Intn(4))
	}

	// Scale damage for tier
	mechanic.Damage = int(float64(mechanic.Damage) * tier.DifficultyMultiplier() * 0.5)

	return mechanic
}
