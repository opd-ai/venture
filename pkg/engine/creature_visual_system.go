// Package engine provides a creature visual classification system.
// CreatureVisualClassifierSystem scans entities that lack a CreatureVisualComponent
// and infers one from existing components (faction, stats, collider).
// This provides a fallback so entities spawned without explicit creature visual
// data still receive a reasonable visual classification.
package engine

import (
	"github.com/sirupsen/logrus"
)

// CreatureVisualClassifierSystem infers creature visual classification for
// entities that were not given an explicit CreatureVisualComponent at spawn time.
// It runs once per entity (idempotent — skips entities that already have the
// component) and sets a form based on heuristic analysis of other components.
type CreatureVisualClassifierSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewCreatureVisualClassifierSystem creates a new classifier system.
func NewCreatureVisualClassifierSystem(world *World) *CreatureVisualClassifierSystem {
	var entry *logrus.Entry
	if world != nil && world.GetLogger() != nil {
		entry = world.GetLogger().WithFields(logrus.Fields{
			"system_name": "creature_visual_classifier",
		})
	}
	return &CreatureVisualClassifierSystem{
		world:  world,
		logger: entry,
	}
}

// Update scans entities and assigns a CreatureVisualComponent to any enemy entity
// that doesn't already have one.
func (s *CreatureVisualClassifierSystem) Update(entities []*Entity, deltaTime float64) {
	for _, e := range entities {
		if e.HasComponent("creature_visual") {
			continue
		}
		// Only classify enemies (team 2) — players and NPCs are humanoid by default.
		teamComp, ok := e.GetComponent("team")
		if !ok {
			continue
		}
		team, ok := teamComp.(*TeamComponent)
		if !ok || team.TeamID != 2 {
			continue
		}

		form := s.inferForm(e)
		sizeClass := s.inferSizeClass(e)

		comp := &CreatureVisualComponent{
			Form:       form,
			SizeClass:  sizeClass,
			VisualTags: nil,
		}
		e.AddComponent(comp)

		// Mark animation dirty so sprite regenerates with correct template.
		if animComp := e.GetAnimation(); animComp != nil {
			animComp.Dirty = true
		}

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  e.ID,
				"form":       string(form),
				"size_class": sizeClass,
			}).Debug("inferred creature visual")
		}
	}
}

// inferForm determines creature form from available components.
func (s *CreatureVisualClassifierSystem) inferForm(e *Entity) CreatureForm {
	// Check faction hints.
	if factionComp, ok := e.GetComponent("faction"); ok {
		if fc, ok := factionComp.(*FactionComponent); ok {
			switch fc.FactionID {
			case "undead_faction":
				return FormUndead
			case "beast_faction", "animal_faction":
				return FormQuadruped
			case "mechanical_faction":
				return FormMechanical
			}
		}
	}

	// Heuristic: high-damage + huge collider → likely boss creature (flying/blob).
	attackComp, hasAttack := e.GetComponent("attack")
	colliderComp, hasCollider := e.GetComponent("collider")

	if hasAttack && hasCollider {
		attack, aOk := attackComp.(*AttackComponent)
		collider, cOk := colliderComp.(*ColliderComponent)
		if aOk && cOk {
			if attack.Damage > 30 && collider.Width >= 64 {
				return FormFlying // Huge boss → likely dragon-type
			}
		}
	}

	// Tiny enemies are likely insects/arachnids.
	if hasCollider {
		if collider, ok := colliderComp.(*ColliderComponent); ok {
			if collider.Width <= 16 {
				return FormArachnid
			}
		}
	}

	return FormHumanoid
}

// inferSizeClass determines size class from collider dimensions.
func (s *CreatureVisualClassifierSystem) inferSizeClass(e *Entity) string {
	colliderComp, ok := e.GetComponent("collider")
	if !ok {
		return "medium"
	}
	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		return "medium"
	}
	switch {
	case collider.Width <= 16:
		return "tiny"
	case collider.Width <= 24:
		return "small"
	case collider.Width <= 40:
		return "medium"
	case collider.Width <= 56:
		return "large"
	default:
		return "huge"
	}
}
