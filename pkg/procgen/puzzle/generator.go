// Package puzzle provides procedural puzzle generation for Phase 11.2.
// This file implements the Generator that creates solvable constraint-based puzzles
// using deterministic seed-based generation.
package puzzle

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
)

// PuzzleType defines the category of puzzle.
type PuzzleType string

const (
	// PuzzleTypePressurePlate requires stepping on specific plates
	PuzzleTypePressurePlate PuzzleType = "pressure_plate"
	// PuzzleTypeLeverSequence requires activating levers in correct order
	PuzzleTypeLeverSequence PuzzleType = "lever_sequence"
	// PuzzleTypeBlockPushing requires pushing blocks to specific locations
	PuzzleTypeBlockPushing PuzzleType = "block_pushing"
	// PuzzleTypeTimedChallenge requires completing within time limit
	PuzzleTypeTimedChallenge PuzzleType = "timed_challenge"
	// PuzzleTypeMemoryPattern requires repeating a shown pattern
	PuzzleTypeMemoryPattern PuzzleType = "memory_pattern"
	// PuzzleTypeColorMatching requires matching colors/symbols
	PuzzleTypeColorMatching PuzzleType = "color_matching"
)

// PuzzleTemplate defines puzzle generation parameters.
type PuzzleTemplate struct {
	Type             PuzzleType
	MinElements      int        // Minimum number of puzzle elements
	MaxElements      int        // Maximum number of puzzle elements
	MinComplexity    int        // Minimum solution complexity (1-10)
	MaxComplexity    int        // Maximum solution complexity (1-10)
	TimeLimitRange   [2]float64 // Time limit range in seconds (0 = no limit)
	MaxAttemptsRange [2]int     // Max attempts range (0 = unlimited)
}

// Puzzle represents a generated puzzle with solution.
type Puzzle struct {
	ID           string          // Unique puzzle identifier
	Type         PuzzleType      // Puzzle category
	Difficulty   int             // Overall difficulty (1-10)
	Solution     []string        // Correct solution sequence
	ElementCount int             // Number of interactive elements
	Elements     []PuzzleElement // Puzzle element details
	TimeLimit    float64         // Time limit in seconds (0 = no limit)
	MaxAttempts  int             // Maximum attempts (0 = unlimited)
	HintText     string          // Player-facing hint
	Description  string          // Puzzle description
	RewardType   string          // Type of reward (door, chest, etc.)
}

// PuzzleElement represents an interactive puzzle element.
type PuzzleElement struct {
	ID           string      // Element identifier
	ElementType  string      // Type (plate, lever, block, etc.)
	Position     [2]int      // Grid position (x, y)
	State        interface{} // Element-specific state
	Interactable bool        // Whether player can interact
}

// Generator creates procedural puzzles using constraint solving.
type Generator struct {
	templates map[PuzzleType]PuzzleTemplate
}

// NewGenerator creates a new puzzle generator with standard templates.
func NewGenerator() *Generator {
	g := &Generator{
		templates: make(map[PuzzleType]PuzzleTemplate),
	}
	g.initializeTemplates()
	return g
}

// initializeTemplates sets up standard puzzle templates.
func (g *Generator) initializeTemplates() {
	g.templates[PuzzleTypePressurePlate] = PuzzleTemplate{
		Type:             PuzzleTypePressurePlate,
		MinElements:      2,
		MaxElements:      8,
		MinComplexity:    1,
		MaxComplexity:    7,
		TimeLimitRange:   [2]float64{0, 60},
		MaxAttemptsRange: [2]int{0, 5},
	}

	g.templates[PuzzleTypeLeverSequence] = PuzzleTemplate{
		Type:             PuzzleTypeLeverSequence,
		MinElements:      3,
		MaxElements:      6,
		MinComplexity:    2,
		MaxComplexity:    10,
		TimeLimitRange:   [2]float64{0, 45},
		MaxAttemptsRange: [2]int{0, 10},
	}

	g.templates[PuzzleTypeBlockPushing] = PuzzleTemplate{
		Type:             PuzzleTypeBlockPushing,
		MinElements:      1,
		MaxElements:      4,
		MinComplexity:    3,
		MaxComplexity:    8,
		TimeLimitRange:   [2]float64{0, 120},
		MaxAttemptsRange: [2]int{0, 0},
	}

	g.templates[PuzzleTypeTimedChallenge] = PuzzleTemplate{
		Type:             PuzzleTypeTimedChallenge,
		MinElements:      3,
		MaxElements:      10,
		MinComplexity:    4,
		MaxComplexity:    9,
		TimeLimitRange:   [2]float64{10, 30},
		MaxAttemptsRange: [2]int{0, 3},
	}

	g.templates[PuzzleTypeMemoryPattern] = PuzzleTemplate{
		Type:             PuzzleTypeMemoryPattern,
		MinElements:      4,
		MaxElements:      9,
		MinComplexity:    2,
		MaxComplexity:    8,
		TimeLimitRange:   [2]float64{0, 30},
		MaxAttemptsRange: [2]int{0, 5},
	}

	g.templates[PuzzleTypeColorMatching] = PuzzleTemplate{
		Type:             PuzzleTypeColorMatching,
		MinElements:      3,
		MaxElements:      6,
		MinComplexity:    2,
		MaxComplexity:    7,
		TimeLimitRange:   [2]float64{0, 45},
		MaxAttemptsRange: [2]int{0, 8},
	}
}

// Generate creates a new puzzle using the provided seed and parameters.
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Create RNG from seed
	rng := rand.New(rand.NewSource(seed))

	// Select puzzle type based on difficulty and depth
	puzzleType := g.selectPuzzleType(rng, params)

	// Get template for selected type
	template, exists := g.templates[puzzleType]
	if !exists {
		return nil, fmt.Errorf("no template for puzzle type: %s", puzzleType)
	}

	// Calculate difficulty (1-10 scale)
	difficulty := g.calculateDifficulty(params)

	// Generate puzzle based on type
	var puzzle *Puzzle
	var err error

	switch puzzleType {
	case PuzzleTypePressurePlate:
		puzzle, err = g.generatePressurePlatePuzzle(rng, template, difficulty, params)
	case PuzzleTypeLeverSequence:
		puzzle, err = g.generateLeverSequencePuzzle(rng, template, difficulty, params)
	case PuzzleTypeBlockPushing:
		puzzle, err = g.generateBlockPushingPuzzle(rng, template, difficulty, params)
	case PuzzleTypeTimedChallenge:
		puzzle, err = g.generateTimedChallengePuzzle(rng, template, difficulty, params)
	case PuzzleTypeMemoryPattern:
		puzzle, err = g.generateMemoryPatternPuzzle(rng, template, difficulty, params)
	case PuzzleTypeColorMatching:
		puzzle, err = g.generateColorMatchingPuzzle(rng, template, difficulty, params)
	default:
		return nil, fmt.Errorf("unsupported puzzle type: %s", puzzleType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate %s puzzle: %w", puzzleType, err)
	}

	return puzzle, nil
}

// Validate verifies the generated puzzle meets quality standards.
func (g *Generator) Validate(result interface{}) error {
	puzzle, ok := result.(*Puzzle)
	if !ok {
		return fmt.Errorf("invalid puzzle type: expected *Puzzle")
	}

	// Verify basic properties
	if puzzle.ID == "" {
		return fmt.Errorf("puzzle missing ID")
	}

	if puzzle.ElementCount < 1 {
		return fmt.Errorf("puzzle has no elements")
	}

	if len(puzzle.Solution) == 0 {
		return fmt.Errorf("puzzle has no solution")
	}

	// Verify difficulty in valid range
	if puzzle.Difficulty < 1 || puzzle.Difficulty > 10 {
		return fmt.Errorf("puzzle difficulty %d out of range [1-10]", puzzle.Difficulty)
	}

	// Verify element count matches
	if len(puzzle.Elements) != puzzle.ElementCount {
		return fmt.Errorf("element count mismatch: declared=%d, actual=%d",
			puzzle.ElementCount, len(puzzle.Elements))
	}

	// Verify solution references valid elements
	elementIDs := make(map[string]bool)
	for _, elem := range puzzle.Elements {
		elementIDs[elem.ID] = true
	}

	for _, solutionID := range puzzle.Solution {
		if !elementIDs[solutionID] {
			return fmt.Errorf("solution references non-existent element: %s", solutionID)
		}
	}

	// Type-specific validation
	switch puzzle.Type {
	case PuzzleTypeTimedChallenge:
		if puzzle.TimeLimit <= 0 {
			return fmt.Errorf("timed challenge must have positive time limit")
		}
	}

	return nil
}

// selectPuzzleType chooses a puzzle type based on parameters.
func (g *Generator) selectPuzzleType(rng *rand.Rand, params procgen.GenerationParams) PuzzleType {
	// Difficulty influences puzzle type distribution
	difficulty := params.Difficulty

	// Early game: simpler puzzles (pressure plates, levers)
	if difficulty < 0.3 || params.Depth < 3 {
		types := []PuzzleType{PuzzleTypePressurePlate, PuzzleTypeLeverSequence}
		return types[rng.Intn(len(types))]
	}

	// Mid game: varied puzzles
	if difficulty < 0.7 || params.Depth < 7 {
		types := []PuzzleType{
			PuzzleTypePressurePlate,
			PuzzleTypeLeverSequence,
			PuzzleTypeMemoryPattern,
			PuzzleTypeColorMatching,
		}
		return types[rng.Intn(len(types))]
	}

	// Late game: all puzzle types including complex ones
	types := []PuzzleType{
		PuzzleTypePressurePlate,
		PuzzleTypeLeverSequence,
		PuzzleTypeBlockPushing,
		PuzzleTypeTimedChallenge,
		PuzzleTypeMemoryPattern,
		PuzzleTypeColorMatching,
	}
	return types[rng.Intn(len(types))]
}

// calculateDifficulty determines puzzle difficulty from generation parameters.
func (g *Generator) calculateDifficulty(params procgen.GenerationParams) int {
	// Base difficulty from params.Difficulty (0.0-1.0)
	baseDiff := params.Difficulty * 6.0 // 0-6 range

	// Add depth scaling (1-4 additional difficulty)
	depthBonus := float64(params.Depth) * 0.3
	if depthBonus > 4.0 {
		depthBonus = 4.0
	}

	difficulty := int(baseDiff + depthBonus + 1.0) // +1 ensures minimum 1
	if difficulty > 10 {
		difficulty = 10
	}

	return difficulty
}

// generatePressurePlatePuzzle creates a pressure plate stepping puzzle.
func (g *Generator) generatePressurePlatePuzzle(rng *rand.Rand, template PuzzleTemplate, difficulty int, params procgen.GenerationParams) (*Puzzle, error) {
	// Element count based on difficulty
	elementCount := template.MinElements + int(float64(template.MaxElements-template.MinElements)*float64(difficulty)/10.0)

	// Generate puzzle ID
	puzzleID := fmt.Sprintf("pressure_%d", rng.Int63())

	// Create elements (pressure plates)
	elements := make([]PuzzleElement, elementCount)
	for i := 0; i < elementCount; i++ {
		elements[i] = PuzzleElement{
			ID:           fmt.Sprintf("plate_%d", i),
			ElementType:  "pressure_plate",
			Position:     [2]int{rng.Intn(10), rng.Intn(10)},
			State:        false, // unpressed
			Interactable: true,
		}
	}

	// Generate solution: which plates must be pressed
	solutionLength := 1 + difficulty/2
	if solutionLength > elementCount {
		solutionLength = elementCount
	}

	// Use CSP to ensure solvable solution
	solution := make([]string, 0, solutionLength)
	usedIndices := make(map[int]bool)

	for len(solution) < solutionLength {
		idx := rng.Intn(elementCount)
		if !usedIndices[idx] {
			solution = append(solution, elements[idx].ID)
			usedIndices[idx] = true
		}
	}

	// Generate hint
	hint := fmt.Sprintf("Step on %d pressure plates to unlock the door", solutionLength)

	puzzle := &Puzzle{
		ID:           puzzleID,
		Type:         PuzzleTypePressurePlate,
		Difficulty:   difficulty,
		Solution:     solution,
		ElementCount: elementCount,
		Elements:     elements,
		TimeLimit:    0, // No time limit by default
		MaxAttempts:  0, // Unlimited attempts
		HintText:     hint,
		Description:  "Ancient pressure plates guard this passage",
		RewardType:   "door",
	}

	return puzzle, nil
}

// generateLeverSequencePuzzle creates a lever activation sequence puzzle.
func (g *Generator) generateLeverSequencePuzzle(rng *rand.Rand, template PuzzleTemplate, difficulty int, params procgen.GenerationParams) (*Puzzle, error) {
	// Element count based on difficulty
	elementCount := template.MinElements + int(float64(template.MaxElements-template.MinElements)*float64(difficulty)/10.0)

	// Generate puzzle ID
	puzzleID := fmt.Sprintf("lever_%d", rng.Int63())

	// Create elements (levers)
	elements := make([]PuzzleElement, elementCount)
	for i := 0; i < elementCount; i++ {
		elements[i] = PuzzleElement{
			ID:           fmt.Sprintf("lever_%d", i),
			ElementType:  "lever",
			Position:     [2]int{rng.Intn(10), rng.Intn(10)},
			State:        "off", // off/on
			Interactable: true,
		}
	}

	// Generate solution sequence (order matters!)
	solutionLength := 2 + difficulty/2
	if solutionLength > elementCount {
		solutionLength = elementCount
	}

	solution := make([]string, 0, solutionLength)
	usedIndices := make(map[int]bool)

	for len(solution) < solutionLength {
		idx := rng.Intn(elementCount)
		if !usedIndices[idx] {
			solution = append(solution, elements[idx].ID)
			usedIndices[idx] = true
		}
	}

	// Generate hint
	hint := fmt.Sprintf("Activate %d levers in the correct sequence", solutionLength)

	puzzle := &Puzzle{
		ID:           puzzleID,
		Type:         PuzzleTypeLeverSequence,
		Difficulty:   difficulty,
		Solution:     solution,
		ElementCount: elementCount,
		Elements:     elements,
		TimeLimit:    0,
		MaxAttempts:  5 + difficulty, // Limited attempts increase with difficulty
		HintText:     hint,
		Description:  "The levers must be pulled in a specific order",
		RewardType:   "door",
	}

	return puzzle, nil
}

// generateBlockPushingPuzzle creates a block pushing puzzle.
func (g *Generator) generateBlockPushingPuzzle(rng *rand.Rand, template PuzzleTemplate, difficulty int, params procgen.GenerationParams) (*Puzzle, error) {
	// Element count based on difficulty
	elementCount := template.MinElements + int(float64(template.MaxElements-template.MinElements)*float64(difficulty)/10.0)

	puzzleID := fmt.Sprintf("block_%d", rng.Int63())

	// Create blocks and target positions
	elements := make([]PuzzleElement, elementCount*2) // blocks + targets

	for i := 0; i < elementCount; i++ {
		// Block
		elements[i] = PuzzleElement{
			ID:           fmt.Sprintf("block_%d", i),
			ElementType:  "pushable_block",
			Position:     [2]int{rng.Intn(8), rng.Intn(8)},
			State:        map[string]interface{}{"on_target": false},
			Interactable: true,
		}

		// Target position
		elements[elementCount+i] = PuzzleElement{
			ID:           fmt.Sprintf("target_%d", i),
			ElementType:  "block_target",
			Position:     [2]int{rng.Intn(8) + 2, rng.Intn(8) + 2},
			State:        map[string]interface{}{"occupied": false},
			Interactable: false,
		}
	}

	// Solution: each block must be on its corresponding target
	// Format: just block IDs (PuzzleSystem will check target positions)
	solution := make([]string, elementCount)
	for i := 0; i < elementCount; i++ {
		solution[i] = fmt.Sprintf("block_%d", i)
	}

	hint := fmt.Sprintf("Push %d blocks onto their targets", elementCount)

	puzzle := &Puzzle{
		ID:           puzzleID,
		Type:         PuzzleTypeBlockPushing,
		Difficulty:   difficulty,
		Solution:     solution,
		ElementCount: elementCount * 2,
		Elements:     elements,
		TimeLimit:    0,
		MaxAttempts:  0,
		HintText:     hint,
		Description:  "Heavy blocks must be positioned precisely",
		RewardType:   "chest",
	}

	return puzzle, nil
}

// generateTimedChallengePuzzle creates a time-limited puzzle.
func (g *Generator) generateTimedChallengePuzzle(rng *rand.Rand, template PuzzleTemplate, difficulty int, params procgen.GenerationParams) (*Puzzle, error) {
	// Timed challenge is a variant of lever or pressure plate
	baseType := PuzzleTypePressurePlate
	if rng.Float64() < 0.5 {
		baseType = PuzzleTypeLeverSequence
	}

	var puzzle *Puzzle
	var err error

	if baseType == PuzzleTypePressurePlate {
		puzzle, err = g.generatePressurePlatePuzzle(rng, template, difficulty, params)
	} else {
		puzzle, err = g.generateLeverSequencePuzzle(rng, template, difficulty, params)
	}

	if err != nil {
		return nil, err
	}

	// Override type and add time limit
	puzzle.Type = PuzzleTypeTimedChallenge
	puzzle.TimeLimit = template.TimeLimitRange[0] +
		rng.Float64()*(template.TimeLimitRange[1]-template.TimeLimitRange[0])

	// Reduce time limit as difficulty increases
	puzzle.TimeLimit = puzzle.TimeLimit * (1.0 - float64(difficulty)/20.0)

	puzzle.HintText = fmt.Sprintf("%s (Time Limit: %.0fs)", puzzle.HintText, puzzle.TimeLimit)
	puzzle.Description = "Time is of the essence - " + strings.ToLower(puzzle.Description)

	return puzzle, nil
}

// generateMemoryPatternPuzzle creates a pattern memory puzzle.
func (g *Generator) generateMemoryPatternPuzzle(rng *rand.Rand, template PuzzleTemplate, difficulty int, params procgen.GenerationParams) (*Puzzle, error) {
	// Element count based on difficulty
	elementCount := template.MinElements + int(float64(template.MaxElements-template.MinElements)*float64(difficulty)/10.0)

	puzzleID := fmt.Sprintf("memory_%d", rng.Int63())

	// Create elements (symbols/patterns)
	elements := make([]PuzzleElement, elementCount)
	symbols := []string{"circle", "square", "triangle", "diamond", "star", "hexagon", "cross", "moon", "sun"}

	for i := 0; i < elementCount; i++ {
		elements[i] = PuzzleElement{
			ID:           fmt.Sprintf("symbol_%d", i),
			ElementType:  "memory_symbol",
			Position:     [2]int{i % 3, i / 3},
			State:        symbols[i%len(symbols)],
			Interactable: true,
		}
	}

	// Solution: random sequence of elements to activate
	solutionLength := 2 + difficulty/3
	if solutionLength > elementCount {
		solutionLength = elementCount
	}

	solution := make([]string, solutionLength)
	for i := 0; i < solutionLength; i++ {
		solution[i] = elements[rng.Intn(elementCount)].ID
	}

	hint := fmt.Sprintf("Remember and repeat the pattern of %d symbols", solutionLength)

	puzzle := &Puzzle{
		ID:           puzzleID,
		Type:         PuzzleTypeMemoryPattern,
		Difficulty:   difficulty,
		Solution:     solution,
		ElementCount: elementCount,
		Elements:     elements,
		TimeLimit:    0,
		MaxAttempts:  3 + difficulty/2,
		HintText:     hint,
		Description:  "Watch carefully and repeat the pattern",
		RewardType:   "door",
	}

	return puzzle, nil
}

// generateColorMatchingPuzzle creates a color/symbol matching puzzle.
func (g *Generator) generateColorMatchingPuzzle(rng *rand.Rand, template PuzzleTemplate, difficulty int, params procgen.GenerationParams) (*Puzzle, error) {
	// Element count based on difficulty
	elementCount := template.MinElements + int(float64(template.MaxElements-template.MinElements)*float64(difficulty)/10.0)

	puzzleID := fmt.Sprintf("color_%d", rng.Int63())

	// Create elements (colored tiles/switches)
	colors := []string{"red", "blue", "green", "yellow", "purple", "orange", "white", "black"}
	elements := make([]PuzzleElement, elementCount)

	for i := 0; i < elementCount; i++ {
		elements[i] = PuzzleElement{
			ID:           fmt.Sprintf("tile_%d", i),
			ElementType:  "colored_tile",
			Position:     [2]int{rng.Intn(10), rng.Intn(10)},
			State:        colors[rng.Intn(len(colors))],
			Interactable: true,
		}
	}

	// Solution: activate tiles of specific colors
	numColors := 1 + difficulty/3
	if numColors > len(colors) {
		numColors = len(colors)
	}

	targetColors := make([]string, numColors)
	usedColors := make(map[string]bool)

	for i := 0; i < numColors; i++ {
		for {
			color := colors[rng.Intn(len(colors))]
			if !usedColors[color] {
				targetColors[i] = color
				usedColors[color] = true
				break
			}
		}
	}

	// Build solution from elements matching target colors
	solution := make([]string, 0)
	for _, elem := range elements {
		if colorStr, ok := elem.State.(string); ok {
			if usedColors[colorStr] {
				solution = append(solution, elem.ID)
			}
		}
	}

	// Ensure solution has at least one element
	if len(solution) == 0 {
		// If no matching colors, just use first element
		if len(elements) > 0 {
			solution = append(solution, elements[0].ID)
		}
	}

	hint := fmt.Sprintf("Activate all %s tiles", strings.Join(targetColors, ", "))

	puzzle := &Puzzle{
		ID:           puzzleID,
		Type:         PuzzleTypeColorMatching,
		Difficulty:   difficulty,
		Solution:     solution,
		ElementCount: elementCount,
		Elements:     elements,
		TimeLimit:    0,
		MaxAttempts:  0,
		HintText:     hint,
		Description:  "Match the correct colors to proceed",
		RewardType:   "door",
	}

	return puzzle, nil
}
