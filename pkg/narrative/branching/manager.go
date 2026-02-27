package branching

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager handles narrative progression and player choices
type Manager struct {
	mu             sync.RWMutex
	graph          *StoryGraph
	playerProgress map[string]map[string]*PlayerProgress // playerID -> arcID -> progress
	logger         *logrus.Entry
}

// NewManager creates a new narrative manager
func NewManager() *Manager {
	return &Manager{
		graph: &StoryGraph{
			Arcs:         make(map[string]*StoryArc),
			Consequences: make(map[string]*Consequence),
		},
		playerProgress: make(map[string]map[string]*PlayerProgress),
		logger:         logrus.WithField("system_name", "branching_manager"),
	}
}

// SetLogger sets a custom logger for the manager
func (m *Manager) SetLogger(logger *logrus.Entry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if logger != nil {
		m.logger = logger
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
		err := fmt.Errorf("arc %s not found", arcID)
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"error":     err.Error(),
		}).Debug("failed to start arc: not found")
		return nil, err
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
		// INTENTIONAL time.Now() EXCEPTION: StartTime and LastUpdate are metadata
		// for analytics/debugging only. They do NOT affect story generation or
		// choice outcomes. All narrative generation logic is deterministic and
		// seed-based. These timestamps are for observability, not game state.
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Completed:  false,
	}

	// Initialize alignment to neutral
	progress.Alignment[AlignmentGoodEvil] = 0.0
	progress.Alignment[AlignmentLawChaos] = 0.0
	progress.Alignment[AlignmentHonorDishonor] = 0.0

	m.playerProgress[playerID][arcID] = progress

	m.logger.WithFields(logrus.Fields{
		"player_id":     playerID,
		"arc_id":        arcID,
		"start_node_id": arc.StartNodeID,
	}).Debug("player started story arc")

	return progress, nil
}

// MakeChoice processes a player's choice and advances the story
func (m *Manager) MakeChoice(playerID, arcID, choiceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, arc, currentNode, err := m.validateChoiceContext(playerID, arcID)
	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"choice_id": choiceID,
			"error":     err.Error(),
		}).Debug("failed to make choice: invalid context")
		return err
	}

	selectedChoice, err := m.findChoice(currentNode, choiceID)
	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"choice_id": choiceID,
			"node_id":   currentNode.ID,
			"error":     err.Error(),
		}).Debug("failed to make choice: choice not found")
		return err
	}

	if err := m.checkRequirements(progress, selectedChoice.Requirements); err != nil {
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"choice_id": choiceID,
			"error":     err.Error(),
		}).Debug("failed to make choice: requirements not met")
		return fmt.Errorf("requirements not met: %w", err)
	}

	m.applyChoiceEffects(progress, selectedChoice)
	progress.ChoicesMade[progress.CurrentNodeID] = choiceID

	m.logger.WithFields(logrus.Fields{
		"player_id":    playerID,
		"arc_id":       arcID,
		"choice_id":    choiceID,
		"next_node_id": selectedChoice.NextNodeID,
	}).Debug("player made choice")

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

	arc, exists := m.graph.Arcs[arcID]
	if !exists {
		return nil, nil, nil, fmt.Errorf("arc %s not found", arcID)
	}

	currentNode, exists := arc.Nodes[progress.CurrentNodeID]
	if !exists {
		return nil, nil, nil, fmt.Errorf("current node %s not found in arc %s", progress.CurrentNodeID, arcID)
	}

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
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"error":     err.Error(),
		}).Debug("failed to advance story: no progress found")
		return err
	}

	if progress.Completed {
		err := fmt.Errorf("arc %s already completed", arcID)
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"error":     err.Error(),
		}).Debug("failed to advance story: arc completed")
		return err
	}

	arc, exists := m.graph.Arcs[arcID]
	if !exists {
		err := fmt.Errorf("arc %s not found", arcID)
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"error":     err.Error(),
		}).Debug("failed to advance story: arc not found")
		return err
	}

	currentNode, exists := arc.Nodes[progress.CurrentNodeID]
	if !exists {
		err := fmt.Errorf("current node %s not found in arc %s", progress.CurrentNodeID, arcID)
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"node_id":   progress.CurrentNodeID,
			"error":     err.Error(),
		}).Debug("failed to advance story: node not found")
		return err
	}

	if currentNode.Type == NodeTypeChoice {
		err := fmt.Errorf("current node %s is a choice node, use MakeChoice instead", progress.CurrentNodeID)
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"node_id":   progress.CurrentNodeID,
			"error":     err.Error(),
		}).Debug("failed to advance story: is choice node")
		return err
	}

	if currentNode.NextNodeID == "" {
		err := fmt.Errorf("current node %s has no next node", progress.CurrentNodeID)
		m.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"arc_id":    arcID,
			"node_id":   progress.CurrentNodeID,
			"error":     err.Error(),
		}).Debug("failed to advance story: no next node")
		return err
	}

	m.logger.WithFields(logrus.Fields{
		"player_id":    playerID,
		"arc_id":       arcID,
		"from_node_id": progress.CurrentNodeID,
		"to_node_id":   currentNode.NextNodeID,
	}).Debug("player advancing story")

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

	arc, exists := m.graph.Arcs[arcID]
	if !exists {
		return nil, fmt.Errorf("arc %s not found", arcID)
	}

	node, exists := arc.Nodes[progress.CurrentNodeID]
	if !exists {
		return nil, fmt.Errorf("current node %s not found in arc %s", progress.CurrentNodeID, arcID)
	}

	return node, nil
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

	var triggered []string
	for consequenceID, consequence := range m.graph.Consequences {
		if m.isConsequenceAlreadyTriggered(progress, consequenceID) {
			continue
		}

		if m.shouldTriggerConsequence(progress, consequence) {
			m.applyConsequenceEffects(progress, consequence)
			triggered = append(triggered, consequenceID)
			m.trackTriggeredConsequence(progress, consequenceID)
		}
	}

	return triggered
}

// isConsequenceAlreadyTriggered checks if a consequence was already triggered for this progress.
func (m *Manager) isConsequenceAlreadyTriggered(progress *PlayerProgress, consequenceID string) bool {
	triggeredList := toStringSlice(progress.Variables["triggered_consequences"])
	for _, id := range triggeredList {
		if id == consequenceID {
			return true
		}
	}
	return false
}

// shouldTriggerConsequence evaluates if consequence conditions are met.
func (m *Manager) shouldTriggerConsequence(progress *PlayerProgress, consequence *Consequence) bool {
	return m.evaluateConditions(progress, consequence.TriggerConditions)
}

// applyConsequenceEffects applies the consequence effects to player progress.
func (m *Manager) applyConsequenceEffects(progress *PlayerProgress, consequence *Consequence) {
	for key, value := range consequence.Effects {
		progress.Variables[key] = value
	}
}

// trackTriggeredConsequence records that a consequence was triggered.
func (m *Manager) trackTriggeredConsequence(progress *PlayerProgress, consequenceID string) {
	existingList := toStringSlice(progress.Variables["triggered_consequences"])
	if existingList == nil {
		existingList = []string{}
	}
	progress.Variables["triggered_consequences"] = append(existingList, consequenceID)
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
	// INTENTIONAL time.Now() EXCEPTION: LastUpdate is metadata for observability only
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
		if err := m.checkSingleRequirement(progress, key, required); err != nil {
			return err
		}
	}
	return nil
}

// checkSingleRequirement validates a single requirement against player progress.
func (m *Manager) checkSingleRequirement(progress *PlayerProgress, key string, required interface{}) error {
	actual, exists := progress.Variables[key]
	if !exists {
		return fmt.Errorf("requirement %s not met: variable not found", key)
	}

	switch req := required.(type) {
	case int:
		return m.checkIntRequirement(key, req, actual)
	case float64:
		return m.checkFloatRequirement(key, req, actual)
	case bool:
		return m.checkBoolRequirement(key, req, actual)
	case string:
		return m.checkStringRequirement(key, req, actual)
	}
	return nil
}

// checkIntRequirement validates an integer requirement.
func (m *Manager) checkIntRequirement(key string, required int, actual interface{}) error {
	switch act := actual.(type) {
	case int:
		if act < required {
			return fmt.Errorf("requirement %s not met: need %d, have %d", key, required, act)
		}
	case float64:
		if int(act) < required {
			return fmt.Errorf("requirement %s not met: need %d, have %v", key, required, actual)
		}
	default:
		return fmt.Errorf("requirement %s not met: type mismatch", key)
	}
	return nil
}

// checkFloatRequirement validates a float requirement.
func (m *Manager) checkFloatRequirement(key string, required float64, actual interface{}) error {
	switch act := actual.(type) {
	case float64:
		if act < required {
			return fmt.Errorf("requirement %s not met: need %f, have %f", key, required, act)
		}
	case int:
		if float64(act) < required {
			return fmt.Errorf("requirement %s not met: need %f, have %v", key, required, actual)
		}
	default:
		return fmt.Errorf("requirement %s not met: type mismatch", key)
	}
	return nil
}

// checkBoolRequirement validates a boolean requirement.
func (m *Manager) checkBoolRequirement(key string, required bool, actual interface{}) error {
	act, ok := actual.(bool)
	if !ok || act != required {
		return fmt.Errorf("requirement %s not met: need %v, have %v", key, required, actual)
	}
	return nil
}

// checkStringRequirement validates a string requirement.
func (m *Manager) checkStringRequirement(key, required string, actual interface{}) error {
	act, ok := actual.(string)
	if !ok || act != required {
		return fmt.Errorf("requirement %s not met: need %s, have %v", key, required, actual)
	}
	return nil
}

func (m *Manager) evaluateConditions(progress *PlayerProgress, conditions map[string]interface{}) bool {
	for key, expected := range conditions {
		if !m.evaluateSingleCondition(progress, key, expected) {
			return false
		}
	}
	return true
}

// evaluateSingleCondition evaluates a single condition against player progress.
func (m *Manager) evaluateSingleCondition(progress *PlayerProgress, key string, expected interface{}) bool {
	actual, exists := progress.Variables[key]
	if !exists {
		return false
	}

	switch exp := expected.(type) {
	case int:
		return m.evaluateIntCondition(exp, actual)
	case float64:
		return m.evaluateFloatCondition(exp, actual)
	case bool:
		return m.evaluateBoolCondition(exp, actual)
	case string:
		return m.evaluateStringCondition(exp, actual)
	}
	return false
}

// evaluateIntCondition evaluates an integer condition.
func (m *Manager) evaluateIntCondition(expected int, actual interface{}) bool {
	switch act := actual.(type) {
	case int:
		return act == expected
	case float64:
		return int(act) == expected
	}
	return false
}

// evaluateFloatCondition evaluates a float condition.
func (m *Manager) evaluateFloatCondition(expected float64, actual interface{}) bool {
	switch act := actual.(type) {
	case float64:
		return act == expected
	case int:
		return float64(act) == expected
	}
	return false
}

// evaluateBoolCondition evaluates a boolean condition.
func (m *Manager) evaluateBoolCondition(expected bool, actual interface{}) bool {
	act, ok := actual.(bool)
	return ok && act == expected
}

// evaluateStringCondition evaluates a string condition.
func (m *Manager) evaluateStringCondition(expected string, actual interface{}) bool {
	act, ok := actual.(string)
	return ok && act == expected
}

// toStringSlice safely converts interface{} to []string, handling both
// []string (direct assignment) and []interface{} (from JSON deserialization).
func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}

	// Handle direct []string type
	if ss, ok := v.([]string); ok {
		return ss
	}

	// Handle []interface{} from JSON deserialization
	if si, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(si))
		for _, item := range si {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return nil
}
