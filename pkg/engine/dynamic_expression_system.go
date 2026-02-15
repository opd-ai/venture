// Package engine provides the DynamicExpressionSystem which reactively updates
// NpcFacialDetailComponent expression state based on entity health, active status
// effects, and combat state. This connects the combat and status effect systems
// to the facial rendering pipeline so that NPC facial expressions change
// dynamically during gameplay (e.g., low health → pained, in combat → hostile,
// poisoned → scared).
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// DynamicExpressionSystem updates the ExpressionType field of
// NpcFacialDetailComponent based on entity state. It reads HealthComponent,
// StatusEffectComponent, and AIComponent to determine the appropriate
// expression, then marks the entity's sprite as dirty for re-rendering.
type DynamicExpressionSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	updateInterval float64
	timeSinceCheck float64
}

// NewDynamicExpressionSystem creates a new dynamic expression system.
func NewDynamicExpressionSystem(world *World, seed int64) *DynamicExpressionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "dynamic_expression")
	}

	sys := &DynamicExpressionSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 0.5,
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("dynamic expression system created")
	}
	return sys
}

// Update checks entity states and updates facial expressions accordingly.
func (s *DynamicExpressionSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	updated := 0
	for _, entity := range entities {
		if s.updateExpression(entity) {
			updated++
		}
	}

	if s.logger != nil && updated > 0 && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("entities_updated", updated).Debug("expressions updated")
	}
}

// updateExpression evaluates entity state and sets the appropriate expression.
// Returns true if the expression was changed.
func (s *DynamicExpressionSystem) updateExpression(entity *Entity) bool {
	comp, ok := entity.GetComponent("npc_facial_detail")
	if !ok {
		return false
	}
	faceComp, ok := comp.(*NpcFacialDetailComponent)
	if !ok {
		return false
	}

	newExpr := s.determineExpression(entity)
	if newExpr == faceComp.ExpressionType {
		return false
	}

	faceComp.ExpressionType = newExpr

	// Mark animation as dirty so sprite re-renders with new expression
	if animComp, ok := entity.GetComponent("animation"); ok {
		if anim, ok := animComp.(*AnimationComponent); ok {
			anim.Dirty = true
		}
	}

	return true
}

// determineExpression evaluates entity state to pick the best expression.
// Priority: dead > low health (pained) > status effects > combat state > neutral.
func (s *DynamicExpressionSystem) determineExpression(entity *Entity) string {
	// Check health — dead or low health
	health := entity.GetHealth()
	if health != nil {
		if health.Current <= 0 {
			return "scared" // dying/dead entities look scared
		}
		if health.Max > 0 && health.Current/health.Max < 0.25 {
			return "scared" // very low HP
		}
	}

	// Check status effects
	if statusComp, ok := entity.GetComponent("status_effect"); ok {
		if se, ok := statusComp.(*StatusEffectComponent); ok {
			if !se.IsExpired() {
				switch se.EffectType {
				case "poison", "burn", "bleed":
					return "scared"
				case "stun", "freeze":
					return "scared"
				case "rage", "berserk":
					return "hostile"
				case "heal", "shield", "buff":
					return "friendly"
				}
			}
		}
	}

	// Check AI/combat state
	if aiComp, ok := entity.GetComponent("ai"); ok {
		if ai, ok := aiComp.(*AIComponent); ok {
			switch ai.State {
			case AIStateAttack, AIStateChase:
				return "hostile"
			case AIStateFlee:
				return "scared"
			case AIStateIdle, AIStatePatrol, AIStateReturn:
				return "neutral"
			}
		}
	}

	return "neutral"
}
