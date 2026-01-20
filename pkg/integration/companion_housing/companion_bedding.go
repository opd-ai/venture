package companion_housing

import "time"

// CompanionBedding represents companion rest locations in player houses.
// This file contains the CompanionBedding type and its methods for calculating
// loyalty bonuses based on bedding quality.
//
// Code relocated from: types.go

// CompanionBedding represents a companion rest location in a house.
type CompanionBedding struct {
	FurnitureID  string         // Unique furniture identifier
	HouseID      string         // Owner house identifier
	CompanionID  uint64         // Companion entity ID (0 if unassigned)
	Quality      BeddingQuality // Bedding quality tier
	LastRestTime time.Time      // Last time companion rested here
}

// LoyaltyBonus calculates daily loyalty gain for this bedding quality.
// Base loyalty gain (no housing) is 0.05/day.
// With housing: 0.05-0.2/day based on bedding quality.
func (b *CompanionBedding) LoyaltyBonus() float64 {
	return float64(b.Quality) * 0.1
}
