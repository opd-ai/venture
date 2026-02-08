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
	// LastRender contains the most recent render output for external consumption.
	LastRender *RenderOutput
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

// PrepareRender computes the current visual state for the memory game.
// This validates screen dimensions and populates internal render data.
// Implements engine.MiniGame interface.
func (m *MemoryGame) PrepareRender(screenWidth, screenHeight int) error {
	if err := validateScreenDimensions(screenWidth, screenHeight); err != nil {
		return fmt.Errorf("memory game prepare render: %w", err)
	}
	if m.rng == nil {
		return fmt.Errorf("memory game prepare render: game not initialized")
	}

	status := m.determineGameStatus()
	matchedCount := m.countMatched()
	elements := m.buildRenderElements(screenWidth, screenHeight, matchedCount)

	m.LastRender = &RenderOutput{
		Title:    "Memory Game",
		Status:   status,
		Width:    screenWidth,
		Height:   screenHeight,
		Elements: elements,
	}
	return nil
}

// GetRenderOutput returns the computed visual state from the last PrepareRender call.
// Implements engine.MiniGame interface.
func (m *MemoryGame) GetRenderOutput() engine.MiniGameRenderOutput {
	return m.LastRender
}

// Render draws the memory game to the screen.
// Computes visual state including card pairs, matched status, and attempt counter.
// DEPRECATED: Use PrepareRender() and GetRenderOutput() instead.
// Kept for backward compatibility with existing code.
func (m *MemoryGame) Render(screen engine.ImageProvider) error {
	w, h, err := validateScreen(screen)
	if err != nil {
		return fmt.Errorf("memory game render: %w", err)
	}
	return m.PrepareRender(w, h)
}

// determineGameStatus returns the current status string for the game.
func (m *MemoryGame) determineGameStatus() string {
	if m.completed {
		if m.playerWon {
			return "Won"
		}
		return "Lost"
	}
	return "Playing"
}

// countMatched returns the number of matched pairs.
func (m *MemoryGame) countMatched() int {
	count := 0
	for _, v := range m.matched {
		if v {
			count++
		}
	}
	return count
}

// buildRenderElements creates all visual elements for the memory game.
func (m *MemoryGame) buildRenderElements(w, h, matchedCount int) []RenderElement {
	elements := make([]RenderElement, 0, 3+m.numPairs)
	elements = append(elements, m.createHeaderElement(w, matchedCount))
	elements = append(elements, m.createCardElements(w)...)
	elements = append(elements, m.createProgressBar(w, h, matchedCount))
	return elements
}

// createHeaderElement creates the header text element showing game status.
func (m *MemoryGame) createHeaderElement(w, matchedCount int) RenderElement {
	mode := "Pairs"
	if m.sequenceMode {
		mode = "Sequence"
	}
	return RenderElement{
		Type:  "text",
		X:     w / 2,
		Y:     10,
		Label: fmt.Sprintf("Mode: %s  Attempts: %d / %d  Matched: %d / %d", mode, m.attempts, m.maxAttempts, matchedCount, m.numPairs),
	}
}

// createCardElements creates render elements for all card pairs.
func (m *MemoryGame) createCardElements(w int) []RenderElement {
	cols := m.calculateColumnCount()
	cardW := (w - 40) / cols
	cardH := 60

	cards := make([]RenderElement, 0, m.numPairs)
	for i := 0; i < m.numPairs; i++ {
		col := i % cols
		row := i / cols
		cards = append(cards, RenderElement{
			Type:        "card",
			X:           20 + col*cardW,
			Y:           50 + row*(cardH+10),
			W:           cardW - 5,
			H:           cardH,
			Label:       fmt.Sprintf("Pair %d", i+1),
			Value:       float64(i + 1),
			Highlighted: m.matched[i],
		})
	}
	return cards
}

// calculateColumnCount determines the number of columns for card layout.
func (m *MemoryGame) calculateColumnCount() int {
	if m.numPairs > 8 {
		return 6
	}
	return 4
}

// createProgressBar creates the progress bar element.
func (m *MemoryGame) createProgressBar(w, h, matchedCount int) RenderElement {
	progress := float64(matchedCount) / float64(m.numPairs)
	return RenderElement{
		Type:  "progress",
		X:     20,
		Y:     h - 40,
		W:     w - 40,
		H:     20,
		Label: "Match Progress",
		Value: progress,
	}
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
