package engine

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// NPCDialogSystem manages NPC conversations using Markov chain dialog generation.
// This system initializes generators, processes player inputs, and generates responses.
// It integrates with ConversationManager for multi-party conversation support.
type NPCDialogSystem struct {
	world *World

	// corpusCache stores trained corpora per genre to avoid re-training
	corpusCache map[string]*dialog.Corpus

	// generatorCache stores trained generators per genre+seed combination
	generatorCache map[string]*dialog.MarkovGenerator

	// defaultOrder is the Markov chain order (2 or 3)
	defaultOrder dialog.MarkovOrder

	// worldSeed is the base seed for deterministic generation
	worldSeed int64

	// conversationManager handles multi-party conversations and turn-taking
	conversationManager *ConversationManager
}

// NewNPCDialogSystem creates an NPC dialog system.
func NewNPCDialogSystem(world *World, worldSeed int64) *NPCDialogSystem {
	return &NPCDialogSystem{
		world:               world,
		corpusCache:         make(map[string]*dialog.Corpus),
		generatorCache:      make(map[string]*dialog.MarkovGenerator),
		defaultOrder:        dialog.Order2, // Default to order 2 for good balance
		worldSeed:           worldSeed,
		conversationManager: NewConversationManager(world),
	}
}

// Update processes dialog components (minimal processing, dialog is event-driven).
func (s *NPCDialogSystem) Update(deltaTime float64) {
	// Dialog system is event-driven (triggered by player interactions)
	// No continuous update logic required
}

// InitializeNPCDialog initializes or retrieves a generator for an NPC.
//
// This method:
//  1. Checks if the NPC already has a generator (cached)
//  2. If not, creates and trains a new generator using genre-specific corpus
//  3. Caches the generator to avoid re-training
//
// Returns error if corpus cannot be loaded or training fails.
func (s *NPCDialogSystem) InitializeNPCDialog(entity *Entity, genreID string, personality *dialog.Personality, seed int64) error {
	// Get or create NPCDialogComponent
	dialogComp := s.getOrCreateDialogComponent(entity, genreID, personality, seed)

	// Check if generator already exists in component
	if dialogComp.Generator != nil {
		return nil // Already initialized
	}

	// Check if generator exists in cache
	cacheKey := fmt.Sprintf("%s-%d", genreID, seed)
	if cachedGen, exists := s.generatorCache[cacheKey]; exists {
		// Use cached generator
		dialogComp.Generator = cachedGen
		return nil
	}

	// Get corpus for genre
	corpus, err := s.getCorpus(genreID)
	if err != nil {
		return fmt.Errorf("failed to load corpus for genre %s: %w", genreID, err)
	}

	// Create and train generator
	gen := dialog.NewMarkovGenerator(seed, genreID, s.defaultOrder)
	gen.TrainFromCorpus(corpus.Sentences)

	// Cache generator
	s.generatorCache[cacheKey] = gen

	// Assign to component
	dialogComp.Generator = gen

	// Successfully initialized (logging removed for simplicity)

	return nil
}

// GenerateResponse creates an NPC response to player input.
//
// This method:
//  1. Validates the entity has an NPCDialogComponent
//  2. Records the player input for context
//  3. Generates a response using the Markov generator or templates
//  4. Records the NPC response
//  5. Returns the response text
//
// Returns error if component missing or generation fails.
func (s *NPCDialogSystem) GenerateResponse(entity *Entity, playerInput string) (string, error) {
	// Get dialog component
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return "", err
	}

	// Ensure generator is initialized
	if dialogComp.Generator == nil {
		return "", fmt.Errorf("dialog generator not initialized for entity %d", entity.ID)
	}

	// Record player input
	dialogComp.AddPlayerInput(playerInput)

	// Generate conversation ID if not set
	if dialogComp.CurrentConversationID == "" {
		dialogComp.CurrentConversationID = fmt.Sprintf("npc-%d-%d", entity.ID, time.Now().UnixNano())
	}

	// Prepare generation parameters
	params := dialog.GenerateParams{
		PlayerInput:    playerInput,
		ConversationID: dialogComp.CurrentConversationID,
		MaxWords:       30,
		MinWords:       10,
		Temperature:    0.7,
	}

	// Apply personality adjustments
	if dialogComp.NPCPersonality != nil {
		dialogComp.NPCPersonality.ApplyToGenerator(&params)
	}

	// Generate response (deterministic or non-deterministic)
	var response string
	if dialogComp.DeterministicMode {
		response = dialogComp.Generator.GenerateDeterministic(params)
	} else {
		response = dialogComp.Generator.Generate(params)
	}

	// Fallback to template if generation failed
	if response == "" {
		response = s.getTemplateResponse(dialogComp, playerInput)
	}

	// Record NPC response
	dialogComp.AddNPCResponse(response)

	// Successfully generated response (logging removed for simplicity)

	return response, nil
}

// GenerateGreeting creates an initial NPC greeting.
//
// This uses personality-based templates for consistent first impressions.
func (s *NPCDialogSystem) GenerateGreeting(entity *Entity) (string, error) {
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return "", err
	}

	if dialogComp.NPCPersonality == nil {
		return "Hello.", nil
	}

	greeting := dialogComp.NPCPersonality.GetGreeting(dialogComp.GenreID)

	// Record as first interaction
	dialogComp.AddNPCResponse(greeting)

	return greeting, nil
}

// ResetConversation clears conversation history for an NPC.
func (s *NPCDialogSystem) ResetConversation(entity *Entity) error {
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return err
	}

	dialogComp.ResetConversation()

	// Successfully reset (logging removed for simplicity)

	return nil
}

// SetDeterministicMode enables/disables deterministic dialog for an NPC.
func (s *NPCDialogSystem) SetDeterministicMode(entity *Entity, enabled bool) error {
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return err
	}

	dialogComp.SetDeterministicMode(enabled)

	// Successfully set mode (logging removed for simplicity)

	return nil
}

// SetDialogState updates the conversation state for an NPC.
func (s *NPCDialogSystem) SetDialogState(entity *Entity, state string) error {
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return err
	}

	dialogComp.SetDialogState(state)

	// Successfully set state (logging removed for simplicity)

	return nil
}

// getOrCreateDialogComponent retrieves or creates an NPCDialogComponent.
func (s *NPCDialogSystem) getOrCreateDialogComponent(entity *Entity, genreID string, personality *dialog.Personality, seed int64) *NPCDialogComponent {
	// Check if component exists
	if comp, exists := entity.GetComponent("npcdialog"); exists && comp != nil {
		if dialogComp, ok := comp.(*NPCDialogComponent); ok {
			return dialogComp
		}
	}

	// Create new component
	dialogComp := NewNPCDialogComponent(genreID, personality, seed)
	entity.AddComponent(dialogComp)

	return dialogComp
}

// getDialogComponent retrieves an NPCDialogComponent or returns error.
func (s *NPCDialogSystem) getDialogComponent(entity *Entity) (*NPCDialogComponent, error) {
	comp, exists := entity.GetComponent("npcdialog")
	if !exists || comp == nil {
		return nil, fmt.Errorf("entity %d has no NPCDialogComponent", entity.ID)
	}

	dialogComp, ok := comp.(*NPCDialogComponent)
	if !ok {
		return nil, fmt.Errorf("entity %d has invalid NPCDialogComponent type", entity.ID)
	}

	return dialogComp, nil
}

// getCorpus retrieves a corpus from cache or loads it.
func (s *NPCDialogSystem) getCorpus(genreID string) (*dialog.Corpus, error) {
	// Check cache
	if corpus, exists := s.corpusCache[genreID]; exists {
		return corpus, nil
	}

	// Load corpus
	corpus := dialog.GetCorpus(genreID)
	if corpus == nil {
		return nil, fmt.Errorf("unknown genre ID: %s", genreID)
	}

	// Cache corpus
	s.corpusCache[genreID] = corpus

	return corpus, nil
}

// getTemplateResponse returns a fallback template-based response.
//
// Used when Markov generation fails or produces empty output.
func (s *NPCDialogSystem) getTemplateResponse(dialogComp *NPCDialogComponent, playerInput string) string {
	// Use personality-based greeting as fallback
	if dialogComp.NPCPersonality != nil {
		return dialogComp.NPCPersonality.GetGreeting(dialogComp.GenreID)
	}

	// Ultimate fallback
	return "I have nothing to say."
}

// GetConversationHistory retrieves recent conversation history for an NPC.
func (s *NPCDialogSystem) GetConversationHistory(entity *Entity, n int) ([]string, error) {
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return nil, err
	}

	return dialogComp.GetRecentContext(n), nil
}

// GetResponseHistory retrieves recent NPC responses.
func (s *NPCDialogSystem) GetResponseHistory(entity *Entity, n int) ([]string, error) {
	dialogComp, err := s.getDialogComponent(entity)
	if err != nil {
		return nil, err
	}

	if n <= 0 || len(dialogComp.ResponseHistory) == 0 {
		return []string{}, nil
	}

	start := len(dialogComp.ResponseHistory) - n
	if start < 0 {
		start = 0
	}

	return dialogComp.ResponseHistory[start:], nil
}

// StartMultiPartyConversation initializes a conversation with multiple players and an NPC.
// Returns the conversation ID for tracking.
func (s *NPCDialogSystem) StartMultiPartyConversation(npcID uint64, playerIDs []uint64) (string, error) {
	conv, err := s.conversationManager.StartConversation(npcID, playerIDs)
	if err != nil {
		return "", fmt.Errorf("failed to start conversation: %w", err)
	}

	return conv.ID, nil
}

// QueuePlayerInput queues a player's dialog request to an NPC.
// Returns a channel that will receive the NPC's response asynchronously.
func (s *NPCDialogSystem) QueuePlayerInput(npcID, playerID uint64, input string) (<-chan *DialogResponse, error) {
	req, err := s.conversationManager.QueueDialogRequest(npcID, playerID, input)
	if err != nil {
		return nil, err
	}

	return req.ResponseChan, nil
}

// ProcessQueuedDialogs processes pending dialog requests for all NPCs.
// This should be called periodically (e.g., in Update loop).
func (s *NPCDialogSystem) ProcessQueuedDialogs(deltaTime float64) {
	// Get all NPCs with dialog queues
	s.conversationManager.mu.RLock()
	npcIDs := make([]uint64, 0, len(s.conversationManager.npcQueues))
	for npcID := range s.conversationManager.npcQueues {
		npcIDs = append(npcIDs, npcID)
	}
	s.conversationManager.mu.RUnlock()

	// Process each NPC's queue
	for _, npcID := range npcIDs {
		req, err := s.conversationManager.ProcessNextDialogRequest(npcID)
		if err != nil || req == nil {
			continue
		}

		// Get NPC entity
		npcEntity, ok := s.world.GetEntity(npcID)
		if !ok || npcEntity == nil {
			s.conversationManager.CompleteDialogRequest(npcID, req.RequestID, "", fmt.Errorf("NPC entity not found"))
			continue
		}

		// Generate response
		response, err := s.GenerateResponse(npcEntity, req.PlayerInput)

		// Complete request
		s.conversationManager.CompleteDialogRequest(npcID, req.RequestID, response, err)
	}
}

// GetDialogQueueStatus returns the status of an NPC's dialog queue.
func (s *NPCDialogSystem) GetDialogQueueStatus(npcID uint64) (queueSize int, hasActive bool, err error) {
	return s.conversationManager.GetDialogQueueStatus(npcID)
}

// GetConversationMessages retrieves all messages in a conversation.
func (s *NPCDialogSystem) GetConversationMessages(convID string) ([]ConversationMessage, error) {
	return s.conversationManager.GetConversationMessages(convID)
}

// AddConversationMessage adds a message to a conversation (for tracking multi-party chat).
func (s *NPCDialogSystem) AddConversationMessage(convID string, senderID uint64, senderName, content string) error {
	return s.conversationManager.AddMessage(convID, senderID, senderName, content)
}

// CleanupStaleConversations removes inactive conversations (called periodically).
func (s *NPCDialogSystem) CleanupStaleConversations() int {
	return s.conversationManager.CleanupStaleConversations()
}
