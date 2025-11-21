package games

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetReportCaller(true)
	log.SetLevel(logrus.DebugLevel)
}

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
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
	}).Debug("Creating new dice game instance")

	game := &DiceGame{}

	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
	}).Debug("Dice game instance created")

	return game
}

// Initialize sets up the dice game with the given seed and difficulty.
func (d *DiceGame) Initialize(seed int64, difficulty float64) error {
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"seed":          seed,
		"difficulty":    difficulty,
	}).Debug("Initializing dice game")

	if difficulty < 0 || difficulty > 1.0 {
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"seed":          seed,
			"difficulty":    difficulty,
			"valid_range":   "0.0-1.0",
		}).Error("Difficulty validation failed")
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

	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"seed":          seed,
		"difficulty":    difficulty,
		"num_dice":      d.numDice,
		"dice_sides":    d.diceSides,
		"target_rolls":  d.targetRolls,
		"bet_amount":    d.betAmount,
	}).Info("Dice game initialized successfully")

	return nil
}

// Update advances the dice game state by playing one roll.
func (d *DiceGame) Update(deltaTime float64) error {
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"delta_time":    deltaTime,
		"completed":     d.completed,
		"player_wins":   d.playerWins,
		"opponent_wins": d.opponentWins,
		"target_rolls":  d.targetRolls,
	}).Debug("Update called")

	if d.completed {
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
		}).Debug("Game already completed, skipping update")
		return nil
	}

	// Roll dice for both players
	playerRoll := d.rollDice()
	opponentRoll := d.rollDice()

	// AI gets difficulty-based bonus
	aiBonus := int(d.difficulty * float64(d.numDice))
	opponentRoll += aiBonus

	log.WithFields(logrus.Fields{
		"minigame_type":  "dice",
		"player_roll":    playerRoll,
		"opponent_roll":  opponentRoll - aiBonus,
		"ai_bonus":       aiBonus,
		"final_opponent": opponentRoll,
		"difficulty":     d.difficulty,
	}).Debug("Dice rolled for current round")

	// Determine round winner
	previousPlayerWins := d.playerWins
	previousOpponentWins := d.opponentWins

	if playerRoll > opponentRoll {
		d.playerWins++
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"winner":        "player",
			"player_roll":   playerRoll,
			"opponent_roll": opponentRoll,
			"player_wins":   d.playerWins,
		}).Debug("Player wins round")
	} else if opponentRoll > playerRoll {
		d.opponentWins++
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"winner":        "opponent",
			"player_roll":   playerRoll,
			"opponent_roll": opponentRoll,
			"opponent_wins": d.opponentWins,
		}).Debug("Opponent wins round")
	} else {
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"result":        "tie",
			"roll_value":    playerRoll,
		}).Debug("Round tied, no winner")
	}

	// Check for game completion
	if d.playerWins >= d.targetRolls {
		d.completed = true
		d.playerWon = true
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"winner":        "player",
			"player_wins":   d.playerWins,
			"opponent_wins": d.opponentWins,
			"target_rolls":  d.targetRolls,
		}).Info("Dice game completed - player won")
	} else if d.opponentWins >= d.targetRolls {
		d.completed = true
		d.playerWon = false
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"winner":        "opponent",
			"player_wins":   d.playerWins,
			"opponent_wins": d.opponentWins,
			"target_rolls":  d.targetRolls,
		}).Info("Dice game completed - opponent won")
	}

	log.WithFields(logrus.Fields{
		"minigame_type":         "dice",
		"state_changed":         (previousPlayerWins != d.playerWins || previousOpponentWins != d.opponentWins),
		"game_completed":        d.completed,
		"current_player_wins":   d.playerWins,
		"current_opponent_wins": d.opponentWins,
	}).Debug("Update completed")

	return nil
}

// rollDice rolls the configured number of dice and returns the sum.
func (d *DiceGame) rollDice() int {
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"num_dice":      d.numDice,
		"dice_sides":    d.diceSides,
	}).Debug("Rolling dice")

	sum := 0
	for i := 0; i < d.numDice; i++ {
		roll := d.rng.Intn(d.diceSides) + 1
		sum += roll
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"die_index":     i,
			"roll_value":    roll,
			"running_sum":   sum,
		}).Debug("Individual die rolled")
	}

	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"num_dice":      d.numDice,
		"total_sum":     sum,
	}).Debug("Dice roll completed")

	return sum
}

// Render draws the dice game to the screen.
func (d *DiceGame) Render(screen engine.ImageProvider) error {
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"completed":     d.completed,
		"player_wins":   d.playerWins,
		"opponent_wins": d.opponentWins,
	}).Debug("Render called")

	// Minimal implementation - actual rendering in Phase 27.3
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
	}).Debug("Render completed (minimal implementation)")

	return nil
}

// IsComplete returns true when the game has finished.
func (d *DiceGame) IsComplete() bool {
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"completed":     d.completed,
	}).Debug("IsComplete checked")

	return d.completed
}

// GetReward returns the reward for winning the dice game.
func (d *DiceGame) GetReward() *engine.Reward {
	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"completed":     d.completed,
		"player_won":    d.playerWon,
	}).Debug("GetReward called")

	if !d.completed || !d.playerWon {
		log.WithFields(logrus.Fields{
			"minigame_type": "dice",
			"completed":     d.completed,
			"player_won":    d.playerWon,
			"reason":        "game not completed or player did not win",
		}).Debug("No reward - game not won by player")
		return nil
	}

	// Reward includes bet winnings
	goldReward := d.betAmount * 2
	xpReward := 15.0 + (d.difficulty * 30.0)

	log.WithFields(logrus.Fields{
		"minigame_type": "dice",
		"gold_reward":   goldReward,
		"xp_reward":     xpReward,
		"bet_amount":    d.betAmount,
		"difficulty":    d.difficulty,
	}).Info("Reward calculated for winning dice game")

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
