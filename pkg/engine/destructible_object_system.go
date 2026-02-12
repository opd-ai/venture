// Package engine provides the destructible object system for Phase 11.3.
// Environmental Destruction & Manipulation
//
// This file implements DestructibleObjectSystem which handles damage to
// destructible objects, debris generation, and explosion/poison effects.
//
// Design Philosophy:
// - Destructible objects take damage from player attacks, spells, and projectiles
// - When destroyed, objects spawn debris entities with physics
// - Explosive objects create area damage and ignite nearby tiles
// - Poison containers spawn lingering poison hazard entities
// - Server-authoritative for multiplayer synchronization
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// DestructibleObjectSystem manages destructible object lifecycle.
type DestructibleObjectSystem struct {
	world          *World
	rng            *rand.Rand
	logger         *logrus.Entry
	tileSize       int
	fireSystem     *FirePropagationSystem // For igniting tiles on explosions
	debrisLifetime float64                // How long debris entities persist
}

// NewDestructibleObjectSystem creates a new destructible object system.
func NewDestructibleObjectSystem(tileSize int, seed int64) *DestructibleObjectSystem {
	return NewDestructibleObjectSystemWithLogger(tileSize, seed, nil)
}

// NewDestructibleObjectSystemWithLogger creates a system with a logger.
func NewDestructibleObjectSystemWithLogger(tileSize int, seed int64, logger *logrus.Logger) *DestructibleObjectSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system":   "destructible_object",
			"tileSize": tileSize,
			"seed":     seed,
		})
		logEntry.Debug("destructible object system created")
	}

	return &DestructibleObjectSystem{
		tileSize:       tileSize,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		debrisLifetime: 5.0, // Default: debris lasts 5 seconds
	}
}

// SetWorld sets the ECS world reference.
func (s *DestructibleObjectSystem) SetWorld(world *World) {
	s.world = world
}

// SetFireSystem sets the fire propagation system reference for igniting tiles.
func (s *DestructibleObjectSystem) SetFireSystem(fireSystem *FirePropagationSystem) {
	s.fireSystem = fireSystem
}

// Update implements the System interface.
// Processes destroyed objects and triggers their effects.
func (s *DestructibleObjectSystem) Update(entities []*Entity, deltaTime float64) {
	// Find all destructible objects
	for _, entity := range entities {
		comp, ok := entity.GetComponent("destructibleObject")
		if !ok {
			continue
		}

		destructibleObj, ok := comp.(*DestructibleObjectComponent)
		if !ok {
			continue
		}

		// Process newly destroyed objects
		if destructibleObj.IsDestroyed {
			s.processDestroyedObject(entity, destructibleObj)
		}
	}
}

// processDestroyedObject handles the destruction of an object.
func (s *DestructibleObjectSystem) processDestroyedObject(entity *Entity, destructibleObj *DestructibleObjectComponent) {
	// Get object position
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":   entity.ID,
			"objectType": destructibleObj.ObjectType.String(),
			"x":          pos.X,
			"y":          pos.Y,
		}).Debug("object destroyed")
	}

	// Handle explosive objects
	if destructibleObj.IsExplosive() {
		s.createExplosion(pos.X, pos.Y, destructibleObj)
	}

	// Handle poison containers
	if destructibleObj.EmitsPoison() {
		s.createPoisonCloud(pos.X, pos.Y, destructibleObj)
	}

	// Spawn debris
	s.spawnDebris(pos.X, pos.Y, destructibleObj)

	// Remove the destroyed object entity
	if s.world != nil {
		s.world.RemoveEntity(entity.ID)
	}
}

// createExplosion creates an explosion effect with area damage.
func (s *DestructibleObjectSystem) createExplosion(x, y float64, destructibleObj *DestructibleObjectComponent) {
	if s.world == nil {
		return
	}

	s.applyExplosionDamageToEntities(x, y, destructibleObj)
	s.igniteExplosionArea(x, y, destructibleObj)
	s.logExplosionEvent(x, y, destructibleObj)
}

// applyExplosionDamageToEntities deals area damage to entities in explosion radius.
func (s *DestructibleObjectSystem) applyExplosionDamageToEntities(x, y float64, destructibleObj *DestructibleObjectComponent) {
	entities := s.world.GetEntities()
	radiusSq := destructibleObj.ExplosionRadius * destructibleObj.ExplosionRadius

	for _, entity := range entities {
		entityPos := s.getEntityPosition(entity)
		if entityPos == nil {
			continue
		}

		distSq := s.calculateDistanceSquared(entityPos.X, entityPos.Y, x, y)
		if distSq <= radiusSq {
			s.dealExplosionDamage(entity, distSq, radiusSq, destructibleObj.ExplosionDamage)
		}
	}
}

// getEntityPosition retrieves the position component of an entity.
func (s *DestructibleObjectSystem) getEntityPosition(entity *Entity) *PositionComponent {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil
	}
	entityPos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil
	}
	return entityPos
}

// calculateDistanceSquared calculates squared distance between two points.
func (s *DestructibleObjectSystem) calculateDistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return dx*dx + dy*dy
}

// dealExplosionDamage applies damage to an entity based on distance from explosion.
func (s *DestructibleObjectSystem) dealExplosionDamage(entity *Entity, distSq, radiusSq, baseDamage float64) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	distPercent := distSq / radiusSq
	damageMultiplier := 1.0 - distPercent
	damage := baseDamage * damageMultiplier

	health.TakeDamage(damage)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"targetID": entity.ID,
			"damage":   damage,
			"distance": distPercent,
		}).Debug("explosion damage applied")
	}
}

// igniteExplosionArea ignites tiles in the explosion radius.
func (s *DestructibleObjectSystem) igniteExplosionArea(x, y float64, destructibleObj *DestructibleObjectComponent) {
	if s.fireSystem != nil {
		intensity := 0.8
		s.fireSystem.IgniteTilesInArea(x, y, destructibleObj.ExplosionRadius*0.5, intensity)
	}
}

// logExplosionEvent logs details about the explosion event.
func (s *DestructibleObjectSystem) logExplosionEvent(x, y float64, destructibleObj *DestructibleObjectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":      x,
			"y":      y,
			"radius": destructibleObj.ExplosionRadius,
			"damage": destructibleObj.ExplosionDamage,
		}).Info("explosion created")
	}
}

// createPoisonCloud spawns a poison hazard entity.
func (s *DestructibleObjectSystem) createPoisonCloud(x, y float64, destructibleObj *DestructibleObjectComponent) {
	if s.world == nil {
		return
	}

	// Create hazard entity
	entity := s.world.CreateEntity()

	// Add position
	posComp := &PositionComponent{X: x, Y: y}
	entity.AddComponent(posComp)

	// Add hazard component
	hazardComp := NewHazardComponent(HazardPoison, destructibleObj.PoisonDuration, 48.0) // 1.5 tiles radius
	entity.AddComponent(hazardComp)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":        x,
			"y":        y,
			"duration": destructibleObj.PoisonDuration,
		}).Debug("poison cloud created")
	}
}

// spawnDebris creates debris entities with physics.
func (s *DestructibleObjectSystem) spawnDebris(x, y float64, destructibleObj *DestructibleObjectComponent) {
	if s.world == nil {
		return
	}

	for i := 0; i < destructibleObj.DebrisCount; i++ {
		// Create debris entity
		entity := s.world.CreateEntity()

		// Random offset from origin
		offsetX := (s.rng.Float64() - 0.5) * 20.0 // -10 to +10 pixels
		offsetY := (s.rng.Float64() - 0.5) * 20.0

		// Add position
		posComp := &PositionComponent{
			X: x + offsetX,
			Y: y + offsetY,
		}
		entity.AddComponent(posComp)

		// Add velocity (debris flies outward from explosion)
		velocityMag := 50.0 + s.rng.Float64()*100.0 // 50-150 pixels/sec
		angle := s.rng.Float64() * 2.0 * 3.14159265 // Random direction
		velComp := &VelocityComponent{
			VX: velocityMag * float64(math.Cos(angle)),
			VY: velocityMag * float64(math.Sin(angle)),
		}
		entity.AddComponent(velComp)

		// Add debris component
		angularVel := (s.rng.Float64() - 0.5) * 10.0 // -5 to +5 rad/sec
		debrisComp := NewDebrisComponent(destructibleObj.ObjectType, s.debrisLifetime, angularVel)
		entity.AddComponent(debrisComp)

		// Add rotation component for visual effect
		rotComp := &RotationComponent{
			Angle: s.rng.Float64() * 2.0 * 3.14159265, // Random initial rotation
		}
		entity.AddComponent(rotComp)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":           x,
			"y":           y,
			"debrisCount": destructibleObj.DebrisCount,
		}).Debug("debris spawned")
	}
}

// ApplyDamageToObject applies damage to a destructible object at a position.
// This is a helper method for external systems (e.g., combat, projectiles).
// Returns true if the object was destroyed.
func (s *DestructibleObjectSystem) ApplyDamageToObject(x, y, damage, radius float64) bool {
	if s.world == nil {
		return false
	}

	entities := s.world.GetEntitiesWith("destructibleObject")
	for _, entity := range entities {
		if s.tryDamageEntity(entity, x, y, damage, radius) {
			return true
		}
	}

	return false
}

// tryDamageEntity attempts to damage a single entity if it's in range.
func (s *DestructibleObjectSystem) tryDamageEntity(entity *Entity, x, y, damage, radius float64) bool {
	pos := s.getEntityPosition(entity)
	if pos == nil {
		return false
	}

	distSq := s.calculateDistanceSquared(pos.X, pos.Y, x, y)
	if distSq > radius*radius {
		return false
	}

	destructibleObj := s.getDestructibleComponent(entity)
	if destructibleObj == nil {
		return false
	}

	return destructibleObj.TakeDamage(damage)
}

// getDestructibleComponent extracts and validates the destructible object component.
func (s *DestructibleObjectSystem) getDestructibleComponent(entity *Entity) *DestructibleObjectComponent {
	objComp, ok := entity.GetComponent("destructibleObject")
	if !ok {
		return nil
	}
	destructibleObj, ok := objComp.(*DestructibleObjectComponent)
	if !ok {
		return nil
	}
	return destructibleObj
}
