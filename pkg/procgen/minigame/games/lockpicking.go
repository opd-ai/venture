package games

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

// LockPickingGame implements timing-based lock-picking challenges.
// Players must align pins within timing windows to unlock.
// Game duration: 0.5-2 minutes typical.
//
// Phase 27.2: Mini-Game Types
type LockPickingGame struct {
	rng           *rand.Rand
	difficulty    float64
	numPins       int
	timingWindow  float64
	maxFailures   int
	failures      int
	currentPinIdx int
	pinPositions  []float64
	timeElapsed   float64
	completed     bool
	playerWon     bool
}

// NewLockPickingGame creates a new lock-picking game instance.
func NewLockPickingGame() *LockPickingGame {
	return &LockPickingGame{}
}

// Initialize sets up the lock-picking game with the given seed and difficulty.
func (l *LockPickingGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	l.rng = rand.New(rand.NewSource(seed))
	l.difficulty = difficulty

	// Scale parameters based on difficulty
	// Easy: 3 pins, 0.5s window, 5 failures allowed
	// Hard: 7 pins, 0.1s window, 2 failures allowed
	l.numPins = 3 + int(difficulty*4)
	l.timingWindow = 0.5 - difficulty*0.4
	l.maxFailures = 5 - int(difficulty*3)

	// Generate random pin positions (0.0 to 1.0)
	l.pinPositions = make([]float64, l.numPins)
	for i := range l.pinPositions {
		l.pinPositions[i] = l.rng.Float64()
	}

	l.completed = false
	l.playerWon = false
	l.failures = 0
	l.currentPinIdx = 0
	l.timeElapsed = 0

	return nil
}

// Update advances the lock-picking game by attempting to pick the current pin.
func (l *LockPickingGame) Update(deltaTime float64) error {
	if l.completed {
		return nil
	}

	l.timeElapsed += deltaTime

	// Simulate player attempting to pick pin
	// Success rate depends on difficulty and timing
	successRate := 0.7 - l.difficulty*0.4
	if l.rng.Float64() < successRate {
		// Successfully picked pin
		l.currentPinIdx++

		// Check if all pins picked
		if l.currentPinIdx >= l.numPins {
			l.completed = true
			l.playerWon = true
		}
	} else {
		// Failed attempt
		l.failures++

		// Check if too many failures
		if l.failures >= l.maxFailures {
			l.completed = true
			l.playerWon = false
		}
	}

	return nil
}

// Render draws the lock-picking game to the screen.
func (l *LockPickingGame) Render(screen engine.ImageProvider) error {
	// Minimal implementation - actual rendering in Phase 27.3
	return nil
}

// IsComplete returns true when the game has finished.
func (l *LockPickingGame) IsComplete() bool {
	return l.completed
}

// GetReward returns the reward for successfully picking the lock.
func (l *LockPickingGame) GetReward() *engine.Reward {
	if !l.completed || !l.playerWon {
		return nil
	}

	// Reward scales with difficulty and speed
	timeBonus := 1.0
	if l.timeElapsed < 10.0 {
		timeBonus = 2.0 - l.timeElapsed/10.0
	}
	goldReward := int(30.0 * timeBonus * (1.0 + l.difficulty))
	xpReward := 15.0 + (l.difficulty * 30.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
