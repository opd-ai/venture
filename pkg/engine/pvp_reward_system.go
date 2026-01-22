// Package engine provides the PvP reward system for competitive PvP rewards.
// PvPRewardSystem manages Honor Points distribution, seasonal rewards, tournament
// rewards, and PvP achievement tracking.
package engine

import (
	"math/rand"

	log "github.com/sirupsen/logrus"
)

// PvPVendorItem represents an item for sale at the PvP vendor.
type PvPVendorItem struct {
	// ID is the unique identifier
	ID string `json:"id"`
	// Name is the display name
	Name string `json:"name"`
	// Description explains the item
	Description string `json:"description"`
	// Type is the reward type
	Type PvPRewardType `json:"type"`
	// HonorCost is how much Honor Points are required
	HonorCost int `json:"honor_cost"`
	// RankRequirement is the minimum rank to purchase
	RankRequirement RankTier `json:"rank_requirement"`
	// Stock is remaining quantity (-1 for unlimited)
	Stock int `json:"stock"`
	// ItemSeed for generating items
	ItemSeed int64 `json:"item_seed,omitempty"`
	// Rarity of the item
	Rarity string `json:"rarity"`
}

// HonorRewardConfig defines how much honor is earned for various activities.
type HonorRewardConfig struct {
	// MatchWin is base honor for winning a match
	MatchWin int
	// MatchLoss is base honor for losing a match
	MatchLoss int
	// TournamentWin is bonus for winning a tournament
	TournamentWin int
	// TournamentParticipation is base for entering a tournament
	TournamentParticipation int
	// TopPlacement is bonus for top tournament placements
	TopPlacement int
	// WinStreakBonus is per-win streak multiplier
	WinStreakBonus int
	// RatingBonusThreshold is rating above which bonus applies
	RatingBonusThreshold int
	// HighRatingMultiplier increases honor for high-rated players
	HighRatingMultiplier float64
}

// DefaultHonorConfig returns the default honor reward configuration.
func DefaultHonorConfig() HonorRewardConfig {
	return HonorRewardConfig{
		MatchWin:                25,
		MatchLoss:               5,
		TournamentWin:           500,
		TournamentParticipation: 50,
		TopPlacement:            100,
		WinStreakBonus:          5,
		RatingBonusThreshold:    1600,
		HighRatingMultiplier:    1.5,
	}
}

// PvPRewardSystem manages PvP rewards, achievements, and the honor vendor.
type PvPRewardSystem struct {
	world        *World
	honorConfig  HonorRewardConfig
	achievements []PvPAchievementDef
	vendorItems  []PvPVendorItem
	seed         int64
}

// NewPvPRewardSystem creates a new PvP reward system.
func NewPvPRewardSystem(world *World, seed int64) *PvPRewardSystem {
	log.WithFields(log.Fields{
		"system_name": "pvp_reward",
		"seed":        seed,
	}).Debug("Creating PvP reward system")

	sys := &PvPRewardSystem{
		world:       world,
		honorConfig: DefaultHonorConfig(),
		seed:        seed,
	}

	// Generate achievements
	sys.achievements = GeneratePvPAchievements(seed)

	// Generate vendor inventory
	sys.vendorItems = sys.generateVendorInventory(seed)

	return sys
}

// Update processes all entities with PvP reward components.
func (s *PvPRewardSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("pvp_reward") {
			continue
		}

		s.processAchievements(entity)
	}
}

// processAchievements checks achievement progress for an entity.
func (s *PvPRewardSystem) processAchievements(entity *Entity) {
	rewardComp := s.getPvPRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	ratingComp := s.getPvPRatingComponent(entity)
	tournamentComp := s.getTournamentComponent(entity)

	for _, ach := range s.achievements {
		if rewardComp.HasAchievement(ach.ID) {
			continue
		}

		completed := s.checkAchievementCompletion(rewardComp, ratingComp, tournamentComp, ach)
		if completed {
			s.grantAchievementReward(entity, rewardComp, ach)
		}
	}
}

// checkAchievementCompletion checks if an achievement's requirements are met.
func (s *PvPRewardSystem) checkAchievementCompletion(
	rewardComp *PvPRewardComponent,
	ratingComp *PvPRatingComponent,
	tournamentComp *TournamentComponent,
	ach PvPAchievementDef,
) bool {
	switch ach.Requirement {
	case "wins":
		if ratingComp == nil {
			return false
		}
		return ratingComp.Wins >= ach.RequiredAmount

	case "streak":
		if ratingComp == nil {
			return false
		}
		return ratingComp.MatchStreak >= ach.RequiredAmount

	case "rating":
		if ratingComp == nil {
			return false
		}
		return ratingComp.PeakRating >= ach.RequiredAmount

	case "tournament_participation":
		return rewardComp.TournamentParticipations >= ach.RequiredAmount

	case "tournament_wins":
		return rewardComp.TournamentWins >= ach.RequiredAmount

	case "honor_earned":
		return rewardComp.TotalHonorEarned >= ach.RequiredAmount

	default:
		progress := rewardComp.GetAchievementProgress(ach.ID)
		if progress != nil {
			return progress.Completed
		}
		return false
	}
}

// grantAchievementReward grants rewards for completing a PvP achievement.
func (s *PvPRewardSystem) grantAchievementReward(entity *Entity, rewardComp *PvPRewardComponent, ach PvPAchievementDef) {
	rewardComp.CompleteAchievement(ach.ID)

	// Handle reward type-specific grants
	switch ach.Reward.Type {
	case PvPRewardHonor:
		rewardComp.AddHonor(ach.Reward.Value)

	case PvPRewardTitle:
		rewardComp.AddTitle(ach.Reward.ID)
		rewardComp.AddReward(ach.Reward)

	case PvPRewardMount:
		rewardComp.AddMount(ach.Reward.ID)
		rewardComp.AddReward(ach.Reward)

	case PvPRewardCosmetic:
		rewardComp.AddCosmetic(ach.Reward.ID)
		rewardComp.AddReward(ach.Reward)

	case PvPRewardItem:
		rewardComp.AddReward(ach.Reward)
	}

	log.WithFields(log.Fields{
		"system_name":    "pvp_reward",
		"entity_id":      entity.ID,
		"achievement_id": ach.ID,
		"reward_type":    ach.Reward.Type,
		"reward_name":    ach.Reward.Name,
	}).Info("Granted PvP achievement reward")
}

// GrantMatchReward grants honor and updates achievements after a match.
func (s *PvPRewardSystem) GrantMatchReward(entity *Entity, won bool, ratingChange, streak int) {
	rewardComp := s.getPvPRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	ratingComp := s.getPvPRatingComponent(entity)

	// Calculate base honor
	var honor int
	if won {
		honor = s.honorConfig.MatchWin
	} else {
		honor = s.honorConfig.MatchLoss
	}

	// Apply streak bonus for wins
	if won && streak > 1 {
		honor += s.honorConfig.WinStreakBonus * (streak - 1)
	}

	// Apply high rating multiplier
	if ratingComp != nil && ratingComp.Rating >= s.honorConfig.RatingBonusThreshold {
		honor = int(float64(honor) * s.honorConfig.HighRatingMultiplier)
	}

	rewardComp.AddHonor(honor)

	// Update highest rank
	if ratingComp != nil {
		rewardComp.UpdateHighestRank(ratingComp.RankTier)
	}

	log.WithFields(log.Fields{
		"system_name": "pvp_reward",
		"entity_id":   entity.ID,
		"won":         won,
		"honor":       honor,
		"streak":      streak,
	}).Debug("Granted match reward")
}

// GrantTournamentReward grants honor for tournament participation and placement.
func (s *PvPRewardSystem) GrantTournamentReward(entity *Entity, placement, totalParticipants int, won bool) {
	rewardComp := s.getPvPRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	// Base participation reward
	honor := s.honorConfig.TournamentParticipation

	// Placement bonus (top 4 get extra)
	if placement <= 4 {
		honor += s.honorConfig.TopPlacement * (5 - placement)
	}

	// Tournament win bonus
	if won {
		honor += s.honorConfig.TournamentWin
		rewardComp.RecordTournamentWin()
	} else {
		rewardComp.RecordTournamentParticipation()
	}

	rewardComp.AddHonor(honor)

	log.WithFields(log.Fields{
		"system_name":  "pvp_reward",
		"entity_id":    entity.ID,
		"placement":    placement,
		"participants": totalParticipants,
		"won":          won,
		"honor":        honor,
	}).Info("Granted tournament reward")
}

// DistributeSeasonRewards distributes end-of-season rewards based on highest rank.
func (s *PvPRewardSystem) DistributeSeasonRewards(entities []*Entity, seasonID string, rewardSeed int64) {
	seasonRewards := GenerateSeasonRewards(seasonID, rewardSeed)

	for _, entity := range entities {
		rewardComp := s.getPvPRewardComponent(entity)
		if rewardComp == nil {
			continue
		}

		highestRank := rewardComp.GetHighestRank(seasonID)

		// Grant rewards for each tier up to highest achieved
		for _, tier := range RankTierOrder {
			if getTierIndex(tier) > getTierIndex(highestRank) {
				break
			}

			tierRewards, exists := seasonRewards[tier]
			if !exists {
				continue
			}

			rewardComp.AddSeasonalReward(tier, seasonID, tierRewards)
		}

		log.WithFields(log.Fields{
			"system_name":  "pvp_reward",
			"entity_id":    entity.ID,
			"season_id":    seasonID,
			"highest_rank": highestRank,
		}).Info("Distributed season rewards")
	}
}

// PurchaseFromVendor attempts to purchase an item from the PvP vendor.
func (s *PvPRewardSystem) PurchaseFromVendor(entity *Entity, vendorItemID string) bool {
	rewardComp := s.getPvPRewardComponent(entity)
	if rewardComp == nil {
		return false
	}

	vendorItem, itemIndex := s.findVendorItem(vendorItemID)
	if vendorItem == nil {
		s.logVendorItemNotFound(entity.ID, vendorItemID)
		return false
	}

	if !s.validatePurchaseRequirements(entity, vendorItem) {
		return false
	}

	if !rewardComp.SpendHonor(vendorItem.HonorCost) {
		s.logInsufficientHonor(entity.ID, vendorItem.HonorCost, rewardComp.GetHonor())
		return false
	}

	s.decreaseStock(itemIndex, vendorItem)
	s.grantVendorReward(entity, rewardComp, vendorItem, vendorItemID)
	s.logPurchaseSuccess(entity.ID, vendorItemID, vendorItem.HonorCost)

	return true
}

// findVendorItem locates a vendor item by ID and returns it with its index.
func (s *PvPRewardSystem) findVendorItem(vendorItemID string) (*PvPVendorItem, int) {
	for i := range s.vendorItems {
		if s.vendorItems[i].ID == vendorItemID {
			return &s.vendorItems[i], i
		}
	}
	return nil, -1
}

// validatePurchaseRequirements checks rank and stock requirements.
func (s *PvPRewardSystem) validatePurchaseRequirements(entity *Entity, vendorItem *PvPVendorItem) bool {
	if !s.checkRankRequirement(entity, vendorItem) {
		return false
	}
	return s.checkStock(vendorItem)
}

// checkRankRequirement validates the player meets the rank requirement.
func (s *PvPRewardSystem) checkRankRequirement(entity *Entity, vendorItem *PvPVendorItem) bool {
	if vendorItem.RankRequirement == "" {
		return true
	}

	ratingComp := s.getPvPRatingComponent(entity)
	if ratingComp == nil {
		return true
	}

	if getTierIndex(ratingComp.RankTier) < getTierIndex(vendorItem.RankRequirement) {
		log.WithFields(log.Fields{
			"system_name":    "pvp_reward",
			"vendor_item_id": vendorItem.ID,
			"required_rank":  vendorItem.RankRequirement,
			"current_rank":   ratingComp.RankTier,
		}).Debug("Rank requirement not met")
		return false
	}
	return true
}

// checkStock validates the item is in stock.
func (s *PvPRewardSystem) checkStock(vendorItem *PvPVendorItem) bool {
	if vendorItem.Stock == 0 {
		log.WithFields(log.Fields{
			"system_name":    "pvp_reward",
			"vendor_item_id": vendorItem.ID,
		}).Debug("Vendor item out of stock")
		return false
	}
	return true
}

// decreaseStock reduces the stock count for limited items.
func (s *PvPRewardSystem) decreaseStock(itemIndex int, vendorItem *PvPVendorItem) {
	if vendorItem.Stock > 0 {
		s.vendorItems[itemIndex].Stock--
	}
}

// grantVendorReward creates and grants the purchased reward to the player.
func (s *PvPRewardSystem) grantVendorReward(entity *Entity, rewardComp *PvPRewardComponent, vendorItem *PvPVendorItem, vendorItemID string) {
	reward := PvPReward{
		ID:          vendorItemID + "_purchased",
		Type:        vendorItem.Type,
		Name:        vendorItem.Name,
		Description: vendorItem.Description,
		Value:       1,
		ItemSeed:    vendorItem.ItemSeed,
		Rarity:      vendorItem.Rarity,
	}

	applyRewardByType(rewardComp, vendorItem.Type, vendorItemID)
	rewardComp.AddReward(reward)
}

// applyRewardByType applies type-specific reward grants.
func applyRewardByType(rewardComp *PvPRewardComponent, rewardType PvPRewardType, vendorItemID string) {
	switch rewardType {
	case PvPRewardTitle:
		rewardComp.AddTitle(vendorItemID)
	case PvPRewardMount:
		rewardComp.AddMount(vendorItemID)
	case PvPRewardCosmetic:
		rewardComp.AddCosmetic(vendorItemID)
	}
}

// logVendorItemNotFound logs when a vendor item is not found.
func (s *PvPRewardSystem) logVendorItemNotFound(entityID uint64, vendorItemID string) {
	log.WithFields(log.Fields{
		"system_name":    "pvp_reward",
		"entity_id":      entityID,
		"vendor_item_id": vendorItemID,
	}).Debug("Vendor item not found")
}

// logInsufficientHonor logs when the player has insufficient honor.
func (s *PvPRewardSystem) logInsufficientHonor(entityID uint64, cost, available int) {
	log.WithFields(log.Fields{
		"system_name": "pvp_reward",
		"entity_id":   entityID,
		"cost":        cost,
		"available":   available,
	}).Debug("Insufficient honor for vendor purchase")
}

// logPurchaseSuccess logs a successful vendor purchase.
func (s *PvPRewardSystem) logPurchaseSuccess(entityID uint64, vendorItemID string, cost int) {
	log.WithFields(log.Fields{
		"system_name":    "pvp_reward",
		"entity_id":      entityID,
		"vendor_item_id": vendorItemID,
		"cost":           cost,
	}).Info("Vendor purchase successful")
}

// GetVendorInventory returns all available vendor items.
func (s *PvPRewardSystem) GetVendorInventory() []PvPVendorItem {
	return s.vendorItems
}

// GetVendorItemsForRank returns vendor items available at a specific rank.
func (s *PvPRewardSystem) GetVendorItemsForRank(rank RankTier) []PvPVendorItem {
	available := make([]PvPVendorItem, 0)
	rankIdx := getTierIndex(rank)

	for _, item := range s.vendorItems {
		if item.RankRequirement == "" {
			available = append(available, item)
			continue
		}

		reqIdx := getTierIndex(item.RankRequirement)
		if rankIdx >= reqIdx {
			available = append(available, item)
		}
	}

	return available
}

// GetPlayerPvPStats returns PvP statistics for a player.
func (s *PvPRewardSystem) GetPlayerPvPStats(entity *Entity) map[string]interface{} {
	rewardComp := s.getPvPRewardComponent(entity)
	if rewardComp == nil {
		return nil
	}

	ratingComp := s.getPvPRatingComponent(entity)

	stats := map[string]interface{}{
		"honor_points":              rewardComp.HonorPoints,
		"total_honor_earned":        rewardComp.TotalHonorEarned,
		"tournament_wins":           rewardComp.TournamentWins,
		"tournament_participations": rewardComp.TournamentParticipations,
		"achievements_completed":    len(rewardComp.CompletedAchievements),
		"titles_earned":             len(rewardComp.EarnedTitles),
		"mounts_earned":             len(rewardComp.EarnedMounts),
		"cosmetics_earned":          len(rewardComp.EarnedCosmetics),
		"rewards_earned":            len(rewardComp.EarnedRewards),
	}

	if ratingComp != nil {
		stats["current_rating"] = ratingComp.Rating
		stats["peak_rating"] = ratingComp.PeakRating
		stats["rank_tier"] = ratingComp.RankTier
		stats["rank_division"] = ratingComp.RankDivision
		stats["wins"] = ratingComp.Wins
		stats["losses"] = ratingComp.Losses
		stats["win_rate"] = ratingComp.GetWinRate()
	}

	return stats
}

// GetAchievements returns all PvP achievements.
func (s *PvPRewardSystem) GetAchievements() []PvPAchievementDef {
	return s.achievements
}

// GetPlayerAchievementProgress returns achievement progress for a player.
func (s *PvPRewardSystem) GetPlayerAchievementProgress(entity *Entity) map[string]PvPAchievementProgress {
	rewardComp := s.getPvPRewardComponent(entity)
	if rewardComp == nil {
		return nil
	}

	progress := make(map[string]PvPAchievementProgress)
	for _, ach := range s.achievements {
		if rewardComp.HasAchievement(ach.ID) {
			progress[ach.ID] = PvPAchievementProgress{
				AchievementID:   ach.ID,
				CurrentProgress: ach.RequiredAmount,
				RequiredAmount:  ach.RequiredAmount,
				Completed:       true,
			}
		} else if p := rewardComp.GetAchievementProgress(ach.ID); p != nil {
			progress[ach.ID] = *p
		} else {
			// Calculate current progress based on requirement
			currentProgress := s.calculateAchievementProgress(entity, rewardComp, ach)
			progress[ach.ID] = PvPAchievementProgress{
				AchievementID:   ach.ID,
				CurrentProgress: currentProgress,
				RequiredAmount:  ach.RequiredAmount,
				Completed:       currentProgress >= ach.RequiredAmount,
			}
		}
	}

	return progress
}

// calculateAchievementProgress calculates current progress for an achievement.
func (s *PvPRewardSystem) calculateAchievementProgress(entity *Entity, rewardComp *PvPRewardComponent, ach PvPAchievementDef) int {
	ratingComp := s.getPvPRatingComponent(entity)

	switch ach.Requirement {
	case "wins":
		if ratingComp != nil {
			return ratingComp.Wins
		}
	case "streak":
		if ratingComp != nil && ratingComp.MatchStreak > 0 {
			return ratingComp.MatchStreak
		}
	case "rating":
		if ratingComp != nil {
			return ratingComp.PeakRating
		}
	case "tournament_participation":
		return rewardComp.TournamentParticipations
	case "tournament_wins":
		return rewardComp.TournamentWins
	case "honor_earned":
		return rewardComp.TotalHonorEarned
	}

	return 0
}

// generateVendorInventory generates the PvP vendor inventory.
func (s *PvPRewardSystem) generateVendorInventory(seed int64) []PvPVendorItem {
	rng := rand.New(rand.NewSource(seed))
	items := make([]PvPVendorItem, 0, 15)

	// Basic items (no rank requirement)
	items = append(items, PvPVendorItem{
		ID:          "pvp_vendor_health_potion",
		Name:        "Warrior's Healing Draught",
		Description: "A potent healing potion for PvP combat",
		Type:        PvPRewardItem,
		HonorCost:   50,
		Stock:       -1,
		ItemSeed:    rng.Int63(),
		Rarity:      "common",
	})

	items = append(items, PvPVendorItem{
		ID:          "pvp_vendor_damage_potion",
		Name:        "Berserker's Elixir",
		Description: "Temporarily increases combat damage",
		Type:        PvPRewardItem,
		HonorCost:   75,
		Stock:       -1,
		ItemSeed:    rng.Int63(),
		Rarity:      "common",
	})

	// Silver tier items
	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_silver_insignia",
		Name:            "Silver Combatant Insignia",
		Description:     "A badge marking your PvP prowess",
		Type:            PvPRewardItem,
		HonorCost:       200,
		RankRequirement: RankSilver,
		Stock:           -1,
		ItemSeed:        rng.Int63(),
		Rarity:          "uncommon",
	})

	// Gold tier items
	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_gold_weapon",
		Name:            "Gladiator's Blade",
		Description:     "A weapon forged for arena combat",
		Type:            PvPRewardItem,
		HonorCost:       500,
		RankRequirement: RankGold,
		Stock:           5,
		ItemSeed:        rng.Int63(),
		Rarity:          "rare",
	})

	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_gold_title",
		Name:            "Duelist",
		Description:     "A title for skilled duelists",
		Type:            PvPRewardTitle,
		HonorCost:       750,
		RankRequirement: RankGold,
		Stock:           -1,
		Rarity:          "rare",
	})

	// Platinum tier items
	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_platinum_armor",
		Name:            "Platinum Champion's Armor",
		Description:     "Armor worn by proven champions",
		Type:            PvPRewardItem,
		HonorCost:       1000,
		RankRequirement: RankPlatinum,
		Stock:           3,
		ItemSeed:        rng.Int63(),
		Rarity:          "rare",
	})

	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_platinum_cosmetic",
		Name:            "Warrior's Flame",
		Description:     "A fiery aura for battle",
		Type:            PvPRewardCosmetic,
		HonorCost:       1500,
		RankRequirement: RankPlatinum,
		Stock:           -1,
		Rarity:          "epic",
	})

	// Diamond tier items
	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_diamond_mount",
		Name:            "Warhorse",
		Description:     "A battle-trained mount",
		Type:            PvPRewardMount,
		HonorCost:       3000,
		RankRequirement: RankDiamond,
		Stock:           2,
		ItemSeed:        rng.Int63(),
		Rarity:          "epic",
	})

	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_diamond_title",
		Name:            "Warlord",
		Description:     "A title for battlefield leaders",
		Type:            PvPRewardTitle,
		HonorCost:       2000,
		RankRequirement: RankDiamond,
		Stock:           -1,
		Rarity:          "epic",
	})

	// Master tier items
	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_master_weapon",
		Name:            "Master's Deathblade",
		Description:     "A legendary weapon of destruction",
		Type:            PvPRewardItem,
		HonorCost:       5000,
		RankRequirement: RankMaster,
		Stock:           1,
		ItemSeed:        rng.Int63(),
		Rarity:          "legendary",
	})

	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_master_cosmetic",
		Name:            "Champion's Radiance",
		Description:     "A blinding aura of victory",
		Type:            PvPRewardCosmetic,
		HonorCost:       4000,
		RankRequirement: RankMaster,
		Stock:           -1,
		Rarity:          "legendary",
	})

	// Legend tier items
	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_legend_mount",
		Name:            "Nightmare Charger",
		Description:     "A mythical mount from the arena",
		Type:            PvPRewardMount,
		HonorCost:       10000,
		RankRequirement: RankLegend,
		Stock:           1,
		ItemSeed:        rng.Int63(),
		Rarity:          "legendary",
	})

	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_legend_title",
		Name:            "Grand Marshal",
		Description:     "The highest PvP honor",
		Type:            PvPRewardTitle,
		HonorCost:       7500,
		RankRequirement: RankLegend,
		Stock:           -1,
		Rarity:          "legendary",
	})

	items = append(items, PvPVendorItem{
		ID:              "pvp_vendor_legend_cosmetic",
		Name:            "Legend's Crown",
		Description:     "A crown of pure glory",
		Type:            PvPRewardCosmetic,
		HonorCost:       8000,
		RankRequirement: RankLegend,
		Stock:           -1,
		Rarity:          "legendary",
	})

	log.WithFields(log.Fields{
		"system_name": "pvp_reward",
		"items_count": len(items),
		"seed":        seed,
	}).Debug("Generated PvP vendor inventory")

	return items
}

// getPvPRewardComponent retrieves the PvPRewardComponent from an entity.
func (s *PvPRewardSystem) getPvPRewardComponent(entity *Entity) *PvPRewardComponent {
	comp, ok := entity.GetComponent("pvp_reward")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*PvPRewardComponent)
}

// getPvPRatingComponent retrieves the PvPRatingComponent from an entity.
func (s *PvPRewardSystem) getPvPRatingComponent(entity *Entity) *PvPRatingComponent {
	comp, ok := entity.GetComponent("pvp_rating")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*PvPRatingComponent)
}

// getTournamentComponent retrieves the TournamentComponent from an entity.
func (s *PvPRewardSystem) getTournamentComponent(entity *Entity) *TournamentComponent {
	comp, ok := entity.GetComponent("tournament")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*TournamentComponent)
}

// SetHonorConfig updates the honor reward configuration.
func (s *PvPRewardSystem) SetHonorConfig(config HonorRewardConfig) {
	s.honorConfig = config
}

// GetHonorConfig returns the current honor reward configuration.
func (s *PvPRewardSystem) GetHonorConfig() HonorRewardConfig {
	return s.honorConfig
}
