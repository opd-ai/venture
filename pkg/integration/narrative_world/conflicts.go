package narrative_world

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/story"
)

// CheckConflict determines if two companions have a personality conflict
func (m *StoryEventManager) CheckConflict(comp1, comp2 *engine.CompanionComponent, comp1ID, comp2ID uint64, personality1, personality2 *learning.PersonalityEvolution) (CompanionConflict, bool) {
	// Check if conflict already exists
	for _, conflict := range m.conflicts {
		if (conflict.Companion1 == comp1ID && conflict.Companion2 == comp2ID) ||
			(conflict.Companion1 == comp2ID && conflict.Companion2 == comp1ID) {
			return conflict, true
		}
	}

	// Calculate conflict probability based on personality clash
	conflictProb := m.calculateConflictProbability(personality1, personality2)

	rng := rand.New(rand.NewSource(int64(comp1ID) + int64(comp2ID)))

	// 10-20% base chance, modified by personality compatibility
	if rng.Float64() < conflictProb*m.conflictChance {
		conflictType := m.determineConflictType(personality1, personality2, rng)

		conflict := CompanionConflict{
			Companion1:      comp1ID,
			Companion2:      comp2ID,
			ConflictType:    conflictType,
			Description:     m.generateConflictDescription(conflictType, comp1, comp2),
			Severity:        0.3 + rng.Float64()*0.7, // 0.3-1.0
			ResolutionQuest: nil,                     // Generated on demand
			Active:          true,
			TimeSinceStart:  0,
		}

		m.conflicts = append(m.conflicts, conflict)
		return conflict, true
	}

	return CompanionConflict{}, false
}

// calculateConflictProbability determines likelihood of conflict based on personality traits
func (m *StoryEventManager) calculateConflictProbability(p1, p2 *learning.PersonalityEvolution) float64 {
	if p1 == nil || p2 == nil {
		return 0.5 // Default 50% modifier if no personality data
	}

	conflictScore := 0.0
	comparisons := 0

	// Opposing trait pairs increase conflict
	opposingPairs := [][2]learning.PersonalityTrait{
		{learning.TraitCautious, learning.TraitBrave},
		{learning.TraitShy, learning.TraitOutgoing},
		{learning.TraitAggressive, learning.TraitPacifist},
		{learning.TraitLoyal, learning.TraitIndependent},
		{learning.TraitCurious, learning.TraitPractical},
	}

	for _, pair := range opposingPairs {
		strength1A := p1.Traits[pair[0]]
		strength1B := p1.Traits[pair[1]]
		strength2A := p2.Traits[pair[0]]
		strength2B := p2.Traits[pair[1]]

		// If one companion has high trait A and other has high trait B (opposing)
		if strength1A > 0.7 && strength2B > 0.7 {
			conflictScore += strength1A * strength2B
			comparisons++
		}
		if strength1B > 0.7 && strength2A > 0.7 {
			conflictScore += strength1B * strength2A
			comparisons++
		}
	}

	if comparisons == 0 {
		return 0.5 // No strong opposing traits
	}

	return conflictScore / float64(comparisons)
}

// determineConflictType selects conflict type based on personalities
func (m *StoryEventManager) determineConflictType(p1, p2 *learning.PersonalityEvolution, rng *rand.Rand) ConflictType {
	if p1 == nil || p2 == nil {
		return ConflictType(rng.Intn(5))
	}

	// Aggressive vs Pacifist -> Personality clash
	if (p1.Traits[learning.TraitAggressive] > 0.7 && p2.Traits[learning.TraitPacifist] > 0.7) ||
		(p2.Traits[learning.TraitAggressive] > 0.7 && p1.Traits[learning.TraitPacifist] > 0.7) {
		return ConflictPersonality
	}

	// Both aggressive -> Rivalry
	if p1.Traits[learning.TraitAggressive] > 0.7 && p2.Traits[learning.TraitAggressive] > 0.7 {
		return ConflictRivalry
	}

	// Practical vs Curious -> Beliefs
	if (p1.Traits[learning.TraitPractical] > 0.7 && p2.Traits[learning.TraitCurious] > 0.7) ||
		(p2.Traits[learning.TraitPractical] > 0.7 && p1.Traits[learning.TraitCurious] > 0.7) {
		return ConflictBeliefs
	}

	// Random for others
	return ConflictType(rng.Intn(5))
}

// generateConflictDescription creates a description for the conflict
func (m *StoryEventManager) generateConflictDescription(conflictType ConflictType, comp1, comp2 *engine.CompanionComponent) string {
	templates := map[ConflictType]string{
		ConflictPersonality:         "Your companions have fundamentally incompatible personalities",
		ConflictRivalry:             "Your companions compete for your attention and approval",
		ConflictBeliefs:             "Your companions disagree on fundamental principles",
		ConflictPastHistory:         "Your companions share a troubled past",
		ConflictResourceCompetition: "Your companions compete for limited resources",
	}

	return templates[conflictType]
}

// GenerateCrossCompanionStory creates a narrative involving multiple companions
func (m *StoryEventManager) GenerateCrossCompanionStory(companionIDs []uint64, seed int64) (*CrossCompanionStory, error) {
	if len(companionIDs) < 2 {
		return nil, fmt.Errorf("cross-companion story requires at least 2 companions, got %d", len(companionIDs))
	}

	rng := rand.New(rand.NewSource(seed))

	// Generate branching narrative
	params := procgen.GenerationParams{
		Difficulty: 0.5 + rng.Float64()*0.5, // 0.5-1.0
		Depth:      5 + rng.Intn(10),        // 5-14
		GenreID:    "fantasy",
	}

	narrative, err := m.narrativeGen.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cross-companion narrative: %w", err)
	}

	// Gather memories from all participants
	events := make([]MemoryEvent, 0)
	for _, compID := range companionIDs {
		if memory, exists := m.memories[compID]; exists {
			// Add recent important events (last 5 important events)
			count := 0
			for i := len(memory.Events) - 1; i >= 0 && count < 5; i-- {
				if memory.Events[i].Importance >= 0.7 {
					events = append(events, memory.Events[i])
					count++
				}
			}
		}
	}

	storyTitles := []string{
		"Bonds of Adversity",
		"Shared Destiny",
		"Trials of Friendship",
		"Conflicting Paths",
		"United Purpose",
		"Tangled Fates",
	}

	storyDescriptions := []string{
		"Your companions' stories intertwine in unexpected ways",
		"A shared challenge tests the bonds between your companions",
		"Past events resurface, forcing your companions to confront their relationships",
		"A common enemy threatens to tear your companions apart",
		"Your companions must work together to overcome a mutual obstacle",
	}

	story := &CrossCompanionStory{
		StoryID:      fmt.Sprintf("cross-story-%d", seed),
		Title:        storyTitles[rng.Intn(len(storyTitles))],
		Description:  storyDescriptions[rng.Intn(len(storyDescriptions))],
		Participants: companionIDs,
		Events:       events,
		Narrative:    narrative.(*story.BranchingNarrative),
		Outcome:      OutcomeUnresolved,
		Active:       true,
	}

	m.crossStories = append(m.crossStories, story)
	return story, nil
}

// GetDialogueContext retrieves memory-based dialogue context
func (m *StoryEventManager) GetDialogueContext(companionID uint64, maxRecent int) *DialogueContext {
	memory, exists := m.memories[companionID]
	if !exists || len(memory.Events) == 0 {
		return m.createEmptyDialogueContext()
	}

	recentEvents := m.extractRecentEvents(memory.Events, maxRecent)
	importantEvents := m.extractImportantEvents(memory.Events)
	relatedTo := m.extractRelatedEntityIDs(memory.Events, companionID)

	return &DialogueContext{
		RecentEvents:    recentEvents,
		ImportantEvents: importantEvents,
		RelatedTo:       relatedTo,
	}
}

// createEmptyDialogueContext creates a dialogue context with empty slices.
func (m *StoryEventManager) createEmptyDialogueContext() *DialogueContext {
	return &DialogueContext{
		RecentEvents:    make([]MemoryEvent, 0),
		ImportantEvents: make([]MemoryEvent, 0),
		RelatedTo:       make([]uint64, 0),
	}
}

// extractRecentEvents retrieves the most recent memory events up to maxRecent.
func (m *StoryEventManager) extractRecentEvents(events []MemoryEvent, maxRecent int) []MemoryEvent {
	recentCount := maxRecent
	if recentCount > len(events) {
		recentCount = len(events)
	}

	recentEvents := make([]MemoryEvent, recentCount)
	copy(recentEvents, events[len(events)-recentCount:])

	return recentEvents
}

// extractImportantEvents filters memory events by importance threshold.
func (m *StoryEventManager) extractImportantEvents(events []MemoryEvent) []MemoryEvent {
	importantEvents := make([]MemoryEvent, 0)
	for _, event := range events {
		if event.Importance >= 0.7 {
			importantEvents = append(importantEvents, event)
		}
	}
	return importantEvents
}

// extractRelatedEntityIDs collects unique entity IDs from event participants.
func (m *StoryEventManager) extractRelatedEntityIDs(events []MemoryEvent, companionID uint64) []uint64 {
	relatedMap := make(map[uint64]bool)
	for _, event := range events {
		for _, participant := range event.Participants {
			if participant != companionID {
				relatedMap[participant] = true
			}
		}
	}

	relatedTo := make([]uint64, 0, len(relatedMap))
	for id := range relatedMap {
		relatedTo = append(relatedTo, id)
	}

	return relatedTo
}

// UpdateConflict updates conflict state over time
func (m *StoryEventManager) UpdateConflict(conflictIndex int, deltaTime time.Duration) {
	if conflictIndex < 0 || conflictIndex >= len(m.conflicts) {
		return
	}

	conflict := &m.conflicts[conflictIndex]
	conflict.TimeSinceStart += deltaTime

	// Conflicts may escalate or diminish over time
	// Severity increases by 0.1 per day if unresolved
	daysElapsed := conflict.TimeSinceStart.Hours() / 24.0
	severityIncrease := daysElapsed * 0.1

	conflict.Severity = math.Min(1.0, conflict.Severity+severityIncrease)
}

// ResolveConflict marks a conflict as resolved
func (m *StoryEventManager) ResolveConflict(comp1ID, comp2ID uint64) bool {
	for i := range m.conflicts {
		conflict := &m.conflicts[i]
		if (conflict.Companion1 == comp1ID && conflict.Companion2 == comp2ID) ||
			(conflict.Companion1 == comp2ID && conflict.Companion2 == comp1ID) {
			conflict.Active = false
			return true
		}
	}
	return false
}

// CompleteQuest marks a quest as completed and applies consequences
func (m *StoryEventManager) CompleteQuest(companionID uint64, questID string) (*PersonalQuest, error) {
	quests, exists := m.activeQuests[companionID]
	if !exists {
		return nil, fmt.Errorf("no active quests for companion %d", companionID)
	}

	for _, quest := range quests {
		if quest.QuestID == questID {
			// Check all objectives completed
			allComplete := true
			for _, obj := range quest.Objectives {
				if !obj.Completed {
					allComplete = false
					break
				}
			}

			if !allComplete {
				return nil, fmt.Errorf("quest %s has incomplete objectives", questID)
			}

			quest.Completed = true

			// Record completion as important memory event
			m.RecordMemory(companionID, EventTypeBonding, fmt.Sprintf("Completed quest: %s", quest.Title))

			return quest, nil
		}
	}

	return nil, fmt.Errorf("quest %s not found for companion %d", questID, companionID)
}

// UpdateQuestObjective updates progress on a quest objective
func (m *StoryEventManager) UpdateQuestObjective(companionID uint64, questID string, objectiveIndex, progress int) error {
	quests, exists := m.activeQuests[companionID]
	if !exists {
		return fmt.Errorf("no active quests for companion %d", companionID)
	}

	for _, quest := range quests {
		if quest.QuestID == questID {
			if objectiveIndex < 0 || objectiveIndex >= len(quest.Objectives) {
				return fmt.Errorf("invalid objective index %d", objectiveIndex)
			}

			objective := &quest.Objectives[objectiveIndex]
			objective.Progress = progress

			if objective.Progress >= objective.Required {
				objective.Completed = true
			}

			return nil
		}
	}

	return fmt.Errorf("quest %s not found", questID)
}

// GetActiveQuests returns all active quests for a companion
func (m *StoryEventManager) GetActiveQuests(companionID uint64) []*PersonalQuest {
	quests, exists := m.activeQuests[companionID]
	if !exists {
		return make([]*PersonalQuest, 0)
	}
	return quests
}

// GetMemoryCount returns the number of stored memories for a companion
func (m *StoryEventManager) GetMemoryCount(companionID uint64) int {
	memory, exists := m.memories[companionID]
	if !exists {
		return 0
	}
	return len(memory.Events)
}

// GetTotalEventsRecorded returns total events ever recorded (including pruned)
func (m *StoryEventManager) GetTotalEventsRecorded(companionID uint64) int {
	memory, exists := m.memories[companionID]
	if !exists {
		return 0
	}
	return memory.TotalEvents
}

// GetActiveConflicts returns all active conflicts
func (m *StoryEventManager) GetActiveConflicts() []CompanionConflict {
	active := make([]CompanionConflict, 0)
	for _, conflict := range m.conflicts {
		if conflict.Active {
			active = append(active, conflict)
		}
	}
	return active
}

// GetActiveCrossStories returns all active cross-companion stories
func (m *StoryEventManager) GetActiveCrossStories() []*CrossCompanionStory {
	active := make([]*CrossCompanionStory, 0)
	for _, story := range m.crossStories {
		if story.Active {
			active = append(active, story)
		}
	}
	return active
}

// SetConflictChance sets the probability of conflicts (0.1-0.2 typical)
func (m *StoryEventManager) SetConflictChance(chance float64) {
	if chance < 0.0 {
		chance = 0.0
	}
	if chance > 1.0 {
		chance = 1.0
	}
	m.conflictChance = chance
}

// SetMaxMemoryEvents sets the maximum memory events per companion
func (m *StoryEventManager) SetMaxMemoryEvents(max int) {
	if max < 1 {
		max = 1 // Minimum 1 event
	}
	if max > 200 {
		max = 200 // Maximum 200 events
	}
	m.maxMemoryEvents = max
}
