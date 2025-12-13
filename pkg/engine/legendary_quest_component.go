package engine

import "errors"

// LegendaryQuestComponent tracks a player's legendary quest progress.
type LegendaryQuestComponent struct {
	QuestID         string
	CurrentPhase    int
	PhasesCompleted []bool
	StartedAt       float64
}

// Type returns the component type.
func (c *LegendaryQuestComponent) Type() string {
	return "legendary_quest"
}

// LegendaryItemComponent represents a legendary reward item.
type LegendaryItemComponent struct {
	RewardID    string
	Name        string
	Description string
	Stats       map[string]int
	Rarity      float64
	Unique      bool
}

// Type returns the component type.
func (c *LegendaryItemComponent) Type() string {
	return "legendary_item"
}

// Errors for legendary quest system.
var (
	ErrNoActiveQuest = errors.New("no active legendary quest")
)
