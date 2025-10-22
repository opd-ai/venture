package network

import (
	"math"
	"testing"
	"time"
)

func TestNewClientPredictor(t *testing.T) {
	predictor := NewClientPredictor()

	if predictor == nil {
		t.Fatal("NewClientPredictor returned nil")
	}

	if predictor.maxHistory != 128 {
		t.Errorf("Expected maxHistory 128, got %d", predictor.maxHistory)
	}

	if len(predictor.stateHistory) != 0 {
		t.Errorf("Expected empty stateHistory, got length %d", len(predictor.stateHistory))
	}
}

func TestClientPredictor_SetInitialState(t *testing.T) {
	predictor := NewClientPredictor()

	pos := Position{X: 100, Y: 200}
	vel := Velocity{VX: 10, VY: 20}

	predictor.SetInitialState(pos, vel)

	state := predictor.GetCurrentState()

	if state.Position.X != 100 || state.Position.Y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", state.Position.X, state.Position.Y)
	}

	if state.Velocity.VX != 10 || state.Velocity.VY != 20 {
		t.Errorf("Expected velocity (10, 20), got (%f, %f)", state.Velocity.VX, state.Velocity.VY)
	}

	if state.Sequence != 0 {
		t.Errorf("Expected sequence 0, got %d", state.Sequence)
	}
}

func TestClientPredictor_PredictInput(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	// Predict movement to the right
	deltaTime := 0.1
	predicted := predictor.PredictInput(100, 0, deltaTime)

	if predicted.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", predicted.Sequence)
	}

	// Velocity should increase
	if predicted.Velocity.VX != 10 {
		t.Errorf("Expected VX 10, got %f", predicted.Velocity.VX)
	}

	// Position should move
	if predicted.Position.X != 1.0 {
		t.Errorf("Expected X 1.0, got %f", predicted.Position.X)
	}

	// History should contain the prediction
	if len(predictor.stateHistory) != 1 {
		t.Errorf("Expected history length 1, got %d", len(predictor.stateHistory))
	}
}

func TestClientPredictor_PredictInput_Multiple(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	deltaTime := 0.05

	// Make several predictions
	for i := 0; i < 5; i++ {
		predictor.PredictInput(100, 0, deltaTime)
	}

	state := predictor.GetCurrentState()

	if state.Sequence != 5 {
		t.Errorf("Expected sequence 5, got %d", state.Sequence)
	}

	if len(predictor.stateHistory) != 5 {
		t.Errorf("Expected history length 5, got %d", len(predictor.stateHistory))
	}

	// Position should have accumulated
	if state.Position.X <= 0 {
		t.Errorf("Expected X > 0, got %f", state.Position.X)
	}
}

func TestClientPredictor_ReconcileServerState_NoError(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	deltaTime := 0.05

	// Make 3 predictions
	for i := 0; i < 3; i++ {
		predictor.PredictInput(100, 0, deltaTime)
	}

	// Server acknowledges sequence 2 with same position
	state2 := predictor.stateHistory[1]
	corrected := predictor.ReconcileServerState(2, state2.Position, state2.Velocity)

	// Since there's no error, should keep current prediction
	if corrected.Sequence != 3 {
		t.Errorf("Expected sequence 3 after reconciliation, got %d", corrected.Sequence)
	}

	// History should be trimmed
	if len(predictor.stateHistory) != 1 {
		t.Errorf("Expected history length 1 after reconciliation, got %d", len(predictor.stateHistory))
	}
}

func TestClientPredictor_ReconcileServerState_WithError(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	deltaTime := 0.05

	// Make 5 predictions
	for i := 0; i < 5; i++ {
		predictor.PredictInput(100, 0, deltaTime)
	}

	// Server acknowledges sequence 2 but with different position (prediction error)
	serverPos := Position{X: 0, Y: 0} // Server says we didn't move much
	serverVel := Velocity{VX: 5, VY: 0}

	corrected := predictor.ReconcileServerState(2, serverPos, serverVel)

	// Should replay inputs 3, 4, 5 from the corrected position
	if corrected.Sequence != 5 {
		t.Errorf("Expected sequence 5 after reconciliation, got %d", corrected.Sequence)
	}

	// Position should be corrected but still ahead of server's ack
	if corrected.Position.X <= 0 {
		t.Errorf("Expected X > 0 after replay, got %f", corrected.Position.X)
	}

	// History should only contain replayed states
	if len(predictor.stateHistory) != 3 {
		t.Errorf("Expected history length 3 (states 3,4,5), got %d", len(predictor.stateHistory))
	}
}

func TestClientPredictor_ReconcileServerState_OldSequence(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	deltaTime := 0.05

	// Make 10 predictions
	for i := 0; i < 10; i++ {
		predictor.PredictInput(100, 0, deltaTime)
	}

	// Server acknowledges an old sequence that's no longer in history
	serverPos := Position{X: 50, Y: 0}
	serverVel := Velocity{VX: 10, VY: 0}

	corrected := predictor.ReconcileServerState(1, serverPos, serverVel)

	// Should trust server completely and reset
	if abs(corrected.Position.X-50) > 1.0 || abs(corrected.Position.Y-0) > 1.0 {
		t.Errorf("Expected position close to (50, 0), got (%f, %f)", corrected.Position.X, corrected.Position.Y)
	}

	// History should be cleared or nearly empty
	if len(predictor.stateHistory) > 10 {
		t.Errorf("Expected small or empty history, got length %d", len(predictor.stateHistory))
	}
}

func TestClientPredictor_GetPredictionError(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	// Move to position
	// VX = 0 + 30*0.1 = 3, X = 0 + 3*0.1 = 0.3
	// VY = 0 + 40*0.1 = 4, Y = 0 + 4*0.1 = 0.4
	predictor.PredictInput(30, 40, 0.1)

	actualPos := Position{X: 4, Y: 5}
	error := predictor.GetPredictionError(actualPos)

	// Error should be sqrt((4-0.3)^2 + (5-0.4)^2) = sqrt(3.7^2 + 4.6^2) ≈ 5.9
	expected := math.Sqrt(3.7*3.7 + 4.6*4.6)
	if math.Abs(error-expected) > 0.1 {
		t.Errorf("Expected error ≈%f, got %f", expected, error)
	}
}

func TestClientPredictor_ConcurrentAccess(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	done := make(chan bool, 3)

	// Concurrent predictions
	go func() {
		for i := 0; i < 100; i++ {
			predictor.PredictInput(10, 0, 0.01)
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Concurrent reconciliations
	go func() {
		for i := 0; i < 50; i++ {
			predictor.ReconcileServerState(uint32(i), Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})
			time.Sleep(2 * time.Microsecond)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 200; i++ {
			predictor.GetCurrentState()
			time.Sleep(time.Microsecond / 2)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should not panic and state should be valid
	state := predictor.GetCurrentState()
	if state.Sequence < 0 {
		t.Errorf("Invalid sequence after concurrent access: %d", state.Sequence)
	}
}

func TestClientPredictor_HistoryTrimming(t *testing.T) {
	predictor := NewClientPredictor()
	predictor.maxHistory = 10 // Set small max for testing
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	// Add more states than max
	for i := 0; i < 20; i++ {
		predictor.PredictInput(10, 0, 0.01)
	}

	// History should be trimmed to max
	if len(predictor.stateHistory) != 10 {
		t.Errorf("Expected history length %d, got %d", predictor.maxHistory, len(predictor.stateHistory))
	}

	// Latest sequences should be preserved
	firstSeq := predictor.stateHistory[0].Sequence
	if firstSeq != 11 { // Should have trimmed sequences 1-10
		t.Errorf("Expected first sequence 11 after trimming, got %d", firstSeq)
	}
}

func TestPosition_Struct(t *testing.T) {
	pos := Position{X: 123.45, Y: 678.90}

	if pos.X != 123.45 {
		t.Errorf("Expected X 123.45, got %f", pos.X)
	}

	if pos.Y != 678.90 {
		t.Errorf("Expected Y 678.90, got %f", pos.Y)
	}
}

func TestVelocity_Struct(t *testing.T) {
	vel := Velocity{VX: -12.34, VY: 56.78}

	if vel.VX != -12.34 {
		t.Errorf("Expected VX -12.34, got %f", vel.VX)
	}

	if vel.VY != 56.78 {
		t.Errorf("Expected VY 56.78, got %f", vel.VY)
	}
}

func TestPredictedState_Struct(t *testing.T) {
	now := time.Now()
	state := PredictedState{
		Sequence:  42,
		Timestamp: now,
		Position:  Position{X: 10, Y: 20},
		Velocity:  Velocity{VX: 1, VY: 2},
	}

	if state.Sequence != 42 {
		t.Errorf("Expected sequence 42, got %d", state.Sequence)
	}

	if state.Position.X != 10 {
		t.Errorf("Expected position X 10, got %f", state.Position.X)
	}

	if state.Velocity.VX != 1 {
		t.Errorf("Expected velocity VX 1, got %f", state.Velocity.VX)
	}
}

func TestHelperFunctions_Abs(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.001, 0.001},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%f) = %f, expected %f", tt.input, result, tt.expected)
		}
	}
}

func TestHelperFunctions_Sqrt(t *testing.T) {
	tests := []struct {
		input     float64
		expected  float64
		tolerance float64
	}{
		{0.0, 0.0, 0.001},
		{1.0, 1.0, 0.001},
		{4.0, 2.0, 0.001},
		{9.0, 3.0, 0.001},
		{16.0, 4.0, 0.001},
		{2.0, math.Sqrt(2), 0.001},
	}

	for _, tt := range tests {
		result := sqrt(tt.input)
		if math.Abs(result-tt.expected) > tt.tolerance {
			t.Errorf("sqrt(%f) = %f, expected %f (tolerance %f)",
				tt.input, result, tt.expected, tt.tolerance)
		}
	}
}

// Benchmark prediction operations
func BenchmarkClientPredictor_PredictInput(b *testing.B) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		predictor.PredictInput(100, 50, 0.016)
	}
}

func BenchmarkClientPredictor_ReconcileServerState(b *testing.B) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	// Add some history
	for i := 0; i < 10; i++ {
		predictor.PredictInput(100, 50, 0.016)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		predictor.ReconcileServerState(5, Position{X: 1, Y: 1}, Velocity{VX: 10, VY: 5})
	}
}

func BenchmarkClientPredictor_GetCurrentState(b *testing.B) {
	predictor := NewClientPredictor()
	predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})
	predictor.PredictInput(100, 50, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		predictor.GetCurrentState()
	}
}
