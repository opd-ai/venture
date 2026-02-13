package narrative_world

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// SerializableMemoryEvent represents a MemoryEvent in JSON-serializable form.
type SerializableMemoryEvent struct {
	Timestamp    int64    `json:"timestamp"`
	Type         int      `json:"type"`
	Description  string   `json:"description"`
	Participants []uint64 `json:"participants"`
	Location     string   `json:"location"`
	Importance   float64  `json:"importance"`
}

// SerializableCompanionMemory represents a CompanionMemory in JSON-serializable form.
type SerializableCompanionMemory struct {
	CompanionID uint64                    `json:"companion_id"`
	Events      []SerializableMemoryEvent `json:"events"`
	MaxEvents   int                       `json:"max_events"`
	TotalEvents int                       `json:"total_events"`
}

// SerializableQuestObjective represents a QuestObjective in JSON-serializable form.
type SerializableQuestObjective struct {
	Description string `json:"description"`
	Type        int    `json:"type"`
	Target      string `json:"target"`
	Progress    int    `json:"progress"`
	Required    int    `json:"required"`
	Completed   bool   `json:"completed"`
}

// SerializableConsequence represents a Consequence in JSON-serializable form.
type SerializableConsequence struct {
	Type        int     `json:"type"`
	Description string  `json:"description"`
	Permanent   bool    `json:"permanent"`
	Severity    float64 `json:"severity"`
}

// SerializablePersonalQuest represents a PersonalQuest in JSON-serializable form.
// Note: StoryBranches is excluded as BranchingNarrative is complex and typically regenerated.
type SerializablePersonalQuest struct {
	QuestID       string                       `json:"quest_id"`
	CompanionID   uint64                       `json:"companion_id"`
	CompanionType int                          `json:"companion_type"`
	Title         string                       `json:"title"`
	Description   string                       `json:"description"`
	Objectives    []SerializableQuestObjective `json:"objectives"`
	UnlockLoyalty float64                      `json:"unlock_loyalty"`
	Completed     bool                         `json:"completed"`
	Started       bool                         `json:"started"`
	Consequences  []SerializableConsequence    `json:"consequences"`
	// PersonalityReqs excluded - regenerated from quest templates
}

// SerializableCompanionConflict represents a CompanionConflict in JSON-serializable form.
type SerializableCompanionConflict struct {
	Companion1       uint64  `json:"companion_1"`
	Companion2       uint64  `json:"companion_2"`
	ConflictType     int     `json:"conflict_type"`
	Description      string  `json:"description"`
	Severity         float64 `json:"severity"`
	Active           bool    `json:"active"`
	TimeSinceStartNs int64   `json:"time_since_start_ns"` // Duration as nanoseconds
}

// SerializableCrossCompanionStory represents a CrossCompanionStory in JSON-serializable form.
type SerializableCrossCompanionStory struct {
	StoryID      string                    `json:"story_id"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	Participants []uint64                  `json:"participants"`
	Events       []SerializableMemoryEvent `json:"events"`
	Outcome      int                       `json:"outcome"`
	Active       bool                      `json:"active"`
	// Narrative excluded - regenerated as needed
}

// SerializableStoryEventManager represents the StoryEventManager state for persistence.
type SerializableStoryEventManager struct {
	Seed            int64                                   `json:"seed"`
	Memories        map[uint64]*SerializableCompanionMemory `json:"memories"`
	ActiveQuests    map[uint64][]*SerializablePersonalQuest `json:"active_quests"`
	Conflicts       []SerializableCompanionConflict         `json:"conflicts"`
	CrossStories    []*SerializableCrossCompanionStory      `json:"cross_stories"`
	ConflictChance  float64                                 `json:"conflict_chance"`
	MaxMemoryEvents int                                     `json:"max_memory_events"`
}

// Serialize encodes the StoryEventManager state to JSON for persistence.
func (m *StoryEventManager) Serialize() ([]byte, error) {
	// Convert memories
	memories := make(map[uint64]*SerializableCompanionMemory, len(m.memories))
	for id, mem := range m.memories {
		memories[id] = serializeCompanionMemory(mem)
	}

	// Convert active quests
	activeQuests := make(map[uint64][]*SerializablePersonalQuest, len(m.activeQuests))
	for id, quests := range m.activeQuests {
		serialized := make([]*SerializablePersonalQuest, len(quests))
		for i, q := range quests {
			serialized[i] = serializePersonalQuest(q)
		}
		activeQuests[id] = serialized
	}

	// Convert conflicts
	conflicts := make([]SerializableCompanionConflict, len(m.conflicts))
	for i, c := range m.conflicts {
		conflicts[i] = serializeCompanionConflict(c)
	}

	// Convert cross stories
	crossStories := make([]*SerializableCrossCompanionStory, len(m.crossStories))
	for i, s := range m.crossStories {
		crossStories[i] = serializeCrossCompanionStory(s)
	}

	state := SerializableStoryEventManager{
		Seed:            m.seed,
		Memories:        memories,
		ActiveQuests:    activeQuests,
		Conflicts:       conflicts,
		CrossStories:    crossStories,
		ConflictChance:  m.conflictChance,
		MaxMemoryEvents: m.maxMemoryEvents,
	}

	return json.Marshal(state)
}

// Deserialize decodes the StoryEventManager state from JSON persistence data.
func (m *StoryEventManager) Deserialize(data []byte) error {
	var state SerializableStoryEventManager
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal StoryEventManager: %w", err)
	}

	// Restore seed and settings
	m.seed = state.Seed
	m.conflictChance = state.ConflictChance
	m.maxMemoryEvents = state.MaxMemoryEvents

	// Restore memories
	m.memories = make(map[uint64]*CompanionMemory, len(state.Memories))
	for id, mem := range state.Memories {
		m.memories[id] = deserializeCompanionMemory(mem)
	}

	// Restore active quests
	m.activeQuests = make(map[uint64][]*PersonalQuest, len(state.ActiveQuests))
	for id, quests := range state.ActiveQuests {
		restored := make([]*PersonalQuest, len(quests))
		for i, q := range quests {
			restored[i] = deserializePersonalQuest(q)
		}
		m.activeQuests[id] = restored
	}

	// Restore conflicts
	m.conflicts = make([]CompanionConflict, len(state.Conflicts))
	for i, c := range state.Conflicts {
		m.conflicts[i] = deserializeCompanionConflict(c)
	}

	// Restore cross stories
	m.crossStories = make([]*CrossCompanionStory, len(state.CrossStories))
	for i, s := range state.CrossStories {
		m.crossStories[i] = deserializeCrossCompanionStory(s)
	}

	return nil
}

// Helper functions for serialization

func serializeMemoryEvent(e MemoryEvent) SerializableMemoryEvent {
	return SerializableMemoryEvent{
		Timestamp:    e.Timestamp,
		Type:         int(e.Type),
		Description:  e.Description,
		Participants: e.Participants,
		Location:     e.Location,
		Importance:   e.Importance,
	}
}

func deserializeMemoryEvent(e SerializableMemoryEvent) MemoryEvent {
	return MemoryEvent{
		Timestamp:    e.Timestamp,
		Type:         EventType(e.Type),
		Description:  e.Description,
		Participants: e.Participants,
		Location:     e.Location,
		Importance:   e.Importance,
	}
}

func serializeCompanionMemory(m *CompanionMemory) *SerializableCompanionMemory {
	events := make([]SerializableMemoryEvent, len(m.Events))
	for i, e := range m.Events {
		events[i] = serializeMemoryEvent(e)
	}
	return &SerializableCompanionMemory{
		CompanionID: m.CompanionID,
		Events:      events,
		MaxEvents:   m.MaxEvents,
		TotalEvents: m.TotalEvents,
	}
}

func deserializeCompanionMemory(m *SerializableCompanionMemory) *CompanionMemory {
	events := make([]MemoryEvent, len(m.Events))
	for i, e := range m.Events {
		events[i] = deserializeMemoryEvent(e)
	}
	return &CompanionMemory{
		CompanionID: m.CompanionID,
		Events:      events,
		MaxEvents:   m.MaxEvents,
		TotalEvents: m.TotalEvents,
	}
}

func serializeQuestObjective(o QuestObjective) SerializableQuestObjective {
	return SerializableQuestObjective{
		Description: o.Description,
		Type:        int(o.Type),
		Target:      o.Target,
		Progress:    o.Progress,
		Required:    o.Required,
		Completed:   o.Completed,
	}
}

func deserializeQuestObjective(o SerializableQuestObjective) QuestObjective {
	return QuestObjective{
		Description: o.Description,
		Type:        ObjectiveType(o.Type),
		Target:      o.Target,
		Progress:    o.Progress,
		Required:    o.Required,
		Completed:   o.Completed,
	}
}

func serializeConsequence(c Consequence) SerializableConsequence {
	return SerializableConsequence{
		Type:        int(c.Type),
		Description: c.Description,
		Permanent:   c.Permanent,
		Severity:    c.Severity,
	}
}

func deserializeConsequence(c SerializableConsequence) Consequence {
	return Consequence{
		Type:        ConsequenceType(c.Type),
		Description: c.Description,
		Permanent:   c.Permanent,
		Severity:    c.Severity,
	}
}

func serializePersonalQuest(q *PersonalQuest) *SerializablePersonalQuest {
	objectives := make([]SerializableQuestObjective, len(q.Objectives))
	for i, o := range q.Objectives {
		objectives[i] = serializeQuestObjective(o)
	}
	consequences := make([]SerializableConsequence, len(q.Consequences))
	for i, c := range q.Consequences {
		consequences[i] = serializeConsequence(c)
	}
	return &SerializablePersonalQuest{
		QuestID:       q.QuestID,
		CompanionID:   q.CompanionID,
		CompanionType: int(q.CompanionType),
		Title:         q.Title,
		Description:   q.Description,
		Objectives:    objectives,
		UnlockLoyalty: q.UnlockLoyalty,
		Completed:     q.Completed,
		Started:       q.Started,
		Consequences:  consequences,
	}
}

func deserializePersonalQuest(q *SerializablePersonalQuest) *PersonalQuest {
	objectives := make([]QuestObjective, len(q.Objectives))
	for i, o := range q.Objectives {
		objectives[i] = deserializeQuestObjective(o)
	}
	consequences := make([]Consequence, len(q.Consequences))
	for i, c := range q.Consequences {
		consequences[i] = deserializeConsequence(c)
	}
	return &PersonalQuest{
		QuestID:       q.QuestID,
		CompanionID:   q.CompanionID,
		CompanionType: engine.CompanionType(q.CompanionType),
		Title:         q.Title,
		Description:   q.Description,
		Objectives:    objectives,
		UnlockLoyalty: q.UnlockLoyalty,
		Completed:     q.Completed,
		Started:       q.Started,
		Consequences:  consequences,
		// StoryBranches and PersonalityReqs would need regeneration
	}
}

func serializeCompanionConflict(c CompanionConflict) SerializableCompanionConflict {
	return SerializableCompanionConflict{
		Companion1:       c.Companion1,
		Companion2:       c.Companion2,
		ConflictType:     int(c.ConflictType),
		Description:      c.Description,
		Severity:         c.Severity,
		Active:           c.Active,
		TimeSinceStartNs: c.TimeSinceStart.Nanoseconds(),
	}
}

func deserializeCompanionConflict(c SerializableCompanionConflict) CompanionConflict {
	return CompanionConflict{
		Companion1:     c.Companion1,
		Companion2:     c.Companion2,
		ConflictType:   ConflictType(c.ConflictType),
		Description:    c.Description,
		Severity:       c.Severity,
		Active:         c.Active,
		TimeSinceStart: time.Duration(c.TimeSinceStartNs),
		// ResolutionQuest would need regeneration if needed
	}
}

func serializeCrossCompanionStory(s *CrossCompanionStory) *SerializableCrossCompanionStory {
	events := make([]SerializableMemoryEvent, len(s.Events))
	for i, e := range s.Events {
		events[i] = serializeMemoryEvent(e)
	}
	return &SerializableCrossCompanionStory{
		StoryID:      s.StoryID,
		Title:        s.Title,
		Description:  s.Description,
		Participants: s.Participants,
		Events:       events,
		Outcome:      int(s.Outcome),
		Active:       s.Active,
	}
}

func deserializeCrossCompanionStory(s *SerializableCrossCompanionStory) *CrossCompanionStory {
	events := make([]MemoryEvent, len(s.Events))
	for i, e := range s.Events {
		events[i] = deserializeMemoryEvent(e)
	}
	return &CrossCompanionStory{
		StoryID:      s.StoryID,
		Title:        s.Title,
		Description:  s.Description,
		Participants: s.Participants,
		Events:       events,
		Outcome:      StoryOutcome(s.Outcome),
		Active:       s.Active,
		// Narrative would need regeneration if needed
	}
}
