package engine

import (
	"math"
)

// Phase 10.3: Screen shake and hit-stop configuration constants
const (
	// Combat shake parameters
	CombatShakeScaleFactor        = 10.0 // Multiplier for damage/maxHP ratio
	CombatShakeMinIntensity       = 1.0  // Minimum shake intensity (pixels)
	CombatShakeMaxIntensity       = 15.0 // Maximum shake intensity (pixels)
	CombatShakeBaseDuration       = 0.1  // Base shake duration (seconds)
	CombatShakeAdditionalDuration = 0.2  // Additional duration scaling (seconds)

	// Projectile shake parameters
	ProjectileShakeScaleFactor        = 8.0  // Multiplier for damage/maxHP ratio
	ProjectileShakeMinIntensity       = 0.5  // Minimum shake intensity (pixels)
	ProjectileShakeMaxIntensity       = 12.0 // Maximum shake intensity (pixels)
	ProjectileShakeBaseDuration       = 0.08 // Base shake duration (seconds)
	ProjectileShakeAdditionalDuration = 0.15 // Additional duration scaling (seconds)

	// Critical hit and explosion bonuses
	CriticalHitShakeMultiplier    = 1.5  // Intensity multiplier for critical hits
	CriticalHitDurationMultiplier = 1.3  // Duration multiplier for critical hits
	CriticalHitStopDuration       = 0.08 // Hit-stop duration for critical hits (seconds)
	ExplosionShakeMultiplier      = 1.5  // Intensity multiplier for explosions
	ExplosionDurationMultiplier   = 1.2  // Duration multiplier for explosions
	ExplosionHitStopDuration      = 0.06 // Hit-stop duration for explosions (seconds)
)

// ProjectileSystem manages projectile physics, collision detection, and lifecycle.
type ProjectileSystem struct {
	world *World
	// Quadtree for efficient spatial queries (optional, can be nil for simple collision)
	quadtree *Quadtree
	// Terrain collision checker for wall collision (optional)
	terrainChecker *TerrainCollisionChecker
	// Phase 10.3: Camera system for screen shake on projectile hits
	camera *CameraSystem
}

// NewProjectileSystem creates a new projectile system.
func NewProjectileSystem(w *World) *ProjectileSystem {
	return &ProjectileSystem{
		world:    w,
		quadtree: nil, // Initialize later if spatial partitioning is available
		camera:   nil, // Optional camera for visual feedback
	}
}

// SetQuadtree assigns a quadtree for efficient spatial collision detection.
func (s *ProjectileSystem) SetQuadtree(qt *Quadtree) {
	s.quadtree = qt
}

// SetTerrainChecker assigns a terrain collision checker for wall collision detection.
func (s *ProjectileSystem) SetTerrainChecker(checker *TerrainCollisionChecker) {
	s.terrainChecker = checker
}

// SetCamera sets the camera reference for screen shake feedback (Phase 10.3).
func (s *ProjectileSystem) SetCamera(camera *CameraSystem) {
	s.camera = camera
}

// Update processes all projectiles: movement, aging, collision detection.
func (s *ProjectileSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Get all projectile entities
	projectiles := s.world.GetEntitiesWith("projectile", "position", "velocity")

	for _, entity := range projectiles {
		s.updateProjectile(entity, deltaTime)
	}
}

// updateProjectile handles a single projectile's physics and collision.
func (s *ProjectileSystem) updateProjectile(entity *Entity, deltaTime float64) {
	projComp, ok := entity.GetComponent("projectile")
	if !ok {
		return
	}
	projComponent, ok := projComp.(*ProjectileComponent)
	if !ok {
		return
	}

	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	posComponent, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		return
	}
	velComponent, ok := velComp.(*VelocityComponent)
	if !ok {
		return
	}

	// Age the projectile
	projComponent.Age += deltaTime
	if projComponent.IsExpired() {
		s.despawnProjectile(entity)
		return
	}

	// Store old position for collision resolution
	oldX, oldY := posComponent.X, posComponent.Y

	// Move projectile
	posComponent.X += velComponent.VX * deltaTime
	posComponent.Y += velComponent.VY * deltaTime

	// Check wall collision
	if s.checkWallCollision(entity, oldX, oldY) {
		if projComponent.CanBounce() {
			s.handleBounce(entity, velComponent, posComponent, oldX, oldY)
			if projComponent.DecrementBounce() {
				// Handle explosion if explosive
				if projComponent.Explosive {
					s.handleExplosion(entity, posComponent)
				}
				s.despawnProjectile(entity)
			}
		} else {
			// Handle explosion if explosive
			if projComponent.Explosive {
				s.handleExplosion(entity, posComponent)
			}
			s.despawnProjectile(entity)
		}
		return
	}

	// Check entity collision
	hitEntity := s.checkEntityCollision(entity, posComponent, projComponent)
	if hitEntity != nil {
		s.handleEntityHit(entity, hitEntity, projComponent, posComponent)
	}
}

// checkWallCollision checks if projectile hit a wall.
func (s *ProjectileSystem) checkWallCollision(entity *Entity, oldX, oldY float64) bool {
	// If no terrain checker is set, skip wall collision
	if s.terrainChecker == nil {
		return false
	}

	posComp, ok := entity.GetComponent("position")
	if !ok {
		return false
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return false
	}

	// Use a small bounding box for the projectile
	const projectileSize = 4.0
	return s.terrainChecker.CheckCollision(pos.X, pos.Y, projectileSize, projectileSize)
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
	entities := s.world.GetEntitiesWith("position", "health")

	// DEBUG: Log collision check
	_ = entities // prevent unused warning if logging is disabled

	for _, entity := range entities {
		// Skip self (owner)
		if entity.ID == projComp.OwnerID {
			continue
		}

		// Skip the projectile entity itself
		if entity.ID == projEntity.ID {
			continue
		}

		entityPosComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entityPos, ok := entityPosComp.(*PositionComponent)
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
	healthComp, ok := hitEntity.GetComponent("health")
	if ok {
		health, ok := healthComp.(*HealthComponent)
		if ok {
			health.Current -= projComp.Damage
			projComp.HasHit = true

			// Phase 10.3: Trigger screen shake on projectile hit
			if s.camera != nil {
				// Calculate shake based on damage
				maxHP := health.Max
				shakeIntensity := CalculateShakeIntensity(projComp.Damage, maxHP,
					ProjectileShakeScaleFactor, ProjectileShakeMinIntensity, ProjectileShakeMaxIntensity)
				shakeDuration := CalculateShakeDuration(shakeIntensity,
					ProjectileShakeBaseDuration, ProjectileShakeAdditionalDuration, ProjectileShakeMaxIntensity)

				// Explosive projectiles get extra shake
				if projComp.Explosive {
					shakeIntensity *= ExplosionShakeMultiplier
					shakeDuration *= ExplosionDurationMultiplier
					// Trigger brief hit-stop for explosions
					s.camera.TriggerHitStop(ExplosionHitStopDuration, 0.0)
				}

				s.camera.ShakeAdvanced(shakeIntensity, shakeDuration)
			}
		}
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
	projComp, ok := projEntity.GetComponent("projectile")
	if !ok {
		return
	}
	proj, ok := projComp.(*ProjectileComponent)
	if !ok || !proj.Explosive {
		return
	}

	// Get all entities within explosion radius
	entities := s.world.GetEntitiesWith("position", "health")

	for _, entity := range entities {
		// Skip owner
		if entity.ID == proj.OwnerID {
			continue
		}

		entityPosComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entityPos, ok := entityPosComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance to explosion center
		dx := entityPos.X - posComp.X
		dy := entityPos.Y - posComp.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Apply damage based on distance (linear falloff)
		if dist <= proj.ExplosionRadius {
			healthComp, ok := entity.GetComponent("health")
			if ok {
				health, ok := healthComp.(*HealthComponent)
				if ok {
					// Full damage at center, 0 at edge
					damageFactor := 1.0 - (dist / proj.ExplosionRadius)
					damage := proj.Damage * damageFactor
					health.Current -= damage
				}
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
	return len(s.world.GetEntitiesWith("projectile"))
}
