// Package engine provides mini-game station spawning for Phase 27.3.
// Spawns interactive game stations in taverns and safe zones.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen"
)

// SpawnMiniGameStation creates a mini-game station entity at the specified location.
// Station type and difficulty are determined by world depth and genre.
//
// Phase 27.3: Mini-Game Integration
func SpawnMiniGameStation(world *World, x, y float64, seed int64, params procgen.GenerationParams) (*Entity, error) {
	if world == nil {
		return nil, fmt.Errorf("world cannot be nil")
	}

	// Create deterministic RNG for this station
	rng := rand.New(rand.NewSource(seed))

	// Select game type based on genre and randomness
	gameType := selectGameType(rng, params.GenreID)

	// Calculate difficulty based on depth
	difficulty := calculateStationDifficulty(params.Depth, params.Difficulty)

	// Create station entity
	entity := world.CreateEntity()

	// Add position
	posComp := &PositionComponent{X: x, Y: y}
	entity.AddComponent(posComp)

	// Add velocity (stations don't move)
	velComp := &VelocityComponent{VX: 0, VY: 0}
	entity.AddComponent(velComp)

	// Add mini-game station component
	stationComp := NewMiniGameStationComponent(gameType, difficulty)
	entity.AddComponent(stationComp)

	// Add context action for interaction
	actionText := fmt.Sprintf("Play %s", gameType.String())
	contextAction := NewContextActionComponent(ActionPlayGame, actionText)
	contextAction.InteractionRange = 64.0 // Slightly larger range for stations
	entity.AddComponent(contextAction)

	// Add collider so player can't walk through stations
	collider := &ColliderComponent{
		Width:     32.0,
		Height:    32.0,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -16.0,
		OffsetY:   -16.0,
	}
	entity.AddComponent(collider)

	// Add team component (neutral team = 0)
	teamComp := &TeamComponent{TeamID: 0}
	entity.AddComponent(teamComp)

	// Add sprite component (Ebiten sprite for rendering)
	stationSprite := &EbitenSprite{
		Image:   ebiten.NewImage(32, 32),
		Width:   32,
		Height:  32,
		Visible: true,
		Layer:   5, // Below player layer
	}
	entity.AddComponent(stationSprite)

	// Add animation component (idle animation)
	stationSeed := seed + int64(gameType)*1000
	animComp := NewAnimationComponent(stationSeed)
	animComp.CurrentState = AnimationStateIdle
	animComp.FrameTime = 0.3 // Slow idle animation
	animComp.Loop = true
	animComp.Playing = true
	animComp.FrameCount = 2 // Simple 2-frame idle
	entity.AddComponent(animComp)

	return entity, nil
}

// SpawnMultipleStations spawns multiple mini-game stations in a safe zone.
// Returns slice of created station entities.
//
// Phase 27.3: Mini-Game Integration
func SpawnMultipleStations(world *World, centerX, centerY float64, count int, seed int64, params procgen.GenerationParams) ([]*Entity, error) {
	if world == nil {
		return nil, fmt.Errorf("world cannot be nil")
	}
	if count <= 0 {
		return []*Entity{}, nil
	}

	stations := make([]*Entity, 0, count)

	for i := 0; i < count; i++ {
		// Calculate position (grid arrangement)
		x := centerX + float64(i%3-1)*64.0
		y := centerY + float64(i/3)*64.0

		// Create station with unique seed
		stationSeed := seed + int64(i)*1000
		station, err := SpawnMiniGameStation(world, x, y, stationSeed, params)
		if err != nil {
			return stations, fmt.Errorf("failed to spawn station %d: %w", i, err)
		}

		stations = append(stations, station)
	}

	return stations, nil
}

// selectGameType chooses a mini-game type based on genre and RNG.
// Genre influences which games are more common.
func selectGameType(rng *rand.Rand, genreID string) MiniGameType {
	// Genre-specific game distribution
	switch genreID {
	case "fantasy":
		// Fantasy: card games, dice, puzzles, ritual
		roll := rng.Float64()
		if roll < 0.3 {
			return MiniGameCard
		} else if roll < 0.5 {
			return MiniGameDice
		} else if roll < 0.7 {
			return MiniGamePuzzle
		} else if roll < 0.9 {
			return MiniGameMemory
		}
		return MiniGameRitual

	case "sci-fi":
		// Sci-fi: hacking, puzzles, memory
		roll := rng.Float64()
		if roll < 0.5 {
			return MiniGameHacking
		} else if roll < 0.75 {
			return MiniGamePuzzle
		}
		return MiniGameMemory

	case "horror":
		// Horror: ritual, lock-picking, memory (scary games)
		roll := rng.Float64()
		if roll < 0.4 {
			return MiniGameRitual
		} else if roll < 0.7 {
			return MiniGameLockPicking
		}
		return MiniGameMemory

	case "cyberpunk":
		// Cyberpunk: hacking, card, dice
		roll := rng.Float64()
		if roll < 0.4 {
			return MiniGameHacking
		} else if roll < 0.7 {
			return MiniGameCard
		}
		return MiniGameDice

	case "post-apocalyptic":
		// Post-apocalyptic: dice, lock-picking, puzzle
		roll := rng.Float64()
		if roll < 0.4 {
			return MiniGameDice
		} else if roll < 0.7 {
			return MiniGameLockPicking
		}
		return MiniGamePuzzle

	default:
		// Default: equal distribution
		return MiniGameType(rng.Intn(7))
	}
}

// calculateStationDifficulty determines station difficulty based on world depth.
// Higher depth = harder stations with better rewards.
func calculateStationDifficulty(depth int, baseDifficulty float64) float64 {
	// Base difficulty from params
	difficulty := baseDifficulty

	// Scale with depth (each level adds 5% difficulty, cap at 1.0)
	depthBonus := float64(depth) * 0.05
	difficulty += depthBonus

	// Clamp to valid range
	if difficulty < 0.0 {
		difficulty = 0.0
	}
	if difficulty > 1.0 {
		difficulty = 1.0
	}

	return difficulty
}

// GetMiniGameStationPrompt returns a formatted interaction prompt for a mini-game station.
// Includes cost and level requirements if applicable.
func GetMiniGameStationPrompt(station *Entity) string {
	stationCompRaw, ok := station.GetComponent("minigameStation")
	if !ok {
		return "Play Game"
	}

	stationComp, ok := stationCompRaw.(*MiniGameStationComponent)
	if !ok {
		return "Play Game"
	}

	prompt := fmt.Sprintf("Play %s", stationComp.GameType.String())

	// Add cost if applicable
	if stationComp.EntryCost > 0 {
		prompt += fmt.Sprintf(" (%dg)", stationComp.EntryCost)
	}

	// Add level requirement if applicable
	if stationComp.RequiresLevel > 0 {
		prompt += fmt.Sprintf(" [Lvl %d+]", stationComp.RequiresLevel)
	}

	return prompt
}
