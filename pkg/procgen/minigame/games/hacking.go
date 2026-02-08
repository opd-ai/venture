// HackingGame implements terminal/console code-breaking puzzles for sci-fi themes.
// This file contains the HackingGame implementation where players must
// deduce an alphanumeric code using feedback hints.
package games

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/engine"
)

// HackingGame implements terminal/console puzzles for sci-fi genre.
// Players must guess or deduce a code within attempt limits.
// Game duration: 1-3 minutes typical.
//
// Phase 27.2: Mini-Game Types
type HackingGame struct {
	rng         *rand.Rand
	difficulty  float64
	codeLength  int
	maxAttempts int
	attempts    int
	code        string
	guesses     []string
	hints       []string
	completed   bool
	playerWon   bool
	// LastRender contains the most recent render output for external consumption.
	LastRender *RenderOutput
}

// NewHackingGame creates a new hacking game instance.
func NewHackingGame() *HackingGame {
	return &HackingGame{}
}

// Initialize sets up the hacking game with the given seed and difficulty.
func (h *HackingGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	h.rng = rand.New(rand.NewSource(seed))
	h.difficulty = difficulty

	// Scale parameters based on difficulty
	// Easy: 4 chars, 10 attempts
	// Hard: 8 chars, 6 attempts
	h.codeLength = 4 + int(difficulty*4)
	h.maxAttempts = 10 - int(difficulty*4)

	// Generate random code (alphanumeric)
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	codeBuilder := strings.Builder{}
	for i := 0; i < h.codeLength; i++ {
		codeBuilder.WriteByte(chars[h.rng.Intn(len(chars))])
	}
	h.code = codeBuilder.String()

	h.guesses = make([]string, 0)
	h.hints = make([]string, 0)
	h.completed = false
	h.playerWon = false
	h.attempts = 0

	return nil
}

// Update advances the hacking game by making a simulated guess.
func (h *HackingGame) Update(deltaTime float64) error {
	if h.completed {
		return nil
	}

	h.attempts++

	// Generate a guess (simulated player input)
	guess := h.generateGuess()
	h.guesses = append(h.guesses, guess)

	// Check if guess is correct
	if guess == h.code {
		h.completed = true
		h.playerWon = true
		return nil
	}

	// Generate hint based on guess similarity
	hint := h.generateHint(guess)
	h.hints = append(h.hints, hint)

	// Check if out of attempts
	if h.attempts >= h.maxAttempts {
		h.completed = true
		h.playerWon = false
	}

	return nil
}

// generateGuess creates a simulated player guess.
func (h *HackingGame) generateGuess() string {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	guessBuilder := strings.Builder{}

	// Gradually improve guesses based on attempt number
	correctChance := 0.1 * float64(h.attempts) / float64(h.maxAttempts)

	for i := 0; i < h.codeLength; i++ {
		if h.rng.Float64() < correctChance {
			// Sometimes guess correct character
			guessBuilder.WriteByte(h.code[i])
		} else {
			// Random character
			guessBuilder.WriteByte(chars[h.rng.Intn(len(chars))])
		}
	}

	return guessBuilder.String()
}

// generateHint creates feedback for a guess.
func (h *HackingGame) generateHint(guess string) string {
	correctPos := 0
	correctChar := 0

	for i := 0; i < len(guess) && i < len(h.code); i++ {
		if guess[i] == h.code[i] {
			correctPos++
		} else if strings.ContainsRune(h.code, rune(guess[i])) {
			correctChar++
		}
	}

	return fmt.Sprintf("%d correct, %d misplaced", correctPos, correctChar)
}

// PrepareRender computes the current visual state for the hacking game.
// This validates screen dimensions and populates internal render data.
// Implements engine.MiniGame interface.
func (hg *HackingGame) PrepareRender(screenWidth, screenHeight int) error {
	if err := validateScreenDimensions(screenWidth, screenHeight); err != nil {
		return fmt.Errorf("hacking game prepare render: %w", err)
	}
	if hg.rng == nil {
		return fmt.Errorf("hacking game prepare render: game not initialized")
	}

	status := "Playing"
	if hg.completed {
		if hg.playerWon {
			status = "Won"
		} else {
			status = "Lost"
		}
	}

	elements := make([]RenderElement, 0, 3+len(hg.guesses))

	// Header
	elements = append(elements, RenderElement{
		Type:  "text",
		X:     screenWidth / 2,
		Y:     10,
		Label: fmt.Sprintf("CODE LENGTH: %d  ATTEMPTS: %d / %d", hg.codeLength, hg.attempts, hg.maxAttempts),
	})

	// Terminal lines (guesses + hints)
	lineH := 20
	startY := 50
	for i := 0; i < len(hg.guesses); i++ {
		hint := ""
		if i < len(hg.hints) {
			hint = hg.hints[i]
		}
		elements = append(elements, RenderElement{
			Type:  "terminal",
			X:     20,
			Y:     startY + i*lineH,
			W:     screenWidth - 40,
			H:     lineH,
			Label: fmt.Sprintf("> %s  [%s]", hg.guesses[i], hint),
			Value: float64(i + 1),
		})
	}

	// Attempt progress
	progress := float64(hg.attempts) / float64(hg.maxAttempts)
	elements = append(elements, RenderElement{
		Type:  "progress",
		X:     20,
		Y:     screenHeight - 40,
		W:     screenWidth - 40,
		H:     20,
		Label: "Attempts Used",
		Value: progress,
	})

	hg.LastRender = &RenderOutput{
		Title:    "Hacking",
		Status:   status,
		Width:    screenWidth,
		Height:   screenHeight,
		Elements: elements,
	}
	return nil
}

// GetRenderOutput returns the computed visual state from the last PrepareRender call.
// Implements engine.MiniGame interface.
func (hg *HackingGame) GetRenderOutput() engine.MiniGameRenderOutput {
	return hg.LastRender
}

// Render draws the hacking game to the screen.
// Computes visual state including terminal display with guesses and hints.
// DEPRECATED: Use PrepareRender() and GetRenderOutput() instead.
// Kept for backward compatibility with existing code.
func (hg *HackingGame) Render(screen engine.ImageProvider) error {
	w, h, err := validateScreen(screen)
	if err != nil {
		return fmt.Errorf("hacking game render: %w", err)
	}
	return hg.PrepareRender(w, h)
}

// IsComplete returns true when the game has finished.
func (h *HackingGame) IsComplete() bool {
	return h.completed
}

// GetReward returns the reward for successfully hacking the code.
func (h *HackingGame) GetReward() *engine.Reward {
	if !h.completed || !h.playerWon {
		return nil
	}

	// Reward scales with difficulty and speed
	attemptsLeft := h.maxAttempts - h.attempts
	bonusMultiplier := 1.0 + float64(attemptsLeft)/float64(h.maxAttempts)
	goldReward := int(45.0 * bonusMultiplier * (1.0 + h.difficulty))
	xpReward := 22.0 + (h.difficulty * 45.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
