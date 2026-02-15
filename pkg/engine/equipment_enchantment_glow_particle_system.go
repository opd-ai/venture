// Package engine provides the EquipmentEnchantmentGlowParticleSystem for ambient
// enchantment glow particles on entities with rare+ equipped items.
// This system reads equipment rarity from EquipmentComponent and spawns continuous
// rarity-colored glow particles via ParticleSystem, connecting equipment data
// with visual feedback that was previously unused.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// EquipmentEnchantmentGlowParticleSystem spawns ambient glow particles around
// entities with Uncommon or higher rarity equipped items. Particle color, count,
// and pulse speed scale with the highest rarity item equipped.
type EquipmentEnchantmentGlowParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Cooldown tracking per entity to avoid particle spam
	cooldowns map[uint64]float64

	// Emit interval in seconds (based on pulse speed)
	baseInterval float64
}

// NewEquipmentEnchantmentGlowParticleSystem creates a new enchantment glow particle system.
func NewEquipmentEnchantmentGlowParticleSystem(world *World, seed int64) *EquipmentEnchantmentGlowParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_enchantment_glow_particle")
		logEntry.Debug("equipment enchantment glow particle system created")
	}

	return &EquipmentEnchantmentGlowParticleSystem{
		world:        world,
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
		logger:       logEntry,
		cooldowns:    make(map[uint64]float64, 64),
		baseInterval: 1.5, // Emit every 1.5s base, scaled by pulse speed
	}
}

// SetParticleSystem sets the particle system used for spawning glow effects.
func (s *EquipmentEnchantmentGlowParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *EquipmentEnchantmentGlowParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities with equipment and spawns enchantment glow particles.
func (s *EquipmentEnchantmentGlowParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if !entity.HasComponent("equipment") {
			continue
		}

		// Get highest rarity from equipped items
		highestRarity, hasRareItem := s.getHighestEquippedRarity(entity)
		if !hasRareItem {
			// Clean up cooldown for entities that no longer have rare items
			delete(s.cooldowns, entity.ID)
			continue
		}

		// Get enchantment glow config for this rarity
		glow := sprites.GetEnchantmentFromRarity(highestRarity.String())
		if !glow.Active {
			continue
		}

		// Update cooldown
		cd := s.cooldowns[entity.ID]
		cd -= deltaTime
		if cd > 0 {
			s.cooldowns[entity.ID] = cd
			continue
		}

		// Get entity position
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Spawn glow particles
		s.spawnGlowParticles(pos.X, pos.Y, glow, highestRarity)

		// Reset cooldown: higher pulse speed = more frequent emissions
		interval := s.baseInterval / math.Max(glow.PulseSpeed, 0.1)
		s.cooldowns[entity.ID] = interval
	}
}

// getHighestEquippedRarity scans all equipment slots and returns the highest rarity found.
func (s *EquipmentEnchantmentGlowParticleSystem) getHighestEquippedRarity(entity *Entity) (item.Rarity, bool) {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return item.RarityCommon, false
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return item.RarityCommon, false
	}

	slots := []EquipmentSlot{
		SlotMainHand, SlotOffHand, SlotHead, SlotChest,
		SlotLegs, SlotBoots, SlotGloves,
		SlotAccessory1, SlotAccessory2, SlotAccessory3,
	}

	highest := item.RarityCommon
	found := false

	for _, slot := range slots {
		itm := equipComp.GetEquipped(slot)
		if itm == nil {
			continue
		}
		if itm.Rarity > highest {
			highest = itm.Rarity
		}
		found = true
	}

	// Only return true if we found at least Uncommon
	if !found || highest < item.RarityUncommon {
		return highest, false
	}

	return highest, true
}

// spawnGlowParticles creates ambient glow particles based on enchantment configuration.
func (s *EquipmentEnchantmentGlowParticleSystem) spawnGlowParticles(x, y float64, glow sprites.EnchantmentGlow, rarity item.Rarity) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(rarity)

	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    glow.ParticleCount,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.0,
		SpreadX:  24.0, // Tight spread around entity
		SpreadY:  24.0,
		Gravity:  -15.0, // Gentle upward float
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["enchantment_glow"] = true
	config.Custom["color_hint"] = glow.Color
	config.Custom["intensity"] = glow.Intensity
	config.Custom["rarity"] = rarity.String()

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"x":      x,
			"y":      y,
			"rarity": rarity.String(),
			"color":  glow.Color,
			"count":  glow.ParticleCount,
		}).Debug("enchantment glow particles spawned")
	}
}

// getGenreColorModifier returns a genre-specific color hint overlay.
func (s *EquipmentEnchantmentGlowParticleSystem) getGenreColorModifier() string {
	switch s.genreID {
	case "horror":
		return "dark"
	case "cyberpunk":
		return "neon"
	case "sci-fi":
		return "electric"
	default:
		return ""
	}
}
