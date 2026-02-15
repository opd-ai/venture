// Package engine provides the EquipmentGleamSweepSystem which reads per-entity
// MaterialSheenComponent data and equipment rarity to animate a genre-aware
// specular highlight sweep across equipped items. Metal gets sharp fast sweeps,
// crystal gets prismatic slow sweeps, cloth/leather get subtle diffuse sweeps.
// This bridges EquipmentMaterialSheenSystem output with an animated gleam band.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// gleamMaterialProfile controls sweep behavior per dominant material.
type gleamMaterialProfile struct {
	SpeedMult    float64 // Multiplier on base sweep speed
	WidthMult    float64 // Multiplier on sweep band width
	IntensityAdd float64 // Additive intensity bonus
}

var gleamMaterialProfiles = map[string]gleamMaterialProfile{
	"Metal":   {SpeedMult: 1.4, WidthMult: 0.7, IntensityAdd: 0.15},
	"Crystal": {SpeedMult: 0.6, WidthMult: 1.3, IntensityAdd: 0.10},
	"Energy":  {SpeedMult: 1.8, WidthMult: 0.5, IntensityAdd: 0.20},
	"Leather": {SpeedMult: 0.8, WidthMult: 1.1, IntensityAdd: 0.00},
	"Cloth":   {SpeedMult: 0.7, WidthMult: 1.2, IntensityAdd: -0.05},
	"Wood":    {SpeedMult: 0.9, WidthMult: 1.0, IntensityAdd: 0.00},
}

// gleamGenrePreset holds genre-specific sweep color tinting and timing.
type gleamGenrePreset struct {
	R, G, B       float64
	CooldownScale float64 // Multiplier on cooldown duration
	IntensityMult float64 // Multiplier on overall intensity
}

// EquipmentGleamSweepSystem animates a specular highlight sweep across entities
// with equipped items. It reads MaterialSheenComponent for material data and
// writes GleamSweepComponent with animated sweep state.
type EquipmentGleamSweepSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  gleamGenrePreset

	// Throttle new-entity detection to avoid per-frame scans
	scanInterval   float64
	timeSinceScan  float64
	baseCooldown   float64
	baseSweepSpeed float64
}

// NewEquipmentGleamSweepSystem creates a new equipment gleam sweep system.
func NewEquipmentGleamSweepSystem(world *World, seed int64) *EquipmentGleamSweepSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_gleam_sweep")
		logEntry.Debug("equipment gleam sweep system created")
	}

	sys := &EquipmentGleamSweepSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		scanInterval:   1.0,
		baseCooldown:   4.0,
		baseSweepSpeed: 0.4,
	}
	sys.preset = sys.getGenrePreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware sweep color and timing.
func (s *EquipmentGleamSweepSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for gleam sweep")
	}
}

// getGenrePreset returns genre-specific gleam sweep parameters.
func (s *EquipmentGleamSweepSystem) getGenrePreset(genreID string) gleamGenrePreset {
	switch genreID {
	case "horror":
		return gleamGenrePreset{R: 0.85, G: 0.6, B: 0.5, CooldownScale: 1.5, IntensityMult: 0.5}
	case "cyberpunk":
		return gleamGenrePreset{R: 0.3, G: 1.0, B: 0.95, CooldownScale: 0.6, IntensityMult: 1.4}
	case "sci-fi", "scifi":
		return gleamGenrePreset{R: 0.7, G: 0.85, B: 1.0, CooldownScale: 0.8, IntensityMult: 1.2}
	case "post-apocalyptic", "postapoc":
		return gleamGenrePreset{R: 1.0, G: 0.8, B: 0.5, CooldownScale: 1.3, IntensityMult: 0.6}
	case "fantasy":
		return gleamGenrePreset{R: 1.0, G: 0.95, B: 0.85, CooldownScale: 1.0, IntensityMult: 1.0}
	default:
		return gleamGenrePreset{R: 1.0, G: 1.0, B: 1.0, CooldownScale: 1.0, IntensityMult: 1.0}
	}
}

// Update advances gleam sweep animations and attaches components to new entities.
func (s *EquipmentGleamSweepSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	fullScan := false
	if s.timeSinceScan >= s.scanInterval {
		s.timeSinceScan = 0
		fullScan = true
	}

	for _, entity := range entities {
		comp, _ := entity.GetComponent("gleam_sweep")
		gleam, hasGleam := comp.(*GleamSweepComponent)

		if !hasGleam {
			if !fullScan {
				continue
			}
			// Only attach to entities with material sheen data
			if !entity.HasComponent("material_sheen") {
				continue
			}
			gleam = NewGleamSweepComponent()
			entity.AddComponent(gleam)
			s.configureGleam(entity, gleam)
		}

		if !gleam.Enabled {
			if fullScan {
				s.configureGleam(entity, gleam)
			}
			continue
		}

		s.advanceSweep(gleam, deltaTime)
	}
}

// configureGleam reads MaterialSheenComponent to set sweep parameters.
func (s *EquipmentGleamSweepSystem) configureGleam(entity *Entity, gleam *GleamSweepComponent) {
	sheenComp, ok := entity.GetComponent("material_sheen")
	if !ok {
		gleam.Enabled = false
		return
	}
	sheen, ok := sheenComp.(*MaterialSheenComponent)
	if !ok || !sheen.Enabled {
		gleam.Enabled = false
		return
	}

	// Base intensity from material sheen
	baseIntensity := sheen.SheenIntensity * sheen.Reflectivity

	// Apply material-specific profile
	profile := gleamMaterialProfiles[sheen.DominantMaterial]
	if profile.SpeedMult == 0 {
		// Unknown material: use neutral defaults
		profile = gleamMaterialProfile{SpeedMult: 1.0, WidthMult: 1.0, IntensityAdd: 0.0}
	}

	intensity := (baseIntensity + profile.IntensityAdd) * s.preset.IntensityMult
	intensity = clampGleamFloat(intensity, 0.0, 1.0)

	// Skip if intensity too low to be visible
	if intensity < 0.05 {
		gleam.Enabled = false
		return
	}

	gleam.SweepSpeed = s.baseSweepSpeed * profile.SpeedMult
	gleam.SweepWidth = 0.15 * profile.WidthMult
	gleam.Intensity = intensity
	gleam.CooldownDuration = s.baseCooldown * s.preset.CooldownScale

	// Genre-tinted color blended with material sheen color
	gleam.ColorR = (s.preset.R + sheen.ColorR) * 0.5
	gleam.ColorG = (s.preset.G + sheen.ColorG) * 0.5
	gleam.ColorB = (s.preset.B + sheen.ColorB) * 0.5

	gleam.MaterialHint = sheen.DominantMaterial
	gleam.Enabled = true
}

// advanceSweep animates the sweep position or ticks cooldown.
func (s *EquipmentGleamSweepSystem) advanceSweep(gleam *GleamSweepComponent, dt float64) {
	if gleam.Active {
		gleam.SweepPosition += gleam.SweepSpeed * dt
		if gleam.SweepPosition > 1.0+gleam.SweepWidth {
			// Sweep finished, enter cooldown
			gleam.Active = false
			gleam.SweepPosition = 0.0
			// Jitter cooldown ±20% for visual variety
			jitter := 0.8 + s.rng.Float64()*0.4
			gleam.CooldownRemaining = gleam.CooldownDuration * jitter
		}
	} else {
		gleam.CooldownRemaining -= dt
		if gleam.CooldownRemaining <= 0 {
			gleam.Active = true
			gleam.SweepPosition = -gleam.SweepWidth
		}
	}
}

// clampGleamFloat clamps v to [min, max].
func clampGleamFloat(v, min, max float64) float64 {
	return math.Min(math.Max(v, min), max)
}
