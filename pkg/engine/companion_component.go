package engine

// CompanionType represents different types of companions
type CompanionType int

const (
	CompanionTypePet CompanionType = iota
	CompanionTypeSummon
	CompanionTypeHireling
	CompanionTypeElemental
	CompanionTypeUndead
	CompanionTypeRobot
	CompanionTypeSpirit
	CompanionTypeInsect
)

// BehaviorMode defines companion AI behavior
type BehaviorMode int

const (
	BehaviorAggressive BehaviorMode = iota
	BehaviorDefensive
	BehaviorPassive
)

// CommandType represents commands that can be given to companions
type CommandType int

const (
	CommandFollow CommandType = iota
	CommandStay
	CommandAttack
	CommandDefend
	CommandGather
	CommandScout
)

// CompanionComponent tracks companion state
type CompanionComponent struct {
	OwnerID       uint64
	CompanionType CompanionType
	Loyalty       float64 // 0-100, affects obedience
	Experience    float64
	Level         int
	Behavior      BehaviorMode
	Commands      []CommandType
	Permadeath    bool          // If true, companion dies permanently
	BondingPerks  []BondingPerk // Unlocked perks based on loyalty
	TimeWithOwner float64       // Total time spent near owner (for bonding)
}

// BondingPerk represents a perk unlocked through bonding
type BondingPerk int

const (
	// PerkNone is a placeholder for no perk
	PerkNone BondingPerk = iota
	// PerkExtraHealth increases companion max HP by 20%
	PerkExtraHealth
	// PerkExtraDamage increases companion attack by 15%
	PerkExtraDamage
	// PerkFasterLearning doubles skill learning rate
	PerkFasterLearning
	// PerkLoyalGuard gives companion 30% damage reduction
	PerkLoyalGuard
	// PerkSharedExperience gives owner 10% of companion's XP
	PerkSharedExperience
	// PerkAutoRevive allows companion to revive once per day
	PerkAutoRevive
)

// String returns the perk name
func (p BondingPerk) String() string {
	switch p {
	case PerkExtraHealth:
		return "Extra Health"
	case PerkExtraDamage:
		return "Extra Damage"
	case PerkFasterLearning:
		return "Faster Learning"
	case PerkLoyalGuard:
		return "Loyal Guard"
	case PerkSharedExperience:
		return "Shared Experience"
	case PerkAutoRevive:
		return "Auto Revive"
	default:
		return "None"
	}
}

// Type returns the component type
func (c CompanionComponent) Type() string {
	return "companion"
}

// HasPerk checks if companion has a specific bonding perk
func (c *CompanionComponent) HasPerk(perk BondingPerk) bool {
	for _, p := range c.BondingPerks {
		if p == perk {
			return true
		}
	}
	return false
}

// AddPerk adds a bonding perk if not already present
func (c *CompanionComponent) AddPerk(perk BondingPerk) {
	if !c.HasPerk(perk) {
		c.BondingPerks = append(c.BondingPerks, perk)
	}
}

// CompanionStatsComponent holds companion-specific stats
type CompanionStatsComponent struct {
	Attack  float64
	Defense float64
	Speed   float64
	HP      float64
	MaxHP   float64
}

// Type returns the component type
func (c CompanionStatsComponent) Type() string {
	return "companionstats"
}
