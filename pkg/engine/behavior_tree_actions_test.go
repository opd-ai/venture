// Package engine provides tests for behavior tree actions and conditions.
package engine

import (
	"testing"
)

// TestHasTargetCondition tests the HasTarget condition.
func TestHasTargetCondition(t *testing.T) {
	entity := NewEntity(1)
	bb := NewBlackboard()

	cond := NewHasTargetCondition()

	// Test without target
	status := cond.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure without target, got %v", status)
	}

	// Test with target
	target := NewEntity(2)
	bb.Set("target", target)
	status = cond.Tick(entity, bb, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected Success with target, got %v", status)
	}
}

// TestHealthBelowCondition tests the HealthBelow condition.
func TestHealthBelowCondition(t *testing.T) {
	entity := NewEntity(1)
	bb := NewBlackboard()

	// Add health component
	health := &HealthComponent{Current: 20, Max: 100}
	entity.AddComponent(health)

	// Test health at 20% (20/100)
	cond := NewHealthBelowCondition(0.3) // Below 30%
	status := cond.Tick(entity, bb, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected Success when health below threshold, got %v", status)
	}

	// Test health not below threshold
	cond = NewHealthBelowCondition(0.1) // Below 10%
	status = cond.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure when health above threshold, got %v", status)
	}

	// Test without health component
	entity2 := NewEntity(2)
	status = cond.Tick(entity2, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure without health component, got %v", status)
	}
}

// TestTargetInRangeCondition tests the TargetInRange condition.
func TestTargetInRangeCondition(t *testing.T) {
	entity := NewEntity(1)
	target := NewEntity(2)
	bb := NewBlackboard()

	// Add position components
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	target.AddComponent(&PositionComponent{X: 50, Y: 0}) // 50 pixels away

	bb.Set("target", target)

	// Test target in range
	cond := NewTargetInRangeCondition(100.0)
	status := cond.Tick(entity, bb, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected Success when target in range, got %v", status)
	}

	// Test target out of range
	cond = NewTargetInRangeCondition(30.0)
	status = cond.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure when target out of range, got %v", status)
	}

	// Test without target
	bb.Set("target", nil)
	status = cond.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure without target, got %v", status)
	}
}

// TestMoveToTargetAction tests the MoveToTarget action.
func TestMoveToTargetAction(t *testing.T) {
	entity := NewEntity(1)
	target := NewEntity(2)
	bb := NewBlackboard()

	// Add position components
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	target.AddComponent(&PositionComponent{X: 100, Y: 0})

	// Add velocity component
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	bb.Set("target", target)

	action := NewMoveToTargetAction(50.0)
	status := action.Tick(entity, bb, 0.016)

	// Should return Running (moving towards target)
	if status != NodeRunning {
		t.Errorf("Expected Running when moving, got %v", status)
	}

	// Check velocity was set
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 && vel.VY == 0 {
		t.Error("Expected velocity to be set")
	}

	// Test when already at target
	target.AddComponent(&PositionComponent{X: 0, Y: 0})
	status = action.Tick(entity, bb, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected Success when at target, got %v", status)
	}
}

// TestAttackTargetAction tests the AttackTarget action.
func TestAttackTargetAction(t *testing.T) {
	entity := NewEntity(1)
	target := NewEntity(2)
	bb := NewBlackboard()

	// Add attack component
	entity.AddComponent(&AttackComponent{
		Damage:        10,
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})

	bb.Set("target", target)

	action := NewAttackTargetAction()

	// First attack should succeed
	status := action.Tick(entity, bb, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected Success on attack, got %v", status)
	}

	// Check cooldown was set
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.CooldownTimer <= 0 {
		t.Error("Expected CooldownTimer to be set")
	}

	// Second attack should be on cooldown
	status = action.Tick(entity, bb, 0.016)
	if status != NodeRunning {
		t.Errorf("Expected Running on cooldown, got %v", status)
	}

	// Test without target
	bb.Set("target", nil)
	status = action.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure without target, got %v", status)
	}
}

// TestFleeFromTargetAction tests the FleeFromTarget action.
func TestFleeFromTargetAction(t *testing.T) {
	entity := NewEntity(1)
	target := NewEntity(2)
	bb := NewBlackboard()

	// Add position components
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	target.AddComponent(&PositionComponent{X: 100, Y: 50})

	// Add velocity component
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	bb.Set("target", target)

	action := NewFleeFromTargetAction(50.0)
	status := action.Tick(entity, bb, 0.016)

	// Should return Running (fleeing from target)
	if status != NodeRunning {
		t.Errorf("Expected Running when fleeing, got %v", status)
	}

	// Check velocity is away from target (negative X direction)
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX >= 0 {
		t.Errorf("Expected negative velocity (fleeing left), got VX=%f", vel.VX)
	}
}

// TestWanderAction tests the Wander action.
func TestWanderAction(t *testing.T) {
	entity := NewEntity(1)
	bb := NewBlackboard()

	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	action := NewWanderAction(50.0)

	// First tick should set wander target
	status := action.Tick(entity, bb, 0.016)
	if status != NodeRunning {
		t.Errorf("Expected Running when starting wander, got %v", status)
	}

	// Velocity should be set
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 && vel.VY == 0 {
		t.Error("Expected velocity to be set during wander")
	}

	// Continue ticking until wander completes
	for i := 0; i < 400; i++ { // Increased iterations (wander can take up to 5 seconds, 400*0.016 = 6.4s)
		status = action.Tick(entity, bb, 0.016)
		if status == NodeSuccess {
			break
		}
	}

	if status != NodeSuccess {
		t.Error("Expected wander to complete eventually")
	}

	// Velocity should be cleared after wander
	velComp, _ = entity.GetComponent("velocity")
	vel = velComp.(*VelocityComponent)
	if vel.VX != 0 || vel.VY != 0 {
		t.Error("Expected velocity to be cleared after wander")
	}
}

// TestWaitAction tests the Wait action.
func TestWaitAction(t *testing.T) {
	entity := NewEntity(1)
	bb := NewBlackboard()

	action := NewWaitAction(1.0) // Wait 1 second

	// First tick should start waiting
	status := action.Tick(entity, bb, 0.016)
	if status != NodeRunning {
		t.Errorf("Expected Running when waiting, got %v", status)
	}

	// Continue for ~1 second
	totalTime := 0.0
	for i := 0; i < 100; i++ { // Simulate ~1.6 seconds
		deltaTime := 0.016
		status = action.Tick(entity, bb, deltaTime)
		totalTime += deltaTime
		if status == NodeSuccess {
			break
		}
	}

	if status != NodeSuccess {
		t.Error("Expected wait to complete after duration")
	}
}

// TestClearTargetAction tests the ClearTarget action.
func TestClearTargetAction(t *testing.T) {
	entity := NewEntity(1)
	bb := NewBlackboard()

	// Set a target
	target := NewEntity(2)
	bb.Set("target", target)

	action := NewClearTargetAction()
	status := action.Tick(entity, bb, 0.016)

	if status != NodeSuccess {
		t.Errorf("Expected Success, got %v", status)
	}

	// Check target was cleared
	clearedTarget, ok := bb.GetEntity("target")
	if ok && clearedTarget != nil {
		t.Error("Expected target to be cleared")
	}
}

// TestFindTargetAction tests the FindTarget action.
func TestFindTargetAction(t *testing.T) {
	// Create a minimal world for testing
	world := NewWorld()

	// Create test entity with team
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TeamComponent{TeamID: 1})
	world.AddEntity(entity)

	// Create enemy entity
	enemy := NewEntity(2)
	enemy.AddComponent(&PositionComponent{X: 150, Y: 100}) // 50 pixels away
	enemy.AddComponent(&TeamComponent{TeamID: 2})          // Different team
	world.AddEntity(enemy)

	// Process entity additions
	world.Update(0)

	bb := NewBlackboard()
	action := NewFindTargetAction(200.0, world)

	status := action.Tick(entity, bb, 0.016)

	// Should find the enemy
	if status != NodeSuccess {
		t.Errorf("Expected Success when enemy in range, got %v", status)
	}

	// Check target was set
	target, ok := bb.GetEntity("target")
	if !ok || target == nil {
		t.Fatal("Expected target to be set")
	}
	if target.ID != 2 {
		t.Errorf("Expected target ID 2, got %d", target.ID)
	}

	// Test with no enemies in range
	enemy.Components["position"] = &PositionComponent{X: 500, Y: 500} // Too far
	bb.Set("target", nil)

	status = action.Tick(entity, bb, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected Failure when no enemies in range, got %v", status)
	}
}
