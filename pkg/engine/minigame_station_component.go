// Package engine provides the core ECS (Entity-Component-System) framework.
// This file defines mini-game station components for Phase 27.3: Mini-Game Integration.
package engine

// MiniGameStationComponent marks an entity as a mini-game station (tavern game table, etc.).
// When player interacts with F key, starts the specified mini-game.
//
// Phase 27.3: Mini-Game Integration - Tavern Stations
type MiniGameStationComponent struct {
	// GameType specifies which mini-game this station offers
	GameType MiniGameType

	// Difficulty is the base difficulty for games at this station (0.0-1.0)
	Difficulty float64

	// StationName is the display name (e.g., "Card Table", "Dice Den")
	StationName string

	// IsOccupied indicates if someone is currently playing
	IsOccupied bool

	// OccupiedByPlayerID is the entity ID of the player using this station
	OccupiedByPlayerID uint64

	// EntryCost is gold required to play (-1 for free)
	EntryCost int

	// RequiresLevel is minimum player level (0 for no requirement)
	RequiresLevel int
}

// Type returns the component type identifier "minigameStation".
func (m MiniGameStationComponent) Type() string {
	return "minigameStation"
}

// NewMiniGameStationComponent creates a new mini-game station component.
// Provides default values for common tavern game stations.
func NewMiniGameStationComponent(gameType MiniGameType, difficulty float64) *MiniGameStationComponent {
	return &MiniGameStationComponent{
		GameType:           gameType,
		Difficulty:         clampDifficulty(difficulty),
		StationName:        getDefaultStationName(gameType),
		IsOccupied:         false,
		OccupiedByPlayerID: 0,
		EntryCost:          getDefaultEntryCost(gameType),
		RequiresLevel:      0,
	}
}

// CanPlay returns true if the station is available and player meets requirements.
func (m *MiniGameStationComponent) CanPlay(playerLevel, playerGold int) bool {
	if m.IsOccupied {
		return false
	}
	if playerLevel < m.RequiresLevel {
		return false
	}
	if m.EntryCost > 0 && playerGold < m.EntryCost {
		return false
	}
	return true
}

// StartGame marks the station as occupied by a player.
func (m *MiniGameStationComponent) StartGame(playerID uint64) {
	m.IsOccupied = true
	m.OccupiedByPlayerID = playerID
}

// EndGame marks the station as available again.
func (m *MiniGameStationComponent) EndGame() {
	m.IsOccupied = false
	m.OccupiedByPlayerID = 0
}

// clampDifficulty ensures difficulty is in valid range [0.0, 1.0].
func clampDifficulty(difficulty float64) float64 {
	if difficulty < 0.0 {
		return 0.0
	}
	if difficulty > 1.0 {
		return 1.0
	}
	return difficulty
}

// getDefaultStationName returns genre-appropriate station names for game types.
func getDefaultStationName(gameType MiniGameType) string {
	switch gameType {
	case MiniGameCard:
		return "Card Table"
	case MiniGameDice:
		return "Dice Den"
	case MiniGamePuzzle:
		return "Puzzle Board"
	case MiniGameMemory:
		return "Memory Game"
	case MiniGameLockPicking:
		return "Practice Lock"
	case MiniGameHacking:
		return "Terminal"
	case MiniGameRitual:
		return "Ritual Circle"
	default:
		return "Game Station"
	}
}

// getDefaultEntryCost returns typical entry cost for different game types.
// Returns gold amount, or -1 for free games.
func getDefaultEntryCost(gameType MiniGameType) int {
	switch gameType {
	case MiniGameCard:
		return 20 // Gambling games cost money
	case MiniGameDice:
		return 15
	case MiniGamePuzzle:
		return -1 // Practice puzzles are free
	case MiniGameMemory:
		return -1
	case MiniGameLockPicking:
		return -1
	case MiniGameHacking:
		return 10 // Sci-fi hacking stations
	case MiniGameRitual:
		return 5 // Small cost for ritual materials
	default:
		return -1
	}
}
