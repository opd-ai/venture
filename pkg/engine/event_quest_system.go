// Package engine provides the event quest system for managing seasonal event quests.
// EventQuestSystem handles the lifecycle of event-specific quests, including
// generation, expiration, and integration with the Living World NPCs.
package engine

import (
	"time"

	"github.com/sirupsen/logrus"
)

// EventQuestSystem manages event-specific quests for all player entities.
// It processes the world's SeasonalEventComponent to generate, manage,
// and expire quests based on active events.
type EventQuestSystem struct {
	world *World
	clock GameClock
}

// NewEventQuestSystem creates a new event quest system.
func NewEventQuestSystem(world *World, clock GameClock) *EventQuestSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "event_quest",
	}).Debug("Creating event quest system")

	return &EventQuestSystem{
		world: world,
		clock: clock,
	}
}

// Update processes all entities with event quest components.
// It checks for active events, generates quests, and expires completed events.
func (s *EventQuestSystem) Update(entities []*Entity, deltaTime float64) {
	if s.clock == nil {
		return
	}

	currentTime := s.clock.Now()

	// Get the world entity with seasonal events
	worldEntity := s.getWorldEntity(entities)
	if worldEntity == nil {
		return
	}

	eventComp := s.getSeasonalEventComponent(worldEntity)
	if eventComp == nil {
		return
	}

	// Get active events
	activeEvents := eventComp.GetActiveEvents()

	// Process each player entity with event quest component
	for _, entity := range entities {
		if !entity.HasComponent("event_quest") {
			continue
		}

		comp, ok := entity.GetComponent("event_quest")
		if !ok || comp == nil {
			continue
		}
		questComp := comp.(*EventQuestComponent)

		// Generate quests for new active events
		s.generateQuestsForActiveEvents(questComp, activeEvents, eventComp.CalendarSeed)

		// Check for expired quests
		s.expireTimedOutQuests(questComp, currentTime)

		// Clean up quests for ended events
		s.cleanupEndedEventQuests(questComp, activeEvents)

		// Check for quest completion
		s.checkQuestCompletions(questComp)
	}
}

// getWorldEntity finds the world entity in the provided entity list.
func (s *EventQuestSystem) getWorldEntity(entities []*Entity) *Entity {
	for _, entity := range entities {
		if entity.HasComponent("seasonal_event") {
			return entity
		}
	}
	return nil
}

// getSeasonalEventComponent extracts the seasonal event component from an entity.
func (s *EventQuestSystem) getSeasonalEventComponent(entity *Entity) *SeasonalEventComponent {
	comp, ok := entity.GetComponent("seasonal_event")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*SeasonalEventComponent)
}

// generateQuestsForActiveEvents creates quests for newly active events.
func (s *EventQuestSystem) generateQuestsForActiveEvents(
	questComp *EventQuestComponent,
	activeEvents []EventInstance,
	calendarSeed int64,
) {
	for _, event := range activeEvents {
		// Check if we already generated quests for this event
		if questComp.LastGenerationEventID == event.Definition.ID {
			continue
		}

		// Check if there are already quests for this event
		if len(questComp.GetAvailableQuestsForEvent(event.Definition.ID)) > 0 {
			continue
		}
		if len(questComp.GetActiveQuestsForEvent(event.Definition.ID)) > 0 {
			continue
		}

		// Generate new quests for this event
		seed := calendarSeed + event.Definition.Seed
		quests := GenerateEventQuests(event, seed)

		// Add to available quests
		questComp.AvailableQuests = append(questComp.AvailableQuests, quests...)

		questComp.LastGenerationEventID = event.Definition.ID

		logrus.WithFields(logrus.Fields{
			"system_name":  "event_quest",
			"event_id":     event.Definition.ID,
			"event_name":   event.Definition.Name,
			"quests_added": len(quests),
		}).Info("Generated quests for active event")
	}
}

// expireTimedOutQuests moves quests past their expiration to the expired list.
func (s *EventQuestSystem) expireTimedOutQuests(
	questComp *EventQuestComponent,
	currentTime time.Time,
) {
	for i := len(questComp.ActiveQuests) - 1; i >= 0; i-- {
		quest := questComp.ActiveQuests[i]
		if currentTime.After(quest.ExpiresAt) {
			questComp.ExpireQuest(quest.Definition.ID)

			logrus.WithFields(logrus.Fields{
				"system_name": "event_quest",
				"quest_id":    quest.Definition.ID,
				"event_id":    quest.Definition.EventID,
			}).Info("Event quest expired due to timeout")
		}
	}
}

// cleanupEndedEventQuests removes quests for events that are no longer active.
func (s *EventQuestSystem) cleanupEndedEventQuests(
	questComp *EventQuestComponent,
	activeEvents []EventInstance,
) {
	// Build set of active event IDs
	activeEventIDs := make(map[string]bool)
	for _, event := range activeEvents {
		activeEventIDs[event.Definition.ID] = true
	}

	// Find event IDs with active/available quests but no longer active
	eventsToClear := make(map[string]bool)

	for _, quest := range questComp.ActiveQuests {
		if !activeEventIDs[quest.Definition.EventID] {
			eventsToClear[quest.Definition.EventID] = true
		}
	}

	for _, quest := range questComp.AvailableQuests {
		if !activeEventIDs[quest.EventID] {
			eventsToClear[quest.EventID] = true
		}
	}

	// Clear quests for ended events
	for eventID := range eventsToClear {
		questComp.ClearEventQuests(eventID)
	}
}

// checkQuestCompletions checks active quests for completion.
func (s *EventQuestSystem) checkQuestCompletions(questComp *EventQuestComponent) {
	for i := len(questComp.ActiveQuests) - 1; i >= 0; i-- {
		quest := questComp.ActiveQuests[i]

		// Check if all objectives are complete
		complete := true
		for j, obj := range quest.Definition.Objectives {
			if quest.Progress[j] < obj.Required {
				complete = false
				break
			}
		}

		if complete {
			questComp.CompleteQuest(quest.Definition.ID)
		}
	}
}

// GetEventNPCDialogOptions returns dialog options for event NPCs.
// Used to integrate event quests with the Living World NPC schedules.
func (s *EventQuestSystem) GetEventNPCDialogOptions(
	playerEntity *Entity,
	npcEntity *Entity,
) []DialogOption {
	options := make([]DialogOption, 0)

	// Check if player has event quest component
	comp, ok := playerEntity.GetComponent("event_quest")
	if !ok || comp == nil {
		return options
	}
	questComp := comp.(*EventQuestComponent)

	// Get available quests
	for _, q := range questComp.AvailableQuests {
		options = append(options, DialogOption{
			Text:    "Tell me about: " + q.Name,
			Action:  ActionOfferEventQuest,
			Enabled: true,
			Payload: q.ID,
		})
	}

	// Get active quests (for turn-in)
	for _, q := range questComp.ActiveQuests {
		if questComp.IsQuestComplete(q.Definition.ID) {
			options = append(options, DialogOption{
				Text:    "I've completed: " + q.Definition.Name,
				Action:  ActionCompleteEventQuest,
				Enabled: true,
				Payload: q.Definition.ID,
			})
		}
	}

	return options
}

// AcceptEventQuest handles a player accepting an event quest from dialog.
func (s *EventQuestSystem) AcceptEventQuest(
	playerEntity *Entity,
	worldEntity *Entity,
	questID string,
) bool {
	// Get player's event quest component
	comp, ok := playerEntity.GetComponent("event_quest")
	if !ok || comp == nil {
		return false
	}
	questComp := comp.(*EventQuestComponent)

	// Get the event end time
	eventComp := s.getSeasonalEventComponent(worldEntity)
	if eventComp == nil {
		return false
	}

	// Find the event for this quest
	var expiresAt time.Time
	for _, q := range questComp.AvailableQuests {
		if q.ID == questID {
			event := eventComp.GetEventByID(q.EventID)
			if event != nil {
				expiresAt = event.EndTime
			} else {
				// Default to 7 days if event not found
				expiresAt = s.clock.Now().AddDate(0, 0, 7)
			}
			break
		}
	}

	return questComp.AcceptQuest(questID, expiresAt)
}

// TurnInEventQuest handles a player turning in a completed event quest.
func (s *EventQuestSystem) TurnInEventQuest(
	playerEntity *Entity,
	questID string,
) (bool, *EventQuestDefinition) {
	// Get player's event quest component
	comp, ok := playerEntity.GetComponent("event_quest")
	if !ok || comp == nil {
		return false, nil
	}
	questComp := comp.(*EventQuestComponent)

	// Get the quest before completing it
	activeQuest := questComp.GetActiveQuest(questID)
	if activeQuest == nil {
		return false, nil
	}

	// Check if complete
	if !questComp.IsQuestComplete(questID) {
		return false, nil
	}

	// Make a copy of the definition for reward processing
	defCopy := activeQuest.Definition

	// Complete the quest
	if !questComp.CompleteQuest(questID) {
		return false, nil
	}

	return true, &defCopy
}

// GetEventQuestProgress returns the progress summary for an event quest.
func (s *EventQuestSystem) GetEventQuestProgress(
	playerEntity *Entity,
	questID string,
) (current, required int, complete bool) {
	comp, ok := playerEntity.GetComponent("event_quest")
	if !ok || comp == nil {
		return 0, 0, false
	}
	questComp := comp.(*EventQuestComponent)

	quest := questComp.GetActiveQuest(questID)
	if quest == nil {
		return 0, 0, false
	}

	totalCurrent := 0
	totalRequired := 0
	for i, obj := range quest.Definition.Objectives {
		totalCurrent += quest.Progress[i]
		totalRequired += obj.Required
	}

	return totalCurrent, totalRequired, totalCurrent >= totalRequired
}
