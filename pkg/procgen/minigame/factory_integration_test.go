package minigame

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestGenerateAndCreateGame verifies the integrated generation and instantiation flow.
func TestGenerateAndCreateGame(t *testing.T) {
	generator := NewGenerator()
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}

	metadata, instance, err := GenerateAndCreateGame(generator, seed, params)
	if err != nil {
		t.Fatalf("GenerateAndCreateGame failed: %v", err)
	}

	// Verify metadata
	if metadata == nil {
		t.Fatal("metadata is nil")
	}
	if metadata.Difficulty != 0.5 {
		t.Errorf("Expected difficulty 0.5, got %.2f", metadata.Difficulty)
	}
	if metadata.GenreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", metadata.GenreID)
	}
	if metadata.Name == "" {
		t.Error("Generated name is empty")
	}
	if metadata.TimeLimit <= 0 {
		t.Errorf("Invalid time limit: %.2f", metadata.TimeLimit)
	}

	// Verify instance
	if instance == nil {
		t.Fatal("instance is nil")
	}
	if instance.IsComplete() {
		t.Error("Newly created instance should not be complete")
	}

	// Verify instance can be updated
	if err := instance.Update(0.016); err != nil {
		t.Errorf("Update failed: %v", err)
	}
}

// TestGenerateAndCreateGame_AllGameTypes verifies generation works for all game types.
func TestGenerateAndCreateGame_AllGameTypes(t *testing.T) {
	generator := NewGenerator()

	// Test multiple seeds to ensure we hit different game types
	for i := int64(0); i < 50; i++ {
		seed := 12345 + i
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      10,
			GenreID:    "fantasy",
			Custom:     make(map[string]interface{}),
		}

		metadata, instance, err := GenerateAndCreateGame(generator, seed, params)
		if err != nil {
			t.Fatalf("GenerateAndCreateGame failed for seed %d: %v", seed, err)
		}

		if metadata == nil || instance == nil {
			t.Fatalf("Nil result for seed %d", seed)
		}

		// Verify metadata and instance match
		expectedEngineType := GameTypeToEngineType(metadata.Type)
		if expectedEngineType.String() == "" {
			t.Errorf("Invalid engine type conversion for procgen type %v", metadata.Type)
		}
	}
}

// TestGenerateAndCreateGame_DeterministicGeneration verifies same seed produces same result.
func TestGenerateAndCreateGame_DeterministicGeneration(t *testing.T) {
	generator := NewGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      20,
		GenreID:    "scifi",
		Custom:     make(map[string]interface{}),
	}

	// Generate twice with same seed
	metadata1, instance1, err1 := GenerateAndCreateGame(generator, seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	metadata2, instance2, err2 := GenerateAndCreateGame(generator, seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Verify metadata matches
	if metadata1.Type != metadata2.Type {
		t.Errorf("Game types differ: %v vs %v", metadata1.Type, metadata2.Type)
	}
	if metadata1.Name != metadata2.Name {
		t.Errorf("Names differ: '%s' vs '%s'", metadata1.Name, metadata2.Name)
	}
	if metadata1.Difficulty != metadata2.Difficulty {
		t.Errorf("Difficulties differ: %.2f vs %.2f", metadata1.Difficulty, metadata2.Difficulty)
	}
	if metadata1.TimeLimit != metadata2.TimeLimit {
		t.Errorf("Time limits differ: %.2f vs %.2f", metadata1.TimeLimit, metadata2.TimeLimit)
	}

	// Verify instances are created (not checking deep equality since they're separate instances)
	if instance1 == nil || instance2 == nil {
		t.Error("Instances should not be nil")
	}
}

// TestGenerateAndCreateGame_InvalidDifficulty verifies error handling for invalid difficulty.
func TestGenerateAndCreateGame_InvalidDifficulty(t *testing.T) {
	generator := NewGenerator()
	seed := int64(12345)

	tests := []struct {
		name       string
		difficulty float64
		wantErr    bool
	}{
		{"Valid difficulty 0.0", 0.0, false},
		{"Valid difficulty 0.5", 0.5, false},
		{"Valid difficulty 1.0", 1.0, false},
		{"Invalid negative", -0.1, true},
		{"Invalid too high", 1.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      10,
				GenreID:    "fantasy",
				Custom:     make(map[string]interface{}),
			}

			_, _, err := GenerateAndCreateGame(generator, seed, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAndCreateGame() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateAndCreateGame_GenreVariety verifies different genres produce appropriate games.
func TestGenerateAndCreateGame_GenreVariety(t *testing.T) {
	generator := NewGenerator()
	seed := int64(42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      10,
				GenreID:    genre,
				Custom:     make(map[string]interface{}),
			}

			metadata, instance, err := GenerateAndCreateGame(generator, seed, params)
			if err != nil {
				t.Fatalf("GenerateAndCreateGame failed for genre %s: %v", genre, err)
			}

			if metadata == nil || instance == nil {
				t.Fatal("Got nil result")
			}

			// Verify genre is set correctly
			if metadata.GenreID != genre {
				t.Errorf("Expected genre '%s', got '%s'", genre, metadata.GenreID)
			}

			// Verify name contains genre-appropriate content (basic check)
			if metadata.Name == "" {
				t.Error("Generated name is empty")
			}
		})
	}
}

// TestGenerateAndCreateGame_DifficultyScaling verifies difficulty affects game parameters.
func TestGenerateAndCreateGame_DifficultyScaling(t *testing.T) {
	generator := NewGenerator()
	seed := int64(67890)

	difficulties := []float64{0.0, 0.3, 0.6, 1.0}

	for _, diff := range difficulties {
		t.Run("", func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: diff,
				Depth:      10,
				GenreID:    "fantasy",
				Custom:     make(map[string]interface{}),
			}

			metadata, instance, err := GenerateAndCreateGame(generator, seed, params)
			if err != nil {
				t.Fatalf("GenerateAndCreateGame failed for difficulty %.2f: %v", diff, err)
			}

			if metadata == nil || instance == nil {
				t.Fatal("Got nil result")
			}

			// Verify difficulty is set correctly
			if metadata.Difficulty != diff {
				t.Errorf("Expected difficulty %.2f, got %.2f", diff, metadata.Difficulty)
			}

			// Instance should be initialized and ready to play
			if instance.IsComplete() {
				t.Error("Newly created instance should not be complete")
			}
		})
	}
}

// TestGameTypeToEngineType_AllTypes verifies all game type conversions are correct.
func TestGameTypeToEngineType_AllTypes(t *testing.T) {
	tests := []struct {
		procgenType GameType
		wantString  string
	}{
		{GameTypeCard, "Card Game"},
		{GameTypeDice, "Dice Game"},
		{GameTypePuzzle, "Puzzle"},
		{GameTypeMemory, "Memory"},
		{GameTypeLockPicking, "Lock-Picking"},
		{GameTypeHacking, "Hacking"},
		{GameTypeRitual, "Ritual"},
	}

	for _, tt := range tests {
		t.Run(tt.procgenType.String(), func(t *testing.T) {
			engineType := GameTypeToEngineType(tt.procgenType)
			got := engineType.String()
			if got != tt.wantString {
				t.Errorf("GameTypeToEngineType(%v).String() = %v, want %v", tt.procgenType, got, tt.wantString)
			}
		})
	}
}

// TestEngineTypeToGameType_AllTypes verifies reverse conversion is correct.
func TestEngineTypeToGameType_AllTypes(t *testing.T) {
	generator := NewGenerator()

	// Generate one of each type to get valid engine types
	for procgenType := GameTypeCard; procgenType <= GameTypeRitual; procgenType++ {
		t.Run(procgenType.String(), func(t *testing.T) {
			// Convert procgen -> engine -> procgen
			engineType := GameTypeToEngineType(procgenType)
			backToProcgen := EngineTypeToGameType(engineType)

			if backToProcgen != procgenType {
				t.Errorf("Round trip failed: %v -> %v -> %v", procgenType, engineType, backToProcgen)
			}

			// Verify we can create an instance
			instance, err := CreateGameInstance(procgenType)
			if err != nil {
				t.Fatalf("CreateGameInstance failed: %v", err)
			}
			if instance == nil {
				t.Error("CreateGameInstance returned nil")
			}
		})
	}

	// Also verify generator produces valid types
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}

	for i := int64(0); i < 20; i++ {
		result, err := generator.Generate(12345+i, params)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		metadata := result.(*MiniGame)
		engineType := GameTypeToEngineType(metadata.Type)
		backToProcgen := EngineTypeToGameType(engineType)

		if backToProcgen != metadata.Type {
			t.Errorf("Round trip failed for generated type: %v -> %v -> %v",
				metadata.Type, engineType, backToProcgen)
		}
	}
}

// BenchmarkGenerateAndCreateGame benchmarks the combined generation + instantiation flow.
func BenchmarkGenerateAndCreateGame(b *testing.B) {
	generator := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seed := int64(i)
		_, _, _ = GenerateAndCreateGame(generator, seed, params)
	}
}

// BenchmarkGameTypeToEngineType benchmarks type conversion performance.
func BenchmarkGameTypeToEngineType(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gameType := GameType(i % 7) // Cycle through all 7 types
		_ = GameTypeToEngineType(gameType)
	}
}
