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
// Entity is a pure data structure following ECS principles.
// For query operations, use the standalone functions: IsHostile, IsBoss, GetThreatLevel.
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
