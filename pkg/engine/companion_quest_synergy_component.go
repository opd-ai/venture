// Package engine provides the companion-quest synergy component.
// This component tracks synergy bonuses between companions and quests,
// allowing companion skills to provide quest objective bonuses and rewards.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// CompanionSkillType represents companion specializations that affect quests.
type CompanionSkillType int

const (
	// SkillNone represents no specialization
	SkillNone CompanionSkillType = iota
	// SkillTracker improves explore/scout objectives
	SkillTracker
	// SkillHunter improves kill objectives
	SkillHunter
	// SkillGatherer improves collect objectives
	SkillGatherer
	// SkillGuardian improves escort objectives
	SkillGuardian
	// SkillDiplomat improves talk/social objectives
	SkillDiplomat
	// SkillCombatant improves boss fight objectives
	SkillCombatant
)

// String returns the skill type name.
func (s CompanionSkillType) String() string {
	switch s {
	case SkillTracker:
		return "Tracker"
	case SkillHunter:
		return "Hunter"
	case SkillGatherer:
		return "Gatherer"
	case SkillGuardian:
		return "Guardian"
	case SkillDiplomat:
		return "Diplomat"
	case SkillCombatant:
		return "Combatant"
	default:
		return "None"
	}
}

// MatchesQuestType returns true if this skill synergizes with the quest type.
func (s CompanionSkillType) MatchesQuestType(qt quest.QuestType) bool {
	switch s {
	case SkillTracker:
		return qt == quest.TypeExplore
	case SkillHunter:
		return qt == quest.TypeKill
	case SkillGatherer:
		return qt == quest.TypeCollect
	case SkillGuardian:
		return qt == quest.TypeEscort
	case SkillDiplomat:
		return qt == quest.TypeTalk || qt == quest.TypeFactionConflict
	case SkillCombatant:
		return qt == quest.TypeBoss
	default:
		return false
	}
}

// QuestSynergyBonus tracks bonuses applied to a specific quest from companions.
type QuestSynergyBonus struct {
	// QuestID is the quest receiving bonuses
	QuestID string
	// CompanionID is the entity ID of the companion providing bonuses
	CompanionID uint64
	// SkillType is the companion's specialization
	SkillType CompanionSkillType
	// ObjectiveBonus is the progress multiplier for matching objectives (1.0 = no bonus)
	ObjectiveBonus float64
	// RewardBonus is the reward multiplier at completion (1.0 = no bonus)
	RewardBonus float64
	// LoyaltyAtStart was the companion's loyalty when quest started
	LoyaltyAtStart float64
	// Active indicates if this synergy is currently being applied
	Active bool
}

// CompanionQuestSynergyComponent tracks companion-quest synergy bonuses.
// Attached to player entities to track their companions' quest contributions.
type CompanionQuestSynergyComponent struct {
	// ActiveSynergies maps quest ID to synergy bonuses from companions
	ActiveSynergies map[string]*QuestSynergyBonus
	// CompletedSynergies tracks historical synergies for completed quests
	CompletedSynergies []*QuestSynergyBonus
	// TotalBonusXP is cumulative bonus XP earned from companion synergies
	TotalBonusXP int
	// TotalBonusGold is cumulative bonus gold earned from companion synergies
	TotalBonusGold int
	// QuestsCompletedWithSynergy counts quests that had active synergies
	QuestsCompletedWithSynergy int
}

// Type returns the component type identifier.
func (c *CompanionQuestSynergyComponent) Type() string {
	return "companion_quest_synergy"
}

// NewCompanionQuestSynergyComponent creates a new synergy tracking component.
func NewCompanionQuestSynergyComponent() *CompanionQuestSynergyComponent {
	return &CompanionQuestSynergyComponent{
		ActiveSynergies:    make(map[string]*QuestSynergyBonus),
		CompletedSynergies: make([]*QuestSynergyBonus, 0),
	}
}

// AddSynergy registers a new companion-quest synergy.
func (c *CompanionQuestSynergyComponent) AddSynergy(questID string, companionID uint64, skillType CompanionSkillType, loyalty float64) {
	// Calculate bonuses based on skill match and loyalty
	objectiveBonus := 1.0
	rewardBonus := 1.0

	// Base bonus from having a companion (5%)
	objectiveBonus += 0.05
	rewardBonus += 0.05

	// Loyalty bonus (up to 15% at 100 loyalty)
	loyaltyMultiplier := loyalty / 100.0
	objectiveBonus += loyaltyMultiplier * 0.15
	rewardBonus += loyaltyMultiplier * 0.15

	c.ActiveSynergies[questID] = &QuestSynergyBonus{
		QuestID:        questID,
		CompanionID:    companionID,
		SkillType:      skillType,
		ObjectiveBonus: objectiveBonus,
		RewardBonus:    rewardBonus,
		LoyaltyAtStart: loyalty,
		Active:         true,
	}
}

// ApplySkillMatch adds bonus when companion skill matches quest type.
func (c *CompanionQuestSynergyComponent) ApplySkillMatch(questID string, matches bool) {
	synergy, exists := c.ActiveSynergies[questID]
	if !exists || !synergy.Active {
		return
	}

	if matches {
		// Skill match adds 25% to objective progress and 20% to rewards
		synergy.ObjectiveBonus += 0.25
		synergy.RewardBonus += 0.20
	}
}

// GetObjectiveBonus returns the objective progress multiplier for a quest.
func (c *CompanionQuestSynergyComponent) GetObjectiveBonus(questID string) float64 {
	synergy, exists := c.ActiveSynergies[questID]
	if !exists || !synergy.Active {
		return 1.0
	}
	return synergy.ObjectiveBonus
}

// GetRewardBonus returns the reward multiplier for a quest.
func (c *CompanionQuestSynergyComponent) GetRewardBonus(questID string) float64 {
	synergy, exists := c.ActiveSynergies[questID]
	if !exists || !synergy.Active {
		return 1.0
	}
	return synergy.RewardBonus
}

// CompleteSynergy moves a synergy to completed and records stats.
func (c *CompanionQuestSynergyComponent) CompleteSynergy(questID string, bonusXP, bonusGold int) {
	synergy, exists := c.ActiveSynergies[questID]
	if !exists {
		return
	}

	synergy.Active = false
	c.CompletedSynergies = append(c.CompletedSynergies, synergy)
	delete(c.ActiveSynergies, questID)

	c.TotalBonusXP += bonusXP
	c.TotalBonusGold += bonusGold
	c.QuestsCompletedWithSynergy++
}

// RemoveSynergy removes a synergy (for abandoned quests).
func (c *CompanionQuestSynergyComponent) RemoveSynergy(questID string) {
	delete(c.ActiveSynergies, questID)
}

// HasActiveSynergy checks if a quest has an active companion synergy.
func (c *CompanionQuestSynergyComponent) HasActiveSynergy(questID string) bool {
	synergy, exists := c.ActiveSynergies[questID]
	return exists && synergy.Active
}

// GetCompanionSkillForType derives companion skill from companion type.
// This maps companion types to their natural quest skill affinity.
func GetCompanionSkillForType(companionType CompanionType) CompanionSkillType {
	switch companionType {
	case CompanionTypePet:
		return SkillTracker // Pets are good at tracking/exploring
	case CompanionTypeSummon:
		return SkillCombatant // Summons excel in combat
	case CompanionTypeHireling:
		return SkillGatherer // Hirelings gather resources
	case CompanionTypeElemental:
		return SkillCombatant // Elementals fight
	case CompanionTypeUndead:
		return SkillHunter // Undead hunt prey
	case CompanionTypeRobot:
		return SkillTracker // Robots scout efficiently
	case CompanionTypeSpirit:
		return SkillDiplomat // Spirits aid in social interaction
	case CompanionTypeInsect:
		return SkillGatherer // Insects gather
	default:
		return SkillNone
	}
}
