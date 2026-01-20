package dialog

import (
	"math/rand"
	"sort"
)

// Utility functions for the dialog package.
// This file contains shared helper functions used across multiple components.
//
// Code relocated from: personality.go and markov.go

// clamp restricts a value to the range [min, max].
// Originally from: personality.go
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// max returns the larger of two integers.
// Originally from: personality.go
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers.
// Originally from: personality.go
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// selectMostFrequentWord chooses the most common word from candidates.
// Originally from: markov.go
func selectMostFrequentWord(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	// Count occurrences
	counts := make(map[string]int)
	for _, word := range candidates {
		counts[word]++
	}
	// Find most frequent
	var maxWord string
	maxCount := 0
	for word, count := range counts {
		if count > maxCount {
			maxCount = count
			maxWord = word
		}
	}
	return maxWord
}

// selectWeightedWord chooses a word using temperature-based weighting.
// Originally from: markov.go
func selectWeightedWord(candidates []string, rng *rand.Rand, temperature float64) string {
	if len(candidates) == 0 {
		return ""
	}
	if temperature == 0 {
		return selectMostFrequentWord(candidates)
	}

	// Build frequency map
	freq := buildFrequencyMap(candidates)

	// Sort words for deterministic iteration
	sortedWords := sortWords(freq)

	// Apply temperature weighting
	weights, totalWeight := calculateTemperatureWeights(freq, sortedWords, temperature)

	// Select word using weighted random selection
	if totalWeight == 0 {
		return sortedWords[0]
	}

	r := rng.Float64() * totalWeight
	cumulative := 0.0
	for i, word := range sortedWords {
		cumulative += weights[i]
		if r <= cumulative {
			return word
		}
	}

	// Fallback to last word
	return sortedWords[len(sortedWords)-1]
}

// buildFrequencyMap counts word occurrences.
// Originally from: markov.go
func buildFrequencyMap(candidates []string) map[string]float64 {
	freq := make(map[string]float64)
	for _, word := range candidates {
		freq[word]++
	}
	return freq
}

// sortWords returns sorted words from frequency map.
// Originally from: markov.go
func sortWords(freq map[string]float64) []string {
	words := make([]string, 0, len(freq))
	for word := range freq {
		words = append(words, word)
	}
	sort.Strings(words) // Sort for deterministic iteration
	return words
}

// calculateTemperatureWeights applies temperature scaling to word frequencies.
// Originally from: markov.go
func calculateTemperatureWeights(freq map[string]float64, sortedWords []string, temperature float64) ([]float64, float64) {
	weights := make([]float64, len(sortedWords))
	totalWeight := 0.0

	for i, word := range sortedWords {
		// Temperature affects how much we favor frequent words
		// Low temperature = more deterministic (favor frequent)
		// High temperature = more random (flatten distribution)
		weight := freq[word]
		if temperature > 0 && temperature != 1.0 {
			// Apply temperature scaling: weight^(1/temperature)
			// This transforms the distribution
			weight = 1.0 / temperature
			if freq[word] > 1 {
				// Only scale non-unique words
				factor := 1.0
				for j := 1.0; j < freq[word]; j++ {
					factor *= weight
				}
				weight = factor
			}
		}
		weights[i] = weight
		totalWeight += weight
	}

	return weights, totalWeight
}

// tokenize splits text into words.
// Originally from: markov.go
func tokenize(text string) []string {
	// Simple whitespace tokenization
	// Could be enhanced with punctuation handling
	return splitOnWhitespace(text)
}

func splitOnWhitespace(text string) []string {
	var words []string
	word := ""
	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
