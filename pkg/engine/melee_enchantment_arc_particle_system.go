// Package engine provides the MeleeEnchantmentArcParticleSystem which spawns
// rarity-colored particles along the melee swing arc path for enchanted weapons.
// This bridges the MeleeSwingArcComponent (active swing visual state) with
// equipment enchantment data (rarity/color) via ParticleSystem, producing
// genre-aware enchantment trails during melee attacks.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// enchantArcGenrePreset holds genre-specific modifiers for enchantment arc particles.
type enchantArcGenrePreset struct {
	ParticleType particles.ParticleType
	GravityMul   float64 // multiplier on base gravity
	SizeMul      float64 // multiplier on particle size
	DurationMul  float64 // multiplier on particle lifetime
}

// enchantArcGenrePresets defines how different genres modify arc particles.
var enchantArcGenrePresets = map[string]enchantArcGenrePreset{
	"fantasy":   {ParticleType: particles.ParticleSparkle, GravityMul: 1.0, SizeMul: 1.0, DurationMul: 1.0},
	"scifi":     {ParticleType: particles.ParticleSpark, GravityMul: 0.5, SizeMul: 0.8, DurationMul: 1.2},
	"horror":    {ParticleType: particles.ParticleSmoke, GravityMul: 1.5, SizeMul: 1.2, DurationMul: 0.8},
	"cyberpunk": {ParticleType: particles.ParticleSpark, GravityMul: 0.3, SizeMul: 0.9, DurationMul: 1.1},
	"postapoc":  {ParticleType: particles.ParticleEmber, GravityMul: 1.3, SizeMul: 1.1, DurationMul: 0.9},
}

// enchantArcRarityConfig maps rarity to particle intensity for arc trails.
type enchantArcRarityConfig struct {
	ColorHint     string
	ParticleCount int
	Intensity     float64 // 0.0–1.0
	SpreadFactor  float64 // lateral spread from arc path
}

var enchantArcRarityConfigs = map[item.Rarity]enchantArcRarityConfig{
	item.RarityUncommon:  {ColorHint: "green", ParticleCount: 2, Intensity: 0.2, SpreadFactor: 6.0},
	item.RarityRare:      {ColorHint: "blue", ParticleCount: 4, Intensity: 0.4, SpreadFactor: 8.0},
	item.RarityEpic:      {ColorHint: "purple", ParticleCount: 6, Intensity: 0.6, SpreadFactor: 10.0},
	item.RarityLegendary: {ColorHint: "gold", ParticleCount: 10, Intensity: 0.8, SpreadFactor: 12.0},
}

// MeleeEnchantmentArcParticleSystem spawns enchantment-colored particles along
// active melee swing arcs for entities with Uncommon+ rarity weapons.
type MeleeEnchantmentArcParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
	genrePreset    enchantArcGenrePreset

	// Track which arcs we already emitted for to avoid per-frame spam
	emittedArcs map[uint64]bool

	// Throttle: only check every N seconds
	checkTimer    float64
	checkInterval float64
}

// NewMeleeEnchantmentArcParticleSystem creates a new enchantment arc particle system.
func NewMeleeEnchantmentArcParticleSystem(world *World, seed int64) *MeleeEnchantmentArcParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "melee_enchantment_arc_particle")
		logEntry.Debug("melee enchantment arc particle system created")
	}

	return &MeleeEnchantmentArcParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		genrePreset:   enchantArcGenrePresets["fantasy"],
		emittedArcs:   make(map[uint64]bool, 32),
		checkTimer:    0,
		checkInterval: 0.05, // 20 Hz check rate
	}
}

// SetParticleSystem sets the particle system for spawning effects.
func (s *MeleeEnchantmentArcParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre for genre-aware particle visuals.
func (s *MeleeEnchantmentArcParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if preset, ok := enchantArcGenrePresets[genreID]; ok {
		s.genrePreset = preset
	}
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update checks entities with active swing arcs and enchanted weapons,
// spawning rarity-colored particles along the arc path.
func (s *MeleeEnchantmentArcParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.checkTimer += deltaTime
	if s.checkTimer < s.checkInterval {
		return
	}
	s.checkTimer = 0

	for _, entity := range entities {
		arcComp, hasArc := entity.GetComponent("melee_swing_arc")
		if !hasArc {
			continue
		}
		arc, ok := arcComp.(*MeleeSwingArcComponent)
		if !ok || !arc.Active {
			// Arc ended, clear emitted state
			delete(s.emittedArcs, entity.ID)
			continue
		}

		// Only emit once per arc activation
		if s.emittedArcs[entity.ID] {
			continue
		}

		// Get weapon rarity
		rarity, hasEnchant := s.getWeaponRarity(entity)
		if !hasEnchant {
			continue
		}

		rarityConf, hasConf := enchantArcRarityConfigs[rarity]
		if !hasConf {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Spawn particles along the arc path
		s.spawnArcParticles(pos.X, pos.Y, arc, rarityConf, entity.ID)
		s.emittedArcs[entity.ID] = true
	}
}

// getWeaponRarity returns the main-hand weapon's rarity if Uncommon+.
func (s *MeleeEnchantmentArcParticleSystem) getWeaponRarity(entity *Entity) (item.Rarity, bool) {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return item.RarityCommon, false
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return item.RarityCommon, false
	}

	weapon := equipComp.GetEquipped(SlotMainHand)
	if weapon == nil {
		return item.RarityCommon, false
	}

	// Only projectile-free weapons (melee) get arc particles
	if weapon.Stats.IsProjectile {
		return item.RarityCommon, false
	}

	if weapon.Rarity < item.RarityUncommon {
		return weapon.Rarity, false
	}

	return weapon.Rarity, true
}

// spawnArcParticles spawns enchantment particles along the swing arc path.
func (s *MeleeEnchantmentArcParticleSystem) spawnArcParticles(
	cx, cy float64,
	arc *MeleeSwingArcComponent,
	conf enchantArcRarityConfig,
	entityID uint64,
) {
	preset := s.genrePreset

	// Distribute particles evenly along the arc angle range
	angleRange := arc.ArcAngleEnd - arc.ArcAngleStart
	if angleRange == 0 {
		return
	}

	for i := 0; i < conf.ParticleCount; i++ {
		// Position along arc
		t := float64(i) / math.Max(float64(conf.ParticleCount-1), 1.0)
		angle := arc.ArcAngleStart + t*angleRange
		px := cx + math.Cos(angle)*arc.ArcRadius
		py := cy + math.Sin(angle)*arc.ArcRadius

		effectSeed := s.seed + int64(entityID) + int64(i)*7

		config := particles.Config{
			Type:     preset.ParticleType,
			Count:    1,
			GenreID:  s.genreID,
			Seed:     effectSeed,
			Duration: 0.4 * preset.DurationMul,
			SpreadX:  conf.SpreadFactor,
			SpreadY:  conf.SpreadFactor,
			Gravity:  -10.0 * preset.GravityMul,
			MinSize:  1.5 * preset.SizeMul,
			MaxSize:  3.5 * preset.SizeMul,
			Custom:   make(map[string]interface{}),
		}

		config.Custom["enchantment_arc"] = true
		config.Custom["color_hint"] = conf.ColorHint
		config.Custom["intensity"] = conf.Intensity
		config.Custom["rarity"] = sprites.GetEnchantmentFromRarity(conf.ColorHint).Color

		s.particleSystem.SpawnParticles(s.world, config, px, py)
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"particle_count": conf.ParticleCount,
			"color":          conf.ColorHint,
			"arc_radius":     arc.ArcRadius,
		}).Debug("enchantment arc particles spawned")
	}
}
