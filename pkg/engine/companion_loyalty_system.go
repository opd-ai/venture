package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// LoyaltyChangeReason describes why loyalty changed
type LoyaltyChangeReason int

const (
	LoyaltyReasonFed LoyaltyChangeReason = iota
	LoyaltyReasonHealed
	LoyaltyReasonDamaged
	LoyaltyReasonAbandoned
	LoyaltyReasonTimeTogether
	LoyaltyReasonOwnerDied
	LoyaltyReasonVictory
)

// String returns a human-readable reason
func (r LoyaltyChangeReason) String() string {
	switch r {
	case LoyaltyReasonFed:
		return "Fed"
	case LoyaltyReasonHealed:
		return "Healed"
	case LoyaltyReasonDamaged:
		return "Damaged"
	case LoyaltyReasonAbandoned:
		return "Abandoned"
	case LoyaltyReasonTimeTogether:
		return "Time Together"
	case LoyaltyReasonOwnerDied:
		return "Owner Died"
	case LoyaltyReasonVictory:
		return "Victory"
	default:
		return "Unknown"
	}
}

// LoyaltyChange represents a loyalty modification
type LoyaltyChange struct {
	CompanionID uint64
	Amount      float64
	Reason      LoyaltyChangeReason
}

// CompanionLoyaltySystem manages companion loyalty changes
type CompanionLoyaltySystem struct {
	world  *World
	logger *logrus.Logger

	// PendingLoyaltyChanges stores loyalty changes to apply
	PendingLoyaltyChanges []LoyaltyChange

	// TimeAccumulator tracks time for passive loyalty gain
	TimeAccumulator float64
}

// NewCompanionLoyaltySystem creates a new companion loyalty system
func NewCompanionLoyaltySystem(world *World, logger *logrus.Logger) *CompanionLoyaltySystem {
	if logger == nil {
		logger = logrus.New()
	}

	return &CompanionLoyaltySystem{
		world:                 world,
		logger:                logger,
		PendingLoyaltyChanges: make([]LoyaltyChange, 0),
		TimeAccumulator:       0.0,
	}
}

// Update processes loyalty changes and passive loyalty gain
func (s *CompanionLoyaltySystem) Update(deltaTime float64) {
	// Process pending loyalty changes
	if len(s.PendingLoyaltyChanges) > 0 {
		for _, change := range s.PendingLoyaltyChanges {
			s.applyLoyaltyChange(change)
		}
		s.PendingLoyaltyChanges = s.PendingLoyaltyChanges[:0]
	}

	// Passive loyalty gain from time spent together
	s.TimeAccumulator += deltaTime
	if s.TimeAccumulator >= 60.0 { // Every minute
		s.TimeAccumulator = 0.0
		s.applyPassiveLoyaltyGain()
	}

	// Check for low loyalty disobedience
	s.checkDisobedience()
}

// QueueLoyaltyChange adds a loyalty change to the pending queue
func (s *CompanionLoyaltySystem) QueueLoyaltyChange(change LoyaltyChange) {
	s.PendingLoyaltyChanges = append(s.PendingLoyaltyChanges, change)

	s.logger.WithFields(logrus.Fields{
		"companionID": change.CompanionID,
		"amount":      change.Amount,
		"reason":      change.Reason.String(),
	}).Debug("Loyalty change queued")
}

// ModifyLoyalty directly modifies a companion's loyalty
func (s *CompanionLoyaltySystem) ModifyLoyalty(companionID uint64, amount float64, reason LoyaltyChangeReason) {
	s.QueueLoyaltyChange(LoyaltyChange{
		CompanionID: companionID,
		Amount:      amount,
		Reason:      reason,
	})
}

// applyLoyaltyChange applies a loyalty change to a companion
func (s *CompanionLoyaltySystem) applyLoyaltyChange(change LoyaltyChange) {
	entity, ok := s.world.GetEntity(change.CompanionID)
	if !ok || entity == nil {
		s.logger.WithFields(logrus.Fields{
			"companionID": change.CompanionID,
		}).Warn("Cannot apply loyalty change: companion not found")
		return
	}

	companionCompRaw, ok := entity.GetComponent("companion")
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"companionID": change.CompanionID,
		}).Warn("Cannot apply loyalty change: no companion component")
		return
	}

	comp := companionCompRaw.(*CompanionComponent)
	oldLoyalty := comp.Loyalty

	// Apply loyalty change with clamping
	comp.Loyalty = math.Max(0.0, math.Min(100.0, comp.Loyalty+change.Amount))

	s.logger.WithFields(logrus.Fields{
		"companionID": change.CompanionID,
		"oldLoyalty":  oldLoyalty,
		"newLoyalty":  comp.Loyalty,
		"change":      change.Amount,
		"reason":      change.Reason.String(),
	}).Info("Companion loyalty changed")
}

// applyPassiveLoyaltyGain gives all companions a small loyalty boost from time together
func (s *CompanionLoyaltySystem) applyPassiveLoyaltyGain() {
	entities := s.world.GetEntitiesWith("companion", "position")

	for _, entity := range entities {
		companionCompRaw, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		comp := companionCompRaw.(*CompanionComponent)

		// Only apply if companion is near owner
		if s.isNearOwner(entity, comp.OwnerID) {
			// Small passive gain (0.5 loyalty per minute)
			comp.Loyalty = math.Min(100.0, comp.Loyalty+0.5)
		}
	}
}

// isNearOwner checks if companion is within range of its owner
func (s *CompanionLoyaltySystem) isNearOwner(companion *Entity, ownerID uint64) bool {
	owner, ok := s.world.GetEntity(ownerID)
	if !ok || owner == nil {
		return false
	}

	companionPosRaw, ok := companion.GetComponent("position")
	if !ok {
		return false
	}
	companionPos := companionPosRaw.(*PositionComponent)

	ownerPosRaw, ok := owner.GetComponent("position")
	if !ok {
		return false
	}
	ownerPos := ownerPosRaw.(*PositionComponent)

	// Within 100 units (tiles)
	dx := companionPos.X - ownerPos.X
	dy := companionPos.Y - ownerPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= 100.0
}

// checkDisobedience checks for low loyalty disobedience
func (s *CompanionLoyaltySystem) checkDisobedience() {
	entities := s.world.GetEntitiesWith("companion")

	for _, entity := range entities {
		companionCompRaw, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		comp := companionCompRaw.(*CompanionComponent)

		// Low loyalty companions may become passive or flee
		if comp.Loyalty < 30.0 {
			// Chance of switching to passive behavior
			if comp.Behavior != BehaviorPassive {
				comp.Behavior = BehaviorPassive
				s.logger.WithFields(logrus.Fields{
					"companionID": entity.ID,
					"loyalty":     comp.Loyalty,
				}).Info("Companion became passive due to low loyalty")
			}
		} else if comp.Loyalty < 50.0 {
			// At medium-low loyalty, companions may disobey aggressive commands
			if comp.Behavior == BehaviorAggressive {
				// 50% chance to switch to defensive
				if comp.Loyalty < 40.0 {
					comp.Behavior = BehaviorDefensive
					s.logger.WithFields(logrus.Fields{
						"companionID": entity.ID,
						"loyalty":     comp.Loyalty,
					}).Debug("Companion switched to defensive due to low loyalty")
				}
			}
		}
	}
}

// GetLoyalty returns the loyalty of a companion
func (s *CompanionLoyaltySystem) GetLoyalty(companionID uint64) float64 {
	entity, ok := s.world.GetEntity(companionID)
	if !ok || entity == nil {
		return 0.0
	}

	companionCompRaw, ok := entity.GetComponent("companion")
	if !ok {
		return 0.0
	}

	return companionCompRaw.(*CompanionComponent).Loyalty
}

// GetLoyaltyThreshold returns the loyalty level category
func (s *CompanionLoyaltySystem) GetLoyaltyThreshold(loyalty float64) string {
	if loyalty >= 80.0 {
		return "Devoted"
	} else if loyalty >= 60.0 {
		return "Loyal"
	} else if loyalty >= 40.0 {
		return "Neutral"
	} else if loyalty >= 20.0 {
		return "Distant"
	}
	return "Rebellious"
}
