// Package engine provides the WeaponMaterialImpactParticleSystem for material-aware
// combat impact visuals. When melee attacks land, this system spawns impact particles
// whose type, color, and behavior depend on the weapon's material (Metal=sparks,
// Crystal=shards, Energy=flashes, Wood=splinters, Leather=dust, Cloth=wisps).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// materialImpactProfile defines visual properties for a material's impact particles.
type materialImpactProfile struct {
	ParticleType particles.ParticleType
	ColorHint    string
	MinSize      float64
	MaxSize      float64
	Gravity      float64
	Duration     float64
	SpreadX      float64
	SpreadY      float64
	BaseCount    int
}

// materialImpactProfiles maps each material to its impact visual profile.
var materialImpactProfiles = map[sprites.MaterialType]materialImpactProfile{
	sprites.MaterialMetal: {
		ParticleType: particles.ParticleSpark,
		ColorHint:    "orange",
		MinSize:      2.0, MaxSize: 5.0,
		Gravity: 60.0, Duration: 0.3,
		SpreadX: 35.0, SpreadY: 35.0,
		BaseCount: 8,
	},
	sprites.MaterialCrystal: {
		ParticleType: particles.ParticleSparkle,
		ColorHint:    "cyan",
		MinSize:      2.0, MaxSize: 6.0,
		Gravity: 40.0, Duration: 0.45,
		SpreadX: 30.0, SpreadY: 30.0,
		BaseCount: 6,
	},
	sprites.MaterialEnergy: {
		ParticleType: particles.ParticleSparkle,
		ColorHint:    "purple",
		MinSize:      3.0, MaxSize: 7.0,
		Gravity: -30.0, Duration: 0.4,
		SpreadX: 45.0, SpreadY: 45.0,
		BaseCount: 7,
	},
	sprites.MaterialWood: {
		ParticleType: particles.ParticleDust,
		ColorHint:    "brown",
		MinSize:      2.0, MaxSize: 4.0,
		Gravity: 50.0, Duration: 0.35,
		SpreadX: 25.0, SpreadY: 25.0,
		BaseCount: 5,
	},
	sprites.MaterialLeather: {
		ParticleType: particles.ParticleDust,
		ColorHint:    "tan",
		MinSize:      1.0, MaxSize: 3.0,
		Gravity: 20.0, Duration: 0.25,
		SpreadX: 20.0, SpreadY: 20.0,
		BaseCount: 4,
	},
	sprites.MaterialCloth: {
		ParticleType: particles.ParticleSmoke,
		ColorHint:    "white",
		MinSize:      1.0, MaxSize: 3.0,
		Gravity: -15.0, Duration: 0.4,
		SpreadX: 20.0, SpreadY: 20.0,
		BaseCount: 3,
	},
}

// WeaponMaterialImpactParticleSystem spawns material-aware impact particles
// when melee attacks land. It reads the attacker's weapon material type and
// selects a distinct particle profile (sparks, shards, dust, etc.).
type WeaponMaterialImpactParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
}

// NewWeaponMaterialImpactParticleSystem creates a new weapon material impact particle system.
func NewWeaponMaterialImpactParticleSystem(world *World, seed int64) *WeaponMaterialImpactParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weapon_material_impact_particle")
		logEntry.Debug("weapon material impact particle system created")
	}

	return &WeaponMaterialImpactParticleSystem{
		world:  world,
		seed:   seed,
		rng:    rand.New(rand.NewSource(seed)),
		logger: logEntry,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *WeaponMaterialImpactParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware material resolution.
func (s *WeaponMaterialImpactParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update is a no-op; this system is callback-driven via OnMeleeImpact.
func (s *WeaponMaterialImpactParticleSystem) Update(entities []*Entity, deltaTime float64) {}

// OnMeleeImpact is called when a melee attack lands. Register this as a
// CombatSystem damage callback to spawn material-specific impact particles.
func (s *WeaponMaterialImpactParticleSystem) OnMeleeImpact(attacker, target *Entity, damage float64) {
	if s.particleSystem == nil || s.world == nil || attacker == nil || target == nil {
		return
	}

	weapon := s.getEquippedWeapon(attacker)
	if weapon == nil {
		return
	}
	if weapon.Stats.IsProjectile {
		return
	}

	pos := target.GetPosition()
	if pos == nil {
		return
	}

	material := sprites.GetMaterialTypeFromTags(weapon.Tags, s.genreID)
	s.spawnImpactParticles(pos.X, pos.Y, material, weapon.Rarity, damage)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"attacker_id": attacker.ID,
			"target_id":   target.ID,
			"material":    material.String(),
			"rarity":      weapon.Rarity.String(),
			"damage":      damage,
		}).Debug("weapon material impact particles spawned")
	}
}

// getEquippedWeapon retrieves the main-hand weapon from an entity's equipment.
func (s *WeaponMaterialImpactParticleSystem) getEquippedWeapon(entity *Entity) *item.Item {
	comp, has := entity.GetComponent("equipment")
	if !has || comp == nil {
		return nil
	}
	equip, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equip.GetEquipped(SlotMainHand)
}

// spawnImpactParticles creates material-specific impact particles at the hit location.
func (s *WeaponMaterialImpactParticleSystem) spawnImpactParticles(x, y float64, mat sprites.MaterialType, rarity item.Rarity, damage float64) {
	profile := s.getProfile(mat)

	count := profile.BaseCount + s.rarityBonus(rarity)
	if damage > 50 {
		count += 2
	}
	if count > 24 {
		count = 24
	}

	effectSeed := s.seed + int64(x*997) + int64(y*991) + int64(mat)*37 + int64(damage*7)

	config := particles.Config{
		Type:     profile.ParticleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: profile.Duration,
		SpreadX:  profile.SpreadX,
		SpreadY:  profile.SpreadY,
		Gravity:  profile.Gravity,
		MinSize:  profile.MinSize,
		MaxSize:  profile.MaxSize,
		Custom:   make(map[string]interface{}),
	}
	config.Custom["material_impact"] = true
	config.Custom["material"] = mat.String()
	config.Custom["color_hint"] = profile.ColorHint

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getProfile returns the impact profile for a material, defaulting to Metal.
func (s *WeaponMaterialImpactParticleSystem) getProfile(mat sprites.MaterialType) materialImpactProfile {
	if p, ok := materialImpactProfiles[mat]; ok {
		return p
	}
	return materialImpactProfiles[sprites.MaterialMetal]
}

// rarityBonus returns extra particle count based on weapon rarity.
func (s *WeaponMaterialImpactParticleSystem) rarityBonus(rarity item.Rarity) int {
	switch rarity {
	case item.RarityUncommon:
		return 1
	case item.RarityRare:
		return 2
	case item.RarityEpic:
		return 4
	case item.RarityLegendary:
		return 6
	default:
		return 0
	}
}

// SpawnImpactEffect allows external systems to trigger material impact particles directly.
func (s *WeaponMaterialImpactParticleSystem) SpawnImpactEffect(x, y float64, mat sprites.MaterialType, rarity item.Rarity, damage float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnImpactParticles(x, y, mat, rarity, damage)
}
