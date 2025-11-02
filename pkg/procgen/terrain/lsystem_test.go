package terrain

import (
	"strings"
	"testing"
)

func TestSymbolConstants(t *testing.T) {
	tests := []struct {
		name   string
		symbol Symbol
		char   rune
	}{
		{"Start", SymbolStart, 'S'},
		{"End", SymbolEnd, 'E'},
		{"Combat", SymbolCombat, 'C'},
		{"Treasure", SymbolTreasure, 'T'},
		{"Puzzle", SymbolPuzzle, 'P'},
		{"Corridor", SymbolCorridor, '-'},
		{"Branch", SymbolBranch, '+'},
		{"Shop", SymbolShop, '$'},
		{"Rest", SymbolRest, 'R'},
		{"Secret", SymbolSecret, '?'},
		{"Empty", SymbolEmpty, '.'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rune(tt.symbol) != tt.char {
				t.Errorf("Symbol %s = %c, want %c", tt.name, tt.symbol, tt.char)
			}
		})
	}
}

func TestNewLSystemGenerator(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		Seed:         12345,
		MinRoomCount: 5,
		MaxRoomCount: 20,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-C", Weight: 1.0},
		},
	}

	gen := NewLSystemGenerator(config)

	if gen == nil {
		t.Fatal("NewLSystemGenerator returned nil")
	}

	if gen.config.Axiom != config.Axiom {
		t.Errorf("Axiom = %s, want %s", gen.config.Axiom, config.Axiom)
	}

	if gen.config.Seed != config.Seed {
		t.Errorf("Seed = %d, want %d", gen.config.Seed, config.Seed)
	}

	if gen.rng == nil {
		t.Error("RNG not initialized")
	}
}

func TestLSystemGenerator_Generate_SimpleRule(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   2,
		Seed:         12345,
		MinRoomCount: 1,
		MaxRoomCount: 100,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-C", Weight: 1.0},
			{From: SymbolCombat, To: "C-E", Weight: 1.0},
		},
	}

	gen := NewLSystemGenerator(config)
	result := gen.Generate()

	// Iteration 0: S
	// Iteration 1: S-C
	// Iteration 2: S-C-C-E
	expected := "S-C-C-E"
	if result != expected {
		t.Errorf("Generate() = %s, want %s", result, expected)
	}
}

func TestLSystemGenerator_Generate_NoRules(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "SCT",
		Iterations:   3,
		Seed:         12345,
		MinRoomCount: 1,
		MaxRoomCount: 100,
		Rules:        []ProductionRule{}, // No rules
	}

	gen := NewLSystemGenerator(config)
	result := gen.Generate()

	// Without rules, string should remain unchanged
	if result != config.Axiom {
		t.Errorf("Generate() with no rules = %s, want %s", result, config.Axiom)
	}
}

func TestLSystemGenerator_Generate_TerminalSymbols(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   2,
		Seed:         12345,
		MinRoomCount: 1,
		MaxRoomCount: 100,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-T", Weight: 1.0},
			// T (treasure) has no rule - terminal
		},
	}

	gen := NewLSystemGenerator(config)
	result := gen.Generate()

	// Iteration 0: S
	// Iteration 1: S-T (S expands to S-T, T stays as T since no rule)
	// Iteration 2: S-T-T (S expands again, T stays, new T added)
	// Should contain S and at least one T
	if !strings.Contains(result, "S") {
		t.Errorf("Generate() = %s, should contain S", result)
	}
	if !strings.Contains(result, "T") {
		t.Errorf("Generate() = %s, should contain T", result)
	}

	// Count rooms - should be at least 2
	roomCount := gen.countRooms(result)
	if roomCount < 2 {
		t.Errorf("countRooms() = %d, want at least 2", roomCount)
	}
}

func TestLSystemGenerator_Generate_MaxRoomLimit(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   10, // Many iterations
		Seed:         12345,
		MinRoomCount: 1,
		MaxRoomCount: 5, // But limit rooms
		Rules: []ProductionRule{
			{From: SymbolStart, To: "SC", Weight: 1.0}, // No corridor to avoid confusion
			{From: SymbolCombat, To: "CC", Weight: 1.0}, // Exponential expansion
		},
	}

	gen := NewLSystemGenerator(config)
	result := gen.Generate()

	roomCount := gen.countRooms(result)
	// Should stop at or just slightly above max (due to expansion in one iteration)
	if roomCount > config.MaxRoomCount+5 {
		t.Errorf("countRooms() = %d, significantly exceeds MaxRoomCount %d", roomCount, config.MaxRoomCount)
	}
}

func TestLSystemGenerator_Generate_MinRoomRequirement(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   1, // Single iteration
		Seed:         12345,
		MinRoomCount: 10, // But require 10 rooms
		MaxRoomCount: 50,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-C", Weight: 1.0},
			{From: SymbolCombat, To: "C-C", Weight: 1.0},
		},
	}

	gen := NewLSystemGenerator(config)
	result := gen.Generate()

	roomCount := gen.countRooms(result)
	if roomCount < config.MinRoomCount {
		t.Errorf("countRooms() = %d, below MinRoomCount %d", roomCount, config.MinRoomCount)
	}
}

func TestLSystemGenerator_Generate_Determinism(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		Seed:         67890,
		MinRoomCount: 5,
		MaxRoomCount: 20,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-C", Weight: 0.5},
			{From: SymbolStart, To: "S-P", Weight: 0.5},
			{From: SymbolCombat, To: "C-T", Weight: 1.0},
			{From: SymbolPuzzle, To: "P-?", Weight: 1.0},
		},
	}

	// Generate twice with same seed
	gen1 := NewLSystemGenerator(config)
	result1 := gen1.Generate()

	gen2 := NewLSystemGenerator(config)
	result2 := gen2.Generate()

	if result1 != result2 {
		t.Errorf("Non-deterministic generation:\nFirst:  %s\nSecond: %s", result1, result2)
	}
}

func TestLSystemGenerator_Generate_DifferentSeeds(t *testing.T) {
	baseConfig := LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		MinRoomCount: 5,
		MaxRoomCount: 20,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-C", Weight: 0.5},
			{From: SymbolStart, To: "S-P", Weight: 0.5},
		},
	}

	config1 := baseConfig
	config1.Seed = 11111
	gen1 := NewLSystemGenerator(config1)
	result1 := gen1.Generate()

	config2 := baseConfig
	config2.Seed = 22222
	gen2 := NewLSystemGenerator(config2)
	result2 := gen2.Generate()

	// Different seeds should (very likely) produce different results
	// Note: There's a tiny chance they could be the same, but extremely unlikely
	if result1 == result2 {
		t.Logf("Warning: Different seeds produced same result (unlikely but possible)")
	}
}

func TestLSystemGenerator_CountRooms(t *testing.T) {
	gen := NewLSystemGenerator(LSystemConfig{Seed: 123})

	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"Empty", "", 0},
		{"Only corridors", "----", 0},
		{"Single room", "S", 1},
		{"Multiple rooms", "S-C-T", 3},
		{"Rooms and corridors", "S-C-P-?-E", 5},
		{"With branches", "S-+-C-T", 3},
		{"Complex", "S-C-+-P-?-T-E", 6},
		{"Non-room symbols", ".-+-.-", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := gen.countRooms(tt.input)
			if count != tt.expect {
				t.Errorf("countRooms(%q) = %d, want %d", tt.input, count, tt.expect)
			}
		})
	}
}

func TestLSystemGenerator_IsRoomSymbol(t *testing.T) {
	gen := NewLSystemGenerator(LSystemConfig{Seed: 123})

	tests := []struct {
		name   string
		symbol Symbol
		isRoom bool
	}{
		{"Start", SymbolStart, true},
		{"End", SymbolEnd, true},
		{"Combat", SymbolCombat, true},
		{"Treasure", SymbolTreasure, true},
		{"Puzzle", SymbolPuzzle, true},
		{"Shop", SymbolShop, true},
		{"Rest", SymbolRest, true},
		{"Secret", SymbolSecret, true},
		{"Corridor", SymbolCorridor, false},
		{"Branch", SymbolBranch, false},
		{"Empty", SymbolEmpty, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.isRoomSymbol(tt.symbol)
			if result != tt.isRoom {
				t.Errorf("isRoomSymbol(%c) = %v, want %v", tt.symbol, result, tt.isRoom)
			}
		})
	}
}

func TestLSystemGenerator_StochasticRules(t *testing.T) {
	config := LSystemConfig{
		Axiom:        "S",
		Iterations:   1,
		Seed:         12345,
		MinRoomCount: 1,
		MaxRoomCount: 100,
		Rules: []ProductionRule{
			{From: SymbolStart, To: "S-C", Weight: 0.8},
			{From: SymbolStart, To: "S-P", Weight: 0.2},
		},
	}

	gen := NewLSystemGenerator(config)
	result := gen.Generate()

	// Should have chosen one of the two expansions
	if !strings.Contains(result, "S-C") && !strings.Contains(result, "S-P") {
		t.Errorf("Generate() = %s, expected S-C or S-P", result)
	}
}

func TestGetFantasyConfig(t *testing.T) {
	seed := int64(12345)
	config := GetFantasyConfig(seed)

	if config.Axiom != "S" {
		t.Errorf("Axiom = %s, want S", config.Axiom)
	}

	if config.Seed != seed {
		t.Errorf("Seed = %d, want %d", config.Seed, seed)
	}

	if len(config.Rules) == 0 {
		t.Error("No rules defined for fantasy config")
	}

	if config.MinRoomCount <= 0 {
		t.Errorf("MinRoomCount = %d, want > 0", config.MinRoomCount)
	}

	if config.MaxRoomCount <= config.MinRoomCount {
		t.Errorf("MaxRoomCount = %d, should be > MinRoomCount %d", config.MaxRoomCount, config.MinRoomCount)
	}
}

func TestGetSciFiConfig(t *testing.T) {
	seed := int64(12345)
	config := GetSciFiConfig(seed)

	if config.Axiom != "S" {
		t.Errorf("Axiom = %s, want S", config.Axiom)
	}

	if len(config.Rules) == 0 {
		t.Error("No rules defined for sci-fi config")
	}

	// Sci-fi should have more rooms (modular design)
	if config.MaxRoomCount < 20 {
		t.Errorf("MaxRoomCount = %d, sci-fi should support many rooms", config.MaxRoomCount)
	}
}

func TestGetHorrorConfig(t *testing.T) {
	seed := int64(12345)
	config := GetHorrorConfig(seed)

	if config.Axiom != "S" {
		t.Errorf("Axiom = %s, want S", config.Axiom)
	}

	if len(config.Rules) == 0 {
		t.Error("No rules defined for horror config")
	}

	// Horror should have fewer, more linear rooms
	if config.MaxRoomCount > 20 {
		t.Errorf("MaxRoomCount = %d, horror should be more compact", config.MaxRoomCount)
	}
}

func TestGetCyberpunkConfig(t *testing.T) {
	seed := int64(12345)
	config := GetCyberpunkConfig(seed)

	if config.Axiom != "S" {
		t.Errorf("Axiom = %s, want S", config.Axiom)
	}

	if len(config.Rules) == 0 {
		t.Error("No rules defined for cyberpunk config")
	}

	// Should have shop-related rules (black market)
	hasShopRule := false
	for _, rule := range config.Rules {
		if strings.Contains(rule.To, "$") {
			hasShopRule = true
			break
		}
	}

	if !hasShopRule {
		t.Error("Cyberpunk config should include shop-related rules")
	}
}

func TestGetPostApocalypticConfig(t *testing.T) {
	seed := int64(12345)
	config := GetPostApocalypticConfig(seed)

	if config.Axiom != "S" {
		t.Errorf("Axiom = %s, want S", config.Axiom)
	}

	if len(config.Rules) == 0 {
		t.Error("No rules defined for post-apocalyptic config")
	}

	// Should have treasure/loot focus
	hasTreasureRule := false
	for _, rule := range config.Rules {
		if strings.Contains(rule.To, "T") {
			hasTreasureRule = true
			break
		}
	}

	if !hasTreasureRule {
		t.Error("Post-apocalyptic config should include treasure-related rules")
	}
}

func TestGetConfigForGenre(t *testing.T) {
	seed := int64(12345)

	tests := []struct {
		name           string
		genre          string
		minRooms       int
		maxRooms       int
		expectDefault  bool
	}{
		{"Fantasy", "fantasy", 8, 20, false},
		{"SciFi", "sci-fi", 10, 25, false},
		{"SciFi Alt", "scifi", 10, 25, false},
		{"Horror", "horror", 6, 15, false},
		{"Cyberpunk", "cyberpunk", 12, 25, false},
		{"PostApoc", "post-apocalyptic", 8, 18, false},
		{"PostApoc Alt", "postapocalyptic", 8, 18, false},
		{"Unknown", "unknown", 8, 20, true}, // Defaults to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetConfigForGenre(tt.genre, seed)

			if config.Seed != seed {
				t.Errorf("Seed = %d, want %d", config.Seed, seed)
			}

			if config.MinRoomCount != tt.minRooms {
				t.Errorf("MinRoomCount = %d, want %d", config.MinRoomCount, tt.minRooms)
			}

			if config.MaxRoomCount != tt.maxRooms {
				t.Errorf("MaxRoomCount = %d, want %d", config.MaxRoomCount, tt.maxRooms)
			}

			if len(config.Rules) == 0 {
				t.Error("No rules defined")
			}
		})
	}
}

func TestGenreConfigs_Determinism(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	seed := int64(99999)

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := GetConfigForGenre(genre, seed)
			gen1 := NewLSystemGenerator(config)
			result1 := gen1.Generate()

			config = GetConfigForGenre(genre, seed)
			gen2 := NewLSystemGenerator(config)
			result2 := gen2.Generate()

			if result1 != result2 {
				t.Errorf("Non-deterministic generation for %s:\nFirst:  %s\nSecond: %s", genre, result1, result2)
			}
		})
	}
}

func TestGenreConfigs_GenerateValid(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	seed := int64(777)

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := GetConfigForGenre(genre, seed)
			gen := NewLSystemGenerator(config)
			result := gen.Generate()

			if len(result) == 0 {
				t.Errorf("Generated empty string for %s", genre)
			}

			// Should start with S (start room)
			if !strings.HasPrefix(result, "S") {
				t.Errorf("Result doesn't start with S: %s", result)
			}

			// Should have at least minimum rooms
			roomCount := gen.countRooms(result)
			if roomCount < config.MinRoomCount {
				t.Errorf("Room count %d < MinRoomCount %d for %s", roomCount, config.MinRoomCount, genre)
			}

			// Should not exceed maximum rooms
			if roomCount > config.MaxRoomCount {
				t.Errorf("Room count %d > MaxRoomCount %d for %s", roomCount, config.MaxRoomCount, genre)
			}
		})
	}
}

// BenchmarkLSystemGenerator_Generate benchmarks dungeon generation.
func BenchmarkLSystemGenerator_Generate(b *testing.B) {
	config := GetFantasyConfig(12345)
	gen := NewLSystemGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Generate()
	}
}

// BenchmarkLSystemGenerator_GenreConfigs benchmarks all genre configurations.
func BenchmarkLSystemGenerator_GenreConfigs(b *testing.B) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			config := GetConfigForGenre(genre, 12345)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				gen := NewLSystemGenerator(config)
				_ = gen.Generate()
			}
		})
	}
}
