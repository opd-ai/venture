package branching

import (
	"fmt"
	"sync"
	"time"
)

// Manager handles narrative progression and player choices
type Manager struct {
	mu             sync.RWMutex
	graph          *StoryGraph
	playerProgress map[string]map[string]*PlayerProgress // playerID -> arcID -> progress
}

// NewManager creates a new narrative manager
func NewManager() *Manager {
	return &Manager{
		graph: &StoryGraph{
			Arcs:         make(map[string]*StoryArc),
			Consequences: make(map[string]*Consequence),
		},
		playerProgress: make(map[string]map[string]*PlayerProgress),
	}
}

// RegisterArc adds a story arc to the graph
func (m *Manager) RegisterArc(arc *StoryArc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.graph.Arcs[arc.ID] = arc
}

// RegisterConsequence adds a consequence to the graph
func (m *Manager) RegisterConsequence(consequence *Consequence) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.graph.Consequences[consequence.ID] = consequence
}

// StartArc begins a story arc for a player
func (m *Manager) StartArc(playerID, arcID string) (*PlayerProgress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	arc, exists := m.graph.Arcs[arcID]
	if !exists {
		return nil, fmt.Errorf("arc %s not found", arcID)
	}

	if _, exists := m.playerProgress[playerID]; !exists {
		m.playerProgress[playerID] = make(map[string]*PlayerProgress)
	}

	progress := &PlayerProgress{
		ArcID:         arcID,
		CurrentNodeID: arc.StartNodeID,
		VisitedNodes:  []string{arc.StartNodeID},
		ChoicesMade:   make(map[string]string),
		Alignment:     make(map[AlignmentAxis]float64),
		Faction:       make(map[string]float64),
		Variables:     make(map[string]interface{}),
		StartTime:     time.Now(),
		LastUpdate:    time.Now(),
		Completed:     false,
	}

	// Initialize alignment to neutral
	progress.Alignment[AlignmentGoodEvil] = 0.0
	progress.Alignment[AlignmentLawChaos] = 0.0
	progress.Alignment[AlignmentHonorDishonor] = 0.0

	m.playerProgress[playerID][arcID] = progress

	return progress, nil
}

// MakeChoice processes a player's choice and advances the story
func (m *Manager) MakeChoice(playerID, arcID, choiceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, arc, currentNode, err := m.validateChoiceContext(playerID, arcID)
	if err != nil {
		return err
	}

	selectedChoice, err := m.findChoice(currentNode, choiceID)
	if err != nil {
		return err
	}

	if err := m.checkRequirements(progress, selectedChoice.Requirements); err != nil {
		return fmt.Errorf("requirements not met: %w", err)
	}

	m.applyChoiceEffects(progress, selectedChoice)
	progress.ChoicesMade[progress.CurrentNodeID] = choiceID

	return m.advanceToNode(progress, arc, selectedChoice.NextNodeID)
}

// validateChoiceContext validates the player progress and current node for making a choice.
func (m *Manager) validateChoiceContext(playerID, arcID string) (*PlayerProgress, *StoryArc, *StoryNode, error) {
	progress, err := m.getProgress(playerID, arcID)
	if err != nil {
		return nil, nil, nil, err
	}

	if progress.Completed {
		return nil, nil, nil, fmt.Errorf("arc %s already completed", arcID)
	}

	arc := m.graph.Arcs[arcID]
	currentNode := arc.Nodes[progress.CurrentNodeID]

	if currentNode.Type != NodeTypeChoice {
		return nil, nil, nil, fmt.Errorf("current node %s is not a choice node", progress.CurrentNodeID)
	}

	return progress, arc, currentNode, nil
}

// findChoice locates a choice in the current node by ID.
func (m *Manager) findChoice(currentNode *StoryNode, choiceID string) (*Choice, error) {
	for i := range currentNode.Choices {
		if currentNode.Choices[i].ID == choiceID {
			return &currentNode.Choices[i], nil
		}
	}
	return nil, fmt.Errorf("choice %s not found in node %s", choiceID, currentNode.ID)
}

// applyChoiceEffects applies alignment shifts and faction changes from a choice.
func (m *Manager) applyChoiceEffects(progress *PlayerProgress, choice *Choice) {
	applyAlignmentShifts(progress, choice.AlignmentShift)
	applyFactionChanges(progress, choice.FactionChange)
}

// applyAlignmentShifts applies and clamps alignment changes.
func applyAlignmentShifts(progress *PlayerProgress, shifts map[AlignmentAxis]float64) {
	for axis, shift := range shifts {
		progress.Alignment[axis] += shift
		if progress.Alignment[axis] > 1.0 {
			progress.Alignment[axis] = 1.0
		} else if progress.Alignment[axis] < -1.0 {
			progress.Alignment[axis] = -1.0
		}
	}
}

// applyFactionChanges applies and clamps faction reputation changes.
func applyFactionChanges(progress *PlayerProgress, changes map[string]float64) {
	for faction, change := range changes {
		if _, exists := progress.Faction[faction]; !exists {
			progress.Faction[faction] = 0.0
		}
		progress.Faction[faction] += change
		if progress.Faction[faction] > 1.0 {
			progress.Faction[faction] = 1.0
		} else if progress.Faction[faction] < -1.0 {
			progress.Faction[faction] = -1.0
		}
	}
}

// AdvanceStory moves to the next node in linear narrative sections
func (m *Manager) AdvanceStory(playerID, arcID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, err := m.getProgress(playerID, arcID)
	if err != nil {
		return err
	}

	if progress.Completed {
		return fmt.Errorf("arc %s already completed", arcID)
	}

	arc := m.graph.Arcs[arcID]
	currentNode := arc.Nodes[progress.CurrentNodeID]

	if currentNode.Type == NodeTypeChoice {
		return fmt.Errorf("current node %s is a choice node, use MakeChoice instead", progress.CurrentNodeID)
	}

	if currentNode.NextNodeID == "" {
		return fmt.Errorf("current node %s has no next node", progress.CurrentNodeID)
	}

	return m.advanceToNode(progress, arc, currentNode.NextNodeID)
}

// GetProgress returns a player's progress in a story arc
func (m *Manager) GetProgress(playerID, arcID string) (*PlayerProgress, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getProgress(playerID, arcID)
}

// GetCurrentNode returns the current node for a player
func (m *Manager) GetCurrentNode(playerID, arcID string) (*StoryNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, err := m.getProgress(playerID, arcID)
	if err != nil {
		return nil, err
	}

	arc := m.graph.Arcs[arcID]
	return arc.Nodes[progress.CurrentNodeID], nil
}

// GetAlignment returns a player's current alignment
func (m *Manager) GetAlignment(playerID, arcID string) (map[AlignmentAxis]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, err := m.getProgress(playerID, arcID)
	if err != nil {
		return nil, err
	}

	// Return a copy
	alignment := make(map[AlignmentAxis]float64)
	for axis, value := range progress.Alignment {
		alignment[axis] = value
	}

	return alignment, nil
}

// GetFactionReputation returns a player's current faction reputation
func (m *Manager) GetFactionReputation(playerID, arcID string) (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, err := m.getProgress(playerID, arcID)
	if err != nil {
		return nil, err
	}

	// Return a copy
	reputation := make(map[string]float64)
	for faction, value := range progress.Faction {
		reputation[faction] = value
	}

	return reputation, nil
}

// CheckConsequences evaluates all consequences and triggers applicable ones
func (m *Manager) CheckConsequences(playerID, arcID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, err := m.getProgress(playerID, arcID)
	if err != nil {
		return nil
	}

	triggered := []string{}

	for consequenceID, consequence := range m.graph.Consequences {
		// Skip already triggered
		alreadyTriggered := false
		for _, id := range progress.Variables["triggered_consequences"].([]string) {
			if id == consequenceID {
				alreadyTriggered = true
				break
			}
		}
		if alreadyTriggered {
			continue
		}

		// Check trigger conditions
		if m.evaluateConditions(progress, consequence.TriggerConditions) {
			// Apply effects
			for key, value := range consequence.Effects {
				progress.Variables[key] = value
			}

			triggered = append(triggered, consequenceID)

			// Track triggered consequence
			if progress.Variables["triggered_consequences"] == nil {
				progress.Variables["triggered_consequences"] = []string{}
			}
			progress.Variables["triggered_consequences"] = append(
				progress.Variables["triggered_consequences"].([]string),
				consequenceID,
			)
		}
	}

	return triggered
}

// GetArc returns a story arc by ID
func (m *Manager) GetArc(arcID string) (*StoryArc, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	arc, exists := m.graph.Arcs[arcID]
	if !exists {
		return nil, fmt.Errorf("arc %s not found", arcID)
	}

	return arc, nil
}

// Helper methods

func (m *Manager) getProgress(playerID, arcID string) (*PlayerProgress, error) {
	playerArcs, exists := m.playerProgress[playerID]
	if !exists {
		return nil, fmt.Errorf("no progress found for player %s", playerID)
	}

	progress, exists := playerArcs[arcID]
	if !exists {
		return nil, fmt.Errorf("no progress found for arc %s", arcID)
	}

	return progress, nil
}

func (m *Manager) advanceToNode(progress *PlayerProgress, arc *StoryArc, nextNodeID string) error {
	nextNode, exists := arc.Nodes[nextNodeID]
	if !exists {
		return fmt.Errorf("next node %s not found", nextNodeID)
	}

	// Check requirements
	if err := m.checkRequirements(progress, nextNode.Requirements); err != nil {
		return fmt.Errorf("cannot advance to node %s: %w", nextNodeID, err)
	}

	// Apply effects
	for key, value := range nextNode.Effects {
		progress.Variables[key] = value
	}

	// Update progress
	progress.CurrentNodeID = nextNodeID
	progress.VisitedNodes = append(progress.VisitedNodes, nextNodeID)
	progress.LastUpdate = time.Now()

	// Check if we reached an ending
	if nextNode.Type == NodeTypeEnding {
		progress.Completed = true
		progress.EndingReached = nextNodeID
	}

	return nil
}

func (m *Manager) checkRequirements(progress *PlayerProgress, requirements map[string]interface{}) error {
	for key, required := range requirements {
		actual, exists := progress.Variables[key]
		if !exists {
			return fmt.Errorf("requirement %s not met: variable not found", key)
		}

		// Type-specific comparison
		switch req := required.(type) {
		case int:
			act, ok := actual.(int)
			if !ok || act < req {
				return fmt.Errorf("requirement %s not met: need %d, have %v", key, req, actual)
			}
		case float64:
			act, ok := actual.(float64)
			if !ok || act < req {
				return fmt.Errorf("requirement %s not met: need %f, have %v", key, req, actual)
			}
		case bool:
			act, ok := actual.(bool)
			if !ok || act != req {
				return fmt.Errorf("requirement %s not met: need %v, have %v", key, req, actual)
			}
		}
	}

	return nil
}

func (m *Manager) evaluateConditions(progress *PlayerProgress, conditions map[string]interface{}) bool {
	for key, expected := range conditions {
		actual, exists := progress.Variables[key]
		if !exists {
			return false
		}

		// Type-specific comparison
		switch exp := expected.(type) {
		case int:
			act, ok := actual.(int)
			if !ok || act != exp {
				return false
			}
		case float64:
			act, ok := actual.(float64)
			if !ok || act != exp {
				return false
			}
		case bool:
			act, ok := actual.(bool)
			if !ok || act != exp {
				return false
			}
		case string:
			act, ok := actual.(string)
			if !ok || act != exp {
				return false
			}
		}
	}

	return true
}
