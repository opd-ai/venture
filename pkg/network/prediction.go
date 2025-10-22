// Package network provides client-side prediction for responsive gameplay.
// This file implements client-side prediction and server reconciliation
// to ensure smooth gameplay despite network latency.
package network

import (
	"sync"
	"time"
)

// PredictedState represents a client-side predicted state for an input
type PredictedState struct {
	Sequence  uint32    // Input sequence number
	Timestamp time.Time // When the input was sent
	Position  Position  // Predicted position
	Velocity  Velocity  // Predicted velocity
}

// Position represents a 2D position
type Position struct {
	X, Y float64
}

// Velocity represents 2D velocity
type Velocity struct {
	VX, VY float64
}

// ClientPredictor handles client-side prediction and reconciliation
type ClientPredictor struct {
	mu sync.RWMutex

	// History of predicted states for reconciliation
	stateHistory []PredictedState

	// Maximum number of states to keep in history
	maxHistory int

	// Current predicted state
	currentState PredictedState

	// Last acknowledged sequence from server
	lastAckedSeq uint32

	// Current sequence number
	currentSeq uint32
}

// NewClientPredictor creates a new client-side predictor
func NewClientPredictor() *ClientPredictor {
	return &ClientPredictor{
		stateHistory: make([]PredictedState, 0, 128),
		maxHistory:   128, // Keep last 128 states (6.4 seconds at 20Hz)
		currentSeq:   0,
	}
}

// PredictInput predicts the result of applying an input and stores it for reconciliation
func (cp *ClientPredictor) PredictInput(dx, dy float64, deltaTime float64) PredictedState {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.currentSeq++

	// Apply movement prediction
	newVX := cp.currentState.Velocity.VX + dx*deltaTime
	newVY := cp.currentState.Velocity.VY + dy*deltaTime

	// Apply velocity to position
	newX := cp.currentState.Position.X + newVX*deltaTime
	newY := cp.currentState.Position.Y + newVY*deltaTime

	predicted := PredictedState{
		Sequence:  cp.currentSeq,
		Timestamp: time.Now(),
		Position:  Position{X: newX, Y: newY},
		Velocity:  Velocity{VX: newVX, VY: newVY},
	}

	// Store in history
	cp.stateHistory = append(cp.stateHistory, predicted)

	// Trim history if needed
	if len(cp.stateHistory) > cp.maxHistory {
		cp.stateHistory = cp.stateHistory[1:]
	}

	cp.currentState = predicted
	return predicted
}

// ReconcileServerState reconciles the client's predicted state with authoritative server state
// Returns the corrected current state after replaying unacknowledged inputs
func (cp *ClientPredictor) ReconcileServerState(serverSeq uint32, serverPos Position, serverVel Velocity) PredictedState {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.lastAckedSeq = serverSeq

	// Find the state that corresponds to the server's acknowledged sequence
	var stateIndex = -1
	for i, state := range cp.stateHistory {
		if state.Sequence == serverSeq {
			stateIndex = i
			break
		}
	}

	// If we don't have the state anymore, trust the server completely
	if stateIndex == -1 {
		cp.currentState = PredictedState{
			Sequence:  serverSeq,
			Timestamp: time.Now(),
			Position:  serverPos,
			Velocity:  serverVel,
		}
		cp.stateHistory = make([]PredictedState, 0, cp.maxHistory)
		return cp.currentState
	}

	// Check if there's a significant difference (prediction error)
	predicted := cp.stateHistory[stateIndex]
	errorX := serverPos.X - predicted.Position.X
	errorY := serverPos.Y - predicted.Position.Y
	errorThreshold := 1.0 // 1 unit difference threshold

	if abs(errorX) < errorThreshold && abs(errorY) < errorThreshold {
		// Prediction was accurate, no correction needed
		// Just remove old states
		cp.stateHistory = cp.stateHistory[stateIndex+1:]
		return cp.currentState
	}

	// Prediction error detected, need to correct and replay
	// Start from the server's authoritative state
	correctedState := PredictedState{
		Sequence:  serverSeq,
		Timestamp: time.Now(),
		Position:  serverPos,
		Velocity:  serverVel,
	}

	// Replay all inputs that came after the acknowledged one
	inputsToReplay := cp.stateHistory[stateIndex+1:]
	for i, oldState := range inputsToReplay {
		// Calculate deltaTime between states
		var deltaTime float64
		if i > 0 {
			deltaTime = oldState.Timestamp.Sub(inputsToReplay[i-1].Timestamp).Seconds()
		} else if stateIndex >= 0 {
			deltaTime = oldState.Timestamp.Sub(predicted.Timestamp).Seconds()
		} else {
			deltaTime = 0.05 // Default to 50ms
		}

		// Clamp deltaTime
		if deltaTime > 0.1 {
			deltaTime = 0.1
		}

		// Re-apply the input
		dx := oldState.Velocity.VX - correctedState.Velocity.VX
		dy := oldState.Velocity.VY - correctedState.Velocity.VY

		correctedState.Velocity.VX += dx
		correctedState.Velocity.VY += dy
		correctedState.Position.X += correctedState.Velocity.VX * deltaTime
		correctedState.Position.Y += correctedState.Velocity.VY * deltaTime
		correctedState.Sequence = oldState.Sequence
	}

	// Update history to only keep replayed states
	cp.stateHistory = inputsToReplay
	cp.currentState = correctedState

	return correctedState
}

// GetCurrentState returns the current predicted state
func (cp *ClientPredictor) GetCurrentState() PredictedState {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.currentState
}

// SetInitialState sets the initial state (usually from server)
func (cp *ClientPredictor) SetInitialState(pos Position, vel Velocity) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.currentState = PredictedState{
		Sequence:  0,
		Timestamp: time.Now(),
		Position:  pos,
		Velocity:  vel,
	}
	cp.stateHistory = nil
	cp.currentSeq = 0
}

// GetPredictionError calculates the prediction error between predicted and actual position
func (cp *ClientPredictor) GetPredictionError(actualPos Position) float64 {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	dx := actualPos.X - cp.currentState.Position.X
	dy := actualPos.Y - cp.currentState.Position.Y
	return sqrt(dx*dx + dy*dy)
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	// Simple implementation for square root
	if x == 0 {
		return 0
	}
	if x < 0 {
		return 0 // Invalid input
	}

	// Newton's method
	guess := x
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}
