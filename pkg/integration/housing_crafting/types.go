package housing_crafting

// StationType represents the type of crafting station
type StationType int

const (
	StationTypeForge StationType = iota
	StationTypeAlchemy
	StationTypeEnchanting
	StationTypeCooking
	StationTypeTailoring
	StationTypeWoodworking
	StationTypeInscription
	StationTypeEngineering
)

// String returns the human-readable name of the station type
func (st StationType) String() string {
	switch st {
	case StationTypeForge:
		return "Forge"
	case StationTypeAlchemy:
		return "Alchemy"
	case StationTypeEnchanting:
		return "Enchanting"
	case StationTypeCooking:
		return "Cooking"
	case StationTypeTailoring:
		return "Tailoring"
	case StationTypeWoodworking:
		return "Woodworking"
	case StationTypeInscription:
		return "Inscription"
	case StationTypeEngineering:
		return "Engineering"
	default:
		return "Unknown"
	}
}

// QualityTier represents the quality level of a crafting station
type QualityTier int

const (
	QualityBasic QualityTier = iota
	QualityStandard
	QualityAdvanced
	QualityMaster
)

// String returns the human-readable name of the quality tier
func (qt QualityTier) String() string {
	switch qt {
	case QualityBasic:
		return "Basic"
	case QualityStandard:
		return "Standard"
	case QualityAdvanced:
		return "Advanced"
	case QualityMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// Multiplier returns the crafting bonus multiplier for this quality tier
func (qt QualityTier) Multiplier() float64 {
	switch qt {
	case QualityBasic:
		return 1.0
	case QualityStandard:
		return 1.2
	case QualityAdvanced:
		return 1.5
	case QualityMaster:
		return 2.0
	default:
		return 1.0
	}
}

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
