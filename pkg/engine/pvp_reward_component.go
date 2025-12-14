// Package engine provides the PvP reward component for competitive PvP rewards.
// PvPRewardComponent tracks Honor Points, seasonal rewards, tournament wins,
// and PvP-specific achievements for players.
package engine

import (
	"encoding/json"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// PvPRewardType categorizes the kind of PvP reward.
type PvPRewardType string

const (
	// PvPRewardHonor represents Honor Points currency.
	PvPRewardHonor PvPRewardType = "honor"
	// PvPRewardItem represents a PvP-exclusive item.
	PvPRewardItem PvPRewardType = "item"
	// PvPRewardTitle represents a PvP cosmetic title.
	PvPRewardTitle PvPRewardType = "title"
	// PvPRewardMount represents a PvP mount reward.
	PvPRewardMount PvPRewardType = "mount"
	// PvPRewardCosmetic represents a visual cosmetic (aura, effect).
	PvPRewardCosmetic PvPRewardType = "cosmetic"
	// PvPRewardAchievement represents a PvP achievement unlock.
	PvPRewardAchievement PvPRewardType = "achievement"
)

// PvPReward represents a reward earned through PvP activities.
type PvPReward struct {
	// ID is the unique identifier for this reward
	ID string `json:"id"`
	// SeasonID links this reward to a specific season (empty for permanent)
	SeasonID string `json:"season_id,omitempty"`
	// Type categorizes the reward
	Type PvPRewardType `json:"type"`
	// Name is the display name
	Name string `json:"name"`
	// Description explains the reward
	Description string `json:"description"`
	// Value is the numerical value (honor amount, item quantity)
	Value int `json:"value"`
	// ItemSeed is the seed for generating exclusive items
	ItemSeed int64 `json:"item_seed,omitempty"`
	// Rarity indicates how rare the reward is
	Rarity string `json:"rarity"`
	// RankRequirement is the minimum rank to earn this reward
	RankRequirement RankTier `json:"rank_requirement,omitempty"`
	// Claimed indicates if the reward has been claimed
	Claimed bool `json:"claimed"`
}

// PvPAchievementDef defines a PvP achievement.
type PvPAchievementDef struct {
	// ID is the unique identifier
	ID string `json:"id"`
	// Name is the achievement name
	Name string `json:"name"`
	// Description explains the achievement
	Description string `json:"description"`
	// Requirement is what must be done (e.g., "wins", "rating", "streak")
	Requirement string `json:"requirement"`
	// RequiredAmount is the target number
	RequiredAmount int `json:"required_amount"`
	// Reward is granted when achievement is completed
	Reward PvPReward `json:"reward"`
}

// PvPAchievementProgress tracks progress toward a PvP achievement.
type PvPAchievementProgress struct {
	// AchievementID is the achievement being tracked
	AchievementID string `json:"achievement_id"`
	// CurrentProgress is the current value
	CurrentProgress int `json:"current_progress"`
	// RequiredAmount is the target
	RequiredAmount int `json:"required_amount"`
	// Completed indicates if requirement is met
	Completed bool `json:"completed"`
}

// SeasonRewardTier defines rewards for reaching a rank tier in a season.
type SeasonRewardTier struct {
	// Tier is the rank tier this reward is for
	Tier RankTier `json:"tier"`
	// SeasonID is the season this reward belongs to
	SeasonID string `json:"season_id"`
	// Rewards are the items granted for reaching this tier
	Rewards []PvPReward `json:"rewards"`
	// Claimed indicates if these rewards have been claimed
	Claimed bool `json:"claimed"`
}

// PvPRewardComponent tracks PvP rewards for a player entity.
type PvPRewardComponent struct {
	// HonorPoints is the PvP currency balance
	HonorPoints int `json:"honor_points"`
	// TotalHonorEarned is lifetime honor earned
	TotalHonorEarned int `json:"total_honor_earned"`
	// EarnedRewards contains all PvP rewards earned
	EarnedRewards []PvPReward `json:"earned_rewards"`
	// SeasonalRewards tracks tier rewards per season
	SeasonalRewards []SeasonRewardTier `json:"seasonal_rewards"`
	// TournamentWins is total tournament victories
	TournamentWins int `json:"tournament_wins"`
	// TournamentParticipations is total tournaments entered
	TournamentParticipations int `json:"tournament_participations"`
	// AchievementProgress tracks progress toward PvP achievements
	AchievementProgress []PvPAchievementProgress `json:"achievement_progress"`
	// CompletedAchievements contains finished achievement IDs
	CompletedAchievements []string `json:"completed_achievements"`
	// EarnedTitles contains unlocked PvP titles
	EarnedTitles []string `json:"earned_titles"`
	// ActiveTitle is the currently displayed PvP title
	ActiveTitle string `json:"active_title"`
	// EarnedMounts contains unlocked PvP mounts
	EarnedMounts []string `json:"earned_mounts"`
	// ActiveMount is the currently equipped PvP mount
	ActiveMount string `json:"active_mount"`
	// EarnedCosmetics contains unlocked visual cosmetics
	EarnedCosmetics []string `json:"earned_cosmetics"`
	// ActiveCosmetic is the currently displayed cosmetic
	ActiveCosmetic string `json:"active_cosmetic"`
	// HighestSeasonRank tracks peak rank per season
	HighestSeasonRank map[string]RankTier `json:"highest_season_rank"`
	// CurrentSeasonID is the active season
	CurrentSeasonID string `json:"current_season_id"`
}

// NewPvPRewardComponent creates a new PvP reward component.
func NewPvPRewardComponent(seasonID string) *PvPRewardComponent {
	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"season_id":      seasonID,
	}).Debug("Creating PvP reward component")

	return &PvPRewardComponent{
		HonorPoints:           0,
		TotalHonorEarned:      0,
		EarnedRewards:         make([]PvPReward, 0),
		SeasonalRewards:       make([]SeasonRewardTier, 0),
		AchievementProgress:   make([]PvPAchievementProgress, 0),
		CompletedAchievements: make([]string, 0),
		EarnedTitles:          make([]string, 0),
		EarnedMounts:          make([]string, 0),
		EarnedCosmetics:       make([]string, 0),
		HighestSeasonRank:     make(map[string]RankTier),
		CurrentSeasonID:       seasonID,
	}
}

// Type returns the component type identifier.
func (c *PvPRewardComponent) Type() string {
	return "pvp_reward"
}

// AddHonor adds Honor Points to the player's balance.
func (c *PvPRewardComponent) AddHonor(amount int) {
	if amount <= 0 {
		return
	}

	c.HonorPoints += amount
	c.TotalHonorEarned += amount

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"amount":         amount,
		"new_total":      c.HonorPoints,
	}).Debug("Added Honor Points")
}

// SpendHonor spends Honor Points.
// Returns true if successful, false if insufficient funds.
func (c *PvPRewardComponent) SpendHonor(amount int) bool {
	if amount <= 0 {
		return false
	}

	if c.HonorPoints < amount {
		logrus.WithFields(logrus.Fields{
			"component_type": "pvp_reward",
			"requested":      amount,
			"available":      c.HonorPoints,
		}).Debug("Insufficient Honor Points")
		return false
	}

	c.HonorPoints -= amount

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"spent":          amount,
		"remaining":      c.HonorPoints,
	}).Debug("Spent Honor Points")

	return true
}

// GetHonor returns the current Honor Points balance.
func (c *PvPRewardComponent) GetHonor() int {
	return c.HonorPoints
}

// AddReward adds an earned PvP reward.
func (c *PvPRewardComponent) AddReward(reward PvPReward) {
	c.EarnedRewards = append(c.EarnedRewards, reward)

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"reward_id":      reward.ID,
		"reward_type":    reward.Type,
		"reward_name":    reward.Name,
	}).Info("Earned PvP reward")
}

// ClaimReward marks a reward as claimed.
func (c *PvPRewardComponent) ClaimReward(rewardID string) bool {
	for i := range c.EarnedRewards {
		if c.EarnedRewards[i].ID == rewardID && !c.EarnedRewards[i].Claimed {
			c.EarnedRewards[i].Claimed = true

			logrus.WithFields(logrus.Fields{
				"component_type": "pvp_reward",
				"reward_id":      rewardID,
			}).Info("Claimed PvP reward")
			return true
		}
	}
	return false
}

// GetUnclaimedRewards returns all unclaimed rewards.
func (c *PvPRewardComponent) GetUnclaimedRewards() []PvPReward {
	unclaimed := make([]PvPReward, 0)
	for _, r := range c.EarnedRewards {
		if !r.Claimed {
			unclaimed = append(unclaimed, r)
		}
	}
	return unclaimed
}

// GetRewardsForSeason returns all rewards for a specific season.
func (c *PvPRewardComponent) GetRewardsForSeason(seasonID string) []PvPReward {
	result := make([]PvPReward, 0)
	for _, r := range c.EarnedRewards {
		if r.SeasonID == seasonID {
			result = append(result, r)
		}
	}
	return result
}

// RecordTournamentWin records a tournament victory.
func (c *PvPRewardComponent) RecordTournamentWin() {
	c.TournamentWins++
	c.TournamentParticipations++

	logrus.WithFields(logrus.Fields{
		"component_type":    "pvp_reward",
		"tournament_wins":   c.TournamentWins,
		"total_tournaments": c.TournamentParticipations,
	}).Debug("Recorded tournament win")
}

// RecordTournamentParticipation records tournament participation without winning.
func (c *PvPRewardComponent) RecordTournamentParticipation() {
	c.TournamentParticipations++
}

// AddTitle unlocks a PvP title.
func (c *PvPRewardComponent) AddTitle(titleID string) bool {
	// Check if already has this title
	for _, t := range c.EarnedTitles {
		if t == titleID {
			return false
		}
	}

	c.EarnedTitles = append(c.EarnedTitles, titleID)

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"title_id":       titleID,
	}).Info("Unlocked PvP title")

	return true
}

// SetActiveTitle sets the displayed PvP title.
func (c *PvPRewardComponent) SetActiveTitle(titleID string) bool {
	for _, t := range c.EarnedTitles {
		if t == titleID {
			c.ActiveTitle = titleID
			return true
		}
	}
	return false
}

// HasTitle checks if the player has a specific title.
func (c *PvPRewardComponent) HasTitle(titleID string) bool {
	for _, t := range c.EarnedTitles {
		if t == titleID {
			return true
		}
	}
	return false
}

// AddMount unlocks a PvP mount.
func (c *PvPRewardComponent) AddMount(mountID string) bool {
	// Check if already has this mount
	for _, m := range c.EarnedMounts {
		if m == mountID {
			return false
		}
	}

	c.EarnedMounts = append(c.EarnedMounts, mountID)

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"mount_id":       mountID,
	}).Info("Unlocked PvP mount")

	return true
}

// SetActiveMount sets the equipped PvP mount.
func (c *PvPRewardComponent) SetActiveMount(mountID string) bool {
	for _, m := range c.EarnedMounts {
		if m == mountID {
			c.ActiveMount = mountID
			return true
		}
	}
	return false
}

// HasMount checks if the player has a specific mount.
func (c *PvPRewardComponent) HasMount(mountID string) bool {
	for _, m := range c.EarnedMounts {
		if m == mountID {
			return true
		}
	}
	return false
}

// AddCosmetic unlocks a visual cosmetic.
func (c *PvPRewardComponent) AddCosmetic(cosmeticID string) bool {
	// Check if already has this cosmetic
	for _, cs := range c.EarnedCosmetics {
		if cs == cosmeticID {
			return false
		}
	}

	c.EarnedCosmetics = append(c.EarnedCosmetics, cosmeticID)

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"cosmetic_id":    cosmeticID,
	}).Info("Unlocked PvP cosmetic")

	return true
}

// SetActiveCosmetic sets the displayed cosmetic effect.
func (c *PvPRewardComponent) SetActiveCosmetic(cosmeticID string) bool {
	for _, cs := range c.EarnedCosmetics {
		if cs == cosmeticID {
			c.ActiveCosmetic = cosmeticID
			return true
		}
	}
	return false
}

// HasCosmetic checks if the player has a specific cosmetic.
func (c *PvPRewardComponent) HasCosmetic(cosmeticID string) bool {
	for _, cs := range c.EarnedCosmetics {
		if cs == cosmeticID {
			return true
		}
	}
	return false
}

// UpdateHighestRank updates the highest rank achieved in the current season.
func (c *PvPRewardComponent) UpdateHighestRank(tier RankTier) {
	currentHighest, exists := c.HighestSeasonRank[c.CurrentSeasonID]
	if !exists {
		c.HighestSeasonRank[c.CurrentSeasonID] = tier
		return
	}

	// Compare tier indices
	currentIdx := getTierIndex(currentHighest)
	newIdx := getTierIndex(tier)

	if newIdx > currentIdx {
		c.HighestSeasonRank[c.CurrentSeasonID] = tier

		logrus.WithFields(logrus.Fields{
			"component_type": "pvp_reward",
			"season_id":      c.CurrentSeasonID,
			"old_highest":    currentHighest,
			"new_highest":    tier,
		}).Info("New highest season rank")
	}
}

// GetHighestRank returns the highest rank for a season.
func (c *PvPRewardComponent) GetHighestRank(seasonID string) RankTier {
	if tier, exists := c.HighestSeasonRank[seasonID]; exists {
		return tier
	}
	return RankBronze
}

// AddSeasonalReward adds a tier reward for the season.
func (c *PvPRewardComponent) AddSeasonalReward(tier RankTier, seasonID string, rewards []PvPReward) {
	// Check if already has this tier reward
	for _, sr := range c.SeasonalRewards {
		if sr.Tier == tier && sr.SeasonID == seasonID {
			return
		}
	}

	c.SeasonalRewards = append(c.SeasonalRewards, SeasonRewardTier{
		Tier:     tier,
		SeasonID: seasonID,
		Rewards:  rewards,
		Claimed:  false,
	})

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"tier":           tier,
		"season_id":      seasonID,
		"rewards_count":  len(rewards),
	}).Info("Added seasonal reward tier")
}

// ClaimSeasonalReward claims rewards for a tier in a season.
func (c *PvPRewardComponent) ClaimSeasonalReward(tier RankTier, seasonID string) []PvPReward {
	for i := range c.SeasonalRewards {
		if c.SeasonalRewards[i].Tier == tier && c.SeasonalRewards[i].SeasonID == seasonID {
			if c.SeasonalRewards[i].Claimed {
				return nil
			}

			c.SeasonalRewards[i].Claimed = true

			// Add rewards to earned rewards
			for _, r := range c.SeasonalRewards[i].Rewards {
				c.AddReward(r)
			}

			logrus.WithFields(logrus.Fields{
				"component_type": "pvp_reward",
				"tier":           tier,
				"season_id":      seasonID,
			}).Info("Claimed seasonal rewards")

			return c.SeasonalRewards[i].Rewards
		}
	}
	return nil
}

// GetUnclaimedSeasonalRewards returns all unclaimed seasonal reward tiers.
func (c *PvPRewardComponent) GetUnclaimedSeasonalRewards() []SeasonRewardTier {
	unclaimed := make([]SeasonRewardTier, 0)
	for _, sr := range c.SeasonalRewards {
		if !sr.Claimed {
			unclaimed = append(unclaimed, sr)
		}
	}
	return unclaimed
}

// UpdateAchievementProgress updates progress for a PvP achievement.
func (c *PvPRewardComponent) UpdateAchievementProgress(achievementID string, progress, required int) bool {
	// Find existing progress
	for i := range c.AchievementProgress {
		if c.AchievementProgress[i].AchievementID == achievementID {
			c.AchievementProgress[i].CurrentProgress = progress
			if progress >= required && !c.AchievementProgress[i].Completed {
				c.AchievementProgress[i].Completed = true
				return true
			}
			return false
		}
	}

	// Create new progress entry
	newProgress := PvPAchievementProgress{
		AchievementID:   achievementID,
		CurrentProgress: progress,
		RequiredAmount:  required,
		Completed:       progress >= required,
	}
	c.AchievementProgress = append(c.AchievementProgress, newProgress)

	return newProgress.Completed
}

// IncrementAchievementProgress increments progress for a PvP achievement.
func (c *PvPRewardComponent) IncrementAchievementProgress(achievementID string, amount, required int) bool {
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
	newProgress := PvPAchievementProgress{
		AchievementID:   achievementID,
		CurrentProgress: amount,
		RequiredAmount:  required,
		Completed:       amount >= required,
	}
	c.AchievementProgress = append(c.AchievementProgress, newProgress)

	return newProgress.Completed
}

// CompleteAchievement marks an achievement as completed.
func (c *PvPRewardComponent) CompleteAchievement(achievementID string) bool {
	// Check if already completed
	for _, id := range c.CompletedAchievements {
		if id == achievementID {
			return false
		}
	}

	c.CompletedAchievements = append(c.CompletedAchievements, achievementID)

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"achievement_id": achievementID,
	}).Info("Completed PvP achievement")

	return true
}

// HasAchievement checks if an achievement is completed.
func (c *PvPRewardComponent) HasAchievement(achievementID string) bool {
	for _, id := range c.CompletedAchievements {
		if id == achievementID {
			return true
		}
	}
	return false
}

// GetAchievementProgress returns progress for a specific achievement.
func (c *PvPRewardComponent) GetAchievementProgress(achievementID string) *PvPAchievementProgress {
	for i := range c.AchievementProgress {
		if c.AchievementProgress[i].AchievementID == achievementID {
			return &c.AchievementProgress[i]
		}
	}
	return nil
}

// StartNewSeason transitions to a new season.
func (c *PvPRewardComponent) StartNewSeason(newSeasonID string) {
	c.CurrentSeasonID = newSeasonID

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"new_season_id":  newSeasonID,
	}).Info("Started new PvP season")
}

// Serialize encodes the component to bytes for persistence.
func (c *PvPRewardComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type":     "pvp_reward",
		"honor_points":       c.HonorPoints,
		"rewards_count":      len(c.EarnedRewards),
		"achievements_count": len(c.CompletedAchievements),
		"tournament_wins":    c.TournamentWins,
		"titles_count":       len(c.EarnedTitles),
		"mounts_count":       len(c.EarnedMounts),
		"cosmetics_count":    len(c.EarnedCosmetics),
	}).Debug("Serializing PvP reward component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "pvp_reward",
			"error":          err.Error(),
		}).Error("Failed to serialize PvP reward component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *PvPRewardComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"bytes":          len(data),
	}).Debug("Deserializing PvP reward component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "pvp_reward",
			"error":          err.Error(),
		}).Error("Failed to deserialize PvP reward component")
		return err
	}

	// Initialize maps if nil after deserialization
	if c.HighestSeasonRank == nil {
		c.HighestSeasonRank = make(map[string]RankTier)
	}

	return nil
}

// GeneratePvPAchievements generates the standard set of PvP achievements.
func GeneratePvPAchievements(seed int64) []PvPAchievementDef {
	rng := rand.New(rand.NewSource(seed))
	achievements := make([]PvPAchievementDef, 0, 15)

	// Win-based achievements
	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_first_blood",
		Name:           "First Blood",
		Description:    "Win your first PvP match",
		Requirement:    "wins",
		RequiredAmount: 1,
		Reward: PvPReward{
			ID:    "pvp_first_blood_reward",
			Type:  PvPRewardHonor,
			Name:  "First Blood Bonus",
			Value: 50 + rng.Intn(25),
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_veteran",
		Name:           "PvP Veteran",
		Description:    "Win 50 PvP matches",
		Requirement:    "wins",
		RequiredAmount: 50,
		Reward: PvPReward{
			ID:     "pvp_veteran_reward",
			Type:   PvPRewardTitle,
			Name:   "The Veteran",
			Rarity: "rare",
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_centurion",
		Name:           "Centurion",
		Description:    "Win 100 PvP matches",
		Requirement:    "wins",
		RequiredAmount: 100,
		Reward: PvPReward{
			ID:     "pvp_centurion_reward",
			Type:   PvPRewardTitle,
			Name:   "Centurion",
			Rarity: "epic",
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_gladiator",
		Name:           "Gladiator",
		Description:    "Win 500 PvP matches",
		Requirement:    "wins",
		RequiredAmount: 500,
		Reward: PvPReward{
			ID:       "pvp_gladiator_reward",
			Type:     PvPRewardMount,
			Name:     "Gladiator's War Mount",
			Rarity:   "legendary",
			ItemSeed: rng.Int63(),
		},
	})

	// Streak achievements
	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_hot_streak",
		Name:           "Hot Streak",
		Description:    "Win 5 matches in a row",
		Requirement:    "streak",
		RequiredAmount: 5,
		Reward: PvPReward{
			ID:    "pvp_hot_streak_reward",
			Type:  PvPRewardHonor,
			Name:  "Streak Bonus",
			Value: 100 + rng.Intn(50),
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_unstoppable",
		Name:           "Unstoppable",
		Description:    "Win 10 matches in a row",
		Requirement:    "streak",
		RequiredAmount: 10,
		Reward: PvPReward{
			ID:     "pvp_unstoppable_reward",
			Type:   PvPRewardCosmetic,
			Name:   "Unstoppable Aura",
			Rarity: "epic",
		},
	})

	// Rating achievements
	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_rising_star",
		Name:           "Rising Star",
		Description:    "Reach Gold rank",
		Requirement:    "rating",
		RequiredAmount: RankThreshold[RankGold],
		Reward: PvPReward{
			ID:    "pvp_rising_star_reward",
			Type:  PvPRewardHonor,
			Name:  "Rising Star Bonus",
			Value: 200,
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_elite",
		Name:           "Elite",
		Description:    "Reach Diamond rank",
		Requirement:    "rating",
		RequiredAmount: RankThreshold[RankDiamond],
		Reward: PvPReward{
			ID:     "pvp_elite_reward",
			Type:   PvPRewardTitle,
			Name:   "The Elite",
			Rarity: "epic",
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_legend",
		Name:           "Legend",
		Description:    "Reach Legend rank",
		Requirement:    "rating",
		RequiredAmount: RankThreshold[RankLegend],
		Reward: PvPReward{
			ID:       "pvp_legend_reward",
			Type:     PvPRewardMount,
			Name:     "Legendary Steed",
			Rarity:   "legendary",
			ItemSeed: rng.Int63(),
		},
	})

	// Tournament achievements
	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_tournament_novice",
		Name:           "Tournament Novice",
		Description:    "Participate in a tournament",
		Requirement:    "tournament_participation",
		RequiredAmount: 1,
		Reward: PvPReward{
			ID:    "pvp_tournament_novice_reward",
			Type:  PvPRewardHonor,
			Name:  "Tournament Bonus",
			Value: 75 + rng.Intn(25),
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_tournament_champion",
		Name:           "Tournament Champion",
		Description:    "Win a tournament",
		Requirement:    "tournament_wins",
		RequiredAmount: 1,
		Reward: PvPReward{
			ID:     "pvp_tournament_champion_reward",
			Type:   PvPRewardTitle,
			Name:   "Champion",
			Rarity: "rare",
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_tournament_master",
		Name:           "Tournament Master",
		Description:    "Win 10 tournaments",
		Requirement:    "tournament_wins",
		RequiredAmount: 10,
		Reward: PvPReward{
			ID:       "pvp_tournament_master_reward",
			Type:     PvPRewardMount,
			Name:     "Champion's Charger",
			Rarity:   "legendary",
			ItemSeed: rng.Int63(),
		},
	})

	// Honor achievements
	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_honorable",
		Name:           "Honorable",
		Description:    "Earn 1000 Honor Points",
		Requirement:    "honor_earned",
		RequiredAmount: 1000,
		Reward: PvPReward{
			ID:    "pvp_honorable_reward",
			Type:  PvPRewardHonor,
			Name:  "Honor Bonus",
			Value: 250,
		},
	})

	achievements = append(achievements, PvPAchievementDef{
		ID:             "pvp_decorated",
		Name:           "Decorated",
		Description:    "Earn 10000 Honor Points",
		Requirement:    "honor_earned",
		RequiredAmount: 10000,
		Reward: PvPReward{
			ID:     "pvp_decorated_reward",
			Type:   PvPRewardCosmetic,
			Name:   "Medal of Honor Aura",
			Rarity: "legendary",
		},
	})

	logrus.WithFields(logrus.Fields{
		"component_type":     "pvp_reward",
		"achievements_count": len(achievements),
		"seed":               seed,
	}).Debug("Generated PvP achievements")

	return achievements
}

// GenerateSeasonRewards generates rewards for each rank tier in a season.
func GenerateSeasonRewards(seasonID string, seed int64) map[RankTier][]PvPReward {
	rng := rand.New(rand.NewSource(seed))
	rewards := make(map[RankTier][]PvPReward)

	for _, tier := range RankTierOrder {
		tierRewards := make([]PvPReward, 0, 3)

		// Honor reward scales with tier
		honorBase := 100 * (getTierIndex(tier) + 1)
		tierRewards = append(tierRewards, PvPReward{
			ID:       seasonID + "_" + string(tier) + "_honor",
			SeasonID: seasonID,
			Type:     PvPRewardHonor,
			Name:     tierToName(tier) + " Season Bonus",
			Value:    honorBase + rng.Intn(honorBase/2),
			Rarity:   getTierRarity(tier),
		})

		// Title for Gold and above
		if getTierIndex(tier) >= getTierIndex(RankGold) {
			tierRewards = append(tierRewards, PvPReward{
				ID:              seasonID + "_" + string(tier) + "_title",
				SeasonID:        seasonID,
				Type:            PvPRewardTitle,
				Name:            "Season " + tierToName(tier),
				Description:     "Reached " + tierToName(tier) + " in Season " + seasonID,
				Rarity:          getTierRarity(tier),
				RankRequirement: tier,
			})
		}

		// Mount for Diamond and above
		if getTierIndex(tier) >= getTierIndex(RankDiamond) {
			tierRewards = append(tierRewards, PvPReward{
				ID:              seasonID + "_" + string(tier) + "_mount",
				SeasonID:        seasonID,
				Type:            PvPRewardMount,
				Name:            tierToName(tier) + " War Mount",
				Description:     "A mount for " + tierToName(tier) + " warriors",
				Rarity:          getTierRarity(tier),
				RankRequirement: tier,
				ItemSeed:        rng.Int63(),
			})
		}

		// Cosmetic for Master and above
		if getTierIndex(tier) >= getTierIndex(RankMaster) {
			tierRewards = append(tierRewards, PvPReward{
				ID:              seasonID + "_" + string(tier) + "_cosmetic",
				SeasonID:        seasonID,
				Type:            PvPRewardCosmetic,
				Name:            tierToName(tier) + " Aura",
				Description:     "A glowing aura for " + tierToName(tier) + " players",
				Rarity:          getTierRarity(tier),
				RankRequirement: tier,
			})
		}

		rewards[tier] = tierRewards
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "pvp_reward",
		"season_id":      seasonID,
		"tiers":          len(rewards),
		"seed":           seed,
	}).Debug("Generated season rewards")

	return rewards
}

// getTierRarity returns the rarity level for a rank tier.
func getTierRarity(tier RankTier) string {
	switch tier {
	case RankBronze, RankSilver:
		return "common"
	case RankGold:
		return "uncommon"
	case RankPlatinum:
		return "rare"
	case RankDiamond:
		return "epic"
	case RankMaster, RankLegend:
		return "legendary"
	default:
		return "common"
	}
}
