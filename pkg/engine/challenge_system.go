// Package engine provides the ChallengeSystem for managing daily/weekly challenges.
// This file implements the system that generates, tracks, and resets challenges.
//
// Phase 98: Daily/Weekly Challenges (V18.0)
package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ChallengeSystem manages daily and weekly challenge lifecycle.
type ChallengeSystem struct {
	mu sync.RWMutex

	// world is the ECS world reference.
	world *World

	// logger for structured logging.
	logger *logrus.Entry

	// onRewardGranted is called when a reward is granted to a player.
	onRewardGranted func(entityID uint64, reward *ChallengeReward)

	// onChallengeCompleted is called when a challenge is completed.
	onChallengeCompleted func(entityID uint64, challenge *Challenge)

	// onStreakChanged is called when a streak changes.
	onStreakChanged func(entityID uint64, streakType ChallengeType, newStreak int)

	// lastUpdateTime tracks the last update for throttling.
	lastUpdateTime time.Time

	// updateIntervalSeconds controls how often full checks run (default: 1 second).
	updateIntervalSeconds float64
}

// NewChallengeSystem creates a new challenge system.
func NewChallengeSystem(world *World) *ChallengeSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "challenge_system",
	})
	logger.Debug("Creating challenge system")

	return &ChallengeSystem{
		world:                 world,
		logger:                logger,
		updateIntervalSeconds: 1.0,
		lastUpdateTime:        time.Time{},
	}
}

// SetRewardCallback sets the callback for reward grants.
func (s *ChallengeSystem) SetRewardCallback(cb func(entityID uint64, reward *ChallengeReward)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onRewardGranted = cb
}

// SetCompletionCallback sets the callback for challenge completion.
func (s *ChallengeSystem) SetCompletionCallback(cb func(entityID uint64, challenge *Challenge)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onChallengeCompleted = cb
}

// SetStreakCallback sets the callback for streak changes.
func (s *ChallengeSystem) SetStreakCallback(cb func(entityID uint64, streakType ChallengeType, newStreak int)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onStreakChanged = cb
}

// Update processes challenge resets and generates new challenges.
func (s *ChallengeSystem) Update(entities []*Entity, deltaTime float64) {
	currentTime := time.Now()

	// Throttle updates
	if currentTime.Sub(s.lastUpdateTime).Seconds() < s.updateIntervalSeconds {
		return
	}
	s.lastUpdateTime = currentTime

	for _, entity := range entities {
		if !entity.HasComponent("daily_challenge") {
			continue
		}

		compRaw, ok := entity.GetComponent("daily_challenge")
		if !ok {
			continue
		}
		comp, ok := compRaw.(*DailyChallengeComponent)
		if !ok {
			continue
		}

		// Check for daily reset
		dailyReset := comp.CheckDailyReset(currentTime)
		if dailyReset {
			s.logger.WithFields(logrus.Fields{
				"entityID":     entity.ID,
				"daily_streak": comp.GetDailyStreak(),
			}).Debug("Daily challenges reset")

			// Generate new daily challenges
			dailyChallenges := comp.GenerateDailyChallenges(currentTime)
			comp.SetActiveChallenges(dailyChallenges, comp.GetActiveWeeklyChallenges())

			s.mu.RLock()
			if s.onStreakChanged != nil {
				s.onStreakChanged(entity.ID, ChallengeTypeDaily, comp.GetDailyStreak())
			}
			s.mu.RUnlock()
		}

		// Check for weekly reset
		weeklyReset := comp.CheckWeeklyReset(currentTime)
		if weeklyReset {
			s.logger.WithFields(logrus.Fields{
				"entityID":      entity.ID,
				"weekly_streak": comp.GetWeeklyStreak(),
			}).Debug("Weekly challenges reset")

			// Generate new weekly challenges
			weeklyChallenges := comp.GenerateWeeklyChallenges(currentTime)
			comp.SetActiveChallenges(comp.GetActiveDailyChallenges(), weeklyChallenges)

			s.mu.RLock()
			if s.onStreakChanged != nil {
				s.onStreakChanged(entity.ID, ChallengeTypeWeekly, comp.GetWeeklyStreak())
			}
			s.mu.RUnlock()
		}

		// Initialize challenges if needed
		if len(comp.GetActiveDailyChallenges()) == 0 {
			dailyChallenges := comp.GenerateDailyChallenges(currentTime)
			comp.SetActiveChallenges(dailyChallenges, comp.GetActiveWeeklyChallenges())
		}
		if len(comp.GetActiveWeeklyChallenges()) == 0 {
			weeklyChallenges := comp.GenerateWeeklyChallenges(currentTime)
			comp.SetActiveChallenges(comp.GetActiveDailyChallenges(), weeklyChallenges)
		}
	}
}

// TrackProgress updates challenge progress for an entity based on a tracking key.
// Returns true if progress was updated.
func (s *ChallengeSystem) TrackProgress(entityID uint64, trackingKey string, amount int) bool {
	if s.world == nil {
		return false
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return false
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return false
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return false
	}

	updated := comp.UpdateProgress(trackingKey, amount)
	if updated != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":     entityID,
			"challenge_id": updated.ID,
			"progress":     updated.Progress,
			"target":       updated.Target,
			"tracking_key": trackingKey,
		}).Debug("Challenge progress updated")

		if updated.IsCompleted {
			s.mu.RLock()
			if s.onChallengeCompleted != nil {
				s.onChallengeCompleted(entityID, updated)
			}
			s.mu.RUnlock()

			s.logger.WithFields(logrus.Fields{
				"entityID":     entityID,
				"challenge_id": updated.ID,
				"name":         updated.Name,
			}).Info("Challenge completed")
		}
		return true
	}
	return false
}

// ClaimReward claims the reward for a completed challenge.
// Returns the reward with bonuses applied, or nil if cannot claim.
func (s *ChallengeSystem) ClaimReward(entityID uint64, challengeID string) *ChallengeReward {
	if s.world == nil {
		return nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return nil
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return nil
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return nil
	}

	reward := comp.ClaimReward(challengeID)
	if reward != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":     entityID,
			"challenge_id": challengeID,
			"xp":           reward.XP,
			"gold":         reward.Gold,
			"bonus":        reward.BonusMultiplier,
		}).Info("Challenge reward claimed")

		s.mu.RLock()
		if s.onRewardGranted != nil {
			s.onRewardGranted(entityID, reward)
		}
		s.mu.RUnlock()
	}
	return reward
}

// RerollChallenge rerolls a daily challenge for an entity.
// Returns the new challenge or nil if reroll failed.
func (s *ChallengeSystem) RerollChallenge(entityID uint64, challengeID string) *Challenge {
	if s.world == nil {
		return nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return nil
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return nil
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return nil
	}

	newChallenge := comp.RerollChallenge(challengeID, time.Now())
	if newChallenge != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":      entityID,
			"old_challenge": challengeID,
			"new_challenge": newChallenge.ID,
			"rerolls_left":  comp.GetRerollsRemaining(),
		}).Debug("Challenge rerolled")
	}
	return newChallenge
}

// GetEntityChallenges returns all active challenges for an entity.
func (s *ChallengeSystem) GetEntityChallenges(entityID uint64) (daily, weekly []*Challenge) {
	if s.world == nil {
		return nil, nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return nil, nil
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return nil, nil
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return nil, nil
	}

	return comp.GetActiveDailyChallenges(), comp.GetActiveWeeklyChallenges()
}

// GetEntityStreak returns the current streak for an entity.
func (s *ChallengeSystem) GetEntityStreak(entityID uint64, streakType ChallengeType) int {
	if s.world == nil {
		return 0
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return 0
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return 0
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return 0
	}

	if streakType == ChallengeTypeDaily {
		return comp.GetDailyStreak()
	}
	return comp.GetWeeklyStreak()
}

// GetCompletionPercent returns the completion percentage for an entity's challenges.
func (s *ChallengeSystem) GetCompletionPercent(entityID uint64, challengeType ChallengeType) float64 {
	if s.world == nil {
		return 0
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return 0
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return 0
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return 0
	}

	if challengeType == ChallengeTypeDaily {
		return comp.GetDailyCompletionPercent()
	}
	return comp.GetWeeklyCompletionPercent()
}

// InitializePlayerChallenges sets up challenges for a new player entity.
func (s *ChallengeSystem) InitializePlayerChallenges(entityID uint64, baseSeed int64) error {
	if s.world == nil {
		return fmt.Errorf("world is nil")
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	// Create component if not exists
	if !entity.HasComponent("daily_challenge") {
		comp := NewDailyChallengeComponent(baseSeed)
		entity.AddComponent(comp)

		// Generate initial challenges
		currentTime := time.Now()
		dailyChallenges := comp.GenerateDailyChallenges(currentTime)
		weeklyChallenges := comp.GenerateWeeklyChallenges(currentTime)
		comp.SetActiveChallenges(dailyChallenges, weeklyChallenges)
		comp.LastDailyReset = currentTime.Unix()
		comp.LastWeeklyReset = currentTime.Unix()

		s.logger.WithFields(logrus.Fields{
			"entityID":     entityID,
			"daily_count":  len(dailyChallenges),
			"weekly_count": len(weeklyChallenges),
		}).Info("Player challenges initialized")
	}

	return nil
}

// GetChallengeByID returns a specific challenge by ID for an entity.
func (s *ChallengeSystem) GetChallengeByID(entityID uint64, challengeID string) *Challenge {
	if s.world == nil {
		return nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return nil
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return nil
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return nil
	}

	for _, ch := range comp.GetActiveDailyChallenges() {
		if ch.ID == challengeID {
			return ch
		}
	}
	for _, ch := range comp.GetActiveWeeklyChallenges() {
		if ch.ID == challengeID {
			return ch
		}
	}
	return nil
}

// GetTotalStats returns lifetime statistics for an entity's challenges.
func (s *ChallengeSystem) GetTotalStats(entityID uint64) (totalCompleted, totalXP, totalGold int) {
	if s.world == nil {
		return 0, 0, 0
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || !entity.HasComponent("daily_challenge") {
		return 0, 0, 0
	}

	compRaw, ok := entity.GetComponent("daily_challenge")
	if !ok {
		return 0, 0, 0
	}
	comp, ok := compRaw.(*DailyChallengeComponent)
	if !ok {
		return 0, 0, 0
	}

	comp.mu.RLock()
	defer comp.mu.RUnlock()

	return comp.TotalChallengesCompleted, comp.TotalXPEarned, comp.TotalGoldEarned
}

// TrackCombatProgress is a convenience method for combat-related progress.
func (s *ChallengeSystem) TrackCombatProgress(entityID uint64, enemiesKilled, damageDealt, criticalHits, bossesKilled int) {
	if enemiesKilled > 0 {
		s.TrackProgress(entityID, "enemies_killed", enemiesKilled)
	}
	if damageDealt > 0 {
		s.TrackProgress(entityID, "damage_dealt", damageDealt)
	}
	if criticalHits > 0 {
		s.TrackProgress(entityID, "critical_hits", criticalHits)
	}
	if bossesKilled > 0 {
		s.TrackProgress(entityID, "bosses_killed", bossesKilled)
	}
}

// TrackGatheringProgress is a convenience method for gathering-related progress.
func (s *ChallengeSystem) TrackGatheringProgress(entityID uint64, resourcesGathered, fishCaught, herbsGathered, rareResources int) {
	if resourcesGathered > 0 {
		s.TrackProgress(entityID, "resources_gathered", resourcesGathered)
	}
	if fishCaught > 0 {
		s.TrackProgress(entityID, "fish_caught", fishCaught)
	}
	if herbsGathered > 0 {
		s.TrackProgress(entityID, "herbs_gathered", herbsGathered)
	}
	if rareResources > 0 {
		s.TrackProgress(entityID, "rare_resources_gathered", rareResources)
	}
}

// TrackExplorationProgress is a convenience method for exploration-related progress.
func (s *ChallengeSystem) TrackExplorationProgress(entityID uint64, areasDiscovered, distanceTraveled, secretsFound, dungeonsCompleted int) {
	if areasDiscovered > 0 {
		s.TrackProgress(entityID, "areas_discovered", areasDiscovered)
	}
	if distanceTraveled > 0 {
		s.TrackProgress(entityID, "distance_traveled", distanceTraveled)
	}
	if secretsFound > 0 {
		s.TrackProgress(entityID, "secrets_found", secretsFound)
	}
	if dungeonsCompleted > 0 {
		s.TrackProgress(entityID, "dungeons_completed", dungeonsCompleted)
	}
}

// TrackSocialProgress is a convenience method for social-related progress.
func (s *ChallengeSystem) TrackSocialProgress(entityID uint64, tradesCompleted, messagesSent, playersHelped, guildActivities int) {
	if tradesCompleted > 0 {
		s.TrackProgress(entityID, "trades_completed", tradesCompleted)
	}
	if messagesSent > 0 {
		s.TrackProgress(entityID, "messages_sent", messagesSent)
	}
	if playersHelped > 0 {
		s.TrackProgress(entityID, "players_helped", playersHelped)
	}
	if guildActivities > 0 {
		s.TrackProgress(entityID, "guild_activities_completed", guildActivities)
	}
}

// TrackCraftingProgress is a convenience method for crafting-related progress.
func (s *ChallengeSystem) TrackCraftingProgress(entityID uint64, itemsCrafted, qualityItemsCrafted, recipesUsed, rareItemsCrafted int) {
	if itemsCrafted > 0 {
		s.TrackProgress(entityID, "items_crafted", itemsCrafted)
	}
	if qualityItemsCrafted > 0 {
		s.TrackProgress(entityID, "quality_items_crafted", qualityItemsCrafted)
	}
	if recipesUsed > 0 {
		s.TrackProgress(entityID, "recipes_used", recipesUsed)
	}
	if rareItemsCrafted > 0 {
		s.TrackProgress(entityID, "rare_items_crafted", rareItemsCrafted)
	}
}
