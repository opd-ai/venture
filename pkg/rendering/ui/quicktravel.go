// Package ui provides quick-travel and enhanced tooltip systems for Phase 60.1.
package ui

import (
	"fmt"
	"math"
	"sync"
)

// TravelDestination represents a quick-travel location
type TravelDestination struct {
	ID          string
	Name        string
	Description string
	X, Y        float64
	Cost        int // gold cost
	Unlocked    bool
	Category    string // "House", "Guild Hall", "City", "Dungeon", "Landmark"
}

// QuickTravelManager manages fast travel destinations
type QuickTravelManager struct {
	mu           sync.RWMutex
	destinations map[string]*TravelDestination
	playerGold   int
}

// NewQuickTravelManager creates a new quick-travel manager
func NewQuickTravelManager() *QuickTravelManager {
	return &QuickTravelManager{
		destinations: make(map[string]*TravelDestination),
	}
}

// RegisterDestination adds a new travel destination
func (qtm *QuickTravelManager) RegisterDestination(dest *TravelDestination) {
	qtm.mu.Lock()
	defer qtm.mu.Unlock()
	qtm.destinations[dest.ID] = dest
}

// UnlockDestination makes a destination available for travel
func (qtm *QuickTravelManager) UnlockDestination(id string) error {
	qtm.mu.Lock()
	defer qtm.mu.Unlock()

	dest, exists := qtm.destinations[id]
	if !exists {
		return fmt.Errorf("destination not found: %s", id)
	}

	dest.Unlocked = true
	return nil
}

// CalculateCost computes travel cost based on distance
func (qtm *QuickTravelManager) CalculateCost(fromX, fromY float64, destID string) (int, error) {
	qtm.mu.RLock()
	defer qtm.mu.RUnlock()

	dest, exists := qtm.destinations[destID]
	if !exists {
		return 0, fmt.Errorf("destination not found: %s", destID)
	}

	// Calculate Euclidean distance
	dx := dest.X - fromX
	dy := dest.Y - fromY
	distance := math.Sqrt(dx*dx + dy*dy)

	// Base cost + distance multiplier (10 gold per 100 units)
	baseCost := 100
	distanceCost := int(distance / 100 * 10)

	// Clamp between 100 and 1000 gold
	totalCost := baseCost + distanceCost
	if totalCost < 100 {
		totalCost = 100
	}
	if totalCost > 1000 {
		totalCost = 1000
	}

	return totalCost, nil
}

// CanTravel checks if player can afford travel
func (qtm *QuickTravelManager) CanTravel(fromX, fromY float64, destID string, playerGold int) (bool, error) {
	cost, err := qtm.CalculateCost(fromX, fromY, destID)
	if err != nil {
		return false, err
	}

	qtm.mu.RLock()
	dest := qtm.destinations[destID]
	qtm.mu.RUnlock()

	if dest == nil {
		return false, fmt.Errorf("destination not found: %s", destID)
	}

	if !dest.Unlocked {
		return false, fmt.Errorf("destination %s is locked", destID)
	}

	return playerGold >= cost, nil
}

// Travel performs quick travel to a destination
func (qtm *QuickTravelManager) Travel(fromX, fromY float64, destID string, playerGold *int) (*TravelDestination, int, error) {
	cost, err := qtm.CalculateCost(fromX, fromY, destID)
	if err != nil {
		return nil, 0, err
	}

	qtm.mu.RLock()
	defer qtm.mu.RUnlock()

	dest, exists := qtm.destinations[destID]
	if !exists {
		return nil, 0, fmt.Errorf("destination not found: %s", destID)
	}

	if !dest.Unlocked {
		return nil, 0, fmt.Errorf("destination %s is locked", destID)
	}

	if *playerGold < cost {
		return nil, cost, fmt.Errorf("insufficient gold: need %d, have %d", cost, *playerGold)
	}

	// Deduct cost
	*playerGold -= cost

	return dest, cost, nil
}

// ListUnlocked returns all unlocked destinations
func (qtm *QuickTravelManager) ListUnlocked() []*TravelDestination {
	qtm.mu.RLock()
	defer qtm.mu.RUnlock()

	result := make([]*TravelDestination, 0)
	for _, dest := range qtm.destinations {
		if dest.Unlocked {
			result = append(result, dest)
		}
	}
	return result
}

// ListByCategory returns destinations in a category
func (qtm *QuickTravelManager) ListByCategory(category string) []*TravelDestination {
	qtm.mu.RLock()
	defer qtm.mu.RUnlock()

	result := make([]*TravelDestination, 0)
	for _, dest := range qtm.destinations {
		if dest.Category == category {
			result = append(result, dest)
		}
	}
	return result
}

// Tooltip represents an enhanced UI tooltip
type Tooltip struct {
	Title        string
	Description  []string               // multiple lines
	Stats        map[string]interface{} // key-value stats
	Bonuses      []string               // bonus descriptions
	Requirements []string               // requirements
	Cost         int                    // item cost
	Rarity       string                 // item rarity
}

// TooltipBuilder constructs tooltips with integration bonuses
type TooltipBuilder struct {
	tooltip *Tooltip
}

// NewTooltipBuilder creates a new tooltip builder
func NewTooltipBuilder(title string) *TooltipBuilder {
	return &TooltipBuilder{
		tooltip: &Tooltip{
			Title:        title,
			Description:  make([]string, 0),
			Stats:        make(map[string]interface{}),
			Bonuses:      make([]string, 0),
			Requirements: make([]string, 0),
		},
	}
}

// AddDescription adds a description line
func (tb *TooltipBuilder) AddDescription(desc string) *TooltipBuilder {
	tb.tooltip.Description = append(tb.tooltip.Description, desc)
	return tb
}

// AddStat adds a stat to the tooltip
func (tb *TooltipBuilder) AddStat(name string, value interface{}) *TooltipBuilder {
	tb.tooltip.Stats[name] = value
	return tb
}

// AddBonus adds a bonus description
func (tb *TooltipBuilder) AddBonus(bonus string) *TooltipBuilder {
	tb.tooltip.Bonuses = append(tb.tooltip.Bonuses, bonus)
	return tb
}

// AddRequirement adds a requirement
func (tb *TooltipBuilder) AddRequirement(req string) *TooltipBuilder {
	tb.tooltip.Requirements = append(tb.tooltip.Requirements, req)
	return tb
}

// SetCost sets the item cost
func (tb *TooltipBuilder) SetCost(cost int) *TooltipBuilder {
	tb.tooltip.Cost = cost
	return tb
}

// SetRarity sets the item rarity
func (tb *TooltipBuilder) SetRarity(rarity string) *TooltipBuilder {
	tb.tooltip.Rarity = rarity
	return tb
}

// Build returns the constructed tooltip
func (tb *TooltipBuilder) Build() *Tooltip {
	return tb.tooltip
}

// FormatTooltip generates a formatted tooltip string
func FormatTooltip(t *Tooltip) string {
	result := fmt.Sprintf("=== %s ===\n", t.Title)

	if t.Rarity != "" {
		result += fmt.Sprintf("Rarity: %s\n", t.Rarity)
	}

	if len(t.Description) > 0 {
		result += "\n"
		for _, desc := range t.Description {
			result += fmt.Sprintf("%s\n", desc)
		}
	}

	if len(t.Stats) > 0 {
		result += "\nStats:\n"
		for name, value := range t.Stats {
			result += fmt.Sprintf("  %s: %v\n", name, value)
		}
	}

	if len(t.Bonuses) > 0 {
		result += "\nBonuses:\n"
		for _, bonus := range t.Bonuses {
			result += fmt.Sprintf("  + %s\n", bonus)
		}
	}

	if len(t.Requirements) > 0 {
		result += "\nRequirements:\n"
		for _, req := range t.Requirements {
			result += fmt.Sprintf("  - %s\n", req)
		}
	}

	if t.Cost > 0 {
		result += fmt.Sprintf("\nCost: %d gold\n", t.Cost)
	}

	return result
}

// CreateItemTooltip creates a tooltip for an item with integration bonuses
func CreateItemTooltip(name, rarity string, damage, defense int, craftingBonus float64) *Tooltip {
	tb := NewTooltipBuilder(name)
	tb.SetRarity(rarity)
	tb.AddDescription("A procedurally generated item")

	if damage > 0 {
		tb.AddStat("Damage", damage)
	}
	if defense > 0 {
		tb.AddStat("Defense", defense)
	}

	// Add integration bonuses
	if craftingBonus > 1.0 {
		bonusPercent := int((craftingBonus - 1.0) * 100)
		tb.AddBonus(fmt.Sprintf("Crafted in quality station (+%d%% stats)", bonusPercent))
	}

	return tb.Build()
}

// CreateStationTooltip creates a tooltip for a crafting station
func CreateStationTooltip(name string, quality, recipeCount int) *Tooltip {
	qualityName := "Basic"
	if quality == 2 {
		qualityName = "Standard"
	} else if quality == 3 {
		qualityName = "Advanced"
	} else if quality == 4 {
		qualityName = "Master"
	}

	tb := NewTooltipBuilder(name)
	tb.AddDescription(fmt.Sprintf("A %s quality crafting station", qualityName))
	tb.AddStat("Quality", qualityName)
	tb.AddStat("Unlocked Recipes", recipeCount)

	multiplier := 1.0 + float64(quality-1)*0.3 // 1.0, 1.3, 1.6, 1.9
	bonusPercent := int((multiplier - 1.0) * 100)
	if bonusPercent > 0 {
		tb.AddBonus(fmt.Sprintf("+%d%% crafting bonus", bonusPercent))
	}

	return tb.Build()
}

// CreateCompanionTooltip creates a tooltip for a companion
func CreateCompanionTooltip(name string, level, loyalty int, skills []string) *Tooltip {
	tb := NewTooltipBuilder(name)
	tb.AddDescription("Your loyal companion")
	tb.AddStat("Level", level)
	tb.AddStat("Loyalty", fmt.Sprintf("%d%%", loyalty))

	if len(skills) > 0 {
		tb.AddDescription("\nSkills:")
		for _, skill := range skills {
			tb.AddDescription(fmt.Sprintf("  - %s", skill))
		}
	}

	if loyalty >= 70 {
		tb.AddBonus("High loyalty: +10% combat effectiveness")
	}

	return tb.Build()
}
