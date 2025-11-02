// Package engine provides interaction system for Phase 11.2.
// Puzzle Interaction System
//
// This file implements the InteractionSystem which handles player-to-puzzle
// element interactions using the F key (or touch input on mobile).
package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sirupsen/logrus"
)

// InteractionSystem manages player interactions with puzzle elements and other interactable entities.
type InteractionSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewInteractionSystem creates a new interaction system.
func NewInteractionSystem(world *World) *InteractionSystem {
	logger := world.GetLogger()
	return &InteractionSystem{
		world:  world,
		logger: logger,
	}
}

// Update checks for interaction key presses and processes player-puzzle interactions.
func (s *InteractionSystem) Update(entities []*Entity, deltaTime float64) {
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
		}
	}
}

// activatePuzzleElement handles the interaction with a puzzle element.
func (s *InteractionSystem) activatePuzzleElement(player *Entity, element *Entity, puzzleElem *PuzzleElementComponent) {
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

			s.logger.WithFields(logrus.Fields{
				"puzzleID":    puzzle.PuzzleID,
				"elementName": puzzleElem.ElementName,
				"elementType": puzzleElem.ElementType,
				"newState":    puzzleElem.State,
				"progress":    len(puzzle.CurrentProgress),
				"solution":    len(puzzle.Solution),
			}).Debug("puzzle element activated")

			break
		}
	}
}
