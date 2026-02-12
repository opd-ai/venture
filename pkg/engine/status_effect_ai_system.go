// Package engine provides the StatusEffectAISystem which bridges status effects
// with AI behavior, disabling actions for entities affected by control effects.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectAISystem disables AI actions for entities under control status effects.
// This bridges the StatusEffectSystem with AISystem by pausing AI decision-making
// when entities are stunned, frozen, feared, or paralyzed.
//
// Supported control effects:
//   - "stun": Complete action lock (AI cannot act)
//   - "frozen": Complete action lock (too cold to act)
//   - "fear": Forces flee state, prevents other actions
//   - "paralyze": Complete action lock (muscle paralysis)
//   - "sleep": Complete action lock (unconscious)
//   - "charm": Prevents hostile actions toward charmer
type StatusEffectAISystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry

	// Control effect types that disable AI actions
	controlEffects map[string]bool

	// Cache for tracking which entities have control effects applied
	// Maps entityID -> true if currently under control effect
	controlledCache map[uint64]bool
}

// NewStatusEffectAISystem creates a new status effect AI system.
func NewStatusEffectAISystem(world *World, seed int64) *StatusEffectAISystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_ai")
	}

	s := &StatusEffectAISystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		controlledCache: make(map[uint64]bool, 64),
		controlEffects:  make(map[string]bool),
	}

	// Initialize control effect types
	s.controlEffects["stun"] = true
	s.controlEffects["stunned"] = true
	s.controlEffects["frozen"] = true
	s.controlEffects["freeze"] = true
	s.controlEffects["fear"] = true
	s.controlEffects["feared"] = true
	s.controlEffects["paralyze"] = true
	s.controlEffects["paralyzed"] = true
	s.controlEffects["sleep"] = true
	s.controlEffects["asleep"] = true
	s.controlEffects["charm"] = true
	s.controlEffects["charmed"] = true

	return s
}

// Update processes all entities and disables AI for those under control effects.
func (s *StatusEffectAISystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.controlledCache {
		delete(s.controlledCache, k)
	}

	for _, entity := range entities {
		aiComp := s.getAIComponent(entity)
		if aiComp == nil {
			continue
		}

		controlEffect := s.findControlEffect(entity)
		if controlEffect == "" {
			// No control effect - restore AI if it was disabled
			if aiComp.Disabled {
				s.restoreAI(entity, aiComp)
			}
			continue
		}

		// Apply control effect to AI
		s.applyControlEffect(entity, aiComp, controlEffect)
		s.controlledCache[entity.ID] = true
	}
}

// getAIComponent retrieves the AI component from an entity.
func (s *StatusEffectAISystem) getAIComponent(entity *Entity) *AIComponent {
	comp, ok := entity.GetComponent("ai")
	if !ok {
		return nil
	}
	aiComp, ok := comp.(*AIComponent)
	if !ok {
		return nil
	}
	return aiComp
}

// findControlEffect checks if entity has any control status effect.
// Returns the effect type if found, empty string otherwise.
func (s *StatusEffectAISystem) findControlEffect(entity *Entity) string {
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}

		if s.controlEffects[effect.EffectType] && !effect.IsExpired() {
			return effect.EffectType
		}
	}
	return ""
}

// applyControlEffect disables AI based on the control effect type.
func (s *StatusEffectAISystem) applyControlEffect(entity *Entity, aiComp *AIComponent, effectType string) {
	// Already disabled - skip redundant logging
	if aiComp.Disabled && aiComp.DisabledReason == effectType {
		return
	}

	previousState := aiComp.State

	// Handle fear specially - force flee state
	if effectType == "fear" || effectType == "feared" {
		if aiComp.State != AIStateFlee {
			aiComp.ChangeState(AIStateFlee)
			aiComp.DisabledReason = effectType
			s.logControlApplied(entity, effectType, previousState, AIStateFlee)
		}
		return
	}

	// Handle charm - prevent hostile actions but allow movement
	if effectType == "charm" || effectType == "charmed" {
		// Mark as disabled for attack purposes
		aiComp.Disabled = true
		aiComp.DisabledReason = effectType
		// Force idle if attacking
		if aiComp.State == AIStateAttack || aiComp.State == AIStateChase {
			aiComp.ChangeState(AIStateIdle)
			aiComp.Target = nil
		}
		s.logControlApplied(entity, effectType, previousState, aiComp.State)
		return
	}

	// Full disable for stun, frozen, paralyze, sleep
	aiComp.Disabled = true
	aiComp.DisabledReason = effectType
	aiComp.PreviousState = previousState

	s.logControlApplied(entity, effectType, previousState, aiComp.State)
}

// restoreAI re-enables AI after control effect expires.
func (s *StatusEffectAISystem) restoreAI(entity *Entity, aiComp *AIComponent) {
	previousReason := aiComp.DisabledReason
	aiComp.Disabled = false
	aiComp.DisabledReason = ""

	// Restore to previous state if it was saved
	if aiComp.PreviousState != AIStateIdle || aiComp.State == AIStateFlee {
		aiComp.ChangeState(aiComp.PreviousState)
		aiComp.PreviousState = AIStateIdle
	}

	s.logControlRemoved(entity, previousReason)
}

// logControlApplied logs when a control effect disables AI.
func (s *StatusEffectAISystem) logControlApplied(entity *Entity, effectType string, from, to AIState) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"effect_type":    effectType,
		"previous_state": from.String(),
		"new_state":      to.String(),
	}).Debug("AI disabled by control effect")
}

// logControlRemoved logs when a control effect expires.
func (s *StatusEffectAISystem) logControlRemoved(entity *Entity, effectType string) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"effect_type": effectType,
	}).Debug("AI restored after control effect expired")
}

// IsEntityControlled returns true if the entity is under a control effect.
func (s *StatusEffectAISystem) IsEntityControlled(entityID uint64) bool {
	return s.controlledCache[entityID]
}

// RegisterControlEffect adds a custom control effect type.
func (s *StatusEffectAISystem) RegisterControlEffect(effectType string) {
	s.controlEffects[effectType] = true
}

// UnregisterControlEffect removes a control effect type.
func (s *StatusEffectAISystem) UnregisterControlEffect(effectType string) {
	delete(s.controlEffects, effectType)
}
