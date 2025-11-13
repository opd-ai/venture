// Package games contains implementations of mini-game types for Phase 27.2.
// Each game implements the engine.MiniGame interface with deterministic
// gameplay, genre-appropriate theming, and difficulty scaling.
package games

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

// CardGame implements a procedural card game with deck and AI opponent.
// Players compete in rounds trying to reach a target number of wins.
// Game duration: 5-10 minutes typical.
//
// Phase 27.2: Mini-Game Types
type CardGame struct {
	rng          *rand.Rand
	difficulty   float64
	deckSize     int
	handSize     int
	targetWins   int
	playerWins   int
	opponentWins int
	currentRound int
	completed    bool
	playerWon    bool
	playerHand   []int // Card values
	opponentHand []int
	deck         []int
}

// NewCardGame creates a new card game instance.
func NewCardGame() *CardGame {
	return &CardGame{}
}

// Initialize sets up the card game with the given seed and difficulty.
func (c *CardGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	c.rng = rand.New(rand.NewSource(seed))
	c.difficulty = difficulty

	// Scale game parameters based on difficulty
	// Easy: 20 card deck, 5 card hand, 3 wins needed
	// Hard: 52 card deck, 7 card hand, 5 wins needed
	c.deckSize = 20 + int(difficulty*32)
	c.handSize = 5 + int(difficulty*2)
	c.targetWins = 3 + int(difficulty*2)

	// Initialize deck with values 1-10
	c.deck = make([]int, 0, c.deckSize)
	for i := 0; i < c.deckSize; i++ {
		c.deck = append(c.deck, (i%10)+1)
	}

	// Shuffle deck deterministically
	c.rng.Shuffle(len(c.deck), func(i, j int) {
		c.deck[i], c.deck[j] = c.deck[j], c.deck[i]
	})

	// Deal initial hands
	c.dealHands()

	c.completed = false
	c.playerWon = false
	c.playerWins = 0
	c.opponentWins = 0
	c.currentRound = 0

	return nil
}

// dealHands deals cards from the deck to both players.
func (c *CardGame) dealHands() {
	if len(c.deck) < c.handSize*2 {
		// Reshuffle deck if needed
		c.deck = make([]int, 0, c.deckSize)
		for i := 0; i < c.deckSize; i++ {
			c.deck = append(c.deck, (i%10)+1)
		}
		c.rng.Shuffle(len(c.deck), func(i, j int) {
			c.deck[i], c.deck[j] = c.deck[j], c.deck[i]
		})
	}

	c.playerHand = make([]int, c.handSize)
	c.opponentHand = make([]int, c.handSize)

	for i := 0; i < c.handSize; i++ {
		c.playerHand[i] = c.deck[i]
		c.opponentHand[i] = c.deck[i+c.handSize]
	}

	c.deck = c.deck[c.handSize*2:]
}

// Update advances the card game state.
// Automatically plays rounds with simple AI.
func (c *CardGame) Update(deltaTime float64) error {
	if c.completed {
		return nil
	}

	// Play one round per update (simulated gameplay)
	c.playRound()

	// Check for game completion
	if c.playerWins >= c.targetWins {
		c.completed = true
		c.playerWon = true
	} else if c.opponentWins >= c.targetWins {
		c.completed = true
		c.playerWon = false
	}

	return nil
}

// playRound plays one round of the card game.
func (c *CardGame) playRound() {
	c.currentRound++

	// Simple card comparison: sum of hand values
	playerSum := 0
	for _, card := range c.playerHand {
		playerSum += card
	}

	opponentSum := 0
	for _, card := range c.opponentHand {
		opponentSum += card
	}

	// Add AI difficulty adjustment: opponent gets advantage based on difficulty
	aiBonus := int(c.difficulty * 5)
	opponentSum += aiBonus

	// Determine round winner
	if playerSum > opponentSum {
		c.playerWins++
	} else if opponentSum > playerSum {
		c.opponentWins++
	}
	// Tie = no winner

	// Deal new hands for next round
	c.dealHands()
}

// Render draws the card game to the screen.
// For Phase 27.2, this is a minimal implementation (actual rendering in Phase 27.3).
func (c *CardGame) Render(screen engine.ImageProvider) error {
	// Minimal implementation - actual rendering happens in integration phase
	return nil
}

// IsComplete returns true when the game has finished.
func (c *CardGame) IsComplete() bool {
	return c.completed
}

// GetReward returns the reward for winning the card game.
func (c *CardGame) GetReward() *engine.Reward {
	if !c.completed || !c.playerWon {
		return nil
	}

	// Reward scales with difficulty
	goldReward := 50 + int(c.difficulty*100)
	xpReward := 25.0 + (c.difficulty * 50.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil, // No item rewards for basic card game
	}
}
