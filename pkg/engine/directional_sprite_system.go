// Package engine provides the DirectionalSpriteSystem which generates
// 4-directional sprite variants (up/down/left/right) for entities with
// sprite and animation components. The render system selects the correct
// directional image based on entity facing, producing visually correct
// top-down sprites that change with movement direction.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// maxDirectionalGenerationsPerFrame limits how many entities get directional
// sprites generated in a single frame to avoid frame drops.
const maxDirectionalGenerationsPerFrame = 4

// DirectionalSpriteSystem generates 4-directional sprite variants for entities.
// It runs after AnimationSystem (which creates the base sprite) and before
// SpriteFinalizerSystem (which applies outlines and rim lighting).
// Entities with populated DirectionalImages will have the render system
// automatically select the correct sprite based on their facing direction.
type DirectionalSpriteSystem struct {
	spriteGenerator *sprites.Generator
	logger          *logrus.Entry
	genreID         string

	// processed tracks entity IDs that already have directional sprites.
	// Cleared when genre changes to allow re-generation with new palette.
	processed map[uint64]int64 // entityID -> seed at generation time
}

// NewDirectionalSpriteSystem creates a new directional sprite generation system.
func NewDirectionalSpriteSystem(spriteGenerator *sprites.Generator) *DirectionalSpriteSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "directional_sprite",
	})

	sys := &DirectionalSpriteSystem{
		spriteGenerator: spriteGenerator,
		logger:          logger,
		processed:       make(map[uint64]int64, 64),
	}

	logger.Debug("directional sprite system created")
	return sys
}

// SetGenre updates the genre used for sprite generation, clearing the cache
// so entities get re-generated with the new genre palette.
func (s *DirectionalSpriteSystem) SetGenre(genreID string) {
	if s.genreID != genreID {
		s.genreID = genreID
		s.processed = make(map[uint64]int64, 64)
		s.logger.WithField("genre", genreID).Debug("genre changed, clearing directional cache")
	}
}

// Update generates directional sprites for entities that need them.
func (s *DirectionalSpriteSystem) Update(entities []*Entity, deltaTime float64) {
	generated := 0

	for _, entity := range entities {
		if generated >= maxDirectionalGenerationsPerFrame {
			break
		}

		spriteComp := s.getSpriteComponent(entity)
		if spriteComp == nil || spriteComp.Image == nil {
			continue
		}

		animComp := s.getAnimationComponent(entity)
		if animComp == nil {
			continue
		}

		// Check if this entity already has current directional sprites
		if s.isAlreadyProcessed(entity, animComp) {
			continue
		}

		// Generate directional sprites
		if err := s.generateDirectionalSprites(entity, spriteComp, animComp); err != nil {
			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"seed":      animComp.Seed,
				}).WithError(err).Debug("directional sprite generation failed")
			}
			continue
		}

		generated++
	}
}

// isAlreadyProcessed checks if an entity already has up-to-date directional sprites.
func (s *DirectionalSpriteSystem) isAlreadyProcessed(entity *Entity, anim *AnimationComponent) bool {
	prevSeed, exists := s.processed[entity.ID]
	if !exists {
		return false
	}
	// Re-generate if the animation seed changed (entity was regenerated)
	return prevSeed == anim.Seed
}

// generateDirectionalSprites creates 4-directional sprites and stores them on the entity.
func (s *DirectionalSpriteSystem) generateDirectionalSprites(entity *Entity, spriteComp *EbitenSprite, animComp *AnimationComponent) error {
	config := s.buildSpriteConfig(entity, spriteComp, animComp)

	dirSprites, err := s.spriteGenerator.GenerateDirectionalSprites(config)
	if err != nil {
		return err
	}

	// Store directional sprites on the sprite component
	if spriteComp.DirectionalImages == nil {
		spriteComp.DirectionalImages = make(map[int]*ebiten.Image, 4)
	}
	for dir, img := range dirSprites {
		spriteComp.DirectionalImages[dir] = img
	}

	// Reset finalized flag so SpriteFinalizerSystem re-processes these sprites
	spriteComp.Finalized = false
	// Reset depth flag so SpriteDepthEnhanceSystem re-processes
	spriteComp.DepthProcessed = false
	// Reset color temperature flag so SpriteColorTemperatureSystem re-processes
	spriteComp.ColorTempProcessed = false

	// Track this entity as processed
	s.processed[entity.ID] = animComp.Seed

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"seed":       animComp.Seed,
			"directions": len(dirSprites),
		}).Debug("directional sprites generated")
	}

	return nil
}

// buildSpriteConfig creates a sprite Config for directional generation,
// mirroring the AnimationSystem's config building approach.
func (s *DirectionalSpriteSystem) buildSpriteConfig(entity *Entity, spriteComp *EbitenSprite, animComp *AnimationComponent) sprites.Config {
	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      int(spriteComp.Width),
		Height:     int(spriteComp.Height),
		Seed:       animComp.Seed,
		Complexity: 0.7,
		GenreID:    s.getGenreID(entity),
		Custom:     make(map[string]interface{}),
	}

	config.Custom["useAerial"] = true

	entityType := s.determineEntityType(entity)
	config.Custom["entityType"] = entityType

	// Pass avatar traits seed for consistent variety across directions
	config.Custom["avatarSeed"] = animComp.Seed

	// Configure player-specific settings
	if entity.HasComponent("input") {
		config.Custom["entityType"] = "humanoid"
		if entity.HasComponent("equipment") {
			config.Custom["hasWeapon"] = true
			config.Custom["hasShield"] = false
		}
	}

	return config
}

// determineEntityType returns the entity type string for template selection.
func (s *DirectionalSpriteSystem) determineEntityType(entity *Entity) string {
	// Check CreatureVisualComponent for classified form
	if cvComp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := cvComp.(*CreatureVisualComponent); ok && cv.Form != FormHumanoid {
			return string(cv.Form)
		}
	}

	// Check NPC role visual component for role-specific humanoid template
	if roleComp, ok := entity.GetComponent("npc_role_visual"); ok {
		if npcRole, ok := roleComp.(*NpcRoleVisualComponent); ok && npcRole.Role != "" {
			return npcRole.Role
		}
	}

	// Check for boss-level entities
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			if attack.Damage >= 50 {
				return "boss"
			}
		}
	}

	// Check size-based type
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			if health.Max >= 500 {
				return "quadruped"
			}
		}
	}

	return "humanoid"
}

// getGenreID returns the genre for this entity, falling back to system default.
func (s *DirectionalSpriteSystem) getGenreID(entity *Entity) string {
	if genreComp, hasGenre := entity.GetComponent("genre"); hasGenre {
		if genre, ok := genreComp.(*GenreComponent); ok {
			return genre.GetPrimaryGenre()
		}
	}
	if s.genreID != "" {
		return s.genreID
	}
	return "fantasy"
}

// getSpriteComponent retrieves the sprite component from an entity.
func (s *DirectionalSpriteSystem) getSpriteComponent(entity *Entity) *EbitenSprite {
	return entity.GetSprite()
}

// getAnimationComponent retrieves the animation component from an entity.
func (s *DirectionalSpriteSystem) getAnimationComponent(entity *Entity) *AnimationComponent {
	comp, ok := entity.GetComponent("animation")
	if !ok {
		return nil
	}
	anim, ok := comp.(*AnimationComponent)
	if !ok {
		return nil
	}
	return anim
}

// GetProcessedCount returns the number of entities with generated directional sprites.
func (s *DirectionalSpriteSystem) GetProcessedCount() int {
	return len(s.processed)
}

// ClearCache resets the processed entity cache, forcing re-generation.
func (s *DirectionalSpriteSystem) ClearCache() {
	s.processed = make(map[uint64]int64, 64)
}
