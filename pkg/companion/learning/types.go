package learning

import (
	"time"
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
	Traits     map[PersonalityTrait]float64 // Trait -> strength (0.0-1.0)
	Changes    []PersonalityChange          // History of trait changes
	LastUpdate time.Time
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
