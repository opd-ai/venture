// Package entity provides query functions for entity analysis.
// This file contains standalone functions that operate on Entity data,
// maintaining ECS compliance by keeping logic separate from data structures.
package entity

// IsHostile returns true if the entity is hostile to players.
// Hostile entities include monsters, bosses, and minions.
func IsHostile(e *Entity) bool {
	if e == nil {
		return false
	}
	return e.Type == TypeMonster || e.Type == TypeBoss || e.Type == TypeMinion
}

// IsBoss returns true if the entity is a boss.
func IsBoss(e *Entity) bool {
	if e == nil {
		return false
	}
	return e.Type == TypeBoss
}

// GetThreatLevel returns a numerical threat assessment (0-100) for an entity.
// Threat is calculated based on stats and entity type.
func GetThreatLevel(e *Entity) int {
	if e == nil {
		return 0
	}
	// Calculate threat based on stats and type
	baseThreat := e.Stats.Health/10 + e.Stats.Damage*5 + e.Stats.Defense*2

	// Modify based on type before applying level
	typeMultiplier := 1.0
	switch e.Type {
	case TypeBoss:
		typeMultiplier = 3.0
	case TypeMonster:
		typeMultiplier = 2.0
	case TypeMinion:
		typeMultiplier = 0.5
	}

	threat := int(float64(baseThreat) * typeMultiplier * float64(e.Stats.Level) * 0.1)

	// Cap at 100
	if threat > 100 {
		threat = 100
	}

	return threat
}
