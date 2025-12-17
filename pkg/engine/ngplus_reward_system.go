// Package engine provides the NG+ Reward system for ECS.
// This file implements NGPlusRewardSystem which manages distribution of
// NG+ exclusive content including achievements, items, titles, and challenges.
//
// Phase 114: NG+ Exclusive Content
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// NGPlusRewardSystem manages NG+ exclusive content distribution.
// It monitors player progression and awards exclusive content based on
// NG+ cycle, achievements, and challenge completions.
type NGPlusRewardSystem struct {
	world  *World
	logger *logrus.Entry

	// callbacks for reward events
	onAchievementUnlocked func(entityID uint64, achievementID string)
	onItemAwarded         func(entityID uint64, itemID string)
	onTitleUnlocked       func(entityID uint64, titleID string)
	onChallengeCompleted  func(entityID uint64, challengeID string)

	// seed for deterministic item generation
	seed int64

	// rewardCheckInterval controls how often rewards are checked (in update calls)
	rewardCheckInterval int
	updateCounter       int
}

// NewNGPlusRewardSystem creates a new NG+ reward distribution system.
func NewNGPlusRewardSystem(world *World, seed int64) *NGPlusRewardSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "ngplus_reward")
		logEntry.Debug("NG+ Reward system created")
	}
	return &NGPlusRewardSystem{
		world:               world,
		logger:              logEntry,
		seed:                seed,
		rewardCheckInterval: 60, // Check every 60 updates (~1 second at 60 FPS)
		updateCounter:       0,
	}
}

// Update processes all entities with NG+ components and distributes rewards.
func (s *NGPlusRewardSystem) Update(entities []*Entity, deltaTime float64) {
	s.updateCounter++
	if s.updateCounter < s.rewardCheckInterval {
		return
	}
	s.updateCounter = 0

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks and awards NG+ exclusive content for a single entity.
func (s *NGPlusRewardSystem) processEntity(entity *Entity) {
	// Get NG+ component to check cycle
	ngpComp, ok := entity.GetComponent("newgameplus")
	if !ok {
		return
	}
	ngp, ok := ngpComp.(*NewGamePlusComponent)
	if !ok {
		return
	}

	cycle := ngp.GetCycle()
	if cycle <= 0 {
		return // Not in NG+ yet
	}

	// Get or create reward component
	rewardComp := s.getOrCreateRewardComponent(entity)

	// Check for milestone achievements
	s.checkAchievements(entity, ngp, rewardComp)

	// Check for cycle-based item rewards
	s.checkItemRewards(entity, cycle, rewardComp)

	// Check for title unlocks
	s.checkTitleUnlocks(entity, ngp, rewardComp)

	// Check for dialog variation unlocks
	s.checkDialogVariations(entity, cycle, rewardComp)

	// Update highest tier
	tier := GetTierForCycle(cycle)
	rewardComp.UpdateHighestTier(tier)
}

// getOrCreateRewardComponent gets or creates an NGPlusRewardComponent for the entity.
func (s *NGPlusRewardSystem) getOrCreateRewardComponent(entity *Entity) *NGPlusRewardComponent {
	if comp, ok := entity.GetComponent("ngplus_reward"); ok {
		if reward, ok := comp.(*NGPlusRewardComponent); ok {
			return reward
		}
	}

	reward := NewNGPlusRewardComponent()
	entity.AddComponent(reward)
	return reward
}

// checkAchievements awards cycle-based achievements.
func (s *NGPlusRewardSystem) checkAchievements(entity *Entity, ngp *NewGamePlusComponent, reward *NGPlusRewardComponent) {
	cycle := ngp.GetCycle()
	achievements := GetNGPlusAchievements()

	for _, ach := range achievements {
		// Skip achievements that require special conditions
		if ach.ID == "ngp_speedrun" || ach.ID == "ngp_nodeaths" ||
			ach.ID == "ngp_allbosses" || ach.ID == "ngp_collector" {
			continue // These are handled by their respective systems
		}

		// Check cycle-based achievements
		if cycle >= ach.MinCycle && !reward.HasAchievement(ach.ID) {
			if reward.UnlockAchievement(ach.ID) {
				s.notifyAchievementUnlocked(entity.ID, ach.ID)
			}
		}
	}

	// Check collector achievement
	s.checkCollectorAchievement(entity, reward)
}

// checkCollectorAchievement checks if player has all NG+ exclusive items.
func (s *NGPlusRewardSystem) checkCollectorAchievement(entity *Entity, reward *NGPlusRewardComponent) {
	if reward.HasAchievement("ngp_collector") {
		return
	}

	items := GetNGPlusLegendaryItems()
	for _, item := range items {
		if !reward.HasItem(item.ID) {
			return
		}
	}

	// Has all items - award achievement
	if reward.UnlockAchievement("ngp_collector") {
		s.notifyAchievementUnlocked(entity.ID, "ngp_collector")
	}
}

// checkItemRewards awards NG+ exclusive legendary items at cycle milestones.
func (s *NGPlusRewardSystem) checkItemRewards(entity *Entity, cycle int, reward *NGPlusRewardComponent) {
	items := GetNGPlusLegendaryItems()

	for _, item := range items {
		if cycle >= item.MinCycle && !reward.HasItem(item.ID) {
			// Use deterministic selection for first item of each tier
			if s.shouldAwardItemThisCycle(entity.ID, item.MinCycle, cycle) {
				if reward.AddItem(item.ID) {
					s.notifyItemAwarded(entity.ID, item.ID)
				}
			}
		}
	}
}

// shouldAwardItemThisCycle determines if an item should be awarded this cycle.
// Awards one item per tier on first reaching that tier.
func (s *NGPlusRewardSystem) shouldAwardItemThisCycle(entityID uint64, itemMinCycle, currentCycle int) bool {
	// Award on first cycle that meets the requirement
	return currentCycle == itemMinCycle
}

// checkTitleUnlocks awards NG+ exclusive titles.
func (s *NGPlusRewardSystem) checkTitleUnlocks(entity *Entity, ngp *NewGamePlusComponent, reward *NGPlusRewardComponent) {
	cycle := ngp.GetCycle()
	titles := GetNGPlusTitles()

	for _, title := range titles {
		// Skip special condition titles
		if title.ID == "title_deathless" || title.ID == "title_speedrunner" {
			continue // These require special achievements
		}

		if cycle >= title.MinCycle && !reward.HasTitle(title.ID) {
			if reward.UnlockTitle(title.ID) {
				s.notifyTitleUnlocked(entity.ID, title.ID)
			}
		}
	}

	// Check deathless title
	if reward.HasNoDeathCompletion() && !reward.HasTitle("title_deathless") {
		if reward.UnlockTitle("title_deathless") {
			s.notifyTitleUnlocked(entity.ID, "title_deathless")
		}
	}

	// Check speedrunner title (needs a time attack completion under 2 hours)
	if reward.GetTimeAttackBest("challenge_time_2h") > 0 &&
		reward.GetTimeAttackBest("challenge_time_2h") < 7200000 &&
		!reward.HasTitle("title_speedrunner") {
		if reward.UnlockTitle("title_speedrunner") {
			s.notifyTitleUnlocked(entity.ID, "title_speedrunner")
		}
	}
}

// checkDialogVariations unlocks NPC dialog variations based on cycle.
func (s *NGPlusRewardSystem) checkDialogVariations(entity *Entity, cycle int, reward *NGPlusRewardComponent) {
	// Define dialog variations that unlock at different cycles
	variations := []struct {
		ID       string
		MinCycle int
	}{
		{"npc_blacksmith_ngplus", 1},
		{"npc_merchant_ngplus", 1},
		{"npc_innkeeper_ngplus", 1},
		{"npc_questgiver_ngplus", 2},
		{"npc_sage_ngplus", 3},
		{"npc_king_ngplus", 5},
		{"npc_final_boss_ngplus", 5},
		{"npc_secret_vendor_ngplus", 7},
		{"npc_elder_ngplus", 10},
		{"npc_cosmic_ngplus", 10},
	}

	for _, v := range variations {
		if cycle >= v.MinCycle {
			reward.UnlockNPCDialogVariation(v.ID)
		}
	}
}

// AwardSpeedrunAchievement awards the speedrun achievement if the time qualifies.
func (s *NGPlusRewardSystem) AwardSpeedrunAchievement(entity *Entity, completionTimeMs int64) {
	reward := s.getOrCreateRewardComponent(entity)

	// Record the time
	reward.RecordTimeAttack("challenge_time_2h", completionTimeMs)

	// Check for achievement
	if completionTimeMs < 7200000 && !reward.HasAchievement("ngp_speedrun") {
		if reward.UnlockAchievement("ngp_speedrun") {
			s.notifyAchievementUnlocked(entity.ID, "ngp_speedrun")
		}
	}
}

// AwardNoDeathAchievement awards the no-death achievement if a run was completed.
func (s *NGPlusRewardSystem) AwardNoDeathAchievement(entity *Entity, cycle int) {
	reward := s.getOrCreateRewardComponent(entity)

	// Complete the no-death run
	reward.CompleteNoDeathRun(cycle)

	// Award achievement
	if !reward.HasAchievement("ngp_nodeaths") {
		if reward.UnlockAchievement("ngp_nodeaths") {
			s.notifyAchievementUnlocked(entity.ID, "ngp_nodeaths")
		}
	}
}

// AwardAllBossesAchievement awards the boss slayer achievement.
func (s *NGPlusRewardSystem) AwardAllBossesAchievement(entity *Entity, cycle int) {
	if cycle < 5 {
		return // Requires NG+5 or higher
	}

	reward := s.getOrCreateRewardComponent(entity)

	if !reward.HasAchievement("ngp_allbosses") {
		if reward.UnlockAchievement("ngp_allbosses") {
			s.notifyAchievementUnlocked(entity.ID, "ngp_allbosses")
		}
	}
}

// StartNoDeathChallenge starts a no-death run for the entity.
func (s *NGPlusRewardSystem) StartNoDeathChallenge(entity *Entity, cycle int) {
	reward := s.getOrCreateRewardComponent(entity)
	reward.StartNoDeathRun(cycle)
	reward.ActivateChallenge("challenge_nodeaths")
}

// FailNoDeathChallenge marks a no-death run as failed.
func (s *NGPlusRewardSystem) FailNoDeathChallenge(entity *Entity, cycle int, location string) {
	reward := s.getOrCreateRewardComponent(entity)
	reward.FailNoDeathRun(cycle, location)
}

// GetAvailableChallenges returns challenges available for the current NG+ level.
func (s *NGPlusRewardSystem) GetAvailableChallenges(cycle int) []NGPlusChallengeDefinition {
	allChallenges := GetNGPlusChallenges()
	available := []NGPlusChallengeDefinition{}

	for _, c := range allChallenges {
		if cycle >= c.MinCycle {
			available = append(available, c)
		}
	}

	return available
}

// GetCycleLegendaryItem returns the deterministic legendary item for a cycle.
func (s *NGPlusRewardSystem) GetCycleLegendaryItem(cycle int) string {
	return GenerateDeterministicLegendaryItem(s.seed, cycle)
}

// GetPlayerRewardSummary returns a summary of the player's NG+ rewards.
func (s *NGPlusRewardSystem) GetPlayerRewardSummary(entity *Entity) RewardSummary {
	summary := RewardSummary{}

	rewardComp, ok := entity.GetComponent("ngplus_reward")
	if !ok {
		return summary
	}

	reward, ok := rewardComp.(*NGPlusRewardComponent)
	if !ok {
		return summary
	}

	summary.AchievementCount = len(reward.GetAchievements())
	summary.ItemCount = len(reward.GetItems())
	summary.TitleCount = len(reward.GetTitles())
	summary.ChallengesCompleted = len(reward.GetCompletedChallenges())
	summary.HighestTier = reward.GetHighestTierReached()
	summary.CurrentTitle = reward.GetCurrentTitle()
	summary.HasNoDeathCompletion = reward.HasNoDeathCompletion()

	return summary
}

// RewardSummary contains a summary of NG+ rewards for display.
type RewardSummary struct {
	AchievementCount     int
	ItemCount            int
	TitleCount           int
	ChallengesCompleted  int
	HighestTier          int
	CurrentTitle         string
	HasNoDeathCompletion bool
}

// SetOnAchievementUnlocked sets a callback for achievement unlocks.
func (s *NGPlusRewardSystem) SetOnAchievementUnlocked(callback func(entityID uint64, achievementID string)) {
	s.onAchievementUnlocked = callback
}

// SetOnItemAwarded sets a callback for item awards.
func (s *NGPlusRewardSystem) SetOnItemAwarded(callback func(entityID uint64, itemID string)) {
	s.onItemAwarded = callback
}

// SetOnTitleUnlocked sets a callback for title unlocks.
func (s *NGPlusRewardSystem) SetOnTitleUnlocked(callback func(entityID uint64, titleID string)) {
	s.onTitleUnlocked = callback
}

// SetOnChallengeCompleted sets a callback for challenge completions.
func (s *NGPlusRewardSystem) SetOnChallengeCompleted(callback func(entityID uint64, challengeID string)) {
	s.onChallengeCompleted = callback
}

// notifyAchievementUnlocked calls the achievement callback if set.
func (s *NGPlusRewardSystem) notifyAchievementUnlocked(entityID uint64, achievementID string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"achievement_id": achievementID,
		}).Info("NG+ achievement unlocked")
	}
	if s.onAchievementUnlocked != nil {
		s.onAchievementUnlocked(entityID, achievementID)
	}
}

// notifyItemAwarded calls the item callback if set.
func (s *NGPlusRewardSystem) notifyItemAwarded(entityID uint64, itemID string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"item_id":   itemID,
		}).Info("NG+ item awarded")
	}
	if s.onItemAwarded != nil {
		s.onItemAwarded(entityID, itemID)
	}
}

// notifyTitleUnlocked calls the title callback if set.
func (s *NGPlusRewardSystem) notifyTitleUnlocked(entityID uint64, titleID string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"title_id":  titleID,
		}).Info("NG+ title unlocked")
	}
	if s.onTitleUnlocked != nil {
		s.onTitleUnlocked(entityID, titleID)
	}
}

// notifyChallengeCompleted calls the challenge callback if set.
func (s *NGPlusRewardSystem) notifyChallengeCompleted(entityID uint64, challengeID string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"challenge_id": challengeID,
		}).Info("NG+ challenge completed")
	}
	if s.onChallengeCompleted != nil {
		s.onChallengeCompleted(entityID, challengeID)
	}
}

// CompleteChallenge marks a challenge as completed and awards any related content.
func (s *NGPlusRewardSystem) CompleteChallenge(entity *Entity, challengeID string) {
	reward := s.getOrCreateRewardComponent(entity)

	if reward.CompleteChallenge(challengeID) {
		s.notifyChallengeCompleted(entity.ID, challengeID)

		// Check for related achievements
		switch challengeID {
		case "challenge_nodeaths":
			if !reward.HasAchievement("ngp_nodeaths") {
				if reward.UnlockAchievement("ngp_nodeaths") {
					s.notifyAchievementUnlocked(entity.ID, "ngp_nodeaths")
				}
			}
		case "challenge_time_2h":
			if !reward.HasAchievement("ngp_speedrun") {
				if reward.UnlockAchievement("ngp_speedrun") {
					s.notifyAchievementUnlocked(entity.ID, "ngp_speedrun")
				}
			}
		}
	}
}

// IsNGPlusExclusiveItem returns true if the item ID is an NG+ exclusive.
func IsNGPlusExclusiveItem(itemID string) bool {
	items := GetNGPlusLegendaryItems()
	for _, item := range items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

// GetNGPlusItemDefinition returns the definition for an NG+ item, if it exists.
func GetNGPlusItemDefinition(itemID string) (NGPlusLegendaryItemDefinition, bool) {
	items := GetNGPlusLegendaryItems()
	for _, item := range items {
		if item.ID == itemID {
			return item, true
		}
	}
	return NGPlusLegendaryItemDefinition{}, false
}

// GetNGPlusAchievementDefinition returns the definition for an NG+ achievement.
func GetNGPlusAchievementDefinition(achievementID string) (NGPlusAchievementDefinition, bool) {
	achievements := GetNGPlusAchievements()
	for _, a := range achievements {
		if a.ID == achievementID {
			return a, true
		}
	}
	return NGPlusAchievementDefinition{}, false
}

// GetNGPlusTitleDefinition returns the definition for an NG+ title.
func GetNGPlusTitleDefinition(titleID string) (NGPlusTitleDefinition, bool) {
	titles := GetNGPlusTitles()
	for _, t := range titles {
		if t.ID == titleID {
			return t, true
		}
	}
	return NGPlusTitleDefinition{}, false
}

// GenerateNPCDialogVariation returns NG+ specific dialog for an NPC if available.
// Uses deterministic generation based on seed, NPC ID, and cycle.
func GenerateNPCDialogVariation(seed int64, npcID string, cycle int) string {
	if cycle <= 0 {
		return "" // No variation for first playthrough
	}

	// Generate deterministic variation
	rng := rand.New(rand.NewSource(seed + int64(len(npcID)) + int64(cycle)*100))

	// NG+ dialog variations acknowledge the player's repeated journeys
	variations := []string{
		"Ah, you've returned once more. The cycles continue...",
		"I sense the weight of many journeys upon you, traveler.",
		"Your eyes hold the wisdom of one who has walked this path before.",
		"The threads of fate wind ever tighter around you.",
		"Each cycle reveals new truths, does it not?",
		"Welcome back, eternal challenger. What secrets do you seek this time?",
		"The cosmos remembers your name, wanderer of cycles.",
		"Time loops for those who defy its natural flow.",
		"You carry the burden of infinite rebirths with grace.",
		"The eternal dance of renewal brings you here again.",
	}

	idx := rng.Intn(len(variations))
	return variations[idx]
}

// GetUIIndicatorForNGPlusContent returns a UI indicator string for exclusive content.
func GetUIIndicatorForNGPlusContent(contentType string) string {
	switch contentType {
	case "achievement":
		return "[NG+] "
	case "item":
		return "⟳ "
	case "title":
		return "★ "
	case "challenge":
		return "⚔ "
	default:
		return "[NG+] "
	}
}

// SetRewardCheckInterval sets how often rewards are checked (in update calls).
func (s *NGPlusRewardSystem) SetRewardCheckInterval(interval int) {
	if interval > 0 {
		s.rewardCheckInterval = interval
	}
}

// ForceRewardCheck forces an immediate check of all rewards for an entity.
func (s *NGPlusRewardSystem) ForceRewardCheck(entity *Entity) {
	s.processEntity(entity)
}
