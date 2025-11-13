package minigame

// CardGameState represents the state of a card game
type CardGameState struct {
	DeckSize     int
	HandSize     int
	TargetWins   int
	PlayerWins   int
	OpponentWins int
	CurrentRound int
}

// DiceGameState represents the state of a dice game
type DiceGameState struct {
	NumDice      int
	DiceSides    int
	BetAmount    int
	TargetRoll   int
	PlayerRoll   int
	OpponentRoll int
}

// PuzzleGameState represents the state of a sliding puzzle
type PuzzleGameState struct {
	GridSize int
	Moves    int
	MaxMoves int
	Grid     [][]int // Puzzle grid state
}

// MemoryGameState represents the state of a memory game
type MemoryGameState struct {
	NumPairs     int
	MaxAttempts  int
	Attempts     int
	SequenceMode bool
	Matched      []bool // Tracks matched pairs
}

// LockPickingGameState represents the state of a lock-picking game
type LockPickingGameState struct {
	NumPins       int
	TimingWindow  float64
	MaxFailures   int
	Failures      int
	CurrentPinIdx int
	PinPositions  []float64 // Target positions for each pin
}

// HackingGameState represents the state of a hacking game
type HackingGameState struct {
	CodeLength  int
	MaxAttempts int
	Attempts    int
	Code        string   // Target code
	Hints       []string // Feedback from previous attempts
}

// RitualGameState represents the state of a ritual game
type RitualGameState struct {
	NumSymbols    int
	DrawAccuracy  float64
	CurrentSymbol int
	Completed     bool
	Symbols       []Symbol // Symbol patterns to draw
}

// Symbol represents a symbol pattern for ritual drawing
type Symbol struct {
	Name   string
	Points []Point // Points defining the symbol shape
}

// Point represents a 2D coordinate
type Point struct {
	X float64
	Y float64
}
