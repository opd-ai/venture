package narrative_world

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/story"
)

// StoryEventManager manages companion-driven narrative events
type StoryEventManager struct {
	seed            int64
	genreID         string
	memories        map[uint64]*CompanionMemory // CompanionID -> Memory
	activeQuests    map[uint64][]*PersonalQuest // CompanionID -> Quests
	conflicts       []CompanionConflict
	crossStories    []*CrossCompanionStory
	narrativeGen    *story.BranchingNarrativeGenerator
	questTemplates  map[engine.CompanionType][]QuestTemplate
	conflictChance  float64 // 0.10-0.20 (10-20% of interactions)
	maxMemoryEvents int     // 50-100 per companion
	timeProvider    TimeProvider
}

// QuestTemplate defines quest generation parameters
type QuestTemplate struct {
	TitlePattern    string
	DescPattern     string
	ObjectiveTypes  []ObjectiveType
	ConsequenceType ConsequenceType
	PersonalityFit  []learning.PersonalityTrait
	MinLoyalty      float64
}

// ManagerOption is a functional option for configuring StoryEventManager.
type ManagerOption func(*StoryEventManager)

// WithTimeProvider sets a custom TimeProvider for deterministic timestamps.
// Use FixedTimeProvider or IncrementingTimeProvider for testing,
// or a game-clock-based provider for deterministic multiplayer.
func WithTimeProvider(tp TimeProvider) ManagerOption {
	return func(m *StoryEventManager) {
		if tp != nil {
			m.timeProvider = tp
		}
	}
}

// WithGenreID sets the genre for story content generation.
func WithGenreID(genreID string) ManagerOption {
	return func(m *StoryEventManager) {
		if genreID != "" {
			m.genreID = genreID
		}
	}
}

// NewStoryEventManager creates a new story event manager.
// Optional ManagerOption functions may be provided to customize behavior:
//   - WithTimeProvider(tp): sets a custom TimeProvider for deterministic timestamps
//
// If no TimeProvider is provided, the package-level defaultTimeProvider is used.
func NewStoryEventManager(seed int64, opts ...ManagerOption) *StoryEventManager {
	manager := &StoryEventManager{
		seed:            seed,
		memories:        make(map[uint64]*CompanionMemory),
		activeQuests:    make(map[uint64][]*PersonalQuest),
		conflicts:       make([]CompanionConflict, 0),
		crossStories:    make([]*CrossCompanionStory, 0),
		narrativeGen:    story.NewBranchingNarrativeGenerator(),
		questTemplates:  make(map[engine.CompanionType][]QuestTemplate),
		conflictChance:  0.15, // 15% default
		maxMemoryEvents: 75,   // Default 75 events
		timeProvider:    nil,  // Use package default if not set
		genreID:         "fantasy",
	}

	// Apply options
	for _, opt := range opts {
		opt(manager)
	}

	manager.initializeQuestTemplates()
	return manager
}

// now returns the current timestamp using the manager's TimeProvider if set,
// otherwise falls back to the package-level defaultTimeProvider.
func (m *StoryEventManager) now() int64 {
	if m.timeProvider != nil {
		return m.timeProvider.Now()
	}
	return now()
}

// initializeQuestTemplates creates 3-5 quest templates per companion type
func (m *StoryEventManager) initializeQuestTemplates() {
	// Pet quests (5 templates)
	m.questTemplates[engine.CompanionTypePet] = []QuestTemplate{
		{
			TitlePattern:    "Lost Home",
			DescPattern:     "Your companion wants to revisit their place of origin",
			ObjectiveTypes:  []ObjectiveType{ObjectiveVisit, ObjectiveExplore},
			ConsequenceType: ConsequenceLoyaltyChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitCurious, learning.TraitShy},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Pack Reunion",
			DescPattern:     "Your companion seeks their former pack members",
			ObjectiveTypes:  []ObjectiveType{ObjectiveTalk, ObjectiveVisit},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitOutgoing, learning.TraitLoyal},
			MinLoyalty:      0.75,
		},
		{
			TitlePattern:    "Territorial Defense",
			DescPattern:     "Protect your companion's claimed territory from invaders",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveProtect},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitBrave, learning.TraitAggressive},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Lost Treasure",
			DescPattern:     "Your companion buried something valuable long ago",
			ObjectiveTypes:  []ObjectiveType{ObjectiveExplore, ObjectiveCollect},
			ConsequenceType: ConsequenceItemGain,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitPractical, learning.TraitCurious},
			MinLoyalty:      0.8,
		},
		{
			TitlePattern:    "Old Enemy",
			DescPattern:     "A creature from your companion's past threatens your bond",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat},
			ConsequenceType: ConsequenceDeath,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitBrave, learning.TraitLoyal},
			MinLoyalty:      0.9,
		},
	}

	// Summon quests (4 templates)
	m.questTemplates[engine.CompanionTypeSummon] = []QuestTemplate{
		{
			TitlePattern:    "Planar Instability",
			DescPattern:     "Your summon's connection to this plane is weakening",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveVisit},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitIndependent},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Elemental Balance",
			DescPattern:     "Restore balance to the elemental forces sustaining your summon",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveCollect},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitPractical},
			MinLoyalty:      0.75,
		},
		{
			TitlePattern:    "Binding Contract",
			DescPattern:     "Strengthen the magical contract binding your summon",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveTalk},
			ConsequenceType: ConsequenceLoyaltyChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitLoyal},
			MinLoyalty:      0.8,
		},
		{
			TitlePattern:    "Otherworldly Threat",
			DescPattern:     "A planar entity seeks to reclaim your summon",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveProtect},
			ConsequenceType: ConsequenceDeath,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitBrave},
			MinLoyalty:      0.85,
		},
	}

	// Hireling quests (5 templates)
	m.questTemplates[engine.CompanionTypeHireling] = []QuestTemplate{
		{
			TitlePattern:    "Personal Debt",
			DescPattern:     "Your hireling owes a debt that threatens their freedom",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveTalk},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitPractical, learning.TraitCautious},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Family Matters",
			DescPattern:     "Your hireling's family needs assistance",
			ObjectiveTypes:  []ObjectiveType{ObjectiveProtect, ObjectiveDefeat},
			ConsequenceType: ConsequenceLoyaltyChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitLoyal},
			MinLoyalty:      0.75,
		},
		{
			TitlePattern:    "Career Advancement",
			DescPattern:     "Help your hireling achieve their professional goals",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveCollect},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitAggressive, learning.TraitPractical},
			MinLoyalty:      0.8,
		},
		{
			TitlePattern:    "Betrayal from the Past",
			DescPattern:     "A former employer seeks revenge on your hireling",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveTalk},
			ConsequenceType: ConsequenceRelationshipChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitCautious},
			MinLoyalty:      0.85,
		},
		{
			TitlePattern:    "Ultimate Loyalty Test",
			DescPattern:     "Your hireling must choose between you and their past",
			ObjectiveTypes:  []ObjectiveType{ObjectiveTalk},
			ConsequenceType: ConsequenceDeath,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitLoyal, learning.TraitBrave},
			MinLoyalty:      0.9,
		},
	}

	// Elemental quests (3 templates)
	m.questTemplates[engine.CompanionTypeElemental] = []QuestTemplate{
		{
			TitlePattern:    "Elemental Rift",
			DescPattern:     "Seal a rift that threatens your elemental's existence",
			ObjectiveTypes:  []ObjectiveType{ObjectiveVisit, ObjectiveDefeat},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitBrave},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Pure Element",
			DescPattern:     "Find a source of pure elemental energy",
			ObjectiveTypes:  []ObjectiveType{ObjectiveExplore, ObjectiveCollect},
			ConsequenceType: ConsequenceLoyaltyChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitCurious},
			MinLoyalty:      0.8,
		},
		{
			TitlePattern:    "Return to Origin",
			DescPattern:     "Your elemental yearns to return to their home plane",
			ObjectiveTypes:  []ObjectiveType{ObjectiveVisit, ObjectiveTalk},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitIndependent},
			MinLoyalty:      0.85,
		},
	}

	// Add templates for remaining types
	m.addRemainingTemplates()
}

func (m *StoryEventManager) addRemainingTemplates() {
	// Undead quests (4 templates)
	m.questTemplates[engine.CompanionTypeUndead] = []QuestTemplate{
		{
			TitlePattern:    "Lingering Regret",
			DescPattern:     "Your undead companion seeks closure from their past life",
			ObjectiveTypes:  []ObjectiveType{ObjectiveVisit, ObjectiveTalk},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitShy, learning.TraitCautious},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Necromantic Power",
			DescPattern:     "Strengthen the dark magic sustaining your companion",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveDefeat},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitPractical},
			MinLoyalty:      0.75,
		},
		{
			TitlePattern:    "Final Rest",
			DescPattern:     "Your companion seeks the peace of true death",
			ObjectiveTypes:  []ObjectiveType{ObjectiveTalk, ObjectiveVisit},
			ConsequenceType: ConsequenceDeath,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitPacifist},
			MinLoyalty:      0.9,
		},
		{
			TitlePattern:    "Curse Breaking",
			DescPattern:     "Break the curse that binds your companion to undeath",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveDefeat},
			ConsequenceType: ConsequenceLoyaltyChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitBrave},
			MinLoyalty:      0.8,
		},
	}

	// Robot quests (3 templates)
	m.questTemplates[engine.CompanionTypeRobot] = []QuestTemplate{
		{
			TitlePattern:    "System Upgrade",
			DescPattern:     "Find components to upgrade your robot's systems",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveExplore},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitPractical},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "AI Awakening",
			DescPattern:     "Your robot is developing true consciousness",
			ObjectiveTypes:  []ObjectiveType{ObjectiveTalk, ObjectiveVisit},
			ConsequenceType: ConsequenceLoyaltyChange,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitCurious},
			MinLoyalty:      0.85,
		},
		{
			TitlePattern:    "Original Creator",
			DescPattern:     "Your robot seeks their creator for answers",
			ObjectiveTypes:  []ObjectiveType{ObjectiveVisit, ObjectiveTalk},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitIndependent},
			MinLoyalty:      0.9,
		},
	}

	// Spirit quests (3 templates)
	m.questTemplates[engine.CompanionTypeSpirit] = []QuestTemplate{
		{
			TitlePattern:    "Spiritual Anchor",
			DescPattern:     "Your spirit needs an anchor to remain in the mortal realm",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveVisit},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitCautious},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Ethereal Vision",
			DescPattern:     "Your spirit companion sees a threat only they can perceive",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveProtect},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitBrave},
			MinLoyalty:      0.8,
		},
		{
			TitlePattern:    "Transcendence",
			DescPattern:     "Your spirit seeks to ascend to a higher plane",
			ObjectiveTypes:  []ObjectiveType{ObjectiveTalk, ObjectiveVisit},
			ConsequenceType: ConsequenceDeath,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitIndependent},
			MinLoyalty:      0.9,
		},
	}

	// Insect quests (3 templates)
	m.questTemplates[engine.CompanionTypeInsect] = []QuestTemplate{
		{
			TitlePattern:    "Hive Calling",
			DescPattern:     "Your insect companion feels the pull of the collective hive",
			ObjectiveTypes:  []ObjectiveType{ObjectiveVisit, ObjectiveTalk},
			ConsequenceType: ConsequenceDeparture,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitShy},
			MinLoyalty:      0.7,
		},
		{
			TitlePattern:    "Metamorphosis",
			DescPattern:     "Your companion is ready to undergo a transformative change",
			ObjectiveTypes:  []ObjectiveType{ObjectiveCollect, ObjectiveProtect},
			ConsequenceType: ConsequenceSkillUnlock,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitCurious},
			MinLoyalty:      0.8,
		},
		{
			TitlePattern:    "Queen's Command",
			DescPattern:     "The hive queen demands your companion's return",
			ObjectiveTypes:  []ObjectiveType{ObjectiveDefeat, ObjectiveTalk},
			ConsequenceType: ConsequenceDeath,
			PersonalityFit:  []learning.PersonalityTrait{learning.TraitLoyal},
			MinLoyalty:      0.85,
		},
	}
}

// GeneratePersonalQuest creates a quest for a companion at loyalty 0.7+
func (m *StoryEventManager) GeneratePersonalQuest(companionID uint64, companion *engine.CompanionComponent, seed int64) (*PersonalQuest, error) {
	if companion.Loyalty < 0.7 {
		return nil, fmt.Errorf("companion loyalty %.2f below minimum 0.7 for personal quests", companion.Loyalty)
	}

	templates, exists := m.questTemplates[companion.CompanionType]
	if !exists || len(templates) == 0 {
		return nil, fmt.Errorf("no quest templates for companion type %d", companion.CompanionType)
	}

	rng := rand.New(rand.NewSource(seed))

	// Select template based on loyalty (higher loyalty = more severe quests)
	var selectedTemplate QuestTemplate
	templateFound := false
	for _, template := range templates {
		if companion.Loyalty >= template.MinLoyalty {
			selectedTemplate = template
			templateFound = true
			if rng.Float64() < 0.7 { // 70% chance to use first matching
				break
			}
		}
	}

	if !templateFound {
		return nil, fmt.Errorf("no matching quest template for companion type %d at loyalty %.2f", companion.CompanionType, companion.Loyalty)
	}

	// Generate branching narrative for quest
	params := procgen.GenerationParams{
		Difficulty: companion.Loyalty, // Higher loyalty = harder quest
		Depth:      companion.Level,
		GenreID:    m.genreID,
	}

	narrative, err := m.narrativeGen.Generate(seed+int64(companionID), params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quest narrative: %w", err)
	}

	// Generate objectives
	numObjectives := 2 + rng.Intn(3) // 2-4 objectives
	objectives := make([]QuestObjective, numObjectives)
	for i := 0; i < numObjectives; i++ {
		objType := selectedTemplate.ObjectiveTypes[rng.Intn(len(selectedTemplate.ObjectiveTypes))]
		objectives[i] = QuestObjective{
			Description: m.generateObjectiveDescription(objType, companion.CompanionType, rng),
			Type:        objType,
			Target:      m.generateObjectiveTarget(objType, rng),
			Progress:    0,
			Required:    1 + rng.Intn(5), // 1-5 required
			Completed:   false,
		}
	}

	// Generate consequences
	consequences := []Consequence{
		{
			Type:        selectedTemplate.ConsequenceType,
			Description: m.generateConsequenceDescription(selectedTemplate.ConsequenceType),
			Permanent:   true,                    // All quest consequences are permanent
			Severity:    0.5 + rng.Float64()*0.5, // 0.5-1.0
		},
	}

	quest := &PersonalQuest{
		QuestID:         fmt.Sprintf("quest-%d-%d", companionID, seed),
		CompanionID:     companionID,
		CompanionType:   companion.CompanionType,
		Title:           selectedTemplate.TitlePattern,
		Description:     selectedTemplate.DescPattern,
		Objectives:      objectives,
		UnlockLoyalty:   selectedTemplate.MinLoyalty,
		Completed:       false,
		Started:         false,
		StoryBranches:   narrative.(*story.BranchingNarrative),
		Consequences:    consequences,
		PersonalityReqs: selectedTemplate.PersonalityFit,
	}

	// Track active quest
	m.activeQuests[companionID] = append(m.activeQuests[companionID], quest)

	return quest, nil
}

// RecordMemory adds a significant event to companion memory
func (m *StoryEventManager) RecordMemory(companionID uint64, eventType EventType, description string) {
	memory, exists := m.memories[companionID]
	if !exists {
		memory = &CompanionMemory{
			CompanionID: companionID,
			Events:      make([]MemoryEvent, 0, m.maxMemoryEvents),
			MaxEvents:   m.maxMemoryEvents,
			TotalEvents: 0,
		}
		m.memories[companionID] = memory
	}

	event := MemoryEvent{
		Timestamp:    m.now(), // Deterministic via manager's TimeProvider
		Type:         eventType,
		Description:  description,
		Participants: []uint64{companionID},
		Location:     "unknown", // Could be enhanced with location tracking
		Importance:   m.calculateImportance(eventType),
	}

	memory.Events = append(memory.Events, event)
	memory.TotalEvents++

	// Prune oldest events if over limit (keep most important)
	if len(memory.Events) > memory.MaxEvents {
		m.pruneOldMemories(memory)
	}
}

// calculateImportance assigns importance to event types
func (m *StoryEventManager) calculateImportance(eventType EventType) float64 {
	switch eventType {
	case EventTypeSacrifice, EventTypeBetray:
		return 1.0 // Maximum importance
	case EventTypeDanger, EventTypeConflict:
		return 0.8
	case EventTypeBonding, EventTypeDiscovery:
		return 0.7
	case EventTypeCombat:
		return 0.5
	case EventTypeTreasure:
		return 0.4
	default:
		return 0.3
	}
}

// pruneOldMemories removes least important old events
func (m *StoryEventManager) pruneOldMemories(memory *CompanionMemory) {
	// Keep events sorted by importance * recency
	type scoredEvent struct {
		event MemoryEvent
		score float64
	}

	scored := make([]scoredEvent, len(memory.Events))
	currentTime := m.now() // Deterministic via manager's TimeProvider

	// Seconds in a year for recency calculation
	const secondsPerYear = 365 * 24 * 3600

	for i, event := range memory.Events {
		// Calculate recency based on Unix timestamp difference (seconds)
		ageSeconds := currentTime - event.Timestamp
		recency := 1.0 - (float64(ageSeconds) / float64(secondsPerYear)) // Decay over year
		if recency < 0 {
			recency = 0
		}
		scored[i] = scoredEvent{
			event: event,
			score: event.Importance * (0.7 + 0.3*recency), // 70% importance, 30% recency
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Keep top MaxEvents
	memory.Events = make([]MemoryEvent, 0, memory.MaxEvents)
	for i := 0; i < memory.MaxEvents && i < len(scored); i++ {
		memory.Events = append(memory.Events, scored[i].event)
	}
}

// Helper functions for quest generation
func (m *StoryEventManager) generateObjectiveDescription(objType ObjectiveType, compType engine.CompanionType, rng *rand.Rand) string {
	descriptions := map[ObjectiveType][]string{
		ObjectiveDefeat:  {"Defeat threatening enemies", "Vanquish hostile forces", "Eliminate dangerous foes"},
		ObjectiveCollect: {"Gather valuable items", "Collect rare artifacts", "Find special materials"},
		ObjectiveVisit:   {"Travel to significant location", "Reach important destination", "Explore sacred site"},
		ObjectiveProtect: {"Defend against attacks", "Guard important person", "Protect valuable objective"},
		ObjectiveTalk:    {"Speak with key individual", "Negotiate with important NPC", "Gain critical information"},
		ObjectiveExplore: {"Discover hidden area", "Uncover secret location", "Map unexplored region"},
	}

	options := descriptions[objType]
	return options[rng.Intn(len(options))]
}

func (m *StoryEventManager) generateObjectiveTarget(objType ObjectiveType, rng *rand.Rand) string {
	targets := []string{"enemy camp", "ancient ruin", "sacred grove", "cursed tomb", "hidden temple", "abandoned fortress"}
	return targets[rng.Intn(len(targets))]
}

func (m *StoryEventManager) generateConsequenceDescription(consType ConsequenceType) string {
	switch consType {
	case ConsequenceLoyaltyChange:
		return "Your companion's loyalty will significantly change"
	case ConsequenceDeparture:
		return "Your companion may leave your service permanently"
	case ConsequenceDeath:
		return "Your companion risks permanent death"
	case ConsequenceRelationshipChange:
		return "Your relationship with your companion will fundamentally change"
	case ConsequenceItemGain:
		return "You will receive a unique reward"
	case ConsequenceSkillUnlock:
		return "Your companion will unlock new abilities"
	default:
		return "Unknown consequence"
	}
}
