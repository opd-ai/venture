// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// PvPRatingSystem manages player PvP ratings and rank transitions.
type PvPRatingSystem struct {
	world           *World
	baseKFactor     float64
	decayDays       int
	decayAmount     int
	minRating       int
	lastDecayCheck  time.Time
	decayCheckHours int
}

// NewPvPRatingSystem creates a new PvP rating system.
func NewPvPRatingSystem(world *World) *PvPRatingSystem {
	log.WithFields(log.Fields{
		"system_name": "pvp_rating",
		"k_factor":    32.0,
	}).Debug("Creating PvP rating system")

	return &PvPRatingSystem{
		world:           world,
		baseKFactor:     32.0,
		decayDays:       14,
		decayAmount:     25,
		minRating:       0,
		lastDecayCheck:  time.Now(),
		decayCheckHours: 24,
	}
}

// Update processes rating updates for entities with PvP rating components.
func (s *PvPRatingSystem) Update(entities []*Entity, deltaTime float64) {
	now := time.Now()

	// Check for rating decay periodically
	if now.Sub(s.lastDecayCheck).Hours() >= float64(s.decayCheckHours) {
		s.processRatingDecay(entities, now)
		s.lastDecayCheck = now
	}
}

// processRatingDecay applies rating decay to inactive players.
func (s *PvPRatingSystem) processRatingDecay(entities []*Entity, now time.Time) {
	for _, entity := range entities {
		if !entity.HasComponent("pvp_rating") {
			continue
		}

		comp, ok := entity.GetComponent("pvp_rating")
		if !ok {
			continue
		}
		rating, ok := comp.(*PvPRatingComponent)
		if !ok {
			continue
		}

		// Only apply decay to players above Silver and inactive for decayDays
		if rating.RankTier == RankBronze || rating.RankTier == RankSilver {
			continue
		}

		if rating.LastMatch.IsZero() {
			continue
		}

		daysSinceMatch := int(now.Sub(rating.LastMatch).Hours() / 24)
		if daysSinceMatch < s.decayDays {
			continue
		}

		// Apply decay
		decayPeriods := (daysSinceMatch - s.decayDays) / 7
		if decayPeriods < 1 {
			decayPeriods = 1
		}

		newRating := rating.Rating - (s.decayAmount * decayPeriods)
		if newRating < RankThreshold[RankSilver] {
			newRating = RankThreshold[RankSilver]
		}

		if newRating != rating.Rating {
			log.WithFields(log.Fields{
				"entityID":      entity.ID,
				"old_rating":    rating.Rating,
				"new_rating":    newRating,
				"days_inactive": daysSinceMatch,
			}).Debug("Applied rating decay")

			rating.Rating = newRating
			s.updateRankFromRating(rating)
		}
	}
}

// RecordMatchResult updates ratings for winner and loser after a match.
// Returns the new ratings for winner and loser.
func (s *PvPRatingSystem) RecordMatchResult(winner, loser *Entity, matchTime time.Time) (winnerNewRating, loserNewRating int, err error) {
	winnerComp, ok1 := winner.GetComponent("pvp_rating")
	loserComp, ok2 := loser.GetComponent("pvp_rating")

	if !ok1 || !ok2 || winnerComp == nil || loserComp == nil {
		log.WithFields(log.Fields{
			"winner_id": winner.ID,
			"loser_id":  loser.ID,
		}).Error("Missing PvP rating component")
		return 0, 0, ErrMissingComponent
	}

	winnerRating, ok1 := winnerComp.(*PvPRatingComponent)
	loserRating, ok2 := loserComp.(*PvPRatingComponent)

	if !ok1 || !ok2 {
		log.WithFields(log.Fields{
			"winner_id": winner.ID,
			"loser_id":  loser.ID,
		}).Error("Invalid PvP rating component type")
		return 0, 0, ErrInvalidComponent
	}

	// Calculate K-factor based on experience
	winnerK := s.calculateKFactor(winnerRating)
	loserK := s.calculateKFactor(loserRating)

	// Calculate new ratings
	newWinnerRating, newLoserRating := CalculateELO(
		winnerRating.Rating,
		loserRating.Rating,
		(winnerK+loserK)/2,
	)

	// Update winner
	winnerRating.Rating = newWinnerRating
	winnerRating.Wins++
	winnerRating.LastMatch = matchTime
	if winnerRating.MatchStreak > 0 {
		winnerRating.MatchStreak++
	} else {
		winnerRating.MatchStreak = 1
	}
	if newWinnerRating > winnerRating.PeakRating {
		winnerRating.PeakRating = newWinnerRating
	}
	s.updateRankFromRating(winnerRating)

	// Update loser
	loserRating.Rating = newLoserRating
	if loserRating.Rating < s.minRating {
		loserRating.Rating = s.minRating
	}
	loserRating.Losses++
	loserRating.LastMatch = matchTime
	if loserRating.MatchStreak < 0 {
		loserRating.MatchStreak--
	} else {
		loserRating.MatchStreak = -1
	}
	s.updateRankFromRating(loserRating)

	log.WithFields(log.Fields{
		"winner_id":         winner.ID,
		"loser_id":          loser.ID,
		"winner_old_rating": winnerRating.Rating - (newWinnerRating - winnerRating.Rating),
		"winner_new_rating": newWinnerRating,
		"loser_old_rating":  loserRating.Rating - (newLoserRating - loserRating.Rating),
		"loser_new_rating":  newLoserRating,
	}).Debug("Recorded match result")

	return newWinnerRating, newLoserRating, nil
}

// calculateKFactor determines the K-factor based on player experience.
func (s *PvPRatingSystem) calculateKFactor(rating *PvPRatingComponent) float64 {
	matches := rating.GetTotalMatches()

	// Higher K-factor for new players (faster calibration)
	if matches < 10 {
		return s.baseKFactor * 2.0
	}
	if matches < 30 {
		return s.baseKFactor * 1.5
	}

	// Lower K-factor for high-rated players
	if rating.Rating > 2000 {
		return s.baseKFactor * 0.75
	}

	return s.baseKFactor
}

// updateRankFromRating updates the tier and division based on current rating.
func (s *PvPRatingSystem) updateRankFromRating(rating *PvPRatingComponent) {
	oldTier := rating.RankTier
	oldDivision := rating.RankDivision

	rating.RankTier = GetTierFromRating(rating.Rating)
	rating.RankDivision = GetDivisionFromRating(rating.Rating)

	if oldTier != rating.RankTier || oldDivision != rating.RankDivision {
		log.WithFields(log.Fields{
			"old_rank": string(oldTier) + " " + divisionToRoman(oldDivision),
			"new_rank": string(rating.RankTier) + " " + divisionToRoman(rating.RankDivision),
		}).Debug("Rank changed")
	}
}

// CalculateELO computes new ratings after a match using the ELO algorithm.
// Returns the new ratings for winner and loser.
func CalculateELO(winnerRating, loserRating int, kFactor float64) (newWinner, newLoser int) {
	expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserRating-winnerRating)/400.0))
	expectedLoser := 1.0 - expectedWinner

	newWinner = winnerRating + int(math.Round(kFactor*(1.0-expectedWinner)))
	newLoser = loserRating + int(math.Round(kFactor*(0.0-expectedLoser)))

	return newWinner, newLoser
}

// ResetSeasonRatings resets all player ratings for a new season.
func (s *PvPRatingSystem) ResetSeasonRatings(entities []*Entity, newSeasonID string) {
	for _, entity := range entities {
		if !entity.HasComponent("pvp_rating") {
			continue
		}

		comp, ok := entity.GetComponent("pvp_rating")
		if !ok {
			continue
		}
		rating, ok := comp.(*PvPRatingComponent)
		if !ok {
			continue
		}

		// Soft reset: move rating toward 1000 based on peak
		softResetRating := (rating.PeakRating + 1000) / 2

		rating.Rating = softResetRating
		rating.PeakRating = softResetRating
		rating.Wins = 0
		rating.Losses = 0
		rating.SeasonID = newSeasonID
		rating.MatchStreak = 0
		s.updateRankFromRating(rating)

		log.WithFields(log.Fields{
			"entityID":   entity.ID,
			"new_season": newSeasonID,
			"new_rating": softResetRating,
		}).Debug("Reset season rating")
	}
}

// GetPlayerRank returns the current rank display string for an entity.
func (s *PvPRatingSystem) GetPlayerRank(entity *Entity) string {
	comp, ok := entity.GetComponent("pvp_rating")
	if !ok || comp == nil {
		return "Unranked"
	}

	rating, ok := comp.(*PvPRatingComponent)
	if !ok {
		return "Unranked"
	}

	return rating.GetRankDisplay()
}

// GetLeaderboard returns the top N players sorted by rating.
func (s *PvPRatingSystem) GetLeaderboard(entities []*Entity, limit int) []*Entity {
	var rankedEntities []*Entity

	for _, entity := range entities {
		if entity.HasComponent("pvp_rating") {
			comp, ok := entity.GetComponent("pvp_rating")
			if !ok {
				continue
			}
			if rating, ok := comp.(*PvPRatingComponent); ok {
				if rating.IsPlacementComplete() {
					rankedEntities = append(rankedEntities, entity)
				}
			}
		}
	}

	// Sort by rating descending
	for i := 0; i < len(rankedEntities)-1; i++ {
		for j := i + 1; j < len(rankedEntities); j++ {
			compI, _ := rankedEntities[i].GetComponent("pvp_rating")
			compJ, _ := rankedEntities[j].GetComponent("pvp_rating")
			ratingI := compI.(*PvPRatingComponent)
			ratingJ := compJ.(*PvPRatingComponent)
			if ratingJ.Rating > ratingI.Rating {
				rankedEntities[i], rankedEntities[j] = rankedEntities[j], rankedEntities[i]
			}
		}
	}

	if limit > 0 && limit < len(rankedEntities) {
		return rankedEntities[:limit]
	}
	return rankedEntities
}

// ErrMissingComponent indicates a required component is missing.
var ErrMissingComponent = &componentError{"missing required component"}

// ErrInvalidComponent indicates a component type mismatch.
var ErrInvalidComponent = &componentError{"invalid component type"}

type componentError struct {
	msg string
}

func (e *componentError) Error() string {
	return e.msg
}
