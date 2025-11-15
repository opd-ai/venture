package dialog

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
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

// Generate creates a response based on parameters.
//
// Generation process:
//  1. Derives a seed from player input, conversation ID, and timestamp (non-deterministic)
//  2. Selects a starting prefix (sentence beginning)
//  3. Walks the Markov chain, selecting next words based on probabilities
//  4. Stops at MaxWords or natural sentence ending
//  5. Ensures MinWords is met (re-generates if too short)
//
// Returns empty string if generation fails (corpus not trained, invalid params).
func (m *MarkovGenerator) Generate(params GenerateParams) string {
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

	// Temperature 0.0: deterministic selection (most frequent)
	if temperature < 0.01 {
		// Count frequencies
		freq := make(map[string]int)
		maxFreq := 0
		mostCommon := candidates[0]

		for _, word := range candidates {
			freq[word]++
			if freq[word] > maxFreq {
				maxFreq = freq[word]
				mostCommon = word
			}
		}
		return mostCommon
	}

	// Temperature 1.0: uniform random
	if temperature > 0.99 {
		return candidates[rng.Intn(len(candidates))]
	}

	// Temperature 0.0-1.0: weighted selection
	// Build frequency map
	freq := make(map[string]float64)
	for _, word := range candidates {
		freq[word]++
	}

	// Sort keys for deterministic iteration order
	sortedWords := make([]string, 0, len(freq))
	for word := range freq {
		sortedWords = append(sortedWords, word)
	}
	sort.Strings(sortedWords)

	// Apply temperature to frequencies (higher temp = more uniform)
	weights := make([]float64, 0, len(freq))
	words := make([]string, 0, len(freq))
	totalWeight := 0.0

	for _, word := range sortedWords {
		count := freq[word]
		// Weight = count^(1/temperature)
		// Higher temperature reduces weight differences
		weight := 1.0
		if temperature > 0 {
			weight = count / temperature
		}

		weights = append(weights, weight)
		words = append(words, word)
		totalWeight += weight
	}

	// Select word by weighted random
	r := rng.Float64() * totalWeight
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return words[i]
		}
	}

	// Fallback (should never reach here)
	return words[len(words)-1]
}

// deriveRuntimeSeed creates a seed from player input, conversation ID, and timestamp.
//
// This introduces controlled non-determinism:
//   - Same conversation with same input at different times = different responses
//   - Different conversations with same input = different responses
//
// The hash combines:
//  1. Base seed (world consistency)
//  2. Player input (context)
//  3. Conversation ID (thread uniqueness)
//  4. Timestamp (temporal variation)
func (m *MarkovGenerator) deriveRuntimeSeed(playerInput, conversationID string) int64 {
	h := sha256.New()

	// Write base seed
	binary.Write(h, binary.LittleEndian, m.seed)

	// Write player input
	h.Write([]byte(playerInput))

	// Write conversation ID
	h.Write([]byte(conversationID))

	// Write timestamp (source of non-determinism)
	timestamp := time.Now().UnixNano()
	binary.Write(h, binary.LittleEndian, timestamp)

	// Extract int64 from hash
	hash := h.Sum(nil)
	seed := int64(binary.LittleEndian.Uint64(hash[:8]))

	return seed
}

// makePrefix creates a space-separated prefix string from words.
func (m *MarkovGenerator) makePrefix(words []string) string {
	return strings.Join(words, " ")
}

// tokenize splits a sentence into words.
//
// Preserves punctuation attached to words (e.g., "Hello!" -> ["Hello!"])
// Normalizes whitespace and removes empty tokens.
func tokenize(text string) []string {
	// Split on whitespace
	rawTokens := strings.Fields(text)

	tokens := make([]string, 0, len(rawTokens))
	for _, token := range rawTokens {
		if token != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
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
