package narrative_world

import (
	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System wraps StoryEventManager as an ECS system.
type System struct {
	world   *engine.World
	manager *StoryEventManager
	logger  *logrus.Entry
	seed    int64
}

// NewSystem creates a new narrative world system.
func NewSystem(world *engine.World, seed int64) *System {
	logger := logrus.WithField("system", "narrative_world")
	return &System{
		world:   world,
		manager: NewStoryEventManager(seed),
		logger:  logger,
		seed:    seed,
	}
}

// Update processes entities with companion components to track narrative events.
func (s *System) Update(entities []*engine.Entity, deltaTime float64) {
	// Track companion interactions and record memories
	companions := make([]*engine.Entity, 0)
	for _, entity := range entities {
		if entity.HasComponent("companion") {
			companions = append(companions, entity)
		}
	}

	// Check for personality conflicts between active companions
	for i := 0; i < len(companions); i++ {
		for j := i + 1; j < len(companions); j++ {
			s.checkCompanionConflict(companions[i], companions[j])
		}
	}

	// Update quest progress and conflict states
	for _, companion := range companions {
		s.updateCompanionQuests(companion)
	}
}

// checkCompanionConflict detects personality conflicts between companions.
func (s *System) checkCompanionConflict(entity1, entity2 *engine.Entity) {
	comp1, ok1 := entity1.GetComponent("companion")
	comp2, ok2 := entity2.GetComponent("companion")
	if !ok1 || !ok2 {
		return
	}

	companionComp1, ok1 := comp1.(*engine.CompanionComponent)
	companionComp2, ok2 := comp2.(*engine.CompanionComponent)
	if !ok1 || !ok2 {
		return
	}

	// Get personality evolution if available
	var personality1, personality2 *learning.PersonalityEvolution
	if entity1.HasComponent("companion_learning") {
		if comp, ok := entity1.GetComponent("companion_learning"); ok {
			if learningComp, ok := comp.(*learning.CompanionLearningComponent); ok {
				personality1 = learningComp.Personality
			}
		}
	}
	if entity2.HasComponent("companion_learning") {
		if comp, ok := entity2.GetComponent("companion_learning"); ok {
			if learningComp, ok := comp.(*learning.CompanionLearningComponent); ok {
				personality2 = learningComp.Personality
			}
		}
	}

	// Check for conflict
	conflict, exists := s.manager.CheckConflict(
		companionComp1, companionComp2,
		entity1.ID, entity2.ID,
		personality1, personality2,
	)

	if exists {
		s.logger.WithFields(logrus.Fields{
			"companion1":    entity1.ID,
			"companion2":    entity2.ID,
			"conflict_type": conflict.ConflictType.String(),
			"severity":      conflict.Severity,
		}).Debug("Companion conflict detected")

		// Record memory event for both companions
		s.manager.RecordMemory(entity1.ID, EventTypeConflict,
			"Personality clash with companion")
		s.manager.RecordMemory(entity2.ID, EventTypeConflict,
			"Personality clash with companion")
	}
}

// updateCompanionQuests checks if companions are eligible for personal quests.
func (s *System) updateCompanionQuests(entity *engine.Entity) {
	comp, ok := entity.GetComponent("companion")
	if !ok {
		return
	}

	companionComp, ok := comp.(*engine.CompanionComponent)
	if !ok {
		return
	}

	// Check if companion meets loyalty requirement for personal quest
	if companionComp.Loyalty >= 0.7 {
		// Check if companion already has active quests
		activeQuests := s.manager.GetActiveQuests(entity.ID)

		// Generate new quest if fewer than 2 active quests
		if len(activeQuests) < 2 {
			quest, err := s.manager.GeneratePersonalQuest(entity.ID, companionComp, s.seed)
			if err == nil {
				s.logger.WithFields(logrus.Fields{
					"companion_id":   entity.ID,
					"companion_type": companionComp.CompanionType,
					"quest_title":    quest.Title,
					"loyalty":        companionComp.Loyalty,
				}).Info("Generated companion personal quest")

				// Record quest generation as important memory
				s.manager.RecordMemory(entity.ID, EventTypeBonding,
					"Personal quest unlocked: "+quest.Title)
			}
		}
	}
}

// RecordCombatEvent records a combat memory for a companion.
func (s *System) RecordCombatEvent(companionID uint64, description string) {
	s.manager.RecordMemory(companionID, EventTypeCombat, description)
}

// RecordBondingEvent records a bonding memory for a companion.
func (s *System) RecordBondingEvent(companionID uint64, description string) {
	s.manager.RecordMemory(companionID, EventTypeBonding, description)
}

// GetDialogueContext retrieves memory-based dialogue context for a companion.
func (s *System) GetDialogueContext(companionID uint64) *DialogueContext {
	return s.manager.GetDialogueContext(companionID, 10)
}

// GetActiveQuests returns active quests for a companion.
func (s *System) GetActiveQuests(companionID uint64) []*PersonalQuest {
	return s.manager.GetActiveQuests(companionID)
}

// GetActiveConflicts returns all active companion conflicts.
func (s *System) GetActiveConflicts() []CompanionConflict {
	return s.manager.GetActiveConflicts()
}

// CompleteQuest marks a quest as completed and applies consequences.
func (s *System) CompleteQuest(companionID uint64, questID string) error {
	quest, err := s.manager.CompleteQuest(companionID, questID)
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"companion_id": companionID,
		"quest_id":     questID,
		"quest_title":  quest.Title,
	}).Info("Companion quest completed")

	return nil
}

// UpdateQuestObjective updates progress on a quest objective.
func (s *System) UpdateQuestObjective(companionID uint64, questID string, objectiveIndex, progress int) error {
	return s.manager.UpdateQuestObjective(companionID, questID, objectiveIndex, progress)
}

// GetManager returns the underlying StoryEventManager for advanced usage.
func (s *System) GetManager() *StoryEventManager {
	return s.manager
}
