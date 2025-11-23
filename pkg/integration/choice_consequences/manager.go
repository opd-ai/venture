package choice_consequences

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ChoiceTracker manages player choice tracking and consequence application.
type ChoiceTracker struct {
	mu             sync.RWMutex
	players        map[string]*PlayerState // Player ID -> state
	questBranches  map[string]*QuestBranch // Quest ID -> branch info
	classQuests    []*ClassSpecificQuest   // Class-specific quests
	npcMemoryLimit int                     // Max events per NPC (default 50)
	choiceLimit    int                     // Max choices tracked per player (default 200)
}

// PlayerState tracks all choice-related state for a player.
type PlayerState struct {
	PlayerID           string                      // Player entity ID
	ChoiceHistory      []*PlayerChoice             // All choices made
	NPCRelationships   map[string]*NPCRelationship // NPC ID -> relationship
	ContentLocks       map[string]*ContentLock     // Content ID -> lock
	Alignment          *PlayerAlignment            // Current alignment
	CompanionReactions []*CompanionReaction        // Recent companion reactions
	LastUpdate         int64                       // Last state update timestamp
}

// NewChoiceTracker creates a new choice tracker.
func NewChoiceTracker() *ChoiceTracker {
	return &ChoiceTracker{
		players:        make(map[string]*PlayerState),
		questBranches:  make(map[string]*QuestBranch),
		classQuests:    make([]*ClassSpecificQuest, 0),
		npcMemoryLimit: 50,
		choiceLimit:    200,
	}
}

// RecordChoice records a player choice and applies consequences.
func (ct *ChoiceTracker) RecordChoice(playerID string, choice *PlayerChoice) error {
	if err := validateChoice(choice); err != nil {
		return err
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	state := ct.getOrCreateState(playerID)
	addChoiceToHistory(state, choice, ct.choiceLimit)
	applyChoiceEffects(ct, state, choice)
	state.LastUpdate = time.Now().Unix()

	return nil
}

// validateChoice validates the choice parameters.
func validateChoice(choice *PlayerChoice) error {
	if choice == nil {
		return fmt.Errorf("choice cannot be nil")
	}
	if choice.ChoiceID == "" {
		return fmt.Errorf("choice ID cannot be empty")
	}
	return nil
}

// addChoiceToHistory adds a choice to history with LRU eviction.
func addChoiceToHistory(state *PlayerState, choice *PlayerChoice, choiceLimit int) {
	state.ChoiceHistory = append(state.ChoiceHistory, choice)

	if len(state.ChoiceHistory) > choiceLimit {
		newHistory := make([]*PlayerChoice, 0, choiceLimit)

		for _, c := range state.ChoiceHistory {
			if c.Irreversible {
				newHistory = append(newHistory, c)
			}
		}

		start := len(state.ChoiceHistory) - (choiceLimit - len(newHistory))
		if start < 0 {
			start = 0
		}
		for i := start; i < len(state.ChoiceHistory); i++ {
			if !state.ChoiceHistory[i].Irreversible {
				newHistory = append(newHistory, state.ChoiceHistory[i])
			}
		}

		state.ChoiceHistory = newHistory
	}
}

// applyChoiceEffects applies all effects of a choice to player state.
func applyChoiceEffects(ct *ChoiceTracker, state *PlayerState, choice *PlayerChoice) {
	if choice.MoralAlignment != nil {
		state.Alignment.ApplyShift(choice.MoralAlignment)
	}

	for _, npcID := range choice.NPCsAffected {
		ct.updateNPCRelationship(state, npcID, choice)
	}

	if choice.Irreversible && len(choice.Consequences) > 0 {
		for _, consequence := range choice.Consequences {
			ct.applyConsequence(state, choice.ChoiceID, consequence)
		}
	}
}

// getOrCreateState gets or creates player state.
func (ct *ChoiceTracker) getOrCreateState(playerID string) *PlayerState {
	if state, exists := ct.players[playerID]; exists {
		return state
	}

	state := &PlayerState{
		PlayerID:         playerID,
		ChoiceHistory:    make([]*PlayerChoice, 0, 100),
		NPCRelationships: make(map[string]*NPCRelationship),
		ContentLocks:     make(map[string]*ContentLock),
		Alignment: &PlayerAlignment{
			GoodEvil:      0.0,
			LawChaos:      0.0,
			HonorDishonor: 0.0,
			UpdatedAt:     time.Now().Unix(),
		},
		CompanionReactions: make([]*CompanionReaction, 0, 20),
		LastUpdate:         time.Now().Unix(),
	}

	ct.players[playerID] = state
	return state
}

// updateNPCRelationship updates NPC attitude based on choice.
func (ct *ChoiceTracker) updateNPCRelationship(state *PlayerState, npcID string, choice *PlayerChoice) {
	rel, exists := state.NPCRelationships[npcID]
	if !exists {
		rel = &NPCRelationship{
			NPCID:           npcID,
			Attitude:        0.0,
			TrustLevel:      0.0,
			LastUpdate:      time.Now().Unix(),
			MemorableEvents: make([]MemorableEvent, 0, ct.npcMemoryLimit),
			DialogueUnlocks: make([]string, 0),
			QuestsAffected:  make([]string, 0),
		}
		state.NPCRelationships[npcID] = rel
	}

	// Calculate impact based on alignment shift
	impact := 0.0
	if choice.MoralAlignment != nil {
		// Good actions tend to improve NPC relations
		impact = choice.MoralAlignment.GoodEvil * 0.3
		impact += choice.MoralAlignment.HonorDishonor * 0.2
	}

	// Add memorable event
	event := MemorableEvent{
		EventID:     fmt.Sprintf("%s_%s", choice.ChoiceID, npcID),
		ChoiceID:    choice.ChoiceID,
		Timestamp:   choice.Timestamp,
		Impact:      impact,
		Description: fmt.Sprintf("Player made choice: %s", choice.ChoiceID),
	}

	rel.MemorableEvents = append(rel.MemorableEvents, event)

	// Enforce memory limit (keep most impactful events)
	if len(rel.MemorableEvents) > ct.npcMemoryLimit {
		// Sort by absolute impact (most significant events)
		for i := 0; i < len(rel.MemorableEvents)-1; i++ {
			for j := i + 1; j < len(rel.MemorableEvents); j++ {
				if abs(rel.MemorableEvents[i].Impact) < abs(rel.MemorableEvents[j].Impact) {
					rel.MemorableEvents[i], rel.MemorableEvents[j] = rel.MemorableEvents[j], rel.MemorableEvents[i]
				}
			}
		}
		rel.MemorableEvents = rel.MemorableEvents[:ct.npcMemoryLimit]
	}

	// Update attitude and trust
	rel.Attitude += impact
	rel.Attitude = clamp(rel.Attitude, -1.0, 1.0)

	if rel.Attitude > 0.0 {
		rel.TrustLevel += 0.05 // Positive actions build trust slowly
	} else if rel.Attitude < 0.0 {
		rel.TrustLevel -= 0.1 // Negative actions destroy trust quickly
	}
	rel.TrustLevel = clamp(rel.TrustLevel, 0.0, 1.0)

	rel.LastUpdate = time.Now().Unix()
}

// applyConsequence applies a consequence to lock content.
func (ct *ChoiceTracker) applyConsequence(state *PlayerState, choiceID, consequence string) {
	// Parse consequence ID to determine lock type and content
	// Format: "lock_quest_<quest_id>" or "lock_npc_<npc_id>" etc.
	lockType := LockTypeQuest
	contentID := consequence

	if len(consequence) > 11 && consequence[:11] == "lock_quest_" {
		lockType = LockTypeQuest
		contentID = consequence[11:]
	} else if len(consequence) > 9 && consequence[:9] == "lock_npc_" {
		lockType = LockTypeNPC
		contentID = consequence[9:]
	} else if len(consequence) > 10 && consequence[:10] == "lock_area_" {
		lockType = LockTypeArea
		contentID = consequence[10:]
	}

	lock := &ContentLock{
		ContentID:        contentID,
		LockedBy:         choiceID,
		LockType:         lockType,
		Timestamp:        time.Now().Unix(),
		Permanent:        true, // Most consequences are permanent
		UnlockConditions: make([]string, 0),
	}

	state.ContentLocks[contentID] = lock
}

// IsContentAvailable checks if content is available to a player.
func (ct *ChoiceTracker) IsContentAvailable(playerID, contentID string) bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	state, exists := ct.players[playerID]
	if !exists {
		return true // No choices made yet, all content available
	}

	if lock, locked := state.ContentLocks[contentID]; locked {
		if lock.Permanent {
			return false
		}

		// Check if unlock conditions are met
		if len(lock.UnlockConditions) == 0 {
			return false
		}

		for _, condition := range lock.UnlockConditions {
			if !ct.hasChoice(state, condition) {
				return false
			}
		}

		// All unlock conditions met, remove lock
		delete(state.ContentLocks, contentID)
		return true
	}

	return true
}

// hasChoice checks if player has made a specific choice.
func (ct *ChoiceTracker) hasChoice(state *PlayerState, choiceID string) bool {
	for _, choice := range state.ChoiceHistory {
		if choice.ChoiceID == choiceID {
			return true
		}
	}
	return false
}

// GetNPCAttitude returns the player's attitude with an NPC.
func (ct *ChoiceTracker) GetNPCAttitude(playerID, npcID string) float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	state, exists := ct.players[playerID]
	if !exists {
		return 0.0 // Neutral
	}

	rel, exists := state.NPCRelationships[npcID]
	if !exists {
		return 0.0 // Neutral
	}

	return rel.Attitude
}

// GetAlignment returns the player's current moral alignment.
func (ct *ChoiceTracker) GetAlignment(playerID string) *PlayerAlignment {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	state, exists := ct.players[playerID]
	if !exists {
		return &PlayerAlignment{
			GoodEvil:      0.0,
			LawChaos:      0.0,
			HonorDishonor: 0.0,
			UpdatedAt:     time.Now().Unix(),
		}
	}

	return state.Alignment
}

// RegisterQuestBranch registers a quest branch with prerequisites.
func (ct *ChoiceTracker) RegisterQuestBranch(branch *QuestBranch) error {
	if branch == nil {
		return fmt.Errorf("branch cannot be nil")
	}
	if branch.QuestID == "" {
		return fmt.Errorf("quest ID cannot be empty")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.questBranches[branch.QuestID] = branch
	return nil
}

// IsQuestBranchAvailable checks if a quest branch is available to a player.
func (ct *ChoiceTracker) IsQuestBranchAvailable(playerID, questID, branchID string) bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	branch, exists := ct.questBranches[questID]
	if !exists {
		return true // Quest not registered, assume available
	}

	if branch.BranchID != branchID {
		return false // Wrong branch
	}

	state, exists := ct.players[playerID]
	if !exists {
		return len(branch.Prerequisites) == 0 // Available if no prerequisites
	}

	// Check all prerequisites
	for _, prereq := range branch.Prerequisites {
		if !ct.hasChoice(state, prereq) {
			return false
		}
	}

	return true
}

// RegisterClassQuest registers a class-specific quest.
func (ct *ChoiceTracker) RegisterClassQuest(quest *ClassSpecificQuest) error {
	if quest == nil {
		return fmt.Errorf("quest cannot be nil")
	}
	if quest.QuestID == "" {
		return fmt.Errorf("quest ID cannot be empty")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.classQuests = append(ct.classQuests, quest)
	return nil
}

// IsClassQuestAvailable checks if a class-specific quest is available.
func (ct *ChoiceTracker) IsClassQuestAvailable(playerID, questID, playerClass string, playerLevel int) bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	quest := ct.findClassQuest(questID)
	if quest == nil {
		return false
	}

	if !ct.checkClassRequirement(quest, playerClass) {
		return false
	}

	if !ct.checkLevelRequirement(quest, playerLevel) {
		return false
	}

	return ct.checkAlignmentAndPrerequisites(playerID, quest)
}

// findClassQuest searches for a class quest by ID.
func (ct *ChoiceTracker) findClassQuest(questID string) *ClassSpecificQuest {
	for _, q := range ct.classQuests {
		if q.QuestID == questID {
			return q
		}
	}
	return nil
}

// checkClassRequirement validates class requirement for quest.
func (ct *ChoiceTracker) checkClassRequirement(quest *ClassSpecificQuest, playerClass string) bool {
	return quest.RequiredClass == playerClass
}

// checkLevelRequirement validates level requirement for quest.
func (ct *ChoiceTracker) checkLevelRequirement(quest *ClassSpecificQuest, playerLevel int) bool {
	return playerLevel >= quest.MinLevel
}

// checkAlignmentAndPrerequisites validates alignment and prerequisite choices.
func (ct *ChoiceTracker) checkAlignmentAndPrerequisites(playerID string, quest *ClassSpecificQuest) bool {
	state, exists := ct.players[playerID]
	if !exists {
		return len(quest.Prerequisites) == 0
	}

	if quest.AlignmentReq != nil && !state.Alignment.ChecksAlignment(quest.AlignmentReq) {
		return false
	}

	for _, prereq := range quest.Prerequisites {
		if !ct.hasChoice(state, prereq) {
			return false
		}
	}

	return true
}

// RecordCompanionReaction records how a companion reacted to a choice.
func (ct *ChoiceTracker) RecordCompanionReaction(playerID string, reaction *CompanionReaction) error {
	if reaction == nil {
		return fmt.Errorf("reaction cannot be nil")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	state := ct.getOrCreateState(playerID)
	state.CompanionReactions = append(state.CompanionReactions, reaction)

	// Keep last 20 reactions
	if len(state.CompanionReactions) > 20 {
		state.CompanionReactions = state.CompanionReactions[len(state.CompanionReactions)-20:]
	}

	return nil
}

// GetCompanionReactions returns recent companion reactions for a player.
func (ct *ChoiceTracker) GetCompanionReactions(playerID, companionID string) []*CompanionReaction {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	state, exists := ct.players[playerID]
	if !exists {
		return nil
	}

	reactions := make([]*CompanionReaction, 0)
	for _, r := range state.CompanionReactions {
		if r.CompanionID == companionID {
			reactions = append(reactions, r)
		}
	}

	return reactions
}

// Save saves all player choice data to a file.
func (ct *ChoiceTracker) Save(filename string) error {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	encoder := json.NewEncoder(gzWriter)
	if err := encoder.Encode(ct.players); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}

// Load loads player choice data from a file.
func (ct *ChoiceTracker) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	ct.mu.Lock()
	defer ct.mu.Unlock()

	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&ct.players); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}

// GetChoiceCount returns the number of choices a player has made.
func (ct *ChoiceTracker) GetChoiceCount(playerID string) int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	state, exists := ct.players[playerID]
	if !exists {
		return 0
	}

	return len(state.ChoiceHistory)
}

// GetNPCRelationshipCount returns the number of NPC relationships tracked.
func (ct *ChoiceTracker) GetNPCRelationshipCount(playerID string) int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	state, exists := ct.players[playerID]
	if !exists {
		return 0
	}

	return len(state.NPCRelationships)
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
