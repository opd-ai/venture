// Package engine provides the matchmaking component for PvP queue management.
// MatchmakingComponent tracks player queue status, preferences, and match history.
package engine

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// MatchmakingMode defines the type of PvP match.
type MatchmakingMode string

const (
	// MatchmakingMode1v1 is a duel between two players.
	MatchmakingMode1v1 MatchmakingMode = "1v1"
	// MatchmakingMode2v2 is a team battle with 2 players per team.
	MatchmakingMode2v2 MatchmakingMode = "2v2"
	// MatchmakingModeFFA is a free-for-all with multiple players.
	MatchmakingModeFFA MatchmakingMode = "ffa"
)

// MatchmakingState tracks the current queue state.
type MatchmakingState string

const (
	// MatchmakingStateIdle means the player is not in queue.
	MatchmakingStateIdle MatchmakingState = "idle"
	// MatchmakingStateQueued means the player is waiting for a match.
	MatchmakingStateQueued MatchmakingState = "queued"
	// MatchmakingStateMatched means a match has been found.
	MatchmakingStateMatched MatchmakingState = "matched"
	// MatchmakingStateInMatch means the player is currently in a match.
	MatchmakingStateInMatch MatchmakingState = "in_match"
)

// MatchmakingPreferences holds player queue preferences.
type MatchmakingPreferences struct {
	// PreferredModes lists modes the player wants to queue for
	PreferredModes []MatchmakingMode `json:"preferred_modes"`
	// AcceptCrossServer allows matching with players from other servers
	AcceptCrossServer bool `json:"accept_cross_server"`
	// MaxPingMs is the maximum acceptable ping for a match
	MaxPingMs int `json:"max_ping_ms"`
}

// QueueEntry represents a player's position in the matchmaking queue.
type QueueEntry struct {
	// PlayerID is the entity ID of the queued player
	PlayerID uint64 `json:"player_id"`
	// Rating is the player's current ELO rating
	Rating int `json:"rating"`
	// Mode is the matchmaking mode for this queue entry
	Mode MatchmakingMode `json:"mode"`
	// QueuedAt is when the player joined the queue
	QueuedAt time.Time `json:"queued_at"`
	// RatingRange is the current acceptable rating difference
	RatingRange int `json:"rating_range"`
	// ServerID is the home server of this player
	ServerID string `json:"server_id"`
	// TeamID is the team identifier for team modes (empty for solo)
	TeamID string `json:"team_id"`
}

// MatchResult records the outcome of a match.
type MatchResult struct {
	// MatchID is the unique match identifier
	MatchID string `json:"match_id"`
	// Mode is the matchmaking mode
	Mode MatchmakingMode `json:"mode"`
	// Participants lists all player entity IDs in the match
	Participants []uint64 `json:"participants"`
	// WinnerIDs lists the winning player entity IDs
	WinnerIDs []uint64 `json:"winner_ids"`
	// LoserIDs lists the losing player entity IDs
	LoserIDs []uint64 `json:"loser_ids"`
	// Duration is how long the match lasted
	Duration time.Duration `json:"duration"`
	// CompletedAt is when the match ended
	CompletedAt time.Time `json:"completed_at"`
	// RatingChanges maps player ID to rating change (+/-)
	RatingChanges map[uint64]int `json:"rating_changes"`
}

// MatchmakingComponent tracks a player's matchmaking state and history.
type MatchmakingComponent struct {
	// State is the current matchmaking state
	State MatchmakingState `json:"state"`
	// CurrentMode is the mode the player is queued for
	CurrentMode MatchmakingMode `json:"current_mode"`
	// QueuedAt is when the player joined the current queue
	QueuedAt time.Time `json:"queued_at"`
	// Preferences holds the player's queue preferences
	Preferences MatchmakingPreferences `json:"preferences"`
	// CurrentMatchID is the ID of the current/pending match
	CurrentMatchID string `json:"current_match_id"`
	// MatchHistory stores recent match results
	MatchHistory []MatchResult `json:"match_history"`
	// MaxHistorySize limits stored match history
	MaxHistorySize int `json:"max_history_size"`
	// TotalQueueTime tracks cumulative queue time for stats
	TotalQueueTime time.Duration `json:"total_queue_time"`
	// TotalMatches tracks total matches played
	TotalMatches int `json:"total_matches"`
	// TeamID is the player's team for team modes
	TeamID string `json:"team_id"`
	// ServerID is the player's home server
	ServerID string `json:"server_id"`
}

// NewMatchmakingComponent creates a new matchmaking component.
func NewMatchmakingComponent(serverID string) *MatchmakingComponent {
	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"server_id":      serverID,
	}).Debug("Creating matchmaking component")

	return &MatchmakingComponent{
		State:       MatchmakingStateIdle,
		CurrentMode: MatchmakingMode1v1,
		Preferences: MatchmakingPreferences{
			PreferredModes:    []MatchmakingMode{MatchmakingMode1v1},
			AcceptCrossServer: true,
			MaxPingMs:         200,
		},
		MatchHistory:   make([]MatchResult, 0),
		MaxHistorySize: 50,
		ServerID:       serverID,
	}
}

// Type returns the component type identifier.
func (c *MatchmakingComponent) Type() string {
	return "matchmaking"
}

// EnterQueue puts the player in the matchmaking queue.
func (c *MatchmakingComponent) EnterQueue(mode MatchmakingMode) bool {
	if c.State != MatchmakingStateIdle {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"current_state":  c.State,
			"reason":         "not_idle",
		}).Debug("Cannot enter queue")
		return false
	}

	c.State = MatchmakingStateQueued
	c.CurrentMode = mode
	c.QueuedAt = time.Now()

	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"mode":           mode,
	}).Info("Entered matchmaking queue")

	return true
}

// LeaveQueue removes the player from the matchmaking queue.
func (c *MatchmakingComponent) LeaveQueue() bool {
	if c.State != MatchmakingStateQueued {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"current_state":  c.State,
			"reason":         "not_queued",
		}).Debug("Cannot leave queue")
		return false
	}

	// Track queue time
	c.TotalQueueTime += time.Since(c.QueuedAt)

	c.State = MatchmakingStateIdle
	c.QueuedAt = time.Time{}

	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
	}).Info("Left matchmaking queue")

	return true
}

// AcceptMatch accepts a found match.
func (c *MatchmakingComponent) AcceptMatch(matchID string) bool {
	if c.State != MatchmakingStateMatched {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"current_state":  c.State,
			"reason":         "not_matched",
		}).Debug("Cannot accept match")
		return false
	}

	c.State = MatchmakingStateInMatch
	c.CurrentMatchID = matchID

	// Track queue time
	c.TotalQueueTime += time.Since(c.QueuedAt)

	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"match_id":       matchID,
	}).Info("Accepted match")

	return true
}

// DeclineMatch declines a found match and returns to idle.
func (c *MatchmakingComponent) DeclineMatch() bool {
	if c.State != MatchmakingStateMatched {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"current_state":  c.State,
			"reason":         "not_matched",
		}).Debug("Cannot decline match")
		return false
	}

	c.State = MatchmakingStateIdle
	c.CurrentMatchID = ""

	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
	}).Info("Declined match")

	return true
}

// CompleteMatch records the completion of a match.
func (c *MatchmakingComponent) CompleteMatch(result MatchResult) {
	if c.State != MatchmakingStateInMatch {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"current_state":  c.State,
		}).Warn("CompleteMatch called when not in match")
	}

	c.State = MatchmakingStateIdle
	c.CurrentMatchID = ""
	c.TotalMatches++

	// Add to history
	c.MatchHistory = append(c.MatchHistory, result)
	if len(c.MatchHistory) > c.MaxHistorySize {
		c.MatchHistory = c.MatchHistory[len(c.MatchHistory)-c.MaxHistorySize:]
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"match_id":       result.MatchID,
		"duration":       result.Duration.String(),
	}).Info("Match completed")
}

// MarkMatched marks the player as matched (waiting for acceptance).
func (c *MatchmakingComponent) MarkMatched(matchID string) bool {
	if c.State != MatchmakingStateQueued {
		return false
	}

	c.State = MatchmakingStateMatched
	c.CurrentMatchID = matchID
	return true
}

// IsQueued returns true if the player is currently in queue.
func (c *MatchmakingComponent) IsQueued() bool {
	return c.State == MatchmakingStateQueued
}

// IsInMatch returns true if the player is currently in a match.
func (c *MatchmakingComponent) IsInMatch() bool {
	return c.State == MatchmakingStateInMatch
}

// GetQueueDuration returns how long the player has been in queue.
func (c *MatchmakingComponent) GetQueueDuration() time.Duration {
	if c.State != MatchmakingStateQueued && c.State != MatchmakingStateMatched {
		return 0
	}
	return time.Since(c.QueuedAt)
}

// GetAverageQueueTime returns the average queue time across all matches.
func (c *MatchmakingComponent) GetAverageQueueTime() time.Duration {
	if c.TotalMatches == 0 {
		return 0
	}
	return c.TotalQueueTime / time.Duration(c.TotalMatches)
}

// GetRecentMatches returns the last N matches from history.
func (c *MatchmakingComponent) GetRecentMatches(count int) []MatchResult {
	if count <= 0 || len(c.MatchHistory) == 0 {
		return nil
	}
	if count > len(c.MatchHistory) {
		count = len(c.MatchHistory)
	}
	start := len(c.MatchHistory) - count
	return c.MatchHistory[start:]
}

// SetPreferences updates the player's matchmaking preferences.
func (c *MatchmakingComponent) SetPreferences(prefs MatchmakingPreferences) {
	c.Preferences = prefs
	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"modes":          prefs.PreferredModes,
		"cross_server":   prefs.AcceptCrossServer,
	}).Debug("Updated matchmaking preferences")
}

// SetTeam sets the player's team for team modes.
func (c *MatchmakingComponent) SetTeam(teamID string) {
	c.TeamID = teamID
}

// ClearTeam removes the player from their team.
func (c *MatchmakingComponent) ClearTeam() {
	c.TeamID = ""
}

// Serialize converts the component to JSON bytes.
func (c *MatchmakingComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"state":          c.State,
		"total_matches":  c.TotalMatches,
	}).Debug("Serializing matchmaking component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"error":          err.Error(),
		}).Error("Failed to serialize matchmaking component")
		return nil, err
	}
	return data, nil
}

// Deserialize loads the component from JSON bytes.
func (c *MatchmakingComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "matchmaking",
		"bytes":          len(data),
	}).Debug("Deserializing matchmaking component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "matchmaking",
			"error":          err.Error(),
		}).Error("Failed to deserialize matchmaking component")
		return err
	}
	return nil
}

// GetWinCount returns the number of wins in match history.
func (c *MatchmakingComponent) GetWinCount(playerID uint64) int {
	wins := 0
	for _, match := range c.MatchHistory {
		for _, winnerID := range match.WinnerIDs {
			if winnerID == playerID {
				wins++
				break
			}
		}
	}
	return wins
}

// GetLossCount returns the number of losses in match history.
func (c *MatchmakingComponent) GetLossCount(playerID uint64) int {
	losses := 0
	for _, match := range c.MatchHistory {
		for _, loserID := range match.LoserIDs {
			if loserID == playerID {
				losses++
				break
			}
		}
	}
	return losses
}
