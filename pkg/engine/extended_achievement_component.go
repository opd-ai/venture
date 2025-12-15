// Package engine provides extended achievement tracking across all game systems.
// This file implements the ExtendedAchievementComponent and related types
// for comprehensive achievement tracking with tiered progression.
//
// Phase 83: Extended Achievement Categories (V15.0)
package engine

import (
	"encoding/json"
	"sync"
)

// AchievementCategory represents the category of an achievement.
type AchievementCategory int

const (
	// AchievementCategoryCombat tracks combat-related achievements.
	AchievementCategoryCombat AchievementCategory = iota
	// AchievementCategoryQuest tracks quest completion achievements.
	AchievementCategoryQuest
	// AchievementCategoryCrafting tracks crafting achievements.
	AchievementCategoryCrafting
	// AchievementCategoryExploration tracks exploration achievements.
	AchievementCategoryExploration
	// AchievementCategorySocial tracks social interaction achievements.
	AchievementCategorySocial
	// AchievementCategoryPvP tracks PvP-related achievements.
	AchievementCategoryPvP
)

// String returns the string representation of the category.
func (c AchievementCategory) String() string {
	names := []string{"Combat", "Quest", "Crafting", "Exploration", "Social", "PvP"}
	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

// AchievementTier represents the tier level of an achievement.
type AchievementTier int

const (
	// AchievementTierNone indicates no tier unlocked.
	AchievementTierNone AchievementTier = iota
	// AchievementTierBronze is the first tier.
	AchievementTierBronze
	// AchievementTierSilver is the second tier.
	AchievementTierSilver
	// AchievementTierGold is the third tier.
	AchievementTierGold
	// AchievementTierPlatinum is the highest tier.
	AchievementTierPlatinum
)

// String returns the string representation of the tier.
func (t AchievementTier) String() string {
	names := []string{"None", "Bronze", "Silver", "Gold", "Platinum"}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

// Points returns the achievement points for this tier.
func (t AchievementTier) Points() int {
	points := []int{0, 10, 25, 50, 100}
	if int(t) < len(points) {
		return points[t]
	}
	return 0
}

// AchievementDefinition defines an achievement's properties.
type AchievementDefinition struct {
	ID          string
	Name        string
	Description string
	Category    AchievementCategory
	// Thresholds for each tier [Bronze, Silver, Gold, Platinum]
	Thresholds [4]int64
}

// AchievementEntry tracks progress for a single achievement.
type AchievementEntry struct {
	ID          string                    `json:"id"`
	Category    AchievementCategory       `json:"category"`
	CurrentTier AchievementTier           `json:"current_tier"`
	Progress    int64                     `json:"progress"`
	UnlockedAt  map[AchievementTier]int64 `json:"unlocked_at"`
}

// NewAchievementEntry creates a new achievement entry.
func NewAchievementEntry(id string, category AchievementCategory) *AchievementEntry {
	return &AchievementEntry{
		ID:          id,
		Category:    category,
		CurrentTier: AchievementTierNone,
		Progress:    0,
		UnlockedAt:  make(map[AchievementTier]int64),
	}
}

// ExtendedAchievementComponent tracks achievements across all game systems.
type ExtendedAchievementComponent struct {
	mu             sync.RWMutex
	Achievements   map[string]*AchievementEntry `json:"achievements"`
	TotalPoints    int                          `json:"total_points"`
	CategoryPoints map[AchievementCategory]int  `json:"category_points"`
}

// Type returns the component type identifier.
func (c *ExtendedAchievementComponent) Type() string {
	return "extended_achievement"
}

// NewExtendedAchievementComponent creates a new extended achievement component.
func NewExtendedAchievementComponent() *ExtendedAchievementComponent {
	return &ExtendedAchievementComponent{
		Achievements:   make(map[string]*AchievementEntry),
		TotalPoints:    0,
		CategoryPoints: make(map[AchievementCategory]int),
	}
}

// GetAchievement returns an achievement entry by ID, creating if not exists.
func (c *ExtendedAchievementComponent) GetAchievement(id string, category AchievementCategory) *AchievementEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.Achievements[id]; exists {
		return entry
	}

	entry := NewAchievementEntry(id, category)
	c.Achievements[id] = entry
	return entry
}

// GetProgress returns the current progress for an achievement.
func (c *ExtendedAchievementComponent) GetProgress(id string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.Achievements[id]; exists {
		return entry.Progress
	}
	return 0
}

// GetTier returns the current tier for an achievement.
func (c *ExtendedAchievementComponent) GetTier(id string) AchievementTier {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.Achievements[id]; exists {
		return entry.CurrentTier
	}
	return AchievementTierNone
}

// SetProgress updates the progress for an achievement.
// Returns true if a new tier was unlocked.
func (c *ExtendedAchievementComponent) SetProgress(id string, category AchievementCategory, progress int64, thresholds [4]int64, timestamp int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.Achievements[id]
	if !exists {
		entry = NewAchievementEntry(id, category)
		c.Achievements[id] = entry
	}

	entry.Progress = progress

	// Check for tier upgrades
	newTier := c.calculateTier(progress, thresholds)
	if newTier > entry.CurrentTier {
		// Award points for each tier upgraded
		for t := entry.CurrentTier + 1; t <= newTier; t++ {
			points := t.Points()
			c.TotalPoints += points
			c.CategoryPoints[category] += points
			entry.UnlockedAt[t] = timestamp
		}
		entry.CurrentTier = newTier
		return true
	}

	return false
}

// IncrementProgress adds to the progress for an achievement.
// Returns true if a new tier was unlocked.
func (c *ExtendedAchievementComponent) IncrementProgress(id string, category AchievementCategory, amount int64, thresholds [4]int64, timestamp int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.Achievements[id]
	if !exists {
		entry = NewAchievementEntry(id, category)
		c.Achievements[id] = entry
	}

	entry.Progress += amount

	// Check for tier upgrades
	newTier := c.calculateTier(entry.Progress, thresholds)
	if newTier > entry.CurrentTier {
		// Award points for each tier upgraded
		for t := entry.CurrentTier + 1; t <= newTier; t++ {
			points := t.Points()
			c.TotalPoints += points
			c.CategoryPoints[category] += points
			entry.UnlockedAt[t] = timestamp
		}
		entry.CurrentTier = newTier
		return true
	}

	return false
}

// calculateTier determines the tier based on progress and thresholds.
func (c *ExtendedAchievementComponent) calculateTier(progress int64, thresholds [4]int64) AchievementTier {
	if progress >= thresholds[3] {
		return AchievementTierPlatinum
	}
	if progress >= thresholds[2] {
		return AchievementTierGold
	}
	if progress >= thresholds[1] {
		return AchievementTierSilver
	}
	if progress >= thresholds[0] {
		return AchievementTierBronze
	}
	return AchievementTierNone
}

// GetTotalPoints returns the total achievement points earned.
func (c *ExtendedAchievementComponent) GetTotalPoints() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TotalPoints
}

// GetCategoryPoints returns points earned in a specific category.
func (c *ExtendedAchievementComponent) GetCategoryPoints(category AchievementCategory) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CategoryPoints[category]
}

// GetAchievementsByCategory returns all achievements in a category.
func (c *ExtendedAchievementComponent) GetAchievementsByCategory(category AchievementCategory) []*AchievementEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*AchievementEntry
	for _, entry := range c.Achievements {
		if entry.Category == category {
			result = append(result, entry)
		}
	}
	return result
}

// GetUnlockedCount returns the count of achievements with at least Bronze tier.
func (c *ExtendedAchievementComponent) GetUnlockedCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, entry := range c.Achievements {
		if entry.CurrentTier >= AchievementTierBronze {
			count++
		}
	}
	return count
}

// GetMaxTierCount returns the count of achievements at Platinum tier.
func (c *ExtendedAchievementComponent) GetMaxTierCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, entry := range c.Achievements {
		if entry.CurrentTier == AchievementTierPlatinum {
			count++
		}
	}
	return count
}

// Serialize converts the component to JSON bytes for persistence.
func (c *ExtendedAchievementComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return json.Marshal(struct {
		Achievements   map[string]*AchievementEntry `json:"achievements"`
		TotalPoints    int                          `json:"total_points"`
		CategoryPoints map[AchievementCategory]int  `json:"category_points"`
	}{
		Achievements:   c.Achievements,
		TotalPoints:    c.TotalPoints,
		CategoryPoints: c.CategoryPoints,
	})
}

// Deserialize restores the component from JSON bytes.
func (c *ExtendedAchievementComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var temp struct {
		Achievements   map[string]*AchievementEntry `json:"achievements"`
		TotalPoints    int                          `json:"total_points"`
		CategoryPoints map[AchievementCategory]int  `json:"category_points"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	c.Achievements = temp.Achievements
	c.TotalPoints = temp.TotalPoints
	c.CategoryPoints = temp.CategoryPoints

	// Ensure maps are initialized
	if c.Achievements == nil {
		c.Achievements = make(map[string]*AchievementEntry)
	}
	if c.CategoryPoints == nil {
		c.CategoryPoints = make(map[AchievementCategory]int)
	}

	// Ensure UnlockedAt maps are initialized for each entry
	for _, entry := range c.Achievements {
		if entry.UnlockedAt == nil {
			entry.UnlockedAt = make(map[AchievementTier]int64)
		}
	}

	return nil
}
