package housing_crafting

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// CraftingStation represents a crafting station placed in a player's house.
// This file contains the CraftingStation type and its methods for querying
// skill bonuses and recipe availability.
//
// Code relocated from: types.go

// CraftingStation represents a crafting station placed in a player's house
type CraftingStation struct {
	ID          string      // Unique station ID
	Type        StationType // Station type
	Quality     QualityTier // Quality tier
	OwnerID     string      // Player ID who owns the house
	HouseID     string      // House ID where station is located
	FurnitureID uint64      // Entity ID of the furniture representing this station

	// Skill bonuses provided by this station
	SkillBonus map[string]int // Skill name → XP bonus percentage (e.g., "smithing" → 50 for +50%)

	// Active recipes unlocked by this station
	ActiveRecipes []string // Recipe IDs available at this quality tier
}

// GetSkillTrainingBonus returns the XP bonus percentage for a skill
func (cs *CraftingStation) GetSkillTrainingBonus(skillName string) int {
	if cs.SkillBonus == nil {
		return 0
	}
	return cs.SkillBonus[skillName]
}

// HasRecipe checks if the station has unlocked a specific recipe
func (cs *CraftingStation) HasRecipe(recipeID string) bool {
	for _, id := range cs.ActiveRecipes {
		if id == recipeID {
			return true
		}
	}
	return false
}

// Serialize encodes the crafting station to JSON bytes for persistence.
// Returns the encoded bytes and any error encountered during marshaling.
func (cs *CraftingStation) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"type":       "crafting_station",
		"station_id": cs.ID,
		"owner_id":   cs.OwnerID,
		"house_id":   cs.HouseID,
	}).Debug("Serializing crafting station")

	data, err := json.Marshal(cs)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "crafting_station",
			"error": err.Error(),
		}).Error("Failed to serialize crafting station")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"type":  "crafting_station",
		"bytes": len(data),
	}).Debug("Crafting station serialized successfully")

	return data, nil
}

// Deserialize decodes the crafting station from JSON bytes.
// Returns any error encountered during unmarshaling.
func (cs *CraftingStation) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"type":  "crafting_station",
		"bytes": len(data),
	}).Debug("Deserializing crafting station")

	if err := json.Unmarshal(data, cs); err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "crafting_station",
			"error": err.Error(),
		}).Error("Failed to deserialize crafting station")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"type":       "crafting_station",
		"station_id": cs.ID,
		"owner_id":   cs.OwnerID,
		"house_id":   cs.HouseID,
	}).Debug("Crafting station deserialized successfully")

	return nil
}
