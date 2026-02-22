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
// Note: Uses seed=0 for the blackboard RNG. For deterministic behavior with unique
// seeds per entity, use NewBehaviorTreeComponentWithSeed instead.
func NewBehaviorTreeComponent(root BehaviorNode, treeName string) *BehaviorTreeComponent {
	return &BehaviorTreeComponent{
		Root:       root,
		Blackboard: NewBlackboard(),
		Enabled:    true,
		TreeName:   treeName,
	}
}

// NewBehaviorTreeComponentWithSeed creates a behavior tree component with a seeded RNG.
// Using a unique seed per entity ensures deterministic yet varied AI behavior.
// The seed should typically be derived from the entity's unique ID combined with
// a world seed to ensure consistent behavior across save/load cycles.
func NewBehaviorTreeComponentWithSeed(root BehaviorNode, treeName string, seed int64) *BehaviorTreeComponent {
	return &BehaviorTreeComponent{
		Root:       root,
		Blackboard: NewBlackboardWithSeed(seed),
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
