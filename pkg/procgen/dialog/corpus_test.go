package dialog

import (
	"testing"
)

// TestGetCorpus verifies corpus retrieval for all genres.
func TestGetCorpus(t *testing.T) {
	tests := []struct {
		genreID      string
		wantNil      bool
		minSentences int
	}{
		{"fantasy", false, 50},
		{"scifi", false, 50},
		{"horror", false, 50},
		{"cyberpunk", false, 50},
		{"postapoc", false, 50},
		{"unknown", true, 0},
		{"", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			corpus := GetCorpus(tt.genreID)

			if tt.wantNil {
				if corpus != nil {
					t.Errorf("GetCorpus(%q) = %v, want nil", tt.genreID, corpus)
				}
				return
			}

			if corpus == nil {
				t.Fatalf("GetCorpus(%q) returned nil", tt.genreID)
			}

			if corpus.GenreID != tt.genreID {
				t.Errorf("corpus.GenreID = %q, want %q", corpus.GenreID, tt.genreID)
			}

			if len(corpus.Sentences) < tt.minSentences {
				t.Errorf("corpus has %d sentences, want >= %d", len(corpus.Sentences), tt.minSentences)
			}
		})
	}
}

// TestGetFantasyCorpus verifies fantasy corpus content.
func TestGetFantasyCorpus(t *testing.T) {
	corpus := GetFantasyCorpus()

	if corpus == nil {
		t.Fatal("GetFantasyCorpus returned nil")
	}

	if corpus.GenreID != "fantasy" {
		t.Errorf("GenreID = %q, want fantasy", corpus.GenreID)
	}

	if len(corpus.Sentences) == 0 {
		t.Error("fantasy corpus has no sentences")
	}

	// Verify fantasy-specific vocabulary
	found := false
	for _, s := range corpus.Sentences {
		if containsAny(s, []string{"dungeon", "dragon", "magic", "sword"}) {
			found = true
			break
		}
	}

	if !found {
		t.Error("fantasy corpus lacks expected fantasy vocabulary")
	}
}

// TestGetSciFiCorpus verifies sci-fi corpus content.
func TestGetSciFiCorpus(t *testing.T) {
	corpus := GetSciFiCorpus()

	if corpus == nil {
		t.Fatal("GetSciFiCorpus returned nil")
	}

	if corpus.GenreID != "scifi" {
		t.Errorf("GenreID = %q, want scifi", corpus.GenreID)
	}

	if len(corpus.Sentences) == 0 {
		t.Error("scifi corpus has no sentences")
	}

	// Verify sci-fi vocabulary
	found := false
	for _, s := range corpus.Sentences {
		if containsAny(s, []string{"ship", "space", "cyber", "system"}) {
			found = true
			break
		}
	}

	if !found {
		t.Error("scifi corpus lacks expected sci-fi vocabulary")
	}
}

// TestGetHorrorCorpus verifies horror corpus content.
func TestGetHorrorCorpus(t *testing.T) {
	corpus := GetHorrorCorpus()

	if corpus == nil {
		t.Fatal("GetHorrorCorpus returned nil")
	}

	if corpus.GenreID != "horror" {
		t.Errorf("GenreID = %q, want horror", corpus.GenreID)
	}

	if len(corpus.Sentences) == 0 {
		t.Error("horror corpus has no sentences")
	}

	// Verify horror vocabulary
	found := false
	for _, s := range corpus.Sentences {
		if containsAny(s, []string{"dark", "death", "fear", "monster"}) {
			found = true
			break
		}
	}

	if !found {
		t.Error("horror corpus lacks expected horror vocabulary")
	}
}

// TestGetCyberpunkCorpus verifies cyberpunk corpus content.
func TestGetCyberpunkCorpus(t *testing.T) {
	corpus := GetCyberpunkCorpus()

	if corpus == nil {
		t.Fatal("GetCyberpunkCorpus returned nil")
	}

	if corpus.GenreID != "cyberpunk" {
		t.Errorf("GenreID = %q, want cyberpunk", corpus.GenreID)
	}

	if len(corpus.Sentences) == 0 {
		t.Error("cyberpunk corpus has no sentences")
	}

	// Verify cyberpunk vocabulary
	found := false
	for _, s := range corpus.Sentences {
		if containsAny(s, []string{"chrome", "corp", "hack", "cyber"}) {
			found = true
			break
		}
	}

	if !found {
		t.Error("cyberpunk corpus lacks expected cyberpunk vocabulary")
	}
}

// TestGetPostApocalypticCorpus verifies post-apocalyptic corpus content.
func TestGetPostApocalypticCorpus(t *testing.T) {
	corpus := GetPostApocalypticCorpus()

	if corpus == nil {
		t.Fatal("GetPostApocalypticCorpus returned nil")
	}

	if corpus.GenreID != "postapoc" {
		t.Errorf("GenreID = %q, want postapocalyptic", corpus.GenreID)
	}

	if len(corpus.Sentences) == 0 {
		t.Error("postapocalyptic corpus has no sentences")
	}

	// Verify post-apocalyptic vocabulary
	found := false
	for _, s := range corpus.Sentences {
		if containsAny(s, []string{"wasteland", "radi", "scav", "surviv"}) {
			found = true
			break
		}
	}

	if !found {
		t.Error("postapocalyptic corpus lacks expected post-apocalyptic vocabulary")
	}
}

// TestGetAllCorpora verifies retrieval of all corpora.
func TestGetAllCorpora(t *testing.T) {
	all := GetAllCorpora()

	if len(all) != 5 {
		t.Errorf("GetAllCorpora returned %d corpora, want 5", len(all))
	}

	// Verify all genres present
	genres := make(map[string]bool)
	for _, corpus := range all {
		if corpus != nil {
			genres[corpus.GenreID] = true
		}
	}

	expected := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range expected {
		if !genres[genre] {
			t.Errorf("GetAllCorpora missing genre %q", genre)
		}
	}
}

// TestGetAvailableGenres verifies genre list retrieval.
func TestGetAvailableGenres(t *testing.T) {
	genres := GetAvailableGenres()

	if len(genres) != 5 {
		t.Errorf("GetAvailableGenres returned %d genres, want 5", len(genres))
	}

	// Verify expected genres - these should match GetCorpus() switch cases
	expected := map[string]bool{
		"fantasy":         true,
		"scifi":           true,
		"horror":          true,
		"cyberpunk":       true,
		"postapoc": true,
	}

	for _, genre := range genres {
		if !expected[genre] {
			t.Errorf("unexpected genre %q in list", genre)
		}
	}
}

// containsAny checks if string contains any of the given substrings.
func containsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if len(s) >= len(sub) {
			// Simple substring check
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
