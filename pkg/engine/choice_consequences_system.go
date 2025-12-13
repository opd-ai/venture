package engine

import (
	"github.com/opd-ai/venture/pkg/integration/choice_consequences"
	"github.com/sirupsen/logrus"
)

// ChoiceConsequencesSystem manages persistent choice tracking and consequences.
// It tracks player decisions, NPC relationships, and content availability based on past choices.
type ChoiceConsequencesSystem struct {
	world   *World
	tracker *choice_consequences.ChoiceTracker
	logger  *logrus.Entry
}

// NewChoiceConsequencesSystem creates a new choice consequences system.
func NewChoiceConsequencesSystem(world *World) *ChoiceConsequencesSystem {
	logger := logrus.WithField("system", "choice_consequences")
	return &ChoiceConsequencesSystem{
		world:   world,
		tracker: choice_consequences.NewChoiceTracker(),
		logger:  logger,
	}
}

// Update processes entities with choice tracker components.
func (s *ChoiceConsequencesSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("choice_tracker") {
			continue
		}

		comp, ok := entity.GetComponent("choice_tracker")
		if !ok {
			continue
		}
		choiceComp, ok := comp.(*choice_consequences.ChoiceTrackerComponent)
		if !ok {
			s.logger.WithFields(logrus.Fields{
				"entityID":       entity.ID,
				"component_type": "choice_tracker",
			}).Warn("Component type assertion failed")
			continue
		}
		s.syncComponentState(choiceComp)
	}
}

// syncComponentState synchronizes component state with the tracker.
func (s *ChoiceConsequencesSystem) syncComponentState(comp *choice_consequences.ChoiceTrackerComponent) {
	// Update component from tracker state
	alignment := s.tracker.GetAlignment(comp.PlayerID)
	if alignment != nil {
		comp.Alignment = alignment
	}

	choiceCount := s.tracker.GetChoiceCount(comp.PlayerID)
	if len(comp.ChoiceHistory) != choiceCount {
		s.logger.WithFields(logrus.Fields{
			"player_id":    comp.PlayerID,
			"choice_count": choiceCount,
			"comp_count":   len(comp.ChoiceHistory),
		}).Debug("Choice count mismatch, component may be stale")
	}
}

// RecordChoice records a player choice and applies consequences.
func (s *ChoiceConsequencesSystem) RecordChoice(playerID string, choice *choice_consequences.PlayerChoice) error {
	if err := s.tracker.RecordChoice(playerID, choice); err != nil {
		s.logger.WithError(err).WithField("player_id", playerID).Error("Failed to record choice")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"player_id":    playerID,
		"choice_id":    choice.ChoiceID,
		"story_node":   choice.StoryNodeID,
		"irreversible": choice.Irreversible,
	}).Debug("Recorded player choice")

	return nil
}

// IsContentAvailable checks if content is available to a player based on past choices.
func (s *ChoiceConsequencesSystem) IsContentAvailable(playerID, contentID string) bool {
	return s.tracker.IsContentAvailable(playerID, contentID)
}

// GetNPCAttitude returns the player's attitude with an NPC.
func (s *ChoiceConsequencesSystem) GetNPCAttitude(playerID, npcID string) float64 {
	return s.tracker.GetNPCAttitude(playerID, npcID)
}

// GetAlignment returns the player's current moral alignment.
func (s *ChoiceConsequencesSystem) GetAlignment(playerID string) *choice_consequences.PlayerAlignment {
	return s.tracker.GetAlignment(playerID)
}

// RegisterQuestBranch registers a quest branch with prerequisites.
func (s *ChoiceConsequencesSystem) RegisterQuestBranch(branch *choice_consequences.QuestBranch) error {
	return s.tracker.RegisterQuestBranch(branch)
}

// IsQuestBranchAvailable checks if a quest branch is available to a player.
func (s *ChoiceConsequencesSystem) IsQuestBranchAvailable(playerID, questID, branchID string) bool {
	return s.tracker.IsQuestBranchAvailable(playerID, questID, branchID)
}

// RegisterClassQuest registers a class-specific quest.
func (s *ChoiceConsequencesSystem) RegisterClassQuest(quest *choice_consequences.ClassSpecificQuest) error {
	return s.tracker.RegisterClassQuest(quest)
}

// IsClassQuestAvailable checks if a class-specific quest is available.
func (s *ChoiceConsequencesSystem) IsClassQuestAvailable(playerID, questID, playerClass string, playerLevel int) bool {
	return s.tracker.IsClassQuestAvailable(playerID, questID, playerClass, playerLevel)
}

// RecordCompanionReaction records how a companion reacted to a choice.
func (s *ChoiceConsequencesSystem) RecordCompanionReaction(playerID string, reaction *choice_consequences.CompanionReaction) error {
	return s.tracker.RecordCompanionReaction(playerID, reaction)
}

// GetCompanionReactions returns recent companion reactions for a player.
func (s *ChoiceConsequencesSystem) GetCompanionReactions(playerID, companionID string) []*choice_consequences.CompanionReaction {
	return s.tracker.GetCompanionReactions(playerID, companionID)
}

// Save saves all choice data to a file.
func (s *ChoiceConsequencesSystem) Save(filename string) error {
	return s.tracker.Save(filename)
}

// Load loads choice data from a file.
func (s *ChoiceConsequencesSystem) Load(filename string) error {
	return s.tracker.Load(filename)
}
