// Package engine provides behavior tree nodes for AI decision-making.
// This file defines the node types and interfaces for composing behavior trees.
package engine

import (
	"fmt"
	"math/rand"
)

// NodeStatus represents the result of a behavior tree node execution.
type NodeStatus int

const (
	// NodeSuccess indicates the node completed successfully.
	NodeSuccess NodeStatus = iota
	// NodeFailure indicates the node failed to complete.
	NodeFailure
	// NodeRunning indicates the node is still executing.
	NodeRunning
)

// String returns the string representation of a node status.
func (s NodeStatus) String() string {
	switch s {
	case NodeSuccess:
		return "Success"
	case NodeFailure:
		return "Failure"
	case NodeRunning:
		return "Running"
	default:
		return "Unknown"
	}
}

// BehaviorNode is the interface that all behavior tree nodes must implement.
type BehaviorNode interface {
	// Tick executes the node logic and returns the result status.
	Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus
	// Reset resets the node state for fresh execution.
	Reset()
	// String returns a string representation of the node for debugging.
	String() string
}

// Blackboard is a shared data structure for behavior tree state.
// It stores key-value pairs that nodes can read and write to share information.
type Blackboard struct {
	data map[string]interface{}
	rng  *rand.Rand
}

// NewBlackboard creates a new empty blackboard with a default RNG.
// For deterministic behavior, use NewBlackboardWithSeed instead.
func NewBlackboard() *Blackboard {
	return &Blackboard{
		data: make(map[string]interface{}),
		rng:  rand.New(rand.NewSource(0)),
	}
}

// NewBlackboardWithSeed creates a new blackboard with a seeded RNG.
// Using the same seed ensures deterministic AI behavior.
func NewBlackboardWithSeed(seed int64) *Blackboard {
	return &Blackboard{
		data: make(map[string]interface{}),
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// GetRNG returns the seeded random number generator for deterministic behavior.
// All randomness in behavior tree actions should use this RNG.
func (b *Blackboard) GetRNG() *rand.Rand {
	return b.rng
}

// SetRNG sets the random number generator for this blackboard.
// This allows updating the seed for different scenarios.
func (b *Blackboard) SetRNG(rng *rand.Rand) {
	b.rng = rng
}

// Set stores a value in the blackboard.
func (b *Blackboard) Set(key string, value interface{}) {
	b.data[key] = value
}

// Get retrieves a value from the blackboard.
// Returns the value and true if found, nil and false otherwise.
func (b *Blackboard) Get(key string) (interface{}, bool) {
	val, ok := b.data[key]
	return val, ok
}

// GetFloat64 retrieves a float64 value from the blackboard.
// Returns the value and true if found and correct type, 0.0 and false otherwise.
func (b *Blackboard) GetFloat64(key string) (float64, bool) {
	val, ok := b.data[key]
	if !ok {
		return 0.0, false
	}
	floatVal, ok := val.(float64)
	return floatVal, ok
}

// GetBool retrieves a bool value from the blackboard.
func (b *Blackboard) GetBool(key string) (bool, bool) {
	val, ok := b.data[key]
	if !ok {
		return false, false
	}
	boolVal, ok := val.(bool)
	return boolVal, ok
}

// GetEntity retrieves an entity pointer from the blackboard.
func (b *Blackboard) GetEntity(key string) (*Entity, bool) {
	val, ok := b.data[key]
	if !ok {
		return nil, false
	}
	entityVal, ok := val.(*Entity)
	return entityVal, ok
}

// Clear removes all data from the blackboard.
func (b *Blackboard) Clear() {
	b.data = make(map[string]interface{})
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
func (s *SequenceNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	for s.currentIndex < len(s.children) {
		status := s.children[s.currentIndex].Tick(entity, blackboard, deltaTime)

		switch status {
		case NodeFailure:
			if s.resetOnFinish {
				s.Reset()
			}
			return NodeFailure
		case NodeRunning:
			return NodeRunning
		case NodeSuccess:
			s.currentIndex++
			continue
		}
	}

	// All children succeeded
	if s.resetOnFinish {
		s.Reset()
	}
	return NodeSuccess
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
func (s *SelectorNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	for s.currentIndex < len(s.children) {
		status := s.children[s.currentIndex].Tick(entity, blackboard, deltaTime)

		switch status {
		case NodeSuccess:
			if s.resetOnFinish {
				s.Reset()
			}
			return NodeSuccess
		case NodeRunning:
			return NodeRunning
		case NodeFailure:
			s.currentIndex++
			continue
		}
	}

	// All children failed
	if s.resetOnFinish {
		s.Reset()
	}
	return NodeFailure
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
func (p *ParallelNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	hasRunning := false
	hasFailure := false

	for _, child := range p.children {
		status := child.Tick(entity, blackboard, deltaTime)

		switch status {
		case NodeRunning:
			hasRunning = true
		case NodeFailure:
			hasFailure = true
		}
	}

	if hasFailure {
		return NodeFailure
	}
	if hasRunning {
		return NodeRunning
	}
	return NodeSuccess
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
func (i *InverterNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	status := i.child.Tick(entity, blackboard, deltaTime)

	switch status {
	case NodeSuccess:
		return NodeFailure
	case NodeFailure:
		return NodeSuccess
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
func (r *RepeatNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	if r.maxRepeats > 0 && r.currentRepeat >= r.maxRepeats {
		r.Reset()
		return NodeSuccess
	}

	status := r.child.Tick(entity, blackboard, deltaTime)

	switch status {
	case NodeSuccess:
		r.currentRepeat++
		r.child.Reset()
		if r.maxRepeats > 0 && r.currentRepeat >= r.maxRepeats {
			r.Reset()
			return NodeSuccess
		}
		return NodeRunning
	case NodeFailure:
		r.Reset()
		return NodeFailure
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
