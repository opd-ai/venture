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
	projComponent, posComponent, velComponent := s.getProjectileComponents(entity)
	if projComponent == nil {
		return
	}

	if s.updateProjectileAge(entity, projComponent, deltaTime) {
		return
	}

	oldX, oldY := posComponent.X, posComponent.Y
	s.moveProjectile(posComponent, velComponent, deltaTime)
	s.spawnTrailParticles(entity, posComponent, deltaTime)

	if s.handleWallCollisionLogic(entity, projComponent, velComponent, posComponent, oldX, oldY) {
		return
	}

	hitEntity := s.checkEntityCollision(entity, posComponent, projComponent)
	if hitEntity != nil {
		s.handleEntityHit(entity, hitEntity, projComponent, posComponent)
	}
}

// getProjectileComponents retrieves and validates all required components for a projectile.
func (s *ProjectileSystem) getProjectileComponents(entity *Entity) (*ProjectileComponent, *PositionComponent, *VelocityComponent) {
	// Use typed getters for position and velocity (~93x faster than map lookup + type assertion)
	posComponent := entity.GetPosition()
	velComponent := entity.GetVelocity()
	if posComponent == nil || velComponent == nil {
		return nil, nil, nil
	}

	// Projectile component uses generic access (not a hot path component)
	projComp, ok := entity.GetComponent("projectile")
	if !ok {
		return nil, nil, nil
	}
	projComponent, ok := projComp.(*ProjectileComponent)
	if !ok {
		return nil, nil, nil
	}

	return projComponent, posComponent, velComponent
}

// updateProjectileAge ages the projectile and despawns if expired.
// Returns true if projectile was despawned.
func (s *ProjectileSystem) updateProjectileAge(entity *Entity, projComponent *ProjectileComponent, deltaTime float64) bool {
	projComponent.Age += deltaTime
	if projComponent.IsExpired() {
		s.despawnProjectile(entity)
		return true
	}
	return false
}

// moveProjectile updates the projectile's position based on velocity.
func (s *ProjectileSystem) moveProjectile(posComponent *PositionComponent, velComponent *VelocityComponent, deltaTime float64) {
	posComponent.X += velComponent.VX * deltaTime
	posComponent.Y += velComponent.VY * deltaTime
}

// handleWallCollisionLogic processes wall collision, bouncing, and explosion.
// Returns true if projectile should stop processing.
func (s *ProjectileSystem) handleWallCollisionLogic(entity *Entity, projComponent *ProjectileComponent,
	velComponent *VelocityComponent, posComponent *PositionComponent, oldX, oldY float64,
) bool {
	if !s.checkWallCollision(entity, oldX, oldY) {
		return false
	}

	if projComponent.CanBounce() {
		s.handleBounce(entity, velComponent, posComponent, oldX, oldY)
		if projComponent.DecrementBounce() {
			s.handleExplosionAndDespawn(entity, projComponent, posComponent)
		}
	} else {
		s.handleExplosionAndDespawn(entity, projComponent, posComponent)
	}
	return true
}

// handleExplosionAndDespawn handles explosion effect if projectile is explosive, then despawns.
func (s *ProjectileSystem) handleExplosionAndDespawn(entity *Entity, projComponent *ProjectileComponent, posComponent *PositionComponent) {
	if projComponent.Explosive {
		s.handleExplosion(entity, posComponent)
	}
	s.despawnProjectile(entity)
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

	for _, entity := range entities {
		// Skip self (owner)
		if entity.ID == projComp.OwnerID {
			continue
		}

		// Skip the projectile entity itself
		if entity.ID == projEntity.ID {
			continue
		}

		// Use typed getter for zero-overhead access (eliminates map lookup + type assertion)
		entityPos := entity.GetPosition()
		if entityPos == nil {
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
	// Apply damage using typed getter for zero-overhead access
	health := hitEntity.GetHealth()
	if health != nil {
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
	proj := s.getExplosiveProjectile(projEntity)
	if proj == nil {
		return
	}

	s.applyExplosionDamage(posComp, proj)
	s.spawnExplosionParticles(posComp.X, posComp.Y, proj.ExplosionRadius)
	s.triggerExplosionScreenShake(proj.ExplosionRadius)
}

// getExplosiveProjectile retrieves the projectile component if it's explosive.
func (s *ProjectileSystem) getExplosiveProjectile(projEntity *Entity) *ProjectileComponent {
	projComp, ok := projEntity.GetComponent("projectile")
	if !ok {
		return nil
	}
	proj, ok := projComp.(*ProjectileComponent)
	if !ok || !proj.Explosive {
		return nil
	}
	return proj
}

// applyExplosionDamage applies damage to all entities within explosion radius.
func (s *ProjectileSystem) applyExplosionDamage(posComp *PositionComponent, proj *ProjectileComponent) {
	entities := s.world.GetEntitiesWith("position", "health")

	for _, entity := range entities {
		if entity.ID == proj.OwnerID {
			continue
		}

		entityPos := s.getEntityPosition(entity)
		if entityPos == nil {
			continue
		}

		s.damageEntityFromExplosion(entity, entityPos, posComp, proj)
	}
}

// getEntityPosition retrieves the position component from an entity.
// Uses typed getter for zero-overhead access (eliminates map lookup + type assertion).
func (s *ProjectileSystem) getEntityPosition(entity *Entity) *PositionComponent {
	return entity.GetPosition()
}

// damageEntityFromExplosion calculates and applies explosion damage to an entity.
func (s *ProjectileSystem) damageEntityFromExplosion(entity *Entity, entityPos, explosionPos *PositionComponent, proj *ProjectileComponent) {
	dx := entityPos.X - explosionPos.X
	dy := entityPos.Y - explosionPos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > proj.ExplosionRadius {
		return
	}

	// Use typed getter for zero-overhead access
	health := entity.GetHealth()
	if health == nil {
		return
	}

	damageFactor := 1.0 - (dist / proj.ExplosionRadius)
	damage := proj.Damage * damageFactor
	health.Current -= damage
}

// triggerExplosionScreenShake triggers camera shake for explosion effect.
func (s *ProjectileSystem) triggerExplosionScreenShake(explosionRadius float64) {
	if s.camera == nil {
		return
	}

	shakeIntensity := 8.0 + (explosionRadius / 20.0)
	if shakeIntensity > ExplosionShakeMaxIntensity {
		shakeIntensity = ExplosionShakeMaxIntensity
	}
	shakeDuration := 0.3
	s.camera.ShakeAdvanced(shakeIntensity, shakeDuration)
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

	// Phase 6: Add trail component for visual enhancement
	// Determine trail type based on projectile type
	var trail *TrailComponent
	switch projComp.ProjectileType {
	case "fireball", "magic_missile", "ice_shard", "lightning_bolt":
		// Magical projectiles get glowing trails
		magicColor := spriteComp.Color.(color.RGBA)
		trail = NewMagicTrailComponent(&magicColor)
	case "arrow", "bullet", "bolt":
		// Physical projectiles get subtle trails
		trail = NewPhysicalTrailComponent()
	default:
		// Default trail for unknown types
		trail = NewTrailComponent()
	}
	entity.AddComponent(trail)

	return entity
}

// GetProjectileCount returns the number of active projectiles.
func (s *ProjectileSystem) GetProjectileCount() int {
	if s.world == nil {
		return 0
	}
	return len(s.world.GetEntitiesWith("projectile"))
}

// spawnTrailParticles spawns particle trail effects for projectiles with TrailComponent.
// This creates visually appealing trails that enhance projectile visibility and game feel.
func (s *ProjectileSystem) spawnTrailParticles(entity *Entity, pos *PositionComponent, deltaTime float64) {
	// Check if entity has a trail component
	trailComp, ok := entity.GetComponent("trail")
	if !ok {
		return
	}
	trail, ok := trailComp.(*TrailComponent)
	if !ok || !trail.Enabled {
		return
	}

	// Update spawn timer
	trail.TimeSinceLastSpawn += deltaTime

	// Calculate spawn interval from spawn rate
	spawnInterval := 1.0 / trail.SpawnRate

	// Spawn particles if enough time has passed
	particlesSpawned := 0
	for trail.TimeSinceLastSpawn >= spawnInterval {
		trail.TimeSinceLastSpawn -= spawnInterval
		particlesSpawned++

		// Safety limit: don't spawn too many particles in one frame
		if particlesSpawned > 10 {
			trail.TimeSinceLastSpawn = 0
			break
		}

		s.spawnTrailParticle(entity, pos, trail)
	}
}

// spawnTrailParticle spawns a single trail particle at the projectile's position.
func (s *ProjectileSystem) spawnTrailParticle(entity *Entity, pos *PositionComponent, trail *TrailComponent) {
	if s.world == nil {
		return
	}

	// Determine particle color
	particleColor := color.RGBA{R: 200, G: 200, B: 200, A: 255} // Default gray
	if trail.Color != nil {
		particleColor = *trail.Color
	} else {
		// Try to get color from sprite component
		if spriteComp, ok := entity.GetComponent("sprite"); ok {
			if sprite, ok := spriteComp.(*EbitenSprite); ok {
				if rgba, ok := sprite.Color.(color.RGBA); ok {
					particleColor = rgba
				}
			}
		}
	}

	// Create particle config
	config := particles.Config{
		Type:     particles.ParticleSparkle, // Sparkle type for trails
		Count:    1,
		GenreID:  s.genreID,
		Seed:     s.seed + int64(entity.ID),
		Duration: trail.ParticleLifetime,
		SpreadX:  trail.SpreadX,
		SpreadY:  trail.SpreadY,
		Gravity:  0.0, // No gravity for projectile trails
		MinSize:  trail.ParticleSize,
		MaxSize:  trail.ParticleSize,
		Custom:   make(map[string]interface{}),
	}

	// Generate particle system directly (particle generator returns *ParticleSystem, not interface{})
	particleSystem, err := s.particleGenerator.Generate(config)
	if err != nil {
		return
	}

	// Position particle at projectile location
	if len(particleSystem.Particles) > 0 {
		particleSystem.Particles[0].X = pos.X
		particleSystem.Particles[0].Y = pos.Y
		particleSystem.Particles[0].Color = particleColor

		// Set fade rate based on trail settings
		particleSystem.Particles[0].Life = trail.FadeRate
	}

	// Create particle entity
	particleEntity := s.world.CreateEntity()
	particleEntity.AddComponent(&PositionComponent{X: pos.X, Y: pos.Y})

	// Add particle emitter component (one-shot)
	emitter := NewParticleEmitterComponent(0, config, 1)
	emitter.AddSystem(particleSystem)
	particleEntity.AddComponent(emitter)
}
