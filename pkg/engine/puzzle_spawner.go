// Package engine provides puzzle spawning functionality for Phase 11.2.
// Procedural Puzzle Integration System
//
// This file implements functions to spawn puzzles in dungeon terrain during
// world generation, creating constraint-solving challenges for players.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/puzzle"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// SpawnPuzzlesInTerrain spawns procedural puzzles in dungeon rooms.
// Returns the number of puzzles spawned and any error encountered.
//
// Parameters:
//   - world: The game world to spawn puzzles in
//   - terrainData: The generated terrain containing rooms
//   - seed: Random seed for deterministic puzzle generation
//   - params: Generation parameters (difficulty, depth, genre)
//   - targetCount: Desired number of puzzles to spawn
//
// Puzzles are placed in rooms with sufficient space (at least 5x5 tiles).
// Each puzzle spawns a main puzzle entity and multiple puzzle element entities.
func SpawnPuzzlesInTerrain(world *World, terrainData *terrain.Terrain, seed int64, params procgen.GenerationParams, targetCount int) (int, error) {
	if err := validateSpawnInputs(world, terrainData, targetCount); err != nil {
		return 0, err
	}

	logger := world.GetLogger()
	puzzleGen := puzzle.NewGenerator()
	rng := rand.New(rand.NewSource(seed))

	suitableRooms := findSuitableRooms(terrainData.Rooms)
	if len(suitableRooms) == 0 {
		logNoSuitableRooms(logger)
		return 0, nil
	}

	actualCount := calculatePuzzleCount(suitableRooms, targetCount, rng)
	shuffleRooms(suitableRooms, rng)

	puzzlesSpawned := spawnPuzzlesInRooms(world, puzzleGen, suitableRooms, actualCount, seed, params, rng, logger)

	logSpawningComplete(logger, puzzlesSpawned, actualCount, len(suitableRooms))
	return puzzlesSpawned, nil
}

// validateSpawnInputs checks that world and terrain are valid and targetCount is positive.
func validateSpawnInputs(world *World, terrainData *terrain.Terrain, targetCount int) error {
	if world == nil {
		return fmt.Errorf("world cannot be nil")
	}
	if terrainData == nil {
		return fmt.Errorf("terrain cannot be nil")
	}
	if targetCount <= 0 {
		return nil // No puzzles to spawn, not an error
	}
	return nil
}

// findSuitableRooms returns rooms that are at least 5x5 tiles, excluding the starting room.
func findSuitableRooms(rooms []*terrain.Room) []*terrain.Room {
	suitableRooms := make([]*terrain.Room, 0)
	for i, room := range rooms {
		if i == 0 {
			continue // Skip starting room
		}
		if room.Width >= 5 && room.Height >= 5 {
			suitableRooms = append(suitableRooms, room)
		}
	}
	return suitableRooms
}

// calculatePuzzleCount determines how many puzzles to spawn based on room availability.
func calculatePuzzleCount(suitableRooms []*terrain.Room, targetCount int, rng *rand.Rand) int {
	minPuzzles := int(float64(len(suitableRooms)) * 0.4)
	maxPuzzles := int(float64(len(suitableRooms)) * 0.6)
	if maxPuzzles < minPuzzles {
		maxPuzzles = minPuzzles
	}

	actualCount := minPuzzles + rng.Intn(maxPuzzles-minPuzzles+1)
	if actualCount > targetCount {
		actualCount = targetCount
	}
	if actualCount > len(suitableRooms) {
		actualCount = len(suitableRooms)
	}
	return actualCount
}

// shuffleRooms randomizes room order for puzzle placement.
func shuffleRooms(rooms []*terrain.Room, rng *rand.Rand) {
	rng.Shuffle(len(rooms), func(i, j int) {
		rooms[i], rooms[j] = rooms[j], rooms[i]
	})
}

// spawnPuzzlesInRooms generates and spawns puzzles in selected rooms.
func spawnPuzzlesInRooms(world *World, puzzleGen *puzzle.Generator, rooms []*terrain.Room, count int, seed int64, params procgen.GenerationParams, rng *rand.Rand, logger *logrus.Entry) int {
	puzzlesSpawned := 0
	for i := 0; i < count && i < len(rooms); i++ {
		room := rooms[i]
		puz := generateAndValidatePuzzle(puzzleGen, seed, i, room, params, logger)
		if puz == nil {
			continue
		}

		if err := spawnPuzzleInRoom(world, puz, room, rng); err != nil {
			logSpawnError(logger, err, puz.ID)
			continue
		}

		puzzlesSpawned++
		logPuzzleSpawned(logger, puz, room)
	}
	return puzzlesSpawned
}

// generateAndValidatePuzzle creates and validates a puzzle, returning nil on failure.
func generateAndValidatePuzzle(puzzleGen *puzzle.Generator, seed int64, roomIndex int, room *terrain.Room, params procgen.GenerationParams, logger *logrus.Entry) *puzzle.Puzzle {
	puzzleSeed := seed + int64(roomIndex)*1000 + int64(room.X)*100 + int64(room.Y)*10
	result, err := puzzleGen.Generate(puzzleSeed, params)
	if err != nil {
		logGenerationError(logger, err, roomIndex)
		return nil
	}

	puz, ok := result.(*puzzle.Puzzle)
	if !ok {
		logInvalidType(logger, roomIndex)
		return nil
	}

	if err := puzzleGen.Validate(puz); err != nil {
		logValidationError(logger, err, puz.ID)
		return nil
	}

	return puz
}

// logNoSuitableRooms logs a warning when no suitable rooms are found.
func logNoSuitableRooms(logger *logrus.Entry) {
	if logger != nil {
		logger.Warn("no suitable rooms found for puzzle spawning")
	}
}

// logGenerationError logs puzzle generation failures.
func logGenerationError(logger *logrus.Entry, err error, roomIndex int) {
	if logger != nil {
		logger.WithError(err).WithField("roomIndex", roomIndex).Warn("failed to generate puzzle")
	}
}

// logInvalidType logs when puzzle generation returns wrong type.
func logInvalidType(logger *logrus.Entry, roomIndex int) {
	if logger != nil {
		logger.WithField("roomIndex", roomIndex).Warn("puzzle generation returned invalid type")
	}
}

// logValidationError logs puzzle validation failures.
func logValidationError(logger *logrus.Entry, err error, puzzleID string) {
	if logger != nil {
		logger.WithError(err).WithField("puzzleID", puzzleID).Warn("puzzle validation failed")
	}
}

// logSpawnError logs errors spawning puzzles in rooms.
func logSpawnError(logger *logrus.Entry, err error, puzzleID string) {
	if logger != nil {
		logger.WithError(err).WithField("puzzleID", puzzleID).Warn("failed to spawn puzzle in room")
	}
}

// logPuzzleSpawned logs successful puzzle spawning with details.
func logPuzzleSpawned(logger *logrus.Entry, puz *puzzle.Puzzle, room *terrain.Room) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"puzzleID":   puz.ID,
			"puzzleType": puz.Type,
			"difficulty": puz.Difficulty,
			"roomX":      room.X,
			"roomY":      room.Y,
			"elements":   puz.ElementCount,
		}).Debug("spawned puzzle in room")
	}
}

// logSpawningComplete logs the final summary of puzzle spawning.
func logSpawningComplete(logger *logrus.Entry, spawned, attempted, suitableCount int) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"puzzlesSpawned": spawned,
			"puzzlesAttempt": attempted,
			"suitableRooms":  suitableCount,
		}).Info("puzzle spawning complete")
	}
}

// spawnPuzzleInRoom creates the puzzle entity and all element entities in the specified room.
func spawnPuzzleInRoom(world *World, puz *puzzle.Puzzle, room *terrain.Room, rng *rand.Rand) error {
	// Create main puzzle entity
	puzzleEntity := world.CreateEntity()

	// Create puzzle component
	puzzleComp := NewPuzzleComponent(
		puz.ID,
		PuzzleType(puz.Type),
		puz.Difficulty,
	)
	puzzleComp.Solution = puz.Solution
	puzzleComp.TimeLimit = puz.TimeLimit
	puzzleComp.MaxAttempts = puz.MaxAttempts
	puzzleComp.HintText = puz.HintText
	puzzleComp.Description = puz.Description

	// Add position at center of room
	centerX := float64(room.X + room.Width/2)
	centerY := float64(room.Y + room.Height/2)
	puzzleEntity.AddComponent(&PositionComponent{
		X: centerX * 32.0, // Convert grid to pixels
		Y: centerY * 32.0,
	})

	// Add puzzle component
	puzzleEntity.AddComponent(puzzleComp)

	// Spawn puzzle elements in the room
	for i, elem := range puz.Elements {
		elementEntity := world.CreateEntity()

		// Calculate element position within room bounds
		// Use grid positions from puzzle, offset by room position
		elemX := room.X + elem.Position[0]
		elemY := room.Y + elem.Position[1]

		// Ensure element is within room bounds
		if elemX < room.X || elemX >= room.X+room.Width ||
			elemY < room.Y || elemY >= room.Y+room.Height {
			// If out of bounds, place randomly in room
			elemX = room.X + 1 + rng.Intn(room.Width-2)
			elemY = room.Y + 1 + rng.Intn(room.Height-2)
		}

		// Add position component (convert grid to pixels)
		elementEntity.AddComponent(&PositionComponent{
			X: float64(elemX) * 32.0,
			Y: float64(elemY) * 32.0,
		})

		// Add puzzle element component
		// Convert interface{} state to int (default to 0 if not convertible)
		stateInt := 0
		if s, ok := elem.State.(int); ok {
			stateInt = s
		} else if s, ok := elem.State.(float64); ok {
			stateInt = int(s)
		}

		puzzleElemComp := &PuzzleElementComponent{
			ElementName:      elem.ID,
			PuzzleID:         puz.ID,
			ElementType:      elem.ElementType,
			State:            stateInt,
			TargetState:      1, // Default target state
			InteractionRange: 48.0,
			CooldownTime:     0.5,
			CooldownElapsed:  0.0,
			IsInteractable:   elem.Interactable,
		}
		elementEntity.AddComponent(puzzleElemComp)

		// Store element entity ID in puzzle component
		puzzleComp.ElementIDs = append(puzzleComp.ElementIDs, elementEntity.ID)

		// Log element spawn
		if logger := world.GetLogger(); logger != nil {
			logger.WithFields(logrus.Fields{
				"puzzleID":   puz.ID,
				"elementID":  elem.ID,
				"elementNum": i + 1,
				"position":   fmt.Sprintf("(%d, %d)", elemX, elemY),
			}).Debug("spawned puzzle element")
		}
	}

	return nil
}
