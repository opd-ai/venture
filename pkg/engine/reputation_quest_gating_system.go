// Package engine provides reputation-based quest gating system.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationQuestGatingSystem monitors reputation changes and updates quest availability.
// It integrates with the reputation system to gate quests behind faction standing requirements.
type ReputationQuestGatingSystem struct {
	world  *World
	logger *logrus.Logger
	rng    *rand.Rand
}

// NewReputationQuestGatingSystem creates a new reputation quest gating system.
func NewReputationQuestGatingSystem(world *World, logger *logrus.Logger, seed int64) *ReputationQuestGatingSystem {
	if logger == nil {
		logger = logrus.New()
	}

	logger.WithFields(logrus.Fields{
		"system_name": "reputation_quest_gating",
		"seed":        seed,
	}).Debug("Creating reputation quest gating system")

	return &ReputationQuestGatingSystem{
		world:  world,
		logger: logger,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// Update checks all entities with reputation and quest gating components
// and updates quest availability based on current reputation standings.
func (s *ReputationQuestGatingSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if entity == nil {
			continue
		}

		// Entity needs both reputation and quest gating components
		repComp := s.getReputationComponent(entity)
		gatingComp := s.getGatingComponent(entity)

		if repComp == nil || gatingComp == nil {
			continue
		}

		// Check each gated quest against current reputation
		s.updateQuestAvailability(entity, repComp, gatingComp)

		// Process notifications
		s.processNotifications(entity, gatingComp)
	}
}

// getReputationComponent retrieves the reputation component from an entity.
func (s *ReputationQuestGatingSystem) getReputationComponent(entity *Entity) *ReputationComponent {
	comp, ok := entity.GetComponent("reputation")
	if !ok || comp == nil {
		return nil
	}
	repComp, ok := comp.(*ReputationComponent)
	if !ok {
		return nil
	}
	return repComp
}

// getGatingComponent retrieves the quest gating component from an entity.
func (s *ReputationQuestGatingSystem) getGatingComponent(entity *Entity) *ReputationQuestGatingComponent {
	comp, ok := entity.GetComponent("reputation_quest_gating")
	if !ok || comp == nil {
		return nil
	}
	gatingComp, ok := comp.(*ReputationQuestGatingComponent)
	if !ok {
		return nil
	}
	return gatingComp
}

// updateQuestAvailability checks each registered quest and updates unlock status.
func (s *ReputationQuestGatingSystem) updateQuestAvailability(
	entity *Entity,
	repComp *ReputationComponent,
	gatingComp *ReputationQuestGatingComponent,
) {
	for questID, gatedQuest := range gatingComp.GatedQuests {
		// Skip already unlocked or locked out quests
		if gatingComp.UnlockedQuests[questID] {
			continue
		}
		if _, lockedOut := gatingComp.LockedOutQuests[questID]; lockedOut {
			continue
		}

		// Check if quest requirements are now met
		if gatingComp.IsQuestAvailable(questID, repComp.Factions) {
			gatingComp.MarkQuestUnlocked(questID)

			s.logger.WithFields(logrus.Fields{
				"system_name": "reputation_quest_gating",
				"entity_id":   entity.ID,
				"quest_id":    questID,
				"faction_id":  gatedQuest.FactionID,
			}).Info("Quest unlocked due to reputation")
		}
	}
}

// processNotifications handles recent unlock/lockout notifications.
func (s *ReputationQuestGatingSystem) processNotifications(entity *Entity, gatingComp *ReputationQuestGatingComponent) {
	unlocks := gatingComp.GetRecentUnlocks()
	for _, questID := range unlocks {
		quest := gatingComp.GatedQuests[questID]
		if quest != nil && quest.UnlockMessage != "" {
			s.logger.WithFields(logrus.Fields{
				"system_name": "reputation_quest_gating",
				"entity_id":   entity.ID,
				"quest_id":    questID,
				"message":     quest.UnlockMessage,
			}).Debug("Quest unlock notification")
		}
	}

	lockouts := gatingComp.GetRecentLockouts()
	for _, questID := range lockouts {
		reason := gatingComp.LockedOutQuests[questID]
		s.logger.WithFields(logrus.Fields{
			"system_name": "reputation_quest_gating",
			"entity_id":   entity.ID,
			"quest_id":    questID,
			"reason":      reason,
		}).Debug("Quest lockout notification")
	}
}

// RegisterFactionQuests registers a set of quests gated by faction reputation.
// This is a convenience method for bulk registration during game initialization.
func (s *ReputationQuestGatingSystem) RegisterFactionQuests(
	entity *Entity,
	factionID string,
	questConfigs []FactionQuestConfig,
) {
	gatingComp := s.getGatingComponent(entity)
	if gatingComp == nil {
		// Create component if missing
		gatingComp = NewReputationQuestGatingComponent()
		entity.AddComponent(gatingComp)
	}

	for _, config := range questConfigs {
		gatedQuest := &GatedQuest{
			QuestID:   config.QuestID,
			FactionID: factionID,
			Requirements: []QuestReputationRequirement{
				{
					FactionID:      factionID,
					MinTier:        config.MinTier,
					MaxTier:        TierRevered, // No max by default
					FailureMessage: config.FailureMessage,
				},
			},
			UnlockMessage:         config.UnlockMessage,
			RewardReputationBonus: config.ReputationBonus,
			IsExclusive:           config.IsExclusive,
			ExcludesFactions:      config.ExcludesFactions,
		}
		gatingComp.RegisterGatedQuest(gatedQuest)
	}

	s.logger.WithFields(logrus.Fields{
		"system_name": "reputation_quest_gating",
		"entity_id":   entity.ID,
		"faction_id":  factionID,
		"quest_count": len(questConfigs),
	}).Debug("Registered faction quests")
}

// FactionQuestConfig is a simplified configuration for faction-gated quests.
type FactionQuestConfig struct {
	QuestID          string
	MinTier          ReputationTier
	FailureMessage   string
	UnlockMessage    string
	ReputationBonus  float64
	IsExclusive      bool
	ExcludesFactions []string
}

// OnQuestCompleted should be called when a gated quest is completed.
// It updates faction progress and handles exclusivity lockouts.
func (s *ReputationQuestGatingSystem) OnQuestCompleted(entity *Entity, questID string) {
	gatingComp := s.getGatingComponent(entity)
	if gatingComp == nil {
		return
	}

	lockedOut := gatingComp.RecordFactionQuestCompletion(questID)

	if len(lockedOut) > 0 {
		s.logger.WithFields(logrus.Fields{
			"system_name":   "reputation_quest_gating",
			"entity_id":     entity.ID,
			"quest_id":      questID,
			"locked_out":    len(lockedOut),
			"locked_quests": lockedOut,
		}).Info("Quest completion caused lockouts")
	}

	// Apply reputation bonus if configured
	quest := gatingComp.GatedQuests[questID]
	if quest != nil && quest.RewardReputationBonus != 0 {
		repComp := s.getReputationComponent(entity)
		if repComp != nil {
			repComp.AdjustReputation(quest.FactionID, quest.RewardReputationBonus)
			s.logger.WithFields(logrus.Fields{
				"system_name": "reputation_quest_gating",
				"entity_id":   entity.ID,
				"quest_id":    questID,
				"faction_id":  quest.FactionID,
				"bonus":       quest.RewardReputationBonus,
			}).Debug("Applied quest reputation bonus")
		}
	}
}

// GetAvailableQuests returns all quests that the entity can currently accept.
func (s *ReputationQuestGatingSystem) GetAvailableQuests(entity *Entity) []string {
	gatingComp := s.getGatingComponent(entity)
	repComp := s.getReputationComponent(entity)

	if gatingComp == nil || repComp == nil {
		return nil
	}

	var available []string
	for questID := range gatingComp.GatedQuests {
		if gatingComp.IsQuestAvailable(questID, repComp.Factions) {
			available = append(available, questID)
		}
	}
	return available
}

// GetBlockedQuests returns all quests with unmet requirements and their reasons.
func (s *ReputationQuestGatingSystem) GetBlockedQuests(entity *Entity) map[string]string {
	gatingComp := s.getGatingComponent(entity)
	repComp := s.getReputationComponent(entity)

	if gatingComp == nil || repComp == nil {
		return nil
	}

	blocked := make(map[string]string)
	for questID := range gatingComp.GatedQuests {
		if !gatingComp.IsQuestAvailable(questID, repComp.Factions) {
			reason := gatingComp.GetQuestBlockReason(questID, repComp.Factions)
			blocked[questID] = reason
		}
	}
	return blocked
}

// GenerateFactionQuestLine creates a chain of quests with increasing reputation requirements.
// This uses deterministic generation based on the system's seed.
func (s *ReputationQuestGatingSystem) GenerateFactionQuestLine(
	factionID string,
	baseSeed int64,
	questCount int,
) []FactionQuestConfig {
	rng := rand.New(rand.NewSource(baseSeed))

	configs := make([]FactionQuestConfig, questCount)
	tiers := []ReputationTier{TierNeutral, TierFriendly, TierFriendly, TierHonored, TierHonored, TierRevered}

	for i := 0; i < questCount; i++ {
		tierIndex := i
		if tierIndex >= len(tiers) {
			tierIndex = len(tiers) - 1
		}

		configs[i] = FactionQuestConfig{
			QuestID:          factionID + "_quest_" + string(rune('a'+i)),
			MinTier:          tiers[tierIndex],
			FailureMessage:   "You need " + tiers[tierIndex].String() + " standing with " + factionID,
			UnlockMessage:    "New quest available from " + factionID + "!",
			ReputationBonus:  float64(10 + rng.Intn(10)), // 10-19 bonus
			IsExclusive:      i == questCount-1,          // Final quest is exclusive
			ExcludesFactions: nil,                        // Set by caller if needed
		}
	}

	return configs
}

// GetFactionQuestProgress returns the number of quests completed for a faction.
func (s *ReputationQuestGatingSystem) GetFactionQuestProgress(entity *Entity, factionID string) int {
	gatingComp := s.getGatingComponent(entity)
	if gatingComp == nil {
		return 0
	}
	return gatingComp.GetFactionQuestCount(factionID)
}

// GetQuestReputationRequirements returns the requirements for a specific quest.
func (s *ReputationQuestGatingSystem) GetQuestReputationRequirements(entity *Entity, questID string) []QuestReputationRequirement {
	gatingComp := s.getGatingComponent(entity)
	if gatingComp == nil {
		return nil
	}

	quest := gatingComp.GatedQuests[questID]
	if quest == nil {
		return nil
	}

	return quest.Requirements
}
