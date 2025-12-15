// Package engine provides achievement notification tracking for the UI layer.
// AchievementNotificationComponent queues achievement unlock notifications
// for display and manages their lifecycle.
//
// Phase 85: Achievement Notifications & Rewards (V15.0)
package engine

import (
	"encoding/json"
	"sync"
)

// AchievementNotification represents a queued achievement unlock notification.
type AchievementNotification struct {
	// AchievementID is the achievement that was unlocked.
	AchievementID string `json:"achievement_id"`
	// AchievementName is the display name.
	AchievementName string `json:"achievement_name"`
	// Description is the achievement description.
	Description string `json:"description"`
	// Category is the achievement category.
	Category AchievementCategory `json:"category"`
	// Tier is the tier that was unlocked.
	Tier AchievementTier `json:"tier"`
	// Points is the achievement points earned for this tier.
	Points int `json:"points"`
	// Timestamp is when the achievement was unlocked (Unix seconds).
	Timestamp int64 `json:"timestamp"`
	// Displayed indicates if this notification has been shown.
	Displayed bool `json:"displayed"`
	// Rewards is the list of rewards granted.
	Rewards []AchievementReward `json:"rewards,omitempty"`
}

// AchievementRewardType categorizes achievement rewards.
type AchievementRewardType string

const (
	// AchievementRewardXP represents experience point rewards.
	AchievementRewardXP AchievementRewardType = "xp"
	// AchievementRewardItem represents item rewards.
	AchievementRewardItem AchievementRewardType = "item"
	// AchievementRewardTitle represents title/cosmetic rewards.
	AchievementRewardTitle AchievementRewardType = "title"
	// AchievementRewardCurrency represents gold/currency rewards.
	AchievementRewardCurrency AchievementRewardType = "currency"
)

// AchievementReward represents a reward granted for achievement unlock.
type AchievementReward struct {
	// Type categorizes the reward.
	Type AchievementRewardType `json:"type"`
	// Name is the display name of the reward.
	Name string `json:"name"`
	// Value is the numerical value (XP amount, item quantity, gold).
	Value int `json:"value"`
	// ItemSeed is used for procedural item generation (if Type is item).
	ItemSeed int64 `json:"item_seed,omitempty"`
}

// AchievementRewardDefinition defines rewards for each achievement tier.
type AchievementRewardDefinition struct {
	// AchievementID is the achievement this reward definition is for.
	AchievementID string
	// TierRewards maps each tier to its rewards.
	TierRewards map[AchievementTier][]AchievementReward
}

// AchievementNotificationComponent queues and tracks achievement notifications.
type AchievementNotificationComponent struct {
	mu sync.RWMutex
	// PendingNotifications is the queue of notifications to display.
	PendingNotifications []AchievementNotification `json:"pending_notifications"`
	// DisplayedNotifications is the history of displayed notifications.
	DisplayedNotifications []AchievementNotification `json:"displayed_notifications"`
	// TotalAchievementPoints is the cumulative points from all achievements.
	TotalAchievementPoints int `json:"total_achievement_points"`
	// PlaySoundOnUnlock indicates if a sound should play on unlock.
	PlaySoundOnUnlock bool `json:"play_sound_on_unlock"`
	// MaxHistorySize limits stored notification history.
	MaxHistorySize int `json:"max_history_size"`
}

// Type returns the component type identifier.
func (c *AchievementNotificationComponent) Type() string {
	return "achievement_notification"
}

// NewAchievementNotificationComponent creates a new achievement notification component.
func NewAchievementNotificationComponent() *AchievementNotificationComponent {
	return &AchievementNotificationComponent{
		PendingNotifications:   make([]AchievementNotification, 0),
		DisplayedNotifications: make([]AchievementNotification, 0),
		PlaySoundOnUnlock:      true,
		MaxHistorySize:         100,
	}
}

// QueueNotification adds a notification to the pending queue.
func (c *AchievementNotificationComponent) QueueNotification(notification AchievementNotification) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.PendingNotifications = append(c.PendingNotifications, notification)
	c.TotalAchievementPoints += notification.Points
}

// PopNotification removes and returns the next pending notification.
// Returns nil if the queue is empty.
func (c *AchievementNotificationComponent) PopNotification() *AchievementNotification {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.PendingNotifications) == 0 {
		return nil
	}

	notification := c.PendingNotifications[0]
	c.PendingNotifications = c.PendingNotifications[1:]
	notification.Displayed = true

	// Add to history, respecting max size
	c.DisplayedNotifications = append(c.DisplayedNotifications, notification)
	if len(c.DisplayedNotifications) > c.MaxHistorySize {
		c.DisplayedNotifications = c.DisplayedNotifications[1:]
	}

	return &notification
}

// PeekNotification returns the next pending notification without removing it.
func (c *AchievementNotificationComponent) PeekNotification() *AchievementNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.PendingNotifications) == 0 {
		return nil
	}

	notification := c.PendingNotifications[0]
	return &notification
}

// GetPendingCount returns the number of pending notifications.
func (c *AchievementNotificationComponent) GetPendingCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.PendingNotifications)
}

// GetHistoryCount returns the number of displayed notifications in history.
func (c *AchievementNotificationComponent) GetHistoryCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.DisplayedNotifications)
}

// GetTotalPoints returns the total achievement points earned.
func (c *AchievementNotificationComponent) GetTotalPoints() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TotalAchievementPoints
}

// SetPlaySound enables or disables unlock sounds.
func (c *AchievementNotificationComponent) SetPlaySound(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PlaySoundOnUnlock = enabled
}

// ShouldPlaySound returns whether unlock sounds are enabled.
func (c *AchievementNotificationComponent) ShouldPlaySound() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PlaySoundOnUnlock
}

// ClearPending removes all pending notifications.
func (c *AchievementNotificationComponent) ClearPending() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PendingNotifications = make([]AchievementNotification, 0)
}

// GetRecentHistory returns the most recent N displayed notifications.
func (c *AchievementNotificationComponent) GetRecentHistory(count int) []AchievementNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()

	historyLen := len(c.DisplayedNotifications)
	if count > historyLen {
		count = historyLen
	}
	if count <= 0 {
		return nil
	}

	result := make([]AchievementNotification, count)
	copy(result, c.DisplayedNotifications[historyLen-count:])
	return result
}

// GetHistoryByCategory returns displayed notifications for a specific category.
func (c *AchievementNotificationComponent) GetHistoryByCategory(category AchievementCategory) []AchievementNotification {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []AchievementNotification
	for _, n := range c.DisplayedNotifications {
		if n.Category == category {
			result = append(result, n)
		}
	}
	return result
}

// Serialize converts the component to JSON bytes for persistence.
func (c *AchievementNotificationComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return json.Marshal(struct {
		PendingNotifications   []AchievementNotification `json:"pending_notifications"`
		DisplayedNotifications []AchievementNotification `json:"displayed_notifications"`
		TotalAchievementPoints int                       `json:"total_achievement_points"`
		PlaySoundOnUnlock      bool                      `json:"play_sound_on_unlock"`
		MaxHistorySize         int                       `json:"max_history_size"`
	}{
		PendingNotifications:   c.PendingNotifications,
		DisplayedNotifications: c.DisplayedNotifications,
		TotalAchievementPoints: c.TotalAchievementPoints,
		PlaySoundOnUnlock:      c.PlaySoundOnUnlock,
		MaxHistorySize:         c.MaxHistorySize,
	})
}

// Deserialize restores the component from JSON bytes.
func (c *AchievementNotificationComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var temp struct {
		PendingNotifications   []AchievementNotification `json:"pending_notifications"`
		DisplayedNotifications []AchievementNotification `json:"displayed_notifications"`
		TotalAchievementPoints int                       `json:"total_achievement_points"`
		PlaySoundOnUnlock      bool                      `json:"play_sound_on_unlock"`
		MaxHistorySize         int                       `json:"max_history_size"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	c.PendingNotifications = temp.PendingNotifications
	c.DisplayedNotifications = temp.DisplayedNotifications
	c.TotalAchievementPoints = temp.TotalAchievementPoints
	c.PlaySoundOnUnlock = temp.PlaySoundOnUnlock
	c.MaxHistorySize = temp.MaxHistorySize

	if c.PendingNotifications == nil {
		c.PendingNotifications = make([]AchievementNotification, 0)
	}
	if c.DisplayedNotifications == nil {
		c.DisplayedNotifications = make([]AchievementNotification, 0)
	}
	if c.MaxHistorySize == 0 {
		c.MaxHistorySize = 100
	}

	return nil
}
