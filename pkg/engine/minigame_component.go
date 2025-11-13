// Package engine provides the core ECS (Entity-Component-System) framework.
// This file defines mini-game components for Phase 27.1: Mini-Game Framework.
package engine

// MiniGameType represents types of mini-games available in Venture.
//
// Phase 27.1: Mini-Game Framework
type MiniGameType int

const (
	// MiniGameCard represents a procedural card game with deck and AI opponent
	MiniGameCard MiniGameType = iota
	// MiniGameDice represents a custom dice game with betting mechanics
	MiniGameDice
	// MiniGamePuzzle represents sliding tiles or pattern matching puzzles
	MiniGamePuzzle
	// MiniGameMemory represents card pairs or sequence repetition games
	MiniGameMemory
	// MiniGameLockPicking represents timing-based lock-picking challenges
	MiniGameLockPicking
	// MiniGameHacking represents terminal/console puzzles (sci-fi genre)
	MiniGameHacking
	// MiniGameRitual represents spell pattern drawing (fantasy/horror genre)
	MiniGameRitual
)

// String returns the human-readable name for this mini-game type.
func (m MiniGameType) String() string {
	switch m {
	case MiniGameCard:
		return "Card Game"
	case MiniGameDice:
		return "Dice Game"
	case MiniGamePuzzle:
		return "Puzzle"
	case MiniGameMemory:
		return "Memory"
	case MiniGameLockPicking:
		return "Lock-Picking"
	case MiniGameHacking:
		return "Hacking"
	case MiniGameRitual:
		return "Ritual"
	default:
		return "Unknown"
	}
}

// Reward represents mini-game completion rewards.
// Rewards are awarded to the player upon successful mini-game completion.
//
// Phase 27.1: Mini-Game Framework
type Reward struct {
	// Gold is the amount of gold currency awarded
	Gold int
	// XP is the experience points awarded
	XP float64
	// Items contains entity IDs of items to be awarded
	Items []uint64
}

// MiniGameComponent tracks mini-game state for an entity.
// When attached to an entity, it indicates that entity is hosting or participating
// in a mini-game. The component stores the game type, state, and reward information.
//
// Phase 27.1: Mini-Game Framework
type MiniGameComponent struct {
	// GameType specifies which mini-game is being played
	GameType MiniGameType
	// Active indicates whether the mini-game is currently running
	Active bool
	// State stores game-specific state data (type depends on GameType)
	State interface{}
	// Difficulty affects game parameters (0.0-1.0 scale)
	Difficulty float64
	// TimeLimit is the maximum time allowed in seconds (0 = no limit)
	TimeLimit float64
	// TimeElapsed tracks how long the game has been running in seconds
	TimeElapsed float64
	// Reward is the reward that will be awarded upon successful completion
	Reward *Reward
	// GameInstance holds the MiniGame interface implementation for this game
	// This allows the system to call Update, Render, IsComplete, etc.
	GameInstance MiniGame
}

// Type returns the component type identifier "minigame".
func (m MiniGameComponent) Type() string {
	return "minigame"
}
