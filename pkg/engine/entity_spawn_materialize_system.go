// Package engine provides the EntitySpawnMaterializeSystem for genre-aware
// visual materialization effects when entities first appear in the world.
// New entities fade in from transparent to opaque over a short duration,
// with genre-specific particle bursts during the materialization period.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpawnMaterializeComponent stores per-entity materialization state.
// Attached automatically when an entity spawns, removed once fully materialized.
type SpawnMaterializeComponent struct {
	// Elapsed time since spawn (seconds)
	Elapsed float64

	// Total materialization duration (seconds)
	Duration float64

	// Current opacity factor (0.0=invisible, 1.0=fully materialized)
	Opacity float64

	// Genre-driven particle color (RGB 0.0-1.0)
	ParticleR float64
	ParticleG float64
	ParticleB float64

	// Number of particles to emit during materialization
	ParticleCount int

	// Particles already emitted
	ParticlesEmitted int

	// Whether materialization is complete
	Complete bool
}

// Type returns the component type identifier.
func (c *SpawnMaterializeComponent) Type() string {
	return "spawn_materialize"
}

// genreMaterializePreset holds per-genre visual parameters.
type genreMaterializePreset struct {
	R, G, B       float64 // Particle color
	Duration      float64 // Materialization time (seconds)
	ParticleCount int     // Particle burst count
}

// EntitySpawnMaterializeSystem applies fade-in opacity and genre-aware
// particle color to newly-spawned entities. It detects entities that have
// a sprite but no spawn_materialize component (new spawns), attaches the
// component, then animates opacity from 0→1 over the materialization duration.
type EntitySpawnMaterializeSystem struct {
	world   *World
	logger  *logrus.Entry
	rng     *rand.Rand
	genreID string
	presets map[string]genreMaterializePreset

	// Track known entities to detect new spawns
	knownEntities map[uint64]bool

	// Throttle new-entity scanning
	scanInterval  float64
	timeSinceScan float64
	firstScanDone bool
}

// NewEntitySpawnMaterializeSystem creates the system with genre defaults.
func NewEntitySpawnMaterializeSystem(world *World, seed int64) *EntitySpawnMaterializeSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "entity_spawn_materialize")
	}

	sys := &EntitySpawnMaterializeSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		presets:       buildMaterializePresets(),
		knownEntities: make(map[uint64]bool, 256),
		scanInterval:  0.25,
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("entity spawn materialize system created")
	}
	return sys
}

// SetGenre configures the active genre for materialization visuals.
func (s *EntitySpawnMaterializeSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update animates materializing entities and detects new spawns.
func (s *EntitySpawnMaterializeSystem) Update(entities []*Entity, deltaTime float64) {
	// Animate active materializations every frame
	s.animateMaterializations(entities, deltaTime)

	// Throttle new-entity scanning
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval && s.firstScanDone {
		return
	}
	s.timeSinceScan = 0
	s.firstScanDone = true

	s.detectNewSpawns(entities)
}

// animateMaterializations updates opacity on entities with spawn_materialize.
func (s *EntitySpawnMaterializeSystem) animateMaterializations(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("spawn_materialize")
		if !ok {
			continue
		}
		mc, ok := comp.(*SpawnMaterializeComponent)
		if !ok || mc.Complete {
			continue
		}

		mc.Elapsed += deltaTime

		if mc.Duration <= 0 {
			mc.Opacity = 1.0
			mc.Complete = true
			continue
		}

		// Smooth ease-out: opacity = 1 - (1 - t)^2
		t := mc.Elapsed / mc.Duration
		if t >= 1.0 {
			t = 1.0
			mc.Complete = true
		}
		mc.Opacity = 1.0 - (1.0-t)*(1.0-t)

		// Emit particles proportional to progress
		targetEmitted := int(math.Floor(t * float64(mc.ParticleCount)))
		if targetEmitted > mc.ParticlesEmitted {
			mc.ParticlesEmitted = targetEmitted
		}
	}
}

// detectNewSpawns finds entities with a sprite component that are not yet known.
func (s *EntitySpawnMaterializeSystem) detectNewSpawns(entities []*Entity) {
	preset, ok := s.presets[s.genreID]
	if !ok {
		preset = s.presets["fantasy"]
	}

	newCount := 0
	for _, entity := range entities {
		id := entity.ID
		if s.knownEntities[id] {
			continue
		}
		s.knownEntities[id] = true

		// Only materialize entities with visible sprites
		if !entity.HasComponent("sprite") {
			continue
		}
		// Skip entities that already have the component (e.g., deserialized)
		if entity.HasComponent("spawn_materialize") {
			continue
		}

		// Per-entity variation ±15% on duration
		durationVariation := 1.0 + (s.rng.Float64()*0.3 - 0.15)
		duration := preset.Duration * durationVariation

		mc := &SpawnMaterializeComponent{
			Elapsed:       0,
			Duration:      duration,
			Opacity:       0,
			ParticleR:     clampMaterialize(preset.R + s.rng.Float64()*0.06 - 0.03),
			ParticleG:     clampMaterialize(preset.G + s.rng.Float64()*0.06 - 0.03),
			ParticleB:     clampMaterialize(preset.B + s.rng.Float64()*0.06 - 0.03),
			ParticleCount: preset.ParticleCount,
		}
		entity.AddComponent(mc)
		newCount++
	}

	if newCount > 0 && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":     s.genreID,
			"new_count": newCount,
		}).Debug("spawn materialize components assigned")
	}
}

// buildMaterializePresets returns genre-specific materialization visuals.
//
//   - Fantasy: golden sparkle (warm materialization)
//   - Horror: dark red smoke (emerging from shadows)
//   - Scifi: cyan digital assembly (teleport beam)
//   - Cyberpunk: neon magenta flicker (holographic)
//   - Postapoc: brown dust (debris settling)
func buildMaterializePresets() map[string]genreMaterializePreset {
	return map[string]genreMaterializePreset{
		"fantasy":   {R: 0.95, G: 0.80, B: 0.30, Duration: 0.6, ParticleCount: 8},
		"horror":    {R: 0.70, G: 0.10, B: 0.10, Duration: 0.9, ParticleCount: 6},
		"scifi":     {R: 0.20, G: 0.85, B: 0.95, Duration: 0.4, ParticleCount: 10},
		"cyberpunk": {R: 0.90, G: 0.20, B: 0.80, Duration: 0.35, ParticleCount: 10},
		"postapoc":  {R: 0.65, G: 0.50, B: 0.30, Duration: 0.7, ParticleCount: 5},
	}
}

// clampMaterialize ensures a value stays in [0.0, 1.0].
func clampMaterialize(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
