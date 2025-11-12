package engine

import (
	"math"
)

// ExpressionSystem handles player expressions and emotes.
// Integrates with animation and audio systems to provide visual and auditory feedback.
//
// Phase 26.1: Expression Framework
type ExpressionSystem struct {
	world          *World
	audioManager   *AudioManager // Optional audio integration
	expressionDefs map[ExpressionType]Expression
}

// NewExpressionSystem creates a new expression system.
// audioManager can be nil if audio integration is not needed.
func NewExpressionSystem(world *World, audioManager *AudioManager) *ExpressionSystem {
	s := &ExpressionSystem{
		world:          world,
		audioManager:   audioManager,
		expressionDefs: make(map[ExpressionType]Expression),
	}

	// Initialize expression definitions
	s.initializeExpressions()

	return s
}

// initializeExpressions creates Expression definitions for all expression types.
func (s *ExpressionSystem) initializeExpressions() {
	// Create a BaseExpression for each ExpressionType
	expressionTypes := []ExpressionType{
		ExpressionWave, ExpressionCheer, ExpressionDance, ExpressionLaugh,
		ExpressionCry, ExpressionSit, ExpressionPoint, ExpressionSalute,
		ExpressionShrug, ExpressionThumbsUp, ExpressionFacepalm, ExpressionSleep,
	}

	for _, expType := range expressionTypes {
		s.expressionDefs[expType] = NewBaseExpression(expType)
	}
}

// Update processes active expressions and updates animation states.
func (s *ExpressionSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("expression")

	for _, entity := range entities {
		expCompRaw, ok := entity.GetComponent("expression")
		if !ok {
			continue
		}
		expComp := expCompRaw.(*ExpressionComponent)

		// Update expression timer
		if expComp.ExpressionTime > 0 && !math.IsInf(expComp.ExpressionTime, 1) {
			expComp.ExpressionTime -= deltaTime
			
			// If expression finished, clear it
			if expComp.ExpressionTime <= 0 {
				expComp.ExpressionTime = 0
				// Could trigger OnComplete callback here if needed
			}
		}

		// Update cooldown
		if expComp.Cooldown > 0 {
			expComp.Cooldown -= deltaTime
			if expComp.Cooldown < 0 {
				expComp.Cooldown = 0
			}
		}

		// Update animation if entity has AnimationComponent
		if expComp.ExpressionTime > 0 {
			s.updateExpressionAnimation(entity, expComp)
		}
	}
}

// updateExpressionAnimation synchronizes the entity's animation with the active expression.
func (s *ExpressionSystem) updateExpressionAnimation(entity *Entity, expComp *ExpressionComponent) {
	// Get the expression definition
	expr, ok := s.expressionDefs[expComp.ActiveExpression]
	if !ok {
		return
	}

	// If entity has an animation component, update it with expression animation
	animCompRaw, hasAnim := entity.GetComponent("animation")
	if !hasAnim {
		return
	}

	animComp := animCompRaw.(*AnimationComponent)
	anim := expr.GetAnimation()

	// Sync animation properties
	animComp.FrameCount = anim.GetFrameCount()
	animComp.FrameTime = anim.GetFrameTime()
	animComp.Loop = anim.ShouldLoop()
	animComp.Playing = true
}

// TriggerExpression starts an expression for the given entity.
// Returns true if the expression was successfully triggered, false if on cooldown.
func (s *ExpressionSystem) TriggerExpression(entityID uint64, expressionType ExpressionType) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}

	// Get expression definition first
	expr, ok := s.expressionDefs[expressionType]
	if !ok {
		return false
	}

	// Get or create expression component
	expCompRaw, ok := entity.GetComponent("expression")
	var expComp *ExpressionComponent
	
	if !ok {
		// Create new component
		expComp = &ExpressionComponent{
			ActiveExpression: expressionType,
			ExpressionTime:   expr.GetDuration(),
			Cooldown:         3.0, // 3 second cooldown
		}
		entity.AddComponent(expComp)
	} else {
		// Use existing component
		expComp = expCompRaw.(*ExpressionComponent)
		
		// Check cooldown (3 second spam prevention)
		if expComp.Cooldown > 0 {
			return false
		}
		
		// Set new expression
		expComp.ActiveExpression = expressionType
		expComp.ExpressionTime = expr.GetDuration()
		expComp.Cooldown = 3.0 // 3 second cooldown
	}

	// Play sound effect if available
	if s.audioManager != nil {
		soundEffect := expr.GetSoundEffect()
		if soundEffect != "" {
			// Use a deterministic seed based on entity ID and expression type
			effectSeed := int64(entityID) + int64(expressionType)
			s.audioManager.PlaySFX(soundEffect, effectSeed)
		}
	}

	// Update animation
	s.updateExpressionAnimation(entity, expComp)

	return true
}

// CancelExpression stops the current expression for an entity.
// Useful for canceling infinite duration expressions (Sit, Sleep).
func (s *ExpressionSystem) CancelExpression(entityID uint64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}

	expCompRaw, ok := entity.GetComponent("expression")
	if !ok {
		return false
	}

	expComp := expCompRaw.(*ExpressionComponent)
	expComp.ExpressionTime = 0

	return true
}

// GetDuration returns the duration of an expression type.
// Deprecated: Use Expression.GetDuration() instead.
func (s *ExpressionSystem) GetDuration(expressionType ExpressionType) float64 {
	expr, ok := s.expressionDefs[expressionType]
	if !ok {
		return 3.0
	}
	return expr.GetDuration()
}

// IsOnCooldown returns whether the entity's expressions are on cooldown.
func (s *ExpressionSystem) IsOnCooldown(entityID uint64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}

	expCompRaw, ok := entity.GetComponent("expression")
	if !ok {
		return false
	}

	expComp := expCompRaw.(*ExpressionComponent)
	return expComp.Cooldown > 0
}

// GetActiveExpression returns the currently active expression for an entity.
// Returns nil if no expression is active.
func (s *ExpressionSystem) GetActiveExpression(entityID uint64) *ExpressionType {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil
	}

	expCompRaw, ok := entity.GetComponent("expression")
	if !ok {
		return nil
	}

	expComp := expCompRaw.(*ExpressionComponent)
	if expComp.ExpressionTime <= 0 {
		return nil
	}

	return &expComp.ActiveExpression
}
