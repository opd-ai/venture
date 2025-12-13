// Package engine provides ECS components for branching narrative support.
package engine

import (
	"github.com/opd-ai/venture/pkg/narrative/branching"
)

// BranchingNarrativeComponent tracks player progress through a branching story arc.
// It stores the current story arc, player progress, and active story nodes.
type BranchingNarrativeComponent struct {
	ArcID          string                    // Current story arc ID
	Progress       *branching.PlayerProgress // Player's progress through the story
	ActiveArc      *branching.StoryArc       // The active story arc
	Manager        *branching.Manager        // Narrative manager for choice processing
	PendingChoices []branching.Choice        // Choices available at current node
	LastUpdate     float64                   // Time since last narrative update
}

// Type returns the component type identifier.
func (b *BranchingNarrativeComponent) Type() string {
	return "branching_narrative"
}
