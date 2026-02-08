// Package engine provides the matchmaking system for skill-based PvP matching.
// MatchmakingSystem manages player queues, creates fair matches, and supports
// cross-server matchmaking via federation.
package engine

import (
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// MatchmakingConfig holds configuration for the matchmaking system.
type MatchmakingConfig struct {
	// InitialRatingRange is the starting acceptable rating difference
	InitialRatingRange int
	// RatingRangeExpansionRate is how much the range expands per second
	RatingRangeExpansionRate float64
	// MaxRatingRange is the maximum acceptable rating difference
	MaxRatingRange int
	// MatchAcceptTimeout is how long players have to accept a match
	MatchAcceptTimeout time.Duration
	// MinPlayersPerMode is the minimum players needed per mode
	MinPlayersPerMode map[MatchmakingMode]int
	// MaxPlayersPerMode is the maximum players per mode
	MaxPlayersPerMode map[MatchmakingMode]int
}

// DefaultMatchmakingConfig returns the default matchmaking configuration.
func DefaultMatchmakingConfig() MatchmakingConfig {
	return MatchmakingConfig{
		InitialRatingRange:       200,
		RatingRangeExpansionRate: 10.0, // +10 rating per second
		MaxRatingRange:           500,
		MatchAcceptTimeout:       30 * time.Second,
		MinPlayersPerMode: map[MatchmakingMode]int{
			MatchmakingMode1v1: 2,
			MatchmakingMode2v2: 4,
			MatchmakingModeFFA: 3,
		},
		MaxPlayersPerMode: map[MatchmakingMode]int{
			MatchmakingMode1v1: 2,
			MatchmakingMode2v2: 4,
			MatchmakingModeFFA: 8,
		},
	}
}

// PendingMatch represents a match waiting for player acceptance.
type PendingMatch struct {
	// MatchID is the unique match identifier
	MatchID string
	// Mode is the matchmaking mode
	Mode MatchmakingMode
	// Participants are the entities in this match
	Participants []*Entity
	// AcceptedBy tracks which players have accepted
	AcceptedBy map[uint64]bool
	// CreatedAt is when the match was created
	CreatedAt time.Time
	// AverageRating is the average rating of participants
	AverageRating int
}

// MatchmakingSystem manages player queues and creates fair matches.
type MatchmakingSystem struct {
	world   *World
	config  MatchmakingConfig
	queues  map[MatchmakingMode][]*QueueEntry
	pending map[string]*PendingMatch
	rng     *rand.Rand
}

// NewMatchmakingSystem creates a new matchmaking system.
func NewMatchmakingSystem(world *World, seed int64) *MatchmakingSystem {
	log.WithFields(log.Fields{
		"system_name":          "matchmaking",
		"initial_rating_range": 200,
	}).Debug("Creating matchmaking system")

	return &MatchmakingSystem{
		world:   world,
		config:  DefaultMatchmakingConfig(),
		queues:  make(map[MatchmakingMode][]*QueueEntry),
		pending: make(map[string]*PendingMatch),
		rng:     rand.New(rand.NewSource(seed)),
	}
}

// NewMatchmakingSystemWithConfig creates a matchmaking system with custom config.
func NewMatchmakingSystemWithConfig(world *World, seed int64, config MatchmakingConfig) *MatchmakingSystem {
	log.WithFields(log.Fields{
		"system_name":          "matchmaking",
		"initial_rating_range": config.InitialRatingRange,
	}).Debug("Creating matchmaking system with config")

	return &MatchmakingSystem{
		world:   world,
		config:  config,
		queues:  make(map[MatchmakingMode][]*QueueEntry),
		pending: make(map[string]*PendingMatch),
		rng:     rand.New(rand.NewSource(seed)),
	}
}

// Update processes matchmaking for queued players.
func (s *MatchmakingSystem) Update(entities []*Entity, deltaTime float64) {
	now := time.Now()

	// Process expired pending matches
	s.processExpiredMatches(now)

	// Expand rating ranges for long-waiting players
	s.expandRatingRanges(now)

	// Try to create matches for each mode
	for mode := range s.queues {
		s.tryCreateMatch(mode, now)
	}
}

// AddToQueue adds a player entity to the matchmaking queue.
func (s *MatchmakingSystem) AddToQueue(entity *Entity, mode MatchmakingMode) bool {
	mmComp, ok := entity.GetComponent("matchmaking")
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "matchmaking",
			"entityID":    entity.ID,
			"reason":      "missing_component",
		}).Debug("Cannot add to queue")
		return false
	}
	matchmaking, ok := mmComp.(*MatchmakingComponent)
	if !ok {
		return false
	}

	// Get player rating
	rating := 1000 // Default
	if pvpComp, ok := entity.GetComponent("pvp_rating"); ok {
		if pvpRating, ok := pvpComp.(*PvPRatingComponent); ok {
			rating = pvpRating.Rating
		}
	}

	// Enter queue in component
	if !matchmaking.EnterQueue(mode) {
		return false
	}

	// Create queue entry
	entry := &QueueEntry{
		PlayerID:    entity.ID,
		Rating:      rating,
		Mode:        mode,
		QueuedAt:    time.Now(),
		RatingRange: s.config.InitialRatingRange,
		ServerID:    matchmaking.ServerID,
		TeamID:      matchmaking.TeamID,
	}

	// Add to queue
	s.queues[mode] = append(s.queues[mode], entry)

	log.WithFields(log.Fields{
		"system_name": "matchmaking",
		"entityID":    entity.ID,
		"mode":        mode,
		"rating":      rating,
	}).Info("Player added to queue")

	return true
}

// RemoveFromQueue removes a player entity from the matchmaking queue.
func (s *MatchmakingSystem) RemoveFromQueue(entity *Entity) bool {
	mmComp, ok := entity.GetComponent("matchmaking")
	if !ok {
		return false
	}
	matchmaking, ok := mmComp.(*MatchmakingComponent)
	if !ok {
		return false
	}

	// Leave queue in component
	if !matchmaking.LeaveQueue() {
		return false
	}

	// Remove from queue
	for mode, queue := range s.queues {
		for i, entry := range queue {
			if entry.PlayerID == entity.ID {
				s.queues[mode] = append(queue[:i], queue[i+1:]...)
				log.WithFields(log.Fields{
					"system_name": "matchmaking",
					"entityID":    entity.ID,
					"mode":        mode,
				}).Info("Player removed from queue")
				return true
			}
		}
	}

	return true
}

// AcceptMatch accepts a pending match for a player.
func (s *MatchmakingSystem) AcceptMatch(entity *Entity, matchID string) bool {
	pending, exists := s.pending[matchID]
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "matchmaking",
			"match_id":    matchID,
			"reason":      "match_not_found",
		}).Debug("Cannot accept match")
		return false
	}

	mmComp, ok := entity.GetComponent("matchmaking")
	if !ok {
		return false
	}
	matchmaking, ok := mmComp.(*MatchmakingComponent)
	if !ok {
		return false
	}

	if !matchmaking.AcceptMatch(matchID) {
		return false
	}

	pending.AcceptedBy[entity.ID] = true

	log.WithFields(log.Fields{
		"system_name": "matchmaking",
		"match_id":    matchID,
		"entityID":    entity.ID,
		"accepted":    len(pending.AcceptedBy),
		"required":    len(pending.Participants),
	}).Debug("Player accepted match")

	// Check if all players accepted
	if s.allPlayersAccepted(pending) {
		s.startMatch(pending)
	}

	return true
}

// DeclineMatch declines a pending match for a player.
func (s *MatchmakingSystem) DeclineMatch(entity *Entity, matchID string) bool {
	pending, exists := s.pending[matchID]
	if !exists {
		return false
	}

	mmComp, ok := entity.GetComponent("matchmaking")
	if !ok {
		return false
	}
	matchmaking, ok := mmComp.(*MatchmakingComponent)
	if !ok {
		return false
	}

	matchmaking.DeclineMatch()

	// Cancel the pending match and return other players to queue
	s.cancelPendingMatch(pending, entity.ID)

	log.WithFields(log.Fields{
		"system_name": "matchmaking",
		"match_id":    matchID,
		"entityID":    entity.ID,
	}).Info("Player declined match")

	return true
}

// GetQueueSize returns the current queue size for a mode.
func (s *MatchmakingSystem) GetQueueSize(mode MatchmakingMode) int {
	return len(s.queues[mode])
}

// GetQueueSizes returns queue sizes for all modes.
func (s *MatchmakingSystem) GetQueueSizes() map[MatchmakingMode]int {
	sizes := make(map[MatchmakingMode]int)
	for mode, queue := range s.queues {
		sizes[mode] = len(queue)
	}
	return sizes
}

// GetEstimatedWaitTime estimates queue wait time for a rating.
func (s *MatchmakingSystem) GetEstimatedWaitTime(mode MatchmakingMode, rating int) time.Duration {
	queue := s.queues[mode]
	if len(queue) == 0 {
		return 60 * time.Second // Default estimate
	}

	// Count players within rating range
	inRange := 0
	for _, entry := range queue {
		diff := mmAbs(entry.Rating - rating)
		if diff <= s.config.MaxRatingRange {
			inRange++
		}
	}

	if inRange == 0 {
		return 120 * time.Second // Longer wait expected
	}

	// Rough estimate: more players = shorter wait
	baseWait := 30 * time.Second
	return baseWait / time.Duration(inRange)
}

// tryCreateMatch attempts to create a match for a specific mode.
func (s *MatchmakingSystem) tryCreateMatch(mode MatchmakingMode, now time.Time) {
	queue := s.queues[mode]
	minPlayers := s.config.MinPlayersPerMode[mode]
	maxPlayers := s.config.MaxPlayersPerMode[mode]

	if len(queue) < minPlayers {
		return
	}

	s.sortQueueByRating(queue)

	matched, matchedIndices := s.findCompatiblePlayers(queue, maxPlayers)

	if len(matched) < minPlayers {
		return
	}

	s.createPendingMatch(mode, matched, matchedIndices, now)
}

// sortQueueByRating sorts the matchmaking queue by player rating.
func (s *MatchmakingSystem) sortQueueByRating(queue []*QueueEntry) {
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Rating < queue[j].Rating
	})
}

// findCompatiblePlayers finds players within acceptable rating ranges for a match.
func (s *MatchmakingSystem) findCompatiblePlayers(queue []*QueueEntry, maxPlayers int) ([]*QueueEntry, []int) {
	var matched []*QueueEntry
	var matchedIndices []int

	for i, entry := range queue {
		if len(matched) >= maxPlayers {
			break
		}

		if s.shouldAddPlayerToMatch(entry, matched) {
			matched = append(matched, entry)
			matchedIndices = append(matchedIndices, i)
		}
	}

	return matched, matchedIndices
}

// shouldAddPlayerToMatch checks if a player is compatible with existing matched players.
func (s *MatchmakingSystem) shouldAddPlayerToMatch(entry *QueueEntry, matched []*QueueEntry) bool {
	if len(matched) == 0 {
		return true
	}

	for _, m := range matched {
		if !s.arePlayersCompatible(entry, m) {
			return false
		}
	}

	return true
}

// arePlayersCompatible checks if two players are within acceptable rating range.
func (s *MatchmakingSystem) arePlayersCompatible(entry1, entry2 *QueueEntry) bool {
	diff := mmAbs(entry1.Rating - entry2.Rating)
	maxRange := mmMin(entry1.RatingRange, entry2.RatingRange)
	return diff <= maxRange
}

// createPendingMatch creates a new pending match from matched entries.
func (s *MatchmakingSystem) createPendingMatch(mode MatchmakingMode, entries []*QueueEntry, indices []int, now time.Time) {
	matchID := uuid.New().String()

	// Calculate average rating
	totalRating := 0
	for _, e := range entries {
		totalRating += e.Rating
	}
	avgRating := totalRating / len(entries)

	// Get entities
	participants := make([]*Entity, 0, len(entries))
	for _, entry := range entries {
		// Find entity from world
		if entity, ok := s.world.GetEntity(entry.PlayerID); ok {
			participants = append(participants, entity)
			// Mark as matched in component
			if mmComp, ok := entity.GetComponent("matchmaking"); ok {
				if matchmaking, ok := mmComp.(*MatchmakingComponent); ok {
					matchmaking.MarkMatched(matchID)
				}
			}
		}
	}

	pending := &PendingMatch{
		MatchID:       matchID,
		Mode:          mode,
		Participants:  participants,
		AcceptedBy:    make(map[uint64]bool),
		CreatedAt:     now,
		AverageRating: avgRating,
	}

	s.pending[matchID] = pending

	// Remove from queue (in reverse order to preserve indices)
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	for _, idx := range indices {
		s.queues[mode] = append(s.queues[mode][:idx], s.queues[mode][idx+1:]...)
	}

	log.WithFields(log.Fields{
		"system_name":    "matchmaking",
		"match_id":       matchID,
		"mode":           mode,
		"players":        len(participants),
		"average_rating": avgRating,
	}).Info("Created pending match")
}

// allPlayersAccepted checks if all players in a pending match have accepted.
func (s *MatchmakingSystem) allPlayersAccepted(pending *PendingMatch) bool {
	for _, participant := range pending.Participants {
		if !pending.AcceptedBy[participant.ID] {
			return false
		}
	}
	return true
}

// startMatch starts a fully accepted match.
func (s *MatchmakingSystem) startMatch(pending *PendingMatch) {
	log.WithFields(log.Fields{
		"system_name": "matchmaking",
		"match_id":    pending.MatchID,
		"mode":        pending.Mode,
		"players":     len(pending.Participants),
	}).Info("Starting match")

	// Remove from pending
	delete(s.pending, pending.MatchID)

	// The actual match logic would be handled by a separate combat/arena system
}

// cancelPendingMatch cancels a pending match and returns players to queue.
func (s *MatchmakingSystem) cancelPendingMatch(pending *PendingMatch, declinedByID uint64) {
	for _, participant := range pending.Participants {
		if participant.ID == declinedByID {
			continue
		}

		mmComp, ok := participant.GetComponent("matchmaking")
		if !ok {
			continue
		}
		matchmaking, ok := mmComp.(*MatchmakingComponent)
		if !ok {
			continue
		}

		// Return to idle state
		matchmaking.State = MatchmakingStateIdle
		matchmaking.CurrentMatchID = ""

		// Re-add to queue
		s.AddToQueue(participant, pending.Mode)
	}

	delete(s.pending, pending.MatchID)

	log.WithFields(log.Fields{
		"system_name": "matchmaking",
		"match_id":    pending.MatchID,
		"declined_by": declinedByID,
	}).Info("Pending match cancelled")
}

// processExpiredMatches handles matches that timed out waiting for acceptance.
func (s *MatchmakingSystem) processExpiredMatches(now time.Time) {
	for matchID, pending := range s.pending {
		if now.Sub(pending.CreatedAt) > s.config.MatchAcceptTimeout {
			log.WithFields(log.Fields{
				"system_name": "matchmaking",
				"match_id":    matchID,
			}).Info("Pending match expired")

			// Return all players to queue
			for _, participant := range pending.Participants {
				mmComp, ok := participant.GetComponent("matchmaking")
				if !ok {
					continue
				}
				matchmaking, ok := mmComp.(*MatchmakingComponent)
				if !ok {
					continue
				}

				matchmaking.State = MatchmakingStateIdle
				matchmaking.CurrentMatchID = ""
				s.AddToQueue(participant, pending.Mode)
			}

			delete(s.pending, matchID)
		}
	}
}

// expandRatingRanges expands rating ranges for players waiting in queue.
func (s *MatchmakingSystem) expandRatingRanges(now time.Time) {
	for _, queue := range s.queues {
		for _, entry := range queue {
			waitTime := now.Sub(entry.QueuedAt).Seconds()
			expansion := int(waitTime * s.config.RatingRangeExpansionRate)
			newRange := s.config.InitialRatingRange + expansion
			if newRange > s.config.MaxRatingRange {
				newRange = s.config.MaxRatingRange
			}
			entry.RatingRange = newRange
		}
	}
}

// GetPendingMatchCount returns the number of pending matches.
func (s *MatchmakingSystem) GetPendingMatchCount() int {
	return len(s.pending)
}

// GetPlayerQueuePosition returns the player's position in their queue (1-indexed).
func (s *MatchmakingSystem) GetPlayerQueuePosition(entityID uint64) int {
	for _, queue := range s.queues {
		for i, entry := range queue {
			if entry.PlayerID == entityID {
				return i + 1
			}
		}
	}
	return 0
}

// mmAbs is a helper function for absolute value.
func mmAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// mmMin is a helper function for minimum.
func mmMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
