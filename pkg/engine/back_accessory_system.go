// Package engine provides the BackAccessorySystem which assigns seed-based
// back-worn accessories (capes, cloaks, quivers, backpacks, banners, scarves,
// wing-capes) to humanoid entities. The system reads entity role data and
// genre context to select appropriate accessories, then attaches
// BackAccessoryComponent so the sprite generation pipeline can render the
// overlay. Non-humanoid entities are skipped.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// BackAccessorySystem scans entities with sprite components and assigns
// back accessory types based on entity role, genre, and deterministic seed.
type BackAccessorySystem struct {
	world  *World
	logger *logrus.Entry
	seed   int64

	genreID string

	scanInterval  float64
	timeSinceScan float64
}

// NewBackAccessorySystem creates a new back accessory system.
func NewBackAccessorySystem(world *World, seed int64) *BackAccessorySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "back_accessory")
		logEntry.Debug("back accessory system created")
	}

	return &BackAccessorySystem{
		world:         world,
		logger:        logEntry,
		seed:          seed,
		genreID:       "fantasy",
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
}

// SetGenre configures genre-aware back accessory selection.
func (s *BackAccessorySystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for back accessory")
	}
}

// Update scans entities and attaches back accessory components to humanoid entities.
func (s *BackAccessorySystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity == nil || !entity.HasComponent("sprite") {
			continue
		}

		// Check for existing back accessory component
		comp, has := entity.GetComponent("back_accessory")
		if has {
			existing, ok := comp.(*BackAccessoryComponent)
			if ok && existing.Genre == s.genreID {
				continue // Already assigned with current genre
			}
			// Genre changed, reassign
			if ok {
				s.assignBackAccessory(existing, entity)
			}
			continue
		}

		// Determine entity type for role mapping
		entityType := s.resolveEntityType(entity)
		if entityType == "" {
			continue
		}

		// Only assign back accessories to humanoid entities
		if !sprites.IsHumanoidEntity(entityType) {
			continue
		}

		ba := &BackAccessoryComponent{}
		s.assignBackAccessory(ba, entity)
		entity.AddComponent(ba)
	}
}

// assignBackAccessory populates a BackAccessoryComponent from entity seed and role.
func (s *BackAccessorySystem) assignBackAccessory(ba *BackAccessoryComponent, entity *Entity) {
	entitySeed := s.seed ^ int64(entity.ID)
	role := s.resolveRole(entity)

	baType := sprites.SelectBackAccessoryForRole(role, s.genreID, entitySeed)
	ba.AccessoryType = int(baType)
	ba.Genre = s.genreID
	ba.Role = role
}

// resolveEntityType determines the entity type string from component inspection.
func (s *BackAccessorySystem) resolveEntityType(entity *Entity) string {
	if entity.HasComponent("ai") || entity.HasComponent("dialog") {
		return "npc"
	}
	if entity.HasComponent("input") {
		return "player"
	}
	if entity.HasComponent("health") && entity.HasComponent("sprite") {
		return "humanoid"
	}
	return ""
}

// resolveRole determines the visual role string for back accessory selection.
func (s *BackAccessorySystem) resolveRole(entity *Entity) string {
	entityType := s.resolveEntityType(entity)
	role := sprites.MapEntityTypeToRole(entityType)
	if role != "" {
		return string(role)
	}

	// Use seed to provide a generic NPC role for variety
	rng := rand.New(rand.NewSource(s.seed ^ int64(entity.ID) ^ 0xBACC))
	genericRoles := []string{"npc", "merchant", "warrior", "mage", "rogue"}
	return genericRoles[rng.Intn(len(genericRoles))]
}
