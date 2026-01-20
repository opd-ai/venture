// Package entity provides entity type definitions.
// This file defines entity and stats structs with their methods.
// Enum types moved to: enums.go
// Templates moved to: templates.go
package entity

// Stats represents the core statistics of an entity.
type Stats struct {
	// Health represents hit points
	Health int
	// MaxHealth is the maximum health value
	MaxHealth int
	// Damage is the base attack damage
	Damage int
	// Defense reduces incoming damage
	Defense int
	// Speed affects movement and attack rate
	Speed float64
	// Level represents the entity's power level
	Level int
}

// Entity represents a generated game entity (monster or NPC).
type Entity struct {
	// Name is the procedurally generated name
	Name string
	// Type categorizes the entity
	Type EntityType
	// Size indicates physical dimensions
	Size EntitySize
	// Rarity indicates how special/rare the entity is
	Rarity Rarity
	// Stats contains all numerical attributes
	Stats Stats
	// Seed is the generation seed for this entity
	Seed int64
	// Tags are additional descriptive labels
	Tags []string
}

// IsHostile returns true if the entity is hostile to players.
func (e *Entity) IsHostile() bool {
	return e.Type == TypeMonster || e.Type == TypeBoss || e.Type == TypeMinion
}

// IsBoss returns true if the entity is a boss.
func (e *Entity) IsBoss() bool {
	return e.Type == TypeBoss
}

// GetThreatLevel returns a numerical threat assessment (0-100).
func (e *Entity) GetThreatLevel() int {
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
