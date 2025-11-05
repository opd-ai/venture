// Package engine provides the context action system for interactive objects.
// This file implements ContextActionSystem which processes proximity-based
// context actions and displays prompts to the player.
//
// Phase 11.3: Context Actions System
// Manages context-sensitive interactions with world objects
package engine

import (
	"math"
	"sort"

	"github.com/sirupsen/logrus"
)

// ContextActionSystem manages context-sensitive actions and player interactions.
// It finds nearby interactive objects, determines which action to show,
// and triggers actions when the player presses the interaction key.
type ContextActionSystem struct {
	world  *World
	logger *logrus.Entry

	// nearbyActions caches actions near the player (updated each frame)
	nearbyActions []*nearbyAction

	// activeAction is the currently highlighted action (closest/highest priority)
	activeAction *nearbyAction
}

// nearbyAction combines an entity with its context action for sorting.
type nearbyAction struct {
	entity   *Entity
	action   *ContextActionComponent
	distance float64
	priority int // Higher priority = shown first if equidistant
}

// NewContextActionSystem creates a new context action system.
func NewContextActionSystem(world *World) *ContextActionSystem {
	return NewContextActionSystemWithLogger(world, nil)
}

// NewContextActionSystemWithLogger creates a new context action system with a logger.
func NewContextActionSystemWithLogger(world *World, logger *logrus.Logger) *ContextActionSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "context_action")
	}

	return &ContextActionSystem{
		world:         world,
		logger:        logEntry,
		nearbyActions: make([]*nearbyAction, 0, 16),
	}
}

// Update finds nearby interactive objects and determines active action.
func (cas *ContextActionSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear previous frame's nearby actions
	cas.nearbyActions = cas.nearbyActions[:0]
	cas.activeAction = nil

	// Find player entity
	var player *Entity
	var playerPos *PositionComponent
	for _, entity := range entities {
		if entity.HasTag("player") {
			player = entity
			if posComp, ok := entity.GetComponent("position"); ok {
				playerPos = posComp.(*PositionComponent)
			}
			break
		}
	}

	if player == nil || playerPos == nil {
		return
	}

	// Find all entities with context actions near the player
	for _, entity := range entities {
		if entity == player {
			continue
		}

		// Check for context action component
		actionComp, ok := entity.GetComponent("contextaction")
		if !ok {
			continue
		}

		action := actionComp.(*ContextActionComponent)
		if !action.Enabled {
			continue
		}

		// Check for position component
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos := posComp.(*PositionComponent)

		// Calculate distance to player
		dx := pos.X - playerPos.X
		dy := pos.Y - playerPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Check if within interaction range
		if distance <= action.InteractionRange {
			cas.nearbyActions = append(cas.nearbyActions, &nearbyAction{
				entity:   entity,
				action:   action,
				distance: distance,
				priority: cas.getActionPriority(action.ActionType),
			})
		}
	}

	// If no nearby actions, nothing to do
	if len(cas.nearbyActions) == 0 {
		return
	}

	// Sort by distance (closest first), then by priority
	sort.Slice(cas.nearbyActions, func(i, j int) bool {
		distDiff := cas.nearbyActions[i].distance - cas.nearbyActions[j].distance
		if math.Abs(distDiff) < 10.0 { // Within 10 pixels, use priority
			return cas.nearbyActions[i].priority > cas.nearbyActions[j].priority
		}
		return cas.nearbyActions[i].distance < cas.nearbyActions[j].distance
	})

	// Set the closest/highest priority as active
	cas.activeAction = cas.nearbyActions[0]

	if cas.logger != nil && cas.logger.Logger.GetLevel() >= logrus.DebugLevel {
		cas.logger.WithFields(logrus.Fields{
			"action":   cas.activeAction.action.ActionType.String(),
			"distance": cas.activeAction.distance,
			"entity":   cas.activeAction.entity.ID,
		}).Debug("Active context action updated")
	}
}

// getActionPriority returns priority for action types (higher = more important).
// Talk > Activate > Open > Read > Pickup > Close > Push/Pull
func (cas *ContextActionSystem) getActionPriority(actionType ContextActionType) int {
	switch actionType {
	case ActionTalk:
		return 100 // Highest priority - NPCs
	case ActionActivate:
		return 90 // Levers, switches
	case ActionOpen:
		return 80 // Doors, chests
	case ActionRead:
		return 70 // Signs, books
	case ActionPickup:
		return 60 // Items
	case ActionClose:
		return 50 // Close door
	case ActionPush:
		return 40 // Push object
	case ActionPull:
		return 40 // Pull object
	default:
		return 0
	}
}

// GetActiveAction returns the currently active context action if any.
// Returns nil if no action is available.
func (cas *ContextActionSystem) GetActiveAction() (entity *Entity, action *ContextActionComponent) {
	if cas.activeAction != nil {
		return cas.activeAction.entity, cas.activeAction.action
	}
	return nil, nil
}

// GetActivePrompt returns the prompt text for the active action.
// Returns empty string if no action is active.
func (cas *ContextActionSystem) GetActivePrompt() string {
	if cas.activeAction != nil {
		return cas.activeAction.action.Prompt
	}
	return ""
}

// TriggerActiveAction activates the currently active context action.
// This should be called when the player presses the interaction key (F).
// Returns error if no action is active or activation fails.
func (cas *ContextActionSystem) TriggerActiveAction() error {
	if cas.activeAction == nil {
		return nil // No error, just no action to trigger
	}

	entity := cas.activeAction.entity
	action := cas.activeAction.action

	if cas.logger != nil && cas.logger.Logger.GetLevel() >= logrus.InfoLevel {
		cas.logger.WithFields(logrus.Fields{
			"action": action.ActionType.String(),
			"entity": entity.ID,
		}).Info("Context action triggered")
	}

	// Trigger the action
	if err := action.Activate(entity, cas.world); err != nil {
		if cas.logger != nil {
			cas.logger.WithError(err).WithField("action", action.ActionType.String()).Error("Context action failed")
		}
		return err
	}

	return nil
}

// GetNearbyActionsCount returns the number of nearby interactive objects.
func (cas *ContextActionSystem) GetNearbyActionsCount() int {
	return len(cas.nearbyActions)
}
