// Package engine provides the WeaponMaterialParticleSystem which spawns
// genre-aware ambient particles based on the material type of the equipped
// weapon. Metal weapons emit metallic glint sparks, crystal weapons emit
// prismatic sparkles, energy weapons emit electrical arcs, and wood weapons
// emit subtle dust motes. This bridges EquipmentComponent weapon data with
// ParticleSystem for idle weapon visual feedback.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// materialParticleConfig holds per-material particle visual settings.
type materialParticleConfig struct {
	ParticleType particles.ParticleType
	Count        int
	SpreadX      float64
	SpreadY      float64
	Gravity      float64
	MinSize      float64
	MaxSize      float64
	Duration     float64
	ColorHint    string
}

// WeaponMaterialParticleSystem spawns ambient particles around entities based
// on the material type of their equipped main-hand weapon. Particle type, count,
// and color scale with material visual properties (sheen, reflectivity).
//
// Genre presets modulate emit interval:
//   - Fantasy: 1.0x (standard)
//   - Sci-fi: 0.8x (more frequent, tech emphasis)
//   - Horror: 1.4x (subdued, ominous)
//   - Cyberpunk: 0.7x (flashy, neon-influenced)
//   - Post-apocalyptic: 1.3x (dusty, sparse)
type WeaponMaterialParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Cooldown tracking per entity
	cooldowns map[uint64]float64

	// Base emit interval in seconds
	baseInterval float64

	// Genre multiplier for emit interval
	genreMultiplier float64

	// Material configs cache
	materialConfigs map[sprites.MaterialType]materialParticleConfig
}

// NewWeaponMaterialParticleSystem creates a new weapon material particle system.
func NewWeaponMaterialParticleSystem(world *World, seed int64) *WeaponMaterialParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weapon_material_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("weapon material particle system created")
		}
	}

	s := &WeaponMaterialParticleSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		cooldowns:       make(map[uint64]float64, 64),
		baseInterval:    2.0,
		genreMultiplier: 1.0,
		materialConfigs: buildMaterialConfigs(),
	}

	return s
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *WeaponMaterialParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre sets the genre ID and updates the genre-specific emit multiplier.
func (s *WeaponMaterialParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	switch genreID {
	case "sci-fi":
		s.genreMultiplier = 0.8
	case "cyberpunk":
		s.genreMultiplier = 0.7
	case "horror":
		s.genreMultiplier = 1.4
	case "post-apocalyptic":
		s.genreMultiplier = 1.3
	default:
		s.genreMultiplier = 1.0
	}
}

// Update processes entities with equipped weapons and spawns material particles.
func (s *WeaponMaterialParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if !entity.HasComponent("equipment") {
			continue
		}

		mat, hasMat := s.getWeaponMaterial(entity)
		if !hasMat {
			delete(s.cooldowns, entity.ID)
			continue
		}

		// Update cooldown
		cd := s.cooldowns[entity.ID]
		cd -= deltaTime
		if cd > 0 {
			s.cooldowns[entity.ID] = cd
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		cfg, ok := s.materialConfigs[mat]
		if !ok {
			continue
		}

		s.spawnMaterialParticles(pos.X, pos.Y, cfg, mat)

		// Reset cooldown scaled by material sheen and genre
		props := sprites.GetMaterialVisualProperties(mat)
		sheenFactor := math.Max(0.3, 1.0-props.Sheen*0.6)
		interval := s.baseInterval * sheenFactor * s.genreMultiplier
		s.cooldowns[entity.ID] = interval
	}
}

// getWeaponMaterial extracts the material type from the entity's main-hand weapon.
func (s *WeaponMaterialParticleSystem) getWeaponMaterial(entity *Entity) (sprites.MaterialType, bool) {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return sprites.MaterialMetal, false
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return sprites.MaterialMetal, false
	}

	weapon := equipComp.GetEquipped(SlotMainHand)
	if weapon == nil {
		return sprites.MaterialMetal, false
	}

	// Determine material from tags first, then weapon type
	if len(weapon.Tags) > 0 {
		mat := sprites.GetMaterialTypeFromTags(weapon.Tags, s.genreID)
		return mat, true
	}

	if weapon.Type == item.TypeWeapon {
		mat := sprites.GetMaterialTypeFromWeaponType(weapon.WeaponType.String(), s.genreID)
		return mat, true
	}

	return sprites.MaterialMetal, false
}

// spawnMaterialParticles creates particles based on material configuration.
func (s *WeaponMaterialParticleSystem) spawnMaterialParticles(x, y float64, cfg materialParticleConfig, mat sprites.MaterialType) {
	effectSeed := s.seed + int64(x*97) + int64(y*131) + int64(mat)

	config := particles.Config{
		Type:     cfg.ParticleType,
		Count:    cfg.Count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: cfg.Duration,
		SpreadX:  cfg.SpreadX,
		SpreadY:  cfg.SpreadY,
		Gravity:  cfg.Gravity,
		MinSize:  cfg.MinSize,
		MaxSize:  cfg.MaxSize,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["weapon_material"] = mat.String()
	config.Custom["color_hint"] = cfg.ColorHint

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"x":        x,
			"y":        y,
			"material": mat.String(),
			"count":    cfg.Count,
		}).Debug("weapon material particles spawned")
	}
}

// buildMaterialConfigs creates the per-material particle configurations.
func buildMaterialConfigs() map[sprites.MaterialType]materialParticleConfig {
	return map[sprites.MaterialType]materialParticleConfig{
		sprites.MaterialMetal: {
			ParticleType: particles.ParticleSpark,
			Count:        3,
			SpreadX:      16.0,
			SpreadY:      20.0,
			Gravity:      -10.0,
			MinSize:      1.0,
			MaxSize:      2.0,
			Duration:     0.6,
			ColorHint:    "silver",
		},
		sprites.MaterialCrystal: {
			ParticleType: particles.ParticleSparkle,
			Count:        4,
			SpreadX:      18.0,
			SpreadY:      18.0,
			Gravity:      -8.0,
			MinSize:      1.0,
			MaxSize:      3.0,
			Duration:     0.8,
			ColorHint:    "prismatic",
		},
		sprites.MaterialEnergy: {
			ParticleType: particles.ParticleSpark,
			Count:        5,
			SpreadX:      20.0,
			SpreadY:      22.0,
			Gravity:      -12.0,
			MinSize:      1.0,
			MaxSize:      2.5,
			Duration:     0.5,
			ColorHint:    "electric_blue",
		},
		sprites.MaterialWood: {
			ParticleType: particles.ParticleDust,
			Count:        2,
			SpreadX:      14.0,
			SpreadY:      16.0,
			Gravity:      5.0,
			MinSize:      0.5,
			MaxSize:      1.5,
			Duration:     1.0,
			ColorHint:    "brown",
		},
		sprites.MaterialLeather: {
			ParticleType: particles.ParticleDust,
			Count:        1,
			SpreadX:      12.0,
			SpreadY:      14.0,
			Gravity:      3.0,
			MinSize:      0.5,
			MaxSize:      1.0,
			Duration:     0.8,
			ColorHint:    "tan",
		},
		sprites.MaterialCloth: {
			ParticleType: particles.ParticleDust,
			Count:        1,
			SpreadX:      10.0,
			SpreadY:      12.0,
			Gravity:      2.0,
			MinSize:      0.5,
			MaxSize:      1.0,
			Duration:     0.7,
			ColorHint:    "white",
		},
	}
}
