package engine

import (
	"fmt"
	"math"
)

// AIState represents the current behavior state of an AI-controlled entity.
type AIState int

const (
	// AIStateIdle means the entity is stationary and passive.
	AIStateIdle AIState = iota
	// AIStatePatrol means the entity is moving along a patrol route.
	AIStatePatrol
	// AIStateDetect means the entity has noticed a potential target.
	AIStateDetect
	// AIStateChase means the entity is actively pursuing a target.
	AIStateChase
	// AIStateAttack means the entity is in combat with a target.
	AIStateAttack
	// AIStateFlee means the entity is retreating from danger.
	AIStateFlee
	// AIStateReturn means the entity is returning to its spawn point.
	AIStateReturn
)

// String returns the string representation of an AI state.
func (s AIState) String() string {
	switch s {
	case AIStateIdle:
		return "Idle"
	case AIStatePatrol:
		return "Patrol"
	case AIStateDetect:
		return "Detect"
	case AIStateChase:
		return "Chase"
	case AIStateAttack:
		return "Attack"
	case AIStateFlee:
		return "Flee"
	case AIStateReturn:
		return "Return"
	default:
		return "Unknown"
	}
}

// AIComponent manages the behavior state and decision-making for an AI-controlled entity.
// It works with movement, combat, and team components to create intelligent enemies.
type AIComponent struct {
	// Current behavior state
	State AIState

	// Target entity being pursued or attacked (nil if no target)
	Target *Entity

	// Spawn position for returning after combat
	SpawnX, SpawnY float64

	// Detection range for finding enemies (in pixels)
	DetectionRange float64

	// Range at which to start fleeing (based on health percentage, 0.0-1.0)
	FleeHealthThreshold float64

	// Maximum distance from spawn before returning (0 = unlimited)
	MaxChaseDistance float64

	// Time until next decision update (in seconds)
	DecisionTimer float64

	// How often to make decisions (in seconds)
	DecisionInterval float64

	// Time spent in current state (for state-specific behaviors)
	StateTimer float64

	// Speed multiplier for different states (e.g., faster when fleeing)
	PatrolSpeed float64
	ChaseSpeed  float64
	FleeSpeed   float64
	ReturnSpeed float64
}

// Type returns the component type identifier.
func (a AIComponent) Type() string {
	return "ai"
}

// NewAIComponent creates a new AI component with sensible defaults.
func NewAIComponent(spawnX, spawnY float64) *AIComponent {
	return &AIComponent{
		State:               AIStateIdle,
		Target:              nil,
		SpawnX:              spawnX,
		SpawnY:              spawnY,
		DetectionRange:      200.0, // Can detect enemies within 200 pixels
		FleeHealthThreshold: 0.2,   // Flee when below 20% health
		MaxChaseDistance:    500.0, // Don't chase more than 500 pixels from spawn
		DecisionTimer:       0.0,
		DecisionInterval:    0.5, // Make decisions twice per second
		StateTimer:          0.0,
		PatrolSpeed:         0.5, // Half speed when patrolling
		ChaseSpeed:          1.0, // Normal speed when chasing
		FleeSpeed:           1.5, // 50% faster when fleeing
		ReturnSpeed:         0.8, // Slightly slower when returning
	}
}

// ShouldUpdateDecision checks if it's time to make a new AI decision.
func (a *AIComponent) ShouldUpdateDecision(deltaTime float64) bool {
	a.DecisionTimer -= deltaTime
	if a.DecisionTimer <= 0 {
		a.DecisionTimer = a.DecisionInterval
		return true
	}
	return false
}

// UpdateStateTimer updates the time spent in the current state.
func (a *AIComponent) UpdateStateTimer(deltaTime float64) {
	a.StateTimer += deltaTime
}

// ChangeState transitions to a new AI state and resets the state timer.
func (a *AIComponent) ChangeState(newState AIState) {
	if a.State != newState {
		a.State = newState
		a.StateTimer = 0.0
	}
}

// GetSpeedMultiplier returns the appropriate speed multiplier for the current state.
func (a *AIComponent) GetSpeedMultiplier() float64 {
	switch a.State {
	case AIStatePatrol:
		return a.PatrolSpeed
	case AIStateChase, AIStateAttack:
		return a.ChaseSpeed
	case AIStateFlee:
		return a.FleeSpeed
	case AIStateReturn:
		return a.ReturnSpeed
	default:
		return 1.0
	}
}

// IsAggressiveState returns true if the AI is in a combat-related state.
func (a *AIComponent) IsAggressiveState() bool {
	return a.State == AIStateChase || a.State == AIStateAttack
}

// HasTarget returns true if the AI has a valid target.
func (a *AIComponent) HasTarget() bool {
	return a.Target != nil
}

// ClearTarget removes the current target.
func (a *AIComponent) ClearTarget() {
	a.Target = nil
}

// GetDistanceFromSpawn calculates how far the entity is from its spawn point.
func (a *AIComponent) GetDistanceFromSpawn(currentX, currentY float64) float64 {
	dx := currentX - a.SpawnX
	dy := currentY - a.SpawnY
	return math.Sqrt(dx*dx + dy*dy)
}

// ShouldReturnToSpawn checks if the entity has wandered too far from spawn.
func (a *AIComponent) ShouldReturnToSpawn(currentX, currentY float64) bool {
	if a.MaxChaseDistance <= 0 {
		return false // No distance limit
	}
	return a.GetDistanceFromSpawn(currentX, currentY) > a.MaxChaseDistance
}

// String returns a string representation of the component.
func (a *AIComponent) String() string {
	targetInfo := "none"
	if a.HasTarget() {
		targetInfo = fmt.Sprintf("entity-%d", a.Target.ID)
	}
	return fmt.Sprintf("AI State: %s, Target: %s, Detection: %.0f, Timer: %.2f",
		a.State.String(), targetInfo, a.DetectionRange, a.StateTimer)
}
