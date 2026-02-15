// Package engine provides the LootRarityBeamSystem which spawns genre-aware
// vertical beam particles on ground items based on their rarity tier.
// Items with Uncommon or higher rarity emit upward-flowing colored particles,
// making valuable drops visible from a distance. This bridges the ItemComponent
// data on dropped entities with ParticleSystem for loot visibility feedback.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// lootBeamConfig holds per-rarity beam visual settings.
type lootBeamConfig struct {
	ColorHint     string
	ParticleCount int
	Intensity     float64
	SpreadX       float64
	BeamHeight    float64
	MinSize       float64
	MaxSize       float64
	PulseSpeed    float64
}

// LootRarityBeamSystem spawns vertical beam particles above ground items
// (entities with "item" + "position" but not "input" or "equipment") based
// on their rarity. Common items get no beam; Uncommon through Legendary
// get progressively more visible beams with rarity-appropriate colors.
//
// Genre presets modulate beam style:
//   - Fantasy: warm, magical particle rise
//   - Sci-fi: electric, sharp vertical lines
//   - Horror: dim, flickering, desaturated
//   - Cyberpunk: neon, bright, wide spread
//   - Post-apocalyptic: dusty, muted glow
type LootRarityBeamSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Cooldown tracking per entity to avoid particle spam
	cooldowns map[uint64]float64

	// Base emit interval in seconds
	baseInterval float64

	// Genre multiplier for emit interval
	genreMultiplier float64

	// Cached beam configs per rarity
	beamConfigs map[item.Rarity]lootBeamConfig
}

// NewLootRarityBeamSystem creates a new loot rarity beam system.
func NewLootRarityBeamSystem(world *World, seed int64) *LootRarityBeamSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "loot_rarity_beam")
		logEntry.Debug("loot rarity beam system created")
	}

	s := &LootRarityBeamSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		cooldowns:       make(map[uint64]float64, 64),
		baseInterval:    2.0,
		genreMultiplier: 1.0,
		beamConfigs:     make(map[item.Rarity]lootBeamConfig, 4),
	}
	s.initBeamConfigs()
	return s
}

// SetParticleSystem sets the particle system used for spawning beam effects.
func (s *LootRarityBeamSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware beam styling.
func (s *LootRarityBeamSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.genreMultiplier = genreEmitMultiplier(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes ground item entities and spawns rarity beam particles.
func (s *LootRarityBeamSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if !entity.HasComponent("item") || !entity.HasComponent("position") {
			continue
		}
		// Skip equipped/held items (entities with input are players)
		if entity.HasComponent("input") || entity.HasComponent("equipment") {
			continue
		}

		rarity, ok := s.getItemRarity(entity)
		if !ok || rarity < item.RarityUncommon {
			continue
		}

		cfg, hasCfg := s.beamConfigs[rarity]
		if !hasCfg {
			continue
		}

		// Cooldown check
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

		s.spawnBeamParticles(pos.X, pos.Y, cfg, rarity)

		// Reset cooldown: scale by genre and pulse speed
		interval := s.baseInterval * s.genreMultiplier / math.Max(cfg.PulseSpeed, 0.1)
		s.cooldowns[entity.ID] = interval
	}
}

// getItemRarity extracts the rarity from an entity's ItemComponent.
func (s *LootRarityBeamSystem) getItemRarity(entity *Entity) (item.Rarity, bool) {
	comp, ok := entity.GetComponent("item")
	if !ok || comp == nil {
		return item.RarityCommon, false
	}
	itemComp, ok := comp.(*ItemComponent)
	if !ok || itemComp.Item == nil {
		return item.RarityCommon, false
	}
	return itemComp.Item.Rarity, true
}

// spawnBeamParticles creates vertical beam particles at the given position.
func (s *LootRarityBeamSystem) spawnBeamParticles(x, y float64, cfg lootBeamConfig, rarity item.Rarity) {
	effectSeed := s.seed + int64(x*73) + int64(y*97) + int64(rarity)*31

	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    cfg.ParticleCount,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.2,
		SpreadX:  cfg.SpreadX,
		SpreadY:  cfg.BeamHeight,
		Gravity:  -25.0, // Strong upward flow for beam effect
		MinSize:  cfg.MinSize,
		MaxSize:  cfg.MaxSize,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["loot_beam"] = true
	config.Custom["color_hint"] = cfg.ColorHint
	config.Custom["intensity"] = cfg.Intensity
	config.Custom["rarity"] = rarity.String()

	s.particleSystem.SpawnParticles(s.world, config, x, y-cfg.BeamHeight*0.5)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"x":         x,
			"y":         y,
			"rarity":    rarity.String(),
			"color":     cfg.ColorHint,
			"particles": cfg.ParticleCount,
		}).Debug("loot beam particles spawned")
	}
}

// initBeamConfigs initializes the per-rarity beam visual configurations.
// Colors follow enchantment standard: Uncommon=Green, Rare=Blue, Epic=Purple, Legendary=Gold.
func (s *LootRarityBeamSystem) initBeamConfigs() {
	s.beamConfigs[item.RarityUncommon] = lootBeamConfig{
		ColorHint:     "green",
		ParticleCount: 3,
		Intensity:     0.3,
		SpreadX:       6.0,
		BeamHeight:    20.0,
		MinSize:       1.0,
		MaxSize:       2.0,
		PulseSpeed:    0.5,
	}
	s.beamConfigs[item.RarityRare] = lootBeamConfig{
		ColorHint:     "blue",
		ParticleCount: 5,
		Intensity:     0.5,
		SpreadX:       8.0,
		BeamHeight:    28.0,
		MinSize:       1.0,
		MaxSize:       2.5,
		PulseSpeed:    0.7,
	}
	s.beamConfigs[item.RarityEpic] = lootBeamConfig{
		ColorHint:     "purple",
		ParticleCount: 8,
		Intensity:     0.65,
		SpreadX:       10.0,
		BeamHeight:    36.0,
		MinSize:       1.5,
		MaxSize:       3.0,
		PulseSpeed:    0.9,
	}
	s.beamConfigs[item.RarityLegendary] = lootBeamConfig{
		ColorHint:     "gold",
		ParticleCount: 12,
		Intensity:     0.8,
		SpreadX:       12.0,
		BeamHeight:    44.0,
		MinSize:       1.5,
		MaxSize:       3.5,
		PulseSpeed:    1.2,
	}
}

// genreEmitMultiplier returns the genre-specific multiplier for beam emit interval.
func genreEmitMultiplier(genreID string) float64 {
	switch genreID {
	case "horror":
		return 1.5 // Dim, infrequent flicker
	case "cyberpunk":
		return 0.7 // Flashy, neon-frequent
	case "sci-fi":
		return 0.85 // Clean, moderate
	case "postapoc":
		return 1.3 // Sparse, muted
	default:
		return 1.0 // Fantasy standard
	}
}
