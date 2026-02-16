package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// TestNPCDialogComponentCreation verifies component creation.
func TestNPCDialogComponentCreation(t *testing.T) {
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)
	comp := &NPCDialogComponent{
		NPCPersonality:      personality,
		ConversationHistory: make([]string, 0),
		ResponseHistory:     make([]string, 0),
		DialogState:         "idle",
		TopicMemory:         make(map[string]bool),
		GenreID:             "fantasy",
	}

	if comp.NPCPersonality != personality {
		t.Error("NPCPersonality not set correctly")
	}
	if comp.DialogState != "idle" {
		t.Errorf("DialogState = %s, want 'idle'", comp.DialogState)
	}
	if len(comp.ConversationHistory) != 0 {
		t.Errorf("ConversationHistory length = %d, want 0", len(comp.ConversationHistory))
	}
}

// TestNPCDialogComponentType verifies Type() method.
func TestNPCDialogComponentType(t *testing.T) {
	comp := &NPCDialogComponent{}
	if comp.Type() != "npcdialog" {
		t.Errorf("Type() = %s, want 'npcdialog'", comp.Type())
	}
}

// TestAddPlayerInput verifies player input recording.
func TestAddPlayerInput(t *testing.T) {
	comp := &NPCDialogComponent{
		ConversationHistory: make([]string, 0),
	}

	// Add single input
	comp.AddPlayerInput("Hello")
	if len(comp.ConversationHistory) != 1 {
		t.Fatalf("history length = %d, want 1", len(comp.ConversationHistory))
	}
	if comp.ConversationHistory[0] != "Hello" {
		t.Errorf("history[0] = %q, want 'Hello'", comp.ConversationHistory[0])
	}

	// Add multiple inputs
	for i := 0; i < 5; i++ {
		comp.AddPlayerInput("Input")
	}
	if len(comp.ConversationHistory) != 6 {
		t.Errorf("history length = %d, want 6", len(comp.ConversationHistory))
	}
}

// TestAddPlayerInputLimit verifies 10-item history limit.
func TestAddPlayerInputLimit(t *testing.T) {
	comp := &NPCDialogComponent{
		ConversationHistory: make([]string, 0),
	}

	// Add 15 inputs (exceeds 10 limit)
	for i := 0; i < 15; i++ {
		comp.AddPlayerInput("Input")
	}

	// Should only keep last 10
	if len(comp.ConversationHistory) != 10 {
		t.Errorf("history length = %d, want 10", len(comp.ConversationHistory))
	}
}

// TestAddNPCResponse verifies NPC response recording.
func TestAddNPCResponse(t *testing.T) {
	comp := &NPCDialogComponent{
		ResponseHistory: make([]string, 0),
	}

	comp.AddNPCResponse("Greetings, traveler")
	if len(comp.ResponseHistory) != 1 {
		t.Fatalf("response history length = %d, want 1", len(comp.ResponseHistory))
	}
	if comp.ResponseHistory[0] != "Greetings, traveler" {
		t.Errorf("response[0] = %q, want 'Greetings, traveler'", comp.ResponseHistory[0])
	}
}

// TestGetRecentContext verifies context retrieval.
func TestGetRecentContext(t *testing.T) {
	comp := &NPCDialogComponent{
		ConversationHistory: []string{"Hello", "How are you?", "Goodbye"},
	}

	context := comp.GetRecentContext(2)
	expected := []string{"How are you?", "Goodbye"} // Last 2 items

	if len(context) != 2 {
		t.Fatalf("GetRecentContext(2) length = %d, want 2", len(context))
	}
	if context[0] != expected[0] || context[1] != expected[1] {
		t.Errorf("GetRecentContext(2) = %v, want %v", context, expected)
	}

	// Test with limit exceeding history
	context = comp.GetRecentContext(10)
	if len(context) != 3 {
		t.Errorf("GetRecentContext(10) length = %d, want 3", len(context))
	}

	// Test with zero limit
	context = comp.GetRecentContext(0)
	if len(context) != 0 {
		t.Errorf("GetRecentContext(0) length = %d, want 0", len(context))
	}
}

// TestGetRecentContextEmpty verifies empty history handling.
func TestGetRecentContextEmpty(t *testing.T) {
	comp := &NPCDialogComponent{
		ConversationHistory: []string{},
	}

	context := comp.GetRecentContext(5)
	if len(context) != 0 {
		t.Errorf("GetRecentContext with empty history length = %d, want 0", len(context))
	}
}

// TestSetDialogState verifies dialog state updates.
func TestSetDialogState(t *testing.T) {
	comp := &NPCDialogComponent{
		DialogState: "idle",
	}

	comp.SetDialogState("trading")
	if comp.DialogState != "trading" {
		t.Errorf("DialogState = %s, want 'trading'", comp.DialogState)
	}

	comp.SetDialogState("combat")
	if comp.DialogState != "combat" {
		t.Errorf("DialogState = %s, want 'combat'", comp.DialogState)
	}
}

// TestRememberTopic verifies topic memory.
func TestRememberTopic(t *testing.T) {
	comp := &NPCDialogComponent{
		TopicMemory: make(map[string]bool),
	}

	// Remember new topic
	comp.RememberTopic("dragons")
	if !comp.TopicMemory["dragons"] {
		t.Error("TopicMemory['dragons'] should be true")
	}

	// Check using HasDiscussedTopic
	if !comp.HasDiscussedTopic("dragons") {
		t.Error("HasDiscussedTopic('dragons') should be true")
	}

	// Remember different topic
	comp.RememberTopic("quest")
	if !comp.HasDiscussedTopic("quest") {
		t.Error("HasDiscussedTopic('quest') should be true")
	}

	// Verify both topics exist
	if len(comp.TopicMemory) != 2 {
		t.Errorf("TopicMemory size = %d, want 2", len(comp.TopicMemory))
	}

	// Verify undiscussed topic
	if comp.HasDiscussedTopic("treasure") {
		t.Error("HasDiscussedTopic('treasure') should be false")
	}
}

// TestResetConversation verifies conversation reset.
func TestResetConversation(t *testing.T) {
	comp := &NPCDialogComponent{
		ConversationHistory: []string{"Hello", "How are you?"},
		ResponseHistory:     []string{"Greetings", "I am well"},
		TopicMemory:         map[string]bool{"dragons": true, "quest": true},
		DialogState:         "trading",
	}

	comp.ResetConversation()

	// Verify conversation fields reset
	if len(comp.ConversationHistory) != 0 {
		t.Errorf("ConversationHistory length = %d, want 0", len(comp.ConversationHistory))
	}
	if len(comp.ResponseHistory) != 0 {
		t.Errorf("ResponseHistory length = %d, want 0", len(comp.ResponseHistory))
	}
	if comp.DialogState != "greeting" {
		t.Errorf("DialogState = %s, want 'greeting'", comp.DialogState)
	}

	// TopicMemory persists across conversations (NPC remembers)
	if len(comp.TopicMemory) != 2 {
		t.Errorf("TopicMemory size = %d, want 2 (topics persist)", len(comp.TopicMemory))
	}
}

// TestGetTimeSinceLastInteraction verifies interaction timing.
func TestGetTimeSinceLastInteraction(t *testing.T) {
	now := time.Now()
	comp := &NPCDialogComponent{
		LastInteractionTime: now.Add(-5 * time.Second),
	}

	elapsed := comp.GetTimeSinceLastInteraction()

	// Should be approximately 5 seconds
	if elapsed < 4*time.Second || elapsed > 6*time.Second {
		t.Errorf("GetTimeSinceLastInteraction = %v, want ~5s", elapsed)
	}
}

// TestGetTimeSinceLastInteractionZero verifies zero-time handling.
func TestGetTimeSinceLastInteractionZero(t *testing.T) {
	comp := &NPCDialogComponent{
		LastInteractionTime: time.Time{}, // Zero time
	}

	elapsed := comp.GetTimeSinceLastInteraction()

	// Should return 0 for zero time
	if elapsed != 0 {
		t.Errorf("GetTimeSinceLastInteraction with zero time = %v, want 0", elapsed)
	}
}

// TestNPCDialogComponentConcurrentAccess verifies thread-safety of basic operations.
func TestNPCDialogComponentConcurrentAccess(t *testing.T) {
	comp := &NPCDialogComponent{
		ConversationHistory: make([]string, 0),
		ResponseHistory:     make([]string, 0),
		TopicMemory:         make(map[string]bool),
	}

	// Note: NPCDialogComponent is not thread-safe by design.
	// This test just verifies sequential operations don't panic.
	// In production, the system layer handles synchronization.

	for i := 0; i < 10; i++ {
		comp.AddPlayerInput("test")
		comp.AddNPCResponse("response")
		comp.RememberTopic("topic")
		_ = comp.GetRecentContext(5)
	}

	if len(comp.ConversationHistory) != 10 {
		t.Errorf("ConversationHistory length = %d, want 10", len(comp.ConversationHistory))
	}
}

// TestNPCDialogComponentGeneratorField verifies generator field.
func TestNPCDialogComponentGeneratorField(t *testing.T) {
	comp := &NPCDialogComponent{
		Generator: nil, // Initially nil
	}

	if comp.Generator != nil {
		t.Error("Initial Generator should be nil")
	}

	// System would initialize generator
	gen := dialog.NewMarkovGenerator(12345, "fantasy", dialog.Order2)
	comp.Generator = gen

	if comp.Generator != gen {
		t.Error("Generator not set correctly")
	}
}

// TestNPCDialogComponent_TopicMemory tests topic tracking methods.
func TestNPCDialogComponent_TopicMemory(t *testing.T) {
	comp := NewNPCDialogComponent("fantasy", nil, 42)

	if comp.HasDiscussedTopic("weather") {
		t.Error("Should not have discussed topic initially")
	}

	comp.RememberTopic("weather")

	if !comp.HasDiscussedTopic("weather") {
		t.Error("Should have discussed topic after remembering")
	}
}

// TestNPCDialogComponent_ConversationState tests conversation state methods.
func TestNPCDialogComponent_ConversationState(t *testing.T) {
	comp := NewNPCDialogComponent("fantasy", nil, 42)

	if !comp.IsFirstInteraction() {
		t.Error("Should be first interaction initially")
	}
	if comp.GetConversationLength() != 0 {
		t.Errorf("Expected conversation length 0, got %d", comp.GetConversationLength())
	}

	comp.AddPlayerInput("hello")

	if comp.IsFirstInteraction() {
		t.Error("Should not be first interaction after input")
	}
	if comp.GetConversationLength() != 1 {
		t.Errorf("Expected conversation length 1, got %d", comp.GetConversationLength())
	}
}

// TestNPCDialogComponent_DeterministicMode tests deterministic mode toggle.
func TestNPCDialogComponent_DeterministicMode(t *testing.T) {
	comp := NewNPCDialogComponent("fantasy", nil, 42)

	if comp.IsDeterministicMode() {
		t.Error("Should not be deterministic initially")
	}

	comp.SetDeterministicMode(true)

	if !comp.IsDeterministicMode() {
		t.Error("Should be deterministic after enabling")
	}
}

// TestNPCDialogComponent_ClearHistory tests complete history clearing.
func TestNPCDialogComponent_ClearHistory(t *testing.T) {
	comp := NewNPCDialogComponent("fantasy", nil, 42)

	comp.AddPlayerInput("test input")
	comp.AddNPCResponse("test response")
	comp.RememberTopic("weather")

	comp.ClearHistory()

	if len(comp.ConversationHistory) != 0 {
		t.Error("ConversationHistory should be empty after clear")
	}
	if len(comp.ResponseHistory) != 0 {
		t.Error("ResponseHistory should be empty after clear")
	}
	if comp.HasDiscussedTopic("weather") {
		t.Error("TopicMemory should be cleared")
	}
}

// TestNPCDialogComponent_TimeSinceLastInteraction tests interaction timing.
func TestNPCDialogComponent_TimeSinceLastInteraction(t *testing.T) {
	comp := NewNPCDialogComponent("fantasy", nil, 42)

	// No interaction yet - should return zero duration
	duration := comp.GetTimeSinceLastInteraction()
	if duration != 0 {
		t.Errorf("Expected zero duration for no interactions, got %v", duration)
	}

	// After an input, should return non-zero duration
	comp.AddPlayerInput("hello")
	duration = comp.GetTimeSinceLastInteraction()
	if duration < 0 {
		t.Error("Duration should be non-negative after interaction")
	}
}
