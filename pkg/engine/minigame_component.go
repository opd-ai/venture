package engine

// MiniGameType represents types of mini-games
type MiniGameType int

const (
	MiniGameCard MiniGameType = iota
	MiniGameDice
	MiniGamePuzzle
	MiniGameMemory
	MiniGameLockPicking
	MiniGameHacking
	MiniGameRitual
)

// Reward represents mini-game rewards
type Reward struct {
	Gold  int
	XP    float64
	Items []uint64 // Item entity IDs
}

// MiniGameComponent tracks mini-game state
type MiniGameComponent struct {
	GameType    MiniGameType
	Active      bool
	State       interface{} // Game-specific state
	Difficulty  float64
	TimeLimit   float64
	TimeElapsed float64
	Reward      *Reward
}

// Type returns the component type
func (m MiniGameComponent) Type() string {
	return "minigame"
}
