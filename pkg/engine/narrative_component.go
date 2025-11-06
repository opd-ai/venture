// Package engine provides narrative components for dynamic story generation.
// This file defines components for tracking narrative events, story beats,
// and player choices that shape emergent storylines.
package engine

import (
	"fmt"
	"time"
)

// NarrativeEventType represents the category of narrative event.
type NarrativeEventType int

const (
	// EventDiscovery represents finding something important (location, item, information).
	EventDiscovery NarrativeEventType = iota
	// EventConflict represents confrontation with enemies or factions.
	EventConflict
	// EventAlliance represents forming partnerships or truces.
	EventAlliance
	// EventBetrayal represents broken trust or deception.
	EventBetrayal
	// EventRevelation represents learning critical information that changes understanding.
	EventRevelation
	// EventSacrifice represents a difficult choice with permanent consequences.
	EventSacrifice
	// EventVictory represents successfully completing a major challenge.
	EventVictory
	// EventDefeat represents failure with narrative consequences.
	EventDefeat
)

// String returns the string representation of a narrative event type.
func (t NarrativeEventType) String() string {
	switch t {
	case EventDiscovery:
		return "Discovery"
	case EventConflict:
		return "Conflict"
	case EventAlliance:
		return "Alliance"
	case EventBetrayal:
		return "Betrayal"
	case EventRevelation:
		return "Revelation"
	case EventSacrifice:
		return "Sacrifice"
	case EventVictory:
		return "Victory"
	case EventDefeat:
		return "Defeat"
	default:
		return "Unknown"
	}
}

// StoryAct represents the three-act structure position.
type StoryAct int

const (
	// ActSetup is the introduction and establishment (Act 1).
	ActSetup StoryAct = iota
	// ActConfrontation is the main conflict and rising action (Act 2).
	ActConfrontation
	// ActResolution is the climax and conclusion (Act 3).
	ActResolution
)

// String returns the string representation of a story act.
func (a StoryAct) String() string {
	switch a {
	case ActSetup:
		return "Setup"
	case ActConfrontation:
		return "Confrontation"
	case ActResolution:
		return "Resolution"
	default:
		return "Unknown"
	}
}

// NarrativeEvent represents a single story beat that occurred during gameplay.
type NarrativeEvent struct {
	// Type of event (discovery, conflict, alliance, etc.)
	Type NarrativeEventType

	// Timestamp when the event occurred
	Timestamp time.Time

	// Description of what happened (procedurally generated)
	Description string

	// Entity IDs involved in the event
	InvolvedEntities []int

	// Location where event occurred (X, Y coordinates)
	LocationX, LocationY float64

	// Act in which this event occurred
	Act StoryAct

	// Importance score (0.0-1.0, higher = more significant to the story)
	Importance float64

	// Whether this event was player-triggered or world-generated
	PlayerTriggered bool

	// Consequences that resulted from this event
	Consequences []string
}

// NarrativeComponent tracks the narrative state and story events for the world.
// It maintains a history of significant events that shape the emergent story.
type NarrativeComponent struct {
	// Current story act (setup, confrontation, resolution)
	CurrentAct StoryAct

	// All narrative events that have occurred, ordered chronologically
	EventHistory []NarrativeEvent

	// Current main objective description
	MainObjective string

	// Progress towards resolving the main story arc (0.0-1.0)
	StoryProgress float64

	// Active story threads (unresolved plot points)
	ActiveThreads []string

	// Resolved story threads (completed plot points)
	ResolvedThreads []string

	// World state flags that affect narrative (e.g., "defeated_boss_dragon", "saved_village")
	WorldStateFlags map[string]bool

	// Relationship values with factions/NPCs (-1.0 to 1.0, negative = hostile, positive = friendly)
	Relationships map[string]float64

	// Major decisions the player has made
	PlayerDecisions []PlayerDecision

	// Next event trigger conditions
	TriggerConditions []TriggerCondition
}

// PlayerDecision represents a significant choice made by the player.
type PlayerDecision struct {
	// Description of the decision
	Description string

	// When the decision was made
	Timestamp time.Time

	// Possible outcomes that could have happened
	AlternativeOutcomes []string

	// Actual outcome that occurred
	ActualOutcome string

	// Long-term consequences
	Consequences []string
}

// TriggerCondition defines what needs to happen for a narrative event to trigger.
type TriggerCondition struct {
	// Type of trigger (exploration, combat_victory, dialogue_complete, puzzle_solved)
	TriggerType string

	// Required world state flags
	RequiredFlags []string

	// Event that will trigger when conditions are met
	EventToTrigger NarrativeEvent

	// Whether this trigger has been activated
	Activated bool
}

// Type returns the component type identifier.
func (n *NarrativeComponent) Type() string {
	return "narrative"
}

// NewNarrativeComponent creates a new narrative component with initial story state.
func NewNarrativeComponent() *NarrativeComponent {
	return &NarrativeComponent{
		CurrentAct:        ActSetup,
		EventHistory:      make([]NarrativeEvent, 0),
		MainObjective:     "",
		StoryProgress:     0.0,
		ActiveThreads:     make([]string, 0),
		ResolvedThreads:   make([]string, 0),
		WorldStateFlags:   make(map[string]bool),
		Relationships:     make(map[string]float64),
		PlayerDecisions:   make([]PlayerDecision, 0),
		TriggerConditions: make([]TriggerCondition, 0),
	}
}

const (
	// StoryProgressMultiplier controls how much each event contributes to story progress.
	StoryProgressMultiplier = 0.1

	// ActTwoThreshold is the story progress required to advance from Act 1 to Act 2.
	ActTwoThreshold = 0.33

	// ActThreeThreshold is the story progress required to advance from Act 2 to Act 3.
	ActThreeThreshold = 0.66
)

// AddEvent records a new narrative event.
func (n *NarrativeComponent) AddEvent(event NarrativeEvent) {
	event.Timestamp = time.Now()
	event.Act = n.CurrentAct
	n.EventHistory = append(n.EventHistory, event)

	// Update story progress based on event importance
	n.StoryProgress += event.Importance * StoryProgressMultiplier
	if n.StoryProgress > 1.0 {
		n.StoryProgress = 1.0
	}
}

// SetWorldFlag sets a world state flag to true or false.
func (n *NarrativeComponent) SetWorldFlag(flag string, value bool) {
	n.WorldStateFlags[flag] = value
}

// HasWorldFlag checks if a world state flag is set to true.
func (n *NarrativeComponent) HasWorldFlag(flag string) bool {
	return n.WorldStateFlags[flag]
}

// ModifyRelationship adjusts the relationship value with a faction or NPC.
// Delta is added to the current value, clamped to [-1.0, 1.0].
func (n *NarrativeComponent) ModifyRelationship(key string, delta float64) {
	current := n.Relationships[key]
	current += delta
	if current > 1.0 {
		current = 1.0
	} else if current < -1.0 {
		current = -1.0
	}
	n.Relationships[key] = current
}

// GetRelationship returns the relationship value with a faction or NPC.
// Returns 0.0 if no relationship has been established.
func (n *NarrativeComponent) GetRelationship(key string) float64 {
	return n.Relationships[key]
}

// AddDecision records a player decision with its consequences.
func (n *NarrativeComponent) AddDecision(decision PlayerDecision) {
	decision.Timestamp = time.Now()
	n.PlayerDecisions = append(n.PlayerDecisions, decision)
}

// ProgressToNextAct advances the story to the next act if conditions are met.
func (n *NarrativeComponent) ProgressToNextAct() error {
	switch n.CurrentAct {
	case ActSetup:
		if n.StoryProgress >= ActTwoThreshold {
			n.CurrentAct = ActConfrontation
			return nil
		}
		return fmt.Errorf("insufficient progress for Act 2: %.2f < %.2f", n.StoryProgress, ActTwoThreshold)
	case ActConfrontation:
		if n.StoryProgress >= ActThreeThreshold {
			n.CurrentAct = ActResolution
			return nil
		}
		return fmt.Errorf("insufficient progress for Act 3: %.2f < %.2f", n.StoryProgress, ActThreeThreshold)
	case ActResolution:
		return fmt.Errorf("already at final act")
	default:
		return fmt.Errorf("unknown act: %v", n.CurrentAct)
	}
}

// AddStoryThread adds a new unresolved plot point.
func (n *NarrativeComponent) AddStoryThread(thread string) {
	n.ActiveThreads = append(n.ActiveThreads, thread)
}

// ResolveStoryThread marks a plot point as resolved.
func (n *NarrativeComponent) ResolveStoryThread(thread string) {
	// Remove from active threads
	for i, t := range n.ActiveThreads {
		if t == thread {
			n.ActiveThreads = append(n.ActiveThreads[:i], n.ActiveThreads[i+1:]...)
			break
		}
	}
	// Add to resolved threads
	n.ResolvedThreads = append(n.ResolvedThreads, thread)
}

// GetEventsByType returns all events of a specific type.
func (n *NarrativeComponent) GetEventsByType(eventType NarrativeEventType) []NarrativeEvent {
	var events []NarrativeEvent
	for _, event := range n.EventHistory {
		if event.Type == eventType {
			events = append(events, event)
		}
	}
	return events
}

// GetEventsByAct returns all events that occurred in a specific act.
func (n *NarrativeComponent) GetEventsByAct(act StoryAct) []NarrativeEvent {
	var events []NarrativeEvent
	for _, event := range n.EventHistory {
		if event.Act == act {
			events = append(events, event)
		}
	}
	return events
}

// CheckTriggerConditions evaluates all trigger conditions and activates matching ones.
func (n *NarrativeComponent) CheckTriggerConditions() []NarrativeEvent {
	var triggeredEvents []NarrativeEvent

	for i := range n.TriggerConditions {
		condition := &n.TriggerConditions[i]
		if condition.Activated {
			continue
		}

		// Check if all required flags are set
		allFlagsMet := true
		for _, flag := range condition.RequiredFlags {
			if !n.HasWorldFlag(flag) {
				allFlagsMet = false
				break
			}
		}

		if allFlagsMet {
			condition.Activated = true
			triggeredEvents = append(triggeredEvents, condition.EventToTrigger)
		}
	}

	return triggeredEvents
}
