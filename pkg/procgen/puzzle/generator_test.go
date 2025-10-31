package puzzle

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	// Verify templates were initialized
	expectedTypes := []PuzzleType{
		PuzzleTypePressurePlate,
		PuzzleTypeLeverSequence,
		PuzzleTypeBlockPushing,
		PuzzleTypeTimedChallenge,
		PuzzleTypeMemoryPattern,
		PuzzleTypeColorMatching,
	}

	for _, puzzleType := range expectedTypes {
		if _, exists := gen.templates[puzzleType]; !exists {
			t.Errorf("Template for %s not initialized", puzzleType)
		}
	}
}

func TestGenerateDeterminism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Generate puzzle twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	puzzle1, ok1 := result1.(*Puzzle)
	puzzle2, ok2 := result2.(*Puzzle)

	if !ok1 || !ok2 {
		t.Fatal("Generated results are not Puzzle types")
	}

	// Verify determinism
	if puzzle1.Type != puzzle2.Type {
		t.Errorf("Puzzle types differ: %s != %s", puzzle1.Type, puzzle2.Type)
	}

	if puzzle1.Difficulty != puzzle2.Difficulty {
		t.Errorf("Difficulties differ: %d != %d", puzzle1.Difficulty, puzzle2.Difficulty)
	}

	if puzzle1.ElementCount != puzzle2.ElementCount {
		t.Errorf("Element counts differ: %d != %d", puzzle1.ElementCount, puzzle2.ElementCount)
	}

	if len(puzzle1.Solution) != len(puzzle2.Solution) {
		t.Errorf("Solution lengths differ: %d != %d", len(puzzle1.Solution), len(puzzle2.Solution))
	}
}

func TestGenerateAllPuzzleTypes(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name       string
		difficulty float64
		depth      int
	}{
		{"easy", 0.2, 1},
		{"medium", 0.5, 5},
		{"hard", 0.8, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    "fantasy",
			}

			// Generate multiple puzzles to cover different types
			for i := 0; i < 20; i++ {
				seed := int64(1000 + i)
				result, err := gen.Generate(seed, params)
				if err != nil {
					t.Errorf("Generation failed for seed %d: %v", seed, err)
					continue
				}

				puzzle, ok := result.(*Puzzle)
				if !ok {
					t.Errorf("Result is not a Puzzle type")
					continue
				}

				// Validate puzzle
				if err := gen.Validate(puzzle); err != nil {
					t.Errorf("Validation failed for puzzle %s: %v", puzzle.ID, err)
				}

				// Verify basic properties
				if puzzle.Difficulty < 1 || puzzle.Difficulty > 10 {
					t.Errorf("Difficulty %d out of range [1-10]", puzzle.Difficulty)
				}

				if puzzle.ElementCount < 1 {
					t.Errorf("Puzzle has no elements")
				}

				if len(puzzle.Solution) == 0 {
					t.Errorf("Puzzle has no solution")
				}
			}
		})
	}
}

func TestCalculateDifficulty(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name        string
		difficulty  float64
		depth       int
		expectedMin int
		expectedMax int
	}{
		{"very_easy", 0.0, 1, 1, 2},
		{"easy", 0.3, 2, 2, 4},
		{"medium", 0.5, 5, 4, 7},
		{"hard", 0.8, 8, 7, 10},
		{"very_hard", 1.0, 10, 9, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
			}

			diff := gen.calculateDifficulty(params)

			if diff < tt.expectedMin || diff > tt.expectedMax {
				t.Errorf("Difficulty %d not in expected range [%d-%d]",
					diff, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestPressurePlatePuzzle(t *testing.T) {
	gen := NewGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Generate multiple puzzles and force pressure plate type
	for i := 0; i < 5; i++ {
		template := gen.templates[PuzzleTypePressurePlate]
		difficulty := gen.calculateDifficulty(params)

		puzzle, err := gen.generatePressurePlatePuzzle(
			selectRNG(seed+int64(i)),
			template,
			difficulty,
			params,
		)

		if err != nil {
			t.Fatalf("Failed to generate pressure plate puzzle: %v", err)
		}

		if puzzle.Type != PuzzleTypePressurePlate {
			t.Errorf("Expected type %s, got %s", PuzzleTypePressurePlate, puzzle.Type)
		}

		// Verify all elements are pressure plates
		for _, elem := range puzzle.Elements {
			if elem.ElementType != "pressure_plate" {
				t.Errorf("Expected pressure_plate, got %s", elem.ElementType)
			}
		}

		// Verify solution references valid elements
		for _, solID := range puzzle.Solution {
			found := false
			for _, elem := range puzzle.Elements {
				if elem.ID == solID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Solution references non-existent element: %s", solID)
			}
		}
	}
}

func TestLeverSequencePuzzle(t *testing.T) {
	gen := NewGenerator()
	template := gen.templates[PuzzleTypeLeverSequence]
	difficulty := 5

	puzzle, err := gen.generateLeverSequencePuzzle(
		selectRNG(54321),
		template,
		difficulty,
		procgen.GenerationParams{},
	)

	if err != nil {
		t.Fatalf("Failed to generate lever sequence puzzle: %v", err)
	}

	if puzzle.Type != PuzzleTypeLeverSequence {
		t.Errorf("Expected type %s, got %s", PuzzleTypeLeverSequence, puzzle.Type)
	}

	// Verify max attempts is set (lever puzzles have limited attempts)
	if puzzle.MaxAttempts == 0 {
		t.Error("Lever sequence should have maximum attempts limit")
	}

	// Verify all elements are levers
	for _, elem := range puzzle.Elements {
		if elem.ElementType != "lever" {
			t.Errorf("Expected lever, got %s", elem.ElementType)
		}
	}
}

func TestBlockPushingPuzzle(t *testing.T) {
	gen := NewGenerator()
	template := gen.templates[PuzzleTypeBlockPushing]
	difficulty := 6

	puzzle, err := gen.generateBlockPushingPuzzle(
		selectRNG(11111),
		template,
		difficulty,
		procgen.GenerationParams{},
	)

	if err != nil {
		t.Fatalf("Failed to generate block pushing puzzle: %v", err)
	}

	if puzzle.Type != PuzzleTypeBlockPushing {
		t.Errorf("Expected type %s, got %s", PuzzleTypeBlockPushing, puzzle.Type)
	}

	// Should have blocks and targets
	hasBlocks := false
	hasTargets := false

	for _, elem := range puzzle.Elements {
		if elem.ElementType == "pushable_block" {
			hasBlocks = true
		}
		if elem.ElementType == "block_target" {
			hasTargets = true
		}
	}

	if !hasBlocks {
		t.Error("Block puzzle should have pushable blocks")
	}

	if !hasTargets {
		t.Error("Block puzzle should have target positions")
	}
}

func TestTimedChallengePuzzle(t *testing.T) {
	gen := NewGenerator()
	template := gen.templates[PuzzleTypeTimedChallenge]
	difficulty := 7

	puzzle, err := gen.generateTimedChallengePuzzle(
		selectRNG(22222),
		template,
		difficulty,
		procgen.GenerationParams{},
	)

	if err != nil {
		t.Fatalf("Failed to generate timed challenge puzzle: %v", err)
	}

	if puzzle.Type != PuzzleTypeTimedChallenge {
		t.Errorf("Expected type %s, got %s", PuzzleTypeTimedChallenge, puzzle.Type)
	}

	// Must have time limit
	if puzzle.TimeLimit <= 0 {
		t.Error("Timed challenge must have positive time limit")
	}
}

func TestMemoryPatternPuzzle(t *testing.T) {
	gen := NewGenerator()
	template := gen.templates[PuzzleTypeMemoryPattern]
	difficulty := 4

	puzzle, err := gen.generateMemoryPatternPuzzle(
		selectRNG(33333),
		template,
		difficulty,
		procgen.GenerationParams{},
	)

	if err != nil {
		t.Fatalf("Failed to generate memory pattern puzzle: %v", err)
	}

	if puzzle.Type != PuzzleTypeMemoryPattern {
		t.Errorf("Expected type %s, got %s", PuzzleTypeMemoryPattern, puzzle.Type)
	}

	// Verify all elements are memory symbols
	for _, elem := range puzzle.Elements {
		if elem.ElementType != "memory_symbol" {
			t.Errorf("Expected memory_symbol, got %s", elem.ElementType)
		}
	}

	// Should have max attempts
	if puzzle.MaxAttempts == 0 {
		t.Error("Memory pattern should have maximum attempts limit")
	}
}

func TestColorMatchingPuzzle(t *testing.T) {
	gen := NewGenerator()
	template := gen.templates[PuzzleTypeColorMatching]
	difficulty := 5

	puzzle, err := gen.generateColorMatchingPuzzle(
		selectRNG(44444),
		template,
		difficulty,
		procgen.GenerationParams{},
	)

	if err != nil {
		t.Fatalf("Failed to generate color matching puzzle: %v", err)
	}

	if puzzle.Type != PuzzleTypeColorMatching {
		t.Errorf("Expected type %s, got %s", PuzzleTypeColorMatching, puzzle.Type)
	}

	// Verify all elements are colored tiles
	for _, elem := range puzzle.Elements {
		if elem.ElementType != "colored_tile" {
			t.Errorf("Expected colored_tile, got %s", elem.ElementType)
		}

		// Verify state is a string (color name)
		if _, ok := elem.State.(string); !ok {
			t.Errorf("Colored tile state should be string, got %T", elem.State)
		}
	}
}

func TestValidation(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		puzzle  *Puzzle
		wantErr bool
	}{
		{
			name: "valid_puzzle",
			puzzle: &Puzzle{
				ID:           "test_1",
				Type:         PuzzleTypePressurePlate,
				Difficulty:   5,
				Solution:     []string{"elem_0", "elem_1"},
				ElementCount: 3,
				Elements: []PuzzleElement{
					{ID: "elem_0", ElementType: "pressure_plate"},
					{ID: "elem_1", ElementType: "pressure_plate"},
					{ID: "elem_2", ElementType: "pressure_plate"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing_id",
			puzzle: &Puzzle{
				ID:           "",
				ElementCount: 1,
				Elements:     []PuzzleElement{{ID: "elem_0"}},
				Solution:     []string{"elem_0"},
				Difficulty:   5,
			},
			wantErr: true,
		},
		{
			name: "no_elements",
			puzzle: &Puzzle{
				ID:           "test_2",
				ElementCount: 0,
				Elements:     []PuzzleElement{},
				Solution:     []string{},
				Difficulty:   5,
			},
			wantErr: true,
		},
		{
			name: "no_solution",
			puzzle: &Puzzle{
				ID:           "test_3",
				ElementCount: 2,
				Elements:     []PuzzleElement{{ID: "elem_0"}, {ID: "elem_1"}},
				Solution:     []string{},
				Difficulty:   5,
			},
			wantErr: true,
		},
		{
			name: "invalid_difficulty",
			puzzle: &Puzzle{
				ID:           "test_4",
				ElementCount: 1,
				Elements:     []PuzzleElement{{ID: "elem_0"}},
				Solution:     []string{"elem_0"},
				Difficulty:   15, // Out of range [1-10]
			},
			wantErr: true,
		},
		{
			name: "element_count_mismatch",
			puzzle: &Puzzle{
				ID:           "test_5",
				ElementCount: 5,                               // Declared 5
				Elements:     []PuzzleElement{{ID: "elem_0"}}, // But only 1
				Solution:     []string{"elem_0"},
				Difficulty:   5,
			},
			wantErr: true,
		},
		{
			name: "solution_invalid_reference",
			puzzle: &Puzzle{
				ID:           "test_6",
				ElementCount: 2,
				Elements:     []PuzzleElement{{ID: "elem_0"}, {ID: "elem_1"}},
				Solution:     []string{"elem_99"}, // Non-existent element
				Difficulty:   5,
			},
			wantErr: true,
		},
		{
			name: "timed_without_limit",
			puzzle: &Puzzle{
				ID:           "test_7",
				Type:         PuzzleTypeTimedChallenge,
				ElementCount: 1,
				Elements:     []PuzzleElement{{ID: "elem_0"}},
				Solution:     []string{"elem_0"},
				Difficulty:   5,
				TimeLimit:    0, // Invalid for timed challenge
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.puzzle)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function for tests
func selectRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Generate a puzzle once
	result, _ := gen.Generate(12345, params)
	puzzle := result.(*Puzzle)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Validate(puzzle)
	}
}
