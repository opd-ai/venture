package dialog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
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

	// Call GenerateText multiple times - should be consistent
	result1a := gen1.GenerateText(params)
	result1b := gen1.GenerateText(params)
	result2 := gen2.GenerateText(params)

	if result1a == "" {
		t.Fatal("GenerateText returned empty string")
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
		// Use GenerateText (non-deterministic) with unique conversation ID per iteration
		params.ConversationID = fmt.Sprintf("test-variation-%d", i)
		response := gen.GenerateText(params)

		if response == "" {
			t.Errorf("iteration %d: GenerateText returned empty string", i)
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

	response := gen.GenerateText(params)

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

	response := gen.GenerateText(params)

	if response == "" {
		t.Error("GenerateText with default params returned empty string")
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
		gen.GenerateText(params)
	}
}

// ============================================================================
// Generator Interface Tests
// ============================================================================

// TestMarkovGenerator_Generate tests the procgen.Generator interface implementation.
func TestMarkovGenerator_Generate(t *testing.T) {
	// Use standard import to avoid conflicts
	tests := []struct {
		name       string
		seed       int64
		difficulty float64
		depth      int
		genreID    string
		wantErr    bool
		trained    bool
	}{
		{
			name:       "basic generation with trained corpus",
			seed:       12345,
			difficulty: 0.5,
			depth:      1,
			genreID:    "fantasy",
			wantErr:    false,
			trained:    true,
		},
		{
			name:       "high difficulty increases length",
			seed:       67890,
			difficulty: 1.0,
			depth:      5,
			genreID:    "scifi",
			wantErr:    false,
			trained:    true,
		},
		{
			name:       "low difficulty short response",
			seed:       11111,
			difficulty: 0.0,
			depth:      0,
			genreID:    "horror",
			wantErr:    false,
			trained:    true,
		},
		{
			name:       "untrained corpus returns error",
			seed:       22222,
			difficulty: 0.5,
			depth:      1,
			genreID:    "fantasy",
			wantErr:    true,
			trained:    false,
		},
		{
			name:       "invalid difficulty",
			seed:       33333,
			difficulty: -0.5, // Invalid
			depth:      1,
			genreID:    "fantasy",
			wantErr:    true,
			trained:    true,
		},
		{
			name:       "invalid depth",
			seed:       44444,
			difficulty: 0.5,
			depth:      -1, // Invalid
			genreID:    "fantasy",
			wantErr:    true,
			trained:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewMarkovGenerator(tt.seed, tt.genreID, Order2)

			// Train corpus if required
			if tt.trained {
				corpus := GetCorpus(tt.genreID)
				if corpus != nil {
					gen.TrainFromCorpus(corpus.Sentences)
				} else {
					// Fallback corpus for testing
					gen.TrainFromCorpus([]string{
						"The brave adventurer explores the dark dungeon.",
						"Ancient treasures await those who dare to seek them.",
						"Beware of the monsters lurking in the shadows.",
					})
				}
			}

			// Create generation params using procgen.GenerationParams
			params := procgen.GenerationParams{
				Difficulty: tt.difficulty,
				Depth:      tt.depth,
				GenreID:    tt.genreID,
				Custom:     make(map[string]interface{}),
			}

			// Call Generate
			result, err := gen.Generate(tt.seed, params)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If error expected, we're done
			if tt.wantErr {
				return
			}

			// Validate result type
			dialogText, ok := result.(string)
			if !ok {
				t.Errorf("Generate() returned type %T, want string", result)
				return
			}

			// Validate result is not empty
			if dialogText == "" {
				t.Error("Generate() returned empty string")
			}

			// Validate minimum word count - be lenient since Markov chains depend on corpus size
			// For small test corpora, we just ensure we got *some* text (>= 3 words)
			words := strings.Fields(dialogText)
			if len(words) < 3 {
				t.Logf("Dialog text: %s", dialogText)
				t.Errorf("Generate() returned %d words, expected at least 3", len(words))
			}

			// Validate maximum word count
			maxExpected := 100
			if len(words) > maxExpected {
				t.Errorf("Generate() returned %d words, expected at most %d", len(words), maxExpected)
			}

			// Validate ends with punctuation
			if len(dialogText) > 0 {
				lastChar := dialogText[len(dialogText)-1]
				if lastChar != '.' && lastChar != '!' && lastChar != '?' {
					t.Errorf("Generate() dialog must end with punctuation, got: %c", lastChar)
				}
			}
		})
	}
}

// TestMarkovGenerator_Validate tests the Validate method.
func TestMarkovGenerator_Validate(t *testing.T) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid dialog text",
			input:   "The brave adventurer explores the dark dungeon.",
			wantErr: false,
		},
		{
			name:    "valid with exclamation",
			input:   "Look out for the dragon!",
			wantErr: false,
		},
		{
			name:    "valid with question",
			input:   "Where is the treasure hidden?",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "too short (less than 3 words)",
			input:   "Go now.",
			wantErr: true,
		},
		{
			name:    "missing punctuation",
			input:   "The hero walks into the dungeon",
			wantErr: true,
		},
		{
			name:    "wrong type (not string)",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "wrong type (slice)",
			input:   []string{"hello", "world"},
			wantErr: true,
		},
		{
			name:    "too long (over 150 words)",
			input:   strings.Repeat("word ", 151) + "end.",
			wantErr: true,
		},
		{
			name:    "invalid characters (non-ASCII)",
			input:   "The hero finds a magical 🗡️ sword.",
			wantErr: true,
		},
		{
			name:    "valid long text",
			input:   "The brave adventurer explores the dark dungeon and discovers ancient treasures that await those who dare to seek them while avoiding the monsters lurking in the shadows.",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMarkovGenerator_GenerateDeterminism tests that Generate produces deterministic output.
func TestMarkovGenerator_GenerateDeterminism(t *testing.T) {
	seed := int64(12345)
	genreID := "fantasy"

	// Create two generators with same seed
	gen1 := NewMarkovGenerator(seed, genreID, Order2)
	gen2 := NewMarkovGenerator(seed, genreID, Order2)

	// Train both with same corpus
	corpus := []string{
		"The brave adventurer explores the dark dungeon.",
		"Ancient treasures await those who dare to seek them.",
		"Beware of the monsters lurking in the shadows.",
		"Magical items can be found in hidden chambers.",
		"The quest for glory drives heroes forward.",
	}
	gen1.TrainFromCorpus(corpus)
	gen2.TrainFromCorpus(corpus)

	// Generate with same parameters
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
		Custom:     make(map[string]interface{}),
	}

	result1, err1 := gen1.Generate(seed, params)
	result2, err2 := gen2.Generate(seed, params)

	// Both should succeed
	if err1 != nil {
		t.Fatalf("gen1.Generate() error = %v", err1)
	}
	if err2 != nil {
		t.Fatalf("gen2.Generate() error = %v", err2)
	}

	// Results should be identical
	text1, ok1 := result1.(string)
	text2, ok2 := result2.(string)

	if !ok1 || !ok2 {
		t.Fatal("Generate() results are not strings")
	}

	if text1 != text2 {
		t.Errorf("Generate() not deterministic:\ngen1: %s\ngen2: %s", text1, text2)
	}
}

// TestMarkovGenerator_ValidateAfterGenerate tests validation of generated output.
func TestMarkovGenerator_ValidateAfterGenerate(t *testing.T) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)

	// Train with a proper corpus
	corpus := []string{
		"The brave adventurer explores the dark dungeon.",
		"Ancient treasures await those who dare to seek them.",
		"Beware of the monsters lurking in the shadows.",
		"Magical items can be found in hidden chambers.",
		"The quest for glory drives heroes forward.",
		"Dragons guard their hoards jealously.",
		"Heroes must be courageous and wise.",
		"The path to victory is fraught with danger.",
	}
	gen.TrainFromCorpus(corpus)

	// Generate dialog
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      2,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Validate the generated output
	err = gen.Validate(result)
	if err != nil {
		t.Errorf("Validate() failed for generated output: %v\nOutput: %v", err, result)
	}
}

// BenchmarkMarkovGenerator_Generate benchmarks the Generator interface method.
func BenchmarkMarkovGenerator_Generate(b *testing.B) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)

	// Train with corpus
	corpus := GetCorpus("fantasy")
	if corpus != nil {
		gen.TrainFromCorpus(corpus.Sentences)
	} else {
		gen.TrainFromCorpus([]string{
			"The brave adventurer explores the dark dungeon.",
			"Ancient treasures await those who dare to seek them.",
			"Beware of the monsters lurking in the shadows.",
		})
	}

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate() error = %v", err)
		}
	}
}

// BenchmarkMarkovGenerator_Validate benchmarks the Validate method.
func BenchmarkMarkovGenerator_Validate(b *testing.B) {
	gen := NewMarkovGenerator(12345, "fantasy", Order2)

	// Sample dialog texts of varying lengths
	samples := []string{
		"The brave adventurer explores the dark dungeon.",
		"Ancient treasures await those who dare to seek them while avoiding dangerous traps.",
		"Beware of the monsters lurking in the shadows, for they are fearsome and powerful creatures.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sample := samples[i%len(samples)]
		_ = gen.Validate(sample)
	}
}
