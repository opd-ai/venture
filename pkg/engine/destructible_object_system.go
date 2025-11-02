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
	pos := posComp.(*PositionComponent)

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

	// Deal area damage to entities in explosion radius
	entities := s.world.GetEntities()
	for _, entity := range entities {
		// Get entity position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entityPos := posComp.(*PositionComponent)

		// Calculate distance
		dx := entityPos.X - x
		dy := entityPos.Y - y
		distSq := dx*dx + dy*dy
		radiusSq := destructibleObj.ExplosionRadius * destructibleObj.ExplosionRadius

		// Check if in explosion radius
		if distSq <= radiusSq {
			// Get health component if entity has one
			healthComp, ok := entity.GetComponent("health")
			if !ok {
				continue
			}
			health, ok := healthComp.(*HealthComponent)
			if !ok {
				continue
			}

			// Apply explosion damage (full damage at center, reduced at edges)
			distPercent := distSq / radiusSq
			damageMultiplier := 1.0 - distPercent // 1.0 at center, 0.0 at edge
			damage := destructibleObj.ExplosionDamage * damageMultiplier

			health.TakeDamage(damage)

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"targetID": entity.ID,
					"damage":   damage,
					"distance": distPercent,
				}).Debug("explosion damage applied")
			}
		}
	}

	// Ignite tiles in explosion radius if fire system is available
	if s.fireSystem != nil {
		intensity := 0.8 // Explosions create intense fires
		s.fireSystem.IgniteTilesInArea(x, y, destructibleObj.ExplosionRadius*0.5, intensity)
	}

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

	// Find destructible objects in range
	entities := s.world.GetEntitiesWith("destructibleObject")
	for _, entity := range entities {
		// Get position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos := posComp.(*PositionComponent)

		// Check if within damage radius
		dx := pos.X - x
		dy := pos.Y - y
		distSq := dx*dx + dy*dy

		if distSq <= radius*radius {
			// Get destructible component
			objComp, ok := entity.GetComponent("destructibleObject")
			if !ok {
				continue
			}
			destructibleObj := objComp.(*DestructibleObjectComponent)

			// Apply damage
			destroyed := destructibleObj.TakeDamage(damage)
			if destroyed {
				return true
			}
		}
	}

	return false
}
