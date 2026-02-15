// Package engine provides the EntityDeathDissolveSystem for genre-aware
// visual dissolve effects when entities die. Dead entities fade from opaque
// to transparent over a configurable duration, with genre-specific dissolve
// colors and decay curves. This is the visual counterpart to
// EntitySpawnMaterializeSystem (spawn=fade-in, death=dissolve-out).
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// DeathDissolveComponent stores per-entity dissolve animation state.
// Attached when an entity's health reaches zero; removed once fully dissolved.
type DeathDissolveComponent struct {
	// Elapsed time since dissolve started (seconds)
	Elapsed float64

	// Total dissolve duration (seconds)
	Duration float64

	// Current opacity factor (1.0=fully visible, 0.0=fully dissolved)
	Opacity float64

	// Genre-driven dissolve tint color (RGB 0.0-1.0)
	TintR float64
	TintG float64
	TintB float64

	// Dissolve style index for rendering (0=fade, 1=burn, 2=glitch, 3=melt, 4=dust)
	Style int

	// Whether dissolve animation is complete
	Complete bool
}

// Type returns the component type identifier.
func (c *DeathDissolveComponent) Type() string {
	return "death_dissolve"
}

// dissolveGenrePreset holds per-genre dissolve visual parameters.
type dissolveGenrePreset struct {
	R, G, B  float64 // Dissolve tint color
	Duration float64 // Dissolve time (seconds)
	Style    int     // Dissolve style index
}

// EntityDeathDissolveSystem detects entities whose health has reached zero
// and applies a genre-aware dissolve animation, fading their opacity from
// 1.0 to 0.0 over a short duration.
type EntityDeathDissolveSystem struct {
	world   *World
	logger  *logrus.Entry
	rng     *rand.Rand
	genreID string
	presets map[string]dissolveGenrePreset

	// Throttle dead-entity scanning (not every frame)
	scanInterval  float64
	timeSinceScan float64
}

// NewEntityDeathDissolveSystem creates the system with genre defaults.
func NewEntityDeathDissolveSystem(world *World, seed int64) *EntityDeathDissolveSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "entity_death_dissolve")
	} else {
		logEntry = logrus.WithField("system_name", "entity_death_dissolve")
	}

	sys := &EntityDeathDissolveSystem{
		world:        world,
		logger:       logEntry,
		rng:          rand.New(rand.NewSource(seed)),
		genreID:      "fantasy",
		presets:      buildDissolvePresets(),
		scanInterval: 0.15,
	}

	if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("entity death dissolve system created")
	}
	return sys
}

// SetGenre configures the active genre for dissolve visuals.
func (s *EntityDeathDissolveSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update animates active dissolves and scans for newly dead entities.
func (s *EntityDeathDissolveSystem) Update(entities []*Entity, deltaTime float64) {
	s.animateDissolves(entities, deltaTime)

	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0
	s.detectDeadEntities(entities)
}

// animateDissolves updates opacity on entities with death_dissolve component.
func (s *EntityDeathDissolveSystem) animateDissolves(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("death_dissolve")
		if !ok {
			continue
		}
		dc, ok := comp.(*DeathDissolveComponent)
		if !ok || dc.Complete {
			continue
		}

		dc.Elapsed += deltaTime

		if dc.Duration <= 0 {
			dc.Opacity = 0.0
			dc.Complete = true
			continue
		}

		// Smooth ease-in curve: opacity = (1 - t)^2
		// Starts slow, accelerates toward end — feels like "dissolving away"
		t := dc.Elapsed / dc.Duration
		if t >= 1.0 {
			t = 1.0
			dc.Complete = true
		}
		dc.Opacity = (1.0 - t) * (1.0 - t)
	}
}

// detectDeadEntities finds entities with health ≤ 0 that need dissolve effects.
func (s *EntityDeathDissolveSystem) detectDeadEntities(entities []*Entity) {
	preset, ok := s.presets[s.genreID]
	if !ok {
		preset = s.presets["fantasy"]
	}

	newCount := 0
	for _, entity := range entities {
		// Must have both sprite and health to dissolve
		if !entity.HasComponent("sprite") || !entity.HasComponent("health") {
			continue
		}
		// Skip if already dissolving
		if entity.HasComponent("death_dissolve") {
			continue
		}

		hComp, ok := entity.GetComponent("health")
		if !ok {
			continue
		}
		hc, ok := hComp.(*HealthComponent)
		if !ok || hc.Current > 0 {
			continue
		}

		// Per-entity variation ±20% on duration
		durationVariation := 1.0 + (s.rng.Float64()*0.4 - 0.2)
		duration := preset.Duration * durationVariation

		// Small color jitter for visual variety
		dc := &DeathDissolveComponent{
			Elapsed:  0,
			Duration: math.Max(0.1, duration),
			Opacity:  1.0,
			TintR:    clampDissolve(preset.R + s.rng.Float64()*0.08 - 0.04),
			TintG:    clampDissolve(preset.G + s.rng.Float64()*0.08 - 0.04),
			TintB:    clampDissolve(preset.B + s.rng.Float64()*0.08 - 0.04),
			Style:    preset.Style,
		}
		entity.AddComponent(dc)
		newCount++
	}

	if newCount > 0 && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":     s.genreID,
			"new_count": newCount,
		}).Debug("death dissolve components assigned")
	}
}

// buildDissolvePresets returns genre-specific dissolve visuals.
//
//   - Fantasy: golden fade (magical dissipation)
//   - Horror: dark crimson melt (visceral horror)
//   - Scifi: cyan glitch (digital breakdown)
//   - Cyberpunk: neon magenta glitch (holographic failure)
//   - Postapoc: grey-brown dust (crumbling to ash)
func buildDissolvePresets() map[string]dissolveGenrePreset {
	return map[string]dissolveGenrePreset{
		"fantasy":   {R: 0.95, G: 0.85, B: 0.40, Duration: 0.8, Style: 0},  // fade
		"horror":    {R: 0.60, G: 0.05, B: 0.05, Duration: 1.2, Style: 3},  // melt
		"scifi":     {R: 0.15, G: 0.80, B: 0.90, Duration: 0.5, Style: 2},  // glitch
		"cyberpunk": {R: 0.85, G: 0.15, B: 0.75, Duration: 0.45, Style: 2}, // glitch
		"postapoc":  {R: 0.55, G: 0.45, B: 0.30, Duration: 1.0, Style: 4},  // dust
	}
}

// clampDissolve ensures a value stays in [0.0, 1.0].
func clampDissolve(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
