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

	// Error threshold for reconciliation (higher for high-latency)
	errorThreshold float64
}

// NewClientPredictor creates a new client-side predictor.
// Uses 256 states at 20Hz = 12.8 seconds of history, adequate for normal latency.
// For high-latency mode, use NewHighLatencyClientPredictor().
func NewClientPredictor() *ClientPredictor {
	return &ClientPredictor{
		stateHistory:   make([]PredictedState, 0, 256),
		maxHistory:     256, // Keep last 256 states (12.8 seconds at 20Hz)
		currentSeq:     0,
		errorThreshold: 1.0, // Standard 1-unit prediction error tolerance
	}
}

// NewHighLatencyClientPredictor creates a client-side predictor optimized for high-latency scenarios.
// Designed for Tor/onion services and connections with 200-5000ms latency.
// Uses larger history buffer (512 states = 25.6s at 20Hz) and relaxed error threshold
// to handle delayed server reconciliation and prediction corrections.
func NewHighLatencyClientPredictor() *ClientPredictor {
	return &ClientPredictor{
		stateHistory:   make([]PredictedState, 0, 512),
		maxHistory:     512, // Keep last 512 states (25.6 seconds at 20Hz)
		currentSeq:     0,
		errorThreshold: 5.0, // Relaxed 5-unit tolerance for high-latency prediction errors
	}
}

// PredictInput predicts the result of applying an input and stores it for reconciliation
func (cp *ClientPredictor) PredictInput(dx, dy, deltaTime float64) PredictedState {
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

	stateIndex := findStateIndex(cp.stateHistory, serverSeq)
	if stateIndex == -1 {
		return cp.resetToServerState(serverSeq, serverPos, serverVel)
	}

	if isPredictionAccurate(cp.stateHistory[stateIndex], serverPos, cp.errorThreshold) {
		cp.stateHistory = cp.stateHistory[stateIndex+1:]
		return cp.currentState
	}

	return cp.correctAndReplay(stateIndex, serverSeq, serverPos, serverVel)
}

// findStateIndex locates the state matching the server sequence.
func findStateIndex(history []PredictedState, serverSeq uint32) int {
	for i, state := range history {
		if state.Sequence == serverSeq {
			return i
		}
	}
	return -1
}

// resetToServerState resets client state to match server when history is lost.
func (cp *ClientPredictor) resetToServerState(serverSeq uint32, serverPos Position, serverVel Velocity) PredictedState {
	cp.currentState = PredictedState{
		Sequence:  serverSeq,
		Timestamp: time.Now(),
		Position:  serverPos,
		Velocity:  serverVel,
	}
	cp.stateHistory = make([]PredictedState, 0, cp.maxHistory)
	return cp.currentState
}

// isPredictionAccurate checks if prediction error is within threshold.
func isPredictionAccurate(predicted PredictedState, serverPos Position, threshold float64) bool {
	errorX := serverPos.X - predicted.Position.X
	errorY := serverPos.Y - predicted.Position.Y
	return abs(errorX) < threshold && abs(errorY) < threshold
}

// correctAndReplay corrects prediction error and replays subsequent inputs.
func (cp *ClientPredictor) correctAndReplay(stateIndex int, serverSeq uint32, serverPos Position, serverVel Velocity) PredictedState {
	correctedState := PredictedState{
		Sequence:  serverSeq,
		Timestamp: time.Now(),
		Position:  serverPos,
		Velocity:  serverVel,
	}

	inputsToReplay := cp.stateHistory[stateIndex+1:]
	predicted := cp.stateHistory[stateIndex]

	for i, oldState := range inputsToReplay {
		deltaTime := calculateReplayDelta(i, oldState, inputsToReplay, predicted)
		applyInputReplay(&correctedState, oldState, deltaTime)
	}

	cp.stateHistory = inputsToReplay
	cp.currentState = correctedState
	return correctedState
}

// calculateReplayDelta computes delta time for input replay.
func calculateReplayDelta(i int, oldState PredictedState, inputsToReplay []PredictedState, predicted PredictedState) float64 {
	var deltaTime float64
	if i > 0 {
		deltaTime = oldState.Timestamp.Sub(inputsToReplay[i-1].Timestamp).Seconds()
	} else {
		deltaTime = oldState.Timestamp.Sub(predicted.Timestamp).Seconds()
	}

	if deltaTime > 0.1 {
		deltaTime = 0.1
	}
	if deltaTime < 0 {
		deltaTime = 0.05
	}
	return deltaTime
}

// applyInputReplay applies a single input during state replay.
func applyInputReplay(corrected *PredictedState, oldState PredictedState, deltaTime float64) {
	dx := oldState.Velocity.VX - corrected.Velocity.VX
	dy := oldState.Velocity.VY - corrected.Velocity.VY

	corrected.Velocity.VX += dx
	corrected.Velocity.VY += dy
	corrected.Position.X += corrected.Velocity.VX * deltaTime
	corrected.Position.Y += corrected.Velocity.VY * deltaTime
	corrected.Sequence = oldState.Sequence
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
