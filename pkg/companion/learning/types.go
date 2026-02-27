package learning

import (
	"encoding/json"
	"errors"
	"time"
)

// Sentinel errors for companion learning operations.
var (
	ErrSkillNotFound           = errors.New("skill not found")
	ErrInsufficientSkillPoints = errors.New("insufficient skill points")
	ErrPrerequisiteNotFound    = errors.New("prerequisite not found")
	ErrPrerequisiteNotMet      = errors.New("prerequisite not met")
)

// SkillType represents different companion skill categories.
type SkillType int

const (
	SkillCombat SkillType = iota
	SkillUtility
	SkillSocial
	SkillCrafting
	SkillMagic
	SkillDefense
	SkillHealing
	SkillStealth
)

func (s SkillType) String() string {
	switch s {
	case SkillCombat:
		return "Combat"
	case SkillUtility:
		return "Utility"
	case SkillSocial:
		return "Social"
	case SkillCrafting:
		return "Crafting"
	case SkillMagic:
		return "Magic"
	case SkillDefense:
		return "Defense"
	case SkillHealing:
		return "Healing"
	case SkillStealth:
		return "Stealth"
	default:
		return "Unknown"
	}
}

// Skill represents a learnable companion skill.
type Skill struct {
	Type        SkillType
	Name        string
	Description string
	Level       int     // 0-10
	Experience  float64 // XP towards next level
	MaxLevel    int     // Maximum achievable level
}

// SkillNode represents a node in the skill tree.
type SkillNode struct {
	Skill         *Skill
	Prerequisites []string // Skill names that must be learned first
	Cost          int      // Skill points required
}

// SkillProgression tracks companion skill development.
type SkillProgression struct {
	Skills          map[string]*Skill // Skill name -> Skill
	AvailablePoints int               // Unspent skill points
	TotalXP         float64           // Total experience earned
	SkillTree       map[string]*SkillNode
}

// PersonalityTrait represents a behavioral characteristic.
type PersonalityTrait int

const (
	TraitCautious PersonalityTrait = iota
	TraitBrave
	TraitShy
	TraitOutgoing
	TraitAggressive
	TraitPacifist
	TraitLoyal
	TraitIndependent
	TraitCurious
	TraitPractical
)

func (p PersonalityTrait) String() string {
	switch p {
	case TraitCautious:
		return "Cautious"
	case TraitBrave:
		return "Brave"
	case TraitShy:
		return "Shy"
	case TraitOutgoing:
		return "Outgoing"
	case TraitAggressive:
		return "Aggressive"
	case TraitPacifist:
		return "Pacifist"
	case TraitLoyal:
		return "Loyal"
	case TraitIndependent:
		return "Independent"
	case TraitCurious:
		return "Curious"
	case TraitPractical:
		return "Practical"
	default:
		return "Unknown"
	}
}

// PersonalityEvolution tracks changes in companion personality.
type PersonalityEvolution struct {
	Traits       map[PersonalityTrait]float64 // Trait -> strength (0.0-1.0)
	Changes      []PersonalityChange          // History of trait changes
	MaxChanges   int                          // Maximum changes to store (LRU)
	LastUpdate   time.Time
	timeProvider TimeProvider // injected time source for deterministic timestamps
}

// PersonalityChange records a shift in personality.
type PersonalityChange struct {
	Trait     PersonalityTrait
	OldValue  float64
	NewValue  float64
	Reason    string
	Timestamp time.Time
}

// EventType categorizes memorable events.
type EventType int

const (
	EventCombat EventType = iota
	EventDialog
	EventTrade
	EventQuest
	EventExploration
	EventCrafting
	EventDeath
	EventRevival
	EventGift
	EventBetrayal
)

func (e EventType) String() string {
	switch e {
	case EventCombat:
		return "Combat"
	case EventDialog:
		return "Dialog"
	case EventTrade:
		return "Trade"
	case EventQuest:
		return "Quest"
	case EventExploration:
		return "Exploration"
	case EventCrafting:
		return "Crafting"
	case EventDeath:
		return "Death"
	case EventRevival:
		return "Revival"
	case EventGift:
		return "Gift"
	case EventBetrayal:
		return "Betrayal"
	default:
		return "Unknown"
	}
}

// MemorableEvent represents a significant interaction.
type MemorableEvent struct {
	Type        EventType
	Description string
	Timestamp   time.Time
	Importance  float64 // 0.0-1.0
	PlayerID    string
	Location    string
}

// EventMemory stores companion memories of player interactions.
type EventMemory struct {
	Events       []MemorableEvent
	MaxEvents    int // Maximum events to store (LRU)
	TotalEvents  int // Total events ever recorded
	FirstEventAt time.Time
}

// CompanionLearningComponent is the ECS component for companion AI learning.
type CompanionLearningComponent struct {
	CompanionID  string
	SkillTree    *SkillProgression
	Personality  *PersonalityEvolution
	Memory       *EventMemory
	LearningRate float64              // Multiplier for XP gain (0.5-2.0)
	LastSkillUse map[string]time.Time // Skill name -> last use time
}

// Type returns the component type identifier.
func (c CompanionLearningComponent) Type() string {
	return "companion_learning"
}

// companionLearningData is the serialization format for CompanionLearningComponent.
// Uses JSON for human readability and easy debugging.
type companionLearningData struct {
	CompanionID  string                `json:"companion_id"`
	SkillTree    *skillProgressionData `json:"skill_tree"`
	Personality  *personalityData      `json:"personality"`
	Memory       *eventMemoryData      `json:"memory"`
	LearningRate float64               `json:"learning_rate"`
	LastSkillUse map[string]int64      `json:"last_skill_use"` // Unix timestamps
}

type skillProgressionData struct {
	Skills          map[string]*skillData `json:"skills"`
	AvailablePoints int                   `json:"available_points"`
	TotalXP         float64               `json:"total_xp"`
}

type skillData struct {
	Type          int      `json:"type"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Level         int      `json:"level"`
	Experience    float64  `json:"experience"`
	MaxLevel      int      `json:"max_level"`
	Prerequisites []string `json:"prerequisites,omitempty"` // Skill names that must be learned first
	Cost          int      `json:"cost,omitempty"`          // Skill points required (0 means use default)
}

type personalityData struct {
	Traits     map[int]float64         `json:"traits"`
	Changes    []personalityChangeData `json:"changes"`
	MaxChanges int                     `json:"max_changes"`
	LastUpdate int64                   `json:"last_update"` // Unix timestamp
}

type personalityChangeData struct {
	Trait     int     `json:"trait"`
	OldValue  float64 `json:"old_value"`
	NewValue  float64 `json:"new_value"`
	Reason    string  `json:"reason"`
	Timestamp int64   `json:"timestamp"` // Unix timestamp
}

type eventMemoryData struct {
	Events       []memorableEventData `json:"events"`
	MaxEvents    int                  `json:"max_events"`
	TotalEvents  int                  `json:"total_events"`
	FirstEventAt int64                `json:"first_event_at"` // Unix timestamp
}

type memorableEventData struct {
	Type        int     `json:"type"`
	Description string  `json:"description"`
	Timestamp   int64   `json:"timestamp"` // Unix timestamp
	Importance  float64 `json:"importance"`
	PlayerID    string  `json:"player_id"`
	Location    string  `json:"location"`
}

// Serialize converts the CompanionLearningComponent to JSON bytes for persistence.
// This enables save/load functionality and network synchronization.
func (c *CompanionLearningComponent) Serialize() ([]byte, error) {
	data := companionLearningData{
		CompanionID:  c.CompanionID,
		LearningRate: c.LearningRate,
	}

	// Serialize skill tree
	if c.SkillTree != nil {
		data.SkillTree = &skillProgressionData{
			Skills:          make(map[string]*skillData),
			AvailablePoints: c.SkillTree.AvailablePoints,
			TotalXP:         c.SkillTree.TotalXP,
		}
		for name, skill := range c.SkillTree.Skills {
			sd := &skillData{
				Type:        int(skill.Type),
				Name:        skill.Name,
				Description: skill.Description,
				Level:       skill.Level,
				Experience:  skill.Experience,
				MaxLevel:    skill.MaxLevel,
			}
			// Include skill node data if available
			if node, ok := c.SkillTree.SkillTree[name]; ok {
				sd.Prerequisites = node.Prerequisites
				sd.Cost = node.Cost
			}
			data.SkillTree.Skills[name] = sd
		}
	}

	// Serialize personality
	if c.Personality != nil {
		data.Personality = &personalityData{
			Traits:     make(map[int]float64),
			Changes:    make([]personalityChangeData, 0, len(c.Personality.Changes)),
			MaxChanges: c.Personality.MaxChanges,
			LastUpdate: c.Personality.LastUpdate.Unix(),
		}
		for trait, value := range c.Personality.Traits {
			data.Personality.Traits[int(trait)] = value
		}
		for _, change := range c.Personality.Changes {
			data.Personality.Changes = append(data.Personality.Changes, personalityChangeData{
				Trait:     int(change.Trait),
				OldValue:  change.OldValue,
				NewValue:  change.NewValue,
				Reason:    change.Reason,
				Timestamp: change.Timestamp.Unix(),
			})
		}
	}

	// Serialize memory
	if c.Memory != nil {
		data.Memory = &eventMemoryData{
			Events:       make([]memorableEventData, 0, len(c.Memory.Events)),
			MaxEvents:    c.Memory.MaxEvents,
			TotalEvents:  c.Memory.TotalEvents,
			FirstEventAt: c.Memory.FirstEventAt.Unix(),
		}
		for _, event := range c.Memory.Events {
			data.Memory.Events = append(data.Memory.Events, memorableEventData{
				Type:        int(event.Type),
				Description: event.Description,
				Timestamp:   event.Timestamp.Unix(),
				Importance:  event.Importance,
				PlayerID:    event.PlayerID,
				Location:    event.Location,
			})
		}
	}

	// Serialize LastSkillUse
	if c.LastSkillUse != nil {
		data.LastSkillUse = make(map[string]int64)
		for name, t := range c.LastSkillUse {
			data.LastSkillUse[name] = t.Unix()
		}
	}

	return json.Marshal(data)
}

// Deserialize restores the CompanionLearningComponent from JSON bytes.
// Returns an error if the data is invalid or cannot be parsed.
func (c *CompanionLearningComponent) Deserialize(data []byte) error {
	var d companionLearningData
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	c.CompanionID = d.CompanionID
	c.LearningRate = d.LearningRate

	// Deserialize skill tree
	if d.SkillTree != nil {
		c.SkillTree = &SkillProgression{
			Skills:          make(map[string]*Skill),
			AvailablePoints: d.SkillTree.AvailablePoints,
			TotalXP:         d.SkillTree.TotalXP,
			SkillTree:       make(map[string]*SkillNode),
		}
		for name, sd := range d.SkillTree.Skills {
			skill := &Skill{
				Type:        SkillType(sd.Type),
				Name:        sd.Name,
				Description: sd.Description,
				Level:       sd.Level,
				Experience:  sd.Experience,
				MaxLevel:    sd.MaxLevel,
			}
			c.SkillTree.Skills[name] = skill
			// Restore skill tree nodes with prerequisites and cost from serialized data
			cost := sd.Cost
			if cost == 0 {
				cost = DefaultSkillCost // Use default if not serialized (backward compat)
			}
			c.SkillTree.SkillTree[name] = &SkillNode{
				Skill:         skill,
				Prerequisites: sd.Prerequisites,
				Cost:          cost,
			}
		}
	}

	// Deserialize personality
	if d.Personality != nil {
		c.Personality = &PersonalityEvolution{
			Traits:     make(map[PersonalityTrait]float64),
			Changes:    make([]PersonalityChange, 0, len(d.Personality.Changes)),
			MaxChanges: d.Personality.MaxChanges,
			LastUpdate: time.Unix(d.Personality.LastUpdate, 0),
		}
		for trait, value := range d.Personality.Traits {
			c.Personality.Traits[PersonalityTrait(trait)] = value
		}
		for _, change := range d.Personality.Changes {
			c.Personality.Changes = append(c.Personality.Changes, PersonalityChange{
				Trait:     PersonalityTrait(change.Trait),
				OldValue:  change.OldValue,
				NewValue:  change.NewValue,
				Reason:    change.Reason,
				Timestamp: time.Unix(change.Timestamp, 0),
			})
		}
	}

	// Deserialize memory
	if d.Memory != nil {
		c.Memory = &EventMemory{
			Events:       make([]MemorableEvent, 0, len(d.Memory.Events)),
			MaxEvents:    d.Memory.MaxEvents,
			TotalEvents:  d.Memory.TotalEvents,
			FirstEventAt: time.Unix(d.Memory.FirstEventAt, 0),
		}
		for _, event := range d.Memory.Events {
			c.Memory.Events = append(c.Memory.Events, MemorableEvent{
				Type:        EventType(event.Type),
				Description: event.Description,
				Timestamp:   time.Unix(event.Timestamp, 0),
				Importance:  event.Importance,
				PlayerID:    event.PlayerID,
				Location:    event.Location,
			})
		}
	}

	// Deserialize LastSkillUse
	if d.LastSkillUse != nil {
		c.LastSkillUse = make(map[string]time.Time)
		for name, ts := range d.LastSkillUse {
			c.LastSkillUse[name] = time.Unix(ts, 0)
		}
	}

	return nil
}
