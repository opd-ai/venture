package engine

import (
	"image/color"
	"math"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
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
	// Particle generator for explosion effects
	particleGenerator *particles.Generator
	// Genre ID and seed for sprite/particle generation
	genreID string
	seed    int64
}

// NewProjectileSystem creates a new projectile system.
func NewProjectileSystem(w *World) *ProjectileSystem {
	return &ProjectileSystem{
		world:             w,
		quadtree:          nil, // Initialize later if spatial partitioning is available
		camera:            nil, // Optional camera for visual feedback
		particleGenerator: particles.NewGenerator(),
		genreID:           "fantasy", // Default genre
		seed:              12345,     // Default seed
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

// SetGenre sets the genre ID for visual generation.
func (s *ProjectileSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// SetSeed sets the seed for deterministic generation.
func (s *ProjectileSystem) SetSeed(seed int64) {
	s.seed = seed
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

	// Phase 10.2: Spawn explosion particle effect
	s.spawnExplosionParticles(posComp.X, posComp.Y, proj.ExplosionRadius)

	// Phase 10.2: Trigger screen shake for explosion
	if s.camera != nil {
		// Use a substantial shake for explosions
		shakeIntensity := 8.0 + (proj.ExplosionRadius / 20.0) // Scale with explosion radius
		if shakeIntensity > ExplosionShakeMaxIntensity {
			shakeIntensity = ExplosionShakeMaxIntensity
		}
		shakeDuration := 0.3 // Fixed duration for explosions
		s.camera.ShakeAdvanced(shakeIntensity, shakeDuration)
	}
}

// ExplosionShakeMaxIntensity is maximum shake intensity for explosions
const ExplosionShakeMaxIntensity = 15.0

// spawnExplosionParticles creates a particle effect at the explosion location.
func (s *ProjectileSystem) spawnExplosionParticles(x, y, radius float64) {
	if s.particleGenerator == nil || s.world == nil {
		return
	}

	// Calculate particle count based on explosion radius
	// Larger explosions have more particles
	particleCount := int(20 + radius/5.0)
	if particleCount > 100 {
		particleCount = 100 // Cap at 100 particles
	}

	// Create particle configuration for explosion
	config := particles.Config{
		Type:     particles.ParticleSpark, // Bright spark particles for explosion
		Count:    particleCount,
		GenreID:  s.genreID,
		Seed:     s.seed + int64(x+y), // Vary seed based on position
		Duration: 0.5,                 // Particles last 0.5 seconds
		SpreadX:  radius * 2.0,        // Radial spread based on explosion radius
		SpreadY:  radius * 2.0,
		Gravity:  50.0, // Slight downward gravity
		MinSize:  2.0,
		MaxSize:  6.0,
	}

	// Generate particle system
	particleSystem, err := s.particleGenerator.Generate(config)
	if err != nil {
		// Failed to generate particles, continue without them
		return
	}

	// Position particles at explosion center
	for i := range particleSystem.Particles {
		particleSystem.Particles[i].X += x
		particleSystem.Particles[i].Y += y
	}

	// Create explosion entity with particle emitter
	explosionEntity := s.world.CreateEntity()
	explosionEntity.AddComponent(&PositionComponent{X: x, Y: y})

	// Create one-shot particle emitter (EmitRate = 0)
	emitter := NewParticleEmitterComponent(0, config, 1)
	emitter.AddSystem(particleSystem)
	explosionEntity.AddComponent(emitter)
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

	// Phase 10.2: Add sprite component for visual representation
	spriteSize := 8 // Default projectile sprite size (8x8 pixels)
	if projComp.Explosive {
		spriteSize = 12 // Larger sprite for explosive projectiles
	}

	// Generate procedural sprite using seed for deterministic generation
	spriteSeed := s.seed + int64(entity.ID)
	projectileType := projComp.ProjectileType
	if projectileType == "" {
		projectileType = "bullet" // Default type
	}

	spriteImage := sprites.GenerateProjectileSprite(spriteSeed, projectileType, s.genreID, spriteSize)

	// Create sprite component with generated image
	spriteComp := NewSpriteComponent(float64(spriteSize), float64(spriteSize), color.RGBA{255, 255, 255, 255})
	spriteComp.Image = spriteImage

	// Calculate rotation from velocity for proper orientation
	rotation := math.Atan2(vy, vx)
	spriteComp.Rotation = rotation

	entity.AddComponent(spriteComp)

	return entity
}

// GetProjectileCount returns the number of active projectiles.
func (s *ProjectileSystem) GetProjectileCount() int {
	if s.world == nil {
		return 0
	}
	return len(s.world.GetEntitiesWith("projectile"))
}
