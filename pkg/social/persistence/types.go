package persistence

import (
	"time"
)

// TrustLevel represents a tier of trust between players
type TrustLevel int

const (
	// TrustLevelStranger is the lowest trust level (0.0-0.3)
	TrustLevelStranger TrustLevel = iota
	// TrustLevelAcquaintance is basic familiarity (0.3-0.6)
	TrustLevelAcquaintance
	// TrustLevelFriend is established friendship (0.6-0.8)
	TrustLevelFriend
	// TrustLevelTrusted is highest trust (0.8-1.0)
	TrustLevelTrusted
)

// String returns the human-readable name of a TrustLevel
func (t TrustLevel) String() string {
	switch t {
	case TrustLevelStranger:
		return "Stranger"
	case TrustLevelAcquaintance:
		return "Acquaintance"
	case TrustLevelFriend:
		return "Friend"
	case TrustLevelTrusted:
		return "Trusted"
	default:
		return "Unknown"
	}
}

// TrustRecord stores trust information between two players
type TrustRecord struct {
	// PlayerA is the first player ID (lexicographically sorted)
	PlayerA string
	// PlayerB is the second player ID (lexicographically sorted)
	PlayerB string
	// Score is the trust value (0.0-1.0)
	Score float64
	// LastUpdate is when the trust was last modified
	LastUpdate time.Time
	// Interactions is the number of positive interactions
	Interactions int
}

// GetTrustLevel returns the trust tier for a given score
func GetTrustLevel(score float64) TrustLevel {
	if score < 0.3 {
		return TrustLevelStranger
	} else if score < 0.6 {
		return TrustLevelAcquaintance
	} else if score < 0.8 {
		return TrustLevelFriend
	}
	return TrustLevelTrusted
}

// CanTradeRarity returns true if the trust level allows trading items of the given rarity
func CanTradeRarity(level TrustLevel, rarity string) bool {
	rarityLevel := map[string]int{
		"common":    0,
		"uncommon":  1,
		"rare":      2,
		"epic":      3,
		"legendary": 4,
	}

	maxRarity := map[TrustLevel]int{
		TrustLevelStranger:     0, // common only
		TrustLevelAcquaintance: 1, // common + uncommon
		TrustLevelFriend:       2, // up to rare
		TrustLevelTrusted:      4, // all items
	}

	itemLevel, exists := rarityLevel[rarity]
	if !exists {
		return false
	}

	maxLevel, exists := maxRarity[level]
	if !exists {
		return false
	}

	return itemLevel <= maxLevel
}

const (
	// DecayRatePerDay is the trust decay rate (0.01 per day)
	DecayRatePerDay = 0.01
	// MinTrustScore is the minimum trust value
	MinTrustScore = 0.0
	// MaxTrustScore is the maximum trust value
	MaxTrustScore = 1.0
)
