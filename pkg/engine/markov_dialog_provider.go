package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// MarkovDialogProvider implements DialogProvider using Markov chain text generation.
// Provides varied, genre-appropriate dialog for NPCs based on trained corpora.
type MarkovDialogProvider struct {
	generator      *dialog.MarkovGenerator
	npcName        string
	genreID        string
	personality    *dialog.Personality
	conversationID string
}

// NewMarkovDialogProvider creates a dialog provider with Markov chain generation.
//
// Parameters:
//   - seed: Base seed for deterministic generation
//   - genreID: Genre for corpus selection ("fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic")
//   - npcName: Name of the NPC for greeting personalization
//   - personality: Optional personality traits (can be nil for default)
func NewMarkovDialogProvider(seed int64, genreID, npcName string, personality *dialog.Personality) *MarkovDialogProvider {
	// Create generator
	gen := dialog.NewMarkovGenerator(seed, genreID, dialog.Order2)

	// Get and train corpus
	corpus := dialog.GetCorpus(genreID)
	if corpus != nil {
		gen.TrainFromCorpus(corpus.Sentences)
	}

	// Generate conversation ID from NPC name and seed
	conversationID := fmt.Sprintf("npc-%s-%d", npcName, seed)

	return &MarkovDialogProvider{
		generator:      gen,
		npcName:        npcName,
		genreID:        genreID,
		personality:    personality,
		conversationID: conversationID,
	}
}

// GetDialog returns procedurally generated dialog text and options.
func (m *MarkovDialogProvider) GetDialog() (string, []DialogOption) {
	// Generate greeting using Markov chain
	params := dialog.GenerateParams{
		PlayerInput:    "hello",
		ConversationID: m.conversationID,
		MaxWords:       25,
		MinWords:       8,
		Temperature:    0.7,
	}

	text := m.generator.Generate(params)

	// Fallback to simple greeting if generation fails
	if text == "" {
		text = fmt.Sprintf("Greetings, traveler. I am %s.", m.npcName)
	}

	// Standard options for merchant-type NPCs
	options := []DialogOption{
		{
			Text:    "Browse your wares",
			Action:  ActionOpenShop,
			Enabled: true,
		},
		{
			Text:    "Tell me more",
			Action:  ActionNone,
			Enabled: true,
		},
		{
			Text:    "Farewell",
			Action:  ActionCloseDialog,
			Enabled: true,
		},
	}

	return text, options
}

// GenerateResponse creates a response to player input.
// This allows for dynamic conversation beyond the initial greeting.
func (m *MarkovDialogProvider) GenerateResponse(playerInput string) string {
	params := dialog.GenerateParams{
		PlayerInput:    playerInput,
		ConversationID: m.conversationID,
		MaxWords:       30,
		MinWords:       10,
		Temperature:    0.7,
	}

	response := m.generator.Generate(params)

	// Fallback
	if response == "" {
		response = "I understand."
	}

	return response
}
