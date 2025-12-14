// Package engine provides the event reward component for seasonal event rewards.
// EventRewardComponent tracks rewards earned by players during seasonal events,
// including exclusive items, achievements, cosmetic rewards, and event currency.
package engine

import (
	"encoding/json"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// EventRewardType categorizes the kind of reward.
type EventRewardType string

const (
	// EventRewardItem represents an event-exclusive item reward.
	EventRewardItem EventRewardType = "item"
	// EventRewardCurrency represents event currency (tokens, tickets).
	EventRewardCurrency EventRewardType = "currency"
	// EventRewardTitle represents a cosmetic title reward.
	EventRewardTitle EventRewardType = "title"
	// EventRewardEffect represents a visual effect reward.
	EventRewardEffect EventRewardType = "effect"
	// EventRewardAchievement represents an event achievement unlock.
	EventRewardAchievement EventRewardType = "achievement"
)

// EventReward represents a reward earned during an event.
type EventReward struct {
	// ID is the unique identifier for this reward
	ID string `json:"id"`
	// EventID links this reward to a specific event
	EventID string `json:"event_id"`
	// Type categorizes the reward
	Type EventRewardType `json:"type"`
	// Name is the display name
	Name string `json:"name"`
	// Description explains the reward
	Description string `json:"description"`
	// Value is the numerical value (currency amount, item quantity)
	Value int `json:"value"`
	// ItemSeed is the seed for generating exclusive items
	ItemSeed int64 `json:"item_seed,omitempty"`
	// Rarity indicates how rare the reward is
	Rarity string `json:"rarity"`
	// Claimed indicates if the reward has been claimed
	Claimed bool `json:"claimed"`
}

// EventAchievementDef defines an achievement for event participation.
type EventAchievementDef struct {
	// ID is the unique identifier
	ID string `json:"id"`
	// EventID links to the event (empty for cross-event achievements)
	EventID string `json:"event_id,omitempty"`
	// Name is the achievement name
	Name string `json:"name"`
	// Description explains the achievement
	Description string `json:"description"`
	// Requirement is what must be done (e.g., "complete_quests", "earn_currency")
	Requirement string `json:"requirement"`
	// RequiredAmount is the target number
	RequiredAmount int `json:"required_amount"`
	// Reward is granted when achievement is completed
	Reward EventReward `json:"reward"`
}

// EventCosmeticTitle represents an earnable title.
type EventCosmeticTitle struct {
	// ID is the unique identifier
	ID string `json:"id"`
	// EventID links to the event
	EventID string `json:"event_id"`
	// DisplayName is the title text
	DisplayName string `json:"display_name"`
	// Rarity indicates how rare the title is
	Rarity string `json:"rarity"`
}

// EventVisualEffect represents an earnable visual effect.
type EventVisualEffect struct {
	// ID is the unique identifier
	ID string `json:"id"`
	// EventID links to the event
	EventID string `json:"event_id"`
	// Name is the effect name
	Name string `json:"name"`
	// EffectType categorizes the effect (aura, trail, glow)
	EffectType string `json:"effect_type"`
	// ColorHex is the primary color
	ColorHex string `json:"color_hex"`
	// Intensity controls effect strength (0.0-1.0)
	Intensity float64 `json:"intensity"`
}

// EventRewardProgress tracks progress toward achievements.
type EventRewardProgress struct {
	// AchievementID is the achievement being tracked
	AchievementID string `json:"achievement_id"`
	// CurrentProgress is the current value
	CurrentProgress int `json:"current_progress"`
	// RequiredAmount is the target
	RequiredAmount int `json:"required_amount"`
	// Completed indicates if requirement is met
	Completed bool `json:"completed"`
}

// EventRewardComponent tracks event rewards for a player entity.
// It manages earned rewards, event currency, achievements, and cosmetics.
type EventRewardComponent struct {
	// EarnedRewards contains all rewards earned from events
	EarnedRewards []EventReward `json:"earned_rewards"`
	// EventCurrency tracks currency per event (eventID -> amount)
	EventCurrency map[string]int `json:"event_currency"`
	// EarnedTitles contains unlocked cosmetic titles
	EarnedTitles []EventCosmeticTitle `json:"earned_titles"`
	// ActiveTitle is the currently displayed title
	ActiveTitle string `json:"active_title"`
	// EarnedEffects contains unlocked visual effects
	EarnedEffects []EventVisualEffect `json:"earned_effects"`
	// ActiveEffect is the currently displayed effect
	ActiveEffect string `json:"active_effect"`
	// AchievementProgress tracks progress toward achievements
	AchievementProgress []EventRewardProgress `json:"achievement_progress"`
	// CompletedAchievements contains finished achievement IDs
	CompletedAchievements []string `json:"completed_achievements"`
	// TotalEventsParticipated is lifetime event count
	TotalEventsParticipated int `json:"total_events_participated"`
	// TotalQuestsCompleted is lifetime event quests completed
	TotalQuestsCompleted int `json:"total_quests_completed"`
	// TotalCurrencyEarned is lifetime currency earned across all events
	TotalCurrencyEarned int `json:"total_currency_earned"`
}

// NewEventRewardComponent creates a new event reward component.
func NewEventRewardComponent() *EventRewardComponent {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
	}).Debug("Creating event reward component")

	return &EventRewardComponent{
		EarnedRewards:         make([]EventReward, 0),
		EventCurrency:         make(map[string]int),
		EarnedTitles:          make([]EventCosmeticTitle, 0),
		EarnedEffects:         make([]EventVisualEffect, 0),
		AchievementProgress:   make([]EventRewardProgress, 0),
		CompletedAchievements: make([]string, 0),
	}
}

// Type returns the component type identifier.
func (c *EventRewardComponent) Type() string {
	return "event_reward"
}

// AddCurrency adds event currency for a specific event.
func (c *EventRewardComponent) AddCurrency(eventID string, amount int) {
	if amount <= 0 {
		return
	}

	c.EventCurrency[eventID] += amount
	c.TotalCurrencyEarned += amount

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"event_id":       eventID,
		"amount":         amount,
		"new_total":      c.EventCurrency[eventID],
	}).Debug("Added event currency")
}

// SpendCurrency spends event currency for a specific event.
// Returns true if successful, false if insufficient funds.
func (c *EventRewardComponent) SpendCurrency(eventID string, amount int) bool {
	if amount <= 0 {
		return false
	}

	current := c.EventCurrency[eventID]
	if current < amount {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_reward",
			"event_id":       eventID,
			"requested":      amount,
			"available":      current,
		}).Debug("Insufficient event currency")
		return false
	}

	c.EventCurrency[eventID] -= amount

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"event_id":       eventID,
		"spent":          amount,
		"remaining":      c.EventCurrency[eventID],
	}).Debug("Spent event currency")

	return true
}

// GetCurrency returns the current currency for an event.
func (c *EventRewardComponent) GetCurrency(eventID string) int {
	return c.EventCurrency[eventID]
}

// AddReward adds an earned reward.
func (c *EventRewardComponent) AddReward(reward EventReward) {
	c.EarnedRewards = append(c.EarnedRewards, reward)

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"reward_id":      reward.ID,
		"event_id":       reward.EventID,
		"reward_type":    reward.Type,
		"reward_name":    reward.Name,
	}).Info("Earned event reward")
}

// ClaimReward marks a reward as claimed.
func (c *EventRewardComponent) ClaimReward(rewardID string) bool {
	for i := range c.EarnedRewards {
		if c.EarnedRewards[i].ID == rewardID && !c.EarnedRewards[i].Claimed {
			c.EarnedRewards[i].Claimed = true

			logrus.WithFields(logrus.Fields{
				"component_type": "event_reward",
				"reward_id":      rewardID,
			}).Info("Claimed event reward")
			return true
		}
	}
	return false
}

// GetUnclaimedRewards returns all unclaimed rewards.
func (c *EventRewardComponent) GetUnclaimedRewards() []EventReward {
	unclaimed := make([]EventReward, 0)
	for _, r := range c.EarnedRewards {
		if !r.Claimed {
			unclaimed = append(unclaimed, r)
		}
	}
	return unclaimed
}

// GetRewardsForEvent returns all rewards for a specific event.
func (c *EventRewardComponent) GetRewardsForEvent(eventID string) []EventReward {
	result := make([]EventReward, 0)
	for _, r := range c.EarnedRewards {
		if r.EventID == eventID {
			result = append(result, r)
		}
	}
	return result
}

// AddTitle unlocks a cosmetic title.
func (c *EventRewardComponent) AddTitle(title EventCosmeticTitle) {
	// Check if already has this title
	for _, t := range c.EarnedTitles {
		if t.ID == title.ID {
			return
		}
	}

	c.EarnedTitles = append(c.EarnedTitles, title)

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"title_id":       title.ID,
		"display_name":   title.DisplayName,
	}).Info("Unlocked event title")
}

// SetActiveTitle sets the displayed title.
func (c *EventRewardComponent) SetActiveTitle(titleID string) bool {
	// Check if player has this title
	for _, t := range c.EarnedTitles {
		if t.ID == titleID {
			c.ActiveTitle = titleID
			return true
		}
	}
	return false
}

// GetActiveTitle returns the active title, if any.
func (c *EventRewardComponent) GetActiveTitle() *EventCosmeticTitle {
	for i := range c.EarnedTitles {
		if c.EarnedTitles[i].ID == c.ActiveTitle {
			return &c.EarnedTitles[i]
		}
	}
	return nil
}

// AddEffect unlocks a visual effect.
func (c *EventRewardComponent) AddEffect(effect EventVisualEffect) {
	// Check if already has this effect
	for _, e := range c.EarnedEffects {
		if e.ID == effect.ID {
			return
		}
	}

	c.EarnedEffects = append(c.EarnedEffects, effect)

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"effect_id":      effect.ID,
		"effect_name":    effect.Name,
	}).Info("Unlocked event effect")
}

// SetActiveEffect sets the displayed visual effect.
func (c *EventRewardComponent) SetActiveEffect(effectID string) bool {
	// Check if player has this effect
	for _, e := range c.EarnedEffects {
		if e.ID == effectID {
			c.ActiveEffect = effectID
			return true
		}
	}
	return false
}

// GetActiveEffect returns the active visual effect, if any.
func (c *EventRewardComponent) GetActiveEffect() *EventVisualEffect {
	for i := range c.EarnedEffects {
		if c.EarnedEffects[i].ID == c.ActiveEffect {
			return &c.EarnedEffects[i]
		}
	}
	return nil
}

// UpdateAchievementProgress updates progress for an achievement.
func (c *EventRewardComponent) UpdateAchievementProgress(achievementID string, progress, required int) bool {
	// Find existing progress
	for i := range c.AchievementProgress {
		if c.AchievementProgress[i].AchievementID == achievementID {
			c.AchievementProgress[i].CurrentProgress = progress
			if progress >= required && !c.AchievementProgress[i].Completed {
				c.AchievementProgress[i].Completed = true
				return true // Newly completed
			}
			return false
		}
	}

	// Create new progress entry
	newProgress := EventRewardProgress{
		AchievementID:   achievementID,
		CurrentProgress: progress,
		RequiredAmount:  required,
		Completed:       progress >= required,
	}
	c.AchievementProgress = append(c.AchievementProgress, newProgress)

	return newProgress.Completed
}

// IncrementAchievementProgress increments progress for an achievement.
func (c *EventRewardComponent) IncrementAchievementProgress(achievementID string, amount, required int) bool {
	// Find existing progress
	for i := range c.AchievementProgress {
		if c.AchievementProgress[i].AchievementID == achievementID {
			c.AchievementProgress[i].CurrentProgress += amount
			if c.AchievementProgress[i].CurrentProgress > required {
				c.AchievementProgress[i].CurrentProgress = required
			}
			if c.AchievementProgress[i].CurrentProgress >= required && !c.AchievementProgress[i].Completed {
				c.AchievementProgress[i].Completed = true
				return true
			}
			return false
		}
	}

	// Create new progress entry
	newProgress := EventRewardProgress{
		AchievementID:   achievementID,
		CurrentProgress: amount,
		RequiredAmount:  required,
		Completed:       amount >= required,
	}
	c.AchievementProgress = append(c.AchievementProgress, newProgress)

	return newProgress.Completed
}

// CompleteAchievement marks an achievement as completed.
func (c *EventRewardComponent) CompleteAchievement(achievementID string) bool {
	// Check if already completed
	for _, id := range c.CompletedAchievements {
		if id == achievementID {
			return false
		}
	}

	c.CompletedAchievements = append(c.CompletedAchievements, achievementID)

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"achievement_id": achievementID,
	}).Info("Completed event achievement")

	return true
}

// HasAchievement checks if an achievement is completed.
func (c *EventRewardComponent) HasAchievement(achievementID string) bool {
	for _, id := range c.CompletedAchievements {
		if id == achievementID {
			return true
		}
	}
	return false
}

// GetAchievementProgress returns progress for a specific achievement.
func (c *EventRewardComponent) GetAchievementProgress(achievementID string) *EventRewardProgress {
	for i := range c.AchievementProgress {
		if c.AchievementProgress[i].AchievementID == achievementID {
			return &c.AchievementProgress[i]
		}
	}
	return nil
}

// RecordEventParticipation records that the player participated in an event.
func (c *EventRewardComponent) RecordEventParticipation() {
	c.TotalEventsParticipated++
}

// RecordQuestCompletion records that a player completed an event quest.
func (c *EventRewardComponent) RecordQuestCompletion() {
	c.TotalQuestsCompleted++
}

// Serialize encodes the component to bytes for persistence.
func (c *EventRewardComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type":   "event_reward",
		"rewards_count":    len(c.EarnedRewards),
		"titles_count":     len(c.EarnedTitles),
		"effects_count":    len(c.EarnedEffects),
		"achievements":     len(c.CompletedAchievements),
		"total_events":     c.TotalEventsParticipated,
		"total_quests":     c.TotalQuestsCompleted,
		"total_currency":   c.TotalCurrencyEarned,
		"currency_entries": len(c.EventCurrency),
	}).Debug("Serializing event reward component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_reward",
			"error":          err.Error(),
		}).Error("Failed to serialize event reward component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *EventRewardComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"bytes":          len(data),
	}).Debug("Deserializing event reward component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_reward",
			"error":          err.Error(),
		}).Error("Failed to deserialize event reward component")
		return err
	}

	// Initialize maps if nil after deserialization
	if c.EventCurrency == nil {
		c.EventCurrency = make(map[string]int)
	}

	return nil
}

// GenerateEventRewards generates rewards for a seasonal event.
// Creates currency rewards, item rewards, and cosmetic rewards.
func GenerateEventRewards(event EventInstance, seed int64) []EventReward {
	rng := rand.New(rand.NewSource(seed))
	rewards := make([]EventReward, 0, 6)

	// Generate currency rewards
	rewards = append(rewards, generateCurrencyReward(rng, event, "small", 10, 25))
	rewards = append(rewards, generateCurrencyReward(rng, event, "medium", 50, 100))
	rewards = append(rewards, generateCurrencyReward(rng, event, "large", 150, 300))

	// Generate item reward
	rewards = append(rewards, generateItemReward(rng, event))

	// Generate title reward
	rewards = append(rewards, generateTitleReward(rng, event))

	// Generate effect reward
	rewards = append(rewards, generateEffectReward(rng, event))

	logrus.WithFields(logrus.Fields{
		"component_type": "event_reward",
		"event_id":       event.Definition.ID,
		"event_name":     event.Definition.Name,
		"rewards_count":  len(rewards),
		"seed":           seed,
	}).Debug("Generated event rewards")

	return rewards
}

// generateCurrencyReward creates a currency reward.
func generateCurrencyReward(rng *rand.Rand, event EventInstance, tier string, minVal, maxVal int) EventReward {
	amount := minVal + rng.Intn(maxVal-minVal+1)
	currencyNames := getThemeCurrencyName(event.Definition.Theme)

	return EventReward{
		ID:          event.Definition.ID + "_currency_" + tier,
		EventID:     event.Definition.ID,
		Type:        EventRewardCurrency,
		Name:        currencyNames[0] + " (" + tier + ")",
		Description: "Event currency for the " + event.Definition.Name,
		Value:       amount,
		Rarity:      tier,
	}
}

// generateItemReward creates an event-exclusive item reward.
func generateItemReward(rng *rand.Rand, event EventInstance) EventReward {
	itemNames := getThemeItemRewards(event.Definition.Theme)
	itemName := itemNames[rng.Intn(len(itemNames))]

	return EventReward{
		ID:          event.Definition.ID + "_item",
		EventID:     event.Definition.ID,
		Type:        EventRewardItem,
		Name:        itemName,
		Description: "An exclusive item from the " + event.Definition.Name,
		Value:       1,
		ItemSeed:    rng.Int63(),
		Rarity:      "rare",
	}
}

// generateTitleReward creates a cosmetic title reward.
func generateTitleReward(rng *rand.Rand, event EventInstance) EventReward {
	titles := getThemeTitles(event.Definition.Theme)
	title := titles[rng.Intn(len(titles))]

	return EventReward{
		ID:          event.Definition.ID + "_title",
		EventID:     event.Definition.ID,
		Type:        EventRewardTitle,
		Name:        title,
		Description: "A title earned during the " + event.Definition.Name,
		Value:       1,
		Rarity:      "epic",
	}
}

// generateEffectReward creates a visual effect reward.
func generateEffectReward(rng *rand.Rand, event EventInstance) EventReward {
	effects := getThemeEffects(event.Definition.Theme)
	effect := effects[rng.Intn(len(effects))]

	return EventReward{
		ID:          event.Definition.ID + "_effect",
		EventID:     event.Definition.ID,
		Type:        EventRewardEffect,
		Name:        effect,
		Description: "A visual effect from the " + event.Definition.Name,
		Value:       1,
		Rarity:      "legendary",
	}
}

// getThemeCurrencyName returns currency names themed for the event.
func getThemeCurrencyName(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"Blossom Token", "Spring Petal"}
	case EventThemeSummer:
		return []string{"Sunstone", "Golden Coin"}
	case EventThemeAutumn:
		return []string{"Harvest Coin", "Amber Token"}
	case EventThemeWinter:
		return []string{"Frost Crystal", "Star Fragment"}
	default:
		return []string{"Event Token", "Festival Coin"}
	}
}

// getThemeItemRewards returns exclusive item names for the event.
func getThemeItemRewards(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"Verdant Crown", "Blossom Staff", "Renewal Amulet", "Springbloom Armor"}
	case EventThemeSummer:
		return []string{"Sunfire Blade", "Radiant Shield", "Solar Circlet", "Phoenix Cloak"}
	case EventThemeAutumn:
		return []string{"Harvest Scythe", "Twilight Bow", "Cornucopia Ring", "Amber Vest"}
	case EventThemeWinter:
		return []string{"Frost Scepter", "Starlight Blade", "Hearthfire Helm", "Snowdrift Boots"}
	default:
		return []string{"Festival Trophy", "Event Memento"}
	}
}

// getThemeTitles returns cosmetic titles for the event.
func getThemeTitles(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"Herald of Spring", "Blossom Keeper", "The Renewed"}
	case EventThemeSummer:
		return []string{"Sun Champion", "Radiant One", "The Illuminated"}
	case EventThemeAutumn:
		return []string{"Harvest Lord", "Twilight Walker", "The Bountiful"}
	case EventThemeWinter:
		return []string{"Frost Warden", "Starlight Guardian", "The Hearthtender"}
	default:
		return []string{"Festival Champion", "Event Veteran"}
	}
}

// getThemeEffects returns visual effect names for the event.
func getThemeEffects(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"Petal Aura", "Verdant Glow", "Butterfly Trail"}
	case EventThemeSummer:
		return []string{"Solar Flare", "Golden Radiance", "Heat Shimmer"}
	case EventThemeAutumn:
		return []string{"Falling Leaves", "Amber Glow", "Twilight Mist"}
	case EventThemeWinter:
		return []string{"Snowfall Aura", "Frost Crystals", "Starlight Sparkle"}
	default:
		return []string{"Festival Glow", "Event Aura"}
	}
}

// GenerateEventAchievements generates achievements for a seasonal event.
func GenerateEventAchievements(event EventInstance, seed int64) []EventAchievementDef {
	rng := rand.New(rand.NewSource(seed))
	achievements := make([]EventAchievementDef, 0, 5)

	// Quest completion achievement
	achievements = append(achievements, EventAchievementDef{
		ID:             event.Definition.ID + "_quest_master",
		EventID:        event.Definition.ID,
		Name:           event.Definition.Name + " Quest Master",
		Description:    "Complete all 3 event quests",
		Requirement:    "complete_quests",
		RequiredAmount: 3,
		Reward: EventReward{
			ID:      event.Definition.ID + "_quest_master_reward",
			EventID: event.Definition.ID,
			Type:    EventRewardCurrency,
			Name:    "Quest Master Bonus",
			Value:   100 + rng.Intn(50),
			Rarity:  "rare",
		},
	})

	// Currency earning achievement
	achievements = append(achievements, EventAchievementDef{
		ID:             event.Definition.ID + "_collector",
		EventID:        event.Definition.ID,
		Name:           event.Definition.Name + " Collector",
		Description:    "Earn 500 event currency",
		Requirement:    "earn_currency",
		RequiredAmount: 500,
		Reward: EventReward{
			ID:      event.Definition.ID + "_collector_reward",
			EventID: event.Definition.ID,
			Type:    EventRewardTitle,
			Name:    "The Collector",
			Value:   1,
			Rarity:  "epic",
		},
	})

	// Participation achievement
	achievements = append(achievements, EventAchievementDef{
		ID:             event.Definition.ID + "_participant",
		EventID:        event.Definition.ID,
		Name:           event.Definition.Name + " Participant",
		Description:    "Participate in the event",
		Requirement:    "participate",
		RequiredAmount: 1,
		Reward: EventReward{
			ID:      event.Definition.ID + "_participant_reward",
			EventID: event.Definition.ID,
			Type:    EventRewardCurrency,
			Name:    "Participation Bonus",
			Value:   25 + rng.Intn(25),
			Rarity:  "common",
		},
	})

	// Boss defeat achievement
	achievements = append(achievements, EventAchievementDef{
		ID:             event.Definition.ID + "_champion",
		EventID:        event.Definition.ID,
		Name:           event.Definition.Name + " Champion",
		Description:    "Defeat the event boss",
		Requirement:    "defeat_boss",
		RequiredAmount: 1,
		Reward: EventReward{
			ID:       event.Definition.ID + "_champion_reward",
			EventID:  event.Definition.ID,
			Type:     EventRewardItem,
			Name:     "Champion's Trophy",
			Value:    1,
			ItemSeed: rng.Int63(),
			Rarity:   "legendary",
		},
	})

	// Exploration achievement
	achievements = append(achievements, EventAchievementDef{
		ID:             event.Definition.ID + "_explorer",
		EventID:        event.Definition.ID,
		Name:           event.Definition.Name + " Explorer",
		Description:    "Discover the hidden event location",
		Requirement:    "explore_location",
		RequiredAmount: 1,
		Reward: EventReward{
			ID:      event.Definition.ID + "_explorer_reward",
			EventID: event.Definition.ID,
			Type:    EventRewardEffect,
			Name:    "Explorer's Aura",
			Value:   1,
			Rarity:  "rare",
		},
	})

	logrus.WithFields(logrus.Fields{
		"component_type":     "event_reward",
		"event_id":           event.Definition.ID,
		"achievements_count": len(achievements),
		"seed":               seed,
	}).Debug("Generated event achievements")

	return achievements
}
