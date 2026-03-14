// types.go defines data types for the companion housing integration.
// See doc.go for the full package overview and usage examples.
package companion_housing

// Shared type definitions for the companion_housing package.
// This file contains BeddingQuality and TrainingAreaType enumerations used
// throughout the package for categorizing furniture and determining bonuses.

// BeddingQuality represents the quality of companion bedding furniture.
// Higher quality provides larger loyalty bonuses.
type BeddingQuality float64

const (
	BeddingBasic    BeddingQuality = 0.5 // +0.05 loyalty/day
	BeddingStandard BeddingQuality = 1.0 // +0.1 loyalty/day
	BeddingAdvanced BeddingQuality = 1.5 // +0.15 loyalty/day
	BeddingLuxury   BeddingQuality = 2.0 // +0.2 loyalty/day
)

// TrainingAreaType represents specialized training furniture types.
type TrainingAreaType string

const (
	TrainingCombat    TrainingAreaType = "combat_dummy"    // Melee/ranged training
	TrainingAgility   TrainingAreaType = "agility_course"  // Speed/dodge training
	TrainingMagic     TrainingAreaType = "magic_focus"     // Spell training
	TrainingObedience TrainingAreaType = "obedience_post"  // Command training
	TrainingStrength  TrainingAreaType = "strength_rack"   // Power training
	TrainingEndurance TrainingAreaType = "endurance_wheel" // Stamina training
)

// String returns human-readable training area names.
func (t TrainingAreaType) String() string {
	switch t {
	case TrainingCombat:
		return "Combat Training Dummy"
	case TrainingAgility:
		return "Agility Obstacle Course"
	case TrainingMagic:
		return "Magic Focus Crystal"
	case TrainingObedience:
		return "Obedience Training Post"
	case TrainingStrength:
		return "Strength Training Rack"
	case TrainingEndurance:
		return "Endurance Training Wheel"
	default:
		return "Unknown Training Area"
	}
}

// XPMultiplier returns the skill XP multiplier for this training area type.
// Base companion skill training is 1.0x, training areas provide 1.25x-1.5x.
func (t TrainingAreaType) XPMultiplier() float64 {
	switch t {
	case TrainingCombat, TrainingMagic:
		return 1.5 // High-priority combat skills
	case TrainingAgility, TrainingStrength:
		return 1.35 // Medium-priority physical skills
	case TrainingObedience, TrainingEndurance:
		return 1.25 // Support skills
	default:
		return 1.0
	}
}
