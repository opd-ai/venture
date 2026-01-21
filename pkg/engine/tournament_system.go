// Package engine provides the tournament system for competitive PvP events.
// TournamentSystem manages tournament lifecycle, brackets, and progression.
package engine

import (
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// TournamentSystem manages scheduled competitive tournaments.
type TournamentSystem struct {
	world                *World
	activeTournaments    map[string]*TournamentInstance
	scheduledTournaments []*TournamentInstance
	definitions          []TournamentDefinition
	rng                  *rand.Rand
	lastScheduleCheck    time.Time
}

// NewTournamentSystem creates a new tournament system.
func NewTournamentSystem(world *World, seed int64) *TournamentSystem {
	log.WithFields(log.Fields{
		"system_name": "tournament",
	}).Debug("Creating tournament system")

	return &TournamentSystem{
		world:                world,
		activeTournaments:    make(map[string]*TournamentInstance),
		scheduledTournaments: make([]*TournamentInstance, 0),
		definitions:          GetDefaultTournamentDefinitions(seed),
		rng:                  rand.New(rand.NewSource(seed)),
		lastScheduleCheck:    time.Now(),
	}
}

// Update processes tournament state changes.
func (s *TournamentSystem) Update(entities []*Entity, deltaTime float64) {
	now := time.Now()

	// Check for tournaments to start
	s.processScheduledTournaments(now)

	// Update active tournaments
	for _, tournament := range s.activeTournaments {
		s.updateTournament(tournament, now)
	}

	// Schedule new tournaments periodically
	if now.Sub(s.lastScheduleCheck) > time.Hour {
		s.scheduleRecurringTournaments(now)
		s.lastScheduleCheck = now
	}
}

// CreateTournament creates a new tournament instance.
func (s *TournamentSystem) CreateTournament(def TournamentDefinition, startTime time.Time) *TournamentInstance {
	instance := &TournamentInstance{
		InstanceID:   uuid.New().String(),
		Definition:   def,
		Phase:        TournamentPhaseRegistration,
		Participants: make([]uint64, 0),
		Seeds:        make(map[uint64]int),
		Bracket:      make([]BracketMatch, 0),
		CurrentRound: 0,
		CreatedAt:    time.Now(),
		StartsAt:     startTime,
	}

	s.scheduledTournaments = append(s.scheduledTournaments, instance)

	log.WithFields(log.Fields{
		"system_name":   "tournament",
		"tournament_id": instance.InstanceID,
		"name":          def.Name,
		"starts_at":     startTime.Format(time.RFC3339),
	}).Info("Created tournament")

	return instance
}

// RegisterPlayer registers a player for a tournament.
func (s *TournamentSystem) RegisterPlayer(entity *Entity, tournamentID string) bool {
	tournament := s.findTournamentByID(tournamentID)
	if tournament == nil {
		log.WithFields(log.Fields{
			"system_name":   "tournament",
			"tournament_id": tournamentID,
			"reason":        "not_found",
		}).Debug("Cannot register player")
		return false
	}

	if !s.validateRegistrationPhase(tournament, tournamentID) {
		return false
	}

	if !s.checkTournamentCapacity(tournament, tournamentID) {
		return false
	}

	if s.isPlayerAlreadyRegistered(tournament, entity.ID) {
		return false
	}

	if !s.validateRatingRequirement(tournament, entity, tournamentID) {
		return false
	}

	tc, ok := s.getTournamentComponent(entity)
	if !ok {
		return false
	}

	s.registerPlayerInTournament(tournament, entity, tournamentID, tc)

	return true
}

// findTournamentByID searches for a tournament in scheduled and active lists.
func (s *TournamentSystem) findTournamentByID(tournamentID string) *TournamentInstance {
	for _, t := range s.scheduledTournaments {
		if t.InstanceID == tournamentID {
			return t
		}
	}

	if t, exists := s.activeTournaments[tournamentID]; exists {
		return t
	}

	return nil
}

// validateRegistrationPhase checks if the tournament is in registration phase.
func (s *TournamentSystem) validateRegistrationPhase(tournament *TournamentInstance, tournamentID string) bool {
	if tournament.Phase != TournamentPhaseRegistration {
		log.WithFields(log.Fields{
			"system_name":   "tournament",
			"tournament_id": tournamentID,
			"phase":         tournament.Phase,
			"reason":        "wrong_phase",
		}).Debug("Cannot register player")
		return false
	}
	return true
}

// checkTournamentCapacity verifies the tournament has not reached max participants.
func (s *TournamentSystem) checkTournamentCapacity(tournament *TournamentInstance, tournamentID string) bool {
	if len(tournament.Participants) >= tournament.Definition.MaxParticipants {
		log.WithFields(log.Fields{
			"system_name":   "tournament",
			"tournament_id": tournamentID,
			"reason":        "full",
		}).Debug("Cannot register player")
		return false
	}
	return true
}

// isPlayerAlreadyRegistered checks if the player is already in the tournament.
func (s *TournamentSystem) isPlayerAlreadyRegistered(tournament *TournamentInstance, entityID uint64) bool {
	for _, p := range tournament.Participants {
		if p == entityID {
			return true
		}
	}
	return false
}

// validateRatingRequirement checks if the player meets the rating requirement.
func (s *TournamentSystem) validateRatingRequirement(tournament *TournamentInstance, entity *Entity, tournamentID string) bool {
	if tournament.Definition.RatingRequirement <= 0 {
		return true
	}

	pvpComp, ok := entity.GetComponent("pvp_rating")
	if !ok {
		return true
	}

	pvpRating, ok := pvpComp.(*PvPRatingComponent)
	if !ok {
		return true
	}

	if pvpRating.Rating < tournament.Definition.RatingRequirement {
		log.WithFields(log.Fields{
			"system_name":   "tournament",
			"tournament_id": tournamentID,
			"rating":        pvpRating.Rating,
			"required":      tournament.Definition.RatingRequirement,
			"reason":        "rating_too_low",
		}).Debug("Cannot register player")
		return false
	}

	return true
}

// getTournamentComponent retrieves and validates the tournament component from an entity.
func (s *TournamentSystem) getTournamentComponent(entity *Entity) (*TournamentComponent, bool) {
	tournComp, ok := entity.GetComponent("tournament")
	if !ok {
		return nil, false
	}
	tc, ok := tournComp.(*TournamentComponent)
	if !ok {
		return nil, false
	}
	return tc, true
}

// registerPlayerInTournament performs the actual registration and seeding.
func (s *TournamentSystem) registerPlayerInTournament(tournament *TournamentInstance, entity *Entity, tournamentID string, tc *TournamentComponent) {
	tournament.Participants = append(tournament.Participants, entity.ID)

	seed := s.calculatePlayerSeed(tournament, entity)
	tournament.Seeds[entity.ID] = seed

	tc.EnterTournament(tournamentID, seed)

	log.WithFields(log.Fields{
		"system_name":   "tournament",
		"tournament_id": tournamentID,
		"entityID":      entity.ID,
		"participants":  len(tournament.Participants),
	}).Info("Player registered for tournament")
}

// calculatePlayerSeed determines the seed value for a player based on rating or registration order.
func (s *TournamentSystem) calculatePlayerSeed(tournament *TournamentInstance, entity *Entity) int {
	seed := len(tournament.Participants)
	if pvpComp, ok := entity.GetComponent("pvp_rating"); ok {
		if pvpRating, ok := pvpComp.(*PvPRatingComponent); ok {
			seed = pvpRating.Rating
		}
	}
	return seed
}

// UnregisterPlayer removes a player from a tournament.
func (s *TournamentSystem) UnregisterPlayer(entity *Entity, tournamentID string) bool {
	var tournament *TournamentInstance
	for _, t := range s.scheduledTournaments {
		if t.InstanceID == tournamentID {
			tournament = t
			break
		}
	}

	if tournament == nil {
		return false
	}

	if tournament.Phase != TournamentPhaseRegistration {
		return false
	}

	// Remove from participants
	for i, p := range tournament.Participants {
		if p == entity.ID {
			tournament.Participants = append(tournament.Participants[:i], tournament.Participants[i+1:]...)
			delete(tournament.Seeds, entity.ID)
			break
		}
	}

	// Update component
	tournComp, ok := entity.GetComponent("tournament")
	if ok {
		if tc, ok := tournComp.(*TournamentComponent); ok {
			tc.LeaveTournament()
		}
	}

	log.WithFields(log.Fields{
		"system_name":   "tournament",
		"tournament_id": tournamentID,
		"entityID":      entity.ID,
	}).Info("Player unregistered from tournament")

	return true
}

// RecordMatchResult records the outcome of a tournament match.
func (s *TournamentSystem) RecordMatchResult(tournamentID, matchID string, winnerID, loserID uint64) bool {
	tournament, exists := s.activeTournaments[tournamentID]
	if !exists {
		return false
	}

	match := s.findMatch(tournament, matchID)
	if match == nil || match.Completed {
		return false
	}

	s.updateMatchResults(match, winnerID, loserID)
	s.updatePlayerComponents(tournament, winnerID, loserID, match)
	s.advanceWinner(tournament, match)
	s.logMatchResult(tournamentID, matchID, winnerID, loserID)
	s.checkTournamentComplete(tournament)

	return true
}

// findMatch searches for a match in both main and losers brackets.
func (s *TournamentSystem) findMatch(tournament *TournamentInstance, matchID string) *BracketMatch {
	for i := range tournament.Bracket {
		if tournament.Bracket[i].MatchID == matchID {
			return &tournament.Bracket[i]
		}
	}

	for i := range tournament.LosersBracket {
		if tournament.LosersBracket[i].MatchID == matchID {
			return &tournament.LosersBracket[i]
		}
	}

	return nil
}

// updateMatchResults records the winner and loser for a completed match.
func (s *TournamentSystem) updateMatchResults(match *BracketMatch, winnerID, loserID uint64) {
	match.WinnerID = winnerID
	match.LoserID = loserID
	match.Completed = true
	match.CompletedAt = time.Now()
}

// updatePlayerComponents updates tournament components for winner and loser.
func (s *TournamentSystem) updatePlayerComponents(tournament *TournamentInstance, winnerID, loserID uint64, match *BracketMatch) {
	isDoubleElim := tournament.Definition.Format == TournamentFormatDoubleElim

	if winner, ok := s.world.GetEntity(winnerID); ok {
		if tc, ok := winner.GetComponent("tournament"); ok {
			if tournComp, ok := tc.(*TournamentComponent); ok {
				tournComp.RecordWin()
			}
		}
	}

	if loser, ok := s.world.GetEntity(loserID); ok {
		if tc, ok := loser.GetComponent("tournament"); ok {
			if tournComp, ok := tc.(*TournamentComponent); ok {
				tournComp.RecordLoss(isDoubleElim)

				if isDoubleElim && tournComp.LossesInTournament == 1 {
					s.moveToLosersBracket(tournament, loserID, match.Round)
				}
			}
		}
	}
}

// logMatchResult logs the tournament match result.
func (s *TournamentSystem) logMatchResult(tournamentID, matchID string, winnerID, loserID uint64) {
	log.WithFields(log.Fields{
		"system_name":   "tournament",
		"tournament_id": tournamentID,
		"match_id":      matchID,
		"winner":        winnerID,
		"loser":         loserID,
	}).Info("Recorded tournament match result")
}

// GetActiveTournaments returns all active tournaments.
func (s *TournamentSystem) GetActiveTournaments() []*TournamentInstance {
	result := make([]*TournamentInstance, 0, len(s.activeTournaments))
	for _, t := range s.activeTournaments {
		result = append(result, t)
	}
	return result
}

// GetScheduledTournaments returns all scheduled tournaments.
func (s *TournamentSystem) GetScheduledTournaments() []*TournamentInstance {
	return s.scheduledTournaments
}

// GetTournament returns a tournament by ID.
func (s *TournamentSystem) GetTournament(tournamentID string) *TournamentInstance {
	if t, exists := s.activeTournaments[tournamentID]; exists {
		return t
	}
	for _, t := range s.scheduledTournaments {
		if t.InstanceID == tournamentID {
			return t
		}
	}
	return nil
}

// GetPlayerMatches returns upcoming/pending matches for a player.
func (s *TournamentSystem) GetPlayerMatches(entityID uint64) []BracketMatch {
	var matches []BracketMatch

	for _, tournament := range s.activeTournaments {
		for _, match := range tournament.Bracket {
			if !match.Completed && (match.Player1ID == entityID || match.Player2ID == entityID) {
				matches = append(matches, match)
			}
		}
		for _, match := range tournament.LosersBracket {
			if !match.Completed && (match.Player1ID == entityID || match.Player2ID == entityID) {
				matches = append(matches, match)
			}
		}
	}

	return matches
}

// processScheduledTournaments starts tournaments that are ready.
func (s *TournamentSystem) processScheduledTournaments(now time.Time) {
	remaining := make([]*TournamentInstance, 0)

	for _, tournament := range s.scheduledTournaments {
		if now.After(tournament.StartsAt) {
			if len(tournament.Participants) >= tournament.Definition.MinParticipants {
				s.startTournament(tournament)
			} else {
				// Cancel if not enough participants
				tournament.Phase = TournamentPhaseCancelled
				s.notifyParticipantsCancelled(tournament)
				log.WithFields(log.Fields{
					"system_name":   "tournament",
					"tournament_id": tournament.InstanceID,
					"participants":  len(tournament.Participants),
					"required":      tournament.Definition.MinParticipants,
				}).Info("Tournament cancelled - insufficient participants")
			}
		} else {
			remaining = append(remaining, tournament)
		}
	}

	s.scheduledTournaments = remaining
}

// startTournament begins tournament competition.
func (s *TournamentSystem) startTournament(tournament *TournamentInstance) {
	tournament.Phase = TournamentPhaseSeeding

	// Sort participants by seed (rating)
	sorted := make([]uint64, len(tournament.Participants))
	copy(sorted, tournament.Participants)
	sort.Slice(sorted, func(i, j int) bool {
		return tournament.Seeds[sorted[i]] > tournament.Seeds[sorted[j]]
	})

	// Assign final seeds
	for i, pid := range sorted {
		tournament.Seeds[pid] = i + 1
	}

	// Generate bracket
	if tournament.Definition.Format == TournamentFormatDoubleElim {
		tournament.Bracket, tournament.LosersBracket = GenerateDoubleElimBracket(
			tournament.Definition.Seed,
			sorted,
		)
	} else {
		tournament.Bracket = GenerateSingleElimBracket(
			tournament.Definition.Seed,
			sorted,
		)
	}

	tournament.TotalRounds = CalculateTotalRounds(len(sorted))
	tournament.CurrentRound = 1
	tournament.Phase = TournamentPhaseInProgress

	s.activeTournaments[tournament.InstanceID] = tournament

	log.WithFields(log.Fields{
		"system_name":   "tournament",
		"tournament_id": tournament.InstanceID,
		"participants":  len(sorted),
		"rounds":        tournament.TotalRounds,
	}).Info("Tournament started")
}

// updateTournament processes tournament state.
func (s *TournamentSystem) updateTournament(tournament *TournamentInstance, now time.Time) {
	if tournament.Phase != TournamentPhaseInProgress {
		return
	}

	// Check for matches that should auto-progress (byes, timeouts)
	for i := range tournament.Bracket {
		match := &tournament.Bracket[i]
		if match.Completed {
			continue
		}

		// Auto-complete byes
		if match.Player1ID == 0 && match.Player2ID != 0 {
			match.WinnerID = match.Player2ID
			match.Completed = true
			match.CompletedAt = now
			s.advanceWinner(tournament, match)
		} else if match.Player2ID == 0 && match.Player1ID != 0 {
			match.WinnerID = match.Player1ID
			match.Completed = true
			match.CompletedAt = now
			s.advanceWinner(tournament, match)
		}
	}
}

// advanceWinner moves the winner to the next round.
func (s *TournamentSystem) advanceWinner(tournament *TournamentInstance, match *BracketMatch) {
	if match.WinnerID == 0 {
		return
	}

	// Find next match in bracket
	nextRound := match.Round + 1
	if nextRound > tournament.TotalRounds {
		return
	}

	nextPos := match.Position / 2

	for i := range tournament.Bracket {
		if tournament.Bracket[i].Round == nextRound && tournament.Bracket[i].Position == nextPos {
			if match.Position%2 == 0 {
				tournament.Bracket[i].Player1ID = match.WinnerID
			} else {
				tournament.Bracket[i].Player2ID = match.WinnerID
			}
			break
		}
	}
}

// moveToLosersBracket moves a player to the losers bracket.
func (s *TournamentSystem) moveToLosersBracket(tournament *TournamentInstance, playerID uint64, fromRound int) {
	if len(tournament.LosersBracket) == 0 {
		return
	}

	// Find appropriate losers bracket match
	for i := range tournament.LosersBracket {
		match := &tournament.LosersBracket[i]
		if match.Player1ID == 0 {
			match.Player1ID = playerID
			return
		}
		if match.Player2ID == 0 {
			match.Player2ID = playerID
			return
		}
	}
}

// checkTournamentComplete checks if the tournament is finished.
func (s *TournamentSystem) checkTournamentComplete(tournament *TournamentInstance) {
	// Check if all matches are complete
	allComplete := true
	for _, match := range tournament.Bracket {
		if !match.Completed {
			allComplete = false
			break
		}
	}

	if !allComplete {
		return
	}

	// For double elimination, also check losers bracket
	if tournament.Definition.Format == TournamentFormatDoubleElim {
		for _, match := range tournament.LosersBracket {
			if !match.Completed && (match.Player1ID != 0 || match.Player2ID != 0) {
				allComplete = false
				break
			}
		}
	}

	if !allComplete {
		return
	}

	// Determine winner (last match winner)
	if len(tournament.Bracket) > 0 {
		finalMatch := tournament.Bracket[len(tournament.Bracket)-1]
		tournament.WinnerID = finalMatch.WinnerID
		tournament.RunnerUpID = finalMatch.LoserID
	}

	tournament.Phase = TournamentPhaseComplete
	tournament.EndsAt = time.Now()

	// Notify participants and update their components
	s.completeTournamentForPlayers(tournament)

	log.WithFields(log.Fields{
		"system_name":   "tournament",
		"tournament_id": tournament.InstanceID,
		"winner":        tournament.WinnerID,
		"runner_up":     tournament.RunnerUpID,
	}).Info("Tournament completed")

	// Remove from active
	delete(s.activeTournaments, tournament.InstanceID)
}

// completeTournamentForPlayers updates player components with results.
func (s *TournamentSystem) completeTournamentForPlayers(tournament *TournamentInstance) {
	placements := s.calculatePlacements(tournament)

	for _, playerID := range tournament.Participants {
		entity, ok := s.world.GetEntity(playerID)
		if !ok {
			continue
		}

		tc, ok := entity.GetComponent("tournament")
		if !ok {
			continue
		}

		tournComp, ok := tc.(*TournamentComponent)
		if !ok {
			continue
		}

		placement := placements[playerID]
		result := TournamentResult{
			TournamentID:      tournament.InstanceID,
			TournamentName:    tournament.Definition.Name,
			Placement:         placement,
			TotalParticipants: len(tournament.Participants),
			MatchesWon:        tournComp.MatchesWon,
			MatchesLost:       tournComp.LossesInTournament,
			CompletedAt:       tournament.EndsAt,
		}

		tournComp.CompleteTournament(result)
	}
}

// calculatePlacements determines final placement for each player.
func (s *TournamentSystem) calculatePlacements(tournament *TournamentInstance) map[uint64]int {
	placements := make(map[uint64]int)

	if tournament.WinnerID != 0 {
		placements[tournament.WinnerID] = 1
	}
	if tournament.RunnerUpID != 0 {
		placements[tournament.RunnerUpID] = 2
	}

	// Calculate other placements based on when eliminated
	for _, match := range tournament.Bracket {
		if match.LoserID != 0 {
			if _, exists := placements[match.LoserID]; !exists {
				// Placement based on round eliminated
				placement := 1 << (tournament.TotalRounds - match.Round + 1)
				placements[match.LoserID] = placement
			}
		}
	}

	// Default placement for those not assigned
	nextPlacement := len(tournament.Participants)
	for _, pid := range tournament.Participants {
		if _, exists := placements[pid]; !exists {
			placements[pid] = nextPlacement
			nextPlacement--
		}
	}

	return placements
}

// notifyParticipantsCancelled clears tournament state for participants.
func (s *TournamentSystem) notifyParticipantsCancelled(tournament *TournamentInstance) {
	for _, playerID := range tournament.Participants {
		entity, ok := s.world.GetEntity(playerID)
		if !ok {
			continue
		}

		tc, ok := entity.GetComponent("tournament")
		if !ok {
			continue
		}

		if tournComp, ok := tc.(*TournamentComponent); ok {
			tournComp.LeaveTournament()
		}
	}
}

// scheduleRecurringTournaments creates new tournament instances.
func (s *TournamentSystem) scheduleRecurringTournaments(now time.Time) {
	for _, def := range s.definitions {
		// Check if we need a new instance
		exists := false
		for _, scheduled := range s.scheduledTournaments {
			if scheduled.Definition.ID == def.ID {
				exists = true
				break
			}
		}

		if !exists {
			// Calculate next start time
			var startTime time.Time
			switch def.Frequency {
			case TournamentFrequencyDaily:
				// Next day at noon
				startTime = time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, now.Location())
			case TournamentFrequencyWeekly:
				// Next Sunday at noon
				daysUntilSunday := (7 - int(now.Weekday())) % 7
				if daysUntilSunday == 0 && now.Hour() >= 12 {
					daysUntilSunday = 7
				}
				startTime = time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, 12, 0, 0, 0, now.Location())
			case TournamentFrequencyMonthly:
				// First of next month at noon
				startTime = time.Date(now.Year(), now.Month()+1, 1, 12, 0, 0, 0, now.Location())
			default:
				continue
			}

			s.CreateTournament(def, startTime)
		}
	}
}

// GetTournamentCount returns the number of active and scheduled tournaments.
func (s *TournamentSystem) GetTournamentCount() (active, scheduled int) {
	return len(s.activeTournaments), len(s.scheduledTournaments)
}

// CancelTournament cancels a tournament before it starts.
func (s *TournamentSystem) CancelTournament(tournamentID string) bool {
	for i, t := range s.scheduledTournaments {
		if t.InstanceID == tournamentID && t.Phase == TournamentPhaseRegistration {
			t.Phase = TournamentPhaseCancelled
			s.notifyParticipantsCancelled(t)
			s.scheduledTournaments = append(s.scheduledTournaments[:i], s.scheduledTournaments[i+1:]...)
			log.WithFields(log.Fields{
				"system_name":   "tournament",
				"tournament_id": tournamentID,
			}).Info("Tournament cancelled")
			return true
		}
	}
	return false
}

// AddTournamentDefinition adds a custom tournament type.
func (s *TournamentSystem) AddTournamentDefinition(def TournamentDefinition) {
	s.definitions = append(s.definitions, def)
}
