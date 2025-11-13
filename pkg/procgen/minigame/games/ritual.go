package games

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

// RitualGame implements spell pattern drawing for fantasy/horror genres.
// Players must draw symbols accurately to complete the ritual.
// Game duration: 2-5 minutes typical.
//
// Phase 27.2: Mini-Game Types
type RitualGame struct {
	rng           *rand.Rand
	difficulty    float64
	numSymbols    int
	drawAccuracy  float64
	currentSymbol int
	symbols       []Symbol
	completed     bool
	playerWon     bool
}

// Symbol represents a pattern to draw in the ritual.
type Symbol struct {
	Name   string
	Points []Point
}

// Point represents a 2D coordinate.
type Point struct {
	X, Y float64
}

// NewRitualGame creates a new ritual game instance.
func NewRitualGame() *RitualGame {
	return &RitualGame{}
}

// Initialize sets up the ritual game with the given seed and difficulty.
func (r *RitualGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	r.rng = rand.New(rand.NewSource(seed))
	r.difficulty = difficulty

	// Scale parameters based on difficulty
	// Easy: 3 symbols, 0.7 accuracy required
	// Hard: 7 symbols, 0.95 accuracy required
	r.numSymbols = 3 + int(difficulty*4)
	r.drawAccuracy = 0.7 + difficulty*0.25

	// Generate ritual symbols
	r.symbols = r.generateSymbols()

	r.completed = false
	r.playerWon = false
	r.currentSymbol = 0

	return nil
}

// generateSymbols creates a set of procedural symbols for the ritual.
func (r *RitualGame) generateSymbols() []Symbol {
	symbols := make([]Symbol, r.numSymbols)
	symbolTypes := []string{"Circle", "Star", "Triangle", "Spiral", "Pentagram", "Cross", "Rune"}

	for i := 0; i < r.numSymbols; i++ {
		symbolType := symbolTypes[r.rng.Intn(len(symbolTypes))]
		symbols[i] = r.createSymbol(symbolType, i)
	}

	return symbols
}

// createSymbol generates points for a specific symbol type.
func (r *RitualGame) createSymbol(symbolType string, seed int) Symbol {
	// Use seed to generate deterministic but varied symbol
	localRng := rand.New(rand.NewSource(int64(seed) + r.rng.Int63()))
	numPoints := 8 + localRng.Intn(8)
	points := make([]Point, numPoints)

	// Generate points based on symbol type
	switch symbolType {
	case "Circle":
		for i := 0; i < numPoints; i++ {
			angle := 2 * math.Pi * float64(i) / float64(numPoints)
			points[i] = Point{
				X: 0.5 + 0.4*math.Cos(angle),
				Y: 0.5 + 0.4*math.Sin(angle),
			}
		}
	case "Star":
		for i := 0; i < numPoints; i++ {
			angle := 2 * math.Pi * float64(i) / float64(numPoints)
			radius := 0.4
			if i%2 == 1 {
				radius = 0.2
			}
			points[i] = Point{
				X: 0.5 + radius*math.Cos(angle),
				Y: 0.5 + radius*math.Sin(angle),
			}
		}
	default:
		// Generic procedural pattern
		for i := 0; i < numPoints; i++ {
			points[i] = Point{
				X: localRng.Float64(),
				Y: localRng.Float64(),
			}
		}
	}

	return Symbol{
		Name:   symbolType,
		Points: points,
	}
}

// Update advances the ritual game by drawing the current symbol.
func (r *RitualGame) Update(deltaTime float64) error {
	if r.completed {
		return nil
	}

	// Simulate player drawing symbol
	successRate := 0.6 + (1.0-r.difficulty)*0.3
	if r.rng.Float64() < successRate {
		// Successfully drew symbol
		r.currentSymbol++

		// Check if all symbols drawn
		if r.currentSymbol >= r.numSymbols {
			r.completed = true
			r.playerWon = true
		}
	} else {
		// Failed to draw accurately enough
		r.completed = true
		r.playerWon = false
	}

	return nil
}

// Render draws the ritual game to the screen.
func (r *RitualGame) Render(screen engine.ImageProvider) error {
	// Minimal implementation - actual rendering in Phase 27.3
	return nil
}

// IsComplete returns true when the game has finished.
func (r *RitualGame) IsComplete() bool {
	return r.completed
}

// GetReward returns the reward for completing the ritual.
func (r *RitualGame) GetReward() *engine.Reward {
	if !r.completed || !r.playerWon {
		return nil
	}

	// Ritual completion grants significant rewards
	goldReward := 60 + int(r.difficulty*120)
	xpReward := 30.0 + (r.difficulty * 60.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
