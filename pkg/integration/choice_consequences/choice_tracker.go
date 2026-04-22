package choice_consequences

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// choice_tracker.go implements the core ChoiceTracker manager that coordinates
// all choice-consequence functionality. This includes recording choices, managing
// NPC relationships, tracking content locks, handling quest branches, and persisting
// player choice state across sessions.
//
// The ChoiceTracker uses thread-safe operations (sync.RWMutex) to support
// concurrent access from multiple game systems.

// trackerLogger provides structured logging for choice tracker operations.
var trackerLogger = log.WithField("system", "choice_tracker")
var createChoiceTrackerFile = func(filename string) (io.WriteCloser, error) {
	return os.Create(filename)
}

func setSaveCloseError(retErr *error, closeErr error) {
	if closeErr != nil && *retErr == nil {
		*retErr = fmt.Errorf("failed to close file: %w", closeErr)
	}
}

// ChoiceTracker manages player choice tracking and consequence application.
type ChoiceTracker struct {
	mu                     sync.RWMutex
	players                map[string]*PlayerState // Player ID -> state
	questBranches          map[string]*QuestBranch // Quest ID -> branch info
	classQuests            []*ClassSpecificQuest   // Class-specific quests
	npcMemoryLimit         int                     // Max events per NPC (default 50)
	choiceLimit            int                     // Max choices tracked per player (default 200)
	companionReactionLimit int                     // Max companion reactions tracked per player (default 20)
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
		players:                make(map[string]*PlayerState),
		questBranches:          make(map[string]*QuestBranch),
		classQuests:            make([]*ClassSpecificQuest, 0),
		npcMemoryLimit:         50,
		choiceLimit:            200,
		companionReactionLimit: 20,
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
	state.LastUpdate = now()

	trackerLogger.WithFields(log.Fields{
		"player_id":     playerID,
		"choice_id":     choice.ChoiceID,
		"story_node_id": choice.StoryNodeID,
		"irreversible":  choice.Irreversible,
		"npcs_affected": len(choice.NPCsAffected),
		"consequences":  len(choice.Consequences),
	}).Debug("Player choice recorded")

	return nil
}

// validateChoice validates the choice parameters.
func validateChoice(choice *PlayerChoice) error {
	if choice == nil {
		trackerLogger.Error("Attempted to record nil choice")
		return fmt.Errorf("choice cannot be nil")
	}
	if choice.ChoiceID == "" {
		trackerLogger.WithField("story_node_id", choice.StoryNodeID).Error("Choice ID cannot be empty")
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
			UpdatedAt:     now(),
		},
		CompanionReactions: make([]*CompanionReaction, 0, 20),
		LastUpdate:         now(),
	}

	ct.players[playerID] = state
	return state
}

// updateNPCRelationship updates NPC attitude based on choice.
// updateNPCRelationship updates relationship state based on a player's choice.
func (ct *ChoiceTracker) updateNPCRelationship(state *PlayerState, npcID string, choice *PlayerChoice) {
	rel := ct.getOrCreateRelationship(state, npcID)
	impact := ct.calculateChoiceImpact(choice)

	ct.recordMemorableEvent(rel, choice, npcID, impact)
	ct.updateRelationshipValues(rel, impact)
}

// getOrCreateRelationship retrieves or creates an NPC relationship entry.
func (ct *ChoiceTracker) getOrCreateRelationship(state *PlayerState, npcID string) *NPCRelationship {
	rel, exists := state.NPCRelationships[npcID]
	if !exists {
		rel = &NPCRelationship{
			NPCID:           npcID,
			Attitude:        0.0,
			TrustLevel:      0.0,
			LastUpdate:      now(),
			MemorableEvents: make([]MemorableEvent, 0, ct.npcMemoryLimit),
			DialogueUnlocks: make([]string, 0),
			QuestsAffected:  make([]string, 0),
		}
		state.NPCRelationships[npcID] = rel
	}
	return rel
}

// calculateChoiceImpact calculates relationship impact from a choice's moral alignment.
func (ct *ChoiceTracker) calculateChoiceImpact(choice *PlayerChoice) float64 {
	if choice.MoralAlignment == nil {
		return 0.0
	}

	impact := choice.MoralAlignment.GoodEvil * 0.3
	impact += choice.MoralAlignment.HonorDishonor * 0.2
	return impact
}

// recordMemorableEvent adds a new memorable event and enforces memory limits.
func (ct *ChoiceTracker) recordMemorableEvent(rel *NPCRelationship, choice *PlayerChoice, npcID string, impact float64) {
	event := MemorableEvent{
		EventID:     fmt.Sprintf("%s_%s", choice.ChoiceID, npcID),
		ChoiceID:    choice.ChoiceID,
		Timestamp:   choice.Timestamp,
		Impact:      impact,
		Description: fmt.Sprintf("Player made choice: %s", choice.ChoiceID),
	}

	rel.MemorableEvents = append(rel.MemorableEvents, event)
	ct.enforceMemoryLimit(rel)
}

// enforceMemoryLimit keeps only the most impactful events within memory limit.
func (ct *ChoiceTracker) enforceMemoryLimit(rel *NPCRelationship) {
	if len(rel.MemorableEvents) <= ct.npcMemoryLimit {
		return
	}

	ct.sortEventsByImpact(rel.MemorableEvents)
	rel.MemorableEvents = rel.MemorableEvents[:ct.npcMemoryLimit]
}

// sortEventsByImpact sorts events by absolute impact in descending order.
func (ct *ChoiceTracker) sortEventsByImpact(events []MemorableEvent) {
	sort.Slice(events, func(i, j int) bool {
		return abs(events[i].Impact) > abs(events[j].Impact)
	})
}

// updateRelationshipValues updates attitude and trust based on choice impact.
func (ct *ChoiceTracker) updateRelationshipValues(rel *NPCRelationship, impact float64) {
	rel.Attitude += impact
	rel.Attitude = clamp(rel.Attitude, -1.0, 1.0)

	ct.adjustTrustLevel(rel)
	rel.LastUpdate = now()
}

// adjustTrustLevel modifies trust level based on current attitude.
func (ct *ChoiceTracker) adjustTrustLevel(rel *NPCRelationship) {
	if rel.Attitude > 0.0 {
		rel.TrustLevel += 0.05
	} else if rel.Attitude < 0.0 {
		rel.TrustLevel -= 0.1
	}
	rel.TrustLevel = clamp(rel.TrustLevel, 0.0, 1.0)
}

// lockTypePrefixes maps consequence prefixes to their LockType values.
var lockTypePrefixes = []struct {
	prefix   string
	lockType LockType
}{
	{"lock_quest_", LockTypeQuest},
	{"lock_npc_", LockTypeNPC},
	{"lock_area_", LockTypeArea},
	{"lock_dialogue_", LockTypeDialogue},
	{"lock_reward_", LockTypeReward},
	{"lock_companion_", LockTypeCompanion},
}

// applyConsequence applies a consequence to lock content.
func (ct *ChoiceTracker) applyConsequence(state *PlayerState, choiceID, consequence string) {
	// Parse consequence ID to determine lock type and content
	// Format: "lock_quest_<quest_id>" or "lock_npc_<npc_id>" etc.
	lockType := LockTypeQuest
	contentID := consequence

	for _, ltp := range lockTypePrefixes {
		if strings.HasPrefix(consequence, ltp.prefix) {
			lockType = ltp.lockType
			contentID = consequence[len(ltp.prefix):]
			break
		}
	}

	lock := &ContentLock{
		ContentID:        contentID,
		LockedBy:         choiceID,
		LockType:         lockType,
		Timestamp:        now(),
		Permanent:        true, // Most consequences are permanent
		UnlockConditions: make([]string, 0),
	}

	state.ContentLocks[contentID] = lock
}

// IsContentAvailable checks if content is available to a player.
func (ct *ChoiceTracker) IsContentAvailable(playerID, contentID string) bool {
	ct.mu.RLock()
	state, exists := ct.players[playerID]
	if !exists {
		ct.mu.RUnlock()
		return true // No choices made yet, all content available
	}

	lock, locked := state.ContentLocks[contentID]
	if !locked {
		ct.mu.RUnlock()
		return true
	}

	if lock.Permanent {
		ct.mu.RUnlock()
		return false
	}

	// Check if unlock conditions are met
	if len(lock.UnlockConditions) == 0 {
		ct.mu.RUnlock()
		return false
	}

	for _, condition := range lock.UnlockConditions {
		if !ct.hasChoice(state, condition) {
			ct.mu.RUnlock()
			return false
		}
	}

	// All unlock conditions met, need write lock to remove lock
	ct.mu.RUnlock()
	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Re-check state under write lock (may have changed)
	state, exists = ct.players[playerID]
	if !exists {
		return true
	}
	if _, stillLocked := state.ContentLocks[contentID]; stillLocked {
		delete(state.ContentLocks, contentID)
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
			UpdatedAt:     now(),
		}
	}

	return state.Alignment
}

// RegisterQuestBranch registers a quest branch with prerequisites.
func (ct *ChoiceTracker) RegisterQuestBranch(branch *QuestBranch) error {
	if branch == nil {
		trackerLogger.Error("Attempted to register nil quest branch")
		return fmt.Errorf("branch cannot be nil")
	}
	if branch.QuestID == "" {
		trackerLogger.Error("Quest branch ID cannot be empty")
		return fmt.Errorf("quest ID cannot be empty")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.questBranches[branch.QuestID] = branch

	trackerLogger.WithFields(log.Fields{
		"quest_id":      branch.QuestID,
		"branch_id":     branch.BranchID,
		"prerequisites": len(branch.Prerequisites),
	}).Debug("Quest branch registered")

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
		trackerLogger.Error("Attempted to register nil class quest")
		return fmt.Errorf("quest cannot be nil")
	}
	if quest.QuestID == "" {
		trackerLogger.Error("Class quest ID cannot be empty")
		return fmt.Errorf("quest ID cannot be empty")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.classQuests = append(ct.classQuests, quest)

	trackerLogger.WithFields(log.Fields{
		"quest_id":       quest.QuestID,
		"required_class": quest.RequiredClass,
		"min_level":      quest.MinLevel,
	}).Debug("Class quest registered")

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
		trackerLogger.WithField("player_id", playerID).Error("Attempted to record nil companion reaction")
		return fmt.Errorf("reaction cannot be nil")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	state := ct.getOrCreateState(playerID)
	state.CompanionReactions = append(state.CompanionReactions, reaction)

	// Keep last companionReactionLimit reactions
	if len(state.CompanionReactions) > ct.companionReactionLimit {
		state.CompanionReactions = state.CompanionReactions[len(state.CompanionReactions)-ct.companionReactionLimit:]
	}

	trackerLogger.WithFields(log.Fields{
		"player_id":     playerID,
		"companion_id":  reaction.CompanionID,
		"choice_id":     reaction.ChoiceID,
		"loyalty_delta": reaction.LoyaltyDelta,
		"approval":      reaction.Approval,
	}).Debug("Companion reaction recorded")

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

// SaveTo writes all player choice data to the provided io.Writer as gzip-compressed JSON.
// This method is WASM-compatible as it accepts any io.Writer implementation, enabling
// callers to provide platform-specific storage backends (e.g., localStorage via bytes.Buffer
// on WASM, or file-based storage on desktop).
func (ct *ChoiceTracker) SaveTo(w io.Writer) error {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	trackerLogger.WithFields(log.Fields{
		"player_count": len(ct.players),
	}).Debug("Saving choice tracker data to writer")

	gzWriter := gzip.NewWriter(w)

	encoder := json.NewEncoder(gzWriter)
	if err := encoder.Encode(ct.players); err != nil {
		gzWriter.Close()
		trackerLogger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to encode choice tracker data")
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to flush gzip writer: %w", err)
	}

	trackerLogger.WithFields(log.Fields{
		"player_count": len(ct.players),
	}).Debug("Choice tracker data serialized successfully")

	return nil
}

// Save saves all player choice data to a file.
// For WASM builds, use SaveTo with a bytes.Buffer and persist via localStorage.
func (ct *ChoiceTracker) Save(filename string) (err error) {
	trackerLogger.WithFields(log.Fields{
		"filename": filename,
	}).Debug("Saving choice tracker data to file")

	file, err := createChoiceTrackerFile(filename)
	if err != nil {
		trackerLogger.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("Failed to create save file")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			setSaveCloseError(&err, closeErr)
			trackerLogger.WithFields(log.Fields{
				"filename": filename,
				"error":    closeErr.Error(),
			}).Error("Failed to close choice tracker save file")
		}
	}()

	if err := ct.SaveTo(file); err != nil {
		return err
	}

	trackerLogger.WithFields(log.Fields{
		"filename": filename,
	}).Info("Choice tracker data saved successfully")

	return nil
}

// LoadFrom loads player choice data from the provided io.Reader containing gzip-compressed JSON.
// This method is WASM-compatible as it accepts any io.Reader implementation, enabling
// callers to provide platform-specific storage backends (e.g., bytes.Reader backed by
// localStorage on WASM, or file-based storage on desktop).
func (ct *ChoiceTracker) LoadFrom(r io.Reader) error {
	trackerLogger.Debug("Loading choice tracker data from reader")

	gzReader, err := gzip.NewReader(r)
	if err != nil {
		trackerLogger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to create gzip reader")
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	ct.mu.Lock()
	defer ct.mu.Unlock()

	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&ct.players); err != nil {
		trackerLogger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to decode choice tracker data")
		return fmt.Errorf("failed to decode data: %w", err)
	}

	trackerLogger.WithFields(log.Fields{
		"player_count": len(ct.players),
	}).Debug("Choice tracker data deserialized successfully")

	return nil
}

// Load loads player choice data from a file.
// For WASM builds, use LoadFrom with a bytes.Reader backed by localStorage data.
func (ct *ChoiceTracker) Load(filename string) error {
	trackerLogger.WithField("filename", filename).Debug("Loading choice tracker data from file")

	file, err := os.Open(filename)
	if err != nil {
		trackerLogger.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error("Failed to open save file")
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if err := ct.LoadFrom(file); err != nil {
		return err
	}

	trackerLogger.WithFields(log.Fields{
		"filename": filename,
	}).Info("Choice tracker data loaded successfully")

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
