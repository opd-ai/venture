// Package engine provides the EquipmentChangeFlashSystem for visual feedback
// when equipment is changed. This system connects EquipmentComponent slot
// tracking with ParticleSystem to spawn genre-aware flash particles whenever
// an entity equips or unequips an item, giving immediate visual confirmation.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// EquipmentChangeFlashSystem detects equipment slot changes and spawns a brief
// genre-aware particle burst at the entity's position. It tracks the last known
// item ID per slot per entity to detect transitions.
type EquipmentChangeFlashSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// lastEquipState caches item IDs per slot per entity to detect changes.
	lastEquipState map[uint64]map[EquipmentSlot]string
}

// NewEquipmentChangeFlashSystem creates a new equipment change flash system.
func NewEquipmentChangeFlashSystem(world *World, seed int64) *EquipmentChangeFlashSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_change_flash")
		logEntry.Debug("equipment change flash system created")
	}

	return &EquipmentChangeFlashSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		lastEquipState: make(map[uint64]map[EquipmentSlot]string, 64),
	}
}

// SetParticleSystem sets the particle system used for spawning flash effects.
func (s *EquipmentChangeFlashSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware flash colors.
func (s *EquipmentChangeFlashSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// allSlots is the list of equipment slots checked for changes.
var equipChangeFlashSlots = []EquipmentSlot{
	SlotMainHand, SlotOffHand, SlotHead, SlotChest,
	SlotLegs, SlotBoots, SlotGloves,
	SlotAccessory1, SlotAccessory2, SlotAccessory3,
}

// Update checks each entity's equipment for changes and spawns flash particles.
func (s *EquipmentChangeFlashSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if !entity.HasComponent("equipment") {
			continue
		}

		comp, ok := entity.GetComponent("equipment")
		if !ok || comp == nil {
			continue
		}
		equipComp, ok := comp.(*EquipmentComponent)
		if !ok {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		prev, tracked := s.lastEquipState[entity.ID]
		if !tracked {
			// First time seeing this entity; snapshot state without flashing.
			s.lastEquipState[entity.ID] = s.snapshotSlots(equipComp)
			continue
		}

		changed := false
		for _, slot := range equipChangeFlashSlots {
			current := s.itemIDForSlot(equipComp, slot)
			if current != prev[slot] {
				changed = true
				break
			}
		}

		if changed {
			s.spawnFlashParticles(pos.X, pos.Y, entity.ID)
			s.lastEquipState[entity.ID] = s.snapshotSlots(equipComp)
		}
	}
}

// snapshotSlots returns a map of slot → item ID for the current equipment state.
func (s *EquipmentChangeFlashSystem) snapshotSlots(ec *EquipmentComponent) map[EquipmentSlot]string {
	snap := make(map[EquipmentSlot]string, len(equipChangeFlashSlots))
	for _, slot := range equipChangeFlashSlots {
		snap[slot] = s.itemIDForSlot(ec, slot)
	}
	return snap
}

// itemIDForSlot returns the item ID for a given slot, or empty string if empty.
func (s *EquipmentChangeFlashSystem) itemIDForSlot(ec *EquipmentComponent, slot EquipmentSlot) string {
	itm := ec.GetEquipped(slot)
	if itm == nil {
		return ""
	}
	return itm.ID
}

// spawnFlashParticles emits a short burst of genre-aware sparkle particles.
func (s *EquipmentChangeFlashSystem) spawnFlashParticles(x, y float64, entityID uint64) {
	effectSeed := s.seed + int64(entityID)*31

	config := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    6,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.4,
		SpreadX:  20.0,
		SpreadY:  20.0,
		Gravity:  -10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["equipment_change_flash"] = true
	config.Custom["color_hint"] = s.genreFlashColor()

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// genreFlashColor returns a genre-appropriate flash color name.
func (s *EquipmentChangeFlashSystem) genreFlashColor() string {
	switch s.genreID {
	case "fantasy":
		return "gold"
	case "scifi":
		return "cyan"
	case "horror":
		return "dark_red"
	case "cyberpunk":
		return "neon_green"
	case "postapoc":
		return "amber"
	default:
		return "white"
	}
}

// GetLastEquipState returns the cached equipment state for testing.
func (s *EquipmentChangeFlashSystem) GetLastEquipState() map[uint64]map[EquipmentSlot]string {
	return s.lastEquipState
}
