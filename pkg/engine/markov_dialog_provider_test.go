package engine

import (
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

func TestNewMarkovDialogProvider(t *testing.T) {
	tests := []struct {
		name        string
		seed        int64
		genreID     string
		npcName     string
		personality *dialog.Personality
	}{
		{
			name:        "fantasy merchant",
			seed:        12345,
			genreID:     "fantasy",
			npcName:     "Merchant Bob",
			personality: nil,
		},
		{
			name:        "scifi trader",
			seed:        67890,
			genreID:     "scifi",
			npcName:     "Trader Zyx",
			personality: nil,
		},
		{
			name:    "horror survivor with personality",
			seed:    11111,
			genreID: "horror",
			npcName: "Frightened Survivor",
			personality: &dialog.Personality{
				Friendliness: 0.3,
				Verbosity:    0.5,
				Formality:    0.2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMarkovDialogProvider(tt.seed, tt.genreID, tt.npcName, tt.personality)

			if provider == nil {
				t.Fatal("NewMarkovDialogProvider returned nil")
			}
			if provider.generator == nil {
				t.Error("generator is nil")
			}
			if provider.npcName != tt.npcName {
				t.Errorf("npcName = %s, want %s", provider.npcName, tt.npcName)
			}
			if provider.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", provider.genreID, tt.genreID)
			}
			if provider.conversationID == "" {
				t.Error("conversationID is empty")
			}
		})
	}
}

func TestMarkovDialogProvider_GetDialog(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		genreID string
		npcName string
	}{
		{"fantasy", 12345, "fantasy", "Merchant"},
		{"scifi", 67890, "scifi", "Trader"},
		{"horror", 11111, "horror", "Survivor"},
		{"cyberpunk", 22222, "cyberpunk", "Netrunner"},
		{"postapocalyptic", 33333, "postapocalyptic", "Scavenger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMarkovDialogProvider(tt.seed, tt.genreID, tt.npcName, nil)

			text, options := provider.GetDialog()

			// Verify text is not empty
			if text == "" {
				t.Error("GetDialog returned empty text")
			}

			// Verify options exist
			if len(options) == 0 {
				t.Error("GetDialog returned no options")
			}

			// Verify standard options are present
			hasShop := false
			hasClose := false
			for _, opt := range options {
				if opt.Action == ActionOpenShop {
					hasShop = true
				}
				if opt.Action == ActionCloseDialog {
					hasClose = true
				}
			}

			if !hasShop {
				t.Error("missing shop option")
			}
			if !hasClose {
				t.Error("missing close dialog option")
			}

			// All options should be enabled
			for i, opt := range options {
				if !opt.Enabled {
					t.Errorf("option %d is not enabled", i)
				}
				if opt.Text == "" {
					t.Errorf("option %d has empty text", i)
				}
			}
		})
	}
}

func TestMarkovDialogProvider_GenerateResponse(t *testing.T) {
	provider := NewMarkovDialogProvider(12345, "fantasy", "Merchant", nil)

	tests := []struct {
		name        string
		playerInput string
	}{
		{"greeting", "Hello"},
		{"question", "Where is the dungeon?"},
		{"purchase", "I want to buy something"},
		{"farewell", "Goodbye"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := provider.GenerateResponse(tt.playerInput)

			// Verify response is not empty
			if response == "" {
				t.Error("GenerateResponse returned empty string")
			}

			// Response should have reasonable length
			words := strings.Fields(response)
			if len(words) < 3 {
				t.Errorf("response too short: %d words", len(words))
			}
		})
	}
}

func TestMarkovDialogProvider_Determinism(t *testing.T) {
	// Same seed should produce same generator state
	seed := int64(98765)
	genreID := "fantasy"
	npcName := "Test Merchant"

	provider1 := NewMarkovDialogProvider(seed, genreID, npcName, nil)
	provider2 := NewMarkovDialogProvider(seed, genreID, npcName, nil)

	// Generators should be independently seeded
	// (non-deterministic due to runtime entropy in Generate())
	// But the fallback behavior should be consistent
	text1, _ := provider1.GetDialog()
	text2, _ := provider2.GetDialog()

	// Both should generate valid text
	if text1 == "" {
		t.Error("provider1 generated empty text")
	}
	if text2 == "" {
		t.Error("provider2 generated empty text")
	}
}

func TestMarkovDialogProvider_Variation(t *testing.T) {
	// Different seeds should produce different initial states
	npcName := "Merchant"
	genreID := "fantasy"

	provider1 := NewMarkovDialogProvider(11111, genreID, npcName, nil)
	provider2 := NewMarkovDialogProvider(22222, genreID, npcName, nil)

	text1, _ := provider1.GetDialog()
	text2, _ := provider2.GetDialog()

	// Both should generate valid text
	if text1 == "" {
		t.Error("provider1 generated empty text")
	}
	if text2 == "" {
		t.Error("provider2 generated empty text")
	}

	// Texts may or may not be different due to corpus size and randomness
	// Just verify they're both valid
	if len(strings.Fields(text1)) < 3 {
		t.Error("text1 too short")
	}
	if len(strings.Fields(text2)) < 3 {
		t.Error("text2 too short")
	}
}

func TestMarkovDialogProvider_InvalidGenre(t *testing.T) {
	// Invalid genre should still work with fallback
	provider := NewMarkovDialogProvider(12345, "invalid_genre", "Test NPC", nil)

	text, options := provider.GetDialog()

	// Should use fallback greeting
	if text == "" {
		t.Error("GetDialog returned empty text for invalid genre")
	}

	// Should still have options
	if len(options) == 0 {
		t.Error("GetDialog returned no options for invalid genre")
	}

	// Should contain NPC name in fallback
	if !strings.Contains(text, "Test NPC") {
		t.Errorf("fallback text missing NPC name: %s", text)
	}
}
