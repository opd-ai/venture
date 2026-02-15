// Package engine provides a system that propagates entity size class data from
// CreatureVisualComponent into the sprite generation pipeline. This ensures
// the size-based anatomy scaling (Tiny/Small/Medium/Large/Huge) is applied
// when generating or regenerating entity sprites, producing proportionally
// distinct silhouettes for different-sized entities.
package engine

import (
	"github.com/sirupsen/logrus"
)

// SizeSpriteScalingSystem ensures that entity size class information from
// CreatureVisualComponent is propagated to the animation pipeline. When an
// entity's size class changes or is first assigned, this system marks the
// entity's animation as dirty so the sprite regenerates with correct
// size-based anatomy proportions.
type SizeSpriteScalingSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewSizeSpriteScalingSystem creates a new size sprite scaling system.
func NewSizeSpriteScalingSystem(world *World) *SizeSpriteScalingSystem {
	var entry *logrus.Entry
	if world != nil && world.GetLogger() != nil {
		entry = world.GetLogger().WithFields(logrus.Fields{
			"system_name": "size_sprite_scaling",
		})
	}
	return &SizeSpriteScalingSystem{
		world:  world,
		logger: entry,
	}
}

// Update scans entities with CreatureVisualComponent and ensures size class
// data is synchronized into the sprite pipeline. When an entity's cached size
// differs from its current CreatureVisualComponent.SizeClass, the animation
// is marked dirty to trigger sprite regeneration with new proportions.
func (s *SizeSpriteScalingSystem) Update(entities []*Entity, deltaTime float64) {
	for _, e := range entities {
		cvComp, ok := e.GetComponent("creature_visual")
		if !ok {
			continue
		}
		cv, ok := cvComp.(*CreatureVisualComponent)
		if !ok || cv.SizeClass == "" || cv.SizeClass == "medium" {
			continue
		}

		// Check if size class is already cached on animation component
		animComp := e.GetAnimation()
		if animComp == nil {
			continue
		}

		// Use a custom field on the entity to track propagated size class.
		// If it matches the current CV size class, skip (already propagated).
		sizeComp, hasSizeMarker := e.GetComponent("size_sprite_marker")
		if hasSizeMarker {
			if marker, ok := sizeComp.(*SizeSpriteMarkerComponent); ok {
				if marker.PropagatedSizeClass == cv.SizeClass {
					continue
				}
			}
		}

		// Set the marker so we don't re-process this entity every frame
		e.AddComponent(&SizeSpriteMarkerComponent{
			PropagatedSizeClass: cv.SizeClass,
		})

		// Mark animation dirty to trigger sprite regeneration with size scaling
		animComp.Dirty = true

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  e.ID,
				"size_class": cv.SizeClass,
				"form":       string(cv.Form),
			}).Debug("propagated size class for sprite scaling")
		}
	}
}

// SizeSpriteMarkerComponent is a lightweight marker that tracks which size class
// has been propagated to the sprite pipeline. This prevents the system from
// redundantly marking animation dirty every frame.
type SizeSpriteMarkerComponent struct {
	PropagatedSizeClass string
}

// Type implements the Component interface.
func (c *SizeSpriteMarkerComponent) Type() string { return "size_sprite_marker" }
