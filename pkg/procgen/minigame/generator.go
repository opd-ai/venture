package minigame

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// GameType represents different mini-game types
type GameType int

const (
	GameTypeCard GameType = iota
	GameTypeDice
	GameTypePuzzle
	GameTypeMemory
	GameTypeLockPicking
	GameTypeHacking
	GameTypeRitual
)

// String returns the string representation of GameType
func (g GameType) String() string {
	switch g {
	case GameTypeCard:
		return "Card"
	case GameTypeDice:
		return "Dice"
	case GameTypePuzzle:
		return "Puzzle"
	case GameTypeMemory:
		return "Memory"
	case GameTypeLockPicking:
		return "LockPicking"
	case GameTypeHacking:
		return "Hacking"
	case GameTypeRitual:
		return "Ritual"
	default:
		return "Unknown"
	}
}

// MiniGame represents a generated mini-game instance
type MiniGame struct {
	Type       GameType
	Name       string
	Difficulty float64
	TimeLimit  float64 // seconds
	Rules      string
	State      interface{}
	GenreID    string
}

// Generator implements procgen.Generator for mini-games
type Generator struct{}

// NewGenerator creates a new mini-game generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a new mini-game
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if params.Difficulty < 0 || params.Difficulty > 1.0 {
		return nil, fmt.Errorf("difficulty must be between 0 and 1, got %.2f", params.Difficulty)
	}

	rng := rand.New(rand.NewSource(seed))

	// Select game type based on genre
	gameType := g.selectGameType(rng, params.GenreID)

	var game *MiniGame
	switch gameType {
	case GameTypeCard:
		game = g.generateCardGame(rng, params)
	case GameTypeDice:
		game = g.generateDiceGame(rng, params)
	case GameTypePuzzle:
		game = g.generatePuzzleGame(rng, params)
	case GameTypeMemory:
		game = g.generateMemoryGame(rng, params)
	case GameTypeLockPicking:
		game = g.generateLockPickingGame(rng, params)
	case GameTypeHacking:
		game = g.generateHackingGame(rng, params)
	case GameTypeRitual:
		game = g.generateRitualGame(rng, params)
	default:
		return nil, fmt.Errorf("unknown game type: %d", gameType)
	}

	return game, nil
}

// Validate checks if a mini-game meets quality standards
func (g *Generator) Validate(result interface{}) error {
	game, ok := result.(*MiniGame)
	if !ok {
		return fmt.Errorf("result is not a *MiniGame")
	}

	if game.Name == "" {
		return fmt.Errorf("game name is empty")
	}

	if game.TimeLimit <= 0 {
		return fmt.Errorf("time limit must be positive, got %.2f", game.TimeLimit)
	}

	if game.Difficulty < 0 || game.Difficulty > 1.0 {
		return fmt.Errorf("difficulty out of range: %.2f", game.Difficulty)
	}

	if game.Rules == "" {
		return fmt.Errorf("game rules are empty")
	}

	if game.State == nil {
		return fmt.Errorf("game state is nil")
	}

	return nil
}

func (g *Generator) selectGameType(rng *rand.Rand, genreID string) GameType {
	// Genre-specific game type preferences
	switch genreID {
	case "scifi":
		// Prefer hacking in sci-fi
		roll := rng.Float64()
		if roll < 0.3 {
			return GameTypeHacking
		}
	case "fantasy", "horror":
		// Prefer rituals in fantasy/horror
		roll := rng.Float64()
		if roll < 0.25 {
			return GameTypeRitual
		}
	}

	// Default: random selection
	return GameType(rng.Intn(7))
}

func (g *Generator) generateCardGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	deckSize := 20 + rng.Intn(33) // 20-52 cards
	handSize := 5 + rng.Intn(3)   // 5-7 cards

	state := &CardGameState{
		DeckSize:   deckSize,
		HandSize:   handSize,
		TargetWins: 2 + int(params.Difficulty*3), // 2-5 wins needed
	}

	name := g.generateCardGameName(rng, params.GenreID)
	timeLimit := 300.0 + rng.Float64()*300.0 // 5-10 minutes

	return &MiniGame{
		Type:       GameTypeCard,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("First to %d wins. Draw cards to beat opponent.", state.TargetWins),
		State:      state,
		GenreID:    params.GenreID,
	}
}

func (g *Generator) generateDiceGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	numDice := 2 + rng.Intn(4)   // 2-5 dice
	diceSides := 6 + rng.Intn(7) // 6-12 sides

	state := &DiceGameState{
		NumDice:    numDice,
		DiceSides:  diceSides,
		BetAmount:  10 + rng.Intn(91), // 10-100 gold
		TargetRoll: int(float64(numDice*diceSides) * (0.5 + params.Difficulty*0.3)),
	}

	name := g.generateDiceGameName(rng, params.GenreID)
	timeLimit := 120.0 + rng.Float64()*180.0 // 2-5 minutes

	return &MiniGame{
		Type:       GameTypeDice,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("Roll %dd%d. Beat %d to win %d gold.", numDice, diceSides, state.TargetRoll, state.BetAmount),
		State:      state,
		GenreID:    params.GenreID,
	}
}

func (g *Generator) generatePuzzleGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	gridSize := 3 + int(params.Difficulty*2) // 3x3 to 5x5

	state := &PuzzleGameState{
		GridSize: gridSize,
		Moves:    0,
		MaxMoves: (gridSize*gridSize)*5 + int(params.Difficulty*20),
	}

	name := g.generatePuzzleGameName(rng, params.GenreID)
	timeLimit := 180.0 + rng.Float64()*240.0 // 3-7 minutes

	return &MiniGame{
		Type:       GameTypePuzzle,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("Solve %dx%d sliding puzzle in under %d moves.", gridSize, gridSize, state.MaxMoves),
		State:      state,
		GenreID:    params.GenreID,
	}
}

func (g *Generator) generateMemoryGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	numPairs := 6 + int(params.Difficulty*6) // 6-12 pairs

	state := &MemoryGameState{
		NumPairs:     numPairs,
		MaxAttempts:  numPairs*2 + int(params.Difficulty*10),
		SequenceMode: rng.Float64() < 0.3, // 30% chance of sequence mode
	}

	name := g.generateMemoryGameName(rng, params.GenreID)
	timeLimit := 120.0 + rng.Float64()*120.0 // 2-4 minutes

	gameMode := "pairs"
	if state.SequenceMode {
		gameMode = "sequence"
	}

	return &MiniGame{
		Type:       GameTypeMemory,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("Match %d %s within %d attempts.", numPairs, gameMode, state.MaxAttempts),
		State:      state,
		GenreID:    params.GenreID,
	}
}

func (g *Generator) generateLockPickingGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	numPins := 3 + int(params.Difficulty*4)     // 3-7 pins
	timingWindow := 0.5 - params.Difficulty*0.3 // 0.5s (easy) to 0.2s (hard)

	state := &LockPickingGameState{
		NumPins:       numPins,
		TimingWindow:  timingWindow,
		MaxFailures:   3,
		CurrentPinIdx: 0,
	}

	name := g.generateLockPickingGameName(rng, params.GenreID)
	timeLimit := 30.0 + float64(numPins)*15.0 // 0.5-2 minutes based on pins

	return &MiniGame{
		Type:       GameTypeLockPicking,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("Pick %d pins with %.2fs timing window. Max 3 failures.", numPins, timingWindow),
		State:      state,
		GenreID:    params.GenreID,
	}
}

func (g *Generator) generateHackingGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	codeLength := 4 + int(params.Difficulty*4)  // 4-8 digits
	maxAttempts := 5 + int(params.Difficulty*5) // 5-10 attempts

	state := &HackingGameState{
		CodeLength:  codeLength,
		MaxAttempts: maxAttempts,
		Attempts:    0,
		Hints:       make([]string, 0),
	}

	name := g.generateHackingGameName(rng, params.GenreID)
	timeLimit := 60.0 + rng.Float64()*120.0 // 1-3 minutes

	return &MiniGame{
		Type:       GameTypeHacking,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("Crack %d-digit code in %d attempts. Use feedback hints.", codeLength, maxAttempts),
		State:      state,
		GenreID:    params.GenreID,
	}
}

func (g *Generator) generateRitualGame(rng *rand.Rand, params procgen.GenerationParams) *MiniGame {
	numSymbols := 3 + int(params.Difficulty*4)  // 3-7 symbols
	drawAccuracy := 0.7 - params.Difficulty*0.2 // 70% (easy) to 50% (hard)

	state := &RitualGameState{
		NumSymbols:    numSymbols,
		DrawAccuracy:  drawAccuracy,
		CurrentSymbol: 0,
		Completed:     false,
	}

	name := g.generateRitualGameName(rng, params.GenreID)
	timeLimit := 120.0 + rng.Float64()*180.0 // 2-5 minutes

	return &MiniGame{
		Type:       GameTypeRitual,
		Name:       name,
		Difficulty: params.Difficulty,
		TimeLimit:  timeLimit,
		Rules:      fmt.Sprintf("Draw %d symbols with %.0f%% accuracy.", numSymbols, drawAccuracy*100),
		State:      state,
		GenreID:    params.GenreID,
	}
}

// generateGenreName selects a random name from nameMap[genreID], falling back
// to nameMap["default"] if the genre is not found or has no entries.
func generateGenreName(rng *rand.Rand, genreID string, nameMap map[string][]string) string {
	if names, ok := nameMap[genreID]; ok && len(names) > 0 {
		return names[rng.Intn(len(names))]
	}
	if names, ok := nameMap["default"]; ok && len(names) > 0 {
		return names[rng.Intn(len(names))]
	}
	return "Unknown Game"
}

// Name generation functions
func (g *Generator) generateCardGameName(rng *rand.Rand, genreID string) string {
	prefixes := []string{"Dragon", "Shadow", "Ancient", "Royal", "Mystic"}
	suffixes := []string{"Hold'em", "Draw", "War", "Conquest", "Duel"}

	if genreID == "scifi" {
		prefixes = []string{"Cyber", "Quantum", "Neural", "Data", "System"}
		suffixes = []string{"Protocol", "Sequence", "Matrix", "Grid", "Network"}
	}

	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (g *Generator) generateDiceGameName(rng *rand.Rand, genreID string) string {
	return generateGenreName(rng, genreID, map[string][]string{
		"scifi":   {"Probability Core", "Random Gen", "Chaos Dice", "Entropy Roll", "Quantum Luck"},
		"default": {"Bones", "Fortune's Favor", "Luck's Roll", "Fate's Dice", "Gambler's Choice"},
	})
}

func (g *Generator) generatePuzzleGameName(rng *rand.Rand, genreID string) string {
	return generateGenreName(rng, genreID, map[string][]string{
		"scifi":   {"Circuit Solver", "Data Grid", "Matrix Puzzle", "Logic Gate", "Neural Net"},
		"default": {"Sliding Stones", "Ancient Tiles", "Mystic Grid", "Rune Puzzle", "Maze of Mind"},
	})
}

func (g *Generator) generateMemoryGameName(rng *rand.Rand, genreID string) string {
	return generateGenreName(rng, genreID, map[string][]string{
		"scifi":   {"Memory Buffer", "Data Recall", "Cache Match", "Pattern Scan", "Neural Link"},
		"default": {"Memory Match", "Recall Challenge", "Mind Game", "Pattern Memory", "Symbol Sequence"},
	})
}

func (g *Generator) generateLockPickingGameName(rng *rand.Rand, genreID string) string {
	return generateGenreName(rng, genreID, map[string][]string{
		"scifi":   {"Firewall Breach", "Security Override", "Encryption Crack", "Access Hack", "Code Break"},
		"default": {"Lock Pick", "Tumbler Trial", "Key Master", "Pin Puzzle", "Lock Breaker"},
	})
}

func (g *Generator) generateHackingGameName(rng *rand.Rand, genreID string) string {
	return generateGenreName(rng, genreID, map[string][]string{
		"cyberpunk": {"ICE Breaker", "Netrun Protocol", "Black ICE Bypass", "Cortex Hack", "Data Heist"},
		"default":   {"Code Breaker", "System Hack", "Terminal Access", "Network Infiltration", "Data Breach"},
	})
}

func (g *Generator) generateRitualGameName(rng *rand.Rand, genreID string) string {
	return generateGenreName(rng, genreID, map[string][]string{
		"horror":  {"Dark Ritual", "Blood Rite", "Cursed Ceremony", "Unholy Pact", "Shadow Binding"},
		"default": {"Mystic Ritual", "Spell Weaving", "Rune Casting", "Arcane Ceremony", "Magic Circle"},
	})
}
