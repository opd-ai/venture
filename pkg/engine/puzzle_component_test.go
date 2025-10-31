package engine

import (
	"math"
	"testing"
	"time"
)

func TestPuzzleStateString(t *testing.T) {
	tests := []struct {
		state    PuzzleState
		expected string
	}{
		{PuzzleStateUnsolved, "Unsolved"},
		{PuzzleStateSolving, "Solving"},
		{PuzzleStateSolved, "Solved"},
		{PuzzleStateFailed, "Failed"},
		{PuzzleState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNewPuzzleComponent(t *testing.T) {
	tests := []struct {
		name       string
		puzzleID   string
		puzzleType PuzzleType
		difficulty int
	}{
		{"pressure_plate_easy", "puzzle_001", PuzzleTypePressurePlate, 3},
		{"lever_sequence_hard", "puzzle_002", PuzzleTypeLeverSequence, 8},
		{"block_pushing_medium", "puzzle_003", PuzzleTypeBlockPushing, 5},
		{"timed_challenge", "puzzle_004", PuzzleTypeTimedChallenge, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzle := NewPuzzleComponent(tt.puzzleID, tt.puzzleType, tt.difficulty)

			if puzzle.PuzzleID != tt.puzzleID {
				t.Errorf("Expected PuzzleID %s, got %s", tt.puzzleID, puzzle.PuzzleID)
			}
			if puzzle.PuzzleType != tt.puzzleType {
				t.Errorf("Expected Type %s, got %s", tt.puzzleType, puzzle.PuzzleType)
			}
			if puzzle.Difficulty != tt.difficulty {
				t.Errorf("Expected Difficulty %d, got %d", tt.difficulty, puzzle.Difficulty)
			}
			if puzzle.State != PuzzleStateUnsolved {
				t.Errorf("Expected initial state Unsolved, got %v", puzzle.State)
			}
			if len(puzzle.Solution) != 0 {
				t.Errorf("Expected empty solution, got %d elements", len(puzzle.Solution))
			}
		})
	}
}

func TestPuzzleComponentType(t *testing.T) {
	puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)
	if puzzle.Type() != "puzzle" {
		t.Errorf("Expected component type 'puzzle', got '%s'", puzzle.Type())
	}
}

func TestPuzzleAddElement(t *testing.T) {
	puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)

	puzzle.AddElement(101, "plate1")
	puzzle.AddElement(102, "plate2")
	puzzle.AddElement(103, "plate3")

	if len(puzzle.ElementIDs) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(puzzle.ElementIDs))
	}

	expectedIDs := []uint64{101, 102, 103}
	for i, expected := range expectedIDs {
		if puzzle.ElementIDs[i] != expected {
			t.Errorf("Element %d: expected ID %d, got %d", i, expected, puzzle.ElementIDs[i])
		}
	}
}

func TestPuzzleRecordProgress(t *testing.T) {
	puzzle := NewPuzzleComponent("test", PuzzleTypeLeverSequence, 5)

	puzzle.RecordProgress("lever1")
	puzzle.RecordProgress("lever3")
	puzzle.RecordProgress("lever2")

	if len(puzzle.CurrentProgress) != 3 {
		t.Errorf("Expected 3 progress items, got %d", len(puzzle.CurrentProgress))
	}

	expected := []string{"lever1", "lever3", "lever2"}
	for i, exp := range expected {
		if puzzle.CurrentProgress[i] != exp {
			t.Errorf("Progress %d: expected %s, got %s", i, exp, puzzle.CurrentProgress[i])
		}
	}
}

func TestPuzzleCheckSolution(t *testing.T) {
	tests := []struct {
		name     string
		solution []string
		progress []string
		expected bool
	}{
		{
			name:     "correct_solution",
			solution: []string{"A", "B", "C"},
			progress: []string{"A", "B", "C"},
			expected: true,
		},
		{
			name:     "wrong_order",
			solution: []string{"A", "B", "C"},
			progress: []string{"A", "C", "B"},
			expected: false,
		},
		{
			name:     "incomplete",
			solution: []string{"A", "B", "C"},
			progress: []string{"A", "B"},
			expected: false,
		},
		{
			name:     "too_many",
			solution: []string{"A", "B"},
			progress: []string{"A", "B", "C"},
			expected: false,
		},
		{
			name:     "empty_solution",
			solution: []string{},
			progress: []string{},
			expected: true,
		},
		{
			name:     "wrong_elements",
			solution: []string{"A", "B", "C"},
			progress: []string{"X", "Y", "Z"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzle := NewPuzzleComponent("test", PuzzleTypeLeverSequence, 5)
			puzzle.Solution = tt.solution
			puzzle.CurrentProgress = tt.progress

			result := puzzle.CheckSolution()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPuzzleIsSolved(t *testing.T) {
	tests := []struct {
		name     string
		state    PuzzleState
		expected bool
	}{
		{"unsolved", PuzzleStateUnsolved, false},
		{"solving", PuzzleStateSolving, false},
		{"solved", PuzzleStateSolved, true},
		{"failed", PuzzleStateFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)
			puzzle.State = tt.state

			result := puzzle.IsSolved()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPuzzleMarkSolved(t *testing.T) {
	puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)
	solverID := uint64(42)

	beforeTime := time.Now()
	puzzle.MarkSolved(solverID)
	afterTime := time.Now()

	if puzzle.State != PuzzleStateSolved {
		t.Errorf("Expected state Solved, got %v", puzzle.State)
	}
	if puzzle.SolvedBy != solverID {
		t.Errorf("Expected SolvedBy %d, got %d", solverID, puzzle.SolvedBy)
	}
	if puzzle.SolvedAt.Before(beforeTime) || puzzle.SolvedAt.After(afterTime) {
		t.Errorf("SolvedAt time not within expected range")
	}
}

func TestPuzzleReset(t *testing.T) {
	puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)
	puzzle.State = PuzzleStateSolving
	puzzle.CurrentProgress = []string{"A", "B"}
	puzzle.TimeElapsed = 10.5
	puzzle.Attempts = 2

	puzzle.Reset()

	if puzzle.State != PuzzleStateUnsolved {
		t.Errorf("Expected state Unsolved after reset, got %v", puzzle.State)
	}
	if len(puzzle.CurrentProgress) != 0 {
		t.Errorf("Expected empty progress after reset, got %d items", len(puzzle.CurrentProgress))
	}
	if puzzle.TimeElapsed != 0 {
		t.Errorf("Expected TimeElapsed 0 after reset, got %f", puzzle.TimeElapsed)
	}
	if puzzle.Attempts != 3 {
		t.Errorf("Expected Attempts incremented to 3, got %d", puzzle.Attempts)
	}
}

func TestPuzzleHasTimedOut(t *testing.T) {
	tests := []struct {
		name        string
		timeLimit   float64
		timeElapsed float64
		expected    bool
	}{
		{"no_limit", 0, 100, false},
		{"within_limit", 60, 30, false},
		{"at_limit", 60, 60, true},
		{"exceeded", 60, 65, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzle := NewPuzzleComponent("test", PuzzleTypeTimedChallenge, 5)
			puzzle.TimeLimit = tt.timeLimit
			puzzle.TimeElapsed = tt.timeElapsed

			result := puzzle.HasTimedOut()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPuzzleHasMaxAttemptsReached(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		attempts    int
		expected    bool
	}{
		{"unlimited", 0, 100, false},
		{"within_limit", 3, 2, false},
		{"at_limit", 3, 3, true},
		{"exceeded", 3, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)
			puzzle.MaxAttempts = tt.maxAttempts
			puzzle.Attempts = tt.attempts

			result := puzzle.HasMaxAttemptsReached()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPuzzleGetProgressPercent(t *testing.T) {
	const epsilon = 1e-9 // Tolerance for floating point comparison

	tests := []struct {
		name            string
		solution        []string
		progress        []string
		expectedPercent float64
	}{
		{"empty", []string{}, []string{}, 0},
		{"zero_progress", []string{"A", "B", "C"}, []string{}, 0},
		{"one_third", []string{"A", "B", "C"}, []string{"A"}, 33.333333333333336},
		{"two_thirds", []string{"A", "B", "C"}, []string{"A", "B"}, 66.66666666666667},
		{"complete", []string{"A", "B", "C"}, []string{"A", "B", "C"}, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzle := NewPuzzleComponent("test", PuzzleTypeLeverSequence, 5)
			puzzle.Solution = tt.solution
			puzzle.CurrentProgress = tt.progress

			result := puzzle.GetProgressPercent()
			if math.Abs(result-tt.expectedPercent) > epsilon {
				t.Errorf("Expected %.2f%%, got %.2f%%", tt.expectedPercent, result)
			}
		})
	}
}

func TestPuzzleGetNextHint(t *testing.T) {
	puzzle := NewPuzzleComponent("test", PuzzleTypePressurePlate, 5)
	puzzle.Hints = []string{"Hint 1", "Hint 2", "Hint 3"}

	// Get first hint
	hint, err := puzzle.GetNextHint()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hint != "Hint 1" {
		t.Errorf("Expected 'Hint 1', got '%s'", hint)
	}
	if puzzle.HintsUsed != 1 {
		t.Errorf("Expected HintsUsed 1, got %d", puzzle.HintsUsed)
	}

	// Get second hint
	hint, err = puzzle.GetNextHint()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hint != "Hint 2" {
		t.Errorf("Expected 'Hint 2', got '%s'", hint)
	}

	// Get third hint
	hint, err = puzzle.GetNextHint()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hint != "Hint 3" {
		t.Errorf("Expected 'Hint 3', got '%s'", hint)
	}

	// Try to get fourth hint (should fail)
	_, err = puzzle.GetNextHint()
	if err == nil {
		t.Error("Expected error when no hints remaining, got nil")
	}
}

// PuzzleElementComponent tests

func TestNewPuzzleElementComponent(t *testing.T) {
	element := NewPuzzleElementComponent("plate1", "puzzle_001", "pressure_plate")

	if element.ElementName != "plate1" {
		t.Errorf("Expected ElementName 'plate1', got '%s'", element.ElementName)
	}
	if element.PuzzleID != "puzzle_001" {
		t.Errorf("Expected PuzzleID 'puzzle_001', got '%s'", element.PuzzleID)
	}
	if element.ElementType != "pressure_plate" {
		t.Errorf("Expected ElementType 'pressure_plate', got '%s'", element.ElementType)
	}
	if element.State != 0 {
		t.Errorf("Expected initial State 0, got %d", element.State)
	}
	if !element.IsInteractable {
		t.Error("Expected IsInteractable true by default")
	}
}

func TestPuzzleElementComponentType(t *testing.T) {
	element := NewPuzzleElementComponent("test", "puzzle_001", "lever")
	if element.Type() != "puzzleElement" {
		t.Errorf("Expected component type 'puzzleElement', got '%s'", element.Type())
	}
}

func TestPuzzleElementActivate(t *testing.T) {
	tests := []struct {
		name            string
		isInteractable  bool
		cooldownTime    float64
		cooldownElapsed float64
		initialState    int
		expectedState   int
		shouldActivate  bool
	}{
		{"normal_activation", true, 0.5, 0.5, 0, 1, true},
		{"toggle_back", true, 0.5, 0.5, 1, 0, true},
		{"on_cooldown", true, 0.5, 0.2, 0, 0, false},
		{"not_interactable", false, 0.5, 0.5, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := NewPuzzleElementComponent("test", "puzzle_001", "lever")
			element.IsInteractable = tt.isInteractable
			element.CooldownTime = tt.cooldownTime
			element.CooldownElapsed = tt.cooldownElapsed
			element.State = tt.initialState

			result := element.Activate()
			if result != tt.shouldActivate {
				t.Errorf("Expected Activate() to return %v, got %v", tt.shouldActivate, result)
			}
			if element.State != tt.expectedState {
				t.Errorf("Expected State %d, got %d", tt.expectedState, element.State)
			}
			if tt.shouldActivate && element.CooldownElapsed != 0 {
				t.Errorf("Expected cooldown reset to 0, got %f", element.CooldownElapsed)
			}
		})
	}
}

func TestPuzzleElementSetState(t *testing.T) {
	element := NewPuzzleElementComponent("test", "puzzle_001", "rune")
	element.CooldownTime = 0.5
	element.CooldownElapsed = 0.5

	// Should succeed
	result := element.SetState(5)
	if !result {
		t.Error("Expected SetState to succeed")
	}
	if element.State != 5 {
		t.Errorf("Expected State 5, got %d", element.State)
	}

	// Should fail due to cooldown
	result = element.SetState(10)
	if result {
		t.Error("Expected SetState to fail due to cooldown")
	}
	if element.State != 5 {
		t.Errorf("Expected State to remain 5, got %d", element.State)
	}
}

func TestPuzzleElementIsInCorrectState(t *testing.T) {
	tests := []struct {
		name        string
		state       int
		targetState int
		expected    bool
	}{
		{"correct", 5, 5, true},
		{"incorrect", 3, 7, false},
		{"zero_correct", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := NewPuzzleElementComponent("test", "puzzle_001", "plate")
			element.State = tt.state
			element.TargetState = tt.targetState

			result := element.IsInCorrectState()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPuzzleElementIsOnCooldown(t *testing.T) {
	tests := []struct {
		name            string
		cooldownTime    float64
		cooldownElapsed float64
		expected        bool
	}{
		{"on_cooldown", 1.0, 0.5, true},
		{"cooldown_complete", 1.0, 1.0, false},
		{"cooldown_exceeded", 1.0, 1.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := NewPuzzleElementComponent("test", "puzzle_001", "lever")
			element.CooldownTime = tt.cooldownTime
			element.CooldownElapsed = tt.cooldownElapsed

			result := element.IsOnCooldown()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPuzzleElementUpdateCooldown(t *testing.T) {
	element := NewPuzzleElementComponent("test", "puzzle_001", "plate")
	element.CooldownTime = 1.0
	element.CooldownElapsed = 0.3

	element.UpdateCooldown(0.2)

	expected := 0.5
	if element.CooldownElapsed != expected {
		t.Errorf("Expected CooldownElapsed %.2f, got %.2f", expected, element.CooldownElapsed)
	}

	// Update beyond cooldown time
	element.UpdateCooldown(1.0)

	// Should cap at cooldown time (not exceed)
	if element.CooldownElapsed < element.CooldownTime {
		t.Errorf("Expected CooldownElapsed to complete, got %.2f", element.CooldownElapsed)
	}
}

func TestPuzzleElementReset(t *testing.T) {
	element := NewPuzzleElementComponent("test", "puzzle_001", "lever")
	element.State = 5
	element.CooldownElapsed = 0.7
	element.IsInteractable = false

	element.Reset()

	if element.State != 0 {
		t.Errorf("Expected State 0 after reset, got %d", element.State)
	}
	if element.CooldownElapsed != 0 {
		t.Errorf("Expected CooldownElapsed 0 after reset, got %f", element.CooldownElapsed)
	}
	if !element.IsInteractable {
		t.Error("Expected IsInteractable true after reset")
	}
}
