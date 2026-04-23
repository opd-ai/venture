// Package behavior provides the core behavior tree composite and decorator nodes
// for AI decision-making. These nodes contain no game-engine-specific logic; they
// delegate all entity-touching work to leaf nodes (ActionNode, ConditionNode) that
// live in pkg/engine. This separation breaks the circular dependency that would
// arise if behavior-tree infrastructure imported engine types directly.
package behavior

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

// BehaviorNode is the interface that all behavior tree nodes must implement.
type BehaviorNode interface {
	// Tick executes the node logic and returns the result status.
	Tick(entity aitypes.EntityContext, blackboard *aitypes.Blackboard, deltaTime float64) aitypes.NodeStatus
	// Reset resets the node state for fresh execution.
	Reset()
	// String returns a string representation of the node for debugging.
	String() string
}

// SequenceNode executes children in order until one fails.
// Returns Success if all children succeed, Failure if any fails, Running if any is running.
type SequenceNode struct {
	children      []BehaviorNode
	currentIndex  int
	name          string
	resetOnFinish bool
}

// NewSequenceNode creates a new sequence node.
func NewSequenceNode(name string, children ...BehaviorNode) *SequenceNode {
	return &SequenceNode{
		children:      children,
		currentIndex:  0,
		name:          name,
		resetOnFinish: true,
	}
}

// Tick executes the sequence node logic.
func (s *SequenceNode) Tick(entity aitypes.EntityContext, blackboard *aitypes.Blackboard, deltaTime float64) aitypes.NodeStatus {
	for s.currentIndex < len(s.children) {
		status := s.children[s.currentIndex].Tick(entity, blackboard, deltaTime)

		switch status {
		case aitypes.NodeFailure:
			if s.resetOnFinish {
				s.Reset()
			}
			return aitypes.NodeFailure
		case aitypes.NodeRunning:
			return aitypes.NodeRunning
		case aitypes.NodeSuccess:
			s.currentIndex++
			continue
		}
	}

	if s.resetOnFinish {
		s.Reset()
	}
	return aitypes.NodeSuccess
}

// Reset resets the sequence to start from the beginning.
func (s *SequenceNode) Reset() {
	s.currentIndex = 0
	for _, child := range s.children {
		child.Reset()
	}
}

// String returns a string representation of the node.
func (s *SequenceNode) String() string {
	return fmt.Sprintf("Sequence(%s, %d children)", s.name, len(s.children))
}

// SelectorNode executes children in order until one succeeds.
// Returns Success if any child succeeds, Failure if all fail, Running if any is running.
type SelectorNode struct {
	children      []BehaviorNode
	currentIndex  int
	name          string
	resetOnFinish bool
}

// NewSelectorNode creates a new selector node.
func NewSelectorNode(name string, children ...BehaviorNode) *SelectorNode {
	return &SelectorNode{
		children:      children,
		currentIndex:  0,
		name:          name,
		resetOnFinish: true,
	}
}

// Tick executes the selector node logic.
func (s *SelectorNode) Tick(entity aitypes.EntityContext, blackboard *aitypes.Blackboard, deltaTime float64) aitypes.NodeStatus {
	for s.currentIndex < len(s.children) {
		status := s.children[s.currentIndex].Tick(entity, blackboard, deltaTime)

		switch status {
		case aitypes.NodeSuccess:
			if s.resetOnFinish {
				s.Reset()
			}
			return aitypes.NodeSuccess
		case aitypes.NodeRunning:
			return aitypes.NodeRunning
		case aitypes.NodeFailure:
			s.currentIndex++
			continue
		}
	}

	if s.resetOnFinish {
		s.Reset()
	}
	return aitypes.NodeFailure
}

// Reset resets the selector to start from the beginning.
func (s *SelectorNode) Reset() {
	s.currentIndex = 0
	for _, child := range s.children {
		child.Reset()
	}
}

// String returns a string representation of the node.
func (s *SelectorNode) String() string {
	return fmt.Sprintf("Selector(%s, %d children)", s.name, len(s.children))
}

// ParallelNode executes all children simultaneously.
// Returns Success if all succeed, Failure if any fails, Running if any is running.
type ParallelNode struct {
	children []BehaviorNode
	name     string
}

// NewParallelNode creates a new parallel node.
func NewParallelNode(name string, children ...BehaviorNode) *ParallelNode {
	return &ParallelNode{
		children: children,
		name:     name,
	}
}

// Tick executes all children in parallel.
func (p *ParallelNode) Tick(entity aitypes.EntityContext, blackboard *aitypes.Blackboard, deltaTime float64) aitypes.NodeStatus {
	hasRunning := false
	hasFailure := false

	for _, child := range p.children {
		status := child.Tick(entity, blackboard, deltaTime)

		switch status {
		case aitypes.NodeRunning:
			hasRunning = true
		case aitypes.NodeFailure:
			hasFailure = true
		}
	}

	if hasFailure {
		return aitypes.NodeFailure
	}
	if hasRunning {
		return aitypes.NodeRunning
	}
	return aitypes.NodeSuccess
}

// Reset resets all children.
func (p *ParallelNode) Reset() {
	for _, child := range p.children {
		child.Reset()
	}
}

// String returns a string representation of the node.
func (p *ParallelNode) String() string {
	return fmt.Sprintf("Parallel(%s, %d children)", p.name, len(p.children))
}

// InverterNode inverts the success/failure of its child.
// Success becomes Failure, Failure becomes Success, Running stays Running.
type InverterNode struct {
	child BehaviorNode
	name  string
}

// NewInverterNode creates a new inverter decorator node.
func NewInverterNode(name string, child BehaviorNode) *InverterNode {
	return &InverterNode{
		child: child,
		name:  name,
	}
}

// Tick executes the child and inverts the result.
func (i *InverterNode) Tick(entity aitypes.EntityContext, blackboard *aitypes.Blackboard, deltaTime float64) aitypes.NodeStatus {
	status := i.child.Tick(entity, blackboard, deltaTime)

	switch status {
	case aitypes.NodeSuccess:
		return aitypes.NodeFailure
	case aitypes.NodeFailure:
		return aitypes.NodeSuccess
	default:
		return status
	}
}

// Reset resets the child node.
func (i *InverterNode) Reset() {
	i.child.Reset()
}

// String returns a string representation of the node.
func (i *InverterNode) String() string {
	return fmt.Sprintf("Inverter(%s)", i.name)
}

// RepeatNode repeats its child a specified number of times or until failure.
type RepeatNode struct {
	child         BehaviorNode
	maxRepeats    int // -1 for infinite
	currentRepeat int
	name          string
}

// NewRepeatNode creates a new repeat decorator node.
func NewRepeatNode(name string, maxRepeats int, child BehaviorNode) *RepeatNode {
	return &RepeatNode{
		child:         child,
		maxRepeats:    maxRepeats,
		currentRepeat: 0,
		name:          name,
	}
}

// Tick executes the child repeatedly.
func (r *RepeatNode) Tick(entity aitypes.EntityContext, blackboard *aitypes.Blackboard, deltaTime float64) aitypes.NodeStatus {
	if r.maxRepeats > 0 && r.currentRepeat >= r.maxRepeats {
		r.Reset()
		return aitypes.NodeSuccess
	}

	status := r.child.Tick(entity, blackboard, deltaTime)

	switch status {
	case aitypes.NodeSuccess:
		r.currentRepeat++
		r.child.Reset()
		if r.maxRepeats > 0 && r.currentRepeat >= r.maxRepeats {
			r.Reset()
			return aitypes.NodeSuccess
		}
		return aitypes.NodeRunning
	case aitypes.NodeFailure:
		r.Reset()
		return aitypes.NodeFailure
	default:
		return status
	}
}

// Reset resets the repeat counter.
func (r *RepeatNode) Reset() {
	r.currentRepeat = 0
	r.child.Reset()
}

// String returns a string representation of the node.
func (r *RepeatNode) String() string {
	return fmt.Sprintf("Repeat(%s, max=%d)", r.name, r.maxRepeats)
}
