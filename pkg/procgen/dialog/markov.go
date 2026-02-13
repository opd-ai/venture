package dialog

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
)

// MarkovOrder defines the n-gram size for the Markov chain.
// Order 2 uses 2-word prefixes, Order 3 uses 3-word prefixes.
// Higher order produces more coherent but less varied text.
type MarkovOrder int

const (
	// Order2 uses 2-word prefixes for Markov chain generation.
	// Provides good balance between coherence and variation.
	Order2 MarkovOrder = 2

	// Order3 uses 3-word prefixes for Markov chain generation.
	// Produces more coherent but less varied text.
	Order3 MarkovOrder = 3
)

// GenerateParams contains parameters for dialog generation.
type GenerateParams struct {
	// PlayerInput is the player's message or question.
	PlayerInput string

	// ConversationID uniquely identifies the conversation thread.
	ConversationID string

	// MaxWords limits the response length (default: 30).
	MaxWords int

	// MinWords ensures minimum response length (default: 10).
	MinWords int

	// Temperature controls randomness (0.0=deterministic, 1.0=max variation).
	// Default: 0.7
	Temperature float64
}

// MarkovGenerator generates text using Markov chains.
// Supports both deterministic (seeded) and non-deterministic (runtime entropy) modes.
type MarkovGenerator struct {
	genreID      string
	order        MarkovOrder
	chain        map[string][]string // prefix -> [possible next words]
	prefixStarts []string            // valid sentence-starting prefixes
	seed         int64               // base seed for deterministic mode
	rng          *rand.Rand          // seeded RNG instance
}

// NewMarkovGenerator creates a new Markov chain text generator.
//
// The seed parameter is used for:
//   - Deterministic mode: Seed is used directly for reproducible generation
//   - Non-deterministic mode: Seed is combined with runtime entropy
//
// The genreID parameter should match one of the supported genres:
// "fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"
//
// The order parameter controls n-gram size (Order2 or Order3).
func NewMarkovGenerator(seed int64, genreID string, order MarkovOrder) *MarkovGenerator {
	if order < Order2 || order > Order3 {
		order = Order2 // Default to order 2 if invalid
	}

	return &MarkovGenerator{
		genreID:      genreID,
		order:        order,
		chain:        make(map[string][]string),
		prefixStarts: make([]string, 0, 100),
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
	}
}

// TrainFromCorpus builds the Markov chain from training text.
//
// Training text should be an array of sentences. The function:
//  1. Tokenizes each sentence into words
//  2. Builds n-gram prefixes based on the generator's order
//  3. Records possible next words for each prefix
//  4. Identifies sentence-starting prefixes
//
// This method should be called once before generating any text.
func (m *MarkovGenerator) TrainFromCorpus(sentences []string) {
	// Reset chain state
	m.chain = make(map[string][]string)
	m.prefixStarts = make([]string, 0, 100)
	seenPrefixes := make(map[string]bool)

	for _, sentence := range sentences {
		words := tokenize(sentence)
		if len(words) < int(m.order)+1 {
			continue // Skip sentences too short to generate n-grams
		}

		// Record sentence-starting prefix
		startPrefix := m.makePrefix(words[:int(m.order)])
		if !seenPrefixes[startPrefix] {
			m.prefixStarts = append(m.prefixStarts, startPrefix)
			seenPrefixes[startPrefix] = true
		}

		// Build n-grams
		for i := 0; i < len(words)-int(m.order); i++ {
			prefix := m.makePrefix(words[i : i+int(m.order)])
			nextWord := words[i+int(m.order)]

			m.chain[prefix] = append(m.chain[prefix], nextWord)
		}
	}
}

// GenerateText creates a response based on parameters.
// This is the internal method used for dialog generation. For the standard Generator interface,
// use the Generate method instead.
//
// Generation process:
//  1. Derives a seed from player input, conversation ID, and timestamp (non-deterministic)
//  2. Selects a starting prefix (sentence beginning)
//  3. Walks the Markov chain, selecting next words based on probabilities
//  4. Stops at MaxWords or natural sentence ending
//  5. Ensures MinWords is met (re-generates if too short)
//
// Returns empty string if generation fails (corpus not trained, invalid params).
func (m *MarkovGenerator) GenerateText(params GenerateParams) string {
	// Apply defaults
	if params.MaxWords <= 0 {
		params.MaxWords = 30
	}
	if params.MinWords <= 0 {
		params.MinWords = 10
	}
	if params.Temperature <= 0 {
		params.Temperature = 0.7
	}

	// Validate chain is trained
	if len(m.chain) == 0 || len(m.prefixStarts) == 0 {
		return "" // No corpus trained
	}

	// Derive runtime seed from context (non-deterministic)
	runtimeSeed := m.deriveRuntimeSeed(params.PlayerInput, params.ConversationID)
	localRNG := rand.New(rand.NewSource(runtimeSeed))

	// Generate response (may retry if too short)
	for attempt := 0; attempt < 5; attempt++ {
		words := m.generateSequence(localRNG, params.MaxWords, params.Temperature)

		if len(words) >= params.MinWords {
			return strings.Join(words, " ")
		}
	}

	// Fallback: return whatever we got
	words := m.generateSequence(localRNG, params.MaxWords, params.Temperature)
	return strings.Join(words, " ")
}

// GenerateDeterministic creates a reproducible response using only the base seed.
//
// This mode is used for testing and when -deterministic-dialog=true is set.
// Unlike Generate(), this does NOT use runtime entropy, so the same parameters
// always produce the same output.
func (m *MarkovGenerator) GenerateDeterministic(params GenerateParams) string {
	// Apply defaults
	if params.MaxWords <= 0 {
		params.MaxWords = 30
	}
	if params.MinWords <= 0 {
		params.MinWords = 10
	}
	if params.Temperature <= 0 {
		params.Temperature = 0.7
	}

	// Validate chain is trained
	if len(m.chain) == 0 || len(m.prefixStarts) == 0 {
		return ""
	}

	// Use base seed (deterministic)
	deterministicRNG := rand.New(rand.NewSource(m.seed))

	// Generate response
	for attempt := 0; attempt < 5; attempt++ {
		words := m.generateSequence(deterministicRNG, params.MaxWords, params.Temperature)

		if len(words) >= params.MinWords {
			return strings.Join(words, " ")
		}
	}

	// Fallback
	words := m.generateSequence(deterministicRNG, params.MaxWords, params.Temperature)
	return strings.Join(words, " ")
}

// generateSequence walks the Markov chain to produce a word sequence.
func (m *MarkovGenerator) generateSequence(rng *rand.Rand, maxWords int, temperature float64) []string {
	// Select starting prefix
	if len(m.prefixStarts) == 0 {
		return []string{} // No valid start prefixes
	}
	startPrefix := m.prefixStarts[rng.Intn(len(m.prefixStarts))]

	// Initialize with starting prefix words
	words := strings.Split(startPrefix, " ")

	// Walk chain
	for len(words) < maxWords {
		// Get current prefix (last N words)
		currentPrefix := m.makePrefix(words[len(words)-int(m.order):])

		// Look up possible next words
		nextWords, exists := m.chain[currentPrefix]
		if !exists || len(nextWords) == 0 {
			break // Dead end, terminate sentence
		}

		// Select next word with temperature-adjusted randomness
		nextWord := m.selectNextWord(nextWords, rng, temperature)
		words = append(words, nextWord)

		// Natural sentence ending (period, exclamation, question)
		if strings.HasSuffix(nextWord, ".") || strings.HasSuffix(nextWord, "!") || strings.HasSuffix(nextWord, "?") {
			break
		}
	}

	// Ensure sentence ends with punctuation
	if len(words) > 0 {
		lastWord := words[len(words)-1]
		if !strings.HasSuffix(lastWord, ".") && !strings.HasSuffix(lastWord, "!") && !strings.HasSuffix(lastWord, "?") {
			words[len(words)-1] = lastWord + "."
		}
	}

	return words
}

// selectNextWord chooses the next word from candidates with temperature-adjusted randomness.
//
// Temperature effects:
//   - 0.0: Always select most common word (deterministic)
//   - 0.5: Bias toward common words, allow some variation
//   - 1.0: Uniform random selection (maximum variation)
func (m *MarkovGenerator) selectNextWord(candidates []string, rng *rand.Rand, temperature float64) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	if temperature < 0.01 {
		return selectMostFrequentWord(candidates)
	}

	if temperature > 0.99 {
		return candidates[rng.Intn(len(candidates))]
	}

	return selectWeightedWord(candidates, rng, temperature)
}

// deriveRuntimeSeed creates a deterministic seed from player input and conversation ID.
//
// This ensures reproducible dialog generation (same seed + input = same output):
//   - Different conversations with same input = different responses
//   - Same conversation with same input = same response
//
// The hash combines:
//  1. Base seed (world consistency)
//  2. Player input (context)
//  3. Conversation ID (thread uniqueness)
func (m *MarkovGenerator) deriveRuntimeSeed(playerInput, conversationID string) int64 {
	h := sha256.New()

	// Write base seed - if write fails, use simpler fallback seed derivation
	if err := binary.Write(h, binary.LittleEndian, m.seed); err != nil {
		// Fallback: combine seed with string hashes directly
		return m.seed ^ int64(hash64(playerInput)) ^ int64(hash64(conversationID))
	}

	// Write player input
	h.Write([]byte(playerInput))

	// Write conversation ID
	h.Write([]byte(conversationID))

	// Extract int64 from hash
	hash := h.Sum(nil)
	seed := int64(binary.LittleEndian.Uint64(hash[:8]))

	return seed
}

// makePrefix creates a space-separated prefix string from words.
func (m *MarkovGenerator) makePrefix(words []string) string {
	return strings.Join(words, " ")
}

// GetChainSize returns the number of prefixes in the trained chain.
// Useful for diagnostics and testing.
func (m *MarkovGenerator) GetChainSize() int {
	return len(m.chain)
}

// GetPrefixStartsCount returns the number of valid sentence-starting prefixes.
// Useful for diagnostics and testing.
func (m *MarkovGenerator) GetPrefixStartsCount() int {
	return len(m.prefixStarts)
}

// GetOrder returns the Markov chain order.
func (m *MarkovGenerator) GetOrder() MarkovOrder {
	return m.order
}

// GetGenreID returns the genre identifier.
func (m *MarkovGenerator) GetGenreID() string {
	return m.genreID
}

// Reset clears the trained chain state.
// Useful for re-training with different corpus.
func (m *MarkovGenerator) Reset() {
	m.chain = make(map[string][]string)
	m.prefixStarts = make([]string, 0, 100)
}

// String returns a human-readable description of the generator.
func (m *MarkovGenerator) String() string {
	return fmt.Sprintf("MarkovGenerator{genre=%s, order=%d, chainSize=%d, prefixStarts=%d}",
		m.genreID, m.order, len(m.chain), len(m.prefixStarts))
}

// Generate implements the procgen.Generator interface.
// It generates dialog text using Markov chain generation based on the seed and parameters.
// The genreID in params is used to select appropriate corpus if different from the generator's genre.
//
// Returns a string containing the generated dialog text.
func (m *MarkovGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Validate parameters
	if err := procgen.ValidateParams(params); err != nil {
		return nil, fmt.Errorf("invalid generation parameters: %w", err)
	}

	// Check if corpus is trained
	if len(m.chain) == 0 || len(m.prefixStarts) == 0 {
		return nil, fmt.Errorf("markov chain not trained, call TrainFromCorpus first")
	}

	// Create dialog generation params from procgen params
	// Use difficulty to adjust length: higher difficulty = longer responses
	// Use depth to adjust complexity: higher depth = more temperature variation
	minWords := 10 + int(params.Difficulty*10)     // 10-20 words
	maxWords := 30 + int(params.Difficulty*20)     // 30-50 words
	temperature := 0.5 + (params.Difficulty * 0.3) // 0.5-0.8 temperature

	// Add depth-based variation
	if params.Depth > 0 {
		// Deeper dungeons = slightly longer dialog
		maxWords += params.Depth
		if maxWords > 100 {
			maxWords = 100 // Cap at 100 words
		}
	}

	dialogParams := GenerateParams{
		PlayerInput:    "greeting",                  // Default greeting
		ConversationID: fmt.Sprintf("gen-%d", seed), // Unique conversation ID
		MaxWords:       maxWords,
		MinWords:       minWords,
		Temperature:    temperature,
	}

	// Generate dialog using deterministic mode (seed-based)
	result := m.GenerateDeterministic(dialogParams)

	// Validate non-empty result
	if result == "" {
		return nil, fmt.Errorf("failed to generate dialog text")
	}

	return result, nil
}

// Validate implements the procgen.Generator interface.
// It checks if the generated dialog text is valid.
func (m *MarkovGenerator) Validate(result interface{}) error {
	// Type check
	dialogText, ok := result.(string)
	if !ok {
		return fmt.Errorf("result is not a string, got type %T", result)
	}

	// Check for empty result
	if len(dialogText) == 0 {
		return fmt.Errorf("generated dialog text is empty")
	}

	// Check minimum length (should be at least a few words)
	words := strings.Fields(dialogText)
	if len(words) < 3 {
		return fmt.Errorf("generated dialog text has too few words (%d), need at least 3", len(words))
	}

	// Check maximum length (prevent runaway generation)
	if len(words) > 150 {
		return fmt.Errorf("generated dialog text has too many words (%d), maximum is 150", len(words))
	}

	// Check for proper punctuation at end
	lastChar := dialogText[len(dialogText)-1]
	if lastChar != '.' && lastChar != '!' && lastChar != '?' {
		return fmt.Errorf("generated dialog text must end with punctuation (. ! ?), got: %c", lastChar)
	}

	// Validate that text contains only printable ASCII characters and common punctuation
	for _, char := range dialogText {
		if char < 32 || char > 126 {
			// Allow only printable ASCII
			return fmt.Errorf("generated dialog text contains invalid character (code %d)", char)
		}
	}

	return nil
}
