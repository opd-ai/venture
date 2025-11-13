package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// AlignmentSystem manages moral alignment shifts based on player actions.
// It tracks alignment changes and updates entity alignment over time.
type AlignmentSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewAlignmentSystem creates a new alignment system.
func NewAlignmentSystem(world *World) *AlignmentSystem {
	return NewAlignmentSystemWithLogger(world, nil)
}

// NewAlignmentSystemWithLogger creates a new alignment system with a logger.
func NewAlignmentSystemWithLogger(world *World, logger *logrus.Logger) *AlignmentSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "alignment")
	}
	return &AlignmentSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes alignment for all entities (currently event-driven, not time-based).
func (s *AlignmentSystem) Update(deltaTime float64) {
	// Alignment changes are primarily event-driven through RecordDeed
	// This method is here for consistency with other systems
}

// RecordDeed records a significant action and applies alignment changes.
func (s *AlignmentSystem) RecordDeed(entityID uint64, action string, lawDelta, goodDelta float64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil // Entity doesn't exist, ignore
	}

	// Get or create reputation component
	var repComp *ReputationComponent
	comp, ok := entity.GetComponent("reputation")
	if !ok {
		repComp = &ReputationComponent{
			Factions:  make(map[string]float64),
			Alignment: Alignment{LawAxis: 0, GoodAxis: 0},
			KarmaDeed: []Deed{},
		}
		entity.AddComponent(repComp)
	} else {
		repComp = comp.(*ReputationComponent)
	}

	// Apply alignment change (clamped to -1.0 to +1.0)
	repComp.Alignment.LawAxis = math.Max(-1.0, math.Min(1.0, repComp.Alignment.LawAxis+lawDelta))
	repComp.Alignment.GoodAxis = math.Max(-1.0, math.Min(1.0, repComp.Alignment.GoodAxis+goodDelta))

	// Record deed in history
	deed := Deed{
		Action:    action,
		Timestamp: 0, // TODO: Add timestamp when time system is available
		AlignmentChange: Alignment{
			LawAxis:  lawDelta,
			GoodAxis: goodDelta,
		},
	}
	repComp.KarmaDeed = append(repComp.KarmaDeed, deed)

	// Keep deed history limited to last 100 actions
	if len(repComp.KarmaDeed) > 100 {
		repComp.KarmaDeed = repComp.KarmaDeed[len(repComp.KarmaDeed)-100:]
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":  entityID,
			"action":    action,
			"lawDelta":  lawDelta,
			"goodDelta": goodDelta,
			"lawAxis":   repComp.Alignment.LawAxis,
			"goodAxis":  repComp.Alignment.GoodAxis,
		}).Debug("deed recorded")
	}

	return nil
}

// GetAlignment returns the current alignment for an entity.
// Returns default True Neutral (0,0) if entity has no reputation component.
func (s *AlignmentSystem) GetAlignment(entityID uint64) Alignment {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return Alignment{LawAxis: 0, GoodAxis: 0}
	}

	comp, ok := entity.GetComponent("reputation")
	if !ok {
		return Alignment{LawAxis: 0, GoodAxis: 0}
	}

	repComp := comp.(*ReputationComponent)
	return repComp.Alignment
}

// GetAlignmentDescription returns a human-readable description of alignment.
func (s *AlignmentSystem) GetAlignmentDescription(entityID uint64) string {
	alignment := s.GetAlignment(entityID)

	// Threshold for axis classification
	const threshold = 0.2

	// Determine law axis
	var lawDesc string
	if alignment.LawAxis > 0.6 {
		lawDesc = "Lawful"
	} else if alignment.LawAxis < -0.6 {
		lawDesc = "Chaotic"
	} else if math.Abs(alignment.LawAxis) < threshold {
		lawDesc = "Neutral"
	}

	// Determine good axis
	var goodDesc string
	if alignment.GoodAxis > 0.6 {
		goodDesc = "Good"
	} else if alignment.GoodAxis < -0.6 {
		goodDesc = "Evil"
	} else if math.Abs(alignment.GoodAxis) < threshold {
		goodDesc = "Neutral"
	}

	// Handle True Neutral case
	if math.Abs(alignment.LawAxis) < threshold && math.Abs(alignment.GoodAxis) < threshold {
		return "True Neutral"
	}

	// Handle single-axis dominance
	if math.Abs(alignment.LawAxis) < threshold {
		return goodDesc
	}
	if math.Abs(alignment.GoodAxis) < threshold {
		return lawDesc
	}

	// Combine both axes
	return lawDesc + " " + goodDesc
}

// Common deed actions with suggested alignment shifts
const (
	DeedKillInnocent       = "Kill Innocent"        // -0.05 law, -0.1 good
	DeedKillHostile        = "Kill Hostile"         // -0.01 law, +0.02 good (defending self/others)
	DeedSteal              = "Steal"                // -0.05 law, -0.05 good
	DeedHelp               = "Help"                 // +0.01 law, +0.05 good
	DeedBreakLaw           = "Break Law"            // -0.08 law, 0 good
	DeedUpholdLaw          = "Uphold Law"           // +0.08 law, +0.02 good
	DeedLie                = "Lie"                  // -0.02 law, -0.02 good
	DeedTellTruth          = "Tell Truth"           // +0.02 law, +0.01 good
	DeedBetray             = "Betray"               // -0.1 law, -0.15 good
	DeedHonorAgreement     = "Honor Agreement"      // +0.05 law, +0.03 good
	DeedDonateToCharity    = "Donate to Charity"    // 0 law, +0.08 good
	DeedRobPoor            = "Rob Poor"             // -0.05 law, -0.12 good
	DeedProtectWeak        = "Protect Weak"         // +0.03 law, +0.10 good
	DeedExploitWeak        = "Exploit Weak"         // -0.03 law, -0.10 good
	DeedFulfillContract    = "Fulfill Contract"     // +0.06 law, +0.02 good
	DeedBreakContract      = "Break Contract"       // -0.06 law, -0.03 good
	DeedSacrificeForGreed  = "Sacrifice for Greed"  // 0 law, -0.15 good
	DeedSacrificeForOthers = "Sacrifice for Others" // +0.05 law, +0.20 good
)

// RecordCommonDeed records a common deed using predefined alignment changes.
func (s *AlignmentSystem) RecordCommonDeed(entityID uint64, deed string) error {
	switch deed {
	case DeedKillInnocent:
		return s.RecordDeed(entityID, deed, -0.05, -0.1)
	case DeedKillHostile:
		return s.RecordDeed(entityID, deed, -0.01, 0.02)
	case DeedSteal:
		return s.RecordDeed(entityID, deed, -0.05, -0.05)
	case DeedHelp:
		return s.RecordDeed(entityID, deed, 0.01, 0.05)
	case DeedBreakLaw:
		return s.RecordDeed(entityID, deed, -0.08, 0)
	case DeedUpholdLaw:
		return s.RecordDeed(entityID, deed, 0.08, 0.02)
	case DeedLie:
		return s.RecordDeed(entityID, deed, -0.02, -0.02)
	case DeedTellTruth:
		return s.RecordDeed(entityID, deed, 0.02, 0.01)
	case DeedBetray:
		return s.RecordDeed(entityID, deed, -0.1, -0.15)
	case DeedHonorAgreement:
		return s.RecordDeed(entityID, deed, 0.05, 0.03)
	case DeedDonateToCharity:
		return s.RecordDeed(entityID, deed, 0, 0.08)
	case DeedRobPoor:
		return s.RecordDeed(entityID, deed, -0.05, -0.12)
	case DeedProtectWeak:
		return s.RecordDeed(entityID, deed, 0.03, 0.10)
	case DeedExploitWeak:
		return s.RecordDeed(entityID, deed, -0.03, -0.10)
	case DeedFulfillContract:
		return s.RecordDeed(entityID, deed, 0.06, 0.02)
	case DeedBreakContract:
		return s.RecordDeed(entityID, deed, -0.06, -0.03)
	case DeedSacrificeForGreed:
		return s.RecordDeed(entityID, deed, 0, -0.15)
	case DeedSacrificeForOthers:
		return s.RecordDeed(entityID, deed, 0.05, 0.20)
	default:
		// Unknown deed, no alignment change
		return s.RecordDeed(entityID, deed, 0, 0)
	}
}
