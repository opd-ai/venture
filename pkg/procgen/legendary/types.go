package legendary

import (
	"time"

	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/world/raids"
)

// PhaseType represents the type of objective in a quest phase.
type PhaseType int

const (
	PhaseExploration PhaseType = iota
	PhaseCombat
	PhaseCrafting
	PhaseCollection
	PhaseRaid
	PhaseFinal
)

// String returns the human-readable name of the phase type.
func (p PhaseType) String() string {
	switch p {
	case PhaseExploration:
		return "Exploration"
	case PhaseCombat:
		return "Combat"
	case PhaseCrafting:
		return "Crafting"
	case PhaseCollection:
		return "Collection"
	case PhaseRaid:
		return "Raid"
	case PhaseFinal:
		return "Final"
	default:
		return "Unknown"
	}
}

// RewardType represents the type of legendary reward.
type RewardType int

const (
	RewardItem RewardType = iota
	RewardTitle
	RewardMount
	RewardCompanion
	RewardAchievement
	RewardAccountBonus
)

// String returns the human-readable name of the reward type.
func (r RewardType) String() string {
	switch r {
	case RewardItem:
		return "Item"
	case RewardTitle:
		return "Title"
	case RewardMount:
		return "Mount"
	case RewardCompanion:
		return "Companion"
	case RewardAchievement:
		return "Achievement"
	case RewardAccountBonus:
		return "AccountBonus"
	default:
		return "Unknown"
	}
}

// QuestPhase represents a single phase of a legendary quest.
type QuestPhase struct {
	ID          string
	Type        PhaseType
	Name        string
	Description string
	ServerID    string
	LocationX   int
	LocationY   int

	// Combat objectives
	BossName   string
	EntityType entity.EntityType
	KillCount  int

	// Raid objectives
	RaidID    string
	RaidTier  raids.RaidTier
	BossIndex int

	// Crafting objectives
	RecipeID    string
	ItemName    string
	Quantity    int
	StationTier int

	// Collection objectives
	MaterialIDs []string
	Quantities  []int

	// Reward for completing this phase
	XPReward    int
	GoldReward  int
	ItemRewards []string
}

// LegendaryReward represents a unique reward for completing a legendary quest.
type LegendaryReward struct {
	ID             string
	Name           string
	Description    string
	Type           RewardType
	ItemID         string
	Title          string
	MountID        string
	CompanionID    string
	AchievementID  string
	AccountBonusID string
	BonusPercent   float64
	IsUnique       bool
}

// LegendaryQuest represents a complete legendary quest chain.
type LegendaryQuest struct {
	ID              string
	Name            string
	Description     string
	Lore            string
	MinLevel        int
	EstimatedHours  int
	ServersRequired int
	RaidsRequired   []string
	Phases          []*QuestPhase
	Rewards         []*LegendaryReward
	CreatedAt       time.Time
}

// PlayerProgress represents a player's progress on a legendary quest.
type PlayerProgress struct {
	QuestID           string
	PlayerID          string
	CurrentPhase      int
	PhaseProgress     float64
	ServersVisited    []string
	RaidsCompleted    []string
	ItemsCrafted      []string
	MaterialsGathered map[string]int
	StartedAt         time.Time
	LastUpdated       time.Time
	CompletedAt       *time.Time
	IsCompleted       bool
}

// ProgressTracker tracks player progress across all legendary quests.
type ProgressTracker struct {
	progress map[string]map[string]*PlayerProgress // questID -> playerID -> progress
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		progress: make(map[string]map[string]*PlayerProgress),
	}
}

// GetProgress retrieves a player's progress on a quest.
func (pt *ProgressTracker) GetProgress(questID, playerID string) *PlayerProgress {
	if questMap, ok := pt.progress[questID]; ok {
		return questMap[playerID]
	}
	return nil
}

// UpdatePhase updates a player's phase progress.
func (pt *ProgressTracker) UpdatePhase(questID, playerID string, phase int, progress float64) {
	if pt.progress[questID] == nil {
		pt.progress[questID] = make(map[string]*PlayerProgress)
	}

	p := pt.progress[questID][playerID]
	if p == nil {
		p = &PlayerProgress{
			QuestID:           questID,
			PlayerID:          playerID,
			CurrentPhase:      phase,
			PhaseProgress:     progress,
			ServersVisited:    make([]string, 0),
			RaidsCompleted:    make([]string, 0),
			ItemsCrafted:      make([]string, 0),
			MaterialsGathered: make(map[string]int),
			StartedAt:         time.Now(),
			LastUpdated:       time.Now(),
		}
		pt.progress[questID][playerID] = p
	}

	p.CurrentPhase = phase
	p.PhaseProgress = progress
	p.LastUpdated = time.Now()
}

// CompleteQuest marks a quest as completed for a player.
func (pt *ProgressTracker) CompleteQuest(questID, playerID string) {
	if p := pt.GetProgress(questID, playerID); p != nil {
		now := time.Now()
		p.CompletedAt = &now
		p.IsCompleted = true
		p.LastUpdated = now
	}
}

// AddServerVisited adds a server to the player's visited list.
func (pt *ProgressTracker) AddServerVisited(questID, playerID, serverID string) {
	if p := pt.GetProgress(questID, playerID); p != nil {
		for _, s := range p.ServersVisited {
			if s == serverID {
				return
			}
		}
		p.ServersVisited = append(p.ServersVisited, serverID)
		p.LastUpdated = time.Now()
	}
}

// AddRaidCompleted adds a raid to the player's completed list.
func (pt *ProgressTracker) AddRaidCompleted(questID, playerID, raidID string) {
	if p := pt.GetProgress(questID, playerID); p != nil {
		for _, r := range p.RaidsCompleted {
			if r == raidID {
				return
			}
		}
		p.RaidsCompleted = append(p.RaidsCompleted, raidID)
		p.LastUpdated = time.Now()
	}
}

// AddMaterial adds or increments material count for a player.
func (pt *ProgressTracker) AddMaterial(questID, playerID, materialID string, quantity int) {
	if p := pt.GetProgress(questID, playerID); p != nil {
		p.MaterialsGathered[materialID] += quantity
		p.LastUpdated = time.Now()
	}
}
