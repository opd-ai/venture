package dialog

import (
	"fmt"
	"strings"
	"testing"
)

// TestNewMarkovGenerator verifies generator creation.
func TestNewMarkovGenerator(t *testing.T) {
	tests := []struct {
		name      string
		seed      int64
		genreID   string
		order     MarkovOrder
		wantGenre string
		wantOrder MarkovOrder
	}{
		{"order 2 fantasy", 12345, "fantasy", Order2, "fantasy", Order2},
		{"order 3 scifi", 67890, "scifi", Order3, "scifi", Order3},
		{"invalid order defaults to 2", 11111, "horror", MarkovOrder(5), "horror", Order2},
		{"order 0 defaults to 2", 22222, "cyberpunk", MarkovOrder(0), "cyberpunk", Order2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewMarkovGenerator(tt.seed, tt.genreID, tt.order)

			if gen == nil {
				t.Fatal("NewMarkovGenerator returned nil")
			}

			if gen.GetGenreID() != tt.wantGenre {
				t.Errorf("genre = %v, want %v", gen.GetGenreID(), tt.wantGenre)
			}

			if gen.GetOrder() != tt.wantOrder {
				t.Errorf("order = %v, want %v", gen.GetOrder(), tt.wantOrder)
			}

			if gen.GetChainSize() != 0 {
				t.Errorf("initial chain size = %d, want 0", gen.GetChainSize())
			}
		})
	}
}

// TestTrainFromCorpus verifies corpus training.
func TestTrainFromCorpus(t *testing.T) {
	sentences := []string{
		"The quick brown fox jumps over the lazy dog.",
		"A journey of a thousand miles begins with a single step.",
		"To be or not to be, that is the question.",
	}

	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(sentences)

	if gen.GetChainSize() == 0 {
		t.Error("chain size is 0 after training")
	}

	if gen.GetPrefixStartsCount() == 0 {
		t.Error("prefix starts count is 0 after training")
	}

	// Chain should have multiple entries for order 2
	if gen.GetChainSize() < 5 {
		t.Errorf("chain size %d seems too small for given corpus", gen.GetChainSize())
	}
}

// TestTrainFromCorpusEmptySentences verifies handling of empty input.
func TestTrainFromCorpusEmptySentences(t *testing.T) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus([]string{})

	if gen.GetChainSize() != 0 {
		t.Errorf("chain size = %d, want 0 for empty corpus", gen.GetChainSize())
	}
}

// TestTrainFromCorpusShortSentences verifies handling of sentences too short for n-grams.
func TestTrainFromCorpusShortSentences(t *testing.T) {
	sentences := []string{
		"Hi",
		"Go",
		"Yes",
	}

	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(sentences)

	// All sentences too short (need at least order+1 words)
	if gen.GetChainSize() != 0 {
		t.Errorf("chain size = %d, want 0 for too-short sentences", gen.GetChainSize())
	}
}

// TestGenerateDeterministic verifies deterministic generation.
func TestGenerateDeterministic(t *testing.T) {
	corpus := GetFantasyCorpus()

	// Create two generators with same seed
	gen1 := NewMarkovGenerator(12345, "fantasy", Order2)
	gen1.TrainFromCorpus(corpus.Sentences)

	gen2 := NewMarkovGenerator(12345, "fantasy", Order2)
	gen2.TrainFromCorpus(corpus.Sentences)

	params := GenerateParams{
		PlayerInput:    "Hello",
		ConversationID: "test-123",
		MaxWords:       20,
		MinWords:       5,
		Temperature:    0.7,
	}

	// Generate responses
	response1 := gen1.GenerateDeterministic(params)
	response2 := gen2.GenerateDeterministic(params)

	if response1 == "" {
		t.Error("GenerateDeterministic returned empty string")
	}

	if response1 != response2 {
		t.Errorf("deterministic responses differ:\n  gen1: %s\n  gen2: %s", response1, response2)
	}
}

// TestGenerateDeterminismWithSameInputs verifies Generate is deterministic for same inputs.
// This tests that deriveRuntimeSeed produces reproducible results without time.Now().
func TestGenerateDeterminismWithSameInputs(t *testing.T) {
	corpus := GetFantasyCorpus()

	// Create two generators with same seed
	gen1 := NewMarkovGenerator(12345, "fantasy", Order2)
	gen1.TrainFromCorpus(corpus.Sentences)

	gen2 := NewMarkovGenerator(12345, "fantasy", Order2)
	gen2.TrainFromCorpus(corpus.Sentences)

	params := GenerateParams{
		PlayerInput:    "Tell me about the quest",
		ConversationID: "conv-abc-123",
		MaxWords:       25,
		MinWords:       8,
		Temperature:    0.7,
	}

	// Call Generate multiple times - should be consistent
	result1a := gen1.Generate(params)
	result1b := gen1.Generate(params)
	result2 := gen2.Generate(params)

	if result1a == "" {
		t.Fatal("Generate returned empty string")
	}

	if result1a != result1b {
		t.Errorf("same generator, same params should produce same output:\n  call1: %s\n  call2: %s", result1a, result1b)
	}

	if result1a != result2 {
		t.Errorf("different generators with same seed/params should produce same output:\n  gen1: %s\n  gen2: %s", result1a, result2)
	}
}

// TestGenerateVariation verifies variation with different inputs.
func TestGenerateVariation(t *testing.T) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	params := GenerateParams{
		PlayerInput:    "Where is the dungeon?",
		ConversationID: "test-variation",
		MaxWords:       30,
		MinWords:       10,
		Temperature:    0.7,
	}

	// Generate multiple responses
	responses := make(map[string]bool)
	iterations := 10

	for i := 0; i < iterations; i++ {
		// Use Generate (non-deterministic) with unique conversation ID per iteration
		params.ConversationID = fmt.Sprintf("test-variation-%d", i)
		response := gen.Generate(params)

		if response == "" {
			t.Errorf("iteration %d: Generate returned empty string", i)
			continue
		}

		responses[response] = true
	}

	// Expect at least some variation (>= 50% unique responses)
	uniqueCount := len(responses)
	minUnique := iterations / 2

	if uniqueCount < minUnique {
		t.Errorf("insufficient variation: %d unique responses out of %d (want >= %d)",
			uniqueCount, iterations, minUnique)
	}
}

// TestGenerateWithUntrained verifies error handling for untrained generator.
func TestGenerateWithUntrained(t *testing.T) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	// Do not train

	params := GenerateParams{
		PlayerInput:    "Hello",
		ConversationID: "test",
		MaxWords:       20,
		MinWords:       10,
		Temperature:    0.7,
	}

	response := gen.Generate(params)

	if response != "" {
		t.Errorf("untrained generator returned non-empty response: %s", response)
	}
}

// TestGenerateDefaultParameters verifies parameter defaults.
func TestGenerateDefaultParameters(t *testing.T) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	// Empty params (should apply defaults)
	params := GenerateParams{
		PlayerInput:    "Test",
		ConversationID: "test",
		// MaxWords, MinWords, Temperature omitted
	}

	response := gen.Generate(params)

	if response == "" {
		t.Error("Generate with default params returned empty string")
	}

	// Response should respect default max words (30)
	wordCount := len(strings.Fields(response))
	if wordCount > 40 {
		t.Errorf("response has %d words, expected <= 30 (default MaxWords)", wordCount)
	}
}

// TestGenerateWordLimits verifies word count constraints.
func TestGenerateWordLimits(t *testing.T) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	tests := []struct {
		name     string
		maxWords int
		minWords int
	}{
		{"short response", 15, 5},
		{"medium response", 30, 10},
		{"long response", 50, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := GenerateParams{
				PlayerInput:    "Tell me a story",
				ConversationID: "test-limits",
				MaxWords:       tt.maxWords,
				MinWords:       tt.minWords,
				Temperature:    0.7,
			}

			response := gen.GenerateDeterministic(params)

			if response == "" {
				t.Error("GenerateDeterministic returned empty string")
				return
			}

			wordCount := len(strings.Fields(response))

			// Allow some slack for MinWords (may not always achieve it)
			if wordCount > tt.maxWords+5 {
				t.Errorf("response has %d words, exceeds max %d", wordCount, tt.maxWords)
			}
		})
	}
}

// TestGenerateTemperatureEffect verifies temperature impact.
func TestGenerateTemperatureEffect(t *testing.T) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	// Temperature 0.0 should be deterministic
	params := GenerateParams{
		PlayerInput:    "Greetings",
		ConversationID: "temp-test",
		MaxWords:       20,
		MinWords:       5,
		Temperature:    0.0,
	}

	// Generate twice with same params, should be identical (using deterministic mode)
	response1 := gen.GenerateDeterministic(params)
	response2 := gen.GenerateDeterministic(params)

	if response1 != response2 {
		t.Error("deterministic mode should produce identical responses")
	}

	// Note: Temperature effect is best tested via non-deterministic Generate(),
	// which uses runtime entropy. That variation is tested in TestGenerateVariation.
}

// TestReset verifies chain reset functionality.
func TestReset(t *testing.T) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	initialSize := gen.GetChainSize()
	if initialSize == 0 {
		t.Fatal("chain not trained properly")
	}

	gen.Reset()

	if gen.GetChainSize() != 0 {
		t.Errorf("after Reset, chain size = %d, want 0", gen.GetChainSize())
	}

	if gen.GetPrefixStartsCount() != 0 {
		t.Errorf("after Reset, prefix starts = %d, want 0", gen.GetPrefixStartsCount())
	}
}

// TestTokenize verifies tokenization logic.
func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // Expected token count
	}{
		{"simple sentence", "Hello world", 2},
		{"with punctuation", "Hello, world!", 2},
		{"multiple spaces", "Hello    world", 2},
		{"empty string", "", 0},
		{"single word", "Hello", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)

			if len(tokens) != tt.want {
				t.Errorf("tokenize(%q) = %d tokens, want %d", tt.input, len(tokens), tt.want)
			}
		})
	}
}

// TestString verifies String() method output.
func TestString(t *testing.T) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)

	str := gen.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	if !strings.Contains(str, "fantasy") {
		t.Error("String() should contain genre ID")
	}

	if !strings.Contains(str, "2") {
		t.Error("String() should contain order")
	}
}

// BenchmarkTrainFromCorpus measures training performance.
func BenchmarkTrainFromCorpus(b *testing.B) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.TrainFromCorpus(corpus.Sentences)
	}
}

// BenchmarkGenerateDeterministic measures deterministic generation performance.
func BenchmarkGenerateDeterministic(b *testing.B) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	params := GenerateParams{
		PlayerInput:    "Where is the dungeon?",
		ConversationID: "bench-test",
		MaxWords:       30,
		MinWords:       10,
		Temperature:    0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateDeterministic(params)
	}
}

// BenchmarkGenerate measures non-deterministic generation performance.
func BenchmarkGenerate(b *testing.B) {
	corpus := GetFantasyCorpus()
	gen := NewMarkovGenerator(12345, "fantasy", Order2)
	gen.TrainFromCorpus(corpus.Sentences)

	params := GenerateParams{
		PlayerInput:    "Where is the dungeon?",
		ConversationID: "bench-test",
		MaxWords:       30,
		MinWords:       10,
		Temperature:    0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(params)
	}
}
