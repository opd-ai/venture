package network

import (
	"regexp"
	"strings"
	"sync"
)

// ProfanityFilter provides client-side profanity filtering.
// This is opt-in and configurable by users.
type ProfanityFilter struct {
	mu          sync.RWMutex
	enabled     bool
	patterns    []*regexp.Regexp
	wordList    []string
	replacement string
}

// NewProfanityFilter creates a new profanity filter with default word list.
func NewProfanityFilter() *ProfanityFilter {
	pf := &ProfanityFilter{
		enabled:     false, // Opt-in by default
		wordList:    defaultProfanityWords(),
		replacement: "***",
	}
	pf.compilePatterns()
	return pf
}

// Enable enables the profanity filter.
func (pf *ProfanityFilter) Enable() {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.enabled = true
}

// Disable disables the profanity filter.
func (pf *ProfanityFilter) Disable() {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.enabled = false
}

// IsEnabled returns whether the filter is enabled.
func (pf *ProfanityFilter) IsEnabled() bool {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.enabled
}

// SetWordList updates the profanity word list and recompiles patterns.
func (pf *ProfanityFilter) SetWordList(words []string) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.wordList = words
	pf.compilePatterns()
}

// AddWord adds a word to the filter list.
func (pf *ProfanityFilter) AddWord(word string) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.wordList = append(pf.wordList, word)
	pf.compilePatterns()
}

// RemoveWord removes a word from the filter list.
func (pf *ProfanityFilter) RemoveWord(word string) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	
	filtered := make([]string, 0, len(pf.wordList))
	for _, w := range pf.wordList {
		if !strings.EqualFold(w, word) {
			filtered = append(filtered, w)
		}
	}
	pf.wordList = filtered
	pf.compilePatterns()
}

// GetWordList returns a copy of the current word list.
func (pf *ProfanityFilter) GetWordList() []string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	
	words := make([]string, len(pf.wordList))
	copy(words, pf.wordList)
	return words
}

// SetReplacement sets the replacement string for filtered words.
func (pf *ProfanityFilter) SetReplacement(replacement string) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.replacement = replacement
}

// Filter applies profanity filtering to text if enabled.
// Returns the filtered text.
func (pf *ProfanityFilter) Filter(text string) string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if !pf.enabled {
		return text
	}

	filtered := text
	for _, pattern := range pf.patterns {
		filtered = pattern.ReplaceAllString(filtered, pf.replacement)
	}
	return filtered
}

// ContainsProfanity checks if text contains profanity (without filtering).
func (pf *ProfanityFilter) ContainsProfanity(text string) bool {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if !pf.enabled {
		return false
	}

	for _, pattern := range pf.patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// compilePatterns compiles regex patterns from word list.
// Must be called with lock held.
func (pf *ProfanityFilter) compilePatterns() {
	pf.patterns = make([]*regexp.Regexp, 0, len(pf.wordList))
	
	for _, word := range pf.wordList {
		if word == "" {
			continue
		}
		
		// Create case-insensitive pattern with word boundaries
		// Pattern: \b<word>\b with common character substitutions
		pattern := buildPattern(word)
		regex, err := regexp.Compile("(?i)" + pattern)
		if err == nil {
			pf.patterns = append(pf.patterns, regex)
		}
	}
}

// buildPattern creates a regex pattern from a profanity word.
// Handles common character substitutions (leet speak).
func buildPattern(word string) string {
	// Character substitution map (common leet speak)
	substitutions := map[rune]string{
		'a': "[a@4]",
		'e': "[e3]",
		'i': "[i1!]",
		'o': "[o0]",
		's': "[s$5]",
		't': "[t7]",
		'l': "[l1]",
	}

	var pattern strings.Builder
	pattern.WriteString(`\b`)
	
	for _, char := range strings.ToLower(word) {
		if sub, ok := substitutions[char]; ok {
			pattern.WriteString(sub)
		} else {
			pattern.WriteString(regexp.QuoteMeta(string(char)))
		}
	}
	
	pattern.WriteString(`\b`)
	return pattern.String()
}

// defaultProfanityWords returns the default profanity word list.
// This is a minimal list for demonstration. Users can customize via SetWordList.
func defaultProfanityWords() []string {
	return []string{
		// Common swear words (minimal list - users should configure)
		"damn",
		"hell",
		"crap",
		"idiot",
		"stupid",
		// Add more as needed, but keep list minimal by default
		// Users can extend via configuration
	}
}

// LoadWordListFromFile loads a profanity word list from a file.
// Each line should contain one word. Lines starting with # are comments.
func (pf *ProfanityFilter) LoadWordListFromFile(filepath string) error {
	// Note: This would read from file, but we keep it simple for now
	// since the project uses zero external assets. Users can provide
	// word lists via API instead.
	return nil // Placeholder - implement if needed
}
