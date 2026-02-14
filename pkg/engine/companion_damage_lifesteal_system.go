// Package engine provides the CompanionDamageLifestealSystem for healing owners
// when their companions deal physical damage. This connects CompanionComponent
// (loyalty, bonding perks) with CombatSystem damage events and HealthComponent
// healing, creating a sustain synergy for pet-based builds.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// CompanionDamageLifestealSystem heals companion owners for a percentage of
// damage their companions deal. The heal amount scales with companion loyalty
// and bonding perks (PerkExtraHealth increases lifesteal).
//
// Base lifesteal: 0.05% per loyalty point (max 5% at 100 loyalty)
// PerkExtraHealth bonus: +3% lifesteal
// PerkSharedExperience bonus: +2% lifesteal
//
// Genre modifiers:
//   - Fantasy: 1.3x (magical bonds)
//   - Scifi: 0.7x (mechanical companions)
//   - Horror: 1.5x (dark pacts)
//   - Cyberpunk: 0.8x (augmented companions)
//   - Postapoc: 1.1x (survival bonds)
type CompanionDamageLifestealSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Base lifesteal per loyalty point (0-100)
	baseLifestealPerLoyalty float64

	// Bonus lifesteal from perks
	perkExtraHealthBonus float64
	perkSharedExpBonus   float64

	// Genre-specific multipliers
	genreMultipliers map[string]float64

	// Particle configuration
	particleCount int
	spreadFactor  float64
}

// NewCompanionDamageLifestealSystem creates a new companion damage lifesteal system.
func NewCompanionDamageLifestealSystem(world *World, seed int64) *CompanionDamageLifestealSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_damage_lifesteal")
		logEntry.Debug("companion damage lifesteal system created")
	}

	return &CompanionDamageLifestealSystem{
		world:                   world,
		seed:                    seed,
		rng:                     rand.New(rand.NewSource(seed)),
		logger:                  logEntry,
		baseLifestealPerLoyalty: 0.0005, // 0.05% per loyalty = max 5% at 100
		perkExtraHealthBonus:    0.03,   // +3% from PerkExtraHealth
		perkSharedExpBonus:      0.02,   // +2% from PerkSharedExperience
		genreMultipliers: map[string]float64{
			"fantasy":   1.3,
			"scifi":     0.7,
			"horror":    1.5,
			"cyberpunk": 0.8,
			"postapoc":  1.1,
		},
		particleCount: 6,
		spreadFactor:  45.0,
	}
}

// SetParticleSystem sets the particle system for spawning heal effects.
func (s *CompanionDamageLifestealSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors and multipliers.
func (s *CompanionDamageLifestealSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *CompanionDamageLifestealSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnCompanionDamageDealt
}

// OnCompanionDamageDealt is called when any entity deals damage; it checks if
// the attacker is a companion and heals the owner accordingly.
// Register this with CombatSystem.AddDamageCallback().
func (s *CompanionDamageLifestealSystem) OnCompanionDamageDealt(attacker, target *Entity, damage float64) {
	if attacker == nil || damage <= 0 || s.world == nil {
		return
	}

	// Check if attacker is a companion
	compComp, hasComp := attacker.GetComponent("companion")
	if !hasComp || compComp == nil {
		return
	}

	companion, ok := compComp.(*CompanionComponent)
	if !ok || companion.OwnerID == 0 {
		return
	}

	// Find the owner entity
	owner := s.world.GetEntity(companion.OwnerID)
	if owner == nil {
		return
	}

	// Get owner's health component
	healthComp, hasHealth := owner.GetComponent("health")
	if !hasHealth || healthComp == nil {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Calculate lifesteal percentage based on loyalty
	lifestealPercent := companion.Loyalty * s.baseLifestealPerLoyalty

	// Add perk bonuses
	if companion.HasPerk(PerkExtraHealth) {
		lifestealPercent += s.perkExtraHealthBonus
	}
	if companion.HasPerk(PerkSharedExperience) {
		lifestealPercent += s.perkSharedExpBonus
	}

	// Apply genre multiplier
	if mult, exists := s.genreMultipliers[s.genreID]; exists {
		lifestealPercent *= mult
	}

	// Cap at 15% max lifesteal
	if lifestealPercent > 0.15 {
		lifestealPercent = 0.15
	}

	// Calculate heal amount
	healAmount := damage * lifestealPercent

	// Cap healing at 15% of owner's max health per hit
	maxHealPerHit := health.Max * 0.15
	if healAmount > maxHealPerHit {
		healAmount = maxHealPerHit
	}

	// Skip tiny heals
	if healAmount < 0.1 {
		return
	}

	// Apply healing
	oldHealth := health.Current
	health.Heal(healAmount)
	actualHeal := health.Current - oldHealth

	// Spawn particles if meaningful healing occurred
	if actualHeal > 0.5 {
		s.spawnHealParticles(owner, actualHeal)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"companion_id":  attacker.ID,
			"owner_id":      owner.ID,
			"damage":        damage,
			"loyalty":       companion.Loyalty,
			"lifesteal_pct": lifestealPercent,
			"heal_amount":   healAmount,
			"actual_healed": actualHeal,
		}).Debug("companion lifesteal applied")
	}
}

// spawnHealParticles creates healing particle effects at owner location.
func (s *CompanionDamageLifestealSystem) spawnHealParticles(owner *Entity, healAmount float64) {
	if s.particleSystem == nil {
		return
	}

	pos := owner.GetPosition()
	if pos == nil {
		return
	}

	// Scale particle count based on heal amount
	count := s.particleCount
	if healAmount > 10 {
		count = int(float64(count) * 1.3)
	}
	if healAmount > 25 {
		count = int(float64(count) * 1.5)
	}
	if count > 15 {
		count = 15
	}

	// Deterministic seed for this heal event
	seedOffset := int64(owner.ID) + int64(healAmount*100)
	localRng := rand.New(rand.NewSource(s.seed + seedOffset))

	// Genre-aware particle colors
	var particleColor particles.ParticleColor
	switch s.genreID {
	case "fantasy":
		particleColor = particles.ParticleColor{R: 0.3, G: 1.0, B: 0.5, A: 0.9} // Green healing
	case "scifi":
		particleColor = particles.ParticleColor{R: 0.2, G: 0.8, B: 1.0, A: 0.9} // Cyan nanites
	case "horror":
		particleColor = particles.ParticleColor{R: 0.8, G: 0.1, B: 0.2, A: 0.9} // Dark red blood
	case "cyberpunk":
		particleColor = particles.ParticleColor{R: 1.0, G: 0.4, B: 0.8, A: 0.9} // Magenta stim
	case "postapoc":
		particleColor = particles.ParticleColor{R: 0.7, G: 0.9, B: 0.3, A: 0.9} // Sickly green
	default:
		particleColor = particles.ParticleColor{R: 0.4, G: 1.0, B: 0.4, A: 0.9} // Default green
	}

	// Spawn rising heal particles
	for i := 0; i < count; i++ {
		offsetX := (localRng.Float64() - 0.5) * s.spreadFactor
		offsetY := (localRng.Float64() - 0.5) * s.spreadFactor

		particle := &particles.Particle{
			X:        pos.X + offsetX,
			Y:        pos.Y + offsetY,
			VX:       (localRng.Float64() - 0.5) * 15,
			VY:       -20 - localRng.Float64()*30, // Rise upward
			Color:    particleColor,
			Size:     2.0 + localRng.Float64()*2,
			Lifetime: 0.6 + localRng.Float64()*0.4,
			Age:      0,
			Behavior: particles.BehaviorRising,
		}

		s.particleSystem.AddParticle(particle)
	}
}

// GetLifestealForCompanion returns the current lifesteal percentage for a companion.
// This can be used by UI systems to display the bonus.
func (s *CompanionDamageLifestealSystem) GetLifestealForCompanion(companion *CompanionComponent) float64 {
	if companion == nil {
		return 0
	}

	lifesteal := companion.Loyalty * s.baseLifestealPerLoyalty

	if companion.HasPerk(PerkExtraHealth) {
		lifesteal += s.perkExtraHealthBonus
	}
	if companion.HasPerk(PerkSharedExperience) {
		lifesteal += s.perkSharedExpBonus
	}

	if mult, exists := s.genreMultipliers[s.genreID]; exists {
		lifesteal *= mult
	}

	if lifesteal > 0.15 {
		lifesteal = 0.15
	}

	return lifesteal
}
