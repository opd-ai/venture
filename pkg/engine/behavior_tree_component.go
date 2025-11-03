// Package engine provides behavior tree components for AI.
// This file defines BehaviorTreeComponent which stores behavior tree state.
package engine

// BehaviorTreeComponent holds a behavior tree for an entity.
type BehaviorTreeComponent struct {
	// Root node of the behavior tree
	Root BehaviorNode

	// Blackboard for shared state between nodes
	Blackboard *Blackboard

	// Whether the behavior tree is active
	Enabled bool

	// Tree name for debugging
	TreeName string
}

// Type returns the component type identifier.
func (b BehaviorTreeComponent) Type() string {
	return "behaviortree"
}

// NewBehaviorTreeComponent creates a new behavior tree component.
func NewBehaviorTreeComponent(root BehaviorNode, treeName string) *BehaviorTreeComponent {
	return &BehaviorTreeComponent{
		Root:       root,
		Blackboard: NewBlackboard(),
		Enabled:    true,
		TreeName:   treeName,
	}
}

// Reset resets the behavior tree to its initial state.
func (b *BehaviorTreeComponent) Reset() {
	if b.Root != nil {
		b.Root.Reset()
	}
	b.Blackboard.Clear()
}

// Tick executes the behavior tree for one frame.
func (b *BehaviorTreeComponent) Tick(entity *Entity, deltaTime float64) NodeStatus {
	if !b.Enabled || b.Root == nil {
		return NodeFailure
	}
	return b.Root.Tick(entity, b.Blackboard, deltaTime)
}
