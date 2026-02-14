// Package engine provides the WeaponSwingParticleSystem for visual weapon swing feedback.
// This system connects CombatSystem damage events with ParticleSystem to spawn
// rarity-aware weapon trail particles when melee attacks land.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeaponSwingParticleSystem spawns weapon trail particles when melee attacks land.
// It connects EquipmentComponent (weapon rarity) with ParticleSystem to provide
// genre-aware visual feedback that scales with weapon quality.
type WeaponSwingParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	baseParticleCount int
	spreadFactor      float64
}

// NewWeaponSwingParticleSystem creates a new weapon swing particle system.
func NewWeaponSwingParticleSystem(world *World, seed int64) *WeaponSwingParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weapon_swing_particle")
		logEntry.Debug("weapon swing particle system created")
	}

	return &WeaponSwingParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		baseParticleCount: 6,    // Modest base count for weapon trails
		spreadFactor:      40.0, // Tighter spread than combat effects
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *WeaponSwingParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *WeaponSwingParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *WeaponSwingParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnDamageDealt, no per-frame processing needed
}

// OnDamageDealt is called when damage is dealt to spawn weapon swing particles.
// This method should be registered as an additional callback with the CombatSystem.
func (s *WeaponSwingParticleSystem) OnDamageDealt(attacker, target *Entity, damage float64) {
	if s.particleSystem == nil || s.world == nil || attacker == nil || target == nil {
		return
	}

	// Get attacker's equipped weapon
	weapon := s.getEquippedWeapon(attacker)
	if weapon == nil {
		return // No weapon equipped, skip particles
	}

	// Only spawn for melee weapons (projectile weapons handled separately)
	if weapon.Stats.IsProjectile {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn weapon swing particles at target location
	s.spawnSwingParticles(pos.X, pos.Y, weapon.Rarity, damage)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"attacker_id": attacker.ID,
			"target_id":   target.ID,
			"weapon":      weapon.Name,
			"rarity":      weapon.Rarity.String(),
			"damage":      damage,
		}).Debug("weapon swing particles spawned")
	}
}

// getEquippedWeapon retrieves the weapon from attacker's equipment component.
func (s *WeaponSwingParticleSystem) getEquippedWeapon(entity *Entity) *item.Item {
	equipComp, hasEquip := entity.GetComponent("equipment")
	if !hasEquip || equipComp == nil {
		return nil
	}

	equip, ok := equipComp.(*EquipmentComponent)
	if !ok {
		return nil
	}

	return equip.GetEquipped(SlotMainHand)
}

// spawnSwingParticles creates the weapon swing particle effect based on rarity.
func (s *WeaponSwingParticleSystem) spawnSwingParticles(x, y float64, rarity item.Rarity, damage float64) {
	// Calculate particle count based on rarity (more particles for rarer weapons)
	count := s.getParticleCount(rarity)

	// Use deterministic seed offset for this specific swing
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(damage*10)

	// Get particle color based on rarity (matches enchantment glow colors)
	colorHint := s.getRarityColorHint(rarity)

	// Create trail config for weapon swing effect
	config := particles.Config{
		Type:     particles.ParticleSpark, // Trail-like effect via spark particles
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.35, // Short duration for quick swing feedback
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -20.0, // Slight upward drift
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as weapon swing with rarity info
	config.Custom["weapon_swing"] = true
	config.Custom["rarity"] = rarity.String()
	config.Custom["color_hint"] = colorHint

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleCount returns particle count scaled by weapon rarity.
func (s *WeaponSwingParticleSystem) getParticleCount(rarity item.Rarity) int {
	// Base count + rarity bonus
	// Common: 6, Uncommon: 8, Rare: 10, Epic: 13, Legendary: 16
	switch rarity {
	case item.RarityCommon:
		return s.baseParticleCount
	case item.RarityUncommon:
		return s.baseParticleCount + 2
	case item.RarityRare:
		return s.baseParticleCount + 4
	case item.RarityEpic:
		return s.baseParticleCount + 7
	case item.RarityLegendary:
		return s.baseParticleCount + 10
	default:
		return s.baseParticleCount
	}
}

// getRarityColorHint returns color hint for rarity-based particle coloring.
// Colors match the enchantment glow system for visual consistency.
func (s *WeaponSwingParticleSystem) getRarityColorHint(rarity item.Rarity) string {
	switch rarity {
	case item.RarityCommon:
		return "white"
	case item.RarityUncommon:
		return "green"
	case item.RarityRare:
		return "blue"
	case item.RarityEpic:
		return "purple"
	case item.RarityLegendary:
		return "gold"
	default:
		return "white"
	}
}

// SpawnSwingEffect allows external systems to trigger swing particles directly.
// This is useful for special attacks or abilities that want weapon trail effects.
func (s *WeaponSwingParticleSystem) SpawnSwingEffect(x, y float64, rarity item.Rarity, damage float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnSwingParticles(x, y, rarity, damage)
}
