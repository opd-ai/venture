// Package engine provides the achievement notification system.
// AchievementNotificationSystem manages achievement unlock notifications,
// distributes rewards, and integrates with audio for unlock sounds.
//
// Phase 85: Achievement Notifications & Rewards (V15.0)
package engine

import (
	"math/rand"
	"sync"

	"github.com/sirupsen/logrus"
)

// Default reward values per tier.
var defaultTierRewards = map[AchievementTier][]AchievementReward{
	AchievementTierBronze: {
		{Type: AchievementRewardXP, Name: "XP Bonus", Value: 50},
	},
	AchievementTierSilver: {
		{Type: AchievementRewardXP, Name: "XP Bonus", Value: 150},
		{Type: AchievementRewardCurrency, Name: "Gold", Value: 25},
	},
	AchievementTierGold: {
		{Type: AchievementRewardXP, Name: "XP Bonus", Value: 300},
		{Type: AchievementRewardCurrency, Name: "Gold", Value: 75},
	},
	AchievementTierPlatinum: {
		{Type: AchievementRewardXP, Name: "XP Bonus", Value: 500},
		{Type: AchievementRewardCurrency, Name: "Gold", Value: 150},
		{Type: AchievementRewardTitle, Name: "Achievement Title", Value: 1},
	},
}

// AchievementNotificationSystem manages achievement unlock notifications and rewards.
type AchievementNotificationSystem struct {
	world  *World
	logger *logrus.Entry
	mu     sync.RWMutex

	// Custom reward definitions per achievement (optional override).
	customRewards map[string]AchievementRewardDefinition

	// Callback for when rewards are granted.
	onRewardGranted func(entityID uint64, reward AchievementReward)

	// Callback for playing unlock sound.
	onPlayUnlockSound func(entityID uint64, tier AchievementTier)

	// Seed for generating item rewards.
	rewardSeed int64
}

// NewAchievementNotificationSystem creates a new achievement notification system.
func NewAchievementNotificationSystem(world *World) *AchievementNotificationSystem {
	var logger *logrus.Entry
	if world != nil && world.logger != nil {
		logger = world.logger.Logger.WithField("system", "achievement_notification")
	}

	return &AchievementNotificationSystem{
		world:         world,
		logger:        logger,
		customRewards: make(map[string]AchievementRewardDefinition),
		rewardSeed:    12345,
	}
}

// SetRewardSeed sets the seed used for generating item rewards.
func (s *AchievementNotificationSystem) SetRewardSeed(seed int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rewardSeed = seed
}

// SetOnRewardGrantedCallback sets a callback for when rewards are granted.
func (s *AchievementNotificationSystem) SetOnRewardGrantedCallback(callback func(entityID uint64, reward AchievementReward)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onRewardGranted = callback
}

// SetOnPlayUnlockSoundCallback sets a callback for playing unlock sounds.
func (s *AchievementNotificationSystem) SetOnPlayUnlockSoundCallback(callback func(entityID uint64, tier AchievementTier)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onPlayUnlockSound = callback
}

// RegisterCustomRewards registers custom rewards for an achievement.
func (s *AchievementNotificationSystem) RegisterCustomRewards(def AchievementRewardDefinition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customRewards[def.AchievementID] = def
}

// Update processes pending notifications for display.
// This is called each frame to pop and display notifications.
func (s *AchievementNotificationSystem) Update(entities []*Entity, deltaTime float64) {
	// The actual display logic would be handled by the UI layer.
	// This system just manages the queue and grants rewards.
	// UI systems can call PopNotification to get notifications to display.
}

// HandleAchievementUnlock is called when an achievement tier is unlocked.
// This creates a notification, determines rewards, and queues everything.
func (s *AchievementNotificationSystem) HandleAchievementUnlock(entityID uint64, achievementID string, tier AchievementTier) {
	if s.world == nil {
		return
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return
	}

	// Get achievement definition
	def, defExists := GetAchievementDefinition(achievementID)
	if !defExists {
		if s.logger != nil {
			s.logger.WithField("achievement_id", achievementID).Warn("Unknown achievement ID")
		}
		return
	}

	// Calculate rewards
	rewards := s.calculateRewards(achievementID, tier)

	// Create notification
	notification := AchievementNotification{
		AchievementID:   achievementID,
		AchievementName: def.Name,
		Description:     def.Description,
		Category:        def.Category,
		Tier:            tier,
		Points:          tier.Points(),
		Timestamp:       s.world.Clock.Now().Unix(),
		Displayed:       false,
		Rewards:         rewards,
	}

	// Get or create notification component
	comp := s.getOrCreateComponent(entity)

	// Queue notification
	comp.QueueNotification(notification)

	// Grant rewards
	s.grantRewards(entityID, entity, rewards)

	// Play sound if enabled
	if comp.ShouldPlaySound() {
		s.playUnlockSound(entityID, tier)
	}

	// Log the unlock
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"achievement_id": achievementID,
			"tier":           tier.String(),
			"points":         tier.Points(),
			"rewards_count":  len(rewards),
		}).Info("Achievement unlocked and notification queued")
	}
}

// calculateRewards determines the rewards for an achievement tier unlock.
func (s *AchievementNotificationSystem) calculateRewards(achievementID string, tier AchievementTier) []AchievementReward {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for custom rewards
	if custom, exists := s.customRewards[achievementID]; exists {
		if tierRewards, hasRewards := custom.TierRewards[tier]; hasRewards {
			return s.copyRewards(tierRewards)
		}
	}

	// Use default tier rewards
	if defaultRewards, exists := defaultTierRewards[tier]; exists {
		return s.copyRewards(defaultRewards)
	}

	return nil
}

// copyRewards creates a copy of rewards with unique item seeds.
func (s *AchievementNotificationSystem) copyRewards(rewards []AchievementReward) []AchievementReward {
	result := make([]AchievementReward, len(rewards))
	rng := rand.New(rand.NewSource(s.rewardSeed))

	for i, r := range rewards {
		result[i] = r
		if r.Type == AchievementRewardItem && r.ItemSeed == 0 {
			result[i].ItemSeed = rng.Int63()
		}
	}

	return result
}

// grantRewards applies the rewards to the entity.
func (s *AchievementNotificationSystem) grantRewards(entityID uint64, entity *Entity, rewards []AchievementReward) {
	s.mu.RLock()
	callback := s.onRewardGranted
	s.mu.RUnlock()

	for _, reward := range rewards {
		// Apply reward based on type
		switch reward.Type {
		case AchievementRewardXP:
			s.grantXP(entity, reward.Value)
		case AchievementRewardCurrency:
			s.grantCurrency(entity, reward.Value)
		case AchievementRewardTitle:
			s.grantTitle(entity, reward.Name)
		case AchievementRewardItem:
			s.grantItem(entity, reward)
		}

		// Notify callback
		if callback != nil {
			callback(entityID, reward)
		}
	}

	// Update statistics if component exists
	s.updateStatistics(entity, len(rewards))
}

// grantXP adds experience points to the entity.
func (s *AchievementNotificationSystem) grantXP(entity *Entity, amount int) {
	compRaw, exists := entity.GetComponent("experience")
	if !exists {
		return
	}

	if comp, ok := compRaw.(*ExperienceComponent); ok {
		comp.AddXP(amount)
	}
}

// grantCurrency adds gold/currency to the entity.
func (s *AchievementNotificationSystem) grantCurrency(entity *Entity, amount int) {
	compRaw, exists := entity.GetComponent("inventory")
	if !exists {
		return
	}

	if comp, ok := compRaw.(*InventoryComponent); ok {
		comp.Gold += amount
	}
}

// grantTitle adds a title to the entity.
func (s *AchievementNotificationSystem) grantTitle(entity *Entity, title string) {
	// Titles are stored in the extended achievement component as unlocked items
	compRaw, exists := entity.GetComponent("extended_achievement")
	if !exists {
		return
	}

	if comp, ok := compRaw.(*ExtendedAchievementComponent); ok {
		// Store in achievements map as a special entry (title unlocks are tracked)
		_ = comp // Title storage would be implemented in a titles component
	}
}

// grantItem generates and grants an item to the entity.
func (s *AchievementNotificationSystem) grantItem(entity *Entity, reward AchievementReward) {
	// Item granting would integrate with the item generator and inventory system.
	// For now, we log the intent - actual item creation uses pkg/procgen/item.
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"item_name": reward.Name,
			"item_seed": reward.ItemSeed,
			"quantity":  reward.Value,
		}).Debug("Item reward would be granted")
	}
}

// updateStatistics increments the achievement unlock counter.
func (s *AchievementNotificationSystem) updateStatistics(entity *Entity, rewardCount int) {
	compRaw, exists := entity.GetComponent("player_statistics")
	if !exists {
		return
	}

	if comp, ok := compRaw.(*PlayerStatisticsComponent); ok {
		comp.IncrementStat("general_achievements_unlocked", 1)
	}
}

// playUnlockSound triggers the unlock sound effect.
func (s *AchievementNotificationSystem) playUnlockSound(entityID uint64, tier AchievementTier) {
	s.mu.RLock()
	callback := s.onPlayUnlockSound
	s.mu.RUnlock()

	if callback != nil {
		callback(entityID, tier)
	}
}

// getOrCreateComponent gets or creates an AchievementNotificationComponent.
func (s *AchievementNotificationSystem) getOrCreateComponent(entity *Entity) *AchievementNotificationComponent {
	compRaw, exists := entity.GetComponent("achievement_notification")
	if exists {
		if comp, ok := compRaw.(*AchievementNotificationComponent); ok {
			return comp
		}
	}

	comp := NewAchievementNotificationComponent()
	entity.AddComponent(comp)
	return comp
}

// GetPendingNotifications returns the pending notification count for an entity.
func (s *AchievementNotificationSystem) GetPendingNotifications(entityID uint64) int {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("achievement_notification")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*AchievementNotificationComponent)
	if !ok {
		return 0
	}

	return comp.GetPendingCount()
}

// PopNotification returns and removes the next notification for an entity.
func (s *AchievementNotificationSystem) PopNotification(entityID uint64) *AchievementNotification {
	if s.world == nil {
		return nil
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return nil
	}

	compRaw, exists := entity.GetComponent("achievement_notification")
	if !exists {
		return nil
	}

	comp, ok := compRaw.(*AchievementNotificationComponent)
	if !ok {
		return nil
	}

	return comp.PopNotification()
}

// GetTotalAchievementPoints returns the total points for an entity.
func (s *AchievementNotificationSystem) GetTotalAchievementPoints(entityID uint64) int {
	if s.world == nil {
		return 0
	}

	entity, exists := s.world.GetEntity(entityID)
	if !exists || entity == nil {
		return 0
	}

	compRaw, exists := entity.GetComponent("achievement_notification")
	if !exists {
		return 0
	}

	comp, ok := compRaw.(*AchievementNotificationComponent)
	if !ok {
		return 0
	}

	return comp.GetTotalPoints()
}

// GetDefaultTierRewards returns the default rewards for a tier.
func GetDefaultTierRewards(tier AchievementTier) []AchievementReward {
	if rewards, exists := defaultTierRewards[tier]; exists {
		result := make([]AchievementReward, len(rewards))
		copy(result, rewards)
		return result
	}
	return nil
}
