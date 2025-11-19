package branching

import (
	"time"
)

// NodeType represents the type of narrative node
type NodeType int

const (
	NodeTypeStart NodeType = iota
	NodeTypeChoice
	NodeTypeEvent
	NodeTypeConsequence
	NodeTypeEnding
)

func (n NodeType) String() string {
	switch n {
	case NodeTypeStart:
		return "Start"
	case NodeTypeChoice:
		return "Choice"
	case NodeTypeEvent:
		return "Event"
	case NodeTypeConsequence:
		return "Consequence"
	case NodeTypeEnding:
		return "Ending"
	default:
		return "Unknown"
	}
}

// EndingType represents the type of story ending
type EndingType int

const (
	EndingTypeHeroic EndingType = iota
	EndingTypeTragic
	EndingTypeNeutral
	EndingTypeMystery
	EndingTypeTriumph
	EndingTypeBetrayal
)

func (e EndingType) String() string {
	switch e {
	case EndingTypeHeroic:
		return "Heroic"
	case EndingTypeTragic:
		return "Tragic"
	case EndingTypeNeutral:
		return "Neutral"
	case EndingTypeMystery:
		return "Mystery"
	case EndingTypeTriumph:
		return "Triumph"
	case EndingTypeBetrayal:
		return "Betrayal"
	default:
		return "Unknown"
	}
}

// AlignmentAxis represents moral alignment axes
type AlignmentAxis string

const (
	AlignmentGoodEvil      AlignmentAxis = "good_evil"
	AlignmentLawChaos      AlignmentAxis = "law_chaos"
	AlignmentHonorDishonor AlignmentAxis = "honor_dishonor"
)

// Choice represents a player choice in the narrative
type Choice struct {
	ID             string
	Text           string
	Requirements   map[string]interface{}    // Requirements to make this choice (e.g., "gold": 100)
	AlignmentShift map[AlignmentAxis]float64 // Effect on moral alignment (-1.0 to 1.0)
	FactionChange  map[string]float64        // Effect on faction reputation
	NextNodeID     string
}

// StoryNode represents a node in the narrative graph
type StoryNode struct {
	ID           string
	Type         NodeType
	Title        string
	Description  string
	Choices      []Choice
	NextNodeID   string                 // For linear nodes (events, consequences)
	Requirements map[string]interface{} // Requirements to reach this node
	Effects      map[string]interface{} // Effects applied when reaching this node
}

// StoryArc represents a complete narrative arc with multiple possible paths
type StoryArc struct {
	ID          string
	Title       string
	Description string
	GenreID     string
	StartNodeID string
	Nodes       map[string]*StoryNode
	Endings     map[string]EndingType
	Seed        int64
}

// PlayerProgress tracks a player's progress through a story arc
type PlayerProgress struct {
	ArcID         string
	CurrentNodeID string
	VisitedNodes  []string
	ChoicesMade   map[string]string         // NodeID -> ChoiceID
	Alignment     map[AlignmentAxis]float64 // Current moral alignment
	Faction       map[string]float64        // Current faction reputation
	Variables     map[string]interface{}    // Story-specific variables
	StartTime     time.Time
	LastUpdate    time.Time
	Completed     bool
	EndingReached string
}

// Consequence represents an effect of player choices
type Consequence struct {
	ID                string
	Description       string
	TriggerConditions map[string]interface{} // Conditions that trigger this consequence
	Effects           map[string]interface{} // Effects applied when triggered
	QuestImpact       []string               // Quest IDs affected by this consequence
	NPCImpact         []string               // NPC IDs affected by this consequence
}

// StoryGraph represents the complete narrative structure
type StoryGraph struct {
	Arcs         map[string]*StoryArc
	Consequences map[string]*Consequence
}

// Component for ECS integration
type NarrativeComponent struct {
	ActiveArcs            []string
	Progress              map[string]*PlayerProgress
	TriggeredConsequences []string
}

func (c NarrativeComponent) Type() string {
	return "narrative"
}
