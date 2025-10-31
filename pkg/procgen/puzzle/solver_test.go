package puzzle

import (
	"fmt"
	"testing"
)

func TestNewCSP(t *testing.T) {
	csp := NewCSP(12345)

	if csp == nil {
		t.Fatal("NewCSP returned nil")
	}
	if csp.Variables == nil {
		t.Error("Variables map is nil")
	}
	if csp.Constraints == nil {
		t.Error("Constraints slice is nil")
	}
	if csp.rng == nil {
		t.Error("RNG is nil")
	}
}

func TestCSPAddVariable(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		domain  []interface{}
		wantErr bool
	}{
		{"valid_variable", "var1", []interface{}{1, 2, 3}, false},
		{"valid_empty_domain", "var2", []interface{}{}, false},
		{"valid_string_domain", "var3", []interface{}{"a", "b", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csp := NewCSP(12345)
			err := csp.AddVariable(tt.varName, tt.domain)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddVariable() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				variable, exists := csp.Variables[tt.varName]
				if !exists {
					t.Errorf("Variable %s not added", tt.varName)
				}
				if variable.Name != tt.varName {
					t.Errorf("Variable name = %s, want %s", variable.Name, tt.varName)
				}
				if len(variable.Domain) != len(tt.domain) {
					t.Errorf("Domain size = %d, want %d", len(variable.Domain), len(tt.domain))
				}
			}
		})
	}
}

func TestCSPAddVariableDuplicate(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2})

	err := csp.AddVariable("var1", []interface{}{3, 4})
	if err == nil {
		t.Error("Expected error when adding duplicate variable")
	}
}

func TestCSPAddConstraint(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2})
	csp.AddVariable("var2", []interface{}{3, 4})

	constraint := func(assignments map[string]interface{}) bool {
		return true
	}

	err := csp.AddConstraint([]string{"var1", "var2"}, constraint)
	if err != nil {
		t.Errorf("AddConstraint() error = %v, want nil", err)
	}

	if len(csp.Constraints) != 1 {
		t.Errorf("Constraints count = %d, want 1", len(csp.Constraints))
	}
}

func TestCSPAddConstraintMissingVariable(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2})

	constraint := func(assignments map[string]interface{}) bool {
		return true
	}

	err := csp.AddConstraint([]string{"var1", "missing"}, constraint)
	if err == nil {
		t.Error("Expected error when adding constraint with missing variable")
	}
}

func TestCSPSolveSimple(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2, 3})
	csp.AddVariable("var2", []interface{}{4, 5, 6})

	// No constraints - any assignment should work
	solution, err := csp.Solve()

	if err != nil {
		t.Errorf("Solve() error = %v, want nil", err)
	}
	if solution == nil {
		t.Fatal("Solution is nil")
	}
	if len(solution) != 2 {
		t.Errorf("Solution size = %d, want 2", len(solution))
	}
}

func TestCSPSolveWithConstraint(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2, 3})
	csp.AddVariable("var2", []interface{}{1, 2, 3})

	// Constraint: var1 must be less than var2
	csp.AddConstraint([]string{"var1", "var2"}, func(assignments map[string]interface{}) bool {
		v1, ok1 := assignments["var1"]
		v2, ok2 := assignments["var2"]

		if !ok1 || !ok2 {
			return true
		}

		val1, _ := v1.(int)
		val2, _ := v2.(int)

		return val1 < val2
	})

	solution, err := csp.Solve()

	if err != nil {
		t.Errorf("Solve() error = %v, want nil", err)
	}
	if solution == nil {
		t.Fatal("Solution is nil")
	}

	// Verify constraint holds
	v1 := solution["var1"].(int)
	v2 := solution["var2"].(int)

	if v1 >= v2 {
		t.Errorf("Constraint violated: %d >= %d", v1, v2)
	}
}

func TestCSPSolveUnsolvable(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1})
	csp.AddVariable("var2", []interface{}{1})

	// Impossible constraint: both must equal 1 but also must be different
	csp.AddConstraint([]string{"var1", "var2"}, func(assignments map[string]interface{}) bool {
		v1, ok1 := assignments["var1"]
		v2, ok2 := assignments["var2"]

		if !ok1 || !ok2 {
			return true
		}

		return v1 != v2
	})

	_, err := csp.Solve()

	if err == nil {
		t.Error("Expected error for unsolvable CSP")
	}
}

func TestCSPReset(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2})

	// Manually assign value
	csp.Variables["var1"].Value = 1

	csp.Reset()

	if csp.Variables["var1"].Value != nil {
		t.Error("Reset() did not clear variable value")
	}
}

func TestCSPGetVariableNames(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1})
	csp.AddVariable("var2", []interface{}{2})
	csp.AddVariable("var3", []interface{}{3})

	names := csp.GetVariableNames()

	if len(names) != 3 {
		t.Errorf("GetVariableNames() returned %d names, want 3", len(names))
	}

	// Check all names are present
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	for _, expected := range []string{"var1", "var2", "var3"} {
		if !found[expected] {
			t.Errorf("Variable %s not in returned names", expected)
		}
	}
}

func TestCSPGetDomainSize(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1, 2, 3})

	size := csp.GetDomainSize("var1")
	if size != 3 {
		t.Errorf("GetDomainSize(var1) = %d, want 3", size)
	}

	size = csp.GetDomainSize("missing")
	if size != 0 {
		t.Errorf("GetDomainSize(missing) = %d, want 0", size)
	}
}

func TestCSPGetConstraintCount(t *testing.T) {
	csp := NewCSP(12345)
	csp.AddVariable("var1", []interface{}{1})
	csp.AddVariable("var2", []interface{}{2})

	if csp.GetConstraintCount() != 0 {
		t.Errorf("Initial constraint count = %d, want 0", csp.GetConstraintCount())
	}

	csp.AddConstraint([]string{"var1", "var2"}, func(m map[string]interface{}) bool { return true })

	if csp.GetConstraintCount() != 1 {
		t.Errorf("After AddConstraint count = %d, want 1", csp.GetConstraintCount())
	}
}

// PuzzleSolver Tests

func TestNewPuzzleSolver(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	if solver == nil {
		t.Fatal("NewPuzzleSolver returned nil")
	}
	if solver.csp == nil {
		t.Error("Solver CSP is nil")
	}
}

func TestPuzzleSolverAddElement(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	err := solver.AddElement("plate1", []int{0, 1, 2})
	if err != nil {
		t.Errorf("AddElement() error = %v, want nil", err)
	}

	// Verify element was added to CSP
	if solver.csp.GetDomainSize("plate1") != 3 {
		t.Errorf("Element domain size = %d, want 3", solver.csp.GetDomainSize("plate1"))
	}
}

func TestPuzzleSolverAddSequenceConstraint(t *testing.T) {
	tests := []struct {
		name     string
		sequence []string
		wantErr  bool
	}{
		{"valid_sequence", []string{"plate1", "plate2", "plate3"}, false},
		{"two_element_sequence", []string{"plate1", "plate2"}, false},
		{"single_element_error", []string{"plate1"}, true},
		{"empty_sequence_error", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solver := NewPuzzleSolver(12345)
			for _, elem := range tt.sequence {
				solver.AddElement(elem, []int{1, 2, 3, 4, 5})
			}

			err := solver.AddSequenceConstraint(tt.sequence)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddSequenceConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPuzzleSolverSequenceConstraintSolution(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	solver.AddElement("first", []int{1, 2, 3})
	solver.AddElement("second", []int{1, 2, 3})
	solver.AddElement("third", []int{1, 2, 3})

	err := solver.AddSequenceConstraint([]string{"first", "second", "third"})
	if err != nil {
		t.Fatalf("AddSequenceConstraint() error = %v", err)
	}

	solution, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve() error = %v", err)
	}

	// Verify sequence order
	if solution["first"] >= solution["second"] {
		t.Errorf("Sequence violated: first(%d) >= second(%d)", solution["first"], solution["second"])
	}
	if solution["second"] >= solution["third"] {
		t.Errorf("Sequence violated: second(%d) >= third(%d)", solution["second"], solution["third"])
	}
}

func TestPuzzleSolverAddUniquenessConstraint(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
		wantErr  bool
	}{
		{"valid_uniqueness", []string{"plate1", "plate2", "plate3"}, false},
		{"two_elements", []string{"plate1", "plate2"}, false},
		{"single_element_error", []string{"plate1"}, true},
		{"empty_error", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solver := NewPuzzleSolver(12345)
			for _, elem := range tt.elements {
				solver.AddElement(elem, []int{1, 2, 3, 4, 5})
			}

			err := solver.AddUniquenessConstraint(tt.elements)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddUniquenessConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPuzzleSolverUniquenessConstraintSolution(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	solver.AddElement("plate1", []int{1, 2, 3})
	solver.AddElement("plate2", []int{1, 2, 3})
	solver.AddElement("plate3", []int{1, 2, 3})

	err := solver.AddUniquenessConstraint([]string{"plate1", "plate2", "plate3"})
	if err != nil {
		t.Fatalf("AddUniquenessConstraint() error = %v", err)
	}

	solution, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve() error = %v", err)
	}

	// Verify all values are unique
	values := []int{solution["plate1"], solution["plate2"], solution["plate3"]}
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] == values[j] {
				t.Errorf("Uniqueness violated: values[%d] == values[%d] == %d", i, j, values[i])
			}
		}
	}
}

func TestPuzzleSolverAddSumConstraint(t *testing.T) {
	tests := []struct {
		name      string
		elements  []string
		targetSum int
		wantErr   bool
	}{
		{"valid_sum", []string{"plate1", "plate2"}, 5, false},
		{"empty_error", []string{}, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solver := NewPuzzleSolver(12345)
			for _, elem := range tt.elements {
				solver.AddElement(elem, []int{1, 2, 3, 4, 5})
			}

			err := solver.AddSumConstraint(tt.elements, tt.targetSum)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddSumConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPuzzleSolverSumConstraintSolution(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	solver.AddElement("plate1", []int{1, 2, 3})
	solver.AddElement("plate2", []int{1, 2, 3})

	targetSum := 4

	err := solver.AddSumConstraint([]string{"plate1", "plate2"}, targetSum)
	if err != nil {
		t.Fatalf("AddSumConstraint() error = %v", err)
	}

	solution, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve() error = %v", err)
	}

	// Verify sum
	actualSum := solution["plate1"] + solution["plate2"]
	if actualSum != targetSum {
		t.Errorf("Sum = %d, want %d", actualSum, targetSum)
	}
}

func TestPuzzleSolverComplexConstraints(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	// Create 4 elements
	for i := 1; i <= 4; i++ {
		solver.AddElement(fmt.Sprintf("plate%d", i), []int{1, 2, 3, 4})
	}

	// Add sequence constraint: plate1 < plate2 < plate3 < plate4
	err := solver.AddSequenceConstraint([]string{"plate1", "plate2", "plate3", "plate4"})
	if err != nil {
		t.Fatalf("AddSequenceConstraint() error = %v", err)
	}

	solution, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve() error = %v", err)
	}

	// Verify sequence
	if solution["plate1"] >= solution["plate2"] ||
		solution["plate2"] >= solution["plate3"] ||
		solution["plate3"] >= solution["plate4"] {
		t.Errorf("Sequence constraint violated in solution: %v", solution)
	}
}

func TestPuzzleSolverGetCSP(t *testing.T) {
	solver := NewPuzzleSolver(12345)

	csp := solver.GetCSP()

	if csp == nil {
		t.Error("GetCSP() returned nil")
	}
	if csp != solver.csp {
		t.Error("GetCSP() did not return internal CSP")
	}
}
