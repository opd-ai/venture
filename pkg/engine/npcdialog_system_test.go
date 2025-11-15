package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// TestNewNPCDialogSystem verifies system creation.
func TestNewNPCDialogSystem(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.worldSeed != 12345 {
		t.Errorf("worldSeed = %d, want 12345", system.worldSeed)
	}
	if system.defaultOrder != dialog.Order2 {
		t.Errorf("defaultOrder = %v, want Order2", system.defaultOrder)
	}
	if system.corpusCache == nil {
		t.Error("corpusCache not initialized")
	}
	if system.generatorCache == nil {
		t.Error("generatorCache not initialized")
	}
}

// TestInitializeNPCDialog verifies NPC dialog initialization.
func TestInitializeNPCDialog(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	// Verify component was created
	comp, exists := entity.GetComponent("npcdialog")
	if !exists {
		t.Fatal("NPCDialogComponent not created")
	}

	dialogComp := comp.(*NPCDialogComponent)
	if dialogComp.Generator == nil {
		t.Error("Generator not initialized")
	}
	if dialogComp.NPCPersonality != personality {
		t.Error("Personality not set correctly")
	}
	if dialogComp.GenreID != "fantasy" {
		t.Errorf("GenreID = %s, want 'fantasy'", dialogComp.GenreID)
	}
}

// TestInitializeNPCDialogInvalidGenre verifies error handling for invalid genre.
func TestInitializeNPCDialogInvalidGenre(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "invalid-genre", personality, 54321)
	if err == nil {
		t.Error("expected error for invalid genre, got nil")
	}
	if !strings.Contains(err.Error(), "unknown genre") {
		t.Errorf("error message = %q, want 'unknown genre'", err.Error())
	}
}

// TestInitializeNPCDialogCaching verifies generator caching.
func TestInitializeNPCDialogCaching(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()

	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	// Initialize first entity
	err := system.InitializeNPCDialog(entity1, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog(entity1) failed: %v", err)
	}

	// Initialize second entity with same genre/seed
	err = system.InitializeNPCDialog(entity2, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog(entity2) failed: %v", err)
	}

	// Verify both use same cached generator
	comp1, _ := entity1.GetComponent("npcdialog")
	comp2, _ := entity2.GetComponent("npcdialog")

	gen1 := comp1.(*NPCDialogComponent).Generator
	gen2 := comp2.(*NPCDialogComponent).Generator

	// Generators should be the same instance (cached)
	if gen1 != gen2 {
		t.Error("generators not cached correctly")
	}
}

// TestGenerateResponse verifies response generation.
func TestGenerateResponse(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	response, err := system.GenerateResponse(entity, "Where is the dungeon?")
	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if response == "" {
		t.Error("GenerateResponse returned empty string")
	}

	// Verify response was recorded
	comp, _ := entity.GetComponent("npcdialog")
	dialogComp := comp.(*NPCDialogComponent)

	if len(dialogComp.ResponseHistory) != 1 {
		t.Errorf("ResponseHistory length = %d, want 1", len(dialogComp.ResponseHistory))
	}
	if len(dialogComp.ConversationHistory) != 1 {
		t.Errorf("ConversationHistory length = %d, want 1", len(dialogComp.ConversationHistory))
	}
}

// TestGenerateResponseVariation verifies response variation (Acceptance Criteria #1).
// Generate 5+ unique responses per NPC for same input.
func TestGenerateResponseVariation(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	playerInput := "Where is the treasure?"
	responses := make(map[string]bool)
	iterations := 10

	for i := 0; i < iterations; i++ {
		// Reset conversation to avoid context influence
		system.ResetConversation(entity)

		response, err := system.GenerateResponse(entity, playerInput)
		if err != nil {
			t.Fatalf("iteration %d: GenerateResponse failed: %v", i, err)
		}

		if response == "" {
			t.Errorf("iteration %d: empty response", i)
			continue
		}

		responses[response] = true
	}

	uniqueCount := len(responses)
	minUnique := 5 // Acceptance criteria: 5+ unique responses

	if uniqueCount < minUnique {
		t.Errorf("insufficient variation: %d unique responses out of %d (want >= %d)",
			uniqueCount, iterations, minUnique)
	}
}

// TestGenerateResponseDeterministic verifies deterministic mode (Acceptance Criteria #2).
// Deterministic mode produces identical dialog given same seed.
func TestGenerateResponseDeterministic(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	// Enable deterministic mode
	system.SetDeterministicMode(entity, true)

	playerInput := "Tell me about the dungeon."
	var firstResponse string

	// Generate responses 10 times
	for i := 0; i < 10; i++ {
		// Reset conversation to ensure consistent state
		system.ResetConversation(entity)

		response, err := system.GenerateResponse(entity, playerInput)
		if err != nil {
			t.Fatalf("iteration %d: GenerateResponse failed: %v", i, err)
		}

		if i == 0 {
			firstResponse = response
		} else if response != firstResponse {
			t.Errorf("deterministic mode failed: iteration %d response differs\n  first: %s\n  current: %s",
				i, firstResponse, response)
		}
	}
}

// TestGenerateResponsePerformance verifies response generation <50ms (Acceptance Criteria #4).
func TestGenerateResponsePerformance(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	playerInput := "What can you tell me?"
	maxDuration := 50 * time.Millisecond

	start := time.Now()
	_, err = system.GenerateResponse(entity, playerInput)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if duration > maxDuration {
		t.Errorf("generation too slow: %v (want < %v)", duration, maxDuration)
	}
}

// TestGenerateResponseWithoutInitialization verifies error handling.
func TestGenerateResponseWithoutInitialization(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()

	_, err := system.GenerateResponse(entity, "Hello")
	if err == nil {
		t.Error("expected error for uninitialized entity, got nil")
	}
}

// TestGenerateResponseFallback verifies template fallback (Acceptance Criteria #5).
func TestGenerateResponseFallback(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	// Initialize with empty corpus (simulates generation failure)
	dialogComp := NewNPCDialogComponent("fantasy", personality, 54321)
	entity.AddComponent(dialogComp)

	// Create generator but don't train it
	gen := dialog.NewMarkovGenerator(54321, "fantasy", dialog.Order2)
	dialogComp.Generator = gen

	// Should fall back to template
	response, err := system.GenerateResponse(entity, "Hello")
	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if response == "" {
		t.Error("fallback returned empty string")
	}
}

// TestGenerateGreeting verifies greeting generation.
func TestGenerateGreeting(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	greeting, err := system.GenerateGreeting(entity)
	if err != nil {
		t.Fatalf("GenerateGreeting failed: %v", err)
	}

	if greeting == "" {
		t.Error("GenerateGreeting returned empty string")
	}

	// Verify greeting was recorded
	comp, _ := entity.GetComponent("npcdialog")
	dialogComp := comp.(*NPCDialogComponent)

	if len(dialogComp.ResponseHistory) != 1 {
		t.Errorf("ResponseHistory length = %d, want 1", len(dialogComp.ResponseHistory))
	}
}

// TestNPCDialogSystemResetConversation verifies conversation reset via system.
func TestNPCDialogSystemResetConversation(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	// Generate some dialog
	system.GenerateResponse(entity, "Hello")
	system.GenerateResponse(entity, "Goodbye")

	// Reset conversation
	err = system.ResetConversation(entity)
	if err != nil {
		t.Fatalf("ResetConversation failed: %v", err)
	}

	// Verify history cleared
	comp, _ := entity.GetComponent("npcdialog")
	dialogComp := comp.(*NPCDialogComponent)

	if len(dialogComp.ConversationHistory) != 0 {
		t.Errorf("ConversationHistory not cleared: %d items", len(dialogComp.ConversationHistory))
	}
	if len(dialogComp.ResponseHistory) != 0 {
		t.Errorf("ResponseHistory not cleared: %d items", len(dialogComp.ResponseHistory))
	}
}

// TestNPCDialogSystemSetDialogState verifies dialog state management via system.
func TestNPCDialogSystemSetDialogState(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	err = system.SetDialogState(entity, "trading")
	if err != nil {
		t.Fatalf("SetDialogState failed: %v", err)
	}

	comp, _ := entity.GetComponent("npcdialog")
	dialogComp := comp.(*NPCDialogComponent)

	if dialogComp.DialogState != "trading" {
		t.Errorf("DialogState = %s, want 'trading'", dialogComp.DialogState)
	}
}

// TestGetConversationHistory verifies history retrieval.
func TestGetConversationHistory(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	// Generate some dialog
	system.GenerateResponse(entity, "Hello")
	system.GenerateResponse(entity, "How are you?")
	system.GenerateResponse(entity, "Goodbye")

	history, err := system.GetConversationHistory(entity, 2)
	if err != nil {
		t.Fatalf("GetConversationHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("history length = %d, want 2", len(history))
	}
	if history[0] != "How are you?" {
		t.Errorf("history[0] = %q, want 'How are you?'", history[0])
	}
	if history[1] != "Goodbye" {
		t.Errorf("history[1] = %q, want 'Goodbye'", history[1])
	}
}

// TestGetResponseHistory verifies response history retrieval.
func TestGetResponseHistory(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	err := system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	if err != nil {
		t.Fatalf("InitializeNPCDialog failed: %v", err)
	}

	// Generate some dialog
	system.GenerateResponse(entity, "Hello")
	system.GenerateResponse(entity, "Goodbye")

	history, err := system.GetResponseHistory(entity, 1)
	if err != nil {
		t.Fatalf("GetResponseHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("history length = %d, want 1", len(history))
	}
}

// TestMultipleGenres verifies system works with multiple genres.
func TestMultipleGenres(t *testing.T) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	for _, genre := range genres {
		entity := world.CreateEntity()
		personality := dialog.NewPersonality(dialog.PersonalityHelpful)

		err := system.InitializeNPCDialog(entity, genre, personality, 54321)
		if err != nil {
			t.Errorf("genre %s: InitializeNPCDialog failed: %v", genre, err)
			continue
		}

		response, err := system.GenerateResponse(entity, "Hello")
		if err != nil {
			t.Errorf("genre %s: GenerateResponse failed: %v", genre, err)
			continue
		}

		if response == "" {
			t.Errorf("genre %s: empty response", genre)
		}
	}
}

// TestGenreAppropriateVocabulary verifies genre-specific vocabulary (Acceptance Criteria #6).
func TestGenreAppropriateVocabulary(t *testing.T) {
	tests := []struct {
		genre    string
		keywords []string // Words that should appear in corpus
	}{
		{"fantasy", []string{"dungeon", "dragon", "magic", "sword"}},
		{"scifi", []string{"system", "protocol", "data", "technology"}},
		{"horror", []string{"dark", "fear", "shadow", "terror"}},
		{"cyberpunk", []string{"corporate", "network", "cyber", "neon"}},
		{"postapocalyptic", []string{"wasteland", "survivor", "radiation", "ruins"}},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			corpus := dialog.GetCorpus(tt.genre)
			if corpus == nil {
				t.Fatalf("corpus for %s not found", tt.genre)
			}

			// Check that corpus contains genre-appropriate keywords
			corpusText := strings.Join(corpus.Sentences, " ")
			corpusLower := strings.ToLower(corpusText)

			foundCount := 0
			for _, keyword := range tt.keywords {
				if strings.Contains(corpusLower, strings.ToLower(keyword)) {
					foundCount++
				}
			}

			// At least 50% of keywords should be present
			minFound := len(tt.keywords) / 2
			if foundCount < minFound {
				t.Errorf("genre %s: only %d/%d keywords found (want >= %d)",
					tt.genre, foundCount, len(tt.keywords), minFound)
			}
		})
	}
}

// BenchmarkGenerateResponse measures response generation performance.
func BenchmarkGenerateResponse(b *testing.B) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	system.InitializeNPCDialog(entity, "fantasy", personality, 54321)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.GenerateResponse(entity, "Where is the dungeon?")
	}
}

// BenchmarkInitializeNPCDialog measures initialization performance.
func BenchmarkInitializeNPCDialog(b *testing.B) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := world.CreateEntity()
		system.InitializeNPCDialog(entity, "fantasy", personality, int64(i))
	}
}

// BenchmarkGenerateResponseDeterministic measures deterministic generation performance.
func BenchmarkGenerateResponseDeterministic(b *testing.B) {
	world := NewWorld()
	system := NewNPCDialogSystem(world, 12345)

	entity := world.CreateEntity()
	personality := dialog.NewPersonality(dialog.PersonalityHelpful)

	system.InitializeNPCDialog(entity, "fantasy", personality, 54321)
	system.SetDeterministicMode(entity, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.ResetConversation(entity)
		system.GenerateResponse(entity, "Tell me about the quest.")
	}
}
