// Package engine provides the ProjectileTrailParticleSystem for genre-aware projectile trails.
// This system connects ProjectileComponent with ParticleSystem to spawn trailing particles
// behind moving projectiles based on their type (arrow, fireball, ice_shard, bullet),
// with genre-aware particle colors and behaviors.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ProjectileTrailParticleSystem spawns trailing particles behind moving projectiles.
// It reads ProjectileComponent.ProjectileType to select appropriate particle effects
// and uses genre-aware coloring for visual consistency.
type ProjectileTrailParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Per-entity cooldown to limit trail spawn rate
	lastSpawn     map[uint64]float64
	spawnInterval float64 // seconds between trail particles per projectile
}

// NewProjectileTrailParticleSystem creates a new projectile trail particle system.
func NewProjectileTrailParticleSystem(world *World, seed int64) *ProjectileTrailParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "projectile_trail_particle")
		logEntry.Debug("projectile trail particle system created")
	}

	return &ProjectileTrailParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		spawnInterval: 0.05, // 50ms between trail particles
		lastSpawn:     make(map[uint64]float64, 64),
	}
}

// SetParticleSystem sets the particle system used for spawning trail effects.
func (s *ProjectileTrailParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ProjectileTrailParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes projectile entities and spawns trailing particles.
func (s *ProjectileTrailParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}

	// Clean up despawned projectile cooldowns periodically
	if len(s.lastSpawn) > 256 {
		s.cleanupCooldowns(entities)
	}
}

// processEntity handles trail spawning for a single projectile entity.
func (s *ProjectileTrailParticleSystem) processEntity(entity *Entity, deltaTime float64) {
	if !entity.HasComponent("projectile") {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	comp, ok := entity.GetComponent("projectile")
	if !ok {
		return
	}
	proj, ok := comp.(*ProjectileComponent)
	if !ok || proj == nil {
		return
	}

	// Don't trail expired or hit projectiles
	if proj.IsExpired() || proj.HasHit {
		delete(s.lastSpawn, entity.ID)
		return
	}

	// Check cooldown
	elapsed := s.lastSpawn[entity.ID] + deltaTime
	if elapsed < s.spawnInterval {
		s.lastSpawn[entity.ID] = elapsed
		return
	}
	s.lastSpawn[entity.ID] = 0

	config := s.getTrailConfig(proj.ProjectileType, pos.X, pos.Y)
	if config == nil {
		return
	}

	s.particleSystem.SpawnParticles(s.world, *config, pos.X, pos.Y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"projectile_type": proj.ProjectileType,
			"x":               pos.X,
			"y":               pos.Y,
		}).Debug("projectile trail particles spawned")
	}
}

// getTrailConfig returns particle configuration based on projectile type.
func (s *ProjectileTrailParticleSystem) getTrailConfig(projectileType string, x, y float64) *particles.Config {
	effectSeed := s.seed + int64(x*73) + int64(y*97)

	switch projectileType {
	case "fireball":
		return s.getFireballTrail(effectSeed)
	case "ice_shard":
		return s.getIceTrail(effectSeed)
	case "arrow":
		return s.getArrowTrail(effectSeed)
	case "bullet":
		return s.getBulletTrail(effectSeed)
	default:
		return s.getDefaultTrail(effectSeed)
	}
}

// getFireballTrail returns ember/flame trail for fire projectiles.
func (s *ProjectileTrailParticleSystem) getFireballTrail(seed int64) *particles.Config {
	pType := particles.ParticleEmber
	if s.genreID == "horror" {
		pType = particles.ParticleSmoke
	}
	return &particles.Config{
		Type:     pType,
		Count:    3,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.3,
		SpreadX:  6.0,
		SpreadY:  6.0,
		Gravity:  -15.0, // Embers rise
		MinSize:  2.0,
		MaxSize:  5.0,
		ZLayer:   particles.ZLayerEntity,
		Custom:   map[string]interface{}{"projectile_trail": true, "trail_type": "fireball"},
	}
}

// getIceTrail returns sparkle trail for ice projectiles.
func (s *ProjectileTrailParticleSystem) getIceTrail(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.35,
		SpreadX:  5.0,
		SpreadY:  5.0,
		Gravity:  10.0, // Slight downward drift
		MinSize:  1.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerEntity,
		Custom:   map[string]interface{}{"projectile_trail": true, "trail_type": "ice"},
	}
}

// getArrowTrail returns subtle dust trail for arrow projectiles.
func (s *ProjectileTrailParticleSystem) getArrowTrail(seed int64) *particles.Config {
	count := 1
	if s.genreID == "fantasy" {
		count = 2
	}
	return &particles.Config{
		Type:     particles.ParticleDust,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.2,
		SpreadX:  4.0,
		SpreadY:  3.0,
		Gravity:  20.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"projectile_trail": true, "trail_type": "arrow"},
	}
}

// getBulletTrail returns smoke trail for bullet projectiles.
func (s *ProjectileTrailParticleSystem) getBulletTrail(seed int64) *particles.Config {
	pType := particles.ParticleSmoke
	if s.genreID == "cyberpunk" || s.genreID == "scifi" {
		pType = particles.ParticleSpark
	}
	return &particles.Config{
		Type:     pType,
		Count:    1,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.15,
		SpreadX:  3.0,
		SpreadY:  3.0,
		Gravity:  5.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerEntity,
		Custom:   map[string]interface{}{"projectile_trail": true, "trail_type": "bullet"},
	}
}

// getDefaultTrail returns a generic magic trail for unknown projectile types.
func (s *ProjectileTrailParticleSystem) getDefaultTrail(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleMagic,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.25,
		SpreadX:  5.0,
		SpreadY:  5.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerEntity,
		Custom:   map[string]interface{}{"projectile_trail": true, "trail_type": "magic"},
	}
}

// cleanupCooldowns removes entries for entities no longer in the active set.
func (s *ProjectileTrailParticleSystem) cleanupCooldowns(entities []*Entity) {
	active := make(map[uint64]struct{}, len(entities))
	for _, e := range entities {
		active[e.ID] = struct{}{}
	}
	for id := range s.lastSpawn {
		if _, ok := active[id]; !ok {
			delete(s.lastSpawn, id)
		}
	}
}
