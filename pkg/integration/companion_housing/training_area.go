package companion_housing

import "time"

// TrainingArea represents skill training locations in player houses.
// This file contains the TrainingArea type and its methods for calculating
// XP bonuses based on training area specialization.
//
// Code relocated from: types.go

// TrainingArea represents a skill training location in a house.
type TrainingArea struct {
	FurnitureID    string               // Unique furniture identifier
	HouseID        string               // Owner house identifier
	Type           TrainingAreaType     // Training area specialization
	ActiveSessions map[uint64]time.Time // Companion ID → session start time
}

// XPBonus calculates the XP multiplier for training in this area.
// Default companion skill XP is 1.0x, training areas provide 1.25x-1.5x.
func (t *TrainingArea) XPBonus() float64 {
	return t.Type.XPMultiplier()
}
