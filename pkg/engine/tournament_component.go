// Package engine provides the tournament component for competitive PvP events.
// TournamentComponent tracks player participation in scheduled tournaments.
package engine

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// TournamentFormat defines the bracket type.
type TournamentFormat string

const (
	// TournamentFormatSingleElim is single elimination (lose once = out).
	TournamentFormatSingleElim TournamentFormat = "single_elimination"
	// TournamentFormatDoubleElim is double elimination (lose twice = out).
	TournamentFormatDoubleElim TournamentFormat = "double_elimination"
	// TournamentFormatRoundRobin has all players face each other.
	TournamentFormatRoundRobin TournamentFormat = "round_robin"
)

// TournamentPhase tracks the tournament lifecycle.
type TournamentPhase string

const (
	// TournamentPhaseRegistration is when players can sign up.
	TournamentPhaseRegistration TournamentPhase = "registration"
	// TournamentPhaseSeeding is when brackets are generated.
	TournamentPhaseSeeding TournamentPhase = "seeding"
	// TournamentPhaseInProgress is during active competition.
	TournamentPhaseInProgress TournamentPhase = "in_progress"
	// TournamentPhaseComplete is when the tournament has ended.
	TournamentPhaseComplete TournamentPhase = "complete"
	// TournamentPhaseCancelled is when the tournament was cancelled.
	TournamentPhaseCancelled TournamentPhase = "cancelled"
)

// TournamentFrequency defines how often tournaments run.
type TournamentFrequency string

const (
	TournamentFrequencyDaily   TournamentFrequency = "daily"
	TournamentFrequencyWeekly  TournamentFrequency = "weekly"
	TournamentFrequencyMonthly TournamentFrequency = "monthly"
	TournamentFrequencySpecial TournamentFrequency = "special"
)

// TournamentDefinition defines a tournament template.
type TournamentDefinition struct {
	// ID is the unique tournament type identifier
	ID string `json:"id"`
	// Name is the display name
	Name string `json:"name"`
	// Format is the bracket type
	Format TournamentFormat `json:"format"`
	// Frequency is how often this tournament runs
	Frequency TournamentFrequency `json:"frequency"`
	// MinParticipants is the minimum required to start
	MinParticipants int `json:"min_participants"`
	// MaxParticipants is the maximum allowed
	MaxParticipants int `json:"max_participants"`
	// RegistrationDuration is how long registration is open
	RegistrationDuration time.Duration `json:"registration_duration"`
	// MatchTimeLimit is the max duration per match
	MatchTimeLimit time.Duration `json:"match_time_limit"`
	// RatingRequirement is the minimum rating to enter (0 = none)
	RatingRequirement int `json:"rating_requirement"`
	// EntryFee is any currency cost to enter (0 = free)
	EntryFee int `json:"entry_fee"`
	// Seed for deterministic bracket generation
	Seed int64 `json:"seed"`
	// EventID links to seasonal event (empty = standalone)
	EventID string `json:"event_id"`
}

// BracketMatch represents a single match in the bracket.
type BracketMatch struct {
	// MatchID is the unique match identifier
	MatchID string `json:"match_id"`
	// Round is the tournament round (1 = first round)
	Round int `json:"round"`
	// Position is the match position within the round
	Position int `json:"position"`
	// Player1ID is the first player entity ID (0 = bye/TBD)
	Player1ID uint64 `json:"player1_id"`
	// Player2ID is the second player entity ID (0 = bye/TBD)
	Player2ID uint64 `json:"player2_id"`
	// WinnerID is the winning player (0 = not yet played)
	WinnerID uint64 `json:"winner_id"`
	// LoserID is the losing player (0 = not yet played)
	LoserID uint64 `json:"loser_id"`
	// Completed indicates if the match has finished
	Completed bool `json:"completed"`
	// ScheduledAt is when the match is scheduled
	ScheduledAt time.Time `json:"scheduled_at"`
	// CompletedAt is when the match ended
	CompletedAt time.Time `json:"completed_at"`
}

// TournamentInstance represents an active tournament.
type TournamentInstance struct {
	// InstanceID is the unique instance identifier
	InstanceID string `json:"instance_id"`
	// Definition is the tournament template
	Definition TournamentDefinition `json:"definition"`
	// Phase is the current tournament phase
	Phase TournamentPhase `json:"phase"`
	// Participants lists registered player entity IDs
	Participants []uint64 `json:"participants"`
	// Seeds maps player ID to seed position (1 = top seed)
	Seeds map[uint64]int `json:"seeds"`
	// Bracket is the match schedule
	Bracket []BracketMatch `json:"bracket"`
	// CurrentRound is the active round
	CurrentRound int `json:"current_round"`
	// TotalRounds is the number of rounds needed
	TotalRounds int `json:"total_rounds"`
	// WinnerID is the tournament champion (0 = not yet determined)
	WinnerID uint64 `json:"winner_id"`
	// RunnerUpID is second place (0 = not yet determined)
	RunnerUpID uint64 `json:"runner_up_id"`
	// CreatedAt is when the tournament was created
	CreatedAt time.Time `json:"created_at"`
	// StartsAt is when competition begins
	StartsAt time.Time `json:"starts_at"`
	// EndsAt is when the tournament concluded
	EndsAt time.Time `json:"ends_at"`
	// LosersBracket is for double elimination (optional)
	LosersBracket []BracketMatch `json:"losers_bracket,omitempty"`
}

// TournamentComponent tracks a player's tournament participation.
type TournamentComponent struct {
	// CurrentTournamentID is the tournament the player is in (empty = none)
	CurrentTournamentID string `json:"current_tournament_id"`
	// BracketPosition is the player's position in the bracket
	BracketPosition int `json:"bracket_position"`
	// Seed is the player's seeding (1 = top seed)
	Seed int `json:"seed"`
	// Eliminated indicates if the player is out
	Eliminated bool `json:"eliminated"`
	// LossesInTournament tracks losses (for double elim)
	LossesInTournament int `json:"losses_in_tournament"`
	// MatchesWon tracks wins in current tournament
	MatchesWon int `json:"matches_won"`
	// TotalTournamentsEntered is lifetime stat
	TotalTournamentsEntered int `json:"total_tournaments_entered"`
	// TotalTournamentWins is lifetime wins
	TotalTournamentWins int `json:"total_tournament_wins"`
	// TotalTournamentMatches is lifetime matches in tournaments
	TotalTournamentMatches int `json:"total_tournament_matches"`
	// TournamentHistory stores recent tournament results
	TournamentHistory []TournamentResult `json:"tournament_history"`
	// MaxHistorySize limits stored history
	MaxHistorySize int `json:"max_history_size"`
	// IsSpectating indicates if watching a tournament
	IsSpectating bool `json:"is_spectating"`
	// SpectatingTournamentID is the tournament being watched
	SpectatingTournamentID string `json:"spectating_tournament_id"`
}

// TournamentResult records a player's tournament performance.
type TournamentResult struct {
	// TournamentID is the tournament instance ID
	TournamentID string `json:"tournament_id"`
	// TournamentName is the display name
	TournamentName string `json:"tournament_name"`
	// Placement is the final position (1 = winner)
	Placement int `json:"placement"`
	// TotalParticipants is how many entered
	TotalParticipants int `json:"total_participants"`
	// MatchesWon is wins in this tournament
	MatchesWon int `json:"matches_won"`
	// MatchesLost is losses in this tournament
	MatchesLost int `json:"matches_lost"`
	// CompletedAt is when the tournament ended
	CompletedAt time.Time `json:"completed_at"`
}

// NewTournamentComponent creates a new tournament component.
func NewTournamentComponent() *TournamentComponent {
	logrus.WithFields(logrus.Fields{
		"component_type": "tournament",
	}).Debug("Creating tournament component")

	return &TournamentComponent{
		TournamentHistory: make([]TournamentResult, 0),
		MaxHistorySize:    50,
	}
}

// Type returns the component type identifier.
func (c *TournamentComponent) Type() string {
	return "tournament"
}

// EnterTournament registers the player for a tournament.
func (c *TournamentComponent) EnterTournament(tournamentID string, seed int) bool {
	if c.CurrentTournamentID != "" {
		logrus.WithFields(logrus.Fields{
			"component_type":     "tournament",
			"current_tournament": c.CurrentTournamentID,
			"reason":             "already_in_tournament",
		}).Debug("Cannot enter tournament")
		return false
	}

	c.CurrentTournamentID = tournamentID
	c.Seed = seed
	c.BracketPosition = 0
	c.Eliminated = false
	c.LossesInTournament = 0
	c.MatchesWon = 0
	c.TotalTournamentsEntered++

	logrus.WithFields(logrus.Fields{
		"component_type": "tournament",
		"tournament_id":  tournamentID,
		"seed":           seed,
	}).Info("Entered tournament")

	return true
}

// LeaveTournament withdraws the player from a tournament.
func (c *TournamentComponent) LeaveTournament() bool {
	if c.CurrentTournamentID == "" {
		return false
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "tournament",
		"tournament_id":  c.CurrentTournamentID,
	}).Info("Left tournament")

	c.CurrentTournamentID = ""
	c.Seed = 0
	c.BracketPosition = 0
	c.Eliminated = false
	c.LossesInTournament = 0
	c.MatchesWon = 0

	return true
}

// RecordWin records a tournament match win.
func (c *TournamentComponent) RecordWin() {
	c.MatchesWon++
	c.TotalTournamentMatches++

	logrus.WithFields(logrus.Fields{
		"component_type": "tournament",
		"matches_won":    c.MatchesWon,
	}).Debug("Recorded tournament win")
}

// RecordLoss records a tournament match loss.
func (c *TournamentComponent) RecordLoss(isDoubleElim bool) {
	c.LossesInTournament++
	c.TotalTournamentMatches++

	// In single elimination, one loss eliminates
	// In double elimination, two losses eliminates
	if isDoubleElim {
		if c.LossesInTournament >= 2 {
			c.Eliminated = true
		}
	} else {
		c.Eliminated = true
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "tournament",
		"losses":         c.LossesInTournament,
		"eliminated":     c.Eliminated,
	}).Debug("Recorded tournament loss")
}

// CompleteTournament finalizes tournament participation.
func (c *TournamentComponent) CompleteTournament(result TournamentResult) {
	if result.Placement == 1 {
		c.TotalTournamentWins++
	}

	// Add to history
	c.TournamentHistory = append(c.TournamentHistory, result)
	if len(c.TournamentHistory) > c.MaxHistorySize {
		c.TournamentHistory = c.TournamentHistory[len(c.TournamentHistory)-c.MaxHistorySize:]
	}

	// Reset current tournament state
	c.CurrentTournamentID = ""
	c.Seed = 0
	c.BracketPosition = 0
	c.Eliminated = false
	c.LossesInTournament = 0
	c.MatchesWon = 0

	logrus.WithFields(logrus.Fields{
		"component_type": "tournament",
		"tournament_id":  result.TournamentID,
		"placement":      result.Placement,
	}).Info("Tournament completed")
}

// IsInTournament returns true if the player is in a tournament.
func (c *TournamentComponent) IsInTournament() bool {
	return c.CurrentTournamentID != ""
}

// IsEliminated returns true if the player has been eliminated.
func (c *TournamentComponent) IsEliminated() bool {
	return c.Eliminated
}

// StartSpectating begins watching a tournament.
func (c *TournamentComponent) StartSpectating(tournamentID string) bool {
	if c.IsInTournament() {
		return false
	}

	c.IsSpectating = true
	c.SpectatingTournamentID = tournamentID
	return true
}

// StopSpectating stops watching a tournament.
func (c *TournamentComponent) StopSpectating() {
	c.IsSpectating = false
	c.SpectatingTournamentID = ""
}

// GetRecentTournaments returns the last N tournament results.
func (c *TournamentComponent) GetRecentTournaments(count int) []TournamentResult {
	if count <= 0 || len(c.TournamentHistory) == 0 {
		return nil
	}
	if count > len(c.TournamentHistory) {
		count = len(c.TournamentHistory)
	}
	start := len(c.TournamentHistory) - count
	return c.TournamentHistory[start:]
}

// GetWinRate returns the tournament win rate as a percentage.
func (c *TournamentComponent) GetWinRate() float64 {
	if c.TotalTournamentsEntered == 0 {
		return 0.0
	}
	return float64(c.TotalTournamentWins) / float64(c.TotalTournamentsEntered) * 100.0
}

// GetAveragePlacement returns the average tournament placement.
func (c *TournamentComponent) GetAveragePlacement() float64 {
	if len(c.TournamentHistory) == 0 {
		return 0.0
	}
	total := 0
	for _, result := range c.TournamentHistory {
		total += result.Placement
	}
	return float64(total) / float64(len(c.TournamentHistory))
}

// Serialize converts the component to JSON bytes.
func (c *TournamentComponent) Serialize() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "tournament",
			"error":          err.Error(),
		}).Error("Failed to serialize tournament component")
		return nil, err
	}
	return data, nil
}

// Deserialize loads the component from JSON bytes.
func (c *TournamentComponent) Deserialize(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "tournament",
			"error":          err.Error(),
		}).Error("Failed to deserialize tournament component")
		return err
	}
	return nil
}

// GenerateSingleElimBracket generates a single elimination bracket.
// Deterministic given the same seed and participants.
func GenerateSingleElimBracket(seed int64, participants []uint64) []BracketMatch {
	rng := rand.New(rand.NewSource(seed))
	n := len(participants)
	if n < 2 {
		return nil
	}

	// Shuffle participants for seeding variety
	shuffled := make([]uint64, n)
	copy(shuffled, participants)
	rng.Shuffle(n, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Calculate number of rounds needed
	rounds := 0
	for pow := 1; pow < n; pow *= 2 {
		rounds++
	}

	// Pad to nearest power of 2 with byes
	bracketSize := 1
	for bracketSize < n {
		bracketSize *= 2
	}

	// Place participants (byes are 0)
	slots := make([]uint64, bracketSize)
	for i := 0; i < n; i++ {
		slots[i] = shuffled[i]
	}

	// Generate first round matches
	var bracket []BracketMatch
	matchID := 0
	for round := 1; round <= rounds; round++ {
		matchesInRound := bracketSize / (1 << round)
		for pos := 0; pos < matchesInRound; pos++ {
			match := BracketMatch{
				MatchID:  generateMatchID(seed, round, pos),
				Round:    round,
				Position: pos,
			}

			if round == 1 {
				// First round: assign from slots
				match.Player1ID = slots[pos*2]
				match.Player2ID = slots[pos*2+1]

				// Handle byes
				if match.Player1ID == 0 && match.Player2ID != 0 {
					match.WinnerID = match.Player2ID
					match.Completed = true
				} else if match.Player2ID == 0 && match.Player1ID != 0 {
					match.WinnerID = match.Player1ID
					match.Completed = true
				}
			}

			bracket = append(bracket, match)
			matchID++
		}
	}

	return bracket
}

// GenerateDoubleElimBracket generates a double elimination bracket.
func GenerateDoubleElimBracket(seed int64, participants []uint64) (winners, losers []BracketMatch) {
	// Winners bracket is same as single elim
	winners = GenerateSingleElimBracket(seed, participants)
	if winners == nil {
		return nil, nil
	}

	n := len(participants)
	// Losers bracket has more rounds (approximately 2x winners - 1)
	rounds := 0
	for pow := 1; pow < n; pow *= 2 {
		rounds++
	}
	losersRounds := rounds * 2

	// Generate empty losers bracket structure
	losers = make([]BracketMatch, 0)
	for round := 1; round <= losersRounds; round++ {
		// Losers bracket has complex match counts
		matchesInRound := 1 << ((losersRounds - round) / 2)
		if matchesInRound < 1 {
			matchesInRound = 1
		}
		for pos := 0; pos < matchesInRound; pos++ {
			match := BracketMatch{
				MatchID:  generateMatchID(seed+1000, round, pos),
				Round:    round,
				Position: pos,
			}
			losers = append(losers, match)
		}
	}

	return winners, losers
}

// CalculateTotalRounds returns rounds needed for a bracket size.
func CalculateTotalRounds(participantCount int) int {
	if participantCount < 2 {
		return 0
	}
	rounds := 0
	for pow := 1; pow < participantCount; pow *= 2 {
		rounds++
	}
	return rounds
}

// generateMatchID creates a deterministic match ID.
func generateMatchID(seed int64, round, position int) string {
	rng := rand.New(rand.NewSource(seed + int64(round*1000) + int64(position)))
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

// GetDefaultTournamentDefinitions returns standard tournament types.
func GetDefaultTournamentDefinitions(baseSeed int64) []TournamentDefinition {
	return []TournamentDefinition{
		{
			ID:                   "daily_1v1",
			Name:                 "Daily 1v1 Championship",
			Format:               TournamentFormatSingleElim,
			Frequency:            TournamentFrequencyDaily,
			MinParticipants:      4,
			MaxParticipants:      32,
			RegistrationDuration: 2 * time.Hour,
			MatchTimeLimit:       15 * time.Minute,
			RatingRequirement:    0,
			EntryFee:             0,
			Seed:                 baseSeed,
		},
		{
			ID:                   "weekly_ranked",
			Name:                 "Weekly Ranked Tournament",
			Format:               TournamentFormatDoubleElim,
			Frequency:            TournamentFrequencyWeekly,
			MinParticipants:      8,
			MaxParticipants:      64,
			RegistrationDuration: 24 * time.Hour,
			MatchTimeLimit:       20 * time.Minute,
			RatingRequirement:    1000,
			EntryFee:             100,
			Seed:                 baseSeed + 1,
		},
		{
			ID:                   "monthly_grand",
			Name:                 "Monthly Grand Championship",
			Format:               TournamentFormatDoubleElim,
			Frequency:            TournamentFrequencyMonthly,
			MinParticipants:      16,
			MaxParticipants:      128,
			RegistrationDuration: 7 * 24 * time.Hour,
			MatchTimeLimit:       30 * time.Minute,
			RatingRequirement:    1200,
			EntryFee:             500,
			Seed:                 baseSeed + 2,
		},
	}
}
