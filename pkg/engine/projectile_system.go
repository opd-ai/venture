package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/world"
)

// ProjectileSystem manages projectile physics, collision detection, and lifecycle.
type ProjectileSystem struct {
	world *world.WorldState
	// Quadtree for efficient spatial queries (optional, can be nil for simple collision)
	quadtree *Quadtree
}

// NewProjectileSystem creates a new projectile system.
func NewProjectileSystem(w *world.WorldState) *ProjectileSystem {
	return &ProjectileSystem{
		world:    w,
		quadtree: nil, // Initialize later if spatial partitioning is available
	}
}

// SetQuadtree assigns a quadtree for efficient spatial collision detection.
func (s *ProjectileSystem) SetQuadtree(qt *Quadtree) {
	s.quadtree = qt
}

// Update processes all projectiles: movement, aging, collision detection.
func (s *ProjectileSystem) Update(deltaTime float64) {
	if s.world == nil {
		return
	}

	// Get all projectile entities
	projectiles := s.world.GetEntitiesWithComponents("projectile", "position", "velocity")

	for _, entity := range projectiles {
		s.updateProjectile(entity, deltaTime)
	}
}

// updateProjectile handles a single projectile's physics and collision.
func (s *ProjectileSystem) updateProjectile(entity *Entity, deltaTime float64) {
	projComp, ok := entity.GetComponent("projectile").(*ProjectileComponent)
	if !ok {
		return
	}

	posComp, ok := entity.GetComponent("position").(*PositionComponent)
	if !ok {
		return
	}

	velComp, ok := entity.GetComponent("velocity").(*VelocityComponent)
	if !ok {
		return
	}

	// Age the projectile
	projComp.Age += deltaTime
	if projComp.IsExpired() {
		s.despawnProjectile(entity)
		return
	}

	// Store old position for collision resolution
	oldX, oldY := posComp.X, posComp.Y

	// Move projectile
	posComp.X += velComp.VX * deltaTime
	posComp.Y += velComp.VY * deltaTime

	// Check wall collision
	if s.checkWallCollision(entity, oldX, oldY) {
		if projComp.CanBounce() {
			s.handleBounce(entity, velComp, posComp, oldX, oldY)
			if projComp.DecrementBounce() {
				// Handle explosion if explosive
				if projComp.Explosive {
					s.handleExplosion(entity, posComp)
				}
				s.despawnProjectile(entity)
			}
		} else {
			// Handle explosion if explosive
			if projComp.Explosive {
				s.handleExplosion(entity, posComp)
			}
			s.despawnProjectile(entity)
		}
		return
	}

	// Check entity collision
	hitEntity := s.checkEntityCollision(entity, posComp, projComp)
	if hitEntity != nil {
		s.handleEntityHit(entity, hitEntity, projComp, posComp)
	}
}

// checkWallCollision checks if projectile hit a wall.
func (s *ProjectileSystem) checkWallCollision(entity *Entity, oldX, oldY float64) bool {
	posComp, ok := entity.GetComponent("position").(*PositionComponent)
	if !ok {
		return false
	}

	// Get terrain map
	if s.world.CurrentMap == nil {
		return false
	}

	tileX := int(posComp.X / 32) // Assuming 32-pixel tiles
	tileY := int(posComp.Y / 32)

	// Check if position is out of bounds
	if tileX < 0 || tileY < 0 || tileX >= s.world.CurrentMap.Width || tileY >= s.world.CurrentMap.Height {
		return true
	}

	// Check if tile is walkable
	tile := s.world.CurrentMap.Tiles[tileY][tileX]
	return tile.Type == world.TileWall || tile.Type == world.TileDoor
}

// handleBounce reflects projectile velocity off a wall.
func (s *ProjectileSystem) handleBounce(entity *Entity, velComp *VelocityComponent, posComp *PositionComponent, oldX, oldY float64) {
	// Simple bounce: reverse velocity component that caused collision
	// More sophisticated: calculate normal and reflect properly
	// For simplicity, we'll just reverse both components for now
	velComp.VX = -velComp.VX
	velComp.VY = -velComp.VY

	// Reset position to before collision
	posComp.X = oldX
	posComp.Y = oldY
}

// checkEntityCollision checks if projectile hit any entity.
func (s *ProjectileSystem) checkEntityCollision(projEntity *Entity, posComp *PositionComponent, projComp *ProjectileComponent) *Entity {
	// Get all entities with position and health (potential targets)
	entities := s.world.GetEntitiesWithComponents("position", "health")

	for _, entity := range entities {
		// Skip self (owner)
		if entity.ID == projComp.OwnerID {
			continue
		}

		// Skip the projectile entity itself
		if entity.ID == projEntity.ID {
			continue
		}

		entityPos, ok := entity.GetComponent("position").(*PositionComponent)
		if !ok {
			continue
		}

		// Simple circle collision (assuming entities have ~16 pixel radius)
		dx := posComp.X - entityPos.X
		dy := posComp.Y - entityPos.Y
		distSq := dx*dx + dy*dy

		const collisionRadius = 16.0
		if distSq <= collisionRadius*collisionRadius {
			return entity
		}
	}

	return nil
}

// handleEntityHit processes damage and pierce logic when projectile hits entity.
func (s *ProjectileSystem) handleEntityHit(projEntity, hitEntity *Entity, projComp *ProjectileComponent, posComp *PositionComponent) {
	// Apply damage
	healthComp, ok := hitEntity.GetComponent("health").(*HealthComponent)
	if ok {
		healthComp.CurrentHealth -= projComp.Damage
		projComp.HasHit = true
	}

	// Handle explosion
	if projComp.Explosive {
		s.handleExplosion(projEntity, posComp)
	}

	// Check if projectile should be destroyed
	if projComp.DecrementPierce() {
		s.despawnProjectile(projEntity)
	}
}

// handleExplosion applies area damage around explosion point.
func (s *ProjectileSystem) handleExplosion(projEntity *Entity, posComp *PositionComponent) {
	projComp, ok := projEntity.GetComponent("projectile").(*ProjectileComponent)
	if !ok || !projComp.Explosive {
		return
	}

	// Get all entities within explosion radius
	entities := s.world.GetEntitiesWithComponents("position", "health")

	for _, entity := range entities {
		// Skip owner
		if entity.ID == projComp.OwnerID {
			continue
		}

		entityPos, ok := entity.GetComponent("position").(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance to explosion center
		dx := entityPos.X - posComp.X
		dy := entityPos.Y - posComp.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Apply damage based on distance (linear falloff)
		if dist <= projComp.ExplosionRadius {
			healthComp, ok := entity.GetComponent("health").(*HealthComponent)
			if ok {
				// Full damage at center, 0 at edge
				damageFactor := 1.0 - (dist / projComp.ExplosionRadius)
				damage := projComp.Damage * damageFactor
				healthComp.CurrentHealth -= damage
			}
		}
	}

	// TODO: Spawn explosion particle effect
	// TODO: Trigger screen shake if damage is significant
}

// despawnProjectile removes a projectile from the world.
func (s *ProjectileSystem) despawnProjectile(entity *Entity) {
	// Mark entity for removal
	// In a proper implementation, this would add to a removal queue
	// For now, we'll remove the projectile component to mark it as inactive
	if s.world != nil {
		s.world.RemoveEntity(entity.ID)
	}
}

// SpawnProjectile creates a new projectile entity in the world.
func (s *ProjectileSystem) SpawnProjectile(x, y, vx, vy float64, projComp *ProjectileComponent) *Entity {
	if s.world == nil {
		return nil
	}

	// Create new entity
	entity := s.world.CreateEntity()

	// Add position
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Add velocity
	entity.AddComponent(&VelocityComponent{VX: vx, VY: vy})

	// Add projectile component
	entity.AddComponent(projComp)

	// TODO: Add sprite component for visual representation

	return entity
}

// GetProjectileCount returns the number of active projectiles.
func (s *ProjectileSystem) GetProjectileCount() int {
	if s.world == nil {
		return 0
	}
	return len(s.world.GetEntitiesWithComponents("projectile"))
}
