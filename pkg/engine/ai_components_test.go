package engine

import (
	"strings"
	"testing"
)

// TestAIState_String tests the string representation of AI states
func TestAIState_String(t *testing.T) {
	tests := []struct {
		name  string
		state AIState
		want  string
	}{
		{"Idle state", AIStateIdle, "Idle"},
		{"Patrol state", AIStatePatrol, "Patrol"},
		{"Detect state", AIStateDetect, "Detect"},
		{"Chase state", AIStateChase, "Chase"},
		{"Attack state", AIStateAttack, "Attack"},
		{"Flee state", AIStateFlee, "Flee"},
		{"Return state", AIStateReturn, "Return"},
		{"Unknown state", AIState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("AIState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAIComponent_Type tests the component type identifier
func TestAIComponent_Type(t *testing.T) {
	ai := NewAIComponent(0, 0)
	if ai.Type() != "ai" {
		t.Errorf("AIComponent.Type() = %v, want 'ai'", ai.Type())
	}
}

// TestNewAIComponent tests the constructor with default values
func TestNewAIComponent(t *testing.T) {
	spawnX, spawnY := 100.0, 200.0
	ai := NewAIComponent(spawnX, spawnY)

	if ai == nil {
		t.Fatal("NewAIComponent returned nil")
	}

	// Verify initial state
	if ai.State != AIStateIdle {
		t.Errorf("Initial state = %v, want %v", ai.State, AIStateIdle)
	}

	// Verify spawn position
	if ai.SpawnX != spawnX {
		t.Errorf("SpawnX = %v, want %v", ai.SpawnX, spawnX)
	}
	if ai.SpawnY != spawnY {
		t.Errorf("SpawnY = %v, want %v", ai.SpawnY, spawnY)
	}

	// Verify no initial target
	if ai.Target != nil {
		t.Error("Initial target should be nil")
	}

	// Verify default ranges
	if ai.DetectionRange != 200.0 {
		t.Errorf("DetectionRange = %v, want 200.0", ai.DetectionRange)
	}
	if ai.FleeHealthThreshold != 0.2 {
		t.Errorf("FleeHealthThreshold = %v, want 0.2", ai.FleeHealthThreshold)
	}
	if ai.MaxChaseDistance != 500.0 {
		t.Errorf("MaxChaseDistance = %v, want 500.0", ai.MaxChaseDistance)
	}

	// Verify timers
	if ai.DecisionTimer != 0.0 {
		t.Errorf("DecisionTimer = %v, want 0.0", ai.DecisionTimer)
	}
	if ai.DecisionInterval != 0.5 {
		t.Errorf("DecisionInterval = %v, want 0.5", ai.DecisionInterval)
	}
	if ai.StateTimer != 0.0 {
		t.Errorf("StateTimer = %v, want 0.0", ai.StateTimer)
	}

	// Verify speed multipliers
	if ai.PatrolSpeed != 0.5 {
		t.Errorf("PatrolSpeed = %v, want 0.5", ai.PatrolSpeed)
	}
	if ai.ChaseSpeed != 1.0 {
		t.Errorf("ChaseSpeed = %v, want 1.0", ai.ChaseSpeed)
	}
	if ai.FleeSpeed != 1.5 {
		t.Errorf("FleeSpeed = %v, want 1.5", ai.FleeSpeed)
	}
	if ai.ReturnSpeed != 0.8 {
		t.Errorf("ReturnSpeed = %v, want 0.8", ai.ReturnSpeed)
	}
}

// TestAIComponent_ShouldUpdateDecision tests decision timer logic
func TestAIComponent_ShouldUpdateDecision(t *testing.T) {
	ai := NewAIComponent(0, 0)
	ai.DecisionInterval = 1.0

	// First update should trigger immediately (timer at 0)
	if !ai.ShouldUpdateDecision(0.1) {
		t.Error("Should update decision when timer is 0")
	}

	// Timer should be reset to interval
	if ai.DecisionTimer != 1.0 {
		t.Errorf("DecisionTimer = %v, want 1.0 after reset", ai.DecisionTimer)
	}

	// Subsequent updates should not trigger until interval passes
	if ai.ShouldUpdateDecision(0.3) {
		t.Error("Should not update decision before interval passes")
	}
	tolerance := 0.0001
	if diff := ai.DecisionTimer - 0.7; diff < -tolerance || diff > tolerance {
		t.Errorf("DecisionTimer = %v, want 0.7", ai.DecisionTimer)
	}

	if ai.ShouldUpdateDecision(0.3) {
		t.Error("Should not update decision yet")
	}
	if diff := ai.DecisionTimer - 0.4; diff < -tolerance || diff > tolerance {
		t.Errorf("DecisionTimer = %v, want 0.4", ai.DecisionTimer)
	}

	// Now it should trigger
	if !ai.ShouldUpdateDecision(0.5) {
		t.Error("Should update decision after interval passes")
	}
	if ai.DecisionTimer != 1.0 {
		t.Errorf("DecisionTimer = %v, want 1.0 after reset", ai.DecisionTimer)
	}
}

// TestAIComponent_UpdateStateTimer tests state timer updates
func TestAIComponent_UpdateStateTimer(t *testing.T) {
	ai := NewAIComponent(0, 0)

	// Initial state timer should be 0
	if ai.StateTimer != 0.0 {
		t.Errorf("Initial StateTimer = %v, want 0.0", ai.StateTimer)
	}

	// Update by 1 second
	ai.UpdateStateTimer(1.0)
	if ai.StateTimer != 1.0 {
		t.Errorf("StateTimer = %v, want 1.0", ai.StateTimer)
	}

	// Update by 0.5 seconds
	ai.UpdateStateTimer(0.5)
	if ai.StateTimer != 1.5 {
		t.Errorf("StateTimer = %v, want 1.5", ai.StateTimer)
	}

	// Multiple small updates
	for i := 0; i < 10; i++ {
		ai.UpdateStateTimer(0.1)
	}
	tolerance := 0.0001
	expected := 2.5
	if diff := ai.StateTimer - expected; diff < -tolerance || diff > tolerance {
		t.Errorf("StateTimer = %v, want %v", ai.StateTimer, expected)
	}
}

// TestAIComponent_ChangeState tests state transitions
func TestAIComponent_ChangeState(t *testing.T) {
	ai := NewAIComponent(0, 0)
	ai.StateTimer = 5.0

	// Change to a new state
	ai.ChangeState(AIStateChase)

	if ai.State != AIStateChase {
		t.Errorf("State = %v, want %v", ai.State, AIStateChase)
	}
	if ai.StateTimer != 0.0 {
		t.Errorf("StateTimer should reset to 0 on state change, got %v", ai.StateTimer)
	}

	// Change to the same state (should not reset timer)
	ai.StateTimer = 3.0
	ai.ChangeState(AIStateChase)

	if ai.StateTimer != 3.0 {
		t.Errorf("StateTimer should not reset when changing to same state, got %v", ai.StateTimer)
	}

	// Change to another state
	ai.ChangeState(AIStateFlee)

	if ai.State != AIStateFlee {
		t.Errorf("State = %v, want %v", ai.State, AIStateFlee)
	}
	if ai.StateTimer != 0.0 {
		t.Errorf("StateTimer should reset on state change, got %v", ai.StateTimer)
	}
}

// TestAIComponent_GetSpeedMultiplier tests speed multiplier for different states
func TestAIComponent_GetSpeedMultiplier(t *testing.T) {
	ai := NewAIComponent(0, 0)

	tests := []struct {
		name  string
		state AIState
		want  float64
	}{
		{"Idle state", AIStateIdle, 1.0},
		{"Patrol state", AIStatePatrol, 0.5},
		{"Detect state", AIStateDetect, 1.0},
		{"Chase state", AIStateChase, 1.0},
		{"Attack state", AIStateAttack, 1.0},
		{"Flee state", AIStateFlee, 1.5},
		{"Return state", AIStateReturn, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai.State = tt.state
			got := ai.GetSpeedMultiplier()
			if got != tt.want {
				t.Errorf("GetSpeedMultiplier() for state %v = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

// TestAIComponent_GetSpeedMultiplier_CustomSpeeds tests custom speed values
func TestAIComponent_GetSpeedMultiplier_CustomSpeeds(t *testing.T) {
	ai := NewAIComponent(0, 0)

	// Set custom speeds
	ai.PatrolSpeed = 0.3
	ai.ChaseSpeed = 1.5
	ai.FleeSpeed = 2.0
	ai.ReturnSpeed = 0.6

	tests := []struct {
		state AIState
		want  float64
	}{
		{AIStatePatrol, 0.3},
		{AIStateChase, 1.5},
		{AIStateFlee, 2.0},
		{AIStateReturn, 0.6},
	}

	for _, tt := range tests {
		ai.State = tt.state
		got := ai.GetSpeedMultiplier()
		if got != tt.want {
			t.Errorf("GetSpeedMultiplier() for state %v = %v, want %v", tt.state, got, tt.want)
		}
	}
}

// TestAIComponent_IsAggressiveState tests aggressive state detection
func TestAIComponent_IsAggressiveState(t *testing.T) {
	ai := NewAIComponent(0, 0)

	tests := []struct {
		name  string
		state AIState
		want  bool
	}{
		{"Idle is not aggressive", AIStateIdle, false},
		{"Patrol is not aggressive", AIStatePatrol, false},
		{"Detect is not aggressive", AIStateDetect, false},
		{"Chase is aggressive", AIStateChase, true},
		{"Attack is aggressive", AIStateAttack, true},
		{"Flee is not aggressive", AIStateFlee, false},
		{"Return is not aggressive", AIStateReturn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai.State = tt.state
			got := ai.IsAggressiveState()
			if got != tt.want {
				t.Errorf("IsAggressiveState() for state %v = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

// TestAIComponent_HasTarget tests target detection
func TestAIComponent_HasTarget(t *testing.T) {
	ai := NewAIComponent(0, 0)

	// Initially no target
	if ai.HasTarget() {
		t.Error("Should not have target initially")
	}

	// Assign a target
	target := NewEntity(123)
	ai.Target = target

	if !ai.HasTarget() {
		t.Error("Should have target after assignment")
	}

	// Clear target
	ai.Target = nil

	if ai.HasTarget() {
		t.Error("Should not have target after clearing")
	}
}

// TestAIComponent_ClearTarget tests target clearing
func TestAIComponent_ClearTarget(t *testing.T) {
	ai := NewAIComponent(0, 0)
	target := NewEntity(456)
	ai.Target = target

	if !ai.HasTarget() {
		t.Error("Should have target before clearing")
	}

	ai.ClearTarget()

	if ai.HasTarget() {
		t.Error("Should not have target after ClearTarget()")
	}
	if ai.Target != nil {
		t.Error("Target should be nil after ClearTarget()")
	}
}

// TestAIComponent_GetDistanceFromSpawn tests distance calculation
func TestAIComponent_GetDistanceFromSpawn(t *testing.T) {
	ai := NewAIComponent(100, 100)

	tests := []struct {
		name    string
		x, y    float64
		wantMin float64
		wantMax float64
	}{
		{"At spawn", 100, 100, 0, 0.001},
		{"100 units east", 200, 100, 99.9, 100.1},
		{"100 units north", 100, 200, 99.9, 100.1},
		{"Diagonal 100,100", 200, 200, 141.4, 141.5}, // sqrt(100^2 + 100^2) ≈ 141.42
		{"Far away", 500, 500, 565, 566},             // sqrt(400^2 + 400^2) ≈ 565.69
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ai.GetDistanceFromSpawn(tt.x, tt.y)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetDistanceFromSpawn(%v, %v) = %v, want between %v and %v",
					tt.x, tt.y, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestAIComponent_ShouldReturnToSpawn tests return-to-spawn logic
func TestAIComponent_ShouldReturnToSpawn(t *testing.T) {
	tests := []struct {
		name               string
		spawnX, spawnY     float64
		currentX, currentY float64
		maxDistance        float64
		want               bool
	}{
		{"At spawn", 100, 100, 100, 100, 500, false},
		{"Within range", 100, 100, 200, 100, 500, false},
		{"Just beyond range", 100, 100, 601, 100, 500, true},
		{"No distance limit", 100, 100, 1000, 1000, 0, false},
		{"Negative limit (unlimited)", 100, 100, 1000, 1000, -1, false},
		{"Exactly at limit", 100, 100, 600, 100, 500, false}, // 500 units away
		{"Just over limit", 100, 100, 601, 100, 500, true},   // 501 units away
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := NewAIComponent(tt.spawnX, tt.spawnY)
			ai.MaxChaseDistance = tt.maxDistance

			got := ai.ShouldReturnToSpawn(tt.currentX, tt.currentY)
			if got != tt.want {
				distance := ai.GetDistanceFromSpawn(tt.currentX, tt.currentY)
				t.Errorf("ShouldReturnToSpawn() = %v, want %v (distance: %v, max: %v)",
					got, tt.want, distance, tt.maxDistance)
			}
		})
	}
}

// TestAIComponent_String tests string representation
func TestAIComponent_String(t *testing.T) {
	ai := NewAIComponent(0, 0)

	// Test without target
	str := ai.String()
	if !strings.Contains(str, "Idle") {
		t.Errorf("String should contain state 'Idle', got: %s", str)
	}
	if !strings.Contains(str, "none") {
		t.Errorf("String should contain 'none' for no target, got: %s", str)
	}
	if !strings.Contains(str, "200") {
		t.Errorf("String should contain detection range '200', got: %s", str)
	}

	// Test with target
	target := NewEntity(789)
	ai.Target = target
	ai.State = AIStateChase
	ai.StateTimer = 2.5

	str = ai.String()
	if !strings.Contains(str, "Chase") {
		t.Errorf("String should contain state 'Chase', got: %s", str)
	}
	if !strings.Contains(str, "entity-789") {
		t.Errorf("String should contain target ID 'entity-789', got: %s", str)
	}
	if !strings.Contains(str, "2.5") {
		t.Errorf("String should contain timer '2.5', got: %s", str)
	}
}

// TestAIComponent_Integration tests full workflow
func TestAIComponent_Integration(t *testing.T) {
	// Create AI at spawn position
	ai := NewAIComponent(100, 100)

	// Verify initial state
	if ai.State != AIStateIdle {
		t.Fatalf("Initial state should be Idle, got %v", ai.State)
	}
	if ai.HasTarget() {
		t.Error("Should not have target initially")
	}

	// Transition to patrol
	ai.ChangeState(AIStatePatrol)
	if ai.State != AIStatePatrol {
		t.Error("Failed to change to Patrol state")
	}
	if ai.GetSpeedMultiplier() != 0.5 {
		t.Error("Patrol speed should be 0.5")
	}

	// Detect a target
	target := NewEntity(1)
	ai.Target = target
	ai.ChangeState(AIStateDetect)

	if !ai.HasTarget() {
		t.Error("Should have target in Detect state")
	}

	// Chase the target
	ai.ChangeState(AIStateChase)
	if !ai.IsAggressiveState() {
		t.Error("Chase should be an aggressive state")
	}
	if ai.GetSpeedMultiplier() != 1.0 {
		t.Error("Chase speed should be 1.0")
	}

	// Update decision timer
	shouldUpdate := ai.ShouldUpdateDecision(0.1)
	if !shouldUpdate {
		t.Error("Should update decision on first call")
	}

	// Simulate moving far from spawn
	ai.MaxChaseDistance = 500
	if ai.ShouldReturnToSpawn(700, 100) {
		// If too far, return to spawn
		ai.ChangeState(AIStateReturn)
		ai.ClearTarget()

		if ai.State != AIStateReturn {
			t.Error("Should be in Return state")
		}
		if ai.HasTarget() {
			t.Error("Should not have target when returning")
		}
		if ai.GetSpeedMultiplier() != 0.8 {
			t.Error("Return speed should be 0.8")
		}
	}

	// Back at spawn
	ai.ChangeState(AIStateIdle)
	if ai.IsAggressiveState() {
		t.Error("Idle should not be aggressive")
	}
}

// TestAIComponent_StateTransitions tests various state transitions
func TestAIComponent_StateTransitions(t *testing.T) {
	ai := NewAIComponent(0, 0)

	// Test all possible transitions
	transitions := []AIState{
		AIStateIdle,
		AIStatePatrol,
		AIStateDetect,
		AIStateChase,
		AIStateAttack,
		AIStateFlee,
		AIStateReturn,
		AIStateIdle,
	}

	for i, state := range transitions {
		ai.ChangeState(state)
		if ai.State != state {
			t.Errorf("Transition %d: failed to change to %v", i, state)
		}
	}
}

// TestAIComponent_TimerAccuracy tests timer precision
func TestAIComponent_TimerAccuracy(t *testing.T) {
	ai := NewAIComponent(0, 0)
	ai.DecisionInterval = 0.1

	// Run many small updates
	updateCount := 0
	totalTime := 0.0
	deltaTime := 0.001 // 1ms updates

	for i := 0; i < 1000; i++ {
		if ai.ShouldUpdateDecision(deltaTime) {
			updateCount++
		}
		totalTime += deltaTime
	}

	// Should have approximately 10 updates (1s / 0.1s interval)
	expectedUpdates := int(totalTime / ai.DecisionInterval)
	tolerance := 2 // Allow +/- 2 updates due to rounding

	if updateCount < expectedUpdates-tolerance || updateCount > expectedUpdates+tolerance {
		t.Errorf("Update count = %d, expected approximately %d (tolerance ±%d)",
			updateCount, expectedUpdates, tolerance)
	}
}
