package housing_crafting

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

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

// Serialize encodes the skill training facility to JSON bytes for persistence.
// Returns the encoded bytes and any error encountered during marshaling.
func (stf *SkillTrainingFacility) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"type":        "skill_training_facility",
		"facility_id": stf.ID,
		"owner_id":    stf.OwnerID,
		"house_id":    stf.HouseID,
	}).Debug("Serializing skill training facility")

	data, err := json.Marshal(stf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "skill_training_facility",
			"error": err.Error(),
		}).Error("Failed to serialize skill training facility")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"type":  "skill_training_facility",
		"bytes": len(data),
	}).Debug("Skill training facility serialized successfully")

	return data, nil
}

// Deserialize decodes the skill training facility from JSON bytes.
// Returns any error encountered during unmarshaling.
func (stf *SkillTrainingFacility) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"type":  "skill_training_facility",
		"bytes": len(data),
	}).Debug("Deserializing skill training facility")

	if err := json.Unmarshal(data, stf); err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "skill_training_facility",
			"error": err.Error(),
		}).Error("Failed to deserialize skill training facility")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"type":        "skill_training_facility",
		"facility_id": stf.ID,
		"owner_id":    stf.OwnerID,
		"house_id":    stf.HouseID,
	}).Debug("Skill training facility deserialized successfully")

	return nil
}
