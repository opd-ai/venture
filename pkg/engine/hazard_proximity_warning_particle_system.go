// Package engine provides the HazardProximityWarningParticleSystem for visual
// hazard proximity feedback. This system connects HazardSystem zone tracking
// with ParticleSystem to spawn genre-aware warning particles around entities
// that are near (but not yet inside) environmental hazards, giving players
// advance visual notice of danger.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// hazardWarningPreset holds genre-specific visual parameters for warning particles.
type hazardWarningPreset struct {
	ParticleType particles.ParticleType
	MinSize      float64
	MaxSize      float64
	Duration     float64
	Gravity      float64
	SpreadFactor float64
}

// HazardProximityWarningParticleSystem spawns warning particles around entities
// approaching environmental hazards. It queries the HazardZoneTracker for nearby
// zones and emits genre-aware particles at the entity's edge facing the hazard.
type HazardProximityWarningParticleSystem struct {
	world          *World
	hazardSystem   *HazardSystem
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Warning detection radius beyond hazard edge
	warningMargin float64
	// Emit timing
	emitInterval  float64
	timeSinceEmit float64
	baseCount     int
	// Per-entity cooldown to avoid particle spam
	lastWarned map[uint64]float64
}

// NewHazardProximityWarningParticleSystem creates a new hazard proximity warning system.
func NewHazardProximityWarningParticleSystem(world *World, seed int64) *HazardProximityWarningParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "hazard_proximity_warning")
		logEntry.Debug("hazard proximity warning particle system created")
	}

	return &HazardProximityWarningParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		genreID:       "fantasy",
		warningMargin: 48.0, // Warn within 48px of hazard edge
		emitInterval:  0.35,
		timeSinceEmit: 0.0,
		baseCount:     4,
		lastWarned:    make(map[uint64]float64, 32),
	}
}

// SetHazardSystem links the hazard system for zone queries.
func (s *HazardProximityWarningParticleSystem) SetHazardSystem(hs *HazardSystem) {
	s.hazardSystem = hs
	if s.logger != nil {
		s.logger.Debug("hazard system linked")
	}
}

// SetParticleSystem sets the particle system for spawning effects.
func (s *HazardProximityWarningParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre configures genre-aware particle visuals.
func (s *HazardProximityWarningParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update scans entities with position and health for nearby hazard zones,
// spawning warning particles when an entity is within warningMargin of a zone edge.
func (s *HazardProximityWarningParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.hazardSystem == nil || s.world == nil {
		return
	}

	tracker := s.hazardSystem.GetZoneTracker()
	if tracker == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.emitInterval {
		return
	}
	s.timeSinceEmit = 0

	// Decay per-entity cooldowns
	for id, t := range s.lastWarned {
		s.lastWarned[id] = t - s.emitInterval
		if s.lastWarned[id] <= 0 {
			delete(s.lastWarned, id)
		}
	}

	for _, entity := range entities {
		// Only warn entities that can be damaged
		if entity.GetHealth() == nil {
			continue
		}
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Skip if recently warned
		if cd, ok := s.lastWarned[entity.ID]; ok && cd > 0 {
			continue
		}

		// Query zones within warning radius (entity position + margin)
		zones := tracker.GetZonesInRadius(pos.X, pos.Y, s.warningMargin)
		if len(zones) == 0 {
			continue
		}

		// Find the closest zone the entity is NOT fully inside
		for _, zone := range zones {
			dx := pos.X - zone.X
			dy := pos.Y - zone.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			// Entity is inside the hazard — skip (hazard system handles damage already)
			if dist <= zone.Radius*0.8 {
				continue
			}

			// Entity is in the warning band: between zone.Radius*0.8 and zone.Radius+warningMargin
			s.spawnWarningParticles(pos.X, pos.Y, zone, dx, dy, dist, entity.ID)
			s.lastWarned[entity.ID] = s.emitInterval * 2

			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entity_id":   entity.ID,
					"hazard_type": zone.HazardType.String(),
					"distance":    dist,
				}).Debug("hazard proximity warning spawned")
			}
			break // One warning per entity per cycle
		}
	}
}

// spawnWarningParticles creates particles between entity and hazard zone.
func (s *HazardProximityWarningParticleSystem) spawnWarningParticles(
	ex, ey float64, zone *HazardZone, dx, dy, dist float64, entityID uint64,
) {
	preset := s.getGenrePreset()

	// Normalize direction from hazard to entity
	ndx, ndy := 0.0, 0.0
	if dist > 0.001 {
		ndx = dx / dist
		ndy = dy / dist
	}

	// Spawn particles on the side of the entity facing the hazard
	spawnX := ex - ndx*8
	spawnY := ey - ndy*8

	effectSeed := s.seed + int64(entityID*997) + int64(zone.ID*773)
	count := s.baseCount + s.hazardCountBonus(zone.HazardType)

	config := particles.Config{
		Type:     s.hazardParticleType(zone.HazardType, preset.ParticleType),
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: preset.Duration,
		SpreadX:  preset.SpreadFactor,
		SpreadY:  preset.SpreadFactor,
		Gravity:  preset.Gravity,
		MinSize:  preset.MinSize,
		MaxSize:  preset.MaxSize,
		Custom:   make(map[string]interface{}),
	}
	config.Custom["hazard_warning"] = true
	config.Custom["hazard_type"] = zone.HazardType.String()

	s.particleSystem.SpawnParticles(s.world, config, spawnX, spawnY)
}

// hazardParticleType overrides particle type for specific hazard types.
func (s *HazardProximityWarningParticleSystem) hazardParticleType(
	ht HazardType, defaultType particles.ParticleType,
) particles.ParticleType {
	switch ht {
	case HazardPoison:
		return particles.ParticleSmoke
	case HazardOil:
		return particles.ParticleDebris
	case HazardSmoke:
		return particles.ParticleSmoke
	default:
		return defaultType
	}
}

// hazardCountBonus returns extra particle count for more dangerous hazards.
func (s *HazardProximityWarningParticleSystem) hazardCountBonus(ht HazardType) int {
	switch ht {
	case HazardPoison:
		return 2
	case HazardOil:
		return 1
	default:
		return 0
	}
}

// getGenrePreset returns genre-specific particle parameters.
func (s *HazardProximityWarningParticleSystem) getGenrePreset() hazardWarningPreset {
	switch s.genreID {
	case "horror":
		return hazardWarningPreset{
			ParticleType: particles.ParticleSmoke,
			MinSize:      3.0,
			MaxSize:      7.0,
			Duration:     1.5,
			Gravity:      -20.0,
			SpreadFactor: 24.0,
		}
	case "scifi":
		return hazardWarningPreset{
			ParticleType: particles.ParticleSpark,
			MinSize:      1.5,
			MaxSize:      4.0,
			Duration:     0.8,
			Gravity:      -60.0,
			SpreadFactor: 18.0,
		}
	case "cyberpunk":
		return hazardWarningPreset{
			ParticleType: particles.ParticleSpark,
			MinSize:      2.0,
			MaxSize:      5.0,
			Duration:     0.9,
			Gravity:      -40.0,
			SpreadFactor: 20.0,
		}
	case "postapoc":
		return hazardWarningPreset{
			ParticleType: particles.ParticleDebris,
			MinSize:      2.5,
			MaxSize:      6.0,
			Duration:     1.3,
			Gravity:      -15.0,
			SpreadFactor: 22.0,
		}
	default: // fantasy
		return hazardWarningPreset{
			ParticleType: particles.ParticleMagic,
			MinSize:      2.0,
			MaxSize:      5.0,
			Duration:     1.0,
			Gravity:      -50.0,
			SpreadFactor: 20.0,
		}
	}
}
