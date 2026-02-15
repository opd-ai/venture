// Package engine provides the ClothingPatternSystem which manages per-entity
// clothing pattern state. It attaches ClothingPatternComponent to entities that
// have sprites, populates pattern data from deterministic seed-based generation,
// and marks entities dirty when genre changes require pattern regeneration.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// clothingGenrePreset holds genre-specific pattern probability and intensity biases.
type clothingGenrePreset struct {
	// PatternProbability is the likelihood an entity gets any pattern (0-1).
	PatternProbability float64
	// IntensityBias is added to base intensity (-0.2 to +0.2).
	IntensityBias float64
	// PreferredTypes lists pattern types favoured for this genre (nil = all equal).
	PreferredTypes []int
}

// ClothingPatternSystem scans entities with sprite components and attaches/updates
// ClothingPatternComponent with seed-derived clothing patterns.
type ClothingPatternSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  clothingGenrePreset

	scanInterval  float64
	timeSinceScan float64
}

// NewClothingPatternSystem creates a new clothing pattern system.
func NewClothingPatternSystem(world *World, seed int64) *ClothingPatternSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "clothing_pattern")
		logEntry.Debug("clothing pattern system created")
	}

	sys := &ClothingPatternSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
	sys.preset = sys.getGenrePreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware pattern preferences.
func (s *ClothingPatternSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for clothing patterns")
	}
}

// getGenrePreset returns genre-specific clothing pattern parameters.
func (s *ClothingPatternSystem) getGenrePreset(genreID string) clothingGenrePreset {
	switch genreID {
	case "horror":
		return clothingGenrePreset{
			PatternProbability: 0.35,
			IntensityBias:      -0.10,
			PreferredTypes:     []int{int(sprites.PatternBorder), int(sprites.PatternGradientV)},
		}
	case "cyberpunk":
		return clothingGenrePreset{
			PatternProbability: 0.80,
			IntensityBias:      0.10,
			PreferredTypes:     []int{int(sprites.PatternHStripes), int(sprites.PatternVStripes), int(sprites.PatternDiamondLattice)},
		}
	case "sci-fi", "scifi":
		return clothingGenrePreset{
			PatternProbability: 0.65,
			IntensityBias:      0.05,
			PreferredTypes:     []int{int(sprites.PatternVStripes), int(sprites.PatternBorder), int(sprites.PatternGradientV)},
		}
	case "post-apocalyptic", "postapoc":
		return clothingGenrePreset{
			PatternProbability: 0.40,
			IntensityBias:      -0.05,
			PreferredTypes:     []int{int(sprites.PatternBorder), int(sprites.PatternHerringbone)},
		}
	default: // fantasy
		return clothingGenrePreset{
			PatternProbability: 0.60,
			IntensityBias:      0.0,
			PreferredTypes:     nil,
		}
	}
}

// Update scans entities and attaches/configures clothing pattern components.
func (s *ClothingPatternSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		if !entity.HasComponent("sprite") {
			continue
		}

		comp, has := entity.GetComponent("clothing_pattern")
		if !has {
			cp := NewClothingPatternComponent()
			s.populateFromSeed(cp, entity)
			entity.AddComponent(cp)
			continue
		}

		existing, ok := comp.(*ClothingPatternComponent)
		if !ok {
			continue
		}

		if existing.GenreID != s.genreID {
			s.populateFromSeed(existing, entity)
			existing.GenreID = s.genreID
			existing.Dirty = true
		}
	}
}

// populateFromSeed fills a ClothingPatternComponent using the entity's ID as seed.
func (s *ClothingPatternSystem) populateFromSeed(cp *ClothingPatternComponent, entity *Entity) {
	seed := int64(entity.ID)
	patternSet := sprites.GenerateClothingPatternSet(seed)

	s.applyPattern(cp, "torso", patternSet.TorsoPattern)
	s.applyPattern(cp, "arm", patternSet.ArmPattern)
	s.applyPattern(cp, "leg", patternSet.LegPattern)

	cp.GenreID = s.genreID
	cp.Enabled = true
}

// applyPattern writes a single pattern to the appropriate component fields.
func (s *ClothingPatternSystem) applyPattern(cp *ClothingPatternComponent, region string, p sprites.ClothingPattern) {
	ptype := int(p.Type)
	intensity := p.Intensity + s.preset.IntensityBias
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1.0 {
		intensity = 1.0
	}

	switch region {
	case "torso":
		cp.TorsoPatternType = ptype
		cp.TorsoPatternR = p.PatternColor.R
		cp.TorsoPatternG = p.PatternColor.G
		cp.TorsoPatternB = p.PatternColor.B
		cp.TorsoScale = p.Scale
		cp.TorsoIntensity = intensity
	case "arm":
		cp.ArmPatternType = ptype
		cp.ArmPatternR = p.PatternColor.R
		cp.ArmPatternG = p.PatternColor.G
		cp.ArmPatternB = p.PatternColor.B
		cp.ArmScale = p.Scale
		cp.ArmIntensity = intensity
	case "leg":
		cp.LegPatternType = ptype
		cp.LegPatternR = p.PatternColor.R
		cp.LegPatternG = p.PatternColor.G
		cp.LegPatternB = p.PatternColor.B
		cp.LegScale = p.Scale
		cp.LegIntensity = intensity
	}
}
