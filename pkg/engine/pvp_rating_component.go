// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"encoding/json"
	"time"
)

// RankTier represents a competitive PvP rank tier.
type RankTier string

const (
	RankBronze   RankTier = "bronze"
	RankSilver   RankTier = "silver"
	RankGold     RankTier = "gold"
	RankPlatinum RankTier = "platinum"
	RankDiamond  RankTier = "diamond"
	RankMaster   RankTier = "master"
	RankLegend   RankTier = "legend"
)

// RankThreshold defines the minimum rating for each tier.
var RankThreshold = map[RankTier]int{
	RankBronze:   0,
	RankSilver:   1000,
	RankGold:     1200,
	RankPlatinum: 1400,
	RankDiamond:  1600,
	RankMaster:   1800,
	RankLegend:   2000,
}

// RankTierOrder defines the ordering of tiers from lowest to highest.
var RankTierOrder = []RankTier{
	RankBronze,
	RankSilver,
	RankGold,
	RankPlatinum,
	RankDiamond,
	RankMaster,
	RankLegend,
}

// PvPRatingComponent tracks a player's competitive PvP rating and rank.
type PvPRatingComponent struct {
	Rating       int       `json:"rating"`        // ELO rating (starting 1000)
	PeakRating   int       `json:"peak_rating"`   // Highest rating this season
	RankTier     RankTier  `json:"rank_tier"`     // Current rank tier
	RankDivision int       `json:"rank_division"` // 1-3 (I, II, III)
	Wins         int       `json:"wins"`          // Total wins this season
	Losses       int       `json:"losses"`        // Total losses this season
	SeasonID     string    `json:"season_id"`     // Current season identifier
	LastMatch    time.Time `json:"last_match"`    // Timestamp of last match
	MatchStreak  int       `json:"match_streak"`  // Positive for wins, negative for losses
}

// Type returns the component type identifier.
func (c *PvPRatingComponent) Type() string {
	return "pvp_rating"
}

// NewPvPRatingComponent creates a new PvP rating component with default values.
func NewPvPRatingComponent(seasonID string) *PvPRatingComponent {
	return &PvPRatingComponent{
		Rating:       1000,
		PeakRating:   1000,
		RankTier:     RankSilver,
		RankDivision: 3,
		Wins:         0,
		Losses:       0,
		SeasonID:     seasonID,
		LastMatch:    time.Time{},
		MatchStreak:  0,
	}
}

// GetWinRate returns the win rate as a percentage (0-100).
func (c *PvPRatingComponent) GetWinRate() float64 {
	total := c.Wins + c.Losses
	if total == 0 {
		return 0.0
	}
	return float64(c.Wins) / float64(total) * 100.0
}

// GetTotalMatches returns the total number of matches played.
func (c *PvPRatingComponent) GetTotalMatches() int {
	return c.Wins + c.Losses
}

// GetRankDisplay returns a human-readable rank string (e.g., "Gold II").
func (c *PvPRatingComponent) GetRankDisplay() string {
	tierName := tierToName(c.RankTier)
	divisionName := divisionToRoman(c.RankDivision)
	return tierName + " " + divisionName
}

// tierToName converts a RankTier to a display name.
func tierToName(tier RankTier) string {
	switch tier {
	case RankBronze:
		return "Bronze"
	case RankSilver:
		return "Silver"
	case RankGold:
		return "Gold"
	case RankPlatinum:
		return "Platinum"
	case RankDiamond:
		return "Diamond"
	case RankMaster:
		return "Master"
	case RankLegend:
		return "Legend"
	default:
		return "Unranked"
	}
}

// divisionToRoman converts a division number to Roman numerals.
func divisionToRoman(division int) string {
	switch division {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	default:
		return "III"
	}
}

// GetTierFromRating returns the appropriate tier for a given rating.
func GetTierFromRating(rating int) RankTier {
	tier := RankBronze
	for _, t := range RankTierOrder {
		if rating >= RankThreshold[t] {
			tier = t
		}
	}
	return tier
}

// GetDivisionFromRating returns the division (1-3) within a tier for a given rating.
func GetDivisionFromRating(rating int) int {
	tier := GetTierFromRating(rating)
	tierMin := RankThreshold[tier]
	tierIdx := getTierIndex(tier)

	var tierMax int
	if tierIdx >= len(RankTierOrder)-1 {
		tierMax = tierMin + 400
	} else {
		tierMax = RankThreshold[RankTierOrder[tierIdx+1]]
	}

	tierRange := tierMax - tierMin
	divisionSize := tierRange / 3
	offset := rating - tierMin

	if divisionSize <= 0 {
		return 1
	}

	division := 3 - (offset / divisionSize)
	if division < 1 {
		division = 1
	}
	if division > 3 {
		division = 3
	}
	return division
}

// getTierIndex returns the index of a tier in RankTierOrder.
func getTierIndex(tier RankTier) int {
	for i, t := range RankTierOrder {
		if t == tier {
			return i
		}
	}
	return 0
}

// Serialize converts the component to JSON bytes.
func (c *PvPRatingComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *PvPRatingComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}

// IsPlacementComplete returns true if the player has completed placement matches.
func (c *PvPRatingComponent) IsPlacementComplete() bool {
	return c.GetTotalMatches() >= 10
}

// GetPlacementProgress returns the number of placement matches completed out of 10.
func (c *PvPRatingComponent) GetPlacementProgress() (completed, total int) {
	matches := c.GetTotalMatches()
	if matches > 10 {
		matches = 10
	}
	return matches, 10
}
