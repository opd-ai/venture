package learning

import (
	"fmt"
	"time"
)

// CompanionLearningSystem manages companion AI updates.
type CompanionLearningSystem struct {
	manager        *Manager
	updateInterval time.Duration
	lastUpdate     time.Time
}

// NewCompanionLearningSystem creates a new system for ECS integration.
func NewCompanionLearningSystem(updateInterval time.Duration) *CompanionLearningSystem {
	return &CompanionLearningSystem{
		manager:        NewManager(),
		updateInterval: updateInterval,
		lastUpdate:     time.Now(),
	}
}

// Update processes companion learning for all companions.
func (s *CompanionLearningSystem) Update(deltaTime float64) {
	now := time.Now()
	if now.Sub(s.lastUpdate) < s.updateInterval {
		return
	}

	s.lastUpdate = now

	// Process each companion
	for companionID, comp := range s.manager.companions {
		s.updateCompanion(companionID, comp, deltaTime)
	}
}

// updateCompanion performs periodic companion updates.
func (s *CompanionLearningSystem) updateCompanion(companionID string, comp *CompanionLearningComponent, deltaTime float64) {
	// Decay unused skills slightly
	for skillName, skill := range comp.SkillTree.Skills {
		lastUse, ok := comp.LastSkillUse[skillName]
		if !ok || time.Since(lastUse) > 24*time.Hour {
			// Very slow skill decay for unused skills
			if skill.Experience > 0 {
				skill.Experience -= 0.1 * deltaTime
				if skill.Experience < 0 {
					skill.Experience = 0
				}
			}
		}
	}

	// Normalize personality traits that are opposing
	s.normalizeOpposingTraits(comp.Personality)
}

// normalizeOpposingTraits balances opposing personality dimensions.
func (s *CompanionLearningSystem) normalizeOpposingTraits(pe *PersonalityEvolution) {
	// Cautious <-> Brave
	s.balanceTraits(pe, TraitCautious, TraitBrave)

	// Shy <-> Outgoing
	s.balanceTraits(pe, TraitShy, TraitOutgoing)

	// Aggressive <-> Pacifist
	s.balanceTraits(pe, TraitAggressive, TraitPacifist)

	// Loyal <-> Independent
	s.balanceTraits(pe, TraitLoyal, TraitIndependent)

	// Curious <-> Practical
	s.balanceTraits(pe, TraitCurious, TraitPractical)
}

// balanceTraits ensures opposing traits sum to roughly 1.0.
func (s *CompanionLearningSystem) balanceTraits(pe *PersonalityEvolution, trait1, trait2 PersonalityTrait) {
	val1 := pe.Traits[trait1]
	val2 := pe.Traits[trait2]

	sum := val1 + val2
	if sum > 1.2 || sum < 0.8 {
		// Normalize to sum of 1.0 while preserving ratio
		if sum > 0 {
			pe.Traits[trait1] = val1 / sum
			pe.Traits[trait2] = val2 / sum
		} else {
			pe.Traits[trait1] = 0.5
			pe.Traits[trait2] = 0.5
		}
	}
}

// GetManager returns the underlying manager.
func (s *CompanionLearningSystem) GetManager() *Manager {
	return s.manager
}

// RecordSkillUse marks a skill as recently used.
func RecordSkillUse(comp *CompanionLearningComponent, skillName string) {
	comp.LastSkillUse[skillName] = time.Now()
}

// GetSkillBonus calculates a bonus multiplier based on skill level.
func GetSkillBonus(comp *CompanionLearningComponent, skillName string) float64 {
	skill, ok := comp.SkillTree.Skills[skillName]
	if !ok {
		return 1.0
	}

	return 1.0 + (float64(skill.Level) * 0.1)
}

// GetPersonalityInfluence returns how much a trait influences behavior.
func GetPersonalityInfluence(comp *CompanionLearningComponent, trait PersonalityTrait) float64 {
	value, ok := comp.Personality.Traits[trait]
	if !ok {
		return 0.5
	}
	return value
}

// IsSkillMaxed checks if a skill is at maximum level.
func IsSkillMaxed(comp *CompanionLearningComponent, skillName string) bool {
	skill, ok := comp.SkillTree.Skills[skillName]
	if !ok {
		return false
	}
	return skill.Level >= skill.MaxLevel
}

// GetTotalSkillPoints returns total skill points spent.
func GetTotalSkillPoints(comp *CompanionLearningComponent) int {
	total := 0
	for _, node := range comp.SkillTree.SkillTree {
		if node.Skill.Level > 0 {
			total += node.Cost
		}
	}
	return total
}

// GetSkillsByType returns all skills of a specific type.
func GetSkillsByType(comp *CompanionLearningComponent, skillType SkillType) []*Skill {
	var skills []*Skill
	for _, skill := range comp.SkillTree.Skills {
		if skill.Type == skillType {
			skills = append(skills, skill)
		}
	}
	return skills
}

// GetMemorySummary generates a summary of companion memories.
func GetMemorySummary(comp *CompanionLearningComponent) string {
	if len(comp.Memory.Events) == 0 {
		return "No memories yet"
	}

	eventCounts := make(map[EventType]int)
	for _, event := range comp.Memory.Events {
		eventCounts[event.Type]++
	}

	summary := fmt.Sprintf("Total memories: %d\n", comp.Memory.TotalEvents)
	for eventType, count := range eventCounts {
		summary += fmt.Sprintf("  %s: %d\n", eventType.String(), count)
	}

	return summary
}

// CalculateLearningProgress returns overall learning progress (0.0-1.0).
func CalculateLearningProgress(comp *CompanionLearningComponent) float64 {
	totalSkills := len(comp.SkillTree.Skills)
	if totalSkills == 0 {
		return 0.0
	}

	totalLevels := 0
	maxPossibleLevels := 0

	for _, skill := range comp.SkillTree.Skills {
		totalLevels += skill.Level
		maxPossibleLevels += skill.MaxLevel
	}

	if maxPossibleLevels == 0 {
		return 0.0
	}

	return float64(totalLevels) / float64(maxPossibleLevels)
}

// ShouldLearnNewSkill determines if companion should auto-learn a skill.
func ShouldLearnNewSkill(comp *CompanionLearningComponent, skillName string) bool {
	if comp.SkillTree.AvailablePoints <= 0 {
		return false
	}

	canLearn, _ := comp.SkillTree.CanLearnSkill(skillName)
	if !canLearn {
		return false
	}

	// Check if skill aligns with personality
	node := comp.SkillTree.SkillTree[skillName]
	skill := node.Skill

	dominant := comp.Personality.GetDominantTrait()

	switch skill.Type {
	case SkillCombat:
		return dominant == TraitAggressive || dominant == TraitBrave
	case SkillDefense:
		return dominant == TraitCautious || dominant == TraitPacifist
	case SkillSocial:
		return dominant == TraitOutgoing || dominant == TraitLoyal
	case SkillUtility:
		return dominant == TraitCurious || dominant == TraitPractical
	case SkillStealth:
		return dominant == TraitCautious || dominant == TraitIndependent
	default:
		return true
	}
}
