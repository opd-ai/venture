// Package engine provides the event reward system for seasonal event rewards.
// EventRewardSystem manages reward distribution, achievement tracking, and
// event vendor functionality for seasonal events.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// EventVendorItem represents an item for sale at an event vendor.
type EventVendorItem struct {
	// ID is the unique identifier for this vendor item
	ID string `json:"id"`
	// EventID links to the event
	EventID string `json:"event_id"`
	// Name is the display name
	Name string `json:"name"`
	// Description explains the item
	Description string `json:"description"`
	// RewardType is what the player receives
	RewardType EventRewardType `json:"reward_type"`
	// CurrencyCost is how much event currency is required
	CurrencyCost int `json:"currency_cost"`
	// Stock is remaining quantity (-1 for unlimited)
	Stock int `json:"stock"`
	// ItemSeed for generating items
	ItemSeed int64 `json:"item_seed,omitempty"`
	// Rarity of the item
	Rarity string `json:"rarity"`
}

// EventRewardSystem manages event rewards, achievements, and vendors.
// It processes player entities with EventRewardComponent and EventQuestComponent.
type EventRewardSystem struct {
	world *World
	clock GameClock
	// VendorInventory tracks vendor items per event
	VendorInventory map[string][]EventVendorItem
	// EventAchievements tracks achievements per event
	EventAchievements map[string][]EventAchievementDef
	// GeneratedRewards tracks rewards generated per event
	GeneratedRewards map[string][]EventReward
}

// NewEventRewardSystem creates a new event reward system.
func NewEventRewardSystem(world *World, clock GameClock) *EventRewardSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "event_reward",
	}).Debug("Creating event reward system")

	return &EventRewardSystem{
		world:             world,
		clock:             clock,
		VendorInventory:   make(map[string][]EventVendorItem),
		EventAchievements: make(map[string][]EventAchievementDef),
		GeneratedRewards:  make(map[string][]EventReward),
	}
}

// Update processes all entities with event reward components.
// It checks for completed quests, updates achievements, and distributes rewards.
func (s *EventRewardSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("event_reward") {
			continue
		}

		s.processPlayerRewards(entity)
		s.processAchievements(entity)
	}
}

// processPlayerRewards checks for completed event quests and grants rewards.
func (s *EventRewardSystem) processPlayerRewards(entity *Entity) {
	rewardComp := s.getEventRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	questComp := s.getEventQuestComponent(entity)
	if questComp == nil {
		return
	}

	// Check for newly completed quests
	for _, quest := range questComp.CompletedQuests {
		// Grant quest completion reward
		s.grantQuestReward(entity, rewardComp, quest)
	}
}

// processAchievements checks achievement progress and grants rewards.
func (s *EventRewardSystem) processAchievements(entity *Entity) {
	rewardComp := s.getEventRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	questComp := s.getEventQuestComponent(entity)

	// Get active events
	seasonalComp := s.getSeasonalEventComponent()
	if seasonalComp == nil {
		return
	}

	for _, event := range seasonalComp.GetActiveEvents() {
		achievements := s.getOrGenerateAchievements(event)

		for _, ach := range achievements {
			if rewardComp.HasAchievement(ach.ID) {
				continue
			}

			completed := s.checkAchievementCompletion(rewardComp, questComp, ach, event.Definition.ID)
			if completed {
				s.grantAchievementReward(entity, rewardComp, ach)
			}
		}
	}
}

// checkAchievementCompletion checks if an achievement's requirements are met.
func (s *EventRewardSystem) checkAchievementCompletion(
	rewardComp *EventRewardComponent,
	questComp *EventQuestComponent,
	ach EventAchievementDef,
	eventID string,
) bool {
	switch ach.Requirement {
	case "participate":
		return rewardComp.TotalEventsParticipated >= ach.RequiredAmount

	case "complete_quests":
		if questComp == nil {
			return false
		}
		completed := 0
		for _, q := range questComp.CompletedQuests {
			if q.Definition.EventID == eventID {
				completed++
			}
		}
		return completed >= ach.RequiredAmount

	case "earn_currency":
		return rewardComp.GetCurrency(eventID) >= ach.RequiredAmount

	case "defeat_boss":
		if questComp == nil {
			return false
		}
		for _, q := range questComp.CompletedQuests {
			if q.Definition.EventID == eventID && q.Definition.QuestType == EventQuestBoss {
				return true
			}
		}
		return false

	case "explore_location":
		if questComp == nil {
			return false
		}
		for _, q := range questComp.CompletedQuests {
			if q.Definition.EventID == eventID && q.Definition.QuestType == EventQuestExploration {
				return true
			}
		}
		return false

	default:
		progress := rewardComp.GetAchievementProgress(ach.ID)
		if progress != nil {
			return progress.Completed
		}
		return false
	}
}

// grantQuestReward grants rewards for completing an event quest.
func (s *EventRewardSystem) grantQuestReward(entity *Entity, rewardComp *EventRewardComponent, quest EventQuestInstance) {
	// Grant XP and gold from quest reward
	reward := quest.Definition.Reward

	// Add event currency based on quest type
	var currencyAmount int
	switch quest.Definition.QuestType {
	case EventQuestCollection:
		currencyAmount = 25
	case EventQuestExploration:
		currencyAmount = 50
	case EventQuestBoss:
		currencyAmount = 100
	default:
		currencyAmount = 25
	}

	rewardComp.AddCurrency(quest.Definition.EventID, currencyAmount)
	rewardComp.RecordQuestCompletion()

	logrus.WithFields(logrus.Fields{
		"system_name":     "event_reward",
		"entity_id":       entity.ID,
		"quest_id":        quest.Definition.ID,
		"currency_earned": currencyAmount,
		"xp_earned":       reward.XP,
		"gold_earned":     reward.Gold,
	}).Info("Granted quest completion reward")
}

// grantAchievementReward grants rewards for completing an event achievement.
func (s *EventRewardSystem) grantAchievementReward(entity *Entity, rewardComp *EventRewardComponent, ach EventAchievementDef) {
	rewardComp.CompleteAchievement(ach.ID)
	rewardComp.AddReward(ach.Reward)

	// Handle reward type-specific grants
	switch ach.Reward.Type {
	case EventRewardCurrency:
		rewardComp.AddCurrency(ach.EventID, ach.Reward.Value)

	case EventRewardTitle:
		title := EventCosmeticTitle{
			ID:          ach.Reward.ID,
			EventID:     ach.EventID,
			DisplayName: ach.Reward.Name,
			Rarity:      ach.Reward.Rarity,
		}
		rewardComp.AddTitle(title)

	case EventRewardEffect:
		effect := EventVisualEffect{
			ID:         ach.Reward.ID,
			EventID:    ach.EventID,
			Name:       ach.Reward.Name,
			EffectType: "aura",
			Intensity:  0.7,
		}
		rewardComp.AddEffect(effect)
	}

	logrus.WithFields(logrus.Fields{
		"system_name":    "event_reward",
		"entity_id":      entity.ID,
		"achievement_id": ach.ID,
		"reward_type":    ach.Reward.Type,
		"reward_name":    ach.Reward.Name,
	}).Info("Granted achievement reward")
}

// PurchaseFromVendor attempts to purchase an item from an event vendor.
func (s *EventRewardSystem) PurchaseFromVendor(entity *Entity, eventID, vendorItemID string) bool {
	rewardComp := s.getEventRewardComponent(entity)
	if rewardComp == nil {
		return false
	}

	vendor := s.GetVendorInventory(eventID)
	var vendorItem *EventVendorItem
	var itemIndex int
	for i := range vendor {
		if vendor[i].ID == vendorItemID {
			vendorItem = &vendor[i]
			itemIndex = i
			break
		}
	}

	if vendorItem == nil {
		logrus.WithFields(logrus.Fields{
			"system_name":    "event_reward",
			"entity_id":      entity.ID,
			"event_id":       eventID,
			"vendor_item_id": vendorItemID,
		}).Debug("Vendor item not found")
		return false
	}

	// Check stock
	if vendorItem.Stock == 0 {
		logrus.WithFields(logrus.Fields{
			"system_name":    "event_reward",
			"vendor_item_id": vendorItemID,
		}).Debug("Vendor item out of stock")
		return false
	}

	// Check currency
	if !rewardComp.SpendCurrency(eventID, vendorItem.CurrencyCost) {
		logrus.WithFields(logrus.Fields{
			"system_name": "event_reward",
			"entity_id":   entity.ID,
			"cost":        vendorItem.CurrencyCost,
			"available":   rewardComp.GetCurrency(eventID),
		}).Debug("Insufficient currency for vendor purchase")
		return false
	}

	// Decrease stock
	if vendorItem.Stock > 0 {
		s.VendorInventory[eventID][itemIndex].Stock--
	}

	// Grant reward
	reward := EventReward{
		ID:          vendorItemID + "_purchased",
		EventID:     eventID,
		Type:        vendorItem.RewardType,
		Name:        vendorItem.Name,
		Description: vendorItem.Description,
		Value:       1,
		ItemSeed:    vendorItem.ItemSeed,
		Rarity:      vendorItem.Rarity,
	}
	rewardComp.AddReward(reward)

	// Handle type-specific grants
	switch vendorItem.RewardType {
	case EventRewardTitle:
		rewardComp.AddTitle(EventCosmeticTitle{
			ID:          vendorItemID,
			EventID:     eventID,
			DisplayName: vendorItem.Name,
			Rarity:      vendorItem.Rarity,
		})

	case EventRewardEffect:
		rewardComp.AddEffect(EventVisualEffect{
			ID:         vendorItemID,
			EventID:    eventID,
			Name:       vendorItem.Name,
			EffectType: "aura",
			Intensity:  0.8,
		})
	}

	logrus.WithFields(logrus.Fields{
		"system_name":    "event_reward",
		"entity_id":      entity.ID,
		"vendor_item_id": vendorItemID,
		"cost":           vendorItem.CurrencyCost,
	}).Info("Vendor purchase successful")

	return true
}

// GetVendorInventory returns the vendor inventory for an event.
func (s *EventRewardSystem) GetVendorInventory(eventID string) []EventVendorItem {
	if inventory, exists := s.VendorInventory[eventID]; exists {
		return inventory
	}

	// Generate vendor inventory for this event
	seasonalComp := s.getSeasonalEventComponent()
	if seasonalComp == nil {
		return nil
	}

	event := seasonalComp.GetEventByID(eventID)
	if event == nil {
		return nil
	}

	inventory := GenerateEventVendorInventory(*event, event.Definition.Seed)
	s.VendorInventory[eventID] = inventory

	return inventory
}

// getOrGenerateAchievements returns achievements for an event, generating if needed.
func (s *EventRewardSystem) getOrGenerateAchievements(event EventInstance) []EventAchievementDef {
	if achievements, exists := s.EventAchievements[event.Definition.ID]; exists {
		return achievements
	}

	achievements := GenerateEventAchievements(event, event.Definition.Seed)
	s.EventAchievements[event.Definition.ID] = achievements

	return achievements
}

// RegisterEventParticipation records that a player participated in an active event.
func (s *EventRewardSystem) RegisterEventParticipation(entity *Entity) {
	rewardComp := s.getEventRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	seasonalComp := s.getSeasonalEventComponent()
	if seasonalComp == nil {
		return
	}

	activeEvents := seasonalComp.GetActiveEvents()
	if len(activeEvents) > 0 {
		rewardComp.RecordEventParticipation()

		logrus.WithFields(logrus.Fields{
			"system_name":   "event_reward",
			"entity_id":     entity.ID,
			"active_events": len(activeEvents),
		}).Debug("Registered event participation")
	}
}

// GrantEventCurrency grants event currency to a player (for miscellaneous activities).
func (s *EventRewardSystem) GrantEventCurrency(entity *Entity, eventID string, amount int, reason string) {
	rewardComp := s.getEventRewardComponent(entity)
	if rewardComp == nil {
		return
	}

	rewardComp.AddCurrency(eventID, amount)

	logrus.WithFields(logrus.Fields{
		"system_name": "event_reward",
		"entity_id":   entity.ID,
		"event_id":    eventID,
		"amount":      amount,
		"reason":      reason,
	}).Debug("Granted event currency")
}

// GetPlayerEventStats returns event statistics for a player.
func (s *EventRewardSystem) GetPlayerEventStats(entity *Entity) map[string]interface{} {
	rewardComp := s.getEventRewardComponent(entity)
	if rewardComp == nil {
		return nil
	}

	return map[string]interface{}{
		"total_events_participated": rewardComp.TotalEventsParticipated,
		"total_quests_completed":    rewardComp.TotalQuestsCompleted,
		"total_currency_earned":     rewardComp.TotalCurrencyEarned,
		"achievements_completed":    len(rewardComp.CompletedAchievements),
		"titles_earned":             len(rewardComp.EarnedTitles),
		"effects_earned":            len(rewardComp.EarnedEffects),
		"rewards_earned":            len(rewardComp.EarnedRewards),
	}
}

// getEventRewardComponent retrieves the EventRewardComponent from an entity.
func (s *EventRewardSystem) getEventRewardComponent(entity *Entity) *EventRewardComponent {
	comp, ok := entity.GetComponent("event_reward")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*EventRewardComponent)
}

// getEventQuestComponent retrieves the EventQuestComponent from an entity.
func (s *EventRewardSystem) getEventQuestComponent(entity *Entity) *EventQuestComponent {
	comp, ok := entity.GetComponent("event_quest")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*EventQuestComponent)
}

// getSeasonalEventComponent retrieves the SeasonalEventComponent from the world.
func (s *EventRewardSystem) getSeasonalEventComponent() *SeasonalEventComponent {
	if s.world == nil {
		return nil
	}

	// Find the world entity with seasonal event component
	entities := s.world.GetEntities()
	for _, entity := range entities {
		comp, ok := entity.GetComponent("seasonal_event")
		if ok && comp != nil {
			return comp.(*SeasonalEventComponent)
		}
	}

	return nil
}

// GenerateEventVendorInventory generates vendor items for a seasonal event.
func GenerateEventVendorInventory(event EventInstance, seed int64) []EventVendorItem {
	rng := rand.New(rand.NewSource(seed))
	items := make([]EventVendorItem, 0, 8)

	theme := event.Definition.Theme

	// Currency exchange items (common)
	items = append(items, EventVendorItem{
		ID:           event.Definition.ID + "_vendor_potion",
		EventID:      event.Definition.ID,
		Name:         getThemePotionName(theme),
		Description:  "A special potion from the " + event.Definition.Name,
		RewardType:   EventRewardItem,
		CurrencyCost: 25 + rng.Intn(10),
		Stock:        10,
		ItemSeed:     rng.Int63(),
		Rarity:       "common",
	})

	// Uncommon items
	items = append(items, EventVendorItem{
		ID:           event.Definition.ID + "_vendor_accessory",
		EventID:      event.Definition.ID,
		Name:         getThemeAccessoryName(theme),
		Description:  "An accessory from the " + event.Definition.Name,
		RewardType:   EventRewardItem,
		CurrencyCost: 75 + rng.Intn(25),
		Stock:        5,
		ItemSeed:     rng.Int63(),
		Rarity:       "uncommon",
	})

	// Rare item
	rareItems := getThemeItemRewards(theme)
	items = append(items, EventVendorItem{
		ID:           event.Definition.ID + "_vendor_rare",
		EventID:      event.Definition.ID,
		Name:         rareItems[rng.Intn(len(rareItems))],
		Description:  "A rare item from the " + event.Definition.Name,
		RewardType:   EventRewardItem,
		CurrencyCost: 200 + rng.Intn(100),
		Stock:        2,
		ItemSeed:     rng.Int63(),
		Rarity:       "rare",
	})

	// Title (epic)
	titles := getThemeTitles(theme)
	items = append(items, EventVendorItem{
		ID:           event.Definition.ID + "_vendor_title",
		EventID:      event.Definition.ID,
		Name:         titles[rng.Intn(len(titles))],
		Description:  "A title from the " + event.Definition.Name,
		RewardType:   EventRewardTitle,
		CurrencyCost: 500 + rng.Intn(200),
		Stock:        -1, // Unlimited
		Rarity:       "epic",
	})

	// Effect (legendary)
	effects := getThemeEffects(theme)
	items = append(items, EventVendorItem{
		ID:           event.Definition.ID + "_vendor_effect",
		EventID:      event.Definition.ID,
		Name:         effects[rng.Intn(len(effects))],
		Description:  "A visual effect from the " + event.Definition.Name,
		RewardType:   EventRewardEffect,
		CurrencyCost: 1000 + rng.Intn(500),
		Stock:        1,
		Rarity:       "legendary",
	})

	logrus.WithFields(logrus.Fields{
		"system_name": "event_reward",
		"event_id":    event.Definition.ID,
		"items_count": len(items),
		"seed":        seed,
	}).Debug("Generated event vendor inventory")

	return items
}

// getThemePotionName returns potion names for the theme.
func getThemePotionName(theme EventTheme) string {
	switch theme {
	case EventThemeSpring:
		return "Bloom Elixir"
	case EventThemeSummer:
		return "Sunfire Tonic"
	case EventThemeAutumn:
		return "Harvest Brew"
	case EventThemeWinter:
		return "Frost Draught"
	default:
		return "Festival Potion"
	}
}

// getThemeAccessoryName returns accessory names for the theme.
func getThemeAccessoryName(theme EventTheme) string {
	switch theme {
	case EventThemeSpring:
		return "Blossom Pendant"
	case EventThemeSummer:
		return "Solar Bracelet"
	case EventThemeAutumn:
		return "Amber Brooch"
	case EventThemeWinter:
		return "Frost Crystal Ring"
	default:
		return "Festival Charm"
	}
}
