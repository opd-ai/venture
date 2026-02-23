package narrative_world

import (
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/story"
)

// EventType categorizes significant companion memories
type EventType int

const (
	EventTypeCombat EventType = iota
	EventTypeTreasure
	EventTypeDanger
	EventTypeBonding
	EventTypeConflict
	EventTypeDiscovery
	EventTypeSacrifice
	EventTypeBetray
)

func (e EventType) String() string {
	switch e {
	case EventTypeCombat:
		return "Combat"
	case EventTypeTreasure:
		return "Treasure"
	case EventTypeDanger:
		return "Danger"
	case EventTypeBonding:
		return "Bonding"
	case EventTypeConflict:
		return "Conflict"
	case EventTypeDiscovery:
		return "Discovery"
	case EventTypeSacrifice:
		return "Sacrifice"
	case EventTypeBetray:
		return "Betray"
	default:
		return "Unknown"
	}
}

// MemoryEvent represents a significant event in companion memory
type MemoryEvent struct {
	Timestamp    int64 // Unix timestamp in seconds (deterministic via TimeProvider)
	Type         EventType
	Description  string
	Participants []uint64 // Entity IDs involved
	Location     string   // Where it happened
	Importance   float64  // 0.0-1.0, affects recall probability
}

// PersonalQuest represents a companion-specific quest
type PersonalQuest struct {
	QuestID         string
	CompanionID     uint64
	CompanionType   engine.CompanionType
	Title           string
	Description     string
	Objectives      []QuestObjective
	UnlockLoyalty   float64 // Minimum loyalty to unlock (0.7 default)
	Completed       bool
	Started         bool
	StoryBranches   *story.BranchingNarrative   // Branching quest outcomes
	Consequences    []Consequence               // What happens on completion/failure
	PersonalityReqs []learning.PersonalityTrait // Required personality traits
}

// QuestObjective represents a single quest goal
type QuestObjective struct {
	Description string
	Type        ObjectiveType
	Target      string // Entity type, location, or item
	Progress    int    // Current progress
	Required    int    // Required amount
	Completed   bool
}

// ObjectiveType categorizes quest objectives
type ObjectiveType int

const (
	ObjectiveDefeat ObjectiveType = iota
	ObjectiveCollect
	ObjectiveVisit
	ObjectiveProtect
	ObjectiveTalk
	ObjectiveExplore
)

func (o ObjectiveType) String() string {
	switch o {
	case ObjectiveDefeat:
		return "Defeat"
	case ObjectiveCollect:
		return "Collect"
	case ObjectiveVisit:
		return "Visit"
	case ObjectiveProtect:
		return "Protect"
	case ObjectiveTalk:
		return "Talk"
	case ObjectiveExplore:
		return "Explore"
	default:
		return "Unknown"
	}
}

// Consequence represents quest outcomes
type Consequence struct {
	Type        ConsequenceType
	Description string
	Permanent   bool    // Cannot be reversed
	Severity    float64 // 0.0-1.0, impact level
}

// ConsequenceType categorizes quest consequences
type ConsequenceType int

const (
	ConsequenceLoyaltyChange ConsequenceType = iota
	ConsequenceDeparture
	ConsequenceDeath
	ConsequenceRelationshipChange
	ConsequenceItemGain
	ConsequenceSkillUnlock
)

func (c ConsequenceType) String() string {
	switch c {
	case ConsequenceLoyaltyChange:
		return "Loyalty Change"
	case ConsequenceDeparture:
		return "Departure"
	case ConsequenceDeath:
		return "Death"
	case ConsequenceRelationshipChange:
		return "Relationship Change"
	case ConsequenceItemGain:
		return "Item Gain"
	case ConsequenceSkillUnlock:
		return "Skill Unlock"
	default:
		return "Unknown"
	}
}

// CompanionConflict represents tension between companions
// CompanionConflict represents an active or resolved conflict between two companions.
// Note on time representation:
//   - TimeSinceStart uses time.Duration for in-memory elapsed time tracking
//   - This differs from MemoryEvent.Timestamp which uses Unix seconds (int64)
//   - For serialization, TimeSinceStart is stored as nanoseconds to preserve precision
type CompanionConflict struct {
	Companion1      uint64
	Companion2      uint64
	ConflictType    ConflictType
	Description     string
	Severity        float64        // 0.0-1.0
	ResolutionQuest *PersonalQuest // Optional quest to resolve
	Active          bool
	TimeSinceStart  time.Duration  // Elapsed time since conflict began (updated via deltaTime)
}

// ConflictType categorizes companion conflicts
type ConflictType int

const (
	ConflictPersonality ConflictType = iota
	ConflictRivalry
	ConflictBeliefs
	ConflictPastHistory
	ConflictResourceCompetition
)

func (c ConflictType) String() string {
	switch c {
	case ConflictPersonality:
		return "Personality Clash"
	case ConflictRivalry:
		return "Rivalry"
	case ConflictBeliefs:
		return "Conflicting Beliefs"
	case ConflictPastHistory:
		return "Past History"
	case ConflictResourceCompetition:
		return "Resource Competition"
	default:
		return "Unknown"
	}
}

// CrossCompanionStory represents a narrative involving multiple companions
type CrossCompanionStory struct {
	StoryID      string
	Title        string
	Description  string
	Participants []uint64 // Companion entity IDs
	Events       []MemoryEvent
	Narrative    *story.BranchingNarrative
	Outcome      StoryOutcome
	Active       bool
}

// StoryOutcome represents the result of a cross-companion story
type StoryOutcome int

const (
	OutcomeUnresolved StoryOutcome = iota
	OutcomeFriendship
	OutcomeRomance
	OutcomeRivalry
	OutcomeBetrayal
	OutcomeSacrifice
)

func (s StoryOutcome) String() string {
	switch s {
	case OutcomeUnresolved:
		return "Unresolved"
	case OutcomeFriendship:
		return "Friendship"
	case OutcomeRomance:
		return "Romance"
	case OutcomeRivalry:
		return "Rivalry"
	case OutcomeBetrayal:
		return "Betrayal"
	case OutcomeSacrifice:
		return "Sacrifice"
	default:
		return "Unknown"
	}
}

// CompanionMemory tracks all memories for a companion
type CompanionMemory struct {
	CompanionID uint64
	Events      []MemoryEvent
	MaxEvents   int // Limit to prevent memory bloat (50-100)
	TotalEvents int // Track even after pruning old events
}

// DialogueContext provides context for memory-based dialogue
type DialogueContext struct {
	RecentEvents    []MemoryEvent // Last 10 events
	ImportantEvents []MemoryEvent // High importance events
	RelatedTo       []uint64      // Other entities in context
}
