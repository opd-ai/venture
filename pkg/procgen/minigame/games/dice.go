// DiceGame implements custom dice game with betting mechanics.
// This file contains the DiceGame implementation where players compete
// against an AI opponent by rolling multiple dice per round.
package games

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

// DiceGame implements a custom dice game with betting mechanics.
// Players bet on dice rolls and try to beat the opponent.
// Game duration: 2-5 minutes typical.
//
// Phase 27.2: Mini-Game Types
type DiceGame struct {
	rng          *rand.Rand
	difficulty   float64
	numDice      int
	diceSides    int
	betAmount    int
	targetRolls  int
	playerWins   int
	opponentWins int
	completed    bool
	playerWon    bool
}

// NewDiceGame creates a new dice game instance.
func NewDiceGame() *DiceGame {
	return &DiceGame{}
}

// Initialize sets up the dice game with the given seed and difficulty.
func (d *DiceGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	d.rng = rand.New(rand.NewSource(seed))
	d.difficulty = difficulty

	// Scale game parameters based on difficulty
	// Easy: 2 dice, 6 sides, 5 target rolls
	// Hard: 5 dice, 12 sides, 10 target rolls
	d.numDice = 2 + int(difficulty*3)
	d.diceSides = 6 + int(difficulty*6)
	d.targetRolls = 5 + int(difficulty*5)
	d.betAmount = 10 + int(difficulty*40)

	d.completed = false
	d.playerWon = false
	d.playerWins = 0
	d.opponentWins = 0

	return nil
}

// Update advances the dice game state by playing one roll.
func (d *DiceGame) Update(deltaTime float64) error {
	if d.completed {
		return nil
	}

	// Roll dice for both players
	playerRoll := d.rollDice()
	opponentRoll := d.rollDice()

	// AI gets difficulty-based bonus
	aiBonus := int(d.difficulty * float64(d.numDice))
	opponentRoll += aiBonus

	// Determine round winner
	if playerRoll > opponentRoll {
		d.playerWins++
	} else if opponentRoll > playerRoll {
		d.opponentWins++
	}

	// Check for game completion
	if d.playerWins >= d.targetRolls {
		d.completed = true
		d.playerWon = true
	} else if d.opponentWins >= d.targetRolls {
		d.completed = true
		d.playerWon = false
	}

	return nil
}

// rollDice rolls the configured number of dice and returns the sum.
func (d *DiceGame) rollDice() int {
	sum := 0
	for i := 0; i < d.numDice; i++ {
		sum += d.rng.Intn(d.diceSides) + 1
	}
	return sum
}

// Render draws the dice game to the screen.
func (d *DiceGame) Render(screen engine.ImageProvider) error {
	// Minimal implementation - actual rendering in Phase 27.3
	return nil
}

// IsComplete returns true when the game has finished.
func (d *DiceGame) IsComplete() bool {
	return d.completed
}

// GetReward returns the reward for winning the dice game.
func (d *DiceGame) GetReward() *engine.Reward {
	if !d.completed || !d.playerWon {
		return nil
	}

	// Reward includes bet winnings
	goldReward := d.betAmount * 2
	xpReward := 15.0 + (d.difficulty * 30.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
