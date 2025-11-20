// Package companion_housing integrates V8 housing with V4 companions, enabling
// companions to rest in player houses, gain loyalty bonuses, train skills, and
// access shared storage. This creates a home-base mechanic for companions that
// deepens the player-companion bond.
//
// Key Features:
//   - Companion bedding furniture grants daily loyalty bonuses
//   - Training areas accelerate companion skill progression
//   - Shared storage chests accessible by both player and companions
//   - Visual idle states when companions rest in houses
//
// Integration Points:
//   - V8 Housing: pkg/world/housing/ (ownership, furniture placement)
//   - V4 Companions: pkg/engine/companion_component.go (loyalty, skills)
//   - V8 Companion Learning: pkg/companion/learning/ (skill progression)
//
// Example:
//
//	manager := companion_housing.NewPetHomeManager()
//	manager.AddBedding(houseID, "companion_bed_1", 1.0) // Quality 1.0
//	bonus := manager.GetLoyaltyBonus(companionID, houseID)
//	// bonus = 0.1 per day (vs 0.05 without housing)
package companion_housing

import (
	"time"
)

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

// StorageChest represents shared inventory accessible by companions.
type StorageChest struct {
	FurnitureID    string   // Unique furniture identifier
	HouseID        string   // Owner house identifier
	Capacity       int      // Total slot count (50-100 typical)
	SharedWithPets bool     // If true, companions can deposit/withdraw
	Items          []string // Item IDs stored in chest
}

// AddItem adds an item to the chest if space available.
// Returns false if chest is full.
func (s *StorageChest) AddItem(itemID string) bool {
	if len(s.Items) >= s.Capacity {
		return false
	}
	s.Items = append(s.Items, itemID)
	return true
}

// RemoveItem removes an item from the chest if present.
// Returns false if item not found.
func (s *StorageChest) RemoveItem(itemID string) bool {
	for i, id := range s.Items {
		if id == itemID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}
	return false
}

// AvailableSlots returns number of empty slots in chest.
func (s *StorageChest) AvailableSlots() int {
	return s.Capacity - len(s.Items)
}
