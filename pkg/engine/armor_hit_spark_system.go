// Package engine provides the ArmorHitSparkSystem which spawns genre-aware
// deflection particles when armored entities take damage. It monitors
// HealthComponent decreases on entities with EquipmentComponent armor slots,
// resolves the armor's material type, and spawns material-distinct spark
// particles via ParticleSystem. This bridges combat damage → equipment →
// visual particle feedback from the defender's armor perspective.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// armorSparkProfile defines visual properties for armor deflection particles per material.
type armorSparkProfile struct {
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

// armorSparkProfiles maps each armor material to its deflection visual profile.
var armorSparkProfiles = map[sprites.MaterialType]armorSparkProfile{
	sprites.MaterialMetal: {
		ParticleType: particles.ParticleSpark,
		ColorHint:    "white",
		MinSize:      2.0, MaxSize: 5.0,
		Gravity: 70.0, Duration: 0.25,
		SpreadX: 30.0, SpreadY: 30.0,
		BaseCount: 6,
	},
	sprites.MaterialLeather: {
		ParticleType: particles.ParticleDust,
		ColorHint:    "brown",
		MinSize:      1.5, MaxSize: 3.5,
		Gravity: 25.0, Duration: 0.30,
		SpreadX: 22.0, SpreadY: 22.0,
		BaseCount: 4,
	},
	sprites.MaterialCloth: {
		ParticleType: particles.ParticleSmoke,
		ColorHint:    "gray",
		MinSize:      1.0, MaxSize: 3.0,
		Gravity: -10.0, Duration: 0.35,
		SpreadX: 18.0, SpreadY: 18.0,
		BaseCount: 3,
	},
	sprites.MaterialWood: {
		ParticleType: particles.ParticleDust,
		ColorHint:    "tan",
		MinSize:      2.0, MaxSize: 4.0,
		Gravity: 55.0, Duration: 0.28,
		SpreadX: 25.0, SpreadY: 25.0,
		BaseCount: 5,
	},
	sprites.MaterialCrystal: {
		ParticleType: particles.ParticleSparkle,
		ColorHint:    "prismatic",
		MinSize:      2.0, MaxSize: 6.0,
		Gravity: 35.0, Duration: 0.40,
		SpreadX: 28.0, SpreadY: 28.0,
		BaseCount: 5,
	},
	sprites.MaterialEnergy: {
		ParticleType: particles.ParticleSparkle,
		ColorHint:    "electric",
		MinSize:      2.5, MaxSize: 6.0,
		Gravity: -20.0, Duration: 0.35,
		SpreadX: 40.0, SpreadY: 40.0,
		BaseCount: 6,
	},
}

// armorSparkGenreScale holds genre-specific intensity scaling for armor sparks.
type armorSparkGenreScale struct {
	CountMul    float64
	SizeMul     float64
	DurationMul float64
}

// ArmorHitSparkSystem spawns material-aware deflection particles when armored
// entities receive damage. The particle type, color, and intensity depend on
// the equipped armor's material and rarity.
type ArmorHitSparkSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Previous health values keyed by entity ID for damage detection
	prevHealth map[uint64]float64

	// Genre-specific scaling
	genreScale armorSparkGenreScale

	// Throttle cleanup of stale entries
	cleanupTimer    float64
	cleanupInterval float64
}

// NewArmorHitSparkSystem creates a new armor hit spark system.
func NewArmorHitSparkSystem(world *World, seed int64) *ArmorHitSparkSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "armor_hit_spark",
	})

	sys := &ArmorHitSparkSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logger,
		prevHealth:      make(map[uint64]float64, 128),
		genreID:         "fantasy",
		cleanupInterval: 5.0,
	}
	sys.genreScale = sys.getGenreScale("fantasy")
	logger.Debug("armor hit spark system created")
	return sys
}

// SetParticleSystem sets the particle system used for spawning deflection effects.
func (s *ArmorHitSparkSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware spark scaling.
func (s *ArmorHitSparkSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.genreScale = s.getGenreScale(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// getGenreScale returns genre-specific multipliers for spark intensity.
func (s *ArmorHitSparkSystem) getGenreScale(genreID string) armorSparkGenreScale {
	switch genreID {
	case "horror":
		return armorSparkGenreScale{CountMul: 0.7, SizeMul: 0.8, DurationMul: 1.3}
	case "cyberpunk":
		return armorSparkGenreScale{CountMul: 1.3, SizeMul: 1.1, DurationMul: 0.8}
	case "scifi":
		return armorSparkGenreScale{CountMul: 1.1, SizeMul: 1.0, DurationMul: 0.9}
	case "postapoc":
		return armorSparkGenreScale{CountMul: 1.0, SizeMul: 1.2, DurationMul: 1.0}
	default:
		return armorSparkGenreScale{CountMul: 1.0, SizeMul: 1.0, DurationMul: 1.0}
	}
}

// Update checks all entities for health decreases and spawns armor deflection particles.
func (s *ArmorHitSparkSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.cleanupTimer += deltaTime

	for _, entity := range entities {
		health := entity.GetHealth()
		if health == nil {
			continue
		}

		prev, tracked := s.prevHealth[entity.ID]
		s.prevHealth[entity.ID] = health.Current

		if !tracked {
			continue
		}

		damage := prev - health.Current
		if damage <= 0 {
			continue
		}

		armorItem, armorSlot := s.getBestArmorSlot(entity)
		if armorItem == nil {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		material := s.resolveArmorMaterial(armorItem, armorSlot)
		s.spawnArmorSparks(pos.X, pos.Y, material, armorItem.Rarity, damage)
	}

	// Periodic cleanup of stale entries
	if s.cleanupTimer >= s.cleanupInterval {
		s.cleanupTimer = 0
		s.cleanupStaleEntries(entities)
	}
}

// getBestArmorSlot returns the highest-rarity armor item from body armor slots.
func (s *ArmorHitSparkSystem) getBestArmorSlot(entity *Entity) (*item.Item, EquipmentSlot) {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil, SlotChest
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil, SlotChest
	}

	armorSlots := []EquipmentSlot{SlotChest, SlotHead, SlotLegs, SlotBoots, SlotGloves, SlotOffHand}

	var bestItem *item.Item
	bestSlot := SlotChest
	bestRarity := item.Rarity(-1)

	for _, slot := range armorSlots {
		itm := equipComp.GetEquipped(slot)
		if itm == nil {
			continue
		}
		if itm.Type != item.TypeArmor {
			continue
		}
		if itm.Rarity > bestRarity {
			bestItem = itm
			bestSlot = slot
			bestRarity = itm.Rarity
		}
	}

	return bestItem, bestSlot
}

// resolveArmorMaterial determines the material type from armor tags and slot.
func (s *ArmorHitSparkSystem) resolveArmorMaterial(itm *item.Item, slot EquipmentSlot) sprites.MaterialType {
	mat := sprites.GetMaterialTypeFromTags(itm.Tags, s.genreID)
	if mat != sprites.MaterialMetal {
		return mat
	}
	return sprites.GetMaterialTypeFromArmorType(slot.String(), s.genreID)
}

// spawnArmorSparks creates material-specific deflection particles at the hit location.
func (s *ArmorHitSparkSystem) spawnArmorSparks(x, y float64, mat sprites.MaterialType, rarity item.Rarity, damage float64) {
	profile := s.getProfile(mat)

	count := int(math.Ceil(float64(profile.BaseCount) * s.genreScale.CountMul))
	count += s.armorRarityBonus(rarity)
	if damage > 30 {
		count += int(math.Min(damage/20.0, 4))
	}
	if count > 20 {
		count = 20
	}

	effectSeed := s.seed + int64(x*991) + int64(y*997) + int64(mat)*41 + int64(damage*11)

	config := particles.Config{
		Type:     profile.ParticleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: profile.Duration * s.genreScale.DurationMul,
		SpreadX:  profile.SpreadX,
		SpreadY:  profile.SpreadY,
		Gravity:  profile.Gravity,
		MinSize:  profile.MinSize * s.genreScale.SizeMul,
		MaxSize:  profile.MaxSize * s.genreScale.SizeMul,
		Custom:   make(map[string]interface{}),
	}
	config.Custom["armor_deflection"] = true
	config.Custom["material"] = mat.String()
	config.Custom["color_hint"] = profile.ColorHint

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getProfile returns the spark profile for a material, defaulting to Metal.
func (s *ArmorHitSparkSystem) getProfile(mat sprites.MaterialType) armorSparkProfile {
	if p, ok := armorSparkProfiles[mat]; ok {
		return p
	}
	return armorSparkProfiles[sprites.MaterialMetal]
}

// armorRarityBonus returns extra particle count based on armor rarity.
func (s *ArmorHitSparkSystem) armorRarityBonus(rarity item.Rarity) int {
	switch rarity {
	case item.RarityUncommon:
		return 1
	case item.RarityRare:
		return 2
	case item.RarityEpic:
		return 3
	case item.RarityLegendary:
		return 5
	default:
		return 0
	}
}

// cleanupStaleEntries removes prevHealth entries for entities no longer in the update set.
func (s *ArmorHitSparkSystem) cleanupStaleEntries(entities []*Entity) {
	active := make(map[uint64]struct{}, len(entities))
	for _, e := range entities {
		active[e.ID] = struct{}{}
	}
	for id := range s.prevHealth {
		if _, ok := active[id]; !ok {
			delete(s.prevHealth, id)
		}
	}
}
