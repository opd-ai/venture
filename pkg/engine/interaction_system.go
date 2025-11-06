// Package engine provides interaction system for Phase 11.2 and Phase 11.3.
// Puzzle Interaction System + Context Actions
//
// This file implements the InteractionSystem which handles player interactions:
// - Puzzle element interactions (Phase 11.2)
// - Context-sensitive actions (Phase 11.3): doors, levers, chests, NPCs
// - Carriable object pickup (Phase 11.3)
// All interactions use the F key (or touch input on mobile).
package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sirupsen/logrus"
)

// InteractionSystem manages player interactions with interactive entities.
type InteractionSystem struct {
	world       *World
	logger      *logrus.Entry
	carrySystem *CarrySystem // For pickup/throw mechanics
}

// NewInteractionSystem creates a new interaction system.
func NewInteractionSystem(world *World) *InteractionSystem {
	logger := world.GetLogger()
	return &InteractionSystem{
		world:  world,
		logger: logger,
	}
}

// SetCarrySystem sets the carry system reference for pickup mechanics.
func (s *InteractionSystem) SetCarrySystem(carrySystem *CarrySystem) {
	s.carrySystem = carrySystem
}

// Update checks for interaction key presses and processes player interactions.
func (s *InteractionSystem) Update(entities []*Entity, deltaTime float64) {
	// Update context action cooldowns for all entities
	contextActions := s.world.GetEntitiesWith("contextAction")
	for _, entity := range contextActions {
		if comp, ok := entity.GetComponent("contextAction"); ok {
			if ctx, ok := comp.(*ContextActionComponent); ok {
				ctx.Update(deltaTime)
			}
		}
	}

	// Check if F key was just pressed (or touch input on mobile)
	interactionPressed := inpututil.IsKeyJustPressed(ebiten.KeyF)

	if !interactionPressed {
		return
	}

	// Get player entity
	players := s.world.GetEntitiesWith("player")
	if len(players) == 0 {
		return
	}
	player := players[0]

	// Get player position
	playerPosComp, ok := player.GetComponent("position")
	if !ok {
		return
	}
	playerPos := playerPosComp.(*PositionComponent)

	// Priority order for interactions:
	// 1. Context actions (doors, levers, etc.)
	// 2. Carriable objects (pickup)
	// 3. Puzzle elements

	// Try context actions first
	if s.tryContextActions(player, playerPos) {
		return
	}

	// Try carriable object pickup
	if s.tryCarriablePickup(player, playerPos) {
		return
	}

	// Fall back to puzzle element interactions
	s.tryPuzzleInteraction(player, playerPos)
}

// tryContextActions attempts to interact with context action entities.
// Returns true if an interaction was performed.
func (s *InteractionSystem) tryContextActions(player *Entity, playerPos *PositionComponent) bool {
	// Find nearby context actions
	contextActions := s.world.GetEntitiesWith("contextAction")
	for _, entity := range contextActions {
		// Get entity position
		entPosComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entPos := entPosComp.(*PositionComponent)

		// Get context action component
		ctxComp, ok := entity.GetComponent("contextAction")
		if !ok {
			continue
		}
		contextAction := ctxComp.(*ContextActionComponent)

		// Skip if not available or on cooldown
		if !contextAction.CanInteract() {
			continue
		}

		// Check distance to player
		dx := entPos.X - playerPos.X
		dy := entPos.Y - playerPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// If within interaction range, activate action
		if distance <= contextAction.InteractionRange {
			s.activateContextAction(player, entity, contextAction)
			return true
		}
	}

	return false
}

// activateContextAction handles context action activation.
func (s *InteractionSystem) activateContextAction(player, entity *Entity, contextAction *ContextActionComponent) {
	// Activate the action
	contextAction.Activate()

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":   entity.ID,
			"actionType": contextAction.ActionType.String(),
			"actionText": contextAction.ActionText,
		}).Debug("context action activated")
	}

	// Perform action-specific behavior
	switch contextAction.ActionType {
	case ActionOpen:
		s.handleOpenAction(entity)
	case ActionClose:
		s.handleCloseAction(entity)
	case ActionActivate:
		s.handleActivateAction(entity)
		// Other actions can be added here
	}
}

// handleOpenAction handles opening doors, chests, etc.
func (s *InteractionSystem) handleOpenAction(entity *Entity) {
	// Update action text to "Close" if it's a door
	if ctxComp, ok := entity.GetComponent("contextAction"); ok {
		if ctx, ok := ctxComp.(*ContextActionComponent); ok {
			ctx.ActionType = ActionClose
			ctx.ActionText = "Close"
		}
	}

	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("door opened")
	}
}

// handleCloseAction handles closing doors.
func (s *InteractionSystem) handleCloseAction(entity *Entity) {
	// Update action text to "Open"
	if ctxComp, ok := entity.GetComponent("contextAction"); ok {
		if ctx, ok := ctxComp.(*ContextActionComponent); ok {
			ctx.ActionType = ActionOpen
			ctx.ActionText = "Open"
		}
	}

	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("door closed")
	}
}

// handleActivateAction handles lever/switch activation.
func (s *InteractionSystem) handleActivateAction(entity *Entity) {
	if s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Debug("lever activated")
	}
	// Can trigger puzzle state changes, open gates, etc.
}

// tryCarriablePickup attempts to pick up a carriable object.
// Returns true if a pickup was performed.
func (s *InteractionSystem) tryCarriablePickup(player *Entity, playerPos *PositionComponent) bool {
	if s.carrySystem == nil {
		return false
	}

	// Check if player is already carrying something
	if s.carrySystem.IsCarrying(player.ID) {
		// Drop the carried object
		s.carrySystem.DropObject(player.ID)
		return true
	}

	// Find nearby carriable object
	maxDistance := 48.0 // 1.5 tiles
	objectID, distance := s.carrySystem.FindNearbyCarriableObject(playerPos.X, playerPos.Y, maxDistance)

	if objectID != 0 {
		// Try to pick up the object
		if s.carrySystem.TryPickup(player.ID, objectID) {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"playerID": player.ID,
					"objectID": objectID,
					"distance": distance,
				}).Debug("object picked up")
			}
			return true
		}
	}

	return false
}

// tryPuzzleInteraction attempts to interact with puzzle elements.
func (s *InteractionSystem) tryPuzzleInteraction(player *Entity, playerPos *PositionComponent) {
	// Find nearby puzzle elements
	puzzleElements := s.world.GetEntitiesWith("puzzleElement")

	for _, element := range puzzleElements {
		// Get element position
		elemPosComp, ok := element.GetComponent("position")
		if !ok {
			continue
		}
		elemPos := elemPosComp.(*PositionComponent)

		// Get puzzle element component
		elemComp, ok := element.GetComponent("puzzleElement")
		if !ok {
			continue
		}
		puzzleElem := elemComp.(*PuzzleElementComponent)

		// Skip if not interactable or on cooldown
		if !puzzleElem.IsInteractable || puzzleElem.CooldownElapsed > 0 {
			continue
		}

		// Check distance to player
		dx := elemPos.X - playerPos.X
		dy := elemPos.Y - playerPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// If within interaction range, activate element
		if distance <= puzzleElem.InteractionRange {
			s.activatePuzzleElement(player, element, puzzleElem)
			return
		}
	}
}

// activatePuzzleElement handles the interaction with a puzzle element.
func (s *InteractionSystem) activatePuzzleElement(player, element *Entity, puzzleElem *PuzzleElementComponent) {
	// Toggle element state
	puzzleElem.State = (puzzleElem.State + 1) % 2 // Toggle between 0 and 1

	// Set cooldown
	puzzleElem.CooldownElapsed = puzzleElem.CooldownTime

	// Record activation
	puzzleElem.ActivatedBy = player.ID

	// Find parent puzzle and record progress
	puzzles := s.world.GetEntitiesWith("puzzle")
	for _, puzzleEntity := range puzzles {
		puzzleComp, ok := puzzleEntity.GetComponent("puzzle")
		if !ok {
			continue
		}

		puzzle := puzzleComp.(*PuzzleComponent)
		if puzzle.PuzzleID == puzzleElem.PuzzleID {
			// Record progress
			puzzle.RecordProgress(puzzleElem.ElementName)

			// Start solving if not yet started
			if puzzle.State == PuzzleStateUnsolved {
				puzzle.State = PuzzleStateSolving
			}

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"puzzleID":    puzzle.PuzzleID,
					"elementName": puzzleElem.ElementName,
					"elementType": puzzleElem.ElementType,
					"newState":    puzzleElem.State,
					"progress":    len(puzzle.CurrentProgress),
					"solution":    len(puzzle.Solution),
				}).Debug("puzzle element activated")
			}

			break
		}
	}
}
