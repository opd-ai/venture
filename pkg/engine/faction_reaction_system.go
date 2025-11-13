package engine

import (
	"math"
)

// FactionReactionSystem handles NPC behavior based on player reputation
type FactionReactionSystem struct {
	world *World
}

// NewFactionReactionSystem creates a new faction reaction system
func NewFactionReactionSystem(world *World) *FactionReactionSystem {
	return &FactionReactionSystem{world: world}
}

// Update processes faction-based reactions
func (s *FactionReactionSystem) Update(deltaTime float64) {
	// Faction reactions are event-driven, not time-based
	// This is here for consistency with other systems
}

// GetReactionLevel returns the reaction level based on reputation
// Returns: "hostile", "unfriendly", "neutral", "friendly", "honored"
func (s *FactionReactionSystem) GetReactionLevel(entityID uint64, faction string) string {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return "neutral"
	}

	repCompRaw, ok := entity.GetComponent("reputation")
	if !ok {
		return "neutral"
	}

	repComp := repCompRaw.(*ReputationComponent)
	reputation := repComp.Factions[faction]

	if reputation <= -75 {
		return "hostile"
	} else if reputation <= -25 {
		return "unfriendly"
	} else if reputation <= 25 {
		return "neutral"
	} else if reputation <= 75 {
		return "friendly"
	} else {
		return "honored"
	}
}

// GetPriceModifier returns the price multiplier based on reputation
// Hostile: 2.0x, Unfriendly: 1.5x, Neutral: 1.0x, Friendly: 0.85x, Honored: 0.7x
func (s *FactionReactionSystem) GetPriceModifier(entityID uint64, faction string) float64 {
	reaction := s.GetReactionLevel(entityID, faction)

	switch reaction {
	case "hostile":
		return 2.0
	case "unfriendly":
		return 1.5
	case "neutral":
		return 1.0
	case "friendly":
		return 0.85
	case "honored":
		return 0.7
	default:
		return 1.0
	}
}

// ShouldAttackOnSight returns true if NPC should attack player
func (s *FactionReactionSystem) ShouldAttackOnSight(entityID uint64, faction string) bool {
	return s.GetReactionLevel(entityID, faction) == "hostile"
}

// CanAcceptQuest returns true if player can accept quests from faction
func (s *FactionReactionSystem) CanAcceptQuest(entityID uint64, faction string, minReputation float64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}

	repCompRaw, ok := entity.GetComponent("reputation")
	if !ok {
		return minReputation <= 0
	}

	repComp := repCompRaw.(*ReputationComponent)
	return repComp.Factions[faction] >= minReputation
}

// GetDialogOptions returns available dialog options based on reputation
func (s *FactionReactionSystem) GetDialogOptions(entityID uint64, faction string) []string {
	reaction := s.GetReactionLevel(entityID, faction)

	baseOptions := []string{"Talk", "Leave"}

	switch reaction {
	case "hostile":
		return baseOptions // Limited options
	case "unfriendly":
		return append(baseOptions, "Trade")
	case "neutral":
		return append(baseOptions, "Trade", "Quests")
	case "friendly":
		return append(baseOptions, "Trade", "Quests", "Services")
	case "honored":
		return append(baseOptions, "Trade", "Quests", "Services", "Special Requests")
	default:
		return baseOptions
	}
}

// GetAlignmentDescription returns a text description of alignment
func (s *FactionReactionSystem) GetAlignmentDescription(entityID uint64) string {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return "True Neutral"
	}

	repCompRaw, ok := entity.GetComponent("reputation")
	if !ok {
		return "True Neutral"
	}

	repComp := repCompRaw.(*ReputationComponent)
	
	lawDesc := s.getLawAxisDescription(repComp.Alignment.LawAxis)
	goodDesc := s.getGoodAxisDescription(repComp.Alignment.GoodAxis)

	// Handle pure neutral case
	if math.Abs(repComp.Alignment.LawAxis) < 0.2 && math.Abs(repComp.Alignment.GoodAxis) < 0.2 {
		return "True Neutral"
	}

	// Handle single-axis dominance
	if math.Abs(repComp.Alignment.LawAxis) < 0.2 {
		return goodDesc
	}
	if math.Abs(repComp.Alignment.GoodAxis) < 0.2 {
		return lawDesc
	}

	// Combine both axes
	return lawDesc + " " + goodDesc
}

func (s *FactionReactionSystem) getLawAxisDescription(lawAxis float64) string {
	if lawAxis > 0.6 {
		return "Lawful"
	} else if lawAxis < -0.6 {
		return "Chaotic"
	} else {
		return "Neutral"
	}
}

func (s *FactionReactionSystem) getGoodAxisDescription(goodAxis float64) string {
	if goodAxis > 0.6 {
		return "Good"
	} else if goodAxis < -0.6 {
		return "Evil"
	} else {
		return "Neutral"
	}
}

// GetReputationThreshold returns the reputation value for a given level
func (s *FactionReactionSystem) GetReputationThreshold(level string) float64 {
	switch level {
	case "hostile":
		return -75.0
	case "unfriendly":
		return -25.0
	case "neutral":
		return 0.0
	case "friendly":
		return 25.0
	case "honored":
		return 75.0
	default:
		return 0.0
	}
}
