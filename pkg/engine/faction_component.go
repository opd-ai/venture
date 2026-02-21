package engine

// FactionComponent represents an entity's faction membership and reputation.
// Used for tracking NPC allegiances and player reputation with different groups.
type FactionComponent struct {
	// FactionID is the unique identifier of the faction this entity belongs to
	FactionID string

	// Reputation tracks player's standing with this faction (-100 to +100)
	// -100 to -50: Hostile (attack on sight)
	// -49 to 0: Suspicious (no trading, poor prices)
	// 1 to 50: Neutral (normal interaction)
	// 51 to 100: Friendly (discounts, special quests)
	Reputation int

	// IsPlayerFaction indicates if this tracks the player's reputation
	// When true, Reputation field tracks player's standing with FactionID
	// When false, entity is a member of FactionID
	IsPlayerFaction bool
}

// Type returns the component type identifier
func (f FactionComponent) Type() string {
	return "faction"
}

// GetReputationLevel returns a string describing the reputation level
func (f FactionComponent) GetReputationLevel() string {
	if f.Reputation >= 51 {
		return "Friendly"
	} else if f.Reputation >= 1 {
		return "Neutral"
	} else if f.Reputation >= -49 {
		return "Suspicious"
	}
	return "Hostile"
}

// IsHostile returns true if reputation is hostile (attack on sight)
func (f FactionComponent) IsHostile() bool {
	return f.Reputation <= -50
}

// IsSuspicious returns true if reputation is suspicious (no trading)
func (f FactionComponent) IsSuspicious() bool {
	return f.Reputation >= -49 && f.Reputation <= 0
}

// IsNeutral returns true if reputation is neutral
func (f FactionComponent) IsNeutral() bool {
	return f.Reputation >= 1 && f.Reputation <= 50
}

// IsFriendly returns true if reputation is friendly
func (f FactionComponent) IsFriendly() bool {
	return f.Reputation >= 51
}

// GetPriceMultiplier returns the commerce price multiplier based on reputation
// Hostile: no trading (returns 0)
// Suspicious: 150% markup
// Neutral: normal price
// Friendly: progressive discount (up to 25% off at 100 reputation)
func (f FactionComponent) GetPriceMultiplier() float64 {
	if f.IsHostile() {
		return 0.0 // No trading
	}
	if f.IsSuspicious() {
		return 1.5 // 50% markup
	}
	if f.IsNeutral() {
		return 1.0 // Normal price
	}
	// Friendly: 0.75 to 1.0 (25% discount at max reputation)
	discount := float64(f.Reputation-50) / 50.0 * 0.25
	return 1.0 - discount
}

// Faction represents a procedurally generated faction in the game world
type Faction struct {
	// ID is the unique identifier for this faction
	ID string

	// Name is the procedurally generated name
	Name string

	// Type categorizes the faction
	Type FactionType

	// GenreID indicates which genre this faction belongs to
	GenreID string

	// Description provides lore about the faction
	Description string

	// Relationships maps other faction IDs to relationship values
	// -100 to -50: Enemy (at war)
	// -49 to 0: Unfriendly (distrustful)
	// 1 to 50: Neutral (trade only)
	// 51 to 100: Allied (mutual defense)
	Relationships map[string]int

	// TerritoryColor is used for visualization (RGBA)
	TerritoryColor [4]uint8

	// MemberCount is the estimated number of NPCs in this faction
	MemberCount int
}

// FactionType categorizes factions by their nature
type FactionType string

const (
	FactionTypeKingdom     FactionType = "kingdom"
	FactionTypeGuild       FactionType = "guild"
	FactionTypeCult        FactionType = "cult"
	FactionTypeCorporation FactionType = "corporation"
	FactionTypeGang        FactionType = "gang"
	FactionTypeRebels      FactionType = "rebels"
	FactionTypeMerchants   FactionType = "merchants"
)

// String returns the string representation of the faction type
func (ft FactionType) String() string {
	return string(ft)
}

// GetRelationship returns the relationship value with another faction
func (f *Faction) GetRelationship(otherFactionID string) int {
	if rel, ok := f.Relationships[otherFactionID]; ok {
		return rel
	}
	return 0 // Default neutral
}

// IsEnemy returns true if this faction is at war with another.
// Deprecated: For ECS compliance, prefer using FactionIsEnemy(f, otherFactionID) helper function.
// This method is retained for backward compatibility.
func (f *Faction) IsEnemy(otherFactionID string) bool {
	return FactionIsEnemy(f, otherFactionID)
}

// IsAlly returns true if this faction is allied with another.
// Deprecated: For ECS compliance, prefer using FactionIsAlly(f, otherFactionID) helper function.
// This method is retained for backward compatibility.
func (f *Faction) IsAlly(otherFactionID string) bool {
	return FactionIsAlly(f, otherFactionID)
}

// FactionIsEnemy returns true if the faction is at war with another faction.
// This is the ECS-compliant standalone function for checking faction enemy status.
// A faction is considered an enemy if the relationship value is -50 or below.
func FactionIsEnemy(f *Faction, otherFactionID string) bool {
	if f == nil {
		return false
	}
	return f.GetRelationship(otherFactionID) <= -50
}

// FactionIsAlly returns true if the faction is allied with another faction.
// This is the ECS-compliant standalone function for checking faction ally status.
// A faction is considered an ally if the relationship value is 51 or above.
func FactionIsAlly(f *Faction, otherFactionID string) bool {
	if f == nil {
		return false
	}
	return f.GetRelationship(otherFactionID) >= 51
}

// FactionGetRelationship returns the relationship value with another faction.
// This is the ECS-compliant standalone function for getting faction relationships.
// Returns 0 (neutral) if the faction is nil or no relationship exists.
func FactionGetRelationship(f *Faction, otherFactionID string) int {
	if f == nil {
		return 0
	}
	return f.GetRelationship(otherFactionID)
}

// ReputationChange represents an event that modifies faction reputation
type ReputationChange struct {
	// FactionID identifies which faction's reputation changes
	FactionID string

	// Amount is the reputation change (-100 to +100)
	Amount int

	// Reason describes why reputation changed
	Reason string
}

// Common reputation change amounts
const (
	ReputationKillMember    = -10 // Killing a faction member
	ReputationCompleteQuest = 15  // Completing a faction quest
	ReputationBetray        = -50 // Betraying the faction
	ReputationRescue        = 20  // Rescuing a faction member
	ReputationSteal         = -5  // Stealing from faction
	ReputationDonate        = 5   // Donating to faction
	ReputationKillEnemy     = 10  // Killing enemy of faction
	ReputationKillAlly      = -20 // Killing ally of faction
)
