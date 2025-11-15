package network

import (
	"strings"
	"testing"
)

// TestNewProfanityFilter tests profanity filter creation
func TestNewProfanityFilter(t *testing.T) {
	pf := NewProfanityFilter()
	if pf == nil {
		t.Fatal("NewProfanityFilter returned nil")
	}

	// Should be disabled by default (opt-in)
	if pf.IsEnabled() {
		t.Error("Filter should be disabled by default")
	}

	// Should have default word list
	if len(pf.wordList) == 0 {
		t.Error("Default word list is empty")
	}

	// Should have default replacement
	if pf.replacement != "***" {
		t.Errorf("Expected replacement '***', got %q", pf.replacement)
	}
}

// TestEnableDisable tests enabling and disabling the filter
func TestEnableDisable(t *testing.T) {
	pf := NewProfanityFilter()

	// Initially disabled
	if pf.IsEnabled() {
		t.Error("Filter should start disabled")
	}

	// Enable
	pf.Enable()
	if !pf.IsEnabled() {
		t.Error("Filter should be enabled after Enable()")
	}

	// Disable
	pf.Disable()
	if pf.IsEnabled() {
		t.Error("Filter should be disabled after Disable()")
	}
}

// TestFilterDisabled tests that filtering is skipped when disabled
func TestFilterDisabled(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Disable() // Ensure disabled

	text := "This contains damn and hell"
	filtered := pf.Filter(text)

	if filtered != text {
		t.Error("Disabled filter should not modify text")
	}
}

// TestFilterEnabled tests basic profanity filtering
func TestFilterEnabled(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Enable()

	tests := []struct {
		name     string
		input    string
		contains string // What the output should NOT contain
	}{
		{"single word", "This is damn annoying", "damn"},
		{"multiple words", "Damn this crap", "damn"},
		{"case insensitive", "DAMN this CRAP", "DAMN"},
		{"mixed case", "DaMn this CrAp", "DaMn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := pf.Filter(tt.input)

			// Filtered text should not contain the profanity
			if strings.Contains(strings.ToLower(filtered), strings.ToLower(tt.contains)) {
				t.Errorf("Filtered text still contains %q: %q", tt.contains, filtered)
			}

			// Filtered text should contain replacement
			if !strings.Contains(filtered, "***") {
				t.Errorf("Filtered text missing replacement: %q", filtered)
			}
		})
	}
}

// TestFilterLeetSpeak tests filtering of leet speak substitutions
func TestFilterLeetSpeak(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Enable()

	tests := []struct {
		name  string
		input string
	}{
		{"@ substitution", "d@mn this"},
		{"number substitution", "h3ll yeah"},
		{"mixed substitution", "d4mn th1s cr@p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := pf.Filter(tt.input)

			// Should contain replacement (filter detected leet speak)
			if !strings.Contains(filtered, "***") {
				t.Errorf("Leet speak not filtered: input=%q, output=%q", tt.input, filtered)
			}
		})
	}
}

// TestContainsProfanity tests profanity detection without filtering
func TestContainsProfanity(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Enable()

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"contains profanity", "This is damn annoying", true},
		{"clean text", "This is very annoying", false},
		{"profanity in caps", "This is DAMN annoying", true},
		{"leet speak", "This is d@mn annoying", true},
		{"partial match", "condemnation is not profanity", false}, // Word boundary
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pf.ContainsProfanity(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for text: %q", tt.expected, result, tt.text)
			}
		})
	}
}

// TestContainsProfanityDisabled tests that detection respects enabled state
func TestContainsProfanityDisabled(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Disable()

	if pf.ContainsProfanity("This is damn annoying") {
		t.Error("Disabled filter should not detect profanity")
	}
}

// TestSetWordList tests custom word list setting
func TestSetWordList(t *testing.T) {
	pf := NewProfanityFilter()

	customWords := []string{"badword", "nasty", "rude"}
	pf.SetWordList(customWords)

	wordList := pf.GetWordList()
	if len(wordList) != len(customWords) {
		t.Errorf("Expected %d words, got %d", len(customWords), len(wordList))
	}

	// Verify words match
	for i, word := range customWords {
		if wordList[i] != word {
			t.Errorf("Word mismatch at index %d: expected %q, got %q", i, word, wordList[i])
		}
	}
}

// TestAddWord tests adding words to the filter
func TestAddWord(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Enable()

	initialCount := len(pf.GetWordList())
	pf.AddWord("newbadword")

	newCount := len(pf.GetWordList())
	if newCount != initialCount+1 {
		t.Errorf("Expected word list size %d, got %d", initialCount+1, newCount)
	}

	// Verify filtering works with new word
	filtered := pf.Filter("This is a newbadword test")
	if strings.Contains(strings.ToLower(filtered), "newbadword") {
		t.Error("Newly added word not filtered")
	}
}

// TestRemoveWord tests removing words from the filter
func TestRemoveWord(t *testing.T) {
	pf := NewProfanityFilter()
	pf.SetWordList([]string{"bad", "worse", "worst"})
	pf.Enable()

	pf.RemoveWord("worse")

	wordList := pf.GetWordList()
	if len(wordList) != 2 {
		t.Errorf("Expected 2 words after removal, got %d", len(wordList))
	}

	// Verify removed word no longer filtered
	filtered := pf.Filter("This is worse")
	if strings.Contains(filtered, "***") {
		t.Error("Removed word still being filtered")
	}

	// Verify remaining words still filtered
	filtered2 := pf.Filter("This is bad")
	if !strings.Contains(filtered2, "***") {
		t.Error("Remaining words not being filtered")
	}
}

// TestSetReplacement tests custom replacement string
func TestSetReplacement(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Enable()
	pf.SetReplacement("[censored]")

	filtered := pf.Filter("This is damn annoying")
	if !strings.Contains(filtered, "[censored]") {
		t.Errorf("Custom replacement not used: %q", filtered)
	}
	if strings.Contains(filtered, "***") {
		t.Error("Old replacement still present")
	}
}

// TestWordBoundaries tests that filter respects word boundaries
func TestWordBoundaries(t *testing.T) {
	pf := NewProfanityFilter()
	pf.SetWordList([]string{"hell"})
	pf.Enable()

	tests := []struct {
		name         string
		text         string
		shouldFilter bool
	}{
		{"standalone word", "Go to hell", true},
		{"part of word", "Hello world", false},
		{"part of word 2", "Shell script", false},
		{"at start", "Hell yeah", true},
		{"at end", "What the hell", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := pf.Filter(tt.text)
			containsReplacement := strings.Contains(filtered, "***")

			if tt.shouldFilter && !containsReplacement {
				t.Errorf("Expected filtering but got: %q", filtered)
			}
			if !tt.shouldFilter && containsReplacement {
				t.Errorf("Unexpected filtering: input=%q, output=%q", tt.text, filtered)
			}
		})
	}
}

// TestEmptyWordList tests behavior with empty word list
func TestEmptyWordList(t *testing.T) {
	pf := NewProfanityFilter()
	pf.SetWordList([]string{})
	pf.Enable()

	text := "This could be anything"
	filtered := pf.Filter(text)

	if filtered != text {
		t.Error("Filter with empty word list should not modify text")
	}

	if pf.ContainsProfanity(text) {
		t.Error("Empty word list should not detect profanity")
	}
}

// TestProfanityFilterConcurrentAccess tests thread-safe access to profanity filter
func TestProfanityFilterConcurrentAccess(t *testing.T) {
	pf := NewProfanityFilter()
	pf.Enable()

	done := make(chan bool)
	iterations := 100

	// Writer goroutine
	go func() {
		for i := 0; i < iterations; i++ {
			pf.AddWord("test")
			pf.RemoveWord("test")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < iterations; i++ {
			pf.Filter("test message")
			pf.ContainsProfanity("test message")
			pf.GetWordList()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

// BenchmarkFilter benchmarks text filtering
func BenchmarkFilter(b *testing.B) {
	pf := NewProfanityFilter()
	pf.Enable()
	text := "This is a test message with some damn and hell words in it"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.Filter(text)
	}
}

// BenchmarkFilterLongText benchmarks filtering of long text
func BenchmarkFilterLongText(b *testing.B) {
	pf := NewProfanityFilter()
	pf.Enable()

	// Build long text
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("This is a test sentence with some words. ")
	}
	text := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.Filter(text)
	}
}

// BenchmarkContainsProfanity benchmarks profanity detection
func BenchmarkContainsProfanity(b *testing.B) {
	pf := NewProfanityFilter()
	pf.Enable()
	text := "This is a clean message with no profanity"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.ContainsProfanity(text)
	}
}

// BenchmarkAddRemoveWord benchmarks word list modifications
func BenchmarkAddRemoveWord(b *testing.B) {
	pf := NewProfanityFilter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.AddWord("testword")
		pf.RemoveWord("testword")
	}
}
