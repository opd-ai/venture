// Package puzzle provides procedural puzzle generation.
// Phase 11.2: Constraint Solver
//
// This file implements a Constraint Satisfaction Problem (CSP) solver
// using backtracking search to ensure puzzles are solvable.
package puzzle

import (
	"fmt"
	"math/rand"
)

// Variable represents a decision point in the puzzle.
type Variable struct {
	// Name uniquely identifies this variable
	Name string

	// Domain contains possible values for this variable
	Domain []interface{}

	// Value is the assigned value (nil if unassigned)
	Value interface{}
}

// Constraint represents a relationship between variables.
type Constraint struct {
	// Variables involved in this constraint
	Variables []string

	// IsSatisfied checks if the constraint is satisfied given assignments
	IsSatisfied func(assignments map[string]interface{}) bool
}

// CSP represents a Constraint Satisfaction Problem.
type CSP struct {
	// Variables in the problem
	Variables map[string]*Variable

	// Constraints on variables
	Constraints []*Constraint

	// Random number generator for variable ordering
	rng *rand.Rand
}

// NewCSP creates a new constraint satisfaction problem.
func NewCSP(seed int64) *CSP {
	return &CSP{
		Variables:   make(map[string]*Variable),
		Constraints: []*Constraint{},
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// AddVariable adds a variable to the CSP.
func (csp *CSP) AddVariable(name string, domain []interface{}) error {
	if _, exists := csp.Variables[name]; exists {
		return fmt.Errorf("variable %s already exists", name)
	}

	csp.Variables[name] = &Variable{
		Name:   name,
		Domain: domain,
		Value:  nil,
	}

	return nil
}

// AddConstraint adds a constraint to the CSP.
func (csp *CSP) AddConstraint(variables []string, isSatisfied func(map[string]interface{}) bool) error {
	// Verify all variables exist
	for _, varName := range variables {
		if _, exists := csp.Variables[varName]; !exists {
			return fmt.Errorf("variable %s does not exist", varName)
		}
	}

	csp.Constraints = append(csp.Constraints, &Constraint{
		Variables:   variables,
		IsSatisfied: isSatisfied,
	})

	return nil
}

// Solve attempts to find a solution using backtracking search.
func (csp *CSP) Solve() (map[string]interface{}, error) {
	assignments := make(map[string]interface{})
	solution := csp.backtrack(assignments)

	if solution == nil {
		return nil, fmt.Errorf("no solution found")
	}

	return solution, nil
}

// backtrack performs recursive backtracking search.
func (csp *CSP) backtrack(assignments map[string]interface{}) map[string]interface{} {
	// Base case: all variables assigned
	if len(assignments) == len(csp.Variables) {
		return assignments
	}

	// Select unassigned variable
	variable := csp.selectUnassignedVariable(assignments)
	if variable == nil {
		return nil
	}

	// Try each value in domain
	for _, value := range variable.Domain {
		// Assign value
		assignments[variable.Name] = value

		// Check if consistent
		if csp.isConsistent(variable.Name, assignments) {
			// Recursively solve
			result := csp.backtrack(assignments)
			if result != nil {
				return result
			}
		}

		// Backtrack
		delete(assignments, variable.Name)
	}

	return nil
}

// selectUnassignedVariable selects the next variable to assign.
// Uses Minimum Remaining Values (MRV) heuristic: choose variable with smallest domain.
func (csp *CSP) selectUnassignedVariable(assignments map[string]interface{}) *Variable {
	var best *Variable
	minRemaining := -1

	for name, variable := range csp.Variables {
		// Skip assigned variables
		if _, assigned := assignments[name]; assigned {
			continue
		}

		remaining := len(variable.Domain)

		if best == nil || remaining < minRemaining {
			best = variable
			minRemaining = remaining
		}
	}

	return best
}

// isConsistent checks if current assignment is consistent with constraints.
func (csp *CSP) isConsistent(varName string, assignments map[string]interface{}) bool {
	for _, constraint := range csp.Constraints {
		// Check if this constraint involves the variable
		involves := false
		for _, v := range constraint.Variables {
			if v == varName {
				involves = true
				break
			}
		}

		if !involves {
			continue
		}

		// Check if all variables in constraint are assigned
		allAssigned := true
		for _, v := range constraint.Variables {
			if _, assigned := assignments[v]; !assigned {
				allAssigned = false
				break
			}
		}

		// Only check constraint if all involved variables are assigned
		if allAssigned {
			if !constraint.IsSatisfied(assignments) {
				return false
			}
		}
	}

	return true
}

// Reset clears all assignments and allows re-solving.
func (csp *CSP) Reset() {
	for _, variable := range csp.Variables {
		variable.Value = nil
	}
}

// GetVariableNames returns list of all variable names.
func (csp *CSP) GetVariableNames() []string {
	names := make([]string, 0, len(csp.Variables))
	for name := range csp.Variables {
		names = append(names, name)
	}
	return names
}

// GetDomainSize returns the size of a variable's domain.
func (csp *CSP) GetDomainSize(varName string) int {
	if variable, exists := csp.Variables[varName]; exists {
		return len(variable.Domain)
	}
	return 0
}

// GetConstraintCount returns the number of constraints.
func (csp *CSP) GetConstraintCount() int {
	return len(csp.Constraints)
}

// PuzzleSolver wraps CSP for specific puzzle generation.
type PuzzleSolver struct {
	csp *CSP
}

// NewPuzzleSolver creates a new puzzle solver.
func NewPuzzleSolver(seed int64) *PuzzleSolver {
	return &PuzzleSolver{
		csp: NewCSP(seed),
	}
}

// AddElement adds a puzzle element as a variable.
func (ps *PuzzleSolver) AddElement(name string, possibleStates []int) error {
	// Convert int slice to interface slice
	domain := make([]interface{}, len(possibleStates))
	for i, state := range possibleStates {
		domain[i] = state
	}

	return ps.csp.AddVariable(name, domain)
}

// AddSequenceConstraint ensures elements are activated in specific order.
func (ps *PuzzleSolver) AddSequenceConstraint(sequence []string) error {
	if len(sequence) < 2 {
		return fmt.Errorf("sequence must have at least 2 elements")
	}

	// Each pair must be in order
	for i := 0; i < len(sequence)-1; i++ {
		first := sequence[i]
		second := sequence[i+1]

		err := ps.csp.AddConstraint([]string{first, second}, func(assignments map[string]interface{}) bool {
			v1, ok1 := assignments[first]
			v2, ok2 := assignments[second]

			if !ok1 || !ok2 {
				return true // Not all assigned yet
			}

			// First element's state must be less than second element's state
			// This ensures order
			state1, ok1 := v1.(int)
			state2, ok2 := v2.(int)

			if !ok1 || !ok2 {
				return false
			}

			return state1 < state2
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// AddUniquenessConstraint ensures all elements have different states.
func (ps *PuzzleSolver) AddUniquenessConstraint(elements []string) error {
	if len(elements) < 2 {
		return fmt.Errorf("uniqueness requires at least 2 elements")
	}

	// Each pair must have different values
	for i := 0; i < len(elements); i++ {
		for j := i + 1; j < len(elements); j++ {
			elem1 := elements[i]
			elem2 := elements[j]

			err := ps.csp.AddConstraint([]string{elem1, elem2}, func(assignments map[string]interface{}) bool {
				v1, ok1 := assignments[elem1]
				v2, ok2 := assignments[elem2]

				if !ok1 || !ok2 {
					return true // Not all assigned yet
				}

				return v1 != v2
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddSumConstraint ensures sum of states equals target.
func (ps *PuzzleSolver) AddSumConstraint(elements []string, targetSum int) error {
	if len(elements) == 0 {
		return fmt.Errorf("sum constraint requires at least 1 element")
	}

	return ps.csp.AddConstraint(elements, func(assignments map[string]interface{}) bool {
		sum := 0
		for _, elem := range elements {
			v, ok := assignments[elem]
			if !ok {
				return true // Not all assigned yet
			}

			state, ok := v.(int)
			if !ok {
				return false
			}

			sum += state
		}

		return sum == targetSum
	})
}

// Solve attempts to find a valid puzzle solution.
func (ps *PuzzleSolver) Solve() (map[string]int, error) {
	solution, err := ps.csp.Solve()
	if err != nil {
		return nil, err
	}

	// Convert interface{} map to int map
	intSolution := make(map[string]int)
	for key, value := range solution {
		intValue, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("invalid value type for %s", key)
		}
		intSolution[key] = intValue
	}

	return intSolution, nil
}

// GetCSP returns the underlying CSP (for advanced usage).
func (ps *PuzzleSolver) GetCSP() *CSP {
	return ps.csp
}
