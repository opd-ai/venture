// Package engine provides the CarryOver component for New Game Plus.
// This file implements CarryOverComponent for tracking and managing
// item/currency/skill selections that carry over between NG+ cycles.
//
// Phase 112: Carry-Over System
package engine

import (
	"encoding/json"
	"sync"
)

// CarryOverCategory represents a category of carry-over items.
type CarryOverCategory string

const (
	// CarryOverCurrency represents gold and other currencies
	CarryOverCurrency CarryOverCategory = "currency"
	// CarryOverEquipment represents weapons, armor, accessories
	CarryOverEquipment CarryOverCategory = "equipment"
	// CarryOverSkills represents learned skills and abilities
	CarryOverSkills CarryOverCategory = "skills"
	// CarryOverCosmetics represents visual customizations (always carried over)
	CarryOverCosmetics CarryOverCategory = "cosmetics"
	// CarryOverAchievements represents unlocked achievements (always carried over)
	CarryOverAchievements CarryOverCategory = "achievements"
)

// CarryOverComponent tracks what items and progress will be carried over to the next NG+ cycle.
// Players can select specific items within slot limits defined by their NG+ level.
type CarryOverComponent struct {
	mu sync.RWMutex

	// SelectedEquipment contains item IDs selected for carry-over
	// Limited by CarryOverSlots from NewGamePlusComponent
	SelectedEquipment []string `json:"selected_equipment"`

	// CurrencyCarryOver tracks how much of each currency type to carry over
	// Values are raw amounts; percentage limit is applied during transfer
	CurrencyCarryOver map[string]int64 `json:"currency_carry_over"`

	// SkillsToKeep contains skill IDs that will be preserved
	// Limited to 50% of total skills at base, +5% per NG+ level
	SkillsToKeep []string `json:"skills_to_keep"`

	// CosmeticsUnlocked contains all unlocked cosmetics (always carried over)
	CosmeticsUnlocked []string `json:"cosmetics_unlocked"`

	// AchievementsUnlocked contains all unlocked achievements (always carried over)
	AchievementsUnlocked []string `json:"achievements_unlocked"`

	// SelectionLocked prevents changes once NG+ transition begins
	SelectionLocked bool `json:"selection_locked"`

	// SelectionConfirmed indicates player has confirmed their selections
	SelectionConfirmed bool `json:"selection_confirmed"`

	// EquipmentSlotLimit is the maximum equipment items that can be carried over
	// Copied from NewGamePlusComponent at selection time
	EquipmentSlotLimit int `json:"equipment_slot_limit"`

	// SkillSlotLimit is the maximum skills that can be carried over
	// Calculated as: min(total_skills * (0.5 + 0.05 * ng_cycle), total_skills)
	SkillSlotLimit int `json:"skill_slot_limit"`

	// CurrencyPercentLimit is the maximum percentage of currency to carry over
	// Copied from NewGamePlusComponent at selection time
	CurrencyPercentLimit float64 `json:"currency_percent_limit"`

	// TransferComplete indicates the carry-over has been applied
	TransferComplete bool `json:"transfer_complete"`
}

// Type returns the component type identifier.
func (c *CarryOverComponent) Type() string {
	return "carryover"
}

// NewCarryOverComponent creates a new carry-over component with default settings.
func NewCarryOverComponent() *CarryOverComponent {
	return &CarryOverComponent{
		SelectedEquipment:    []string{},
		CurrencyCarryOver:    make(map[string]int64),
		SkillsToKeep:         []string{},
		CosmeticsUnlocked:    []string{},
		AchievementsUnlocked: []string{},
		SelectionLocked:      false,
		SelectionConfirmed:   false,
		EquipmentSlotLimit:   3, // Default from base NG+
		SkillSlotLimit:       5, // Default initial limit
		CurrencyPercentLimit: 50.0,
		TransferComplete:     false,
	}
}

// SetLimitsFromNGPlus updates carry-over limits based on NG+ component.
// Should be called when preparing for NG+ transition.
func (c *CarryOverComponent) SetLimitsFromNGPlus(ngp *NewGamePlusComponent, totalSkills int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ngp == nil {
		return
	}

	c.EquipmentSlotLimit = ngp.GetCarryOverSlots()
	c.CurrencyPercentLimit = ngp.GetCurrencyCarryOverPercent()

	// Skill limit: 50% base + 5% per NG+ cycle, max 100%
	cycle := ngp.GetCycle()
	skillPercent := 0.5 + (0.05 * float64(cycle))
	if skillPercent > 1.0 {
		skillPercent = 1.0
	}
	c.SkillSlotLimit = int(float64(totalSkills) * skillPercent)
	if c.SkillSlotLimit < 1 && totalSkills > 0 {
		c.SkillSlotLimit = 1 // Always allow at least 1 skill
	}
}

// IsLocked returns true if selections are locked for transfer.
func (c *CarryOverComponent) IsLocked() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SelectionLocked
}

// Lock prevents further selection changes.
func (c *CarryOverComponent) Lock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SelectionLocked = true
}

// Unlock allows selection changes (for cancel/restart scenarios).
func (c *CarryOverComponent) Unlock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SelectionLocked = false
	c.SelectionConfirmed = false
}

// CanSelectEquipment checks if more equipment can be selected.
func (c *CarryOverComponent) CanSelectEquipment() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.SelectionLocked {
		return false
	}
	return len(c.SelectedEquipment) < c.EquipmentSlotLimit
}

// SelectEquipment adds an equipment item ID to carry-over selection.
// Returns true if successful, false if locked or at limit.
func (c *CarryOverComponent) SelectEquipment(itemID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return false
	}
	if len(c.SelectedEquipment) >= c.EquipmentSlotLimit {
		return false
	}

	// Check for duplicates
	for _, id := range c.SelectedEquipment {
		if id == itemID {
			return false // Already selected
		}
	}

	c.SelectedEquipment = append(c.SelectedEquipment, itemID)
	return true
}

// DeselectEquipment removes an equipment item from carry-over selection.
// Returns true if the item was found and removed.
func (c *CarryOverComponent) DeselectEquipment(itemID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return false
	}

	for i, id := range c.SelectedEquipment {
		if id == itemID {
			c.SelectedEquipment = append(c.SelectedEquipment[:i], c.SelectedEquipment[i+1:]...)
			return true
		}
	}
	return false
}

// IsEquipmentSelected checks if an item is in the carry-over selection.
func (c *CarryOverComponent) IsEquipmentSelected(itemID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, id := range c.SelectedEquipment {
		if id == itemID {
			return true
		}
	}
	return false
}

// GetSelectedEquipment returns a copy of selected equipment IDs.
func (c *CarryOverComponent) GetSelectedEquipment() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.SelectedEquipment))
	copy(result, c.SelectedEquipment)
	return result
}

// GetEquipmentSelectionCount returns the number of selected equipment items.
func (c *CarryOverComponent) GetEquipmentSelectionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.SelectedEquipment)
}

// GetEquipmentSlotLimit returns the maximum equipment slots available.
func (c *CarryOverComponent) GetEquipmentSlotLimit() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.EquipmentSlotLimit
}

// SetCurrencyCarryOver sets the amount of a currency type to carry over.
// The actual transferred amount will be capped by CurrencyPercentLimit.
func (c *CarryOverComponent) SetCurrencyCarryOver(currencyType string, amount int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return
	}
	if c.CurrencyCarryOver == nil {
		c.CurrencyCarryOver = make(map[string]int64)
	}
	c.CurrencyCarryOver[currencyType] = amount
}

// GetCurrencyCarryOver returns the amount of currency to carry over.
func (c *CarryOverComponent) GetCurrencyCarryOver(currencyType string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CurrencyCarryOver[currencyType]
}

// GetAllCurrencyCarryOver returns a copy of all currency carry-over amounts.
func (c *CarryOverComponent) GetAllCurrencyCarryOver() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.CurrencyCarryOver))
	for k, v := range c.CurrencyCarryOver {
		result[k] = v
	}
	return result
}

// GetCurrencyPercentLimit returns the maximum percentage of currency that can be carried over.
func (c *CarryOverComponent) GetCurrencyPercentLimit() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CurrencyPercentLimit
}

// CalculateFinalCurrencyAmount returns the actual amount that will be transferred after applying the limit.
func (c *CarryOverComponent) CalculateFinalCurrencyAmount(currencyType string, totalAmount int64) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	maxAllowed := int64(float64(totalAmount) * c.CurrencyPercentLimit / 100.0)
	requested := c.CurrencyCarryOver[currencyType]

	if requested > maxAllowed {
		return maxAllowed
	}
	return requested
}

// CanSelectSkill checks if more skills can be selected.
func (c *CarryOverComponent) CanSelectSkill() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.SelectionLocked {
		return false
	}
	return len(c.SkillsToKeep) < c.SkillSlotLimit
}

// SelectSkill adds a skill ID to carry-over selection.
// Returns true if successful, false if locked or at limit.
func (c *CarryOverComponent) SelectSkill(skillID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return false
	}
	if len(c.SkillsToKeep) >= c.SkillSlotLimit {
		return false
	}

	// Check for duplicates
	for _, id := range c.SkillsToKeep {
		if id == skillID {
			return false
		}
	}

	c.SkillsToKeep = append(c.SkillsToKeep, skillID)
	return true
}

// DeselectSkill removes a skill from carry-over selection.
func (c *CarryOverComponent) DeselectSkill(skillID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return false
	}

	for i, id := range c.SkillsToKeep {
		if id == skillID {
			c.SkillsToKeep = append(c.SkillsToKeep[:i], c.SkillsToKeep[i+1:]...)
			return true
		}
	}
	return false
}

// IsSkillSelected checks if a skill is in the carry-over selection.
func (c *CarryOverComponent) IsSkillSelected(skillID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, id := range c.SkillsToKeep {
		if id == skillID {
			return true
		}
	}
	return false
}

// GetSelectedSkills returns a copy of selected skill IDs.
func (c *CarryOverComponent) GetSelectedSkills() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.SkillsToKeep))
	copy(result, c.SkillsToKeep)
	return result
}

// GetSkillSelectionCount returns the number of selected skills.
func (c *CarryOverComponent) GetSkillSelectionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.SkillsToKeep)
}

// GetSkillSlotLimit returns the maximum skill slots available.
func (c *CarryOverComponent) GetSkillSlotLimit() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SkillSlotLimit
}

// AddCosmetic registers an unlocked cosmetic (always carries over).
func (c *CarryOverComponent) AddCosmetic(cosmeticID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicates
	for _, id := range c.CosmeticsUnlocked {
		if id == cosmeticID {
			return false
		}
	}

	c.CosmeticsUnlocked = append(c.CosmeticsUnlocked, cosmeticID)
	return true
}

// HasCosmetic checks if a cosmetic is unlocked.
func (c *CarryOverComponent) HasCosmetic(cosmeticID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, id := range c.CosmeticsUnlocked {
		if id == cosmeticID {
			return true
		}
	}
	return false
}

// GetCosmetics returns a copy of all unlocked cosmetics.
func (c *CarryOverComponent) GetCosmetics() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.CosmeticsUnlocked))
	copy(result, c.CosmeticsUnlocked)
	return result
}

// AddAchievement registers an unlocked achievement (always carries over).
func (c *CarryOverComponent) AddAchievement(achievementID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicates
	for _, id := range c.AchievementsUnlocked {
		if id == achievementID {
			return false
		}
	}

	c.AchievementsUnlocked = append(c.AchievementsUnlocked, achievementID)
	return true
}

// HasAchievement checks if an achievement is unlocked.
func (c *CarryOverComponent) HasAchievement(achievementID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, id := range c.AchievementsUnlocked {
		if id == achievementID {
			return true
		}
	}
	return false
}

// GetAchievements returns a copy of all unlocked achievements.
func (c *CarryOverComponent) GetAchievements() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.AchievementsUnlocked))
	copy(result, c.AchievementsUnlocked)
	return result
}

// ConfirmSelection marks the selection as confirmed by the player.
// Returns false if already locked or no selections made.
func (c *CarryOverComponent) ConfirmSelection() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return false
	}

	c.SelectionConfirmed = true
	return true
}

// IsConfirmed returns true if the player has confirmed selections.
func (c *CarryOverComponent) IsConfirmed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SelectionConfirmed
}

// IsTransferComplete returns true if carry-over has been applied.
func (c *CarryOverComponent) IsTransferComplete() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TransferComplete
}

// MarkTransferComplete marks the carry-over as applied.
func (c *CarryOverComponent) MarkTransferComplete() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TransferComplete = true
}

// ClearSelections resets all selections (except cosmetics and achievements).
// Used when starting fresh or canceling selection.
func (c *CarryOverComponent) ClearSelections() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SelectionLocked {
		return
	}

	c.SelectedEquipment = []string{}
	c.CurrencyCarryOver = make(map[string]int64)
	c.SkillsToKeep = []string{}
	c.SelectionConfirmed = false
}

// Reset fully resets the component for a new cycle.
// Preserves cosmetics and achievements, clears everything else.
func (c *CarryOverComponent) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.SelectedEquipment = []string{}
	c.CurrencyCarryOver = make(map[string]int64)
	c.SkillsToKeep = []string{}
	c.SelectionLocked = false
	c.SelectionConfirmed = false
	c.TransferComplete = false
	// Note: CosmeticsUnlocked and AchievementsUnlocked are preserved
}

// GetSummary returns a summary of current carry-over selections.
func (c *CarryOverComponent) GetSummary() CarryOverSummary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CarryOverSummary{
		EquipmentCount:   len(c.SelectedEquipment),
		EquipmentLimit:   c.EquipmentSlotLimit,
		SkillCount:       len(c.SkillsToKeep),
		SkillLimit:       c.SkillSlotLimit,
		CurrencyTypes:    len(c.CurrencyCarryOver),
		CurrencyLimit:    c.CurrencyPercentLimit,
		CosmeticCount:    len(c.CosmeticsUnlocked),
		AchievementCount: len(c.AchievementsUnlocked),
		IsLocked:         c.SelectionLocked,
		IsConfirmed:      c.SelectionConfirmed,
		IsComplete:       c.TransferComplete,
	}
}

// CarryOverSummary provides a snapshot of carry-over state.
type CarryOverSummary struct {
	EquipmentCount   int     `json:"equipment_count"`
	EquipmentLimit   int     `json:"equipment_limit"`
	SkillCount       int     `json:"skill_count"`
	SkillLimit       int     `json:"skill_limit"`
	CurrencyTypes    int     `json:"currency_types"`
	CurrencyLimit    float64 `json:"currency_limit"`
	CosmeticCount    int     `json:"cosmetic_count"`
	AchievementCount int     `json:"achievement_count"`
	IsLocked         bool    `json:"is_locked"`
	IsConfirmed      bool    `json:"is_confirmed"`
	IsComplete       bool    `json:"is_complete"`
}

// Serialize converts the component to JSON for persistence.
func (c *CarryOverComponent) Serialize() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// Deserialize restores the component from JSON data.
func (c *CarryOverComponent) Deserialize(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var temp CarryOverComponent
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	c.SelectedEquipment = temp.SelectedEquipment
	c.CurrencyCarryOver = temp.CurrencyCarryOver
	c.SkillsToKeep = temp.SkillsToKeep
	c.CosmeticsUnlocked = temp.CosmeticsUnlocked
	c.AchievementsUnlocked = temp.AchievementsUnlocked
	c.SelectionLocked = temp.SelectionLocked
	c.SelectionConfirmed = temp.SelectionConfirmed
	c.EquipmentSlotLimit = temp.EquipmentSlotLimit
	c.SkillSlotLimit = temp.SkillSlotLimit
	c.CurrencyPercentLimit = temp.CurrencyPercentLimit
	c.TransferComplete = temp.TransferComplete

	// Initialize nil maps/slices
	if c.SelectedEquipment == nil {
		c.SelectedEquipment = []string{}
	}
	if c.CurrencyCarryOver == nil {
		c.CurrencyCarryOver = make(map[string]int64)
	}
	if c.SkillsToKeep == nil {
		c.SkillsToKeep = []string{}
	}
	if c.CosmeticsUnlocked == nil {
		c.CosmeticsUnlocked = []string{}
	}
	if c.AchievementsUnlocked == nil {
		c.AchievementsUnlocked = []string{}
	}

	return nil
}
