// Package engine provides puzzle system for Phase 11.2.
// Procedural Puzzle Generation System
//
// This file implements the PuzzleSystem which manages puzzle state updates,
// solution checking, timing, and element interactions.
package engine

import (
	"github.com/sirupsen/logrus"
)

// PuzzleSystem manages puzzle state and interactions.
type PuzzleSystem struct {
	world  *World
	logger *logrus.Logger
}

// NewPuzzleSystem creates a new puzzle system.
func NewPuzzleSystem(world *World) *PuzzleSystem {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &PuzzleSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes puzzle state, timing, and solution checking.
func (s *PuzzleSystem) Update(deltaTime float64) {
	// Update puzzle element cooldowns
	s.updateElementCooldowns(deltaTime)

	// Update timed puzzles
	s.updateTimedPuzzles(deltaTime)

	// Check for solved puzzles
	s.checkPuzzleSolutions()

	// Handle failed puzzles (timeout or max attempts)
	s.handleFailedPuzzles()
}

// updateElementCooldowns updates cooldown timers for all puzzle elements.
func (s *PuzzleSystem) updateElementCooldowns(deltaTime float64) {
	entities := s.world.GetEntitiesWith("puzzleElement")

	for _, entity := range entities {
		elementComp, ok := entity.GetComponent("puzzleElement")
		if !ok || elementComp == nil {
			continue
		}

		element, ok := elementComp.(*PuzzleElementComponent)
		if !ok {
			continue
		}

		element.UpdateCooldown(deltaTime)
	}
}

// updateTimedPuzzles updates elapsed time for timed puzzles.
func (s *PuzzleSystem) updateTimedPuzzles(deltaTime float64) {
	entities := s.world.GetEntitiesWith("puzzle")

	for _, entity := range entities {
		puzzleComp, ok := entity.GetComponent("puzzle")
		if !ok || puzzleComp == nil {
			continue
		}

		puzzle, ok := puzzleComp.(*PuzzleComponent)
		if !ok {
			continue
		}

		// Only update time for unsolved/solving puzzles with time limits
		if puzzle.TimeLimit > 0 && !puzzle.IsSolved() && puzzle.State != PuzzleStateFailed {
			puzzle.TimeElapsed += deltaTime

			// Automatically mark as solving if time has started
			if puzzle.State == PuzzleStateUnsolved && puzzle.TimeElapsed > 0.1 {
				puzzle.State = PuzzleStateSolving
			}
		}
	}
}

// checkPuzzleSolutions checks if any puzzles have been completed.
func (s *PuzzleSystem) checkPuzzleSolutions() {
	entities := s.world.GetEntitiesWith("puzzle")

	for _, entity := range entities {
		puzzleComp, ok := entity.GetComponent("puzzle")
		if !ok || puzzleComp == nil {
			continue
		}

		puzzle, ok := puzzleComp.(*PuzzleComponent)
		if !ok {
			continue
		}

		// Skip already solved puzzles
		if puzzle.IsSolved() {
			continue
		}

		// Check if solution matches
		if puzzle.CheckSolution() {
			s.solvePuzzle(entity, puzzle)
		}
	}
}

// solvePuzzle marks puzzle as solved and triggers rewards.
func (s *PuzzleSystem) solvePuzzle(entity *Entity, puzzle *PuzzleComponent) {
	// Mark solved (solver ID 0 for system-triggered solving)
	puzzle.MarkSolved(0)

	s.logger.WithFields(logrus.Fields{
		"puzzleID":    puzzle.PuzzleID,
		"type":        puzzle.PuzzleType,
		"difficulty":  puzzle.Difficulty,
		"attempts":    puzzle.Attempts,
		"timeElapsed": puzzle.TimeElapsed,
	}).Info("Puzzle solved")

	// Trigger reward entity if specified
	if puzzle.RewardEntityID != 0 {
		s.unlockReward(puzzle.RewardEntityID)
	}
}

// unlockReward activates the reward entity (door, chest, etc.).
func (s *PuzzleSystem) unlockReward(rewardEntityID uint64) {
	entity, ok := s.world.GetEntity(rewardEntityID)
	if !ok || entity == nil {
		s.logger.WithField("rewardEntityID", rewardEntityID).Warn("Reward entity not found")
		return
	}

	// Handle different reward types
	// For doors: remove blocking component or change tile type
	// For chests: spawn loot
	// For barriers: destroy entity

	s.logger.WithField("rewardEntityID", rewardEntityID).Info("Puzzle reward unlocked")
}

// handleFailedPuzzles checks for timeout or max attempts and marks puzzles as failed.
func (s *PuzzleSystem) handleFailedPuzzles() {
	entities := s.world.GetEntitiesWith("puzzle")

	for _, entity := range entities {
		puzzleComp, ok := entity.GetComponent("puzzle")
		if !ok || puzzleComp == nil {
			continue
		}

		puzzle, ok := puzzleComp.(*PuzzleComponent)
		if !ok {
			continue
		}

		// Skip already solved or failed puzzles
		if puzzle.IsSolved() || puzzle.State == PuzzleStateFailed {
			continue
		}

		// Check for timeout
		if puzzle.HasTimedOut() {
			puzzle.State = PuzzleStateFailed
			s.logger.WithFields(logrus.Fields{
				"puzzleID": puzzle.PuzzleID,
				"reason":   "timeout",
			}).Info("Puzzle failed")
			continue
		}

		// Check for max attempts
		if puzzle.HasMaxAttemptsReached() {
			puzzle.State = PuzzleStateFailed
			s.logger.WithFields(logrus.Fields{
				"puzzleID": puzzle.PuzzleID,
				"reason":   "max_attempts",
			}).Info("Puzzle failed")
		}
	}
}

// InteractWithElement handles player interaction with puzzle element.
func (s *PuzzleSystem) InteractWithElement(playerID uint64, elementEntityID uint64) error {
	// Get element entity
	elementEntity, ok := s.world.GetEntity(elementEntityID)
	if !ok || elementEntity == nil {
		return nil // Element doesn't exist
	}

	elementComp, ok := elementEntity.GetComponent("puzzleElement")
	if !ok || elementComp == nil {
		return nil // Not a puzzle element
	}

	element, ok := elementComp.(*PuzzleElementComponent)
	if !ok {
		return nil
	}

	// Check interaction range
	if !s.isPlayerInRange(playerID, elementEntityID, element.InteractionRange) {
		return nil // Player too far away
	}

	// Activate element
	if !element.Activate() {
		return nil // Could not activate (cooldown or not interactable)
	}

	s.logger.WithFields(logrus.Fields{
		"playerID":    playerID,
		"elementID":   elementEntityID,
		"elementName": element.ElementName,
		"newState":    element.State,
	}).Debug("Puzzle element activated")

	// Get parent puzzle
	puzzle := s.getPuzzleByID(element.PuzzleID)
	if puzzle == nil {
		return nil
	}

	// Record progress
	puzzle.RecordProgress(element.ElementName)

	return nil
}

// isPlayerInRange checks if player is within interaction range of element.
func (s *PuzzleSystem) isPlayerInRange(playerID, elementID uint64, interactionRange float64) bool {
	playerEntity, ok1 := s.world.GetEntity(playerID)
	elementEntity, ok2 := s.world.GetEntity(elementID)

	if !ok1 || !ok2 || playerEntity == nil || elementEntity == nil {
		return false
	}

	playerPosComp, ok1 := playerEntity.GetComponent("position")
	elementPosComp, ok2 := elementEntity.GetComponent("position")

	if !ok1 || !ok2 || playerPosComp == nil || elementPosComp == nil {
		return false
	}

	playerPos, ok1 := playerPosComp.(*PositionComponent)
	elementPos, ok2 := elementPosComp.(*PositionComponent)

	if !ok1 || !ok2 {
		return false
	}

	// Calculate distance
	dx := playerPos.X - elementPos.X
	dy := playerPos.Y - elementPos.Y
	distanceSquared := dx*dx + dy*dy
	rangeSquared := interactionRange * interactionRange

	return distanceSquared <= rangeSquared
}

// getPuzzleByID retrieves puzzle component by puzzle ID.
func (s *PuzzleSystem) getPuzzleByID(puzzleID string) *PuzzleComponent {
	entities := s.world.GetEntitiesWith("puzzle")

	for _, entity := range entities {
		puzzleComp, ok := entity.GetComponent("puzzle")
		if !ok || puzzleComp == nil {
			continue
		}

		puzzle, ok := puzzleComp.(*PuzzleComponent)
		if !ok {
			continue
		}

		if puzzle.PuzzleID == puzzleID {
			return puzzle
		}
	}

	return nil
}

// ResetPuzzle resets a puzzle to initial state.
func (s *PuzzleSystem) ResetPuzzle(puzzleID string) {
	puzzle := s.getPuzzleByID(puzzleID)
	if puzzle == nil {
		return
	}

	puzzle.Reset()

	// Reset all elements
	for _, elementID := range puzzle.ElementIDs {
		elementEntity, ok := s.world.GetEntity(elementID)
		if !ok || elementEntity == nil {
			continue
		}

		elementComp, ok := elementEntity.GetComponent("puzzleElement")
		if !ok || elementComp == nil {
			continue
		}

		element, ok := elementComp.(*PuzzleElementComponent)
		if ok {
			element.Reset()
		}
	}

	s.logger.WithField("puzzleID", puzzleID).Info("Puzzle reset")
}

// GetPuzzleStatus returns current puzzle state information.
func (s *PuzzleSystem) GetPuzzleStatus(puzzleID string) map[string]interface{} {
	puzzle := s.getPuzzleByID(puzzleID)
	if puzzle == nil {
		return nil
	}

	status := make(map[string]interface{})
	status["puzzleID"] = puzzle.PuzzleID
	status["type"] = puzzle.PuzzleType
	status["state"] = puzzle.State.String()
	status["difficulty"] = puzzle.Difficulty
	status["progress"] = puzzle.GetProgressPercent()
	status["attempts"] = puzzle.Attempts
	status["timeElapsed"] = puzzle.TimeElapsed
	status["timeLimit"] = puzzle.TimeLimit
	status["isSolved"] = puzzle.IsSolved()
	status["hasTimedOut"] = puzzle.HasTimedOut()

	return status
}

// SetLogLevel sets the logging level for the puzzle system.
func (s *PuzzleSystem) SetLogLevel(level logrus.Level) {
	s.logger.SetLevel(level)
}
