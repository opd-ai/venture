// Package engine provides the CompanionAuraParticleSystem for visual feedback
// when companions with active bonding perks are near their owners.
// This connects CompanionComponent bonding perks with ParticleSystem to spawn
// genre-aware aura particles indicating active perk effects.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// CompanionAuraParticleSystem spawns subtle aura particles around companions
// that have active bonding perks. It provides visual feedback when companions
// are providing bonuses to their owners through unlocked perks.
type CompanionAuraParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Pulse timing for aura effects
	pulseInterval float64 // Seconds between aura pulses
	timeSinceEmit float64 // Accumulator

	// Particle configuration
	baseParticleCount int
	spreadFactor      float64
	effectDuration    float64

	// Distance threshold - only show aura when near owner
	auraDistanceMax float64
}

// NewCompanionAuraParticleSystem creates a new companion aura particle system.
func NewCompanionAuraParticleSystem(world *World, seed int64) *CompanionAuraParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_aura_particle")
		logEntry.Debug("companion aura particle system created")
	}

	return &CompanionAuraParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		pulseInterval:     1.5, // Pulse every 1.5 seconds (subtle)
		timeSinceEmit:     0.0,
		baseParticleCount: 6,     // Fewer particles for subtle effect
		spreadFactor:      40.0,  // Smaller radius than combat effects
		effectDuration:    1.2,   // Lingering glow
		auraDistanceMax:   150.0, // Only show when within 150px of owner
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *CompanionAuraParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *CompanionAuraParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes companions and spawns aura particles for those with active perks.
func (s *CompanionAuraParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.pulseInterval {
		return
	}

	for _, entity := range entities {
		if !entity.HasComponent("companion") {
			continue
		}

		companionComp := s.getCompanionComponent(entity)
		if companionComp == nil {
			continue
		}

		// Check if companion has any perks
		if len(companionComp.BondingPerks) == 0 {
			continue
		}

		// Get companion position
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Check if near owner
		owner, _ := s.world.GetEntity(companionComp.OwnerID)
		if owner == nil {
			continue
		}
		ownerPos := owner.GetPosition()
		if ownerPos == nil {
			continue
		}

		// Calculate distance to owner
		dx := ownerPos.X - pos.X
		dy := ownerPos.Y - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Only show aura when close to owner (perks are active)
		if distance > s.auraDistanceMax {
			continue
		}

		// Spawn aura particles based on perks
		s.spawnAuraParticles(entity.ID, pos.X, pos.Y, companionComp.BondingPerks)
	}

	// Reset timer after processing
	s.timeSinceEmit = 0
}

// getCompanionComponent safely extracts the companion component.
func (s *CompanionAuraParticleSystem) getCompanionComponent(entity *Entity) *CompanionComponent {
	comp, ok := entity.GetComponent("companion")
	if !ok {
		return nil
	}
	cc, ok := comp.(*CompanionComponent)
	if !ok {
		return nil
	}
	return cc
}

// spawnAuraParticles creates subtle aura effects around companion based on perks.
func (s *CompanionAuraParticleSystem) spawnAuraParticles(entityID uint64, x, y float64, perks []BondingPerk) {
	effectSeed := s.seed + int64(entityID*1000) + int64(x) + int64(y)

	// Determine dominant perk type for particle style
	perkType := s.getDominantPerkType(perks)
	particleType := s.getParticleTypeForPerk(perkType)

	// Scale particle count by number of perks (more perks = stronger aura)
	count := s.baseParticleCount + len(perks)*2
	if count > 20 {
		count = 20
	}

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -15.0, // Gentle upward drift
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom: map[string]interface{}{
			"companion_aura": true,
			"perk_count":     len(perks),
		},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entityID,
			"perk_count":    len(perks),
			"dominant_perk": perkType.String(),
			"x":             x,
			"y":             y,
		}).Debug("companion aura particles spawned")
	}
}

// getDominantPerkType returns the highest-tier perk for visual style.
func (s *CompanionAuraParticleSystem) getDominantPerkType(perks []BondingPerk) BondingPerk {
	if len(perks) == 0 {
		return PerkNone
	}

	// Higher perks take visual priority
	priority := map[BondingPerk]int{
		PerkAutoRevive:       6,
		PerkSharedExperience: 5,
		PerkLoyalGuard:       4,
		PerkFasterLearning:   3,
		PerkExtraDamage:      2,
		PerkExtraHealth:      1,
		PerkNone:             0,
	}

	var dominant BondingPerk
	maxPriority := -1

	for _, perk := range perks {
		if p, ok := priority[perk]; ok && p > maxPriority {
			maxPriority = p
			dominant = perk
		}
	}

	return dominant
}

// getParticleTypeForPerk returns genre-aware particle type based on perk.
func (s *CompanionAuraParticleSystem) getParticleTypeForPerk(perk BondingPerk) particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return s.getFantasyPerkParticle(perk)
	case "scifi":
		return s.getSciFiPerkParticle(perk)
	case "horror":
		return s.getHorrorPerkParticle(perk)
	case "cyberpunk":
		return s.getCyberpunkPerkParticle(perk)
	case "postapoc":
		return s.getPostApocPerkParticle(perk)
	default:
		return s.getFantasyPerkParticle(perk)
	}
}

// getFantasyPerkParticle returns fantasy-themed particle for perk type.
func (s *CompanionAuraParticleSystem) getFantasyPerkParticle(perk BondingPerk) particles.ParticleType {
	switch perk {
	case PerkExtraHealth:
		return particles.ParticleSparkle // Golden health sparkles
	case PerkExtraDamage:
		return particles.ParticleEmber // Fiery damage glow
	case PerkFasterLearning:
		return particles.ParticleMagic // Arcane learning
	case PerkLoyalGuard:
		return particles.ParticleSparkle // Shield sparkles
	case PerkSharedExperience:
		return particles.ParticleMagic // Magical bond
	case PerkAutoRevive:
		return particles.ParticleMagic // Life magic
	default:
		return particles.ParticleSparkle
	}
}

// getSciFiPerkParticle returns sci-fi themed particle for perk type.
func (s *CompanionAuraParticleSystem) getSciFiPerkParticle(perk BondingPerk) particles.ParticleType {
	switch perk {
	case PerkExtraHealth:
		return particles.ParticleSpark // Energy shield
	case PerkExtraDamage:
		return particles.ParticleSpark // Weapon boost
	case PerkFasterLearning:
		return particles.ParticleMagic // Data stream hologram
	case PerkLoyalGuard:
		return particles.ParticleSpark // Force field
	case PerkSharedExperience:
		return particles.ParticleMagic // Neural link
	case PerkAutoRevive:
		return particles.ParticleSpark // Revive beacon
	default:
		return particles.ParticleSpark
	}
}

// getHorrorPerkParticle returns horror-themed particle for perk type.
func (s *CompanionAuraParticleSystem) getHorrorPerkParticle(perk BondingPerk) particles.ParticleType {
	switch perk {
	case PerkExtraHealth:
		return particles.ParticleSmoke // Life mist
	case PerkExtraDamage:
		return particles.ParticleEmber // Dark embers
	case PerkFasterLearning:
		return particles.ParticleSmoke // Eerie knowledge
	case PerkLoyalGuard:
		return particles.ParticleSmoke // Shadow ward
	case PerkSharedExperience:
		return particles.ParticleSmoke // Soul bond
	case PerkAutoRevive:
		return particles.ParticleSmoke // Death's embrace
	default:
		return particles.ParticleSmoke
	}
}

// getCyberpunkPerkParticle returns cyberpunk-themed particle for perk type.
func (s *CompanionAuraParticleSystem) getCyberpunkPerkParticle(perk BondingPerk) particles.ParticleType {
	switch perk {
	case PerkExtraHealth:
		return particles.ParticleSpark // Nano repair
	case PerkExtraDamage:
		return particles.ParticleEmber // Overclocked weapon
	case PerkFasterLearning:
		return particles.ParticleSpark // Data download
	case PerkLoyalGuard:
		return particles.ParticleSpark // Defensive matrix
	case PerkSharedExperience:
		return particles.ParticleSpark // Neural sync
	case PerkAutoRevive:
		return particles.ParticleSpark // Backup restore
	default:
		return particles.ParticleSpark
	}
}

// getPostApocPerkParticle returns post-apocalyptic themed particle for perk type.
func (s *CompanionAuraParticleSystem) getPostApocPerkParticle(perk BondingPerk) particles.ParticleType {
	switch perk {
	case PerkExtraHealth:
		return particles.ParticleDust // Survival grit
	case PerkExtraDamage:
		return particles.ParticleEmber // Scavenger fire
	case PerkFasterLearning:
		return particles.ParticleDust // Wasteland wisdom
	case PerkLoyalGuard:
		return particles.ParticleDust // Protective dust
	case PerkSharedExperience:
		return particles.ParticleDust // Shared struggle
	case PerkAutoRevive:
		return particles.ParticleDust // Second chance
	default:
		return particles.ParticleDust
	}
}
