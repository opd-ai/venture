package learning

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// defaultLogger is used when no custom logger is provided.
var defaultLogger = logrus.New()

func init() {
	defaultLogger.SetReportCaller(true)
	defaultLogger.WithField("package", "companion_learning").Debug("Companion learning package initialized")
}

// Manager handles companion learning operations.
// Thread-safe for concurrent access from multiple goroutines.
type Manager struct {
	mu           sync.RWMutex
	companions   map[string]*CompanionLearningComponent
	timeProvider TimeProvider
	logger       *logrus.Logger
}

// NewManager creates a new companion learning manager.
// Uses real wall-clock time and the default package logger.
// For deterministic behavior or custom logging, use NewManagerWithOptions.
func NewManager() *Manager {
	return NewManagerWithOptions(DefaultTimeProvider(), nil)
}

// NewManagerWithTimeProvider creates a manager with custom time source.
// This enables deterministic testing and reproducible state.
// Uses the default package logger. For custom logging, use NewManagerWithOptions.
func NewManagerWithTimeProvider(timeProvider TimeProvider) *Manager {
	return NewManagerWithOptions(timeProvider, nil)
}

// NewManagerWithOptions creates a manager with custom time source and logger.
// Pass nil for logger to use the default package logger.
// This enables integration with engine-level structured logging.
func NewManagerWithOptions(timeProvider TimeProvider, logger *logrus.Logger) *Manager {
	if logger == nil {
		logger = defaultLogger
	}
	logger.Debug("Creating new companion learning manager")

	m := &Manager{
		companions:   make(map[string]*CompanionLearningComponent),
		timeProvider: timeProvider,
		logger:       logger,
	}

	logger.WithFields(logrus.Fields{
		"initial_companion_count": 0,
	}).Info("Companion learning manager created")

	return m
}

// AddCompanion registers a companion for learning tracking.
//
// Parameters:
//   - companionID: Unique identifier for the companion
//   - learningRate: Rate at which the companion learns (multiplier for XP gain)
//
// Learning Rate Behavior:
// This function implements "fail-soft" error handling for invalid learning rates.
// If learningRate <= 0, it automatically clamps the value to 1.0 (default) and logs
// a warning. This prevents crashes from invalid input while ensuring companions can
// always learn at a reasonable rate.
//
// Design Rationale:
// The fail-soft approach prioritizes system stability over strict validation. In a
// game context, it's better to use a sensible default than to reject companion creation
// entirely. This allows the game to continue functioning even if a bug or user error
// provides an invalid learning rate.
//
// Expected Learning Rate Range: 0.5 (slow learner) to 2.0 (fast learner)
// Default: 1.0 (normal learning speed)
//
// Thread Safety: Safe for concurrent calls.
func (m *Manager) AddCompanion(companionID string, learningRate float64) *CompanionLearningComponent {
	m.logger.WithFields(logrus.Fields{
		"companion_id":  companionID,
		"learning_rate": learningRate,
	}).Debug("Adding companion to learning manager")

	if learningRate <= 0 {
		m.logger.WithFields(logrus.Fields{
			"companion_id":          companionID,
			"invalid_learning_rate": learningRate,
			"default_learning_rate": DefaultLearningRate,
		}).Warn("Invalid learning rate provided, using default")
		learningRate = DefaultLearningRate
	}

	comp := &CompanionLearningComponent{
		CompanionID:  companionID,
		SkillTree:    NewSkillProgression(),
		Personality:  NewPersonalityEvolutionWithTimeProvider(m.timeProvider),
		Memory:       NewEventMemoryWithTimeProvider(DefaultMaxEvents, m.timeProvider),
		LearningRate: learningRate,
		LastSkillUse: make(map[string]time.Time),
	}

	m.mu.Lock()
	m.companions[companionID] = comp
	totalCompanions := len(m.companions)
	m.mu.Unlock()

	m.logger.WithFields(logrus.Fields{
		"companion_id":     companionID,
		"learning_rate":    learningRate,
		"total_companions": totalCompanions,
	}).Info("Companion added to learning manager")

	return comp
}

// GetCompanion retrieves a companion's learning component.
func (m *Manager) GetCompanion(companionID string) (*CompanionLearningComponent, bool) {
	m.logger.WithFields(logrus.Fields{
		"companion_id": companionID,
	}).Debug("Retrieving companion from learning manager")

	m.mu.RLock()
	comp, ok := m.companions[companionID]
	m.mu.RUnlock()

	if !ok {
		m.logger.WithFields(logrus.Fields{
			"companion_id": companionID,
		}).Debug("Companion not found in learning manager")
	} else {
		m.logger.WithFields(logrus.Fields{
			"companion_id": companionID,
		}).Debug("Companion retrieved successfully")
	}

	return comp, ok
}

// RemoveCompanion removes a companion from tracking.
func (m *Manager) RemoveCompanion(companionID string) {
	m.logger.WithFields(logrus.Fields{
		"companion_id": companionID,
	}).Debug("Removing companion from learning manager")

	m.mu.Lock()
	_, existed := m.companions[companionID]
	delete(m.companions, companionID)
	remainingCompanions := len(m.companions)
	m.mu.Unlock()

	if existed {
		m.logger.WithFields(logrus.Fields{
			"companion_id":         companionID,
			"remaining_companions": remainingCompanions,
		}).Info("Companion removed from learning manager")
	} else {
		m.logger.WithFields(logrus.Fields{
			"companion_id": companionID,
		}).Warn("Attempted to remove non-existent companion")
	}
}

// NewSkillProgression creates a new skill progression system.
func NewSkillProgression() *SkillProgression {
	defaultLogger.Debug("Creating new skill progression system")

	sp := &SkillProgression{
		Skills:          make(map[string]*Skill),
		AvailablePoints: 0,
		TotalXP:         0,
		SkillTree:       make(map[string]*SkillNode),
	}

	sp.initializeSkillTree()

	defaultLogger.WithFields(logrus.Fields{
		"total_skills":     len(sp.Skills),
		"available_points": sp.AvailablePoints,
		"total_xp":         sp.TotalXP,
	}).Info("Skill progression system created")

	return sp
}

// initializeSkillTree sets up the default skill tree.
func (sp *SkillProgression) initializeSkillTree() {
	defaultLogger.Debug("Initializing skill tree with default skills")

	skillsAdded := 0

	// Combat skills
	sp.addSkillNode("Basic Attack", SkillCombat, "Improved melee damage", nil, 1, 10)
	sp.addSkillNode("Power Strike", SkillCombat, "Devastating attack", []string{"Basic Attack"}, 1, 10)
	sp.addSkillNode("Combat Mastery", SkillCombat, "Enhanced combat effectiveness", []string{"Power Strike"}, 2, 10)
	skillsAdded += 3

	// Defense skills
	sp.addSkillNode("Block", SkillDefense, "Reduce incoming damage", nil, 1, 10)
	sp.addSkillNode("Iron Skin", SkillDefense, "Increased armor", []string{"Block"}, 1, 10)
	sp.addSkillNode("Defensive Stance", SkillDefense, "Maximum protection", []string{"Iron Skin"}, 2, 10)
	skillsAdded += 3

	// Utility skills
	sp.addSkillNode("Gather", SkillUtility, "Collect resources faster", nil, 1, 10)
	sp.addSkillNode("Scout", SkillUtility, "Reveal hidden paths", []string{"Gather"}, 1, 10)
	sp.addSkillNode("Tracking", SkillUtility, "Follow enemy trails", []string{"Scout"}, 2, 10)
	skillsAdded += 3

	// Social skills
	sp.addSkillNode("Persuasion", SkillSocial, "Convince NPCs", nil, 1, 10)
	sp.addSkillNode("Charm", SkillSocial, "Better trade prices", []string{"Persuasion"}, 1, 10)
	sp.addSkillNode("Leadership", SkillSocial, "Inspire allies", []string{"Charm"}, 2, 10)
	skillsAdded += 3

	// Healing skills
	sp.addSkillNode("First Aid", SkillHealing, "Basic healing", nil, 1, 10)
	sp.addSkillNode("Restoration", SkillHealing, "Powerful healing", []string{"First Aid"}, 1, 10)
	sp.addSkillNode("Regeneration", SkillHealing, "Continuous healing", []string{"Restoration"}, 2, 10)
	skillsAdded += 3

	// Magic skills
	sp.addSkillNode("Mana Control", SkillMagic, "Increased mana pool", nil, 1, 10)
	sp.addSkillNode("Spell Power", SkillMagic, "Stronger spells", []string{"Mana Control"}, 1, 10)
	sp.addSkillNode("Arcane Mastery", SkillMagic, "Ultimate magic power", []string{"Spell Power"}, 2, 10)
	skillsAdded += 3

	// Crafting skills
	sp.addSkillNode("Apprentice Smith", SkillCrafting, "Basic crafting", nil, 1, 10)
	sp.addSkillNode("Master Smith", SkillCrafting, "Advanced crafting", []string{"Apprentice Smith"}, 1, 10)
	sp.addSkillNode("Legendary Smith", SkillCrafting, "Epic items", []string{"Master Smith"}, 2, 10)
	skillsAdded += 3

	// Stealth skills
	sp.addSkillNode("Sneak", SkillStealth, "Move undetected", nil, 1, 10)
	sp.addSkillNode("Backstab", SkillStealth, "Critical strikes", []string{"Sneak"}, 1, 10)
	sp.addSkillNode("Shadow Walk", SkillStealth, "Become invisible", []string{"Backstab"}, 2, 10)
	skillsAdded += 3

	defaultLogger.WithFields(logrus.Fields{
		"skills_added": skillsAdded,
		"total_skills": len(sp.Skills),
	}).Info("Skill tree initialized")
}

// addSkillNode adds a skill to the tree.
func (sp *SkillProgression) addSkillNode(name string, skillType SkillType, description string, prerequisites []string, cost, maxLevel int) {
	defaultLogger.WithFields(logrus.Fields{
		"skill_name":         name,
		"skill_type":         skillType.String(),
		"prerequisite_count": len(prerequisites),
		"cost":               cost,
		"max_level":          maxLevel,
	}).Debug("Adding skill node to tree")

	skill := &Skill{
		Type:        skillType,
		Name:        name,
		Description: description,
		Level:       0,
		Experience:  0,
		MaxLevel:    maxLevel,
	}

	node := &SkillNode{
		Skill:         skill,
		Prerequisites: prerequisites,
		Cost:          cost,
	}

	sp.Skills[name] = skill
	sp.SkillTree[name] = node
}

// AddExperience adds XP to a specific skill.
func (sp *SkillProgression) AddExperience(skillName string, xp, learningRate float64) error {
	defaultLogger.WithFields(logrus.Fields{
		"skill_name":    skillName,
		"xp":            xp,
		"learning_rate": learningRate,
	}).Debug("Adding experience to skill")

	skill, ok := sp.Skills[skillName]
	if !ok {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name": skillName,
		}).Warn("Skill not found when adding experience")
		return fmt.Errorf("%w: %s", ErrSkillNotFound, skillName)
	}

	adjustedXP := xp * learningRate
	oldTotalXP := sp.TotalXP
	sp.TotalXP += adjustedXP

	if skill.Level >= skill.MaxLevel {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name":  skillName,
			"skill_level": skill.Level,
			"max_level":   skill.MaxLevel,
			"xp_added":    adjustedXP,
		}).Debug("Skill at max level, XP counted towards total only")
		return nil
	}

	oldLevel := skill.Level
	oldExperience := skill.Experience
	skill.Experience += adjustedXP

	levelsGained := 0
	xpNeeded := float64(skill.Level+1) * SkillXPPerLevel
	for skill.Experience >= xpNeeded && skill.Level < skill.MaxLevel {
		skill.Experience -= xpNeeded
		skill.Level++
		sp.AvailablePoints++
		levelsGained++
		xpNeeded = float64(skill.Level+1) * SkillXPPerLevel
	}

	if levelsGained > 0 {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name":       skillName,
			"old_level":        oldLevel,
			"new_level":        skill.Level,
			"levels_gained":    levelsGained,
			"available_points": sp.AvailablePoints,
		}).Info("Skill leveled up")
	} else {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name":   skillName,
			"skill_level":  skill.Level,
			"old_xp":       oldExperience,
			"new_xp":       skill.Experience,
			"xp_added":     adjustedXP,
			"old_total_xp": oldTotalXP,
			"new_total_xp": sp.TotalXP,
		}).Debug("Experience added to skill")
	}

	return nil
}

// CanLearnSkill checks if prerequisites are met.
func (sp *SkillProgression) CanLearnSkill(skillName string) (bool, error) {
	defaultLogger.WithFields(logrus.Fields{
		"skill_name": skillName,
	}).Debug("Checking if skill can be learned")

	node, ok := sp.SkillTree[skillName]
	if !ok {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name": skillName,
		}).Warn("Skill not found in tree")
		return false, fmt.Errorf("%w: %s", ErrSkillNotFound, skillName)
	}

	if sp.AvailablePoints < node.Cost {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name":       skillName,
			"available_points": sp.AvailablePoints,
			"cost":             node.Cost,
		}).Debug("Insufficient skill points")
		return false, fmt.Errorf("%w: have %d, need %d", ErrInsufficientSkillPoints, sp.AvailablePoints, node.Cost)
	}

	for _, prereq := range node.Prerequisites {
		prereqSkill, ok := sp.Skills[prereq]
		if !ok {
			defaultLogger.WithFields(logrus.Fields{
				"skill_name":   skillName,
				"prerequisite": prereq,
			}).Error("Prerequisite skill not found")
			return false, fmt.Errorf("%w: %s", ErrPrerequisiteNotFound, prereq)
		}
		if prereqSkill.Level < 1 {
			defaultLogger.WithFields(logrus.Fields{
				"skill_name":         skillName,
				"prerequisite":       prereq,
				"prerequisite_level": prereqSkill.Level,
			}).Debug("Prerequisite not met")
			return false, fmt.Errorf("%w: %s", ErrPrerequisiteNotMet, prereq)
		}
	}

	defaultLogger.WithFields(logrus.Fields{
		"skill_name":       skillName,
		"available_points": sp.AvailablePoints,
		"cost":             node.Cost,
	}).Debug("Skill can be learned")
	return true, nil
}

// LearnSkill allocates points to a skill and increments its level.
func (sp *SkillProgression) LearnSkill(skillName string) error {
	defaultLogger.WithFields(logrus.Fields{
		"skill_name": skillName,
	}).Debug("Attempting to learn skill")

	canLearn, err := sp.CanLearnSkill(skillName)
	if !canLearn {
		defaultLogger.WithFields(logrus.Fields{
			"skill_name": skillName,
			"error":      err.Error(),
		}).Warn("Cannot learn skill")
		return err
	}

	node := sp.SkillTree[skillName]
	oldPoints := sp.AvailablePoints
	sp.AvailablePoints -= node.Cost

	// Increment skill level so it can satisfy prerequisites
	node.Skill.Level++

	defaultLogger.WithFields(logrus.Fields{
		"skill_name":       skillName,
		"cost":             node.Cost,
		"old_points":       oldPoints,
		"remaining_points": sp.AvailablePoints,
		"new_level":        node.Skill.Level,
	}).Info("Skill learned successfully")

	return nil
}

// NewPersonalityEvolution creates a new personality system.
// Uses real wall-clock time. For deterministic behavior, use NewPersonalityEvolutionWithTimeProvider.
func NewPersonalityEvolution() *PersonalityEvolution {
	return NewPersonalityEvolutionWithTimeProvider(DefaultTimeProvider())
}

// NewPersonalityEvolutionWithTimeProvider creates a personality system with custom time source.
func NewPersonalityEvolutionWithTimeProvider(timeProvider TimeProvider) *PersonalityEvolution {
	defaultLogger.Debug("Creating new personality evolution system")

	pe := &PersonalityEvolution{
		Traits:       make(map[PersonalityTrait]float64),
		Changes:      []PersonalityChange{},
		MaxChanges:   DefaultMaxPersonalityChanges,
		LastUpdate:   timeProvider.Now(),
		timeProvider: timeProvider,
	}

	// Initialize traits with neutral values
	pe.Traits[TraitCautious] = TraitDefaultValue
	pe.Traits[TraitBrave] = TraitDefaultValue
	pe.Traits[TraitShy] = TraitDefaultValue
	pe.Traits[TraitOutgoing] = TraitDefaultValue
	pe.Traits[TraitAggressive] = TraitDefaultValue
	pe.Traits[TraitPacifist] = TraitDefaultValue
	pe.Traits[TraitLoyal] = TraitDefaultValue
	pe.Traits[TraitIndependent] = TraitDefaultValue
	pe.Traits[TraitCurious] = TraitDefaultValue
	pe.Traits[TraitPractical] = TraitDefaultValue

	defaultLogger.WithFields(logrus.Fields{
		"trait_count":   len(pe.Traits),
		"default_value": TraitDefaultValue,
	}).Info("Personality evolution system created")

	return pe
}

// AdjustTrait modifies a personality trait.
// Uses the injected time provider for deterministic timestamps.
func (pe *PersonalityEvolution) AdjustTrait(trait PersonalityTrait, delta float64, reason string) {
	defaultLogger.WithFields(logrus.Fields{
		"trait":  trait.String(),
		"delta":  delta,
		"reason": reason,
	}).Debug("Adjusting personality trait")

	oldValue := pe.Traits[trait]
	newValue := oldValue + delta

	// Clamp to [TraitMinValue, TraitMaxValue]
	if newValue < TraitMinValue {
		newValue = TraitMinValue
		defaultLogger.WithFields(logrus.Fields{
			"trait":      trait.String(),
			"attempted":  oldValue + delta,
			"clamped_to": TraitMinValue,
		}).Debug("Trait value clamped to minimum")
	}
	if newValue > TraitMaxValue {
		newValue = TraitMaxValue
		defaultLogger.WithFields(logrus.Fields{
			"trait":      trait.String(),
			"attempted":  oldValue + delta,
			"clamped_to": TraitMaxValue,
		}).Debug("Trait value clamped to maximum")
	}

	// Use time provider for deterministic timestamps
	tp := pe.timeProvider
	if tp == nil {
		tp = DefaultTimeProvider()
	}
	now := tp.Now()

	pe.Traits[trait] = newValue
	pe.LastUpdate = now

	change := PersonalityChange{
		Trait:     trait,
		OldValue:  oldValue,
		NewValue:  newValue,
		Reason:    reason,
		Timestamp: now,
	}
	pe.Changes = append(pe.Changes, change)

	// LRU eviction to prevent unbounded memory growth
	if pe.MaxChanges > 0 && len(pe.Changes) > pe.MaxChanges {
		pe.Changes = pe.Changes[len(pe.Changes)-pe.MaxChanges:]
	}

	defaultLogger.WithFields(logrus.Fields{
		"trait":         trait.String(),
		"old_value":     oldValue,
		"new_value":     newValue,
		"delta":         delta,
		"reason":        reason,
		"total_changes": len(pe.Changes),
	}).Info("Personality trait adjusted")
}

// GetDominantTrait returns the strongest personality trait.
// On ties, returns the trait with lowest enum value for determinism.
func (pe *PersonalityEvolution) GetDominantTrait() PersonalityTrait {
	defaultLogger.Debug("Determining dominant personality trait")

	// Use deterministic tie-breaking by iterating in enum order
	allTraits := []PersonalityTrait{
		TraitCautious, TraitBrave, TraitShy, TraitOutgoing,
		TraitAggressive, TraitPacifist, TraitLoyal, TraitIndependent,
		TraitCurious, TraitPractical,
	}

	var dominant PersonalityTrait
	maxValue := 0.0

	for _, trait := range allTraits {
		if value, exists := pe.Traits[trait]; exists && value > maxValue {
			maxValue = value
			dominant = trait
		}
	}

	defaultLogger.WithFields(logrus.Fields{
		"dominant_trait": dominant.String(),
		"trait_value":    maxValue,
	}).Debug("Dominant trait determined")

	return dominant
}

// NewEventMemory creates a new event memory system.
// Uses real wall-clock time. For deterministic behavior, use NewEventMemoryWithTimeProvider.
func NewEventMemory(maxEvents int) *EventMemory {
	return NewEventMemoryWithTimeProvider(maxEvents, DefaultTimeProvider())
}

// NewEventMemoryWithTimeProvider creates an event memory system with custom time source.
func NewEventMemoryWithTimeProvider(maxEvents int, timeProvider TimeProvider) *EventMemory {
	defaultLogger.WithFields(logrus.Fields{
		"max_events": maxEvents,
	}).Debug("Creating new event memory system")

	em := &EventMemory{
		Events:       []MemorableEvent{},
		MaxEvents:    maxEvents,
		TotalEvents:  0,
		FirstEventAt: timeProvider.Now(),
	}

	defaultLogger.WithFields(logrus.Fields{
		"max_events": maxEvents,
	}).Info("Event memory system created")

	return em
}

// AddEvent records a memorable event.
func (em *EventMemory) AddEvent(event MemorableEvent) {
	defaultLogger.WithFields(logrus.Fields{
		"event_type":  event.Type.String(),
		"description": event.Description,
		"importance":  event.Importance,
		"player_id":   event.PlayerID,
	}).Debug("Adding event to memory")

	em.Events = append(em.Events, event)
	oldTotalEvents := em.TotalEvents
	em.TotalEvents++

	if em.TotalEvents == 1 {
		em.FirstEventAt = event.Timestamp
		defaultLogger.WithFields(logrus.Fields{
			"first_event_time": em.FirstEventAt,
		}).Debug("First event recorded")
	}

	evicted := false
	if len(em.Events) > em.MaxEvents {
		em.Events = em.Events[len(em.Events)-em.MaxEvents:]
		evicted = true
	}

	defaultLogger.WithFields(logrus.Fields{
		"event_type":    event.Type.String(),
		"old_total":     oldTotalEvents,
		"new_total":     em.TotalEvents,
		"stored_events": len(em.Events),
		"evicted":       evicted,
	}).Debug("Event added to memory")
}

// GetRecentEvents returns the N most recent events.
func (em *EventMemory) GetRecentEvents(n int) []MemorableEvent {
	defaultLogger.WithFields(logrus.Fields{
		"requested_count": n,
		"total_events":    len(em.Events),
	}).Debug("Retrieving recent events")

	if n > len(em.Events) {
		n = len(em.Events)
	}

	events := em.Events[len(em.Events)-n:]

	defaultLogger.WithFields(logrus.Fields{
		"requested_count": n,
		"returned_count":  len(events),
	}).Debug("Recent events retrieved")

	return events
}

// GetEventsByType filters events by type.
func (em *EventMemory) GetEventsByType(eventType EventType) []MemorableEvent {
	defaultLogger.WithFields(logrus.Fields{
		"event_type":   eventType.String(),
		"total_events": len(em.Events),
	}).Debug("Filtering events by type")

	var filtered []MemorableEvent
	for _, event := range em.Events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	defaultLogger.WithFields(logrus.Fields{
		"event_type":     eventType.String(),
		"filtered_count": len(filtered),
		"total_events":   len(em.Events),
	}).Debug("Events filtered by type")

	return filtered
}

// ProcessCombatAction updates skills and personality based on combat behavior.
// Uses the companion's internal time provider for deterministic timestamps.
// Safe to call with nil comp (returns immediately).
func ProcessCombatAction(comp *CompanionLearningComponent, aggressive, successful bool) {
	if comp == nil {
		return
	}
	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"aggressive":   aggressive,
		"successful":   successful,
	}).Debug("Processing combat action")

	xp := 10.0
	if successful {
		xp = 20.0
	}

	if aggressive {
		err := comp.SkillTree.AddExperience("Basic Attack", xp, comp.LearningRate)
		if err != nil {
			defaultLogger.WithFields(logrus.Fields{
				"companion_id": comp.CompanionID,
				"skill_name":   "Basic Attack",
				"error":        err.Error(),
			}).Warn("Failed to add combat XP")
		}
		comp.Personality.AdjustTrait(TraitAggressive, 0.01, "engaged in aggressive combat")
		comp.Personality.AdjustTrait(TraitPacifist, -0.01, "engaged in aggressive combat")
	} else {
		err := comp.SkillTree.AddExperience("Block", xp, comp.LearningRate)
		if err != nil {
			defaultLogger.WithFields(logrus.Fields{
				"companion_id": comp.CompanionID,
				"skill_name":   "Block",
				"error":        err.Error(),
			}).Warn("Failed to add defense XP")
		}
		comp.Personality.AdjustTrait(TraitCautious, 0.01, "used defensive tactics")
	}

	// Use personality's time provider for deterministic timestamps
	tp := getTimeProviderFromPersonality(comp.Personality)
	comp.Memory.AddEvent(MemorableEvent{
		Type:        EventCombat,
		Description: fmt.Sprintf("Combat action (aggressive=%v, successful=%v)", aggressive, successful),
		Timestamp:   tp.Now(),
		Importance:  0.6,
	})

	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"aggressive":   aggressive,
		"successful":   successful,
		"xp_awarded":   xp,
	}).Info("Combat action processed")
}

// getTimeProviderFromPersonality retrieves the time provider from a PersonalityEvolution.
// Falls back to real time if no time provider is set.
func getTimeProviderFromPersonality(pe *PersonalityEvolution) TimeProvider {
	if pe != nil && pe.timeProvider != nil {
		return pe.timeProvider
	}
	return DefaultTimeProvider()
}

// ProcessSocialInteraction updates skills and personality based on social behavior.
// Uses the companion's internal time provider for deterministic timestamps.
// Safe to call with nil comp (returns immediately).
func ProcessSocialInteraction(comp *CompanionLearningComponent, playerID string, positive bool) {
	if comp == nil {
		return
	}
	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"player_id":    playerID,
		"positive":     positive,
	}).Debug("Processing social interaction")

	xp := 15.0
	if positive {
		err := comp.SkillTree.AddExperience("Persuasion", xp, comp.LearningRate)
		if err != nil {
			defaultLogger.WithFields(logrus.Fields{
				"companion_id": comp.CompanionID,
				"skill_name":   "Persuasion",
				"error":        err.Error(),
			}).Warn("Failed to add persuasion XP")
		}
		comp.Personality.AdjustTrait(TraitOutgoing, 0.02, "positive social interaction")
		comp.Personality.AdjustTrait(TraitShy, -0.02, "positive social interaction")
		comp.Personality.AdjustTrait(TraitLoyal, 0.01, "bonded with player")
	} else {
		comp.Personality.AdjustTrait(TraitShy, 0.01, "negative social interaction")
		comp.Personality.AdjustTrait(TraitOutgoing, -0.01, "negative social interaction")
	}

	// Use personality's time provider for deterministic timestamps
	tp := getTimeProviderFromPersonality(comp.Personality)
	comp.Memory.AddEvent(MemorableEvent{
		Type:        EventDialog,
		Description: fmt.Sprintf("Social interaction with %s (positive=%v)", playerID, positive),
		Timestamp:   tp.Now(),
		Importance:  0.5,
		PlayerID:    playerID,
	})

	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"player_id":    playerID,
		"positive":     positive,
		"xp_awarded":   xp,
	}).Info("Social interaction processed")
}

// ProcessExploration updates skills and personality based on exploration behavior.
// Uses the companion's internal time provider for deterministic timestamps.
// Safe to call with nil comp (returns immediately).
func ProcessExploration(comp *CompanionLearningComponent, discovered bool) {
	if comp == nil {
		return
	}
	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"discovered":   discovered,
	}).Debug("Processing exploration action")

	xp := 8.0
	if discovered {
		xp = 12.0
	}

	err := comp.SkillTree.AddExperience("Scout", xp, comp.LearningRate)
	if err != nil {
		defaultLogger.WithFields(logrus.Fields{
			"companion_id": comp.CompanionID,
			"skill_name":   "Scout",
			"error":        err.Error(),
		}).Warn("Failed to add scout XP")
	}
	comp.Personality.AdjustTrait(TraitCurious, 0.015, "explored new area")
	comp.Personality.AdjustTrait(TraitPractical, -0.005, "took exploratory risk")

	// Use personality's time provider for deterministic timestamps
	tp := getTimeProviderFromPersonality(comp.Personality)
	comp.Memory.AddEvent(MemorableEvent{
		Type:        EventExploration,
		Description: fmt.Sprintf("Explored area (discovered=%v)", discovered),
		Timestamp:   tp.Now(),
		Importance:  0.4,
	})

	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"discovered":   discovered,
		"xp_awarded":   xp,
	}).Info("Exploration action processed")
}

// GeneratePersonalityDescription creates a text description of personality.
func GeneratePersonalityDescription(pe *PersonalityEvolution) string {
	if pe == nil {
		return "No personality data"
	}
	defaultLogger.Debug("Generating personality description")

	dominant := pe.GetDominantTrait()
	description := fmt.Sprintf("Primarily %s with varying degrees of other traits", dominant.String())

	defaultLogger.WithFields(logrus.Fields{
		"dominant_trait": dominant.String(),
		"description":    description,
	}).Debug("Personality description generated")

	return description
}

// AdaptBehaviorToCombatStyle learns from player's combat preferences.
func AdaptBehaviorToCombatStyle(comp *CompanionLearningComponent, seed int64) {
	if comp == nil {
		return
	}
	defaultLogger.WithFields(logrus.Fields{
		"companion_id": comp.CompanionID,
		"seed":         seed,
	}).Debug("Adapting behavior to combat style")

	rng := rand.New(rand.NewSource(seed))

	combatEvents := comp.Memory.GetEventsByType(EventCombat)
	if len(combatEvents) < 5 {
		defaultLogger.WithFields(logrus.Fields{
			"companion_id":  comp.CompanionID,
			"combat_events": len(combatEvents),
			"required":      5,
		}).Debug("Insufficient combat events for behavior adaptation")
		return
	}

	recentCombat := combatEvents
	if len(recentCombat) > 10 {
		recentCombat = recentCombat[len(recentCombat)-10:]
	}

	// Count aggressive combat actions by parsing event descriptions
	// Events are stored as "Combat action (aggressive=true/false, successful=...)"
	aggressiveCount := 0
	for _, event := range recentCombat {
		if strings.Contains(event.Description, "aggressive=true") {
			aggressiveCount++
		}
	}

	// Use seed-based RNG only for minor personality variation, not core logic
	_ = rng // Seed preserved for potential future personality variation

	if aggressiveCount > len(recentCombat)/2 {
		defaultLogger.WithFields(logrus.Fields{
			"companion_id":     comp.CompanionID,
			"aggressive_count": aggressiveCount,
			"total_events":     len(recentCombat),
			"adaptation":       "aggressive",
		}).Info("Adapting to aggressive combat style")
		comp.Personality.AdjustTrait(TraitAggressive, 0.05, "learned aggressive combat style")
		comp.Personality.AdjustTrait(TraitBrave, 0.03, "learned aggressive combat style")
	} else {
		defaultLogger.WithFields(logrus.Fields{
			"companion_id":     comp.CompanionID,
			"aggressive_count": aggressiveCount,
			"total_events":     len(recentCombat),
			"adaptation":       "defensive",
		}).Info("Adapting to defensive combat style")
		comp.Personality.AdjustTrait(TraitCautious, 0.05, "learned defensive combat style")
		comp.Personality.AdjustTrait(TraitPacifist, 0.02, "learned defensive combat style")
	}
}
