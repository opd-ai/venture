package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// NPCDialogComponent tracks conversation state and history for NPCs using the Markov dialog system.
// This is separate from DialogComponent which handles simple choice-based dialogs.
type NPCDialogComponent struct {
	// NPCPersonality defines the NPC's character traits.
	NPCPersonality *dialog.Personality

	// CurrentConversationID uniquely identifies the active conversation.
	// Format: "npc-{entityID}-{timestamp}"
	CurrentConversationID string

	// ConversationHistory stores recent player inputs for context.
	// Limited to last 10 interactions to preserve memory.
	ConversationHistory []string

	// ResponseHistory stores NPC responses for reference.
	// Limited to last 10 responses.
	ResponseHistory []string

	// LastInteractionTime tracks when the NPC last spoke.
	LastInteractionTime time.Time

	// DialogState tracks conversation state for multi-turn dialogs.
	// Can be "greeting", "trading", "questing", "combat", etc.
	DialogState string

	// TopicMemory remembers topics discussed (for avoiding repetition).
	TopicMemory map[string]bool

	// Generator is the trained Markov generator for this NPC.
	// Cached to avoid re-training on every interaction.
	Generator *dialog.MarkovGenerator

	// GenreID determines the corpus and vocabulary.
	GenreID string

	// DeterministicMode forces template-based responses (for testing).
	DeterministicMode bool
}

// Type returns the component type identifier.
func (d *NPCDialogComponent) Type() string {
	return "npcdialog"
}

// NewNPCDialogComponent creates an NPC dialog component with default values.
func NewNPCDialogComponent(genreID string, personality *dialog.Personality, seed int64) *NPCDialogComponent {
	if personality == nil {
		personality = dialog.NewPersonality(dialog.PersonalityHelpful)
	}

	return &NPCDialogComponent{
		NPCPersonality:      personality,
		ConversationHistory: make([]string, 0, 10),
		ResponseHistory:     make([]string, 0, 10),
		DialogState:         "greeting",
		TopicMemory:         make(map[string]bool),
		GenreID:             genreID,
		Generator:           nil, // Initialized by NPCDialogSystem
		DeterministicMode:   false,
	}
}

// AddPlayerInput records a player message and updates conversation context.
func (d *NPCDialogComponent) AddPlayerInput(input string) {
	d.ConversationHistory = append(d.ConversationHistory, input)

	// Limit history to last 10 interactions
	if len(d.ConversationHistory) > 10 {
		d.ConversationHistory = d.ConversationHistory[len(d.ConversationHistory)-10:]
	}

	d.LastInteractionTime = time.Now()
}

// AddNPCResponse records an NPC response.
func (d *NPCDialogComponent) AddNPCResponse(response string) {
	d.ResponseHistory = append(d.ResponseHistory, response)

	// Limit history to last 10 responses
	if len(d.ResponseHistory) > 10 {
		d.ResponseHistory = d.ResponseHistory[len(d.ResponseHistory)-10:]
	}
}

// GetRecentContext returns the last N player inputs for context.
func (d *NPCDialogComponent) GetRecentContext(n int) []string {
	if n <= 0 || len(d.ConversationHistory) == 0 {
		return []string{}
	}

	start := len(d.ConversationHistory) - n
	if start < 0 {
		start = 0
	}

	return d.ConversationHistory[start:]
}

// SetDialogState updates the conversation state.
func (d *NPCDialogComponent) SetDialogState(state string) {
	d.DialogState = state
}

// GetDialogState returns the current conversation state.
func (d *NPCDialogComponent) GetDialogState() string {
	return d.DialogState
}

// RememberTopic marks a topic as discussed.
func (d *NPCDialogComponent) RememberTopic(topic string) {
	d.TopicMemory[topic] = true
}

// HasDiscussedTopic checks if a topic was previously discussed.
func (d *NPCDialogComponent) HasDiscussedTopic(topic string) bool {
	return d.TopicMemory[topic]
}

// ResetConversation clears conversation history (for new conversation thread).
func (d *NPCDialogComponent) ResetConversation() {
	d.ConversationHistory = make([]string, 0, 10)
	d.ResponseHistory = make([]string, 0, 10)
	d.DialogState = "greeting"
	d.CurrentConversationID = ""
	// TopicMemory persists across conversations (NPC remembers previous discussions)
}

// GetTimeSinceLastInteraction returns duration since last player interaction.
func (d *NPCDialogComponent) GetTimeSinceLastInteraction() time.Duration {
	if d.LastInteractionTime.IsZero() {
		return time.Duration(0)
	}
	return time.Since(d.LastInteractionTime)
}

// IsFirstInteraction returns true if this is the first conversation.
func (d *NPCDialogComponent) IsFirstInteraction() bool {
	return len(d.ConversationHistory) == 0
}

// GetConversationLength returns the number of exchanges in current conversation.
func (d *NPCDialogComponent) GetConversationLength() int {
	return len(d.ConversationHistory)
}

// SetDeterministicMode enables/disables template-based responses.
func (d *NPCDialogComponent) SetDeterministicMode(enabled bool) {
	d.DeterministicMode = enabled
}

// IsDeterministicMode returns true if using template-based dialog.
func (d *NPCDialogComponent) IsDeterministicMode() bool {
	return d.DeterministicMode
}

// ClearHistory removes all conversation and response history.
// Useful for memory management or when NPC should "forget" past interactions.
func (d *NPCDialogComponent) ClearHistory() {
	d.ConversationHistory = make([]string, 0, 10)
	d.ResponseHistory = make([]string, 0, 10)
	d.TopicMemory = make(map[string]bool)
}
