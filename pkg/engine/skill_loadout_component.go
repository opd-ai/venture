// Package engine provides skill loadout functionality for character customization.
// This file implements SkillLoadoutComponent which stores multiple skill configurations
// that players can save and swap between for different gameplay situations.
package engine

import (
	"encoding/json"
	"fmt"
)

// SkillLoadout represents a saved skill point configuration.
// Players can save their current skill setup and restore it later.
type SkillLoadout struct {
	Name        string         `json:"name"`         // User-defined loadout name
	Description string         `json:"description"`  // Optional description
	SkillLevels map[string]int `json:"skill_levels"` // Skill ID -> level mapping
	TreeID      string         `json:"tree_id"`      // Associated skill tree ID
	CreatedAt   float64        `json:"created_at"`   // Timestamp when created
	LastUsedAt  float64        `json:"last_used_at"` // Timestamp of last use
}

// Copy creates a deep copy of the loadout.
func (l *SkillLoadout) Copy() *SkillLoadout {
	skillLevels := make(map[string]int, len(l.SkillLevels))
	for k, v := range l.SkillLevels {
		skillLevels[k] = v
	}
	return &SkillLoadout{
		Name:        l.Name,
		Description: l.Description,
		SkillLevels: skillLevels,
		TreeID:      l.TreeID,
		CreatedAt:   l.CreatedAt,
		LastUsedAt:  l.LastUsedAt,
	}
}

// TotalPointsUsed calculates total skill points invested in this loadout.
// Note: This is an approximation; actual cost depends on skill requirements.
func (l *SkillLoadout) TotalPointsUsed() int {
	total := 0
	for _, level := range l.SkillLevels {
		total += level
	}
	return total
}

// SkillLoadoutComponent stores multiple saved skill configurations.
// Players can save up to MaxLoadouts configurations and swap between them.
type SkillLoadoutComponent struct {
	Loadouts       []*SkillLoadout `json:"loadouts"`       // Saved loadout configurations
	ActiveIndex    int             `json:"active_index"`   // Currently active loadout (-1 = none)
	MaxLoadouts    int             `json:"max_loadouts"`   // Maximum allowed loadouts
	SwapCooldown   float64         `json:"swap_cooldown"`  // Cooldown between loadout swaps (seconds)
	LastSwapTime   float64         `json:"last_swap_time"` // Timestamp of last swap
	SwapCost       int             `json:"swap_cost"`      // Gold cost per swap (0 = free)
	QuickSlots     [3]int          `json:"quick_slots"`    // Indices of loadouts assigned to quick swap (1-3 keys)
	UnlockedSlots  int             `json:"unlocked_slots"` // Number of loadout slots unlocked (starts at 2)
	IsDirty        bool            `json:"-"`              // Flag for UI refresh
	PendingRestore int             `json:"-"`              // Index of loadout pending restoration (-1 = none)
}

// Type returns the component type identifier.
func (s *SkillLoadoutComponent) Type() string {
	return "skill_loadout"
}

// NewSkillLoadoutComponent creates a new skill loadout component with default settings.
func NewSkillLoadoutComponent() *SkillLoadoutComponent {
	return &SkillLoadoutComponent{
		Loadouts:       make([]*SkillLoadout, 0),
		ActiveIndex:    -1,
		MaxLoadouts:    10,
		SwapCooldown:   30.0, // 30 seconds between swaps by default
		LastSwapTime:   0,
		SwapCost:       0, // Free swaps by default
		QuickSlots:     [3]int{-1, -1, -1},
		UnlockedSlots:  2, // Start with 2 loadout slots
		IsDirty:        false,
		PendingRestore: -1,
	}
}

// SaveLoadout creates a new loadout from current skill configuration.
// Returns the index of the saved loadout, or -1 if saving failed.
func (s *SkillLoadoutComponent) SaveLoadout(name, description string, skillLevels map[string]int, treeID string, currentTime float64) int {
	if len(s.Loadouts) >= s.UnlockedSlots {
		return -1 // No available slots
	}

	// Create deep copy of skill levels
	levels := make(map[string]int, len(skillLevels))
	for k, v := range skillLevels {
		if v > 0 {
			levels[k] = v
		}
	}

	loadout := &SkillLoadout{
		Name:        name,
		Description: description,
		SkillLevels: levels,
		TreeID:      treeID,
		CreatedAt:   currentTime,
		LastUsedAt:  currentTime,
	}

	s.Loadouts = append(s.Loadouts, loadout)
	s.IsDirty = true
	return len(s.Loadouts) - 1
}

// UpdateLoadout overwrites an existing loadout with current skill configuration.
// Returns true if successful.
func (s *SkillLoadoutComponent) UpdateLoadout(index int, skillLevels map[string]int, currentTime float64) bool {
	if index < 0 || index >= len(s.Loadouts) {
		return false
	}

	// Create deep copy of skill levels
	levels := make(map[string]int, len(skillLevels))
	for k, v := range skillLevels {
		if v > 0 {
			levels[k] = v
		}
	}

	s.Loadouts[index].SkillLevels = levels
	s.Loadouts[index].LastUsedAt = currentTime
	s.IsDirty = true
	return true
}

// RenameLoadout changes the name and description of a loadout.
func (s *SkillLoadoutComponent) RenameLoadout(index int, name, description string) bool {
	if index < 0 || index >= len(s.Loadouts) {
		return false
	}
	s.Loadouts[index].Name = name
	s.Loadouts[index].Description = description
	s.IsDirty = true
	return true
}

// DeleteLoadout removes a loadout at the specified index.
func (s *SkillLoadoutComponent) DeleteLoadout(index int) bool {
	if index < 0 || index >= len(s.Loadouts) {
		return false
	}

	// Update quick slots that reference this or higher indices
	for i := range s.QuickSlots {
		if s.QuickSlots[i] == index {
			s.QuickSlots[i] = -1
		} else if s.QuickSlots[i] > index {
			s.QuickSlots[i]--
		}
	}

	// Update active index
	if s.ActiveIndex == index {
		s.ActiveIndex = -1
	} else if s.ActiveIndex > index {
		s.ActiveIndex--
	}

	// Remove loadout
	s.Loadouts = append(s.Loadouts[:index], s.Loadouts[index+1:]...)
	s.IsDirty = true
	return true
}

// GetLoadout retrieves a loadout by index. Returns nil if not found.
func (s *SkillLoadoutComponent) GetLoadout(index int) *SkillLoadout {
	if index < 0 || index >= len(s.Loadouts) {
		return nil
	}
	return s.Loadouts[index]
}

// GetLoadoutByName finds a loadout by name. Returns index and loadout, or -1/nil if not found.
func (s *SkillLoadoutComponent) GetLoadoutByName(name string) (int, *SkillLoadout) {
	for i, loadout := range s.Loadouts {
		if loadout.Name == name {
			return i, loadout
		}
	}
	return -1, nil
}

// CanSwapLoadout checks if a loadout swap is currently allowed.
// Returns true if cooldown has passed.
func (s *SkillLoadoutComponent) CanSwapLoadout(currentTime float64) bool {
	return currentTime >= s.LastSwapTime+s.SwapCooldown
}

// GetSwapCooldownRemaining returns seconds remaining until swap is available.
func (s *SkillLoadoutComponent) GetSwapCooldownRemaining(currentTime float64) float64 {
	remaining := (s.LastSwapTime + s.SwapCooldown) - currentTime
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RequestLoadoutRestore queues a loadout for restoration.
// Actual restoration happens in the system's Update to allow validation.
func (s *SkillLoadoutComponent) RequestLoadoutRestore(index int, currentTime float64) bool {
	if index < 0 || index >= len(s.Loadouts) {
		return false
	}
	if !s.CanSwapLoadout(currentTime) {
		return false
	}
	s.PendingRestore = index
	return true
}

// MarkSwapComplete updates timestamps after a successful loadout swap.
func (s *SkillLoadoutComponent) MarkSwapComplete(index int, currentTime float64) {
	s.LastSwapTime = currentTime
	s.ActiveIndex = index
	s.PendingRestore = -1
	if index >= 0 && index < len(s.Loadouts) {
		s.Loadouts[index].LastUsedAt = currentTime
	}
	s.IsDirty = true
}

// AssignQuickSlot assigns a loadout to a quick slot (0-2 for keys F1-F3).
func (s *SkillLoadoutComponent) AssignQuickSlot(slot, loadoutIndex int) bool {
	if slot < 0 || slot >= 3 {
		return false
	}
	if loadoutIndex != -1 && (loadoutIndex < 0 || loadoutIndex >= len(s.Loadouts)) {
		return false
	}
	s.QuickSlots[slot] = loadoutIndex
	s.IsDirty = true
	return true
}

// GetQuickSlotLoadout returns the loadout assigned to a quick slot.
func (s *SkillLoadoutComponent) GetQuickSlotLoadout(slot int) *SkillLoadout {
	if slot < 0 || slot >= 3 {
		return nil
	}
	return s.GetLoadout(s.QuickSlots[slot])
}

// UnlockSlot increases the number of available loadout slots.
// Returns true if a slot was unlocked.
func (s *SkillLoadoutComponent) UnlockSlot() bool {
	if s.UnlockedSlots >= s.MaxLoadouts {
		return false
	}
	s.UnlockedSlots++
	s.IsDirty = true
	return true
}

// GetLoadoutCount returns the number of saved loadouts.
func (s *SkillLoadoutComponent) GetLoadoutCount() int {
	return len(s.Loadouts)
}

// GetAvailableSlots returns the number of unused loadout slots.
func (s *SkillLoadoutComponent) GetAvailableSlots() int {
	return s.UnlockedSlots - len(s.Loadouts)
}

// GetLoadoutNames returns a slice of all loadout names for UI display.
func (s *SkillLoadoutComponent) GetLoadoutNames() []string {
	names := make([]string, len(s.Loadouts))
	for i, loadout := range s.Loadouts {
		names[i] = loadout.Name
	}
	return names
}

// IsActiveLoadout checks if a loadout index is the currently active one.
func (s *SkillLoadoutComponent) IsActiveLoadout(index int) bool {
	return s.ActiveIndex == index
}

// ClearDirtyFlag resets the dirty flag after UI refresh.
func (s *SkillLoadoutComponent) ClearDirtyFlag() {
	s.IsDirty = false
}

// Serialize encodes the component to JSON for persistence.
func (s *SkillLoadoutComponent) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize decodes the component from JSON.
func (s *SkillLoadoutComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// CalculateSkillDifference returns skills that differ between loadout and current config.
// Returns maps of skills to add and skills to remove.
func (s *SkillLoadoutComponent) CalculateSkillDifference(loadoutIndex int, currentSkillLevels map[string]int) (toAdd, toRemove map[string]int) {
	toAdd = make(map[string]int)
	toRemove = make(map[string]int)

	loadout := s.GetLoadout(loadoutIndex)
	if loadout == nil {
		return toAdd, toRemove
	}

	// Find skills to add/increase
	for skillID, targetLevel := range loadout.SkillLevels {
		currentLevel := currentSkillLevels[skillID]
		if targetLevel > currentLevel {
			toAdd[skillID] = targetLevel - currentLevel
		}
	}

	// Find skills to remove/decrease
	for skillID, currentLevel := range currentSkillLevels {
		targetLevel := loadout.SkillLevels[skillID]
		if currentLevel > targetLevel {
			toRemove[skillID] = currentLevel - targetLevel
		}
	}

	return toAdd, toRemove
}

// ValidateLoadoutCompatibility checks if a loadout can be applied to a skill tree.
// Returns error describing incompatibility, or nil if compatible.
func (s *SkillLoadoutComponent) ValidateLoadoutCompatibility(loadoutIndex int, treeID string) error {
	loadout := s.GetLoadout(loadoutIndex)
	if loadout == nil {
		return fmt.Errorf("loadout not found at index %d", loadoutIndex)
	}
	if loadout.TreeID != "" && loadout.TreeID != treeID {
		return fmt.Errorf("loadout was created for tree %s, not %s", loadout.TreeID, treeID)
	}
	return nil
}
