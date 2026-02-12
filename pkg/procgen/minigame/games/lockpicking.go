// LockPickingGame implements timing-based lock-picking challenges.
// This file contains the LockPickingGame implementation where players must
// align pins within timing windows to successfully pick a lock.
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
	// LastRender contains the most recent render output for external consumption.
	LastRender *RenderOutput
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

// PrepareRender computes the current visual state for the lock-picking game.
// This validates screen dimensions and populates internal render data.
// Implements engine.MiniGame interface.
func (l *LockPickingGame) PrepareRender(screenWidth, screenHeight int) error {
	if err := validateScreenDimensions(screenWidth, screenHeight); err != nil {
		return fmt.Errorf("lockpicking game prepare render: %w", err)
	}
	if l.rng == nil {
		return fmt.Errorf("lockpicking game prepare render: game not initialized")
	}

	status := "Playing"
	if l.completed {
		if l.playerWon {
			status = "Won"
		} else {
			status = "Lost"
		}
	}

	elements := make([]RenderElement, 0, 3+l.numPins)

	// Status display
	elements = append(elements, RenderElement{
		Type:  "text",
		X:     screenWidth / 2,
		Y:     10,
		Label: fmt.Sprintf("Pin %d / %d  Failures: %d / %d  Time: %.1fs", l.currentPinIdx, l.numPins, l.failures, l.maxFailures, l.timeElapsed),
	})

	// Pin display
	pinW := (screenWidth - 40) / l.numPins
	if pinW > 80 {
		pinW = 80
	}
	pinH := screenHeight / 2
	startX := (screenWidth - l.numPins*pinW) / 2
	for i := 0; i < l.numPins; i++ {
		picked := i < l.currentPinIdx
		active := i == l.currentPinIdx
		elements = append(elements, RenderElement{
			Type:        "pin",
			X:           startX + i*pinW,
			Y:           screenHeight/4 + int(l.pinPositions[i]*float64(pinH)/2),
			W:           pinW - 5,
			H:           20,
			Label:       fmt.Sprintf("Pin %d", i+1),
			Value:       l.pinPositions[i],
			Highlighted: picked || active,
		})
	}

	// Timing window indicator
	elements = append(elements, RenderElement{
		Type:  "progress",
		X:     20,
		Y:     screenHeight - 40,
		W:     screenWidth - 40,
		H:     20,
		Label: fmt.Sprintf("Timing Window: %.2fs", l.timingWindow),
		Value: l.timingWindow,
	})

	l.LastRender = &RenderOutput{
		Title:    "Lock Picking",
		Status:   status,
		Width:    screenWidth,
		Height:   screenHeight,
		Elements: elements,
	}
	return nil
}

// GetRenderOutput returns the computed visual state from the last PrepareRender call.
// Implements engine.MiniGame interface.
func (l *LockPickingGame) GetRenderOutput() engine.MiniGameRenderOutput {
	if l.LastRender == nil {
		return nil
	}
	return l.LastRender
}

// Render draws the lock-picking game to the screen.
// Computes visual state including pin positions, timing window, and failure count.
// DEPRECATED: Use PrepareRender() and GetRenderOutput() instead.
// Kept for backward compatibility with existing code.
func (l *LockPickingGame) Render(screen engine.ImageProvider) error {
	w, h, err := validateScreen(screen)
	if err != nil {
		return fmt.Errorf("lockpicking game render: %w", err)
	}
	return l.PrepareRender(w, h)
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
