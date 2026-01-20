// MemoryGame implements card pair matching and sequence repetition games.
// This file contains the MemoryGame implementation where players must
// match pairs within a limited number of attempts.
package games

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

// MemoryGame implements card pairs and sequence repetition games.
// Players must match pairs or repeat sequences within attempt limits.
// Game duration: 2-4 minutes typical.
//
// Phase 27.2: Mini-Game Types
type MemoryGame struct {
	rng          *rand.Rand
	difficulty   float64
	numPairs     int
	maxAttempts  int
	attempts     int
	matched      []bool
	completed    bool
	playerWon    bool
	sequenceMode bool
}

// NewMemoryGame creates a new memory game instance.
func NewMemoryGame() *MemoryGame {
	return &MemoryGame{}
}

// Initialize sets up the memory game with the given seed and difficulty.
func (m *MemoryGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	m.rng = rand.New(rand.NewSource(seed))
	m.difficulty = difficulty

	// Scale parameters based on difficulty
	// Easy: 4 pairs, 20 attempts
	// Hard: 12 pairs, 30 attempts
	m.numPairs = 4 + int(difficulty*8)
	m.maxAttempts = 20 + int(difficulty*10)

	// Sequence mode for harder difficulties
	m.sequenceMode = difficulty > 0.6

	m.matched = make([]bool, m.numPairs)
	m.completed = false
	m.playerWon = false
	m.attempts = 0

	return nil
}

// Update advances the memory game by making a simulated match attempt.
func (m *MemoryGame) Update(deltaTime float64) error {
	if m.completed {
		return nil
	}

	m.attempts++

	// Simulate player attempting to match pairs
	successRate := 0.6 - m.difficulty*0.3 // Harder games = lower success rate

	if m.rng.Float64() < successRate {
		// Find unmatched pair and match it
		for i := range m.matched {
			if !m.matched[i] {
				m.matched[i] = true
				break
			}
		}
	}

	// Check for completion
	allMatched := true
	for _, matched := range m.matched {
		if !matched {
			allMatched = false
			break
		}
	}

	if allMatched {
		m.completed = true
		m.playerWon = true
	} else if m.attempts >= m.maxAttempts {
		m.completed = true
		m.playerWon = false
	}

	return nil
}

// Render draws the memory game to the screen.
func (m *MemoryGame) Render(screen engine.ImageProvider) error {
	// Minimal implementation - actual rendering in Phase 27.3
	return nil
}

// IsComplete returns true when the game has finished.
func (m *MemoryGame) IsComplete() bool {
	return m.completed
}

// GetReward returns the reward for completing the memory game.
func (m *MemoryGame) GetReward() *engine.Reward {
	if !m.completed || !m.playerWon {
		return nil
	}

	// Reward scales with difficulty and remaining attempts
	attemptsLeft := m.maxAttempts - m.attempts
	bonusMultiplier := 1.0 + float64(attemptsLeft)/float64(m.maxAttempts)
	goldReward := int(35.0 * bonusMultiplier * (1.0 + m.difficulty))
	xpReward := 18.0 + (m.difficulty * 35.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
