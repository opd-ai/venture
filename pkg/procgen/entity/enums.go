// Package entity provides entity type enums and classifications.
// This file defines all enumeration types used for entity categorization.
// Originally from: types.go
package entity

// EntityType represents the classification of an entity.
type EntityType int

const (
	// TypeMonster represents hostile creatures that attack the player
	TypeMonster EntityType = iota
	// TypeNPC represents non-hostile characters (merchants, quest givers)
	TypeNPC
	// TypeBoss represents rare, powerful boss enemies
	TypeBoss
	// TypeMinion represents weak, common enemies often found in groups
	TypeMinion
)

// String returns the string representation of an entity type.
func (t EntityType) String() string {
	switch t {
	case TypeMonster:
		return "monster"
	case TypeNPC:
		return "npc"
	case TypeBoss:
		return "boss"
	case TypeMinion:
		return "minion"
	default:
		return "unknown"
	}
}

// EntitySize represents the physical size category of an entity.
type EntitySize int

const (
	// SizeTiny represents very small entities (rats, insects)
	SizeTiny EntitySize = iota
	// SizeSmall represents small entities (goblins, kobolds)
	SizeSmall
	// SizeMedium represents human-sized entities
	SizeMedium
	// SizeLarge represents large entities (ogres, bears)
	SizeLarge
	// SizeHuge represents massive entities (dragons, giants)
	SizeHuge
)

// String returns the string representation of an entity size.
func (s EntitySize) String() string {
	switch s {
	case SizeTiny:
		return "tiny"
	case SizeSmall:
		return "small"
	case SizeMedium:
		return "medium"
	case SizeLarge:
		return "large"
	case SizeHuge:
		return "huge"
	default:
		return "unknown"
	}
}

// Rarity represents how rare/special an entity is.
type Rarity int

const (
	// RarityCommon represents frequently encountered entities
	RarityCommon Rarity = iota
	// RarityUncommon represents moderately rare entities
	RarityUncommon
	// RarityRare represents rare entities with better stats
	RarityRare
	// RarityEpic represents very rare, powerful entities
	RarityEpic
	// RarityLegendary represents extremely rare, unique entities
	RarityLegendary
)

// String returns the string representation of a rarity level.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "common"
	case RarityUncommon:
		return "uncommon"
	case RarityRare:
		return "rare"
	case RarityEpic:
		return "epic"
	case RarityLegendary:
		return "legendary"
	default:
		return "unknown"
	}
}

// MerchantType represents the behavior pattern of a merchant NPC.
// Originally from: merchant.go
type MerchantType int

const (
	// MerchantFixed represents stationary shopkeepers in settlements
	MerchantFixed MerchantType = iota
	// MerchantNomadic represents wandering merchants that spawn periodically
	MerchantNomadic
)

// String returns the string representation of a merchant type.
func (m MerchantType) String() string {
	switch m {
	case MerchantFixed:
		return "fixed"
	case MerchantNomadic:
		return "nomadic"
	default:
		return "unknown"
	}
}
