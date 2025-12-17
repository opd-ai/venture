// Package engine provides the New Game Plus component for ECS.
// This file implements NewGamePlusComponent for tracking NG+ cycles,
// legacy statistics, and cross-playthrough progression.
//
// Phase 111: NG+ Core Component & Persistence
package engine

import (
	"encoding/json"
	"sync"
	"time"
)

// NewGamePlusComponent tracks NG+ progression across playthroughs.
// It maintains cycle count, cumulative statistics, and carry-over unlocks.
type NewGamePlusComponent struct {
	mu sync.RWMutex

	// Cycle is the current NG+ cycle (0 = first playthrough, 1 = NG+1, etc.)
	Cycle int `json:"cycle"`

	// MaxCycleReached is the highest NG+ cycle ever achieved
	MaxCycleReached int `json:"max_cycle_reached"`

	// LegacyStats accumulates statistics across all playthroughs
	LegacyStats map[string]int64 `json:"legacy_stats"`

	// TotalPlaytime is cumulative playtime in seconds across all cycles
	TotalPlaytime int64 `json:"total_playtime"`

	// CycleStartTime is the Unix timestamp when current cycle started
	CycleStartTime int64 `json:"cycle_start_time"`

	// CurrentCyclePlaytime is playtime in current cycle (seconds)
	CurrentCyclePlaytime int64 `json:"current_cycle_playtime"`

	// CarryOverSlots is the number of equipment slots unlocked for carry-over
	// Base is 3, increases by 1 per NG+ level up to max of 10
	CarryOverSlots int `json:"carry_over_slots"`

	// UnlockedBonuses lists permanent bonuses earned across playthroughs
	UnlockedBonuses []string `json:"unlocked_bonuses"`

	// CurrencyCarryOverPercent is the percentage of currency that carries over
	// Base 50%, +5% per NG+ level, max 100%
	CurrencyCarryOverPercent float64 `json:"currency_carry_over_percent"`

	// CompletedCycles tracks when each cycle was completed
	CompletedCycles []CycleRecord `json:"completed_cycles"`
}

// CycleRecord stores information about a completed playthrough cycle.
type CycleRecord struct {
	CycleNumber   int   `json:"cycle_number"`
	CompletedAt   int64 `json:"completed_at"`
	PlaytimeHours int   `json:"playtime_hours"`
	EnemiesKilled int64 `json:"enemies_killed"`
	QuestsCleared int   `json:"quests_cleared"`
	DeathCount    int   `json:"death_count"`
}

// Type returns the component type identifier.
func (n *NewGamePlusComponent) Type() string {
	return "newgameplus"
}

// NewNewGamePlusComponent creates a new NG+ component for first playthrough.
func NewNewGamePlusComponent() *NewGamePlusComponent {
	return &NewGamePlusComponent{
		Cycle:                    0,
		MaxCycleReached:          0,
		LegacyStats:              make(map[string]int64),
		TotalPlaytime:            0,
		CycleStartTime:           time.Now().Unix(),
		CurrentCyclePlaytime:     0,
		CarryOverSlots:           3, // Base carry-over slots
		UnlockedBonuses:          []string{},
		CurrencyCarryOverPercent: 50.0, // Base 50%
		CompletedCycles:          []CycleRecord{},
	}
}

// GetCycle returns the current NG+ cycle (thread-safe).
func (n *NewGamePlusComponent) GetCycle() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Cycle
}

// IsNewGamePlus returns true if this is an NG+ playthrough.
func (n *NewGamePlusComponent) IsNewGamePlus() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Cycle > 0
}

// GetLegacyStat returns a cumulative stat value across all playthroughs.
func (n *NewGamePlusComponent) GetLegacyStat(statName string) int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.LegacyStats[statName]
}

// AddToLegacyStat increments a legacy stat by the given amount.
func (n *NewGamePlusComponent) AddToLegacyStat(statName string, amount int64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.LegacyStats == nil {
		n.LegacyStats = make(map[string]int64)
	}
	n.LegacyStats[statName] += amount
}

// GetCarryOverSlots returns the number of equipment carry-over slots.
func (n *NewGamePlusComponent) GetCarryOverSlots() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.CarryOverSlots
}

// GetCurrencyCarryOverPercent returns the currency carry-over percentage.
func (n *NewGamePlusComponent) GetCurrencyCarryOverPercent() float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.CurrencyCarryOverPercent
}

// UpdatePlaytime adds elapsed time to current and total playtime.
func (n *NewGamePlusComponent) UpdatePlaytime(deltaSeconds int64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.CurrentCyclePlaytime += deltaSeconds
	n.TotalPlaytime += deltaSeconds
}

// GetTotalPlaytime returns cumulative playtime across all cycles.
func (n *NewGamePlusComponent) GetTotalPlaytime() int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.TotalPlaytime
}

// GetCurrentCyclePlaytime returns playtime in the current cycle.
func (n *NewGamePlusComponent) GetCurrentCyclePlaytime() int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.CurrentCyclePlaytime
}

// HasBonus checks if a permanent bonus is unlocked.
func (n *NewGamePlusComponent) HasBonus(bonusID string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, b := range n.UnlockedBonuses {
		if b == bonusID {
			return true
		}
	}
	return false
}

// UnlockBonus adds a permanent bonus if not already unlocked.
func (n *NewGamePlusComponent) UnlockBonus(bonusID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, b := range n.UnlockedBonuses {
		if b == bonusID {
			return false // Already unlocked
		}
	}
	n.UnlockedBonuses = append(n.UnlockedBonuses, bonusID)
	return true
}

// StartNewCycle initiates a new NG+ cycle, preserving legacy data.
// This should be called after completing the game to begin NG+.
// statsSnapshot contains current cycle stats to record.
func (n *NewGamePlusComponent) StartNewCycle(statsSnapshot map[string]int64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Record the completed cycle
	cycleRecord := CycleRecord{
		CycleNumber:   n.Cycle,
		CompletedAt:   time.Now().Unix(),
		PlaytimeHours: int(n.CurrentCyclePlaytime / 3600),
		EnemiesKilled: statsSnapshot["enemies_killed"],
		QuestsCleared: int(statsSnapshot["quests_completed"]),
		DeathCount:    int(statsSnapshot["deaths"]),
	}
	n.CompletedCycles = append(n.CompletedCycles, cycleRecord)

	// Accumulate stats into legacy
	for statName, value := range statsSnapshot {
		if n.LegacyStats == nil {
			n.LegacyStats = make(map[string]int64)
		}
		n.LegacyStats[statName] += value
	}

	// Increment cycle
	n.Cycle++
	if n.Cycle > n.MaxCycleReached {
		n.MaxCycleReached = n.Cycle
	}

	// Reset cycle-specific tracking
	n.CycleStartTime = time.Now().Unix()
	n.CurrentCyclePlaytime = 0

	// Upgrade carry-over benefits (capped)
	if n.CarryOverSlots+1 <= 10 {
		n.CarryOverSlots = n.CarryOverSlots + 1
	} else {
		n.CarryOverSlots = 10
	}
	if n.CurrencyCarryOverPercent+5.0 <= 100.0 {
		n.CurrencyCarryOverPercent = n.CurrencyCarryOverPercent + 5.0
	} else {
		n.CurrencyCarryOverPercent = 100.0
	}
}

// GetCompletedCycles returns all completed cycle records.
func (n *NewGamePlusComponent) GetCompletedCycles() []CycleRecord {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]CycleRecord, len(n.CompletedCycles))
	copy(result, n.CompletedCycles)
	return result
}

// GetMaxCycleReached returns the highest NG+ level ever reached.
func (n *NewGamePlusComponent) GetMaxCycleReached() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.MaxCycleReached
}

// Serialize converts the component to JSON for persistence.
func (n *NewGamePlusComponent) Serialize() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return json.Marshal(n)
}

// Deserialize restores the component from JSON data.
func (n *NewGamePlusComponent) Deserialize(data []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var temp NewGamePlusComponent
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	n.Cycle = temp.Cycle
	n.MaxCycleReached = temp.MaxCycleReached
	n.LegacyStats = temp.LegacyStats
	n.TotalPlaytime = temp.TotalPlaytime
	n.CycleStartTime = temp.CycleStartTime
	n.CurrentCyclePlaytime = temp.CurrentCyclePlaytime
	n.CarryOverSlots = temp.CarryOverSlots
	n.UnlockedBonuses = temp.UnlockedBonuses
	n.CurrencyCarryOverPercent = temp.CurrencyCarryOverPercent
	n.CompletedCycles = temp.CompletedCycles

	if n.LegacyStats == nil {
		n.LegacyStats = make(map[string]int64)
	}
	if n.UnlockedBonuses == nil {
		n.UnlockedBonuses = []string{}
	}
	if n.CompletedCycles == nil {
		n.CompletedCycles = []CycleRecord{}
	}

	return nil
}

// GetNGPlusLabel returns a display label for the current NG+ level.
// Returns "" for first playthrough, "NG+" for first NG+, "NG+2" for second, etc.
func (n *NewGamePlusComponent) GetNGPlusLabel() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.Cycle == 0 {
		return ""
	}
	if n.Cycle == 1 {
		return "NG+"
	}
	return "NG+" + string('0'+byte(n.Cycle%10)) // Simple for cycles 2-9
}

// GetNGPlusLabelFull returns a full display label including high cycles.
func (n *NewGamePlusComponent) GetNGPlusLabelFull() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.Cycle == 0 {
		return "First Playthrough"
	}
	if n.Cycle == 1 {
		return "New Game Plus"
	}
	return "New Game Plus " + formatCycleNumber(n.Cycle)
}

// formatCycleNumber formats a cycle number for display.
func formatCycleNumber(cycle int) string {
	if cycle <= 9 {
		return string('0' + byte(cycle))
	}
	// For cycles 10+, use numeric format
	result := ""
	for cycle > 0 {
		result = string('0'+byte(cycle%10)) + result
		cycle /= 10
	}
	return result
}
