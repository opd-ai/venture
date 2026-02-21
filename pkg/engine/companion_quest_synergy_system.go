// Package engine provides the companion-quest synergy system.
// This system integrates companion skills with quest objectives, providing
// bonuses to quest progress and rewards based on companion specializations.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/sirupsen/logrus"
)

// CompanionQuestSynergySystem manages the integration between companions and quests.
// It detects when companions provide synergy bonuses to active quests and
// applies objective progress multipliers and reward bonuses.
type CompanionQuestSynergySystem struct {
	world           *World
	logger          *logrus.Entry
	updateInterval  float64
	updateTimer     float64
	synergyDistance float64 // Max distance for companion to provide quest synergy
}

// NewCompanionQuestSynergySystem creates a new companion-quest synergy system.
func NewCompanionQuestSynergySystem(world *World) *CompanionQuestSynergySystem {
	logger := logrus.WithField("system_name", "companion_quest_synergy")
	logger.Debug("Creating companion-quest synergy system")

	return &CompanionQuestSynergySystem{
		world:           world,
		logger:          logger,
		updateInterval:  0.5, // Check synergies twice per second
		updateTimer:     0.0,
		synergyDistance: 300.0, // Companion must be within 300 pixels
	}
}

// Update processes companion-quest synergies for all player entities.
func (s *CompanionQuestSynergySystem) Update(entities []*Entity, deltaTime float64) {
	s.updateTimer += deltaTime
	if s.updateTimer < s.updateInterval {
		return
	}
	s.updateTimer -= s.updateInterval

	for _, entity := range entities {
		s.processEntitySynergies(entity)
	}
}

// processEntitySynergies handles synergy processing for a single player entity.
func (s *CompanionQuestSynergySystem) processEntitySynergies(entity *Entity) {
	// Entity needs quest tracker and synergy components
	questComp, hasQuests := entity.GetComponent("questtracker")
	if !hasQuests {
		return
	}
	questTracker := questComp.(*QuestTrackerComponent)

	// Get or create synergy component
	synergyComp, hasSynergy := entity.GetComponent("companion_quest_synergy")
	var synergy *CompanionQuestSynergyComponent
	if !hasSynergy {
		synergy = NewCompanionQuestSynergyComponent()
		entity.AddComponent(synergy)
	} else {
		synergy = synergyComp.(*CompanionQuestSynergyComponent)
	}

	// Find active companions owned by this entity
	companions := s.findActiveCompanions(entity.ID)
	if len(companions) == 0 {
		return
	}

	// Process each active quest
	for _, trackedQuest := range questTracker.ActiveQuests {
		if trackedQuest.Status != QuestStatusActive {
			continue
		}

		s.processQuestSynergy(entity, synergy, trackedQuest, companions)
	}
}

// findActiveCompanions returns companions owned by the entity that are nearby.
func (s *CompanionQuestSynergySystem) findActiveCompanions(ownerID uint64) []*Entity {
	allCompanions := s.world.GetEntitiesWith("companion", "position")
	result := make([]*Entity, 0)

	// Get owner position for distance check
	owner, ownerExists := s.world.GetEntity(ownerID)
	if !ownerExists || owner == nil {
		return result
	}
	ownerPosComp, hasOwnerPos := owner.GetComponent("position")
	if !hasOwnerPos {
		return result
	}
	ownerPos := ownerPosComp.(*PositionComponent)

	for _, companion := range allCompanions {
		compComp, ok := companion.GetComponent("companion")
		if !ok {
			continue
		}
		companionData := compComp.(*CompanionComponent)

		// Check ownership
		if companionData.OwnerID != ownerID {
			continue
		}

		// Check distance
		compPosComp, hasCompPos := companion.GetComponent("position")
		if !hasCompPos {
			continue
		}
		compPos := compPosComp.(*PositionComponent)

		dx := compPos.X - ownerPos.X
		dy := compPos.Y - ownerPos.Y
		distSq := dx*dx + dy*dy
		if distSq <= s.synergyDistance*s.synergyDistance {
			result = append(result, companion)
		}
	}

	return result
}

// processQuestSynergy handles synergy for a single quest with available companions.
func (s *CompanionQuestSynergySystem) processQuestSynergy(
	player *Entity,
	synergy *CompanionQuestSynergyComponent,
	trackedQuest *TrackedQuest,
	companions []*Entity,
) {
	questID := trackedQuest.Quest.ID
	questType := trackedQuest.Quest.Type

	// Check if synergy already registered for this quest
	if synergy.HasActiveSynergy(questID) {
		return
	}

	// Find best matching companion for this quest type
	var bestCompanion *Entity
	var bestSkill CompanionSkillType
	var bestLoyalty float64
	bestMatch := false

	for _, companion := range companions {
		compComp, _ := companion.GetComponent("companion")
		companionData := compComp.(*CompanionComponent)

		skill := GetCompanionSkillForType(companionData.CompanionType)
		matches := skill.MatchesQuestType(questType)

		// Prefer matching skills, then highest loyalty
		if bestCompanion == nil ||
			(matches && !bestMatch) ||
			(matches == bestMatch && companionData.Loyalty > bestLoyalty) {

			bestCompanion = companion
			bestSkill = skill
			bestLoyalty = companionData.Loyalty
			bestMatch = matches
		}
	}

	if bestCompanion == nil {
		return
	}

	// Register synergy
	synergy.AddSynergy(questID, bestCompanion.ID, bestSkill, bestLoyalty)
	synergy.ApplySkillMatch(questID, bestMatch)

	s.logger.WithFields(logrus.Fields{
		"player_id":    player.ID,
		"quest_id":     questID,
		"quest_type":   questType.String(),
		"companion_id": bestCompanion.ID,
		"skill_type":   bestSkill.String(),
		"skill_match":  bestMatch,
		"obj_bonus":    synergy.GetObjectiveBonus(questID),
		"reward_bonus": synergy.GetRewardBonus(questID),
	}).Info("Companion-quest synergy registered")
}

// ApplyObjectiveProgress applies synergy bonuses to quest objective progress.
// Call this from quest progress update logic to get boosted progress.
func (s *CompanionQuestSynergySystem) ApplyObjectiveProgress(playerID uint64, questID string, baseProgress int) int {
	player, exists := s.world.GetEntity(playerID)
	if !exists || player == nil {
		return baseProgress
	}

	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	if !hasSynergy {
		return baseProgress
	}
	synergy := synergyComp.(*CompanionQuestSynergyComponent)

	bonus := synergy.GetObjectiveBonus(questID)
	boostedProgress := int(float64(baseProgress) * bonus)

	if boostedProgress > baseProgress {
		s.logger.WithFields(logrus.Fields{
			"player_id":        playerID,
			"quest_id":         questID,
			"base_progress":    baseProgress,
			"boosted_progress": boostedProgress,
			"bonus_multiplier": bonus,
		}).Debug("Applied synergy objective bonus")
	}

	return boostedProgress
}

// ApplyRewardBonus applies synergy bonuses to quest rewards.
// Returns boosted XP and gold amounts.
func (s *CompanionQuestSynergySystem) ApplyRewardBonus(playerID uint64, questID string, baseXP, baseGold int) (int, int) {
	player, exists := s.world.GetEntity(playerID)
	if !exists || player == nil {
		return baseXP, baseGold
	}

	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	if !hasSynergy {
		return baseXP, baseGold
	}
	synergy := synergyComp.(*CompanionQuestSynergyComponent)

	bonus := synergy.GetRewardBonus(questID)
	boostedXP := int(float64(baseXP) * bonus)
	boostedGold := int(float64(baseGold) * bonus)

	bonusXP := boostedXP - baseXP
	bonusGold := boostedGold - baseGold

	// Record completion
	synergy.CompleteSynergy(questID, bonusXP, bonusGold)

	s.logger.WithFields(logrus.Fields{
		"player_id":    playerID,
		"quest_id":     questID,
		"base_xp":      baseXP,
		"boosted_xp":   boostedXP,
		"base_gold":    baseGold,
		"boosted_gold": boostedGold,
	}).Info("Applied synergy reward bonus")

	return boostedXP, boostedGold
}

// OnQuestAccepted should be called when a player accepts a new quest.
// This triggers initial synergy evaluation.
func (s *CompanionQuestSynergySystem) OnQuestAccepted(playerID uint64, q *quest.Quest) {
	player, exists := s.world.GetEntity(playerID)
	if !exists || player == nil {
		return
	}

	// Ensure player has synergy component
	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	var synergy *CompanionQuestSynergyComponent
	if !hasSynergy {
		synergy = NewCompanionQuestSynergyComponent()
		player.AddComponent(synergy)
	} else {
		synergy = synergyComp.(*CompanionQuestSynergyComponent)
	}

	// Find companions and evaluate synergy
	companions := s.findActiveCompanions(playerID)
	if len(companions) == 0 {
		return
	}

	// Find best matching companion
	var bestCompanion *Entity
	var bestSkill CompanionSkillType
	var bestLoyalty float64

	for _, companion := range companions {
		compComp, _ := companion.GetComponent("companion")
		companionData := compComp.(*CompanionComponent)

		skill := GetCompanionSkillForType(companionData.CompanionType)
		if bestCompanion == nil || companionData.Loyalty > bestLoyalty {
			bestCompanion = companion
			bestSkill = skill
			bestLoyalty = companionData.Loyalty
		}
	}

	if bestCompanion != nil {
		synergy.AddSynergy(q.ID, bestCompanion.ID, bestSkill, bestLoyalty)
		synergy.ApplySkillMatch(q.ID, bestSkill.MatchesQuestType(q.Type))

		s.logger.WithFields(logrus.Fields{
			"player_id":    playerID,
			"quest_id":     q.ID,
			"quest_name":   q.Name,
			"companion_id": bestCompanion.ID,
			"skill_type":   bestSkill.String(),
		}).Info("Quest synergy established on accept")
	}
}

// OnQuestAbandoned should be called when a player abandons a quest.
func (s *CompanionQuestSynergySystem) OnQuestAbandoned(playerID uint64, questID string) {
	player, exists := s.world.GetEntity(playerID)
	if !exists || player == nil {
		return
	}

	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	if !hasSynergy {
		return
	}
	synergy := synergyComp.(*CompanionQuestSynergyComponent)

	synergy.RemoveSynergy(questID)

	s.logger.WithFields(logrus.Fields{
		"player_id": playerID,
		"quest_id":  questID,
	}).Debug("Quest synergy removed on abandon")
}

// GetSynergyStats returns synergy statistics for a player.
func (s *CompanionQuestSynergySystem) GetSynergyStats(playerID uint64) (totalBonusXP, totalBonusGold, questsWithSynergy int) {
	player, exists := s.world.GetEntity(playerID)
	if !exists || player == nil {
		return 0, 0, 0
	}

	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	if !hasSynergy {
		return 0, 0, 0
	}
	synergy := synergyComp.(*CompanionQuestSynergyComponent)

	return synergy.TotalBonusXP, synergy.TotalBonusGold, synergy.QuestsCompletedWithSynergy
}

// GetActiveSynergyBonus returns the current bonus for an active quest.
func (s *CompanionQuestSynergySystem) GetActiveSynergyBonus(playerID uint64, questID string) (objectiveBonus, rewardBonus float64) {
	player, exists := s.world.GetEntity(playerID)
	if !exists || player == nil {
		return 1.0, 1.0
	}

	synergyComp, hasSynergy := player.GetComponent("companion_quest_synergy")
	if !hasSynergy {
		return 1.0, 1.0
	}
	synergy := synergyComp.(*CompanionQuestSynergyComponent)

	return synergy.GetObjectiveBonus(questID), synergy.GetRewardBonus(questID)
}
