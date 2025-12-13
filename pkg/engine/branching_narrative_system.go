// Package engine provides the branching narrative system for story progression.
package engine

import (
	"fmt"
	"strconv"

	"github.com/opd-ai/venture/pkg/narrative/branching"
	"github.com/sirupsen/logrus"
)

// BranchingNarrativeSystem manages branching story arcs and player choices.
// It processes narrative progression, presents choices to players, and applies consequences.
type BranchingNarrativeSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewBranchingNarrativeSystem creates a new branching narrative system.
func NewBranchingNarrativeSystem(world *World) *BranchingNarrativeSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "branching_narrative")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Branching narrative system created")
		}
	}
	return &BranchingNarrativeSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes branching narrative components and checks for narrative progression.
func (bns *BranchingNarrativeSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("branching_narrative")
		if !ok {
			continue
		}

		narComp, ok := comp.(*BranchingNarrativeComponent)
		if !ok {
			continue
		}

		narComp.LastUpdate += deltaTime

		// Update every second to avoid excessive processing
		if narComp.LastUpdate < 1.0 {
			continue
		}
		narComp.LastUpdate = 0

		// Check if we need to update pending choices
		if narComp.Manager != nil && narComp.ArcID != "" {
			playerID := strconv.FormatUint(entity.ID, 10)
			currentNode, err := narComp.Manager.GetCurrentNode(playerID, narComp.ArcID)
			if err == nil && currentNode != nil && currentNode.Type == branching.NodeTypeChoice {
				// Update pending choices
				narComp.PendingChoices = currentNode.Choices
			} else {
				narComp.PendingChoices = nil
			}
		}
	}
}

// StartStoryArc initializes a new story arc for an entity.
// Returns error if entity doesn't have branching narrative component.
func (bns *BranchingNarrativeSystem) StartStoryArc(entity *Entity, arc *branching.StoryArc, manager *branching.Manager) error {
	comp, ok := entity.GetComponent("branching_narrative")
	if !ok {
		return fmt.Errorf("entity does not have branching narrative component")
	}

	narComp, ok := comp.(*BranchingNarrativeComponent)
	if !ok {
		return fmt.Errorf("invalid branching narrative component type")
	}

	// Register the arc with the manager
	manager.RegisterArc(arc)

	// Start the arc
	playerID := strconv.FormatUint(entity.ID, 10)
	progress, err := manager.StartArc(playerID, arc.ID)
	if err != nil {
		return fmt.Errorf("failed to start arc: %w", err)
	}

	narComp.ArcID = arc.ID
	narComp.Progress = progress
	narComp.ActiveArc = arc
	narComp.Manager = manager

	// Set initial pending choices if start node is a choice
	currentNode, err := manager.GetCurrentNode(playerID, arc.ID)
	if err == nil && currentNode != nil && currentNode.Type == branching.NodeTypeChoice {
		narComp.PendingChoices = currentNode.Choices
	}

	if bns.logger != nil && bns.logger.Logger.GetLevel() >= logrus.InfoLevel {
		bns.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"arc_id":    arc.ID,
			"arc_title": arc.Title,
		}).Info("Started story arc")
	}

	return nil
}

// MakeChoice processes a player choice in the narrative.
// Returns error if choice is invalid or requirements aren't met.
func (bns *BranchingNarrativeSystem) MakeChoice(entity *Entity, choiceID string) error {
	comp, ok := entity.GetComponent("branching_narrative")
	if !ok {
		return fmt.Errorf("entity does not have branching narrative component")
	}

	narComp, ok := comp.(*BranchingNarrativeComponent)
	if !ok {
		return fmt.Errorf("invalid branching narrative component type")
	}

	if narComp.Manager == nil || narComp.ArcID == "" {
		return fmt.Errorf("no active story arc")
	}

	// Find the choice
	var selectedChoice *branching.Choice
	for i := range narComp.PendingChoices {
		if narComp.PendingChoices[i].ID == choiceID {
			selectedChoice = &narComp.PendingChoices[i]
			break
		}
	}

	if selectedChoice == nil {
		return fmt.Errorf("choice not found: %s", choiceID)
	}

	// Make the choice
	playerID := strconv.FormatUint(entity.ID, 10)
	err := narComp.Manager.MakeChoice(playerID, narComp.ArcID, choiceID)
	if err != nil {
		return fmt.Errorf("failed to make choice: %w", err)
	}

	if bns.logger != nil && bns.logger.Logger.GetLevel() >= logrus.InfoLevel {
		bns.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"choice_id": choiceID,
			"choice":    selectedChoice.Text,
		}).Info("Player made narrative choice")
	}

	// Clear pending choices and wait for next update to refresh
	narComp.PendingChoices = nil

	return nil
}

// GetCurrentNode returns the current narrative node for an entity.
func (bns *BranchingNarrativeSystem) GetCurrentNode(entity *Entity) (*branching.StoryNode, error) {
	comp, ok := entity.GetComponent("branching_narrative")
	if !ok {
		return nil, fmt.Errorf("entity does not have branching narrative component")
	}

	narComp, ok := comp.(*BranchingNarrativeComponent)
	if !ok {
		return nil, fmt.Errorf("invalid branching narrative component type")
	}

	if narComp.Manager == nil || narComp.ArcID == "" {
		return nil, fmt.Errorf("no active story arc")
	}

	playerID := strconv.FormatUint(entity.ID, 10)
	return narComp.Manager.GetCurrentNode(playerID, narComp.ArcID)
}

// GetAlignment returns the player's current alignment values.
func (bns *BranchingNarrativeSystem) GetAlignment(entity *Entity) (map[branching.AlignmentAxis]float64, error) {
	comp, ok := entity.GetComponent("branching_narrative")
	if !ok {
		return nil, fmt.Errorf("entity does not have branching narrative component")
	}

	narComp, ok := comp.(*BranchingNarrativeComponent)
	if !ok {
		return nil, fmt.Errorf("invalid branching narrative component type")
	}

	if narComp.Manager == nil || narComp.ArcID == "" {
		return nil, fmt.Errorf("no narrative manager")
	}

	playerID := strconv.FormatUint(entity.ID, 10)
	return narComp.Manager.GetAlignment(playerID, narComp.ArcID)
}

// GetFactionReputation returns the player's faction reputation.
func (bns *BranchingNarrativeSystem) GetFactionReputation(entity *Entity) (map[string]float64, error) {
	comp, ok := entity.GetComponent("branching_narrative")
	if !ok {
		return nil, fmt.Errorf("entity does not have branching narrative component")
	}

	narComp, ok := comp.(*BranchingNarrativeComponent)
	if !ok {
		return nil, fmt.Errorf("invalid branching narrative component type")
	}

	if narComp.Manager == nil || narComp.ArcID == "" {
		return nil, fmt.Errorf("no narrative manager")
	}

	playerID := strconv.FormatUint(entity.ID, 10)
	return narComp.Manager.GetFactionReputation(playerID, narComp.ArcID)
}
