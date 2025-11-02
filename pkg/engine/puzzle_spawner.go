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
	if world == nil {
		return 0, fmt.Errorf("world cannot be nil")
	}
	if terrainData == nil {
		return 0, fmt.Errorf("terrain cannot be nil")
	}
	if targetCount <= 0 {
		return 0, nil // No puzzles to spawn
	}

	logger := world.GetLogger()

	// Create puzzle generator
	puzzleGen := puzzle.NewGenerator()

	// Create RNG for deterministic room selection
	rng := rand.New(rand.NewSource(seed))

	// Find suitable rooms for puzzles (exclude starting room at index 0)
	suitableRooms := make([]*terrain.Room, 0)
	for i, room := range terrainData.Rooms {
		if i == 0 {
			continue // Skip starting room
		}

		// Room must be at least 5x5 to accommodate puzzle elements
		if room.Width >= 5 && room.Height >= 5 {
			suitableRooms = append(suitableRooms, room)
		}
	}

	if len(suitableRooms) == 0 {
		logger.Warn("no suitable rooms found for puzzle spawning")
		return 0, nil
	}

	// Calculate actual puzzle count (40-60% of suitable rooms)
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

	// Shuffle rooms for random selection
	rng.Shuffle(len(suitableRooms), func(i, j int) {
		suitableRooms[i], suitableRooms[j] = suitableRooms[j], suitableRooms[i]
	})

	// Spawn puzzles in selected rooms
	puzzlesSpawned := 0
	for i := 0; i < actualCount && i < len(suitableRooms); i++ {
		room := suitableRooms[i]

		// Generate puzzle with unique seed per room
		puzzleSeed := seed + int64(i)*1000 + int64(room.X)*100 + int64(room.Y)*10
		result, err := puzzleGen.Generate(puzzleSeed, params)
		if err != nil {
			logger.WithError(err).WithField("roomIndex", i).Warn("failed to generate puzzle")
			continue
		}

		puz, ok := result.(*puzzle.Puzzle)
		if !ok {
			logger.WithField("roomIndex", i).Warn("puzzle generation returned invalid type")
			continue
		}

		// Validate puzzle
		if err := puzzleGen.Validate(puz); err != nil {
			logger.WithError(err).WithField("puzzleID", puz.ID).Warn("puzzle validation failed")
			continue
		}

		// Spawn puzzle entity and elements in the room
		if err := spawnPuzzleInRoom(world, puz, room, rng); err != nil {
			logger.WithError(err).WithField("puzzleID", puz.ID).Warn("failed to spawn puzzle in room")
			continue
		}

		puzzlesSpawned++

		logger.WithFields(logrus.Fields{
			"puzzleID":   puz.ID,
			"puzzleType": puz.Type,
			"difficulty": puz.Difficulty,
			"roomX":      room.X,
			"roomY":      room.Y,
			"elements":   puz.ElementCount,
		}).Debug("spawned puzzle in room")
	}

	logger.WithFields(logrus.Fields{
		"puzzlesSpawned": puzzlesSpawned,
		"puzzlesAttempt": actualCount,
		"suitableRooms":  len(suitableRooms),
	}).Info("puzzle spawning complete")

	return puzzlesSpawned, nil
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
		world.GetLogger().WithFields(logrus.Fields{
			"puzzleID":   puz.ID,
			"elementID":  elem.ID,
			"elementNum": i + 1,
			"position":   fmt.Sprintf("(%d, %d)", elemX, elemY),
		}).Debug("spawned puzzle element")
	}

	return nil
}
