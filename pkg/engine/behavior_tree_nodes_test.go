// Package engine provides tests for behavior tree nodes.
package engine

import (
	"testing"
)

// TestNodeStatus tests the NodeStatus string representation.
func TestNodeStatus(t *testing.T) {
	tests := []struct {
		status   NodeStatus
		expected string
	}{
		{NodeSuccess, "Success"},
		{NodeFailure, "Failure"},
		{NodeRunning, "Running"},
		{NodeStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("NodeStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestBlackboard tests the Blackboard data storage.
func TestBlackboard(t *testing.T) {
	bb := NewBlackboard()

	// Test Set and Get
	bb.Set("key1", "value1")
	val, ok := bb.Get("key1")
	if !ok {
		t.Fatal("Expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("Get() = %v, want %v", val, "value1")
	}

	// Test missing key
	_, ok = bb.Get("missing")
	if ok {
		t.Error("Expected missing key to return false")
	}

	// Test GetFloat64
	bb.Set("float", 42.5)
	floatVal, ok := bb.GetFloat64("float")
	if !ok || floatVal != 42.5 {
		t.Errorf("GetFloat64() = %v, %v, want 42.5, true", floatVal, ok)
	}

	// Test GetFloat64 wrong type
	bb.Set("notfloat", "string")
	_, ok = bb.GetFloat64("notfloat")
	if ok {
		t.Error("Expected GetFloat64() on string to return false")
	}

	// Test GetBool
	bb.Set("bool", true)
	boolVal, ok := bb.GetBool("bool")
	if !ok || !boolVal {
		t.Errorf("GetBool() = %v, %v, want true, true", boolVal, ok)
	}

	// Test GetEntity
	entity := NewEntity(123)
	bb.Set("entity", entity)
	entityVal, ok := GetEntityFromBlackboard(bb, "entity")
	if !ok || entityVal.ID != 123 {
		t.Errorf("GetEntity() failed")
	}

	// Test Clear
	bb.Clear()
	_, ok = bb.Get("key1")
	if ok {
		t.Error("Expected key1 to not exist after Clear()")
	}
}

// TestSequenceNode tests the sequence node behavior.
func TestSequenceNode(t *testing.T) {
	tests := []struct {
		name           string
		childResults   []NodeStatus
		expectedStatus NodeStatus
	}{
		{
			name:           "all success",
			childResults:   []NodeStatus{NodeSuccess, NodeSuccess, NodeSuccess},
			expectedStatus: NodeSuccess,
		},
		{
			name:           "first fails",
			childResults:   []NodeStatus{NodeFailure, NodeSuccess, NodeSuccess},
			expectedStatus: NodeFailure,
		},
		{
			name:           "middle fails",
			childResults:   []NodeStatus{NodeSuccess, NodeFailure, NodeSuccess},
			expectedStatus: NodeFailure,
		},
		{
			name:           "first running",
			childResults:   []NodeStatus{NodeRunning, NodeSuccess, NodeSuccess},
			expectedStatus: NodeRunning,
		},
		{
			name:           "middle running",
			childResults:   []NodeStatus{NodeSuccess, NodeRunning, NodeSuccess},
			expectedStatus: NodeRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock children that return predefined status
			children := make([]BehaviorNode, len(tt.childResults))
			for i, status := range tt.childResults {
				result := status // Capture for closure
				children[i] = NewActionNode("mock", func(*Entity, *Blackboard, float64) NodeStatus {
					return result
				})
			}

			seq := NewSequenceNode("test", children...)
			entity := NewEntity(1)
			bb := NewBlackboard()

			status := seq.Tick(entity, bb, 0.016)
			if status != tt.expectedStatus {
				t.Errorf("Tick() = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

// TestSelectorNode tests the selector node behavior.
func TestSelectorNode(t *testing.T) {
	tests := []struct {
		name           string
		childResults   []NodeStatus
		expectedStatus NodeStatus
	}{
		{
			name:           "first succeeds",
			childResults:   []NodeStatus{NodeSuccess, NodeFailure, NodeFailure},
			expectedStatus: NodeSuccess,
		},
		{
			name:           "middle succeeds",
			childResults:   []NodeStatus{NodeFailure, NodeSuccess, NodeFailure},
			expectedStatus: NodeSuccess,
		},
		{
			name:           "all fail",
			childResults:   []NodeStatus{NodeFailure, NodeFailure, NodeFailure},
			expectedStatus: NodeFailure,
		},
		{
			name:           "first running",
			childResults:   []NodeStatus{NodeRunning, NodeSuccess, NodeFailure},
			expectedStatus: NodeRunning,
		},
		{
			name:           "middle running",
			childResults:   []NodeStatus{NodeFailure, NodeRunning, NodeSuccess},
			expectedStatus: NodeRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]BehaviorNode, len(tt.childResults))
			for i, status := range tt.childResults {
				result := status
				children[i] = NewActionNode("mock", func(*Entity, *Blackboard, float64) NodeStatus {
					return result
				})
			}

			sel := NewSelectorNode("test", children...)
			entity := NewEntity(1)
			bb := NewBlackboard()

			status := sel.Tick(entity, bb, 0.016)
			if status != tt.expectedStatus {
				t.Errorf("Tick() = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

// TestParallelNode tests the parallel node behavior.
func TestParallelNode(t *testing.T) {
	tests := []struct {
		name           string
		childResults   []NodeStatus
		expectedStatus NodeStatus
	}{
		{
			name:           "all success",
			childResults:   []NodeStatus{NodeSuccess, NodeSuccess, NodeSuccess},
			expectedStatus: NodeSuccess,
		},
		{
			name:           "one fails",
			childResults:   []NodeStatus{NodeSuccess, NodeFailure, NodeSuccess},
			expectedStatus: NodeFailure,
		},
		{
			name:           "one running",
			childResults:   []NodeStatus{NodeSuccess, NodeRunning, NodeSuccess},
			expectedStatus: NodeRunning,
		},
		{
			name:           "running and failure",
			childResults:   []NodeStatus{NodeRunning, NodeFailure, NodeSuccess},
			expectedStatus: NodeFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]BehaviorNode, len(tt.childResults))
			for i, status := range tt.childResults {
				result := status
				children[i] = NewActionNode("mock", func(*Entity, *Blackboard, float64) NodeStatus {
					return result
				})
			}

			par := NewParallelNode("test", children...)
			entity := NewEntity(1)
			bb := NewBlackboard()

			status := par.Tick(entity, bb, 0.016)
			if status != tt.expectedStatus {
				t.Errorf("Tick() = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

// TestInverterNode tests the inverter decorator.
func TestInverterNode(t *testing.T) {
	tests := []struct {
		childStatus    NodeStatus
		expectedStatus NodeStatus
	}{
		{NodeSuccess, NodeFailure},
		{NodeFailure, NodeSuccess},
		{NodeRunning, NodeRunning}, // Running is not inverted
	}

	for _, tt := range tests {
		t.Run(tt.childStatus.String(), func(t *testing.T) {
			child := NewActionNode("mock", func(*Entity, *Blackboard, float64) NodeStatus {
				return tt.childStatus
			})

			inv := NewInverterNode("test", child)
			entity := NewEntity(1)
			bb := NewBlackboard()

			status := inv.Tick(entity, bb, 0.016)
			if status != tt.expectedStatus {
				t.Errorf("Tick() = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

// TestRepeatNode tests the repeat decorator.
func TestRepeatNode(t *testing.T) {
	t.Run("repeat 3 times", func(t *testing.T) {
		counter := 0
		child := NewActionNode("counter", func(*Entity, *Blackboard, float64) NodeStatus {
			counter++
			return NodeSuccess
		})

		repeat := NewRepeatNode("test", 3, child)
		entity := NewEntity(1)
		bb := NewBlackboard()

		// First 2 ticks should return Running
		for i := 0; i < 2; i++ {
			status := repeat.Tick(entity, bb, 0.016)
			if status != NodeRunning {
				t.Errorf("Tick %d = %v, want Running", i+1, status)
			}
		}

		// 3rd tick should return Success
		status := repeat.Tick(entity, bb, 0.016)
		if status != NodeSuccess {
			t.Errorf("Final tick = %v, want Success", status)
		}

		// Counter should be 3
		if counter != 3 {
			t.Errorf("Counter = %d, want 3", counter)
		}
	})

	t.Run("child fails", func(t *testing.T) {
		child := NewActionNode("fail", func(*Entity, *Blackboard, float64) NodeStatus {
			return NodeFailure
		})

		repeat := NewRepeatNode("test", 5, child)
		entity := NewEntity(1)
		bb := NewBlackboard()

		status := repeat.Tick(entity, bb, 0.016)
		if status != NodeFailure {
			t.Errorf("Tick() = %v, want Failure", status)
		}
	})
}

// TestConditionNode tests condition nodes.
func TestConditionNode(t *testing.T) {
	t.Run("condition true", func(t *testing.T) {
		cond := NewConditionNode("test", func(*Entity, *Blackboard) bool {
			return true
		})

		entity := NewEntity(1)
		bb := NewBlackboard()

		status := cond.Tick(entity, bb, 0.016)
		if status != NodeSuccess {
			t.Errorf("Tick() = %v, want Success", status)
		}
	})

	t.Run("condition false", func(t *testing.T) {
		cond := NewConditionNode("test", func(*Entity, *Blackboard) bool {
			return false
		})

		entity := NewEntity(1)
		bb := NewBlackboard()

		status := cond.Tick(entity, bb, 0.016)
		if status != NodeFailure {
			t.Errorf("Tick() = %v, want Failure", status)
		}
	})
}

// TestBehaviorTreeIntegration tests a complete behavior tree.
func TestBehaviorTreeIntegration(t *testing.T) {
	// Create a simple behavior tree: Selector(Condition->Action, DefaultAction)
	bb := NewBlackboard()
	entity := NewEntity(1)

	successAction := NewActionNode("success", func(*Entity, *Blackboard, float64) NodeStatus {
		return NodeSuccess
	})

	failAction := NewActionNode("fail", func(*Entity, *Blackboard, float64) NodeStatus {
		return NodeFailure
	})

	condition := NewConditionNode("check", func(e *Entity, b *Blackboard) bool {
		val, ok := b.GetBool("trigger")
		return ok && val
	})

	tree := NewSelectorNode("root",
		NewSequenceNode("conditional",
			condition,
			successAction,
		),
		failAction,
	)

	// Test with condition false - should execute fail action
	bb.Set("trigger", false)
	status := tree.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure when condition false, got %v", status)
	}

	// Test with condition true - should execute success action
	bb.Set("trigger", true)
	status = tree.Tick(entity, bb, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected Success when condition true, got %v", status)
	}
}
