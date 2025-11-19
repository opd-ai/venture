package learning

import (
	"fmt"
	"math/rand"
	"time"
)

// Manager handles companion learning operations.
type Manager struct {
	companions map[string]*CompanionLearningComponent
}

// NewManager creates a new companion learning manager.
func NewManager() *Manager {
	return &Manager{
		companions: make(map[string]*CompanionLearningComponent),
	}
}

// AddCompanion registers a companion for learning tracking.
func (m *Manager) AddCompanion(companionID string, learningRate float64) *CompanionLearningComponent {
	if learningRate <= 0 {
		learningRate = 1.0
	}

	comp := &CompanionLearningComponent{
		CompanionID:  companionID,
		SkillTree:    NewSkillProgression(),
		Personality:  NewPersonalityEvolution(),
		Memory:       NewEventMemory(1000),
		LearningRate: learningRate,
		LastSkillUse: make(map[string]time.Time),
	}

	m.companions[companionID] = comp
	return comp
}

// GetCompanion retrieves a companion's learning component.
func (m *Manager) GetCompanion(companionID string) (*CompanionLearningComponent, bool) {
	comp, ok := m.companions[companionID]
	return comp, ok
}

// RemoveCompanion removes a companion from tracking.
func (m *Manager) RemoveCompanion(companionID string) {
	delete(m.companions, companionID)
}

// NewSkillProgression creates a new skill progression system.
func NewSkillProgression() *SkillProgression {
	sp := &SkillProgression{
		Skills:          make(map[string]*Skill),
		AvailablePoints: 0,
		TotalXP:         0,
		SkillTree:       make(map[string]*SkillNode),
	}

	sp.initializeSkillTree()
	return sp
}

// initializeSkillTree sets up the default skill tree.
func (sp *SkillProgression) initializeSkillTree() {
	// Combat skills
	sp.addSkillNode("Basic Attack", SkillCombat, "Improved melee damage", nil, 1, 10)
	sp.addSkillNode("Power Strike", SkillCombat, "Devastating attack", []string{"Basic Attack"}, 1, 10)
	sp.addSkillNode("Combat Mastery", SkillCombat, "Enhanced combat effectiveness", []string{"Power Strike"}, 2, 10)

	// Defense skills
	sp.addSkillNode("Block", SkillDefense, "Reduce incoming damage", nil, 1, 10)
	sp.addSkillNode("Iron Skin", SkillDefense, "Increased armor", []string{"Block"}, 1, 10)
	sp.addSkillNode("Defensive Stance", SkillDefense, "Maximum protection", []string{"Iron Skin"}, 2, 10)

	// Utility skills
	sp.addSkillNode("Gather", SkillUtility, "Collect resources faster", nil, 1, 10)
	sp.addSkillNode("Scout", SkillUtility, "Reveal hidden paths", []string{"Gather"}, 1, 10)
	sp.addSkillNode("Tracking", SkillUtility, "Follow enemy trails", []string{"Scout"}, 2, 10)

	// Social skills
	sp.addSkillNode("Persuasion", SkillSocial, "Convince NPCs", nil, 1, 10)
	sp.addSkillNode("Charm", SkillSocial, "Better trade prices", []string{"Persuasion"}, 1, 10)
	sp.addSkillNode("Leadership", SkillSocial, "Inspire allies", []string{"Charm"}, 2, 10)

	// Healing skills
	sp.addSkillNode("First Aid", SkillHealing, "Basic healing", nil, 1, 10)
	sp.addSkillNode("Restoration", SkillHealing, "Powerful healing", []string{"First Aid"}, 1, 10)
	sp.addSkillNode("Regeneration", SkillHealing, "Continuous healing", []string{"Restoration"}, 2, 10)

	// Magic skills
	sp.addSkillNode("Mana Control", SkillMagic, "Increased mana pool", nil, 1, 10)
	sp.addSkillNode("Spell Power", SkillMagic, "Stronger spells", []string{"Mana Control"}, 1, 10)
	sp.addSkillNode("Arcane Mastery", SkillMagic, "Ultimate magic power", []string{"Spell Power"}, 2, 10)

	// Crafting skills
	sp.addSkillNode("Apprentice Smith", SkillCrafting, "Basic crafting", nil, 1, 10)
	sp.addSkillNode("Master Smith", SkillCrafting, "Advanced crafting", []string{"Apprentice Smith"}, 1, 10)
	sp.addSkillNode("Legendary Smith", SkillCrafting, "Epic items", []string{"Master Smith"}, 2, 10)

	// Stealth skills
	sp.addSkillNode("Sneak", SkillStealth, "Move undetected", nil, 1, 10)
	sp.addSkillNode("Backstab", SkillStealth, "Critical strikes", []string{"Sneak"}, 1, 10)
	sp.addSkillNode("Shadow Walk", SkillStealth, "Become invisible", []string{"Backstab"}, 2, 10)
}

// addSkillNode adds a skill to the tree.
func (sp *SkillProgression) addSkillNode(name string, skillType SkillType, description string, prerequisites []string, cost, maxLevel int) {
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
	skill, ok := sp.Skills[skillName]
	if !ok {
		return fmt.Errorf("skill not found: %s", skillName)
	}

	adjustedXP := xp * learningRate
	sp.TotalXP += adjustedXP

	if skill.Level >= skill.MaxLevel {
		return nil // Max level reached, still count XP towards total
	}

	skill.Experience += adjustedXP

	// Level up if enough XP
	xpNeeded := float64(skill.Level+1) * 100.0
	for skill.Experience >= xpNeeded && skill.Level < skill.MaxLevel {
		skill.Experience -= xpNeeded
		skill.Level++
		sp.AvailablePoints++
		xpNeeded = float64(skill.Level+1) * 100.0
	}

	return nil
}

// CanLearnSkill checks if prerequisites are met.
func (sp *SkillProgression) CanLearnSkill(skillName string) (bool, error) {
	node, ok := sp.SkillTree[skillName]
	if !ok {
		return false, fmt.Errorf("skill not found: %s", skillName)
	}

	if sp.AvailablePoints < node.Cost {
		return false, fmt.Errorf("insufficient skill points: have %d, need %d", sp.AvailablePoints, node.Cost)
	}

	for _, prereq := range node.Prerequisites {
		prereqSkill, ok := sp.Skills[prereq]
		if !ok {
			return false, fmt.Errorf("prerequisite not found: %s", prereq)
		}
		if prereqSkill.Level < 1 {
			return false, fmt.Errorf("prerequisite not met: %s", prereq)
		}
	}

	return true, nil
}

// LearnSkill allocates points to a skill.
func (sp *SkillProgression) LearnSkill(skillName string) error {
	canLearn, err := sp.CanLearnSkill(skillName)
	if !canLearn {
		return err
	}

	node := sp.SkillTree[skillName]
	sp.AvailablePoints -= node.Cost
	return nil
}

// NewPersonalityEvolution creates a new personality system.
func NewPersonalityEvolution() *PersonalityEvolution {
	pe := &PersonalityEvolution{
		Traits:     make(map[PersonalityTrait]float64),
		Changes:    []PersonalityChange{},
		LastUpdate: time.Now(),
	}

	// Initialize traits with neutral values
	pe.Traits[TraitCautious] = 0.5
	pe.Traits[TraitBrave] = 0.5
	pe.Traits[TraitShy] = 0.5
	pe.Traits[TraitOutgoing] = 0.5
	pe.Traits[TraitAggressive] = 0.5
	pe.Traits[TraitPacifist] = 0.5
	pe.Traits[TraitLoyal] = 0.5
	pe.Traits[TraitIndependent] = 0.5
	pe.Traits[TraitCurious] = 0.5
	pe.Traits[TraitPractical] = 0.5

	return pe
}

// AdjustTrait modifies a personality trait.
func (pe *PersonalityEvolution) AdjustTrait(trait PersonalityTrait, delta float64, reason string) {
	oldValue := pe.Traits[trait]
	newValue := oldValue + delta

	// Clamp to [0.0, 1.0]
	if newValue < 0.0 {
		newValue = 0.0
	}
	if newValue > 1.0 {
		newValue = 1.0
	}

	pe.Traits[trait] = newValue
	pe.LastUpdate = time.Now()

	change := PersonalityChange{
		Trait:     trait,
		OldValue:  oldValue,
		NewValue:  newValue,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	pe.Changes = append(pe.Changes, change)
}

// GetDominantTrait returns the strongest personality trait.
func (pe *PersonalityEvolution) GetDominantTrait() PersonalityTrait {
	var dominant PersonalityTrait
	maxValue := 0.0

	for trait, value := range pe.Traits {
		if value > maxValue {
			maxValue = value
			dominant = trait
		}
	}

	return dominant
}

// NewEventMemory creates a new event memory system.
func NewEventMemory(maxEvents int) *EventMemory {
	return &EventMemory{
		Events:       []MemorableEvent{},
		MaxEvents:    maxEvents,
		TotalEvents:  0,
		FirstEventAt: time.Now(),
	}
}

// AddEvent records a memorable event.
func (em *EventMemory) AddEvent(event MemorableEvent) {
	em.Events = append(em.Events, event)
	em.TotalEvents++

	if em.TotalEvents == 1 {
		em.FirstEventAt = event.Timestamp
	}

	// LRU eviction if over limit
	if len(em.Events) > em.MaxEvents {
		em.Events = em.Events[len(em.Events)-em.MaxEvents:]
	}
}

// GetRecentEvents returns the N most recent events.
func (em *EventMemory) GetRecentEvents(n int) []MemorableEvent {
	if n > len(em.Events) {
		n = len(em.Events)
	}
	return em.Events[len(em.Events)-n:]
}

// GetEventsByType filters events by type.
func (em *EventMemory) GetEventsByType(eventType EventType) []MemorableEvent {
	var filtered []MemorableEvent
	for _, event := range em.Events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// ProcessCombatAction updates skills and personality based on combat behavior.
func ProcessCombatAction(comp *CompanionLearningComponent, aggressive, successful bool) {
	// Award XP based on action
	xp := 10.0
	if successful {
		xp = 20.0
	}

	if aggressive {
		_ = comp.SkillTree.AddExperience("Basic Attack", xp, comp.LearningRate)
		comp.Personality.AdjustTrait(TraitAggressive, 0.01, "engaged in aggressive combat")
		comp.Personality.AdjustTrait(TraitPacifist, -0.01, "engaged in aggressive combat")
	} else {
		_ = comp.SkillTree.AddExperience("Block", xp, comp.LearningRate)
		comp.Personality.AdjustTrait(TraitCautious, 0.01, "used defensive tactics")
	}

	// Record event
	comp.Memory.AddEvent(MemorableEvent{
		Type:        EventCombat,
		Description: fmt.Sprintf("Combat action (aggressive=%v, successful=%v)", aggressive, successful),
		Timestamp:   time.Now(),
		Importance:  0.6,
	})
}

// ProcessSocialInteraction updates skills and personality based on social behavior.
func ProcessSocialInteraction(comp *CompanionLearningComponent, playerID string, positive bool) {
	xp := 15.0
	if positive {
		_ = comp.SkillTree.AddExperience("Persuasion", xp, comp.LearningRate)
		comp.Personality.AdjustTrait(TraitOutgoing, 0.02, "positive social interaction")
		comp.Personality.AdjustTrait(TraitShy, -0.02, "positive social interaction")
		comp.Personality.AdjustTrait(TraitLoyal, 0.01, "bonded with player")
	} else {
		comp.Personality.AdjustTrait(TraitShy, 0.01, "negative social interaction")
		comp.Personality.AdjustTrait(TraitOutgoing, -0.01, "negative social interaction")
	}

	comp.Memory.AddEvent(MemorableEvent{
		Type:        EventDialog,
		Description: fmt.Sprintf("Social interaction with %s (positive=%v)", playerID, positive),
		Timestamp:   time.Now(),
		Importance:  0.5,
		PlayerID:    playerID,
	})
}

// ProcessExploration updates skills and personality based on exploration behavior.
func ProcessExploration(comp *CompanionLearningComponent, discovered bool) {
	xp := 8.0
	if discovered {
		xp = 12.0
	}

	_ = comp.SkillTree.AddExperience("Scout", xp, comp.LearningRate)
	comp.Personality.AdjustTrait(TraitCurious, 0.015, "explored new area")
	comp.Personality.AdjustTrait(TraitPractical, -0.005, "took exploratory risk")

	comp.Memory.AddEvent(MemorableEvent{
		Type:        EventExploration,
		Description: fmt.Sprintf("Explored area (discovered=%v)", discovered),
		Timestamp:   time.Now(),
		Importance:  0.4,
	})
}

// GeneratePersonalityDescription creates a text description of personality.
func GeneratePersonalityDescription(pe *PersonalityEvolution) string {
	dominant := pe.GetDominantTrait()
	return fmt.Sprintf("Primarily %s with varying degrees of other traits", dominant.String())
}

// AdaptBehaviorToCombatStyle learns from player's combat preferences.
func AdaptBehaviorToCombatStyle(comp *CompanionLearningComponent, seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// Check recent combat events
	combatEvents := comp.Memory.GetEventsByType(EventCombat)
	if len(combatEvents) < 5 {
		return // Not enough data
	}

	// Analyze last 10 combat events for patterns
	recentCombat := combatEvents
	if len(recentCombat) > 10 {
		recentCombat = recentCombat[len(recentCombat)-10:]
	}

	// Simple pattern detection (would be more sophisticated in production)
	aggressiveCount := 0
	for range recentCombat {
		if rng.Float64() > 0.5 { // Simulated aggression detection
			aggressiveCount++
		}
	}

	if aggressiveCount > len(recentCombat)/2 {
		comp.Personality.AdjustTrait(TraitAggressive, 0.05, "learned aggressive combat style")
		comp.Personality.AdjustTrait(TraitBrave, 0.03, "learned aggressive combat style")
	} else {
		comp.Personality.AdjustTrait(TraitCautious, 0.05, "learned defensive combat style")
		comp.Personality.AdjustTrait(TraitPacifist, 0.02, "learned defensive combat style")
	}
}
