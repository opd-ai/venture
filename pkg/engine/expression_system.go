package engine

import (
	"math"
)

// ExpressionSystem handles player expressions and emotes
type ExpressionSystem struct {
	world *World
}

// NewExpressionSystem creates a new expression system
func NewExpressionSystem(world *World) *ExpressionSystem {
	return &ExpressionSystem{world: world}
}

// Update processes active expressions
func (s *ExpressionSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("expression")
	
	for _, entity := range entities {
		expCompRaw, ok := entity.GetComponent("expression")
		if !ok {
			continue
		}
		expComp := expCompRaw.(*ExpressionComponent)
		
		// Update expression timer
		if expComp.ExpressionTime > 0 {
			expComp.ExpressionTime -= deltaTime
		}
		
		// Update cooldown
		if expComp.Cooldown > 0 {
			expComp.Cooldown -= deltaTime
		}
	}
}

// TriggerExpression starts an expression
func (s *ExpressionSystem) TriggerExpression(entityID uint64, expressionType ExpressionType) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}
	
	expCompRaw, ok := entity.GetComponent("expression")
	if !ok {
		// Add component if it doesn't exist
		expComp := &ExpressionComponent{
			ActiveExpression: expressionType,
			ExpressionTime:   3.0, // 3 second duration
			Cooldown:         3.0, // 3 second cooldown
		}
		entity.AddComponent(expComp)
		return true
	}
	
	expComp := expCompRaw.(*ExpressionComponent)
	
	// Check cooldown
	if expComp.Cooldown > 0 {
		return false
	}
	
	// Set new expression
	expComp.ActiveExpression = expressionType
	expComp.ExpressionTime = 3.0
	expComp.Cooldown = 3.0
	
	return true
}

// GetDuration returns the duration of an expression type
func (s *ExpressionSystem) GetDuration(expressionType ExpressionType) float64 {
	// Most expressions are 3 seconds
	switch expressionType {
	case ExpressionDance:
		return 5.0 // Dancing is longer
	case ExpressionSit, ExpressionSleep:
		return math.Inf(1) // These last until canceled
	default:
		return 3.0
	}
}
