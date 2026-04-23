// Package engine provides behavior tree nodes for AI decision-making.
// This file re-exports the foundational types from the aitypes and behavior
// sub-packages so that all existing engine code continues to compile without
// modification, while the definitions themselves live in import-cycle-free packages.
package engine

import (
	"github.com/opd-ai/venture/pkg/engine/ai/behavior"
	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

// NodeStatus is a type alias for aitypes.NodeStatus.
// All existing engine code using NodeStatus continues to compile unchanged.
type NodeStatus = aitypes.NodeStatus

// Re-export NodeStatus constants so callers in this package need no new imports.
const (
	// NodeSuccess indicates the node completed successfully.
	NodeSuccess = aitypes.NodeSuccess
	// NodeFailure indicates the node failed to complete.
	NodeFailure = aitypes.NodeFailure
	// NodeRunning indicates the node is still executing.
	NodeRunning = aitypes.NodeRunning
)

// Blackboard is a type alias for aitypes.Blackboard.
// All existing engine code using *Blackboard continues to compile unchanged.
type Blackboard = aitypes.Blackboard

// NewBlackboard creates a new empty blackboard with a default RNG.
func NewBlackboard() *Blackboard {
	return aitypes.NewBlackboard()
}

// NewBlackboardWithSeed creates a new blackboard with a seeded RNG.
func NewBlackboardWithSeed(seed int64) *Blackboard {
	return aitypes.NewBlackboardWithSeed(seed)
}

// GetEntityFromBlackboard retrieves a *Entity stored in a blackboard under key.
// This replaces the former blackboard.GetEntity(key) method, which could not be
// defined on aitypes.Blackboard without creating an import cycle.
func GetEntityFromBlackboard(bb *Blackboard, key string) (*Entity, bool) {
	val, ok := bb.Get(key)
	if !ok {
		return nil, false
	}
	entity, ok := val.(*Entity)
	return entity, ok
}

// Type aliases for the composite nodes defined in pkg/engine/ai/behavior.
// Existing engine code that names these types (e.g. *SequenceNode) continues
// to compile without modification.

// SequenceNode executes children in order until one fails.
type SequenceNode = behavior.SequenceNode

// NewSequenceNode creates a new sequence node.
func NewSequenceNode(name string, children ...BehaviorNode) *SequenceNode {
	return behavior.NewSequenceNode(name, children...)
}

// SelectorNode executes children in order until one succeeds.
type SelectorNode = behavior.SelectorNode

// NewSelectorNode creates a new selector node.
func NewSelectorNode(name string, children ...BehaviorNode) *SelectorNode {
	return behavior.NewSelectorNode(name, children...)
}

// ParallelNode executes all children simultaneously.
type ParallelNode = behavior.ParallelNode

// NewParallelNode creates a new parallel node.
func NewParallelNode(name string, children ...BehaviorNode) *ParallelNode {
	return behavior.NewParallelNode(name, children...)
}

// InverterNode inverts the success/failure of its child.
type InverterNode = behavior.InverterNode

// NewInverterNode creates a new inverter decorator node.
func NewInverterNode(name string, child BehaviorNode) *InverterNode {
	return behavior.NewInverterNode(name, child)
}

// RepeatNode repeats its child a specified number of times or until failure.
type RepeatNode = behavior.RepeatNode

// NewRepeatNode creates a new repeat decorator node.
func NewRepeatNode(name string, maxRepeats int, child BehaviorNode) *RepeatNode {
	return behavior.NewRepeatNode(name, maxRepeats, child)
}
