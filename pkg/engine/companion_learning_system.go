package engine

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
)

// CompanionLearningSystem integrates companion AI learning into the ECS framework.
type CompanionLearningSystem struct {
	world          *World
	learningSystem *learning.CompanionLearningSystem
}

// NewCompanionLearningSystem creates a new system for companion AI learning.
func NewCompanionLearningSystem(world *World) *CompanionLearningSystem {
	return &CompanionLearningSystem{
		world:          world,
		learningSystem: learning.NewCompanionLearningSystem(time.Second),
	}
}

// Update processes companion learning updates.
func (cls *CompanionLearningSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity has the required components
		if !entity.HasComponent("companion") || !entity.HasComponent("companion_learning") {
			continue
		}

		learningCompRaw, ok := entity.GetComponent("companion_learning")
		if !ok {
			continue
		}
		companionCompRaw, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}

		learningComp := learningCompRaw.(*learning.CompanionLearningComponent)
		companionComp := companionCompRaw.(*CompanionComponent)

		cls.processCompanionActions(entity, learningComp, companionComp, deltaTime)
	}

	cls.learningSystem.Update(deltaTime)
}

// processCompanionActions analyzes companion behavior and updates learning.
func (cls *CompanionLearningSystem) processCompanionActions(entity *Entity, learningComp *learning.CompanionLearningComponent, companionComp *CompanionComponent, deltaTime float64) {
	healthCompRaw, ok := entity.GetComponent("health")
	if ok {
		health := healthCompRaw.(*HealthComponent)
		if health.Current < health.Max*0.3 {
			learningComp.Personality.AdjustTrait(learning.TraitCautious, 0.02*deltaTime, "low health survival instinct")
		}
	}

	switch companionComp.Behavior {
	case BehaviorAggressive:
		learningComp.Personality.AdjustTrait(learning.TraitAggressive, 0.01*deltaTime, "aggressive behavior")
	case BehaviorDefensive:
		learningComp.Personality.AdjustTrait(learning.TraitCautious, 0.01*deltaTime, "defensive behavior")
	case BehaviorPassive:
		learningComp.Personality.AdjustTrait(learning.TraitPacifist, 0.01*deltaTime, "passive behavior")
	}
}

// ProcessCombatAction records combat experience for a companion.
func (cls *CompanionLearningSystem) ProcessCombatAction(companionID uint64, aggressive, successful bool) error {
	entity, ok := cls.world.GetEntity(companionID)
	if !ok || entity == nil {
		return fmt.Errorf("companion entity not found: %d", companionID)
	}

	learningCompRaw, ok := entity.GetComponent("companion_learning")
	if !ok {
		return fmt.Errorf("companion has no learning component: %d", companionID)
	}

	comp := learningCompRaw.(*learning.CompanionLearningComponent)
	learning.ProcessCombatAction(comp, aggressive, successful)

	return nil
}

// ProcessSocialInteraction records social experience for a companion.
func (cls *CompanionLearningSystem) ProcessSocialInteraction(companionID uint64, playerID string, positive bool) error {
	entity, ok := cls.world.GetEntity(companionID)
	if !ok || entity == nil {
		return fmt.Errorf("companion entity not found: %d", companionID)
	}

	learningCompRaw, ok := entity.GetComponent("companion_learning")
	if !ok {
		return fmt.Errorf("companion has no learning component: %d", companionID)
	}

	comp := learningCompRaw.(*learning.CompanionLearningComponent)
	learning.ProcessSocialInteraction(comp, playerID, positive)

	return nil
}

// AddCompanionLearning initializes learning for a companion entity.
func (cls *CompanionLearningSystem) AddCompanionLearning(companionID uint64, learningRate float64) error {
	entity, ok := cls.world.GetEntity(companionID)
	if !ok || entity == nil {
		return fmt.Errorf("companion entity not found: %d", companionID)
	}

	manager := cls.learningSystem.GetManager()
	compID := fmt.Sprintf("companion_%d", companionID)
	learningComp := manager.AddCompanion(compID, learningRate)

	entity.AddComponent(learningComp)

	return nil
}

// GetCompanionSkillBonus calculates skill bonus for a companion.
func (cls *CompanionLearningSystem) GetCompanionSkillBonus(companionID uint64, skillName string) float64 {
	entity, ok := cls.world.GetEntity(companionID)
	if !ok || entity == nil {
		return 1.0
	}

	learningCompRaw, ok := entity.GetComponent("companion_learning")
	if !ok {
		return 1.0
	}

	comp := learningCompRaw.(*learning.CompanionLearningComponent)
	return learning.GetSkillBonus(comp, skillName)
}

// GetPersonalityInfluence returns personality influence for decision making.
func (cls *CompanionLearningSystem) GetPersonalityInfluence(companionID uint64, trait learning.PersonalityTrait) float64 {
	entity, ok := cls.world.GetEntity(companionID)
	if !ok || entity == nil {
		return 0.5
	}

	learningCompRaw, ok := entity.GetComponent("companion_learning")
	if !ok {
		return 0.5
	}

	comp := learningCompRaw.(*learning.CompanionLearningComponent)
	return learning.GetPersonalityInfluence(comp, trait)
}

// GetManager returns the underlying learning manager for direct access.
func (cls *CompanionLearningSystem) GetManager() *learning.Manager {
	return cls.learningSystem.GetManager()
}
