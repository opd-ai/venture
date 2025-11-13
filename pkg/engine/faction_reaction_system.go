package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// FactionReactionSystem modifies NPC behavior based on reputation.
// It adjusts prices, aggression levels, and dialogue availability based on
// the player's standing with the NPC's faction.
type FactionReactionSystem struct {
	world  *World
	logger *logrus.Logger
}

// NewFactionReactionSystem creates a new FactionReactionSystem.
func NewFactionReactionSystem(world *World, logger *logrus.Logger) *FactionReactionSystem {
	if logger == nil {
		logger = logrus.New()
	}

	return &FactionReactionSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes faction reactions for all NPCs.
// This is called every frame but optimizations ensure minimal overhead.
func (s *FactionReactionSystem) Update(deltaTime float64) {
	// Faction reactions are checked on-demand (during interactions)
	// rather than every frame for performance reasons
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

// GetPriceModifier returns the price multiplier based on reputation.
// Returns values from 0.5 (50% discount at Revered) to 2.0 (200% markup at Hated).
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

// IsHostile returns true if the NPC should be aggressive towards the player.
// NPCs become hostile at reputation < -50 (Hostile or Hated).
func (s *FactionReactionSystem) IsHostile(playerID, npcID uint64) bool {
	npc, ok := s.world.GetEntity(npcID)
	if !ok || npc == nil {
		return false
	}

	player, ok := s.world.GetEntity(playerID)
	if !ok || player == nil {
		return false
	}

	npcFaction := s.getEntityFaction(npc)
	if npcFaction == "" {
		return false // No faction, default to non-hostile
	}

	repComp, ok := player.GetComponent("reputation")
	if !ok || repComp == nil {
		return false // No reputation data, assume neutral
	}

	reputation, ok := repComp.(*ReputationComponent)
	if !ok {
		return false
	}

	return reputation.IsHostile(npcFaction)
}

// CanTrade returns true if the NPC is willing to trade with the player.
// NPCs refuse to trade at Hostile or Hated reputation levels.
func (s *FactionReactionSystem) CanTrade(playerID, npcID uint64) bool {
	npc, ok := s.world.GetEntity(npcID)
	if !ok || npc == nil {
		return false
	}

	player, ok := s.world.GetEntity(playerID)
	if !ok || player == nil {
		return false
	}

	npcFaction := s.getEntityFaction(npc)
	if npcFaction == "" {
		return true // No faction restrictions
	}

	repComp, ok := player.GetComponent("reputation")
	if !ok || repComp == nil {
		return true // No reputation data, allow trade
	}

	reputation, ok := repComp.(*ReputationComponent)
	if !ok {
		return true
	}

	// Refuse trade if hostile or hated
	rep := reputation.GetReputation(npcFaction)
	return rep >= -50.0
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

// getEntityFaction returns the faction of an entity.
// This is a helper method that checks for a faction component.
func (s *FactionReactionSystem) getEntityFaction(entity *Entity) string {
	// Check for faction component
	comp, ok := entity.GetComponent("faction")
	if ok && comp != nil {
		// Future: Extract faction name from FactionComponent
		return ""
	}

	return ""
}
