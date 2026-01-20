package choice_consequences

import "time"

// alignment.go contains types and logic for tracking player moral alignment.
// This includes alignment shifts from choices, alignment requirements for content,
// and methods for applying alignment changes with clamping to valid ranges.
//
// Code relocated from: types.go

// AlignmentShift represents changes to a player's moral alignment from a choice.
type AlignmentShift struct {
	GoodEvil      float64 // -1.0 (evil) to +1.0 (good)
	LawChaos      float64 // -1.0 (chaotic) to +1.0 (lawful)
	HonorDishonor float64 // -1.0 (dishonorable) to +1.0 (honorable)
}

// PlayerAlignment tracks a player's cumulative moral alignment.
type PlayerAlignment struct {
	GoodEvil      float64 // -1.0 (pure evil) to +1.0 (pure good)
	LawChaos      float64 // -1.0 (pure chaos) to +1.0 (pure law)
	HonorDishonor float64 // -1.0 (dishonorable) to +1.0 (honorable)
	UpdatedAt     int64   // Last alignment change timestamp
}

// AlignmentRequirement specifies alignment ranges needed for content access.
type AlignmentRequirement struct {
	MinGoodEvil      float64 // Minimum good/evil alignment
	MaxGoodEvil      float64 // Maximum good/evil alignment
	MinLawChaos      float64 // Minimum law/chaos alignment
	MaxLawChaos      float64 // Maximum law/chaos alignment
	MinHonorDishonor float64 // Minimum honor/dishonor alignment
	MaxHonorDishonor float64 // Maximum honor/dishonor alignment
}

// ChecksAlignment checks if player's alignment meets requirements.
func (pa *PlayerAlignment) ChecksAlignment(req *AlignmentRequirement) bool {
	if req == nil {
		return true
	}

	if pa.GoodEvil < req.MinGoodEvil || pa.GoodEvil > req.MaxGoodEvil {
		return false
	}
	if pa.LawChaos < req.MinLawChaos || pa.LawChaos > req.MaxLawChaos {
		return false
	}
	if pa.HonorDishonor < req.MinHonorDishonor || pa.HonorDishonor > req.MaxHonorDishonor {
		return false
	}

	return true
}

// ApplyShift applies an alignment shift to the player's alignment.
func (pa *PlayerAlignment) ApplyShift(shift *AlignmentShift) {
	if shift == nil {
		return
	}

	pa.GoodEvil += shift.GoodEvil
	pa.LawChaos += shift.LawChaos
	pa.HonorDishonor += shift.HonorDishonor

	// Clamp values to [-1.0, 1.0]
	pa.GoodEvil = clamp(pa.GoodEvil, -1.0, 1.0)
	pa.LawChaos = clamp(pa.LawChaos, -1.0, 1.0)
	pa.HonorDishonor = clamp(pa.HonorDishonor, -1.0, 1.0)

	pa.UpdatedAt = time.Now().Unix()
}
