// Package engine provides the HeadgearAssignmentSystem which assigns seed-based
// headgear types to humanoid entities. The system reads entity role data and
// genre context to select appropriate headgear, then attaches HeadgearComponent
// so the sprite generation pipeline can render the overlay. Non-humanoid entities
// are skipped. The headgear type is fed into Config.Custom["headgearType"] during
// sprite generation.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// HeadgearAssignmentSystem scans entities with sprite components and assigns
// headgear types based on entity role, genre, and deterministic seed.
type HeadgearAssignmentSystem struct {
	world  *World
	logger *logrus.Entry
	seed   int64

	genreID string

	scanInterval  float64
	timeSinceScan float64
}

// NewHeadgearAssignmentSystem creates a new headgear assignment system.
func NewHeadgearAssignmentSystem(world *World, seed int64) *HeadgearAssignmentSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "headgear_assignment")
		logEntry.Debug("headgear assignment system created")
	}

	return &HeadgearAssignmentSystem{
		world:         world,
		logger:        logEntry,
		seed:          seed,
		genreID:       "fantasy",
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
}

// SetGenre configures genre-aware headgear selection.
func (s *HeadgearAssignmentSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for headgear assignment")
	}
}

// Update scans entities and attaches headgear components to humanoid entities.
func (s *HeadgearAssignmentSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity == nil || !entity.HasComponent("sprite") {
			continue
		}

		// Check for existing headgear component
		comp, has := entity.GetComponent("headgear")
		if has {
			existing, ok := comp.(*HeadgearComponent)
			if ok && existing.Genre == s.genreID {
				continue // Already assigned with current genre
			}
			// Genre changed, reassign
			if ok {
				s.assignHeadgear(existing, entity)
			}
			continue
		}

		// Determine entity type for role mapping
		entityType := s.resolveEntityType(entity)
		if entityType == "" {
			continue
		}

		// Only assign headgear to humanoid entities
		if !sprites.IsHumanoidEntity(entityType) {
			continue
		}

		hg := &HeadgearComponent{}
		s.assignHeadgear(hg, entity)
		entity.AddComponent(hg)
	}
}

// assignHeadgear populates a HeadgearComponent from entity seed and role.
func (s *HeadgearAssignmentSystem) assignHeadgear(hg *HeadgearComponent, entity *Entity) {
	entitySeed := s.seed ^ int64(entity.ID)
	role := s.resolveRole(entity)

	hgType := sprites.SelectHeadgearForRole(role, s.genreID, entitySeed)
	hg.HeadgearType = int(hgType)
	hg.Genre = s.genreID
	hg.Role = role
}

// resolveEntityType determines the entity type string from component inspection.
func (s *HeadgearAssignmentSystem) resolveEntityType(entity *Entity) string {
	// Check if entity has NPC-related components
	if entity.HasComponent("ai") || entity.HasComponent("dialog") {
		return "npc"
	}
	if entity.HasComponent("player_input") {
		return "player"
	}
	// Entities with health and sprite are likely game entities
	if entity.HasComponent("health") && entity.HasComponent("sprite") {
		return "humanoid"
	}

	return ""
}

// resolveRole determines the visual role string for headgear selection.
func (s *HeadgearAssignmentSystem) resolveRole(entity *Entity) string {
	entityType := s.resolveEntityType(entity)
	role := sprites.MapEntityTypeToRole(entityType)
	if role != "" {
		return string(role)
	}

	// Use seed to provide a generic NPC role for variety
	rng := rand.New(rand.NewSource(s.seed ^ int64(entity.ID) ^ 0xBEAD))
	genericRoles := []string{"npc", "merchant", "warrior", "mage", "rogue"}
	return genericRoles[rng.Intn(len(genericRoles))]
}
