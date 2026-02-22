// Package branching provides branching narrative type definitions.
// This file defines narrative data structures and components.
// Enum types moved to: enums.go
package branching

import (
	"time"
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

// Consequence represents a delayed or cascading effect triggered by player choices
// in branching narratives. Consequences enable cause-and-effect storytelling where
// player decisions earlier in the narrative can trigger meaningful changes later.
//
// Each consequence has trigger conditions that must be satisfied (e.g., player chose
// a particular dialogue option, visited a specific location, or has a certain
// alignment). When conditions are met, the consequence's effects are applied and
// any affected quests and NPCs are updated accordingly.
//
// Example consequence: "Spare the thief" choice at Node A triggers a consequence
// at Node C where the thief returns to help the player, affecting the "Retrieve
// Artifact" quest and the "ThiefGuild" NPC faction.
type Consequence struct {
	ID                string
	Description       string
	TriggerConditions map[string]interface{} // Conditions that trigger this consequence
	Effects           map[string]interface{} // Effects applied when triggered
	QuestImpact       []string               // Quest IDs affected by this consequence
	NPCImpact         []string               // NPC IDs affected by this consequence
}

// StoryGraph represents the complete narrative structure for a game world,
// containing all story arcs and their interconnected consequences. The graph
// serves as a container for procedurally generated narratives that can span
// multiple arcs with shared consequences.
//
// Story arcs are self-contained narrative sequences with their own node graphs,
// while consequences can cross arc boundaries to create a cohesive world where
// player choices in one arc affect outcomes in another.
//
// The graph is typically populated by the Generator and managed by the Manager,
// which handles player progress tracking and consequence triggering.
type StoryGraph struct {
	Arcs         map[string]*StoryArc
	Consequences map[string]*Consequence
}

// NarrativeComponent is an ECS component that tracks a player entity's narrative
// state including active story arcs, progress through each arc, and consequences
// that have been triggered. This component enables the branching narrative system
// to operate within the ECS architecture.
//
// ActiveArcs lists the IDs of story arcs the player is currently participating in.
// Progress maps arc IDs to the player's PlayerProgress for each arc, tracking
// visited nodes, choices made, alignment shifts, and completion status.
// TriggeredConsequences records consequence IDs that have already fired, preventing
// duplicate triggering.
type NarrativeComponent struct {
	ActiveArcs            []string
	Progress              map[string]*PlayerProgress
	TriggeredConsequences []string
}

func (c NarrativeComponent) Type() string {
	return "narrative"
}
