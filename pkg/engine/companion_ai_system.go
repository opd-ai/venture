package engine

import (
	"math"
)

// CompanionAISystem handles companion AI behavior
type CompanionAISystem struct {
	world *World
}

// NewCompanionAISystem creates a new companion AI system
func NewCompanionAISystem(world *World) *CompanionAISystem {
	return &CompanionAISystem{world: world}
}

// Update processes companion AI behavior
func (s *CompanionAISystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("companion", "position")

	for _, entity := range entities {
		companionCompRaw, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp := companionCompRaw.(*CompanionComponent)

		posCompRaw, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		posComp := posCompRaw.(*PositionComponent)

		// Get owner entity
		owner, ok := s.world.GetEntity(companionComp.OwnerID)
		if !ok || owner == nil {
			continue
		}

		ownerPosRaw, ok := owner.GetComponent("position")
		if !ok {
			continue
		}
		ownerPos := ownerPosRaw.(*PositionComponent)

		// Execute behavior based on mode
		switch companionComp.Behavior {
		case BehaviorAggressive:
			s.aggressiveBehavior(entity, posComp, owner, ownerPos, deltaTime)
		case BehaviorDefensive:
			s.defensiveBehavior(entity, posComp, owner, ownerPos, deltaTime)
		case BehaviorPassive:
			s.passiveBehavior(entity, posComp, owner, ownerPos, deltaTime)
		}
	}
}

func (s *CompanionAISystem) aggressiveBehavior(entity *Entity, pos *PositionComponent, owner *Entity, ownerPos *PositionComponent, deltaTime float64) {
	// Find nearby enemies and attack
	enemies := s.findNearbyEnemies(pos, 200.0)
	if len(enemies) > 0 {
		s.attackEnemy(entity, enemies[0])
	} else {
		s.followOwner(entity, pos, ownerPos, deltaTime)
	}
}

func (s *CompanionAISystem) defensiveBehavior(entity *Entity, pos *PositionComponent, owner *Entity, ownerPos *PositionComponent, deltaTime float64) {
	// Stay near owner and defend
	dist := s.distance(pos, ownerPos)
	if dist > 50.0 {
		s.followOwner(entity, pos, ownerPos, deltaTime)
	}
}

func (s *CompanionAISystem) passiveBehavior(entity *Entity, pos *PositionComponent, owner *Entity, ownerPos *PositionComponent, deltaTime float64) {
	// Just follow owner
	s.followOwner(entity, pos, ownerPos, deltaTime)
}

func (s *CompanionAISystem) followOwner(entity *Entity, pos, ownerPos *PositionComponent, deltaTime float64) {
	dist := s.distance(pos, ownerPos)
	if dist > 30.0 {
		// Move towards owner
		dx := ownerPos.X - pos.X
		dy := ownerPos.Y - pos.Y
		mag := math.Sqrt(dx*dx + dy*dy)
		if mag > 0 {
			// Get companion stats for speed
			speed := 100.0 // Default speed
			if statsCompRaw, ok := entity.GetComponent("companionstats"); ok {
				speed = statsCompRaw.(*CompanionStatsComponent).Speed
			}

			pos.X += (dx / mag) * speed * deltaTime
			pos.Y += (dy / mag) * speed * deltaTime
		}
	}
}

func (s *CompanionAISystem) findNearbyEnemies(pos *PositionComponent, radius float64) []*Entity {
	enemies := []*Entity{}
	entities := s.world.GetEntitiesWith("ai", "position")

	for _, entity := range entities {
		enemyPosRaw, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		enemyPos := enemyPosRaw.(*PositionComponent)
		if s.distance(pos, enemyPos) <= radius {
			enemies = append(enemies, entity)
		}
	}

	return enemies
}

func (s *CompanionAISystem) attackEnemy(companion, enemy *Entity) {
	// Attack logic handled by combat system
	// This just sets the target
	if aiCompRaw, ok := companion.GetComponent("ai"); ok {
		aiCompRaw.(*AIComponent).Target = enemy
	}
}

func (s *CompanionAISystem) distance(pos1, pos2 *PositionComponent) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}
