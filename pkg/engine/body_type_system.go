// Package engine provides the BodyTypeSystem which assigns seed-derived body
// types to humanoid entities. It attaches BodyTypeComponent and feeds the
// body type index into the sprite generation pipeline via Config.Custom,
// producing dramatically different silhouettes per entity.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// BodyTypeSystem scans entities with sprite components and attaches/updates
// BodyTypeComponent with seed-derived body type assignments.
type BodyTypeSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	scanInterval  float64
	timeSinceScan float64
}

// NewBodyTypeSystem creates a new body type system.
func NewBodyTypeSystem(world *World, seed int64) *BodyTypeSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "body_type")
		logEntry.Debug("body type system created")
	}

	return &BodyTypeSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
}

// SetGenre configures genre-aware body type distribution.
func (s *BodyTypeSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for body type system")
	}
}

// Update scans entities and attaches body type components.
func (s *BodyTypeSystem) Update(entities []*Entity, deltaTime float64) {
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

		comp, has := entity.GetComponent("body_type")
		if !has {
			bt := NewBodyTypeComponent()
			s.populateFromSeed(bt, entity)
			entity.AddComponent(bt)
			continue
		}

		existing, ok := comp.(*BodyTypeComponent)
		if !ok {
			continue
		}

		// Re-derive if genre changed
		if existing.GenreID != s.genreID {
			s.populateFromSeed(existing, entity)
			existing.Dirty = true
		}
	}
}

// populateFromSeed derives a body type from the entity's ID.
func (s *BodyTypeSystem) populateFromSeed(bt *BodyTypeComponent, entity *Entity) {
	seed := int64(entity.ID)
	derived := sprites.DeriveBodyType(seed, s.genreID)
	bt.BodyType = int(derived)
	bt.GenreID = s.genreID
	bt.Assigned = true
}
