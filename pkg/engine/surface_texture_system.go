// Package engine provides the SurfaceTextureSystem which assigns creature-form-
// specific procedural surface textures to entities. It reads the
// CreatureVisualComponent to determine what form the entity has (quadruped,
// serpentine, arachnid, etc.) and populates a SurfaceTextureComponent with the
// appropriate texture type, colors, and intensities derived from the entity seed.
// Genre changes trigger re-population.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// SurfaceTextureSystem scans entities with creature_visual components and
// attaches/updates SurfaceTextureComponent with form-appropriate surface
// texture parameters.
type SurfaceTextureSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	scanInterval  float64
	timeSinceScan float64
}

// NewSurfaceTextureSystem creates a new surface texture system.
func NewSurfaceTextureSystem(world *World, seed int64) *SurfaceTextureSystem {
	var logEntry *logrus.Entry
	if world != nil && world.GetLogger() != nil {
		logEntry = world.GetLogger().WithField("system_name", "surface_texture")
		logEntry.Debug("surface texture system created")
	}

	return &SurfaceTextureSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
}

// SetGenre configures genre-aware texture preferences.
func (s *SurfaceTextureSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for surface textures")
	}
}

// Update scans entities and attaches/configures surface texture components.
func (s *SurfaceTextureSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		// Only process entities that have a creature visual classification
		cvComp, hasCv := entity.GetComponent("creature_visual")
		if !hasCv {
			continue
		}
		cv, ok := cvComp.(*CreatureVisualComponent)
		if !ok {
			continue
		}

		// Skip humanoids (they use clothing patterns, not surface textures)
		if cv.Form == FormHumanoid {
			continue
		}

		comp, has := entity.GetComponent("surface_texture")
		if !has {
			st := NewSurfaceTextureComponent()
			s.populateFromForm(st, entity, cv)
			entity.AddComponent(st)
			continue
		}

		existing, ok := comp.(*SurfaceTextureComponent)
		if !ok {
			continue
		}

		// Re-populate if genre changed
		if existing.GenreID != s.genreID {
			s.populateFromForm(existing, entity, cv)
			existing.Dirty = true
		}
	}
}

// populateFromForm fills a SurfaceTextureComponent using creature form and seed.
func (s *SurfaceTextureSystem) populateFromForm(st *SurfaceTextureComponent, entity *Entity, cv *CreatureVisualComponent) {
	seed := int64(entity.ID)
	texSet := sprites.GenerateSurfaceTextureSet(seed, string(cv.Form), s.genreID)

	st.TextureType = int(texSet.TorsoTexture.Type)

	st.HeadIntensity = texSet.HeadTexture.Intensity
	st.TorsoIntensity = texSet.TorsoTexture.Intensity
	st.LimbIntensity = texSet.LimbTexture.Intensity

	st.HeadScale = texSet.HeadTexture.Scale
	st.TorsoScale = texSet.TorsoTexture.Scale
	st.LimbScale = texSet.LimbTexture.Scale

	st.PrimaryR = texSet.TorsoTexture.PrimaryColor.R
	st.PrimaryG = texSet.TorsoTexture.PrimaryColor.G
	st.PrimaryB = texSet.TorsoTexture.PrimaryColor.B
	st.PrimaryA = texSet.TorsoTexture.PrimaryColor.A

	st.SecondaryR = texSet.TorsoTexture.SecondaryColor.R
	st.SecondaryG = texSet.TorsoTexture.SecondaryColor.G
	st.SecondaryB = texSet.TorsoTexture.SecondaryColor.B
	st.SecondaryA = texSet.TorsoTexture.SecondaryColor.A

	st.GenreID = s.genreID
	st.Enabled = true

	// Mark animation dirty so sprite regenerates with texture
	if animComp := entity.GetAnimation(); animComp != nil {
		animComp.Dirty = true
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"form":         string(cv.Form),
			"texture_type": sprites.SurfaceTextureType(st.TextureType).String(),
		}).Debug("surface texture assigned")
	}
}

// GetActiveTextureCount returns the number of entities with active surface textures.
func (s *SurfaceTextureSystem) GetActiveTextureCount() int {
	if s.world == nil {
		return 0
	}
	count := 0
	for _, entity := range s.world.GetEntities() {
		if comp, has := entity.GetComponent("surface_texture"); has {
			if st, ok := comp.(*SurfaceTextureComponent); ok && st.Enabled {
				count++
			}
		}
	}
	return count
}
