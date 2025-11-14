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

// Serialize converts CompanionComponent to bytes for network transmission.
// Format: ownerID(8) + type(1) + loyalty(8) + exp(8) + level(4) + behavior(1) +
// commandCount(4) + commands(4*N) + permadeath(1) + perkCount(4) + perks(1*N) + timeWithOwner(8)
func (c *CompanionComponent) Serialize() []byte {
	// Calculate size: 8+1+8+8+4+1+4+(4*cmdCount)+1+4+(1*perkCount)+8
	cmdCount := len(c.Commands)
	perkCount := len(c.BondingPerks)
	size := 47 + (4 * cmdCount) + perkCount

	buf := make([]byte, size)
	offset := 0

	// OwnerID (8 bytes)
	writeUint64(buf[offset:], c.OwnerID)
	offset += 8

	// CompanionType (1 byte)
	buf[offset] = byte(c.CompanionType)
	offset++

	// Loyalty (8 bytes)
	writeFloat64(buf[offset:], c.Loyalty)
	offset += 8

	// Experience (8 bytes)
	writeFloat64(buf[offset:], c.Experience)
	offset += 8

	// Level (4 bytes)
	writeInt32(buf[offset:], int32(c.Level))
	offset += 4

	// Behavior (1 byte)
	buf[offset] = byte(c.Behavior)
	offset++

	// Commands (4 bytes count + 4 bytes per command)
	writeInt32(buf[offset:], int32(cmdCount))
	offset += 4
	for _, cmd := range c.Commands {
		writeInt32(buf[offset:], int32(cmd))
		offset += 4
	}

	// Permadeath (1 byte)
	writeBool(buf[offset:], c.Permadeath)
	offset++

	// BondingPerks (4 bytes count + 1 byte per perk)
	writeInt32(buf[offset:], int32(perkCount))
	offset += 4
	for _, perk := range c.BondingPerks {
		buf[offset] = byte(perk)
		offset++
	}

	// TimeWithOwner (8 bytes)
	writeFloat64(buf[offset:], c.TimeWithOwner)

	return buf
}

// Deserialize restores CompanionComponent from bytes.
func (c *CompanionComponent) Deserialize(data []byte) error {
	if len(data) < 47 {
		return ErrInvalidComponentData
	}

	offset := 0

	// OwnerID
	c.OwnerID = readUint64(data[offset:])
	offset += 8

	// CompanionType
	c.CompanionType = CompanionType(data[offset])
	offset++

	// Loyalty
	c.Loyalty = readFloat64(data[offset:])
	offset += 8

	// Experience
	c.Experience = readFloat64(data[offset:])
	offset += 8

	// Level
	c.Level = int(readInt32(data[offset:]))
	offset += 4

	// Behavior
	c.Behavior = BehaviorMode(data[offset])
	offset++

	// Commands
	cmdCount := int(readInt32(data[offset:]))
	offset += 4
	c.Commands = make([]CommandType, cmdCount)
	for i := 0; i < cmdCount; i++ {
		c.Commands[i] = CommandType(readInt32(data[offset:]))
		offset += 4
	}

	// Permadeath
	c.Permadeath = readBool(data[offset:])
	offset++

	// BondingPerks
	perkCount := int(readInt32(data[offset:]))
	offset += 4
	c.BondingPerks = make([]BondingPerk, perkCount)
	for i := 0; i < perkCount; i++ {
		c.BondingPerks[i] = BondingPerk(data[offset])
		offset++
	}

	// TimeWithOwner
	c.TimeWithOwner = readFloat64(data[offset:])

	return nil
}
