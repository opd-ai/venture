package housing_crafting

// SkillTrainingFacility represents dedicated skill training areas in player houses.
// This file contains the SkillTrainingFacility type and its methods for checking
// trainable skills and calculating XP bonuses.
//
// Code relocated from: types.go

// SkillTrainingFacility represents a dedicated skill training area in a house
type SkillTrainingFacility struct {
	ID          string // Unique facility ID
	OwnerID     string // Player ID who owns the house
	HouseID     string // House ID where facility is located
	FurnitureID uint64 // Entity ID of the furniture representing this facility

	// Skills that can be trained here
	TrainableSkills []string // Skill names (e.g., "smithing", "alchemy")

	// XP multiplier for training (1.5 = +50% XP)
	XPMultiplier float64
}

// CanTrainSkill checks if a skill can be trained at this facility
func (stf *SkillTrainingFacility) CanTrainSkill(skillName string) bool {
	for _, skill := range stf.TrainableSkills {
		if skill == skillName {
			return true
		}
	}
	return false
}

// GetXPBonus calculates the XP bonus for a given skill
func (stf *SkillTrainingFacility) GetXPBonus(baseXP float64) float64 {
	if stf.XPMultiplier <= 0 {
		return baseXP
	}
	return baseXP * stf.XPMultiplier
}
