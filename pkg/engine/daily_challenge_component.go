// Package engine provides daily and weekly challenge components for the ECS.
// This file implements the DailyChallengeComponent for tracking challenge progress,
// streaks, and rewards with deterministic daily/weekly rotation.
//
// Phase 98: Daily/Weekly Challenges (V18.0)
package engine

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"
)

// ChallengeType defines whether a challenge is daily or weekly.
type ChallengeType string

const (
	// ChallengeTypeDaily resets every 24 hours.
	ChallengeTypeDaily ChallengeType = "daily"
	// ChallengeTypeWeekly resets every 7 days.
	ChallengeTypeWeekly ChallengeType = "weekly"
)

// ChallengeCategory defines the activity type for challenges.
type ChallengeCategory string

const (
	// ChallengeCategoryCombat for combat-related challenges.
	ChallengeCategoryCombat ChallengeCategory = "combat"
	// ChallengeCategoryGathering for resource gathering challenges.
	ChallengeCategoryGathering ChallengeCategory = "gathering"
	// ChallengeCategoryExploration for exploration challenges.
	ChallengeCategoryExploration ChallengeCategory = "exploration"
	// ChallengeCategorySocial for social interaction challenges.
	ChallengeCategorySocial ChallengeCategory = "social"
	// ChallengeCategoryCrafting for crafting challenges.
	ChallengeCategoryCrafting ChallengeCategory = "crafting"
)

// AllChallengeCategories returns all challenge categories.
func AllChallengeCategories() []ChallengeCategory {
	return []ChallengeCategory{
		ChallengeCategoryCombat,
		ChallengeCategoryGathering,
		ChallengeCategoryExploration,
		ChallengeCategorySocial,
		ChallengeCategoryCrafting,
	}
}

// String returns the string representation of the category.
func (c ChallengeCategory) String() string {
	return string(c)
}

// ChallengeReward defines the reward for completing a challenge.
type ChallengeReward struct {
	// XP is experience points awarded.
	XP int `json:"xp"`
	// Gold is currency awarded.
	Gold int `json:"gold"`
	// ItemID is an optional item reward (empty if none).
	ItemID string `json:"item_id,omitempty"`
	// BonusMultiplier for streak bonuses (1.0 = no bonus).
	BonusMultiplier float64 `json:"bonus_multiplier"`
}

// Challenge represents a single daily or weekly challenge.
type Challenge struct {
	// ID is the unique identifier for this challenge instance.
	ID string `json:"id"`
	// DefinitionID is the challenge type (e.g., "kill_enemies", "gather_herbs").
	DefinitionID string `json:"definition_id"`
	// Type is daily or weekly.
	Type ChallengeType `json:"type"`
	// Category is the activity type.
	Category ChallengeCategory `json:"category"`
	// Name is the display name.
	Name string `json:"name"`
	// Description explains what the player needs to do.
	Description string `json:"description"`
	// Target is the goal amount to complete.
	Target int `json:"target"`
	// Progress is the current progress toward target.
	Progress int `json:"progress"`
	// Reward is what the player earns on completion.
	Reward ChallengeReward `json:"reward"`
	// IsCompleted indicates the challenge is finished.
	IsCompleted bool `json:"is_completed"`
	// IsRewardClaimed indicates the reward has been claimed.
	IsRewardClaimed bool `json:"is_reward_claimed"`
	// ExpiresAt is the unix timestamp when this challenge expires.
	ExpiresAt int64 `json:"expires_at"`
	// CreatedAt is the unix timestamp when this challenge was generated.
	CreatedAt int64 `json:"created_at"`
}

// ChallengeDefinition defines a template for generating challenges.
type ChallengeDefinition struct {
	ID          string
	Category    ChallengeCategory
	Name        string
	Description string
	MinTarget   int
	MaxTarget   int
	BaseXP      int
	BaseGold    int
	TrackingKey string // Key used to track progress (e.g., "enemies_killed")
}

// DefaultDailyChallengeDefinitions returns the standard daily challenge templates.
func DefaultDailyChallengeDefinitions() []ChallengeDefinition {
	return []ChallengeDefinition{
		// Combat
		{ID: "kill_enemies", Category: ChallengeCategoryCombat, Name: "Enemy Slayer", Description: "Defeat %d enemies", MinTarget: 10, MaxTarget: 30, BaseXP: 100, BaseGold: 50, TrackingKey: "enemies_killed"},
		{ID: "deal_damage", Category: ChallengeCategoryCombat, Name: "Damage Dealer", Description: "Deal %d total damage", MinTarget: 500, MaxTarget: 2000, BaseXP: 80, BaseGold: 40, TrackingKey: "damage_dealt"},
		{ID: "crit_hits", Category: ChallengeCategoryCombat, Name: "Critical Striker", Description: "Land %d critical hits", MinTarget: 5, MaxTarget: 15, BaseXP: 120, BaseGold: 60, TrackingKey: "critical_hits"},
		// Gathering
		{ID: "gather_resources", Category: ChallengeCategoryGathering, Name: "Resource Collector", Description: "Gather %d resources", MinTarget: 15, MaxTarget: 40, BaseXP: 90, BaseGold: 45, TrackingKey: "resources_gathered"},
		{ID: "catch_fish", Category: ChallengeCategoryGathering, Name: "Angler", Description: "Catch %d fish", MinTarget: 5, MaxTarget: 15, BaseXP: 110, BaseGold: 55, TrackingKey: "fish_caught"},
		{ID: "gather_herbs", Category: ChallengeCategoryGathering, Name: "Herbalist", Description: "Gather %d herbs", MinTarget: 10, MaxTarget: 25, BaseXP: 85, BaseGold: 42, TrackingKey: "herbs_gathered"},
		// Exploration
		{ID: "discover_areas", Category: ChallengeCategoryExploration, Name: "Explorer", Description: "Discover %d new areas", MinTarget: 2, MaxTarget: 5, BaseXP: 150, BaseGold: 75, TrackingKey: "areas_discovered"},
		{ID: "travel_distance", Category: ChallengeCategoryExploration, Name: "Wanderer", Description: "Travel %d units", MinTarget: 1000, MaxTarget: 3000, BaseXP: 70, BaseGold: 35, TrackingKey: "distance_traveled"},
		{ID: "find_secrets", Category: ChallengeCategoryExploration, Name: "Secret Seeker", Description: "Find %d hidden secrets", MinTarget: 1, MaxTarget: 3, BaseXP: 200, BaseGold: 100, TrackingKey: "secrets_found"},
		// Social
		{ID: "complete_trades", Category: ChallengeCategorySocial, Name: "Trader", Description: "Complete %d trades", MinTarget: 2, MaxTarget: 5, BaseXP: 100, BaseGold: 50, TrackingKey: "trades_completed"},
		{ID: "send_messages", Category: ChallengeCategorySocial, Name: "Socialite", Description: "Send %d chat messages", MinTarget: 10, MaxTarget: 30, BaseXP: 50, BaseGold: 25, TrackingKey: "messages_sent"},
		{ID: "help_players", Category: ChallengeCategorySocial, Name: "Helper", Description: "Assist %d players", MinTarget: 1, MaxTarget: 3, BaseXP: 150, BaseGold: 75, TrackingKey: "players_helped"},
		// Crafting
		{ID: "craft_items", Category: ChallengeCategoryCrafting, Name: "Artisan", Description: "Craft %d items", MinTarget: 3, MaxTarget: 10, BaseXP: 110, BaseGold: 55, TrackingKey: "items_crafted"},
		{ID: "craft_quality", Category: ChallengeCategoryCrafting, Name: "Quality Crafter", Description: "Craft %d uncommon+ items", MinTarget: 1, MaxTarget: 3, BaseXP: 180, BaseGold: 90, TrackingKey: "quality_items_crafted"},
		{ID: "use_recipes", Category: ChallengeCategoryCrafting, Name: "Recipe Master", Description: "Use %d different recipes", MinTarget: 2, MaxTarget: 5, BaseXP: 130, BaseGold: 65, TrackingKey: "recipes_used"},
	}
}

// DefaultWeeklyChallengeDefinitions returns the standard weekly challenge templates.
func DefaultWeeklyChallengeDefinitions() []ChallengeDefinition {
	return []ChallengeDefinition{
		// Combat
		{ID: "weekly_boss_kills", Category: ChallengeCategoryCombat, Name: "Boss Hunter", Description: "Defeat %d bosses", MinTarget: 3, MaxTarget: 7, BaseXP: 500, BaseGold: 250, TrackingKey: "bosses_killed"},
		{ID: "weekly_dungeons", Category: ChallengeCategoryCombat, Name: "Dungeon Delver", Description: "Complete %d dungeons", MinTarget: 3, MaxTarget: 5, BaseXP: 600, BaseGold: 300, TrackingKey: "dungeons_completed"},
		// Gathering
		{ID: "weekly_rare_resources", Category: ChallengeCategoryGathering, Name: "Rare Collector", Description: "Gather %d rare resources", MinTarget: 10, MaxTarget: 25, BaseXP: 400, BaseGold: 200, TrackingKey: "rare_resources_gathered"},
		{ID: "weekly_fishing_variety", Category: ChallengeCategoryGathering, Name: "Fish Collector", Description: "Catch %d different fish types", MinTarget: 5, MaxTarget: 10, BaseXP: 450, BaseGold: 225, TrackingKey: "fish_types_caught"},
		// Exploration
		{ID: "weekly_world_tour", Category: ChallengeCategoryExploration, Name: "World Traveler", Description: "Visit %d different biomes", MinTarget: 3, MaxTarget: 5, BaseXP: 550, BaseGold: 275, TrackingKey: "biomes_visited"},
		// Social
		{ID: "weekly_guild_activity", Category: ChallengeCategorySocial, Name: "Guild Champion", Description: "Complete %d guild activities", MinTarget: 5, MaxTarget: 10, BaseXP: 500, BaseGold: 250, TrackingKey: "guild_activities_completed"},
		// Crafting
		{ID: "weekly_masterwork", Category: ChallengeCategoryCrafting, Name: "Master Craftsman", Description: "Craft %d rare+ items", MinTarget: 3, MaxTarget: 7, BaseXP: 600, BaseGold: 300, TrackingKey: "rare_items_crafted"},
	}
}

// DailyChallengeComponent tracks active challenges, completion, and streaks.
type DailyChallengeComponent struct {
	mu sync.RWMutex

	// ActiveDailyChallenges contains the current daily challenges (max 5).
	ActiveDailyChallenges []*Challenge `json:"active_daily_challenges"`

	// ActiveWeeklyChallenges contains the current weekly challenges (max 3).
	ActiveWeeklyChallenges []*Challenge `json:"active_weekly_challenges"`

	// CompletedTodayIDs tracks challenge IDs completed in current daily cycle.
	CompletedTodayIDs []string `json:"completed_today_ids"`

	// DailyStreak is consecutive days with all daily challenges completed.
	DailyStreak int `json:"daily_streak"`

	// LongestDailyStreak is the all-time highest streak.
	LongestDailyStreak int `json:"longest_daily_streak"`

	// WeeklyStreak is consecutive weeks with all weekly challenges completed.
	WeeklyStreak int `json:"weekly_streak"`

	// LongestWeeklyStreak is the all-time highest weekly streak.
	LongestWeeklyStreak int `json:"longest_weekly_streak"`

	// TotalChallengesCompleted is lifetime completed challenges.
	TotalChallengesCompleted int `json:"total_challenges_completed"`

	// TotalXPEarned is lifetime XP from challenges.
	TotalXPEarned int `json:"total_xp_earned"`

	// TotalGoldEarned is lifetime gold from challenges.
	TotalGoldEarned int `json:"total_gold_earned"`

	// LastDailyReset is unix timestamp of last daily reset.
	LastDailyReset int64 `json:"last_daily_reset"`

	// LastWeeklyReset is unix timestamp of last weekly reset.
	LastWeeklyReset int64 `json:"last_weekly_reset"`

	// RerollsRemaining is how many times challenges can be rerolled today.
	RerollsRemaining int `json:"rerolls_remaining"`

	// MaxDailyRerolls is the maximum rerolls per day.
	MaxDailyRerolls int `json:"max_daily_rerolls"`

	// BaseSeed is the base seed for deterministic generation.
	BaseSeed int64 `json:"base_seed"`
}

// NewDailyChallengeComponent creates a new challenge component with defaults.
func NewDailyChallengeComponent(baseSeed int64) *DailyChallengeComponent {
	return &DailyChallengeComponent{
		ActiveDailyChallenges:    make([]*Challenge, 0, 5),
		ActiveWeeklyChallenges:   make([]*Challenge, 0, 3),
		CompletedTodayIDs:        make([]string, 0),
		DailyStreak:              0,
		LongestDailyStreak:       0,
		WeeklyStreak:             0,
		LongestWeeklyStreak:      0,
		TotalChallengesCompleted: 0,
		TotalXPEarned:            0,
		TotalGoldEarned:          0,
		LastDailyReset:           0,
		LastWeeklyReset:          0,
		RerollsRemaining:         3,
		MaxDailyRerolls:          3,
		BaseSeed:                 baseSeed,
	}
}

// Type returns the component type identifier.
func (c *DailyChallengeComponent) Type() string {
	return "daily_challenge"
}

// createDateSeed generates a deterministic seed from a date
func createDateSeed(baseSeed int64, date time.Time) int64 {
	dateSeed := int64(date.Year())*10000 + int64(date.YearDay())
	return baseSeed + dateSeed
}

// shuffleChallengeDefinitions performs Fisher-Yates shuffle on challenge definitions
func shuffleChallengeDefinitions(definitions []ChallengeDefinition, rng *rand.Rand) []ChallengeDefinition {
	shuffled := make([]ChallengeDefinition, len(definitions))
	copy(shuffled, definitions)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// selectChallengesByCategory selects one challenge from each category until limit reached
func (c *DailyChallengeComponent) selectChallengesByCategory(shuffled []ChallengeDefinition, date time.Time, rng *rand.Rand, limit int) ([]*Challenge, map[ChallengeCategory]bool) {
	challenges := make([]*Challenge, 0, limit)
	usedCategories := make(map[ChallengeCategory]bool)

	for _, def := range shuffled {
		if len(challenges) >= limit {
			break
		}
		if usedCategories[def.Category] {
			continue
		}
		challenge := c.generateChallengeFromDefinition(def, ChallengeTypeDaily, date, rng)
		challenges = append(challenges, challenge)
		usedCategories[def.Category] = true
	}

	return challenges, usedCategories
}

// fillRemainingChallenges fills remaining challenge slots with unused definitions
func (c *DailyChallengeComponent) fillRemainingChallenges(shuffled []ChallengeDefinition, challenges []*Challenge, date time.Time, rng *rand.Rand, limit int) []*Challenge {
	for _, def := range shuffled {
		if len(challenges) >= limit {
			break
		}
		alreadyUsed := false
		for _, ch := range challenges {
			if ch.DefinitionID == def.ID {
				alreadyUsed = true
				break
			}
		}
		if !alreadyUsed {
			challenge := c.generateChallengeFromDefinition(def, ChallengeTypeDaily, date, rng)
			challenges = append(challenges, challenge)
		}
	}
	return challenges
}

// GenerateDailyChallenges generates 5 daily challenges for the given date.
// Uses deterministic generation: same date + baseSeed = same challenges.
func (c *DailyChallengeComponent) GenerateDailyChallenges(date time.Time) []*Challenge {
	c.mu.Lock()
	defer c.mu.Unlock()

	seed := createDateSeed(c.BaseSeed, date)
	rng := rand.New(rand.NewSource(seed))

	definitions := DefaultDailyChallengeDefinitions()
	shuffled := shuffleChallengeDefinitions(definitions, rng)

	challenges, _ := c.selectChallengesByCategory(shuffled, date, rng, 5)
	challenges = c.fillRemainingChallenges(shuffled, challenges, date, rng, 5)

	return challenges
}

// GenerateWeeklyChallenges generates 3 weekly challenges for the given week.
func (c *DailyChallengeComponent) GenerateWeeklyChallenges(weekStart time.Time) []*Challenge {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create week-based seed for determinism
	_, week := weekStart.ISOWeek()
	weekSeed := int64(weekStart.Year())*100 + int64(week)
	rng := rand.New(rand.NewSource(c.BaseSeed + weekSeed*1000))

	definitions := DefaultWeeklyChallengeDefinitions()
	shuffled := make([]ChallengeDefinition, len(definitions))
	copy(shuffled, definitions)

	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Generate 3 weekly challenges
	challenges := make([]*Challenge, 0, 3)
	for i := 0; i < 3 && i < len(shuffled); i++ {
		challenge := c.generateChallengeFromDefinition(shuffled[i], ChallengeTypeWeekly, weekStart, rng)
		challenges = append(challenges, challenge)
	}

	return challenges
}

// generateChallengeFromDefinition creates a challenge from a definition.
func (c *DailyChallengeComponent) generateChallengeFromDefinition(def ChallengeDefinition, challengeType ChallengeType, baseTime time.Time, rng *rand.Rand) *Challenge {
	// Random target within range
	target := def.MinTarget
	if def.MaxTarget > def.MinTarget {
		target = def.MinTarget + rng.Intn(def.MaxTarget-def.MinTarget+1)
	}

	// Calculate expiry
	var expiresAt int64
	if challengeType == ChallengeTypeDaily {
		// Expires at midnight next day (UTC)
		nextDay := baseTime.AddDate(0, 0, 1)
		expiresAt = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, time.UTC).Unix()
	} else {
		// Expires at start of next week
		daysUntilMonday := (8 - int(baseTime.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextWeek := baseTime.AddDate(0, 0, daysUntilMonday)
		expiresAt = time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 0, 0, 0, 0, time.UTC).Unix()
	}

	// Scale rewards for weeklies
	xp := def.BaseXP
	gold := def.BaseGold
	if challengeType == ChallengeTypeWeekly {
		xp = int(float64(xp) * 1.5)
		gold = int(float64(gold) * 1.5)
	}

	// Create unique ID based on definition + time
	id := def.ID + "_" + baseTime.Format("20060102")
	if challengeType == ChallengeTypeWeekly {
		_, week := baseTime.ISOWeek()
		id = def.ID + "_w" + baseTime.Format("2006") + "_" + string(rune('0'+week/10)) + string(rune('0'+week%10))
	}

	return &Challenge{
		ID:              id,
		DefinitionID:    def.ID,
		Type:            challengeType,
		Category:        def.Category,
		Name:            def.Name,
		Description:     def.Description,
		Target:          target,
		Progress:        0,
		Reward:          ChallengeReward{XP: xp, Gold: gold, BonusMultiplier: 1.0},
		IsCompleted:     false,
		IsRewardClaimed: false,
		ExpiresAt:       expiresAt,
		CreatedAt:       baseTime.Unix(),
	}
}

// SetActiveChallenges sets the active daily and weekly challenges.
func (c *DailyChallengeComponent) SetActiveChallenges(daily, weekly []*Challenge) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ActiveDailyChallenges = daily
	c.ActiveWeeklyChallenges = weekly
}

// GetActiveDailyChallenges returns the current daily challenges.
func (c *DailyChallengeComponent) GetActiveDailyChallenges() []*Challenge {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*Challenge, len(c.ActiveDailyChallenges))
	for i, ch := range c.ActiveDailyChallenges {
		chCopy := *ch
		result[i] = &chCopy
	}
	return result
}

// GetActiveWeeklyChallenges returns the current weekly challenges.
func (c *DailyChallengeComponent) GetActiveWeeklyChallenges() []*Challenge {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*Challenge, len(c.ActiveWeeklyChallenges))
	for i, ch := range c.ActiveWeeklyChallenges {
		chCopy := *ch
		result[i] = &chCopy
	}
	return result
}

// UpdateProgress updates progress for a challenge matching the tracking key.
// Returns the challenge if it was updated, nil otherwise.
func (c *DailyChallengeComponent) UpdateProgress(trackingKey string, amount int) *Challenge {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check daily challenges
	for _, ch := range c.ActiveDailyChallenges {
		def := c.findDefinitionByID(ch.DefinitionID)
		if def != nil && def.TrackingKey == trackingKey && !ch.IsCompleted {
			ch.Progress += amount
			if ch.Progress >= ch.Target {
				ch.Progress = ch.Target
				ch.IsCompleted = true
				c.CompletedTodayIDs = append(c.CompletedTodayIDs, ch.ID)
			}
			chCopy := *ch
			return &chCopy
		}
	}

	// Check weekly challenges
	for _, ch := range c.ActiveWeeklyChallenges {
		def := c.findDefinitionByID(ch.DefinitionID)
		if def != nil && def.TrackingKey == trackingKey && !ch.IsCompleted {
			ch.Progress += amount
			if ch.Progress >= ch.Target {
				ch.Progress = ch.Target
				ch.IsCompleted = true
			}
			chCopy := *ch
			return &chCopy
		}
	}

	return nil
}

// findDefinitionByID finds a challenge definition by ID.
func (c *DailyChallengeComponent) findDefinitionByID(id string) *ChallengeDefinition {
	for _, def := range DefaultDailyChallengeDefinitions() {
		if def.ID == id {
			return &def
		}
	}
	for _, def := range DefaultWeeklyChallengeDefinitions() {
		if def.ID == id {
			return &def
		}
	}
	return nil
}

// ClaimReward claims the reward for a completed challenge.
// Returns the reward with streak bonuses applied, or nil if cannot claim.
func (c *DailyChallengeComponent) ClaimReward(challengeID string) *ChallengeReward {
	c.mu.Lock()
	defer c.mu.Unlock()

	challenge := c.findChallengeByID(challengeID)
	if !c.canClaimReward(challenge) {
		return nil
	}

	streakBonus := c.calculateStreakBonus(challenge)
	reward := c.buildReward(challenge, streakBonus)

	c.markRewardClaimed(challenge, &reward)

	return &reward
}

// findChallengeByID searches for a challenge by ID in both daily and weekly lists.
func (c *DailyChallengeComponent) findChallengeByID(challengeID string) *Challenge {
	for _, ch := range c.ActiveDailyChallenges {
		if ch.ID == challengeID {
			return ch
		}
	}

	for _, ch := range c.ActiveWeeklyChallenges {
		if ch.ID == challengeID {
			return ch
		}
	}

	return nil
}

// canClaimReward checks if a challenge reward can be claimed.
func (c *DailyChallengeComponent) canClaimReward(challenge *Challenge) bool {
	return challenge != nil && challenge.IsCompleted && !challenge.IsRewardClaimed
}

// calculateStreakBonus determines the bonus multiplier based on streak.
func (c *DailyChallengeComponent) calculateStreakBonus(challenge *Challenge) float64 {
	if challenge.Type == ChallengeTypeDaily {
		return c.calculateDailyStreakBonus()
	}
	if challenge.Type == ChallengeTypeWeekly {
		return c.calculateWeeklyStreakBonus()
	}
	return 1.0
}

// calculateDailyStreakBonus computes the daily streak bonus (10% per day, max 100%).
func (c *DailyChallengeComponent) calculateDailyStreakBonus() float64 {
	if c.DailyStreak <= 0 {
		return 1.0
	}

	bonus := float64(c.DailyStreak) * 0.1
	if bonus > 1.0 {
		bonus = 1.0
	}
	return 1.0 + bonus
}

// calculateWeeklyStreakBonus computes the weekly streak bonus (25% per week, max 100%).
func (c *DailyChallengeComponent) calculateWeeklyStreakBonus() float64 {
	if c.WeeklyStreak <= 0 {
		return 1.0
	}

	bonus := float64(c.WeeklyStreak) * 0.25
	if bonus > 1.0 {
		bonus = 1.0
	}
	return 1.0 + bonus
}

// buildReward constructs the final reward with bonuses applied.
func (c *DailyChallengeComponent) buildReward(challenge *Challenge, streakBonus float64) ChallengeReward {
	return ChallengeReward{
		XP:              int(float64(challenge.Reward.XP) * streakBonus),
		Gold:            int(float64(challenge.Reward.Gold) * streakBonus),
		ItemID:          challenge.Reward.ItemID,
		BonusMultiplier: streakBonus,
	}
}

// markRewardClaimed updates the challenge and component statistics after claiming.
func (c *DailyChallengeComponent) markRewardClaimed(challenge *Challenge, reward *ChallengeReward) {
	challenge.IsRewardClaimed = true
	c.TotalChallengesCompleted++
	c.TotalXPEarned += reward.XP
	c.TotalGoldEarned += reward.Gold
}

// RerollChallenge rerolls a daily challenge if rerolls are available.
// Returns the new challenge or nil if reroll failed.
func (c *DailyChallengeComponent) RerollChallenge(challengeID string, currentTime time.Time) *Challenge {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.RerollsRemaining <= 0 {
		return nil
	}

	// Find and remove the challenge
	idx := -1
	for i, ch := range c.ActiveDailyChallenges {
		if ch.ID == challengeID && !ch.IsCompleted {
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil
	}

	// Get definitions not currently in use
	usedDefs := make(map[string]bool)
	for _, ch := range c.ActiveDailyChallenges {
		usedDefs[ch.DefinitionID] = true
	}

	definitions := DefaultDailyChallengeDefinitions()
	available := make([]ChallengeDefinition, 0)
	for _, def := range definitions {
		if !usedDefs[def.ID] {
			available = append(available, def)
		}
	}

	if len(available) == 0 {
		return nil
	}

	// Generate new challenge
	rerollSeed := c.BaseSeed + currentTime.UnixNano()
	rng := rand.New(rand.NewSource(rerollSeed))
	newDef := available[rng.Intn(len(available))]
	newChallenge := c.generateChallengeFromDefinition(newDef, ChallengeTypeDaily, currentTime, rng)

	c.ActiveDailyChallenges[idx] = newChallenge
	c.RerollsRemaining--

	chCopy := *newChallenge
	return &chCopy
}

// CheckDailyReset checks if daily challenges should reset.
// Returns true if reset occurred.
func (c *DailyChallengeComponent) CheckDailyReset(currentTime time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC).Unix()
	lastResetDay := time.Unix(c.LastDailyReset, 0).UTC()
	lastResetDayStart := time.Date(lastResetDay.Year(), lastResetDay.Month(), lastResetDay.Day(), 0, 0, 0, 0, time.UTC).Unix()

	if currentDay > lastResetDayStart {
		// Check streak before reset
		allCompleted := true
		for _, ch := range c.ActiveDailyChallenges {
			if !ch.IsCompleted {
				allCompleted = false
				break
			}
		}

		if allCompleted && len(c.ActiveDailyChallenges) > 0 {
			c.DailyStreak++
			if c.DailyStreak > c.LongestDailyStreak {
				c.LongestDailyStreak = c.DailyStreak
			}
		} else {
			c.DailyStreak = 0
		}

		c.ActiveDailyChallenges = make([]*Challenge, 0, 5)
		c.CompletedTodayIDs = make([]string, 0)
		c.RerollsRemaining = c.MaxDailyRerolls
		c.LastDailyReset = currentTime.Unix()
		return true
	}

	return false
}

// CheckWeeklyReset checks if weekly challenges should reset.
// Returns true if reset occurred.
func (c *DailyChallengeComponent) CheckWeeklyReset(currentTime time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get start of current week (Monday)
	daysFromMonday := (int(currentTime.Weekday()) + 6) % 7
	currentWeekStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()-daysFromMonday, 0, 0, 0, 0, time.UTC).Unix()

	lastResetTime := time.Unix(c.LastWeeklyReset, 0).UTC()
	lastResetDaysFromMonday := (int(lastResetTime.Weekday()) + 6) % 7
	lastWeekStart := time.Date(lastResetTime.Year(), lastResetTime.Month(), lastResetTime.Day()-lastResetDaysFromMonday, 0, 0, 0, 0, time.UTC).Unix()

	if currentWeekStart > lastWeekStart {
		// Check streak before reset
		allCompleted := true
		for _, ch := range c.ActiveWeeklyChallenges {
			if !ch.IsCompleted {
				allCompleted = false
				break
			}
		}

		if allCompleted && len(c.ActiveWeeklyChallenges) > 0 {
			c.WeeklyStreak++
			if c.WeeklyStreak > c.LongestWeeklyStreak {
				c.LongestWeeklyStreak = c.WeeklyStreak
			}
		} else {
			c.WeeklyStreak = 0
		}

		c.ActiveWeeklyChallenges = make([]*Challenge, 0, 3)
		c.LastWeeklyReset = currentTime.Unix()
		return true
	}

	return false
}

// GetDailyStreak returns the current daily streak.
func (c *DailyChallengeComponent) GetDailyStreak() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DailyStreak
}

// GetWeeklyStreak returns the current weekly streak.
func (c *DailyChallengeComponent) GetWeeklyStreak() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.WeeklyStreak
}

// GetTotalCompleted returns total challenges completed.
func (c *DailyChallengeComponent) GetTotalCompleted() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TotalChallengesCompleted
}

// GetRerollsRemaining returns remaining rerolls for today.
func (c *DailyChallengeComponent) GetRerollsRemaining() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RerollsRemaining
}

// GetDailyCompletionPercent returns completion percentage for today's daily challenges.
func (c *DailyChallengeComponent) GetDailyCompletionPercent() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ActiveDailyChallenges) == 0 {
		return 0
	}

	completed := 0
	for _, ch := range c.ActiveDailyChallenges {
		if ch.IsCompleted {
			completed++
		}
	}
	return float64(completed) / float64(len(c.ActiveDailyChallenges)) * 100
}

// GetWeeklyCompletionPercent returns completion percentage for weekly challenges.
func (c *DailyChallengeComponent) GetWeeklyCompletionPercent() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ActiveWeeklyChallenges) == 0 {
		return 0
	}

	completed := 0
	for _, ch := range c.ActiveWeeklyChallenges {
		if ch.IsCompleted {
			completed++
		}
	}
	return float64(completed) / float64(len(c.ActiveWeeklyChallenges)) * 100
}

// IsChallengeExpired checks if a challenge has expired.
func (c *DailyChallengeComponent) IsChallengeExpired(challengeID string, currentTime time.Time) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, ch := range c.ActiveDailyChallenges {
		if ch.ID == challengeID {
			return currentTime.Unix() > ch.ExpiresAt
		}
	}
	for _, ch := range c.ActiveWeeklyChallenges {
		if ch.ID == challengeID {
			return currentTime.Unix() > ch.ExpiresAt
		}
	}
	return true
}

// GetStreakBonus returns the current streak bonus multiplier (1.0 = no bonus).
func (c *DailyChallengeComponent) GetStreakBonus(challengeType ChallengeType) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if challengeType == ChallengeTypeDaily && c.DailyStreak > 0 {
		bonus := float64(c.DailyStreak) * 0.1
		if bonus > 1.0 {
			bonus = 1.0
		}
		return 1.0 + bonus
	} else if challengeType == ChallengeTypeWeekly && c.WeeklyStreak > 0 {
		bonus := float64(c.WeeklyStreak) * 0.25
		if bonus > 1.0 {
			bonus = 1.0
		}
		return 1.0 + bonus
	}
	return 1.0
}

// Serialize converts the component to JSON bytes.
func (c *DailyChallengeComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *DailyChallengeComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
