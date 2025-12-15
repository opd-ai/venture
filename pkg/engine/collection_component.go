// Package engine provides collection tracking components for the ECS.
// This file implements components for collectible items and player collection progress.
// Phase 97: Collection System (V18.0)

package engine

import (
	"encoding/json"
	"sync"
)

// CollectionCategory defines the category of collectibles.
type CollectionCategory string

const (
	// CollectionCategoryFish for caught fish types.
	CollectionCategoryFish CollectionCategory = "fish"
	// CollectionCategoryResources for gathered resource types.
	CollectionCategoryResources CollectionCategory = "resources"
	// CollectionCategoryCreatures for encountered creature types.
	CollectionCategoryCreatures CollectionCategory = "creatures"
	// CollectionCategoryArtifacts for discovered artifacts.
	CollectionCategoryArtifacts CollectionCategory = "artifacts"
	// CollectionCategoryLore for found lore items and books.
	CollectionCategoryLore CollectionCategory = "lore"
	// CollectionCategoryRecipes for learned crafting recipes.
	CollectionCategoryRecipes CollectionCategory = "recipes"
	// CollectionCategoryCosmetics for unlocked cosmetic items.
	CollectionCategoryCosmetics CollectionCategory = "cosmetics"
	// CollectionCategoryAchievements for earned achievements.
	CollectionCategoryAchievements CollectionCategory = "achievements"
)

// AllCollectionCategories returns all collection categories.
func AllCollectionCategories() []CollectionCategory {
	return []CollectionCategory{
		CollectionCategoryFish,
		CollectionCategoryResources,
		CollectionCategoryCreatures,
		CollectionCategoryArtifacts,
		CollectionCategoryLore,
		CollectionCategoryRecipes,
		CollectionCategoryCosmetics,
		CollectionCategoryAchievements,
	}
}

// String returns the string representation of the category.
func (c CollectionCategory) String() string {
	return string(c)
}

// CollectionRarity defines rarity levels for collectibles.
type CollectionRarity string

const (
	// CollectionRarityCommon for frequently found items.
	CollectionRarityCommon CollectionRarity = "common"
	// CollectionRarityUncommon for moderately rare items.
	CollectionRarityUncommon CollectionRarity = "uncommon"
	// CollectionRarityRare for rare items.
	CollectionRarityRare CollectionRarity = "rare"
	// CollectionRarityEpic for very rare items.
	CollectionRarityEpic CollectionRarity = "epic"
	// CollectionRarityLegendary for extremely rare items.
	CollectionRarityLegendary CollectionRarity = "legendary"
)

// Points returns the collection points for this rarity.
func (r CollectionRarity) Points() int {
	switch r {
	case CollectionRarityCommon:
		return 1
	case CollectionRarityUncommon:
		return 3
	case CollectionRarityRare:
		return 5
	case CollectionRarityEpic:
		return 10
	case CollectionRarityLegendary:
		return 25
	default:
		return 1
	}
}

// CollectionMilestone defines completion thresholds for rewards.
type CollectionMilestone struct {
	// Threshold is the percentage (0-100) required to unlock this milestone.
	Threshold int `json:"threshold"`
	// Name is the display name of the milestone.
	Name string `json:"name"`
	// RewardType is the type of reward (title, cosmetic, bonus).
	RewardType string `json:"reward_type"`
	// RewardID is the identifier for the reward item.
	RewardID string `json:"reward_id"`
	// Points is bonus collection points awarded at this milestone.
	Points int `json:"points"`
}

// DefaultMilestones returns the standard completion milestones.
func DefaultMilestones() []CollectionMilestone {
	return []CollectionMilestone{
		{Threshold: 25, Name: "Novice Collector", RewardType: "title", RewardID: "title_novice_collector", Points: 50},
		{Threshold: 50, Name: "Keen Collector", RewardType: "title", RewardID: "title_keen_collector", Points: 100},
		{Threshold: 75, Name: "Expert Collector", RewardType: "cosmetic", RewardID: "cosmetic_collector_badge", Points: 200},
		{Threshold: 100, Name: "Master Collector", RewardType: "bonus", RewardID: "bonus_collection_luck", Points: 500},
	}
}

// CollectedEntry tracks a single discovered collectible.
type CollectedEntry struct {
	// ID is the unique identifier for this collectible.
	ID string `json:"id"`
	// Name is the display name.
	Name string `json:"name"`
	// Category is the collection category.
	Category CollectionCategory `json:"category"`
	// Rarity is the item rarity.
	Rarity CollectionRarity `json:"rarity"`
	// Description is flavor text for the collectible.
	Description string `json:"description"`
	// DiscoveredAt is the unix timestamp of first discovery.
	DiscoveredAt int64 `json:"discovered_at"`
	// Count is the number of times this item has been collected.
	Count int `json:"count"`
}

// CollectionComponent tracks a player's collectibles and completion progress.
type CollectionComponent struct {
	mu sync.RWMutex

	// Discovered maps collectible IDs to their entries.
	Discovered map[string]*CollectedEntry `json:"discovered"`

	// CategoryCounts tracks discovered count per category.
	CategoryCounts map[CollectionCategory]int `json:"category_counts"`

	// TotalInCategory tracks total possible per category.
	TotalInCategory map[CollectionCategory]int `json:"total_in_category"`

	// TotalPoints is the cumulative collection points earned.
	TotalPoints int `json:"total_points"`

	// ClaimedRewards tracks milestone rewards that have been claimed.
	ClaimedRewards map[string]bool `json:"claimed_rewards"`

	// UnlockedMilestones tracks milestones achieved per category.
	UnlockedMilestones map[CollectionCategory][]int `json:"unlocked_milestones"`

	// FavoriteCollectibles are player-marked favorites.
	FavoriteCollectibles []string `json:"favorite_collectibles"`
}

// NewCollectionComponent creates a new collection component with defaults.
func NewCollectionComponent() *CollectionComponent {
	c := &CollectionComponent{
		Discovered:           make(map[string]*CollectedEntry),
		CategoryCounts:       make(map[CollectionCategory]int),
		TotalInCategory:      make(map[CollectionCategory]int),
		TotalPoints:          0,
		ClaimedRewards:       make(map[string]bool),
		UnlockedMilestones:   make(map[CollectionCategory][]int),
		FavoriteCollectibles: make([]string, 0),
	}

	// Initialize categories with default totals
	for _, cat := range AllCollectionCategories() {
		c.CategoryCounts[cat] = 0
		c.TotalInCategory[cat] = GetCategoryTotal(cat)
		c.UnlockedMilestones[cat] = make([]int, 0)
	}

	return c
}

// GetCategoryTotal returns the total collectibles in a category.
func GetCategoryTotal(cat CollectionCategory) int {
	// Default totals based on game content
	totals := map[CollectionCategory]int{
		CollectionCategoryFish:         14, // From Phase 96 fishing system
		CollectionCategoryResources:    6,  // From Phase 95 gathering system
		CollectionCategoryCreatures:    50, // Various enemy types
		CollectionCategoryArtifacts:    25, // Quest rewards and hidden items
		CollectionCategoryLore:         30, // Books and lore items
		CollectionCategoryRecipes:      40, // Crafting recipes
		CollectionCategoryCosmetics:    20, // Cosmetic unlocks
		CollectionCategoryAchievements: 60, // From Phase 83 achievements
	}
	if total, ok := totals[cat]; ok {
		return total
	}
	return 10
}

// Type returns the component type identifier.
func (c *CollectionComponent) Type() string {
	return "collection"
}

// AddCollectible adds a new collectible or increments count if already discovered.
// Returns true if this is a new discovery, false if already discovered.
func (c *CollectionComponent) AddCollectible(id, name string, category CollectionCategory, rarity CollectionRarity, description string, timestamp int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.Discovered[id]; exists {
		entry.Count++
		return false
	}

	c.Discovered[id] = &CollectedEntry{
		ID:           id,
		Name:         name,
		Category:     category,
		Rarity:       rarity,
		Description:  description,
		DiscoveredAt: timestamp,
		Count:        1,
	}

	c.CategoryCounts[category]++
	c.TotalPoints += rarity.Points()

	return true
}

// HasCollectible checks if a collectible has been discovered.
func (c *CollectionComponent) HasCollectible(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.Discovered[id]
	return exists
}

// GetCollectible returns a collectible entry by ID.
func (c *CollectionComponent) GetCollectible(id string) *CollectedEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if entry, exists := c.Discovered[id]; exists {
		// Return a copy
		entryCopy := *entry
		return &entryCopy
	}
	return nil
}

// GetCategoryProgress returns discovered/total for a category.
func (c *CollectionComponent) GetCategoryProgress(category CollectionCategory) (discovered, total int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CategoryCounts[category], c.TotalInCategory[category]
}

// GetCategoryCompletionPercent returns completion percentage (0-100) for a category.
func (c *CollectionComponent) GetCategoryCompletionPercent(category CollectionCategory) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.TotalInCategory[category]
	if total == 0 {
		return 0
	}
	return float64(c.CategoryCounts[category]) / float64(total) * 100
}

// GetOverallCompletionPercent returns overall completion percentage.
func (c *CollectionComponent) GetOverallCompletionPercent() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalDiscovered, totalPossible int
	for _, cat := range AllCollectionCategories() {
		totalDiscovered += c.CategoryCounts[cat]
		totalPossible += c.TotalInCategory[cat]
	}

	if totalPossible == 0 {
		return 0
	}
	return float64(totalDiscovered) / float64(totalPossible) * 100
}

// GetTotalPoints returns total collection points.
func (c *CollectionComponent) GetTotalPoints() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TotalPoints
}

// GetDiscoveredCount returns total discovered collectibles.
func (c *CollectionComponent) GetDiscoveredCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Discovered)
}

// GetDiscoveredInCategory returns all discovered entries in a category.
func (c *CollectionComponent) GetDiscoveredInCategory(category CollectionCategory) []*CollectedEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*CollectedEntry, 0)
	for _, entry := range c.Discovered {
		if entry.Category == category {
			entryCopy := *entry
			entries = append(entries, &entryCopy)
		}
	}
	return entries
}

// CheckMilestone checks if a milestone threshold is reached for a category.
// Returns the milestone if newly reached, nil otherwise.
func (c *CollectionComponent) CheckMilestone(category CollectionCategory, threshold int) *CollectionMilestone {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already unlocked
	for _, t := range c.UnlockedMilestones[category] {
		if t == threshold {
			return nil
		}
	}

	// Check if threshold is reached
	total := c.TotalInCategory[category]
	if total == 0 {
		return nil
	}
	percent := float64(c.CategoryCounts[category]) / float64(total) * 100

	if int(percent) >= threshold {
		c.UnlockedMilestones[category] = append(c.UnlockedMilestones[category], threshold)
		for _, m := range DefaultMilestones() {
			if m.Threshold == threshold {
				return &m
			}
		}
	}
	return nil
}

// ClaimReward marks a reward as claimed. Returns true if successfully claimed.
func (c *CollectionComponent) ClaimReward(rewardID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ClaimedRewards[rewardID] {
		return false
	}
	c.ClaimedRewards[rewardID] = true
	return true
}

// HasClaimedReward checks if a reward has been claimed.
func (c *CollectionComponent) HasClaimedReward(rewardID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ClaimedRewards[rewardID]
}

// AddFavorite adds a collectible to favorites.
func (c *CollectionComponent) AddFavorite(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already favorite
	for _, fav := range c.FavoriteCollectibles {
		if fav == id {
			return false
		}
	}

	// Check if collectible exists
	if _, exists := c.Discovered[id]; !exists {
		return false
	}

	c.FavoriteCollectibles = append(c.FavoriteCollectibles, id)
	return true
}

// RemoveFavorite removes a collectible from favorites.
func (c *CollectionComponent) RemoveFavorite(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, fav := range c.FavoriteCollectibles {
		if fav == id {
			c.FavoriteCollectibles = append(c.FavoriteCollectibles[:i], c.FavoriteCollectibles[i+1:]...)
			return true
		}
	}
	return false
}

// IsFavorite checks if a collectible is marked as favorite.
func (c *CollectionComponent) IsFavorite(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, fav := range c.FavoriteCollectibles {
		if fav == id {
			return true
		}
	}
	return false
}

// GetFavorites returns all favorite collectible IDs.
func (c *CollectionComponent) GetFavorites() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.FavoriteCollectibles))
	copy(result, c.FavoriteCollectibles)
	return result
}

// SetCategoryTotal sets the total collectibles for a category.
func (c *CollectionComponent) SetCategoryTotal(category CollectionCategory, total int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TotalInCategory[category] = total
}

// Serialize converts the component to JSON bytes.
func (c *CollectionComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *CollectionComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}

// CollectibleComponent marks an entity as a collectible item.
type CollectibleComponent struct {
	mu sync.RWMutex

	// CollectibleID is the unique identifier for this collectible type.
	CollectibleID string `json:"collectible_id"`

	// Name is the display name.
	Name string `json:"name"`

	// Category is the collection category.
	Category CollectionCategory `json:"category"`

	// Rarity is the item rarity.
	Rarity CollectionRarity `json:"rarity"`

	// Description is flavor text.
	Description string `json:"description"`

	// IsCollected indicates if this instance has been collected.
	IsCollected bool `json:"is_collected"`

	// CollectedBy is the entity ID of the collector (0 if not collected).
	CollectedBy uint64 `json:"collected_by"`

	// CollectedAt is the unix timestamp of collection.
	CollectedAt int64 `json:"collected_at"`

	// RequiredLevel is the minimum level to collect (0 = no requirement).
	RequiredLevel int `json:"required_level"`

	// IsHidden indicates collectible requires special conditions to see.
	IsHidden bool `json:"is_hidden"`

	// HintText provides a clue for hidden collectibles.
	HintText string `json:"hint_text"`
}

// NewCollectibleComponent creates a new collectible component.
func NewCollectibleComponent(id, name string, category CollectionCategory, rarity CollectionRarity, description string) *CollectibleComponent {
	return &CollectibleComponent{
		CollectibleID: id,
		Name:          name,
		Category:      category,
		Rarity:        rarity,
		Description:   description,
		IsCollected:   false,
		CollectedBy:   0,
		CollectedAt:   0,
		RequiredLevel: 0,
		IsHidden:      false,
		HintText:      "",
	}
}

// Type returns the component type identifier.
func (c *CollectibleComponent) Type() string {
	return "collectible"
}

// CanCollect checks if the collectible can be collected by a player at the given level.
func (c *CollectibleComponent) CanCollect(playerLevel int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.IsCollected && playerLevel >= c.RequiredLevel
}

// Collect marks this collectible as collected.
func (c *CollectibleComponent) Collect(collectorID uint64, timestamp int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsCollected {
		return false
	}

	c.IsCollected = true
	c.CollectedBy = collectorID
	c.CollectedAt = timestamp
	return true
}

// GetCollectibleID returns the collectible type ID.
func (c *CollectibleComponent) GetCollectibleID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CollectibleID
}

// GetCategory returns the collection category.
func (c *CollectibleComponent) GetCategory() CollectionCategory {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Category
}

// GetRarity returns the rarity.
func (c *CollectibleComponent) GetRarity() CollectionRarity {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Rarity
}

// GetName returns the display name.
func (c *CollectibleComponent) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Name
}

// GetDescription returns the description.
func (c *CollectibleComponent) GetDescription() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Description
}

// IsAlreadyCollected returns whether this instance has been collected.
func (c *CollectibleComponent) IsAlreadyCollected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsCollected
}

// SetHidden marks the collectible as hidden with an optional hint.
func (c *CollectibleComponent) SetHidden(hidden bool, hint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsHidden = hidden
	c.HintText = hint
}

// IsHiddenCollectible returns whether this collectible is hidden.
func (c *CollectibleComponent) IsHiddenCollectible() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsHidden
}

// GetHint returns the hint text for hidden collectibles.
func (c *CollectibleComponent) GetHint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HintText
}

// SetRequiredLevel sets the minimum level to collect.
func (c *CollectibleComponent) SetRequiredLevel(level int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RequiredLevel = level
}

// GetRequiredLevel returns the required level.
func (c *CollectibleComponent) GetRequiredLevel() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RequiredLevel
}

// Serialize converts the component to JSON bytes.
func (c *CollectibleComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *CollectibleComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
