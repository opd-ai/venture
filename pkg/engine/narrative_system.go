// Package engine provides the narrative system for dynamic story progression.
// This file implements NarrativeSystem which tracks events, processes triggers,
// and manages story arc advancement based on player actions and world state.
package engine

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// BossDetectionRange is the detection range threshold that indicates a boss entity.
	BossDetectionRange = 300.0
)

// NarrativeSystem manages narrative progression and event tracking.
// It monitors gameplay events, checks trigger conditions, and advances story arcs.
type NarrativeSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewNarrativeSystem creates a new narrative system.
func NewNarrativeSystem(world *World) *NarrativeSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "narrative")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Narrative system created")
		}
	}
	return &NarrativeSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes narrative components and checks trigger conditions.
func (ns *NarrativeSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity has narrative component (typically world entity)
		narComp, ok := entity.GetComponent("narrative")
		if !ok {
			continue
		}

		narrative := narComp.(*NarrativeComponent)

		// Check for triggered events
		triggeredEvents := narrative.CheckTriggerConditions()
		for _, event := range triggeredEvents {
			narrative.AddEvent(event)

			if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.InfoLevel {
				ns.logger.WithFields(logrus.Fields{
					"event_type":     event.Type.String(),
					"description":    event.Description,
					"importance":     event.Importance,
					"story_progress": narrative.StoryProgress,
				}).Info("Narrative event triggered")
			}
		}

		// Check for act progression
		if ns.shouldProgressAct(narrative) {
			if err := narrative.ProgressToNextAct(); err == nil {
				if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.InfoLevel {
					ns.logger.WithFields(logrus.Fields{
						"new_act":        narrative.CurrentAct.String(),
						"progress":       narrative.StoryProgress,
						"active_threads": len(narrative.ActiveThreads),
					}).Info("Story progressed to next act")
				}
			}
		}
	}
}

// shouldProgressAct determines if the story should advance to the next act.
func (ns *NarrativeSystem) shouldProgressAct(narrative *NarrativeComponent) bool {
	// Check if enough progress has been made
	switch narrative.CurrentAct {
	case ActSetup:
		return narrative.StoryProgress >= ActTwoThreshold
	case ActConfrontation:
		return narrative.StoryProgress >= ActThreeThreshold
	case ActResolution:
		return false // Already at final act
	default:
		return false
	}
}

// TriggerEvent manually triggers a narrative event (called by other systems).
func (ns *NarrativeSystem) TriggerEvent(narrative *NarrativeComponent, eventType NarrativeEventType, description string, importance float64, involvedEntities []int, locationX, locationY float64) {
	event := NarrativeEvent{
		Type:             eventType,
		Description:      description,
		InvolvedEntities: involvedEntities,
		LocationX:        locationX,
		LocationY:        locationY,
		Importance:       importance,
		PlayerTriggered:  true,
	}

	narrative.AddEvent(event)

	if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.InfoLevel {
		ns.logger.WithFields(logrus.Fields{
			"event_type":     eventType.String(),
			"description":    description,
			"importance":     importance,
			"story_progress": narrative.StoryProgress,
		}).Info("Narrative event manually triggered")
	}
}

// OnCombatVictory is called when the player wins a combat encounter.
func (ns *NarrativeSystem) OnCombatVictory(narrative *NarrativeComponent, enemyEntity *Entity, locationX, locationY float64) {
	// Check if this was a boss fight (high importance)
	importance := 0.3 // Default importance

	if enemyEntity != nil {
		// Check for boss or elite status
		if aiComp, ok := enemyEntity.GetComponent("ai"); ok {
			ai := aiComp.(*AIComponent)
			// Boss entities typically have high detection range
			if ai.DetectionRange > BossDetectionRange {
				importance = 0.8 // Boss fight
			}
		}
	}

	description := "Defeated enemy in combat"
	if importance > 0.5 {
		description = "Defeated powerful enemy"
	}

	ns.TriggerEvent(narrative, EventVictory, description, importance, []int{}, locationX, locationY)
}

// OnDiscovery is called when the player discovers something significant.
func (ns *NarrativeSystem) OnDiscovery(narrative *NarrativeComponent, discoveryType string, locationX, locationY float64) {
	description := fmt.Sprintf("Discovered %s", discoveryType)
	importance := 0.4 // Discovery importance

	ns.TriggerEvent(narrative, EventDiscovery, description, importance, []int{}, locationX, locationY)
}

// OnDialogueComplete is called when a dialogue interaction finishes.
func (ns *NarrativeSystem) OnDialogueComplete(narrative *NarrativeComponent, npcEntity *Entity, choiceIndex int) {
	description := "Completed dialogue with NPC"
	importance := 0.2 // Basic dialogue importance

	entityID := -1
	if npcEntity != nil {
		entityID = int(npcEntity.ID)
	}

	ns.TriggerEvent(narrative, EventAlliance, description, importance, []int{entityID}, 0, 0)
}

// OnPuzzleSolved is called when the player solves a puzzle.
func (ns *NarrativeSystem) OnPuzzleSolved(narrative *NarrativeComponent, puzzleType string, locationX, locationY float64) {
	description := fmt.Sprintf("Solved %s puzzle", puzzleType)
	importance := 0.5 // Puzzle importance

	ns.TriggerEvent(narrative, EventDiscovery, description, importance, []int{}, locationX, locationY)
}

// OnPlayerDeath is called when the player dies.
func (ns *NarrativeSystem) OnPlayerDeath(narrative *NarrativeComponent, locationX, locationY float64) {
	description := "Player was defeated"
	importance := 0.6 // Death is significant

	ns.TriggerEvent(narrative, EventDefeat, description, importance, []int{}, locationX, locationY)
}

// OnQuestComplete is called when the player completes a quest.
func (ns *NarrativeSystem) OnQuestComplete(narrative *NarrativeComponent, questTitle string) {
	description := fmt.Sprintf("Completed quest: %s", questTitle)
	importance := 0.7 // Quest completion is important

	ns.TriggerEvent(narrative, EventVictory, description, importance, []int{}, 0, 0)
}

// GetStoryStatus returns a summary of the current narrative state.
func (ns *NarrativeSystem) GetStoryStatus(narrative *NarrativeComponent) StoryStatus {
	return StoryStatus{
		CurrentAct:    narrative.CurrentAct,
		Progress:      narrative.StoryProgress,
		MainObjective: narrative.MainObjective,
		ActiveThreads: len(narrative.ActiveThreads),
		EventCount:    len(narrative.EventHistory),
		RecentEvents:  ns.getRecentEvents(narrative, 5),
	}
}

// StoryStatus provides a summary of narrative state for UI display.
type StoryStatus struct {
	CurrentAct    StoryAct
	Progress      float64
	MainObjective string
	ActiveThreads int
	EventCount    int
	RecentEvents  []NarrativeEvent
}

// getRecentEvents returns the most recent events up to the specified count.
func (ns *NarrativeSystem) getRecentEvents(narrative *NarrativeComponent, count int) []NarrativeEvent {
	if len(narrative.EventHistory) == 0 {
		return []NarrativeEvent{}
	}

	start := len(narrative.EventHistory) - count
	if start < 0 {
		start = 0
	}

	return narrative.EventHistory[start:]
}

// SetMainObjective updates the current main objective text.
func (ns *NarrativeSystem) SetMainObjective(narrative *NarrativeComponent, objective string) {
	narrative.MainObjective = objective

	if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.InfoLevel {
		ns.logger.WithField("objective", objective).Info("Main objective updated")
	}
}

// AddStoryTrigger adds a new trigger condition for a narrative event.
func (ns *NarrativeSystem) AddStoryTrigger(narrative *NarrativeComponent, triggerType string, requiredFlags []string, eventToTrigger NarrativeEvent) {
	trigger := TriggerCondition{
		TriggerType:    triggerType,
		RequiredFlags:  requiredFlags,
		EventToTrigger: eventToTrigger,
		Activated:      false,
	}

	narrative.TriggerConditions = append(narrative.TriggerConditions, trigger)

	if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.DebugLevel {
		ns.logger.WithFields(logrus.Fields{
			"trigger_type": triggerType,
			"flags":        requiredFlags,
			"event_type":   eventToTrigger.Type.String(),
		}).Debug("Story trigger added")
	}
}

// CheckWorldFlag is a helper to check if a world state flag is set.
func (ns *NarrativeSystem) CheckWorldFlag(narrative *NarrativeComponent, flag string) bool {
	return narrative.HasWorldFlag(flag)
}

// SetWorldFlag is a helper to set a world state flag.
func (ns *NarrativeSystem) SetWorldFlag(narrative *NarrativeComponent, flag string, value bool) {
	narrative.SetWorldFlag(flag, value)

	if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.DebugLevel {
		ns.logger.WithFields(logrus.Fields{
			"flag":  flag,
			"value": value,
		}).Debug("World flag updated")
	}
}

// GetRelationshipStatus returns the relationship level with a faction or NPC.
func (ns *NarrativeSystem) GetRelationshipStatus(narrative *NarrativeComponent, key string) string {
	value := narrative.GetRelationship(key)

	if value >= 0.75 {
		return "Allied"
	} else if value >= 0.25 {
		return "Friendly"
	} else if value >= -0.25 {
		return "Neutral"
	} else if value > -0.75 {
		return "Hostile"
	} else {
		return "Enemy"
	}
}

// RecordPlayerDecision records a significant player choice.
func (ns *NarrativeSystem) RecordPlayerDecision(narrative *NarrativeComponent, description, chosenOption string, alternatives, consequences []string) {
	decision := PlayerDecision{
		Description:         description,
		Timestamp:           time.Now(),
		AlternativeOutcomes: alternatives,
		ActualOutcome:       chosenOption,
		Consequences:        consequences,
	}

	narrative.AddDecision(decision)

	if ns.logger != nil && ns.logger.Logger.GetLevel() >= logrus.InfoLevel {
		ns.logger.WithFields(logrus.Fields{
			"description": description,
			"choice":      chosenOption,
		}).Info("Player decision recorded")
	}
}
