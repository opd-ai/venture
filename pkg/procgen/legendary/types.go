package legendary

import (
	"time"

	"github.com/opd-ai/venture/pkg/world/raids"
)

// PhaseType categorizes quest phase activities.
type PhaseType int

const (
	// PhaseKill requires defeating specific enemies
	PhaseKill PhaseType = iota
	// PhaseCollect requires gathering items
	PhaseCollect
	// PhaseCraft requires creating items
	PhaseCraft
	// PhaseRaid requires clearing a raid encounter
	PhaseRaid
	// PhaseTravel requires visiting specific servers
	PhaseTravel
	// PhaseExplore requires discovering locations
	PhaseExplore
	// PhaseTalk requires NPC interactions
	PhaseTalk
	// PhaseChallenge requires special trials
	PhaseChallenge
)

// String returns the phase type name.
func (p PhaseType) String() string {
	switch p {
	case PhaseKill:
		return "Kill"
	case PhaseCollect:
		return "Collect"
	case PhaseCraft:
		return "Craft"
	case PhaseRaid:
		return "Raid"
	case PhaseTravel:
		return "Travel"
	case PhaseExplore:
		return "Explore"
	case PhaseTalk:
		return "Talk"
	case PhaseChallenge:
		return "Challenge"
	default:
		return "Unknown"
	}
}

// QuestPhase represents a single step in a legendary quest chain.
type QuestPhase struct {
	// PhaseNumber is the sequence position (1-based)
	PhaseNumber int
	// Name is phase title
	Name string
	// Description is phase objective text
	Description string
	// Type categorizes this phase
	Type PhaseType
	// Requirements are what must be completed
	Requirements *PhaseRequirements
	// Completed tracks if this phase is done
	Completed bool
	// CompletedAt tracks completion timestamp
	CompletedAt time.Time
}

// PhaseRequirements specifies what must be done to complete a phase.
type PhaseRequirements struct {
	// Kill requirements
	KillTargets   map[string]int // entity type -> count
	KillBosses    []string       // boss names
	KillCompleted map[string]int // progress tracking

	// Collection requirements
	CollectItems     map[string]int // item type -> count
	CollectCompleted map[string]int // progress tracking

	// Crafting requirements
	CraftItems     []CraftRequirement
	CraftCompleted map[string]bool // item ID -> completed

	// Raid requirements
	RaidEncounters     []*RaidRequirement
	RaidCompleted      map[string]bool // raid ID -> completed
	RaidBossesDefeated map[string]bool // boss ID -> defeated

	// Travel requirements
	ServersToVisit      []string        // server IDs
	ServersVisited      map[string]bool // server ID -> visited
	MinServers          int             // minimum unique servers
	LocationsToDiscover []string        // location names
	LocationsDiscovered map[string]bool // location -> discovered

	// Dialogue requirements
	NPCsToTalk   []string        // NPC IDs
	NPCsTalkedTo map[string]bool // NPC ID -> talked

	// Challenge requirements (special trials)
	Challenges          []*Challenge
	ChallengesCompleted map[string]bool // challenge ID -> completed
}

// CraftRequirement specifies an item that must be crafted.
type CraftRequirement struct {
	// ItemType is the item category
	ItemType string
	// ItemName is specific item
	ItemName string
	// Quantity required
	Quantity int
	// StationQuality minimum (Basic, Standard, Advanced, Master)
	StationQuality string
	// Completed tracks progress
	Completed bool
}

// RaidRequirement specifies a raid that must be cleared.
type RaidRequirement struct {
	// RaidID uniquely identifies the raid
	RaidID string
	// RaidName is human-readable
	RaidName string
	// Tier is difficulty level
	Tier raids.RaidTier
	// BossesToKill lists specific bosses (empty = all)
	BossesToKill []string
	// MinPartySize required
	MinPartySize int
	// MaxDeaths allowed (0 = no limit)
	MaxDeaths int
	// TimeLimit in minutes (0 = no limit)
	TimeLimit int
}

// Challenge represents a special trial requirement.
type Challenge struct {
	// ID uniquely identifies this challenge
	ID string
	// Name is challenge title
	Name string
	// Description explains the challenge
	Description string
	// Type categorizes the challenge
	Type ChallengeType
	// Difficulty scaling
	Difficulty float64
	// Completed tracks success
	Completed bool
}

// ChallengeType categorizes special challenges.
type ChallengeType int

const (
	// ChallengeSurvival requires surviving for duration
	ChallengeSurvival ChallengeType = iota
	// ChallengePuzzle requires solving a puzzle
	ChallengePuzzle
	// ChallengeCombat requires defeating enemies with constraints
	ChallengeCombat
	// ChallengeSpeed requires completing task quickly
	ChallengeSpeed
	// ChallengePerfection requires no mistakes
	ChallengePerfection
)

// String returns the challenge type name.
func (c ChallengeType) String() string {
	switch c {
	case ChallengeSurvival:
		return "Survival"
	case ChallengePuzzle:
		return "Puzzle"
	case ChallengeCombat:
		return "Combat"
	case ChallengeSpeed:
		return "Speed"
	case ChallengePerfection:
		return "Perfection"
	default:
		return "Unknown"
	}
}

// PhaseProgress returns completion status for current phase (0.0-1.0).
func (p *QuestPhase) PhaseProgress() float64 {
	if p.Requirements == nil {
		return 0.0
	}

	req := p.Requirements
	totalTasks := 0
	completedTasks := 0

	totalTasks, completedTasks = countKillProgress(req, totalTasks, completedTasks)
	totalTasks, completedTasks = countCollectionProgress(req, totalTasks, completedTasks)
	totalTasks, completedTasks = countCraftingProgress(req, totalTasks, completedTasks)
	totalTasks, completedTasks = countRaidProgress(req, totalTasks, completedTasks)
	totalTasks, completedTasks = countTravelProgress(req, totalTasks, completedTasks)
	totalTasks, completedTasks = countNPCProgress(req, totalTasks, completedTasks)
	totalTasks, completedTasks = countChallengeProgress(req, totalTasks, completedTasks)

	if totalTasks == 0 {
		return 0.0
	}

	return float64(completedTasks) / float64(totalTasks)
}

// countKillProgress tallies kill target completion.
func countKillProgress(req *PhaseRequirements, total, completed int) (int, int) {
	for target, required := range req.KillTargets {
		total++
		if current := req.KillCompleted[target]; current >= required {
			completed++
		}
	}
	return total, completed
}

// countCollectionProgress tallies item collection completion.
func countCollectionProgress(req *PhaseRequirements, total, completed int) (int, int) {
	for item, required := range req.CollectItems {
		total++
		if current := req.CollectCompleted[item]; current >= required {
			completed++
		}
	}
	return total, completed
}

// countCraftingProgress tallies crafting requirement completion.
func countCraftingProgress(req *PhaseRequirements, total, completed int) (int, int) {
	for _, craft := range req.CraftItems {
		total++
		if craft.Completed {
			completed++
		}
	}
	return total, completed
}

// countRaidProgress tallies raid encounter completion.
func countRaidProgress(req *PhaseRequirements, total, completed int) (int, int) {
	for range req.RaidEncounters {
		total++
	}
	if req.RaidCompleted != nil {
		for _, done := range req.RaidCompleted {
			if done {
				completed++
				break
			}
		}
	}
	return total, completed
}

// countTravelProgress tallies server visit completion.
func countTravelProgress(req *PhaseRequirements, total, completed int) (int, int) {
	if req.MinServers > 0 {
		total++
		if len(req.ServersVisited) >= req.MinServers {
			completed++
		}
	}
	return total, completed
}

// countNPCProgress tallies NPC interaction completion.
func countNPCProgress(req *PhaseRequirements, total, completed int) (int, int) {
	for range req.NPCsToTalk {
		total++
	}
	if req.NPCsTalkedTo != nil {
		for _, talked := range req.NPCsTalkedTo {
			if talked {
				completed++
				break
			}
		}
	}
	return total, completed
}

// countChallengeProgress tallies challenge completion.
func countChallengeProgress(req *PhaseRequirements, total, completed int) (int, int) {
	for _, challenge := range req.Challenges {
		total++
		if challenge.Completed {
			completed++
		}
	}
	return total, completed
}

// NewPhaseRequirements creates an initialized requirements struct.
func NewPhaseRequirements() *PhaseRequirements {
	return &PhaseRequirements{
		KillCompleted:       make(map[string]int),
		CollectCompleted:    make(map[string]int),
		CraftCompleted:      make(map[string]bool),
		RaidCompleted:       make(map[string]bool),
		RaidBossesDefeated:  make(map[string]bool),
		ServersVisited:      make(map[string]bool),
		LocationsDiscovered: make(map[string]bool),
		NPCsTalkedTo:        make(map[string]bool),
		ChallengesCompleted: make(map[string]bool),
	}
}

// LegendaryQuest represents a multi-phase legendary quest.
type LegendaryQuest struct {
	// ID uniquely identifies this legendary quest
	ID string
	// Name is the quest title
	Name string
	// Description provides quest lore
	Description string
	// Phases are sequential quest steps
	Phases []*QuestPhase
	// Rewards are granted upon completion
	Rewards *LegendaryRewards
	// RequiredLevel is minimum level to start
	RequiredLevel int
	// Seed for deterministic generation
	Seed int64
	// EstimatedHours is completion time estimate
	EstimatedHours float64
	// StartedAt tracks when quest was accepted
	StartedAt time.Time
	// CompletedAt tracks when quest was finished
	CompletedAt time.Time
}

// LegendaryRewards contains all rewards for completing the quest.
type LegendaryRewards struct {
	// Items are legendary items granted
	Items []LegendaryItem
	// Titles are account-wide titles
	Titles []string
	// Gold reward amount
	Gold int
	// Experience reward
	Experience int
	// PrestigeLevels granted
	PrestigeLevels int
	// Achievements unlocked
	Achievements []string
	// Cosmetics unlocked (visual effects, mounts, etc.)
	Cosmetics []string
}

// Progress returns overall quest completion (0.0-1.0).
func (q *LegendaryQuest) Progress() float64 {
	if len(q.Phases) == 0 {
		return 0.0
	}

	completed := 0
	for _, phase := range q.Phases {
		if phase.Completed {
			completed++
		}
	}

	return float64(completed) / float64(len(q.Phases))
}

// CurrentPhase returns the active phase (first incomplete).
func (q *LegendaryQuest) CurrentPhase() *QuestPhase {
	for _, phase := range q.Phases {
		if !phase.Completed {
			return phase
		}
	}
	return nil // All phases complete
}

// IsComplete returns true if all phases are done.
func (q *LegendaryQuest) IsComplete() bool {
	for _, phase := range q.Phases {
		if !phase.Completed {
			return false
		}
	}
	return len(q.Phases) > 0
}

// QuestTemplate defines a legendary quest archetype.
type QuestTemplate struct {
	// NamePattern for quest title generation
	NamePattern string
	// MinPhases in quest chain
	MinPhases int
	// MaxPhases in quest chain
	MaxPhases int
	// PhaseTypes allowed
	PhaseTypes []PhaseType
	// MinServers for travel requirements
	MinServers int
	// MaxServers for travel requirements
	MaxServers int
	// RequiresRaid if true, must include raid phase
	RequiresRaid bool
	// RequiresCrafting if true, must include craft phase
	RequiresCrafting bool
	// EstimatedHoursMin for completion
	EstimatedHoursMin float64
	// EstimatedHoursMax for completion
	EstimatedHoursMax float64
	// RewardTier affects reward quality
	RewardTier int
}

// LegendaryItem represents a legendary item reward (simplified for quest rewards).
// Full item generation happens via pkg/procgen/item package.
type LegendaryItem struct {
	Name   string
	Rarity Rarity
}

// Rarity represents item quality tiers.
type Rarity int

const (
	RarityCommon Rarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

// String returns the rarity name.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityEpic:
		return "Epic"
	case RarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
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
	Progress map[string]map[string]*PlayerProgress `json:"progress"` // questID -> playerID -> progress
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		Progress: make(map[string]map[string]*PlayerProgress),
	}
}

// GetProgress retrieves a player's progress on a quest.
func (pt *ProgressTracker) GetProgress(questID, playerID string) *PlayerProgress {
	if questMap, ok := pt.Progress[questID]; ok {
		return questMap[playerID]
	}
	return nil
}

// UpdatePhase updates a player's phase progress.
func (pt *ProgressTracker) UpdatePhase(questID, playerID string, phase int, progress float64) {
	if pt.Progress[questID] == nil {
		pt.Progress[questID] = make(map[string]*PlayerProgress)
	}

	p := pt.Progress[questID][playerID]
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
		pt.Progress[questID][playerID] = p
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
	if pt.Progress[questID] == nil {
		pt.Progress[questID] = make(map[string]*PlayerProgress)
	}

	p := pt.Progress[questID][playerID]
	if p == nil {
		p = &PlayerProgress{
			QuestID:           questID,
			PlayerID:          playerID,
			CurrentPhase:      0,
			PhaseProgress:     0.0,
			ServersVisited:    make([]string, 0),
			RaidsCompleted:    make([]string, 0),
			ItemsCrafted:      make([]string, 0),
			MaterialsGathered: make(map[string]int),
			StartedAt:         time.Now(),
			LastUpdated:       time.Now(),
		}
		pt.Progress[questID][playerID] = p
	}

	for _, s := range p.ServersVisited {
		if s == serverID {
			return
		}
	}
	p.ServersVisited = append(p.ServersVisited, serverID)
	p.LastUpdated = time.Now()
}

// AddRaidCompleted adds a raid to the player's completed list.
func (pt *ProgressTracker) AddRaidCompleted(questID, playerID, raidID string) {
	if pt.Progress[questID] == nil {
		pt.Progress[questID] = make(map[string]*PlayerProgress)
	}

	p := pt.Progress[questID][playerID]
	if p == nil {
		p = &PlayerProgress{
			QuestID:           questID,
			PlayerID:          playerID,
			CurrentPhase:      0,
			PhaseProgress:     0.0,
			ServersVisited:    make([]string, 0),
			RaidsCompleted:    make([]string, 0),
			ItemsCrafted:      make([]string, 0),
			MaterialsGathered: make(map[string]int),
			StartedAt:         time.Now(),
			LastUpdated:       time.Now(),
		}
		pt.Progress[questID][playerID] = p
	}

	for _, r := range p.RaidsCompleted {
		if r == raidID {
			return
		}
	}
	p.RaidsCompleted = append(p.RaidsCompleted, raidID)
	p.LastUpdated = time.Now()
}

// AddMaterial adds or increments material count for a player.
func (pt *ProgressTracker) AddMaterial(questID, playerID, materialID string, quantity int) {
	if pt.Progress[questID] == nil {
		pt.Progress[questID] = make(map[string]*PlayerProgress)
	}

	p := pt.Progress[questID][playerID]
	if p == nil {
		p = &PlayerProgress{
			QuestID:           questID,
			PlayerID:          playerID,
			CurrentPhase:      0,
			PhaseProgress:     0.0,
			ServersVisited:    make([]string, 0),
			RaidsCompleted:    make([]string, 0),
			ItemsCrafted:      make([]string, 0),
			MaterialsGathered: make(map[string]int),
			StartedAt:         time.Now(),
			LastUpdated:       time.Now(),
		}
		pt.Progress[questID][playerID] = p
	}

	p.MaterialsGathered[materialID] += quantity
	p.LastUpdated = time.Now()
}
